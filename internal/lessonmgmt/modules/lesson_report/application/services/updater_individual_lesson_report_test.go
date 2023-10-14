package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/application/services/form_partner"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_lesson_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson_report/repositories"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdaterIndividualLessonReport_Update(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	lessonReportRepo := new(mock_repositories.MockLessonReportRepo)
	lessonReportDetailRepo := new(mock_repositories.MockLessonReportDetailRepo)
	partnerFormConfigRepo := new(mock_repositories.MockPartnerFormConfigRepo)
	lessonMemberRepo := new(mock_lesson_repositories.MockLessonMemberRepo)
	lessonReport := &domain.LessonReport{LessonReportID: "LessonReportID",
		FormConfigID: "FormConfigID",
		LessonID:     "LessonID"}
	tcs := []struct {
		name     string
		req      *domain.LessonReport
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "success",
			req: &domain.LessonReport{LessonReportID: "LessonReportID",
				FormConfigID: "FormConfigID",
				LessonID:     "LessonID"},
			setup: func(ctx context.Context) {
				lessonReportDetails := domain.LessonReportDetails{
					&domain.LessonReportDetail{LessonReportDetailID: "LessonReportDetailID", StudentID: "StudentID"},
				}
				lessonReportDetailRepo.
					On("GetDetailByLessonReportID", ctx, db, mock.Anything).
					Return(lessonReportDetails, nil).
					Once()

				mapFieldValuesOfStudent := map[string]domain.LessonReportFields{
					"StudentID": {&domain.LessonReportField{FieldID: "lesson_content", Value: &domain.AttributeValue{String: "1"}}},
				}

				partnerFormConfigRepo.
					On("GetMapStudentFieldValuesByDetailID", ctx, db, mock.Anything).
					Return(mapFieldValuesOfStudent, nil).
					Once()

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonReportRepo.
					On("Update", ctx, tx, mock.Anything).
					Return(lessonReport, nil).
					Once()

				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembers", ctx, tx, mock.Anything).
					Return(nil).
					Once()

			},
		},
		{
			name: "fail",
			req: &domain.LessonReport{LessonReportID: "LessonReportID",
				FormConfigID: "FormConfigID",
				LessonID:     "LessonID"},
			hasError: true,
			setup: func(ctx context.Context) {
				lessonReportDetailRepo.
					On("GetDetailByLessonReportID", ctx, db, mock.Anything).
					Return(make(domain.LessonReportDetails, 0), fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).
					Once()

			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: "1",
				},
			}
			ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			tc.setup(ctx)

			service := UpdaterIndividualLessonReport{
				DB:                     db,
				LessonReportRepo:       lessonReportRepo,
				LessonReportDetailRepo: lessonReportDetailRepo,
				PartnerFormConfigRepo:  partnerFormConfigRepo,
				LessonMemberRepo:       lessonMemberRepo,
			}
			formPartner := form_partner.InitFormPartner("resourceId")
			evictionPartner := &form_partner.AllPartner{FormPartner: formPartner}
			err := service.Update(evictionPartner, ctx, tc.req)

			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

		})
	}
}

func TestUpdaterIndividualLessonReport_UpdateForRenseikaAndBestco(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	lessonReportRepo := new(mock_repositories.MockLessonReportRepo)
	lessonReportDetailRepo := new(mock_repositories.MockLessonReportDetailRepo)
	partnerFormConfigRepo := new(mock_repositories.MockPartnerFormConfigRepo)
	lessonMemberRepo := new(mock_lesson_repositories.MockLessonMemberRepo)
	lessonReport := &domain.LessonReport{LessonReportID: "LessonReportID",
		FormConfigID: "FormConfigID",
		LessonID:     "LessonID"}
	tcs := []struct {
		name     string
		req      *domain.LessonReport
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "success",
			req: &domain.LessonReport{LessonReportID: "LessonReportID",
				FormConfigID: "FormConfigID",
				LessonID:     "LessonID"},
			setup: func(ctx context.Context) {
				lessonReportDetails := domain.LessonReportDetails{
					&domain.LessonReportDetail{LessonReportDetailID: "LessonReportDetailID", StudentID: "StudentID"},
				}
				lessonReportDetailRepo.
					On("GetDetailByLessonReportID", ctx, db, mock.Anything).
					Return(lessonReportDetails, nil).
					Once()

				mapFieldValuesOfStudent := map[string]domain.LessonReportFields{
					"StudentID": {&domain.LessonReportField{FieldID: "lesson_content", Value: &domain.AttributeValue{String: "1"}},
						&domain.LessonReportField{FieldID: "lesson_homework", Value: &domain.AttributeValue{String: "1"}},
					},
				}

				partnerFormConfigRepo.
					On("GetMapStudentFieldValuesByDetailID", ctx, db, mock.Anything).
					Return(mapFieldValuesOfStudent, nil).
					Once()

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonReportRepo.
					On("Update", ctx, tx, mock.Anything).
					Return(lessonReport, nil).
					Once()

				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembers", ctx, tx, mock.Anything).
					Return(nil).
					Once()

			},
		},
		{
			name: "fail",
			req: &domain.LessonReport{LessonReportID: "LessonReportID",
				FormConfigID: "FormConfigID",
				LessonID:     "LessonID"},
			hasError: true,
			setup: func(ctx context.Context) {
				lessonReportDetailRepo.
					On("GetDetailByLessonReportID", ctx, db, mock.Anything).
					Return(make(domain.LessonReportDetails, 0), fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).
					Once()

			},
		},
		{
			name: "should update success for multi field",
			req: &domain.LessonReport{LessonReportID: "LessonReportID",
				FormConfigID: "FormConfigID",
				LessonID:     "LessonID"},
			setup: func(ctx context.Context) {
				lessonReportDetails := domain.LessonReportDetails{
					&domain.LessonReportDetail{LessonReportDetailID: "LessonReportDetailID", StudentID: "StudentID"},
				}
				lessonReportDetailRepo.
					On("GetDetailByLessonReportID", ctx, db, mock.Anything).
					Return(lessonReportDetails, nil).
					Once()

				mapFieldValuesOfStudent := map[string]domain.LessonReportFields{
					"StudentID": {&domain.LessonReportField{FieldID: "1", Value: &domain.AttributeValue{String: "1"}}},
				}

				partnerFormConfigRepo.
					On("GetMapStudentFieldValuesByDetailID", ctx, db, mock.Anything).
					Return(mapFieldValuesOfStudent, nil).
					Once()

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonReportRepo.
					On("Update", ctx, tx, mock.Anything).
					Return(lessonReport, nil).
					Once()

				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembers", ctx, tx, mock.Anything).
					Return(nil).
					Once()

			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: "1",
				},
			}
			ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			tc.setup(ctx)

			service := UpdaterIndividualLessonReport{
				DB:                     db,
				LessonReportRepo:       lessonReportRepo,
				LessonReportDetailRepo: lessonReportDetailRepo,
				PartnerFormConfigRepo:  partnerFormConfigRepo,
				LessonMemberRepo:       lessonMemberRepo,
			}
			formPartner := form_partner.InitFormPartner("resourceId")
			evictionPartner := &form_partner.ReseikaiAndBestcoPartner{FormPartner: formPartner}
			err := service.Update(evictionPartner, ctx, tc.req)

			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

		})
	}
}
