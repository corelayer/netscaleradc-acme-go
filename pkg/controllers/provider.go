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
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/corelayer/netscaleradc-nitro-go/pkg/registry"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"

	"github.com/corelayer/netscaleradc-acme-go/pkg/models"
	"github.com/corelayer/netscaleradc-acme-go/pkg/models/config"
)

type Provider struct {
	Name   string
	Input  chan models.Certificate
	Output chan models.Certificate

	provider challenge.Provider
}

func NewProvider(name string) *Provider {
	p := Provider{
		Name:   name,
		Input:  make(chan models.Certificate),
		Output: make(chan models.Certificate),
	}
	return &p
}

func (p *Provider) Run(ctx context.Context) {
	slog.Debug("Starting provider", "provider", p.Name)

	var err error
	for {
		select {
		case c, ok := <-p.Input:
			if !ok {
				return
			}
			c, err = p.process(c)
			if err != nil {
				// TODO Run error handling
				continue
			}
			p.Output <- c
		case <-ctx.Done():
			p.Stop()
			return
		default:
			slog.Debug("Waiting for input", "provider", p.Name)
			time.Sleep(500 * time.Millisecond)
		}

	}
}

func (p *Provider) Stop() {
	slog.Debug("Stopping provider", "provider", p.Name)
	close(p.Input)
	// TODO Stop close output channel?
}

func (p *Provider) process(c models.Certificate) (models.Certificate, error) {
	slog.Debug("Processing certificate", "provider", p.Name, "certificate", c.Name, "status", "started")

	var (
		err    error
		client *lego.Client
	)

	err = c.ProviderConfig.ApplyEnvironmentVariables()
	if err != nil {
		slog.Warn("Could not apply environment variables", "provider", p.Name, "certificate", c.Name, "status", "error")
		return c, err
	}

	defer func() {
		err = c.ProviderConfig.ResetEnvironmentVariables()
	}()
	if err != nil {
		slog.Warn("Could not reset environment variables", "provider", p.Name, "certificate", c.Name, "status", "error")
		return c, err
	}

	client, err = p.configureAcmeClient()
	if err != nil {
		slog.Warn("Could not configure acme client", "provider", p.Name, "certificate", c.Name, "status", "error")
		return c, err
	}

	c.Data, err = p.executeAcmeRequest(client, c.Request.Domains)
	if err != nil {
		slog.Warn("Could not request certificate", "provider", p.Name, "certificate", c.Name, "status", "error")
		return c, err
	}

	slog.Debug("Processing certificate", "provider", p.Name, "certificate", c.Name, "status", "completed")
	return c, nil
}

func (p *Provider) configureAcmeClient() (*lego.Client, error) {
	return nil, nil
}

func (p *Provider) executeAcmeRequest(c *lego.Client, domains []string) (models.CertificateData, error) {
	var (
		err error
		req certificate.ObtainRequest
		res *certificate.Resource
	)
	req = certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	res, err = c.Certificate.Obtain(req)
	if err != nil {
		return models.CertificateData{}, err
	}

	data := models.CertificateData{
		PublicKey:  res.Certificate,
		PrivateKey: res.PrivateKey,
		Issuer:     res.IssuerCertificate,
	}
	return data, nil
}

func (l Launcher) executeAcmeRequest2(cert config.Certificate) (*certificate.Resource, error) {
	var (
		err    error
		client *lego.Client
	)
	slog.Info("execute acme request for certificate", "certificate", cert.Name)

	client, err = l.getLegoClient(cert.Request.User, cert.Request.Challenge.Service, cert.Request.GetKeyType(), cert.Request.Timeout)
	if err != nil {
		return nil, err
	}

	var environment registry.Environment
	environment, err = l.getEnvironment(cert.Request.Target)
	if err != nil {
		slog.Debug("could not find organization environment for acme request", "organization", cert.Request.Target.Organization, "environment", cert.Request.Target.Environment)
		return nil, fmt.Errorf("could not find environment %s for organization %s for acme request with message %w", cert.Request.Target.Environment, cert.Request.Target.Organization, err)
	}

	var provider challenge.Provider
	provider, err = cert.Request.GetChallengeProvider(environment, l.timestamp)
	if err != nil {
		return nil, err
	}

	switch cert.Request.Challenge.Type {
	case config.ACME_CHALLENGE_TYPE_HTTP:
		err = client.Challenge.SetHTTP01Provider(provider)
	case config.ACME_CHALLENGE_TYPE_DNS:
		if cert.Request.Challenge.DisableDnsPropagationCheck {
			err = client.Challenge.SetDNS01Provider(provider, dns01.DisableCompletePropagationRequirement())
		} else {
			err = client.Challenge.SetDNS01Provider(provider)
		}
	case config.ACME_CHALLENGE_TYPE_TLS_ALPN:
		err = client.Challenge.SetTLSALPN01Provider(provider)
	default:
		err = fmt.Errorf("invalid challenge type")
	}
	if err != nil {
		return nil, err
	}
}

func (l Launcher) getLegoClient2(username string, service string, keyType certcrypto.KeyType, timeout int) (*lego.Client, error) {
	var (
		err            error
		url            string
		account        *models.Account
		requestTimeout time.Duration
		client         *lego.Client
	)

	l.registrationMutex.Lock()
	slog.Debug("locking for acme user account validation", "user", username, "service", service)
	url, err = l.getServiceUrl(service)
	if err != nil {
		slog.Debug("could not find service", "service", service)
		return nil, fmt.Errorf("could not find user %s for service %s with message: %w", username, url, err)
	}

	account, err = l.getAccount(username, url)
	if err != nil {
		slog.Debug("could not find user", "username", username, "service", url)
		return nil, fmt.Errorf("could not find user %s for service %s with message: %w", username, url, err)
	}

	requestTimeout, err = time.ParseDuration(strconv.Itoa(timeout) + "s")
	legoConfig := lego.NewConfig(*account)
	legoConfig.CADirURL = url
	legoConfig.Certificate.KeyType = keyType
	legoConfig.Certificate.Timeout = requestTimeout

	client, err = lego.NewClient(legoConfig)
	if err != nil {
		slog.Debug("could not create lego client", "username", username)
		return nil, fmt.Errorf("could not create lego client for user %s with error %w", username, err)
	}

	// Query registration
	if account.GetRegistration() == nil {
		slog.Debug("register acme account for user", "username", username, "service", url)
		// New users will need to register
		var reg *registration.Resource
		var emptyEab = config.ExternalAccountBinding{}

		if account.ExternalAccountBinding == emptyEab {
			reg, err = client.Registration.Register(
				registration.RegisterOptions{
					TermsOfServiceAgreed: true,
				})
		} else {
			reg, err = client.Registration.RegisterWithExternalAccountBinding(
				registration.RegisterEABOptions{
					TermsOfServiceAgreed: true,
					Kid:                  account.ExternalAccountBinding.Kid,
					HmacEncoded:          account.ExternalAccountBinding.HmacEncoded,
				})
		}
		if err != nil {
			slog.Debug("could not register acme account for user", "username", username, "service", url, "error", err)
			return nil, fmt.Errorf("could not register user %s for acme request on service %s with message: %w", username, url, err)
		}
		account.Registration = reg
	}
	l.registrationMutex.Unlock()
	slog.Debug("unlocking for acme user account validation", "user", username, "service", url)

	return client, nil
}
