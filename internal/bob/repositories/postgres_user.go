package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type PostgresUserRepo struct {
}

func (p *PostgresUserRepo) Get(ctx context.Context, db database.QueryExecer) ([]*entities.PostgresUser, error) {
	postgresUsers := entities.PostgresUsers{}
	e := new(entities.PostgresUser)
	fields := database.GetFieldNames(e)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s", strings.Join(fields, ","), e.TableName())
	err := database.Select(ctx, db, selectStmt).ScanAll(&postgresUsers)
	if err != nil {
		return nil, fmt.Errorf("error when call database.Select.ScanAll: %v", err)
	}
	return postgresUsers, nil
}

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
