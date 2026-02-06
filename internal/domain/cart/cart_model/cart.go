package cartmodel

const (
	CartItemTabName = "cart_item_tab"
)

// CartItem 购物车商品项
type CartItem struct {
	Id       uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserId   uint64 `gorm:"column:user_id" json:"user_id"`   // 用户ID
	SkuId    uint64 `gorm:"column:sku_id" json:"sku_id"`     // 商品ID
	Quantity int    `gorm:"column:quantity" json:"quantity"` // 数量
	Ctime    uint32 `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime    uint32 `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

func (c *CartItem) TableName() string {
	return CartItemTabName
}

// CartItemWithSku 购物车商品项（包含商品信息）
type CartItemWithSku struct {
	CartItem
	SkuCode   string  `json:"sku_code"`
	SkuName   string  `json:"sku_name"`
	SkuAvatar string  `json:"sku_avatar"`
	Cost      float32 `json:"cost"`
	SkuStatus uint8   `json:"sku_status"`
}
