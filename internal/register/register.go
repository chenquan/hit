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

// 注册服务

package register

import (
	"context"
	"fmt"
	"github.com/chenquan/hit/internal/consts"
	"github.com/etcd-io/etcd/clientv3"
	"log"
	"os"
	"time"
)

// 注册
// 注册
type Config struct {
	Endpoints   []string `json:"endpoints"`    // ETCD节点列表
	LeaseTtl    int64    `json:"lease_ttl"`    // 续租时间
	DialTimeout int64    `json:"dial_timeout"` // 超时时间
	NodeAddr    string   `json:"node_addr"`    // 缓存服务节点地址,列如:192.168.1.11
	NodeName    string   `json:"node_name"`    // 缓存服务节点名称,例如:node1
	Protocol    string   `json:"protocol"`     //协议.目前只支持http
	Port        string   `json:"port"`         //端口.默认:2020
}

func New(config *Config) *Server {

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
	client := &Server{client: cli, name: config.NodeName}
	if err := client.setLease(config.LeaseTtl); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	// 监听etcd租约
	go client.ListenLeaseRespChan()
	return client
}

type Server struct {
	client        *clientv3.Client
	leaseResp     *clientv3.LeaseGrantResponse
	canclefunc    func()
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	name          string
}

//设置租约
func (e *Server) setLease(ttl int64) error {

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

func (e *Server) ListenLeaseRespChan() {
	for {
		select {
		case leaseKeepResp := <-e.keepAliveChan:
			if leaseKeepResp == nil {
				log.Println("已经关闭续租功能.")
				return
			} else {
				log.Printf("续租成功节点:%s.", e.name)
			}
		}
	}
}

func (e *Server) RegisterNode(name, addr string) error {
	name = consts.DefaultEctdPath + name
	log.Println("注册 name:", name, "addr:", addr)
	kv := clientv3.NewKV(e.client)
	_, err := kv.Put(context.TODO(), name, addr, clientv3.WithLease(e.leaseResp.ID))
	return err
}

//撤销租约
func (e *Server) RevokeLease() error {
	e.canclefunc()
	time.Sleep(2 * time.Second)
	_, err := e.client.Lease.Revoke(context.TODO(), e.leaseResp.ID)
	return err
}
