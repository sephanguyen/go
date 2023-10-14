package mastermgmt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"
)

func (s *suite) theInvalidLocationTypeLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*mpb.ImportLocationTypeRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*mpb.ImportLocationTypeResponse)
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

func (s *suite) theValidLocationTypeLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	validRows := stepState.ValidCsvRows

	names := sliceutils.Map(validRows, func(s string) string {
		split := strings.Split(s, ",")
		return strings.ToLower(split[0])
	})

	locationTypes, err := s.selectLocTypeByNames(ctx, names)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	dbLocTypes := sliceutils.Map(locationTypes, func(lt *repo.LocationType) string {
		return fmt.Sprintf("%s,%s,%s", lt.Name.String, lt.DisplayName.String, lt.ParentName.String)
	})

	slices.Sort(validRows)
	slices.Sort(dbLocTypes)
	if slices.Compare(validRows, dbLocTypes) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row.\nexpect:%s\ngot:%s", strings.Join(validRows, ";"), strings.Join(dbLocTypes, ";"))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingLocationType(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	ctx, err := s.upsertedLocationTypeSubscribe(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.upsertedLocationTypeSubscribe: %w", err)
	}
	stepState.Response, stepState.ResponseErr = mpb.NewLocationManagementGRPCServiceClient(s.MasterMgmtConn).
		ImportLocationType(contextWithToken(s, ctx), stepState.Request.(*mpb.ImportLocationTypeRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertedLocationTypeSubscribe(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(), nats.AckWait(30 * time.Second),
			nats.MaxDeliver(10),
			nats.Bind(constants.StreamSyncLocationTypeUpserted, constants.DurableSyncLocationTypeUpsertedImporter),
			nats.DeliverSubject(constants.DeliverSyncLocationTypeUpsertedImporter)},
	}

	handlerUpsertedLocationTypeSubscription := func(ctx context.Context, data []byte) (bool, error) {
		r := &npb.EventSyncLocationType{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return false, err
		}
		stepState.FoundChanForJetStream <- "UpsertLocationType"
		return false, nil
	}
	sub, err := s.JSM.QueueSubscribe(constants.SubjectSyncLocationTypeUpserted, constants.QueueSyncLocationTypeUpsertedImporter, opts, handlerUpsertedLocationTypeSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) setResourcePathToContext(ctx context.Context, schoolID string) context.Context {
	stepState := StepStateFromContext(ctx)
	claim := interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: schoolID,
			DefaultRole:  entities.UserGroupAdmin,
			UserGroup:    entities.UserGroupAdmin,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, &claim)
	return StepStateToContext(ctx, stepState)
}

func (s *suite) importLocationTypeOtherSchool(ctx context.Context, name, parent_name, display_name string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	request := &mpb.ImportLocationTypeRequest{
		Payload: []byte(fmt.Sprintf(`name,display_name,parent_name
		%s`, fmt.Sprintf("%s,%s,%s", name, display_name, parent_name))),
	}
	stepState.Response, stepState.ResponseErr = mpb.NewLocationManagementGRPCServiceClient(s.MasterMgmtConn).
		ImportLocationType(contextWithToken(s, ctx), request)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("NewLocationManagementGRPCServiceClient.ImportLocationType err %s", stepState.ResponseErr)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnNumberLocationTypeFailed(ctx context.Context, numFailed int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*mpb.ImportLocationTypeResponse)
	if resp != nil && resp.TotalFailed != int32(numFailed) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect %d failed, but got %d", numFailed, resp.TotalFailed)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aLocationTypeValidRequestPayloadWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	numID := 10
	validRow1 := fmt.Sprintf("brand_%d,Chi nhánh quận 1,org", numID)
	validRow2 := fmt.Sprintf("center_%d,Chi nhánh quận 2,brand_%d", numID, numID)
	validRow3 := fmt.Sprintf("place_%d,Chi nhánh quận 3,center_%d", numID, numID)
	invalidEmptyRow1 := ",Chi nhánh 2,,1"
	invalidEmptyRow2 := ",aa,,1"
	invalidValueRow1 := "org2,Name,not_exist"
	invalidValueRow2 := "org,Chi nhánh quận 1,"
	invalidValueRow3 := "org,Chi nhánh quận 3,"
	invalidValueRow4 := fmt.Sprintf("place,%s,center", string([]byte{0xff, 0xfe, 0xfd}))
	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "all valid rows":
		stepState.Request = &mpb.ImportLocationTypeRequest{
			Payload: []byte(fmt.Sprintf(`name,display_name,parent_name
			%s
			%s
			%s`, validRow1, validRow2, validRow3)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3}
	case "empty value row":
		stepState.Request = &mpb.ImportLocationTypeRequest{
			Payload: []byte(fmt.Sprintf(`name,display_name,parent_name
					%s
					%s`, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "valid and invalid rows":
		stepState.Request = &mpb.ImportLocationTypeRequest{
			Payload: []byte(fmt.Sprintf(`name,display_name,parent_name
					%s
					%s
					%s
					%s
					%s`, invalidValueRow1, invalidValueRow2, invalidValueRow3, invalidValueRow4, validRow1)),
		}
		stepState.ValidCsvRows = []string{validRow1}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2, invalidValueRow3, invalidValueRow4}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aLocationTypeInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &mpb.ImportLocationTypeRequest{}
	case "header only":
		stepState.Request = &mpb.ImportLocationTypeRequest{
			Payload: []byte(`name,display_name,parent_name`),
		}
	case "number of column is not equal 4":
		stepState.Request = &mpb.ImportLocationTypeRequest{
			Payload: []byte(`name,display_name,parent_name
			Name1,display,org
			Name2,display,org`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &mpb.ImportLocationTypeRequest{
			Payload: []byte(`name,display_name,parent_name
					name4,Display
					Center location type 1`),
		}
	case "wrong name column name in header":
		stepState.Request = &mpb.ImportLocationTypeRequest{
			Payload: []byte(`Name1,display_name,parent_name
					Name1,Display1`),
		}
	case "wrong display_name column name in header":
		stepState.Request = &mpb.ImportLocationTypeRequest{
			Payload: []byte(`name,displayName,parent_name
					Name1,Display1`),
		}
	case "wrong parent_name column name in header":
		stepState.Request = &mpb.ImportLocationTypeRequest{
			Payload: []byte(`name,display_name,ParentName
					Name1,Display1`),
		}
	case "wrong is_archived column name in header":
		stepState.Request = &mpb.ImportLocationTypeRequest{
			Payload: []byte(`name,display_name,parent_name,ISDELETED
					Name1,Display1,org,1`),
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectLocTypeByNames(ctx context.Context, names []string) ([]*repo.LocationType, error) {
	var allEntities []*repo.LocationType
	stmt :=
		`
		SELECT
			name,
			display_name,
			parent_name
		FROM
			location_types
		where deleted_at is null and name = ANY ($1)
		order by updated_at desc limit 20
		`
	rows, err := s.BobDBTrace.Query(
		ctx,
		stmt,
		names,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query location type")
	}
	defer rows.Close()
	for rows.Next() {
		e := &repo.LocationType{}
		err := rows.Scan(
			&e.Name,
			&e.DisplayName,
			&e.ParentName,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan location type")
		}
		allEntities = append(allEntities, e)
	}
	return allEntities, nil
}

func (s *suite) aLocationTypeValues(ctx context.Context, locationType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	orgName := "org"
	err := s.insertOrgLocationTypes(ctx, orgName)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	types := strings.Split(locationType, ",")
	if len(types) != 3 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("location type must 3 value")
	}
	validRow1 := fmt.Sprintf("%s,Chi nhánh quận 1,%s,0", types[0], orgName)
	validRow2 := fmt.Sprintf("%s,Chi nhánh quận 2,%s,0", types[1], types[0])
	validRow3 := fmt.Sprintf("%s,Chi nhánh quận 3,%s,0", types[2], types[1])
	stepState.Request = &mpb.ImportLocationTypeRequest{
		Payload: []byte(fmt.Sprintf(`name,display_name,parent_name
		%s
		%s
		%s`, validRow1, validRow2, validRow3)),
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertOrgLocationTypes(ctx context.Context, orgName string) error {
	stepState := StepStateFromContext(ctx)
	randomStr := idutil.ULIDNow()
	name := database.Text(orgName)
	stmt := fmt.Sprintf(`INSERT INTO location_types (location_type_id, name, display_name, resource_path, updated_at, created_at, is_archived)
	VALUES ($1, $2, 'Org', '%s', now(), now(), false) ON CONFLICT ON CONSTRAINT unique__location_type_name_resource_path DO NOTHING`, fmt.Sprint(constants.ManabieSchool))
	_, err := s.BobDBTrace.Exec(ctx, stmt, randomStr, name)
	if err != nil {
		return fmt.Errorf("cannot insert location type, err: %s", err)
	}
	var locationTypeID string
	queryLocationTypeOrg := fmt.Sprintf("SELECT location_type_id FROM location_types l WHERE l.name = $1 AND l.deleted_at IS NULL and l.resource_path = '%s'", fmt.Sprint(constants.ManabieSchool))
	if err := s.BobDBTrace.QueryRow(ctx, queryLocationTypeOrg, orgName).Scan(&locationTypeID); err != nil {
		return fmt.Errorf("cannot get location type org, err: %s", err)
	}
	stepState.LocationTypeOrgID = locationTypeID
	return nil
}

func (s *suite) someLocationsWithLocationTypes(ctx context.Context, typeName, parentName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertOrgLocationTypes(ctx, "org")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	locationTypeID := idutil.ULIDNow()
	locationID := idutil.ULIDNow()
	typeName = fmt.Sprintf("type_%s_%s", typeName, stepState.Random)

	queryType := `INSERT INTO location_types (location_type_id, name, display_name, parent_name, updated_at, created_at, is_archived)
	VALUES ($1, $2, $3, $4, now(), now(), false) ON CONFLICT ON CONSTRAINT unique__location_type_name_resource_path DO NOTHING`

	_, err = s.BobDBTrace.Exec(ctx, queryType, locationTypeID, typeName, typeName, parentName)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert location type, err: %s", err)
	}

	query := `INSERT INTO locations (location_id,name,location_type,is_archived,created_at,updated_at) VALUES($1,'locations',$2, false, now(), now()) 
	ON CONFLICT ON CONSTRAINT locations_pkey DO NOTHING`
	_, err = s.BobDBTrace.Exec(ctx, query, locationID, locationTypeID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert location, err: %s", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) locationTypeWithParentExistInDB(ctx context.Context, typeName, parentName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	typeName = fmt.Sprintf("type_%s_%s", typeName, stepState.Random)
	row := s.BobDB.QueryRow(ctx, `SELECT COUNT(*) FROM location_types l
		WHERE l.name = $1 AND l.parent_name = $2 AND l.deleted_at IS NULL`, &typeName, &parentName)
	var total int
	if err := row.Scan(&total); err != nil {
		return ctx, err
	}
	if total == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Expect location type %s and parent %s exist, but got empty", typeName, parentName)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminUpdateParentOfLocationType(ctx context.Context, typeName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	typeName = fmt.Sprintf("type_%s_%s", typeName, stepState.Random)
	request := &mpb.ImportLocationTypeRequest{
		Payload: []byte(fmt.Sprintf(`name,display_name,parent_name,is_archived
		%s`, fmt.Sprintf("%s,display_name,org,0", typeName))),
	}
	stepState.Response, stepState.ResponseErr = mpb.NewLocationManagementGRPCServiceClient(s.MasterMgmtConn).
		ImportLocationType(contextWithToken(s, ctx), request)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("NewLocationManagementGRPCServiceClient.ImportLocationType err %s", stepState.ResponseErr)
	}
	resp := stepState.Response.(*mpb.ImportLocationTypeResponse)
	expectErr := fmt.Sprintf(`unable to import location type item: locations with type %s exist`, typeName)
	if resp.Errors[0].GetError() != expectErr {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect %s, but got %s", expectErr, resp.Errors[0].GetError())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) mustStoreLocationTypeImportLogs(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), nil
	}
	type payload struct {
		DeletedNames  []string `json:"deleted_names"`
		UpsertedNames []string `json:"upserted_names"`
	}
	query := `SELECT payload->'deleted_names' as deleted_names ,payload->'upserted_names' as upserted_names
	 FROM mastermgmt_import_log where import_type = $1 order by created_at desc limit 10`
	rows, err := s.BobDBTrace.Query(ctx, query, repo.ImportTypeLocationType)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()
	payloads := make([]payload, 0)
	for rows.Next() {
		pl := payload{}
		err := rows.Scan(&pl.DeletedNames, &pl.UpsertedNames)
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
		expectedPayload.UpsertedNames = append(expectedPayload.UpsertedNames, strings.TrimSpace(columns[0]))
	}
	invalidRows := stepState.InvalidCsvRows
	for _, row := range invalidRows {
		columns := strings.Split(row, ",")
		expectedPayload.DeletedNames = append(expectedPayload.DeletedNames, strings.TrimSpace(columns[0]))
	}
	found := false
	for _, pl := range payloads {
		if stringutil.SliceElementsMatch(pl.DeletedNames, expectedPayload.DeletedNames) &&
			stringutil.SliceElementsMatch(pl.UpsertedNames, expectedPayload.UpsertedNames) {
			found = true
			break
		}
	}
	if !found {
		expectedJSON, _ := json.Marshal(expectedPayload)
		gotJSON, _ := json.Marshal(payloads)
		return StepStateToContext(ctx, stepState), fmt.Errorf("payload log not correct.\nexpected:%s\ngot:%s", string(expectedJSON), string(gotJSON))
	}
	return StepStateToContext(ctx, stepState), nil
}
