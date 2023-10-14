package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type ConversionTaskRepo struct{}

func (c *ConversionTaskRepo) CreateTasks(ctx context.Context, db database.QueryExecer, tasks []*entities.ConversionTask) error {
	b := &pgx.Batch{}
	now := time.Now().UTC()

	for _, task := range tasks {
		if err := multierr.Combine(
			task.CreatedAt.Set(now),
			task.UpdatedAt.Set(now),
		); err != nil {
			return err
		}

		names, values := task.FieldMap()

		query := fmt.Sprintf(`
			INSERT INTO %s (%s) VALUES (%s) ON CONFLICT DO NOTHING`,
			task.TableName(),
			strings.Join(names, ","),
			database.GeneratePlaceholders(len(names)),
		)

		b.Queue(query, values...)
	}

	results := db.SendBatch(ctx, b)
	defer results.Close()

	for i := 0; i < b.Len(); i++ {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("results.Exec: %w", err)
		}
	}

	return nil
}

func (c *ConversionTaskRepo) UpdateTasks(ctx context.Context, db database.QueryExecer, tasks []*entities.ConversionTask) error {
	b := &pgx.Batch{}

	for _, task := range tasks {
		query := "UPDATE conversion_tasks SET status = $1, conversion_response = $2, updated_at = NOW() WHERE task_uuid = $3"
		b.Queue(query, &task.Status, &task.ConversionResponse, &task.TaskUUID)
	}

	results := db.SendBatch(ctx, b)
	defer results.Close()

	for i := 0; i < b.Len(); i++ {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("results.Exec: %w", err)
		}
	}

	return nil
}

func (c *ConversionTaskRepo) RetrieveResourceURL(ctx context.Context, db database.QueryExecer, jobUUID pgtype.Text) (string, string, error) {
	task := &entities.ConversionTask{}

	query := fmt.Sprintf(`SELECT resource_url, resource_path FROM %s WHERE task_uuid = $1`, task.TableName())

	var resourceURL, resourcePath pgtype.Text
	if err := database.Select(ctx, db, query, &jobUUID).ScanFields(&resourceURL, &resourcePath); err != nil {
		return "", "", err
	}

	return resourceURL.String, resourcePath.String, nil
}
