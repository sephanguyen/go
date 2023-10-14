package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type QuestionGroupRepo struct{}

func (q *QuestionGroupRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.QuestionGroup) (int64, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuestionGroupRepo.Upsert")
	defer span.End()

	if e.QuestionGroupID.Status != pgtype.Present || len(e.QuestionGroupID.String) == 0 {
		e.QuestionGroupID = database.Text(idutil.ULIDNow())
	}
	fieldNames, args := e.FieldMapUpsert()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf(`
			INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT question_group_pk 
			DO UPDATE SET name = $3, description = $4, rich_description = $5, updated_at = now(), deleted_at = NULL`,
		e.TableName(), strings.Join(fieldNames, ","), placeHolders)

	cmdTag, err := db.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("exec db: %w", err)
	}

	return cmdTag.RowsAffected(), nil
}

func (q *QuestionGroupRepo) FindByID(ctx context.Context, db database.Ext, id string) (*entities.QuestionGroup, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuestionGroupRepo.FindByID")
	defer span.End()

	qr := &entities.QuestionGroup{}
	fields, values := qr.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM question_group
		WHERE question_group_id = $1
		AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)

	err := db.QueryRow(ctx, query, &id).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return qr, nil
}

func (q *QuestionGroupRepo) GetByQuestionGroupIDAndLoID(ctx context.Context, db database.QueryExecer, questionGroupID pgtype.Text, loID pgtype.Text) (*entities.QuestionGroup, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuestionGroupRepo.GetByQuestionGroupIDAndLoID")
	defer span.End()

	e := &entities.QuestionGroup{}

	fields, _ := e.FieldMap()
	stmt := fmt.Sprintf(`
			SELECT %s
			FROM %s
			WHERE deleted_at is NULL
			AND question_group_id = $1::TEXT
			AND learning_material_id = $2::TEXT`,
		strings.Join(fields, ","), e.TableName())

	if err := database.Select(ctx, db, stmt, &questionGroupID, &loID).ScanOne(e); err != nil {
		return nil, err
	}

	return e, nil
}

func (q *QuestionGroupRepo) GetQuestionGroupsByIDs(ctx context.Context, db database.QueryExecer, ids ...string) (entities.QuestionGroups, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuestionGroupRepo.GetQuestionGroupsByIDs")
	defer span.End()
	fields, _ := (&entities.QuestionGroup{}).FieldMap()
	fieldString := strings.Join(fields, ", qg.")
	fieldString = "qg." + fieldString
	query := fmt.Sprintf(`
			SELECT %s , CAST (count(q.quiz_id) AS INTEGER) AS total_children, CAST (sum(q.point) AS INTEGER) AS total_points FROM question_group qg 
			LEFT JOIN quizzes q ON q.deleted_at IS null AND qg.question_group_id = q.question_group_id  
			WHERE qg.question_group_id = ANY($1) AND qg.deleted_at IS NULL  
			GROUP BY qg.question_group_id;`, fieldString,
	)
	rows, err := db.Query(ctx, query, database.TextArray(ids))
	if err != nil {
		return nil, fmt.Errorf("db.Query :%v", err)
	}
	defer rows.Close()

	var res entities.QuestionGroups
	for rows.Next() {
		e := &entities.QuestionGroup{}
		_, values := e.FieldMap()
		values = append(values, e.TotalChildren())
		values = append(values, e.TotalPoints())
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("rows.Scan :%w", err)
		}
		res = append(res, e)
	}

	return res, nil
}

func (q *QuestionGroupRepo) DeleteByID(ctx context.Context, db database.QueryExecer, questionGroupID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "QuestionGroupRepo.DeleteByID")
	defer span.End()

	cmd, err := db.Exec(
		ctx,
		`
			UPDATE question_group
			SET deleted_at = NOW()
			WHERE question_group_id = $1
		`,
		&questionGroupID)

	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("question group not found: %w", pgx.ErrNoRows)
	}

	return nil
}
