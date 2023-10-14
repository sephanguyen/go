package database

import (
	"testing"

	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPGConnString(t *testing.T) {
	t.Parallel()

	type testcase struct {
		desc               string
		connstr            string
		iamLoginEnabled    bool
		expectedErr        error
		expectedDBUser     string
		expectedDBPassword string
		expectedDBHost     string
		expectedDBPort     string
		expectedDBName     string
	}

	testcases := []testcase{
		{
			desc:               "conn string with password",
			connstr:            `postgres://tokyo_hasura:aAbBcC123456@127.0.0.1:5432/tokyo_bob?sslmode=disable`,
			iamLoginEnabled:    false,
			expectedErr:        nil,
			expectedDBUser:     "tokyo_hasura",
			expectedDBPassword: "aAbBcC123456",
			expectedDBHost:     "127.0.0.1",
			expectedDBPort:     "5432",
			expectedDBName:     "tokyo_bob",
		},
		{
			desc:               "conn string with IAM login",
			connstr:            `postgres://prod-bob-h%40student-coach-e1e95.iam@127.0.0.1:5432/tokyo_bob?sslmode=disable`,
			iamLoginEnabled:    true,
			expectedErr:        nil,
			expectedDBUser:     "prod-bob-h%40student-coach-e1e95.iam",
			expectedDBPassword: "",
			expectedDBHost:     "127.0.0.1",
			expectedDBPort:     "5432",
			expectedDBName:     "tokyo_bob",
		},
		{
			desc:               "conn string with IAM login, with empty password",
			connstr:            `postgres://prod-bob-h%40student-coach-e1e95.iam:@127.0.0.1:5432/tokyo_bob?sslmode=disable`,
			iamLoginEnabled:    true,
			expectedErr:        nil,
			expectedDBUser:     "prod-bob-h%40student-coach-e1e95.iam",
			expectedDBPassword: "",
			expectedDBHost:     "127.0.0.1",
			expectedDBPort:     "5432",
			expectedDBName:     "tokyo_bob",
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			p, err := NewPGConnString(vr.PartnerTokyo, vr.EnvProduction, vr.ServiceBob, tc.connstr, EnableIAMLogin(tc.iamLoginEnabled))
			require.Equal(t, tc.expectedErr, err)
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedDBUser, p.dbUser)
				assert.Equal(t, tc.expectedDBPassword, p.dbPassword)
				assert.Equal(t, tc.expectedDBHost, p.dbHost)
				assert.Equal(t, tc.expectedDBPort, p.dbPort)
				assert.Equal(t, tc.expectedDBName, p.dbName)
			}
		})
	}
}

func TestPGConnString_AssertAll(t *testing.T) {
	t.Parallel()
	t.Run("uat.bob", func(t *testing.T) {
		t.Parallel()
		c, err := NewPGConnString(
			vr.PartnerManabie,
			vr.EnvUAT,
			vr.ServiceKafkaConnect,
			"postgres://uat_kafka_connector:example1@localhost:5432/uat_bob?sslmode=disable",
			RequireDBUser("kafka_connector", true),
			RequireDBPassword("example1"),
			RequireDBName("bob"),
		)
		require.NoError(t, err)
		c.AssertAll(t)
	})

	t.Run("stag.jprep", func(t *testing.T) {
		t.Parallel()
		c, err := NewPGConnString(
			vr.PartnerJPREP,
			vr.EnvStaging,
			vr.ServiceBob,
			"postgres://bob:example1@localhost:5432/stag_bob?sslmode=disable",
			RequireDBPassword("example1"),
		)
		require.NoError(t, err)
		c.AssertAll(t)
	})
}

func TestPGConnString_WrongDBUser(t *testing.T) {
	t.Parallel()
	c, err := NewPGConnString(
		vr.PartnerAIC,
		vr.EnvProduction,
		vr.ServiceKafkaConnect,
		"postgres://nsg_kafka_connector:example1@127.0.0.1:5432/aic_bob?sslmode=disable",
		RequireDBUser("kafka_connector", true),
		RequireDBPassword("example1"),
	)
	require.NoError(t, err)
	require.EqualError(t, c.ValidateDBUser(), `expected DB user to be one of [aic_kafka_connector], got "nsg_kafka_connector"`)
}

func TestPGConnString_WrongPassword(t *testing.T) {
	t.Parallel()
	c, err := NewPGConnString(
		vr.PartnerTokyo,
		vr.EnvProduction,
		vr.ServiceKafkaConnect,
		"postgres://nsg_kafka_connector:wrong_password@127.0.0.1:5432/aic_bob?sslmode=disable",
		RequireDBUser("kafka_connector", true),
		RequireDBPassword("example1"),
	)
	require.NoError(t, err)
	require.EqualError(t, c.ValidateDBPassword(), `expected DB password to be "example1", got "wrong_password"`)
}

func TestPGConnString_WrongDBHost(t *testing.T) {
	t.Parallel()
	c, err := NewPGConnString(
		vr.PartnerTokyo,
		vr.EnvProduction,
		vr.ServiceKafkaConnect,
		"postgres://kec_kafka_connector:example1@wronghost:5432/aic_bob?sslmode=disable",
		RequireDBUser("kafka_connector", true),
		RequireDBPassword("example1"),
	)
	require.NoError(t, err)
	require.EqualError(t, c.ValidateDBHost(), `expected DB host to be "localhost" or "127.0.0.1", got "wronghost"`)
}

func TestPGConnString_WrongDBPort(t *testing.T) {
	t.Parallel()
	c, err := NewPGConnString(
		vr.PartnerSynersia,
		vr.EnvProduction,
		vr.ServiceKafkaConnect,
		"postgres://kafka_connector:example1@localhost:23456/aic_bob?sslmode=disable",
		RequireDBUser("kafka_connector", true),
		RequireDBPassword("example1"),
	)
	require.NoError(t, err)
	require.EqualError(t, c.ValidateDBPort(), `expected DB port to be "5432", got "23456"`)
}

func TestPGConnString_WrongDBName(t *testing.T) {
	t.Parallel()
	c, err := NewPGConnString(
		vr.PartnerRenseikai,
		vr.EnvProduction,
		vr.ServiceKafkaConnect,
		"postgres://kafka_connector:example1@localhost:5432/eureka?sslmode=disable",
		RequireDBUser("kafka_connector", true),
		RequireDBPassword("example1"),
		RequireDBName("bob"),
	)
	require.NoError(t, err)
	require.EqualError(t, c.ValidateDBName(), `expected DB name to be "bob", got "eureka"`)
}
