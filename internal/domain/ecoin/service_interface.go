package ecoin

import (
	"context"

	"wdkr-marketplace-service/internal/domain/ecoin/ecoin_model"
)

// AddEcoinRequest 增加积分请求
type AddEcoinRequest struct {
	UserId      uint64  `json:"user_id"`
	Amount      float64 `json:"amount"`
	SourceType  string  `json:"source_type"`
	SourceId    string  `json:"source_id"`
	Description string  `json:"description"`
}

// DeductEcoinRequest 扣除积分请求
type DeductEcoinRequest struct {
	UserId      uint64  `json:"user_id"`
	Amount      float64 `json:"amount"`
	SourceType  string  `json:"source_type"`
	SourceId    string  `json:"source_id"`
	Description string  `json:"description"`
}

// EcoinTransactionListRequest 积分流水查询请求
type EcoinTransactionListRequest struct {
	UserId uint64 `json:"user_id"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

// EcoinTransactionListResponse 积分流水查询响应
type EcoinTransactionListResponse struct {
	Total int64                           `json:"total"`
	List  []*ecoin_model.EcoinTransaction `json:"list"`
}

// EcoinStockGroupListRequest 积分库存分组查询请求
type EcoinStockGroupListRequest struct {
	UserId uint64 `json:"user_id"`
}

// EcoinStockGroupListResponse 积分库存分组查询响应
type EcoinStockGroupListResponse struct {
	TotalStock float64                            `json:"total_stock"`
	List       []*ecoin_model.UserEcoinStockGroup `json:"list"`
}

// EcoinService 积分服务接口
type EcoinService interface {
	// GetUserEcoin 获取用户积分信息
	GetUserEcoin(ctx context.Context, userId uint64) (*ecoin_model.UserEcoin, error)

	// AddEcoin 增加积分
	AddEcoin(ctx context.Context, req *AddEcoinRequest) (*ecoin_model.EcoinTransaction, error)

	// DeductEcoin 扣除积分
	DeductEcoin(ctx context.Context, req *DeductEcoinRequest) (*ecoin_model.EcoinTransaction, error)

	// GetEcoinTransactionList 获取积分流水列表
	GetEcoinTransactionList(ctx context.Context, req *EcoinTransactionListRequest) (*EcoinTransactionListResponse, error)

	// GetEcoinTransaction 根据ID获取积分流水
	GetEcoinTransaction(ctx context.Context, id uint64) (*ecoin_model.EcoinTransaction, error)

	// InitUserEcoin 初始化用户积分账户
	InitUserEcoin(ctx context.Context, userId uint64) (*ecoin_model.UserEcoin, bool, error)

	// GetEcoinStockGroupList 获取积分库存分组
	GetEcoinStockGroupList(ctx context.Context, req *EcoinStockGroupListRequest) (*EcoinStockGroupListResponse, error)

	// ExpireEcoinStock 过期积分处理（定时任务调用）
	ExpireEcoinStock(ctx context.Context, now uint32, limit int) (int, error)

	// DeductEcoinInTx 在已有数据库事务内扣除积分（不再开启新事务；无 Redis 幂等）
	DeductEcoinInTx(ctx context.Context, req *DeductEcoinRequest) (*ecoin_model.EcoinTransaction, error)

	// AddEcoinInTx 在已有数据库事务内增加积分（不再开启新事务；无 Redis 幂等）
	AddEcoinInTx(ctx context.Context, req *AddEcoinRequest) (*ecoin_model.EcoinTransaction, error)
}
