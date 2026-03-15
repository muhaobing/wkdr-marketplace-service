package auth_utils

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muhaobing/std-go/go-common/crypto"
)

func GenAuthToken(sessionId string, key string) (string, error) {
	token := fmt.Sprintf("%s.%d", sessionId, time.Now().UnixMicro())
	suffix, err := crypto.AesEncrypt(token, key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%s", token, suffix), nil
}

func ParseAuthToken(token string, key string) (string, error) {
	elems := strings.Split(token, ".")
	if len(elems) != 3 {
		return "", errors.New("invalid token")
	}
	sessionId := elems[0]
	timestamp := elems[1]
	suffix, err := crypto.AesEncrypt(fmt.Sprintf("%s.%s", sessionId, timestamp), key)
	if err != nil {
		return "", err
	}
	if suffix != elems[2] {
		return "", errors.New("invalid token")
	}
	return sessionId, nil
}

func WrapGinContext(ctx *gin.Context, session string) {
	ctx.Set(contextKeySession, session)
}

func FromContext(ctx context.Context) string {
	return ctx.Value(contextKeySession).(string)
}
