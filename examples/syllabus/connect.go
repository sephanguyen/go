package syllabus

import (
	"context"
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func Connect(host string, insecure bool) *grpc.ClientConn {
	dialWithTransportSecurityOption := grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	if insecure {
		dialWithTransportSecurityOption = grpc.WithInsecure()
	}

	conn, err := grpc.Dial(host, dialWithTransportSecurityOption)
	if err != nil {
		panic(err.Error())
	}

	return conn
}

func AuthorizedContext(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx,
		"pkg", "com.manabie.liz",
		"version", "1.0.0",
		"token", token)
}
