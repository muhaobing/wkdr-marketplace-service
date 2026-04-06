package ecoinbill_model

import "fmt"

const (
	// ShardCount 分表数量（ecoin_bill_tab_0 .. ecoin_bill_tab_9）
	ShardCount = 10
)

const (
	StatusIncomplete = 1 // 预扣中（未完成确认）
	StatusCompleted  = 2 // 已确认扣款
	StatusCancelled  = 3 // 已取消并退款
)

// TableNameForShard 按分片下标返回表名（0–9）；shard 须在合法范围内由调用方保证。
func TableNameForShard(shard int) string {
	return fmt.Sprintf("ecoin_bill_tab_%08d", shard)
}

// TableNameForUser 按用户 ID 取模得到分表名。
func TableNameForUser(userID uint64) string {
	return TableNameForShard(int(userID % ShardCount))
}

// EcoinBill 积分账单（按 user_id 分表；主键为 bill_id）
type EcoinBill struct {
	BillId            string  `gorm:"column:bill_id;primaryKey" json:"bill_id"`
	UserId            uint64  `gorm:"column:user_id" json:"user_id"`
	CompanyId         uint64  `gorm:"column:company_id" json:"company_id"` // 0=个人积分；>0=企业积分
	Amount            float64 `gorm:"column:amount" json:"amount"`
	Status            uint8   `gorm:"column:status" json:"status"`
	DeductUserTxId    uint64  `gorm:"column:deduct_user_tx_id" json:"deduct_user_tx_id"`
	DeductCompanyTxId uint64  `gorm:"column:deduct_company_tx_id" json:"deduct_company_tx_id"`
	Ctime             uint32  `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime             uint32  `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
}

func (p *EcoinBill) IsIncomplete() bool {
	return p.Status == StatusIncomplete
}
