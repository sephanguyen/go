package usermgmt

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc/importstudent"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/cucumber/godog"
	"github.com/gocarina/gocsv"
	"github.com/pkg/errors"
)

var (
	/* "1" */ potentialEnrollmentStatusEnum = fmt.Sprintf("%d", upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL)
	/* "2" */ enrolledEnrollmentStatusEnum = fmt.Sprintf("%d", upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED)
	/* "3" */ withdrawnEnrollmentStatusEnum = fmt.Sprintf("%d", upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN)
	/* "4" */ graduatedEnrollmentStatusEnum = fmt.Sprintf("%d", upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED)
	/* "5" */ loaEnrollmentStatusEnum = fmt.Sprintf("%d", upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_LOA)
	/* "6" */ temporaryEnrollmentStatusEnum = fmt.Sprintf("%d", upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY)
	/* "7" */ nonPotentialEnrollmentStatusEnum = fmt.Sprintf("%d", upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL)
)

func (s *suite) studentsWereUpsertedSuccessfulByImport(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*upb.ImportStudentRequest)
	resp := stepState.Response.(*upb.UpsertStudentResponse)
	// err := stepState.ResponseErr

	errors := resp.Messages
	if len(errors) > 0 {
		return ctx, fmt.Errorf("expected: there is not error, actual: but there are errors, list of errors: %v", errors)
	}
	// errors := []*upb.ErrorMessage{}
	// for _, detail := range status.Convert(err).Details() {
	// 	errorMessages, ok := detail.(*upb.ErrorMessages)
	// 	if !ok {
	// 		return ctx, fmt.Errorf("errorMessages, not ok: detail.(*upb.ErrorMessages), err: %v", err)
	// 	}
	// 	errors = append(errors, errorMessages.Messages...)
	// }
	// return ctx, fmt.Errorf("expected: there is not error, actual: but there are errors, err: %v, list of errors: %v", err, errors)

	// if err != nil {
	// 	errors := []*upb.ErrorMessage{}
	// 	for _, detail := range status.Convert(err).Details() {
	// 		errorMessages, ok := detail.(*upb.ErrorMessages)
	// 		if !ok {
	// 			return ctx, fmt.Errorf("errorMessages, not ok: detail.(*upb.ErrorMessages), err: %v", err)
	// 		}
	// 		errors = append(errors, errorMessages.Messages...)
	// 	}
	// 	return ctx, fmt.Errorf("expected: there is not error, actual: but there are errors, err: %v, list of errors: %v", err, errors)
	// }

	mapEmailAndStudentID := make(map[string]string, len(resp.GetStudentProfiles()))

	for _, studentProfile := range resp.GetStudentProfiles() {
		mapEmailAndStudentID[studentProfile.GetEmail()] = studentProfile.Id
	}

	students, err := importstudent.ConvertPayloadToImportStudentData(req.Payload)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("importstudent.ConvertPayloadToImportStudentData err: %v", err)
	}

	domainStudents := make(aggregate.DomainStudents, 0, len(students))

	for i, student := range students {
		student.IDAttr = field.NewString(mapEmailAndStudentID[student.EmailAttr.String()])
		s, _ := importstudent.ToDomainStudentsV2(student, i, true)
		domainStudents = append(domainStudents, s)
	}

	service := s.InitStudentValidationManager()

	studentsToUpsert := make(aggregate.DomainStudents, 0)
	studentsToCreate, studentsToUpdate, _ := service.FullyValidate(ctx, s.BobDBTrace, domainStudents, true)
	studentsToUpsert = append(studentsToUpsert, studentsToCreate...)
	studentsToUpsert = append(studentsToUpsert, studentsToUpdate...)
	if ctx, err = s.verifyStudentsInBD(ctx, studentsToUpsert); err != nil {
		return ctx, fmt.Errorf("verifyStudentInBDAfterCreatedSuccessfully err:%v", err)
	}
	// @an-tang will fix in this ticket [LT-42562]
	// if ctx, err = s.verifyLocationInNatsEvent(ctx, userIDs); err != nil {
	// 	return ctx, fmt.Errorf("verifyLocationInNatsEvent err:%v", err)
	// }
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateStudentsByImport(ctx context.Context, numberOfStudents int, conditions string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	createReq := stepState.Request.(*upb.ImportStudentRequest)

	createResp := s.Response.(*upb.UpsertStudentResponse)

	mapEmailAndStudentID := make(map[string]string, 0)

	for _, student := range createResp.StudentProfiles {
		mapEmailAndStudentID[student.Email] = student.Id
	}
	students, err := importstudent.ConvertPayloadToImportStudentData(createReq.Payload)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("importstudent.ConvertPayloadToImportStudentData err: %v", err)
	}
	// edit csv with condition
	const editString = "edited"

	for _, student := range students {
		student.IDAttr = field.NewString(mapEmailAndStudentID[student.EmailAttr.String()])
	}

	for i := 0; i < numberOfStudents; i++ {
		student := students[i]
		editEnrollmentStatusHistory(conditions, student)
		switch conditions {
		case "edit first_name":
			student.FirstNameAttr = field.NewString(student.FirstNameAttr.String() + editString)
		case "edit last_name":
			student.LastNameAttr = field.NewString(student.LastNameAttr.String() + editString)
		case "edit first_name_phonetic":
			student.FirstNamePhoneticAttr = field.NewString(student.FirstNamePhoneticAttr.String() + editString)
		case "edit last_name_phonetic":
			student.LastNamePhoneticAttr = field.NewString(student.LastNamePhoneticAttr.String() + editString)
		case "edit birthday":
			student.BirthdayAttr = field.NewString("1999/12/12")
		case "edit gender":
			student.GenderAttr = field.NewString("2")
		case "edit grade":
			student.GradeAttr = field.NewString("grade_02")
		case "editing to non-existing user_id":
			student.IDAttr = field.NewString("non-existing-user-id")
		case "editing to duplicated user_id":
			student.IDAttr = field.NewString(mapEmailAndStudentID[students[0].EmailAttr.String()])
		case "editing to non-existing grade":
			student.GradeAttr = field.NewString("non-existing-grade")
		case "editing to wrong format birthday":
			student.BirthdayAttr = field.NewString("12/12/2000")
		case "editing to out of range gender":
			student.GenderAttr = field.NewString("10")
		case "editing to text gender":
			student.GenderAttr = field.NewString("text")
		case "editing to non-existing school":
			student.SchoolAttr = field.NewString("non-existing-school")
		case "editing to non-existing school_course":
			student.SchoolCourseAttr = field.NewString("non-existing-school-course")
		case "editing to empty first_name":
			student.FirstNameAttr = field.NewString("")
		case "editing to empty last_name":
			student.LastNameAttr = field.NewString("")
		case "editing to empty external_user_id":
			student.ExternalUserIDAttr = field.NewString("")
		case "editing to empty external_user_id with spaces":
			student.ExternalUserIDAttr = field.NewString("		")
		case "editing to duplicated external_user_id":
			student.ExternalUserIDAttr = field.NewString(students[0].ExternalUserIDAttr.String())
		case "editing to duplicated email":
			student.EmailAttr = field.NewString(students[0].EmailAttr.String())
		case "edit external_user_id with spaces":
			student.ExternalUserIDAttr = field.NewString(student.ExternalUserIDAttr.String() + "	")
		case "edit to have external_user_id":
			id := idutil.ULIDNow()
			student.ExternalUserIDAttr = field.NewString(id + "external_user_id")
			// Because we complexity of the condition, we need to split the condition and skip the default case temporarily
			// default:
			// 	return StepStateToContext(ctx, stepState), fmt.Errorf("invalid condition: %s", conditions)
		case "keep existing username":
			student.UserNameAttr = field.NewString(student.UserNameAttr.String())
		case "editing to another available username":
			student.UserNameAttr = field.NewString(student.UserNameAttr.String() + idutil.ULIDNow())
		case "editing to another available username with email format":
			student.UserNameAttr = field.NewString(fmt.Sprintf("user+%s@%s.com", idutil.ULIDNow(), idutil.ULIDNow()))
		case "editing to duplicated username":
			// update the rest students to have the same username with the first student
			student.UserNameAttr = field.NewString(students[0].UserNameAttr.String())
		case "editing to existing username":
			username, err := s.getUsernameByUsingOther(ctx)
			if err != nil {
				return nil, err
			}
			student.UserNameAttr = field.NewString(username)
		case "editing to existing username and upper case":
			username, err := s.getUsernameByUsingOther(ctx)
			if err != nil {
				return nil, err
			}
			student.UserNameAttr = field.NewString(strings.ToUpper(username))
		case "editing to multiple spaces username":
			student.UserNameAttr = field.NewString("   ")
		case "editing to single space username":
			student.UserNameAttr = field.NewString(" ")
		case "editing to empty username":
			student.UserNameAttr = field.NewString("")
		case "editing to special characters username":
			student.UserNameAttr = field.NewString(":'(")
		case "editing to existing external_user_id":
			if _, err := s.createStudentByGRPC(ctx, "general info", "all fields"); err != nil {
				return ctx, fmt.Errorf("s.createStudentByGRPC err:%v", err)
			}
			stepState := StepStateFromContext(ctx)
			resp := stepState.Response.(*upb.UpsertStudentResponse)
			externalUserID := resp.StudentProfiles[0].ExternalUserId
			student.ExternalUserIDAttr = field.NewString(externalUserID)
		}
	}
	payload, err := gocsv.MarshalBytes(&students)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("gocsv.MarshalBytes: %s", err)
	}
	// Update with new payload
	updateRep := &upb.ImportStudentRequest{Payload: payload}
	stepState.RequestSentAt = time.Now()
	updateResp, err := upb.NewStudentServiceClient(s.UserMgmtConn).
		ImportStudentV2(contextWithToken(ctx), updateRep)

	stepState.ResponseErr = err
	stepState.Response = updateResp
	stepState.Request = updateRep
	stepState.Request1 = createReq // created request to verify after update unsuccessful
	stepState.Response1 = createResp
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createStudentsByImport(ctx context.Context, numberOfStudents int, conditions string, folder string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	path := fmt.Sprintf("usermgmt/testdata/csv/%s/create_%d_student_%s.csv", folder, numberOfStudents, strings.ReplaceAll(conditions, " ", "_"))

	csv, err := os.ReadFile(path)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("os.ReadFile err: %v", err)
	}

	students, err := importstudent.ConvertPayloadToImportStudentData(csv)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("importstudent.ConvertPayloadToImportStudentData err: %v", err)
	}

	for i, student := range students {
		id := idutil.ULIDNow()
		students[i].EmailAttr = field.NewString(id + student.EmailAttr.String())

		if !student.ExternalUserIDAttr.TrimSpace().IsEmpty() {
			students[i].ExternalUserIDAttr = field.NewString(id + student.ExternalUserIDAttr.String())
		}
		switch conditions {
		case "with existing username", "with existing username and upper case":
			break
		default:
			if !student.UserNameAttr.TrimSpace().IsEmpty() {
				students[i].UserNameAttr = field.NewString(id + student.UserNameAttr.String())
			}
		}
	}

	payload, err := gocsv.MarshalBytes(&students)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("gocsv.MarshalBytes: %s", err)
	}
	req := &upb.ImportStudentRequest{Payload: payload}

	ctx, err = s.createSubscriptionForCreatedStudentByImport(ctx, req)
	if err != nil {
		return ctx, errors.Wrap(err, "s.createSubscriptionForCreatedStudentByImport")
	}
	// create student by import
	stepState.RequestSentAt = time.Now()
	resp, err := upb.NewStudentServiceClient(s.UserMgmtConn).
		ImportStudentV2(contextWithToken(ctx), req)
	stepState.ResponseErr = err
	stepState.Response = resp
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentWereCreatedUnsuccessfulByImportWithError(ctx context.Context, stringCode string, csvField string, rowCSV string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// err := stepState.ResponseErr
	resp := stepState.Response.(*upb.UpsertStudentResponse)
	req := stepState.Request.(*upb.ImportStudentRequest)
	errors := resp.Messages
	// if err != nil {
	// 	for _, detail := range status.Convert(err).Details() {
	// 		errorMessages, ok := detail.(*upb.ErrorMessages)
	// 		if !ok {
	// 			return ctx, fmt.Errorf("errorMessages, not ok: detail.(*upb.ErrorMessages), err: %v", err)
	// 		}
	// 		errors = append(errors, errorMessages.Messages...)
	// 	}
	// }
	code, err := strconv.ParseInt(stringCode, 10, 64)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("stringCode - strconv.Atoi err: %v", err)
	}
	row, err := strconv.ParseInt(rowCSV, 10, 64)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("rowCSV - strconv.Atoi err: %v", err)
	}
	if len(errors) == 0 {
		return ctx, fmt.Errorf("expected error but got none")
	}
	for _, err := range errors {
		codeErr := err.Code
		fieldName := err.FieldName
		index := err.Index
		if codeErr != int32(code) {
			return ctx, fmt.Errorf("code is incorrect: %s, code: %d, field name: %s, row: %d", err.Error, err.Code, err.FieldName, err.Index)
		}
		if fieldName != csvField {
			return ctx, fmt.Errorf("field name is incorrect: %s, code: %d, field name: %s, row: %d", err.Error, err.Code, err.FieldName, err.Index)
		}
		if index != int32(row) {
			return ctx, fmt.Errorf("row is incorrect: %s, code: %d, field name: %s, row: %d", err.Error, err.Code, err.FieldName, err.Index)
		}
	}
	students, err := importstudent.ConvertPayloadToImportStudentData(req.Payload)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("importstudent.ConvertPayloadToImportStudentData err: %v", err)
	}
	emails := []string{}
	for _, student := range students {
		emails = append(emails, student.EmailAttr.String())
	}
	ctx, err = s.verifyUsersNotInBD(ctx, emails)
	if err != nil {
		return ctx, fmt.Errorf("s.verifyUsersNotInBD err: %v", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentWereUpdatedUnsuccessfulByImportWithError(ctx context.Context, stringCode string, csvField string, rowCSV string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	createdReq := stepState.Request1.(*upb.ImportStudentRequest)
	createdResp := stepState.Response1.(*upb.UpsertStudentResponse)
	updateResp := stepState.Response.(*upb.UpsertStudentResponse)
	// err := stepState.ResponseErr
	errors := updateResp.Messages
	// if err != nil {
	// 	for _, detail := range status.Convert(err).Details() {
	// 		errorMessages, ok := detail.(*upb.ErrorMessages)
	// 		if !ok {
	// 			return ctx, fmt.Errorf("errorMessages, ok := detail.(*upb.ErrorMessages) err: %v", err)
	// 		}
	// 		errors = append(errors, errorMessages.Messages...)
	// 	}
	// }
	code, err := strconv.ParseInt(stringCode, 10, 64)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("stringCode - strconv.Atoi err: %v", err)
	}
	row, err := strconv.ParseInt(rowCSV, 10, 64)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("rowCSV - strconv.Atoi err: %v", err)
	}
	if len(errors) == 0 {
		return ctx, fmt.Errorf("expected error but got none")
	}
	for _, err := range errors {
		codeErr := err.Code
		fieldName := err.FieldName
		index := err.Index
		if codeErr != int32(code) {
			return ctx, fmt.Errorf("code is incorrect: %s, code: %d, field name: %s, row: %d", err.Error, err.Code, err.FieldName, err.Index)
		}
		if fieldName != csvField {
			return ctx, fmt.Errorf("field name is incorrect: %s, code: %d, field name: %s, row: %d", err.Error, err.Code, err.FieldName, err.Index)
		}
		if index != int32(row) {
			return ctx, fmt.Errorf("row is incorrect: %s, code: %d, field name: %s, row: %d", err.Error, err.Code, err.FieldName, err.Index)
		}
	}

	// verify students were not updated
	mapEmailAndStudentID := make(map[string]string, len(createdResp.GetStudentProfiles()))
	for _, studentProfile := range createdResp.GetStudentProfiles() {
		mapEmailAndStudentID[studentProfile.GetEmail()] = studentProfile.Id
	}

	students, err := importstudent.ConvertPayloadToImportStudentData(createdReq.Payload)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("importstudent.ConvertPayloadToImportStudentData err: %v", err)
	}

	domainStudents := make(aggregate.DomainStudents, 0, len(students))

	for i, student := range students {
		student.IDAttr = field.NewString(mapEmailAndStudentID[student.EmailAttr.String()])
		s, _ := importstudent.ToDomainStudentsV2(student, i, true)
		domainStudents = append(domainStudents, s)
	}

	service := s.InitStudentValidationManager()

	studentsToUpsert := make(aggregate.DomainStudents, 0)
	studentsToCreate, studentsToUpdate, _ := service.FullyValidate(ctx, s.BobDBTrace, domainStudents, true)
	studentsToUpsert = append(studentsToUpsert, studentsToCreate...)
	studentsToUpsert = append(studentsToUpsert, studentsToUpdate...)

	if ctx, err = s.verifyStudentsInBD(ctx, studentsToUpsert); err != nil {
		return ctx, fmt.Errorf("verifyStudentInBDAfterCreatedSuccessfully err:%v", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func editEnrollmentStatusHistory(conditions string, student *importstudent.StudentCSV) {
	switch conditions {
	case "empty enrollment_status_histories":
		student.EnrollmentStatusAttr = field.NewNullString()
		student.LocationAttr = field.NewNullString()
		student.FirstNameAttr = field.NewString(student.FirstNameAttr.String() + "edited")
	case "changing to enrollment status temporary":
		student.EnrollmentStatusAttr = field.NewString(temporaryEnrollmentStatusEnum)
	case "changing to enrollment status non-potential":
		student.EnrollmentStatusAttr = field.NewString(nonPotentialEnrollmentStatusEnum)
	case "changing to enrollment status non-potential and new date":
		student.EnrollmentStatusAttr = field.NewString(nonPotentialEnrollmentStatusEnum)
		student.StatusStartDateAttr = field.NewString(field.NewDate(time.Now()).Date().Format(constant.DateLayout))
	case "changing to enrollment status graduate and new date":
		student.EnrollmentStatusAttr = field.NewString(graduatedEnrollmentStatusEnum)
		student.StatusStartDateAttr = field.NewString(field.NewDate(time.Now()).Date().Format(constant.DateLayout))
	case "changing to enrollment status withdraw and new date":
		student.EnrollmentStatusAttr = field.NewString(withdrawnEnrollmentStatusEnum)
		student.StatusStartDateAttr = field.NewString(field.NewDate(time.Now()).Date().Format(constant.DateLayout))
	case "changing to enrollment status potential and new date":
		student.EnrollmentStatusAttr = field.NewString(potentialEnrollmentStatusEnum)
		student.StatusStartDateAttr = field.NewString(field.NewDate(time.Now()).Date().Format(constant.DateLayout))
	case "changing to enrollment status LOA and new date":
		student.EnrollmentStatusAttr = field.NewString(loaEnrollmentStatusEnum)
		student.StatusStartDateAttr = field.NewString(field.NewDate(time.Now()).Date().Format(constant.DateLayout))
	case "changing to enrollment status enrolled new date":
		student.EnrollmentStatusAttr = field.NewString(enrolledEnrollmentStatusEnum)
		student.StatusStartDateAttr = field.NewString(field.NewDate(time.Now()).Date().Format(constant.DateLayout))
	case "changing temporary status to enrollment status potential":
		enrollmentStatusStr := student.EnrollmentStatusAttr.String()
		student.EnrollmentStatusAttr = field.NewString(strings.ReplaceAll(enrollmentStatusStr, temporaryEnrollmentStatusEnum, potentialEnrollmentStatusEnum))
	case "adding new location with enrollment status potential with feature date":
		student.LocationAttr = field.NewString("location-id-3")
		student.EnrollmentStatusAttr = field.NewString("1")
		student.StatusStartDateAttr = field.NewString(field.NewDate(time.Now().Add(30 * time.Hour)).Date().Format(constant.DateLayout))
	case "adding new location with enrollment status potential with current date":
		student.LocationAttr = field.NewString("location-id-3")
		student.EnrollmentStatusAttr = field.NewString("1")
		student.StatusStartDateAttr = field.NewString(field.NewDate(time.Now()).Date().Format(constant.DateLayout))
	case "adding new location with enrollment status temporary with feature date":
		student.LocationAttr = field.NewString("location-id-3")
		student.EnrollmentStatusAttr = field.NewString("6")
		student.StatusStartDateAttr = field.NewString(field.NewDate(time.Now().Add(30 * time.Hour)).Date().Format(constant.DateLayout))
	case "adding new location with enrollment status non-potential with feature date":
		student.LocationAttr = field.NewString("location-id-3")
		student.EnrollmentStatusAttr = field.NewString("7")
		student.StatusStartDateAttr = field.NewString(field.NewDate(time.Now().Add(30 * time.Hour)).Date().Format(constant.DateLayout))
	}
}

func (s *suite) studentsWereImportedWithErrorsCollection(ctx context.Context, numberRowsFailed int, numberRowsSuccessful int, tableConditions *godog.Table) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*upb.ImportStudentRequest)
	// err := stepState.ResponseErr
	rows := tableConditions.Rows
	if numberRowsFailed != 0 {
		resp := stepState.Response.(*upb.UpsertStudentResponse)
		errors := resp.Messages
		// if err != nil {
		// 	for _, detail := range status.Convert(err).Details() {
		// 		errorMessages, ok := detail.(*upb.ErrorMessages)
		// 		if !ok {
		// 			return ctx, fmt.Errorf("errorMessages, not ok: detail.(*upb.ErrorMessages), err: %v", err)
		// 		}
		// 		errors = append(errors, errorMessages.Messages...)
		// 	}
		// }

		if len(errors) != numberRowsFailed {
			return ctx, fmt.Errorf("len(errors) != numberRowsFailed, len(errors): %v, numberRowsFailed: %v, err: %v", len(errors), numberRowsFailed, errors)
		}

		for i := 1; i < len(rows); i++ {
			v := rows[i]
			condition := v.Cells[0].Value
			field := v.Cells[1].Value
			code := v.Cells[2].Value
			atRow := v.Cells[3].Value

			csvRow, err := strconv.ParseInt(atRow, 10, 32)
			if err != nil {
				return ctx, fmt.Errorf("csvRow-strconv.Atoi err: %v", err)
			}
			codeErr, err := strconv.ParseInt(code, 10, 32)
			if err != nil {
				return ctx, fmt.Errorf("codeErr-strconv.Atoi err: %v", err)
			}
			for _, error := range errors {
				if error.GetIndex() == int32(csvRow) {
					if error.FieldName != field {
						return ctx, fmt.Errorf("expected: field name is %v, actual: but field name is %v | condition: %s", field, error.FieldName, condition)
					}
					if error.Code != int32(codeErr) {
						return ctx, fmt.Errorf("expected: code is %v, actual: but code is %v | condition: %s", code, error.Code, condition)
					}
				}
			}
		}
	}

	if numberRowsSuccessful != 0 {
		students, err := importstudent.ConvertPayloadToImportStudentData(req.Payload)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("importstudent.ConvertPayloadToImportStudentData err: %v", err)
		}

		insertedStudents := make([]*importstudent.StudentCSV, 0, numberRowsSuccessful)
		failedIndexStudents := make([]int, 0, numberRowsFailed)

		for i := 1; i < len(rows); i++ {
			v := rows[i]
			atRow := v.Cells[3].Value
			csvRow, err := strconv.ParseInt(atRow, 10, 32)
			if err != nil {
				return ctx, fmt.Errorf("csvRow-strconv.Atoi err: %v", err)
			}
			failedIndexStudents = append(failedIndexStudents, int(csvRow-2))
		}

		emails := make([]string, 0, numberRowsSuccessful)
		for idx, student := range students {
			if !utils.InArrayInt(idx, failedIndexStudents) {
				insertedStudents = append(insertedStudents, student)
				emails = append(emails, student.EmailAttr.String())
			}
		}

		userRepo := repository.DomainUserRepo{}

		users, err := userRepo.GetByEmailsInsensitiveCase(ctx, s.BobDBTrace, emails)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("userRepo.GetByEmailsInsensitiveCase err: %v", err)
		}

		if len(users) != numberRowsSuccessful {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected: number of users is %v, actual: but number of users is %v", numberRowsSuccessful, len(users))
		}

		mapEmailAndUserID := make(map[string]string, len(users))
		for _, user := range users {
			mapEmailAndUserID[user.Email().String()] = user.UserID().String()
		}

		domainStudents := make(aggregate.DomainStudents, 0, len(users))
		for i, student := range insertedStudents {
			student.IDAttr = field.NewString(mapEmailAndUserID[student.EmailAttr.String()])
			if !student.IDAttr.IsEmpty() {
				s, _ := importstudent.ToDomainStudentsV2(student, i, true)
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
	if numberRowsFailed == 0 && numberRowsSuccessful == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("numberRowsSuccessful is 0 and numberRowsFailed is 0")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) InitStudentValidationManager() service.StudentValidationManager {
	return service.StudentValidationManager{
		UserRepo:                    &repository.DomainUserRepo{},
		UserGroupRepo:               &repository.DomainUserGroupRepo{},
		LocationRepo:                &repository.DomainLocationRepo{},
		GradeRepo:                   &repository.DomainGradeRepo{},
		SchoolRepo:                  &repository.DomainSchoolRepo{},
		SchoolCourseRepo:            &repository.DomainSchoolCourseRepo{},
		PrefectureRepo:              &repository.DomainPrefectureRepo{},
		TagRepo:                     &repository.DomainTagRepo{},
		InternalConfigurationRepo:   &repository.DomainInternalConfigurationRepo{},
		EnrollmentStatusHistoryRepo: &repository.DomainEnrollmentStatusHistoryRepo{},
		StudentRepo:                 &repository.DomainStudentRepo{},
	}
}
