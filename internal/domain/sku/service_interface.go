package sku

import (
	"context"

	skumodel "wdkr-marketplace-service/internal/domain/sku/sku_model"
)

// CreateSkuRequest 创建商品请求
type CreateSkuRequest struct {
	BizCode        string  `json:"biz_code"`        // 业务编码
	SkuCode        string  `json:"sku_code"`        // 商品代码
	SkuName        string  `json:"sku_name"`        // 商品名称
	SkuAvatar      string  `json:"sku_avatar"`      // 商品图标
	SkuDesc        string  `json:"sku_desc"`        // 商品描述
	Cost           float32 `json:"cost"`            // 商品售价(积分)
	DeliveryMethod string  `json:"delivery_method"` // 履约回调接口
}

// EditSkuRequest 编辑商品请求
type EditSkuRequest struct {
	Id             uint64  `json:"id"`              // 商品ID
	SkuName        string  `json:"sku_name"`        // 商品名称
	SkuAvatar      string  `json:"sku_avatar"`      // 商品图标
	SkuDesc        string  `json:"sku_desc"`        // 商品描述
	Cost           float32 `json:"cost"`            // 商品售价(积分)
	DeliveryMethod string  `json:"delivery_method"` // 履约回调接口
}

// ListSkuRequest 商品列表请求
type ListSkuRequest struct {
	BizCode string `json:"biz_code"` // 业务编码
	Status  *uint8 `json:"status"`   // 上架状态过滤（可选）
	Offset  int    `json:"offset"`   // 偏移量
	Limit   int    `json:"limit"`    // 每页数量
}

// ListSkuResponse 商品列表响应
type ListSkuResponse struct {
	Total int64           `json:"total"` // 总数
	List  []*skumodel.Sku `json:"list"`  // 商品列表
}

// FulfillSkuRequest 商品履约请求
type FulfillSkuRequest struct {
	SkuId     uint64 `json:"sku_id"`      // 商品ID
	BizUserId string `json:"biz_user_id"` // 业务用户ID
}

// FulfillSkuResponse 商品履约响应
type FulfillSkuResponse struct {
	Success bool   `json:"success"` // 履约是否成功
	Message string `json:"message"` // 履约结果消息
}

// SkuService SKU服务接口
type SkuService interface {
	// CreateSku 创建商品（默认未上架状态）
	CreateSku(ctx context.Context, req *CreateSkuRequest) (*skumodel.Sku, error)

	// GetSkuById 根据ID获取商品
	GetSkuById(ctx context.Context, id uint64) (*skumodel.Sku, error)

	// GetSkuByCode 根据业务编码和商品编码获取商品
	GetSkuByCode(ctx context.Context, bizCode, skuCode string) (*skumodel.Sku, error)

	// EditSku 编辑商品信息
	EditSku(ctx context.Context, req *EditSkuRequest) (*skumodel.Sku, error)

	// ListSku 上架商品（将商品状态改为已上架）
	ListingSku(ctx context.Context, id uint64) error

	// DelistSku 下架商品（将商品状态改为未上架）
	DelistingSku(ctx context.Context, id uint64) error

	// ListSkus 获取商品列表
	ListSkus(ctx context.Context, req *ListSkuRequest) (*ListSkuResponse, error)

	// DeleteSku 删除商品（仅允许删除未上架的商品）
	DeleteSku(ctx context.Context, id uint64) error

	// FulfillSku 商品履约（调用商品的履约回调接口）
	FulfillSku(ctx context.Context, req *FulfillSkuRequest) (*FulfillSkuResponse, error)
}
