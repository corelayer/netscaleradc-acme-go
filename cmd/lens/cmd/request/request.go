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
	"log/slog"
	"os"

	"github.com/corelayer/clapp/pkg/clapp"
	"github.com/spf13/cobra"

	"github.com/corelayer/netscaleradc-acme-go/pkg/controllers/command"
	"github.com/corelayer/netscaleradc-acme-go/pkg/global"
	"github.com/corelayer/netscaleradc-acme-go/pkg/models/config"
)

var Command = clapp.Command{
	Cobra: &cobra.Command{
		Use:   "request",
		Short: "Request mode",
		Long:  global.LENS_BANNER + "\n\n" + global.LENS_TITLE + " - Request Mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			// Get flag values from command
			var configFile string
			var envFile string
			var path string
			var search []string
			var name string
			var all bool

			configFile, err = cmd.Flags().GetString("configFile")
			if err != nil {
				slog.Error("could not find flag", "flag", "configFile")
				return err
			}

			envFile, err = cmd.Flags().GetString("envFile")
			if err != nil {
				slog.Error("could not find flag", "flag", "envFile")
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
				slog.Error("could not find flag", "loglevel", "all")
				return err
			}

			var level slog.Leveler
			switch logLevelFlag {
			case "error":
				level = slog.LevelError
			case "warn":
				level = slog.LevelWarn
			case "info":
				level = slog.LevelInfo
			case "debug":
				level = slog.LevelDebug
			default:
				level = slog.LevelInfo
			}

			// logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
			slog.SetDefault(logger)

			// Setup application environment variables
			appEnvFile := clapp.NewConfiguration(envFile, path, search)
			viperEnv := appEnvFile.GetViper()
			viperEnv.SetEnvPrefix("lens")
			viperEnv.AutomaticEnv()
			err = viperEnv.ReadInConfig()
			if err != nil {
				slog.Error("could not read configuration", "file", viperEnv.ConfigFileUsed(), "error", err)
				return err
			}

			// Setup application configuration
			appConfigFile := clapp.NewConfiguration(configFile, path, search)
			viperFile := appConfigFile.GetViper()

			err = viperFile.ReadInConfig()
			if err != nil {
				slog.Error("could not read configuration", "error", err)
				return err
			}

			var appConfig config.Application
			err = viperFile.Unmarshal(&appConfig)
			if err != nil {
				slog.Error("could not unmarshal configuration", "error", err)
				return err
			}

			err = appConfig.UpdateEnvironmentVariables(viperEnv)
			if err != nil {
				slog.Error("could not update environment variables in config", "error", err)
				return err
			}

			appConfig.Services = append(appConfig.Services, config.ACME_SERVICE_LETSENCRYPT_STAGING)
			appConfig.Services = append(appConfig.Services, config.ACME_SERVICE_LETSENCRYPT_PRODUCTION)

			var c command.Request
			if name != "" {
				c = command.Request{
					Config:     appConfig,
					Request:    name,
					RequestAll: false,
				}
			}

			if all {
				c = command.Request{
					Config:     appConfig,
					Request:    name,
					RequestAll: all,
				}
			}
			err = c.Execute()
			return err
		},
		SilenceErrors: true,
		SilenceUsage:  false,
	},
}

func init() {
	Command.Cobra.Flags().StringP("name", "n", "", "request name")
	Command.Cobra.Flags().BoolP("all", "a", false, "request all")

	Command.Cobra.MarkFlagsMutuallyExclusive("name", "all")

}
