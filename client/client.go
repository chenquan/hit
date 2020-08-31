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
// 客户端
package client

// Discovery 服务发现
type Discovery interface {
	// 拉取所有节点
	PullAllNodes() ([]string, error)
	// 拉取指定前戳节点
	PullNodes(prefix string) ([]string, error)
	// 关闭
	Close()
	// 获取节点数据
	GetNodes() map[string]string
}

// 缓存机制
type Hitor interface {
	Get(key string) ([]byte, error)
	Put(kye string, value []byte) error
	SetName(name string)
	Name() string
}
