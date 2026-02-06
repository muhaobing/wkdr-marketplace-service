package order

import (
	"context"

	ordermodel "wdkr-marketplace-service/internal/domain/order/order_model"
)

// SkuOrderItem SKU订单项
type SkuOrderItem struct {
	SkuId    uint64 `json:"sku_id"`   // SKU ID
	Quantity int    `json:"quantity"` // 购买数量
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	UserId          uint64          `json:"user_id"`   // 用户ID
	SkuItems        []*SkuOrderItem `json:"sku_items"` // SKU列表
	PayType         string          `json:"pay_type"`  // 支付类型：ecoin/money
	Remark          string          `json:"remark"`    // 备注
	IsEcoinRecharge bool            `json:"is_ecoin_recharge"`
	EcoinStock      int             `json:"ecoin_stock"`
}

// CreateOrderResponse 创建订单响应
type CreateOrderResponse struct {
	Order *ordermodel.Order `json:"order"` // 订单信息
}

// PayOrderRequest 支付订单请求（货币支付）
type PayOrderRequest struct {
	OrderNo   string `json:"order_no"`   // 订单号
	Channel   string `json:"channel"`    // 支付渠道：wechat/alipay
	PayMethod string `json:"pay_method"` // 支付方式：native/jsapi/h5
	ClientIP  string `json:"client_ip"`  // 客户端IP（H5支付需要）
	OpenId    string `json:"open_id"`    // 用户OpenID（JSAPI支付需要）
}

// PayOrderResponse 支付订单响应
type PayOrderResponse struct {
	OrderNo   string `json:"order_no"`             // 订单号
	CodeUrl   string `json:"code_url,omitempty"`   // Native支付二维码URL
	H5Url     string `json:"h5_url,omitempty"`     // H5支付跳转URL
	PrepayId  string `json:"prepay_id,omitempty"`  // JSAPI预支付ID
	AppId     string `json:"app_id,omitempty"`     // JSAPI AppID
	TimeStamp string `json:"time_stamp,omitempty"` // JSAPI 时间戳
	NonceStr  string `json:"nonce_str,omitempty"`  // JSAPI 随机串
	Package   string `json:"package,omitempty"`    // JSAPI Package
	SignType  string `json:"sign_type,omitempty"`  // JSAPI 签名类型
	PaySign   string `json:"pay_sign,omitempty"`   // JSAPI 签名
}

// CancelOrderRequest 取消订单请求
type CancelOrderRequest struct {
	OrderNo string `json:"order_no"` // 订单号
	Reason  string `json:"reason"`   // 取消原因
}

// ListOrdersRequest 订单列表请求
type ListOrdersRequest struct {
	UserId uint64 `json:"user_id"` // 用户ID
	Status *uint8 `json:"status"`  // 订单状态过滤（可选）
	Offset int    `json:"offset"`  // 偏移量
	Limit  int    `json:"limit"`   // 每页数量
}

// ListOrdersResponse 订单列表响应
type ListOrdersResponse struct {
	Total int64               `json:"total"` // 总数
	List  []*ordermodel.Order `json:"list"`  // 订单列表
}

// RefundOrderRequest 退款订单请求
type RefundOrderRequest struct {
	OrderNo string `json:"order_no"` // 订单号
	Reason  string `json:"reason"`   // 退款原因
}

// OrderService 订单服务接口
type OrderService interface {
	// CreateOrder 创建订单
	// 积分支付会立即扣除积分并完成支付
	// 货币支付会创建待支付订单
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)

	// GetOrderByOrderNo 根据订单号获取订单详情
	// 包含订单明细项
	GetOrderByOrderNo(ctx context.Context, orderNo string) (*ordermodel.Order, error)

	// PayOrder 支付订单（货币支付）
	// 调用支付模块创建支付订单，返回支付凭证
	PayOrder(ctx context.Context, req *PayOrderRequest) (*PayOrderResponse, error)

	// CancelOrder 取消订单
	// 仅允许取消待支付的订单
	// 积分支付订单无法取消（已支付）
	CancelOrder(ctx context.Context, req *CancelOrderRequest) error

	// HandlePaymentSuccess 处理支付成功回调
	// 更新订单状态为已支付，并触发履约
	HandlePaymentSuccess(ctx context.Context, orderNo string, payTime uint32) error

	// FulfillOrder 履约订单
	// 调用SKU的履约接口，为用户发放商品
	FulfillOrder(ctx context.Context, orderNo string, bizUserId string) error

	// RefundOrder 退款订单
	// 积分支付退还积分，货币支付调用支付模块退款
	RefundOrder(ctx context.Context, req *RefundOrderRequest) error

	// ListOrders 获取订单列表
	ListOrders(ctx context.Context, req *ListOrdersRequest) (*ListOrdersResponse, error)

	// SyncOrderStatus 同步订单状态
	// 主动查询支付状态并更新订单
	SyncOrderStatus(ctx context.Context, orderNo string) (*ordermodel.Order, error)
}
