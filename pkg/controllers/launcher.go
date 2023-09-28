/*
 * Copyright 2023 CoreLayer BV
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package controllers

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro"
	nitroConfig "github.com/corelayer/netscaleradc-nitro-go/pkg/nitro/resource/config"
	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro/resource/controllers"
	"github.com/corelayer/netscaleradc-nitro-go/pkg/registry"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"

	"github.com/corelayer/netscaleradc-acme-go/pkg/models"
	"github.com/corelayer/netscaleradc-acme-go/pkg/models/config"
)

const (
	LENS_CERTIFICATE_PATH = "/nsconfig/ssl/LENS/"
)

type Launcher struct {
	loader         Loader
	organizations  []registry.Organization
	users          map[string]*models.User
	providerParams []config.ProviderParameters
	timestamp      string

	providerChannels     map[string]chan config.Certificate
	installationChannels map[config.Target]chan config.Certificate
	errorChannel         chan error
	channelMapMutex      *sync.RWMutex

	registrationMutex *sync.Mutex
}

func NewLauncher(path string, organizations []registry.Organization, users []config.AcmeUser, params []config.ProviderParameters) (*Launcher, error) {
	var (
		err    error
		output *Launcher
	)
	output = &Launcher{
		loader:               NewLoader(path),
		organizations:        organizations,
		providerParams:       params,
		timestamp:            time.Now().Format("20060102150405"),
		providerChannels:     make(map[string]chan config.Certificate),
		installationChannels: make(map[config.Target]chan config.Certificate),
		errorChannel:         make(chan error),
		channelMapMutex:      &sync.RWMutex{},
		registrationMutex:    &sync.Mutex{},
	}

	output.users, err = output.initializeUsers(users)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (l Launcher) Request(name string) error {
	var (
		err   error
		certs map[string]config.Certificate
	)
	certs, err = l.loader.Get(name)
	if err != nil {
		return err
	}

	return l.processCertificates(certs)
}

func (l Launcher) RequestAll() error {
	var (
		err   error
		certs map[string]config.Certificate
	)
	certs, err = l.loader.GetAll()
	if err != nil {
		return err
	}

	return l.processCertificates(certs)
}

func (l Launcher) processCertificates(certs map[string]config.Certificate) error {
	var (
		providers     = make(map[string]int)
		installations = make(map[config.Target]int)

		wgProvider     sync.WaitGroup
		wgInstallation sync.WaitGroup
	)

	for _, c := range certs {
		if _, foundProvider := providers[c.Request.Challenge.Provider]; foundProvider {
			slog.Debug("found provider", "certificate", c.Name, "provider", c.Request.Challenge.Provider)
			providers[c.Request.Challenge.Provider] += 1
		} else {
			slog.Debug("adding provider", "certificate", c.Name, "provider", c.Request.Challenge.Provider)
			providers[c.Request.Challenge.Provider] = 1
		}

		for _, i := range c.Installation {
			if _, foundInstallation := installations[i.Target]; foundInstallation {
				slog.Debug("found installation target", "certificate", c.Name, "target", i.Target)
				installations[i.Target] += 1
			} else {
				slog.Debug("adding installation target", "certificate", c.Name, "target", i.Target)
				installations[i.Target] = 1
			}
		}
	}

	// Create channel per provider and launch processor
	for k, v := range providers {
		l.channelMapMutex.Lock()
		l.providerChannels[k] = make(chan config.Certificate, v)
		wgProvider.Add(1)
		go l.certificateProviderProcessor(k, l.providerChannels[k], &wgProvider)
		l.channelMapMutex.Unlock()
	}

	// Create channel per installation target and launch processor
	for i, v := range installations {
		l.channelMapMutex.Lock()
		l.installationChannels[i] = make(chan config.Certificate, v)
		wgInstallation.Add(1)
		go l.certificateInstallationProcessor(i, l.installationChannels[i], &wgInstallation)
		l.channelMapMutex.Unlock()
	}

	// Push certificates to their respective provider channel
	for _, c := range certs {
		slog.Info("process certificate", "certificate", c.Name, "provider", c.Request.Challenge.Provider)
		l.channelMapMutex.RLock()
		l.providerChannels[c.Request.Challenge.Provider] <- c
		l.channelMapMutex.RUnlock()
	}

	// Provider channels can be closed as soon as all certificate configurations are in the pipeline
	for _, ch := range l.providerChannels {
		l.channelMapMutex.Lock()
		close(ch)
		l.channelMapMutex.Unlock()
	}
	wgProvider.Wait()

	// Installation channels can be closed as soon as all provider processors have finished
	for _, ch := range l.installationChannels {
		l.channelMapMutex.Lock()
		close(ch)
		l.channelMapMutex.Unlock()
	}
	wgInstallation.Wait()

	// Error channel can be closed as soon as all installation processors have finished
	close(l.errorChannel)

	// TODO ADD CHECK FOR ERRORS which occurred in errorProcessor
	return nil
}

func (l Launcher) certificateProviderProcessor(p string, ch <-chan config.Certificate, wg *sync.WaitGroup) {
	var (
		err error
	)

	defer wg.Done()

	slog.Info("launching provider processor", "provider", p)
	for r := range ch {
		r.Resource, err = l.executeAcmeRequest(r)
		if err != nil {
			slog.Error("error occurred while processing request", "certificate", r.Name, "error", err)
			l.errorChannel <- err
		}
		for _, i := range r.Installation {
			l.channelMapMutex.Lock()
			l.installationChannels[i.Target] <- r
			l.channelMapMutex.Unlock()
		}
	}
	slog.Info("terminating provider processor", "provider", p)
}

func (l Launcher) certificateInstallationProcessor(t config.Target, ch <-chan config.Certificate, wg *sync.WaitGroup) {
	var (
		err error
	)

	defer wg.Done()

	slog.Info("launching installation processor", "target", t)
	for r := range ch {
		if r.Resource == nil {
			slog.Error("no certificate found to install", "target", t, "certificate", r.Name)
			l.errorChannel <- fmt.Errorf("no certificate found to install on target %s for %s", t, r.Name)
			continue
		}
		for _, i := range r.Installation {
			if i.Target == t {
				err = l.updateEnvironment(i, r.Name, r.Resource)
				if err != nil {
					slog.Error("RECEIVER ERROR", "errror", err)
					l.errorChannel <- err
					return
				}
			}
		}
	}
	slog.Info("terminating installation processor", "target", t)
}

func (l Launcher) errorProcessor(wg *sync.WaitGroup) {
	defer wg.Done()

	slog.Debug("launching error processor")
	for err := range l.errorChannel {
		slog.Error("error processing certificate", "error", err)
	}
}

func (l Launcher) initializeUsers(users []config.AcmeUser) (map[string]*models.User, error) {
	var (
		err    error
		output map[string]*models.User
	)
	output = make(map[string]*models.User, len(users))
	for _, u := range users {
		if _, exists := output[u.Name]; exists {
			slog.Error("user already exists", "username", u.Name)
			return nil, fmt.Errorf("user exists: %s", u.Name)
		}

		for _, v := range output {
			if u.Email == v.Email {
				slog.Error("user e-mail address already exists", "username", u.Name)
				return nil, fmt.Errorf("user e-mail address already exists for user: %s", u.Name)
			}
		}

		var user *models.User
		user, err = models.NewUser(u.Email)
		if err != nil {
			slog.Error("error creating user", "username", u.Name, "error", err)
			return nil, err
		}

		output[u.Name] = user
		slog.Debug("user added", "user", u.Name)
	}
	return output, nil
}

func (l Launcher) getUser(username string) (*models.User, error) {
	if _, exists := l.users[username]; !exists {
		slog.Error("user does not exist", "username", username)
		return nil, fmt.Errorf("user does not exist")
	}
	return l.users[username], nil
}

func (l Launcher) getLegoClient(username string, url string, keyType certcrypto.KeyType) (*lego.Client, error) {
	var (
		err    error
		user   *models.User
		client *lego.Client
	)

	l.registrationMutex.Lock()
	slog.Debug("locking for acme user account validation", "user", username)
	user, err = l.getUser(username)
	if err != nil {
		slog.Error("could not find user", "username", username)
		return nil, fmt.Errorf("could not find user %s with error %w", username, err)
	}

	legoConfig := lego.NewConfig(*user)
	legoConfig.CADirURL = url
	legoConfig.Certificate.KeyType = keyType

	client, err = lego.NewClient(legoConfig)
	if err != nil {
		slog.Error("could not create lego client", "username", username)
		return nil, fmt.Errorf("could not create lego client for user %s with error %w", username, err)
	}

	// Query registration
	if user.GetRegistration() == nil {
		slog.Debug("register acme account for user", "username", username)
		// New users will need to register
		var reg *registration.Resource
		reg, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			slog.Error("could not register acme account for user", "username", username, "error", err)
			return nil, fmt.Errorf("could not register user %s for acme request with error %w", username, err)
		}
		user.Registration = reg
	}
	l.registrationMutex.Unlock()
	slog.Debug("unlocking for acme user account validation", "user", username)

	return client, nil
}

func (l Launcher) executeAcmeRequest(cert config.Certificate) (*certificate.Resource, error) {
	var (
		err     error
		client  *lego.Client
		domains []string
	)
	slog.Info("execute acme request for certificate", "certificate", cert.Name)

	client, err = l.getLegoClient(cert.Request.AcmeUser, cert.Request.GetServiceUrl(), cert.Request.GetKeyType())
	if err != nil {
		return nil, err
	}

	var environment registry.Environment
	environment, err = l.getEnvironment(cert.Request.Target)
	if err != nil {
		slog.Debug("could not find organization environment for acme request", "organization", cert.Request.Target.Organization, "environment", cert.Request.Target.Environment)
		return nil, fmt.Errorf("could not find environment %s for organization %s for acme request with message %w", cert.Request.Target.Environment, cert.Request.Target.Organization, err)
	}

	var provider challenge.Provider
	provider, err = cert.Request.GetChallengeProvider(environment, l.timestamp)
	if err != nil {
		return nil, err
	}

	var providerParams config.ProviderParameters
	if cert.Request.Challenge.ProviderParameters != "" {
		providerParams, err = l.getProviderParameters(cert.Request.Challenge.ProviderParameters)
		if err != nil {
			return nil, err
		}

		err = providerParams.ApplyEnvironmentVariables()
		if err != nil {
			return nil, err
		}

		defer func() {
			err = providerParams.ResetEnvironmentVariables()
		}()
		if err != nil {
			return nil, err
		}

	}

	switch cert.Request.Challenge.Type {
	case config.ACME_CHALLENGE_TYPE_HTTP:
		err = client.Challenge.SetHTTP01Provider(provider)
	case config.ACME_CHALLENGE_TYPE_DNS:
		if cert.Request.Challenge.DisableDnsPropagationCheck {
			err = client.Challenge.SetDNS01Provider(provider, dns01.DisableCompletePropagationRequirement())
		} else {
			err = client.Challenge.SetDNS01Provider(provider)
		}
	case config.ACME_CHALLENGE_TYPE_TLS_ALPN:
		err = client.Challenge.SetTLSALPN01Provider(provider)
	default:
		err = fmt.Errorf("invalid challenge type")
	}
	if err != nil {
		return nil, err
	}

	// Get domains for ACME request
	if domains, err = cert.Request.GetDomains(); err != nil {
		slog.Error("invalid domain in request", "certificate", cert.Name, "error", err)
		return nil, err
	}

	// Execute ACME request
	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  false,
	}

	var certificates *certificate.Resource
	certificates, err = client.Certificate.Obtain(request)
	if err != nil {
		slog.Error("could not obtain certificate", "error", err)
		return nil, fmt.Errorf("could not obtain certificate with message %w", err)
	}

	block, _ := pem.Decode(certificates.Certificate)
	if block == nil {
		panic("failed to parse PEM block containing the public key")
	}
	pub, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic("failed to parse DER encoded public key: " + err.Error())
	}
	slog.Debug("certificate information", "cn", pub.Subject.CommonName, "SAN", pub.DNSNames)

	return certificates, nil
}

func (l Launcher) getCertificateFilename(name string) string {
	return name + "_" + l.timestamp + ".cer"
}

func (l Launcher) getPrivateKeyFilename(name string) string {
	return name + "_" + l.timestamp + ".key"
}

func (l Launcher) getSslCertKeyName(name string) string {
	return "LENS_" + name
}

func (l Launcher) uploadCertificates(c *nitro.Client, t config.Target, name string, cert *certificate.Resource) error {
	var (
		err error
	)
	slog.Info("upload certificate files to target", "target", t, "certificate", name)
	controller := controllers.NewSystemFileController(c)

	slog.Debug("uploading certificate public key to target", "target", t, "certificate", name)
	_, err = controller.Add(l.getCertificateFilename(name), LENS_CERTIFICATE_PATH, cert.Certificate)
	if err != nil {
		return fmt.Errorf("could not upload certificate public key to organization %s environment %s with message %w", t.Organization, t.Environment, err)
	}

	slog.Debug("uploading certificate private key to target", "target", t, "certificate", name)
	_, err = controller.Add(l.getPrivateKeyFilename(name), LENS_CERTIFICATE_PATH, cert.PrivateKey)
	if err != nil {
		return fmt.Errorf("could not upload certificate private key to organization %s environment %s with message %w", t.Organization, t.Environment, err)
	}
	return nil
}

func (l Launcher) configureSslCertKey(c *nitro.Client, name string, t config.Target) error {
	var (
		err       error
		unwrapErr error
	)
	slog.Info("configure ssl certkey on target", "target", t, "certificate", name)

	controller := controllers.NewSslCertKeyController(c)

	// Check if certificate exists
	if _, err = controller.Get(l.getSslCertKeyName(name), nil); err != nil {
		unwrapErr = errors.Unwrap(err)
		if !errors.Is(unwrapErr, nitro.NSERR_SSL_NOCERT) {
			slog.Error("could not verify if certificate exists on target", "target", t, "certificate", name, "error", err)
			return fmt.Errorf("could not verify if certificate exists in organization %s environment %s with message %w", t.Organization, t.Environment, err)
		} else {
			slog.Debug("creating ssl certkey on target", "target", t, "certificate", name)
			if _, err = controller.Add(l.getSslCertKeyName(name), LENS_CERTIFICATE_PATH+l.getCertificateFilename(name), LENS_CERTIFICATE_PATH+l.getPrivateKeyFilename(name)); err != nil {
				slog.Error("could not add certificate to environment", "target", t, "certificate", name, "error", err)
				return fmt.Errorf("could not add certificate to organization %s environment %s with message %w", t.Organization, t.Environment, err)
			}
		}
	} else {
		slog.Debug("updating ssl certkey on target", "target", t, "certificate", name)
		if _, err = controller.Update(l.getSslCertKeyName(name), LENS_CERTIFICATE_PATH+l.getCertificateFilename(name), LENS_CERTIFICATE_PATH+l.getPrivateKeyFilename(name), true); err != nil {
			slog.Error("could not update certificate exists in environment", "target", t, "certificate", name, "error", err)
			return fmt.Errorf("could not update certificate in organization %s environment %s with message %w", t.Organization, t.Environment, err)

		}
	}

	return nil
}

func (l Launcher) configureCertificates(c *nitro.Client, i config.Installation, name string) error {
	var (
		err error
	)

	err = l.configureSslCertKey(c, name, i.Target)
	if err != nil {
		return err
	}

	if len(i.SslVirtualServers) > 0 {
		err = l.bindSslVservers(c, name, i)
		if err != nil {
			return err
		}
	}

	if len(i.SslServices) > 0 {
		err = l.bindSslService(c, name, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l Launcher) updateEnvironment(i config.Installation, name string, cert *certificate.Resource) error {
	var (
		err    error
		e      registry.Environment
		client *nitro.Client
	)
	slog.Info("install certificate on target", "target", i.Target, "certificate", name)

	e, err = l.getEnvironment(i.Target)
	if err != nil {
		slog.Error("could not get environment for organization", "target", i.Target, "certificate", name)
		l.errorChannel <- fmt.Errorf("could not get environment %s for organization %s with message %w", i.Target.Environment, i.Target.Organization, err)
	}

	client, err = e.GetPrimaryNitroClient()

	err = l.uploadCertificates(client, i.Target, name, cert)
	if err != nil {
		return err
	}

	if i.ReplaceDefaultCertificate {
		err = l.replaceDefaultCertificate(client, i.Target, LENS_CERTIFICATE_PATH+l.getCertificateFilename(name), LENS_CERTIFICATE_PATH+l.getPrivateKeyFilename(name))
		if err != nil {
			slog.Error("could not replace default certificate", "target", i.Target)
			return err
		}
	} else {
		err = l.configureCertificates(client, i, name)
		if err != nil {
			//
			return err
		}
	}

	slog.Info("saving config on target", "target", i.Target)
	if err = client.SaveConfig(); err != nil {
		slog.Error("error saving config", "target", i.Target, "error", err)
		return err
	}
	slog.Info("process complete", "target", i.Target, "certificate", name)
	return nil
}

func (l Launcher) replaceDefaultCertificate(c *nitro.Client, t config.Target, certFilename string, keyFilename string) error {
	var (
		err error
	)
	slog.Info("replacing default certificate on target", "target", t)
	controller := controllers.NewSslCertKeyController(c)
	_, err = controller.Update("ns-server-certificate", certFilename, keyFilename, true)
	return err
}

func (l Launcher) bindSslVservers(c *nitro.Client, name string, i config.Installation) error {
	var (
		err error
	)
	slog.Info("bind certificate to ssl vservers", "target", i.Target)
	certKeyName := l.getSslCertKeyName(name)
	controller := controllers.NewSslCertKeyController(c)

	var bindings *nitro.Response[nitroConfig.SslCertKeySslVserverBinding]
	if bindings, err = controller.GetSslVserverBinding(certKeyName, nil); err != nil {
		slog.Error("could not verify if certificate exists", "target", i.Target, "certificate", name, "error", err)
		return fmt.Errorf("could not verify if certificate exists in organization %s environment %s with message %w", i.Target.Organization, i.Target.Environment, err)
	}
	if len(bindings.Data) == 0 {
		for _, bindTo := range i.SslVirtualServers {
			slog.Debug("bind certificate to ssl vserver", "target", i.Target, "certificate", name, "vserver", bindTo.Name)
			if _, err = controller.BindSslVserver(bindTo.Name, certKeyName, bindTo.SniEnabled); err != nil {
				slog.Error("could not bind certificate to vserver", "target", i.Target, "certificate", name, "error", err)
				// return fmt.Errorf("could not bind certificate %s to vserver in environment %s with message %w", certKeyName, e.Name, err)
			}
		}
	} else {
		// TODO UPDATE FLOW --> check if vserver name in SslVirtualServers exists before trying to bind
		slog.Debug("found existing bindings for certificate", "target", i.Target, "certificate", name, "count", len(bindings.Data))
		for _, bindTo := range i.SslVirtualServers {
			for _, boundTo := range bindings.Data {
				if bindTo.Name == boundTo.ServerName {
					slog.Debug("certificate already bound to vserver", "target", i.Target, "certificate", name, "vserver", bindTo.Name)
					continue
				} else {
					slog.Debug("binding certificate to vserver", "target", i.Target, "certificate", name, "vserver", bindTo.Name)
					if _, err = controller.BindSslVserver(bindTo.Name, certKeyName, bindTo.SniEnabled); err != nil {
						slog.Error("could not bind certificate to vserver", "target", i.Target, "certificate", name, "vserver", bindTo.Name, "error", err)
						// return fmt.Errorf("could not bind certificate %s to vserver in environment %s with message %w", certKeyName, e.Name, err)
						// TODO WHY NO RETURN?
					}
				}
			}
		}
	}
	return err
}

func (l Launcher) bindSslService(c *nitro.Client, name string, i config.Installation) error {
	var (
		err error
	)
	slog.Info("bind certificate to ssl services", "target", i.Target)
	certKeyName := l.getSslCertKeyName(name)
	controller := controllers.NewSslCertKeyController(c)

	var bindings *nitro.Response[nitroConfig.SslCertKeyServiceBinding]
	if bindings, err = controller.GetServiceBinding(certKeyName, nil); err != nil {
		slog.Error("could not verify if certificate exists on target", "target", i.Target, "certificate", name, "error", err)
		return fmt.Errorf("could not verify if certificate exists in organization %s environment %s with message %w", i.Target.Organization, i.Target.Environment, err)
	}
	if len(bindings.Data) == 0 {
		for _, bindTo := range i.SslServices {
			slog.Debug("bind certificate to ssl service", "target", i.Target, "certificate", name, "service", bindTo.Name)
			if _, err = controller.BindSslService(bindTo.Name, certKeyName, bindTo.SniEnabled); err != nil {
				slog.Error("could not bind certificate to ssl service", "organization", i.Target.Organization, "environment", i.Target.Environment, "certificate", certKeyName, "error", err)
				// return fmt.Errorf("could not bind certificate %s to service in environment %s with message %w", certKeyName, e.Name, err)
			}
		}
	} else {
		// TODO UPDATE FLOW --> check if service name in SslVirtualServers exists before trying to bind
		slog.Debug("found existing bindings for certificate", "target", i.Target, "certificate", name, "count", len(bindings.Data))
		for _, bindTo := range i.SslServices {
			for _, boundTo := range bindings.Data {
				if bindTo.Name == boundTo.ServiceName {
					slog.Debug("certificate already bound to ssl service", "target", i.Target, "certificate", name, "service", bindTo.Name)
					continue
				} else {
					slog.Debug("binding certificate to service", "target", i.Target, "certificate", name, "service", bindTo.Name)
					if _, err = controller.BindSslService(bindTo.Name, certKeyName, bindTo.SniEnabled); err != nil {
						slog.Error("could not bind certificate to ssl service", "target", i.Target, "certificate", name, "service", bindTo.Name, "error", err)
						// return fmt.Errorf("could not bind certificate %s to service in environment %s with message %w", certKeyName, e.Name, err)
						// TODO WHY NO RETURN?
					}
				}
			}
		}
	}
	return err
}

func (l Launcher) getEnvironment(t config.Target) (registry.Environment, error) {
	for _, org := range l.organizations {
		if t.Organization == org.Name {
			for _, env := range org.Environments {
				if t.Environment == env.Name {
					return env, nil
				}
			}
			break
		}
	}
	return registry.Environment{}, fmt.Errorf("could not find environment %s for organization %s", t.Environment, t.Organization)
}

func (l Launcher) getProviderParameters(name string) (config.ProviderParameters, error) {
	for _, p := range l.providerParams {
		if name == p.Name {
			return p, nil
		}
	}
	return config.ProviderParameters{}, fmt.Errorf("could not find provider parameters for %s", name)
}
