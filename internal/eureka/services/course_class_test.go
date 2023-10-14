package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCourseClassService_SyncCourseClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}

	t.Run("[ActionKind UPSERT] should create CourseClass success", func(t *testing.T) {
		// Arrange
		courseClassRepo := &mock_repositories.MockCourseClassRepo{}
		c := &CourseClassService{
			DB:              db,
			CourseClassRepo: courseClassRepo,
		}
		courseIDs := []string{"courseID1", "courseID2"}
		courseClassRepo.On("BulkUpsert", mock.Anything, db, mock.AnythingOfType("[]*entities.CourseClass")).
			Run(func(args mock.Arguments) {
				s := args[2].([]*entities.CourseClass)
				assert.Contains(t, courseIDs, s[0].CourseID.String)
				assert.Equal(t, s[0].ClassID.String, "1")
			}).
			Return(nil)

		// Action
		err := c.SyncCourseClass(ctx, &npb.EventMasterRegistration{
			Classes: []*npb.EventMasterRegistration_Class{
				{
					ClassId:    1,
					CourseId:   courseIDs[0],
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
				},
			},
		})

		// Assert
		assert.Nil(t, err)
	})

	t.Run("[ActionKind UPSERT] should throw err", func(t *testing.T) {
		// Arrange
		courseClassRepo := &mock_repositories.MockCourseClassRepo{}
		c := &CourseClassService{
			DB:              db,
			CourseClassRepo: courseClassRepo,
		}
		courseIDs := []string{"courseID1", "courseID2"}
		courseClassRepo.On("BulkUpsert", mock.Anything, db, mock.AnythingOfType("[]*entities.CourseClass")).
			Run(func(args mock.Arguments) {
				s := args[2].([]*entities.CourseClass)
				assert.Contains(t, courseIDs, s[0].CourseID.String)
				assert.Equal(t, s[0].ClassID.String, "1")
			}).
			Return(errors.New("error insert"))

		// Action
		err := c.SyncCourseClass(ctx, &npb.EventMasterRegistration{
			Classes: []*npb.EventMasterRegistration_Class{
				{
					ClassId:    1,
					CourseId:   courseIDs[0],
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
				},
			},
		})

		// Assert
		assert.NotNil(t, err)
		assert.EqualError(t, err, "err s.CourseClassRepo.BulkUpsert: error insert")
	})

	t.Run("[ActionKind DELETE] should soft delete CourseClass success", func(t *testing.T) {
		// Arrange
		courseClassRepo := &mock_repositories.MockCourseClassRepo{}
		c := &CourseClassService{
			DB:              db,
			CourseClassRepo: courseClassRepo,
		}
		courseIDs := []string{"courseID1", "courseID2"}
		classIDs := []string{"1", "2"}
		courseClassRepo.On("Delete", mock.Anything, db, mock.AnythingOfType("[]*entities.CourseClass")).
			Run(func(args mock.Arguments) {
				inputs := args[2].([]*entities.CourseClass)
				for i, c := range inputs {
					assert.Equal(t, courseIDs[i], c.CourseID.String)
					assert.Equal(t, classIDs[i], c.ClassID.String)
				}
			}).
			Return(nil)

		// Action
		err := c.SyncCourseClass(ctx, &npb.EventMasterRegistration{
			Classes: []*npb.EventMasterRegistration_Class{
				{
					ClassId:    uint64(1),
					CourseId:   courseIDs[0],
					ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
				},
				{
					ClassId:    uint64(2),
					CourseId:   courseIDs[1],
					ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
				},
			},
		})

		// Assert
		assert.Nil(t, err)
	})

	t.Run("[ActionKind DELETE] should throw err", func(t *testing.T) {
		// Arrange
		courseClassRepo := &mock_repositories.MockCourseClassRepo{}
		c := &CourseClassService{
			DB:              db,
			CourseClassRepo: courseClassRepo,
		}
		courseIDs := []string{"courseID1", "courseID2"}
		classIDs := []string{"1", "2"}
		courseClassRepo.On("Delete", mock.Anything, db, mock.AnythingOfType("[]*entities.CourseClass")).
			Run(func(args mock.Arguments) {
				inputs := args[2].([]*entities.CourseClass)
				for i, c := range inputs {
					assert.Equal(t, courseIDs[i], c.CourseID.String)
					assert.Equal(t, classIDs[i], c.ClassID.String)
				}
			}).
			Return(errors.New("error softdelete"))

		// Action
		err := c.SyncCourseClass(ctx, &npb.EventMasterRegistration{
			Classes: []*npb.EventMasterRegistration_Class{
				{
					ClassId:    uint64(1),
					CourseId:   courseIDs[0],
					ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
				},
				{
					ClassId:    uint64(2),
					CourseId:   courseIDs[1],
					ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
				},
			},
		})

		// Assert
		assert.NotNil(t, err)
		assert.EqualError(t, err, "err s.CourseClassRepo.Delete: error softdelete")
	})
}
