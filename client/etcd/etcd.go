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
	"github.com/chenquan/hit/client/backend"
	"github.com/chenquan/hit/client/hit"
	"github.com/chenquan/hit/internal/consistenthash"
	"github.com/chenquan/hit/internal/consts"
	"github.com/chenquan/hit/internal/logging"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/etcd-io/etcd/clientv3"
	"log"

	"os"
	"sync"
)

//
//type Config struct {
//	Endpoints []string `json:"endpoints"` // etcd服务节点
//	Replicas  int      `json:"replicas"`  // 虚拟节点个数
//}

type Client struct {
	client *clientv3.Client         // etcd客户端
	peers  *consistenthash.Map      // 存储哈希一致性数据
	nodes  map[string]backend.Nodor // key 节点名称,节点结构体
	lock   sync.RWMutex             // 锁,用于
	wg     sync.WaitGroup           // 锁,用于关闭etcd client
}

func NewClient(config *hit.Config) *Client {

	//var config Config
	//if path == "" {
	//	path = "hit.toml"
	//}
	//if _, err := toml.DecodeFile(path, &config); err != nil {
	//	fmt.Println(err)
	//	os.Exit(0)
	//}
	var etcdConfig = clientv3.Config{
		Endpoints: config.Endpoints,
	}

	cli, err := clientv3.New(etcdConfig)
	if err != nil {
		fmt.Println("Error Open", etcdConfig.Endpoints, err)
		os.Exit(0)
	}

	return &Client{client: cli, nodes: make(map[string]backend.Nodor), peers: consistenthash.New(config.Replicas, nil)}
}

// PullAllNodes 拉取所有节点
func (c *Client) PullAllNodes() ([]string, error) {
	c.wg.Add(1)
	defer c.wg.Done()
	return c.PullNodes("")
}

// PullNodes 拉取指定prefix节点
func (c *Client) PullNodes(prefix string) ([]string, error) {
	c.wg.Add(1)
	defer c.wg.Done()

	prefix = consts.DefaultEctdPath + prefix
	response, err := c.client.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	addrs := c.extractAddrs(response)
	// 监听节点变化
	go c.watcher(prefix)

	return addrs, nil
}

// Close 优雅关闭etcd
func (c *Client) Close() {
	// 等待所有etcd操作完成，关闭
	c.wg.Wait()
	_ = c.client.Close()
}

// extractAddrs 扩展新的地址
func (c *Client) extractAddrs(response *clientv3.GetResponse) []string {
	addrs := make([]string, 0)
	if response == nil || response.Kvs == nil {
		return addrs
	}
	for i := range response.Kvs {
		if v := response.Kvs[i].Value; v != nil {
			c.putNode(string(response.Kvs[i].Key), string(response.Kvs[i].Value))
			addrs = append(addrs, string(v))
		}
	}
	return addrs
}

// watcher 监听
func (c *Client) watcher(prefix string) {
	watchChan := c.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for wc := range watchChan {
		for _, ev := range wc.Events {
			switch ev.Type {
			case mvccpb.PUT:
				c.putNode(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE:
				c.delNode(string(ev.Kv.Key))
			}
		}
	}
}

// putNode 更新节点
func (c *Client) putNode(name string, addr string) {
	c.wg.Add(1)
	defer c.wg.Done()
	c.lock.Lock()
	defer c.lock.Unlock()
	addr = addr + consts.DefaultBasePath
	c.nodes[name] = NewNode(addr)
	c.peers.Add(name)
	logging.LogAction("PUT", fmt.Sprintf("Node name:%s, addr:%s", name, addr))
}

// delNode 删除节点
func (c *Client) delNode(name string) {
	c.wg.Add(1)
	defer c.wg.Done()
	c.lock.Lock()
	defer c.lock.Unlock()

	value, exist := c.nodes[name]
	if exist {
		c.peers.Del(name)
		delete(c.nodes, name)
		logging.LogAction("DELETE", fmt.Sprintf("Node name:%s addr%s", name, value))
	}

}

// GetLocalAllNodes 获取当前本地所以节点
func (c *Client) GetLocalAllNodes() map[string]backend.Nodor {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.nodes
}

// Log 记录日志
func (c *Client) Log(format string, v ...interface{}) {
	log.Printf("[Hit] %s.", fmt.Sprintf(format, v...))
}

// PickNode 为当前key选取一个合适的远程节点
func (c *Client) PickNode(key string) (backend.Nodor, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	// 获取一个合适的节点
	if nodeName := c.peers.Get(key); nodeName != "" {
		peer := c.nodes[nodeName]
		c.Log("Pick peer %s", peer.Url())
		return peer, true
	}
	return nil, false
}
