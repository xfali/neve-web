// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description:

package midware

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/xfali/fig"
	"github.com/xfali/goutils/idUtil"
	"github.com/xfali/xlog"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"
)

const (
	REQEUST_ID = "_REQEUST_ID"
)

type logFunc func(fmt string, args ...interface{})

type HttpLogger interface {
	LogHttp() gin.HandlerFunc
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
	case "panic":
		util.logFunc = util.Logger.Panicf
	case "fatal":
		util.logFunc = util.Logger.Fatalf
	default:
		util.logFunc = util.Logger.Infof
	}
}
