package examples

import (
	"context"
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const (
	// Stag point to normal Staging API version
	Stag = "api.staging.manabie.io:443"
)

func AuthorizedContext(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx,
		"pkg", "com.manabie.liz",
		"version", "1.0.0",
		"token", token)
}

// SimplifiedDial can be use with listed Domain, using insecure when calling localhost to bypass TLS
func SimplifiedDial(host string, insecure bool) *grpc.ClientConn {
	dialWithTransportSecurityOption := grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	if insecure {
		dialWithTransportSecurityOption = grpc.WithInsecure()
	}

	conn, err := grpc.Dial(host, dialWithTransportSecurityOption, grpc.WithBlock())
	if err != nil {
		panic(err.Error())
	}

	return conn
}
