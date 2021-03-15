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
	"net/http"
	"reflect"
	"strings"
	"time"
)

const (
	REQEUST_ID = "_REQEUST_ID"
)

type ResponseBodyLogWriter struct {
	gin.ResponseWriter
	body strings.Builder
}

func (w *ResponseBodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
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

type LogConfig struct {
	RequestHeader  bool
	RequestBody    bool
	ResponseHeader bool
	ResponseBody   bool
}

type LogHttpUtil struct {
	Logger xlog.Logger
	Conf   LogConfig
}

func NewLogHttpUtil(conf fig.Properties, logger xlog.Logger) *LogHttpUtil {
	ret := &LogHttpUtil{
		Logger: logger,
	}
	conf.GetValue("Log", &ret.Conf)
	return ret
}

func (util *LogHttpUtil) LogHttp() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		method := c.Request.Method
		requestId := idUtil.RandomId(16)
		params := c.Params
		querys := c.Request.URL.RawQuery
		reqHeader := ""
		if util.Conf.RequestHeader {
			reqHeader = getHeaderStr(c.Request.Header)
		}

		c.Set(REQEUST_ID, requestId)

		util.Logger.Infof("[Request %s] [path]: %s , [client ip]: %s , [method]: %s , %s , [params]: %v , [query]: %s \n",
			requestId, path, clientIP, method, reqHeader, params, querys)

		var blw *ResponseBodyLogWriter
		if util.Conf.ResponseBody {
			blw = &ResponseBodyLogWriter{ResponseWriter: c.Writer}
			c.Writer = blw
		}

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		//执行时间
		latency := end.Sub(start)

		statusCode := c.Writer.Status()

		var data string
		if util.Conf.ResponseBody {
			data = blw.body.String()
		}
		respHeader := ""
		if util.Conf.ResponseHeader {
			rh := c.Writer.Header()
			if rh != nil {
				respHeader = getHeaderStr(rh.Clone())
			}
		}
		respId, _ := c.Get(REQEUST_ID)
		util.Logger.Infof("[Response %s] [latency]: %d ms, [status]: %d , %s , [data]: %s\n",
			respId.(string), latency/time.Millisecond, statusCode, respHeader, data)
	}
}

func getHeaderStr(header http.Header) string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("[header]: ")
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
		}
	}
	return buf.String()
}
