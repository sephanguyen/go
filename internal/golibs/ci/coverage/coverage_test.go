package coverage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	mock_coverage "github.com/manabie-com/backend/mock/golibs/ci/coverage"
	dpb "github.com/manabie-com/backend/pkg/manabuf/draft/v1"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCUpdateCoverage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// test input validation
	type testcase struct {
		desc         string
		ref          string
		reponame     string
		coveragefile string
		address      string
		key          string
		expectedErr  error
	}

	testcases := []testcase{
		{desc: "missing ref", expectedErr: errEmptyRef},
		{desc: "missing repo name", ref: "develop", expectedErr: errEmptyRepositoryName},
		{desc: "missing secret key", ref: "develop", reponame: "manabie-com/backend", expectedErr: errEmptySecretKey},
		{
			desc: "missing server address", ref: "develop",
			reponame: "manabie-com/backend", key: "123", expectedErr: errEmptyServerAddr,
		},
		{
			desc: "missing server address", ref: "develop",
			reponame: "manabie-com/backend", key: "123", address: "draft-address", expectedErr: errEmptyCoverageFilepath,
		},
	}
	for _, tc := range testcases {
		c := C{
			Ref:              tc.ref,
			CoverageFilepath: tc.coveragefile,
			RepositoryName:   tc.reponame,
			SecretKey:        tc.key,
			ServerAddr:       tc.address,
		}
		err := c.UpdateCoverage(ctx)
		require.Equal(t, tc.expectedErr, err, "test case %q failed", tc.desc)
	}

	// test functionality
	draftClient := mock_coverage.NewSendCoverageServiceClient(t)
	c := C{
		Ref:              "develop",
		CoverageFilepath: getSampleCovFile(t),
		RepositoryName:   "manabie-com/backend",
		SecretKey:        "123",
		ServerAddr:       "draft-server-address",
		grpcClient:       draftClient,
	}
	draftClient.
		On("UpdateTargetCoverage", ctx, &dpb.UpdateTargetCoverageRequest{
			BranchName:  "develop",
			Repository:  "manabie-com/backend",
			Key:         "123",
			Coverage:    float32(getSampleCov()),
			Integration: false,
		}).
		Return(&dpb.UpdateTargetCoverageResponse{}, nil).
		Once()
	err := c.UpdateCoverage(ctx)
	require.NoError(t, err)
}

var (
	sampleCov float64 = 68.9
	once      sync.Once
)

func getSampleCovFile(t *testing.T) string {
	sampleFilepath := filepath.Join(t.TempDir(), "cover")
	f, err := os.OpenFile(sampleFilepath, os.O_RDWR|os.O_CREATE, 0o755)
	if err != nil {
		t.Fatalf("failed to create mock coverage file: %s", err)
	}
	defer f.Close()
	content := fmt.Sprintf(`github.com/manabie-com/backend/internal/zeus/subscriptions/activity_log_subscription.go:50: CreateActivityLog 0.0%%
total: 	(statements)  	%.1f%%`, sampleCov)
	_, err = f.WriteString(content)
	if err != nil {
		t.Fatalf("failed to write sample coverage content: %s", err)
	}
	return sampleFilepath
}

func getSampleCov() float64 {
	return sampleCov
}

func TestCCompareCoverage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// validate inputs
	type testcase struct {
		desc         string
		baseref      string
		headref      string
		reponame     string
		coveragefile string
		address      string
		key          string
		expectedErr  error
	}
	testcases := []testcase{
		{desc: "missing --base-ref", expectedErr: errEmptyBaseRef},
		{desc: "missing --head-ref", baseref: "develop", expectedErr: errEmptyHeadRef},
		{desc: "missing --repo", baseref: "develop", headref: "feature", expectedErr: errEmptyRepositoryName},
		{
			desc: "missing --key", baseref: "develop", headref: "feature",
			reponame: "manabie-com/backend", expectedErr: errEmptySecretKey,
		},
		{
			desc: "missing --address", baseref: "develop", headref: "feature",
			reponame: "manabie-com/backend", key: "123", expectedErr: errEmptyServerAddr,
		},
		{
			desc: "missing --coverage-file", baseref: "develop", headref: "feature",
			reponame: "manabie-com/backend", key: "123", address: "server address", expectedErr: errEmptyCoverageFilepath,
		},
	}
	for _, tc := range testcases {
		c := C{
			BaseRef:           tc.baseref,
			HeadRef:           tc.headref,
			CoverageFilepath:  tc.coveragefile,
			RepositoryName:    tc.reponame,
			SecretKey:         tc.key,
			ServerAddr:        tc.address,
			IsIntegrationTest: false,
		}
		err := c.CompareCoverage(ctx)
		require.Equal(t, tc.expectedErr, err, "test case %q failed", tc.desc)
	}

	// test functionality
	draftClient := mock_coverage.NewSendCoverageServiceClient(t)
	c := C{
		BaseRef:          "develop",
		HeadRef:          "feature-branch",
		CoverageFilepath: getSampleCovFile(t),
		RepositoryName:   "manabie-com/backend",
		SecretKey:        "123",
		ServerAddr:       "draft-server-address",
		grpcClient:       draftClient,
	}
	draftClient.
		On("SendCoverage", ctx, &dpb.SendCoverageRequest{
			Coverage:     float32(sampleCov),
			BranchName:   "feature-branch",
			Repository:   "manabie-com/backend",
			Key:          "123",
			TargetBranch: "develop",
			Integration:  false,
		}).
		Once().
		Return(&dpb.SendCoverageResponse{
			Message: dpb.SendCoverageResponse_PASS,
		}, nil)
	err := c.CompareCoverage(ctx)
	require.NoError(t, err)

	c.BaseRef = "new-branch-without-coverage"
	draftClient.
		On("SendCoverage", ctx, &dpb.SendCoverageRequest{
			Coverage:     float32(sampleCov),
			BranchName:   "feature-branch",
			Repository:   "manabie-com/backend",
			Key:          "123",
			TargetBranch: "new-branch-without-coverage",
			Integration:  false,
		}).Once().Return(nil, status.Error(codes.Unknown, "target branch does not exist"))
	err = c.CompareCoverage(ctx)
	require.NoError(t, err)
}
