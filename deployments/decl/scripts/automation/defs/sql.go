package defs

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

type SQLData struct {
	databases      map[string]int    // key: dbname, value: filename index
	users          map[string]string // key: username, value: password
	dbWithHasura   map[string]struct{}
	dbWithHasuraV2 map[string]struct{}
	svcGrants      []svcGrant
	rawStmts       map[string][]string // list of extra raw statements
}

func NewSQL() *SQLData {
	return &SQLData{
		databases:      map[string]int{},
		users:          map[string]string{},
		dbWithHasura:   map[string]struct{}{},
		dbWithHasuraV2: map[string]struct{}{},
		rawStmts:       map[string][]string{},
	}
}

func SQLFromServices(svcs []Service) (*SQLData, error) {
	res := NewSQL()
	for _, s := range svcs {
		if err := res.updateFromService(s); err != nil {
			return nil, fmt.Errorf("failed to update data from service %q: %s", s.Name, err)
		}
	}
	return res, nil
}

func (d *SQLData) updateFromService(s Service) error {
	if s.Postgresql.CreateDB {
		if err := d.AddDatabases(s.Name); err != nil {
			return err
		}
	}
	if s.Postgresql.CreateDB && s.Hasura.V2Enabled {
		if err := d.AddDatabases(s.HasuraMetadataDBName()); err != nil {
			return err
		}
	}
	if !s.DisableIAM {
		if err := d.addUser(s.Name); err != nil {
			return err
		}
	}
	if s.Postgresql.CreateDB && (s.Hasura.Enabled || s.Hasura.V2Enabled) {
		if err := d.addUser(s.HasuraDBUser()); err != nil {
			return err
		}
	}
	if s.Hasura.Enabled {
		if err := d.addDatabaseWithHasura(s.Name); err != nil {
			return err
		}
	}
	if s.Hasura.V2Enabled {
		if err := d.addDatabaseWithHasuraV2(s.Name); err != nil {
			return err
		}
	}
	for _, v := range s.Postgresql.Grants {
		d.addSvcUserPermission(v.DBName, s.Name, v.GrantDelete)
	}
	return nil
}

func (d *SQLData) AddRawStmts(filesuffix string, stmts ...string) {
	d.rawStmts[filesuffix] = append(d.rawStmts[filesuffix], stmts...)
}

func (d *SQLData) AddDatabases(dbnames ...string) error {
	for _, name := range dbnames {
		if _, exist := d.databases[name]; exist {
			return fmt.Errorf("duplicated database %q", name)
		}
		d.databases[name] = 0 // will be indexed using normalize() later
	}
	return nil
}

func (d *SQLData) addUser(dbusers ...string) error {
	for _, name := range dbusers {
		if _, exist := d.users[name]; exist {
			return fmt.Errorf("duplicated user %q", name)
		}
		d.users[name] = "example" // common password in local
	}
	return nil
}

func (d *SQLData) addDatabaseWithHasura(dbname string) error {
	if _, exist := d.dbWithHasura[dbname]; exist {
		return fmt.Errorf("duplicated database with hasura %q", dbname)
	}
	d.dbWithHasura[dbname] = struct{}{}
	return nil
}

func (d *SQLData) addDatabaseWithHasuraV2(dbname string) error {
	if _, exist := d.dbWithHasuraV2[dbname]; exist {
		return fmt.Errorf("duplicated database with hasurav2 %q", dbname)
	}
	d.dbWithHasuraV2[dbname] = struct{}{}
	return nil
}

func (d *SQLData) addSvcUserPermission(dbname string, dbuser string, candelete bool) {
	d.svcGrants = append(d.svcGrants, svcGrant{
		database:  dbname,
		user:      dbuser,
		candelete: candelete,
	})
}

func (d *SQLData) Databases() map[string]int {
	return d.databases
}

func (d *SQLData) Users() map[string]string {
	return d.users
}

func (d *SQLData) DBWithHasura() map[string]struct{} {
	return d.dbWithHasura
}

func (d *SQLData) DBWithHasuraV2() map[string]struct{} {
	return d.dbWithHasuraV2
}

func (d *SQLData) RawStmts(filename string) []string {
	res := []string{}
	for k, v := range d.rawStmts {
		if strings.HasSuffix(filename, k) {
			res = append(res, v...)
		}
	}
	return res
}

func (d *SQLData) GenerateTo(dirpath string) error {
	if err := d.generateInitSQL(dirpath); err != nil {
		return err
	}

	perdbdata := d.generatePerDBData()
	for _, v := range perdbdata {
		if err := v.generateSQL(dirpath); err != nil {
			return err
		}
	}

	return nil
}

func (d *SQLData) generatePerDBData() []sqlDataPerDB {
	d.indexMapOrder(&d.databases)
	res := make([]sqlDataPerDB, 0, len(d.databases))
	for dbname, dbindex := range d.databases {
		perdbdata := sqlDataPerDB{
			index:    dbindex,
			database: dbname,
		}
		for _, p := range d.svcGrants {
			if p.database == dbname {
				perdbdata.svcPermissions = append(perdbdata.svcPermissions, p)
			}
		}
		res = append(res, perdbdata)
	}
	return res
}

func (d *SQLData) indexMapOrder(m *map[string]int) {
	keys := make([]string, 0, len(*m))
	for k := range *m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		(*m)[k] = i + 1
	}
}

func (d *SQLData) generateInitSQL(dirpath string) error {
	outfile := "0.init.sql"
	f, err := os.Create(filepath.Join(dirpath, outfile))
	if err != nil {
		return err
	}
	defer f.Close()

	const tpl = `{{range $k, $v := .Databases}}CREATE DATABASE {{$k}};
{{end}}
{{range $k, $v := .Users}}CREATE USER "{{$k}}" WITH PASSWORD '{{$v}}';
{{end}}
-- grant permission for hasura user to create schema hdb_catalog
{{range $k, $v := .DBWithHasura}}GRANT ALL PRIVILEGES ON DATABASE "{{$k}}" TO hasura;
{{end}}
{{range $k, $v := .DBWithHasuraV2}}GRANT CREATE ON DATABASE "{{$k}}_hasura_metadata" TO {{$k}}_hasura;
{{end}}
{{range .RawStmts "0.init.sql"}}{{.}}
{{end}}
`
	t := template.Must(template.New(outfile).Option("missingkey=error").Parse(tpl))
	return t.Execute(f, d)
}

type sqlDataPerDB struct {
	index          int
	database       string
	svcPermissions []svcGrant
}

func (d *sqlDataPerDB) generateSQL(dirpath string) error {
	outfile := fmt.Sprintf("%d.%s.sql", d.index, d.database)
	f, err := os.Create(filepath.Join(dirpath, outfile))
	if err != nil {
		return err
	}
	defer f.Close()

	const tpl = `\connect {{.DBName}}

{{range .ServiceGrants}}{{.GrantStatements}}{{end}}`
	t := template.Must(template.New(outfile).Option("missingkey=error").Parse(tpl))
	return t.Execute(f, d)
}

func (d *sqlDataPerDB) DBName() (string, error) {
	if d.database == "" {
		return "", fmt.Errorf("missing database name")
	}
	return d.database, nil
}

func (d *sqlDataPerDB) ServiceGrants() []svcGrant {
	return d.svcPermissions
}

type svcGrant struct {
	database  string
	user      string
	candelete bool
}

func (sg *svcGrant) GrantStatements() (string, error) {
	tpl := `GRANT USAGE ON SCHEMA public TO {{.DBName}};
GRANT {{if .CanDelete}}SELECT, INSERT, UPDATE, DELETE{{else}}SELECT, INSERT, UPDATE{{end}} ON ALL TABLES IN SCHEMA public TO {{.DBName}};
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO {{.DBName}};
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO {{.DBName}};
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT SELECT, INSERT, UPDATE ON TABLES TO {{.DBName}};
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO {{.DBName}};
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO {{.DBName}};
	`
	t := template.Must(template.New("grant").Option("missingkey=error").Parse(tpl))
	buf := bytes.Buffer{}
	if err := t.Execute(&buf, sg); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (sg *svcGrant) DBName() (string, error) {
	if sg.database == "" {
		return "", fmt.Errorf("missing database name")
	}
	return sg.database, nil
}

func (sg *svcGrant) CanDelete() bool {
	return sg.candelete
}
