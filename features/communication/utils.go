package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/try"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/bxcodec/faker/v3/support/slice"
	"github.com/google/go-cmp/cmp"
	"github.com/r3labs/diff/v3"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	LauoutTimeFormat = "2006-01-02T15:04:05Z07:00"
)

func contextWithResourcePath(ctx context.Context, rp string) context.Context {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim == nil {
		claim = &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{},
		}
	}
	claim.Manabie.ResourcePath = rp
	return interceptors.ContextWithJWTClaims(ctx, claim)
}

func contextWithUserID(ctx context.Context, userID string) context.Context {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim == nil {
		claim = &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{},
		}
	}
	claim.Manabie.UserID = userID
	return interceptors.ContextWithJWTClaims(ctx, claim)
}

func contextWithResourcePathAndUserID(ctx context.Context, rp string, userID string) context.Context {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim == nil {
		claim = &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{},
		}
	}
	claim.Manabie.ResourcePath = rp
	claim.Manabie.UserID = userID
	return interceptors.ContextWithJWTClaims(ctx, claim)
}

func i32ToStr(i int32) string {
	return strconv.Itoa(int(i))
}

func strToI32(str string) int32 {
	i, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		panic(err)
	}
	return int32(i)
}

func doRetry(f func() (bool, error)) error {
	return try.Do(func(attempt int) (bool, error) {
		retry, err := f()
		if err != nil {
			if retry {
				time.Sleep(2 * time.Second)
				return attempt < 10, err
			}
			return false, err
		}
		return false, nil
	})
}

func i32resourcePathFromCtx(ctx context.Context) int32 {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim != nil && claim.Manabie != nil {
		rp := claim.Manabie.ResourcePath
		ret, err := strconv.ParseInt(rp, 10, 32)
		if err != nil {
			panic(fmt.Errorf("ctx has invalid resource path %w", err))
		}
		return int32(ret)
	}
	panic("ctx has no resource path")
}

func resourcePathFromCtx(ctx context.Context) string {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim != nil && claim.Manabie != nil {
		return claim.Manabie.ResourcePath
	}
	panic("ctx has no resource path")
}

func EnsureSchoolAdminToken(ctx context.Context, s *common.Suite) (context.Context, error) {
	if !s.ContextHasToken(ctx) {
		ctx2, err := s.ASignedInWithSchool(ctx, "school admin", i32resourcePathFromCtx(ctx))
		if err != nil {
			return ctx, err
		}
		return ctx2, nil
	}
	return ctx, nil
}

var (
	alwaysEqual = cmp.Comparer(func(_, _ interface{}) bool { return true })

	defaultCmpOptions = []cmp.Option{
		// Use proto.Equal for protobufs
		cmp.Comparer(proto.Equal),
		// Use big.Rat.Cmp for big.Rats
		cmp.Comparer(func(x, y *big.Rat) bool {
			if x == nil || y == nil {
				return x == y
			}
			return x.Cmp(y) == 0
		}),
		// NaNs compare equal
		cmp.FilterValues(func(x, y float64) bool {
			return math.IsNaN(x) && math.IsNaN(y)
		}, alwaysEqual),
		cmp.FilterValues(func(x, y float32) bool {
			return math.IsNaN(float64(x)) && math.IsNaN(float64(y))
		}, alwaysEqual),
	}
)

// protoEqual tests two values for equality.
func protoEqual(x, y interface{}, opts ...cmp.Option) bool {
	// Put default options at the end. Order doesn't matter.
	opts = append(opts[:len(opts):len(opts)], defaultCmpOptions...)
	return cmp.Equal(x, y, opts...)
}

// protoDiff reports the differences between two values.
// protoDiff(x, y) == "" iff Equal(x, y).
func protoDiff(x, y interface{}, opts ...cmp.Option) string {
	// Put default options at the end. Order doesn't matter.
	opts = append(opts[:len(opts):len(opts)], defaultCmpOptions...)
	return cmp.Diff(x, y, opts...)
}

func aSampleComposedNotification(receiverIds []string, schoolID int32, target string, isImportant bool) *cpb.Notification {
	var userGroupFilter *cpb.NotificationTargetGroup_UserGroupFilter
	switch target {
	case "student":
		userGroupFilter = &cpb.NotificationTargetGroup_UserGroupFilter{UserGroups: []cpb.UserGroup{cpb.UserGroup_USER_GROUP_STUDENT}}
	case "parent":
		userGroupFilter = &cpb.NotificationTargetGroup_UserGroupFilter{UserGroups: []cpb.UserGroup{cpb.UserGroup_USER_GROUP_PARENT}}
	default:
		userGroupFilter = &cpb.NotificationTargetGroup_UserGroupFilter{UserGroups: []cpb.UserGroup{cpb.UserGroup_USER_GROUP_STUDENT, cpb.UserGroup_USER_GROUP_PARENT}}
	}

	infoNotification := &cpb.Notification{
		Data: `{"somekey":"somevalue"}`,
		// ReceiverIds: receiverIds,
		GenericReceiverIds: receiverIds,
		Message: &cpb.NotificationMessage{
			Title: "noti title",
			Content: &cpb.RichText{
				Raw:      "raw rich text",
				Rendered: "rendered rich text",
			},
		},
		Type:   cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED,
		Event:  cpb.NotificationEvent_NOTIFICATION_EVENT_NONE,
		Status: cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT,
		TargetGroup: &cpb.NotificationTargetGroup{
			CourseFilter:    &cpb.NotificationTargetGroup_CourseFilter{Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE},
			GradeFilter:     &cpb.NotificationTargetGroup_GradeFilter{Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE},
			LocationFilter:  &cpb.NotificationTargetGroup_LocationFilter{Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE},
			ClassFilter:     &cpb.NotificationTargetGroup_ClassFilter{Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE},
			SchoolFilter:    &cpb.NotificationTargetGroup_SchoolFilter{Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE},
			UserGroupFilter: userGroupFilter,
		},
		SchoolId:    schoolID,
		IsImportant: isImportant,
	}

	return infoNotification
}

func retrievePushedNotification(ctx context.Context, noti *grpc.ClientConn, deviceToken string) (*npb.RetrievePushedNotificationMessageResponse, error) {
	respNoti, err := npb.NewInternalServiceClient(noti).RetrievePushedNotificationMessages(
		ctx,
		&npb.RetrievePushedNotificationMessageRequest{
			DeviceToken: deviceToken,
			Limit:       1,
			Since:       timestamppb.Now(),
		})

	if err != nil {
		return nil, err
	}
	return respNoti, nil
}

type NewNotiSvcProxy struct {
	notiSvc *grpc.ClientConn
}

func (p *NewNotiSvcProxy) randomlyProxyToNewNotificationSVCMiddleware(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	shouldProxy := rand.Intn(2) == 0
	shouldConvertToNewProto := rand.Intn(2) == 0
	if !shouldProxy {
		return invoker(ctx, method, req, reply, cc, opts...)
	}
	parts := strings.Split(method, "/")
	var svcNamePrefix string
	switch parts[1] {
	case "bob.v1.NotificationReaderService":
		svcNamePrefix = "notificationmgmt.v1.NotificationReaderService"
	case "bob.v1.NotificationModifierService":
		// we still redirect traffic to notimgmt, but new proto does not have this endpoint
		shouldConvertToNewProto = false
		svcNamePrefix = "bob.v1.NotificationModifierService"
	case "yasuo.v1.NotificationModifierService":
		svcNamePrefix = "notificationmgmt.v1.NotificationModifierService"
	}
	if svcNamePrefix == "" {
		return invoker(ctx, method, req, reply, cc, opts...)
	}
	if shouldConvertToNewProto {
		method = fmt.Sprintf("/%s/%s", svcNamePrefix, parts[2])
	}
	return p.notiSvc.Invoke(ctx, method, req, reply, opts...)
}

// Similar to protoEqual but this function doesn't take into account of interface's fields/values's ordering.
// Return true if x,y are equal.
//
//	ok, diff := protoEqualWithoutOrder(x, y, diff.AllowTypeMismatch(true), diff.DisableStructValues())
//	if !ok {
//		fmt.Println(diff)
//	}
func protoEqualWithoutOrder(x, y interface{}, ignoreField []string, opts ...func(d *diff.Differ) error) (bool, string) {
	// we don't need to compare these fields for proto interfaces
	doNotDescendPaths := []string{"state", "sizeCache", "unknownFields"}

	if len(ignoreField) > 0 {
		doNotDescendPaths = append(doNotDescendPaths, ignoreField...)
	}

	opts = append(opts, diff.Filter(func(path []string, parent reflect.Type, field reflect.StructField) bool {
		return !slice.Contains(doNotDescendPaths, path[0])
	}))

	chl, _ := diff.Diff(x, y, opts...)
	if len(chl) == 0 {
		return true, ""
	}
	var p []byte
	p, err := json.MarshalIndent(chl, "", "\t")
	if err != nil {
		return false, err.Error()
	}
	return false, string(p)
}

func makeAnswersListForOnlyRequiredQuestion(questions []*cpb.Question) []*cpb.Answer {
	answers := []*cpb.Answer{}
	for _, q := range questions {
		if !q.Required {
			continue
		}
		switch q.Type {
		case cpb.QuestionType_QUESTION_TYPE_CHECK_BOX:
			answers = append(answers, &cpb.Answer{
				Answer:                  q.Choices[0],
				QuestionnaireQuestionId: q.QuestionnaireQuestionId,
			})
			answers = append(answers, &cpb.Answer{
				Answer:                  q.Choices[1],
				QuestionnaireQuestionId: q.QuestionnaireQuestionId,
			})
		case cpb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE:
			answers = append(answers, &cpb.Answer{
				Answer:                  q.Choices[1],
				QuestionnaireQuestionId: q.QuestionnaireQuestionId,
			})
		case cpb.QuestionType_QUESTION_TYPE_FREE_TEXT:
			answers = append(answers, &cpb.Answer{Answer: idutil.ULIDNow(), QuestionnaireQuestionId: q.QuestionnaireQuestionId})
		}
	}
	return answers
}
