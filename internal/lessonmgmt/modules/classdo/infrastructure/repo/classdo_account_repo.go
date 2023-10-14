package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/domain"

	"github.com/jackc/pgx/v4"
)

type ClassDoAccountRepo struct{}

func (c *ClassDoAccountRepo) UpsertClassDoAccounts(ctx context.Context, db database.QueryExecer, classDoAccounts domain.ClassDoAccounts) error {
	ctx, span := interceptors.StartSpan(ctx, "ClassDoAccountRepo.UpsertClassDoAccounts")
	defer span.End()

	b := &pgx.Batch{}
	for _, classDoAccount := range classDoAccounts {
		dto, err := NewClassDoAccountFromDomain(classDoAccount)
		if err != nil {
			return err
		}
		c.UpsertQueue(b, dto)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()
	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}

	return nil
}

func (c *ClassDoAccountRepo) UpsertQueue(b *pgx.Batch, e *ClassDoAccount) {
	fields, values := e.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fields))

	query := fmt.Sprintf(`INSERT INTO %s (%s) 
			VALUES (%s) ON CONFLICT ON CONSTRAINT pk__classdo_account DO 
			UPDATE SET classdo_email = $2, classdo_api_key = $3, updated_at = now(), deleted_at = $6 `,
		e.TableName(),
		strings.Join(fields, ", "),
		placeHolders,
	)

	b.Queue(query, values...)
}

func (c *ClassDoAccountRepo) GetAllClassDoAccounts(ctx context.Context, db database.QueryExecer) ([]*ClassDoAccount, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassDoAccountRepo.GetAllClassDoAccounts")
	defer span.End()

	classDoAccount := &ClassDoAccount{}
	fields, _ := classDoAccount.FieldMap()

	query := fmt.Sprintf(`SELECT %s 
			FROM %s WHERE deleted_at IS NULL`,
		strings.Join(fields, ","),
		classDoAccount.TableName(),
	)

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	allClassDoAccounts := []*ClassDoAccount{}
	for rows.Next() {
		item := &ClassDoAccount{}
		if err := rows.Scan(database.GetScanFields(item, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		allClassDoAccounts = append(allClassDoAccounts, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return allClassDoAccounts, nil
}

func (c *ClassDoAccountRepo) GetClassDoAccountByID(ctx context.Context, db database.QueryExecer, id string) (*ClassDoAccount, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassDoAccountRepo.GetClassDoAccountByID")
	defer span.End()

	classDoAccount := &ClassDoAccount{}
	fields, values := classDoAccount.FieldMap()

	query := fmt.Sprintf(`SELECT %s 
			FROM %s WHERE classdo_id = $1
			    	AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		classDoAccount.TableName(),
	)

	err := db.QueryRow(ctx, query, id).Scan(values...)

	if err != nil {
		return nil, fmt.Errorf("db.QueryRow %s: %s", id, err)
	}

	return classDoAccount, nil
}
