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
echo "Archiving netscaleradc-acme"
echo "--------------------------"
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

cd "$DIR" || exit

rm -rf output/archives/*

OUTBASEBIN="output/bin"
OUTBASEZIP="output/archives"

echo "Archive executables per platform"
GOOS=linux
GOARCH=amd64
OUTEXT="tgz"
OUTPUT="${OUTBASEZIP}/${GOOS}_${GOARCH}.${OUTEXT}"
INPUT="${OUTBASEBIN}/${GOOS}/${GOARCH}/*"
echo " - $GOOS-$GOARCH --> $OUTBASEZIP"
tar -c -a -f $OUTPUT $INPUT

GOOS=linux
GOARCH=arm64
OUTEXT="tgz"
OUTPUT="${OUTBASEZIP}/${GOOS}_${GOARCH}.${OUTEXT}"
INPUT="${OUTBASEBIN}/${GOOS}/${GOARCH}/*"
echo " - $GOOS-$GOARCH --> $OUTBASEZIP"
tar -c -a -f $OUTPUT $INPUT

GOOS=windows
GOARCH=amd64
OUTEXT="zip"
OUTPUT="${OUTBASEZIP}/${GOOS}_${GOARCH}.${OUTEXT}"
INPUT="${OUTBASEBIN}/${GOOS}/${GOARCH}/*"
echo " - $GOOS-$GOARCH --> $OUTBASEZIP"
tar -c -a -f $OUTPUT $INPUT

GOOS=windows
GOARCH=arm64
OUTEXT="zip"
OUTPUT="${OUTBASEZIP}/${GOOS}_${GOARCH}.${OUTEXT}"
INPUT="${OUTBASEBIN}/${GOOS}/${GOARCH}/*"
echo " - $GOOS-$GOARCH --> $OUTBASEZIP"
tar -c -a -f $OUTPUT $INPUT

GOOS=darwin
GOARCH=amd64
OUTEXT="tgz"
OUTPUT="${OUTBASEZIP}/${GOOS}_${GOARCH}.${OUTEXT}"
INPUT="${OUTBASEBIN}/${GOOS}/${GOARCH}/*"
echo " - $GOOS-$GOARCH --> $OUTBASEZIP"
tar -c -a -f $OUTPUT $INPUT

GOOS=darwin
GOARCH=arm64
OUTEXT="tgz"
OUTPUT="${OUTBASEZIP}/${GOOS}_${GOARCH}.${OUTEXT}"
INPUT="${OUTBASEBIN}/${GOOS}/${GOARCH}/*"
echo " - $GOOS-$GOARCH --> $OUTBASEZIP"
tar -c -a -f $OUTPUT $INPUT

GOOS=freebsd
GOARCH=amd64
OUTEXT="tgz"
OUTPUT="${OUTBASEZIP}/${GOOS}_${GOARCH}.${OUTEXT}"
INPUT="${OUTBASEBIN}/${GOOS}/${GOARCH}/*"
echo " - $GOOS-$GOARCH --> $OUTBASEZIP"
tar -c -a -f $OUTPUT $INPUT

GOOS=freebsd
GOARCH=arm64
OUTEXT="tgz"
OUTPUT="${OUTBASEZIP}/${GOOS}_${GOARCH}.${OUTEXT}"
INPUT="${OUTBASEBIN}/${GOOS}/${GOARCH}/*"
echo " - $GOOS-$GOARCH --> $OUTBASEZIP"
tar -c -a -f $OUTPUT $INPUT
