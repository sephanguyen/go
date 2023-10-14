package study_plan

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) aSignedIn(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// reset token
	stepState.Token = ""
	userID, authToken, err := s.AuthHelper.AUserSignedInAsRole(ctx, arg)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	//TODO: no need if you're not use it. Just an example.
	switch arg {
	case "student":
		stepState.Student.Token = authToken
		stepState.Student.ID = userID
	case "school admin", "admin":
		stepState.SchoolAdmin.Token = authToken
		stepState.SchoolAdmin.ID = userID
	case "teacher", "current teacher":
		stepState.Teacher.Token = authToken
		stepState.Teacher.ID = userID
	case "parent":
		stepState.Parent.Token = authToken
		stepState.Parent.ID = userID
	case "hq staff":
		stepState.HQStaff.Token = authToken
		stepState.HQStaff.ID = userID
	default:
		stepState.Student.Token = authToken
		stepState.Student.ID = userID
	}
	stepState.Token = authToken
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) listStudyPlanWithLearningMaterial(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	bookID, _, topicIDs, err := utils.AValidBookContent(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, s.EurekaDB, constant.ManabieSchool)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	learningMaterialIDs, err := utils.GenerateFlashcard(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, topicIDs[0])

	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	courseID, err := utils.GenerateCourse(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.YasuoConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	studyPlanID, err := utils.GenerateStudyPlan(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, courseID, bookID)

	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.LearningMaterialIDs = learningMaterialIDs
	stepState.StudyPlanID = studyPlanID
	return utils.StepStateToContext(ctx, stepState), err
}

func (s *Suite) adminUpsertMasterStudyPlanInfo(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if len(stepState.LearningMaterialIDs) > 0 {
		stepState.LearningMaterialID = stepState.LearningMaterialIDs[0]
	}

	req := &sspb.UpsertMasterInfoRequest{
		MasterItems: []*sspb.MasterStudyPlan{
			{
				MasterStudyPlanIdentify: &sspb.MasterStudyPlanIdentify{
					StudyPlanId:        stepState.StudyPlanID,
					LearningMaterialId: stepState.LearningMaterialID,
				},
				AvailableFrom: timestamppb.Now(),
				AvailableTo:   timestamppb.New(time.Now().Add(time.Hour)),
				StartDate:     timestamppb.Now(),
				EndDate:       timestamppb.New(time.Now().Add(time.Hour)),
				Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
				SchoolDate:    timestamppb.Now(),
			},
		},
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = sspb.NewStudyPlanClient(s.EurekaConn).UpsertMasterInfo(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsStatusCode(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	return utils.StepStateToContext(ctx, stepState), utils.ValidateStatusCode(stepState.ResponseErr, arg)
}

func (s *Suite) ourSystemStoresIndividualStudyPlanCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	msp := &entities.MasterStudyPlan{}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE study_plan_id = $1", strings.Join(database.GetFieldNames(msp), ","), msp.TableName())
	if err := database.Select(ctx, s.EurekaDB, query, stepState.StudyPlanID).ScanOne(msp); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	req := stepState.Request.(*sspb.UpsertMasterInfoRequest)
	mspReq := req.GetMasterItems()[0]

	if mspReq.MasterStudyPlanIdentify.LearningMaterialId != msp.LearningMaterialID.String {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf(fmt.Sprintf("Incorrect learning material id, expected %v got %v", mspReq.MasterStudyPlanIdentify.LearningMaterialId, msp.LearningMaterialID.String))
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminUpdateInfoOfMasterStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := stepState.Request.(*sspb.UpsertMasterInfoRequest)
	stepState.StartDate = timestamppb.Now()
	req.MasterItems[0].StartDate = timestamppb.Now()

	if stepState.Response, stepState.ResponseErr = sspb.NewStudyPlanClient(s.EurekaConn).UpsertMasterInfo(s.AuthHelper.SignedCtx(ctx, stepState.Token), req); stepState.ResponseErr != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf(fmt.Sprintf("unable to upsert a master study plan: %v", stepState.ResponseErr.Error()))
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemUpdatesStartDateForIndividualStudyPlanCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	msp := &entities.MasterStudyPlan{}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE study_plan_id = $1", strings.Join(database.GetFieldNames(msp), ","), msp.TableName())
	if err := database.Select(ctx, s.EurekaDB, query, stepState.StudyPlanID).ScanOne(msp); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	req := stepState.Request.(*sspb.UpsertMasterInfoRequest)
	var reqTime pgtype.Timestamptz
	reqTime.Set(req.GetMasterItems()[0].StartDate.AsTime())

	if reqTime.Time.Equal(msp.StartDate.Time) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf(fmt.Sprintf("Incorrect start data, expected %v got %v", reqTime.Time, msp.StartDate.Time))
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
