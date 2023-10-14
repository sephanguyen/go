package services

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type LoReaderService struct {
	DB  database.Ext
	Env string
	*CourseReaderService
	LearningObjectiveRepo interface {
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.LearningObjective, error)
		RetrieveByTopicIDs(ctx context.Context, db database.QueryExecer, topicIds pgtype.TextArray) ([]*entities.LearningObjective, error)
	}
	QuizSetRepo interface {
		CountQuizOnLO(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (map[string]int32, error)
	}
	StudentsLearningObjectivesCompletenessRepo interface {
		Find(ctx context.Context, db database.QueryExecer, studentId pgtype.Text, loIds pgtype.TextArray) (map[pgtype.Text]*entities.StudentsLearningObjectivesCompleteness, error)
	}
}

func NewLOReaderService(
	db database.Ext,
	courseReaderSvc *CourseReaderService,
) *LoReaderService {
	return &LoReaderService{
		DB:                    db,
		CourseReaderService:   courseReaderSvc,
		LearningObjectiveRepo: &repositories.LearningObjectiveRepo{},
		QuizSetRepo:           &repositories.QuizSetRepo{},
		StudentsLearningObjectivesCompletenessRepo: &repositories.StudentsLearningObjectivesCompletenessRepo{},
	}
}

func (s *LoReaderService) RetrieveLOs(ctx context.Context, req *pb.RetrieveLOsRequest) (*pb.RetrieveLOsResponse, error) {
	var (
		learningObjectives []*entities.LearningObjective
		err                error
	)

	switch {
	case len(req.TopicIds) > 0:
		learningObjectives, err = s.LearningObjectiveRepo.RetrieveByTopicIDs(ctx, s.DB, database.TextArray(req.TopicIds))
	case len(req.LoIds) > 0:
		learningObjectives, err = s.LearningObjectiveRepo.RetrieveByIDs(ctx, s.DB, database.TextArray(req.LoIds))
	default:
		return &pb.RetrieveLOsResponse{}, nil
	}
	if err != nil {
		return nil, err
	}

	var loIDs []string
	if len(req.LoIds) > 0 {
		loIDs = req.LoIds
	} else {
		loIDs = make([]string, 0, len(learningObjectives))
		for _, lo := range learningObjectives {
			loIDs = append(loIDs, lo.ID.String)
		}
	}
	pgLOs := database.TextArray(loIDs)

	resp := new(pb.RetrieveLOsResponse)
	resp.TotalQuestions, err = s.QuizSetRepo.CountQuizOnLO(ctx, s.DB, pgLOs)
	if err != nil {
		return nil, err
	}

	loCompletenesses := make(map[pgtype.Text]*entities.StudentsLearningObjectivesCompleteness)

	if req.WithCompleteness || req.WithAchievementCrown {
		if req.StudentId == "" {
			req.StudentId = interceptors.UserIDFromContext(ctx)
		}

		loCompletenesses, err = s.StudentsLearningObjectivesCompletenessRepo.Find(ctx, s.DB, database.Text(req.StudentId), pgLOs)
		if err != nil {
			return nil, err
		}
	}

	resp.LearningObjectives = make([]*cpb.LearningObjective, 0, len(learningObjectives))
	for _, lo := range learningObjectives {
		var (
			prerequisites  []string
			maximumAttempt *wrapperspb.Int32Value
		)

		if lo.MaximumAttempt.Status == pgtype.Present {
			maximumAttempt = wrapperspb.Int32(lo.MaximumAttempt.Int)
		}

		_ = lo.Prerequisites.AssignTo(&prerequisites)
		resp.LearningObjectives = append(resp.LearningObjectives, &cpb.LearningObjective{
			Info: &cpb.ContentBasicInfo{
				Id:           lo.ID.String,
				Name:         lo.Name.String,
				Country:      cpb.Country(cpb.Country_value[lo.Country.String]),
				Subject:      cpb.Subject(cpb.Subject_value[lo.Subject.String]),
				Grade:        int32(lo.Grade.Int),
				SchoolId:     lo.SchoolID.Int,
				DisplayOrder: int32(lo.DisplayOrder.Int),
				MasterId:     lo.MasterLoID.String,
				CreatedAt:    timestamppb.New(lo.CreatedAt.Time),
				UpdatedAt:    timestamppb.New(lo.UpdatedAt.Time),
			},
			TopicId:        lo.TopicID.String,
			Video:          lo.Video.String,
			StudyGuide:     lo.StudyGuide.String,
			Prerequisites:  prerequisites,
			Type:           cpb.LearningObjectiveType(cpb.LearningObjectiveType_value[lo.Type.String]),
			Instruction:    lo.Instruction.String,
			GradeToPass:    wrapperspb.Int32(lo.GradeToPass.Int),
			ManualGrading:  lo.ManualGrading.Bool,
			TimeLimit:      wrapperspb.Int32(lo.TimeLimit.Int),
			MaximumAttempt: maximumAttempt,
			ApproveGrading: lo.ApproveGrading.Bool,
			GradeCapping:   lo.GradeCapping.Bool,
			ReviewOption:   cpb.ExamLOReviewOption(cpb.ExamLOReviewOption_value[lo.ReviewOption.String]),
			VendorType:     cpb.LearningMaterialVendorType(cpb.LearningMaterialVendorType_value[lo.VendorType.String]),
		})

		if req.WithCompleteness {
			if loCompleteness, ok := loCompletenesses[lo.ID]; ok {
				resp.Completenesses = append(resp.Completenesses, &cpb.Completenes{
					QuizFinished:         loCompleteness.IsFinishedQuiz.Bool,
					VideoFinished:        loCompleteness.IsFinishedVideo.Bool,
					StudyGuideFinished:   loCompleteness.IsFinishedStudyGuide.Bool,
					FirstQuizCorrectness: loCompleteness.FirstQuizCorrectness.Float,
				})
			}
		}

		if req.WithAchievementCrown {
			if c, ok := loCompletenesses[lo.ID]; ok {
				resp.Crowns = append(resp.Crowns, getAchievementCrownV1(c.HighestQuizScore.Float))
			} else {
				resp.Crowns = append(resp.Crowns, cpb.AchievementCrown_ACHIEVEMENT_CROWN_NONE)
			}
		}
	}

	return resp, nil
}

func getAchievementCrownV1(score float32) cpb.AchievementCrown {
	switch {
	case score == 100:
		return cpb.AchievementCrown_ACHIEVEMENT_CROWN_GOLD
	case score >= 80:
		return cpb.AchievementCrown_ACHIEVEMENT_CROWN_SILVER
	case score >= 60:
		return cpb.AchievementCrown_ACHIEVEMENT_CROWN_BRONZE
	default:
		return cpb.AchievementCrown_ACHIEVEMENT_CROWN_NONE
	}
}
