package mastermgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

func (s *suite) insertClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseIDs := stepState.CourseIDs
	locationIDs := stepState.CenterIDs
	fields := []string{"class_id", "name", "course_id", "location_id", "school_id", "created_at", "updated_at"}
	query := fmt.Sprintf("INSERT INTO class (%s) VALUES ($1,$2,$3,$4,$5,$6,$7)",
		strings.Join(fields, ","))
	var courseID, locationID string
	classID := idutil.ULIDNow()
	schoolID := golibs.ResourcePathFromCtx(ctx)
	className := "name"
	if len(courseIDs) > 0 {
		courseID = courseIDs[0]
	}
	if len(locationIDs) > 0 {
		locationID = locationIDs[0]
	}
	now := time.Now()
	stepState.RequestSentAt = now
	_, err := s.BobDBTrace.Exec(ctx, query, classID, className, courseID, locationID, schoolID, now, now)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	stepState.CurrentClassId = classID
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	name := "updated-name"
	stepState.NameOfData = name
	stepState.Request = &mpb.UpdateClassRequest{
		ClassId: stepState.CurrentClassId,
		Name:    name,
	}
	ctx, err := s.subscribeEventClass(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.subscribe: %w", err)
	}
	stepState.Response, stepState.ResponseErr = mpb.NewClassServiceClient(s.MasterMgmtConn).UpdateClass(contextWithToken(s, ctx), stepState.Request.(*mpb.UpdateClassRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updatedClassProperly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		name      string
		updatedAt time.Time
	)
	query := "SELECT name,updated_at FROM class WHERE class_id = $1 AND deleted_at IS NULL"
	err := s.BobDBTrace.QueryRow(ctx, query, stepState.CurrentClassId).Scan(&name, &updatedAt)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	if name != stepState.NameOfData {
		return StepStateToContext(ctx, stepState), fmt.Errorf("`name` is not correct")
	}
	if updatedAt.Equal(stepState.RequestSentAt) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("`updated_at` is not correct")
	}
	return StepStateToContext(ctx, stepState), nil
}
