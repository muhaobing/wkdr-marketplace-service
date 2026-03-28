package sku_model

const (
	SkuTabName = "sku_tab"

	// SkuStatusOffline 未上架状态
	SkuStatusOffline uint8 = 0
	// SkuStatusOnline 已上架状态
	SkuStatusOnline uint8 = 1

	// FulfillModeCallback 履约模式：接口回调
	FulfillModeCallback uint8 = 0
	// FulfillModeEcoinGrant 履约模式：积分发放
	FulfillModeEcoinGrant uint8 = 1
)

type Sku struct {
	Id                 uint64  `gorm:"column:id" json:"id"`
	BizCode            string  `gorm:"column:biz_code" json:"biz_code"`                         // BizCode 商品所属的业务编码
	SkuCode            string  `gorm:"column:sku_code" json:"sku_code"`                         // SkuCode 商品代码，在 BizCode 下唯一
	SkuName            string  `gorm:"column:sku_name" json:"sku_name"`                         // SkuName 商品名
	SkuAvatar          string  `gorm:"column:sku_avatar" json:"sku_avatar"`                     // SkuAvatar 商品图标
	SkuDesc            string  `gorm:"column:sku_desc" json:"sku_desc"`                         // SkuDesc 商品描述
	SkuStatus          uint8   `gorm:"column:sku_status" json:"sku_status"`                     // SkuStatus 上架状态。0-未上架，1-已上架
	Cost               float32 `gorm:"column:cost" json:"cost"`                                 // Cost 商品售价
	DeliveryMethod     string  `gorm:"column:delivery_method" json:"delivery_method"`           // DeliveryMethod 商品履约回调接口（FulfillModeCallback 时必填）。接口为http POST请求，body固定为{"sku_code": "商品代码", "biz_user_id": "用户ID"}，response为{"retcode": 0, "message": ""}，retcode为0表示成功，非0表示失败
	FulfillMode        uint8   `gorm:"column:fulfill_mode" json:"fulfill_mode"`                 // FulfillMode 履约模式：0-接口回调，1-积分发放
	FulfillEcoinAmount float64 `gorm:"column:fulfill_ecoin_amount" json:"fulfill_ecoin_amount"` // FulfillEcoinAmount 积分发放模式下每件商品发放的积分数（×下单数量）
	MultiSelect        uint8   `gorm:"column:multi_select" json:"multi_select"`                 // MultiSelect 是否支持多选下单。0-不支持，1-支持
	Ctime              uint32  `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime              uint32  `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

func (s *Sku) TableName() string {
	return SkuTabName
}

// IsOnline 检查商品是否已上架
func (s *Sku) IsOnline() bool {
	return s.SkuStatus == SkuStatusOnline
}

// IsOffline 检查商品是否未上架
func (s *Sku) IsOffline() bool {
	return s.SkuStatus == SkuStatusOffline
}
