package controller

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	lrd "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	user_infras "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LessonReaderService struct {
	wrapperConnection     *support.WrapperDBConnection
	retrieveLessonCommand application.RetrieveLessonCommand
	lessonQueryHandler    queries.LessonQueryHandler
	env                   string
	unleashClientIns      unleashclient.ClientInstance
}

func NewLessonReaderService(
	wrapperConnection *support.WrapperDBConnection,
	lessonRepo infrastructure.LessonRepo,
	searchRepo infrastructure.SearchRepo,
	lessonTeacherRepo infrastructure.LessonTeacherRepo,
	lessonMemberRepo infrastructure.LessonMemberRepo,
	lessonGroupRepo infrastructure.LessonGroupRepo,
	lessonClassroomRepo infrastructure.LessonClassroomRepo,
	userRepo user_infras.UserRepo,
	env string,
	unleashClientIns unleashclient.ClientInstance,
) *LessonReaderService {
	return &LessonReaderService{
		wrapperConnection: wrapperConnection,
		lessonQueryHandler: queries.LessonQueryHandler{
			WrapperConnection:   wrapperConnection,
			LessonRepo:          lessonRepo,
			LessonTeacherRepo:   lessonTeacherRepo,
			LessonMemberRepo:    lessonMemberRepo,
			LessonClassroomRepo: lessonClassroomRepo,
			UserRepo:            userRepo,
		},
		retrieveLessonCommand: application.RetrieveLessonCommand{
			WrapperConnection: wrapperConnection,
			LessonRepo:        lessonRepo,
			LessonMemberRepo:  lessonMemberRepo,
			LessonGroupRepo:   lessonGroupRepo,
			SearchRepo:        searchRepo,
		},
		env:              env,
		unleashClientIns: unleashClientIns,
	}
}

func (l *LessonReaderService) RetrieveLessonsV2(ctx context.Context, req *lpb.RetrieveLessonsRequest) (*lpb.RetrieveLessonsResponse, error) {
	if err := validateRetrieveLessonsRequest(ctx, req); err != nil {
		return nil, err
	}
	args, isEmptyIntersectedLocation := filterArgsFromRequestPayload(ctx, req)

	isUnleashToggledGivenName, err := l.unleashClientIns.IsFeatureEnabled("Lesson_Student_SearchInGivenNameColumn", l.env)
	if err != nil {
		return nil, fmt.Errorf("l.connectToUnleash for Lesson_Student_SearchInGivenNameColumn: %w", err)
	}
	if isUnleashToggledGivenName {
		args.SearchInGivenNameColumn = true
	} else {
		args.SearchInGivenNameColumn = false
	}

	if isEmptyIntersectedLocation {
		return &lpb.RetrieveLessonsResponse{
			Items:      []*lpb.RetrieveLessonsResponse_Lesson{},
			TotalItems: uint32(0),
			NextPage: &cpb.Paging{
				Limit: req.Paging.Limit,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: "",
				},
			},
			PreviousPage: &cpb.Paging{
				Limit: req.Paging.Limit,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: "",
				},
			},
			TotalLesson: uint32(0),
		}, nil
	}

	result := l.lessonQueryHandler.RetrieveLesson(ctx, &args)

	if result.Error != nil {
		return nil, result.Error
	}

	items := []*lpb.RetrieveLessonsResponse_Lesson{}
	for _, lesson := range result.Lessons {
		items = append(items, toLessonPb(lesson))
	}
	lastItem := ""
	lessonLen := len(result.Lessons)
	if lessonLen > 0 {
		lastItem = result.Lessons[lessonLen-1].LessonID
	}

	return &lpb.RetrieveLessonsResponse{
		Items:      items,
		TotalItems: result.Total,
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: lastItem,
			},
		},
		PreviousPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: result.OffsetID,
			},
		},
		TotalLesson: result.Total,
	}, nil
}

func (l *LessonReaderService) RetrieveLessonByID(ctx context.Context, req *lpb.RetrieveLessonByIDRequest) (*lpb.RetrieveLessonByIDResponse, error) {
	lesson, err := l.retrieveLessonCommand.GetLessonByID(ctx, req.LessonId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf(`cannot get lesson by id %s`, req.LessonId))
	}
	lessonMembers := make([]*lpb.LessonMember, 0, len(lesson.Learners))
	for _, learner := range lesson.Learners {
		lessonMembers = append(lessonMembers, &lpb.LessonMember{
			StudentId:        learner.LearnerID,
			CourseId:         learner.CourseID,
			LocationId:       learner.LocationID,
			AttendanceStatus: lpb.StudentAttendStatus(lpb.StudentAttendStatus_value[string(learner.AttendStatus)]),
		})
	}
	mediaIDs := make([]string, 0)
	if lesson.Material != nil {
		mediaIDs = lesson.Material.MediaIDs
	}
	lessonRes := &lpb.Lesson{
		LessonId:         lesson.LessonID,
		LocationId:       lesson.LocationID,
		TeacherIds:       lesson.Teachers.GetIDs(),
		LearnerMembers:   lessonMembers,
		MediaIds:         mediaIDs,
		TeachingMethod:   cpb.LessonTeachingMethod(cpb.LessonTeachingMethod_value[string(lesson.TeachingMethod)]),
		TeachingMedium:   cpb.LessonTeachingMedium(cpb.LessonTeachingMedium_value[string(lesson.TeachingMedium)]),
		SchedulingStatus: cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[string(lesson.SchedulingStatus)]),
		StartTime:        timestamppb.New(lesson.StartTime),
		EndTime:          timestamppb.New(lesson.EndTime),
		CreatedAt:        timestamppb.New(lesson.CreatedAt),
		UpdatedAt:        timestamppb.New(lesson.UpdatedAt),
		LessonCapacity:   uint32(lesson.LessonCapacity),
	}
	return &lpb.RetrieveLessonByIDResponse{
		Lesson: lessonRes,
	}, nil
}

func validateRetrieveLessonsRequest(ctx context.Context, req *lpb.RetrieveLessonsRequest) error {
	if req.GetPaging() == nil {
		return status.Error(codes.Internal, "missing paging info")
	}
	if req.GetCurrentTime() == nil {
		return status.Error(codes.Internal, "missing current time")
	}

	dows := req.GetFilter().GetDateOfWeeks()
	timeZone := req.GetFilter().GetTimeZone()

	if (len(dows) > 0 || req.GetFilter().GetFromTime() != nil || req.GetFilter().GetToTime() != nil) && timeZone == "" {
		return status.Error(codes.Internal, "missing timezone")
	}
	return nil
}

func (l *LessonReaderService) RetrieveLessons(ctx context.Context, req *lpb.RetrieveLessonsRequest) (*lpb.RetrieveLessonsResponse, error) {
	var (
		lessons   []*domain.Lesson
		total     uint32
		prePageID string
		err       error
	)

	if err = validateRetrieveLessonsRequest(ctx, req); err != nil {
		return nil, err
	}
	args, isExisted := filterArgsFromRequest(ctx, req)

	if isExisted {
		lessons, total, prePageID, err = l.retrieveLessonCommand.Search(ctx, &args)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	items := make([]*lpb.RetrieveLessonsResponse_Lesson, 0, len(lessons))
	for _, lesson := range lessons {
		items = append(items, toLessonPb(lesson))
	}
	lastItem := ""
	if len(lessons) > 0 {
		lastItem = lessons[len(lessons)-1].LessonID
	}

	return &lpb.RetrieveLessonsResponse{
		Items:      items,
		TotalItems: total,
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: lastItem,
			},
		},
		PreviousPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: prePageID,
			},
		},
		TotalLesson: total,
	}, nil
}

func (l *LessonReaderService) RetrieveLessonsOnCalendar(ctx context.Context, req *lpb.RetrieveLessonsOnCalendarRequest) (*lpb.RetrieveLessonsOnCalendarResponse, error) {
	if err := validateLessonsOnCalendarRequest(req); err != nil {
		return nil, err
	}

	args, isLocationExisting := filterArgsFromLessonsOnCalendarRequest(req)
	if !isLocationExisting {
		return &lpb.RetrieveLessonsOnCalendarResponse{
			Items: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson{},
		}, nil
	}

	result := l.lessonQueryHandler.RetrieveLessonsOnCalendar(ctx, &args)
	if result.Error != nil {
		return nil, result.Error
	}

	items := make([]*lpb.RetrieveLessonsOnCalendarResponse_Lesson, 0, len(result.Lessons))
	for _, lesson := range result.Lessons {
		items = append(items, toLessonOnCalendarPb(lesson))
	}

	return &lpb.RetrieveLessonsOnCalendarResponse{
		Items: items,
	}, nil
}

func validateLessonsOnCalendarRequest(req *lpb.RetrieveLessonsOnCalendarRequest) error {
	if len(req.GetCalendarView().Enum().String()) == 0 {
		return status.Error(codes.Internal, "request missing calendar view")
	}

	if len(req.GetLocationId()) == 0 {
		return status.Error(codes.Internal, "request missing location ID")
	}

	if len(req.GetTimezone()) == 0 {
		return status.Error(codes.Internal, "request missing timezone")
	}

	if req.GetToDate().AsTime().Before(req.GetFromDate().AsTime()) {
		return status.Error(codes.Internal, "to date cannot be before from date")
	}

	return nil
}

func filterArgsFromLessonsOnCalendarRequest(req *lpb.RetrieveLessonsOnCalendarRequest) (args payloads.GetLessonListOnCalendarArgs, isLocationExisting bool) {
	args = payloads.GetLessonListOnCalendarArgs{
		View:                                payloads.CalendarView(req.GetCalendarView().Enum().String()),
		FromDate:                            req.FromDate.AsTime(),
		ToDate:                              req.ToDate.AsTime(),
		Timezone:                            req.GetTimezone(),
		LocationID:                          "",
		StudentIDs:                          req.GetFilter().GetStudentIds(),
		CourseIDs:                           req.GetFilter().GetCourseIds(),
		TeacherIDs:                          req.GetFilter().GetTeacherIds(),
		ClassIDs:                            req.GetFilter().GetClassIds(),
		IsIncludeNoneAssignedTeacherLessons: req.GetFilter().GetNoneAssignedTeacherLessons(),
	}

	selectedLocation := req.GetLocationId()
	selectedLocationLength := len(selectedLocation)

	locationList := req.GetLocationIds()
	locationListLength := len(locationList)

	if selectedLocationLength > 0 && locationListLength > 0 {
		for _, locationID := range locationList {
			if selectedLocation == locationID {
				args.LocationID = selectedLocation
			}
		}

		if len(args.LocationID) == 0 {
			return args, false
		}
	} else {
		args.LocationID = selectedLocation
	}

	return args, true
}

func toLessonOnCalendarPb(l *domain.Lesson) *lpb.RetrieveLessonsOnCalendarResponse_Lesson {
	lessonLearnersLength := len(l.Learners)
	lessonLearners := make([]*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonMember, 0, lessonLearnersLength)
	lessonTeachersLength := len(l.Teachers)
	lessonTeachers := make([]*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonTeacher, 0, lessonTeachersLength)
	lessonClassroomsLength := len(l.Classrooms)
	lessonClassrooms := make([]*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonClassroom, 0, lessonClassroomsLength)

	lessonGroupGrade := ""

	for _, learner := range l.Learners {
		lessonLearners = append(lessonLearners, &lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonMember{
			StudentId:        learner.LearnerID,
			CourseId:         learner.CourseID,
			Grade:            learner.Grade,
			StudentName:      learner.LearnerName,
			CourseName:       learner.CourseName,
			AttendanceStatus: lpb.StudentAttendStatus(lpb.StudentAttendStatus_value[string(learner.AttendStatus)]),
		})
	}

	for _, teacher := range l.Teachers {
		lessonTeachers = append(lessonTeachers, &lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonTeacher{
			TeacherId:   teacher.TeacherID,
			TeacherName: teacher.Name,
		})
	}

	for _, classroom := range l.Classrooms {
		lessonClassrooms = append(lessonClassrooms, &lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonClassroom{
			ClassroomId:   classroom.ClassroomID,
			ClassroomName: classroom.ClassroomName,
			RoomArea:      classroom.ClassroomArea,
		})
	}

	return &lpb.RetrieveLessonsOnCalendarResponse_Lesson{
		TeachingMethod:   cpb.LessonTeachingMethod(cpb.LessonTeachingMethod_value[string(l.TeachingMethod)]),
		LessonId:         l.LessonID,
		LessonName:       l.Name,
		StartTime:        timestamppb.New(l.StartTime),
		EndTime:          timestamppb.New(l.EndTime),
		LessonTeachers:   lessonTeachers,
		CourseId:         l.CourseID,
		CourseName:       l.CourseName,
		ClassId:          l.ClassID,
		ClassName:        l.ClassName,
		LessonMembers:    lessonLearners,
		GroupGrade:       lessonGroupGrade,
		SchedulingStatus: cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[string(l.SchedulingStatus)]),
		LessonClassrooms: lessonClassrooms,
		SchedulerId:      l.SchedulerID,
		LessonCapacity:   uint32(l.LessonCapacity),
	}
}

func toLessonPb(l *domain.Lesson) *lpb.RetrieveLessonsResponse_Lesson {
	return &lpb.RetrieveLessonsResponse_Lesson{
		Id:               l.LessonID,
		Name:             l.Name,
		CenterId:         l.LocationID,
		StartTime:        timestamppb.New(l.StartTime),
		EndTime:          timestamppb.New(l.EndTime),
		TeacherIds:       l.GetTeacherIDs(),
		TeachingMethod:   cpb.LessonTeachingMethod(cpb.LessonTeachingMethod_value[string(l.TeachingMethod)]),
		TeachingMedium:   cpb.LessonTeachingMedium(cpb.LessonTeachingMedium_value[string(l.TeachingMedium)]),
		CourseId:         l.CourseID,
		ClassId:          l.ClassID,
		SchedulingStatus: cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[string(l.SchedulingStatus)]),
		LessonCapacity:   uint32(l.LessonCapacity),
		EndAt: func() *timestamppb.Timestamp {
			if l.EndAt == nil {
				return nil
			}
			return timestamppb.New(*l.EndAt)
		}(),
		ZoomLink:    l.ZoomLink,
		ClassDoLink: l.ClassDoLink,
	}
}

func parseSecondsToTime(total int64) string {
	hours := total / 3600
	minutes := (total % 3600) / 60
	seconds := total % 60
	if total < 0 {
		hours = 0
		minutes = 0
		seconds = 0
	}
	return fmt.Sprintf(`%02d:%02d:%02d`, hours, minutes, seconds)
}

func filterArgsFromRequestPayload(ctx context.Context, req *lpb.RetrieveLessonsRequest) (args payloads.GetLessonListArg, isEmptyIntersectedLocation bool) {
	args = payloads.GetLessonListArg{
		LessonTime: "future",
		Compare:    ">=",
		Limit:      req.Paging.GetLimit(),
		LessonID:   req.GetPaging().GetOffsetString(),
		SchoolID:   golibs.ResourcePathFromCtx(ctx),
		CourseIDs:  req.GetFilter().GetCourseIds(),
		TeacherIDs: req.GetFilter().GetTeacherIds(),
		StudentIDs: req.GetFilter().GetStudentIds(),
		Grades:     req.GetFilter().GetGrades(),
		GradesV2:   req.GetFilter().GetGradesV2(),
		TimeZone:   req.GetFilter().GetTimeZone(),
		KeyWord:    req.GetKeyword(),
		ClassIDs:   req.GetFilter().GetClassIds(),
	}

	locations := req.GetLocationIds()
	filterLocations := req.GetFilter().GetLocationIds()
	locationsLen := len(locations)
	filterLocationsLen := len(filterLocations)

	if locationsLen > 0 && filterLocationsLen > 0 {
		intersect, _, _ := golibs.Compare(locations, filterLocations)
		if len(intersect) == 0 {
			return args, true
		}
		args.LocationIDs = intersect
	} else {
		switch {
		case filterLocationsLen != 0:
			args.LocationIDs = filterLocations
		case locationsLen != 0:
			args.LocationIDs = locations
		}
	}

	if req.GetLessonTime() == lpb.LessonTime_LESSON_TIME_PAST {
		args.LessonTime = "past"
		args.Compare = "<"
	}

	if fromTime := req.GetFilter().GetFromTime(); fromTime != nil {
		totalSeconds := req.GetFilter().GetFromTime().Seconds
		args.FromTime = parseSecondsToTime(totalSeconds)
	}

	if toTime := req.GetFilter().GetToTime(); toTime != nil {
		totalSeconds := req.GetFilter().GetToTime().Seconds
		args.ToTime = parseSecondsToTime(totalSeconds)
	}

	if dows := req.GetFilter().GetDateOfWeeks(); len(dows) > 0 {
		valDows := make([]domain.DateOfWeek, len(dows))
		for i, s := range dows {
			valDows[i] = domain.DateOfWeek(cpb.DateOfWeek_value[s.String()])
		}
		args.Dow = valDows
	}

	if fromDate := req.Filter.GetFromDate(); fromDate != nil {
		args.FromDate = fromDate.AsTime()
	}

	if toDate := req.Filter.GetToDate(); toDate != nil {
		args.ToDate = toDate.AsTime()
	}

	if currentTime := req.GetCurrentTime(); currentTime != nil {
		args.CurrentTime = currentTime.AsTime()
	}

	if status := req.GetFilter().GetSchedulingStatus(); len(status) > 0 {
		lessonStatuses := make([]domain.LessonSchedulingStatus, 0, len(status))
		for _, v := range status {
			lessonStatuses = append(lessonStatuses, domain.LessonSchedulingStatus(v.String()))
		}
		args.LessonSchedulingStatuses = lessonStatuses
	}

	if courseTypeIDs := req.GetFilter().GetCourseTypeIds(); len(courseTypeIDs) > 0 {
		args.CourseTypesIDs = courseTypeIDs
	}

	if reportStatus := req.GetFilter().GetLessonReportStatus(); len(reportStatus) > 0 {
		reportStatuses := make([]lrd.LessonReportStatus, 0, len(reportStatus))
		for _, r := range reportStatus {
			switch r {
			case lpb.RetrieveLessonsFilter_LESSON_REPORT_STATUS_NONE:
				reportStatuses = append(reportStatuses, lrd.ReportStatusNone)
			case lpb.RetrieveLessonsFilter_LESSON_REPORT_STATUS_DRAFT:
				reportStatuses = append(reportStatuses, lrd.ReportStatusDraft)
			case lpb.RetrieveLessonsFilter_LESSON_REPORT_STATUS_SUBMITTED:
				reportStatuses = append(reportStatuses, lrd.ReportStatusSubmitted)
			}
		}
		args.LessonReportStatus = reportStatuses
	}

	return args, false
}

func filterArgsFromRequest(ctx context.Context, req *lpb.RetrieveLessonsRequest) (args domain.ListLessonArgs, isExisted bool) {
	args = domain.ListLessonArgs{
		KeyWord:  "",
		LessonID: "",

		Limit:      req.Paging.Limit,
		SchoolID:   "",
		LessonTime: "future",
		Compare:    ">=",

		TimeZone: "",
		FromTime: "",
		ToTime:   "",
	}

	locations := req.GetLocationIds()
	filterLocations := req.GetFilter().GetLocationIds()
	locationsLen := len(locations)
	filterLocationsLen := len(filterLocations)

	if locationsLen > 0 && filterLocationsLen > 0 {
		intersect, _, _ := golibs.Compare(locations, filterLocations)
		if len(intersect) == 0 {
			return args, false
		}
		args.LocationIDs = intersect
	} else {
		switch {
		case filterLocationsLen != 0:
			args.LocationIDs = filterLocations
		case locationsLen != 0:
			args.LocationIDs = locations
		}
	}

	if req.GetLessonTime() == lpb.LessonTime_LESSON_TIME_PAST {
		args.LessonTime = "past"
		args.Compare = "<"
	}

	if courses := req.GetFilter().GetCourseIds(); len(courses) > 0 {
		args.CourseIDs = courses
	}

	if fromTime := req.GetFilter().GetFromTime(); fromTime != nil {
		totalSeconds := req.GetFilter().GetFromTime().Seconds
		args.FromTime = parseSecondsToTime(totalSeconds)
	}

	if toTime := req.GetFilter().GetToTime(); toTime != nil {
		totalSeconds := req.GetFilter().GetToTime().Seconds
		args.ToTime = parseSecondsToTime(totalSeconds)
	}

	if teachers := req.GetFilter().GetTeacherIds(); len(teachers) > 0 {
		args.TeacherIDs = teachers
	}

	if students := req.GetFilter().GetStudentIds(); len(students) > 0 {
		args.StudentIDs = students
	}

	if fromDate := req.GetFilter().GetFromDate(); fromDate != nil {
		args.FromDate = fromDate.AsTime()
	}

	if toDate := req.GetFilter().GetToDate(); toDate != nil {
		args.ToDate = toDate.AsTime()
	}

	if currentTime := req.GetCurrentTime(); currentTime != nil {
		args.CurrentTime = currentTime.AsTime()
	}

	if keyWord := req.GetKeyword(); keyWord != "" {
		args.KeyWord = keyWord
	}

	if grades := req.GetFilter().GetGrades(); len(grades) > 0 {
		args.Grades = grades
	}

	if gradesV2 := req.GetFilter().GetGradesV2(); len(gradesV2) > 0 {
		args.GradesV2 = gradesV2
	}

	if timeZone := req.GetFilter().GetTimeZone(); timeZone != "" {
		args.TimeZone = timeZone
	}

	if dows := req.GetFilter().GetDateOfWeeks(); len(dows) > 0 {
		valDows := make([]domain.DateOfWeek, len(dows))
		for i, s := range dows {
			valDows[i] = domain.DateOfWeek(cpb.DateOfWeek_value[s.String()])
		}
		args.Dow = valDows
	}

	if req.GetPaging().GetOffsetString() != "" {
		args.LessonID = req.Paging.GetOffsetString()
	}

	args.SchoolID = golibs.ResourcePathFromCtx(ctx)

	return args, true
}

func (l *LessonReaderService) RetrieveStudentsByLesson(ctx context.Context, req *lpb.ListStudentsByLessonRequest) (*lpb.ListStudentsByLessonResponse, error) {
	args := &domain.ListStudentsByLessonArgs{
		LessonID: req.LessonId,
		Limit:    10,
		UserName: "",
		UserID:   "",
	}
	if paging := req.Paging; paging != nil {
		if limit := paging.Limit; 1 <= limit && limit <= 100 {
			args.Limit = limit
		}
		if c := paging.GetOffsetMultipleCombined(); c != nil {
			args.UserName = c.Combined[0].OffsetString
			args.UserID = c.Combined[1].OffsetString
		}
	}

	students, err := l.retrieveLessonCommand.RetrieveLessonMembersByLessonArgs(ctx, args)
	if err != nil {
		return nil, err
	}
	if len(students) == 0 {
		return &lpb.ListStudentsByLessonResponse{}, nil
	}

	pbStudents := make([]*cpb.BasicProfile, 0, len(students))
	for _, student := range students {
		pbStudents = append(pbStudents, toCommonBasicProfile(student))
	}

	lastItem := students[len(students)-1]
	nextPage := &cpb.Paging{
		Limit: args.Limit,
		Offset: &cpb.Paging_OffsetMultipleCombined{
			OffsetMultipleCombined: &cpb.Paging_MultipleCombined{
				Combined: []*cpb.Paging_Combined{
					{
						OffsetString: lastItem.GetName(),
					},
					{
						OffsetString: lastItem.ID,
					},
				},
			},
		},
	}

	return &lpb.ListStudentsByLessonResponse{
		Students: pbStudents,
		NextPage: nextPage,
	}, nil
}

func toCommonBasicProfile(e *domain.User) *cpb.BasicProfile {
	basicProfile := &cpb.BasicProfile{
		UserId:      e.ID,
		Name:        e.GetName(),
		Avatar:      e.Avatar,
		Group:       cpb.UserGroup(cpb.UserGroup_value[e.Group]),
		FacebookId:  e.FacebookID,
		AppleUserId: e.AppleUser.ID,
	}
	if !e.LastLoginDate.IsZero() {
		basicProfile.LastLoginDate = timestamppb.New(e.LastLoginDate)
	}
	return basicProfile
}

func (l *LessonReaderService) RetrieveLessonMedias(ctx context.Context, req *lpb.ListLessonMediasRequest) (*lpb.ListLessonMediasResponse, error) {
	limit := uint32(req.Paging.Limit)
	var offset string
	if len(req.Paging.GetOffsetString()) > 0 {
		offset = req.Paging.GetOffsetString()
	}
	medias, err := l.retrieveLessonCommand.RetrieveMediasByLessonArgs(ctx, &domain.ListMediaByLessonArgs{
		LessonID: req.LessonId,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}
	if len(medias) == 0 {
		return &lpb.ListLessonMediasResponse{}, nil
	}

	mediasPb := make([]*lpb.Media, 0, len(medias))
	for _, media := range medias {
		mediaPb := toMediaLpb(media)
		mediasPb = append(mediasPb, mediaPb)
	}

	resp := &lpb.ListLessonMediasResponse{
		Items: mediasPb,
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: medias[len(medias)-1].ID,
			},
		},
	}

	return resp, nil
}

func toMediaLpb(src *media_domain.Media) *lpb.Media {
	comments := toCommentsLpb(src.Comments)
	pbImages := make([]*lpb.ConvertedImage, 0, len(src.ConvertedImages))
	for _, c := range src.ConvertedImages {
		pbImages = append(pbImages, &lpb.ConvertedImage{
			Width:    c.Width,
			Height:   c.Height,
			ImageUrl: c.ImageURL,
		})
	}

	return &lpb.Media{
		MediaId:   src.ID,
		Name:      src.Name,
		Resource:  src.Resource,
		CreatedAt: timestamppb.New(src.CreatedAt),
		UpdatedAt: timestamppb.New(src.UpdatedAt),
		Comments:  comments,
		Type:      lpb.MediaType(lpb.MediaType_value[string(src.Type)]),
		Images:    pbImages,
	}
}

func toCommentsLpb(comments []media_domain.Comment) []*lpb.Comment {
	dst := make([]*lpb.Comment, 0, len(comments))
	for _, comment := range comments {
		dst = append(dst, &lpb.Comment{
			Comment: comment.Comment,
		})
	}
	return dst
}
