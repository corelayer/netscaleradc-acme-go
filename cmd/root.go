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
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "lens",
	// CompletionOptions: cobra.CompletionOptions{
	// 	DisableDefaultCmd: true},
	Short: "Let's Encrypt integration for NetScaler ADC",
	Long: `Let's Encrypt integration for NetScaler ADC
Complete documentation is available at http://github.com/corelayer/netscaleradc-acme-go`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Execute root")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
