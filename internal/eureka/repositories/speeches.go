package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/yasuo/entities"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type SpeechesRepository struct {
}

type SpeechConfig struct {
	Language string `json:"language"`
}

func (s *SpeechesRepository) UpsertSpeeches(ctx context.Context, db database.QueryExecer, data []*entities.Speeches) ([]*entities.Speeches, error) {
	ctx, span := interceptors.StartSpan(ctx, " SpeechesRepository.Upsert")
	defer span.End()

	queueFn := func(b *pgx.Batch, s *entities.Speeches) {
		fieldNames := database.GetFieldNames(s)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`
			INSERT INTO flashcard_speeches(%s) VALUES(%s)
				ON CONFLICT ON CONSTRAINT flashcard_speeches_pk
			DO UPDATE SET
				sentence = $2,
				link = $3,
				settings = $4,
				type = $5,
				quiz_id = $6,
				created_at = $7,
				updated_at = $8,
				deleted_at = $9,
				created_by = $10,
				updated_by = $11
			RETURNING %s
		`,
			strings.Join(fieldNames, ","),
			placeHolders,
			strings.Join(fieldNames, ","),
		)
		b.Queue(query, database.GetScanFields(s, fieldNames)...)
	}

	b := &pgx.Batch{}
	var d pgtype.Timestamptz
	if err := d.Set(time.Now()); err != nil {
		return nil, fmt.Errorf("cannot set time for quiz: %w", err)
	}

	for _, each := range data {
		each.CreatedAt = d
		each.UpdatedAt = d

		queueFn(b, each)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	resp := make([]*entities.Speeches, 0)
	for i := 0; i < b.Len(); i++ {
		speech := new(entities.Speeches)
		_, values := speech.FieldMap()
		if err := result.QueryRow().Scan(values...); err != nil {
			return nil, fmt.Errorf("batchResults.QueryRow: %w", err)
		}

		resp = append(resp, speech)
	}

	return resp, nil
}

type CheckExistedSpeechReq struct {
	Text   pgtype.Text
	Config pgtype.JSONB
}

func (s *SpeechesRepository) CheckExistedSpeech(ctx context.Context, db database.QueryExecer, input *CheckExistedSpeechReq) (bool, *entities.Speeches) {
	ctx, span := interceptors.StartSpan(ctx, "SpeechesRepository.CheckExistedSpeech")
	defer span.End()

	e := new(entities.Speeches)
	query := "SELECT link FROM flashcard_speeches WHERE sentence = $1::TEXT AND settings = $2::JSONB AND deleted_at IS NULL LIMIT 1"

	err := db.QueryRow(ctx, query, &input.Text, &input.Config).Scan(&e.Link)
	if err != nil {
		return false, nil
	}

	if e.Link.String == "" || e.Link.Status == pgtype.Null {
		return false, nil
	}

	return true, e
}

func (s *SpeechesRepository) RetrieveSpeeches(ctx context.Context, db database.QueryExecer, sentences pgtype.TextArray, settings pgtype.JSONBArray) ([]*entities.Speeches, error) {
	ctx, span := interceptors.StartSpan(ctx, "SpeechesRepository.RetrieveSpeeches")
	defer span.End()

	e := new(entities.Speeches)
	fieldNames, _ := e.FieldMap()
	query := fmt.Sprintf(`
		SELECT DISTINCT ON (sentence, settings) sp.%s
		FROM %s sp
		JOIN UNNEST($1::TEXT[], $2::JSONB[]) AS un(sentence, settings)
		ON un.sentence = sp.sentence AND COALESCE(un.settings, '{}') = COALESCE(sp.settings,'{}')
		WHERE NULLIF(link, '') IS NOT NULL   
		AND deleted_at IS NULL
			`, strings.Join(fieldNames, ", sp."), e.TableName())
	var result []*entities.Speeches
	rows, err := db.Query(ctx, query, sentences, settings)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var row entities.Speeches
		if err := rows.Scan(database.GetScanFields(&row, fieldNames)...); err != nil {
			return nil, err
		}
		result = append(result, &row)
	}
	return result, nil
}

func (s *SpeechesRepository) RetrieveAllSpeaches(ctx context.Context, db database.QueryExecer, limit, offset pgtype.Int8) ([]*entities.Speeches, error) {
	ctx, span := interceptors.StartSpan(ctx, "SpeechesRepository.RetrieveAllSpeaches")
	defer span.End()

	e := new(entities.Speeches)
	fieldNames, _ := e.FieldMap()

	query := fmt.Sprintf(`
	SELECT DISTINCT ON (sentence, settings) %s
	FROM %s sp
	WHERE NULLIF(link, '') IS NOT NULL
	AND deleted_at IS NULL
	LIMIT $1
	OFFSET $2
		`, strings.Join(fieldNames, ","), e.TableName())
	var result []*entities.Speeches
	rows, err := db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	for rows.Next() {
		var row entities.Speeches
		if err := rows.Scan(database.GetScanFields(&row, fieldNames)...); err != nil {
			return nil, err
		}
		result = append(result, &row)
	}
	return result, nil
}
