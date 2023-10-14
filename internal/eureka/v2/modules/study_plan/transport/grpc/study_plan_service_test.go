package grpc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_usecase "github.com/manabie-com/backend/mock/eureka/v2/modules/study_plan/usecase"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStudyPlanService_UpsertStudyPlan(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	courseID := idutil.ULIDNow()
	name := idutil.ULIDNow()
	academicYear := "academic_year_id"
	spStatus := pb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE
	req := &pb.UpsertStudyPlanRequest{
		CourseId:     courseID,
		Name:         name,
		AcademicYear: academicYear,
		Status:       spStatus,
	}

	t.Run("Return invalid arguments when course id is missing", func(t *testing.T) {
		t.Parallel()
		// arrange
		studyPlanUsecase := &mock_usecase.MockStudyPlanUsecaseImpl{}
		sut := NewStudyPlanService(studyPlanUsecase)
		req := &pb.UpsertStudyPlanRequest{
			CourseId:     "",
			Name:         "name",
			AcademicYear: "academic_year_id",
		}
		rootErr := errors.NewValidationError("req must have course id", nil)
		expectedErr := status.Error(codes.InvalidArgument, fmt.Errorf("%w", rootErr).Error())

		// act
		resp, err := sut.UpsertStudyPlan(ctx, req)

		// assert
		assert.Equal(t, err, expectedErr)

		assert.Nil(t, resp)
	})
	t.Run("Return invalid arguments when name is missing", func(t *testing.T) {
		t.Parallel()
		// arrange
		studyPlanUsecase := &mock_usecase.MockStudyPlanUsecaseImpl{}
		sut := NewStudyPlanService(studyPlanUsecase)
		req := &pb.UpsertStudyPlanRequest{
			CourseId:     courseID,
			Name:         "",
			AcademicYear: "academic_year_id",
		}
		rootErr := errors.NewValidationError("req must have name", nil)
		expectedErr := status.Error(codes.InvalidArgument, fmt.Errorf("%w", rootErr).Error())

		// act
		resp, err := sut.UpsertStudyPlan(ctx, req)

		// assert
		assert.Equal(t, err, expectedErr)

		assert.Nil(t, resp)
	})

	t.Run("Return invalid arguments when academic year is missing", func(t *testing.T) {
		t.Parallel()
		// arrange
		studyPlanUsecase := &mock_usecase.MockStudyPlanUsecaseImpl{}
		sut := NewStudyPlanService(studyPlanUsecase)
		req := &pb.UpsertStudyPlanRequest{
			CourseId:     courseID,
			Name:         "name",
			AcademicYear: "",
		}
		rootErr := errors.NewValidationError("req must have academic year", nil)
		expectedErr := status.Error(codes.InvalidArgument, fmt.Errorf("%w", rootErr).Error())

		// act
		resp, err := sut.UpsertStudyPlan(ctx, req)

		// assert
		assert.Equal(t, err, expectedErr)

		assert.Nil(t, resp)
	})

	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		// arrange
		studyPlanUsecase := &mock_usecase.MockStudyPlanUsecaseImpl{}
		sut := NewStudyPlanService(studyPlanUsecase)
		studyPlanUsecase.On("UpsertStudyPlan", ctx, mock.Anything).Once().Return("study_plan_id", nil)

		// act
		resp, err := sut.UpsertStudyPlan(ctx, req)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, resp.StudyPlanId, "study_plan_id")
	})

}
