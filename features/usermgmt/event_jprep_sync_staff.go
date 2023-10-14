package usermgmt

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/common"
	enigma_entites "github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"google.golang.org/protobuf/proto"
)

func (s *suite) toStaffSyncMsg(ctx context.Context, status, actionKind string, total int) ([]*npb.EventUserRegistration_Staff, error) {
	stepState := StepStateFromContext(ctx)

	if total == 0 {
		return []*npb.EventUserRegistration_Staff{}, nil
	}

	staffs := []*npb.EventUserRegistration_Staff{}

	switch status {
	case "new":
		for i := 0; i < total; i++ {
			staffs = append(staffs, &npb.EventUserRegistration_Staff{
				ActionKind: npb.ActionKind(npb.ActionKind_value[actionKind]),
				StaffId:    idutil.ULIDNow(),
				Name:       idutil.ULIDNow(),
			})
		}
	case "existed":
		stepState.Request = nil
		ctx, err := s.jprepSyncStaffsWithActionAndStaffsWithAction(ctx, strconv.Itoa(total), npb.ActionKind_ACTION_KIND_UPSERTED.String(), "0", "")
		if err != nil {
			return nil, err
		}
		if _, err = s.theseStaffsMustBeStoreInOurSystem(ctx); err != nil {
			return nil, err
		}

		for _, eventStaff := range stepState.Request.([]*npb.EventUserRegistration_Staff) {
			eventStaff.ActionKind = npb.ActionKind(npb.ActionKind_value[actionKind])
			staffs = append(staffs, eventStaff)
		}
	}

	return staffs, nil
}

func (s *suite) syncStaffRequest(ctx context.Context, staffs []*npb.EventUserRegistration_Staff) error {
	stepState := StepStateFromContext(ctx)
	signature := idutil.ULIDNow()
	partnerSyncDataLog, err := createPartnerSyncDataLog(ctx, s.BobDBTrace, signature, 0)
	if err != nil {
		return fmt.Errorf("create partner sync data log error: %w", err)
	}
	stepState.PartnerSyncDataLogID = partnerSyncDataLog.PartnerSyncDataLogID.String

	partnerSyncDataLogSplit, err := createLogSyncDataSplit(ctx, s.BobDBTrace, string(enigma_entites.KindStaff))
	if err != nil {
		return fmt.Errorf("create partner sync data log split error: %w", err)
	}

	req := &npb.EventUserRegistration{
		RawPayload: []byte("{}"),
		Signature:  signature,
		Staffs:     staffs,
		LogId:      partnerSyncDataLogSplit.PartnerSyncDataLogSplitID.String,
	}
	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("error when marshal request: %w", err)
	}
	if _, err = s.JSM.PublishContext(ctx, constants.SubjectUserRegistrationNatsJS, data); err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	return nil
}

func (s *suite) jprepSyncStaffsWithActionAndStaffsWithAction(ctx context.Context, numberOfNewStaff, newStaffAction, numberOfExistedStaff, existedStaffAction string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.JPREPSchool, StaffRoleSchoolAdmin)

	total, err := strconv.Atoi(numberOfNewStaff)
	if err != nil {
		return ctx, err
	}
	stepState.RequestSentAt = time.Now()

	newStaffs, err := s.toStaffSyncMsg(ctx, "new", newStaffAction, total)
	if err != nil {
		return ctx, fmt.Errorf("s.toStaffSyncMsg new: %v", err)
	}
	staffs := []*npb.EventUserRegistration_Staff{}
	staffs = append(staffs, newStaffs...)

	total, err = strconv.Atoi(numberOfExistedStaff)
	if err != nil {
		return ctx, err
	}

	existedStaffs, err := s.toStaffSyncMsg(ctx, "existed", existedStaffAction, total)
	if err != nil {
		return ctx, fmt.Errorf("s.toStaffSyncMsg existed: %v", err)
	}

	staffs = append(staffs, existedStaffs...)
	stepState.Request = staffs

	if err := s.syncStaffRequest(ctx, staffs); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error when sync staff request: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theseStaffsMustBeStoreInOurSystem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, staff := range stepState.Request.([]*npb.EventUserRegistration_Staff) {
		// ignore deleted staff
		if staff.ActionKind == npb.ActionKind_ACTION_KIND_DELETED {
			continue
		}

		token, err := s.tryGenerateExchangeToken(staff.StaffId, pb.USER_GROUP_TEACHER.String())
		if err != nil {
			return ctx, err
		}
		ctx := common.ValidContext(ctx, constants.JPREPSchool, staff.StaffId, token)

		resp, err := pb.NewUserServiceClient(s.BobConn).GetTeacherProfiles(ctx, &pb.GetTeacherProfilesRequest{})
		if err != nil {
			return ctx, err
		}

		if len(resp.Profiles) != 1 {
			return ctx, fmt.Errorf("not found staff")
		}

		staffResp := resp.Profiles[0]
		if staffResp.Id != staff.StaffId {
			return ctx, fmt.Errorf("staffID does not match")
		}
		if staffResp.Name != staff.Name {
			return ctx, fmt.Errorf("name does not match")
		}
		if staffResp.Country != pb.COUNTRY_JP {
			return ctx, fmt.Errorf("country does not match")
		}

		if err := s.checkUserAccessPath(ctx, staff.StaffId); err != nil {
			return ctx, fmt.Errorf("checkUserAccessPath: %w", err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) jprepSyncSyncDeletedStaffWithAction(ctx context.Context, syncAction string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// init staff sync request
	existedStaffs, err := s.toStaffSyncMsg(ctx, "existed", syncAction, 1)
	if err != nil {
		return ctx, err
	}
	// assign for checkAfterSignedInGetSelfProfile
	stepState.Request = existedStaffs

	// sync delete staff first
	for _, staffReq := range existedStaffs {
		staffReq.ActionKind = npb.ActionKind_ACTION_KIND_DELETED
	}
	if err := s.syncStaffRequest(ctx, existedStaffs); err != nil {
		return ctx, fmt.Errorf("error when sync staff request with ActionKind_ACTION_KIND_DELETED: %w", err)
	}

	// re-sync staff with specify sync action
	for _, staffReq := range existedStaffs {
		staffReq.ActionKind = npb.ActionKind(npb.ActionKind_value[syncAction])
	}
	if err := s.syncStaffRequest(ctx, existedStaffs); err != nil {
		return ctx, fmt.Errorf("error when sync staff request with %s: %w", syncAction, err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theyLoginOurSystemAndGetSelfProfileInfo(ctx context.Context, statusAction string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, staffReq := range stepState.Request.([]*npb.EventUserRegistration_Staff) {
		// sign with current staff
		token, err := s.tryGenerateExchangeToken(staffReq.StaffId, pb.USER_GROUP_TEACHER.String())
		if err != nil {
			switch statusAction {
			case "can":
				return StepStateToContext(ctx, stepState), fmt.Errorf("error when signing with the current user: %w", err)

				// user is soft_deleted cannot sign in our system
			case "cannot":
				return StepStateToContext(ctx, stepState), nil
			}
		}

		ctx = common.ValidContext(ctx, constants.JPREPSchool, staffReq.StaffId, token)
		// get self-profile with signed staff
		_, err = pb.NewUserServiceClient(s.BobConn).GetTeacherProfiles(ctx, &pb.GetTeacherProfilesRequest{})
		if err != nil && statusAction == "can" {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect staff can login but can't get profile: %s", err.Error())
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) tryGenerateExchangeToken(staffID, userGroup string) (string, error) {
	var (
		token string
		err   error
	)
	retryTime := 5

	err = try.Do(func(attempt int) (bool, error) {
		token, err = s.generateExchangeToken(staffID, userGroup)
		if err == nil {
			return false, nil
		}

		if attempt < retryTime {
			time.Sleep(time.Second)
			return true, err
		}

		return false, err
	})

	return token, err
}
