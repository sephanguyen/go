package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	db "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// StudentAssignmentReaderService implement business logic
type StudentAssignmentReaderService struct {
	DB             database.Ext
	SubmissionRepo interface {
		List(context.Context, database.QueryExecer, *repositories.StudentSubmissionFilter) (entities.StudentSubmissions, error)
		ListV2(context.Context, database.QueryExecer, *repositories.StudentSubmissionFilter) (entities.StudentSubmissions, error)
		RetrieveByStudyPlanItemIDs(context.Context, database.QueryExecer, pgtype.TextArray) (entities.StudentSubmissions, error)
	}
	StudentRepo interface {
		FindStudentsByCourseID(context.Context, database.QueryExecer, pgtype.Text) (*pgtype.TextArray, error)
		FindStudentsByClassIDs(context.Context, database.QueryExecer, pgtype.TextArray) (*pgtype.TextArray, error)
		FindStudentsByCourseLocation(ctx context.Context, db database.QueryExecer, courseID pgtype.Text, locationID pgtype.TextArray) (*pgtype.TextArray, error)
		FindStudentsByLocation(ctx context.Context, db database.QueryExecer, locationIDs pgtype.TextArray) (*pgtype.TextArray, error)
	}
	GradeRepo interface {
		RetrieveByIDs(context.Context, database.QueryExecer, pgtype.TextArray) (entities.StudentSubmissionGrades, error)
	}
	StudyPlanRepo interface {
		FindByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.StudyPlan, error)
	}
	StudyPlanItemRepo interface {
		FindByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.StudyPlanItem, error)
	}
}

func (s *StudentAssignmentReaderService) getStudentIDs(ctx context.Context, req *pb.ListSubmissionsRequest) (*pgtype.TextArray, error) {
	if len(req.ClassIds) > 0 {
		return s.StudentRepo.FindStudentsByClassIDs(ctx, s.DB, database.TextArray(req.ClassIds))
	}

	if req.CourseId != nil && req.CourseId.Value != "" {
		return s.StudentRepo.FindStudentsByCourseID(ctx, s.DB, database.Text(req.CourseId.Value))
	}

	return nil, nil
}

// ListSubmissions should be called by admin
func (s *StudentAssignmentReaderService) ListSubmissions(ctx context.Context, req *pb.ListSubmissionsRequest) (*pb.ListSubmissionsResponse, error) {
	if req.Paging == nil {
		return nil, status.Error(codes.InvalidArgument, "empty paging")
	}

	filter := &repositories.StudentSubmissionFilter{
		Limit: uint(req.Paging.Limit),
	}

	if err := multierr.Combine(
		filter.OffsetID.Set(nil),
		filter.CreatedAt.Set(nil),
		filter.StudentIDs.Set(nil),
		filter.Statuses.Set(nil),
		filter.StartDate.Set(nil),
		filter.EndDate.Set(nil),
		filter.AssignmentName.Set(nil),
		filter.CourseID.Set(nil),
		filter.ClassIDs.Set(nil),
	); err != nil {
		return nil, err
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

	if req.SearchText != nil && req.SearchText.Value != "" {
		if req.SearchType == pb.SearchType_SEARCH_TYPE_ASSIGNMENT_NAME {
			filter.AssignmentName.Set("%" + req.SearchText.Value + "%")
		}
	}

	studentIDs, err := s.getStudentIDs(ctx, req)
	if err != nil {
		return nil, err
	}

	if studentIDs != nil {
		filter.StudentIDs = *studentIDs
	}

	var itemsMap map[string]*entities.StudyPlanItem
	var coursesMap map[string]string
	var ss entities.StudentSubmissions
	err = database.ExecInTxWithRetry(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		ss, err = s.SubmissionRepo.List(ctx, tx, filter)
		if err != nil {
			// return nil, err
			return err
		}
		if len(ss) == 0 {
			return nil
		}

		itemsMap, coursesMap, err = s.getStudyPlansInfo(ctx, tx, ss)
		if err != nil {
			// return nil, err
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	items := toSubmissionsProto(ss, itemsMap, coursesMap)
	var next *cpb.Paging
	if len(items) == int(req.Paging.Limit) {
		next = &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetCombined{
				OffsetCombined: &cpb.Paging_Combined{
					OffsetString: items[len(items)-1].SubmissionId,
					OffsetTime:   items[len(items)-1].UpdatedAt,
				},
			},
		}
	}

	return &pb.ListSubmissionsResponse{
		NextPage: next,
		Items:    items,
	}, nil
}

func (s *StudentAssignmentReaderService) getStudyPlansInfo(ctx context.Context, db database.QueryExecer, submissions []*entities.StudentSubmission) (map[string]*entities.StudyPlanItem, map[string]string, error) {
	var studyPlanIDs, studyPlanItemIDs []string
	for _, sub := range submissions {
		studyPlanItemIDs = append(studyPlanItemIDs, sub.StudyPlanItemID.String)
	}
	studyPlanItems, err := s.StudyPlanItemRepo.FindByIDs(ctx, db, database.TextArray(golibs.Uniq(studyPlanItemIDs)))
	if err != nil {
		return nil, nil, err
	}

	for _, item := range studyPlanItems {
		studyPlanIDs = append(studyPlanIDs, item.StudyPlanID.String)
	}
	studyPlans, err := s.StudyPlanRepo.FindByIDs(ctx, db, database.TextArray(golibs.Uniq(studyPlanIDs)))
	if err != nil {
		return nil, nil, err
	}

	itemsMap := make(map[string]*entities.StudyPlanItem, len(studyPlanItems))
	for _, item := range studyPlanItems {
		itemsMap[item.ID.String] = item
	}
	coursesMap := make(map[string]string)
	for _, studyPlan := range studyPlans {
		coursesMap[studyPlan.ID.String] = studyPlan.CourseID.String
	}

	return itemsMap, coursesMap, nil
}

func toSubmissionsProto(e entities.StudentSubmissions, studyPlanItems map[string]*entities.StudyPlanItem, courses map[string]string) []*pb.StudentSubmission {
	results := make([]*pb.StudentSubmission, 0, len(e))

	for _, s := range e {
		i := &pb.StudentSubmission{
			SubmissionId:       s.ID.String,
			AssignmentId:       s.AssignmentID.String,
			StudyPlanItemId:    s.StudyPlanItemID.String,
			StudentId:          s.StudentID.String,
			CreatedAt:          timestamppb.New(s.CreatedAt.Time),
			UpdatedAt:          timestamppb.New(s.UpdatedAt.Time),
			Status:             pb.SubmissionStatus(pb.SubmissionStatus_value[s.Status.String]),
			Note:               s.Note.String,
			Duration:           s.Duration.Int,
			CorrectScore:       s.CorrectScore.Float,
			TotalScore:         s.TotalScore.Float,
			UnderstandingLevel: pb.SubmissionUnderstandingLevel(pb.SubmissionUnderstandingLevel_value[s.UnderstandingLevel.String]),
		}

		if s.CompleteDate.Status == pgtype.Present {
			i.CompleteDate = timestamppb.New(s.CompleteDate.Time)
		}

		if s.SubmissionGradeID.Status == pgtype.Present {
			i.SubmissionGradeId = wrapperspb.String(s.SubmissionGradeID.String)
		}

		s.SubmissionContent.AssignTo(&i.SubmissionContent)

		studyPlanItem := studyPlanItems[i.StudyPlanItemId]
		i.StartDate = timestamppb.New(studyPlanItem.StartDate.Time)
		i.EndDate = timestamppb.New(studyPlanItem.EndDate.Time)
		i.CourseId = courses[studyPlanItem.StudyPlanID.String]

		results = append(results, i)
	}
	return results
}

// RetrieveSubmissions return submissions by studyPlanItemIDs
func (s *StudentAssignmentReaderService) RetrieveSubmissions(ctx context.Context, req *pb.RetrieveSubmissionsRequest) (*pb.RetrieveSubmissionsResponse, error) {
	if len(req.StudyPlanItemIds) == 0 {
		return &pb.RetrieveSubmissionsResponse{
			Items: []*pb.StudentSubmission{},
		}, nil
	}

	var itemsMap map[string]*entities.StudyPlanItem
	var coursesMap map[string]string
	var ss entities.StudentSubmissions

	err := database.ExecInTxWithRetry(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		ss, err = s.SubmissionRepo.RetrieveByStudyPlanItemIDs(ctx, tx, database.TextArray(req.StudyPlanItemIds))
		if err != nil {
			// return nil, err
			return err
		}
		if len(ss) == 0 {
			return nil
		}

		itemsMap, coursesMap, err = s.getStudyPlansInfo(ctx, tx, ss)
		if err != nil {
			// return nil, err
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &pb.RetrieveSubmissionsResponse{
		Items: toSubmissionsProto(ss, itemsMap, coursesMap),
	}, nil
}

func toGradesProto(e entities.StudentSubmissionGrades) []*pb.RetrieveSubmissionGradesRespose_Grade {
	results := make([]*pb.RetrieveSubmissionGradesRespose_Grade, 0, len(e))
	for _, g := range e {
		var grade float64
		g.Grade.AssignTo(&grade)

		rg := &pb.RetrieveSubmissionGradesRespose_Grade{
			SubmissionGradeId: g.ID.String,
			Grade: &pb.SubmissionGrade{
				SubmissionId: g.StudentSubmissionID.String,
				Note:         g.GraderComment.String,
				Grade:        grade,
			},
		}
		g.GradeContent.AssignTo(&rg.Grade.GradeContent)
		results = append(results, rg)
	}

	return results
}

// RetrieveSubmissionGrades is a gRPC method
func (s *StudentAssignmentReaderService) RetrieveSubmissionGrades(ctx context.Context,
	req *pb.RetrieveSubmissionGradesRequest) (*pb.RetrieveSubmissionGradesRespose, error) {
	grades, err := s.GradeRepo.RetrieveByIDs(ctx, s.DB, database.TextArray(req.SubmissionGradeIds))
	if err != nil {
		return nil, err
	}

	return &pb.RetrieveSubmissionGradesRespose{
		Grades: toGradesProto(grades),
	}, nil
}

func (s *StudentAssignmentReaderService) ListSubmissionsV2(ctx context.Context, req *pb.ListSubmissionsV2Request) (*pb.ListSubmissionsV2Response, error) {
	if req.GetPaging() == nil {
		return nil, status.Error(codes.InvalidArgument, "empty paging")
	}

	filter := &repositories.StudentSubmissionFilter{
		Limit: uint(req.Paging.Limit),
	}

	if err := multierr.Combine(
		filter.OffsetID.Set(nil),
		filter.CreatedAt.Set(nil),
		filter.StudentIDs.Set(nil),
		filter.Statuses.Set(nil),
		filter.StartDate.Set(nil),
		filter.EndDate.Set(nil),
		filter.AssignmentName.Set(nil),
		filter.CourseID.Set(nil),
		filter.ClassIDs.Set(nil),
		filter.LocationIDs.Set(nil),
		filter.StudentName.Set(nil),
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

	if req.SearchText != nil && req.SearchText.Value != "" {
		if req.SearchType == pb.SearchType_SEARCH_TYPE_ASSIGNMENT_NAME {
			_ = filter.AssignmentName.Set(req.SearchText.Value)
		}
	}

	if len(req.LocationIds) > 0 {
		_ = filter.LocationIDs.Set(req.LocationIds)
	}

	if req.StudentName.GetValue() != "" {
		cleanValue := db.ReplaceSpecialChars(req.StudentName.Value)
		_ = filter.StudentName.Set(cleanValue)
	}

	var itemsMap map[string]*entities.StudyPlanItem
	var coursesMap map[string]string
	var ss entities.StudentSubmissions
	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		ss, err = s.SubmissionRepo.ListV2(ctx, tx, filter)
		if err != nil {
			// return nil, err
			return err
		}
		if len(ss) == 0 {
			return nil
		}

		itemsMap, coursesMap, err = s.getStudyPlansInfo(ctx, tx, ss)
		if err != nil {
			// return nil, err
			return err
		}

		return nil
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}
	items := toSubmissionsProto(ss, itemsMap, coursesMap)
	var next *cpb.Paging
	if len(items) == int(req.Paging.Limit) {
		next = &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetCombined{
				OffsetCombined: &cpb.Paging_Combined{
					OffsetString: items[len(items)-1].SubmissionId,
					OffsetTime:   items[len(items)-1].UpdatedAt,
				},
			},
		}
	}

	return &pb.ListSubmissionsV2Response{
		NextPage: next,
		Items:    items,
	}, nil
}
