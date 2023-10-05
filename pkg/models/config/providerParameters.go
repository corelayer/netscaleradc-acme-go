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
	"log/slog"
	"os"
)

type ProviderParameters struct {
	Name      string                `json:"name" yaml:"name" mapstructure:"name"`
	Variables []EnvironmentVariable `json:"variables" yaml:"variables" mapstructure:"variables"`
}

func (p ProviderParameters) ApplyEnvironmentVariables() error {
	var err error
	slog.Debug("applying provider parameters", "name", p.Name)
	for _, v := range p.Variables {
		slog.Debug("applying provider parameter", "name", p.Name, "variable", v.Name)
		err = os.Setenv(v.Name, v.Value)
		if err != nil {
			return err
		}
	}
	slog.Debug("applying provider parameters completed", "name", p.Name)
	return nil
}

func (p ProviderParameters) ResetEnvironmentVariables() error {
	var err error
	slog.Debug("resetting provider parameters", "name", p.Name)
	for _, v := range p.Variables {
		slog.Debug("resetting provider parameter", "name", p.Name, "variable", v.Name)
		err = os.Unsetenv(v.Name)
		if err != nil {
			return err
		}
	}
	slog.Debug("resetting provider parameters completed", "name", p.Name)
	return nil
}
