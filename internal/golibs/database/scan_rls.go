package database

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
)

func ScanRLS(ctx context.Context, db *pgxpool.Pool) error {
	tableInfoList, err := loadAllTableNames(ctx, db)
	if err != nil {
		return err
	}
	ignoreMap, err := loadIgnoreTableJSON()
	if err != nil {
		return err
	}
	dbName, err := getCurrentDBName(ctx, db)
	if err != nil {
		return err
	}
	bypassRLSAccounts, err := getBypassRLSAccount(ctx, db)
	if err != nil {
		return err
	}
	ignoreACTables, err := loadIgnoreACScanTables()
	if err != nil {
		return err
	}
	for _, table := range tableInfoList {
		policy, err := snapshotRLSPolicy(ctx, db, table.Name.String)
		if err != nil {
			return err
		}
		schema := &tableSchema{
			TableName: table.Name.String,
			Policies:  policy,
			Owner:     table.Owner.String,
			Type:      table.Type.String,
		}
		err = VerifyRLS(dbName, schema, ignoreMap, bypassRLSAccounts, ignoreACTables)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadIgnoreTableJSON() (map[string]map[string]bool, error) {
	dir := "./migrations/public_tables.json"

	// load ignore table json
	ignoreTableMap, err := LoadIgnoreTableJSON(dir)
	if err != nil {
		return nil, err
	}
	return ignoreTableMap, nil
}

func loadAllTableNames(ctx context.Context, db *pgxpool.Pool) ([]*table, error) {
	var tableInfoList []*table
	rows, err := db.Query(ctx, `SELECT table_name, table_type,
		CONCAT(tableowner, viewowner) as "owner" 
	FROM   information_schema.tables t
	LEFT JOIN pg_catalog.pg_views pv  
		ON t.table_name = pv.viewname 
		LEFT JOIN pg_catalog.pg_tables pt 
		ON t.table_name = pt.tablename 
	WHERE  table_schema = 'public'
	ORDER BY table_name;
   `)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var t table
		if err := rows.Scan(&t.Name, &t.Type, &t.Owner); err != nil {
			return nil, err
		}
		tableInfoList = append(tableInfoList, &t)
	}
	return tableInfoList, nil
}

func getCurrentDBName(ctx context.Context, db *pgxpool.Pool) (string, error) {
	var currentDBName string
	err := db.QueryRow(ctx, "select current_database()").Scan(&currentDBName)
	if err != nil {
		return "", err
	}
	results := strings.Split(currentDBName, "_")
	// remove database prefix if exist
	switch len(results) {
	case 1:
		return results[0], nil
	case 2:
		return results[1], nil
	default:
		return "", errors.New("invalid database name")
	}
}

func getBypassRLSAccount(ctx context.Context, db *pgxpool.Pool) ([]string, error) {
	accounts := []string{}
	rows, err := db.Query(ctx, `SELECT usename FROM pg_catalog.pg_user WHERE usebypassrls = true;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var account string
		err = rows.Scan(&account)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}
