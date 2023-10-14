package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/services"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type BobInternalReaderServiceClient interface {
	RetrieveTopics(ctx context.Context, in *bpb.RetrieveTopicsRequest, opts ...grpc.CallOption) (*bpb.RetrieveTopicsResponse, error)
}

type StudyPlanReaderService struct {
	DB            database.Ext
	StudyPlanRepo interface {
		RetrieveByCourseID(ctx context.Context, db database.QueryExecer, args *repositories.RetrieveStudyPlanByCourseArgs) ([]*entities.StudyPlan, error)
	}

	StudyPlanItemRepo interface {
		FetchByStudyProgressRequest(ctx context.Context, db database.QueryExecer, courseID pgtype.Text, bookID pgtype.Text, studentID pgtype.Text) ([]*entities.StudyPlanItem, error)
	}

	StudentStudyPlanRepo interface {
		GetBookIDsBelongsToStudentStudyPlan(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, bookIDs pgtype.TextArray) ([]string, error)
		GetByStudyPlanStudentAndLO(ctx context.Context, db database.QueryExecer, studyPlanIDs, studentIDs, LoIDs pgtype.TextArray) ([]*entities.StudentStudyPlan, error)
	}

	AssignmentRepo interface {
		CalculateHigestScore(ctx context.Context, db database.QueryExecer, assignmetIDs pgtype.TextArray) ([]*repositories.CalculateHighestScoreResponse, error)
		CalculateTaskAssignmentHighestScore(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*repositories.CalculateHighestScoreResponse, error)
	}

	ShuffledQuizSetRepo interface {
		CalculateHighestSubmissionScore(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*repositories.CalculateHighestScoreResponse, error)
	}

	StudentLearningTimeDailyRepo interface {
		Retrieve(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, queryEnhancers ...repositories.QueryEnhancer) ([]*entities.StudentLearningTimeDaily, error)
		RetrieveV2(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, queryEnhancers ...repositories.QueryEnhancer) ([]*repositories.StudentLearningTimeDailyV2, error)
	}

	StudentLOCompletenessRepo interface {
		RetrieveFinishedLOs(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]*entities.StudentsLearningObjectivesCompleteness, error)
	}

	StudentEventLogRepo interface {
		RetrieveStudentEventLogsByStudyPlanItemIDs(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*entities.StudentEventLog, error)
	}

	InternalReaderService BobInternalReaderServiceClient

	LearningTimeCalculator *LearningTimeCalculator
}

func NewStudyPlanReaderService(db database.Ext, internalReaderService BobInternalReaderServiceClient) pb.StudyPlanReaderServiceServer {
	return &StudyPlanReaderService{
		DB:                           db,
		StudyPlanRepo:                &repositories.StudyPlanRepo{},
		StudentStudyPlanRepo:         &repositories.StudentStudyPlanRepo{},
		StudyPlanItemRepo:            &repositories.StudyPlanItemRepo{},
		AssignmentRepo:               &repositories.AssignmentRepo{},
		StudentEventLogRepo:          &repositories.StudentEventLogRepo{},
		InternalReaderService:        internalReaderService,
		LearningTimeCalculator:       &LearningTimeCalculator{},
		ShuffledQuizSetRepo:          &repositories.ShuffledQuizSetRepo{},
		StudentLearningTimeDailyRepo: &repositories.StudentLearningTimeDailyRepo{},
		StudentLOCompletenessRepo:    &repositories.StudentsLearningObjectivesCompletenessRepo{},
	}
}

func (s *StudyPlanReaderService) ListStudyPlanByCourse(ctx context.Context, req *pb.ListStudyPlanByCourseRequest) (*pb.ListStudyPlanByCourseResponse, error) {
	if req.CourseId == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid argument: course id have to not empty")
	}
	args := &repositories.RetrieveStudyPlanByCourseArgs{
		CourseID:      database.Text(req.CourseId),
		Limit:         10,
		StudyPlanName: pgtype.Text{Status: pgtype.Null},
		StudyPlanID:   pgtype.Text{Status: pgtype.Null},
	}

	if paging := req.Paging; paging != nil {
		if limit := paging.Limit; 1 <= limit && limit <= 100 {
			args.Limit = limit
		}
		if c := paging.GetOffsetMultipleCombined(); c != nil {
			args.StudyPlanName = database.Text(c.GetCombined()[0].GetOffsetString())
			args.StudyPlanID = database.Text(c.GetCombined()[1].GetOffsetString())
		}
	}
	studyPlans, err := s.StudyPlanRepo.RetrieveByCourseID(ctx, s.DB, args)
	if err != nil {
		return nil, fmt.Errorf("StudyPlanRepo.RetrieveByCourseID: %w", err)
	}

	if len(studyPlans) == 0 {
		return &pb.ListStudyPlanByCourseResponse{}, nil
	}

	pbStudyPlans := toPbStudyPlans(studyPlans)
	lastStudyPlans := studyPlans[len(pbStudyPlans)-1]
	nextPage := &cpb.Paging{
		Limit: args.Limit,
		Offset: &cpb.Paging_OffsetMultipleCombined{
			OffsetMultipleCombined: &cpb.Paging_MultipleCombined{
				Combined: []*cpb.Paging_Combined{
					{
						OffsetString: lastStudyPlans.Name.String,
					},
					{
						OffsetString: lastStudyPlans.ID.String,
					},
				},
			},
		},
	}

	return &pb.ListStudyPlanByCourseResponse{
		StudyPlans: pbStudyPlans,
		NextPage:   nextPage,
	}, nil
}

func toPbStudyPlans(es []*entities.StudyPlan) []*pb.StudyPlan {
	res := make([]*pb.StudyPlan, 0, len(es))
	for _, e := range es {
		res = append(res, &pb.StudyPlan{
			StudyPlanId:         e.ID.String,
			Name:                e.Name.String,
			Status:              pb.StudyPlanStatus(pb.StudyPlanStatus_value[e.Status.String]),
			BookId:              e.BookID.String,
			TrackSchoolProgress: e.TrackSchoolProgress.Bool,
			Grades:              database.FromInt4Array(e.Grades),
		})
	}
	return res
}

func (s *StudyPlanReaderService) GetBookIDsBelongsToStudentStudyPlan(ctx context.Context, req *pb.GetBookIDsBelongsToStudentStudyPlanRequest) (*pb.GetBookIDsBelongsToStudentStudyPlanResponse, error) {
	if req.StudentId == "" {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("GetBookIDsBelongsToStudentStudyPlan: studentID is not provided").Error())
	}

	if len(req.BookIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("GetBookIDsBelongsToStudentStudyPlan: bookIDs is not provided").Error())
	}

	var studentIDReq pgtype.Text
	var bookIDsReq pgtype.TextArray

	err := multierr.Combine(
		studentIDReq.Set(req.StudentId),
		bookIDsReq.Set(req.BookIds),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("GetBookIDsBelongsToStudentStudyPlan: value invalid").Error())
	}

	bookIDs, err := s.StudentStudyPlanRepo.GetBookIDsBelongsToStudentStudyPlan(ctx, s.DB, studentIDReq, bookIDsReq)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("StudentStudyPlanRepo.GetBookIDsBelongsToStudentStudyPlan: %w", err).Error())
	}

	return &pb.GetBookIDsBelongsToStudentStudyPlanResponse{
		BookIds: bookIDs,
	}, nil
}

func (s *StudyPlanReaderService) validateStudentBookStudyProgressReq(ctx context.Context, req *pb.StudentBookStudyProgressRequest) error {
	if req.CourseId == "" {
		return status.Error(codes.InvalidArgument, fmt.Errorf("StudyPlanReaderService.StudentBookStudyProgress: course_id is empty").Error())
	}

	if req.BookId == "" {
		return status.Error(codes.InvalidArgument, fmt.Errorf("StudyPlanReaderService.StudentBookStudyProgress: book_id is empty").Error())
	}

	if req.StudentId == "" {
		return status.Error(codes.InvalidArgument, fmt.Errorf("StudyPlanReaderService.StudentBookStudyProgress: student_id is empty").Error())
	}
	return nil
}

// nolint
// Api for calculating student learning progress
// This will fetch all student study plan items from information given by request (course_id, book_id, student_id)
// Then will calculate highest scores for both LO and Assignments through study plan item ids
// The result would be returned array of topic scores and chapter scores
func (s *StudyPlanReaderService) StudentBookStudyProgress(ctx context.Context, req *pb.StudentBookStudyProgressRequest) (*pb.StudentBookStudyProgressResponse, error) {

	if err := s.validateStudentBookStudyProgressReq(ctx, req); err != nil {
		return nil, err
	}

	studyPlanItems, err := s.StudyPlanItemRepo.FetchByStudyProgressRequest(ctx, s.DB, database.Text(req.CourseId), database.Text(req.BookId), database.Text(req.StudentId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("StudyPlanItemRepo.FetchByStudyProgressRequest: %w", err).Error())
	}
	type ContentInfo struct {
		ID         string
		Percentage float32
		Completed  bool
	}

	assStudyPlanItemIDs := make([]string, 0)
	loStudyPlanItemIDs := make([]string, 0)
	topicContent := make(map[string][]ContentInfo)
	chapterContent := make(map[string][]string)

	for _, each := range studyPlanItems {
		contentStructure := new(entities.ContentStructure)
		if err := each.ContentStructure.AssignTo(contentStructure); err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("StudyPlanItemRepo.FetchByStudyProgressRequest: %w", err).Error())
		}

		contentInfo := ContentInfo{
			Percentage: -1,
		}
		if contentStructure.AssignmentID != "" {
			contentInfo.ID = each.ID.String
			if each.CompletedAt.Status == pgtype.Present {
				assStudyPlanItemIDs = append(assStudyPlanItemIDs, each.ID.String)
			}
		}
		if contentStructure.LoID != "" {

			contentInfo.ID = each.ID.String
			if each.CompletedAt.Status == pgtype.Present {
				loStudyPlanItemIDs = append(loStudyPlanItemIDs, each.ID.String)
			}
		}

		if each.CompletedAt.Status == pgtype.Present {
			contentInfo.Completed = true
		}

		topicContent[contentStructure.TopicID] = append(topicContent[contentStructure.TopicID], contentInfo)
		chapterContent[contentStructure.ChapterID] = append(chapterContent[contentStructure.ChapterID], contentStructure.TopicID)
	}

	for chapterID := range chapterContent {
		chapterContent[chapterID] = golibs.GetUniqueElementStringArray(chapterContent[chapterID])
	}

	if len(assStudyPlanItemIDs) > 0 {
		assignments, err := s.AssignmentRepo.CalculateHigestScore(ctx, s.DB, database.TextArray(assStudyPlanItemIDs))
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("AssignmentRepo.CalculateHigestScore: %w", err).Error())
		}

		taskAssignments, err := s.AssignmentRepo.CalculateTaskAssignmentHighestScore(ctx, s.DB, database.TextArray(assStudyPlanItemIDs))
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("AssignmentRepo.CalculateTaskAssignmentHighestScore: %w", err).Error())
		}

		for i, contents := range topicContent {
			for j, content := range contents {
				// non task assignment
				for _, assignment := range assignments {
					if content.ID == assignment.StudyPlanItemID.String && assignment.Percentage.Float >= 0 {
						topicContent[i][j].Percentage = assignment.Percentage.Float
					}
				}
				// task assignment
				for _, taskAssignment := range taskAssignments {
					if content.ID == taskAssignment.StudyPlanItemID.String && taskAssignment.Percentage.Float >= 0 {
						topicContent[i][j].Percentage = taskAssignment.Percentage.Float
					}
				}
			}
		}
	}

	if len(loStudyPlanItemIDs) > 0 {
		cctx, err := interceptors.GetOutgoingContext(ctx)

		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("StudyPlanReaderService.GetOutgoingContext: %w", err).Error())
		}
		loHighestScores, err := s.GetLOHighestScoresByStudyPlanItemIDs(cctx, &pb.GetLOHighestScoresByStudyPlanItemIDsRequest{
			StudyPlanItemIds: loStudyPlanItemIDs,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("StudyPlanReaderService.GetLOHighestScoresByStudyPlanItemIDs: %w", err).Error())
		}

		for i, contents := range topicContent {
			for j, content := range contents {
				for _, loHighestScore := range loHighestScores.GetLoHighestScores() {
					if content.ID == loHighestScore.StudyPlanItemId && loHighestScore.Percentage >= 0 {
						topicContent[i][j].Percentage = loHighestScore.Percentage
					}
				}
			}
		}
	}

	cctx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("StudyPlanReaderService.GetOutgoingContext: %w", err).Error())
	}

	// get topics that have not been deleted
	//TODO: ??? Why we need use bob client here ?
	topicIds := make([]string, 0)
	if resp, err := s.InternalReaderService.RetrieveTopics(cctx, &bpb.RetrieveTopicsRequest{
		BookIds: []string{req.BookId},
	}); err == nil {
		for _, topic := range resp.Items {
			topicIds = append(topicIds, topic.Info.Id)
		}
	}

	res := new(pb.StudentBookStudyProgressResponse)
	for topic, each := range topicContent {
		var numberCompleted, numLOCompleted int32 = 0, 0
		var topicScore float32 = 0
		for _, content := range each {
			if content.Percentage >= 0 {
				topicScore += content.Percentage
				numLOCompleted++
			}

			if content.Completed {
				numberCompleted++
			}
		}

		item := &pb.StudentTopicStudyProgress{
			TopicId:                topic,
			CompletedStudyPlanItem: wrapperspb.Int32(numberCompleted),
			TotalStudyPlanItem:     wrapperspb.Int32(int32(len(each))),
		}

		if topicScore >= 0 && numLOCompleted > 0 {
			item.AverageScore = wrapperspb.Int32(int32(float64(topicScore)/float64(numLOCompleted) + 0.5))
		} else {
			item.AverageScore = nil
		}

		res.TopicProgress = append(res.TopicProgress, item)
	}

	for chapter, each := range chapterContent {
		var chapterScore, numberCompleted int32 = 0, 0
		for _, topic := range res.TopicProgress {
			if containsStr(each, topic.TopicId) && containsStr(topicIds, topic.TopicId) && topic.AverageScore != nil && topic.AverageScore.Value >= 0 {
				chapterScore += topic.AverageScore.Value
				numberCompleted++
			}
		}

		item := &pb.StudentChapterStudyProgress{
			ChapterId: chapter,
		}

		if chapterScore >= 0 && numberCompleted > 0 {
			item.AverageScore = wrapperspb.Int32(int32(float64(chapterScore)/float64(numberCompleted) + 0.5))
		} else {
			item.AverageScore = nil
		}

		res.ChapterProgress = append(res.ChapterProgress, item)
	}
	return res, nil
}

func containsStr(s []string, target string) bool {
	for _, val := range s {
		if target == val {
			return true
		}
	}
	return false
}

func (s *StudyPlanReaderService) RetrieveStudyPlanItemEventLogs(ctx context.Context, req *pb.RetrieveStudyPlanItemEventLogsRequest) (*pb.RetrieveStudyPlanItemEventLogsResponse, error) {
	resp := &pb.RetrieveStudyPlanItemEventLogsResponse{}
	if len(req.StudyPlanItemId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "study_plan_item_id is empty")
	}

	studentEventLogs, err := s.StudentEventLogRepo.RetrieveStudentEventLogsByStudyPlanItemIDs(ctx, s.DB, database.TextArray(req.StudyPlanItemId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.StudentEventLogRepo.RetrieveStudentEventLogsByStudyPlanItemIDs: %v", err).Error())
	}

	var studyPlanItemIDs []string
	// key studyPlanItemID
	mStudyPlanItemIDAndLogs := make(map[string][]*entities.StudentEventLog)
	// key student_id and session_id
	// group logs by student_id and session_id to calculate learningTime and completedAt
	mLogs := make(map[string]map[string][]*entities.StudentEventLog)

	for _, studentEventLog := range studentEventLogs {
		payload := make(map[string]interface{})
		if err := studentEventLog.Payload.AssignTo(&payload); err != nil {
			return nil, err
		}

		studyPlanItemID, ok := payload["study_plan_item_id"].(string)
		if !ok {
			continue
		}
		sessionID, ok := payload["session_id"].(string)
		if !ok {
			continue
		}
		if _, ok := payload["event"].(string); !ok {
			continue
		}

		studentID := studentEventLog.StudentID.String

		if _, ok := mStudyPlanItemIDAndLogs[studyPlanItemID]; !ok {
			studyPlanItemIDs = append(studyPlanItemIDs, studyPlanItemID)
		}
		mStudyPlanItemIDAndLogs[studyPlanItemID] = append(mStudyPlanItemIDAndLogs[studyPlanItemID], studentEventLog)

		if _, ok := mLogs[studentID]; !ok {
			mLogs[studentID] = make(map[string][]*entities.StudentEventLog)
		}
		mLogs[studentID][sessionID] = append(mLogs[studentID][sessionID], studentEventLog)
	}

	// calculate learningTime and completedAt
	if err := s.calculateEventLogTime(mLogs); err != nil {
		return nil, err
	}

	// map event logs response
	if err := s.mapEventLogsResponse(studyPlanItemIDs, mStudyPlanItemIDAndLogs, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *StudyPlanReaderService) calculateEventLogTime(mLogs map[string]map[string][]*entities.StudentEventLog) error {
	for _, mLogsByStudentID := range mLogs {
		for _, logsBySessionID := range mLogsByStudentID {
			learningTime, completedAt, err := s.LearningTimeCalculator.Calculate(logsBySessionID)
			if err != nil {
				return err
			}

			for _, log := range logsBySessionID {
				payload := make(map[string]interface{})
				if err := log.Payload.AssignTo(&payload); err != nil {
					return err
				}

				payload["learning_time"] = learningTime.Seconds()
				if payload["event"].(string) == services.LOEventCompleted {
					if completedAt != nil { // avoid panic
						payload["completed_time"] = completedAt.Format(time.RFC3339)
					}
				}

				log.Payload.Set(payload)
			}
		}
	}
	return nil
}

func (s *StudyPlanReaderService) mapEventLogsResponse(
	studyPlanItemIDs []string, mStudyPlanItemIDAndLogs map[string][]*entities.StudentEventLog,
	resp *pb.RetrieveStudyPlanItemEventLogsResponse,
) error {
	for _, studyPlanItemID := range studyPlanItemIDs {
		logs := mStudyPlanItemIDAndLogs[studyPlanItemID]

		studyPlanItemLog := &pb.RetrieveStudyPlanItemEventLogsResponse_StudyPlanItemLog{
			StudyPlanItemId: studyPlanItemID,
		}

		for _, log := range logs {
			payload := make(map[string]interface{})
			if err := log.Payload.AssignTo(&payload); err != nil {
				return err
			}

			respLog := &pb.RetrieveStudyPlanItemEventLogsResponse_Log{
				SessionId:    payload["session_id"].(string),
				LearningTime: int32(payload["learning_time"].(float64)),
				CreatedAt:    timestamppb.New(log.CreatedAt.Time),
			}

			if payload["event"].(string) == services.LOEventCompleted {
				if payload["completed_time"] != nil {
					completedTime, err := time.Parse(time.RFC3339, payload["completed_time"].(string))
					if err != nil {
						return err
					}
					respLog.CompletedAt = timestamppb.New(completedTime)
				}
			}

			studyPlanItemLog.Logs = append(studyPlanItemLog.Logs, respLog)
		}

		resp.Items = append(resp.Items, studyPlanItemLog)
	}
	return nil
}

func (s *StudyPlanReaderService) GetLOHighestScoresByStudyPlanItemIDs(ctx context.Context, req *pb.GetLOHighestScoresByStudyPlanItemIDsRequest) (*pb.GetLOHighestScoresByStudyPlanItemIDsResponse, error) {
	shuflledQuizSets, err := s.ShuffledQuizSetRepo.CalculateHighestSubmissionScore(ctx, s.DB, database.TextArray(req.StudyPlanItemIds))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "StudyPlanReaderService.GetLOHighestScoresByStudyPlanItemIDs.CalculateHigestSubmissionScore: %v", err)
	}

	resp := &pb.GetLOHighestScoresByStudyPlanItemIDsResponse{}
	for _, each := range shuflledQuizSets {
		resp.LoHighestScores = append(resp.LoHighestScores, &pb.GetLOHighestScoresByStudyPlanItemIDsResponse_LOHighestScore{
			StudyPlanItemId: each.StudyPlanItemID.String,
			Percentage:      each.Percentage.Float,
		})
	}
	return resp, nil
}

func getAchievementCrown(score float32) pb.AchievementCrown {
	switch {
	case score == 100:
		return pb.AchievementCrown_ACHIEVEMENT_CROWN_GOLD
	case score >= 80:
		return pb.AchievementCrown_ACHIEVEMENT_CROWN_SILVER
	case score >= 60:
		return pb.AchievementCrown_ACHIEVEMENT_CROWN_BRONZE
	default:
		return pb.AchievementCrown_ACHIEVEMENT_CROWN_NONE
	}
}

func (s *StudyPlanReaderService) RetrieveStat(ctx context.Context, req *pb.RetrieveStatRequest) (*pb.RetrieveStatResponse, error) {
	var (
		totalLearningTime int32
		finishedLOs       []*entities.StudentsLearningObjectivesCompleteness
	)

	learningTimeByDailies, err := s.StudentLearningTimeDailyRepo.Retrieve(ctx, s.DB, database.Text(req.StudentId), nil, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.StudentLearningTimeDailyRepo.Retrieve: %w", err).Error())
	}
	for _, d := range learningTimeByDailies {
		totalLearningTime += d.LearningTime.Int
	}

	finishedLOs, err = s.StudentLOCompletenessRepo.RetrieveFinishedLOs(ctx, s.DB, database.Text(req.StudentId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.StudentLOCompletenessRepo.RetrieveFinishedLOs: %w", err).Error())
	}

	crownsMap := make(map[string]int32)
	// if student finishes any unassigned learning objective,
	// count that unassigned learning objective to the total lo.
	for _, f := range finishedLOs {
		if c := getAchievementCrown(f.HighestQuizScore.Float); c != pb.AchievementCrown_ACHIEVEMENT_CROWN_NONE {
			crownsMap[c.String()]++
		}
	}

	crowns := make([]*pb.StudentStatCrown, 0, len(crownsMap))
	for achievementCrown, total := range crownsMap {
		crowns = append(crowns, &pb.StudentStatCrown{
			AchievementCrown: achievementCrown,
			Total:            total,
		})
	}

	return &pb.RetrieveStatResponse{
		StudentStat: &pb.StudentStat{
			TotalLearningTime: totalLearningTime,
			TotalLoFinished:   int32(len(finishedLOs)),
			Crowns:            crowns,
		},
	}, nil
}

func (s *StudyPlanReaderService) RetrieveStatV2(ctx context.Context, req *pb.RetrieveStatRequest) (*pb.RetrieveStatResponse, error) {
	var (
		totalLearningTime int32
		finishedLOs       []*entities.StudentsLearningObjectivesCompleteness
	)

	learningTimeByDailies, err := s.StudentLearningTimeDailyRepo.RetrieveV2(ctx, s.DB, database.Text(req.StudentId), nil, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.StudentLearningTimeDailyRepo.RetrieveV2: %w", err).Error())
	}
	for _, d := range learningTimeByDailies {
		if d.LearningTime.Int > 0 {
			totalLearningTime += d.LearningTime.Int
		}
	}

	finishedLOs, err = s.StudentLOCompletenessRepo.RetrieveFinishedLOs(ctx, s.DB, database.Text(req.StudentId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.StudentLOCompletenessRepo.RetrieveFinishedLOs: %w", err).Error())
	}

	crownsMap := make(map[string]int32)
	// if student finishes any unassigned learning objective,
	// count that unassigned learning objective to the total lo.
	for _, f := range finishedLOs {
		if c := getAchievementCrown(f.HighestQuizScore.Float); c != pb.AchievementCrown_ACHIEVEMENT_CROWN_NONE {
			crownsMap[c.String()]++
		}
	}

	crowns := make([]*pb.StudentStatCrown, 0, len(crownsMap))
	for achievementCrown, total := range crownsMap {
		crowns = append(crowns, &pb.StudentStatCrown{
			AchievementCrown: achievementCrown,
			Total:            total,
		})
	}

	return &pb.RetrieveStatResponse{
		StudentStat: &pb.StudentStat{
			TotalLearningTime: totalLearningTime,
			TotalLoFinished:   int32(len(finishedLOs)),
			Crowns:            crowns,
		},
	}, nil
}

func (s *StudyPlanReaderService) GetStudentStudyPlan(ctx context.Context, req *pb.GetStudentStudyPlanRequest) (*pb.GetStudentStudyPlanResponse, error) {
	studyPlanIDs := req.GetStudyPlanIds()
	studentIDs := req.GetStudentIds()
	learningMaterialIDs := req.GetLearningMaterialIds()

	studentStudyPlans, err := s.StudentStudyPlanRepo.GetByStudyPlanStudentAndLO(ctx, s.DB, database.TextArray(studyPlanIDs), database.TextArray(studentIDs), database.TextArray(learningMaterialIDs))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("StudentStudyPlanRepo.GetByStudyPlanStudentAndLO: %w", err).Error())
	}
	studentStudyPlanRes := []*pb.GetStudentStudyPlanResponse_StudentStudyPlan{}
	for _, row := range studentStudyPlans {
		tmp := pb.GetStudentStudyPlanResponse_StudentStudyPlan{
			StudyPlanId: row.StudyPlanID.String,
			StudentId:   row.StudentID.String,
		}
		studentStudyPlanRes = append(studentStudyPlanRes, &tmp)
	}
	result := &pb.GetStudentStudyPlanResponse{
		StudentStudyPlans: studentStudyPlanRes,
	}

	return result, nil
}
