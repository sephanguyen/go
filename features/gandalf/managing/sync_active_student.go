package managing

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	eurekaCmd "github.com/manabie-com/backend/cmd/server/eureka"
	"github.com/manabie-com/backend/features/yasuo"
	"github.com/manabie-com/backend/internal/eureka/configurations"
	eureka_entities "github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	bobPb "github.com/manabie-com/backend/pkg/genproto/bob"
	yasuoPb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) schooldAdminCreatesAnExistedCourse(ctx context.Context) (context.Context, error) {
	yasuoState := yasuo.StepStateFromContext(ctx)

	stepState := GandalfStepStateFromContext(ctx)
	stepState.YasuoStepState.CurrentCourseID = idutil.ULIDNow()
	id := rand.Int31()

	country := bobPb.COUNTRY_VN
	grade, _ := i18n.ConvertIntGradeToString(country, 7)
	req := &yasuoPb.UpsertCoursesRequest{
		Courses: []*yasuoPb.UpsertCoursesRequest_Course{
			{
				Id:       stepState.YasuoStepState.CurrentCourseID,
				Name:     fmt.Sprintf("course-%d", id),
				Country:  country,
				Subject:  bobPb.SUBJECT_BIOLOGY,
				SchoolId: yasuoState.CurrentSchoolID,
				Grade:    grade,
			},
		},
	}
	stepState.GandalfStateResponse, stepState.GandalfStateResponseErr = yasuoPb.NewCourseServiceClient(s.yasuoConn).
		UpsertCourses(contextWithToken(ctx, yasuoState.AuthToken), req)
	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) schoolAdminCreatesStudentWithCoursePackage(ctx context.Context) (context.Context, error) {
	yasuoState := yasuo.StepStateFromContext(ctx)

	stepState := GandalfStepStateFromContext(ctx)
	randomId := idutil.ULIDNow()
	stepState.GandalfStateRequest = &ypb.CreateStudentRequest{
		SchoolId: 1,
		StudentProfile: &ypb.CreateStudentRequest_StudentProfile{
			Email:            fmt.Sprintf("%v@example.com", randomId),
			Password:         fmt.Sprintf("password-%v", randomId),
			Name:             fmt.Sprintf("user-%v", randomId),
			CountryCode:      cpb.Country_COUNTRY_VN,
			EnrollmentStatus: cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			PhoneNumber:      fmt.Sprintf("phone-number-%v", randomId),
			Grade:            5,
		},
		StudentPackageProfiles: []*ypb.CreateStudentRequest_StudentPackageProfile{
			{
				CourseId: stepState.YasuoStepState.CurrentCourseID,
				Start:    timestamppb.Now(),
				End:      timestamppb.New(time.Now().Add(time.Hour)),
			},
		},
	}
	stepState.GandalfStateResponse, stepState.GandalfStateResponseErr = ypb.NewUserModifierServiceClient(s.yasuoConn).
		CreateStudent(contextWithToken(ctx, yasuoState.AuthToken), stepState.GandalfStateRequest.(*ypb.CreateStudentRequest))
	return GandalfStepStateToContext(ctx, stepState), nil
}
func (s *suite) eurekaStoreStudentCourseInfo(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	getStudentCourseInfoFunc := func() error {
		e := &eureka_entities.CourseStudent{}
		fields := database.GetFieldNames(e)
		stmt :=
			fmt.Sprintf(
				`
		SELECT %s 
		FROM
			course_students
		WHERE
			student_id = $1 AND deleted_at IS NULL;
	`, strings.Join(fields, ","))

		rows, err := s.eurekaDB.Query(
			ctx,
			stmt,
			stepState.GandalfStateResponse.(*ypb.CreateStudentResponse).StudentProfile.Student.UserProfile.UserId,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		courseStudents := []*eureka_entities.CourseStudent{}
		for rows.Next() {
			cs := &eureka_entities.CourseStudent{}

			err := rows.Scan(database.GetScanFields(cs, database.GetFieldNames(cs))...)
			if err != nil {
				return err

			}
			courseStudents = append(courseStudents, cs)
		}

		req := stepState.GandalfStateRequest.(*ypb.CreateStudentRequest)
		studentPackages := req.StudentPackageProfiles
		if len(studentPackages) != len(courseStudents) {
			return fmt.Errorf("expect student has %d number of course but got %d", len(studentPackages), len(courseStudents))
		}
		for i := range courseStudents {
			c := courseStudents[i]
			pk := studentPackages[i]
			if c.CourseID.String != pk.CourseId {
				return fmt.Errorf("expect course id in student package %s but got course id %s", pk.CourseId, c.CourseID.String)
			}
			if !pk.Start.AsTime().Round(time.Second).Equal(c.StartAt.Time.Round(time.Second)) {
				return fmt.Errorf("expect student package start at %v but got %v", pk.Start.AsTime(), c.StartAt.Time)
			}
			if !pk.End.AsTime().Round(time.Second).Equal(c.EndAt.Time.Round(time.Second)) {
				return fmt.Errorf("expect student package end at %v but got %v", pk.End.AsTime(), c.EndAt.Time)
			}
		}
		return nil
	}

	err := try.Do(func(attempt int) (bool, error) {
		err := getStudentCourseInfoFunc()
		if err == nil {
			return false, nil
		}

		time.Sleep(200 * time.Millisecond)
		return attempt < 5, nil
	})
	return ctx, err
}

func (s *suite) deleteStartDateAndEndDateOfThisStudentCourse(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	setStartAtToNullFunc := func() error {
		stmt := `
		UPDATE course_students 
		SET start_at = NULL,
		end_at = NULL
		WHERE student_id = $1 AND course_id = $2 AND deleted_at IS NULL;
	`

		studentID := stepState.GandalfStateResponse.(*ypb.CreateStudentResponse).StudentProfile.Student.UserProfile.UserId
		courseID := stepState.YasuoStepState.CurrentCourseID

		cmd, err := s.eurekaDB.Exec(
			ctx,
			stmt,
			database.Text(studentID),
			database.Text(courseID),
		)
		if err != nil {
			return err

		}
		if cmd.RowsAffected() != 1 {
			return fmt.Errorf("no row affected")
		}
		return nil
	}

	err := try.Do(func(attempt int) (bool, error) {
		err := setStartAtToNullFunc()
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			return attempt < 10, err
		}

		return false, nil
	})

	return ctx, err
}
func (s *suite) runMigrationToolSyncActiveStudent(ctx context.Context) (context.Context, error) {
	yasuoState := yasuo.StepStateFromContext(ctx)
	rsc := bootstrap.NewResources().WithServiceName(s.Cfg.Common.Name).WithDatabaseC(ctx, s.Cfg.PostgresV2.Databases).WithLoggerC(&s.Cfg.Common)
	_ = eurekaCmd.RunSyncActiveStudent(contextWithToken(ctx, yasuoState.AuthToken), configurations.Config{
		Common:     s.Cfg.Common,
		PostgresV2: s.Cfg.PostgresV2,
	}, rsc)
	return ctx, nil
}
func (s *suite) afterSyncActiveStudentEurekaStoreCorrectStudentCourseInfo(ctx context.Context) (context.Context, error) {
	return s.eurekaStoreStudentCourseInfo(ctx)
}
