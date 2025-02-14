// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type MockLessonReportDetailRepo struct {
	mock.Mock
}

func (r *MockLessonReportDetailRepo) DeleteByLessonReportID(arg1 context.Context, arg2 database.Ext, arg3 pgtype.Text) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockLessonReportDetailRepo) DeleteFieldValuesByDetails(arg1 context.Context, arg2 database.Ext, arg3 pgtype.TextArray) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockLessonReportDetailRepo) GetByLessonReportID(arg1 context.Context, arg2 database.Ext, arg3 pgtype.Text) (entities.LessonReportDetails, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(entities.LessonReportDetails), args.Error(1)
}

func (r *MockLessonReportDetailRepo) GetFieldValuesByDetailIDs(arg1 context.Context, arg2 database.Ext, arg3 pgtype.TextArray) (entities.PartnerDynamicFormFieldValues, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(entities.PartnerDynamicFormFieldValues), args.Error(1)
}

func (r *MockLessonReportDetailRepo) Upsert(arg1 context.Context, arg2 database.Ext, arg3 pgtype.Text, arg4 entities.LessonReportDetails) error {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Error(0)
}

func (r *MockLessonReportDetailRepo) UpsertFieldValues(arg1 context.Context, arg2 database.Ext, arg3 []*entities.PartnerDynamicFormFieldValue) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockLessonReportDetailRepo) UpsertFieldValuesQueue(arg1 *pgx.Batch, arg2 *entities.PartnerDynamicFormFieldValue) {
	_ = r.Called(arg1, arg2)
	return
}

func (r *MockLessonReportDetailRepo) UpsertQueue(arg1 *pgx.Batch, arg2 *entities.LessonReportDetail) {
	_ = r.Called(arg1, arg2)
	return
}
