package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"wdkr-marketplace-service/internal/common/config"
)

const (
	// JWTValidationHandlerKey 注册到 restserver 的中间件名
	JWTValidationHandlerKey = "jwt-validation-handler"
	// JWTBodyField HTTP JSON body 中存放 JWT 字符串的字段名
	JWTBodyField = "jwt"
)

var jwtReservedClaimKeys = map[string]struct{}{
	"account": {},
	"iss":     {}, "sub": {}, "aud": {},
	"exp": {}, "nbf": {}, "iat": {}, "jti": {},
}

type JWTValidationHandler struct{}

func (c *JWTValidationHandler) Name() string {
	return JWTValidationHandlerKey
}

func (c *JWTValidationHandler) Handle(ctx *gin.Context) {
	conf := config.GetConf()
	path := ctx.Request.URL.Path

	if !shouldApplyJWTAuth(path, conf) {
		ctx.Next()
		return
	}

	j := conf.JWT
	if !j.HasJWTClients() {
		ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"retcode": -1,
			"message": "jwt auth not configured",
		})
		return
	}

	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"retcode": -1,
			"message": "read body failed",
		})
		return
	}
	_ = ctx.Request.Body.Close()

	var wrapper struct {
		JWT string `json:"jwt"`
	}
	if err := json.Unmarshal(bodyBytes, &wrapper); err != nil || strings.TrimSpace(wrapper.JWT) == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"retcode": -1,
			"message": "invalid body: jwt field required",
		})
		return
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(wrapper.JWT, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		mc, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("invalid claims type")
		}
		acc, ok := stringClaim(mc, "account")
		if !ok {
			return nil, fmt.Errorf("missing account in jwt payload")
		}
		sec, ok := conf.JWT.SecretForAccount(acc)
		if !ok {
			return nil, fmt.Errorf("unknown account")
		}
		return []byte(sec), nil
	})
	if err != nil || token == nil || !token.Valid {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"retcode": -1,
			"message": "invalid jwt",
		})
		return
	}

	maxTTL := conf.JWT.EffectiveJWTExpirationSeconds()
	if err := validateJWTTTL(claims, maxTTL); err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"retcode": -1,
			"message": err.Error(),
		})
		return
	}

	payload := make(map[string]interface{})
	for k, v := range claims {
		if _, reserved := jwtReservedClaimKeys[k]; reserved {
			continue
		}
		payload[k] = v
	}

	out, err := json.Marshal(payload)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"retcode": -1,
			"message": "marshal payload failed",
		})
		return
	}

	ctx.Request.Body = io.NopCloser(bytes.NewReader(out))
	ctx.Request.ContentLength = int64(len(out))
	ctx.Request.Header.Set("Content-Type", "application/json")

	ctx.Next()
}

// validateJWTTTL 要求携带 iat、exp；exp 未过期，且 exp-iat 不超过配置的最大存活时间
func validateJWTTTL(claims jwt.MapClaims, maxSeconds uint32) error {
	iat, ok := claimUnix(claims, "iat")
	if !ok {
		return fmt.Errorf("jwt iat required")
	}
	exp, ok := claimUnix(claims, "exp")
	if !ok {
		return fmt.Errorf("jwt exp required")
	}
	now := time.Now().Unix()
	if exp <= now {
		return fmt.Errorf("jwt expired")
	}
	if exp <= iat {
		return fmt.Errorf("invalid jwt exp/iat")
	}
	if exp-iat > int64(maxSeconds) {
		return fmt.Errorf("jwt ttl exceeds %d seconds", maxSeconds)
	}
	// 允许小幅时钟偏差：签发时间不应远晚于当前（默认 60 秒内）
	if iat > now+60 {
		return fmt.Errorf("jwt iat in future")
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
	switch x := v.(type) {
	case string:
		return x, true
	case fmt.Stringer:
		return x.String(), true
	default:
		return "", false
	}
}

func shouldApplyJWTAuth(path string, conf *config.Conf) bool {
	if conf == nil {
		return false
	}
	j := conf.JWT
	if len(j.Paths) == 0 {
		return false
	}
	for _, prefix := range j.Paths {
		if prefix == "" {
			continue
		}
		if path == prefix || strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
