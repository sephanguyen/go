// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type MockPresetStudyPlanRepo struct {
	mock.Mock
}

func (r *MockPresetStudyPlanRepo) BulkImport(arg1 context.Context, arg2 database.QueryExecer, arg3 []*entities.PresetStudyPlan, arg4 []*entities.PresetStudyPlanWeekly) error {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Error(0)
}

func (r *MockPresetStudyPlanRepo) CreatePresetStudyPlan(arg1 context.Context, arg2 database.Ext, arg3 []*entities.PresetStudyPlan) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockPresetStudyPlanRepo) CreatePresetStudyPlanWeekly(arg1 context.Context, arg2 database.QueryExecer, arg3 []*entities.PresetStudyPlanWeekly) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockPresetStudyPlanRepo) RetrievePresetStudyPlanWeeklies(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) ([]*entities.PresetStudyPlanWeekly, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.PresetStudyPlanWeekly), args.Error(1)
}

func (r *MockPresetStudyPlanRepo) RetrievePresetStudyPlans(arg1 context.Context, arg2 database.QueryExecer, arg3 string, arg4 string, arg5 string, arg6 int) ([]*entities.PresetStudyPlan, error) {
	args := r.Called(arg1, arg2, arg3, arg4, arg5, arg6)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.PresetStudyPlan), args.Error(1)
}

func (r *MockPresetStudyPlanRepo) RetrieveStudentCompletedTopics(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text, arg4 pgtype.TextArray) ([]*entities.Topic, error) {
	args := r.Called(arg1, arg2, arg3, arg4)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Topic), args.Error(1)
}

func (r *MockPresetStudyPlanRepo) RetrieveStudentPresetStudyPlanWeeklies(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text, arg4 *pgtype.Timestamptz, arg5 *pgtype.Timestamptz, arg6 bool) ([]repositories.Topic, error) {
	args := r.Called(arg1, arg2, arg3, arg4, arg5, arg6)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.Topic), args.Error(1)
}

func (r *MockPresetStudyPlanRepo) RetrieveStudentPresetStudyPlans(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text, arg4 *pgtype.Timestamptz, arg5 *pgtype.Timestamptz) ([]*repositories.PlanWithStartDate, error) {
	args := r.Called(arg1, arg2, arg3, arg4, arg5)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repositories.PlanWithStartDate), args.Error(1)
}

func (r *MockPresetStudyPlanRepo) RetrieveStudyAheadTopicsOfPresetStudyPlan(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text, arg4 pgtype.Text) ([]repositories.AheadTopic, error) {
	args := r.Called(arg1, arg2, arg3, arg4)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.AheadTopic), args.Error(1)
}

func (r *MockPresetStudyPlanRepo) UpdatePresetStudyPlanWeeklyEndTime(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text, arg4 pgtype.Timestamptz) error {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Error(0)
}
