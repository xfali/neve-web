// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description:

package midware

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/xfali/fig"
	"github.com/xfali/goutils/idUtil"
	"github.com/xfali/xlog"
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

type LogHttpUtil struct {
	Logger      xlog.Logger
	LogRespBody bool
}

func NewLogHttpUtil(conf fig.Properties, logger xlog.Logger) *LogHttpUtil {
	return &LogHttpUtil{
		Logger:      logger,
		LogRespBody: conf.Get("Log.ResponseBody", "false") == "true",
	}
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

		c.Set(REQEUST_ID, requestId)

		util.Logger.Infof("[Request %s] [path]: %s, [client ip]: %s, [method]: %s, [params]: %v, [query]: %s\n",
			requestId, path, clientIP, method, params, querys)

		var blw *ResponseBodyLogWriter
		if util.LogRespBody {
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
		if util.LogRespBody {
			data = blw.body.String()
		}
		respId, _ := c.Get(REQEUST_ID)
		util.Logger.Infof("[Response %s] [latency]: %d ms, [status]: %d, [data]: %s\n",
			respId.(string), latency/time.Millisecond, statusCode, data)
	}
}
