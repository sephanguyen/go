package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/golibs/tools"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/features"
	spb "github.com/manabie-com/backend/pkg/manabuf/spike/v1"

	"github.com/spf13/cobra"
)

func genUsermgmtRepo(_ *cobra.Command, args []string) error {
	tools.AddImport("github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity")
	repos := map[string]interface{}{
		"user":                              &repository.UserRepo{},
		"teacher":                           &repository.TeacherRepo{},
		"school_admin":                      &repository.SchoolAdminRepo{},
		"staff":                             &repository.StaffRepo{},
		"user_group":                        &repository.UserGroupRepo{},
		"student":                           &repository.StudentRepo{},
		"parent":                            &repository.ParentRepo{},
		"student_parent":                    &repository.StudentParentRepo{},
		"school_info":                       &repository.SchoolInfoRepo{},
		"user_access_path":                  &repository.UserAccessPathRepo{},
		"student_comment":                   &repository.StudentCommentRepo{},
		"granted_role":                      &repository.GrantedRoleRepo{},
		"role":                              &repository.RoleRepo{},
		"user_group_v2":                     &repository.UserGroupV2Repo{},
		"granted_role_access_path":          &repository.GrantedRoleAccessPathRepo{},
		"usr_email":                         &repository.UsrEmailRepo{},
		"user_group_member":                 &repository.UserGroupsMemberRepo{},
		"import_user_event":                 &repository.ImportUserEventRepo{},
		"permission":                        &repository.PermissionRepo{},
		"school_history":                    &repository.SchoolHistoryRepo{},
		"school_repo":                       &repository.SchoolRepo{},
		"school_course":                     &repository.SchoolCourseRepo{},
		"user_address":                      &repository.UserAddressRepo{},
		"prefecture":                        &repository.PrefectureRepo{},
		"user_phone_number":                 &repository.UserPhoneNumberRepo{},
		"organization":                      &repository.OrganizationRepo{},
		"domain_grade":                      &repository.DomainGradeRepo{},
		"grade_organization":                &repository.GradeOrganizationRepo{},
		"domain_student":                    &repository.DomainStudentRepo{},
		"domain_user_group":                 &repository.DomainUserGroupRepo{},
		"domain_user":                       &repository.DomainUserRepo{},
		"domain_api_keypair":                &repository.DomainAPIKeypairRepo{},
		"domain_tagged_user":                &repository.DomainTaggedUserRepo{},
		"domain_tag":                        &repository.DomainTagRepo{},
		"domain_user_address":               &repository.DomainUserAddressRepo{},
		"domain_school_history":             &repository.DomainSchoolHistoryRepo{},
		"domain_location":                   &repository.DomainLocationRepo{},
		"domain_school":                     &repository.DomainSchoolRepo{},
		"domain_school_course":              &repository.DomainSchoolCourseRepo{},
		"domain_prefecture":                 &repository.DomainPrefectureRepo{},
		"domain_usr_email":                  &repository.DomainUsrEmailRepo{},
		"domain_user_access_path":           &repository.DomainUserAccessPathRepo{},
		"domain_user_group_member":          &repository.DomainUserGroupMemberRepo{},
		"domain_enrollment_status_history":  &repository.DomainEnrollmentStatusHistoryRepo{},
		"student_enrollment_status_history": &repository.StudentEnrollmentStatusHistoryRepo{},
		"domain_parent":                     &repository.DomainParentRepo{},
		"domain_student_package":            &repository.DomainStudentPackageRepo{},
		"domain_student_parent":             &repository.DomainStudentParentRelationshipRepo{},
		"domain_internal_configuration":     &repository.DomainInternalConfigurationRepo{},
		"domain_course":                     &repository.DomainCourseRepo{},
		"domain_role":                       &repository.DomainRoleRepo{},
	}

	tools.MockRepository("mock_repositories", filepath.Join(args[0], "repositories"), "usermgmt", repos)

	mockObjects := map[string][]interface{}{
		"internal/usermgmt/service":           {&service.DomainParent{}},
		"internal/usermgmt/config":            {&features.FeatureManager{}},
		"internal/usermgmt/external_services": {spb.NewEmailModifierServiceClient(nil)},
	}
	return tools.GenMockStructs(mockObjects)
}

func newGenUsermgmtCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "usermgmt [../../mock/usermgmt]",
		Short: "generate usermgmt repository type",
		Args:  cobra.ExactArgs(1),
		RunE:  genUsermgmtRepo,
	}
}
