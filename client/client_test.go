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
	"github.com/etcd-io/etcd/clientv3"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"101.132.38.180:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		// handle error!
		fmt.Println("连接失败")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	_, err = cli.Put(ctx, "chenquan", "chenquan的PC")
	cancel()
	if err != nil {
		fmt.Println("存储失败")
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	getResponse, err := cli.Get(ctx, "chenquan")
	cancel()

	if err != nil {
		fmt.Println("获取失败")
	}
	fmt.Println(string(getResponse.Kvs[0].Value))

	defer cli.Close()
}
