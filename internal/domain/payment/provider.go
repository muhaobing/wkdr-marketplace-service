package payment

import (
	"fmt"
	"os"
	"strings"

	"wdkr-marketplace-service/internal/common/config"
	"wdkr-marketplace-service/internal/domain/payment/channel"
	"wdkr-marketplace-service/internal/domain/payment/channel/wechat"
	"wdkr-marketplace-service/internal/domain/payment/repo"
)

// NewWechatPayConfig 从配置文件获取微信支付配置
func NewWechatPayConfig() *wechat.WechatPayConfig {
	cfg := config.GetWechatPayConfig()
	if cfg == nil || cfg.AppID == "" {
		return wechat.NewMockConfig()
	}

	privateKey := cfg.PrivateKey
	if cfg.PrivateKeyPath != "" {
		data, err := os.ReadFile(cfg.PrivateKeyPath)
		if err != nil {
			fmt.Printf("[WARN] failed to read private key file: %s, err: %v\n", cfg.PrivateKeyPath, err)
		} else {
			privateKey = strings.TrimSpace(string(data))
		}
	}

	return &wechat.WechatPayConfig{
		AppID:           cfg.AppID,
		MchID:           cfg.MchID,
		APIKey:          cfg.APIKey,
		SerialNo:        cfg.SerialNo,
		PrivateKey:      privateKey,
		NotifyURL:       cfg.NotifyURL,
		RefundNotifyURL: cfg.RefundNotifyURL,
	}
}

// NewWechatPayChannel 创建微信支付渠道
func NewWechatPayChannel(config *wechat.WechatPayConfig) *wechat.WechatPayChannel {
	return wechat.NewWechatPayChannel(config)
}

// ProvidePaymentChannels 提供所有支付渠道列表
// 目前仅支持微信支付
func ProvidePaymentChannels(wechatChannel *wechat.WechatPayChannel) []channel.PaymentChannel {
	return []channel.PaymentChannel{wechatChannel}
}

// ProvidePaymentService 提供 PaymentService
// 内部初始化所有支付渠道
func ProvidePaymentService(paymentRepo repo.PaymentRepo, channels []channel.PaymentChannel) PaymentService {
	return NewPaymentService(paymentRepo, channels...)
}
