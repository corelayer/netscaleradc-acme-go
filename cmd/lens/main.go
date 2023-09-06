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

package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/corelayer/clapp/pkg/clapp"

	"github.com/corelayer/netscaleradc-acme-go/cmd/lens/cmd/configure"
	"github.com/corelayer/netscaleradc-acme-go/cmd/lens/cmd/daemon"
	"github.com/corelayer/netscaleradc-acme-go/cmd/lens/cmd/request"
)

// Banner generated at https://patorjk.com/software/taag/#p=display&v=3&f=Ivrit&t=NetScaler%20ADC%20-%20ACME
var banner = "\n\n  _   _      _   ____            _                _    ____   ____              _    ____ __  __ _____ \n | \\ | | ___| |_/ ___|  ___ __ _| | ___ _ __     / \\  |  _ \\ / ___|            / \\  / ___|  \\/  | ____|\n |  \\| |/ _ \\ __\\___ \\ / __/ _` | |/ _ \\ '__|   / _ \\ | | | | |      _____    / _ \\| |   | |\\/| |  _|  \n | |\\  |  __/ |_ ___) | (_| (_| | |  __/ |     / ___ \\| |_| | |___  |_____|  / ___ \\ |___| |  | | |___ \n |_| \\_|\\___|\\__|____/ \\___\\__,_|_|\\___|_|    /_/   \\_\\____/ \\____|         /_/   \\_\\____|_|  |_|_____|\n                                                                                                       "

var configSearchPaths = []string{
	filepath.Join("/", "etc", "corelayer", "lens"),
	filepath.Join("/nsconfig", "ssl", "LENS"),
	filepath.Join("$HOME", ".lens"),
	filepath.Join("$PWD"),
}

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}

}

func run() error {
	var err error

	var configFileFlag string
	var configPathFlag string
	var configSearchPathFlag []string
	var logLevelFlag string

	app := clapp.NewApplication("lens", "Let's Encrypt for NetScaler ADC", "", "")
	app.Command.PersistentFlags().StringVarP(&configFileFlag, "file", "f", "config.yaml", "config file name")
	app.Command.PersistentFlags().StringVarP(&configPathFlag, "path", "p", "", "config file path, do not use with -s")
	app.Command.PersistentFlags().StringSliceVarP(&configSearchPathFlag, "search", "s", configSearchPaths, "config file search paths, do not use with -p")
	app.Command.PersistentFlags().StringVarP(&logLevelFlag, "loglevel", "l", "", "log level")
	app.Command.MarkFlagsMutuallyExclusive("path", "search")

	if err = app.Command.MarkPersistentFlagDirname("path"); err != nil {
		return err
	}
	if err = app.Command.MarkPersistentFlagFilename("file", "yaml", "yml"); err != nil {
		return err
	}

	app.RegisterCommands([]clapp.Commander{
		daemon.Command,
		configure.Command,
		request.Command,
	})

	var level slog.Leveler
	switch logLevelFlag {
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

	return app.Run()
}
