package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/jackc/pgx/v4"
	bobRepo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	"github.com/manabie-com/backend/internal/yasuo/configurations"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	mock_bob_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_metrics "github.com/manabie-com/backend/mock/notification/infra/metrics"
	mock_tag_repositories "github.com/manabie-com/backend/mock/notification/modules/tagmgmt/repositories"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	mock_services "github.com/manabie-com/backend/mock/notification/services"
	mock_domain_services "github.com/manabie-com/backend/mock/notification/services/domain"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNotificationModifierService_UpsertNotification(t *testing.T) {
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}
	config := &configurations.Config{
		Storage: configs.StorageConfig{
			Endpoint: "endpoint",
			Bucket:   "testBucket",
		},
	}
	userID := idutil.ULIDNow()

	notification := utils.GenSampleNotification()
	sampleQuestionnaire := utils.GenSampleQuestionnaire()
	notification.EditorId = userID
	notification.TargetGroup.CourseFilter = &cpb.NotificationTargetGroup_CourseFilter{
		Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
		CourseIds: []string{"course_id_1", "course_id_2"},
	}
	notification.TargetGroup.LocationFilter = &cpb.NotificationTargetGroup_LocationFilter{
		Type:        cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
		LocationIds: []string{"location_id_1", "location_id_2"},
	}
	notification.TargetGroup.ClassFilter = &cpb.NotificationTargetGroup_ClassFilter{
		Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
		ClassIds: []string{"class_id_1", "class_id_2"},
	}
	notification.TargetGroup.GradeFilter = &cpb.NotificationTargetGroup_GradeFilter{
		Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
		GradeIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
	}
	notification.TargetGroup.SchoolFilter = &cpb.NotificationTargetGroup_SchoolFilter{
		Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
		SchoolIds: []string{"school_1", "school_2"},
	}
	notification.ScheduledAt = timestamppb.New(time.Now().Add(time.Hour))
	individualIDs := []string{"individual_1", "individual_2", "individual_3"}
	notification.ReceiverIds = append(notification.ReceiverIds, individualIDs...)
	tagIDs := []string{"tag-id-1", "tag-id-2", "tag-id-3"}

	infoNotificationRepo := &mock_repositories.MockInfoNotificationRepo{}
	infoNotificationMsgRepo := &mock_repositories.MockInfoNotificationMsgRepo{}
	userInfoNotificationRepo := &mock_repositories.MockUsersInfoNotificationRepo{}
	studentRepo := &mock_bob_repositories.MockStudentRepo{}
	studentParentRepo := &mock_bob_repositories.MockStudentParentRepo{}
	questionnaireReppo := &mock_repositories.MockQuestionnaireRepo{}
	questionnaireQuestionReppo := &mock_repositories.MockQuestionnaireQuestionRepo{}
	locationRepo := &mock_repositories.MockLocationRepo{}
	infoNotificationAccessPathRepo := &mock_repositories.MockInfoNotificationAccessPathRepo{}
	notificationDataRetentionSvc := &mock_domain_services.MockDataRetentionService{}
	notificationInternalUserRepo := &mock_repositories.MockNotificationInternalUserRepo{}
	notificationLocationFilter := &mock_repositories.MockNotificationLocationFilterRepo{}
	notificationCourseFilter := &mock_repositories.MockNotificationCourseFilterRepo{}
	notificationClassFilter := &mock_repositories.MockNotificationClassFilterRepo{}

	mockNotificationMetric := &mock_metrics.NotificationMetrics{}
	mockNotificationMetric.On("RecordNotificationCreated", mock.Anything)

	tagRepo := &mock_tag_repositories.MockTagRepo{}
	infoNotificationTagRepo := &mock_repositories.MockInfoNotificationTagRepo{}

	internalUser := &entities.NotificationInternalUser{
		UserID:   database.Text("internal-user-id"),
		IsSystem: database.Bool(true),
	}

	svc := &NotificationModifierService{
		Env:                            "local",
		DB:                             mockDB,
		StorageConfig:                  config.Storage,
		InfoNotificationRepo:           infoNotificationRepo,
		InfoNotificationMsgRepo:        infoNotificationMsgRepo,
		UserNotificationRepo:           userInfoNotificationRepo,
		QuestionnaireRepo:              questionnaireReppo,
		QuestionnaireQuestionRepo:      questionnaireQuestionReppo,
		StudentRepo:                    studentRepo,
		StudentParentRepo:              studentParentRepo,
		NotificationMetrics:            mockNotificationMetric,
		TagRepo:                        tagRepo,
		InfoNotificationTagRepo:        infoNotificationTagRepo,
		LocationRepo:                   locationRepo,
		InfoNotificationAccessPathRepo: infoNotificationAccessPathRepo,
		DataRetentionService:           notificationDataRetentionSvc,
		NotificationInternalUserRepo:   notificationInternalUserRepo,
		NotificationLocationFilterRepo: notificationLocationFilter,
		NotificationCourseFilterRepo:   notificationCourseFilter,
		NotificationClassFilterRepo:    notificationClassFilter,
	}

	locationFiltersEnts := make(entities.NotificationLocationFilters, 0)
	for _, locationID := range notification.TargetGroup.LocationFilter.LocationIds {
		locationFiltersEnts = append(locationFiltersEnts, &entities.NotificationLocationFilter{
			NotificationID: database.Text(notification.NotificationId),
			LocationID:     database.Text(locationID),
		})
	}
	courseFiltersEnts := make(entities.NotificationCourseFilters, 0)
	for _, courseID := range notification.TargetGroup.CourseFilter.CourseIds {
		courseFiltersEnts = append(courseFiltersEnts, &entities.NotificationCourseFilter{
			NotificationID: database.Text(notification.NotificationId),
			CourseID:       database.Text(courseID),
		})
	}
	classFiltersEnts := make(entities.NotificationClassFilters, 0)
	for _, classID := range notification.TargetGroup.ClassFilter.ClassIds {
		classFiltersEnts = append(classFiltersEnts, &entities.NotificationClassFilter{
			NotificationID: database.Text(notification.NotificationId),
			ClassID:        database.Text(classID),
		})
	}

	makeContext := func(ctx context.Context, userID string) context.Context {
		ctxWithRp := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: strconv.Itoa(constant.ManabieSchool),
				UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			},
		})
		if userID != "" {
			ctx = interceptors.ContextWithUserID(ctxWithRp, userID)
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "pkg", "manabie", "version", "1.0.0", "token", idutil.ULIDNow())
		return ctx
	}

	testCases := []struct {
		Name  string
		Req   *npb.UpsertNotificationRequest
		Err   error
		Setup func(ctx context.Context) context.Context
	}{
		{
			Name: "happy case",
			Err:  nil,
			Req: &npb.UpsertNotificationRequest{
				Notification: notification,
				TagIds:       tagIDs,
			},
			Setup: func(ctx context.Context) context.Context {
				ctx = makeContext(ctx, userID)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)

				uploader := &mock_services.Uploader{}

				uploader.On("UploadWithContext", ctx, &s3manager.UploadInput{
					Bucket:      aws.String(config.Storage.Bucket),
					Key:         aws.String("/content/" + utils.GetMD5String(notification.Message.Content.Rendered) + ".html"),
					Body:        strings.NewReader(notification.Message.Content.Rendered),
					ACL:         aws.String("public-read"),
					ContentType: aws.String("text/html; charset=UTF-8"),
				}).Once().Return(nil, nil)
				svc.Uploader = uploader
				infoNotification, _ := mappers.PbToInfoNotificationEnt(notification)
				infoNotification.ReceiverNames = database.TextArray([]string{})
				infoNotificationMsg, _ := mappers.PbToInfoNotificationMsgEnt(notification.Message)

				// set latest editor id
				_ = infoNotification.EditorID.Set(userID)
				_ = infoNotification.CreatedUserID.Set(userID)
				_ = infoNotification.Owner.Set(constant.ManabieSchool)
				url, _ := generateUploadURL(svc.StorageConfig.Endpoint, svc.StorageConfig.Bucket, notification.Message.Content.Rendered)
				infoNotificationMsg.Content = database.JSONB(&entities.RichText{
					Raw:         notification.Message.Content.Raw,
					RenderedURL: url,
				})
				notificationDataRetentionSvc.On("AssignIndividualRetentionNamesForNotification", ctx, mockDB, mock.Anything).Return(infoNotification, nil)
				infoNotificationMsgRepo.On("Upsert", ctx, mockTx, infoNotificationMsg).Once().Return(nil)
				_ = infoNotification.NotificationMsgID.Set(infoNotificationMsg.NotificationMsgID)
				infoNotificationRepo.On("Upsert", ctx, mockTx, infoNotification).Once().Return(infoNotification.NotificationID.String, nil)

				studentIDs := []string{"student_id_1", "student_id_2", "student_id_3"}
				grades := []int32{10, 11, 12}
				filter := bobRepo.FindStudentFilter{}
				filter.StudentIDs.Set(studentIDs)
				filter.GradeIDs.Set(grades)
				studentGradeMap := map[string]int{"student_id_1": 10, "student_id_2": 11, "student_id_3": 12}
				studentRepo.On("FindStudents", ctx, mockTx, filter).Once().Return(studentGradeMap, nil)

				filter.StudentIDs.Set(individualIDs)
				filter.GradeIDs.Set(nil)

				individualGradeMap := map[string]int{"individual_1": 10, "individual_2": 11, "individual_3": 12}
				studentRepo.On("FindStudents", ctx, mockTx, filter).Once().Return(individualGradeMap, nil)
				parentIDs := []string{"parent_id_1", "parent_id_2", "parent_id_3"}
				studentParentRepo.On("FindParentIDs", ctx, mockTx, mock.Anything).Once().Return(parentIDs, nil)
				userInfoNotificationRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				userInfoNotificationRepo.On("Remove", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)

				tagRepo.On("CheckTagIDsExist", ctx, mockDB, mock.Anything).Once().Return(true, nil)
				ifntTagMap := make(map[string]entities.InfoNotificationsTags)
				ifntTagMap[infoNotification.NotificationID.String] = entities.InfoNotificationsTags{
					&entities.InfoNotificationTag{
						NotificationID: infoNotification.NotificationID,
						TagID:          database.Text("tag-id-1"),
					},
				}
				infoNotificationTagRepo.On("GetByNotificationIDs", ctx, mockTx, database.TextArray([]string{infoNotification.NotificationID.String})).Once().Return(ifntTagMap, nil)
				infoNotificationTagRepo.On("BulkUpsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				infoNotificationTagRepo.On("SoftDelete", ctx, mockTx, mock.Anything).Once().Return(nil)

				locationRepo.On("GetLocationAccessPathsByIDs", ctx, mockTx, mock.Anything).Once().Return(map[string]string{
					"location_id_1": "location_id_1",
					"location_id_2": "location_id_2",
				}, nil)

				infoNotificationAccessPathRepo.On("GetByNotificationIDAndNotInLocationIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(entities.InfoNotificationAccessPaths{}, nil)

				listAccessPath := entities.InfoNotificationAccessPaths{}
				notificationInternalUserRepo.On("GetByOrgID", ctx, mockTx, strconv.Itoa(constant.ManabieSchool)).Once().Return(internalUser, nil)

				infoNotificationAccessPathRepo.On("GetByNotificationIDAndNotInLocationIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(listAccessPath, nil)
				infoNotificationAccessPathRepo.On("SoftDelete", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				infoNotificationAccessPathRepo.On("BulkUpsert", ctx, mockTx, mock.Anything).Once().Return(nil)

				notificationLocationFilter.On("SoftDeleteByNotificationID", ctx, mockTx, notification.NotificationId).Once().Return(nil)
				notificationLocationFilter.On("BulkUpsert", ctx, mockTx, locationFiltersEnts).Once().Return(nil)

				notificationCourseFilter.On("SoftDeleteByNotificationID", ctx, mockTx, notification.NotificationId).Once().Return(nil)
				notificationCourseFilter.On("BulkUpsert", ctx, mockTx, courseFiltersEnts).Once().Return(nil)

				notificationClassFilter.On("SoftDeleteByNotificationID", ctx, mockTx, notification.NotificationId).Once().Return(nil)
				notificationClassFilter.On("BulkUpsert", ctx, mockTx, classFiltersEnts).Once().Return(nil)

				return ctx
			},
		},
		{
			Name: "happy case questionnaire",
			Err:  nil,
			Req: &npb.UpsertNotificationRequest{
				Notification:  notification,
				Questionnaire: sampleQuestionnaire,
			},
			Setup: func(ctx context.Context) context.Context {
				ctx = makeContext(ctx, userID)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)

				uploader := &mock_services.Uploader{}

				uploader.On("UploadWithContext", ctx, &s3manager.UploadInput{
					Bucket:      aws.String(config.Storage.Bucket),
					Key:         aws.String("/content/" + utils.GetMD5String(notification.Message.Content.Rendered) + ".html"),
					Body:        strings.NewReader(notification.Message.Content.Rendered),
					ACL:         aws.String("public-read"),
					ContentType: aws.String("text/html; charset=UTF-8"),
				}).Once().Return(nil, nil)
				svc.Uploader = uploader
				infoNotification, _ := mappers.PbToInfoNotificationEnt(notification)
				infoNotification.ReceiverNames = database.TextArray([]string{})
				infoNotificationMsg, _ := mappers.PbToInfoNotificationMsgEnt(notification.Message)
				questionnaire, _ := mappers.PbToQuestionnaireEnt(sampleQuestionnaire)
				questionnaireQuestion, _ := mappers.PbToQuestionnaireQuestionEnts(sampleQuestionnaire)
				// set latest editor id
				_ = infoNotification.EditorID.Set(userID)
				_ = infoNotification.CreatedUserID.Set(userID)
				_ = infoNotification.Owner.Set(constant.ManabieSchool)
				url, _ := generateUploadURL(svc.StorageConfig.Endpoint, svc.StorageConfig.Bucket, notification.Message.Content.Rendered)
				infoNotificationMsg.Content = database.JSONB(&entities.RichText{
					Raw:         notification.Message.Content.Raw,
					RenderedURL: url,
				})
				notificationDataRetentionSvc.On("AssignIndividualRetentionNamesForNotification", ctx, mockDB, mock.Anything).Return(infoNotification, nil)
				questionnaireReppo.On("Upsert", ctx, mockTx, questionnaire).Once().Return(nil)

				questionnaireQuestionReppo.On("BulkForceUpsert", ctx, mockTx, questionnaireQuestion).Once().Return(nil)
				_ = infoNotification.QuestionnaireID.Set(questionnaire.QuestionnaireID)

				infoNotificationMsgRepo.On("Upsert", ctx, mockTx, infoNotificationMsg).Once().Return(nil)
				_ = infoNotification.NotificationMsgID.Set(infoNotificationMsg.NotificationMsgID)
				infoNotificationRepo.On("Upsert", ctx, mockTx, infoNotification).Once().Return(infoNotification.NotificationID.String, nil)

				studentIDs := []string{"student_id_1", "student_id_2", "student_id_3"}
				grades := []int32{10, 11, 12}
				filter := bobRepo.FindStudentFilter{}
				filter.StudentIDs.Set(studentIDs)
				filter.GradeIDs.Set(grades)
				studentGradeMap := map[string]int{"student_id_1": 10, "student_id_2": 11, "student_id_3": 12}
				studentRepo.On("FindStudents", ctx, mockTx, filter).Once().Return(studentGradeMap, nil)

				filter.StudentIDs.Set(individualIDs)
				filter.GradeIDs.Set(nil)

				individualGradeMap := map[string]int{"individual_1": 10, "individual_2": 11, "individual_3": 12}
				studentRepo.On("FindStudents", ctx, mockTx, filter).Once().Return(individualGradeMap, nil)
				parentIDs := []string{"parent_id_1", "parent_id_2", "parent_id_3"}
				studentParentRepo.On("FindParentIDs", ctx, mockTx, mock.Anything).Once().Return(parentIDs, nil)
				userInfoNotificationRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				userInfoNotificationRepo.On("Remove", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)

				tagRepo.On("CheckTagIDsExist", ctx, mockDB, mock.Anything).Once().Return(true, nil)
				ifntTagMap := make(map[string]entities.InfoNotificationsTags)
				ifntTagMap[infoNotification.NotificationID.String] = entities.InfoNotificationsTags{
					&entities.InfoNotificationTag{
						NotificationID: infoNotification.NotificationID,
						TagID:          database.Text("tag-id-1"),
					},
				}
				infoNotificationTagRepo.On("GetByNotificationIDs", ctx, mockTx, database.TextArray([]string{infoNotification.NotificationID.String})).Once().Return(ifntTagMap, nil)
				infoNotificationTagRepo.On("BulkUpsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				infoNotificationTagRepo.On("SoftDelete", ctx, mockTx, mock.Anything).Once().Return(nil)

				locationRepo.On("GetLocationAccessPathsByIDs", ctx, mockTx, mock.Anything).Once().Return(map[string]string{
					"location_id_1": "location_id_1",
					"location_id_2": "location_id_2",
				}, nil)

				infoNotificationAccessPathRepo.On("GetByNotificationIDAndNotInLocationIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(entities.InfoNotificationAccessPaths{}, nil)

				listAccessPath := entities.InfoNotificationAccessPaths{}
				notificationInternalUserRepo.On("GetByOrgID", ctx, mockTx, strconv.Itoa(constant.ManabieSchool)).Once().Return(internalUser, nil)

				infoNotificationAccessPathRepo.On("GetByNotificationIDAndNotInLocationIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(listAccessPath, nil)
				infoNotificationAccessPathRepo.On("SoftDelete", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				infoNotificationAccessPathRepo.On("BulkUpsert", ctx, mockTx, mock.Anything).Once().Return(nil)

				notificationLocationFilter.On("SoftDeleteByNotificationID", ctx, mockTx, notification.NotificationId).Once().Return(nil)
				notificationLocationFilter.On("BulkUpsert", ctx, mockTx, locationFiltersEnts).Once().Return(nil)

				notificationCourseFilter.On("SoftDeleteByNotificationID", ctx, mockTx, notification.NotificationId).Once().Return(nil)
				notificationCourseFilter.On("BulkUpsert", ctx, mockTx, courseFiltersEnts).Once().Return(nil)

				notificationClassFilter.On("SoftDeleteByNotificationID", ctx, mockTx, notification.NotificationId).Once().Return(nil)
				notificationClassFilter.On("BulkUpsert", ctx, mockTx, classFiltersEnts).Once().Return(nil)

				return ctx
			},
		},
		{
			Name: "error GetByNotificationIDs no rows",
			Err:  fmt.Errorf("rpc error: code = Internal desc = ExecInTxWithRetry: svc.attachTagsToNotification: svc.InfoNotificationTagRepo.GetByNotificationIDs: %v", pgx.ErrNoRows),
			Req: &npb.UpsertNotificationRequest{
				Notification: notification,
				TagIds:       tagIDs,
			},
			Setup: func(ctx context.Context) context.Context {
				ctx = makeContext(ctx, userID)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockTx.On("Rollback", mock.Anything).Return(pgx.ErrNoRows)

				uploader := &mock_services.Uploader{}

				uploader.On("UploadWithContext", ctx, &s3manager.UploadInput{
					Bucket:      aws.String(config.Storage.Bucket),
					Key:         aws.String("/content/" + utils.GetMD5String(notification.Message.Content.Rendered) + ".html"),
					Body:        strings.NewReader(notification.Message.Content.Rendered),
					ACL:         aws.String("public-read"),
					ContentType: aws.String("text/html; charset=UTF-8"),
				}).Once().Return(nil, nil)
				svc.Uploader = uploader
				infoNotification, _ := mappers.PbToInfoNotificationEnt(notification)
				infoNotification.ReceiverNames = database.TextArray([]string{})
				infoNotificationMsg, _ := mappers.PbToInfoNotificationMsgEnt(notification.Message)

				_ = infoNotification.EditorID.Set(userID)
				_ = infoNotification.CreatedUserID.Set(userID)
				_ = infoNotification.Owner.Set(constant.ManabieSchool)
				url, _ := generateUploadURL(svc.StorageConfig.Endpoint, svc.StorageConfig.Bucket, notification.Message.Content.Rendered)
				infoNotificationMsg.Content = database.JSONB(&entities.RichText{
					Raw:         notification.Message.Content.Raw,
					RenderedURL: url,
				})

				notificationDataRetentionSvc.On("AssignIndividualRetentionNamesForNotification", ctx, mockDB, mock.Anything).Return(infoNotification, nil)
				infoNotificationMsgRepo.On("Upsert", ctx, mockTx, infoNotificationMsg).Once().Return(nil)
				_ = infoNotification.NotificationMsgID.Set(infoNotificationMsg.NotificationMsgID)
				infoNotificationRepo.On("Upsert", ctx, mockTx, infoNotification).Once().Return(infoNotification.NotificationID.String, nil)

				tagRepo.On("CheckTagIDsExist", ctx, mockDB, mock.Anything).Once().Return(true, nil)
				ifntTagMap := make(map[string]entities.InfoNotificationsTags)
				ifntTagMap[infoNotification.NotificationID.String] = entities.InfoNotificationsTags{
					&entities.InfoNotificationTag{
						NotificationID: infoNotification.NotificationID,
						TagID:          database.Text("tag-id-1"),
					},
				}
				infoNotificationTagRepo.On("GetByNotificationIDs", ctx, mockTx, database.TextArray([]string{infoNotification.NotificationID.String})).Once().Return(ifntTagMap, pgx.ErrNoRows)
				infoNotificationTagRepo.On("BulkUpsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				infoNotificationTagRepo.On("SoftDelete", ctx, mockTx, mock.Anything).Once().Return(nil)

				locationRepo.On("GetLocationAccessPathsByIDs", ctx, mockTx, mock.Anything).Once().Return(map[string]string{
					"location_id_1": "location_id_1",
					"location_id_2": "location_id_2",
				}, nil)

				infoNotificationAccessPathRepo.On("GetByNotificationIDAndNotInLocationIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(entities.InfoNotificationAccessPaths{}, nil)

				listAccessPath := entities.InfoNotificationAccessPaths{}
				notificationInternalUserRepo.On("GetByOrgID", ctx, mockTx, strconv.Itoa(constant.ManabieSchool)).Once().Return(internalUser, nil)

				infoNotificationAccessPathRepo.On("GetByNotificationIDAndNotInLocationIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(listAccessPath, nil)
				infoNotificationAccessPathRepo.On("SoftDelete", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				infoNotificationAccessPathRepo.On("BulkUpsert", ctx, mockTx, mock.Anything).Once().Return(nil)

				notificationLocationFilter.On("SoftDeleteByNotificationID", ctx, mockTx, notification.NotificationId).Once().Return(nil)
				notificationLocationFilter.On("BulkUpsert", ctx, mockTx, locationFiltersEnts).Once().Return(nil)

				notificationCourseFilter.On("SoftDeleteByNotificationID", ctx, mockTx, notification.NotificationId).Once().Return(nil)
				notificationCourseFilter.On("BulkUpsert", ctx, mockTx, courseFiltersEnts).Once().Return(nil)

				notificationClassFilter.On("SoftDeleteByNotificationID", ctx, mockTx, notification.NotificationId).Once().Return(nil)
				notificationClassFilter.On("BulkUpsert", ctx, mockTx, classFiltersEnts).Once().Return(nil)

				return ctx
			},
		},
		{
			Name: "happy case Tag insert",
			Err:  nil,
			Req: &npb.UpsertNotificationRequest{
				Notification: notification,
				TagIds:       tagIDs,
			},
			Setup: func(ctx context.Context) context.Context {
				ctx = makeContext(ctx, userID)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockTx.On("Rollback", mock.Anything).Return(nil)

				uploader := &mock_services.Uploader{}

				uploader.On("UploadWithContext", ctx, &s3manager.UploadInput{
					Bucket:      aws.String(config.Storage.Bucket),
					Key:         aws.String("/content/" + utils.GetMD5String(notification.Message.Content.Rendered) + ".html"),
					Body:        strings.NewReader(notification.Message.Content.Rendered),
					ACL:         aws.String("public-read"),
					ContentType: aws.String("text/html; charset=UTF-8"),
				}).Once().Return(nil, nil)
				svc.Uploader = uploader
				infoNotification, _ := mappers.PbToInfoNotificationEnt(notification)
				infoNotification.ReceiverNames = database.TextArray([]string{})
				infoNotificationMsg, _ := mappers.PbToInfoNotificationMsgEnt(notification.Message)

				_ = infoNotification.EditorID.Set(userID)
				_ = infoNotification.CreatedUserID.Set(userID)
				_ = infoNotification.Owner.Set(constant.ManabieSchool)
				url, _ := generateUploadURL(svc.StorageConfig.Endpoint, svc.StorageConfig.Bucket, notification.Message.Content.Rendered)
				infoNotificationMsg.Content = database.JSONB(&entities.RichText{
					Raw:         notification.Message.Content.Raw,
					RenderedURL: url,
				})

				notificationDataRetentionSvc.On("AssignIndividualRetentionNamesForNotification", ctx, mockDB, mock.Anything).Return(infoNotification, nil)
				infoNotificationMsgRepo.On("Upsert", ctx, mockTx, infoNotificationMsg).Once().Return(nil)
				_ = infoNotification.NotificationMsgID.Set(infoNotificationMsg.NotificationMsgID)
				infoNotificationRepo.On("Upsert", ctx, mockTx, infoNotification).Once().Return(infoNotification.NotificationID.String, nil)
				infoNotificationRepo.On("Find", ctx, mockDB, mock.Anything).Once().Return(nil, nil, nil)

				tagRepo.On("CheckTagIDsExist", ctx, mockDB, mock.Anything).Once().Return(true, nil)
				ifntTagMap := make(map[string]entities.InfoNotificationsTags)
				newNotiID := idutil.ULIDNow()
				ifntTagMap[newNotiID] = entities.InfoNotificationsTags{
					&entities.InfoNotificationTag{
						NotificationID: database.Text(newNotiID),
						TagID:          database.Text("tag-id-1"),
					},
				}
				infoNotificationTagRepo.On("GetByNotificationIDs", ctx, mockTx, database.TextArray([]string{infoNotification.NotificationID.String})).Once().Return(ifntTagMap, nil)
				infoNotificationTagRepo.On("BulkUpsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				infoNotificationTagRepo.On("SoftDelete", ctx, mockTx, mock.Anything).Once().Return(nil)

				locationRepo.On("GetLocationAccessPathsByIDs", ctx, mockTx, mock.Anything).Once().Return(map[string]string{
					"location_id_1": "location_id_1",
					"location_id_2": "location_id_2",
				}, nil)

				infoNotificationAccessPathRepo.On("GetByNotificationIDAndNotInLocationIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(entities.InfoNotificationAccessPaths{}, nil)

				listAccessPath := entities.InfoNotificationAccessPaths{}
				notificationInternalUserRepo.On("GetByOrgID", ctx, mockTx, strconv.Itoa(constant.ManabieSchool)).Once().Return(internalUser, nil)

				infoNotificationAccessPathRepo.On("GetByNotificationIDAndNotInLocationIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(listAccessPath, nil)
				infoNotificationAccessPathRepo.On("SoftDelete", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				infoNotificationAccessPathRepo.On("BulkUpsert", ctx, mockTx, mock.Anything).Once().Return(nil)

				notificationLocationFilter.On("SoftDeleteByNotificationID", ctx, mockTx, notification.NotificationId).Once().Return(nil)
				notificationLocationFilter.On("BulkUpsert", ctx, mockTx, locationFiltersEnts).Once().Return(nil)

				notificationCourseFilter.On("SoftDeleteByNotificationID", ctx, mockTx, notification.NotificationId).Once().Return(nil)
				notificationCourseFilter.On("BulkUpsert", ctx, mockTx, courseFiltersEnts).Once().Return(nil)

				notificationClassFilter.On("SoftDeleteByNotificationID", ctx, mockTx, notification.NotificationId).Once().Return(nil)
				notificationClassFilter.On("BulkUpsert", ctx, mockTx, classFiltersEnts).Once().Return(nil)

				return ctx
			},
		},
		{
			Name: "case userID is empty",
			Err:  fmt.Errorf("rpc error: code = InvalidArgument desc = user_id doesn't exist in request"),
			Req: &npb.UpsertNotificationRequest{
				Notification: notification,
			},
			Setup: func(ctx context.Context) context.Context {
				ctx = makeContext(ctx, "")

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)

				uploader := &mock_services.Uploader{}

				uploader.On("UploadWithContext", ctx, &s3manager.UploadInput{
					Bucket:      aws.String(config.Storage.Bucket),
					Key:         aws.String("/content/" + utils.GetMD5String(notification.Message.Content.Rendered) + ".html"),
					Body:        strings.NewReader(notification.Message.Content.Rendered),
					ACL:         aws.String("public-read"),
					ContentType: aws.String("text/html; charset=UTF-8"),
				}).Once().Return(nil, nil)
				svc.Uploader = uploader
				infoNotification, _ := mappers.PbToInfoNotificationEnt(notification)
				infoNotification.ReceiverNames = database.TextArray([]string{})
				infoNotificationMsg, _ := mappers.PbToInfoNotificationMsgEnt(notification.Message)

				// set latest editor id
				_ = infoNotification.EditorID.Set(userID)
				_ = infoNotification.CreatedUserID.Set(userID)
				_ = infoNotification.Owner.Set(constant.ManabieSchool)
				url, _ := generateUploadURL(svc.StorageConfig.Endpoint, svc.StorageConfig.Bucket, notification.Message.Content.Rendered)
				infoNotificationMsg.Content = database.JSONB(&entities.RichText{
					Raw:         notification.Message.Content.Raw,
					RenderedURL: url,
				})
				notificationDataRetentionSvc.On("AssignIndividualRetentionNamesForNotification", ctx, mockDB, mock.Anything).Return(infoNotification, nil)
				infoNotificationMsgRepo.On("Upsert", ctx, mockTx, infoNotificationMsg).Once().Return(nil)
				_ = infoNotification.NotificationMsgID.Set(infoNotificationMsg.NotificationMsgID)
				infoNotificationRepo.On("Upsert", ctx, mockTx, infoNotification).Once().Return(infoNotification.NotificationID.String, nil)

				studentIDs := []string{"student_id_1", "student_id_2", "student_id_3"}
				grades := []int32{10, 11, 12}
				filter := bobRepo.FindStudentFilter{}
				filter.StudentIDs.Set(studentIDs)
				filter.GradeIDs.Set(grades)
				studentGradeMap := map[string]int{"student_id_1": 10, "student_id_2": 11, "student_id_3": 12}
				studentRepo.On("FindStudents", ctx, mockTx, filter).Once().Return(studentGradeMap, nil)

				filter.StudentIDs.Set(individualIDs)
				filter.GradeIDs.Set(nil)

				individualGradeMap := map[string]int{"individual_1": 10, "individual_2": 11, "individual_3": 12}
				studentRepo.On("FindStudents", ctx, mockTx, filter).Once().Return(individualGradeMap, nil)
				parentIDs := []string{"parent_id_1", "parent_id_2", "parent_id_3"}
				studentParentRepo.On("FindParentIDs", ctx, mockTx, mock.Anything).Once().Return(parentIDs, nil)
				userInfoNotificationRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				userInfoNotificationRepo.On("Remove", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)

				locationRepo.On("GetLocationAccessPathsByIDs", ctx, mockTx, mock.Anything).Once().Return(map[string]string{
					"location_id_1": "location_id_1",
					"location_id_2": "location_id_2",
				}, nil)

				infoNotificationAccessPathRepo.On("GetByNotificationIDAndNotInLocationIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(entities.InfoNotificationAccessPaths{}, nil)

				listAccessPath := entities.InfoNotificationAccessPaths{}
				notificationInternalUserRepo.On("GetByOrgID", ctx, mockTx, strconv.Itoa(constant.ManabieSchool)).Once().Return(internalUser, nil)

				infoNotificationAccessPathRepo.On("GetByNotificationIDAndNotInLocationIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(listAccessPath, nil)
				infoNotificationAccessPathRepo.On("SoftDelete", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				infoNotificationAccessPathRepo.On("BulkUpsert", ctx, mockTx, mock.Anything).Once().Return(nil)

				notificationLocationFilter.On("SoftDeleteByNotificationID", ctx, mockTx, notification.NotificationId).Once().Return(nil)
				notificationLocationFilter.On("BulkUpsert", ctx, mockTx, locationFiltersEnts).Once().Return(nil)

				notificationCourseFilter.On("SoftDeleteByNotificationID", ctx, mockTx, notification.NotificationId).Once().Return(nil)
				notificationCourseFilter.On("BulkUpsert", ctx, mockTx, courseFiltersEnts).Once().Return(nil)

				notificationClassFilter.On("SoftDeleteByNotificationID", ctx, mockTx, notification.NotificationId).Once().Return(nil)
				notificationClassFilter.On("BulkUpsert", ctx, mockTx, classFiltersEnts).Once().Return(nil)

				return ctx
			},
		},
	}

	ctx := context.Background()
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx2 := testCase.Setup(ctx)
			_, err := svc.UpsertNotification(ctx2, testCase.Req)
			if testCase.Err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testCase.Err.Error(), err.Error())
			}
		})
	}
}
