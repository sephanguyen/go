package application

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	media_module "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
)

type RetrieveLessonCommand struct {
	WrapperConnection *support.WrapperDBConnection

	// ports
	SearchRepo       infrastructure.SearchRepo
	LessonRepo       infrastructure.LessonRepo
	LessonMemberRepo infrastructure.LessonMemberRepo
	LessonGroupRepo  infrastructure.LessonGroupRepo
}

func (r *RetrieveLessonCommand) GetLessonByID(ctx context.Context, lessonID string) (*domain.Lesson, error) {
	conn, err := r.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	lesson, err := r.LessonRepo.GetLessonByID(ctx, conn, lessonID)
	if err != nil {
		return nil, fmt.Errorf("LessonRepo.GetLessonByID err: %w", err)
	}

	return lesson, nil
}

func (r *RetrieveLessonCommand) GetLessonByIDs(ctx context.Context, lessonIDs []string) ([]*domain.Lesson, error) {
	conn, err := r.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	lesson, err := r.LessonRepo.GetLessonByIDs(ctx, conn, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("LessonRepo.GetLessonByIDs err: %w", err)
	}

	return lesson, nil
}

func (r *RetrieveLessonCommand) Search(ctx context.Context, args *domain.ListLessonArgs) (lessons []*domain.Lesson, total uint32, offsetID string, err error) {
	lessons, total, offsetID, err = r.SearchRepo.Search(ctx, args)
	if err != nil {
		return nil, 0, "", fmt.Errorf("Search: %w", err)
	}

	return lessons, total, offsetID, nil
}
func (r *RetrieveLessonCommand) RetrieveLessonMembersByLessonArgs(ctx context.Context, args *domain.ListStudentsByLessonArgs) ([]*domain.User, error) {
	conn, err := r.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	students, err := r.LessonMemberRepo.ListStudentsByLessonArgs(ctx, conn, args)
	if err != nil {
		return nil, fmt.Errorf("LessonMemberRepo.ListStudentsByLessonArgs err: %w", err)
	}

	return students, nil
}

func (r *RetrieveLessonCommand) RetrieveMediasByLessonArgs(ctx context.Context, args *domain.ListMediaByLessonArgs) (media_module.Medias, error) {
	conn, err := r.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	medias, err := r.LessonGroupRepo.ListMediaByLessonArgs(ctx, conn, args)
	if err != nil {
		return nil, fmt.Errorf("LessonGroupRepo.ListMediaByLessonArgs err: %w", err)
	}

	return medias, nil
}
