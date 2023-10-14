package fatima

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	fatimaPb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) aValidEvent_Upsert(ctx context.Context) error {
	timeNow := time.Now()
	s.StartAt = &timestamppb.Timestamp{
		Seconds: int64(timeNow.AddDate(0, 0, -1).Unix()),
	}
	s.EndAt = &timestamppb.Timestamp{
		Seconds: int64(timeNow.AddDate(0, 0, 1).Unix()),
	}

	studentID, err := s.insertStudentIntoBob(auth.InjectFakeJwtToken(ctx, fmt.Sprint(constants.ManabieSchool)))
	if err != nil {
		return err
	}

	s.UserID = studentID
	s.CourseIDs = []string{"course-teacher-20", "course-live-teacher-21"}
	event := s.aEvtSyncStudentPackage_Upsert()
	s.Event = event
	return nil
}

func (s *suite) aValidEvent_Delete(ctx context.Context) error {
	err := s.aValidEvent_Upsert(ctx)
	if err != nil {
		return err
	}

	errPublishEvent := s.sendSyncStudentPackageEvent()
	if errPublishEvent != nil {
		return err
	}

	for index := 0; index < len(s.Event.(*npb.EventSyncStudentPackage).StudentPackages); index++ {
		s.Event.(*npb.EventSyncStudentPackage).StudentPackages[index].ActionKind = npb.ActionKind_ACTION_KIND_DELETED
	}

	return nil
}

func (s *suite) aEvtSyncStudentPackage_Upsert() *npb.EventSyncStudentPackage {
	event := &npb.EventSyncStudentPackage{
		StudentPackages: []*npb.EventSyncStudentPackage_StudentPackage{
			{
				StudentId:  s.UserID,
				ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
				Packages: []*npb.EventSyncStudentPackage_Package{
					{
						CourseIds: s.CourseIDs,
						StartDate: s.StartAt,
						EndDate:   s.EndAt,
					},
					{
						CourseIds: s.CourseIDs,
						StartDate: s.StartAt,
						EndDate:   s.EndAt,
					},
				},
			},
		},
	}
	return event
}

func (s *suite) sendSyncStudentPackageEvent() error {
	data, err := proto.Marshal(s.Event.(*npb.EventSyncStudentPackage))
	if err != nil {
		return err
	}

	ctx := context.Background()
	_, err = s.JSM.PublishContext(auth.InjectFakeJwtToken(ctx, fmt.Sprint(constants.ManabieSchool)), constants.SubjectSyncStudentPackage, data)
	if err != nil {
		return fmt.Errorf("sendSyncStudentPackageEvent %w", err)
	}
	return nil
}

func (s *suite) getCourseOfStudent() map[string][]string {
	courses := make(map[string]bool)
	courseOfStudents := make(map[string][]string)
	for _, studentPackage := range s.Event.(*npb.EventSyncStudentPackage).StudentPackages {
		studentId := studentPackage.StudentId
		for _, item := range studentPackage.Packages {
			for _, course := range item.CourseIds {
				if isExist := courses[fmt.Sprintf("%s-%s", studentId, course)]; !isExist {
					courses[fmt.Sprintf("%s-%s", studentId, course)] = true
					courseOfStudents[studentId] = append(courseOfStudents[studentId], course)
				}
			}
		}
	}
	return courseOfStudents
}

func (s *suite) fatimaMustCreateStudentPackage() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	courseOfStudents := s.getCourseOfStudent()
	totalCourses := 0
	studentIds := []string{}
	for studentId, courses := range courseOfStudents {
		studentIds = append(studentIds, studentId)
		totalCourses += len(courses)
	}

	studentPackages := []*entities.StudentPackage{}
	e := &entities.StudentPackage{}
	fields, _ := e.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s 
		FROM student_packages
		WHERE student_id = ANY($1)`, strings.Join(fields, ", "))
	if err := try.Do(func(attempt int) (retry bool, err error) {
		time.Sleep(350 * time.Millisecond)
		rows, err := db.Query(ctx, query, studentIds)
		if err != nil {
			return attempt < 5, err
		}
		defer rows.Close()

		for rows.Next() {
			e := &entities.StudentPackage{}
			err := rows.Scan(database.GetScanFields(e, database.GetFieldNames(e))...)
			if err != nil {
				return attempt < 5, err
			}
			studentPackages = append(studentPackages, e)
		}
		if err != nil {
			return attempt < 5, err
		}
		if len(studentPackages) != len(studentIds) {
			return attempt < 5, fmt.Errorf("fatima does not create studentPackage correctly")
		}

		for _, v := range studentPackages {
			locIDs := database.FromTextArray(v.LocationIDs)
			if !slices.Contains(locIDs, constants.JPREPOrgLocation) {
				return false, fmt.Errorf("fatima does not create studentPackage correctly: wrong location id for JPREP")
			}
		}

		return attempt < 5, err
	}); err != nil {
		return err
	}

	err := s.fatimaMustReturnCourseByStudentID()
	if err != nil {
		return err
	}
	return nil
}

func (s *suite) fatimaSaveStudentPackageAccessPath() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	courseOfStudents := s.getCourseOfStudent()
	totalCourses := 0
	studentIds := []string{}
	for studentId, courses := range courseOfStudents {
		studentIds = append(studentIds, studentId)
		totalCourses += len(courses)
	}

	studentPackageAps := []*entities.StudentPackageAccessPath{}
	e := &entities.StudentPackageAccessPath{}
	fields, _ := e.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s 
		FROM student_package_access_path
		WHERE student_id = ANY($1)`, strings.Join(fields, ", "))
	if err := try.Do(func(attempt int) (retry bool, err error) {
		time.Sleep(350 * time.Millisecond)
		rows, err := db.Query(ctx, query, studentIds)
		if err != nil {
			return attempt < 5, err
		}
		defer rows.Close()

		for rows.Next() {
			e := &entities.StudentPackageAccessPath{}
			err := rows.Scan(database.GetScanFields(e, database.GetFieldNames(e))...)
			if err != nil {
				return attempt < 5, err
			}
			studentPackageAps = append(studentPackageAps, e)
		}
		if err != nil {
			return attempt < 5, err
		}
		for _, v := range studentPackageAps {
			if constants.JPREPOrgLocation != v.LocationID.String {
				return false, fmt.Errorf("fatima does not create studentPackageAccessPath correctly: wrong location id for JPREP")
			}
		}

		return attempt < 5, err
	}); err != nil {
		return err
	}
	return nil
}

func (s *suite) fatimaMustUpdateStudentPackage() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	courseOfStudents := s.getCourseOfStudent()
	totalCourses := 0
	studentIds := []string{}
	for studentId, courses := range courseOfStudents {
		studentIds = append(studentIds, studentId)
		totalCourses += len(courses)
	}

	count := 0
	query := "SELECT count(*) FROM student_packages WHERE student_id = ANY($1) AND is_active = FALSE AND start_at = $2 AND end_at = $3"
	if err := try.Do(func(attempt int) (retry bool, err error) {
		defer func() {
			if retry {
				time.Sleep(1 * time.Second)
			}
		}()
		err = s.DB.QueryRow(ctx, query, studentIds, s.StartAt.AsTime(), s.EndAt.AsTime()).Scan(&count)
		if err != nil {
			return true, err
		}
		if count != len(studentIds) {
			return true, fmt.Errorf("fatimaMustUpdateStudentPackage: update studentPackage expected: %d, but got: %d", count, len(studentIds))
		}

		return attempt < 5, err
	}); err != nil {
		return err
	}

	return nil
}

func (s *suite) fatimaMustReturnCourseByStudentID() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	studentID := s.Event.(*npb.EventSyncStudentPackage).StudentPackages[0].StudentId
	token, err := s.generateExchangeToken(studentID, entity.UserGroupStudent)
	if err != nil {
		return fmt.Errorf("generateExchangeToken: %w", err)
	}

	if err := checkUserExisted(auth.InjectFakeJwtToken(ctx, fmt.Sprint(constants.ManabieSchool)), s.DB, studentID); err != nil {
		return err
	}

	s.AuthToken = token
	courses, err := fatimaPb.NewAccessibilityReadServiceClient(s.Conn).RetrieveAccessibility(contextWithToken(s, ctx), &fatimaPb.RetrieveAccessibilityRequest{})
	if err != nil {
		return err
	}

	for _, courseID := range s.CourseIDs {
		if course := courses.Courses[courseID]; course == nil {
			return fmt.Errorf("can not find course %s of student %s", courseID, studentID)
		}
	}

	return nil
}
