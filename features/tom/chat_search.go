package tom

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/godogutil"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/try"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/cucumber/godog"
	gogoproto "github.com/gogo/protobuf/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type commonState struct {
	schoolID        string
	locationPool    map[string]string
	coursePool      map[string]string
	studentChatPool map[string]chatInfo
	parentChatPool  map[string]chatInfo
}

type filterSuiteState struct {
	language          string
	studentCoursesMap map[string][]string
	courses           []string
	studentIDs        []string
	studentParentsMap map[string][]string
	parentStudentMap  map[string]string
	chatStudentMap    map[string]string
	teacher           string
	listRequest       *tpb.ListConversationsInSchoolRequest
	listResponse      *tpb.ListConversationsInSchoolResponse
	// schoolID            string
	defaultLocationID     string
	defaultLocationTypeID string
	totalChats            []chatInfo
	newlyRepliedChat      string
	filterExpectedChats   []func(previousResult []chatInfo) ([]chatInfo, error)
	updatedStudentName    string
}

func initFilterChatGroupSuite(previousMap map[string]interface{}, ctx *godog.ScenarioContext, s *suite) {
	s.filterSuiteState = filterSuiteState{
		listRequest: &tpb.ListConversationsInSchoolRequest{Paging: &cpb.Paging{Limit: 100}}, studentCoursesMap: make(map[string][]string), studentParentsMap: make(map[string][]string), chatStudentMap: make(map[string]string), parentStudentMap: make(map[string]string),
	}
	s.commonState = commonState{
		coursePool:      map[string]string{},
		locationPool:    map[string]string{},
		studentChatPool: map[string]chatInfo{},
		parentChatPool:  map[string]chatInfo{},
	}
	mergedMap := map[string]interface{}{
		`^each student joined some courses at random$`:                                                                          s.eachStudentJoinedSomeCoursesAtRandom,
		`^parents\' chats are created$`:                                                                                         s.parentsChatsAreCreated,
		`^students and each has from (\d+) to (\d+) parents randomly$`:                                                          s.studentsAndEachHasFromToParentsRandomly,
		`^students\' chats are created$`:                                                                                        s.studentsChatsAreCreated,
		`^only chats with status "([^"]*)" are returned$`:                                                                       s.onlyChatsWithStatusAreReturned,
		`^some parents\' chat has new message from parent$`:                                                                     s.someParentsChatHasNewMessageFromParent,
		`^some student\'s chat has new message from student$`:                                                                   s.someStudentsChatHasNewMessageFromStudent,
		`^teacher replies to some of those chat$`:                                                                               s.teacherRepliesToSomeOfThoseChat,
		`^filter chats with type message type "([^"]*)" only$`:                                                                  s.filterChatsWithTypeMessageTypeOnly,
		`^filter chats with type contact type "([^"]*)" only$`:                                                                  s.filterChatsWithTypeContactTypeOnly,
		`^only chats with type "([^"]*)" are returned$`:                                                                         s.onlyChatsWithTypeAreReturned,
		`^filter by a course id$`:                                                                                               s.filterByACourseID,
		`^chats that belong to that course are returned$`:                                                                       s.chatsThatBelongToThatCourseAreReturned,
		`^filter by multiple courses$`:                                                                                          s.filterByMultipleCourses,
		`^chats with type "([^"]*)", "([^"]*)" that belong to that course are returned$`:                                        s.chatsWithTypeThatBelongToThatCourseAreReturned,
		`^chats with type "([^"]*)", "([^"]*)" that belong to those courses are returned$`:                                      s.chatsWithTypeThatBelongToThoseCoursesAreReturned,
		`^chats that belong to those courses are returned$`:                                                                     s.chatsThatBelongToThoseCoursesAreReturned,
		`^a result of unreplied chats filter$`:                                                                                  s.aResultOfUnrepliedChatsFilter,
		`^teacher replies to an unreplied chat$`:                                                                                s.teacherRepliesToAnUnrepliedChat,
		`^only chats with status "([^"]*)" are returned including previous replied chat$`:                                       s.onlyChatsWithStatusAreReturnedIncludingPreviousRepliedChat,
		`^chats with name including the partial names are returned$`:                                                            s.chatsWithNameIncludingThePartialNamesAreReturned,
		`^use partial names of a student to search$`:                                                                            s.usePartialNamesOfAStudentToSearch,
		`^"([^"]*)" language is used$`:                                                                                          s.languageIsUsed,
		`^chats with name including the full names are returned$`:                                                               s.chatsWithNameIncludingTheFullNamesAreReturned,
		`^use full names of a student to search$`:                                                                               s.useFullNamesOfAStudentToSearch,
		`^chats with those names and belong to those courses are returned$`:                                                     s.chatsWithThoseNamesAndBelongToThoseCoursesAreReturned,
		`^filter by a course that the student does not belong to$`:                                                              s.filterByACourseThatTheStudentDoesNotBelongTo,
		`^nothing is returned$`:                                                                                                 s.nothingIsReturned,
		`^chats with type "([^"]*)" having name including the name of the student are returned$`:                                s.chatsWithTypeHavingNameIncludingTheNameOfTheStudentAreReturned,
		`^chats with type "([^"]*)" belonging to those courses and having name including the name of the student are returned$`: s.chatsWithTypeBelongingToThoseCoursesAndHavingNameIncludingTheNameOfTheStudentAreReturned,
		`^chats with status "([^"]*)" with type "([^"]*)" belonging to those courses and having name including the name of the student are returned$`: s.chatsWithStatusWithTypeBelongingToThoseCoursesAndHavingNameIncludingTheNameOfTheStudentAreReturned,
		`^a new school is created$`:                                  s.aNewSchoolIsCreated,
		`^returned chats have correct student ids$`:                  s.returnedChatsHaveCorrectStudentIds,
		`^a teacher account in db$`:                                  s.aTeacherAccountInDB,
		`^chat pagination (\d+) item$`:                               s.chatPaginationItem,
		`^all chat is returned$`:                                     s.allChatIsReturned,
		`^pagination by offset of "([^"]*)" limit (\d+)$`:            s.paginationByOffsetOfLimit,
		`^pagination return "([^"]*)" number of items$`:              s.paginationReturnNumberOfItems,
		`^a student name is updated with event "([^"]*)"$`:           s.aStudentNameIsUpdatedWithEvent,
		`^student and parent chats with updated names are returned$`: s.studentAndParentChatsWithUpdatedNamesAreReturned,
		`^use updated name of student to search$`:                    s.useUpdatedNameOfStudentToSearch,

		// search with locations
		`^a new school is created with location "([^"]*)"$`:                                                                   s.aNewSchoolIsCreatedWithLocation,
		`^locations "([^"]*)" children of location "([^"]*)" are created$`:                                                    s.locationsChildrenOfLocationAreCreated,
		`^student chat "([^"]*)" and parent chat "([^"]*)" are created with locations "([^"]*)"$`:                             s.studentChatAndParentChatAreCreatedWithLocations,
		`^filter chats with location "([^"]*)"$`:                                                                              s.filterChatsWithLocation,
		`^chats with ids "([^"]*)" are returned$`:                                                                             s.chatsWithIdsAreReturned,
		`^usermgmt send event upsert user profile for student of chat "([^"]*)" with locations "([^"]*)"$`:                    s.usermgmtSendEventUpsertUserProfileForStudentOfChatWithLocations,
		`^student chat "([^"]*)" parent chat "([^"]*)" same parent with chat "([^"]*)" are created with locations "([^"]*)"$`: s.studentChatParentChatSameParentWithChatAreCreatedWithLocations,
		`^filter chats with filter combination "([^"]*)"$`:                                                                    s.filterChatsWithFilterCombination,
		`^mappings between student and course "([^"]*)"$`:                                                                     s.mappingsBetweenStudentAndCourse,
	}

	applyMergedSteps(ctx, previousMap, mergedMap)
}

func (s *suite) paginationReturnNumberOfItems(ctx context.Context, script string) (context.Context, error) {
	var want int
	switch script {
	case "num conversations - 1":
		want = len(s.filterSuiteState.totalChats) - 1
	case "0":
		want = 0
	default:
		return ctx, fmt.Errorf("unknown param %s", script)
	}
	err := s.sendListConversationRequestAndAssert(ctx, func(res *tpb.ListConversationsInSchoolResponse, _ []chatInfo) error {
		if len(res.GetItems()) != want {
			return fmt.Errorf("want %d item, has %d", want, len(res.GetItems()))
		}
		return nil
	})
	return ctx, err
}

func (s *suite) paginationByOffsetOfLimit(ctx context.Context, offsetType string, limit int) (context.Context, error) {
	var (
		offsetString  string
		offsetInteger int64
	)
	switch offsetType {
	case "next page":
		nextPage := s.filterSuiteState.listResponse.NextPage.GetOffsetCombined()
		offsetInteger = nextPage.GetOffsetInteger()
		offsetString = nextPage.GetOffsetString()
	case "first item":
		firstItem := s.filterSuiteState.listResponse.GetItems()[0]
		offsetInteger = firstItem.LastMessage.UpdatedAt.AsTime().UnixMilli()
		offsetString = firstItem.ConversationId
	default:
		return ctx, fmt.Errorf("invalid offset type %s", offsetType)
	}
	s.filterSuiteState.listRequest.Paging = &cpb.Paging{Limit: uint32(limit), Offset: &cpb.Paging_OffsetCombined{
		OffsetCombined: &cpb.Paging_Combined{OffsetString: offsetString, OffsetInteger: offsetInteger},
	}}

	return ctx, nil
}

func (s *suite) allChatIsReturned(ctx context.Context) (context.Context, error) {
	allChat := len(s.filterSuiteState.totalChats)
	err := s.sendListConversationRequestAndAssert(ctx, func(res *tpb.ListConversationsInSchoolResponse, _ []chatInfo) error {
		if len(res.GetItems()) != allChat {
			return fmt.Errorf("want %d item, has %d", allChat, len(res.GetItems()))
		}
		return nil
	})
	return ctx, err
}

func (s *suite) chatPaginationItem(ctx context.Context, limit int) (context.Context, error) {
	s.filterSuiteState.listRequest.Paging = &cpb.Paging{Limit: uint32(limit)}
	return ctx, nil
}

func (s *suite) aTeacherAccountInDB(ctx context.Context) (context.Context, error) {
	profile, token, err := s.CommonSuite.CreateTeacher(ctx)
	if err != nil {
		return ctx, err
	}
	s.teacherID = profile.StaffId
	s.filterSuiteState.teacher = s.teacherID
	s.TeacherToken = token
	return ctx, nil
}

func (s *suite) returnedChatsHaveCorrectStudentIds(ctx context.Context) (context.Context, error) {
	if s.filterSuiteState.listResponse == nil {
		return ctx, fmt.Errorf("expect previous step to return a response")
	}
	for _, item := range s.filterSuiteState.listResponse.GetItems() {
		studentID := item.StudentId
		expectedStudentID, ok := s.filterSuiteState.chatStudentMap[item.ConversationId]
		if !ok {
			return ctx, fmt.Errorf("chat-student map does not include conversation %s, check db", item.ConversationId)
		}
		if expectedStudentID != studentID {
			return ctx, fmt.Errorf("expecting conversation %s to have student %s, but %s returned", item.ConversationId, expectedStudentID, studentID)
		}
	}
	return ctx, nil
}

func (s *suite) aGenerateSchool(ctx context.Context) (int32, string, string, error) {
	random := idutil.ULIDNow()
	ctx, err := s.CommonSuite.ASchoolNameCountryCityDistrict(ctx, random, random, random, random)
	if err != nil {
		return 0, "", "", err
	}
	ctx, err = s.CommonSuite.AdminInsertsSchools(ctx)
	if err != nil {
		return 0, "", "", err
	}

	state := common.StepStateFromContext(ctx)
	school := state.CurrentSchoolID
	resourcePath := strconv.Itoa(int(school))
	err = s.CommonSuite.GenerateOrganizationAuth(ctx, resourcePath)
	if err != nil {
		return 0, "", "", fmt.Errorf("s.CommonSuite.GenerateOrganizationAuth %s", err)
	}

	// default location
	locationID, locationTypeID, err := s.CommonSuite.CreateLocationWithDB(ctx, resourcePath, "org", "", "")
	if err != nil {
		return 0, "", "", err
	}

	// permission and role
	err = s.CommonSuite.GenerateOrganizationRoleAndPermission(ctx, locationID, resourcePath)
	if err != nil {
		return 0, "", "", fmt.Errorf("s.CommonSuite.GenerateOrganizationRoleAndPermission %s", err)
	}
	return state.CurrentSchoolID, locationID, locationTypeID, err
}

func (s *suite) aNewSchoolIsCreated(ctx context.Context) (context.Context, error) {
	schoolID, locationID, locationTypeID, err := s.aGenerateSchool(ctx)
	if err != nil {
		return ctx, err
	}

	s.commonState.schoolID = strconv.Itoa(int(schoolID))
	s.filterSuiteState.defaultLocationID = locationID
	s.filterSuiteState.defaultLocationTypeID = locationTypeID
	return contextWithResourcePath(ctx, strconv.Itoa(int(schoolID))), nil
}

func (s *suite) chatsWithStatusWithTypeBelongingToThoseCoursesAndHavingNameIncludingTheNameOfTheStudentAreReturned(ctx context.Context, replyStatus, contactType string) (context.Context, error) {
	filteredCourses := s.filterSuiteState.listRequest.CourseIds
	name := s.filterSuiteState.listRequest.Name.Value
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("reply status", replyStatus))
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("contact", contactType))
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("courses", filteredCourses...))
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("partial name", name))
	return ctx, s.sendListConversationRequestAndAssert(ctx, s.assertConversationListResultWithName)
}

func (s *suite) chatsWithTypeThatBelongToThatCourseAreReturned(ctx context.Context, replyStatus, contactType string) (context.Context, error) {
	filteredCourses := s.filterSuiteState.listRequest.CourseIds
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("reply status", replyStatus))
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("contact", contactType))
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("courses", filteredCourses...))
	return ctx, s.sendListConversationRequestAndAssert(ctx, s.assertConversationListResultWithName)
}

func (s *suite) chatsWithTypeThatBelongToThoseCoursesAreReturned(ctx context.Context, replyStatus, contactType string) (context.Context, error) {
	filteredCourses := s.filterSuiteState.listRequest.CourseIds
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("reply status", replyStatus))
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("contact", contactType))
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("courses", filteredCourses...))
	return ctx, s.sendListConversationRequestAndAssert(ctx, s.assertConversationListResultWithIDs)
}

func (s *suite) chatsWithTypeBelongingToThoseCoursesAndHavingNameIncludingTheNameOfTheStudentAreReturned(ctx context.Context, contactType string) (context.Context, error) {
	filteredCourses := s.filterSuiteState.listRequest.CourseIds
	name := s.filterSuiteState.listRequest.Name.Value
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("contact", contactType))
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("courses", filteredCourses...))
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("partial name", name))
	return ctx, s.sendListConversationRequestAndAssert(ctx, s.assertConversationListResultWithName)
}

func (s *suite) chatsWithTypeHavingNameIncludingTheNameOfTheStudentAreReturned(ctx context.Context, contactType string) (context.Context, error) {
	name := s.filterSuiteState.listRequest.Name.Value
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("contact", contactType))
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("partial name", name))
	return ctx, s.sendListConversationRequestAndAssert(ctx, s.assertConversationListResultWithName)
}

func (s *suite) nothingIsReturned(ctx context.Context) (context.Context, error) {
	return ctx, s.sendListConversationRequestAndAssert(ctx, func(res *tpb.ListConversationsInSchoolResponse, expected []chatInfo) error {
		if len(res.GetItems()) == 0 {
			return nil
		}
		return fmt.Errorf("%d items are returned instead of nothing", len(res.GetItems()))
	})
}

func (s *suite) filterByACourseThatTheStudentDoesNotBelongTo(ctx context.Context) (context.Context, error) {
	chosenCourse := idutil.ULIDNow()
	s.filterSuiteState.listRequest.CourseIds = []string{chosenCourse}
	return ctx, nil
}

func (s *suite) chatsWithThoseNamesAndBelongToThoseCoursesAreReturned(ctx context.Context) (context.Context, error) {
	name := s.filterSuiteState.listRequest.Name.Value
	courses := s.filterSuiteState.listRequest.CourseIds
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("courses", courses...))
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("full name", name))
	return ctx, s.sendListConversationRequestAndAssert(ctx, s.assertConversationListResultWithName)
}

func (s *suite) chatsWithNameIncludingTheFullNamesAreReturned(ctx context.Context) (context.Context, error) {
	name := s.filterSuiteState.listRequest.Name.Value
	courses := s.filterSuiteState.listRequest.CourseIds
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("courses", courses...))
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("full name", name))
	return ctx, s.sendListConversationRequestAndAssert(ctx, s.assertConversationListResultWithName)
}

func (s *suite) languageIsUsed(ctx context.Context, language string) (context.Context, error) {
	_, ok := studentNames[language]
	if !ok {
		return ctx, fmt.Errorf("language %s is not supported", language)
	}
	s.filterSuiteState.language = language
	return ctx, nil
}

func (s *suite) studentAndParentChatsWithUpdatedNamesAreReturned(ctx context.Context) error {
	return s.sendListConversationRequestAndAssert(ctx, s.assertUpdatedParentStudentChat)
}

func (s *suite) assertUpdatedParentStudentChat(res *tpb.ListConversationsInSchoolResponse, _ []chatInfo) error {
	updatedName := s.filterSuiteState.updatedStudentName
	if len(res.GetItems()) != 2 {
		return fmt.Errorf("want 2, has %d chats returned", len(res.GetItems()))
	}
	for _, item := range res.GetItems() {
		if item.GetConversationName() != updatedName {
			return fmt.Errorf("want chat with name %s, has %s", updatedName, item.GetConversationName())
		}
	}
	return nil
}

func (s *suite) chatsWithNameIncludingThePartialNamesAreReturned(ctx context.Context) (context.Context, error) {
	name := s.filterSuiteState.listRequest.Name.Value
	filterExpectedChat := s.generateFilterFunc("partial name", name)
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, filterExpectedChat)
	return ctx, s.sendListConversationRequestAndAssert(ctx, s.assertConversationListResultWithName)
}

var (
	idToName     = map[string]string{}
	studentNames = map[string][]string{}
	nameGenLock  = &sync.Mutex{}
	langEng      = "english"
)

func readIntoSlice(f io.Reader) ([]string, error) {
	sl := []string{}
	reader := bufio.NewReader(f)
	line := 0
	for {
		bs, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return sl, nil
			}
			return nil, fmt.Errorf("failed to read file at line %d: %s", line, err)
		}
		if len(bs) == 0 {
			return nil, fmt.Errorf("unexpected blank line in seed data file at line %d", line)
		}
		sl = append(sl, string(bs))
		line++
	}
}

func seedLanguagesData() {
	workDir, err := os.Getwd()
	zapLogger := logger.NewZapLogger("info", true)
	if err != nil {
		zapLogger.Panic("failed to get current working directory")
	}
	filenames := []string{"hiragana", "kanji", "katakana", "english"}
	for _, languageform := range filenames {
		file, err := os.Open(filepath.Join(workDir, "tom", "data", languageform+".txt"))
		if err != nil {
			zapLogger.Panic(fmt.Sprintf("failed to open %s: %s", languageform, err))
		}
		defer file.Close()
		slc, err := readIntoSlice(file)
		if err != nil {
			zapLogger.Panic(fmt.Sprintf("cannot read file %s data: %s", languageform, err))
		}
		studentNames[languageform] = slc
	}
}
func init() { seedLanguagesData() }
func makePartialName(name string, language string) string {
	if language == "english" {
		half := len(name) / 2
		return strings.TrimSpace(name[:half])
	}
	chars := []rune(name)
	half := len(chars) / 2
	return strings.TrimSpace(string(chars[:half]))
}

func (s *suite) useFullNamesOfAStudentToSearch(ctx context.Context) (context.Context, error) {
	language := s.filterSuiteState.language
	if language == "" {
		return ctx, errors.New("no language is selected")
	}
	astudent := s.filterSuiteState.studentIDs[0]
	chatName := pseudoNameForStudentChat(astudent, language)
	s.filterSuiteState.listRequest.Name = wrapperspb.String(chatName)
	return ctx, nil
}

func (s *suite) aStudentNameIsUpdatedWithEvent(ctx context.Context, _ string) error {
	language := s.filterSuiteState.language
	if language == "" {
		return errors.New("no language is selected")
	}
	aStudent := s.filterSuiteState.studentIDs[0]
	chatName := pseudoNameForStudentChat(aStudent, language)
	s.filterSuiteState.updatedStudentName = chatName + " updated"
	evt := &pb.EvtUserInfo{
		UserId: aStudent,
		Name:   s.filterSuiteState.updatedStudentName,
	}
	data, err := gogoproto.Marshal(evt)
	if err != nil {
		return err
	}
	_, err = s.JSM.TracedPublishAsync(ctx, "Publish SubjectUserDeviceTokenUpdated", constants.SubjectUserDeviceTokenUpdated, data)
	return err
}

func (s *suite) useUpdatedNameOfStudentToSearch(ctx context.Context) error {
	s.filterSuiteState.listRequest.Name = wrapperspb.String(s.filterSuiteState.updatedStudentName)
	return nil
}

func (s *suite) usePartialNamesOfAStudentToSearch(ctx context.Context) (context.Context, error) {
	language := s.filterSuiteState.language
	if language == "" {
		return ctx, errors.New("no language is selected")
	}
	aStudent := s.filterSuiteState.studentIDs[0]
	chatName := pseudoNameForStudentChat(aStudent, language)
	partialName := makePartialName(chatName, language)
	s.filterSuiteState.listRequest.Name = wrapperspb.String(partialName)
	return ctx, nil
}

func (s *suite) onlyChatsWithStatusAreReturnedIncludingPreviousRepliedChat(ctx context.Context, replyStatus string) (context.Context, error) {
	filterExpectedChat := s.generateFilterFunc("exclude id", s.filterSuiteState.newlyRepliedChat)
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("reply status", replyStatus))
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, filterExpectedChat)
	return ctx, s.sendListConversationRequestAndAssert(ctx, s.assertConversationListResultWithIDs)
}

func (s *suite) teacherRepliesToAnUnrepliedChat(ctx context.Context) (context.Context, error) {
	var tobereplied chatInfo
	for idx, info := range s.filterSuiteState.totalChats {
		if !info.replied {
			tobereplied = info
			break
		}
		if idx == len(s.filterSuiteState.totalChats) {
			return ctx, fmt.Errorf("unproper seed data, no unreplied chat to proceed")
		}
	}
	s.conversationID = tobereplied.id

	ctx, err := godogutil.MultiErrChain(
		ctx,
		s.teacherJoinsChat, tobereplied.id,
		s.userSendsItemWithContent, "text", "hello world", s.filterSuiteState.teacher, cpb.UserGroup_USER_GROUP_TEACHER,
	)
	if err != nil {
		return ctx, err
	}
	s.filterSuiteState.newlyRepliedChat = tobereplied.id

	return ctx, nil
}

func (s *suite) aResultOfUnrepliedChatsFilter(ctx context.Context) (context.Context, error) {
	return godogutil.MultiErrChain(ctx,
		s.someStudentsChatHasNewMessageFromStudent,
		s.someParentsChatHasNewMessageFromParent,
		s.teacherRepliesToSomeOfThoseChat,
		s.filterChatsWithTypeMessageTypeOnly, "Unreplied",
	)
}

func (s *suite) chatsThatBelongToThoseCoursesAreReturned(ctx context.Context) (context.Context, error) {
	courses := s.filterSuiteState.listRequest.CourseIds
	filterExpectedChat := s.generateFilterFunc("courses", courses...)
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, filterExpectedChat)
	return ctx, s.sendListConversationRequestAndAssert(ctx, s.assertConversationListResultWithIDs)
}

func (s *suite) chatsThatBelongToThatCourseAreReturned(ctx context.Context) (context.Context, error) {
	courses := s.filterSuiteState.listRequest.CourseIds
	filterExpectedChat := s.generateFilterFunc("courses", courses...)
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, filterExpectedChat)
	return ctx, s.sendListConversationRequestAndAssert(ctx, s.assertConversationListResultWithIDs)
}

func (s *suite) filterByMultipleCourses(ctx context.Context) (context.Context, error) {
	s.filterSuiteState.listRequest.CourseIds = s.filterSuiteState.courses[:2]
	return ctx, nil
}

func intersect(a, b []string) bool {
	aMap := map[string]bool{}
	for _, item := range a {
		aMap[item] = true
	}
	for _, item := range b {
		if aMap[item] {
			return true
		}
	}
	return false
}

func (s *suite) filterByACourseID(ctx context.Context) (context.Context, error) {
	aCourse := s.filterSuiteState.courses[0]
	s.filterSuiteState.listRequest.CourseIds = []string{aCourse}
	return ctx, nil
}

func (s *suite) sendListConversationRequestAndAssert(ctx context.Context, fnAssert func(*tpb.ListConversationsInSchoolResponse, []chatInfo) error) error {
	expected := s.filterSuiteState.totalChats
	for _, filter := range s.filterSuiteState.filterExpectedChats {
		newlyFiltered, err := filter(expected)
		if err != nil {
			return err
		}
		expected = newlyFiltered
	}
	schoolID, _ := strconv.Atoi(s.commonState.schoolID)
	token, err := s.generateExchangeToken(s.filterSuiteState.teacher, cpb.UserGroup_USER_GROUP_TEACHER.String(), applicantID, int64(schoolID), s.ShamirConn)
	if err != nil {
		return err
	}
	enventually := func(attempt int) (bool, error) {
		res, err := tpb.NewChatReaderServiceClient(s.Conn).
			ListConversationsInSchoolWithLocations(
				contextWithToken(ctx, token),
				s.filterSuiteState.listRequest)
		if err != nil {
			return false, err
		}
		s.filterSuiteState.listResponse = res
		err = fnAssert(res, expected)
		if err != nil {
			time.Sleep(2 * time.Second)
			if attempt < 10 {
				return true, err
			}
			s.debugRequestResponse()
			return false, err
		}
		return false, nil
	}
	return try.Do(enventually)
}

// nolint
func (s *suite) generateFilterFunc(filtertype string, args ...string) func([]chatInfo) ([]chatInfo, error) {
	switch filtertype {
	case "full name":
		text := args[0]
		return func(previous []chatInfo) ([]chatInfo, error) {
			filtered := filterChats(previous, func(c chatInfo) bool { return strings.Contains(c.name, text) })
			return filtered, nil
		}
	case "partial name":
		text := args[0]
		return func(previous []chatInfo) ([]chatInfo, error) {
			filtered := filterChats(previous, func(c chatInfo) bool { return strings.Contains(c.name, text) })
			return filtered, nil
		}
	case "exclude id":
		id := args[0]
		return func(previous []chatInfo) ([]chatInfo, error) {
			filtered := filterChats(previous, func(c chatInfo) bool { return c.id != id })
			return filtered, nil
		}
	case "contact":
		contactType := args[0]
		return func(previous []chatInfo) ([]chatInfo, error) {
			switch contactType {
			case "All":
				return previous, nil
			case "Parent":
			case "Student":
			default:
				return nil, fmt.Errorf("unknown chat type status %s in scenario", contactType)
			}
			filtered := filterChats(previous, func(c chatInfo) bool { return c.chatType == contactType })
			return filtered, nil
		}
	case "reply status":
		return func(previous []chatInfo) ([]chatInfo, error) {
			var isReplied bool
			switch s.filterSuiteState.listRequest.TeacherStatus {
			case tpb.TeacherConversationStatus_TEACHER_CONVERSATION_STATUS_ALL:
				return previous, nil
			case tpb.TeacherConversationStatus_TEACHER_CONVERSATION_STATUS_NOT_REPLIED:
				isReplied = false
			default:
				isReplied = true
			}
			filtered := filterChats(previous, func(c chatInfo) bool { return c.replied == isReplied })
			return filtered, nil
		}
	case "courses":
		courses := args
		return func(previous []chatInfo) ([]chatInfo, error) {
			newExpected := filterChats(previous, func(c chatInfo) bool { return intersect(c.courses, courses) })
			return newExpected, nil
		}
	default:
		panic(fmt.Sprintf("not support filter type %s", filtertype))
	}
}

func (s *suite) onlyChatsWithTypeAreReturned(ctx context.Context, contactType string) (context.Context, error) {
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("contact", contactType))
	return ctx, s.sendListConversationRequestAndAssert(ctx, s.assertConversationListResultWithIDs)
}

func filterChats(slice []chatInfo, h func(chatInfo) bool) []chatInfo {
	ret := []chatInfo{}
	for _, info := range slice {
		if h(info) {
			ret = append(ret, info)
		}
	}
	return ret
}

func (s *suite) filterChatsWithTypeContactTypeOnly(contactType string) error {
	switch contactType {
	case "All":
		s.filterSuiteState.listRequest.Type = []tpb.ConversationType{tpb.ConversationType_CONVERSATION_PARENT, tpb.ConversationType_CONVERSATION_STUDENT}
	case "Parent":
		s.filterSuiteState.listRequest.Type = []tpb.ConversationType{tpb.ConversationType_CONVERSATION_PARENT}
	case "Student":
		s.filterSuiteState.listRequest.Type = []tpb.ConversationType{tpb.ConversationType_CONVERSATION_STUDENT}
	default:
		return fmt.Errorf("unknown chat type status %s in scenario", contactType)
	}

	return nil
}

func (s *suite) assertConversationListResultWithName(res *tpb.ListConversationsInSchoolResponse, expecteds []chatInfo) error {
	checkList := make(map[string]bool)
	for _, item := range expecteds {
		checkList[item.id] = false
	}
	notExpected := []*tpb.Conversation{}

	for _, returnedItem := range res.GetItems() {
		if _, ok := checkList[returnedItem.ConversationId]; !ok {
			notExpected = append(notExpected, returnedItem)
		} else {
			checkList[returnedItem.ConversationId] = true
		}
	}
	missingExpected := []string{}
	for id, checked := range checkList {
		if !checked {
			missingExpected = append(missingExpected, id)
		}
	}
	var errStr, warnStr string
	var returnErr error

	if len(notExpected) > 0 {
		names := make([]string, 0, len(notExpected))
		for _, item := range notExpected {
			names = append(names, item.GetConversationName())
		}
		warnStr += fmt.Sprintf("these additional names were returned: %v\n", names)
		warnStr += fmt.Sprintf("used %s to search\n", s.filterSuiteState.listRequest.GetName().GetValue())
		s.ZapLogger.Info(warnStr)
	}
	if len(missingExpected) > 0 {
		// map id with name for debug
		chatmap := make(map[string]chatInfo)
		for _, item := range expecteds {
			chatmap[item.id] = item
		}
		mapidtonames := mapSlice(missingExpected, func(id string) string { return chatmap[id].name })
		errStr += fmt.Sprintf("these names are expected to be returned, but was not: %v", mapidtonames)
		returnErr = errors.New(errStr)
	}
	return returnErr
}

func mapSlice(sl []string, f func(string) string) []string {
	ret := make([]string, 0, len(sl))
	for _, item := range sl {
		ret = append(ret, f(item))
	}
	return ret
}

func (s *suite) debugRequestResponse() {
	rq := s.filterSuiteState.listRequest
	debug := fmt.Sprintf("schoolID: %s, name: %s, courses: %v, chatType: %v, teacher status: %s, paging: %s, join status: %s\n", s.commonState.schoolID, rq.GetName(), rq.GetCourseIds(), rq.GetType(), rq.GetTeacherStatus().String(), rq.GetPaging().String(), rq.GetJoinStatus().String())
	s.ZapLogger.Info(debug)
	res := s.filterSuiteState.listResponse
	for _, item := range res.GetItems() {
		debug = fmt.Sprintf("conversation: %s, stauts: %s, isreplied: %v, student: %s", item.GetConversationId(), item.Status.String(), item.IsReplied, item.StudentId)
		s.ZapLogger.Info(debug)
	}
}

func (s *suite) assertConversationListResultWithIDs(res *tpb.ListConversationsInSchoolResponse, expecteds []chatInfo) error {
	checkList := make(map[string]bool)
	for _, item := range expecteds {
		checkList[item.id] = false
	}
	notExpected := []string{}
	for _, returnedItem := range res.GetItems() {
		if _, ok := checkList[returnedItem.GetConversationId()]; !ok {
			notExpected = append(notExpected, returnedItem.GetConversationId())
		} else {
			checkList[returnedItem.GetConversationId()] = true
		}
	}
	missingExpected := []string{}
	for id, checked := range checkList {
		if !checked {
			missingExpected = append(missingExpected, id)
		}
	}
	var errStr string
	if len(notExpected) > 0 {
		errStr = fmt.Sprintf("not expect ids to be returned: %v", notExpected)
	}
	if len(missingExpected) > 0 {
		errStr = fmt.Sprintf("%s\nthese ids are expected to be returned, but was not: %v", errStr, missingExpected)
	}
	if errStr == "" {
		return nil
	}
	return errors.New(errStr)
}

func (s *suite) onlyChatsWithStatusAreReturned(ctx context.Context, _ string) (context.Context, error) {
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, s.generateFilterFunc("reply status"))
	return ctx, s.sendListConversationRequestAndAssert(ctx, s.assertConversationListResultWithIDs)
}

func (s *suite) teacherJoinsChat(ctx context.Context, chatID string) (context.Context, error) {
	s.teacherID = s.filterSuiteState.teacher
	s.ConversationIDs = []string{chatID}
	s.conversationID = chatID
	token, err := s.generateExchangeToken(s.teacherID, cpb.UserGroup_USER_GROUP_TEACHER.String(), applicantID, s.getSchool(), s.ShamirConn)
	if err != nil {
		return ctx, err
	}
	s.TeacherToken = token
	return godogutil.MultiErrChain(ctx,
		// s.aValidToken, "current teacher",
		s.teacherJoinConversations,
		s.returnsStatusCode, "OK",
		s.teacherMustBeMemberOfConversations,
	)
}

func (s *suite) teacherRepliesToSomeOfThoseChat(ctx context.Context) (context.Context, error) {
	if s.filterSuiteState.teacher == "" {
		s.filterSuiteState.teacher = idutil.ULIDNow()
	}
	tok, err := s.genTeacherTokens([]string{s.filterSuiteState.teacher})
	if err != nil {
		return ctx, err
	}
	ctx, err = s.makeUsersSubscribeV2Ctx(ctx, []string{s.filterSuiteState.teacher}, []string{tok[0]})
	if err != nil {
		return ctx, fmt.Errorf("s.makeUsersSubscribeV2: %w", err)
	}

	newChats := []chatInfo{}
	reply := true
	for _, info := range s.filterSuiteState.totalChats {
		if !reply {
			newChats = append(newChats, info)
			reply = true
			continue
		}
		s.conversationID = info.id
		// replied = append(replied, chatID)
		ctx, err := godogutil.MultiErrChain(ctx,
			s.teacherJoinsChat, info.id,
			s.userSendsItemWithContent, "text", "hello world", s.filterSuiteState.teacher, cpb.UserGroup_USER_GROUP_TEACHER,
		)
		if err != nil {
			return ctx, fmt.Errorf("multierr.Combine: %w", err)
		}
		info.replied = true
		newChats = append(newChats, info)
		reply = false
	}
	s.filterSuiteState.totalChats = newChats
	return ctx, nil
}

func (s *suite) someStudentsChatHasNewMessageFromStudent(ctx context.Context) (context.Context, error) {
	count := 0
	countUntil := countTotalChatWithType(s.filterSuiteState.totalChats, "student") / 2
	if countUntil < 1 {
		countUntil = 1
	}
	for _, info := range s.filterSuiteState.totalChats {
		if info.chatType != "Student" {
			continue
		}
		s.conversationID = info.id
		studentID := info.studentID
		tok, err := s.genStudentToken(studentID)
		if err != nil {
			return ctx, err
		}
		ctx, err := godogutil.MultiErrChain(ctx,
			s.makeUsersSubscribeV2Ctx, []string{studentID}, []string{tok},
			s.userSendsItemWithContent, "text", "hello world", studentID, cpb.UserGroup_USER_GROUP_STUDENT,
		)
		if err != nil {
			return ctx, err
		}
		count++
		if count == countUntil {
			break
		}
	}
	return ctx, nil
}

func countTotalChatWithType(chats []chatInfo, chatType string) (ret int) {
	for _, info := range chats {
		if info.chatType == chatType {
			ret++
		}
	}
	return
}

func (s *suite) someParentsChatHasNewMessageFromParent(ctx context.Context) (context.Context, error) {
	count := 0
	totalParentChats := countTotalChatWithType(s.filterSuiteState.totalChats, "Parent")
	countUntil := totalParentChats / 2
	if countUntil < 1 {
		countUntil = 1
	}
	for _, info := range s.filterSuiteState.totalChats {
		if info.chatType != "Parent" {
			continue
		}
		s.conversationID = info.id
		// pick first parent to send new message
		parentSendingMsg := info.parentIDs[0]
		tok, err := s.genParentToken(parentSendingMsg)
		if err != nil {
			return ctx, err
		}
		ctx, err := godogutil.MultiErrChain(ctx,
			s.makeUsersSubscribeV2Ctx, []string{parentSendingMsg}, []string{tok},
			s.userSendsItemWithContent, "text", "hello world", parentSendingMsg, cpb.UserGroup_USER_GROUP_PARENT,
		)
		if err != nil {
			return ctx, err
		}
		count++
		if count >= countUntil {
			break
		}
	}
	return ctx, nil
}

func (s *suite) filterChatsWithTypeMessageTypeOnly(ctx context.Context, repliedStatus string) (context.Context, error) {
	switch repliedStatus {
	case "Replied":
		s.filterSuiteState.listRequest.TeacherStatus = tpb.TeacherConversationStatus_TEACHER_CONVERSATION_STATUS_REPLIED
	case "All":
		s.filterSuiteState.listRequest.TeacherStatus = tpb.TeacherConversationStatus_TEACHER_CONVERSATION_STATUS_ALL
	case "Unreplied":
		s.filterSuiteState.listRequest.TeacherStatus = tpb.TeacherConversationStatus_TEACHER_CONVERSATION_STATUS_NOT_REPLIED
	default:
		return ctx, fmt.Errorf("unknown message reply status %s in scenario", repliedStatus)
	}
	return ctx, nil
}

func pickRandomFromCourses(courses []string, min, max int) []string {
	permutations := rand.Perm(len(courses))
	pickedCourses := []string{}
	// nolint
	for i := 0; i < rand.Intn(max)+min; i++ {
		pickedCourses = append(pickedCourses, courses[permutations[i]])
	}
	return pickedCourses
}

func (s *suite) studentsChatsAreCreated(ctx context.Context) (context.Context, error) {
	// unrepliedChats := []string{}
	newChats := []chatInfo{}
	for _, studentID := range s.filterSuiteState.studentIDs {
		s.studentID = studentID
		ctx, err := godogutil.MultiErrChain(ctx,
			s.aEvtUserWithMessageAndLanguageAndSchoolID, "CreateStudent", s.filterSuiteState.language, s.commonState.schoolID, []string{s.filterSuiteState.defaultLocationID},
			s.yasuoSendEventEvtUser,
			s.studentMustBeInConversation,
		)
		if err != nil {
			return ctx, err
		}
		// unrepliedChats = append(unrepliedChats, s.conversationID)
		s.filterSuiteState.chatStudentMap[s.conversationID] = studentID
		newChats = append(newChats, chatInfo{
			name:      pseudoNameForStudentChat(studentID, s.filterSuiteState.language),
			id:        s.conversationID,
			chatType:  "Student",
			courses:   s.filterSuiteState.studentCoursesMap[studentID],
			studentID: studentID,
			replied:   false, // replied is false by default
		})
	}
	s.filterSuiteState.totalChats = append(s.filterSuiteState.totalChats, newChats...)
	return ctx, nil
}

func (s *suite) waitForAllParentsChatCreation(ctx context.Context) error {
	newChats := []chatInfo{}
	// time.Sleep(5 * time.Second)
	for studentID, parents := range s.filterSuiteState.studentParentsMap {
		for _, parent := range parents {
			s.filterSuiteState.parentStudentMap[parent] = studentID
		}
		s.parentChats = make(map[string]chatInfo)
		s.childrenIDs = []string{studentID}
		s.parentIDs = parents
		ctx2, err := s.allParentsAreAddedIntoChatGroups(ctx)
		if err != nil {
			return err
		}
		ctx = ctx2
		if len(s.parentChats) != 1 {
			panic(s.parentChats)
		}

		for _, info := range s.parentChats {
			newChats = append(newChats, chatInfo{
				studentID: studentID,
				name:      info.name,
				id:        info.id,
				chatType:  "Parent",
				replied:   false,
				courses:   s.filterSuiteState.studentCoursesMap[studentID],
				parentIDs: parents,
			})

			s.filterSuiteState.chatStudentMap[info.id] = studentID
		}
	}
	s.filterSuiteState.totalChats = append(s.filterSuiteState.totalChats, newChats...)
	return nil
}

func (s *suite) parentsChatsAreCreated(ctx context.Context) (context.Context, error) {
	for studentID, parents := range s.filterSuiteState.studentParentsMap {
		chatName := pseudoNameForParentChat(studentID, s.filterSuiteState.language)
		for _, parentID := range parents {
			request := &bpb.EvtUser{
				Message: &bpb.EvtUser_CreateParent_{
					CreateParent: &bpb.EvtUser_CreateParent{
						StudentId:   studentID,
						StudentName: chatName, // expect this to be chat name
						SchoolId:    s.commonState.schoolID,
						ParentId:    parentID,
					},
				},
			}

			data, err := proto.Marshal(request)
			if err != nil {
				return ctx, err
			}

			_, err = s.JSM.TracedPublishAsync(ctx, "nats.TracedPublishAsync", constants.SubjectUserCreated, data)
			if err != nil {
				return ctx, fmt.Errorf("s.JSM.TracedPublishAsync: %w", err)
			}
		}
	}
	err := s.waitForAllParentsChatCreation(ctx)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (s *suite) studentJoinsCoursesAtLocations(ctx context.Context, studentCoursesMap map[string][]string, loc []string) (context.Context, error) {
	for studentID, courses := range studentCoursesMap {
		evt := &npb.EventStudentPackage{
			StudentPackage: &npb.EventStudentPackage_StudentPackage{
				IsActive:  true,
				StudentId: studentID,
				Package: &npb.EventStudentPackage_Package{
					CourseIds:   courses,
					StartDate:   timestamppb.Now(),
					EndDate:     timestamppb.New(time.Now().AddDate(1, 0, 0)),
					LocationIds: loc,
				},
			},
		}
		data, _ := proto.Marshal(evt)
		_, err := s.JSM.TracedPublishAsync(ctx, "studentJoinsCourses", constants.SubjectStudentPackageEventNats, data)
		if err != nil {
			return ctx, fmt.Errorf("failed to publish event error %w", err)
		}
	}
	return ctx, nil
}

func (s *suite) studentsJoinsCourses(ctx context.Context, studentCoursesMap map[string][]string) (context.Context, error) {
	return s.studentJoinsCoursesAtLocations(ctx, studentCoursesMap, nil)
}

func (s *suite) eachStudentJoinedSomeCoursesAtRandom(ctx context.Context) (context.Context, error) {
	numCourses := 10
	courses := []string{}
	for i := 0; i < numCourses; i++ {
		courses = append(courses, idutil.ULIDNow())
	}
	// at least 1-course package, and multiple-course pakcage
	packages := [][]string{
		{courses[0]},
		{courses[0], courses[1]},
	}
	for i := 0; i < 3; i++ {
		packages = append(packages, pickRandomFromCourses(courses, 2, 3))
	}
	studentCoursesMap := make(map[string][]string)
	for idx, stuID := range s.filterSuiteState.studentIDs {
		studentCoursesMap[stuID] = packages[idx%len(packages)]
	}
	ctx, err := s.studentsJoinsCourses(ctx, studentCoursesMap)
	if err != nil {
		return ctx, err
	}
	s.filterSuiteState.courses = courses
	s.filterSuiteState.studentCoursesMap = studentCoursesMap
	for idx, item := range s.filterSuiteState.totalChats {
		courses := studentCoursesMap[item.studentID]
		s.filterSuiteState.totalChats[idx].courses = courses
	}
	return ctx, nil
}

func (s *suite) studentsAndEachHasFromToParentsRandomly(ctx context.Context, minParent, maxParent int) (context.Context, error) {
	studentNum := 2
	studentIDs := []string{}
	studentParentsMap := make(map[string][]string)
	studentNameMap := make(map[string]string)
	for i := 0; i < studentNum; i++ {
		rand := idutil.ULIDNow()
		name := pseudoNameForStudentChat(rand, s.filterSuiteState.language)
		stu, par, err := s.CommonSuite.CreateStudentWithParent(ctx, []string{s.filterSuiteState.defaultLocationID}, &common.CreateStudentWithParentOpt{
			StudentName: name,
		})
		if err != nil {
			return ctx, err
		}

		studentID := stu.UserProfile.UserId
		studentNameMap[studentID] = name
		parents := []string{par.UserProfile.UserId}
		studentIDs = append(studentIDs, studentID)
		studentParentsMap[studentID] = parents
	}

	s.filterSuiteState.studentIDs = studentIDs
	s.filterSuiteState.studentParentsMap = studentParentsMap
	stuConvMap, err := s.findStudentsConvIDs(ctx, s.filterSuiteState.studentIDs, s.commonState.schoolID, []string{s.filterSuiteState.defaultLocationID})
	if err != nil {
		return ctx, err
	}

	newChats := []chatInfo{}
	for _, stu := range studentIDs {
		conv := stuConvMap[stu]
		s.filterSuiteState.chatStudentMap[conv] = stu
		newChats = append(newChats, chatInfo{
			name:      studentNameMap[stu],
			id:        conv,
			chatType:  "Student",
			courses:   s.filterSuiteState.studentCoursesMap[stu],
			studentID: stu,
			replied:   false, // replied is false by default
		})
	}
	s.filterSuiteState.totalChats = append(s.filterSuiteState.totalChats, newChats...)

	err = s.waitForAllParentsChatCreation(ctx)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}
