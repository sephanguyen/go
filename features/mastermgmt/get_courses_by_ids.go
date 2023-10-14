package mastermgmt

import (
	"context"
	"errors"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"k8s.io/utils/strings/slices"
)

func (s *suite) getCoursesByIDs(ctx context.Context, idType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	cIDs := []string{}
	// existing and mixed cases
	if idType != "non-existing" {
		ids, err := s.seedCourses(ctx, 3)
		if err != nil {
			return nil, err
		}
		cIDs = ids

		stepState.SeedingCourseIDs = ids
	}
	if idType != "existing" {
		// non existing / mixed case
		cIDs = append(cIDs, idutil.ULIDNow())
	}

	req := &mpb.GetCoursesByIDsRequest{
		CourseIds: cIDs,
	}
	stepState.Response, stepState.ResponseErr = mpb.NewMasterDataCourseServiceClient(s.MasterMgmtConn).
		GetCoursesByIDs(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) mustReturnCorrectCourses(ctx context.Context, courseType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	res := stepState.Response.(*mpb.GetCoursesByIDsResponse)

	switch courseType {
	case "empty":
		{
			if len(res.Courses) != 0 {
				return StepStateToContext(ctx, stepState), errors.New("given non-existing ids, but received course(s)")
			}
		}
	case "valid", "mixed":
		{
			if len(res.Courses) != len(stepState.SeedingCourseIDs) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("given %d existing ids, but received %d course(s)", len(res.Courses), len(stepState.SeedingCourseIDs))
			}

			wrongItems := sliceutils.Filter(res.Courses, func(c *mpb.Course) bool {
				return !slices.Contains(stepState.SeedingCourseIDs, c.Id)
			})

			if len(wrongItems) > 0 {
				return StepStateToContext(ctx, stepState), errors.New("wrong items in case ids existing")
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) seedCourses(ctx context.Context, count int) ([]string, error) {
	courseIDs := make([]string, count)
	for j := 0; j < count; j++ {
		rdcID := idutil.ULIDNow()
		iStmt := `INSERT INTO courses (
					course_id,
					name,
					grade,
					updated_at,
					created_at)
				VALUES ($1, $2, $3, NOW(), NOW())`
		_, err := s.BobDB.Exec(ctx, iStmt, rdcID, rdcID+"name", 0)
		if err != nil {
			return nil, fmt.Errorf("cannot seed courses, err: %s", err)
		}

		courseIDs[j] = rdcID
	}
	return courseIDs, nil
}
