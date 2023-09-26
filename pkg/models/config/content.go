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

package config

import (
	"bufio"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type Content struct {
	CommonName                  string   `json:"commonName" yaml:"commonName" mapstructure:"commonName"`
	SubjectAlternativeNames     []string `json:"subjectAlternativeNames" yaml:"subjectAlternativeNames" mapstructure:"subjectAlternativeNames"`
	SubjectAlternativeNamesFile string   `json:"subjectAlternativeNamesFile" yaml:"subjectAlternativeNamesFile" mapstructure:"subjectAlternativeNamesFile"`
}

func (c Content) GetDomains(basePath string) ([]string, error) {
	var (
		err     error
		output  []string
		domains []string
	)
	output = append(output, c.CommonName)

	if len(c.SubjectAlternativeNames) > 0 {
		output = append(output, c.SubjectAlternativeNames...)
	}

	domains, err = c.GetDomainsFromFile(basePath)
	if err != nil {
		return output, err
	}

	output = append(output, domains...)
	return output, c.validateDomains(output)
}

func (c Content) GetDomainsFromFile(basePath string) ([]string, error) {
	var (
		err      error
		filename string
		f        *os.File
		output   []string
	)
	if c.SubjectAlternativeNamesFile != "" {
		_, err = os.Stat(c.SubjectAlternativeNamesFile)
		if err == nil {
			filename = c.SubjectAlternativeNamesFile
		} else {
			fullPath := filepath.Join(basePath, c.SubjectAlternativeNamesFile)
			_, err = os.Stat(fullPath)
			if err == nil {
				filename = fullPath
			} else {
				slog.Error("could not read subject alternative names from file", "filename", fullPath, "error", err)
				return output, err
			}
		}

		f, err = os.Open(filename)
		defer func(f *os.File) {
			err = f.Close()
			if err != nil {
				slog.Error("could not close file", "filename", filename, "error", err)
			}
		}(f)

		if err != nil {
			slog.Error("could not read subject alternative names from file", "filename", filename, "error", err)
			return output, err
		}

		fs := bufio.NewScanner(f)
		fs.Split(bufio.ScanLines)

		for fs.Scan() {
			output = append(output, fs.Text())
		}
	}
	return output, nil
}

func (c Content) validateDomains(domains []string) error {
	var err error
	for _, domain := range domains {
		// Skip wildcard domain validation
		if strings.HasPrefix(domain, "*") {
			continue
		}

		if _, err = net.LookupHost(domain); err != nil {
			return err
		}
	}
	return nil
}
