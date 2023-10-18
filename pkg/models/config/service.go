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

// Let's Encrypt Service Environments
const (
	ACME_SERVICE_LETSENCRYPT_PRODUCTION_NAME = "LE_PRODUCTION"
	ACME_SERVICE_LETSENCRYPT_PRODUCTION_URL  = "https://acme-v02.api.letsencrypt.org/directory"
	ACME_SERVICE_LETSENCRYPT_STAGING_NAME    = "LE_STAGING"
	ACME_SERVICE_LETSENCRYPT_STAGING_URL     = "https://acme-staging-v02.api.letsencrypt.org/directory"
)

var (
	ACME_SERVICE_LETSENCRYPT_PRODUCTION = Service{
		Name: ACME_SERVICE_LETSENCRYPT_PRODUCTION_NAME,
		Url:  ACME_SERVICE_LETSENCRYPT_PRODUCTION_URL,
	}

	ACME_SERVICE_LETSENCRYPT_STAGING = Service{
		Name: ACME_SERVICE_LETSENCRYPT_STAGING_NAME,
		Url:  ACME_SERVICE_LETSENCRYPT_STAGING_URL,
	}
)

type Service struct {
	Name string `json:"name" yaml:"name" mapstructure:"name"`
	Url  string `json:"email" yaml:"email" mapstructure:"email"`
}
