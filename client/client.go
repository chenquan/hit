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

package client

import (
	"fmt"
	"github.com/chenquan/hit/client/etcd"
	"github.com/chenquan/hit/internal/cache"
	cachebackend "github.com/chenquan/hit/internal/cache/backend/cache"
	"github.com/chenquan/hit/internal/cache/lru"
	pb "github.com/chenquan/hit/internal/remotecache"
	"github.com/chenquan/hit/internal/utils"

	"log"
	"sync"
)

//  Loader 一个缓存名称空间，相关的数据加载分布
type Group struct {
	name      string
	getter    Getter
	mainCache *cache.SyncCache
	nodes     NodePicker
	loader    *utils.Loader
}

// GetterFunc 通过函数实现Getter
type GetterFunc func(key string) ([]byte, error)

//  Getter 加载key的数据
type Getter interface {
	Get(key string) ([]byte, error)
}

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup create a new instance of Loader
func NewGroupDefault(name string, cacheBytes int64, getter Getter) *Group {
	group := NewGroup(name, cache.NewSyncCacheDefault(cacheBytes), getter)

	return group
}

func NewGroup(name string, mainCache *cache.SyncCache, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: mainCache,
		loader:    &utils.Loader{},
	}
	client := etcd.NewClient("")
	nodes, err := client.PullNodes(name)
	if err != nil {
		log.Println()
	} else {
		log.Println("获取到初始节点:", nodes)
	}
	g.RegisterPeers(client)
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}
func (g *Group) RegisterPeers(nodes NodePicker) {
	if g.nodes != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.nodes = nodes
}

// Get 通过key获取value
func (g *Group) Get(key string) (cachebackend.Valuer, error) {
	if key == "" {
		return &lru.Value{}, fmt.Errorf("key is required")
	}
	// 从本地缓存中获取数据
	if v, ok := g.mainCache.Get(key); ok {
		log.Println("[Hit] hit", key)
		return v, nil
	}
	// 从非本地缓存中获取
	return g.load(key)
}

// load 当存在节点时,从节点获取数据,否则从本地获取数据
func (g *Group) load(key string) (value cachebackend.Valuer, err error) {
	do, err := g.loader.Do(key, func() (interface{}, error) {
		if g.nodes != nil {
			// 存在节点时,从节点获取数据
			if peer, ok := g.nodes.PickNode(key); ok {
				if value, err = g.getFromNode(peer, key); err == nil {
					return value, nil
				}
				log.Println("[Hit] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return do.(*lru.Value), nil
	}
	return
}

// getFromPeer 从节点获取存储
func (g *Group) getFromNode(peer NodeGetter, key string) (cachebackend.Valuer, error) {
	// 从节点获取存储
	in := &pb.GetRequest{Group: g.name, Key: key}
	out := &pb.GetResponse{}
	err := peer.Get(in, out)
	if err != nil {
		return &lru.Value{}, err
	}
	return lru.NewValue(out.Value), nil
}

// getLocally 从本地非缓存
func (g *Group) getLocally(key string) (cachebackend.Valuer, error) {
	// 从源数据中获取
	value, err := g.getter.Get(key)
	if err != nil {
		return &lru.Value{}, err

	}
	newValue := lru.NewValue(value)
	g.populateCache(key, newValue)
	return newValue, nil
}

// populateCache 填充数据到缓存中
func (g *Group) populateCache(key string, value cachebackend.Valuer) {
	g.mainCache.Add(key, value)
}
