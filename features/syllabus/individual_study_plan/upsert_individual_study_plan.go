package individual_study_plan

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
	yasuoPb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func generateIndividualStudyPlanRequest(spID, lmID, studentID string) *sspb.UpsertIndividualInfoRequest {
	req := &sspb.UpsertIndividualInfoRequest{
		IndividualItems: []*sspb.StudyPlanItem{
			{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudyPlanId:        spID,
					LearningMaterialId: lmID,
					StudentId: &wrapperspb.StringValue{
						Value: studentID,
					},
				},
				Status: sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
			},
		},
	}

	return req
}

func generateUpsertStudyPlanRequest(bookID, courseID string) *epb.UpsertStudyPlanRequest {
	req := &epb.UpsertStudyPlanRequest{
		Name:                fmt.Sprintf("studyplan-%s", bookID),
		SchoolId:            constants.ManabieSchool,
		TrackSchoolProgress: true,
		Grades:              []int32{3, 4},
		Status:              epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
		BookId:              bookID,
		CourseId:            courseID,
	}

	return req
}

func generateUpsertCourseRequest(courseID string, schoolID int32) *yasuoPb.UpsertCoursesRequest {
	req := &yasuoPb.UpsertCoursesRequest{
		Courses: []*yasuoPb.UpsertCoursesRequest_Course{
			{
				Id:       courseID,
				Name:     "course",
				Country:  1,
				Subject:  bpb.SUBJECT_BIOLOGY,
				SchoolId: schoolID,
			},
		},
	}

	return req
}

func (s *Suite) ourSystemStoresIndividualStudyPlanCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	isp := &entities.IndividualStudyPlan{}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE study_plan_id = $1", strings.Join(database.GetFieldNames(isp), ","), isp.TableName())
	if err := database.Select(ctx, s.EurekaDB, query, stepState.StudyPlanID).ScanOne(isp); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	req := stepState.Request.(*sspb.UpsertIndividualInfoRequest)
	ispReq := req.GetIndividualItems()[0]

	if ispReq.StudyPlanItemIdentity.LearningMaterialId != isp.LearningMaterialID.String {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf(fmt.Sprintf("Incorrect learning material id, expected %v got %v", ispReq.StudyPlanItemIdentity.LearningMaterialId, isp.LearningMaterialID.String))
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminUpdateStartDateForIndividualStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := stepState.Request.(*sspb.UpsertIndividualInfoRequest)
	stepState.StartDate = timestamppb.Now()
	req.IndividualItems[0].StartDate = timestamppb.Now()

	if stepState.Response, stepState.ResponseErr = sspb.NewStudyPlanClient(s.EurekaConn).UpsertIndividual(s.AuthHelper.SignedCtx(ctx, stepState.Token), req); stepState.ResponseErr != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf(fmt.Sprintf("unable to upsert a individual study plan: %v", stepState.ResponseErr.Error()))
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemUpdatesStartDateForIndividualStudyPlanCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	isp := &entities.IndividualStudyPlan{}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE study_plan_id = $1", strings.Join(database.GetFieldNames(isp), ","), isp.TableName())
	if err := database.Select(ctx, s.EurekaDB, query, stepState.StudyPlanID).ScanOne(isp); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	req := stepState.Request.(*sspb.UpsertIndividualInfoRequest)
	var reqTime pgtype.Timestamptz
	reqTime.Set(req.GetIndividualItems()[0].StartDate.AsTime())

	if reqTime.Time.Equal(isp.StartDate.Time) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf(fmt.Sprintf("Incorrect start data, expected %v got %v", reqTime.Time, isp.StartDate.Time))
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aValidStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ctx, err1 := s.adminInsertCourse(ctx)
	ctx, err2 := s.adminInsertStudyPlan(ctx)
	if err := multierr.Combine(err1, err2); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a valid study plan: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminInsertIndividualStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.StudentID, _, _ = s.AuthHelper.AUserSignedInAsRole(ctx, "student")
	if len(stepState.LearningMaterialIDs) > 0 {
		stepState.LearningMaterialID = stepState.LearningMaterialIDs[0]
	}

	req := generateIndividualStudyPlanRequest(stepState.StudyPlanID, stepState.LearningMaterialID, stepState.StudentID)
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = sspb.NewStudyPlanClient(s.EurekaConn).UpsertIndividual(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminInsertCourse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.CourseID = idutil.ULIDNow()

	if _, err := yasuoPb.NewCourseServiceClient(s.YasuoConn).UpsertCourses(s.AuthHelper.SignedCtx(ctx, stepState.Token), generateUpsertCourseRequest(stepState.CourseID, constants.ManabieSchool)); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert course: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminInsertStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if resp, err := epb.NewCourseModifierServiceClient(s.EurekaConn).AddBooks(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.AddBooksRequest{
		CourseId: stepState.CourseID,
		BookIds:  []string{stepState.BookID},
	}); err != nil || !resp.Successful {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert course book: %w", err)
	}

	resp, err := epb.NewStudyPlanModifierServiceClient(s.EurekaConn).UpsertStudyPlan(s.AuthHelper.SignedCtx(ctx, stepState.Token), generateUpsertStudyPlanRequest(stepState.BookID, stepState.CourseID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan: %w", err)
	}
	stepState.StudyPlanID = resp.StudyPlanId
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) thereIsAFlashcardExistedInTopic(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	insertFlashcardtReq := &sspb.InsertFlashcardRequest{
		Flashcard: &sspb.FlashcardBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: stepState.TopicIDs[0],
				Name:    fmt.Sprintf("flashcard-name+%v", stepState.TopicIDs[0]),
			},
		},
	}
	resp, err := sspb.NewFlashcardClient(s.EurekaConn).InsertFlashcard(s.AuthHelper.SignedCtx(ctx, stepState.Token), insertFlashcardtReq)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	stepState.LearningMaterialIDs = append(stepState.LearningMaterialIDs, resp.GetLearningMaterialId())
	return utils.StepStateToContext(ctx, stepState), nil
}
