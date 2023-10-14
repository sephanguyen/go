package valueobj

import (
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type HasOrganizationID interface {
	OrganizationID() field.String
}

type HasSchoolID interface {
	SchoolID() field.Int32
}

type HasUserID interface {
	UserID() field.String
}

type HasUserIDs []HasUserID

type HasStudentID interface {
	StudentID() field.String
}

type HasParentID interface {
	ParentID() field.String
}

type RandomHasUserID struct {
	RandomUserID field.String
}

func (r *RandomHasUserID) UserID() field.String {
	return r.RandomUserID
}

type HasPartnerInternalID interface {
	PartnerInternalID() field.String
}

type HasPartnerInternalIDs []HasPartnerInternalID

type HasLocationID interface {
	LocationID() field.String
}

type RandomHasLocationID struct{}

func (*RandomHasLocationID) LocationID() field.String {
	return field.NewString(idutil.ULIDNow())
}

type HasTagID interface {
	TagID() field.String
}

type HasUserAddressID interface {
	UserAddressID() field.String
}

type HasPrefectureID interface {
	PrefectureID() field.String
}

// user group
type HasUserGroupID interface {
	UserGroupID() field.String
}

type HasRoleID interface {
	RoleID() field.String
}

type HasPermissionID interface {
	PermissionID() field.String
}

type HasCountry interface {
	Country() field.String
}

type HasSchoolInfoID interface {
	SchoolID() field.String
}

type HasSchoolCourseID interface {
	SchoolCourseID() field.String
}

type HasGradeID interface {
	GradeID() field.String
}

type HasStudentPackageID interface {
	StudentPackageID() field.String
}

type HasPackageID interface {
	PackageID() field.String
}

type HasCreatedAt interface {
	CreatedAt() field.Time
}

type HasLoginEmail interface {
	LoginEmail() field.String
}
