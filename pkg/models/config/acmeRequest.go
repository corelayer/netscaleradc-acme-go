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

const (
	// NetScaler specific challenge types are defined in package netscaleradc
	ACME_CHALLENGE_TYPE_HTTP = "http"
	ACME_CHALLENGE_TYPE_DNS  = "dns"
)

const (
	ACME_SERVICE_LETSENCRYPT_PRODUCTION     = "LE_PRODUCTION"
	ACME_SERVICE_LETSENCRYPT_PRODUCTION_URL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	ACME_SERVICE_LETSENCRYPT_STAGING        = "LE_STAGING"
	ACME_SERVICE_LETSENCRYPT_STAGING_URL    = "https://acme-staging-v02.api.letsencrypt.org/directory"
)

type AcmeRequest struct {
	Organization            string   `json:"organization" yaml:"organization" mapstructure:"organization"`
	Environment             string   `json:"environment" yaml:"environment" mapstructure:"environment"`
	Username                string   `json:"username" yaml:"username" mapstructure:"username"`
	ChallengeService        string   `json:"service" yaml:"service" mapstructure:"service"`
	ChallengeType           string   `json:"type" yaml:"type" mapstructure:"type"`
	KeyType                 string   `json:"keytype" yaml:"keyType" mapstructure:"keyType"`
	CommonName              string   `json:"commonName" yaml:"commonName" mapstructure:"commonName"`
	SubjectAlternativeNames []string `json:"subjectAlternativeNames" yaml:"subjectAlternativeNames" mapstructure:"subjectAlternativeNames"`
}

func (r AcmeRequest) GetDomains() []string {
	var output []string
	output = append(output, r.CommonName)
	output = append(output, r.SubjectAlternativeNames...)
	return output
}

func (r AcmeRequest) GetServiceUrl() string {
	switch r.ChallengeService {
	case ACME_SERVICE_LETSENCRYPT_PRODUCTION:
		return ACME_SERVICE_LETSENCRYPT_PRODUCTION_URL
	case ACME_SERVICE_LETSENCRYPT_STAGING:
		return ACME_SERVICE_LETSENCRYPT_STAGING_URL
	default:
		return r.ChallengeService
	}
}

func (r AcmeRequest) GetChallengeProvider(environment registry.Environment, timestamp string) (challenge.Provider, error) {
	switch r.ChallengeType {
	case netscaleradc.ACME_CHALLENGE_TYPE_NETSCALER_HTTP_GLOBAL:
		return netscaleradc.NewGlobalHttpProvider(environment, 10, timestamp)
	case netscaleradc.ACME_CHALLENGE_TYPE_NETSCALER_ADNS:
		return netscaleradc.NewADnsProvider(environment, 10)
	case ACME_CHALLENGE_TYPE_HTTP:
		return http01.NewProviderServer("", "12346"), nil
	default:
		return dns.NewDNSChallengeProviderByName(r.ChallengeType)
	}
}

func (r AcmeRequest) GetKeyType() certcrypto.KeyType {
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
