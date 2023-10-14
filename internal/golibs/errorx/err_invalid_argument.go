package errorx

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InvalidArgumentReason string

const (
	InvalidArgumentReasonIsEmpty                  InvalidArgumentReason = "IS_EMPTY"
	InvalidArgumentReasonSmallerThanMinimumLength InvalidArgumentReason = "SMALLER_THAN_MINIMUM_LENGTH"
	InvalidArgumentReasonGreaterThanMaximumLength InvalidArgumentReason = "GREATER_THAN_MAXIMUM_LENGTH"
)

type InvalidArgumentError interface {
	FieldName() string
	Reason() InvalidArgumentReason
	Error() string
}

func invalidArgumentErrToString(invalidArgumentErr InvalidArgumentError) *status.Status {
	stt := status.New(codes.InvalidArgument, invalidArgumentErr.Error())
	details := &errdetails.BadRequest_FieldViolation{
		Field:       invalidArgumentErr.FieldName(),
		Description: string(invalidArgumentErr.Reason()),
	}

	sttWithDetail, err := stt.WithDetails(details)
	if err != nil {
		return status.New(codes.InvalidArgument, invalidArgumentErr.Error())
	}
	return sttWithDetail
}
