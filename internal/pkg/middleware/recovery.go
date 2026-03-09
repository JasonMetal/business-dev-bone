package middleware

import (
	"business-dev-bone/pkg/component-base/log"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return gin.RecoveryWithWriter(log.GetWriter())
}
