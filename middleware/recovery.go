package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

const (
	RecoveryHandlerKey = "recovery-handler"
)

type RecoveryHandler struct{}

func (c *RecoveryHandler) Name() string {
	return RecoveryHandlerKey
}

func (c *RecoveryHandler) Handle(ctx *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			// 打印堆栈信息
			stack := string(debug.Stack())
			fmt.Printf("[Recovery] panic recovered:\n%v\n%s\n", err, stack)

			// 返回 500 错误
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"retcode": -1,
				"message": "Internal Server Error",
			})
		}
	}()

	ctx.Next()
}
