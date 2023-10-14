package entryexitmanagement

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/yasuo/constant"
)

func (s *suite) thisStudentHasExistingEntryAndExitRecord(ctx context.Context) error {
	// get student
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	s.stepState.UserGroupInContext = constant.UserGroupStudent
	studentID := s.UserGroupCredentials[s.UserGroupInContext].UserID

	// existing entry & exit records created with school admin permission
	err := s.signedInAsAccountWithResourcePath(ctx, "school admin", s.stepState.ResourcePath)
	if err != nil {
		return err
	}
	err = s.createEntryExitRecord(ctx, studentID)
	if err != nil {
		return err
	}
	return nil
}
