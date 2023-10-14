package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure"

	"github.com/jackc/pgx/v4"
)

type CourseCommandHandler struct {
	DB database.Ext

	// ports
	CourseRepo           infrastructure.CourseRepo
	CourseAccessPathRepo infrastructure.CourseAccessPathRepo
}

func (c *CourseCommandHandler) UpsertCourses(ctx context.Context, payload UpdateCoursesCommand) error {
	return database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := c.CourseRepo.Upsert(ctx, tx, payload.Courses)
		if err != nil {
			return err
		}
		if len(payload.CourseIDs) > 0 {
			if err := c.CourseAccessPathRepo.Delete(ctx, c.DB, payload.CourseIDs); err != nil {
				return fmt.Errorf("CourseAccessPathRepo.Delete: %w", err)
			}
		}
		err = c.CourseAccessPathRepo.Upsert(ctx, tx, payload.CourseAccessPaths)
		if err != nil {
			return err
		}
		err = c.CourseRepo.LinkSubjects(ctx, tx, payload.Courses)
		if err != nil {
			return fmt.Errorf("LinkSubjects: %w", err)
		}
		return nil
	})
}

func (c *CourseCommandHandler) ImportCourses(ctx context.Context, payload ImportCoursesPayload) (err error) {
	var getPartnerID = func(c *domain.Course) string {
		return c.PartnerID
	}
	var skipEmptyPartnerID = func(c *domain.Course) bool {
		return c.PartnerID == ""
	}
	partnerIDs := sliceutils.MapSkip(payload.Courses, getPartnerID, skipEmptyPartnerID)

	existingCourses, err := c.CourseRepo.GetByPartnerIDs(ctx, c.DB, partnerIDs)
	if err != nil {
		return err
	}

	coursePartnerMap := make(map[string]*domain.Course, len(existingCourses))
	for _, v := range existingCourses {
		coursePartnerMap[v.PartnerID] = v
	}

	// if exists then get the existing id, if not, generate an id
	for _, c := range payload.Courses {
		if c.PartnerID == "" {
			continue
		}
		ec, ok := coursePartnerMap[c.PartnerID]
		if ok {
			c.CourseID = ec.CourseID
		}
	}

	err = database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err = c.CourseRepo.Import(ctx, tx, payload.Courses)
		return err
	})

	if err != nil {
		return fmt.Errorf("CourseRepo.Import: %w", err)
	}
	return nil
}
