package utils

import (
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GroupErrorFunc(errors ...error) (err error) {
	for i := range errors {
		err = errors[i]
		if err != nil {
			return err
		}
	}
	return
}

func StatusErrWithDetail(code codes.Code, message string, detail proto.Message) error {
	stt := status.New(code, message)
	if detail != nil {
		stt, _ = stt.WithDetails(detail)
	}

	return stt.Err()
}
