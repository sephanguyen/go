package utils

import (
	"fmt"

	"google.golang.org/grpc/status"
)

func ValidateStatusCode(grpcError error, arg string) error {
	stt, ok := status.FromError(grpcError)
	if !ok {
		return fmt.Errorf("returned error is not status.Status, err: %s", grpcError)
	}
	if stt.Code().String() != arg {
		return fmt.Errorf("expecting %s, got %s status code, message: %s", arg, stt.Code().String(), stt.Message())
	}
	return nil
}

func ContainsStr(s []string, target string) bool {
	for _, val := range s {
		if target == val {
			return true
		}
	}
	return false
}
