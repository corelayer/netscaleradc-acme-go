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

	"github.com/corelayer/netscaleradc-acme-go/pkg/controllers"
	"github.com/corelayer/netscaleradc-acme-go/pkg/models/config"
)

type Request struct {
	Config     config.Application
	Request    string
	RequestAll bool
}

func (c Request) Execute() error {
	var (
		err      error
		launcher *controllers.Launcher
	)
	launcher, err = controllers.NewLauncher(c.Config.ConfigPath, c.Config.Organizations, c.Config.Users)
	if err != nil {
		return err
	}

	if c.Request != "" {
		return launcher.Request(c.Request)
	}
	if c.RequestAll {
		return launcher.RequestAll()

	}
	return fmt.Errorf("no valid execution target")
}
