package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	Success = 20001
	Fail    = 40001
)

var responseMessage = map[int]string{
	Success: "Success Transaction",
	Fail:    "Fail Transaction",
}

type ResponseData struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func SuccessResponse(c echo.Context, code int, data interface{}) error {
	return c.JSON(
		http.StatusOK,
		ResponseData{
			Code:    code,
			Message: responseMessage[code],
			Data:    data,
		},
	)
}

func FailResponse(c echo.Context, code int, data interface{}) error {
	return c.JSON(
		http.StatusBadGateway,
		ResponseData{
			Code:    code,
			Message: responseMessage[code],
			Data:    data,
		},
	)
}
