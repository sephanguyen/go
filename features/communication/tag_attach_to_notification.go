package communication

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TagAttachToNotificationSuite struct {
	*common.NotificationSuite
	dbTagMap              map[string]string
	upsertNotiReq         *npb.UpsertNotificationRequest
	createdNotiID         string
	dbAttachedTagIDs      []string
	requestAttachedTagIDs []string
	ifntTagRepo           *repositories.InfoNotificationTagRepo
	expectedError         error
}

var (
	ExpectedUpsertError = status.Error(codes.InvalidArgument, "some tags do not exist")
	SpecialCaseFlag     = -1
)

func (c *SuiteConstructor) InitTagAttachToNotification(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &TagAttachToNotificationSuite{
		NotificationSuite: dep.notiCommonSuite,
		ifntTagRepo:       &repositories.InfoNotificationTagRepo{},
		dbTagMap:          make(map[string]string),
	}
	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^a valid composed notification$`:                                                    s.aValidComposedNotification,
		`^association with "([^"]*)" tags are reenabled in database$`:                        s.associationWithTagsAreReenabledInDatabase,
		`^association with "([^"]*)" tags are soft deleted in database$`:                     s.associationWithTagsAreSoftDeletedInDatabase,
		`^"([^"]*)" attached tags data and names are correctly stored database$`:             s.attachedTagsDataAndNamesAreCorrectlyStoredDatabase,
		`^notification is attached with tags "([^"]*)" in database$`:                         s.notificationIsAttachedWithTagsInDatabase,
		`^user send upsert notification request to attach "([^"]*)" tags to notification$`:   s.userSendUpsertNotificationRequestToAttachTagsToNotification,
		`^user send upsert notification request to remove "([^"]*)" tags from notification$`: s.userSendUpsertNotificationRequestToSoftDeleteTags,
		`^school admin creates "([^"]*)" students$`:                                          s.CreatesNumberOfStudents,
		`^school admin create some tags named "([^"]*)"$`:                                    s.CreatesTagsWithNames,
		`^current staff discards notification$`:                                              s.CurrentStaffDiscardsNotification,
		`^notification is discarded$`:                                                        s.NotificationIsDiscarded,
		`^returns "([^"]*)" status code$`:                                                    s.CheckReturnStatusCode,
		`^admin archived "([^"]*)" tags$`:                                                    s.adminArchivedTags,
	}
	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *TagAttachToNotificationSuite) aValidComposedNotification(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	if len(commonState.Students) != 1 {
		return ctx, fmt.Errorf("number of student is not 1")
	}
	noti := aSampleComposedNotification([]string{commonState.Students[0].ID}, i32resourcePathFromCtx(ctx), "all", false)
	req := &npb.UpsertNotificationRequest{
		Notification: noti,
	}
	s.upsertNotiReq = req
	return ctx, nil
}

func (s *TagAttachToNotificationSuite) associationWithTagsAreReenabledInDatabase(ctx context.Context, reassignTagNamesStr string) (context.Context, error) {
	return s.checkInfoNotificationTagExist(ctx, reassignTagNamesStr, true)
}

func (s *TagAttachToNotificationSuite) associationWithTagsAreSoftDeletedInDatabase(ctx context.Context, deleteTagNamesStr string) (context.Context, error) {
	return s.checkInfoNotificationTagExist(ctx, deleteTagNamesStr, false)
}

func (s *TagAttachToNotificationSuite) checkInfoNotificationTagExist(ctx context.Context, tagNamesStr string, shouldExist bool) (context.Context, error) {
	tagIDs := s.toTagIDArray(tagNamesStr)
	query := `
	SELECT COUNT(*)
	FROM	info_notifications_tags ifnt
	WHERE   ifnt.tag_id = ANY($1::text[]) AND ifnt.notification_id = $2
	`
	deletedAt := ""
	if shouldExist {
		deletedAt = " AND ifnt.deleted_at IS NULL"
	} else {
		deletedAt = " AND ifnt.deleted_at IS NOT NULL"
	}
	query += deletedAt

	row := s.BobDBConn.QueryRow(ctx, query, database.TextArray(tagIDs), database.Text(s.createdNotiID))
	var count int
	if err := row.Scan(&count); err != nil {
		return ctx, fmt.Errorf("err scan: %w", err)
	}
	if count != len(tagIDs) {
		return ctx, fmt.Errorf("expected %d info_notifications_tags records, got %d", len(tagIDs), count)
	}
	return ctx, nil
}

func (s *TagAttachToNotificationSuite) attachedTagsDataAndNamesAreCorrectlyStoredDatabase(ctx context.Context, numTagsCorrectStoredStr string) (context.Context, error) {
	numIfntTagsExpc := s.toNumCount(numTagsCorrectStoredStr)
	if numIfntTagsExpc == SpecialCaseFlag {
		if !errors.Is(s.expectedError, ExpectedUpsertError) {
			return ctx, fmt.Errorf("expected %v to occurred, not received", ExpectedUpsertError)
		}
		return ctx, nil
	}

	res, err := s.ifntTagRepo.GetByNotificationIDs(ctx, s.BobDBConn, database.TextArray([]string{s.createdNotiID}))
	if err != nil {
		return ctx, fmt.Errorf("ifntTagRepo.GetByNotificationIDs: %v", err)
	}
	if ifntTags, ok := res[s.createdNotiID]; ok {
		countNumIfntTags := 0
		for _, ifntTag := range ifntTags {
			if !slices.Contains(s.requestAttachedTagIDs, ifntTag.TagID.String) {
				return ctx, fmt.Errorf("error attached tag is not correctly stored, unexpected info_notifications_tag ID: %v", ifntTag.NotificationTagID.String)
			}
			countNumIfntTags++
		}
		if countNumIfntTags != numIfntTagsExpc {
			return ctx, fmt.Errorf("error expected %d info_notifications_tags records, got %d", numIfntTagsExpc, countNumIfntTags)
		}
	} else if numIfntTagsExpc != 0 {
		// GetByNotificationIDs result 0 record means either:
		// 1. UpsertNotification request removed all associated Tags -> expected 0 tag results
		// 2. Error somewhere occurred
		return ctx, fmt.Errorf("error info notifications tags are not correctly stored")
	}
	return ctx, nil
}

func (s *TagAttachToNotificationSuite) notificationIsAttachedWithTagsInDatabase(ctx context.Context, attachedTagNamesStr string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	for _, tag := range commonState.Tags {
		s.dbTagMap[tag.Name] = tag.ID
	}
	tagIDs := s.toTagIDArray(attachedTagNamesStr)
	req := &npb.UpsertNotificationRequest{
		Notification: s.upsertNotiReq.Notification,
	}
	if len(tagIDs) > 0 {
		req.TagIds = tagIDs
	}
	res, err := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(
		ctx,
		req,
	)
	commonState.ResponseErr = err
	if err == nil {
		s.createdNotiID = res.NotificationId
		s.upsertNotiReq.Notification.NotificationId = res.NotificationId
		s.dbAttachedTagIDs = tagIDs
	}

	commonState.Notification = s.upsertNotiReq.Notification
	return common.StepStateToContext(ctx, commonState), nil
}

func (s *TagAttachToNotificationSuite) userSendUpsertNotificationRequestToAttachTagsToNotification(ctx context.Context, attachedTagNamesStr string) (context.Context, error) {
	tagIDs := s.toTagIDArray(attachedTagNamesStr)
	req := &npb.UpsertNotificationRequest{
		Notification: s.upsertNotiReq.Notification,
	}
	if len(tagIDs) > 0 {
		req.TagIds = tagIDs
		s.requestAttachedTagIDs = tagIDs
	}
	res, err := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(
		ctx,
		req,
	)
	if err != nil {
		// handle an expected error when upserting non-existed TagID
		if errors.Is(err, ExpectedUpsertError) {
			s.expectedError = err
			return ctx, nil
		}
		return ctx, fmt.Errorf("error attach tags to notification: %v", err)
	}
	if res.NotificationId != s.createdNotiID {
		return ctx, fmt.Errorf("expected NotiID %s, got %s", s.createdNotiID, res.NotificationId)
	}
	return ctx, nil
}

func (s *TagAttachToNotificationSuite) userSendUpsertNotificationRequestToSoftDeleteTags(ctx context.Context, deleteTagNamesStr string) (context.Context, error) {
	// original: tag1,tag2,tag3
	// delete: tag1,tag3
	// upsert request: tag2
	tagIDs := s.toTagIDArray(deleteTagNamesStr)
	_, _, isnTagIDs := golibs.Compare(s.dbAttachedTagIDs, tagIDs)
	req := &npb.UpsertNotificationRequest{
		Notification: s.upsertNotiReq.Notification,
		TagIds:       isnTagIDs,
	}
	res, err := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(
		ctx,
		req,
	)
	if err != nil {
		return ctx, fmt.Errorf("error attach tags to notification: %v", err)
	}
	if res.NotificationId != s.createdNotiID {
		return ctx, fmt.Errorf("expected NotiID %s, got %s", s.createdNotiID, res.NotificationId)
	}
	return ctx, nil
}

// util funcs
func (s *TagAttachToNotificationSuite) toTagNameArray(tagNames string) []string {
	tagNamesArr := []string{}
	if tagNames != "NULL" {
		tagNamesArr = strings.Split(tagNames, ",")
	}

	return tagNamesArr
}

func (s *TagAttachToNotificationSuite) toTagIDArray(tagNames string) []string {
	tagNameArr := s.toTagNameArray(tagNames)
	tagIDs := []string{}
	for _, tagName := range tagNameArr {
		if val, ok := s.dbTagMap[tagName]; ok {
			tagIDs = append(tagIDs, val)
		} else {
			// generate a fake TagID for case upsert request has non-existed TagID
			tagIDs = append(tagIDs, idutil.ULIDNow())
		}
	}

	return tagIDs
}

func (s *TagAttachToNotificationSuite) toNumCount(num string) int {
	i, _ := strconv.Atoi(num)
	return i
}

func (s *TagAttachToNotificationSuite) adminArchivedTags(ctx context.Context, attachedTagNamesStr string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	for _, tag := range commonState.Tags {
		s.dbTagMap[tag.Name] = tag.ID
	}
	tagIDs := s.toTagIDArray(attachedTagNamesStr)
	query := `
		UPDATE tags SET is_archived = TRUE
		WHERE tag_id = ANY($1::TEXT[]);
	`
	_, err := s.BobDBConn.Exec(ctx, query, database.TextArray(tagIDs))
	if err != nil {
		return ctx, fmt.Errorf("failed archive tag: %v", err)
	}

	return common.StepStateToContext(ctx, commonState), nil
}
