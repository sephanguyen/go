package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jackc/pgtype"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type CourseReaderService struct {
	DB database.Ext
	pb.UnimplementedCourseReaderServiceServer

	StudentPackageAccessPathRepo interface {
		GetByCourseIDAndLocationIDs(ctx context.Context, db database.QueryExecer, courseID pgtype.Text, locationIDs pgtype.TextArray) ([]*entities.StudentPackageAccessPath, error)
	}

	UserMgmtUserReader interface {
		SearchBasicProfile(ctx context.Context, in *upb.SearchBasicProfileRequest, opts ...grpc.CallOption) (*upb.SearchBasicProfileResponse, error)
	}
}

func (s *CourseReaderService) ListStudentByCourse(ctx context.Context, req *pb.ListStudentByCourseRequest) (*pb.ListStudentByCourseResponse, error) {
	headers, ok := metadata.FromIncomingContext(ctx)
	var pkg, token, version string
	if ok {
		pkg = headers["pkg"][0]
		token = headers["token"][0]
		version = headers["version"][0]
	}

	studentPackages, err := s.StudentPackageAccessPathRepo.GetByCourseIDAndLocationIDs(ctx, s.DB, database.Text(req.CourseId), database.TextArray(req.LocationIds))
	if err != nil {
		return nil, fmt.Errorf("s.StudentPackageAccessPathRepo.GetByCourseIDAndLocationIDs: %w", err)
	}

	basicProfileReq := toPbSearchBasicProfileReq(studentPackages, req)

	rsp, err := s.UserMgmtUserReader.SearchBasicProfile(metadata.AppendToOutgoingContext(ctx, "pkg", pkg, "version", version, "token", token), basicProfileReq)
	if err != nil {
		return nil, fmt.Errorf("s.UserMgmtUserReader.SearchBasicProfile: %w", err)
	}

	return &pb.ListStudentByCourseResponse{
		Profiles: rsp.Profiles,
		NextPage: rsp.NextPage,
	}, nil
}

func toPbSearchBasicProfileReq(studentPackages []*entities.StudentPackageAccessPath, req *pb.ListStudentByCourseRequest) *upb.SearchBasicProfileRequest {
	ids := make([]string, len(studentPackages))
	for _, sp := range studentPackages {
		ids = append(ids, sp.StudentID.String)
	}
	return &upb.SearchBasicProfileRequest{
		UserIds:     golibs.Uniq(ids),
		SearchText:  &wrappers.StringValue{Value: req.SearchText},
		Paging:      req.Paging,
		LocationIds: req.LocationIds,
	}
}
