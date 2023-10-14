package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type TaskAssignmentService struct {
	sspb.UnimplementedTaskAssignmentServer
	DB database.Ext

	BookRepo                    IBookRepository
	ChapterRepo                 IChapterRepository
	BookChapterRepo             IBookChapterRepository
	TopicRepo                   ITopicRepository
	CourseBookRepo              ICourseBookRepository
	StudyPlanRepo               IStudyPlanRepository
	StudentStudyPlanRepo        IStudentStudyPlanRepository
	AssignmentRepo              IAssignmentRepository
	StudyPlanItemRepo           IStudyPlanItemRepository
	AssignmentStudyPlanItemRepo IAssignmentStudyPlanItemRepository
	LoStudyPlanItemRepo         ILoStudyPlanItemRepository
	LearningObjectiveRepo       ILearningObjectiveRepository
	MasterStudyPlanRepo         IMasterStudyPlanRepository
	IndividualStudyPlanRepo     IIndividualStudyPlanRepository

	TaskAssignmentRepo interface {
		Insert(ctx context.Context, db database.QueryExecer, e *entities.TaskAssignment) error
		List(ctx context.Context, db database.QueryExecer, learningMaterialIds pgtype.TextArray) ([]*entities.TaskAssignment, error)
		Update(ctx context.Context, db database.QueryExecer, e *entities.TaskAssignment) error
		Upsert(ctx context.Context, db database.QueryExecer, e *entities.TaskAssignment) error
	}

	BobStudentReaderService interface {
		RetrieveStudentProfile(ctx context.Context, in *bpb.RetrieveStudentProfileRequest, opts ...grpc.CallOption) (*bpb.RetrieveStudentProfileResponse, error)
	}
}

func NewTaskAssignmentService(
	db database.Ext,
	bobStudentReaderService bpb.StudentReaderServiceClient,
) sspb.TaskAssignmentServer {
	return &TaskAssignmentService{
		DB:                          db,
		BookRepo:                    new(repositories.BookRepo),
		TopicRepo:                   new(repositories.TopicRepo),
		TaskAssignmentRepo:          new(repositories.TaskAssignmentRepo),
		ChapterRepo:                 new(repositories.ChapterRepo),
		BookChapterRepo:             new(repositories.BookChapterRepo),
		CourseBookRepo:              new(repositories.CourseBookRepo),
		StudyPlanRepo:               new(repositories.StudyPlanRepo),
		StudentStudyPlanRepo:        new(repositories.StudentStudyPlanRepo),
		AssignmentRepo:              new(repositories.AssignmentRepo),
		StudyPlanItemRepo:           new(repositories.StudyPlanItemRepo),
		AssignmentStudyPlanItemRepo: new(repositories.AssignmentStudyPlanItemRepo),
		LoStudyPlanItemRepo:         new(repositories.LoStudyPlanItemRepo),
		LearningObjectiveRepo:       new(repositories.LearningObjectiveRepo),
		BobStudentReaderService:     bobStudentReaderService,
		MasterStudyPlanRepo:         new(repositories.MasterStudyPlanRepo),
		IndividualStudyPlanRepo:     new(repositories.IndividualStudyPlan),
	}
}

func toTaskAssignmentEnt(req *sspb.TaskAssignmentBase) (*entities.TaskAssignment, error) {
	e := &entities.TaskAssignment{}
	database.AllNullEntity(e)
	id := req.Base.LearningMaterialId
	if id == "" {
		id = idutil.ULIDNow()
	}
	now := time.Now()
	err := multierr.Combine(
		e.ID.Set(id),
		e.TopicID.Set(req.Base.TopicId),
		e.Name.Set(req.Base.Name),
		e.Type.Set(sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT.String()),
		e.IsPublished.Set(false),
		e.SetDefaultVendorType(),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),

		e.Attachments.Set(req.Attachments),
		e.Instruction.Set(req.Instruction),
		e.RequireDuration.Set(req.RequireDuration),
		e.RequireCompleteDate.Set(req.RequireCompleteDate),
		e.RequireUnderstandingLevel.Set(req.RequireUnderstandingLevel),
		e.RequireCorrectness.Set(req.RequireCorrectness),
		e.RequireAttachment.Set(req.RequireAttachment),
		e.RequireAssignmentNote.Set(req.RequireAssignmentNote),
	)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func validateInsertTaskAssignmentReq(req *sspb.TaskAssignmentBase) error {
	if req.Base.LearningMaterialId != "" {
		return fmt.Errorf("LearningMaterialId must be empty")
	}
	if req.Base.TopicId == "" {
		return fmt.Errorf("TopicId must not be empty")
	}
	return nil
}

func (s *TaskAssignmentService) InsertTaskAssignment(ctx context.Context, req *sspb.InsertTaskAssignmentRequest) (*sspb.InsertTaskAssignmentResponse, error) {
	if err := validateInsertTaskAssignmentReq(req.TaskAssignment); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("validateTaskAssignmentReq: %w", err).Error())
	}
	fc, err := toTaskAssignmentEnt(req.TaskAssignment)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("toTaskAssignmentEnt: %w", err).Error())
	}
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		tp, err := s.TopicRepo.RetrieveByID(ctx, tx, fc.TopicID, repositories.WithUpdateLock())
		if err != nil {
			return fmt.Errorf("s.TopicRepo.RetrieveByID: %w", err)
		}
		if err := fc.DisplayOrder.Set(tp.LODisplayOrderCounter.Int + 1); err != nil {
			return fmt.Errorf("fc.DisplayOrder.Set: %w", err)
		}
		if err := s.TaskAssignmentRepo.Insert(ctx, tx, fc); err != nil {
			return fmt.Errorf("s.TaskAssignmentRepo.Insert: %w", err)
		}
		if err := s.TopicRepo.UpdateLODisplayOrderCounter(ctx, tx, tp.ID, database.Int4(1)); err != nil {
			return fmt.Errorf("s.TopicRepo.UpdateLODisplayOrderCounter: %w", err)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}
	return &sspb.InsertTaskAssignmentResponse{
		LearningMaterialId: fc.LearningMaterial.ID.String,
	}, nil
}

func validateUpdateTaskAssignmentReq(req *sspb.TaskAssignmentBase) error {
	if req.Base.LearningMaterialId == "" {
		return fmt.Errorf("empty learning_material_id")
	}
	return nil
}

func (s *TaskAssignmentService) UpdateTaskAssignment(ctx context.Context, req *sspb.UpdateTaskAssignmentRequest) (*sspb.UpdateTaskAssignmentResponse, error) {
	if err := validateUpdateTaskAssignmentReq(req.TaskAssignment); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("validateUpdateTaskAssignmentReq: %w", err).Error())
	}

	taskAssignment, err := toTaskAssignmentEnt(req.TaskAssignment)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("toTaskAssignmentEnt: %w", err).Error())
	}

	if err := s.TaskAssignmentRepo.Update(ctx, s.DB, taskAssignment); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("unable to update task assignment: %w", err).Error())
	}

	return &sspb.UpdateTaskAssignmentResponse{}, nil
}

func ToTaskAssignmentPb(src *entities.TaskAssignment) (*sspb.TaskAssignmentBase, error) {
	var attachments []string
	err := src.Attachments.AssignTo(&attachments)
	if err != nil {
		return nil, err
	}
	return &sspb.TaskAssignmentBase{
		Base: &sspb.LearningMaterialBase{
			LearningMaterialId: src.LearningMaterial.ID.String,
			TopicId:            src.LearningMaterial.TopicID.String,
			Name:               src.LearningMaterial.Name.String,
			Type:               src.LearningMaterial.Type.String,
			DisplayOrder:       wrapperspb.Int32(int32(src.DisplayOrder.Int)),
		},
		Attachments:               attachments,
		Instruction:               src.Instruction.String,
		RequireDuration:           src.RequireDuration.Bool,
		RequireCompleteDate:       src.RequireCompleteDate.Bool,
		RequireUnderstandingLevel: src.RequireUnderstandingLevel.Bool,
		RequireCorrectness:        src.RequireCorrectness.Bool,
		RequireAttachment:         src.RequireAttachment.Bool,
		RequireAssignmentNote:     src.RequireAssignmentNote.Bool,
	}, nil
}

func (s *TaskAssignmentService) ListTaskAssignment(ctx context.Context, req *sspb.ListTaskAssignmentRequest) (*sspb.ListTaskAssignmentResponse, error) {
	ids := req.LearningMaterialIds
	if len(ids) == 0 {
		return nil, status.Error(codes.InvalidArgument, "LearningMaterialIds must not be empty")
	}

	taskAssignments, err := s.TaskAssignmentRepo.List(ctx, s.DB, database.TextArray(req.LearningMaterialIds))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, fmt.Errorf("task assignment not found: %w", err).Error())
		}
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.TaskAssignmentRepo.List: %w", err).Error())
	}

	taskAssignmentsPb := make([]*sspb.TaskAssignmentBase, 0, len(taskAssignments))
	for _, taskAssignment := range taskAssignments {
		assignmentBase, err := ToTaskAssignmentPb(taskAssignment)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "error convert task assignment to pb: %v", err)
		}
		taskAssignmentsPb = append(taskAssignmentsPb, assignmentBase)
	}
	return &sspb.ListTaskAssignmentResponse{
		TaskAssignments: taskAssignmentsPb,
	}, nil
}

func (s *TaskAssignmentService) UpsertAdhocTaskAssignment(ctx context.Context, req *sspb.UpsertAdhocTaskAssignmentRequest) (*sspb.UpsertAdhocTaskAssignmentResponse, error) {
	if err := s.verifyUpsertAdHocTaskAssignmentRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("s.verifyUpsertAdHocTaskAssignmentRequest: %w", err).Error())
	}

	// retrieve student's profile form Bob
	studentProfile, err := s.RetrieveStudentProfile(ctx, req.StudentId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.RetrieveStudentProfile: %w", err).Error())
	}

	var LearningMaterialID string
	if err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		factory := NewStudentAdHocFactory(
			tx,
			s.BookRepo,
			s.ChapterRepo,
			s.BookChapterRepo,
			s.TopicRepo,
			s.CourseBookRepo,
			s.StudyPlanRepo,
			s.StudentStudyPlanRepo,
			s.AssignmentRepo,
			s.StudyPlanItemRepo,
			s.AssignmentStudyPlanItemRepo,
			s.LoStudyPlanItemRepo,
			s.LearningObjectiveRepo,
			s.MasterStudyPlanRepo,
			s.IndividualStudyPlanRepo,
		)
		// set and validate student and grade
		if err := factory.SetStudent(studentProfile); err != nil {
			return fmt.Errorf("factory.SetStudent: %w", err)
		}

		// get existed or create new adhoc book contents
		var book *entities.Book
		var topicID string

		// try retrieve existed adhoc book
		if book, err = s.BookRepo.RetrieveAdHocBookByCourseIDAndStudentID(
			ctx, tx, database.Text(req.CourseId), database.Text(req.StudentId),
		); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("s.BookRepo.RetrieveAdHocBookByCourseIDAndStudentID: %w", err)
		}
		if book != nil {
			// retrieve topic if adhoc book existed
			topics, err := s.TopicRepo.FindByBookIDs(
				ctx, tx,
				database.TextArrayVariadic(book.ID.String),
				pgtype.TextArray{Status: pgtype.Null},
				pgtype.Int4{Status: pgtype.Null},
				pgtype.Int4{Status: pgtype.Null},
			)
			if err != nil {
				return fmt.Errorf("s.TopicRepo.FindByBookIDs: %w", err)
			}
			if len(topics) == 0 {
				return fmt.Errorf("book (%s) must have a topic", book.ID.String)
			}

			topicID = topics[0].ID.String
		} else {
			// create new adhoc book
			bookContent, err := factory.CreateAdHocBookContent(ctx, CreateBookContentInput{
				BookName:    req.BookName,
				ChapterName: req.ChapterName,
				TopicName:   req.TopicName,
			})
			if err != nil {
				return fmt.Errorf("factory.CreateAdHocBookContent: %w", err)
			}
			book = bookContent.Book
			topicID = bookContent.Topic.ID.String
		}

		// assert task assignment data
		if req.TaskAssignment.Base == nil {
			req.TaskAssignment.Base = &sspb.LearningMaterialBase{}
		}
		req.TaskAssignment.Base.TopicId = topicID
		// upsert task assignment
		taskAssignmentEnt, err := toTaskAssignmentEnt(req.TaskAssignment)

		if err != nil {
			return fmt.Errorf("toTaskAssignmentEnt: %w", err)
		}
		// create  course_book and study plan
		if err := factory.CreateAdHocStudyPlan(ctx, CreateStudyPlanInput{
			StudyPlanName:      req.StudyPlanName,
			BookID:             book.ID.String,
			CourseID:           req.CourseId,
			LearningMaterialID: taskAssignmentEnt.ID.String,
			StarDate:           req.StartDate,
			EndDate:            req.EndDate,
		}); err != nil {
			return fmt.Errorf("factory.CreateAdHocStudyPlan: %w", err)
		}

		if err := s.TaskAssignmentRepo.Upsert(ctx, tx, taskAssignmentEnt); err != nil {
			return fmt.Errorf("s.TaskAssignmentRepo.Upsert: %w", err)
		}

		LearningMaterialID = taskAssignmentEnt.ID.String

		return nil
	}); err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}

	return &sspb.UpsertAdhocTaskAssignmentResponse{
		LearningMaterialId: LearningMaterialID,
	}, nil
}

func (s *TaskAssignmentService) verifyUpsertAdHocTaskAssignmentRequest(req *sspb.UpsertAdhocTaskAssignmentRequest) error {
	if req.CourseId == "" {
		return fmt.Errorf("CourseId must not be empty")
	}

	if req.StudentId == "" {
		return fmt.Errorf("StudentId must not be empty")
	}

	if req.StartDate == nil {
		return fmt.Errorf("StartDate must not be empty")
	}

	if req.TaskAssignment == nil {
		return fmt.Errorf("TaskAssignment must not be empty")
	}

	return nil
}

// Retrieve student from Bob and validate
func (s *TaskAssignmentService) RetrieveStudentProfile(ctx context.Context, studentID string) (*bpb.StudentProfile, error) {
	cctx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("interceptors.GetOutgoingContext: %w", err)
	}
	resp, err := s.BobStudentReaderService.RetrieveStudentProfile(cctx, &bpb.RetrieveStudentProfileRequest{
		StudentIds: []string{studentID},
	})
	if err != nil {
		return nil, fmt.Errorf("s.BobStudentReaderService.RetrieveStudentProfile: %w", err)
	}

	if len(resp.Items) < 1 ||
		resp.Items[0].GetProfile().GetSchool() == nil {
		return nil, fmt.Errorf("student must belongs to a school")
	}
	return resp.Items[0].Profile, nil
}
