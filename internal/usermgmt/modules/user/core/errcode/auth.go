package errcode

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrInternalFailedToImportAuthUsersToFirebaseAuth     = errors.New("failed to import auth users to firebase auth")
	ErrInternalFailedToImportAuthUsersToIdentityPlatform = errors.New("failed to import auth users to identity platform")
)

type ErrInternalFailedToImportAuthErr struct {
	Err error
}

func (err ErrInternalFailedToImportAuthErr) Error() string {
	return fmt.Sprintf("failed to import auth user: %v", err.Err)
}

func (err ErrInternalFailedToImportAuthErr) ErrorCode() int {
	return InternalError
}

type ErrAuthProfilesHaveIssueWhenImport struct {
	ErrMessages []string
	TenantID    string
}

func (err ErrAuthProfilesHaveIssueWhenImport) Error() string {
	return fmt.Sprintf("failed to import auth users, tenant id: '%s', err: %s", err.TenantID, strings.Join(err.ErrMessages, ", "))
}

func (err ErrAuthProfilesHaveIssueWhenImport) ErrorCode() int {
	return InvalidData
}

type ErrFailedToImportAuthUsersToTenantErr struct {
	TenantID string
	Err      error
}

func (err ErrFailedToImportAuthUsersToTenantErr) Error() string {
	var msg string
	if err.Err != nil {
		msg = err.Err.Error()
	}
	return fmt.Sprintf("failed to import auth users, tenant id: '%s', err: %s", err.TenantID, msg)
}

func (err ErrFailedToImportAuthUsersToTenantErr) ErrorCode() int {
	return InternalError
}

type ErrScryptIsInvalidErr struct {
	Err error
}

func (err ErrScryptIsInvalidErr) Error() string {
	return fmt.Sprintf("scrypt is invalid: %s", err.Err)
}

func (err ErrScryptIsInvalidErr) ErrorCode() int {
	return InternalError
}

type ErrTenantOfOrgNotFound struct {
	OrganizationID string
}

func (err ErrTenantOfOrgNotFound) Error() string {
	return fmt.Sprintf("organization id: '%s' not found", err.OrganizationID)
}

func (err ErrTenantOfOrgNotFound) ErrorCode() int {
	return InternalError
}

type ErrIdentityPlatformTenantNotFound struct {
	TenantID string
}

func (err ErrIdentityPlatformTenantNotFound) Error() string {
	return fmt.Sprintf("tenant id: '%s' not found", err.TenantID)
}

func (err ErrIdentityPlatformTenantNotFound) ErrorCode() int {
	return InternalError
}
