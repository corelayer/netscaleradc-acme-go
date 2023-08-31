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
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"sync"

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
	Config config.Application
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

	wg := sync.WaitGroup{}
	for key, currentConfig := range configs {
		wg.Add(1)
		go func(k string, conf *viper.Viper) {
			var uConfig config.Certificate
			err = conf.Unmarshal(&uConfig)
			if err != nil {
				slog.Debug("could not unmarshal config", "config", k)
				// return fmt.Errorf("could not unmarshal config for %s with message %w", k, err)
			}
			slog.Info("certificate config loaded for processing", "name", uConfig.Name)
			err = c.launchLego(uConfig)
			if err != nil {
				slog.Debug("failed to process acme request", "name", uConfig.Name, "error", err)
				// return fmt.Errorf("failed to process acme request for %s with message %w", uConfig.Name, err)
			}
		}(key, currentConfig)
	}
	wg.Wait()
	return nil
}

func (c Daemon) launchLego(config config.Certificate) error {
	var (
		err    error
		user   *models.User
		client *lego.Client
	)
	user, err = models.NewUser(c.Config.User.Email)
	if err != nil {
		slog.Debug("could not create user", "email", c.Config.User.Email, "config", config.Name)
		return fmt.Errorf("could not create user %s for config %s with message %w", c.Config.User.Email, config.Name, err)
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
		return fmt.Errorf("could not find environment %s for organization %s for acme request with message %w", config.AcmeRequest.Environment, config.AcmeRequest.Organization, err)
	}

	var provider challenge.Provider
	provider, err = netscaleradc.NewGlobalHttpProvider(environment, 10)
	if err != nil {
		return err
	}
	err = client.Challenge.SetHTTP01Provider(provider)

	// New users will need to register
	var reg *registration.Resource
	reg, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		slog.Error("could not register user", "error", err)
		return fmt.Errorf("could not register user %s for acme request with message %w", c.Config.User.Email, err)
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
		return fmt.Errorf("could not obtain certificate with message %w", err)
	}

	fmt.Println(certificates.Certificate)
	return nil
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
