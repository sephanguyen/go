package communication

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/PuerkitoBio/goquery"
	"github.com/jackc/pgtype"
)

func (s *suite) schoolAdminSeesPeopleDisplayInNotificationListOnCMS(ctx context.Context, numReadPerTotal, _ string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	expectNumRead, _ := strconv.Atoi(strings.Split(numReadPerTotal, "/")[0])
	expectTotal, _ := strconv.Atoi(strings.Split(numReadPerTotal, "/")[1])

	query := `SELECT 
				COUNT(DISTINCT (user_id, notification_id)) FILTER (WHERE uin.status = 'USER_NOTIFICATION_STATUS_READ') AS read, 
				COUNT(*) AS total 
				FROM users_info_notifications uin 
				WHERE notification_id = $1 AND deleted_at IS NULL;`

	var numRead pgtype.Int8
	var total pgtype.Int8
	err := s.bobDB.QueryRow(ctx, query, database.Text(stepState.notification.NotificationId)).Scan(&numRead, &total)
	if err != nil {
		return ctx, err
	}

	if expectNumRead != int(numRead.Int) {
		return ctx, fmt.Errorf("expect num of read user is %d but got %d", expectNumRead, numRead.Int)
	}
	if expectTotal != int(total.Int) {
		return ctx, fmt.Errorf("expect total user is %d but got %d", expectTotal, total.Int)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) schoolAdminSeesTheStatusOfIsChangedTo(ctx context.Context, userAccountsArg, readStatus string) (context.Context, error) {
	var err error
	stepState := StepStateFromContext(ctx)
	userAccount := strings.Split(userAccountsArg, "&")
	for _, acc := range userAccount {
		acc := strings.TrimSpace(acc)
		userID := ""
		switch acc {
		case student:
			userID = stepState.profile.defaultStudent.id
		case parent:
			userID = stepState.profile.defaultParent.id
		default:
			return ctx, fmt.Errorf("not supported this type of account %s", acc)
		}
		token := s.getToken(acc)
		ctx, err = s.RetrieveNotificationDetail(ctx, stepState.notification.NotificationId, token, userID, cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_READ)
		if err != nil {
			return ctx, err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) interactsTheHyperlinkInTheContentOnLearnerApp(ctx context.Context, role string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	token := s.getToken(role)
	req := &bpb.RetrieveNotificationDetailRequest{
		NotificationId: stepState.notification.NotificationId,
	}
	resp, err := bpb.NewNotificationReaderServiceClient(s.bobConn).RetrieveNotificationDetail(contextWithToken(ctx, token), req)
	if err != nil {
		return ctx, err
	}
	urlHTML := resp.Item.Message.Content.Rendered
	return StepStateToContext(ctx, stepState), checkRenderedHTMLWithLink(urlHTML, stepState.attachedLinkInNoti)
}

//nolint:gosec
func checkRenderedHTMLWithLink(url string, expectLink string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	aTag := doc.Find("a")
	link, ok := aTag.Attr("href")
	if !ok {
		return fmt.Errorf("cannot find <a> tag in html")
	}
	if link != expectLink {
		return fmt.Errorf("expect attached link in notification is %s but got %s", expectLink, link)
	}

	return nil
}

//nolint: goconst
func (s *suite) schoolAdminHasCreatedNotificationThatContentIncludesHyperlink(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	noti, err := s.newNotification(stepState.schoolID, stepState.profile.schoolAdmin.id)
	if err != nil {
		return ctx, err
	}
	receiverIDs := make([]string, 0)
	receiverIDs = append(receiverIDs, stepState.profile.defaultStudent.id)
	noti = s.notificationWithReceiver(receiverIDs, noti)

	stepState.attachedLinkInNoti = "https://google.com"
	noti.Message.Content = &cpb.RichText{
		Rendered: fmt.Sprintf(`<a href="%s">google</a>`, stepState.attachedLinkInNoti),
		Raw:      fmt.Sprintf(`<a href="%s">google</a>`, stepState.attachedLinkInNoti),
	}
	stepState.notification = noti

	return s.upsertNotification(ctx, noti)
}

func (s *suite) schoolAdminHasSentNotificationToStudentAndParent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	return s.sendNotification(ctx, stepState.notification)
}
