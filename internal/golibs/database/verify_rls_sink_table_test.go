package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRLSForSinkTable(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	errStr := `table locations in db calendar missing AC ; table locations in db entryexitmgmt missing AC ; table courses in db eureka missing AC ; table locations in db eureka missing AC ; table students in db eureka missing AC ; table user_access_paths in db eureka missing AC ; table users in db eureka missing AC ; table courses in db fatima missing AC ; table locations in db fatima missing AC ; table students in db fatima missing AC ; table user_access_paths in db fatima missing AC ; table users in db fatima missing AC ; table locations in db invoicemgmt missing AC ; table locations in db mastermgmt missing AC ; table courses in db timesheet missing AC ; table locations in db timesheet missing AC ; table staff in db timesheet missing AC ; table user_access_paths in db tom missing AC ; table users in db tom missing AC ; table student_product in db bob missing AC ; table lessons in db timesheet missing AC `
	assert.Equal(errStr, VerifyACForAllSinkTable().Error())
}
