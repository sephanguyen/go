package lessonmgmt

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) aValidCourseTeachingTimePayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rows := []string{}
	addedCourseIDs := make(map[string]bool)
	header := "course_id,course_name,preparation_time,break_time,action"

	for i := 0; i < len(stepState.StudentIDWithCourseID); i += 2 {
		courseID := stepState.StudentIDWithCourseID[i+1]
		if _, ok := addedCourseIDs[courseID]; ok {
			continue
		}
		addedCourseIDs[courseID] = true
		prepTime, _ := rand.Int(rand.Reader, big.NewInt(300))
		breakTime, _ := rand.Int(rand.Reader, big.NewInt(60))
		rows = append(rows, fmt.Sprintf("%s,,%d,%d,upsert", courseID, prepTime, breakTime))
	}

	stepState.Request = &lpb.ImportCourseTeachingTimeRequest{
		Payload: []byte(fmt.Sprintf(`%s
		%s
		`, header, strings.Join(rows, "\n"))),
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) importCourseTeachingTime(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = lpb.NewLessonExecutorServiceClient(s.LessonMgmtConn).
		ImportCourseTeachingTime(contextWithToken(s, ctx), stepState.Request.(*lpb.ImportCourseTeachingTimeRequest))
	return StepStateToContext(ctx, stepState), nil
}
