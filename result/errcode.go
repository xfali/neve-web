/*
 * Copyright (C) 2019-2024, Xiongfa Li.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package result

var (
	OK              = Result{Code: 0, Msg: "ok", HttpStatus: 200}
	InternalError   = Result{Code: -1, Msg: "internal error", HttpStatus: 500}
	ConnectError    = Result{Code: 1001, Msg: "connect error", HttpStatus: 500}
	SettingNilError = Result{Code: 1002, Msg: "setting is nil", HttpStatus: 500}
)
