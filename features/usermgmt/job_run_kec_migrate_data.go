package usermgmt

import (
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"cloud.google.com/go/storage"
	"github.com/gocarina/gocsv"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) setUpDataToMigrateEnrollmentStatus(ctx context.Context) (context.Context, error) {
	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)
	stepState := StepStateFromContext(ctx)
	stepState.BucketNameJobMigrationStatus = "kec-migrate"
	stepState.ObjectNameJobMigrationStatus = "test.csv"
	_, err := s.generateGradeMaster(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateGradeMaster err: %v", err)
	}
	locationIDs := []string{
		"01GTE5A3DA1N7A6CB1Y9FQWM9W",
		"01GTE5ADBPZ6FR9HRRA7256BD8",
		"01GTE5ANF473SXZ4J4Q94BE638",
	}
	for _, locationID := range locationIDs {
		stmt := `INSERT INTO locationIDs (
				location_id,
				name,
				location_type,
				partner_internal_id,
                parent_location_id,
				created_at,
				updated_at
			)
			VALUES ($1, $2, $3, $4, $5, now(), now()) ON CONFLICT ON CONSTRAINT locations_pkey DO NOTHING;`
		_, err := s.BobPostgresDBTrace.Exec(ctx, stmt, locationID, locationID, "01FR4M51XJY9E77GSN4QZ1Q9M1", "1", constants.ManabieOrgLocation)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	userIDToExternalID := map[string]string{
		"01GSZMW0WD7Q7RXEMS2QEPVDAA": "01GT9CDWXBE7ZVCDKHNVCWY44K",
		"01GSZMWBV3MMWE20H1G66PPKFC": "01GT9CNBR53W1EGH2KTMBT2425",
		"01GSZMWN0ZSZX3C3B1ADQ4CVZM": "01GT9CNSJPJVSJRRTES5W9SHSW",
		"01GSZMX1XB6HREM28G58P0AHN9": "01GT9CNYJN5EA441Q79QT7D9CN",
	}
	client, err := storage.NewClient(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer client.Close()
	rc, err := client.Bucket(stepState.BucketNameJobMigrationStatus).Object(stepState.ObjectNameJobMigrationStatus).NewReader(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rc.Close()
	reader := csv.NewReader(rc)
	var data usermgmt.KecDatas
	if err = gocsv.UnmarshalCSV(reader, &data); err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "gocsv.UnmarshalCSV error")
	}
	s.UserIDToExternalID = userIDToExternalID
	for userID, externalUserID := range userIDToExternalID {
		ctx, err = s.insertUserAndStudentWithExternalID(ctx, userID, externalUserID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.insertUserAndStudentWithID err: %v", err)
		}
	}
	for _, row := range data {
		stmt := `INSERT INTO user_access_paths (user_id, location_id, created_at, updated_at, resource_path) VALUES ($1, $2, now(), now(), $3) ON CONFLICT ON CONSTRAINT user_access_paths_pk DO NOTHING;`
		userID, ok := getKeyByValueMap(stepState.UserIDToExternalID, row.StudentID.String())
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("userID not in stepState.UserIDToExternalID")
		}
		_, err := s.BobDBTrace.Exec(ctx, stmt, userID, row.LocationID.String(), fmt.Sprint(constants.ManabieSchool))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertUserAndStudentWithExternalID(ctx context.Context, id string, externalID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	name := database.Text(fmt.Sprintf("Student %s", id))
	stmt := `INSERT INTO users
		(user_id, user_external_id, name, user_group, country, updated_at, created_at)
		VALUES ($1, $2, $3, $4, $5, now(), now());`
	_, err := s.BobDBTrace.Exec(ctx, stmt, id, externalID, name, cpb.UserGroup_USER_GROUP_STUDENT.String(), "COUNTRY_VN")
	if err != nil {
		return ctx, err
	}

	gradeID := s.ManabieGradeIDs[0]
	stmt = `INSERT INTO students
		(student_id, current_grade, enrollment_status, grade_id, billing_date, updated_at, created_at)
		VALUES ($1, $2, $3, $4, now(), now(), now());`
	_, err = s.BobDBTrace.Exec(ctx, stmt, id, 1, "STUDENT_ENROLLMENT_STATUS_POTENTIAL", gradeID)
	if err != nil {
		return ctx, err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) runJobMigrateEnrollmentStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	zl := logger.NewZapLogger(s.Cfg.Common.Log.ApplicationLevel, true)
	rsc := bootstrap.NewResources().WithLogger(zl).WithDatabaseC(ctx, s.Cfg.PostgresV2.Databases)
	defer rsc.Cleanup() //nolint:errcheck

	err := usermgmt.RunMigrateKecEnrollmentStatus(
		ctx,
		configurations.Config{Common: s.Cfg.Common, PostgresV2: s.Cfg.PostgresV2},
		rsc,
		ManabieSchoolResourcePath,
		stepState.BucketNameJobMigrationStatus,
		stepState.ObjectNameJobMigrationStatus,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("usermgmt.RunMigrateKecEnrollmentStatus: %s", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkDataJobMigrationEnrollmentStatusCorrect(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	client, err := storage.NewClient(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer client.Close()
	rc, err := client.Bucket(stepState.BucketNameJobMigrationStatus).Object(stepState.ObjectNameJobMigrationStatus).NewReader(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	var data usermgmt.KecDatas
	if err = gocsv.UnmarshalCSV(reader, &data); err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "gocsv.UnmarshalCSV error")
	}

	for _, row := range data {
		var existAccessPath bool
		studentID, _ := getKeyByValueMap(stepState.UserIDToExternalID, row.StudentID.String())
		if studentID != "" {
			continue
		}
		checkAccessPathStmt := `SELECT EXISTS (SELECT 1 FROM user_access_paths where user_id = $1 and location_id = $2);`
		err := s.BobDBTrace.QueryRow(ctx, checkAccessPathStmt, studentID, row.LocationID.String()).Scan(&existAccessPath)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if !existAccessPath {
			return StepStateToContext(ctx, stepState), fmt.Errorf("user access path doesn't exist")
		}
		startDate, err := time.Parse(usermgmt.KecFormatDate, row.StartDate.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		var endDate *time.Time
		if field.IsPresent(row.EndDate) {
			endDateVal, err := time.Parse(usermgmt.KecFormatDate, row.EndDate.String())
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			endDate = &endDateVal
		}
		var existEnrollmentStatus bool
		var checkEnrollmentStatusStmt string
		if endDate != nil {
			checkEnrollmentStatusStmt = `SELECT EXISTS
    				(SELECT 1 FROM student_enrollment_status_history WHERE student_id = $1 AND location_id = $2 AND enrollment_status = $3 AND start_date = $4 AND end_date = $5)`

			err = s.BobDBTrace.QueryRow(ctx, checkEnrollmentStatusStmt, studentID, row.LocationID.String(), row.EnrollmentStatus.String(), startDate, endDate).Scan(&existEnrollmentStatus)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			if !existEnrollmentStatus {
				return StepStateToContext(ctx, stepState), fmt.Errorf("student enrollment status history doesn't exist")
			}
		} else {
			checkEnrollmentStatusStmt = `SELECT EXISTS
    				(SELECT 1 FROM student_enrollment_status_history WHERE student_id = $1 AND location_id = $2 AND enrollment_status = $3 AND start_date = $4 AND end_date IS NULL)`

			err = s.BobDBTrace.QueryRow(ctx, checkEnrollmentStatusStmt, studentID, row.LocationID.String(), row.EnrollmentStatus.String(), startDate).Scan(&existEnrollmentStatus)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			if !existEnrollmentStatus {
				return StepStateToContext(ctx, stepState), fmt.Errorf("student enrollment status history doesn't exist")
			}
		}
	}
	ctx, err = s.tearDownMigrationEnrollmentStatusHistory(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) tearDownMigrationEnrollmentStatusHistory(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	userIDs := make([]string, 0)
	for userID := range s.UserIDToExternalID {
		userIDs = append(userIDs, userID)
	}
	if err := database.ExecInTx(ctx, s.BobDBTrace, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM student_enrollment_status_history WHERE student_id=any($1);`, userIDs)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `DELETE FROM user_access_paths WHERE user_id=any($1);`, userIDs)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `DELETE FROM users WHERE user_id=any($1);`, userIDs)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `DELETE FROM students WHERE student_id=any($1);`, userIDs)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
