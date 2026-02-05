package wechat

// ============================================================
// 微信支付配置
// TODO: 请替换以下所有 Mock 值为真实的微信支付配置
// ============================================================

// WechatPayConfig 微信支付配置
type WechatPayConfig struct {
	// AppID 微信AppID（公众号/小程序）
	AppID string

	// MchID 微信支付商户号
	MchID string

	// APIKey APIv3密钥（32字节）
	APIKey string

	// SerialNo 商户API证书序列号
	SerialNo string

	// PrivateKey 商户API私钥（PEM格式）
	PrivateKey string

	// NotifyURL 支付结果回调通知地址
	NotifyURL string

	// RefundNotifyURL 退款结果回调通知地址（可选）
	RefundNotifyURL string
}

// ============================================================
// Mock配置 - 仅用于开发测试
// TODO: 部署前请替换为真实配置！
// ============================================================

// NewMockConfig 创建Mock配置（仅用于开发测试）
func NewMockConfig() *WechatPayConfig {
	return &WechatPayConfig{
		// TODO: 替换为真实的微信AppID
		AppID: "wx_mock_appid_replace_me",

		// TODO: 替换为真实的商户号
		MchID: "1234567890",

		// TODO: 替换为真实的APIv3密钥（32字节）
		APIKey: "mock_api_key_32_bytes_replace_me",

		// TODO: 替换为真实的证书序列号
		SerialNo: "MOCK_SERIAL_NO_REPLACE_ME",

		// TODO: 替换为真实的商户私钥
		PrivateKey: `-----BEGIN PRIVATE KEY-----
MOCK_PRIVATE_KEY_REPLACE_ME
This is a placeholder for the merchant's private key.
Please replace with actual PEM formatted private key.
-----END PRIVATE KEY-----`,

		// TODO: 替换为真实的回调地址
		NotifyURL: "https://your-domain.com/api/payment/notify/wechat",

		// TODO: 替换为真实的退款回调地址（可选）
		RefundNotifyURL: "https://your-domain.com/api/payment/refund/notify/wechat",
	}
}

// 微信支付API地址
const (
	// 沙箱环境
	SandboxBaseURL = "https://api.mch.weixin.qq.com/sandboxnew"

	// 正式环境
	ProductionBaseURL = "https://api.mch.weixin.qq.com"

	// Native支付下单
	UnifiedOrderURL = "/v3/pay/transactions/native"

	// JSAPI支付下单
	JSAPIOrderURL = "/v3/pay/transactions/jsapi"

	// H5支付下单
	H5OrderURL = "/v3/pay/transactions/h5"

	// 查询订单
	QueryOrderURL = "/v3/pay/transactions/out-trade-no/%s"

	// 关闭订单
	CloseOrderURL = "/v3/pay/transactions/out-trade-no/%s/close"

	// 申请退款
	RefundURL = "/v3/refund/domestic/refunds"

	// 查询退款
	QueryRefundURL = "/v3/refund/domestic/refunds/%s"
)
