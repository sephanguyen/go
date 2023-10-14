package eureka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

func (s *suite) quizTestsInfor(ctx context.Context, arg1 int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp, ok := stepState.Response.(*epb.RetrieveQuizTestsResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), errors.New("returns response is no RetrieveQuizTestsResponse err")
	}

	numTests := arg1
	if numTests != len(resp.Items[stepState.StudyPlanItemID].Items) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("returns num of quiz tests expect %v but got %v", numTests, len(resp.Items))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentsDoQuizExamOfAStudyPlanItem(ctx context.Context, numStudentsStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	numStudents, _ := strconv.Atoi(numStudentsStr)
	s.aStudyPlanItemId(ctx)

	limit := rand.Intn(len(stepState.Quizzes)) + 1
	stepState.QuizOptions = make(map[string]map[string][]*cpb.QuizOption)
	for i := 0; i < numStudents; i++ {
		ctx, err1 := s.aValidStudentAccount(ctx)
		ctx, err2 := s.doQuizExam(ctx, limit, false)
		if err := multierr.Combine(err1, err2); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) teacherGetQuizTestOfAStudyPlanItem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.aUserSignedInTeacher(ctx)
	if err := try.Do(func(attempt int) (bool, error) {
		time.Sleep(1 * time.Second)
		stepState.Response, stepState.ResponseErr = epb.NewQuizReaderServiceClient(s.Conn).RetrieveQuizTests(s.signedCtx(ctx), &epb.RetrieveQuizTestsRequest{
			StudyPlanItemId: []string{stepState.StudyPlanItemID},
		})
		resp := stepState.Response.(*epb.RetrieveQuizTestsResponse)
		if len(resp.Items) == 0 {
			return attempt < 10, errors.New("not match quiz items length")
		}
		return false, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) teacherGetQuizTestWithoutStudyPlanItemId(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.aUserSignedInTeacher(ctx)

	stepState.Response, stepState.ResponseErr = epb.NewQuizReaderServiceClient(s.Conn).RetrieveQuizTests(s.signedCtx(ctx), &epb.RetrieveQuizTestsRequest{
		StudyPlanItemId: []string{},
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) compareQuizTestsListWithBobService(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.BobResponse, stepState.ResponseErr = epb.NewQuizReaderServiceClient(s.Conn).RetrieveQuizTests(s.signedCtx(ctx), &epb.RetrieveQuizTestsRequest{
		StudyPlanItemId: []string{stepState.StudyPlanItemID},
	})
	bobBytes, _ := json.Marshal(stepState.BobResponse)
	eurekaBytes, _ := json.Marshal(stepState.Response)
	if string(bobBytes) != string(eurekaBytes) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not match with bob response")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aLearningObjectiveBelongedToATopic(ctx context.Context, topic string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.aSignedIn(ctx, "school admin")
	if len(stepState.TopicIDs) == 0 {
		return StepStateToContext(ctx, stepState), errors.New("topic can't empty")
	}
	t1 := stepState.TopicIDs[0]

	lo := s.generateValidLearningObjectiveEntity(t1)
	if _, err := database.Insert(ctx, lo, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	now := time.Now()
	topicLO := &entities.TopicsLearningObjectives{
		TopicID:      database.Text(t1),
		LoID:         database.Text(lo.ID.String),
		DisplayOrder: database.Int2(lo.DisplayOrder.Int),
		CreatedAt:    database.Timestamptz(now),
		UpdatedAt:    database.Timestamptz(now),
		DeletedAt:    pgtype.Timestamptz{Status: pgtype.Null},
	}
	if _, err := database.Insert(ctx, topicLO, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.LoID = lo.ID.String
	stepState.Request = lo.ID.String
	return StepStateToContext(ctx, stepState), nil
}
