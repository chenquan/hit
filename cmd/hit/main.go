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
	"github.com/chenquan/hit"
	remotehttp "github.com/chenquan/hit/remote/http"
	"log"
	"net/http"
	"strings"
)

var db = map[string]string{
	"Tom":      "630",
	"Jack":     "589",
	"Sam":      "567",
	"chenquan": "chenquan",
	"zero":     "zero",
}

func main() {
	var addr string
	var apiPort string
	var api bool
	var peerAddrs []string
	var groups []string

	flag.StringVar(&addr, "addr", "http://localhost:8001", "Protoc server port")
	flag.StringVar(&apiPort, "apiPort", "9999", "Server api port")
	flag.Var(newSliceValue([]string{"http://localhost:8001", "http://localhost:8002", "http://localhost:8003"}, &peerAddrs),
		"peerAddrs",
		"Peer node addrs")
	flag.Var(newSliceValue([]string{"default", "tmp"}, &groups),
		"groups",
		"Group cache names")
	flag.BoolVar(&api, "api", true, "Start a api server?")
	flag.Parse()
	for _, name := range groups {
		createNewGroupDefault(name)
	}
	if api {
		go startAPIServer(apiPort)
	}
	startCacheServer(addr, peerAddrs, groups)

}
func createNewGroupDefault(name string) *hit.Group {
	return hit.NewGroupDefault(name, 2<<10, hit.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}
func startCacheServer(addr string, peerAddrs []string, groups []string) {

	peers := remotehttp.NewHTTPPool(addr)
	peers.Set(peerAddrs...)
	for _, group := range groups {
		g := hit.GetGroup(group)
		log.Println("[Hit] group:", group)
		g.RegisterPeers(peers)
	}
	log.Println("[Hit] running at", addr)
	ports := strings.SplitN(addr, ":", 3)
	if len(ports) != 3 {
		log.Fatalln("[Hit] ERROR: cache server Addr", addr)
	}
	log.Fatal(http.ListenAndServe(":"+ports[2], peers))
}
func startAPIServer(apiPort string) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			group := r.URL.Query().Get("group")
			if group == "" {
				group = "default"
			}
			g := hit.GetGroup(group)
			//  查询
			if r.Method == "GET" {
				key := r.URL.Query().Get("key")

				if key == "" {
					http.Error(w, "key is required", http.StatusBadRequest)
					return
				}

				if g == nil {
					http.Error(w, "group not found", http.StatusBadRequest)
					return
				}
				view, err := g.Get(key)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Write(view.ByteSlice())
			}

			// TODO 新增
			if r.Method == "POST" {

			}
		}))

	log.Println("[Hit] api server port is running at http://localhost:" + apiPort)

	log.Fatal(http.ListenAndServe(":"+apiPort, nil))

}

type stringSlice []string

func newSliceValue(vals []string, p *[]string) *stringSlice {
	*p = vals
	return (*stringSlice)(p)
}
func (s *stringSlice) Set(val string) error {
	*s = strings.Split(val, ",")
	return nil
}

func (s *stringSlice) Get() interface{} { return []string(*s) }

func (s *stringSlice) String() string { return strings.Join(*s, ",") }
