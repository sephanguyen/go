package errcode

import (
	"fmt"

	"github.com/pkg/errors"
)

var ErrCannotGetTenant = errors.New("cannot get tenant")

// TenantDoesNotExistErr legacy error before apply hexagon, please avoid using this
type TenantDoesNotExistErr struct {
	OrganizationID string
}

func (err TenantDoesNotExistErr) Error() string {
	return fmt.Sprintf(`tenant for organization with id: "%s" does not exist`, err.OrganizationID)
}
