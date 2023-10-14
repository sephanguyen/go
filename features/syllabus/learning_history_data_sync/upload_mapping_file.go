// nolint
package learning_history_data_sync

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/features/syllabus/utils"
	repositories "github.com/manabie-com/backend/internal/eureka/repositories/learning_history_data_sync"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) validCourseExamLoQuestionTagInDB(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.StudentID = idutil.ULIDNow()

	courseID, err := utils.GenerateCourse(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.YasuoConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.CourseID = courseID
	_, err = utils.AValidCourseWithIDs(ctx, s.EurekaDB, []string{stepState.StudentID}, stepState.CourseID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidCourseWithIDs: %w", err)
	}
	questionTagTypeID := idutil.ULIDNow()
	err = utils.GenerateQuestionTagType(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaDB, questionTagTypeID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateQuestionTagType: %s", err.Error())
	}
	questionTagIDs := make([]string, 0)
	uuid := idutil.ULIDNow()
	for i := 0; i < 4; i++ {
		questionTagIDs = append(questionTagIDs, uuid+strconv.Itoa(i))
	}
	err = utils.GenerateQuestionTags(ctx, s.EurekaDB, questionTagIDs, questionTagTypeID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateQuestionTags: %s", err.Error())
	}
	stepState.QuestionTagIDs = questionTagIDs

	_, _, topicIDs, err := utils.AValidBookContent(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, s.EurekaDB, constant.ManabieSchool)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidBookContent: %w", err)
	}
	resp, err := sspb.NewExamLOClient(s.EurekaConn).InsertExamLO(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.InsertExamLORequest{
		ExamLo: &sspb.ExamLOBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: topicIDs[0],
				Name:    "exam-lo-name",
			},
			Instruction:   "instruction",
			GradeToPass:   wrapperspb.Int32(10),
			ManualGrading: false,
			TimeLimit:     wrapperspb.Int32(100),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to insert exam lo: %w", err)
	}
	stepState.ExamLOID = resp.LearningMaterialId

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUploadMappingFile(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	headerMappingCourseID := "manabie_course_id,withus_course_id,is_archived"
	row1MappingCourseID := fmt.Sprintf("%s,withus_course_id_1,false", stepState.CourseID)

	headerMappingExamLoID := "exam_lo_id,material_code,is_archived"
	row1MappingExamLoID := fmt.Sprintf("%s,material_code_1,false", stepState.ExamLOID)

	headerMappingQuestionTag := "manabie_tag_id,manabie_tag_name,withus_tag_name,is_archived"
	row1MappingQuestionTag := fmt.Sprintf("%s,manabie_tag_name_1,withus_tag_name_1,false", stepState.QuestionTagIDs[0])

	headerFailedSync := "recipient_id,email_address,is_archived"
	row1FailedSync := "recipient_id_1,email_address_1,false"
	row2FailedSync := "recipient_id_2,email_address_2,false"

	req := &sspb.UploadMappingFileRequest{
		MappingCourseId: []byte(fmt.Sprintf(`%s
		%s`, headerMappingCourseID, row1MappingCourseID)),
		MappingExamLoId: []byte(fmt.Sprintf(`%s
		%s`, headerMappingExamLoID, row1MappingExamLoID)),
		MappingQuestionTag: []byte(fmt.Sprintf(`%s
		%s`, headerMappingQuestionTag, row1MappingQuestionTag)),
		FailedSyncEmailRecipients: []byte(fmt.Sprintf(`%s
		%s
		%s`, headerFailedSync, row1FailedSync, row2FailedSync)),
	}

	stepState.Response, stepState.ResponseErr = sspb.NewLearningHistoryDataSyncServiceClient(s.EurekaConn).UploadMappingFile(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)

	return utils.StepStateToContext(ctx, stepState), nil
}

// nolint
func (s *Suite) svFileIsUploaded(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	repo := &repositories.LearningHistoryDataSyncRepo{}
	mappingCourseID, err := repo.RetrieveMappingCourseID(ctx, s.EurekaDB)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	mappingExamLoID, err := repo.RetrieveMappingExamLoID(ctx, s.EurekaDB)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	mappingQuestionTag, err := repo.RetrieveMappingQuestionTag(ctx, s.EurekaDB)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	failedSyncEmailRecipients, err := repo.RetrieveFailedSyncEmailRecipient(ctx, s.EurekaDB)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	isExist := false
	for _, mapping := range mappingCourseID {
		if mapping.ManabieCourseID.String == stepState.CourseID && mapping.WithusCourseID.String == "withus_course_id_1" {
			isExist = true
		}
	}
	if !isExist {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("mapping course id is not uploaded")
	}

	isExist = false
	for _, mapping := range mappingExamLoID {
		if mapping.ExamLoID.String == stepState.ExamLOID && mapping.MaterialCode.String == "material_code_1" {
			isExist = true
		}
	}
	if !isExist {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("mapping exam lo id is not uploaded")
	}

	isExist = false
	for _, mapping := range mappingQuestionTag {
		if mapping.ManabieTagID.String == stepState.QuestionTagIDs[0] && mapping.WithusTagName.String == "withus_tag_name_1" {
			isExist = true
		}
	}
	if !isExist {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("mapping question tag is not uploaded")
	}

	isExist = false
	for _, mapping := range failedSyncEmailRecipients {
		if mapping.RecipientID.String == "recipient_id_1" && mapping.EmailAddress.String == "email_address_1" {
			isExist = true
		}
	}
	if !isExist {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("failed sync email recipient is not uploaded")
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
