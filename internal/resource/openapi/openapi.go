package openapi

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/muhaobing-eng/std-go/restserver/registry"

	"wdkr-marketplace-service/internal/common/utils/http_utils"
	"wdkr-marketplace-service/internal/domain/ecoin"
	"wdkr-marketplace-service/internal/domain/order"
	"wdkr-marketplace-service/internal/domain/payment"
	"wdkr-marketplace-service/internal/domain/payment/payment_model"
	"wdkr-marketplace-service/internal/domain/user"
)

// OpenAPIResource OpenAPI接口资源（面向内部平台及外部支付回调）
type OpenAPIResource struct {
	ecoinService   ecoin.EcoinService
	paymentService payment.PaymentService
	userService    user.UserService
	orderService   order.OrderService
}

// NewOpenAPIResource 创建OpenAPI资源实例
func NewOpenAPIResource(ecoinService ecoin.EcoinService, paymentService payment.PaymentService, userService user.UserService, orderService order.OrderService) *OpenAPIResource {
	return &OpenAPIResource{
		ecoinService:   ecoinService,
		paymentService: paymentService,
		userService:    userService,
		orderService:   orderService,
	}
}

// ==================== 用户接口 ====================

// BindUserRequest 用户绑定请求
type BindUserRequest struct {
	BizCode   string `json:"biz_code" binding:"required"`    // 业务平台代码
	BizUserId uint64 `json:"biz_user_id" binding:"required"` // 业务平台用户ID
	TelNo     string `json:"tel_no"`                         // 手机号
	Email     string `json:"email"`                          // 邮箱
	Secret    string `json:"secret" binding:"required"`      // 用户密钥
}

// BindUser 绑定用户
// POST /openapi/user/bind
func (r *OpenAPIResource) BindUser(ctx *gin.Context) {
	var req BindUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	resp, err := r.userService.BindUser(ctx.Request.Context(), &user.BindUserRequest{
		BizCode:   req.BizCode,
		BizUserId: req.BizUserId,
		TelNo:     req.TelNo,
		Email:     req.Email,
		Secret:    req.Secret,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, resp, nil)
}

// UnbindUserRequest 用户解绑请求
type UnbindUserRequest struct {
	UserId  uint   `json:"user_id" binding:"required"`  // 商城用户ID
	BizCode string `json:"biz_code" binding:"required"` // 业务平台代码
}

// UnbindUser 解绑用户
// POST /openapi/user/unbind
func (r *OpenAPIResource) UnbindUser(ctx *gin.Context) {
	var req UnbindUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	err := r.userService.UnbindUser(ctx.Request.Context(), &user.UnbindUserRequest{
		UserId:  req.UserId,
		BizCode: req.BizCode,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, nil, nil)
}

// BizLoginRequest 业务平台登录请求
type BizLoginRequest struct {
	BizCode   string `json:"biz_code" binding:"required"`    // 业务平台代码
	BizUserId uint64 `json:"biz_user_id" binding:"required"` // 业务平台用户ID
	Secret    string `json:"secret" binding:"required"`      // 用户密钥
}

// BizLogin 业务平台登录
// POST /openapi/user/login
func (r *OpenAPIResource) BizLogin(ctx *gin.Context) {
	var req BizLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	resp, err := r.userService.BizLogin(ctx.Request.Context(), &user.BizLoginRequest{
		BizCode:   req.BizCode,
		BizUserId: req.BizUserId,
		Secret:    req.Secret,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, resp, nil)
}

// GetUserBindings 获取用户绑定信息
// GET /openapi/user/:user_id/bindings
func (r *OpenAPIResource) GetUserBindings(ctx *gin.Context) {
	userIdStr := ctx.Param("user_id")
	userId, err := strconv.ParseUint(userIdStr, 10, 32)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	bindings, err := r.userService.GetBindingsByUserId(ctx.Request.Context(), uint(userId))
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, bindings, nil)
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

// GetUserEcoin 获取用户积分信息
// GET /openapi/ecoin/:user_id
func (r *OpenAPIResource) GetUserEcoin(ctx *gin.Context) {
	userIdStr := ctx.Param("user_id")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	ecoinInfo, err := r.ecoinService.GetUserEcoin(ctx.Request.Context(), userId)
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

// GetEcoinTransactions 获取积分流水列表
// GET /openapi/ecoin/:user_id/transactions
func (r *OpenAPIResource) GetEcoinTransactions(ctx *gin.Context) {
	userIdStr := ctx.Param("user_id")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	type ListRequest struct {
		Offset int `form:"offset"` // 偏移量
		Limit  int `form:"limit"`  // 每页数量
	}

	var req ListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	transactions, err := r.ecoinService.GetEcoinTransactionList(ctx.Request.Context(), &ecoin.EcoinTransactionListRequest{
		UserId: userId,
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
			// 用户接口
			group.POST("/user/bind", r.BindUser)
			group.POST("/user/unbind", r.UnbindUser)
			group.POST("/user/login", r.BizLogin)
			group.GET("/user/:user_id/bindings", r.GetUserBindings)

			// 积分接口
			group.POST("/ecoin/add", r.AddEcoin)
			group.POST("/ecoin/deduct", r.DeductEcoin)
			group.POST("/ecoin/init", r.InitUserEcoin)
			group.GET("/ecoin/:user_id", r.GetUserEcoin)
			group.GET("/ecoin/:user_id/transactions", r.GetEcoinTransactions)

			// 支付回调接口
			group.POST("/callback/wechat/pay", r.WechatPayNotify)
			group.POST("/callback/wechat/refund", r.WechatRefundNotify)
			group.POST("/callback/alipay/pay", r.AlipayPayNotify)
		}
	}
}
