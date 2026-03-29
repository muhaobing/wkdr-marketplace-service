package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const CorsHandlerKey = "cors-handler"

// CorsHandler 处理浏览器跨域与 OPTIONS 预检（须排在鉴权等中间件之前）
type CorsHandler struct{}

func (c *CorsHandler) Name() string {
	return CorsHandlerKey
}

func (c *CorsHandler) Handle(ctx *gin.Context) {
	origin := ctx.GetHeader("Origin")
	if origin != "" && corsOriginAllowed(origin) {
		ctx.Header("Access-Control-Allow-Origin", origin)
		ctx.Header("Access-Control-Allow-Credentials", "true")
	}
	ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	ctx.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
	ctx.Header("Access-Control-Max-Age", "86400")

	if ctx.Request.Method == http.MethodOptions {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}
	ctx.Next()
}

func corsOriginAllowed(origin string) bool {
	// 与 Nginx server_name 中 lawmind 域名一致；含 www / 非 www、http / https
	for _, o := range []string{
		"http://www.lawmind.top",
		"http://lawmind.top",
		"https://www.lawmind.top",
		"https://lawmind.top",
	} {
		if origin == o {
			return true
		}
	}
	// 本地开发：Vite 等
	if strings.HasPrefix(origin, "http://127.0.0.1:") || strings.HasPrefix(origin, "http://localhost:") {
		return true
	}
	return false
}
