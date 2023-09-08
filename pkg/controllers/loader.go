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
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/corelayer/netscaleradc-acme-go/pkg/models/config"
)

type Loader struct {
	basePath   string
	extensions []string
}

func NewLoader(path string) Loader {
	return Loader{
		basePath:   path,
		extensions: []string{".yaml", ".yml"},
	}

}

func (l Loader) GetAll() (map[string]config.Certificate, error) {
	var (
		err    error
		vipers map[string]*viper.Viper
		output map[string]config.Certificate
	)

	slog.Debug("loading configurations")
	vipers, err = l.loadVipers()
	if err != nil {
		slog.Debug("could not get configurations", "error", err)
		return nil, err
	}

	output = make(map[string]config.Certificate, len(vipers))
	for k, v := range vipers {
		var cert config.Certificate
		cert, err = l.loadCertificateConfig(v)
		if err != nil {
			return nil, err
		}
		output[k] = cert
	}
	return output, nil
}

func (l Loader) Get(name string) (config.Certificate, error) {
	var (
		err    error
		vipers map[string]*viper.Viper
		found  bool
		output config.Certificate
	)

	slog.Debug("loading configurations")
	vipers, err = l.loadVipers()
	if err != nil {
		return config.Certificate{}, err
	}

	slog.Debug("searching available configurations", "config", name)
	if _, found = vipers[name]; !found {
		return config.Certificate{}, fmt.Errorf("could not get configuration %s", name)
	}

	output, err = l.loadCertificateConfig(vipers[name])
	if err != nil {
		return config.Certificate{}, err
	}
	return output, nil
}

func (l Loader) getConfigFiles() ([]string, error) {
	var (
		err   error
		files []string
	)
	files, err = l.walkConfigPath(l.basePath)
	if err != nil {
		slog.Debug("could get config files", "error", err)
		return nil, err
	}
	return files, nil
}

func (l Loader) loadCertificateConfig(v *viper.Viper) (config.Certificate, error) {
	var (
		err    error
		output config.Certificate
	)
	err = v.Unmarshal(&output)
	if err != nil {
		slog.Debug("could not unmarshal config", "config", v.GetString("name"))
		return config.Certificate{}, err
	}

	output.AcmeRequest = output.AcmeRequest.SetPath(l.basePath)

	return output, nil
}

func (l Loader) loadViper(path string) (*viper.Viper, error) {
	var (
		err    error
		output *viper.Viper
	)

	output = viper.New()
	output.SetConfigFile(path)
	err = output.ReadInConfig()
	if err != nil {
		slog.Error("could not read config from file", "path", path, "error", err)
		return nil, err
	}
	return output, nil
}

func (l Loader) loadVipers() (map[string]*viper.Viper, error) {
	var (
		err    error
		files  []string
		vipers map[string]*viper.Viper
	)

	files, err = l.getConfigFiles()
	if err != nil {
		return nil, err
	}

	vipers = make(map[string]*viper.Viper, len(files))
	for _, file := range files {
		var currentViper *viper.Viper
		currentViper, err = l.loadViper(file)
		if err != nil {
			return nil, err
		}

		currentName := currentViper.Get("name")
		if currentName == "" {
			slog.Error("could not read config name from file", "file", file)
			return nil, fmt.Errorf("could not read config name from file %s", file)
		}
		vipers[currentViper.GetString("name")] = currentViper
	}
	return vipers, nil
}

func (l Loader) walkConfigPath(path string) ([]string, error) {
	var (
		err    error
		files  []os.DirEntry
		output []string
	)

	files, err = os.ReadDir(path)
	if err != nil {
		slog.Error("cannot list files in config directory", "error", err)
		return output, fmt.Errorf("cannot list files in config directory with message %w", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			if l.isValidConfigFile(file) {
				output = append(output, path+"/"+file.Name())
			}
		} else {
			var subDirFiles []string
			subDirFiles, err = l.walkConfigPath(path + "/" + file.Name())
			if err != nil {
				return output, err
			}
			output = append(output, subDirFiles...)
		}
	}
	return output, err
}

func (l Loader) isValidConfigFile(e os.DirEntry) bool {
	ext := filepath.Ext(e.Name())
	for _, v := range l.extensions {
		if v == ext {
			return true
		}
	}
	return false
}
