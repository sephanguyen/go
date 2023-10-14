package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
)

type StudentEnrollmentStatusHistoryRepo struct{}

func (r *StudentEnrollmentStatusHistoryRepo) GetLatestStatusByStudentIDAndLocationID(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
	locationID string,
) (
	entities.StudentEnrollmentStatusHistory,
	error,
) {
	studentEnrollmentStatusHistory := &entities.StudentEnrollmentStatusHistory{}
	fieldNames, fieldValues := studentEnrollmentStatusHistory.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM
			%s
		WHERE 
			student_id = $1
		AND
			location_id = $2
		AND
			deleted_at IS NULL
		ORDER BY created_at DESC
	`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		studentEnrollmentStatusHistory.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentID, locationID)
	err := row.Scan(fieldValues...)
	if err != nil {
		return entities.StudentEnrollmentStatusHistory{}, err
	}
	return *studentEnrollmentStatusHistory, nil
}

func (r *StudentEnrollmentStatusHistoryRepo) GetCurrentStatusByStudentIDAndLocationID(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
	locationID string,
) (
	entities.StudentEnrollmentStatusHistory,
	error,
) {
	studentEnrollmentStatusHistory := &entities.StudentEnrollmentStatusHistory{}
	fieldNames, fieldValues := studentEnrollmentStatusHistory.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM
			%s
		WHERE 
			student_id = $1
		AND
			location_id = $2
		AND
			deleted_at IS NULL
		AND 
			(start_date < NOW() AND 
				(end_date > NOW() OR end_date IS NULL))
		ORDER BY created_at DESC
	`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		studentEnrollmentStatusHistory.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentID, locationID)
	err := row.Scan(fieldValues...)
	if err != nil {
		return entities.StudentEnrollmentStatusHistory{}, err
	}
	return *studentEnrollmentStatusHistory, nil
}

func (r *StudentEnrollmentStatusHistoryRepo) GetListStudentEnrollmentStatusHistoryByStudentID(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
) (
	[]*entities.StudentEnrollmentStatusHistory,
	error,
) {
	var studentEnrollmentStatusHistoryList []*entities.StudentEnrollmentStatusHistory
	studentEnrollmentStatusHistory := &entities.StudentEnrollmentStatusHistory{}
	fieldNames, _ := studentEnrollmentStatusHistory.FieldMap()
	stmt :=
		`
		SELECT distinct on (location_id) %s
		FROM
			%s
		WHERE 
		    student_id = $1
		AND
			deleted_at IS NULL
	    ORDER BY location_id, created_at DESC
	`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		studentEnrollmentStatusHistory.TableName(),
	)
	rows, err := db.Query(ctx, stmt, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		studentEnrollmentStatusHistory := new(entities.StudentEnrollmentStatusHistory)
		_, fieldValues := studentEnrollmentStatusHistory.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		studentEnrollmentStatusHistoryList = append(studentEnrollmentStatusHistoryList, studentEnrollmentStatusHistory)
	}
	return studentEnrollmentStatusHistoryList, nil
}

func (r *StudentEnrollmentStatusHistoryRepo) GetListEnrolledStatusByStudentIDAndTime(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
	time time.Time,
) (
	[]*entities.StudentEnrollmentStatusHistory,
	error,
) {
	var studentEnrollmentStatusHistoryList []*entities.StudentEnrollmentStatusHistory
	studentEnrollmentStatusHistory := &entities.StudentEnrollmentStatusHistory{}
	fieldNames, _ := studentEnrollmentStatusHistory.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM
			%s
		WHERE 
			student_id = $1
		AND
			deleted_at IS NULL
		AND 
			(start_date <= $2 AND 
				(end_date >= $2 OR end_date IS NULL))
		AND ( enrollment_status = 'STUDENT_ENROLLMENT_STATUS_ENROLLED' OR enrollment_status = 'STUDENT_ENROLLMENT_STATUS_LOA')
		ORDER BY created_at DESC
	`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		studentEnrollmentStatusHistory.TableName(),
	)
	rows, err := db.Query(ctx, stmt, studentID, time)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		studentEnrollmentStatusHistory := new(entities.StudentEnrollmentStatusHistory)
		_, fieldValues := studentEnrollmentStatusHistory.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		studentEnrollmentStatusHistoryList = append(studentEnrollmentStatusHistoryList, studentEnrollmentStatusHistory)
	}
	return studentEnrollmentStatusHistoryList, nil
}

func (r *StudentEnrollmentStatusHistoryRepo) GetListEnrolledStudentEnrollmentStatusByStudentID(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
) (
	[]*entities.StudentEnrollmentStatusHistory,
	error,
) {
	var studentEnrollmentStatusHistoryList []*entities.StudentEnrollmentStatusHistory
	studentEnrollmentStatusHistory := &entities.StudentEnrollmentStatusHistory{}
	fieldNames, _ := studentEnrollmentStatusHistory.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM
			%s
		WHERE 
			student_id = $1
		AND
			deleted_at IS NULL
		AND 
			(start_date <= NOW() AND 
				(end_date >= NOW() OR end_date IS NULL))
		AND ( enrollment_status = 'STUDENT_ENROLLMENT_STATUS_ENROLLED' OR enrollment_status = 'STUDENT_ENROLLMENT_STATUS_LOA')
		ORDER BY created_at DESC
	`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		studentEnrollmentStatusHistory.TableName(),
	)
	rows, err := db.Query(ctx, stmt, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		studentEnrollmentStatusHistory := new(entities.StudentEnrollmentStatusHistory)
		_, fieldValues := studentEnrollmentStatusHistory.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		studentEnrollmentStatusHistoryList = append(studentEnrollmentStatusHistoryList, studentEnrollmentStatusHistory)
	}
	return studentEnrollmentStatusHistoryList, nil
}

func (r *StudentEnrollmentStatusHistoryRepo) GetLatestStatusEnrollmentByStudentIDAndLocationIDs(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
	locationIDs []string,
) (
	[]*entities.StudentEnrollmentStatusHistory,
	error,
) {
	var studentEnrollmentStatusHistoryList []*entities.StudentEnrollmentStatusHistory
	studentEnrollmentStatusHistory := &entities.StudentEnrollmentStatusHistory{}
	fieldNames, _ := studentEnrollmentStatusHistory.FieldMap()
	stmt :=
		`
		SELECT distinct on (location_id) %s
		FROM
			%s
		WHERE 
		    student_id = $1
		AND
		    location_id = ANY($2)
		AND
			deleted_at IS NULL
	    ORDER BY location_id, created_at DESC
	`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		studentEnrollmentStatusHistory.TableName(),
	)
	rows, err := db.Query(ctx, stmt, studentID, locationIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		studentEnrollmentStatusHistory := new(entities.StudentEnrollmentStatusHistory)
		_, fieldValues := studentEnrollmentStatusHistory.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		studentEnrollmentStatusHistoryList = append(studentEnrollmentStatusHistoryList, studentEnrollmentStatusHistory)
	}
	return studentEnrollmentStatusHistoryList, nil
}
