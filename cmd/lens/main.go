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
	"os"
	"path/filepath"

	"github.com/corelayer/clapp/pkg/clapp"

	"github.com/corelayer/netscaleradc-acme-go/cmd/lens/cmd/request"
)

const (
	LENS_TITLE  = "Let's Encrypt for NetScaler ADC"
	LENS_BANNER = "    __    _______   _______\n   / /   / ____/ | / / ___/\n  / /   / __/ /  |/ /\\__ \\\n / /___/ /___/ /|  /___/ /\n/_____/_____/_/ |_//____/"
	// https://patorjk.com/software/taag/#p=display&v=2&f=Slant&t=LENS
)

var configSearchPaths = []string{
	filepath.Join("/", "etc", "corelayer", "lens"),
	filepath.Join("/nsconfig", "ssl", "LENS"),
	filepath.Join("$HOME", ".lens"),
	filepath.Join("$PWD"),
	filepath.Join("%APPDATA%", "corelayer", "lens"),
	filepath.Join("%LOCALAPPDATA%", "corelayer", "lens"),
	filepath.Join("%PROGRAMDATA%", "corelayer", "lens"),
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

	app := clapp.NewApplication("lens", LENS_TITLE, LENS_BANNER+"\n\n"+LENS_TITLE, "")
	app.Command.PersistentFlags().StringVarP(&configFileFlag, "file", "f", "config.yaml", "config file name")
	app.Command.PersistentFlags().StringVarP(&configPathFlag, "path", "p", "", "config file path, do not use with -s")
	app.Command.PersistentFlags().StringVarP(&logLevelFlag, "loglevel", "l", "", "log level")
	app.Command.PersistentFlags().StringSliceVarP(&configSearchPathFlag, "search", "s", configSearchPaths, "config file search paths, do not use with -p")

	app.Command.MarkFlagsMutuallyExclusive("path", "search")

	if err = app.Command.MarkPersistentFlagDirname("path"); err != nil {
		return err
	}
	if err = app.Command.MarkPersistentFlagFilename("file", "yaml", "yml"); err != nil {
		return err
	}

	app.RegisterCommands([]clapp.Commander{
		// daemon.Command,
		// configure.Command,
		request.Command,
	})

	return app.Run()
}
