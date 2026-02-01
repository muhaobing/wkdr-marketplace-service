package ecoin_model

const (
	// 积分操作类型
	TransactionTypeAdd    = 1 // 增加积分
	TransactionTypeDeduct = 2 // 扣除积分

	// 积分来源类型
	SourceTypeSystem   = "system"   // 系统赠送
	SourceTypeOrder    = "order"    // 订单奖励
	SourceTypeConsume  = "consume"  // 积分消费
	SourceTypeRefund   = "refund"   // 退款返还
	SourceTypeTransfer = "transfer" // 转账

	// 交易状态
	TransactionStatusPending   = 0 // 处理中
	TransactionStatusCompleted = 1 // 已完成
	TransactionStatusFailed    = 2 // 已失败

	// 表名
	EcoinTransactionTabName = "ecoin_transaction"
)

type EcoinTransaction struct {
	Id          uint64  `gorm:"column:id;primary_key;auto_increment" json:"id"`
	UserId      uint64  `gorm:"column:user_id;not null" json:"user_id"`
	Amount      float64 `gorm:"column:amount;not null" json:"amount"`                   // 积分变动数量，正数表示增加，负数表示扣除
	BeforeStock float64 `gorm:"column:before_stock;not null" json:"before_stock"`       // 操作前积分余额
	AfterStock  float64 `gorm:"column:after_stock;not null" json:"after_stock"`         // 操作后积分余额
	TxType      int     `gorm:"column:tx_type;not null" json:"tx_type"`                 // 交易类型：1-增加，2-扣除
	SourceType  string  `gorm:"column:source_type;not null;size:32" json:"source_type"` // 来源类型
	SourceId    string  `gorm:"column:source_id;size:64" json:"source_id"`              // 来源业务ID
	Description string  `gorm:"column:description;size:255" json:"description"`         // 描述信息
	Status      int     `gorm:"column:status;not null;default:0" json:"status"`         // 交易状态：0-处理中，1-已完成，2-已失败
	Ctime       uint32  `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime       uint32  `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

func (t *EcoinTransaction) TableName() string {
	return EcoinTransactionTabName
}
