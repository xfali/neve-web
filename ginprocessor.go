// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package neveweb

import (
	"github.com/gin-gonic/gin"
	"github.com/xfali/neve-web/gineve"
	"github.com/xfali/neve-web/gineve/midware/loghttp"
	"github.com/xfali/neve-web/gineve/midware/recovery"
	"github.com/xfali/xlog"
)

type ginOpts struct{}

var GinOpt ginOpts

func NewGinProcessor(opts ...gineve.Opt) *gineve.Processor {
	return gineve.NewProcessor(opts...)
}

func (opt ginOpts) WithLogger(logger xlog.Logger) gineve.Opt {
	return gineve.OptSetLogger(logger)
}

func (opt ginOpts) WithPanicHandler(h recovery.PanicHandler) gineve.Opt {
	return gineve.OptSetPanicHandler(h)
}

// 配置默认的HttpLogger
// 如果all为true，则将为所有的接口添加该logger；
// 如果all为false，则用户需要使用inject Logger的方式，手工添加接口日志。
func (opt ginOpts) WithHttpLogger(logger loghttp.HttpLogger, all bool) gineve.Opt {
	return gineve.OptSetDefaultHttpLogger(logger, all)
}

func (opt ginOpts) AddFilters(filters ...gin.HandlerFunc) gineve.Opt {
	return gineve.OptAddFilters(filters...)
}
