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

	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro/resource/controllers"
	"github.com/corelayer/netscaleradc-nitro-go/pkg/registry"
)

type GlobalHttpProvider struct {
	rsaController  *controllers.ResponderActionController
	rspController  *controllers.ResponderPolicyController
	rspbController *controllers.ResponderGlobalResponderPolicyBindingController

	rsaPrefix string
	rspPrefix string
}

// NewGlobalHttpProvider returns a HTTPProvider instance with a configured list of hosts
func NewGlobalHttpProvider(environment registry.Environment) (*GlobalHttpProvider, error) {
	c := &GlobalHttpProvider{}

	return c, c.initialize(environment)
}

// Present the ACME challenge to the provider.
// Parameter domain is the fqdn for which the challenge will be provided
// Parameter endpoint is the path to which ACME will look  for the challenge (/.well-known/acme-challenge/<token>)
// Parameter keyAuth is the value which must be returned for a successful challenge
func (p *GlobalHttpProvider) Present(domain string, token string, keyAuth string) error {
	var err error
	slog.Info("presenting token for domain", "domain", domain)

	rsaAction := "HTTP/1.1 200 OK\r\nContent-Type/text-plain\r\n\r\n" + keyAuth
	if _, err = p.rsaController.Add(p.getResponderActionName(domain), "respondwith", rsaAction); err != nil {
		slog.Error("could not create responder action for acme challenge", "domain", domain)
		return fmt.Errorf("could not create responder action for acme challenge for domain %s with error %w", domain, err)
	}

	rspRule := "HTTP.REQ.HOSTNAME.EQ(\"" + domain + "\") && HTTP.REQ.URL.EQ(\"/.well-known/acme-challenge/" + token + "\")"
	if _, err = p.rspController.Add(p.getResponderPolicyName(domain), rspRule, p.getResponderActionName(domain), ""); err != nil {
		slog.Error("could not create responder policy for acme challenge", "domain", domain)
		return fmt.Errorf("could not create responder policy for acme challenge for domain %s with error %w", domain, err)
	}

	var priority float64
	priority, err = p.getPriority()
	if err != nil {
		return fmt.Errorf("could not get valid global policy binding priority for acme challenge for domain %s with message %w", domain, err)
	}
	if _, err = p.rspbController.Add(p.getResponderPolicyName(domain), priority, ""); err != nil {
		slog.Error("could not bind global responder policy for acme challenge", "domain", domain)
		return fmt.Errorf("could not bind global responder policy for acme challenge for domain %s with message %w", domain, err)
	}
	return nil
}

func (p *GlobalHttpProvider) CleanUp(domain string, token string, keyAuth string) error {
	var err error
	slog.Info("cleanup token for domain", "domain", domain)

	if _, err = p.rspbController.Delete(p.getResponderPolicyName(domain)); err != nil {
		slog.Error("could not unbind global responder policy for acme challenge", "domain", domain)
		return fmt.Errorf("could not unbind global responder policy for acme challenge for domain %s with message %w", domain, err)
	}

	if _, err = p.rspController.Delete(p.getResponderPolicyName(domain)); err != nil {
		slog.Error("could not remove responder policy for acme challenge", "domain", domain)
		return fmt.Errorf("could not remove responder policy for acme challenge for domain %s with error %w", domain, err)
	}

	if _, err = p.rsaController.Delete(p.getResponderActionName(domain)); err != nil {
		slog.Error("could not remove responder action for acme challenge", "domain", domain)
		return fmt.Errorf("could not remove responder action for acme challenge for domain %s with error %w", domain, err)
	}

	return nil
}

func (p *GlobalHttpProvider) priorityExists(priority float64, usedPriorities []float64) bool {
	for _, usedPriority := range usedPriorities {
		if priority == usedPriority {
			return true
		}
	}
	return false
}

func (p *GlobalHttpProvider) getPriority() (float64, error) {
	var (
		err                error
		priority           float64
		usedPriorities     []float64
		validPriorityFound bool
	)
	priority = 33500
	validPriorityFound = false

	usedPriorities, err = p.getPolicyBindingPriorities()
	if err != nil {
		return float64(0), fmt.Errorf("could not determine valid binding priority with error %w", err)
	}
	for !validPriorityFound {
		priority = priority + 1
		validPriorityFound = p.priorityExists(priority, usedPriorities)
	}
	return priority, nil
}

func (p *GlobalHttpProvider) getPolicyBindingPriorities() ([]float64, error) {
	var output []float64
	slog.Debug("get used priorities for global responder policy bindings")
	bindings, err := p.rspbController.List(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("could not get list of globally bound responder policies with error %w", err)
	}

	for _, binding := range bindings.Data {
		output = append(output, binding.Priority)
	}
	return output, nil
}

func (p *GlobalHttpProvider) getResponderActionName(domain string) string {
	return p.rsaPrefix + domain
}

func (p *GlobalHttpProvider) getResponderPolicyName(domain string) string {
	return p.rspPrefix + domain
}

func (p *GlobalHttpProvider) initialize(e registry.Environment) error {
	slog.Debug("initialize nitro client for primary node for environment %s", e.Nodes)
	client, err := e.GetPrimaryNitroClient()
	if err != nil {
		return fmt.Errorf("failed to initialize GlobalHttpProvider with error %w", err)
	}

	slog.Debug("initialize nitro controllers for responder functionality")
	p.rsaController = controllers.NewResponderActionController(client)
	p.rspController = controllers.NewResponderPolicyController(client)
	p.rspbController = controllers.NewResponderGlobalResponderPolicyBindingController(client)

	p.rsaPrefix = "RSA_LENS_"
	p.rspPrefix = "RSP_LENS_"

	return nil
}
