// This file contains the new configuration approach for Postgres.
// TODO @anhpngt: merge to common.go
package configs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/sqladmin/v1"
	"gopkg.in/yaml.v3"
)

// PostgresConfigV2 is used to connect to one or many Postgres databases.
type PostgresConfigV2 struct {
	// Databases contains the config to connect to a single database by database name.
	Databases map[string]PostgresDatabaseConfig `yaml:",inline"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
// It is to implement a deep-merge behavior when chain-unmarshalling.
func (pc *PostgresConfigV2) UnmarshalYAML(v *yaml.Node) error {
	type yamlPostgresConfigV2 PostgresConfigV2 // prevent infinite recursion
	newConfig := yamlPostgresConfigV2{}
	if err := v.Decode(&newConfig); err != nil {
		return err
	}
	newConfig2 := PostgresConfigV2(newConfig)
	return pc.mergeFrom(&newConfig2)
}

// mergeFrom deep-merges, in-place, other into pc.
// Values from other override values in pc.
func (pc *PostgresConfigV2) mergeFrom(other *PostgresConfigV2) error {
	if pc.Databases == nil {
		pc.Databases = make(map[string]PostgresDatabaseConfig)
	}
	for dbname, dbconf2 := range other.Databases {
		mergedDBConf := pc.Databases[dbname].merge(dbconf2)
		pc.Databases[dbname] = *mergedDBConf
	}
	return nil
}

// PostgresDatabaseConfig represents config to connect to a Postgres database.
// When Instance is specified, it is assumed that we are connecting to a
// Cloud SQL instance from Google Cloud Platform.
//
// Use database.NewPool to begin a database connection with this config.
type PostgresDatabaseConfig struct {
	// CloudSQLInstance is the fully qualified instance name to connect to.
	// Format: "project:region:name"
	// Example: "staging-manabie-online:asia-southeast1:manabie-common-88e1ee71"
	CloudSQLInstance string `yaml:"cloudsql_instance"`

	// CloudSQLUsePublicIP indicates whether to use public IP or private IP (VPC)
	// to connect to the Cloud SQL instance.
	// See https://pkg.go.dev/cloud.google.com/go/cloudsqlconn#WithPrivateIP
	CloudSQLUsePublicIP bool `yaml:"cloudsql_use_public_ip"`

	// CloudSQLAutoIAMAuthN is similar to the -enable_iam_login or -auto-iam-authn
	// argument in Cloud SQL Auth Proxy.
	//
	// Currently, Application Default Credentials is used to authenticate by default.
	// No other modes are supported.
	//
	// Should be true for non-local environment.
	CloudSQLAutoIAMAuthN bool `yaml:"cloudsql_auto_iam_authn"`

	// CloudSQLImpersonateServiceAccountEmail, when specified, is the service account email
	// to be used to connect to Cloud SQL and perform Cloud IAM Authentication.
	//
	// It is not to be confused with User field, which is the PostgreSQL user account
	// to run SQL statements (although they usually refers to the same account).
	//
	// Currently, Application Default Credentials is used to authenticate by default.
	// No other modes are supported.
	//
	// Example: stag-bob@staging-manabie-online.iam.gserviceaccount.com
	// (in that case, the user field is usually "stag-bob@staging-manabie-online.iam").
	CloudSQLImpersonateServiceAccountEmail string `yaml:"cloudsql_impersonate_service_account_email"`

	// These are the usual PostgreSQL connection parameters.
	// User and DBName are required, while the rest are only required
	// when not connecting to a Cloud SQL instance.
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	DBName   string `yaml:"dbname"`

	// Some additional parameters for PostgreSQL connection.
	MaxConns          int32         `yaml:"max_conns"`
	RetryAttempts     int           `yaml:"retry_attempts"`
	RetryWaitInterval time.Duration `yaml:"retry_wait_interval"`
	MaxConnIdleTime   time.Duration `yaml:"max_conn_idle_time"`

	// ShardID is the idenfier of the current database shard.
	// It is a pointer to distinguish zero vs unspecified.
	ShardID *int `yaml:"shard_id"`
}

// mergeFrom merges other into c and return the resulting struct.
// Values from other override values in c.
func (c PostgresDatabaseConfig) merge(other PostgresDatabaseConfig) *PostgresDatabaseConfig {
	v1 := reflect.ValueOf(c)
	v2 := reflect.ValueOf(other)
	out := reflect.New(reflect.TypeOf(c))
	for i, n := 0, v2.NumField(); i < n; i++ {
		f := v2.Field(i)
		if f.IsZero() {
			f = v1.Field(i)
		}
		out.Elem().Field(i).Set(f)
	}
	return out.Interface().(*PostgresDatabaseConfig)
}

// PGXConnectionString returns the database connection string.
// See https://www.postgresql.org/docs/current/libpq-connect.html.
//
// The returned connection string should only be used with pgx interface.
// To use with database/sql interface, use ConnectionString instead.
func (c PostgresDatabaseConfig) PGXConnectionString() (string, error) {
	if err := c.validate(); err != nil {
		return "", fmt.Errorf("postgres config is invalid: %s", err)
	}

	if c.IsCloudSQL() {
		return fmt.Sprintf("user='%s' password='%s' dbname='%s' sslmode='disable' application_name='%s'",
			c.User, c.Password, c.DBName, c.User,
		), nil
	}
	return fmt.Sprintf(
		"user='%s' password='%s' host='%s' port='%s' dbname='%s' sslmode='disable' application_name='%s'",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.User,
	), nil
}

// ConnectionString returns the database connection string.
// See https://www.postgresql.org/docs/current/libpq-connect.html.
//
// The returned connection string should be only used with database/sql interface.
// To use with pgx interface, use PGXConnectionString instead.
func (c *PostgresDatabaseConfig) ConnectionString() (string, error) {
	if err := c.validate(); err != nil {
		return "", fmt.Errorf("postgres config is invalid: %s", err)
	}
	if c.IsCloudSQL() {
		// We set c.CloudSQLInstance as host here, following the example
		// at https://github.com/googlecloudplatform/cloud-sql-go-connector#using-the-dialer-with-databasesql
		return fmt.Sprintf("user='%s' password='%s' host='%s' dbname='%s' sslmode='disable' application_name='%s'",
			c.User, c.Password, c.CloudSQLInstance, c.DBName, c.User,
		), nil
	}
	return fmt.Sprintf(
		"user='%s' password='%s' host='%s' port='%s' dbname='%s' sslmode='disable' application_name='%s'",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.User,
	), nil
}

func (c *PostgresDatabaseConfig) validate() error {
	if c.User == "" {
		return errors.New(`"user" cannot be empty`)
	}
	if c.DBName == "" {
		return errors.New(`"dbname" cannot be empty`)
	}
	if !c.IsCloudSQL() {
		if c.Host == "" {
			return errors.New(`"host" cannot be empty when not connecting to Cloud SQL`)
		}
		if c.Port == "" {
			return errors.New(`"port" cannot be empty when not connecting to Cloud SQL`)
		}
	}
	return nil
}

// IsCloudSQL reports whether this config is intended for a Google's Cloud SQL.
func (c *PostgresDatabaseConfig) IsCloudSQL() bool {
	return c.CloudSQLInstance != ""
}

// DefaultCloudSQLConnOpts returns the commonly used options to create cloudsqlconn.Dialer.
func (c *PostgresDatabaseConfig) DefaultCloudSQLConnOpts(ctx context.Context) ([]cloudsqlconn.Option, error) {
	opts := []cloudsqlconn.Option{
		cloudsqlconn.WithUserAgent(c.User),
	}

	cloudSQLDialOpts := []cloudsqlconn.DialOption{}
	if c.CloudSQLUsePublicIP {
		cloudSQLDialOpts = append(cloudSQLDialOpts, cloudsqlconn.WithPublicIP())
	} else {
		cloudSQLDialOpts = append(cloudSQLDialOpts, cloudsqlconn.WithPrivateIP())
	}
	opts = append(opts, cloudsqlconn.WithDefaultDialOptions(cloudSQLDialOpts...))

	credOpts, err := c.credentialsOpt(ctx)
	if err != nil {
		return nil, fmt.Errorf("c.credentialsOpt: %s", err)
	}
	opts = append(opts, credOpts)

	if c.CloudSQLAutoIAMAuthN {
		opts = append(opts, cloudsqlconn.WithIAMAuthN())
	}

	return opts, nil
}

const iamLoginScope = "https://www.googleapis.com/auth/sqlservice.login"

// credentialsOpt returns the necessary token source for cloudsqlconn.Dialer.
// Reference: https://github.com/GoogleCloudPlatform/cloud-sql-proxy/blob/d022c5683a301722e55692ae3ca1d62cf0e6d017/internal/proxy/proxy.go#L251
func (c PostgresDatabaseConfig) credentialsOpt(ctx context.Context) (cloudsqlconn.Option, error) {
	impersonateTarget := c.CloudSQLImpersonateServiceAccountEmail
	if impersonateTarget != "" {
		log.Printf("impersonating %q", impersonateTarget)
		apiTS, err := impersonate.CredentialsTokenSource(
			ctx,
			impersonate.CredentialsConfig{
				TargetPrincipal: impersonateTarget,
				Scopes:          []string{sqladmin.SqlserviceAdminScope},
			},
		)
		if err != nil {
			return nil, err
		}

		if c.CloudSQLAutoIAMAuthN {
			iamLoginTS, err := impersonate.CredentialsTokenSource(
				ctx,
				impersonate.CredentialsConfig{
					TargetPrincipal: impersonateTarget,
					Scopes:          []string{iamLoginScope},
				},
			)
			if err != nil {
				return nil, err
			}
			return cloudsqlconn.WithIAMAuthNTokenSources(apiTS, iamLoginTS), nil
		}
		return cloudsqlconn.WithTokenSource(apiTS), nil
	}

	// if not impersonating, return a no-op option
	return cloudsqlconn.WithOptions(), nil
}
