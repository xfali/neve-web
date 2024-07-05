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

package loghttp

func OptLogReqHeader(flag bool) LogOpt {
	return func(setter Setter) {
		setter.Set(LogReqHeaderKey, flag)
	}
}

func EnableLogReqHeader() LogOpt {
	return OptLogReqHeader(true)
}

func DisableLogReqHeader() LogOpt {
	return OptLogReqHeader(false)
}

func OptLogReqBody(flag bool) LogOpt {
	return func(setter Setter) {
		setter.Set(LogReqBodyKey, flag)
	}
}

func EnableLogReqBody() LogOpt {
	return OptLogReqBody(true)
}

func DisableLogReqBody() LogOpt {
	return OptLogReqBody(false)
}

func OptLogRespHeader(flag bool) LogOpt {
	return func(setter Setter) {
		setter.Set(LogRespHeaderKey, flag)
	}
}

func EnableLogRespHeader() LogOpt {
	return OptLogRespHeader(true)
}

func DisableLogRespHeader() LogOpt {
	return OptLogRespHeader(false)
}

func OptLogRespBody(flag bool) LogOpt {
	return func(setter Setter) {
		setter.Set(LogRespBodyKey, flag)
	}
}

func EnableLogRespBody() LogOpt {
	return OptLogRespBody(true)
}

func DisableLogRespBody() LogOpt {
	return OptLogRespBody(false)
}

func OptLogLevel(lv string) LogOpt {
	return func(setter Setter) {
		setter.Set(LogLevelKey, lv)
	}
}
