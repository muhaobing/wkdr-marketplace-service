package companyecoin_model

const CompanyEcoinTransactionTabName = "company_ecoin_transaction_tab"

// CompanyEcoinTransaction 企业积分流水（记录操作人）
type CompanyEcoinTransaction struct {
	Id             uint64  `gorm:"column:id" json:"id"`
	CompanyId      uint64  `gorm:"column:company_id" json:"company_id"`
	OperatorUserId uint64  `gorm:"column:operator_user_id" json:"operator_user_id"`
	OperatorLabel  string  `gorm:"-" json:"operator_label,omitempty"` // 接口层填充：操作人展示名
	Amount         float64 `gorm:"column:amount" json:"amount"`
	BeforeStock    float64 `gorm:"column:before_stock" json:"before_stock"`
	AfterStock     float64 `gorm:"column:after_stock" json:"after_stock"`
	TxType         int     `gorm:"column:tx_type" json:"tx_type"`
	SourceType     string  `gorm:"column:source_type" json:"source_type"`
	SourceId       string  `gorm:"column:source_id" json:"source_id"`
	Description    string  `gorm:"column:description" json:"description"`
	Status         int     `gorm:"column:status" json:"status"`
	Ctime          uint32  `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime          uint32  `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

func (t *CompanyEcoinTransaction) TableName() string {
	return CompanyEcoinTransactionTabName
}
