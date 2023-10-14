package errorx

import (
	"regexp"
	"strings"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/runtime/protoiface"
)

func ReturnFirstErr(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

func GRPCErr(err error, details ...protoiface.MessageV1) error {
	e, ok := err.(errcode.Error)
	if !ok {
		e = errcode.Error{
			Code: errcode.InternalError,
			Err:  err,
		}
	}

	var statusCode codes.Code

	switch e.Code {
	case errcode.InternalError:
		statusCode = codes.Internal

	default:
		statusCode = codes.InvalidArgument
	}

	s := status.New(statusCode, err.Error())
	s, _ = s.WithDetails(details...)

	return s.Err()
}

func PbErrorMessage(err error) *upb.ErrorMessage {
	e, _ := err.(errcode.Error)

	fieldName := ExtractFieldName(e.FieldName)
	fieldName = getLastFieldName(fieldName)

	return &upb.ErrorMessage{
		FieldName: fieldName,
		Error:     e.Error(),
		Code:      int32(e.Code),
		Index:     int32(e.Index),
	}
}

func getLastFieldName(str string) string {
	strSlice := strings.Split(str, ".")
	return strSlice[len(strSlice)-1]
}

func ExtractFieldName(msg string) string {
	re := regexp.MustCompile(constant.ExtractTextBetweenQuotesPattern)
	matches := re.FindStringSubmatch(msg)

	if len(matches) >= 2 {
		msg = matches[1]
	}
	return msg
}
