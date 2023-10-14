package interceptors

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
)

func GetOutgoingContext(ctx context.Context) (context.Context, error) {
	headers, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, fmt.Errorf("missing header")
	}
	p := headers["pkg"]
	tkn := headers["token"]
	v := headers["version"]
	if len(p) == 0 {
		return ctx, fmt.Errorf("missing pkg")
	}
	if len(tkn) == 0 {
		return ctx, fmt.Errorf("missing token")
	}
	if len(v) == 0 {
		return ctx, fmt.Errorf("missing version")
	}

	pkg := p[0]
	token := tkn[0]
	version := v[0]

	return metadata.AppendToOutgoingContext(ctx, "pkg", pkg, "version", version, "token", token), nil
}

func NewIncomingContext(ctx context.Context) context.Context {
	m := map[string]string{
		"pkg":     "com.manabie.liz",
		"version": "1.0.0",
		"token":   "token",
	}
	md := metadata.New(m)
	ctx = metadata.NewIncomingContext(ctx, md)
	return ctx
}
