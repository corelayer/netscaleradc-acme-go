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

	"github.com/corelayer/clapp/pkg/clapp"

	"github.com/corelayer/netscaleradc-acme-go/cmd/lens/cmd/daemon"
)

// Banner generated at https://patorjk.com/software/taag/#p=display&v=3&f=Ivrit&t=NetScaler%20ADC%20-%20ACME
var banner = "\n\n  _   _      _   ____            _                _    ____   ____              _    ____ __  __ _____ \n | \\ | | ___| |_/ ___|  ___ __ _| | ___ _ __     / \\  |  _ \\ / ___|            / \\  / ___|  \\/  | ____|\n |  \\| |/ _ \\ __\\___ \\ / __/ _` | |/ _ \\ '__|   / _ \\ | | | | |      _____    / _ \\| |   | |\\/| |  _|  \n | |\\  |  __/ |_ ___) | (_| (_| | |  __/ |     / ___ \\| |_| | |___  |_____|  / ___ \\ |___| |  | | |___ \n |_| \\_|\\___|\\__|____/ \\___\\__,_|_|\\___|_|    /_/   \\_\\____/ \\____|         /_/   \\_\\____|_|  |_|_____|\n                                                                                                       "

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// filename := "config.yaml"
	// paths := []string{
	// 	filepath.Join("etc", "corelayer", "lens"),
	// 	filepath.Join("nsconfig", "ssl", "LENS"),
	// 	filepath.Join("$HOME", ".lens"),
	// 	filepath.Join("$PWD"),
	// }
	// config, err := clapp.NewConfiguration(filename, paths, false)
	// if err != nil {
	// 	slog.Error("could not create new configuration", "error", err)
	// 	os.Exit(1)
	// }

	app := clapp.NewApplication("lens", "Let's Encrypt for NetScaler ADC", "", "")
	app.RegisterCommands([]clapp.Commander{
		daemon.Command,
	})
	app.Run()
}
