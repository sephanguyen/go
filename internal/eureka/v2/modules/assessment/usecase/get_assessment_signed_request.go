package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/helper"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
	learnosity_init "github.com/manabie-com/backend/internal/golibs/learnosity/init"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

func (a *AssessmentUsecaseImpl) GetAssessmentSignedRequest(ctx context.Context, session domain.Session, hostDomain, config string) (string, error) {
	lm, err := a.LearningMaterialRepo.GetByID(ctx, session.LearningMaterialID)
	if err != nil {
		return "", errors.New("LearningMaterialRepo.GetByID", err)
	}

	var signedRequest any
	if err = database.ExecInTx(ctx, a.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		assessment, err := a.AssessmentRepo.GetOneByLMAndCourseID(ctx, tx, session.CourseID, session.LearningMaterialID)
		if err != nil && !errors.CheckErrType(errors.ErrNoRowsExisted, err) {
			return errors.New("AssessmentRepo.GetOneByLMAndCourseID", err)
		}

		now := time.Now()
		if errors.CheckErrType(errors.ErrNoRowsExisted, err) {
			assessment = &domain.Assessment{
				ID:                   idutil.ULIDNow(),
				CourseID:             session.CourseID,
				LearningMaterialID:   session.LearningMaterialID,
				LearningMaterialType: lm.Type,
			}
			err := assessment.Validate()
			if err != nil {
				return errors.New("assessment.Validate", err)
			}

			assessment.ID, err = a.AssessmentRepo.Upsert(ctx, tx, now, *assessment)
			if err != nil {
				return errors.New("AssessmentRepo.Insert", err)
			}
		}

		latestSession, err := a.AssessmentSessionRepo.GetLatestByIdentity(ctx, tx, assessment.ID, session.UserID)
		if err != nil && !errors.CheckErrType(errors.ErrNoRowsExisted, err) {
			return errors.New("AssessmentSessionRepo.GetLatestByIdentity", err)
		}

		if errors.CheckErrType(errors.ErrNoRowsExisted, err) {
			latestSession = domain.Session{
				ID:           uuid.New().String(),
				AssessmentID: assessment.ID,
				UserID:       session.UserID,
				Status:       domain.SessionStatusNone,
			}
			err = latestSession.Validate()
			if err != nil {
				return errors.New("Session.Validate", err)
			}

			if err = a.AssessmentSessionRepo.Insert(ctx, tx, now, latestSession); err != nil {
				return errors.New("AssessmentSessionRepo.Insert", err)
			}
		} else {
			dataSecurity := helper.NewLearnositySecurity(ctx, a.LearnosityConfig, "localhost", now)

			dataRequest := learnosity.Request{
				"activity_id": []string{latestSession.AssessmentID},
				"session_id":  []string{latestSession.ID},
				"user_id":     []string{session.UserID},
			}

			sessionStatuses, err := a.LearnositySessionRepo.GetSessionStatuses(ctx, dataSecurity, dataRequest)
			if err != nil {
				return errors.New("LearnositySessionRepo.GetSessionStatuses", err)
			}

			if len(sessionStatuses) > 0 && sessionStatuses[0].Status == domain.SessionStatusCompleted {
				latestSession = domain.Session{
					ID:           uuid.New().String(),
					AssessmentID: assessment.ID,
					UserID:       session.UserID,
					Status:       domain.SessionStatusNone,
				}
				err = latestSession.Validate()
				if err != nil {
					return errors.New("Session.Validate", err)
				}

				if err = a.AssessmentSessionRepo.Insert(ctx, tx, now, latestSession); err != nil {
					return errors.New("AssessmentSessionRepo.Insert", err)
				}
			}
		}

		initSecurity := helper.NewLearnositySecurity(ctx, a.LearnosityConfig, hostDomain, now)

		itemsRequest := learnosity.Request{
			"activity_template_id": session.LearningMaterialID,
			"name":                 lm.Name,
			"rendering_type":       learnosity.RenderingTypeAssess,
			"activity_id":          latestSession.AssessmentID,
			"session_id":           latestSession.ID,
			"user_id":              session.UserID,
		}

		if config != "" {
			configMap := make(map[string]any, 0)
			if err := json.Unmarshal([]byte(config), &configMap); err != nil {
				return errors.New("json.Unmarshal", err)
			}
			if configMap != nil {
				itemsRequest["config"] = configMap
			}
		}

		reqStr, err := json.Marshal(itemsRequest)
		if err != nil {
			return errors.New("json.Marshal", err)
		}

		init := learnosity_init.New(learnosity.ServiceItems, initSecurity, learnosity.RequestString(reqStr))

		signedRequest, err = init.Generate(true)
		if err != nil {
			return errors.New("init.Generate", err)
		}

		return nil
	}); err != nil {
		return "", err
	}

	return signedRequest.(string), nil
}
