package mock

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhaobing/std-go/restserver/registry"
)

type MockResource struct{}

func NewMockResource() *MockResource {
	return &MockResource{}
}

type deliveryCallbackRequest struct {
	SkuCode   string `json:"sku_code"`
	BizUserId string `json:"biz_user_id"`
}

// DeliveryCallback 模拟履约回调接口
func (r *MockResource) DeliveryCallback(ctx *gin.Context) {
	var req deliveryCallbackRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"retcode": 1, "message": "invalid request: " + err.Error()})
		return
	}

	log.Printf("[MockDelivery] sku_code=%s, biz_user_id=%s", req.SkuCode, req.BizUserId)

	ctx.JSON(http.StatusOK, gin.H{
		"retcode": 0,
		"message": "mock fulfill success",
	})
}

func (r *MockResource) Router() registry.Registry {
	return func(router *gin.Engine) {
		group := router.Group("/mock")
		group.POST("/delivery_callback", r.DeliveryCallback)
	}
}
