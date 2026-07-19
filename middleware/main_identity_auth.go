package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"wdkr-marketplace-service/internal/common/config"
	"wdkr-marketplace-service/internal/common/utils/auth_utils"
	"wdkr-marketplace-service/internal/domain/user"
)

const MainIdentityAuthHandlerKey = "main-identity-auth-handler"

const (
	bizCodeLawMindToC         = "LawMind_ToC"
	bizCodeLawMindEnterprise  = "LawMind_Enterprise"
	mainAuthHeaderKey         = "Authorization"
)

var userServiceForAuth user.UserService

// SetUserServiceForAuth 由 main 在初始化 Resources 后注入，供中间件解析/建号。
func SetUserServiceForAuth(svc user.UserService) {
	userServiceForAuth = svc
}

type MainIdentityAuthHandler struct{}

func (c *MainIdentityAuthHandler) Name() string {
	return MainIdentityAuthHandlerKey
}

func (c *MainIdentityAuthHandler) Handle(ctx *gin.Context) {
	conf := config.GetConf()
	path := ctx.Request.URL.Path

	if shouldSkipMainIdentityAuth(path, conf) {
		ctx.Next()
		return
	}

	if userServiceForAuth == nil {
		ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"retcode": -1,
			"message": "main identity auth not configured",
		})
		return
	}

	claims, err := parseMainSiteJWT(ctx, conf)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"retcode": -1,
			"message": err.Error(),
		})
		return
	}

	bizUserID, err := claimUint64(claims, "userId")
	if err != nil || bizUserID == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"retcode": -1,
			"message": "invalid userId in token",
		})
		return
	}

	accountType, _ := claimString(claims, "accountType")
	mainCompanyID, _ := claimUint64(claims, "companyId")
	bizCode := resolveBizCode(accountType, mainCompanyID)
	displayLabel, _ := claimString(claims, "username")
	if displayLabel == "" {
		displayLabel, _ = claimString(claims, "sub")
	}

	u, err := userServiceForAuth.ResolveOrCreateByBiz(ctx.Request.Context(), &user.ResolveOrCreateByBizRequest{
		BizCode:       bizCode,
		BizUserId:     bizUserID,
		MainCompanyId: mainCompanyID,
		DisplayLabel:  displayLabel,
	})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"retcode": -1,
			"message": err.Error(),
		})
		return
	}

	if requireAdminPath(path, conf) {
		role, _ := claimString(claims, "role")
		if !isMainSiteAdmin(role) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"retcode": -1,
				"message": "admin permission required",
			})
			return
		}
	}

	sessionJSON, err := u.ToJSON()
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"retcode": -1,
			"message": "serialize user failed",
		})
		return
	}
	auth_utils.WrapGinContext(ctx, sessionJSON)
	ctx.Next()
}

func shouldSkipMainIdentityAuth(path string, conf *config.Conf) bool {
	if conf == nil {
		return false
	}
	for _, whitePath := range conf.Auth.Whitelist {
		if pathMatchesWhitelistEntry(path, whitePath) {
			return true
		}
	}
	return false
}

func requireAdminPath(path string, conf *config.Conf) bool {
	if conf == nil {
		return false
	}
	for _, adminPath := range conf.Auth.AdminPaths {
		if strings.HasPrefix(path, adminPath) {
			return true
		}
	}
	return false
}

func resolveBizCode(accountType string, companyID uint64) string {
	at := strings.ToUpper(strings.TrimSpace(accountType))
	if at == "COMPANY" || at == "OPERATIONS" {
		return bizCodeLawMindEnterprise
	}
	if companyID > 0 {
		return bizCodeLawMindEnterprise
	}
	return bizCodeLawMindToC
}

func isMainSiteAdmin(role string) bool {
	r := strings.ToUpper(strings.TrimSpace(role))
	return r == "SYSTEM_ADMIN" || r == "SUPER_ADMIN"
}

func parseMainSiteJWT(ctx *gin.Context, conf *config.Conf) (jwt.MapClaims, error) {
	if conf == nil || strings.TrimSpace(conf.MainJWT.Secret) == "" {
		return nil, errors.New("main jwt secret not configured")
	}

	tokenStr := strings.TrimSpace(ctx.GetHeader(mainAuthHeaderKey))
	if tokenStr == "" {
		return nil, errors.New("missing authorization")
	}
	if strings.HasPrefix(strings.ToLower(tokenStr), "bearer ") {
		tokenStr = strings.TrimSpace(tokenStr[7:])
	}
	if tokenStr == "" {
		return nil, errors.New("missing authorization token")
	}

	claims := jwt.MapClaims{}
	// 使用 JSON Number 解析，避免雪花 ID (userId/companyId) 超过 float64 安全整数范围导致精度丢失。
	parser := jwt.NewParser(jwt.WithJSONNumber())
	token, err := parser.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(conf.MainJWT.Secret), nil
	})
	if err != nil || token == nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	exp, ok := claimUnix(claims, "exp")
	if !ok || exp <= time.Now().Unix() {
		return nil, errors.New("token expired")
	}
	return claims, nil
}

func claimString(claims jwt.MapClaims, key string) (string, bool) {
	v, ok := claims[key]
	if !ok {
		return "", false
	}
	switch x := v.(type) {
	case string:
		return x, true
	default:
		return fmt.Sprintf("%v", x), true
	}
}

func claimUint64(claims jwt.MapClaims, key string) (uint64, error) {
	n, ok := claimUnix(claims, key)
	if !ok || n < 0 {
		return 0, errors.New("missing numeric claim")
	}
	return uint64(n), nil
}
