package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	econs "github.com/manabie-com/backend/internal/eureka/constants"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type BobCourseClientServiceClient interface {
	RetrieveCoursesByIDs(ctx context.Context, in *bob_pb.RetrieveCoursesByIDsRequest, opts ...grpc.CallOption) (*bob_pb.RetrieveCoursesResponse, error)
}

type StatisticService struct {
	sspb.UnimplementedStatisticsServer
	DB database.Ext

	UserMgmtService             UserMgmtService
	StudentLearningTimeDaiyRepo interface {
		RetrieveV2(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, queryEnhancers ...repositories.QueryEnhancer) ([]*repositories.StudentLearningTimeDailyV2, error)
	}

	StudentLOCompletenessRepo interface {
		RetrieveFinishedLOs(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]*entities.StudentsLearningObjectivesCompleteness, error)
	}
	StudentLearningTimeDailyRepo interface {
		RetrieveV2(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, queryEnhancers ...repositories.QueryEnhancer) ([]*entities.StudentLearningTimeDaily, error)
	}
	CourseClient interface {
		RetrieveCoursesByIDs(ctx context.Context, in *bob_pb.RetrieveCoursesByIDsRequest, opts ...grpc.CallOption) (*bob_pb.RetrieveCoursesResponse, error)
	}
	StudentReaderClient BobStudentReaderServiceClient

	ExamLORepo interface {
		GetScores(ctx context.Context, db database.QueryExecer, courseIDs, studyPlanIDs, studentIDs pgtype.TextArray, getGradeToPassScore pgtype.Bool) ([]*entities.ExamLoScore, error)
	}

	LearningMaterialRepo interface {
		FindInfoByStudyPlanItemIdentity(ctx context.Context, db database.QueryExecer, studentID, studyPlanID pgtype.Text, learningMaterialID pgtype.Text) ([]*repositories.LearningMaterialInfo, error)
	}

	StatisticsRepo interface {
		GetStudentProgress(ctx context.Context, db database.QueryExecer, studentID, studyPlanID, courseID pgtype.Text) ([]*repositories.LearningMaterialProgress, []*repositories.StudentTopicProgress, []*repositories.StudentChapterProgress, error)
		GetStudentTopicProgress(ctx context.Context, db database.QueryExecer, studentID, studyPlanID pgtype.Text) ([]*repositories.StudentTopicProgress, error)
		GetStudentChapterProgress(ctx context.Context, db database.QueryExecer, studentID, studyPlanID pgtype.Text) ([]*repositories.StudentChapterProgress, error)
	}

	StudentRepo interface {
		FilterByGradeBookView(ctx context.Context, db database.QueryExecer,
			studentIDs,
			studyPlanIDs pgtype.TextArray,
			courseIDs pgtype.TextArray,
			grades pgtype.Int4Array,
			gradeIds pgtype.TextArray,
			studentName pgtype.Text,
			locationIDs pgtype.TextArray,
			limit,
			offset int64,
		) ([]*repositories.StudentInfo, error)

		FilterOutDeletedStudentIDs(
			ctx context.Context,
			db database.QueryExecer,
			studentIDs []string,
		) ([]string, error)
	}

	StudentSubmissionRepo interface {
		RetrieveByStudyPlanIdentities(context.Context, database.QueryExecer, []*repositories.StudyPlanItemIdentity) ([]*repositories.StudentSubmissionInfo, error)
	}

	CourseStudyPlanRepo interface {
		ListCourseStatisticV3(ctx context.Context, db database.QueryExecer, args *repositories.ListCourseStatisticItemsArgsV3) ([]*repositories.TopicStatistic, []*repositories.LearningMaterialStatistic, error)
		ListCourseStatisticV4(ctx context.Context, db database.QueryExecer, args *repositories.ListCourseStatisticItemsArgsV3) ([]*repositories.TopicStatistic, []*repositories.LearningMaterialStatistic, error)
	}

	CourseStudentRepo interface {
		FindStudentByCourseID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) ([]string, error)
		FindStudentTagByCourseID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) ([]*entities.StudentTag, error)
	}

	ClassStudentRepo interface {
		GetClassStudentByCourseAndClassIds(ctx context.Context, db database.QueryExecer, courseIDs, classIDs pgtype.TextArray) ([]*entities.ClassStudent, error)
		GetClassStudentByCourse(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray) ([]*entities.ClassStudent, error)
	}

	CourseStudentAccessPathRepo interface {
		GetByLocationsStudentsAndCourse(ctx context.Context, db database.QueryExecer, locationIDs, studentIDs, courseIDs pgtype.TextArray) ([]*entities.CourseStudentsAccessPath, error)
	}
}

func NewStatisticService(db database.Ext, courseClient bob_pb.CourseClient, studentReaderClient bpb.StudentReaderServiceClient) sspb.StatisticsServer {
	return &StatisticService{
		DB:                          db,
		CourseClient:                courseClient,
		StudentReaderClient:         studentReaderClient,
		ExamLORepo:                  &repositories.ExamLORepo{},
		StudentRepo:                 &repositories.StudentRepo{},
		StudentSubmissionRepo:       &repositories.StudentSubmissionRepo{},
		StatisticsRepo:              &repositories.StatisticsRepo{},
		LearningMaterialRepo:        &repositories.LearningMaterialRepo{},
		CourseStudyPlanRepo:         &repositories.CourseStudyPlanRepo{},
		CourseStudentRepo:           &repositories.CourseStudentRepo{},
		ClassStudentRepo:            &repositories.ClassStudentRepo{},
		CourseStudentAccessPathRepo: &repositories.CourseStudentAccessPathRepo{},
	}
}

func (s *StatisticService) ListGradeBook(ctx context.Context, req *sspb.GradeBookRequest) (*sspb.GradeBookResponse, error) {
	limit, offset := getLimitOffset(req)
	studentInfos, err := s.StudentRepo.FilterByGradeBookView(
		ctx,
		s.DB,
		database.TextArray(req.GetStudentIds()),
		database.TextArray(req.GetStudyPlanIds()),
		database.TextArray(req.GetCourseIds()),
		database.Int4Array(req.GetGrades()),
		database.TextArray(req.GetGradeIds()),
		database.Text(req.GetStudentName()),
		database.TextArray(req.GetLocationIds()),
		limit.Int,
		offset.Int,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.StudentRepo.FilterByGradeBookView: %w", err).Error())
	}

	studentIds := make([]string, 0)
	courseIds := pgtype.TextArray{
		Status: pgtype.Present,
		Dimensions: []pgtype.ArrayDimension{
			{
				Length:     0,
				LowerBound: 1,
			},
		},
	}
	mapStudentGrade := make(map[string]*wrapperspb.Int32Value)
	getGradeToPassScore := pgtype.Bool{Status: pgtype.Null}
	for _, info := range studentInfos {
		studentIds = append(studentIds, info.StudentID.String)
		courseIds.Elements = append(courseIds.Elements, info.CourseIDs.Elements...)
		courseIds.Dimensions[0].Length += int32(len(info.CourseIDs.Elements))
		if _, ok := mapStudentGrade[info.StudentID.String]; !ok {
			mapStudentGrade[info.StudentID.String] = &wrapperspb.Int32Value{
				Value: int32(info.Grade.Int),
			}
		}
	}

	if req.Setting.Enum() != nil && sspb.GradeBookSetting(req.Setting.Number()) == sspb.GradeBookSetting_GRADE_TO_PASS_SCORE {
		getGradeToPassScore.Set(true)
	}

	examLoScores, err := s.ExamLORepo.GetScores(
		ctx,
		s.DB,
		courseIds,
		database.TextArray(req.GetStudyPlanIds()),
		database.TextArray(studentIds),
		getGradeToPassScore,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ExamLORepo.GetScores: %w", err).Error())
	}

	courseIDs := make([]string, 0)
	mapStudentStudyPlan := make(map[entities.StudentStudyPlanIdentity]*sspb.GradeBookResponse_StudentGradeItem)
	for _, examLoScore := range examLoScores {
		key := entities.StudentStudyPlanIdentity{
			StudentID:   examLoScore.StudentID.String,
			CourseID:    examLoScore.CourseID.String,
			StudyPlanID: examLoScore.StudyPlanID.String,
		}

		if _, ok := mapStudentStudyPlan[key]; !ok {
			studentGradeItem := &sspb.GradeBookResponse_StudentGradeItem{
				StudentId:   key.StudentID,
				StudentName: examLoScore.StudentName.String,
				Grade: &wrapperspb.UInt32Value{
					Value: uint32(examLoScore.Grade.Int),
				},
				GradeId:       examLoScore.GradeID.String,
				StudyPlanId:   key.StudyPlanID,
				StudyPlanName: examLoScore.StudyPlanName.String,
				CourseId:      key.CourseID,
				TotalExamLos: &wrapperspb.UInt32Value{
					Value: uint32(examLoScore.TotalExamLOs.Int),
				},
				TotalCompletedExamLos: &wrapperspb.UInt32Value{
					Value: uint32(examLoScore.TotalCompletedExamLOs.Int),
				},
				TotalGradeToPass: &wrapperspb.UInt32Value{
					Value: uint32(examLoScore.TotalGradeToPass.Int),
				},
				TotalPassed: &wrapperspb.UInt32Value{
					Value: uint32(examLoScore.TotalPassed.Int),
				},
				Results: []*sspb.GradeBookResponse_ExamResult{},
			}
			mapStudentStudyPlan[key] = studentGradeItem
		}

		mapStudentStudyPlan[key].Results = append(mapStudentStudyPlan[key].Results, &sspb.GradeBookResponse_ExamResult{
			LmId:          examLoScore.LearningMaterialID.String,
			TotalAttempts: uint32(examLoScore.TotalAttempts.Int),
			GradePoint: &wrapperspb.UInt32Value{
				Value: uint32(examLoScore.GradePoint.Int),
			},
			Failed: !examLoScore.PassedExamLo.Bool,
			TotalPoint: &wrapperspb.UInt32Value{
				Value: uint32(examLoScore.TotalPoint.Int),
			},
			LmName:              examLoScore.ExamLOName.String,
			Status:              sspb.SubmissionStatus(sspb.SubmissionStatus_value[examLoScore.Status.String]),
			IsGradeToPass:       examLoScore.IsGradeToPass.Bool,
			ReviewOption:        sspb.ExamLOReviewOption(sspb.ExamLOReviewOption_value[examLoScore.ReviewOption.String]),
			DueDate:             timestamppb.New(examLoScore.DueDate.Time),
			ChapterDisplayOrder: int32(examLoScore.ChapterDisplayOrder.Int),
			TopicDisplayOrder:   int32(examLoScore.TopicDisplayOrder.Int),
			LmDisplayOrder:      int32(examLoScore.LmDisplayOrder.Int),
		})

		courseIDs = append(courseIDs, examLoScore.CourseID.String)
	}

	mdCtx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}
	courses, err := s.CourseClient.RetrieveCoursesByIDs(mdCtx, &bob_pb.RetrieveCoursesByIDsRequest{
		Ids: golibs.GetUniqueElementStringArray(courseIDs),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.CourseClient.RetrieveCoursesByIDs: %w", err).Error())
	}

	mapCourseInfo := make(map[string]string, 0) // id - name
	for _, course := range courses.GetCourses() {
		if _, ok := mapCourseInfo[course.GetId()]; !ok {
			mapCourseInfo[course.GetId()] = course.Name
		}
	}

	studentGradeItems := make([]*sspb.GradeBookResponse_StudentGradeItem, 0)
	for key, item := range mapStudentStudyPlan {
		item.CourseName = mapCourseInfo[key.CourseID]
		studentGradeItems = append(studentGradeItems, item)
	}

	getInfoForSortStudent := func(item *sspb.GradeBookResponse_StudentGradeItem) (string, string, string) {
		return item.StudentName, item.CourseName, item.StudyPlanName
	}

	getInfoForSortBookTree := func(item *sspb.GradeBookResponse_ExamResult) (int, int, int) {
		return int(item.ChapterDisplayOrder), int(item.TopicDisplayOrder), int(item.LmDisplayOrder)
	}

	sort.SliceStable(studentGradeItems, func(i, j int) bool {
		studentName1, courseName1, studyPlanName1 := getInfoForSortStudent(studentGradeItems[i])
		studentName2, courseName2, studyPlanName2 := getInfoForSortStudent(studentGradeItems[j])
		if studentName1 != studentName2 {
			return studentName1 < studentName2
		}
		if courseName1 != courseName2 {
			return courseName1 < courseName2
		}
		return studyPlanName1 < studyPlanName2
	})

	for _, item := range studentGradeItems {
		sort.SliceStable(item.Results, func(i, j int) bool {
			chapterDO1, topicDO1, lmDO1 := getInfoForSortBookTree(item.Results[i])
			chapterDO2, topicDO2, lmDO2 := getInfoForSortBookTree(item.Results[j])
			if chapterDO1 != chapterDO2 {
				return chapterDO1 < chapterDO2
			}
			if topicDO1 != topicDO2 {
				return topicDO1 < topicDO2
			}
			return lmDO1 < lmDO2
		})
	}

	return &sspb.GradeBookResponse{
		StudentGradeItems: studentGradeItems,
		NextPage:          getNextPaging(limit, offset),
	}, nil
}

func getLimitOffset(req *sspb.GradeBookRequest) (limit, offset pgtype.Int8) {
	limit = database.Int8(constant.PageLimit)
	offset = database.Int8(0)

	if req.Paging != nil && req.Paging.Limit != 0 {
		_ = limit.Set(req.Paging.Limit)
		_ = offset.Set(req.Paging.GetOffsetInteger())
	}

	return limit, offset
}

func getNextPaging(limit, offset pgtype.Int8) *cpb.Paging {
	return &cpb.Paging{
		Limit: uint32(limit.Int),
		Offset: &cpb.Paging_OffsetInteger{
			OffsetInteger: limit.Int + offset.Int,
		},
	}
}

func studentTopicProgressToPb(tp *repositories.StudentTopicProgress) *sspb.StudentTopicStudyProgress {
	var completedSPItems *wrapperspb.Int32Value
	if tp.CompletedSPItems.Status == pgtype.Present {
		completedSPItems = wrapperspb.Int32(int32(tp.CompletedSPItems.Int))
	}
	var totalSPItems *wrapperspb.Int32Value
	if tp.TotalSpItems.Status == pgtype.Present {
		totalSPItems = wrapperspb.Int32(int32(tp.TotalSpItems.Int))
	}
	var averageScore *wrapperspb.Int32Value
	if tp.AverageScore.Status == pgtype.Present {
		averageScore = wrapperspb.Int32(int32(tp.AverageScore.Int))
	}
	return &sspb.StudentTopicStudyProgress{
		TopicId:                tp.TopicID.String,
		CompletedStudyPlanItem: completedSPItems,
		TotalStudyPlanItem:     totalSPItems,
		AverageScore:           averageScore,
		TopicName:              tp.TopicName.String,
	}
}

func studentChapterProgressToPb(cp *repositories.StudentChapterProgress) *sspb.StudentChapterStudyProgress {
	return &sspb.StudentChapterStudyProgress{
		ChapterId:    cp.ChapterID.String,
		AverageScore: wrapperspb.Int32(int32(cp.AverageScore.Int)),
	}
}

func learningMaterialResultToPb(info *repositories.LearningMaterialProgress) *sspb.LearningMaterialResult {
	return &sspb.LearningMaterialResult{
		LearningMaterial: &sspb.LearningMaterialBase{
			LearningMaterialId: info.LearningMaterialID.String,
			TopicId:            info.TopicID.String,
			Type:               info.Type.String,
			Name:               info.Name.String,
			DisplayOrder:       wrapperspb.Int32(int32(info.LmDisplayOrder.Int)),
		},
		IsCompleted: info.IsCompleted.Bool,
		Crown:       getAchievementCrownV2(float32(info.HighestScore.Int)),
	}
}

func studyPlanTreeResultToPb(info *repositories.LearningMaterialProgress) *sspb.StudyPlanTree {
	var availableFrom *timestamppb.Timestamp
	if info.AvailableFrom.Status == pgtype.Present {
		availableFrom = timestamppb.New(info.AvailableFrom.Time)
	}
	var availableTo *timestamppb.Timestamp
	if info.AvailableTo.Status == pgtype.Present {
		availableTo = timestamppb.New(info.AvailableTo.Time)
	}
	var startDate *timestamppb.Timestamp
	if info.StartDate.Status == pgtype.Present {
		startDate = timestamppb.New(info.StartDate.Time)
	}
	var endDate *timestamppb.Timestamp
	if info.EndDate.Status == pgtype.Present {
		endDate = timestamppb.New(info.EndDate.Time)
	}
	var completedAt *timestamppb.Timestamp
	if info.CompletedAt.Status == pgtype.Present {
		completedAt = timestamppb.New(info.CompletedAt.Time)
	}
	var schoolDate *timestamppb.Timestamp
	if info.SchoolDate.Status == pgtype.Present {
		schoolDate = timestamppb.New(info.SchoolDate.Time)
	}
	return &sspb.StudyPlanTree{
		StudyPlanId: info.StudyPlanID.String,
		BookTree: &sspb.BookTree{
			BookId:              info.BookID.String,
			ChapterId:           info.ChapterID.String,
			ChapterDisplayOrder: int32(info.ChapterDisplayOrder.Int),
			TopicId:             info.TopicID.String,
			TopicDisplayOrder:   int32(info.TopicDisplayOrder.Int),
			LearningMaterialId:  info.LearningMaterialID.String,
			LmDisplayOrder:      int32(info.LmDisplayOrder.Int),
			LmType:              getLearningMaterialType(info.Type.String),
		},
		AvailableFrom: availableFrom,
		AvailableTo:   availableTo,
		StartDate:     startDate,
		EndDate:       endDate,
		CompletedAt:   completedAt,
		Status:        getStudyPlanItemStatus(info.Status.String),
		SchoolDate:    schoolDate,
	}
}

func retrieveGetStudentProgressRequest(ctx context.Context, req *sspb.GetStudentProgressRequest) (studentID, studyPlanID, courseID pgtype.Text, err error) {
	if req.StudyPlanItemIdentity.StudentId == nil || req.StudyPlanItemIdentity.StudentId.Value == "" {
		userID := interceptors.UserIDFromContext(ctx)
		studentID = database.Text(userID)
	} else {
		studentID = database.Text(req.StudyPlanItemIdentity.StudentId.Value)
	}
	studyPlanID = pgtype.Text{Status: pgtype.Null}
	if req.StudyPlanItemIdentity.StudyPlanId != "" {
		studyPlanID = database.Text(req.StudyPlanItemIdentity.StudyPlanId)
	}
	if req.CourseId == "" {
		return studentID, studyPlanID, courseID, fmt.Errorf("course_id is required")
	}
	courseID = database.Text(req.CourseId)
	return studentID, studyPlanID, courseID, nil
}

func (s *StatisticService) GetStudentProgress(ctx context.Context, req *sspb.GetStudentProgressRequest) (*sspb.GetStudentProgressResponse, error) {
	studentID, studyPlanID, courseID, err := retrieveGetStudentProgressRequest(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	topicProgress := make([]*repositories.StudentTopicProgress, 0)
	chapterProgress := make([]*repositories.StudentChapterProgress, 0)
	lmProgressInfo := make([]*repositories.LearningMaterialProgress, 0)

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		lmProgressInfo, topicProgress, chapterProgress, err = s.StatisticsRepo.GetStudentProgress(ctx, tx, studentID, studyPlanID, courseID)
		if err != nil {
			return fmt.Errorf("StatisticsRepo.GetStudentTopicProgress: %w", err)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	studyPlanIDs := make([]string, 0, len(chapterProgress))
	for _, cp := range chapterProgress {
		studyPlanIDs = append(studyPlanIDs, cp.StudyPlanID.String)
	}
	uniqueStudyPlanIDs := golibs.GetUniqueElementStringArray(studyPlanIDs)

	studentStudyPlanProgresses := make([]*sspb.GetStudentProgressResponse_StudentStudyPlanProgress, 0, len(uniqueStudyPlanIDs))
	for _, studyPlanID := range uniqueStudyPlanIDs {
		topicProgressPb := make([]*sspb.StudentTopicStudyProgress, 0)
		for _, tp := range topicProgress {
			if tp.StudyPlanID.String == studyPlanID {
				topicProgressPb = append(topicProgressPb, studentTopicProgressToPb(tp))
			}
		}

		chapterProgressPb := make([]*sspb.StudentChapterStudyProgress, 0)
		for _, cp := range chapterProgress {
			if cp.StudyPlanID.String == studyPlanID {
				chapterProgressPb = append(chapterProgressPb, studentChapterProgressToPb(cp))
			}
		}

		lmResults := make([]*sspb.LearningMaterialResult, 0)
		studyPlanTrees := make([]*sspb.StudyPlanTree, 0)
		for _, info := range lmProgressInfo {
			if info.StudyPlanID.String == studyPlanID {
				lmResults = append(lmResults, learningMaterialResultToPb(info))
				studyPlanTrees = append(studyPlanTrees, studyPlanTreeResultToPb(info))
			}
		}

		studentStudyPlanProgresses = append(studentStudyPlanProgresses, &sspb.GetStudentProgressResponse_StudentStudyPlanProgress{
			StudyPlanId:             studyPlanID,
			TopicProgress:           topicProgressPb,
			ChapterProgress:         chapterProgressPb,
			LearningMaterialResults: lmResults,
			StudyPlanTrees:          studyPlanTrees,
		})
	}

	return &sspb.GetStudentProgressResponse{
		StudentStudyPlanProgresses: studentStudyPlanProgresses,
	}, nil
}
func validateCourseStatisticRequest(req *sspb.CourseStatisticRequest) error {
	if req.GetCourseId() == "" {
		return errors.New("Missing course")
	}
	if req.GetStudyPlanId() == "" {
		return errors.New("Missing study plan")
	}
	return nil
}

func (s *StatisticService) RetrieveCourseStatisticV2(ctx context.Context, req *sspb.CourseStatisticRequest) (*sspb.CourseStatisticResponse, error) {
	if err := validateCourseStatisticRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("validateCourseStatisticRequest %w", err).Error())
	}

	// Get all student_ids exists in the course
	studentIDs, err := s.CourseStudentRepo.FindStudentByCourseID(ctx, s.DB, database.Text(req.GetCourseId()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.CourseStudentRepo.FindStudentByCourseID %w", err).Error())
	}

	// Get student_ids by school_id, all_student in course
	if req.GetSchoolId() != "" || req.GetUnassigned() {
		cctx, err := interceptors.GetOutgoingContext(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := s.StudentReaderClient.RetrieveStudentSchoolHistory(cctx, &bpb.RetrieveStudentSchoolHistoryRequest{
			StudentIds: studentIDs,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("s.StudentReaderClient.RetrieveStudentSchoolHistory %w", err).Error())
		}

		switch req.GetSchool().(type) {
		case *sspb.CourseStatisticRequest_Unassigned:
			var assigned []string
			for _, school := range resp.GetSchools() {
				assigned = append(assigned, school.GetStudentIds()...)
			}
			studentIDs = sliceutils.Filter(studentIDs, func(studentID string) bool {
				return !sliceutils.Contains(assigned, studentID)
			})
		case *sspb.CourseStatisticRequest_SchoolId:
			studentIDs = resp.GetSchools()[req.GetSchoolId()].GetStudentIds()
		}
	}

	// Get student_ids by class or unassign in this course
	listClassIDs := req.GetClassId()
	if len(listClassIDs) > 0 {
		var classStudentIds []string
		courseIDsConvert := []string{req.GetCourseId()}

		indexOfUnAssignID := slices.Index(listClassIDs, econs.UnAssignClassID)
		if indexOfUnAssignID != -1 {
			listClassIDs = slices.Delete(listClassIDs, indexOfUnAssignID, indexOfUnAssignID+1)

			var allClassStudentIds []string
			var allUnAssignClassStudentIds []string
			allClassStudents, err := s.ClassStudentRepo.GetClassStudentByCourse(ctx, s.DB, database.TextArray(courseIDsConvert))
			if err != nil {
				return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ClassStudentRepo.GetClassStudentByCourse %w", err).Error())
			}
			for _, classStudent := range allClassStudents {
				allClassStudentIds = append(allClassStudentIds, classStudent.StudentID.String)
			}

			allStudentIDsInCourse, err := s.CourseStudentRepo.FindStudentByCourseID(ctx, s.DB, database.Text(req.GetCourseId()))
			if err != nil {
				return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ClassStudentRepo.FindStudentByCourseID %w", err).Error())
			}

			for _, studentID := range allStudentIDsInCourse {
				if !slices.Contains(allClassStudentIds, studentID) {
					allUnAssignClassStudentIds = append(allUnAssignClassStudentIds, studentID)
				}
			}
			classStudentIds = append(classStudentIds, allUnAssignClassStudentIds...)
		}

		if len(listClassIDs) != 0 {
			// get list students class of this course
			classStudents, err := s.ClassStudentRepo.GetClassStudentByCourseAndClassIds(ctx, s.DB, database.TextArray(courseIDsConvert), database.TextArray(listClassIDs))
			if err != nil {
				return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ClassStudentRepo.GetClassStudentByCourseAndClassIds %w", err).Error())
			}

			for _, classStudent := range classStudents {
				classStudentIds = append(classStudentIds, classStudent.StudentID.String)
			}
		}
		classStudentIds = sliceutils.RemoveDuplicates(classStudentIds)
		studentIDs = sliceutils.Intersect(studentIDs, classStudentIds)
	}

	// filter students by locations
	listLocationIDsFilter := req.GetLocationIds()
	if len(listLocationIDsFilter) > 0 {
		courseStudentAccessPaths, err := s.CourseStudentAccessPathRepo.GetByLocationsStudentsAndCourse(ctx, s.DB, database.TextArray(listLocationIDsFilter), database.TextArray(studentIDs), database.TextArray([]string{req.GetCourseId()}))
		if err != nil && err != pgx.ErrNoRows {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("s.CourseStudentAccessPathRepo.GetByLocationsStudentsAndCourse %w", err).Error())
		}
		tempStudentIDs := []string{}
		for _, courseStudentAccessPath := range courseStudentAccessPaths {
			tempStudentIDs = append(tempStudentIDs, courseStudentAccessPath.StudentID.String)
		}
		studentIDs = tempStudentIDs
	}

	validStudentIDs, err := s.StudentRepo.FilterOutDeletedStudentIDs(ctx, s.DB, studentIDs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.StudentRepo.FilterOutDeletedStudents %w", err).Error())
	}

	args := &repositories.ListCourseStatisticItemsArgsV3{
		CourseID:    database.Text(req.CourseId),
		StudyPlanID: database.Text(req.StudyPlanId),
		ClassID:     pgtype.TextArray{Status: pgtype.Null},
		StudentIDs:  database.TextArray(validStudentIDs),
		TagIDs:      pgtype.TextArray{Status: pgtype.Null},
		LocationIDs: pgtype.TextArray{Status: pgtype.Null},
	}

	if len(req.ClassId) != 0 {
		args.ClassID = database.TextArray(req.ClassId)
	}

	if tagIDs := req.GetStudentTagIds(); len(tagIDs) != 0 {
		args.TagIDs = database.TextArray(tagIDs)
	}

	if len(req.LocationIds) != 0 {
		args.LocationIDs = database.TextArray(req.LocationIds)
	}

	topics, lms, err := s.CourseStudyPlanRepo.ListCourseStatisticV4(ctx, s.DB, args)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	resp := &sspb.CourseStatisticResponse{}
	topicsStatistic := []*sspb.CourseStatisticResponse_TopicStatistic{}
	mapTopicLms := map[string][]*repositories.LearningMaterialStatistic{}
	for _, lm := range lms {
		if _, ok := mapTopicLms[lm.TopicID]; !ok || ok {
			mapTopicLms[lm.TopicID] = append(mapTopicLms[lm.TopicID], lm)
		}
	}

	for _, topic := range topics {
		topicpb := &sspb.CourseStatisticResponse_TopicStatistic{}
		topicpb.TopicId = topic.TopicID
		topicpb.AverageScore = topic.AverageScore.Int
		topicpb.CompletedStudent = topic.CompletedStudent.Int
		topicpb.TotalAssignedStudent = topic.TotalAssignStudent.Int
		lmspb := []*sspb.CourseStatisticResponse_TopicStatistic_LearningMaterialStatistic{}

		if _, ok := mapTopicLms[topic.TopicID]; !ok {
			return nil, status.Errorf(codes.Internal, "Topic not exist in LearningMaterialStatistic")
		}

		for _, lm := range mapTopicLms[topic.TopicID] {
			lmpb := &sspb.CourseStatisticResponse_TopicStatistic_LearningMaterialStatistic{}
			lmpb.LearningMaterialId = lm.LearningMaterialID
			lmpb.TotalAssignedStudent = lm.TotalAssignStudent.Int
			lmpb.CompletedStudent = lm.CompletedStudent.Int
			lmpb.AverageScore = lm.AverageScore.Int

			lmspb = append(lmspb, lmpb)
		}

		if len(lmspb) == 0 {
			return nil, status.Errorf(codes.Internal, "LearningMaterialStatistic len is 0")
		}

		topicpb.LearningMaterialStatistic = lmspb

		topicsStatistic = append(topicsStatistic, topicpb)
	}

	resp.TopicStatistic = topicsStatistic

	return resp, nil
}

func (s *StatisticService) RetrieveCourseStatistic(ctx context.Context, req *sspb.CourseStatisticRequest) (*sspb.CourseStatisticResponse, error) {
	if err := validateCourseStatisticRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("validateCourseStatisticRequest %w", err).Error())
	}

	// Get all student_ids exists in the course
	studentIDs, err := s.CourseStudentRepo.FindStudentByCourseID(ctx, s.DB, database.Text(req.GetCourseId()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.CourseStudentRepo.FindStudentByCourseID %w", err).Error())
	}

	// Get student_ids by school_id, all_student in course
	if req.GetSchoolId() != "" || req.GetUnassigned() {
		cctx, err := interceptors.GetOutgoingContext(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := s.StudentReaderClient.RetrieveStudentSchoolHistory(cctx, &bpb.RetrieveStudentSchoolHistoryRequest{
			StudentIds: studentIDs,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("s.StudentReaderClient.RetrieveStudentSchoolHistory %w", err).Error())
		}

		switch req.GetSchool().(type) {
		case *sspb.CourseStatisticRequest_Unassigned:
			var assigned []string
			for _, school := range resp.GetSchools() {
				assigned = append(assigned, school.GetStudentIds()...)
			}
			studentIDs = sliceutils.Filter(studentIDs, func(studentID string) bool {
				return !sliceutils.Contains(assigned, studentID)
			})
		case *sspb.CourseStatisticRequest_SchoolId:
			studentIDs = resp.GetSchools()[req.GetSchoolId()].GetStudentIds()
		}
	}

	// Get student_ids by class or unassign in this course
	listClassIDs := req.GetClassId()
	if len(listClassIDs) > 0 {
		var classStudentIds []string
		courseIDsConvert := []string{req.GetCourseId()}

		indexOfUnAssignID := slices.Index(listClassIDs, econs.UnAssignClassID)
		if indexOfUnAssignID != -1 {
			listClassIDs = slices.Delete(listClassIDs, indexOfUnAssignID, indexOfUnAssignID+1)

			var allClassStudentIds []string
			var allUnAssignClassStudentIds []string
			allClassStudents, err := s.ClassStudentRepo.GetClassStudentByCourse(ctx, s.DB, database.TextArray(courseIDsConvert))
			if err != nil {
				return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ClassStudentRepo.GetClassStudentByCourse %w", err).Error())
			}
			for _, classStudent := range allClassStudents {
				allClassStudentIds = append(allClassStudentIds, classStudent.StudentID.String)
			}

			allStudentIDsInCourse, err := s.CourseStudentRepo.FindStudentByCourseID(ctx, s.DB, database.Text(req.GetCourseId()))
			if err != nil {
				return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ClassStudentRepo.FindStudentByCourseID %w", err).Error())
			}

			for _, studentID := range allStudentIDsInCourse {
				if !slices.Contains(allClassStudentIds, studentID) {
					allUnAssignClassStudentIds = append(allUnAssignClassStudentIds, studentID)
				}
			}
			classStudentIds = append(classStudentIds, allUnAssignClassStudentIds...)
		}

		if len(listClassIDs) != 0 {
			// get list students class of this course
			classStudents, err := s.ClassStudentRepo.GetClassStudentByCourseAndClassIds(ctx, s.DB, database.TextArray(courseIDsConvert), database.TextArray(listClassIDs))
			if err != nil {
				return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ClassStudentRepo.GetClassStudentByCourseAndClassIds %w", err).Error())
			}

			for _, classStudent := range classStudents {
				classStudentIds = append(classStudentIds, classStudent.StudentID.String)
			}
		}
		classStudentIds = sliceutils.RemoveDuplicates(classStudentIds)
		studentIDs = sliceutils.Intersect(studentIDs, classStudentIds)
	}

	// filter students by locations
	listLocationIDsFilter := req.GetLocationIds()
	if len(listLocationIDsFilter) > 0 {
		courseStudentAccessPaths, err := s.CourseStudentAccessPathRepo.GetByLocationsStudentsAndCourse(ctx, s.DB, database.TextArray(listLocationIDsFilter), database.TextArray(studentIDs), database.TextArray([]string{req.GetCourseId()}))
		if err != nil && err != pgx.ErrNoRows {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("s.CourseStudentAccessPathRepo.GetByLocationsStudentsAndCourse %w", err).Error())
		}
		tempStudentIDs := []string{}
		for _, courseStudentAccessPath := range courseStudentAccessPaths {
			tempStudentIDs = append(tempStudentIDs, courseStudentAccessPath.StudentID.String)
		}
		studentIDs = tempStudentIDs
	}

	validStudentIDs, err := s.StudentRepo.FilterOutDeletedStudentIDs(ctx, s.DB, studentIDs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.StudentRepo.FilterOutDeletedStudents %w", err).Error())
	}

	args := &repositories.ListCourseStatisticItemsArgsV3{
		CourseID:    database.Text(req.CourseId),
		StudyPlanID: database.Text(req.StudyPlanId),
		ClassID:     pgtype.TextArray{Status: pgtype.Null},
		StudentIDs:  database.TextArray(validStudentIDs),
		TagIDs:      pgtype.TextArray{Status: pgtype.Null},
	}
	if len(req.ClassId) != 0 {
		args.ClassID = database.TextArray(req.ClassId)
	}
	if tagIDs := req.GetStudentTagIds(); len(tagIDs) != 0 {
		args.TagIDs = database.TextArray(tagIDs)
	}

	topics, lms, err := s.CourseStudyPlanRepo.ListCourseStatisticV3(ctx, s.DB, args)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	resp := &sspb.CourseStatisticResponse{}
	topicsStatistic := []*sspb.CourseStatisticResponse_TopicStatistic{}
	mapTopicLms := map[string][]*repositories.LearningMaterialStatistic{}
	for _, lm := range lms {
		if _, ok := mapTopicLms[lm.TopicID]; !ok || ok {
			mapTopicLms[lm.TopicID] = append(mapTopicLms[lm.TopicID], lm)
		}
	}

	for _, topic := range topics {
		topicpb := &sspb.CourseStatisticResponse_TopicStatistic{}
		topicpb.TopicId = topic.TopicID
		topicpb.AverageScore = int32(topic.AverageScore.Int)
		topicpb.CompletedStudent = int32(topic.CompletedStudent.Int)
		topicpb.TotalAssignedStudent = int32(topic.TotalAssignStudent.Int)
		lmspb := []*sspb.CourseStatisticResponse_TopicStatistic_LearningMaterialStatistic{}

		if _, ok := mapTopicLms[topic.TopicID]; !ok {
			return nil, status.Errorf(codes.Internal, "Topic not exist in LearningMaterialStatistic")
		}

		for _, lm := range mapTopicLms[topic.TopicID] {
			lmpb := &sspb.CourseStatisticResponse_TopicStatistic_LearningMaterialStatistic{}
			lmpb.LearningMaterialId = lm.LearningMaterialID
			lmpb.TotalAssignedStudent = int32(lm.TotalAssignStudent.Int)
			lmpb.CompletedStudent = int32(lm.CompletedStudent.Int)
			lmpb.AverageScore = int32(lm.AverageScore.Int)

			lmspb = append(lmspb, lmpb)
		}

		if len(lmspb) == 0 {
			return nil, status.Errorf(codes.Internal, "LearningMaterialStatistic len is 0")
		}

		topicpb.LearningMaterialStatistic = lmspb

		topicsStatistic = append(topicsStatistic, topicpb)
	}

	resp.TopicStatistic = topicsStatistic

	return resp, nil
}

func getAchievementCrownV2(score float32) sspb.AchievementCrown {
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

func getLearningMaterialType(lmType string) sspb.LearningMaterialType {
	switch lmType {
	case sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE.String():
		return sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE
	case sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String():
		return sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD
	case sspb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT.String():
		return sspb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT
	case sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT.String():
		return sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT
	default:
		return sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO
	}
}

func getStudyPlanItemStatus(status string) sspb.StudyPlanItemStatus {
	switch status {
	case sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE.String():
		return sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE
	case sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ARCHIVED.String():
		return sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ARCHIVED
	default:
		return sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE
	}
}

func (s *StatisticService) ListSubmissions(ctx context.Context, req *sspb.ListSubmissionsRequest) (*sspb.ListSubmissionsResponse, error) {
	if err := s.validateListSubmissionsReq(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("validateListSubmissionsReq: %w", err).Error())
	}

	studyPlanItemIdentities := sliceutils.Map(req.StudyPlanItemIdentities, func(el *sspb.StudyPlanItemIdentity) *repositories.StudyPlanItemIdentity {
		return &repositories.StudyPlanItemIdentity{
			StudyPlanID:        database.Text(el.StudyPlanId),
			LearningMaterialID: database.Text(el.LearningMaterialId),
			StudentID:          database.Text(el.StudentId.Value),
		}
	})

	submissions, err := s.StudentSubmissionRepo.RetrieveByStudyPlanIdentities(ctx, s.DB, studyPlanItemIdentities)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.StudentSubmissionRepo.RetrieveByStudyPlanIdentities: %w", err).Error())
	}

	pbSubmissions := sliceutils.Map(submissions, func(el *repositories.StudentSubmissionInfo) *sspb.Submission {
		return s.toSubmissionProto(el.StudentSubmission, el.CourseID, el.StartDate, el.EndDate)
	})

	return &sspb.ListSubmissionsResponse{
		Submissions: pbSubmissions,
	}, nil
}

func (s *StatisticService) validateListSubmissionsReq(req *sspb.ListSubmissionsRequest) error {
	for i, identity := range req.StudyPlanItemIdentities {
		if identity.StudyPlanId == "" {
			return fmt.Errorf("StudyPlanItemIdentities[%d]: StudyPlanId must not empty", i)
		}

		if identity.LearningMaterialId == "" {
			return fmt.Errorf("StudyPlanItemIdentities[%d]: LearningMaterialId must not empty", i)
		}

		if identity.StudentId.GetValue() == "" {
			return fmt.Errorf("StudyPlanItemIdentities[%d]: StudentId must not empty", i)
		}
	}
	return nil
}

func (s *StatisticService) toSubmissionProto(e entities.StudentSubmission, courseID pgtype.Text, startDate, endDate pgtype.Timestamptz) *sspb.Submission {
	pb := &sspb.Submission{
		SubmissionId: e.ID.String,
		StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
			StudyPlanId:        e.StudyPlanID.String,
			LearningMaterialId: e.LearningMaterialID.String,
			StudentId:          wrapperspb.String(e.StudentID.String),
		},
		Note:               e.Note.String,
		CreatedAt:          timestamppb.New(e.CreatedAt.Time),
		UpdatedAt:          timestamppb.New(e.UpdatedAt.Time),
		Status:             sspb.SubmissionStatus(sspb.SubmissionStatus_value[e.Status.String]),
		CourseId:           courseID.String,
		StartDate:          timestamppb.New(startDate.Time),
		EndDate:            timestamppb.New(endDate.Time),
		Duration:           e.Duration.Int,
		UnderstandingLevel: sspb.SubmissionUnderstandingLevel(sspb.SubmissionUnderstandingLevel_value[e.UnderstandingLevel.String]),
	}
	if err := e.SubmissionContent.AssignTo(&pb.SubmissionContent); err != nil {
		pb.SubmissionContent = []*sspb.SubmissionContent{}
	}
	if e.SubmissionGradeID.Status == pgtype.Present {
		pb.SubmissionGradeId = wrapperspb.String(e.SubmissionGradeID.String)
	}
	if e.CorrectScore.Status == pgtype.Present {
		pb.CorrectScore = wrapperspb.Float(e.CorrectScore.Float)
	}
	if e.TotalScore.Status == pgtype.Present {
		pb.TotalScore = wrapperspb.Float(e.TotalScore.Float)
	}
	if e.CompleteDate.Status == pgtype.Present {
		pb.CompleteDate = timestamppb.New(e.CompleteDate.Time)
	}

	return pb
}

func validateRetrieveLearningProgressRequestV2(ctx context.Context, req *sspb.RetrieveLearningProgressRequest) error {
	// if !dateRangeValid(req.From, req.To) {
	// 	return status.Error(codes.InvalidArgument, codes.InvalidArgument.String())
	// }

	if !canProcessStudentData(ctx, req.StudentId) {
		return status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	return nil
}

func (s *StatisticService) normalizeTimeRetrieveLearningProgress(ctx context.Context, req *sspb.RetrieveLearningProgressRequest) (*pgtype.Timestamptz, *pgtype.Timestamptz, error) {
	mdCtx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, nil, status.Errorf(codes.Unauthenticated, codes.Unauthenticated.String())
	}
	resp, err := s.UserMgmtService.SearchBasicProfile(mdCtx, &upb.SearchBasicProfileRequest{
		UserIds: []string{req.StudentId},
		Paging: &cpb.Paging{
			Limit: 1,
		},
	})
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "RetrieveUserProfile: %v", err.Error())
	}
	if len(resp.Profiles) == 0 {
		return nil, nil, status.Error(codes.NotFound, codes.NotFound.String())
	}
	country := pb.Country(resp.Profiles[0].Country)

	// For this api, the client should send:
	//  - req.From is Monday 00:00:00 on student's local time
	//  - req.To is Sunday 23:59:59 on student's local time
	// and because both req.From and req.To are using google.protobuf.Timestamp,
	// so if the student's country is VN, which is UTC+07, the expected data are:
	//  - req.From is Sunday 17:00:00 +00
	//  - req.To is next Sunday 16:59:59 +00
	//  but currently the data client send are:
	//  - req.From is Monday 00:00:00 +00
	//  - req.To is Sunday 23:59:59 +00
	// so we must change the req.From and req.To to match with expected data above.
	// TODO: remove this when the client send same with expected data.
	tFrom := req.From.AsTime()
	if tFrom.Hour() == 0 && tFrom.Minute() == 0 && tFrom.Second() == 0 {
		if country == pb.Country(cpb.Country_COUNTRY_VN) {
			tFrom = tFrom.Add(-7 * time.Hour)
		}
	}
	from := new(pgtype.Timestamptz)
	from.Set(tFrom)

	tTo := req.To.AsTime()
	if tTo.Hour() == 23 && tTo.Minute() == 59 && tTo.Second() == 59 {
		if country == pb.Country(cpb.Country_COUNTRY_VN) {
			tTo = tTo.Add(-7 * time.Hour)
		}
	}
	to := new(pgtype.Timestamptz)
	to.Set(tTo)
	return from, to, nil
}

func (s *StatisticService) RetrieveLearningProgress(ctx context.Context, req *sspb.RetrieveLearningProgressRequest) (*sspb.RetrieveLearningProgressResponse, error) {
	if err := validateRetrieveLearningProgressRequestV2(ctx, req); err != nil {
		return nil, err
	}
	from, to, err := s.normalizeTimeRetrieveLearningProgress(ctx, req)
	if err != nil {
		return nil, err
	}

	learningTimeByDailies, err := s.StudentLearningTimeDaiyRepo.RetrieveV2(ctx, s.DB, database.Text(req.StudentId), from, to)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.StudentLearningTimeDaiyRepo.RetrieveV2 %w", err).Error())
	}

	var ret []*sspb.RetrieveLearningProgressResponse_DailyLearningTime
	for from.Time.Before(to.Time) {
		lt := &sspb.RetrieveLearningProgressResponse_DailyLearningTime{
			Day: timestamppb.New(from.Time),
		}
		for _, d := range learningTimeByDailies {
			if d.Day.Time.Equal(from.Time) && d.LearningTime.Int > 0 {
				lt.TotalTimeSpentInDay = int64(d.LearningTime.Int)
				break
			}
		}
		ret = append(ret, lt)
		from.Time = from.Time.Add(24 * time.Hour)
	}

	return &sspb.RetrieveLearningProgressResponse{Dailies: ret}, nil
}

func (s *StatisticService) RetrieveSchoolHistoryByStudentInCourse(ctx context.Context, req *sspb.RetrieveSchoolHistoryByStudentInCourseRequest) (*sspb.RetrieveSchoolHistoryByStudentInCourseResponse, error) {
	studentIDs, err := s.CourseStudentRepo.FindStudentByCourseID(ctx, s.DB, database.Text(req.GetCourseId()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.CourseStudentRepo.FindStudentByCourseID %w", err).Error())
	}

	if len(studentIDs) == 0 {
		return &sspb.RetrieveSchoolHistoryByStudentInCourseResponse{}, nil
	}

	cctx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, err
	}
	schoolInfos, err := s.StudentReaderClient.RetrieveStudentSchoolHistory(cctx, &bpb.RetrieveStudentSchoolHistoryRequest{
		StudentIds: studentIDs,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.StudentReaderClient.RetrieveStudentSchoolHistory %w", err).Error())
	}

	schools := make(map[string]*sspb.RetrieveSchoolHistoryByStudentInCourseResponse_School, 0)
	for schoolID, si := range schoolInfos.GetSchools() {
		schools[schoolID] = &sspb.RetrieveSchoolHistoryByStudentInCourseResponse_School{
			SchoolId:   si.GetSchoolId(),
			SchoolName: si.GetSchoolName(),
		}
	}

	return &sspb.RetrieveSchoolHistoryByStudentInCourseResponse{
		Schools: schools,
	}, nil
}

func (s *StatisticService) ListTagByStudentInCourse(ctx context.Context, req *sspb.ListTagByStudentInCourseRequest) (*sspb.ListTagByStudentInCourseResponse, error) {
	if req.CourseId == "" {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("cannot empty course_id").Error())
	}

	tags, err := s.CourseStudentRepo.FindStudentTagByCourseID(ctx, s.DB, database.Text(req.CourseId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.CourseStudentRepo.FindStudentTagByCourseID %w", err).Error())
	}
	if len(tags) == 0 {
		return &sspb.ListTagByStudentInCourseResponse{}, nil
	}

	studentTags := make([]*sspb.ListTagByStudentInCourseResponse_StudentTag, 0, len(tags))
	for _, tag := range tags {
		studentTags = append(studentTags, &sspb.ListTagByStudentInCourseResponse_StudentTag{
			TagId:   tag.ID.String,
			TagName: tag.Name.String,
		})
	}

	return &sspb.ListTagByStudentInCourseResponse{
		StudentTags: studentTags,
	}, nil
}

func differenceSlice(slice1 []string, slice2 []string) []string {
	var diff []string
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			if !found {
				diff = append(diff, s1)
			}
		}
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}
