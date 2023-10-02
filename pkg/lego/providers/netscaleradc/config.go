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

package netscaleradc

import (
	"log/slog"

	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro"
	"github.com/go-acme/lego/v4/platform/config/env"
)

const (
	envNamespace = "NETSCALERADC_"

	EnvName                      = envNamespace + "NAME"
	EnvAddress                   = envNamespace + "ADDRESS"
	EnvUsername                  = envNamespace + "USER"
	EnvPassword                  = envNamespace + "PASS"
	EnvUseSsl                    = envNamespace + "USE_SSL"
	EnvValidateServerCertificate = envNamespace + "VALIDATE_SERVER_CERTIFICATE"
	EnvTimeout                   = envNamespace + "TIMEOUT"
)

type Config struct {
	Name                      string
	Address                   string
	username                  string
	password                  string
	useSsl                    bool
	validateServerCertificate bool
	timeout                   int
}

func (c Config) GetClient() (*nitro.Client, error) {
	return nitro.NewClient(c.Name, c.Address, c.getCredentials(), c.getConnectionSettings())
}

func (c Config) getCredentials() nitro.Credentials {
	return nitro.Credentials{
		Username: c.username,
		Password: c.password,
	}
}

func (c Config) getConnectionSettings() nitro.ConnectionSettings {
	return nitro.ConnectionSettings{
		UseSsl:                    c.useSsl,
		Timeout:                   c.timeout,
		UserAgent:                 "",
		ValidateServerCertificate: c.validateServerCertificate,
		LogTlsSecrets:             false,
		LogTlsSecretsDestination:  "",
		AutoLogin:                 false,
	}
}

func NewConfig() (*Config, error) {
	var (
		err    error
		values map[string]string
	)

	values, err = env.Get(EnvName, EnvAddress, EnvUsername, EnvPassword)
	if err != nil {
		slog.Error("could not get environment variables", "error", err)
		return nil, err
	}

	return &Config{
		Name:                      values[EnvName],
		Address:                   values[EnvAddress],
		username:                  values[EnvUsername],
		password:                  values[EnvPassword],
		useSsl:                    env.GetOrDefaultBool(EnvUseSsl, true),
		validateServerCertificate: env.GetOrDefaultBool(EnvValidateServerCertificate, true),
		timeout:                   env.GetOrDefaultInt(EnvTimeout, 5000),
	}, nil
}
