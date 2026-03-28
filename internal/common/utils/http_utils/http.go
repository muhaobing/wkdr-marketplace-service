package http_utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CommonResponse struct {
	Retcode int         `json:"retcode"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func WriteResponse(ctx *gin.Context, data interface{}, err error) {
	if err != nil {
		ctx.JSON(http.StatusOK, &CommonResponse{
			Retcode: -1,
			Message: err.Error(),
		})
	} else {
		ctx.JSON(http.StatusOK, &CommonResponse{
			Retcode: 0,
			Message: "success",
			Data:    data,
		})
	}
}

// WriteResponseWithRetcode 指定 retcode（如 err_code.UserBindingNotFound），HTTP 仍为 200
func WriteResponseWithRetcode(ctx *gin.Context, retcode int, message string) {
	ctx.JSON(http.StatusOK, &CommonResponse{
		Retcode: retcode,
		Message: message,
	})
}
