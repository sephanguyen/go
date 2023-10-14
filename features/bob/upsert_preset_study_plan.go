package bob

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	types "github.com/gogo/protobuf/types"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func generatePresetStudyPlan() *pb.PresetStudyPlan {
	num := rand.Int()
	startDate := "August 09"
	startTime := timeutil.PSPStartDate(pb.COUNTRY_VN, startDate)
	p := &pb.PresetStudyPlan{Id: fmt.Sprintf("%d", num), Name: "Random name", Country: pb.COUNTRY_VN, Grade: "G12", Subject: pb.SUBJECT_MATHS, CreatedAt: &types.Timestamp{Seconds: time.Now().Unix()}, UpdatedAt: &types.Timestamp{Seconds: time.Now().Unix()}, StartDate: &types.Timestamp{Seconds: startTime.Unix()}}
	return p
}
func (s *suite) validPresetStudyPlan(ctx context.Context, presetNum int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	sp := make([]*pb.PresetStudyPlan, 0, presetNum)
	for i := 0; i < presetNum; i++ {
		sp = append(sp, generatePresetStudyPlan())
	}
	stepState.Request = &pb.UpsertPresetStudyPlansRequest{
		PresetStudyPlans: sp,
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validPresetStudyPlanWithoutId(ctx context.Context, presetNum int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	sp := make([]*pb.PresetStudyPlan, 0, presetNum)
	for i := 0; i < presetNum; i++ {
		presetStudyPlan := generatePresetStudyPlan()
		presetStudyPlan.Id = ""
		sp = append(sp, presetStudyPlan)
	}
	stepState.Request = &pb.UpsertPresetStudyPlansRequest{
		PresetStudyPlans: sp,
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userUpsertPresetStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewCourseClient(s.Conn).UpsertPresetStudyPlans(s.signedCtx(ctx), stepState.Request.(*pb.UpsertPresetStudyPlansRequest))
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobMustStoreAllPresetStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}
	query := `SELECT preset_study_plan_id, name, country, grade, subject, updated_at, created_at
	FROM public.preset_study_plans WHERE preset_study_plan_id=$1 ORDER BY created_at ASC;`
	returnedIds := stepState.Response.(*pb.UpsertPresetStudyPlansResponse).PresetStudyPlanIds
	studyPlans := stepState.Request.(*pb.UpsertPresetStudyPlansRequest).PresetStudyPlans
	for i, studyPlan := range studyPlans {
		if studyPlan.Id == "" {
			//return id is same order as requested preset study plan
			studyPlan.Id = returnedIds[i]
		}
		row := s.DB.QueryRow(ctx, query, &studyPlan.Id)
		eStudyPlan := new(entities_bob.PresetStudyPlan)
		row.Scan(&eStudyPlan.ID, &eStudyPlan.Name, &eStudyPlan.Country, &eStudyPlan.Grade, &eStudyPlan.Subject, &eStudyPlan.UpdatedAt, &eStudyPlan.CreatedAt)
		if !isEqualPresetStudyPlanEnAndPb(eStudyPlan, studyPlan) {
			return StepStateToContext(ctx, stepState), errors.New("bob store wrong preset study plan information")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsAListOfStoredStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ids := stepState.Response.(*pb.UpsertPresetStudyPlansResponse).PresetStudyPlanIds
	studyPlans := stepState.Request.(*pb.UpsertPresetStudyPlansRequest).PresetStudyPlans
	if len(ids) != len(studyPlans) {
		return StepStateToContext(ctx, stepState), errors.New("bob did not return all stored study plan")
	}
	return StepStateToContext(ctx, stepState), nil
}

func isEqualPresetStudyPlanEnAndPb(e *entities_bob.PresetStudyPlan, p *pb.PresetStudyPlan) bool {
	if p.UpdatedAt == nil {
		p.UpdatedAt = &types.Timestamp{Seconds: e.UpdatedAt.Time.Unix()}
	}
	if p.CreatedAt == nil {
		p.CreatedAt = &types.Timestamp{Seconds: e.CreatedAt.Time.Unix()}
	}
	updatedAt := &types.Timestamp{Seconds: e.UpdatedAt.Time.Unix()}
	createdAt := &types.Timestamp{Seconds: e.CreatedAt.Time.Unix()}

	grade := "G" + strconv.FormatInt(int64(e.Grade.Int), 10)
	return (e.ID.String == p.Id) && (e.Name.String == p.Name) && (pb.Country(pb.Country_value[e.Country.String]) == p.Country) && (grade == p.Grade) && (pb.Subject(pb.Subject_value[e.Subject.String]) == p.Subject) && updatedAt.Equal(p.UpdatedAt) && createdAt.Equal(p.CreatedAt)

}
