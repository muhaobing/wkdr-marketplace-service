package order_model

const (
	OrderTabName = "order_tab"

	// 订单状态
	OrderStatusPending   uint8 = 0 // 待支付
	OrderStatusPaid      uint8 = 1 // 已支付
	OrderStatusFulfilled uint8 = 2 // 已履约（已完成）
	OrderStatusCancelled uint8 = 3 // 已取消
	OrderStatusRefunded  uint8 = 4 // 已退款

	// 支付类型
	PayTypeEcoin string = "ecoin" // 积分支付
	PayTypeMoney string = "money" // 货币支付
)

// Order 订单（支持跨业务下单）
type Order struct {
	Id             uint64  `gorm:"column:id;primaryKey" json:"id"`
	OrderNo        string  `gorm:"column:order_no" json:"order_no"`                 // 订单号（唯一）
	UserId         uint64  `gorm:"column:user_id" json:"user_id"`                   // 用户ID
	ItemCount      int     `gorm:"column:item_count" json:"item_count"`             // 商品种类数量
	TotalQuantity  int     `gorm:"column:total_quantity" json:"total_quantity"`     // 商品总数量
	OriginalAmount float32 `gorm:"column:original_amount" json:"original_amount"`   // 原价
	PayAmount      float32 `gorm:"column:pay_amount" json:"pay_amount"`             // 实付金额
	PayType        string  `gorm:"column:pay_type" json:"pay_type"`                 // 支付类型：ecoin/money
	PaymentOrderNo string  `gorm:"column:payment_order_no" json:"payment_order_no"` // 支付订单号（货币支付时使用）
	Status         uint8   `gorm:"column:status" json:"status"`                     // 订单状态
	PayTime        uint32  `gorm:"column:pay_time" json:"pay_time"`                 // 支付时间
	FulfillTime    uint32  `gorm:"column:fulfill_time" json:"fulfill_time"`         // 履约时间
	CancelTime     uint32  `gorm:"column:cancel_time" json:"cancel_time"`           // 取消时间
	CancelReason   string  `gorm:"column:cancel_reason" json:"cancel_reason"`       // 取消原因
	Remark         string  `gorm:"column:remark" json:"remark"`                     // 备注
	Ctime          uint32  `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime          uint32  `gorm:"column:mtime;autoUpdateTime" json:"mtime"`

	// 非数据库字段，用于关联查询
	Items []*OrderItem `gorm:"-" json:"items,omitempty"`
}

func (o *Order) TableName() string {
	return OrderTabName
}

// IsPending 是否待支付
func (o *Order) IsPending() bool {
	return o.Status == OrderStatusPending
}

// IsPaid 是否已支付
func (o *Order) IsPaid() bool {
	return o.Status == OrderStatusPaid
}

// IsFulfilled 是否已履约
func (o *Order) IsFulfilled() bool {
	return o.Status == OrderStatusFulfilled
}

// IsCancelled 是否已取消
func (o *Order) IsCancelled() bool {
	return o.Status == OrderStatusCancelled
}

// IsRefunded 是否已退款
func (o *Order) IsRefunded() bool {
	return o.Status == OrderStatusRefunded
}

// CanCancel 是否可以取消
func (o *Order) CanCancel() bool {
	return o.Status == OrderStatusPending
}

// CanPay 是否可以支付
func (o *Order) CanPay() bool {
	return o.Status == OrderStatusPending
}

// CanRefund 是否可以退款
func (o *Order) CanRefund() bool {
	return o.Status == OrderStatusPaid || o.Status == OrderStatusFulfilled
}

// IsEcoinPay 是否为积分支付
func (o *Order) IsEcoinPay() bool {
	return o.PayType == PayTypeEcoin
}

// IsMoneyPay 是否为货币支付
func (o *Order) IsMoneyPay() bool {
	return o.PayType == PayTypeMoney
}
