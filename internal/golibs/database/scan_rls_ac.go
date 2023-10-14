package database

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/multierr"
)

const fileACStageScanDir = "./accesscontrol/stage.json"

func loadIgnoreACScanTables() (map[string]map[string]bool, error) {
	return loadIgnoreACTablesBy(fileACStageScanDir)
}

func ScanPostgresAC(ctx context.Context, db *pgxpool.Pool) error {
	existedSvc, stages, err := loadACStagesBy(fileACStageScanDir)
	if err != nil {
		return err
	}
	dbName, err := getCurrentDBName(ctx, db)
	if err != nil {
		return err
	}
	if _, valid := existedSvc[dbName]; !valid {
		return nil
	}
	tableNames := existedSvc[dbName]
	for _, tableName := range tableNames {
		policy, err := snapshotRLSPolicy(ctx, db, tableName)
		if err != nil {
			return err
		}
		schema := &tableSchema{
			TableName: tableName,
			Policies:  policy,
		}
		err = VerifyPostgresRls(dbName, schema, stages[tableName])
		if err != nil {
			return err
		}
	}

	return nil
}

const sinkDir = "/accesscontrol/connectors"

func DetectSinkTableMissingAC(ctx context.Context, db *pgxpool.Pool) error {
	dbName, err := getCurrentDBName(ctx, db)
	if err != nil {
		return err
	}

	groupedAC, err := groupACByServiceAndTable(fileACStageScanDir)
	if err != nil {
		return err
	}
	_, ok := groupedAC[dbName]
	if !ok {
		return nil
	}

	sinkTables, err := getSinkTablesBy(sinkDir, dbName)
	if err != nil {
		return err
	}

	var errs error
	for _, sinkTable := range sinkTables {
		err := verifySinkTable(sinkTable, groupedAC)
		errs = multierr.Append(errs, err)
	}

	return errs
}
