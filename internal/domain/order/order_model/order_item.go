package order_model

const (
	OrderItemTabName = "order_item_tab"

	// 履约状态
	FulfillStatusPending uint8 = 0 // 待履约
	FulfillStatusSuccess uint8 = 1 // 履约成功
	FulfillStatusFailed  uint8 = 2 // 履约失败
)

// OrderItem 订单明细项
type OrderItem struct {
	Id            uint64  `gorm:"column:id;primaryKey" json:"id"`
	OrderId       uint64  `gorm:"column:order_id" json:"order_id"`             // 订单ID
	OrderNo       string  `gorm:"column:order_no" json:"order_no"`             // 订单号
	SkuId         uint64  `gorm:"column:sku_id" json:"sku_id"`                 // SKU ID
	SkuCode       string  `gorm:"column:sku_code" json:"sku_code"`             // SKU编码（冗余）
	SkuName       string  `gorm:"column:sku_name" json:"sku_name"`             // SKU名称（冗余）
	SkuAvatar     string  `gorm:"column:sku_avatar" json:"sku_avatar"`         // SKU图标（冗余）
	Quantity      int     `gorm:"column:quantity" json:"quantity"`             // 数量
	UnitPrice     float32 `gorm:"column:unit_price" json:"unit_price"`         // 单价
	TotalPrice    float32 `gorm:"column:total_price" json:"total_price"`       // 小计
	FulfillStatus uint8   `gorm:"column:fulfill_status" json:"fulfill_status"` // 履约状态
	FulfillTime   uint32  `gorm:"column:fulfill_time" json:"fulfill_time"`     // 履约时间
	FulfillMsg    string  `gorm:"column:fulfill_msg" json:"fulfill_msg"`       // 履约消息
	Ctime         uint32  `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime         uint32  `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

func (o *OrderItem) TableName() string {
	return OrderItemTabName
}

// IsFulfillPending 是否待履约
func (o *OrderItem) IsFulfillPending() bool {
	return o.FulfillStatus == FulfillStatusPending
}

// IsFulfillSuccess 是否履约成功
func (o *OrderItem) IsFulfillSuccess() bool {
	return o.FulfillStatus == FulfillStatusSuccess
}

// IsFulfillFailed 是否履约失败
func (o *OrderItem) IsFulfillFailed() bool {
	return o.FulfillStatus == FulfillStatusFailed
}
