package coverage

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/logger"
	dpb "github.com/manabie-com/backend/pkg/manabuf/draft/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type SendCoverageServiceClient interface {
	dpb.SendCoverageServiceClient
}

type C struct {
	BaseRef           string
	HeadRef           string
	Ref               string
	CoverageFilepath  string
	RepositoryName    string
	SecretKey         string
	ServerAddr        string
	IsIntegrationTest bool

	// These are used as global variables in cmd only
	LogLevelString   string
	TimeoutInSeconds int

	grpcConn   *grpc.ClientConn
	grpcClient SendCoverageServiceClient
}

// Dial creates a GRPC connection to the server and setups the underlying client.
// Before program exits, Close must be called to clean up those resources.
func (c *C) Dial(ctx context.Context) error {
	grpcconn, err := grpc.DialContext(ctx, c.ServerAddr,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS13})),
		grpc.WithBlock())
	if err != nil {
		return err
	}
	c.grpcConn = grpcconn
	c.grpcClient = dpb.NewSendCoverageServiceClient(grpcconn)
	return nil
}

// Close closes the underlying GRPC connection. It must be called
// before the program exits.
func (c *C) Close() {
	err := c.grpcConn.Close()
	if err != nil {
		logger.Errorf("Close error: %s", err)
	}
}

var (
	errEmptyRef              = errors.New("--ref cannot be empty")
	errEmptyBaseRef          = errors.New("--base-ref cannot be empty")
	errEmptyHeadRef          = errors.New("--head-ref cannot be empty")
	errEmptyRepositoryName   = errors.New("--repo cannot be empty")
	errEmptySecretKey        = errors.New("--key cannot be empty")
	errEmptyServerAddr       = errors.New("--address cannot be empty")
	errEmptyCoverageFilepath = errors.New("--coverage-file cannot by empty")
)

func (c *C) UpdateCoverage(ctx context.Context) error {
	if c.Ref == "" {
		return errEmptyRef
	}
	if c.RepositoryName == "" {
		return errEmptyRepositoryName
	}
	if c.SecretKey == "" {
		return errEmptySecretKey
	}
	if c.ServerAddr == "" {
		return errEmptyServerAddr
	}
	if c.CoverageFilepath == "" {
		return errEmptyCoverageFilepath
	}

	coverage, err := c.readCoverageFromFile(c.CoverageFilepath)
	if err != nil {
		return err
	}

	logger.Infof("--ref: %s", c.Ref)
	logger.Infof("--repo: %s", c.RepositoryName)
	logger.Infof("--coverage-file: %s (coverage: %f)", c.CoverageFilepath, coverage)
	logger.Infof("--address: %s", c.ServerAddr)
	_, err = c.grpcClient.UpdateTargetCoverage(ctx, &dpb.UpdateTargetCoverageRequest{
		BranchName:  c.Ref,
		Repository:  c.RepositoryName,
		Key:         c.SecretKey,
		Coverage:    float32(coverage),
		Integration: c.IsIntegrationTest,
	})
	if err != nil {
		return fmt.Errorf("UpdateTargetCoverage: %s", err)
	}
	return nil
}

func (c C) readCoverageFromFile(filename string) (float64, error) {
	f, err := os.Open(filename)
	if err != nil {
		return 0, fmt.Errorf("os.Open: %w", err)
	}
	defer f.Close()
	return c.readCoverage(f)
}

var (
	reStr = `[-]?\d[\d,]*[\.]?[\d{2}]*`
	re    = regexp.MustCompile(reStr)
)

func (c C) readCoverage(f io.Reader) (float64, error) {
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	var line string
	for scanner.Scan() {
		line = scanner.Text()
	}

	matches := re.FindAllString(line, -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("failed to look up coverage in file using regexp %q", reStr)
	}

	return strconv.ParseFloat(matches[len(matches)-1], 64)
}

func (c *C) CompareCoverage(ctx context.Context) error {
	if c.BaseRef == "" {
		return errEmptyBaseRef
	}
	if c.HeadRef == "" {
		return errEmptyHeadRef
	}
	if c.RepositoryName == "" {
		return errEmptyRepositoryName
	}
	if c.SecretKey == "" {
		return errEmptySecretKey
	}
	if c.ServerAddr == "" {
		return errEmptyServerAddr
	}
	if c.CoverageFilepath == "" {
		return errEmptyCoverageFilepath
	}

	coverage, err := c.readCoverageFromFile(c.CoverageFilepath)
	if err != nil {
		return err
	}

	logger.Infof("--base-ref: %s", c.BaseRef)
	logger.Infof("--head-ref: %s", c.HeadRef)
	logger.Infof("--repo: %s", c.RepositoryName)
	logger.Infof("--coverage-file: %s (coverage: %f)", c.CoverageFilepath, coverage)
	logger.Infof("--address: %s", c.ServerAddr)

	res, err := c.grpcClient.SendCoverage(ctx, &dpb.SendCoverageRequest{
		Coverage:     float32(coverage),
		BranchName:   c.HeadRef,
		Repository:   c.RepositoryName,
		Key:          c.SecretKey,
		TargetBranch: c.BaseRef,
		Integration:  c.IsIntegrationTest,
	})
	if err != nil {
		// Ignore error when target branch does not exist
		// This can happen in cases of hotfixes.
		if strings.Contains(err.Error(), "target branch does not exist") {
			logger.Warnf("bypassing error %q", err)
			return nil
		}
		return err
	}
	logger.Infof("response: %s", res.GetMessage())
	return nil
}
