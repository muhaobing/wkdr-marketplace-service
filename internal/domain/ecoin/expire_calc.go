package ecoin

import (
	"time"
)

// 积分过期按业务日历（上海时区）计算，与服务器本地时区无关。
var ecoinBizLocation = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.UTC
	}
	return loc
}()

// CalcEcoinExpireTimeFromUnix 根据购买时刻（Unix 秒）计算该批次积分过期时间戳（秒）：
// - 默认：购买所在自然月最后一日 23:59:59（上海时区）；
// - 若购买日落在当月最后 5 个自然日内（含当月最后一天），则顺延至次月最后一日 23:59:59。
func CalcEcoinExpireTimeFromUnix(purchaseUnix uint32) uint32 {
	t := time.Unix(int64(purchaseUnix), 0).In(ecoinBizLocation)
	y, month, day := t.Date()
	lastDayOfMonth := time.Date(y, month+1, 0, 0, 0, 0, 0, ecoinBizLocation).Day()
	last5StartDay := lastDayOfMonth - 4 // 与当月最后一天共 5 天：lastDay-4 .. lastDay
	var end time.Time
	if day >= last5StartDay {
		end = time.Date(y, month+2, 0, 23, 59, 59, 0, ecoinBizLocation)
	} else {
		end = time.Date(y, month+1, 0, 23, 59, 59, 0, ecoinBizLocation)
	}
	return uint32(end.Unix())
}
