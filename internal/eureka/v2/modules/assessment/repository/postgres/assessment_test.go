package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/constants"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssessmentRepo_GetManyByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mockDB := testutil.NewMockDB()
	repo := &AssessmentRepo{}
	keys := []string{"key", "key-1"}

	t.Run("query failed return DB error", func(t *testing.T) {
		// arrange
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.TextArray(keys))
		expectedErr := errors.NewDBError("AssessmentRepo.GetManyByIDs", puddle.ErrClosedPool)

		// act
		c, err := repo.GetManyByIDs(ctx, mockDB.DB, keys)

		// assert
		assert.Nil(t, c)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("scan failed return conversion error", func(t *testing.T) {
		// arrange
		asm := &dto.Assessment{}
		_, val := asm.FieldMap()
		mockDB.DB.
			On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(keys)).
			Once().
			Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Close").Once().Return(nil)
		scanErr := fmt.Errorf("%s", "some err")
		mockDB.Rows.On("Scan", val...).Once().Return(scanErr)
		expectedErr := errors.NewConversionError("AssessmentRepo.scanAssessment", scanErr)

		// act
		c, err := repo.GetManyByIDs(ctx, mockDB.DB, keys)

		// assert
		assert.Nil(t, c)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("select succeeded", func(t *testing.T) {
		// arrange
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(keys))
		e := &dto.Assessment{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		// act
		_, err := repo.GetManyByIDs(ctx, mockDB.DB, keys)

		// assert
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"id":         {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestAssessmentRepo_GetManyByLMAndCourseIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	repo := &AssessmentRepo{}
	var getValueHolders = func(ass []domain.Assessment) []interface{} {
		values := make([]interface{}, 0, len(ass))
		for _, asm := range ass {
			values = append(values, asm.LearningMaterialID, asm.CourseID)
		}
		return values
	}

	t.Run("query failed returns DB error", func(t *testing.T) {
		// arrange
		n := 5
		assessments := getRandomAssessments(n)
		values := getValueHolders(assessments)
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"),
			values[0],
			values[1],
			values[2],
			values[3],
			values[4],
			values[5],
			values[6],
			values[7],
			values[8],
			values[9])
		expectedErr := errors.NewDBError("AssessmentRepo.scanAssessment", puddle.ErrClosedPool)

		// act
		c, err := repo.GetManyByLMAndCourseIDs(ctx, mockDB.DB, assessments)

		// assert
		assert.Nil(t, c)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("scan failed returns conversion error", func(t *testing.T) {
		// arrange
		n := 2
		assessments := getRandomAssessments(n)
		values := getValueHolders(assessments)
		asm := &dto.Assessment{}
		_, val := asm.FieldMap()
		mockDB.DB.
			On("Query", mock.Anything, mock.AnythingOfType("string"),
				values[0],
				values[1],
				values[2],
				values[3]).
			Once().
			Return(mockDB.Rows, nil)

		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Close").Once().Return(nil)
		scanErr := fmt.Errorf("%s", "some err")
		mockDB.Rows.On("Scan", val...).Once().Return(scanErr)
		expectedErr := errors.NewConversionError("AssessmentRepo.scanAssessment", scanErr)

		// act
		c, err := repo.GetManyByLMAndCourseIDs(ctx, mockDB.DB, assessments)

		// assert
		assert.Nil(t, c)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("select succeeded returns nil err", func(t *testing.T) {
		// arrange
		n := 5
		assessments := getRandomAssessments(n)
		values := getValueHolders(assessments)
		mockDB.MockQueryArgs(t, nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			values[0],
			values[1],
			values[2],
			values[3],
			values[4],
			values[5],
			values[6],
			values[7],
			values[8],
			values[9])
		e := &dto.Assessment{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		// act
		_, err := repo.GetManyByLMAndCourseIDs(ctx, mockDB.DB, assessments)

		// assert
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func getRandomAssessments(n int) []domain.Assessment {
	assessments := make([]domain.Assessment, n)
	for i := 0; i < n; i++ {
		assessments[i] = domain.Assessment{
			ID:                   idutil.ULIDNow(),
			CourseID:             idutil.ULIDNow(),
			LearningMaterialID:   idutil.ULIDNow(),
			LearningMaterialType: constants.LearningObjective,
		}
	}
	return assessments
}

func TestAssessmentRepo_GetOneByLMAndCourseID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	repo := &AssessmentRepo{}

	type Request struct {
		CourseID           string
		LearningMaterialID string
	}
	request := Request{
		CourseID:           "course_id",
		LearningMaterialID: "lm_id",
	}

	testCases := []struct {
		Name             string
		Ctx              context.Context
		Request          any
		MockDB           *testutil.MockDB
		Setup            func(ctx context.Context, mockDB *testutil.MockDB)
		ExpectedResponse any
		ExpectedError    error
	}{
		{
			Name:    "happy case",
			Ctx:     ctx,
			Request: request,
			MockDB:  testutil.NewMockDB(),
			Setup: func(ctx context.Context, mockDB *testutil.MockDB) {
				assessmentDto := dto.Assessment{
					ID:                 database.Text("id"),
					CourseID:           database.Text("course_id"),
					LearningMaterialID: database.Text("lm_id"),
					RefTable:           database.Varchar("learning_objective"),
				}
				fields, values := assessmentDto.FieldMap()

				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
					database.Text(request.CourseID), database.Text(request.LearningMaterialID))
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			ExpectedResponse: &domain.Assessment{
				ID:                   "id",
				CourseID:             "course_id",
				LearningMaterialID:   "lm_id",
				LearningMaterialType: constants.LearningObjective,
			},
			ExpectedError: nil,
		},
		{
			Name:    "unexpected error",
			Ctx:     ctx,
			Request: request,
			MockDB:  testutil.NewMockDB(),
			Setup: func(ctx context.Context, mockDB *testutil.MockDB) {
				assessmentDto := dto.Assessment{}
				fields, values := assessmentDto.FieldMap()

				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything,
					database.Text(request.CourseID), database.Text(request.LearningMaterialID))
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			ExpectedResponse: domain.Assessment{},
			ExpectedError:    errors.NewNoRowsExistedError("database.Select", fmt.Errorf("err db.Query: %w", pgx.ErrNoRows)),
		},
		{
			Name:    "unexpected error",
			Ctx:     ctx,
			Request: request,
			MockDB:  testutil.NewMockDB(),
			Setup: func(ctx context.Context, mockDB *testutil.MockDB) {
				assessmentDto := dto.Assessment{}
				fields, values := assessmentDto.FieldMap()

				mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything, mock.Anything,
					database.Text(request.CourseID), database.Text(request.LearningMaterialID))
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			ExpectedResponse: domain.Assessment{},
			ExpectedError:    errors.NewDBError("database.Select", fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed)),
		},
		{
			Name:    "ref table is empty",
			Ctx:     ctx,
			Request: request,
			MockDB:  testutil.NewMockDB(),
			Setup: func(ctx context.Context, mockDB *testutil.MockDB) {
				assessmentDto := dto.Assessment{
					ID:                 database.Text("id"),
					CourseID:           database.Text("course_id"),
					LearningMaterialID: database.Text("lm_id"),
					RefTable:           database.Varchar(""),
				}
				fields, values := assessmentDto.FieldMap()

				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
					database.Text(request.CourseID), database.Text(request.LearningMaterialID))
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			ExpectedResponse: domain.Assessment{
				ID:                   "id",
				CourseID:             "course_id",
				LearningMaterialID:   "lm_id",
				LearningMaterialType: constants.LearningObjective,
			},
			ExpectedError: errors.NewConversionError("result.ToEntity", domain.ErrInvalidLearningMaterialType),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(tc.Ctx, tc.MockDB)
			res, err := repo.GetOneByLMAndCourseID(tc.Ctx, tc.MockDB.DB, tc.Request.(Request).CourseID, tc.Request.(Request).LearningMaterialID)
			if err != nil {
				assert.Equal(t, tc.ExpectedError.Error(), err.Error())
			} else {
				assert.Equal(t, tc.ExpectedResponse, res)
			}
		})
	}
}

func TestAssessmentRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	repo := &AssessmentRepo{}

	now := time.Now()
	request := domain.Assessment{
		ID:                   "id",
		CourseID:             "course_id",
		LearningMaterialID:   "lm_id",
		LearningMaterialType: constants.LearningObjective,
	}

	testCases := []struct {
		Name             string
		Ctx              context.Context
		Request          any
		MockDB           *testutil.MockDB
		Setup            func(ctx context.Context, mockDB *testutil.MockDB)
		ExpectedResponse any
		ExpectedError    error
	}{
		{
			Name:    "happy case",
			Ctx:     ctx,
			Request: request,
			MockDB:  testutil.NewMockDB(),
			Setup: func(ctx context.Context, mockDB *testutil.MockDB) {
				assessmentDto := dto.Assessment{
					BaseEntity: dto.BaseEntity{
						CreatedAt: database.Timestamptz(now),
						UpdatedAt: database.Timestamptz(now),
						DeletedAt: pgtype.Timestamptz{
							Status: pgtype.Null,
						},
					},
					ID:                 database.Text(request.ID),
					CourseID:           database.Text(request.CourseID),
					LearningMaterialID: database.Text(request.LearningMaterialID),
					RefTable:           database.Varchar("learning_objective"),
				}
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything,
					&assessmentDto.ID, &assessmentDto.CourseID, &assessmentDto.LearningMaterialID, &assessmentDto.RefTable,
					&assessmentDto.CreatedAt, &assessmentDto.UpdatedAt, &assessmentDto.DeletedAt)

				fields := []string{"id"}
				values := []interface{}{&request.ID}
				mockDB.MockRowScanFields(nil, fields, values)
			},
			ExpectedResponse: "id",
			ExpectedError:    nil,
		},
		{
			Name:    "unexpected error",
			Ctx:     ctx,
			Request: request,
			MockDB:  testutil.NewMockDB(),
			Setup: func(ctx context.Context, mockDB *testutil.MockDB) {
				assessmentDto := dto.Assessment{
					BaseEntity: dto.BaseEntity{
						CreatedAt: database.Timestamptz(now),
						UpdatedAt: database.Timestamptz(now),
						DeletedAt: pgtype.Timestamptz{
							Status: pgtype.Null,
						},
					},
					ID:                 database.Text(request.ID),
					CourseID:           database.Text(request.CourseID),
					LearningMaterialID: database.Text(request.LearningMaterialID),
					RefTable:           database.Varchar("learning_objective"),
				}
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything,
					&assessmentDto.ID, &assessmentDto.CourseID, &assessmentDto.LearningMaterialID, &assessmentDto.RefTable,
					&assessmentDto.CreatedAt, &assessmentDto.UpdatedAt, &assessmentDto.DeletedAt)

				fields := []string{"id"}
				values := []interface{}{&request.ID}
				mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
			},
			ExpectedResponse: "",
			ExpectedError:    errors.NewDBError("db.QueryRow", pgx.ErrNoRows),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(tc.Ctx, tc.MockDB)
			res, err := repo.Upsert(tc.Ctx, tc.MockDB.DB, now, tc.Request.(domain.Assessment))
			if err != nil {
				assert.Equal(t, tc.ExpectedError.Error(), err.Error())
			} else {
				assert.Equal(t, tc.ExpectedResponse, res)
			}
		})
	}
}

func TestAssessmentRepo_GetVirtualByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &AssessmentRepo{}

		assessmentDto := dto.AssessmentExtended{
			ID:                 database.Text("id"),
			CourseID:           database.Text("course_id"),
			LearningMaterialID: database.Text("learning_material_id"),
			RefTable:           database.Varchar("learning_objective"),
			ManualGrading:      database.Bool(true),
		}
		fields, values := assessmentDto.FieldMap()

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Text("id"))
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		expectedResult := domain.Assessment{
			ID:                   "id",
			CourseID:             "course_id",
			LearningMaterialID:   "learning_material_id",
			LearningMaterialType: constants.LearningObjective,
			ManualGrading:        true,
		}

		// actual
		result, err := repo.GetVirtualByID(ctx, mockDB.DB, "id")

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expectedResult, result)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("unexpected error", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &AssessmentRepo{}

		assessmentDto := dto.AssessmentExtended{
			ID:                 database.Text("id"),
			CourseID:           database.Text("course_id"),
			LearningMaterialID: database.Text("learning_material_id"),
			RefTable:           database.Varchar("learning_objective"),
			ManualGrading:      database.Bool(true),
		}
		fields, values := assessmentDto.FieldMap()

		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything, mock.Anything, database.Text("id"))
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		expectedResult := domain.Assessment{}
		expectedErr := errors.NewDBError("database.Select", fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed))

		// actual
		result, err := repo.GetVirtualByID(ctx, mockDB.DB, "id")

		// assert
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, expectedResult, result)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}
