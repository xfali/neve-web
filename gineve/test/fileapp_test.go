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

package test

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xfali/fig"
	"github.com/xfali/neve-core"
	"github.com/xfali/neve-core/bean"
	"github.com/xfali/neve-core/processor"
	"github.com/xfali/neve-web"
	"github.com/xfali/neve-web/gineve"
	"github.com/xfali/neve-web/gineve/midware/loghttp"
	"github.com/xfali/neve-web/result"
	"github.com/xfali/xlog"
	"net/http"
	"testing"
)

type webBean struct {
	V          string             //`fig:"Log.Level"`
	P          print              `inject:"testProcess.print"`
	HttpLogger loghttp.HttpLogger `inject:""`
}

func (b *webBean) HttpRoutes(engine gin.IRouter) {
	//loghttp := midware.LogHttpUtil{
	//	Logger:        log.GetLogger(),
	//	LogReqBody:    true,
	//	LogReqHeader:  true,
	//	LogRespBody:   true,
	//	LogRespHeader: true,
	//}
	engine.GET("test", b.HttpLogger.LogHttp(), func(context *gin.Context) {
		context.JSON(http.StatusOK, result.Ok(b.V))
	})

	engine.GET("panic", b.HttpLogger.LogHttp(), func(context *gin.Context) {
		panic("test!")
	})

	engine.POST("test", b.HttpLogger.LogHttp(), func(context *gin.Context) {
		d, err := context.GetRawData()
		if err != nil {
			context.AbortWithStatus(http.StatusBadRequest)
			return
		}

		context.JSON(http.StatusOK, result.Ok(string(d)))
	})

	engine.POST("/no/req/header", b.HttpLogger.OptLogHttp(loghttp.DisableLogReqHeader()), func(context *gin.Context) {
		d, err := context.GetRawData()
		if err != nil {
			context.AbortWithStatus(http.StatusBadRequest)
			return
		}

		context.Writer.WriteString(string(d))
	})

	engine.POST("/no/req/body", b.HttpLogger.OptLogHttp(loghttp.DisableLogReqBody()), func(context *gin.Context) {
		d, err := context.GetRawData()
		if err != nil {
			context.AbortWithStatus(http.StatusBadRequest)
			return
		}

		context.Writer.WriteString(string(d))
	})

	engine.POST("/no/resp/header", b.HttpLogger.OptLogHttp(loghttp.DisableLogRespHeader()), func(context *gin.Context) {
		d, err := context.GetRawData()
		if err != nil {
			context.AbortWithStatus(http.StatusBadRequest)
			return
		}

		context.Writer.WriteString(string(d))
	})

	engine.POST("/no/resp/body", b.HttpLogger.OptLogHttp(loghttp.DisableLogRespBody()), func(context *gin.Context) {
		d, err := context.GetRawData()
		if err != nil {
			context.AbortWithStatus(http.StatusBadRequest)
			return
		}

		context.Writer.WriteString(string(d))
	})

	engine.POST("/error", b.HttpLogger.OptLogHttp(loghttp.OptLogLevel("error")), func(context *gin.Context) {
		d, err := context.GetRawData()
		if err != nil {
			context.AbortWithStatus(http.StatusBadRequest)
			return
		}
		xlog.Infoln(string(d))

		context.Writer.WriteString(string(d))
	})
}

func TestWebAndValue(t *testing.T) {
	app := neve.NewFileConfigApplication("assets/config-test.yaml")
	app.RegisterBean(neveweb.NewGinProcessor())
	app.RegisterBean(processor.NewValueProcessor())
	app.RegisterBean(&testProcess{})
	app.RegisterBean(&webBean{})
	app.Run()
}

type filter struct{}

func (f *filter) FilterHandler(context *gin.Context) {
	v, have := context.Get("hello")
	if !have {
		xlog.Panic("not have!")
	}
	if v.(string) != "world" {
		xlog.Panic("not match!")
	}
	xlog.Infoln(v)
	if context.FullPath() == "/panic" {
		context.Abort()
	} else {
		context.Next()
	}
}

func TestWebFilters(t *testing.T) {
	app := neve.NewFileConfigApplication("assets/config-test.yaml")
	app.RegisterBean(gineve.NewProcessor(gineve.OptAddFilters(func(context *gin.Context) {
		context.Set("hello", "world")
		context.Next()
	})))
	app.RegisterBean(processor.NewValueProcessor())
	app.RegisterBean(&filter{})
	app.RegisterBean(&webBean{})
	app.Run()
}

type testProcess struct{}

type print interface {
	Print(str string)
}

type dummy struct{}

func (d *dummy) Print(str string) {
	fmt.Println("dummy!", str)
}

func (p *testProcess) Init(conf fig.Properties, container bean.Container) error {
	container.RegisterByName("testProcess.print", &dummy{})
	return nil
}

func (p *testProcess) Classify(o interface{}) (bool, error) {
	switch v := o.(type) {
	case print:
		v.Print("test")
	}
	return true, nil
}

func (p *testProcess) Process() error {
	return nil
}

func (p *testProcess) Close() error {
	return nil
}
