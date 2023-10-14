package coverage

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/draft/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func createTargetCoverage(cmd *cobra.Command, args []string) error {
	if address == "" {
		return fmt.Errorf("address is missing")
	}
	if branch == "" {
		return fmt.Errorf("branch is missing")
	}
	if repository == "" {
		return fmt.Errorf("repository is missing")
	}
	if key == "" {
		return fmt.Errorf("secret key is missing")
	}

	// read coverage from file args[0]
	coverage, err := readCoverage(args[0])
	if err != nil {
		return fmt.Errorf("readCoverage(args[0]): %v", err)
	}

	// declare connection to grpc server
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))) // #nosec G402
	if err != nil {
		return fmt.Errorf("grpc.Dial: %v", err)
	}

	defer conn.Close()

	// new draft client
	client := pb.NewSendCoverageServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// call send coverage
	_, err = client.CreateTargetCoverage(ctx, &pb.CreateTargetCoverageRequest{
		Coverage:    float32(coverage),
		BranchName:  branch,
		Repository:  repository,
		Key:         key,
		Integration: integration,
	})
	if err != nil {
		return fmt.Errorf("client.CreateTargetCoverage: %v", err)
	}
	return nil
}
