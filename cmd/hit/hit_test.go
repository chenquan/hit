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

package main

import (
	"fmt"
	"reflect"
	"testing"
)

type Test interface {
	Name() string
	SetName(string)
}
type Atest struct {
	value string
}

func (t *Atest) Name() string {
	return t.value
}
func (t *Atest) SetName(value string) {
	t.value = value
}

func Test1(t *testing.T) {
	var test1 Test
	test1 = &Atest{
		value: "11",
	}
	fmt.Println(test1)
	var test2 Test
	tmp := test1.(*Atest)
	tmp2 := *tmp
	test2 = &tmp2
	test2.SetName("222")

	test3 := reflect.New(reflect.ValueOf(test1).Elem().Type()).Interface().(Test)
	test3.SetName("333")
	fmt.Println(test3)
	fmt.Println(test2)
	fmt.Println(test1)

}
