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
	"github.com/BurntSushi/toml"
	"github.com/chenquan/hit/client/backend"
	"github.com/chenquan/hit/client/etcd"
	"github.com/chenquan/hit/client/hit"
	"github.com/chenquan/hit/internal/cache"
	cachebackend "github.com/chenquan/hit/internal/cache/backend/cache"
	"github.com/chenquan/hit/internal/cache/lru"
	"github.com/chenquan/hit/internal/consts"
	pb "github.com/chenquan/hit/internal/remotecache"
	"github.com/chenquan/hit/internal/utils"
	"os"
	"time"

	"log"
	"sync"
)

type Hit struct {
	client *etcd.Client
	groups map[string]*Group
	rwLock sync.RWMutex
}

func NewHit(config *hit.Config) *Hit {
	return &Hit{
		client: etcd.NewClient(config),
		groups: make(map[string]*Group),
	}
}
func NewHitFromPath(path string) *Hit {

	var config hit.Config
	if path == "" {
		path = "hit.toml"
	}
	if _, err := toml.DecodeFile(path, &config); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	return NewHit(&config)
}

// NewGroup create a new instance of Loader
func (h *Hit) NewGroupDefault(name, nodeName string, cacheBytes int64, getter Getter) *Group {
	group := h.NewGroup(name, nodeName, cache.NewSyncCacheDefault(cacheBytes), getter)
	return group
}
func (h *Hit) NewGroup(name, nodeName string, mainCache *cache.SyncCache, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	h.rwLock.Lock()
	defer h.rwLock.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: mainCache,
		loader:    &utils.Loader{},
	}
	nodes, err := h.client.PullNodes(nodeName)
	if err != nil {
		log.Println()
	} else {
		log.Println("获取到初始节点:", nodes)
	}
	g.registerPeers(h.client)
	h.groups[name] = g
	return g
}

func (h *Hit) GetGroup(name string) *Group {
	h.rwLock.RLock()
	defer h.rwLock.RUnlock()
	g := h.groups[name]
	return g
}

type Group struct {
	name      string
	getter    Getter
	mainCache *cache.SyncCache
	nodes     backend.NodePicker
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

func (g *Group) registerPeers(nodes backend.NodePicker) {
	if g.nodes != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.nodes = nodes
}

// Get 通过key获取value
func (g *Group) Get(key string) (cachebackend.Valuer, error) {
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}
	// 从本地缓存(一级缓存)中获取数据
	if v, ok := g.mainCache.Get(key); ok {
		// 检查数据是否过期
		if v.Expire() <= time.Now().Unix() {
			// 过期删除
			g.mainCache.Remove(key)
		} else {
			log.Println("[Hit] hit 一级缓存数据", key)
			return v, nil
		}
	}

	// 从远程节点中获取
	return g.load(key)
}

func (g *Group) Set(key string, value cachebackend.Valuer, isLocalCache bool) (newValue cachebackend.Valuer, err error) {
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}
	// 写入本地缓存
	if isLocalCache {
		newValue = lru.NewValue(value.Bytes(), time.Now().Add(consts.DefaultLocalCacheDuration).Unix(), value.GroupName())
		g.populateCache(key, newValue)
	}
	if g.nodes != nil {
		// 存在节点时,从节点获取数据
		if peer, ok := g.nodes.PickNode(key); ok {
			if value, err = g.setFromNode(peer, key, value); err == nil {
				return value, nil
			}
			log.Println("[Hit] Failed to get from peer", err)
		}
	}
	return newValue, nil
}

// load 当存在节点时,从节点获取数据,否则从本地DB获取数据
func (g *Group) load(key string) (value cachebackend.Valuer, err error) {
	do, err := g.loader.Do(key, func() (interface{}, error) {
		log.Println("[Hit] hit 获取远程节点数据")
		if g.nodes != nil {
			// 存在节点时,从节点获取数据
			if peer, ok := g.nodes.PickNode(key); ok {
				if value, err = g.getFromNode(peer, key); err == nil {
					// 克隆一个新值,存入本地(一级)缓存
					newValue := lru.NewValue(value.Bytes(), time.Now().Add(consts.DefaultLocalCacheDuration).Unix(), value.GroupName())
					g.populateCache(key, newValue)
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
func (g *Group) getFromNode(peer backend.NodeGetter, key string) (cachebackend.Valuer, error) {
	// 从节点获取存储
	in := &pb.GetRequest{Group: g.name, Key: key}
	out := &pb.GetResponse{}
	err := peer.Get(in, out)
	if err != nil {
		return &lru.Value{}, err
	}
	return lru.NewValue(out.Data.Value, out.Data.Expire, out.Data.Group), nil
}

// getFromPeer 从节点获取存储
func (g *Group) setFromNode(peer backend.NodeSetter, key string, value cachebackend.Valuer) (cachebackend.Valuer, error) {
	// 从节点获取存储
	in := &pb.SetRequest{Group: g.name, Key: key, Value: value.Bytes()}
	out := &pb.SetResponse{}
	err := peer.Set(in, out)
	if err != nil {
		return nil, err
	}
	return lru.NewValue(out.Data.Value, out.Data.Expire, out.Data.Group), nil
}

// getFromPeer 从节点获取存储
func (g *Group) delFromNode(peer backend.NodeDeler, key string) error {
	// 从节点获取存储
	in := &pb.DelRequest{Group: g.name, Key: key}
	out := &pb.DelResponse{}
	err := peer.Del(in, out)
	if err != nil {
		return err
	}
	return nil
}

// getLocally 从本地DB
func (g *Group) getLocally(key string) (cachebackend.Valuer, error) {
	// 从DB数据中获取
	value, err := g.getter.Get(key)
	if err != nil {
		return nil, err
	}
	newValue := lru.NewValue(value, time.Now().Add(consts.DefaultLocalCacheDuration).Unix(), g.name)
	g.populateCache(key, newValue)
	return newValue, nil
}

// populateCache 填充数据到缓存中
func (g *Group) populateCache(key string, value cachebackend.Valuer) {
	g.mainCache.Add(key, value)
}
