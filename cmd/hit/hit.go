/*
 *    Copyright 2020 Chen Quan
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
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/chenquan/hit/internal/consts"
	"github.com/chenquan/hit/internal/register"
	"github.com/chenquan/hit/internal/server"
	"log"
	"net/http"
	"os"
)

func main() {
	var path string
	flag.StringVar(&path, "path", "hit.toml", "配置文件地址")
	flag.Parse()
	config := handleConfig(path)

	// 注册节点
	serverRegister := register.New(config)
	addr := fmt.Sprintf("%s://%s:%s", config.Protocol, config.NodeAddr, config.Port)

	_ = serverRegister.RegisterNode(config.NodeName, addr)

	switch config.Protocol {
	case consts.ProtocolHTTP:
		httpPool := server.NewHTTPPool()
		_ = http.ListenAndServe(":"+config.Port, httpPool)
	}

}

func handleConfig(path string) *register.Config {
	// 存储配置文件信息
	var config register.Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		log.Println(err)
		os.Exit(0)
	}
	if config.Endpoints == nil || len(config.Endpoints) == 0 {
		log.Println("Endpoints 不能为空")
		os.Exit(0)
	}
	if config.NodeAddr == "" {
		log.Println("NodeAddr 不能为空")
		os.Exit(0)
	}
	if config.NodeName == "" {
		log.Println("NodeName 不能为空")
		os.Exit(0)
	}
	// 配置信息处理
	if config.Port == "" {
		config.Port = consts.DefaultPost
	}
	if config.Protocol == "" {
		config.Protocol = consts.ProtocolHTTP
	}
	if config.LeaseTtl == 0 {
		config.LeaseTtl = 10
	}
	if config.DialTimeout == 0 {
		config.DialTimeout = 5
	}
	return &config
}
