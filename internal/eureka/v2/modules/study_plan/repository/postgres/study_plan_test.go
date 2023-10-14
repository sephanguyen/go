package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStudyPlanRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	repo := &StudyPlanRepo{}

	now := time.Now()
	request := domain.StudyPlan{
		ID:           "id",
		CourseID:     "course_id",
		Name:         "name",
		AcademicYear: "academic_year_id",
		Status:       domain.StudyPlanStatusActive,
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
				studyPlanDto := dto.StudyPlan{
					BaseEntity: dto.BaseEntity{
						CreatedAt: database.Timestamptz(now),
						UpdatedAt: database.Timestamptz(now),
						DeletedAt: pgtype.Timestamptz{
							Status: pgtype.Null,
						},
					},
					ID:           database.Text(request.ID),
					CourseID:     database.Text(request.CourseID),
					Name:         database.Text(request.Name),
					AcademicYear: database.Text(request.AcademicYear),
					Status:       database.Text(string(request.Status)),
				}
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything,
					&studyPlanDto.ID, &studyPlanDto.Name, &studyPlanDto.CourseID, &studyPlanDto.AcademicYear, &studyPlanDto.Status,
					&studyPlanDto.CreatedAt, &studyPlanDto.UpdatedAt, &studyPlanDto.DeletedAt)

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
				studyPlanDto := dto.StudyPlan{
					BaseEntity: dto.BaseEntity{
						CreatedAt: database.Timestamptz(now),
						UpdatedAt: database.Timestamptz(now),
						DeletedAt: pgtype.Timestamptz{
							Status: pgtype.Null,
						},
					},
					ID:           database.Text(request.ID),
					CourseID:     database.Text(request.CourseID),
					Name:         database.Text(request.Name),
					AcademicYear: database.Text(request.AcademicYear),
					Status:       database.Text(string(request.Status)),
				}
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything,
					&studyPlanDto.ID, &studyPlanDto.Name, &studyPlanDto.CourseID, &studyPlanDto.AcademicYear, &studyPlanDto.Status,
					&studyPlanDto.CreatedAt, &studyPlanDto.UpdatedAt, &studyPlanDto.DeletedAt)

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
			res, err := repo.Upsert(tc.Ctx, tc.MockDB.DB, now, tc.Request.(domain.StudyPlan))
			if err != nil {
				assert.Equal(t, tc.ExpectedError.Error(), err.Error())
			} else {
				assert.Equal(t, tc.ExpectedResponse, res)
			}
		})
	}
}
