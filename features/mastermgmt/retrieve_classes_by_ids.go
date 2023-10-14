package mastermgmt

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

func getInputClasses() []*mpb.RetrieveClassByIDsResponse_Class {
	return []*mpb.RetrieveClassByIDsResponse_Class{
		{ClassId: "class-id-1", Name: "class-1"},
		{ClassId: "class-id-2", Name: "class-2"},
		{ClassId: "class-id-3", Name: "class-3"},
	}
}

func (s *suite) aListOfClassInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.Random = fmt.Sprint(rand.Intn(2999) + 1)
	classList := getInputClasses()
	stepState.RequestSentAt = time.Now()

	for _, cl := range classList {
		courseIDs := stepState.CourseIDs
		locationIDs := stepState.CenterIDs
		fields := []string{"class_id", "name", "course_id", "location_id", "school_id", "created_at", "updated_at"}
		query := fmt.Sprintf("INSERT INTO class (%s) VALUES ($1,$2,$3,$4,$5,$6,$7)",
			strings.Join(fields, ","))
		var courseID, locationID string

		classID := cl.ClassId + "-" + s.Random
		schoolID := golibs.ResourcePathFromCtx(ctx)
		className := cl.Name
		if len(courseIDs) > 0 {
			courseID = courseIDs[0]
		}
		if len(locationIDs) > 0 {
			locationID = locationIDs[0]
		}

		_, err := s.BobDBTrace.Exec(ctx, query, classID, className, courseID, locationID, schoolID, stepState.RequestSentAt, stepState.RequestSentAt)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}
		s.ClassIds = append(s.ClassIds, classID)
	}
	s.CurrentClassId = s.ClassIds[0]

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) retrieveClassesByIDs(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &mpb.RetrieveClassByIDsRequest{
		ClassIds: s.ClassIds,
	}

	stepState.Response, stepState.ResponseErr = mpb.NewClassServiceClient(s.MasterMgmtConn).RetrieveClassesByIDs(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) mustReturnCorrectClasses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*mpb.RetrieveClassByIDsResponse)
	res := rsp.GetClasses()
	req := getInputClasses()

	if len(res) != len(req) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("length classes of response are incorrect")
	}
	for _, v := range res {
		if s.checkIsNotEqualClasses(req, v.ClassId, v.Name) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("some ids or names of classes in response are incorrect")
		}
		locationID := stepState.CenterIDs[0]
		if v.LocationId != locationID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected location %s but got %s", locationID, v.LocationId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkIsNotEqualClasses(cs []*mpb.RetrieveClassByIDsResponse_Class, classID string, className string) bool {
	for _, v := range cs {
		classIDWithRandom := v.ClassId + "-" + s.Random
		if classIDWithRandom == classID && v.Name == className {
			return false
		}
	}
	return true
}
