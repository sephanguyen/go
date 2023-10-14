package course_statistical

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/bob/constants"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) retrieveSchoolHistoryByStudentInCourse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	stepState.Response, stepState.ResponseErr = sspb.NewStatisticsClient(s.EurekaConn).RetrieveSchoolHistoryByStudentInCourse(ctx, &sspb.RetrieveSchoolHistoryByStudentInCourseRequest{
		CourseId: stepState.CourseID,
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentsExistsInSchoolHistory(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	for _, student := range stepState.Students {
		schoolID := idutil.ULIDNow()
		stepState.SchoolIDs = append(stepState.SchoolIDs, schoolID)

		if err := s.generateSchoolHistory(ctx, s.BobDB, student.ID, schoolID); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("generateSchoolHistory by student_id: %s, %w", student.ID, err)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) thereAreNumberSchoolInfo(ctx context.Context, num int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	resp := stepState.Response.(*sspb.RetrieveSchoolHistoryByStudentInCourseResponse)

	if got := len(resp.GetSchools()); got != num {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected to result of school_info is %d, got %d", num, got)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) generateSchoolHistory(ctx context.Context, db database.Ext, studentID, schoolID string) error {
	stepState := utils.StepStateFromContext[StepState](ctx)
	levelID := idutil.ULIDNow()
	locationID := idutil.ULIDNow()

	{ // School level
		sequence := rand.Intn(999999999)
		isArchived := false
		stmt := `INSERT INTO public.school_level (school_level_id, school_level_name, sequence, is_archived) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`
		_, err := db.Exec(ctx, stmt, levelID, levelID, sequence, isArchived)
		if err != nil {
			return fmt.Errorf("db.Exec: %v", err)
		}
	}

	{ // School Info
		schoolName := fmt.Sprintf("school name %s", schoolID)
		phonetic := fmt.Sprintf("phonetic %s", schoolID)
		stmt := `INSERT INTO school_info
		(school_id, school_name, school_name_phonetic, school_level_id, address, is_archived, school_partner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())`

		_, err := db.Exec(ctx, stmt, schoolID, schoolName, phonetic, levelID, "address", false, idutil.ULIDNow())
		if err != nil {
			return fmt.Errorf("db.Exec: %v", err)
		}
	}

	{
		e := &bob_entities.Location{}
		database.AllNullEntity(e)
		if err := multierr.Combine(
			e.LocationID.Set(locationID),
			e.Name.Set(fmt.Sprintf("location-%s", locationID)),
			e.IsArchived.Set(false),
			e.CreatedAt.Set(time.Now()),
			e.UpdatedAt.Set(time.Now()),
		); err != nil {
			return err
		}

		if _, err := database.Insert(ctx, e, db.Exec); err != nil {
			return fmt.Errorf("db.Exec: location %v", err)
		}
	}

	stmt := `SELECT email FROM users WHERE user_id = $1`
	var studentEmail string

	if err := db.QueryRow(ctx, stmt, studentID).Scan(&studentEmail); err != nil {
		return fmt.Errorf("db.QueryRow: email %v", err)
	}

	if _, err := upb.NewUserModifierServiceClient(s.UserMgmtConn).UpdateStudent(
		s.AuthHelper.SignedCtx(ctx, stepState.Token),
		&upb.UpdateStudentRequest{
			StudentProfile: &upb.UpdateStudentRequest_StudentProfile{
				Id:               studentID,
				Name:             fmt.Sprintf("name %s", studentID),
				Grade:            5,
				EnrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
				Email:            studentEmail,
				LocationIds:      []string{locationID},
			},
			SchoolId: constants.ManabieSchool,
			SchoolHistories: []*upb.SchoolHistory{
				{
					SchoolId:       schoolID,
					SchoolCourseId: stepState.CourseID,
					StartDate:      timestamppb.Now(),
				},
			},
		},
	); err != nil {
		return fmt.Errorf("updateStudent %w", err)
	}

	// Update current school
	{
		stmt := `UPDATE school_history SET is_current = true WHERE school_id = $1 AND student_id = $2 AND deleted_at IS NULL`
		_, err := db.Exec(ctx, stmt, &schoolID, &studentID)
		if err != nil {
			return fmt.Errorf("err db.Exec: %w", err)
		}
	}

	return nil
}
