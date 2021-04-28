# neve-web

neve-web是neve的WEB扩展组件，用于集成WEB相关服务。

内置WEB中间件为[gin](https://github.com/gin-gonic/gin)

## 安装
```
go get github.com/xfali/neve-web
```

## 使用
  
### 1. neve集成（依赖[neve-core](https://github.com/xfali/neve-core)）
```
app := neve.NewFileConfigApplication("assets/config-test.yaml")
app.RegisterBean(gineve.NewProcessor())
//注册值注入处理器，用于根据配置注入值（非必须）
app.RegisterBean(processor.NewValueProcessor())
//注册其他对象
app.RegisterBean(&testProcess{})
app.RegisterBean(&webBean{})
app.Run()
```

### 2. 配置
在config-example.yaml中配置示例如下：
```
neve:
  web:
    log:
      requestHeader: true
      requestBody: true
      responseHeader: true
      responseBody: true
      level: "warn"

    server:
      contextPath: ""
      port: 8080
      readTimeout: 15
      writeTimeout: 15
      idleTimeout: 15
```
* 【neve.web.log】配置rest的日志输出，包含request header、body，response header、body以及配置日志级别，根据项目需要进行配置。
* 【neve.web.server】配置WEB服务的端口、读写超时等配置，contextPath配置总的根路由路径，如contextPath: "/order"

### 3. 注册路由
注册的bean实现 HttpRoutes(engine gin.IRouter)方法
```
//webBean通过app.RegisterBean(&webBean{})注册，并实现下列方法：

func (b *webBean) HttpRoutes(engine gin.IRouter) {
	engine.GET("test", b.HttpLogger.LogHttp(), func(context *gin.Context) {
		context.JSON(http.StatusOK, result.Ok(b.V))
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
```

### 4. 输出日志配置
注入loghttp.HttpLogger，在gin.IRouter中添加该handler
```
type webBean struct {
	V          string //`fig:"Log.Level"`
	//注入
	HttpLogger loghttp.HttpLogger `inject:""`
}
func (b *webBean) HttpRoutes(engine gin.IRouter) {
    //使用“b.HttpLogger.LogHttp()”配置，作为首个handler
	engine.GET("test", b.HttpLogger.LogHttp(), func(context *gin.Context) {
		context.JSON(http.StatusOK, result.Ok(b.V))
	})
}
```
