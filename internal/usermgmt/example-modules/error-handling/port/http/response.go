package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/manabie-com/backend/internal/usermgmt/example-modules/error-handling/core/errcode"

	"github.com/gin-gonic/gin"
)

func responseHandler(handler func(c *gin.Context) (interface{}, error)) func(c *gin.Context) {
	return func(c *gin.Context) {
		responseData, err := handler(c)

		if err == nil {
			c.JSON(
				http.StatusOK,
				Response{
					Data:    responseData,
					Code:    errcode.DomainCodeOK,
					Message: "success",
				},
			)
			return
		}

		switch err := err.(type) {
		// If error tpe is Domain Error (new error handling standard, prioritize this error handling flow)
		case errcode.DomainError:
			c.JSON(
				DomainErrorToHTTPResponseStatusCode(err),
				Response{
					Data:    responseData,
					Code:    err.DomainCode(),
					Message: err.DomainError(),
				},
			)
		// If not then go on with the legacy flow
		default:
			// Warning this is sample package, do not apply this in production pkg
			panic("please handle error correctly")
		}
	}
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func JSONDecode(src io.Reader, dst interface{}) error {
	if err := json.NewDecoder(src).Decode(dst); err != nil {
		switch err := err.(type) {
		case *json.UnsupportedValueError, *json.InvalidUnmarshalError:
			return err
		default:
			return InternalError{RawError: err}
		}
	}
	return nil
}
