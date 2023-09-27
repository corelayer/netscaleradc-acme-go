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
	"github.com/go-acme/lego/v4/challenge/dns01"
)

const (
	ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS = "netscaler-adns"
)

// ADnsProvider manages ACME requests for NetScaler ADC Authoritative DNS service
type ADnsProvider struct {
	nitroClient *nitro.Client
	dnsTxtRec   *controllers.DnsTxtRecController

	maxRetries int
}

// NewADnsProvider returns a HTTPProvider instance with a configured list of hosts
func NewADnsProvider(environment registry.Environment, maxRetries int) (*ADnsProvider, error) {
	c := &ADnsProvider{
		maxRetries: maxRetries,
	}

	return c, c.initialize(environment)
}

// Present the ACME challenge to the provider.
// domain is the fqdn for which the challenge will be provided
// Parameter endpoint is the path to which ACME will look  for the challenge (/.well-known/acme-challenge/<token>)
// Parameter keyAuth is the value which must be returned for a successful challenge
func (p *ADnsProvider) Present(domain string, token string, keyAuth string) error {
	var err error
	slog.Info("ns acme request", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS, "domain", domain)

	// Get challenge information to
	info := dns01.GetChallengeInfo(domain, keyAuth)

	// Add DNS record to ADNS zone on NetScaler ADC
	slog.Debug("ns acme request: create dns record", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS, "domain", domain)
	if _, err = p.dnsTxtRec.Add(info.FQDN, []string{info.Value}, 30); err != nil {
		slog.Error("ns acme request: could not create dns record", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS, "domain", domain, "error", err)
		return fmt.Errorf("ns acme request: could not create dns record %s: %w", domain, err)
	}

	slog.Debug("ns acme request: completed", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS, "domain", domain)
	return nil
}

func (p *ADnsProvider) CleanUp(domain string, token string, keyAuth string) error {
	var err error
	slog.Info("ns acme cleanup", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS, "domain", domain)

	// Get DNS01 Challenge info
	info := dns01.GetChallengeInfo(domain, keyAuth)

	slog.Debug("ns acme cleanup: remove dns record", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS, "domain", domain)
	var res *nitro.Response[config.DnsTxtRec]
	// Limit data transfer by limiting returned fields
	if res, err = p.dnsTxtRec.Get(info.FQDN, []string{"string", "recordid"}); err != nil {
		slog.Error("ns acme cleanup: could not get recordid", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS, "domain", domain, "error", err)
		return fmt.Errorf("ns acme cleanup: could not get recordid %s: %w", domain, err)

	}

	for _, rec := range res.Data {
		// Loop over array of returned records
		slog.Debug("ns acme cleanup: processing record", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS, "domain", domain, "recordid", rec.RecordId)
		for _, data := range rec.Data {
			// Only remove record if keyAuth matches the current acme request
			slog.Debug("ns acme cleanup: processing record data", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS, "domain", domain)
			if data != info.Value {
				slog.Debug("ns acme cleanup: skipping record", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS, "domain", domain)
				continue
			}

			slog.Debug("ns acme cleanup: found record to remove", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS, "domain", domain)
			if _, err = p.dnsTxtRec.Delete(info.FQDN, rec.RecordId); err != nil {
				slog.Error("ns acme cleanup: could not remove dns record", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS, "domain", domain, "error", err)
				return fmt.Errorf("ns acme cleanup: could not remove dns record %s: %w", domain, err)
			}
			slog.Debug("ns acme cleanup: removed dns record", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS, "domain", domain)
		}
	}

	slog.Debug("ns acme cleanup: completed", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS, "domain", domain)
	return nil
}

func (p *ADnsProvider) initialize(e registry.Environment) error {
	slog.Debug("ns acme adns provider initialization", "environment", e.Name)
	client, err := e.GetPrimaryNitroClient()
	if err != nil {
		slog.Error("ns acme adns provider initialization failed", "error", err)
		return fmt.Errorf("ns acme adns provider initialization failed: %w", err)
	}

	p.nitroClient = client
	p.dnsTxtRec = controllers.NewDnsTxtRecController(client)
	slog.Debug("ns acme adns provider initialization completed")
	return nil
}
