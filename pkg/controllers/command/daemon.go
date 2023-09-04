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

package command

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"

	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro"
	nitroConfig "github.com/corelayer/netscaleradc-nitro-go/pkg/nitro/resource/config"
	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro/resource/controllers"
	"github.com/corelayer/netscaleradc-nitro-go/pkg/registry"
	"github.com/go-acme/lego/certcrypto"
	"github.com/go-acme/lego/certificate"
	"github.com/go-acme/lego/challenge"
	"github.com/go-acme/lego/lego"
	"github.com/go-acme/lego/registration"
	"github.com/spf13/viper"

	"github.com/corelayer/netscaleradc-acme-go/pkg/lego/providers/http/netscaleradc"
	"github.com/corelayer/netscaleradc-acme-go/pkg/models"
	"github.com/corelayer/netscaleradc-acme-go/pkg/models/config"
)

type Daemon struct {
	Config    config.Application
	Timestamp string
}

func (c Daemon) Execute() error {
	var err error
	if _, err = net.Listen("tcp", c.Config.Daemon.Address+":"+strconv.Itoa(c.Config.Daemon.Port)); err != nil {
		slog.Error("a daemon is already running on the same address")
		return err
	}
	slog.Info("Running daemon", "address", c.Config.Daemon.Address, "port", c.Config.Daemon.Port)

	var files []string
	files, err = c.listConfigFiles(c.Config.ConfigPath)
	if err != nil {
		slog.Debug("could not read config snippets", "error", err)
		return err
	}

	var configs map[string]*viper.Viper
	configs, err = c.getVipers(files)
	if err != nil {
		slog.Debug("could not load config from file", "error", err)
		return err
	}

	for key, currentConfig := range configs {
		var uConfig config.Certificate
		err = currentConfig.Unmarshal(&uConfig)
		if err != nil {
			slog.Debug("could not unmarshal config", "config", key)
			// return fmt.Errorf("could not unmarshal config for %s with message %w", k, err)
		}
		slog.Info("certificate config loaded for processing", "name", uConfig.Name)
		var certificates *certificate.Resource
		certificates, err = c.launchLego(uConfig)
		if err != nil {
			slog.Debug("failed to process acme request", "name", uConfig.Name, "error", err)
			// return fmt.Errorf("failed to process acme request for %s with message %w", uConfig.Name, err)
		}

		err = c.updateNetScaler(uConfig, certificates)
		if err != nil {
			slog.Error("aborting acme request", "domain", uConfig.Name)
		}

	}
	return nil
}

func (c Daemon) updateNetScaler(certConfig config.Certificate, acmeCert *certificate.Resource) error {
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
		env, err = c.Config.GetEnvironment(b.Organization, b.Environment)
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
		certFilename := certConfig.Name + "_" + c.Timestamp + ".cer"
		pkeyFilename := certConfig.Name + "_" + c.Timestamp + ".key"
		location := "/nsconfig/ssl/CERTS/LENS/"
		slog.Debug("uploading certificate public key to environment", "environment", e.Name)
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
			fmt.Printf("####\r\nERROR: %s\r\nUNWRAP: %s\r\n####\r\rn", err, uErr)
			if !errors.Is(uErr, NSERR_SSL_NOCERT) {
				slog.Error("could not verify if certificate exists in environment", "environment", e.Name, "error", err)
				return fmt.Errorf("could not verify if certificate exists in environment %s with message %w", e.Name, err)
			} else {
				slog.Info("creating ssl certkey in environment", "environment", e.Name)
				if _, err = certc.Add(certKeyName, location+certFilename, location+pkeyFilename); err != nil {
					slog.Error("could not add certificate to environment", "environment", e.Name, "error", err)
					return fmt.Errorf("could not add certificate to environment %s with message %w", e.Name, err)
				}
			}
		} else {
			slog.Info("updating ssl certkey in environment", "environment", e.Name)
			if _, err = certc.Update(certKeyName, location+certFilename, location+pkeyFilename); err != nil {
				slog.Error("could not update certificate exists in environment", "environment", e.Name, "error", err)
				return fmt.Errorf("could not update certificate in environment %s with message %w", e.Name, err)

			}
		}

		for _, b := range certConfig.Bindpoints {
			var bindings *nitro.Response[nitroConfig.SslCertKey_SslVserver_Binding]
			if bindings, err = certc.GetSslVserverBinding(certKeyName, nil); err != nil {
				slog.Error("could not verify if certificate exists in environment", "environment", e.Name, "error", err)
				return fmt.Errorf("could not verify if certificate exists in environment %s with message %w", e.Name, err)
			}
			if len(bindings.Data) == 0 {
				for _, bindTo := range b.SslVservers {
					// TODO ADD LOGGING FOR EACH SSL VSERVER
					if _, err = certc.Bind(bindTo.Name, certKeyName, bindTo.SniEnabled); err != nil {
						slog.Error("could not bind certificate to vserver", "environment", e.Name, "certkey", certKeyName, "error", err)
						// return fmt.Errorf("could not bind certificate %s to vserver in environment %s with message %w", certKeyName, e.Name, err)
					}
				}
			} else {
				// TODO UPDATE FLOW --> check if vserver name in SslVservers exists before trying to bind
				slog.Debug("found existing bindings for certificate", "environment", e.Name, "certkey", certKeyName, "count", len(bindings.Data))
				for _, bindTo := range b.SslVservers {
					for _, boundTo := range bindings.Data {
						if bindTo.Name == boundTo.ServerName {
							slog.Debug("certificate already bound to vserver", "certificate", certKeyName, "vserver", bindTo.Name)
							break
						}
					}
					slog.Debug("binding certificate to vserver", "certificate", certKeyName, "vserver", bindTo.Name)
					if _, err = certc.Bind(bindTo.Name, certKeyName, bindTo.SniEnabled); err != nil {
						slog.Error("could not bind certificate to vserver", "environment", e.Name, "certkey", certKeyName, "error", err)
						// return fmt.Errorf("could not bind certificate %s to vserver in environment %s with message %w", certKeyName, e.Name, err)
					}
				}
			}
		}
	}
	return nil
}

var NSERR_SSL_NOCERT = nitro.Error{}.WithCode(1540).WithMessage("Certificate does not exist ")

func (c Daemon) launchLego(config config.Certificate) (*certificate.Resource, error) {
	var (
		err    error
		user   *models.User
		client *lego.Client
	)
	user, err = models.NewUser(c.Config.User.Email)
	if err != nil {
		slog.Debug("could not create user", "email", c.Config.User.Email, "config", config.Name)
		return nil, fmt.Errorf("could not create user %s for config %s with message %w", c.Config.User.Email, config.Name, err)
	}

	legoConfig := lego.NewConfig(user)
	legoConfig.CADirURL = lego.LEDirectoryStaging
	legoConfig.Certificate.KeyType = certcrypto.RSA2048

	client, err = lego.NewClient(legoConfig)
	if err != nil {
		slog.Error("could not create lego client", "config", config.Name)
	}

	var environment registry.Environment
	environment, err = c.Config.GetEnvironment(config.AcmeRequest.Organization, config.AcmeRequest.Environment)
	if err != nil {
		slog.Debug("could not find organization environment for acme request", "organization", config.AcmeRequest.Organization, "environment", config.AcmeRequest.Environment)
		return nil, fmt.Errorf("could not find environment %s for organization %s for acme request with message %w", config.AcmeRequest.Environment, config.AcmeRequest.Organization, err)
	}

	var provider challenge.Provider
	provider, err = netscaleradc.NewGlobalHttpProvider(environment, 10, c.Timestamp)
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
		return nil, fmt.Errorf("could not register user %s for acme request with message %w", c.Config.User.Email, err)
	}
	user.Registration = reg

	request := certificate.ObtainRequest{
		Domains: config.AcmeRequest.GetDomains(),
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

func (c Daemon) getVipers(files []string) (map[string]*viper.Viper, error) {
	var (
		err    error
		vipers = make(map[string]*viper.Viper, len(files))
	)
	for _, file := range files {
		fileViper := viper.New()
		fileViper.SetConfigFile(file)
		err = fileViper.ReadInConfig()
		if err != nil {
			slog.Error("could not read config from file", "file", file, "error", err)
			continue
		}
		vipers[file] = fileViper
	}
	return vipers, nil
}

func (c Daemon) listConfigFiles(path string) ([]string, error) {
	var (
		err    error
		files  []os.DirEntry
		output []string
	)

	files, err = os.ReadDir(path)
	if err != nil {
		slog.Error("cannot list files in config directory", "error", err)
		return output, fmt.Errorf("cannot list files in config directory with message %w", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			output = append(output, path+"/"+file.Name())
		} else {
			var subDirFiles []string
			subDirFiles, err = c.listConfigFiles(path + "/" + file.Name())
			if err != nil {
				return output, err
			}
			output = append(output, subDirFiles...)
		}
	}
	return output, err
}
