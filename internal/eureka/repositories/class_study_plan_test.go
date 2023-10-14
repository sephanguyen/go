package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClassStudyPlanBulkUpsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	classStudyPlanRepo := &ClassStudyPlanRepo{}
	validClassStudyPlanReq := []*entities.ClassStudyPlan{
		{
			ClassID:     database.Int4(1),
			StudyPlanID: database.Text("study-plan-id-1"),
		},
		{
			ClassID:     database.Int4(1),
			StudyPlanID: database.Text("study-plan-id-2"),
		},
	}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validClassStudyPlanReq,
			expectedErr: nil,
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validClassStudyPlanReq); i++ {
					_, field := validClassStudyPlanReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
		{
			name:        "exec error",
			req:         validClassStudyPlanReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertClassStudyPlan error: exec error"),
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validClassStudyPlanReq); i++ {
					_, field := validClassStudyPlanReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, fmt.Errorf("exec error"))
			},
		},
		{
			name:        "no row affected",
			req:         validClassStudyPlanReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertClassStudyPlan error: no row affected"),
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validClassStudyPlanReq); i++ {
					_, field := validClassStudyPlanReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`0`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := classStudyPlanRepo.BulkUpsert(ctx, db, testCase.req.([]*entities.ClassStudyPlan))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
