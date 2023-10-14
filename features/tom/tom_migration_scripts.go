package tom

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	bobcmd "github.com/manabie-com/backend/cmd/server/bob"
	tomcmd "github.com/manabie-com/backend/cmd/server/tom"
	yasuocmd "github.com/manabie-com/backend/cmd/server/yasuo"
	"github.com/manabie-com/backend/features/bob"
	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/usermgmt"
	bobCfg "github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/elastic"
	"github.com/manabie-com/backend/internal/golibs/godogutil"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	tomCfg "github.com/manabie-com/backend/internal/tom/configurations"
	"github.com/manabie-com/backend/internal/tom/constants"
	"github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/tom/repositories"
	yasuoCfg "github.com/manabie-com/backend/internal/yasuo/configurations"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/cucumber/godog"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var (
	yasuoSyncConfig yasuoCfg.Config
	bobSyncConfig   bobCfg.Config
	tomSyncConfig   tomCfg.Config
)

type tomMigrationScriptKey struct{}

func TomMigrationScriptStateFromCtx(ctx context.Context) *StepState {
	return ctx.Value(tomMigrationScriptKey{}).(*StepState)
}

func TomMigrationScriptStateToCtx(ctx context.Context, state *StepState) context.Context {
	return context.WithValue(ctx, tomMigrationScriptKey{}, state)
}

func tomMigrationScriptSteps(ctx *godog.ScenarioContext, s *TomMigrationScripts) {
	steps := map[string]interface{}{
		// support chat sync
		`^remove all location of conversations in current school$`:                                               s.removeAllLocationOfConversationsInCurrentSchool,
		`^a new school is created with location default "([^"]*)"$`:                                              s.aNewSchoolIsCreatedWithLocationDefault,
		`^accounts and chats of "([^"]*)" students are created with location "([^"]*)", each has (\d+) parents$`: s.accountsAndChatsOfStudentsAreCreatedWithLocationEachHasParents,
		`^locations "([^"]*)" children of location "([^"]*)" are created$`:                                       s.locationsChildrenOfLocationAreCreated,
		`^there are "([^"]*)" support chats with exact locations "([^"]*)"$`:                                     s.thereAreSupportChatsWithExactLocations,
		`^a new school is created$`: s.aNewSchoolIsCreated,
		`^accounts and chats of "([^"]*)" students are created, each has (\d+) parents$`: s.accountsAndChatsOfStudentsAreCreatedEachHasParents,
		`^force remove all support conversations of this school$`:                        s.forceRemoveAllSupportConversationsOfThisSchool,
		`^force remove "([^"]*)" student chats and associated parent chats$`:             s.forceRemoveStudentChatsAndAssociatedParentChats,
		`^total conversations in this school is "([^"]*)"$`:                              s.totalConversationsInThisSchoolIs,
		`^number of parent membership created equal "([^"]*)"$`:                          s.numberOfParentMembershipCreatedEqual,
		`^number of student chats created equal "([^"]*)"$`:                              s.numberOfStudentChatsCreatedEqual,

		// lesson chat sync
		`^accounts of "([^"]*)" students are created$`:                            s.accountsOfStudentsAreCreated,
		`^force remove all lesson conversations$`:                                 s.forceRemoveAllLessonConversations,
		`^"([^"]*)" live lesson chats are created including all of the students$`: s.liveLessonChatsAreCreatedIncludingAllOfTheStudents,
		`^number of lesson chats created equal "([^"]*)"$`:                        s.numberOfLessonChatsCreatedEqual,
		`^number of lesson chats updated equal "([^"]*)"$`:                        s.numberOfLessonChatsUpdatedEqual,
		`^force remove all students from "([^"]*)" remaining lesson chats$`:       s.forceRemoveAllStudentsFromRemainingLessonChats,
		`^force remove "([^"]*)" lesson chats$`:                                   s.forceRemoveLessonChat,
		`^students are added to lesson chats after sync$`:                         s.studentsAreAddedToLessonChatsAfterSync,

		`^db has "([^"]*)" user device token with resource path of this school$`: s.dbHasUserDeviceTokenWithResourcePathOfThisSchool,
		`^those users have resource path in Bob DB$`:                             s.thoseUsersHaveResourcePathInBobDB,
		`^those users have device token in Tom DB without resource path$`:        s.thoseUsersHaveDeviceTokenInTomDBWithoutResourcePath,

		`^all lesson conversation have correct resource path$`:                                s.allLessonConversationHaveCorrectResourcePath,
		`^run "([^"]*)" script$`:                                                              s.runScript,
		`^a ctx with resource_path of current school$`:                                        s.aCtxWithResourcepathOfCurrentSchool,
		`^those lesson chats have empty resource path$`:                                       s.thoseLessonChatsHaveEmptyResourcePath,
		`^another school "([^"]*)" is created with "([^"]*)" conversations on elasticsearch$`: s.anotherSchoolIsCreatedWithConversationsOnElasticsearch,
		`^delete chat data on elasticsearch for current school and school "([^"]*)"$`:         s.deleteChatDataOnElasticsearchForCurrentSchoolAndSchool,
		`^listing chat on elasticsearch with "([^"]*)" returns "([^"]*)" items$`:              s.listingChatOnElasticsearchWithReturnsItems,
		`^wait for "([^"]*)" chats of those schools created on elasticsearch$`:                s.waitForChatsOfThoseSchoolsCreatedOnElasticsearch,
		`^wait for elasticsearch consumers to ack all msg from stream chat$`:                  s.waitForElasticsearchConsumersToAckAllMsgFromStreamChat,
	}

	for pattern, stepFunc := range steps {
		ctx.Step(pattern, stepFunc)
	}
}

func (s *TomMigrationScripts) aCtxWithResourcepathOfCurrentSchool(ctx context.Context) (context.Context, error) {
	stepState := TomMigrationScriptStateFromCtx(ctx)

	school := stepState.schoolID
	return contextWithResourcePath(ctx, school), nil
}

func (s *TomMigrationScripts) accountsOfStudentsAreCreated(ctx context.Context, numstudent int) (context.Context, error) {
	ctx, err := EnsureSchoolAdminToken(ctx, s.commonSuite)
	if err != nil {
		return ctx, err
	}

	stepState := TomMigrationScriptStateFromCtx(ctx)

	// school := stepState.schoolID
	totalStudentsInLessons := make([]string, 0, numstudent)

	for i := 0; i < numstudent; i++ {
		stu, err := s.commonSuite.CreateStudent(ctx, []string{stepState.orgLocation}, nil)
		if err != nil {
			return ctx, err
		}
		id := stu.UserProfile.UserId
		totalStudentsInLessons = append(totalStudentsInLessons, id)
	}
	stepState.totalStudentsInLesson = totalStudentsInLessons
	return TomMigrationScriptStateToCtx(ctx, stepState), nil
}

func newOldTomGandalfSuite(conf *common.Config, c *common.Connections) *TomMigrationScripts {
	s := &TomMigrationScripts{
		Connections:  c,
		searchClient: searchClient,
		StepState: StepState{
			lessonConvMap:        map[string]string{},
			removedLessonConvMap: map[string]string{},
			locationPool:         map[string]string{},
		},
		jsm:         c.JSM,
		ZapLogger:   zapLogger,
		ApplicantID: applicantID,
		commonSuite: newCommonSuite(),
	}

	// s.newYasuoSuite(c)
	s.newBobSuite(c)
	s.newUsermgmtSuite(c)
	yasuoSyncConfig = yasuoCfg.Config{
		Common:     conf.Common,
		PostgresV2: conf.PostgresV2,
		NatsJS:     conf.NatsJS,
	}
	tomSyncConfig = tomCfg.Config{
		Common:        conf.Common,
		PostgresV2:    conf.PostgresV2,
		NatsJS:        conf.NatsJS,
		ElasticSearch: conf.ElasticSearch,
	}
	bobSyncConfig = bobCfg.Config{
		Common:     conf.Common,
		PostgresV2: conf.PostgresV2,
		NatsJS:     conf.NatsJS,
	}
	return s
}

func (s *TomMigrationScripts) newBobSuite(c *common.Connections) {
	s.bobSuite = &bob.Suite{}
	s.bobSuite.DB = c.BobDB
	s.bobSuite.DBPostgres = c.BobPostgresDB
	s.bobSuite.Conn = c.BobConn
	s.bobSuite.ZapLogger = s.ZapLogger
	s.bobSuite.JSM = c.JSM
	s.bobSuite.ShamirConn = c.ShamirConn
	s.bobSuite.ApplicantID = s.ApplicantID
}

func (s *TomMigrationScripts) newUsermgmtSuite(c *common.Connections) {
	s.userMgmtSuite = usermgmt.NewSuite(c.UserMgmtConn, c.ShamirConn, c.BobDBTrace, s.ZapLogger, firebaseAddr, applicantID)
}

type TomMigrationScripts struct {
	bobSuite      *bob.Suite
	userMgmtSuite *usermgmt.Suite
	ZapLogger     *zap.Logger
	*common.Connections
	commonSuite  *common.Suite
	searchClient *elastic.SearchFactoryImpl
	StepState
	jsm         nats.JetStreamManagement
	ApplicantID string
}

//nolint:structcheck
type StepState struct {
	// currentSchool string
	school2 string

	userIDs        []string
	locationPool   map[string]string
	schoolID       string
	orgLocation    string
	createdStudent int
	createdParent  int
	syncedStudent  int
	syncedParent   int
	// YasuoStepState YasuoStepState

	// totalLessons          []string
	totalStudentsInLesson []string
	// totalLessonConversation           []string
	lessonConvMap        map[string]string
	removedLessonConvMap map[string]string
	createdLessonNum     int
	updatedLessonNum     int
}

func (s *TomMigrationScripts) locationsChildrenOfLocationAreCreated(ctx context.Context, childrenLocs string, parentLoc string) (context.Context, error) {
	st := TomMigrationScriptStateFromCtx(ctx)
	parentLocID := s.getLocation(ctx, parentLoc)
	for _, loc := range strings.Split(strings.TrimSpace(childrenLocs), ",") {
		locationID, err := s.generateNewLocation(ctx, st.schoolID, parentLocID)
		if err != nil {
			return ctx, err
		}
		st.locationPool[loc] = locationID
	}
	return TomMigrationScriptStateToCtx(ctx, st), nil
}

func (s *TomMigrationScripts) generateNewLocation(ctx context.Context, resourcePath string, parentLocID string) (string, error) {
	locationID := idutil.ULIDNow()
	parentLocation := pgtype.Text{Status: pgtype.Null}
	accessPath := locationID
	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})
	if parentLocID != "" {
		var parentAp string
		query := `
			SELECT access_path FROM locations WHERE location_id = $1 AND resource_path = $2
		`
		if err := s.BobPostgresDBTrace.QueryRow(ctx2, query, parentLocID, resourcePath).Scan(&parentAp); err != nil {
			return "", fmt.Errorf("finding access_path for of parent location %s", parentLocID)
		}
		accessPath = fmt.Sprintf("%s/%s", parentAp, locationID)
		parentLocation = database.Text(parentLocID)
	}

	stmt := `
		INSERT INTO public.locations
			(location_id, name, location_type, parent_location_id, updated_at, created_at, access_path, resource_path)
		VALUES ($1, $1, NULL, $2, now(), now(), $3, $4)
	`

	if _, err := s.BobPostgresDBTrace.Exec(ctx2, stmt, locationID, parentLocation, accessPath, resourcePath); err != nil {
		return "", err
	}
	return locationID, nil
}

func (s *TomMigrationScripts) getLocation(ctx context.Context, identifier string) string {
	st := TomMigrationScriptStateFromCtx(ctx)
	id, ok := st.locationPool[identifier]
	if !ok {
		panic(fmt.Sprintf("location %s do not exist in pool, maybe missing scenario step", identifier))
	}
	return id
}

func (s *TomMigrationScripts) getLocations(ctx context.Context, identifiers string) []string {
	st := TomMigrationScriptStateFromCtx(ctx)
	ret := []string{}
	for _, identifier := range strings.Split(strings.TrimSpace(identifiers), ",") {
		id, ok := st.locationPool[identifier]
		if !ok {
			panic(fmt.Sprintf("location %s do not exist in pool, maybe missing scenario step", identifier))
		}
		ret = append(ret, id)
	}
	return ret
}

var searchbyschoolquery = `
{
	"query": {
		"bool": {
		"filter": [
			{
			"terms": {
				"owner": [
					%s
				]
			}
			}
		]
		}
	}
}`

func (s *TomMigrationScripts) listingChatOnElasticsearchWithReturnsItems(ctx context.Context, schooltype string, expectChat int) (context.Context, error) {
	stepState := TomMigrationScriptStateFromCtx(ctx)
	var schoolparam, resourcepath string
	switch schooltype {
	case "current school":
		schoolparam = fmt.Sprintf(`"%s"`, stepState.schoolID)
		resourcepath = stepState.schoolID
	case "school 2":
		schoolparam = fmt.Sprintf(`"%s"`, stepState.school2)
		resourcepath = stepState.school2
	default:
		return ctx, fmt.Errorf("unknown school %s", schooltype)
	}
	err := try.Do(func(attempt int) (bool, error) {
		rdr := strings.NewReader(fmt.Sprintf(searchbyschoolquery, schoolparam))
		res, err := s.searchClient.Search(ctx, constants.ESConversationIndexName, rdr)
		if err != nil {
			return false, err
		}
		defer res.Body.Close()
		var count int
		type Cont struct {
			ResourcePath string `json:"resource_path"`
		}
		err = elastic.ParseSearchResponse(res.Body, func(h *elastic.SearchHit) error {
			var c Cont
			err := json.Unmarshal(h.Source, &c)
			if err != nil {
				return err
			}
			if c.ResourcePath != resourcepath {
				return fmt.Errorf("return document has resource path %s instead of %s", c.ResourcePath, resourcepath)
			}
			count++
			return nil
		})
		if err != nil {
			return false, err
		}
		// eventual consistent
		if count != expectChat {
			time.Sleep(2 * time.Second)
			return attempt < 10, fmt.Errorf("count query result not match expected: %d vs %d", count, expectChat)
		}
		return false, nil
	})
	return ctx, err
}

func (s *TomMigrationScripts) waitForChatsOfThoseSchoolsCreatedOnElasticsearch(ctx context.Context, totalchat int) (context.Context, error) {
	stepState := TomMigrationScriptStateFromCtx(ctx)
	schoolparam := fmt.Sprintf(`"%s","%s"`, stepState.schoolID, stepState.school2)
	err := try.Do(func(attempt int) (bool, error) {
		rdr := strings.NewReader(fmt.Sprintf(searchbyschoolquery, schoolparam))
		res, err := s.searchClient.Count(constants.ESConversationIndexName, rdr)
		if err != nil {
			return false, err
		}
		defer res.Body.Close()
		type Cont struct {
			Count int `json:"count"`
		}
		var c Cont
		err = json.NewDecoder(res.Body).Decode(&c)
		if err != nil {
			return false, err
		}
		if c.Count != totalchat {
			time.Sleep(2 * time.Second)
			return attempt < 10, fmt.Errorf("count query result not match expected: %d vs %d", c.Count, totalchat)
		}
		return false, nil
	})
	return ctx, err
}

func (s *TomMigrationScripts) waitForElasticsearchConsumersToAckAllMsgFromStreamChat(ctx context.Context) (context.Context, error) {
	str := "chat"
	erg := errgroup.Group{}
	cons := []string{
		"durable_chat_chat_created_elastic",
		"durable_chat_chat_members_updated_elastic",
		"durable_chat_chat_message_created_elastic",
		"durable_chat_chat_updated_elastic",
	}
	for idx := range cons {
		consumer := cons[idx]
		erg.Go(func() error {
			return try.Do(func(attempt int) (bool, error) {
				info, err := s.jsm.GetJS().ConsumerInfo(str, consumer)
				if err != nil {
					return false, err
				}
				delivered := info.Delivered.Consumer
				acked := info.AckFloor.Consumer
				if delivered != acked {
					time.Sleep(2 * time.Second)
					return attempt < 5, fmt.Errorf("consumer is still handling incoming msg from nats")
				}
				return false, nil
			})
		})
	}
	return ctx, erg.Wait()
}

func (s *TomMigrationScripts) deleteChatDataOnElasticsearchForCurrentSchoolAndSchool(ctx context.Context, _ string) (context.Context, error) {
	// sleep a bit to let data on Elasticsearch sync
	time.Sleep(2 * time.Second)
	stepState := TomMigrationScriptStateFromCtx(ctx)
	schoolparam := fmt.Sprintf(`"%s","%s"`, stepState.schoolID, stepState.school2)
	rdr := strings.NewReader(fmt.Sprintf(searchbyschoolquery, schoolparam))
	res, err := s.searchClient.DeletebyQuery(constants.ESConversationIndexName, rdr)
	if err != nil {
		return ctx, err
	}
	defer res.Body.Close()
	bs, err := io.ReadAll(res.Body)
	if err != nil {
		return ctx, err
	}
	fmt.Printf("deleting conversation %s using delete by query res: %v\n", schoolparam, string(bs))

	return ctx, nil
}

func (s *TomMigrationScripts) anotherSchoolIsCreatedWithConversationsOnElasticsearch(ctx context.Context, _ string, numconversation int) (context.Context, error) {
	stepState := TomMigrationScriptStateFromCtx(ctx)

	schoolID, locationID, _, err := s.commonSuite.NewOrgWithOrgLocation(ctx)
	if err != nil {
		return ctx, err
	}

	tempCtx, err := s.commonSuite.ASignedInWithSchool(contextWithResourcePath(ctx, i32ToStr(schoolID)), "school admin", schoolID)
	if err != nil {
		return ctx, err
	}

	stepState.school2 = strconv.Itoa(int(schoolID))

	for i := 0; i < numconversation; i++ {
		_, err := s.commonSuite.CreateStudent(tempCtx, []string{locationID}, nil)
		if err != nil {
			return ctx, fmt.Errorf("s.commonsuite.CreateStudent: %w", err)
		}
	}
	return TomMigrationScriptStateToCtx(ctx, stepState), nil
}

func stringMapKeys(m map[string]string) []string {
	ret := make([]string, 0, len(m))
	for k := range m {
		ret = append(ret, k)
	}
	return ret
}

func stringMapValues(m map[string]string) []string {
	ret := make([]string, 0, len(m))
	for _, v := range m {
		ret = append(ret, v)
	}
	return ret
}

func (s *TomMigrationScripts) studentsAreAddedToLessonChatsAfterSync(ctx context.Context) (context.Context, error) {
	st := TomMigrationScriptStateFromCtx(ctx)
	studentCheckList := map[string]struct{}{}
	for _, id := range st.totalStudentsInLesson {
		studentCheckList[id] = struct{}{}
	}
	err := try.Do(func(attempt int) (bool, error) {
		res, err := tpb.NewConversationReaderServiceClient(s.TomConn).ListConversationByLessons(ctx, &tpb.ListConversationByLessonsRequest{
			LessonIds:      stringMapKeys(st.lessonConvMap),
			OrganizationId: resourcePathFromCtx(ctx),
		})
		if err != nil {
			return false, err
		}
		if len(res.GetConversations()) != len(st.lessonConvMap) {
			time.Sleep(2 * time.Second)
			return attempt < 5, fmt.Errorf("still %d out of %d lesson synced", len(res.GetConversations()), len(st.lessonConvMap))
		}
		for _, item := range res.GetConversations() {
			if len(item.GetUsers()) > len(studentCheckList) {
				return false, fmt.Errorf("after synced has %d users,more than %d expected", len(item.GetUsers()), len(studentCheckList))
			}
			if len(item.GetUsers()) < len(studentCheckList) {
				time.Sleep(2 * time.Second)
				return attempt < 5, fmt.Errorf("still %d out of %d synced", len(item.GetUsers()), len(studentCheckList))
			}
			for _, u := range item.GetUsers() {
				_, exist := studentCheckList[u.GetId()]
				if !exist {
					return false, fmt.Errorf("invalid user in lesson chat after synced")
				}
			}
		}
		return false, nil
	})
	return TomMigrationScriptStateToCtx(ctx, st), err
}

func (s *TomMigrationScripts) forceRemoveAllStudentsFromRemainingLessonChats(ctx context.Context, removed int) (context.Context, error) {
	st := TomMigrationScriptStateFromCtx(ctx)
	remaining := []string{}
	for lesson := range st.lessonConvMap {
		if _, exist := st.removedLessonConvMap[lesson]; exist {
			continue
		}
		remaining = append(remaining, lesson)
	}
	if len(remaining) < removed {
		return ctx, fmt.Errorf("%d remaining lessons are not sufficient for %d removal", len(remaining), removed)
	}
	removeAllScripts := []string{
		`DELETE FROM conversation_members WHERE conversation_members.conversation_id in (
			select conversation_id from conversations c left join conversation_lesson cl using(conversation_id)
			where  cl.lesson_id=ANY($1)
		);`,
	}
	for _, item := range removeAllScripts {
		_, err := s.TomPostgresDB.Exec(context.Background(), item, database.TextArray(remaining[:removed]))
		if err != nil {
			return ctx, err
		}
	}
	return TomMigrationScriptStateToCtx(ctx, st), nil
}

func (s *TomMigrationScripts) forceRemoveLessonChat(ctx context.Context, removedLessons int) (context.Context, error) {
	st := TomMigrationScriptStateFromCtx(ctx)
	lessonConvMap := st.lessonConvMap
	count := 0
	removedLesson := make([]string, 0, removedLessons)
	removedConv := make([]string, 0, removedLessons)
	removedLessonConversationMap := map[string]string{}
	for lesson, conv := range lessonConvMap {
		removedLesson = append(removedLesson, lesson)
		removedConv = append(removedConv, conv)
		removedLessonConversationMap[lesson] = conv
		count++
		if count == removedLessons {
			break
		}
	}
	ctx, err := s.forceRemoveLessonConversation(ctx, removedLesson, removedConv)
	if err != nil {
		return ctx, err
	}
	st.removedLessonConvMap = removedLessonConversationMap
	return TomMigrationScriptStateToCtx(ctx, st), nil
}

func (s *TomMigrationScripts) numberOfLessonChatsUpdatedEqual(ctx context.Context, want int) (context.Context, error) {
	st := TomMigrationScriptStateFromCtx(ctx)
	if st.updatedLessonNum != want {
		return ctx, fmt.Errorf("want %d lesson chat updated, has %d", want, st.updatedLessonNum)
	}
	return ctx, nil
}

func (s *TomMigrationScripts) numberOfLessonChatsCreatedEqual(ctx context.Context, want int) (context.Context, error) {
	st := TomMigrationScriptStateFromCtx(ctx)
	if st.createdLessonNum != want {
		return ctx, fmt.Errorf("want %d lesson chat created, has %d", want, st.createdLessonNum)
	}
	return ctx, nil
}

func (s *TomMigrationScripts) forceRemoveLessonConversation(ctx context.Context, lessons []string, convs []string) (context.Context, error) {
	st := TomMigrationScriptStateFromCtx(ctx)
	removeAllScripts := []string{
		`DELETE FROM messages m WHERE m.conversation_id in (
			select conversation_id from conversations c left join conversation_lesson cl using(conversation_id)
			where  cl.lesson_id=ANY($1)
		);`,
		`DELETE FROM conversation_members WHERE conversation_members.conversation_id in (
			select conversation_id from conversations c left join conversation_lesson cl using(conversation_id)
			where  cl.lesson_id=ANY($1)
		);`,
		`DELETE FROM conversation_lesson WHERE lesson_id=ANY($1);`,
	}
	for _, item := range removeAllScripts {
		_, err := s.TomPostgresDB.Exec(context.Background(), item, database.TextArray(lessons))
		if err != nil {
			return ctx, err
		}
	}
	_, err := s.TomPostgresDB.Exec(context.Background(), "DELETE FROM conversations WHERE conversation_id=ANY($1)", database.TextArray(convs))
	if err != nil {
		return ctx, err
	}
	return TomMigrationScriptStateToCtx(ctx, st), nil
}

func (s *TomMigrationScripts) forceRemoveAllLessonConversations(ctx context.Context) (context.Context, error) {
	st := TomMigrationScriptStateFromCtx(ctx)
	convs := make([]string, 0, len(st.lessonConvMap))
	lessons := make([]string, 0, len(st.lessonConvMap))
	for lesson, conv := range st.lessonConvMap {
		convs = append(convs, conv)
		lessons = append(lessons, lesson)
	}
	return s.forceRemoveLessonConversation(ctx, lessons, convs)
}

func (s *TomMigrationScripts) liveLessonChatsAreCreatedIncludingAllOfTheStudents(ctx context.Context, numLesson int) (context.Context, error) {
	st := TomMigrationScriptStateFromCtx(ctx)

	bobstate := bob.StepStateFromContext(ctx)
	bobstate.StudentIds = st.totalStudentsInLesson
	bobstate.CurrentSchoolID = strToI32(st.schoolID)
	ctx = bob.StepStateToContext(ctx, bobstate)

	ctx, err := godogutil.MultiErrChain(ctx,
		s.bobSuite.CreateTeacherAccounts,
		s.bobSuite.CreateLiveCourse,
		s.bobSuite.SignedInSchoolAdmin,
	)
	if err != nil {
		return ctx, err
	}
	lessons := make([]string, 0, numLesson)
	for i := 0; i < numLesson; i++ {
		lessonName := fmt.Sprintf("lesson %d", i)
		start := time.Now().Format(time.RFC3339)
		end := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

		ctx, err := godogutil.MultiErrChain(ctx,
			s.bobSuite.UserCreateLiveLesson, lessonName, start, end, "",
			s.bobSuite.ReturnsStatusCode, "OK",
		)
		if err != nil {
			return ctx, err
		}
		bobstate := bob.StepStateFromContext(ctx)
		lessonID := bobstate.Response.(*bpb.CreateLiveLessonResponse).Id
		lessons = append(lessons, lessonID)
	}
	checkLessonSyncQuery := `
select cl.conversation_id,cl.lesson_id,count(cm.user_id) from conversation_lesson cl left join
conversation_members cm using (conversation_id) where cl.lesson_id=ANY($1)
group by cl.lesson_id;
	`
	finalLessonConvMap := map[string]string{}
	err = try.Do(func(attempt int) (bool, error) {
		lessonConvMap := map[string]string{}
		rows, err := s.TomDB.Query(ctx, checkLessonSyncQuery, database.TextArray(lessons))
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return false, err
			}
			time.Sleep(2 * time.Second)
			return attempt < 5, err
		}
		defer rows.Close()
		for rows.Next() {
			var totalUser pgtype.Int8
			var convID, lessonID pgtype.Text
			err := rows.Scan(&convID, &lessonID, &totalUser)
			if err != nil {
				return false, err
			}
			if int(totalUser.Int) != len(st.totalStudentsInLesson) {
				time.Sleep(2 * time.Second)
				return attempt < 5, fmt.Errorf("out of sync student in lesson")
			}
			lessonConvMap[lessonID.String] = convID.String
		}
		if len(lessonConvMap) != len(lessons) {
			time.Sleep(2 * time.Second)
			return attempt < 5, fmt.Errorf("want %d lesson chats in db, has %d", len(lessons), len(lessonConvMap))
		}
		finalLessonConvMap = lessonConvMap
		return false, nil
	})
	if err != nil {
		return ctx, err
	}
	st.lessonConvMap = finalLessonConvMap
	return TomMigrationScriptStateToCtx(ctx, st), nil
}

func (s *TomMigrationScripts) dbHasUserDeviceTokenWithResourcePathOfThisSchool(ctx context.Context, numuser int) (context.Context, error) {
	stepState := TomMigrationScriptStateFromCtx(ctx)
	schoolID := stepState.schoolID
	err := try.Do(func(attempt int) (bool, error) {
		selectq := `
	select count(*) from user_device_tokens where resource_path=$1
	`
		row := s.TomDB.QueryRow(ctx, selectq, schoolID)
		var hasusers int
		err := row.Scan(&hasusers)
		if err != nil {
			return false, err
		}
		if hasusers != numuser {
			time.Sleep(2 * time.Second)
			return attempt < 5, fmt.Errorf("want %d, has %d user device token in db with resource path %s", numuser, hasusers, schoolID)
		}
		return false, nil
	})
	return ctx, err
}

func (s *TomMigrationScripts) thoseUsersHaveResourcePathInBobDB(ctx context.Context) (context.Context, error) {
	st := TomMigrationScriptStateFromCtx(ctx)
	schoolID := st.schoolID
	_, err := s.BobDB.Exec(ctx, `
	update users set resource_path = $1 where user_id = any($2)
	`, database.Text(schoolID), database.TextArray(st.userIDs))
	return ctx, err
}

func (s *TomMigrationScripts) thoseLessonChatsHaveEmptyResourcePath(ctx context.Context) (context.Context, error) {
	st := TomMigrationScriptStateFromCtx(ctx)
	lessonChats := st.lessonConvMap
	chatids := stringMapValues(lessonChats)
	cl := interceptors.JWTClaimsFromContext(ctx)
	cl.Manabie.ResourcePath = ""
	ctx2 := interceptors.ContextWithJWTClaims(ctx, cl)
	_, err := s.TomDB.Exec(ctx2, "update conversation_lesson set resource_path = null where conversation_id = ANY($1)", database.TextArray(chatids))
	if err != nil {
		return ctx, err
	}
	_, err = s.TomDB.Exec(ctx2, "update conversation_members set resource_path = null where conversation_id = ANY($1)", database.TextArray(chatids))
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (s *TomMigrationScripts) thoseUsersHaveDeviceTokenInTomDBWithoutResourcePath(ctx context.Context) (context.Context, error) {
	cl := interceptors.JWTClaimsFromContext(ctx)
	cl.Manabie.ResourcePath = ""
	ctx2 := interceptors.ContextWithJWTClaims(ctx, cl)
	st := TomMigrationScriptStateFromCtx(ctx)
	userIDs := st.userIDs
	repo := &repositories.UserDeviceTokenRepo{}
	fmt.Printf("%v\n", st.userIDs)
	for _, id := range userIDs {
		user := &core.UserDeviceToken{}
		database.AllNullEntity(user)
		user.UserID = database.Text(id)
		user.Token = database.Text(id)
		err := repo.Upsert(ctx2, s.TomDB, user)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func (s *TomMigrationScripts) allLessonConversationHaveCorrectResourcePath(ctx context.Context) (context.Context, error) {
	st := TomMigrationScriptStateFromCtx(ctx)
	lessonConvMap := st.lessonConvMap
	schoolID := st.schoolID
	convIDs := stringMapValues(lessonConvMap)
	checkLessonConv := `
	select count(*) from %s alias where alias.conversation_id =any($1) and alias.resource_path != $2`
	tables := []string{"conversations", "conversation_lesson", "messages", "conversation_members"}
	for _, t := range tables {
		err := try.Do(func(attempt int) (bool, error) {
			var count int
			err := s.TomDB.QueryRow(ctx, fmt.Sprintf(checkLessonConv, t), database.TextArray(convIDs), database.Text(schoolID)).Scan(&count)
			if err != nil {
				return false, err
			}
			if count != 0 {
				time.Sleep(2 * time.Second)
				return attempt < 5, fmt.Errorf("still %d %s has incorrect resource path", count, t)
			}
			return false, nil
		})
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func (s *TomMigrationScripts) accountsAndChatsOfStudentsAreCreatedWithLocationEachHasParents(ctx context.Context, numstudent int, locLabel string, parentsPerStudent int) (context.Context, error) {
	ctx2, err := EnsureSchoolAdminToken(ctx, s.commonSuite)
	if err != nil {
		return ctx, err
	}
	ctx = ctx2

	locID := s.getLocation(ctx, locLabel)
	stepState := TomMigrationScriptStateFromCtx(ctx)

	school := strToI32(stepState.schoolID)

	for i := 0; i < numstudent; i++ {
		stu, _, err := s.commonSuite.CreateStudentWithParent(ctx, []string{locID}, nil)
		if err != nil {
			return ctx, fmt.Errorf("s.commonsuite.CreateStudentWithParent: %w", err)
		}
		for i := 0; i < parentsPerStudent-1; i++ {
			_, err := s.commonSuite.CreateParentForStudent(ctx, stu.UserProfile.UserId)
			if err != nil {
				return ctx, err
			}
		}
	}
	err = s.waitUntilChatsCreated(ctx, school, numstudent+numstudent*parentsPerStudent)
	if err != nil {
		return ctx, err
	}
	return TomMigrationScriptStateToCtx(ctx, stepState), nil
}

func (s *TomMigrationScripts) accountsAndChatsOfStudentsAreCreatedEachHasParents(ctx context.Context, numstudent int, parentsPerStudent int) (context.Context, error) {
	ctx2, err := EnsureSchoolAdminToken(ctx, s.commonSuite)
	if err != nil {
		return ctx, err
	}
	ctx = ctx2
	stepState := TomMigrationScriptStateFromCtx(ctx)

	school := strToI32(stepState.schoolID)
	userIDs := []string{}

	for i := 0; i < numstudent; i++ {
		stu, par, err := s.commonSuite.CreateStudentWithParent(ctx, []string{stepState.orgLocation}, nil)
		if err != nil {
			return ctx, fmt.Errorf("s.commonsuite.CreateStudentWithParent: %w", err)
		}
		userIDs = append(userIDs, stu.UserProfile.UserId)
		userIDs = append(userIDs, par.UserProfile.UserId)
		for i := 0; i < parentsPerStudent-1; i++ {
			par2, err := s.commonSuite.CreateParentForStudent(ctx, stu.UserProfile.UserId)
			if err != nil {
				return ctx, err
			}
			userIDs = append(userIDs, par2.UserProfile.UserId)
		}
	}
	stepState.createdStudent = numstudent
	stepState.createdParent = numstudent * parentsPerStudent
	stepState.userIDs = userIDs
	err = s.waitUntilChatsCreated(ctx, school, numstudent+numstudent*parentsPerStudent)
	if err != nil {
		return ctx, err
	}
	return TomMigrationScriptStateToCtx(ctx, stepState), nil
}

func (s *TomMigrationScripts) waitUntilChatsCreated(ctx context.Context, schoolID int32, chatNum int) error {
	schoolStr := strconv.Itoa(int(schoolID))
	checkChats := `
		select count(*) from conversation_members cm left join conversations c
		on cm.conversation_id=c.conversation_id where c.owner = $1;
	`

	return try.Do(func(attempt int) (bool, error) {
		row := s.TomDB.QueryRow(ctx, checkChats, database.Text(schoolStr))
		var count int
		err := row.Scan(&count)
		if err != nil {
			return false, err
		}
		if count == chatNum {
			return false, nil
		}
		if count > chatNum {
			return false, fmt.Errorf("total conversations (%d) are created more than expected(%d), db is not clean or bug in logic", count, chatNum)
		}
		time.Sleep(2 * time.Second)
		return attempt < 10, fmt.Errorf("chats are not created yet")
	})
}

func (s *TomMigrationScripts) aNewSchoolIsCreatedWithLocationDefault(ctx context.Context, defaultLocLabel string) (context.Context, error) {
	stepState := TomMigrationScriptStateFromCtx(ctx)
	schoolID, loc, _, err := s.commonSuite.NewOrgWithOrgLocation(ctx)
	if err != nil {
		return ctx, nil
	}
	stepState.schoolID = i32ToStr(schoolID)
	stepState.orgLocation = loc
	stepState.locationPool[defaultLocLabel] = loc

	ctx = contextWithResourcePath(ctx, stepState.schoolID)

	return TomMigrationScriptStateToCtx(ctx, stepState), nil
}

// DEPRECATE don't use it anymore
func (s *TomMigrationScripts) aNewSchoolIsCreated(ctx context.Context) (context.Context, error) {
	stepState := TomMigrationScriptStateFromCtx(ctx)
	schoolID, org, _, err := s.commonSuite.NewOrgWithOrgLocation(ctx)
	if err != nil {
		return ctx, err
	}
	stepState.schoolID = i32ToStr(schoolID)
	stepState.orgLocation = org

	schoolText := strconv.Itoa(int(schoolID))

	ctx = contextWithResourcePath(ctx, schoolText)

	return TomMigrationScriptStateToCtx(ctx, stepState), nil
}

func (s *TomMigrationScripts) removeAllLocationOfConversationsInCurrentSchool(ctx context.Context) (context.Context, error) {
	stepState := TomMigrationScriptStateFromCtx(ctx)

	schoolID := stepState.schoolID
	// TODO: use some other connection to bypassrls if rls is enabled for local
	hardDelete := `
delete from conversation_locations where conversation_id=any(select conversation_id from conversations where resource_path=$1 limit 1)
	`
	softDelete := `
	update conversation_locations set deleted_at=now() where resource_path=$1
	`
	tag, err := s.TomPostgresDB.Exec(ctx, hardDelete, schoolID)
	if err != nil {
		return ctx, err
	}
	if tag.RowsAffected() != 1 {
		return ctx, fmt.Errorf("expect hard delete return 1 item")
	}
	_, err = s.TomPostgresDB.Exec(ctx, softDelete, schoolID)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (s *TomMigrationScripts) thereAreSupportChatsWithExactLocations(ctx context.Context, expectConv int, locs string) (context.Context, error) {
	checkLocsQuery := `
select st.student_id, (
	select array_agg(location_id) as locs from conversation_locations cl
	where cl.deleted_at is null and cl.conversation_id = c.conversation_id
) as locs from conversation_students st left join conversations c using(conversation_id)
where st.resource_path=$1` // in case rls is not yet enabled
	locIDs := s.getLocations(ctx, locs)
	stepState := TomMigrationScriptStateFromCtx(ctx)

	schoolID := stepState.schoolID
	err := doRetry(func() (bool, error) {
		rows, err := s.TomDB.Query(ctx, checkLocsQuery, schoolID)
		if err != nil {
			return false, err
		}
		defer rows.Close()
		debugErr := ""
		var matchCount int
		for rows.Next() {
			var (
				dbLocIDs pgtype.TextArray
				stuID    string
			)
			err := rows.Scan(&stuID, &dbLocIDs)
			if err != nil {
				return false, err
			}
			if !stringutil.SliceElementsMatch(locIDs, database.FromTextArray(dbLocIDs)) {
				debugErr += fmt.Sprintf("student %s's conversation has locations %v in db, expect %v\n", stuID, database.FromTextArray(dbLocIDs), locIDs)
				continue
			}
			matchCount++
		}
		if matchCount != expectConv {
			return true, fmt.Errorf(debugErr)
		}
		return false, nil
	})
	return ctx, err
}

func (s *TomMigrationScripts) totalConversationsInThisSchoolIs(ctx context.Context, expectConv int) (context.Context, error) {
	stepState := TomMigrationScriptStateFromCtx(ctx)

	schoolID := stepState.schoolID
	err := doRetry(func() (bool, error) {
		q := "select count(*) from conversations where resource_path = $1 and status=$2"
		var totalConv pgtype.Int8
		err := s.TomDB.QueryRow(ctx, q, schoolID, tpb.ConversationStatus_CONVERSATION_STATUS_NONE.String()).Scan(&totalConv)
		if err != nil {
			return false, err
		}
		if int(totalConv.Int) != expectConv {
			return true, fmt.Errorf("want %d, has %d conversation in db with resource path %s", expectConv, totalConv, schoolID)
		}
		return false, nil
	})
	return ctx, err
}

func (s *TomMigrationScripts) forceRemoveAllSupportConversationsOfThisSchool(ctx context.Context) (context.Context, error) {
	stepState := TomMigrationScriptStateFromCtx(ctx)

	schoolID := stepState.schoolID
	removeAllScripts := []string{
		`DELETE FROM conversation_locations cl WHERE cl.conversation_id in (
			select conversation_id from conversations where conversations.owner=$1
		);`,
		`DELETE FROM messages WHERE messages.conversation_id in (
			select conversation_id from conversations where conversations.owner=$1
		);`,
		`DELETE FROM conversation_students WHERE conversation_students.conversation_id in (
			select conversation_id from conversations where conversations.owner=$1
		);`,
		`DELETE FROM conversation_members WHERE conversation_members.conversation_id in (
			select conversation_id from conversations where conversations.owner=$1
		);`,
		`DELETE FROM conversations where conversations.owner=$1`,
	}
	for _, item := range removeAllScripts {
		_, err := s.TomPostgresDB.Exec(ctx, item, database.Text(schoolID))
		if err != nil {
			return TomMigrationScriptStateToCtx(ctx, stepState), err
		}
	}

	return TomMigrationScriptStateToCtx(ctx, stepState), nil
}

func (s *TomMigrationScripts) runScript(ctx context.Context, scriptName string) (context.Context, error) {
	stepState := TomMigrationScriptStateFromCtx(ctx)

	schoolID := stepState.schoolID
	ctx = auth.InjectFakeJwtToken(ctx, schoolID)
	bobDB := s.Connections.BobDBTrace
	jsm := s.Connections.JSM

	var err error

	switch scriptName {
	case "YasuoSyncUserConversation Create":
		s.syncedStudent, s.syncedParent = yasuocmd.SyncConversation(ctx, yasuoSyncConfig, bobDB, jsm, schoolID, schoolID, yasuocmd.SyncCreate)
	case "YasuoSyncUserConversation Update":
		s.syncedStudent, s.syncedParent = yasuocmd.SyncConversation(ctx, yasuoSyncConfig, bobDB, jsm, schoolID, schoolID, yasuocmd.SyncUpdate)
	case "BobSyncLessonConversation":
		ctx = auth.InjectFakeJwtToken(ctx, schoolID)
		stepState.createdLessonNum, stepState.updatedLessonNum, err = bobcmd.SyncLessonConversation(ctx, bobSyncConfig, zap.NewNop().Sugar(), bobDB, jsm, schoolID, schoolID, s.TomConn)
	case "SynConversationDocument":
		ctx = auth.InjectFakeJwtToken(ctx, schoolID)
		tomcmd.SyncConversationDocument(ctx, tomSyncConfig, s.TomDBTrace, s.searchClient, schoolID, schoolID, s.EurekaConn)
	}
	time.Sleep(2 * time.Second)

	return TomMigrationScriptStateToCtx(ctx, stepState), err
}

func (s *TomMigrationScripts) forceRemoveStudentChatsAndAssociatedParentChats(ctx context.Context, numStudentChats int) (context.Context, error) {
	stepState := TomMigrationScriptStateFromCtx(ctx)

	tx, err := s.TomPostgresDB.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		return TomMigrationScriptStateToCtx(ctx, stepState), err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(context.Background())
		}
	}()

	school := stepState.schoolID
	query := `
		SELECT c.conversation_id
		FROM conversations c
		LEFT JOIN conversation_students cs ON cs.conversation_id = c.conversation_id
		WHERE cs.student_id in
			(SELECT distinct(cs.student_id)
			FROM conversation_students cs
			LEFT JOIN conversations c ON cs.conversation_id = c.conversation_id
			WHERE c.owner=$1
			ORDER BY cs.student_id
			OFFSET 0
			LIMIT $2)
	`
	rows, err := tx.Query(ctx, query, database.Text(school), numStudentChats)
	if err != nil {
		return TomMigrationScriptStateToCtx(ctx, stepState), err
	}
	var removedConvs []string
	defer rows.Close()
	for rows.Next() {
		convID := ""
		err = rows.Scan(&convID)
		if err != nil {
			return TomMigrationScriptStateToCtx(ctx, stepState), err
		}
		removedConvs = append(removedConvs, convID)
	}
	removePartialScripts := []string{
		`DELETE FROM messages WHERE messages.conversation_id = ANY($1)`,
		`DELETE FROM conversation_students WHERE conversation_students.conversation_id = ANY($1)`,
		`DELETE FROM conversation_members WHERE conversation_members.conversation_id = ANY($1)`,
		`DELETE FROM conversations WHERE conversation_id=ANY($1)`,
	}
	for _, item := range removePartialScripts {
		_, err = tx.Exec(ctx, item, database.TextArray(removedConvs))
		if err != nil {
			return TomMigrationScriptStateToCtx(ctx, stepState), err
		}
	}
	cmerr := tx.Commit(context.Background())
	if cmerr != nil {
		return TomMigrationScriptStateToCtx(ctx, stepState), cmerr
	}

	return TomMigrationScriptStateToCtx(ctx, stepState), nil
}

func (s *TomMigrationScripts) numberOfStudentChatsCreatedEqual(expectedSyncStudents int) error {
	if s.syncedStudent != expectedSyncStudents {
		return fmt.Errorf("want %d students synced, got %d", expectedSyncStudents, s.syncedStudent)
	}
	return nil
}

func (s *TomMigrationScripts) numberOfParentMembershipCreatedEqual(expectedSyncParents int) error {
	if s.syncedParent != expectedSyncParents {
		return fmt.Errorf("want %d parents synced, got %d", expectedSyncParents, s.syncedParent)
	}
	return nil
}
