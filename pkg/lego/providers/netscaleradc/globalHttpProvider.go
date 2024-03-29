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
	"time"

	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro"
	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro/resource/config"
	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro/resource/controllers"
	"github.com/corelayer/netscaleradc-nitro-go/pkg/registry"
)

const (
	ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL = "netscaler-http-global"
)

// GlobalHttpProvider manages ACME requests for NetScaler ADC using globally bound responder policies
type GlobalHttpProvider struct {
	client         *nitro.Client
	rsaController  *controllers.ResponderActionController
	rspController  *controllers.ResponderPolicyController
	rspbController *controllers.ResponderGlobalResponderPolicyBindingController

	rspbBindtype string
	rsaPrefix    string
	rspPrefix    string
	timestamp    string

	maxRetries int
}

// NewGlobalHttpProvider returns a HTTPProvider instance with a configured list of hosts
func NewGlobalHttpProvider(e registry.Environment, maxRetries int, timestamp string) (*GlobalHttpProvider, error) {
	var (
		err error
		c   *nitro.Client
		p   *GlobalHttpProvider
	)

	slog.Debug("ns acme provider: initialize from configuration", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "environment", e.Name)
	c, err = e.GetPrimaryNitroClient()
	if err != nil {
		slog.Error("ns acme provider: client initialization from configuration failed", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "environment", e.Name, "client", c.Name, "error", err)
		return nil, fmt.Errorf("ns acme %s provider initialization from configuration failed: %w", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, err)
	}

	p = &GlobalHttpProvider{
		client:     c,
		maxRetries: maxRetries,
		timestamp:  timestamp,
	}
	p.initialize()

	slog.Debug("ns acme provider: initialization from configuration completed", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "environment", e.Name)
	return p, nil
}

// NewGlobalHttpProvider returns an HTTPProvider instance from environment variable settings
func NewGlobalHttpProviderFromEnv(maxRetries int, timestamp string) (*GlobalHttpProvider, error) {
	var (
		err error
		c   *Config
		n   *nitro.Client
		p   *GlobalHttpProvider
	)

	slog.Debug("ns acme provider: initialize from configuration", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "environment", "os")
	c, err = NewConfig()
	if err != nil {
		slog.Error("ns acme provider: client initialization from environment failed", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS, "environment", "os", "client", c.Name, "error", err)
		return nil, err
	}

	n, err = c.GetClient()
	if err != nil {
		slog.Error("ns acme provider: initialization from environment failed", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS, "environment", "os", "error", err)
		return nil, err
	}

	p = &GlobalHttpProvider{
		client:     n,
		maxRetries: maxRetries,
		timestamp:  timestamp,
	}
	p.initialize()

	slog.Debug("ns acme provider: initialization from environment completed", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_ADNS, "environment", "os")
	return p, nil
}

// Present the ACME challenge to the provider before validation
//
//	domain is the fqdn for which the challenge will be provided
//	token is the path to which ACME will look  for the challenge (/.well-known/acme-challenge/<token>)
//	keyAuth is the value which must be returned for a successful challenge
func (p *GlobalHttpProvider) Present(domain string, token string, keyAuth string) error {
	var err error
	slog.Info("ns acme request: start", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain)

	rsaActionName := p.getResponderActionName(domain)
	rspPolicyName := p.getResponderPolicyName(domain)
	rsaAction := "\"HTTP/1.1 200 OK\\r\\n\\r\\n" + keyAuth + "\""
	rspRule := "HTTP.REQ.HOSTNAME.EQ(\"" + domain + "\") && HTTP.REQ.URL.EQ(\"/.well-known/acme-challenge/" + token + "\")"

	// Create responder action
	slog.Debug("ns acme request: create responder action", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain, "resource", rsaActionName)
	if _, err = p.rsaController.Add(rsaActionName, "respondwith", rsaAction); err != nil {
		slog.Error("ns acme request: could not create responder action", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain, "resource", rsaActionName)
		return fmt.Errorf("ns acme request: could not create responder action %s for %s: %w", rsaActionName, domain, err)
	}

	// Create responder policy
	slog.Debug("ns acme request: create responder policy", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain, "resource", rspPolicyName)
	if _, err = p.rspController.Add(rspPolicyName, rspRule, rsaActionName, ""); err != nil {
		slog.Error("ns acme request: could not create responder policy", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain, "resource", rspPolicyName)
		return fmt.Errorf("ns acme request: could not create responder policy %s for %s: %w", rspPolicyName, domain, err)
	}

	// Bind responder policy to global REQ_OVERRIDE
	// We need REQ_OVERRIDE, otherwise responder policies bound to a csvserver/lbvserver get a higher priority
	if err = p.bindResponderPolicy(domain); err != nil {
		slog.Error("ns acme request: could not bind global responder policy", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain, "resource", rspPolicyName)
		return fmt.Errorf("ns acme request: could not bind global responder policy %s for %s: %w", rspPolicyName, domain, err)
	}

	slog.Debug("ns acme request: completed", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain)
	return nil
}

// CleanUp the ACME challenge on the provider after validation
//
//	domain is the fqdn for which the challenge will be provided
//	token is the path to which ACME will look  for the challenge (/.well-known/acme-challenge/<token>)
//	keyAuth is the value which must be returned for a successful challenge
func (p *GlobalHttpProvider) CleanUp(domain string, token string, keyAuth string) error {
	var err error
	slog.Info("ns acme cleanup: start", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain)

	rspPolicyName := p.getResponderPolicyName(domain)
	rsaActionName := p.getResponderActionName(domain)

	// Unbind responder policy from global REQ_OVERRIDE
	slog.Debug("ns acme cleanup: unbind global responder policy", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain, "resource", rspPolicyName)
	if _, err = p.rspbController.Delete(rspPolicyName, p.rspbBindtype); err != nil {
		slog.Error("ns acme cleanup: could not unbind global responder policy", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain, "resource", rspPolicyName)
		return fmt.Errorf("ns acme cleanup: could not unbind global responder policy %s for %s: %w", rspPolicyName, domain, err)
	}

	slog.Debug("ns acme cleanup: remove responder policy", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain, "resource", rspPolicyName)
	if _, err = p.rspController.Delete(rspPolicyName); err != nil {
		slog.Error("ns acme cleanup: could not remove responder policy", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain, "resource", rspPolicyName)
		return fmt.Errorf("ns acme cleanup: could not remove responder policy %s for %s: %w", rspPolicyName, domain, err)
	}

	slog.Debug("ns acme cleanup: remove responder action", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain, "resource", rsaActionName)
	if _, err = p.rsaController.Delete(rsaActionName); err != nil {
		slog.Error("ns acme cleanup: could not remove responder action", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain, "resource", rsaActionName)
		return fmt.Errorf("ns acme cleanup: could not remove responder action %s for %s: %w", rsaActionName, domain, err)
	}

	slog.Debug("ns acme cleanup: completed", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain)
	return nil
}

// bindResponderPolicy will bind the responder policy globally on NetScaler
func (p *GlobalHttpProvider) bindResponderPolicy(domain string) error {
	var (
		successfullyBoundPolicy = false
		retries                 = 0
		err                     error
		priority                string
		rspPolicyName           = p.getResponderPolicyName(domain)
	)

	for !successfullyBoundPolicy {
		slog.Debug("ns acme request: search for valid binding priority", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain, "resource", rspPolicyName)

		retries += 1
		priority, err = p.getPriority()
		if err != nil {
			slog.Error("ns acme request: could not find valid policy binding priority", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain, "error", err)
			return fmt.Errorf("ns acme request: could not find valid policy binding priority for %s: %w", domain, err)
		}

		if _, err = p.rspbController.Add(rspPolicyName, p.rspbBindtype, priority, "END"); err != nil {
			if retries >= p.maxRetries {
				slog.Error("ns acme request: exceeded max retries to bind global responder policy", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "domain", domain, "resource", rspPolicyName)
				return fmt.Errorf("ns acme request: exceeded max retries to bind global responder policy %s for %s: %w", rspPolicyName, domain, err)
			}
			// If the attempt to bind the policy at the current priority fails, continue to the next iteration to increase the priority
			continue
		}
		// The binding completed successfully, exit the loop
		successfullyBoundPolicy = true
	}
	return nil
}

// getPolicyBindingPriorities will get all global responder binding priorities currently in use on NetScaler
func (p *GlobalHttpProvider) getPolicyBindingPriorities() ([]string, error) {
	var (
		err      error
		output   []string
		bindings *nitro.Response[config.ResponderGlobalResponderPolicyBinding]
	)
	slog.Debug("ns acme request: retrieve existing priorities", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL)

	// Create custom Nitro Request
	// Limit data transfer by limiting returned fields
	nitroRequest := &nitro.Request[config.ResponderGlobalResponderPolicyBinding]{
		Arguments: map[string]string{
			"type": p.rspbBindtype,
		},
		Attributes: []string{"priority"},
	}

	// Execute Nitro Request
	bindings, err = nitro.ExecuteNitroRequest[config.ResponderGlobalResponderPolicyBinding](p.client, nitroRequest)
	if err != nil {
		slog.Error("ns acme request: could not retrieve existing priorities", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL)
		return nil, fmt.Errorf("ns acme request: could not retrieve existing priorities: %w", err)
	}

	// If no priorities are found, the nitro request will return an empty slice, so we can return immediately
	if len(bindings.Data) == 0 {
		slog.Debug("ns acme request: no global responder policy bindings found", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "count", len(bindings.Data))
		return output, nil
	}

	// If there are policies bound, add existing priorities to the list
	for _, binding := range bindings.Data {
		slog.Debug("ns acme request: add existing priority to list", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "priority", binding.Priority)
		output = append(output, binding.Priority)
	}
	return output, nil
}

// getPriority finds an available priority for binding the responder policy
func (p *GlobalHttpProvider) getPriority() (string, error) {
	var (
		err                error
		priority           float64 = 33500
		usedPriorities     []string
		validPriorityFound bool = false
	)
	slog.Debug("ns acme request: find valid priority for binding", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL)

	usedPriorities, err = p.getPolicyBindingPriorities()
	if err != nil {
		return "", err
	}

	// If there are no existing priorities, use the deault value + 1
	if len(usedPriorities) == 0 {
		priority = priority + 1
		slog.Debug("ns acme request: using default priority", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "priority", priority)
		return fmt.Sprintf("%g", priority), nil
	}

	// Existing priorities are found, find available priority
	for !validPriorityFound {
		priority = priority + 1
		validPriorityFound = !p.priorityExists(priority, usedPriorities)
	}
	slog.Debug("ns acme request: found available priority", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "priority", priority)
	return fmt.Sprintf("%g", priority), nil
}

// priorityExists checks if a priority is present in a slice or priorities
//
//	priority is the desired priority
//	usedPriorities is the current list of priorities in use
func (p *GlobalHttpProvider) priorityExists(priority float64, usedPriorities []string) bool {
	if len(usedPriorities) == 0 {
		return false
	}

	for _, usedPriority := range usedPriorities {
		// Convert priority to string --> exponent as needed, necessary digits only
		if fmt.Sprintf("%g", priority) == usedPriority {
			slog.Debug("ns acme request: priority is in use", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "priority", priority)
			return true
		}
	}
	slog.Debug("ns acme request: priority is not in use", "provider", ACME_CHALLENGE_PROVIDER_NETSCALER_HTTP_GLOBAL, "priority", priority)
	return false
}

// getResponderActionName generates the name for the responder action
func (p *GlobalHttpProvider) getResponderActionName(domain string) string {
	return p.rsaPrefix + domain + "_" + p.timestamp
}

// getResponderPolicyName generates the name for the responder policy
func (p *GlobalHttpProvider) getResponderPolicyName(domain string) string {
	return p.rspPrefix + domain + "_" + p.timestamp
}

func (p *GlobalHttpProvider) initialize() {
	p.rsaController = controllers.NewResponderActionController(p.client)
	p.rspController = controllers.NewResponderPolicyController(p.client)
	p.rspbController = controllers.NewResponderGlobalResponderPolicyBindingController(p.client)

	if p.timestamp == "" {
		p.timestamp = time.Now().Format("20060102150405")
	}
	p.rspbBindtype = "REQ_OVERRIDE"
	p.rsaPrefix = "RSA_LENS_"
	p.rspPrefix = "RSP_LENS_"
}
