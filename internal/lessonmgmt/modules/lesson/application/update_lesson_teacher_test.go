package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mọck_lesson_user_repositories "github.com/manabie-com/backend/mock/lessonmgmt/user/repositories"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdaterLessonTeacher_UpdateLessonTeacherNames(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	userIds := pgtype.TextArray{}
	_ = userIds.Set([]string{"user-1"})
	lessonTeacherRepo := new(mock_repositories.MockLessonTeacherRepo)
	userRepo := new(mọck_lesson_user_repositories.MockUserRepo)
	tcs := []struct {
		name     string
		req      user_domain.Users
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "success",
			req: user_domain.Users{{
				ID: "user_1", FullName: "full name"}},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				updateLessonTeachers := []*domain.UpdateLessonTeacherName{{
					TeacherID: "user_1",
					FullName:  "full name",
				}}
				lessonTeacherRepo.
					On("UpdateLessonTeacherNames", ctx, db, updateLessonTeachers).
					Return(nil).
					Once()

			},
		},
		{
			name: "failed",
			req: user_domain.Users{{
				ID: "user_1",
			}},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				userRepo.
					On("Retrieve", ctx, db, userIds, mock.Anything).
					Return(nil, errors.New("Internal Error")).
					Once()
				lessonTeacherRepo.
					On("UpdateLessonTeacherNames", ctx, db, mock.Anything).
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

			service := UpdaterLessonTeacher{
				DB:                db,
				LessonTeacherRepo: lessonTeacherRepo,
			}
			err := service.UpdaterLessonTeacherNames(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

		})
	}
}
