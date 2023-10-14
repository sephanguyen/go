package exam_lo

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"sync"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) userInsertAValidExamLO(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := &sspb.InsertExamLORequest{
		ExamLo: &sspb.ExamLOBase{
			Base: &sspb.LearningMaterialBase{
				Name:    "Exam LO",
				TopicId: stepState.TopicID,
			},
		},
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = sspb.NewExamLOClient(s.EurekaConn).InsertExamLO(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	if stepState.ResponseErr == nil {
		stepState.TopicLODisplayOrderCounter++
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) thereAreExamLOsExistedInTopic(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	n := rand.Intn(5) + 3 // maximum = 8
	wg := &sync.WaitGroup{}
	var err error
	cErrs := make(chan error, n)
	cIDs := make(chan string, n)

	defer func() {
		close(cErrs)
		close(cIDs)
	}()
	genAndInsert := func(ctx context.Context, i int, wg *sync.WaitGroup) {
		defer wg.Done()
		insertExamLOReq := &sspb.InsertExamLORequest{
			ExamLo: &sspb.ExamLOBase{
				Base: &sspb.LearningMaterialBase{
					TopicId: stepState.TopicIDs[0],
					Name:    fmt.Sprintf("exam-lo-name+%d", i),
				},
				MaximumAttempt: wrapperspb.Int32(int32(i + 2)),
				ApproveGrading: i%2 == 0,
				GradeCapping:   i%2 != 0,
				ReviewOption:   sspb.ExamLOReviewOption(int32(math.Min(1, float64(i%2)))),
			},
		}
		resp, err := sspb.NewExamLOClient(s.EurekaConn).InsertExamLO(s.AuthHelper.SignedCtx(ctx, stepState.Token), insertExamLOReq)
		if err != nil {
			cErrs <- err
		}
		cIDs <- resp.LearningMaterialId
	}
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go genAndInsert(ctx, i, wg)
	}
	go func() {
		wg.Wait()
	}()
	for i := 0; i < n; i++ {
		select {
		case errTemp := <-cErrs:
			err = multierr.Combine(err, errTemp)
		case id := <-cIDs:
			stepState.LearningMaterialIDs = append(stepState.LearningMaterialIDs, id)
			stepState.TopicLODisplayOrderCounter++
		}
	}
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemGeneratesACorrectDisplayOrderForExamLO(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	e := &entities.ExamLO{
		LearningMaterial: entities.LearningMaterial{
			ID: database.Text(stepState.Response.(*sspb.InsertExamLOResponse).LearningMaterialId),
		},
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE learning_material_id = $1", strings.Join(database.GetFieldNames(e), ","), e.TableName())
	if err := database.Select(ctx, s.EurekaDB, query, e.ID).ScanOne(e); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	if int32(e.DisplayOrder.Int) != stepState.TopicLODisplayOrderCounter {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Incorrect ExamLO DisplayOrder: expected %d, got %d", stepState.TopicLODisplayOrderCounter, e.DisplayOrder.Int)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemUpdatesTopicLODisplayOrderCounterCorrectlyWithNewExamLo(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	topic := &entities.Topic{
		ID: database.Text(stepState.TopicID),
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE topic_id = $1", strings.Join(database.GetFieldNames(topic), ","), topic.TableName())
	if err := database.Select(ctx, s.EurekaDB, query, topic.ID).ScanOne(topic); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	if topic.LODisplayOrderCounter.Int != stepState.TopicLODisplayOrderCounter {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Incorrect Topic LODisplayOrderCounter: expected %d, got %d", stepState.TopicLODisplayOrderCounter, topic.LODisplayOrderCounter.Int)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userInsertAExamLOWithoutField(ctx context.Context, field string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	req := &sspb.InsertExamLORequest{
		ExamLo: &sspb.ExamLOBase{
			Base: &sspb.LearningMaterialBase{
				Name:    "Exam LO",
				TopicId: stepState.TopicID,
			},
			Instruction:   "instruction",
			GradeToPass:   wrapperspb.Int32(1),
			ManualGrading: true,
			TimeLimit:     wrapperspb.Int32(1),
		},
	}

	switch field {
	case "instruction":
		req.ExamLo.Instruction = ""
	case "manual_grading":
		req.ExamLo.ManualGrading = false
	case "time_limit":
		req.ExamLo.TimeLimit = nil
	case "grade_to_pass":
		req.ExamLo.GradeToPass = nil
	case "maximum_attempt":
		req.ExamLo.MaximumAttempt = wrapperspb.Int32(5)
	case "approve_grading":
		req.ExamLo.ApproveGrading = true
	case "grade_capping":
		req.ExamLo.GradeCapping = false
	case "review_option":
		req.ExamLo.ReviewOption = sspb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE
	default:
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("field %s not belong to exam LO", field)
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = sspb.NewExamLOClient(s.EurekaConn).InsertExamLO(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustCreateExamLOWithDefaultValue(ctx context.Context, field string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	insertResp := stepState.Response.(*sspb.InsertExamLOResponse)
	insertReq := stepState.Request.(*sspb.InsertExamLORequest)

	stepState.LearningMaterialIDs = []string{insertResp.LearningMaterialId}
	stepState.ExamLOBase = insertReq.ExamLo

	stepState.Response, stepState.ResponseErr = sspb.NewExamLOClient(s.EurekaConn).ListExamLO(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListExamLORequest{
		LearningMaterialIds: stepState.LearningMaterialIDs,
	})
	listExamLOResp := stepState.Response.(*sspb.ListExamLOResponse)
	expectedDefaultInt32Value := &wrapperspb.Int32Value{}
	for _, exam := range listExamLOResp.ExamLos {
		switch field {
		case "instruction":
			if exam.Instruction != stepState.ExamLOBase.Instruction {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam LO instruction: want %s, got %s", stepState.ExamLOBase.Instruction, exam.Instruction)
			}
		case "manual_grading":
			if exam.ManualGrading != stepState.ExamLOBase.ManualGrading {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam LO manual_grading: want %t, got %t", stepState.ExamLOBase.ManualGrading, exam.ManualGrading)
			}
		case "time_limit":
			if !reflect.DeepEqual(exam.TimeLimit, expectedDefaultInt32Value) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam LO time_limit: want %v, got %v", expectedDefaultInt32Value, exam.TimeLimit)
			}
		case "grade_to_pass":
			if !reflect.DeepEqual(exam.GradeToPass, expectedDefaultInt32Value) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam LO grade_to_pass: want %v, got %v", expectedDefaultInt32Value, exam.GradeToPass)
			}
		case "maximum_attempt":
			if exam.MaximumAttempt.GetValue() != stepState.ExamLOBase.MaximumAttempt.GetValue() {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam LO maximum_attempt: want %d, got %d", stepState.ExamLOBase.MaximumAttempt.GetValue(), exam.MaximumAttempt.GetValue())
			}
		case "approve_grading":
			if exam.ApproveGrading != stepState.ExamLOBase.ApproveGrading {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam LO approve_grading: want %t, got %t", stepState.ExamLOBase.ApproveGrading, exam.ApproveGrading)
			}
		case "grade_capping":
			if exam.GradeCapping != stepState.ExamLOBase.GradeCapping {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam LO grade_capping: want %t, got %t", stepState.ExamLOBase.GradeCapping, exam.GradeCapping)
			}
		case "review_option":
			if exam.ReviewOption.String() != stepState.ExamLOBase.ReviewOption.String() {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam LO review_option: want %s, got %s", stepState.ExamLOBase.ReviewOption.String(), exam.ReviewOption.String())
			}
		default:
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("field %s not belong to exam LO", field)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
