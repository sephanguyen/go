package eureka

import (
	"context"
	"fmt"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/segmentio/ksuid"
	"go.uber.org/multierr"
)

func (s *suite) eurekaMustReturnCorrectListOfBasicProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error when call ListStudentByCourse %w", stepState.ResponseErr)
	}
	resp := stepState.Response.(*epb.ListStudentByCourseResponse)
	if (stepState.NumberOfId) != len(resp.Profiles) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("total student retrieved not correctly, expected %d - got %d", stepState.NumberOfId, len(resp.Profiles))
	}
	for i := 1; i < len(resp.Profiles); i++ {
		if resp.Profiles[i].Name < resp.Profiles[i-1].Name {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error alphabet sort by student name")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) callMultiListStudentByCoursePaging(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseReader := epb.NewCourseReaderServiceClient(conn)
	limit := 5
	resp, err := courseReader.ListStudentByCourse(contextWithToken(s, ctx), &epb.ListStudentByCourseRequest{
		CourseId: stepState.CourseID,
		Paging:   &cpb.Paging{Limit: 5},
	})

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	resp2, err2 := courseReader.ListStudentByCourse(contextWithToken(s, ctx), &epb.ListStudentByCourseRequest{
		CourseId: stepState.CourseID,
		Paging:   resp.NextPage,
	})
	stepState.Response = resp2
	stepState.ResponseErr = err2
	stepState.NumberOfId -= limit
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) CallListStudentByCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aValidAuthenticationToken(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	courseReader := epb.NewCourseReaderServiceClient(conn)
	resp, err := courseReader.ListStudentByCourse(contextWithToken(s, ctx), &epb.ListStudentByCourseRequest{
		CourseId: stepState.CourseID,
		Paging:   &cpb.Paging{Limit: 100},
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Response = resp
	stepState.ResponseErr = err
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) CallListStudentByCourseWithSearchText(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aValidAuthenticationToken(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	courseReader := epb.NewCourseReaderServiceClient(conn)
	resp, err := courseReader.ListStudentByCourse(contextWithToken(s, ctx), &epb.ListStudentByCourseRequest{
		CourseId:   stepState.CourseID,
		SearchText: "に歩",
		Paging:     &cpb.Paging{Limit: 1},
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Response = resp
	stepState.ResponseErr = err
	stepState.NumberOfId = 1
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidCourseStudentBackground(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseID := idutil.ULIDNow()
	stepState.CourseID = courseID
	// insert multi user to bob db
	if ctx, err := s.insertMultiUserIntoBob(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	// get all user have just inserted
	ctx, err := s.aValidCourseWithIds(ctx, stepState.StudentIDs, courseID)

	return StepStateToContext(ctx, stepState), err
}
func (s *suite) insertMultiUserIntoBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.NumberOfId = 10
	studentIDs := make([]string, 0, stepState.NumberOfId)
	for i := 0; i < stepState.NumberOfId; i++ {
		if ctx, err := s.insertUserIntoBob(ctx); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		studentIDs = append(studentIDs, stepState.UserId)
	}
	stepState.StudentIDs = studentIDs
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) getUserIds(ctx context.Context) (context.Context, []string, error) {
	stepState := StepStateFromContext(ctx)
	limit := stepState.NumberOfId
	ids := make([]string, 0, limit)
	query := fmt.Sprintf(`SELECT user_id from "users" order by user_id DESC limit %d`, limit)
	rows, err := s.DB.Query(ctx, query)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, fmt.Errorf("error when query user_id from bob db %w", err)
	}
	defer rows.Close()

	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return StepStateToContext(ctx, stepState), nil, fmt.Errorf("rows.Err: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), nil, fmt.Errorf("rows.Err: %w", err)
	}
	return StepStateToContext(ctx, stepState), ids, nil
}
func (s *suite) aJapaneseStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, ids, err := s.getUserIds(ctx)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	var studentID string
	if len(ids) > 0 {
		studentID = ids[0]
	} else {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not found any student")
	}
	stmtTpl := `UPDATE "users" SET name='に歩きたい乗り' WHERE user_id=$1::TEXT`
	_, err = s.DB.Exec(ctx, stmtTpl, studentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) insertUserIntoBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	now := time.Now()
	stepState.UserId = idutil.ULIDNow()
	user := &entities_bob.User{}
	database.AllNullEntity(user)
	userName := "valid-user-import-by-eureka" + stepState.UserId
	num := idutil.ULIDNow()

	err := multierr.Combine(
		user.Country.Set("COUNTRY_VN"),
		user.PhoneNumber.Set(fmt.Sprintf("+849%s", num)),
		user.Email.Set(fmt.Sprintf("valid-%s@email.com", num)),
		user.LastName.Set(userName),
		user.Group.Set("USER_GROUP_STUDENT"),
		user.ID.Set(stepState.UserId),
		user.CreatedAt.Set(now),
		user.UpdatedAt.Set(now),
		user.ResourcePath.Set(1),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil
	}
	_, err = database.Insert(ctx, user, s.BobDB.Exec)

	return StepStateToContext(ctx, stepState), err
}
func (s *suite) aValidCourseWithIds(ctx context.Context, Ids []string, courseID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if courseID == "" {
		stepState.CourseID = courseID
	}

	for i := 0; i < len(Ids); i++ {
		courseStudent, err := generateCourseByStudentId(Ids[i], courseID)
		stepState.CourseStudents = append(stepState.CourseStudents, courseStudent)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		cmd, err := database.Insert(ctx, courseStudent, s.DB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if cmd.RowsAffected() != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error insert course student")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func generateCourseByStudentId(studentID, courseID string) (*entities.CourseStudent, error) {
	var c entities.CourseStudent
	database.AllNullEntity(&c)
	now := timeutil.Now()
	err := multierr.Combine(
		c.ID.Set(ksuid.New().String()),
		c.CourseID.Set(courseID),
		c.StudentID.Set(studentID),
		c.CreatedAt.Set(now),
		c.UpdatedAt.Set(now),
		c.StartAt.Set(now.Add(-time.Hour)),
		c.EndAt.Set(now.Add(time.Hour*24*5)),
	)
	return &c, err
}
