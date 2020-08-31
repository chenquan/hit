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

package async

import (
	"context"
	"fmt"
	"github.com/chenquan/hit/internal/logging"
	"runtime/debug"
	"time"
)

// Repeat 异步重复执行以预定的间隔的动作
func Repeat(ctx context.Context, interval time.Duration, action func()) context.CancelFunc {

	// 创建取消上下文
	ctx, cancel := context.WithCancel(ctx)
	safeAction := func() {
		defer handlePanic()
		action()
	}

	// 首次同步执行操作
	safeAction()
	timer := time.NewTicker(interval)
	go func() {

		for {
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				safeAction()
			}
		}
	}()

	return cancel
}

// handlePanic 处理panic并以日志输出
func handlePanic() {
	if r := recover(); r != nil {
		logging.LogAction("async", fmt.Sprintf("panic recovered: %ss \n %s", r, debug.Stack()))
	}
}
