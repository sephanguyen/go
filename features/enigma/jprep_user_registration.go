package enigma

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/enigma/dto"
	entities_enigma "github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *suite) aRequestWithStudentAndStaffInvalid(ctx context.Context, student, staff int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.CurrentUserID = idutil.ULIDNow()
	// student, miss studentID
	students := make([]dto.Student, 0, student)
	for i := 1; i <= student; i++ {
		students = append(students, dto.Student{
			ActionKind: dto.ActionKindUpserted,
			LastName:   "Last name " + idutil.ULIDNow(),
			GivenName:  "Given name " + idutil.ULIDNow(),
		})
	}
	// staff, miss staffID
	staffs := make([]dto.Staff, 0, staff)
	for i := 1; i <= staff; i++ {
		staffs = append(staffs, dto.Staff{
			ActionKind: dto.ActionKindUpserted,
			Name:       "Name " + idutil.ULIDNow(),
		})
	}
	request := &dto.UserRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Students []dto.Student `json:"m_student"`
			Staffs   []dto.Staff   `json:"m_staff"`
		}{
			Students: students,
			Staffs:   staffs,
		},
	}

	s.Request = request
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aRequestWithStudentAndStaff(ctx context.Context, student, staff int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.CurrentUserID = idutil.ULIDNow()
	// student
	students := make([]dto.Student, 0, student)
	for i := 1; i <= student; i++ {
		students = append(students, dto.Student{
			ActionKind: dto.ActionKindUpserted,
			StudentID:  strconv.Itoa(i),
			LastName:   "Last name " + idutil.ULIDNow(),
			GivenName:  "Given name " + idutil.ULIDNow(),
		})
	}
	// staff
	staffs := make([]dto.Staff, 0, staff)
	for i := 1; i <= staff; i++ {
		staffs = append(staffs, dto.Staff{
			ActionKind: dto.ActionKindUpserted,
			StaffID:    strconv.Itoa(i),
			Name:       "Name " + idutil.ULIDNow(),
		})
	}
	request := &dto.UserRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Students []dto.Student `json:"m_student"`
			Staffs   []dto.Staff   `json:"m_staff"`
		}{
			Students: students,
			Staffs:   staffs,
		},
	}

	s.Request = request

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aRequestWithStudentAndStaffInvalidPayload(ctx context.Context, student, staff int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.CurrentUserID = idutil.ULIDNow()
	// student
	students := make([]dto.Student, 0, student)
	for i := 1; i <= student; i++ {
		students = append(students, dto.Student{
			ActionKind: dto.ActionKindUpserted,
			LastName:   "Last name " + idutil.ULIDNow(),
			GivenName:  "Given name " + idutil.ULIDNow(),
		})
	}
	// staff
	staffs := make([]dto.Staff, 0, staff)
	for i := 1; i <= staff; i++ {
		staffs = append(staffs, dto.Staff{
			ActionKind: dto.ActionKindUpserted,
			Name:       "Name " + idutil.ULIDNow(),
		})
	}
	request := &dto.UserRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Students []dto.Student `json:"m_student"`
			Staffs   []dto.Staff   `json:"m_staff"`
		}{
			Students: students,
			Staffs:   staffs,
		},
	}

	s.Request = request

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) stepAValidJPREPSignatureInItsHeader(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	data, err := json.Marshal(s.Request)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	sig, err := s.generateSignature(s.JprepKey, string(data))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	s.JPREPSignature = sig
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateSignature(key, message string) (string, error) {
	sig := hmac.New(sha256.New, []byte(key))
	if _, err := sig.Write([]byte(message)); err != nil {
		return "", err
	}
	return hex.EncodeToString(sig.Sum(nil)), nil
}

func (s *suite) stepPerformUserRegistrationRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	url := fmt.Sprintf("%s/jprep/user-registration", s.EnigmaSrvURL)
	bodyBytes, err := s.makeHTTPRequest(http.MethodPut, url)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if bodyBytes == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("body is nil")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) setResourcePathToContext(ctx context.Context, schoolID string) context.Context {
	stepState := StepStateFromContext(ctx)
	claim := interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: schoolID,
			DefaultRole:  entities.UserGroupAdmin,
			UserGroup:    entities.UserGroupAdmin,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, &claim)
	return StepStateToContext(ctx, stepState)
}

func (s *suite) aPartnerDataSyncNotExistInDB(ctx context.Context, schoolID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.setResourcePathToContext(ctx, schoolID)
	row := s.BobDB.QueryRow(ctx, `SELECT count(*) FROM public.partner_sync_data_log p WHERE p.signature = $1 `, s.JPREPSignature)
	var count int
	if err := row.Scan(&count); err != nil {
		return ctx, err
	}
	if count > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Expect partner_sync_data_log not found")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aPartnerDataSyncAlreadyExistInDB(ctx context.Context, schoolID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.setResourcePathToContext(ctx, schoolID)
	row := s.BobDB.QueryRow(ctx, `SELECT partner_sync_data_log_id FROM public.partner_sync_data_log p WHERE p.signature = $1 `, s.JPREPSignature)
	var logID string
	if err := row.Scan(&logID); err != nil {
		return ctx, err
	}
	if len(logID) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Expect partner_sync_data_log exist, but got empty")
	}
	stepState.PartnerSyncLogID = logID

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aPartnerDataSyncSplitAlreadyExistInDB(ctx context.Context, schoolID string, n_logs int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.setResourcePathToContext(ctx, schoolID)
	row := s.BobDB.QueryRow(ctx, `SELECT count(*) from partner_sync_data_log_split where partner_sync_data_log_id = $1 `, stepState.PartnerSyncLogID)
	var total int
	if err := row.Scan(&total); err != nil {
		return ctx, err
	}
	if total != n_logs {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Expect %d partner_sync_data_log_split exist, but got %d", n_logs, total)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aPayloadLogsMatchWithRequest(ctx context.Context, schoolID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.setResourcePathToContext(ctx, schoolID)
	var logs []*entities_enigma.PartnerSyncDataLogSplit
	stmt := `SELECT partner_sync_data_log_split_id, kind, status, payload from partner_sync_data_log_split where partner_sync_data_log_id = $1`
	rows, err := s.BobDBTrace.Query(
		ctx,
		stmt,
		stepState.PartnerSyncLogID,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "query partner_sync_data_log_split")
	}
	defer rows.Close()
	for rows.Next() {
		e := &entities_enigma.PartnerSyncDataLogSplit{}
		err := rows.Scan(
			&e.PartnerSyncDataLogSplitID,
			&e.Kind,
			&e.Status,
			&e.Payload,
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), errors.WithMessage(err, "rows.Scan partner_sync_data_log_split")
		}
		logs = append(logs, e)
	}
	for _, log := range logs {
		switch log.Kind.String {
		case string(entities_enigma.KindStudent):
			return s.logCorrectStudent(ctx, log.Payload.Bytes)
		case string(entities_enigma.KindStaff):
			return s.logCorrectStaff(ctx, log.Payload.Bytes)
		case string(entities_enigma.KindLesson):
			return s.logCorrectLesson(ctx, log.Payload.Bytes)
		case string(entities_enigma.KindCourse):
			return s.logCorrectCourse(ctx, log.Payload.Bytes)
		case string(entities_enigma.KindClass):
			return s.logCorrectClass(ctx, log.Payload.Bytes)
		case string(entities_enigma.KindAcademicYear):
			return s.logCorrectAcademicYear(ctx, log.Payload.Bytes)
		case string(entities_enigma.KindStudentLessons):
			return s.logCorrectStudentLessons(ctx, log.Payload.Bytes)
		default:
			return StepStateToContext(ctx, stepState), fmt.Errorf("Kind %s invalid", log.Kind.String)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) logCorrectStudent(ctx context.Context, payload []byte) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	students := []*npb.EventUserRegistration_Student{}
	err := json.Unmarshal(payload, &students)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("json.Unmarshal students: %w", err)
	}
	for _, student := range students {
		found := false
		for _, studentReq := range s.Request.(*dto.UserRegistrationRequest).Payload.Students {
			if student.StudentId == studentReq.StudentID {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("can't find student %s", student.StudentId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) logCorrectStaff(ctx context.Context, payload []byte) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	staffs := []*npb.EventUserRegistration_Staff{}
	err := json.Unmarshal(payload, &staffs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("json.Unmarshal staff: %w", err)
	}
	for _, staff := range staffs {
		found := false
		for _, staffReq := range s.Request.(*dto.UserRegistrationRequest).Payload.Staffs {
			if staff.StaffId == staffReq.StaffID {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("can't find staff %s", staff.Name)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedInWithSchool(ctx context.Context, role, school string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	schoolID, _ := strconv.Atoi(school)
	stepState.CurrentSchoolID = int32(schoolID)
	switch role {
	case "school admin":
		return s.aSignedInSchoolAdminWithSchoolID(ctx,
			entities.UserGroupSchoolAdmin, schoolID)
	case "unauthenticated":
		stepState.AuthToken = "random-token"
		return StepStateToContext(ctx, stepState), nil
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedInSchoolAdminWithSchoolID(ctx context.Context, group string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()
	ctx, err := s.aValidSchoolAdminProfileWithId(ctx, id, group, schoolID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken, err = s.generateExchangeToken(id, group)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	stepState.CurrentUserID = id
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidSchoolAdminProfileWithId(ctx context.Context, id, userGroup string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	c := entities.SchoolAdmin{}
	database.AllNullEntity(&c)
	if userGroup == "" {
		userGroup = entities.UserGroupSchoolAdmin
	}

	c.SchoolAdminID.Set(id)
	c.SchoolID.Set(schoolID)
	now := time.Now()
	if err := c.UpdatedAt.Set(now); err != nil {
		return nil, err
	}
	if err := c.CreatedAt.Set(now); err != nil {
		return nil, err
	}

	num := rand.Int()
	u := entities.User{}
	database.AllNullEntity(&u)
	err := multierr.Combine(
		u.ID.Set(c.SchoolAdminID),
		u.LastName.Set(fmt.Sprintf("valid-school-admin-%d", num)),
		u.PhoneNumber.Set(fmt.Sprintf("+848%d", num)),
		u.Email.Set(fmt.Sprintf("valid-school-admin-%d@email.com", num)),
		u.Avatar.Set(fmt.Sprintf("http://valid-school-admin-%d", num)),
		u.Country.Set(pb.COUNTRY_VN.String()),
		u.Group.Set(userGroup),
		u.DeviceToken.Set(nil),
		u.AllowNotification.Set(true),
		u.CreatedAt.Set(c.CreatedAt),
		u.UpdatedAt.Set(c.UpdatedAt),
		u.IsTester.Set(nil),
		u.FacebookID.Set(nil),
	)

	userRepo := repositories.UserRepo{}

	err = userRepo.Create(ctx, s.BobDBTrace, &u)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	schoolAdminRepo := repositories.SchoolAdminRepo{}
	err = schoolAdminRepo.CreateMultiple(ctx, s.BobDBTrace, []*entities.SchoolAdmin{&c})

	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	ug := entities.UserGroup{}
	database.AllNullEntity(&ug)
	err = multierr.Combine(
		ug.UserID.Set(id),
		ug.GroupID.Set(userGroup),
		ug.UpdatedAt.Set(now),
		ug.CreatedAt.Set(now),
		ug.IsOrigin.Set(true),
		ug.Status.Set(entities.UserGroupStatusActive),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	userGroupRepo := repositories.UserGroupRepo{}
	err = userGroupRepo.Upsert(ctx, s.BobDBTrace, &ug)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
