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

package models

import (
	"os"

	"github.com/go-acme/lego/v4/certcrypto"
)

type Certificate struct {
	Name           string
	ProviderConfig ProviderConfig
	Request        CertificateRequest
	Data           CertificateData
}

type CertificateData struct {
	PublicKey  []byte
	PrivateKey []byte
	Issuer     []byte
}

type CertificateRequest struct {
	Domains []string
	KeyType certcrypto.KeyType
}

type ProviderConfig struct {
	EnvironmentVariables []EnvironmentVariable
}

func (c ProviderConfig) ApplyEnvironmentVariables() error {
	var err error

	if c.EnvironmentVariables == nil {
		return nil
	}

	for _, v := range c.EnvironmentVariables {
		err = os.Setenv(v.Name, v.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c ProviderConfig) ResetEnvironmentVariables() error {
	var err error

	if c.EnvironmentVariables == nil {
		return nil
	}

	for _, v := range c.EnvironmentVariables {
		err = os.Unsetenv(v.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

type EnvironmentVariable struct {
	Name  string
	Value string
}
