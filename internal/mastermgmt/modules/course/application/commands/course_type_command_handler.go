package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure"

	"github.com/jackc/pgx/v4"
)

type CourseTypeCommandHandler struct {
	DB database.Ext

	// ports
	CourseTypeRepo infrastructure.CourseTypeRepo
}

func (c *CourseTypeCommandHandler) ImportCourseTypes(ctx context.Context, payload ImportCourseTypesPayload) (err error) {
	err = database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err = c.CourseTypeRepo.Import(ctx, c.DB, payload.CourseTypes)
		return err
	})
	if err != nil {
		return fmt.Errorf("CourseTypeRepo.Import: %w", err)
	}
	return nil
}
