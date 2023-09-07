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
	"log/slog"
	"net"
	"strconv"

	"github.com/corelayer/netscaleradc-acme-go/pkg/controllers"
	"github.com/corelayer/netscaleradc-acme-go/pkg/models/config"
)

type Daemon struct {
	Config config.Application
}

func (c Daemon) Execute() error {
	var (
		err      error
		launcher *controllers.Launcher
	)
	if _, err = net.Listen("tcp", c.Config.Daemon.Address+":"+strconv.Itoa(c.Config.Daemon.Port)); err != nil {
		slog.Error("a daemon is already running on the same address")
		return err
	}
	slog.Info("Running daemon", "address", c.Config.Daemon.Address, "port", c.Config.Daemon.Port)

	launcher, err = controllers.NewLauncher(c.Config.ConfigPath, c.Config.Organizations, c.Config.Users)
	return launcher.RequestAll()
}
