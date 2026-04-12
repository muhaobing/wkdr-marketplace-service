package marketplace

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/muhaobing/std-go/restserver/registry"

	"wdkr-marketplace-service/internal/common/config"
	"wdkr-marketplace-service/internal/common/constant/err_code"
	"wdkr-marketplace-service/internal/common/constant/sys_err"
	"wdkr-marketplace-service/internal/common/utils/auth_utils"
	"wdkr-marketplace-service/internal/common/utils/http_utils"
	bizcoderepo "wdkr-marketplace-service/internal/domain/bizcode/repo"
	"wdkr-marketplace-service/internal/domain/cart"
	companyrepo "wdkr-marketplace-service/internal/domain/company/repo"
	"wdkr-marketplace-service/internal/domain/companyecoin"
	companyecoin_model "wdkr-marketplace-service/internal/domain/companyecoin/companyecoin_model"
	"wdkr-marketplace-service/internal/domain/ecoin"
	"wdkr-marketplace-service/internal/domain/order"
	"wdkr-marketplace-service/internal/domain/sku"
	skumodel "wdkr-marketplace-service/internal/domain/sku/sku_model"
	"wdkr-marketplace-service/internal/domain/user"
)

// MarketplaceResource 商城接口资源（面向用户）
type MarketplaceResource struct {
	skuService          sku.SkuService
	orderService        order.OrderService
	ecoinService        ecoin.EcoinService
	companyEcoinService companyecoin.CompanyEcoinService
	userService         user.UserService
	cartService         cart.CartService
	bizCodeRepo         bizcoderepo.BizCodeRepo
	companyRepo         companyrepo.CompanyRepo
}

// NewMarketplaceResource 创建商城资源实例
func NewMarketplaceResource(
	skuService sku.SkuService,
	orderService order.OrderService,
	ecoinService ecoin.EcoinService,
	companyEcoinService companyecoin.CompanyEcoinService,
	userService user.UserService,
	cartService cart.CartService,
	bizCodeRepo bizcoderepo.BizCodeRepo,
	companyRepo companyrepo.CompanyRepo,
) *MarketplaceResource {
	return &MarketplaceResource{
		skuService:          skuService,
		orderService:        orderService,
		ecoinService:        ecoinService,
		companyEcoinService: companyEcoinService,
		userService:         userService,
		cartService:         cartService,
		bizCodeRepo:         bizCodeRepo,
		companyRepo:         companyRepo,
	}
}

func writeMarketplaceErr(ctx *gin.Context, err error) {
	if errors.Is(err, sys_err.ErrInsufficientEcoin) {
		http_utils.WriteResponseWithRetcode(ctx, err_code.EcoinInsufficientBalance, err.Error())
		return
	}
	http_utils.WriteResponse(ctx, nil, err)
}

// ==================== 商品接口 ====================

// ListSkusRequest 商品列表请求
type ListSkusRequest struct {
	SkuName string `form:"sku_name"` // 商品名称（模糊查询，可选）
	Offset  int    `form:"offset"`   // 偏移量
	Limit   int    `form:"limit"`    // 每页数量
}

// ListSkus 获取商品列表（仅上架商品，支持商品名模糊查询）
// GET /marketplace/skus
// sku_scope 筛选由服务端根据 session 中用户 company_id（0=个人访客，>0=企业访客）决定，不接受前端传参。
func (r *MarketplaceResource) ListSkus(ctx *gin.Context) {
	var req ListSkusRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	u, err := auth_utils.UserFromGinContext(ctx)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	visitorEnterprise := u.CompanyId > 0

	// 只查询已上架的商品
	onlineStatus := skumodel.SkuStatusOnline

	resp, err := r.skuService.ListSkus(ctx.Request.Context(), &sku.ListSkuRequest{
		SkuName:                        req.SkuName,
		Status:                         &onlineStatus,
		MarketplaceSkuScopeFilter:      true,
		MarketplaceVisitorIsEnterprise: visitorEnterprise,
		Offset:                         req.Offset,
		Limit:                          req.Limit,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, resp, nil)
}

// GetSkuDetail 获取商品详情
// GET /marketplace/skus/:id
// 是否可见由服务端根据 session 中 company_id 与 sku_scope 判定，不接受 sku_scope 查询参数。
func (r *MarketplaceResource) GetSkuDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	skuInfo, err := r.skuService.GetSkuById(ctx.Request.Context(), id)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	// 只返回已上架的商品
	if !skuInfo.IsOnline() {
		http_utils.WriteResponse(ctx, nil, nil)
		return
	}

	u, err := auth_utils.UserFromGinContext(ctx)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	visitorEnterprise := u.CompanyId > 0
	if !skumodel.SkuVisibleToMarketplaceVisitor(skuInfo.SkuScope, visitorEnterprise) {
		http_utils.WriteResponse(ctx, nil, nil)
		return
	}

	http_utils.WriteResponse(ctx, skuInfo, nil)
}

// ==================== 订单接口 ====================

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	UserId   uint64                `json:"user_id" binding:"required"`   // 用户ID
	SkuItems []*order.SkuOrderItem `json:"sku_items" binding:"required"` // SKU列表
	PayType  string                `json:"pay_type"`                     // 支付类型：ecoin/money，不传默认money
	Remark   string                `json:"remark"`                       // 备注
}

// CreateOrder 创建订单
// POST /marketplace/checkout
func (r *MarketplaceResource) CreateOrder(ctx *gin.Context) {
	var req CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	payType := req.PayType
	if payType == "" {
		payType = "money"
	}

	resp, err := r.orderService.CreateOrder(ctx.Request.Context(), &order.CreateOrderRequest{
		UserId:   req.UserId,
		SkuItems: req.SkuItems,
		PayType:  payType,
		Remark:   req.Remark,
	})
	if err != nil {
		writeMarketplaceErr(ctx, err)
		return
	}

	http_utils.WriteResponse(ctx, resp, nil)
}

// GetOrderDetail 获取订单详情
// GET /marketplace/orders/:order_no
func (r *MarketplaceResource) GetOrderDetail(ctx *gin.Context) {
	orderNo := ctx.Param("order_no")
	if orderNo == "" {
		http_utils.WriteResponse(ctx, nil, nil)
		return
	}

	orderInfo, err := r.orderService.GetOrderByOrderNo(ctx.Request.Context(), orderNo)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, orderInfo, nil)
}

// ListOrdersRequest 订单列表请求
type ListOrdersRequest struct {
	UserId uint64 `form:"user_id" binding:"required"` // 用户ID
	Status *uint8 `form:"status"`                     // 订单状态过滤（可选）
	Offset int    `form:"offset"`                     // 偏移量
	Limit  int    `form:"limit"`                      // 每页数量
}

// ListOrders 获取订单列表
// GET /marketplace/orders
func (r *MarketplaceResource) ListOrders(ctx *gin.Context) {
	var req ListOrdersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	resp, err := r.orderService.ListOrders(ctx.Request.Context(), &order.ListOrdersRequest{
		UserId: req.UserId,
		Status: req.Status,
		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, resp, nil)
}

// CancelOrderRequest 取消订单请求
type CancelOrderRequest struct {
	Reason string `json:"reason"` // 取消原因
}

// CancelOrder 取消订单
// POST /marketplace/orders/:order_no/cancel
func (r *MarketplaceResource) CancelOrder(ctx *gin.Context) {
	orderNo := ctx.Param("order_no")
	if orderNo == "" {
		http_utils.WriteResponse(ctx, nil, nil)
		return
	}

	var req CancelOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	err := r.orderService.CancelOrder(ctx.Request.Context(), &order.CancelOrderRequest{
		OrderNo: orderNo,
		Reason:  req.Reason,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, nil, nil)
}

// PayOrderRequest 支付订单请求
type PayOrderRequest struct {
	Channel   string `json:"channel" binding:"required"` // 支付渠道：ecoin/wechat/alipay
	PayMethod string `json:"pay_method"`                 // 支付方式：native/jsapi/h5（积分支付无需提供）
	ClientIP  string `json:"client_ip"`                  // 客户端IP（H5支付需要）
	OpenId    string `json:"open_id"`                    // 用户OpenID（JSAPI支付需要）
}

// PayOrder 支付订单
// POST /marketplace/orders/:order_no/pay
func (r *MarketplaceResource) PayOrder(ctx *gin.Context) {
	orderNo := ctx.Param("order_no")
	if orderNo == "" {
		http_utils.WriteResponse(ctx, nil, nil)
		return
	}

	var req PayOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	// 如果未提供 ClientIP，尝试从请求中获取
	clientIP := req.ClientIP
	if clientIP == "" {
		clientIP = ctx.ClientIP()
	}

	resp, err := r.orderService.PayOrder(ctx.Request.Context(), &order.PayOrderRequest{
		OrderNo:   orderNo,
		Channel:   req.Channel,
		PayMethod: req.PayMethod,
		ClientIP:  clientIP,
		OpenId:    req.OpenId,
	})
	if err != nil {
		writeMarketplaceErr(ctx, err)
		return
	}

	http_utils.WriteResponse(ctx, resp, nil)
}

// SyncOrderStatus 同步订单状态
// POST /marketplace/orders/:order_no/sync
func (r *MarketplaceResource) SyncOrderStatus(ctx *gin.Context) {
	orderNo := ctx.Param("order_no")
	if orderNo == "" {
		http_utils.WriteResponse(ctx, nil, nil)
		return
	}

	orderInfo, err := r.orderService.SyncOrderStatus(ctx.Request.Context(), orderNo)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, orderInfo, nil)
}

// ==================== 积分接口 ====================

// GetEcoinBalanceRequest 获取积分余额请求
type GetEcoinBalanceRequest struct {
	UserId uint64 `form:"user_id" binding:"required"` // 用户ID
}

// GetEcoinBalance 获取用户积分余额
// GET /marketplace/ecoin/balance
func (r *MarketplaceResource) GetEcoinBalance(ctx *gin.Context) {
	var req GetEcoinBalanceRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	u, err := r.userService.GetUserById(ctx.Request.Context(), uint(req.UserId))
	if err != nil || u == nil {
		http_utils.WriteResponse(ctx, nil, errors.New("user not found"))
		return
	}
	if u.CompanyId > 0 {
		ce, err := r.companyEcoinService.GetCompanyEcoin(ctx.Request.Context(), u.CompanyId)
		if err != nil {
			http_utils.WriteResponse(ctx, nil, err)
			return
		}
		http_utils.WriteResponse(ctx, gin.H{
			"user_id":         u.Id,
			"company_id":      u.CompanyId,
			"available_stock": ce.AvailableStock,
			"stock_groups":    ce.StockGroups,
			"id":              ce.Id,
			"ctime":           ce.Ctime,
			"mtime":           ce.Mtime,
		}, nil)
		return
	}

	ecoinInfo, err := r.ecoinService.GetUserEcoin(ctx.Request.Context(), req.UserId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, ecoinInfo, nil)
}

// GetEcoinStockGroups 获取用户积分分组
// GET /marketplace/ecoin/stock_groups
func (r *MarketplaceResource) GetEcoinStockGroups(ctx *gin.Context) {
	var req GetEcoinBalanceRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	u, err := r.userService.GetUserById(ctx.Request.Context(), uint(req.UserId))
	if err != nil || u == nil {
		http_utils.WriteResponse(ctx, nil, errors.New("user not found"))
		return
	}
	if u.CompanyId > 0 {
		ce, err := r.companyEcoinService.GetCompanyEcoin(ctx.Request.Context(), u.CompanyId)
		if err != nil {
			http_utils.WriteResponse(ctx, nil, err)
			return
		}
		http_utils.WriteResponse(ctx, gin.H{
			"total_stock": ce.AvailableStock,
			"list":        ce.StockGroups,
			"company_id":  u.CompanyId,
			"user_id":     u.Id,
		}, nil)
		return
	}

	resp, err := r.ecoinService.GetEcoinStockGroupList(ctx.Request.Context(), &ecoin.EcoinStockGroupListRequest{
		UserId: req.UserId,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	http_utils.WriteResponse(ctx, resp, nil)
}

// GetEcoinTransactionsRequest 获取积分流水请求
type GetEcoinTransactionsRequest struct {
	UserId uint64 `form:"user_id" binding:"required"` // 用户ID
	Offset int    `form:"offset"`                     // 偏移量
	Limit  int    `form:"limit"`                      // 每页数量
}

// GetEcoinTransactions 获取积分流水列表
// GET /marketplace/ecoin/transactions
func (r *MarketplaceResource) GetEcoinTransactions(ctx *gin.Context) {
	var req GetEcoinTransactionsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	u, err := r.userService.GetUserById(ctx.Request.Context(), uint(req.UserId))
	if err != nil || u == nil {
		http_utils.WriteResponse(ctx, nil, errors.New("user not found"))
		return
	}
	if u.CompanyId > 0 {
		resp, err := r.companyEcoinService.GetTransactionList(ctx.Request.Context(), &companyecoin.TransactionListRequest{
			CompanyId: u.CompanyId,
			Offset:    req.Offset,
			Limit:     req.Limit,
		})
		if err != nil {
			http_utils.WriteResponse(ctx, nil, err)
			return
		}
		if resp != nil {
			for _, tx := range resp.List {
				r.fillCompanyTxOperatorLabel(ctx.Request.Context(), tx)
			}
		}
		http_utils.WriteResponse(ctx, resp, nil)
		return
	}

	resp, err := r.ecoinService.GetEcoinTransactionList(ctx.Request.Context(), &ecoin.EcoinTransactionListRequest{
		UserId: req.UserId,
		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, resp, nil)
}

func (r *MarketplaceResource) fillCompanyTxOperatorLabel(ctx context.Context, tx *companyecoin_model.CompanyEcoinTransaction) {
	if tx == nil {
		return
	}
	if tx.OperatorUserId == 0 {
		tx.OperatorLabel = "系统"
		return
	}
	op, err := r.userService.GetUserById(ctx, uint(tx.OperatorUserId))
	if err != nil || op == nil {
		tx.OperatorLabel = fmt.Sprintf("用户 #%d", tx.OperatorUserId)
		return
	}
	if op.Email != "" {
		tx.OperatorLabel = op.Email
		return
	}
	if op.TelNo != "" {
		tx.OperatorLabel = op.TelNo
		return
	}
	tx.OperatorLabel = fmt.Sprintf("用户 #%d", op.Id)
}

// RechargeEcoinRequest 积分充值请求
// GetRechargeConfig 获取积分充值配置
// GET /marketplace/ecoin/recharge_config
func (r *MarketplaceResource) GetRechargeConfig(ctx *gin.Context) {
	unitPrice := config.GetConf().EcoinUnitPrice
	// 最低充值数量 = ceil(0.01 / unitPrice)，保证支付金额 >= 1 分
	minAmount := int(math.Ceil(0.01 / float64(unitPrice)))
	if minAmount < 1 {
		minAmount = 1
	}

	http_utils.WriteResponse(ctx, map[string]interface{}{
		"unit_price": unitPrice,
		"min_amount": minAmount,
	}, nil)
}

type RechargeEcoinRequest struct {
	UserId    uint64 `json:"user_id" binding:"required"`  // 用户ID
	Amount    int    `json:"amount" binding:"required"`   // 充值积分数量
	PayType   string `json:"pay_type" binding:"required"` // 支付类型：money
	Channel   string `json:"channel"`                     // 支付渠道：wechat
	PayMethod string `json:"pay_method"`                  // 支付方式：native/jsapi/h5
}

// RechargeEcoin 积分充值
// POST /marketplace/ecoin/recharge
func (r *MarketplaceResource) RechargeEcoin(ctx *gin.Context) {
	var req RechargeEcoinRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	// 创建充值订单（不需要SKU，直接使用特殊参数）
	orderResp, err := r.orderService.CreateOrder(ctx.Request.Context(), &order.CreateOrderRequest{
		UserId:          req.UserId,
		SkuItems:        nil,
		PayType:         req.PayType,
		Remark:          "积分充值",
		IsEcoinRecharge: true,
		EcoinUnits:      req.Amount,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	// 如果是货币支付，需要发起支付
	if req.PayType == "money" && req.Channel != "" {
		clientIP := ctx.ClientIP()
		payResp, err := r.orderService.PayOrder(ctx.Request.Context(), &order.PayOrderRequest{
			OrderNo:   orderResp.Order.OrderNo,
			Channel:   req.Channel,
			PayMethod: req.PayMethod,
			ClientIP:  clientIP,
		})
		if err != nil {
			http_utils.WriteResponse(ctx, nil, err)
			return
		}

		// 返回支付信息
		http_utils.WriteResponse(ctx, map[string]interface{}{
			"order":        orderResp.Order,
			"payment_info": payResp,
		}, nil)
		return
	}

	http_utils.WriteResponse(ctx, orderResp, nil)
}

// ==================== 支付方式 ====================

// PaymentMethod 支付方式
type PaymentMethod struct {
	Channel   string `json:"channel"`    // 支付渠道
	Name      string `json:"name"`       // 支付名称
	PayMethod string `json:"pay_method"` // 支付方式
	Icon      string `json:"icon"`       // 图标
}

// GetPaymentMethods 获取支付方式列表
// GET /marketplace/payment-methods
func (r *MarketplaceResource) GetPaymentMethods(ctx *gin.Context) {
	// 返回支持的支付方式
	methods := []*PaymentMethod{
		{
			Channel:   "ecoin",
			Name:      "积分支付",
			PayMethod: "ecoin",
			Icon:      "",
		},
		{
			Channel:   "wechat",
			Name:      "微信扫码支付",
			PayMethod: "native",
			Icon:      "",
		},
	}

	http_utils.WriteResponse(ctx, methods, nil)
}

// ==================== 用户登录接口 ====================

// LoginRequest 登录请求（电话号码/邮箱登录）
type LoginRequest struct {
	TelNo     string `json:"tel_no"`     // 手机号
	Email     string `json:"email"`      // 邮箱
	Secret    string `json:"secret"`     // 用户密钥
	LoginKind string `json:"login_kind"` // personal | enterprise
	CompanyId uint64 `json:"company_id"` // 企业登录必填
}

// Login 用户登录（电话号码/邮箱）
// POST /marketplace/login
func (r *MarketplaceResource) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	resp, err := r.userService.Login(ctx.Request.Context(), &user.LoginRequest{
		TelNo:     req.TelNo,
		Email:     req.Email,
		Secret:    req.Secret,
		LoginKind: req.LoginKind,
		CompanyId: req.CompanyId,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, resp, nil)
}

// BindUserRequest 用户绑定（与业务平台账号关联，成功后返回 session）
type BindUserRequest struct {
	BizCode     string `json:"biz_code" binding:"required"`    // 业务平台代码
	BizUserId   uint64 `json:"biz_user_id,string" binding:"required"` // 业务平台用户 ID
	TelNo       string `json:"tel_no"`                         // 手机号
	Email       string `json:"email"`                          // 邮箱
	Password    string `json:"password" binding:"required"`    // 密码
	CompanyId   uint64 `json:"company_id"`                     // 企业：已有企业 ID
	CompanyName string `json:"company_name"`                   // 企业：企业名称（可新建）
}

// ListCompanies 企业名称下拉（公开）
// GET /marketplace/companies?q=&limit=
func (r *MarketplaceResource) ListCompanies(ctx *gin.Context) {
	q := strings.TrimSpace(ctx.Query("q"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	list, err := r.companyRepo.SearchByName(ctx.Request.Context(), q, limit)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, c := range list {
		out = append(out, gin.H{"id": c.Id, "name": c.Name})
	}
	http_utils.WriteResponse(ctx, out, nil)
}

// ListBizCodes 业务平台编码枚举（供绑定页下拉；数据由 biz_code_enum_tab 维护）
// GET /marketplace/biz_codes
func (r *MarketplaceResource) ListBizCodes(ctx *gin.Context) {
	list, err := r.bizCodeRepo.ListAll(ctx.Request.Context())
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, row := range list {
		out = append(out, gin.H{
			"code":  row.Code,
			"name":  row.Name,
			"scope": row.Scope,
		})
	}
	http_utils.WriteResponse(ctx, out, nil)
}

// BindUser 绑定用户
// POST /marketplace/user/bind（免 session，与 /marketplace/login 相同）
func (r *MarketplaceResource) BindUser(ctx *gin.Context) {
	var req BindUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	resp, err := r.userService.BindUser(ctx.Request.Context(), &user.BindUserRequest{
		Password:    req.Password,
		BizCode:     req.BizCode,
		BizUserId:   req.BizUserId,
		TelNo:       req.TelNo,
		Email:       req.Email,
		CompanyId:   req.CompanyId,
		CompanyName: req.CompanyName,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, resp, nil)
}

// CheckBizBinding 校验 biz_code + biz_user_id 是否已有绑定（免登录，供登录页预检）
// GET /marketplace/user/bind/check?biz_code=&biz_user_id=
func (r *MarketplaceResource) CheckBizBinding(ctx *gin.Context) {
	bizCode := strings.TrimSpace(ctx.Query("biz_code"))
	bizUserIdStr := strings.TrimSpace(ctx.Query("biz_user_id"))
	if bizCode == "" {
		http_utils.WriteResponse(ctx, nil, errors.New("biz_code is required"))
		return
	}
	if bizUserIdStr == "" {
		http_utils.WriteResponse(ctx, nil, errors.New("biz_user_id is required"))
		return
	}
	bizUserId, err := strconv.ParseUint(bizUserIdStr, 10, 64)
	if err != nil || bizUserId == 0 {
		http_utils.WriteResponse(ctx, nil, errors.New("biz_user_id must be a positive integer"))
		return
	}
	u, err := r.userService.GetUserByBiz(ctx.Request.Context(), bizCode, bizUserId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	http_utils.WriteResponse(ctx, map[string]bool{"bound": u != nil}, nil)
}

// UnbindUserBody 解绑请求体（当前用户从 session 解析）
type UnbindUserBody struct {
	BizCode string `json:"biz_code" binding:"required"` // 业务平台代码
}

// UnbindUser 解绑
// POST /marketplace/user/unbind
func (r *MarketplaceResource) UnbindUser(ctx *gin.Context) {
	u, err := auth_utils.UserFromGinContext(ctx)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	var req UnbindUserBody
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	if err := r.userService.UnbindUser(ctx.Request.Context(), &user.UnbindUserRequest{
		UserId:  u.Id,
		BizCode: req.BizCode,
	}); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, nil, nil)
}

// ListUserBindings 当前用户的业务平台绑定列表
// GET /marketplace/user/bindings
func (r *MarketplaceResource) ListUserBindings(ctx *gin.Context) {
	u, err := auth_utils.UserFromGinContext(ctx)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	bindings, err := r.userService.GetBindingsByUserId(ctx.Request.Context(), u.Id)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, bindings, nil)
}

// ChangePasswordBody 修改登录密钥
type ChangePasswordBody struct {
	OldSecret string `json:"old_secret" binding:"required"`
	NewSecret string `json:"new_secret" binding:"required"`
}

// ChangePassword 修改当前用户登录密钥
// POST /marketplace/user/password
func (r *MarketplaceResource) ChangePassword(ctx *gin.Context) {
	u, err := auth_utils.UserFromGinContext(ctx)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	var req ChangePasswordBody
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	if err := r.userService.ChangePassword(ctx.Request.Context(), u.Id, &user.ChangePasswordRequest{
		OldSecret: req.OldSecret,
		NewSecret: req.NewSecret,
	}); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	http_utils.WriteResponse(ctx, gin.H{"ok": true}, nil)
}

// UpdateProfile 更新当前用户手机号、邮箱（唯一性由服务层校验，并刷新 Redis session）
// POST /marketplace/user/profile
func (r *MarketplaceResource) UpdateProfile(ctx *gin.Context) {
	u, err := auth_utils.UserFromGinContext(ctx)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	var reqBody user.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	authHeader := strings.TrimSpace(ctx.GetHeader("Authorization"))
	token := authHeader
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	}
	if token == "" {
		http_utils.WriteResponse(ctx, nil, errors.New("missing authorization"))
		return
	}
	conf := config.GetConf()
	if conf == nil {
		http_utils.WriteResponse(ctx, nil, errors.New("config not initialized"))
		return
	}
	sessionId, err := auth_utils.ParseAuthToken(token, conf.Auth.AesKey)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	fresh, err := r.userService.UpdateProfile(ctx.Request.Context(), u.Id, &reqBody, sessionId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	http_utils.WriteResponse(ctx, fresh, nil)
}

// ==================== 购物车接口 ====================

// AddToCartRequest 添加购物车请求
type AddToCartRequest struct {
	UserId   uint64 `json:"user_id" binding:"required"`  // 用户ID
	SkuId    uint64 `json:"sku_id" binding:"required"`   // 商品ID
	Quantity int    `json:"quantity" binding:"required"` // 数量
}

// AddToCart 添加商品到购物车
// POST /marketplace/shopping_cart/add
func (r *MarketplaceResource) AddToCart(ctx *gin.Context) {
	var req AddToCartRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	item, err := r.cartService.AddToCart(ctx.Request.Context(), &cart.AddToCartRequest{
		UserId:   req.UserId,
		SkuId:    req.SkuId,
		Quantity: req.Quantity,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, item, nil)
}

// RemoveFromCartRequest 移除购物车请求
type RemoveFromCartRequest struct {
	UserId uint64 `json:"user_id" binding:"required"` // 用户ID
	SkuId  uint64 `json:"sku_id" binding:"required"`  // 商品ID
}

// RemoveFromCart 从购物车移除商品
// POST /marketplace/shopping_cart/remove
func (r *MarketplaceResource) RemoveFromCart(ctx *gin.Context) {
	var req RemoveFromCartRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	err := r.cartService.RemoveFromCart(ctx.Request.Context(), &cart.RemoveFromCartRequest{
		UserId: req.UserId,
		SkuId:  req.SkuId,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, nil, nil)
}

// UpdateCartItemRequest 更新购物车商品数量请求
type UpdateCartItemRequest struct {
	UserId   uint64 `json:"user_id" binding:"required"` // 用户ID
	SkuId    uint64 `json:"sku_id" binding:"required"`  // 商品ID
	Quantity int    `json:"quantity"`                   // 数量（0表示删除）
}

// UpdateCartItem 更新购物车商品数量
// POST /marketplace/shopping_cart/update
func (r *MarketplaceResource) UpdateCartItem(ctx *gin.Context) {
	var req UpdateCartItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	item, err := r.cartService.UpdateCartItem(ctx.Request.Context(), &cart.UpdateCartItemRequest{
		UserId:   req.UserId,
		SkuId:    req.SkuId,
		Quantity: req.Quantity,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, item, nil)
}

// ClearCartRequest 清空购物车请求
type ClearCartRequest struct {
	UserId uint64 `json:"user_id" binding:"required"` // 用户ID
}

// ClearCart 清空购物车
// POST /marketplace/shopping_cart/clear
func (r *MarketplaceResource) ClearCart(ctx *gin.Context) {
	var req ClearCartRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	err := r.cartService.ClearCart(ctx.Request.Context(), req.UserId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, nil, nil)
}

// GetCartListRequest 获取购物车列表请求
type GetCartListRequest struct {
	UserId uint64 `form:"user_id" binding:"required"` // 用户ID
}

// GetCartList 获取购物车列表
// GET /marketplace/shopping_cart/list
func (r *MarketplaceResource) GetCartList(ctx *gin.Context) {
	var req GetCartListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	resp, err := r.cartService.GetCartList(ctx.Request.Context(), req.UserId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, resp, nil)
}

// CartCheckoutRequest 购物车下单请求
type CartCheckoutRequest struct {
	UserId  uint64   `json:"user_id" binding:"required"` // 用户ID
	SkuIds  []uint64 `json:"sku_ids" binding:"required"` // 要下单的商品ID列表
	PayType string   `json:"pay_type"`                   // 支付类型：ecoin/money，不传默认money
	Remark  string   `json:"remark"`                     // 备注
}

// CartCheckout 购物车下单（下单并移除对应商品）
// POST /marketplace/shopping_cart/checkout
func (r *MarketplaceResource) CartCheckout(ctx *gin.Context) {
	var req CartCheckoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	payType := req.PayType
	if payType == "" {
		payType = "money"
	}

	resp, err := r.cartService.Checkout(ctx.Request.Context(), &cart.CheckoutRequest{
		UserId:  req.UserId,
		SkuIds:  req.SkuIds,
		PayType: payType,
		Remark:  req.Remark,
	})
	if err != nil {
		writeMarketplaceErr(ctx, err)
		return
	}

	http_utils.WriteResponse(ctx, resp, nil)
}

// Router 注册路由
func (r *MarketplaceResource) Router() registry.Registry {
	return func(router *gin.Engine) {
		group := router.Group("/marketplace")
		{
			// 用户登录接口
			group.POST("/login", r.Login)
			group.GET("/biz_codes", r.ListBizCodes)
			group.GET("/companies", r.ListCompanies)
			group.POST("/user/bind", r.BindUser)
			group.GET("/user/bind/check", r.CheckBizBinding)
			group.POST("/user/unbind", r.UnbindUser)
			group.GET("/user/bindings", r.ListUserBindings)
			group.POST("/user/password", r.ChangePassword)
			group.POST("/user/profile", r.UpdateProfile)

			// 商品接口
			group.GET("/skus", r.ListSkus)
			group.GET("/skus/:id", r.GetSkuDetail)

			// 订单接口
			group.POST("/checkout", r.CreateOrder)
			group.GET("/orders", r.ListOrders)
			group.GET("/orders/:order_no", r.GetOrderDetail)
			group.POST("/orders/:order_no/cancel", r.CancelOrder)
			group.POST("/orders/:order_no/pay", r.PayOrder)
			group.POST("/orders/:order_no/sync", r.SyncOrderStatus)

			// 积分接口
			ecoinGroup := group.Group("/ecoin")
			{
				ecoinGroup.GET("/balance", r.GetEcoinBalance)
				ecoinGroup.GET("/stock_groups", r.GetEcoinStockGroups)
				ecoinGroup.GET("/transactions", r.GetEcoinTransactions)
				ecoinGroup.GET("/recharge_config", r.GetRechargeConfig)
				ecoinGroup.POST("/recharge", r.RechargeEcoin)
			}

			// 支付方式
			group.GET("/payment-methods", r.GetPaymentMethods)

			// 购物车接口
			cartGroup := group.Group("/shopping_cart")
			{
				cartGroup.POST("/add", r.AddToCart)
				cartGroup.POST("/remove", r.RemoveFromCart)
				cartGroup.POST("/update", r.UpdateCartItem)
				cartGroup.POST("/clear", r.ClearCart)
				cartGroup.GET("/list", r.GetCartList)
				cartGroup.POST("/checkout", r.CartCheckout)
			}
		}
	}
}
