package commands

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/constants"
	infra_scheduler "github.com/manabie-com/backend/internal/calendar/infrastructure"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/producers"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	infra_lesson_report "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	user_infras "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"
	zoom_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"
	zoom_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/infrastructure"
	zoom_service "github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/service"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LessonCommandHandler struct {
	WrapperConnection *support.WrapperDBConnection
	JSM               nats.JetStreamManagement
	Env               string
	UnleashClientIns  unleashclient.ClientInstance

	// ports
	LessonRepo                   infrastructure.LessonRepo
	LessonReportRepo             infra_lesson_report.LessonReportRepo
	SchedulerRepo                infra_scheduler.SchedulerPort
	StudentSubscriptionRepo      user_infras.StudentSubscriptionRepo
	LessonProducer               producers.LessonProducer
	ClassroomRepo                infrastructure.ClassroomRepo
	ReallocationRepo             infrastructure.ReallocationRepo
	LessonMemberRepo             infrastructure.LessonMemberRepo
	UserAccessPathRepo           infrastructure.UserAccessPathPort
	StudentEnrollmentHistoryRepo infrastructure.StudentEnrollmentStatusHistoryPort
	ZoomService                  zoom_service.ZoomServiceInterface
	ZoomAccountRepo              zoom_repo.ZoomAccountRepo
	MasterDataPort               infrastructure.MasterDataPort
	DateInfoRepo                 infrastructure.DateInfoRepo
	UserModulePort               infrastructure.UserModulePort
	SchedulerClient              clients.SchedulerClientInterface
	LessonPublisher              infrastructure.LessonPublisher
}

func canGenerateZoomLink(startTime *time.Time, endTime *time.Time) bool {
	durationRecurrence := zoom_domain.GetDurationByWeek(*startTime, *endTime)

	if durationRecurrence <= zoom_domain.MaximumRepeatInterval {
		return false
	}
	startTimeGenerate := startTime.AddDate(0, 0, 1*int(zoom_domain.MaximumRepeatInterval))

	if timeutil.EqualDate(startTimeGenerate, *endTime) {
		return false
	}
	return startTimeGenerate.Before(*endTime)
}

/*
because a link recurring will have data
so we have multi-link and multi-occurrence for a lesson in the recurring chain, so we need to calculate zoomId and occurrence
*/
func calculateIndexZoomLinkAndIndexOccurrence(startTime *time.Time, lessonTime *time.Time, zoomLinks []*zoom_domain.GenerateZoomLinkResponse) (int, int) {
	totalLink := len(zoomLinks)
	// caculate week of lesson in chain
	durationWeek := zoom_domain.GetDurationByWeek(*startTime, *lessonTime)
	maxWeekForALink := int(zoom_domain.MaximumRepeatInterval)
	// get index of zoom
	indexZoomLink := int(durationWeek) / maxWeekForALink
	safeIndexZoomLink := support.Max(0, support.Min(indexZoomLink, totalLink-1))
	totalOccurrence := len(zoomLinks[safeIndexZoomLink].Occurrences)
	// get index of occurenceID
	indexOccurrence := durationWeek - float64(safeIndexZoomLink*maxWeekForALink)
	safeIdxOccurrence := support.Max(0, support.Min(int(indexOccurrence), totalOccurrence-1))
	return safeIndexZoomLink, safeIdxOccurrence
}

func (l *LessonCommandHandler) CreateLessonOneTime(ctx context.Context, payload CreateLesson) (*domain.Lesson, error) {
	lesson := payload.Lesson
	conn, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if err = database.ExecInTx(ctx, conn, func(ctx context.Context, tx pgx.Tx) (err error) {
		lesson.SaveOneTime()
		if err = lesson.IsValid(ctx, tx); err != nil {
			return fmt.Errorf("invalid lesson: %w", err)
		}
		isUnleashToggled, err := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_BackOffice_ReallocateStudents", l.Env)
		if err != nil {
			return fmt.Errorf("l.connectToUnleash: %w", err)
		}
		if isUnleashToggled {
			if err = l.handleStudentReallocatedDiffLocation(ctx, tx, lesson); err != nil {
				return fmt.Errorf("l.handleStudentReallocatedDiffLocation: %w", err)
			}
		}

		isUnleashTeachingTimeToggled, err := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_CourseTeachingTime", l.Env)
		if err != nil {
			return fmt.Errorf("l.connectToUnleash: %w", err)
		}
		if isUnleashTeachingTimeToggled {
			if err = l.AddLessonCourseTeachingTime(ctx, tx, []*domain.Lesson{lesson}, false, payload.TimeZone); err != nil {
				return fmt.Errorf("lesson.AddLessonCourseTeachingTime: %w", err)
			}
		}
		createSchedulerResp, err := l.SchedulerClient.CreateScheduler(ctx, clients.CreateReqCreateScheduler(lesson.StartTime, lesson.StartTime, constants.FrequencyOnce))
		if err != nil {
			logger := ctxzap.Extract(ctx)
			logger.Error(
				"Create Scheduler error",
				zap.String("lesson_id", lesson.LessonID),
				zap.Error(err),
			)

			return fmt.Errorf("create scheduler fail: %w ", err)
		}
		schedulerID := createSchedulerResp.SchedulerId
		lesson.SchedulerID = schedulerID
		lesson, err = l.LessonRepo.InsertLesson(ctx, tx, lesson)
		if err != nil {
			return fmt.Errorf("could not persist lesson: lessonRepo.InsertLesson: %w", err)
		}
		lesson.Persisted = true

		return nil
	}); err != nil {
		return nil, err
	}
	return lesson, nil
}

func (l *LessonCommandHandler) CreateLessonOneTimeForImportLesson(ctx context.Context, db database.Ext, lesson *domain.Lesson, schedulerID string) error {
	return database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) (err error) {
		if err = lesson.IsValid(ctx, tx); err != nil {
			return fmt.Errorf("invalid lesson: %w", err)
		}

		lesson.SchedulerID = schedulerID
		lesson, err = l.LessonRepo.InsertLesson(ctx, tx, lesson)
		if err != nil {
			return fmt.Errorf("could not persist lesson: lessonRepo.InsertLesson: %w", err)
		}
		lesson.Persisted = true

		return nil
	})
}

func (l *LessonCommandHandler) CreateRecurringLesson(ctx context.Context, payload CreateRecurringLesson) (*domain.RecurringLesson, error) {
	conn, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	baseLesson := payload.Lesson
	rrlue := payload.RRuleCmd

	dateInfo, err := baseLesson.GetDateInfoByDateAndCenterID(ctx, conn, rrlue.StartTime, rrlue.UntilDate, baseLesson.LocationID)
	if err != nil {
		return nil, err
	}
	if err := baseLesson.CheckClosedDate(rrlue.StartTime, dateInfo); err != nil {
		return nil, err
	}
	skipDates := baseLesson.GetNonRegularDatesExceptFirstDate(ctx, dateInfo)
	recurRule, err := domain.NewRecurrenceRule(domain.Option{
		Freq:         domain.WEEKLY,
		StartTime:    rrlue.StartTime,
		EndTime:      rrlue.EndTime,
		UntilDate:    rrlue.UntilDate,
		ExcludeDates: skipDates,
	})
	if err != nil {
		return nil, fmt.Errorf("could not init new recurrence rule: %w ", err)
	}

	recurSet := recurRule.ExceptFirst()
	freq := domain.FrequencyName[recurRule.Option.Freq]
	createSchedulerResp, err := l.SchedulerClient.CreateScheduler(ctx, clients.CreateReqCreateScheduler(rrlue.StartTime, rrlue.UntilDate, constants.Frequency(freq)))

	if err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Error(
			"Create Scheduler error",
			zap.String("lesson_id", baseLesson.LessonID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("create scheduler fail: %w ", err)
	}
	schedulerID := createSchedulerResp.SchedulerId

	linksZoomAttach := []*zoom_domain.GenerateZoomLinkResponse{}
	zoomInfo := payload.ZoomInfo
	if zoomInfo != nil {
		baseLesson.ZoomID = zoomInfo.ZoomID
		baseLesson.ZoomLink = zoomInfo.ZoomLink
		baseLesson.ZoomOwnerID = zoomInfo.ZoomAccountID
		if len(zoomInfo.ZoomOccurrences) > 0 {
			baseLesson.ZoomOccurrenceID = zoomInfo.ZoomOccurrences[0].OccurrenceID
		}
		if zoomInfo.ZoomLink != "" {
			zoomID, err := strconv.Atoi(zoomInfo.ZoomID)
			if err != nil {
				return nil, fmt.Errorf("zoomId invalid %w ", err)
			}
			linksZoomAttach = []*zoom_domain.GenerateZoomLinkResponse{
				{
					ZoomID: zoomID,
					URL:    payload.Lesson.ZoomLink,
					Occurrences: sliceutils.Map(zoomInfo.ZoomOccurrences, func(zoomOccurrence *ZoomOccurrence) *zoom_domain.OccurrenceOfZoomResponse {
						return &zoom_domain.OccurrenceOfZoomResponse{
							OccurrenceID: zoomOccurrence.OccurrenceID,
							StartTime:    zoomOccurrence.StartTime,
						}
					}),
				},
			}
			startTime := payload.RRuleCmd.StartTime
			endDate := payload.RRuleCmd.UntilDate
			// should skip the first 60 weeks because it was sent from FE
			startTimeGenerate := startTime.AddDate(0, 0, zoom_domain.DaysOfWeek*int(zoom_domain.MaximumRepeatInterval))
			endTimeGenerate := payload.RRuleCmd.EndTime.AddDate(0, 0, zoom_domain.DaysOfWeek*int(zoom_domain.MaximumRepeatInterval))
			if canGenerateZoomLink(&startTime, &endDate) {
				requestsGenerateZoomMeeting, err := zoom_domain.ConverterMultiZoomGenerateMeetingRequest(&zoom_domain.MultiZoomGenerateMeetingRequest{
					StartTime:   startTimeGenerate,
					EndTime:     endTimeGenerate,
					TimeZone:    payload.TimeZone,
					EndDateTime: payload.RRuleCmd.UntilDate,
				})
				if err != nil {
					return nil, fmt.Errorf("could not convert zoom meeting request: %w ", err)
				}
				zoomAccount, err := l.ZoomAccountRepo.GetZoomAccountByID(ctx, conn, payload.ZoomInfo.ZoomAccountID)
				if err != nil {
					return nil, fmt.Errorf("could not found zoom account id: %w ", err)
				}
				links, err := l.ZoomService.GenerateMultiZoomLink(ctx, zoomAccount.Email, requestsGenerateZoomMeeting)
				if err != nil {
					return nil, fmt.Errorf("could not generate zoom meeting request: %w ", err)
				}
				linksZoomAttach = append(linksZoomAttach, links...)
			}
		}
	}

	lessonRecurring := &domain.RecurringLesson{
		ID: schedulerID,
	}
	baseLesson.SchedulerID = schedulerID
	learner := baseLesson.Learners
	studentIDWithCourseID := make([]string, 0, len(learner)*2)
	for _, l := range learner {
		studentIDWithCourseID = append(studentIDWithCourseID, l.LearnerID, l.CourseID)
	}
	var studentSubscriptions user_domain.StudentSubscriptions
	if len(studentIDWithCourseID) > 0 {
		var locationID []string
		if len(baseLesson.LocationID) > 0 {
			locationID = append(locationID, baseLesson.LocationID)
		}
		studentSubscriptions, err = l.StudentSubscriptionRepo.GetStudentCourseSubscriptions(ctx, conn, locationID, studentIDWithCourseID...)
		if err != nil {
			return nil, fmt.Errorf("l.StudentSubscriptionRepo.GetStudentCourseSubscriptions: %w", err)
		}
	}
	studentSubMap := make(map[string]time.Time)
	for _, ss := range studentSubscriptions {
		studentSubMap[ss.StudentWithCourseID()] = ss.EndAt
	}
	lessons := []*domain.Lesson{baseLesson}
	loc := timeutil.Location(payload.TimeZone)
	totalLinkZoom := len(linksZoomAttach)
	for _, r := range recurSet {
		lesson := *baseLesson
		lesson.ResetID()
		lesson.StartTime = r.StartTime
		lesson.EndTime = r.EndTime
		lesson.SchedulerID = schedulerID
		lesson.Material = &domain.LessonMaterial{}
		learners := domain.LessonLearners{}
		for _, learner := range lesson.Learners {
			if learner.IsReallocated() {
				continue
			}
			ln := *learner
			endAt, ok := studentSubMap[ln.StudentWithCourse()]
			if !ok {
				return nil, fmt.Errorf("cannot find student course duration(student_id=%s,course_id=%s)", ln.LearnerID, ln.CourseID)
			}
			if r.StartTime.In(loc).Format(domain.Ymd) <= endAt.In(loc).Format(domain.Ymd) {
				ln.EmptyAttendanceInfo()
				learners = append(learners, &ln)
			}
		}
		lesson.Learners = learners
		if zoomInfo != nil {
			if totalLinkZoom > 0 {
				indexLink, indexOccurrence := calculateIndexZoomLinkAndIndexOccurrence(&rrlue.StartTime, &lesson.StartTime, linksZoomAttach)
				zoomLink := linksZoomAttach[indexLink]
				lesson.ZoomID = fmt.Sprint(zoomLink.ZoomID)
				lesson.ZoomLink = zoomLink.URL
				if len(zoomLink.Occurrences) > 0 {
					lesson.ZoomOccurrenceID = zoomLink.Occurrences[indexOccurrence].OccurrenceID
				}
			}
			lesson.ZoomOwnerID = payload.ZoomInfo.ZoomAccountID
		}
		lessons = append(lessons, &lesson)
	}
	lessonRecurring.Lessons = lessons
	err = database.ExecInTx(ctx, conn, func(ctx context.Context, tx pgx.Tx) (err error) {
		lessonRecurring.Save()
		if err = lessonRecurring.IsValid(ctx, tx); err != nil {
			return fmt.Errorf("invalid lesson: %w", err)
		}

		isUnleashToggled, err := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_BackOffice_ReallocateStudents", l.Env)
		if err != nil {
			return fmt.Errorf("l.connectToUnleash: %w", err)
		}
		if isUnleashToggled {
			if err = l.handleStudentReallocatedDiffLocation(ctx, tx, lessonRecurring.GetBaseLesson()); err != nil {
				return fmt.Errorf("l.handleStudentReallocatedDiffLocation: %w", err)
			}
		}
		isUnleashTeachingTimeToggled, err := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_CourseTeachingTime", l.Env)
		if err != nil {
			return fmt.Errorf("l.connectToUnleash: %w", err)
		}
		if isUnleashTeachingTimeToggled {
			if err = l.AddLessonCourseTeachingTime(ctx, tx, lessonRecurring.Lessons, false, payload.TimeZone); err != nil {
				return fmt.Errorf("lesson.AddLessonCourseTeachingTime: %w", err)
			}
		}
		_, err = l.LessonRepo.UpsertLessons(ctx, tx, lessonRecurring)
		if err != nil {
			return fmt.Errorf("could not persist lesson: lessonRepo.UpsertLessons: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return lessonRecurring, nil
}

func (l *LessonCommandHandler) AddLessonCourseTeachingTime(ctx context.Context, db database.Ext, lessons []*domain.Lesson, isReCompute bool, timezone string) error {
	courseIDs := []string{}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone")
	}
	now := time.Now().In(location)

	for _, lesson := range lessons {
		startTime := lesson.StartTime.In(location)
		if isReCompute && startTime.Before(now) {
			continue
		}
		if lesson.CourseID != "" {
			courseIDs = append(courseIDs, lesson.CourseID)
		}

		if len(lesson.Learners) > 0 {
			courseIDs = append(courseIDs, lesson.Learners.GetCourseIDsOfLessonLearners()...)
		}
	}
	courseIDs = golibs.Uniq(courseIDs)
	if len(courseIDs) == 0 {
		return nil
	}

	coursesMap, err := l.MasterDataPort.GetCourseTeachingTimeByIDs(ctx, db, courseIDs)
	if err != nil {
		return fmt.Errorf("l.MasterDataPort.GetCourseTeachingTimeByIDs: %w", err)
	}

	for _, lesson := range lessons {
		course, ok := coursesMap[lesson.CourseID]
		if lesson.TeachingMethod == domain.LessonTeachingMethodGroup {
			if !ok { // reset to NULL value incase recompute deleted course time
				lesson.BreakTime = -1
				lesson.PreparationTime = -1
				continue
			}
			lesson.BreakTime = course.BreakTime
			lesson.PreparationTime = course.PreparationTime
			continue
		}

		maxBreakTime := int32(0)
		lessonCourseIDs := lesson.Learners.GetCourseIDsOfLessonLearners()
		courseExist := false
		for _, lcID := range lessonCourseIDs {
			cEntity, exist := coursesMap[lcID]
			if !exist {
				continue
			}
			courseExist = true
			if cEntity.BreakTime > maxBreakTime {
				maxBreakTime = cEntity.BreakTime
				lesson.PreparationTime = cEntity.PreparationTime
				lesson.BreakTime = cEntity.BreakTime
			}
			if cEntity.BreakTime == maxBreakTime && cEntity.PreparationTime > lesson.PreparationTime {
				lesson.PreparationTime = cEntity.PreparationTime
				lesson.BreakTime = cEntity.BreakTime
			}
		}
		if !courseExist { // reset to NULL value in case recompute deleted course time
			lesson.PreparationTime = -1
			lesson.BreakTime = -1
		}
	}

	return nil
}

func (l *LessonCommandHandler) handleStudentReallocatedDiffLocation(ctx context.Context, tx pgx.Tx, lesson *domain.Lesson) error {
	userReallocated := lesson.Learners.GetStudentReallocatedDiffLocation(lesson.LocationID)
	if len(userReallocated) > 0 {
		locationMap, err := l.UserAccessPathRepo.GetLocationAssignedByUserID(ctx, tx, userReallocated)
		if err != nil {
			return fmt.Errorf("l.UserAccessPathRepo.GetLocationAssignedByUserID: %w", err)
		}
		accessPathData := map[string]string{}
		for userID, locationIDs := range locationMap {
			if !slices.Contains(locationIDs, lesson.LocationID) {
				accessPathData[userID] = lesson.LocationID
			}
		}
		if len(accessPathData) > 0 {
			isEnable, err := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_BackOffice_TemporaryLocationAssignment", l.Env)
			if err != nil {
				isEnable = false
			}
			if isEnable {
				studentEnrollmentStatus := make([]*npb.LessonReallocateStudentEnrollmentStatusEvent_StudentEnrollmentStatusInfo, 0, len(accessPathData))
				for userID, locationID := range accessPathData {
					item := &npb.LessonReallocateStudentEnrollmentStatusEvent_StudentEnrollmentStatusInfo{
						StudentId:        userID,
						LocationId:       locationID,
						StartDate:        timestamppb.Now(),
						EndDate:          timestamppb.New(time.Now().AddDate(0, 0, 14)),
						EnrollmentStatus: npb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY,
					}
					temporaryLocationAssignment := domain.TemporaryLocationAssignment{
						LessonReallocateStudentEnrollmentStatusEvent_StudentEnrollmentStatusInfo: item,
					}
					if err := temporaryLocationAssignment.Validate(); err != nil {
						return err
					}
					studentEnrollmentStatus = append(studentEnrollmentStatus, item)
				}
				if err = l.LessonPublisher.PublishTemporaryLocationAssignment(ctx, l.JSM, &npb.LessonReallocateStudentEnrollmentStatusEvent{
					StudentEnrollmentStatus: studentEnrollmentStatus,
				}); err != nil {
					return fmt.Errorf("l.LessonPublisher.PublishTemporaryLocationAssignment: %w", err)
				}
			} else {
				organization, err := interceptors.OrganizationFromContext(ctx)
				if err != nil {
					return err
				}
				userAccessPaths := []*user_domain.UserAccessPath{}

				for userID, locationID := range accessPathData {
					userAccessPaths = append(userAccessPaths, &user_domain.UserAccessPath{
						UserID:     userID,
						LocationID: locationID,
					})
					temporaryLocationAssignment := domain.NewTemporaryLocationAssignment(
						domain.TemporaryLocationAssignmentAttribute{
							StudentID:      userID,
							LocationID:     locationID,
							StartDate:      time.Now(),
							EndDate:        lesson.StartTime.AddDate(0, 0, 14),
							OrganizationID: organization.OrganizationID().String(),
						},
					)

					if err = l.StudentEnrollmentHistoryRepo.Create(ctx, tx, temporaryLocationAssignment); err != nil {
						return fmt.Errorf("l.StudentEnrollmentHistoryRepo.Create : %w", err)
					}
				}

				if err = l.UserAccessPathRepo.Create(ctx, tx, userAccessPaths); err != nil {
					return fmt.Errorf("l.UserAccessPathRepo.Create : %w", err)
				}
			}
		}
	}
	return nil
}

func (l *LessonCommandHandler) ImportLesson(ctx context.Context, req *lpb.ImportLessonRequest) ([]*domain.Lesson, []*lpb.ImportLessonResponse_ImportLessonError, error) {
	var (
		errorCSVs []*lpb.ImportLessonResponse_ImportLessonError
		tz        = "UTC"
	)
	if req.GetTimeZone() != "" {
		tz = req.GetTimeZone()
	}

	isUsingVersion2, err := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_BackOffice_ImportLessonByCSVV2", l.Env)
	if err != nil {
		return nil, nil, fmt.Errorf("l.connectToUnleash: %w", err)
	}

	lessonPayloads, errorCSVs := l.buildImportLessonArgs(ctx, req.Payload, tz, isUsingVersion2)
	if len(errorCSVs) > 0 {
		return nil, errorCSVs, err
	}

	lessons := []*domain.Lesson{}
	for _, lessonPayload := range lessonPayloads.Payloads {
		switch lessonPayload.SavingMethod {
		case lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME:
			payload := CreateLesson{Lesson: lessonPayload.Lesson, TimeZone: tz}
			lesson, err := l.CreateLessonOneTime(ctx, payload)
			if err != nil {
				return nil, errorCSVs, status.Errorf(codes.Internal, err.Error())
			}
			lessons = append(lessons, lesson)
		case lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE:
			ruleCmd := CreateRecurringLesson{
				Lesson: lessonPayload.Lesson,
				RRuleCmd: RecurrenceRuleCommand{
					StartTime: lessonPayload.StartTime,
					EndTime:   lessonPayload.EndTime,
					UntilDate: lessonPayload.UntilDate,
				},
				TimeZone: tz,
			}
			recurLesson, err := l.CreateRecurringLesson(ctx, ruleCmd)
			if err != nil {
				return nil, errorCSVs, status.Error(codes.Internal, err.Error())
			}
			lessons = recurLesson.Lessons
		default:
			return nil, errorCSVs, status.Error(codes.Internal, fmt.Sprintf(`unexpected saving option method %T`, lessonPayload.SavingMethod))
		}
	}
	return lessons, errorCSVs, err
}

func (l *LessonCommandHandler) ImportLessonV2(ctx context.Context, req *lpb.ImportLessonRequest) ([]*domain.Lesson, []*lpb.ImportLessonResponse_ImportLessonError, error) {
	var (
		errorCSVs []*lpb.ImportLessonResponse_ImportLessonError
		tz        = "UTC"
	)
	if req.GetTimeZone() != "" {
		tz = req.GetTimeZone()
	}

	isUsingVersion2, err := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_BackOffice_ImportLessonByCSVV2", l.Env)
	if err != nil {
		return nil, nil, fmt.Errorf("l.connectToUnleash: %w", err)
	}

	isUnleashTeachingTimeToggled, err := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_CourseTeachingTime", l.Env)
	if err != nil {
		return nil, nil, fmt.Errorf("l.connectToUnleash: %w", err)
	}

	conn, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, nil, err
	}

	lessonPayloads, errorCSVs := l.buildImportLessonArgs(ctx, req.Payload, tz, isUsingVersion2)
	if len(errorCSVs) > 0 {
		return nil, errorCSVs, err
	}

	createManySchedulersWithIdentityReq := make([]*cpb.CreateSchedulerWithIdentityRequest, 0)

	lessons := sliceutils.Map(lessonPayloads.Payloads, func(i ImportLessonCommand) *domain.Lesson {
		lesson := i.Lesson
		lesson.SaveOneTime()
		createManySchedulersWithIdentityReq = append(createManySchedulersWithIdentityReq, &cpb.CreateSchedulerWithIdentityRequest{
			Identity: lesson.LessonID,
			Request:  clients.CreateReqCreateScheduler(lesson.StartTime, lesson.StartTime, constants.FrequencyOnce),
		})
		return lesson
	})

	if isUnleashTeachingTimeToggled {
		if err = l.AddLessonCourseTeachingTime(ctx, conn, lessons, false, tz); err != nil {
			return nil, nil, fmt.Errorf("AddLessonCourseTeachingTime: %w", err)
		}
	}

	createSchedulersResp, err := l.SchedulerClient.CreateManySchedulers(ctx, &cpb.CreateManySchedulersRequest{
		Schedulers: createManySchedulersWithIdentityReq,
	})

	if err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Error(
			"Create Scheduler error",
			zap.Error(err),
		)
		return nil, nil, err
	}

	mapSchedulers := createSchedulersResp.MapSchedulers

	for _, lesson := range lessons {
		err = l.CreateLessonOneTimeForImportLesson(ctx, conn, lesson, mapSchedulers[lesson.LessonID])
		if err != nil {
			return nil, errorCSVs, status.Errorf(codes.Internal, err.Error())
		}
	}
	return lessons, errorCSVs, err
}

func (l *LessonCommandHandler) buildImportLessonArgs(ctx context.Context, data []byte, tz string, isUsingVersion2 bool) (payloads *ImportLessonPayload, errorCSVs []*lpb.ImportLessonResponse_ImportLessonError) {
	sc1 := scanner.NewCSVScanner(bytes.NewReader(data))
	columnsIndex := map[string]int{
		"partner_internal_id": 0,
		"start_date_time":     1,
		"end_date_time":       2,
		"teaching_method":     3,
	}

	if isUsingVersion2 {
		columnsIndexV2 := map[string]int{
			"teaching_medium":    4,
			"teacher_ids":        5,
			"student_course_ids": 6,
		}
		maps.Copy(columnsIndex, columnsIndexV2)
	}
	errors := ValidateImportFileHeader(sc1, columnsIndex)
	if len(errors) > 0 {
		errorCSVs = convertErrToErrResForEachLineCSV(errors)
		return
	}

	pIDs := make([]string, 0, len(sc1.GetRow()))
	studentIDWithCourseIDs := []string{}
	for sc1.Scan() {
		if sc1.Text("partner_internal_id") != "" {
			pIDs = append(pIDs, sc1.Text("partner_internal_id"))
			if isUsingVersion2 {
				for _, sCourseID := range strings.Split(sc1.Text("student_course_ids"), "_") {
					if studentCourseIDs := strings.Split(sCourseID, "/"); len(studentCourseIDs) > 1 {
						studentIDWithCourseIDs = append(studentIDWithCourseIDs, studentCourseIDs[0], studentCourseIDs[1])
					}
				}
			}
		}
	}

	sc2 := scanner.NewCSVScanner(bytes.NewReader(data))
	payloads = NewImportLessonPayload().
		WithTimeZone(tz).
		WithScanner(sc2).
		WithPartnerInternalIDs(pIDs).
		WithStudentCourseIDs(studentIDWithCourseIDs).
		WithMasterDataPort(l.MasterDataPort).
		WithUserModulePort(l.UserModulePort).
		WithDateInfoRepo(l.DateInfoRepo).
		WithVersion2(isUsingVersion2).
		WithStudentSubscriptionRepo(l.StudentSubscriptionRepo)

	conn, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, []*lpb.ImportLessonResponse_ImportLessonError{
			{
				RowNumber: int32(0),
				Error:     fmt.Sprintf("%s", err),
			},
		}
	}
	errors = payloads.prepareData(ctx, conn)
	if len(errors) > 0 {
		errorCSVs = convertErrToErrResForEachLineCSV(errors)
	}

	errors = payloads.WithScanner(sc2).buildImportLessonPayload(ctx)
	if len(errors) > 0 {
		errorCSVs = convertErrToErrResForEachLineCSV(errors)
	}
	return payloads, errorCSVs
}

func convertErrToErrResForEachLineCSV(errors map[int]error) []*lpb.ImportLessonResponse_ImportLessonError {
	errorCSVs := []*lpb.ImportLessonResponse_ImportLessonError{}
	for line, err := range errors {
		errorCSVs = append(errorCSVs, &lpb.ImportLessonResponse_ImportLessonError{
			RowNumber: int32(line),
			Error:     fmt.Sprintf("unable to parse this lesson item: %s", err),
		})
	}
	return errorCSVs
}

func ValidateImportFileHeader(sc scanner.CSVScanner, columnsIndex map[string]int) map[int]error {
	totalRows := len(sc.GetRow())
	errors := make(map[int]error)
	currentRow := 1

	if totalRows == 0 {
		errors[currentRow] = fmt.Errorf("request payload empty")
	}
	if totalRows < len(columnsIndex) {
		errors[currentRow] = fmt.Errorf("invalid format: number of column should be greater than or equal %d", len(columnsIndex))
	}
	for colName, colIndex := range columnsIndex {
		if i, ok := sc.Head[colName]; !ok || i != colIndex && colIndex != -1 {
			errors[currentRow] = fmt.Errorf("invalid format: the column have index %d (toLowerCase) should be '%s'", colIndex, colName)
		}
	}
	return errors
}
