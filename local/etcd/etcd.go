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

package etcd

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	utils "github.com/chenquan/go-utils"
	"github.com/chenquan/hit/local"
	"github.com/etcd-io/etcd/clientv3"
	"os"
)

var Client etcd

func Step(path string) {
	var etcdConfig clientv3.Config
	if path == "" {
		path = "etcd.toml"
	}
	if _, err := toml.DecodeFile(path, &etcdConfig); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	cli, err := clientv3.New(etcdConfig)
	if err != nil {
		fmt.Println("Error Open", etcdConfig.Endpoints, err)
		os.Exit(0)
	}

	Client = etcd{client: cli, ctx: context.Background(), kv: clientv3.NewKV(cli)}
}

type etcd struct {
	client *clientv3.Client
	ctx    context.Context
	kv     clientv3.KV
}

func (e *etcd) Close() {
	_ = e.client.Close()
}

var _ local.Local = &etcd{}

func (e *etcd) Get(key string) ([]byte, error) {
	//if value, err := e.kv.Get(e.ctx, key);err!=nil{
	//	return nil,false
	//}else {
	//}
	if response, err := e.kv.Get(e.ctx, key); err != nil {
		return nil, err
	} else {
		fmt.Println(response)
	}

	panic("implement me")

}
func (e *etcd) Put(key string, value []byte) error {
	_, err := e.client.Put(e.ctx, key, utils.ToString(&value))
	return err
}
