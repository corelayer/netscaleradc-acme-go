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
	"fmt"
	"net"
	"time"
)

type Daemon struct {
}

func (c *Daemon) Execute() {
	if _, err := net.Listen("tcp", "127.0.0.1:12345"); err != nil {
		fmt.Println("An daemon is already running")
		return
	}
	fmt.Println("Running daemon")
	time.Sleep(10 * time.Second)
}
