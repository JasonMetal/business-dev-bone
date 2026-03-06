package core

import (
	"business-dev-bone/pkg/component-base/core"
	"github.com/gin-gonic/gin"
)

// installRoutes 只注册基础路由，无业务
func installRoutes(e *gin.Engine) {
	e.NoRoute(func(c *gin.Context) {
		core.WriteResponse(c, nil, gin.H{"message": "not found"})
	})
	e.GET("/ping", func(c *gin.Context) {
		core.WriteResponse(c, nil, gin.H{"message": "pong"})
	})
}
