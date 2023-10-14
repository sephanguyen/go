package entryexitmanagement

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/yasuo/constant"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"
)

func (s *suite) loginsWithResourcePathFrom(ctx context.Context, role, organization string) (context.Context, error) {
	organizationStrSplit := strings.Split(organization, " ")
	resourcePathOrdinalStr := organizationStrSplit[len(organizationStrSplit)-1]
	s.stepState.ResourcePath = resourcePathOrdinalStr
	resourcePathOrdinal, err := strconv.Atoi(resourcePathOrdinalStr)

	if err != nil {
		return ctx, err
	}
	s.CurrentSchoolID = int32(resourcePathOrdinal)
	return ctx, s.signedInAsAccountWithResourcePath(ctx, role, resourcePathOrdinalStr)
}

func (s *suite) thisSchoolAdminCreatesANewStudentEntryAndExitRecord(ctx context.Context) error {
	nCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	s.stepState.UserGroupInContext = constant.UserGroupSchoolAdmin
	err := s.createStudentWithResourcePath(ctx)
	if err != nil {
		return err
	}

	err = s.createEntryExitRecord(nCtx, s.Response.(*upb.CreateStudentResponse).StudentProfile.Student.UserProfile.UserId)
	if err != nil {
		return err
	}
	return nil
}

func (s *suite) thisSchoolAdminSeeTheNewStudentEntryAndExitRecord(ctx context.Context, result string) (context.Context, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithToken(s, ctx)

	studentID := s.Request.(*eepb.CreateEntryExitRequest).EntryExitPayload.StudentId
	switch result {
	case "can":
		err := canSeeEntryExitRecordsOnBackOffice(ctx, studentID)
		if err != nil {
			return ctx, errors.Wrap(err, "error canSeeEntryExitRecordsOnBackOffice")
		}
	case "cannot":
		err := cannotSeeEntryExitRecordsOnBackOffice(ctx, studentID)
		if err != nil {
			return ctx, errors.Wrap(err, "error cannotSeeEntryExitRecordsOnBackOffice")
		}
	}

	return ctx, nil
}

type EntryExitRecord struct {
	ID        int    `graphql:"entryexit_id"`
	StudentID string `graphql:"student_id"`
}

func queryEntryExitRecords(ctx context.Context, studentID string) ([]*EntryExitRecord, error) {
	// Pre-setup for hasura query using admin secret
	if err := trackTableForHasuraQuery("student_entryexit_records"); err != nil {
		return nil, errors.Wrap(err, "trackTableForHasuraQuery()")
	}
	if err := createSelectPermissionForHasuraQuery("student_entryexit_records"); err != nil {
		return nil, errors.Wrap(err, "createSelectPermissionForHasuraQuery()")
	}
	query := `query ($student_id: String!) {
			student_entryexit_records(where: {student_id: {_eq: $student_id}, deleted_at: {_is_null:true}}) {
				entryexit_id
				student_id
			}
		}`
	if err := addQueryToAllowListForHasuraQuery(query); err != nil {
		return nil, errors.Wrap(err, "addQueryToAllowListForHasuraQuery()")
	}

	// Query newly created teacher from hasura
	var profileQuery struct {
		EntryExitRecords []*EntryExitRecord `graphql:"student_entryexit_records(where: {student_id: {_eq: $student_id}, deleted_at: {_is_null:true}})"`
	}

	variables := map[string]interface{}{
		"student_id": graphql.String(studentID),
	}
	err := queryHasura(ctx, &profileQuery, variables, bobHasuraAdminUrl+"/v1/graphql")
	if err != nil {
		return nil, errors.Wrap(err, "queryHasura")
	}
	return profileQuery.EntryExitRecords, nil
}

func canSeeEntryExitRecordsOnBackOffice(ctx context.Context, studentID string) error {
	entryExitRecords, err := queryEntryExitRecords(ctx, studentID)
	if err != nil {
		return errors.Wrap(err, "queryEntryExitRecords")
	}

	if len(entryExitRecords) < 1 {
		return fmt.Errorf("can't find student entry exit with student id: %s", studentID)
	}

	queriedEntryExitRecord := entryExitRecords[0]
	if queriedEntryExitRecord.StudentID != studentID {
		return fmt.Errorf(`expected student id for this record is: "%s" but actual returned student id is "%s"`, studentID, queriedEntryExitRecord.StudentID)
	}

	return nil
}

func cannotSeeEntryExitRecordsOnBackOffice(ctx context.Context, studentID string) error {
	entryExitRecords, err := queryEntryExitRecords(ctx, studentID)
	if err != nil {
		return errors.Wrap(err, "queryEntryExitRecords")
	}

	if len(entryExitRecords) > 0 {
		return fmt.Errorf("expected no records but got %d records", len(entryExitRecords))
	}

	return nil
}
