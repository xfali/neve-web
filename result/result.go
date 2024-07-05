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

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
)

type Result struct {
	Code int64       `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data,omitempty"`

	Err error `json:"error,omitempty"`

	HttpStatus int `json:"-"`
}

func Ok(data interface{}) Result {
	return Result{Code: OK.Code, Msg: OK.Msg, Data: data, HttpStatus: OK.HttpStatus}
}

func (result *Result) WriteJson(ctx *gin.Context) {
	ctx.JSON(result.HttpStatus, *result)
}

func (result *Result) SetCode(code int64) *Result {
	result.Code = code
	return result
}

func (result *Result) SetMessage(v string) *Result {
	result.Msg = v
	return result
}

func (result *Result) SetData(v interface{}) *Result {
	result.Data = v
	return result
}

func (result *Result) SetError(v error) *Result {
	result.Err = v
	return result
}

func (result *Result) SetHttpStatus(v int) *Result {
	result.HttpStatus = v
	return result
}

func (result *Result) Clone() *Result {
	ret := *result
	return &ret
}

func (result *Result) Error() string {
	return result.String()
}

func (result *Result) String() string {
	b, _ := json.Marshal(result)
	return string(b)
}
