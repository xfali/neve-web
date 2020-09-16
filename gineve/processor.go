// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package gineve

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/xfali/fig"
	"github.com/xfali/neve/neve-core/container"
	"github.com/xfali/neve/neve-utils/log"
	"github.com/xfali/neve/neve-web/gineve/midware"
	"github.com/xfali/neve/neve-web/result"
	"github.com/xfali/xlog"
	"net/http"
	"time"
)

const (
	ConfigLogRequestBody = "Log.RequestBody"
)

//func init() {
//	neve.RegisterProcessor(NewProcessor())
//}

type Processor struct {
	conf   fig.Properties
	logger xlog.Logger
	server *http.Server

	compList []Component
}

func NewProcessor() *Processor {
	return &Processor{
		logger: log.GetLogger(),
	}
}

func (p *Processor) Init(conf fig.Properties, container container.Container) error {
	p.conf = conf
	return nil
}

func (p *Processor) Classify(o interface{}) (bool, error) {
	switch v := o.(type) {
	case Component:
		err := p.parseBean(v, o)
		return true, err
	}
	return false, nil
}

func (p *Processor) Process() error {
	return p.start(p.conf)
}

func (p *Processor) Close() error {
	if p.server != nil {
		return p.server.Close()
	}
	return nil
}

func (p *Processor) start(conf fig.Properties) error {
	if conf.Get(ConfigLogRequestBody, "false") == "true" {
		binding.Validator = &midware.RequestBodyLogWriter{V: binding.Validator}
	}

	r := gin.New()
	//r.Use(gin.Logger())
	//r.Use(gin.Recovery())
	panicU := &midware.RecoveryUtil{
		Logger: p.logger,
		PanicHandler: func(ctx *gin.Context, err interface{}) {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, result.InternalError)
		},
	}
	r.Use(panicU.Recovery())
	//logU := &midware.LogHttpUtil{
	//	Logger:      p.logger,
	//	LogRespBody: conf.Get(ConfigLogRequestBody, "false") == "true",
	//}
	//r.Use(logU.LogHttp())

	for _, v := range p.compList {
		Register(r)
	}

	type serverConf struct {
		Port         int
		ReadTimeout  int
		WriteTimeout int
		IdleTimeout  int
	}
	servConf := serverConf{}
	conf.GetValue("Server", &servConf)
	if servConf.Port == 0 {
		servConf.Port = 8080
	}
	if servConf.ReadTimeout == 0 {
		servConf.ReadTimeout = 15
	}
	if servConf.WriteTimeout == 0 {
		servConf.WriteTimeout = 15
	}
	if servConf.IdleTimeout == 0 {
		servConf.IdleTimeout = 15
	}
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", servConf.Port),
		Handler:        r,
		ReadTimeout:    time.Duration(servConf.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(servConf.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(servConf.IdleTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go s.ListenAndServe()

	p.server = s

	return nil
}

func (p *Processor) parseBean(comp Component, o interface{}) error {
	p.compList = append(p.compList, comp)

	return nil
}
