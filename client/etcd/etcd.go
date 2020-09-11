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

package etcd

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/chenquan/hit/client"
	"github.com/chenquan/hit/consistenthash"
	"github.com/chenquan/hit/internal/consts"
	"github.com/chenquan/hit/internal/logging"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/etcd-io/etcd/clientv3"
	"log"

	"os"
	"sync"
)

var Client etcd

type Config struct {
	Endpoints []string `json:"endpoints"` // etcd服务节点
	Replicas  int      `json:"replicas"`  // 虚拟节点个数
}

func Step(path string) {

	var config Config
	if path == "" {
		path = "etcd.toml"
	}
	if _, err := toml.DecodeFile(path, &config); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	var etcdConfig = clientv3.Config{
		Endpoints: config.Endpoints,
	}

	cli, err := clientv3.New(etcdConfig)
	if err != nil {
		fmt.Println("Error Open", etcdConfig.Endpoints, err)
		os.Exit(0)
	}

	Client = etcd{client: cli, nodes: make(map[string]client.Nodor), peers: consistenthash.New(config.Replicas, nil)}
}

//var _ client.Discovery = new(etcd)

type etcd struct {
	client *clientv3.Client        // etcd客户端
	peers  *consistenthash.Map     // 存储哈希一致性数据
	nodes  map[string]client.Nodor // key 节点名称,节点结构体
	lock   sync.RWMutex            // 锁,用于
	wg     sync.WaitGroup          // 锁,用于关闭etcd client
}

func (e *etcd) PullAllNodes() ([]string, error) {
	e.wg.Add(1)
	defer e.wg.Done()
	return e.PullNodes("")
}

func (e *etcd) PullNodes(prefix string) ([]string, error) {
	e.wg.Add(1)
	defer e.wg.Done()

	prefix = consts.DefaultPath + prefix
	response, err := e.client.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	addrs := e.extractAddrs(response)
	// 监听节点变化
	go e.watcher(prefix)

	return addrs, nil
}

// Close 优雅关闭etcd
func (e *etcd) Close() {
	// 等待所有etcd操作完成，关闭
	e.wg.Wait()
	_ = e.client.Close()
}

func (e *etcd) extractAddrs(response *clientv3.GetResponse) []string {
	addrs := make([]string, 0)
	if response == nil || response.Kvs == nil {
		return addrs
	}
	for i := range response.Kvs {
		if v := response.Kvs[i].Value; v != nil {
			e.putNode(string(response.Kvs[i].Key), string(response.Kvs[i].Value))
			addrs = append(addrs, string(v))
		}
	}
	return addrs
}

// watcher 监听
func (e *etcd) watcher(prefix string) {
	watchChan := e.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for wc := range watchChan {
		for _, ev := range wc.Events {
			switch ev.Type {
			case mvccpb.PUT:
				e.putNode(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE:
				e.delNode(string(ev.Kv.Key))
			}
		}
	}
}

// putNode 更新节点
func (e *etcd) putNode(name string, addr string) {
	e.wg.Add(1)
	defer e.wg.Done()
	e.lock.Lock()
	defer e.lock.Unlock()

	e.nodes[name] = client.NewNode(addr)
	e.peers.Add(name)
	log.Println("consistenthash", e.peers)
	logging.LogAction("PUT", fmt.Sprintf("Node name:%s, addr:%s", name, addr))
}

// delNode 删除节点
func (e *etcd) delNode(name string) {
	e.wg.Add(1)
	defer e.wg.Done()
	e.lock.Lock()
	defer e.lock.Unlock()

	value, exist := e.nodes[name]
	if exist {
		e.peers.Del(name)
		delete(e.nodes, name)
		log.Println("consistenthash", e.peers)
		logging.LogAction("DELETE", fmt.Sprintf("Node name:%s addr%s", name, value))
	}

}

// 获取当前本地所以节点
func (e *etcd) GetLocalAllNodes() map[string]client.Nodor {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.nodes
}

// 记录日志
func (e *etcd) Log(format string, v ...interface{}) {
	log.Printf("[Hit] %s.", fmt.Sprintf(format, v...))
}

// 为当前key选取一个合适的远程节点
func (e *etcd) PickNode(key string) (client.Nodor, bool) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	// 获取一个合适的节点
	if nodeName := e.peers.Get(key); nodeName != "" {
		peer := e.nodes[nodeName]
		e.Log("Pick peer %s", peer.Url())
		return peer, true
	}
	return nil, false
}
