package entryexitmgmt

import (
	"context"
	"fmt"
	"time"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

func (s *suite) recordsFoundWithDefaultLimitAreDisplayedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*eepb.RetrieveEntryExitRecordsRequest)
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = eepb.NewEntryExitServiceClient(s.EntryExitMgmtConn).RetrieveEntryExitRecords(contextWithToken(ctx), req)
	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) parentIsAtTheEntryExitScreen(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// create request
	req := &eepb.RetrieveEntryExitRecordsRequest{
		Paging: &cpb.Paging{
			// default limit is 100
			// for testing just set it to 3 to create small data to test
			Limit: uint32(3),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 0,
			},
		},
		// default filter date all
		RecordFilter: eepb.RecordFilter_ALL,
	}

	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) loginsLearnerApp(ctx context.Context, role string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, role)
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) parentSelectsThisExistingStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request.(*eepb.RetrieveEntryExitRecordsRequest).StudentId = stepState.StudentID
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentHasEntryAndExitRecord(ctx context.Context, existOrNot, dateRecords string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if existOrNot != "existing" {
		return StepStateToContext(ctx, stepState), nil
	}

	entryexitRepo := s.StudentEntryExitRecordsRepo
	var entryAt time.Time
	var exitAt time.Time

	switch dateRecords {
	case "Last month":
		entryAt = time.Now().AddDate(0, -1, 0)
		exitAt = time.Now().Add(8*time.Hour).AddDate(0, -1, 0)
	case "Last year", "Last year this month":
		entryAt = time.Now().AddDate(-1, 0, 0)
		exitAt = time.Now().Add(8*time.Hour).AddDate(-1, 0, 0)
	default:
		// this year and this month
		entryAt = time.Now()
		exitAt = time.Now().Add(8 * time.Hour)
	}

	// create records
	for i := 0; i < 7; i++ {
		entryexitEntity, err := generateEntryExitRecord(stepState.StudentID, entryAt, exitAt)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if err := entryexitRepo.Create(ctx, s.EntryExitMgmtDBTrace, entryexitEntity); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) parentChecksTheFilterForRecords(ctx context.Context, dateRecordsFilter string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch dateRecordsFilter {
	case "This month":
		stepState.Request.(*eepb.RetrieveEntryExitRecordsRequest).RecordFilter = eepb.RecordFilter_THIS_MONTH
	case "This year":
		stepState.Request.(*eepb.RetrieveEntryExitRecordsRequest).RecordFilter = eepb.RecordFilter_THIS_YEAR
	case "Last month":
		stepState.Request.(*eepb.RetrieveEntryExitRecordsRequest).RecordFilter = eepb.RecordFilter_LAST_MONTH
	default:
		stepState.Request.(*eepb.RetrieveEntryExitRecordsRequest).RecordFilter = eepb.RecordFilter_ALL
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) parentScrollsDownToDisplayAllRecords(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// the initial response is added on the response count record
	countRecord := int32(len(stepState.Response.(*eepb.RetrieveEntryExitRecordsResponse).EntryExitRecords))
	// get the previous request
	req := stepState.Request.(*eepb.RetrieveEntryExitRecordsRequest)
	// assign the initial response next page for pagination
	stepStateResponse := stepState.Response
	for {
		req.Paging = stepStateResponse.(*eepb.RetrieveEntryExitRecordsResponse).NextPage
		stepState.RequestSentAt = time.Now()
		stepState.Response, stepState.ResponseErr = eepb.NewEntryExitServiceClient(s.EntryExitMgmtConn).RetrieveEntryExitRecords(contextWithToken(ctx), req)
		stepStateResponse = stepState.Response
		if countRecord == countRecord+int32(len(stepState.Response.(*eepb.RetrieveEntryExitRecordsResponse).EntryExitRecords)) {
			break
		}
		countRecord += int32(len(stepState.Response.(*eepb.RetrieveEntryExitRecordsResponse).EntryExitRecords))
	}
	stepState.RetrieveRecordCount = countRecord
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) allRecordsFoundAreDisplayedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// get all records set to limit 10 since less test record are created
	req := stepState.Request.(*eepb.RetrieveEntryExitRecordsRequest)
	req.Paging.Limit = 10
	req.Paging.Offset = &cpb.Paging_OffsetInteger{
		OffsetInteger: 0,
	}
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = eepb.NewEntryExitServiceClient(s.EntryExitMgmtConn).RetrieveEntryExitRecords(contextWithToken(ctx), req)
	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	if int32(len(stepState.Response.(*eepb.RetrieveEntryExitRecordsResponse).EntryExitRecords)) != stepState.RetrieveRecordCount {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve all records: unexpected row affected")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) noRecordsFoundDisplayedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*eepb.RetrieveEntryExitRecordsRequest)
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = eepb.NewEntryExitServiceClient(s.EntryExitMgmtConn).RetrieveEntryExitRecords(contextWithToken(ctx), req)
	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	if stepState.Response != nil && int32(len(stepState.Response.(*eepb.RetrieveEntryExitRecordsResponse).EntryExitRecords)) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("there are records for this student")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) parentHasAnotherExistingStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.thisParentHasAnExistingStudent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisParentHasAnExistingStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	parentID := stepState.CurrentParentID
	ctx, err := s.thereIsAnExistingStudent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	err = s.createStudentParentRelationship(
		ctx,
		stepState.StudentID,
		[]string{parentID},
		upb.FamilyRelationship_name[int32(upb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER)],
	)
	if err != nil {
		return ctx, err
	}

	time.Sleep(3 * time.Second) // added for kafka sync delay

	return StepStateToContext(ctx, stepState), nil
}
