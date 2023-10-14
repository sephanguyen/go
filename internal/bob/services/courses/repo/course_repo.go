package repo

import (
	"context"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type CourseRepo interface {
	FindByID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) (*entities.Course, error)
	FindByIDs(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error)
	FindByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (entities.Courses, error)
	GetPresetStudyPlanIDsByCourseIDs(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) ([]string, error)
	Upsert(ctx context.Context, db database.Ext, cc []*entities.Course) error
	UpdateStartAndEndDate(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) error
}

var _ CourseRepo = new(CourseRepoMock)

type CourseRepoMock struct {
	FindByIDMock                         func(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) (*entities.Course, error)
	FindByIDsMock                        func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error)
	FindByLessonIDMock                   func(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (entities.Courses, error)
	GetPresetStudyPlanIDsByCourseIDsMock func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) ([]string, error)
	UpsertMock                           func(ctx context.Context, db database.Ext, cc []*entities.Course) error
	UpdateStartAndEndDateMock            func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) error
}

func (c CourseRepoMock) FindByID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) (*entities.Course, error) {
	return c.FindByIDMock(ctx, db, courseID)
}

func (c CourseRepoMock) FindByIDs(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
	return c.FindByIDsMock(ctx, db, courseIDs)
}

func (c CourseRepoMock) FindByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (entities.Courses, error) {
	return c.FindByLessonIDMock(ctx, db, lessonID)
}

func (c CourseRepoMock) GetPresetStudyPlanIDsByCourseIDs(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) ([]string, error) {
	return c.GetPresetStudyPlanIDsByCourseIDsMock(ctx, db, courseIDs)
}

func (c CourseRepoMock) Upsert(ctx context.Context, db database.Ext, cc []*entities.Course) error {
	return c.UpsertMock(ctx, db, cc)
}

func (c CourseRepoMock) UpdateStartAndEndDate(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) error {
	return c.UpdateStartAndEndDateMock(ctx, db, courseIDs)
}
