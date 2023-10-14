// Package features define some features for a domain user.
package features

import (
	"context"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4"
)

// FeatureManager is a struct that represents a set of features for a domain user in the source code.
// This struct has two main components: feature flags and config.
//
// Feature flags are used to control the development flow of the project using Unleash, a feature management system.
// If there is any problem or incident during the development,
// the feature flags can be toggled to roll back the feature.
//
// Config is used to enable or disable the features for the domain user according to their needs and preferences
type FeatureManager struct {
	UnleashClient unleashclient.ClientInstance
	Env           string
	DB            database.QueryExecer

	InternalConfigurationRepo interface {
		GetByKey(ctx context.Context, db database.QueryExecer, configKey string) (entity.DomainConfiguration, error)
	}
}

func (d *FeatureManager) isEnableUsernameConfig(ctx context.Context, db database.QueryExecer) bool {
	zapLogger := ctxzap.Extract(ctx).Sugar()

	config, err := d.InternalConfigurationRepo.GetByKey(ctx, db, constant.KeyAuthUsernameConfig)
	if err != nil {
		if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
			return false
		}
		zapLogger.Errorw("failed to get configuration", "err", err)
		return false
	}
	return config.ConfigValue().String() == constant.ConfigValueOn
}

func (d *FeatureManager) IsEnableUsername(ctx context.Context, org valueobj.HasOrganizationID) bool {
	isEnableUsernameUnleash := unleash.IsFeatureUserNameStudentParentEnabled(d.UnleashClient, d.Env, org)
	return isEnableUsernameUnleash && d.isEnableUsernameConfig(ctx, d.DB)
}

func (d *FeatureManager) IsEnableUsernameStudentParentStaff(ctx context.Context, org valueobj.HasOrganizationID) bool {
	usernameStudentParent := d.IsEnableUsername(ctx, org)
	usernameStaff := unleash.IsFeatureStaffUsernameEnabled(d.UnleashClient, d.Env, org)
	return usernameStudentParent && usernameStaff
}

func (d *FeatureManager) IsEnableDecouplingUserAndAuthDB(org valueobj.HasOrganizationID) bool {
	return unleash.IsFeatureDecouplingUserAndAuthDBEnable(d.UnleashClient, org, d.Env)
}

func (d *FeatureManager) FeatureUsernameToStudentFeatureOption(ctx context.Context, org valueobj.HasOrganizationID, option unleash.DomainStudentFeatureOption) unleash.DomainStudentFeatureOption {
	newOption := option
	newOption.EnableUsername = d.IsEnableUsername(ctx, org)
	return newOption
}

func (d *FeatureManager) FeatureUsernameToParentFeatureOption(ctx context.Context, org valueobj.HasOrganizationID, option unleash.DomainParentFeatureOption) unleash.DomainParentFeatureOption {
	newOption := option
	newOption.EnableUsername = d.IsEnableUsername(ctx, org)
	return newOption
}

func (d *FeatureManager) FeatureAutoDeactivateAndReactivateStudentsV2ToStudentFeatureOption(_ context.Context, org valueobj.HasOrganizationID, option unleash.DomainStudentFeatureOption) unleash.DomainStudentFeatureOption {
	newOption := option
	// check config if existing
	newOption.EnableAutoDeactivateAndReactivateStudentV2 = unleash.IsFeatureAutoDeactivateAndReactivateStudentsV2Enabled(d.UnleashClient, d.Env, org)
	return newOption
}

func (d *FeatureManager) FeatureDisableAutoDeactivateStudentsToStudentFeatureOption(_ context.Context, org valueobj.HasOrganizationID, option unleash.DomainStudentFeatureOption) unleash.DomainStudentFeatureOption {
	newOption := option
	// check config if existing
	newOption.DisableAutoDeactivateAndReactivateStudent = unleash.IsDisableAutoDeactivateStudents(d.UnleashClient, d.Env, org)
	return newOption
}

func (d *FeatureManager) FeatureExperimentalBulkInsertEnrollmentStatusHistoriesToStudentFeatureOption(_ context.Context, org valueobj.HasOrganizationID, option unleash.DomainStudentFeatureOption) unleash.DomainStudentFeatureOption {
	newOption := option
	// check config if existing
	newOption.EnableExperimentalBulkInsertEnrollmentStatusHistories = unleash.IsExperimentalBulkInsertEnrollmentStatusHistories(d.UnleashClient, d.Env, org)
	return newOption
}
