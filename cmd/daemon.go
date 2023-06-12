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

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/corelayer/netscaleradc-acme-go/pkg/controllers"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Daemon mode",
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		c := controllers.Daemon{}
		c.Execute()
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
