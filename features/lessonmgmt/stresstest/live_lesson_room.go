package stresstest

import (
	"context"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
	tpb "github.com/manabie-com/backend/pkg/genproto/tom"
	pb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb_v1 "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb_v1 "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"golang.org/x/sync/errgroup"
)

type ScenarioSimulateAStudyOnLiveLessonRoom struct {
	id              string
	st              *StressTest
	teacherSuites   []*Suite
	studentSuites   []*Suite
	lessonID        string
	courseID        string
	locationID      string
	teacherAccounts []*AccountInfo
	studentAccounts []*AccountInfo

	logs *ScenarioSimulateAStudyOnLiveLessonRoomLogs
}

type ScenarioSimulateAStudyOnLiveLessonRoomLogs struct {
	NumberFetchFailedRoomState int32
}

func NewScenarioSimulateAStudyOnLiveLessonRoom(st *StressTest, courseID, locationID string, numTeacher, numStudent int) (*ScenarioSimulateAStudyOnLiveLessonRoom, error) {
	if numTeacher < 2 || numTeacher > len(st.teacherAccounts) {
		return nil, fmt.Errorf("numTeacher could be not less than 2 or > teacherAccounts")
	}

	if numStudent < 2 || numStudent > len(st.studentAccounts) {
		return nil, fmt.Errorf("numStudent could be not less than 2 or > studentAccounts")
	}
	return &ScenarioSimulateAStudyOnLiveLessonRoom{
		id:              idutil.ULIDNow(),
		st:              st,
		courseID:        courseID,
		locationID:      locationID,
		teacherAccounts: st.teacherAccounts[0:numTeacher],
		studentAccounts: st.studentAccounts[0:numStudent],
		logs:            &ScenarioSimulateAStudyOnLiveLessonRoomLogs{},
	}, nil
}

// createLesson will create a lesson which contain all students, teachers and have some materials
func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) createLesson(ctx context.Context) error {
	suiteTpl := scn.st.NewSuite()
	suiteTpl.lessonSuite.CommonSuite.CurrentCourseID = scn.courseID
	suiteTpl.lessonSuite.CommonSuite.CenterIDs = []string{scn.locationID}
	ctxTpl := common.StepStateToContext(context.Background(), suiteTpl.lessonSuite.CommonSuite.StepState)
	teacherIDs := make([]string, 0, len(scn.teacherAccounts))
	for _, teacher := range scn.teacherAccounts {
		teacherIDs = append(teacherIDs, teacher.ID)
	}
	studentIDs := make([]string, 0, len(scn.studentAccounts))
	for _, student := range scn.studentAccounts {
		studentIDs = append(studentIDs, student.ID)
	}
	lessonID, err := suiteTpl.CreateALiveLessonWithTeachersAndStudents(ctxTpl, teacherIDs, studentIDs)
	if err != nil {
		return fmt.Errorf("CreateALiveLessonWithTeachersAndStudents: %w", err)
	}
	scn.lessonID = lessonID

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) Prepare(ctx context.Context) error {
	// create lesson which contain all students teachers, and have some materials
	if err := scn.createLesson(ctx); err != nil {
		return fmt.Errorf("CreateLesson: %w", err)
	}

	signedInFunc := func(acc *AccountInfo) (*Suite, error) {
		suite := scn.st.NewSuite()
		suite.lessonSuite.CommonSuite.StepState.CurrentLessonID = scn.lessonID
		suite.lessonSuite.CommonSuite.StepState.CurrentCourseID = scn.courseID
		ctxS := common.StepStateToContext(context.Background(), suite.lessonSuite.CommonSuite.StepState)
		if err := suite.ASignedInWithAccInfo(ctxS, acc); err != nil {
			return nil, fmt.Errorf("suite.ASignedInWithAccInfo: %w", err)
		}
		return suite, nil
	}

	teachersSuites := make([]*Suite, 0, len(scn.teacherAccounts))
	// create suites for attendees
	for _, acc := range scn.teacherAccounts {
		suite, err := signedInFunc(acc)
		if err != nil {
			return err
		}
		teachersSuites = append(teachersSuites, suite)
	}
	scn.teacherSuites = teachersSuites

	studentSuites := make([]*Suite, 0, len(scn.studentAccounts))
	for _, acc := range scn.studentAccounts {
		suite, err := signedInFunc(acc)
		if err != nil {
			return err
		}
		studentSuites = append(studentSuites, suite)
	}
	scn.studentSuites = studentSuites

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) EnterSiteAndPrepareJoinRoom(ctx context.Context) error {
	g := new(errgroup.Group)
	for i := range scn.teacherSuites {
		suite := scn.teacherSuites[i]
		num := i
		g.Go(func() error {
			if err := suite.AfterLoginAndEnterALiveLessonRoom_Teacher(ctx); err != nil {
				return fmt.Errorf("Scenario_SimualteAStudyOnLiveLessonRoom_Teacher: suite %d : %w", num, err)
			}
			return nil
		})
	}

	for i := range scn.studentSuites {
		suite := scn.studentSuites[i]
		num := i
		g.Go(func() error {
			if err := suite.AfterLoginAndEnterALiveLessonRoom_Student(ctx); err != nil {
				return fmt.Errorf("AfterLoginAndEnterALiveLessonRoom_Student: suite %d : %w", num, err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

// Execute will Execute Scenario: A live class room
//
//		Given '2' teacher and '10' students singed in to site
//		And all of them belong to a live lesson
//		When all attendees join live room of above lesson
//	 And 'first' teacher request recording live room successfully
//	 And 'first' teacher share a pdf in '9 seconds'
//	 And there are '7' students raise hand but after that have '2' students hand off
//	 And there is one student spam raise hand button '10 time' in '2 seconds'
//	 And 'second' teacher fold hand of all students
//	 And waiting the student who be chosen answer question in '5 seconds'
//	 And 'second' teacher stop sharing and share a video in '10 seconds'
//	 And 'second' teacher pause video in '5 second'
//	 And 'first' teacher open a polling with '5' options
//	 And waiting all student answer question in '10 seconds'
//	 And 'first' teacher stop the polling and check answer of students
//	 And waiting for explaining answer in '6 seconds'
//	 And 'first' teacher end the polling
//	 And 'second' teacher resume video and let it play in '5 second'
//	 And 'second' teacher 'stop' video
//	 And 'second' teacher share a pdf in '5 seconds'
//	 And 'first' teacher enable annotation for '3' students
//	 And students who be enabled annotation will interact with together on whiteboard in '10 seconds'
//		And 'second' teacher enable annotation for 'all' students
//	 And all students will interact with together on whiteboard in '5 seconds'
//	 And 'first' teacher disable annotation for 'all' students
//	 And 'first' teacher stop recording live room successfully
//		Then 'first' teacher end live room
func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) Execute(ctx context.Context) (err error) {
	// fetch room states
	scn.RepeatFetchRoomStateOfAllAttendeesWithInterval(5 * time.Second)

	fmt.Println("All attendees joined successfully and begin running scenario:", scn.id)
	//  And 'first' teacher request recording live room successfully
	if err = scn.teacherRequestRecordingVideo(scn.teacherSuites[0]); err != nil {
		return fmt.Errorf("and 'first' teacher request recording live room successfully: %w", err)
	}

	// And 'first' teacher share a pdf in '9 seconds'
	if err = scn.teacherSharePDF(scn.teacherSuites[0]); err != nil {
		return fmt.Errorf("and 'first' teacher share a pdf in '9 seconds': %w", err)
	}
	scn.FetchRoomStateOfAllAttendees()
	time.Sleep(9 * time.Second)

	// And there are '7' students raise hand but after that have '2' students hand off
	g := new(errgroup.Group)
	for i := 0; i < 7; i++ {
		suite := scn.studentSuites[i]
		g.Go(func() error {
			if err = scn.studentRaiseHand(suite); err != nil {
				return fmt.Errorf("and there are '7' students raise hand: %w", err)
			}
			return nil
		})
	}

	// And there is one student spam raise hand button '10 time' in '2 seconds'
	g.Go(func() error {
		if err = scn.lastStudentSpamRaiseHand(10, 200*time.Millisecond); err != nil {
			return fmt.Errorf("and there is one student spam raise hand button '10 time' in '2 seconds': %w", err)
		}
		return nil
	})
	if err = g.Wait(); err != nil {
		return err
	}

	// but after that have '2' students hand off
	g = new(errgroup.Group)
	for i := 0; i < 2; i++ {
		suite := scn.studentSuites[i]
		g.Go(func() error {
			if err = scn.studentHandOff(suite); err != nil {
				return fmt.Errorf("after that have '2' students hand off: %w", err)
			}
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		return err
	}

	// And 'second' teacher fold hand of all students
	if err = scn.handOfAllStudents(scn.teacherSuites[1]); err != nil {
		return fmt.Errorf("and 'second' teacher fold hand of all students': %w", err)
	}

	// And waiting the student who be chosen answer question in '5 seconds'
	time.Sleep(5 * time.Second)

	// And 'second' teacher stop sharing and share a video in '10 seconds'
	if err = scn.teacherStopSharing(scn.teacherSuites[1]); err != nil {
		return fmt.Errorf("and 'second' teacher stop sharing': %w", err)
	}
	scn.FetchRoomStateOfAllAttendees()

	if err = scn.teacherShareVideo(scn.teacherSuites[1]); err != nil {
		return fmt.Errorf("and 'second' teacher stop sharing and share a video in '10 seconds'': %w", err)
	}
	scn.FetchRoomStateOfAllAttendees()
	time.Sleep(10 * time.Second)

	// And 'second' teacher pause video in '5 second'
	if err = scn.teacherPauseVideo(scn.teacherSuites[1]); err != nil {
		return fmt.Errorf("and 'second' teacher pause video in '5 second'': %w", err)
	}
	scn.FetchRoomStateOfAllAttendees()
	time.Sleep(5 * time.Second)

	// And 'first' teacher open a polling with '5' options
	if err = scn.teacherOpenAPolling(scn.teacherSuites[0], 5); err != nil {
		return fmt.Errorf("and 'first' teacher open a polling with '5' options': %w", err)
	}
	scn.FetchRoomStateOfAllAttendees()

	// And waiting all student answer question in '10 seconds'
	g = new(errgroup.Group)
	for i := range scn.studentSuites {
		suite := scn.studentSuites[i]
		g.Go(func() error {
			if err = scn.studentSubmitPollingAnswer(suite, "A"); err != nil {
				return fmt.Errorf("and waiting all student answer question in '10 seconds': %w", err)
			}
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		return err
	}
	time.Sleep(10 * time.Second)

	// And 'first' teacher stop the polling and check answer of students
	if err = scn.teacherStopPolling(scn.teacherSuites[0]); err != nil {
		return fmt.Errorf("and 'first' teacher stop the polling and check answer of students: %w", err)
	}
	scn.FetchRoomStateOfAllAttendees()

	// And waiting for explaining answer in '6 seconds'
	time.Sleep(6 * time.Second)

	// And 'first' teacher end the polling
	if err = scn.teacherEndPolling(scn.teacherSuites[0]); err != nil {
		return fmt.Errorf("and 'first' teacher end the polling: %w", err)
	}
	scn.FetchRoomStateOfAllAttendees()

	// And 'second' teacher resume video and let it play in '5 second'
	if err = scn.teacherResumeVideo(scn.teacherSuites[1]); err != nil {
		return fmt.Errorf("and 'second' teacher resume video and let it play in '5 second'': %w", err)
	}
	scn.FetchRoomStateOfAllAttendees()
	time.Sleep(5 * time.Second)

	// And 'second' teacher 'stop' video
	if err = scn.teacherStopVideo(scn.teacherSuites[1]); err != nil {
		return fmt.Errorf("and 'second' teacher 'stop' video': %w", err)
	}
	scn.FetchRoomStateOfAllAttendees()

	// And 'second' teacher share a pdf in '5 seconds'
	if err = scn.teacherSharePDF(scn.teacherSuites[1]); err != nil {
		return fmt.Errorf("and 'second' teacher share a pdf in '5 seconds': %w", err)
	}
	time.Sleep(5 * time.Second)
	scn.FetchRoomStateOfAllAttendees()

	// And 'first' teacher enable annotation for '3' students
	studentIDs := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		studentIDs = append(studentIDs, scn.studentAccounts[i].ID)
	}
	if err = scn.teacherEnableAnnotation(scn.teacherSuites[0], studentIDs); err != nil {
		return fmt.Errorf("and 'first' teacher enable annotation for '3' students: %w", err)
	}

	// And students who be enabled annotation will interact with together on whiteboard in '10 seconds'
	time.Sleep(10 * time.Second)

	// And 'second' teacher enable annotation for 'all' students
	studentIDs = make([]string, 0, len(scn.studentAccounts))
	for _, acc := range scn.studentAccounts {
		studentIDs = append(studentIDs, acc.ID)
	}
	if err = scn.teacherEnableAnnotation(scn.teacherSuites[1], studentIDs); err != nil {
		return fmt.Errorf("and 'second' teacher enable annotation for 'all' students: %w", err)
	}

	// And all students will interact with together on whiteboard in '5 seconds'
	time.Sleep(5 * time.Second)

	// And 'first' teacher disable annotation for 'all' students
	if err = scn.teacherDisableAnnotation(scn.teacherSuites[0], studentIDs); err != nil {
		return fmt.Errorf("and 'first' teacher disable annotation for 'all' students: %w", err)
	}

	// And 'first' teacher stop recording live room successfully
	if err = scn.teacherStopRecordingVideo(scn.teacherSuites[0]); err != nil {
		return fmt.Errorf("'first' teacher stop recording live room successfully: %w", err)
	}

	// Then 'first' teacher end live room
	if err = scn.teacherEndLiveRoom(scn.teacherSuites[0]); err != nil {
		return fmt.Errorf("then 'first' teacher end live room: %w", err)
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) teacherSharePDF(suite *Suite) error {
	ctx, err := suite.lessonSuite.
		UserShareAMaterialWithTypeIsPdfInLiveLessonRoom(ContextForSuite(suite))
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) teacherShareVideo(suite *Suite) error {
	ctx, err := suite.lessonSuite.
		UserShareAMaterialWithTypeIsVideoInLiveLessonRoom(ContextForSuite(suite))
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) teacherStopSharing(suite *Suite) error {
	ctx, err := suite.lessonSuite.
		UserStopSharingMaterialInLiveLessonRoom(ContextForSuite(suite))
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) teacherPauseVideo(suite *Suite) error {
	ctx, err := suite.lessonSuite.
		UserPauseVideoInLiveLessonRoom(ContextForSuite(suite))
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) teacherResumeVideo(suite *Suite) error {
	ctx, err := suite.lessonSuite.
		UserResumeVideoInLiveLessonRoom(ContextForSuite(suite))
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) teacherStopVideo(suite *Suite) error {
	ctx, err := suite.lessonSuite.
		UserStopVideoInLiveLessonRoom(ContextForSuite(suite))
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) teacherOpenAPolling(suite *Suite, opts int) error {
	ctx, err := suite.lessonSuite.
		UserStartPollingInLiveLessonRoomWithNumOption(ContextForSuite(suite), opts)
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) teacherEnableAnnotation(suite *Suite, studentIDs []string) error {
	suite.lessonSuite.CommonSuite.StepState.StudentIds = studentIDs
	ctx, err := suite.lessonSuite.
		UserEnableAnnotationInLiveLessonRoom(ContextForSuite(suite))
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) teacherDisableAnnotation(suite *Suite, studentIDs []string) error {
	suite.lessonSuite.CommonSuite.StepState.StudentIds = studentIDs
	ctx, err := suite.lessonSuite.
		UserDisableAnnotationInLiveLessonRoom(ContextForSuite(suite))
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) studentSubmitPollingAnswer(suite *Suite, answer string) error {
	ctx, err := suite.lessonSuite.
		UserSubmitPollingAnswerInLiveLessonRoom(ContextForSuite(suite), answer)
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) teacherStopPolling(suite *Suite) error {
	ctx, err := suite.lessonSuite.
		UserStopPollingInLiveLessonRoom(ContextForSuite(suite))
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) teacherEndPolling(suite *Suite) error {
	ctx, err := suite.lessonSuite.
		UserEndPollingInLiveLessonRoom(ContextForSuite(suite))
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) studentRaiseHand(suite *Suite) error {
	ctx, err := suite.lessonSuite.UserRaiseHandInLiveLessonRoom(ContextForSuite(suite))
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) studentHandOff(suite *Suite) error {
	ctx, err := suite.lessonSuite.UserHandOffInLiveLessonRoom(ContextForSuite(suite))
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) handOfAllStudents(suite *Suite) error {
	ctx, err := suite.lessonSuite.
		UserFoldHandAllLearner(ContextForSuite(suite))
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) teacherRequestRecordingVideo(suite *Suite) error {
	ctx, err := suite.lessonSuite.
		UserRequestRecordingLiveLesson(ContextForSuite(suite))
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) teacherStopRecordingVideo(suite *Suite) error {
	ctx, err := suite.lessonSuite.
		UserStopRecordingLiveLesson(ContextForSuite(suite))
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) teacherEndLiveRoom(suite *Suite) error {
	ctx, err := suite.lessonSuite.
		EndOneOfTheLiveLessonV1(ContextForSuite(suite))
	if err == nil {
		err = HaveNoResponseError(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) lastStudentSpamRaiseHand(num int, interval time.Duration) error {
	for i := 0; i < num; i++ {
		suite := scn.studentSuites[len(scn.studentSuites)-1]
		if err := scn.studentRaiseHand(suite); err != nil {
			return err
		}
		time.Sleep(interval)
	}

	return nil
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) RepeatFetchRoomStateOfAllAttendeesWithInterval(interval time.Duration) {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	n := r.Intn(10) // because time joined not same between attendees

	attendees := scn.teacherSuites
	attendees = append(attendees, scn.studentSuites...)
	for i := range attendees {
		attendee := attendees[i]
		time.AfterFunc(time.Duration(i*n)*time.Millisecond, func() {
			for {
				_, _ = scn.FetchRoomState(attendee)
				time.Sleep(interval)
			}
		})
	}
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) FetchRoomStateOfAllAttendees() {
	attendees := scn.teacherSuites
	attendees = append(attendees, scn.studentSuites...)
	for i := range attendees {
		attendee := attendees[i]
		go func() {
			_, _ = scn.FetchRoomState(attendee)
		}()
	}
}

func (scn *ScenarioSimulateAStudyOnLiveLessonRoom) FetchRoomState(suite *Suite) (*pb.LiveLessonStateResponse, error) {
	res, err := suite.lessonSuite.GetCurrentStateOfLiveLessonRoom(ContextForSuite(suite), suite.lessonSuite.CommonSuite.CurrentLessonID)
	if err != nil {
		atomic.AddInt32(&scn.logs.NumberFetchFailedRoomState, 1)
		return nil, err
	}

	return res, nil
}

// AfterLoginAndEnterALiveLessonRoom_Teacher simulate
// get teacher profile
// RetrieveLocations
// list all courses
func (s *Suite) AfterLoginAndEnterALiveLessonRoom_Teacher(ctx context.Context) error {
	ctx = common.StepStateToContext(ctx, s.lessonSuite.CommonSuite.StepState)
	stepState := common.StepStateFromContext(ctx)

	if _, err := s.lessonSuite.CommonSuite.GetTeacherProfile(ctx); err != nil {
		return fmt.Errorf("GetTeacherProfile: %w", err)
	}
	if err := HaveNoResponseError(ctx); err != nil {
		return fmt.Errorf("GetTeacherProfile: %w", err)
	}

	if _, err := s.lessonSuite.RetrieveLocations(ctx); err != nil {
		return fmt.Errorf("RetrieveLocations: %w", err)
	}
	if err := HaveNoResponseError(ctx); err != nil {
		return fmt.Errorf("RetrieveLocations: %w", err)
	}

	if _, err := s.lessonSuite.CommonSuite.ListCourses(ctx, 100); err != nil {
		return fmt.Errorf("ListCourses: %w", err)
	}
	if err := HaveNoResponseError(ctx); err != nil {
		return fmt.Errorf("ListCourses: %w", err)
	}
	// check do list course contain current course id ?
	isExistCurrentCourse := false
	for _, courseIf := range stepState.Courses {
		course := courseIf.(*cpb_v1.Course)
		if course.Info.Id == stepState.CurrentCourseID {
			isExistCurrentCourse = true
			break
		}
	}
	// TODO: check later
	_ = isExistCurrentCourse

	if _, err := s.lessonSuite.CommonSuite.RetrieveLiveLessonByCourseWithStartTimeAndEndTime(ctx, stepState.CurrentCourseID, "", ""); err != nil {
		return fmt.Errorf("RetrieveLiveLessonByCourseWithStartTimeAndEndTime: %w", err)
	}
	if err := HaveNoResponseError(ctx); err != nil {
		return fmt.Errorf("RetrieveLiveLessonByCourseWithStartTimeAndEndTime: %w", err)
	}
	// check do list lesson contain current lesson id ?
	isExistCurrentLesson := false
	for _, lesson := range stepState.Response.(*bpb.RetrieveLiveLessonResponse).Lessons {
		if lesson.LessonId == stepState.CurrentLessonID {
			isExistCurrentLesson = true
			break
		}
	}
	if !isExistCurrentLesson {
		return fmt.Errorf("lesson id %s is not exist in list lessons", stepState.CurrentLessonID)
	}

	if _, err := s.lessonSuite.CommonSuite.ListStudentsInLesson(ctx); err != nil {
		return fmt.Errorf("ListStudentsInLesson: %w", err)
	}
	if err := HaveNoResponseError(ctx); err != nil {
		return fmt.Errorf("ListStudentsInLesson: %w", err)
	}

	if _, err := s.lessonSuite.CommonSuite.GetLessonMedias(ctx); err != nil {
		return fmt.Errorf("GetLessonMedias: %w", err)
	}
	if err := HaveNoResponseError(ctx); err != nil {
		return fmt.Errorf("GetLessonMedias: %w", err)
	}

	if _, err := s.lessonSuite.CommonSuite.JoinLesson(ctx); err != nil {
		return fmt.Errorf("JoinLesson: %w", err)
	}
	if err := HaveNoResponseError(ctx); err != nil {
		return fmt.Errorf("JoinLesson: %w", err)
	}

	if _, err := s.lessonSuite.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID); err != nil {
		return fmt.Errorf("GetCurrentStateOfLiveLessonRoom: %w", err)
	}
	if err := HaveNoResponseError(ctx); err != nil {
		return fmt.Errorf("GetCurrentStateOfLiveLessonRoom: %w", err)
	}

	if _, err := s.RefreshLiveLessonSession(ctx); err != nil {
		return fmt.Errorf("RefreshLiveLessonSession: %w", err)
	}
	if err := HaveNoResponseError(ctx); err != nil {
		return fmt.Errorf("RefreshLiveLessonSession: %w", err)
	}

	return nil
}

// AfterLoginAndEnterALiveLessonRoom_Student simulate
// get student profile
func (s *Suite) AfterLoginAndEnterALiveLessonRoom_Student(ctx context.Context) error {
	ctx = common.StepStateToContext(ctx, s.lessonSuite.CommonSuite.StepState)
	stepState := common.StepStateFromContext(ctx)

	if _, err := s.lessonSuite.CommonSuite.GetStudentProfileV1(ctx); err != nil {
		return fmt.Errorf("GetStudentProfileV1: %w", err)
	}
	if err := HaveNoResponseError(ctx); err != nil {
		return fmt.Errorf("GetStudentProfileV1: %w", err)
	}

	if _, err := s.lessonSuite.CommonSuite.RetrieveLiveLessonByCourseWithStartTimeAndEndTime(ctx, stepState.CurrentCourseID, "", ""); err != nil {
		return fmt.Errorf("RetrieveLiveLessonByCourseWithStartTimeAndEndTime: %w", err)
	}
	if err := HaveNoResponseError(ctx); err != nil {
		return fmt.Errorf("RetrieveLiveLessonByCourseWithStartTimeAndEndTime: %w", err)
	}
	// check do list lesson contain current lesson id ?
	isExistCurrentLesson := false
	for _, lesson := range stepState.Response.(*bpb.RetrieveLiveLessonResponse).Lessons {
		if lesson.LessonId == stepState.CurrentLessonID {
			isExistCurrentLesson = true
			break
		}
	}
	if !isExistCurrentLesson {
		return fmt.Errorf("lesson id %s is not exist in list lessons", stepState.CurrentLessonID)
	}

	if _, err := s.lessonSuite.CommonSuite.JoinLesson(ctx); err != nil {
		return fmt.Errorf("JoinLesson: %w", err)
	}
	if err := HaveNoResponseError(ctx); err != nil {
		return fmt.Errorf("JoinLesson: %w", err)
	}

	if _, err := s.lessonSuite.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID); err != nil {
		return fmt.Errorf("GetCurrentStateOfLiveLessonRoom: %w", err)
	}
	if err := HaveNoResponseError(ctx); err != nil {
		return fmt.Errorf("GetCurrentStateOfLiveLessonRoom: %w", err)
	}

	if _, err := s.RefreshLiveLessonSession(ctx); err != nil {
		return fmt.Errorf("RefreshLiveLessonSession: %w", err)
	}
	if err := HaveNoResponseError(ctx); err != nil {
		return fmt.Errorf("RefreshLiveLessonSession: %w", err)
	}

	return nil
}

func (s *Suite) RefreshLiveLessonSession(ctx context.Context) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)
	currentTeacher := stepState.CurrentTeacherID
	ctx, err := s.makeUsersSubscribeV2Ctx(ctx, []string{currentTeacher}, []string{stepState.AuthToken})
	if err != nil {
		return ctx, err
	}
	_, err = tpb_v1.NewLessonChatReaderServiceClient(s.lessonSuite.BobConn).
		RefreshLiveLessonSession(helper.GRPCContext(ctx, "token", stepState.AuthToken), &tpb_v1.RefreshLiveLessonSessionRequest{LessonId: stepState.CurrentLessonID})
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (s *Suite) makeUsersSubscribeV2Ctx(ctx context.Context, userIDs []string, tokens []string) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)
	for idx, id := range userIDs {
		token := tokens[idx]
		ctx2, cancel := context.WithCancel(ctx)
		subClient, err := tpb.NewChatServiceClient(s.lessonSuite.BobConn).
			SubscribeV2(helper.GRPCContext(ctx2, "token", token), &tpb.SubscribeV2Request{})
		if err != nil {
			cancel()
			return ctx, err
		}

		// if old stream already exist, cancel it and overwrite
		if oldStream, ok := stepState.SubV2Clients[id]; ok {
			oldStream.Cancel()
		}
		stepState.SubV2Clients[id] = common.CancellableStream{ChatService_SubscribeV2Client: subClient, Cancel: cancel}
	}
	return ctx, nil
}
