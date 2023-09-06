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
	"github.com/go-acme/lego/certcrypto"
	"github.com/go-acme/lego/certificate"
	"github.com/go-acme/lego/challenge"
	"github.com/go-acme/lego/lego"
	"github.com/go-acme/lego/registration"

	"github.com/corelayer/netscaleradc-acme-go/pkg/lego/providers/http/netscaleradc"
	"github.com/corelayer/netscaleradc-acme-go/pkg/models"
	"github.com/corelayer/netscaleradc-acme-go/pkg/models/config"
)

type Launcher struct {
	loader        ConfigLoader
	organizations []registry.Organization
	user          config.AcmeUser
	timestamp     string
}

func NewLauncher(organizations []registry.Organization, path string, user config.AcmeUser) *Launcher {
	return &Launcher{
		organizations: organizations,
		loader:        NewConfigLoader(path),
		user:          user,
		timestamp:     time.Now().Format("20060102150405"),
	}
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

	certificates, err = l.launchLego(req)
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
			certificates, gorErr = l.launchLego(c)
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

func (l Launcher) launchLego(cert config.Certificate) (*certificate.Resource, error) {
	var (
		err    error
		user   *models.User
		client *lego.Client
	)
	user, err = models.NewUser(l.user.Email)
	if err != nil {
		slog.Debug("could not create user", "email", l.user.Email, "config", cert.Name)
		return nil, fmt.Errorf("could not create user %s for config %s with message %w", l.user.Email, cert.Name, err)
	}

	legoConfig := lego.NewConfig(user)
	legoConfig.CADirURL = lego.LEDirectoryStaging
	legoConfig.Certificate.KeyType = certcrypto.RSA2048

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
	provider, err = netscaleradc.NewGlobalHttpProvider(environment, 10, l.timestamp)
	if err != nil {
		return nil, err
	}
	err = client.Challenge.SetHTTP01Provider(provider)

	// New users will need to register
	var reg *registration.Resource
	reg, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	// reg, err = client.Registration.QueryRegistration()
	if err != nil {
		slog.Error("could not register user", "error", err)
		return nil, fmt.Errorf("could not register user %s for acme request with message %w", l.user.Email, err)
	}
	user.Registration = reg

	request := certificate.ObtainRequest{
		Domains: cert.AcmeRequest.GetDomains(),
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
			fmt.Printf("####\r\nERROR: %s\r\nUNWRAP: %s\r\n####\r\n", err, uErr)
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
