package config

import (
	"github.com/muhaobing-eng/std-go/restserver/config"
)

// Conf 应用配置
type Conf struct {
	WechatPay WechatPayConfig `json:"config.wechat_pay"` // 微信支付配置
}

// WechatPayConfig 微信支付配置
type WechatPayConfig struct {
	AppID           string `json:"app_id"`            // 微信AppID（公众号/小程序）
	MchID           string `json:"mch_id"`            // 微信支付商户号
	APIKey          string `json:"api_key"`           // APIv3密钥（32字节）
	SerialNo        string `json:"serial_no"`         // 商户API证书序列号
	PrivateKey      string `json:"private_key"`       // 商户API私钥（PEM格式）
	NotifyURL       string `json:"notify_url"`        // 支付结果回调通知地址
	RefundNotifyURL string `json:"refund_notify_url"` // 退款结果回调通知地址（可选）
}

var globalConf *Conf

func Init() error {
	conf := new(Conf)
	if err := config.LoadWithPrefix("config", conf); err != nil {
		return err
	}
	globalConf = conf
	return nil
}

func Get() *Conf {
	return globalConf
}

// GetWechatPayConfig 获取微信支付配置
func GetWechatPayConfig() *WechatPayConfig {
	if globalConf == nil {
		return nil
	}
	return &globalConf.WechatPay
}
