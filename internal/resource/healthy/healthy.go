package healthy

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhaobing-eng/std-go/restserver/registry"
)

type HealthyResource struct{}

func NewHealthyResource() *HealthyResource {
	return &HealthyResource{}
}

func (r *HealthyResource) Ping(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func (r *HealthyResource) Router() registry.Registry {
	return func(router *gin.Engine) {
		router.GET("/ping", r.Ping)
	}
}
