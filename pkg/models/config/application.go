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

import (
	"fmt"

	"github.com/corelayer/netscaleradc-nitro-go/pkg/registry"
)

type Application struct {
	User          AcmeUser                `json:"user" yaml:"user" mapstructure:"user"`
	ConfigPath    string                  `json:"configPath" yaml:"configPath" mapstructure:"configPath"`
	Daemon        Daemon                  `json:"daemon" yaml:"daemon" mapstructure:"daemon"`
	Organizations []registry.Organization `json:"organizations" yaml:"organizations" mapstructure:"organizations"`
}

func (a *Application) GetEnvironment(organization string, environment string) (registry.Environment, error) {
	for _, org := range a.Organizations {
		if organization == org.Name {
			for _, env := range org.Environments {
				if environment == env.Name {
					return env, nil
				}
			}
			break
		}
	}
	return registry.Environment{}, fmt.Errorf("could not find environment %s for organization %s", environment, organization)
}
