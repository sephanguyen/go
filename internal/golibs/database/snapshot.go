package database

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"time"

	"github.com/manabie-com/backend/internal/golibs/try"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/mod/sumdb/dirhash"
)

// Test database parameters
const (
	dbuser     = "postgres"
	dbpassword = "postgres"
	dbhost     = "postgres"
	dbport     = "5432"
	dbname     = "testdb"
	dbMaxConns = 5
)

var (
	// Connection strings to database
	dbURI = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbuser, dbpassword, dbhost, dbport, dbname)

	// offset to migration directory
	migrationLoc = "../../../migrations"
	internalLoc  = "../../../internal"

	migrationLocDBSchema = "/backend/migrations"
)

// schemaVersionFileName is the name of the file containing the checksum value of each service.
const schemaVersionFileName = "db_schema_versioning.json"

// bypassRLSAccountFileName is the name of file containing account that can bypass RLS.
// nolint
const bypassRLSAccountFileName = "bypass_rls_account.json"

// Matcher for pgtype with PostgreSQL column data type.
// Reference: https://www.postgresql.org/docs/current/datatype.html and several others.
var typemap = map[reflect.Type]func(string) bool{
	reflect.TypeOf(&pgtype.Bool{}):        func(s string) bool { return s == "boolean" },
	reflect.TypeOf(&pgtype.Date{}):        func(s string) bool { return s == "date" },
	reflect.TypeOf(&pgtype.Float4{}):      func(s string) bool { return s == "real" },
	reflect.TypeOf(&pgtype.Int2{}):        func(s string) bool { return s == "smallint" },
	reflect.TypeOf(&pgtype.Int2Array{}):   func(s string) bool { return s == "ARRAY" }, //nolint:goconst
	reflect.TypeOf(&pgtype.Int4{}):        func(s string) bool { return s == "integer" },
	reflect.TypeOf(&pgtype.Int4Array{}):   func(s string) bool { return s == "ARRAY" }, //nolint:goconst
	reflect.TypeOf(&pgtype.Int8{}):        func(s string) bool { return s == "bigint" },
	reflect.TypeOf(&pgtype.JSON{}):        func(s string) bool { return s == "json" },
	reflect.TypeOf(&pgtype.JSONB{}):       func(s string) bool { return s == "jsonb" },
	reflect.TypeOf(&pgtype.Numeric{}):     func(s string) bool { return s == "numeric" },
	reflect.TypeOf(&pgtype.Point{}):       func(s string) bool { return s == "point" },
	reflect.TypeOf(&pgtype.Text{}):        func(s string) bool { return s == "text" || s == "character varying" },
	reflect.TypeOf(&pgtype.TextArray{}):   func(s string) bool { return s == "ARRAY" }, //nolint:goconst
	reflect.TypeOf(&pgtype.Timestamptz{}): func(s string) bool { return s == "timestamp with time zone" },
	reflect.TypeOf(&pgtype.Varchar{}):     func(s string) bool { return s == "text" || s == "character varying" },
	reflect.TypeOf(&pgtype.JSONBArray{}):  func(s string) bool { return s == "ARRAY" }, // nolint:goconst
	reflect.TypeOf(&pgtype.BoolArray{}):   func(s string) bool { return s == "ARRAY" }, // nolint:goconst
}

// Absolute path to the snapshot directory, initialized at init().
// Usually it is backend/mock/testing/testdata/.
var snapshotDir string

func init() {
	_, file, _, _ := runtime.Caller(0)
	snapshotDir = filepath.Join(filepath.Dir(file), "../../../mock/testing/testdata")
}

// newDBConn returns a connection of pgxpool.Pool to the database.
func newDBConn(ctx context.Context, connString string, maxConns int32) (*pgxpool.Pool, error) {
	cf, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("cannot read PG_CONNECTION_URI: %s", err)
	}
	cf.MaxConns = maxConns

	var pool *pgxpool.Pool
	err = try.Do(func(attempt int) (retry bool, err error) {
		pool, err = pgxpool.ConnectConfig(ctx, cf)
		if err != nil {
			time.Sleep(time.Second * 2)
			return attempt < 10, err
		}
		return false, nil
	})

	if err != nil {
		return nil, err
	}

	return pool, nil
}

type SnapShot struct {
	Service      string
	Conn         *pgxpool.Pool
	CacheDir     string
	SQLFileCount int
	SQLHashSum   string
}

type BypassRLSAccountList struct {
	Account []string `json:"account"`
}

// NewSchemaRecorder returns an initialized NewSchemaRecorder.
func NewSchemaRecorder(srv string) *SnapShot {
	sr := SnapShot{
		Service:  srv,
		CacheDir: filepath.Join("/backend/mock/testing/testdata", srv),
	}
	return &sr
}

// Record records all tables in 'public' schema from database to JSONs.
func (sr *SnapShot) Record(ctx context.Context) error {
	if err := sr.connectDB(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %s", err)
	}

	if err := sr.resetDB(ctx); err != nil {
		return fmt.Errorf("failed to reset database %s: %s", dbname, err)
	}

	if err := os.MkdirAll(sr.CacheDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create cache directory: %s", err)
	}

	if err := sr.migrateDB(ctx); err != nil {
		return fmt.Errorf("failed to perform migration: %s", err)
	}

	if err := sr.dumpSchema(); err != nil {
		return fmt.Errorf("failed to dump database: %s", err)
	}

	tbls, err := sr.getTableList(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve table list from database: %s", err)
	}

	log.Printf("Recording %d tables for service %s", len(tbls), sr.Service)
	for _, t := range tbls {
		tblName := t.Name.String
		if t.Schema.String != publicSchema {
			tblName = fmt.Sprintf("%s.%s", t.Schema.String, t.Name.String)
		}

		// snapshot table and write to file
		tblSchema, err2 := sr.snapshotTable(ctx, t)
		if err2 != nil {
			return fmt.Errorf("failed to get table %s schema: %w", tblName, err2)
		}
		cacheFilePath := filepath.Join(sr.CacheDir, tblName+".json")

		tblSchema.Policies, err2 = sr.snapshotRLSPolicy(ctx, tblName)
		if err2 != nil {
			return fmt.Errorf("failed to cache schema: %w", err2)
		}
		err2 = writeJSON(tblSchema, cacheFilePath)
		if err2 != nil {
			return fmt.Errorf("failed to store snap shot: %w", err2)
		}
	}

	// Record the number of SQL files migrated
	cs := checksum{
		Count:   sr.SQLFileCount,
		HashSum: sr.SQLHashSum,
	}
	accountList := BypassRLSAccountList{}
	accountList.Account, err = sr.snapshotBypassRLSAccount(ctx, sr.Conn)
	if err != nil {
		return fmt.Errorf("failed to get bypass RLS account list: %w", err)
	}
	err = writeJSON(accountList, filepath.Join(sr.CacheDir, bypassRLSAccountFileName))
	if err != nil {
		return fmt.Errorf("failed to store bypass RLS account list: %w", err)
	}

	return writeJSON(cs, filepath.Join(sr.CacheDir, schemaVersionFileName))
}

// connectDB connects to the database specified by dbURI.
func (sr *SnapShot) connectDB(ctx context.Context) error {
	conn, err := newDBConn(ctx, dbURI, dbMaxConns)
	if err != nil {
		return err
	}
	sr.Conn = conn
	return nil
}

// resetDB resets the test database by dropping all tables in 'public' schema.
// Reference: https://stackoverflow.com/questions/3327312/how-can-i-drop-all-the-tables-in-a-postgresql-database
func (sr *SnapShot) resetDB(ctx context.Context) error {
	resetQuery := fmt.Sprintf(`
	DROP SCHEMA public CASCADE;
	DROP SCHEMA IF EXISTS manabie CASCADE;
	DROP USER IF EXISTS %s;
	CREATE SCHEMA public;
	GRANT ALL ON SCHEMA public TO postgres;
	GRANT ALL ON SCHEMA public TO public;
	COMMENT ON SCHEMA public IS 'standard public schema';
	CREATE USER "%s" WITH PASSWORD 'example';`,
		sr.Service, sr.Service)
	_, err := sr.Conn.Exec(
		ctx,
		resetQuery,
	)
	return err
}

// migrateDB executes all .sql files in migrations/<srv_name>/ lexicographically.
// If there are any errors, the process stops immediately.
func (sr *SnapShot) migrateDB(context.Context) error {
	migrateDir := filepath.Join(migrationLocDBSchema, sr.Service)
	sqlFiles, err := getFilesInDirectory(migrateDir, ".sql")
	if err != nil {
		return err
	}

	// Do migration
	for _, file := range sqlFiles {
		err = sr.execFrom(file)
		if err != nil {
			return err
		}
	}
	sr.SQLFileCount = len(sqlFiles)
	sr.SQLHashSum, err = dirhash.HashDir(migrateDir, "string", dirhash.DefaultHash)
	if err != nil {
		return fmt.Errorf("failed to compute directory hash: %s", err)
	}
	return nil
}

// dump schema to internal/$service/
func (sr *SnapShot) dumpSchema() error {
	// this function is currently only available for eureka
	if sr.Service != "eureka" {
		return nil
	}

	outputPath := filepath.Join("/backend/internal", sr.Service, sr.Service+".sql")
	dumpCmd := exec.Command("pg_dump", "-h", "postgres", "-U", dbuser, "-d", dbname, "--schema=public", "--schema-only", "-f", outputPath)
	dumpCmd.Stdout = os.Stdout
	dumpCmd.Stderr = os.Stderr

	dumpCmd.Env = os.Environ()
	dumpCmd.Env = append(dumpCmd.Env, "PGPASSWORD="+dbpassword)

	if err := dumpCmd.Run(); err != nil {
		return fmt.Errorf("failed to run pg_dump: %s", err)
	}

	sedCmd := exec.Command("sed", "-i", "/-- Dumped by pg_dump/d", outputPath)
	sedCmd.Stdout = os.Stdout
	sedCmd.Stderr = os.Stderr
	if err := sedCmd.Run(); err != nil {
		return fmt.Errorf("failed to run sed: %s", err)
	}
	return nil
}

// getTableList returns the list of all tables belonging to public schema.
func (sr *SnapShot) getTableList(ctx context.Context) ([]*table, error) {
	tbls := tables{}
	err := Select(
		ctx,
		sr.Conn,
		fmt.Sprintf(
			`SELECT table_name, table_type,
			    CONCAT(tableowner, viewowner) as "owner", table_schema
			 FROM  %s AS t
			 LEFT JOIN pg_catalog.pg_views pv  
			 ON t.table_name = pv.viewname 
			 LEFT JOIN pg_catalog.pg_tables pt 
			 ON t.table_name = pt.tablename 
			 WHERE  table_schema in ('public', 'manabie');`,
			(&table{}).TableName(),
		),
	).ScanAll(&tbls)
	if err != nil {
		return nil, err
	}
	return tbls, nil
}

// execFrom executes SQL queries in file, mainly used in test database migration.
func (sr *SnapShot) execFrom(file string) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	sql := string(b)
	_, err = sr.Conn.Exec(context.Background(), sql)
	if err != nil {
		return fmt.Errorf("error while executing sql file %s: %s", file, err)
	}
	return nil
}

// snapshotTable returns a record of a table with columns' names and types.
func (sr *SnapShot) snapshotTable(ctx context.Context, table *table) (*tableSchema, error) {
	schema := fieldSchemas{}
	err := Select(
		ctx,
		sr.Conn,
		fmt.Sprintf(
			`SELECT column_name, data_type, column_default, is_nullable
			 FROM %s
			 WHERE table_name = ($1)
			 ORDER BY 1, 2, 3, 4;`,
			(&fieldSchema{}).TableName(),
		),
		table.Name,
	).ScanAll(&schema)
	if err != nil {
		return nil, err
	}
	constraint := fieldConstraints{}
	err2 := Select(
		ctx,
		sr.Conn,
		`SELECT T1.constraint_name as constraint_name, column_name, constraint_type  
					FROM information_schema.table_constraints T1 join
						information_schema.key_column_usage T2
					ON T1.constraint_name = T2.constraint_name
					and T1.table_name = ($1) ORDER BY 3, 2, 1 ;`,
		table.Name,
	).ScanAll(&constraint)
	if err2 != nil {
		return nil, err
	}

	return &tableSchema{Schema: schema, Constraint: constraint, TableName: table.Name.String, Type: table.Type.String, Owner: table.Owner.String}, nil
}

func snapshotRLSPolicy(ctx context.Context, conn *pgxpool.Pool, tableName string) ([]*tablePolicy, error) {
	pls := tablePolicies{}
	err := Select(
		ctx,
		conn,
		`SELECT isb.table_name as tablename, policyname, qual, with_check, relrowsecurity, relforcerowsecurity, permissive, roles
				FROM information_schema.tables isb
				LEFT JOIN pg_catalog.pg_policies pp ON
					isb.table_name = pp.tablename
				LEFT JOIN pg_catalog.pg_class pg ON
					pp.tablename::regclass = pg."oid"
				WHERE isb.table_name=$1;`,
		tableName,
	).ScanAll(&pls)
	if err != nil {
		return nil, err
	}
	return pls, nil
}

// snapshotPolicy returns all policy in database.
func (sr *SnapShot) snapshotRLSPolicy(ctx context.Context, tableName string) ([]*tablePolicy, error) {
	return snapshotRLSPolicy(ctx, sr.Conn, tableName)
}

func (sr *SnapShot) snapshotBypassRLSAccount(ctx context.Context, conn *pgxpool.Pool) ([]string, error) {
	return getBypassRLSAccount(ctx, conn)
}
