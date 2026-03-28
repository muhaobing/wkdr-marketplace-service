package openapi

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhaobing/std-go/restserver/registry"

	"wdkr-marketplace-service/internal/common/utils/http_utils"
	"wdkr-marketplace-service/internal/domain/ecoin"
	"wdkr-marketplace-service/internal/domain/order"
	"wdkr-marketplace-service/internal/domain/payment"
	"wdkr-marketplace-service/internal/domain/payment/payment_model"
)

// OpenAPIResource OpenAPI接口资源（面向内部平台及外部支付回调）
type OpenAPIResource struct {
	ecoinService   ecoin.EcoinService
	paymentService payment.PaymentService
	orderService   order.OrderService
}

// NewOpenAPIResource 创建OpenAPI资源实例
func NewOpenAPIResource(ecoinService ecoin.EcoinService, paymentService payment.PaymentService, orderService order.OrderService) *OpenAPIResource {
	return &OpenAPIResource{
		ecoinService:   ecoinService,
		paymentService: paymentService,
		orderService:   orderService,
	}
}

// ==================== 积分接口 ====================

// AddEcoinRequest 增加积分请求
type AddEcoinRequest struct {
	UserId      uint64  `json:"user_id" binding:"required"`     // 用户ID
	Amount      float64 `json:"amount" binding:"required,gt=0"` // 积分数量（必须大于0）
	SourceType  string  `json:"source_type" binding:"required"` // 来源类型
	SourceId    string  `json:"source_id"`                      // 来源业务ID
	Description string  `json:"description"`                    // 描述
}

// AddEcoin 增加积分
// POST /openapi/ecoin/add
func (r *OpenAPIResource) AddEcoin(ctx *gin.Context) {
	var req AddEcoinRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	transaction, err := r.ecoinService.AddEcoin(ctx.Request.Context(), &ecoin.AddEcoinRequest{
		UserId:      req.UserId,
		Amount:      req.Amount,
		SourceType:  req.SourceType,
		SourceId:    req.SourceId,
		Description: req.Description,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, transaction, nil)
}

// DeductEcoinRequest 扣除积分请求
type DeductEcoinRequest struct {
	UserId      uint64  `json:"user_id" binding:"required"`     // 用户ID
	Amount      float64 `json:"amount" binding:"required,gt=0"` // 积分数量（必须大于0）
	SourceType  string  `json:"source_type" binding:"required"` // 来源类型
	SourceId    string  `json:"source_id"`                      // 来源业务ID
	Description string  `json:"description"`                    // 描述
}

// DeductEcoin 扣除积分
// POST /openapi/ecoin/deduct
func (r *OpenAPIResource) DeductEcoin(ctx *gin.Context) {
	var req DeductEcoinRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	transaction, err := r.ecoinService.DeductEcoin(ctx.Request.Context(), &ecoin.DeductEcoinRequest{
		UserId:      req.UserId,
		Amount:      req.Amount,
		SourceType:  req.SourceType,
		SourceId:    req.SourceId,
		Description: req.Description,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, transaction, nil)
}

// EcoinBalanceRequest 查询积分余额（JWT payload）
type EcoinBalanceRequest struct {
	UserId uint64 `json:"user_id" binding:"required"` // 用户ID
}

// PostEcoinBalance 获取用户积分信息
// POST /openapi/ecoin/balance（业务参数在 JWT payload 中）
func (r *OpenAPIResource) PostEcoinBalance(ctx *gin.Context) {
	var req EcoinBalanceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	ecoinInfo, err := r.ecoinService.GetUserEcoin(ctx.Request.Context(), req.UserId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, ecoinInfo, nil)
}

// InitUserEcoin 初始化用户积分账户
// POST /openapi/ecoin/init
func (r *OpenAPIResource) InitUserEcoin(ctx *gin.Context) {
	type InitRequest struct {
		UserId uint64 `json:"user_id" binding:"required"` // 用户ID
	}

	var req InitRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	ecoinInfo, err := r.ecoinService.InitUserEcoin(ctx.Request.Context(), req.UserId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, ecoinInfo, nil)
}

// EcoinTransactionsRequest 积分流水列表（JWT payload）
type EcoinTransactionsRequest struct {
	UserId uint64 `json:"user_id" binding:"required"` // 用户ID
	Offset int    `json:"offset"`                     // 偏移量
	Limit  int    `json:"limit"`                      // 每页数量
}

// PostEcoinTransactions 获取积分流水列表
// POST /openapi/ecoin/transactions（业务参数在 JWT payload 中）
func (r *OpenAPIResource) PostEcoinTransactions(ctx *gin.Context) {
	var req EcoinTransactionsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	transactions, err := r.ecoinService.GetEcoinTransactionList(ctx.Request.Context(), &ecoin.EcoinTransactionListRequest{
		UserId: req.UserId,
		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, transactions, nil)
}

// ==================== 支付回调接口 ====================

// WechatPayNotify 微信支付回调
// POST /openapi/callback/wechat/pay
func (r *OpenAPIResource) WechatPayNotify(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    "FAIL",
			"message": "read body failed",
		})
		return
	}

	log.Println("WechatPayNotify:", string(body))
	result, err := r.paymentService.HandleNotify(ctx.Request.Context(), "wechat", body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FAIL",
			"message": err.Error(),
		})
		return
	}

	// 支付成功 → 联动更新业务订单并触发履约
	if result != nil && result.Status == payment_model.PaymentStatusPaid {
		if err := r.orderService.HandlePaymentSuccess(ctx.Request.Context(), result.BizOrderNo, result.PayTime); err != nil {
			fmt.Printf("[WARN] handle payment success failed for order %s: %v\n", result.BizOrderNo, err)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "成功",
	})
}

// WechatRefundNotify 微信退款回调
// POST /openapi/callback/wechat/refund
func (r *OpenAPIResource) WechatRefundNotify(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    "FAIL",
			"message": "read body failed",
		})
		return
	}

	_, err = r.paymentService.HandleNotify(ctx.Request.Context(), "wechat_refund", body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FAIL",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "成功",
	})
}

// AlipayPayNotify 支付宝支付回调（预留）
// POST /openapi/callback/alipay/pay
func (r *OpenAPIResource) AlipayPayNotify(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.String(http.StatusBadRequest, "fail")
		return
	}

	result, err := r.paymentService.HandleNotify(ctx.Request.Context(), "alipay", body)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "fail")
		return
	}

	if result != nil && result.Status == payment_model.PaymentStatusPaid {
		if err := r.orderService.HandlePaymentSuccess(ctx.Request.Context(), result.BizOrderNo, result.PayTime); err != nil {
			fmt.Printf("[WARN] handle payment success failed for order %s: %v\n", result.BizOrderNo, err)
		}
	}

	ctx.String(http.StatusOK, "success")
}

// Router 注册路由
func (r *OpenAPIResource) Router() registry.Registry {
	return func(router *gin.Engine) {
		group := router.Group("/openapi")
		{
			// 积分接口
			group.POST("/ecoin/add", r.AddEcoin)
			group.POST("/ecoin/deduct", r.DeductEcoin)
			group.POST("/ecoin/init", r.InitUserEcoin)
			group.POST("/ecoin/balance", r.PostEcoinBalance)
			group.POST("/ecoin/transactions", r.PostEcoinTransactions)

			// 支付回调接口
			group.POST("/callback/wechat/pay", r.WechatPayNotify)
			group.POST("/callback/wechat/refund", r.WechatRefundNotify)
			group.POST("/callback/alipay/pay", r.AlipayPayNotify)
		}
	}
}
