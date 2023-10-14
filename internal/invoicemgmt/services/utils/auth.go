package utils

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func SignCtx(ctx context.Context) context.Context {
	headers, ok := metadata.FromIncomingContext(ctx)
	var pkg, token, version string
	if ok {
		pkg = headers["pkg"][0]
		token = headers["token"][0]
		version = headers["version"][0]
	}
	return metadata.AppendToOutgoingContext(ctx, "pkg", pkg, "version", version, "token", token)
}
