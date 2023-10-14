package bob

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/lestrrat-go/jwx/jwt"
	"go.uber.org/multierr"
)

func (s *suite) aLessonWithSomeLessonMembers(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.StudentIds = []string{} // reset
	stepState.courseIds = []string{}  // reset

	ctx, err1 := s.aTeacherAndAClassWithSomeStudents(ctx)
	ctx, err2 := s.aListOfCoursesAreExistedInDBOf(ctx, "above teacher")
	ctx, err3 := s.aStudentWithValidLesson(ctx)

	if err := multierr.Combine(err1, err2, err3); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var ids []string
	now := time.Now()
	total := rand.Intn(10) + 5 // minimum 5 student
	for i := 0; i < total; i++ {
		s.aSignedInStudent(ctx)
		t, _ := jwt.ParseString(stepState.AuthToken)
		userID := t.Subject()
		if i%2 == 0 {
			query := "UPDATE users SET last_login_date = $1 WHERE user_id = $2"
			if _, err := s.DB.Exec(ctx, query, time.Now().UTC().Add(-time.Hour), &userID); err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}

		stepState.StudentIds = append(stepState.StudentIds, userID)
		stepState.courseIds = append(stepState.courseIds, stepState.CurrentLessonID)
		lessonMember := &entities_bob.LessonMember{}
		database.AllNullEntity(lessonMember)
		if err := multierr.Combine(
			lessonMember.LessonID.Set(stepState.CurrentLessonID),
			lessonMember.UserID.Set(userID),
			lessonMember.CreatedAt.Set(now),
			lessonMember.UpdatedAt.Set(now),
		); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if _, err := database.Insert(ctx, lessonMember, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		ids = append(ids, userID)
	}

	stepState.Request = ids
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) listStudentsInLesson(ctx context.Context) (context.Context, error) {
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

		resp, err := bpb.NewClassReaderServiceClient(s.Conn).ListStudentsByLesson(s.signedCtx(ctx), &bpb.ListStudentsByLessonRequest{
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
func (s *suite) teacherListStudentsInThatLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	stepState.AuthToken, err = s.generateExchangeToken(stepState.CurrentTeacherID, entities_bob.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	return s.listStudentsInLesson(ctx)
}
func (s *suite) validateSystemStoreCourseMembers(ctx context.Context, expectedStudentIds []string, students []*cpb.BasicProfile) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(students) != len(expectedStudentIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("total students mismatch, got: %d, want: %d", len(students), len(expectedStudentIds))
	}

	query := `SELECT user_id, (concat(given_name||' ')||name), last_login_date
				FROM users WHERE user_id = ANY($1) ORDER BY concat(given_name || ' ', name)  COLLATE "C" ASC, user_id ASC`
	rows, err := s.DB.Query(ctx, query, expectedStudentIds)

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
func (s *suite) returnsAListOfStudents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	expectedStudentIDs := stepState.Request.([]string)
	students := stepState.Response.([]*cpb.BasicProfile)
	return s.validateSystemStoreCourseMembers(ctx, expectedStudentIDs, students)
}
func (s *suite) studentListStudentsInThatLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aSignedInStudent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	return s.listStudentsInLesson(ctx)
}
func (s *suite) someStudentsHasBeenRemovedFromTheLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	r := &repositories.LessonMemberRepo{}
	totalStudentRemove := rand.Intn(5) + 1
	for i := 0; i < totalStudentRemove; i++ {
		stepState.studentRemovedIds = append(stepState.studentRemovedIds, stepState.StudentIds[i])
		if err := r.SoftDelete(ctx, s.DB, database.Text(stepState.StudentIds[i]), database.TextArray(golibs.Uniq(stepState.courseIds))); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable remove student: %w", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) ourSystemHaveToReturnsAListOfStudentsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	expectedStudents := make([]string, 0, len(stepState.StudentIds)-len(stepState.studentRemovedIds))
	for _, id := range stepState.StudentIds {
		if !contains(stepState.studentRemovedIds, id) {
			expectedStudents = append(expectedStudents, id)
		}
	}
	students := stepState.Response.([]*cpb.BasicProfile)
	return s.validateSystemStoreCourseMembers(ctx, expectedStudents, students)
}
func contains(s []string, target string) bool {
	for _, val := range s {
		if target == val {
			return true
		}
	}
	return false
}
func (s *suite) aLessonWithSomeLessonMembersWithTheirName(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.StudentIds = []string{} // reset
	stepState.courseIds = []string{}  // reset

	ctx, err1 := s.aTeacherAndAClassWithSomeStudents(ctx)
	ctx, err2 := s.aStudentWithValidLesson(ctx)
	if err := multierr.Combine(err1, err2); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var ids []string

	now := time.Now()
	total := rand.Intn(10) + 10 // minimum 10 student
	name := generateRandomName()
	tempName := generateRandomName()
	for i := 0; i < total; i++ {
		s.aSignedInStudentGivenName(ctx, name)
		if i > (total / 2) {
			name = tempName
			if ctx, err := s.updateGivenName(ctx, generateRandomName()); err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
		t, _ := jwt.ParseString(stepState.AuthToken)

		stepState.StudentIds = append(stepState.StudentIds, t.Subject())
		stepState.courseIds = append(stepState.courseIds, stepState.CurrentLessonID)
		lessonMember := &entities_bob.LessonMember{}
		database.AllNullEntity(lessonMember)
		if err := multierr.Combine(
			lessonMember.LessonID.Set(stepState.CurrentLessonID),
			lessonMember.UserID.Set(t.Subject()),
			lessonMember.CreatedAt.Set(now),
			lessonMember.UpdatedAt.Set(now),
		); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if _, err := database.Insert(ctx, lessonMember, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		ids = append(ids, t.Subject())
	}

	stepState.Request = ids
	return StepStateToContext(ctx, stepState), nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateRandomName() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, rand.Intn(15)+5)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
func (s *suite) updateGivenName(ctx context.Context, givenName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	updateGivenName := `UPDATE users SET given_name = $1 WHERE user_id = $2`
	cTag, err := s.DB.Exec(ctx, updateGivenName, givenName, stepState.CurrentUserID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable update given name for user %s: %w", stepState.CurrentUserID, err)
	}

	if cTag.RowsAffected() != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable update given name: unexpected row effected")
	}

	return StepStateToContext(ctx, stepState), nil

}
