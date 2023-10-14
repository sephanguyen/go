package studyplans

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/eureka/services"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	IStudyPlanRepository interface {
		RecursiveSoftDeleteStudyPlanByStudyPlanIDInCourse(ctx context.Context, db database.QueryExecer, courseStudyPlanID pgtype.Text) ([]string, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlan) error
		FindByID(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.Text) (*entities.StudyPlan, error)
		BulkUpdateByMaster(ctx context.Context, db database.QueryExecer, item *entities.StudyPlan) error
		BulkUpdateBook(ctx context.Context, db database.QueryExecer, spbs []*repositories.StudyPlanBook) error
		FindByIDs(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.TextArray) ([]*entities.StudyPlan, error)
		BulkCopy(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) ([]string, []string, error)
	}

	IStudyPlanItemRepository interface {
		BulkCopy(ctx context.Context, db database.QueryExecer, originalStudyPlanIDs pgtype.TextArray, newStudyPlanIDs pgtype.TextArray) error
		BulkSync(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlanItem) ([]*entities.StudyPlanItem, error)
		DeleteStudyPlanItemsByStudyPlans(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlanItem) error
		UpdateWithCopiedFromItem(ctx context.Context, db database.QueryExecer, studyPlanItems []*entities.StudyPlanItem) error
		UpdateSchoolDate(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, studentID pgtype.Text, schoolDate pgtype.Timestamptz) error
		UpdateStatus(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, studentID pgtype.Text, status pgtype.Text) error
		DeleteStudyPlanItemsByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error
	}

	ICourseStudyplanRepository interface {
		FindByCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray) ([]*entities.CourseStudyPlan, error)
		DeleteCourseStudyPlanBy(ctx context.Context, db database.QueryExecer, req *entities.CourseStudyPlan) error
		BulkUpsert(ctx context.Context, db database.QueryExecer, courseStudyPlans []*entities.CourseStudyPlan) error
	}

	IStudentStudyPlanRepository interface {
		DeleteStudentStudyPlans(ctx context.Context, db database.QueryExecer, studyPlanIds pgtype.TextArray) error
		BulkUpsert(ctx context.Context, db database.QueryExecer, studentStudyPlans []*entities.StudentStudyPlan) error
	}

	IAssignmentStudyPlanItemRepository interface {
		CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		BulkInsert(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem) error
		BulkUpsertByStudyPlanItem(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem) error
	}

	ILoStudyPlanItemRepository interface {
		CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		BulkInsert(ctx context.Context, db database.QueryExecer, loStudyPlanItems []*entities.LoStudyPlanItem) error
		DeleteLoStudyPlanItemsByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error
		DeleteLoStudyPlanItemsAndStudyPlanItemByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error
	}

	IStudentRepository interface {
		FindStudentsByCourseID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) (*pgtype.TextArray, error)
	}

	IAssignmentRepository interface {
		RetrieveAssignmentsByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.Assignment, error)
		RetrieveAssignments(ctx context.Context, db database.QueryExecer, assignmentIDs pgtype.TextArray) ([]*entities.Assignment, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, assignments []*entities.Assignment) error
	}

	ICourseBookRepository interface {
		FindByCourseIDAndBookID(ctx context.Context, db database.QueryExecer, bookID, courseID pgtype.Text) (*entities.CoursesBooks, error)
	}

	IStudentEventLogRepository interface {
		Create(ctx context.Context, db database.QueryExecer, ss []*entities.StudentEventLog) error
	}

	IBookRepository interface {
		FindByID(ctx context.Context, db database.QueryExecer, bookID pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Book, error)
		FindByIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) (map[string]*entities.Book, error)
		Upsert(ctx context.Context, db database.Ext, cc []*entities.Book) error
		RetrieveBookTreeByBookID(ctx context.Context, db database.QueryExecer, bookID pgtype.Text) ([]*repositories.BookTreeInfo, error)
		UpdateCurrentChapterDisplayOrder(ctx context.Context, db database.QueryExecer, totalGeneratedChapterDisplayOrder pgtype.Int4, bookID pgtype.Text) error
	}

	ILearningObjectiveRepository interface {
		RetrieveLearningObjectivesByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.LearningObjective, error)
	}
)

type StudyPlanModifierService struct {
	db                          database.Ext
	env                         string
	JSM                         nats.JetStreamManagement
	studyPlanRepo               IStudyPlanRepository
	courseStudyPlanRepo         ICourseStudyplanRepository
	studentStudyPlanRepo        IStudentStudyPlanRepository
	studyPlanItemRepo           IStudyPlanItemRepository
	assignmentStudyPlanItemRepo IAssignmentStudyPlanItemRepository
	loStudyPlanItemRepo         ILoStudyPlanItemRepository
	studentRepo                 IStudentRepository
	assignmentRepo              IAssignmentRepository
	learningObjectiveRepo       ILearningObjectiveRepository
	courseBookRepo              ICourseBookRepository
	bookRepo                    IBookRepository
	studentEventLogRepo         IStudentEventLogRepository
	internalModifierService     *services.InternalModifierService
}

func NewStudyPlanModifierService(db database.Ext, env string, jsm nats.JetStreamManagement) epb.StudyPlanModifierServiceServer {
	return &StudyPlanModifierService{
		db:                          db,
		env:                         env,
		JSM:                         jsm,
		courseStudyPlanRepo:         &repositories.CourseStudyPlanRepo{},
		studyPlanItemRepo:           &repositories.StudyPlanItemRepo{},
		studyPlanRepo:               &repositories.StudyPlanRepo{},
		studentStudyPlanRepo:        &repositories.StudentStudyPlanRepo{},
		assignmentStudyPlanItemRepo: &repositories.AssignmentStudyPlanItemRepo{},
		loStudyPlanItemRepo:         &repositories.LoStudyPlanItemRepo{},
		studentRepo:                 &repositories.StudentRepo{},
		assignmentRepo:              &repositories.AssignmentRepo{},
		learningObjectiveRepo:       &repositories.LearningObjectiveRepo{},
		courseBookRepo:              &repositories.CourseBookRepo{},
		bookRepo:                    &repositories.BookRepo{},
		studentEventLogRepo:         &repositories.StudentEventLogRepo{},
		internalModifierService: &services.InternalModifierService{
			BookRepo:                    &repositories.BookRepo{},
			AssignmentRepo:              &repositories.AssignmentRepo{},
			StudyPlanItemRepo:           &repositories.StudyPlanItemRepo{},
			AssignmentStudyPlanItemRepo: &repositories.AssignmentStudyPlanItemRepo{},
			LoStudyPlanItemRepo:         &repositories.LoStudyPlanItemRepo{},
			LearningObjectiveRepo:       &repositories.LearningObjectiveRepo{},
		},
	}
}

func (sp *StudyPlanModifierService) DeleteStudyPlanBelongsToACourse(ctx context.Context, req *epb.DeleteStudyPlanBelongsToACourseRequest) (*epb.DeleteStudyPlanBelongsToACourseResponse, error) {
	var courseIDs pgtype.TextArray
	if err := courseIDs.Set([]string{req.CourseId}); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("could not get course ID from request: %v", err))
	}
	exsitedStudyPlan, err := sp.courseStudyPlanRepo.FindByCourseIDs(ctx, sp.db, courseIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("courseStudyPlanRepo.FindByCourseIDs: %v", err))
	}

	if len(exsitedStudyPlan) == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("there are no any study plans belong to course %s", req.CourseId))
	}

	var courseID pgtype.Text
	var studyPlanID pgtype.Text
	if err := courseID.Set(req.CourseId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("course_id is not valid %v", err))
	}
	if err := studyPlanID.Set(req.StudyPlanId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("study_plan_id is not valid %v", err))
	}

	if err := database.ExecInTx(ctx, sp.db, func(ctx context.Context, tx pgx.Tx) error {
		if err := sp.courseStudyPlanRepo.DeleteCourseStudyPlanBy(ctx, tx, &entities.CourseStudyPlan{
			CourseID:    courseID,
			StudyPlanID: studyPlanID,
		}); err != nil {
			return fmt.Errorf("could not delete study plan belongs to course %v", err)
		}

		studyPlanIds, err := sp.studyPlanRepo.RecursiveSoftDeleteStudyPlanByStudyPlanIDInCourse(ctx, tx, studyPlanID)
		if err != nil {
			return fmt.Errorf("could not delete study plans %v", err)
		}

		var studyPlanIdsArg pgtype.TextArray
		if err := studyPlanIdsArg.Set(studyPlanIds); err != nil {
			return fmt.Errorf("study_plan_id is not valid %v", err)
		}
		if err := sp.studyPlanItemRepo.DeleteStudyPlanItemsByStudyPlans(ctx, tx, studyPlanIdsArg); err != nil {
			return fmt.Errorf("could not delete study plan items %v", err)
		}

		if err := sp.studentStudyPlanRepo.DeleteStudentStudyPlans(ctx, tx, studyPlanIdsArg); err != nil {
			return fmt.Errorf("could not delete student study plan %v", err)
		}

		return nil
	}); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("transaction for deleting error %v", err))
	}

	return &epb.DeleteStudyPlanBelongsToACourseResponse{}, nil
}

func (sp *StudyPlanModifierService) UpsertStudyPlanItemV2(ctx context.Context, req *epb.UpsertStudyPlanItemV2Request) (*epb.UpsertStudyPlanItemV2Response, error) {
	var ids []string
	r := &services.IAssignStudyPlan{
		StudyPlanRepo:               sp.studyPlanRepo,
		StudyPlanItemRepo:           sp.studyPlanItemRepo,
		AssignmentStudyPlanItemRepo: sp.assignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         sp.loStudyPlanItemRepo,
	}

	if err := sp.validateUpsertStudyPlanItemV2Request(req); err != nil {
		return nil, err
	}

	reqV1 := &epb.UpsertStudyPlanItemRequest{
		StudyPlanItems: req.StudyPlanItems,
	}
	err := database.ExecInTxWithRetry(ctx, sp.db, func(ctx context.Context, tx pgx.Tx) error {
		var errTx error
		ids, errTx = services.UpsertStudyPlanItemWithTx(ctx, reqV1, tx, r)
		if errTx != nil {
			return errTx
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("UpsertStudyPlanItem: %w", err)
	}
	return &epb.UpsertStudyPlanItemV2Response{
		StudyPlanItemIds: ids,
	}, nil
}

func (sp *StudyPlanModifierService) validateUpsertStudyPlanItemV2Request(req *epb.UpsertStudyPlanItemV2Request) error {
	for _, item := range req.StudyPlanItems {
		// validate dates, rule (available_from <= start_date <= end_date <= available_to)
		var times []time.Time

		if item.AvailableFrom != nil {
			times = append(times, item.AvailableFrom.AsTime())
		}
		if item.StartDate != nil {
			times = append(times, item.StartDate.AsTime())
		}
		if item.EndDate != nil {
			times = append(times, item.EndDate.AsTime())
		}
		if item.AvailableTo != nil {
			times = append(times, item.AvailableTo.AsTime())
		}
		for i := 0; i < len(times)-1; i++ {
			t0 := times[i]
			t1 := times[i+1]
			if t0.After(t1) {
				return fmt.Errorf("invalid dates, please follow rule available_from <= start_date <= end_date <= available_to")
			}
		}
	}

	return nil
}

func (sp *StudyPlanModifierService) UpdateStudyPlanItemsSchoolDate(ctx context.Context, req *epb.UpdateStudyPlanItemsSchoolDateRequest) (*epb.UpdateStudyPlanItemsSchoolDateResponse, error) {
	if len(req.StudyPlanItemIds) == 0 {
		return nil, fmt.Errorf("empty study plan item ids")
	}

	if req.StudentId == "" {
		return nil, fmt.Errorf("student id required")
	}

	ids := database.TextArray(req.StudyPlanItemIds)
	studentID := database.Text(req.StudentId)
	schoolDate := database.TimestamptzFromPb(req.SchoolDate)

	if err := sp.studyPlanItemRepo.UpdateSchoolDate(ctx, sp.db, ids, studentID, schoolDate); err != nil {
		return &epb.UpdateStudyPlanItemsSchoolDateResponse{
			IsSuccess: false,
		}, fmt.Errorf("studyPlanItemRepo.UpdateSchoolDate: %w", err)
	}

	return &epb.UpdateStudyPlanItemsSchoolDateResponse{
		IsSuccess: true,
	}, nil
}

func (sp *StudyPlanModifierService) UpdateStudyPlanItemsStatus(ctx context.Context, req *epb.UpdateStudyPlanItemsStatusRequest) (*epb.UpdateStudyPlanItemsStatusResponse, error) {
	if len(req.StudyPlanItemIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("empty study plan item ids").Error())
	}

	if req.StudentId == "" {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("student id required").Error())
	}

	if req.StudyPlanItemStatus == epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("status is invalid").Error())
	}

	ids := database.TextArray(req.StudyPlanItemIds)
	studentID := database.Text(req.StudentId)
	stat := database.Text(req.StudyPlanItemStatus.String())

	if err := sp.studyPlanItemRepo.UpdateStatus(ctx, sp.db, ids, studentID, stat); err != nil {
		return &epb.UpdateStudyPlanItemsStatusResponse{
			IsSuccess: false,
		}, status.Error(codes.Internal, fmt.Errorf("studyPlanItemRepo.UpdateStatus: %w", err).Error())
	}

	return &epb.UpdateStudyPlanItemsStatusResponse{
		IsSuccess: true,
	}, nil
}
func toCourseStudyPlanEn(courseID string, studyPlanID string) (*entities.CourseStudyPlan, error) {
	e := &entities.CourseStudyPlan{}
	now := timeutil.Now()
	database.AllNullEntity(e)
	err := multierr.Combine(
		e.StudyPlanID.Set(studyPlanID),
		e.CourseID.Set(courseID),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	return e, err
}

func toStudentStudyPlanEn(studentID pgtype.Text, studyPlanID string) (*entities.StudentStudyPlan, error) {
	e := &entities.StudentStudyPlan{}
	database.AllNullEntity(e)
	e.Now()
	e.StudentID = studentID

	err := e.StudyPlanID.Set(studyPlanID)
	return e, err
}

func toStudyPlanEntity(src *epb.UpsertStudyPlanRequest) (*entities.StudyPlan, error) {
	now := timeutil.Now()
	e := &entities.StudyPlan{}
	database.AllNullEntity(e)
	if err := multierr.Combine(
		e.ID.Set(idutil.ULIDNow()),
		e.Name.Set(src.Name),
		e.SchoolID.Set(src.SchoolId),
		e.CourseID.Set(src.CourseId),
		e.BookID.Set(src.BookId),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
		e.StudyPlanType.Set(epb.StudyPlanType_STUDY_PLAN_TYPE_COURSE.String()),
		e.Status.Set(src.Status.String()),
		e.TrackSchoolProgress.Set(src.TrackSchoolProgress),
		e.Grades.Set(src.Grades),
	); err != nil {
		return nil, fmt.Errorf("error create study plan: %w", err)
	}
	if len(src.Grades) == 0 {
		e.Grades.Set("{}")
	}
	return e, nil
}

var (
	ErrMissingStudyPlanID = fmt.Errorf("studyPlanId can't be null")
	ErrCantUpdateField    = fmt.Errorf("This field cannot be updated, only update name , grade , track school progress")
)
var (
	ErrMustHaveCourseID = fmt.Errorf("course id can't be null")
	ErrMustHaveBookID   = fmt.Errorf("book id can't be null")
)

func validateUpsertStudyPlan(req *epb.UpsertStudyPlanRequest) error {
	if req.CourseId == "" {
		return ErrMustHaveCourseID
	}
	if req.BookId == "" {
		return ErrMustHaveBookID
	}
	return nil
}

func (s *StudyPlanModifierService) verifyUpsertStudyPlanV2(ctx context.Context, req *epb.UpsertStudyPlanRequest) error {
	if req.BookId == "" {
		return status.Errorf(codes.InvalidArgument, "req must have book id")
	}
	if req.CourseId == "" {
		return status.Errorf(codes.InvalidArgument, "req must have course id")
	}

	if _, err := s.courseBookRepo.FindByCourseIDAndBookID(ctx, s.db, database.Text(req.BookId), database.Text(req.CourseId)); err != nil {
		if err.Error() == pgx.ErrNoRows.Error() {
			return status.Errorf(codes.InvalidArgument, "course id %s not found or book id %s not found", req.CourseId, req.BookId)
		}
		return status.Errorf(codes.Internal, fmt.Errorf("unable to retrieve course book by course id and book id: %w", err).Error())
	}
	return nil
}

func (s *StudyPlanModifierService) UpsertStudyPlan(ctx context.Context, req *epb.UpsertStudyPlanRequest) (*epb.UpsertStudyPlanResponse, error) {
	var studyPlanID string
	e, err := toStudyPlanEntity(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	if req.StudyPlanId != nil {
		studyPlanID = req.StudyPlanId.Value
		if _, err := s.studyPlanRepo.FindByID(ctx, s.db, database.Text(studyPlanID)); err != nil {
			if err.Error() == pgx.ErrNoRows.Error() {
				return nil, status.Errorf(codes.NotFound, "study plan id %v does not exists", studyPlanID)
			}
			return nil, status.Errorf(codes.Internal, fmt.Errorf("unable to retrieve study plan: %w", err).Error())
		}
		if err := e.ID.Set(studyPlanID); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "unable to set study plan id")
		}
		if err := s.studyPlanRepo.BulkUpdateByMaster(ctx, s.db, e); err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("unable to update study plan by master: %w", err).Error())
		}
		return &epb.UpsertStudyPlanResponse{
			StudyPlanId: studyPlanID,
		}, nil
	} else {
		studyPlanID = e.ID.String
	}
	if err := validateUpsertStudyPlan(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	if err := s.verifyUpsertStudyPlanV2(ctx, req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := database.ExecInTxWithRetry(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		studyPlans := []*entities.StudyPlan{e}
		studyPlanIDs := []string{studyPlanID}

		students, err := s.studentRepo.FindStudentsByCourseID(ctx, tx, database.Text(req.CourseId))

		if err != nil {
			if err.Error() != pgx.ErrNoRows.Error() {
				return fmt.Errorf("studentRepo.FindStudentsByCourseID: %w", err)
			} else {
				st := database.TextArray([]string{})
				students = &st
			}
		}

		ssps := make([]*entities.StudentStudyPlan, 0, len(students.Elements))
		for _, student := range students.Elements {
			studyPlan, err := toStudyPlanEntity(req)
			if err != nil {
				return err
			}
			if err := studyPlan.MasterStudyPlan.Set(studyPlanID); err != nil {
				return fmt.Errorf("unable to set master study plan: %w", err)
			}
			ssp, err := toStudentStudyPlanEn(student, studyPlan.ID.String)
			if err != nil {
				return fmt.Errorf("toStudentStudyPlanEn: %w", err)
			}
			ssps = append(ssps, ssp)
			studyPlans = append(studyPlans, studyPlan)
			studyPlanIDs = append(studyPlanIDs, studyPlan.ID.String)
		}

		if err := s.studyPlanRepo.BulkUpsert(ctx, tx, studyPlans); err != nil {
			return fmt.Errorf("studyPlanRepo.BulkUpsert: %w", err)
		}

		csp, err := toCourseStudyPlanEn(req.CourseId, studyPlanID)
		if err != nil {
			return fmt.Errorf("toCourseStudyPlanEn: %w", err)
		}
		if err := s.courseStudyPlanRepo.BulkUpsert(ctx, tx, []*entities.CourseStudyPlan{csp}); err != nil {
			return fmt.Errorf("courseStudyPlanRepo.BulkUpsert: %w", err)
		}
		if len(ssps) > 0 {
			if err := s.studentStudyPlanRepo.BulkUpsert(ctx, tx, ssps); err != nil {
				return fmt.Errorf("studentStudyPlan.BulkUpsert: %w", err)
			}
		}
		if err := services.UpsertStudyPlanItems(ctx, tx, req.BookId, req.CourseId, studyPlans, s.internalModifierService); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &epb.UpsertStudyPlanResponse{
		StudyPlanId: studyPlanID,
	}, nil
}

func toStudentEventLogEntity(ctx context.Context, p *epb.StudentEventLog) *entities.StudentEventLog {
	ep := new(entities.StudentEventLog)
	ep.ID.Set(nil)
	ep.StudentID.Set(interceptors.UserIDFromContext(ctx))
	ep.StudyPlanID.Set(nil)
	ep.LearningMaterialID.Set(nil)
	ep.EventID.Set(p.EventId)
	ep.EventType.Set(p.EventType)

	if p.CreatedAt == nil {
		ep.CreatedAt.Set(nil)
	} else {
		ep.CreatedAt.Set(p.CreatedAt.AsTime())
	}

	if p.Payload == nil {
		ep.Payload.Set(nil)
	} else {
		ep.Payload.Set(p.Payload)
	}

	return ep
}
