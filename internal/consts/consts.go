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

package consts

import "time"

// 默认路径
const (
	DefaultPath               = "hit/"
	ContentType               = "application/octet-stream"
	DefaultLocalCacheDuration = time.Second      // 默认本地缓存时长
	DefaultNodeCacheDuration  = time.Second * 60 // 默认节点缓存时长
)
