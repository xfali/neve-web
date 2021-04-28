// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

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

