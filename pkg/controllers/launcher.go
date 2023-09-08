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
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"

	"github.com/corelayer/netscaleradc-acme-go/pkg/lego/providers/netscaleradc"
	"github.com/corelayer/netscaleradc-acme-go/pkg/models"
	"github.com/corelayer/netscaleradc-acme-go/pkg/models/config"
)

type Launcher struct {
	loader        Loader
	organizations []registry.Organization
	users         map[string]*models.User
	timestamp     string
}

func NewLauncher(path string, organizations []registry.Organization, users []config.AcmeUser) (*Launcher, error) {
	var (
		err    error
		output *Launcher
	)
	output = &Launcher{
		organizations: organizations,
		loader:        NewLoader(path),
		timestamp:     time.Now().Format("20060102150405"),
	}
	output.users, err = output.initialize(users)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (l Launcher) Request(name string) error {
	var (
		err          error
		req          config.Certificate
		certificates *certificate.Resource
	)
	req, err = l.loader.Get(name)
	if err != nil {
		return err
	}

	certificates, err = l.executeAcmeRequest(req)
	if err != nil {
		return err
	}

	slog.Info(certificates.Domain)
	return l.updateNetScaler(req, certificates)
	// return nil
}

func (l Launcher) RequestAll() error {
	var (
		err          error
		req          map[string]config.Certificate
		certificates *certificate.Resource
	)
	req, err = l.loader.GetAll()
	if err != nil {
		return err
	}

	// TODO UPDATE GOROUTINE CALLS TO HANDLE ERRORS
	var wg sync.WaitGroup
	for _, v := range req {
		wg.Add(1)
		go func(c config.Certificate, group *sync.WaitGroup) {
			slog.Debug("Requesting certficate", "domain", c.Name)
			defer group.Done()
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
		}(v, &wg)
	}
	wg.Wait()
	return nil
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

	user, err = l.getUser(cert.AcmeRequest.Username)
	if err != nil {
		slog.Debug("could not find user", "username", cert.AcmeRequest.Username, "config", cert.Name)
		return nil, fmt.Errorf("could not find user %s for config %s with message %w", cert.AcmeRequest.Username, cert.Name, err)
	}

	legoConfig := lego.NewConfig(user)
	legoConfig.CADirURL = cert.AcmeRequest.GetServiceUrl()
	legoConfig.Certificate.KeyType = cert.AcmeRequest.GetKeyType()

	client, err = lego.NewClient(legoConfig)
	if err != nil {
		slog.Error("could not create lego client", "config", cert.Name)
	}

	var environment registry.Environment
	environment, err = l.getEnvironment(cert.AcmeRequest.Organization, cert.AcmeRequest.Environment)
	if err != nil {
		slog.Debug("could not find organization environment for acme request", "organization", cert.AcmeRequest.Organization, "environment", cert.AcmeRequest.Environment)
		return nil, fmt.Errorf("could not find environment %s for organization %s for acme request with message %w", cert.AcmeRequest.Environment, cert.AcmeRequest.Organization, err)
	}

	var provider challenge.Provider
	provider, err = cert.AcmeRequest.GetChallengeProvider(environment, l.timestamp)
	if err != nil {
		return nil, err
	}

	switch cert.AcmeRequest.ChallengeType {
	case netscaleradc.ACME_CHALLENGE_TYPE_NETSCALER_HTTP_GLOBAL:
		err = client.Challenge.SetHTTP01Provider(provider)
	case netscaleradc.ACME_CHALLENGE_TYPE_NETSCALER_ADNS:
		err = client.Challenge.SetDNS01Provider(provider)
	case config.ACME_CHALLENGE_TYPE_HTTP:
		err = client.Challenge.SetHTTP01Provider(provider)
	case config.ACME_CHALLENGE_TYPE_DNS:
		err = client.Challenge.SetDNS01Provider(provider)
	default:
		err = fmt.Errorf("invalid provider")
	}
	if err != nil {
		return nil, err
	}

	// Get domains for ACME request
	if domains, err = cert.AcmeRequest.GetDomains(); err != nil {
		slog.Error("invalid domain in request", "certificate", cert.Name, "error", err)

		return nil, err
	}

	// New users will need to register
	var reg *registration.Resource
	reg, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	// reg, err = client.Registration.QueryRegistration()
	if err != nil {
		slog.Error("could not register user", "error", err)
		return nil, fmt.Errorf("could not register user %s for acme request with message %w", cert.AcmeRequest.Username, err)
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

func (l Launcher) updateNetScaler(certConfig config.Certificate, acmeCert *certificate.Resource) error {
	var (
		err error
	)

	if acmeCert == nil {
		slog.Error("no certificate available for upload")
		return errors.New("no certificate available for upload")
	}
	var environments []registry.Environment
	for _, b := range certConfig.Bindpoints {
		var env registry.Environment
		env, err = l.getEnvironment(b.Organization, b.Environment)
		if err != nil {
			slog.Error("could not get environment for organization")
			return fmt.Errorf("could not get environment %s for organization %s with message %w", b.Environment, b.Organization, err)
		}

		environments = append(environments, env)
	}

	for _, e := range environments {
		var client *nitro.Client
		client, err = e.GetPrimaryNitroClient()

		if err != nil {
			slog.Error("could not get nitro client for environment")
			return fmt.Errorf("could not get nitro client for environment %s with message %w", e.Name, err)
		}
		fc := controllers.NewSystemFileController(client)
		certFilename := certConfig.Name + "_" + l.timestamp + ".cer"
		pkeyFilename := certConfig.Name + "_" + l.timestamp + ".key"
		location := "/nsconfig/ssl/CERTS/LENS/"
		slog.Debug("uploading certificate public key to environment", "environment", e.Name, "certificate", certConfig.Name)
		_, err = fc.Add(certFilename, location, acmeCert.Certificate)
		if err != nil {
			return fmt.Errorf("could not upload certificate public key to environment %s with message %w", e.Name, err)
		}
		slog.Debug("uploading certificate private key to environment", "environment", e.Name)
		_, err = fc.Add(pkeyFilename, location, acmeCert.PrivateKey)
		if err != nil {
			return fmt.Errorf("could not upload certificate private key to environment %s with message %w", e.Name, err)
		}

		certKeyName := "LENS_" + certConfig.Name
		certc := controllers.NewSslCertKeyController(client)
		// Check if certificate exists
		// var certKey *nitro.Response[nitroConfig.SslCertKey]
		var uErr error
		if _, err = certc.Get(certKeyName, nil); err != nil {
			uErr = errors.Unwrap(err)
			if !errors.Is(uErr, nitro.NSERR_SSL_NOCERT) {
				slog.Error("could not verify if certificate exists in environment", "environment", e.Name, "certificate", certConfig.Name, "error", err)
				return fmt.Errorf("could not verify if certificate exists in environment %s with message %w", e.Name, err)
			} else {
				slog.Info("creating ssl certkey in environment", "environment", e.Name, "certificate", certConfig.Name)
				if _, err = certc.Add(certKeyName, location+certFilename, location+pkeyFilename); err != nil {
					slog.Error("could not add certificate to environment", "environment", e.Name, "certificate", certConfig.Name, "error", err)
					return fmt.Errorf("could not add certificate to environment %s with message %w", e.Name, err)
				}
			}
		} else {
			slog.Info("updating ssl certkey in environment", "environment", e.Name)
			if _, err = certc.Update(certKeyName, location+certFilename, location+pkeyFilename); err != nil {
				slog.Error("could not update certificate exists in environment", "environment", e.Name, "certificate", certConfig.Name, "error", err)
				return fmt.Errorf("could not update certificate in environment %s with message %w", e.Name, err)

			}
		}

		for _, b := range certConfig.Bindpoints {
			var bindings *nitro.Response[nitroConfig.SslCertKey_SslVserver_Binding]
			if bindings, err = certc.GetSslVserverBinding(certKeyName, nil); err != nil {
				slog.Error("could not verify if certificate exists in environment", "environment", e.Name, "certificate", certConfig.Name, "error", err)
				return fmt.Errorf("could not verify if certificate exists in environment %s with message %w", e.Name, err)
			}
			if len(bindings.Data) == 0 {
				for _, bindTo := range b.SslVservers {
					// TODO ADD LOGGING FOR EACH SSL VSERVER
					if _, err = certc.Bind(bindTo.Name, certKeyName, bindTo.SniEnabled); err != nil {
						slog.Error("could not bind certificate to vserver", "environment", e.Name, "certificate", certConfig.Name, "error", err)
						// return fmt.Errorf("could not bind certificate %s to vserver in environment %s with message %w", certKeyName, e.Name, err)
					}
				}
			} else {
				// TODO UPDATE FLOW --> check if vserver name in SslVservers exists before trying to bind
				slog.Debug("found existing bindings for certificate", "environment", e.Name, "certificate", certConfig.Name, "count", len(bindings.Data))
				for _, bindTo := range b.SslVservers {
					for _, boundTo := range bindings.Data {
						if bindTo.Name == boundTo.ServerName {
							slog.Debug("certificate already bound to vserver", "certificate", certKeyName, "vserver", bindTo.Name)
							continue
						} else {
							slog.Debug("binding certificate to vserver", "certificate", certKeyName, "vserver", bindTo.Name)
							if _, err = certc.Bind(bindTo.Name, certKeyName, bindTo.SniEnabled); err != nil {
								slog.Error("could not bind certificate to vserver", "environment", e.Name, "certificate", certConfig.Name, "error", err)
								// return fmt.Errorf("could not bind certificate %s to vserver in environment %s with message %w", certKeyName, e.Name, err)
							}
						}
					}
				}
			}
		}
	}
	return nil
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
