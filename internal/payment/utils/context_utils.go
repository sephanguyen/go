package utils

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func SignCtx(ctx context.Context) context.Context {
	headers, ok := metadata.FromIncomingContext(ctx)
	var pkg, token, version string
	if ok {
		pkg = headers[pkgHeaderKey][0]
		token = headers[tokenHeaderKey][0]
		version = headers[versionHeaderKey][0]
	}
	return metadata.AppendToOutgoingContext(ctx, pkgHeaderKey, pkg, versionHeaderKey, version, tokenHeaderKey, token)
}
