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
	"github.com/chenquan/hit/client/hit"
	"github.com/chenquan/hit/internal/cache/lru"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"
)

var hitClient *Hit

func init() {
	config := &hit.Config{
		Endpoints: []string{"localhost:2379"},
		Replicas:  3,
	}
	hitClient = NewHit(config)
}
func BenchmarkSetAndLocal(b *testing.B) {
	var f GetterFunc = func(string2 string) ([]byte, error) {

		return []byte("not found"), nil
	}

	groupDefault := hitClient.NewGroupDefault("node1", "", 1000, f)
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
	groupDefault := hitClient.NewGroupDefault("groupName1", "", 1000, f)

	rand.Seed(time.Now().Unix())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index := strconv.Itoa(rand.Int())
		_, _ = groupDefault.Set("chenquan"+index, lru.NewValue([]byte("data"), time.Now().Add(time.Minute).Unix(), "test"+strconv.Itoa(rand.Int())), false)
		//_, _ = groupDefault.Get("chenquan" + index)
	}
}
func BenchmarkSetAndGet(b *testing.B) {
	var f GetterFunc = func(string2 string) ([]byte, error) {

		return []byte("not found"), nil
	}
	groupDefault := hitClient.NewGroupDefault("groupName1", "", 1000, f)

	rand.Seed(time.Now().Unix())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index := strconv.Itoa(rand.Int())
		_, _ = groupDefault.Set("chenquan"+index, lru.NewValue([]byte("data"), time.Now().Add(time.Minute).Unix(), "test"+strconv.Itoa(rand.Int())), false)
		_, _ = groupDefault.Get("chenquan" + index)
	}
}

func TestClient(t *testing.T) {
	var f GetterFunc = func(string2 string) ([]byte, error) {

		return []byte("not found"), nil
	}
	groupDefault := hitClient.NewGroupDefault("groupName1", "", 1000, f)

	async.Repeat(context.Background(), time.Second*1, func() {
		rand.Seed(time.Now().Unix())
		index := strconv.Itoa(rand.Int())

		data := []byte("data" + index)
		newValue := lru.NewValue(data, time.Now().Add(time.Minute).Unix(), "groupName1")
		groupDefault.Set("chenquan"+index, newValue, false)
		value, err := groupDefault.Get("chenquan" + index)
		if err == nil {
			fmt.Println("原始数据", newValue.String())
			fmt.Println("获取数据", value.String())
			if !reflect.DeepEqual(value, newValue) {
				panic("数据不相等")
			}
		} else {
			fmt.Println(err)
		}
	})
	select {}
}
