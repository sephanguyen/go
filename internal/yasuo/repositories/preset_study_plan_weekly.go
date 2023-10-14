package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type PresetStudyPlanWeeklyRepo struct{}

func (r *PresetStudyPlanWeeklyRepo) Create(ctx context.Context, db database.Ext, plans []*entities.PresetStudyPlanWeekly) error {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanWeeklyRepo.Create")
	defer span.End()

	queueFn := func(b *pgx.Batch, e *entities.PresetStudyPlanWeekly) {
		fieldNames := database.GetFieldNames(e)
		placeHolders := "$1, $2, $3, $4, $5, $6, $7, $8, $9, $10"

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			e.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)
		b.Queue(query, database.GetScanFields(e, fieldNames)...)
	}

	b := &pgx.Batch{}
	var d pgtype.Timestamptz
	err := d.Set(time.Now())
	if err != nil {
		return fmt.Errorf("cannot set time preset study plans: %w", err)
	}

	for _, each := range plans {
		if each.ID.String == "" {
			err = each.ID.Set(idutil.ULIDNow())
			if err != nil {
				return fmt.Errorf("cannot set id for preset study plans: %w", err)
			}
		}
		each.CreatedAt = d
		each.UpdatedAt = d
		queueFn(b, each)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}

	return nil
}

func (r *PresetStudyPlanWeeklyRepo) FindByPresetStudyPlanID(ctx context.Context, db database.QueryExecer, ID pgtype.Text) ([]*entities.PresetStudyPlanWeekly, error) {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanWeeklyRepo.FindByPresetStudyPlanID")
	defer span.End()
	e := &entities.PresetStudyPlanWeekly{}

	fields, _ := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE preset_study_plan_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), e.TableName())

	result := []*entities.PresetStudyPlanWeekly{}
	rows, err := db.Query(ctx, query, &ID)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		c := new(entities.PresetStudyPlanWeekly)
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		result = append(result, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %w", err)
	}

	return result, nil
}

func (r *PresetStudyPlanWeeklyRepo) FindByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (*entities.PresetStudyPlanWeekly, error) {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanWeeklyRepo.FindByLessonID")
	defer span.End()
	e := &entities.PresetStudyPlanWeekly{}

	fields, values := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE lesson_id = $1", strings.Join(fields, ","), e.TableName())

	err := db.QueryRow(ctx, query, &lessonID).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return e, nil
}

func (r *PresetStudyPlanWeeklyRepo) FindByLessonIDs(ctx context.Context, db database.QueryExecer, IDs pgtype.TextArray, isAll bool) (map[pgtype.Text]*entities.PresetStudyPlanWeekly, error) {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanWeeklyRepo.FindByLessonIDs")
	defer span.End()

	e := &entities.PresetStudyPlanWeekly{}

	fields := database.GetFieldNames(e)

	query := fmt.Sprintf("SELECT %s FROM %s WHERE lesson_id = ANY($1)", strings.Join(fields, ","), e.TableName())
	if !isAll {
		query += " AND deleted_at IS NULL"
	}
	result := map[pgtype.Text]*entities.PresetStudyPlanWeekly{}
	rows, err := db.Query(ctx, query, &IDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		c := new(entities.PresetStudyPlanWeekly)
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		result[c.ID] = c
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %w", err)
	}

	return result, nil
}

func (r *PresetStudyPlanWeeklyRepo) GetIDsByLessonIDAndPresetStudyPlanIDs(ctx context.Context, db database.Ext, lessonID pgtype.Text, pspIDs pgtype.TextArray) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanWeeklyRepo.FindByLessonIDAndPresetStudyPlanIDs")
	defer span.End()

	query := fmt.Sprintf(`
		SELECT preset_study_plan_weekly_id
		FROM %s
		WHERE lesson_id = $1
			AND preset_study_plan_id = ANY($2)
			AND deleted_at IS NULL`,
		(&entities.PresetStudyPlanWeekly{}).TableName(),
	)
	rows, err := db.Query(ctx, query, lessonID, pspIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %s", err)
	}
	defer rows.Close()

	var pspwIDs []string
	for rows.Next() {
		var pspwID string
		if err := rows.Scan(&pspwID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %s", err)
		}
		pspwIDs = append(pspwIDs, pspwID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %s", err)
	}
	return pspwIDs, nil
}

func (r *PresetStudyPlanWeeklyRepo) Update(ctx context.Context, db database.QueryExecer, l *entities.PresetStudyPlanWeekly) error {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanWeeklyRepo.Update")
	defer span.End()

	query := "UPDATE preset_study_plans_weekly SET updated_at = now(), start_date = $1, end_date = $2 WHERE preset_study_plan_weekly_id = $3 AND deleted_at IS NULL"
	cmdTag, err := db.Exec(ctx, query, &l.StartDate, &l.EndDate, &l.ID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("cannot update preset study plans weekly")
	}

	return nil
}

func (r *PresetStudyPlanWeeklyRepo) UpdateTimeByLessonAndCourses(ctx context.Context, db database.Ext, lessonID pgtype.Text, courseIDs pgtype.TextArray, startDate, endDate pgtype.Timestamptz) error {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanWeeklyRepo.UpdateTimeByLessonAndCourses")
	defer span.End()

	if len(courseIDs.Elements) == 0 {
		// No courses to update
		return nil
	}

	const query = `
		UPDATE preset_study_plans_weekly
		SET start_date = $3,
			end_date = $4
		WHERE lesson_id = $1
			AND preset_study_plan_id IN (
				SELECT preset_study_plan_id
				FROM courses
				WHERE deleted_at IS NULL
					AND course_id = ANY($2)
			)
			AND deleted_at IS NULL`

	cmdTag, err := db.Exec(ctx, query, lessonID, courseIDs, startDate, endDate)
	if err != nil {
		return fmt.Errorf("db.Exec: %s", err)
	}

	// Since each (lesson_id, course_id) should have exactly one preset_study_plan_weekly_id
	// we can further check the number of rows updated
	if cmdTag.RowsAffected() != int64(len(courseIDs.Elements)) {
		return fmt.Errorf("expect %d row(s) to be updated, got %d", len(courseIDs.Elements), cmdTag.RowsAffected())
	}
	return nil
}

func (r *PresetStudyPlanWeeklyRepo) SoftDelete(ctx context.Context, db database.QueryExecer, pspwIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanWeeklyRepo.SoftDelete")
	defer span.End()

	query := "UPDATE preset_study_plans_weekly SET deleted_at = now(), updated_at = now() WHERE preset_study_plan_weekly_id = ANY($1) AND deleted_at IS NULL"
	cmdTag, err := db.Exec(ctx, query, &pspwIDs)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("cannot delete preset study plans weeklys")
	}

	return nil
}

func (r *PresetStudyPlanWeeklyRepo) SoftDeleteByPresetStudyPlanIDs(ctx context.Context, db database.QueryExecer, pspIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanWeeklyRepo.SoftDeleteByPresetStudyPlanIDs")
	defer span.End()

	query := "UPDATE preset_study_plans_weekly SET deleted_at = now(), updated_at = now() WHERE preset_study_plan_id = ANY($1) AND deleted_at IS NULL"
	cmdTag, err := db.Exec(ctx, query, &pspIDs)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("cannot delete preset study plans weeklys")
	}

	return nil
}
