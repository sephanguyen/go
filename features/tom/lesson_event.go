package tom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/godogutil"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/tom/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func (s *suite) aEvtLessonWithMessage(ctx context.Context, event string) (context.Context, error) {
	switch event {
	case "CreateLesson":
		s.lessonID = idutil.ULID(timeutil.Now())
		s.lessonName = fmt.Sprintf("%s-lesson-name", s.lessonID)
		lessons := []*pb.EvtLesson_Lesson{
			{
				LessonId: s.lessonID,
				Name:     s.lessonName,
			},
		}
		s.Request = &pb.EvtLesson{
			Message: &pb.EvtLesson_CreateLessons_{
				CreateLessons: &pb.EvtLesson_CreateLessons{
					Lessons: lessons,
				},
			},
		}
	case "EndLesson":
		if s.lessonID == "" {
			return ctx, fmt.Errorf("lesson id is empty, check previous step")
		}
		s.Request = &pb.EvtLesson{
			Message: &pb.EvtLesson_EndLiveLesson_{
				EndLiveLesson: &pb.EvtLesson_EndLiveLesson{
					LessonId: s.lessonID,
				},
			},
		}
	case "JoinLesson":
		if s.lessonID == "" {
			s.lessonID = idutil.ULID(timeutil.Now())
		}
		s.Request = &pb.EvtLesson{
			Message: &pb.EvtLesson_JoinLesson_{
				JoinLesson: &pb.EvtLesson_JoinLesson{
					LessonId:  s.lessonID,
					UserGroup: pb.USER_GROUP_STUDENT,
					UserId:    "",
				},
			},
		}
	}

	return ctx, nil
}

func (s *suite) bobSendEventEvtLesson(ctx context.Context) (context.Context, error) {
	data, err := s.Request.(*pb.EvtLesson).Marshal()
	if err != nil {
		return ctx, err
	}
	subject := ""
	s.RequestAt = time.Now()
	message := s.Request.(*pb.EvtLesson).Message
	switch message.(type) {
	case *pb.EvtLesson_CreateLessons_:
		subject = constants.SubjectLessonCreated
	default:
		subject = constants.SubjectLessonUpdated
	}
	_, err = s.JSM.TracedPublish(ctx, "bobSendEventEvtLesson", subject, data)
	if err != nil {
		return ctx, fmt.Errorf("s.JSM.TracedPublish: %w", err)
	}
	return ctx, nil
}

func (s *suite) tomMustCreateConversationForAllLesson(ctx context.Context) (context.Context, error) {
	req := s.Request.(*pb.EvtLesson)
	for _, lesson := range req.GetCreateLessons().Lessons {
		if err := doRetry(func() (bool, error) {
			query := `SELECT cl.conversation_id FROM conversation_lesson cl LEFT JOIN conversations c ON cl.conversation_id = c.conversation_id
    WHERE cl.lesson_id = $1 AND c.name = $2 AND c.status= 'CONVERSATION_STATUS_NONE'`
			var conversationID pgtype.Text
			if err := s.DB.QueryRow(ctx, query, lesson.LessonId, lesson.Name).Scan(&conversationID); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return true, fmt.Errorf("conversation not inserted")
				}
				return false, err
			}
			s.LessonConversationMap[lesson.LessonId] = conversationID.String
			return false, nil
		}); err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func (s *suite) aValidIDInJoinLesson(ctx context.Context, userGroup string) (context.Context, error) {
	var userID string
	switch userGroup {
	case pb.USER_GROUP_STUDENT.String():
		stu, err := s.CommonSuite.CreateStudent(ctx, []string{s.filterSuiteState.defaultLocationID}, nil)
		if err != nil {
			return ctx, err
		}
		s.studentID = stu.UserProfile.UserId
		token, err := generateValidAuthenticationToken(s.studentID, userGroup)
		if err != nil {
			return ctx, err
		}
		s.studentToken = token
		s.StudentsInLesson = append(s.StudentsInLesson, s.studentID)
		userID = s.studentID
	case pb.USER_GROUP_TEACHER.String():
		profile, tok, err := s.CommonSuite.CreateTeacher(ctx)
		if err != nil {
			return ctx, err
		}
		s.teacherID = profile.StaffId
		s.TeacherToken = tok
		s.TeachersInLesson = append(s.TeachersInLesson, s.teacherID)
		userID = s.teacherID
	}

	s.Request.(*pb.EvtLesson).GetJoinLesson().UserId = userID
	s.Request.(*pb.EvtLesson).GetJoinLesson().UserGroup = pb.UserGroup(pb.UserGroup_value[userGroup])

	return ctx, nil
}

func (s *suite) tomAddAboveUserToThisLessonConversation(ctx context.Context, actionType string) (context.Context, error) {
	var maxAttempt = 5
	// nolint:goconst
	switch actionType {
	case "must":
	case "must not":
		maxAttempt = 3
	default:
		panic(fmt.Sprintf("invalid action type %s", actionType))
	}
	var notFoundErr = errors.New("not found user in conversation")
	convID := s.LessonChatState.LessonConversationMap[s.LessonChatState.lessonID]
	err := try.Do(func(attempt int) (retry bool, err error) {
		time.Sleep(2 * time.Second)
		ctx2, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		cStatusRepo := &repositories.ConversationMemberRepo{}
		var (
			userID    = s.Request.(*pb.EvtLesson).GetJoinLesson().UserId
			userGroup = s.Request.(*pb.EvtLesson).GetJoinLesson().UserGroup.String()
		)
		conversationMembers, err := cStatusRepo.FindByConversationID(ctx2, s.DB, database.Text(convID))
		if err != nil {
			return attempt < maxAttempt, err
		}
		found := false
		for _, m := range conversationMembers {
			if m.UserID.String == userID && m.Role.String == userGroup {
				found = true
				break
			}
		}
		if !found {
			return attempt < maxAttempt, notFoundErr
		}
		return false, nil
	})
	if actionType == "must not" {
		if errors.Is(err, notFoundErr) {
			return ctx, nil
		}
		return ctx, err
	}
	return ctx, err
}

func (s *suite) tomMustSendEndLiveLessonMessageAndRemoveAllMembersFromConversation(ctx context.Context) (context.Context, error) {
	return ctx, try.Do(func(attempt int) (bool, error) {
		ctx2, err := godogutil.MultiErrChain(ctx,
			s.returnsStatusCode, "OK",
			s.tomRemoveAllStudentsFromCurrentLessonConversation,
			s.tomStoreMessageInThisConversation, "must", "CODES_MESSAGE_TYPE_END_LIVE_LESSON",
		)
		if err == nil {
			ctx = ctx2
			return false, nil
		}

		time.Sleep(time.Duration(attempt) * 250 * time.Millisecond)
		return attempt < 5, err
	})
}

func (s *suite) tomRemoveAllStudentsFromCurrentLessonConversation(ctx context.Context) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT status
		FROM conversation_members cm
		INNER JOIN conversation_lesson cl ON cl.conversation_id = cm.conversation_id
		WHERE cl.lesson_id = $1
	`

	rows, err := s.DB.Query(ctx2, query, s.Request.(*pb.EvtLesson).GetEndLiveLesson().LessonId)
	if err != nil {
		return ctx, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			return ctx, err
		}
		if status != entities.ConversationStatusInActive {
			return ctx, fmt.Errorf("unexpected status, got: %q, want: %q", status, entities.ConversationStatusInActive)
		}
	}
	if err := rows.Err(); err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (s *suite) bobSendLeaveLessonForOneOfPrevious(ctx context.Context, role string) (context.Context, error) {
	userID := ""
	switch role {
	case "student":
		userID = s.LessonChatState.studentsInLesson[0]
		s.LessonChatState.studentsInLesson = s.LessonChatState.studentsInLesson[1:]
	case "teacher":
		userID = s.LessonChatState.TeachersInLesson[0]
		s.LessonChatState.TeachersInLesson = s.LessonChatState.TeachersInLesson[1:]
	}
	evt := &pb.EvtLesson{
		Message: &pb.EvtLesson_LeaveLesson_{
			LeaveLesson: &pb.EvtLesson_LeaveLesson{
				LessonId: s.LessonChatState.lessonID,
				UserId:   userID,
			},
		},
	}
	s.Request = evt
	return s.bobSendEventEvtLesson(ctx)
}

func (s *suite) aEvtLessonWithMessageCreateLessonWithStudents(ctx context.Context, studentnum int) (context.Context, error) {
	ctx, err := s.aEvtLessonWithMessage(ctx, "CreateLesson")
	if err != nil {
		return ctx, err
	}
	studentIDs := make([]string, 0, studentnum)
	if !s.CommonSuite.ContextHasToken(ctx) {
		ctx2, err := s.CommonSuite.ASignedInWithSchool(ctx, "school admin", constants.ManabieSchool)
		if err != nil {
			return ctx, err
		}
		ctx = ctx2
	}
	for i := 0; i < studentnum; i++ {
		// id := idutil.ULIDNow()
		// studentIDs = append(studentIDs, id)
		stu, err := s.CommonSuite.CreateStudent(ctx, []string{s.filterSuiteState.defaultLocationID}, nil)
		if err != nil {
			return ctx, err
		}
		studentIDs = append(studentIDs, stu.UserProfile.UserId)
	}
	s.LessonChatState.studentsInLesson = studentIDs
	s.Request.(*pb.EvtLesson).GetCreateLessons().Lessons[0].LearnerIds = studentIDs
	return ctx, nil
}

func (s *suite) tomMustRemoveTeacherFromConversation(ctx context.Context) (context.Context, error) {
	teacher := s.Request.(*pb.EvtLesson).GetLeaveLesson().UserId
	convID := s.LessonChatState.LessonConversationMap[s.lessonID]
	repo := &repositories.ConversationMemberRepo{}

	return ctx, try.Do(func(attempt int) (bool, error) {
		members, err := repo.FindByConversationID(ctx, s.DB, database.Text(convID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return false, nil
			}
			return false, err
		}
		if _, ok := members[database.Text(teacher)]; !ok {
			return false, nil
		}
		time.Sleep(1 * time.Second)
		return attempt < 5, fmt.Errorf("teacher is still in conversation")
	})
}

func (s *suite) tomMustNotRemoveStudentFromConversation(ctx context.Context) (context.Context, error) {
	studentID := s.Request.(*pb.EvtLesson).GetLeaveLesson().UserId
	convID := s.LessonChatState.LessonConversationMap[s.lessonID]
	repo := &repositories.ConversationMemberRepo{}

	for i := 0; i < 3; i++ {
		members, err := repo.FindByConversationID(ctx, s.DB, database.Text(convID))
		if err != nil {
			return ctx, err
		}
		if _, ok := members[database.Text(studentID)]; ok {
			time.Sleep(1 * time.Second)
			continue
		}
		return ctx, fmt.Errorf("student should not be removed from conversation")
	}
	return ctx, nil
}

func (s *suite) tomMustCreateConversationMemberForStudentInCreateLesson(ctx context.Context) (context.Context, error) {
	studentIDs := s.LessonChatState.studentsInLesson
	convID := s.LessonChatState.LessonConversationMap[s.lessonID]
	repo := &repositories.ConversationMemberRepo{}

	return ctx, doRetry(func() (bool, error) {
		members, err := repo.FindByConversationID(ctx, s.DB, database.Text(convID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return true, err
			}
			return false, err
		}

		for _, studentID := range studentIDs {
			if membership, ok := members[database.Text(studentID)]; !ok {
				return true, fmt.Errorf("missing student id after sending event CreateLesson")
			} else if membership.Status.String != entities.ConversationStatusActive {
				return false, fmt.Errorf("student membership is not active")
			}
		}
		return false, nil
	})
}

func (s *suite) aLessonConversationWithTeachersAndStudents(ctx context.Context, numTeacher, numStudent int) (context.Context, error) {
	ctx, err := godogutil.MultiErrChain(ctx,
		s.aEvtLessonWithMessageCreateLessonWithStudents, numStudent,
		s.bobSendEventEvtLesson,
		s.tomMustCreateConversationForAllLesson,
		s.tomMustCreateConversationMemberForStudentInCreateLesson,
	)
	if err != nil {
		return ctx, err
	}
	for i := 0; i < numTeacher; i++ {
		ctx, err := godogutil.MultiErrChain(ctx,
			s.aEvtLessonWithMessage, "JoinLesson",
			s.aValidIDInJoinLesson, cpb.UserGroup_USER_GROUP_TEACHER.String(),
			s.bobSendEventEvtLesson,
			s.tomAddAboveUserToThisLessonConversation, "must",
		)
		if err != nil {
			return ctx, err
		}
	}
	if len(s.LessonChatState.TeachersInLesson) > 0 {
		s.LessonChatState.firstTeacher = s.LessonChatState.TeachersInLesson[0]
	}
	return ctx, nil
}

// for Gandalf
func (s *suite) ALessonConversationBackground(ctx context.Context) (context.Context, error) {
	return s.aLessonConversationWithTeachersAndStudents(ctx, 1, 1)
}

func (s *suite) bobSendUpdateLessonWithNewStudentAndWithoutPreviousStudents(ctx context.Context, numStudent int, excludeOldStudents int) (context.Context, error) {
	oldstudent := s.LessonChatState.studentsInLesson
	remainedStudents := oldstudent[excludeOldStudents:]
	for i := 0; i < numStudent; i++ {
		remainedStudents = append(remainedStudents, idutil.ULIDNow())
	}
	s.LessonChatState.studentsInLesson = remainedStudents
	req := &pb.EvtLesson{
		Message: &pb.EvtLesson_UpdateLesson_{
			UpdateLesson: &pb.EvtLesson_UpdateLesson{
				LessonId:   s.LessonChatState.lessonID,
				ClassName:  s.LessonChatState.lessonName,
				LearnerIds: remainedStudents,
			},
		},
	}
	s.Request = req
	return s.bobSendEventEvtLesson(ctx)
}

func (s *suite) yasuoSendEventSyncUserCourseInsertingStudentsAndDeletingPrevousStudentForCurrentLesson(ctx context.Context, newstudents, removedstudents int) (context.Context, error) {
	oldstudent := s.LessonChatState.studentsInLesson
	removeStudents := oldstudent[:removedstudents]
	remainedIDs := oldstudent[removedstudents:]
	newIDs := make([]string, 0, newstudents)
	for i := 0; i < newstudents; i++ {
		newIDs = append(newIDs, idutil.ULIDNow())
	}
	studentLesson := []*npb.EventSyncUserCourse_StudentLesson{}
	for _, id := range newIDs {
		studentLesson = append(studentLesson, &npb.EventSyncUserCourse_StudentLesson{
			ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			LessonIds:  []string{s.LessonChatState.lessonID},
			StudentId:  id,
		})
	}
	for _, id := range removeStudents {
		studentLesson = append(studentLesson, &npb.EventSyncUserCourse_StudentLesson{
			ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
			LessonIds:  []string{s.LessonChatState.lessonID},
			StudentId:  id,
		})
	}
	time.Sleep(2 * time.Second)

	evt := &npb.EventSyncUserCourse{
		StudentLessons: studentLesson,
	}
	s.Request = evt
	newIDs = append(newIDs, remainedIDs...)
	s.LessonChatState.studentsInLesson = newIDs

	data, err := proto.Marshal(evt)
	if err != nil {
		return ctx, err
	}

	_, err = s.JSM.PublishContext(ctx, constants.SubjectSyncStudentLessons, data)
	if err != nil {
		return ctx, fmt.Errorf("s.JSM.PublishContext: %w", err)
	}
	return ctx, nil
}

func (s *suite) tomMustCorrectlyStoreOnlyLatestStudentsInLessonConversation(ctx context.Context) (context.Context, error) {
	convID := s.LessonChatState.LessonConversationMap[s.LessonChatState.lessonID]
	return ctx, try.Do(func(attempt int) (bool, error) {
		repo := &repositories.ConversationMemberRepo{}
		mems, err := repo.FindByConversationID(ctx, s.DB, database.Text(convID))
		if err != nil {
			return false, err
		}
		studentCount := 0
		for _, mem := range mems {
			if mem.Role.String != cpb.UserGroup_USER_GROUP_STUDENT.String() {
				continue
			}
			studentCount++
		}
		for _, id := range s.LessonChatState.studentsInLesson {
			mem, exist := mems[database.Text(id)]
			if !exist || mem.Status.String != entities.ConversationStatusActive {
				time.Sleep(1 * time.Second)
				return attempt < 5, fmt.Errorf("student does not have membership")
			}
		}
		return false, nil
	})
}
