package helpers

import (
	"context"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"google.golang.org/grpc/metadata"
)

func intResourcePathFromCtx(ctx context.Context) int64 {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim != nil && claim.Manabie != nil {
		rp := claim.Manabie.ResourcePath
		intrp, err := strconv.ParseInt(rp, 10, 64)
		if err != nil {
			panic(err)
		}
		return intrp
	}
	panic("ctx has no resource path")
}

func contextWithToken(ctx context.Context, token string) context.Context {
	// clear old auth info in context
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(make(map[string]string)))
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithTokenAndTimeOut(ctx context.Context, token string) (context.Context, context.CancelFunc) {
	newCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	return contextWithToken(newCtx, token), cancel
}
