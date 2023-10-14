package tests

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/execwrapper"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/require"
)

type Postgres struct {
	Connection string `yaml:"connection"`
}

type PostgresMigrate struct {
	Postgres Postgres `yaml:"postgres"`
}

func (PostgresMigrate) Path(p vr.P, e vr.E, s vr.S) string {
	return filepath.Join(
		execwrapper.RootDirectory(),
		fmt.Sprintf("deployments/helm/manabie-all-in-one/charts/%v/secrets/%v/%v/%v_migrate.secrets.encrypted.yaml", s, p, e, s),
	)
}

type postgresMigrateConfig struct {
	Database configs.PostgresDatabaseConfig `yaml:"database"`
}

// migrationConfig is used to configure database migration jobs.
type migrationConfig struct {
	PostgresMigrate postgresMigrateConfig `yaml:"postgres_migrate"`
}

func (migrationConfig) Path(p vr.P, e vr.E, s vr.S) string {
	return (PostgresMigrate{}).Path(p, e, s)
}

func runTestPostgresV2MigratePassword(_ vr.P, e vr.E, s vr.S) bool {
	switch s {
	case vr.ServiceBob,
		vr.ServiceCalendar,
		vr.ServiceDraft,
		vr.ServiceEnigma,
		vr.ServiceEureka,
		vr.ServiceEntryExitMgmt,
		vr.ServiceFatima,
		vr.ServiceInvoiceMgmt,
		vr.ServiceLessonMgmt,
		vr.ServiceNotificationMgmt,
		vr.ServicePayment,
		vr.ServiceShamir,
		vr.ServiceTimesheet,
		vr.ServiceTom,
		vr.ServiceVirtualClassroom,
		vr.ServiceUserMgmt,
		vr.ServiceZeus:
		return true
	default:
		return false
	}
}

func isServiceUsingIAMForMigration(e vr.E) bool {
	return e == vr.EnvStaging || e == vr.EnvUAT
}

func isPartnerMigrationDisabled(p vr.P) bool {
	return p == vr.PartnerSynersia
}

func TestPostgresV2MigratePassword(t *testing.T) {
	// Test common DB
	vr.Iter(t).SkipE(vr.EnvPreproduction).IterPE(func(t *testing.T, p vr.P, e vr.E) {
		slist := commonMigrationServiceList(p, e)
		if len(slist) == 0 {
			return
		}

		firstSvc := slist[0]
		truthCfg, err := configs.LoadAndDecrypt[migrationConfig](p, e, firstSvc)
		require.NoError(t, err)
		if isServiceUsingIAMForMigration(e) || isPartnerMigrationDisabled(p) {
			require.Empty(t, truthCfg.PostgresMigrate.Database.Password)
		} else {
			require.NotEmptyf(t, truthCfg.PostgresMigrate.Database.Password, "postgres password of service %q cannot be empty", firstSvc)
		}

		// load other services and compare the password with the first service's password
		for i := 1; i < len(slist); i++ {
			s := slist[i]
			if !runTestPostgresV2MigratePassword(p, e, s) {
				continue
			}
			t.Run(s.String(), func(t *testing.T) {
				cfg, err := configs.LoadAndDecrypt[migrationConfig](p, e, s)
				require.NoError(t, err)
				if isServiceUsingIAMForMigration(e) || isPartnerMigrationDisabled(p) {
					require.Empty(t, cfg.PostgresMigrate.Database.Password)
				} else {
					require.Equal(t, truthCfg.PostgresMigrate.Database.Password, cfg.PostgresMigrate.Database.Password)
				}
			})
		}
	})

	// Test LMS DB
	// Currently only eureka is located in lms, so the test is shorter
	vr.Iter(t).SkipE(vr.EnvPreproduction).IterPE(func(t *testing.T, p vr.P, e vr.E) {
		slist := lmsMigrationServiceList(p, e)
		if len(slist) == 0 {
			return
		}

		firstSvc := slist[0]
		if !runTestPostgresV2MigratePassword(p, e, firstSvc) {
			return
		}
		truthCfg, err := configs.LoadAndDecrypt[migrationConfig](p, e, firstSvc)
		require.NoError(t, err)
		if isServiceUsingIAMForMigration(e) || isPartnerMigrationDisabled(p) {
			require.Empty(t, truthCfg.PostgresMigrate.Database.Password)
		} else {
			require.NotEmptyf(t, truthCfg.PostgresMigrate.Database.Password, "postgres password of service %q cannot be empty", firstSvc)
		}
	})
}

func commonMigrationServiceList(p vr.P, e vr.E) []vr.S {
	res := []vr.S{
		vr.ServiceInvoiceMgmt,
		vr.ServiceZeus,
		vr.ServiceBob,
		vr.ServiceFatima,
		vr.ServiceLessonMgmt,
		vr.ServiceMasterMgmt,
		vr.ServiceTom,
		vr.ServiceTimesheet,
		vr.ServiceNotificationMgmt,
	}
	if p != vr.PartnerJPREP {
		res = append(res,
			vr.ServiceCalendar,
			vr.ServiceEntryExitMgmt,
		)
	}
	if p == vr.PartnerManabie && (e == vr.EnvLocal || e == vr.EnvStaging) {
		res = append(res, vr.ServiceDraft)
	}
	return res
}

func lmsMigrationServiceList(p vr.P, e vr.E) []vr.S {
	return []vr.S{
		vr.ServiceEureka,
	}
}
