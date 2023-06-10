#!/usr/bin/env bash
#/*
# * Copyright 2023 CoreLayer BV
# *
# *    Licensed under the Apache License, Version 2.0 (the "License");
# *    you may not use this file except in compliance with the License.
# *    You may obtain a copy of the License at
# *
# *        http://www.apache.org/licenses/LICENSE-2.0
# *
# *    Unless required by applicable law or agreed to in writing, software
# *    distributed under the License is distributed on an "AS IS" BASIS,
# *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# *    See the License for the specific language governing permissions and
# *    limitations under the License.
# */

clear
echo "Building netscaleradc-backup"
echo "-------------------------"
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

cd "$DIR" || exit

echo "Cleaning up previous builds and packages"
rm -rf output/bin/*
rm -rf output/pkg/*


echo "Build executables per platform"
OUTPUT="output/bin/linux/amd64/netscaleradc-backup"
echo " - linux-amd64 --> $OUTPUT"
GOOS=linux GOARCH=amd64 go build -o $OUTPUT main.go

OUTPUT="output/bin/windows/amd64/netscaleradc-backup.exe"
echo " - windows-amd64 --> $OUTPUT"
GOOS=windows GOARCH=amd64 go build -o $OUTPUT main.go

OUTPUT="output/bin/darwin/amd64/netscaleradc-backup"
echo " - darwin-amd64 --> $OUTPUT"
GOOS=darwin GOARCH=amd64 go build -o $OUTPUT main.go

OUTPUT="output/bin/darwin/arm64/netscaleradc-backup"
echo " - darwin-arm64 --> $OUTPUT"
GOOS=darwin GOARCH=arm64 go build -o $OUTPUT main.go