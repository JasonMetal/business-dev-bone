package core

import (
	"net/http"

	"business-dev-bone/pkg/component-base/errors"
	"business-dev-bone/pkg/component-base/log"

	"github.com/gin-gonic/gin"
)

// ErrResponse defines the return messages when an error occurred.
// Reference will be omitted if it does not exist.
// swagger:model
type ErrResponse struct {
	// Code defines the business error code.
	Code int `json:"code"`

	// Message contains the detail of this message.
	// This message is suitable to be exposed to external
	Message string `json:"message"`

	// Reference returns the reference document which maybe useful to solve this error.
	Reference string `json:"reference,omitempty"`
}

type Response struct {
	// Code defines the business error code.
	Code int `json:"code"`

	// Message contains the detail of this message.
	// This message is suitable to be exposed to external
	Message string `json:"message"`

	Data interface{} `json:"data"`
}
type CommonResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// WriteResponse write an error or the response data into http response body.
// It use errors.ParseCoder to parse any error into errors.Coder
// errors.Coder contains error code, user-safe error message.
func WriteResponse(c *gin.Context, err error, data interface{}) {
	errHttpStatus := http.StatusOK
	writeResponse(c, err, data, &errHttpStatus)
}

func WriteResponseWithStatusCode(c *gin.Context, err error, data interface{}, statusCode int) {
	writeResponse(c, err, data, &statusCode)
}

// WriteResponse write an error or the response data into http response body.
// It use errors.ParseCoder to parse any error into errors.Coder
// errors.Coder contains error code, user-safe error message and http status code.
func writeResponse(c *gin.Context, err error, data interface{}, errHttpStatus *int) {
	if err != nil {
		log.L(c).Errorf("api response error: %s, %#+v", err.Error(), err)
		coder := errors.ParseCoder(err)

		status := coder.HTTPStatus()
		if errHttpStatus != nil {
			status = *errHttpStatus
		}

		c.JSON(status, Response{
			Code:    coder.Code(),
			Message: coder.String(),
			Data:    data,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}
func WriteResponseWithErrMsg(c *gin.Context, err error, errMsg string, data interface{}, errHttpStatus int) {
	if err != nil {
		log.L(c).Errorf("api response error: %s, %#+v", err.Error(), err)
		coder := errors.ParseCoder(err)

		status := coder.HTTPStatus()
		if &errHttpStatus != nil {
			status = errHttpStatus
		}
		if errMsg != "" {
			c.JSON(status, Response{
				Code:    coder.Code(),
				Message: errMsg,
				Data:    data,
			})
			return
		}
		c.JSON(status, Response{
			Code:    coder.Code(),
			Message: coder.String(),
			Data:    data,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}
