// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package gineve

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xfali/fig"
	"github.com/xfali/neve-core/bean"
	"github.com/xfali/neve-web/gineve/midware/loghttp"
	"github.com/xfali/neve-web/gineve/midware/recovery"
	"github.com/xfali/neve-web/result"
	"github.com/xfali/xlog"
	"net/http"
	"time"
)

const (
	ConfigLogRequestBody = "log.requestBody"
)

//func init() {
//	neve.RegisterProcessor(NewProcessor())
//}

type serverConf struct {
	ContextPath  string
	Port         int
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

type Processor struct {
	conf   fig.Properties
	logger xlog.Logger
	server *http.Server

	compList []Component

	panicHandler recovery.PanicHandler
	httpLogger   loghttp.HttpLogger
	logAll       bool
}

type Opt func(p *Processor)

func NewProcessor(opts ...Opt) *Processor {
	ret := &Processor{
		logger: xlog.GetLogger(),
		panicHandler: func(ctx *gin.Context, err interface{}) {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, result.InternalError)
		},
	}
	for _, v := range opts {
		v(ret)
	}
	return ret
}

func (p *Processor) Init(conf fig.Properties, container bean.Container) error {
	p.conf = conf
	if p.httpLogger == nil {
		p.httpLogger = loghttp.NewFromConfig(conf, p.logger)
	}
	container.Register(p.httpLogger)
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

func (p *Processor) BeanDestroy() error {
	if p.server != nil {
		return p.server.Close()
	}
	return nil
}

func (p *Processor) start(conf fig.Properties) error {
	r := gin.New()
	//r.Use(gin.Logger())
	//r.Use(gin.Recovery())
	panicU := &recovery.RecoveryUtil{
		Logger:       p.logger,
		PanicHandler: p.panicHandler,
	}
	r.Use(panicU.Recovery())
	//logU := &midware.LogHttpUtil{
	//	Logger:      p.logger,
	//	LogRespBody: conf.Get(ConfigLogRequestBody, "false") == "true",
	//}
	if p.logAll {
		r.Use(p.httpLogger.LogHttp())
	}
	//r.Use(logU.LogHttp())
	servConf := serverConf{}
	conf.GetValue("neve.web.server", &servConf)
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

	var router gin.IRouter = r
	if servConf.ContextPath != "" {
		router = router.Group(servConf.ContextPath)
	}
	for _, v := range p.compList {
		v.HttpRoutes(router)
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

func OptSetLogger(logger xlog.Logger) Opt {
	return func(p *Processor) {
		p.logger = logger
	}
}

func OptSetPanicHandler(h recovery.PanicHandler) Opt {
	return func(p *Processor) {
		p.panicHandler = h
	}
}

// 配置默认的HttpLogger
// 如果all为true，则将为所有的接口添加该logger；
// 如果all为false，则用户需要使用inject Logger的方式，手工添加接口日志。
func OptSetDefaultHttpLogger(logger loghttp.HttpLogger, all bool) Opt {
	return func(p *Processor) {
		p.httpLogger = logger
		p.logAll = all
	}
}
