package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

// LessonGroupModifier implement builder for Lesson Group
type LessonGroupModifier struct {
	repo LessonGroupRepo
	db   database.Ext
}

func NewLessonGroupModifier(db database.Ext, repo LessonGroupRepo) *LessonGroupModifier {
	return &LessonGroupModifier{
		db:   db,
		repo: repo,
	}
}

func (l *LessonGroupModifier) CreateWithMedias(ctx context.Context, courseID pgtype.Text, mediaIDs pgtype.TextArray) (*entities.LessonGroup, error) {
	lg := &entities.LessonGroup{}
	database.AllNullEntity(lg)
	lg.MediaIDs = mediaIDs
	lg.CourseID = courseID
	err := l.repo.Create(ctx, l.db, lg)
	if err != nil {
		return nil, fmt.Errorf("s.LessonGroupRepo.Create: %w", err)
	}

	return lg, nil
}
