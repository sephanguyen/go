package defs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Services struct {
	services []Service

	targetService *Service
	mu            sync.Mutex // write mutex for targetService
}

// GenerateSQL generates the SQL statements based on the service definitions
// then writes to files in the specified directory.
func (s *Services) GenerateSQL(dirpath string) error {
	initSQLFilepath := filepath.Join(dirpath, "0.init.sql")
	if err := s.generateInitSQLFile(initSQLFilepath); err != nil {
		return errors.Wrap(err, "failed to generate init SQL")
	}

	for idx, svc := range s.services {
		if !svc.Postgresql.CreateDB {
			continue
		}

		// Note the idx+1, as 0 is reserved for init.sql
		fIdx := idx + 1
		sqlFilePath := filepath.Join(dirpath, fmt.Sprintf("%d.%s.sql", fIdx, svc.Name))
		if err := s.generateSQLFile(sqlFilePath, idx); err != nil {
			return errors.Wrap(err, fmt.Sprintf("generateSQLFile for %q", svc.Name))
		}

		// generate for hasura v2 in a separate file (as different db)
		// but using the same index
		if svc.Hasura.V2Enabled {
			hasurav2SQLFilePath := filepath.Join(dirpath, fmt.Sprintf("%d.%s.sql", fIdx, svc.HasuraMetadataDBName()))
			if err := s.generateHasuraV2SQLFile(hasurav2SQLFilePath, idx); err != nil {
				return errors.Wrap(err, fmt.Sprintf("generateHasuraV2SQLFile for %q", svc.Name))
			}
		}
	}
	return nil
}

func NewServicesFrom(srcFile string) (*Services, error) {
	content, err := os.ReadFile(srcFile)
	if err != nil {
		return nil, err
	}

	svcs := []Service{}
	if err := yaml.Unmarshal(content, &svcs); err != nil {
		return nil, err
	}
	return &Services{services: svcs}, nil
}

func (s *Services) generateInitSQLFile(fp string) error {
	f, err := os.Create(fp)
	if err != nil {
		return errors.Wrap(err, "os.Create")
	}
	defer f.Close()
	return s.generateInitSQL(f)
}

func (s *Services) generateInitSQL(out io.Writer) error {
	const tpl = `{{range .Databases}}{{- printf "CREATE DATABASE %s;\n" . -}}{{end}}
{{range .Users}}{{- printf "CREATE USER %q WITH PASSWORD 'example';\n" . -}}{{end}}
-- grant permission for hasura user to create schema hdb_catalog
{{range .HasuraEnabledServices}}{{- printf "GRANT ALL PRIVILEGES ON DATABASE %q TO hasura;\n" . -}}{{end}}
{{range .HasuraV2EnabledServices}}{{- printf "GRANT CREATE ON DATABASE %q TO %s;\n" .HasuraMetadataDBName .HasuraDBUser -}}{{end}}
`

	return doTemplate(tpl, s, out)
}

func (s *Services) generateSQLFile(fp string, idx int) error {
	f, err := os.Create(fp)
	if err != nil {
		return errors.Wrap(err, "os.Create")
	}
	defer f.Close()
	return s.generateSQL(f, idx)
}

func (s *Services) generateSQL(out io.Writer, idx int) error {
	svc := s.services[idx]
	if !svc.Postgresql.CreateDB {
		return fmt.Errorf("generateSQL call to no create database %q service", svc.Name)
	}

	const tpl = `{{$svc := .TargetService -}}
{{printf "\\connect %s;\n\n" $svc.Name -}}
{{range .Services}}{{.GenerateServiceGrants $svc.Name}}{{end -}}
{{$svc.GenerateHasuraGrants -}}
{{$svc.GenerateHasuraV2Grants -}}
{{$svc.GenerateKafkaGrants -}}
`

	s.mu.Lock()
	defer s.mu.Unlock()
	s.targetService = &svc
	defer func() { s.targetService = nil }()
	return doTemplate(tpl, s, out)
}

func (s *Services) generateHasuraV2SQLFile(fp string, idx int) error {
	f, err := os.Create(fp)
	if err != nil {
		return errors.Wrap(err, "os.Create")
	}
	defer f.Close()
	return s.generateHasuraV2SQL(f, idx)
}

func (s *Services) generateHasuraV2SQL(out io.Writer, idx int) error {
	svc := s.services[idx]
	if !svc.Hasura.V2Enabled {
		return fmt.Errorf("generateHasuraV2SQL to service %q where Hasura v2 is disabled", svc.Name)
	}

	const tpl = `{{printf "\\connect %s;\n\n" .HasuraMetadataDBName -}}
{{.GenerateHasuraV2MetadataGrants -}}`
	return doTemplate(tpl, &svc, out)
}

// Databases returns the names of all databases to be created, including:
// - service databases
// - hasura metadata databases
// This matches "local.databases" variable in stag-apps.hcl.
func (s *Services) Databases() []string {
	out := []string{}
	for _, v := range s.services {
		if v.Postgresql.CreateDB {
			out = append(out, v.Name)
		}
		if v.Postgresql.CreateDB && v.Hasura.V2Enabled {
			out = append(out, v.HasuraMetadataDBName())
		}
	}
	sort.Strings(out)
	return out
}

// Users returns the names of all users to be created
// This matches "local.users" variable in stag-apps.hcl.
func (s *Services) Users() []string {
	out := []string{"hasura"}
	for _, v := range s.services {
		if !v.DisableIAM {
			out = append(out, v.Name)
		}
		if v.Postgresql.CreateDB && (v.Hasura.Enabled || v.Hasura.V2Enabled) {
			out = append(out, v.HasuraDBUser())
		}
	}
	sort.Strings(out)
	return out
}

// HasuraEnabledServices returns names of the services that have Hasura.
func (s *Services) HasuraEnabledServices() []string {
	out := []string{}
	for _, v := range s.services {
		if v.Hasura.Enabled {
			out = append(out, v.Name)
		}
	}
	return out
}

// HasuraV2EnabledServices returns the services that have Hasura v2.
func (s *Services) HasuraV2EnabledServices() []Service {
	out := []Service{}
	for _, v := range s.services {
		if v.Hasura.V2Enabled {
			out = append(out, v)
		}
	}
	return out
}

func (s *Services) TargetService() (*Service, error) {
	if s.targetService == nil {
		return nil, errors.New("s.targetService is nil")
	}
	return s.targetService, nil
}

func (s *Services) Services() []Service {
	return s.services
}
