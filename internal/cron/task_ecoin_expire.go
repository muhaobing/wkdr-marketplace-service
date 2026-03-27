package cron

import (
	"context"
	"fmt"
	"time"

	"wdkr-marketplace-service/internal/domain/ecoin"
)

const ecoinExpireScanLimit = 500

// EcoinExpireTask 积分过期扫描任务
type EcoinExpireTask struct {
	ecoinService ecoin.EcoinService
}

func NewEcoinExpireTask(ecoinService ecoin.EcoinService) *EcoinExpireTask {
	return &EcoinExpireTask{
		ecoinService: ecoinService,
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
	processed, err := t.ecoinService.ExpireEcoinStock(ctx, now, ecoinExpireScanLimit)
	if err != nil {
		return fmt.Errorf("expire ecoin stock: %w", err)
	}
	if processed > 0 {
		fmt.Printf("[EcoinExpireTask] processed expired groups: %d\n", processed)
	}
	return nil
}
