package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	ys_pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EurekaStudentSubmissionReaderService interface {
	RetrieveStudentSubmissionHistoryByLoIDs(ctx context.Context, req *epb.RetrieveStudentSubmissionHistoryByLoIDsRequest, opts ...grpc.CallOption) (*epb.RetrieveStudentSubmissionHistoryByLoIDsResponse, error)
}

// CourseService implement proto CourseServer
type CourseService struct {
	EurekaDBTrace        database.Ext
	DB                   database.Ext
	Cfg                  *configurations.Config
	Env                  string
	EurekaTopicReaderSvc epb.TopicReaderServiceClient
	QuestionRepo         interface {
		CreateAll(context.Context, database.QueryExecer, []entities.Question, []entities.QuestionTagLo) ([]*entities.Question, error)
		ExistMasterQuestion(context.Context, database.QueryExecer, string) (bool, error)
		RetrieveQuizSetsFromLoId(ctx context.Context, db database.QueryExecer, loID pgtype.Text, topicType pgtype.Text, limit, page int) (*repositories.QuestionPagination, error)
		RetrieveQuestionTagLo(ctx context.Context, db database.QueryExecer, questionId pgtype.TextArray) (map[string][]string, error)
		RetrieveQuiz(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.Question, error)
	}
	UserRepo interface {
		UserGroup(context.Context, database.QueryExecer, pgtype.Text) (string, error)
		Get(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (*entities.User, error)
		Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]*entities.User, error)
	}
	TopicRepo interface {
		Retrieve(ctx context.Context, db database.QueryExecer, country, subject, topicType, status string, grade int) ([]*entities.Topic, error)
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.Topic, error)
		BulkImport(ctx context.Context, db database.QueryExecer, topics []*entities.Topic) error
		GetTopicFromLoId(ctx context.Context, db database.QueryExecer, loID pgtype.Text) (*entities.Topic, error)
		UpdateTotalLOs(ctx context.Context, db database.QueryExecer, topicID pgtype.Text) error
		FindByChapterIds(ctx context.Context, db database.QueryExecer, chapterIds []string) ([]*entities.Topic, error)
		BulkUpsertWithoutDisplayOrder(ctx context.Context, db database.QueryExecer, topics []*entities.Topic) error
		RetrieveByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Topic, error)
		UpdateLODisplayOrderCounter(ctx context.Context, db database.QueryExecer, topicID pgtype.Text, number pgtype.Int4) error
	}
	QuestionSetsRepo interface {
		CreateAll(ctx context.Context, db database.QueryExecer, quizsets []*entities.QuestionSets) error
		FindByQuizID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.QuestionSets, error)
	}
	QuizSetRepo interface {
		CountQuizOnLO(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (map[string]int32, error)
	}
	LearningObjectiveRepo interface {
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.LearningObjective, error)
		RetrieveByTopicIDs(ctx context.Context, db database.QueryExecer, topicIds pgtype.TextArray) ([]*entities.LearningObjective, error)
		BulkImport(ctx context.Context, db database.QueryExecer, learningObjectives []*entities.LearningObjective) error
		SuggestByLOName(ctx context.Context, db database.QueryExecer, loName string) ([]*entities.LearningObjective, error)
	}
	TopicLearningObjectiveRepo interface {
		Create(ctx context.Context, db database.QueryExecer, m *entities.TopicsLearningObjectives) error
		BulkImport(ctx context.Context, db database.QueryExecer, m []*entities.TopicsLearningObjectives) error
	}
	StudentsLearningObjectivesCompletenessRepo interface {
		Find(ctx context.Context, db database.QueryExecer, studentId pgtype.Text, loIds pgtype.TextArray) (map[pgtype.Text]*entities.StudentsLearningObjectivesCompleteness, error)
	}
	StudentEventLogRepo interface {
		LogsQuestionSubmitionByLO(ctx context.Context, db database.QueryExecer, studentID string, loIDs pgtype.TextArray) (map[string][]*pb.SubmissionResult, error)
		RetrieveStudentSubmissionsByLO(ctx context.Context, db database.QueryExecer, studentID, loID pgtype.Text, order repositories.SubmissionOrderType, limit, offset int) (*repositories.SubmissionPagination, error)
		GetSubmitAnswerEventLog(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, quizID pgtype.Text) (*entities.StudentEventLog, error)
	}
	PresetStudyPlanRepo interface {
		CreatePresetStudyPlan(ctx context.Context, db database.Ext, presetStudyPlans []*entities.PresetStudyPlan) error
		CreatePresetStudyPlanWeekly(ctx context.Context, db database.QueryExecer, presetStudyPlanWeeklies []*entities.PresetStudyPlanWeekly) error
	}
	StudentTopicCompletenessRepo interface {
		RetrieveByStudentID(context.Context, database.QueryExecer, pgtype.Text, *pgtype.Text) ([]*entities.StudentTopicCompleteness, error)
	}
	ChapterRepo interface {
		FindWithFilter(context.Context, database.QueryExecer, []string, string, string, int, uint32, uint32) ([]*entities.Chapter, int, error)
		FindByIDs(ctx context.Context, db database.QueryExecer, chapterIDs []string) (map[string]*entities.Chapter, error)
		FindByBookID(ctx context.Context, db database.QueryExecer, bookID string) ([]*entities.Chapter, error)
		FindByID(ctx context.Context, db database.QueryExecer, chapterID pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Chapter, error)
		UpdateCurrentTopicDisplayOrder(ctx context.Context, db database.QueryExecer, totalGeneratedTopicDisplayOrder pgtype.Int4, chapterID pgtype.Text) error
	}
	CourseRepo interface {
		CountCourses(context.Context, database.QueryExecer, *repositories.CourseQuery) (int, error)
		RetrieveCourses(context.Context, database.QueryExecer, *repositories.CourseQuery) (entities.Courses, error)
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (entities.Courses, error)
	}
	ClassRepo interface {
		FindJoined(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]*entities.Class, error)
	}
	SchoolAdminRepo interface {
		Get(context.Context, database.QueryExecer, pgtype.Text) (*entities.SchoolAdmin, error)
	}
	TeacherRepo interface {
		GetTeacherHasSchoolIDs(context.Context, database.QueryExecer, string, []int32) (*entities.Teacher, error)
		FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.Teacher, error)
	}
	ActivityLogRepo interface {
		BulkImport(ctx context.Context, db database.QueryExecer, logs []*entities.ActivityLog) error
	}
	LessonRepo interface {
		FindLessonWithTime(ctx context.Context, db database.QueryExecer, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, limit int32, page int32, schedulingStatus pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error)
		FindLessonJoined(ctx context.Context, db database.QueryExecer, userID pgtype.Text, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, limit int32, page int32, schedulingStatus pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error)
		FindLessonJoinedV2(ctx context.Context, db database.QueryExecer, filter *repositories.LessonJoinedV2Filter, limit int32, page int32) ([]*repositories.LessonWithTime, pgtype.Int8, error)
		Find(ctx context.Context, db database.QueryExecer, filter *repositories.LessonFilter) ([]*entities.Lesson, error)
	}
	CourseClassRepo interface {
		Find(ctx context.Context, db database.QueryExecer, ids pgtype.Int4Array) (mapCourseIDsByClassID map[pgtype.Int4]pgtype.TextArray, err error)
		FindClassInCourse(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray, classIDs pgtype.Int4Array) (mapClassIDByCourseID map[pgtype.Text]pgtype.Int4Array, err error)
		FindByCourseIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (mapClassIDByCourseID map[pgtype.Text]pgtype.Int4Array, err error)
	}
	ClassMemberRepo interface {
		FindUsersClass(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]int32, error)
	}
	SchoolRepo interface {
		RetrieveCountries(ctx context.Context, db database.QueryExecer, schoolIDs pgtype.Int4Array) ([]string, error)
	}
	BrightCoveService interface {
		CreateBrightCoveUploadUrl(ctx context.Context, req *ys_pb.CreateBrightCoveUploadUrlRequest) (*ys_pb.CreateBrightCoveUploadUrlResponse, error)
		FinishUploadBrightCove(ctx context.Context, req *ys_pb.FinishUploadBrightCoveRequest) (*ys_pb.FinishUploadBrightCoveResponse, error)
	}
	CourseBookRepo interface {
		FindByCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs []string) (mapBookIDByCourseID map[string][]string, err error)
	}
	BookRepo interface {
		FindByIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) (map[string]*entities.Book, error)
		FindWithFilter(ctx context.Context, db database.QueryExecer, courseID string, limit, offset uint32) ([]*entities.Book, int, error)
	}
	BookChapterRepo interface {
		FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) (map[string][]*entities.BookChapter, error)
	}
	LessonMemberRepo interface {
		Find(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]*entities.LessonMember, error)
		CourseAccessible(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]string, error)
	}
	ConfigRepo interface {
		RetrieveWithResourcePath(ctx context.Context, db database.QueryExecer, country pgtype.Text, group pgtype.Text, keys pgtype.TextArray, resourcePath pgtype.Text) ([]*entities.Config, error)
	}
	EurekaStudentSubmissionReaderService EurekaStudentSubmissionReaderService
	UnleashClientIns                     unleashclient.ClientInstance
}

func (c *CourseService) ListTopic(ctx context.Context, req *pb.ListTopicRequest) (*pb.ListTopicResponse, error) {
	currentUserID := interceptors.UserIDFromContext(ctx)
	uGroup, err := c.UserRepo.UserGroup(ctx, c.DB, database.Text(currentUserID))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "wrong token")
	}

	var published string
	if uGroup == entities.UserGroupStudent {
		published = pb.TOPIC_STATUS_PUBLISHED.String()
	}
	var country string
	if req.Country != pb.COUNTRY_NONE {
		country = req.Country.String()
	}
	var subject string
	if req.Subject != pb.SUBJECT_NONE {
		subject = req.Subject.String()
	}
	var topicType string
	if req.TopicType != pb.TOPIC_TYPE_NONE {
		topicType = req.TopicType.String()
	}
	var grade int
	if req.Grade == "" {
		grade = -1
	} else {
		grade, err = i18n.ConvertStringGradeToInt(req.Country, req.Grade)
	}
	if err != nil {
		return nil, errors.Wrap(err, "c.TopicRepo.Retrieve")
	}
	topics, err := c.TopicRepo.Retrieve(ctx, c.DB, country, subject, topicType, published, grade)
	if err != nil {
		return nil, errors.Wrap(err, "c.TopicRepo.Retrieve")
	}

	ret := make([]*pb.Topic, 0, len(topics))
	for _, topic := range topics {
		ret = append(ret, ToTopicPb(topic))
	}

	return &pb.ListTopicResponse{Topics: ret}, nil
}

// deprecated
func (c *CourseService) ListTopics(ctx context.Context, req *pb.ListTopicsRequest) (*pb.ListTopicsResponse, error) {
	return &pb.ListTopicsResponse{}, nil
}

// UpsertQuizSets deprecated
func (c *CourseService) UpsertQuizSets(ctx context.Context, req *pb.UpsertQuizRequest) (*pb.UpsertQuizResponse, error) {
	return &pb.UpsertQuizResponse{}, nil
}

// TakeTheQuiz deprecated
func (c *CourseService) TakeTheQuiz(ctx context.Context, req *pb.TakeTheQuizRequest) (*pb.TakeTheQuizResponse, error) {
	return &pb.TakeTheQuizResponse{}, nil
}

func (c *CourseService) UpsertQuestions(ctx context.Context, req *pb.UpsertQuestionsRequest) (*pb.UpsertQuestionsResponse, error) {
	return &pb.UpsertQuestionsResponse{}, nil
}

func (c *CourseService) updateTotalLOs(db database.QueryExecer, topicIDs []string, logger *zap.Logger) {
	if len(topicIDs) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, topicID := range topicIDs {
		if err := c.TopicRepo.UpdateTotalLOs(ctx, db, database.Text(topicID)); err != nil {
			logger.Error("c.TopicRepo.UpdateTotalLOs", zap.String("topic_id", topicID), zap.Error(err))
		}
	}
}

func (c *CourseService) UpsertLOs(ctx context.Context, req *pb.UpsertLOsRequest) (*pb.UpsertLOsResponse, error) {
	ids := make([]string, 0, len(req.LearningObjectives))
	var topicIDs []string

	type Group struct {
		Los      []*entities.LearningObjective
		TopicLos []*entities.TopicsLearningObjectives
	}

	topicGroupMap := make(map[string]*Group)

	for _, lo := range req.LearningObjectives {
		e, err := toLoEntity(lo)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		ids = append(ids, e.ID.String)
		loTopicE := &entities.TopicsLearningObjectives{
			TopicID:      e.TopicID,
			LoID:         e.ID,
			DisplayOrder: e.DisplayOrder,
			CreatedAt:    e.CreatedAt,
			UpdatedAt:    e.UpdatedAt,
			DeletedAt:    e.DeletedAt,
		}
		topicID := e.TopicID.String
		if _, ok := topicGroupMap[topicID]; !ok {
			topicIDs = append(topicIDs, topicID)
			topicGroupMap[topicID] = &Group{}
		}
		topicGroupMap[topicID].Los = append(topicGroupMap[topicID].Los, e)
		topicGroupMap[topicID].TopicLos = append(topicGroupMap[topicID].TopicLos, loTopicE)
	}

	topics, err := c.TopicRepo.RetrieveByIDs(ctx, c.EurekaDBTrace, database.TextArray(topicIDs))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("unable to retrieve topics by ids: %w", err).Error())
	}

	if !isTopicsExisted(topicIDs, topics) {
		return nil, status.Errorf(codes.InvalidArgument, "some topics does not exists")
	}
	for topicID, group := range topicGroupMap {
		los := group.Los
		topicLos := group.TopicLos
		if err := database.ExecInTx(ctx, c.EurekaDBTrace, func(ctx context.Context, tx pgx.Tx) error {
			var insertNum int
			topic, err := c.TopicRepo.RetrieveByID(ctx, tx, database.Text(topicID), repositories.WithUpdateLock())
			if err != nil {
				return fmt.Errorf("unable to retrieve topic by id: %w", err)
			}
			if isAutoGenLODisplayOrder(los) {
				existedLos, err := c.LearningObjectiveRepo.RetrieveByIDs(ctx, tx, database.TextArray(ids))
				if err != nil {
					return fmt.Errorf("unable to retrieve los by ids: %w", err)
				}
				m := make(map[string]*entities.LearningObjective)
				for _, lo := range existedLos {
					m[lo.ID.String] = lo
				}
				count := 0
				total := topic.LODisplayOrderCounter.Int
				for i, lo := range los {
					if e, ok := m[lo.ID.String]; !ok {
						lo.DisplayOrder = database.Int2(int16(total) + int16(count) + 1)
						topicLos[i].DisplayOrder = lo.DisplayOrder
						count++
					} else {
						lo.DisplayOrder = e.DisplayOrder
						topicLos[i].DisplayOrder = lo.DisplayOrder
					}
				}
				insertNum = count
			} else {
				insertNum = len(los)
			}

			if err := c.LearningObjectiveRepo.BulkImport(ctx, tx, los); err != nil {
				return fmt.Errorf("unable to bulk import learning objective: %w", err)
			}
			if err := c.TopicLearningObjectiveRepo.BulkImport(ctx, tx, topicLos); err != nil {
				return fmt.Errorf("unable to bulk import topic learning objective: %w", err)
			}
			if err := c.TopicRepo.UpdateLODisplayOrderCounter(ctx, tx, database.Text(topicID), database.Int4(int32(insertNum))); err != nil {
				return fmt.Errorf("unable to update lo display order counter: %w", err)
			}
			if err := c.TopicRepo.UpdateTotalLOs(ctx, tx, database.Text(topicID)); err != nil {
				return fmt.Errorf("unable to update total learning objectives: %w", err)
			}
			return nil
		}); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
				return nil, status.Error(codes.FailedPrecondition, err.Error())
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	if err != nil {
		return nil, err
	}

	return &pb.UpsertLOsResponse{
		LoIds: ids,
	}, nil
}

func toLoEntity(src *pb.LearningObjective) (*entities.LearningObjective, error) {
	if src.Id == "" {
		src.Id = idutil.ULIDNow()
	}

	if src.Type == pb.LEARNING_OBJECTIVE_TYPE_NONE {
		src.Type = pb.LEARNING_OBJECTIVE_TYPE_LEARNING
	}
	grade, _ := i18n.ConvertStringGradeToInt(src.Country, src.Grade)

	e := new(entities.LearningObjective)
	database.AllNullEntity(e)

	err := multierr.Combine(
		e.ID.Set(src.Id),
		e.Name.Set(src.Name),
		e.Country.Set(src.Country.String()),
		e.Grade.Set(grade),
		e.Subject.Set(src.Subject.String()),
		e.TopicID.Set(src.TopicId),
		e.VideoScript.Set(src.VideoScript),
		e.DisplayOrder.Set(src.DisplayOrder),
		e.Prerequisites.Set(src.Prerequisites),
		e.Video.Set(src.Video),
		e.StudyGuide.Set(src.StudyGuide),
		e.SchoolID.Set(src.SchoolId),
		e.Type.Set(src.Type.String()),
	)

	if src.MasterLo != "" {
		err = multierr.Append(err, e.MasterLoID.Set(src.MasterLo))
	}

	if src.CreatedAt != nil {
		err = multierr.Append(err, e.CreatedAt.Set(time.Unix(src.CreatedAt.Seconds, int64(src.CreatedAt.Nanos))))
	} else {
		err = multierr.Append(err, e.CreatedAt.Set(time.Now()))
	}
	if src.UpdatedAt != nil {
		err = multierr.Append(err, e.UpdatedAt.Set(time.Unix(src.UpdatedAt.Seconds, int64(src.UpdatedAt.Nanos))))
	} else {
		err = multierr.Append(err, e.UpdatedAt.Set(time.Now()))
	}

	if err != nil {
		return nil, fmt.Errorf("entities.LearningObjective set: %w", err)
	}

	return e, nil
}

func (c *CourseService) UpsertPresetStudyPlans(ctx context.Context, req *pb.UpsertPresetStudyPlansRequest) (*pb.UpsertPresetStudyPlansResponse, error) {
	ss := make([]*entities.PresetStudyPlan, 0, len(req.PresetStudyPlans))
	ids := make([]string, 0, len(req.PresetStudyPlans))

	for _, presetStudyPlan := range req.PresetStudyPlans {
		presetStudyPlanModel := toPresetStudyPlanEntity(presetStudyPlan)
		ss = append(ss, presetStudyPlanModel)
		ids = append(ids, presetStudyPlan.Id)
	}

	err := c.PresetStudyPlanRepo.CreatePresetStudyPlan(ctx, c.DB, ss)
	if err != nil {
		return nil, toStatusError(err)
	}

	return &pb.UpsertPresetStudyPlansResponse{
		PresetStudyPlanIds: ids,
	}, nil
}

func (c *CourseService) UpsertPresetStudyPlanWeeklies(ctx context.Context, req *pb.UpsertPresetStudyPlanWeekliesRequest) (*pb.UpsertPresetStudyPlanWeekliesResponse, error) {
	ss := make([]*entities.PresetStudyPlanWeekly, 0, len(req.PresetStudyPlanWeeklies))
	ids := make([]string, 0, len(req.PresetStudyPlanWeeklies))

	for _, presetStudyPlanWeekly := range req.PresetStudyPlanWeeklies {
		presetStudyPlanWeeklyModel := toPresetStudyPlanWeeklyEntity(presetStudyPlanWeekly)
		ss = append(ss, presetStudyPlanWeeklyModel)
		ids = append(ids, presetStudyPlanWeeklyModel.ID.String)
	}

	err := c.PresetStudyPlanRepo.CreatePresetStudyPlanWeekly(ctx, c.DB, ss)
	if err != nil {
		return nil, toStatusError(err)
	}

	return &pb.UpsertPresetStudyPlanWeekliesResponse{
		PresetStudyPlanWeeklyIds: ids,
	}, nil
}

func toPresetStudyPlanEntity(src *pb.PresetStudyPlan) *entities.PresetStudyPlan {
	dest := new(entities.PresetStudyPlan)
	database.AllNullEntity(dest)
	if src.Id == "" {
		src.Id = ksuid.New().String()
	}

	dest.ID.Set(src.Id)
	dest.Name.Set(src.Name)
	dest.Country.Set(src.Country.String())
	grade, _ := i18n.ConvertStringGradeToInt(src.Country, src.Grade)
	dest.Grade.Set(grade)
	dest.Subject.Set(src.Subject.String())

	if src.CreatedAt != nil {
		dest.CreatedAt.Set(time.Unix(src.CreatedAt.Seconds, int64(src.CreatedAt.Nanos)))
	} else {
		dest.CreatedAt.Set(time.Now())
	}

	if src.UpdatedAt != nil {
		dest.UpdatedAt.Set(time.Unix(src.UpdatedAt.Seconds, int64(src.UpdatedAt.Nanos)))
	} else {
		dest.UpdatedAt.Set(time.Now())
	}

	if src.StartDate != nil {
		dest.StartDate.Set(time.Unix(src.StartDate.Seconds, int64(src.StartDate.Nanos)))
	} else {
		defaultDate := timeutil.DefaultPSPStartDate(src.Country)
		dest.StartDate.Set(defaultDate)
	}

	return dest
}

func toPresetStudyPlanWeeklyEntity(src *pb.PresetStudyPlanWeekly) *entities.PresetStudyPlanWeekly {
	dest := new(entities.PresetStudyPlanWeekly)
	if src.Id == "" {
		src.Id = ksuid.New().String()
	}

	dest.ID.Set(src.Id)
	dest.PresetStudyPlanID.Set(src.PresetStudyPlanId)
	dest.TopicID.Set(src.TopicId)
	dest.Week.Set(int16(src.Week))
	dest.CreatedAt.Set(time.Now())
	dest.UpdatedAt.Set(time.Now())

	return dest
}

func (c *CourseService) GetHistoryQuizDetail(ctx context.Context, req *pb.GetHistoryQuizDetailRequest) (*pb.GetHistoryQuizDetailResponse, error) {
	return &pb.GetHistoryQuizDetailResponse{}, nil
}

func (c *CourseService) validateQuestion(ctx context.Context, questionList []*pb.Question) error {
	for _, question := range questionList {
		err := c.rulesForMandatoryString(question.Question)
		if err != nil {
			return err
		}

		err = c.rulesForMasterQuestionID(ctx, question)
		if err != nil {
			return err
		}

		if len(question.Answers) <= 1 {
			return status.Error(codes.InvalidArgument, "wrong format data")
		}

		if question.DifficultyLevel < 1 {
			return status.Error(codes.InvalidArgument, "wrong format data")
		}
	}

	return nil
}

func (c *CourseService) validateAnswer(answersList []string) error {
	for _, answer := range answersList {
		err := c.rulesForMandatoryString(answer)
		if err != nil {
			return status.Error(codes.InvalidArgument, "wrong format data")
		}
	}

	return nil
}

func (c *CourseService) rulesForMandatoryString(s string) error {
	if s == "" {
		return status.Error(codes.InvalidArgument, "empty mandatory field")
	}

	return nil
}

func (c *CourseService) rulesForMasterQuestionID(ctx context.Context, q *pb.Question) error {
	// master question id is not manatory
	err := c.rulesForMandatoryString(q.MasterQuestionId)
	if err != nil {
		return nil
	}
	// master question id refer to a question in db
	exist, err := c.QuestionRepo.ExistMasterQuestion(ctx, c.DB, q.MasterQuestionId)
	if !exist || err != nil {
		return status.Error(codes.FailedPrecondition, "master id does not map to question id")
	}

	return nil
}

// ToTopicPb convert topic ent to proto format
func ToTopicPb(p *entities.Topic) *pb.Topic {
	var (
		country     = pb.Country(pb.Country_value[p.Country.String])
		grade, _    = i18n.ConvertIntGradeToString(country, int(p.Grade.Int))
		publishedAt *types.Timestamp
	)

	if p.PublishedAt.Get() != nil {
		publishedAt = &types.Timestamp{Seconds: p.PublishedAt.Time.Unix()}
	} else {
		publishedAt = nil
	}

	topic := &pb.Topic{
		Id:           p.ID.String,
		Name:         p.Name.String,
		Country:      country,
		Grade:        grade,
		Subject:      pb.Subject(pb.Subject_value[p.Subject.String]),
		Type:         pb.TopicType(pb.TopicType_value[p.TopicType.String]),
		Status:       pb.TopicStatus(pb.TopicStatus_value[p.Status.String]),
		DisplayOrder: int32(p.DisplayOrder.Int),
		CreatedAt:    &types.Timestamp{Seconds: p.CreatedAt.Time.Unix()},
		UpdatedAt:    &types.Timestamp{Seconds: p.UpdatedAt.Time.Unix()},
		PublishedAt:  publishedAt,
		TotalLos:     p.TotalLOs.Int,
		ChapterId:    p.ChapterID.String,
		SchoolId:     getSchool(p.SchoolID.Int),
		Instruction:  p.Instruction.String,
		IconUrl:      p.IconURL.String,
	}

	if p.CopiedTopicID.Status == pgtype.Present {
		topic.CopiedTopicId = &types.StringValue{
			Value: p.CopiedTopicID.String,
		}
	}

	numAttachment := len(p.AttachmentNames.Elements)
	if n := len(p.AttachmentURLs.Elements); n < numAttachment {
		numAttachment = n
	}

	for i := 0; i < numAttachment; i++ {
		topic.Attachments = append(topic.Attachments, &pb.Attachment{
			Name: p.AttachmentNames.Elements[i].String,
			Url:  p.AttachmentURLs.Elements[i].String,
		})
	}

	return topic
}

func (c *CourseService) RetrieveGradeMap(ctx context.Context, req *pb.RetrieveGradeMapRequest) (*pb.RetrieveGradeMapResponse, error) {
	gradeMap := make(map[string]*pb.LocalGrade)
	for country, localGrade := range i18n.OutGradeMap {
		grade := make(map[string]int32, len(i18n.OutGradeMap[country]))
		for k, v := range localGrade {
			grade[v] = int32(k)
		}

		gradeMap[country.String()] = &pb.LocalGrade{
			LocalGrade: grade,
		}
	}

	return &pb.RetrieveGradeMapResponse{
		GradeMap: gradeMap,
	}, nil
}

func (c *CourseService) SuggestLO(ctx context.Context, req *pb.SuggestLORequest) (*pb.SuggestLOResponse, error) {
	return &pb.SuggestLOResponse{}, nil
}

func (c *CourseService) toChaptersPb(ctx context.Context, chapters []*entities.Chapter) ([]*pb.Chapter, error) {
	chapterIDs := []string{}
	for _, v := range chapters {
		if v.ID.String != "" {
			chapterIDs = append(chapterIDs, v.ID.String)
		}
	}

	enTopics := map[string][]*entities.Topic{}

	if len(chapterIDs) != 0 {
		topics, err := c.TopicRepo.FindByChapterIds(ctx, c.DB, chapterIDs)
		if err != nil {
			return nil, fmt.Errorf("TopicRepo.FindByChapterIds: %w", err)
		}

		for _, v := range topics {
			chapterID := v.ChapterID.String
			enTopics[chapterID] = append(enTopics[chapterID], v)
		}
	}

	rChapters := []*pb.Chapter{}
	for _, v := range chapters {
		vChapter, err := c.toChapterPb(v, enTopics[v.ID.String])
		if err != nil {
			return nil, fmt.Errorf("toChapterPb: %w", err)
		}

		rChapters = append(rChapters, vChapter)
	}
	return rChapters, nil
}

func (c *CourseService) toChapterPb(src *entities.Chapter, enTopics []*entities.Topic) (*pb.Chapter, error) {
	country := pb.Country(pb.Country_value[src.Country.String])

	grade, _ := i18n.ConvertIntGradeToString(country, int(src.Grade.Int))

	updatedAt, err := types.TimestampProto(src.UpdatedAt.Time)
	if err != nil {
		return nil, err
	}

	createdAt, err := types.TimestampProto(src.CreatedAt.Time)
	if err != nil {
		return nil, err
	}
	topics := []*pb.Topic{}
	for _, v := range enTopics {
		topics = append(topics, ToTopicPb(v))
	}

	return &pb.Chapter{
		ChapterId:    src.ID.String,
		ChapterName:  src.Name.String,
		Country:      country,
		Subject:      pb.Subject(pb.Subject_value[src.Subject.String]),
		Grade:        grade,
		DisplayOrder: int32(src.DisplayOrder.Int),
		Topics:       topics,
		SchoolId:     getSchool(src.SchoolID.Int),
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}

// GetChapterList get chapter list with search by id, name
func (c *CourseService) GetChapterList(ctx context.Context, req *pb.GetChapterListRequest) (*pb.GetChapterListResponse, error) {
	return &pb.GetChapterListResponse{}, nil
}

func (c *CourseService) toBookPb(ctx context.Context, src *entities.Book, enChapters []*entities.Chapter) (*pb.Book, error) {
	updatedAt, err := types.TimestampProto(src.UpdatedAt.Time)
	if err != nil {
		return nil, err
	}

	createdAt, err := types.TimestampProto(src.CreatedAt.Time)
	if err != nil {
		return nil, err
	}
	b := &pb.Book{
		Id:        src.ID.String,
		Name:      src.Name.String,
		Subject:   pb.Subject(pb.Subject_value[src.Subject.String]),
		SchoolId:  getSchool(src.SchoolID.Int),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	country := pb.Country(pb.Country_value[src.Country.String])
	b.Country = country

	if src.Grade.Int > 0 {
		grade, err := i18n.ConvertIntGradeToString(country, int(src.Grade.Int))
		if err != nil {
			return nil, err
		}

		b.Grade = grade
	}

	chapters, err := c.toChaptersPb(ctx, enChapters)
	if err != nil {
		return nil, err
	}

	b.Chapters = chapters

	return b, nil
}

func (c *CourseService) toBooksPb(ctx context.Context, src []*entities.Book) ([]*pb.Book, error) {
	rBooks := []*pb.Book{}

	for _, v := range src {
		chapters, err := c.ChapterRepo.FindByBookID(ctx, c.DB, v.ID.String)
		if err != nil {
			return nil, fmt.Errorf("c.BookRepo.FindByBookID: %w", err)
		}

		book, err := c.toBookPb(ctx, v, chapters)
		if err != nil {
			return nil, fmt.Errorf("c.toBookPb: %w", err)
		}

		rBooks = append(rBooks, book)
	}

	return rBooks, nil
}

func (c *CourseService) toCoursesPb(ctx context.Context, courses []*entities.Course,
	mapTeacherBasicProfile map[string]*pb.BasicProfile, mapCourseClass map[pgtype.Text]pgtype.Int4Array, mapCourseBook map[string][]string, bookChapterMap map[string][]*entities.BookChapter) ([]*pb.RetrieveCoursesResponse_Course, error) {
	pbCourses := make([]*pb.RetrieveCoursesResponse_Course, 0, len(courses))

	courseIDs := []string{}
	chapterIDs := map[string][]string{}

	for _, course := range courses {
		courseIDs = append(courseIDs, course.ID.String)

		// return chapters when have only a book, to avoid beak change
		bookIDs, ok := mapCourseBook[course.ID.String]
		if ok && len(bookIDs) == 1 {
			bookID := bookIDs[0]
			eChapterIDs, ok := bookChapterMap[bookID]
			if ok {
				for _, bc := range eChapterIDs {
					chapterIDs[bookID] = append(chapterIDs[bookID], bc.ChapterID.String)
				}
			}
		}
	}

	for _, course := range courses {
		lessonCourse := []*entities.Lesson{}
		coursePb := toCoursePb(course, lessonCourse)

		var teacherIDs []string
		if err := course.TeacherIDs.AssignTo(&teacherIDs); err != nil {
			return nil, err
		}

		for _, teacherID := range teacherIDs {
			teacherBasicProfile, ok := mapTeacherBasicProfile[teacherID]
			if ok {
				coursePb.Teachers = append(coursePb.Teachers, teacherBasicProfile)
			}
		}

		var classIDs []int32
		pgClassIDs, found := mapCourseClass[course.ID]
		if found {
			_ = pgClassIDs.AssignTo(&classIDs)
		}
		coursePb.ClassIds = classIDs

		bookIDs, ok := mapCourseBook[course.ID.String]
		if ok {
			coursePb.BookIds = bookIDs
			if len(bookIDs) == 1 {
				bookID := bookIDs[0]

				courseChapters := []*entities.Chapter{}
				enChapterIDs, ok := chapterIDs[bookID]
				var err error

				if ok {
					courseChapters, _, err = c.ChapterRepo.FindWithFilter(ctx, c.DB, enChapterIDs, "", "", 0, 0, 0)
					if err != nil {
						return nil, fmt.Errorf("ChapterRepo.FindWithFilter: %w", err)
					}
				}

				rChapters, err := c.toChaptersPb(ctx, courseChapters)
				if err != nil {
					return nil, err
				}
				coursePb.Chapters = rChapters
			}
		}

		pbCourses = append(pbCourses, coursePb)
	}
	return pbCourses, nil
}

func (c *CourseService) RetrieveCourses(ctx context.Context, req *pb.RetrieveCoursesRequest) (*pb.RetrieveCoursesResponse, error) {
	schoolIDs := []int{constants.ManabieSchool}

	if req.CourseType == pb.COURSE_TYPE_NONE {
		req.CourseType = pb.COURSE_TYPE_CONTENT
	}

	// check class joined, to get schoolIDs
	classes, err := c.ClassRepo.FindJoined(ctx, c.DB, database.Text(interceptors.UserIDFromContext(ctx)))
	if err != nil && errors.Cause(err) != pgx.ErrNoRows {
		return nil, toStatusError(err)
	}

	for _, class := range classes {
		schoolIDs = append(schoolIDs, int(class.SchoolID.Int))
	}

	limit := req.Limit
	if req.Limit < 1 || req.Limit > 100 {
		limit = 10
	}

	grade := 0
	if req.Country.String() != "" && req.Grade != "" {
		r, err := i18n.ConvertStringGradeToInt(req.Country, req.Grade)
		if err != nil {
			return nil, err
		}
		grade = r
	}

	var countries []string
	if len(req.Countries) > 0 {
		countries = make([]string, 0, len(req.Countries))
		for _, country := range req.Countries {
			countries = append(countries, country.String())
		}
	}

	if req.Country != pb.COUNTRY_NONE && !inArrayString(req.Country.String(), countries) {
		countries = append(countries, req.Country.String())
	}

	classIDs := make([]int, 0, len(classes))
	if req.IsAssigned {
		for _, c := range classes {
			classIDs = append(classIDs, int(c.ID.Int))
		}

		schoolIDs = []int{}
		if req.CourseType == pb.COURSE_TYPE_CONTENT {
			schoolIDs = []int{constants.ManabieSchool}
		}
	}

	// IsAssigned and classId request is mutually exclusive
	if req.ClassId != 0 {
		classIDs = []int{int(req.ClassId)}
		schoolIDs = []int{}
	}

	var ids []string
	if req.Id != "" {
		ids = []string{req.Id}
	}

	query := &repositories.CourseQuery{
		IDs:       ids,
		Name:      req.Name,
		Countries: countries,
		Subject:   req.Subject.String(),
		Grade:     grade,
		SchoolIDs: schoolIDs,
		ClassIDs:  classIDs,
		Limit:     int(limit),
		Offset:    int((req.Page - 1) * limit),
		Type:      req.CourseType.String(),
		Status:    req.CourseStatus.String(),
	}

	return c.handleRetrieveCourses(ctx, query)
}

func toCoursePb(e *entities.Course, lessons []*entities.Lesson) *pb.RetrieveCoursesResponse_Course {
	status := pb.COURSE_STATUS_ACTIVE
	country := pb.Country(pb.Country_value[e.Country.String])
	grade, _ := i18n.ConvertIntGradeToString(country, int(e.Grade.Int))

	updatedAt, _ := types.TimestampProto(e.UpdatedAt.Time)
	createdAt, _ := types.TimestampProto(e.CreatedAt.Time)

	return &pb.RetrieveCoursesResponse_Course{
		Id:           e.ID.String,
		Name:         e.Name.String,
		Country:      pb.Country(pb.Country_value[e.Country.String]),
		Subject:      pb.Subject(pb.Subject_value[e.Subject.String]),
		Grade:        grade,
		SchoolId:     getSchool(e.SchoolID.Int),
		UpdatedAt:    updatedAt,
		CreatedAt:    createdAt,
		StartDate:    &types.Timestamp{Seconds: e.StartDate.Time.Unix()},
		EndDate:      &types.Timestamp{Seconds: e.EndDate.Time.Unix()},
		CourseStatus: status,
	}
}

func (c *CourseService) RetrieveStudentSubmissions(ctx context.Context, req *pb.RetrieveStudentSubmissionsRequest) (*pb.RetrieveStudentSubmissionsResponse, error) {
	return &pb.RetrieveStudentSubmissionsResponse{}, nil
}

func toBasicProfile(src *entities.User) *pb.BasicProfile {
	return &pb.BasicProfile{
		UserId:      src.ID.String,
		Name:        src.GetName(),
		Avatar:      src.Avatar.String,
		UserGroup:   src.Group.String,
		FacebookId:  src.FacebookID.String,
		AppleUserId: src.AppleUser.ID.String,
	}
}

func (c *CourseService) RetrieveAssignedCourses(ctx context.Context, req *pb.RetrieveCoursesRequest) (*pb.RetrieveCoursesResponse, error) {
	limit := req.Limit
	if req.Limit < 1 || req.Limit > 100 {
		limit = 10
	}

	if req.CourseType == pb.COURSE_TYPE_NONE {
		req.CourseType = pb.COURSE_TYPE_CONTENT
	}
	grade := 0
	if req.Country.String() != "" && req.Grade != "" {
		r, err := i18n.ConvertStringGradeToInt(req.Country, req.Grade)
		if err != nil {
			return nil, err
		}
		grade = r
	}

	classes, err := c.ClassRepo.FindJoined(ctx, c.DB, database.Text(interceptors.UserIDFromContext(ctx)))
	if err != nil && errors.Cause(err) != pgx.ErrNoRows {
		return nil, toStatusError(err)
	}

	classIDs := make([]int, 0, len(classes))
	for _, c := range classes {
		classIDs = append(classIDs, int(c.ID.Int))
	}

	var countries []string
	if len(req.Countries) > 0 {
		countries = make([]string, 0, len(req.Countries))
		for _, country := range req.Countries {
			countries = append(countries, country.String())
		}
	}

	if req.Country != pb.COUNTRY_NONE && !inArrayString(req.Country.String(), countries) {
		countries = append(countries, req.Country.String())
	}

	var ids []string
	if req.Id != "" {
		ids = []string{req.Id}
	}
	query := &repositories.CourseQuery{
		IDs:       ids,
		Name:      req.Name,
		Countries: countries,
		Subject:   req.Subject.String(),
		Grade:     grade,
		SchoolIDs: []int{constants.ManabieSchool},
		ClassIDs:  classIDs,
		Limit:     int(limit),
		Offset:    int((req.Page - 1) * limit),
		Type:      req.CourseType.String(),
		Status:    req.CourseStatus.String(),
	}

	return c.handleRetrieveCourses(ctx, query)
}

func (c *CourseService) TakeTheQuizV2(ctx context.Context, req *pb.TakeTheQuizRequest) (*pb.TakeTheQuizV2Response, error) {
	return &pb.TakeTheQuizV2Response{}, nil
}

func (c *CourseService) CreateBrightCoveUploadUrl(ctx context.Context, req *pb.CreateBrightCoveUploadUrlRequest) (*pb.CreateBrightCoveUploadUrlResponse, error) {
	resp, err := c.BrightCoveService.CreateBrightCoveUploadUrl(ctx, &ys_pb.CreateBrightCoveUploadUrlRequest{
		Name: req.Name,
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateBrightCoveUploadUrlResponse{
		SignedUrl:     resp.SignedUrl,
		ApiRequestUrl: resp.ApiRequestUrl,
		VideoId:       resp.VideoId,
	}, err
}

func (c *CourseService) FinishUploadBrightCove(ctx context.Context, req *pb.FinishUploadBrightCoveRequest) (*pb.FinishUploadBrightCoveResponse, error) {
	_, err := c.BrightCoveService.FinishUploadBrightCove(ctx, &ys_pb.FinishUploadBrightCoveRequest{
		ApiRequestUrl: req.ApiRequestUrl,
		VideoId:       req.VideoId,
	})
	if err != nil {
		return nil, err
	}

	return &pb.FinishUploadBrightCoveResponse{}, err
}

func (c *CourseService) handleRetrieveCourses(ctx context.Context, query *repositories.CourseQuery) (*pb.RetrieveCoursesResponse, error) {
	resp := &pb.RetrieveCoursesResponse{
		Total:   0,
		Courses: []*pb.RetrieveCoursesResponse_Course{},
	}

	total, err := c.CourseRepo.CountCourses(ctx, c.DB, query)
	if err != nil {
		return nil, errors.Wrap(err, "c.CourseRepo.CountCourses")
	}

	if total == 0 {
		return resp, nil
	}

	courses, err := c.CourseRepo.RetrieveCourses(ctx, c.DB, query)
	if err != nil {
		return nil, errors.Wrap(err, "c.CourseRepo.RetrieveCourses")
	}

	if len(courses) == 0 {
		if total > 0 {
			return nil, status.Error(codes.NotFound, "cannot find courses")
		}
		return resp, nil
	}

	resp, err = c.coursesDecoration(ctx, courses)
	if err != nil {
		return nil, err
	}

	resp.Total = int32(total)
	return resp, nil
}

func (c *CourseService) IsJoinCourse(ctx context.Context, courseID, userID string) (bool, error) {
	if courseID == "" {
		return false, status.Error(codes.InvalidArgument, "missing course id")
	}
	if userID == "" {
		return false, status.Error(codes.InvalidArgument, "missing user id")
	}

	courses, err := c.CourseRepo.RetrieveByIDs(ctx, c.DB, database.TextArray([]string{courseID}))
	if err != nil {
		return false, fmt.Errorf("CourseRepo.RetrieveByIDs: %w", err)
	}

	if len(courses) == 0 {
		return false, status.Error(codes.NotFound, "cannot find course")
	}
	course := courses[0]

	if course.TeacherIDs.Status == pgtype.Present && len(course.TeacherIDs.Elements) > 0 {
		for _, teacherID := range course.TeacherIDs.Elements {
			if teacherID.String == userID {
				return true, nil
			}
		}
	}

	classes, err := c.ClassRepo.FindJoined(ctx, c.DB, database.Text(userID))
	if err != nil {
		return false, fmt.Errorf("c.ClassRepo.FindJoined: %w", err)
	}

	if len(classes) == 0 {
		return false, status.Error(codes.NotFound, "user do not join this course")
	}

	mapClassIDs := map[int32]int32{}
	classIDs := []int32{}
	for _, class := range classes {
		classID := class.ID.Int

		mapClassIDs[classID] = classID
		classIDs = append(classIDs, classID)
	}

	courseClasses, err := c.CourseClassRepo.FindClassInCourse(ctx, c.DB, database.TextArray([]string{courseID}), database.Int4Array(classIDs))
	if err != nil {
		return false, fmt.Errorf("c.CourseClassRepo.Find: %w", err)
	}

	if len(courseClasses) == 0 {
		return false, status.Error(codes.NotFound, "user do not join classes course")
	}

	classIDsInCourse, ok := courseClasses[database.Text(courseID)]
	if !ok || len(classIDsInCourse.Elements) == 0 {
		return false, status.Error(codes.NotFound, "user do not join classes course")
	}

	for _, classID := range classIDsInCourse.Elements {
		en, ok := mapClassIDs[classID.Int]
		if ok && en != 0 {
			return true, nil
		}
	}

	return false, status.Error(codes.NotFound, "user do not join this course")
}

func (c *CourseService) RetrieveBooks(ctx context.Context, req *pb.RetrieveBooksRequest) (*pb.RetrieveBooksResponse, error) {
	if req.CourseId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing course id")
	}

	limit := req.Limit
	if req.Limit < 1 || req.Limit > 100 {
		limit = 10
	}

	enBooks, total, err := c.BookRepo.FindWithFilter(ctx, c.DB, req.CourseId, uint32(limit), uint32((req.Page-1)*limit))
	if err != nil {
		return nil, fmt.Errorf("c.BookRepo.FindWithFilter: %w", err)
	}

	bookIDs := []string{}
	for _, v := range enBooks {
		bookIDs = append(bookIDs, v.ID.String)
	}

	books, err := c.toBooksPb(ctx, enBooks)
	if err != nil {
		return nil, fmt.Errorf("c.toBooksPb: %w", err)
	}

	return &pb.RetrieveBooksResponse{
		Books: books,
		Total: int32(total),
	}, nil
}
