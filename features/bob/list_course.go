package bob

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/lestrrat-go/jwx/jwt"
	"go.uber.org/multierr"
)

func (s *suite) aListCoursesRequestMessageSchool(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if arg1 == "manabie" {
		stepState.Request = &bpb.ListCoursesRequest{Filter: &cpb.CommonFilter{SchoolId: constant.ManabieSchool}}
	} else {
		stepState.Request = &bpb.ListCoursesRequest{}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aListCoursesRequestKeyword(ctx context.Context, keyword string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &bpb.ListCoursesRequest{Filter: &cpb.CommonFilter{SchoolId: constant.ManabieSchool}, Keyword: keyword}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsCoursesInListCoursesResponseMatchingFilterOfListCoursesRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, courses := range stepState.PaginatedCourses {
		for _, course := range courses {
			if course.Info.SchoolId != constants.ManabieSchool {
				return StepStateToContext(ctx, stepState), fmt.Errorf("school id in course does not match, expected: %v, got: %v", constants.ManabieSchool, course.Info.SchoolId)
			}
		}
	}
	return s.checkListCourseOrder(ctx)
}
func (s *suite) userListCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = contextWithToken(s, ctx)

	paging := &cpb.Paging{
		Limit: 10,
	}

	stepState.ResponseErr = nil
	stepState.PaginatedCourses = nil
	for {
		stepState.Request.(*bpb.ListCoursesRequest).Paging = paging
		resp, err := bpb.NewCourseReaderServiceClient(s.Conn).ListCourses(ctx, stepState.Request.(*bpb.ListCoursesRequest))
		if err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), nil
		}
		if len(resp.Items) == 0 {
			break
		}
		if len(resp.Items) > int(paging.Limit) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total course: got: %d, want: %d", len(resp.Items), paging.Limit)
		}
		stepState.PaginatedCourses = append(stepState.PaginatedCourses, resp.Items)

		paging = resp.NextPage
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aboveTeacherBelongToSchool(ctx context.Context, school string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var schoolID int32
	switch school {
	case "manabie":
		schoolID = constant.ManabieSchool
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("school %s does not specify", school)
	}

	t, _ := jwt.ParseString(stepState.AuthToken)
	_, err := s.DB.Exec(ctx, `UPDATE teachers SET school_ids = $1 WHERE teacher_id = $2`, []int32{schoolID}, t.Subject())

	return StepStateToContext(ctx, stepState), err
}
func (s *suite) returnsResponseOfListCoursesHaveToCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.returnsCoursesInListCoursesResponseMatchingFilterOfListCoursesRequest(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	counter := 0
	var courseIDWithIcon []string
	limitedCourseIds := stepState.courseIds[:stepState.numberCourseHaveIcon]
	for _, courses := range stepState.PaginatedCourses {
		for _, course := range courses {
			if course.Info.IconUrl != "" && golibs.InArrayString(course.Info.Id, limitedCourseIds) {
				courseIDWithIcon = append(courseIDWithIcon, course.Info.Id)
			}
		}
	}
	courseIDWithIcon = golibs.GetUniqueElementStringArray(courseIDWithIcon)
	counter = len(courseIDWithIcon)
	if counter != stepState.numberCourseHaveIcon {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of icon return wrong, want %d, got %d", stepState.numberCourseHaveIcon, counter)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) someCoursesHaveIconUrl(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resetIconStmtTpl := `UPDATE courses SET icon= NULL`
	_, err := s.DB.Exec(ctx, resetIconStmtTpl)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("reset icon to NULL failed")
	}
	updateStmtTpl := `UPDATE courses SET icon = 'icon' WHERE course_id=ANY($1::_TEXT)`
	rand.Seed(time.Now().Unix())
	stepState.numberCourseHaveIcon = rand.Intn(5) + 1
	sort.Strings(stepState.courseIds)
	commandTag, err := s.DB.Exec(ctx, updateStmtTpl, stepState.courseIds[:stepState.numberCourseHaveIcon])
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed update icon value")
	}
	if commandTag.RowsAffected() == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no raw affected, failed update icon value")
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) checkListCourseOrder(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, courses := range stepState.PaginatedCourses {
		if courses == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("courses have to not null")
		}
		var tmp []*cpb.Course
		copy(tmp, courses)
		if !sort.SliceIsSorted(tmp, func(i, j int) bool {
			return tmp[i].Info.CreatedAt.GetNanos() > tmp[j].Info.CreatedAt.GetNanos()
		}) {
			return StepStateToContext(ctx, stepState), errors.New("courses are not sorted by created_at DESC")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsResponseForStudentListCoursesHaveToCorrectly(ctx context.Context) (context.Context, error) {
	return s.checkListCourseOrder(ctx)
}
func (s *suite) allCoursesAreBelongToAcademicYear(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	now := time.Now()
	rand := rand.Int()
	academicYearID := ""

	switch status {
	case "current":
		e := &entities.AcademicYear{}
		err := multierr.Combine(
			e.ID.Set(fmt.Sprintf("%d-%d-%d", constants.ManabieSchool, now.Year(), rand)),
			e.SchoolID.Set(constants.ManabieSchool),
			e.Name.Set(fmt.Sprintf("%d", now.Year())),
			e.StartYearDate.Set(now),
			e.EndYearDate.Set(now.Add(200*24*time.Hour)),
			e.Status.Set(entities.AcademicYearStatusActive),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}

		aRepo := &repositories.AcademicYearRepo{}
		err = aRepo.Create(ctx, s.DB, e)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}

		academicYearID = e.ID.String
	case "expired":
		now := now.Add(-365 * 24 * time.Hour)
		e := &entities.AcademicYear{}
		err := multierr.Combine(
			e.ID.Set(fmt.Sprintf("%d-%d-%d", constants.ManabieSchool, now.Year(), rand)),
			e.SchoolID.Set(constants.ManabieSchool),
			e.Name.Set(fmt.Sprintf("%d", now.Year())),
			e.StartYearDate.Set(now),
			e.EndYearDate.Set(now.Add(200*24*time.Hour)),
			e.Status.Set(entities.AcademicYearStatusInActive),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}

		aRepo := &repositories.AcademicYearRepo{}
		err = aRepo.Create(ctx, s.DB, e)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}

		academicYearID = e.ID.String
	}

	sql := `DELETE FROM courses_academic_years WHERE course_id = ANY($1)`
	_, err := s.DB.Exec(ctx, sql, stepState.courseIds)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	sql = `INSERT INTO courses_academic_years
SELECT course_id, $1, now(), now(), null
FROM courses WHERE course_id = ANY($2)`
	_, err = s.DB.Exec(ctx, sql, academicYearID, stepState.courseIds)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsNoCoursesInListCoursesResponse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for i := range stepState.PaginatedCourses {
		for j := range stepState.PaginatedCourses[i] {
			if golibs.InArrayString(stepState.PaginatedCourses[i][j].Info.Id, stepState.courseIds) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("don't expect course_id = %s in result", stepState.PaginatedCourses[i][j].Info.Id)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
