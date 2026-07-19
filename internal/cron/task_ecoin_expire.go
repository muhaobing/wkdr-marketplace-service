package cron

import (
	"context"
	"fmt"
	"time"

	"wdkr-marketplace-service/internal/domain/companyecoin"
	"wdkr-marketplace-service/internal/domain/ecoin"
)

const ecoinExpireScanLimit = 500

// EcoinExpireTask 金币过期扫描任务（个人 + 企业）
type EcoinExpireTask struct {
	ecoinService        ecoin.EcoinService
	companyEcoinService companyecoin.CompanyEcoinService
}

func NewEcoinExpireTask(ecoinService ecoin.EcoinService, companyEcoinService companyecoin.CompanyEcoinService) *EcoinExpireTask {
	return &EcoinExpireTask{
		ecoinService:        ecoinService,
		companyEcoinService: companyEcoinService,
	}
}

func (t *EcoinExpireTask) Name() string {
	return "ecoin_expire_scan"
}

func (t *EcoinExpireTask) Ticker() time.Duration {
	return 10 * time.Second
}

func (t *EcoinExpireTask) Handle(ctx context.Context) error {
	now := uint32(time.Now().Unix())
	n1, err := t.ecoinService.ExpireEcoinStock(ctx, now, ecoinExpireScanLimit)
	if err != nil {
		return fmt.Errorf("expire ecoin stock: %w", err)
	}
	if n1 > 0 {
		fmt.Printf("[EcoinExpireTask] processed user expired groups: %d\n", n1)
	}
	n2, err := t.companyEcoinService.ExpireCompanyEcoinStock(ctx, now, ecoinExpireScanLimit)
	if err != nil {
		return fmt.Errorf("expire company ecoin stock: %w", err)
	}
	if n2 > 0 {
		fmt.Printf("[EcoinExpireTask] processed company expired groups: %d\n", n2)
	}
	return nil
}
