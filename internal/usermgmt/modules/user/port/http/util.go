package http

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/errorx"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type DataResponse struct {
	UserID         string `json:"user_id"`
	ExternalUserID string `json:"external_user_id"`
}

type ResponseErrors struct {
	Response
	Errors []Response `json:"errors"`
}

func ParseJSONPayload(req *http.Request, output interface{}) error {
	var data map[string]interface{}

	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		switch err := err.(type) {
		case *json.UnsupportedValueError, *json.InvalidUnmarshalError:
			return err
		default:
			return InternalError{RawErr: errors.Wrap(err, "json.NewDecoder(req.Body).Decode")}
		}
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: jsonUnmarshallerHookFunc(),
		Result:     &output,
		TagName:    "json",
	})
	if err != nil {
		return InternalError{RawErr: errors.Wrap(err, "mapstructure.NewDecoder")}
	}
	err = decoder.Decode(data)
	if err != nil {
		return err
	}
	return nil
}

func jsonUnmarshallerHookFunc() mapstructure.DecodeHookFuncType {
	return func(
		from reflect.Type,
		to reflect.Type,
		data interface{}) (interface{}, error) {
		result := reflect.New(to).Interface()

		unmarshaller, ok := result.(json.Unmarshaler)
		if !ok {
			return data, nil
		}
		dataBytes, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		if err := unmarshaller.UnmarshalJSON(dataBytes); err != nil {
			return nil, err
		}
		return result, nil
	}
}

func ResponseError(c *gin.Context, err error) {
	// This is for logging errors with ginzap middleware
	logError := err

	code := errcode.InternalError
	resp := Response{
		Code:    code,
		Message: errcode.Error{Code: code}.Error(),
	}
	switch err := err.(type) {
	case errcode.Error:
		if err.Err != nil {
			logError = err.Err
		}
		resp.Code = err.Code
		resp.Message = err.Error()
	case errcode.DomainError:
		resp.Code = err.DomainCode()
		resp.Message = err.DomainError()
	case *mapstructure.Error:
		resp.Code = errcode.InvalidData
		for _, msg := range err.Errors {
			fieldName := errorx.ExtractFieldName(msg)
			resp.Message = errcode.Error{Code: resp.Code, FieldName: fieldName}.Error()
			break
		}
	}
	c.Errors = append(c.Errors, c.Error(logError))

	statusCode, _ := strconv.Atoi(strconv.Itoa(resp.Code)[0:3])
	c.AbortWithStatusJSON(statusCode, &resp)
}

func ResponseListErrors(c *gin.Context, errs []error) {
	resp := []Response{}
	for _, err := range errs {
		tem := Response{}
		switch err := err.(type) {
		case errcode.Error:
			tem.Code = err.Code
			tem.Message = err.Error()
		case errcode.DomainError:
			tem.Code = err.DomainCode()
			tem.Message = err.Error()
			tem.Error = err.DomainError()
		case *mapstructure.Error:
			tem.Code = errcode.InvalidData
			for _, msg := range err.Errors {
				fieldName := errorx.ExtractFieldName(msg)
				tem.Message = errcode.Error{Code: tem.Code, FieldName: fieldName}.Error()
				break
			}
		default:
			tem.Code = errcode.InternalError
			tem.Message = err.Error()
		}

		resp = append(resp, tem)
		c.Errors = append(c.Errors, c.Error(err))
	}
	err := ResponseErrors{
		Response: resp[0],
		Errors:   resp,
	}

	statusCode, _ := strconv.Atoi(strconv.Itoa(resp[0].Code)[0:3])

	c.AbortWithStatusJSON(statusCode, err)
}
