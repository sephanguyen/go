package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AssessmentSessionService struct {
	sspb.UnimplementedAssessmentSessionServiceServer
	DB database.Ext

	AssessmentSessionRepo interface {
		GetAssessmentSessionByAssessmentIDs(ctx context.Context, db database.QueryExecer, assessmentIDs pgtype.TextArray) ([]*entities.AssessmentSession, error)
		CountByAssessment(ctx context.Context, db database.QueryExecer, assessmentIDs pgtype.TextArray) (int32, error)
	}

	AssessmentRepo interface {
		GetAssessmentByCourseAndLearningMaterial(ctx context.Context, db database.QueryExecer, courseIDs, learningMaterialIDs pgtype.TextArray) ([]*entities.Assessment, error)
	}

	UserRepo interface {
		GetUserByID(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (*entities.User, error)
		GetUsersByIDs(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray) ([]*entities.User, error)
	}

	CourseStudentAccessPathRepo interface {
		GetByLocationsAndCourse(ctx context.Context, db database.QueryExecer, locationIDs, courseIDs pgtype.TextArray) ([]*entities.CourseStudentsAccessPath, error)
	}
}

func NewAssessmentSessionService(db database.Ext) sspb.AssessmentSessionServiceServer {
	return &AssessmentSessionService{
		DB:                          db,
		AssessmentSessionRepo:       &repositories.AssessmentSessionRepo{},
		AssessmentRepo:              &repositories.AssessmentRepo{},
		UserRepo:                    &repositories.UserRepo{},
		CourseStudentAccessPathRepo: &repositories.CourseStudentAccessPathRepo{},
	}
}

func (s *AssessmentSessionService) GetAssessmentSessionsByCourseAndLM(ctx context.Context, req *sspb.GetAssessmentSessionsByCourseAndLMRequest) (*sspb.GetAssessmentSessionsByCourseAndLMResponse, error) {
	courseIDs := req.GetCourseId()
	learningMaterialIDs := req.GetLearningMaterialId()
	assessmentIDs := []string{}
	assessmentSessionIDs := []string{}
	assessmentSessionWithPagiIDs := []string{}
	offset := req.Paging.GetOffsetInteger()
	limit := req.Paging.GetLimit()
	assessmentSessionsResponse := []*sspb.GetAssessmentSessionsByCourseAndLMResponse_AssessmentSession{}
	tempUsersID := []string{}
	mapUserIDAssessmentID := make(map[string]string)
	mapAssessmentSessionIDUserName := make(map[string]string)
	assessments, err := s.AssessmentRepo.GetAssessmentByCourseAndLearningMaterial(ctx, s.DB, database.TextArray(courseIDs), database.TextArray(learningMaterialIDs))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("AssessmentRepo.GetAssessmentByCourseAndLearningMaterial: %w", err).Error())
	}
	if len(assessments) > 0 {
		for _, assessment := range assessments {
			assessmentIDs = append(assessmentIDs, assessment.ID.String)
		}
	}

	assessmentSessions, err := s.AssessmentSessionRepo.GetAssessmentSessionByAssessmentIDs(ctx, s.DB, database.TextArray(assessmentIDs))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("AssessmentSessionRepo.GetAssessmentSessionByAssessmentIDs: %w", err).Error())
	}

	// filter by locations
	isFilterByLocation := false
	listStudentInLocation := []string{}
	if len(req.GetLocationIds()) > 0 {
		isFilterByLocation = true
		locationIds := req.GetLocationIds()
		courseStudentAccessPath, err := s.CourseStudentAccessPathRepo.GetByLocationsAndCourse(ctx, s.DB, database.TextArray(locationIds), database.TextArray(courseIDs))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("CourseStudentAccessPathRepo.GetByLocationsAndCourse: %w", err).Error())
		}
		for _, row := range courseStudentAccessPath {
			listStudentInLocation = append(listStudentInLocation, row.StudentID.String)
		}
	}

	if len(assessmentSessions) > 0 {
		for _, assessmentSession := range assessmentSessions {
			userID := assessmentSession.UserID.String
			if !slices.Contains(tempUsersID, userID) {
				if isFilterByLocation {
					if slices.Contains(listStudentInLocation, userID) {
						tempUsersID = append(tempUsersID, userID)
						mapUserIDAssessmentID[userID] = assessmentSession.SessionID.String
					}
				} else {
					tempUsersID = append(tempUsersID, userID)
					mapUserIDAssessmentID[userID] = assessmentSession.SessionID.String
				}
			}
		}
	}

	// order by Last Name asc, First Name asc
	users, err := s.UserRepo.GetUsersByIDs(ctx, s.DB, database.TextArray(tempUsersID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("UserRepo.GetUsersByIDs: %w", err).Error())
	}
	for _, user := range users {
		assessmentSessionIDs = append(assessmentSessionIDs, mapUserIDAssessmentID[user.UserID.String])
		mapAssessmentSessionIDUserName[mapUserIDAssessmentID[user.UserID.String]] = user.Name.String
	}

	if len(assessmentSessionIDs) >= int(offset)+int(limit) {
		assessmentSessionWithPagiIDs = assessmentSessionIDs[offset : offset+int64(limit)]
	} else if len(assessmentSessionIDs) >= int(offset) {
		assessmentSessionWithPagiIDs = assessmentSessionIDs[offset:]
	}

	for _, assessmentSessionID := range assessmentSessionWithPagiIDs {
		for _, assessmentSession := range assessmentSessions {
			if assessmentSession.SessionID.String != assessmentSessionID {
				continue
			}

			assessmentSessionsResponse = append(assessmentSessionsResponse, &sspb.GetAssessmentSessionsByCourseAndLMResponse_AssessmentSession{
				SessionId:    assessmentSession.SessionID.String,
				AssessmentId: assessmentSession.AssessmentID.String,
				UserId:       assessmentSession.UserID.String,
				UserName:     mapAssessmentSessionIDUserName[assessmentSession.SessionID.String],
			})
		}
	}

	return &sspb.GetAssessmentSessionsByCourseAndLMResponse{
		AssessmentSessions: assessmentSessionsResponse,
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: req.Paging.GetOffsetInteger() + int64(req.Paging.Limit),
			},
		},
		TotalItems: int32(len(assessmentSessionIDs)),
	}, nil
}
