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

package controllers

//
// import (
// 	"fmt"
// 	"log/slog"
//
// 	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro"
// 	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro/resource/controllers"
// )
//
// type NetScalerOperations struct {
// 	client *nitro.Client
// }
//
// func (c *NetScalerOperations) SslVserverExists(name string) (bool, error) {
// 	var err error
// 	var list []string
//
// 	list, err = c.ListSslVserver()
// 	if err != nil {
// 		slog.Error("could not get list of sslvserver")
// 		return false, err
// 	}
//
// 	for _, item := range list {
// 		if item == name {
// 			return true, nil
// 		}
// 	}
// 	return false, nil
// }
//
// func (c *NetScalerOperations) ListSslVserver() ([]string, error) {
// 	result := []string{
// 		"vserver1",
// 		"vserver2",
// 		"vserver3",
// 	}
// 	return result, nil
// }
//
// func (c *NetScalerOperations) UploadCertificate(name string, location string, publicKey []byte, privateKey []byte) error {
// 	var err error
// 	uploadControl := controllers.NewSystemFileController(c.client)
//
// 	if _, err = uploadControl.Add(name+".cer", location, publicKey); err != nil {
// 		slog.Error("could not upload public key", "name", name, "location", location)
// 		return err
// 	}
// 	if _, err = uploadControl.Add(name+".key", location, privateKey); err != nil {
// 		slog.Error("could not upload private key", "name", name, "location", location)
// 		return err
// 	}
//
// 	return nil
// }
//
// func (c *NetScalerOperations) BindCertificate(name string, vserver string, sni bool) error {
// 	var err error
// 	var exists bool
//
// 	if exists, err = c.SslVserverExists(vserver); err != nil {
// 		return err
// 	}
//
// 	if !exists {
// 		slog.Error("sslvserver does not exist", "sslvserver", name)
// 		return fmt.Errorf("sslvserver %s does not exist", name)
// 	}
//
// 	bindControl := controllers.NewSslCertKeyController(c.client)
// 	if _, err = bindControl.Bind(vserver, name, sni); err != nil {
// 		slog.Error("could not bind certificate to ssl vserver", "name", name, "sslvserver", vserver)
// 		return err
// 	}
// 	return nil
// }
