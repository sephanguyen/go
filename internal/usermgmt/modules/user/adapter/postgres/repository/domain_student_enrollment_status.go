package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	pbu "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
)

type DomainEnrollmentStatusHistoryRepo struct{}

type EnrollmentStatusHistoryAttribute struct {
	UserID              field.String
	LocationID          field.String
	EnrollmentStatus    field.String
	StartDate           field.Time
	EndDate             field.Time
	OrganizationID      field.String
	OrderID             field.String
	OrderSequenceNumber field.Int32
}

type EnrollmentStatusHistory struct {
	EnrollmentStatusHistoryAttribute
	CreatedAtAttr field.Time
	UpdatedAtAttr field.Time
	DeletedAtAttr field.Time
}

func NewEnrollmentStatusHistory(enrollmentStatus entity.DomainEnrollmentStatusHistory) *EnrollmentStatusHistory {
	now := field.NewTime(time.Now().UTC())
	return &EnrollmentStatusHistory{
		EnrollmentStatusHistoryAttribute: EnrollmentStatusHistoryAttribute{
			UserID:              enrollmentStatus.UserID(),
			LocationID:          enrollmentStatus.LocationID(),
			EnrollmentStatus:    enrollmentStatus.EnrollmentStatus(),
			StartDate:           enrollmentStatus.StartDate(),
			EndDate:             enrollmentStatus.EndDate(),
			OrganizationID:      enrollmentStatus.OrganizationID(),
			OrderID:             enrollmentStatus.OrderID(),
			OrderSequenceNumber: enrollmentStatus.OrderSequenceNumber(),
		},
		CreatedAtAttr: now,
		UpdatedAtAttr: now,
		DeletedAtAttr: field.NewNullTime(),
	}
}

func (e *EnrollmentStatusHistory) UserID() field.String {
	return e.EnrollmentStatusHistoryAttribute.UserID
}

func (e *EnrollmentStatusHistory) LocationID() field.String {
	return e.EnrollmentStatusHistoryAttribute.LocationID
}

func (e *EnrollmentStatusHistory) EnrollmentStatus() field.String {
	return e.EnrollmentStatusHistoryAttribute.EnrollmentStatus
}

func (e *EnrollmentStatusHistory) StartDate() field.Time {
	return e.EnrollmentStatusHistoryAttribute.StartDate
}

func (e *EnrollmentStatusHistory) EndDate() field.Time {
	return e.EnrollmentStatusHistoryAttribute.EndDate
}

func (e *EnrollmentStatusHistory) OrderID() field.String {
	return e.EnrollmentStatusHistoryAttribute.OrderID
}

func (e *EnrollmentStatusHistory) OrderSequenceNumber() field.Int32 {
	return e.EnrollmentStatusHistoryAttribute.OrderSequenceNumber
}

func (e *EnrollmentStatusHistory) OrganizationID() field.String {
	return e.EnrollmentStatusHistoryAttribute.OrganizationID
}

func (e *EnrollmentStatusHistory) CreatedAt() field.Time {
	return e.CreatedAtAttr
}

func (e *EnrollmentStatusHistory) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id",
			"location_id",
			"enrollment_status",
			"start_date",
			"end_date",
			"order_id",
			"order_sequence_number",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.EnrollmentStatusHistoryAttribute.UserID,
			&e.EnrollmentStatusHistoryAttribute.LocationID,
			&e.EnrollmentStatusHistoryAttribute.EnrollmentStatus,
			&e.EnrollmentStatusHistoryAttribute.StartDate,
			&e.EnrollmentStatusHistoryAttribute.EndDate,
			&e.EnrollmentStatusHistoryAttribute.OrderID,
			&e.EnrollmentStatusHistoryAttribute.OrderSequenceNumber,
			&e.CreatedAtAttr,
			&e.UpdatedAtAttr,
			&e.DeletedAtAttr,
			&e.EnrollmentStatusHistoryAttribute.OrganizationID,
		}
}

func (e *EnrollmentStatusHistory) TableName() string {
	return "student_enrollment_status_history"
}

func (e *DomainEnrollmentStatusHistoryRepo) Create(ctx context.Context, db database.QueryExecer, enrollmentStatusHistoryToCreate entity.DomainEnrollmentStatusHistory) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainEnrollmentStatusHistoryRepo.Create")
	defer span.End()

	newEnrollmentStatus := NewEnrollmentStatusHistory(enrollmentStatusHistoryToCreate)

	if enrollmentStatusHistoryToCreate.StartDate().Time().IsZero() {
		newEnrollmentStatus.EnrollmentStatusHistoryAttribute.StartDate = newEnrollmentStatus.CreatedAtAttr
	}
	cmdTag, err := database.Insert(ctx, newEnrollmentStatus, db.Exec)
	if err != nil {
		return InternalError{
			RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.Create"),
		}
	}
	if cmdTag.RowsAffected() != 1 {
		return InternalError{
			RawError: errors.Wrap(ErrNoRowAffected, "repo.DomainEnrollmentStatusHistoryRepo.Create"),
		}
	}

	return nil
}

func (e *DomainEnrollmentStatusHistoryRepo) Update(ctx context.Context, db database.QueryExecer, enrollStatusHistDB, enrollStatusHistReq entity.DomainEnrollmentStatusHistory) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainEnrollmentStatusHistoryRepo.Update")
	defer span.End()
	enrollmentStatusHistory := &EnrollmentStatusHistory{}
	startDate := enrollStatusHistReq.StartDate().Time()
	if startDate.IsZero() {
		startDate = time.Now()
	}

	query := fmt.Sprintf(`
	UPDATE %s
	SET start_date = COALESCE($1, start_date),
		end_date = COALESCE($2, end_date),
		enrollment_status = COALESCE($3, enrollment_status)
	WHERE student_id = $4
		AND location_id = $5
		AND start_date = $6
		AND enrollment_status = $7
		AND deleted_at IS NULL;
	`,
		enrollmentStatusHistory.TableName())
	_, err := db.Exec(ctx,
		query,
		database.Timestamptz(startDate),
		database.TimestamptzNull(enrollStatusHistReq.EndDate().Time()),
		database.Text(enrollStatusHistReq.EnrollmentStatus().String()),
		database.Text(enrollStatusHistDB.UserID().String()),
		database.Text(enrollStatusHistDB.LocationID().String()),
		database.Timestamptz(enrollStatusHistDB.StartDate().Time()),
		database.Text(enrollStatusHistDB.EnrollmentStatus().String()),
	)

	if err != nil {
		return InternalError{
			RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.Update"),
		}
	}

	return nil
}

func (e *DomainEnrollmentStatusHistoryRepo) DeactivateEnrollmentStatus(ctx context.Context, db database.QueryExecer, enrollmentStatusHistoryToCreate entity.DomainEnrollmentStatusHistory, endDateReq time.Time) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainEnrollmentStatusHistoryRepo.DeactivateEnrollmentStatus")
	defer span.End()

	enrollmentStatusHistory := &EnrollmentStatusHistory{}

	query := fmt.Sprintf(`UPDATE %s SET end_date = $1 
          							WHERE student_id = $2 
                                     AND location_id = $3
          							 AND enrollment_status = $4
                                     AND start_date = $5 
                                     AND deleted_at IS NULL`,
		enrollmentStatusHistory.TableName())

	_, err := db.Exec(ctx,
		query,
		database.TimestamptzNull(endDateReq),
		database.Text(enrollmentStatusHistoryToCreate.UserID().String()),
		database.Text(enrollmentStatusHistoryToCreate.LocationID().String()),
		database.Text(enrollmentStatusHistoryToCreate.EnrollmentStatus().String()),
		database.Timestamptz(enrollmentStatusHistoryToCreate.StartDate().Time()),
	)
	if err != nil {
		return InternalError{
			RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.DeactivateEnrollmentStatus"),
		}
	}

	return nil
}

func (e *DomainEnrollmentStatusHistoryRepo) SoftDeleteEnrollments(ctx context.Context, db database.QueryExecer, enrollmentStatusHistoryToCreate entity.DomainEnrollmentStatusHistory) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainEnrollmentStatusHistoryRepo.SoftDeleteByUserIDs")
	defer span.End()

	enrollmentStatusHistory := &EnrollmentStatusHistory{}

	query := fmt.Sprintf(`UPDATE %s SET deleted_at = NOW()
             						WHERE student_id = $1 
                                     AND location_id = $2 
                                     AND start_date = $3 
             						 AND enrollment_status = $4
                                     AND deleted_at IS NULL`,
		enrollmentStatusHistory.TableName())
	_, err := db.Exec(ctx,
		query,
		database.Text(enrollmentStatusHistoryToCreate.UserID().String()),
		database.Text(enrollmentStatusHistoryToCreate.LocationID().String()),
		database.Timestamptz(enrollmentStatusHistoryToCreate.StartDate().Time()),
		database.Text(enrollmentStatusHistoryToCreate.EnrollmentStatus().String()),
	)
	if err != nil {
		return InternalError{
			RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.SoftDeleteEnrollments"),
		}
	}

	return nil
}

func (e *DomainEnrollmentStatusHistoryRepo) GetByStudentIDAndLocationID(ctx context.Context, db database.QueryExecer, studentID, locationID string, getCurrent bool) (entity.DomainEnrollmentStatusHistories, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainEnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID")
	defer span.End()

	enrollmentStatusHistory := &EnrollmentStatusHistory{}
	fields, _ := enrollmentStatusHistory.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s  
 									WHERE student_id = $1 
 			    					AND location_id = $2 
 			    					AND deleted_at IS NULL`,
		strings.Join(fields, ","), enrollmentStatusHistory.TableName())

	if getCurrent {
		query += ` AND start_date < NOW() AND ( end_date > NOW() OR end_date IS NULL)`
	}

	rows, err := db.Query(
		ctx,
		query,
		database.Text(studentID),
		database.Text(locationID),
	)
	if err != nil {
		return nil, InternalError{
			RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID"),
		}
	}

	defer rows.Close()

	var result []entity.DomainEnrollmentStatusHistory
	for rows.Next() {
		item := &EnrollmentStatusHistory{}

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{
				RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID"),
			}
		}

		result = append(result, item)
	}

	return result, nil
}

func (e *DomainEnrollmentStatusHistoryRepo) GetLatestEnrollmentStudentOfLocation(ctx context.Context, db database.QueryExecer, studentID, locationID string) ([]entity.DomainEnrollmentStatusHistory, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainEnrollmentStatusHistoryRepo.GetLatestEnrollmentStudentOfLocation")
	defer span.End()

	enrollmentStatusHistory := &EnrollmentStatusHistory{}
	fields, _ := enrollmentStatusHistory.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s  
 									WHERE student_id = $1 
 			    					AND location_id = $2 
 			    					AND deleted_at IS NULL
 			    					ORDER BY created_at DESC
               						FETCH FIRST 2 ROWS ONLY`,
		strings.Join(fields, ","), enrollmentStatusHistory.TableName())

	rows, err := db.Query(
		ctx,
		query,
		database.Text(studentID),
		database.Text(locationID),
	)
	if err != nil {
		return nil, InternalError{
			RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.GetLatestEnrollmentStudentOfLocation"),
		}
	}

	defer rows.Close()

	var result []entity.DomainEnrollmentStatusHistory
	for rows.Next() {
		item := &EnrollmentStatusHistory{}

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{
				RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.GetLatestEnrollmentStudentOfLocation: rows.Scan"),
			}
		}

		result = append(result, item)
	}

	return result, nil
}

func (e *DomainEnrollmentStatusHistoryRepo) GetOutDateEnrollmentStatus(ctx context.Context, db database.QueryExecer, organizationID string) ([]entity.DomainEnrollmentStatusHistory, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainEnrollmentStatusHistoryRepo.GetOutDateEnrollmentStatus")
	defer span.End()

	enrollmentStatusHistory := &EnrollmentStatusHistory{}
	fields, _ := enrollmentStatusHistory.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s  
 									WHERE end_date between now() - interval '1' day AND now()
 			    					AND deleted_at IS NULL 
 									AND resource_path = $1
 									AND enrollment_status = $2`,
		strings.Join(fields, ","), enrollmentStatusHistory.TableName())

	rows, err := db.Query(
		ctx,
		query,
		database.Text(organizationID),
		database.Text(pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String()),
	)
	if err != nil {
		return nil, InternalError{
			RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.GetOutDateEnrollmentStatus"),
		}
	}

	defer rows.Close()

	var result []entity.DomainEnrollmentStatusHistory
	for rows.Next() {
		item := &EnrollmentStatusHistory{}

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{
				RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.GetOutDateEnrollmentStatus"),
			}
		}

		result = append(result, item)
	}

	return result, nil
}

func (e *DomainEnrollmentStatusHistoryRepo) GetByStudentID(ctx context.Context, db database.QueryExecer, studentID string, getCurrent bool) ([]entity.DomainEnrollmentStatusHistory, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainEnrollmentStatusHistoryRepo.GetByStudentID")
	defer span.End()
	enrollmentStatusHistory := &EnrollmentStatusHistory{}
	fields, _ := enrollmentStatusHistory.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE student_id = $1 AND deleted_at IS NULL`,
		strings.Join(fields, ","), enrollmentStatusHistory.TableName())
	if getCurrent {
		query += ` AND start_date < NOW() AND ( end_date > NOW() OR end_date IS NULL)`
	}

	rows, err := db.Query(
		ctx,
		query,
		database.Text(studentID),
	)
	if err != nil {
		return nil, InternalError{
			RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.GetByStudentID"),
		}
	}
	defer rows.Close()

	var result []entity.DomainEnrollmentStatusHistory
	for rows.Next() {
		item := &EnrollmentStatusHistory{}
		_, fieldValues := item.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{
				RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.GetByStudentID"),
			}
		}
		result = append(result, item)
	}
	return result, nil
}

func (e *DomainEnrollmentStatusHistoryRepo) GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate(ctx context.Context, db database.QueryExecer, enrollmentStatusHistoryReq entity.DomainEnrollmentStatusHistory) ([]entity.DomainEnrollmentStatusHistory, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainEnrollmentStatusHistoryRepo.GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate")
	defer span.End()
	enrollmentStatusHistory := &EnrollmentStatusHistory{}
	fields, _ := enrollmentStatusHistory.FieldMap()
	startDate := enrollmentStatusHistoryReq.StartDate().Time()
	endDate := enrollmentStatusHistoryReq.EndDate().Time()

	query := fmt.Sprintf(`
	SELECT %s FROM %s 
			WHERE student_id = $1 
			AND location_id = $2
			AND enrollment_status = $3
			AND date_trunc('day'::text, start_date::timestamp) = date_trunc('day'::text, $4::timestamp)
			AND date_trunc('day'::text, end_date::timestamp) IS NOT DISTINCT FROM date_trunc('day'::text, $5::timestamp)
			AND deleted_at IS NULL
	`,
		strings.Join(fields, ","), enrollmentStatusHistory.TableName())

	if startDate.IsZero() {
		startDate = time.Now()
	}

	params := []interface{}{
		database.Text(enrollmentStatusHistoryReq.UserID().String()),
		database.Text(enrollmentStatusHistoryReq.LocationID().String()),
		database.Text(enrollmentStatusHistoryReq.EnrollmentStatus().String()),
		database.Timestamptz(startDate),
		database.TimestamptzNull(endDate),
	}

	rows, err := db.Query(
		ctx,
		query,
		params...,
	)
	if err != nil {
		return nil, InternalError{
			RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate"),
		}
	}
	defer rows.Close()

	var result []entity.DomainEnrollmentStatusHistory
	for rows.Next() {
		item := &EnrollmentStatusHistory{}
		_, fieldValues := item.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{
				RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate"),
			}
		}
		result = append(result, item)
	}
	return result, nil
}

func (e *DomainEnrollmentStatusHistoryRepo) GetInactiveAndActiveStudents(ctx context.Context, db database.QueryExecer, studentIDs, deactivateEnrollmentStatuses []string) ([]entity.DomainEnrollmentStatusHistory, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainEnrollmentStatusHistoryRepo.GetInactiveAndActiveStudents")
	defer span.End()
	enrollmentStatusHistory := &EnrollmentStatusHistory{}

	query := fmt.Sprintf(`
		WITH inactive_students as (
			SELECT sesh1.student_id, MAX(sesh1.start_date) AS start_date 
			FROM %[1]s sesh1
			WHERE NOT EXISTS 
				(
					SELECT 1
					FROM %[1]s sesh2
					WHERE NOT (sesh2.enrollment_status = ANY($1::text[]))
					AND sesh2.student_id = sesh1.student_id 
					AND (end_date > CLOCK_TIMESTAMP() OR end_date IS NULL)
					AND start_date < CLOCK_TIMESTAMP()
					AND deleted_at IS NULL
				)
			AND (end_date > CLOCK_TIMESTAMP() OR end_date IS NULL)
			AND start_date < CLOCK_TIMESTAMP()
			AND deleted_at IS NULL 
			AND ((ARRAY_LENGTH('{%[2]s}'::text[], 1) IS NULL) or (sesh1.student_id = ANY('{%[2]s}')))
			GROUP BY sesh1.student_id
		)
			SELECT %[1]s.student_id, COALESCE(inactive_students.start_date,NULL) as start_date
			FROM %[1]s left join inactive_students on %[1]s.student_id = inactive_students.student_id
			WHERE %[1]s.deleted_at IS NULL
			AND ((ARRAY_LENGTH('{%[2]s}'::text[], 1) IS NULL) or (%[1]s.student_id = ANY('{%[2]s}')))
			GROUP BY %[1]s.student_id,inactive_students.start_date`,
		enrollmentStatusHistory.TableName(), strings.Join(studentIDs, ","),
	)

	rows, err := db.Query(
		ctx,
		query,
		database.TextArray(
			deactivateEnrollmentStatuses,
		),
	)
	if err != nil {
		return nil, InternalError{
			RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.GetInactiveAndActiveStudents"),
		}
	}
	defer rows.Close()
	var enrollmentStatusHistories []entity.DomainEnrollmentStatusHistory
	for rows.Next() {
		enrollmentStatusHistory := &EnrollmentStatusHistory{}
		_, fieldValues := enrollmentStatusHistory.FieldMap()
		err := rows.Scan(fieldValues[0], fieldValues[3])
		if err != nil {
			return nil, InternalError{
				RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.GetInactiveAndActiveStudents"),
			}
		}
		enrollmentStatusHistories = append(enrollmentStatusHistories, enrollmentStatusHistory)
	}
	return enrollmentStatusHistories, nil
}

func (e *DomainEnrollmentStatusHistoryRepo) UpdateStudentStatusBasedEnrollmentStatus(ctx context.Context, db database.QueryExecer, studentIDs, deactivateEnrollmentStatuses []string) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainEnrollmentStatusHistoryRepo.UpdateStudentStatusBasedEnrollmentStatus")
	defer span.End()
	enrollmentStatusHistory := &EnrollmentStatusHistory{}

	getStudentStatusQuery := fmt.Sprintf(`
		WITH inactive_students as (
			SELECT sesh1.student_id, MAX(sesh1.start_date) AS latest_start_date 
			FROM %[1]s sesh1
			WHERE NOT EXISTS 
				(
					SELECT 1
					FROM %[1]s sesh2
					WHERE NOT (sesh2.enrollment_status = ANY($1::text[]))
					AND sesh2.student_id = sesh1.student_id 
					AND (end_date > CLOCK_TIMESTAMP() OR end_date IS NULL)
					AND start_date < CLOCK_TIMESTAMP()
					AND deleted_at IS NULL
				)
			AND (end_date > CLOCK_TIMESTAMP() OR end_date IS NULL)
			AND start_date < CLOCK_TIMESTAMP()
			AND deleted_at IS NULL 
			AND ((ARRAY_LENGTH('{%[2]s}'::text[], 1) IS NULL) or (sesh1.student_id = ANY('{%[2]s}')))
			GROUP BY sesh1.student_id
		),
		upsert_students AS (
			SELECT %[1]s.student_id, COALESCE(inactive_students.latest_start_date,NULL) as deactivation_date
			FROM %[1]s left join inactive_students on %[1]s.student_id = inactive_students.student_id
			WHERE %[1]s.deleted_at IS NULL
			AND ((ARRAY_LENGTH('{%[2]s}'::text[], 1) IS NULL) or (%[1]s.student_id = ANY('{%[2]s}')))
			GROUP BY %[1]s.student_id,inactive_students.latest_start_date
		) `,
		enrollmentStatusHistory.TableName(), strings.Join(studentIDs, ","),
	)

	updateStatusQuery := `UPDATE users SET deactivated_at = upsert_students.deactivation_date FROM upsert_students
		WHERE users.user_id = upsert_students.student_id 
		AND users.deactivated_at IS DISTINCT FROM upsert_students.deactivation_date`

	query := getStudentStatusQuery + updateStatusQuery
	_, err := db.Exec(
		ctx,
		query,
		database.TextArray(
			deactivateEnrollmentStatuses,
		),
	)
	if err != nil {
		return InternalError{
			RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.UpdateStudentStatusBasedEnrollmentStatus"),
		}
	}
	return nil
}

func (e *DomainEnrollmentStatusHistoryRepo) GetByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) (entity.DomainEnrollmentStatusHistories, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainEnrollmentStatusHistoryRepo.GetByStudentIDs")
	defer span.End()
	enrollmentStatusHistory := &EnrollmentStatusHistory{}
	fields, _ := enrollmentStatusHistory.FieldMap()

	query := fmt.Sprintf(`
	SELECT %s FROM %s 
		WHERE student_id = ANY($1) 
		AND deleted_at IS NULL`,
		strings.Join(fields, ","), enrollmentStatusHistory.TableName())

	rows, err := db.Query(
		ctx,
		query,
		database.TextArray(studentIDs),
	)
	if err != nil {
		return nil, InternalError{
			RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.GetByStudentIDs"),
		}
	}
	defer rows.Close()

	var result []entity.DomainEnrollmentStatusHistory
	for rows.Next() {
		item := &EnrollmentStatusHistory{}
		_, fieldValues := item.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{
				RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.GetByStudentIDs"),
			}
		}
		result = append(result, item)
	}
	return result, nil
}

func (e *DomainEnrollmentStatusHistoryRepo) BulkInsert(ctx context.Context, db database.QueryExecer, reqEnrollmentStatusHistoriesToCreate entity.DomainEnrollmentStatusHistories) error {
	if len(reqEnrollmentStatusHistoriesToCreate) == 0 {
		return nil
	}
	ctx, span := interceptors.StartSpan(ctx, "DomainEnrollmentStatusHistoryRepo.BulkInsert")
	defer span.End()

	repoEnrollmentStatusHistory := NewEnrollmentStatusHistory(entity.DefaultDomainEnrollmentStatusHistory{})
	fields, _ := repoEnrollmentStatusHistory.FieldMap()

	currentLen := 1
	s := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", repoEnrollmentStatusHistory.TableName(), strings.Join(fields, ","))
	var insertValues []interface{}

	queueFn := func(enrollmentStatusHistory *EnrollmentStatusHistory, idx int) {
		fields, values := enrollmentStatusHistory.FieldMap()
		insertValues = append(insertValues, values...)

		insertPlaceHolders := database.GeneratePlaceholdersWithFirstIndex(currentLen, len(fields))
		placeHolders := fmt.Sprintf("(%s)", insertPlaceHolders)
		s += placeHolders
		if idx != len(reqEnrollmentStatusHistoriesToCreate)-1 {
			s += ","
		}
	}

	for idx, reqEnrollmentStatusHistory := range reqEnrollmentStatusHistoriesToCreate {
		repoDomainUser := NewEnrollmentStatusHistory(reqEnrollmentStatusHistory)
		queueFn(repoDomainUser, idx)
		currentLen += len(fields)
	}
	cmdTag, err := db.Exec(ctx, s, insertValues...)
	if err != nil {
		return InternalError{
			RawError: errors.Wrap(err, "repo.DomainEnrollmentStatusHistoryRepo.BulkInsert"),
		}
	}

	if cmdTag.RowsAffected() < 1 {
		return InternalError{
			RawError: errors.Wrapf(ErrNoRowAffected, "repo.DomainEnrollmentStatusHistoryRepo.BulkInsert, rows affected: %d", cmdTag.RowsAffected()),
		}
	}
	return nil
}

func (e *DomainEnrollmentStatusHistoryRepo) GetSameStartDateEnrollmentStatusHistory(ctx context.Context, db database.QueryExecer, enrollmentStatusHistory entity.DomainEnrollmentStatusHistory) (entity.DomainEnrollmentStatusHistories, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainEnrollmentStatusHistoryRepo.GetSameStartDateEnrollmentStatusHistory")
	defer span.End()

	history := new(EnrollmentStatusHistory)
	fields, _ := new(EnrollmentStatusHistory).FieldMap()

	query := fmt.Sprintf(
		`
			SELECT %s FROM %s 
			WHERE
				DATE_TRUNC('second', start_date::TIMESTAMP) = DATE_TRUNC('second', $1::TIMESTAMP) AND
				student_id = $2 AND
				location_id = $3 AND
				enrollment_status = $4
		`,
		strings.Join(fields, ","),
		history.TableName(),
	)
	rows, err := db.Query(
		ctx, query,
		database.Timestamptz(enrollmentStatusHistory.StartDate().Time()),
		database.Text(enrollmentStatusHistory.UserID().String()),
		database.Text(enrollmentStatusHistory.LocationID().String()),
		database.Text(enrollmentStatusHistory.EnrollmentStatus().String()),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]entity.DomainEnrollmentStatusHistory, 0)
	for rows.Next() {
		item := new(EnrollmentStatusHistory)
		_, values := item.FieldMap()
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
