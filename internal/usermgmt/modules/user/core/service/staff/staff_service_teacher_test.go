package staff

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	usvc "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStaffService_UserToTeacher(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	type params struct {
		user      *entity.LegacyUser
		schoolIDs []int32
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: params{
				user: &entity.LegacyUser{
					ID:           pgtype.Text{String: "1", Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: "constants.ManabieOrgLocation", Status: pgtype.Present},
				},
				schoolIDs: []int32{constants.ManabieSchool},
			},
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		teacher := userToTeacher(
			testCase.req.(params).user,
			testCase.req.(params).schoolIDs,
		)

		if testCase.expectedErr != nil {
			assert.NotEmpty(t, teacher)
		}
	}
}

func TestStaffService_CreateTeacher(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx := new(mock_database.Tx)

	teacherRepo := new(mock_repositories.MockTeacherRepo)

	umsvc := &usvc.UserModifierService{
		DB:          tx,
		TeacherRepo: teacherRepo,
	}

	s := &StaffService{
		DB:                  umsvc.DB,
		FirebaseClient:      umsvc.FirebaseClient,
		UserModifierService: umsvc,
	}
	type params struct {
		user      *entity.LegacyUser
		schoolIDs []int64
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: params{
				user: &entity.LegacyUser{
					ID:           pgtype.Text{String: "1", Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: "constants.ManabieOrgLocation", Status: pgtype.Present},
				},
				schoolIDs: []int64{constants.ManabieSchool},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				teacherRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		if testCase.setup != nil {
			testCase.setup(testCase.ctx)
		}

		err := s.createTeacher(ctx, tx, testCase.req.(params).user, testCase.req.(params).schoolIDs)

		if testCase.expectedErr != nil {
			assert.EqualError(t, err, testCase.expectedErr.Error())
		}
	}
}
