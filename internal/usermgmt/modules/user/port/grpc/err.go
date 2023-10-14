package grpc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

type MaximumRowsCSVError struct {
	RequestRows  int
	LimitRowsCSV int
}

func (err MaximumRowsCSVError) Error() string {
	return fmt.Sprintf(`In port layer, the number of rows is more than %d. The request number of rows: %d`, err.LimitRowsCSV, err.RequestRows)
}
func (err MaximumRowsCSVError) DomainError() string {
	return fmt.Sprintf(`The number of rows is more than %d rows`, err.LimitRowsCSV)
}
func (err MaximumRowsCSVError) DomainCode() int {
	return errcode.InvalidMaximumRows
}

type InvalidPayloadSizeCSVError struct {
	RequestSize int
}

func (err InvalidPayloadSizeCSVError) Error() string {
	return fmt.Sprintf(`In port layer, the payload size is more than 5 MB or zero, request size: %d`, err.RequestSize)
}
func (err InvalidPayloadSizeCSVError) DomainError() string {
	return `The payload size is more than 5 MB or zero`
}
func (err InvalidPayloadSizeCSVError) DomainCode() int {
	return errcode.InvalidPayloadSize
}

type InternalError struct {
	RawErr error
}

func (err InternalError) Error() string {
	return fmt.Sprintf(`In port layer, internal error: '%s'`, err.RawErr.Error())
}

func (err InternalError) DomainError() string {
	return "internal error"
}
func (err InternalError) DomainCode() int {
	return errcode.InternalError
}

func ToPbErrorMessageBackOffice(err errcode.DomainError) *upb.ErrorMessage {
	field := GetFieldFromMessageError(err.DomainError())
	index := GetIndexFromMessageError(err.DomainError())

	return &upb.ErrorMessage{
		FieldName: field,
		Code:      int32(err.DomainCode()),
		Index:     int32(index),
		Error:     err.DomainError(),
	}
}

func ToPbErrorMessageImport(err errcode.DomainError) *upb.ErrorMessage {
	field := ToFieldImport(err.DomainError())
	index := GetIndexFromMessageError(err.DomainError())

	return &upb.ErrorMessage{
		FieldName: field,
		Code:      int32(err.DomainCode()),
		Index:     int32(index + 2),
	}
}

func ToFieldImport(messageError string) string {
	if strings.Contains(messageError, string(entity.EnrollmentStatusHistories)) &&
		strings.Contains(messageError, string(entity.StartDateFieldEnrollmentStatusHistory)) {
		return entity.StudentFieldEnrollmentStatusStartDate
	}
	return GetFieldFromMessageError(messageError)
}

func GetIndexFromMessageError(message string) int {
	re := regexp.MustCompile(`\[(\d+)\]`)
	matches := re.FindStringSubmatch(message)
	if len(matches) > 1 {
		number, _ := strconv.Atoi(matches[1])
		return number
	}
	return 0
}
func GetFieldFromMessageError(message string) string {
	re := regexp.MustCompile(`\.(\w+\s)|(\.(\w+(\[\d\]\s)))`)
	matches := re.FindStringSubmatch(message)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}
