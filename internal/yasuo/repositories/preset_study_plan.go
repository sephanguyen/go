package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type PresetStudyPlanRepo struct{}

func (r *PresetStudyPlanRepo) Upsert(ctx context.Context, db database.Ext, preset *entities.PresetStudyPlan) error {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanRepo.Upsert")
	defer span.End()

	now := timeutil.Now()
	err := multierr.Combine(
		preset.CreatedAt.Set(now),
		preset.UpdatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}
	fieldNames := database.GetFieldNames(preset)
	placeHolders := "$1, $2, $3, $4, $5, $6, $7, $8, $9"

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (preset_study_plan_id) DO UPDATE SET name = $2, country = $3, grade = $4, subject = $5, updated_at = $6, start_date = $8, deleted_at = $9", preset.TableName(), strings.Join(fieldNames, ","), placeHolders)
	args := database.GetScanFields(preset, fieldNames)

	ct, err := db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return errors.New("cannot insert preset study plan")
	}
	return nil
}

func (r *PresetStudyPlanRepo) Get(ctx context.Context, db database.QueryExecer, pspID pgtype.Text) (*entities.PresetStudyPlan, error) {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanRepo.Get")
	defer span.End()

	e := &entities.PresetStudyPlan{}
	fields, values := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE preset_study_plan_id = $1", strings.Join(fields, ","), e.TableName())

	err := db.QueryRow(ctx, query, &pspID).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return e, nil
}

func (r *PresetStudyPlanRepo) FindByIDs(ctx context.Context, db database.QueryExecer, pspIDs pgtype.TextArray) (map[pgtype.Text]*entities.PresetStudyPlan, error) {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanRepo.FindByIDs")
	defer span.End()

	e := new(entities.PresetStudyPlan)
	fields := database.GetFieldNames(e)

	query := fmt.Sprintf("SELECT %s FROM %s WHERE preset_study_plan_id = ANY($1) AND deleted_at IS NULL", strings.Join(fields, ","), e.TableName())

	p := map[pgtype.Text]*entities.PresetStudyPlan{}
	rows, err := db.Query(ctx, query, &pspIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		c := new(entities.PresetStudyPlan)
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		p[c.ID] = c
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %w", err)
	}

	return p, nil
}

// FindByCourseIDs returns preset_study_plan by course_id in a map.
func (r *PresetStudyPlanRepo) FindByCourseIDs(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.PresetStudyPlan, error) {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanRepo.FindByCourseIDs")
	defer span.End()

	fields, _ := (&entities.PresetStudyPlan{}).FieldMap()
	query := fmt.Sprintf(`
		SELECT tmp.course_id, psp.%s
		FROM preset_study_plans psp
		JOIN (
			SELECT course_id, preset_study_plan_id
			FROM courses
			WHERE course_id = ANY($1)
				AND deleted_at IS NULL
		) tmp USING(preset_study_plan_id)
		WHERE deleted_at IS NULL`,
		strings.Join(fields, ", psp."),
	)
	rows, err := db.Query(ctx, query, courseIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %s", err)
	}
	defer rows.Close()

	pspByCourseID := make(map[pgtype.Text]*entities.PresetStudyPlan)
	for rows.Next() {
		courseID := pgtype.Text{}
		psp := new(entities.PresetStudyPlan)
		scanFields := []interface{}{&courseID}
		_, pspScanFields := psp.FieldMap()
		scanFields = append(scanFields, pspScanFields...)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %s", err)
		}
		pspByCourseID[courseID] = psp
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %s", err)
	}
	return pspByCourseID, nil
}

func (r *PresetStudyPlanRepo) SoftDelete(ctx context.Context, db database.QueryExecer, pspIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanRepo.SoftDelete")
	defer span.End()

	query := "UPDATE preset_study_plans SET deleted_at = now(), updated_at = now() WHERE preset_study_plan_id = ANY($1) AND deleted_at IS NULL"
	cmdTag, err := db.Exec(ctx, query, &pspIDs)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("cannot delete preset study plans")
	}

	return nil
}
