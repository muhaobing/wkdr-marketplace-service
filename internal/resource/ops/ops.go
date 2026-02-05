package ops

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/muhaobing-eng/std-go/restserver/registry"

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
	BizCode        string  `json:"biz_code" binding:"required"` // 业务编码
	SkuCode        string  `json:"sku_code" binding:"required"` // 商品代码
	SkuName        string  `json:"sku_name" binding:"required"` // 商品名称
	SkuAvatar      string  `json:"sku_avatar"`                  // 商品图标
	SkuDesc        string  `json:"sku_desc"`                    // 商品描述
	Cost           float32 `json:"cost" binding:"required"`     // 商品售价(积分)
	DeliveryMethod string  `json:"delivery_method"`             // 履约回调接口
}

// CreateSku 创建商品
// POST /ops/skus
func (r *OpsResource) CreateSku(ctx *gin.Context) {
	var req CreateSkuRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	skuInfo, err := r.skuService.CreateSku(ctx.Request.Context(), &sku.CreateSkuRequest{
		BizCode:        req.BizCode,
		SkuCode:        req.SkuCode,
		SkuName:        req.SkuName,
		SkuAvatar:      req.SkuAvatar,
		SkuDesc:        req.SkuDesc,
		Cost:           req.Cost,
		DeliveryMethod: req.DeliveryMethod,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, skuInfo, nil)
}

// EditSkuRequest 编辑商品请求
type EditSkuRequest struct {
	SkuName        string  `json:"sku_name" binding:"required"` // 商品名称
	SkuAvatar      string  `json:"sku_avatar"`                  // 商品图标
	SkuDesc        string  `json:"sku_desc"`                    // 商品描述
	Cost           float32 `json:"cost" binding:"required"`     // 商品售价(积分)
	DeliveryMethod string  `json:"delivery_method"`             // 履约回调接口
}

// EditSku 编辑商品
// PUT /ops/skus/:id
func (r *OpsResource) EditSku(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	var req EditSkuRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	skuInfo, err := r.skuService.EditSku(ctx.Request.Context(), &sku.EditSkuRequest{
		Id:             id,
		SkuName:        req.SkuName,
		SkuAvatar:      req.SkuAvatar,
		SkuDesc:        req.SkuDesc,
		Cost:           req.Cost,
		DeliveryMethod: req.DeliveryMethod,
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
	BizCode string `form:"biz_code" binding:"required"` // 业务编码
	Status  *uint8 `form:"status"`                      // 上架状态过滤（可选）
	Offset  int    `form:"offset"`                      // 偏移量
	Limit   int    `form:"limit"`                       // 每页数量
}

// ListSkus 获取商品列表
// GET /ops/skus
func (r *OpsResource) ListSkus(ctx *gin.Context) {
	var req ListSkusRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	resp, err := r.skuService.ListSkus(ctx.Request.Context(), &sku.ListSkuRequest{
		BizCode: req.BizCode,
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

// ListingSku 上架商品
// POST /ops/skus/:id/listing
func (r *OpsResource) ListingSku(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	err = r.skuService.ListingSku(ctx.Request.Context(), id)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, nil, nil)
}

// DelistingSku 下架商品
// POST /ops/skus/:id/delisting
func (r *OpsResource) DelistingSku(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	err = r.skuService.DelistingSku(ctx.Request.Context(), id)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	http_utils.WriteResponse(ctx, nil, nil)
}

// DeleteSku 删除商品
// DELETE /ops/skus/:id
func (r *OpsResource) DeleteSku(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	err = r.skuService.DeleteSku(ctx.Request.Context(), id)
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
	Reason string `json:"reason"` // 退款原因
}

// RefundOrder 退款订单
// POST /ops/orders/:order_no/refund
func (r *OpsResource) RefundOrder(ctx *gin.Context) {
	orderNo := ctx.Param("order_no")
	if orderNo == "" {
		http_utils.WriteResponse(ctx, nil, nil)
		return
	}

	var req RefundOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	err := r.orderService.RefundOrder(ctx.Request.Context(), &order.RefundOrderRequest{
		OrderNo: orderNo,
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
	BizUserId string `json:"biz_user_id" binding:"required"` // 业务用户ID
}

// FulfillOrder 履约订单
// POST /ops/orders/:order_no/fulfill
func (r *OpsResource) FulfillOrder(ctx *gin.Context) {
	orderNo := ctx.Param("order_no")
	if orderNo == "" {
		http_utils.WriteResponse(ctx, nil, nil)
		return
	}

	var req FulfillOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	err := r.orderService.FulfillOrder(ctx.Request.Context(), orderNo, req.BizUserId)
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
			group.POST("/skus", r.CreateSku)
			group.PUT("/skus/:id", r.EditSku)
			group.GET("/skus/:id", r.GetSkuDetail)
			group.GET("/skus", r.ListSkus)
			group.POST("/skus/:id/listing", r.ListingSku)
			group.POST("/skus/:id/delisting", r.DelistingSku)
			group.DELETE("/skus/:id", r.DeleteSku)

			// 订单管理接口
			group.GET("/orders", r.ListOrders)
			group.GET("/orders/:order_no", r.GetOrderDetail)
			group.POST("/orders/:order_no/refund", r.RefundOrder)
			group.POST("/orders/:order_no/fulfill", r.FulfillOrder)
		}
	}
}
