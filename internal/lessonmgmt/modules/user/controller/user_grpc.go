package controller

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserGRPCService struct {
	bobDB             database.Ext
	wrapperConnection *support.WrapperDBConnection
	teacherRepo       infrastructure.TeacherRepo
	userRepo          infrastructure.UserRepo
	userBasicInfoRepo infrastructure.UserBasicInfoRepo
}

func NewUserGRPCService(bobDB database.Ext, wrapperConnection *support.WrapperDBConnection, teacherRepo infrastructure.TeacherRepo, userRepo infrastructure.UserRepo, userBasicInfoRepo infrastructure.UserBasicInfoRepo) *UserGRPCService {
	return &UserGRPCService{
		bobDB:             bobDB,
		wrapperConnection: wrapperConnection,
		teacherRepo:       teacherRepo,
		userRepo:          userRepo,
		userBasicInfoRepo: userBasicInfoRepo,
	}
}

func (u *UserGRPCService) GetStudentsManyReferenceByNameOrEmail(ctx context.Context, req *lpb.GetStudentsManyReferenceByNameOrEmailRequest) (*lpb.GetStudentsManyReferenceByNameOrEmailResponse, error) {
	students, err := u.userRepo.GetStudentsManyReferenceByNameOrEmail(ctx, u.bobDB, req.Keyword, req.Limit, req.Offset)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf(`userRepo.GetStudentsManyReferenceByNameOrEmail: %v`, err))
	}
	res := &lpb.GetStudentsManyReferenceByNameOrEmailResponse{}
	res.Students = make([]*lpb.GetStudentsManyReferenceByNameOrEmailResponse_StudentInfo, 0, len(students))
	for _, value := range students {
		res.Students = append(res.Students, &lpb.GetStudentsManyReferenceByNameOrEmailResponse_StudentInfo{
			UserId: value.ID,
			Name:   value.Name,
			Email:  value.Email,
		})
	}
	return res, nil
}

func (u *UserGRPCService) GetTeachers(ctx context.Context, req *lpb.GetTeachersRequest) (*lpb.GetTeachersResponse, error) {
	conn, err := u.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	teachers, err := u.teacherRepo.ListByIDs(ctx, conn, req.TeacherIds)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf(`teacherRepo.ListByIDs: %v`, err))
	}
	if err = teachers.IsValid(); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf(`invalid teachers: %v`, err))
	}

	res := &lpb.GetTeachersResponse{}
	res.Teachers = make([]*lpb.GetTeachersResponse_TeacherInfo, 0, len(teachers))
	for _, teacher := range teachers {
		res.Teachers = append(res.Teachers, &lpb.GetTeachersResponse_TeacherInfo{
			Id:        teacher.ID,
			Name:      teacher.Name,
			CreatedAt: timestamppb.New(teacher.CreatedAt),
			UpdatedAt: timestamppb.New(teacher.UpdatedAt),
		})
	}

	return res, nil
}

func (u *UserGRPCService) GetUserGroup(ctx context.Context, req *lpb.GetUserGroupRequest) (*lpb.GetUserGroupResponse, error) {
	conn, err := u.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	gr, err := u.userRepo.GetUserGroupByUserID(ctx, conn, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf(`userRepo.GetUserGroupByUserIDs: %v`, err))
	}

	return &lpb.GetUserGroupResponse{UserGroup: gr}, nil
}

func (u *UserGRPCService) GetTeachersSameGrantedLocation(ctx context.Context, req *lpb.GetTeachersSameGrantedLocationRequest) (*lpb.GetTeachersSameGrantedLocationResponse, error) {
	limit := int(req.Paging.GetLimit())
	if limit == 0 {
		limit = 30
	}
	query := domain.UserBasicInfoQuery{
		KeyWord: req.Keyword,
		Offset:  int(req.Paging.GetOffsetInteger()),
		Limit:   support.Min(limit, 100),
	}
	if !req.IsAllTeacher {
		query.LocationID = req.LocationId
	}
	conn, err := u.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	data, err := u.userBasicInfoRepo.GetTeachersSameGrantedLocation(ctx, conn, query)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf(`userRepo.GetTeachersSameGrantedLocation: %v`, err))
	}

	return &lpb.GetTeachersSameGrantedLocationResponse{
		Teachers: sliceutils.Map(data, func(userInfo *domain.UserBasicInfo) *lpb.GetTeachersSameGrantedLocationResponse_TeacherInfo {
			return &lpb.GetTeachersSameGrantedLocationResponse_TeacherInfo{
				Id:                userInfo.UserID,
				Name:              userInfo.FullName,
				FirstName:         userInfo.FirstName,
				LastName:          userInfo.LastName,
				FullNamePhonetic:  userInfo.FullNamePhonetic,
				FirstNamePhonetic: userInfo.FirstNamePhonetic,
				LastNamePhonetic:  userInfo.LastNamePhonetic,
				Email:             userInfo.Email,
				CreatedAt:         timestamppb.New(userInfo.CreatedAt),
				UpdatedAt:         timestamppb.New(userInfo.UpdatedAt),
			}
		}),
	}, nil
}
