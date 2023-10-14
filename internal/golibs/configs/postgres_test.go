package configs

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestPostgresDatabaseConfigPGXConnectionString(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		desc                  string
		cloudsqlinstance      string
		cloudsqlimpersonation string
		user                  string
		password              string
		host                  string
		port                  string
		dbname                string
		expectedConnStr       string
	}{
		{
			desc:             "basic cloud sql usage",
			cloudsqlinstance: "staging:asia:vietnam",
			user:             "dev-a@example.com",
			dbname:           "mydb",
			expectedConnStr:  "user='dev-a@example.com' password='' dbname='mydb' sslmode='disable' application_name='dev-a@example.com'",
		},
		{
			desc:            "basic local usage",
			user:            "dev-a@example.com",
			password:        "randompassword",
			host:            "localhost",
			port:            "5432",
			dbname:          "mydb",
			expectedConnStr: "user='dev-a@example.com' password='randompassword' host='localhost' port='5432' dbname='mydb' sslmode='disable' application_name='dev-a@example.com'",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			c := PostgresDatabaseConfig{
				CloudSQLInstance:                       tc.cloudsqlinstance,
				CloudSQLImpersonateServiceAccountEmail: tc.cloudsqlimpersonation,
				User:                                   tc.user,
				Password:                               tc.password,
				Host:                                   tc.host,
				Port:                                   tc.port,
				DBName:                                 tc.dbname,
			}
			connstr, err := c.PGXConnectionString()
			require.NoError(t, err)
			require.Equal(t, tc.expectedConnStr, connstr)
		})
	}
}

func TestPostgresV2CustomUnmarshal(t *testing.T) {
	dataA := []byte(`
bob:
  cloudsql_instance: instanceA1
  user: userA
eureka:
  cloudsql_instance: instanceA2
  user: userA
`)

	dataB := []byte(`
bob:
  cloudsql_instance: instanceB1
  user: userB
tom:
  cloudsql_instance: instanceB3
  user: userB
`)

	actual := &PostgresConfigV2{}
	err := yaml.Unmarshal(dataA, actual)
	require.NoError(t, err)
	err = yaml.Unmarshal(dataB, actual)
	require.NoError(t, err)
	expected := &PostgresConfigV2{
		Databases: map[string]PostgresDatabaseConfig{
			"bob":    {CloudSQLInstance: "instanceB1", User: "userB"},
			"eureka": {CloudSQLInstance: "instanceA2", User: "userA"},
			"tom":    {CloudSQLInstance: "instanceB3", User: "userB"},
		},
	}
	require.Equal(t, expected, actual)
}

func TestPostgresDatabaseConfigMerge(t *testing.T) {
	intptrfunc := func(i int) *int { return &i }
	a := PostgresDatabaseConfig{
		CloudSQLInstance:                       "a",
		CloudSQLUsePublicIP:                    false,
		CloudSQLAutoIAMAuthN:                   false,
		CloudSQLImpersonateServiceAccountEmail: "a",
		User:                                   "a",
		Password:                               "a",
		Host:                                   "a",
		Port:                                   "a",
		DBName:                                 "a",
		MaxConns:                               0,
		RetryAttempts:                          1,
		RetryWaitInterval:                      1,
		MaxConnIdleTime:                        1,
		ShardID:                                intptrfunc(1),
	}
	b := PostgresDatabaseConfig{
		CloudSQLInstance:                       "",
		CloudSQLUsePublicIP:                    false,
		CloudSQLAutoIAMAuthN:                   true,
		CloudSQLImpersonateServiceAccountEmail: "",
		User:                                   "",
		Password:                               "b",
		Host:                                   "b",
		Port:                                   "b",
		DBName:                                 "b",
		MaxConns:                               0,
		RetryAttempts:                          0,
		RetryWaitInterval:                      2,
		MaxConnIdleTime:                        2,
		ShardID:                                intptrfunc(2),
	}
	expected := &PostgresDatabaseConfig{
		CloudSQLInstance:                       "a",
		CloudSQLUsePublicIP:                    false,
		CloudSQLAutoIAMAuthN:                   true,
		CloudSQLImpersonateServiceAccountEmail: "a",
		User:                                   "a",
		Password:                               "b",
		Host:                                   "b",
		Port:                                   "b",
		DBName:                                 "b",
		MaxConns:                               0,
		RetryAttempts:                          1,
		RetryWaitInterval:                      2,
		MaxConnIdleTime:                        2,
		ShardID:                                intptrfunc(2),
	}
	require.NotPanics(t, func() {
		actual := a.merge(b)
		require.Equal(t, expected, actual)
	})
}
