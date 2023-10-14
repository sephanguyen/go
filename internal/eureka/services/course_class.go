package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"go.uber.org/multierr"
)

type CourseClassService struct {
	DB database.Ext

	CourseClassRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.CourseClass) error
		Delete(ctx context.Context, db database.QueryExecer, items []*entities.CourseClass) error
	}
}

func getCourseClassFromReq(req []*npb.EventMasterRegistration_Class) ([]*entities.CourseClass, error) {
	courseClasses := make([]*entities.CourseClass, 0, len(req))

	for _, item := range req {
		courseClass := &entities.CourseClass{}
		database.AllNullEntity(courseClass)
		err := multierr.Combine(
			courseClass.ID.Set(idutil.ULIDNow()),
			courseClass.CourseID.Set(item.CourseId),
			courseClass.ClassID.Set(strconv.Itoa(int(item.ClassId))),
		)
		if err != nil {
			return nil, err
		}
		courseClasses = append(courseClasses, courseClass)
	}
	return courseClasses, nil
}

func (s *CourseClassService) upsertCourseClass(ctx context.Context, courseClasses []*entities.CourseClass) error {
	if len(courseClasses) == 0 {
		return nil
	}

	err := s.CourseClassRepo.BulkUpsert(ctx, s.DB, courseClasses)
	if err != nil {
		return fmt.Errorf("err s.CourseClassRepo.BulkUpsert: %w", err)
	}
	return nil
}

func (s *CourseClassService) softDeleteCourseClass(ctx context.Context, courseClasses []*entities.CourseClass) error {
	if len(courseClasses) == 0 {
		return nil
	}

	err := s.CourseClassRepo.Delete(ctx, s.DB, courseClasses)
	if err != nil {
		return fmt.Errorf("err s.CourseClassRepo.Delete: %w", err)
	}

	return nil
}

// CourseClassService handle EventUserRegistration event, upsert CourseClass if ActionKind=UPSERTED and softDelete if ActionKind=DELETED.
func (s *CourseClassService) SyncCourseClass(ctx context.Context, req *npb.EventMasterRegistration) error {
	courseClasses := make(map[string][]*npb.EventMasterRegistration_Class)
	for _, request := range req.Classes {
		if request.ActionKind == npb.ActionKind_ACTION_KIND_UPSERTED {
			courseClasses[npb.ActionKind_ACTION_KIND_UPSERTED.String()] = append(courseClasses[npb.ActionKind_ACTION_KIND_UPSERTED.String()], request)
		} else {
			courseClasses[npb.ActionKind_ACTION_KIND_DELETED.String()] = append(courseClasses[npb.ActionKind_ACTION_KIND_DELETED.String()], request)
		}
	}
	var err error
	var courseClassUpserted []*entities.CourseClass
	var courseClassDeleted []*entities.CourseClass

	if courseClassUpserted, err = getCourseClassFromReq(courseClasses[npb.ActionKind_ACTION_KIND_UPSERTED.String()]); err != nil {
		return fmt.Errorf("err getCourseClassFromReq upserted: %w", err)
	}
	if courseClassDeleted, err = getCourseClassFromReq(courseClasses[npb.ActionKind_ACTION_KIND_DELETED.String()]); err != nil {
		return fmt.Errorf("err getCourseClassFromReq softdelete: %w", err)
	}

	err = multierr.Combine(
		s.upsertCourseClass(ctx, courseClassUpserted),
		s.softDeleteCourseClass(ctx, courseClassDeleted),
	)
	if err != nil {
		return err
	}
	return nil
}
