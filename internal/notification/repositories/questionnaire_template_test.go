package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Upsert(t *testing.T) {
	t.Parallel()

	genMockArgs := func() []interface{} {
		e := &entities.QuestionnaireTemplate{}
		fields := database.GetFieldNames(e)
		values := database.GetScanFields(e, fields)
		mockValues := make([]interface{}, 0, len(values)+2)
		mockValues = append(mockValues, mock.Anything)
		mockValues = append(mockValues, mock.AnythingOfType("string"))
		for range values {
			mockValues = append(mockValues, mock.Anything)
		}
		return mockValues
	}

	db := &mock_database.Ext{}

	testCases := []struct {
		Name        string
		ErrorExpect error
		Setup       func(ctx context.Context)
	}{
		{
			Name:        "happy case",
			ErrorExpect: nil,
			Setup: func(ctx context.Context) {
				db.On("Exec",
					genMockArgs()...,
				).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
		{
			Name:        "Error poll closed",
			ErrorExpect: fmt.Errorf("QuestionnaireTemplateRepo.Upsert: %w", puddle.ErrClosedPool),
			Setup: func(ctx context.Context) {
				db.On("Exec",
					genMockArgs()...,
				).Once().Return(pgconn.CommandTag([]byte(`0`)), puddle.ErrClosedPool)
			},
		},
		{
			Name:        "No row affected",
			ErrorExpect: fmt.Errorf("QuestionnaireTemplateRepo.Upsert: Questionnaire template is not inserted"),
			Setup: func(ctx context.Context) {
				db.On("Exec",
					genMockArgs()...,
				).Once().Return(pgconn.CommandTag([]byte(`0`)), nil)
			},
		},
	}

	questionnaireTemplateRepo := &QuestionnaireTemplateRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		testCase.Setup(ctx)
		err := questionnaireTemplateRepo.Upsert(ctx, db, &entities.QuestionnaireTemplate{})
		assert.Equal(t, testCase.ErrorExpect, err)
	}
}

func Test_CheckIsExistNameAndType(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	filter := NewCheckTemplateNameFilter()

	testCases := []struct {
		Name        string
		ExistExpect bool
		ErrorExpect error
		Setup       func(ctx context.Context)
	}{
		{
			Name:        "happy case",
			ExistExpect: true,
			ErrorExpect: nil,
			Setup: func(ctx context.Context) {
				countResult := int(1)
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything)
				mockDB.MockRowScanFields(nil, []string{""}, []interface{}{&countResult})
			},
		},
		{
			Name:        "count is 0",
			ExistExpect: false,
			ErrorExpect: nil,
			Setup: func(ctx context.Context) {
				countResult := int(0)
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything)
				mockDB.MockRowScanFields(nil, []string{""}, []interface{}{&countResult})
			},
		},
		{
			Name:        "conn pool closed",
			ExistExpect: false,
			ErrorExpect: fmt.Errorf("QuestionnaireTemplateRepo.CheckIsExistNameAndType: %w", puddle.ErrClosedPool),
			Setup: func(ctx context.Context) {
				countResult := int(0)
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything)
				mockDB.MockRowScanFields(puddle.ErrClosedPool, []string{""}, []interface{}{&countResult})
			},
		},
	}

	questionnaireTemplateRepo := &QuestionnaireTemplateRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		testCase.Setup(ctx)
		t.Run(testCase.Name, func(t *testing.T) {
			isExist, err := questionnaireTemplateRepo.CheckIsExistNameAndType(ctx, mockDB.DB, filter)
			if testCase.ErrorExpect == nil {
				assert.Equal(t, testCase.ExistExpect, isExist)
			} else {
				assert.Equal(t, testCase.ErrorExpect.Error(), err.Error())
			}
		})
	}
}
