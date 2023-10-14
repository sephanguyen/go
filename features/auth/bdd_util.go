package auth

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithToken(ctx context.Context, token string) context.Context {
	return helper.GRPCContext(ctx, "token", token)
}

func newID() string {
	return idutil.ULIDNow()
}

func compareStatusCode(err error, code string) error {
	stt, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("returned error is not status.Status, err: %s", err.Error())
	}
	if stt.Code().String() != code {
		return fmt.Errorf("expecting %s, got %s status code, message: %s", code, stt.Code().String(), stt.Message())
	}
	return nil
}
