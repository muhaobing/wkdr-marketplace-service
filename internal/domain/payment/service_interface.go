package payment

import (
	"context"

	"wdkr-marketplace-service/internal/domain/payment/payment_model"
)

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	BizOrderNo    string `json:"biz_order_no"`   // 业务订单号
	BizType       string `json:"biz_type"`       // 业务类型：recharge/purchase
	UserId        uint64 `json:"user_id"`        // 用户ID
	Channel       string `json:"channel"`        // 支付渠道：wechat/alipay
	PayMethod     string `json:"pay_method"`     // 支付方式：native/jsapi/h5
	Amount        int64  `json:"amount"`         // 支付金额（分）
	Description   string `json:"description"`    // 商品描述
	ClientIP      string `json:"client_ip"`      // 客户端IP（H5支付需要）
	OpenId        string `json:"open_id"`        // 用户OpenID（JSAPI支付需要）
	ExpireMinutes int    `json:"expire_minutes"` // 过期时间（分钟），默认30分钟
}

// CreatePaymentResponse 创建支付响应
type CreatePaymentResponse struct {
	OrderNo   string `json:"order_no"`             // 支付订单号
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

// RefundRequest 退款请求
type RefundRequest struct {
	OrderNo   string `json:"order_no"`   // 支付订单号
	RefundFee int64  `json:"refund_fee"` // 退款金额（分）
	Reason    string `json:"reason"`     // 退款原因
}

// ListPaymentOrdersRequest 获取支付订单列表请求
type ListPaymentOrdersRequest struct {
	UserId uint64 `json:"user_id"` // 用户ID
	Offset int    `json:"offset"`  // 偏移量
	Limit  int    `json:"limit"`   // 每页数量
}

// ListPaymentOrdersResponse 获取支付订单列表响应
type ListPaymentOrdersResponse struct {
	Total int64                         `json:"total"` // 总数
	List  []*payment_model.PaymentOrder `json:"list"`  // 订单列表
}

// PaymentService 支付服务接口
type PaymentService interface {
	// CreatePayment 创建支付订单
	// 返回支付凭证（二维码URL/H5链接/JSAPI参数）
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error)

	// HandleNotify 处理支付回调
	// channel: 支付渠道（wechat/alipay）
	// data: 回调原始数据
	HandleNotify(ctx context.Context, channel string, data []byte) error

	// QueryPayment 查询支付状态
	// 返回最新的支付订单信息
	QueryPayment(ctx context.Context, orderNo string) (*payment_model.PaymentOrder, error)

	// SyncPaymentStatus 同步支付状态
	// 主动向支付渠道查询并更新订单状态
	SyncPaymentStatus(ctx context.Context, orderNo string) (*payment_model.PaymentOrder, error)

	// ClosePayment 关闭支付订单
	// 仅允许关闭待支付的订单
	ClosePayment(ctx context.Context, orderNo string) error

	// Refund 申请退款
	// 仅允许对已支付的订单申请退款
	Refund(ctx context.Context, req *RefundRequest) (*payment_model.PaymentRefund, error)

	// QueryRefund 查询退款状态
	QueryRefund(ctx context.Context, refundNo string) (*payment_model.PaymentRefund, error)

	// SyncRefundStatus 同步退款状态
	// 主动向支付渠道查询并更新退款状态
	SyncRefundStatus(ctx context.Context, refundNo string) (*payment_model.PaymentRefund, error)

	// ListPaymentOrders 获取支付订单列表
	ListPaymentOrders(ctx context.Context, req *ListPaymentOrdersRequest) (*ListPaymentOrdersResponse, error)

	// GetPaymentOrderByBizOrderNo 根据业务订单号获取支付订单
	GetPaymentOrderByBizOrderNo(ctx context.Context, bizOrderNo string) (*payment_model.PaymentOrder, error)
}
