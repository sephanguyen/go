package lessonmgmt

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/helper"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) studentListInThatLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	paging := &cpb.Paging{
		Limit: uint32(rand.Intn(4)) + 1,
	}
	var students []*cpb.BasicProfile
	idx := 0
	for {
		if idx > 100 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected paging: infinite paging")
		}
		idx++
		resp, err := lpb.NewLessonReaderServiceClient(s.Connections.LessonMgmtConn).
			RetrieveStudentsByLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), &lpb.ListStudentsByLessonRequest{
				LessonId: stepState.CurrentLessonID,
				Paging:   paging,
			})
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if len(resp.Students) == 0 {
			break
		}
		if len(resp.Students) > int(paging.Limit) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total students: got: %d, want: %d", len(resp.Students), paging.Limit)
		}

		students = append(students, resp.Students...)

		paging = resp.NextPage
	}
	stepState.Response = students
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsAListOfStudents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	expectedStudentIDs := make([]string, 0, len(stepState.StudentIDWithCourseID)/2)
	for i := 0; i < len(stepState.StudentIDWithCourseID); i += 2 {
		expectedStudentIDs = append(expectedStudentIDs, stepState.StudentIDWithCourseID[i])
	}
	students := stepState.Response.([]*cpb.BasicProfile)
	return s.validateSystemStoreCourseMembers(ctx, expectedStudentIDs, students)
}

func (s *Suite) validateSystemStoreCourseMembers(ctx context.Context, expectedStudentIds []string, students []*cpb.BasicProfile) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(students) != len(expectedStudentIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("total students mismatch, got: %d, want: %d", len(students), len(expectedStudentIds))
	}

	query := `SELECT user_id, (concat(given_name||' ')||name), last_login_date
				FROM users WHERE user_id = ANY($1) ORDER BY concat(given_name || ' ', name)  COLLATE "C" ASC, user_id ASC`
	rows, err := s.BobDBTrace.Query(ctx, query, expectedStudentIds)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()

	var idx int
	for rows.Next() {
		var expectedUserID, expectedName string
		expectedLastLoginDate := &time.Time{}
		if err := rows.Scan(&expectedUserID, &expectedName, &expectedLastLoginDate); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		student := students[idx]
		if student.UserId != expectedUserID {
			return StepStateToContext(ctx, stepState),
				fmt.Errorf("student id mistmatch, got: %q, want: %q", student.UserId, expectedUserID)
		}
		if student.Name != expectedName {
			return StepStateToContext(ctx, stepState),
				fmt.Errorf("student name mistmatch, got: %q, want: %q", student.Name, expectedName)
		}
		if (expectedLastLoginDate == nil && student.LastLoginDate != nil) ||
			(expectedLastLoginDate != nil && !expectedLastLoginDate.Equal(student.LastLoginDate.AsTime())) {
			return StepStateToContext(ctx, stepState),
				fmt.Errorf("student last_login_date mistmatch, got: %q, want: %q", student.LastLoginDate.AsTime(), expectedLastLoginDate)
		}

		idx++
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
