// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xfali/fig"
	"github.com/xfali/neve-core"
	"github.com/xfali/neve-core/bean"
	"github.com/xfali/neve-core/processor"
	"github.com/xfali/neve-web/gineve"
	"github.com/xfali/neve-web/gineve/midware"
	"github.com/xfali/neve-web/result"
	"net/http"
	"testing"
)

type webBean struct {
	V          string //`fig:"Log.Level"`
	P          print  `inject:"testProcess.print"`
	HttpLogger midware.HttpLogger `inject:""`
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
}

func TestWebAndValue(t *testing.T) {
	app := neve.NewFileConfigApplication("assets/config-test.yaml")
	app.RegisterBean(gineve.NewProcessor())
	app.RegisterBean(processor.NewValueProcessor())
	app.RegisterBean(&testProcess{})
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
