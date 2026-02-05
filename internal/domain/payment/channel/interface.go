package channel

import (
	"context"
)

// PaymentChannel 支付渠道抽象接口
// 所有支付渠道（微信、支付宝等）都需要实现此接口
type PaymentChannel interface {
	// GetChannelCode 获取渠道代码
	GetChannelCode() string

	// CreatePayment 创建支付
	// 返回支付凭证（如二维码URL、JSAPI参数等）
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error)

	// QueryPayment 查询支付状态
	QueryPayment(ctx context.Context, orderNo string) (*QueryPaymentResponse, error)

	// ClosePayment 关闭支付订单
	ClosePayment(ctx context.Context, orderNo string) error

	// Refund 申请退款
	Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error)

	// QueryRefund 查询退款状态
	QueryRefund(ctx context.Context, refundNo string) (*QueryRefundResponse, error)

	// VerifyNotify 验证回调签名并解析数据
	VerifyNotify(ctx context.Context, data []byte) (*PaymentNotify, error)
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	OrderNo     string `json:"order_no"`    // 支付订单号
	Amount      int64  `json:"amount"`      // 支付金额（分）
	Description string `json:"description"` // 商品描述
	PayMethod   string `json:"pay_method"`  // 支付方式：native/jsapi/h5
	NotifyUrl   string `json:"notify_url"`  // 回调通知URL
	ExpireTime  int64  `json:"expire_time"` // 过期时间戳
	ClientIP    string `json:"client_ip"`   // 客户端IP（H5支付需要）
	OpenId      string `json:"open_id"`     // 用户OpenID（JSAPI支付需要）
}

// CreatePaymentResponse 创建支付响应
type CreatePaymentResponse struct {
	// Native支付返回二维码URL
	CodeUrl string `json:"code_url,omitempty"`

	// H5支付返回跳转URL
	H5Url string `json:"h5_url,omitempty"`

	// JSAPI支付返回的参数（用于前端调起支付）
	PrepayId  string `json:"prepay_id,omitempty"`
	AppId     string `json:"app_id,omitempty"`
	TimeStamp string `json:"time_stamp,omitempty"`
	NonceStr  string `json:"nonce_str,omitempty"`
	Package   string `json:"package,omitempty"`
	SignType  string `json:"sign_type,omitempty"`
	PaySign   string `json:"pay_sign,omitempty"`
}

// QueryPaymentResponse 查询支付响应
type QueryPaymentResponse struct {
	OrderNo        string `json:"order_no"`         // 支付订单号
	ChannelOrderNo string `json:"channel_order_no"` // 渠道订单号
	Status         string `json:"status"`           // 状态：SUCCESS/NOTPAY/CLOSED/REFUND
	PayTime        int64  `json:"pay_time"`         // 支付时间戳
	Amount         int64  `json:"amount"`           // 支付金额（分）
}

// 支付状态常量
const (
	PayStatusSuccess string = "SUCCESS" // 支付成功
	PayStatusNotPay  string = "NOTPAY"  // 未支付
	PayStatusClosed  string = "CLOSED"  // 已关闭
	PayStatusRefund  string = "REFUND"  // 已退款
)

// RefundRequest 退款请求
type RefundRequest struct {
	OrderNo   string `json:"order_no"`   // 原支付订单号
	RefundNo  string `json:"refund_no"`  // 退款单号
	TotalFee  int64  `json:"total_fee"`  // 原订单金额（分）
	RefundFee int64  `json:"refund_fee"` // 退款金额（分）
	Reason    string `json:"reason"`     // 退款原因
	NotifyUrl string `json:"notify_url"` // 退款回调URL（可选）
}

// RefundResponse 退款响应
type RefundResponse struct {
	RefundNo        string `json:"refund_no"`         // 退款单号
	ChannelRefundNo string `json:"channel_refund_no"` // 渠道退款号
	Status          string `json:"status"`            // 状态：PROCESSING/SUCCESS/FAILED
}

// 退款状态常量
const (
	RefundStatusProcessing string = "PROCESSING" // 处理中
	RefundStatusSuccess    string = "SUCCESS"    // 退款成功
	RefundStatusFailed     string = "FAILED"     // 退款失败
)

// QueryRefundResponse 查询退款响应
type QueryRefundResponse struct {
	RefundNo        string `json:"refund_no"`         // 退款单号
	ChannelRefundNo string `json:"channel_refund_no"` // 渠道退款号
	Status          string `json:"status"`            // 状态
	RefundFee       int64  `json:"refund_fee"`        // 退款金额
}

// PaymentNotify 支付回调通知
type PaymentNotify struct {
	OrderNo        string `json:"order_no"`         // 支付订单号
	ChannelOrderNo string `json:"channel_order_no"` // 渠道订单号
	Status         string `json:"status"`           // 状态：SUCCESS/FAIL
	PayTime        int64  `json:"pay_time"`         // 支付时间戳
	Amount         int64  `json:"amount"`           // 支付金额
}
