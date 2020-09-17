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
	"bytes"
	"fmt"
	"github.com/chenquan/hit/internal/consts"
	pb "github.com/chenquan/hit/internal/remotecache"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"net/http"
	"net/url"
)

// 远程节点
type Node struct {
	url string
}

func NewNode(url string) *Node {
	return &Node{url: url}
}

func (h *Node) Set(in *pb.SetRequest, out *pb.SetResponse) error {
	u := fmt.Sprintf(
		"%v/%v/%v",
		h.url,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	requestBytes, _ := proto.Marshal(in)
	res, err := http.Post(u, consts.ContentType, bytes.NewBuffer(requestBytes))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("register returned: %v", res.Status)
	}

	bytesData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}
	if err = proto.Unmarshal(bytesData, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	if out.Success {
		return nil
	} else {
		return fmt.Errorf("message: %s", out.Message)
	}
}

// 从远程节点获取数据
func (h *Node) Get(in *pb.GetRequest, out *pb.GetResponse) error {

	u := fmt.Sprintf(
		"%v/%v/%v",
		h.url,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("register returned: %v", res.Status)
	}

	bytesData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}
	if err = proto.Unmarshal(bytesData, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	if out.Success {
		return nil
	} else {
		return fmt.Errorf("message: %s", out.Message)
	}
}
func (h *Node) Del(in *pb.DelRequest, out *pb.DelResponse) error {
	u := fmt.Sprintf(
		"%v/%v/%v",
		h.url,
		url.QueryEscape(in.GetKey()),
		url.QueryEscape(in.GetGroup()),
	)
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	bytesBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(bytesBody, out)
	if err != nil {
		return err
	}
	return nil

}

// 获取远程节点地址
func (h *Node) Url() string {
	return h.url
}
