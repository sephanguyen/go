package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lesson_user_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure/repo"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mọck_lesson_user_repositories "github.com/manabie-com/backend/mock/lessonmgmt/user/repositories"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdaterLessonMember_UpdateLessonMemberNames(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	userIds := pgtype.TextArray{}
	_ = userIds.Set([]string{"user-1"})
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)
	userRepo := new(mọck_lesson_user_repositories.MockUserRepo)
	tcs := []struct {
		name     string
		req      *domain.LessonMembers
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "success",
			req: &domain.LessonMembers{{
				LessonID:  "lesson-1",
				StudentID: "user-1",
			}},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				userRepo.
					On("Retrieve", ctx, db, userIds, mock.Anything).
					Return([]*lesson_user_repo.User{
						{
							ID:        database.Text("user-1"),
							FirstName: database.Text("first name"),
							LastName:  database.Text("last name"),
						},
					}, nil).
					Once()
				updateLessonMembers := []*domain.UpdateLessonMemberName{{
					LessonID:      "lesson-1",
					StudentID:     "user-1",
					UserFirstName: "first name",
					UserLastName:  "last name",
				}}
				lessonMemberRepo.
					On("UpdateLessonMemberNames", ctx, db, updateLessonMembers).
					Return(nil).
					Once()

			},
		},
		{
			name: "failed",
			req: &domain.LessonMembers{{
				LessonID:  "lesson-1",
				StudentID: "user-1",
			}},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				userRepo.
					On("Retrieve", ctx, db, userIds, mock.Anything).
					Return(nil, errors.New("Internal Error")).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMemberNames", ctx, db, mock.Anything).
					Return(errors.New("Internal Error")).
					Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: "1",
				},
			}
			ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			tc.setup(ctx)

			service := UpdaterLessonMember{
				DB:               db,
				LessonMemberRepo: lessonMemberRepo,
				UserRepo:         userRepo,
			}
			err := service.UpdateLessonMemberNames(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

		})
	}
}
