package yasuo

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/yasuo/configurations"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	schoolID   string
	schoolName string
	syncType   string
)

func init() {
	bootstrap.RegisterJob("yasuo_sync_user_conversations", RunSyncConversation).
		StringVar(&schoolID, "schoolID", "", "").
		StringVar(&schoolName, "schoolName", "", "")
}

var zlogger *zap.SugaredLogger

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
func SyncConversation(ctx context.Context, _ configurations.Config,
	bobdb *database.DBTrace,
	jsm nats.JetStreamManagement,
	schoolID string, schoolName string, syncType string,
) (int, int) {
	if zlogger == nil {
		zlogger = zap.NewNop().Sugar()
	}
	err := checkSchoolIDMatchSchoolName(ctx, bobdb, schoolID, schoolName)
	if err != nil {
		zlogger.DPanic(err)
	}

	perBatch := 100
	offset := 0
	school := pgtype.Int4{Status: pgtype.Null}
	if schoolID != "" {
		intSchool, err := strconv.ParseInt(schoolID, 10, 64)
		if err != nil {
			zlogger.Fatalf("non numeric schoolID: %s", schoolID)
		}
		school = database.Int4(int32(intSchool))
	}

	var totalParent, totalStudent int
	for {
		ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
		rows, err := bobdb.Query(ctx2, studentScanQuery, school, perBatch, offset)
		cancel()
		if err != nil {
			zlogger.Fatalf(err.Error())
		}
		offset += perBatch

		count := 0
		defer rows.Close()
		for rows.Next() {
			count++
			var (
				studentID, givenName, lastName pgtype.Text
				pgparentIDs, locationIDs       pgtype.TextArray
				schoolID                       pgtype.Int4
			)
			err := rows.Scan(&studentID, &schoolID, &givenName, &lastName, &pgparentIDs, &locationIDs)
			if err != nil {
				zlogger.Errorf("failed to scan a row: %s", err)
			}
			if studentID.Status != pgtype.Present || studentID.String == "" {
				zlogger.Errorf("student id null or does not exist")
			}
			locIDs := database.FromTextArray(locationIDs)
			schoolIDText := strconv.Itoa(int(schoolID.Int))
			studentName := composeName(givenName, lastName)
			switch syncType {
			case SyncCreate:
				createStudentEvt := &upb.EvtUser_CreateStudent{
					StudentId:   studentID.String,
					StudentName: studentName,
					SchoolId:    schoolIDText,
					LocationIds: locIDs,
				}

				publishJsm(ctx, jsm, constants.SubjectUserCreated, &upb.EvtUser{
					Message: &upb.EvtUser_CreateStudent_{
						CreateStudent: createStudentEvt,
					},
				})
				totalStudent++

				parentIDs := database.FromTextArray(pgparentIDs)
				for _, parentID := range parentIDs {
					createParentEvt := &upb.EvtUser_ParentAssignedToStudent{
						StudentId: studentID.String,
						ParentId:  parentID,
					}
					publishJsm(ctx, jsm, constants.SubjectUserCreated, &upb.EvtUser{
						Message: &upb.EvtUser_ParentAssignedToStudent_{
							ParentAssignedToStudent: createParentEvt,
						},
					})
				}
				totalParent += len(parentIDs)
			case SyncUpdate:
				// Tom subscriber will not upsert fields that is empty
				updateProfileEvt := &upb.EvtUser{
					Message: &upb.EvtUser_UpdateStudent_{
						UpdateStudent: &upb.EvtUser_UpdateStudent{
							StudentId:   studentID.String,
							LocationIds: locIDs,
						},
					},
				}
				publishJsm(ctx, jsm, constants.SubjectUserUpdated, updateProfileEvt)
				totalStudent++
			}
		}

		if count == 0 {
			zlogger.Infof("Query return 0 rows, done syncing")
			break
		}
	}
	// wait for async jobs to complete
	return totalStudent, totalParent
}

const (
	SyncUpdate = "update"
	SyncCreate = "create"
)

// TODO: validate schoolID and schoolName
func RunSyncConversation(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	zlogger = rsc.Logger().Sugar()

	// for db RLS query
	ctx = auth.InjectFakeJwtToken(ctx, schoolID)

	totalStudent, totalParent := SyncConversation(ctx, c, rsc.DB(), rsc.NATS(), schoolID, schoolName, syncType)
	zlogger.Infof("Synced total: %d new student(s), %d new parent(s)", totalStudent, totalParent)
	return nil
}

var studentScanQuery = `
select student_id, school_id, u.given_name, u.name,
(
	select array_agg(sp.parent_id) as parent_ids from student_parents sp left join students st2
	on st2.student_id = sp.student_id and sp.deleted_at IS NULL
	where st2.student_id=st.student_id
),
(
	select array_agg(uap.location_id) as locations from user_access_paths uap left join students st2
	on uap.user_id=st2.student_id where uap.user_id=st.student_id and uap.deleted_at IS NULL
)
from students st
left join users u on st.student_id=u.user_id
where st.deleted_at IS NULL and u.deleted_at IS NULL and ($1::integer IS NULL OR st.school_id=$1)
order by st.created_at desc limit $2 offset $3
`

func publishJsm(ctx context.Context, jsm nats.JetStreamManagement, topic string, evt protoreflect.ProtoMessage) {
	msg, _ := proto.Marshal(evt)
	err := try.Do(func(attempt int) (bool, error) {
		_, err := jsm.PublishContext(ctx, topic, msg)
		if err == nil {
			return false, nil
		}
		retry := attempt < 5
		if retry {
			time.Sleep(1 * time.Second)
		}
		return false, err
	})
	if err != nil {
		zlogger.Error("jsm.PublishContext failed", zap.Error(err))
	}
}

func composeName(given, last pgtype.Text) string {
	if given.Status == pgtype.Null || given.String == "" {
		return last.String
	}
	return given.String + " " + last.String
}

type StudentParentKey struct {
	StudentID string
	ParentID  string
}
type StudentJob struct {
	StudentID   pgtype.Text
	StudentName string
	SchoolID    pgtype.Int4
	LocationIDs []string
}
type ParentJob struct {
	StudentID   pgtype.Text
	ParentID    pgtype.Text
	SchoolID    pgtype.Int4
	StudentName string
}
