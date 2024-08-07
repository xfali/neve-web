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
	Host         string
	Port         int
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int

	Tls tlsConf
}

type tlsConf struct {
	Cert string
	Key  string
}

type Processor struct {
	conf   fig.Properties
	logger xlog.Logger
	server *http.Server

	compList []Component

	filters gin.HandlersChain

	panicHandler recovery.PanicHandler
	httpLogger   loghttp.HttpLogger

	srvModifier ServerModifier
	logAll      bool
}

type ServerModifier func(srv *http.Server, engine *gin.Engine)

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
		return true, p.parseBean(v)
	case Filter:
		return true, p.parseFilter(v)
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

	if p.panicHandler != nil {
		panicU := &recovery.RecoveryUtil{
			Logger:       p.logger,
			PanicHandler: p.panicHandler,
		}
		r.Use(panicU.Recovery())
	}
	if p.logAll {
		r.Use(p.httpLogger.LogHttp())
	}

	if len(p.filters) > 0 {
		r.Use(p.filters...)
	}

	servConf := serverConf{}
	err := conf.GetValue("neve.web.server", &servConf)
	if err != nil {
		return err
	}

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

	addr := getServeAddr(servConf)
	s := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    time.Duration(servConf.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(servConf.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(servConf.IdleTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if p.srvModifier != nil {
		p.srvModifier(s, r)
	}

	go func() {
		if servConf.Tls.Cert == "" {
			err := s.ListenAndServe()
			if err != nil {
				p.logger.Errorln(err)
			}
		} else {
			err := s.ListenAndServeTLS(servConf.Tls.Cert, servConf.Tls.Key)
			if err != nil {
				p.logger.Errorln(err)
			}
		}
	}()

	p.server = s

	return nil
}

func (p *Processor) parseBean(comp Component) error {
	p.compList = append(p.compList, comp)
	return nil
}

func (p *Processor) parseFilter(filter Filter) error {
	p.filters = append(p.filters, filter.FilterHandler)
	return nil
}

func getServeAddr(servConf serverConf) string {
	//scheme := u.Scheme
	//if scheme == "" {
	//	if servConf.Tls.Cert == "" {
	//		scheme = "http"
	//	} else {
	//		scheme = "https"
	//	}
	//}
	//if u.Host != "" {
	//	u.Port()
	//	return u.Host
	//}
	//fmt.Sprintf("%s://%s:%d", scheme, servConf.Host, servConf.Port)
	return fmt.Sprintf("%s:%d", servConf.Host, servConf.Port)
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

func OptAddFilters(filters ...gin.HandlerFunc) Opt {
	return func(p *Processor) {
		p.filters = append(p.filters, filters...)
	}
}

func OptSetServerModifier(m ServerModifier) Opt {
	return func(p *Processor) {
		p.srvModifier = m
	}
}
