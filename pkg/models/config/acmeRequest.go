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

type AcmeRequest struct {
	Organization            string   `json:"organization" yaml:"organization" mapstructure:"organization"`
	Environment             string   `json:"environment" yaml:"environment" mapstructure:"environment"`
	CommonName              string   `json:"commonName" yaml:"commonName" mapstructure:"commonName"`
	SubjectAlternativeNames []string `json:"subjectAlternativeNames" yaml:"subjectAlternativeNames" mapstructure:"subjectAlternativeNames"`
}

func (r *AcmeRequest) GetDomains() []string {
	var output []string
	output = append(output, r.CommonName)
	output = append(output, r.SubjectAlternativeNames...)
	return output
}
