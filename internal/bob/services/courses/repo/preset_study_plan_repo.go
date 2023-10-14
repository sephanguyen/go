package repo

import (
	"context"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type PresetStudyPlanRepo interface {
	CreatePresetStudyPlan(ctx context.Context, db database.Ext, presetStudyPlans []*entities.PresetStudyPlan) error
}

var _ PresetStudyPlanRepo = new(PresetStudyPlanRepoMock)

type PresetStudyPlanRepoMock struct {
	CreatePresetStudyPlanMock func(ctx context.Context, db database.Ext, presetStudyPlans []*entities.PresetStudyPlan) error
}

func (p PresetStudyPlanRepoMock) CreatePresetStudyPlan(ctx context.Context, db database.Ext, presetStudyPlans []*entities.PresetStudyPlan) error {
	return p.CreatePresetStudyPlanMock(ctx, db, presetStudyPlans)
}

type PresetStudyPlanWeeklyRepo interface {
	Create(ctx context.Context, db database.Ext, plans []*entities.PresetStudyPlanWeekly) error
	FindByLessonIDs(ctx context.Context, db database.QueryExecer, IDs pgtype.TextArray, isAll bool) (map[pgtype.Text]*entities.PresetStudyPlanWeekly, error)
	GetIDsByLessonIDAndPresetStudyPlanIDs(ctx context.Context, db database.Ext, lessonID pgtype.Text, pspIDs pgtype.TextArray) ([]string, error)
	UpdateTimeByLessonAndCourses(ctx context.Context, db database.Ext, lessonID pgtype.Text, courseIDs pgtype.TextArray, startDate, endDate pgtype.Timestamptz) error
	SoftDelete(ctx context.Context, db database.QueryExecer, pspwIDs pgtype.TextArray) error
}

type PresetStudyPlanWeeklyRepoMock struct {
	CreateMock                                func(ctx context.Context, db database.Ext, plans []*entities.PresetStudyPlanWeekly) error
	FindByLessonIDsMock                       func(ctx context.Context, db database.QueryExecer, IDs pgtype.TextArray, isAll bool) (map[pgtype.Text]*entities.PresetStudyPlanWeekly, error)
	GetIDsByLessonIDAndPresetStudyPlanIDsMock func(ctx context.Context, db database.Ext, lessonID pgtype.Text, pspIDs pgtype.TextArray) ([]string, error)
	UpdateTimeByLessonAndCoursesMock          func(ctx context.Context, db database.Ext, lessonID pgtype.Text, courseIDs pgtype.TextArray, startDate, endDate pgtype.Timestamptz) error
	SoftDeleteMock                            func(ctx context.Context, db database.QueryExecer, PresetStudyPlanWeeklyIDs pgtype.TextArray) error
}

func (p PresetStudyPlanWeeklyRepoMock) Create(ctx context.Context, db database.Ext, plans []*entities.PresetStudyPlanWeekly) error {
	return p.CreateMock(ctx, db, plans)
}

func (p PresetStudyPlanWeeklyRepoMock) FindByLessonIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, isAll bool) (map[pgtype.Text]*entities.PresetStudyPlanWeekly, error) {
	return p.FindByLessonIDsMock(ctx, db, ids, isAll)
}

func (p PresetStudyPlanWeeklyRepoMock) GetIDsByLessonIDAndPresetStudyPlanIDs(ctx context.Context, db database.Ext, lessonID pgtype.Text, pspIDs pgtype.TextArray) ([]string, error) {
	return p.GetIDsByLessonIDAndPresetStudyPlanIDsMock(ctx, db, lessonID, pspIDs)
}

func (p PresetStudyPlanWeeklyRepoMock) UpdateTimeByLessonAndCourses(ctx context.Context, db database.Ext, lessonID pgtype.Text, courseIDs pgtype.TextArray, startDate, endDate pgtype.Timestamptz) error {
	return p.UpdateTimeByLessonAndCoursesMock(ctx, db, lessonID, courseIDs, startDate, endDate)
}

func (p PresetStudyPlanWeeklyRepoMock) SoftDelete(ctx context.Context, db database.QueryExecer, pspwIDs pgtype.TextArray) error {
	return p.SoftDeleteMock(ctx, db, pspwIDs)
}
