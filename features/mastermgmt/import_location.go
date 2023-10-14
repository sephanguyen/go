package mastermgmt

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func (s *suite) theInvalidLocationLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*mpb.ImportLocationRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*mpb.ImportLocationResponse)
	for _, row := range stepState.InvalidCsvRows {
		found := false
		for _, e := range resp.Errors {
			if strings.TrimSpace(reqSplit[e.RowNumber-1]) == row {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid line is not returned in response")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidLocationLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	locations, err := s.selectNewsLocations(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, row := range stepState.ValidCsvRows {
		rowSplit := strings.Split(row, ",")
		partnerInternalID := rowSplit[0]
		name := rowSplit[1]
		partnerInternalParentID := rowSplit[3]
		found := false
		for _, e := range locations {
			if e.AccessPath.String != "" && e.Name.String == name && e.PartnerInternalID.String == partnerInternalID && e.PartnerInternalParentID.String == partnerInternalParentID {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row %v", row)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingLocation(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	ctx, err := s.upsertedLocationSubscribe(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.upsertedLocationSubscribe: %w", err)
	}
	stepState.Response, stepState.ResponseErr = mpb.NewLocationManagementGRPCServiceClient(s.MasterMgmtConn).
		ImportLocation(contextWithToken(s, ctx), stepState.Request.(*mpb.ImportLocationRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertedLocationSubscribe(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(), nats.AckWait(30 * time.Second),
			nats.MaxDeliver(10),
			nats.Bind(constants.StreamSyncLocationUpserted, constants.DurableSyncLocationUpsertedImporter),
			nats.DeliverSubject(constants.DeliverSyncLocationUpsertedImporter),
		},
	}

	handlerUpsertedLocationSubscription := func(ctx context.Context, data []byte) (bool, error) {
		r := &npb.EventSyncLocation{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return false, err
		}
		stepState.FoundChanForJetStream <- "UpsertLocation"
		return false, nil
	}
	sub, err := s.JSM.QueueSubscribe(constants.SubjectSyncLocationUpserted, constants.QueueSyncLocationUpsertedImporter, opts, handlerUpsertedLocationSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importLocationOtherSchool(ctx context.Context, locationID, parentID, name string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.insertOrgLocationTypes(ctx, "org")
	s.insertOrgLocation(ctx)

	request := &mpb.ImportLocationRequest{
		Payload: []byte(fmt.Sprintf(`partner_internal_id,name,location_type,partner_internal_parent_id,is_archived
		%s`, fmt.Sprintf("%s,%s,center,%s,0", locationID, name, parentID))),
	}
	stepState.Response, stepState.ResponseErr = mpb.NewLocationManagementGRPCServiceClient(s.MasterMgmtConn).
		ImportLocation(contextWithToken(s, ctx), request)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("NewLocationManagementGRPCServiceClient.ImportLocation err %s", stepState.ResponseErr)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnNumberLocationFailed(ctx context.Context, numFailed int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*mpb.ImportLocationResponse)
	if resp != nil && resp.TotalFailed != int32(numFailed) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect %d failed, but got %d", numFailed, resp.TotalFailed)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aLocationValidRequestPayloadWith(ctx context.Context, rowCondition string, locationType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertOrgLocation(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	types := strings.Split(locationType, ",")
	if len(types) != 3 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("location type must 3 value")
	}
	numId := rand.Intn(1000)
	validRow1 := fmt.Sprintf("location1_%d,Location 1 %s,%s,", numId, idutil.ULIDNow(), types[0])
	validRow2 := fmt.Sprintf("location2_%d,Location 2 %s,%s,location1_%d", numId, idutil.ULIDNow(), types[1], numId)
	validRow3 := fmt.Sprintf("location3_%d,Location 3 %s,%s,location1_%d", numId, idutil.ULIDNow(), types[1], numId)

	invalidEmptyRow1 := fmt.Sprintf(",Location %s,,,0", idutil.ULIDNow())
	invalidEmptyRow3 := fmt.Sprintf("%s,,center,,0", idutil.ULIDNow())
	invalidValueRow1 := fmt.Sprintf(",Location 1 %s,place,,0", idutil.ULIDNow())
	invalidValueRow2 := fmt.Sprintf(",Location 2 %s,place,,0", idutil.ULIDNow())
	invalidValueRow3 := fmt.Sprintf("location_4,Location 4 %s,place,location_4,0", idutil.ULIDNow())
	invalidValueRow4 := fmt.Sprintf("location_5,Location 5 %s,place,location_6,0", idutil.ULIDNow())
	invalidValueRow5 := fmt.Sprintf("location_7,Location 5 %s,place,location_8,0", idutil.ULIDNow())
	invalidValueRow6 := fmt.Sprintf("location_8,Location 5 %s,brand,location_2,0", idutil.ULIDNow())
	invalidValueRow7 := fmt.Sprintf("location_9,Location 1 %s,org,,0", idutil.ULIDNow())
	invalidValueRow8 := fmt.Sprintf("location_3,Location 3 %s,brand,location_1,0", idutil.ULIDNow())
	invalidValueRow9 := fmt.Sprintf("location1_%d,Location 1 %s,%s,,0", numId, idutil.ULIDNow(), types[0])
	invalidValueRow10 := fmt.Sprintf("location1_%d,Location 1 %s,%s,,0", numId, string([]byte{0xff, 0xfe, 0xfd}), types[0])

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "all valid rows":
		stepState.Request = &mpb.ImportLocationRequest{
			Payload: []byte(fmt.Sprintf(`partner_internal_id,name,location_type,partner_internal_parent_id
			%s
			%s
			%s`, validRow1, validRow2, validRow3)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3}
	case "empty value row":
		stepState.Request = &mpb.ImportLocationRequest{
			Payload: []byte(fmt.Sprintf(`partner_internal_id,name,location_type,partner_internal_parent_id
				%s
				%s`, invalidEmptyRow1, invalidEmptyRow3)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow3}
	case "invalid value row":
		stepState.Request = &mpb.ImportLocationRequest{
			Payload: []byte(fmt.Sprintf(`partner_internal_id,name,location_type,partner_internal_parent_id
				%s
				%s
				%s
				%s`, invalidValueRow1, invalidValueRow2, invalidValueRow3, invalidValueRow4)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2, invalidValueRow3, invalidValueRow4}
	case "valid and invalid rows":
		stepState.Request = &mpb.ImportLocationRequest{
			Payload: []byte(fmt.Sprintf(`partner_internal_id,name,location_type,partner_internal_parent_id
				%s
				%s
				%s
				%s
				%s
				%s
				%s
				%s
				%s
				%s
				%s
				%s`, validRow1, invalidEmptyRow1, invalidValueRow1, invalidValueRow2, invalidValueRow3, invalidValueRow4, invalidValueRow5, invalidValueRow6, invalidValueRow7, invalidValueRow8, invalidValueRow9, invalidValueRow10)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidValueRow1, invalidValueRow2, invalidValueRow3, invalidValueRow4, invalidValueRow5, invalidValueRow6, invalidValueRow7, invalidValueRow8, invalidValueRow9, invalidValueRow10}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aLocationInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &mpb.ImportLocationRequest{}
	case "header only":
		stepState.Request = &mpb.ImportLocationRequest{
			Payload: []byte(`partner_internal_id,name,location_type,partner_internal_parent_id,is_archived`),
		}
	case "number of column is not equal 5":
		stepState.Request = &mpb.ImportLocationRequest{
			Payload: []byte(`partner_internal_id,name,location_type,is_archived
			01FPVK2MVW961BS1BFZHBR7J44,Location 1,center,,0
			01FPVK2MVW961BS1BFZHBR7J55,Location 2,center,,0`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &mpb.ImportLocationRequest{
			Payload: []byte(`partner_internal_id,name,location_type,partner_internal_parent_id,is_archived
				1,Location 1
				2,Location 2
				3,Location 3`),
		}
	case "wrong partner_internal_id column name in header":
		stepState.Request = &mpb.ImportLocationRequest{
			Payload: []byte(`locationID,name,location_type,partner_internal_parent_id,is_archived
				1,Location 1,center,
				2,Location 2,center,
				3,Location 3,center,`),
		}
	case "wrong name column name in header":
		stepState.Request = &mpb.ImportLocationRequest{
			Payload: []byte(`partner_internal_id,nameming,location_type,partner_internal_parent_id,is_archived
			1,Location 1,center,
			2,Location 2,center,
			3,Location 3,center,`),
		}
	case "wrong location_type column name in header":
		stepState.Request = &mpb.ImportLocationRequest{
			Payload: []byte(`partner_internal_id,nameming,locationType,partner_internal_parent_id,is_archived
			1,Location 1,center,
			2,Location 2,center,
			3,Location 3,center,`),
		}
	case "wrong partner_internal_parent_id column name in header":
		stepState.Request = &mpb.ImportLocationRequest{
			Payload: []byte(`partner_internal_id,nameming,location_type,parentID,is_archived
			1,Location 1,center,
			2,Location 2,center,
			3,Location 3,center,`),
		}
	case "wrong is_archived column name in header":
		stepState.Request = &mpb.ImportLocationRequest{
			Payload: []byte(`partner_internal_id,nameming,locationType,partner_internal_parent_id,ISarchived
			1,Location 1,center,
			2,Location 2,center,
			3,Location 3,center,`),
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectNewsLocations(ctx context.Context) ([]*entities.Location, error) {
	var allEntities []*entities.Location
	stmt :=
		`
		SELECT
			partner_internal_id,
			name,
			location_type,
			partner_internal_parent_id,
			access_path
		FROM
			locations
		WHERE deleted_at is null 
		order by updated_at desc limit 50

		`
	rows, err := s.BobDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query location")
	}
	defer rows.Close()
	for rows.Next() {
		e := &entities.Location{}
		err := rows.Scan(
			&e.PartnerInternalID,
			&e.Name,
			&e.LocationType,
			&e.PartnerInternalParentID,
			&e.AccessPath,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan location")
		}
		allEntities = append(allEntities, e)
	}
	return allEntities, nil
}

func (s *suite) insertOrgLocation(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)
	var locationID string
	query := fmt.Sprintf("SELECT location_id FROM locations WHERE parent_location_id is null and resource_path = '%s' limit 1", fmt.Sprint(constants.ManabieSchool))

	if err := database.Select(ctx, s.BobDBTrace, query).ScanFields(&locationID); err != nil {
		if err != pgx.ErrNoRows {
			return err
		}
	}
	if locationID == "" {
		randomStr := idutil.ULIDNow()
		name := database.Text("org")
		stmt := fmt.Sprintf(`INSERT INTO locations (location_id, name, created_at, updated_at, resource_path, location_type, is_archived )
		VALUES ($1, $2, now(), now(), '%s', $3, false)`, fmt.Sprint(constants.ManabieSchool))
		_, err := s.BobDBTrace.Exec(ctx, stmt, randomStr, name, stepState.LocationTypeOrgID)
		if err != nil {
			return fmt.Errorf("cannot insert location org, err: %s", err)
		}
	}

	return nil
}

func (s *suite) mustStoreLocationImportLogs(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), nil
	}
	type payload struct {
		DeletedIds  []string `json:"deleted_ids"`
		UpsertedIds []string `json:"upserted_ids"`
	}
	query := `SELECT payload->'deleted_ids' as deleted_ids ,payload->'upserted_ids' as upserted_ids
	FROM mastermgmt_import_log where import_type = $1 order by created_at desc limit 50`
	rows, err := s.BobDBTrace.Query(ctx, query, repo.ImportTypeLocation)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()
	payloads := make([]payload, 0)
	for rows.Next() {
		pl := payload{}
		err := rows.Scan(&pl.DeletedIds, &pl.UpsertedIds)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("rows.Scan: %w", err)
		}
		payloads = append(payloads, pl)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("rows.Err: %w", err)
	}
	validRows := stepState.ValidCsvRows
	expectedPayload := payload{}
	for _, row := range validRows {
		columns := strings.Split(row, ",")
		expectedPayload.UpsertedIds = append(expectedPayload.UpsertedIds, columns[0])
	}
	found := false
	for _, pl := range payloads {
		if stringutil.SliceElementsMatch(pl.UpsertedIds, expectedPayload.UpsertedIds) {
			found = true
			break
		}
	}
	if !found {
		return StepStateToContext(ctx, stepState), fmt.Errorf("payload log not correct")
	}
	return StepStateToContext(ctx, stepState), nil
}
