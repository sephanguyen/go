package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SubscriptionModifyService struct {
	DB  database.Ext
	JSM nats.JetStreamManagement

	PackageRepo interface {
		Get(ctx context.Context, db database.QueryExecer, ID pgtype.Text) (*entities.Package, error)
		Upsert(ctx context.Context, db database.QueryExecer, e *entities.Package) error
	}

	StudentPackageRepo interface {
		Get(ctx context.Context, db database.QueryExecer, ID pgtype.Text) (*entities.StudentPackage, error)
		Insert(ctx context.Context, db database.QueryExecer, e *entities.StudentPackage) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.StudentPackage) error
		GetByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]*entities.StudentPackage, error)
		GetByStudentPackageIDAndStudentIDAndCourseID(ctx context.Context, db database.QueryExecer, studentPackageID pgtype.Text, studentID pgtype.Text, courseID pgtype.Text) (*entities.StudentPackage, error)
	}

	StudentPackageAccessPathRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, ents []*entities.StudentPackageAccessPath) error
		DeleteByStudentPackageIDs(ctx context.Context, db database.QueryExecer, spIDs pgtype.TextArray) error
	}

	StudentPackageClassRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.StudentPackageClass) error
		DeleteByStudentPackageIDs(ctx context.Context, db database.QueryExecer, spIDs pgtype.TextArray) error
		DeleteByStudentPackageIDAndCourseID(ctx context.Context, db database.QueryExecer, studentPackageID string, courseID string) error
	}
}

func (s *SubscriptionModifyService) CreatePackage(ctx context.Context, req *pb.CreatePackageRequest) (*pb.CreatePackageResponse, error) {
	startAt := req.StartAt.AsTime()
	endAt := req.EndAt.AsTime()

	p := &entities.Package{}
	database.AllNullEntity(p)
	err := multierr.Combine(
		p.ID.Set(idutil.ULIDNow()),
		p.Country.Set(req.Country.String()),
		p.Name.Set(req.Name),
		p.Descriptions.Set(req.Descriptions),
		p.Price.Set(req.Price),
		p.DiscountedPrice.Set(req.DiscountedPrice),
		p.PrioritizeLevel.Set(req.PrioritizeLevel),
		p.StartAt.Set(startAt),
		p.EndAt.Set(endAt),
		p.Duration.Set(req.Duration),
		p.Properties.Set(&entities.PackageProperties{
			CanWatchVideo:     req.Properties.CanWatchVideo,
			CanViewStudyGuide: req.Properties.CanViewStudyGuide,
			CanDoQuiz:         req.Properties.CanDoQuiz,
			LimitOnlineLesson: int(req.Properties.LimitOnlineLession),
			AskTutor: &entities.AskTutorCfg{
				TotalQuestionLimit: int(req.Properties.AskTutor.TotalQuestionLimit),
				LimitDuration:      req.Properties.AskTutor.LimitDuration.String(),
			},
		}),
		p.IsRecommended.Set(req.IsRecommended),
		p.IsActive.Set(true),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.PackageRepo.Upsert(ctx, s.DB, p); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreatePackageResponse{
		PackageId: p.ID.String,
	}, nil
}

func (s *SubscriptionModifyService) ToggleActivePackage(ctx context.Context, req *pb.ToggleActivePackageRequest) (*pb.ToggleActivePackageResponse, error) {
	p, err := s.PackageRepo.Get(ctx, s.DB, database.Text(req.PackageId))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	p.IsActive.Bool = !p.IsActive.Bool
	if err := s.PackageRepo.Upsert(ctx, s.DB, p); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ToggleActivePackageResponse{
		IsActive: p.IsActive.Bool,
	}, nil
}

func (s *SubscriptionModifyService) AddStudentPackage(ctx context.Context, req *pb.AddStudentPackageRequest) (*pb.AddStudentPackageResponse, error) {
	p, err := s.PackageRepo.Get(ctx, s.DB, database.Text(req.PackageId))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if !p.IsActive.Bool {
		return nil, status.Error(codes.InvalidArgument, "package is disabled")
	}

	props, err := p.GetProperties()
	if err != nil {
		return nil, status.Error(codes.Internal, "err parse properties")
	}

	now := time.Now()
	startAt := p.StartAt.Time
	endAt := p.EndAt.Time
	if p.Duration.Int > 0 {
		startAt = now
		endAt = now.Add(time.Duration(p.Duration.Int) * time.Hour * 24)
	}
	studentPackageID := idutil.ULIDNow()
	sp := &entities.StudentPackage{}
	database.AllNullEntity(sp)
	err = multierr.Combine(
		sp.ID.Set(studentPackageID),
		sp.StudentID.Set(req.StudentId),
		sp.PackageID.Set(p.ID.String),
		sp.StartAt.Set(startAt),
		sp.EndAt.Set(endAt),
		sp.Properties.Set(&entities.StudentPackageProps{
			CanWatchVideo:     props.CanWatchVideo,
			CanViewStudyGuide: props.CanViewStudyGuide,
			CanDoQuiz:         props.CanDoQuiz,
			LimitOnlineLesson: props.LimitOnlineLesson,
			AskTutor:          props.AskTutor,
		}),
		sp.IsActive.Set(true),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.StudentPackageRepo.Insert(ctx, s.DB, sp)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	courseIds := append(props.CanDoQuiz, props.CanViewStudyGuide...)
	courseIds = append(courseIds, props.CanWatchVideo...)

	event := &npb.EventStudentPackage{
		StudentPackage: &npb.EventStudentPackage_StudentPackage{
			StudentId: req.StudentId,
			Package: &npb.EventStudentPackage_Package{
				CourseIds:        golibs.Uniq(courseIds),
				StartDate:        timestamppb.New(startAt),
				EndDate:          timestamppb.New(endAt),
				StudentPackageId: studentPackageID,
			},
			IsActive: sp.IsActive.Bool,
		},
	}
	data, err := proto.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("unable marshal: %w", err)
	}

	_, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectStudentPackageEventNats, data)
	if err != nil {
		return nil, fmt.Errorf("err PublishAsync: %w", err)
	}

	return &pb.AddStudentPackageResponse{
		StudentPackageId: sp.ID.String,
	}, nil
}

func (s *SubscriptionModifyService) ToggleActiveStudentPackage(ctx context.Context, req *pb.ToggleActiveStudentPackageRequest) (*pb.ToggleActiveStudentPackageResponse, error) {
	p, err := s.StudentPackageRepo.Get(ctx, s.DB, database.Text(req.StudentPackageId))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	p.IsActive.Bool = !p.IsActive.Bool
	if err := s.StudentPackageRepo.Update(ctx, s.DB, p); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	props, err := p.GetProperties()

	if err != nil {
		return nil, status.Error(codes.Internal, "err parse properties")
	}

	courseIds := append(props.CanDoQuiz, props.CanViewStudyGuide...)
	courseIds = append(courseIds, props.CanWatchVideo...)

	event := &npb.EventStudentPackage{
		StudentPackage: &npb.EventStudentPackage_StudentPackage{
			StudentId: p.StudentID.String,
			Package: &npb.EventStudentPackage_Package{
				CourseIds:        golibs.Uniq(courseIds),
				StartDate:        timestamppb.New(p.StartAt.Time),
				EndDate:          timestamppb.New(p.EndAt.Time),
				StudentPackageId: req.StudentPackageId,
			},
			IsActive: p.IsActive.Bool,
		},
	}

	data, err := proto.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("unable marshal: %w", err)
	}

	_, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectStudentPackageEventNats, data)
	if err != nil {
		return nil, fmt.Errorf("err PublishAsync: %w", err)
	}

	return &pb.ToggleActiveStudentPackageResponse{
		IsActive: p.IsActive.Bool,
	}, nil
}

func (s *SubscriptionModifyService) AddStudentPackageCourse(ctx context.Context, req *pb.AddStudentPackageCourseRequest) (*pb.AddStudentPackageCourseResponse, error) {
	startAt := req.StartAt
	endAt := req.EndAt
	studentPackageID := idutil.ULIDNow()
	sp := &entities.StudentPackage{}
	database.AllNullEntity(sp)
	if req.StudentPackageExtra != nil {
		courseIds := make([]string, 0)
		locationIds := make([]string, 0)
		for _, value := range req.StudentPackageExtra {
			courseIds = append(courseIds, value.CourseId)
			locationIds = append(locationIds, value.LocationId)
		}
		err := multierr.Combine(
			sp.ID.Set(studentPackageID),
			sp.StudentID.Set(req.StudentId),
			sp.PackageID.Set(nil),
			sp.StartAt.Set(startAt.AsTime()),
			sp.EndAt.Set(endAt.AsTime()),
			sp.Properties.Set(&entities.StudentPackageProps{
				CanWatchVideo:     courseIds,
				CanViewStudyGuide: courseIds,
				CanDoQuiz:         courseIds,
			}),
			sp.IsActive.Set(true),
			sp.LocationIDs.Set(locationIds),
		)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	} else {
		err := multierr.Combine(
			sp.ID.Set(studentPackageID),
			sp.StudentID.Set(req.StudentId),
			sp.PackageID.Set(nil),
			sp.StartAt.Set(startAt.AsTime()),
			sp.EndAt.Set(endAt.AsTime()),
			sp.Properties.Set(&entities.StudentPackageProps{
				CanWatchVideo:     req.CourseIds,
				CanViewStudyGuide: req.CourseIds,
				CanDoQuiz:         req.CourseIds,
			}),
			sp.IsActive.Set(true),
			sp.LocationIDs.Set(req.LocationIds),
		)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	if err := database.ExecInTxWithRetry(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := s.StudentPackageRepo.Insert(ctx, tx, sp)
		if err != nil {
			return err
		}
		studentPackageAccessPaths, err := generateListStudentPackageAccessPathsFromStudentPackage(sp)
		if err != nil {
			return err
		}
		if err = s.StudentPackageAccessPathRepo.BulkUpsert(ctx, tx, studentPackageAccessPaths); err != nil {
			return err
		}
		if req.StudentPackageExtra != nil {
			studentPackageClasses, err := generateListStudentPackageClassesFromStudentPackageAndStudentPackagesExtras(sp, req.GetStudentPackageExtra())
			if err != nil {
				return err
			}
			if err = s.StudentPackageClassRepo.BulkUpsert(ctx, tx, studentPackageClasses); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if req.StudentPackageExtra != nil {
		courseIDs := make([]string, 0)
		locationIDs := make([]string, 0)
		for _, packageExtra := range req.StudentPackageExtra {
			courseIDs = append(courseIDs, packageExtra.CourseId)
			locationIDs = append(locationIDs, packageExtra.LocationId)
			event := &npb.EventStudentPackageV2{
				StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
					StudentId: req.StudentId,
					IsActive:  sp.IsActive.Bool,
					Package: &npb.EventStudentPackageV2_PackageV2{
						CourseId:   packageExtra.CourseId,
						LocationId: packageExtra.LocationId,
						ClassId:    packageExtra.ClassId,
						StartDate:  startAt,
						EndDate:    endAt,
					},
				},
			}
			err := s.publishEventWithSubjectToNats(ctx, event, constants.SubjectStudentPackageV2EventNats)
			if err != nil {
				return nil, err
			}
		}

		event := &npb.EventStudentPackage{
			StudentPackage: &npb.EventStudentPackage_StudentPackage{
				StudentId: req.StudentId,
				Package: &npb.EventStudentPackage_Package{
					CourseIds:        golibs.Uniq(courseIDs),
					StartDate:        startAt,
					EndDate:          endAt,
					LocationIds:      locationIDs,
					StudentPackageId: studentPackageID,
				},
				IsActive: sp.IsActive.Bool,
			},
		}
		err := s.publishEventWithSubjectToNats(ctx, event, constants.SubjectStudentPackageEventNats)
		if err != nil {
			return nil, err
		}
	} else {
		event := &npb.EventStudentPackage{
			StudentPackage: &npb.EventStudentPackage_StudentPackage{
				StudentId: req.StudentId,
				Package: &npb.EventStudentPackage_Package{
					CourseIds:        golibs.Uniq(req.CourseIds),
					StartDate:        startAt,
					EndDate:          endAt,
					LocationIds:      req.LocationIds,
					StudentPackageId: studentPackageID,
				},
				IsActive: sp.IsActive.Bool,
			},
		}
		err := s.publishEventWithSubjectToNats(ctx, event, constants.SubjectStudentPackageEventNats)
		if err != nil {
			return nil, err
		}
	}

	return &pb.AddStudentPackageCourseResponse{
		StudentPackageId: sp.ID.String,
	}, nil
}

func toStudentPackageAccessPath(sp *entities.StudentPackage, courseID, locationID string) (*entities.StudentPackageAccessPath, error) {
	spap := &entities.StudentPackageAccessPath{}
	database.AllNullEntity(spap)

	if err := multierr.Combine(
		spap.StudentPackageID.Set(sp.ID),
		spap.CourseID.Set(courseID),
		spap.StudentID.Set(sp.StudentID),
		spap.LocationID.Set(locationID),
	); err != nil {
		return nil, err
	}

	return spap, nil
}

func toStudentPackageClass(sp *entities.StudentPackage, courseID, locationID string, classID string) (*entities.StudentPackageClass, error) {
	spc := &entities.StudentPackageClass{}
	database.AllNullEntity(spc)
	if err := multierr.Combine(
		spc.StudentPackageID.Set(sp.ID),
		spc.CourseID.Set(courseID),
		spc.StudentID.Set(sp.StudentID),
		spc.LocationID.Set(locationID),
		spc.ClassID.Set(classID),
	); err != nil {
		return nil, err
	}
	return spc, nil

}

func generateListStudentPackageAccessPathsFromStudentPackage(sp *entities.StudentPackage) ([]*entities.StudentPackageAccessPath, error) {
	studentPackageAccessPaths := make([]*entities.StudentPackageAccessPath, 0)

	var locationIDs []string
	if err := sp.LocationIDs.AssignTo(&locationIDs); err != nil {
		return nil, err
	}

	courseIDs, err := sp.GetCourseIDs()
	if err != nil {
		return nil, err
	}

	for _, courseID := range courseIDs {
		if len(locationIDs) > 0 {
			for _, locationID := range locationIDs {
				spap, err := toStudentPackageAccessPath(sp, courseID, locationID)
				if err != nil {
					return nil, err
				}
				studentPackageAccessPaths = append(studentPackageAccessPaths, spap)
			}
		} else {
			spap, err := toStudentPackageAccessPath(sp, courseID, "")
			if err != nil {
				return nil, err
			}
			studentPackageAccessPaths = append(studentPackageAccessPaths, spap)
		}
	}

	return studentPackageAccessPaths, nil
}

func generateListStudentPackageClassesFromStudentPackageAndEditTimeStudentPackageExtra(
	sp *entities.StudentPackage,
	packageExtras []*pb.EditTimeStudentPackageRequest_EditTimeStudentPackageExtra,
) ([]*entities.StudentPackageClass, error) {
	studentPackageClasses := make([]*entities.StudentPackageClass, 0)
	for _, packageClass := range packageExtras {
		if packageClass.ClassId == "" {
			continue
		}
		courseIds, err := sp.GetCourseIDs()
		if err != nil {
			return nil, err
		}
		spc, err := toStudentPackageClass(sp, courseIds[0], packageClass.LocationId, packageClass.ClassId)
		if err != nil {
			return nil, err
		}
		studentPackageClasses = append(studentPackageClasses, spc)
	}
	return studentPackageClasses, nil
}

func generateListStudentPackageClassesFromStudentPackageAndStudentPackagesExtras(
	sp *entities.StudentPackage,
	packageExtras []*pb.AddStudentPackageCourseRequest_AddStudentPackageExtra,
) ([]*entities.StudentPackageClass, error) {
	studentPackageClasses := make([]*entities.StudentPackageClass, 0)
	for _, packageClass := range packageExtras {
		if packageClass.ClassId == "" {
			continue
		}
		spc, err := toStudentPackageClass(sp, packageClass.CourseId, packageClass.LocationId, packageClass.ClassId)
		if err != nil {
			return nil, err
		}
		studentPackageClasses = append(studentPackageClasses, spc)
	}
	return studentPackageClasses, nil
}

func (s *SubscriptionModifyService) EditTimeStudentPackage(ctx context.Context, req *pb.EditTimeStudentPackageRequest) (*pb.EditTimeStudentPackageResponse, error) {
	p, err := s.StudentPackageRepo.Get(ctx, s.DB, database.Text(req.StudentPackageId))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	startAt := req.StartAt
	endAt := req.EndAt
	err = p.StartAt.Set(startAt.AsTime())
	if err != nil {
		return nil, fmt.Errorf("EditTimeStudentPackage set student startAt error: %w", err)
	}
	err = p.EndAt.Set(endAt.AsTime())
	if err != nil {
		return nil, fmt.Errorf("EditTimeStudentPackage set student endAt error: %w", err)
	}
	if req.StudentPackageExtra != nil {
		locationIDs := make([]string, 0)
		for _, value := range req.StudentPackageExtra {
			locationIDs = append(locationIDs, value.LocationId)
		}
		err = p.LocationIDs.Set(golibs.Uniq(locationIDs))
		if err != nil {
			return nil, fmt.Errorf("EditTimeStudentPackage set student locationIDs error: %w", err)
		}
	} else {
		err = p.LocationIDs.Set(golibs.Uniq(req.LocationIds))
		if err != nil {
			return nil, fmt.Errorf("EditTimeStudentPackage set student locationIDs error: %w", err)
		}
	}
	if err := database.ExecInTxWithRetry(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.StudentPackageRepo.Update(ctx, tx, p); err != nil {
			return err
		}

		if err := s.StudentPackageAccessPathRepo.DeleteByStudentPackageIDs(ctx, tx, database.TextArray([]string{p.ID.String})); err != nil {
			return err
		}

		studentPackageAccessPaths, err := generateListStudentPackageAccessPathsFromStudentPackage(p)
		if err != nil {
			return err
		}
		if req.StudentPackageExtra != nil {
			if err := s.StudentPackageClassRepo.DeleteByStudentPackageIDs(ctx, tx, database.TextArray([]string{p.ID.String})); err != nil {
				return err
			}
			studentPackageClasses, err := generateListStudentPackageClassesFromStudentPackageAndEditTimeStudentPackageExtra(p, req.StudentPackageExtra)
			if err != nil {
				return err
			}
			if err := s.StudentPackageClassRepo.BulkUpsert(ctx, tx, studentPackageClasses); err != nil {
				return err
			}
		}

		return s.StudentPackageAccessPathRepo.BulkUpsert(ctx, tx, studentPackageAccessPaths)
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	courseIDs, err := p.GetCourseIDs()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if req.StudentPackageExtra != nil {
		courseIds, err := p.GetCourseIDs()
		if err != nil {
			return nil, err
		}
		locationIDs := make([]string, 0)
		for _, packageExtra := range req.StudentPackageExtra {
			locationIDs = append(locationIDs, packageExtra.LocationId)
			event := &npb.EventStudentPackageV2{
				StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
					StudentId: p.StudentID.String,
					Package: &npb.EventStudentPackageV2_PackageV2{
						CourseId:   courseIds[0],
						LocationId: packageExtra.LocationId,
						ClassId:    packageExtra.ClassId,
						StartDate:  startAt,
						EndDate:    endAt,
					},
					IsActive: p.IsActive.Bool,
				},
			}
			err := s.publishEventWithSubjectToNats(ctx, event, constants.SubjectStudentPackageV2EventNats)
			if err != nil {
				return nil, err
			}
		}

		event := &npb.EventStudentPackage{
			StudentPackage: &npb.EventStudentPackage_StudentPackage{
				StudentId: p.StudentID.String,
				Package: &npb.EventStudentPackage_Package{
					CourseIds:        courseIDs,
					StartDate:        startAt,
					EndDate:          endAt,
					LocationIds:      locationIDs,
					StudentPackageId: req.StudentPackageId,
				},
				IsActive: p.IsActive.Bool,
			},
		}

		err = s.publishEventWithSubjectToNats(ctx, event, constants.SubjectStudentPackageEventNats)
		if err != nil {
			return nil, err
		}
	} else {
		event := &npb.EventStudentPackage{
			StudentPackage: &npb.EventStudentPackage_StudentPackage{
				StudentId: p.StudentID.String,
				Package: &npb.EventStudentPackage_Package{
					CourseIds:        courseIDs,
					StartDate:        startAt,
					EndDate:          endAt,
					LocationIds:      req.LocationIds,
					StudentPackageId: req.StudentPackageId,
				},
				IsActive: p.IsActive.Bool,
			},
		}

		err = s.publishEventWithSubjectToNats(ctx, event, constants.SubjectStudentPackageEventNats)
		if err != nil {
			return nil, err
		}

	}

	return &pb.EditTimeStudentPackageResponse{
		StudentPackageId: p.ID.String,
	}, nil
}

func (s *SubscriptionModifyService) ListStudentPackage(ctx context.Context, req *pb.ListStudentPackageRequest) (*pb.ListStudentPackageResponse, error) {
	studentPackages, err := s.StudentPackageRepo.GetByStudentIDs(ctx, s.DB, database.TextArray(req.StudentIds))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("ListStudentPackage.GetByStudentIDs %v", err))
	}

	studentPackagesPb := make([]*pb.StudentPackage, 0)
	for _, p := range studentPackages {
		prop, err := p.GetProperties()
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("ListStudentPackage.GetProperties %v", err))
		}
		askTutor := &pb.PackageProperties_AskTutorCfg{}
		if prop.AskTutor != nil {
			askTutor.TotalQuestionLimit = int32(prop.AskTutor.TotalQuestionLimit)
			askTutor.LimitDuration = bpb.AskDuration(bpb.AskDuration_value[prop.AskTutor.LimitDuration])
		}
		studentPackagesPb = append(studentPackagesPb, &pb.StudentPackage{
			Id:        p.ID.String,
			StudentId: p.StudentID.String,
			PackageId: p.PackageID.String,
			StartAt:   timestamppb.New(p.StartAt.Time),
			EndAt:     timestamppb.New(p.EndAt.Time),
			// properties
			Properties: &pb.PackageProperties{
				CanWatchVideo:      prop.CanWatchVideo,
				CanViewStudyGuide:  prop.CanViewStudyGuide,
				CanDoQuiz:          prop.CanDoQuiz,
				LimitOnlineLession: int32(prop.LimitOnlineLesson),
				AskTutor:           askTutor,
			},
			IsActive:    p.IsActive.Bool,
			CreatedAt:   timestamppb.New(p.CreatedAt.Time),
			UpdatedAt:   timestamppb.New(p.UpdatedAt.Time),
			LocationIds: database.FromTextArray(p.LocationIDs),
		})
	}

	return &pb.ListStudentPackageResponse{StudentPackages: studentPackagesPb}, nil
}

func (s *SubscriptionModifyService) ListStudentPackageV2(req *pb.ListStudentPackageV2Request, stream pb.SubscriptionModifierService_ListStudentPackageV2Server) error {
	studentPackages, err := s.StudentPackageRepo.GetByStudentIDs(context.Background(), s.DB, database.TextArray(req.StudentIds))
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("ListStudentPackageV2.GetByStudentIDs %v", err))
	}

	for _, p := range studentPackages {
		prop, err := p.GetProperties()
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("ListStudentPackageV2.GetProperties %v", err))
		}
		askTutor := &pb.PackageProperties_AskTutorCfg{}
		if prop.AskTutor != nil {
			askTutor.TotalQuestionLimit = int32(prop.AskTutor.TotalQuestionLimit)
			askTutor.LimitDuration = bpb.AskDuration(bpb.AskDuration_value[prop.AskTutor.LimitDuration])
		}
		studentPackagePb := &pb.StudentPackage{
			Id:        p.ID.String,
			StudentId: p.StudentID.String,
			PackageId: p.PackageID.String,
			StartAt:   timestamppb.New(p.StartAt.Time),
			EndAt:     timestamppb.New(p.EndAt.Time),
			// properties
			Properties: &pb.PackageProperties{
				CanWatchVideo:      prop.CanWatchVideo,
				CanViewStudyGuide:  prop.CanViewStudyGuide,
				CanDoQuiz:          prop.CanDoQuiz,
				LimitOnlineLession: int32(prop.LimitOnlineLesson),
				AskTutor:           askTutor,
			},
			IsActive:  p.IsActive.Bool,
			CreatedAt: timestamppb.New(p.CreatedAt.Time),
			UpdatedAt: timestamppb.New(p.UpdatedAt.Time),
		}
		if err := stream.Send(&pb.ListStudentPackageV2Response{
			StudentPackage: studentPackagePb,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *SubscriptionModifyService) RegisterStudentClass(ctx context.Context, req *pb.RegisterStudentClassRequest) error {
	zapLogger := ctxzap.Extract(ctx).Sugar()
	var events []*npb.EventStudentPackageV2
	for _, classInformation := range req.ClassesInformation {
		studentPackage, err := s.StudentPackageRepo.Get(ctx, s.DB, database.Text(classInformation.GetStudentPackageId()))
		if err != nil {
			return status.Error(codes.NotFound, err.Error())
		}
		if err := database.ExecInTxWithRetry(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			if err := s.StudentPackageClassRepo.DeleteByStudentPackageIDAndCourseID(ctx, tx, studentPackage.ID.String, classInformation.CourseId); err != nil {
				return err
			}
			studentPackageClasses, err := generateListStudentPackageClassesFromStudentPackageAndClassInformation(studentPackage, req.ClassesInformation)
			if err != nil {
				return err
			}
			if err := s.StudentPackageClassRepo.BulkUpsert(ctx, tx, studentPackageClasses); err != nil {
				return err
			}
			for _, studentPackageClass := range studentPackageClasses {
				event := &npb.EventStudentPackageV2{
					StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
						StudentId: classInformation.StudentId,
						Package: &npb.EventStudentPackageV2_PackageV2{
							CourseId:   studentPackageClass.CourseID.String,
							LocationId: studentPackageClass.LocationID.String,
							ClassId:    classInformation.ClassId,
							StartDate:  classInformation.StartTime,
							EndDate:    classInformation.EndTime,
						},
						IsActive: studentPackage.IsActive.Bool,
					},
				}
				events = append(events, event)
			}

			return nil
		}); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	for _, event := range events {
		if err := s.publishEventWithSubjectToNats(ctx, event, constants.SubjectStudentPackageV2EventNats); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		zapLogger.Infof(
			"Publish event subject StudentPackageV2.Upserted with student_id: %s, class_id: %s, location_id: %s, course_id: %s",
			event.StudentPackage.StudentId, event.StudentPackage.Package.ClassId, event.StudentPackage.Package.LocationId, event.StudentPackage.Package.CourseId,
		)
	}

	return nil
}

func generateListStudentPackageClassesFromStudentPackageAndClassInformation(
	studentPackage *entities.StudentPackage,
	classesInformation []*pb.RegisterStudentClassRequest_ClassInformation,
) ([]*entities.StudentPackageClass, error) {
	studentPackageClasses := make([]*entities.StudentPackageClass, 0, len(classesInformation))
	for _, packageClass := range classesInformation {
		studentPackageClass, err := toStudentPackageClass(studentPackage, packageClass.CourseId, studentPackage.LocationIDs.Elements[0].String, packageClass.ClassId)
		if err != nil {
			return nil, err
		}
		studentPackageClasses = append(studentPackageClasses, studentPackageClass)
	}
	return studentPackageClasses, nil
}

func (s *SubscriptionModifyService) publishEventWithSubjectToNats(ctx context.Context, event interface{}, subject string) error {
	var data []byte
	var err error
	switch eventType := event.(type) {
	case *npb.EventStudentPackage:
		data, err = proto.Marshal(eventType)
		if err != nil {
			return fmt.Errorf("unable marshal: %w", err)
		}
	case *npb.EventStudentPackageV2:
		data, err = proto.Marshal(eventType)
		if err != nil {
			return fmt.Errorf("unable marshal: %w", err)
		}
	default:
		return fmt.Errorf("failed marshal, unknown type")
	}
	_, err = s.JSM.PublishAsyncContext(ctx, subject, data)
	if err != nil {
		return fmt.Errorf("err PublishAsync: %w", err)
	}
	return nil
}

func (s *SubscriptionModifyService) WrapperRegisterStudentClass(ctx context.Context, req *pb.WrapperRegisterStudentClassRequest) error {
	classesInfoReq := make([]*pb.RegisterStudentClassRequest_ClassInformation, 0, len(req.ReserveClassesInformation))
	for _, reserveClassInfo := range req.ReserveClassesInformation {
		studentPackageID := reserveClassInfo.StudentPackageId
		studentID := reserveClassInfo.StudentId
		courseID := reserveClassInfo.CourseId
		classID := reserveClassInfo.ClassId

		sp, err := s.StudentPackageRepo.GetByStudentPackageIDAndStudentIDAndCourseID(ctx, s.DB, database.Text(studentPackageID), database.Text(studentID), database.Text(courseID))
		if err != nil {
			return status.Error(codes.NotFound, fmt.Errorf("GetByStudentPackageIDAndStudentIDAndCourseID not found: %w", err).Error())
		}

		classStartTime := timestamppb.Now()
		classEndTime := timestamppb.New(sp.EndAt.Time)

		classInfoReq := &pb.RegisterStudentClassRequest_ClassInformation{
			StudentPackageId: studentPackageID,
			StudentId:        studentID,
			CourseId:         courseID,
			ClassId:          classID,
			StartTime:        classStartTime,
			EndTime:          classEndTime,
		}

		classesInfoReq = append(classesInfoReq, classInfoReq)
	}

	registerClassReq := &pb.RegisterStudentClassRequest{
		ClassesInformation: classesInfoReq,
	}

	return s.RegisterStudentClass(ctx, registerClassReq)
}

func (s *SubscriptionModifyService) RetrieveStudentPackagesUnderCourse(_ context.Context, _ *pb.RetrieveStudentPackagesUnderCourseRequest) (*pb.RetrieveStudentPackagesUnderCourseResponse, error) {
	fakeStartTime := timestamppb.New(time.Now().Add(-30 * 24 * time.Hour))
	fakeEndTime := timestamppb.New(time.Now().Add(30 * 24 * time.Hour))
	mockResult := &pb.RetrieveStudentPackagesUnderCourseResponse{
		Items: []*pb.RetrieveStudentPackagesUnderCourseResponse_StudentPackageUnderCourse{
			{
				StudentPackageId: "01GG9CA00V16Q3GZY8X6XYGD8F",
				StudentId:        "01GG9C95FJ7HGZN3JC2KTCJSC0",
				LocationId:       "01G802ZD2PGM9FBH62W520B21E",
				CourseId:         "01GG7TZETTSV8WP525F361SF6X",
				ClassId:          "01GG7V26G1RDHSFN51YVX85ATB",
				StartAt:          fakeStartTime,
				EndAt:            fakeEndTime,
				CreatedAt:        timestamppb.Now(),
				UpdatedAt:        timestamppb.Now(),
			},
			{
				StudentPackageId: "01GGS22E44Y39BPM1TXFVNAPQA",
				StudentId:        "01GGF101WJQD9R5XEH0J6RZZTC",
				LocationId:       "01G802ZD2PGM9FBH62W520B21E",
				CourseId:         "01GG7TZETTSV8WP525F361SF6X",
				ClassId:          "01GG7V26G1RDHSFN51YVX85ATB",
				StartAt:          fakeStartTime,
				EndAt:            fakeEndTime,
				CreatedAt:        timestamppb.Now(),
				UpdatedAt:        timestamppb.Now(),
			},
			{
				StudentPackageId: "01GH85TKV1AWMK0JBRCRKYSBV1",
				StudentId:        "01GH833Z40GNMDW5ZAQWANHMM4",
				LocationId:       "01G802ZD2PGM9FBH62W520B21E",
				CourseId:         "01GG7TZETTSV8WP525F361SF6X",
				StartAt:          fakeStartTime,
				EndAt:            fakeEndTime,
				CreatedAt:        timestamppb.Now(),
				UpdatedAt:        timestamppb.Now(),
			},
			{
				StudentPackageId: "01H3XBY75QZ9CAK5G8YTXQED03",
				StudentId:        "01GSSJJE7WNQZ3N160ABVCXGS7",
				LocationId:       "01G802ZD2PGM9FBH62W520B21E",
				CourseId:         "01GG7TZETTSV8WP525F361SF6X",
				StartAt:          fakeStartTime,
				EndAt:            fakeEndTime,
				CreatedAt:        timestamppb.Now(),
				UpdatedAt:        timestamppb.Now(),
			},
		},
		TotalItems: 4,
	}

	return mockResult, nil
}
