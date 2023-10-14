package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	shuffledAction  = "shuffled"
	keepOrderAction = "keep order"
)

func (s *suite) executeCreateFlashcardStudyTestServiceV2(ctx context.Context, request *pb.CreateFlashCardStudyRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp, err := pb.NewQuizModifierServiceClient(s.Conn).CreateFlashCardStudy(s.signedCtx(ctx), request)
	if err == nil {
		stepState.StudySetID = resp.StudySetId
	}
	stepState.Response = resp
	stepState.ResponseErr = err
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetNextPageOfFlashcardStudy(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	request := &pb.CreateFlashCardStudyRequest{StudyPlanItemId: stepState.StudyPlanItemID, LoId: stepState.LoID, StudentId: stepState.StudentID, StudySetId: stepState.StudySetID, Paging: stepState.NextPage, KeepOrder: false}
	return s.executeCreateFlashcardStudyTestServiceV2(ctx, request)
}

func (s *suite) thatStudentCanFetchTheListOfFlashcardStudyQuizzesPageByPageUsingLimitFlashcardStudyQuizzesPerPage(ctx context.Context, arg1 string) (_ context.Context, err error) {
	stepState := StepStateFromContext(ctx)
	stepState.Limit, _ = strconv.Atoi(arg1)
	for i := 0; i < (len(stepState.QuizItems)/stepState.Limit)-1; i++ {
		ctx, err = s.userGetNextPageOfFlashcardStudy(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		ctx, err = s.returnListOfFlashcardStudyItems(ctx, strconv.Itoa(stepState.Limit))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentDoingALongExamWithFlashcardStudyQuizzesPerPage(ctx context.Context, limitStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Limit, _ = strconv.Atoi(limitStr)
	stepState.Offset = 1
	stepState.AllQuizzesRes = make([]*cpb.Quiz, 0)
	request := &pb.CreateFlashCardStudyRequest{
		StudyPlanItemId: stepState.StudyPlanItemID,
		LoId:            stepState.LoID,
		StudentId:       stepState.StudentID,
		StudySetId:      stepState.StudySetID,
		Paging: &cpb.Paging{
			Limit: uint32(stepState.Limit),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: int64(stepState.Offset),
			},
		},
	}
	ctx, err := s.executeCreateFlashcardStudyTestServiceV2(ctx, request)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.returnListOfFlashcardStudyItems(ctx, strconv.Itoa(stepState.Limit))
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) returnListOfFlashcardStudyItems(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.returnsStatusCode(ctx, "OK")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	numOfFlashcardQuizzes, _ := strconv.Atoi(arg1)

	resp, ok := stepState.Response.(*pb.CreateFlashCardStudyResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not receive create flashcard study response")
	}
	var limit int
	var offset int
	if stepState.NextPage == nil {
		limit = stepState.Limit
		offset = stepState.Offset
	} else {
		limit = int(stepState.NextPage.Limit)
		offset = int(stepState.NextPage.GetOffsetInteger())
	}
	stepState.NextPage = resp.NextPage
	stepState.StudySetID = resp.StudySetId
	for _, flashcardStudyItem := range resp.Items {
		stepState.QuizItems = append(stepState.QuizItems, flashcardStudyItem.Item)
	}
	if len(resp.Items) != numOfFlashcardQuizzes {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect number of flashcard study quizzes are %v but got %v", numOfFlashcardQuizzes, len(resp.Items))
	}

	quizExternalIDs := make([]string, 0)
	query := `SELECT quiz_external_id FROM flashcard_progressions fps INNER JOIN UNNEST(fps.quiz_external_ids) AS quiz_external_id ON study_set_id = $1 LIMIT $2 OFFSET $3;`

	rows, err := s.DB.Query(ctx, query, stepState.StudySetID, limit, offset-1)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()
	for rows.Next() {
		var id pgtype.Text
		err := rows.Scan(&id)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		quizExternalIDs = append(quizExternalIDs, id.String)
	}

	if len(resp.Items) != len(quizExternalIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect number of flashcard study quizzes are %v but got %v", numOfFlashcardQuizzes, len(resp.Items))
	}

	for i := range resp.Items {
		if resp.Items[i].Item.Core.ExternalId != quizExternalIDs[i] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect quiz id %v but got %v", quizExternalIDs, resp.Items[i].Item.Core.ExternalId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userCreateFlashcardStudyWithoutLoID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rand.Seed(time.Now().UnixNano())
	limit := rand.Intn(10)
	offset := rand.Intn(3)

	stepState.Response, stepState.ResponseErr = pb.NewQuizModifierServiceClient(s.Conn).CreateFlashCardStudy(s.signedCtx(ctx), &pb.CreateFlashCardStudyRequest{
		StudyPlanItemId: stepState.StudyPlanItemID,
		LoId:            "",
		StudentId:       stepState.StudentID,
		Paging: &cpb.Paging{
			Limit: uint32(limit),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: int64(offset),
			},
		},
		KeepOrder: false,
	})

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userCreateFlashcardStudyWithoutPaging(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewQuizModifierServiceClient(s.Conn).CreateFlashCardStudy(s.signedCtx(ctx), &pb.CreateFlashCardStudyRequest{
		StudyPlanItemId: stepState.StudyPlanItemID,
		LoId:            "lo-id",
		StudentId:       stepState.StudentID,
		KeepOrder:       false,
	})

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsEmptyFlashcardStudyItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stt, ok := status.FromError(stepState.ResponseErr)

	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}

	if stt.Code() != codes.OK {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting %s, got %s status code, message: %s", codes.OK.String(), stt.Code().String(), stt.Message())
	}

	resp, ok := stepState.Response.(*pb.CreateFlashCardStudyResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not receive create flashcard study response")
	}
	if len(resp.Items) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("want empty flashcard study items but get %v flashcard study items", len(resp.Items))
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userCreateFlashcardStudyWithValidRequestAndOffsetAndLimit(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Offset, _ = strconv.Atoi(arg1)
	stepState.Limit, _ = strconv.Atoi(arg2)
	request := &pb.CreateFlashCardStudyRequest{
		LoId:            stepState.LoID,
		StudentId:       stepState.StudentID,
		StudyPlanItemId: stepState.StudyPlanItemID,
		Paging: &cpb.Paging{
			Limit: uint32(stepState.Limit),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: int64(stepState.Offset),
			},
		},
	}
	ctx, err := s.executeCreateFlashcardStudyTestServiceV2(ctx, request)

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) returnsExpectedListOfFlashcardStudyQuizzesWith(ctx context.Context, action string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch action {
	case shuffledAction:
		// check where quiz_external_ids is shuffled or not, if it is, then return error
		if shuffled, err := s.IsShuffleFlashcardStudyQuizzes(ctx); err != nil || !shuffled {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect shuffled")
		}
	case keepOrderAction:
		// check where quiz_external_ids is shuffled or not, if it is, then return error
		if shuffled, err := s.IsShuffleFlashcardStudyQuizzes(ctx); err != nil || shuffled {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect no shuffled")
		}
	}

	query := `SELECT flp.quiz_external_ids FROM flashcard_progressions flp where flp.study_set_id=$1`
	var quizExternalIDs pgtype.TextArray
	err := s.DB.QueryRow(ctx, query, stepState.StudySetID).Scan(&quizExternalIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(quizExternalIDs.Elements) != len(stepState.QuizItems) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect create flashcard study response %v items but got %v", len(quizExternalIDs.Elements), len(stepState.QuizItems))
	}
	for i := range quizExternalIDs.Elements {
		if quizExternalIDs.Elements[i].String != stepState.QuizItems[i].Core.ExternalId {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect create flashcard study return ordered items %v \n but got %v",
				quizExternalIDs, s.getExternalIDsFromQuizItems(stepState.QuizItems))
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) IsShuffleFlashcardStudyQuizzes(ctx context.Context) (bool, error) {
	stepState := StepStateFromContext(ctx)
	query := `SELECT count(*) FROM quiz_sets qs INNER JOIN flashcard_progressions flp ON qs.quiz_set_id = flp.original_quiz_set_id WHERE flp.study_set_id = $1 AND qs.quiz_external_ids = flp.quiz_external_ids`

	var c pgtype.Int8
	if err := s.DB.QueryRow(ctx, query, stepState.StudySetID).Scan(&c); err != nil {
		return false, err
	}

	if c.Int == 0 {
		// there is no quiz_set.quiz_external_ids match the same sequence with shuffled_quiz_set.quiz_external_ids
		// so quiz_external_ids is shuffled
		return true, nil
	}

	return false, nil
}

func (s *suite) userCreateFlashcardStudyWithAndFlashcardStudyQuizzesPerPage(ctx context.Context, action, limitStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.StudyPlanItemID = idutil.ULIDNow()

	var keepOrder bool
	switch action {
	case shuffledAction:
		keepOrder = false
	case keepOrderAction:
		keepOrder = true
	}

	limit, _ := strconv.Atoi(limitStr)
	stepState.Limit = limit
	stepState.Offset = 1
	stepState.SetID = ""
	nextPage := &cpb.Paging{
		Limit: uint32(stepState.Limit),
		Offset: &cpb.Paging_OffsetInteger{
			OffsetInteger: int64(stepState.Offset),
		},
	}

	for {
		req := &pb.CreateFlashCardStudyRequest{
			StudySetId:      stepState.StudySetID,
			LoId:            stepState.LoID,
			StudentId:       stepState.StudentID,
			StudyPlanItemId: stepState.StudyPlanItemID,
			Paging:          nextPage,
			KeepOrder:       keepOrder,
		}
		resp, err := pb.NewQuizModifierServiceClient(s.Conn).CreateFlashCardStudy(contextWithToken(s, ctx), req)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		items := resp.Items
		nextPage = resp.NextPage
		stepState.StudySetID = resp.StudySetId
		if len(items) == 0 {
			break
		}
		if len(items) > 0 && len(items) > stepState.Limit {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect %v flashcard study quizzes but got %v", limit, len(items))
		}
		for _, item := range items {
			stepState.QuizItems = append(stepState.QuizItems, item.Item)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getExternalIDsFromQuizItems(items []*cpb.Quiz) []string {
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.Core.ExternalId
	}
	return ids
}

func (s *suite) userCreateFlashcardStudyTestWithValidRequestAndLimitInTheFirstTime(ctx context.Context, limit string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.OldToken = stepState.AuthToken
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx, err := s.aValidCourseAndStudyPlanBackground(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = stepState.SchoolAdminToken

	if stepState.LoID == "" {
		ctx, err = s.userCreateLearningObjectives(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("userCreateLearningObjectives: %w", err)
		}
	}
	resp, err := pb.NewTopicReaderServiceClient(s.Conn).ListToDoItemsByTopics(s.signedCtx(ctx), &pb.ListToDoItemsByTopicsRequest{
		StudyPlanId: wrapperspb.String(stepState.StudyPlanID),
		TopicIds:    []string{stepState.TopicID},
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("ListToDoItemsByTopics: %w", err)
	}
	for _, item := range resp.Items {
		flag := false
		for _, todoItem := range item.TodoItems {
			if todoItem.Type == pb.ToDoItemType_TO_DO_ITEM_TYPE_LO {
				stepState.StudyPlanItemID = todoItem.StudyPlanItem.StudyPlanItemId
				stepState.LoID = todoItem.ResourceId
				flag = true
				break
			}
		}
		if flag {
			break
		}
	}
	stepState.Limit, _ = strconv.Atoi(limit)
	stepState.Offset = 1
	stepState.NextPage = nil
	stepState.SetID = ""

	request := &pb.CreateFlashCardStudyRequest{
		StudyPlanItemId: stepState.StudyPlanItemID,
		LoId:            stepState.LoID,
		StudentId:       stepState.StudentID,
		Paging: &cpb.Paging{
			Limit: uint32(stepState.Limit),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
		KeepOrder: false,
	}
	stepState.AuthToken = stepState.OldToken
	return s.executeCreateFlashcardStudyTestServiceV2(ctx, request)
}
