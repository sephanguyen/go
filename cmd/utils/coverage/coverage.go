package coverage

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/draft/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var re = regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)

func verifyCompareCoverageArgs(cmd *cobra.Command, args []string) error {
	return cobra.MinimumNArgs(2)(cmd, args)
}

func compareCoverage(cmd *cobra.Command, args []string) error {
	if address == "" {
		return fmt.Errorf("address is missing")
	}
	if branch == "" {
		return fmt.Errorf("branch is missing")
	}
	if repository == "" {
		return fmt.Errorf("repository is missing")
	}
	if baseBranch == "" {
		return fmt.Errorf("missing base branch")
	}
	if key == "" {
		return fmt.Errorf("secret key is missing")
	}

	// read coverage from file args[0]
	coverage, err := readCoverage(args[0])
	if err != nil {
		return err
	}

	// declare connection to grpc server
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	if err != nil {
		return err
	}

	defer conn.Close()

	// new draft client
	client := pb.NewSendCoverageServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// call send coverage
	r, err := client.SendCoverage(ctx, &pb.SendCoverageRequest{
		Coverage:     float32(coverage),
		BranchName:   branch,
		Repository:   repository,
		Key:          key,
		TargetBranch: baseBranch,
		Integration:  integration,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code coverage of base branch is") {
			return err
		}
		return fmt.Errorf("could not greet: %v", err)
	}
	fmt.Println(r.Message)
	return nil
}

func readCoverage(filename string) (float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0.0, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var line string
	for {
		l, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0.0, err
		}

		line = l
	}

	var coverage string
	for _, element := range re.FindAllString(line, -1) {
		coverage = element
	}

	return strconv.ParseFloat(coverage, 64)
}
