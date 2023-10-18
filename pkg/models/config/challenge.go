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

// ACME Protocol challenge types
const (
	// NetScaler specific challenge types are defined in package netscaleradc
	ACME_CHALLENGE_TYPE_HTTP     = "http-01"
	ACME_CHALLENGE_TYPE_DNS      = "dns-01"
	ACME_CHALLENGE_TYPE_TLS_ALPN = "tls-alpn-01"
)

// Generic webserver provider
// Other providers constants are defined in their respective codebase
const (
	ACME_CHALLENGE_PROVIDER_WEBSERVER = "webserver"
)

type Challenge struct {
	Service                    string `json:"service" yaml:"service" mapstructure:"service"`
	Type                       string `json:"type" yaml:"type" mapstructure:"type"`
	Provider                   string `json:"provider" yaml:"provider" mapstructure:"provider"`
	DisableDnsPropagationCheck bool   `json:"disableDnsPropagationCheck" yaml:"disableDnsPropagationCheck" mapstructure:"disableDnsPropagationCheck"`
	ProviderParameters         string `json:"providerParameters" yaml:"providerParameters" mapstructure:"providerParameters"`
}
