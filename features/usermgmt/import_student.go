package usermgmt

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
)

const (
	NoData = "no data"
)

var (
	studentEnrollmentStatusMap = map[string]string{
		"1": "STUDENT_ENROLLMENT_STATUS_POTENTIAL",
		"2": "STUDENT_ENROLLMENT_STATUS_ENROLLED",
		"3": "STUDENT_ENROLLMENT_STATUS_WITHDRAWN",
		"4": "STUDENT_ENROLLMENT_STATUS_GRADUATED",
		"5": "STUDENT_ENROLLMENT_STATUS_LOA",
	}

	validCSVHeader                            = `last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postal_code,prefecture,city,first_street,second_street`
	validCSVHeaderWithTag                     = validCSVHeader + ",student_tag"
	validCSVHeaderWithEnrollmentStatusHistory = validCSVHeader + ",status_start_date"
)

func updateManabiePartnerInternalID(ctx context.Context, db database.QueryExecer) error {
	stmt := `UPDATE locations SET partner_internal_id = $1::text WHERE location_id = $2::text`
	_, err := db.Exec(ctx, stmt, ManabiePartnerInternalID, constants.ManabieOrgLocation)
	if err != nil {
		return err
	}
	return nil
}

func (s *suite) aStudentInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := updateManabiePartnerInternalID(ctx, s.BobPostgresDB)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("updateManabiePartnerInternalID err: %v", err)
	}

	num := newID()
	validRowWithFirstNameLastName := fmt.Sprintf(`Student %[1]s 01 Last Name,Student %[1]s 01 First Name,,,student-%[1]s-01-with-first-last-name@example.com,1,0,%[2]s,1999/01/12,1,%[3]s`, num, RandPhoneNumberInVN(0), ManabiePartnerInternalID)
	switch invalidFormat {
	case NoData:
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(``),
		}
	case "1001 rows":
		payload := validCSVHeader
		for i := 0; i < 1001; i++ {
			row := fmt.Sprintf("\nStudent %[1]s Last Name,Student %[1]s First Name,,,student-%[1]s@example.com,5,8,,1999/01/12,2,%[2]s,postal-%[1]s,01,city,,", fmt.Sprintf("%s%d", newID(), i), ManabiePartnerInternalID)
			payload += row
			stepState.InvalidCsvRows = append(stepState.ValidCsvRows, row)
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(payload),
		}
	case "number of column is not equal 11 or 16":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postal_code,prefecture,city,first_street,`),
		}
	case "missing mandatory":
		num := newID()
		stepState.InvalidCsvRows = []string{
			fmt.Sprintf(`,student-%[1]s-01@example.com,1,0,%[2]s,1999/01/12,1,%[3]s`, num, RandPhoneNumberInVN(3), ManabiePartnerInternalID),
			fmt.Sprintf(`Student %[1]s 02,,1,0,%[2]s,1999/01/12,1,%[3]s`, num, RandPhoneNumberInVN(3), ManabiePartnerInternalID),
			fmt.Sprintf(`Student %[1]s 03,student-%[1]s-03@example.com,,16,%[2]s,1999/01/12,2,%[3]s`, num, RandPhoneNumberInVN(4), ManabiePartnerInternalID),
			fmt.Sprintf(`Student %[1]s 04,student-%[1]s-04@example.com,1,,%[2]s,1999/01/12,1,%[3]s`, num, RandPhoneNumberInVN(5), ManabiePartnerInternalID),
			fmt.Sprintf(`,,,,%s,1999/01/12,2,%s`, RandPhoneNumberInVN(6), ManabiePartnerInternalID),
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(`name,email,enrollment_status,grade,phone_number,birthday,gender,location
%s
%s
%s
%s
%s`, stepState.InvalidCsvRows[0], stepState.InvalidCsvRows[1], stepState.InvalidCsvRows[2], stepState.InvalidCsvRows[3], stepState.InvalidCsvRows[4])),
		}
	case "with first name last name and wrong first_name column in header":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`last_name,firstName,first_name_phonetic,last_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location`),
		}
	case "with first name last name and wrong last_name column in header":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`LastName,first_name,first_name_phonetic,last_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location`),
		}
	case "with first name last name and wrong first_name_phonetic column in header":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`last_name,first_name,last_name_phonetic,FirstNamePhonetic,email,enrollment_status,grade,phone_number,birthday,gender,location`),
		}
	case "with first name last name and wrong last_name_phonetic column in header":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`last_name,first_name,LastNamePhonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location`),
		}
	case "with first name last name and wrong email column in header":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,Em4il,enrollment_status,grade,phone_number,birthday,gender,location`),
		}
	case "with first name last name and wrong enrollment_status column in header":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,EnrollmentStatus,grade,phone_number,birthday,gender,location`),
		}
	case "with first name last name and wrong grade column in header":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,Gr4de,phone_number,birthday,gender,location`),
		}
	case "with first name last name and wrong phone_number column in header":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,PhoneNumber,birthday,gender,location`),
		}
	case "with first name last name and wrong birthday column in header":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,BirthDay,gender,location`),
		}
	case "with first name last name and wrong gender column in header":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,GEN1DER,location`),
		}
	case "with first name last name and wrong location column in header":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,Loc4tion`),
		}
	case "with first name last name invalid rows":
		num := newID()
		stepState.InvalidCsvRows = []string{
			fmt.Sprintf(`Student %[1]s 01 Last Name,Student 01 First Name,,,student-%[1]s-01@example..com,1,0,%[2]s,1999/01/12,1,%[3]s`, num, RandPhoneNumberInVN(3), ManabiePartnerInternalID),
			fmt.Sprintf(`Student %[1]s 02 Last Name,Student 02 First Name,,,student-%[1]s-02@example.com,0,16,%[2]s,1999/01/12,2,%[3]s`, num, RandPhoneNumberInVN(4), ManabiePartnerInternalID),
			fmt.Sprintf(`Student %[1]s 03 Last Name,Student %[1]s 03 Last Name,,,student-%[1]s-03@example.com,6,0,%[2]s,1999/01/12,1,%[3]s`, num, RandPhoneNumberInVN(5), ManabiePartnerInternalID),
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location
			%s
			%s
			%s`, stepState.InvalidCsvRows[0], stepState.InvalidCsvRows[1], stepState.InvalidCsvRows[2])),
		}
	case "with first name last name missing mandatory":
		num := newID()
		stepState.InvalidCsvRows = []string{
			fmt.Sprintf(`,Student %[1]s 01 First Name,,,student-%[1]s-01@example.com,1,0,%[2]s,1999/01/12,1,%[3]s`, num, RandPhoneNumberInVN(3), ManabiePartnerInternalID),
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location
%s`, stepState.InvalidCsvRows[0])),
		}

	case "with the missing mandatory location field":
		num := newID()
		stepState.InvalidCsvRows = []string{
			fmt.Sprintf(`Student %[1]s Last Name,Student %[1]s First Name,,,student-%[1]s-01@example.com,1,0,%[2]s,1999/01/12,1,`, num, RandPhoneNumberInVN(3)),
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location
%s`, stepState.InvalidCsvRows[0])),
		}
	case "with first name last name email duplication rows":
		stepState.InvalidCsvRows = []string{
			`Student 02 Last Name,Student 02 First Name,,,student-01@example.com,5,8,0981143302,1999/01/12,2,`,
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location
				Student 01 Last Name,Student 01 First Name,,,student-01@example.com,5,8,0981143301,1999/01/12,2,
				%s`, stepState.InvalidCsvRows[0])),
		}
	case "with first name last name phone_number duplication rows":
		stepState.InvalidCsvRows = []string{
			`Student 02 Last Name,Student 02 First Name,,,student-02@example.com,5,8,0981143301,1999/01/12,2,`,
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location
				Student 01 Last Name,Student 01 First Name,,,student-01@example.com,5,8,0981143301,1999/01/12,2,
				%s`, stepState.InvalidCsvRows[0])),
		}
	case "with first name last name email duplication data rows":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(`last_name,first_name,last_name_phonetic,first_name_phonetic,,email,enrollment_status,grade,phone_number,birthday,gender,location
			%s`, validRowWithFirstNameLastName)),
		}
		_, err := s.importingStudent(StepStateToContext(ctx, stepState), schoolAdminType)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.importingStudent err: %v", err)
		}
		stepState.InvalidCsvRows = []string{
			fmt.Sprintf(`Student %[1]s 01 Last Name,Student %[1]s 01 First Name,,,student-%[1]s-01-with-first-last-name@example.com,1,0,%[2]s,1999/01/12,1,%[3]s`, num, RandPhoneNumberInVN(1), ManabiePartnerInternalID),
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location
			%s`, stepState.InvalidCsvRows[0])),
		}
	case "with home address and wrong postal column in header":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postalCode,prefecture,city,first_street,second_street`),
		}
	case "with home address and wrong prefecture column in header":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postalCode,prefe3ture,city,first_street,second_street`),
		}
	case "with home address and wrong city column in header":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postalCode,prefecture,c4ty,first_street,second_street`),
		}
	case "with home address and wrong first street column in header":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postal_code,prefecture,city,firstStreet,second_street`),
		}
	case "with home address and wrong second street in header":
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postal_code,prefecture,city,first_street,secondStreet`),
		}
	case "with home address with invalid prefecture value":
		num := newID()
		stepState.InvalidCsvRows = []string{
			fmt.Sprintf(`Student %[1]s 01 Last Name,Student %[1]s 01 First Name,Student %[1]s 01 Last Name Phonetic,Student %[1]s 01 First Name Phonetic,student-%[1]s-01@example.com,1,0,%[2]s,1999/01/12,1,,,1000,,,`, num, RandPhoneNumberInVN(1)),
			fmt.Sprintf(`Student %[1]s 02 Last Name,Student %[1]s 02 First Name,Student %[1]s 02 Last Name Phoentic,,student-%[1]s-02@example.com,5,16,%[2]s,1999/01/12,1,,,invalid-prefecture,,,`, num, RandPhoneNumberInVN(2)),
			fmt.Sprintf(`Student %[1]s 03 Last Name,Student %[1]s 03 First Name,,Student %[1]s 03 First Name Phonetic,student-%[1]s-03@example.com,1,7,,1999/01/12,1,,,10a,,,`, num),
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(validCSVHeader+`
			%s
			%s
			%s`, stepState.InvalidCsvRows[0], stepState.InvalidCsvRows[1], stepState.InvalidCsvRows[2])),
		}
	case "with school history with invalid schoolID value":
		num := newID()
		stepState.InvalidCsvRows = []string{
			fmt.Sprintf(`Student %[1]s 01 Last Name,Student %[1]s 01 First Name,Student %[1]s 01 Last Name Phonetic,Student %[1]s 01 First Name Phonetic,student-%[1]s-01@example.com,1,0,%[2]s,1999/01/12,1,,,1000,,,,school-id-1,;,2022-01-02T15:04:05.000Z;2022-11-02T15:04:05.000Z,2023-01-02T15:04:05.000Z;2023-01-02T15:04:05.000Z`, num, RandPhoneNumberInVN(1)),
			fmt.Sprintf(`Student %[1]s 02 Last Name,Student %[1]s 02 First Name,Student %[1]s 02 Last Name Phoentic,,student-%[1]s-02@example.com,5,16,%[2]s,1999/01/12,1,,,invalid-prefecture,,,,school-id-1,;,2022-01-02T15:04:05.000Z;2022-11-02T15:04:05.000Z,2023-01-02T15:04:05.000Z;2023-01-02T15:04:05.000Z`, num, RandPhoneNumberInVN(2)),
			fmt.Sprintf(`Student %[1]s 03 Last Name,Student %[1]s 03 First Name,,Student %[1]s 03 First Name Phonetic,student-%[1]s-03@example.com,1,7,,1999/01/12,1,,,10a,,,,school-id-1,;,2022-01-02T15:04:05.000Z;2022-11-02T15:04:05.000Z,2023-01-02T15:04:05.000Z;2023-01-02T15:04:05.000Z`, num),
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(validCSVHeader+`,school,school_course,start_date,end_date
			%s
			%s
			%s`, stepState.InvalidCsvRows[0], stepState.InvalidCsvRows[1], stepState.InvalidCsvRows[2])),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aStudentValidRequestPayloadWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := updateManabiePartnerInternalID(ctx, s.BobDBTrace)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("updateManabiePartnerInternalID err: %v", err)
	}

	gradeID := ""
	stmt := `select partner_internal_id from grade limit 1`
	if err := s.BobDBTrace.QueryRow(ctx, stmt).Scan(&gradeID); err != nil {
		return ctx, err
	}

	switch rowCondition {
	case "no row":
		stepState.ValidCsvRows = []string{}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(`user_id,last_name,first_name,last_name_phonetic,email,grade,school,school_course,enrollment_status,location`),
		}
	case "only mandatory rows":
		payload := `user_id,first_name,last_name,name,email,enrollment_status,grade,phone_number,birthday,gender,location`
		for i := 0; i < 10; i++ {
			row := fmt.Sprintf(`,last_name,first_name,last_name_phonetic,student-%[1]s@example.com,%[2]s,,,6,manabie-location`, strings.ToLower(newID()), gradeID)
			payload += row
			stepState.ValidCsvRows = append(stepState.ValidCsvRows, row)
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(payload),
		}
	case "1000 rows":
		payload := `user_id,first_name,last_name,name,email,enrollment_status,grade,phone_number,birthday,gender,location`
		for i := 0; i < 1000; i++ {
			row := fmt.Sprintf(`,last_name,first_name,last_name_phonetic,student-%[1]s-02@example.com,%[2]s,,,6,manabie-location`, strings.ToLower(newID()), gradeID)
			payload += row
			stepState.ValidCsvRows = append(stepState.ValidCsvRows, row)
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(payload),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aStudentValidRequestPayloadHomeAddressWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := updateManabiePartnerInternalID(ctx, s.BobDBTrace)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("updateManabiePartnerInternalID err: %v", err)
	}

	switch rowCondition {
	case "no row":
		stepState.ValidCsvRows = []string{}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(validCSVHeader),
		}
	case "only mandatory rows":
		num := newID()
		stepState.ValidCsvRows = []string{
			fmt.Sprintf(`Student %[1]s 01 Last Name,Student %[1]s 01 First Name,,,student-%[1]s-01@example.com,1,0,,,,,,,,,`, num),
			fmt.Sprintf(`Student %[1]s 02 Last Name,Student %[1]s 02 First Name,,,student-%[1]s-02@example.com,5,16,,,,,,,,,`, num),
			fmt.Sprintf(`Student %[1]s 03 Last Name,Student %[1]s 03 First Name,,,student-%[1]s-03@example.com,3,0,,,,,,,,,`, num),
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(validCSVHeader+`
			%s
			%s
			%s`, stepState.ValidCsvRows[0], stepState.ValidCsvRows[1], stepState.ValidCsvRows[2])),
		}
	case "valid rows":
		num := newID()
		stepState.ValidCsvRows = []string{
			fmt.Sprintf(`Student %[1]s 01 Last Name,Student %[1]s 01 First Name,Student %[1]s 01 Last Name Phonetic,Student %[1]s 01 First Name Phonetic,student-%[1]s-01@example.com,1,0,%[2]s,1999/01/12,1,,,,,,`, num, RandPhoneNumberInVN(1)),
			fmt.Sprintf(`Student %[1]s 02 Last Name,Student %[1]s 02 First Name,Student %[1]s 02 Last Name Phoentic,,student-%[1]s-02@example.com,5,16,%[2]s,1999/01/12,1,,7000,01,,,`, num, RandPhoneNumberInVN(2)),
			fmt.Sprintf(`Student %[1]s 03 Last Name,Student %[1]s 03 First Name,,Student %[1]s 03 First Name Phonetic,student-%[1]s-03@example.com,1,7,,1999/01/12,1,,900000,02,,,`, num),
			fmt.Sprintf(`Student %[1]s 04 Last Name,Student %[1]s 04 First Name,,,student-%[1]s-04@example.com,5,0,%[2]s,,2,,2000,03,,,`, num, RandPhoneNumberInVN(4)),
			fmt.Sprintf(`Student %[1]s 05 Last Name,Student %[1]s 05 First Name,,,student-%[1]s-05@example.com,1,16,%[2]s,1999/01/12,,,,,,,`, num, RandPhoneNumberInVN(5)),
			fmt.Sprintf(`Student %[1]s 06 Last Name,Student %[1]s 06 First Name,,,student-%[1]s-06@example.com,5,8,%[2]s,1999/01/12,2,,,,,,`, num, RandPhoneNumberInVN(6)),
			fmt.Sprintf(`Student %[1]s 07 Last Name,Student %[1]s 07 First Name,,,student-%[1]s-07@example.com,3,0,,,,,,,,,`, num),
		}

		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(validCSVHeader+`
			%s
			%s
			%s
			%s
			%s
			%s
			%s`, stepState.ValidCsvRows[0], stepState.ValidCsvRows[1], stepState.ValidCsvRows[2], stepState.ValidCsvRows[3], stepState.ValidCsvRows[4], stepState.ValidCsvRows[5], stepState.ValidCsvRows[6])),
		}
	case "valid row with grade master":
		_, err = s.generateGradeMaster(StepStateToContext(ctx, stepState))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateGradeMaster err: %v", err)
		}
		num := newID()
		stepState.ValidCsvRows = []string{
			fmt.Sprintf(`Student %[1]s 01 Last Name,Student %[1]s 01 First Name,,,student-%[1]s-01@example.com,1,%s,,,,,,,,,`, num, stepState.PartnerInternalIDs[0]),
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(validCSVHeader+`
		%s`, stepState.ValidCsvRows[0])),
		}
	case "valid row with student phone number":
		num := newID()
		_, err = s.generateGradeMaster(StepStateToContext(ctx, stepState))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateGradeMaster err: %v", err)
		}
		stepState.ValidCsvRows = []string{
			fmt.Sprintf(`Student %[1]s 01 Last Name,Student %[1]s 01 First Name,student-%[1]s-01@example.com,1,%s,%s,%s,1`, num, stepState.PartnerInternalIDs[0], randomNumericString(9), randomNumericString(9)),
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(`last_name,first_name,email,enrollment_status,grade,student_phone_number,home_phone_number,contact_preference
		%s`, stepState.ValidCsvRows[0])),
		}
	case "1000 rows":
		payload := validCSVHeader
		for i := 0; i < 1000; i++ {
			row := fmt.Sprintf("\nStudent %[1]s Last Name,Student %[1]s First Name,,,student-%[1]s@example.com,5,8,,1999/01/12,2,%[2]s,postal-%[1]s,01,city,,", fmt.Sprintf("%s%d", newID(), i), ManabiePartnerInternalID)
			payload += row
			stepState.ValidCsvRows = append(stepState.ValidCsvRows, row)
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(payload),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aStudentValidRequestTagWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if err := updateManabiePartnerInternalID(ctx, s.BobDBTrace); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("updateManabiePartnerInternalID err: %v", err)
	}

	var err error
	var tagIDs, tagPartnerIDs []string

	switch rowCondition {
	case "no row":
		stepState.ValidCsvRows = []string{}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(validCSVHeaderWithTag),
		}
	case "only mandatory rows":
		num := newID()
		stepState.ValidCsvRows = []string{
			fmt.Sprintf(`Student %[1]s 01 Last Name,Student %[1]s 01 First Name,,,student-%[1]s-01@example.com,1,0,,,,,,,,,`, num),
			fmt.Sprintf(`Student %[1]s 02 Last Name,Student %[1]s 02 First Name,,,student-%[1]s-02@example.com,5,16,,,,,,,,,`, num),
			fmt.Sprintf(`Student %[1]s 03 Last Name,Student %[1]s 03 First Name,,,student-%[1]s-03@example.com,3,0,,,,,,,,,`, num),
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(validCSVHeader+`
			%s
			%s
			%s`, stepState.ValidCsvRows[0], stepState.ValidCsvRows[1], stepState.ValidCsvRows[2])),
		}
	case "valid row":
		tagIDs, tagPartnerIDs, err = s.createAmountTags(ctx, 3, pb.UserTagType_USER_TAG_TYPE_STUDENT.String(), fmt.Sprint(constants.ManabieSchool))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.createAmountTags: %v", err)
		}

		payload := validCSVHeaderWithTag
		for i := 0; i < 3; i++ {
			row := fmt.Sprintf("\nStudent %[1]s Last Name,Student %[1]s First Name,,,student-%[1]s@example.com,5,8,,1999/01/12,2,%[2]s,postal-%[1]s,01,city,,,%[3]s", fmt.Sprintf("%s%d", newID(), i), ManabiePartnerInternalID, strings.Join(tagPartnerIDs, ";"))
			payload += row
			stepState.ValidCsvRows = append(stepState.ValidCsvRows, row)
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(payload),
		}
	case "1000 rows":
		tagIDs, tagPartnerIDs, err = s.createAmountTags(ctx, 3, pb.UserTagType_USER_TAG_TYPE_STUDENT.String(), fmt.Sprint(constants.ManabieSchool))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.createAmountTags: %v", err)
		}

		payload := validCSVHeaderWithTag
		for i := 0; i < 1000; i++ {
			row := fmt.Sprintf("\nStudent %[1]s Last Name,Student %[1]s First Name,,,student-%[1]s@example.com,5,8,,1999/01/12,2,%[2]s,postal-%[1]s,01,city,,,%[3]s", fmt.Sprintf("%s%d", newID(), i), ManabiePartnerInternalID, strings.Join(tagPartnerIDs, ";"))
			payload += row
			stepState.ValidCsvRows = append(stepState.ValidCsvRows, row)
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(payload),
		}
	}

	stepState.TagIDs = tagIDs
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aStudentValidRequestEnrollmentStatusHistoryWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if err := updateManabiePartnerInternalID(ctx, s.BobDBTrace); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("updateManabiePartnerInternalID err: %v", err)
	}

	switch rowCondition {
	case "no row":
		stepState.ValidCsvRows = []string{}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(validCSVHeaderWithEnrollmentStatusHistory),
		}
	case "only mandatory rows":
		num := newID()
		stepState.ValidCsvRows = []string{
			fmt.Sprintf(`Student %[1]s 01 Last Name,Student %[1]s 01 First Name,,,student-%[1]s-01@example.com,1,0,,,,,,,,,`, num),
			fmt.Sprintf(`Student %[1]s 02 Last Name,Student %[1]s 02 First Name,,,student-%[1]s-02@example.com,5,16,,,,,,,,,`, num),
			fmt.Sprintf(`Student %[1]s 03 Last Name,Student %[1]s 03 First Name,,,student-%[1]s-03@example.com,3,0,,,,,,,,,`, num),
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(validCSVHeader+`
			%s
			%s
			%s`, stepState.ValidCsvRows[0], stepState.ValidCsvRows[1], stepState.ValidCsvRows[2])),
		}
	case "1000 rows":
		payload := validCSVHeaderWithEnrollmentStatusHistory
		for i := 0; i < 1000; i++ {
			row := fmt.Sprintf("\nStudent %[1]s Last Name,Student %[1]s First Name,,,student-%[1]s@example.com,5,8,,1999/01/12,2,%[2]s,postal-%[1]s,01,city,,,2022/01/12", fmt.Sprintf("%s%d", newID(), i), ManabiePartnerInternalID)
			payload += row
			stepState.ValidCsvRows = append(stepState.ValidCsvRows, row)
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(payload),
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aStudentValidRequestPayloadSchoolHistoryWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := updateManabiePartnerInternalID(ctx, s.BobDBTrace)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("updateManabiePartnerInternalID err: %v", err)
	}

	schoolInfo1, err := insertRandomSchoolInfo(ctx, s.BobDBTrace, idutil.ULIDNow())
	if err != nil {
		return nil, fmt.Errorf("insertRandomSchoolInfo: %v", err)
	}
	schoolCourse1, err := insertRandomSchoolCourse(ctx, s.BobDBTrace, schoolInfo1.ID.String)
	if err != nil {
		return nil, fmt.Errorf("insertRandomSchoolInfo: %v", err)
	}

	schoolInfo2, err := insertRandomSchoolInfo(ctx, s.BobDBTrace, idutil.ULIDNow())
	if err != nil {
		return nil, fmt.Errorf("insertRandomSchoolInfo: %v", err)
	}
	schoolCourse2, err := insertRandomSchoolCourse(ctx, s.BobDBTrace, schoolInfo2.ID.String)
	if err != nil {
		return nil, fmt.Errorf("insertRandomSchoolInfo: %v", err)
	}

	schoolInfo3, err := insertRandomSchoolInfo(ctx, s.BobDBTrace, idutil.ULIDNow())
	if err != nil {
		return nil, fmt.Errorf("insertRandomSchoolInfo: %v", err)
	}

	switch rowCondition {
	case "no row":
		stepState.ValidCsvRows = []string{}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(validCSVHeader + `,school,school_course,start_date,end_date`),
		}
	case "only mandatory rows":
		num := newID()
		stepState.ValidCsvRows = []string{
			fmt.Sprintf(`Student %[1]s 01 Last Name,Student %[1]s 01 First Name,,,student-%[1]s-01@example.com,1,0,,,,,,,,,,%[2]s;%[3]s,%[6]s;%[7]s,%[4]s;%[4]s,%[5]s;%[5]s`, num, schoolInfo1.PartnerID.String, schoolInfo2.PartnerID.String, "2016/01/02", "2022/01/02", schoolCourse1.PartnerID.String, schoolCourse2.PartnerID.String),
			fmt.Sprintf(`Student %[1]s 02 Last Name,Student %[1]s 02 First Name,,,student-%[1]s-02@example.com,5,16,,,,,,,,,,%[2]s;%[3]s,;,%[4]s;%[4]s,%[5]s;%[5]s`, num, schoolInfo2.PartnerID.String, schoolInfo3.PartnerID.String, "2016/01/02", "2022/01/02"),
			fmt.Sprintf(`Student %[1]s 03 Last Name,Student %[1]s 03 First Name,,,student-%[1]s-03@example.com,3,0,,,,,,,,,,%[2]s;%[3]s,;,%[4]s;%[4]s,%[5]s;%[5]s`, num, schoolInfo3.PartnerID.String, schoolInfo1.PartnerID.String, "2016/01/02", "2022/01/02"),
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(validCSVHeader+`,school,school_course,start_date,end_date
			%s
			%s
			%s`, stepState.ValidCsvRows[0], stepState.ValidCsvRows[1], stepState.ValidCsvRows[2])),
		}
	case "valid rows":
		num := newID()
		stepState.ValidCsvRows = []string{
			fmt.Sprintf(`Student %[1]s 01 Last Name,Student %[1]s 01 First Name,Student %[1]s 01 Last Name Phonetic,Student %[1]s 01 First Name Phonetic,student-%[1]s-01@example.com,1,0,%[2]s,1999/01/12,1,,,,,,,%[3]s;%[4]s,;,%[5]s;%[5]s,%[6]s;%[6]s`, num, RandPhoneNumberInVN(1), schoolInfo1.PartnerID.String, schoolInfo2.PartnerID.String, "2016/01/02", "2022/01/02"),
			fmt.Sprintf(`Student %[1]s 02 Last Name,Student %[1]s 02 First Name,Student %[1]s 02 Last Name Phoentic,,student-%[1]s-02@example.com,5,16,%[2]s,1999/01/12,1,,7000,01,city-test,,,,,,`, num, RandPhoneNumberInVN(2)),
			fmt.Sprintf(`Student %[1]s 03 Last Name,Student %[1]s 03 First Name,,Student %[1]s 03 First Name Phonetic,student-%[1]s-03@example.com,1,7,,1999/01/12,1,,,,,,,,,,`, num),
			fmt.Sprintf(`Student %[1]s 04 Last Name,Student %[1]s 04 First Name,,,student-%[1]s-04@example.com,5,0,%[2]s,,2,,2000,03,,,,,,,`, num, RandPhoneNumberInVN(4)),
			fmt.Sprintf(`Student %[1]s 05 Last Name,Student %[1]s 05 First Name,,,student-%[1]s-05@example.com,1,16,%[2]s,1999/01/12,,,,,,,,,,,`, num, RandPhoneNumberInVN(5)),
			fmt.Sprintf(`Student %[1]s 06 Last Name,Student %[1]s 06 First Name,,,student-%[1]s-06@example.com,5,8,%[2]s,1999/01/12,2,,,,,,,,,,`, num, RandPhoneNumberInVN(6)),
			fmt.Sprintf(`Student %[1]s 07 Last Name,Student %[1]s 07 First Name,,,student-%[1]s-07@example.com,3,0,,,,,,,,,,,,,`, num),
		}

		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(validCSVHeader+`,school,school_course,start_date,end_date
			%s
			%s
			%s
			%s
			%s
			%s
			%s`, stepState.ValidCsvRows[0], stepState.ValidCsvRows[1], stepState.ValidCsvRows[2], stepState.ValidCsvRows[3], stepState.ValidCsvRows[4], stepState.ValidCsvRows[5], stepState.ValidCsvRows[6])),
		}
	case "valid row with grade master":
		_, err = s.generateGradeMaster(StepStateToContext(ctx, stepState))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateGradeMaster err: %v", err)
		}
		num := newID()
		stepState.ValidCsvRows = []string{
			fmt.Sprintf(`Student %[1]s 01 Last Name,Student %[1]s 01 First Name,,,student-%[1]s-01@example.com,1,%[2]s,,,,,,,,,,%[3]s;%[4]s,;,%[5]s;%[5]s,%[6]s;%[6]s`, num, stepState.PartnerInternalIDs[0], schoolInfo1.ID.String, schoolInfo2.ID.String, "2016/01/02", "2022/01/02"),
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(fmt.Sprintf(validCSVHeader+`,school,school_course,start_date,end_date
		%s`, stepState.ValidCsvRows[0])),
		}
	case "1000 rows":
		payload := validCSVHeader + `,school,school_course,start_date,end_date`
		for i := 0; i < 1000; i++ {
			row := fmt.Sprintf("\nStudent %[1]s Last Name,Student %[1]s First Name,,,student-%[1]s@example.com,5,8,,1999/01/12,2,%[2]s,postal-%[1]s,01,city,,,%[3]s;%[4]s,;,%[5]s;%[5]s,%[6]s;%[6]s", fmt.Sprintf("%s%d", newID(), i), ManabiePartnerInternalID, schoolInfo1.ID.String, schoolInfo2.ID.String, time.Now(), time.Now().Add(time.Hour))
			payload += row
			stepState.ValidCsvRows = append(stepState.ValidCsvRows, row)
		}
		stepState.Request = &pb.ImportStudentRequest{
			Payload: []byte(payload),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidStudentLinesWithStudentPhoneNumberAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	const (
		LastName = iota
		FirstName
		Email
		EnrollmentStatus
		Grade
		StudentPhoneNumber
		StudentHomePhoneNumber
	)
	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	resp := stepState.Response.(*pb.ImportStudentResponse)
	if len(resp.Errors) > 0 && len(stepState.InvalidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected resp.Errors is [], but actual resp is %v", resp)
	}
	schoolID := fmt.Sprint(constants.ManabieSchool)
	userRepo := &repository.UserRepo{}
	studentRepo := &repository.StudentRepo{}
	studentPhoneNumberRepo := &repository.UserPhoneNumberRepo{}
	stepState.EvtImportStudents = make([]*pb.EvtImportStudent_ImportStudent, 0, len(stepState.ValidCsvRows))
	userIDs := make([]string, 0, len(stepState.ValidCsvRows))

	for _, row := range stepState.ValidCsvRows {
		rowSplit := strings.Split(row, ",")
		users, err := userRepo.GetByEmail(ctx, s.BobDBTrace, database.TextArray([]string{rowSplit[Email]}))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("userRepo.GetByEmail err: %v", err)
		}
		if len(users) != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found student with email = %s", rowSplit[Email])
		}

		user := users[0]
		student, err := studentRepo.Find(ctx, s.BobDBTrace, user.ID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("studentRepo.Find err: %v", err)
		}

		if rowSplit[Grade] != "" {
			grade, err := strconv.Atoi(strings.TrimSpace(rowSplit[Grade]))
			if err != nil {
				if err := validateGradeMaster(ctx, s.BobDBTrace, student.GradeID.String, student.ID.String); err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("validateGradeMaster: %v", err)
				}
			} else {
				if student.CurrentGrade.Int != int16(grade) {
					return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected grade is %v, actual grade is %v", grade, student.CurrentGrade.Int)
				}
			}
		}

		if user.FirstName.String != rowSplit[FirstName] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected first name is %v, actual first name is %v", rowSplit[FirstName], user.FirstName.String)
		}

		if user.LastName.String != rowSplit[LastName] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected last name is %v, actual last name is %v", rowSplit[LastName], user.LastName.String)
		}

		if user.FullName.String != helper.CombineFirstNameAndLastNameToFullName(strings.TrimSpace(rowSplit[FirstName]), strings.TrimSpace(rowSplit[LastName])) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected name is %v, actual name is %v", helper.CombineFirstNameAndLastNameToFullName(strings.TrimSpace(rowSplit[FirstName]), strings.TrimSpace(rowSplit[LastName])), user.FullName.String)
		}

		if user.Group.String != entity.UserGroupStudent {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected group is %v, actual group is %v", entity.UserGroupStudent, user.Group.String)
		}

		enrollmentStatus := studentEnrollmentStatusMap[strings.TrimSpace(rowSplit[EnrollmentStatus])]
		if student.EnrollmentStatus.String != enrollmentStatus {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected enrollment_status is %v, actual enrollment_status is %v", enrollmentStatus, student.EnrollmentStatus.String)
		}

		if user.ResourcePath.String != schoolID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected resource_path is %v, actual resource_path is %v", constants.ManabieSchool, user.ResourcePath.String)
		}

		userPhoneNumbers, err := studentPhoneNumberRepo.FindByUserID(ctx, s.BobDBTrace, user.ID)
		var userPhoneNumber *entity.UserPhoneNumber
		var userHomePhoneNumber *entity.UserPhoneNumber
		for _, phoneNumber := range userPhoneNumbers {
			switch phoneNumber.PhoneNumberType.String {
			case entity.StudentPhoneNumber:
				userPhoneNumber = phoneNumber
			case entity.StudentHomePhoneNumber:
				userHomePhoneNumber = phoneNumber
			}
		}
		if userPhoneNumber.PhoneNumber.String != rowSplit[StudentPhoneNumber] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected student phone number is %v, actual student phone number is %v", rowSplit[StudentPhoneNumber], userPhoneNumber.PhoneNumber.String)
		}

		if userHomePhoneNumber.PhoneNumber.String != rowSplit[StudentHomePhoneNumber] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected student home phone number is %v, actual student home phone number is %v", rowSplit[StudentHomePhoneNumber], userHomePhoneNumber.PhoneNumber.String)
		}

		stepState.EvtImportStudents = append(stepState.EvtImportStudents, &pb.EvtImportStudent_ImportStudent{
			StudentId:   student.ID.String,
			StudentName: user.GetName(),
			SchoolId:    fmt.Sprint(student.SchoolID.Int),
		})
		userIDs = append(userIDs, user.ID.String)
	}

	if err := s.validateUsersHasUserGroupWithRole(ctx, userIDs, schoolID, constant.RoleStudent); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validateUsersHasUserGroupWithRole: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidStudentLinesWithHomeAddressAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	const (
		LastName = iota
		FirstName
		LastNamePhonetic
		FirstNamePhonetic
		Email
		EnrollmentStatus
		Grade
		PhoneNumber
		Birthday
		Gender
		Location
	)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	resp := stepState.Response.(*pb.ImportStudentResponse)
	if len(resp.Errors) > 0 && len(stepState.InvalidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected resp.Errors is [], but actual resp is %v", resp)
	}

	schoolID := fmt.Sprint(constants.ManabieSchool)
	userRepo := &repository.UserRepo{}
	studentRepo := &repository.StudentRepo{}
	stepState.EvtImportStudents = make([]*pb.EvtImportStudent_ImportStudent, 0, len(stepState.ValidCsvRows))
	userIDs := make([]string, 0, len(stepState.ValidCsvRows))

	for _, row := range stepState.ValidCsvRows {
		rowSplit := strings.Split(row, ",")
		users, err := userRepo.GetByEmail(ctx, s.BobDBTrace, database.TextArray([]string{rowSplit[Email]}))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("userRepo.GetByEmail err: %v", err)
		}
		if len(users) != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found student with email = %s", rowSplit[Email])
		}

		user := users[0]
		student, err := studentRepo.Find(ctx, s.BobDBTrace, user.ID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("studentRepo.Find err: %v", err)
		}

		if rowSplit[Grade] != "" {
			grade, err := strconv.Atoi(strings.TrimSpace(rowSplit[Grade]))
			if err != nil {
				if err := validateGradeMaster(ctx, s.BobDBTrace, student.GradeID.String, student.ID.String); err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("validateGradeMaster: %v", err)
				}
			} else {
				if student.CurrentGrade.Int != int16(grade) {
					return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected grade is %v, actual grade is %v", grade, student.CurrentGrade.Int)
				}
			}
		}

		if user.FirstName.String != rowSplit[FirstName] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected first name is %v, actual first name is %v", rowSplit[FirstName], user.FirstName.String)
		}

		if user.LastName.String != rowSplit[LastName] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected last name is %v, actual last name is %v", rowSplit[LastName], user.LastName.String)
		}

		if user.FirstNamePhonetic.String != rowSplit[FirstNamePhonetic] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected first name phonetic is %v, actual first name phonetic is %v", rowSplit[FirstNamePhonetic], user.FirstNamePhonetic.String)
		}

		if user.LastNamePhonetic.String != rowSplit[LastNamePhonetic] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected last name phonetic is %v, actual last name phonetic is %v", rowSplit[LastNamePhonetic], user.LastNamePhonetic.String)
		}

		if user.FullName.String != helper.CombineFirstNameAndLastNameToFullName(strings.TrimSpace(rowSplit[FirstName]), strings.TrimSpace(rowSplit[LastName])) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected name is %v, actual name is %v", helper.CombineFirstNameAndLastNameToFullName(strings.TrimSpace(rowSplit[FirstName]), strings.TrimSpace(rowSplit[LastName])), user.FullName.String)
		}

		if user.PhoneNumber.String != rowSplit[PhoneNumber] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected phone_number is %v, actual phone_number is %v", rowSplit[PhoneNumber], user.PhoneNumber.String)
		}

		if user.Group.String != entity.UserGroupStudent {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected group is %v, actual group is %v", entity.UserGroupStudent, user.Group.String)
		}

		enrollmentStatus := studentEnrollmentStatusMap[strings.TrimSpace(rowSplit[EnrollmentStatus])]
		if student.EnrollmentStatus.String != enrollmentStatus {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected enrollment_status is %v, actual enrollment_status is %v", enrollmentStatus, student.EnrollmentStatus.String)
		}

		var gender string
		if rowSplit[Gender] != "" {
			genderInt, err := strconv.Atoi(strings.TrimSpace(rowSplit[Gender]))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("strconv.Atoi err: %v", err)
			}
			gender = pb.Gender(genderInt).String()
		}
		if user.Gender.String != gender {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected gender is %v, actual gender is %v", gender, user.Gender.String)
		}

		var birthday time.Time
		if rowSplit[Birthday] != "" {
			birthday, err = time.Parse("2006/01/02", rowSplit[Birthday])
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("time.Parse err: %v", err)
			}
		} else if user.Birthday.Status != pgtype.Null {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected birthday is nil, actual birthday is %v", user.Birthday)
		}
		if user.Birthday.Time != birthday {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected birthday is %v, actual birthday is %v", birthday, user.Birthday.Time)
		}

		var locationIDs []string
		if strings.TrimSpace(rowSplit[Location]) != "" {
			locationIDs = strings.Split(strings.TrimSpace(rowSplit[Location]), ";")
			err = s.validateLocation(ctx, student, locationIDs)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("s.validateLocation err: %v", err)
			}
		}

		if user.ResourcePath.String != schoolID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected resource_path is %v, actual resource_path is %v", constants.ManabieSchool, user.ResourcePath.String)
		}

		stepState.EvtImportStudents = append(stepState.EvtImportStudents, &pb.EvtImportStudent_ImportStudent{
			StudentId:   student.ID.String,
			StudentName: user.GetName(),
			SchoolId:    fmt.Sprint(student.SchoolID.Int),
			LocationIds: locationIDs,
		})
		userIDs = append(userIDs, user.ID.String)
	}

	if err := s.validateUsersHasUserGroupWithRole(ctx, userIDs, schoolID, constant.RoleStudent); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validateUsersHasUserGroupWithRole: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidStudentLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	const (
		Name = iota
		EnrollmentStatus
		Grade
		PhoneNumber
		Email
		Birthday
		Gender   = 9
		Location = 10
	)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	switch resp := stepState.Response.(type) {
	case *pb.ImportStudentResponse:
		if len(resp.Errors) > 0 && len(stepState.InvalidCsvRows) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected resp.Errors is [], but actual resp is %v", resp)
		}
	case *pb.UpsertStudentResponse:
		if len(resp.Messages) > 0 && len(stepState.InvalidCsvRows) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected resp.Errors is [], but actual resp is %v", resp)
		}
	}

	schoolID := fmt.Sprint(constants.ManabieSchool)
	userRepo := &repository.UserRepo{}
	studentRepo := &repository.StudentRepo{}
	stepState.EvtImportStudents = make([]*pb.EvtImportStudent_ImportStudent, 0, len(stepState.ValidCsvRows))
	userIDs := make([]string, 0, len(stepState.ValidCsvRows))

	for _, row := range stepState.ValidCsvRows {
		rowSplit := strings.Split(row, ",")
		users, err := userRepo.GetByEmail(ctx, s.BobDBTrace, database.TextArray([]string{rowSplit[Email]}))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("userRepo.GetByEmail err: %v", err)
		}
		if len(users) != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found student with email = %s", rowSplit[Email])
		}

		user := users[0]
		student, err := studentRepo.Find(ctx, s.BobDBTrace, user.ID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("studentRepo.Find err: %v", err)
		}

		if rowSplit[Grade] != "" {
			grade, err := strconv.Atoi(strings.TrimSpace(rowSplit[Grade]))
			if err != nil {
				if err := validateGradeMaster(ctx, s.BobDBTrace, student.GradeID.String, student.ID.String); err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("validateGradeMaster: %v", err)
				}
			} else {
				if student.CurrentGrade.Int != int16(grade) {
					return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected grade is %v, actual grade is %v", grade, student.CurrentGrade.Int)
				}
			}
		}

		if user.FullName.String != strings.TrimSpace(rowSplit[Name]) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected name is %v, actual name is %v", strings.TrimSpace(rowSplit[Name]), user.FullName.String)
		}

		if user.PhoneNumber.String != rowSplit[PhoneNumber] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected phone_number is %v, actual phone_number is %v", rowSplit[PhoneNumber], user.PhoneNumber.String)
		}

		if user.Group.String != entity.UserGroupStudent {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected group is %v, actual group is %v", entity.UserGroupStudent, user.Group.String)
		}

		enrollmentStatus := studentEnrollmentStatusMap[strings.TrimSpace(rowSplit[EnrollmentStatus])]
		if student.EnrollmentStatus.String != enrollmentStatus {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected enrollment_status is %v, actual enrollment_status is %v", enrollmentStatus, student.EnrollmentStatus.String)
		}

		var gender string
		if rowSplit[Gender] != "" {
			genderInt, err := strconv.Atoi(strings.TrimSpace(rowSplit[Gender]))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("strconv.Atoi err: %v", err)
			}
			gender = pb.Gender(genderInt).String()
		}
		if user.Gender.String != gender {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected gender is %v, actual gender is %v", gender, user.Gender.String)
		}

		var birthday time.Time
		if rowSplit[Birthday] != "" {
			birthday, err = time.Parse("2006/01/02", rowSplit[Birthday])
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("time.Parse err: %v", err)
			}
		} else if user.Birthday.Status != pgtype.Null {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected birthday is nil, actual birthday is %v", user.Birthday)
		}
		if user.Birthday.Time != birthday {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected birthday is %v, actual birthday is %v", birthday, user.Birthday.Time)
		}

		var locationIDs []string
		if strings.TrimSpace(rowSplit[Location]) != "" {
			locationIDs = strings.Split(strings.TrimSpace(rowSplit[Location]), ";")
			err = s.validateLocation(ctx, student, locationIDs)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("s.validateLocation err: %v", err)
			}
		}

		if user.ResourcePath.String != schoolID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: expected resource_path is %v, actual resource_path is %v", constants.ManabieSchool, user.ResourcePath.String)
		}

		stepState.EvtImportStudents = append(stepState.EvtImportStudents, &pb.EvtImportStudent_ImportStudent{
			StudentId:   student.ID.String,
			StudentName: user.GetName(),
			SchoolId:    fmt.Sprint(student.SchoolID.Int),
			LocationIds: locationIDs,
		})
		userIDs = append(userIDs, user.ID.String)
	}

	if err := s.validateUsersHasUserGroupWithRole(ctx, userIDs, schoolID, constant.RoleStudent); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validateUsersHasUserGroupWithRole: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingStudent(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewStudentServiceClient(s.UserMgmtConn).
		ImportStudent(contextWithToken(ctx), stepState.Request.(*pb.ImportStudentRequest))

	return StepStateToContext(ctx, stepState), nil
}

func RandPhoneNumberInVN(index int) string {
	random := fmt.Sprint(time.Now().UnixMicro() + int64(index))
	return fmt.Sprintf("+8498%s", random[len(random)-7:])
}

func (s *suite) theInvalidStudentLinesAreReturnedWithError(ctx context.Context, errCode string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp := stepState.Response.(*pb.ImportStudentResponse)

	if stepState.ResponseErr != nil {
		if !strings.Contains(stepState.ResponseErr.Error(), errCode) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("test failed: expected %v contains %v", stepState.ResponseErr.Error(), errCode)
		}
	} else {
		if len(resp.Errors) != len(stepState.InvalidCsvRows) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("test failed: expected total errors is %v, actual is %v, %v", len(stepState.InvalidCsvRows), len(resp.Errors), resp.Errors)
		}

		for i := range resp.Errors {
			if resp.Errors[i].Error != errCode {
				return StepStateToContext(ctx, stepState), fmt.Errorf("test failed: expected error code is %v, actual is %v", errCode, resp.Errors[i].Error)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validateLocation(ctx context.Context, student *entity.LegacyStudent, locationIDs []string) error {
	ctx = contextWithToken(ctx)

	stmt :=
		`
	SELECT DISTINCT
		uap.user_id,
		uap.location_id,
		uap.access_path,
		uap.resource_path,
		l.partner_internal_id
	FROM
		user_access_paths uap,
		locations l
	WHERE 
		l.location_id = uap.location_id
		AND uap.user_id = $1
		AND l.partner_internal_id = ANY ($2)
		AND uap.resource_path = $3
	`
	ids := pgtype.TextArray{}
	if err := ids.Set(locationIDs); err != nil {
		return err
	}

	rows, err := s.BobDBTrace.Query(
		ctx,
		stmt,
		student.ID.String,
		ids,
		fmt.Sprint(constants.ManabieSchool),
	)
	if err != nil {
		return fmt.Errorf("validateLocation: query locations stored fail %s", err.Error())
	}
	defer rows.Close()
	userAccessPaths := []*entity.UserAccessPath{}
	var PartnerInternalID pgtype.Text

	for rows.Next() {
		uap := &entity.UserAccessPath{}
		if err := rows.Scan(
			&uap.UserID,
			&uap.LocationID,
			&uap.AccessPath,
			&uap.ResourcePath,
			&PartnerInternalID,
		); err != nil {
			return err
		}

		userAccessPaths = append(userAccessPaths, uap)
	}

	if len(userAccessPaths) != len(locationIDs) {
		return fmt.Errorf("validateLocation fail: expect stored %d user_access_paths, but actual %d", len(locationIDs), len(userAccessPaths))
	}

	for _, uap := range userAccessPaths {
		if uap.UserID.String != student.ID.String {
			return fmt.Errorf("validateLocation fail: user_id stored not equal, expected: %s but actual: %s", student.ID.String, uap.UserID.String)
		}

		if !golibs.InArrayString(PartnerInternalID.String, locationIDs) {
			return fmt.Errorf("validateLocation fail: location_id %s stored not in locationIDs request", PartnerInternalID.String)
		}

		if uap.ResourcePath.String != fmt.Sprint(constants.ManabieSchool) {
			return fmt.Errorf("validateLocation fail: resource_path stored not equal, expected: %v but actual: %v", constants.ManabieSchool, uap.ResourcePath.String)
		}
	}

	return nil
}

func (s *suite) userTagsAreImported(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	countTaggedUsers := 0
	if err := s.BobDB.QueryRow(
		ctx,
		`
		SELECT count(*)
		FROM tagged_user
		WHERE tag_id = ANY($1) AND
		      deleted_at IS NULL
		`,
		database.TextArray(stepState.TagIDs),
	).Scan(&countTaggedUsers); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	expectedTaggedUsers := len(stepState.TagIDs) * len(stepState.ValidCsvRows)
	if countTaggedUsers != expectedTaggedUsers {
		return StepStateToContext(ctx, stepState), fmt.Errorf(
			"imported student dont have enough tag, expecting %d tagged user returned, got %d",
			expectedTaggedUsers,
			countTaggedUsers,
		)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) enrollmentStatusHistoryAreImported(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	resp := stepState.Response.(*pb.ImportStudentResponse)
	if len(resp.Errors) > 0 && len(stepState.InvalidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected resp.Errors is [], but actual resp is %v", resp)
	}
	locationID := stepState.LocationID
	userRepo := &repository.UserRepo{}
	enrollmentStatusHistoryRepo := &repository.DomainEnrollmentStatusHistoryRepo{}
	stepState.EvtImportStudents = make([]*pb.EvtImportStudent_ImportStudent, 0, len(stepState.ValidCsvRows))

	for _, row := range stepState.ValidCsvRows {
		rowSplit := strings.Split(row, ",")
		users, err := userRepo.GetByEmail(ctx, s.BobDBTrace, database.TextArray([]string{rowSplit[4]}))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("userRepo.GetByEmail err: %v", err)
		}
		if len(users) != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found student with email = %s", rowSplit[4])
		}

		user := users[0]
		locations, err := (&repository.DomainLocationRepo{}).GetByPartnerInternalIDs(ctx, s.BobPostgresDB, []string{locationID})
		if err != nil {
			return ctx, fmt.Errorf("(&repository.DomainLocationRepo{}).GetByPartnerInternalIDs err: %v", err)
		}

		enrollmentStatusHistoryRes, err := enrollmentStatusHistoryRepo.
			GetByStudentIDAndLocationID(ctx, s.BobPostgresDB,
				user.UserID.String,
				locations[0].LocationID().String(),
				false,
			)
		if err != nil {
			return ctx, fmt.Errorf("(&repository.DomainEnrollmentStatusHistoryRepo{}).GetByStudentIDAndLocationID err: %v", err)
		}
		if len(enrollmentStatusHistoryRes) != 0 {
			return StepStateToContext(ctx, stepState), nil
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
