package usermgmt

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
)

func (s *suite) importingSchoolInfo(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewSchoolInfoServiceClient(s.UserMgmtConn).
		ImportSchoolInfo(contextWithToken(ctx), stepState.Request.(*pb.ImportSchoolInfoRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertSomeSchools(ctx context.Context) error {
	for i := 0; i < 3; i++ {
		id := newID()
		name := database.Text(fmt.Sprintf("School %v", id))
		phonetic := database.Text(fmt.Sprintf("S%v", id))
		isArchived := database.Bool(rand.Int()%2 == 0)
		stmt := `INSERT INTO school_info
		(school_id, school_name, school_name_phonetic, school_level_id, address, is_archived, created_at, updated_at, resource_path)
		VALUES ($1, $2, $3, $4, $5, $6, now(), now(), $7)`
		_, err := s.BobPostgresDBTrace.Exec(auth.InjectFakeJwtToken(ctx, fmt.Sprint(constant.ManabieSchool)), stmt, id, name, phonetic, "Ho Chi Minh", "Go Vap", isArchived, fmt.Sprint(constant.ManabieSchool))
		if err != nil {
			return fmt.Errorf("cannot insert school, err: %s", err)
		}
	}
	return nil
}

func (s *suite) selectAllSchoolInfo(ctx context.Context) ([]*entity.SchoolInfo, error) {
	var allEntities []*entity.SchoolInfo
	stmt :=
		`
		SELECT 
			school_id,
			school_name,
			school_name_phonetic,
			school_level_id,
			address,
			is_archived,
			resource_path
		FROM
			school_info
		`
	rows, err := s.BobPostgresDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query school_info")
	}
	defer rows.Close()
	for rows.Next() {
		e := &entity.SchoolInfo{}
		err := rows.Scan(
			&e.ID,
			&e.Name,
			&e.NamePhonetic,
			&e.LevelID,
			&e.Address,
			&e.IsArchived,
			&e.ResourcePath,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan school_info")
		}
		allEntities = append(allEntities, e)
	}
	return allEntities, nil
}

func (s *suite) anSchoolInfoValidRequestPayloadWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomeSchools(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	existingSchoolInfo, err := s.selectAllSchoolInfo(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	validRow1 := fmt.Sprintf(",School %[1]s,S%[1]s,school_level_id-%[1]s,Address %[1]s,1", newID())
	validRow2 := fmt.Sprintf(",School %[1]s,S%[1]s,school_level_id-%[1]s,Address %[1]s,1", newID())
	validRow3 := fmt.Sprintf("%[1]s,School %[2]s,S%[2]s,school_level_id-%[2]s,Address %[2]s,0", existingSchoolInfo[0].ID.String, newID())
	invalidEmptyRow1 := fmt.Sprintf(",School %[1]s,S%[1]s,school_level_id-%[1]s,Address %[1]s,", newID())
	invalidEmptyRow2 := fmt.Sprintf("%[1]s,School %s,S%[2]s,school_level_id-%[2]s,Address %[2]s,", existingSchoolInfo[1].ID.String, newID())
	invalidValueRow1 := fmt.Sprintf(",School %[1]s,S%[1]s,school_level_id-%[1]s,Address %[1]s,Archived", newID())
	invalidValueRow2 := fmt.Sprintf("%[1]s,School %[2]s,S%[2]s,school_level_id-%[2]s,Address %[2]s,Archived", existingSchoolInfo[2].ID.String, newID())

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "all valid rows":
		stepState.Request = &pb.ImportSchoolInfoRequest{
			Payload: []byte(fmt.Sprintf(`school_id,school_name,school_name_phonetic,school_level_id,address,is_archived
			%s
			%s`, validRow1, validRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
	case "empty value row":
		stepState.Request = &pb.ImportSchoolInfoRequest{
			Payload: []byte(fmt.Sprintf(`school_id,school_name,school_name_phonetic,school_level_id,address,is_archived
			%s
			%s`, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":
		stepState.Request = &pb.ImportSchoolInfoRequest{
			Payload: []byte(fmt.Sprintf(`school_id,school_name,school_name_phonetic,school_level_id,address,is_archived
			%s
			%s`, invalidValueRow1, invalidValueRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportSchoolInfoRequest{
			Payload: []byte(fmt.Sprintf(`school_id,school_name,school_name_phonetic,school_level_id,address,is_archived
			%s
			%s
			%s
			%s
			%s
			%s
			%s`, validRow1, validRow2, validRow3, invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2}
	case "500 rows":
		payload := `school_id,school_name,school_name_phonetic,school_level_id,address,is_archived`
		for i := 0; i < 500; i++ {
			row := fmt.Sprintf("\n,School %[1]s,S%[1]s,school_level_id-%[1]s,Address %[1]s,1", fmt.Sprintf("%s%d", newID(), i))
			payload += row
			stepState.ValidCsvRows = append(stepState.ValidCsvRows, row)
		}
		stepState.Request = &pb.ImportSchoolInfoRequest{
			Payload: []byte(payload),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anSchoolInfoInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case NoData:
		stepState.Request = &pb.ImportSchoolInfoRequest{}
	case "header only":
		stepState.Request = &pb.ImportSchoolInfoRequest{
			Payload: []byte(`school_id,school_name,school_name_phonetic,school_level_id,address,is_archived`),
		}
	case "number of column is not equal 6":
		stepState.Request = &pb.ImportSchoolInfoRequest{
			Payload: []byte(`school_id,school_name
				1,School 1
				2,School 2
				3,School 3`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportSchoolInfoRequest{
			Payload: []byte(`school_id,school_name,
			1,School 1
			2,School 2
			3,School 3`),
		}
	case "wrong school_id column name in header":
		stepState.Request = &pb.ImportSchoolInfoRequest{
			Payload: []byte(`Number,school_name,school_name_phonetic,school_level_id,address,is_archived
			,School 1,S1,school_level_id-1,Address 1,0`),
		}
	case "wrong school_name column name in header":
		stepState.Request = &pb.ImportSchoolInfoRequest{
			Payload: []byte(`school_id,School Name,school_name_phonetic,school_level_id,address,is_archived
			,School 1,S1,school_level_id-1,Address 1,0`),
		}
	case "wrong school_name_phonetic column name in header":
		stepState.Request = &pb.ImportSchoolInfoRequest{
			Payload: []byte(`school_id,school_name,School Name Phonetic,school_level_id,address,is_archived
			,School 1,S1,school_level_id-1,Address 1,0`),
		}
	case "wrong school_level_id column name in header":
		stepState.Request = &pb.ImportSchoolInfoRequest{
			Payload: []byte(`school_id,school_name,school_name_phonetic,School level id,address,is_archived
			,School 1,S1,school_level_id-1,Address 1,0`),
		}
	case "wrong address column name in header":
		stepState.Request = &pb.ImportSchoolInfoRequest{
			Payload: []byte(`school_id,school_name,school_name_phonetic,school_level_id,Addresses,is_archived
			,School 1,S1,school_level_id-1,Address 1,0`),
		}
	case "wrong is_archived column name in header":
		stepState.Request = &pb.ImportSchoolInfoRequest{
			Payload: []byte(`school_id,school_name,school_name_phonetic,school_level_id,address,IsArchived
			,School 1,S1,school_level_id-1,Address 1,0`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidSchoolInfoLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	const (
		id = iota
		name
		namePhonetic
		levelID
		address
		isArchived
	)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allSchoolInfo, err := s.selectAllSchoolInfo(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// we should use map for allSchoolInfo, but it leads to some more code and not many items in
	// stepState.ValidCsvRows and allSchoolInfo, so we can do like below to make it simple
	for _, row := range stepState.ValidCsvRows {
		rowSplit := strings.Split(row, ",")
		isArchived, err := strconv.ParseBool(rowSplit[isArchived])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		found := false
		for _, e := range allSchoolInfo {
			if e.Name.String == rowSplit[name] &&
				e.NamePhonetic.String == rowSplit[namePhonetic] &&
				e.LevelID.String == rowSplit[levelID] &&
				e.Address.String == rowSplit[address] &&
				e.IsArchived.Bool == isArchived &&
				e.ResourcePath.String == fmt.Sprint(constants.ManabieSchool) {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theInvalidSchoolInfoLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.ImportSchoolInfoRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pb.ImportSchoolInfoResponse)

	// we cannot pass array struct to array interface{}, so we need to convert it first.
	respErrors := make([]ImportCSVErrors, len(resp.Errors))
	for i := range resp.Errors {
		respErrors[i] = resp.Errors[i]
	}
	err := checkInvalidRows(stepState.InvalidCsvRows, reqSplit, respErrors)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("checkInvalidRows: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}
