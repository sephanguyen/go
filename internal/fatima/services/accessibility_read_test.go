package services

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/fatima/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCourseService_RetrieveAccessibility(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	studentPackageRepo := &mock_repositories.MockStudentPackageRepo{}

	c := &AccessibilityReadService{
		DB:                 db,
		StudentPackageRepo: studentPackageRepo,
	}

	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	t.Run("err find current package", func(t *testing.T) {
		studentPackageRepo.On("CurrentPackage", ctx, db, database.Text(userID)).Once().
			Return(nil, puddle.ErrClosedPool)

		resp, err := c.RetrieveAccessibility(ctx, &fpb.RetrieveAccessibilityRequest{})
		assert.Nil(t, resp)
		assert.EqualError(t, err, status.Error(codes.Internal, puddle.ErrClosedPool.Error()).Error())
	})

	t.Run("success with 1 package", func(t *testing.T) {
		props := entities.StudentPackageProps{
			CanWatchVideo:     []string{"course_1", "course_2"},
			CanViewStudyGuide: []string{"course_1", "course_3"},
			CanDoQuiz:         []string{"course_2", "course_4"},
		}

		pgProps := pgtype.JSONB{}
		_ = pgProps.Set(props)

		studentPackageRepo.On("CurrentPackage", ctx, db, database.Text(userID)).Once().
			Return([]*entities.StudentPackage{
				{Properties: pgProps},
			}, nil)

		resp, err := c.RetrieveAccessibility(ctx, &fpb.RetrieveAccessibilityRequest{})
		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, &fpb.RetrieveAccessibilityResponse{
			Courses: map[string]*fpb.RetrieveAccessibilityResponse_CourseAccessibility{
				"course_1": {CanWatchVideo: true, CanViewStudyGuide: true, CanDoQuiz: false},
				"course_2": {CanWatchVideo: true, CanViewStudyGuide: false, CanDoQuiz: true},
				"course_3": {CanWatchVideo: false, CanViewStudyGuide: true, CanDoQuiz: false},
				"course_4": {CanWatchVideo: false, CanViewStudyGuide: false, CanDoQuiz: true},
			},
		}, resp)
	})

	t.Run("success with 2 package", func(t *testing.T) {
		prop1 := entities.StudentPackageProps{
			CanWatchVideo:     []string{"course_1", "course_2"},
			CanViewStudyGuide: []string{"course_1", "course_3"},
			CanDoQuiz:         []string{"course_2", "course_4"},
		}

		pgProp1 := pgtype.JSONB{}
		_ = pgProp1.Set(prop1)

		prop2 := entities.StudentPackageProps{
			CanWatchVideo:     []string{"course_1", "course_2"},
			CanViewStudyGuide: []string{"course_1", "course_3"},
			CanDoQuiz:         []string{"course_5", "course_7"},
		}

		pgProp2 := pgtype.JSONB{}
		_ = pgProp2.Set(prop2)

		studentPackageRepo.On("CurrentPackage", ctx, db, database.Text(userID)).Once().
			Return([]*entities.StudentPackage{
				{Properties: pgProp1},
				{Properties: pgProp2},
			}, nil)

		resp, err := c.RetrieveAccessibility(ctx, &fpb.RetrieveAccessibilityRequest{})
		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, &fpb.RetrieveAccessibilityResponse{
			Courses: map[string]*fpb.RetrieveAccessibilityResponse_CourseAccessibility{
				"course_1": {CanWatchVideo: true, CanViewStudyGuide: true, CanDoQuiz: false},
				"course_2": {CanWatchVideo: true, CanViewStudyGuide: false, CanDoQuiz: true},
				"course_3": {CanWatchVideo: false, CanViewStudyGuide: true, CanDoQuiz: false},
				"course_4": {CanWatchVideo: false, CanViewStudyGuide: false, CanDoQuiz: true},
				"course_5": {CanWatchVideo: false, CanViewStudyGuide: false, CanDoQuiz: true},
				"course_7": {CanWatchVideo: false, CanViewStudyGuide: false, CanDoQuiz: true},
			},
		}, resp)
	})
}

func TestCourseService_RetrieveStudentAccessibility(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	studentPackageRepo := &mock_repositories.MockStudentPackageRepo{}

	c := &AccessibilityReadService{
		DB:                 db,
		StudentPackageRepo: studentPackageRepo,
	}

	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	t.Run("no student id", func(t *testing.T) {
		resp, err := c.RetrieveStudentAccessibility(ctx, &fpb.RetrieveStudentAccessibilityRequest{})
		assert.Nil(t, resp)
		assert.EqualError(t, err, status.Error(codes.InvalidArgument, "student id cannot be empty").Error())
	})

	t.Run("err find current package", func(t *testing.T) {
		studentPackageRepo.On("CurrentPackage", ctx, db, database.Text(userID)).Once().
			Return(nil, puddle.ErrClosedPool)

		resp, err := c.RetrieveStudentAccessibility(ctx, &fpb.RetrieveStudentAccessibilityRequest{UserId: userID})
		assert.Nil(t, resp)
		assert.EqualError(t, err, status.Error(codes.Internal, puddle.ErrClosedPool.Error()).Error())
	})

	t.Run("success with 1 package", func(t *testing.T) {
		props := entities.StudentPackageProps{
			CanWatchVideo:     []string{"course_1", "course_2"},
			CanViewStudyGuide: []string{"course_1", "course_3"},
			CanDoQuiz:         []string{"course_2", "course_4"},
		}

		pgProps := pgtype.JSONB{}
		_ = pgProps.Set(props)

		studentPackageRepo.On("CurrentPackage", ctx, db, database.Text(userID)).Once().
			Return([]*entities.StudentPackage{
				{Properties: pgProps},
			}, nil)

		resp, err := c.RetrieveStudentAccessibility(ctx, &fpb.RetrieveStudentAccessibilityRequest{UserId: userID})
		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, &fpb.RetrieveStudentAccessibilityResponse{
			Courses: map[string]*cpb.CourseAccessibility{
				"course_1": {CanWatchVideo: true, CanViewStudyGuide: true, CanDoQuiz: false},
				"course_2": {CanWatchVideo: true, CanViewStudyGuide: false, CanDoQuiz: true},
				"course_3": {CanWatchVideo: false, CanViewStudyGuide: true, CanDoQuiz: false},
				"course_4": {CanWatchVideo: false, CanViewStudyGuide: false, CanDoQuiz: true},
			},
		}, resp)
	})

	t.Run("success with 2 package", func(t *testing.T) {
		prop1 := entities.StudentPackageProps{
			CanWatchVideo:     []string{"course_1", "course_2"},
			CanViewStudyGuide: []string{"course_1", "course_3"},
			CanDoQuiz:         []string{"course_2", "course_4"},
		}

		pgProp1 := pgtype.JSONB{}
		_ = pgProp1.Set(prop1)

		prop2 := entities.StudentPackageProps{
			CanWatchVideo:     []string{"course_1", "course_2"},
			CanViewStudyGuide: []string{"course_1", "course_3"},
			CanDoQuiz:         []string{"course_5", "course_7"},
		}

		pgProp2 := pgtype.JSONB{}
		_ = pgProp2.Set(prop2)

		studentPackageRepo.On("CurrentPackage", ctx, db, database.Text(userID)).Once().
			Return([]*entities.StudentPackage{
				{Properties: pgProp1},
				{Properties: pgProp2},
			}, nil)

		resp, err := c.RetrieveStudentAccessibility(ctx, &fpb.RetrieveStudentAccessibilityRequest{UserId: userID})
		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, &fpb.RetrieveStudentAccessibilityResponse{
			Courses: map[string]*cpb.CourseAccessibility{
				"course_1": {CanWatchVideo: true, CanViewStudyGuide: true, CanDoQuiz: false},
				"course_2": {CanWatchVideo: true, CanViewStudyGuide: false, CanDoQuiz: true},
				"course_3": {CanWatchVideo: false, CanViewStudyGuide: true, CanDoQuiz: false},
				"course_4": {CanWatchVideo: false, CanViewStudyGuide: false, CanDoQuiz: true},
				"course_5": {CanWatchVideo: false, CanViewStudyGuide: false, CanDoQuiz: true},
				"course_7": {CanWatchVideo: false, CanViewStudyGuide: false, CanDoQuiz: true},
			},
		}, resp)
	})
}
