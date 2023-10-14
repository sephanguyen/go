package services

import (
	"context"
	"fmt"
	"sort"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type CourseReaderService struct {
	pb.UnimplementedCourseReaderServiceServer
	DB  database.Ext
	Env string

	CourseClassRepo interface {
		FindClassIDByCourseID(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.Text) ([]string, error)
	}
	BobUserReader interface {
		SearchBasicProfile(ctx context.Context, in *bpb.SearchBasicProfileRequest, opts ...grpc.CallOption) (*bpb.SearchBasicProfileResponse, error)
	}
	CourseStudentRepo interface {
		FindStudentByCourseID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) ([]string, error)
		SearchStudents(ctx context.Context, db database.QueryExecer, filter *repositories.SearchStudentsFilter) (map[string][]string, []string, error)
	}

	StudyPlanItemRepo interface {
		RetrieveBookIDByStudyPlanID(ctx context.Context, db database.Ext, studyPlanID pgtype.Text) (string, error)
	}

	StudentStudyPlanRepo interface {
		GetBookIDsBelongsToStudentStudyPlan(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, bookIDs pgtype.TextArray) ([]string, error)
	}

	BookRepo interface {
		ListBooks(ctx context.Context, db database.QueryExecer, args *repositories.ListBooksArgs) ([]*entities.Book, error)
	}

	TopicRepo interface {
		FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs, topicIDs pgtype.TextArray, limit, offset pgtype.Int4) ([]*entities.Topic, error)
		FindByChapterIDs(ctx context.Context, db database.QueryExecer, chapterIDs pgtype.TextArray) ([]*entities.Topic, error)
	}

	CourseStudyPlanRepo interface {
		ListCourseStatisticItems(ctx context.Context, db database.QueryExecer, args *repositories.ListCourseStatisticItemsArgs) ([]*repositories.CourseStatisticItem, error)
		ListCourseStatisticItemsV2(ctx context.Context, db database.QueryExecer, args *repositories.ListCourseStatisticItemsArgsV2) ([]*repositories.CourseStatisticItemV2, error)
	}

	ShuffledQuizSetRepo interface {
		CalculateHighestSubmissionScore(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*repositories.CalculateHighestScoreResponse, error)
	}

	AssignmentRepo interface {
		CalculateHigestScore(ctx context.Context, db database.QueryExecer, assignmetIDs pgtype.TextArray) ([]*repositories.CalculateHighestScoreResponse, error)
		CalculateTaskAssignmentHighestScore(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*repositories.CalculateHighestScoreResponse, error)
	}

	ChapterRepo interface {
		FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs pgtype.TextArray) ([]*entities.Chapter, error)
	}
	LearningObjectiveRepoV2 interface {
		RetrieveLearningObjectivesByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.LearningObjectiveV2, error)
		CountLearningObjectivesByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) (int, error)
	}

	CourseBookRepo interface {
		FindByCourseIDsV2(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray) ([]*entities.CoursesBooks, error)
	}

	ClassService interface {
		RetrieveClassesByIDs(ctx context.Context, in *mpb.RetrieveClassByIDsRequest, opts ...grpc.CallOption) (*mpb.RetrieveClassByIDsResponse, error)
	}

	AssessmentRepo interface {
		GetAssessmentByCourseAndLearningMaterial(ctx context.Context, db database.QueryExecer, courseIDs, learningMaterialIDs pgtype.TextArray) ([]*entities.Assessment, error)
	}

	StudentRepo interface {
		FilterOutDeletedStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]string, error)
	}

	CourseStudentAccessPathRepo interface {
		GetByLocationsStudentsAndCourse(ctx context.Context, db database.QueryExecer, locationIDs, studentIDs, courseIDs pgtype.TextArray) ([]*entities.CourseStudentsAccessPath, error)
	}
}

func NewCourseReaderService(
	db database.Ext,
	bobUserReader bpb.UserReaderServiceClient,
	classService mpb.ClassServiceClient,
) *CourseReaderService {
	return &CourseReaderService{
		DB:                          db,
		CourseClassRepo:             &repositories.CourseClassRepo{},
		BobUserReader:               bobUserReader,
		CourseStudentRepo:           &repositories.CourseStudentRepo{},
		StudyPlanItemRepo:           &repositories.StudyPlanItemRepo{},
		StudentStudyPlanRepo:        &repositories.StudentStudyPlanRepo{},
		BookRepo:                    &repositories.BookRepo{},
		TopicRepo:                   &repositories.TopicRepo{},
		CourseStudyPlanRepo:         &repositories.CourseStudyPlanRepo{},
		AssignmentRepo:              &repositories.AssignmentRepo{},
		ShuffledQuizSetRepo:         &repositories.ShuffledQuizSetRepo{},
		ChapterRepo:                 &repositories.ChapterRepo{},
		LearningObjectiveRepoV2:     &repositories.LearningObjectiveRepoV2{},
		CourseBookRepo:              &repositories.CourseBookRepo{},
		AssessmentRepo:              &repositories.AssessmentRepo{},
		StudentRepo:                 &repositories.StudentRepo{},
		CourseStudentAccessPathRepo: &repositories.CourseStudentAccessPathRepo{},
		ClassService:                classService,
	}
}

func (crs *CourseReaderService) GetLOsByCourse(ctx context.Context, req *pb.GetLOsByCourseRequest) (*pb.GetLOsByCourseResponse, error) {
	var (
		bookIDs       []string
		topicIDs      []string
		responseLOs   []*pb.GetLOsByCourseResponse_LearningObject
		topics        []*entities.Topic
		LOs           []*entities.LearningObjectiveV2
		LOsPagination []*entities.LearningObjectiveV2
		offset        = req.Paging.GetOffsetInteger()
		limit         = req.Paging.GetLimit()
	)

	courseBooks, err := crs.CourseBookRepo.FindByCourseIDsV2(ctx, crs.DB, database.TextArray(req.CourseId))
	if err != nil {
		return nil, fmt.Errorf("s.BookRepo.GetBooksByCourse: %w", err)
	}

	for _, courseBook := range courseBooks {
		bookIDs = append(bookIDs, courseBook.BookID.String)
	}

	chapters, err := crs.ChapterRepo.FindByBookIDs(ctx, crs.DB, database.TextArray(bookIDs))
	if err != nil {
		return nil, fmt.Errorf("s.ChapterRepo.FindByBookIDs: %w", err)
	}

	for _, chapter := range chapters {
		topicsChapter, err := crs.TopicRepo.FindByChapterIDs(ctx, crs.DB, database.TextArray([]string{chapter.ID.String}))
		if err != nil {
			return nil, fmt.Errorf("s.TopicRepo.FindByChapterIDs: %w", err)
		}
		topics = append(topics, topicsChapter...)
	}

	for _, topic := range topics {
		topicIDs = append(topicIDs, topic.ID.String)
		LOsTopic, err := crs.LearningObjectiveRepoV2.RetrieveLearningObjectivesByTopicIDs(ctx, crs.DB, database.TextArray([]string{topic.ID.String}))
		if err != nil {
			return nil, fmt.Errorf("s.LearningObjectiveRepoV2.RetrieveLearningObjectivesByTopicIDs: %w", err)
		}
		LOs = append(LOs, LOsTopic...)
	}

	totalItems, err := crs.LearningObjectiveRepoV2.CountLearningObjectivesByTopicIDs(ctx, crs.DB, database.TextArray(topicIDs))
	if err != nil {
		return nil, fmt.Errorf("s.LearningObjectiveRepoV2.CountLearningObjectivesByTopicIDs: %w", err)
	}

	// pagination
	if len(LOs) > 0 {
		if len(LOs) >= int(offset)+int(limit) {
			LOsPagination = LOs[offset : offset+int64(limit)]
		} else if len(LOs) >= int(offset) {
			LOsPagination = LOs[offset:]
		}
	}

	mapTopicLo := make(map[string]string)

	for _, lo := range LOsPagination {
		for _, topic := range topics {
			if lo.TopicID == topic.ID {
				mapTopicLo[lo.LearningMaterial.ID.String] = topic.Name.String
			}
		}
	}

	for _, lo := range LOsPagination {
		assessments, err := crs.AssessmentRepo.GetAssessmentByCourseAndLearningMaterial(ctx, crs.DB, database.TextArray(req.CourseId), database.TextArray([]string{lo.LearningMaterial.ID.String}))
		if err != nil {
			return nil, fmt.Errorf("s.AssessmentRepo.GetAssessmentByCourseAndLearningMaterial: %w", err)
		}
		activityID := ""
		if len(assessments) > 0 {
			activityID = assessments[0].ID.String
		}
		temp := pb.GetLOsByCourseResponse_LearningObject{
			ActivityId:         activityID,
			LoName:             lo.Name.String,
			TopicName:          mapTopicLo[lo.ID.String],
			LearningMaterialId: lo.ID.String,
		}

		responseLOs = append(responseLOs, &temp)
	}

	return &pb.GetLOsByCourseResponse{
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: req.Paging.GetOffsetInteger() + int64(req.Paging.Limit),
			},
		},
		LOs:        responseLOs,
		TotalItems: int32(totalItems),
	}, nil
}

func (crs *CourseReaderService) ListClassByCourse(ctx context.Context, req *pb.ListClassByCourseRequest) (*pb.ListClassByCourseResponse, error) {
	classIDs, err := crs.CourseClassRepo.FindClassIDByCourseID(ctx, crs.DB, database.Text(req.CourseId))
	if err != nil {
		return nil, fmt.Errorf("s.CourseClassRepo.FindClassIDByCourseID: %w", err)
	}

	if len(req.LocationIds) != 0 && len(classIDs) != 0 {
		mdCtx, err := interceptors.GetOutgoingContext(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "GetOutgoingContext: %v", err)
		}
		res, err := crs.ClassService.RetrieveClassesByIDs(mdCtx, &mpb.RetrieveClassByIDsRequest{
			ClassIds: classIDs,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("could not get classes's location ids %v", err))
		}
		filterClassIDs := make([]string, 0)
		for _, class := range res.Classes {
			if slices.Contains(req.LocationIds, class.LocationId) {
				filterClassIDs = append(filterClassIDs, class.ClassId)
			}
		}
		classIDs = filterClassIDs
	}

	if len(classIDs) == 0 {
		classIDs = make([]string, 0)
	}

	return &pb.ListClassByCourseResponse{
		ClassIds: classIDs,
	}, nil
}

func (crs *CourseReaderService) ListStudentByCourse(ctx context.Context, req *pb.ListStudentByCourseRequest) (*pb.ListStudentByCourseResponse, error) {
	headers, ok := metadata.FromIncomingContext(ctx)
	var pkg, token, version string
	if ok {
		pkg = headers["pkg"][0]
		token = headers["token"][0]
		version = headers["version"][0]
	}

	if req.CourseId == "" {
		return nil, fmt.Errorf("CourseReaderService.ListStudentByCourse: No CourseId")
	}

	ids, err := crs.CourseStudentRepo.FindStudentByCourseID(ctx, crs.DB, database.Text(req.CourseId))
	if err != nil {
		return nil, fmt.Errorf("s.CourseStudentRepo.FindStudentByCourseID: %w", err)
	}

	if ids == nil {
		return &pb.ListStudentByCourseResponse{}, nil
	}
	rsp, err := crs.BobUserReader.SearchBasicProfile(metadata.AppendToOutgoingContext(ctx, "pkg", pkg, "version", version, "token", token), &bpb.SearchBasicProfileRequest{
		UserIds:    ids,
		SearchText: &wrappers.StringValue{Value: req.SearchText},
		Paging:     req.Paging,
	})
	if err != nil {
		return nil, err
	}
	return &pb.ListStudentByCourseResponse{
		Profiles: rsp.Profiles,
		NextPage: rsp.NextPage,
	}, nil
}

func (crs *CourseReaderService) ListStudentIDsByCourse(ctx context.Context, req *pb.ListStudentIDsByCourseRequest) (*pb.ListStudentIDsByCourseResponse, error) {
	if req.Paging == nil {
		req.Paging = &cpb.Paging{
			Limit:  100,
			Offset: nil,
		}
	}
	if req.Paging.Limit == 0 {
		req.Paging.Limit = 100
	}
	filter := &repositories.SearchStudentsFilter{}
	_ = filter.CourseIDs.Set(nil)
	_ = filter.StudentIDs.Set(nil)
	if req.CourseIds != nil {
		filter.CourseIDs = database.TextArray(req.CourseIds)
	}
	err := multierr.Combine(
		filter.Offset.Set(req.Paging.GetOffsetString()),
		filter.Limit.Set(req.Paging.Limit),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("cannot set SearchStudentsFilter %v", err))
	}
	studentCourses, studentIDs, err := crs.CourseStudentRepo.SearchStudents(ctx, crs.DB, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(studentIDs) == 0 {
		return &pb.ListStudentIDsByCourseResponse{}, nil
	}

	studentCoursesPb := make([]*pb.ListStudentIDsByCourseResponse_StudentCourses, 0, len(studentCourses))
	for studentID, courseIDs := range studentCourses {
		studentCoursesPb = append(studentCoursesPb, &pb.ListStudentIDsByCourseResponse_StudentCourses{
			StudentId: studentID,
			CourseIds: courseIDs,
		})
	}

	return &pb.ListStudentIDsByCourseResponse{
		StudentCourses: studentCoursesPb,
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: studentIDs[len(studentIDs)-1],
			},
		},
	}, nil
}

func (crs *CourseReaderService) ListCourseIDsByStudents(ctx context.Context, req *pb.ListCourseIDsByStudentsRequest) (*pb.ListCourseIDsByStudentsResponse, error) {
	if req.GetStudentIds() == nil || len(req.GetStudentIds()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "student ids have to not empty")
	}
	filter := &repositories.SearchStudentsFilter{}
	multierr.Combine(
		filter.CourseIDs.Set(nil),
		filter.StudentIDs.Set(nil),
		filter.Limit.Set(nil),
		filter.Offset.Set(nil))
	if req.StudentIds != nil {
		filter.StudentIDs = database.TextArray(req.StudentIds)
	}
	studentCourses, _, err := crs.CourseStudentRepo.SearchStudents(ctx, crs.DB, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("CourseStudentRepo.SearchStudents: %w", err).Error())
	}

	studentCoursesPb := make([]*pb.ListCourseIDsByStudentsResponse_StudentCourses, 0, len(studentCourses))
	for studentID, courseIDs := range studentCourses {
		studentCoursesPb = append(studentCoursesPb, &pb.ListCourseIDsByStudentsResponse_StudentCourses{
			StudentId: studentID,
			CourseIds: courseIDs,
		})
	}
	return &pb.ListCourseIDsByStudentsResponse{
		StudentCourses: studentCoursesPb,
	}, nil
}

func (crs *CourseReaderService) ListStudentIDsByCourseV2(req *pb.ListStudentIDsByCourseV2Request, stream pb.CourseReaderService_ListStudentIDsByCourseV2Server) error {
	filter := &repositories.SearchStudentsFilter{}
	_ = filter.CourseIDs.Set(nil)
	_ = filter.StudentIDs.Set(nil)
	if req.CourseIds != nil {
		filter.CourseIDs = database.TextArray(req.CourseIds)
	}

	ctx := stream.Context()
	ctx = interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: fmt.Sprintf("%v", req.SchoolId),
			UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
		},
	})

	var numOfStudentPerPage int64 = 2000
	var offset pgtype.Text = pgtype.Text{Status: pgtype.Null}
	var limit pgtype.Int8 = database.Int8(numOfStudentPerPage)

	for {
		err := multierr.Combine(
			filter.Offset.Set(offset),
			filter.Limit.Set(limit),
		)
		if err != nil {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("cannot set SearchStudentsFilter %v", err))
		}
		studentCourses, studentIDs, err := crs.SearchStudents(ctx, filter)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		if len(studentIDs) == 0 {
			break
		}

		for studentID, courseIDs := range studentCourses {
			e := &pb.ListStudentIDsByCourseV2Response_StudentCourses{
				StudentId: studentID,
				CourseIds: courseIDs,
			}
			if err := stream.Send(&pb.ListStudentIDsByCourseV2Response{
				StudentCourses: e,
			}); err != nil {
				return err
			}
		}

		offset = database.Text(studentIDs[len(studentIDs)-1])
	}

	return nil
}

func (crs *CourseReaderService) SearchStudents(ctx context.Context, filter *repositories.SearchStudentsFilter) (map[string][]string, []string, error) {
	studentCourses, studentIDs, err := crs.CourseStudentRepo.SearchStudents(ctx, crs.DB, filter)
	if err != nil {
		return nil, nil, fmt.Errorf("CourseStudentRepo.SearchStudents %v", err)
	}

	return studentCourses, studentIDs, nil
}

func (crs *CourseReaderService) ListTopicsByStudyPlan(ctx context.Context, req *pb.ListTopicsByStudyPlanRequest) (*pb.ListTopicsByStudyPlanResponse, error) {
	if err := crs.verifyListTopicsByStudyPlanRequest(req); err != nil {
		return nil, err
	}

	if req.Paging.Limit <= 0 {
		req.Paging.Limit = 100
	}

	offset := req.Paging.GetOffsetInteger()
	limit := req.Paging.Limit

	bookID, err := crs.StudyPlanItemRepo.RetrieveBookIDByStudyPlanID(ctx, crs.DB, database.Text(req.StudyPlanId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("crs.StudyPlanItemRepo.RetrieveBookIDByStudyPlanID: %w", err).Error())
	}

	topics, err := crs.retrieveTopics(ctx, req.Paging, []string{bookID}, nil)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("crs.retrieveTopics: %w", err).Error())
	}

	topicPbs := make([]*cpb.Topic, 0, len(topics))
	for _, topic := range topics {
		topicPbs = append(topicPbs, ToTopicPbV1(topic))
	}

	return &pb.ListTopicsByStudyPlanResponse{
		Items: topicPbs,
		NextPage: &cpb.Paging{
			Limit: uint32(limit),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: int64(limit) + int64(offset),
			},
		},
	}, nil
}

func (crs *CourseReaderService) verifyListTopicsByStudyPlanRequest(req *pb.ListTopicsByStudyPlanRequest) error {
	if req.StudyPlanId == "" {
		return status.Error(codes.InvalidArgument, "req must have study plan id")
	}

	if req.Paging == nil {
		return status.Error(codes.InvalidArgument, "req must have paging field")
	}

	if req.Paging.GetOffsetInteger() < 0 {
		return status.Error(codes.InvalidArgument, "offset must be positive")
	}

	return nil
}

func (crs *CourseReaderService) retrieveTopics(ctx context.Context, paging *cpb.Paging, bookIDs, topicIDs []string) (topics []*entities.Topic, err error) {
	if paging != nil {
		if paging.GetOffsetInteger() < 0 {
			return nil, status.Error(codes.InvalidArgument, "offset must be positive")
		}

		if paging.Limit <= 0 {
			paging.Limit = 100
		}

		offset := paging.GetOffsetInteger()
		limit := paging.GetLimit()

		topics, err = crs.TopicRepo.FindByBookIDs(ctx, crs.DB, database.TextArray(bookIDs), database.TextArray(topicIDs), database.Int4(int32(limit)), database.Int4(int32(offset)))
	} else {
		topics, err = crs.TopicRepo.FindByBookIDs(ctx, crs.DB, database.TextArray(bookIDs), database.TextArray(topicIDs), pgtype.Int4{Status: pgtype.Null}, pgtype.Int4{Status: pgtype.Null})
	}
	//TODO: hardcode here, will remove this repo later
	if err != nil && err != fmt.Errorf("database.Select: %w", pgx.ErrNoRows) {
		return nil, status.Error(codes.Internal, fmt.Errorf("crs.TopicRepo.FindByBookIDs: %w", err).Error())
	}

	return topics, nil
}

func (crs *CourseReaderService) validateCourseStatisticRequest(ctx context.Context, req *pb.RetrieveCourseStatisticRequest) (*repositories.ListCourseStatisticItemsArgs, error) {
	if len(req.CourseId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Missing course")
	}
	if len(req.StudyPlanId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Missing study plan")
	}
	args := &repositories.ListCourseStatisticItemsArgs{
		CourseID:    database.Text(req.CourseId),
		StudyPlanID: database.Text(req.StudyPlanId),
		ClassID:     pgtype.Text{Status: pgtype.Null},
	}
	if len(req.ClassId) != 0 {
		args.ClassID = database.Text(req.ClassId)
	}
	return args, nil
}

func (crs *CourseReaderService) validateCourseStatisticRequestV2(ctx context.Context, req *pb.RetrieveCourseStatisticRequestV2) (*repositories.ListCourseStatisticItemsArgsV2, error) {
	if len(req.CourseId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Missing course")
	}
	if len(req.StudyPlanId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Missing study plan")
	}
	args := &repositories.ListCourseStatisticItemsArgsV2{
		CourseID:    database.Text(req.CourseId),
		StudyPlanID: database.Text(req.StudyPlanId),
		ClassID:     pgtype.TextArray{Status: pgtype.Null},
	}

	if len(req.ClassId) != 0 {
		args.ClassID = database.TextArray(req.ClassId)
	}
	return args, nil
}

// [Topic][StudyPlanItem] => [(CourseStatisticItem)...]
type statisticMapStudyPlanItem = map[string]map[string][]*repositories.CourseStatisticItem

// [Topic][Student] => [(CourseStatisticItem)...]
type statisticMapStudent = map[string]map[string][]*repositories.CourseStatisticItem

// [StudyPlanItem]order
type statisticMapStudyPlanItemOrder = map[string]int

// [Topic][RootStudyPlan] => [(CourseStatisticItemV2)...]
type statisticMapStudyPlanItemV2 = map[string]map[string][]*repositories.CourseStatisticItemV2

// [Topic][Student] => [(CourseStatisticItemV2)...]
type statisticMapStudentV2 = map[string]map[string][]*repositories.CourseStatisticItemV2

func (crs *CourseReaderService) RetrieveCourseStatistic(ctx context.Context, req *pb.RetrieveCourseStatisticRequest) (*pb.RetrieveCourseStatisticResponse, error) {
	listCourseStatisticItemsArgs, err := crs.validateCourseStatisticRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	courseStatisticItems, err := crs.CourseStudyPlanRepo.ListCourseStatisticItems(ctx, crs.DB, listCourseStatisticItemsArgs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "CourseStudyPlanRepo.ListCourseStatisticItems %v", err.Error())
	}

	assignmentStudyPlanItemIDs, taskAssignmentStudyPlanItemIDs, LOStudyPlanItemIDs := getStudyPlanItemIDByTypeFromCourseStatisticItems(courseStatisticItems)

	assignmentScores, err := crs.AssignmentRepo.CalculateHigestScore(ctx, crs.DB, assignmentStudyPlanItemIDs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "AssignmentRepo.CalculateHigestScore %v", err.Error())
	}

	taskAssignmentScores, err := crs.AssignmentRepo.CalculateTaskAssignmentHighestScore(ctx, crs.DB, taskAssignmentStudyPlanItemIDs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "AssignmentRepo.CalculateTaskAssignmentHighestScore %v", err.Error())
	}

	loScores, err := crs.ShuffledQuizSetRepo.CalculateHighestSubmissionScore(ctx, crs.DB, LOStudyPlanItemIDs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ShuffledQuizSetRepo.CalculateHighestSubmissionScore %v", err.Error())
	}

	scores := mergeStudyPlanItemScore(assignmentScores, taskAssignmentScores, loScores)

	statisticMapStudyPlanItem, statisticMapStudent, statisticMapOrder := createStatisticMaps(courseStatisticItems, scores)
	resp := &pb.RetrieveCourseStatisticResponse{}

	for i, courseStatisticItem := range courseStatisticItems {
		// new topic item
		if i == 0 || courseStatisticItem.ContentStructure.TopicID != courseStatisticItems[i-1].ContentStructure.TopicID {
			topicItem := calculateTopicStudyPlanItemStatistic(statisticMapStudyPlanItem, statisticMapStudent, statisticMapOrder, courseStatisticItem.ContentStructure.TopicID)
			resp.CourseStatisticItems = append(resp.CourseStatisticItems, topicItem)
		}
	}

	return resp, nil
}

func (crs *CourseReaderService) RetrieveCourseStatisticV2(ctx context.Context, req *pb.RetrieveCourseStatisticRequestV2) (*pb.RetrieveCourseStatisticResponseV2, error) {
	listCourseStatisticArgs, err := crs.validateCourseStatisticRequestV2(ctx, req)
	if err != nil {
		return nil, err
	}

	courseStatisticItems, err := crs.CourseStudyPlanRepo.ListCourseStatisticItemsV2(ctx, crs.DB, listCourseStatisticArgs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "CourseStudyPlanRepo.ListCourseStatisticItemsV2 %v", err.Error())
	}

	assignmentStudyPlanItemIDs, taskAssignmentStudyPlanItemIDs, LOStudyPlanItemIDs := getStudyPlanItemIDByTypeFromCourseStatisticItemsV2(courseStatisticItems)

	assignmentScores, err := crs.AssignmentRepo.CalculateHigestScore(ctx, crs.DB, assignmentStudyPlanItemIDs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "AssignmentRepo.CalculateHigestScore %v", err.Error())
	}

	taskAssignmentScores, err := crs.AssignmentRepo.CalculateTaskAssignmentHighestScore(ctx, crs.DB, taskAssignmentStudyPlanItemIDs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "AssignmentRepo.CalculateTaskAssignmentHighestScore %v", err.Error())
	}

	loScores, err := crs.ShuffledQuizSetRepo.CalculateHighestSubmissionScore(ctx, crs.DB, LOStudyPlanItemIDs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ShuffledQuizSetRepo.CalculateHighestSubmissionScore %v", err.Error())
	}

	scores := mergeStudyPlanItemScore(assignmentScores, taskAssignmentScores, loScores)

	statisticMapStudyPlanItem, statisticMapStudent, statisticMapOrder := createStatisticMapsV2(courseStatisticItems, scores)
	resp := &pb.RetrieveCourseStatisticResponseV2{}

	for i, topicStatisticItem := range courseStatisticItems {
		// new topic item
		if i == 0 || topicStatisticItem.ContentStructure.TopicID != courseStatisticItems[i-1].ContentStructure.TopicID {
			topicItem := calculateCourseStatisticV2(statisticMapStudyPlanItem, statisticMapStudent, statisticMapOrder, topicStatisticItem.ContentStructure.TopicID)
			resp.TopicStatistic = append(resp.TopicStatistic, topicItem)
		}
	}

	return resp, nil
}

func calculateTopicStudyPlanItemStatistic(statisticMapLO statisticMapStudyPlanItem, statisticMapStudent statisticMapStudent, orderMap statisticMapStudyPlanItemOrder, topicID string) *pb.RetrieveCourseStatisticResponse_CourseStatisticItem {
	topicItem := &pb.RetrieveCourseStatisticResponse_CourseStatisticItem{
		TopicId: topicID,
	}

	sumAverageScore := int32(0)

	// countStudentsHaveScoreStudyPlanItem counts number of study plans which has student's scores
	countStudentsHaveScoreStudyPlanItem := int32(0)

	for studyPlanItemID := range statisticMapLO[topicID] {
		studyPlanItem := &pb.RetrieveCourseStatisticResponse_CourseStatisticItem_StudyPlanItemStatisticItem{
			StudyPlanItemId: studyPlanItemID,
		}

		var countStudentsHaveScore int32 = 0

		studyPlanItem.TotalAssignedStudent, studyPlanItem.CompletedStudent, studyPlanItem.AverageScore, countStudentsHaveScore = calculateStudyPlanItemStatisticAssignedCompletedScore(statisticMapLO, topicID, studyPlanItemID)

		if countStudentsHaveScore > 0 {
			sumAverageScore += studyPlanItem.AverageScore
			countStudentsHaveScoreStudyPlanItem++
		}

		topicItem.StudyPlanItemStatisticItems = append(topicItem.StudyPlanItemStatisticItems, studyPlanItem)
	}

	if countStudentsHaveScoreStudyPlanItem > 0 {
		topicItem.AverageScore = int32(float32(sumAverageScore)/float32(countStudentsHaveScoreStudyPlanItem) + 0.5)
	} else {
		topicItem.AverageScore = -1
	}

	topicItem.TotalAssignedStudent, topicItem.CompletedStudent = calculateTopicStatisticAssignedCompleted(statisticMapStudent, topicID)

	// original order
	sort.Slice(topicItem.StudyPlanItemStatisticItems, func(i, j int) bool {
		iID := topicItem.StudyPlanItemStatisticItems[i].StudyPlanItemId
		jID := topicItem.StudyPlanItemStatisticItems[j].StudyPlanItemId
		return orderMap[iID] < orderMap[jID]
	})

	return topicItem
}

func calculateCourseStatisticV2(statisticMapLO statisticMapStudyPlanItemV2, statisticMapStudent statisticMapStudentV2, orderMap statisticMapStudyPlanItemOrder, topicID string) *pb.RetrieveCourseStatisticResponseV2_TopicStatistic {
	topicItem := &pb.RetrieveCourseStatisticResponseV2_TopicStatistic{
		TopicId: topicID,
	}

	sumAverageScore := int32(0)

	// countStudentsHaveScoreStudyPlanItem counts number of study plans which has student's scores
	countStudentsHaveScoreStudyPlanItem := int32(0)

	for rootStudyPlanItemID := range statisticMapLO[topicID] {
		learningMaterial := &pb.RetrieveCourseStatisticResponseV2_TopicStatistic_LearningMaterialStatistic{StudyPlanItemId: rootStudyPlanItemID}

		var countStudentsHaveScore int32

		learningMaterial.TotalAssignedStudent, learningMaterial.CompletedStudent, learningMaterial.AverageScore, countStudentsHaveScore, learningMaterial.LearningMaterialId = calculateStudyPlanItemStatisticAssignedCompletedScoreV2(statisticMapLO, topicID, rootStudyPlanItemID)
		if countStudentsHaveScore > 0 {
			sumAverageScore += learningMaterial.AverageScore
			countStudentsHaveScoreStudyPlanItem++
		}

		topicItem.LearningMaterialStatistic = append(topicItem.LearningMaterialStatistic, learningMaterial)
	}

	if countStudentsHaveScoreStudyPlanItem > 0 {
		topicItem.AverageScore = int32(float32(sumAverageScore)/float32(countStudentsHaveScoreStudyPlanItem) + 0.5)
	} else {
		topicItem.AverageScore = -1
	}

	topicItem.TotalAssignedStudent, topicItem.CompletedStudent = calculateTopicStatisticAssignedCompletedV2(statisticMapStudent, topicID)

	// original order
	sort.Slice(topicItem.LearningMaterialStatistic, func(i, j int) bool {
		iID := topicItem.LearningMaterialStatistic[i].StudyPlanItemId
		jID := topicItem.LearningMaterialStatistic[j].StudyPlanItemId
		return orderMap[iID] < orderMap[jID]
	})

	return topicItem
}

// studyPlanItemId is oringinal (or master) study plan item id
func calculateStudyPlanItemStatisticAssignedCompletedScore(statisticMapStudyPlanItem statisticMapStudyPlanItem, topicID, studyPlanItemID string) (int32, int32, int32, int32) {
	items := statisticMapStudyPlanItem[topicID][studyPlanItemID]

	totalAssignedStudent := int32(0)
	completedStudent := int32(0)
	averageScore := int32(0)
	sumAverageScore := float32(0)
	countStudentsHaveScore := int32(0)

	for _, item := range items {
		if item.Status == pb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE.String() {
			totalAssignedStudent++
			if item.CompletedAt.Status == pgtype.Present {
				completedStudent++
				// exist score
				if item.Score >= 0 {
					countStudentsHaveScore++
					sumAverageScore += item.Score
				}
			}
		}
	}

	if countStudentsHaveScore > 0 {
		averageScore = int32(sumAverageScore/float32(countStudentsHaveScore) + 0.5)
	} else {
		averageScore = -1
	}

	return totalAssignedStudent, completedStudent, averageScore, countStudentsHaveScore
}

// This function calculate totalAssignedStudent, completedStudent, averageScore, countStudentsHaveScore
// of all studyPlanItemID and merge the reuslt in a RootStudyplanItemID
func calculateStudyPlanItemStatisticAssignedCompletedScoreV2(statisticMapStudyPlanItem statisticMapStudyPlanItemV2, topicID, rootStudyPlanItemID string) (int32, int32, int32, int32, string) {
	items := statisticMapStudyPlanItem[topicID][rootStudyPlanItemID]

	totalAssignedStudent := int32(0)
	completedStudent := int32(0)
	averageScore := int32(0)
	sumAverageScore := float32(0)
	countStudentsHaveScore := int32(0)
	var lmID string

	for _, item := range items {
		if item.Status == pb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE.String() {
			totalAssignedStudent++
			if item.CompletedAt.Status == pgtype.Present {
				completedStudent++
				// exist score
				if item.Score >= 0 {
					countStudentsHaveScore++
					sumAverageScore += item.Score
				}
			}
		}
	}

	// Learning materialID is the same with 1 rootsStudyPlanItemID
	if len(items) > 0 {
		lmID = items[0].LearningMaterialID
	}

	if countStudentsHaveScore > 0 {
		averageScore = int32(sumAverageScore/float32(countStudentsHaveScore) + 0.5)
	} else {
		averageScore = -1
	}

	return totalAssignedStudent, completedStudent, averageScore, countStudentsHaveScore, lmID
}

func calculateTopicStatisticAssignedCompleted(statisticMapStudent statisticMapStudent, topicID string) (int32, int32) {
	totalAssignedStudent := int32(0)
	completedStudent := int32(0)

	for _ /*studentID*/, items := range statisticMapStudent[topicID] {
		countStudentActiveItem := 0
		countStudentCompletedItem := 0
		for _, item := range items {
			if item.Status == pb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE.String() {
				countStudentActiveItem++
				if item.CompletedAt.Status == pgtype.Present {
					countStudentCompletedItem++
				}
			}
		}
		// at least 1 active item in this topic for this student
		if countStudentActiveItem > 0 {
			totalAssignedStudent++
			// this student completes all items for this topic
			if countStudentCompletedItem == countStudentActiveItem {
				completedStudent++
			}
		}
	}
	return totalAssignedStudent, completedStudent
}

func calculateTopicStatisticAssignedCompletedV2(statisticMapStudent statisticMapStudentV2, topicID string) (int32, int32) {
	totalAssignedStudent := int32(0)
	completedStudent := int32(0)

	for _ /*studentID*/, items := range statisticMapStudent[topicID] {
		countStudentActiveItem := 0
		countStudentCompletedItem := 0
		for _, item := range items {
			if item.Status == pb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE.String() {
				countStudentActiveItem++
				if item.CompletedAt.Status == pgtype.Present {
					countStudentCompletedItem++
				}
			}
		}
		// at least 1 active item in this topic for this student
		if countStudentActiveItem > 0 {
			totalAssignedStudent++
			// this student completes all items for this topic
			if countStudentCompletedItem == countStudentActiveItem {
				completedStudent++
			}
		}
	}
	return totalAssignedStudent, completedStudent
}

func createStatisticMaps(items []*repositories.CourseStatisticItem, scores map[string]float32) (statisticMapStudyPlanItem, statisticMapStudent, statisticMapStudyPlanItemOrder) {
	spi := make(statisticMapStudyPlanItem)
	student := make(statisticMapStudent)
	order := make(statisticMapStudyPlanItemOrder)
	for i, courseStatisticItem := range items {
		order[courseStatisticItem.RootStudyPlanItemID] = i
		if score, found := scores[courseStatisticItem.StudyPlanItemID]; found {
			courseStatisticItem.Score = score
		} else {
			courseStatisticItem.Score = -1
		}
		if spi[courseStatisticItem.ContentStructure.TopicID] == nil {
			spi[courseStatisticItem.ContentStructure.TopicID] = make(map[string][]*repositories.CourseStatisticItem)
		}
		if student[courseStatisticItem.ContentStructure.TopicID] == nil {
			student[courseStatisticItem.ContentStructure.TopicID] = make(map[string][]*repositories.CourseStatisticItem)
		}

		spi[courseStatisticItem.ContentStructure.TopicID][courseStatisticItem.RootStudyPlanItemID] = append(spi[courseStatisticItem.ContentStructure.TopicID][courseStatisticItem.RootStudyPlanItemID], courseStatisticItem)
		student[courseStatisticItem.ContentStructure.TopicID][courseStatisticItem.StudentID] = append(student[courseStatisticItem.ContentStructure.TopicID][courseStatisticItem.StudentID], courseStatisticItem)
	}
	return spi, student, order
}

func createStatisticMapsV2(items []*repositories.CourseStatisticItemV2, scores map[string]float32) (statisticMapStudyPlanItemV2, statisticMapStudentV2, statisticMapStudyPlanItemOrder) {
	spi := make(statisticMapStudyPlanItemV2)
	student := make(statisticMapStudentV2)
	order := make(statisticMapStudyPlanItemOrder)
	for i, topicStatisticItem := range items {
		order[topicStatisticItem.RootStudyPlanItemID] = i
		if score, found := scores[topicStatisticItem.StudyPlanItemID]; found {
			topicStatisticItem.Score = score
		} else {
			topicStatisticItem.Score = -1
		}
		if spi[topicStatisticItem.ContentStructure.TopicID] == nil {
			spi[topicStatisticItem.ContentStructure.TopicID] = make(map[string][]*repositories.CourseStatisticItemV2)
		}
		if student[topicStatisticItem.ContentStructure.TopicID] == nil {
			student[topicStatisticItem.ContentStructure.TopicID] = make(map[string][]*repositories.CourseStatisticItemV2)
		}

		spi[topicStatisticItem.ContentStructure.TopicID][topicStatisticItem.RootStudyPlanItemID] = append(spi[topicStatisticItem.ContentStructure.TopicID][topicStatisticItem.RootStudyPlanItemID], topicStatisticItem)
		student[topicStatisticItem.ContentStructure.TopicID][topicStatisticItem.StudentID] = append(student[topicStatisticItem.ContentStructure.TopicID][topicStatisticItem.StudentID], topicStatisticItem)
	}

	return spi, student, order
}

func getStudyPlanItemIDByTypeFromCourseStatisticItems(items []*repositories.CourseStatisticItem) (pgtype.TextArray, pgtype.TextArray, pgtype.TextArray) {
	var assignment, taskAssignment, lo []string
	for _, item := range items {
		if item.CompletedAt.Status != pgtype.Present {
			continue
		}
		if item.Status != pb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE.String() {
			continue
		}
		if len(item.ContentStructure.AssignmentID) != 0 {
			assignment = append(assignment, item.StudyPlanItemID)
			// TODO(giahuy): separate task assignment
			taskAssignment = append(taskAssignment, item.StudyPlanItemID)
		}
		if len(item.ContentStructure.LoID) != 0 {
			lo = append(lo, item.StudyPlanItemID)
		}
	}
	assignment = golibs.Uniq(assignment)
	taskAssignment = golibs.Uniq(taskAssignment)
	lo = golibs.Uniq(lo)
	return database.TextArray(assignment), database.TextArray(taskAssignment), database.TextArray(lo)
}

func getStudyPlanItemIDByTypeFromCourseStatisticItemsV2(items []*repositories.CourseStatisticItemV2) (pgtype.TextArray, pgtype.TextArray, pgtype.TextArray) {
	var assignment, taskAssignment, lo []string
	for _, item := range items {
		if item.CompletedAt.Status != pgtype.Present {
			continue
		}
		if item.Status != pb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE.String() {
			continue
		}
		if len(item.ContentStructure.AssignmentID) != 0 {
			assignment = append(assignment, item.StudyPlanItemID)
			taskAssignment = append(taskAssignment, item.StudyPlanItemID)
		}
		if len(item.ContentStructure.LoID) != 0 {
			lo = append(lo, item.StudyPlanItemID)
		}
	}
	assignment = golibs.Uniq(assignment)
	taskAssignment = golibs.Uniq(taskAssignment)
	lo = golibs.Uniq(lo)
	return database.TextArray(assignment), database.TextArray(taskAssignment), database.TextArray(lo)
}

func mergeStudyPlanItemScore(inputs ...[]*repositories.CalculateHighestScoreResponse) map[string]float32 {
	result := make(map[string]float32)
	for _, scores := range inputs {
		for _, score := range scores {
			result[score.StudyPlanItemID.String] = score.Percentage.Float
		}
	}
	return result
}

func (crs *CourseReaderService) GetStudentsAccessPath(ctx context.Context, req *pb.GetStudentsAccessPathRequest) (*pb.GetStudentsAccessPathResponse, error) {
	listLocationIDsFilter := req.GetLocationIds()
	studentIDs := req.GetStudentIds()
	courseIDs := req.GetCourseIds()
	courseStudentAccessPaths := []*entities.CourseStudentsAccessPath{}
	validStudentIDs, err := crs.StudentRepo.FilterOutDeletedStudentIDs(ctx, crs.DB, studentIDs)
	if err != nil && err != pgx.ErrNoRows {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.StudentRepo.FilterOutDeletedStudentIDs %w", err).Error())
	}
	if len(listLocationIDsFilter) > 0 {
		courseStudentAccessPaths, err = crs.CourseStudentAccessPathRepo.GetByLocationsStudentsAndCourse(ctx, crs.DB, database.TextArray(listLocationIDsFilter), database.TextArray(validStudentIDs), database.TextArray(courseIDs))
		if err != nil && err != pgx.ErrNoRows {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("s.CourseStudentAccessPathRepo.GetByLocationsStudentsAndCourse %w", err).Error())
		}
	}
	resCSAP := []*pb.GetStudentsAccessPathResponse_CourseStudentAccessPathObject{}
	for _, row := range courseStudentAccessPaths {
		resCSAP = append(resCSAP, &pb.GetStudentsAccessPathResponse_CourseStudentAccessPathObject{
			CourseStudentId: row.CourseStudentID.String,
			CourseId:        row.CourseID.String,
			StudentId:       row.StudentID.String,
		})
	}

	return &pb.GetStudentsAccessPathResponse{
		CourseStudentAccesssPaths: resCSAP,
	}, nil
}
