package eureka

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	eureka "github.com/manabie-com/backend/internal/eureka/constants"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) AValidToken(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.UserId = idutil.ULIDNow()
	ctx, err := s.aValidToken(ctx, userGroup)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) aValidToken(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.UserId = idutil.ULIDNow()
	var oldGroup string
	switch userGroup {
	case "teacher", "current teacher":
		userGroup = constant.RoleTeacher
		oldGroup = entities.UserGroupTeacher
	case "student":
		userGroup = constant.RoleStudent
		oldGroup = entities.UserGroupStudent
	case "parent":
		userGroup = constant.RoleParent
		oldGroup = entities.UserGroupParent
	case "school admin", "admin":
		userGroup = constant.RoleSchoolAdmin
		oldGroup = entities.UserGroupSchoolAdmin
	case "hq staff":
		userGroup = eureka.RoleHQStaff
		oldGroup = entities.UserGroupParent
		/*
			TODO: we'll change belows roles correctly when user team adds them
			For now, just using "constant.RoleParent" temporary instead
		*/
	case "center lead", "center manager", "center staff":
		userGroup = constant.RoleParent
		oldGroup = entities.UserGroupParent
	default:
		userGroup = constant.RoleStudent
		oldGroup = entities.UserGroupStudent
	}
	ctx, err := s.aValidUser(ctx, stepState.UserId, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUserInDB: %w", err)
	}
	token, err := s.generateExchangeToken(stepState.UserId, oldGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateExchangeToken: %w", err)
	}
	stepState.AuthToken = token

	switch userGroup {
	case constant.RoleSchoolAdmin:
		stepState.SchoolAdminToken = stepState.AuthToken
	case constant.RoleStudent:
		stepState.StudentToken = stepState.AuthToken
	case constant.RoleTeacher:
		stepState.TeacherToken = stepState.AuthToken
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eurekaMustStoreCorrectStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.UpsertStudyPlanResponse)

	query := "SELECT count(*) FROM study_plans WHERE study_plan_id = $1"
	var count int32
	if err := s.DB.QueryRow(ctx, query, rsp.StudyPlanId).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if int(count) != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("can not find study plan")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreateStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.aSignedIn(ctx, "school admin")
	ctx = auth.InjectFakeJwtToken(ctx, stepState.SchoolID)
	ctx = contextWithToken(s, ctx)
	if stepState.BookID == "" {
		if ctx, err := s.createBook(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create book: %w", err)
		}
	}
	if stepState.CourseID == "" {
		if ctx, err := s.createACourse(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create course: %w", err)
		}
	}

	req := &pb.UpsertStudyPlanRequest{
		SchoolId:            constants.ManabieSchool,
		Name:                idutil.ULIDNow(),
		CourseId:            stepState.CourseID,
		BookId:              stepState.BookID,
		TrackSchoolProgress: true,
		Grades:              []int32{3, 4},
		Status:              pb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
	}
	stepState.Request = req

	stepState.Response, stepState.ResponseErr = pb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlan(s.signedCtx(ctx), req)
	resp := stepState.Response.(*pb.UpsertStudyPlanResponse)
	stepState.StudyPlanID = resp.StudyPlanId
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) listPaginatedStudyPlans(ctx context.Context, studentID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var studyPlans [][]*pb.StudyPlan
	var offset string

	for {
		limit := rand.Intn(2) + 1
		resp, err := pb.NewAssignmentReaderServiceClient(s.Conn).ListStudyPlans(contextWithToken(s, ctx), &pb.ListStudyPlansRequest{
			StudentId: studentID,
			SchoolId:  constants.ManabieSchool,
			CourseId:  stepState.CourseID,
			Paging: &cpb.Paging{
				Limit: uint32(limit),
				Offset: &cpb.Paging_OffsetString{
					OffsetString: offset,
				},
			},
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if len(resp.Items) == 0 {
			break
		}
		if len(resp.Items) > limit {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total study plans: got: %d, want: %d", len(resp.Items), limit)
		}

		studyPlans = append(studyPlans, resp.Items)

		offset = resp.NextPage.GetOffsetString()
	}

	stepState.PaginatedStudyPlans = append(stepState.PaginatedStudyPlans, studyPlans)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) teacherListStudyPlansForEachStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	teacherID := idutil.ULIDNow()

	if _, err := s.aValidUser(ctx, teacherID, constant.RoleTeacher); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create teacher: %w", err)
	}
	token, err := s.generateExchangeToken(teacherID, entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = token

	for _, studentID := range stepState.StudentIDs {
		if ctx, err := s.listPaginatedStudyPlans(ctx, studentID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsAListOfAssignedStudyPlansOfEachStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for i, studentID := range stepState.StudentIDs {
		ctx, expectedPlans, err := s.getStudyPlansByStudent(ctx, studentID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		var total int

		paginatedPlans := stepState.PaginatedStudyPlans[i]
		for _, plans := range paginatedPlans {
			for _, plan := range plans {
				if !golibs.InArrayString(plan.StudyPlanId, expectedPlans) {
					return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected study plan id: %q in list study plans %v of student: %q", plan.StudyPlanId, expectedPlans, studentID)
				}
				total++
			}
		}
		if total != len(expectedPlans) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("total assigned study plans don't match: got: %d, expected: %d", total, len(expectedPlans))
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
