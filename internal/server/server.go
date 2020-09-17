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
	return NewGroup(name, cache.NewSyncCacheDefault(cacheBytes))
}

func NewGroup(name string, mainCache *cache.SyncCache) *Group {

	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		mainCache: mainCache,
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]

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
	basePath string
}

func NewHTTPPool() *HTTPPool {
	return &HTTPPool{
		basePath: consts.DefaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Hit] %s", fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	s := r.URL.Path[len(p.basePath)+1:]
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	switch r.Method {
	case http.MethodGet:
		get(groupName, key, w, r)
	case http.MethodPost:
		set(groupName, key, w, r)
	case http.MethodDelete:
		del(groupName, key, w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
func get(groupName string, key string, w http.ResponseWriter, r *http.Request) {
	group := GetGroup(groupName)
	if group == nil {
		group = NewGroupDefault(groupName, 1000)

	}
	valuer, err := group.Get(key)
	if err == nil {
		data := &pb.Data{
			Group:  groupName,
			Value:  valuer.Bytes(),
			Expire: valuer.Expire(),
		}
		bytes, _ := proto.Marshal(&pb.GetResponse{Success: true, Message: "success", Data: data})
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(bytes)
		return
	}

	bytes, _ := proto.Marshal(&pb.GetResponse{Success: false, Message: "fail"})
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(bytes)

}
func set(groupName string, key string, w http.ResponseWriter, r *http.Request) {
	bytesData, err := ioutil.ReadAll(r.Body)
	requestBody := &pb.SetRequest{}
	err = proto.Unmarshal(bytesData, requestBody)
	if err == nil {
		group := GetGroup(groupName)
		if group == nil {
			group = NewGroupDefault(groupName, 1000)

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
			bytes, _ := proto.Marshal(&pb.SetResponse{Success: true, Message: "success", Data: data})
			w.Header().Set("Content-Type", "application/octet-stream")
			_, _ = w.Write(bytes)
			return
		}

	}
	bytes, _ := proto.Marshal(&pb.GetResponse{Success: false, Message: "fail"})
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(bytes)

}
func del(groupName string, key string, w http.ResponseWriter, r *http.Request) {
	message := "fail"
	success := false
	group := GetGroup(groupName)
	if group != nil {
		err := group.Delete(key)
		if err != nil {
		} else {
			success = true
			message = "success"
		}

	}

	bytes, _ := proto.Marshal(&pb.DelResponse{Success: success, Message: message})
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(bytes)

}
