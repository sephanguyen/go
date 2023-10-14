package defs

import (
	"bytes"
	"io"
	"text/template"

	"github.com/pkg/errors"
)

type Service struct {
	Name       string `yaml:"name"`
	Postgresql struct {
		CreateDB bool `yaml:"createdb"`
		Grants   []struct {
			DBName      string `yaml:"dbname"`
			GrantDelete bool   `yaml:"grant_delete"`
		} `yaml:"grants"`
	} `yaml:"postgresql"`
	Hasura struct {
		Enabled   bool `yaml:"enabled"`
		V2Enabled bool `yaml:"v2_enabled"`
	} `yaml:"hasura"`
	Kafka struct {
		Enabled     bool `yaml:"enabled"`
		GrantDelete bool `yaml:"grant_delete"`
	} `yaml:"kafka"`
	J4 struct {
		AllowDBAccess bool `yaml:"allow_db_access"`
	}
	DisableIAM bool `yaml:"disable_iam"`
}

func (s *Service) HasuraMetadataDBName() string {
	return s.Name + "_hasura_metadata"
}

func (s *Service) HasuraDBUser() string {
	return s.Name + "_hasura"
}

func (s *Service) GenerateServiceGrants(dbname string) (string, error) {
	for _, grant := range s.Postgresql.Grants {
		if grant.DBName == dbname {
			return s.generateGrantSQL(s.Name, grant.GrantDelete)
		}
	}
	return "", nil
}

func (s *Service) generateGrantSQL(svcName string, candelete bool) (string, error) {
	type data struct {
		Name      string
		CanDelete bool
	}
	const tpl = `GRANT USAGE ON SCHEMA public TO {{.Name}};
GRANT SELECT, INSERT, UPDATE{{if .CanDelete}}, DELETE{{end}} ON ALL TABLES IN SCHEMA public TO {{.Name}};
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO {{.Name}};
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO {{.Name}};
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT SELECT, INSERT, UPDATE{{if .CanDelete}}, DELETE{{end}} ON TABLES TO {{.Name}};
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO {{.Name}};
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO {{.Name}};

`
	return doTemplateS(tpl, data{Name: svcName, CanDelete: candelete})
}

func (s *Service) GenerateHasuraGrants() (string, error) {
	if !s.Hasura.Enabled {
		return "", nil
	}
	const tpl = `GRANT USAGE ON SCHEMA public TO hasura;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO hasura;
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO hasura;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO hasura;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT SELECT, INSERT, UPDATE ON TABLES TO hasura;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO hasura;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO hasura;

`
	return doTemplateS(tpl, s)
}

func (s *Service) GenerateHasuraV2Grants() (string, error) {
	if !s.Hasura.Enabled && !s.Hasura.V2Enabled {
		return "", nil
	}
	const tpl = `GRANT USAGE ON SCHEMA public TO {{.HasuraDBUser}};
GRANT SELECT, INSERT, UPDATE{{if eq .Name "draft"}}, DELETE{{end}} ON ALL TABLES IN SCHEMA public TO {{.HasuraDBUser}};
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO {{.HasuraDBUser}};
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO {{.HasuraDBUser}};
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT SELECT, INSERT, UPDATE ON TABLES TO {{.HasuraDBUser}};
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO {{.HasuraDBUser}};
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO {{.HasuraDBUser}};

`
	return doTemplateS(tpl, s)
}

func (s *Service) GenerateHasuraV2MetadataGrants() (string, error) {
	if !s.Hasura.V2Enabled {
		return "", nil
	}
	const tpl = `{{if .Hasura.V2Enabled}}GRANT USAGE ON SCHEMA public TO {{.HasuraDBUser}};
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO {{.HasuraDBUser}};
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO {{.HasuraDBUser}};
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO {{.HasuraDBUser}};
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO {{.HasuraDBUser}};
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO {{.HasuraDBUser}};
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO {{.HasuraDBUser}};

{{end}}`
	return doTemplateS(tpl, s)
}

func (s *Service) GenerateKafkaGrants() (string, error) {
	if !s.Kafka.Enabled {
		return "", nil
	}
	const tpl = `{{if .Kafka.Enabled }}GRANT USAGE ON SCHEMA public TO kafka_connector;
GRANT SELECT, INSERT, UPDATE{{if .Kafka.GrantDelete}}, DELETE{{end}} ON ALL TABLES IN SCHEMA public TO kafka_connector;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO kafka_connector;

{{end}}`
	return doTemplateS(tpl, s)
}

// doTemplate is a helper function to quickly parse and execute template.
func doTemplate(tpl string, data any, out io.Writer) error {
	t, err := template.New("").Option("missingkey=error").Parse(tpl)
	if err != nil {
		return errors.Wrap(err, "template.New")
	}
	if err := t.Execute(out, data); err != nil {
		return errors.Wrap(err, "template.Execute")
	}
	return nil
}

// doTemplateS is similar to doTemplate but output to string instead.
func doTemplateS(tpl string, data any) (string, error) {
	buf := bytes.Buffer{}
	if err := doTemplate(tpl, data, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
