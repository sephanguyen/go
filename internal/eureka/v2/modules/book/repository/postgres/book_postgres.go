package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/constants"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type BookRepo struct {
	DB database.Ext
}

const queryPublishedBookContent = `SELECT
		b.book_id AS id,
		b.name AS name,
		json_agg(
			jsonb_build_object(
				'id', c.chapter_id,
				'name', c.name,
				'order', c.display_order,
				'topics', (
					SELECT json_agg(
							   jsonb_build_object(
								   'id', t.topic_id,
								   'name', t.name,
								   'order', t.display_order,
								   'iconUrl', t.icon_url,
								   'materials', (
									   SELECT json_agg(
												  jsonb_build_object(
														  'id', l.learning_material_id,
														  'name', l.name,
														  'order', l.display_order,
														  'type', l.type
													  )
												ORDER BY l.display_order
										  )
									   FROM learning_material AS l
									   WHERE l.topic_id = t.topic_id
										 AND l.deleted_at is NULL
										 AND l.is_published = TRUE
								   )
							   )
						   	ORDER BY t.display_order
						   )
				FROM topics AS t
				WHERE t.chapter_id = c.chapter_id
				AND t.deleted_at is NULL
				)
			)
			ORDER BY c.display_order
		)
    FILTER (WHERE c.chapter_id IS NOT NULL)
	AS chapters
	FROM
		books AS b
			LEFT JOIN
		chapters AS c ON b.book_id = c.book_id AND c.deleted_at is NULL
	WHERE
			b.book_id = $1
	AND b.deleted_at is NULL
	AND b.is_v2 = TRUE
	GROUP BY
		b.book_id, b.name;`

func (br *BookRepo) Upsert(ctx context.Context, books []domain.Book) error {
	ctx, span := interceptors.StartSpan(ctx, "BookRepo.Upsert")
	defer span.End()
	dtos, err := convertDomainsToDtos(books)
	if err != nil {
		return errors.NewConversionError("BookRepo.Upsert", err)
	}

	upsertBookQuery := `
		INSERT INTO %s (%s) 
		VALUES (%s) 
		ON CONFLICT ON CONSTRAINT books_pk DO UPDATE 
		SET 
			name = excluded.name,
			updated_at = excluded.updated_at`

	batch := &pgx.Batch{}
	bookHolder := dto.BookDto{}
	tableName := bookHolder.TableName()
	fields, _ := bookHolder.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fields))

	for _, book := range dtos {
		fields, values := book.FieldMap()
		query := fmt.Sprintf(upsertBookQuery,
			tableName, strings.Join(fields, ", "), placeHolders)
		batch.Queue(query, values...)
	}

	result := br.DB.SendBatch(ctx, batch)
	defer result.Close()

	for i := 0; i < batch.Len(); i++ {
		ct, err := result.Exec()
		if err != nil {
			return errors.NewDBError("batchResults.Exec", err)
		}
		if ct.RowsAffected() != 1 {
			return errors.NewNoRowsUpdatedError("ct.RowsAffected: Books are not upserted", nil)
		}
	}
	return nil
}

func (br *BookRepo) GetPublishedBookContent(ctx context.Context, bookID string) (domain.Book, error) {
	ctx, span := interceptors.StartSpan(ctx, "BookRepo.GetPublishedBookContent")
	defer span.End()
	var book domain.Book

	bookContent := new(dto.BookContent)

	err := database.Select(ctx, br.DB, queryPublishedBookContent, bookID).
		ScanFields(&bookContent.ID, &bookContent.Name, &bookContent.ChapterJSON)
	if err != nil {
		if errors.IsPgxNoRows(err) {
			return book, errors.NewNoRowsExistedError("BookRepo.GetPublishedBookContent", nil)
		}
		return book, errors.NewDBError("BookRepo.GetPublishedBookContent", err)
	}

	var chapterDtos []dto.Chapter
	if bookContent.ChapterJSON.Status == pgtype.Present {
		err = json.Unmarshal([]byte(bookContent.ChapterJSON.String), &chapterDtos)
		if err != nil {
			return book, errors.NewConversionError("BookRepo.GetPublishedBookContent", err)
		}
	}

	return buildPublishedBookContent(bookContent.ID.String, bookContent.Name.String, chapterDtos), nil
}

func convertDomainsToDtos(books []domain.Book) ([]dto.BookDto, error) {
	bookDtos := make([]dto.BookDto, len(books))
	now := time.Now()

	for i, book := range books {
		bookDto, err := dto.NewBookDtoFromEntity(book)
		if err != nil {
			return nil, fmt.Errorf("NewBookDtoFromEntity: %w", err)
		}
		err = multierr.Combine(
			bookDto.CreatedAt.Set(now),
			bookDto.UpdatedAt.Set(now),
		)
		if err != nil {
			return nil, fmt.Errorf("NewBookDtoFromEntity: %w", err)
		}
		bookDtos[i] = bookDto
	}

	return bookDtos, nil
}

func buildPublishedBookContent(bookID, bookName string, chapterDtos []dto.Chapter) domain.Book {
	chapters := make([]domain.Chapter, len(chapterDtos))

	for ic, c := range chapterDtos {
		topics := make([]domain.Topic, len(c.Topics))

		for it, t := range c.Topics {
			materials := make([]domain.LearningMaterial, len(t.LearningMaterials))

			for im, m := range t.LearningMaterials {
				material := domain.LearningMaterial{
					ID:           m.ID,
					Name:         m.Name,
					DisplayOrder: m.Order,
					Published:    true,
					Type:         constants.LearningMaterialType(m.Type),
				}
				materials[im] = material
			}

			topic := domain.Topic{
				ID:                t.ID,
				Name:              t.Name,
				DisplayOrder:      t.Order,
				IconURL:           t.IconURL,
				LearningMaterials: materials,
			}

			topics[it] = topic
		}

		chapter := domain.Chapter{
			ID:           c.ID,
			Name:         c.Name,
			DisplayOrder: c.Order,
			Topics:       topics,
		}

		chapters[ic] = chapter
	}

	return domain.Book{
		ID:       bookID,
		Name:     bookName,
		Chapters: chapters,
	}
}

const queryGetBookHierarchyFlattenByLearningMaterialID = `
		SELECT b.book_id, c.chapter_id, t.topic_id, lm.learning_material_id
		FROM
			books b
			JOIN chapters c
				USING (book_id)
			JOIN topics t
				USING(chapter_id)
			JOIN learning_material lm
				USING (topic_id)
			WHERE COALESCE(b.deleted_at, c.deleted_at, t.deleted_at, lm.deleted_at) IS NULL
			AND learning_material_id = $1::TEXT
		LIMIT 1`

func (br *BookRepo) GetBookHierarchyFlattenByLearningMaterialID(ctx context.Context, learningMaterialID string) (domain.BookHierarchyFlatten, error) {
	bHierarchyFlatten := new(dto.BookHierarchyFlatten)

	err := database.Select(ctx, br.DB, queryGetBookHierarchyFlattenByLearningMaterialID, learningMaterialID).ScanFields(&bHierarchyFlatten.BookID, &bHierarchyFlatten.ChapterID, &bHierarchyFlatten.TopicID, &bHierarchyFlatten.LearningMaterialID)

	if err != nil {
		if errors.IsPgxNoRows(err) {
			return domain.BookHierarchyFlatten{}, errors.NewNoRowsExistedError("BookRepo.GetBookHierarchyFlattenByLearningMaterialID", nil)
		}

		return domain.BookHierarchyFlatten{}, errors.NewDBError("BookRepo.GetBookHierarchyFlattenByLearningMaterialID database.Select", err)
	}

	return bHierarchyFlatten.ToEntity(), nil
}
