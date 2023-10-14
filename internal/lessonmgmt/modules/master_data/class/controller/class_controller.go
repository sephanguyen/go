package controller

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ClassReaderService struct {
	ClassUseCase      domain.ClassUseCase
	WrapperConnection *support.WrapperDBConnection
}

func (c *ClassReaderService) GetByStudentSubscription(ctx context.Context, in *lpb.GetByStudentSubscriptionRequest) (*lpb.GetByStudentSubscriptionResponse, error) {
	conn, err := c.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	studentSubWithClassUnassigned, err := c.ClassUseCase.GetByStudentSubscription(ctx, conn, in.GetStudentSubscriptionId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	classUnassigned := make([]*lpb.GetByStudentSubscriptionResponse_ClassUnassigned, 0, len(studentSubWithClassUnassigned))
	for _, ss := range studentSubWithClassUnassigned {
		classUnassigned = append(classUnassigned, &lpb.GetByStudentSubscriptionResponse_ClassUnassigned{
			StudentSubscriptionId: ss.StudentSubscriptionID,
			IsClassUnassigned:     ss.IsClassUnAssigned,
		})
	}
	return &lpb.GetByStudentSubscriptionResponse{
		ClassUnassigned: classUnassigned,
	}, nil
}
