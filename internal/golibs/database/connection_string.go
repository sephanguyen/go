package database

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/assert"
)

var (
	dbConnReStr        = `^postgres://([a-z_]+):(\w+)@([a-zA-Z0-9\.]+):(\d+)/([a-z_]+)\?sslmode=disable$`
	dbConnRe           = regexp.MustCompile(dbConnReStr)
	dbConnReWithIAMStr = `^postgres://([\w\-\%\.]+)(:\w*)?@([\w\d\.]+):(\d+)/(\w+)\?sslmode=disable$`
	dbConnReWithIAM    = regexp.MustCompile(dbConnReWithIAMStr)
	dbConnKakfaRe      = regexp.MustCompile(`^jdbc:postgresql://([a-zA-Z0-9\.]+):(\d+)/([a-z_]+)\?user=([a-z_]+)\&password=(\w+)\&stringtype=unspecified$`)
)

// isServiceUsingIAMLogin determines whether the service is using IAM login.
// The logic is hard-coded.
func isServiceUsingIAMLogin(p vr.P, e vr.E, s vr.S) bool {
	if p == vr.PartnerJPREP && e == vr.EnvStaging {
		// JPREP Staging does not use IAM login.
		// TODO: write a terraform module for it
		return false
	}

	noList := []vr.S{
		vr.ServiceKafkaConnect,
	}
	for _, x := range noList {
		if s == x {
			return false
		}
	}

	return true
}

type PGConnStringOption func(*PGConnString) error

// RequireDBUser asserts that the connection string uses the provided user `dbuser`
// with some environment-specific prefix if enabled.
// Using this option disables IAM login option.
// Not activated when EnableIAMLogin is specified.
func RequireDBUser(dbuser string, enablePrefix bool) PGConnStringOption {
	return func(pg *PGConnString) error {
		pg.dbUserOverride = dbuser
		pg.dbUserPrefixEnabled = enablePrefix
		return nil
	}
}

// RequireIAMDBUser is similar to RequireDBUser, but only activated
// when IAM login is enabled (possibly with EnableIAMLogin).
func RequireIAMDBUser(dbuser string) PGConnStringOption {
	return func(pg *PGConnString) error {
		pg.dbIAMUserOverride = dbuser
		return nil
	}
}

func RequireDBPassword(password string) PGConnStringOption {
	return func(pg *PGConnString) error {
		pg.dbPasswordOverride = password
		return nil
	}
}

func RequireDBName(dbname string) PGConnStringOption {
	return func(pg *PGConnString) error {
		pg.dbNameOverride = dbname
		return nil
	}
}

// ForcePassword makes the PGConnString assume that the connection is using a password,
// so that it will not try to use IAM login.
func ForcePassword(b bool) PGConnStringOption {
	return func(pg *PGConnString) error {
		pg.forcePassword = b
		return nil
	}
}

func EnableIAMLogin(b bool) PGConnStringOption {
	return func(ps *PGConnString) error {
		ps.iamLoginEnabled = b
		return nil
	}
}

// PGConnString is a struct that represents a PostgreSQL connection string.
// This struct should be used to assert the connection string.
type PGConnString struct {
	partner     vr.P
	environment vr.E
	service     vr.S
	connString  string
	dbUser      string
	dbPassword  string
	dbHost      string
	dbPort      string
	dbName      string

	iamLoginEnabled bool

	// The following fields are used to override the values extracted from the connection string.
	// They are managed by the opts ...PGConnStringOption provided.
	dbUserOverride      string
	dbIAMUserOverride   string
	dbUserPrefixEnabled bool
	dbPasswordOverride  string
	dbNameOverride      string
	forcePassword       bool
}

// NewPGConnString returns a new PGConnString instance.
func NewPGConnString(p vr.P, e vr.E, s vr.S, connstr string, opts ...PGConnStringOption) (*PGConnString, error) {
	pgConnStr := &PGConnString{
		partner:             p,
		environment:         e,
		service:             s,
		connString:          connstr,
		dbUserPrefixEnabled: true,
	}
	for _, opt := range opts {
		if err := opt(pgConnStr); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	var matches []string
	if pgConnStr.iamLoginEnabled {
		matches = dbConnReWithIAM.FindStringSubmatch(connstr)
		if len(matches) != 6 {
			return nil, fmt.Errorf("connection string %q failed to match regexp %q", connstr, dbConnReWithIAMStr)
		}
		pgConnStr.dbUser = matches[1]
		pgConnStr.dbPassword = strings.TrimPrefix(matches[2], ":")
		pgConnStr.dbHost = matches[3]
		pgConnStr.dbPort = matches[4]
		pgConnStr.dbName = matches[5]
	} else if pgConnStr.forcePassword || !isServiceUsingIAMLogin(p, e, s) {
		matches = dbConnRe.FindStringSubmatch(connstr)
		if len(matches) != 6 {
			return nil, fmt.Errorf("connection string %q failed to match regexp %q", connstr, dbConnReStr)
		}
		pgConnStr.dbUser = matches[1]
		pgConnStr.dbPassword = matches[2]
		pgConnStr.dbHost = matches[3]
		pgConnStr.dbPort = matches[4]
		pgConnStr.dbName = matches[5]
	} else {
		matches = dbConnReWithIAM.FindStringSubmatch(connstr)
		if len(matches) != 6 {
			return nil, fmt.Errorf("connection string %q failed to match regexp %q", connstr, dbConnReWithIAMStr)
		}
		pgConnStr.dbUser = matches[1]
		pgConnStr.dbPassword = strings.TrimPrefix(matches[2], ":")
		pgConnStr.dbHost = matches[3]
		pgConnStr.dbPort = matches[4]
		pgConnStr.dbName = matches[5]
	}

	return pgConnStr, nil
}

func NewPGConnStringForKakfaProperties(p vr.P, e vr.E, connStr string, opts ...PGConnStringOption) (*PGConnString, error) {
	matches := dbConnKakfaRe.FindStringSubmatch(connStr)
	if len(matches) != 6 {
		return nil, fmt.Errorf("failed to parse connection string: expected 6 elements after parsing: 1 fullmatch plus 5 submatches, got %d", len(matches))
	}
	pgConnStr := &PGConnString{
		partner:     p,
		environment: e,
		service:     vr.ServiceKafkaConnect,
		connString:  connStr,
		dbUser:      matches[4],
		dbPassword:  matches[5],
		dbHost:      matches[1],
		dbPort:      matches[2],
		dbName:      matches[3],
	}

	for _, opt := range opts {
		if err := opt(pgConnStr); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	return pgConnStr, nil
}

// DBPassword returns the password extracted from the connection string.
func (pg *PGConnString) DBPassword() string {
	return pg.dbPassword
}

// AssertAll is a convenient function that asserts all the fields of the connection string.
func (pg *PGConnString) AssertAll(t *testing.T) {
	assert.NoError(t, pg.ValidateDBUser())
	assert.NoError(t, pg.ValidateDBPassword())
	assert.NoError(t, pg.ValidateDBHost())
	assert.NoError(t, pg.ValidateDBPort())
	assert.NoError(t, pg.ValidateDBName())
}

// ValidateDBUser returns an error if the database user is incorrect.
func (pg *PGConnString) ValidateDBUser() error {
	expectedDBUsers := pg.getExpectedDBUser()
	for _, allowedDBUser := range expectedDBUsers {
		if pg.dbUser == allowedDBUser {
			return nil
		}
	}
	return fmt.Errorf("expected DB user to be one of %v, got %q", expectedDBUsers, pg.dbUser)
}

// ValidateDBPassword returns an error if the database password is incorrect.
func (pg *PGConnString) ValidateDBPassword() error {
	if pg.isIAMLoginEnabled() {
		if pg.dbPassword != "" {
			return fmt.Errorf("expected DB password to be empty, got %q", pg.dbPassword)
		}
	} else {
		if pg.dbPasswordOverride != "" {
			if pg.dbPassword != pg.dbPasswordOverride {
				return fmt.Errorf("expected DB password to be %q, got %q", pg.dbPasswordOverride, pg.dbPassword)
			}
		} else {
			if pg.dbPassword == "" {
				return fmt.Errorf("expected DB password to be non-empty, got empty")
			}
		}
	}
	return nil
}

// ValidateDBHost returns an error if the database host is incorrect.
func (pg *PGConnString) ValidateDBHost() error {
	if pg.dbHost != "127.0.0.1" && pg.dbHost != "localhost" {
		return fmt.Errorf(`expected DB host to be "localhost" or "127.0.0.1", got %q`, pg.dbHost)
	}
	return nil
}

// ValidateDBPort returns an error if the database port is incorrect.
func (pg *PGConnString) ValidateDBPort() error {
	if pg.dbPort != "5432" {
		return fmt.Errorf(`expected DB port to be "5432", got %q`, pg.dbPort)
	}
	return nil
}

// ValidateDBName returns an error if the database name is incorrect.
func (pg *PGConnString) ValidateDBName() error {
	expectedDBName := pg.getExpectedDBName()
	if pg.dbName != expectedDBName {
		return fmt.Errorf("expected DB name to be %q, got %q", expectedDBName, pg.dbName)
	}
	return nil
}

func (pg *PGConnString) isIAMLoginEnabled() bool {
	return (!pg.forcePassword && isServiceUsingIAMLogin(pg.partner, pg.environment, pg.service)) || pg.iamLoginEnabled
}

func (pg *PGConnString) getExpectedDBUser() []string {
	if pg.isIAMLoginEnabled() {
		if pg.dbIAMUserOverride != "" {
			return []string{
				fmt.Sprintf("%v@%v.iam", pg.dbIAMUserOverride, vr.GCPProjectID(pg.partner, pg.environment)),
				fmt.Sprintf("%v%%40%v.iam", pg.dbIAMUserOverride, vr.GCPProjectID(pg.partner, pg.environment)),
			}
		}
		return []string{
			fmt.Sprintf("%v-%v@%v.iam", pg.environment, pg.service, vr.GCPProjectID(pg.partner, pg.environment)),
			fmt.Sprintf("%v-%v%%40%v.iam", pg.environment, pg.service, vr.GCPProjectID(pg.partner, pg.environment)),
		}
	}

	name := pg.service.String()
	if pg.dbUserOverride != "" {
		name = pg.dbUserOverride
	}
	if !pg.dbUserPrefixEnabled {
		return []string{name}
	}
	return []string{fmt.Sprintf("%s%s", pg.customDBUserPrefix(pg.partner, pg.environment), name)}
}

func (pg *PGConnString) customDBUserPrefix(p vr.P, e vr.E) string {
	switch p {
	case vr.PartnerAIC, vr.PartnerGA, vr.PartnerTokyo:
		if e == vr.EnvProduction {
			return p.String() + "_"
		}
	case vr.PartnerJPREP:
		switch e {
		case vr.EnvStaging, vr.EnvUAT, vr.EnvProduction:
			return ""
		}
	case vr.PartnerManabie:
		switch e {
		case vr.EnvStaging, vr.EnvProduction:
			return ""
		case vr.EnvUAT:
			return "uat_"
		}
	case vr.PartnerRenseikai:
		if e == vr.EnvProduction {
			return ""
		}
	case vr.PartnerSynersia:
		if e == vr.EnvProduction {
			return vr.PartnerTokyo.String() + "_"
		}
	}
	panic(fmt.Sprintf("invalid partner-environment combination: %v-%v", p, e))
}

func (pg *PGConnString) getExpectedDBName() string {
	dbNamePrefix := pg.customDBNamePrefix(pg.partner, pg.environment)

	if pg.dbNameOverride != "" {
		return dbNamePrefix + pg.dbNameOverride
	}

	// TODO: clean this switch case up
	var dbName string
	switch pg.service {
	case vr.ServiceShamir, vr.ServiceYasuo, vr.ServiceUserMgmt,
		vr.ServiceMasterMgmt, vr.ServiceLessonMgmt,
		vr.ServiceEnigma, vr.ServiceEntryExitMgmt:
		dbName = vr.ServiceBob.String()
	case vr.ServicePayment:
		dbName = vr.ServiceFatima.String()
	default:
		dbName = pg.service.String()
	}

	return dbNamePrefix + dbName
}

func (pg *PGConnString) customDBNamePrefix(p vr.P, e vr.E) string {
	switch p {
	case vr.PartnerGA, vr.PartnerAIC, vr.PartnerTokyo:
		if e == vr.EnvProduction {
			return p.String() + "_"
		}
	case vr.PartnerJPREP:
		switch e {
		case vr.EnvStaging:
			return "stag_"
		case vr.EnvUAT, vr.EnvProduction:
			return ""
		}
	case vr.PartnerManabie:
		switch e {
		case vr.EnvStaging, vr.EnvProduction:
			return ""
		case vr.EnvUAT:
			return "uat_"
		}
	case vr.PartnerRenseikai:
		if e == vr.EnvProduction {
			return ""
		}
	case vr.PartnerSynersia:
		if e == vr.EnvProduction {
			return vr.PartnerTokyo.String() + "_"
		}
	}
	panic(fmt.Sprintf("invalid partner-environment combination: %s-%s", p, e))
}
