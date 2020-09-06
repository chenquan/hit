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
	"github.com/chenquan/hit/internal/async"
	"testing"
	"time"
)

func TestStep(t *testing.T) {
	Step("")
	if nodes, err := Client.PullAllNodes(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(nodes)
	}
	async.Repeat(context.Background(), time.Second*10, func() {
		nodes := Client.GetNodes()

		fmt.Println(nodes)
	})
	select {}

}
