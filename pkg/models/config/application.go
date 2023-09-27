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
	"reflect"
	"regexp"

	"github.com/corelayer/netscaleradc-nitro-go/pkg/registry"
	"github.com/spf13/viper"
)

type Application struct {
	ConfigPath string `json:"configPath" yaml:"configPath" mapstructure:"configPath"`
	// Daemon        Daemon                  `json:"daemon" yaml:"daemon" mapstructure:"daemon"`
	Organizations []registry.Organization `json:"organizations" yaml:"organizations" mapstructure:"organizations"`
	AcmeUsers     []AcmeUser              `json:"acmeUsers" yaml:"acmeUsers" mapstructure:"acmeUsers"`
}

func (a *Application) UpdateEnvironmentVariables(viperEnv *viper.Viper) error {
	var (
		err error
	)

	r := reflect.ValueOf(a)
	err = reflectValues(r, viperEnv)
	if err != nil {
		return err
	}

	return nil
}

func reflectValues(r reflect.Value, viperEnv *viper.Viper) error {
	var (
		err error
		s   reflect.Value
	)
	if r.Kind() == reflect.Ptr {
		// We can only call r.Elem() if r is a Ptr or interface
		s = r.Elem()
	} else {
		s = r
	}

	switch s.Kind() {
	case reflect.Struct:
		n := s.NumField()
		for i := 0; i < n; i++ {
			f := s.Field(i)
			err = reflectValues(f, viperEnv)
			if err != nil {
				return err
			}
		}
	case reflect.Slice:
		// fmt.Println("\treflectValues - SLICE", s.String())
		for i := 0; i < s.Len(); i++ {
			e := s.Index(i)
			// fmt.Println("\t\tSlice", e.String(), e.Kind())
			err = reflectValues(e, viperEnv)
			if err != nil {
				return err
			}
		}
	case reflect.String:
		// fmt.Println("\treflectValues - STRING", s.String())
		err = updateValue(s, viperEnv)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateValue(r reflect.Value, viperEnv *viper.Viper) error {
	re := regexp.MustCompile(`\${LENS_(?P<var>.*)}`)
	var matches []string
	if matches = re.FindStringSubmatch(r.String()); matches == nil {
		return nil
	}
	slog.Debug("replacing environment variable", "variable", r.String())
	v := viperEnv.Get(matches[1])
	if r.CanSet() {
		r.Set(reflect.ValueOf(v))
	}
	return nil
}
