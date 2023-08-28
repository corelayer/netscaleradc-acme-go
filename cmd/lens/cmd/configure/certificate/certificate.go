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

package certificate

import (
	"github.com/corelayer/clapp/pkg/clapp"
	"github.com/spf13/cobra"

	"github.com/corelayer/netscaleradc-acme-go/pkg/controllers"
)

var Command = clapp.Command{
	Cobra: &cobra.Command{
		Use:   "certificate",
		Short: "Configure certificate settings",
		Long:  "Configure certificate settings or generate default certificate config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var example bool
			var c clapp.CommandController
			example, err = cmd.Flags().GetBool("example")
			if example {
				c = controllers.ConfigureGlobalExample{}
			} else {
				c = controllers.ConfigureGlobal{}
			}

			err = c.Execute()
			return err
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	},
}

func init() {
	Command.Cobra.Flags().BoolP("example", "e", false, "generate example")
}
