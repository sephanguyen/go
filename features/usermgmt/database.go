package usermgmt

import (
	"context"
	"math"
	"strconv"

	"github.com/pkg/errors"
)

func (s *suite) aDatabaseWithShardId(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	/*ctx = contextWithTokenV2(ctx, s.OrgAndSignedInSchoolAdminToken[strconv.Itoa(constants.ManabieSchool)])

	//Random school admin
	randomSchoolAdmin, err := CreateARandomSchoolAdmin(ctx, s.BobPostgresDB, s.TenantManager)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "CreateARandomSchoolAdmin")
	}

	idToken, err := common.GenerateAuthenticationToken(s.FirebaseAddress, randomSchoolAdmin.UserID().String(), entity.UserGroupSchoolAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "generateExchangeToken")
	}

	//Login in to get id token
	//idToken, err := LoginInAuthPlatform(ctx, apiKey, "", randomSchoolAdmin.Email().String(), randomSchoolAdmin.Password().String())
	//if err != nil {
	//	return "", err
	//}

	exchangedToken, err := ExchangeToken(ctx, s.ShamirConn, randomSchoolAdmin.Email().String(), randomSchoolAdmin.Password().String(), idToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	fmt.Println(exchangedToken)*/

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) clientAcquiresAConnection(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := `SELECT current_setting('database.shard_id')`

	var queriedShardID string

	err := s.BobDBTrace.QueryRow(ctx, stmt).Scan(&queriedShardID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "QueryRow.Scan")
	}

	shardId, err := strconv.ParseInt(queriedShardID, 10, 64)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "strconv.ParseInt")
	}
	stepState.ShardID = shardId

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theConnectionShouldHaveCorrespondingShardIdInSessionVariable(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if int64(*s.Cfg.PostgresV2.Databases["bob"].ShardID) != stepState.ShardID {
		return StepStateToContext(ctx, stepState), errors.New("shard id is not correct")
	}

	return StepStateToContext(ctx, stepState), nil
}

func decodeSharedId(shardedId int64) (int64, int64, int64) {
	decodedTime := (shardedId >> 21) & 0x7FFFFFFFFFF
	decodedShardId := (shardedId >> 10) & 0x7FF
	decodedIncrementNum := (shardedId >> 0) & 0x3FF

	return decodedTime, decodedShardId, decodedIncrementNum
}

func (s *suite) clientGenerateShardedIdViaDatabaseFunc(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	maxMillisValueCanStore := int64(math.Pow(float64(2), float64(43)) - 1)
	maxShardIdValue := int64(math.Pow(float64(2), float64(11)) - 1)
	maxSequenceNumValue := int64(math.Pow(float64(2), float64(10)) - 1)

	millisToEncode := maxMillisValueCanStore
	shardIdToEncode := maxShardIdValue
	sequenceNumToEncode := maxSequenceNumValue

	stmt := `SELECT generate_sharded_id($1, $2, $3) `

	var shardedId int64

	err := s.BobDBTrace.QueryRow(ctx, stmt, &millisToEncode, &shardIdToEncode, &sequenceNumToEncode).Scan(&shardedId)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "QueryRow.Scan")
	}

	decodedMillis, decodedShardId, decodedIncrementNum := decodeSharedId(shardedId)

	/*fmt.Println("decodedMillis:", decodedMillis)
	fmt.Println("decodedShardId:", decodedShardId)
	fmt.Println("decodedIncrementNum:", decodedIncrementNum)*/

	switch {
	case decodedMillis != millisToEncode:
		return StepStateToContext(ctx, stepState), errors.New("decoded millis value is different with millis value to encode ")
	case decodedShardId != shardIdToEncode:
		return StepStateToContext(ctx, stepState), errors.New("decoded shard id value is different with shard id value to encode ")
	case decodedIncrementNum != sequenceNumToEncode:
		return StepStateToContext(ctx, stepState), errors.New("decoded sequence num value is different with sequence num value to encode ")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theClientReceiveValidShardedId(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	return StepStateToContext(ctx, stepState), nil
}
