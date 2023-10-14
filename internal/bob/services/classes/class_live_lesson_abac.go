package classes

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ClassServiceABAC struct {
	*ClassService
}

func (rcv *ClassServiceABAC) TeacherRetrieveStreamToken(ctx context.Context, req *pb.TeacherRetrieveStreamTokenRequest) (*pb.TeacherRetrieveStreamTokenResponse, error) {
	return rcv.ClassService.TeacherRetrieveStreamToken(ctx, req)
}

func (rcv *ClassServiceABAC) StudentRetrieveStreamToken(ctx context.Context, req *pb.StudentRetrieveStreamTokenRequest) (*pb.StudentRetrieveStreamTokenResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)

	allowed, err := rcv.StreamSubscriberPermission(ctx, req.LessonId, userID)
	if err != nil {
		return nil, fmt.Errorf("ClassServiceABAC.StudentRetrieveStreamToken: %w", err)
	}
	if !allowed {
		return nil, status.Error(codes.PermissionDenied, "student is not assigned to course")
	}

	return rcv.ClassService.StudentRetrieveStreamToken(ctx, req)
}

func (rcv *ClassServiceABAC) JoinLesson(ctx context.Context, req *pb.JoinLessonRequest) (*pb.JoinLessonResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)
	userGroup, err := rcv.UserRepo.UserGroup(ctx, rcv.DB, database.Text(userID))
	if err != nil {
		return nil, fmt.Errorf("rcv.UserRepo.UserGroup: %w", err)
	}
	if userGroup == entities.UserGroupStudent {
		studentSubscribePermission, err := rcv.StreamSubscriberPermission(ctx, req.LessonId, userID)
		if err != nil {
			return nil, err
		}
		if !studentSubscribePermission {
			return nil, status.Error(codes.PermissionDenied, "student not allowed to join lesson")
		}
	}

	return rcv.ClassService.JoinLesson(ctx, req)
}

func (rcv *ClassServiceABAC) EndLiveLesson(ctx context.Context, req *pb.EndLiveLessonRequest) (*pb.EndLiveLessonResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)
	allowed, err := rcv.ClassService.LiveTeacherPermission(ctx, req.LessonId, userID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, status.Error(codes.PermissionDenied, "teacher is not lecturer")
	}
	return rcv.ClassService.EndLiveLesson(ctx, req)
}

func (rcv *ClassServiceABAC) LeaveLesson(ctx context.Context, req *pb.LeaveLessonRequest) (*pb.LeaveLessonResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)
	userGroup, err := rcv.UserRepo.UserGroup(ctx, rcv.DB, database.Text(userID))
	if err != nil {
		return nil, fmt.Errorf("rcv.UserRepo.UserGroup: %w", err)
	}
	if userGroup == entities.UserGroupStudent && req.UserId != userID {
		return nil, status.Error(codes.PermissionDenied, "student can only leave by him self")
	}

	return rcv.ClassService.LeaveLesson(ctx, req)
}

func (rcv *ClassServiceABAC) UpsertMedia(ctx context.Context, req *pb.UpsertMediaRequest) (*pb.UpsertMediaResponse, error) {
	return rcv.ClassService.UpsertMedia(ctx, req)
}

func (rcv *ClassServiceABAC) RetrieveMedia(ctx context.Context, req *pb.RetrieveMediaRequest) (*pb.RetrieveMediaResponse, error) {
	return rcv.ClassService.RetrieveMedia(ctx, req)
}
