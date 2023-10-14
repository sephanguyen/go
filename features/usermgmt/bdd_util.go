package usermgmt

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	unleash "github.com/manabie-com/backend/features/unleash"
	user_unleash "github.com/manabie-com/backend/features/usermgmt/unleash"
	"github.com/manabie-com/backend/internal/golibs"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

const (
	CommonDateLayout         = "2006/01/02"
	ManabiePartnerInternalID = "manabie-location"
	invalidToken             = "an-invalid-dude-token"

	studentType         = "student"
	teacherType         = "teacher"
	parentType          = "parent"
	schoolAdminType     = "school admin"
	organizationType    = "organization manager"
	unauthenticatedType = "unauthenticated"

	JPREPSchool     = "JPREPSchool"
	SynersiaSchool  = "SynersiaSchool"
	RenseikaiSchool = "RenseikaiSchool"
	TestingSchool   = "TestingSchool"
	GASchool        = "GASchool"
	KECSchool       = "KECSchool"
	AICSchool       = "AICSchool"
	NSGSchool       = "NSGSchool"
)

var SchoolNameWithResourcePath = map[string]int{
	"Manabie School":    -2147483648,
	"JPREP School":      -2147483647,
	"Synersia School":   -2147483646,
	"Renseikai School":  -2147483645,
	"End-to-end School": -2147483644,
	"GA School":         -2147483643,
	"KEC School":        -2147483642,
	"AIC School":        -2147483641,
	"NSG School":        -2147483640,
	"E2E Tokyo":         -2147483639,
	"E2E HCM":           -2147483638,
	"Manabie Demo":      -2147483637,
}

const (
	StaffRoleSchoolAdmin   = "staff granted role school admin"
	StaffRoleHQStaff       = "staff granted role hq staff"
	StaffRoleCentreLead    = "staff granted role centre lead"
	StaffRoleCentreManager = "staff granted role centre manager"
	StaffRoleCentreStaff   = "staff granted role centre staff"
	StaffRoleTeacher       = "staff granted role teacher"
	StaffRoleTeacherLead   = "staff granted role teacher lead"
	UsermgmtScheduleJob    = "RoleUsermgmtScheduleJob"
	Student                = "student"
)

type userOption func(u *entity.LegacyUser)

func newID() string {
	return idutil.ULIDNow()
}

func InitUsermgmtState(ctx context.Context) context.Context {
	return StepStateToContext(ctx, &common.StepState{})
}

func newTenantToCreate() internal_auth_tenant.TenantInfo {
	bigNum := 999999999999999
	random := rand.Intn(bigNum)
	return &entity.Tenant{
		DisplayName:            fmt.Sprintf("test-%v", random),
		PasswordSignUpAllowed:  true,
		EmailLinkSignInEnabled: false,
	}
}

func withID(id string) userOption {
	return func(u *entity.LegacyUser) {
		_ = u.ID.Set(id)
	}
}

func withRole(group string) userOption {
	return func(u *entity.LegacyUser) {
		_ = u.Group.Set(group)
	}
}

type ImportCSVErrors interface {
	GetRowNumber() int32
	GetError() string
}

func checkInvalidRows(invalidCsvRows []string, reqSplit []string, respErrors []ImportCSVErrors) error {
	for _, row := range invalidCsvRows {
		found := false
		for _, e := range respErrors {
			if strings.TrimSpace(reqSplit[e.GetRowNumber()-1]) == row {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid line is not returned in response")
		}
	}
	return nil
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

// do not recommend use this function please use contextWithTokenV2 instead
func contextWithToken(ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", stepState.AuthToken)
}

func contextWithTokenV2(ctx context.Context, token string) context.Context {
	return helper.GRPCContext(ctx, "token", token)
}

func generateValidAuthenticationToken(sub, userGroup string) (string, error) {
	return generateAuthenticationToken(sub, "templates/"+userGroup+".template")
}

func switchSchoolIDStringToSchoolID(schoolID string) int {
	switch schoolID {
	case JPREPSchool:
		return constants.JPREPSchool
	case SynersiaSchool:
		return constants.SynersiaSchool
	case RenseikaiSchool:
		return constants.RenseikaiSchool
	case TestingSchool:
		return constants.TestingSchool
	case GASchool:
		return constants.GASchool
	case KECSchool:
		return constants.KECSchool
	case AICSchool:
		return constants.AICSchool
	case NSGSchool:
		return constants.NSGSchool
	default:
		return constants.ManabieSchool
	}
}

func isEmptyString(s string) bool {
	return s == ""
}

func isBasicProfile(p *pb.StudentProfile) bool {
	return isEmptyString(p.Phone) && isEmptyString(p.Email)
}

func GetRoleFromConstant(role string) string {
	switch role {
	case StaffRoleSchoolAdmin, schoolAdminType:
		return constant.RoleSchoolAdmin
	case StaffRoleHQStaff:
		return constant.RoleHQStaff
	case StaffRoleCentreLead:
		return constant.RoleCentreLead
	case StaffRoleCentreManager:
		return constant.RoleCentreManager
	case StaffRoleCentreStaff:
		return constant.RoleCentreStaff
	case StaffRoleTeacher, teacherType:
		return constant.RoleTeacher
	case StaffRoleTeacherLead:
		return constant.RoleTeacherLead
	case studentType:
		return constant.RoleStudent
	case parentType:
		return constant.RoleParent
	case UsermgmtScheduleJob:
		return constant.RoleUsermgmtScheduleJob
	default:
		return ""
	}
}

func GetLegacyUserGroupFromConstant(role string) string {
	switch role {
	case StaffRoleSchoolAdmin, schoolAdminType, StaffRoleHQStaff,
		StaffRoleCentreLead, StaffRoleCentreManager, StaffRoleCentreStaff, UsermgmtScheduleJob:
		return constant.UserGroupSchoolAdmin
	case StaffRoleTeacher, teacherType, StaffRoleTeacherLead:
		return constant.UserGroupTeacher
	case studentType:
		return constant.UserGroupStudent
	case parentType:
		return constant.UserGroupParent
	default:
		return ""
	}
}

func OrgIDFromCtx(ctx context.Context) int {
	orgID, err := strconv.Atoi(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		fmt.Println(fmt.Errorf("strconv.Atoi(golibs.ResourcePathFromCtx()) %w", err))
	}
	return orgID
}

// TryUntilSuccess must be used with context.WithTimeout
func TryUntilSuccess(ctx context.Context, tryInterval time.Duration, tryFn func(ctx context.Context) (bool, error)) error {
	ticker := time.NewTicker(tryInterval)
	defer ticker.Stop()

	errCh := make(chan error, 1)
	go func(ctx context.Context) {
		for {
			select {
			case <-ticker.C:
				retry, err := tryFn(ctx)
				if retry {
					continue
				}
				select {
				case errCh <- err:
				case <-time.After(time.Second):
				}
				return
			case <-ctx.Done():
				select {
				case errCh <- ctx.Err():
				case <-time.After(time.Second):
				}
				return
			}
		}
	}(ctx)

	return <-errCh
}

func isFeatureToggleEnabled(ctx context.Context, u *unleash.Suite, featureToggleName string) (bool, error) {
	unleashClient := user_unleash.NewDefaultClient(u.UnleashSrvAddr, u.UnleashAPIKey, u.UnleashLocalAdminAPIKey)

	details, err := unleashClient.GetFeatureToggleDetails(ctx, featureToggleName)
	if err != nil {
		return false, errors.Wrap(err, "unleashClient.GetFeatureToggleDetails")
	}

	return details.Enabled, nil
}
