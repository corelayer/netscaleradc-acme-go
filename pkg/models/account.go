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
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	"github.com/go-acme/lego/v4/registration"

	"github.com/corelayer/netscaleradc-acme-go/pkg/models/config"
)

type Account struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey

	ExternalAccountBinding config.ExternalAccountBinding
}

func (a Account) GetEmail() string {
	return a.Email
}
func (a Account) GetRegistration() *registration.Resource {
	return a.Registration
}
func (a Account) GetPrivateKey() crypto.PrivateKey {
	return a.key
}

func NewAccount(email string, eab config.ExternalAccountBinding) (*Account, error) {
	key, err := createPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("could not create new private key for user %s with message %w", email, err)
	}
	return &Account{
		Email:                  email,
		key:                    key,
		ExternalAccountBinding: eab,
	}, nil
}

func createPrivateKey() (*ecdsa.PrivateKey, error) {
	// Create a user. New accounts need an email and private key to start.
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}
