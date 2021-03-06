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

package cache

import "fmt"

type Cache interface {
	Add(key string, valuer Valuer)
	Get(key string) (valuer Valuer, ok bool)
	Remove(key string)
	Clear()
	Len() int
}

//使用Len值计算需要多少字节
type Valuer interface {
	fmt.Stringer
	Len() int
	Bytes() []byte
	Expire() int64
	SetExpire(timestamp int64)
	GroupName() string
}
