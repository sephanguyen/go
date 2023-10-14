package entryexitmanagement

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/entryexitmgmt/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) createStudentParentRelationship(ctx context.Context, studentID string, parentIDs []string, relationship string) error {
	entities := make([]*bob_entities.StudentParent, 0, len(parentIDs))
	for _, parentID := range parentIDs {
		studentParent := &bob_entities.StudentParent{}
		database.AllNullEntity(studentParent)
		err := multierr.Combine(
			studentParent.StudentID.Set(studentID),
			studentParent.ParentID.Set(parentID),
			studentParent.Relationship.Set(relationship),
		)
		if err != nil {
			return err
		}
		entities = append(entities, studentParent)
	}
	if err := (&bob_repo.StudentParentRepo{}).Upsert(ctx, s.bobDBTrace, entities); err != nil {
		return err
	}

	return nil
}

func (s *suite) createEntryExitRecord(ctx context.Context, studentID string) error {
	s.stepState.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithTokenForGrpcCall(s, ctx)
	now := time.Now()
	req := &eepb.CreateEntryExitRequest{
		EntryExitPayload: &eepb.EntryExitPayload{
			StudentId:     studentID,
			EntryDateTime: timestamppb.New(now.Add(-7 * time.Hour)),
			ExitDateTime:  timestamppb.New(now),
		},
	}
	s.Response, s.ResponseErr = eepb.NewEntryExitServiceClient(s.entryExitMgmtConn).CreateEntryExit(ctx, req)
	if s.ResponseErr != nil {
		return errors.Wrap(s.ResponseErr, "CreateEntryExit()")
	}
	s.Request = req
	return nil
}

func (s *suite) loginsLearnerAppWithAResourcePathFrom(ctx context.Context, role string, organization string) error {
	roleStr := (strings.Split(role, " "))[0]
	organizationStrSplit := strings.Split(organization, " ")
	resourcePathOrdinalStr := organizationStrSplit[len(organizationStrSplit)-1]
	s.stepState.ResourcePath = resourcePathOrdinalStr
	resourcePathOrdinal, err := strconv.Atoi(resourcePathOrdinalStr)

	if err != nil {
		return err
	}
	s.CurrentSchoolID = int32(resourcePathOrdinal)
	return s.signedInAsAccountWithResourcePath(ctx, roleStr, resourcePathOrdinalStr)
}

func (s *suite) thisParentHasExistingStudentWithEntryAndExitRecord(ctx context.Context) error {
	// create student
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	studentID := idutil.ULIDNow()

	err := s.aValidStudentInDBWithResourcePath(ctx, studentID, fmt.Sprintf("%d", s.CurrentSchoolID))
	if err != nil {
		return err
	}

	s.stepState.UserGroupInContext = constant.UserGroupParent
	parentID := s.UserGroupCredentials[s.UserGroupInContext].UserID
	// create relationship to parent
	ctx = getContextJWTClaims(ctx, fmt.Sprintf("%d", s.CurrentSchoolID))
	err = s.createStudentParentRelationship(
		ctx,
		studentID,
		[]string{parentID},
		upb.FamilyRelationship_name[int32(upb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER)],
	)
	if err != nil {
		return err
	}
	// existing entry exit records created with admin permission
	err = s.signedInAsAccountWithResourcePath(ctx, "school admin", s.stepState.ResourcePath)
	if err != nil {
		return err
	}
	err = s.createEntryExitRecord(ctx, studentID)
	if err != nil {
		return err
	}
	return nil
}

func (s *suite) visitsItsStudentsEntryAndExitRecordOnLearnerApp(ctx context.Context, role string) error {
	role = (strings.Split(role, " "))[0]

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	switch role {
	case "parent":
		s.stepState.UserGroupInContext = constant.UserGroupParent
	case "student":
		s.stepState.UserGroupInContext = constant.UserGroupStudent
	}

	ctx = contextWithTokenForGrpcCall(s, ctx)
	// get the last request
	latestStudentIDReq := s.Request.(*eepb.CreateEntryExitRequest).EntryExitPayload.StudentId

	if role == "parent" {
		studentParentRepo := &repositories.StudentParentRepo{}
		ctx = getContextJWTClaims(ctx, fmt.Sprintf("%d", s.CurrentSchoolID))
		getParentIDs, err := studentParentRepo.GetParentIDsByStudentID(ctx, s.bobDBTrace, latestStudentIDReq)
		if err != nil {
			return err
		}

		if s.stepState.CurrentParentID != getParentIDs[0] {
			return errors.New("student parent not match")
		}
	}

	req := &eepb.RetrieveEntryExitRecordsRequest{
		StudentId: latestStudentIDReq,
	}
	s.Response, s.ResponseErr = eepb.NewEntryExitServiceClient(s.entryExitMgmtConn).RetrieveEntryExitRecords(ctx, req)

	return nil
}

func (s *suite) onlySeesRecordsFrom(ctx context.Context, role string, organization string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if s.ResponseErr != nil {
		return errors.Wrap(s.ResponseErr, "RetrieveEntryExitRecords()")
	}

	if len(s.Response.(*eepb.RetrieveEntryExitRecordsResponse).EntryExitRecords) == 0 {
		return errors.New("expected to see student entry exit records but got none")
	}

	return nil
}
