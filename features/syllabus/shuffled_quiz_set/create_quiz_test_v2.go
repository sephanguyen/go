package shuffled_quiz_set

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/bob/constants"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/s3"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
	ypb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) schoolAdminAddStudentToACourseHaveAStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Token = stepState.SchoolAdmin.Token
	ctx, err := s.userCreateACourseWithAStudyPlan(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.userCreateACourseWithAStudyPlan: %w", err)
	}
	ctx, err = s.userAddCourseToStudent(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.addStudentToCourse: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreateQuizTestV2(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Token = stepState.Student.Token
	stepState.TopicIDs = append(stepState.TopicIDs, stepState.TopicID)
	stepState.Response, stepState.ResponseErr = sspb.NewQuizClient(s.EurekaConn).CreateQuizTestV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.CreateQuizTestV2Request{
		StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
			StudyPlanId:        stepState.StudyPlanID,
			LearningMaterialId: stepState.LoID,
			StudentId:          wrapperspb.String(stepState.Student.ID),
		},
		Paging: &cpb.Paging{
			Limit: 10,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
		KeepOrder: true,
	})
	if stepState.ResponseErr == nil {
		resp := stepState.Response.(*sspb.CreateQuizTestV2Response)
		stepState.ShuffledQuizSetID = resp.ShuffleQuizSetId
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreateACourseWithAStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if ctx, err := s.createACourse(ctx); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	req := &epb.UpsertStudyPlanRequest{
		Name:                "study-plan",
		CourseId:            stepState.CourseID,
		BookId:              stepState.BookID,
		SchoolId:            constants.ManabieSchool,
		Status:              epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
		TrackSchoolProgress: true,
		Grades:              []int32{1, 2},
	}
	resp, err := epb.NewStudyPlanModifierServiceClient(s.EurekaConn).UpsertStudyPlan(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create study plan: %w", err)
	}
	stepState.Request = req
	stepState.StudyPlanID = resp.StudyPlanId
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createACourse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.CourseID = idutil.ULIDNow()
	_, err := ypb.NewCourseServiceClient(s.YasuoConn).UpsertCourses(s.AuthHelper.SignedCtx(ctx, stepState.Token), &ypb.UpsertCoursesRequest{
		Courses: []*ypb.UpsertCoursesRequest_Course{
			{
				Id:       stepState.CourseID,
				Name:     "course",
				Country:  1,
				Subject:  bpb.SUBJECT_BIOLOGY,
				SchoolId: int32(constants.ManabieSchool),
				BookIds:  []string{stepState.BookID},
			},
		},
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create course: %w", err)
	}

	if _, err := epb.NewCourseModifierServiceClient(s.EurekaConn).AddBooks(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.AddBooksRequest{
		BookIds:  []string{stepState.BookID},
		CourseId: stepState.CourseID,
	}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to add books: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userAddCourseToStudent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stmt := `SELECT email FROM users WHERE user_id = $1`
	var studentEmail string
	err := s.BobDB.QueryRow(ctx, stmt, stepState.Student.ID).Scan(&studentEmail)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.BobDB.QueryRow: %w", err)
	}

	if stepState.SchoolAdmin.Token != "" {
		stepState.Token = stepState.SchoolAdmin.Token
	}

	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	locationID := idutil.ULIDNow()
	e := &bob_entities.Location{}
	database.AllNullEntity(e)
	if err := multierr.Combine(
		e.LocationID.Set(locationID),
		e.Name.Set(fmt.Sprintf("location-%s", locationID)),
		e.IsArchived.Set(false),
		e.CreatedAt.Set(time.Now()),
		e.UpdatedAt.Set(time.Now()),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup Location: %w", err)
	}

	if _, err := database.Insert(ctx, e, s.BobDB.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert Location: %w", err)
	}

	_, err = upb.NewUserModifierServiceClient(s.UserMgmtConn).UpdateStudent(
		ctx,
		&upb.UpdateStudentRequest{
			StudentProfile: &upb.UpdateStudentRequest_StudentProfile{
				Id:               stepState.Student.ID,
				Name:             "test-name",
				Grade:            5,
				EnrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
				Email:            studentEmail,
				LocationIds:      []string{locationID},
			},

			SchoolId: stepState.SchoolIDInt,
		},
	)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to update student: %w", err)
	}

	if _, err := upb.NewUserModifierServiceClient(s.UserMgmtConn).UpsertStudentCoursePackage(ctx, &upb.UpsertStudentCoursePackageRequest{
		StudentId: stepState.Student.ID,
		StudentPackageProfiles: []*upb.UpsertStudentCoursePackageRequest_StudentPackageProfile{{
			Id: &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
				CourseId: stepState.CourseID,
			},
			StartTime: timestamppb.New(time.Now().Add(time.Hour * -20)),
			EndTime:   timestamppb.New(time.Now().Add(time.Hour * 24 * 10)),
		}},
	}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to update student course package: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreateAQuizUsingV2(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	ctx, err := s.createALO(ctx)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), nil
	}
	ctx, err = s.createAQuizV2(ctx)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), nil
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createAQuizV2(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	quizLO := utils.GenerateQuizLOProtobufMessage(1, stepState.LoID)
	stepState.QuizLOList = append(stepState.QuizLOList, quizLO[0])

	if _, err := sspb.NewQuizClient(s.EurekaConn).UpsertFlashcardContent(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.UpsertFlashcardContentRequest{
		FlashcardId: quizLO[0].LoId,
		Quizzes:     []*cpb.QuizCore{quizLO[0].Quiz},
	}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert quiz: %w", err)
	}
	stepState.QuizID = quizLO[0].GetQuiz().ExternalId
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createALO(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.LoID = idutil.ULIDNow()
	req := &epb.UpsertLOsRequest{
		LearningObjectives: []*cpb.LearningObjective{
			{
				Info: &cpb.ContentBasicInfo{
					Id:        stepState.LoID,
					Name:      "name",
					Country:   cpb.Country_COUNTRY_VN,
					Subject:   cpb.Subject_SUBJECT_BIOLOGY,
					SchoolId:  constants.ManabieSchool,
					Grade:     1,
					CreatedAt: timestamppb.Now(),
					UpdatedAt: timestamppb.Now(),
				},
				TopicId: stepState.TopicID,
			},
		},
	}
	if _, err := epb.NewLearningObjectiveModifierServiceClient(s.EurekaConn).UpsertLOs(s.AuthHelper.SignedCtx(ctx, stepState.Token), req); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert lo: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) shuffledQuizTestHaveBeenStored(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	e := &entities.ShuffledQuizSet{}
	stmt := fmt.Sprintf(`SELECT COUNT(*)
	FROM %s 
	WHERE study_plan_id = $1
	AND student_id = $2
	AND learning_material_id = $3
	`, e.TableName())
	var count pgtype.Int8
	if err := s.EurekaDB.QueryRow(ctx, stmt, stepState.StudyPlanID, stepState.Student.ID, stepState.LoID).Scan(&count); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.EurekaDB.QueryRow: %w", err)
	}

	if count.Status != pgtype.Present || count.Int != 1 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't create shuffled quiz test, expect 1 but got %v", count.Int)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) quizBelongToQuestionGroup(ctx context.Context, numItems int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	token := stepState.Token
	ctx, err := s.aSignedIn(ctx, "school admin")
	if err != nil {
		return ctx, err
	}
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	questionGrIDs := stepState.ExistingQuestionHierarchy.GetQuestionGroupIDs()
	quizLO := utils.GenerateQuizLOProtobufMessage(numItems*len(questionGrIDs), stepState.LoID)
	for i, id := range questionGrIDs {
		quizLO[i*2].Quiz.QuestionGroupId, quizLO[i*2+1].Quiz.QuestionGroupId = wrapperspb.String(id), wrapperspb.String(id)
		if err = stepState.ExistingQuestionHierarchy.AddChildrenIDsForQuestionGroup(id, quizLO[i*2].Quiz.ExternalId, quizLO[i*2+1].Quiz.ExternalId); err != nil {
			return ctx, fmt.Errorf("ExistingQuestionHierarchy.AddChildrenIDsForQuestionGroup: %w", err)
		}
	}

	if _, err := utils.UpsertQuizzes(ctx, s.EurekaConn, quizLO); err != nil {
		return ctx, fmt.Errorf("UpsertQuizzes: %w", err)
	}
	stepState.QuizLOList = append(stepState.QuizLOList, quizLO...)
	stepState.Token = token
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetQuizTestResponse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	type questionWithGroup struct {
		ID              string
		QuestionGroupID string
	}
	expectedQuiz := make([]*questionWithGroup, 0)
	expectedQuestionGroup := make([]*cpb.QuestionGroup, 0)
	url, _ := s3.GenerateUploadURL("", "", "rendered rich text")

	for _, v := range stepState.ExistingQuestionHierarchy {
		if v.Type == entities.QuestionHierarchyQuestion {
			expectedQuiz = append(expectedQuiz, &questionWithGroup{
				ID: v.ID,
			})
		} else {
			var totalPoints int32
			for _, id := range v.ChildrenIDs {
				expectedQuiz = append(expectedQuiz, &questionWithGroup{
					ID:              id,
					QuestionGroupID: v.ID,
				})
				quiz := s.getStepStateQuizLOByID(ctx, id)
				if quiz == nil {
					return ctx, fmt.Errorf("could not found quiz id %s in step state", id)
				}
				totalPoints += quiz.Quiz.Point.GetValue()
			}
			expectedQuestionGroup = append(expectedQuestionGroup, &cpb.QuestionGroup{
				QuestionGroupId:    v.ID,
				LearningMaterialId: stepState.LoID,
				TotalChildren:      int32(len(v.ChildrenIDs)),
				TotalPoints:        totalPoints,
				RichDescription: &cpb.RichText{
					Raw:      "raw rich text",
					Rendered: url,
				},
			})
		}
	}

	var (
		quizzes        []*cpb.Quiz
		questionGroups []*cpb.QuestionGroup
	)
	// compare between response's data and expected data
	switch resp := stepState.Response.(type) {
	case *sspb.CreateRetryQuizTestV2Response:
		quizzes = resp.GetQuizzes()
		questionGroups = resp.GetQuestionGroups()
	case *sspb.CreateQuizTestV2Response:
		quizzes = resp.GetQuizzes()
		questionGroups = resp.GetQuestionGroups()
	default:
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("")
	}
	if len(expectedQuiz) != len(quizzes) {
		return ctx, fmt.Errorf("expected %d quiz in response but got %d", len(expectedQuiz), len(quizzes))
	}
	for i := range quizzes {
		if expectedQuiz[i].ID != quizzes[i].Core.ExternalId {
			return ctx, fmt.Errorf("expected in position %d in quiz field response is id %s but got %s", i, expectedQuiz[i].ID, quizzes[i].Core.ExternalId)
		}

		actGrID := ""
		if quizzes[i].Core.GetQuestionGroupId() != nil {
			actGrID = quizzes[i].Core.GetQuestionGroupId().Value
		}
		if expectedQuiz[i].QuestionGroupID != actGrID {
			return ctx, fmt.Errorf("expected quiz %s have question group id is %s but got %s", expectedQuiz[i].ID, expectedQuiz[i].QuestionGroupID, actGrID)
		}
	}

	if len(expectedQuestionGroup) != len(questionGroups) {
		return ctx, fmt.Errorf("expected %d questions in response but got %d", len(expectedQuestionGroup), len(questionGroups))
	}
	for i := range questionGroups {
		if expectedQuestionGroup[i].QuestionGroupId != questionGroups[i].QuestionGroupId {
			return ctx, fmt.Errorf("expected in position %d in question group field response is id %s but got %s", i, expectedQuestionGroup[i].QuestionGroupId, questionGroups[i].QuestionGroupId)
		}

		if expectedQuestionGroup[i].LearningMaterialId != questionGroups[i].LearningMaterialId {
			return ctx, fmt.Errorf("expected question group %s have learning material id is %s but got %s", expectedQuestionGroup[i].QuestionGroupId, expectedQuestionGroup[i].LearningMaterialId, questionGroups[i].LearningMaterialId)
		}

		if expectedQuestionGroup[i].TotalChildren != questionGroups[i].TotalChildren {
			return ctx, fmt.Errorf("expected question group %s have total children is %d but got %d", expectedQuestionGroup[i].QuestionGroupId, expectedQuestionGroup[i].TotalChildren, questionGroups[i].TotalChildren)
		}

		if expectedQuestionGroup[i].TotalPoints != questionGroups[i].TotalPoints {
			return ctx, fmt.Errorf("expected question group %s have total ponts is %d but got %d", expectedQuestionGroup[i].QuestionGroupId, expectedQuestionGroup[i].TotalPoints, questionGroups[i].TotalPoints)
		}

		if !strings.Contains(questionGroups[i].RichDescription.Rendered, strings.ReplaceAll(expectedQuestionGroup[i].RichDescription.Rendered, "//", "")) {
			return ctx, fmt.Errorf("expected question group %s have total rendered rich description is %s but got %s", expectedQuestionGroup[i].QuestionGroupId, expectedQuestionGroup[i].RichDescription.Rendered, questionGroups[i].RichDescription.Rendered)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) getStepStateQuizLOByID(ctx context.Context, id string) *epb.QuizLO {
	stepState := utils.StepStateFromContext[StepState](ctx)
	for _, quiz := range stepState.QuizLOList {
		if quiz.GetQuiz().ExternalId == id {
			return quiz
		}
	}

	return nil
}
