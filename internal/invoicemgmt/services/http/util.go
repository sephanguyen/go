package http

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"

	"github.com/manabie-com/backend/internal/invoicemgmt/services/http/errcode"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func ResponseSuccess(data map[string]interface{}) *Response {
	return &Response{
		Data:    data,
		Code:    20000,
		Message: "success",
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
	case *mapstructure.Error:
		resp.Code = errcode.InvalidData
		for _, msg := range err.Errors {
			fieldName := msg
			resp.Message = errcode.Error{Code: resp.Code, FieldName: fieldName}.Error()
			break
		}
	}
	c.Errors = append(c.Errors, c.Error(logError))

	statusCode, _ := strconv.Atoi(strconv.Itoa(resp.Code)[0:3])
	c.AbortWithStatusJSON(statusCode, &resp)
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

func ParseJSONPayload(req *http.Request, output interface{}) error {
	var data map[string]interface{}

	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		return errors.Wrap(err, "json.NewDecoder(req.Body).Decode(")
	}

	if len(data) == 0 {
		return errors.New("no payload to upsert")
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: jsonUnmarshallerHookFunc(),
		Result:     &output,
		TagName:    "json",
	})
	if err != nil {
		return errors.Wrap(err, "mapstructure.NewDecoder")
	}
	err = decoder.Decode(data)
	if err != nil {
		return err
	}
	return nil
}
