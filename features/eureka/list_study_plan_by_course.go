package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) aUserSignedInTeacher(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.UserId = idutil.ULIDNow()
	ctx, err := s.aValidToken(ctx, "USER_GROUP_TEACHER")
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to generate a valid token: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) AUserSignedInTeacher(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aUserSignedInTeacher(ctx)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) listStudyPlanByCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	paging := &cpb.Paging{
		Limit: uint32(rand.Int31n(3)) + 2,
	}
	for {
		res, err := pb.NewStudyPlanReaderServiceClient(s.Conn).ListStudyPlanByCourse(contextWithToken(s, ctx), &pb.ListStudyPlanByCourseRequest{
			CourseId: stepState.CourseID,
			Paging:   paging,
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("ListStudyPlanByCourse: %w", err)
		}
		if len(res.StudyPlans) == 0 {
			break
		}
		stepState.StudyPlans = append(stepState.StudyPlans, res.StudyPlans...)
		paging = res.NextPage
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) avalidCourseAndSomeStudyPlanBackground(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err1 := s.aValidCourseBackground(ctx)
	ctx, err2 := s.someStudyPlanNameInDB(ctx)
	err := multierr.Combine(
		err1, err2,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) someStudyPlanNameInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	n := rand.Intn(15) + 10
	if stepState.AuthToken == "" {
		ctx, err := s.aValidToken(ctx, "USER_GROUP_TEACHER")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	ctx, studyPlans, courseStudyPlans := s.generateSomeStudyPlans(ctx, n)

	for _, req := range studyPlans {
		_, err := pb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlan(contextWithToken(s, ctx), req)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create some study plans: %w", err)
		}
	}

	courseStudyPlanRepo := &repositories.CourseStudyPlanRepo{}
	if err := courseStudyPlanRepo.BulkUpsert(ctx, s.DB, courseStudyPlans); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("courseStudyPlanRepo.BulkUpsert: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) teacherArchivesSomeStudyPlans(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	numSp := len(stepState.StudyPlanIDs)
	n := rand.Intn(numSp)
	for i := 0; i < n; i++ {
		req := &pb.UpsertStudyPlanRequest{
			StudyPlanId: &wrapperspb.StringValue{
				Value: stepState.StudyPlanIDs[i],
			},
			Status: pb.StudyPlanStatus_STUDY_PLAN_STATUS_ARCHIVED,
		}
		_, err := pb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlan(contextWithToken(s, ctx), req)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan: %w", err)
		}
		stepState.ArchivedStudyPlanIDs = append(stepState.ArchivedStudyPlanIDs, stepState.StudyPlanIDs[i])
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateSomeStudyPlans(ctx context.Context, n int, courseID ...interface{}) (context.Context, []*pb.UpsertStudyPlanRequest, []*entities.CourseStudyPlan) {
	stepState := StepStateFromContext(ctx)
	res := make([]*pb.UpsertStudyPlanRequest, 0, n)
	es := make([]*entities.CourseStudyPlan, 0, n)

	var courseId string

	if len(courseID) > 0 {
		if str := courseID[0].(string); str != "" {
			courseId = str

			upserts := []*pb.UpsertStudyPlanRequest{
				{
					StudyPlanId: &wrapperspb.StringValue{Value: "01F4XATGGDRS006EDEW9X6VAXA"},
					SchoolId:    constants.ManabieSchool,
					Name:        "14-IJHDFULPHC-オハニセカサイツァウ",
					CourseId:    courseId,
				},
				{
					StudyPlanId: &wrapperspb.StringValue{Value: "01F4XB0KJF73E4180EF3D05KVW"},
					SchoolId:    constants.ManabieSchool,
					Name:        "2-AINGQXECQG-バラマジヤモミオメジ",
					CourseId:    courseId,
				},
				{
					StudyPlanId: &wrapperspb.StringValue{Value: "01F4XB0KJEQNJPDFC1J34TEA27"},
					SchoolId:    constants.ManabieSchool,
					Name:        "37-VCHIAYNUWW-モサゥジウチロギミナ",
					CourseId:    courseId,
				},
				{
					StudyPlanId: &wrapperspb.StringValue{Value: "01F4XATGGF9BDE1P9BAY7X70C2"},
					SchoolId:    constants.ManabieSchool,
					Name:        "1-TFCBAAFNWO-フガズデルニワドャポ",
					CourseId:    courseId,
				},
				{
					StudyPlanId: &wrapperspb.StringValue{Value: "01F4XATGGEP7N5ZVJA9H6FDJVR"},
					SchoolId:    constants.ManabieSchool,
					Name:        "83-MJNULARUJM-ユモヅヲクシフユサシ",
					CourseId:    courseId,
				},
			}

			for _, e := range upserts {
				stepState.StudyPlanIDs = append(stepState.StudyPlanIDs, e.StudyPlanId.Value)
				res = append(res, e)
				e := &entities.CourseStudyPlan{
					CourseID:    database.Text(courseId),
					StudyPlanID: database.Text(e.StudyPlanId.Value),
				}
				e.BaseEntity.Now()
				es = append(es, e)
			}
		}
	} else {
		courseId = stepState.CourseID

		for i := 0; i < n; i++ {
			id := idutil.ULIDNow()
			stepState.StudyPlanIDs = append(stepState.StudyPlanIDs, id)
			res = append(res, &pb.UpsertStudyPlanRequest{
				StudyPlanId: &wrapperspb.StringValue{Value: id},
				SchoolId:    constants.ManabieSchool,
				Name: fmt.Sprintf(
					"%s-%s-%s",
					strconv.Itoa(int(helper.RandInt(1, 100))),
					helper.GenerateRandomRune(10, helper.English[0], helper.English[1]),
					helper.GenerateRandomRune(10, helper.JapaneseHiragana[0], helper.JapaneseHiragana[1]),
				),
				CourseId: courseId,
			})
			e := &entities.CourseStudyPlan{
				CourseID:    database.Text(courseId),
				StudyPlanID: database.Text(id),
			}
			e.BaseEntity.Now()
			es = append(es, e)
		}
	}

	return StepStateToContext(ctx, stepState), res, es
}

func (s *suite) ourSystemHaveToReturnListStudyPlanByCourseCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if (len(stepState.StudyPlanIDs) - len(stepState.ArchivedStudyPlanIDs)) != len(stepState.StudyPlans) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected number of study plans, expected: %d, actual: %d", len(stepState.StudyPlanIDs), len(stepState.StudyPlans))
	}

	return StepStateToContext(ctx, stepState), nil
}
