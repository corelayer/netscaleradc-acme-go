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

//
// import (
// 	"fmt"
// 	"log/slog"
// 	"sync"
//
// 	"github.com/corelayer/netscaleradc-nitro-go/pkg/registry"
// 	"github.com/go-acme/lego/v4/certificate"
// 	"github.com/go-acme/lego/v4/challenge"
// 	"github.com/go-acme/lego/v4/challenge/dns01"
// 	"github.com/go-acme/lego/v4/lego"
//
// 	"github.com/corelayer/netscaleradc-acme-go/pkg/models/config"
// )
//
// type CertificateProviderProcessor struct {
// 	parameters []config.ProviderParameters
// }
//
// func (p CertificateProviderProcessor) Start(providerName string, input <-chan config.Certificate, output chan<- config.Certificate, errors chan<- error, wg *sync.WaitGroup) {
// 	var (
// 		err error
// 	)
//
// 	defer wg.Done()
//
// 	slog.Debug("launching provider processor", "provider", providerName)
// 	for r := range input {
// 		slog.Debug("provider sequence started for certificate", "provider", providerName, "certificate", r.Name)
// 		r.Resource, err = p.processCertificate(r)
// 		if err != nil {
// 			errors <- fmt.Errorf("error occurred while processing request for certificate %s using provider %s with message: %w", r.Name, providerName, err)
// 			continue
// 		}
// 		output <- r
// 		slog.Debug("provider sequence completed for certificate", "provider", providerName, "certificate", r.Name)
// 	}
// 	slog.Debug("terminating provider processor", "provider", providerName)
// }
//
// func (p CertificateProviderProcessor) processCertificate(cert config.Certificate) (*certificate.Resource, error) {
// 	var (
// 		err     error
// 		client  *lego.Client
// 		domains []string
// 	)
// 	slog.Info("execute acme request for certificate", "certificate", cert.Name)
//
// 	client, err = l.getLegoClient(cert.Request.User, cert.Request.Challenge.Service, cert.Request.GetKeyType(), cert.Request.Timeout)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var environment registry.Environment
// 	environment, err = l.getEnvironment(cert.Request.Target)
// 	if err != nil {
// 		slog.Debug("could not find organization environment for acme request", "organization", cert.Request.Target.Organization, "environment", cert.Request.Target.Environment)
// 		return nil, fmt.Errorf("could not find environment %s for organization %s for acme request with message %w", cert.Request.Target.Environment, cert.Request.Target.Organization, err)
// 	}
//
// 	defer p.resetEnvironmentVariables(cert.Request.Challenge.ProviderParameters)
//
// 	if cert.Request.Challenge.ProviderParameters != "" {
// 		providerParams, err = l.getProviderParameters(cert.Request.Challenge.ProviderParameters)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		err = providerParams.ApplyEnvironmentVariables()
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
//
// 	err = p.configureChallengeProvider(client, cert.Request, environment, timestamp)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var request certificate.ObtainRequest
// 	request, err = p.getObtainRequest(cert.Name, cert.Request)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return client.Certificate.Obtain(request)
// }
//
// func (p CertificateProviderProcessor) applyEnvironmentVariables(params string) error {
// 	var providerParams config.ProviderParameters
// 	providerParams, err = p.configureEnvironmentVariables(cert.Request.Challenge)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = providerParams.ApplyEnvironmentVariables()
// 	if err != nil {
// 		return nil, err
// 	}
// }
//
// func (p CertificateProviderProcessor) resetEnvironmentVariables(params string) {
// 	var providerParams config.ProviderParameters
// 	providerParams, err = p.configureEnvironmentVariables(cert.Request.Challenge)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = providerParams.ApplyEnvironmentVariables()
// 	if err != nil {
// 		return nil, err
// 	}
// }
//
// func (p CertificateProviderProcessor) configureEnvironmentVariables(c config.Challenge) (config.ProviderParameters, error) {
// 	var (
// 		err            error
// 		providerParams config.ProviderParameters
// 	)
// 	if c.ProviderParameters != "" {
// 		providerParams, err = l.getProviderParameters(c.ProviderParameters)
// 		if err != nil {
// 			return providerParams, err
// 		}
//
// 		err = providerParams.ApplyEnvironmentVariables()
// 		if err != nil {
// 			return providerParams, err
// 		}
// 	}
// }
//
// func (p CertificateProviderProcessor) configureChallengeProvider(c *lego.Client, r config.Request, e registry.Environment, timestamp string) error {
// 	var (
// 		err      error
// 		provider challenge.Provider
// 	)
//
// 	provider, err = r.GetChallengeProvider(e, timestamp)
// 	if err != nil {
// 		return err
// 	}
//
// 	switch r.Challenge.Type {
// 	case config.ACME_CHALLENGE_TYPE_HTTP:
// 		err = c.Challenge.SetHTTP01Provider(provider)
// 	case config.ACME_CHALLENGE_TYPE_DNS:
// 		if r.Challenge.DisableDnsPropagationCheck {
// 			err = c.Challenge.SetDNS01Provider(provider, dns01.DisableCompletePropagationRequirement())
// 		} else {
// 			err = c.Challenge.SetDNS01Provider(provider)
// 		}
// 	case config.ACME_CHALLENGE_TYPE_TLS_ALPN:
// 		err = c.Challenge.SetTLSALPN01Provider(provider)
// 	default:
// 		err = fmt.Errorf("invalid challenge type")
// 	}
// 	return err
// }
//
// func (p CertificateProviderProcessor) getObtainRequest(name string, r config.Request) (certificate.ObtainRequest, error) {
// 	var (
// 		err     error
// 		domains []string
// 	)
// 	// Get domains for ACME request
// 	if domains, err = r.GetDomains(); err != nil {
// 		slog.Debug("invalid domain in request", "certificate", name, "error", err)
// 		return certificate.ObtainRequest{}, fmt.Errorf("invalid domain in request for certificate %s with message: %w", name, err)
// 	}
//
// 	// Execute ACME request
// 	request var providerParams config.ProviderParameters
// 	providerParams, err = p.configureEnvironmentVariables(cert.Request.Challenge)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = providerParams.ApplyEnvironmentVariables()
// 	if err != nil {
// 		return nil, err
// 	}:= certificate.ObtainRequest{
// 		Domains: domains,
// 		Bundle:  true,
// 	}
//
// 	return request, nil
// }
