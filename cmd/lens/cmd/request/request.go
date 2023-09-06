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

package request

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/corelayer/clapp/pkg/clapp"
	"github.com/spf13/cobra"

	"github.com/corelayer/netscaleradc-acme-go/pkg/controllers/command"
	"github.com/corelayer/netscaleradc-acme-go/pkg/models/config"
)

var Command = clapp.Command{
	Cobra: &cobra.Command{
		Use:   "request",
		Short: "request mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			// Get flag values from command
			var file string
			var path string
			var search []string
			var name string
			var all bool

			file, err = cmd.Flags().GetString("file")
			if err != nil {
				slog.Error("could not find flag", "flag", "file")
				return err
			}

			path, err = cmd.Flags().GetString("path")
			if err != nil {
				slog.Error("could not find flag", "flag", "path")
				return err
			}

			search, err = cmd.Flags().GetStringSlice("search")
			if err != nil {
				slog.Error("could not find flag", "flag", "search")
				return err
			}

			name, err = cmd.Flags().GetString("name")
			if err != nil {
				slog.Error("could not find flag", "flag", "name")
				return err
			}

			all, err = cmd.Flags().GetBool("all")
			if err != nil {
				slog.Error("could not find flag", "flag", "all")
				return err
			}

			var logLevelFlag string
			logLevelFlag, err = cmd.Flags().GetString("loglevel")
			if err != nil {
				slog.Error("could not find flag", "flag", "all")
				return err
			}

			// TODO UPDATE LOGLEVEL HANDLING
			var level slog.Leveler
			fmt.Println("LOGLEVELFLAG", logLevelFlag)
			switch logLevelFlag {
			case "info":
				level = slog.LevelInfo
			case "debug":
				fmt.Println("SETTING LOG LEVEL TO DEBUG")
				level = slog.LevelDebug
			default:
				fmt.Println("DEFAULT SETTINGS")
				level = slog.LevelInfo
			}

			// logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
			slog.SetDefault(logger)

			// Setup application configuration
			clappConfig := clapp.NewConfiguration(file, path, search)
			viper := clappConfig.GetViper()

			err = viper.ReadInConfig()
			if err != nil {
				slog.Error("could not read configuration", "error", err)
				return err
			}

			var appConfig config.Application
			err = viper.Unmarshal(&appConfig)
			if err != nil {
				slog.Error("could not unmarshal configuration", "error", err)
				return err
			}

			c := command.Request{
				Config:     appConfig,
				Request:    name,
				RequestAll: all,
			}
			err = c.Execute()
			return err
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	},
}

func init() {
	Command.Cobra.Flags().StringP("name", "n", "", "request name")
	Command.Cobra.Flags().BoolP("all", "a", true, "request all")

	Command.Cobra.MarkFlagsMutuallyExclusive("name", "all")

}
