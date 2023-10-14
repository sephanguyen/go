package question_test

import (
	"container/ring"
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/eureka/services"
	"github.com/manabie-com/backend/internal/eureka/services/question"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/s3"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_services "github.com/manabie-com/backend/mock/eureka/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestService_UpsertQuestionGroup(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	qgRepo := &mock_repositories.MockQuestionGroupRepo{}
	quizSetRepo := &mock_repositories.MockQuizSetRepo{}
	quizRepo := &mock_repositories.MockQuizRepo{}
	yasuoUploadModifierService := &mock_services.YasuoUploadModifierServiceClient{}
	yasuoUploadReaderService := &mock_services.YasuoUploadReaderServiceClient{}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupSchoolAdmin)

	url, _ := s3.GenerateUploadURL("", "", "rendered rich text")
	emptyStringUrl, _ := s3.GenerateUploadURL("", "", "")
	updatedUrl, _ := s3.GenerateUploadURL("", "", "rendered rich text updated")

	currentGroup := []pgtype.JSONB{
		database.JSONB(entities.QuestionHierarchyObj{
			ID:          "current-question-id-0",
			Type:        entities.QuestionHierarchyQuestion,
			ChildrenIDs: nil,
		}),
		database.JSONB(entities.QuestionHierarchyObj{
			ID:          "current-question-group-id",
			Type:        entities.QuestionHierarchyQuestionGroup,
			ChildrenIDs: []string{"current-question-id-1", "current-question-id-2"},
		}),
	}

	questionHierarchy := pgtype.JSONBArray{}
	err := questionHierarchy.Set(currentGroup)
	assert.NoError(t, err)

	tcs := []struct {
		ctx      context.Context
		name     string
		req      *sspb.UpsertQuestionGroupRequest
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			ctx:  ctx,
			name: "update successfully",
			req: &sspb.UpsertQuestionGroupRequest{
				QuestionGroupId:    "id",
				LearningMaterialId: "lo-id",
				Name:               "this is name",
				Description:        "this is Description",
				RichDescription: &cpb.RichText{
					Raw:      "raw rich text",
					Rendered: "rendered rich text",
				},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
					Content: "rendered rich text",
				}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
					Url: url,
				}, nil)
				tx.On("Commit", ctx).Once().Return(nil)
				qgRepo.On(
					"Upsert",
					ctx,
					tx,
					mock.Anything,
				).
					Run(func(args mock.Arguments) {
						gr := args.Get(2).(*entities.QuestionGroup)
						assert.Equal(t, "id", gr.QuestionGroupID.String)
						assert.Equal(t, "lo-id", gr.LearningMaterialID.String)
						assert.Equal(t, "this is name", gr.Name.String)
						assert.Equal(t, "this is Description", gr.Description.String)
						var richDescription entities.RichText
						err := gr.RichDescription.AssignTo(&richDescription)
						assert.NoError(t, err)
						assert.Equal(t, "raw rich text", richDescription.Raw)
						assert.Equal(t, url, richDescription.RenderedURL)
					}).
					Return(int64(1), nil).
					Once()
				qgRepo.On("FindByID", mock.Anything, mockDB, "id").Once().Return(&entities.QuestionGroup{
					QuestionGroupID: database.Text("id"),
					RichDescription: database.JSONB(&entities.RichText{
						Raw:         "raw rich text",
						RenderedURL: url,
					}),
				}, nil)
			},
			hasError: false,
		},
		{
			ctx:  ctx,
			name: "insert successfully when quiz set already have quiz ids and question ids",
			req: &sspb.UpsertQuestionGroupRequest{
				LearningMaterialId: "lo-id",
				Name:               "this is name",
				Description:        "this is Description",
				RichDescription: &cpb.RichText{
					Raw:      "raw rich text",
					Rendered: "rendered rich text",
				},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", ctx).Once().Return(nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
					Content: "rendered rich text",
				}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
					Url: url,
				}, nil)
				qgRepo.On(
					"Upsert",
					ctx,
					tx,
					mock.Anything,
				).
					Run(func(args mock.Arguments) {
						gr := args.Get(2).(*entities.QuestionGroup)
						assert.Empty(t, gr.QuestionGroupID.String)
						assert.Equal(t, "lo-id", gr.LearningMaterialID.String)
						assert.Equal(t, "this is name", gr.Name.String)
						assert.Equal(t, "this is Description", gr.Description.String)
						args.Get(2).(*entities.QuestionGroup).QuestionGroupID = database.Text("new-id")
						var richDescription entities.RichText
						err := gr.RichDescription.AssignTo(&richDescription)
						assert.NoError(t, err)
						assert.Equal(t, "raw rich text", richDescription.Raw)
						assert.Equal(t, url, richDescription.RenderedURL)
					}).
					Once().
					Return(int64(1), nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).
					Once().
					Return(entities.QuizSets{
						{
							ID:                database.Text("quiz-set-id"),
							QuizExternalIDs:   database.TextArray([]string{"quiz_id_0"}),
							QuestionHierarchy: questionHierarchy,
						},
					}, nil)
				quizSetRepo.On("Delete", ctx, tx, database.Text("quiz-set-id")).
					Once().Return(nil)
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Equal(t, pgtype.Present, quizSet.QuizExternalIDs.Status)
						assert.ElementsMatch(t, []string{"quiz_id_0"}, database.FromTextArray(quizSet.QuizExternalIDs))

						actual := make(entities.QuestionHierarchy, 0)
						err = quizSet.QuestionHierarchy.AssignTo(&actual)
						expected := make(entities.QuestionHierarchy, 0)
						tmp := append(currentGroup, database.JSONB(entities.QuestionHierarchyObj{
							ID:          "new-id",
							Type:        entities.QuestionHierarchyQuestionGroup,
							ChildrenIDs: nil,
						}))
						tmp2 := pgtype.JSONBArray{}
						err = tmp2.Set(tmp)
						require.NoError(t, err)
						err = tmp2.AssignTo(&expected)
						require.NoError(t, err)

						assert.ElementsMatch(t, actual, expected)
					}).
					Return(nil)
			},
			hasError: false,
		},
		{
			ctx:  ctx,
			name: "insert successfully when quiz set is not exist",
			req: &sspb.UpsertQuestionGroupRequest{
				LearningMaterialId: "lo-id",
				Name:               "this is name",
				Description:        "this is Description",
				RichDescription: &cpb.RichText{
					Raw:      "raw rich text",
					Rendered: "rendered rich text",
				},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", ctx).Once().Return(nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
					Content: "rendered rich text",
				}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
					Url: url,
				}, nil)
				qgRepo.
					On(
						"Upsert",
						ctx,
						tx,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						gr := args.Get(2).(*entities.QuestionGroup)
						assert.Empty(t, gr.QuestionGroupID.String)
						assert.Equal(t, "lo-id", gr.LearningMaterialID.String)
						assert.Equal(t, "this is name", gr.Name.String)
						assert.Equal(t, "this is Description", gr.Description.String)
						args.Get(2).(*entities.QuestionGroup).QuestionGroupID = database.Text("new-id")
						var richDescription entities.RichText
						err := gr.RichDescription.AssignTo(&richDescription)
						assert.NoError(t, err)
						assert.Equal(t, "raw rich text", richDescription.Raw)
						assert.Equal(t, url, richDescription.RenderedURL)
					}).
					Once().
					Return(int64(1), nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).
					Once().
					Return(entities.QuizSets{}, pgx.ErrNoRows)
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Equal(t, pgtype.Present, quizSet.QuizExternalIDs.Status)
						assert.Len(t, quizSet.QuizExternalIDs.Elements, 0)

						actual := make(entities.QuestionHierarchy, 0)
						err = quizSet.QuestionHierarchy.AssignTo(&actual)
						expected := make(entities.QuestionHierarchy, 0)
						tmp := pgtype.JSONBArray{}
						err = tmp.Set([]pgtype.JSONB{
							database.JSONB(entities.QuestionHierarchyObj{
								ID:          "new-id",
								Type:        entities.QuestionHierarchyQuestionGroup,
								ChildrenIDs: nil,
							}),
						})
						require.NoError(t, err)
						err = tmp.AssignTo(&expected)
						require.NoError(t, err)

						assert.ElementsMatch(t, actual, expected)
					}).
					Return(nil)
			},
			hasError: false,
		},
		{
			ctx:  ctx,
			name: "insert successfully when rich description is empty",
			req: &sspb.UpsertQuestionGroupRequest{
				LearningMaterialId: "lo-id",
				Name:               "this is name",
				Description:        "this is Description",
				RichDescription:    &cpb.RichText{},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", ctx).Once().Return(nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
					Content: "",
				}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
					Url: emptyStringUrl,
				}, nil)
				qgRepo.
					On(
						"Upsert",
						ctx,
						tx,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						gr := args.Get(2).(*entities.QuestionGroup)
						assert.Empty(t, gr.QuestionGroupID.String)
						assert.Equal(t, "lo-id", gr.LearningMaterialID.String)
						assert.Equal(t, "this is name", gr.Name.String)
						assert.Equal(t, "this is Description", gr.Description.String)
						args.Get(2).(*entities.QuestionGroup).QuestionGroupID = database.Text("new-id")
						var richDescription entities.RichText
						err := gr.RichDescription.AssignTo(&richDescription)
						assert.NoError(t, err)
						assert.Equal(t, "", richDescription.Raw)
						assert.Equal(t, emptyStringUrl, richDescription.RenderedURL)
					}).
					Once().
					Return(int64(1), nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).
					Once().
					Return(entities.QuizSets{}, pgx.ErrNoRows)
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Equal(t, pgtype.Present, quizSet.QuizExternalIDs.Status)
						assert.Len(t, quizSet.QuizExternalIDs.Elements, 0)

						actual := make(entities.QuestionHierarchy, 0)
						err = quizSet.QuestionHierarchy.AssignTo(&actual)
						expected := make(entities.QuestionHierarchy, 0)
						tmp := pgtype.JSONBArray{}
						err = tmp.Set([]pgtype.JSONB{
							database.JSONB(entities.QuestionHierarchyObj{
								ID:          "new-id",
								Type:        entities.QuestionHierarchyQuestionGroup,
								ChildrenIDs: nil,
							}),
						})
						require.NoError(t, err)
						err = tmp.AssignTo(&expected)
						require.NoError(t, err)

						assert.ElementsMatch(t, actual, expected)
					}).
					Return(nil)
			},
			hasError: false,
		},
		{
			ctx:  ctx,
			name: "missing lo id in request",
			req: &sspb.UpsertQuestionGroupRequest{
				QuestionGroupId: "id",
				Name:            "this is name",
				Description:     "this is Description",
				RichDescription: &cpb.RichText{
					Raw:      "raw rich text",
					Rendered: "rendered rich text",
				},
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
		{
			ctx:  ctx,
			name: "not exist lo id",
			req: &sspb.UpsertQuestionGroupRequest{
				QuestionGroupId:    "id",
				LearningMaterialId: "lo-id",
				Name:               "this is name",
				Description:        "this is Description",
				RichDescription: &cpb.RichText{
					Raw:      "raw rich text",
					Rendered: "rendered rich text",
				},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", ctx).Once().Return(nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				qgRepo.On("FindByID", mock.Anything, mockDB, "id").Once().Return(&entities.QuestionGroup{
					QuestionGroupID: database.Text("id"),
					RichDescription: database.JSONB(&entities.RichText{
						Raw:         "raw rich text",
						RenderedURL: "rendered rich text",
					}),
				}, nil)
				qgRepo.
					On(
						"Upsert",
						ctx,
						tx,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						gr := args.Get(2).(*entities.QuestionGroup)
						assert.Equal(t, "id", gr.QuestionGroupID.String)
						assert.Equal(t, "lo-id", gr.LearningMaterialID.String)
						assert.Equal(t, "this is name", gr.Name.String)
						assert.Equal(t, "this is Description", gr.Description.String)
						var richDescription entities.RichText
						err := gr.RichDescription.AssignTo(&richDescription)
						assert.NoError(t, err)
						assert.Equal(t, "raw rich text", richDescription.Raw)
						assert.Equal(t, url, richDescription.RenderedURL)
					}).
					Once().
					Return(int64(0), fmt.Errorf("could not found lo id"))
			},
			hasError: true,
		},
		{
			ctx:  ctx,
			name: "insert successfully when rich description is null",
			req: &sspb.UpsertQuestionGroupRequest{
				LearningMaterialId: "lo-id",
				Name:               "this is name",
				Description:        "this is Description",
				RichDescription:    nil,
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", ctx).Once().Return(nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
					Content: "",
				}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
					Url: emptyStringUrl,
				}, nil)
				qgRepo.
					On(
						"Upsert",
						ctx,
						tx,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						gr := args.Get(2).(*entities.QuestionGroup)
						assert.Empty(t, gr.QuestionGroupID.String)
						assert.Equal(t, "lo-id", gr.LearningMaterialID.String)
						assert.Equal(t, "this is name", gr.Name.String)
						assert.Equal(t, "this is Description", gr.Description.String)
						args.Get(2).(*entities.QuestionGroup).QuestionGroupID = database.Text("new-id")
						var richDescription entities.RichText
						err := gr.RichDescription.AssignTo(&richDescription)
						assert.NoError(t, err)
						assert.Equal(t, "", richDescription.Raw)
						assert.Equal(t, emptyStringUrl, richDescription.RenderedURL)
					}).
					Once().
					Return(int64(1), nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).
					Once().
					Return(entities.QuizSets{}, pgx.ErrNoRows)
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Equal(t, pgtype.Present, quizSet.QuizExternalIDs.Status)
						assert.Len(t, quizSet.QuizExternalIDs.Elements, 0)

						actual := make(entities.QuestionHierarchy, 0)
						err = quizSet.QuestionHierarchy.AssignTo(&actual)
						expected := make(entities.QuestionHierarchy, 0)
						tmp := pgtype.JSONBArray{}
						err = tmp.Set([]pgtype.JSONB{
							database.JSONB(entities.QuestionHierarchyObj{
								ID:          "new-id",
								Type:        entities.QuestionHierarchyQuestionGroup,
								ChildrenIDs: nil,
							}),
						})
						require.NoError(t, err)
						err = tmp.AssignTo(&expected)
						require.NoError(t, err)

						assert.ElementsMatch(t, actual, expected)
					}).
					Return(nil)
			},
			hasError: false,
		},
		{
			ctx:  ctx,
			name: "error RetrieveUploadInfo",
			req: &sspb.UpsertQuestionGroupRequest{
				QuestionGroupId:    "id",
				LearningMaterialId: "lo-id",
				Name:               "this is name",
				Description:        "this is Description",
				RichDescription: &cpb.RichText{
					Raw:      "raw rich text",
					Rendered: "rendered rich text",
				},
			},
			setup: func(ctx context.Context) {
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(nil, fmt.Errorf("failed to retrieve"))
			},
			hasError: true,
		},
		{
			ctx:  ctx,
			name: "error FindByID",
			req: &sspb.UpsertQuestionGroupRequest{
				QuestionGroupId:    "id",
				LearningMaterialId: "lo-id",
				Name:               "this is name",
				Description:        "this is Description",
				RichDescription: &cpb.RichText{
					Raw:      "raw rich text",
					Rendered: "rendered rich text",
				},
			},
			setup: func(ctx context.Context) {
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				qgRepo.On("FindByID", mock.Anything, mockDB, "id").Once().Return(nil, pgx.ErrNoRows)
			},
			hasError: true,
		},
		{
			ctx:  ctx,
			name: "update successfully with different rich description",
			req: &sspb.UpsertQuestionGroupRequest{
				QuestionGroupId:    "id",
				LearningMaterialId: "lo-id",
				Name:               "this is name",
				Description:        "this is Description",
				RichDescription: &cpb.RichText{
					Raw:      "raw rich text updated",
					Rendered: "rendered rich text updated",
				},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
					Content: "rendered rich text updated",
				}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
					Url: updatedUrl,
				}, nil)
				tx.On("Commit", ctx).Once().Return(nil)
				qgRepo.On(
					"Upsert",
					ctx,
					tx,
					mock.Anything,
				).
					Run(func(args mock.Arguments) {
						gr := args.Get(2).(*entities.QuestionGroup)
						assert.Equal(t, "id", gr.QuestionGroupID.String)
						assert.Equal(t, "lo-id", gr.LearningMaterialID.String)
						assert.Equal(t, "this is name", gr.Name.String)
						assert.Equal(t, "this is Description", gr.Description.String)
						var richDescription entities.RichText
						err := gr.RichDescription.AssignTo(&richDescription)
						assert.NoError(t, err)
						assert.Equal(t, "raw rich text updated", richDescription.Raw)
						assert.Equal(t, updatedUrl, richDescription.RenderedURL)
					}).
					Return(int64(1), nil).
					Once()
				qgRepo.On("FindByID", mock.Anything, mockDB, "id").Once().Return(&entities.QuestionGroup{
					QuestionGroupID: database.Text("id"),
					RichDescription: database.JSONB(&entities.RichText{
						Raw:         "raw rich text",
						RenderedURL: url,
					}),
				}, nil)
			},
			hasError: false,
		},
		{
			ctx:  ctx,
			name: "error mismatch url",
			req: &sspb.UpsertQuestionGroupRequest{
				QuestionGroupId:    "id",
				LearningMaterialId: "lo-id",
				Name:               "this is name",
				Description:        "this is Description",
				RichDescription: &cpb.RichText{
					Raw:      "raw rich text updated",
					Rendered: "rendered rich text updated",
				},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
					Content: "rendered rich text updated",
				}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
					Url: url,
				}, nil)
				tx.On("Rollback", ctx).Once().Return(nil)
				qgRepo.On(
					"Upsert",
					ctx,
					tx,
					mock.Anything,
				).
					Run(func(args mock.Arguments) {
						gr := args.Get(2).(*entities.QuestionGroup)
						assert.Equal(t, "id", gr.QuestionGroupID.String)
						assert.Equal(t, "lo-id", gr.LearningMaterialID.String)
						assert.Equal(t, "this is name", gr.Name.String)
						assert.Equal(t, "this is Description", gr.Description.String)
						var richDescription entities.RichText
						err := gr.RichDescription.AssignTo(&richDescription)
						assert.NoError(t, err)
						assert.Equal(t, "raw rich text updated", richDescription.Raw)
						assert.Equal(t, updatedUrl, richDescription.RenderedURL)
					}).
					Return(int64(1), nil).
					Once()
				qgRepo.On("FindByID", mock.Anything, mockDB, "id").Once().Return(&entities.QuestionGroup{
					QuestionGroupID: database.Text("id"),
					RichDescription: database.JSONB(&entities.RichText{
						Raw:         "raw rich text",
						RenderedURL: url,
					}),
				}, nil)
			},
			hasError: true,
		},
		{
			ctx:  ctx,
			name: "error UploadHtmlContent",
			req: &sspb.UpsertQuestionGroupRequest{
				QuestionGroupId:    "id",
				LearningMaterialId: "lo-id",
				Name:               "this is name",
				Description:        "this is Description",
				RichDescription: &cpb.RichText{
					Raw:      "raw rich text updated",
					Rendered: "rendered rich text updated",
				},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
					Content: "rendered rich text updated",
				}, mock.Anything).Once().Return(nil, fmt.Errorf("cannot connect to s3"))
				tx.On("Rollback", ctx).Once().Return(nil)
				qgRepo.On(
					"Upsert",
					ctx,
					tx,
					mock.Anything,
				).
					Run(func(args mock.Arguments) {
						gr := args.Get(2).(*entities.QuestionGroup)
						assert.Equal(t, "id", gr.QuestionGroupID.String)
						assert.Equal(t, "lo-id", gr.LearningMaterialID.String)
						assert.Equal(t, "this is name", gr.Name.String)
						assert.Equal(t, "this is Description", gr.Description.String)
						var richDescription entities.RichText
						err := gr.RichDescription.AssignTo(&richDescription)
						assert.NoError(t, err)
						assert.Equal(t, "raw rich text updated", richDescription.Raw)
						assert.Equal(t, updatedUrl, richDescription.RenderedURL)
					}).
					Return(int64(1), nil).
					Once()
				qgRepo.On("FindByID", mock.Anything, mockDB, "id").Once().Return(&entities.QuestionGroup{
					QuestionGroupID: database.Text("id"),
					RichDescription: database.JSONB(&entities.RichText{
						Raw:         "raw rich text",
						RenderedURL: url,
					}),
				}, nil)
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			mdCtx := interceptors.NewIncomingContext(tc.ctx)
			tc.setup(mdCtx)
			s := &services.QuizModifierService{
				QuizSetRepo: quizSetRepo,
			}
			srv := question.NewQuestionService(mockDB, qgRepo, quizSetRepo, quizRepo, s, yasuoUploadReaderService, yasuoUploadModifierService)
			res, err := srv.UpsertQuestionGroup(mdCtx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, t, res.QuestionGroupId)
			}

			mock.AssertExpectationsForObjects(t, mockDB, tx, qgRepo, quizSetRepo)
		})
	}
}

func TestCourseModifierService_UpdateDisplayOrderOfQuizSetV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	quizSetRepo := &mock_repositories.MockQuizSetRepo{}
	qgRepo := &mock_repositories.MockQuestionGroupRepo{}
	quizRepo := &mock_repositories.MockQuizRepo{}
	yasuoUploadModifierService := &mock_services.YasuoUploadModifierServiceClient{}
	yasuoUploadReaderService := &mock_services.YasuoUploadReaderServiceClient{}

	mockDB := &mock_database.Ext{}
	mockTxer := &mock_database.Tx{}
	quizModifierSvc := &services.QuizModifierService{
		QuizSetRepo: quizSetRepo,
	}
	s := question.NewQuestionService(mockDB, qgRepo, quizSetRepo, quizRepo, quizModifierSvc, yasuoUploadReaderService, yasuoUploadModifierService)

	loID := "lo_id_1"
	quizSetID := idutil.ULIDNow()
	quizSetHavingQuizzes := &entities.QuizSet{
		ID:   database.Text(quizSetID),
		LoID: database.Text(loID),
	}
	quizSetHavingQuizzes.QuestionHierarchy.Set(entities.QuestionHierarchy{
		{
			ID:   "quiz_external_id_1",
			Type: entities.QuestionHierarchyQuestion,
		},
		{
			ID:   "quiz_external_id_2",
			Type: entities.QuestionHierarchyQuestion,
		},
		{
			ID:   "quiz_external_id_3",
			Type: entities.QuestionHierarchyQuestion,
		},
		{
			ID:   "quiz_external_id_4",
			Type: entities.QuestionHierarchyQuestion,
		},
	})

	quizSetHavingQuizzesAndQuestionGroup := &entities.QuizSet{
		ID:   database.Text(quizSetID),
		LoID: database.Text(loID),
	}

	quizSetHavingQuizzesAndQuestionGroup.QuestionHierarchy.Set(entities.QuestionHierarchy{
		{
			ID:   "quiz_external_id_1",
			Type: entities.QuestionHierarchyQuestion,
		},
		{
			ID:          "question_group_id_2",
			Type:        entities.QuestionHierarchyQuestionGroup,
			ChildrenIDs: []string{"quiz_external_id_5", "quiz_external_id_6"},
		},
		{
			ID:   "quiz_external_id_3",
			Type: entities.QuestionHierarchyQuestion,
		},
		{
			ID:   "quiz_external_id_4",
			Type: entities.QuestionHierarchyQuestion,
		},
	})

	quizSetHavingQuizzesAndQuestionGroupExpected := &entities.QuizSet{
		ID:   database.Text(quizSetID),
		LoID: database.Text(loID),
	}

	quizSetHavingQuizzesAndQuestionGroupExpected.QuestionHierarchy.Set(entities.QuestionHierarchy{
		{
			ID:   "quiz_external_id_1",
			Type: entities.QuestionHierarchyQuestion,
		},
		{
			ID:   "quiz_external_id_4",
			Type: entities.QuestionHierarchyQuestion,
		},
		{
			ID:   "quiz_external_id_3",
			Type: entities.QuestionHierarchyQuestion,
		},
		{
			ID:          "question_group_id_2",
			Type:        entities.QuestionHierarchyQuestionGroup,
			ChildrenIDs: []string{"quiz_external_id_5", "quiz_external_id_6"},
		},
	})

	questionHierarchyHavingAllQuizzes := []*sspb.QuestionHierarchy{
		{
			Id:   "quiz_external_id_1",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:   "quiz_external_id_4",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:   "quiz_external_id_3",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:   "quiz_external_id_2",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
	}

	questionHierarchyHavingQuizzesAndGroups := []*sspb.QuestionHierarchy{
		{
			Id:   "quiz_external_id_1",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:   "quiz_external_id_4",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:   "quiz_external_id_3",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:          "question_group_id_2",
			Type:        sspb.QuestionHierarchyType_QUESTION_GROUP,
			ChildrenIds: []string{"quiz_external_id_5", "quiz_external_id_6"},
		},
	}

	nonExistQuestionHierarchyHavingAllQuizzes := []*sspb.QuestionHierarchy{
		{
			Id:   "non_exists_quiz_external_id_1",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:   "quiz_external_id_4",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:   "quiz_external_id_3",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:   "quiz_external_id_2",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
	}

	duplicatedQuestionHierarchyHavingAllQuizzes := []*sspb.QuestionHierarchy{
		{
			Id:   "quiz_external_id_1",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:   "quiz_external_id_4",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:   "quiz_external_id_3",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:   "quiz_external_id_4",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
	}
	lengthMismatchQuestionHierarchyHavingAllQuizzes := []*sspb.QuestionHierarchy{
		{
			Id:   "quiz_external_id_1",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:   "quiz_external_id_4",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:   "quiz_external_id_3",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
	}
	typeMismatchQuestionHierarchyHavingAllQuizzes := []*sspb.QuestionHierarchy{
		{
			Id:   "quiz_external_id_1",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:   "quiz_external_id_4",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:   "quiz_external_id_3",
			Type: sspb.QuestionHierarchyType_QUESTION_GROUP,
		},
		{
			Id:   "quiz_external_id_2",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
	}

	wrongChildrenIdsQuestionHierarchyHavingAllQuizzes := []*sspb.QuestionHierarchy{
		{
			Id:   "quiz_external_id_1",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:   "quiz_external_id_4",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:   "quiz_external_id_3",
			Type: sspb.QuestionHierarchyType_QUESTION,
		},
		{
			Id:          "question_group_id_2",
			Type:        sspb.QuestionHierarchyType_QUESTION_GROUP,
			ChildrenIds: []string{"quiz_external_id_5", "quiz_external_id_7"},
		},
	}
	testCases := []struct {
		name        string
		req         interface{}
		setup       func(ctx context.Context)
		ctx         context.Context
		expectedErr error
	}{
		{
			name: "update display order successfully when having all questions",
			ctx:  ctx,
			req: &sspb.UpdateDisplayOrderOfQuizSetV2Request{
				LearningMaterialId: loID,
				QuestionHierarchy:  questionHierarchyHavingAllQuizzes,
			},
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(quizSetHavingQuizzes, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				quizSetRepo.On("Delete", mock.Anything, mockTxer, quizSetHavingQuizzes.ID).Once().Return(nil)
				quizSetRepo.On("Create", mock.Anything, mockTxer, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, loID, quizSet.LoID.String)
						assert.Equal(t, len(quizSet.QuestionHierarchy.Elements), 4)

						var questionHierarchy entities.QuestionHierarchy
						questionHierarchy.UnmarshalJSONBArray(quizSet.QuestionHierarchy)

						expectedQuestionHierarchy := entities.QuestionHierarchy{
							{
								ID:   "quiz_external_id_1",
								Type: entities.QuestionHierarchyQuestion,
							},
							{
								ID:   "quiz_external_id_4",
								Type: entities.QuestionHierarchyQuestion,
							},
							{
								ID:   "quiz_external_id_3",
								Type: entities.QuestionHierarchyQuestion,
							},
							{
								ID:   "quiz_external_id_2",
								Type: entities.QuestionHierarchyQuestion,
							},
						}

						assert.Equal(t, questionHierarchy, expectedQuestionHierarchy)

					}).
					Return(nil)
				mockTxer.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Err when question id not exist",
			ctx:  ctx,
			req: &sspb.UpdateDisplayOrderOfQuizSetV2Request{
				LearningMaterialId: loID,
				QuestionHierarchy:  nonExistQuestionHierarchyHavingAllQuizzes,
			},
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(quizSetHavingQuizzes, nil)
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "unable to validate req: question hierarchy's elements not matched: question id %s not exist in current question hierarchy", nonExistQuestionHierarchyHavingAllQuizzes[0].Id),
		},
		{
			name: "Err get quizset by loID",
			ctx:  ctx,
			req: &sspb.UpdateDisplayOrderOfQuizSetV2Request{
				LearningMaterialId: loID,
				QuestionHierarchy:  questionHierarchyHavingAllQuizzes,
			},
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(nil, pgx.ErrNoRows)
			},
			expectedErr: status.Error(codes.NotFound, fmt.Errorf("UpdateDisplayOrderOfQuizSetV2.QuizSetRepo.GetQuizSetByLoID: %w", pgx.ErrNoRows).Error()),
		},
		{
			name: "Err duplicate question id in req",
			ctx:  ctx,
			req: &sspb.UpdateDisplayOrderOfQuizSetV2Request{
				LearningMaterialId: loID,
				QuestionHierarchy:  duplicatedQuestionHierarchyHavingAllQuizzes,
			},
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(quizSetHavingQuizzes, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("unable to validate req: id validation failed: duplicate question id %s in new question hierarchy", duplicatedQuestionHierarchyHavingAllQuizzes[3].Id).Error()),
		},
		{
			name: "Err length mismatch",
			ctx:  ctx,
			req: &sspb.UpdateDisplayOrderOfQuizSetV2Request{
				LearningMaterialId: loID,
				QuestionHierarchy:  lengthMismatchQuestionHierarchyHavingAllQuizzes,
			},
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(quizSetHavingQuizzes, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("unable to validate req: expected question hierarchy length is %d, got %d", len(quizSetHavingQuizzes.QuestionHierarchy.Elements), len(lengthMismatchQuestionHierarchyHavingAllQuizzes)).Error()),
		},
		{
			name: "Err question type mismatch",
			ctx:  ctx,
			req: &sspb.UpdateDisplayOrderOfQuizSetV2Request{
				LearningMaterialId: loID,
				QuestionHierarchy:  typeMismatchQuestionHierarchyHavingAllQuizzes,
			},
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(quizSetHavingQuizzes, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("unable to validate req: question hierarchy's elements not matched: question type mismatch, expected QUESTION but got QUESTION_GROUP").Error()),
		},
		{
			name: "Err children ids is not equal",
			ctx:  ctx,
			req: &sspb.UpdateDisplayOrderOfQuizSetV2Request{
				LearningMaterialId: loID,
				QuestionHierarchy:  wrongChildrenIdsQuestionHierarchyHavingAllQuizzes,
			},
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(quizSetHavingQuizzesAndQuestionGroup, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("unable to validate req: question hierarchy's elements not matched: mismatch children question id in question id %s", wrongChildrenIdsQuestionHierarchyHavingAllQuizzes[3].Id).Error()),
		},
		{
			name: "update display order successfully when having questions and question group",
			ctx:  ctx,
			req: &sspb.UpdateDisplayOrderOfQuizSetV2Request{
				LearningMaterialId: loID,
				QuestionHierarchy:  questionHierarchyHavingQuizzesAndGroups,
			},
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(quizSetHavingQuizzesAndQuestionGroup, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				quizSetRepo.On("Delete", mock.Anything, mockTxer, quizSetHavingQuizzesAndQuestionGroup.ID).Once().Return(nil)
				quizSetRepo.On("Create", mock.Anything, mockTxer, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, loID, quizSet.LoID.String)
						assert.Equal(t, len(quizSet.QuestionHierarchy.Elements), 4)

						var questionHierarchy entities.QuestionHierarchy
						questionHierarchy.UnmarshalJSONBArray(quizSet.QuestionHierarchy)

						var expectedQuestionHierarchy entities.QuestionHierarchy
						expectedQuestionHierarchy.UnmarshalJSONBArray(quizSetHavingQuizzesAndQuestionGroupExpected.QuestionHierarchy)

						assert.Equal(t, questionHierarchy, expectedQuestionHierarchy)

					}).
					Return(nil)
				mockTxer.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "update display order successfully when req and quizset's question hierarchy is empty",
			ctx:  ctx,
			req: &sspb.UpdateDisplayOrderOfQuizSetV2Request{
				LearningMaterialId: loID,
			},
			setup: func(ctx context.Context) {
				quizSetID := database.Text(idutil.ULIDNow())
				quizSet := &entities.QuizSet{
					ID:   quizSetID,
					LoID: database.Text(loID),
				}
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(quizSet, nil)
				mockTxer.On("Commit", mock.Anything).Once().Return(nil)
				quizSetRepo.On("Delete", mock.Anything, mockTxer, quizSetID).Once().Return(nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				quizSetRepo.On("Create", mock.Anything, mockTxer, quizSet).Once().Return(nil)

			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)

			req := testCase.req.(*sspb.UpdateDisplayOrderOfQuizSetV2Request)
			_, err := s.UpdateDisplayOrderOfQuizSetV2(testCase.ctx, req)

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
			mock.AssertExpectationsForObjects(t, mockDB, mockTxer, quizSetRepo, qgRepo)
		})
	}
}

func TestCourseModifierService_DeleteQuestionGroup(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	quizSetRepo := &mock_repositories.MockQuizSetRepo{}
	qgRepo := &mock_repositories.MockQuestionGroupRepo{}
	quizRepo := &mock_repositories.MockQuizRepo{}
	yasuoUploadModifierService := &mock_services.YasuoUploadModifierServiceClient{}
	yasuoUploadReaderService := &mock_services.YasuoUploadReaderServiceClient{}

	mockDB := &mock_database.Ext{}
	mockTxer := &mock_database.Tx{}
	quizModifierSvc := &services.QuizModifierService{
		QuizSetRepo: quizSetRepo,
	}
	s := question.NewQuestionService(mockDB, qgRepo, quizSetRepo, quizRepo, quizModifierSvc, yasuoUploadReaderService, yasuoUploadModifierService)

	questionGroupID := "group-id"
	loID := "lo-id"
	questionGroup := &entities.QuestionGroup{
		QuestionGroupID:    database.Text(questionGroupID),
		LearningMaterialID: database.Text(loID),
	}

	quizSet := &entities.QuizSet{
		ID:              database.Text("quizset-id"),
		QuizExternalIDs: database.TextArray([]string{"quiz-id-1", "quiz-id-2", "quiz-id-3"}),
		QuestionHierarchy: database.JSONBArray([]interface{}{
			&entities.QuestionHierarchyObj{
				ID:   "quiz-id-1",
				Type: entities.QuestionHierarchyQuestion,
			},
			&entities.QuestionHierarchyObj{
				ID:   "quiz-id-2",
				Type: entities.QuestionHierarchyQuestion,
			},
			&entities.QuestionHierarchyObj{
				ID:          "group-id",
				Type:        entities.QuestionHierarchyQuestionGroup,
				ChildrenIDs: []string{"quiz-id-3"},
			},
		}),
	}

	expectedQuizSet := &entities.QuizSet{
		ID:              database.Text("quizset-id"),
		QuizExternalIDs: database.TextArray([]string{"quiz-id-1", "quiz-id-2"}),
		QuestionHierarchy: database.JSONBArray([]interface{}{
			&entities.QuestionHierarchyObj{
				ID:   "quiz-id-1",
				Type: entities.QuestionHierarchyQuestion,
			},
			&entities.QuestionHierarchyObj{
				ID:   "quiz-id-2",
				Type: entities.QuestionHierarchyQuestion,
			},
		}),
	}

	quizzes := entities.Quizzes{
		{
			ID:              database.Text("id-3"),
			ExternalID:      database.Text("quiz-id-3"),
			QuestionGroupID: database.Text("group-1"),
		},
	}

	testCases := []struct {
		name         string
		req          interface{}
		setup        func(ctx context.Context)
		ctx          context.Context
		expectedErr  error
		expectedResp interface{}
	}{
		{
			name: "error empty question group id",
			ctx:  ctx,
			req: &sspb.DeleteQuestionGroupRequest{
				QuestionGroupId: "",
			},
			setup:       func(ctx context.Context) {},
			expectedErr: status.Errorf(codes.InvalidArgument, fmt.Errorf("invalid question group id").Error()),
		},
		{
			name: "Err when FindByID",
			ctx:  ctx,
			req: &sspb.DeleteQuestionGroupRequest{
				QuestionGroupId: questionGroupID,
			},
			setup: func(ctx context.Context) {
				qgRepo.On("FindByID", mock.Anything, s.DB, questionGroupID).Once().Return(nil, pgx.ErrTxClosed)
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("GroupRepo.FindByID: %w", pgx.ErrTxClosed).Error()),
		},
		{
			name: "Err when GetQuizSetByLoID",
			ctx:  ctx,
			req: &sspb.DeleteQuestionGroupRequest{
				QuestionGroupId: questionGroupID,
			},
			setup: func(ctx context.Context) {
				qgRepo.On("FindByID", mock.Anything, s.DB, questionGroupID).Once().Return(questionGroup, nil)
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(nil, pgx.ErrTxClosed)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("QuizSetRepo.GetQuizSetByLoID: %w", pgx.ErrTxClosed).Error()),
		},
		{
			name: "Err when GetByQuestionGroupID",
			ctx:  ctx,
			req: &sspb.DeleteQuestionGroupRequest{
				QuestionGroupId: questionGroupID,
			},
			setup: func(ctx context.Context) {
				qgRepo.On("FindByID", mock.Anything, s.DB, questionGroupID).Once().Return(questionGroup, nil)
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(quizSet, nil)
				quizRepo.On("GetByQuestionGroupID", mock.Anything, s.DB, database.Text(questionGroupID)).Once().Return(entities.Quizzes{}, pgx.ErrTxClosed)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("QuizRepo.GetByQuestionGroupID: %w", pgx.ErrTxClosed).Error()),
		},
		{
			name: "Err when DeleteGroupByID",
			ctx:  ctx,
			req: &sspb.DeleteQuestionGroupRequest{
				QuestionGroupId: questionGroupID,
			},
			setup: func(ctx context.Context) {
				qgRepo.On("FindByID", mock.Anything, s.DB, questionGroupID).Once().Return(questionGroup, nil)
				qgRepo.On("DeleteByID", mock.Anything, mockTxer, database.Text(questionGroupID)).Once().Return(pgx.ErrTxClosed)

				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(quizSet, nil)
				quizRepo.On("GetByQuestionGroupID", mock.Anything, s.DB, database.Text(questionGroupID)).Once().Return(quizzes, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("QuestionGroupRepo.DeleteByID: %w", pgx.ErrTxClosed).Error()),
		},
		{
			name: "Err when DeleteQuizByQuestionGroupID",
			ctx:  ctx,
			req: &sspb.DeleteQuestionGroupRequest{
				QuestionGroupId: questionGroupID,
			},
			setup: func(ctx context.Context) {
				qgRepo.On("FindByID", mock.Anything, s.DB, questionGroupID).Once().Return(questionGroup, nil)
				qgRepo.On("DeleteByID", mock.Anything, mockTxer, database.Text(questionGroupID)).Once().Return(nil)

				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(quizSet, nil)
				quizRepo.On("GetByQuestionGroupID", mock.Anything, s.DB, database.Text(questionGroupID)).Once().Return(quizzes, nil)
				quizRepo.On("DeleteByQuestionGroupID", mock.Anything, mockTxer, database.Text(questionGroupID)).Once().Return(pgx.ErrTxClosed)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("QuizRepo.DeleteByQuestionGroupID: %w", pgx.ErrTxClosed).Error()),
		},

		{
			name: "Err when DeleteByID",
			ctx:  ctx,
			req: &sspb.DeleteQuestionGroupRequest{
				QuestionGroupId: questionGroupID,
			},
			setup: func(ctx context.Context) {
				qgRepo.On("FindByID", mock.Anything, s.DB, questionGroupID).Once().Return(questionGroup, nil)
				qgRepo.On("DeleteByID", mock.Anything, mockTxer, database.Text(questionGroupID)).Once().Return(nil)

				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(quizSet, nil)
				quizSetRepo.On("Delete", mock.Anything, mockTxer, quizSet.ID).Once().Return(puddle.ErrNotAvailable)

				quizRepo.On("GetByQuestionGroupID", mock.Anything, s.DB, database.Text(questionGroupID)).Once().Return(quizzes, nil)
				quizRepo.On("DeleteByQuestionGroupID", mock.Anything, mockTxer, database.Text(questionGroupID)).Once().Return(nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("QuizSetRepo.Delete: %w", puddle.ErrNotAvailable).Error()),
		},
		{
			name: "Err when Create quizset",
			ctx:  ctx,
			req: &sspb.DeleteQuestionGroupRequest{
				QuestionGroupId: questionGroupID,
			},
			setup: func(ctx context.Context) {
				qgRepo.On("FindByID", mock.Anything, s.DB, questionGroupID).Once().Return(questionGroup, nil)
				qgRepo.On("DeleteByID", mock.Anything, mockTxer, database.Text(questionGroupID)).Once().Return(nil)

				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(quizSet, nil)
				quizSetRepo.On("Delete", mock.Anything, mockTxer, quizSet.ID).Once().Return(nil)
				quizSetRepo.On("Create", mock.Anything, mockTxer, expectedQuizSet).Once().Return(puddle.ErrNotAvailable)

				quizRepo.On("GetByQuestionGroupID", mock.Anything, s.DB, database.Text(questionGroupID)).Once().Return(quizzes, nil)
				quizRepo.On("DeleteByQuestionGroupID", mock.Anything, mockTxer, database.Text(questionGroupID)).Once().Return(nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.Internal, puddle.ErrNotAvailable.Error()),
		},
		{
			name: "Delete question group successfully",
			ctx:  ctx,
			req: &sspb.DeleteQuestionGroupRequest{
				QuestionGroupId: questionGroupID,
			},
			setup: func(ctx context.Context) {
				qgRepo.On("FindByID", mock.Anything, s.DB, questionGroupID).Once().Return(questionGroup, nil)
				qgRepo.On("DeleteByID", mock.Anything, mockTxer, database.Text(questionGroupID)).Once().Return(nil)

				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(quizSet, nil)
				quizSetRepo.On("Delete", mock.Anything, mockTxer, quizSet.ID).Once().Return(nil)
				quizSetRepo.On("Create", mock.Anything, mockTxer, expectedQuizSet).Once().Return(nil)

				quizRepo.On("GetByQuestionGroupID", mock.Anything, s.DB, database.Text(questionGroupID)).Once().Return(quizzes, nil)
				quizRepo.On("DeleteByQuestionGroupID", mock.Anything, mockTxer, database.Text(questionGroupID)).Once().Return(nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Once().Return(nil)

			},
			expectedResp: &sspb.DeleteQuestionGroupResponse{},
		},
		{
			name: "Delete empty question group successfully",
			ctx:  ctx,
			req: &sspb.DeleteQuestionGroupRequest{
				QuestionGroupId: questionGroupID,
			},
			setup: func(ctx context.Context) {
				qgRepo.On("FindByID", mock.Anything, s.DB, questionGroupID).Once().Return(questionGroup, nil)
				qgRepo.On("DeleteByID", mock.Anything, mockTxer, database.Text(questionGroupID)).Once().Return(nil)

				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(&entities.QuizSet{
					ID:              database.Text("quizset-id"),
					QuizExternalIDs: database.TextArray([]string{}),
					QuestionHierarchy: database.JSONBArray([]interface{}{
						&entities.QuestionHierarchyObj{
							ID:          "group-id",
							Type:        entities.QuestionHierarchyQuestionGroup,
							ChildrenIDs: []string{},
						},
					}),
				}, nil)
				quizSetRepo.On("Delete", mock.Anything, mockTxer, quizSet.ID).Once().Return(nil)
				quizSetRepo.On("Create", mock.Anything, mockTxer, &entities.QuizSet{
					ID:                database.Text("quizset-id"),
					QuizExternalIDs:   database.TextArray([]string{}),
					QuestionHierarchy: database.JSONBArray([]interface{}{}),
				}).Once().Return(nil)

				quizRepo.On("GetByQuestionGroupID", mock.Anything, s.DB, database.Text(questionGroupID)).Once().Return(entities.Quizzes{}, nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Once().Return(nil)

			},
			expectedResp: &sspb.DeleteQuestionGroupResponse{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)

			req := testCase.req.(*sspb.DeleteQuestionGroupRequest)

			_, err := s.DeleteQuestionGroup(testCase.ctx, req)

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				res := testCase.expectedResp.(*sspb.DeleteQuestionGroupResponse)

				assert.Nil(t, err)
				assert.Equal(t, res, &sspb.DeleteQuestionGroupResponse{})
			}
			mock.AssertExpectationsForObjects(t, mockDB, mockTxer, quizSetRepo, qgRepo, quizRepo)
		})
	}
}

func TestService_CheckQuestionsCorrectness_OrderingQuestion(t *testing.T) {
	allType := getQuizzes(
		8,
		cpb.QuizType_QUIZ_TYPE_MCQ.String(),
		cpb.QuizType_QUIZ_TYPE_FIB.String(),
		cpb.QuizType_QUIZ_TYPE_POW.String(),
		cpb.QuizType_QUIZ_TYPE_TAD.String(),
		cpb.QuizType_QUIZ_TYPE_MIQ.String(),
		cpb.QuizType_QUIZ_TYPE_MAQ.String(),
		cpb.QuizType_QUIZ_TYPE_ORD.String(),
		cpb.QuizType_QUIZ_TYPE_ESQ.String(),
	)
	quizzesNoOrderingQuestion := allType[0:6]
	quizzesNoEssayQuestion := allType[0:7]
	tcs := []struct {
		name           string
		quizzes        entities.Quizzes
		submit         *epb.SubmitQuizAnswersRequest
		request        []*sspb.CheckQuizCorrectnessRequest
		expectedResult []*entities.QuizAnswer
		haveError      bool
	}{
		{
			name:    "submit all question's answer with ordering answer correct all order options and essay question",
			quizzes: allType,
			submit: &epb.SubmitQuizAnswersRequest{
				SetId: "shuffle-quiz-set-id",
				QuizAnswer: []*epb.QuizAnswer{
					{
						QuizId: allType[1].ExternalID.String,
						Answer: []*epb.Answer{
							{
								Format: &epb.Answer_FilledText{
									FilledText: "1",
								},
							},
							{
								Format: &epb.Answer_FilledText{
									FilledText: "2",
								},
							},
							{
								Format: &epb.Answer_FilledText{
									FilledText: "3",
								},
							},
						},
					},
					{
						QuizId: allType[2].ExternalID.String,
					},
					{
						QuizId: allType[3].ExternalID.String,
					},
					{
						QuizId: allType[4].ExternalID.String,
					},
					{
						QuizId: allType[5].ExternalID.String,
					},
					{
						QuizId: allType[6].ExternalID.String,
						Answer: []*epb.Answer{
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-1",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-2",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-3",
								},
							},
						},
					},
					{
						QuizId: allType[7].ExternalID.String,
						Answer: []*epb.Answer{
							{
								Format: &epb.Answer_FilledText{
									FilledText: "text-1",
								},
							},
							{
								Format: &epb.Answer_FilledText{
									FilledText: "text-2",
								},
							},
							{
								Format: &epb.Answer_FilledText{
									FilledText: "text-3",
								},
							},
						},
					},
				},
			},
			request: []*sspb.CheckQuizCorrectnessRequest{
				{
					QuizId: allType[1].ExternalID.String,
					Answer: []*sspb.Answer{
						{
							Format: &sspb.Answer_FilledText{
								FilledText: "1",
							},
						},
						{
							Format: &sspb.Answer_FilledText{
								FilledText: "2",
							},
						},
						{
							Format: &sspb.Answer_FilledText{
								FilledText: "3",
							},
						},
					},
				},
				{
					QuizId: allType[2].ExternalID.String,
				},
				{
					QuizId: allType[3].ExternalID.String,
				},
				{
					QuizId: allType[4].ExternalID.String,
				},
				{
					QuizId: allType[5].ExternalID.String,
				},
				{
					QuizId: allType[6].ExternalID.String,
					Answer: []*sspb.Answer{
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-1",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-2",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-3",
							},
						},
					},
				},
				{
					QuizId: allType[7].ExternalID.String,
					Answer: []*sspb.Answer{
						{
							Format: &sspb.Answer_FilledText{
								FilledText: "text-1",
							},
						},
						{
							Format: &sspb.Answer_FilledText{
								FilledText: "text-2",
							},
						},
						{
							Format: &sspb.Answer_FilledText{
								FilledText: "text-3",
							},
						},
					},
				},
			},
			expectedResult: []*entities.QuizAnswer{
				nil, nil, nil, nil, nil, nil, {
					QuizID:        "external-id-6",
					QuizType:      cpb.QuizType_QUIZ_TYPE_ORD.String(),
					SubmittedKeys: []string{"key-1", "key-2", "key-3"},
					CorrectKeys:   []string{"key-1", "key-2", "key-3"},
					Correctness:   []bool{true, true, true},
					IsAccepted:    true,
					IsAllCorrect:  true,
					Point:         uint32(allType[6].Point.Int),
				},
				{
					QuizID:     "external-id-7",
					QuizType:   cpb.QuizType_QUIZ_TYPE_ESQ.String(),
					FilledText: []string{"text-1", "text-2", "text-3"},
				},
			},
			haveError: false,
		},
		{
			name:    "submit all question's answer with ordering answer correct all order options and there is not exiting quiz",
			quizzes: allType,
			submit: &epb.SubmitQuizAnswersRequest{
				SetId: "shuffle-quiz-set-id",
				QuizAnswer: []*epb.QuizAnswer{
					{
						QuizId: allType[1].ExternalID.String,
						Answer: []*epb.Answer{
							{
								Format: &epb.Answer_FilledText{
									FilledText: "1",
								},
							},
							{
								Format: &epb.Answer_FilledText{
									FilledText: "2",
								},
							},
							{
								Format: &epb.Answer_FilledText{
									FilledText: "3",
								},
							},
						},
					},
					{
						QuizId: allType[2].ExternalID.String,
					},
					{
						QuizId: allType[3].ExternalID.String,
					},
					{
						QuizId: allType[4].ExternalID.String,
					},
					{
						QuizId: allType[5].ExternalID.String,
					},
					{
						QuizId: allType[6].ExternalID.String,
						Answer: []*epb.Answer{
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-1",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-2",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-3",
								},
							},
						},
					},
					{
						QuizId: allType[7].ExternalID.String,
						Answer: []*epb.Answer{
							{
								Format: &epb.Answer_FilledText{
									FilledText: "text-1",
								},
							},
							{
								Format: &epb.Answer_FilledText{
									FilledText: "text-2",
								},
							},
							{
								Format: &epb.Answer_FilledText{
									FilledText: "text-3",
								},
							},
						},
					},
					{
						QuizId: "not-existing-id",
						Answer: []*epb.Answer{
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-1",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-2",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-3",
								},
							},
						},
					},
				},
			},
			request: []*sspb.CheckQuizCorrectnessRequest{
				{
					QuizId: allType[1].ExternalID.String,
					Answer: []*sspb.Answer{
						{
							Format: &sspb.Answer_FilledText{
								FilledText: "1",
							},
						},
						{
							Format: &sspb.Answer_FilledText{
								FilledText: "2",
							},
						},
						{
							Format: &sspb.Answer_FilledText{
								FilledText: "3",
							},
						},
					},
				},
				{
					QuizId: allType[2].ExternalID.String,
				},
				{
					QuizId: allType[3].ExternalID.String,
				},
				{
					QuizId: allType[4].ExternalID.String,
				},
				{
					QuizId: allType[5].ExternalID.String,
				},
				{
					QuizId: allType[6].ExternalID.String,
					Answer: []*sspb.Answer{
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-1",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-2",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-3",
							},
						},
					},
				},
				{
					QuizId: allType[7].ExternalID.String,
					Answer: []*sspb.Answer{
						{
							Format: &sspb.Answer_FilledText{
								FilledText: "text-1",
							},
						},
						{
							Format: &sspb.Answer_FilledText{
								FilledText: "text-2",
							},
						},
						{
							Format: &sspb.Answer_FilledText{
								FilledText: "text-3",
							},
						},
					},
				},
				{
					QuizId: "not-existing-id",
					Answer: []*sspb.Answer{
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-1",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-2",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-3",
							},
						},
					},
				},
			},
			expectedResult: []*entities.QuizAnswer{
				nil, nil, nil, nil, nil, nil, {
					QuizID:        "external-id-6",
					QuizType:      cpb.QuizType_QUIZ_TYPE_ORD.String(),
					SubmittedKeys: []string{"key-1", "key-2", "key-3"},
					CorrectKeys:   []string{"key-1", "key-2", "key-3"},
					Correctness:   []bool{true, true, true},
					IsAccepted:    true,
					IsAllCorrect:  true,
					Point:         uint32(allType[6].Point.Int),
				},
				{
					QuizID:     "external-id-7",
					QuizType:   cpb.QuizType_QUIZ_TYPE_ESQ.String(),
					FilledText: []string{"text-1", "text-2", "text-3"},
				},
			},
			haveError: false,
		},
		{
			name:    "submit only ordering question answer that correct all order options",
			quizzes: allType,
			submit: &epb.SubmitQuizAnswersRequest{
				SetId: "shuffle-quiz-set-id",
				QuizAnswer: []*epb.QuizAnswer{
					{
						QuizId: allType[6].ExternalID.String,
						Answer: []*epb.Answer{
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-1",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-2",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-3",
								},
							},
						},
					},
				},
			},
			request: []*sspb.CheckQuizCorrectnessRequest{
				{
					QuizId: allType[6].ExternalID.String,
					Answer: []*sspb.Answer{
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-1",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-2",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-3",
							},
						},
					},
				},
			},
			expectedResult: []*entities.QuizAnswer{
				nil, nil, nil, nil, nil, nil, {
					QuizID:        "external-id-6",
					QuizType:      cpb.QuizType_QUIZ_TYPE_ORD.String(),
					SubmittedKeys: []string{"key-1", "key-2", "key-3"},
					CorrectKeys:   []string{"key-1", "key-2", "key-3"},
					Correctness:   []bool{true, true, true},
					IsAccepted:    true,
					IsAllCorrect:  true,
					Point:         uint32(allType[6].Point.Int),
				},
				{
					QuizID:     "external-id-7",
					QuizType:   cpb.QuizType_QUIZ_TYPE_ESQ.String(),
					FilledText: []string{},
				},
			},
			haveError: false,
		},
		{
			name:    "submit answer that have 2 option incorrect order and 1 option correct order",
			quizzes: quizzesNoEssayQuestion,
			submit: &epb.SubmitQuizAnswersRequest{
				SetId: "shuffle-quiz-set-id",
				QuizAnswer: []*epb.QuizAnswer{
					{
						QuizId: allType[6].ExternalID.String,
						Answer: []*epb.Answer{
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-3",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-2",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-1",
								},
							},
						},
					},
				},
			},
			request: []*sspb.CheckQuizCorrectnessRequest{
				{
					QuizId: allType[6].ExternalID.String,
					Answer: []*sspb.Answer{
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-3",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-2",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-1",
							},
						},
					},
				},
			},
			expectedResult: []*entities.QuizAnswer{
				nil, nil, nil, nil, nil, nil, {
					QuizID:        "external-id-6",
					QuizType:      cpb.QuizType_QUIZ_TYPE_ORD.String(),
					SubmittedKeys: []string{"key-3", "key-2", "key-1"},
					CorrectKeys:   []string{"key-1", "key-2", "key-3"},
					Correctness:   []bool{false, true, false},
					IsAccepted:    false,
					IsAllCorrect:  false,
					Point:         0,
				},
			},
			haveError: false,
		},
		{
			name:    "submit answer that incorrect all order options",
			quizzes: quizzesNoEssayQuestion,
			submit: &epb.SubmitQuizAnswersRequest{
				SetId: "shuffle-quiz-set-id",
				QuizAnswer: []*epb.QuizAnswer{
					{
						QuizId: allType[6].ExternalID.String,
						Answer: []*epb.Answer{
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-2",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-3",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-1",
								},
							},
						},
					},
				},
			},
			request: []*sspb.CheckQuizCorrectnessRequest{
				{
					QuizId: allType[6].ExternalID.String,
					Answer: []*sspb.Answer{
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-2",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-3",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-1",
							},
						},
					},
				},
			},
			expectedResult: []*entities.QuizAnswer{
				nil, nil, nil, nil, nil, nil, {
					QuizID:        "external-id-6",
					QuizType:      cpb.QuizType_QUIZ_TYPE_ORD.String(),
					SubmittedKeys: []string{"key-2", "key-3", "key-1"},
					CorrectKeys:   []string{"key-1", "key-2", "key-3"},
					Correctness:   []bool{false, false, false},
					IsAccepted:    false,
					IsAllCorrect:  false,
					Point:         0,
				},
			},
			haveError: false,
		},
		{
			name:    "submit answer that correct all order options but latest options is duplicated",
			quizzes: quizzesNoEssayQuestion,
			submit: &epb.SubmitQuizAnswersRequest{
				SetId: "shuffle-quiz-set-id",
				QuizAnswer: []*epb.QuizAnswer{
					{
						QuizId: allType[6].ExternalID.String,
						Answer: []*epb.Answer{
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-1",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-2",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-3",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-3",
								},
							},
						},
					},
				},
			},
			request: []*sspb.CheckQuizCorrectnessRequest{
				{
					QuizId: allType[6].ExternalID.String,
					Answer: []*sspb.Answer{
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-1",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-2",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-3",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-3",
							},
						},
					},
				},
			},
			haveError: true,
		},
		{
			name:    "submit answer that correct all order options but latest options is not exist",
			quizzes: quizzesNoEssayQuestion,
			submit: &epb.SubmitQuizAnswersRequest{
				SetId: "shuffle-quiz-set-id",
				QuizAnswer: []*epb.QuizAnswer{
					{
						QuizId: allType[6].ExternalID.String,
						Answer: []*epb.Answer{
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-1",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-2",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-3",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-4",
								},
							},
						},
					},
				},
			},
			request: []*sspb.CheckQuizCorrectnessRequest{
				{
					QuizId: allType[6].ExternalID.String,
					Answer: []*sspb.Answer{
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-1",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-2",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-3",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-4",
							},
						},
					},
				},
			},
			haveError: true,
		},
		{
			name:    "submit answer that miss order answer and have a non-existing externalID",
			quizzes: quizzesNoEssayQuestion,
			submit: &epb.SubmitQuizAnswersRequest{
				SetId: "shuffle-quiz-set-id",
				QuizAnswer: []*epb.QuizAnswer{
					{
						QuizId: "not-existing-id",
						Answer: []*epb.Answer{
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-1",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-2",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-3",
								},
							},
						},
					},
				},
			},
			request: []*sspb.CheckQuizCorrectnessRequest{
				{
					QuizId: "not-existing-id",
					Answer: []*sspb.Answer{
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-1",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-2",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-3",
							},
						},
					},
				},
			},
			haveError: true,
		},
		{
			name:    "submit empty list answer ordering option",
			quizzes: quizzesNoEssayQuestion,
			submit: &epb.SubmitQuizAnswersRequest{
				SetId: "shuffle-quiz-set-id",
				QuizAnswer: []*epb.QuizAnswer{
					{
						QuizId: allType[6].ExternalID.String,
						Answer: []*epb.Answer{},
					},
				},
			},
			request: []*sspb.CheckQuizCorrectnessRequest{
				{
					QuizId: allType[6].ExternalID.String,
					Answer: []*sspb.Answer{},
				},
			},
			haveError: true,
		},
		{
			name:    "submit empty quiz answer",
			quizzes: quizzesNoEssayQuestion,
			submit: &epb.SubmitQuizAnswersRequest{
				SetId: "shuffle-quiz-set-id",
				QuizAnswer: []*epb.QuizAnswer{
					{},
				},
			},
			request: []*sspb.CheckQuizCorrectnessRequest{
				{},
			},
			haveError: true,
		},
		{
			name:    "submit wrong type answer",
			quizzes: quizzesNoEssayQuestion,
			submit: &epb.SubmitQuizAnswersRequest{
				SetId: "shuffle-quiz-set-id",
				QuizAnswer: []*epb.QuizAnswer{
					{
						QuizId: allType[6].ExternalID.String,
						Answer: []*epb.Answer{
							{
								Format: &epb.Answer_SelectedIndex{
									SelectedIndex: 1,
								},
							},
							{
								Format: &epb.Answer_SelectedIndex{
									SelectedIndex: 2,
								},
							},
							{
								Format: &epb.Answer_SelectedIndex{
									SelectedIndex: 3,
								},
							},
						},
					},
				},
			},
			request: []*sspb.CheckQuizCorrectnessRequest{
				{
					QuizId: allType[6].ExternalID.String,
					Answer: []*sspb.Answer{
						{
							Format: &sspb.Answer_SelectedIndex{
								SelectedIndex: 1,
							},
						},
						{
							Format: &sspb.Answer_SelectedIndex{
								SelectedIndex: 2,
							},
						},
						{
							Format: &sspb.Answer_SelectedIndex{
								SelectedIndex: 3,
							},
						},
					},
				},
			},
			haveError: true,
		},

		// current question have no any ordering questions
		{
			name:    "submit answer that correct all order options but there are no both existing ordering question and its answer",
			quizzes: quizzesNoOrderingQuestion,
			submit: &epb.SubmitQuizAnswersRequest{
				SetId: "shuffle-quiz-set-id",
				QuizAnswer: []*epb.QuizAnswer{
					{
						QuizId: allType[6].ExternalID.String,
						Answer: []*epb.Answer{
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-1",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-2",
								},
							},
							{
								Format: &epb.Answer_SubmittedKey{
									SubmittedKey: "key-3",
								},
							},
						},
					},
				},
			},
			request: []*sspb.CheckQuizCorrectnessRequest{
				{
					QuizId: allType[6].ExternalID.String,
					Answer: []*sspb.Answer{
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-1",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-2",
							},
						},
						{
							Format: &sspb.Answer_SubmittedKey{
								SubmittedKey: "key-3",
							},
						},
					},
				},
			},
			expectedResult: []*entities.QuizAnswer{
				nil, nil, nil, nil, nil, nil,
			},
		},
		{
			name:    "there are no both existing ordering question and submit answer",
			quizzes: quizzesNoOrderingQuestion,
			submit: &epb.SubmitQuizAnswersRequest{
				SetId:      "shuffle-quiz-set-id",
				QuizAnswer: []*epb.QuizAnswer{},
			},
			request: []*sspb.CheckQuizCorrectnessRequest{},
			expectedResult: []*entities.QuizAnswer{
				nil, nil, nil, nil, nil, nil,
			},
		},
		{
			name:    "submit answer that correct all order options but this question is not ordering type",
			quizzes: quizzesNoOrderingQuestion,
			submit: &epb.SubmitQuizAnswersRequest{
				SetId: "shuffle-quiz-set-id",
				QuizAnswer: []*epb.QuizAnswer{
					{
						QuizId: allType[5].ExternalID.String,
						Answer: []*epb.Answer{
							{
								Format: &epb.Answer_FilledText{
									FilledText: "3213213",
								},
							},
							{
								Format: &epb.Answer_FilledText{
									FilledText: "3213214",
								},
							},
							{
								Format: &epb.Answer_FilledText{
									FilledText: "3213215",
								},
							},
						},
					},
				},
			},
			request: []*sspb.CheckQuizCorrectnessRequest{
				{
					QuizId: allType[5].ExternalID.String,
					Answer: []*sspb.Answer{
						{
							Format: &sspb.Answer_FilledText{
								FilledText: "3213213",
							},
						},
						{
							Format: &sspb.Answer_FilledText{
								FilledText: "3213214",
							},
						},
						{
							Format: &sspb.Answer_FilledText{
								FilledText: "3213215",
							},
						},
					},
				},
			},
			expectedResult: []*entities.QuizAnswer{
				nil, nil, nil, nil, nil, nil,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			s := &question.Service{}
			res1, err1 := s.CheckQuestionsCorrectness(tc.quizzes, question.WithSubmitQuizAnswersRequest(tc.submit))
			res2, err2 := s.CheckQuestionsCorrectness(tc.quizzes, question.WithCheckQuizCorrectnessRequest(tc.request))
			if tc.haveError {
				require.Error(t, err1)
				require.Error(t, err2)
			} else {
				require.NoError(t, err1)
				require.NoError(t, err2)
				require.Len(t, res1, len(tc.expectedResult))
				for i := range res1 {
					if res1[i] == nil {
						continue
					}
					assert.NotZero(t, res1[i].SubmittedAt)
					tc.expectedResult[i].SubmittedAt = res1[i].SubmittedAt
				}
				assert.Equal(t, tc.expectedResult, res1)

				require.Len(t, res2, len(tc.expectedResult))
				for i := range res2 {
					if res2[i] == nil {
						continue
					}
					assert.NotZero(t, res2[i].SubmittedAt)
					tc.expectedResult[i].SubmittedAt = res2[i].SubmittedAt
				}
				assert.Equal(t, tc.expectedResult, res2)
			}
		})
	}
}

func getQuizzes(numOfQuizzes int, kind ...string) entities.Quizzes {
	start := database.Timestamptz(timeutil.Now())
	quizzes := entities.Quizzes{}
	r := ring.New(len(kind))
	for i := 0; i < len(kind); i++ {
		r.Value = kind[i]
		r = r.Next()
	}
	ra := rand.New(rand.NewSource(99))
	for i := 0; i < numOfQuizzes; i++ {
		quiz := &entities.Quiz{
			ID:          database.Text(idutil.ULIDNow()),
			ExternalID:  database.Text(fmt.Sprintf("external-id-%d", i)),
			Country:     database.Text("COUNTRY_VN"),
			SchoolID:    database.Int4(-2147483648),
			Kind:        database.Text("QUIZ_TYPE_MCQ"),
			Question:    database.JSONB(`{"raw": "{\"blocks\":[{\"key\":\"eq20k\",\"text\":\"3213\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/b9dbb04803e7cdde2e072edfd632809f.html"}`),
			Explanation: database.JSONB(`{"raw": "{\"blocks\":[{\"key\":\"4rpf3\",\"text\":\"213213\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/ee8b6810089778c7021a70298399256c.html"}`),
			Options: database.JSONB(`[
				{"key":"key-1" , "label": "1", "configs": [], "content": {"raw": "{\"blocks\":[{\"key\":\"2so83\",\"text\":\"3213213\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html"}, "correctness": false},
				{"key":"key-2" , "label": "2", "configs": [], "content": {"raw": "{\"blocks\":[{\"key\":\"2so83\",\"text\":\"3213214\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html"}, "correctness": false},
				{"key":"key-3" , "label": "3", "configs": [], "content": {"raw": "{\"blocks\":[{\"key\":\"2so83\",\"text\":\"3213215\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html"}, "correctness": false}
				]`),
			TaggedLOs:      database.TextArray([]string{"VN10-CH-01-L-001.1"}),
			DifficultLevel: database.Int4(1),
			CreatedBy:      database.Text("QC6KC30TZWc97APf99ydhonPRct1"),
			ApprovedBy:     database.Text("QC6KC30TZWc97APf99ydhonPRct1"),
			Status:         database.Text("QUIZ_STATUS_APPROVED"),
			UpdatedAt:      start,
			CreatedAt:      start,
			DeletedAt:      pgtype.Timestamptz{},
			Point:          database.Int4(ra.Int31()),
		}
		if len(kind) != 0 {
			quiz.Kind = database.Text(fmt.Sprintf("%v", r.Value))
			r = r.Next()
		}
		quizzes = append(quizzes, quiz)
	}
	return quizzes
}
