package queries

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	master_data_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	user_infras "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure/repo"
)

type ExportUserHandler struct {
	WrapperConnection       *support.WrapperDBConnection
	TeacherRepo             user_infras.TeacherRepo
	UserBasicInfoRepo       user_infras.UserBasicInfoRepo
	LocationRepo            master_data_domain.LocationRepository
	CourseRepo              master_data_domain.CourseRepository
	StudentSubscriptionRepo user_infras.StudentSubscriptionRepo

	UnleashClient unleashclient.ClientInstance
	Env           string
}

func (e *ExportUserHandler) ExportTeacher(ctx context.Context) (data []byte, err error) {
	conn, err := e.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	teachers, err := e.TeacherRepo.ListByGrantedLocation(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("e.TeacherRepo.ListByGrantedLocation: %w", err)
	}
	teacherID := make([]string, len(teachers))
	var uniqueGrantedLocationID []string
	grantedLocationMap := map[string]bool{}
	for tID, grantedLocationID := range teachers {
		teacherID = append(teacherID, tID)
		for i := 0; i < len(grantedLocationID); i++ {
			locationID := grantedLocationID[i]
			if !grantedLocationMap[locationID] {
				uniqueGrantedLocationID = append(uniqueGrantedLocationID, locationID)
				grantedLocationMap[locationID] = true
			}
		}
	}
	userBasicInfo, err := e.UserBasicInfoRepo.GetUser(ctx, conn, teacherID)
	if err != nil {
		return nil, fmt.Errorf("e.UserRepo.GetStudentCurrentGradeByUserIDs: %w", err)
	}
	userBasicInfoMap := map[string]*repo.UserBasicInfo{}
	for _, user := range userBasicInfo {
		userBasicInfoMap[user.UserID.String] = user
	}

	locationDetail, err := e.LocationRepo.GetLocationByID(ctx, conn, uniqueGrantedLocationID)
	if err != nil {
		return nil, fmt.Errorf("e.LocationRepo.GetLocationByID: %w", err)
	}
	locationMap := map[string]*master_data_domain.Location{}
	for _, location := range locationDetail {
		locationMap[location.LocationID.String] = location
	}
	teacherData := [][]string{}
	for teacherID, grantedLocations := range teachers {
		for _, locationID := range grantedLocations {
			var teacherName, locationName, partnerInternalID string
			if teacher, exist := userBasicInfoMap[teacherID]; exist {
				teacherName = teacher.FullName.String
			}
			if location, exist := locationMap[locationID]; exist {
				locationName = location.Name.String
				partnerInternalID = location.PartnerInternalID.String
			}
			line := []string{
				teacherID, teacherName, partnerInternalID, locationID, locationName,
			}
			teacherData = append(teacherData, line)
		}
	}
	sort.Sort(support.SliceOfSlice{
		Data:         teacherData,
		IndexCompare: []int{2, 1}, // sort by partner_internal_id,teacher name asc
	})
	title := []string{"teacher_id", "teacher_name", "partner_internal_id", "granted_location_id", "location_name"}
	csvData := append([][]string{title}, teacherData...)
	return exporter.ToCSV(csvData), nil
}

func (e *ExportUserHandler) ExportEnrolledStudent(ctx context.Context, timezone string) (data []byte, err error) {
	conn, err := e.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	enrolledStudent, err := e.StudentSubscriptionRepo.GetAll(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("e.StudentSubscriptionRepo.GetAll: %w", err)
	}
	studentID := make([]string, 0, len(enrolledStudent))
	var uniqueLocationID, uniqueCourseID []string
	grantedLocationMap := map[string]bool{}
	courseUniqueMap := map[string]bool{}
	for _, es := range enrolledStudent {
		studentID = append(studentID, es.StudentID)
		if !grantedLocationMap[es.LocationID] {
			uniqueLocationID = append(uniqueLocationID, es.LocationID)
			grantedLocationMap[es.LocationID] = true
		}
		if !courseUniqueMap[es.CourseID] {
			uniqueCourseID = append(uniqueCourseID, es.CourseID)
			courseUniqueMap[es.CourseID] = true
		}
	}
	userBasicInfo, err := e.UserBasicInfoRepo.GetUser(ctx, conn, studentID)
	if err != nil {
		return nil, fmt.Errorf("e.UserRepo.GetStudentCurrentGradeByUserIDs: %w", err)
	}
	userBasicInfoMap := map[string]*repo.UserBasicInfo{}
	for _, user := range userBasicInfo {
		userBasicInfoMap[user.UserID.String] = user
	}

	locationDetail, err := e.LocationRepo.GetLocationByID(ctx, conn, uniqueLocationID)
	if err != nil {
		return nil, fmt.Errorf("e.LocationRepo.GetLocationByID: %w", err)
	}
	locationMap := map[string]*master_data_domain.Location{}
	for _, location := range locationDetail {
		locationMap[location.LocationID.String] = location
	}

	courses, err := e.CourseRepo.GetByIDs(ctx, conn, uniqueCourseID)
	if err != nil {
		return nil, fmt.Errorf("e.LocationRepo.GetLocationByID: %w", err)
	}
	courseMap := map[string]*master_data_domain.Course{}
	for _, course := range courses {
		courseMap[course.CourseID.String] = course
	}

	studentData := [][]string{}
	loc, _ := time.LoadLocation(timezone)
	for _, es := range enrolledStudent {
		var studentName, locationName, partnerInternalID, courseName string
		if student, exists := userBasicInfoMap[es.StudentID]; exists {
			studentName = student.FullName.String
		}
		if location, exists := locationMap[es.LocationID]; exists {
			locationName = location.Name.String
			partnerInternalID = location.PartnerInternalID.String
		}
		if course, exists := courseMap[es.CourseID]; exists {
			courseName = course.Name.String
		}
		var enrollmentStatus string
		if es.EnrollmentStatus == string(domain.EnrollmentStatusEnrolled) {
			enrollmentStatus = "Enrolled"
		} else {
			enrollmentStatus = "Potential"
		}
		line := []string{
			es.StudentID,
			studentName,
			enrollmentStatus,
			partnerInternalID,
			es.LocationID,
			locationName,
			es.CourseID,
			courseName,
			es.StartAt.In(loc).Format("2006/01/02"),
			es.EndAt.In(loc).Format("2006/01/02"),
		}
		studentData = append(studentData, line)
	}
	sort.Sort(support.SliceOfSlice{
		Data:         studentData,
		IndexCompare: []int{3, 7, 1}, // sort by partner_internal_id, course_name, student_name asc
	})
	title := []string{"student_id", "student_name", "student_status", "partner_internal_id", "granted_location_id", "location_name", "course_id", "course_name", "course_start_date", "course_end_date"}
	csvData := append([][]string{title}, studentData...)
	return exporter.ToCSV(csvData), nil
}
