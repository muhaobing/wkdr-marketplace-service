package bill_id_gen

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Expected length: EB(2) + YYYYMMDDHHMMSS(14) + hash(8) + shard digit(1) = 25.
const BillIDLength = 25

var (
	ErrInvalidBillID = errors.New("invalid bill_id")
)

// GenerateBillID returns a 25-character bill id:
// EB + YYYYMMDDHHMMSS + first 8 hex chars of MD5(userId|timestamp|random) uppercase + shard digit (userId % 10).
func GenerateBillID(userID uint64) (string, error) {
	ts := time.Now().Format("20060102150405")
	rnd := make([]byte, 16)
	if _, err := rand.Read(rnd); err != nil {
		return "", err
	}
	sum := md5.Sum([]byte(fmt.Sprintf("%d%s%s", userID, ts, hex.EncodeToString(rnd))))
	hash8 := strings.ToUpper(hex.EncodeToString(sum[:])[:8])
	shard := userID % 10
	return fmt.Sprintf("EB%s%s%d", ts, hash8, shard), nil
}

// ShardIndexFromBillID returns the shard table index (0–9) encoded in the last character of bill_id.
func ShardIndexFromBillID(billID string) (int, error) {
	if len(billID) != BillIDLength {
		return -1, ErrInvalidBillID
	}
	if !strings.HasPrefix(billID, "EB") {
		return -1, ErrInvalidBillID
	}
	last := billID[len(billID)-1]
	if last < '0' || last > '9' {
		return -1, ErrInvalidBillID
	}
	n, err := strconv.Atoi(string(last))
	if err != nil {
		return -1, ErrInvalidBillID
	}
	return n, nil
}
