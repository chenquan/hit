/*
 *    Copyright  2020 Chen Quan
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

package remote

import "sync"

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Loader struct {
	mu sync.Mutex // protects m
	m  map[string]*call
}

func (l *Loader) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	l.mu.Lock()
	if l.m == nil {
		l.m = make(map[string]*call)
	}
	if c, ok := l.m[key]; ok {
		l.mu.Unlock()
		// 等待先入协程从远端获取数据
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	l.m[key] = c
	l.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	l.mu.Lock()
	delete(l.m, key)
	l.mu.Unlock()

	return c.val, c.err
}
