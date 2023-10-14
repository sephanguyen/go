package managing

import (
	"sync"

	bobpb "github.com/manabie-com/backend/pkg/genproto/bob"
	tompb "github.com/manabie-com/backend/pkg/genproto/tom"
	yasuopb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	eurekapb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	fatimapb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	shamirpb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"

	"github.com/nats-io/nats.go"
)

func initStepForZeus(s *suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^A user with username is "([^"]*)" and password is "([^"]*)"$`: s.aUserWithUserInfo,
		`^"([^"]*)" publishes a message with subject "([^"]*)"$`:        s.publishesAMessageWithSubject,
		`^"([^"]*)" publishes this message "([^"]*)"$`:                  s.publishesMessageStatus,
		`^"([^"]*)" subscribes a message with subject "([^"]*)"$`:       s.subscribesAMessageWithSubject,
		`^"([^"]*)" subscribes this message "([^"]*)"$`:                 s.subscribesThisMessageSuccessfully,

		`^data in table activity log is empty$`: s.dataInTableActivityLogIsEmpty,
		`^(\d+) request update user profile with user group: "([^"]*)", name: "([^"]*)", phone: "([^"]*)", email: "([^"]*)", school: (\d+)$`: s.requestUpdateUserProfileWithUserGroupNamePhoneEmailSchool,
		`^all of above request are sent$`:                                s.allOfAboveRequestAreSent,
		`^number of record in table activity log is (\d+)$`:              s.numberOfRecordInTableActivityLog,
		`^a lesson conversation background$`:                             s.tomSuite.ALessonConversationBackground,
		`^(\d+) request get conversation using current conversation id$`: s.requestGetConversationByID,
		// `^a school with random name background$`:                                                   s.aSchoolWithRandomNameBackground,
		// `^(\d+) request update school with country "([^"]*)", city "([^"]*)", district "([^"]*)"$`: s.requestUpdateSchool,
		`^a valid course background$`:          s.eurekaSuite.AValidCourseBackground,
		`^(\d+) request list class by course$`: s.requestListClassByCourse,
		`^(\d+) request create package$`:       s.requestCreatePackage,
		`^(\d+) request verify token$`:         s.requestVerifyToken,

		`^"([^"]*)" publishes (\d+) message with subject "([^"]*)" and same Nats-Msg-Id$`: s.publishSomeMessageWithSameNatsMsgID,
		`^total activity log is inserted must be one$`:                                    s.totalRecordIsInsertedMustBeOne,
		`^"([^"]*)" publishes (\d+) message with subject "([^"]*)"$`:                      s.publishSomeMessage,
		`^These activity log are created by Zeus$`:                                        s.theseActivityLogAreCreatedByZeus,
		`^Some message above must be deleted from stream "([^"]*)"$`:                      s.messageMustBeDeletedFromStream,
	}
	return steps
}

type ZeusStepState struct {
	MapJSContext        map[string]nats.JetStreamContext
	MapPublishStatus    map[string]error
	MapSubscribeStatus  map[string]error
	ListNatJSConnection []*nats.Conn
	sync.RWMutex

	BobAuthToken              string
	TomAuthToken              string
	EurekaAuthToken           string
	UpdateUserProfileRequests []*bobpb.UpdateUserProfileRequest
	GetConversationRequests   []*tompb.GetConversationRequest
	UpdateSchoolRequests      []*yasuopb.UpdateSchoolRequest
	GetClassByCourseRequests  []*eurekapb.ListClassByCourseRequest
	CreatePackageRequests     []*fatimapb.CreatePackageRequest
	VerifyTokenRequests       []*shamirpb.VerifyTokenRequest

	CurrentActionType   string
	CurrentResourcePath string

	ActivityLogPublished                []npb.ActivityLogEvtCreated
	SequenceIDs                         []uint64
	CurrentUserNameConnectedToJetStream string
}
