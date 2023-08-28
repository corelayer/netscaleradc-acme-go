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

package configure

import (
	"github.com/corelayer/clapp/pkg/clapp"
	"github.com/spf13/cobra"

	"github.com/corelayer/netscaleradc-acme-go/cmd/lens/cmd/configure/certificate"
	"github.com/corelayer/netscaleradc-acme-go/cmd/lens/cmd/configure/global"
	"github.com/corelayer/netscaleradc-acme-go/pkg/controllers"
)

var Command = clapp.Command{
	Cobra: &cobra.Command{
		Use:              "configure",
		Short:            "Configure mode",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			c := controllers.Configure{}
			err = c.Execute()
			return err
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	},
	SubCommands: []clapp.Commander{
		global.Command,
		certificate.Command,
	},
}
