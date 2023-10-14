package eureka

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
	yasuoPb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) createACourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CourseID = idutil.ULIDNow()
	_, err := yasuoPb.NewCourseServiceClient(s.YasuoConn).UpsertCourses(s.signedCtx(ctx), &yasuoPb.UpsertCoursesRequest{
		Courses: []*yasuoPb.UpsertCoursesRequest_Course{
			{
				Id:       stepState.CourseID,
				Name:     "course",
				Country:  1,
				Subject:  bpb.SUBJECT_BIOLOGY,
				SchoolId: int32(constants.ManabieSchool),
				BookIds:  []string{stepState.BookID},
			},
		},
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create course: %w", err)
	}

	if _, err := epb.NewCourseModifierServiceClient(s.Conn).AddBooks(s.signedCtx(ctx), &epb.AddBooksRequest{
		BookIds:  []string{stepState.BookID},
		CourseId: stepState.CourseID,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to add books: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createACourseBook(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := &entities.CoursesBooks{}
	database.AllNullEntity(e)
	if stepState.CourseID == "" {
		stepState.CourseID = idutil.ULIDNow()
	}
	now := time.Now()
	if err := multierr.Combine(
		e.CourseID.Set(stepState.CourseID),
		e.BookID.Set(stepState.BookID),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to set value: %w", err)
	}
	if _, err := database.Insert(ctx, e, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create course: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreateACourseWithAStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.createACourse(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := &epb.UpsertStudyPlanRequest{
		Name:                "study-plan",
		CourseId:            stepState.CourseID,
		BookId:              stepState.BookID,
		SchoolId:            constants.ManabieSchool,
		Status:              epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
		TrackSchoolProgress: true,
		Grades:              []int32{1, 2},
	}
	resp, err := epb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlan(s.signedCtx(ctx), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create study plan: %w", err)
	}
	stepState.Request = req
	stepState.StudyPlanID = resp.StudyPlanId
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studyPlanOfStudentHaveStoredCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*epb.UpsertStudyPlanRequest)
	var (
		studyPlan entities.StudyPlan
	)
	if err := try.Do(func(attempt int) (retry bool, err error) {
		sspE := &entities.StudentStudyPlan{}
		fields, _ := studyPlan.FieldMap()
		fieldList := make([]string, 0, len(fields))
		for _, field := range fields {
			fieldList = append(fieldList, fmt.Sprintf("sp.%s", field))
		}
		stmt := fmt.Sprintf(`
			SELECT %s
			FROM %s AS sp
			JOIN %s AS ssp
			USING(study_plan_id)
			WHERE ssp.student_id = $1 AND sp.master_study_plan_id = $2
		`, strings.Join(fieldList, ","), studyPlan.TableName(), sspE.TableName())
		if err := s.DB.QueryRow(ctx, stmt, stepState.StudentID, stepState.StudyPlanID).Scan(database.GetScanFields(&studyPlan, fields)...); err != nil {
			if err.Error() == pgx.ErrNoRows.Error() {
				time.Sleep(5 * time.Second)
				return attempt < 5, fmt.Errorf("study plan not created yet")
			}
			return false, fmt.Errorf("unable to retrieve study plan of student: %w", err)
		}
		return false, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if req.BookId != studyPlan.BookID.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("book have stored wrong, expect %v but got %v", req.BookId, studyPlan.BookID.String)
	}
	if req.CourseId != studyPlan.CourseID.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("course have stored wrong, expect %v but got %v", req.CourseId, studyPlan.CourseID.String)
	}
	if req.Status.String() != studyPlan.Status.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("status have stored wrong, expect %v but got %v", req.Status.String(), studyPlan.Status.String)
	}
	if req.TrackSchoolProgress != studyPlan.TrackSchoolProgress.Bool {
		return StepStateToContext(ctx, stepState), fmt.Errorf("track school progress have stored wrong, expect %v but got %v", req.TrackSchoolProgress, studyPlan.TrackSchoolProgress.Bool)
	}
	grades := make([]int32, 0, len(studyPlan.Grades.Elements))
	for _, e := range studyPlan.Grades.Elements {
		grades = append(grades, e.Int)
	}
	if !reflect.DeepEqual(req.Grades, grades) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("grades have stored wrong, expect %v but got %v", req.Grades, grades)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userAddCourseToStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stmt := `SELECT email FROM users WHERE user_id = $1`
	var studentEmail string
	err := s.BobDB.QueryRow(ctx, stmt, stepState.StudentID).Scan(&studentEmail)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if stepState.SchoolAdminToken != "" {
		stepState.AuthToken = stepState.SchoolAdminToken
	}

	ctx = s.signedCtx(ctx)

	locationID := idutil.ULIDNow()
	e := &bob_entities.Location{}
	database.AllNullEntity(e)
	if err := multierr.Combine(
		e.LocationID.Set(locationID),
		e.Name.Set(fmt.Sprintf("location-%s", locationID)),
		e.IsArchived.Set(false),
		e.CreatedAt.Set(time.Now()),
		e.UpdatedAt.Set(time.Now()),
	); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if _, err := database.Insert(ctx, e, s.BobDB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	_, err = upb.NewUserModifierServiceClient(s.UsermgmtConn).UpdateStudent(
		ctx,
		&upb.UpdateStudentRequest{
			StudentProfile: &upb.UpdateStudentRequest_StudentProfile{
				Id:               stepState.StudentID,
				Name:             "test-name",
				Grade:            5,
				EnrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
				Email:            studentEmail,
				LocationIds:      []string{locationID},
			},

			SchoolId: stepState.SchoolIDInt,
		},
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update student: %w", err)
	}

	if _, err := upb.NewUserModifierServiceClient(s.UsermgmtConn).UpsertStudentCoursePackage(ctx, &upb.UpsertStudentCoursePackageRequest{
		StudentId: stepState.StudentID,
		StudentPackageProfiles: []*upb.UpsertStudentCoursePackageRequest_StudentPackageProfile{{
			Id: &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
				CourseId: stepState.CourseID,
			},
			StartTime: timestamppb.New(time.Now().Add(time.Hour * -20)),
			EndTime:   timestamppb.New(time.Now().Add(time.Hour * 24 * 10)),
		}},
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update student course package: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}
