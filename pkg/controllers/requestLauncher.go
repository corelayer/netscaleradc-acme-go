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

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/corelayer/netscaleradc-acme-go/pkg/models"
)

type RequestLauncher struct {
	Providers  map[string]*Provider
	Installers map[string]*Installer
}

func (c *RequestLauncher) Start() {
	slog.Debug("Starting Request Launcher", "providers", len(c.Providers))
	wg := sync.WaitGroup{}
	wg.Add(len(c.Providers))

	ctx, cancel := context.WithCancel(context.Background())
	for _, provider := range c.Providers {
		go func(p *Provider) {
			defer wg.Done()
			p.Run(ctx)
		}(provider)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(5 * time.Second)
		cancel()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(2 * time.Second)
		c.Providers["test1"].Stop()
	}()

	cert := models.Certificate{Name: "Test Certificate"}
	c.Providers["test1"].Input <- cert
	wg.Wait()
}

func (c *RequestLauncher) init() error {
	c.Providers["test1"] = NewProvider("test1")
	c.Providers["test2"] = NewProvider("test2")
	return nil
}

func NewRequestLauncher() (RequestLauncher, error) {
	c := RequestLauncher{
		Providers:  make(map[string]*Provider),
		Installers: make(map[string]*Installer),
	}
	return c, c.init()
}
