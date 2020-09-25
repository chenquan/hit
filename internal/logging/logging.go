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

package logging

import (
	"io/ioutil"
	"log"
	"os"
)

// 遗弃的日志记录
var Discard = log.New(ioutil.Discard, "", 0)

// 默认日志
var Logger = NewStdErr()

// LogError 将错误记录为字符串
func LogError(context string, action string, err error) {
	Logger.Printf("[HIT %s] error during %s (%s)", context, action, err.Error())
}

// LogAction
func LogAction(context string, action string) {
	Logger.Printf("[HIT %s] %s", context, action)
}

// LogTarget
func LogTarget(context, action string, target interface{}) {
	Logger.Printf("[HIT %s] %s (%v)", context, action, target)
}

type Logging interface {
	// Printf 格式化行日志
	Printf(format string, v ...interface{})
}

var _ Logging = new(stderrLogger)

// 错误日志
type stderrLogger log.Logger

func NewStdErr() Logging {
	return (*stderrLogger)(log.New(os.Stderr, "", log.LstdFlags))
}

func (s *stderrLogger) Printf(format string, v ...interface{}) {
	(*log.Logger)(s).Printf(format+"\n", v...)
}
