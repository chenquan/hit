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

package cache

import (
	"github.com/chenquan/hit/internal/cache/backend/cache"
	"github.com/chenquan/hit/internal/cache/lru"
	"sync"
)

type SyncCache struct {
	mu sync.RWMutex
	c  cache.Cache
}

func NewSyncCache(c cache.Cache) *SyncCache {
	return &SyncCache{
		c: c,
	}
}
func NewSyncCacheDefault(cacheBytes int64) *SyncCache {
	return NewSyncCache(lru.NewLRUCache(cacheBytes, nil))
}

func (s *SyncCache) RemoveOldest() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.c.RemoveOldest()
}

func (s *SyncCache) Add(key string, value cache.Valuer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.c.Add(key, value)
}

func (s *SyncCache) Get(key string) (value cache.Valuer, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if v, ok := s.c.Get(key); ok {
		return v.(cache.Valuer), ok
	}
	return
} // Len 缓存列表的条数
func (s *SyncCache) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.c.Len()
}
