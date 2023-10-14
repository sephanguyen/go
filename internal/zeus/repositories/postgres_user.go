package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/zeus/entities"
)

type PostgresNamespaceRepo struct{}

func (p *PostgresNamespaceRepo) Get(ctx context.Context, db database.QueryExecer) ([]*entities.PostgresNamespace, error) {
	postgresNamespaces := entities.PostgresNamespaces{}
	e := new(entities.PostgresNamespace)
	fields := database.GetFieldNames(e)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE nspname = 'public'", strings.Join(fields, ","), e.TableName())
	err := database.Select(ctx, db, selectStmt).ScanAll(&postgresNamespaces)
	if err != nil {
		return nil, fmt.Errorf("PostgresNamespaceRepo.Get: error when call database.Select.ScanAll: %v", err)
	}

	return postgresNamespaces, nil
}
