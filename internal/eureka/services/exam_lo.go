package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/eureka/constants"
	"github.com/manabie-com/backend/internal/eureka/entities"
	db "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type ExamLOService struct {
	sspb.UnimplementedExamLOServer
	DB database.Ext

	TopicRepo interface {
		RetrieveByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Topic, error)
		UpdateLODisplayOrderCounter(ctx context.Context, db database.QueryExecer, topicID pgtype.Text, number pgtype.Int4) error
	}

	QuizRepo interface {
		Search(ctx context.Context, db database.QueryExecer, filter repositories.QuizFilter) (entities.Quizzes, error)
		GetTagNames(ctx context.Context, db database.QueryExecer, externalIDs pgtype.TextArray) (map[string][]string, error)
	}

	LearningTimeCalculatorSvc interface {
		Calculate([]*entities.StudentEventLog) (learningTime time.Duration, completedAt *time.Time, err error)
	}

	ShuffledQuizSetRepo interface {
		Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.ShuffledQuizSet, error)
		ListExternalIDsFromSubmissionHistory(ctx context.Context, db database.QueryExecer, shuffleQuizSetIDs pgtype.TextArray, isAccepted bool) (map[string][]string, error)
		GetSeed(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Text, error)
		GetQuizIdx(context.Context, database.QueryExecer, pgtype.Text, pgtype.Text) (pgtype.Int4, error)
		GetExternalIDs(ctx context.Context, db database.QueryExecer, shuffleQuizSetID pgtype.Text) (pgtype.TextArray, error)
	}

	ExamLORepo interface {
		Insert(ctx context.Context, db database.QueryExecer, e *entities.ExamLO) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.ExamLO) error
		ListByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.ExamLO, error)
		ListExamLOBaseByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.ExamLOBase, error)
		UpsertGradeBookSetting(ctx context.Context, db database.QueryExecer, item *entities.GradeBookSetting) error
		Get(ctx context.Context, db database.QueryExecer, learningMaterialID pgtype.Text) (*entities.ExamLO, error)
	}

	ExamLOSubmissionRepo interface {
		Get(ctx context.Context, db database.QueryExecer, args *repositories.GetExamLOSubmissionArgs) (*entities.ExamLOSubmission, error)
		List(ctx context.Context, db database.QueryExecer, filter *repositories.ExamLOSubmissionFilter) (res []*repositories.ExtendedExamLOSubmission, _ error)
		ListByStudyPlanItemIdentities(ctx context.Context, db database.QueryExecer, studyPlanItemIdentities []*repositories.StudyPlanItemIdentity) ([]*entities.ExamLOSubmission, error)
		ListExamLOSubmissionWithDates(ctx context.Context, db database.QueryExecer, studyPlanItemIdentities []*repositories.StudyPlanItemIdentity) (res []*repositories.ExtendedExamLOSubmission, _ error)
		ListTotalGradePoints(ctx context.Context, db database.QueryExecer, submissionIDs pgtype.TextArray) (res []*repositories.ExamLOSubmissionWithGrade, _ error)
		Delete(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text) (int64, error)
		GetLatestSubmissionID(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text) (pgtype.Text, error)
		Update(ctx context.Context, db database.QueryExecer, e *entities.ExamLOSubmission) error
		GetTotalGradedPoint(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text) (pgtype.Int4, error)
		GetLatestExamLOSubmission(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text) (entities.ExamLOSubmission, error)
		GetInvalidIDsByBulkApproveReject(ctx context.Context, db database.QueryExecer, submissionIDs pgtype.TextArray, statusCond pgtype.TextArray) (pgtype.TextArray, error)
		BulkUpdateApproveReject(ctx context.Context, db database.QueryExecer, args *repositories.BulkUpdateApproveRejectArgs) (int, error)
	}

	ExamLOSubmissionAnswerRepo interface {
		List(ctx context.Context, db database.QueryExecer, filter *repositories.ExamLOSubmissionAnswerFilter) ([]*entities.ExamLOSubmissionAnswer, error)
		Delete(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text) (int64, error)
	}

	ExamLOSubmissionScoreRepo interface {
		List(ctx context.Context, db database.QueryExecer, filter *repositories.ExamLOSubmissionScoreFilter) ([]*entities.ExamLOSubmissionScore, error)
		Delete(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text) (int64, error)
		Upsert(ctx context.Context, db database.QueryExecer, e *entities.ExamLOSubmissionScore) (int, error)
	}

	StudentEventLogRepo interface {
		RetrieveStudentEventLogsByStudyPlanIdentities(context.Context, database.QueryExecer, []*repositories.StudyPlanItemIdentity) ([]*entities.StudentEventLog, error)
		DeleteByStudyPlanIdentities(ctx context.Context, db database.QueryExecer, args repositories.StudyPlanItemIdentity) (int64, error)
	}

	StudentRepo interface {
		FindStudentsByCourseID(context.Context, database.QueryExecer, pgtype.Text) (*pgtype.TextArray, error)
		FindStudentsByClassIDs(context.Context, database.QueryExecer, pgtype.TextArray) (*pgtype.TextArray, error)
		FindStudentsByCourseLocation(ctx context.Context, db database.QueryExecer, courseID pgtype.Text, locationID pgtype.TextArray) (*pgtype.TextArray, error)
		FindStudentsByLocation(ctx context.Context, db database.QueryExecer, locationIDs pgtype.TextArray) (*pgtype.TextArray, error)
	}
	StudyPlanItemRepo interface {
		UpdateCompletedAtToNullByStudyPlanItemIdentity(ctx context.Context, db database.QueryExecer, args repositories.StudyPlanItemIdentity) (int64, error)
	}

	QuestionTagRepo interface {
		GetPointPerTagBySubmissionID(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text) ([]repositories.GetPointPerTagBySubmissionIDData, error)
	}

	AllocateMarkerRepo interface {
		GetTeacherID(ctx context.Context, db database.QueryExecer, args *repositories.StudyPlanItemIdentity) (pgtype.Text, error)
	}

	QuestionGroupRepo interface {
		GetQuestionGroupsByIDs(ctx context.Context, db database.QueryExecer, ids ...string) (entities.QuestionGroups, error)
	}
}

func NewExamLOService(db database.Ext) sspb.ExamLOServer {
	return &ExamLOService{
		DB:                         db,
		TopicRepo:                  new(repositories.TopicRepo),
		ExamLORepo:                 new(repositories.ExamLORepo),
		StudentRepo:                new(repositories.StudentRepo),
		ExamLOSubmissionRepo:       new(repositories.ExamLOSubmissionRepo),
		ExamLOSubmissionAnswerRepo: new(repositories.ExamLOSubmissionAnswerRepo),
		ExamLOSubmissionScoreRepo:  new(repositories.ExamLOSubmissionScoreRepo),
		QuizRepo:                   new(repositories.QuizRepo),
		LearningTimeCalculatorSvc:  &LearningTimeCalculator{},
		ShuffledQuizSetRepo:        new(repositories.ShuffledQuizSetRepo),
		StudentEventLogRepo:        new(repositories.StudentEventLogRepo),
		StudyPlanItemRepo:          new(repositories.StudyPlanItemRepo),
		QuestionTagRepo:            new(repositories.QuestionTagRepo),
		AllocateMarkerRepo:         new(repositories.AllocateMarkerRepo),
		QuestionGroupRepo:          new(repositories.QuestionGroupRepo),
	}
}

func toExamLOEnt(req *sspb.ExamLOBase) (*entities.ExamLO, error) {
	e := &entities.ExamLO{}
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
		e.Type.Set(sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String()),
		e.IsPublished.Set(false),
		e.SetDefaultVendorType(),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
		e.Instruction.Set(req.Instruction),
		e.ManualGrading.Set(req.ManualGrading),
		e.ApproveGrading.Set(req.ApproveGrading),
		e.GradeCapping.Set(req.GradeCapping),
		e.ReviewOption.Set(req.GetReviewOption()),
	)
	if req.GradeToPass != nil {
		err = multierr.Append(
			err,
			e.GradeToPass.Set(req.GradeToPass.Value),
		)
	}
	if req.TimeLimit != nil {
		err = multierr.Append(err,
			e.TimeLimit.Set(req.TimeLimit.Value),
		)
	}
	if req.MaximumAttempt != nil {
		err = multierr.Append(err, e.MaximumAttempt.Set(req.MaximumAttempt.Value))
	}
	if err != nil {
		return nil, err
	}
	return e, nil
}

func ToBaseExamLO(e *entities.ExamLO) *sspb.ExamLOBase {
	return &sspb.ExamLOBase{
		Base: &sspb.LearningMaterialBase{
			LearningMaterialId: e.ID.String,
			TopicId:            e.TopicID.String,
			Name:               e.Name.String,
			Type:               e.Type.String,
			DisplayOrder: &wrapperspb.Int32Value{
				Value: int32(e.DisplayOrder.Int),
			},
		},
	}
}
func validateInsertExamLOReq(req *sspb.ExamLOBase) error {
	if req.Base.LearningMaterialId != "" {
		return fmt.Errorf("LearningMaterialId must be empty")
	}
	if req.Base.TopicId == "" {
		return fmt.Errorf("Topic ID must not be empty")
	}
	if ma := req.MaximumAttempt; ma != nil && (ma.GetValue() < 1 || ma.GetValue() > 99) {
		return fmt.Errorf("maximum_attempt must be null or between 1 to 99")
	}
	return nil
}

func (s *ExamLOService) InsertExamLO(ctx context.Context, req *sspb.InsertExamLORequest) (*sspb.InsertExamLOResponse, error) {
	if err := validateInsertExamLOReq(req.ExamLo); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("validateExamLOReq: %w", err).Error())
	}
	fc, err := toExamLOEnt(req.ExamLo)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("toExamLOEnt: %w", err).Error())
	}
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		tp, err := s.TopicRepo.RetrieveByID(ctx, tx, fc.TopicID, repositories.WithUpdateLock())
		if err != nil {
			return fmt.Errorf("s.TopicRepo.RetrieveByID: %w", err)
		}
		if err := fc.DisplayOrder.Set(tp.LODisplayOrderCounter.Int + 1); err != nil {
			return fmt.Errorf("fc.DisplayOrder.Set: %w", err)
		}
		if err := s.ExamLORepo.Insert(ctx, tx, fc); err != nil {
			return fmt.Errorf("s.ExamLORepo.Insert: %w", err)
		}
		if err := s.TopicRepo.UpdateLODisplayOrderCounter(ctx, tx, tp.ID, database.Int4(1)); err != nil {
			return fmt.Errorf("s.TopicRepo.UpdateLODisplayOrderCounter: %w", err)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}
	return &sspb.InsertExamLOResponse{
		LearningMaterialId: fc.LearningMaterial.ID.String,
	}, nil
}

func validateUpdateExamLOReq(req *sspb.ExamLOBase) error {
	if req.Base.LearningMaterialId == "" {
		return fmt.Errorf("LearningMaterialId must not be empty")
	}
	if ma := req.MaximumAttempt; ma != nil && (ma.GetValue() < 1 || ma.GetValue() > 99) {
		return fmt.Errorf("maximum_attempt must be null or between 1 to 99")
	}
	return nil
}

func (s *ExamLOService) UpdateExamLO(ctx context.Context, req *sspb.UpdateExamLORequest) (*sspb.UpdateExamLOResponse, error) {
	if err := validateUpdateExamLOReq(req.ExamLo); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("validateUpdateExamLOReq: %w", err).Error())
	}
	fc, err := toExamLOEnt(req.ExamLo)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("toExamLOEnt: %w", err).Error())
	}
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.ExamLORepo.Update(ctx, tx, fc); err != nil {
			return fmt.Errorf("s.ExamLORepo.Update: %w", err)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}
	return &sspb.UpdateExamLOResponse{}, nil
}

func (s *ExamLOService) ListExamLO(ctx context.Context, req *sspb.ListExamLORequest) (*sspb.ListExamLOResponse, error) {
	ids := req.LearningMaterialIds
	if len(ids) == 0 {
		return nil, status.Error(codes.InvalidArgument, "LearningMaterialIds must not be empty")
	}

	examLos, err := s.ExamLORepo.ListExamLOBaseByIDs(ctx, s.DB, database.TextArray(ids))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, fmt.Errorf("s.ExamLORepo.ListExamLOBaseByIDs: %w", err).Error())
		}
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ExamLORepo.ListExamLOBaseByIDs: %w", err).Error())
	}

	rspExamLos := make([]*sspb.ExamLOBase, 0, len(examLos))
	for _, elo := range examLos {
		var maximumAttempt *wrapperspb.Int32Value
		if elo.MaximumAttempt.Status == pgtype.Present {
			maximumAttempt = wrapperspb.Int32(elo.MaximumAttempt.Int)
		}

		rspExamLos = append(rspExamLos, &sspb.ExamLOBase{
			Base: &sspb.LearningMaterialBase{
				LearningMaterialId: elo.LearningMaterial.ID.String,
				TopicId:            elo.TopicID.String,
				Name:               elo.Name.String,
				Type:               elo.Type.String,
				DisplayOrder:       wrapperspb.Int32(int32(elo.DisplayOrder.Int)),
			},
			Instruction:    elo.Instruction.String,
			ManualGrading:  elo.ManualGrading.Bool,
			GradeToPass:    wrapperspb.Int32(elo.GradeToPass.Int),
			TimeLimit:      wrapperspb.Int32(elo.TimeLimit.Int),
			TotalQuestion:  elo.TotalQuestion.Int,
			MaximumAttempt: maximumAttempt,
			ApproveGrading: elo.ApproveGrading.Bool,
			GradeCapping:   elo.GradeCapping.Bool,
			ReviewOption:   sspb.ExamLOReviewOption(sspb.ExamLOReviewOption_value[elo.ReviewOption.String]),
		})
	}
	return &sspb.ListExamLOResponse{
		ExamLos: rspExamLos,
	}, nil
}

func (s *ExamLOService) ListHighestResultExamLOSubmission(ctx context.Context, req *sspb.ListHighestResultExamLOSubmissionRequest) (*sspb.ListHighestResultExamLOSubmissionResponse, error) {
	if len(req.StudyPlanItemIdentities) == 0 {
		return &sspb.ListHighestResultExamLOSubmissionResponse{}, nil
	}

	studyPlanItemIdentitiesArg := make([]*repositories.StudyPlanItemIdentity, 0, len(req.StudyPlanItemIdentities))

	for _, studyPlanItemIdentity := range req.StudyPlanItemIdentities {
		studyPlanItemIdentitiesArg = append(studyPlanItemIdentitiesArg, &repositories.StudyPlanItemIdentity{
			StudentID:          database.Text(studyPlanItemIdentity.StudentId.Value),
			StudyPlanID:        database.Text(studyPlanItemIdentity.StudyPlanId),
			LearningMaterialID: database.Text(studyPlanItemIdentity.LearningMaterialId),
		})
	}

	examLOSubmissions, err := s.ExamLOSubmissionRepo.ListByStudyPlanItemIdentities(ctx, s.DB, studyPlanItemIdentitiesArg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ExamLOSubmissionRepo.ListByStudyPlanItemIdentities: %w", err).Error())
	}

	mapStudyPlanItemIdentityResults := make(map[repositories.StudyPlanItemIdentity][]string)
	for _, examLOSubmission := range examLOSubmissions {
		key := repositories.StudyPlanItemIdentity{
			StudentID:          examLOSubmission.StudentID,
			StudyPlanID:        examLOSubmission.StudyPlanID,
			LearningMaterialID: examLOSubmission.LearningMaterialID,
		}
		mapStudyPlanItemIdentityResults[key] = append(mapStudyPlanItemIdentityResults[key], examLOSubmission.Result.String)
	}

	studyPlanItemResults := make([]*sspb.ListHighestResultExamLOSubmissionResponse_StudyPlanItemResult, 0, len(req.StudyPlanItemIdentities))

	for _, studyPlanItemIdentityReq := range req.StudyPlanItemIdentities {
		studyPlanItemResult := &sspb.ListHighestResultExamLOSubmissionResponse_StudyPlanItemResult{
			StudyPlanItemIdentity:        studyPlanItemIdentityReq,
			LatestExamLoSubmissionResult: sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_WAITING_FOR_GRADE,
		}
		studyPlanItemResults = append(studyPlanItemResults, studyPlanItemResult)

		studyPlanIdentityResults, ok := mapStudyPlanItemIdentityResults[repositories.StudyPlanItemIdentity{
			StudentID:          database.Text(studyPlanItemIdentityReq.StudentId.Value),
			StudyPlanID:        database.Text(studyPlanItemIdentityReq.StudyPlanId),
			LearningMaterialID: database.Text(studyPlanItemIdentityReq.LearningMaterialId),
		}]
		if !ok {
			continue
		}

		var containPassed, containFailed bool
		for _, result := range studyPlanIdentityResults {
			switch result {
			case sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_FAILED.String():
				containFailed = true
			case sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String():
				containPassed = true
			}
		}

		if containPassed {
			studyPlanItemResult.LatestExamLoSubmissionResult = sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED
		} else if containFailed {
			studyPlanItemResult.LatestExamLoSubmissionResult = sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_FAILED
		}
	}

	return &sspb.ListHighestResultExamLOSubmissionResponse{
		StudyPlanItemResults: studyPlanItemResults,
	}, nil
}

func (s *ExamLOService) getStudentIDsV2(ctx context.Context, req *sspb.ListExamLOSubmissionRequest) (*pgtype.TextArray, error) {
	if len(req.ClassIds) > 0 {
		return s.StudentRepo.FindStudentsByClassIDs(ctx, s.DB, database.TextArray(req.ClassIds))
	}

	if req.CourseId != nil && req.CourseId.Value != "" {
		return s.StudentRepo.FindStudentsByCourseLocation(ctx, s.DB, database.Text(req.CourseId.Value), database.TextArray(req.LocationIds))
	}

	return s.StudentRepo.FindStudentsByLocation(ctx, s.DB, database.TextArray(req.LocationIds))
}

// fetch the logs chunk by chunk to avoid Seq Scan
// because of too large amount of records will be over the effective_cache_size
// Postgres will choose Seq Scan rather than Index Scan
func (s *ExamLOService) retrieveStudentEventLogsConcurrently(ctx context.Context, studyPlanItemIdentities []*repositories.StudyPlanItemIdentity) ([]*entities.StudentEventLog, error) {
	return retrieveStudentEventLogsConcurrentlyByStudyPlanItemIdentities(ctx, s.DB, studyPlanItemIdentities, s.StudentEventLogRepo)
}

// nolint:errcheck
func (s *ExamLOService) ListExamLOSubmission(ctx context.Context, req *sspb.ListExamLOSubmissionRequest) (*sspb.ListExamLOSubmissionResponse, error) {
	if req.GetPaging() == nil {
		return nil, status.Error(codes.InvalidArgument, "empty paging")
	}

	filter := &repositories.ExamLOSubmissionFilter{
		Limit: uint(req.Paging.Limit),
	}

	if err := multierr.Combine(
		filter.OffsetID.Set(nil),
		filter.CreatedAt.Set(nil),
		filter.StudentIDs.Set(nil),
		filter.Statuses.Set(nil),
		filter.StartDate.Set(nil),
		filter.EndDate.Set(nil),
		filter.CourseID.Set(nil),
		filter.ClassIDs.Set(nil),
		filter.StudentName.Set(nil),
		filter.ExamName.Set(nil),
		filter.LocationIDs.Set(req.LocationIds),
		filter.CorrectorID.Set(nil),
		filter.SubmittedStartDate.Set(nil),
		filter.SubmittedEndDate.Set(nil),
		filter.UpdatedStartDate.Set(nil),
		filter.UpdatedEndDate.Set(nil),
		filter.SubmissionID.Set(nil),
	); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("unable to set value: %w", err).Error())
	}

	if offset := req.Paging.GetOffsetCombined().GetOffsetString(); offset != "" {
		filter.OffsetID.Set(offset)
	}
	if createdAt := req.Paging.GetOffsetCombined().GetOffsetTime(); createdAt != nil {
		filter.CreatedAt.Set(createdAt)
	}

	if len(req.Statuses) > 0 {
		ss := make([]string, 0, len(req.Statuses))
		for _, s := range req.Statuses {
			ss = append(ss, s.String())
		}

		filter.Statuses.Set(ss)
	}

	if req.Start != nil && req.Start.IsValid() {
		filter.StartDate.Set(req.Start.AsTime())
	}
	if req.End != nil && req.End.IsValid() {
		filter.EndDate.Set(req.End.AsTime())
	}

	if req.CourseId != nil && req.CourseId.Value != "" {
		filter.CourseID.Set(req.CourseId.Value)

		if len(req.ClassIds) > 0 {
			filter.ClassIDs.Set(req.ClassIds)
		}
	}

	if req.ExamName != nil && req.ExamName.Value != "" {
		cleanValue := db.ReplaceSpecialChars(req.ExamName.Value)
		filter.ExamName.Set(cleanValue)
	}

	if req.StudentName != nil && req.StudentName.Value != "" {
		cleanValue := db.ReplaceSpecialChars(req.StudentName.Value)
		filter.StudentName.Set(cleanValue)
	}

	if req.CorrectorId != nil && req.CorrectorId.Value != "" {
		filter.CorrectorID.Set(req.CorrectorId.Value)
	}

	if req.SubmittedDate != nil {
		if req.SubmittedDate.Start != nil && req.SubmittedDate.Start.IsValid() {
			filter.SubmittedStartDate.Set(req.SubmittedDate.Start.AsTime())
		}
		if req.SubmittedDate.End != nil && req.SubmittedDate.End.IsValid() {
			filter.SubmittedEndDate.Set(req.SubmittedDate.End.AsTime())
		}
	}
	if req.LastUpdatedDate != nil {
		if req.LastUpdatedDate.Start != nil && req.LastUpdatedDate.Start.IsValid() {
			filter.UpdatedStartDate.Set(req.LastUpdatedDate.Start.AsTime())
		}
		if req.LastUpdatedDate.End != nil && req.LastUpdatedDate.End.IsValid() {
			filter.UpdatedEndDate.Set(req.LastUpdatedDate.End.AsTime())
		}
	}
	if req.SubmissionId != nil && req.SubmissionId.Value != "" {
		cleanValue := db.ReplaceSpecialChars(req.SubmissionId.Value)
		filter.SubmissionID.Set(cleanValue)
	}
	studentIDs, err := s.getStudentIDsV2(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.getStudentIDsV2: %w", err).Error())
	}

	if studentIDs != nil {
		filter.StudentIDs = *studentIDs
	}
	extendedSubmissions, err := s.ExamLOSubmissionRepo.List(ctx, s.DB, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ExamLOSubmissionRepo.List: %w", err).Error())
	}

	items := make([]*sspb.ExamLOSubmission, 0, len(extendedSubmissions))
	for _, extendedSubmission := range extendedSubmissions {
		item := &sspb.ExamLOSubmission{
			SubmissionId: extendedSubmission.SubmissionID.String,
			StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
				StudentId:          wrapperspb.String(extendedSubmission.StudentID.String),
				StudyPlanId:        extendedSubmission.StudyPlanID.String,
				LearningMaterialId: extendedSubmission.LearningMaterialID.String,
			},
			ShuffledQuizSetId: extendedSubmission.ShuffledQuizSetID.String,
			SubmissionStatus:  sspb.SubmissionStatus(sspb.SubmissionStatus_value[extendedSubmission.Status.String]),
			SubmissionResult:  sspb.ExamLOSubmissionResult(sspb.ExamLOSubmissionResult_value[extendedSubmission.Result.String]),
			SubmittedAt:       timestamppb.New(extendedSubmission.CreatedAt.Time),
			UpdatedAt:         timestamppb.New(extendedSubmission.UpdatedAt.Time),
			StartDate:         timestamppb.New(extendedSubmission.StartDate.Time),
			EndDate:           timestamppb.New(extendedSubmission.EndDate.Time),
			CourseId:          extendedSubmission.CourseID.String,
			LastAction:        sspb.ApproveGradingAction(sspb.ApproveGradingAction_value[extendedSubmission.LastAction.String]),
			CorrectorId: &wrapperspb.StringValue{
				Value: extendedSubmission.CorrectorID.String,
			},
			MarkDate: timestamppb.New(extendedSubmission.MarkedAt.Time),
		}
		if extendedSubmission.LastActionAt.Status == pgtype.Present {
			item.LastActionAt = timestamppb.New(extendedSubmission.LastActionAt.Time)
		}
		if extendedSubmission.LastActionBy.Status == pgtype.Present {
			item.LastActionBy = wrapperspb.String(extendedSubmission.LastActionBy.String)
		}

		items = append(items, item)
	}

	paging := &cpb.Paging{
		Limit: req.Paging.Limit,
	}
	if len(extendedSubmissions) != 0 {
		paging.Offset = &cpb.Paging_OffsetCombined{
			OffsetCombined: &cpb.Paging_Combined{
				OffsetString: extendedSubmissions[len(extendedSubmissions)-1].SubmissionID.String,
				OffsetTime:   timestamppb.New(extendedSubmissions[len(extendedSubmissions)-1].CreatedAt.Time),
			},
		}
	}

	return &sspb.ListExamLOSubmissionResponse{
		NextPage: paging,
		Items:    items,
	}, nil
}

func (s *ExamLOService) ListExamLOSubmissionResult(ctx context.Context, req *sspb.ListExamLOSubmissionResultRequest) (*sspb.ListExamLOSubmissionResultResponse, error) {
	if len(req.StudyPlanItemIdentities) == 0 {
		return &sspb.ListExamLOSubmissionResultResponse{}, nil
	}

	studyPlanItemIdentitiesArg := make([]*repositories.StudyPlanItemIdentity, 0, len(req.StudyPlanItemIdentities))

	examLOIDs := make([]string, 0, len(req.StudyPlanItemIdentities))
	for _, studyPlanItemIdentity := range req.StudyPlanItemIdentities {
		studyPlanItemIdentitiesArg = append(studyPlanItemIdentitiesArg, &repositories.StudyPlanItemIdentity{
			StudentID:          database.Text(studyPlanItemIdentity.StudentId.Value),
			StudyPlanID:        database.Text(studyPlanItemIdentity.StudyPlanId),
			LearningMaterialID: database.Text(studyPlanItemIdentity.LearningMaterialId),
		})
		examLOIDs = append(examLOIDs, studyPlanItemIdentity.LearningMaterialId)
	}

	examLOSubmissionsWithDates, err := s.ExamLOSubmissionRepo.ListExamLOSubmissionWithDates(ctx, s.DB, studyPlanItemIdentitiesArg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ExamLOSubmissionRepo.ListExamLOSubmissionWithDates: %w", err).Error())
	}

	examLOs, err := s.ExamLORepo.ListByIDs(ctx, s.DB, database.TextArray(examLOIDs))
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ExamLORepo.ListByIDs: %w", err).Error())
	}
	examLOsMap := make(map[string]*entities.ExamLO)
	for _, examLO := range examLOs {
		examLOsMap[examLO.ID.String] = examLO
	}

	shuffledQuizSetIDs := make([]string, 0, len(examLOSubmissionsWithDates))
	examLOSubmissionIDs := make([]string, 0, len(examLOSubmissionsWithDates))
	examLOSubmissionsWithDatesMap := make(map[repositories.StudyPlanItemIdentity]map[string]*repositories.ExtendedExamLOSubmission)

	for _, examLOSubmissionWithDate := range examLOSubmissionsWithDates {
		studyPlanItemIdentity := repositories.StudyPlanItemIdentity{
			StudentID:          examLOSubmissionWithDate.StudentID,
			StudyPlanID:        examLOSubmissionWithDate.StudyPlanID,
			LearningMaterialID: examLOSubmissionWithDate.LearningMaterialID,
		}
		if _, ok := examLOSubmissionsWithDatesMap[studyPlanItemIdentity]; !ok {
			examLOSubmissionsWithDatesMap[studyPlanItemIdentity] = make(map[string]*repositories.ExtendedExamLOSubmission)
		}
		examLOSubmissionIDs = append(examLOSubmissionIDs, examLOSubmissionWithDate.SubmissionID.String)
		examLOSubmissionsWithDatesMap[studyPlanItemIdentity][examLOSubmissionWithDate.ShuffledQuizSetID.String] = examLOSubmissionWithDate
		shuffledQuizSetIDs = append(shuffledQuizSetIDs, examLOSubmissionWithDate.ShuffledQuizSetID.String)
	}

	examLOSubmissionWithGrades, err := s.ExamLOSubmissionRepo.ListTotalGradePoints(ctx, s.DB, database.TextArray(examLOSubmissionIDs))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.ExamLOSubmissionRepo.ListTotalGradePoints: %w", err).Error())
	}
	examLOSubmissionWithGradeMap := make(map[string]*repositories.ExamLOSubmissionWithGrade)
	for _, examLOSubmissionWithGrade := range examLOSubmissionWithGrades {
		examLOSubmissionWithGradeMap[examLOSubmissionWithGrade.SubmissionID.String] = examLOSubmissionWithGrade
	}

	// list shuffledQuizSets by shuffledQuizSetIDs
	shuffledQuizSets, err := s.ShuffledQuizSetRepo.Retrieve(ctx, s.DB, database.TextArray(shuffledQuizSetIDs))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.ShuffledQuizSetRepo.Retrieve: %w", err).Error())
	}

	// list studentEventLogs by studyPlanItemIdentities
	// group logs by studyPlanItemIdentity and SessionID
	studentEventLogsMap := make(map[repositories.StudyPlanItemIdentity]map[string][]*entities.StudentEventLog)
	for _, studyPlanItemIdentity := range studyPlanItemIdentitiesArg {
		studentEventLogsMap[*studyPlanItemIdentity] = make(map[string][]*entities.StudentEventLog)
	}

	studentEventLogs, err := s.retrieveStudentEventLogsConcurrently(ctx, studyPlanItemIdentitiesArg)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.retrieveStudentEventLogsConcurrently: %w", err).Error())
	}

	for _, studentEventLog := range studentEventLogs {
		payload := make(map[string]interface{})
		if err := studentEventLog.Payload.AssignTo(&payload); err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("log.Payload.AssignTo: %v", err).Error())
		}
		sessionID, ok := payload["session_id"].(string)
		if !ok {
			continue
		}
		studyPlanItemIdentity := &repositories.StudyPlanItemIdentity{
			StudentID:          database.Text(studentEventLog.StudentID.String),
			StudyPlanID:        database.Text(studentEventLog.StudyPlanID.String),
			LearningMaterialID: database.Text(studentEventLog.LearningMaterialID.String),
		}
		studentEventLogsMap[*studyPlanItemIdentity][sessionID] = append(studentEventLogsMap[*studyPlanItemIdentity][sessionID], studentEventLog)
	}

	// list map shuffledQuizSetID and externalQuizIDs -> externalIDsFromSubmissionHistoriesMap
	shuffledQuizSetsMap := make(map[repositories.StudyPlanItemIdentity][]*entities.ShuffledQuizSet)
	var originalShuffledQuizSetIDs []string
	originalShuffledQuizSetIDMap := make(map[string]bool)
	for _, shuffledQuizSet := range shuffledQuizSets {
		studyPlanItemIdentity := repositories.StudyPlanItemIdentity{
			StudentID:          shuffledQuizSet.StudentID,
			StudyPlanID:        shuffledQuizSet.StudyPlanID,
			LearningMaterialID: shuffledQuizSet.LearningMaterialID,
		}
		shuffledQuizSetsMap[studyPlanItemIdentity] = append(shuffledQuizSetsMap[studyPlanItemIdentity], shuffledQuizSet)

		if shuffledQuizSet.OriginalShuffleQuizSetID.Status != pgtype.Present {
			continue
		}

		if _, ok := originalShuffledQuizSetIDMap[shuffledQuizSet.OriginalShuffleQuizSetID.String]; !ok {
			originalShuffledQuizSetIDMap[shuffledQuizSet.OriginalShuffleQuizSetID.String] = true
			originalShuffledQuizSetIDs = append(originalShuffledQuizSetIDs, shuffledQuizSet.OriginalShuffleQuizSetID.String)
		}
	}

	externalIDsFromSubmissionHistoriesMap := make(map[string][]string)
	if len(originalShuffledQuizSetIDs) != 0 {
		externalIDsFromSubmissionHistoriesMap, err = s.ShuffledQuizSetRepo.ListExternalIDsFromSubmissionHistory(ctx, s.DB, database.TextArray(originalShuffledQuizSetIDs), false)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("s.ShuffledQuizSetRepo.ListExternalIDsFromSubmissionHistory: %w", err).Error())
		}
	}

	var totalAttempts int32
	items := make([]*sspb.ListExamLOSubmissionResultResponseItem, 0, len(examLOSubmissionsWithDates))
	highestQuestionScore := &sspb.HighestQuestionScore{}

	for studyPlanItemIdentity := range shuffledQuizSetsMap {
		item := &sspb.ListExamLOSubmissionResultResponseItem{
			StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
				StudyPlanId:        studyPlanItemIdentity.StudyPlanID.String,
				LearningMaterialId: studyPlanItemIdentity.LearningMaterialID.String,
				StudentId:          wrapperspb.String(studyPlanItemIdentity.StudentID.String),
			},
		}

		examLOSubmissionInfors := []*sspb.ExamLOSubmissionInfo{}
		for _, shuffledQuizSet := range shuffledQuizSetsMap[studyPlanItemIdentity] {
			logs := studentEventLogsMap[studyPlanItemIdentity][shuffledQuizSet.SessionID.String]
			learningTime, completedAt, err := s.LearningTimeCalculatorSvc.Calculate(logs)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Errorf("s.LearningTimeCalculatorSvc.Calculate: %v", err).Error())
			}

			createdAt := timestamppb.New(shuffledQuizSet.CreatedAt.Time)
			var completedAtPb *timestamppb.Timestamp
			if completedAt != nil {
				completedAtPb = timestamppb.New(*completedAt)
			}
			var totalQuiz int32
			totalQuiz = int32(len(shuffledQuizSet.QuizExternalIDs.Elements))
			if shuffledQuizSet.OriginalShuffleQuizSetID.Status == pgtype.Present {
				// only the origin attempt, not retry mode, because the default auto plus in below, so we have to sub
				totalAttempts--

				// because we only save the incorrect external_ids + new (external_ids which recently add from admin) to field quiz_external_ids
				// so when we use retry mode, it will handle missing the external_ids which done before,
				// we have to get all with distinct external_ids
				externalQuizIDs := externalIDsFromSubmissionHistoriesMap[shuffledQuizSet.OriginalShuffleQuizSetID.String]

				for _, e := range shuffledQuizSet.QuizExternalIDs.Elements {
					if e.Status == pgtype.Present {
						externalQuizIDs = append(externalQuizIDs, e.String)
					}
				}
				totalQuiz = int32(len(golibs.GetUniqueElementStringArray(externalQuizIDs)))
			}

			infor := &sspb.ExamLOSubmissionInfo{
				ShuffledQuizSetId: shuffledQuizSet.ID.String,
				CreatedAt:         createdAt,
				CompletedAt:       completedAtPb,
				TotalLearningTime: int64(learningTime.Seconds()),
			}

			var countHighestCrown bool
			examLOSubmissionWithDates := examLOSubmissionsWithDatesMap[studyPlanItemIdentity][shuffledQuizSet.ID.String]
			if examLOSubmissionWithDates != nil {
				infor.SubmissionId = examLOSubmissionWithDates.SubmissionID.String
				infor.SubmissionResult = sspb.ExamLOSubmissionResult(sspb.ExamLOSubmissionResult_value[examLOSubmissionWithDates.Result.String])
				infor.SubmissionStatus = sspb.SubmissionStatus(sspb.SubmissionStatus_value[examLOSubmissionWithDates.Status.String])
				infor.TotalPoint = wrapperspb.UInt32(uint32(examLOSubmissionWithDates.TotalPoint.Int))

				if examLOSubmissionGrade, ok := examLOSubmissionWithGradeMap[examLOSubmissionWithDates.SubmissionID.String]; ok {
					infor.TotalGradedPoint = wrapperspb.UInt32(uint32(examLOSubmissionGrade.TotalGradePoint.Int))
				}

				if examLOSubmissionWithDates.Status.String == sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String() {
					examLO := examLOsMap[examLOSubmissionWithDates.LearningMaterialID.String]
					if examLO.ReviewOption.String == sspb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_IMMEDIATELY.String() ||
						(examLO.ReviewOption.String == sspb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE.String() &&
							time.Now().After(examLOSubmissionWithDates.EndDate.Time)) {
						countHighestCrown = true
					}
				}
			}

			totalAttempts++
			if countHighestCrown {
				assignHighestQuestionScore(highestQuestionScore, shuffledQuizSet.TotalCorrectness.Int, totalQuiz)
			}
			examLOSubmissionInfors = append(examLOSubmissionInfors, infor)
		}
		// sort the item decs completed_at
		sortExamLOSubmissionInfo(examLOSubmissionInfors)
		item.ExamLoSubmissions = &sspb.ExamLOSubmissions{
			Items: examLOSubmissionInfors,
		}

		items = append(items, item)
	}

	var highestCrown cpb.AchievementCrown
	if highestQuestionScore != nil && highestQuestionScore.CorrectQuestion != nil {
		score := (float32(highestQuestionScore.GetCorrectQuestion().Value) / float32(highestQuestionScore.GetTotalQuestion())) * 100
		highestCrown = getAchievementCrownV1(score)
	}
	resp := &sspb.ListExamLOSubmissionResultResponse{
		Items:         items,
		HighestCrown:  highestCrown,
		HighestScore:  highestQuestionScore,
		TotalAttempts: totalAttempts,
	}

	return resp, nil
}

func assignHighestQuestionScore(currScore *sspb.HighestQuestionScore, totalCorrectness, totalQuiz int32) {
	if totalQuiz == 0 {
		return
	}
	if (currScore.GetTotalQuestion() == 0 && totalQuiz != 0) ||
		(currScore.CorrectQuestion != nil && float32(currScore.GetCorrectQuestion().Value)/float32(currScore.GetTotalQuestion()) < float32(totalCorrectness)/float32(totalQuiz)) {
		currScore.CorrectQuestion = wrapperspb.Int32(totalCorrectness)
		currScore.TotalQuestion = totalQuiz
	}
}

// sortExamLOSubmissionInfo will sort by completed_at, if two value both nil, will use created_at
func sortExamLOSubmissionInfo(examLOSubmissionInfors []*sspb.ExamLOSubmissionInfo) {
	sort.Slice(examLOSubmissionInfors, func(i, j int) bool {
		if examLOSubmissionInfors[i].CompletedAt == nil && examLOSubmissionInfors[j].CompletedAt == nil {
			return examLOSubmissionInfors[i].CreatedAt.AsTime().After(examLOSubmissionInfors[j].CreatedAt.AsTime())
		}
		return examLOSubmissionInfors[i].CompletedAt.AsTime().After(examLOSubmissionInfors[j].CompletedAt.AsTime())
	})
}

func (s *ExamLOService) ListExamLOSubmissionScore(ctx context.Context, req *sspb.ListExamLOSubmissionScoreRequest) (*sspb.ListExamLOSubmissionScoreResponse, error) {
	if req.SubmissionId == "" {
		return nil, status.Error(codes.InvalidArgument, "empty submission_id")
	}
	if req.ShuffledQuizSetId == "" {
		return nil, status.Error(codes.InvalidArgument, "empty shuffled_quiz_set_id")
	}

	submission, err := s.ExamLOSubmissionRepo.Get(ctx, s.DB, &repositories.GetExamLOSubmissionArgs{
		SubmissionID:      database.Text(req.SubmissionId),
		ShuffledQuizSetID: database.Text(req.ShuffledQuizSetId),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.ExamLOSubmissionRepo.Get: %w", err).Error())
	}

	submissionScores, err := s.ExamLOSubmissionScoreRepo.List(ctx, s.DB, &repositories.ExamLOSubmissionScoreFilter{
		SubmissionID:      database.Text(req.SubmissionId),
		ShuffledQuizSetID: database.Text(req.ShuffledQuizSetId),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.ExamLOSubmissionScoreRepo.List: %w", err).Error())
	}

	type submissionScoreKey struct {
		SubmissionID      string
		ShuffledQuizSetID string
		ExternalQuizID    string
	}
	submissionAnswers, err := s.ExamLOSubmissionAnswerRepo.List(ctx, s.DB, &repositories.ExamLOSubmissionAnswerFilter{
		SubmissionID:      database.Text(req.SubmissionId),
		ShuffledQuizSetID: database.Text(req.ShuffledQuizSetId),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.ExamLOSubmissionAnswerRepo.List: %w", err).Error())
	}

	submissionScoresMap := make(map[submissionScoreKey]*entities.ExamLOSubmissionScore)

	externalQuizIDs := make([]string, 0, len(submissionAnswers))
	for _, submissionAnswer := range submissionAnswers {
		externalQuizIDs = append(externalQuizIDs, submissionAnswer.QuizID.String)
	}

	for _, submissionScore := range submissionScores {
		key := submissionScoreKey{
			SubmissionID:      submissionScore.SubmissionID.String,
			ShuffledQuizSetID: submissionScore.ShuffledQuizSetID.String,
			ExternalQuizID:    submissionScore.QuizID.String,
		}
		submissionScoresMap[key] = submissionScore
	}

	quizzes, err := s.QuizRepo.Search(ctx, s.DB, repositories.QuizFilter{
		ExternalIDs: database.TextArray(externalQuizIDs),
		Status: pgtype.Text{
			Status: pgtype.Null,
		},
		Limit: uint(len(externalQuizIDs)),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.QuizRepo.Search: %w", err).Error())
	}
	tagsPerQuiz, err := s.QuizRepo.GetTagNames(ctx, s.DB, database.TextArray(externalQuizIDs))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.QuizRepo.GetTagNames: %w", err).Error())
	}
	quizzesMap := make(map[string]*entities.Quiz)
	for _, quiz := range quizzes {
		quizzesMap[quiz.ExternalID.String] = quiz
	}

	var totalGradedPoint uint32
	examLoSubmissionScores := make([]*sspb.ExamLOSubmissionScore, 0, len(submissionAnswers))

	for _, submissionAnswer := range submissionAnswers {
		quiz := quizzesMap[submissionAnswer.QuizID.String]
		quizCore, err := toQuizCore(quiz)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("toQuizCore: %w", err).Error())
		}
		if tags, ok := tagsPerQuiz[submissionAnswer.QuizID.String]; ok {
			quizCore.TagNames = append(quizCore.TagNames, tags...)
		}
		originalOption := make([]*cpb.QuizOption, len(quizCore.Options))
		copy(originalOption, quizCore.Options)
		if quizCore.Kind == cpb.QuizType_QUIZ_TYPE_MCQ ||
			quizCore.Kind == cpb.QuizType_QUIZ_TYPE_MAQ ||
			quizCore.Kind == cpb.QuizType_QUIZ_TYPE_ORD {
			seedStr, err := s.ShuffledQuizSetRepo.GetSeed(ctx, s.DB, database.Text(req.ShuffledQuizSetId))
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Errorf("ExamLOService.ListExamLOSubmissionScore.ShuffledQuizSetRepo.GetSeed: %v", err).Error())
			}
			seed, err := strconv.ParseInt(seedStr.String, 10, 64)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Errorf("ExamLOService.ListExamLOSubmissionScore.CannotParseSeed: %v", err).Error())
			}
			idx, err := s.ShuffledQuizSetRepo.GetQuizIdx(ctx, s.DB, database.Text(req.ShuffledQuizSetId), quiz.ExternalID)
			if err != nil || idx.Int == 0 {
				return nil, status.Error(codes.Internal, fmt.Errorf("ExamLOService.ListExamLOSubmissionScore.GetQuizIdx: %v", err).Error())
			}
			r := rand.New(rand.NewSource(seed + int64(idx.Int)))
			r.Shuffle(len(quizCore.Options), func(i, j int) { quizCore.Options[i], quizCore.Options[j] = quizCore.Options[j], quizCore.Options[i] })
		}

		correctIdx := make([]uint32, 0, len(quizCore.Options))
		correctText := make([]string, 0, len(quizCore.Options))
		result := (&sspb.ExamLOSubmissionScore{}).Result
		switch quizCore.Kind {
		case cpb.QuizType_QUIZ_TYPE_MCQ, cpb.QuizType_QUIZ_TYPE_MAQ, cpb.QuizType_QUIZ_TYPE_MIQ:
			for idx, opt := range quizCore.Options {
				if opt.Correctness {
					// must increase index before append because fe and me sides use index with start value is 1 instead of 0 in be side
					correctIdx = append(correctIdx, uint32(idx+1))
				}
			}
		case cpb.QuizType_QUIZ_TYPE_FIB, cpb.QuizType_QUIZ_TYPE_POW, cpb.QuizType_QUIZ_TYPE_TAD:
			var optionsEnt []*entities.QuizOption
			err := quiz.Options.AssignTo(&optionsEnt)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Errorf("ExamLOService.ListExamLOSubmissionScore.QuizOption.AssignTo: %v", err).Error())
			}
			for _, opt := range optionsEnt {
				correctText = append(correctText, opt.GetText())
			}
		case cpb.QuizType_QUIZ_TYPE_ORD:
			correctKeys := make([]string, 0, len(quizCore.Options))
			for _, opt := range originalOption {
				correctKeys = append(correctKeys, opt.Key)
			}

			var submittedKeys []string
			if submissionAnswer.SubmittedKeysAnswer.Status == pgtype.Present {
				if err = submissionAnswer.SubmittedKeysAnswer.AssignTo(&submittedKeys); err != nil {
					return nil, status.Error(codes.Internal, fmt.Errorf("got error when get submitted keys answer: %w", err).Error())
				}
			}
			result = &sspb.ExamLOSubmissionScore_OrderingResult{
				OrderingResult: &cpb.OrderingResult{
					CorrectKeys:   correctKeys,
					SubmittedKeys: submittedKeys,
				},
			}
		case cpb.QuizType_QUIZ_TYPE_ESQ:
		default:
			return nil, status.Error(codes.FailedPrecondition, "Not supported quiz type!")
		}

		var (
			selectedIndex  []uint32
			filledText     []string
			correctness    []bool
			teacherComment string
		)
		if err := multierr.Combine(
			submissionAnswer.StudentIndexAnswer.AssignTo(&selectedIndex),
			submissionAnswer.StudentTextAnswer.AssignTo(&filledText),
			submissionAnswer.IsCorrect.AssignTo(&correctness),
		); err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("multierr.Combine: %w", err).Error())
		}

		// point student gained for that question, if there is no record in score table -> get from answer table
		gradedPoint := uint32(submissionAnswer.Point.Int)
		key := submissionScoreKey{
			SubmissionID:      submissionAnswer.SubmissionID.String,
			ShuffledQuizSetID: submissionAnswer.ShuffledQuizSetID.String,
			ExternalQuizID:    submissionAnswer.QuizID.String,
		}
		submissionScore, ok := submissionScoresMap[key]
		if ok {
			gradedPoint = uint32(submissionScore.Point.Int)
			teacherComment = submissionScore.TeacherComment.String
		}

		examLOSubmissionScore := &sspb.ExamLOSubmissionScore{
			ShuffleQuizSetId: submissionAnswer.ShuffledQuizSetID.String,
			QuizType:         cpb.QuizType(cpb.QuizType_value[quiz.Kind.String]),
			FilledText:       filledText,
			CorrectText:      correctText,
			SelectedIndex:    selectedIndex,
			CorrectIndex:     correctIdx,
			Correctness:      correctness,
			IsAccepted:       submissionAnswer.IsAccepted.Bool,
			Core:             quizCore,
			TeacherComment:   teacherComment,
			GradedPoint:      wrapperspb.UInt32(gradedPoint),
			Point:            wrapperspb.UInt32(uint32(submissionAnswer.Point.Int)),
			Result:           result,
		}
		examLoSubmissionScores = append(examLoSubmissionScores, examLOSubmissionScore)

		totalGradedPoint += gradedPoint
	}

	// get list question group
	qr, err := getQuestionGroupByQuiz(ctx, s.QuestionGroupRepo, s.DB, quizzes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	respQuestionGroups, err := entities.QuestionGroupsToQuestionGroupProtoBufMess(qr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	resp := &sspb.ListExamLOSubmissionScoreResponse{
		SubmissionScores: examLoSubmissionScores,
		SubmissionStatus: sspb.SubmissionStatus(sspb.SubmissionStatus_value[submission.Status.String]),
		SubmissionResult: sspb.ExamLOSubmissionResult(sspb.ExamLOSubmissionResult_value[submission.Result.String]),
		TeacherFeedback:  submission.TeacherFeedback.String,
		TotalGradedPoint: wrapperspb.UInt32(totalGradedPoint),
		TotalPoint:       wrapperspb.UInt32(uint32(submission.TotalPoint.Int)),
		QuestionGroups:   respQuestionGroups,
	}
	return resp, nil
}

func validateDeleteExamLOReq(req *sspb.DeleteExamLOSubmissionRequest) error {
	if req.SubmissionId == "" {
		return fmt.Errorf("cannot empty submission_id")
	}
	return nil
}

func (s *ExamLOService) DeleteExamLOSubmission(ctx context.Context, req *sspb.DeleteExamLOSubmissionRequest) (*sspb.DeleteExamLOSubmissionResponse, error) {
	if err := validateDeleteExamLOReq(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("validateDeleteExamLOReq, err: %w", err).Error())
	}

	// Get the latest exam lo submission combine
	latestExamLOSubmission, err := s.ExamLOSubmissionRepo.GetLatestExamLOSubmission(ctx, s.DB, database.Text(req.SubmissionId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("S.DeleteExamLOSubmission.GetLatestExamLOSubmission, err: %w", err).Error())
	}

	// Check if id exam lo submission is the latest latest = request
	if latestExamLOSubmission.SubmissionID.String != req.SubmissionId {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("S.DeleteExamLOSubmission, err: exam lo submission not the latest, expected %s, get %s", latestExamLOSubmission.SubmissionID.String, req.SubmissionId).Error())
	}

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Delete exam lo submission
		if _, errTx := s.ExamLOSubmissionRepo.Delete(ctx, tx, database.Text(req.SubmissionId)); errTx != nil {
			return fmt.Errorf("s.ExamLOSubmissionRepo.Delete, err: %w", errTx)
		}

		// Delete exam lo submission answer
		if _, errTx := s.ExamLOSubmissionAnswerRepo.Delete(ctx, tx, database.Text(req.SubmissionId)); errTx != nil {
			return fmt.Errorf("s.ExamLOSubmissionAnswerRepo.Delete, err: %w", errTx)
		}

		// Delete exam lo submission score
		if _, errTx := s.ExamLOSubmissionScoreRepo.Delete(ctx, tx, database.Text(req.SubmissionId)); errTx != nil {
			return fmt.Errorf("s.ExamLOSubmissionScoreRepo.Delete, err: %w", errTx)
		}

		// Check if combination studyPlanID, learningMaterialID, studentID still exist
		_, errTx := s.ExamLOSubmissionRepo.GetLatestExamLOSubmission(ctx, tx, database.Text(req.SubmissionId))
		if errTx != nil && !errors.Is(errTx, pgx.ErrNoRows) {
			return fmt.Errorf("S.DeleteExamLOSubmission.GetLatestExamLOSubmission check combination err: %w", errTx)
		}
		if errTx != nil && errors.Is(errTx, pgx.ErrNoRows) { // no combination left{ // no combination left
			// Update studyPlanItem completed to null
			_, errInTx := s.StudyPlanItemRepo.UpdateCompletedAtToNullByStudyPlanItemIdentity(ctx, tx, repositories.StudyPlanItemIdentity{
				StudentID:          latestExamLOSubmission.StudentID,
				StudyPlanID:        latestExamLOSubmission.StudyPlanID,
				LearningMaterialID: latestExamLOSubmission.LearningMaterialID,
			})
			if errInTx != nil {
				return fmt.Errorf("S.DeleteExamLOSubmission.UpdateCompletedAtToNullByStudyPlanItemIdentity, err: %w", errInTx)
			}

			// soft delete student event logs relate
			rowsAffectedLog, errInTx := s.StudentEventLogRepo.DeleteByStudyPlanIdentities(ctx, tx, repositories.StudyPlanItemIdentity{
				StudentID:          latestExamLOSubmission.StudentID,
				StudyPlanID:        latestExamLOSubmission.StudyPlanID,
				LearningMaterialID: latestExamLOSubmission.LearningMaterialID,
			})
			if errInTx != nil {
				return fmt.Errorf("S.DeleteExamLOSubmission.DeleteByStudyPlanIdentities, err: %w", errInTx)
			}
			if rowsAffectedLog == 0 {
				return fmt.Errorf("S.DeleteExamLOSubmission.DeleteByStudyPlanIdentities, not found any row to delete with this lm id: %s, student id: %s and study plan id: %s", latestExamLOSubmission.LearningMaterialID.String, latestExamLOSubmission.StudentID.String, latestExamLOSubmission.StudyPlanID.String)
			}
		}
		return nil
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &sspb.DeleteExamLOSubmissionResponse{}, nil
}
func (s *ExamLOService) UpsertGradeBookSetting(ctx context.Context, req *sspb.UpsertGradeBookSettingRequest) (*sspb.UpsertGradeBookSettingResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)
	now := time.Now()

	if err := s.ExamLORepo.UpsertGradeBookSetting(ctx, s.DB, &entities.GradeBookSetting{
		BaseEntity: entities.BaseEntity{
			CreatedAt: database.Timestamptz(now),
			UpdatedAt: database.Timestamptz(now),
			DeletedAt: pgtype.Timestamptz{
				Status: pgtype.Null,
			},
		},
		Setting:   database.Text(req.GetSetting().String()),
		UpdatedBy: database.Text(userID),
	}); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ExamLORepo.UpsertGradeBookSetting: %w", err).Error())
	}

	return &sspb.UpsertGradeBookSettingResponse{}, nil
}

func validateUpsertLOProgressionRequest(req *sspb.UpsertLOProgressionRequest) error {
	if req.ShuffledQuizSetId == "" {
		return errors.New("req must have ShuffledQuizSetId")
	}

	if req.StudyPlanItemIdentity == nil {
		return errors.New("req must have StudyPlanItemIdentity")
	}

	if req.StudyPlanItemIdentity.LearningMaterialId == "" {
		return errors.New("req must have LearningMaterialId")
	}

	if req.StudyPlanItemIdentity.StudentId == nil || req.StudyPlanItemIdentity.StudentId.Value == "" {
		return errors.New("req must have StudentId")
	}

	if req.StudyPlanItemIdentity.StudyPlanId == "" {
		return errors.New("req must have StudyPlanId")
	}

	if req.SessionId == "" {
		return errors.New("req must have session_id")
	}

	return nil
}

func (s *ExamLOService) GradeAManualGradingExamSubmission(ctx context.Context, req *sspb.GradeAManualGradingExamSubmissionRequest) (*sspb.GradeAManualGradingExamSubmissionResponse, error) {
	if err := validateGradeAManualGradingExamSubmissionRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	latestSubmissionID, err := s.ExamLOSubmissionRepo.GetLatestSubmissionID(ctx, s.DB, database.Text(req.SubmissionId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.ExamLOSubmissionRepo.GetLatestSubmissionID: %w", err).Error())
	}
	if req.SubmissionId != latestSubmissionID.String {
		return nil, status.Error(codes.FailedPrecondition, errors.New("cannot grade the old submission").Error())
	}

	submission, err := s.ExamLOSubmissionRepo.Get(ctx, s.DB, &repositories.GetExamLOSubmissionArgs{
		SubmissionID:      database.Text(req.SubmissionId),
		ShuffledQuizSetID: database.Text(req.ShuffledQuizSetId),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.ExamLOSubmissionRepo.Get: %w", err).Error())
	}

	examLO, err := s.ExamLORepo.Get(ctx, s.DB, submission.LearningMaterialID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.ExamLORepo.Get: %w", err).Error())
	}

	var isAssigned bool
	teacherID, err := s.AllocateMarkerRepo.GetTeacherID(ctx, s.DB, &repositories.StudyPlanItemIdentity{
		StudentID:          submission.StudentID,
		StudyPlanID:        submission.StudyPlanID,
		LearningMaterialID: submission.LearningMaterialID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.AllocateMarkerRepo.GetTeacherID: %w", err).Error())
	}
	if teacherID.String == interceptors.UserIDFromContext(ctx) {
		isAssigned = true
	}

	if err = validateStatusChangeApproveGradingSetup(&ApproveGradingSetup{
		ApproveGrading: examLO.ApproveGrading.Bool,
		Role:           interceptors.UserGroupFromContext(ctx),
		IsAssigned:     isAssigned,
		Status:         submission.Status.String,
		StatusChange:   req.SubmissionStatus.String(),
	}); err != nil {
		return nil, status.Error(codes.FailedPrecondition, fmt.Errorf("validateStatusChangeApproveGradingSetup: %w", err).Error())
	}

	var totalGradedPoint uint32
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		now := time.Now()
		for _, item := range req.TeacherExamGrades {
			score := &entities.ExamLOSubmissionScore{}
			database.AllNullEntity(score)

			err := multierr.Combine(
				score.SubmissionID.Set(req.SubmissionId),
				score.QuizID.Set(item.QuizId),
				score.TeacherID.Set(interceptors.UserIDFromContext(ctx)),
				score.TeacherComment.Set(item.TeacherComment),
				score.ShuffledQuizSetID.Set(req.ShuffledQuizSetId),
				score.CreatedAt.Set(now),
				score.UpdatedAt.Set(now),
			)

			if item.TeacherPointGiven != nil {
				err = multierr.Append(err, score.Point.Set(item.TeacherPointGiven.Value))
			}

			if len(item.Correctness) > 0 {
				err = multierr.Append(err, multierr.Combine(
					score.IsAccepted.Set(item.IsAccepted),
					score.IsCorrect.Set(item.Correctness),
				))
			} else {
				score.IsCorrect = pgtype.BoolArray{
					Elements: []pgtype.Bool{},
					Status:   pgtype.Present,
				}
			}

			if err != nil {
				return fmt.Errorf("can not set up submission score: %w", err)
			}

			if _, err = s.ExamLOSubmissionScoreRepo.Upsert(ctx, tx, score); err != nil {
				return fmt.Errorf("ExamLOSubmissionScoreRepo.Upsert: %w", err)
			}
		}

		resultTotalGradedPoint, err := s.ExamLOSubmissionRepo.GetTotalGradedPoint(ctx, tx, submission.SubmissionID)
		if err != nil {
			return fmt.Errorf("ExamLOSubmissionRepo.GetTotalGradedPoint: %w", err)
		}
		totalGradedPoint = uint32(resultTotalGradedPoint.Int)

		switch req.SubmissionStatus {
		case sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED:
			if examLO.GradeToPass.Status == pgtype.Null {
				submission.Result = database.Text(sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_COMPLETED.String())
			} else {
				if resultTotalGradedPoint.Int >= examLO.GradeToPass.Int {
					submission.Result = database.Text(sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String())
				} else {
					submission.Result = database.Text(sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_FAILED.String())
				}
			}
		case sspb.SubmissionStatus_SUBMISSION_STATUS_MARKED:
			submission.MarkedAt = database.Timestamptz(now)
		}
		submission.Status = database.Text(req.SubmissionStatus.String())
		submission.TeacherFeedback = database.Text(req.TeacherFeedback)
		submission.TeacherID = database.Text(interceptors.UserIDFromContext(ctx))
		submission.UpdatedAt = database.Timestamptz(now)

		err = s.ExamLOSubmissionRepo.Update(ctx, tx, submission)
		if err != nil {
			return fmt.Errorf("s.ExamLOSubmissionRepo.Update: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &sspb.GradeAManualGradingExamSubmissionResponse{
		TotalGradedPoint: wrapperspb.UInt32(totalGradedPoint),
	}

	return resp, nil
}

func validateGradeAManualGradingExamSubmissionRequest(req *sspb.GradeAManualGradingExamSubmissionRequest) error {
	if req.SubmissionId == "" {
		return errors.New("req must have SubmissionId")
	}

	if req.ShuffledQuizSetId == "" {
		return errors.New("req must have ShuffledQuizSetId")
	}

	return nil
}

func validateStatusChangeApproveGradingSetup(setup *ApproveGradingSetup) error {
	statusChange := &StatusChangeMessageHandler{}
	approveGrading := &ApproveGradingMessageHandler{Next: statusChange}
	isAssigned := &IsAssignedMessageHandler{Next: approveGrading}
	role := &RoleMessageHandler{Next: isAssigned}

	return role.GetErrorMessage(setup)
}

type ApproveGradingSetup struct {
	ApproveGrading bool
	Role           string
	IsAssigned     bool
	Status         string
	StatusChange   string
}
type ErrorMessageHandler interface {
	GetErrorMessage(input *ApproveGradingSetup) error
}

type RoleMessageHandler struct {
	Next ErrorMessageHandler
}

func (r *RoleMessageHandler) GetErrorMessage(input *ApproveGradingSetup) error {
	switch input.Role {
	case constants.RoleTeacher, constants.RoleHQStaff, constants.RoleSchoolAdmin, constants.RoleCentreManager:
		return r.Next.GetErrorMessage(input)
	}
	return fmt.Errorf("role %s does not allow change information", input.Role)
}

type IsAssignedMessageHandler struct {
	Next ErrorMessageHandler
}

func (r *IsAssignedMessageHandler) GetErrorMessage(input *ApproveGradingSetup) error {
	if input.IsAssigned {
		return r.Next.GetErrorMessage(input)
	}
	return fmt.Errorf("user is not assigned to change the information")
}

type ApproveGradingMessageHandler struct {
	Next ErrorMessageHandler
}

func (r *ApproveGradingMessageHandler) GetErrorMessage(input *ApproveGradingSetup) error {
	if input.ApproveGrading {
		return r.Next.GetErrorMessage(input)
	}
	return nil
}

type StatusChangeMessageHandler struct {
	Next ErrorMessageHandler
}

func (r *StatusChangeMessageHandler) GetErrorMessage(input *ApproveGradingSetup) error {
	switch input.Status {
	case sspb.SubmissionStatus_SUBMISSION_STATUS_NOT_MARKED.String():
		if input.StatusChange == sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String() {
			return fmt.Errorf("changing %s status to %s status is not allowed", input.Status, input.StatusChange)
		}
	case sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS.String():
		if input.StatusChange == sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String() {
			return fmt.Errorf("changing %s status to %s status is not allowed", input.Status, input.StatusChange)
		}
	case sspb.SubmissionStatus_SUBMISSION_STATUS_MARKED.String(), sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String():
		return fmt.Errorf("user is not assigned to change the information")
	}
	return nil
}

func (s *ExamLOService) BulkApproveRejectSubmission(ctx context.Context, req *sspb.BulkApproveRejectSubmissionRequest) (*sspb.BulkApproveRejectSubmissionResponse, error) {
	if len(req.SubmissionIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, errors.New("req must have SubmissionIds").Error())
	}

	var statusChange string
	statusCond := make([]string, 0)
	respIDs := make([]string, 0)

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		switch req.ApproveGradingAction {
		case sspb.ApproveGradingAction_APPROVE_ACTION_APPROVED:
			statusChange = sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String()
			statusCond = append(statusCond, sspb.SubmissionStatus_SUBMISSION_STATUS_MARKED.String())
		case sspb.ApproveGradingAction_APPROVE_ACTION_REJECTED:
			statusChange = sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS.String()
			statusCond = append(statusCond, sspb.SubmissionStatus_SUBMISSION_STATUS_MARKED.String(),
				sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String())
		}

		now := time.Now()
		rowsAffected, err := s.ExamLOSubmissionRepo.BulkUpdateApproveReject(ctx, tx, &repositories.BulkUpdateApproveRejectArgs{
			SubmissionIDs: database.TextArray(req.SubmissionIds),
			Status:        database.Text(statusChange),
			LastAction:    database.Text(req.ApproveGradingAction.String()),
			LastActionAt:  database.Timestamptz(now),
			LastActionBy:  database.Text(interceptors.UserIDFromContext(ctx)),
			StatusCond:    database.TextArray(statusCond),
			UpdatedAt:     database.Timestamptz(now),
		})
		if err != nil {
			return fmt.Errorf("ExamLOSubmissionRepo.BulkUpdateApproveReject: %w", err)
		}

		if rowsAffected < len(req.SubmissionIds) {
			invalidIDs, err := s.ExamLOSubmissionRepo.GetInvalidIDsByBulkApproveReject(ctx, tx, database.TextArray(req.SubmissionIds), database.TextArray(statusCond))
			if err != nil {
				return fmt.Errorf("ExamLOSubmissionRepo.GetInvalidIDsByBulkApproveReject: %w", err)
			}

			respIDs = database.FromTextArray(invalidIDs)

			switch req.ApproveGradingAction {
			case sspb.ApproveGradingAction_APPROVE_ACTION_APPROVED:
				return errors.New("bulk approved is only allowed which submission has status is Marked")
			case sspb.ApproveGradingAction_APPROVE_ACTION_REJECTED:
				return errors.New("bulk rejected is only allowed which submission has status is Marked or Returned")
			}
		}

		return nil
	}); err != nil {
		return &sspb.BulkApproveRejectSubmissionResponse{
			InvalidSubmissionIds: respIDs,
		}, status.Errorf(codes.Internal, err.Error())
	}

	return &sspb.BulkApproveRejectSubmissionResponse{}, nil
}

func validateRetrieveMetadataTaggingResultRequest(req *sspb.RetrieveMetadataTaggingResultRequest) error {
	if len(req.SubmissionId) == 0 {
		return fmt.Errorf("missing submission id")
	}
	return nil
}

type point struct {
	gradedPoint int32
	totalPoint  int32
}

func (s *ExamLOService) RetrieveMetadataTaggingResult(ctx context.Context, req *sspb.RetrieveMetadataTaggingResultRequest) (*sspb.RetrieveMetadataTaggingResultResponse, error) {
	if err := validateRetrieveMetadataTaggingResultRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validateRetrieveMetadataTaggingResultRequest: %s", err.Error()))
	}

	pointPerTags, err := s.QuestionTagRepo.GetPointPerTagBySubmissionID(ctx, s.DB, database.Text(req.SubmissionId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "s.QuestionTagRepo.GetPointPerTagBySubmissionID: %s", err)
	}

	taggingResults := make([]*sspb.TaggingResult, 0)
	for _, pointPerTag := range pointPerTags {
		taggingResults = append(taggingResults, &sspb.TaggingResult{
			TagId:       pointPerTag.QuestionTagID,
			TagName:     pointPerTag.QuestionTagName,
			GradedPoint: uint32(pointPerTag.GradedPoint),
			TotalPoint:  uint32(pointPerTag.TotalPoint),
		})
	}

	return &sspb.RetrieveMetadataTaggingResultResponse{
		TaggingResults: taggingResults,
	}, nil
}
