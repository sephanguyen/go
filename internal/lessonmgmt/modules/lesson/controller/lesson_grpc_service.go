package controller

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LessonManagementGRPCService struct {
	wrapperConnection *support.WrapperDBConnection

	userModulePort  infrastructure.UserModulePort
	mediaModulePort infrastructure.MediaModulePort
	lessonRepo      infrastructure.LessonRepo
	roomStateRepo   infrastructure.LessonRoomState
}

func NewLessonManagementGRPCService(
	wrapperConnection *support.WrapperDBConnection,
	lessonRepo infrastructure.LessonRepo,
	userModulePort infrastructure.UserModulePort,
	mediaModulePort infrastructure.MediaModulePort,
	roomStateRepo infrastructure.LessonRoomState,
) *LessonManagementGRPCService {
	return &LessonManagementGRPCService{
		wrapperConnection: wrapperConnection,
		userModulePort:    userModulePort,
		mediaModulePort:   mediaModulePort,
		lessonRepo:        lessonRepo,
		roomStateRepo:     roomStateRepo,
	}
}

func (l *LessonManagementGRPCService) CreateLesson(ctx context.Context, req *bpb.CreateLessonRequest) (*bpb.CreateLessonResponse, error) {
	// TODO: will remove when endpoint in lessonmgmt is stable
	return &bpb.CreateLessonResponse{}, nil
}

func (l *LessonManagementGRPCService) getMediaIDsFromMaterials(materials []*bpb.Material) ([]string, error) {
	if len(materials) == 0 {
		return nil, nil
	}
	res := make([]string, 0, len(materials))
	for _, material := range materials {
		switch resource := material.Resource.(type) {
		case *bpb.Material_BrightcoveVideo_:
			return nil, fmt.Errorf("not yet support this type")
		case *bpb.Material_MediaId:
			res = append(res, resource.MediaId)
		default:
			return nil, fmt.Errorf(`unexpected material's type %T`, resource)
		}
	}

	return res, nil
}

func (l *LessonManagementGRPCService) RetrieveLessons(ctx context.Context, req *bpb.RetrieveLessonsRequestV2) (*bpb.RetrieveLessonsResponseV2, error) {
	// TODO: implement this method later
	return &bpb.RetrieveLessonsResponseV2{}, nil
}

func (l *LessonManagementGRPCService) UpdateLesson(ctx context.Context, req *bpb.UpdateLessonRequest) (*bpb.UpdateLessonResponse, error) {
	return &bpb.UpdateLessonResponse{}, nil
}

func (l *LessonManagementGRPCService) DeleteLesson(context.Context, *bpb.DeleteLessonRequest) (*bpb.DeleteLessonResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteLesson not implemented")
}

func (l *LessonManagementGRPCService) ModifyLiveLessonState(ctx context.Context, req *bpb.ModifyLiveLessonStateRequest) (*bpb.ModifyLiveLessonStateResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)
	var command application.StateModifyCommand
	switch c := req.Command.(type) {
	case *bpb.ModifyLiveLessonStateRequest_ShareAMaterial:
		t := &application.ShareMaterialCommand{
			CommanderID: userID,
			LessonID:    req.Id,
			MediaID:     c.ShareAMaterial.MediaId,
			VideoState:  nil,
		}

		switch c.ShareAMaterial.State.(type) {
		case *bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand_VideoState:
			tplState := req.GetShareAMaterial().GetVideoState()
			t.VideoState = &domain.VideoState{
				PlayerState: domain.PlayerState(tplState.PlayerState.String()),
			}
			if tplState.CurrentTime != nil {
				t.VideoState.CurrentTime = domain.Duration(tplState.CurrentTime.AsDuration())
			}
		case *bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand_PdfState:
			break
		}
		command = t
	case *bpb.ModifyLiveLessonStateRequest_StopSharingMaterial:
		command = &application.StopSharingMaterialCommand{
			CommanderID: userID,
			LessonID:    req.Id,
		}
	default:
		return nil, status.Error(codes.Internal, fmt.Sprintf("unhandled state type %T", req.Command))
	}

	permissionChecker := &application.RoomStateCommandPermissionChecker{
		WrapperConnection: l.wrapperConnection,
		UserModule:        l.userModulePort,
	}
	if err := permissionChecker.Execute(ctx, command); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	commandDp := application.RoomStateCommandDispatcher{
		WrapperConnection: l.wrapperConnection,
		LessonRepo:        l.lessonRepo,
		MediaModulePort:   l.mediaModulePort,
		RoomStateRepo:     l.roomStateRepo,
	}
	if err := commandDp.Execute(ctx, command); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &bpb.ModifyLiveLessonStateResponse{}, nil
}
