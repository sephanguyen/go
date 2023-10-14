package jprep

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/enigma/dto"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	yasuoRepo "github.com/manabie-com/backend/internal/yasuo/repositories"

	"github.com/jackc/pgx/v4"
)

func (s *suite) stepARequestNewClassMemberWithStudentPayload() error {
	s.CurrentUserID = idutil.ULIDNow()
	now := time.Now().Format("2006/01/02")
	request := &dto.UserRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Students []dto.Student `json:"m_student"`
			Staffs   []dto.Staff   `json:"m_staff"`
		}{
			Students: []dto.Student{
				{
					ActionKind: dto.ActionKindUpserted,
					StudentID:  s.CurrentUserID,
					LastName:   "Last name " + idutil.ULIDNow(),
					GivenName:  "Given name " + idutil.ULIDNow(),
					StudentDivs: []struct {
						MStudentDivID int `json:"m_student_div_id"`
					}{
						{MStudentDivID: 1},
					},
					Regularcourses: []struct {
						ClassID   int    `json:"m_course_id"`
						Startdate string `json:"startdate"`
						Enddate   string `json:"enddate"`
					}{
						{
							ClassID:   s.CurrentClassID,
							Startdate: now,
							Enddate:   now,
						},
					},
				},
			},
		},
	}

	s.Request = request
	return nil
}

func (s *suite) stepARequestNewClassMemberWithStudentPayloadMissing(missingField string) error {
	s.CurrentUserID = idutil.ULIDNow()
	now := time.Now().Format("2006/01/02")
	request := &dto.UserRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Students []dto.Student `json:"m_student"`
			Staffs   []dto.Staff   `json:"m_staff"`
		}{
			Students: []dto.Student{
				{
					ActionKind: dto.ActionKindUpserted,
					StudentID:  s.CurrentUserID,
					LastName:   "Last name " + idutil.ULIDNow(),
					GivenName:  "Given name " + idutil.ULIDNow(),
					StudentDivs: []struct {
						MStudentDivID int `json:"m_student_div_id"`
					}{
						{MStudentDivID: 1},
					},
					Regularcourses: []struct {
						ClassID   int    `json:"m_course_id"`
						Startdate string `json:"startdate"`
						Enddate   string `json:"enddate"`
					}{
						{
							ClassID:   s.CurrentClassID,
							Startdate: now,
							Enddate:   now,
						},
					},
				},
			},
		},
	}

	switch missingField {
	case "action kind":
		request.Payload.Students[0].ActionKind = ""
	case "studentdivs.mstudentdivid":
		request.Payload.Students[0].StudentDivs[0].MStudentDivID = 0
	case "student id":
		request.Payload.Students[0].StudentID = ""
	case "last name":
		request.Payload.Students[0].LastName = ""
	case "given name":
		request.Payload.Students[0].GivenName = ""
	case "regularcourses.mcourseid":
		request.Payload.Students[0].Regularcourses[0].ClassID = 0
	case "regularcourses.startdate":
		request.Payload.Students[0].Regularcourses[0].Startdate = ""
	case "regularcourses.enddate":
		request.Payload.Students[0].Regularcourses[0].Enddate = ""
	}

	s.Request = request
	return nil
}

func (s *suite) stepARequestNewClassMemberWithStaffPayload() error {
	s.CurrentUserID = idutil.ULIDNow()
	s.CurrentTeacherID = idutil.ULIDNow()

	request := &dto.UserRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Students []dto.Student `json:"m_student"`
			Staffs   []dto.Staff   `json:"m_staff"`
		}{
			Staffs: []dto.Staff{
				{
					ActionKind: dto.ActionKindUpserted,
					StaffID:    s.CurrentTeacherID,
					Name:       "teacher name " + idutil.ULIDNow(),
				},
			},
		},
	}

	s.Request = request
	return nil
}

func (s *suite) stepARequestNewClassMemberWithStaffPayloadMissing(missingField string) error {
	s.CurrentUserID = idutil.ULIDNow()
	s.CurrentTeacherID = idutil.ULIDNow()

	request := &dto.UserRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Students []dto.Student `json:"m_student"`
			Staffs   []dto.Staff   `json:"m_staff"`
		}{
			Staffs: []dto.Staff{
				{
					ActionKind: dto.ActionKindUpserted,
					StaffID:    s.CurrentTeacherID,
					Name:       "teacher name " + idutil.ULIDNow(),
				},
			},
		},
	}

	switch missingField {
	case "action kind":
		request.Payload.Staffs[0].ActionKind = ""
	case "staff id":
		request.Payload.Staffs[0].StaffID = ""
	case "staff name":
		request.Payload.Staffs[0].Name = ""
	}

	s.Request = request
	return nil
}

func (s *suite) stepARequestExistClassMemberWithStudentPayload(actionKind string) error {
	time.Sleep(time.Second * 2)
	now := time.Now().Format("2006/01/02")
	request := &dto.UserRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Students []dto.Student `json:"m_student"`
			Staffs   []dto.Staff   `json:"m_staff"`
		}{
			Students: []dto.Student{
				{
					ActionKind: dto.Action(actionKind),
					StudentID:  s.CurrentUserID,
					LastName:   "Last name " + idutil.ULIDNow(),
					GivenName:  "Given name " + idutil.ULIDNow(),
					StudentDivs: []struct {
						MStudentDivID int `json:"m_student_div_id"`
					}{
						{MStudentDivID: 1},
					},
					Regularcourses: []struct {
						ClassID   int    `json:"m_course_id"`
						Startdate string `json:"startdate"`
						Enddate   string `json:"enddate"`
					}{
						{
							ClassID:   s.CurrentClassID,
							Startdate: now,
							Enddate:   now,
						},
					},
				},
			},
		},
	}

	s.Request = request
	return nil
}

func (s *suite) validateClassMemberResourcePath(ctx context.Context, member *entities.ClassMember) error {
	var resourcePath string
	query := "SELECT resource_path FROM class_members WHERE class_member_id = $1"
	if err := database.Select(ctx, s.bobDBTrace, query, member.ID).ScanFields(&resourcePath); err != nil {
		return err
	}

	if resourcePath != fmt.Sprint(constants.JPREPSchool) {
		return fmt.Errorf("resourcePath does not match, expected: %s, got: %s", fmt.Sprint(constants.JPREPSchool), resourcePath)
	}

	return nil
}

func (s *suite) theTeachersMustBeStoreInOutSystem() error {
	mainProcess := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		count := 0
		query := `SELECT count(*) 
			FROM teachers t, users u, users_groups ug
			WHERE t.teacher_id = $1
			AND t.resource_path = $2 
			AND t.teacher_id = u.user_id
			AND ug.user_id = t.teacher_id
			AND t.deleted_at IS NULL
			AND u.deleted_at IS NULL`

		err := s.bobDB.QueryRow(ctx, query, s.CurrentTeacherID, database.Text(fmt.Sprint(constants.JPREPSchool))).Scan(&count)
		if err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("not found any teachers where teacher_id = %s and resource_path = %d", s.CurrentTeacherID, constants.JPREPSchool)
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) theTeachersMustNotBeStoreInOutSystem() error {
	mainProcess := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		count := 1
		query := `SELECT count(*) 
			FROM teachers
			WHERE teacher_id = $1`

		err := s.bobDB.QueryRow(ctx, query, s.CurrentTeacherID).Scan(&count)
		if err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("teacher must not be store with teacher_id = %s", s.CurrentTeacherID)
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) theStudentsMustBeStoreInOutSystem() error {
	mainProcess := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		student := s.Request.(*dto.UserRegistrationRequest).Payload.Students[0]

		count := 0
		query := `SELECT count(*) 
			FROM students s, users u, users_groups ug
			WHERE s.student_id = $1
			AND s.resource_path = $2
			AND s.student_id = u.user_id
			AND ug.user_id = s.student_id
			AND s.deleted_at IS NULL
			AND u.deleted_at IS NULL`

		if student.ActionKind == dto.ActionKindDeleted {
			query = `SELECT count(*) 
			FROM students s, users u, users_groups ug
			WHERE s.student_id = $1
			AND s.resource_path = $2
			AND s.student_id = u.user_id
			AND ug.user_id = s.student_id
			AND s.deleted_at IS NOT NULL
			AND u.deleted_at IS NOT NULL`
		}

		err := s.bobDB.QueryRow(ctx, query, student.StudentID, database.Text(fmt.Sprint(constants.JPREPSchool))).Scan(&count)
		if err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("not found any students where student_id = %s, resource_path = %d", student.StudentID, constants.JPREPSchool)
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) theStudentsMustNotBeStoreInOutSystem() error {
	mainProcess := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		student := s.Request.(*dto.UserRegistrationRequest).Payload.Students[0]

		count := 1
		query := `SELECT count(*) 
			FROM students
			WHERE student_id = $1`

		err := s.bobDB.QueryRow(ctx, query, student.StudentID).Scan(&count)
		if err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("students must not be store with student_id = %s", student.StudentID)
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) theClassMembersMustBeStoreInOutSystem() error {
	mainProcess := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		classMemberRepo := &repositories.ClassMemberRepo{}
		student := s.Request.(*dto.UserRegistrationRequest).Payload.Students[0]
		status := entities.ClassMemberStatusActive
		if student.ActionKind == dto.ActionKindDeleted {
			status = entities.ClassMemberStatusInactive
		}

		member, err := classMemberRepo.Get(ctx, s.bobDBTrace,
			database.Int4(int32(student.Regularcourses[0].ClassID)),
			database.Text(student.StudentID),
			database.Text(status))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("err find class: %w", err)
		}

		if member == nil {
			return fmt.Errorf("not found member in class")
		}

		err = s.validateClassMemberResourcePath(ctx, member)
		if err != nil {
			return err
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) tomMustRecordMessageJoinClassOfCurrentUser(message string) error {
	mainProcess := func() error {
		query := `SELECT count(message_id) 
				FROM messages 
				WHERE conversation_id = $1 
				AND message = $2 
				AND type = $3 
				AND user_id = $4
				AND deleted_at IS NULL`
		rows, err := s.tomDB.Query(context.Background(), query, s.ConversationID, message, "MESSAGE_TYPE_SYSTEM", s.CurrentUserID)
		if err != nil {
			return err
		}
		defer rows.Close()
		var count int
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err
			}
		}

		if count == 0 {
			return fmt.Errorf("tom cannot create message with conversation_id = %v, message =%v, type = %v, user_id = %v", s.ConversationID, message, "MESSAGE_TYPE_SYSTEM", s.CurrentUserID)
		}

		return nil
	}
	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) eurekaMustStoreClassMemberWithAction(action string) error {
	mainProcess := func() error {
		query := `SELECT COUNT(student_id) 
			FROM class_students 
			WHERE class_id = $1 
			AND student_id = $2 
			AND deleted_at IS NULL`

		if action == "deleted" {
			query = `SELECT COUNT(student_id) 
				FROM class_students 
				WHERE class_id = $1 
				AND student_id = $2 
				AND deleted_at IS NOT NULL`
		}

		rows, err := s.eurekaDB.Query(context.Background(), query, strconv.Itoa(int(s.CurrentClassID)), s.CurrentUserID)
		if err != nil {
			return err
		}
		defer rows.Close()

		var count int
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err
			}
		}

		if count == 0 {
			return fmt.Errorf("not found any class_students where class_id = %d and student_id = %s and action = %s", s.CurrentClassID, s.CurrentUserID, action)
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) eurekaMustStoreCourseStudentsWithAction(action string) error {
	mainProcess := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		courseClassRepo := &yasuoRepo.CourseClassRepo{}
		mapByClass, err := courseClassRepo.FindByClassIDs(
			ctx,
			s.bobDB,
			database.Int4Array([]int32{int32(s.CurrentClassID)}),
		)
		if err != nil {
			return fmt.Errorf("err s.CourseClassRepo.FindByClassIDs: %w", err)
		}
		query := `SELECT COUNT(student_id) 
			FROM course_students 
			WHERE course_id = ANY($1) 
			AND student_id = $2
			AND deleted_at IS NULL`

		if action == "deleted" {
			query = `SELECT COUNT(student_id) 
				FROM course_students 
				WHERE course_id = ANY($1) 
				AND student_id = $2
				AND deleted_at IS NOT NULL`
		}

		rows, err := s.eurekaDB.Query(
			context.Background(),
			query,
			mapByClass[database.Int4(int32(s.CurrentClassID))],
			s.CurrentUserID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		var count int
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err
			}
		}

		if count == 0 {
			return fmt.Errorf("not found any course_students where course_id = %d and student_id = %s and action = %s", s.CurrentClassID, s.CurrentUserID, action)
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}
