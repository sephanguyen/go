package yasuo

import (
	"context"
	"fmt"
	"strconv"
	"time"

	enigma_entites "github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	ypb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"
)

func (s *suite) toStaffSyncMsg(ctx context.Context, status, actionKind string, total int) (context.Context, []*npb.EventUserRegistration_Staff, error) {
	stepState := StepStateFromContext(ctx)

	if total == 0 {
		return ctx, []*npb.EventUserRegistration_Staff{}, nil
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
		ctx, err1 := s.jprepSyncStaffsWithActionAndStaffsWithAction(ctx, strconv.Itoa(total), npb.ActionKind_ACTION_KIND_UPSERTED.String(), "0", "")
		ctx, err2 := s.theseStaffsMustBeStoreInOurSystem(ctx)
		err := multierr.Combine(err1, err2)

		if err != nil {
			return ctx, nil, err
		}

		for _, s := range stepState.Request.([]*npb.EventUserRegistration_Staff) {
			s.ActionKind = npb.ActionKind(npb.ActionKind_value[actionKind])
			staffs = append(staffs, s)
		}
	}

	return StepStateToContext(ctx, stepState), staffs, nil
}

func (s *suite) userGetBasicProfile(ctx context.Context, profile *ypb.GetBasicProfileRequest) (*ypb.GetBasicProfileResponse, error) {
	return ypb.NewUserServiceClient(s.Conn).GetBasicProfile(ctx, profile)
}

func (s *suite) syncStaffRequest(ctx context.Context, staffs []*npb.EventUserRegistration_Staff) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	signature := idutil.ULIDNow()
	ctx, err := s.createPartnerSyncDataLog(ctx, signature, 0)
	if err != nil {
		return ctx, fmt.Errorf("create partner sync data log error: %w", err)
	}

	if ctx, err = s.createLogSyncDataSplit(ctx, string(enigma_entites.KindStaff)); err != nil {
		return ctx, fmt.Errorf("create partner sync data log split error: %w", err)
	}

	req := &npb.EventUserRegistration{
		RawPayload: []byte("{}"),
		Signature:  signature,
		Staffs:     staffs,
		LogId:      stepState.PartnerSyncDataLogSplitId,
	}
	data, err := proto.Marshal(req)
	if err != nil {
		return ctx, fmt.Errorf("error when marshal request: %w", err)
	}
	if _, err = s.JSM.PublishContext(ctx, constants.SubjectUserRegistrationNatsJS, data); err != nil {
		return ctx, fmt.Errorf("publish: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) jprepSyncStaffsWithActionAndStaffsWithAction(ctx context.Context, numberOfNewStaff, newStaffAction, numberOfExistedStaff, existedStaffAction string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = auth.InjectFakeJwtToken(ctx, fmt.Sprint(constants.JPREPSchool))

	total, err := strconv.Atoi(numberOfNewStaff)
	if err != nil {
		return ctx, err
	}
	stepState.RequestSentAt = time.Now()

	ctx, newStaffs, err := s.toStaffSyncMsg(ctx, "new", newStaffAction, total)
	if err != nil {
		return ctx, fmt.Errorf("s.toStaffSyncMsg new: %v", err)
	}

	total, err = strconv.Atoi(numberOfExistedStaff)
	if err != nil {
		return ctx, err
	}

	ctx, existedStaffs, err := s.toStaffSyncMsg(ctx, "existed", existedStaffAction, total)
	if err != nil {
		return ctx, fmt.Errorf("s.toStaffSyncMsg existed: %v", err)
	}

	staffs := append(newStaffs, existedStaffs...)
	stepState.Request = staffs

	if ctx, err = s.syncStaffRequest(ctx, staffs); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error when sync staff request: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theseStaffsMustBeStoreInOurSystem(ctx context.Context) (context.Context, error) {
	time.Sleep(5 * time.Second)
	stepState := StepStateFromContext(ctx)

	var err error
	for _, staff := range stepState.Request.([]*npb.EventUserRegistration_Staff) {
		// ignore deleted staff
		if staff.ActionKind == npb.ActionKind_ACTION_KIND_DELETED {
			continue
		}
		stepState.AuthToken, err = s.generateExchangeToken(staff.StaffId, constant.UserGroupTeacher)
		if err != nil {
			return ctx, err
		}
		ctx = StepStateToContext(ctx, stepState)

		resp, err := pb.NewUserServiceClient(s.BobConn).GetTeacherProfiles(contextWithToken(s, ctx), &pb.GetTeacherProfilesRequest{})
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
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) jprepSyncSyncDeletedStaffWithAction(ctx context.Context, syncAction string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = auth.InjectFakeJwtToken(ctx, fmt.Sprint(constants.JPREPSchool))

	// init staff sync request
	ctx, existedStaffs, err := s.toStaffSyncMsg(ctx, "existed", syncAction, 1)
	if err != nil {
		return ctx, err
	}
	// assign for checkAfterSignedInGetSelfProfile
	stepState.Request = existedStaffs

	// sync delete staff first
	for _, staffReq := range existedStaffs {
		staffReq.ActionKind = npb.ActionKind_ACTION_KIND_DELETED
	}
	if ctx, err = s.syncStaffRequest(ctx, existedStaffs); err != nil {
		return ctx, fmt.Errorf("error when sync staff request with ActionKind_ACTION_KIND_DELETED: %w", err)
	}

	// re-sync staff with specify sync action
	for _, staffReq := range existedStaffs {
		staffReq.ActionKind = npb.ActionKind(npb.ActionKind_value[syncAction])
	}
	if ctx, err = s.syncStaffRequest(ctx, existedStaffs); err != nil {
		return ctx, fmt.Errorf("error when sync staff request with %s: %w", syncAction, err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkAfterSignedInGetSelfProfile(ctx context.Context, statusAction string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentUserGroup = constant.UserGroupTeacher
	mapResponse := map[bool]string{
		true:  "can",
		false: "cannot",
	}

	for _, staffReq := range stepState.Request.([]*npb.EventUserRegistration_Staff) {
		// sign with current staff
		stepState.CurrentUserID = staffReq.StaffId
		ctx, err := s.aSignedInCurrentUser(ctx)
		if err != nil {
			switch statusAction {
			case "can":
				return StepStateToContext(ctx, stepState), fmt.Errorf("error when signing with the current user: %w", err)

				// user is soft_deleted cannot sign in our system
			case "cannot":
				return StepStateToContext(ctx, stepState), nil
			}
		}

		// get self-profile with signed staff
		res, err := s.userGetBasicProfile(s.signedCtx(ctx), &ypb.GetBasicProfileRequest{})

		// check current user can get self profile
		statusGetProfileInfo := res != nil && err == nil
		if mapResponse[statusGetProfileInfo] != statusAction {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect `%s` failed, but got `%s`", statusAction, mapResponse[statusGetProfileInfo])
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aTeacherAccountWithSchoolID(ctx context.Context, schoolID int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := idutil.ULIDNow()
	stepState.CurrentTeacherID = id
	stepState.CurrentSchoolID = schoolID
	return s.aValidUserInDB(StepStateToContext(ctx, stepState), withID(id), withRole(pb.USER_GROUP_TEACHER.String()))
}

func (s *suite) ATeacherAccount(ctx context.Context) (context.Context, error) {
	return s.aTeacherAccountWithSchoolID(ctx, 1)
}
