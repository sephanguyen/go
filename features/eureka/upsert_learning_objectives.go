package eureka

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) generateValidLearningObjectiveEntity(topicID string) *entities.LearningObjective {
	now := time.Now()
	lo := new(entities.LearningObjective)
	database.AllNullEntity(lo)

	err := multierr.Combine(
		lo.ID.Set(idutil.ULIDNow()),
		lo.Name.Set(idutil.ULIDNow()),
		lo.Country.Set(bpb.COUNTRY_VN.String()),
		lo.Grade.Set(12),
		lo.Subject.Set(bpb.SUBJECT_MATHS.String()),
		lo.TopicID.Set(topicID),
		lo.MasterLoID.Set(nil),
		lo.DisplayOrder.Set(1),
		lo.VideoScript.Set(""),
		lo.Prerequisites.Set(nil),
		lo.Video.Set(""),
		lo.StudyGuide.Set("url_study_guide"),
		lo.SchoolID.Set(constants.ManabieSchool),
		lo.UpdatedAt.Set(now),
		lo.CreatedAt.Set(now),
		lo.DeletedAt.Set(nil),
		lo.ApproveGrading.Set(false),
		lo.GradeCapping.Set(false),
		lo.ReviewOption.Set(cpb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_IMMEDIATELY.String()),
		lo.VendorType.Set(cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_MANABIE.String()),
	)

	if err != nil {
		return nil
	}

	return lo
}

func (s *suite) aValidBookContent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	bookResp, err := pb.NewBookModifierServiceClient(s.Conn).UpsertBooks(s.signedCtx(ctx), &pb.UpsertBooksRequest{
		Books: s.generateBooks(1, nil),
	})
	if err != nil {
		if err.Error() == "rpc error: code = PermissionDenied desc = auth: not allowed" {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), nil
		}
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create book: %w", err)
	}
	stepState.BookID = bookResp.BookIds[0]
	chapterResp, err := pb.NewChapterModifierServiceClient(s.Conn).UpsertChapters(s.signedCtx(ctx), &pb.UpsertChaptersRequest{
		Chapters: s.generateChapters(ctx, 1, nil),
		BookId:   stepState.BookID,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create chapter: %w", err)
	}
	stepState.ChapterID = chapterResp.ChapterIds[0]
	topicResp, err := pb.NewTopicModifierServiceClient(s.Conn).Upsert(s.signedCtx(ctx), &pb.UpsertTopicsRequest{
		Topics: s.generateTopics(ctx, 1, nil),
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create topic: %w", err)
	}
	stepState.TopicID = topicResp.TopicIds[0]
	stepState.TopicIDs = append(stepState.TopicIDs, stepState.TopicID)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) learningObjectivesMustBeCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	loRepo := repositories.LearningObjectiveRepo{}

	los, err := loRepo.RetrieveByIDs(ctx, s.DB, database.TextArray(stepState.LOIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve los by ids: %w", err)
	}
	m := make(map[string]bool)
	for _, lo := range los {
		m[lo.ID.String] = true
	}
	for _, loID := range stepState.LOIDs {
		if ok := m[loID]; !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("missing lo %v", loID)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) learningObjectivesMustBeUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	loRepo := repositories.LearningObjectiveRepo{}

	los, err := loRepo.RetrieveByIDs(ctx, s.DB, database.TextArray(stepState.LoIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve los by ids: %w", err)
	}
	for _, lo := range los {
		if !strings.Contains(lo.Name.String, "updated") {
			return StepStateToContext(ctx, stepState), fmt.Errorf("name of lo %s not updated", lo.ID.String)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreateLearningObjectives(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := s.generateLOsReq(ctx)
	stepState.Request = req
	stepState.LearningObjectives = req.LearningObjectives
	for _, lo := range req.LearningObjectives {
		stepState.LoIDs = append(stepState.LoIDs, lo.Info.Id)
	}
	if _, err := pb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(s.signedCtx(ctx), req); err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), nil
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateLearningObjectives(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.UpsertLOsRequest)
	for _, lo := range req.LearningObjectives {
		lo.Info.Name = fmt.Sprintf("%s-updated", lo.Info.Name)
	}

	stepState.Request = req
	if _, err := pb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(s.signedCtx(ctx), req); err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), nil
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreateNLearningObjectivesWithType(ctx context.Context, numberLos int, loType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, pbLOs := s.prepareLO(ctx, stepState.TopicID, numberLos, cpb.LearningObjectiveType(cpb.LearningObjectiveType_value[loType]))

	if ctx, err := s.createLO(ctx, pbLOs); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateLearningObjectivesFields(ctx context.Context, field string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.UpsertLOsRequest)

	for i, lo := range req.LearningObjectives {
		switch field {
		case "instruction":
			lo.Instruction = fmt.Sprintf("%s-updated", lo.Instruction)
		case "grade_to_pass":
			if lo.GradeToPass == nil {
				stepState.UpdatedGradeToPass = 1
			} else {
				stepState.UpdatedGradeToPass = lo.GradeToPass.Value + 1
			}
			lo.GradeToPass = wrapperspb.Int32(stepState.UpdatedGradeToPass)
		case "manual_grading":
			stepState.UpdatedManualGrading = !lo.ManualGrading
			lo.ManualGrading = stepState.UpdatedManualGrading
		case "time_limit":
			if lo.GradeToPass == nil {
				stepState.UpdatedTimeLimit = 1
			} else {
				stepState.UpdatedTimeLimit = lo.TimeLimit.Value + 1
			}
			lo.TimeLimit = wrapperspb.Int32(stepState.UpdatedTimeLimit)
		case "maximum_attempt":
			stepState.UpdatedMaximumAttempt = 99
			lo.MaximumAttempt = wrapperspb.Int32(stepState.UpdatedMaximumAttempt)
		case "approve_grading":
			stepState.UpdatedApproveGrading = i%2 == 0
			lo.ApproveGrading = stepState.UpdatedApproveGrading
		case "grade_capping":
			stepState.UpdatedGradeCapping = i%2 != 0
			lo.GradeCapping = stepState.UpdatedGradeCapping
		case "review_option":
			stepState.UpdatedReviewOption = cpb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE.String()
			lo.ReviewOption = cpb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE
		default:
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid field")
		}
	}

	stepState.Request = req

	if _, err := pb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(s.signedCtx(ctx), req); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert los: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) fieldOfLearningObjectsMustBeUpdated(ctx context.Context, field string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	loRepo := repositories.LearningObjectiveRepo{}

	los, err := loRepo.RetrieveByIDs(ctx, s.DB, database.TextArray(stepState.LoIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve los by ids: %w", err)
	}
	for _, lo := range los {
		var isUpdated bool

		switch field {
		case "instruction":
			isUpdated = strings.Contains(lo.Instruction.String, "updated")
		case "grade_to_pass":
			isUpdated = lo.GradeToPass.Int == stepState.UpdatedGradeToPass
		case "manual_grading":
			isUpdated = lo.ManualGrading.Bool == stepState.UpdatedManualGrading
		case "time_limit":
			isUpdated = lo.TimeLimit.Int == stepState.UpdatedTimeLimit
		case "maximum_attempt":
			isUpdated = lo.MaximumAttempt.Int == stepState.UpdatedMaximumAttempt
		case "approve_grading":
			isUpdated = lo.ApproveGrading.Bool == stepState.UpdatedApproveGrading
		case "grade_capping":
			isUpdated = lo.GradeCapping.Bool == stepState.UpdatedGradeCapping
		case "review_option":
			isUpdated = lo.ReviewOption.String == stepState.UpdatedReviewOption
		default:
			return StepStateToContext(ctx, stepState), fmt.Errorf("Invalid field")
		}

		if !isUpdated {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%s of lo %s not updated", field, lo.ID.String)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreateNLearningObjectivesWithDefaultValuesAndType(ctx context.Context, numberLos int, loType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, pbLOs := s.prepareLOWithDefaultValues(ctx, stepState.TopicID, numberLos, cpb.LearningObjectiveType(cpb.LearningObjectiveType_value[loType]))

	if ctx, err := s.createLO(ctx, pbLOs); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareLOWithDefaultValues(ctx context.Context, topicID string, numberOfLOs int, loType cpb.LearningObjectiveType) (context.Context, []*cpb.LearningObjective) {
	stepState := StepStateFromContext(ctx)
	pbLOs := make([]*cpb.LearningObjective, 0, numberOfLOs)
	for i := 0; i < numberOfLOs; i++ {
		stepState.LoID = idutil.ULIDNow()
		stepState.LoIDs = append(stepState.LoIDs, stepState.LoID)
		pbLOs = append(pbLOs, &cpb.LearningObjective{
			Info: &cpb.ContentBasicInfo{
				Id:           stepState.LoID,
				Name:         fmt.Sprintf("lo-%s-name+%s", loType.String(), stepState.LoID),
				Country:      cpb.Country_COUNTRY_VN,
				Grade:        1,
				Subject:      cpb.Subject_SUBJECT_BIOLOGY,
				DisplayOrder: int32(i + 1),

				SchoolId: stepState.SchoolIDInt,
			},
			Type:    loType,
			TopicId: topicID,
			// Instruction as nil
			GradeToPass:   nil,
			ManualGrading: false,
			TimeLimit:     nil,
			VendorType:    cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_MANABIE,
		})
	}
	return StepStateToContext(ctx, stepState), pbLOs
}

func (s *suite) learningObjectivesMustBeCreatedWithFieldsAsDefaultValue(ctx context.Context, field string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	loRepo := repositories.LearningObjectiveRepo{}

	los, err := loRepo.RetrieveByIDs(ctx, s.DB, database.TextArray(stepState.LoIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve los by ids: %w", err)
	}
	for _, lo := range los {
		var isDefault bool

		switch field {
		case "instruction":
			isDefault = lo.Instruction == pgtype.Text{Status: pgtype.Null}
		case "grade_to_pass":
			isDefault = lo.GradeToPass == pgtype.Int4{Status: pgtype.Null}
		case "manual_grading":
			isDefault = !lo.ManualGrading.Bool // false as default
		case "time_limit":
			isDefault = lo.TimeLimit == pgtype.Int4{Status: pgtype.Null}
		case "maximum_attempt":
			isDefault = lo.MaximumAttempt == pgtype.Int4{Status: pgtype.Null}
		case "approve_grading":
			isDefault = !lo.ApproveGrading.Bool
		case "grade_capping":
			isDefault = !lo.GradeCapping.Bool
		case "review_option":
			isDefault = lo.ReviewOption.String == cpb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_IMMEDIATELY.String()
		case "vendor_type":
			isDefault = lo.VendorType.String == cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_MANABIE.String()
		default:
			return StepStateToContext(ctx, stepState), fmt.Errorf("Invalid field")
		}

		if !isDefault {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%s of lo %v not default", field, lo.Instruction)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
