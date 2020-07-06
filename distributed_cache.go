/*
 *    Copyright  2020 Chen Quan
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

package hit

import (
	"fmt"
	"github.com/chenquan/hit/remote"
	pb "github.com/chenquan/hit/remote/remotecache"
	"log"
	"sync"
)

//  Loader 一个缓存名称空间，相关的数据加载分布
type Group struct {
	name      string
	getter    Getter
	mainCache *SyncCache
	peers     remote.PeerPicker
	loader    *remote.Loader
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
	return NewGroup(name, NewSyncCacheDefault(cacheBytes), getter)
}

func NewGroup(name string, mainCache *SyncCache, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: mainCache,
		loader:    &remote.Loader{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}
func (g *Group) RegisterPeers(peers remote.PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// Get 通过key获取value
func (g *Group) Get(key string) (ReadBytes, error) {
	if key == "" {
		return ReadBytes{}, fmt.Errorf("key is required")
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
func (g *Group) load(key string) (value ReadBytes, err error) {
	do, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			// 存在节点时,从节点获取数据
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[Hit] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return do.(ReadBytes), nil
	}
	return
}

// getFromPeer 从节点获取存储
func (g *Group) getFromPeer(peer remote.PeerGetter, key string) (ReadBytes, error) {
	// 从节点获取存储
	in := &pb.Request{Group: g.name, Key: key}
	out := &pb.Response{}
	err := peer.Get(in, out)
	if err != nil {
		return ReadBytes{}, err
	}
	return ReadBytes{b: out.Value}, nil
}

// getLocally 从本地获取数据(非缓存)
func (g *Group) getLocally(key string) (ReadBytes, error) {
	// 从源数据中获取
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ReadBytes{}, err

	}
	value := ReadBytes{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// populateCache 填充数据到缓存中
func (g *Group) populateCache(key string, value ReadBytes) {
	g.mainCache.Add(key, value)
}
