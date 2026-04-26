package mock

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/muhaobing/std-go/restserver/registry"

	"wdkr-marketplace-service/internal/common/config"
	"wdkr-marketplace-service/internal/common/utils/callback_jwt"
)

type MockResource struct{}

func NewMockResource() *MockResource {
	return &MockResource{}
}

type deliveryCallbackRequest struct {
	SkuCode   string `json:"sku_code"`
	BizUserId string `json:"biz_user_id"`
}

type deliveryJWTWrapper struct {
	JWT string `json:"jwt"`
}

// DeliveryCallback 模拟履约回调接口
func (r *MockResource) DeliveryCallback(ctx *gin.Context) {
	var wrapper deliveryJWTWrapper
	if err := ctx.ShouldBindJSON(&wrapper); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"retcode": 1, "message": "invalid request: " + err.Error()})
		return
	}
	if strings.TrimSpace(wrapper.JWT) == "" {
		ctx.JSON(http.StatusOK, gin.H{"retcode": 1, "message": "invalid request: jwt is required"})
		return
	}

	conf := config.GetConf()
	if conf == nil {
		ctx.JSON(http.StatusOK, gin.H{"retcode": 1, "message": "config not initialized"})
		return
	}

	_, payload, err := callback_jwt.ParseToken(
		wrapper.JWT,
		conf.CallbackJWT.EffectiveExpirationSeconds(),
		func(account string) (string, bool) {
			return conf.CallbackJWT.SecretForBizCode(account)
		},
	)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"retcode": 1, "message": "invalid jwt: " + err.Error()})
		return
	}

	var req deliveryCallbackRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"retcode": 1, "message": "invalid decrypted payload: " + err.Error()})
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
