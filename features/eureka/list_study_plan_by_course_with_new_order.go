package eureka

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) addNewStudyPlans(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, studyPlans, courseStudyPlans := s.generateSomeStudyPlans(ctx, 10, "THANHMSCRN1NRRC3BE6X1XDINH")

	for _, req := range studyPlans {
		if _, err := pb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlan(contextWithToken(s, ctx), req); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create some study plans: %w", err)
		}
	}

	courseStudyPlanRepo := &repositories.CourseStudyPlanRepo{}
	if err := courseStudyPlanRepo.BulkUpsert(ctx, s.DB, courseStudyPlans); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("courseStudyPlanRepo.BulkUpsert: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) allStudyPlansWereInserted(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var count int64
	var count2 int64

	query := `select count(*) from study_plans where course_id = $1;`
	if err := s.DB.QueryRow(ctx, query, "THANHMSCRN1NRRC3BE6X1XDINH").Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	query = `select count(*) from course_study_plans where course_id = $1;`
	if err := s.DB.QueryRow(ctx, query, "THANHMSCRN1NRRC3BE6X1XDINH").Scan(&count2); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if count != count2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("data was failure inserted")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetListStudyPlansAndFilterWithNewOrderCollation(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	pbStudyPlans := make([]*pb.StudyPlan, 0)
	paging := &cpb.Paging{
		Limit: 5,
	}
	res, err := pb.NewStudyPlanReaderServiceClient(s.Conn).ListStudyPlanByCourse(contextWithToken(s, ctx), &pb.ListStudyPlanByCourseRequest{
		CourseId: "THANHMSCRN1NRRC3BE6X1XDINH",
		Paging:   paging,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	pbStudyPlans = append(pbStudyPlans, res.StudyPlans...)

	sortResult := []*pb.UpsertStudyPlansRequest_StudyPlan{
		{
			StudyPlanId: &wrapperspb.StringValue{Value: "01F4XATGGF9BDE1P9BAY7X70C2"},
			SchoolId:    constants.ManabieSchool,
			Country:     cpb.Country_COUNTRY_JP,
			Name:        "1-TFCBAAFNWO-フガズデルニワドャポ",
			Type:        pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE,
			CourseId:    "THANHMSCRN1NRRC3BE6X1XDINH",
		},
		{
			StudyPlanId: &wrapperspb.StringValue{Value: "01F4XB0KJF73E4180EF3D05KVW"},
			SchoolId:    constants.ManabieSchool,
			Country:     cpb.Country_COUNTRY_JP,
			Name:        "2-AINGQXECQG-バラマジヤモミオメジ",
			Type:        pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE,
			CourseId:    "THANHMSCRN1NRRC3BE6X1XDINH",
		},
		{
			StudyPlanId: &wrapperspb.StringValue{Value: "01F4XATGGDRS006EDEW9X6VAXA"},
			SchoolId:    constants.ManabieSchool,
			Country:     cpb.Country_COUNTRY_JP,
			Name:        "14-IJHDFULPHC-オハニセカサイツァウ",
			Type:        pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE,
			CourseId:    "THANHMSCRN1NRRC3BE6X1XDINH",
		},

		{
			StudyPlanId: &wrapperspb.StringValue{Value: "01F4XB0KJEQNJPDFC1J34TEA27"},
			SchoolId:    constants.ManabieSchool,
			Country:     cpb.Country_COUNTRY_JP,
			Name:        "37-VCHIAYNUWW-モサゥジウチロギミナ",
			Type:        pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE,
			CourseId:    "THANHMSCRN1NRRC3BE6X1XDINH",
		},
		{
			StudyPlanId: &wrapperspb.StringValue{Value: "01F4XATGGEP7N5ZVJA9H6FDJVR"},
			SchoolId:    constants.ManabieSchool,
			Country:     cpb.Country_COUNTRY_JP,
			Name:        "83-MJNULARUJM-ユモヅヲクシフユサシ",
			Type:        pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE,
			CourseId:    "THANHMSCRN1NRRC3BE6X1XDINH",
		},
	}

	for i := 0; i < len(pbStudyPlans)-1; i++ {
		if pbStudyPlans[i].Name != sortResult[i].Name {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong order, expected: %s-> %s, actual: %s->%s", pbStudyPlans[i], pbStudyPlans[i+1], pbStudyPlans[i+1], pbStudyPlans[i])
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
