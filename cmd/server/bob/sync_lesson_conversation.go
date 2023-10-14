package bob

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	schoolID   string
	schoolName string
)

func init() {
	bootstrap.RegisterJob("bob_sync_lesson_conversations", RunSyncLessonConversation).
		StringVar(&schoolID, "schoolID", "", "sync for specific school").
		StringVar(&schoolName, "schoolName", "", "should match with school name in secret config, for sanity check")
}

func checkSchoolIDMatchSchoolName(ctx context.Context, db database.QueryExecer, schoolID, schoolName string) error {
	var count pgtype.Int8
	if err := db.QueryRow(ctx, "select count(*) from organizations where organization_id=$1 and name=$2", schoolID, schoolName).Scan(&count); err != nil {
		return err
	}
	if count.Int != 1 {
		return fmt.Errorf("school name %s with schoolID %s has %d count in db", schoolName, schoolID, count.Int)
	}
	return nil
}

// ctx should already be injected with fake jwt claim (for query db)
// and token to call grpc
func SyncLessonConversation(
	ctx context.Context,
	c configurations.Config,
	l *zap.SugaredLogger,
	bobDB *database.DBTrace,
	jsm nats.JetStreamManagement,
	schoolID string,
	schoolName string,
	tomConn *grpc.ClientConn,
) (int, int, error) {
	l.Infof("Sync for schoolID: %s", schoolID)

	if c.Common.Environment != "prod" && c.Common.Environment != "uat" && schoolID == "" {
		return 0, 0, fmt.Errorf("running in non (production/uat) requires a school id")
	}

	if err := checkSchoolIDMatchSchoolName(ctx, bobDB, schoolID, schoolName); err != nil {
		return 0, 0, fmt.Errorf("checkSchoolIDMatchSchoolName: %s", err)
	}

	perBatch := 100
	offset := 0
	if schoolID == "" {
		return 0, 0, fmt.Errorf("empty school id")
	}
	intSchool, err := strconv.ParseInt(schoolID, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("non numeric schoolID: %s", schoolID)
	}
	school := database.Int4(int32(intSchool))

	convReaderService := tpb.NewConversationReaderServiceClient(tomConn)

	var totalCreated, totalUpdated int
	for {
		rows, err := bobDB.Query(ctx, scanActiveLessonQuery, school, perBatch, offset)
		if err != nil {
			return 0, 0, fmt.Errorf("bobDB.Query: %s", err)
		}
		defer rows.Close()
		if err := rows.Err(); err != nil {
			return 0, 0, fmt.Errorf("rows.Err(): %s", err)
		}
		offset += perBatch

		sourceLessonStudents := map[string][]string{}
		lessonNames := map[string]string{}

		count := 0

		for rows.Next() {
			count++
			var (
				lessonID, lessonName pgtype.Text
				studentIDs           pgtype.TextArray
				schoolID             pgtype.Int4
			)

			err := rows.Scan(&lessonID, &lessonName, &studentIDs, &schoolID)
			if err != nil {
				return 0, 0, fmt.Errorf("failed to scan a row: %s", err)
			}
			sourceLessonStudents[lessonID.String] = database.FromTextArray(studentIDs)
			lessonNames[lessonID.String] = lessonName.String
		}
		if err := rows.Err(); err != nil {
			return 0, 0, fmt.Errorf("rows.Err(): %s", err)
		}
		if count == 0 {
			l.Infof("Query return 0 rows, done syncing")
			break
		}

		// check which lesson has no student to send update event
		sourceLessons := make([]string, 0, len(sourceLessonStudents))
		for lesson := range sourceLessonStudents {
			sourceLessons = append(sourceLessons, lesson)
		}

		// else call Tom to check if learners are added correctly
		syncedLessons, err := checkTomFindSyncedLessons(ctx, convReaderService, sourceLessons, schoolID)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to check synced student lessons: %s", err)
		}

		checkList := map[string]struct{}{}
		for lessonID := range syncedLessons {
			checkList[lessonID] = struct{}{}
		}

		for _, sourceLesson := range sourceLessons {
			if _, created := checkList[sourceLesson]; created {
				sourceStudents := sourceLessonStudents[sourceLesson]
				// students are not synced
				if !compareStudentsSynced(sourceStudents, syncedLessons[sourceLesson]) {
					updateLessonEvt := &bpb.EvtLesson{
						Message: &bpb.EvtLesson_UpdateLesson_{
							UpdateLesson: &bpb.EvtLesson_UpdateLesson{
								LessonId:   sourceLesson,
								LearnerIds: sourceStudents,
								ClassName:  lessonNames[sourceLesson],
							},
						},
					}
					publish(ctx, constants.SubjectLessonChatSynced, l, jsm, updateLessonEvt)
					totalUpdated++
				}
			} else {
				sourceStudent := sourceLessonStudents[sourceLesson]
				createLessonEvt := &bpb.EvtLesson{
					Message: &bpb.EvtLesson_CreateLessons_{
						CreateLessons: &bpb.EvtLesson_CreateLessons{
							Lessons: []*bpb.EvtLesson_Lesson{
								{
									LessonId:   sourceLesson,
									LearnerIds: sourceStudent,
									Name:       lessonNames[sourceLesson],
								},
							},
						},
					},
				}
				publish(ctx, constants.SubjectLessonChatSynced, l, jsm, createLessonEvt)

				totalCreated++
			}
		}
	}
	return totalCreated, totalUpdated, nil
}

func compareStudentsSynced(sourcedStudents []string, conv *tpb.Conversation) bool {
	syncedStudents := []string{}
	for _, u := range conv.GetUsers() {
		if u.Group.String() == cpb.UserGroup_USER_GROUP_STUDENT.String() && u.GetIsPresent() {
			syncedStudents = append(syncedStudents, u.Id)
		}
	}
	return stringutil.SliceElementsMatch(sourcedStudents, syncedStudents)
}

func RunSyncLessonConversation(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	// for db RLS query
	ctx = auth.InjectFakeJwtToken(ctx, schoolID)
	l := rsc.Logger().Sugar()
	created, updated, err := SyncLessonConversation(ctx, c, l, rsc.DB(), rsc.NATS(), schoolID, schoolName, rsc.GRPCDial("tom"))
	if err != nil {
		return err
	}
	l.Infof("Synced result: %d new lesson(s) created, %d lesson(s) updated", created, updated)
	return nil
}

func checkTomFindSyncedLessons(ctx context.Context, svc tpb.ConversationReaderServiceClient, lessonIDs []string, schoolID string) (map[string]*tpb.Conversation, error) {
	response, err := svc.ListConversationByLessons(ctx, &tpb.ListConversationByLessonsRequest{
		LessonIds:      lessonIDs,
		OrganizationId: schoolID,
	})
	if err != nil {
		return nil, fmt.Errorf("svc.ListConversationByLessons: %s", err)
	}
	if response.GetConversations() == nil {
		return map[string]*tpb.Conversation{}, nil
	}
	return response.GetConversations(), nil
}

// each returned row is a lesson info with a list of students
var scanActiveLessonQuery = `
select l.lesson_id,l.name,
(select array_agg(st.student_id) as student_members
	from students st
		left join lesson_members lm on st.student_id=lm.user_id
	where lm.lesson_id=l.lesson_id and lm.deleted_at IS NULL and st.deleted_at IS NULL
) as student_members,c.school_id
from lessons l left join courses c using(course_id) 
where l.deleted_at IS NULL
and lesson_type = 'LESSON_TYPE_ONLINE' 
and c.deleted_at IS NULL
and ($1::integer IS NULL or c.school_id=$1) 
order by l.lesson_id limit $2 offset $3;
`

func publish(ctx context.Context, topic string, l *zap.SugaredLogger, jsm nats.JetStreamManagement, evt protoreflect.ProtoMessage) {
	msg, _ := proto.Marshal(evt)
	err := try.Do(func(attempt int) (bool, error) {
		_, err := jsm.PublishContext(ctx, topic, msg)
		if err == nil {
			return false, nil
		}
		retry := attempt < 5
		if retry {
			time.Sleep(1 * time.Second)
			return true, fmt.Errorf("temporary error jsm.PublishContext: %s", err.Error())
		}
		return false, err
	})
	if err != nil {
		l.Error("jsm.PublishContext failed", zap.Error(err))
	}
}
