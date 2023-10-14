package controller

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	commands "github.com/manabie-com/backend/internal/mastermgmt/modules/class/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type ClassService struct {
	DB  database.Ext
	JSM nats.JetStreamManagement

	ClassCommandHandler       commands.ClassCommandHandler
	ClassQueryHandler         queries.ClassQueryHandler
	ExportClassesQueryHandler queries.ExportClassesQueryHandler

	ClassRepo    infrastructure.ClassRepo
	LocationRepo infrastructure.LocationRepo
	CourseRepo   infrastructure.CourseRepo
}

func (c *ClassService) ImportClass(ctx context.Context, req *mpb.ImportClassRequest) (*mpb.ImportClassResponse, error) {
	sc := scanner.NewCSVScanner(bytes.NewReader(req.Payload))
	if len(sc.GetRow()) == 0 {
		return &mpb.ImportClassResponse{}, status.Error(codes.InvalidArgument, "no data in csv file")
	}
	classes := []*domain.Class{}
	errors := []*mpb.ImportClassResponse_ImportClassError{}
	now := time.Now()
	for sc.Scan() {
		classBuilder := domain.NewClass().
			WithClassRepo(c.ClassRepo).
			WithLocationRepo(c.LocationRepo).
			WithCourseRepo(c.CourseRepo).
			WithName(sc.Text("class_name")).
			WithCourseID(sc.Text("course_id")).
			WithLocationID(sc.Text("location_id")).
			WithSchoolID(golibs.ResourcePathFromCtx(ctx)).
			WithModificationTime(now, now)
		class, err := classBuilder.Build(ctx, c.DB)
		if err != nil {
			icErr := &mpb.ImportClassResponse_ImportClassError{
				RowNumber: int32(sc.GetCurRow()),
				Error:     err.Error(),
			}
			errors = append(errors, icErr)
			continue
		}
		classes = append(classes, class)
	}
	if len(classes) == 0 && len(errors) == 0 {
		return &mpb.ImportClassResponse{}, status.Error(codes.InvalidArgument, "no data in csv file")
	}
	if len(errors) > 0 {
		return &mpb.ImportClassResponse{
			Errors: errors,
		}, nil
	}
	var err error
	if len(classes) > 0 {
		payload := commands.CreateClass{Classes: classes}
		err = c.ClassCommandHandler.Create(ctx, payload)
	}
	if err != nil {
		return &mpb.ImportClassResponse{
			Errors: errors,
		}, status.Errorf(codes.Internal, err.Error())
	}
	for _, class := range classes {
		if err := c.PublishClassEvent(ctx, &mpb.EvtClass{
			Message: &mpb.EvtClass_CreateClass_{
				CreateClass: &mpb.EvtClass_CreateClass{
					ClassId:    class.ClassID,
					Name:       class.Name,
					CourseId:   class.CourseID,
					LocationId: class.LocationID,
				},
			},
		}); err != nil {
			ctxzap.Extract(ctx).Warn("ClassService.PublishClassEvent", zap.Error(err))
		}
	}
	resp := &mpb.ImportClassResponse{
		Errors: errors,
	}
	return resp, nil
}

func (c *ClassService) ExportClasses(ctx context.Context, req *mpb.ExportClassesRequest) (res *mpb.ExportClassesResponse, err error) {
	bytes, err := c.ExportClassesQueryHandler.ExportClasses(ctx)
	if err != nil {
		return &mpb.ExportClassesResponse{}, err
	}
	res = &mpb.ExportClassesResponse{
		Data: bytes,
	}
	return res, nil
}

func (c *ClassService) UpdateClass(ctx context.Context, req *mpb.UpdateClassRequest) (*mpb.UpdateClassResponse, error) {
	classID := req.GetClassId()
	resp := &mpb.UpdateClassResponse{}
	if len(classID) == 0 {
		return resp, status.Error(codes.InvalidArgument, "`class_id` cannot be empty")
	}
	payload := commands.UpdateClassById{ID: classID, Name: req.GetName()}
	err := c.ClassCommandHandler.UpdateByID(ctx, payload)
	if err == domain.ErrNotFound {
		return resp, status.Error(codes.NotFound, "`class_id` not found")
	} else if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}
	if err := c.PublishClassEvent(ctx, &mpb.EvtClass{
		Message: &mpb.EvtClass_UpdateClass_{
			UpdateClass: &mpb.EvtClass_UpdateClass{
				ClassId: classID,
				Name:    req.GetName(),
			},
		},
	}); err != nil {
		ctxzap.Extract(ctx).Warn("ClassService.PublishClassEvent", zap.Error(err))
	}
	return resp, nil
}

func (c *ClassService) DeleteClass(ctx context.Context, req *mpb.DeleteClassRequest) (*mpb.DeleteClassResponse, error) {
	classID := req.GetClassId()
	resp := &mpb.DeleteClassResponse{}
	if len(classID) == 0 {
		return resp, status.Error(codes.InvalidArgument, "`class_id` cannot be empty")
	}
	payload := commands.DeleteClassById{ID: classID}
	err := c.ClassCommandHandler.Delete(ctx, payload)
	if err == domain.ErrNotFound {
		return resp, status.Error(codes.NotFound, "`class_id` not found")
	} else if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}
	if err := c.PublishClassEvent(ctx, &mpb.EvtClass{
		Message: &mpb.EvtClass_DeleteClass_{
			DeleteClass: &mpb.EvtClass_DeleteClass{
				ClassId: classID,
			},
		},
	}); err != nil {
		ctxzap.Extract(ctx).Warn("ClassService.PublishClassEvent", zap.Error(err))
	}
	return resp, nil
}

func (c *ClassService) PublishClassEvent(ctx context.Context, msg *mpb.EvtClass) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	msgID, err := c.JSM.PublishAsyncContext(ctx, constants.SubjectMasterMgmtClassUpserted, data)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishClassEvent JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
	}
	return nil
}

func (c *ClassService) RetrieveClassesByIDs(ctx context.Context, req *mpb.RetrieveClassByIDsRequest) (*mpb.RetrieveClassByIDsResponse, error) {
	classIds := req.GetClassIds()
	if len(classIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "RetrieveClassByIDsRequest: classIds is empty")
	}

	payload := queries.GetByIds{IDs: classIds}
	classes, err := c.ClassQueryHandler.GetByIDs(ctx, payload)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf(`ClassCommand.RetrieveByIDs: %v`, err))
	}

	classResult := make([]*mpb.RetrieveClassByIDsResponse_Class, 0, len(classes))
	for _, c := range classes {
		class := &mpb.RetrieveClassByIDsResponse_Class{
			ClassId:    c.ClassID,
			Name:       c.Name,
			LocationId: c.LocationID,
		}
		classResult = append(classResult, class)
	}

	return &mpb.RetrieveClassByIDsResponse{
		Classes: classResult,
	}, nil
}
