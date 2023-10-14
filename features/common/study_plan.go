package common

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/helper"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/gogo/protobuf/types"
	"go.uber.org/multierr"
)

func (s *suite) aListOfValidPresetStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err1 := s.aValidPresetStudyPlan(ctx, "course-live-1-plan"+stepState.Random)
	ctx, err2 := s.aValidPresetStudyPlan(ctx, "course-live-2-plan"+stepState.Random)
	ctx, err3 := s.aValidPresetStudyPlan(ctx, "course-live-3-plan"+stepState.Random)
	ctx, err4 := s.aValidPresetStudyPlan(ctx, "course-live-teacher-1-plan"+stepState.Random)
	ctx, err5 := s.aValidPresetStudyPlan(ctx, "course-live-teacher-2-plan"+stepState.Random)
	ctx, err6 := s.aValidPresetStudyPlan(ctx, "course-live-teacher-3-plan"+stepState.Random)
	ctx, err7 := s.aValidPresetStudyPlan(ctx, "course-live-teacher-4-plan"+stepState.Random)
	ctx, err8 := s.aValidPresetStudyPlan(ctx, "course-live-teacher-5-plan"+stepState.Random)
	ctx, err9 := s.aValidPresetStudyPlan(ctx, "course-live-teacher-6-plan"+stepState.Random)
	ctx, err10 := s.aValidPresetStudyPlan(ctx, "course-live-teacher-7-plan"+stepState.Random)
	ctx, err11 := s.aValidPresetStudyPlan(ctx, "course-live-dont-have-lesson-1-plan"+stepState.Random)
	ctx, err12 := s.aValidPresetStudyPlan(ctx, "course-live-complete-lesson-1-plan"+stepState.Random)
	ctx, err13 := s.aValidPresetStudyPlan(ctx, "course-live-complete-lesson-2-plan"+stepState.Random)
	err := multierr.Combine(err1, err2, err3, err4, err5, err6, err7, err8, err9, err10, err11, err12, err13)
	return ctx, err
}
func (s *suite) aValidPresetStudyPlan(ctx context.Context, id string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	p := generatePresetStudyPlan()
	p.Id = id
	ctx, err := s.InsertAPresetStudyPlan(ctx, p)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), err
}

func generatePresetStudyPlan() *pb.PresetStudyPlan {
	num := rand.Int()
	startDate := "August 09"
	startTime := timeutil.PSPStartDate(pb.COUNTRY_VN, startDate)
	p := &pb.PresetStudyPlan{Id: fmt.Sprintf("%d", num), Name: "Random name", Country: pb.COUNTRY_VN, Grade: "G12", Subject: pb.SUBJECT_MATHS, CreatedAt: &types.Timestamp{Seconds: time.Now().Unix()}, UpdatedAt: &types.Timestamp{Seconds: time.Now().Unix()}, StartDate: &types.Timestamp{Seconds: startTime.Unix()}}
	return p
}

func (s *suite) InsertAPresetStudyPlan(ctx context.Context, p *pb.PresetStudyPlan) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, token := s.generateAnAdminToken(ctx)

	req := &pb.UpsertPresetStudyPlansRequest{
		PresetStudyPlans: []*pb.PresetStudyPlan{p},
	}

	_, err := pb.NewCourseClient(s.BobConn).UpsertPresetStudyPlans(helper.GRPCContext(ctx, "token", token), req)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) genPresetStudyPlanWeekly(ctx context.Context, startDate, endDate time.Time, lessonID, courseID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var topicID string
	if err := s.BobDB.QueryRow(ctx, "SELECT topic_id FROM topics LIMIT 1").Scan(&topicID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	var presetStudyPlanID string
	if err := s.BobDB.QueryRow(ctx, "SELECT preset_study_plan_id FROM courses c WHERE c.course_id =$1", courseID).Scan(&presetStudyPlanID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	presetStudyPlanWeekly := &bob_entities.PresetStudyPlanWeekly{}
	database.AllNullEntity(presetStudyPlanWeekly)

	week := rand.Intn(10000)
	err := multierr.Combine(
		presetStudyPlanWeekly.ID.Set(s.newID()),
		presetStudyPlanWeekly.StartDate.Set(startDate),
		presetStudyPlanWeekly.EndDate.Set(endDate),
		presetStudyPlanWeekly.PresetStudyPlanID.Set(presetStudyPlanID),
		presetStudyPlanWeekly.TopicID.Set(topicID),
		presetStudyPlanWeekly.Week.Set(week),
		presetStudyPlanWeekly.CreatedAt.Set(timeutil.Now()),
		presetStudyPlanWeekly.UpdatedAt.Set(timeutil.Now()),
		presetStudyPlanWeekly.LessonID.Set(lessonID),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	cmdTag, err := database.Insert(ctx, presetStudyPlanWeekly, s.BobDB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if cmdTag.RowsAffected() != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert preset study plan weekly")
	}
	return StepStateToContext(ctx, stepState), nil
}
