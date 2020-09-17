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

package consistenthash

import (
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

// Hash 将字节映射到uint32
type Hash func(data []byte) uint32
type Map struct {
	hash     Hash
	replicas int
	keys     []int // Sorted
	hashMap  map[int]string
	rwm      sync.RWMutex
}

func (m *Map) String() string {
	return fmt.Sprint(m.hashMap)
}

// New creates a Map instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加一些键。
func (m *Map) Add(keys ...string) {
	// 写锁
	m.rwm.Lock()
	defer m.rwm.Unlock()
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Dle 删除一些键。
func (m *Map) Del(keys ...string) {
	// 写锁
	m.rwm.Lock()
	defer m.rwm.Unlock()
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			// 删除hash
			for index, keyHash := range m.keys {
				if keyHash == hash {
					if len(m.keys) > 1 {
						m.keys = append(m.keys[:index], m.keys[index+1:]...)
					} else {
						m.keys = m.keys[:0]
					}
				}
			}
			delete(m.hashMap, hash)
		}
	}
	sort.Ints(m.keys)
}

// Get 获取哈希中与提供的键最接近的项
func (m *Map) Get(key string) string {
	// 读锁
	m.rwm.RLock()
	defer m.rwm.RUnlock()

	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// 二进制搜索满足条件的副本
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
