package enigma

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/enigma/dto"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
)

func (s *suite) requestUserCourseRegistration(ctx context.Context, students int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.CurrentUserID = idutil.ULIDNow()
	now := time.Now()
	studentLessons := make([]dto.StudentLesson, 0, students)
	// student
	for i := 1; i <= students; i++ {
		studentLessons = append(studentLessons, dto.StudentLesson{
			StudentID:  strconv.Itoa(i),
			LessonIDs:  []int{i},
			ActionKind: dto.ActionKindUpserted,
		})
	}

	request := &dto.SyncUserCourseRequest{
		Timestamp: int(now.Unix()),
		Payload: struct {
			StudentLessons []dto.StudentLesson `json:"student_lesson"`
		}{
			StudentLessons: studentLessons,
		},
	}

	s.Request = request

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) requestUserCourseRegistrationInvalidPayload(ctx context.Context, students int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.CurrentUserID = idutil.ULIDNow()
	now := time.Now()
	studentLessons := make([]dto.StudentLesson, 0, students)
	// student
	for i := 1; i <= students; i++ {
		studentLessons = append(studentLessons, dto.StudentLesson{
			ActionKind: dto.ActionKindUpserted,
		})
	}

	request := &dto.SyncUserCourseRequest{
		Timestamp: int(now.Unix()),
		Payload: struct {
			StudentLessons []dto.StudentLesson `json:"student_lesson"`
		}{
			StudentLessons: studentLessons,
		},
	}

	s.Request = request

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) stepPerformUserCourseRegistrationRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	url := fmt.Sprintf("%s/jprep/user-course", s.EnigmaSrvURL)
	bodyBytes, err := s.makeHTTPRequest(http.MethodPut, url)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if bodyBytes == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("body is nil")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) logCorrectStudentLessons(ctx context.Context, payload []byte) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentLessons := []*npb.EventSyncUserCourse_StudentLesson{}
	err := json.Unmarshal(payload, &studentLessons)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("json.Unmarshal student lessons: %w", err)
	}
	for _, studentLesson := range studentLessons {
		found := false
		for _, req := range s.Request.(*dto.SyncUserCourseRequest).Payload.StudentLessons {
			if studentLesson.StudentId == req.StudentID {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("can't find student lessons %s", studentLesson.StudentId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
