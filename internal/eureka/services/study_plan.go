package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type StudyPlanService struct {
	sspb.UnimplementedStudyPlanServer
	JSM nats.JetStreamManagement
	DB  database.Ext

	IndividualStudyPlanItemRepo interface {
		BulkSync(ctx context.Context, db database.QueryExecer, args []*entities.IndividualStudyPlan) ([]*entities.IndividualStudyPlan, error)
	}

	ImportStudyPlanTaskRepo interface {
		Insert(ctx context.Context, db database.QueryExecer, e *entities.ImportStudyPlanTask) error
	}

	StudyPlanItemRepo interface {
		UpdateSchoolDateByStudyPlanItemIdentity(ctx context.Context, db database.QueryExecer, lmID, studyPlanID pgtype.Text, studentIDs pgtype.TextArray, schoolDate pgtype.Timestamptz) error
		BulkUpdateSchoolDate(ctx context.Context, db database.QueryExecer, studyPlanItemIds pgtype.TextArray, schoolDate pgtype.Timestamptz) error
		BulkUpdateStartEndDate(ctx context.Context, db database.QueryExecer, studyPlanItemIds pgtype.TextArray, updateFields sspb.UpdateStudyPlanItemsStartEndDateFields, startDate, endDate pgtype.Timestamptz) (int64, error)
		ListSPItemByIdentity(ctx context.Context, db database.QueryExecer, studyPlanItemIdentities []repositories.StudyPlanItemIdentity) ([]string, error)
		UpdateStudyPlanItemsStatus(ctx context.Context, db database.QueryExecer, studyPlanItemIds pgtype.TextArray, spiStatus pgtype.Text) (int64, error)
	}
	MasterStudyPlanRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.MasterStudyPlan) error
		FindByID(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.Text) ([]*entities.MasterStudyPlan, error)
	}
	StudyPlanRepo interface {
		RetrieveStudyPlanIdentity(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*repositories.RetrieveStudyPlanIdentityResponse, error)
		ListIndividualStudyPlanItems(ctx context.Context, db database.QueryExecer, args *repositories.ListIndividualStudyPlanArgs) ([]*repositories.IndividualStudyPlanItem, error)
		ListStudentToDoItem(ctx context.Context, db database.QueryExecer, args *repositories.ListStudentToDoItemArgs) ([]*repositories.StudentStudyPlanItem, []*repositories.TopicProgress, error)
		ListStudentStudyPlans(ctx context.Context, db database.QueryExecer, args *repositories.ListStudentStudyPlansArgs) ([]*repositories.StudentStudyPlan, error)
	}
	AllocateMarkerRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.AllocateMarker) error
		GetAllocateTeacherByCourseAccess(ctx context.Context, db database.QueryExecer, courseIds pgtype.TextArray) ([]*entities.AllocateTeacherItem, error)
		GetTeacherID(ctx context.Context, db database.QueryExecer, args *repositories.StudyPlanItemIdentity) (pgtype.Text, error)
	}
}

func NewStudyPlanService(db database.Ext, jsm nats.JetStreamManagement) sspb.StudyPlanServer {
	return &StudyPlanService{
		DB:                          db,
		JSM:                         jsm,
		IndividualStudyPlanItemRepo: new(repositories.IndividualStudyPlan),
		StudyPlanItemRepo:           new(repositories.StudyPlanItemRepo),
		MasterStudyPlanRepo:         new(repositories.MasterStudyPlanRepo),
		StudyPlanRepo:               new(repositories.StudyPlanRepo),
		AllocateMarkerRepo:          new(repositories.AllocateMarkerRepo),
		ImportStudyPlanTaskRepo:     new(repositories.ImportStudyPlanTaskRepo),
	}
}

func validateUpsertIndividualRequest(in *sspb.UpsertIndividualInfoRequest) error {
	if in.IndividualItems == nil || len(in.IndividualItems) == 0 {
		return fmt.Errorf("Individual items must not be empty")
	}
	for _, item := range in.GetIndividualItems() {
		if item.StudyPlanItemIdentity.GetLearningMaterialId() == "" {
			return fmt.Errorf("Learning material id must not be empty")
		}

		if item.StudyPlanItemIdentity.GetStudyPlanId() == "" {
			return fmt.Errorf("Study plan id must not be empty")
		}

		if item.StudyPlanItemIdentity.GetStudentId().GetValue() == "" {
			return fmt.Errorf("Student id must not be empty")
		}
	}
	return nil
}

func validateUpsertSchoolDateRequest(in *sspb.UpsertSchoolDateRequest) error {
	if err := validateSPItemIdentities(in.StudyPlanItemIdentities); err != nil {
		return fmt.Errorf("validateSPItemIdentities: %s", err)
	}

	if in.SchoolDate == nil {
		return fmt.Errorf("school date must not be empty")
	}

	return nil
}

func toIndividualStudyPlansEnt(req *sspb.UpsertIndividualInfoRequest) ([]*entities.IndividualStudyPlan, error) {
	ispEntities := []*entities.IndividualStudyPlan{}
	for _, item := range req.GetIndividualItems() {
		e := &entities.IndividualStudyPlan{}
		database.AllNullEntity(e)
		e.Now()
		if err := multierr.Combine(
			e.ID.Set(item.StudyPlanItemIdentity.GetStudyPlanId()),
			e.LearningMaterialID.Set(item.StudyPlanItemIdentity.GetLearningMaterialId()),
			e.StudentID.Set(item.StudyPlanItemIdentity.GetStudentId().GetValue()),
			e.Status.Set(item.GetStatus()),
		); err != nil {
			return nil, err
		}

		if item.AvailableFrom != nil {
			e.AvailableFrom.Set(item.AvailableFrom.AsTime())
		}
		if item.AvailableTo != nil {
			e.AvailableTo.Set(item.AvailableTo.AsTime())
		}
		if item.StartDate != nil {
			e.StartDate.Set(item.StartDate.AsTime())
		}
		if item.EndDate != nil {
			e.EndDate.Set(item.EndDate.AsTime())
		}
		if item.SchoolDate != nil {
			e.SchoolDate.Set(item.GetSchoolDate().AsTime())
		}

		ispEntities = append(ispEntities, e)
	}

	return ispEntities, nil
}

func (s *StudyPlanService) validateImportStudyPlanRequest(ctx context.Context, req *sspb.ImportStudyPlanRequest) (map[int32]string, error) {
	if req.GetStudyPlanItems() == nil || len(req.GetStudyPlanItems()) == 0 {
		return nil, fmt.Errorf("empty study plan items")
	}

	mapRowErr := map[int32]string{}
	lmIds := []string{}
	studyPlans, err := s.MasterStudyPlanRepo.FindByID(ctx, s.DB, database.Text(req.GetStudyPlanItems()[0].StudyPlanId))
	if studyPlans == nil || err != nil {
		return nil, fmt.Errorf("invalid study plan id")
	}

	for _, i := range studyPlans {
		lmIds = append(lmIds, i.LearningMaterialID.String)
	}

	for i, row := range req.GetStudyPlanItems() {
		// validate each row

		if req.GetStudyPlanItems()[0].StudyPlanId != row.StudyPlanId {
			mapRowErr[int32(i)] = "1"
		}

		if !slices.Contains(lmIds, row.LearningMaterialId) {
			mapRowErr[int32(i)] = "2"
		}

		if row.AvailableFrom != nil && row.AvailableTo != nil && row.AvailableTo.AsTime().Before(row.AvailableFrom.AsTime()) {
			mapRowErr[int32(i)] = "5"
		}

		if row.StartDate != nil && (row.StartDate.AsTime().Before(row.AvailableFrom.AsTime()) || row.AvailableTo.AsTime().Before(row.StartDate.AsTime())) {
			mapRowErr[int32(i)] = "7"
		}

		if row.EndDate != nil && (row.EndDate.AsTime().Before(row.AvailableFrom.AsTime()) || row.AvailableTo.AsTime().Before(row.EndDate.AsTime())) {
			mapRowErr[int32(i)] = "9"
		}

		if row.EndDate != nil && row.StartDate != nil && row.EndDate.AsTime().Before(row.StartDate.AsTime()) {
			mapRowErr[int32(i)] = "10"
		}
	}

	//nolint
	if len(mapRowErr) == 0 {
		return nil, nil
	}
	return mapRowErr, fmt.Errorf("invalid request")
}

func (s *StudyPlanService) ImportStudyPlan(ctx context.Context, req *sspb.ImportStudyPlanRequest) (*sspb.ImportStudyPlanResponse, error) {
	// validation
	if mapRowErr, err := s.validateImportStudyPlanRequest(ctx, req); err != nil {
		rowErrs := []*sspb.RowError{}
		for i, v := range mapRowErr {
			rowErrs = append(rowErrs, &sspb.RowError{
				RowNumber: i,
				Err:       v,
			})
		}

		return &sspb.ImportStudyPlanResponse{
			RowErrors: rowErrs,
		}, nil
	}

	// create task
	_, userID, _ := interceptors.GetUserInfoFromContext(ctx)
	taskID := idutil.ULIDNow()
	now := time.Now()
	e := &entities.ImportStudyPlanTask{}
	database.AllNullEntity(e)

	if err := multierr.Combine(
		e.TaskID.Set(taskID),
		e.StudyPlanID.Set(req.GetStudyPlanItems()[0].StudyPlanId),
		e.Status.Set(epb.StudyPlanTaskStatus_STUDY_PLAN_TASK_STATUS_NONE.String()),
		e.ImportedBy.Set(userID),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("multiErr.Combine: %w", err).Error())
	}

	// transaction create task and publish event
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// create task
		if err := s.ImportStudyPlanTaskRepo.Insert(ctx, tx, e); err != nil {
			return fmt.Errorf("s.ImportStudyPlanTaskRepo.Insert: %w", err)
		}

		// publish event
		msg, err := proto.Marshal(&npb.EventImportStudyPlan{
			TaskId:         taskID,
			StudyPlanItems: req.StudyPlanItems,
		})

		if err != nil {
			return fmt.Errorf("EventImportStudyPlan: proto.Marshal: %v", err)
		}

		if _, err = s.JSM.PublishContext(ctx, constants.SubjectStudyPlanItemsImported, msg); err != nil {
			return fmt.Errorf("s.JSM.PublishContext: subject: %q, %v", constants.SubjectStudyPlanItemsImported, err)
		}

		return nil
	}); err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}

	return &sspb.ImportStudyPlanResponse{
		TaskId: taskID,
	}, nil
}

func (s *StudyPlanService) UpsertIndividual(ctx context.Context, in *sspb.UpsertIndividualInfoRequest) (*sspb.UpsertIndividualInfoResponse, error) {
	if err := validateUpsertIndividualRequest(in); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("validateUpsertIndividualRequest: %w", err).Error())
	}

	ispEntities, err := toIndividualStudyPlansEnt(in)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("toIndividualStudyPlansEnt: %w", err).Error())
	}
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		_, err = s.IndividualStudyPlanItemRepo.BulkSync(ctx, tx, ispEntities)
		if err != nil {
			return status.Errorf(codes.Internal, fmt.Errorf("s.StudyPlanItemRepo.: %w", err).Error())
		}

		return nil
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}

	return &sspb.UpsertIndividualInfoResponse{}, nil
}

func (s *StudyPlanService) UpsertSchoolDate(ctx context.Context, req *sspb.UpsertSchoolDateRequest) (*sspb.UpsertSchoolDateResponse, error) {
	if err := validateUpsertSchoolDateRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("validateUpsertSchoolDateRequest: %w", err).Error())
	}
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		studyPlanItemIDs, err := s.StudyPlanItemRepo.ListSPItemByIdentity(ctx, tx,
			convertSPItemIdentitiesFromPbToRepo(req.StudyPlanItemIdentities))
		if err != nil {
			return fmt.Errorf("StudyPlanItemRepo.ListSPItemByIdentity: %s", err)
		}
		err = s.StudyPlanItemRepo.BulkUpdateSchoolDate(ctx, tx,
			database.TextArray(studyPlanItemIDs), database.TimestamptzFromPb(req.SchoolDate))
		if err != nil {
			return fmt.Errorf("StudyPlanItemRepo.BulkUpdateSchoolDate: %s", err.Error())
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "database.ExecInTx: %s", err.Error())
	}

	return &sspb.UpsertSchoolDateResponse{}, nil
}

func (s *StudyPlanService) UpsertMasterInfo(ctx context.Context, req *sspb.UpsertMasterInfoRequest) (*sspb.UpsertMasterInfoResponse, error) {
	if err := validateUpsertMasterInfoRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("validateUpsertMasterInfoRequest: %w", err).Error())
	}

	entityMasterStudyPlans := make([]*entities.MasterStudyPlan, 0)

	for _, item := range req.MasterItems {
		entityMasterStudyPlan, err := toMasterStudyPlanEntity(item)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("toMasterStudyPlanEntity: %w", err).Error())
		}
		entityMasterStudyPlans = append(entityMasterStudyPlans, entityMasterStudyPlan)
	}
	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := s.MasterStudyPlanRepo.BulkUpsert(ctx, tx, entityMasterStudyPlans)
		if err != nil {
			return fmt.Errorf("s.MasterStudyPlanRepo.BulkUpsert: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}
	return &sspb.UpsertMasterInfoResponse{}, nil
}

func (s *StudyPlanService) UpsertAllocateMarker(ctx context.Context, req *sspb.UpsertAllocateMarkerRequest) (*sspb.UpsertAllocateMarkerResponse, error) {
	if err := validateUpsertAllocateMarkerRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("validateUpsertAllocateMarkerRequest: %w", err).Error())
	}

	allocateMarkerEntities := make([]*entities.AllocateMarker, 0)
	submissionIdx := 0
	for _, markerItem := range req.GetAllocateMarkers() {
		allocateMarkerEntities = append(allocateMarkerEntities, toAllocateMarkerEntities(req.Submissions[submissionIdx:(submissionIdx+int(markerItem.GetNumberAllocated()))], markerItem.TeacherId, req.GetCreatedBy())...)
		submissionIdx += int(markerItem.GetNumberAllocated())
	}

	if err := s.AllocateMarkerRepo.BulkUpsert(ctx, s.DB, allocateMarkerEntities); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.AllocateMarkerRepo.BulkUpsert: %w", err).Error())
	}

	return &sspb.UpsertAllocateMarkerResponse{}, nil
}

func (s *StudyPlanService) ListAllocateTeacher(ctx context.Context, req *sspb.ListAllocateTeacherRequest) (*sspb.ListAllocateTeacherResponse, error) {
	locationIDs := pgtype.TextArray{Status: pgtype.Null}
	if req.GetLocationIds() != nil && len(req.GetLocationIds()) > 0 {
		if err := locationIDs.Set(req.GetLocationIds()); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("locationIDs.Set: %w", err).Error())
		}
	}

	allocateTeachers, err := s.AllocateMarkerRepo.GetAllocateTeacherByCourseAccess(ctx, s.DB, locationIDs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.AllocateMarkerRepo.GetAllocateTeacherByCourseAccess: %w", err).Error())
	}

	return &sspb.ListAllocateTeacherResponse{
		AllocateTeachers: convertListAllocateTeachers(allocateTeachers),
	}, nil
}

func validateUpsertMasterInfoRequest(req *sspb.UpsertMasterInfoRequest) error {
	if len(req.MasterItems) == 0 {
		return fmt.Errorf("MasterStudyPlans is empty")
	}
	for _, item := range req.MasterItems {
		studyPlanId := item.MasterStudyPlanIdentify.StudyPlanId
		learningMaterialId := item.MasterStudyPlanIdentify.LearningMaterialId
		if len(studyPlanId) == 0 {
			return fmt.Errorf("StudyPlanId is empty")
		}
		if len(learningMaterialId) == 0 {
			return fmt.Errorf("LearningMaterialId is empty")
		}
		if item.EndDate != nil && item.StartDate != nil && item.EndDate.AsTime().Before(item.StartDate.AsTime()) {
			return fmt.Errorf("StudyPlanId: %s, LearningMaterialId: %s, end_date before start_date", studyPlanId, learningMaterialId)
		}
		if item.AvailableFrom != nil && item.AvailableTo != nil && item.AvailableTo.AsTime().Before(item.AvailableFrom.AsTime()) {
			return fmt.Errorf("StudyPlanId: %s, LearningMaterialId: %s, available_to before available_from", studyPlanId, learningMaterialId)
		}
	}
	return nil
}

func validateUpsertAllocateMarkerRequest(req *sspb.UpsertAllocateMarkerRequest) error {
	for _, submission := range req.GetSubmissions() {
		if submission.StudyPlanItemIdentity == nil {
			return fmt.Errorf("submission must be not empty")
		}
	}

	totalSubmission := 0
	for _, markerItem := range req.GetAllocateMarkers() {
		if int(markerItem.GetNumberAllocated()) <= 0 {
			return fmt.Errorf("number of allocated submission must be not less or equal than zero")
		}

		totalSubmission += int(markerItem.NumberAllocated)
	}

	if req.Submissions == nil || len(req.GetSubmissions()) != totalSubmission {
		return fmt.Errorf("total allocated submission does not equal the total of submission selected")
	}

	return nil
}

func validateListAllocateTeacherRequest(req *sspb.ListAllocateTeacherRequest) error {
	if req.GetLocationIds() == nil || len(req.GetLocationIds()) == 0 {
		return fmt.Errorf("location ids must be not empty")
	}

	return nil
}

func toMasterStudyPlanEntity(src *sspb.MasterStudyPlan) (*entities.MasterStudyPlan, error) {
	dst := &entities.MasterStudyPlan{}
	database.AllNullEntity(dst)
	dst.Now()
	if src.StartDate != nil {
		dst.StartDate = database.TimestamptzFromPb(src.StartDate)
	}
	if src.EndDate != nil {
		dst.EndDate = database.TimestamptzFromPb(src.EndDate)
	}
	if src.AvailableFrom != nil {
		dst.AvailableFrom = database.TimestamptzFromPb(src.AvailableFrom)
	}
	if src.AvailableTo != nil {
		dst.AvailableTo = database.TimestamptzFromPb(src.AvailableTo)
	}
	if src.SchoolDate != nil {
		dst.SchoolDate = database.TimestamptzFromPb(src.SchoolDate)
	}

	setErr := multierr.Combine(
		dst.StudyPlanID.Set(src.MasterStudyPlanIdentify.StudyPlanId),
		dst.LearningMaterialID.Set(src.MasterStudyPlanIdentify.LearningMaterialId),
		dst.Status.Set(src.GetStatus()),
	)
	return dst, setErr
}

func validateUpdateStudyPlanItemsStartEndDateRequest(req *sspb.UpdateStudyPlanItemsStartEndDateRequest) error {
	if err := validateSPItemIdentities(req.StudyPlanItemIdentities); err != nil {
		return fmt.Errorf("validateSPItemIdentities: %s", err)
	}

	if req.Fields != sspb.UpdateStudyPlanItemsStartEndDateFields_ALL &&
		req.Fields != sspb.UpdateStudyPlanItemsStartEndDateFields_START_DATE &&
		req.Fields != sspb.UpdateStudyPlanItemsStartEndDateFields_END_DATE {
		return fmt.Errorf("invalid fields need to update")
	}
	switch req.Fields {
	case sspb.UpdateStudyPlanItemsStartEndDateFields_START_DATE:
		// we don't accept user update start date to null
		if req.StartDate == nil {
			return fmt.Errorf("startdate have to not null")
		}
	case sspb.UpdateStudyPlanItemsStartEndDateFields_END_DATE:
		// we don't accept user update end date to null
		if req.EndDate == nil {
			return fmt.Errorf("enddate have to not null")
		}
	case sspb.UpdateStudyPlanItemsStartEndDateFields_ALL:
		// we don't accept user update start date to null
		if req.StartDate == nil || req.EndDate == nil {
			// we don't accept user update start date && end date to null
			return fmt.Errorf("startdate and enddate have to not null")
		}
		if req.StartDate.AsTime().After(req.EndDate.AsTime()) {
			return fmt.Errorf("startdate after enddate")
		}
	}
	return nil
}

func validateSPItemIdentities(identities []*sspb.StudyPlanItemIdentity) error {
	for _, identity := range identities {
		if len(identity.LearningMaterialId) == 0 {
			return fmt.Errorf("learning material id must not be empty")
		}
		if len(identity.StudyPlanId) == 0 {
			return fmt.Errorf("study plan id must not be empty")
		}
		if identity.StudentId != nil && len(identity.StudentId.Value) == 0 {
			return fmt.Errorf("student id must be nil or have value")
		}
	}
	return nil
}

func convertSPItemIdentitiesFromPbToRepo(pbIdentities []*sspb.StudyPlanItemIdentity) []repositories.StudyPlanItemIdentity {
	repoIdentities := make([]repositories.StudyPlanItemIdentity, 0)
	for _, pbIdentity := range pbIdentities {
		var studentID pgtype.Text
		if pbIdentity.StudentId != nil {
			studentID = pgtype.Text{
				String: pbIdentity.StudentId.Value,
				Status: pgtype.Present,
			}
		} else {
			studentID = pgtype.Text{
				String: "",
				Status: pgtype.Null,
			}
		}
		repoIdentities = append(repoIdentities, repositories.StudyPlanItemIdentity{
			StudentID:          studentID,
			StudyPlanID:        database.Text(pbIdentity.StudyPlanId),
			LearningMaterialID: database.Text(pbIdentity.LearningMaterialId),
		})
	}
	return repoIdentities
}

func (s *StudyPlanService) UpdateStudyPlanItemsStartEndDate(ctx context.Context, req *sspb.UpdateStudyPlanItemsStartEndDateRequest) (*sspb.UpdateStudyPlanItemsStartEndDateResponse, error) {
	if err := validateUpdateStudyPlanItemsStartEndDateRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validateUpdateStudyPlanItemsStartEndDateRequest: %s", err.Error())
	}
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		studyPlanItemIDs, err := s.StudyPlanItemRepo.ListSPItemByIdentity(ctx, tx,
			convertSPItemIdentitiesFromPbToRepo(req.StudyPlanItemIdentities))
		if err != nil {
			return fmt.Errorf("StudyPlanItemRepo.ListSPItemByIdentity: %s", err)
		}
		updateRows, err := s.StudyPlanItemRepo.BulkUpdateStartEndDate(ctx, tx,
			database.TextArray(studyPlanItemIDs), req.Fields,
			database.TimestamptzFromPb(req.StartDate),
			database.TimestamptzFromPb(req.EndDate))
		if err != nil {
			return fmt.Errorf("StudyPlanItemRepo.BulkUpdateStartEndDate: %s", err.Error())
		}

		if int64(len(studyPlanItemIDs)) != updateRows {
			return fmt.Errorf("total updated rows should be equal to total students")
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "database.ExecInTx: %s", err.Error())
	}

	return &sspb.UpdateStudyPlanItemsStartEndDateResponse{}, nil
}

func (s *StudyPlanService) RetrieveStudyPlanIdentity(ctx context.Context, req *sspb.RetrieveStudyPlanIdentityRequest) (*sspb.RetrieveStudyPlanIdentityResponse, error) {
	if req.StudyPlanItemIds == nil || len(req.StudyPlanItemIds) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "missing study plan item ids")
	}

	studyPlanIdentities, err := s.StudyPlanRepo.RetrieveStudyPlanIdentity(ctx, s.DB, database.TextArray(req.StudyPlanItemIds))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "StudyPlanRepo.RetrieveStudyPlanIdentity: %s", err.Error())
	}

	res := &sspb.RetrieveStudyPlanIdentityResponse{
		StudyPlanIdentities: make([]*sspb.StudyPlanIdentity, 0, len(studyPlanIdentities)),
	}

	for _, studyPlanIdentity := range studyPlanIdentities {
		res.StudyPlanIdentities = append(res.StudyPlanIdentities, &sspb.StudyPlanIdentity{
			StudyPlanId:        studyPlanIdentity.StudyPlanID.String,
			StudentId:          studyPlanIdentity.StudentID.String,
			LearningMaterialId: studyPlanIdentity.LearningMaterialID.String,
			StudyPlanItemId:    studyPlanIdentity.StudyPlanItemID.String,
		})
	}

	return res, nil
}

func validateUpdateBulkUpdateStudyPlanItemStatusRequest(req *sspb.BulkUpdateStudyPlanItemStatusRequest) error {
	if err := validateSPItemIdentities(req.StudyPlanItemIdentities); err != nil {
		return fmt.Errorf("validateSPItemIdentities: %s", err)
	}

	if _, ok := sspb.StudyPlanItemStatus_value[req.GetStudyPlanItemStatus().String()]; !ok {
		return fmt.Errorf("invalid status need to update")
	}
	return nil
}

func (s *StudyPlanService) BulkUpdateStudyPlanItemStatus(ctx context.Context, req *sspb.BulkUpdateStudyPlanItemStatusRequest) (*sspb.BulkUpdateStudyPlanItemStatusResponse, error) {
	if err := validateUpdateBulkUpdateStudyPlanItemStatusRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validateUpdateBulkUpdateStudyPlanItemStatusRequest: %s", err.Error())
	}
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		studyPlanItemIDs, err := s.StudyPlanItemRepo.ListSPItemByIdentity(ctx, tx,
			convertSPItemIdentitiesFromPbToRepo(req.StudyPlanItemIdentities))
		if err != nil {
			return fmt.Errorf("StudyPlanItemRepo.ListSPItemByIdentity: %s", err)
		}
		updateRows, err := s.StudyPlanItemRepo.UpdateStudyPlanItemsStatus(ctx, tx,
			database.TextArray(studyPlanItemIDs), database.Text(req.GetStudyPlanItemStatus().String()))
		if err != nil {
			return fmt.Errorf("StudyPlanItemRepo.UpdateStudyPlanItemsStatus: %s", err.Error())
		}

		if int64(len(studyPlanItemIDs)) != updateRows {
			return fmt.Errorf("total updated rows should be equal to total students")
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "database.ExecInTx: %s", err.Error())
	}

	return &sspb.BulkUpdateStudyPlanItemStatusResponse{}, nil
}

func toAllocateMarkerEntities(items []*sspb.UpsertAllocateMarkerRequest_SubmissionItem, teacherID, createdBy string) []*entities.AllocateMarker {
	allocateMarkers := make([]*entities.AllocateMarker, 0)
	now := time.Now()

	for _, item := range items {
		allocateMarkers = append(allocateMarkers, &entities.AllocateMarker{
			AllocateMarkerID:   database.Text(idutil.ULIDNow()),
			TeacherID:          database.Text(teacherID),
			CreatedBy:          database.Text(createdBy),
			StudentID:          database.Text(item.GetStudyPlanItemIdentity().GetStudentId().GetValue()),
			StudyPlanID:        database.Text(item.GetStudyPlanItemIdentity().GetStudyPlanId()),
			LearningMaterialID: database.Text(item.GetStudyPlanItemIdentity().GetLearningMaterialId()),
			BaseEntity: entities.BaseEntity{
				CreatedAt: database.Timestamptz(now),
				UpdatedAt: database.Timestamptz(now),
				DeletedAt: pgtype.Timestamptz{
					Status: pgtype.Null,
				},
			},
		})
	}

	return allocateMarkers
}

func convertListAllocateTeachers(allocateTeachers []*entities.AllocateTeacherItem) []*sspb.ListAllocateTeacherResponse_AllocateTeacherItem {
	results := []*sspb.ListAllocateTeacherResponse_AllocateTeacherItem{}
	for _, item := range allocateTeachers {
		results = append(results, &sspb.ListAllocateTeacherResponse_AllocateTeacherItem{
			TeacherId:                item.TeacherID.String,
			TeacherName:              item.TeacherName.String,
			NumberAssignedSubmission: item.NumberAssignedSubmission,
		})
	}

	return results
}

func convertListToDoItemsRequestToArgs(req *sspb.ListToDoItemRequest) (*repositories.ListIndividualStudyPlanArgs, error) {
	now := timeutil.Now().UTC()
	args := &repositories.ListIndividualStudyPlanArgs{
		StudentID:          database.Text(req.StudentId),
		Limit:              10,
		LearningMaterialID: pgtype.Text{Status: pgtype.Null},
		CourseIDs:          database.TextArray(req.CourseIds),
	}

	switch req.Status {
	case sspb.StudyPlanItemToDoStatus_STUDY_PLAN_ITEM_TO_DO_STATUS_ACTIVE:
		args.Status = "active"
		args.Offset.Set(nil)
	case sspb.StudyPlanItemToDoStatus_STUDY_PLAN_ITEM_TO_DO_STATUS_COMPLETED:
		args.Status = "completed"
		args.Offset.Set(nil)
	case sspb.StudyPlanItemToDoStatus_STUDY_PLAN_ITEM_TO_DO_STATUS_OVERDUE:
		args.Status = "overdue"
		// offset is start_date, and start_date must be exist in 1st page.
		args.Offset.Set(now)
	default:
		err := fmt.Errorf("unknown todo status: %v", req.Status)
		return nil, err
	}
	return args, nil
}

func (s *StudyPlanService) listIndividualStudyPlanItem(ctx context.Context, args *repositories.ListIndividualStudyPlanArgs, paging *cpb.Paging) ([]*repositories.IndividualStudyPlanItem, *cpb.Paging, error) {
	args.LearningMaterialID.Set(nil)

	if paging != nil {
		if limit := paging.Limit; 1 <= limit && limit <= 100 {
			args.Limit = limit
		}
		if c := paging.GetOffsetCombined(); c != nil {
			if c.OffsetString != "" {
				args.LearningMaterialID.Set(c.OffsetString)
			}

			if c.OffsetTime != nil && c.OffsetTime.AsTime().Unix() > 0 {
				args.Offset.Set(c.OffsetTime.AsTime())
			} else {
				args.Offset.Set(nil)
			}
		}
	}
	var (
		items []*repositories.IndividualStudyPlanItem
		err   error
	)
	items, err = s.StudyPlanRepo.ListIndividualStudyPlanItems(ctx, s.DB, args)
	if err != nil {
		return nil, nil, fmt.Errorf("StudyPlanRepo.List%sItems: %w", args.Status, err)
	}

	nextPage := &cpb.Paging{
		Limit: args.Limit,
	}
	if len(items) > 0 {
		lastItem := items[len(items)-1]
		nextPage.Offset = &cpb.Paging_OffsetCombined{
			OffsetCombined: &cpb.Paging_Combined{
				OffsetTime:   timestamppb.New(lastItem.StartDate.Time),
				OffsetString: lastItem.LearningMaterialID.String,
			},
		}
	}

	return items, nextPage, nil
}

func (s *StudyPlanService) toDoItems(items []*repositories.IndividualStudyPlanItem, status sspb.StudyPlanItemToDoStatus) []*sspb.StudyPlanToDoItem {
	toDoItems := make([]*sspb.StudyPlanToDoItem, 0, len(items))
	for _, item := range items {
		individualStudyPlanItem := &sspb.StudyPlanItem{
			StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
				StudyPlanId:        item.StudyPlanID.String,
				LearningMaterialId: item.LearningMaterialID.String,
				StudentId: &wrapperspb.StringValue{
					Value: item.StudentID.String,
				},
			},
			AvailableFrom: timestamppb.New(item.AvailableFrom.Time),
			AvailableTo:   timestamppb.New(item.AvailableTo.Time),
			StartDate:     timestamppb.New(item.StartDate.Time),
			EndDate:       timestamppb.New(item.EndDate.Time),
			Status:        sspb.StudyPlanItemStatus(sspb.StudyPlanItemStatus_value[item.Status.String]),
			SchoolDate:    timestamppb.New(item.SchoolDate.Time),
		}

		if item.CompletedAt.Status == pgtype.Present {
			individualStudyPlanItem.CompletedAt = timestamppb.New(item.CompletedAt.Time)
		}

		toDoItems = append(toDoItems, &sspb.StudyPlanToDoItem{
			IndividualStudyPlanItem: individualStudyPlanItem,
			Status:                  status,
			Crown:                   getAchievementCrownForToDoItem(float32(item.Score.Int)),
			LearningMaterialType:    getLearningMaterialType(item.Type.String),
		})
	}
	return toDoItems
}

func getAchievementCrownForToDoItem(score float32) sspb.AchievementCrown {
	switch {
	case score == 100:
		return sspb.AchievementCrown_ACHIEVEMENT_CROWN_GOLD
	case score >= 80:
		return sspb.AchievementCrown_ACHIEVEMENT_CROWN_SILVER
	case score >= 60:
		return sspb.AchievementCrown_ACHIEVEMENT_CROWN_BRONZE
	default:
		return sspb.AchievementCrown_ACHIEVEMENT_CROWN_NONE
	}
}

func (s *StudyPlanService) ListToDoItem(ctx context.Context, req *sspb.ListToDoItemRequest) (*sspb.ListToDoItemResponse, error) {
	if req.StudentId == "" {
		return nil, status.Error(codes.InvalidArgument, "student id is required")
	}
	if len(req.CourseIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "a list of course ids is required")
	}
	args, err := convertListToDoItemsRequestToArgs(req)
	if err != nil {
		return nil, err
	}

	items, nextPage, err := s.listIndividualStudyPlanItem(ctx, args, req.Page)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return &sspb.ListToDoItemResponse{}, nil
	}

	return &sspb.ListToDoItemResponse{
		TodoItems: s.toDoItems(items, req.Status),
		NextPage:  nextPage,
	}, nil
}

func convertListToDoItemStructuredBookTreeRequestToArgs(req *sspb.ListToDoItemStructuredBookTreeRequest) *repositories.ListStudentToDoItemArgs {
	args := &repositories.ListStudentToDoItemArgs{
		StudentID:   database.Text(req.StudyPlanIdentity.StudentId.Value),
		StudyPlanID: database.Text(req.StudyPlanIdentity.StudyPlanId),
		Limit:       10,
		TopicID:     pgtype.Text{Status: pgtype.Null},
	}
	return args
}

func (s *StudyPlanService) ListItemStructuredBookTree(ctx context.Context, args *repositories.ListStudentToDoItemArgs, paging *cpb.Paging) ([]*repositories.StudentStudyPlanItem, []*repositories.TopicProgress, *cpb.Paging, error) {
	if paging != nil {
		if limit := paging.Limit; 1 <= limit && limit <= 100 {
			args.Limit = limit
		}
		if c := paging.GetOffsetCombined(); c != nil {
			if c.OffsetString != "" {
				if err := args.TopicID.Set(c.OffsetString); err != nil {
					return nil, nil, nil, fmt.Errorf("TopicID.Set(): %w", err)
				}
			}
		}
	}

	items, topicProgress, err := s.StudyPlanRepo.ListStudentToDoItem(ctx, s.DB, args)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("StudyPlanRepo.List: %w", err)
	}

	nextPage := &cpb.Paging{
		Limit: args.Limit,
	}
	if len(items) > 0 {
		lastItem := items[len(items)-1]
		nextPage.Offset = &cpb.Paging_OffsetCombined{
			OffsetCombined: &cpb.Paging_Combined{
				OffsetString: lastItem.TopicID.String,
			},
		}
	}
	return items, topicProgress, nextPage, nil
}

func (s *StudyPlanService) toStudentStudyPlanItemAndStudentTopicStudyProgress(items []*repositories.StudentStudyPlanItem, studentTopicProgress []*repositories.TopicProgress) (toDoItems []*sspb.StudentStudyPlanItem, topicProgress []*sspb.StudentTopicStudyProgress) {
	for _, item := range items {
		toDoItems = append(toDoItems, &sspb.StudentStudyPlanItem{
			LearningMaterial: &sspb.LearningMaterialBase{
				LearningMaterialId: item.LearningMaterialID.String,
				TopicId:            item.TopicID.String,
				Name:               item.Name.String,
				Type:               item.Type.String,
				DisplayOrder:       &wrapperspb.Int32Value{Value: int32(item.DisplayOrder.Int)},
			},
			StartDate:           timestamppb.New(item.StartDate.Time),
			EndDate:             timestamppb.New(item.EndDate.Time),
			CompletedAt:         timestamppb.New(item.CompletedAt.Time),
			SchoolDate:          timestamppb.New(item.SchoolDate.Time),
			StudyPlanItemStatus: sspb.StudyPlanItemStatus(sspb.StudyPlanItemStatus_value[item.StudyPlanItemStatus.String]),
			BookId:              item.BookID.String,
			AvailableFrom:       Timestamptz2TimestampPb(item.AvailableFrom),
			AvailableTo:         Timestamptz2TimestampPb(item.AvailableTo),
		})
	}
	for _, studentTopic := range studentTopicProgress {
		topicProgress = append(topicProgress, &sspb.StudentTopicStudyProgress{
			TopicId: studentTopic.TopicID.String,
			CompletedStudyPlanItem: &wrapperspb.Int32Value{
				Value: int32(studentTopic.CompletedSPItem.Int),
			},
			TotalStudyPlanItem: &wrapperspb.Int32Value{
				Value: int32(studentTopic.TotalSPItem.Int),
			},
			AverageScore: &wrapperspb.Int32Value{
				Value: int32(studentTopic.AverageScore.Int),
			},
			TopicName:    studentTopic.Name.String,
			IconUrl:      studentTopic.IconURL.String,
			DisplayOrder: int32(studentTopic.DisplayOrder.Int),
		})
	}
	return
}

func Timestamptz2TimestampPb(in pgtype.Timestamptz) *timestamppb.Timestamp {
	if in.Status != pgtype.Present {
		return nil
	}
	return timestamppb.New(in.Time)
}

func validateListToDoItemStructuredBookTreeReq(req *sspb.ListToDoItemStructuredBookTreeRequest) error {
	if req.StudyPlanIdentity.StudyPlanId == "" {
		return status.Error(codes.InvalidArgument, "study plan id is required")
	}

	if req.StudyPlanIdentity.StudentId == nil {
		return status.Error(codes.InvalidArgument, "student id is required")
	}
	return nil
}

func (s *StudyPlanService) ListToDoItemStructuredBookTree(ctx context.Context, req *sspb.ListToDoItemStructuredBookTreeRequest) (*sspb.ListToDoItemStructuredBookTreeResponse, error) {
	if err := validateListToDoItemStructuredBookTreeReq(req); err != nil {
		return nil, err
	}

	args := convertListToDoItemStructuredBookTreeRequestToArgs(req)

	items, topics, nextPage, err := s.ListItemStructuredBookTree(ctx, args, req.Page)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return &sspb.ListToDoItemStructuredBookTreeResponse{}, nil
	}
	toDoItems, topicProgress := s.toStudentStudyPlanItemAndStudentTopicStudyProgress(items, topics)

	return &sspb.ListToDoItemStructuredBookTreeResponse{
		TodoItems:       toDoItems,
		TopicProgresses: topicProgress,
		NextPage:        nextPage,
	}, nil
}

func validateListStudentStudyPlan(req *sspb.ListStudentStudyPlansRequest) error {
	if len(req.GetStudentIds()) == 0 {
		return status.Error(codes.InvalidArgument, "student id is required")
	}
	return nil
}

func (s *StudyPlanService) ListStudentStudyPlan(ctx context.Context, req *sspb.ListStudentStudyPlansRequest) (*sspb.ListStudentStudyPlansResponse, error) {
	if err := validateListStudentStudyPlan(req); err != nil {
		return nil, err
	}
	query := &repositories.ListStudentStudyPlansArgs{
		StudentIDs: database.TextArray(req.GetStudentIds()),
		CourseID:   pgtype.Text{Status: pgtype.Null},
		Limit:      10,
		Offset:     pgtype.Text{Status: pgtype.Null},
		Search:     pgtype.Text{Status: pgtype.Null},
		Status:     pgtype.Text{Status: pgtype.Null},
		BookIDs:    pgtype.TextArray{Status: pgtype.Null},
		Grades:     pgtype.Int4Array{Status: pgtype.Null},
	}
	if req.CourseId != "" {
		query.CourseID.Set(req.CourseId)
	}
	if req.Paging != nil {
		if limit := req.Paging.Limit; 1 <= limit && limit <= 100 {
			query.Limit = limit
		}
		if o := req.Paging.GetOffsetString(); o != "" {
			query.Offset = database.Text(o)
		}
	}
	if req.Search != "" {
		_ = query.Search.Set(req.Search)
	}
	if req.Status != "" {
		_ = query.Status.Set(req.Status)
	}
	if bookIDs := req.GetBookIds(); len(bookIDs) != 0 {
		_ = query.BookIDs.Set(bookIDs)
	}
	if len(req.Grades) != 0 {
		_ = query.Grades.Set(req.Grades)
	}

	plans, err := s.StudyPlanRepo.ListStudentStudyPlans(ctx, s.DB, query)
	if err != nil {
		return nil, err
	}
	if len(plans) == 0 {
		return &sspb.ListStudentStudyPlansResponse{}, nil
	}

	plansPb := make([]*sspb.StudentStudyPlanData, 0, len(plans))
	for _, plan := range plans {
		plansPb = append(plansPb, &sspb.StudentStudyPlanData{
			StudyPlanId:         plan.ID.String,
			Name:                plan.Name.String,
			BookId:              plan.BookID.String,
			Status:              sspb.StudyPlanStatus(sspb.StudyPlanStatus_value[plan.Status.String]),
			Grades:              database.FromInt4Array(plan.Grades),
			StudentId:           plan.StudentID.String,
			TrackSchoolProgress: plan.TrackSchoolProgress.Bool,
		})
	}

	return &sspb.ListStudentStudyPlansResponse{
		StudyPlans: plansPb,
		NextPage: &cpb.Paging{
			Limit: query.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: plans[len(plans)-1].ID.String,
			},
		},
	}, nil
}

func (s *StudyPlanService) RetrieveAllocateMarker(ctx context.Context, req *sspb.RetrieveAllocateMarkerRequest) (*sspb.RetrieveAllocateMarkerResponse, error) {
	if err := validateRetrieveAllocateMarkerRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validateRetrieveAllocateMarkerRequest: %s", err.Error())
	}

	studyPlanItemIdentityArg := &repositories.StudyPlanItemIdentity{
		StudentID:          database.Text(req.StudyPlanItemIdentity.StudentId.GetValue()),
		StudyPlanID:        database.Text(req.StudyPlanItemIdentity.StudyPlanId),
		LearningMaterialID: database.Text(req.StudyPlanItemIdentity.LearningMaterialId),
	}
	teacherID, err := s.AllocateMarkerRepo.GetTeacherID(ctx, s.DB, studyPlanItemIdentityArg)
	if err != nil && err != pgx.ErrNoRows {
		return nil, status.Errorf(codes.Internal, "s.AllocateMarkerRepo.GetTeacherID: %s", err.Error())
	}

	var teacherStr string
	if teacherID.Status == pgtype.Present {
		teacherStr = teacherID.String
	}

	return &sspb.RetrieveAllocateMarkerResponse{
		MarkerId: teacherStr,
	}, nil
}

func validateRetrieveAllocateMarkerRequest(req *sspb.RetrieveAllocateMarkerRequest) error {
	if req.StudyPlanItemIdentity.LearningMaterialId == "" {
		return fmt.Errorf("learning_material_id must not be empty")
	}

	if req.StudyPlanItemIdentity.StudentId == nil && req.StudyPlanItemIdentity.StudentId.GetValue() == "" {
		return fmt.Errorf("student_id must not be empty")
	}

	if req.StudyPlanItemIdentity.StudyPlanId == "" {
		return fmt.Errorf("study_plan_id must not be empty")
	}

	return nil
}
