package yasuo

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	repositories_bob "github.com/manabie-com/backend/internal/bob/repositories"
	enigma_entites "github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/yasuo/repositories"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/gogo/protobuf/types"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) toEventMasterRegistrationStudentLesson(ctx context.Context, status, actionKind string, total int) (context.Context, []*npb.EventSyncUserCourse_StudentLesson, error) {
	stepState := StepStateFromContext(ctx)

	if total == 0 {
		return ctx, []*npb.EventSyncUserCourse_StudentLesson{}, nil
	}

	studentLessons := []*npb.EventSyncUserCourse_StudentLesson{}

	switch status {
	case "new":
		stepState.Request = nil

		ctx, err := s.jprefSyncLessonWithActionAndLessonWithAction(StepStateToContext(ctx, stepState), "2", npb.ActionKind_ACTION_KIND_UPSERTED.String(), "0", "")
		if err != nil {
			return ctx, nil, fmt.Errorf("jprefSyncLessonWithActionAndLessonWithAction: %w", err)
		}

		ctx, err = s.theseLessonsMustBeStoreInOurSystem(StepStateToContext(ctx, stepState))
		if err != nil {
			return ctx, nil, fmt.Errorf("theseLessonsMustBeStoreInOurSystem: %w", err)
		}
		lessonIDs := []string{}
		for _, s := range stepState.Request.([]*npb.EventMasterRegistration_Lesson) {
			lessonIDs = append(lessonIDs, s.LessonId)
		}

		for i := 0; i < total; i++ {
			ctx, err := s.aSignedIn(ctx, "student")
			if err != nil {
				return ctx, nil, err
			}

			studentLessons = append(studentLessons, &npb.EventSyncUserCourse_StudentLesson{
				ActionKind: npb.ActionKind(npb.ActionKind_value[actionKind]),
				StudentId:  stepState.CurrentUserID,
				LessonIds:  lessonIDs,
			})
		}
	case "existed":
		stepState.Request = nil

		ctx, err1 := s.jprepSyncLessonMembersWithActionAndLessonMembersWithAction(StepStateToContext(ctx, stepState), strconv.Itoa(total), npb.ActionKind_ACTION_KIND_UPSERTED.String(), "0", "", 0)
		ctx, err2 := s.theseLessonMembersMustBeStoreInOurSystem(StepStateToContext(ctx, stepState))

		err := multierr.Combine(err1, err2)

		if err != nil {
			return ctx, nil, fmt.Errorf("err prepare existed lesson member: %w", err)
		}

		for _, s := range stepState.Request.([]*npb.EventSyncUserCourse_StudentLesson) {
			s.ActionKind = npb.ActionKind(npb.ActionKind_value[actionKind])
			studentLessons = append(studentLessons, s)
		}
	}

	return StepStateToContext(ctx, stepState), studentLessons, nil
}

func (s *suite) jprepSyncLessonMembersWithActionAndLessonMembersWithAction(ctx context.Context, numberOfNewLessonMember, newLessonMemberAction, numberOfExistedLessonMember, existedLessonMemberAction string, hours int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	total, err := strconv.Atoi(numberOfNewLessonMember)
	if err != nil {
		return ctx, err
	}
	stepState.RequestSentAt = time.Now()
	stepState.CurrentSchoolID = constants.JPREPSchool

	ctx, err = s.aSignedIn(ctx, "school admin")
	if err != nil {
		return ctx, err
	}

	ctx, newLessonMembers, err := s.toEventMasterRegistrationStudentLesson(StepStateToContext(ctx, stepState), "new", newLessonMemberAction, total)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	total, err = strconv.Atoi(numberOfExistedLessonMember)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, existedLessonMembers, err := s.toEventMasterRegistrationStudentLesson(StepStateToContext(ctx, stepState), "existed", existedLessonMemberAction, total)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	lessons := append(newLessonMembers, existedLessonMembers...)
	stepState.Request = lessons
	stepState.FoundChanForJetStream = make(chan interface{}, 1)

	handler := func(ctx context.Context, data []byte) (bool, error) {
		r := &npb.EventSyncUserCourse{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return false, err
		}
		// req := stepState.Request.([]*npb.EventSyncUserCourse_StudentLesson)

		type reqtype struct {
			student string
			lesson  string
		}
		reqmap := map[reqtype]npb.ActionKind{}
		for idx, item := range lessons {
			for _, lesson := range lessons[idx].LessonIds {
				reqmap[reqtype{
					student: lessons[idx].StudentId,
					lesson:  lesson,
				}] = item.ActionKind
			}
		}
		for _, item := range r.GetStudentLessons() {
			for _, lesson := range item.LessonIds {
				key := reqtype{
					student: item.StudentId,
					lesson:  lesson,
				}
				action, exist := reqmap[key]
				if !exist {
					return false, errors.New("reqmap[key] does not exist")
				}
				// there is a case where req.action kind = upsert, but resulted event is removed
				// but is in other test case, not here
				if action != item.ActionKind {
					return false, errors.New("action != item.ActionKind")
				}
			}
		}
		stepState.FoundChanForJetStream <- struct{}{}
		return false, nil
	}

	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(stepState.RequestSentAt),
			nats.AckWait(2 * time.Second),
		}}

	sub, err := s.JSM.Subscribe(constants.SubjectSyncStudentLessons, opts, handler)
	if err != nil {
		return ctx, fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	signature := idutil.ULIDNow()
	ctx, err = s.createPartnerSyncDataLog(ctx, signature, time.Duration(hours))
	if err != nil {
		return ctx, fmt.Errorf("create partner sync data log error: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	req := &npb.EventSyncUserCourse{RawPayload: []byte("{}"), Signature: signature, StudentLessons: lessons}
	data, _ := proto.Marshal(req)
	_, err = s.JSM.PublishContext(ctx, constants.SubjectJPREPSyncUserCourseNatsJS, data)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("publish: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) toLiveLessonMsg(ctx context.Context, status, actionKind string, total int) (context.Context, []*npb.EventMasterRegistration_Lesson, error) {
	stepState := StepStateFromContext(ctx)
	if total == 0 {
		return ctx, []*npb.EventMasterRegistration_Lesson{}, nil
	}

	lessons := []*npb.EventMasterRegistration_Lesson{}

	switch status {
	case "new":
		stepState.CurrentSchoolID = constants.JPREPSchool
		now := timestamppb.Now()
		for i := 0; i < total; i++ {
			ctx, err1 := s.aTeacherAccountWithSchoolID(ctx, stepState.CurrentSchoolID)
			ctx, err2 := s.aClass(ctx)
			ctx, err3 := s.aSignedIn(ctx, "school admin")
			ctx, err4 := s.aUpsertLiveCourseRequestWithMissing(ctx, "id")

			ctx, err5 := s.userUpsertLiveCourses(ctx)
			ctx, err6 := s.returnsStatusCode(ctx, "OK")
			ctx, err7 := s.yasuoMustStoreLiveCourse(ctx)

			err := multierr.Combine(err1, err2, err3, err4, err5, err6, err7)
			if err != nil {
				return ctx, nil, fmt.Errorf("err prepare live course: %w", err)
			}

			lessons = append(lessons, &npb.EventMasterRegistration_Lesson{
				ActionKind: npb.ActionKind(npb.ActionKind_value[actionKind]),
				LessonId:   idutil.ULIDNow(),
				CourseId:   stepState.CurrentCourseID,
				StartDate:  now,
				EndDate: &timestamppb.Timestamp{
					Seconds: now.Seconds + 3600,
				},
				LessonGroup: "ABC",
				ClassName:   "className" + idutil.ULIDNow(),
				LessonType:  cpb.LessonType_LESSON_TYPE_ONLINE,
			})
		}
	case "existed":
		stepState.Request = nil

		ctx, err1 := s.jprefSyncLessonWithActionAndLessonWithAction(ctx, strconv.Itoa(total), npb.ActionKind_ACTION_KIND_UPSERTED.String(), "0", "")
		ctx, err2 := s.theseLessonsMustBeStoreInOurSystem(ctx)
		err := multierr.Combine(err1, err2)
		if err != nil {
			return ctx, nil, fmt.Errorf("err prepare existed live lesson: %w", err)
		}

		for _, s := range stepState.Request.([]*npb.EventMasterRegistration_Lesson) {
			s.ActionKind = npb.ActionKind(npb.ActionKind_value[actionKind])
			lessons = append(lessons, s)
		}
	}

	return StepStateToContext(ctx, stepState), lessons, nil
}

func (s *suite) jprefSyncLessonWithActionAndLessonWithAction(ctx context.Context, numberOfNewLesson, newLessonAction, numberOfExistedLesson, existedLessonAction string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	total, err := strconv.Atoi(numberOfNewLesson)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.RequestSentAt = time.Now()
	stepState.CurrentSchoolID = constants.JPREPSchool
	ctx, err = s.aSignedIn(ctx, "school admin")
	if err != nil {
		return ctx, err
	}

	ctx, newLessons, err := s.toLiveLessonMsg(StepStateToContext(ctx, stepState), "new", newLessonAction, total)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	total, err = strconv.Atoi(numberOfExistedLesson)
	if err != nil {
		return ctx, err
	}
	if len(newLessons) > 0 {
		stepState.ExistedLessons = nil
		stepState.ExistedLessons = append(stepState.ExistedLessons, newLessons...)
	}

	ctx, existedLessons, err := s.toLiveLessonMsg(StepStateToContext(ctx, stepState), "existed", existedLessonAction, total)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	lessons := append(newLessons, existedLessons...)

	stepState.Request = lessons

	// create consumer jetstream
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(stepState.RequestSentAt),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}
	handlerLessonCreated := func(ctx context.Context, data []byte) (bool, error) {
		r := &bpb.EvtLesson{}
		err := r.Unmarshal(data)
		if err != nil {
			return true, err
		}
		reqs := stepState.Request.([]*npb.EventMasterRegistration_Lesson)

		reqMap := map[string]*npb.EventMasterRegistration_Lesson{}
		for idx, req := range reqs {
			reqMap[req.LessonId] = reqs[idx]
		}

		for _, msg := range r.GetCreateLessons().GetLessons() {
			req, exist := reqMap[msg.GetLessonId()]
			if !exist {
				return false, fmt.Errorf("req not exist")
			}
			if req.ActionKind != npb.ActionKind_ACTION_KIND_UPSERTED {
				return false, fmt.Errorf("req.ActionKind != npb.ActionKind_ACTION_KIND_UPSERTED")
			}
			if len(msg.LearnerIds) != 0 {
				return false, fmt.Errorf("len(msg.LearnerIds) != 0")
			}
		}

		stepState.FoundChanForJetStream <- r.Message
		return false, nil
	}
	sub, err := s.JSM.Subscribe(constants.SubjectLessonCreated, opts, handlerLessonCreated)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)

	signature := idutil.ULIDNow()
	ctx, err = s.createPartnerSyncDataLog(ctx, signature, 0)
	if err != nil {
		return ctx, fmt.Errorf("create partner sync data log error: %w", err)
	}
	ctx, err = s.createLogSyncDataSplit(ctx, string(enigma_entites.KindLesson))
	if err != nil {
		return ctx, fmt.Errorf("create partner sync data log split error: %w", err)
	}

	req := &npb.EventMasterRegistration{
		RawPayload: []byte("{}"),
		Signature:  signature,
		Lessons:    lessons,
		LogId:      stepState.PartnerSyncDataLogSplitId,
	}
	data, _ := proto.Marshal(req)
	_, err = s.JSM.PublishContext(ctx, constants.SubjectSyncMasterRegistration, data)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Publish: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theseLessonMembersMustBeStoreInOurSystem(ctx context.Context) (context.Context, error) {
	time.Sleep(time.Second)

	stepState := StepStateFromContext(ctx)

	lessonRepo := &repositories_bob.LessonRepo{}
	lessonMemberRepo := &repositories_bob.LessonMemberRepo{}

	for _, l := range stepState.Request.([]*npb.EventSyncUserCourse_StudentLesson) {
		lessonMemebers, err := lessonMemberRepo.Find(ctx, s.DBTrace, database.Text(l.StudentId))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		switch l.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			mapLesson := make(map[string]bool)
			for _, m := range lessonMemebers {
				mapLesson[m.LessonID.String] = true
			}

			for _, id := range l.LessonIds {
				if _, ok := mapLesson[id]; !ok {
					return StepStateToContext(ctx, stepState), fmt.Errorf("not found lessonId: %s", id)
				}
			}
			// check student can retrieve live course
			stepState.AuthToken, err = s.generateExchangeToken(l.StudentId, pb.USER_GROUP_STUDENT.String())
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			lessons, err := lessonRepo.Find(StepStateToContext(ctx, stepState), s.DBTrace, &repositories_bob.LessonFilter{
				LessonID:  database.TextArray(l.LessonIds),
				TeacherID: database.TextArray(nil),
				CourseID:  database.TextArray(nil),
			})
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("er FindLesson: %w", err)
			}

			courseIDs := []string{}
			for _, v := range lessons {
				courseIDs = append(courseIDs, v.CourseID.String)
			}

			resp, err := bpb.NewCourseClient(s.BobConn).RetrieveLiveLesson(contextWithToken(s, ctx), &bpb.RetrieveLiveLessonRequest{
				CourseIds: courseIDs,
				Pagination: &bpb.Pagination{
					Limit: 100,
					Page:  1,
				},
				From: &types.Timestamp{Seconds: time.Now().Unix() + 60},
				To:   &types.Timestamp{Seconds: time.Now().Unix() + 120},
			})
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			if len(resp.Lessons) == 0 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("not found any live lessons")
			}

			for _, lesson := range resp.Lessons {
				if _, ok := mapLesson[lesson.LessonId]; !ok {
					return StepStateToContext(ctx, stepState), fmt.Errorf("check retrieveLiveLesson not found lessonId: %s", lesson.LessonId)
				}
			}
		case npb.ActionKind_ACTION_KIND_DELETED:
			if len(lessonMemebers) != 0 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson members not delete")
			}
		}
	}
	return s.bobMustPushMsgSubjectToNats(ctx, "EventSyncUserCourse", constants.SubjectSyncStudentLessons)
}

func (s *suite) theseNoLessonMembersStoreInOurSystem(ctx context.Context) (context.Context, error) {
	time.Sleep(time.Second)

	stepState := StepStateFromContext(ctx)

	lessonMemberRepo := &repositories_bob.LessonMemberRepo{}

	for _, l := range stepState.Request.([]*npb.EventSyncUserCourse_StudentLesson) {
		lessonMemebers, err := lessonMemberRepo.Find(ctx, s.DBTrace, database.Text(l.StudentId))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if len(lessonMemebers) != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson members exist")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theseLessonsMustBeStoreInOurSystem(ctx context.Context) (context.Context, error) {
	time.Sleep(time.Second * 3)
	stepState := StepStateFromContext(ctx)

	lessonRepo := &repositories_bob.LessonRepo{}
	presetStudyPlanWeeklyRepo := &repositories.PresetStudyPlanWeeklyRepo{}
	lessonGroupRepo := &repositories_bob.LessonGroupRepo{}
	topicRepo := &repositories_bob.TopicRepo{}

	for _, l := range stepState.Request.([]*npb.EventMasterRegistration_Lesson) {
		var (
			lessons []*entities.Lesson
		)
		try.Do(func(attempt int) (retry bool, err error) {
			lessons, err = lessonRepo.Find(auth.InjectFakeJwtToken(ctx, fmt.Sprint(constants.JPREPSchool)), s.DBTrace, &repositories_bob.LessonFilter{
				LessonID: database.TextArray([]string{
					l.LessonId,
				}),
				TeacherID: database.TextArray(nil),
				CourseID:  database.TextArray(nil),
			})
			if err != nil || len(lessons) == 0 {
				time.Sleep(2 * time.Second)
				return attempt < 10, fmt.Errorf("er FindLesson: %w", err)
			}
			return false, nil
		})

		switch l.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			if len(lessons) == 0 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("live lesson not found")
			}

			lesson := lessons[0]
			p, err := presetStudyPlanWeeklyRepo.FindByLessonID(ctx, s.DBTrace, database.Text(l.LessonId))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("err find presetStudyPlanWeekly: %w", err)
			}

			if l.CourseId != lesson.CourseID.String {
				return StepStateToContext(ctx, stepState), fmt.Errorf("courseID does not match, expected: %s, got: %s", l.CourseId, lesson.CourseID.String)
			}

			if p.StartDate.Time.Unix() != l.StartDate.Seconds {
				return StepStateToContext(ctx, stepState), fmt.Errorf("startDate does not match, expected: %s, got: %s", p.StartDate.Time.String(), l.StartDate.AsTime())
			}

			if p.EndDate.Time.Unix() != l.EndDate.Seconds {
				return StepStateToContext(ctx, stepState), fmt.Errorf("endDate does not match, expected: %s, got: %s", p.EndDate.Time.String(), l.EndDate.AsTime())
			}

			if l.LessonGroup != lesson.LessonGroupID.String {
				return StepStateToContext(ctx, stepState), fmt.Errorf("endDate does not match, expected: %s, got: %s", l.LessonGroup, lesson.LessonGroupID.String)
			}

			if l.LessonGroup != "" {
				_, err := lessonGroupRepo.Get(ctx, s.DBTrace, database.Text(l.LessonGroup), lesson.CourseID)
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}
			}

			t, err := topicRepo.RetrieveByID(ctx, s.DBTrace, p.TopicID)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("err find topic: %w", err)
			}

			if t.Name.String != l.ClassName {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lessonName does not match, expected: %s, got: %s", l.ClassName, t.Name.String)
			}

			if lesson.CenterID.String != constants.JPREPOrgLocation {
				return StepStateToContext(ctx, stepState), fmt.Errorf("JPREP center id is not correct: %s", lesson.CenterID.String)
			}
		case npb.ActionKind_ACTION_KIND_DELETED:
			if len(lessons) != 0 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("live lesson not deleted")
			}
		}
	}
	return s.bobMustPushMsgSubjectToNats(StepStateToContext(ctx, stepState), "EvtLesson from Jprep", constants.SubjectLessonCreated)
}
func (s *suite) someExistedLessonInDatabase(ctx context.Context) (context.Context, error) {
	total := rand.Intn(5) + 1
	return s.jprefSyncLessonWithActionAndLessonWithAction(ctx, strconv.Itoa(total), npb.ActionKind_ACTION_KIND_UPSERTED.String(), "0", "")
}
func (s *suite) theseLessonUpdatedType(ctx context.Context, arg string) (context.Context, error) {
	time.Sleep(time.Second * 2)
	stepState := StepStateFromContext(ctx)
	lessonIds := make([]string, 0)
	for _, l := range stepState.ExistedLessons {
		lessonIds = append(lessonIds, l.LessonId)
	}
	stmt := `UPDATE lessons SET lesson_type = $1 WHERE lesson_id = ANY($2::_TEXT)`
	commandTag, err := s.DBTrace.DB.Exec(ctx, stmt, arg, lessonIds)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if commandTag.RowsAffected() == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable update: no row affected")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) jrefSyncSomeNewLessonWithActionAndSomeExistedLessonWithAction(ctx context.Context, newLessonAction, existedLessonAction string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	totalnewLesson := rand.Intn(5) + 1
	ctx, newLessons, err := s.toLiveLessonMsg(ctx, "new", newLessonAction, totalnewLesson)
	if err != nil {
		return ctx, err
	}
	lessons := append(newLessons, stepState.ExistedLessons...)

	stepState.Request = lessons

	req := &npb.EventMasterRegistration{RawPayload: []byte("{}"), Signature: idutil.ULIDNow(), Lessons: lessons}
	data, _ := proto.Marshal(req)
	_, err = s.JSM.PublishContext(ctx, constants.SubjectSyncMasterRegistration, data)
	if err != nil {
		return ctx, fmt.Errorf("Publish: %w", err)
	}
	time.Sleep(time.Second * 3)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theseLessonHaveToDeleted(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(time.Second * 10)
	lessonRepo := repositories.LessonRepo{}
	lessonIds := make([]string, 0, len(stepState.ExistedLessons))
	for _, l := range stepState.ExistedLessons {
		lessonIds = append(lessonIds, l.LessonId)
	}
	err := lessonRepo.SoftDelete(ctx, s.DBTrace.DB, database.TextArray(lessonIds))
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) jprepSyncSomeLessonsToStudent(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.jprepSyncLessonMembersWithActionAndLessonMembersWithAction(ctx, "1",
		npb.ActionKind_ACTION_KIND_UPSERTED.String(),
		"0",
		npb.ActionKind_ACTION_KIND_UPSERTED.String(),
		0,
	)

	ctx, err2 := s.theseLessonMembersMustBeStoreInOurSystem(ctx)
	stepState := StepStateFromContext(ctx)
	// drain old subscription for subject SyncStudentLessonsConversations.Synced, because in next step we have different handler flow for this subject
	if len(stepState.Subs) > 0 {
		if stepState.Subs[0].IsValid() {
			err := stepState.Subs[0].Drain()
			if err != nil {
				return StepStateToContext(ctx, stepState), errors.New("failed to drain old subscription")
			}
		}
	}
	// only care about 1 student
	return StepStateToContext(ctx, stepState), multierr.Combine(err1, err2)
}
func (s *suite) jprepResyncLessonMembersButExcludingALesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	newReq := stepState.Request.([]*npb.EventSyncUserCourse_StudentLesson)[0]
	lessons := newReq.LessonIds
	remainingLessonMember := lessons[:len(lessons)-1]
	stepState.RemovedStudentLessons = lessons[len(lessons)-1:]
	newReq.LessonIds = remainingLessonMember
	stepState.Request = newReq
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	handler := func(ctx context.Context, data []byte) (bool, error) {
		r := &npb.EventSyncUserCourse{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return false, err
		}
		if len(r.GetStudentLessons()) != 1 {
			return false, fmt.Errorf("expected len(r.GetStudentLessons()) = %v, got %v", 1, len(r.GetStudentLessons()))
		}
		if !stringutil.SliceElementsMatch(r.GetStudentLessons()[0].LessonIds, stepState.RemovedStudentLessons) {
			return false, fmt.Errorf("not matched in result of compare two arr of lesson_ids")
		}
		stepState.FoundChanForJetStream <- struct{}{}
		return false, nil
	}

	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(stepState.RequestSentAt),
			nats.AckWait(2 * time.Second),
		},
	}

	sub, err := s.JSM.Subscribe(constants.SubjectSyncStudentLessons, opts, handler)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot subscribe to NATS: %v", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	signature := idutil.ULIDNow()
	ctx, err = s.createPartnerSyncDataLog(ctx, signature, 0)
	if err != nil {
		return ctx, fmt.Errorf("create partner sync data log error: %w", err)
	}
	req := &npb.EventSyncUserCourse{RawPayload: []byte("{}"), Signature: signature, StudentLessons: []*npb.EventSyncUserCourse_StudentLesson{newReq}}
	data, _ := proto.Marshal(req)
	_, err = s.JSM.PublishContext(ctx, constants.SubjectJPREPSyncUserCourseNatsJS, data)
	if err != nil {
		return ctx, fmt.Errorf("s.JSM.PublishContext: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) yasuoMustPushEventRemovingLessonMembersToForExcludedLesson(ctx context.Context, qName string) (context.Context, error) {
	time.Sleep(500 * time.Millisecond)
	stepState := StepStateFromContext(ctx)

	timer := time.NewTimer(time.Minute)
	defer timer.Stop()

	select {
	case <-stepState.FoundChanForJetStream:
		if qName != constants.SubjectSyncStudentLessons {
			return ctx, fmt.Errorf("expect qName = %v", constants.SubjectSyncStudentLessons)
		}
		return ctx, nil
	case <-timer.C:
		return StepStateToContext(ctx, stepState), errors.New("time out")
	}
}

func (s *suite) aUpsertLiveCourseRequestWithMissing(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.UpsertLiveCourseRequest{
		Id:         stepState.CurrentCourseID,
		Name:       "live-course " + s.Random,
		Grade:      "G12",
		Subject:    bpb.SUBJECT_BIOLOGY,
		ClassIds:   []int32{stepState.CurrentClassID},
		TeacherIds: []string{stepState.CurrentTeacherID},
		SchoolId:   int64(stepState.CurrentSchoolID),
		Country:    bpb.COUNTRY_VN,
		StartDate:  &types.Timestamp{Seconds: time.Now().Unix()},
		EndDate:    &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
	}
	if arg1 == "name" {
		req.Name = ""
	}
	if arg1 == "id" {
		req.Id = ""
	}
	if arg1 == "grade" {
		req.Grade = ""
	}
	if arg1 == "subject" {
		req.Subject = bpb.SUBJECT_NONE
	}
	if arg1 == "classIds" {
		req.ClassIds = []int32{}
	}
	if arg1 == "teacherIds" {
		req.TeacherIds = []string{}
	}
	if arg1 == "teacherIdsIsMine" {
		t, _ := jwt.ParseString(stepState.AuthToken)
		req.TeacherIds = []string{t.Subject()}
	}
	if arg1 == "teacherIdsIsMineAndMissingId" {
		req.Id = ""
		t, _ := jwt.ParseString(stepState.AuthToken)
		req.TeacherIds = []string{t.Subject()}
	}
	if arg1 == "schoolId" {
		req.SchoolId = 0
	}
	if arg1 == "country" {
		req.Country = bpb.COUNTRY_NONE
	}
	if arg1 == "invalid date" {
		req.EndDate = &types.Timestamp{Seconds: time.Now().Unix()}
		req.StartDate = &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()}
	}
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpsertLiveCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewCourseServiceClient(s.Conn).UpsertLiveCourse(contextWithToken(s, ctx), stepState.Request.(*pb.UpsertLiveCourseRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) yasuoMustStoreLiveCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	req := stepState.Request.(*pb.UpsertLiveCourseRequest)
	resp := stepState.Response.(*pb.UpsertLiveCourseResponse)
	stepState.CurrentCourseID = resp.Id
	if len(req.TeacherIds) != 0 {
		stepState.CurrentTeacherID = req.TeacherIds[0]
	}

	courseRepo := &repositories_bob.CourseRepo{}
	courseInDB, err := courseRepo.FindByID(ctx, s.DBTrace, database.Text(stepState.CurrentCourseID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	grade, _ := i18n.ConvertStringGradeToInt(req.Country, req.Grade)
	if int32(courseInDB.Grade.Int) != int32(grade) {
		return StepStateToContext(ctx, stepState), errors.Errorf("not same grade id expect %d, got %d", grade, courseInDB.Grade.Int)
	}
	if courseInDB.Name.String != req.Name {
		return StepStateToContext(ctx, stepState), errors.Errorf("not same name expect %s, got %s", req.Name, courseInDB.Name.String)
	}
	if courseInDB.Subject.String != req.Subject.String() {
		return StepStateToContext(ctx, stepState), errors.Errorf("not same subject expect %s, got %s", req.Subject.String(), courseInDB.Subject.String)
	}
	if courseInDB.Country.String != req.Country.String() {
		return StepStateToContext(ctx, stepState), errors.Errorf("not same country expect %s, got %s", req.Country.String(), courseInDB.Country.String)
	}
	if courseInDB.SchoolID.Int != int32(req.SchoolId) {
		return StepStateToContext(ctx, stepState), errors.Errorf("not same school id expect %d, got %d", req.SchoolId, courseInDB.SchoolID.Int)
	}
	if len(courseInDB.TeacherIDs.Elements) != len(req.TeacherIds) {
		return ctx, fmt.Errorf("not same teacher ids expect %v, got %v", req.TeacherIds, courseInDB.TeacherIDs.Elements)
	}

	courseClassRepo := &repositories.CourseClassRepo{}
	courseClassInDB, err := courseClassRepo.FindByCourseID(ctx, s.DBTrace, database.Text(stepState.CurrentCourseID), true)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	classIDs := []int32{}
	for classID, v := range courseClassInDB {
		deletedAt := v.DeletedAt.Time
		if v.Status.String == entities.CourseClassStatusActive {
			if !deletedAt.IsZero() {
				return ctx, fmt.Errorf("deleted at is not null for course %s class %d", v.CourseID.String, v.ClassID.Int)
			}
			classIDs = append(classIDs, classID.Int)
		}

		if v.Status.String == entities.CourseClassStatusInActive && deletedAt.IsZero() {
			return ctx, fmt.Errorf("deleted at is null for course %s class %d", v.CourseID.String, v.ClassID.Int)
		}
	}

	if len(req.ClassIds) != 0 && !reflect.DeepEqual(req.ClassIds, classIDs) {
		return StepStateToContext(ctx, stepState), errors.New("not same class ids")
	}
	if courseInDB.PresetStudyPlanID.String == "" {
		return StepStateToContext(ctx, stepState), errors.New("missing preset study plan")
	}

	// check preset study plan
	presetStudyPlanRepo := &repositories.PresetStudyPlanRepo{}
	presetStudyPlanInDB, err := presetStudyPlanRepo.Get(ctx, s.DBTrace, courseInDB.PresetStudyPlanID)
	if err != nil {
		return ctx, fmt.Errorf("presetStudyPlanRepo.Get: %w", err)
	}
	if presetStudyPlanInDB.Grade.Int != courseInDB.Grade.Int {
		return StepStateToContext(ctx, stepState), errors.Errorf("preset study plan not same grade expect %d got %d", courseInDB.Grade.Int, presetStudyPlanInDB.Grade.Int)
	}
	if presetStudyPlanInDB.Name.String != courseInDB.Name.String {
		return StepStateToContext(ctx, stepState), errors.Errorf("preset study plan not same name expect %s got %s", courseInDB.Name.String, presetStudyPlanInDB.Name.String)
	}
	if presetStudyPlanInDB.Subject.String != courseInDB.Subject.String {
		return StepStateToContext(ctx, stepState), errors.Errorf("preset study plan not same subject expect %s got %s", courseInDB.Subject.String, presetStudyPlanInDB.Subject.String)
	}
	if presetStudyPlanInDB.Country.String != courseInDB.Country.String {
		return StepStateToContext(ctx, stepState), errors.Errorf("preset study plan not same country expect %s got %s", courseInDB.Country.String, presetStudyPlanInDB.Country.String)
	}

	return StepStateToContext(ctx, stepState), nil
}
