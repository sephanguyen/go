package unleash

import (
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
)

const (
	DisableAutoDeactivateStudents                      = "User_StudentManagement_DisableDeactivateStudent"
	ExperimentalBulkInsertEnrollmentStatusHistories    = "User_StudentManagement_Experimental_BulkInsert_Enrollment_Status_Histories"
	FeatureDecouplingUserAndAuthDB                     = "User_UserManagement_DecouplingUserAndAuthDB"
	FeatureIgnoreInvalidRecordsOpenAPI                 = "User_StudentManagement_IgnoreInvalidRecordsOpenAPI"
	FeatureToggleAutoDeactivateAndReactivateStudentsV2 = "User_StudentManagement_DeactivateStudent_V2"
	FeatureToggleStaffUsername                         = "User_StaffManagement_StaffUsername"
	FeatureToggleUserNameStudentParent                 = "User_StudentManagement_StudentParentUsername"
	FeatureUsingMasterReplicatedTable                  = "User_StudentManagement_UsingMasterReplicatedTable"
)

type DomainUserFeatureOption struct {
	EnableIgnoreUpdateEmail bool
	EnableUsername          bool
}

type DomainStudentFeatureOption struct {
	DomainUserFeatureOption
	EnableAutoDeactivateAndReactivateStudentV2            bool
	DisableAutoDeactivateAndReactivateStudent             bool
	EnableExperimentalBulkInsertEnrollmentStatusHistories bool
}

type DomainParentFeatureOption struct {
	DomainUserFeatureOption
}

func IsFeatureUsingMasterReplicatedTable(unleashClient unleashclient.ClientInstance, env string, organization valueobj.HasOrganizationID) bool {
	return isFeatureToggleEnabled(unleashClient, env, organization, FeatureUsingMasterReplicatedTable)
}

func IsFeatureIgnoreInvalidRecordsOpenAPI(unleashClient unleashclient.ClientInstance, env string, organization valueobj.HasOrganizationID) bool {
	return isFeatureToggleEnabled(unleashClient, env, organization, FeatureIgnoreInvalidRecordsOpenAPI)
}

func IsFeatureUserNameStudentParentEnabled(unleashClient unleashclient.ClientInstance, env string, organization valueobj.HasOrganizationID) bool {
	return isFeatureToggleEnabled(unleashClient, env, organization, FeatureToggleUserNameStudentParent)
}

func IsFeatureStaffUsernameEnabled(unleashClient unleashclient.ClientInstance, env string, organization valueobj.HasOrganizationID) bool {
	return isFeatureToggleEnabled(unleashClient, env, organization, FeatureToggleStaffUsername)
}

func IsExperimentalBulkInsertEnrollmentStatusHistories(unleashClient unleashclient.ClientInstance, env string, organization valueobj.HasOrganizationID) bool {
	return isFeatureToggleEnabled(unleashClient, env, organization, ExperimentalBulkInsertEnrollmentStatusHistories)
}

func IsFeatureAutoDeactivateAndReactivateStudentsV2Enabled(unleashClient unleashclient.ClientInstance, env string, organization valueobj.HasOrganizationID) bool {
	return isFeatureToggleEnabled(unleashClient, env, organization, FeatureToggleAutoDeactivateAndReactivateStudentsV2)
}

func IsDisableAutoDeactivateStudents(unleashClient unleashclient.ClientInstance, env string, organization valueobj.HasOrganizationID) bool {
	return isFeatureToggleEnabled(unleashClient, env, organization, DisableAutoDeactivateStudents)
}

func IsFeatureDecouplingUserAndAuthDBEnable(unleashClient unleashclient.ClientInstance, organization valueobj.HasOrganizationID, env string) bool {
	return isFeatureToggleEnabled(unleashClient, env, organization, FeatureDecouplingUserAndAuthDB)
}

func isFeatureToggleEnabled(unleashClient unleashclient.ClientInstance, env string, organization valueobj.HasOrganizationID, toggleName string) bool {
	isFeatureEnabled, err := unleashClient.IsFeatureEnabledOnOrganization(toggleName, env, organization.OrganizationID().String())
	if err != nil {
		isFeatureEnabled = false
	}
	return isFeatureEnabled
}
