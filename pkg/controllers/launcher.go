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
}

func NewLauncher(path string, organizations []registry.Organization, users []config.AcmeUser, params []config.ProviderParameters) (*Launcher, error) {
	var (
		err    error
		output *Launcher
	)
	output = &Launcher{
		loader:         NewLoader(path),
		organizations:  organizations,
		providerParams: params,
		timestamp:      time.Now().Format("20060102150405"),
	}
	output.users, err = output.initialize(users)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (l Launcher) Request(name string) error {
	var (
		err error
		req config.Certificate
		// certificates *certificate.Resource
	)
	req, err = l.loader.Get(name)
	if err != nil {
		return err
	}

	// TODO UPDATE GOROUTINE CALLS TO HANDLE ERRORS
	var wg sync.WaitGroup
	wg.Add(1)
	go l.processRequest(req, &wg)
	wg.Wait()
	return nil

	// certificates, err = l.executeAcmeRequest(req)
	// if err != nil {
	// 	return err
	// }
	//
	// slog.Info(certificates.Domain)
	// return l.updateNetScaler(req, certificates)
	// return nil
}

func (l Launcher) RequestAll() error {
	var (
		err error
		req map[string]config.Certificate
		// certificates *certificate.Resource
	)
	req, err = l.loader.GetAll()
	if err != nil {
		return err
	}

	// TODO UPDATE GOROUTINE CALLS TO HANDLE ERRORS
	var wg sync.WaitGroup
	for _, v := range req {
		wg.Add(1)
		go l.processRequest(v, &wg)
	}
	wg.Wait()
	return nil
}

func (l Launcher) processRequest(c config.Certificate, wg *sync.WaitGroup) {
	var certificates *certificate.Resource
	slog.Debug("Requesting certficate", "domain", c.Name)
	defer wg.Done()
	var gorErr error
	certificates, gorErr = l.executeAcmeRequest(c)
	if gorErr != nil {
		slog.Error(gorErr.Error())
		return
	}

	slog.Info(certificates.Domain)
	gorErr = l.updateNetScaler(c, certificates)
	if gorErr != nil {
		slog.Error(gorErr.Error())
		return
	}
}

func (l Launcher) initialize(users []config.AcmeUser) (map[string]*models.User, error) {
	var (
		err    error
		output map[string]*models.User
	)
	output = make(map[string]*models.User, len(users))
	for _, v := range users {
		slog.Debug("adding user to configuration", "user", v.Name)
		var u *models.User
		u, err = models.NewUser(v.Email)
		if err != nil {
			slog.Error("error adding user", "error", err)
			return nil, err
		}

		if _, exists := output[v.Name]; exists {
			return nil, fmt.Errorf("user exists")
		}

		output[v.Name] = u
		slog.Debug("user added", "user", v.Name)
	}
	return output, nil
}

func (l Launcher) getUser(username string) (*models.User, error) {
	for k, v := range l.users {
		slog.Debug("user in configuration", "username", k, "value", v.Email)
	}
	if _, exists := l.users[username]; !exists {
		slog.Error("user does not exist", "username", username)
		return nil, fmt.Errorf("user does not exist")
	}
	return l.users[username], nil
}

func (l Launcher) executeAcmeRequest(cert config.Certificate) (*certificate.Resource, error) {
	var (
		err     error
		user    *models.User
		client  *lego.Client
		domains []string
	)

	user, err = l.getUser(cert.Request.AcmeUser)
	if err != nil {
		slog.Debug("could not find user", "username", cert.Request.AcmeUser, "config", cert.Name)
		return nil, fmt.Errorf("could not find user %s for config %s with message %w", cert.Request.AcmeUser, cert.Name, err)
	}

	legoConfig := lego.NewConfig(user)
	legoConfig.CADirURL = cert.Request.GetServiceUrl()
	legoConfig.Certificate.KeyType = cert.Request.GetKeyType()

	client, err = lego.NewClient(legoConfig)
	if err != nil {
		slog.Error("could not create lego client", "config", cert.Name)
	}

	var environment registry.Environment
	environment, err = l.getEnvironment(cert.Request.Organization, cert.Request.Environment)
	if err != nil {
		slog.Debug("could not find organization environment for acme request", "organization", cert.Request.Organization, "environment", cert.Request.Environment)
		return nil, fmt.Errorf("could not find environment %s for organization %s for acme request with message %w", cert.Request.Environment, cert.Request.Organization, err)
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

	// New users will need to register
	var reg *registration.Resource
	reg, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	// reg, err = client.Registration.QueryRegistration()
	if err != nil {
		slog.Error("could not register user", "error", err)
		return nil, fmt.Errorf("could not register user %s for acme request with message %w", cert.Request.AcmeUser, err)
	}
	user.Registration = reg

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

func (l Launcher) updateEnvironment(certName string, installation config.Installation, acmeCert *certificate.Resource) error {
	var (
		err    error
		e      registry.Environment
		client *nitro.Client
	)
	e, err = l.getEnvironment(installation.Organization, installation.Environment)
	if err != nil {
		slog.Error("could not get environment for organization")
		return fmt.Errorf("could not get environment %s for organization %s with message %w", installation.Environment, installation.Organization, err)
	}

	client, err = e.GetPrimaryNitroClient()

	if err != nil {
		slog.Error("could not get nitro client for environment")
		return fmt.Errorf("could not get nitro client for environment %s with message %w", e.Name, err)
	}
	fc := controllers.NewSystemFileController(client)
	certFilename := certName + "_" + l.timestamp + ".cer"
	pkeyFilename := certName + "_" + l.timestamp + ".key"
	slog.Debug("uploading certificate public key to environment", "environment", e.Name, "certificate", certName)
	_, err = fc.Add(certFilename, LENS_CERTIFICATE_PATH, acmeCert.Certificate)
	if err != nil {
		return fmt.Errorf("could not upload certificate public key to environment %s with message %w", e.Name, err)
	}
	slog.Debug("uploading certificate private key to environment", "environment", e.Name)
	_, err = fc.Add(pkeyFilename, LENS_CERTIFICATE_PATH, acmeCert.PrivateKey)
	if err != nil {
		return fmt.Errorf("could not upload certificate private key to environment %s with message %w", e.Name, err)
	}

	certc := controllers.NewSslCertKeyController(client)

	if installation.ReplaceDefaultCertificate {
		err = l.replaceDefaultCertificate(client, installation.Environment, LENS_CERTIFICATE_PATH+certFilename, LENS_CERTIFICATE_PATH+pkeyFilename)
		if err != nil {
			slog.Error("could not replace default certificate", "environment", installation.Environment)
			return err
		}
		time.Sleep(5 * time.Second)
	} else {
		certKeyName := "LENS_" + certName

		// Check if certificate exists
		// var certKey *nitro.Response[nitroConfig.SslCertKey]
		var uErr error
		if _, err = certc.Get(certKeyName, nil); err != nil {
			uErr = errors.Unwrap(err)
			if !errors.Is(uErr, nitro.NSERR_SSL_NOCERT) {
				slog.Error("could not verify if certificate exists in environment", "environment", e.Name, "certificate", certName, "error", err)
				return fmt.Errorf("could not verify if certificate exists in environment %s with message %w", e.Name, err)
			} else {
				slog.Info("creating ssl certkey in environment", "environment", e.Name, "certificate", certName)
				if _, err = certc.Add(certKeyName, LENS_CERTIFICATE_PATH+certFilename, LENS_CERTIFICATE_PATH+pkeyFilename); err != nil {
					slog.Error("could not add certificate to environment", "environment", e.Name, "certificate", certName, "error", err)
					return fmt.Errorf("could not add certificate to environment %s with message %w", e.Name, err)
				}
			}
		} else {
			slog.Info("updating ssl certkey in environment", "environment", e.Name)
			if _, err = certc.Update(certKeyName, LENS_CERTIFICATE_PATH+certFilename, LENS_CERTIFICATE_PATH+pkeyFilename, true); err != nil {
				slog.Error("could not update certificate exists in environment", "environment", e.Name, "certificate", certName, "error", err)
				return fmt.Errorf("could not update certificate in environment %s with message %w", e.Name, err)

			}
		}

		err = l.bindSslVservers(client, certKeyName, installation)
		if err != nil {
			return err
		}

		err = l.bindSslService(client, certKeyName, installation)
		if err != nil {
			return err
		}
	}

	// TODO - SAVE CONFIG LOGIC
	slog.Debug("saving config")
	if err = client.SaveConfig(); err != nil {
		slog.Error("error saving config", "environment", e.Name, "error", err)
		return err
	}
	return nil
}

func (l Launcher) updateNetScaler(certConfig config.Certificate, acmeCert *certificate.Resource) error {
	var (
		err error
	)

	if acmeCert == nil {
		slog.Error("no certificate available for upload")
		return errors.New("no certificate available for upload")
	}

	wg := sync.WaitGroup{}
	// TODO updateNetScaler - validate configuration so that org/env does not appear more than once in installation section
	for _, b := range certConfig.Installation {
		wg.Add(1)
		go func(certName string, installation config.Installation, acmeCert *certificate.Resource, w *sync.WaitGroup) {
			defer wg.Done()

			err = l.updateEnvironment(certName, installation, acmeCert)

		}(certConfig.Name, b, acmeCert, &wg)
		// environments = append(environments, env)

	}
	wg.Wait()
	return err
}

func (l Launcher) replaceDefaultCertificate(c *nitro.Client, environment string, certFilename string, keyFilename string) error {
	var (
		err error
	)
	slog.Debug("replacing default certificate", "environment", environment)
	certc := controllers.NewSslCertKeyController(c)
	_, err = certc.Update("ns-server-certificate", certFilename, keyFilename, true)
	return err
}

func (l Launcher) bindSslVservers(c *nitro.Client, certKeyName string, b config.Installation) error {
	var (
		err error
	)
	certc := controllers.NewSslCertKeyController(c)
	var bindings *nitro.Response[nitroConfig.SslCertKeySslVserverBinding]
	if bindings, err = certc.GetSslVserverBinding(certKeyName, nil); err != nil {
		slog.Error("could not verify if certificate exists in environment", "environment", b.Environment, "certificate", certKeyName, "error", err)
		return fmt.Errorf("could not verify if certificate exists in environment %s with message %w", b.Environment, err)
	}
	if len(bindings.Data) == 0 {
		for _, bindTo := range b.SslVirtualServers {
			// TODO ADD LOGGING FOR EACH SSL VSERVER
			if _, err = certc.BindSslVserver(bindTo.Name, certKeyName, bindTo.SniEnabled); err != nil {
				slog.Error("could not bind certificate to vserver", "environment", b.Environment, "certificate", certKeyName, "error", err)
				// return fmt.Errorf("could not bind certificate %s to vserver in environment %s with message %w", certKeyName, e.Name, err)
			}
		}
	} else {
		// TODO UPDATE FLOW --> check if vserver name in SslVirtualServers exists before trying to bind
		slog.Debug("found existing bindings for certificate", "environment", b.Environment, "certificate", certKeyName, "count", len(bindings.Data))
		for _, bindTo := range b.SslVirtualServers {
			for _, boundTo := range bindings.Data {
				if bindTo.Name == boundTo.ServerName {
					slog.Debug("certificate already bound to vserver", "certificate", certKeyName, "vserver", bindTo.Name)
					continue
				} else {
					slog.Debug("binding certificate to vserver", "certificate", certKeyName, "vserver", bindTo.Name)
					if _, err = certc.BindSslVserver(bindTo.Name, certKeyName, bindTo.SniEnabled); err != nil {
						slog.Error("could not bind certificate to vserver", "environment", b.Environment, "certificate", certKeyName, "error", err)
						// return fmt.Errorf("could not bind certificate %s to vserver in environment %s with message %w", certKeyName, e.Name, err)
					}
				}
			}
		}
	}
	return err
}

func (l Launcher) bindSslService(c *nitro.Client, certKeyName string, b config.Installation) error {
	var (
		err error
	)
	certc := controllers.NewSslCertKeyController(c)
	var bindings *nitro.Response[nitroConfig.SslCertKeyServiceBinding]
	if bindings, err = certc.GetServiceBinding(certKeyName, nil); err != nil {
		slog.Error("could not verify if certificate exists in environment", "environment", b.Environment, "certificate", certKeyName, "error", err)
		return fmt.Errorf("could not verify if certificate exists in environment %s with message %w", b.Environment, err)
	}
	if len(bindings.Data) == 0 {
		for _, bindTo := range b.SslServices {
			// TODO ADD LOGGING FOR EACH SSL SERVICE
			if _, err = certc.BindSslService(bindTo.Name, certKeyName, bindTo.SniEnabled); err != nil {
				slog.Error("could not bind certificate to service", "environment", b.Environment, "certificate", certKeyName, "error", err)
				// return fmt.Errorf("could not bind certificate %s to service in environment %s with message %w", certKeyName, e.Name, err)
			}
		}
	} else {
		// TODO UPDATE FLOW --> check if service name in SslVirtualServers exists before trying to bind
		slog.Debug("found existing bindings for certificate", "environment", b.Environment, "certificate", certKeyName, "count", len(bindings.Data))
		for _, bindTo := range b.SslServices {
			for _, boundTo := range bindings.Data {
				if bindTo.Name == boundTo.ServiceName {
					slog.Debug("certificate already bound to service", "certificate", certKeyName, "service", bindTo.Name)
					continue
				} else {
					slog.Debug("binding certificate to service", "certificate", certKeyName, "service", bindTo.Name)
					if _, err = certc.BindSslService(bindTo.Name, certKeyName, bindTo.SniEnabled); err != nil {
						slog.Error("could not bind certificate to service", "environment", b.Environment, "certificate", certKeyName, "error", err)
						// return fmt.Errorf("could not bind certificate %s to service in environment %s with message %w", certKeyName, e.Name, err)
					}
				}
			}
		}
	}
	return err
}

func (l Launcher) getEnvironment(organization string, environment string) (registry.Environment, error) {
	for _, org := range l.organizations {
		if organization == org.Name {
			for _, env := range org.Environments {
				if environment == env.Name {
					return env, nil
				}
			}
			break
		}
	}
	return registry.Environment{}, fmt.Errorf("could not find environment %s for organization %s", environment, organization)
}

func (l Launcher) getProviderParameters(name string) (config.ProviderParameters, error) {
	for _, p := range l.providerParams {
		if name == p.Name {
			return p, nil
		}
	}
	return config.ProviderParameters{}, fmt.Errorf("could not find provider parameters for %s", name)
}
