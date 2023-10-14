package tom

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/try"
	tom_constants "github.com/manabie-com/backend/internal/tom/constants"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/protobuf/proto"
)

func (s *suite) aNewSchoolIsCreatedWithLocation(ctx context.Context, locIdentifier string) (context.Context, error) {
	ctx, err := s.aNewSchoolIsCreated(ctx)

	if locIdentifier == "default" {
		s.commonState.locationPool[locIdentifier] = s.filterSuiteState.defaultLocationID
		s.CommonSuite.DefaultLocationID = s.filterSuiteState.defaultLocationID
		s.CommonSuite.DefaultLocationTypeID = s.filterSuiteState.defaultLocationTypeID
	} else {
		locationID, locationTypeID, err := s.CommonSuite.CreateLocationWithDB(ctx, s.commonState.schoolID, "org", "", "")
		if err != nil {
			return ctx, err
		}
		s.commonState.locationPool[locIdentifier] = locationID
		s.CommonSuite.DefaultLocationID = locationID
		s.CommonSuite.DefaultLocationTypeID = locationTypeID
	}

	return ctx, err
}
func (s *suite) getLocations(identifiers string) []string {
	ret := []string{}
	for _, identifier := range strings.Split(strings.TrimSpace(identifiers), ",") {
		id, ok := s.commonState.locationPool[identifier]
		if !ok {
			panic(fmt.Sprintf("location %s do not exist in pool, maybe missing scenario step", identifier))
		}
		ret = append(ret, id)
	}
	return ret
}

func (s *suite) getCourse(identifier string) string {
	courseID, ok := s.commonState.coursePool[identifier]
	if !ok {
		panic(fmt.Sprintf("unknown course %s", identifier))
	}
	return courseID
}

func (s *suite) getOrCreateCourse(identifier string) string {
	_, ok := s.commonState.coursePool[identifier]
	if !ok {
		s.commonState.coursePool[identifier] = idutil.ULIDNow()
	}
	return s.commonState.coursePool[identifier]
}

func (s *suite) getChatInfoFromPool(identifier string) chatInfo {
	chatinfo, ok := s.commonState.parentChatPool[identifier]
	if !ok {
		chatinfo, ok = s.commonState.studentChatPool[identifier]
		if !ok {
			panic(fmt.Sprintf("chat %s do not exist in pool, maybe missing scenario step", identifier))
		}
	}
	return chatinfo
}
func (s *suite) getChatIDsFromPool(identifiers string) []string {
	ret := []string{}
	for _, identifier := range strings.Split(strings.TrimSpace(identifiers), ",") {
		chatinfo, ok := s.commonState.parentChatPool[identifier]
		if !ok {
			chatinfo, ok = s.commonState.studentChatPool[identifier]
			if !ok {
				panic(fmt.Sprintf("chat %s do not exist in pool, maybe missing scenario step", identifier))
			}
		}
		ret = append(ret, chatinfo.id)
	}
	return ret
}

func (s *suite) getLocation(identifier string) string {
	id, ok := s.commonState.locationPool[identifier]
	if !ok {
		panic(fmt.Sprintf("location %s do not exist in pool, maybe missing scenario step", identifier))
	}
	return id
}

func (s *suite) locationsChildrenOfLocationAreCreated(ctx context.Context, childrenLocs string, parentLoc string) (context.Context, error) {
	parentLocID := s.getLocation(parentLoc)
	for _, loc := range strings.Split(strings.TrimSpace(childrenLocs), ",") {
		locationID, _, err := s.CommonSuite.CreateLocationWithDB(ctx, s.commonState.schoolID, "center", parentLocID, s.CommonSuite.DefaultLocationTypeID)
		if err != nil {
			return ctx, err
		}
		s.commonState.locationPool[loc] = locationID
	}

	return ctx, nil
}

func (s *suite) createStudentParentByAPI(ctx context.Context, locIDs []string) (stu *upb.Student, par *upb.Parent, err error) {
	if !s.CommonSuite.ContextHasToken(ctx) {
		ctx2, err := s.CommonSuite.ASignedInWithSchool(ctx, "school admin", int32ResourcePathFromCtx(ctx))
		if err != nil {
			return nil, nil, err
		}
		ctx = ctx2
	}
	return s.CommonSuite.CreateStudentWithParent(ctx, locIDs, nil)
}
func (s *suite) findStudentParentChat(ctx context.Context, stuid, parid string, schoolID string, locs []string,
	student *upb.Student,
	parent *upb.Parent,
) (context.Context, []chatInfo, error) {
	stuConvMap, err := s.findStudentsConvIDs(ctx, []string{stuid}, schoolID, locs)
	if err != nil {
		return ctx, nil, err
	}
	stuConv := stuConvMap[stuid]
	stuchat := chatInfo{
		name:      pseudoNameForStudentChat(stuid, langEng),
		id:        stuConv,
		chatType:  "Student",
		studentID: stuid,
		student:   student,
		replied:   false, // replied is false by default
	}
	parchatid, parchatname, err := s.findCreatedParentChat(ctx, stuid, []string{parid}, schoolID, locs)
	if err != nil {
		return ctx, nil, err
	}
	parchat := chatInfo{
		name:      parchatname,
		id:        parchatid,
		chatType:  "Parent",
		replied:   false,
		parentIDs: []string{parid},
		parents:   []*upb.Parent{parent},
	}
	return ctx, []chatInfo{stuchat, parchat}, nil
}

func (s *suite) filterChatsWithFilterCombination(ctx context.Context, filterCombination string) (context.Context, error) {
	for _, text := range strings.Split(filterCombination, ",") {
		parts := strings.Split(text, " ")
		filter, param := parts[0], parts[1]
		switch filter {
		case "contact":
			err := s.filterChatsWithTypeContactTypeOnly(param)
			if err != nil {
				return ctx, err
			}
		case "courses":
			ids := []string{}
			for _, courseLabel := range strings.Split(param, "-") {
				ids = append(ids, s.getCourse(courseLabel))
			}
			s.filterSuiteState.listRequest.CourseIds = ids
		default:
			return ctx, fmt.Errorf("unknown filter %s", text)
		}
	}
	return ctx, nil
}
func (s *suite) mappingsBetweenStudentAndCourse(ctx context.Context, studentCourseMappings string) (context.Context, error) {
	for _, item := range strings.Split(studentCourseMappings, ",") {
		parts := strings.Split(item, "-")
		if len(parts) != 3 {
			return ctx, fmt.Errorf("invalid input %s", item)
		}
		stuLabel, courseLabel, locLabel := parts[0], parts[1], parts[2]
		stuID := s.getChatInfoFromPool(stuLabel).studentID
		stucoursemap := map[string][]string{
			stuID: {s.getOrCreateCourse(courseLabel)},
		}
		ctx, err := s.studentJoinsCoursesAtLocations(ctx, stucoursemap, []string{s.getLocation(locLabel)})
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func (s *suite) studentChatParentChatSameParentWithChatAreCreatedWithLocations(ctx context.Context, stuChatLabel, parChatLabel, sharedParentLabel, locs string) (context.Context, error) {
	parChat := s.getChatInfoFromPool(sharedParentLabel)
	par := parChat.parents[0]

	locations := s.getLocations(locs)
	stu, err := s.CommonSuite.CreateStudent(ctx, locations, nil)
	if err != nil {
		return ctx, err
	}
	err = s.CommonSuite.UpdateStudentParent(ctx, stu.UserProfile.UserId, par.UserProfile.UserId, par.UserProfile.Email)
	if err != nil {
		return ctx, err
	}
	ctx, chatinfos, err := s.findStudentParentChat(ctx, stu.UserProfile.UserId, par.UserProfile.UserId, s.commonState.schoolID, locations, stu, par)
	if err != nil {
		return ctx, err
	}

	stuchat, parchat := chatinfos[0], chatinfos[1]
	s.commonState.studentChatPool[stuChatLabel] = stuchat
	s.commonState.parentChatPool[parChatLabel] = parchat

	s.filterSuiteState.totalChats = append(s.filterSuiteState.totalChats, chatinfos...)
	return ctx, nil
}

func (s *suite) studentChatAndParentChatAreCreatedWithLocations(ctx context.Context, stuChatLabel, parChatLabel, locs string) (context.Context, error) {
	var locations []string
	if locs != "" {
		locations = s.getLocations(locs)
	}
	stu, par, err := s.createStudentParentByAPI(ctx, locations)
	if err != nil {
		return ctx, err
	}
	ctx, chatinfos, err := s.findStudentParentChat(ctx, stu.UserProfile.UserId, par.UserProfile.UserId, s.commonState.schoolID, locations, stu, par)
	if err != nil {
		return ctx, err
	}
	stuchat, parchat := chatinfos[0], chatinfos[1]
	s.commonState.studentChatPool[stuChatLabel] = stuchat
	s.commonState.studentChatPool[parChatLabel] = parchat
	// Publish all event first, then start checking db

	s.filterSuiteState.totalChats = append(s.filterSuiteState.totalChats, chatinfos...)
	return ctx, nil
}

func (s *suite) filterChatsWithLocation(ctx context.Context, locations string) (context.Context, error) {
	if locations != "" {
		s.filterSuiteState.listRequest.LocationIds = s.getLocations(locations)
	}
	return ctx, nil
}

// ids = s1,s2,p1,p2 => exactIDs = real id of those chats in db
func (s *suite) chatsWithIdsAreReturned(ctx context.Context, ids string) (context.Context, error) {
	exactIDs := []string{}
	if ids != "" {
		exactIDs = s.getChatIDsFromPool(ids)
	}

	filterExactID := func(previous []chatInfo) ([]chatInfo, error) {
		// keep if given chat has id in the expect ids list
		filtered := filterChats(previous, func(c chatInfo) bool { return intersect(exactIDs, []string{c.id}) })
		return filtered, nil
	}
	s.filterSuiteState.filterExpectedChats = append(s.filterSuiteState.filterExpectedChats, filterExactID)
	return ctx, s.sendListConversatioLocationsAndAssert(ctx, s.assertConversationListResultWithIDs)
}

func (s *suite) sendListConversatioLocationsAndAssert(ctx context.Context, fnAssert func(*tpb.ListConversationsInSchoolResponse, []chatInfo) error) error {
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
	svc := tpb.NewChatReaderServiceClient(s.Conn)

	enventually := func(attempt int) (bool, error) {
		api := svc.ListConversationsInSchoolV2
		res, err := api(contextWithToken(context.Background(), token), s.filterSuiteState.listRequest)
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
func (s *suite) usermgmtSendEventUpsertUserProfileForStudentOfChatWithLocations(ctx context.Context, chatlabel string, locs string) (context.Context, error) {
	chatinfo := s.getChatInfoFromPool(chatlabel)
	studentID := chatinfo.student.UserProfile.UserId

	msg := &upb.EvtUser{
		Message: &upb.EvtUser_UpdateStudent_{
			UpdateStudent: &upb.EvtUser_UpdateStudent{
				StudentId:         studentID,
				DeviceToken:       "new-token",
				AllowNotification: true,
				Name:              fmt.Sprintf("user-name-%s", studentID),
				LocationIds:       s.getLocations(locs),
			},
		},
	}
	s.Request = msg

	data, err := proto.Marshal(msg)
	if err != nil {
		return ctx, err
	}
	_, err = s.JSM.TracedPublish(ctx, "usermgmtSendEventUpsertUserProfileForStudentOfChatWithLocations", constants.SubjectUserUpdated, data)
	return ctx, err
}

func (s *suite) disableChatInLocationConfigTable(ctx context.Context, conversationTypeStr string, locIdentifier string) (context.Context, error) {
	locationID := s.commonState.locationPool[locIdentifier]
	resourcePath, _ := interceptors.ResourcePathFromContext(ctx)

	var configKey string
	switch conversationTypeStr {
	case "student":
		configKey = tom_constants.ChatConfigKeyStudent
	case "parent":
		configKey = tom_constants.ChatConfigKeyParent
	}

	query := `UPDATE public.location_configuration_value
	set config_value=$1
	WHERE config_key=$2 and resource_path = $3 and location_id = $4;`

	_, err := s.masterMgmtDBTrace.Exec(ctx, query, "false", configKey, resourcePath, locationID)

	return ctx, err
}
