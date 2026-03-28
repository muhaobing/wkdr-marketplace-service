package ops

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/muhaobing/std-go/restserver/registry"

	"wdkr-marketplace-service/internal/common/utils/http_utils"
	"wdkr-marketplace-service/internal/domain/order"
	"wdkr-marketplace-service/internal/domain/sku"
)

// OpsResource 运营接口资源（面向运营人员）
type OpsResource struct {
	skuService   sku.SkuService
	orderService order.OrderService
}

// NewOpsResource 创建运营资源实例
func NewOpsResource(
	skuService sku.SkuService,
	orderService order.OrderService,
) *OpsResource {
	return &OpsResource{
		skuService:   skuService,
		orderService: orderService,
	}
}

// ==================== 商品管理接口 ====================

// CreateSkuRequest 创建商品请求
type CreateSkuRequest struct {
	BizCode            string  `json:"biz_code" binding:"required"` // 业务编码
	SkuCode            string  `json:"sku_code" binding:"required"` // 商品代码
	SkuName            string  `json:"sku_name" binding:"required"` // 商品名称
	SkuAvatar          string  `json:"sku_avatar"`                  // 商品图标
	SkuDesc            string  `json:"sku_desc"`                    // 商品描述
	Cost               float32 `json:"cost" binding:"required"`     // 商品售价(积分)
	DeliveryMethod     string  `json:"delivery_method"`             // 履约回调接口（fulfill_mode=0）
	FulfillMode        uint8   `json:"fulfill_mode"`                // 履约模式：0-接口回调，1-积分发放
	FulfillEcoinAmount float64 `json:"fulfill_ecoin_amount"`        // 积分发放：每件发放积分数
	MultiSelect        uint8   `json:"multi_select"`                // 是否支持多选：0-不支持，1-支持
}

// CreateSku 创建商品
// POST /ops/skus/create
func (r *OpsResource) CreateSku(ctx *gin.Context) {
	var req CreateSkuRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	skuInfo, err := r.skuService.CreateSku(ctx.Request.Context(), &sku.CreateSkuRequest{
		BizCode:            req.BizCode,
		SkuCode:            req.SkuCode,
		SkuName:            req.SkuName,
		SkuAvatar:          req.SkuAvatar,
		SkuDesc:            req.SkuDesc,
		Cost:               req.Cost,
		DeliveryMethod:     req.DeliveryMethod,
		FulfillMode:        req.FulfillMode,
		FulfillEcoinAmount: req.FulfillEcoinAmount,
		MultiSelect:        req.MultiSelect,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, skuInfo, nil)
}

// EditSkuRequest 编辑商品请求
type EditSkuRequest struct {
	Id                 uint64  `json:"id" binding:"required"`       // 商品ID
	SkuName            string  `json:"sku_name" binding:"required"` // 商品名称
	SkuAvatar          string  `json:"sku_avatar"`                  // 商品图标
	SkuDesc            string  `json:"sku_desc"`                    // 商品描述
	Cost               float32 `json:"cost" binding:"required"`     // 商品售价(积分)
	DeliveryMethod     string  `json:"delivery_method"`             // 履约回调接口（fulfill_mode=0）
	FulfillMode        uint8   `json:"fulfill_mode"`                // 履约模式：0-接口回调，1-积分发放
	FulfillEcoinAmount float64 `json:"fulfill_ecoin_amount"`        // 积分发放：每件发放积分数
	MultiSelect        uint8   `json:"multi_select"`                // 是否支持多选：0-不支持，1-支持
}

// EditSku 编辑商品
// POST /ops/skus/edit
func (r *OpsResource) EditSku(ctx *gin.Context) {
	var req EditSkuRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	skuInfo, err := r.skuService.EditSku(ctx.Request.Context(), &sku.EditSkuRequest{
		Id:                 req.Id,
		SkuName:            req.SkuName,
		SkuAvatar:          req.SkuAvatar,
		SkuDesc:            req.SkuDesc,
		Cost:               req.Cost,
		DeliveryMethod:     req.DeliveryMethod,
		FulfillMode:        req.FulfillMode,
		FulfillEcoinAmount: req.FulfillEcoinAmount,
		MultiSelect:        req.MultiSelect,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, skuInfo, nil)
}

// GetSkuDetail 获取商品详情
// GET /ops/skus/:id
func (r *OpsResource) GetSkuDetail(ctx *gin.Context) {
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

	http_utils.WriteResponse(ctx, skuInfo, nil)
}

// ListSkusRequest 商品列表请求
type ListSkusRequest struct {
	BizCode string `form:"biz_code"` // 业务编码（可选）
	SkuName string `form:"sku_name"` // 商品名称（模糊查询，可选）
	Status  *uint8 `form:"status"`   // 上架状态过滤（可选）
	Offset  int    `form:"offset"`   // 偏移量
	Limit   int    `form:"limit"`    // 每页数量
}

// ListSkus 获取商品列表（支持多条件组合查询）
// GET /ops/skus
func (r *OpsResource) ListSkus(ctx *gin.Context) {
	var req ListSkusRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	resp, err := r.skuService.ListSkus(ctx.Request.Context(), &sku.ListSkuRequest{
		BizCode: req.BizCode,
		SkuName: req.SkuName,
		Status:  req.Status,
		Offset:  req.Offset,
		Limit:   req.Limit,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, resp, nil)
}

// ListingSkuRequest 上架商品请求
type ListingSkuRequest struct {
	Id uint64 `json:"id" binding:"required"` // 商品ID
}

// ListingSku 上架商品
// POST /ops/skus/listing
func (r *OpsResource) ListingSku(ctx *gin.Context) {
	var req ListingSkuRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	err := r.skuService.ListingSku(ctx.Request.Context(), req.Id)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, nil, nil)
}

// DelistingSkuRequest 下架商品请求
type DelistingSkuRequest struct {
	Id uint64 `json:"id" binding:"required"` // 商品ID
}

// DelistingSku 下架商品
// POST /ops/skus/delisting
func (r *OpsResource) DelistingSku(ctx *gin.Context) {
	var req DelistingSkuRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	err := r.skuService.DelistingSku(ctx.Request.Context(), req.Id)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, nil, nil)
}

// DeleteSkuRequest 删除商品请求
type DeleteSkuRequest struct {
	Id uint64 `json:"id" binding:"required"` // 商品ID
}

// DeleteSku 删除商品
// POST /ops/skus/delete
func (r *OpsResource) DeleteSku(ctx *gin.Context) {
	var req DeleteSkuRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	err := r.skuService.DeleteSku(ctx.Request.Context(), req.Id)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, nil, nil)
}

// ==================== 订单管理接口 ====================

// ListOrdersRequest 订单列表请求
type ListOrdersRequest struct {
	UserId uint64 `form:"user_id"` // 用户ID（可选）
	Status *uint8 `form:"status"`  // 订单状态过滤（可选）
	Offset int    `form:"offset"`  // 偏移量
	Limit  int    `form:"limit"`   // 每页数量
}

// ListOrders 获取订单列表
// GET /ops/orders
func (r *OpsResource) ListOrders(ctx *gin.Context) {
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

// GetOrderDetail 获取订单详情
// GET /ops/orders/:order_no
func (r *OpsResource) GetOrderDetail(ctx *gin.Context) {
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

// RefundOrderRequest 退款订单请求
type RefundOrderRequest struct {
	OrderNo string `json:"order_no" binding:"required"` // 订单号
	Reason  string `json:"reason"`                      // 退款原因
}

// RefundOrder 退款订单
// POST /ops/orders/refund
func (r *OpsResource) RefundOrder(ctx *gin.Context) {
	var req RefundOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	err := r.orderService.RefundOrder(ctx.Request.Context(), &order.RefundOrderRequest{
		OrderNo: req.OrderNo,
		Reason:  req.Reason,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, nil, nil)
}

// FulfillOrderRequest 履约订单请求
type FulfillOrderRequest struct {
	OrderNo   string `json:"order_no" binding:"required"`    // 订单号
	BizUserId string `json:"biz_user_id" binding:"required"` // 业务用户ID
}

// FulfillOrder 履约订单
// POST /ops/orders/fulfill
func (r *OpsResource) FulfillOrder(ctx *gin.Context) {
	var req FulfillOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	err := r.orderService.FulfillOrder(ctx.Request.Context(), req.OrderNo, req.BizUserId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, nil, nil)
}

// Router 注册路由
func (r *OpsResource) Router() registry.Registry {
	return func(router *gin.Engine) {
		group := router.Group("/ops")
		{
			// 商品管理接口
			group.POST("/skus/create", r.CreateSku)
			group.POST("/skus/edit", r.EditSku)
			group.GET("/skus/:id", r.GetSkuDetail)
			group.GET("/skus", r.ListSkus)
			group.POST("/skus/listing", r.ListingSku)
			group.POST("/skus/delisting", r.DelistingSku)
			group.POST("/skus/delete", r.DeleteSku)

			// 订单管理接口
			group.GET("/orders", r.ListOrders)
			group.GET("/orders/:order_no", r.GetOrderDetail)
			group.POST("/orders/refund", r.RefundOrder)
			group.POST("/orders/fulfill", r.FulfillOrder)
		}
	}
}
