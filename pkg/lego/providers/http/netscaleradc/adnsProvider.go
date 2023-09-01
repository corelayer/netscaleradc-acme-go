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
	"fmt"
	"log/slog"

	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro"
	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro/resource/config"
	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro/resource/controllers"
	"github.com/corelayer/netscaleradc-nitro-go/pkg/registry"
)

type ADnsProvider struct {
	nitroClient *nitro.Client
	dnsTxt      *controllers.DnsTxtRecController

	maxRetries int
}

// NewADnsProvider returns a HTTPProvider instance with a configured list of hosts
func NewADnsProvider(environment registry.Environment, maxRetries int, timestamp string) (*ADnsProvider, error) {
	c := &ADnsProvider{
		maxRetries: maxRetries,
	}

	return c, c.initialize(environment)
}

// Present the ACME challenge to the provider.
// Parameter domain is the fqdn for which the challenge will be provided
// Parameter endpoint is the path to which ACME will look  for the challenge (/.well-known/acme-challenge/<token>)
// Parameter keyAuth is the value which must be returned for a successful challenge
func (p *ADnsProvider) Present(domain string, token string, keyAuth string) error {
	var err error
	slog.Info("prepare acme request", "domain", domain)

	slog.Debug("adding dns record", "domain", domain, "action")
	if _, err = p.dnsTxt.Add(domain, []string{token}, 300); err != nil {
		slog.Error("could not create dns record for acme challenge", "domain", domain)
		return fmt.Errorf("could not create dns record for acme challenge for domain %s with error %w", domain, err)
	}

	slog.Debug("prepare acme request completed", "domain", domain)
	return nil
}

func (p *ADnsProvider) CleanUp(domain string, token string, keyAuth string) error {
	var err error
	slog.Info("cleanup acme request", "domain", domain)

	slog.Debug("removing dns record", "domain", domain, "action")
	var d *nitro.Response[config.DnsTxtRec]
	if d, err = p.dnsTxt.Get(domain, []string{"recordid"}); err != nil {
		slog.Error("could not get recordid for dns record for acme challenge", "domain", domain)
		return fmt.Errorf("could not get recordid dns record for acme challenge for domain %s with error %w", domain, err)

	}

	for _, rec := range d.Data {
		for _, data := range rec.Data {
			if data != keyAuth {
				continue
			}
			// Can multiple records exist?
			if _, err = p.dnsTxt.Delete(domain, d.Data[0].RecordId); err != nil {
				slog.Error("could not delete dns record for acme challenge", "domain", domain)
				return fmt.Errorf("could not delete dns record for acme challenge for domain %s with error %w", domain, err)
			}
			break
		}
	}

	slog.Debug("cleanup acme request completed", "domain", domain)
	return nil
}

func (p *ADnsProvider) initialize(e registry.Environment) error {
	slog.Debug("initialize nitro client for primary node for environment", "environment", e.Name)
	client, err := e.GetPrimaryNitroClient()
	if err != nil {
		return fmt.Errorf("failed to initialize ADnsProvider with error %w", err)
	}

	slog.Debug("initialize nitro controllers for responder functionality")
	p.nitroClient = client
	p.dnsTxt = controllers.NewDnsTxtRecController(client)

	return nil
}
