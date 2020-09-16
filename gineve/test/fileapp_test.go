// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xfali/fig"
	"github.com/xfali/neve/neve-core/container"
	"github.com/xfali/neve/neve-utils/log"
	"github.com/xfali/neve/neve-core"
	"github.com/xfali/neve/neve-core/processor"
	"github.com/xfali/neve/neve-web/gineve"
	"github.com/xfali/neve/neve-web/gineve/midware"
	"github.com/xfali/neve/neve-web/result"
	"net/http"
	"testing"
)

type webBean struct {
	V string `fig:"Log.Level"`
	P print  `inject:"testProcess.print"`
}

func (b *webBean) Register(engine gin.IRouter) {
	loghttp := midware.LogHttpUtil{
		Logger:      log.GetLogger(),
		LogRespBody: true,
	}
	engine.GET("test", loghttp.LogHttp(), func(context *gin.Context) {
		context.JSON(http.StatusOK, result.Ok(b.V))
	})

	engine.GET("panic", loghttp.LogHttp(), func(context *gin.Context) {
		panic("test!")
	})
}

func TestWebAndValue(t *testing.T) {
	neve.RegisterProcessor(gineve.NewProcessor(), processor.NewValueProcessor(), &testProcess{})

	app := neve.NewFileConfigApplication("assets/config-test.json")
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

func (p *testProcess) Init(conf fig.Properties, container container.Container) error {
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
