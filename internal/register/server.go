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

package register

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/etcd-io/etcd/clientv3"
	"log"
	"os"
	"time"
)

const DefaultPath = "hit/"

// 注册
type Config struct {
	Endpoints   []string `json:"endpoints"`    // 节点列表
	LeaseTtl    int64    `json:"lease_ttl"`    // 续租时间
	DialTimeout int64    `json:"dial_timeout"` // 续租时间
}

var Client etcd

func Step(path string) {
	var config Config

	if path == "" {
		path = "etcd.toml"
	}
	if _, err := toml.DecodeFile(path, &config); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// 默认配置
	if config.LeaseTtl == 0 {
		config.LeaseTtl = 10
	}
	if config.DialTimeout == 0 {
		config.DialTimeout = 5
	}
	// 配置etcd客户端
	var etcdConfig = clientv3.Config{
		Endpoints:   config.Endpoints,
		DialTimeout: time.Duration(config.DialTimeout) * time.Second,
	}

	cli, err := clientv3.New(etcdConfig)
	if err != nil {
		fmt.Println("Error Open", etcdConfig.Endpoints, err)
		os.Exit(0)
	}
	Client = etcd{client: cli}
	if err := Client.setLease(config.LeaseTtl); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	// 监听etcd租约
	go Client.ListenLeaseRespChan()
}

type etcd struct {
	client        *clientv3.Client
	leaseResp     *clientv3.LeaseGrantResponse
	canclefunc    func()
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	key           string
}

//设置租约
func (e *etcd) setLease(ttl int64) error {

	//设置租约时间
	leaseResp, err := e.client.Lease.Grant(context.TODO(), ttl)
	if err != nil {
		return err
	}

	//设置续租
	ctx, cancelFunc := context.WithCancel(context.TODO())
	leaseRespChan, err := e.client.Lease.KeepAlive(ctx, leaseResp.ID)

	if err != nil {
		return err
	}

	e.leaseResp = leaseResp
	e.canclefunc = cancelFunc
	e.keepAliveChan = leaseRespChan
	return nil
}

func (e *etcd) ListenLeaseRespChan() {
	for {
		select {
		case leaseKeepResp := <-e.keepAliveChan:
			if leaseKeepResp == nil {
				fmt.Printf("已经关闭续租功能\n")
				return
			} else {
				fmt.Printf("续租成功\n")
			}
		}
	}
}

func (e *etcd) RegisterNode(name, addr string) error {
	name = DefaultPath + name
	log.Println("注册:", name)
	kv := clientv3.NewKV(e.client)
	_, err := kv.Put(context.TODO(), name, addr, clientv3.WithLease(e.leaseResp.ID))
	return err
}

//撤销租约
func (e *etcd) RevokeLease() error {
	e.canclefunc()
	time.Sleep(2 * time.Second)
	_, err := e.client.Lease.Revoke(context.TODO(), e.leaseResp.ID)
	return err
}
