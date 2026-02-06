package config

import (
	"github.com/muhaobing-eng/std-go/restserver/config"
)

// Conf 应用配置
type Conf struct {
	Auth      AuthConfig      `yaml:"config.auth"`       // 鉴权配置
	WechatPay WechatPayConfig `yaml:"config.wechat_pay"` // 微信支付配置
}

type AuthConfig struct {
	AesKey     string `yaml:"aes_key"`
	Expiration uint32 `yaml:"expiration"`
}

// WechatPayConfig 微信支付配置
type WechatPayConfig struct {
	AppID           string `yaml:"app_id"`            // 微信AppID（公众号/小程序）
	MchID           string `yaml:"mch_id"`            // 微信支付商户号
	APIKey          string `yaml:"api_key"`           // APIv3密钥（32字节）
	SerialNo        string `yaml:"serial_no"`         // 商户API证书序列号
	PrivateKey      string `yaml:"private_key"`       // 商户API私钥（PEM格式）
	NotifyURL       string `yaml:"notify_url"`        // 支付结果回调通知地址
	RefundNotifyURL string `yaml:"refund_notify_url"` // 退款结果回调通知地址（可选）
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

func GetConf() *Conf {
	return globalConf
}

// GetWechatPayConfig 获取微信支付配置
func GetWechatPayConfig() *WechatPayConfig {
	if globalConf == nil {
		return nil
	}
	return &globalConf.WechatPay
}
