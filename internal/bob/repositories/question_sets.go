package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

// QuestionSetRepo stores
type QuestionSetRepo struct{}

// CreateAll create list of question into DB
func (repo *QuestionSetRepo) CreateAll(ctx context.Context, db database.QueryExecer, quizsets []*entities.QuestionSets) error {
	ctx, span := interceptors.StartSpan(ctx, "QuestionSetRepo.CreateAll")
	defer span.End()

	insertQuizSets := func(b *pgx.Batch, quiz *entities.QuestionSets) {
		fieldNames := database.GetFieldNames(quiz)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT quizsets_pk DO UPDATE SET updated_at = NOW()", quiz.TableName(), strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(quiz, fieldNames)...)
	}

	b := &pgx.Batch{}

	for _, v := range quizsets {
		quiz := v
		insertQuizSets(b, quiz)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(quizsets); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}

		if ct.RowsAffected() != 1 {
			return fmt.Errorf("quizsets not inserted")
		}
	}

	return nil
}

func (repo *QuestionSetRepo) FindByQuizID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.QuestionSets, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuestionSetRepo.FindByQuizID")
	defer span.End()

	e := new(entities.QuestionSets)
	fields := database.GetFieldNames(e)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE question_id = $1", strings.Join(fields, ","), e.TableName())

	row := db.QueryRow(ctx, selectStmt, &id)

	if err := row.Scan(database.GetScanFields(e, fields)...); err != nil {
		return nil, err
	}

	return e, nil
}

func (repo *QuestionSetRepo) FindByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) ([]*entities.QuestionSets, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuestionSetRepo.FindByLoID")
	defer span.End()

	e := new(entities.QuestionSets)
	fields := database.GetFieldNames(e)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE lo_id = ANY($1)", strings.Join(fields, ","), e.TableName())

	rows, err := db.Query(ctx, selectStmt, &loIDs)
	if err != nil {
		return nil, fmt.Errorf("db.QueryEx: %v", err)
	}
	defer rows.Close()

	var quizsets []*entities.QuestionSets
	for rows.Next() {
		quizset := new(entities.QuestionSets)
		if err := rows.Scan(database.GetScanFields(quizset, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %v", err)
		}
		quizsets = append(quizsets, quizset)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %v", err)
	}
	return quizsets, nil
}
