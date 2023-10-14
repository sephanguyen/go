package queries

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure"
)

type StudentSubscriptionQueryHandler struct {
	DB database.Ext

	// ports
	StudentSubscriptionRepo infrastructure.StudentSubscriptionRepo
}

func (c *StudentSubscriptionQueryHandler) GetLocationsBelongToActiveStudentSubscriptionsByCourses(ctx context.Context, payload GetLocationsBelongToActiveStudentSubscriptionsByCourses) (map[string][]string, error) {
	return c.StudentSubscriptionRepo.GetLocationActiveStudentSubscriptions(ctx, c.DB, payload.CourseIDs)
}
