package usermgmt

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	enigma_entities "github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	location_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"google.golang.org/protobuf/proto"
)

func (s *suite) toStudentSyncMsg(ctx context.Context, status, actionKind string, total int) ([]*npb.EventUserRegistration_Student, error) {
	if total == 0 {
		return []*npb.EventUserRegistration_Student{}, nil
	}

	stepState := StepStateFromContext(ctx)
	students := []*npb.EventUserRegistration_Student{}

	switch status {
	case "new":
		for i := 0; i < total; i++ {
			randomID := newID()
			students = append(students, &npb.EventUserRegistration_Student{
				ActionKind:  npb.ActionKind(npb.ActionKind_value[actionKind]),
				StudentId:   idutil.ULIDNow(),
				StudentDivs: []int64{1, 2},
				LastName:    fmt.Sprintf("last_name_%s", randomID),
				GivenName:   fmt.Sprintf("first_name_%s", randomID), // given_name will be first_name (LT-33615)
			})
		}
	case "existed":
		_, err := s.jprepSyncStudentsWithActionAndStudentsWithAction(ctx, strconv.Itoa(total), npb.ActionKind_ACTION_KIND_UPSERTED.String(), "0", "")
		if err != nil {
			return nil, fmt.Errorf("toStudentSyncMsg.jprepSyncStudentsWithActionAndStudentsWithAction: %v", err)
		}

		_, err = s.theseStudentsMustBeStoreInOurSystem(ctx)
		if err != nil {
			return nil, fmt.Errorf("toStudentSyncMsg.theseStudentsMustBeStoreInOurSystem: %v", err)
		}

		for _, student := range stepState.Request.([]*npb.EventUserRegistration_Student) {
			student.ActionKind = npb.ActionKind(npb.ActionKind_value[actionKind])
			students = append(students, student)
		}
	}

	return students, nil
}

func (s *suite) jprepSyncStudentsWithActionAndStudentsWithAction(ctx context.Context, numberOfNewStudent, newStudentAction, numberOfExistedStudent, existedStudentAction string) (context.Context, error) {
	total, err := strconv.Atoi(numberOfNewStudent)
	if err != nil {
		return ctx, err
	}

	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	ctx = s.signedIn(ctx, constants.JPREPSchool, StaffRoleSchoolAdmin)

	newStudents, err := s.toStudentSyncMsg(ctx, "new", newStudentAction, total)
	if err != nil {
		return ctx, err
	}

	students := []*npb.EventUserRegistration_Student{}
	students = append(students, newStudents...)
	total, err = strconv.Atoi(numberOfExistedStudent)
	if err != nil {
		return ctx, err
	}

	existedStudents, err := s.toStudentSyncMsg(ctx, "existed", existedStudentAction, total)
	if err != nil {
		return ctx, err
	}

	students = append(students, existedStudents...)
	stepState.Request = students
	signature := idutil.ULIDNow()
	partnerSyncDataLog, err := createPartnerSyncDataLog(ctx, s.BobDBTrace, signature, 0)
	if err != nil {
		return ctx, fmt.Errorf("create partner sync data log error: %w", err)
	}
	stepState.PartnerSyncDataLogID = partnerSyncDataLog.PartnerSyncDataLogID.String

	partnerSyncDataLogSplit, err := createLogSyncDataSplit(ctx, s.BobDBTrace, string(enigma_entities.KindStudent))
	if err != nil {
		return ctx, fmt.Errorf("create partner sync data log split error: %w", err)
	}

	req := &npb.EventUserRegistration{
		RawPayload: []byte("{}"),
		Signature:  signature,
		Students:   students,
		LogId:      partnerSyncDataLogSplit.PartnerSyncDataLogSplitID.String,
	}

	data, _ := proto.Marshal(req)
	_, err = s.JSM.PublishContext(ctx, constants.SubjectUserRegistrationNatsJS, data)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("publish: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

//nolint:gocyclo
func (s *suite) theseStudentsMustBeStoreInOurSystem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, student := range stepState.Request.([]*npb.EventUserRegistration_Student) {
		if student.ActionKind == npb.ActionKind_ACTION_KIND_DELETED {
			continue
		}

		token, err := s.tryGenerateExchangeToken(student.StudentId, pb.USER_GROUP_STUDENT.String())
		if err != nil {
			return ctx, err
		}

		ctx = common.ValidContext(ctx, constants.JPREPSchool, student.StudentId, token)
		resp, err := pb.NewStudentClient(s.BobConn).GetStudentProfile(ctx, &pb.GetStudentProfileRequest{})
		if err != nil {
			return ctx, fmt.Errorf("theseStudentsMustBeStoreInOurSystem.GetStudentProfile: %s", err.Error())
		}

		if len(resp.Datas) != 1 {
			return ctx, fmt.Errorf("not found student")
		}

		studentResp := resp.Datas[0]
		switch {
		case studentResp.Profile.Id != student.StudentId:
			return ctx, fmt.Errorf("studentID does not match")
		case studentResp.Profile.Name != helper.CombineFirstNameAndLastNameToFullName(student.GivenName, student.LastName):
			return ctx, fmt.Errorf("name does not match")
		case studentResp.Profile.Country != pb.COUNTRY_JP:
			return ctx, fmt.Errorf("country does not match")
		case len(studentResp.Profile.Divs) != len(student.StudentDivs):
			return ctx, fmt.Errorf("divs does not match")
		case studentResp.Profile.School.Id != constants.JPREPSchool:
			return ctx, fmt.Errorf("school does not match")
		}

		for i := range student.StudentDivs {
			if student.StudentDivs[i] != studentResp.Profile.Divs[i] {
				return ctx, fmt.Errorf("student divs does not match")
			}
		}

		if err := s.checkUserAccessPath(ctx, studentResp.Profile.Id); err != nil {
			return ctx, fmt.Errorf("checkUserAccessPath: %w", err)
		}

		if err := s.checkEnrollmentStatusHistory(ctx, studentResp.Profile.Id); err != nil {
			return ctx, fmt.Errorf("checkEnrollmentStatusHistory: %w", err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) storeLogDataSplitWithCorrectStatus(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var statusFromDB string

	query := `SELECT status FROM partner_sync_data_log_split
		WHERE partner_sync_data_log_id = $1 AND partner_sync_data_log_split_id = $2 LIMIT 1`

	retryTime := 3

	err := try.Do(func(attempt int) (bool, error) {
		if err := s.BobDBTrace.QueryRow(ctx, query, stepState.PartnerSyncDataLogID, stepState.PartnerSyncDataLogSplitID).Scan(&statusFromDB); err != nil {
			return true, fmt.Errorf("query partner sync data log id err: %w", err)
		}

		if statusFromDB == status {
			return false, nil
		}

		if attempt < retryTime {
			time.Sleep(time.Second)
			return true, nil
		}

		return false, fmt.Errorf("unexpected status data log split expect %s but status in database is %s", status, statusFromDB)
	})

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) checkUserAccessPath(ctx context.Context, userID string) error {
	ctx = s.signedIn(ctx, constants.JPREPSchool, StaffRoleSchoolAdmin)

	stmt := `SELECT uap.user_id, uap.location_id
						FROM user_access_paths uap
						WHERE uap.user_id = $1`

	rows, err := s.BobDBTrace.Query(ctx, stmt, userID)
	if err != nil {
		return fmt.Errorf("query user_access_path stored fail %s", err.Error())
	}
	defer rows.Close()
	userAccessPaths := []*entity.UserAccessPath{}

	for rows.Next() {
		uap := &entity.UserAccessPath{}
		if err := rows.Scan(
			&uap.UserID,
			&uap.LocationID,
		); err != nil {
			return err
		}

		userAccessPaths = append(userAccessPaths, uap)
	}

	// jprep sync api only add org location for student/teacher
	if len(userAccessPaths) != 1 {
		return fmt.Errorf("unexpected user_access_path stored, expect 1 but got %d", len(userAccessPaths))
	}

	orgLocation, err := (&location_repo.LocationRepo{}).GetLocationOrg(ctx, s.BobDBTrace, golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return fmt.Errorf("get orgLocation failed %w", err)
	}

	if orgLocation.LocationID != userAccessPaths[0].LocationID.String {
		return fmt.Errorf("location in user_access_path is unexpected, expect %s but got %s", orgLocation.LocationID, userAccessPaths[0].LocationID.String)
	}

	return nil
}

func (s *suite) checkEnrollmentStatusHistory(ctx context.Context, userID string) error {
	ctx = s.signedIn(ctx, constants.JPREPSchool, StaffRoleSchoolAdmin)

	stmt := `
		SELECT sesh.student_id, sesh.enrollment_status
		FROM student_enrollment_status_history sesh
			JOIN user_access_paths uap ON sesh.student_id = uap.user_id AND sesh.location_id = uap.location_id
		WHERE sesh.student_id = $1
			AND sesh.deleted_at IS NULL
			AND uap.deleted_at IS NULL`

	rows, err := s.BobDBTrace.Query(ctx, stmt, userID)
	if err != nil {
		return fmt.Errorf("query student_enrollment_status_history stored fail %s", err.Error())
	}
	defer rows.Close()
	enrollmentStatusHistories := []*entity.StudentEnrollmentStatusHistory{}

	for rows.Next() {
		esh := &entity.StudentEnrollmentStatusHistory{}
		if err := rows.Scan(
			&esh.StudentID,
			&esh.EnrollmentStatus,
		); err != nil {
			return err
		}

		enrollmentStatusHistories = append(enrollmentStatusHistories, esh)
	}

	// jprep sync api only add org location for student/teacher
	if len(enrollmentStatusHistories) != 1 {
		return fmt.Errorf("unexpected student_enrollment_status_history stored, expect 1 but got %d", len(enrollmentStatusHistories))
	}

	return nil
}
