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
	"github.com/chenquan/hit/cache"
	"github.com/chenquan/hit/cache/lru"
	"sync"
)

type SyncCache struct {
	mu sync.Mutex
	c  cache.Cache
}

func NewSyncCache(c cache.Cache) *SyncCache {
	return &SyncCache{
		c: c,
	}
}
func NewSyncCacheDefault(cacheBytes int64) *SyncCache {
	return &SyncCache{
		c: lru.NewLRUCache(cacheBytes, nil),
	}
}

func (s *SyncCache) Add(key string, value cache.Value) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.c.Add(key, value)
}

func (s *SyncCache) Get(key string) (value ReadBytes, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if v, ok := s.c.Get(key); ok {
		return v.(ReadBytes), ok
	}
	return
}
