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

package lru

import (
	"container/list"
	"github.com/chenquan/hit/internal/cache/backend/cache"
)

type entry struct {
	key   string
	value cache.Valuer
}
type Cache struct {
	maxBytes     int64                                // 最大内存
	currentBytes int64                                // 当前内存
	ll           *list.List                           // 缓存队列
	cache        map[string]*list.Element             // 缓存字典
	OnEvicted    func(key string, value cache.Valuer) // (可选)移除缓存中某条记录时执行
}

// NewLRUCache 创建LRUCache
func NewLRUCache(maxBytes int64, onEvicted func(string, cache.Valuer)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Len 缓存列表的条数
func (L *Cache) Len() int {
	return L.ll.Len()
}

// Add 添加一个值到缓存中
func (L *Cache) Add(key string, value cache.Valuer) {
	if ele, ok := L.cache[key]; ok {
		L.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		L.currentBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := L.ll.PushFront(&entry{key, value})
		L.cache[key] = ele
		L.currentBytes += int64(len(key)) + int64(value.Len())
	}
	for L.maxBytes != 0 && L.maxBytes < L.currentBytes {
		L.RemoveOldest()
	}
}

// Get 查找键的值
func (L *Cache) Get(key string) (value cache.Valuer, ok bool) {
	if ele, ok := L.cache[key]; ok {
		L.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 删除旧的记录
func (L *Cache) RemoveOldest() {
	ele := L.ll.Back()
	if ele != nil {
		L.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(L.cache, kv.key)
		L.currentBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if L.OnEvicted != nil {
			L.OnEvicted(kv.key, kv.value)
		}
	}
}

type Value struct {
	v []byte
}

func NewValue(v []byte) *Value {
	return &Value{v: v}
}

func (v *Value) Len() int {
	return len(v.v)
}

func (v *Value) Bytes() []byte {

	return cloneBytes(v.v)
}

// cloneBytes 克隆字节码
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
