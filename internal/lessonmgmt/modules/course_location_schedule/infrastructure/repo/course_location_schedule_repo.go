package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/domain"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type CourseLocationScheduleRepo struct{}

func (c *CourseLocationScheduleRepo) UpsertMultiCourseLocationSchedule(ctx context.Context, db database.QueryExecer, arrCourseLocationSchedule []*domain.CourseLocationSchedule) *domain.ImportCourseLocationScheduleError {
	ctx, span := interceptors.StartSpan(ctx, "CourseLocationScheduleRepo.InsertCoursesLocationSchedule")
	defer span.End()
	b := &pgx.Batch{}
	strQuery := `INSERT INTO course_location_schedule (%s) 
	VALUES (%s) ON CONFLICT ON CONSTRAINT course_location_schedule_pk 
	DO UPDATE SET course_id = $2, location_id = $3, academic_weeks = $4, product_type_schedule = $5, frequency = $6, total_no_lessons = $7, updated_at = $9 `
	for _, courseLocationSchedule := range arrCourseLocationSchedule {
		l, err := NewCourseLocationScheduleDTOFromCourseLocationScheduleDomain(courseLocationSchedule)
		if err != nil {
			return &domain.ImportCourseLocationScheduleError{
				Index: -1,
				Err:   err,
			}
		}
		fieldsToCreate, valuesToCreate := l.FieldMap()

		query := fmt.Sprintf(
			strQuery,
			strings.Join(fieldsToCreate, ","),
			database.GeneratePlaceholders(len(fieldsToCreate)),
		)
		b.Queue(query, valuesToCreate...)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			e := err
			if strings.Contains(err.Error(), "unique_course_location_schedule") {
				e = domain.ErrUniqCourseLocationSchedule
			}
			if strings.Contains(err.Error(), "course_location_schedule_fk") {
				e = domain.ErrNotExistsFKCourseLocationSchedule
			}
			return &domain.ImportCourseLocationScheduleError{
				Index: i,
				Err:   e,
			}
		}
	}
	return nil
}

func (c *CourseLocationScheduleRepo) ExportCourseLocationSchedule(ctx context.Context, db database.QueryExecer) ([]*domain.CourseLocationSchedule, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseLocationScheduleRepo.ExportCourseLocationSchedule")
	courseLocationSchedule := &CourseLocationSchedule{}
	fields, _ := courseLocationSchedule.FieldMap()
	defer span.End()
	strQuery := `SELECT cls.course_location_schedule_id, cap.course_id, cap.location_id, 
				 cls.academic_weeks, cls.product_type_schedule, cls.frequency, cls.total_no_lessons,
				 cls.created_at, cls.updated_at 
				 FROM course_location_schedule cls 
				 RIGHT JOIN course_access_paths cap ON cls.course_id = cap.course_id and cls.location_id = cap.location_id 
				 WHERE cls.deleted_at IS NULL AND cap.deleted_at IS NULL`
	rows, err := db.Query(ctx, strQuery)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	arrCourseLocationSchedule := []*domain.CourseLocationSchedule{}
	for rows.Next() {
		courseLocationSchedule := &CourseLocationSchedule{}
		if err := rows.Scan(database.GetScanFields(courseLocationSchedule, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		courseLocationScheduleDomain, err := NewCourseLocationScheduleDomainFromCourseLocationScheduleDTO(courseLocationSchedule)
		if err != nil {
			return nil, errors.Wrap(err, "rows convert error")
		}
		arrCourseLocationSchedule = append(arrCourseLocationSchedule, courseLocationScheduleDomain)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return arrCourseLocationSchedule, nil
}

func (c *CourseLocationScheduleRepo) GetAcademicWeekValid(ctx context.Context, db database.QueryExecer, locationIds []string, dateValid time.Time) (map[string]bool, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseLocationScheduleRepo.GetAcademicWeekValid")
	mapResult := make(map[string]bool)
	defer span.End()
	strQuery := `SELECT aw.location_id, aw.week_order 
				 FROM academic_week aw 
				 WHERE aw.location_id = ANY($1) AND aw.start_date::date > $2::date`
	rows, err := db.Query(ctx, strQuery, locationIds, dateValid)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	for rows.Next() {
		var locationID pgtype.Text
		var weekOrder pgtype.Int2
		if err := rows.Scan(&locationID, &weekOrder); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		mapResult[fmt.Sprintf("%s-%d", locationID.String, weekOrder.Int)] = true
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return mapResult, nil
}

func (c *CourseLocationScheduleRepo) GetByCourseIDAndLocationID(ctx context.Context, db database.QueryExecer, courseID, locationID string) (*domain.CourseLocationSchedule, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseLocationScheduleRepo.GetByCourseIDAndLocationID")
	defer span.End()

	e := &CourseLocationSchedule{}
	fields, args := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE course_id = $1 AND location_id = $2 AND deleted_at IS NULL", strings.Join(fields, ", "), e.TableName())
	err := db.QueryRow(ctx, query, &courseID, &locationID).Scan(args...)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrorNotFound
	} else if err != nil {
		return nil, errors.Wrap(err, "db.QueryRow")
	}
	courseLocationSchedule, err := NewCourseLocationScheduleDomainFromCourseLocationScheduleDTO(e)
	return courseLocationSchedule, err
}
