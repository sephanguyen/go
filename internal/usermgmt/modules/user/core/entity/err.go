package entity

import (
	"fmt"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
)

type Entity string
type Reason string
type NestedFieldName string

const (
	EnrollmentStatusHistories NestedFieldName = "enrollment_status_histories"
	SchoolHistories           NestedFieldName = "school_histories"
	PhoneNumbers              NestedFieldName = "phone_number"
	Address                   NestedFieldName = "address"
)

const (
	StudentEntity Entity = "student"
	UserEntity    Entity = "user"
	ParentEntity  Entity = "parent"
	StaffEntity   Entity = "staff"
	GradeEntity   Entity = "grade"
)

const (
	Empty                                        Reason = "empty"
	Archived                                     Reason = "archived"
	Invalid                                      Reason = "invalid"
	NotMatchingPattern                           Reason = "not matching pattern"
	NotMatchingEnum                              Reason = "not matching enum"
	NotPresentField                              Reason = "not present field"
	StartDateAfterEndDate                        Reason = "start date after end date"
	AlreadyRegistered                            Reason = "already registered"
	NotMatching                                  Reason = "not matching"
	NotMatchingConstants                         Reason = "not matching constants"
	NotInAllowListEnrollmentStatus               Reason = "not in allow list enrollment status"
	MissingActivatedEnrollmentStatus             Reason = "missing activated enrollment status"
	StartDateAfterCurrentDate                    Reason = "start date after current date"
	ChangingNonPotentialToOtherStatus            Reason = "changing non potential to other status"
	ChangingTemporaryToOtherStatus               Reason = "changing temporary to other status"
	ChangingOtherStatusToNonPotential            Reason = "changing other status to non potential"
	ChangingNonERPStatusToOtherStatusAtOrderFlow Reason = "changing non ERP status to other status at Order flow"
	ChangingStartDateWithoutChangingStatus       Reason = "changing start date without changing status"
	ActivatedStartDateAfterReqStartDate          Reason = "activated start date after request start date"
	ChangingStatusWithoutChangingStartDate       Reason = "changing status without changing start date"
	FailedUnmarshal                              Reason = "failed unmarshal"
	SchoolCourseDoesNotBelongToSchool            Reason = "school course does not belong to school"
	InvalidTagType                               Reason = "invalid tag type"
	LocationIsNotLowestLocation                  Reason = "not be the lowest locations"
)

type InvalidFieldError struct {
	EntityName Entity
	Index      int
	FieldName  string
	Reason     Reason
}

func (err InvalidFieldError) Error() string {
	return fmt.Sprintf(`The '%s' of '%s' at %v is %s`, err.FieldName, err.EntityName, err.Index, err.Reason)
}
func (err InvalidFieldError) DomainError() string {
	return fmt.Sprintf(`%s[%v].%s invalid`, err.EntityName, err.Index, err.FieldName)
}
func (err InvalidFieldError) DomainCode() int {
	return errcode.InvalidData
}

type InvalidFieldErrorWithArrayNestedField struct {
	InvalidFieldError
	NestedFieldName NestedFieldName
	NestedIndex     int
}

func (err InvalidFieldErrorWithArrayNestedField) Error() string {
	return fmt.Sprintf(`The '%s' of '%s' at %d in '%s' field at nested index: %d is %s`, err.EntityName, err.NestedFieldName, err.Index, err.FieldName, err.NestedIndex, err.Reason)
}
func (err InvalidFieldErrorWithArrayNestedField) DomainError() string {
	return fmt.Sprintf(`%s[%d].%s[%d].%s invalid`, err.EntityName, err.Index, err.NestedFieldName, err.NestedIndex, err.FieldName)
}

type InvalidFieldErrorWithObjectNestedField struct {
	InvalidFieldError
	NestedFieldName NestedFieldName
}

func (err InvalidFieldErrorWithObjectNestedField) Error() string {
	return fmt.Sprintf(`The '%s' of '%s' at %d in field: '%s' is %s`, err.EntityName, err.NestedFieldName, err.Index, err.NestedFieldName, err.Reason)
}
func (err InvalidFieldErrorWithObjectNestedField) DomainError() string {
	return fmt.Sprintf(`%s[%d].%s.%s invalid`, err.EntityName, err.Index, err.NestedFieldName, err.FieldName)
}

type DuplicatedFieldErrorWithArrayNestedField struct {
	DuplicatedFieldError
	NestedFieldName NestedFieldName
	NestedIndex     int
}

func (err DuplicatedFieldErrorWithArrayNestedField) Error() string {
	return fmt.Sprintf(`The '%s' of '%s' at %d in '%s' field at nested index: %d is duplicated`, err.NestedFieldName, err.EntityName, err.Index, err.DuplicatedField, err.NestedIndex)
}
func (err DuplicatedFieldErrorWithArrayNestedField) DomainError() string {
	return fmt.Sprintf(`%s[%d].%s[%d].%s is duplicated`, err.EntityName, err.Index, err.NestedFieldName, err.NestedIndex, err.DuplicatedField)
}
func (err DuplicatedFieldErrorWithArrayNestedField) DomainCode() int {
	return errcode.DuplicatedData
}

type NotFoundError struct {
	EntityName Entity
	Index      int
	FieldName  string
	FieldValue string
}

func (err NotFoundError) Error() string {
	return fmt.Sprintf(`can not find '%s' with '%s' = '%s' at index %d`, err.EntityName, err.FieldName, err.FieldValue, err.Index)
}
func (err NotFoundError) DomainError() string {
	return fmt.Sprintf(`%s[%v].%s not found`, err.EntityName, err.Index, err.FieldName)
}
func (err NotFoundError) DomainCode() int {
	return errcode.NotFound
}

type NotFoundErrorWithArrayNestedField struct {
	NotFoundError
	NestedFieldName NestedFieldName
	NestedIndex     int
}

func (err NotFoundErrorWithArrayNestedField) Error() string {
	return fmt.Sprintf(`Can not find nested field '%s' at index %d with '%s' = '%s' of %s at index %d`, err.FieldName, err.NestedIndex, err.NestedFieldName, err.FieldValue, err.EntityName, err.Index)
}
func (err NotFoundErrorWithArrayNestedField) DomainError() string {
	return fmt.Sprintf(`%s[%d].%s[%d].%s '%s' not found`, err.EntityName, err.Index, err.NestedFieldName, err.NestedIndex, err.FieldName, err.FieldValue)
}
func (err NotFoundErrorWithArrayNestedField) DomainCode() int {
	return errcode.NotFound
}

type MissingMandatoryFieldError struct {
	EntityName Entity
	Index      int
	FieldName  string
}

func (err MissingMandatoryFieldError) Error() string {
	return fmt.Sprintf(`The field '%s' is required in entity '%s' at index %d`, err.FieldName, err.EntityName, err.Index)
}
func (err MissingMandatoryFieldError) DomainError() string {
	return fmt.Sprintf(`%s[%v].%s required`, err.EntityName, err.Index, err.FieldName)
}
func (err MissingMandatoryFieldError) DomainCode() int {
	return errcode.MissingMandatory
}

type InternalError struct {
	RawErr error
}

func (err InternalError) Error() string {
	return fmt.Sprintf(`In domain layer, internal error with raw error: %v`, err.RawErr)
}
func (err InternalError) DomainError() string {
	return "internal error"
}
func (err InternalError) DomainCode() int {
	return errcode.InternalError
}

type DuplicatedFieldError struct {
	DuplicatedField string
	EntityName      Entity
	Index           int
}

func (err DuplicatedFieldError) Error() string {
	return fmt.Sprintf(`Then field '%s' is duplicated in entity '%s' at index %d`, err.DuplicatedField, err.EntityName, err.Index)
}
func (err DuplicatedFieldError) DomainError() string {
	return fmt.Sprintf(`%s[%v].%s duplicated`, err.EntityName, err.Index, err.DuplicatedField)
}
func (err DuplicatedFieldError) DomainCode() int {
	return errcode.DuplicatedData
}

type ExistingDataError struct {
	FieldName  string
	EntityName Entity
	Index      int
}

func (err ExistingDataError) Error() string {
	return fmt.Sprintf(`existing data in field '%s' in entity '%s' at index %d`, err.FieldName, err.EntityName, err.Index)
}
func (err ExistingDataError) DomainError() string {
	return fmt.Sprintf(`%s[%v].%s existing`, err.EntityName, err.Index, err.FieldName)
}
func (err ExistingDataError) DomainCode() int {
	return errcode.DataExist
}

type UpdateFieldError struct {
	FieldName  string
	EntityName Entity
	Index      int
}

func (err UpdateFieldError) Error() string {
	return fmt.Sprintf(`can not update field '%s' in entity '%s' at index %d`, err.FieldName, err.EntityName, err.Index)
}
func (err UpdateFieldError) DomainError() string {
	return fmt.Sprintf(`%s[%v].%s don't allow to update`, err.EntityName, err.Index, err.FieldName)
}
func (err UpdateFieldError) DomainCode() int {
	return errcode.UpdateFieldFail
}
