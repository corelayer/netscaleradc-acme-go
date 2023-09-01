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

type GlobalHttpProvider struct {
	nitroClient    *nitro.Client
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
func NewGlobalHttpProvider(environment registry.Environment, maxRetries int, timestamp string) (*GlobalHttpProvider, error) {
	c := &GlobalHttpProvider{
		maxRetries: maxRetries,
		timestamp:  timestamp,
	}

	return c, c.initialize(environment)
}

// Present the ACME challenge to the provider.
// Parameter domain is the fqdn for which the challenge will be provided
// Parameter endpoint is the path to which ACME will look  for the challenge (/.well-known/acme-challenge/<token>)
// Parameter keyAuth is the value which must be returned for a successful challenge
func (p *GlobalHttpProvider) Present(domain string, token string, keyAuth string) error {
	var err error
	slog.Info("prepare acme request", "domain", domain)

	slog.Debug("adding responder action", "domain", domain, "action", p.getResponderActionName(domain))
	rsaAction := "\"HTTP/1.1 200 OK\\r\\n\\r\\n" + keyAuth + "\""
	if _, err = p.rsaController.Add(p.getResponderActionName(domain), "respondwith", rsaAction); err != nil {
		slog.Error("could not create responder action for acme challenge", "domain", domain, "action", p.getResponderActionName(domain))
		return fmt.Errorf("could not create responder action %s for acme challenge for domain %s with error %w", p.getResponderActionName(domain), domain, err)
	}

	slog.Debug("adding responder policy", "domain", domain, "policy", p.getResponderPolicyName(domain))
	rspRule := "HTTP.REQ.HOSTNAME.EQ(\"" + domain + "\") && HTTP.REQ.URL.EQ(\"/.well-known/acme-challenge/" + token + "\")"
	if _, err = p.rspController.Add(p.getResponderPolicyName(domain), rspRule, p.getResponderActionName(domain), ""); err != nil {
		slog.Error("could not create responder policy for acme challenge", "domain", domain, "policy", p.getResponderPolicyName(domain))
		return fmt.Errorf("could not create responder policy %s for acme challenge for domain %s with error %w", p.getResponderPolicyName(domain), domain, err)
	}

	if err = p.bindResponderPolicy(domain); err != nil {
		slog.Error("could not bind global responder policy for acme challenge", "domain", domain)
	}

	slog.Debug("prepare acme request completed", "domain", domain)
	return nil
}

func (p *GlobalHttpProvider) CleanUp(domain string, token string, keyAuth string) error {
	var err error
	slog.Info("cleanup acme request", "domain", domain)

	slog.Debug("cleaning up global responder policy binding", "domain", domain, "policy", p.getResponderPolicyName(domain))
	if _, err = p.rspbController.Delete(p.getResponderPolicyName(domain), p.rspbBindtype); err != nil {
		slog.Error("could not unbind global responder policy for acme challenge", "domain", domain, "policy", p.getResponderPolicyName(domain))
		return fmt.Errorf("could not unbind global responder policy %s for acme challenge for domain %s with message %w", p.getResponderPolicyName(domain), domain, err)
	}

	slog.Debug("cleaning up responder policy", "domain", domain, "policy", p.getResponderPolicyName(domain))
	if _, err = p.rspController.Delete(p.getResponderPolicyName(domain)); err != nil {
		slog.Error("could not remove responder policy for acme challenge", "domain", domain, "policy", p.getResponderPolicyName(domain))
		return fmt.Errorf("could not remove responder policy %s for acme challenge for domain %s with error %w", p.getResponderPolicyName(domain), domain, err)
	}

	slog.Debug("cleaning up responder action", "domain", domain, "action", p.getResponderActionName(domain))
	if _, err = p.rsaController.Delete(p.getResponderActionName(domain)); err != nil {
		slog.Error("could not remove responder action for acme challenge", "domain", domain, "action", p.getResponderActionName(domain))
		return fmt.Errorf("could not remove responder action %s for acme challenge for domain %s with error %w", p.getResponderActionName(domain), domain, err)
	}

	slog.Debug("acme request cleanup completed", "domain", domain)
	return nil
}

func (p *GlobalHttpProvider) bindResponderPolicy(domain string) error {
	var (
		successfullyBoundPolicy = false
		retries                 = 0
		err                     error
		priority                string
	)

	for !successfullyBoundPolicy {
		slog.Debug("search for valid binding priority", "domain", domain)

		retries += 1
		priority, err = p.getPriority()
		if err != nil {
			return fmt.Errorf("could not get valid global policy binding priority for acme challenge for domain %s with message %w", domain, err)
		}

		slog.Debug("binding global responder policy", "policy", p.getResponderPolicyName(domain), "priority", priority)
		if _, err = p.rspbController.Add(p.getResponderPolicyName(domain), p.rspbBindtype, priority, ""); err != nil {
			if retries == p.maxRetries {
				return fmt.Errorf("could not bind global responder policy for acme challenge for domain %s with message %w", domain, err)
			}
			continue
		}
		successfullyBoundPolicy = true
	}
	return nil
}

func (p *GlobalHttpProvider) priorityExists(priority float64, usedPriorities []string) bool {
	if len(usedPriorities) == 0 {
		slog.Debug("priorityExists", "exists", false, "priority", priority)
		return false
	}

	for _, usedPriority := range usedPriorities {
		// Convert priority to string --> exponent as needed, necessary digits only
		if fmt.Sprintf("%g", priority) == usedPriority {
			slog.Debug("priorityExists", "exists", true, "priority", priority)
			return true
		}
	}
	slog.Debug("priorityExists processed all usedPriorities", "exists", false, "priority", priority)
	return false
}

func (p *GlobalHttpProvider) generateAcmeUrl(token string) string {
	return "/.well-known/acme-challenge/" + token
}

func (p *GlobalHttpProvider) getPriority() (string, error) {
	var (
		err                error
		priority           float64
		usedPriorities     []string
		validPriorityFound bool
	)
	priority = 33500
	validPriorityFound = false

	usedPriorities, err = p.getPolicyBindingPriorities()
	if err != nil {
		return "", fmt.Errorf("could not determine valid binding priority with error %w", err)
	}

	if len(usedPriorities) == 0 {
		priority = priority + 1
		slog.Debug("no global responder policy bindings found, using default priority", "priority", priority)
		return fmt.Sprintf("%g", priority), nil
	}

	for !validPriorityFound {
		priority = priority + 1
		validPriorityFound = !p.priorityExists(priority, usedPriorities)
	}
	return fmt.Sprintf("%g", priority), nil
}

func (p *GlobalHttpProvider) getPolicyBindingPriorities() ([]string, error) {
	var output []string
	slog.Debug("get used priorities for global responder policy bindings")

	bindingsNitroRequest := &nitro.Request[config.ResponderGlobalResponderPolicyBinding]{
		Arguments: map[string]string{
			"type": "REQ_OVERRIDE",
		},
		Attributes: []string{"priority"},
	}

	var err error
	var bindings *nitro.Response[config.ResponderGlobalResponderPolicyBinding]
	bindings, err = nitro.ExecuteNitroRequest[config.ResponderGlobalResponderPolicyBinding](p.nitroClient, bindingsNitroRequest)
	if err != nil {
		return nil, fmt.Errorf("could not get list of globally bound responder policies with error %w", err)
	}

	if len(bindings.Data) == 0 {
		slog.Debug("no global responder policy bindings found", "count", len(bindings.Data))
		return output, nil
	}

	for _, binding := range bindings.Data {
		slog.Debug("adding priority to list", "priority", binding.Priority)
		output = append(output, binding.Priority)
	}
	return output, nil
}

func (p *GlobalHttpProvider) getResponderActionName(domain string) string {
	return p.rsaPrefix + domain + "_" + p.timestamp
}

func (p *GlobalHttpProvider) getResponderPolicyName(domain string) string {
	return p.rspPrefix + domain + "_" + p.timestamp
}

func (p *GlobalHttpProvider) initialize(e registry.Environment) error {
	slog.Debug("initialize nitro client for primary node for environment", "environment", e.Name)
	client, err := e.GetPrimaryNitroClient()
	if err != nil {
		return fmt.Errorf("failed to initialize GlobalHttpProvider with error %w", err)
	}

	slog.Debug("initialize nitro controllers for responder functionality")
	p.nitroClient = client
	p.rsaController = controllers.NewResponderActionController(client)
	p.rspController = controllers.NewResponderPolicyController(client)
	p.rspbController = controllers.NewResponderGlobalResponderPolicyBindingController(client)

	if p.timestamp == "" {
		p.timestamp = time.Now().Format("20060102150405")
	}
	p.rspbBindtype = "REQ_OVERRIDE"
	p.rsaPrefix = "RSA_LENS_"
	p.rspPrefix = "RSP_LENS_"

	return nil
}
