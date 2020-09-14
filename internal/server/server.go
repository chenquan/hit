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

package server

import (
	"fmt"
	"github.com/chenquan/hit/client/etcd"
	"github.com/chenquan/hit/internal/cache"
	cachebackend "github.com/chenquan/hit/internal/cache/backend/cache"
	"github.com/chenquan/hit/internal/cache/lru"
	_ "github.com/chenquan/hit/internal/consistenthash"
	"github.com/chenquan/hit/internal/consts"
	pb "github.com/chenquan/hit/internal/remotecache"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Group struct {
	name      string
	mainCache *cache.SyncCache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroupDefault(name string, cacheBytes int64) *Group {
	group := NewGroup(name, cache.NewSyncCacheDefault(cacheBytes))

	return group
}

func NewGroup(name string, mainCache *cache.SyncCache) *Group {

	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		mainCache: mainCache,
	}
	client := etcd.NewClient("")
	nodes, err := client.PullNodes(name)
	if err != nil {
		log.Println()
	} else {
		log.Println("获取到初始节点:", nodes)
	}

	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	g := groups[name]
	return g
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
			log.Println("[Hit] hit", key)
			return v, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (g *Group) Add(key string, value cachebackend.Valuer) error {
	if key == "" {
		return fmt.Errorf("key is required")
	}
	g.mainCache.Add(key, value)
	return nil
}
func (g *Group) Delete(key string) error {
	if key == "" {
		return fmt.Errorf("key is required")
	}
	g.mainCache.Remove(key)
	return nil
}

// populateCache 填充数据到缓存中
func (g *Group) populateCache(key string, value cachebackend.Valuer) {
	g.mainCache.Add(key, value)
}

type HTTPPool struct {
	// this peer's base URL, e.g. "https://example.net:8000"
	self      string
	basePath  string
	mu        sync.Mutex // guards peers and httpGetters
	mainCache *cache.SyncCache
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self: self,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Hit] %s %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	switch r.Method {
	case http.MethodGet:
		get(groupName, key, w, r)
		fmt.Println("GET")
		break
	case http.MethodPost:
		set(groupName, key, w, r)
		fmt.Println("POST")
		break
	case http.MethodDelete:
		del(w, r)
		fmt.Println("DELETE")
		break

	}

	//// 获取
	//group := p.mainCache.Get()
	//if group == nil {
	//	http.Error(w, "no such group: "+groupName, http.StatusNotFound)
	//	return
	//}
	//
	//view, err := group.Get(key)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//bytes, err := proto.Marshal(&pb.Response{Value: view.Bytes()})
	//w.Header().Set("Content-Type", "application/octet-stream")
	//w.Write(bytes)

}
func get(groupName string, key string, w http.ResponseWriter, r *http.Request) {
	if group, ok := groups[groupName]; ok {
		valuer, err := group.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := &pb.Data{
			Group:  groupName,
			Value:  valuer.Bytes(),
			Expire: valuer.Expire(),
		}
		bytes, err := proto.Marshal(&pb.GetResponse{Success: true, Message: "success", Data: data})
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(bytes)
	} else {
		bytes, _ := proto.Marshal(&pb.GetResponse{Success: false, Message: "fail"})
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(bytes)
	}

}
func set(groupName string, key string, w http.ResponseWriter, r *http.Request) {
	bytesData, err := ioutil.ReadAll(r.Body)
	requestBody := &pb.SetRequest{}
	err = proto.Unmarshal(bytesData, requestBody)
	if err == nil {
		var group *Group
		var ok bool
		if group, ok = groups[groupName]; !ok {
			group = NewGroupDefault(groupName, 10000)
		}
		value := requestBody.Value
		expire := time.Now().Add(consts.DefaultNodeCacheDuration).Unix()
		err := group.Add(key, lru.NewValue(value, expire, groupName))

		if err == nil {
			data := &pb.Data{
				Group:  groupName,
				Value:  value,
				Expire: expire,
			}
			bytes, _ := proto.Marshal(&pb.SetResponse{Success: false, Message: "success", Data: data})
			w.Header().Set("Content-Type", "application/octet-stream")
			_, _ = w.Write(bytes)
			return
		}

	}
	bytes, _ := proto.Marshal(&pb.GetResponse{Success: false, Message: "fail"})
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(bytes)

}
func del(w http.ResponseWriter, r *http.Request) {

}

//// Set 更新PeerPool列表
//func (p *HTTPPool) Set(peers ...string) {
//	p.mu.Lock()
//	defer p.mu.Unlock()
//	p.peers = consistenthash.New(defaultReplicas, nil)
//	p.peers.Add(peers...)
//	p.httpGetters = make(map[string]*httpGetter, len(peers))
//	for _, peer := range peers {
//		// 远程节点
//		p.httpGetters[peer] = &httpGetter{url: peer + p.basePath}
//	}
//}
//
//// PickPeer picks a peer according to key
//func (p *HTTPPool) PickPeer(key string) (remote.PeerGetter, bool) {
//	p.mu.Lock()
//	defer p.mu.Unlock()
//	// 获取一个合适的节点
//	if peer := p.peers.Get(key); peer != "" && peer != p.self {
//		p.Log("Pick peer %s", peer)
//		return p.httpGetters[peer], true
//	}
//	return nil, false
//}
