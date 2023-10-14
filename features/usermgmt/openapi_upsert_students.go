package usermgmt

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/crypt"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	grpc_port "github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc"
	http_port "github.com/manabie-com/backend/internal/usermgmt/modules/user/port/http"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"

	"github.com/cucumber/godog"
	"github.com/pkg/errors"
)

func (s *suite) setupAPIKey(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)

	_, err := s.aSignedInStaff(ctx, []string{constant.RoleOpenAPI})
	if err != nil {
		return errors.Wrap(err, "s.aSignedInStaff")
	}
	_, err = s.systemRunJobToGenerateAPIKeyWithOrganization(ctx)
	if err != nil {
		return errors.Wrap(err, "s.systemRunJobToGenerateAPIKeyWithOrganization")
	}

	var publicKey, privateKey field.String

	err = try.Do(func(attempt int) (bool, error) {
		query := `SELECT public_key, private_key FROM api_keypair WHERE user_id = $1 AND resource_path = $2`
		err = database.Select(ctx, s.AuthPostgresDBTrace, query, &stepState.CurrentUserID, &stepState.OrganizationID).ScanFields(&publicKey, &privateKey)
		if err == nil {
			return false, nil
		}
		if attempt < retryTimes {
			time.Sleep(time.Millisecond * 200)
			return true, errors.Wrap(err, "database.Select")
		}
		return false, nil
	})
	if err != nil {
		return err
	}

	data, err := json.Marshal(s.Request)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}

	privateKeyByte, _ := crypt.AESDecryptBase64(privateKey.String(), []byte(AESKey), []byte(AESIV))
	sig := hmac.New(sha256.New, privateKeyByte)
	if _, err := sig.Write(data); err != nil {
		return errors.Wrap(err, "sig.Write")
	}

	s.ManabiePublicKey = publicKey.String()
	s.ManabieSignature = hex.EncodeToString(sig.Sum(nil))

	return nil
}

func (s *suite) makeHTTPRequest(method, url string, bodyRequest []byte) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyRequest))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("manabie-public-key", s.ManabiePublicKey)
	req.Header.Set("manabie-signature", s.ManabieSignature)

	if err != nil {
		return nil, err
	}

	client := http.Client{Timeout: time.Duration(30) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	respStruct := http_port.ResponseErrors{}
	err = json.Unmarshal(body, &respStruct)
	if err != nil {
		return nil, err
	}

	s.Response = respStruct

	return body, nil
}

func (s *suite) studentsWereSuccessfullyByOpenAPI(ctx context.Context) (context.Context, error) {
	req := s.Request.(*http_port.UpsertStudentsRequest)
	resp := s.Response.(http_port.ResponseErrors)

	if resp.Code != 20000 {
		return ctx, fmt.Errorf("message: %s, code: %d", resp.Message, resp.Code)
	}
	dataSlice, ok := resp.Data.([]interface{})
	dataResp := make([]http_port.DataResponse, len(dataSlice))

	if ok {
		for i, data := range dataSlice {
			dataMap, ok := data.(map[string]interface{})
			if ok {
				dataResp[i].UserID = dataMap["user_id"].(string)
				dataResp[i].ExternalUserID = dataMap["external_user_id"].(string)
			}
		}
	}

	mapExternalUserIDAndUserID := make(map[string]string)

	for _, user := range dataResp {
		mapExternalUserIDAndUserID[user.ExternalUserID] = user.UserID
	}

	domainStudents := make(aggregate.DomainStudents, 0, len(req.Students))
	for idx, student := range req.Students {
		student.UserID = field.NewString(mapExternalUserIDAndUserID[student.ExternalUserID.String()])
		domainStudent := http_port.ToDomainStudentV2(http_port.DomainStudentImpl{
			NullDomainStudent: entity.NullDomainStudent{},
			StudentProfile:    student,
		}, idx, true)
		domainStudents = append(domainStudents, domainStudent)
	}

	service := s.InitStudentValidationManager()
	studentsToUpsert := make(aggregate.DomainStudents, 0)
	studentsToCreate, studentsToUpdate, _ := service.FullyValidate(ctx, s.BobDBTrace, domainStudents, true)
	studentsToUpsert = append(studentsToUpsert, studentsToCreate...)
	studentsToUpsert = append(studentsToUpsert, studentsToUpdate...)
	if ctx, err := s.verifyStudentsInBD(ctx, studentsToUpsert); err != nil {
		return ctx, fmt.Errorf("verifyStudentInBDAfterCreatedSuccessfully err:%v", err)
	}

	// if ctx, err = s.verifyLocationInNatsEvent(ctx, userIDs); err != nil {
	// 	return ctx, fmt.Errorf("verifyLocationInNatsEvent err:%v", err)
	// }
	return ctx, nil
}

func (s *suite) studentsWereCreatedUnsuccessfullyWithCodeAndField(ctx context.Context, code string, field string) (context.Context, error) {
	resp := s.Response.(http_port.ResponseErrors)
	req := s.Request.(*http_port.UpsertStudentsRequest)
	if resp.Message == "ok" {
		return ctx, fmt.Errorf("expected error when calling OpenAPI")
	}

	if strconv.Itoa(resp.Code) != code {
		return ctx, fmt.Errorf("expected code is %s, actual is %d", code, resp.Code)
	}

	if !strings.Contains(resp.Message, field) {
		return ctx, fmt.Errorf("expected field: %s, but actual message is: %s", field, resp.Message)
	}

	emails := make([]string, 0, len(req.Students))
	for _, student := range req.Students {
		emails = append(emails, student.Email.String())
	}
	ctx, err := s.verifyUsersNotInBD(ctx, emails)
	if err != nil {
		return ctx, fmt.Errorf("verifyUsersNotInBD err:%v", err)
	}
	return ctx, nil
}

func (s *suite) studentsWereUpdatedUnsuccessfullyWithCodeAndField(ctx context.Context, code string, fieldName string) (context.Context, error) {
	respUpdate := s.Response.(http_port.ResponseErrors)
	if respUpdate.Message == "ok" {
		return ctx, fmt.Errorf("expected error when calling OpenAPI")
	}

	if strconv.Itoa(respUpdate.Code) != code {
		return ctx, fmt.Errorf("expected code is %s, actual is %d", code, respUpdate.Code)
	}

	switch respUpdate.Code {
	case errcode.InvalidData, errcode.MissingMandatory, errcode.DuplicatedData, errcode.DataExist:
		if !strings.Contains(respUpdate.Message, fieldName) {
			return ctx, fmt.Errorf("expected message %s contains %s", respUpdate.Message, fieldName)
		}
	}
	// Check student created in db
	reqCreate := s.Request1.(*http_port.UpsertStudentsRequest)
	respCreate := s.Response1.(http_port.ResponseErrors)

	if respCreate.Code != 20000 {
		return ctx, fmt.Errorf("message: %s, code: %d", respCreate.Message, respCreate.Code)
	}
	dataSlice, ok := respCreate.Data.([]interface{})
	dataResp := make([]http_port.DataResponse, len(dataSlice))

	if ok {
		for i, data := range dataSlice {
			dataMap, ok := data.(map[string]interface{})
			if ok {
				dataResp[i].UserID = dataMap["user_id"].(string)
				dataResp[i].ExternalUserID = dataMap["external_user_id"].(string)
			}
		}
	}
	mapExternalUserIDAndUserID := make(map[string]string)

	for _, user := range dataResp {
		mapExternalUserIDAndUserID[user.ExternalUserID] = user.UserID
	}

	domainStudents := make(aggregate.DomainStudents, 0, len(reqCreate.Students))
	for idx, student := range reqCreate.Students {
		student.UserID = field.NewString(mapExternalUserIDAndUserID[student.ExternalUserID.String()])
		domainStudent := http_port.ToDomainStudentV2(http_port.DomainStudentImpl{
			NullDomainStudent: entity.NullDomainStudent{},
			StudentProfile:    student,
		}, idx, true)
		domainStudents = append(domainStudents, domainStudent)
	}

	service := s.InitStudentValidationManager()
	studentsToUpsert := make(aggregate.DomainStudents, 0)
	studentsToCreate, studentsToUpdate, _ := service.FullyValidate(ctx, s.BobDBTrace, domainStudents, true)
	studentsToUpsert = append(studentsToUpsert, studentsToCreate...)
	studentsToUpsert = append(studentsToUpsert, studentsToUpdate...)

	if ctx, err := s.verifyStudentsInBD(ctx, studentsToUpsert); err != nil {
		return ctx, fmt.Errorf("verifyStudentInBDAfterCreatedSuccessfully err:%v", err)
	}
	return ctx, nil
}

func (s *suite) createStudentsByOpenAPI(ctx context.Context, numberOfStudents int, conditions string, folder string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &http_port.UpsertStudentsRequest{}

	path := fmt.Sprintf("usermgmt/testdata/json/%s/create_%d_student_%s.json", folder, numberOfStudents, strings.ReplaceAll(conditions, " ", "_"))

	jsonByte, err := os.ReadFile(path)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("os.ReadFile err: %v", err)
	}

	if err := json.Unmarshal(jsonByte, req); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("json.Unmarshal err: %v", err)
	}

	for idx, student := range req.Students {
		switch conditions {
		case "missing external_user_id", "empty external user id with spaces":
			// Skip make unique external_user_id
			id := idutil.ULIDNow()
			student.Email = field.NewString(id + student.Email.String())
			if !student.UserName.TrimSpace().IsEmpty() {
				student.UserName = field.NewString(id + student.UserName.String())
			}
		case "missing email":
			// Skip make unique email
			id := idutil.ULIDNow()
			student.ExternalUserID = field.NewString(student.ExternalUserID.String() + id)
			if !student.UserName.TrimSpace().IsEmpty() {
				student.UserName = field.NewString(id + student.UserName.String())
			}
		case "with existing username", "with existing username and upper case":
			// Skip make unique username
			id := idutil.ULIDNow()
			student.ExternalUserID = field.NewString(student.ExternalUserID.String() + id)
			student.Email = field.NewString(id + student.Email.String())
		default:
			id := idutil.ULIDNow()
			student.ExternalUserID = field.NewString(id + student.ExternalUserID.String())
			student.Email = field.NewString(id + student.Email.String())
			if !student.UserName.TrimSpace().IsEmpty() {
				student.UserName = field.NewString(id + student.UserName.String())
			}
		}
		req.Students[idx] = student
	}
	stepState.Request = req
	for _, v := range req.Students {
		stepState.StudentEmails = append(stepState.StudentEmails, v.Email.String())
	}
	jsonPayload, err := json.Marshal(req)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("json.Marshal err: %v", err)
	}
	ctx, err = s.createSubscriptionForCreatedStudentByOpenAPI(ctx, req)
	if err != nil {
		return ctx, errors.Wrap(err, "s.createSubscriptionForCreatedStudentByOpenAPI")
	}

	err = s.setupAPIKey(ctx)
	if err != nil {
		return ctx, errors.Wrap(err, "s.setupAPIKey")
	}
	url := fmt.Sprintf(`http://%s%s`, s.Cfg.UserMgmtRestAddr, constant.DomainStudentEndpoint)

	bodyBytes, err := s.makeHTTPRequest(http.MethodPut, url, jsonPayload)
	if err != nil {
		return ctx, errors.Wrap(err, "s.makeHTTPRequest")
	}

	if bodyBytes == nil {
		return ctx, fmt.Errorf("body is nil")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateStudentsByOpenAPI(ctx context.Context, numberOfStudents int, conditions string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(http_port.ResponseErrors)
	req := stepState.Request.(*http_port.UpsertStudentsRequest)

	// req and resp create student
	stepState.Request1 = req
	stepState.Response1 = resp

	dataSlice, ok := resp.Data.([]interface{})
	dataResp := make([]http_port.DataResponse, len(dataSlice))

	if ok {
		for i, data := range dataSlice {
			dataMap, ok := data.(map[string]interface{})
			if ok {
				dataResp[i].UserID = dataMap["user_id"].(string)
				dataResp[i].ExternalUserID = dataMap["external_user_id"].(string)
			}
		}
	}
	mapUserIDAndExternalUserID := make(map[string]string, len(dataResp))

	for _, user := range dataResp {
		mapUserIDAndExternalUserID[user.ExternalUserID] = user.UserID
	}

	for idx, student := range req.Students {
		student.UserID = field.NewString(mapUserIDAndExternalUserID[student.ExternalUserID.String()])
		req.Students[idx] = student
	}

	for idx := 0; idx < numberOfStudents; idx++ {
		student := req.Students[idx]
		switch conditions {
		case "empty enrollment_status_histories":
			student.EnrollmentStatusHistories = nil
			student.FirstName = field.NewString("enrollment_status_histories" + student.FirstName.String())
		case "edit first_name":
			student.FirstName = field.NewString("edited" + student.FirstName.String())
		case "edit last_name":
			student.LastName = field.NewString("edited" + student.LastName.String())
		case "edit email":
			student.Email = field.NewString("edited" + student.Email.String())
		case "edit first_name_phonetic":
			student.FirstNamePhonetic = field.NewString("edited" + student.FirstNamePhonetic.String())
		case "edit last_name_phonetic":
			student.LastNamePhonetic = field.NewString("edited" + student.FirstNamePhonetic.String())
		case "edit birthday":
			student.Birthday = field.NewDate(time.Now())
		case "edit gender":
			if student.Gender.Int32() == 1 {
				student.Gender = field.NewInt32(2)
			} else {
				student.Gender = field.NewInt32(1)
			}
		case "edit grade":
			student.Grade = field.NewString("grade_02")
		case "edit with empty locations, other info still remains":
			student.Locations = nil
		case "changing to enrollment status non-potential":
			for idx := range student.EnrollmentStatusHistories {
				student.EnrollmentStatusHistories[idx].EnrollmentStatus = field.NewInt16(7)
			}
		case "changing to enrollment status temporary":
			for idx := range student.EnrollmentStatusHistories {
				student.EnrollmentStatusHistories[idx].EnrollmentStatus = field.NewInt16(6)
			}
		case "editing to non-existing external_user_id":
			student.ExternalUserID = field.NewString(student.ExternalUserID.String() + "non-existing")
		case "editing to duplicated external_user_id":
			student.ExternalUserID = field.NewString(req.Students[0].ExternalUserID.String())
		case "editing to empty external_user_id":
			student.ExternalUserID = field.NewString("")
		case "editing to empty external_user_id with spaces":
			student.ExternalUserID = field.NewString("	")
		case "editing to non-existing grade":
			student.Grade = field.NewString("non-existing")
		case "editing to out of range gender":
			student.Gender = field.NewInt32(100)
		case "editing to non-existing school":
			student.SchoolHistories = []http_port.SchoolHistoryPayload{
				{
					School:       field.NewString("school_partner_id_1_non_existing"),
					SchoolCourse: field.NewString("school_course_partner_id_01"),
				},
			}
		case "editing to non-existing school_course":
			student.SchoolHistories = []http_port.SchoolHistoryPayload{
				{
					School:       field.NewString("school_partner_id_1"),
					SchoolCourse: field.NewString("school_course_partner_id_01_non_existing"),
				},
			}
		case "editing to empty first_name":
			student.FirstName = field.NewString("")
		case "editing to empty last_name":
			student.LastName = field.NewString("")
		case "edit with external_user_id with spaces, other info still remains":
			student.ExternalUserID = field.NewString(student.ExternalUserID.String() + "    ")
		case "changing to enrollment status graduate and new date":
			for idx := range student.EnrollmentStatusHistories {
				student.EnrollmentStatusHistories[idx].EnrollmentStatus = field.NewInt16(4)
				student.EnrollmentStatusHistories[idx].StartDate = field.NewDate(time.Now())
			}
		case "changing to enrollment status withdraw and new date":
			for idx := range student.EnrollmentStatusHistories {
				student.EnrollmentStatusHistories[idx].EnrollmentStatus = field.NewInt16(3)
				student.EnrollmentStatusHistories[idx].StartDate = field.NewDate(time.Now())
			}
		case "changing to enrollment status LOA and new date":
			for idx := range student.EnrollmentStatusHistories {
				student.EnrollmentStatusHistories[idx].EnrollmentStatus = field.NewInt16(5)
				student.EnrollmentStatusHistories[idx].StartDate = field.NewDate(time.Now())
			}
		case "changing to enrollment status enrolled new date":
			for idx := range student.EnrollmentStatusHistories {
				student.EnrollmentStatusHistories[idx].EnrollmentStatus = field.NewInt16(2)
				student.EnrollmentStatusHistories[idx].StartDate = field.NewDate(time.Now())
			}
		case "changing to enrollment status non-potential new date":
			for idx := range student.EnrollmentStatusHistories {
				student.EnrollmentStatusHistories[idx].EnrollmentStatus = field.NewInt16(7)
				student.EnrollmentStatusHistories[idx].StartDate = field.NewDate(time.Now())
			}
		case "changing to enrollment status potential and new date":
			for idx := range student.EnrollmentStatusHistories {
				student.EnrollmentStatusHistories[idx].EnrollmentStatus = field.NewInt16(1)
				student.EnrollmentStatusHistories[idx].StartDate = field.NewDate(time.Now())
			}
		case "adding new location with enrollment status potential with feature date":
			student.EnrollmentStatusHistories = append(student.EnrollmentStatusHistories, http_port.EnrollmentStatusHistoryPayload{
				EnrollmentStatus: field.NewInt16(1),
				StartDate:        field.NewDate(time.Now().Add(time.Hour * 30)),
				Location:         field.NewString("location-id-3"),
			})

		case "adding new location with enrollment status temporary with feature date":
			student.EnrollmentStatusHistories = append(student.EnrollmentStatusHistories, http_port.EnrollmentStatusHistoryPayload{
				EnrollmentStatus: field.NewInt16(6),
				StartDate:        field.NewDate(time.Now().Add(time.Hour * 30)),
				Location:         field.NewString("location-id-3"),
			})
		case "adding new location with enrollment status non-potential with feature date":
			student.EnrollmentStatusHistories = append(student.EnrollmentStatusHistories, http_port.EnrollmentStatusHistoryPayload{
				EnrollmentStatus: field.NewInt16(7),
				StartDate:        field.NewDate(time.Now().Add(time.Hour * 30)),
				Location:         field.NewString("location-id-3"),
			})
		case "editing to external_user_id was used by parent":
			student.ExternalUserID = field.NewString("parent_external_id_existing_02")
		default:
			return StepStateToContext(ctx, stepState), fmt.Errorf(`this "%v" condition is not supported`, conditions)
		}
		req.Students[idx] = student
	}
	err := s.setupAPIKey(ctx)
	if err != nil {
		return ctx, errors.Wrap(err, "s.setupAPIKey")
	}

	stepState.Request = req

	jsonPayload, err := json.Marshal(req)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("json.Marshal err: %v", err)
	}

	url := fmt.Sprintf(`http://%s%s`, s.Cfg.UserMgmtRestAddr, constant.DomainStudentEndpoint)

	bodyBytes, err := s.makeHTTPRequest(http.MethodPut, url, jsonPayload)
	if err != nil {
		return ctx, errors.Wrap(err, "s.makeHTTPRequest")
	}

	if bodyBytes == nil {
		return ctx, fmt.Errorf("body is nil")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentsWereUpsertedByOpenAPIWithErrorsCollection(ctx context.Context, numberStudentFailed int, numberStudentSuccessful int, tableConditions *godog.Table) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rows := tableConditions.Rows
	if numberStudentFailed != 0 {
		resp := s.Response.(http_port.ResponseErrors)
		errors := resp.Errors

		if len(errors) != numberStudentFailed {
			return ctx, fmt.Errorf("len(errors) != numberStudentFailed, len(errors): %v, numberStudentFailed: %v, err: %v", len(errors), numberStudentFailed, errors)
		}

		for i := 1; i < len(rows); i++ {
			v := rows[i]
			condition := v.Cells[0].Value
			field := v.Cells[1].Value
			code := v.Cells[2].Value
			index := v.Cells[3].Value

			idx, err := strconv.ParseInt(index, 10, 32)
			if err != nil {
				return ctx, fmt.Errorf("csvRow-strconv.ParseInt err: %v", err)
			}
			codeErr, err := strconv.ParseInt(code, 10, 32)
			if err != nil {
				return ctx, fmt.Errorf("codeErr-strconv.ParseInt err: %v", err)
			}
			for _, err := range errors {
				indexError := grpc_port.GetIndexFromMessageError(err.Error)
				fieldName := grpc_port.GetFieldFromMessageError(err.Error)
				if indexError == int(idx) {
					if fieldName != field {
						return ctx, fmt.Errorf("expected: field name is %v, actual: but field name is %v | condition: %s", field, fieldName, condition)
					}
					if err.Code != int(codeErr) {
						return ctx, fmt.Errorf("expected: code is %v, actual: but code is %v | condition: %s", code, err.Code, condition)
					}
				}
			}
		}
	}

	if numberStudentSuccessful != 0 {
		req := s.Request.(*http_port.UpsertStudentsRequest)
		students := req.Students

		insertedStudents := make([]http_port.StudentProfile, 0, numberStudentSuccessful)
		failedIndexStudents := make([]int, 0, numberStudentSuccessful)

		for i := 1; i < len(rows); i++ {
			v := rows[i]
			index := v.Cells[3].Value
			idx, err := strconv.ParseInt(index, 10, 32)
			if err != nil {
				return ctx, fmt.Errorf("OpenAPI index strconv.ParseInt err: %v", err)
			}
			failedIndexStudents = append(failedIndexStudents, int(idx))
		}

		emails := make([]string, 0, numberStudentSuccessful)
		for idx, student := range students {
			if !utils.InArrayInt(idx, failedIndexStudents) {
				insertedStudents = append(insertedStudents, student)
				emails = append(emails, student.Email.String())
			}
		}

		userRepo := repository.DomainUserRepo{}

		users, err := userRepo.GetByEmailsInsensitiveCase(ctx, s.BobDBTrace, emails)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("userRepo.GetByEmailsInsensitiveCase err: %v", err)
		}

		if len(users) != numberStudentSuccessful {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected: number of users is %v, actual: but number of users is %v", numberStudentSuccessful, len(users))
		}

		mapEmailAndUserID := make(map[string]string, len(users))
		for _, user := range users {
			mapEmailAndUserID[user.Email().String()] = user.UserID().String()
		}

		domainStudents := make(aggregate.DomainStudents, 0, len(users))
		for i, student := range insertedStudents {
			student.UserID = field.NewString(mapEmailAndUserID[student.Email.String()])
			if !student.UserID.IsEmpty() {
				s := http_port.ToDomainStudentV2(http_port.DomainStudentImpl{
					StudentProfile: student,
				}, i, true)
				domainStudents = append(domainStudents, s)
			}
		}
		service := s.InitStudentValidationManager()
		studentsToUpsert := make(aggregate.DomainStudents, 0)
		studentsToCreate, studentsToUpdate, _ := service.FullyValidate(ctx, s.BobDBTrace, domainStudents, true)
		studentsToUpsert = append(studentsToUpsert, studentsToCreate...)
		studentsToUpsert = append(studentsToUpsert, studentsToUpdate...)

		if ctx, err = s.verifyStudentsInBD(ctx, studentsToUpsert); err != nil {
			return ctx, fmt.Errorf("verifyStudentInBDAfterCreatedSuccessfully err:%v", err)
		}
	}
	if numberStudentFailed == 0 && numberStudentSuccessful == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("numberRowsSuccessful is 0 and numberRowsFailed is 0")
	}
	return StepStateToContext(ctx, stepState), nil
}
