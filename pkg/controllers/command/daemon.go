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

package command

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"

	"github.com/corelayer/netscaleradc-acme-go/pkg/models/config"
)

type Daemon struct {
	Config config.Application
}

func (c Daemon) Execute() error {
	var err error
	if _, err = net.Listen("tcp", c.Config.Daemon.Address+":"+strconv.Itoa(c.Config.Daemon.Port)); err != nil {
		slog.Error("a daemon is already running on the same address")
		return err
	}
	slog.Info("Running daemon", "address", c.Config.Daemon.Address, "port", c.Config.Daemon.Port)

	var files []string
	files, err = c.listConfigFiles(c.Config.ConfigPath)
	if err != nil {
		slog.Error("could not read config snippets", "error", err)
		return err
	}

	var configs map[string]*viper.Viper
	configs, err = c.getVipers(files)
	if err != nil {
		slog.Error("could not load config from file", "error", err)
		return err
	}

	for _, currentConfig := range configs {
		var uConfig config.Certificate
		err = currentConfig.Unmarshal(&uConfig)
		slog.Info("certificate config loaded for processing", "name", uConfig.Name)
	}
	time.Sleep(10 * time.Second)
	return nil
}

func (c Daemon) getVipers(files []string) (map[string]*viper.Viper, error) {
	var (
		err    error
		vipers = make(map[string]*viper.Viper, len(files))
	)
	for _, file := range files {
		fileViper := viper.New()
		fileViper.SetConfigFile(file)
		err = fileViper.ReadInConfig()
		if err != nil {
			slog.Error("could not read config from file", "file", file, "error", err)
			continue
		}
		vipers[file] = fileViper
	}
	return vipers, nil
}

func (c Daemon) listConfigFiles(path string) ([]string, error) {
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
			output = append(output, path+"/"+file.Name())
		} else {
			var subDirFiles []string
			subDirFiles, err = c.listConfigFiles(path + "/" + file.Name())
			if err != nil {
				return output, err
			}
			output = append(output, subDirFiles...)
		}
	}
	return output, err
}
