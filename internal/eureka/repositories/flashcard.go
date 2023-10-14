package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	dbeureka "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type FlashcardRepo struct{}

func (s *FlashcardRepo) Insert(ctx context.Context, db database.QueryExecer, e *entities.Flashcard) error {
	ctx, span := interceptors.StartSpan(ctx, "FlashcardRepo.Insert")
	defer span.End()
	if _, err := database.Insert(ctx, e, db.Exec); err != nil {
		return fmt.Errorf("database.Insert: %w", err)
	}
	return nil
}

func (s *FlashcardRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.Flashcard) error {
	ctx, span := interceptors.StartSpan(ctx, "FlashcardRepo.Update")
	defer span.End()
	if _, err := database.UpdateFields(ctx, e, db.Exec, "learning_material_id", []string{"name", "updated_at"}); err != nil {
		return fmt.Errorf("database.UpdateFields: %w", err)
	}
	return nil
}

type ListFlashcardArgs struct {
	LearningMaterialIDs pgtype.TextArray
}

func (s *FlashcardRepo) ListFlashcard(ctx context.Context, db database.QueryExecer, args *ListFlashcardArgs) ([]*entities.Flashcard, error) {
	ctx, span := interceptors.StartSpan(ctx, "FlashcardRepo.ListFlashcard")
	defer span.End()
	b := &entities.Flashcard{}
	fields, _ := b.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s
			FROM %s
		WHERE
			deleted_at IS NULL AND learning_material_id = ANY($1::_TEXT)
		ORDER BY display_order ASC, created_at DESC 
		`, strings.Join(fields, ", "), b.TableName())
	flashcards := entities.Flashcards{}
	if err := database.Select(ctx, db, query, &args.LearningMaterialIDs).ScanAll(&flashcards); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return flashcards, nil
}

func (s *FlashcardRepo) ListFlashcardBase(ctx context.Context, db database.QueryExecer, args *ListFlashcardArgs) ([]*entities.FlashcardBase, error) {
	ctx, span := interceptors.StartSpan(ctx, "FlashcardRepo.ListFlashcardBase")
	defer span.End()
	b := &entities.Flashcard{}

	stmt := fmt.Sprintf(`
		SELECT fc.%s, array_length(qs.quiz_external_ids, 1) as total_question
		FROM %s fc
		LEFT JOIN 
			quiz_sets qs ON qs.lo_id = fc.learning_material_id AND qs.deleted_at IS NULL
		WHERE
			fc.deleted_at IS NULL AND fc.learning_material_id = ANY($1::_TEXT)
		ORDER BY fc.display_order ASC, fc.created_at DESC 
		`,
		strings.Join(database.GetFieldNames(b), ", fc."),
		b.TableName(),
	)

	bases := make([]*entities.FlashcardBase, 0)

	rows, err := db.Query(ctx, stmt, args.LearningMaterialIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		b := &entities.FlashcardBase{}

		_, values := b.Flashcard.FieldMap()
		values = append(values, &b.TotalQuestion)

		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		bases = append(bases, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return bases, nil
}

func (s *FlashcardRepo) ListByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.Flashcard, error) {
	ctx, span := interceptors.StartSpan(ctx, "FlashcardRepo.ListByTopicIDs")
	defer span.End()
	flashcards := &entities.Flashcards{}
	e := &entities.Flashcard{}
	query := fmt.Sprintf(queryListByTopicIDs, strings.Join(database.GetFieldNames(e), ","), e.TableName())
	if err := database.Select(ctx, db, query, topicIDs).ScanAll(flashcards); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return *flashcards, nil
}

func (m *FlashcardRepo) BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.Flashcard) error {
	err := dbeureka.BulkUpsert(ctx, db, bulkInsertQuery, items)
	if err != nil {
		return fmt.Errorf("FlashcardRepo database.BulkInsert error: %s", err.Error())
	}
	return nil
}
