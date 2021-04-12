// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description:

package loghttp

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/xfali/fig"
	"github.com/xfali/goutils/idUtil"
	"github.com/xfali/neve-web/buffer"
	"github.com/xfali/xlog"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"
)

const (
	REQEUST_ID = "_REQEUST_ID"

	LogReqHeaderKey  = "neve.web.Log.RequestHeader"
	LogReqBodyKey    = "neve.web.Log.RequestBody"
	LogRespHeaderKey = "neve.web.Log.ResponseHeader"
	LogRespBodyKey   = "neve.web.Log.ResponseBody"
	LogLevelKey      = "neve.web.Log.Level"
)

type logFunc func(fmt string, args ...interface{})

type LogOpt func(setter Setter)

type HttpLogger interface {
	// 获得按配置初始化的日志handler
	LogHttp() gin.HandlerFunc
	// 按参数配置日志handler
	OptLogHttp(opts ...LogOpt) gin.HandlerFunc
}

type Setter interface {
	Set(key string, value interface{})
}

type requestBodyWrapper struct {
	origin io.ReadCloser
	body   *bytes.Buffer
}

func newRequestBodyWrapper(rc io.ReadCloser) *requestBodyWrapper {
	ret := &requestBodyWrapper{
		origin: rc,
		body:   bytes.NewBuffer(nil),
	}
	ret.body.WriteString(" [data]: ")
	return ret
}

func (w *requestBodyWrapper) purge() {
	w.body.Reset()
	w.body = nil
}

func (w *requestBodyWrapper) Read(b []byte) (int, error) {
	i, err := w.origin.Read(b)
	w.body.Write(b[:i])
	return i, err
}

func (w *requestBodyWrapper) getBody() string {
	w.body.WriteString(" ,")
	return w.body.String()
}

func (w *requestBodyWrapper) Close() error {
	return w.origin.Close()
}

type responseBodyLogWrapper struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyLogWrapper) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseBodyLogWrapper) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func newResponseBodyWrapper(w gin.ResponseWriter) *responseBodyLogWrapper {
	ret := &responseBodyLogWrapper{
		ResponseWriter: w,
		body:           bytes.NewBuffer(nil),
	}
	ret.body.WriteString(" [data]: ")
	return ret
}

func (w *responseBodyLogWrapper) getBody() string {
	return w.body.String()
}

func (w *responseBodyLogWrapper) purge() {
	w.body.Reset()
	w.body = nil
}

type RequestBodyLogWriter struct {
	Logger xlog.Logger
	V      binding.StructValidator
}

func (v *RequestBodyLogWriter) ValidateStruct(obj interface{}) error {
	valueType := reflect.TypeOf(obj)
	if valueType.Kind() == reflect.Ptr {
		valueType = valueType.Elem()
	}
	if valueType.Kind() == reflect.Struct {
		b, err := json.Marshal(obj)
		if err != nil {
			v.Logger.Infof("Request Log bind body error: %s\n", err.Error())
		} else {
			v.Logger.Infof("Request Log bind body is %s\n", string(b))
		}
	}

	return v.V.ValidateStruct(obj)
}

func (v *RequestBodyLogWriter) Engine() interface{} {
	return v.V.Engine()
}

type LogHttpUtil struct {
	Logger xlog.Logger

	ignore struct{} `figPx:"neve.web"`
	// with request header log
	LogReqHeader bool `fig:"Log.RequestHeader"`
	// with request body log
	LogReqBody bool `fig:"Log.RequestBody"`
	// with response header log
	LogRespHeader bool `fig:"Log.ResponseHeader"`
	// with response body log
	LogRespBody bool `fig:"Log.ResponseBody"`
	// log level
	Level string `fig:"Log.Level"`

	logFunc logFunc
}

type DefaultHttpLogger LogHttpUtil

func NewLogHttpUtil(conf fig.Properties, logger xlog.Logger) *LogHttpUtil {
	ret := &LogHttpUtil{
		Logger: logger,
	}
	fig.Fill(conf, ret)
	ret.initLog()
	return ret
}

func (util *LogHttpUtil) LogHttp() gin.HandlerFunc {
	return util.log
}

func (util *LogHttpUtil) OptLogHttp(opts ...LogOpt) gin.HandlerFunc {
	return util.clone(opts...).log
}

func (util *LogHttpUtil) clone(opts ...LogOpt) *LogHttpUtil {
	ret := &LogHttpUtil{}
	ret.Logger = util.Logger
	ret.LogReqHeader = util.LogReqHeader
	ret.LogReqBody = util.LogReqBody
	ret.LogRespHeader = util.LogRespHeader
	ret.LogRespBody = util.LogRespBody
	ret.Level = util.Level

	for _, opt := range opts {
		opt(ret)
	}

	ret.initLog()
	return ret
}

func (util *LogHttpUtil) Set(key string, value interface{}) {
	switch key {
	case LogReqHeaderKey:
		if v, ok := value.(bool); ok {
			util.LogReqHeader = v
		}
		break
	case LogReqBodyKey:
		if v, ok := value.(bool); ok {
			util.LogReqBody = v
		}
		break
	case LogRespHeaderKey:
		if v, ok := value.(bool); ok {
			util.LogRespHeader = v
		}
		break
	case LogRespBodyKey:
		if v, ok := value.(bool); ok {
			util.LogRespBody = v
		}
		break
	case LogLevelKey:
		if v, ok := value.(string); ok {
			util.Level = v
		}
		break
	}
}

func (util *LogHttpUtil) log(c *gin.Context) {
	start := time.Now()

	path := c.Request.URL.Path
	clientIP := c.ClientIP()
	method := c.Request.Method
	requestId := idUtil.RandomId(16)
	params := c.Params
	querys := c.Request.URL.RawQuery
	reqHeader := ""
	if util.LogReqHeader {
		reqHeader = getHeaderStr(c.Request.Header)
	}

	//c.Set(REQEUST_ID, requestId)

	var rbw *requestBodyWrapper = nil
	if util.LogReqBody {
		rbw = newRequestBodyWrapper(c.Request.Body)
		c.Request.Body = rbw
		defer rbw.purge()
	}

	var blw *responseBodyLogWrapper
	if util.LogRespBody {
		blw = newResponseBodyWrapper(c.Writer)
		c.Writer = blw
		defer blw.purge()
	}

	// 处理请求
	c.Next()

	reqBody := ""
	if util.LogReqBody {
		reqBody = rbw.getBody()
	}
	//util.Logger.Infof("[Request %s] [path]: %s , [client ip]: %s , [method]: %s , %s , [params]: %v , [query]: %s %s\n",
	//	requestId, path, clientIP, method, reqHeader, params, querys, reqBody)

	// 结束时间
	end := time.Now()
	//执行时间
	latency := end.Sub(start)

	statusCode := c.Writer.Status()

	var data string
	if util.LogRespBody {
		data = blw.getBody()
	}
	respHeader := ""
	if util.LogRespHeader {
		rh := c.Writer.Header()
		if rh != nil {
			respHeader = getHeaderStr(rh.Clone())
		}
	}
	//respId, _ := c.Get(REQEUST_ID)
	util.output("\n[Request\t%s] [path]: %s , [client ip]: %s , [method]: %s %s [params]: %v , [query]: %s%s\n"+
		"[Response\t%s] [latency]: %d ms, [status]: %d %s%s\n",
		requestId, path, clientIP, method, reqHeader, params, querys, reqBody,
		requestId, latency/time.Millisecond, statusCode, respHeader, data)
	//util.Logger.Infof("[Response %s] [latency]: %d ms, [status]: %d , %s , [data]: %s\n",
	//	respId.(string), latency/time.Millisecond, statusCode, respHeader, data)
}

func getHeaderStr(header http.Header) string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(", [header]: ")
	if len(header) > 0 {
		for k, vs := range header {
			buf.WriteString(k)
			buf.WriteString("=")
			for i := range vs {
				buf.WriteString(vs[i])
				if i < len(vs)-1 {
					buf.WriteString(",")
				}
			}
			buf.WriteString(" ")
		}
	}
	buf.WriteString(" ,")
	return buf.String()
}

func (util *LogHttpUtil) output(fmt string, args ...interface{}) {
	if util.logFunc == nil {
		util.Logger.Infof(fmt, args...)
	} else {
		util.logFunc(fmt, args...)
	}
}

func (util *LogHttpUtil) initLog() {
	lv := strings.ToLower(util.Level)
	switch lv {
	case "debug":
		util.logFunc = util.Logger.Debugf
	case "info":
		util.logFunc = util.Logger.Infof
	case "warn":
		util.logFunc = util.Logger.Warnf
	case "error":
		util.logFunc = util.Logger.Errorf
	case "panic":
		util.logFunc = util.Logger.Panicf
	case "fatal":
		util.logFunc = util.Logger.Fatalf
	default:
		util.logFunc = util.Logger.Infof
	}
}

type hLogger struct {
	LogHttpUtil
	pool buffer.Pool
}

func NewHttpLogger(conf fig.Properties, logger xlog.Logger) *hLogger {
	ret := &hLogger{
		LogHttpUtil: *NewLogHttpUtil(conf, logger),
		pool:        buffer.NewPool(),
	}
	return ret
}

func (util *hLogger) clone(opts ...LogOpt) *hLogger {
	ret := &hLogger{}
	ret.Logger = util.Logger
	ret.LogReqHeader = util.LogReqHeader
	ret.LogReqBody = util.LogReqBody
	ret.LogRespHeader = util.LogRespHeader
	ret.LogRespBody = util.LogRespBody
	ret.Level = util.Level
	ret.pool = util.pool

	for _, opt := range opts {
		opt(ret)
	}

	ret.initLog()
	return ret
}

func (util *hLogger) LogHttp() gin.HandlerFunc {
	return util.log
}

func (util *hLogger) OptLogHttp(opts ...LogOpt) gin.HandlerFunc {
	return util.clone(opts...).log
}

func (util *hLogger) log(c *gin.Context) {
	start := time.Now()

	path := c.Request.URL.Path
	clientIP := c.ClientIP()
	method := c.Request.Method
	requestId := idUtil.RandomId(16)
	params := c.Params
	querys := c.Request.URL.RawQuery
	reqHeaderBuf := util.pool.Get()
	defer util.pool.Put(reqHeaderBuf)
	if util.LogReqHeader {
		getHeaderBuffer(reqHeaderBuf, c.Request.Header)
	}

	//c.Set(REQEUST_ID, requestId)

	reqBody := ""
	if util.LogReqBody {
		reqBodyWrapper := buffer.NewReadWriteCloser(util.pool)
		io.Copy(reqBodyWrapper, c.Request.Body)
		c.Request.Body.Close()
		reqBody = string(reqBodyWrapper.Bytes())
		c.Request.Body = reqBodyWrapper
	}

	var blw *responseBodyWriter
	if util.LogRespBody {
		blw = newResponseBodyWriter(c.Writer, buffer.NewReadWriteCloser(util.pool))
		c.Writer = blw
		defer blw.Close()
	}

	if util.LogReqBody {
		util.Logger.Infof("[Request\t%s] [path]: %s , [client ip]: %s , [method]: %s %s [params]: %v , [query]: %s [data]: %s\n",
			requestId, path, clientIP, method, reqHeaderBuf.String(), params, querys, reqBody)
	} else {
		util.Logger.Infof("[Request\t%s] [path]: %s , [client ip]: %s , [method]: %s %s [params]: %v , [query]: %s\n",
			requestId, path, clientIP, method, reqHeaderBuf.String(), params, querys)
	}

	// 处理请求
	c.Next()

	// 结束时间
	end := time.Now()
	//执行时间
	latency := end.Sub(start)

	statusCode := c.Writer.Status()

	var data string
	if util.LogRespBody {
		data = string(blw.getBody())
	}
	respHeaderBuf := util.pool.Get()
	defer util.pool.Put(respHeaderBuf)
	if util.LogRespHeader {
		rh := c.Writer.Header()
		if rh != nil {
			getHeaderBuffer(respHeaderBuf, rh.Clone())
		}
	}
	util.output("[Response\t%s] [path]: %s , [method]: %s , [latency]: %d ms, [status]: %d %s%s\n",
		requestId, path, method, latency/time.Millisecond, statusCode, respHeaderBuf.String(), data)
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *buffer.ReadWriteCloser
}

func newResponseBodyWriter(w gin.ResponseWriter, rwc *buffer.ReadWriteCloser) *responseBodyWriter {
	ret := &responseBodyWriter{
		ResponseWriter: w,
		body:           rwc,
	}
	ret.body.Write([]byte(" [data]: "))
	return ret
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseBodyWriter) WriteString(s string) (int, error) {
	w.body.Write([]byte(s))
	return w.ResponseWriter.WriteString(s)
}

func (w *responseBodyWriter) Close() error {
	return w.body.Close()
}

func (w *responseBodyWriter) getBody() []byte {
	return w.body.Bytes()
}

func getHeaderBuffer(buf *bytes.Buffer, header http.Header) {
	buf.WriteString(", [header]: ")
	if len(header) > 0 {
		for k, vs := range header {
			buf.WriteString(k)
			buf.WriteString("=")
			for i := range vs {
				buf.WriteString(vs[i])
				if i < len(vs)-1 {
					buf.WriteString(",")
				}
			}
			buf.WriteString(" ")
		}
	}
	buf.WriteString(" ,")
}
