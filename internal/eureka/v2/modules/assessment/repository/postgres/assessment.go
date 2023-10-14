package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgx/v4"
)

type AssessmentRepo struct{}

func (a *AssessmentRepo) GetManyByIDs(ctx context.Context, db database.Ext, ids []string) ([]domain.Assessment, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssessmentRepo.GetManyByIDs")
	defer span.End()

	asm := &dto.Assessment{}
	fields, _ := asm.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE id = ANY($1)
		AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		asm.TableName(),
	)

	rows, err := db.Query(ctx, query, database.TextArray(ids))
	if err != nil {
		return nil, errors.NewDBError("AssessmentRepo.GetManyByIDs", err)
	}

	return scanAssessment(rows)
}

func (a *AssessmentRepo) GetManyByLMAndCourseIDs(ctx context.Context, db database.Ext, assessments []domain.Assessment) ([]domain.Assessment, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssessmentRepo.GetManyByLMAndCourseIDs")
	defer span.End()

	placeholders := make([]string, 0, len(assessments))
	values := make([]interface{}, 0, len(assessments))

	for i, asm := range assessments {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", 2*i+1, 2*i+2))
		values = append(values, asm.LearningMaterialID, asm.CourseID)
	}
	asm := &dto.Assessment{}
	fields, _ := asm.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE (learning_material_id, course_id) IN
		(
			%s
		)
		AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		asm.TableName(),
		strings.Join(placeholders, ", "),
	)

	rows, err := db.Query(ctx, query, values...)
	if err != nil {
		return nil, errors.NewDBError("AssessmentRepo.scanAssessment", err)
	}

	return scanAssessment(rows)
}

func scanAssessment(rows pgx.Rows) ([]domain.Assessment, error) {
	var cfs []domain.Assessment
	asm := &dto.Assessment{}
	fields, _ := asm.FieldMap()

	defer rows.Close()
	for rows.Next() {
		cf := new(dto.Assessment)
		if err := rows.Scan(database.GetScanFields(cf, fields)...); err != nil {
			return nil, errors.NewConversionError("AssessmentRepo.scanAssessment", err)
		}
		a, err := cf.ToEntity()
		if err != nil {
			return nil, errors.New("cf.ToEntity", err)
		}
		cfs = append(cfs, a)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.NewConversionError("AssessmentRepo.scanAssessment", err)
	}
	return cfs, nil
}

func (a *AssessmentRepo) GetOneByLMAndCourseID(ctx context.Context, db database.Ext, courseID, lmID string) (*domain.Assessment, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssessmentRepo.GetOneByLMAndCourseIDs")
	defer span.End()

	var result dto.Assessment

	stmt := fmt.Sprintf(`
        SELECT %s
          FROM %s
         WHERE deleted_at IS NULL
           AND course_id = $1
           AND learning_material_id = $2;
	`, strings.Join(database.GetFieldNames(&result), ", "), result.TableName())

	if err := database.Select(ctx, db, stmt, database.Text(courseID), database.Text(lmID)).ScanOne(&result); err != nil {
		if errors.IsPgxNoRows(err) {
			return nil, errors.NewNoRowsExistedError("database.Select", err)
		}
		return nil, errors.NewDBError("database.Select", err)
	}

	assessment, err := result.ToEntity()
	if err != nil {
		return nil, errors.NewConversionError("result.ToEntity", err)
	}

	return &assessment, nil
}

func (a *AssessmentRepo) Upsert(ctx context.Context, db database.Ext, now time.Time, assessment domain.Assessment) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssessmentRepo.Insert")
	defer span.End()

	var returnID string

	assessmentDto := dto.Assessment{}
	if err := assessmentDto.FromEntity(now, assessment); err != nil {
		return "", errors.NewConversionError("assessmentDto.FromEntity", err)
	}

	fields, values := assessmentDto.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fields))

	stmt := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES(%s)
		ON CONFLICT ON CONSTRAINT un_lm_course DO
		UPDATE SET deleted_at = NULL, updated_at = EXCLUDED.updated_at
		RETURNING id;`,
		assessmentDto.TableName(),
		strings.Join(fields, ","),
		placeHolders,
	)

	if err := db.QueryRow(ctx, stmt, values...).Scan(&returnID); err != nil {
		return "", errors.NewDBError("db.QueryRow", err)
	}

	return returnID, nil
}

func (a *AssessmentRepo) GetVirtualByID(ctx context.Context, db database.Ext, id string) (domain.Assessment, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssessmentRepo.GetVirtualByID")
	defer span.End()

	var assessment dto.Assessment
	var result dto.AssessmentExtended

	stmt := fmt.Sprintf(`
        SELECT ass.%s
             , CASE
                   WHEN ass.ref_table = 'learning_objective' THEN lo.manual_grading
                END AS manual_grading
          FROM %s ass
                 -- learning_objective
                 JOIN learning_objective lo ON lo.learning_material_id = ass.learning_material_id
                    AND ass.ref_table = 'learning_objective'
         WHERE ass.deleted_at IS NULL
           AND ass.id = $1
        ;
	`, strings.Join(database.GetFieldNames(&assessment), ", ass."), assessment.TableName())

	if err := database.Select(ctx, db, stmt, database.Text(id)).ScanOne(&result); err != nil {
		if errors.IsPgxNoRows(err) {
			return domain.Assessment{}, errors.NewNoRowsExistedError("database.Select", err)
		}
		return domain.Assessment{}, errors.NewDBError("database.Select", err)
	}

	assessmentVirtual, err := result.ToEntity()
	if err != nil {
		return domain.Assessment{}, errors.NewConversionError("result.ToEntity", err)
	}

	return assessmentVirtual, nil
}
