// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package gineve

import "github.com/gin-gonic/gin"

type Component interface {
	HttpRoutes(engine gin.IRouter)
}

type Filter interface {
	FilterHandler(ctx *gin.Context)
}
