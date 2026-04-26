package callback_jwt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const encryptedPayloadClaimKey = "payload"

// BuildToken 将原始业务包体加密后签发 JWT。
func BuildToken(account, secret string, ttlSeconds uint32, rawPayload []byte) (string, error) {
	if account == "" {
		return "", errors.New("account is required")
	}
	if secret == "" {
		return "", errors.New("secret is required")
	}
	if len(rawPayload) == 0 {
		return "", errors.New("raw payload is required")
	}
	if ttlSeconds == 0 {
		ttlSeconds = 300
	}

	encPayload, err := encrypt(rawPayload, secret)
	if err != nil {
		return "", err
	}

	now := time.Now().Unix()
	claims := jwt.MapClaims{
		"account":                account,
		"iat":                    now,
		"exp":                    now + int64(ttlSeconds),
		encryptedPayloadClaimKey: encPayload,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 校验 JWT 并解密得到原始业务包体。
func ParseToken(tokenString string, maxTTL uint32, secretResolver func(account string) (string, bool)) (account string, rawPayload []byte, err error) {
	if tokenString == "" {
		return "", nil, errors.New("jwt is required")
	}
	if secretResolver == nil {
		return "", nil, errors.New("secret resolver is required")
	}
	if maxTTL == 0 {
		maxTTL = 300
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		mc, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			return nil, errors.New("invalid claims type")
		}
		acc, ok := stringClaim(mc, "account")
		if !ok {
			return nil, errors.New("missing account in jwt payload")
		}
		sec, ok := secretResolver(acc)
		if !ok || sec == "" {
			return nil, errors.New("unknown account")
		}
		return []byte(sec), nil
	})
	if err != nil || token == nil || !token.Valid {
		return "", nil, errors.New("invalid jwt")
	}

	if err := validateJWTTTL(claims, maxTTL); err != nil {
		return "", nil, err
	}

	account, ok := stringClaim(claims, "account")
	if !ok {
		return "", nil, errors.New("missing account in jwt payload")
	}
	secret, ok := secretResolver(account)
	if !ok || secret == "" {
		return "", nil, errors.New("unknown account")
	}

	encPayload, ok := stringClaim(claims, encryptedPayloadClaimKey)
	if !ok {
		return "", nil, errors.New("missing encrypted payload")
	}
	rawPayload, err = decrypt(encPayload, secret)
	if err != nil {
		return "", nil, fmt.Errorf("decrypt payload failed: %w", err)
	}

	return account, rawPayload, nil
}

func validateJWTTTL(claims jwt.MapClaims, maxSeconds uint32) error {
	iat, ok := claimUnix(claims, "iat")
	if !ok {
		return errors.New("jwt iat required")
	}
	exp, ok := claimUnix(claims, "exp")
	if !ok {
		return errors.New("jwt exp required")
	}
	now := time.Now().Unix()
	if exp <= now {
		return errors.New("jwt expired")
	}
	if exp <= iat {
		return errors.New("invalid jwt exp/iat")
	}
	if exp-iat > int64(maxSeconds) {
		return fmt.Errorf("jwt ttl exceeds %d seconds", maxSeconds)
	}
	if iat > now+60 {
		return errors.New("jwt iat in future")
	}
	return nil
}

func claimUnix(claims jwt.MapClaims, key string) (int64, bool) {
	v, ok := claims[key]
	if !ok {
		return 0, false
	}
	switch x := v.(type) {
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return 0, false
		}
		return int64(x), true
	case json.Number:
		n, err := x.Int64()
		return n, err == nil
	case int64:
		return x, true
	case int:
		return int64(x), true
	default:
		return 0, false
	}
}

func stringClaim(claims jwt.MapClaims, key string) (string, bool) {
	v, ok := claims[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func encrypt(plaintext []byte, secret string) (string, error) {
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("new cipher failed: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("new gcm failed: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce failed: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	out := append(nonce, ciphertext...)
	return base64.RawURLEncoding.EncodeToString(out), nil
}

func decrypt(encoded string, secret string) ([]byte, error) {
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode payload failed: %w", err)
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("new cipher failed: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm failed: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return nil, errors.New("invalid encrypted payload")
	}
	nonce, ciphertext := raw[:nonceSize], raw[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt failed: %w", err)
	}
	return plaintext, nil
}
