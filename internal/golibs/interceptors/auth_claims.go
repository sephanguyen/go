package interceptors

import (
	"strings"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/square/go-jose/v3/jwt"
)

// CustomClaims can be use for parsing both Firebase or Cognito claims
type CustomClaims struct {
	jwt.Claims
	*FirebaseClaims
	*JPREPClaims
	Hasura  *HasuraClaims  `json:"https://hasura.io/jwt/claims,omitempty"`
	Manabie *ManabieClaims `json:"manabie,omitempty"`

	JwkURL string `json:"-"`
}

// OrganizationID func implement Organization interface
// in Usermgmt domain
func (c *CustomClaims) OrganizationID() field.String {
	if c == nil || c.Manabie == nil {
		return field.NewUndefinedString()
	}
	return field.NewString(c.Manabie.ResourcePath)
}

func (c *CustomClaims) GetProjectID() string {
	splitStr := strings.Split(c.Issuer, "/")
	return splitStr[len(splitStr)-1]
}

// DefaultRole will prioritize Manabie over Hasura metadata
func (c *CustomClaims) DefaultRole() string {
	if c.Manabie != nil {
		return c.Manabie.DefaultRole
	}

	if c.Hasura != nil {
		return c.Hasura.DefaultRole
	}

	return ""
}

type Organization struct {
	organizationID string
	schoolID       int32
}

func NewOrganization(organizationID string, schoolID int32) *Organization {
	return &Organization{
		organizationID: organizationID,
		schoolID:       schoolID,
	}
}

func (org *Organization) OrganizationID() field.String {
	return field.NewString(org.organizationID)
}

func (org *Organization) SchoolID() field.Int32 {
	return field.NewInt32(org.schoolID)
}

type TokenInfo struct {
	Applicant    string
	UserID       string
	SchoolIds    []int64
	DefaultRole  string
	AllowedRoles []string
	UserGroup    string
	ResourcePath string
}

// HasuraClaims custom claims, used inside FirebaseClaims
type HasuraClaims struct {
	AllowedRoles []string `json:"x-hasura-allowed-roles,omitempty"`
	DefaultRole  string   `json:"x-hasura-default-role,omitempty"`
	UserID       string   `json:"x-hasura-user-id,omitempty"`
	SchoolIDs    string   `json:"x-hasura-school-ids,omitempty"`
	UserGroup    string   `json:"x-hasura-user-group,omitempty"`
	ResourcePath string   `json:"x-hasura-resource-path,omitempty"`
}

// ManabieClaims has exact value as HasuraClaims but without prefix
type ManabieClaims struct {
	AllowedRoles []string `json:"allowed_roles,omitempty"`
	DefaultRole  string   `json:"default_role"`
	UserID       string   `json:"user_id"`
	SchoolIDs    []string `json:"school_ids,omitempty"`
	UserGroup    string   `json:"user_group,omitempty"`
	ResourcePath string   `json:"resource_path,omitempty"`
}

// FirebaseClaims describes firebase's jwt claims structure
type FirebaseClaims struct {
	Email         string           `json:"email,omitempty"`
	EmailVerified bool             `json:"email_verified"`
	Identity      FirebaseIdentity `json:"firebase"`
}

func (v *FirebaseClaims) GetTenantID() string {
	if v == nil {
		return ""
	}
	return v.Identity.Tenant
}

// FirebaseIdentity describes firebase's jwt claims user identity specific
type FirebaseIdentity struct {
	SignInProvider string              `json:"sign_in_provider"`
	Identities     map[string][]string `json:"identities"`
	Tenant         string              `json:"tenant,omitempty"`
}

// JPREPClaims describes JPREP specific value
type JPREPClaims struct {
	StudentDivision string `json:"student_division"`
}
