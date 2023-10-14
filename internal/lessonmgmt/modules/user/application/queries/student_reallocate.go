package queries

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudentReallocateQueryHandler struct {
	WrapperConnection       *support.WrapperDBConnection
	StudentSubscriptionRepo infrastructure.StudentSubscriptionRepo
	Env                     string
	UnleashClientIns        unleashclient.ClientInstance
}

type StudentReallocateRequest struct {
	SearchKey  string
	LessonDate time.Time
	Paging     support.Paging[int]
	Filter     domain.Filter
	Timezone   string
}

type StudentReallocateResponse struct {
	ReallocateStudent []*domain.ReallocateStudent
	Total             uint32
}

func (s *StudentReallocateQueryHandler) RetrieveStudentPendingReallocate(ctx context.Context, req StudentReallocateRequest) (StudentReallocateResponse, error) {
	var limit = 25
	if req.Paging.Limit > 0 {
		limit = req.Paging.Limit
	}

	conn, err := s.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return StudentReallocateResponse{}, err
	}
	rs, total, err := s.StudentSubscriptionRepo.RetrieveStudentPendingReallocate(ctx, conn, domain.RetrieveStudentPendingReallocateDto{
		Limit:      limit,
		Offset:     req.Paging.Offset,
		LessonDate: req.LessonDate,
		SearchKey:  req.SearchKey,
		Timezone:   req.Timezone,
		CourseID:   req.Filter.CourseID,
		LocationID: req.Filter.LocationID,
		GradeID:    req.Filter.GradeID,
		ClassID:    req.Filter.ClassId,
		StartDate:  req.Filter.StartDate,
		EndDate:    req.Filter.EndDate,
	})
	resp := StudentReallocateResponse{}
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}
	resp.ReallocateStudent = rs
	resp.Total = total
	return resp, nil
}
