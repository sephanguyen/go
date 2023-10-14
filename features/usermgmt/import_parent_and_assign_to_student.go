package usermgmt

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/gocarina/gocsv"
	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) aParentValidRequestPayloadWith(numParent, numStudent int, rowCondition string) (*pb.ImportParentsAndAssignToStudentRequest, error) {
	// open sample csv file
	path := fmt.Sprintf(
		"usermgmt/testdata/csv/parents/import_%d_parents_and_assign_to_%d_students_%s.csv",
		numParent, numStudent, strings.Join(strings.Split(rowCondition, " "), "_"),
	)
	csv, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "os.ReadFile")
	}

	parentCSVRows := make([]*service.ImportParentCSVField, 0)
	if err := gocsv.UnmarshalBytes(csv, &parentCSVRows); err != nil {
		return nil, errors.Wrap(err, "gocsv.UnmarshalBytes")
	}

	randomID := idutil.ULIDNow()
	for idx, parentCSVRow := range parentCSVRows {
		parentCSVRows[idx] = updateParentCSVRowByCondition(parentCSVRow, rowCondition, randomID, idx)
	}
	bytes, err := gocsv.MarshalBytes(parentCSVRows)
	if err != nil {
		return nil, errors.Wrap(err, "gocsv.MarshalBytes")
	}

	req := &pb.ImportParentsAndAssignToStudentRequest{Payload: bytes}
	s.Request = req

	csvLines := make([]string, 0, len(parentCSVRows))
	for _, line := range strings.Split(string(bytes), "\n") {
		if line != "" {
			csvLines = append(csvLines, line)
		}
	}

	numberInvalidRows, err := getNumberInvalidRows(rowCondition)
	if err != nil {
		return nil, errors.Wrap(err, "getNumberInvalidRows")
	}

	if numParent == 1001 {
		numberInvalidRows++
	}

	s.NumberInvalidCsvRows = numberInvalidRows
	s.NumberValidCsvRows = len(csvLines[1:]) // exclude header line

	return req, nil
}

func updateParentCSVRowByCondition(row *service.ImportParentCSVField, rowCondition string, randomID string, idx int) *service.ImportParentCSVField {
	if !golibs.InArrayString(rowCondition,
		[]string{
			"have 1 invalid row with external_user_id duplicated in payload",
			"have 1 invalid row with external_user_id existed in database",
		},
	) {
		// if rowCondition not in list above, then update external_user_id
		if strings.TrimSpace(row.ExternalUserID.String()) != "" {
			row.ExternalUserID.Text = fmt.Sprintf("%s.%d", randomID, idx)
		}
	}

	if !golibs.InArrayString(rowCondition,
		[]string{
			"have 1 invalid row with email duplicated in payload",
			"have 1 invalid row with email existed in database",
		},
	) {
		// if rowCondition not in list above, then update email
		if strings.TrimSpace(row.Email.String()) != "" {
			row.Email.Text = fmt.Sprintf("email%s.u%d@manabie.com", randomID, idx)
		}
	}

	if !golibs.InArrayString(rowCondition,
		[]string{
			"have 1 invalid row with existing username and upper case",
			"have 1 invalid row with existing username",
		},
	) {
		// if rowCondition not in list above, then update username
		if strings.TrimSpace(row.UserName.String()) != "" {
			row.UserName.Text = fmt.Sprintf("username%su%d", randomID, idx) + row.UserName.Text
		}
	}
	return row
}

func (s *suite) theValidParentLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	if s.ResponseErr != nil {
		return ctx, s.ResponseErr
	}

	resp := s.Response.(*pb.ImportParentsAndAssignToStudentResponse)
	if len(resp.Errors) > 0 && s.NumberInvalidCsvRows == 0 {
		return ctx, fmt.Errorf("expected resp.Errors is [], but actual resp is %v", resp)
	}

	userRepo := &repository.UserRepo{}
	parentRepo := &repository.ParentRepo{}
	parentIDs := make([]string, 0, s.NumberValidCsvRows)
	schoolID := fmt.Sprint(constants.ManabieSchool)

	request := s.Request.(*pb.ImportParentsAndAssignToStudentRequest)
	parentCSVRows := make([]*service.ImportParentCSVField, 0)
	if err := gocsv.UnmarshalBytes(request.Payload, &parentCSVRows); err != nil {
		return ctx, errors.Wrap(err, "gocsv.UnmarshalBytes")
	}

	for _, parentCSVRow := range parentCSVRows {
		users, err := userRepo.GetByEmail(ctx, s.BobDBTrace, database.TextArray([]string{parentCSVRow.Email.String()}))
		if err != nil {
			return ctx, fmt.Errorf("userRepo.GetByEmail err: %v", err)
		}
		if len(users) == 0 {
			return ctx, fmt.Errorf("not found parent with email = %s", parentCSVRow.Email.String())
		}

		user := users[0]
		parent, err := parentRepo.GetByID(ctx, s.BobDBTrace, user.ID)
		if err != nil {
			return ctx, fmt.Errorf("parentRepo.Find err: %v", err)
		}

		fullName := utils.CombineFirstNameAndLastNameToFullName(parentCSVRow.FirstName.String(), parentCSVRow.LastName.String())
		if user.FullName.String != strings.TrimSpace(fullName) {
			return ctx, fmt.Errorf("failed to import valid csv row: expected name is %v, actual name is %v", strings.TrimSpace(fullName), user.FullName.String)
		}

		// todo: assert primary phone number and secondary phone number
		if user.PhoneNumber.String != parentCSVRow.PhoneNumber.String() {
			return ctx, fmt.Errorf("failed to import valid csv row: expected phone_number is %v, actual phone_number is %v", parentCSVRow.PhoneNumber.String(), user.PhoneNumber.String)
		}

		if user.Group.String != entity.UserGroupParent {
			return ctx, fmt.Errorf("failed to import valid csv row: expected group is %v, actual group is %v", entity.UserGroupParent, user.Group.String)
		}

		if user.Remarks.String != parentCSVRow.Remarks.String() {
			return ctx, fmt.Errorf("failed to import valid csv row: expected remarks is %v, actual remarks is %v", parentCSVRow.Remarks.String(), user.Remarks.String)
		}

		if user.FirstNamePhonetic.String != parentCSVRow.FirstNamePhonetic.String() {
			return ctx, fmt.Errorf("failed to import valid csv row: expected first_name_phonetic is %v, actual first_name_phonetic is %v", parentCSVRow.FirstNamePhonetic.String(), user.FirstNamePhonetic.String)
		}

		if user.LastNamePhonetic.String != parentCSVRow.LastNamePhonetic.String() {
			return ctx, fmt.Errorf("failed to import valid csv row: expected last_name_phonetic is %v, actual last_name_phonetic is %v", parentCSVRow.LastNamePhonetic.String(), user.LastNamePhonetic.String)
		}

		if user.ExternalUserID.String != parentCSVRow.ExternalUserID.String() {
			return ctx, fmt.Errorf("failed to import valid csv row: expected external_user_id is %v, actual external_user_id is %v", parentCSVRow.ExternalUserID.String(), user.ExternalUserID.String)
		}
		if user.UserRole.String != string(constant.UserRoleParent) {
			return ctx, fmt.Errorf("failed to import valid csv row: expected user_role is %v, actual is %v", constant.UserRoleParent, user.UserRole.String)
		}

		relationships := strings.Split(strings.TrimSpace(parentCSVRow.Relationship.String()), ";")
		err = s.validateRelationshipAndLocation(ctx, parent, relationships)
		if err != nil {
			return ctx, fmt.Errorf("s.validateRelationship err: %v", err)
		}

		parentIDs = append(parentIDs, user.GetUID())
	}

	if err := s.validateUsersHasUserGroupWithRole(ctx, parentIDs, schoolID, constant.RoleParent); err != nil {
		return ctx, fmt.Errorf("validateUsersHasUserGroupWithRole: %v", err)
	}

	if _, err := s.userTagsAreImported(ctx); err != nil {
		return ctx, fmt.Errorf("s.userTagsAreImported: %v", err)
	}

	return ctx, nil
}

func (s *suite) theInvalidParentLinesAreReturnedWithError(ctx context.Context, errCode string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp := stepState.Response.(*pb.ImportParentsAndAssignToStudentResponse)

	if stepState.ResponseErr != nil {
		if !strings.Contains(stepState.ResponseErr.Error(), errCode) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("test failed: expected %v contains %v", stepState.ResponseErr.Error(), errCode)
		}
	} else {
		if len(resp.Errors) != stepState.NumberInvalidCsvRows {
			return StepStateToContext(ctx, stepState), fmt.Errorf("test failed: expected total errors is %v, actual is %v, %v", stepState.NumberInvalidCsvRows, len(resp.Errors), resp.Errors)
		}

		for i := range resp.Errors {
			if resp.Errors[i].Error != errCode {
				return StepStateToContext(ctx, stepState), fmt.Errorf("test failed: expected error code is %v, actual is %v - %v", errCode, resp.Errors[i].Error, resp.Errors[i].FieldName)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingParentWithValidPayload(ctx context.Context, _ string, numberParent, numberStudent int, condition string) (context.Context, error) {
	payload, err := s.aParentValidRequestPayloadWith(numberParent, numberStudent, condition)
	if err != nil {
		return ctx, err
	}

	s.Response, s.ResponseErr = pb.NewUserModifierServiceClient(s.UserMgmtConn).ImportParentsAndAssignToStudent(ctx, payload)
	return ctx, nil
}

func (s *suite) validateRelationshipAndLocation(ctx context.Context, parent *entity.Parent, relationships []string) error {
	ctx = contextWithToken(ctx)

	arrRelationships := pgtype.TextArray{}
	studentParents := make([]*entity.StudentParent, 0)
	studentIDs := make([]string, 0)

	stmt := `
		SELECT sp.student_id, sp.parent_id, sp.relationship, sp.resource_path
		FROM student_parents sp
		WHERE sp.parent_id = $1
			AND sp.relationship = ANY ($2)
			AND sp.resource_path = $3
	`
	if err := arrRelationships.Set(relationships); err != nil {
		return err
	}

	rows, err := s.BobDBTrace.Query(ctx, stmt, parent.ID.String, arrRelationships, fmt.Sprint(constants.ManabieSchool))
	if err != nil {
		return fmt.Errorf("validateRelationship: query student parent stored fail %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		sp := &entity.StudentParent{}
		if err := rows.Scan(&sp.StudentID, &sp.ParentID, &sp.Relationship); err != nil {
			return err
		}

		studentIDs = append(studentIDs, sp.StudentID.String)
		studentParents = append(studentParents, sp)
	}

	for _, sp := range studentParents {
		if sp.ParentID.String != parent.ID.String {
			return fmt.Errorf("validateRelationship fail: parent_id stored not equal, expected: %s but actual: %s", parent.ID.String, sp.ParentID.String)
		}

		if !golibs.InArrayString(sp.Relationship.String, relationships) {
			return fmt.Errorf("validateRelationship fail: relationship %s stored not in relationships request", sp.Relationship.String)
		}
	}

	// Validate locations
	const (
		studentLocationsSTMT = `
			SELECT uap.location_id
			FROM user_access_paths uap
			WHERE uap.user_id = ANY ($1)
				AND uap.resource_path = $2
				AND uap.deleted_at IS NULL
		`
		parentLocationsSTMT = `
			SELECT uap.location_id
			FROM user_access_paths uap
			WHERE uap.user_id = $1
				AND uap.resource_path = $2
				AND uap.deleted_at IS NULL
		`
	)

	studentLocations := make([]string, 0)
	parentLocations := map[string]string{}

	rows, err = s.BobDBTrace.Query(ctx, studentLocationsSTMT, studentIDs, fmt.Sprint(constants.ManabieSchool))
	if err != nil {
		return fmt.Errorf("validateRelationship: query student parent stored fail %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		studentLocation := pgtype.Text{}
		if err := rows.Scan(&studentLocation); err != nil {
			return err
		}

		studentLocations = append(studentLocations, studentLocation.String)
	}

	rows, err = s.BobDBTrace.Query(ctx, parentLocationsSTMT, parent.ID.String, fmt.Sprint(constants.ManabieSchool))
	if err != nil {
		return fmt.Errorf("validateRelationship: query student parent stored fail %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		parentLocation := pgtype.Text{}
		if err := rows.Scan(&parentLocation); err != nil {
			return err
		}

		parentLocations[parentLocation.String] = ""
	}

	for _, locationOfStudent := range studentLocations {
		_, ok := parentLocations[locationOfStudent]
		if !ok {
			return fmt.Errorf("locations doesn't apply to parents, expected: %s ", locationOfStudent)
		}
	}

	return nil
}

func (s *suite) createNewStudent(ctx context.Context, req *pb.CreateStudentRequest) (*pb.CreateStudentResponse, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.createStudentSubscription(ctx, stepState.Request)
	if err != nil {
		return nil, fmt.Errorf("s.createStudentSubscription: %w", err)
	}

	resp, err := pb.NewUserModifierServiceClient(s.UserMgmtConn).CreateStudent(contextWithToken(ctx), req)

	return resp, err
}

func createStudentReq(locationIDs []string) *pb.CreateStudentRequest {
	randomID := newID()
	req := &pb.CreateStudentRequest{
		SchoolId: constants.ManabieSchool,
		StudentProfile: &pb.CreateStudentRequest_StudentProfile{
			Email:             fmt.Sprintf("%v@example.com", randomID),
			Password:          fmt.Sprintf("password-%v", randomID),
			Name:              fmt.Sprintf("user-%v", randomID),
			CountryCode:       cpb.Country_COUNTRY_VN,
			EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			PhoneNumber:       fmt.Sprintf("phone-number-%v", randomID),
			StudentExternalId: fmt.Sprintf("student-external-id-%v", randomID),
			StudentNote:       fmt.Sprintf("some random student note %v", randomID),
			Grade:             5,
			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:            pb.Gender_MALE,
			LocationIds:       locationIDs,
		},
	}

	return req
}

func (s *suite) createStudentsByGRPC(ctx context.Context, numberStudents int) ([]*pb.StudentProfileV2, error) {
	profiles := make([]*pb.StudentProfileV2, 0)
	for i := 0; i < numberStudents; i++ {
		if _, err := s.createStudentByGRPC(ctx, "general info", "all fields"); err != nil {
			return nil, errors.Wrap(err, "s.createStudentByGRPC")
		}

		resp := s.Response.(*pb.UpsertStudentResponse)
		if len(resp.StudentProfiles) == 0 {
			return nil, errors.New("resp.StudentProfiles is empty")
		}

		profiles = append(profiles, resp.StudentProfiles...)
	}

	if len(profiles) != numberStudents {
		return nil, errors.Errorf("len(profiles) != numberStudents")
	}
	return profiles, nil
}

func getNumberInvalidRows(str string) (int, error) {
	re := regexp.MustCompile(`have (\d+) invalid row`)
	match := re.FindStringSubmatch(str)
	if len(match) == 0 {
		return 0, nil
	}

	num, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "strconv.ParseInt")
	}
	return int(num), nil
}
