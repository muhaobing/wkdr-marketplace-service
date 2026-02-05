package payment_model

const (
	PaymentRefundTabName = "payment_refund_tab"

	// 退款状态
	RefundStatusProcessing uint8 = 0 // 处理中
	RefundStatusSuccess    uint8 = 1 // 退款成功
	RefundStatusFailed     uint8 = 2 // 退款失败
)

// PaymentRefund 退款记录
type PaymentRefund struct {
	Id              uint64 `gorm:"column:id;primaryKey" json:"id"`
	RefundNo        string `gorm:"column:refund_no" json:"refund_no"`                 // 退款单号（唯一）
	OrderNo         string `gorm:"column:order_no" json:"order_no"`                   // 关联支付订单号
	Channel         string `gorm:"column:channel" json:"channel"`                     // 支付渠道
	Amount          int64  `gorm:"column:amount" json:"amount"`                       // 退款金额（分）
	Reason          string `gorm:"column:reason" json:"reason"`                       // 退款原因
	Status          uint8  `gorm:"column:status" json:"status"`                       // 状态
	ChannelRefundNo string `gorm:"column:channel_refund_no" json:"channel_refund_no"` // 渠道退款号
	Ctime           uint32 `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime           uint32 `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

func (p *PaymentRefund) TableName() string {
	return PaymentRefundTabName
}

// IsProcessing 是否处理中
func (p *PaymentRefund) IsProcessing() bool {
	return p.Status == RefundStatusProcessing
}

// IsSuccess 是否成功
func (p *PaymentRefund) IsSuccess() bool {
	return p.Status == RefundStatusSuccess
}

// IsFailed 是否失败
func (p *PaymentRefund) IsFailed() bool {
	return p.Status == RefundStatusFailed
}
