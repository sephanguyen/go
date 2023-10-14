package eibanam

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	bentities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/enigma/dto"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Helper) InsertJprepSchool() error {
	random := idutil.ULIDNow()
	sch := &bentities.School{
		ID:             database.Int4(constants.JPREPSchool),
		Name:           database.Text(random),
		Country:        database.Text(constant.CountryVN),
		IsSystemSchool: database.Bool(false),
		CreatedAt:      database.Timestamptz(time.Now()),
		UpdatedAt:      database.Timestamptz(time.Now()),
		Point: pgtype.Point{
			P:      pgtype.Vec2{X: 0, Y: 0},
			Status: 2,
		},
	}

	city := &bentities.City{
		Name:         database.Text(random),
		Country:      database.Text(constant.CountryVN),
		CreatedAt:    database.Timestamptz(time.Now()),
		UpdatedAt:    database.Timestamptz(time.Now()),
		DisplayOrder: database.Int2(0),
	}

	district := &bentities.District{
		Name:    database.Text(random),
		Country: database.Text(constant.CountryVN),
		City:    city,
	}
	sch.City = city
	sch.District = district
	repo := &repositories.SchoolRepo{}
	ctx := context.Background()
	err := repo.Import(ctx, s.bobDB, []*bentities.School{sch})
	if err != nil {
		return err
	}
	if err = s.generateOrganizationAuth(ctx, constants.JPREPSchool); err != nil {
		return err
	}
	return nil
}

// copy from enigma
func (h *Helper) TranslateJprepLessonID(lessonID int) string {
	return fmt.Sprintf("JPREP_LESSON_%09d", lessonID)
}

func (h *Helper) JprepCreateLessonForStudents(student []string) (*dto.Lesson, error) {
	courseID, err := h.JprepSyncACourse()
	if err != nil {
		return nil, err
	}
	lesson, err := h.JprepSyncALesson(courseID)
	if err != nil {
		return nil, err
	}
	err = h.JprepSyncStudentToLesson(student, []int{lesson.LessonID})
	if err != nil {
		return nil, err
	}
	return lesson, nil
}

func (h *Helper) CreateLesson(adminToken string, teachers []string, students []string, schoolID int32) (*bpb.CreateLiveLessonRequest, *bpb.CreateLiveLessonResponse, error) {
	courseNum := 1

	// create course
	courses := []string{}
	if courseNum > 0 {
		coursesReq, err := h.CreateACourseViaGRPC(adminToken, schoolID)
		if err != nil {
			return nil, nil, err
		}
		courses = append(courses, coursesReq.Courses[0].Id)
	}

	now := time.Now()
	now = now.Add(-time.Duration(now.Nanosecond()) * time.Nanosecond)
	req := &bpb.CreateLiveLessonRequest{
		Name:       "lesson name " + idutil.ULIDNow(),
		StartTime:  timestamppb.New(now.Add(-1 * time.Hour)),
		EndTime:    timestamppb.New(now.Add(1 * time.Hour)),
		TeacherIds: teachers,
		CourseIds:  courses,
		LearnerIds: students,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ctx = ContextWithTokenForGrpcCall(ctx, adminToken)
	res, err := bpb.NewLessonModifierServiceClient(h.BobConn).
		CreateLiveLesson(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	return req, res, nil
}

func (s *Helper) JprepSyncStudentToLesson(students []string, lessons []int) error {
	studentLessons := make([]dto.StudentLesson, 0, len(students))
	for i := 0; i < len(students); i++ {
		studentLessons = append(studentLessons, dto.StudentLesson{
			StudentID:  students[i],
			LessonIDs:  lessons,
			ActionKind: dto.ActionKindUpserted,
		})
	}
	request := dto.SyncUserCourseRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			StudentLessons []dto.StudentLesson `json:"student_lesson"`
		}{
			StudentLessons: studentLessons,
		},
	}
	data, err := json.Marshal(request)
	if err != nil {
		return err
	}
	sig, err := s.generateSignature(s.JPREPKey, string(data))
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/jprep/user-course", s.EnigmaSrvURL)
	bodyBytes, err := s.makeJPREPHTTPRequest(request, sig, http.MethodPut, url)
	if err != nil {
		return err
	}

	if bodyBytes == nil {
		return fmt.Errorf("body is nil")
	}
	return nil
}

func (s *Helper) JprepSyncALesson(courseID int) (*dto.Lesson, error) {
	now := time.Now()
	lessonID := rand.Intn(1000)
	lessonGroup := idutil.ULIDNow()
	lesson := dto.Lesson{
		ActionKind:    dto.ActionKindUpserted,
		LessonID:      lessonID,
		LessonType:    "online",
		CourseID:      courseID,
		StartDatetime: int(now.Unix()),
		EndDatetime:   int(now.Add(1 * time.Hour).Unix()),
		ClassName:     "class name " + idutil.ULIDNow(),
		Week:          lessonGroup,
	}
	request := &dto.MasterRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Courses       []dto.Course       `json:"m_course_name"`
			Classes       []dto.Class        `json:"m_regular_course"`
			Lessons       []dto.Lesson       `json:"m_lesson"`
			AcademicYears []dto.AcademicYear `json:"m_academic_year"`
		}{
			Lessons: []dto.Lesson{
				lesson,
			},
		},
	}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	sig, err := s.generateSignature(s.JPREPKey, string(data))
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/jprep/master-registration", s.EnigmaSrvURL)
	bodyBytes, err := s.makeJPREPHTTPRequest(request, sig, http.MethodPut, url)
	if err != nil {
		return nil, err
	}

	if bodyBytes == nil {
		return nil, fmt.Errorf("body is nil")
	}
	return &lesson, nil
}

func (s *Helper) JprepSyncACourse() (int, error) {
	start := time.Now().Format("2006/01/02")
	end := time.Now().Add(1 * time.Hour).Format("2006/01/02")
	courseID := rand.Intn(1000)
	request := &dto.MasterRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Courses       []dto.Course       `json:"m_course_name"`
			Classes       []dto.Class        `json:"m_regular_course"`
			Lessons       []dto.Lesson       `json:"m_lesson"`
			AcademicYears []dto.AcademicYear `json:"m_academic_year"`
		}{
			Classes: []dto.Class{
				{
					ActionKind:     dto.ActionKindUpserted,
					ClassName:      "class name " + idutil.ULIDNow(),
					ClassID:        rand.Intn(999999999),
					CourseID:       courseID,
					StartDate:      start,
					EndDate:        end,
					AcademicYearID: 0,
				},
			},
			Courses: []dto.Course{
				{
					ActionKind:         dto.ActionKindUpserted,
					CourseID:           courseID,
					CourseName:         "course-name-with-actionKind-upsert",
					CourseStudentDivID: dto.CourseIDKid,
				},
			},
		},
	}

	data, err := json.Marshal(request)
	if err != nil {
		return 0, err
	}
	sig, err := s.generateSignature(s.JPREPKey, string(data))
	if err != nil {
		return 0, err
	}

	url := fmt.Sprintf("%s/jprep/master-registration", s.EnigmaSrvURL)
	bodyBytes, err := s.makeJPREPHTTPRequest(request, sig, http.MethodPut, url)
	if err != nil {
		return 0, err
	}

	if bodyBytes == nil {
		return 0, fmt.Errorf("body is nil")
	}
	return courseID, nil
}
func (s *Helper) makeJPREPHTTPRequest(request interface{}, sig, method, url string) ([]byte, error) {
	bodyRequest, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyRequest))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("JPREP-Signature", sig)
	if err != nil {
		return nil, err
	}

	client := http.Client{Timeout: time.Duration(30) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil
	}
	return body, nil
}
func (s *Helper) generateSignature(key, message string) (string, error) {
	sig := hmac.New(sha256.New, []byte(key))
	if _, err := sig.Write([]byte(message)); err != nil {
		return "", err
	}
	return hex.EncodeToString(sig.Sum(nil)), nil
}
func (s *Helper) generateOrganizationAuth(ctx context.Context, schoolID int32) error {
	query := fmt.Sprintf(`
	INSERT INTO organization_auths
		(organization_id, auth_project_id, auth_tenant_id)
	VALUES
		($1, 'fake_aud', ''),
		($1, 'dev-manabie-online', ''), 
		($1, 'dev-manabie-online', 'integration-test-1-909wx')
	ON CONFLICT 
		DO NOTHING
	;
	`)

	_, err := s.bobDB.Exec(ctx, query, schoolID)
	return err
}
