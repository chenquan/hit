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
	"github.com/chenquan/hit/internal/async"
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestStep(t *testing.T) {
	Step("")

	async.Repeat(context.Background(), time.Second*5, func() {
		err := Client.RegisterNode("node"+strconv.Itoa(rand.Int()), "http://localhost/"+strconv.Itoa(rand.Int()))
		if err != nil {
			log.Println(err)
		}
	})
	select {}
}
