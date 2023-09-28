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

package config

import (
	"github.com/corelayer/netscaleradc-nitro-go/pkg/registry"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/providers/dns"

	"github.com/corelayer/netscaleradc-acme-go/pkg/lego/providers/netscaleradc"
)

const (
	ACME_KEY_TYPE_EC256   = "EC256"
	ACME_KEY_TYPE_EC384   = "EC384"
	ACME_KEY_TYPE_RSA2048 = "RSA2048"
	ACME_KEY_TYPE_RSA4096 = "RSA4096"
	ACME_KEY_TYPE_RSA8192 = "RSA8192"
)

type Request struct {
	Target    Target    `json:"target" yaml:"target" mapstructure:"target"`
	AcmeUser  string    `json:"acmeUser" yaml:"acmeUser" mapstructure:"acmeUser"`
	Challenge Challenge `json:"challenge" yaml:"challenge" mapstructure:"challenge"`
	KeyType   string    `json:"keyType" yaml:"keyType" mapstructure:"keyType"`
	Content   Content   `json:"content" yaml:"content" mapstructure:"content"`
	basePath  string
}

func (r Request) GetServiceUrl() string {
	switch r.Challenge.Service {
	case ACME_SERVICE_LETSENCRYPT_PRODUCTION:
		return ACME_SERVICE_LETSENCRYPT_PRODUCTION_URL
	case ACME_SERVICE_LETSENCRYPT_STAGING:
		return ACME_SERVICE_LETSENCRYPT_STAGING_URL
	default:
		return r.Challenge.Service
	}
}

func (r Request) GetChallengeProvider(environment registry.Environment, timestamp string) (challenge.Provider, error) {
	switch r.Challenge.Provider {
	case netscaleradc.ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL:
		return netscaleradc.NewGlobalHttpProvider(environment, 10, timestamp)
	case netscaleradc.ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS:
		return netscaleradc.NewADnsProvider(environment, 10)
	case ACME_CHALLENGE_PROVIDER_WEBSERVER:
		return http01.NewProviderServer("", "12346"), nil
	default:
		return dns.NewDNSChallengeProviderByName(r.Challenge.Type)
	}
}

func (r Request) GetDomains() ([]string, error) {
	return r.Content.GetDomains(r.basePath)
}

func (r Request) GetKeyType() certcrypto.KeyType {
	switch r.KeyType {
	case ACME_KEY_TYPE_EC256:
		return certcrypto.EC256
	case ACME_KEY_TYPE_EC384:
		return certcrypto.EC384
	case ACME_KEY_TYPE_RSA2048:
		return certcrypto.RSA2048
	case ACME_KEY_TYPE_RSA4096:
		return certcrypto.RSA4096
	case ACME_KEY_TYPE_RSA8192:
		return certcrypto.RSA8192
	default:
		return certcrypto.RSA4096
	}
}

func (r Request) SetPath(path string) Request {
	r.basePath = path
	return r
}
