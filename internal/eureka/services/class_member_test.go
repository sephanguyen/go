package services

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	bobproto "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClassMemberService_HandleClassEvent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}

	t.Run("[ActionKind UPSERT] should success", func(t *testing.T) {
		// Arrange
		classStudentRepo := &mock_repositories.MockClassStudentRepo{}
		c := &ClassStudentService{
			DB:               db,
			ClassStudentRepo: classStudentRepo,
		}
		classStudentRepo.On("Upsert", mock.Anything, db, mock.AnythingOfType("*entities.ClassStudent")).
			Run(func(args mock.Arguments) {
				s := args[2].(*entities.ClassStudent)
				assert.Equal(t, "123", s.ClassID.String)
				assert.Equal(t, "userId", s.StudentID.String)
			}).
			Return(nil)

		// Action
		err := c.HandleClassEvent(ctx, &bobproto.EvtClassRoom{
			Message: &bobproto.EvtClassRoom_JoinClass_{
				JoinClass: &bobproto.EvtClassRoom_JoinClass{
					UserId:  "userId",
					ClassId: 123,
				},
			},
		})

		// Assert
		assert.Nil(t, err)
	})

	t.Run("[ActionKind UPSERT] should throw err", func(t *testing.T) {
		// Arrange
		classStudentRepo := &mock_repositories.MockClassStudentRepo{}
		c := &ClassStudentService{
			DB:               db,
			ClassStudentRepo: classStudentRepo,
		}
		classStudentRepo.On("Upsert", mock.Anything, db, mock.AnythingOfType("*entities.ClassStudent")).
			Run(func(args mock.Arguments) {
				s := args[2].(*entities.ClassStudent)
				assert.Equal(t, "123", s.ClassID.String)
				assert.Equal(t, "userId", s.StudentID.String)
			}).
			Return(errors.New("error upsert"))

		// Action
		err := c.HandleClassEvent(ctx, &bobproto.EvtClassRoom{
			Message: &bobproto.EvtClassRoom_JoinClass_{
				JoinClass: &bobproto.EvtClassRoom_JoinClass{
					UserId:  "userId",
					ClassId: 123,
				},
			},
		})
		// Assert
		assert.NotNil(t, err)
		assert.EqualError(t, err, "err s.upsertClassStudent: error upsert")
	})

	t.Run("[ActionKind SOFTDELETE] should success", func(t *testing.T) {
		// Arrange
		classStudentRepo := &mock_repositories.MockClassStudentRepo{}
		c := &ClassStudentService{
			DB:               db,
			ClassStudentRepo: classStudentRepo,
		}
		userIds := []string{"userId", "userId2", "userId3"}
		classStudentRepo.On("SoftDelete", mock.Anything, db, mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string")).
			Run(func(args mock.Arguments) {
				studentIds := args[2].([]string)
				classIds := args[3].([]string)
				assert.Contains(t, studentIds, userIds[rand.Intn(3)])
				assert.Contains(t, classIds, "123")
				assert.Equal(t, len(userIds), len(studentIds))
				assert.Equal(t, 1, len(classIds))
			}).
			Return(nil)

		// Action
		err := c.HandleClassEvent(ctx, &bobproto.EvtClassRoom{
			Message: &bobproto.EvtClassRoom_LeaveClass_{
				LeaveClass: &bobproto.EvtClassRoom_LeaveClass{
					UserIds: userIds,
					ClassId: 123,
				},
			},
		})
		// Assert
		assert.Nil(t, err)
	})

	t.Run("[ActionKind SOFTDELETE] should throw error", func(t *testing.T) {
		// Arrange
		classStudentRepo := &mock_repositories.MockClassStudentRepo{}
		c := &ClassStudentService{
			DB:               db,
			ClassStudentRepo: classStudentRepo,
		}
		userIds := []string{"userId", "userId2", "userId3"}
		classStudentRepo.On("SoftDelete", mock.Anything, db, mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string")).
			Run(func(args mock.Arguments) {
				studentIds := args[2].([]string)
				classIds := args[3].([]string)
				assert.Contains(t, studentIds, userIds[rand.Intn(3)])
				assert.Contains(t, classIds, "123")
				assert.Equal(t, len(userIds), len(studentIds))
				assert.Equal(t, 1, len(classIds))
			}).
			Return(errors.New("error softdelete"))

		// Action
		err := c.HandleClassEvent(ctx, &bobproto.EvtClassRoom{
			Message: &bobproto.EvtClassRoom_LeaveClass_{
				LeaveClass: &bobproto.EvtClassRoom_LeaveClass{
					UserIds: userIds,
					ClassId: 123,
				},
			},
		})
		// Assert
		assert.NotNil(t, err)
		assert.EqualError(t, err, "err s.softDeleteClassMember: error softdelete")
	})
}
