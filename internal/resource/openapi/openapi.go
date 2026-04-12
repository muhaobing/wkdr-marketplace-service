package openapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/muhaobing/std-go/restserver/registry"

	"wdkr-marketplace-service/internal/common/constant/err_code"
	"wdkr-marketplace-service/internal/common/constant/sys_err"
	"wdkr-marketplace-service/internal/common/utils/http_utils"
	"wdkr-marketplace-service/internal/domain/companyecoin"
	"wdkr-marketplace-service/internal/domain/ecoin"
	"wdkr-marketplace-service/internal/domain/ecoinbill"
	"wdkr-marketplace-service/internal/domain/order"
	"wdkr-marketplace-service/internal/domain/payment"
	"wdkr-marketplace-service/internal/domain/payment/payment_model"
	"wdkr-marketplace-service/internal/domain/user"
)

// OpenAPIResource OpenAPI接口资源（面向内部平台及外部支付回调）
type OpenAPIResource struct {
	ecoinService        ecoin.EcoinService
	companyEcoinService companyecoin.CompanyEcoinService
	paymentService      payment.PaymentService
	orderService        order.OrderService
	userService         user.UserService
	pointsBillService   ecoinbill.EcoinBillService
}

// NewOpenAPIResource 创建OpenAPI资源实例
func NewOpenAPIResource(ecoinService ecoin.EcoinService, companyEcoinService companyecoin.CompanyEcoinService, paymentService payment.PaymentService, orderService order.OrderService, userService user.UserService, pointsBillService ecoinbill.EcoinBillService) *OpenAPIResource {
	return &OpenAPIResource{
		ecoinService:        ecoinService,
		companyEcoinService: companyEcoinService,
		paymentService:      paymentService,
		orderService:        orderService,
		userService:         userService,
		pointsBillService:   pointsBillService,
	}
}

// fetchUserIDByBizBinding 根据 biz_code + biz_user_id 查绑定得到商城 user_id；无绑定时返回 sys_err.ErrUserBindingNotFound
func (r *OpenAPIResource) fetchUserIDByBizBinding(ctx context.Context, bizCode string, bizUserId uint64) (uint64, error) {
	bizCode = strings.TrimSpace(bizCode)
	if bizCode == "" {
		return 0, errors.New("biz_code is required")
	}
	if bizUserId == 0 {
		return 0, errors.New("biz_user_id is required")
	}
	u, err := r.userService.GetUserByBiz(ctx, bizCode, bizUserId)
	if err != nil {
		return 0, err
	}
	if u == nil {
		return 0, sys_err.ErrUserBindingNotFound
	}
	return uint64(u.Id), nil
}

func writeOpenAPIError(ctx *gin.Context, err error) {
	if errors.Is(err, sys_err.ErrUserBindingNotFound) {
		http_utils.WriteResponseWithRetcode(ctx, err_code.UserBindingNotFound, err.Error())
		return
	}
	if errors.Is(err, sys_err.ErrInsufficientEcoin) {
		http_utils.WriteResponseWithRetcode(ctx, err_code.EcoinInsufficientBalance, err.Error())
		return
	}
	if errors.Is(err, sys_err.ErrEcoinBillNotFound) || errors.Is(err, sys_err.ErrEcoinBillInvalidState) {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	http_utils.WriteResponse(ctx, nil, err)
}

func (r *OpenAPIResource) userCompanyId(ctx context.Context, userId uint64) (uint64, error) {
	u, err := r.userService.GetUserById(ctx, uint(userId))
	if err != nil {
		return 0, err
	}
	if u == nil {
		return 0, errors.New("user not found")
	}
	return u.CompanyId, nil
}

// ==================== 积分接口 ====================

// AddEcoinRequest 增加积分请求（按业务身份定位商城用户）
type AddEcoinRequest struct {
	BizCode     string  `json:"biz_code" binding:"required"`    // 业务平台代码
	BizUserId   uint64  `json:"biz_user_id,string" binding:"required"` // 业务平台用户 ID
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

	userId, err := r.fetchUserIDByBizBinding(ctx.Request.Context(), req.BizCode, req.BizUserId)
	if err != nil {
		writeOpenAPIError(ctx, err)
		return
	}

	cid, err := r.userCompanyId(ctx.Request.Context(), userId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	if cid > 0 {
		tx, err := r.companyEcoinService.AddCompanyEcoin(ctx.Request.Context(), &companyecoin.AddCompanyEcoinRequest{
			CompanyId:      cid,
			OperatorUserId: userId,
			Amount:         req.Amount,
			SourceType:     req.SourceType,
			SourceId:       req.SourceId,
			Description:    req.Description,
		})
		if err != nil {
			http_utils.WriteResponse(ctx, nil, err)
			return
		}
		http_utils.WriteResponse(ctx, tx, nil)
		return
	}

	transaction, err := r.ecoinService.AddEcoin(ctx.Request.Context(), &ecoin.AddEcoinRequest{
		UserId:      userId,
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

// DeductEcoinRequest 扣除积分请求（按业务身份定位商城用户）
type DeductEcoinRequest struct {
	BizCode     string  `json:"biz_code" binding:"required"`    // 业务平台代码
	BizUserId   uint64  `json:"biz_user_id,string" binding:"required"` // 业务平台用户 ID
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

	userId, err := r.fetchUserIDByBizBinding(ctx.Request.Context(), req.BizCode, req.BizUserId)
	if err != nil {
		writeOpenAPIError(ctx, err)
		return
	}

	cid, err := r.userCompanyId(ctx.Request.Context(), userId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	if cid > 0 {
		tx, err := r.companyEcoinService.DeductCompanyEcoin(ctx.Request.Context(), &companyecoin.DeductCompanyEcoinRequest{
			CompanyId:      cid,
			OperatorUserId: userId,
			Amount:         req.Amount,
			SourceType:     req.SourceType,
			SourceId:       req.SourceId,
			Description:    req.Description,
		})
		if err != nil {
			writeOpenAPIError(ctx, err)
			return
		}
		http_utils.WriteResponse(ctx, tx, nil)
		return
	}

	transaction, err := r.ecoinService.DeductEcoin(ctx.Request.Context(), &ecoin.DeductEcoinRequest{
		UserId:      userId,
		Amount:      req.Amount,
		SourceType:  req.SourceType,
		SourceId:    req.SourceId,
		Description: req.Description,
	})
	if err != nil {
		writeOpenAPIError(ctx, err)
		return
	}

	http_utils.WriteResponse(ctx, transaction, nil)
}

// EcoinBalanceRequest 查询积分余额（按业务身份定位商城用户；JWT 仍用于鉴权）
type EcoinBalanceRequest struct {
	BizCode   string `json:"biz_code" binding:"required"`    // 业务平台代码
	BizUserId uint64 `json:"biz_user_id,string" binding:"required"` // 业务平台用户 ID
}

// PostEcoinBalance 获取用户积分信息
// POST /openapi/ecoin/balance（业务参数在 JWT payload 中）
func (r *OpenAPIResource) PostEcoinBalance(ctx *gin.Context) {
	var req EcoinBalanceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	userId, err := r.fetchUserIDByBizBinding(ctx.Request.Context(), req.BizCode, req.BizUserId)
	if err != nil {
		writeOpenAPIError(ctx, err)
		return
	}

	cid, err := r.userCompanyId(ctx.Request.Context(), userId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	if cid > 0 {
		info, err := r.companyEcoinService.GetCompanyEcoin(ctx.Request.Context(), cid)
		if err != nil {
			http_utils.WriteResponse(ctx, nil, err)
			return
		}
		http_utils.WriteResponse(ctx, info, nil)
		return
	}

	ecoinInfo, err := r.ecoinService.GetUserEcoin(ctx.Request.Context(), userId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, ecoinInfo, nil)
}

// EcoinBillCreateRequest 预扣积分（生成积分账单并扣减可用余额）
type EcoinBillCreateRequest struct {
	BizCode   string  `json:"biz_code" binding:"required"`    // 业务平台代码
	BizUserId uint64  `json:"biz_user_id,string" binding:"required"` // 业务平台用户 ID
	Cost      float64 `json:"cost" binding:"required,gt=0"`   // 预扣积分数量
}

// PostEcoinBillCreate 预扣积分并创建积分账单（JWT 鉴权）
// POST /openapi/ecoin/bill/create
// retcode：0 成功返回 bill_id；UserBindingNotFound(-100404) 未绑定；EcoinInsufficientBalance(-100402) 余额不足；15 分钟内未确认则自动取消并退款
func (r *OpenAPIResource) PostEcoinBillCreate(ctx *gin.Context) {
	var req EcoinBillCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	userId, err := r.fetchUserIDByBizBinding(ctx.Request.Context(), req.BizCode, req.BizUserId)
	if err != nil {
		writeOpenAPIError(ctx, err)
		return
	}

	cid, err := r.userCompanyId(ctx.Request.Context(), userId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	bill, err := r.pointsBillService.PreDeduct(ctx.Request.Context(), userId, cid, req.Cost)
	if err != nil {
		writeOpenAPIError(ctx, err)
		return
	}

	http_utils.WriteResponse(ctx, gin.H{
		"bill_id":            bill.BillId,
		"amount":             bill.Amount,
		"expires_in_seconds": 900,
		"status":             "incomplete",
	}, nil)
}

// EcoinBillMutateRequest 确认/取消账单（biz 身份 + bill_id）
type EcoinBillMutateRequest struct {
	BizCode   string `json:"biz_code" binding:"required"`
	BizUserId uint64 `json:"biz_user_id,string" binding:"required"`
	BillId    string `json:"bill_id" binding:"required"`
}

// PostEcoinBillConfirm 确认积分账单（扣款生效，状态 completed）
// POST /openapi/ecoin/bill/confirm
func (r *OpenAPIResource) PostEcoinBillConfirm(ctx *gin.Context) {
	var req EcoinBillMutateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	userId, err := r.fetchUserIDByBizBinding(ctx.Request.Context(), req.BizCode, req.BizUserId)
	if err != nil {
		writeOpenAPIError(ctx, err)
		return
	}
	cid, err := r.userCompanyId(ctx.Request.Context(), userId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	if err := r.pointsBillService.ConfirmBill(ctx.Request.Context(), req.BillId, userId, cid); err != nil {
		writeOpenAPIError(ctx, err)
		return
	}
	http_utils.WriteResponse(ctx, gin.H{"bill_id": req.BillId, "status": "completed"}, nil)
}

// PostEcoinBillCancel 取消积分账单并退回积分（单事务）
// POST /openapi/ecoin/bill/cancel
func (r *OpenAPIResource) PostEcoinBillCancel(ctx *gin.Context) {
	var req EcoinBillMutateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	userId, err := r.fetchUserIDByBizBinding(ctx.Request.Context(), req.BizCode, req.BizUserId)
	if err != nil {
		writeOpenAPIError(ctx, err)
		return
	}
	cid, err := r.userCompanyId(ctx.Request.Context(), userId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	if err := r.pointsBillService.CancelBill(ctx.Request.Context(), req.BillId, userId, cid); err != nil {
		writeOpenAPIError(ctx, err)
		return
	}
	http_utils.WriteResponse(ctx, gin.H{"bill_id": req.BillId, "status": "cancelled"}, nil)
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

	cid, err := r.userCompanyId(ctx.Request.Context(), req.UserId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	if cid > 0 {
		info, _, err := r.companyEcoinService.InitCompanyEcoin(ctx.Request.Context(), cid)
		if err != nil {
			http_utils.WriteResponse(ctx, nil, err)
			return
		}
		http_utils.WriteResponse(ctx, info, nil)
		return
	}

	ecoinInfo, _, err := r.ecoinService.InitUserEcoin(ctx.Request.Context(), req.UserId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, ecoinInfo, nil)
}

// EcoinTransactionsRequest 积分流水列表（按业务身份定位商城用户；JWT payload）
type EcoinTransactionsRequest struct {
	BizCode   string `json:"biz_code" binding:"required"`    // 业务平台代码
	BizUserId uint64 `json:"biz_user_id,string" binding:"required"` // 业务平台用户 ID
	Offset    int    `json:"offset"`                         // 偏移量
	Limit     int    `json:"limit"`                          // 每页数量
}

// PostEcoinTransactions 获取积分流水列表
// POST /openapi/ecoin/transactions（业务参数在 JWT payload 中）
func (r *OpenAPIResource) PostEcoinTransactions(ctx *gin.Context) {
	var req EcoinTransactionsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	userId, err := r.fetchUserIDByBizBinding(ctx.Request.Context(), req.BizCode, req.BizUserId)
	if err != nil {
		writeOpenAPIError(ctx, err)
		return
	}

	cid, err := r.userCompanyId(ctx.Request.Context(), userId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	if cid > 0 {
		list, err := r.companyEcoinService.GetTransactionList(ctx.Request.Context(), &companyecoin.TransactionListRequest{
			CompanyId: cid,
			Offset:    req.Offset,
			Limit:     req.Limit,
		})
		if err != nil {
			http_utils.WriteResponse(ctx, nil, err)
			return
		}
		http_utils.WriteResponse(ctx, list, nil)
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

const maxPayCallbackLogBody = 8192

// logPaymentCallback 使用 fmt 打到 stdout，与项目内 [WARN] 等一致，便于 shell 重定向到 backend.log（标准库 log 默认走 stderr，易与文件日志不一致）
func logPaymentCallback(tag string, body []byte) {
	n := len(body)
	s := string(body)
	if n > maxPayCallbackLogBody {
		s = string(body[:maxPayCallbackLogBody]) + fmt.Sprintf("...(truncated, total_bytes=%d)", n)
	}
	fmt.Printf("[%s] bytes=%d body=%s\n", tag, n, s)
}

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

	logPaymentCallback("WechatPayNotify", body)
	result, err := r.paymentService.HandleNotify(ctx.Request.Context(), "wechat", body)
	if err != nil {
		fmt.Printf("[WechatPayNotify] HandleNotify error: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FAIL",
			"message": err.Error(),
		})
		return
	}
	if result != nil {
		fmt.Printf("[WechatPayNotify] HandleNotify ok biz_order_no=%s payment_status=%d pay_time=%d\n",
			result.BizOrderNo, result.Status, result.PayTime)
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

	logPaymentCallback("WechatRefundNotify", body)
	_, err = r.paymentService.HandleNotify(ctx.Request.Context(), "wechat_refund", body)
	if err != nil {
		fmt.Printf("[WechatRefundNotify] HandleNotify error: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FAIL",
			"message": err.Error(),
		})
		return
	}
	fmt.Printf("[WechatRefundNotify] HandleNotify ok\n")

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

	logPaymentCallback("AlipayPayNotify", body)
	result, err := r.paymentService.HandleNotify(ctx.Request.Context(), "alipay", body)
	if err != nil {
		fmt.Printf("[AlipayPayNotify] HandleNotify error: %v\n", err)
		ctx.String(http.StatusInternalServerError, "fail")
		return
	}
	if result != nil {
		fmt.Printf("[AlipayPayNotify] HandleNotify ok biz_order_no=%s payment_status=%d pay_time=%d\n",
			result.BizOrderNo, result.Status, result.PayTime)
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
			group.POST("/ecoin/bill/create", r.PostEcoinBillCreate)
			group.POST("/ecoin/bill/confirm", r.PostEcoinBillConfirm)
			group.POST("/ecoin/bill/cancel", r.PostEcoinBillCancel)
			group.POST("/ecoin/transactions", r.PostEcoinTransactions)

			// 支付回调接口
			group.POST("/callback/wechat/pay", r.WechatPayNotify)
			group.POST("/callback/wechat/refund", r.WechatRefundNotify)
			group.POST("/callback/alipay/pay", r.AlipayPayNotify)
		}
	}
}
