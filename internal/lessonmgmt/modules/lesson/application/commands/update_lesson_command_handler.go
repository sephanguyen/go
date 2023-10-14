package commands

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/constants"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/calendar/domain/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	zoom_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	v1 "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/jackc/pgx/v4"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (l *LessonCommandHandler) UpdateLessonOneTime(ctx context.Context, payload UpdateLessonOneTimeCommandRequest) (*domain.Lesson, error) {
	lesson := payload.Lesson
	conn, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if err := database.ExecInTx(ctx, conn, func(ctx context.Context, tx pgx.Tx) (err error) {
		lesson.SaveOneTime()
		if err = lesson.IsValid(ctx, tx); err != nil {
			return fmt.Errorf("invalid lesson: %w", err)
		}

		if err = l.handleStudentReallocatedDiffLocation(ctx, tx, lesson); err != nil {
			return fmt.Errorf("l.handleStudentReallocatedDiffLocation: %w", err)
		}

		isUnleashTeachingTimeToggled, err := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_CourseTeachingTime", l.Env)
		if err != nil {
			return fmt.Errorf("l.connectToUnleash: %w", err)
		}
		if isUnleashTeachingTimeToggled {
			if err = l.AddLessonCourseTeachingTime(ctx, tx, []*domain.Lesson{lesson}, true, payload.TimeZone); err != nil {
				return fmt.Errorf("lesson.AddLessonCourseTeachingTime: %w", err)
			}
		}
		currentLesson := payload.CurrentLesson
		if currentLesson.SchedulingStatus == domain.LessonSchedulingStatusCompleted ||
			currentLesson.SchedulingStatus == domain.LessonSchedulingStatusCanceled {
			lesson.SchedulingStatus = currentLesson.SchedulingStatus
		}
		schedulerID := currentLesson.SchedulerID
		if currentLesson.LocationID != lesson.LocationID && len(schedulerID) > 0 {
			scheduler, err := l.SchedulerRepo.GetByID(ctx, tx, schedulerID)
			if err != nil {
				return fmt.Errorf("l.SchedulerRepo.GetByID: %w", err)
			}
			if constants.Frequency(scheduler.Frequency) == constants.FrequencyWeekly {
				createSchedulerResp, err := l.SchedulerClient.CreateScheduler(ctx, clients.CreateReqCreateScheduler(lesson.StartTime, lesson.EndTime, constants.FrequencyOnce))

				if err != nil {
					return fmt.Errorf("l.SchedulerClient.CreateScheduler: %w", err)
				}
				schedulerID = createSchedulerResp.SchedulerId
			}
		}
		lesson.SchedulerID = schedulerID
		lesson, err = l.LessonRepo.UpdateLesson(ctx, tx, lesson)
		if err != nil {
			return fmt.Errorf("could not persist lesson: lessonRepo.UpdateLesson: %w", err)
		}

		learners := lesson.Learners
		oldLearner := currentLesson.Learners
		if err = l.handingReallocateStudent(ctx, tx, learners, oldLearner, lesson.LessonID); err != nil {
			return fmt.Errorf("l.handingReallocateStudent: %w", err)
		}

		// HACK: Update end_date of lesson's courses => replace by hack in course_dto.go
		if len(learners) == 0 {
			err = l.handleDeleteLessonReport(ctx, tx, []string{lesson.LessonID})
			if err != nil {
				return fmt.Errorf("l.HandleDeleteLessonReport: %w", err)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return lesson, nil
}

func shouldUpdateAttachZoomLink(currentLesson *domain.Lesson, scheduler *dto.Scheduler, command *UpdateRecurringLessonCommandRequest) bool {
	// FE always submit zoomInfo if it exists
	if command.ZoomInfo == nil {
		return false
	}
	if command.ZoomInfo.ZoomID == "" {
		return false
	}
	if command.ZoomInfo.ZoomID != "" && command.ZoomInfo.ZoomID != command.CurrentLesson.ZoomID {
		return true
	}
	// if currentLesson is LessonTeachingMediumZoom and end date of scheduler is different, BE should generate new zoom link
	return currentLesson.TeachingMedium == domain.LessonTeachingMediumZoom && timeutil.BeforeDate(scheduler.EndDate, command.RRuleCmd.UntilDate)
}

func shouldDeleteZoomLink(command *UpdateRecurringLessonCommandRequest) bool {
	if command.ZoomInfo == nil {
		return true
	}
	if command.ZoomInfo.ZoomID == "" {
		return true
	}
	return false
}

func (l *LessonCommandHandler) getZoomLinksAttach(ctx context.Context, command UpdateRecurringLessonCommandRequest) ([]*zoom_domain.GenerateZoomLinkResponse, error) {
	zoomInfo := command.ZoomInfo
	if zoomInfo == nil {
		zoomInfo = &ZoomInfo{
			ZoomAccountID: command.CurrentLesson.ZoomOwnerID,
			ZoomLink:      command.CurrentLesson.ZoomLink,
			ZoomID:        command.CurrentLesson.ZoomID,
		}
	}
	zoomID, _ := strconv.Atoi(zoomInfo.ZoomID)
	linksZoomAttach := []*zoom_domain.GenerateZoomLinkResponse{
		{
			ZoomID: zoomID,
			URL:    zoomInfo.ZoomLink,
			Occurrences: sliceutils.Map(zoomInfo.ZoomOccurrences, func(zoomOccurrence *ZoomOccurrence) *zoom_domain.OccurrenceOfZoomResponse {
				return &zoom_domain.OccurrenceOfZoomResponse{
					OccurrenceID: zoomOccurrence.OccurrenceID,
					StartTime:    zoomOccurrence.StartTime,
				}
			}),
		},
	}
	startTime := command.RRuleCmd.StartTime
	endDate := command.RRuleCmd.UntilDate
	// should skip the first 12 weeks because it was sent from FE

	startTimeGenerate := startTime.AddDate(0, 0, zoom_domain.DaysOfWeek*int(zoom_domain.MaximumRepeatInterval))
	endTimeGenerate := command.RRuleCmd.EndTime.AddDate(0, 0, zoom_domain.DaysOfWeek*int(zoom_domain.MaximumRepeatInterval))

	if canGenerateZoomLink(&startTime, &endDate) {
		requestsGenerateZoomMeeting, err := zoom_domain.ConverterMultiZoomGenerateMeetingRequest(&zoom_domain.MultiZoomGenerateMeetingRequest{
			StartTime:   startTimeGenerate,
			EndTime:     endTimeGenerate,
			TimeZone:    command.TimeZone,
			EndDateTime: command.RRuleCmd.UntilDate,
		})
		if err != nil {
			return nil, fmt.Errorf("could not convert zoom meeting request: %w ", err)
		}
		conn, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
		if err != nil {
			return nil, err
		}
		zoomAccount, err := l.ZoomAccountRepo.GetZoomAccountByID(ctx, conn, zoomInfo.ZoomAccountID)
		if err != nil {
			return nil, fmt.Errorf("could not found zoom account id: %w ", err)
		}
		links, err := l.ZoomService.GenerateMultiZoomLink(ctx, zoomAccount.Email, requestsGenerateZoomMeeting)
		if err != nil {
			return nil, fmt.Errorf("could not generate zoom meeting request: %w ", err)
		}
		linksZoomAttach = append(linksZoomAttach, links...)
	}
	return linksZoomAttach, nil
}

func (l *LessonCommandHandler) UpdateRecurringLesson(ctx context.Context, command UpdateRecurringLessonCommandRequest) (res []*domain.Lesson, lessonMap map[string]*domain.Lesson, err error) { //nolint:gocyclo
	selectedLesson := command.SelectedLesson
	currentLesson := command.CurrentLesson
	conn, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, nil, err
	}
	if err = database.ExecInTx(ctx, conn, func(ctx context.Context, tx pgx.Tx) (err error) {
		var stateChanged domain.StateChangedLesson
		if !currentLesson.StartTime.Equal(selectedLesson.StartTime) || !currentLesson.EndTime.Equal(selectedLesson.EndTime) {
			stateChanged.ChanTime = true
		}
		if currentLesson.LocationID != selectedLesson.LocationID {
			stateChanged.ChanLocation = true
		}

		dateInfo, err := selectedLesson.GetDateInfoByDateAndCenterID(ctx, tx, command.RRuleCmd.StartTime, command.RRuleCmd.UntilDate, selectedLesson.LocationID)
		if err != nil {
			return err
		}
		if err = selectedLesson.CheckClosedDate(command.RRuleCmd.StartTime, dateInfo); err != nil {
			return err
		}
		skipDates := selectedLesson.GetNonRegularDatesExceptFirstDate(ctx, dateInfo)
		schedulerID := currentLesson.SchedulerID
		rrlue, err := domain.NewRecurrenceRule(domain.Option{
			Freq:         domain.WEEKLY,
			StartTime:    command.RRuleCmd.StartTime,
			EndTime:      command.RRuleCmd.EndTime,
			UntilDate:    command.RRuleCmd.UntilDate,
			ExcludeDates: skipDates,
		})
		if err != nil {
			return fmt.Errorf("could not init recurrence rule: %w", err)
		}

		recurSet := rrlue.ExceptFirst()
		lessonChain, err := l.LessonRepo.GetLessonBySchedulerID(ctx, tx, schedulerID)
		if err != nil {
			return fmt.Errorf("l.LessonRepo.GetLessonBySchedulerID: %w", err)
		}

		existedScheduler := &entities.Scheduler{
			SchedulerID:   schedulerID,
			SchedulerRepo: l.SchedulerRepo,
		}

		scheduler, err := existedScheduler.Get(ctx, tx)
		if err != nil {
			return fmt.Errorf("get scheduler error: %w", err)
		}

		isShouldUpdateAttachZoomLink := shouldUpdateAttachZoomLink(currentLesson, scheduler, &command)

		if stateChanged.IsChanged() {
			var oldDate time.Time
			for _, ls := range lessonChain {
				if ls.EndTime.Before(currentLesson.EndTime) {
					oldDate = ls.EndTime
				}
			}
			if oldDate.IsZero() {
				oldDate = currentLesson.EndTime
			}
			existedScheduler.EndDate = oldDate
			_, err = l.SchedulerClient.UpdateScheduler(ctx, &v1.UpdateSchedulerRequest{SchedulerId: schedulerID, EndDate: timestamppb.New(oldDate)})
			if err != nil {
				return fmt.Errorf("l.SchedulerClient.UpdateScheduler: %w", err)
			}
			freq := domain.FrequencyName[rrlue.Option.Freq]
			createSchedulerResp, err := l.SchedulerClient.CreateScheduler(ctx, clients.CreateReqCreateScheduler(command.RRuleCmd.StartTime, command.RRuleCmd.UntilDate, constants.Frequency(freq)))

			if err != nil {
				return fmt.Errorf("l.SchedulerRepo.Create: %w", err)
			}
			schedulerID = createSchedulerResp.SchedulerId
		} else {
			existedScheduler.EndDate = command.RRuleCmd.UntilDate
			_, err = l.SchedulerClient.UpdateScheduler(ctx, &v1.UpdateSchedulerRequest{SchedulerId: schedulerID, EndDate: timestamppb.New(existedScheduler.EndDate)})
			if err != nil {
				return fmt.Errorf("l.SchedulerClient.UpdateScheduler: %w", err)
			}
		}

		recurringLesson := &domain.RecurringLesson{
			ID: schedulerID,
		}
		learner := selectedLesson.Learners
		studentIDWithCourseID := make([]string, 0, len(learner)*2)
		for _, l := range learner {
			studentIDWithCourseID = append(studentIDWithCourseID, l.LearnerID, l.CourseID)
		}
		// Get student course duration of all students
		var studentSubscriptions user_domain.StudentSubscriptions
		if len(studentIDWithCourseID) > 0 {
			var locationID []string
			if len(selectedLesson.LocationID) > 0 {
				locationID = append(locationID, selectedLesson.LocationID)
			}
			studentSubscriptions, err = l.StudentSubscriptionRepo.GetStudentCourseSubscriptions(ctx, tx, locationID, studentIDWithCourseID...)
			if err != nil {
				return fmt.Errorf("l.StudentSubscriptionRepo.GetStudentCourseSubscriptions %w: ", err)
			}
		}
		studentSubMap := make(map[string]time.Time)
		for _, ss := range studentSubscriptions {
			studentSubMap[ss.StudentWithCourseID()] = ss.EndAt
		}
		selectedLesson.SchedulerID = schedulerID
		if currentLesson.SchedulingStatus == domain.LessonSchedulingStatusCompleted ||
			currentLesson.SchedulingStatus == domain.LessonSchedulingStatusCanceled {
			selectedLesson.SchedulingStatus = currentLesson.SchedulingStatus
		}
		var followingLessonID domain.FollowingLessonID

		lockedLessonIDs := []string{}

		lessonMap = make(map[string]*domain.Lesson)
		for _, ls := range lessonChain {
			if ls.StartTime.After(currentLesson.StartTime) {
				followingLessonID.Add(ls.LessonID)
				lessonMap[ls.LessonID] = ls
				if ls.IsLocked {
					lockedLessonIDs = append(lockedLessonIDs, ls.LessonID)
				}
			}
		}
		upsertedLesson := []*domain.Lesson{selectedLesson}
		upsertedLessonID := []string{selectedLesson.LessonID}
		loc := timeutil.Location(command.TimeZone)
		linksZoomAttach := []*zoom_domain.GenerateZoomLinkResponse{}
		if isShouldUpdateAttachZoomLink {
			linksZoomAttach, err = l.getZoomLinksAttach(ctx, command)
			if err != nil {
				return err
			}
			// always have a item
			firstLink := linksZoomAttach[0]
			selectedLesson.ZoomID = fmt.Sprint(firstLink.ZoomID)
			selectedLesson.ZoomLink = firstLink.URL
			if len(firstLink.Occurrences) > 0 {
				selectedLesson.ZoomOccurrenceID = firstLink.Occurrences[0].OccurrenceID
			}
		}
		if command.ZoomInfo != nil {
			selectedLesson.ZoomOwnerID = command.ZoomInfo.ZoomAccountID
		}
		zoomAccountID := command.CurrentLesson.ZoomOwnerID
		if command.ZoomInfo != nil {
			zoomAccountID = command.ZoomInfo.ZoomAccountID
		}
		studentRemoved := learner.GetStudentRemoved(currentLesson.Learners)

		isChangedStudentInfo := selectedLesson.Learners.IsChangedStudentInfo(currentLesson.Learners)
		for _, r := range recurSet {
			ls := *selectedLesson
			ls.ResetID()
			ls.StartTime = r.StartTime
			ls.EndTime = r.EndTime
			ls.SchedulerID = schedulerID
			ls.Material = &domain.LessonMaterial{}
			currentLearners := domain.LessonLearners{}
			if id, ok := followingLessonID.Pop(); ok {
				ls.LessonID = id
				ls.Persisted = true
				upsertedLessonID = append(upsertedLessonID, id)
				lesson := lessonMap[ls.LessonID]
				if lesson.IsLocked {
					continue
				}
				currentLearners = lesson.Learners
				schedulingStatus := selectSchedulingStatus(lesson.SchedulingStatus, selectedLesson.SchedulingStatus)
				ls.SchedulingStatus = schedulingStatus
				if shouldDeleteZoomLink(&command) {
					ls.ZoomLink = ""
					ls.ZoomID = ""
					ls.ZoomOwnerID = ""
					ls.ZoomOccurrenceID = ""
				} else {
					ls.ZoomLink = lesson.ZoomLink
					ls.ZoomID = lesson.ZoomID
					ls.ZoomOwnerID = lesson.ZoomOwnerID
					ls.ZoomOccurrenceID = lesson.ZoomOccurrenceID
				}
				ls.Material = lesson.Material
			}
			if isShouldUpdateAttachZoomLink {
				indexLink, indexOccurrence := calculateIndexZoomLinkAndIndexOccurrence(&command.RRuleCmd.StartTime, &ls.StartTime, linksZoomAttach)
				zoomLink := linksZoomAttach[indexLink]
				ls.ZoomID = fmt.Sprint(zoomLink.ZoomID)
				ls.ZoomLink = zoomLink.URL
				if len(zoomLink.Occurrences) > 0 {
					ls.ZoomOccurrenceID = zoomLink.Occurrences[indexOccurrence].OccurrenceID
				}
			}
			if command.ZoomInfo != nil {
				ls.ZoomOwnerID = zoomAccountID
			}
			learners := domain.LessonLearners{}
			if len(ls.LessonID) > 0 {
				if isChangedStudentInfo {
					learnerMap := map[string]bool{}
					for _, ln := range currentLearners {
						if ln.IsReallocated() || !slices.Contains(studentRemoved, ln.LearnerID) {
							learners = append(learners, ln)
							learnerMap[ln.LearnerID] = true
						}
					}
					for _, lLearner := range ls.Learners {
						if lLearner.IsReallocated() || learnerMap[lLearner.LearnerID] {
							continue
						}
						ln := *lLearner
						endAt, exists := studentSubMap[ln.StudentWithCourse()]
						if !exists {
							return fmt.Errorf("UpdateRecurringLesson: cannot find student course (%s,%s) duration", ln.LearnerID, ln.CourseID)
						}
						if ln.IsValidForAllocateToLesson(endAt, r.StartTime, loc) {
							ln.EmptyAttendanceInfo()
							learners = append(learners, &ln)
						}
					}
				} else {
					learners = currentLearners
				}
			} else {
				for _, sLearner := range selectedLesson.Learners {
					if sLearner.IsReallocated() {
						continue
					}
					learners = append(learners, sLearner)
				}
			}
			ls.Learners = learners
			upsertedLesson = append(upsertedLesson, &ls)
		}

		// update schedulerID for many lessons are locked
		if err := l.updateSchedulerID(ctx, tx, lockedLessonIDs, schedulerID); err != nil {
			return err
		}
		if !followingLessonID.IsEmpty() {
			remainingLesson := followingLessonID.GetNoLockedLessons(lockedLessonIDs)
			err := l.LessonRepo.Delete(ctx, tx, remainingLesson)
			if err != nil {
				return fmt.Errorf("l.LessonRepo.Delete %w: ", err)
			}
			// publish event lesson deleted
			if err = l.LessonProducer.PublishLessonEvt(ctx,
				&bpb.EvtLesson{
					Message: &bpb.EvtLesson_DeletedLessons_{
						DeletedLessons: &bpb.EvtLesson_DeletedLessons{
							LessonIds: remainingLesson,
						},
					},
				}); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}
		recurringLesson.Lessons = upsertedLesson
		recurringLesson.Save()
		if err = recurringLesson.IsValid(ctx, tx); err != nil {
			return fmt.Errorf("invalid lesson: %w", err)
		}

		if err = l.handleStudentReallocatedDiffLocation(ctx, tx, recurringLesson.GetBaseLesson()); err != nil {
			return fmt.Errorf("l.handleStudentReallocatedDiffLocation: %w", err)
		}

		isUnleashTeachingTimeToggled, err := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_CourseTeachingTime", l.Env)
		if err != nil {
			return fmt.Errorf("l.connectToUnleash: %w", err)
		}
		if isUnleashTeachingTimeToggled {
			if err = l.AddLessonCourseTeachingTime(ctx, tx, recurringLesson.Lessons, true, command.TimeZone); err != nil {
				return fmt.Errorf("lesson.AddLessonCourseTeachingTime: %w", err)
			}
		}
		if _, err = l.LessonRepo.UpsertLessons(ctx, tx, recurringLesson); err != nil {
			return fmt.Errorf("l.LessonRepo.UpsertLessons: %w", err)
		}

		learners := selectedLesson.Learners
		oldLearner := currentLesson.Learners

		if err = l.handingReallocateStudent(ctx, tx, learners, oldLearner, selectedLesson.LessonID); err != nil {
			return fmt.Errorf("l.handingReallocateStudent: %w", err)
		}

		res = recurringLesson.Lessons
		if len(recurringLesson.GetBaseLesson().Learners) == 0 {
			err = l.handleDeleteLessonReport(ctx, tx, upsertedLessonID)
			if err != nil {
				return fmt.Errorf("l.handleDeleteLessonReport errs: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}
	return
}

func (l *LessonCommandHandler) handingReallocateStudent(ctx context.Context, tx pgx.Tx, learners, oldLearner domain.LessonLearners, lessonId string) error {
	// handle changing attendance status of student from reallocate to another
	studentIds := learners.GetStudentNoPendingReallocate(oldLearner)
	if len(studentIds) > 0 {
		reallocations, err := l.ReallocationRepo.GetFollowingReallocation(ctx, tx, lessonId, studentIds)
		if err != nil {
			return fmt.Errorf("l.ReallocationRepo.GetFollowingReallocation: %w", err)
		}
		reallocateStudent := []string{}
		lessonMemberDeleted := make([]*domain.LessonMember, 0)
		for _, r := range reallocations {
			reallocateStudent = append(reallocateStudent, r.StudentID, r.OriginalLessonID)
			if slices.Contains(studentIds, r.StudentID) && r.OriginalLessonID == lessonId {
				continue
			}
			lessonMemberDeleted = append(lessonMemberDeleted, &domain.LessonMember{
				LessonID:  r.OriginalLessonID,
				StudentID: r.StudentID,
			})
		}
		if err := l.ReallocationRepo.SoftDelete(ctx, tx, reallocateStudent, true); err != nil {
			return fmt.Errorf("l.ReallocationRepo.SoftDelete: %w", err)
		}
		if len(lessonMemberDeleted) > 0 {
			if err = l.LessonMemberRepo.DeleteLessonMembers(ctx, tx, lessonMemberDeleted); err != nil {
				return fmt.Errorf("l.LessonMemberRepo.DeleteLessonMembers: %w", err)
			}
		}
	}
	// handle students are removed from lesson
	oldLearnerID := oldLearner.GetLearnerIDs()
	reallocation, err := l.ReallocationRepo.GetByNewLessonID(ctx, tx, oldLearnerID, lessonId)
	if err != nil {
		return fmt.Errorf("l.ReallocationRepo.GetByNewLessonID: %w", err)
	}
	if len(reallocation) > 0 {
		for _, r := range reallocation {
			learner := oldLearner.GetLearnerByID(r.StudentID)
			learner.Reallocate = &domain.Reallocate{
				OriginalLessonID: r.OriginalLessonID,
			}
		}
	}
	studentRemoved := learners.GetReallocateStudentRemoved(oldLearner)
	if len(studentRemoved) > 0 {
		studentWithLesson := []string{}
		for _, studentId := range studentRemoved {
			studentWithLesson = append(studentWithLesson, studentId, lessonId)
		}
		if err := l.ReallocationRepo.SoftDelete(ctx, tx, studentWithLesson, false); err != nil {
			return fmt.Errorf("l.ReallocationRepo.SoftDelete: %w", err)
		}
		if err := l.ReallocationRepo.CancelIfStudentReallocated(ctx, tx, studentWithLesson); err != nil {
			return fmt.Errorf("l.ReallocationRepo.CancelIfStudentReallocated: %w", err)
		}
	}
	return nil
}

func (l *LessonCommandHandler) UpdateLessonSchedulingStatus(ctx context.Context, cmdReq *UpdateLessonStatusCommandRequest) (*UpdateLessonStatusCommandResponse, error) {
	conn, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	currentLesson, err := l.LessonRepo.GetLessonWithSchedulerInfoByLessonID(ctx, conn, cmdReq.LessonID)
	if err != nil {
		return nil, fmt.Errorf("l.LessonRepo.GetLessonByID: %w", err)
	}
	var updatedLesson []*domain.Lesson
	curStatus := currentLesson.SchedulingStatus
	switch cmdReq.SavingType {
	case lpb.SavingType_THIS_ONE:
		newStatus := domain.LessonSchedulingStatus(cmdReq.SchedulingStatus)
		if !checkUpdatedSchedulingStatus(curStatus, newStatus) {
			return nil, fmt.Errorf("cannot change scheduling status from `%s` to `%s`", curStatus, newStatus)
		}
		if currentLesson.IsLock() {
			return nil, fmt.Errorf("lesson is locked so cannot update lesson status")
		}
		if newStatus == domain.LessonSchedulingStatusCanceled {
			if currentLesson != nil && currentLesson.SchedulerInfo != nil && currentLesson.SchedulerInfo.Freq == "once" && (currentLesson.ZoomID != "" || currentLesson.ClassDoRoomID != "") {
				if currentLesson.ZoomID != "" {
					_, err = l.ZoomService.RetryDeleteZoomLink(ctx, currentLesson.ZoomID)
					if err != nil {
						return nil, fmt.Errorf("ZoomService.RetryDeleteZoomLink: %w", err)
					}
					err = l.LessonRepo.RemoveZoomLinkByLessonID(ctx, conn, currentLesson.LessonID)
					if err != nil {
						return nil, fmt.Errorf("LessonRepo.RemoveZoomLinkByLessonID: %w", err)
					}
				} else {
					err = l.LessonRepo.RemoveClassDoLinkByLessonID(ctx, conn, currentLesson.LessonID)
					if err != nil {
						return nil, fmt.Errorf("LessonRepo.RemoveClassDoLinkByLessonID: %w", err)
					}
				}
			}
		}
		currentLesson.SchedulingStatus = newStatus
		ls, err := l.LessonRepo.UpdateLessonSchedulingStatus(ctx, conn, currentLesson)
		if err != nil {
			return nil, fmt.Errorf("l.LessonRepo.UpdateLessonSchedulingStatus: %w", err)
		}
		ls.PreSchedulingStatus = curStatus
		updatedLesson = append(updatedLesson, ls)
	case lpb.SavingType_THIS_AND_FOLLOWING:
		lessonChain, err := l.LessonRepo.GetLessonBySchedulerID(ctx, conn, currentLesson.SchedulerID)
		if err != nil {
			return nil, fmt.Errorf("l.LessonRepo.GetLessonBySchedulerID: %w", err)
		}
		lessonStatusMap := make(map[string]domain.LessonSchedulingStatus, 0)
		newStatus := domain.LessonSchedulingStatus(cmdReq.SchedulingStatus)
		for _, ls := range lessonChain {
			curStatus := ls.SchedulingStatus
			if curStatus == domain.LessonSchedulingStatusCompleted ||
				curStatus == domain.LessonSchedulingStatusCanceled ||
				curStatus == newStatus {
				continue
			}
			if ls.StartTime.After(currentLesson.StartTime) ||
				ls.StartTime.Equal(currentLesson.StartTime) {
				lessonStatusMap[ls.LessonID] = newStatus
				ls.PreSchedulingStatus = curStatus
				ls.SchedulingStatus = newStatus
				updatedLesson = append(updatedLesson, ls)
			}
		}
		if len(lessonStatusMap) > 0 {
			err = l.LessonRepo.UpdateSchedulingStatus(ctx, conn, lessonStatusMap)
			if err != nil {
				return nil, fmt.Errorf("l.LessonRepo.UpdateSchedulingStatus: %w", err)
			}
		}
	}
	return &UpdateLessonStatusCommandResponse{
		UpdatedLesson: updatedLesson,
	}, nil
}

func checkUpdatedSchedulingStatus(preStatus, newStatus domain.LessonSchedulingStatus) bool {
	if preStatus == newStatus {
		return false
	}
	switch preStatus {
	case domain.LessonSchedulingStatusDraft:
		return newStatus == domain.LessonSchedulingStatusPublished
	case domain.LessonSchedulingStatusCompleted:
		return newStatus == domain.LessonSchedulingStatusCanceled
	case domain.LessonSchedulingStatusPublished:
		return newStatus == domain.LessonSchedulingStatusCompleted ||
			newStatus == domain.LessonSchedulingStatusCanceled ||
			newStatus == domain.LessonSchedulingStatusDraft
	case domain.LessonSchedulingStatusCanceled:
		return newStatus == domain.LessonSchedulingStatusPublished
	default:
		return false
	}
}

func selectSchedulingStatus(curStatus, newStatus domain.LessonSchedulingStatus) domain.LessonSchedulingStatus {
	if curStatus == domain.LessonSchedulingStatusCompleted ||
		curStatus == domain.LessonSchedulingStatusCanceled {
		return curStatus
	}
	if newStatus == domain.LessonSchedulingStatusCompleted ||
		newStatus == domain.LessonSchedulingStatusCanceled {
		return curStatus
	}
	return newStatus
}

func (l *LessonCommandHandler) updateSchedulerID(ctx context.Context, tx pgx.Tx, lockedLessonIDs []string, schedulerID string) error {
	if len(lockedLessonIDs) > 0 {
		if err := l.LessonRepo.UpdateSchedulerID(ctx, tx, lockedLessonIDs, schedulerID); err != nil {
			return fmt.Errorf("could not update lesson schedulerID for the lessons are locked %s", schedulerID)
		}
	}
	return nil
}

func (l *LessonCommandHandler) handleDeleteLessonReport(ctx context.Context, tx pgx.Tx, lessonIDs []string) error {
	// Delete lesson report when don't have any student
	err := l.LessonReportRepo.DeleteReportsBelongToLesson(ctx, tx, lessonIDs)
	if err != nil {
		return fmt.Errorf("LessonReportRepo.DeleteReportsBelongToLesson: %w", err)
	}
	return nil
}

func (l *LessonCommandHandler) BulkUpdateLessonSchedulingStatus(ctx context.Context, req *BulkUpdateLessonSchedulingStatusCommandRequest) (*BulkUpdateLessonSchedulingStatusCommandResponse, error) {
	lessonStatusMap := make(map[string]domain.LessonSchedulingStatus, 0)

	switch req.Action {
	case lpb.LessonBulkAction_LESSON_BULK_ACTION_CANCEL:
		{
			for _, lesson := range req.Lessons {
				if (lesson.SchedulingStatus == domain.LessonSchedulingStatusCompleted || lesson.SchedulingStatus == domain.LessonSchedulingStatusPublished) &&
					!lesson.IsLocked {
					lessonStatusMap[lesson.LessonID] = domain.LessonSchedulingStatusCanceled
					lesson.SchedulingStatus = domain.LessonSchedulingStatusCanceled
				}
			}
		}
	case lpb.LessonBulkAction_LESSON_BULK_ACTION_PUBLISH:
		{
			for _, lesson := range req.Lessons {
				if lesson.SchedulingStatus == domain.LessonSchedulingStatusDraft &&
					!lesson.IsLocked && lesson.CheckLessonValidForPublish() {
					lessonStatusMap[lesson.LessonID] = domain.LessonSchedulingStatusPublished
					lesson.SchedulingStatus = domain.LessonSchedulingStatusPublished
				}
			}
		}
	default:
		return nil, status.Error(codes.Internal, "unsupported action for bulk action")
	}
	conn, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return &BulkUpdateLessonSchedulingStatusCommandResponse{UpdatedLessons: req.Lessons}, err
	}
	err = l.LessonRepo.UpdateSchedulingStatus(ctx, conn, lessonStatusMap)
	return &BulkUpdateLessonSchedulingStatusCommandResponse{UpdatedLessons: req.Lessons}, err
}

func (l *LessonCommandHandler) MarkStudentAsReallocate(ctx context.Context, req *MarkStudentAsReallocateRequest) error {
	conn, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return err
	}
	err = database.ExecInTx(ctx, conn, func(ctx context.Context, tx pgx.Tx) error {
		lessonMember, err := l.LessonMemberRepo.FindByID(ctx, tx, req.Member.LessonID, req.Member.StudentID)

		if err != nil {
			return fmt.Errorf("LessonMemberRepo.FindByID: %v", err)
		}

		if lessonMember.AttendanceStatus != string(domain.StudentAttendStatusAbsent) {
			return fmt.Errorf("student's attendance status is not absent: %s", lessonMember.AttendanceStatus)
		}

		members := []*domain.LessonMember{req.Member}
		if err = l.LessonMemberRepo.UpdateLessonMembersFields(
			ctx,
			tx,
			members,
			repo.UpdateLessonMemberFields{"attendance_status"},
		); err != nil {
			return fmt.Errorf("LessonMemberRepo.UpdateLessonMembersFields: %v", err)
		}

		req.ReAllocations.CourseID = lessonMember.CourseID
		reAllocations := []*domain.Reallocation{req.ReAllocations}
		if err = l.ReallocationRepo.UpsertReallocation(ctx, tx, req.Member.LessonID, reAllocations); err != nil {
			return fmt.Errorf("ReallocationRepo.UpsertReallocation: %v", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (l *LessonCommandHandler) UpdateFutureLessonsWhenCourseChanged(ctx context.Context, courseIDs []string, timezone string) error {
	conn, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return err
	}
	err = database.ExecInTx(ctx, conn, func(ctx context.Context, tx pgx.Tx) error {
		lessons, er := l.LessonRepo.GetFutureLessonsByCourseIDs(ctx, tx, courseIDs, timezone)
		if er != nil {
			return er
		}

		if er = l.AddLessonCourseTeachingTime(ctx, tx, lessons, true, timezone); er != nil {
			return fmt.Errorf("lesson.AddLessonCourseTeachingTime: %w", er)
		}

		if er = l.LessonRepo.UpdateLessonsTeachingTime(ctx, tx, lessons); er != nil {
			return fmt.Errorf("lessonRepo.UpdateLessonsTeachingTime: %w", er)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
