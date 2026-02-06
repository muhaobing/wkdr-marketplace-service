package marketplace

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/muhaobing-eng/std-go/restserver/registry"

	"wdkr-marketplace-service/internal/common/utils/http_utils"
	"wdkr-marketplace-service/internal/domain/cart"
	"wdkr-marketplace-service/internal/domain/ecoin"
	"wdkr-marketplace-service/internal/domain/order"
	"wdkr-marketplace-service/internal/domain/sku"
	skumodel "wdkr-marketplace-service/internal/domain/sku/sku_model"
	"wdkr-marketplace-service/internal/domain/user"
)

// MarketplaceResource 商城接口资源（面向用户）
type MarketplaceResource struct {
	skuService   sku.SkuService
	orderService order.OrderService
	ecoinService ecoin.EcoinService
	userService  user.UserService
	cartService  cart.CartService
}

// NewMarketplaceResource 创建商城资源实例
func NewMarketplaceResource(
	skuService sku.SkuService,
	orderService order.OrderService,
	ecoinService ecoin.EcoinService,
	userService user.UserService,
	cartService cart.CartService,
) *MarketplaceResource {
	return &MarketplaceResource{
		skuService:   skuService,
		orderService: orderService,
		ecoinService: ecoinService,
		userService:  userService,
		cartService:  cartService,
	}
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
func (r *MarketplaceResource) ListSkus(ctx *gin.Context) {
	var req ListSkusRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	// 只查询已上架的商品
	onlineStatus := skumodel.SkuStatusOnline
	resp, err := r.skuService.ListSkus(ctx.Request.Context(), &sku.ListSkuRequest{
		SkuName: req.SkuName,
		Status:  &onlineStatus,
		Offset:  req.Offset,
		Limit:   req.Limit,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, resp, nil)
}

// GetSkuDetail 获取商品详情
// GET /marketplace/skus/:id
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

	http_utils.WriteResponse(ctx, skuInfo, nil)
}

// ==================== 订单接口 ====================

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	UserId   uint64                `json:"user_id" binding:"required"`   // 用户ID
	SkuItems []*order.SkuOrderItem `json:"sku_items" binding:"required"` // SKU列表
	PayType  string                `json:"pay_type" binding:"required"`  // 支付类型：ecoin/money
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

	resp, err := r.orderService.CreateOrder(ctx.Request.Context(), &order.CreateOrderRequest{
		UserId:   req.UserId,
		SkuItems: req.SkuItems,
		PayType:  req.PayType,
		Remark:   req.Remark,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
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
	Channel   string `json:"channel" binding:"required"`    // 支付渠道：wechat/alipay
	PayMethod string `json:"pay_method" binding:"required"` // 支付方式：native/jsapi/h5
	ClientIP  string `json:"client_ip"`                     // 客户端IP（H5支付需要）
	OpenId    string `json:"open_id"`                       // 用户OpenID（JSAPI支付需要）
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
		http_utils.WriteResponse(ctx, nil, err)
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

// GetUserEcoin 获取用户积分
// GET /marketplace/ecoin
func (r *MarketplaceResource) GetUserEcoin(ctx *gin.Context) {
	userIdStr := ctx.Query("user_id")
	if userIdStr == "" {
		http_utils.WriteResponse(ctx, nil, nil)
		return
	}

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
		{
			Channel:   "wechat",
			Name:      "微信H5支付",
			PayMethod: "h5",
			Icon:      "",
		},
		{
			Channel:   "wechat",
			Name:      "微信JSAPI支付",
			PayMethod: "jsapi",
			Icon:      "",
		},
	}

	http_utils.WriteResponse(ctx, methods, nil)
}

// ==================== 用户登录接口 ====================

// LoginRequest 登录请求（电话号码/邮箱登录）
type LoginRequest struct {
	TelNo  string `json:"tel_no"` // 手机号
	Email  string `json:"email"`  // 邮箱
	Secret string `json:"secret"` // 用户密钥
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
		TelNo:  req.TelNo,
		Email:  req.Email,
		Secret: req.Secret,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, resp, nil)
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
	UserId  uint64   `json:"user_id" binding:"required"`  // 用户ID
	SkuIds  []uint64 `json:"sku_ids" binding:"required"`  // 要下单的商品ID列表
	PayType string   `json:"pay_type" binding:"required"` // 支付类型：ecoin/money
	Remark  string   `json:"remark"`                      // 备注
}

// CartCheckout 购物车下单（下单并移除对应商品）
// POST /marketplace/shopping_cart/checkout
func (r *MarketplaceResource) CartCheckout(ctx *gin.Context) {
	var req CartCheckoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	resp, err := r.cartService.Checkout(ctx.Request.Context(), &cart.CheckoutRequest{
		UserId:  req.UserId,
		SkuIds:  req.SkuIds,
		PayType: req.PayType,
		Remark:  req.Remark,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
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
			group.GET("/ecoin", r.GetUserEcoin)

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
