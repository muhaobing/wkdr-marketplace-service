package payment_model

import (
	"database/sql/driver"
	"errors"

	jsoniter "github.com/json-iterator/go"
)

const (
	PaymentOrderTabName = "payment_order_tab"

	// 支付订单状态
	PaymentStatusPending  uint8 = 0 // 待支付
	PaymentStatusPaid     uint8 = 1 // 已支付
	PaymentStatusClosed   uint8 = 2 // 已关闭
	PaymentStatusRefunded uint8 = 3 // 已退款

	// 业务类型
	BizTypeRecharge string = "recharge" // 充值积分
	BizTypePurchase string = "purchase" // 购买商品

	// 支付渠道
	ChannelWechat string = "wechat" // 微信支付
	ChannelAlipay string = "alipay" // 支付宝（预留）

	// 支付方式
	PayMethodNative string = "native" // Native支付（扫码）
	PayMethodJSAPI  string = "jsapi"  // JSAPI支付（微信内/小程序）
	PayMethodH5     string = "h5"     // H5支付（手机浏览器）
)

// PaymentOrder 支付订单
type PaymentOrder struct {
	Id             uint64    `gorm:"column:id;primaryKey" json:"id"`
	OrderNo        string    `gorm:"column:order_no" json:"order_no"`                 // 支付订单号（唯一）
	BizOrderNo     string    `gorm:"column:biz_order_no" json:"biz_order_no"`         // 业务订单号
	BizType        string    `gorm:"column:biz_type" json:"biz_type"`                 // 业务类型：recharge/purchase
	UserId         uint64    `gorm:"column:user_id" json:"user_id"`                   // 用户ID
	Channel        string    `gorm:"column:channel" json:"channel"`                   // 支付渠道：wechat/alipay
	PayMethod      string    `gorm:"column:pay_method" json:"pay_method"`             // 支付方式：native/jsapi/h5
	Amount         int64     `gorm:"column:amount" json:"amount"`                     // 支付金额（分）
	Status         uint8     `gorm:"column:status" json:"status"`                     // 状态
	ChannelOrderNo string    `gorm:"column:channel_order_no" json:"channel_order_no"` // 渠道订单号
	PayTime        uint32    `gorm:"column:pay_time" json:"pay_time"`                 // 支付时间
	ExpireTime     uint32    `gorm:"column:expire_time" json:"expire_time"`           // 过期时间
	NotifyUrl      string    `gorm:"column:notify_url" json:"notify_url"`             // 回调通知URL
	Extra          ExtraData `gorm:"column:extra" json:"extra"`                       // 扩展信息
	Ctime          uint32    `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime          uint32    `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

func (p *PaymentOrder) TableName() string {
	return PaymentOrderTabName
}

// IsPending 是否待支付
func (p *PaymentOrder) IsPending() bool {
	return p.Status == PaymentStatusPending
}

// IsPaid 是否已支付
func (p *PaymentOrder) IsPaid() bool {
	return p.Status == PaymentStatusPaid
}

// IsClosed 是否已关闭
func (p *PaymentOrder) IsClosed() bool {
	return p.Status == PaymentStatusClosed
}

// IsRefunded 是否已退款
func (p *PaymentOrder) IsRefunded() bool {
	return p.Status == PaymentStatusRefunded
}

// ExtraData 扩展数据
type ExtraData map[string]interface{}

func (e ExtraData) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type for ExtraData")
	}
	if len(bytes) == 0 {
		return nil
	}
	return jsoniter.Unmarshal(bytes, &e)
}

func (e ExtraData) Value() (driver.Value, error) {
	if e == nil {
		return nil, nil
	}
	return jsoniter.Marshal(e)
}
