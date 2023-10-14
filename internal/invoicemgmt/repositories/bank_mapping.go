package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
)

type BankMappingRepo struct {
}

func (r *BankMappingRepo) FindAll(ctx context.Context, db database.QueryExecer) ([]*entities.BankMapping, error) {
	ctx, span := interceptors.StartSpan(ctx, "BankMappingRepo.FindAll")
	defer span.End()

	e := &entities.BankMapping{}
	fields, _ := e.FieldMap()

	resourcePath := golibs.ResourcePathFromCtx(ctx)

	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE resource_path = $1", strings.Join(fields, ","), e.TableName())

	rows, err := db.Query(ctx, stmt, resourcePath)
	if err != nil {
		return nil, err
	}

	bankMappings := []*entities.BankMapping{}
	defer rows.Close()
	for rows.Next() {
		bankMapping := new(entities.BankMapping)
		database.AllNullEntity(bankMapping)

		_, fieldValues := bankMapping.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		bankMappings = append(bankMappings, bankMapping)
	}

	return bankMappings, nil
}
