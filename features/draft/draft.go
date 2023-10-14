package draft

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/draft/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/draft/v1"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var createTestCases = []struct {
	branch            string
	repository        string
	integration       bool
	expectCreateError bool
}{
	// good values
	{
		repository:        "github.com/manabie-com/test",
		branch:            "develop",
		integration:       false,
		expectCreateError: false,
	},
	// integration true
	{
		repository:        "github.com/manabie-com/test",
		branch:            "develop",
		integration:       true,
		expectCreateError: false,
	},
	// different branch
	{
		repository:        "github.com/manabie-com/test",
		branch:            "release",
		integration:       false,
		expectCreateError: false,
	},
	// different repository
	{
		repository:        "github.com/manabie-com/test-two",
		branch:            "develop",
		integration:       false,
		expectCreateError: false,
	},
	// Test duplicate create should fail
	{
		repository:        "github.com/manabie-com/test",
		branch:            "develop",
		integration:       false,
		expectCreateError: true,
	},
	// Test duplicate create should fail
	{
		repository:        "github.com/manabie-com/test",
		branch:            "develop",
		integration:       true,
		expectCreateError: true,
	},
	// Test duplicate create should fail
	{
		repository:        "github.com/manabie-com/test",
		branch:            "release",
		integration:       false,
		expectCreateError: true,
	},
}

// channels to force things to run in the order I want :p
var createdDone = make(chan struct{})
var updatedDone = make(chan struct{})

// table-test error wrap
func ttErr(k int, err error) error {
	return errors.Wrapf(err, "test[%d]", k)
}

func (s *suite) createCoverageTest() error {
	defer close(createdDone)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// clear database

	truncateStmt := fmt.Sprintf(
		"TRUNCATE %s; TRUNCATE %s;",
		(&entities.History{}).TableName(),
		(&entities.TargetCoverage{}).TableName(),
	)
	_, err := s.DB.Exec(ctx, truncateStmt)
	if err != nil {
		return fmt.Errorf("clear database error: %s", err)
	}

	c := pb.NewSendCoverageServiceClient(s.Conn)

	for k, tC := range createTestCases {
		_, err := c.CreateTargetCoverage(ctx, &pb.CreateTargetCoverageRequest{
			Coverage:    50.0,
			Key:         "AAAA",
			BranchName:  tC.branch,
			Repository:  tC.repository,
			Integration: tC.integration,
		})
		if (err != nil) != tC.expectCreateError {
			return ttErr(k, errors.Wrapf(err, "expectCreateError was %t, but had error", tC.expectCreateError))
		}
		if tC.expectCreateError {
			continue
		}

		targetCoverage := entities.TargetCoverage{}
		fields, values := targetCoverage.FieldMap()

		query := fmt.Sprintf(
			"SELECT %s FROM %s WHERE branch_name=$1 AND repository=$2 AND integration=$3 LIMIT 1", strings.Join(fields, ","), targetCoverage.TableName())
		// fmt.Println(k, query)
		row := s.DB.QueryRow(ctx, query, tC.branch, tC.repository, tC.integration)
		err = row.Scan(values...)
		if err != nil {
			return ttErr(k, errors.Wrap(err, "row.Scan error"))
		}
	}
	return nil
}

var updateTestCases = []struct {
	newCoverage float32
}{
	// update all created coverages...
	{
		newCoverage: 60.0,
	},
	{
		newCoverage: 61.0,
	},
	{
		newCoverage: 63.0,
	},
	{
		newCoverage: 62.0,
	},
}

func (s *suite) updateCoverageTest() error {
	<-createdDone
	defer close(updatedDone)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := pb.NewSendCoverageServiceClient(s.Conn)

	for k, tC := range updateTestCases {
		_, err := c.UpdateTargetCoverage(ctx, &pb.UpdateTargetCoverageRequest{
			Coverage:    tC.newCoverage,
			Key:         "AAAA",
			BranchName:  createTestCases[k].branch,
			Repository:  createTestCases[k].repository,
			Integration: createTestCases[k].integration,
		})
		if err != nil {
			return ttErr(k, err)
		}

		targetCoverage := entities.TargetCoverage{}
		fields, values := targetCoverage.FieldMap()

		query := fmt.Sprintf(
			"SELECT %s FROM %s WHERE branch_name=$1 AND repository=$2 AND integration=$3 LIMIT 1", strings.Join(fields, ","), targetCoverage.TableName())
		row := s.DB.QueryRow(ctx, query, createTestCases[k].branch, createTestCases[k].repository, createTestCases[k].integration)

		// fmt.Println("update", k, query)
		err = row.Scan(values...)
		if err != nil {
			return ttErr(k, fmt.Errorf("case: %v row.Scan: %v ", tC, err))
		}
		if targetCoverage.Coverage != tC.newCoverage {
			return ttErr(k, fmt.Errorf("coverage should be %f, but was %f", tC.newCoverage, targetCoverage.Coverage))
		}
	}
	return nil
}

func (s *suite) compareCoverageTest() error {
	<-updatedDone

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := pb.NewSendCoverageServiceClient(s.Conn)

	testCases := []struct {
		branch string
	}{
		{
			branch: "feature/LT-1337",
		},
		{
			branch: "feature/LT-1338",
		},
	}

	// checks last history entry for branch to see if it was pass or fail
	requireHistory := func(targetK int, branch string, pass bool) error {
		history := entities.History{}
		fields, values := history.FieldMap()
		query := fmt.Sprintf("SELECT %s FROM %s WHERE branch_name=$1 AND repository=$2 AND integration=$3 ORDER BY id DESC LIMIT 1", strings.Join(fields, ","), history.TableName())
		row := s.DB.QueryRow(ctx, query, branch, createTestCases[targetK].repository, createTestCases[targetK].integration)
		err := row.Scan(values...)
		if err != nil {
			return ttErr(targetK, fmt.Errorf("error querying for history. row.Scan: %s", err))
		}
		if pass {
			if history.Status != pb.SendCoverageResponse_PASS.String() {
				return ttErr(targetK, fmt.Errorf("coverage history status should be pass, but was %s", history.Status))
			}
		} else {
			if history.Status != pb.SendCoverageResponse_FAIL.String() {
				return ttErr(targetK, fmt.Errorf("coverage history status should be fail, but was %s", history.Status))
			}
		}
		return nil
	}

	for _, tC := range testCases {
		for targetK, targetC := range updateTestCases {
			// Pass with greater coverage
			_, err := c.SendCoverage(ctx, &pb.SendCoverageRequest{
				Coverage:     targetC.newCoverage + 10,
				Key:          "AAAA",
				TargetBranch: createTestCases[targetK].branch,
				Repository:   createTestCases[targetK].repository,
				Integration:  createTestCases[targetK].integration,
				BranchName:   tC.branch,
			})
			if err != nil {
				return ttErr(targetK, fmt.Errorf("should not return error for increased coverage: %s", err))
			}
			if err = requireHistory(targetK, tC.branch, true); err != nil {
				return err
			}

			// Pass with equal coverage
			_, err = c.SendCoverage(ctx, &pb.SendCoverageRequest{
				Coverage:     targetC.newCoverage,
				Key:          "AAAA",
				TargetBranch: createTestCases[targetK].branch,
				Repository:   createTestCases[targetK].repository,
				Integration:  createTestCases[targetK].integration,
				BranchName:   tC.branch,
			})
			if err != nil {
				return ttErr(targetK, fmt.Errorf("should not return error for equal coverage: %s", err))
			}
			if err = requireHistory(targetK, tC.branch, true); err != nil {
				return err
			}

			// Fail with lower coverage
			_, err = c.SendCoverage(ctx, &pb.SendCoverageRequest{
				Coverage:     targetC.newCoverage - 10,
				Key:          "AAAA",
				TargetBranch: createTestCases[targetK].branch,
				Repository:   createTestCases[targetK].repository,
				Integration:  createTestCases[targetK].integration,
				BranchName:   tC.branch,
			})
			if err == nil || status.Code(err) != codes.FailedPrecondition {
				return ttErr(targetK, fmt.Errorf("should return FailedPrecondition for lower coverage"))
			}
			if err = requireHistory(targetK, tC.branch, false); err != nil {
				return err
			}
		}
	}

	return nil
}
