package companyecoin

import (
	"context"

	"wdkr-marketplace-service/internal/domain/companyecoin/companyecoin_model"
)

// AddCompanyEcoinRequest 企业入账
type AddCompanyEcoinRequest struct {
	CompanyId      uint64
	OperatorUserId uint64
	Amount         float64
	SourceType     string
	SourceId       string
	Description    string
}

// DeductCompanyEcoinRequest 企业扣款
type DeductCompanyEcoinRequest struct {
	CompanyId      uint64
	OperatorUserId uint64
	Amount         float64
	SourceType     string
	SourceId       string
	Description    string
}

// TransactionListRequest 流水列表
type TransactionListRequest struct {
	CompanyId uint64
	Offset    int
	Limit     int
}

type TransactionListResponse struct {
	Total int64                                         `json:"total"`
	List  []*companyecoin_model.CompanyEcoinTransaction `json:"list"`
}

// CompanyEcoinService 企业积分（按 company_id 串行 FOR UPDATE）
type CompanyEcoinService interface {
	InitCompanyEcoin(ctx context.Context, companyId uint64) (*companyecoin_model.CompanyEcoin, error)
	GetCompanyEcoin(ctx context.Context, companyId uint64) (*companyecoin_model.CompanyEcoin, error)
	AddCompanyEcoin(ctx context.Context, req *AddCompanyEcoinRequest) (*companyecoin_model.CompanyEcoinTransaction, error)
	DeductCompanyEcoin(ctx context.Context, req *DeductCompanyEcoinRequest) (*companyecoin_model.CompanyEcoinTransaction, error)
	GetTransactionList(ctx context.Context, req *TransactionListRequest) (*TransactionListResponse, error)
	ExpireCompanyEcoinStock(ctx context.Context, now uint32, limit int) (int, error)

	// AddCompanyEcoinInTx / DeductCompanyEcoinInTx 在已有数据库事务内操作（不再开启新事务；无 Redis 幂等）
	AddCompanyEcoinInTx(ctx context.Context, req *AddCompanyEcoinRequest) (*companyecoin_model.CompanyEcoinTransaction, error)
	DeductCompanyEcoinInTx(ctx context.Context, req *DeductCompanyEcoinRequest) (*companyecoin_model.CompanyEcoinTransaction, error)
}
