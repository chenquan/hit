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
	"fmt"
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
func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

// Add 添加一个值到缓存中
func (c *Cache) Add(key string, value cache.Valuer) {
	if c.cache == nil {
		c.cache = make(map[string]*list.Element)
		c.ll = list.New()
	}

	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.currentBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.currentBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.currentBytes {
		c.removeOldest()
	}
}

// Get 查找键的值
func (c *Cache) Get(key string) (value cache.Valuer, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 删除旧的记录
func (c *Cache) removeOldest() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.currentBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

//
func (c *Cache) removeElement(e *list.Element) {
	if c.cache == nil {
		return
	}
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	c.currentBytes -= int64(len(kv.key)) + int64(kv.value.Len())
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// Remove 移除指定key的数据
func (c *Cache) Remove(key string) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

// Clear 清空全部数据
func (c *Cache) Clear() {
	if c.cache == nil {
		return
	}
	if c.OnEvicted != nil {
		for _, e := range c.cache {
			kv := e.Value.(*entry)
			c.OnEvicted(kv.key, kv.value)
		}
	}
	c.ll = nil
	c.cache = nil
}

type Value struct {
	data      []byte // 数据
	expire    int64  // 数据到期时间戳
	groupName string // 分组名称
}

func NewValue(data []byte, expire int64, groupName string) *Value {
	return &Value{data: data, expire: expire, groupName: groupName}
}

func (v *Value) Len() int {
	return len(v.data)
}

func (v *Value) Bytes() []byte {

	return cloneBytes(v.data)
}
func (v *Value) Expire() int64 {
	return v.expire
}
func (v *Value) SetExpire(timestamp int64) {
	v.expire = timestamp
}
func (v *Value) GroupName() string {
	return v.groupName
}
func (v *Value) String() string {

	return fmt.Sprintf("{data:%s,expire:%d,groupName:%s}", v.data, v.expire, v.groupName)

}

// cloneBytes 克隆字节码
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
