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

package client

import (
	"context"
	"fmt"
	"github.com/chenquan/go-utils/async"
	"github.com/chenquan/hit/internal/cache/lru"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func BenchmarkSetAndLocal(b *testing.B) {
	var f GetterFunc = func(string2 string) ([]byte, error) {

		return []byte("not found"), nil
	}
	groupDefault := NewGroupDefault("node1", 1000, f)
	rand.Seed(time.Now().Unix())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		index := strconv.Itoa(rand.Int())
		_, _ = groupDefault.Set("chenquan"+index, lru.NewValue([]byte("data"), time.Now().Add(time.Minute).Unix(), "test"+strconv.Itoa(rand.Int())), true)
		//_, _ = groupDefault.Get("chenquan" + index)
	}
}
func BenchmarkSet(b *testing.B) {
	var f GetterFunc = func(string2 string) ([]byte, error) {

		return []byte("not found"), nil
	}
	groupDefault := NewGroupDefault("node1", 1000, f)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rand.Seed(time.Now().Unix())
		index := strconv.Itoa(rand.Int())
		_, _ = groupDefault.Set("chenquan"+index, lru.NewValue([]byte("data"), time.Now().Add(time.Minute).Unix(), "test"+strconv.Itoa(rand.Int())), false)
		//_, _ = groupDefault.Get("chenquan" + index)
	}
}

func TestClient(t *testing.T) {
	var f GetterFunc = func(string2 string) ([]byte, error) {

		return []byte("not found"), nil
	}
	groupDefault := NewGroupDefault("node1", 1000, f)

	async.Repeat(context.Background(), time.Second*6, func() {
		rand.Seed(time.Now().Unix())
		index := strconv.Itoa(rand.Int())
		groupDefault.Set("chenquan"+index, lru.NewValue([]byte("data"), time.Now().Add(time.Minute).Unix(), "test"+strconv.Itoa(rand.Int())), false)
		value, err := groupDefault.Get("chenquan" + index)
		if err == nil {
			fmt.Println("获取数据", value.String())
		} else {
			fmt.Println(err)
		}
	})
	select {}
}
