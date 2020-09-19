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
	"testing"
)

type String string

func (s String) SetExpire(timestamp int64) {
	panic("implement me")
}

func (s String) GroupName() string {
	panic("implement me")
}

func (s String) Bytes() []byte {
	return []byte(s)
}

func (s String) Len() int {
	return len(s)
}
func (s String) Expire() int64 {
	return 0
}
func (s String) String() string {
	return string(s)
}

func TestGet(t *testing.T) {
	lru := NewSyncCacheDefault(int64(0))
	lru.Add("key1", String("1234"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("lru hit key1=1234 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("lru miss key2 failed")
	}
}

func TestRemoveoldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	capSize := len(k1 + k2 + v1 + v2)
	lru := NewSyncCacheDefault(int64(capSize))
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}
