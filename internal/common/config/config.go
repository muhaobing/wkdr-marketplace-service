package config

import (
	"strings"
	"time"

	"github.com/muhaobing/std-go/restserver/config"
)

// DefaultJWTExpirationSeconds JWT 默认最大存活时间（秒），与 Postman 脚本默认 ttl 一致
const DefaultJWTExpirationSeconds uint32 = 300

// Conf 应用配置
type Conf struct {
	Auth                       AuthConfig        `yaml:"auth"`         // 用户 session（AES+Redis）鉴权
	JWT                        JWTConfig         `yaml:"jwt"`          // OpenAPI 等系统间 JWT 鉴权（与 auth 独立）
	CallbackJWT                CallbackJWTConfig `yaml:"callback_jwt"` // SKU 履约回调 JWT（按 biz_code 配置 secret）
	WechatPay                  WechatPayConfig   `yaml:"wechat_pay"`   // 微信支付配置
	EcoinUnitPrice             float32           `yaml:"ecoin_unit_price"`
	EcoinExpireSeconds         uint32            `yaml:"ecoin_expire_seconds"`          // 已废弃：保留键名兼容旧配置；积分过期由域服务按自然月计算，不再读取该秒数
	EcoinIdempotencyTTLSeconds uint32            `yaml:"ecoin_idempotency_ttl_seconds"` // 积分加减幂等 Redis 键 TTL（秒），0 表示默认 12 小时
}

// AuthConfig 前台/运营端 session 鉴权
type AuthConfig struct {
	AesKey     string   `yaml:"aes_key"`     // AES加密密钥
	Expiration uint32   `yaml:"expiration"`  // session过期时间（秒）
	Whitelist  []string `yaml:"whitelist"`   // 不需要 session 鉴权的路径白名单
	AdminPaths []string `yaml:"admin_paths"` // 需要管理员权限的路径前缀
}

// JWTConfig 系统间 JWT：请求体仅含 jwt 字段，业务参数在 payload；account 与 secret 成对，验签时按 payload.account 选用对应 secret
type JWTConfig struct {
	Clients           []JWTClient `yaml:"clients"`            // 多组调用方，每组 account + secret 一一对应
	Paths             []string    `yaml:"paths"`              // 需 JWT 的路径项：无尾斜杠仅匹配自身或「自身/子路径」；尾斜杠则前缀匹配（与 session 白名单规则一致）
	ExpirationSeconds uint32      `yaml:"expiration_seconds"` // JWT 最大存活时间（秒），0 表示使用 DefaultJWTExpirationSeconds；须含 iat、exp，且 exp-iat 不得超过此值
}

// JWTClient 单个 JWT 调用方凭据
type JWTClient struct {
	Account string `yaml:"account"`
	Secret  string `yaml:"secret"`
}

// CallbackJWTConfig SKU 履约回调 JWT：每个 biz_code 一组 secret
type CallbackJWTConfig struct {
	Clients           []CallbackJWTClient `yaml:"clients"`
	ExpirationSeconds uint32              `yaml:"expiration_seconds"` // 0 表示默认 300 秒
}

// CallbackJWTClient 单个业务线回调凭据
type CallbackJWTClient struct {
	BizCode string `yaml:"biz_code"`
	Secret  string `yaml:"secret"`
}

// SecretForAccount 按 payload 中的 account 查找对应 HS256 密钥；无匹配返回 false
func (j *JWTConfig) SecretForAccount(account string) (string, bool) {
	account = strings.TrimSpace(account)
	if account == "" {
		return "", false
	}
	for _, c := range j.Clients {
		if strings.TrimSpace(c.Account) != account {
			continue
		}
		sec := strings.TrimSpace(c.Secret)
		if sec == "" {
			return "", false
		}
		return c.Secret, true
	}
	return "", false
}

// EffectiveJWTExpirationSeconds 返回 JWT 最大存活秒数（用于校验 iat/exp 窗口）
func (j *JWTConfig) EffectiveJWTExpirationSeconds() uint32 {
	if j.ExpirationSeconds == 0 {
		return DefaultJWTExpirationSeconds
	}
	return j.ExpirationSeconds
}

// HasJWTClients 是否配置了至少一组有效的 account+secret
func (j *JWTConfig) HasJWTClients() bool {
	for _, c := range j.Clients {
		if strings.TrimSpace(c.Account) != "" && strings.TrimSpace(c.Secret) != "" {
			return true
		}
	}
	return false
}

// SecretForBizCode 按 biz_code 查找履约回调 JWT 密钥
func (j *CallbackJWTConfig) SecretForBizCode(bizCode string) (string, bool) {
	bizCode = strings.TrimSpace(bizCode)
	if bizCode == "" {
		return "", false
	}
	for _, c := range j.Clients {
		if strings.TrimSpace(c.BizCode) != bizCode {
			continue
		}
		sec := strings.TrimSpace(c.Secret)
		if sec == "" {
			return "", false
		}
		return c.Secret, true
	}
	return "", false
}

// EffectiveExpirationSeconds 返回回调 JWT 最大存活秒数
func (j *CallbackJWTConfig) EffectiveExpirationSeconds() uint32 {
	if j.ExpirationSeconds == 0 {
		return DefaultJWTExpirationSeconds
	}
	return j.ExpirationSeconds
}

// WechatPayConfig 微信支付配置
type WechatPayConfig struct {
	AppID           string `yaml:"app_id"`            // 微信AppID（公众号/小程序）
	MchID           string `yaml:"mch_id"`            // 微信支付商户号
	APIKey          string `yaml:"api_key"`           // APIv3密钥（32字节）
	SerialNo        string `yaml:"serial_no"`         // 商户API证书序列号
	PrivateKey      string `yaml:"private_key"`       // 商户API私钥（PEM内容，与 private_key_path 二选一）
	PrivateKeyPath  string `yaml:"private_key_path"`  // 商户API私钥文件路径（优先级高于 private_key）
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

// GetEcoinIdempotencyTTL 积分加减幂等键过期时间（默认 12 小时）
func GetEcoinIdempotencyTTL() time.Duration {
	if globalConf == nil || globalConf.EcoinIdempotencyTTLSeconds == 0 {
		return 12 * time.Hour
	}
	return time.Duration(globalConf.EcoinIdempotencyTTLSeconds) * time.Second
}
