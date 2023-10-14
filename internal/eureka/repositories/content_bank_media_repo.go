package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type ContentBankMediaRepo struct{}

func (r *ContentBankMediaRepo) Upsert(ctx context.Context, db database.QueryExecer, media *entities.ContentBankMedia) (mediaID string, err error) {
	m := entities.ContentBankMedia{}
	var id pgtype.Text
	fieldNames := []string{"id", "name", "resource", "type", "file_size_bytes", "created_by", "created_at", "updated_at"}
	placeHolders := "$1, $2, $3, $4, $5, $6, $7, $8"
	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s) 
		ON CONFLICT (name,resource_path) WHERE deleted_at is null
		DO UPDATE 
		SET resource = $3, type = $4, file_size_bytes = $5, created_by = $6, updated_at = $8
		RETURNING id;`,
		m.TableName(),
		strings.Join(fieldNames, ","),
		placeHolders,
	)

	args := database.GetScanFields(media, fieldNames)
	if err := db.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return "", fmt.Errorf("db.QueryRow: %w", err)
	}

	return id.String, nil
}

func (r *ContentBankMediaRepo) FindByMediaNames(ctx context.Context, db database.QueryExecer, mediaNames []string) ([]*entities.ContentBankMedia, error) {
	medias := make([]*entities.ContentBankMedia, 0)
	m := entities.ContentBankMedia{}
	fieldNames, _ := m.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM %s WHERE name = ANY($1) AND deleted_at is null`,
		strings.Join(fieldNames, ","),
		m.TableName(),
	)

	rows, err := db.Query(ctx, query, mediaNames)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		media := entities.ContentBankMedia{}
		if err := rows.Scan(database.GetScanFields(&media, fieldNames)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		medias = append(medias, &media)
	}

	return medias, nil
}

func (r *ContentBankMediaRepo) FindByID(ctx context.Context, db database.QueryExecer, mediaID string) (*entities.ContentBankMedia, error) {
	media := &entities.ContentBankMedia{}
	fieldNames, _ := media.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM %s WHERE id = $1 AND deleted_at is null`,
		strings.Join(fieldNames, ","),
		media.TableName(),
	)

	if err := db.QueryRow(ctx, query, mediaID).Scan(database.GetScanFields(media, fieldNames)...); err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return media, nil
}

// Soft delete
func (r *ContentBankMediaRepo) DeleteByID(ctx context.Context, db database.QueryExecer, mediaID string) error {
	media := entities.ContentBankMedia{}
	query := fmt.Sprintf(`
		UPDATE %s SET deleted_at = now() WHERE id = $1`,
		media.TableName(),
	)

	if _, err := db.Exec(ctx, query, mediaID); err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}
