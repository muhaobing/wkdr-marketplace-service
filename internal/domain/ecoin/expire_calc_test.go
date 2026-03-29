package ecoin

import (
	"testing"
	"time"
)

func TestCalcEcoinExpireTimeFromUnix_sameMonthEnd(t *testing.T) {
	// 2026-03-15 12:00:00 Asia/Shanghai -> 当月最后一天 23:59:59
	loc := ecoinBizLocation
	purchase := time.Date(2026, 3, 15, 12, 0, 0, 0, loc)
	exp := CalcEcoinExpireTimeFromUnix(uint32(purchase.Unix()))
	want := time.Date(2026, 3, 31, 23, 59, 59, 0, loc)
	if exp != uint32(want.Unix()) {
		t.Fatalf("got %d want %d (%s)", exp, want.Unix(), want.Format(time.RFC3339))
	}
}

func TestCalcEcoinExpireTimeFromUnix_lastFiveDaysRollsToNextMonth(t *testing.T) {
	loc := ecoinBizLocation
	// 3 月共 31 天，最后 5 天从 27 日起；27 日购买 -> 次月（4 月）末
	purchase := time.Date(2026, 3, 27, 10, 0, 0, 0, loc)
	exp := CalcEcoinExpireTimeFromUnix(uint32(purchase.Unix()))
	want := time.Date(2026, 4, 30, 23, 59, 59, 0, loc)
	if exp != uint32(want.Unix()) {
		t.Fatalf("got %d want %d", exp, want.Unix())
	}
}

func TestCalcEcoinExpireTimeFromUnix_dayBeforeLastFiveWindow(t *testing.T) {
	loc := ecoinBizLocation
	// 3 月 26 日 -> 仍为本月 3 月 31 日
	purchase := time.Date(2026, 3, 26, 23, 0, 0, 0, loc)
	exp := CalcEcoinExpireTimeFromUnix(uint32(purchase.Unix()))
	want := time.Date(2026, 3, 31, 23, 59, 59, 0, loc)
	if exp != uint32(want.Unix()) {
		t.Fatalf("got %d want %d", exp, want.Unix())
	}
}

func TestCalcEcoinExpireTimeFromUnix_februaryLeap(t *testing.T) {
	loc := ecoinBizLocation
	// 2028-02 共 29 天，最后 5 天从 25 日起；24 日 -> 当月 29 日末
	purchase := time.Date(2028, 2, 24, 8, 0, 0, 0, loc)
	exp := CalcEcoinExpireTimeFromUnix(uint32(purchase.Unix()))
	want := time.Date(2028, 2, 29, 23, 59, 59, 0, loc)
	if exp != uint32(want.Unix()) {
		t.Fatalf("got %d want %d", exp, want.Unix())
	}
}

func TestCalcEcoinExpireTimeFromUnix_februaryLeap_roll(t *testing.T) {
	loc := ecoinBizLocation
	purchase := time.Date(2028, 2, 25, 8, 0, 0, 0, loc)
	exp := CalcEcoinExpireTimeFromUnix(uint32(purchase.Unix()))
	want := time.Date(2028, 3, 31, 23, 59, 59, 0, loc)
	if exp != uint32(want.Unix()) {
		t.Fatalf("got %d want %d", exp, want.Unix())
	}
}
