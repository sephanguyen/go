package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/mod/sumdb/dirhash"
)

const mandatoryDefaultValue = "autofillresourcepath()"

var ignoreRLSService = map[string]bool{
	"draft":   true,
	"enigma":  true,
	"unleash": true,
	"zeus":    true,
	"test":    true,
}

// checksum counts or checks the number of sql files in service.
type checksum struct {
	Count   int    `json:"count"`
	HashSum string `json:"hashsum"`
}

// writeJSON encodes data to json and writes to filename.
// Its inverse is readJSON.
func writeJSON(data interface{}, filename string) error {
	b, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, b, 0o644) //nolint:gosec
}

// loadJSON loads and decodes json from filename to data.
// Its inverse is writeJSON.
func loadJSON(filename string, data interface{}) error {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, data)
}

// SchemaVerifier verifies whether database.Entity matches the schema from migration files.
// It checks against either an actual migrated-up database or a cached schema file.
type SchemaVerifier struct {
	Service  string // name of service
	CacheDir string // path to directory containing all JSON files
}

// NewSchemaVerifier returns an initialized SchemaVerifier. SchemaVerifier reads from
// JSON cache saved at "backend/mock/testing/testdata" when checking entities.
func NewSchemaVerifier(srv string) (*SchemaVerifier, error) {
	sv := SchemaVerifier{
		Service:  srv,
		CacheDir: filepath.Join(snapshotDir, srv),
	}

	// Check if the count of sql files is still the same versus when generating schema.
	migrateDir := getMigrationDirectory(srv)
	sqlFiles, err := getFilesInDirectory(migrateDir, ".sql")
	if err != nil {
		return nil, err
	}
	if err := checkSQLFilenames(srv, sqlFiles); err != nil {
		return nil, err
	}

	cs := &checksum{}
	if err := loadJSON(filepath.Join(sv.CacheDir, schemaVersionFileName), cs); err != nil {
		return nil, fmt.Errorf("failed to load JSON from %s", filepath.Join(sv.CacheDir, schemaVersionFileName))
	}
	if cs.Count != len(sqlFiles) {
		return nil, fmt.Errorf("expected %d sql files, found %d (have you run 'make gen-db-schema'?)", cs.Count, len(sqlFiles))
	}
	hashsum, err := dirhash.HashDir(migrateDir, "string", dirhash.DefaultHash)
	if err != nil {
		return nil, fmt.Errorf("failed to compute directory hash: %s", err)
	}
	if cs.HashSum != hashsum {
		return nil, fmt.Errorf("hashsum mismatched (have you run 'make gen-db-schema'?)")
	}
	expectedVersion := cs.Count + 1000
	gotVersion, err := getVersion(sqlFiles[len(sqlFiles)-1])
	if err != nil {
		return nil, err
	}
	if expectedVersion != gotVersion {
		return nil, fmt.Errorf("expected migration file name `%d.migrate.up.sql`, got `%d.migrate.up.sql`", expectedVersion, gotVersion)
	}
	return &sv, nil
}

func checkSQLFilenames(service string, filenames []string) error {
	sqlFilenameReStr := service + `\/[0-9]+_migrate\.up\.sql$`
	sqlFilenameRe := regexp.MustCompile(sqlFilenameReStr)
	for _, fn := range filenames {
		if !sqlFilenameRe.MatchString(fn) {
			return fmt.Errorf("invalid name for migration file %q (does not match regexp %q)", fn, sqlFilenameReStr)
		}
	}
	return nil
}

func (sv *SchemaVerifier) Verify(e Entity) error {
	_, file, _, _ := runtime.Caller(0)
	ignoreFilePath := filepath.Join(filepath.Dir(file), "../../../migrations/public_tables.json")
	log.Println("===", ignoreFilePath)
	ignoreTableMap, err := LoadIgnoreTableJSON(ignoreFilePath)
	if err != nil {
		return err
	}
	if ignoreTableMap[sv.Service][e.TableName()] {
		return nil
	}
	tblSchema := &tableSchema{}
	cacheFilePath := filepath.Join(sv.CacheDir, e.TableName()+".json")
	err = loadJSON(cacheFilePath, tblSchema)
	if err != nil {
		return fmt.Errorf("failed to load schema from %s: %s", cacheFilePath, err)
	}

	err = sv.VerifyEntity(tblSchema, e)
	if err != nil {
		return fmt.Errorf("failed to verify entity %s: %w", e.TableName(), err)
	}
	return err
}

// VerifyEntity returns error if e does not match column names and types in the test database.
// It checks the values returned by e.FieldMap() against schema queried from database.
func (sv *SchemaVerifier) VerifyEntity(tblSchema *tableSchema, e Entity) error {
	if e.TableName() == "pg_catalog.pg_user" || e.TableName() == "pg_catalog.pg_namespace" {
		return nil
	}
	err := sv.checkMandatoryColumn(tblSchema)
	if err != nil {
		return err
	}
	if ok, err := tblSchema.matchEntity(e); !ok {
		return fmt.Errorf("entity %s does not match its table schema: %s", reflect.TypeOf(e), err)
	}
	return nil
}

func (sv *SchemaVerifier) checkMandatoryDefaultValue(field *fieldSchema) error {
	if !ignoreRLSService[sv.Service] &&
		field.FieldName.String == "resource_path" &&
		field.ColumnDefault.String != mandatoryDefaultValue {
		return fmt.Errorf("field %s must has default value %s", field.FieldName.String, mandatoryDefaultValue)
	}
	return nil
}

func (sv *SchemaVerifier) checkMandatoryColumn(tblSchema *tableSchema) error {
	mandatoryFields := map[string]bool{
		"resource_path": true,
	}
	count := 0
	for _, field := range tblSchema.Schema {
		if mandatoryFields[field.FieldName.String] {
			count++
		}
		err := sv.checkMandatoryDefaultValue(field)
		if err != nil {
			return fmt.Errorf("table %s has error: %w", tblSchema.TableName, err)
		}
	}
	if count != len(mandatoryFields) {
		return fmt.Errorf("expecting %d mandatory fields got %d mandatory fields on table: %s", len(mandatoryFields), count, tblSchema.TableName)
	}
	return nil
}

// getMigrationDirectory returns absolute path to "migrations/<srv>/".
func getMigrationDirectory(srv string) string {
	_, fp, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(fp), migrationLoc, srv)
}

// getInternalDirectory returns absolute path to "internal/<srv>/".
func getInternalDirectory(srv string) string {
	_, fp, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(fp), internalLoc, srv)
}

// getFilesInDirectory returns a list of all absolute filepaths in dir having file
// extension ext. For example, getFilesInDirectory("a", ".bcd") returns every file
// that matches "a/*.bcd".
func getFilesInDirectory(dir, ext string) ([]string, error) {
	result := make([]string, 0)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error while accessing %q: %v", path, err)
		}
		if info.IsDir() || filepath.Ext(path) != ext {
			return nil
		}
		result = append(result, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// getDirectories returns a list of all sub directory name in dir.
func getDirectories(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	results := make([]string, 0, len(files))
	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			results = append(results, fileInfo.Name())
		}
	}
	return results, nil
}

func getVersion(path string) (int, error) {
	paths := strings.Split(path, "/")
	fileName := paths[len(paths)-1]
	match, err := regexp.MatchString("^([1-9][0-9]{3})_migrate.up.sql", fileName)
	if !match || err != nil {
		return -1, fmt.Errorf("migration file name %s is not valid", fileName)
	}
	gotVersion, err := strconv.Atoi(strings.Split(fileName, "_")[0])
	return gotVersion, err
}
