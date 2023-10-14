package lessonmgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	user_pbv2 "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) studentInfoWithFirstNameLastNameAndPhoneticName(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	randomID := newID()
	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = constant.ManabieSchool
	}
	req := &user_pbv2.CreateStudentRequest{
		SchoolId: stepState.CurrentSchoolID,
		StudentProfile: &user_pbv2.CreateStudentRequest_StudentProfile{
			Email:             fmt.Sprintf("%v@example.com", randomID),
			Password:          fmt.Sprintf("password-%v", randomID),
			FirstName:         fmt.Sprintf("user-first-name-%v", randomID),
			LastName:          fmt.Sprintf("user-last-name-%v", randomID),
			FirstNamePhonetic: fmt.Sprintf("user-first-name-phonetic%v", randomID),
			LastNamePhonetic:  fmt.Sprintf("user-last-name-phonetic%v", randomID),
			CountryCode:       cpb.Country_COUNTRY_VN,
			EnrollmentStatus:  user_pbv2.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			PhoneNumber:       fmt.Sprintf("phone-number-%v", randomID),
			StudentExternalId: fmt.Sprintf("student-external-id-%v", randomID),
			StudentNote:       fmt.Sprintf("some random student note %v", randomID),
			Grade:             5,
			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:            user_pbv2.Gender_MALE,
			LocationIds:       []string{"01FR4M51XJY9E77GSN4QZ1Q9N1"},
		},
	}
	stepState.Request = req
	ctx, err := s.createStudentWithFirstNameAndLastNameSubscription(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createStudentWithFirstNameAndLastNameSubscription: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createStudentWithFirstNameAndLastNameSubscription(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}

	handleCreateUser := func(ctx context.Context, data []byte) (bool, error) {
		evtUser := &user_pbv2.EvtUser{}
		if err := proto.Unmarshal(data, evtUser); err != nil {
			return false, err
		}

		switch req := stepState.Request.(type) {
		case *user_pbv2.CreateStudentRequest:
			switch msg := evtUser.Message.(type) {
			case *user_pbv2.EvtUser_CreateStudent_:
				if req.StudentProfile.FirstName == msg.CreateStudent.StudentFirstName && req.StudentProfile.LastName == msg.CreateStudent.StudentLastName {
					stepState.FoundChanForJetStream <- evtUser.Message
					return true, nil
				}
			}
		}
		return false, nil
	}

	subs, err := s.JSM.Subscribe(constants.SubjectUserCreated, opts, handleCreateUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("createStudentSubscription: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) createNewStudentAccount(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.createStudentSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createStudentSubscription: %w", err)
	}
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = user_pbv2.NewUserModifierServiceClient(s.UserMgmtConn).CreateStudent(contextWithToken(s, ctx), stepState.Request.(*user_pbv2.CreateStudentRequest))

	if stepState.ResponseErr == nil {
		stepState.CurrentStudentID = stepState.
			Response.(*user_pbv2.CreateStudentResponse).
			GetStudentProfile().
			GetStudent().
			GetUserProfile().
			GetUserId()
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) createStudentSubscription(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}

	handleCreateUser := func(ctx context.Context, data []byte) (bool, error) {
		evtUser := &user_pbv2.EvtUser{}
		if err := proto.Unmarshal(data, evtUser); err != nil {
			return false, err
		}

		switch req := stepState.Request.(type) {
		case *user_pbv2.CreateStudentRequest:
			switch msg := evtUser.Message.(type) {
			case *user_pbv2.EvtUser_CreateStudent_:
				if req.StudentProfile.Name == msg.CreateStudent.StudentName {
					stepState.FoundChanForJetStream <- evtUser.Message
					return true, nil
				}
			}
		}
		return false, nil
	}

	subs, err := s.JSM.Subscribe(constants.SubjectUserCreated, opts, handleCreateUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("createStudentSubscription: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentAccountDataToUpdateWithFirstNameLastNameAndPhoneticName(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	uid := idutil.ULIDNow()
	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = constant.ManabieSchool
	}
	req := &user_pbv2.UpdateStudentRequest{
		SchoolId: stepState.CurrentSchoolID,
		StudentProfile: &user_pbv2.UpdateStudentRequest_StudentProfile{
			Id:                  stepState.CurrentStudentID,
			Grade:               int32(1),
			FirstName:           fmt.Sprintf("bdd-test-updated-first-name-" + uid),
			LastName:            fmt.Sprintf("bdd-test-updated-last-name" + uid),
			FirstNamePhonetic:   fmt.Sprintf("bdd-test-updated-first-name-phonetic" + uid),
			LastNamePhonetic:    fmt.Sprintf("bdd-test-updated-last-name-phonetic" + uid),
			EnrollmentStatus:    user_pbv2.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			EnrollmentStatusStr: "STUDENT_ENROLLMENT_STATUS_ENROLLED",
			StudentExternalId:   fmt.Sprintf("student-external-id-%v", uid),
			StudentNote:         fmt.Sprintf("some random student note edited %v", uid),
			Email:               fmt.Sprintf("student-email-edited-%s@example.com", uid),
			Birthday:            timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:              user_pbv2.Gender_MALE,
			LocationIds:         locationIDs,
		},
	}
	stepState.Request = req
	stepState.CurrentStudentFirstName = req.StudentProfile.FirstName
	stepState.CurrentStudentLastName = req.StudentProfile.LastName
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) updateStudentAccount(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*user_pbv2.UpdateStudentRequest)

	ctx, err := s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.updateStudentSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.updateStudentSubscription: %w", err)
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = tryUpdateStudent(contextWithToken(s, ctx), s.UserMgmtConn, req)
	return StepStateToContext(ctx, stepState), nil
}

func tryUpdateStudent(ctx context.Context, client *grpc.ClientConn, req *user_pbv2.UpdateStudentRequest) (*user_pbv2.UpdateStudentResponse, error) {
	var (
		resp = &user_pbv2.UpdateStudentResponse{}
		err  error
	)

	err = try.Do(func(attempt int) (bool, error) {
		resp, err = user_pbv2.NewUserModifierServiceClient(client).UpdateStudent(ctx, req)
		if err == nil {
			return false, nil
		}
		if attempt < 5 {
			time.Sleep(time.Second)
			return true, err
		}
		return false, err
	})

	return resp, err
}

func (s *Suite) updateStudentSubscription(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}

	handleUpdateUser := func(ctx context.Context, data []byte) (bool, error) {
		evtUser := &user_pbv2.EvtUser{}
		if err := proto.Unmarshal(data, evtUser); err != nil {
			return false, err
		}

		switch req := stepState.Request.(type) {
		case *user_pbv2.UpdateStudentRequest:
			switch msg := evtUser.Message.(type) {
			case *user_pbv2.EvtUser_UpdateStudent_:
				if req.StudentProfile.Id == msg.UpdateStudent.StudentId && req.StudentProfile.FirstName == msg.UpdateStudent.StudentFirstName && req.StudentProfile.LastName == msg.UpdateStudent.StudentLastName {
					stepState.FoundChanForJetStream <- evtUser.Message
					return true, nil
				}
			}
		}
		return false, nil
	}

	subs, err := s.JSM.Subscribe(constants.SubjectUserUpdated, opts, handleUpdateUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("updateStudentSubscription: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) assignStudentToStudentSubscriptions(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentID := stepState.CurrentStudentID
	studentIDWithCourseID := make([]string, 0, len(stepState.StudentIds)*2)
	for _, courseID := range s.CommonSuite.CourseIDs {
		studentIDWithCourseID = append(studentIDWithCourseID, studentID, courseID)
	}
	stepState.StartDate = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	stepState.EndDate = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := s.insertStudentSubscription(ctx, stepState.StartDate, stepState.EndDate, studentIDWithCourseID...)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not insert student subscription: %w", err)
	}
	stepState.StudentIDWithCourseID = studentIDWithCourseID

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) insertStudentSubscription(ctx context.Context, startAt, endAt time.Time, studentIDWithCourseID ...string) ([]string, error) {
	queueFn := func(b *pgx.Batch, studentID, courseID string) string {
		id := idutil.ULIDNow()
		query := `INSERT INTO lesson_student_subscriptions (student_subscription_id, subscription_id, student_id, course_id, start_at, end_at) VALUES ($1, $2, $3, $4, $5, $6)`
		b.Queue(query, id, id, studentID, courseID, startAt, endAt)
		return id
	}

	b := &pgx.Batch{}
	ids := make([]string, 0, len(studentIDWithCourseID))
	for i := 0; i < len(studentIDWithCourseID); i += 2 {
		ids = append(ids, queueFn(b, studentIDWithCourseID[i], studentIDWithCourseID[i+1]))
	}
	result := s.BobDB.SendBatch(ctx, b)
	defer result.Close()

	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		_, err := result.Exec()
		if err != nil {
			return nil, fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}
	return ids, nil
}
func (s *Suite) assignStudentToALesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessonID := s.CommonSuite.CurrentLessonID
	studentID := stepState.CurrentStudentID
	lessonMember := domain.LessonMember{
		LessonID:         lessonID,
		UserID:           studentID,
		AttendanceStatus: cpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT.String(),
		AttendanceRemark: "bad",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	query := `INSERT INTO lesson_members (lesson_id, user_id, attendance_status, attendance_remark, user_first_name, user_last_name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := s.BobDB.Exec(ctx, query, lessonMember.LessonID, lessonMember.UserID, lessonMember.AttendanceStatus, lessonMember.AttendanceRemark, "", "", lessonMember.CreatedAt, lessonMember.UpdatedAt)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentNameIsUpdatedCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.CurrentStudentFirstName == "" || stepState.CurrentStudentLastName == "" {
		return ctx, fmt.Errorf("stepState name is not updated correctly: contains empty value")
	}
	// lesson member
	var lessonMemberCount int
	queryLessonMember := `SELECT count(*) FROM lesson_members WHERE user_first_name = $1 AND user_last_name = $2 AND deleted_at IS NULL`
	err := try.Do(func(attempt int) (bool, error) {
		err := s.BobDBTrace.QueryRow(ctx, queryLessonMember, database.Text(stepState.CurrentStudentFirstName), database.Text(stepState.CurrentStudentLastName)).Scan(&lessonMemberCount)
		if err == nil && lessonMemberCount > 0 {
			return false, nil
		}
		retry := attempt < 5
		if retry {
			time.Sleep(5 * time.Second)
			return true, fmt.Errorf("lesson_members name has not been updated correctly")
		}
		return false, err
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	// student subscriptions
	var studentSubscriptionCount int
	queryStudentSubscription := `SELECT count(*) FROM lesson_student_subscriptions WHERE student_first_name = $1 AND student_last_name = $2 AND deleted_at IS NULL`
	err = try.Do(func(attempt int) (bool, error) {
		err = s.BobDBTrace.QueryRow(ctx, queryStudentSubscription, stepState.CurrentStudentFirstName, stepState.CurrentStudentLastName).Scan(&studentSubscriptionCount)
		if err == nil && lessonMemberCount > 0 {
			return false, nil
		}
		retry := attempt < 5
		if retry {
			time.Sleep(5 * time.Second)
			return true, fmt.Errorf("lesson_student_subscriptions name has not been updated correctly")
		}
		return false, err
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
