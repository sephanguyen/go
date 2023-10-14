package services

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bobv1 "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// services.ValidateAuth( check allowedGroups
func ValidateAuth(ctx context.Context, db database.QueryExecer, userGroup func(context.Context, database.QueryExecer, pgtype.Text) (string, error), allowedGroups ...string) (string, error) {
	currentUserID := interceptors.UserIDFromContext(ctx)
	uGroup, err := userGroup(ctx, db, database.Text(currentUserID))
	if err != nil {
		return "", status.Error(codes.Unknown, err.Error())
	}
	if len(allowedGroups) == 0 {
		allowedGroups = append(allowedGroups, entities.UserGroupStudent, entities.UserGroupAdmin)
	}
	for _, group := range allowedGroups {
		if uGroup == group {
			return currentUserID, nil
		}
	}
	return "", status.Error(codes.PermissionDenied, "user group not allowed")
}

func canProcessStudentData(ctx context.Context, studentID string) bool {
	currentUserID := interceptors.UserIDFromContext(ctx)
	uGroup := interceptors.UserGroupFromContext(ctx)

	switch uGroup {
	case entities.UserGroupStudent:
		// only allows student get his owned plans
		return currentUserID == studentID
	default:
		return true
	}
}

func ConfigMap(configs []*entities.Config) map[string]*entities.Config {
	m := make(map[string]*entities.Config, len(configs))
	for _, v := range configs {
		m[v.Key.String] = v
	}

	return m
}

func inArrayString(s string, arr []string) bool {
	return golibs.InArrayString(s, arr)
}

const runes = "ABCDEFGHJKLMNPQRSTUVWXYZ12345689"

func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return string(b)
}

// Character to use, exclude 0,1,I,O
var (
	charMap = []string{
		"U", "V", "W", "X", "Y", "Z",
		"A", "B", "C", "D", "E", "F", "G", "H", "J",
		"2", "3", "4", "5", "6", "7", "8", "9",
		"K", "L", "M", "N", "P", "Q", "R", "S", "T",
	}
)

const (
	codePrefix = "MO" // Manabie order
	randLen    = 3
	randMin    = 101
	randMax    = 999
)

func getRandValue(seed uint, randMin, randMax, randLen int) int {
	localRand := rand.New(rand.NewSource(int64(seed)))
	rawValue := fmt.Sprintf("%d", localRand.Intn((randMax-randMin)+1)+randMin)
	randValue, _ := strconv.Atoi(rawValue[:randLen])
	return randValue
}

func dec2n(val int64) string {
	base := len(charMap)
	var result string
	for {
		index := val % int64(base)
		result = charMap[index] + result
		if val = int64(math.Floor(float64(val) / float64(base))); val == 0 {
			break
		}
	}
	return result
}

func reserveString(s string) (result string) {
	for _, v := range s {
		result = string(v) + result
	}

	return
}

func sliceIndex(length int, f func(i int) bool) int {
	for i := 0; i < length; i++ {
		if f(i) {
			return i
		}
	}

	return -1
}

func n2dec(val string) (string, error) {
	base := len(charMap)

	var decode int64 = 0
	val = reserveString(val)
	for digit := 0; digit < len(val); digit++ {
		char := string(val[digit])
		if !isContain(charMap, char) {
			return "", errors.New("invalid input")
		}

		index := int64(sliceIndex(len(charMap), func(i int) bool { return charMap[i] == char }))
		decode += index * int64(math.Pow(float64(base), float64(digit)))
	}
	return fmt.Sprintf("%d", decode), nil
}

func isContain(arr []string, value string) bool {
	for _, item := range arr {
		if item == value {
			return true
		}
	}
	return false
}

func EncodeOrderID2String(n uint) string {
	return codePrefix + EncodeNumber2String(n)
}

func DecodeString2OrderID(encoded string) (uint, error) {
	encoded = strings.Replace(encoded, codePrefix, "", -1)
	return DecodeString2Number(encoded)
}

func EncodeNumber2String(n uint) string {
	randVal := getRandValue(n, randMin, randMax, randLen)

	idTemp := fmt.Sprintf("%03d%09d", randVal, n)
	id, _ := strconv.ParseInt(idTemp, 10, 64)
	return dec2n(id)
}

func DecodeString2Number(encoded string) (uint, error) {
	decoded, err := n2dec(encoded)
	if err != nil {
		return 0, err
	}
	if len(decoded) < randLen+1 {
		return 0, errors.New("invalid encoded")
	}
	id, _ := strconv.ParseInt(decoded[randLen:], 10, 64)
	decodedRandVal, _ := strconv.Atoi(decoded[:randLen])

	randVal := getRandValue(uint(id), randMin, randMax, randLen)
	if randVal != decodedRandVal {
		return 0, errors.New("invalid rand value")
	}

	return uint(id), nil
}

func StatusErrWithDetail(code codes.Code, message string, detail proto.Message) error {
	stt := status.New(code, message)
	stt, _ = stt.WithDetails(detail)

	return stt.Err()
}

func StructToMap(s *types.Struct) map[string]interface{} {
	if s == nil {
		return nil
	}
	m := map[string]interface{}{}
	for k, v := range s.Fields {
		m[k] = decodeValue(v)
	}
	return m
}

func decodeValue(v *types.Value) interface{} {
	switch k := v.Kind.(type) {
	case *types.Value_NullValue:
		return nil
	case *types.Value_NumberValue:
		return k.NumberValue
	case *types.Value_StringValue:
		return k.StringValue
	case *types.Value_BoolValue:
		return k.BoolValue
	case *types.Value_StructValue:
		return StructToMap(k.StructValue)
	case *types.Value_ListValue:
		s := make([]interface{}, len(k.ListValue.Values))
		for i, e := range k.ListValue.Values {
			s[i] = decodeValue(e)
		}
		return s
	default:
		panic("protostruct: unknown kind")
	}
}

func getSchool(schoolID int32) int32 {
	if schoolID != 0 {
		return schoolID
	}
	return constants.ManabieSchool
}

// TODO: store msg can not push
func HandlePushMsgFail(ctx context.Context, msg proto.Marshaler, err error) error {
	return err
}

func inArraySubject(v pb.Subject, arr []pb.Subject) bool {
	for _, e := range arr {
		if v == e {
			return true
		}
	}
	return false
}

func FormatPromotionCodeExpiredDate(c pb.Country, expiredDate time.Time) string {
	switch c {
	case pb.COUNTRY_VN:
		return expiredDate.In(timeutil.Timezone(c)).Format("02/01")
	default:
		return expiredDate.Format("January 02")
	}
}

func createStudentOrder(ctx context.Context, db database.QueryExecer, create func(context.Context, database.QueryExecer, *entities.StudentOrder) error, updateReferenceNumber func(context.Context, database.QueryExecer, pgtype.Int4, pgtype.Text) error, order *entities.StudentOrder) error {
	err := create(ctx, db, order)
	if err != nil {
		return fmt.Errorf("s.StudentOrderRepo.Create: %w", err)
	}

	order.ReferenceNumber = database.Text(EncodeOrderID2String(uint(order.ID.Int)))
	err = updateReferenceNumber(ctx, db, order.ID, order.ReferenceNumber)
	if err != nil {
		return fmt.Errorf("updateReferenceNumber: %w", err)
	}

	return nil
}

type userGroupFetcherFunc func(context.Context, database.QueryExecer, pgtype.Text) (string, error)

// CheckUserGroup returns nil if user is expected group
func CheckUserGroup(ctx context.Context, db database.QueryExecer, userID pgtype.Text, expectedGroup, errMsg string, groupFetcher userGroupFetcherFunc) error {
	uGroup, err := groupFetcher(ctx, db, userID)
	if err != nil {
		return toStatusError(err)
	}

	if uGroup != expectedGroup {
		return status.Error(codes.PermissionDenied, errMsg)
	}

	return nil
}

// ToConversionTaskStatus converts cloud convert job status to entity status.
func ToConversionTaskStatus(status string) string {
	switch status {
	case "job.finished":
		return bobv1.ConversionTaskStatus_CONVERSION_TASK_STATUS_FINISHED.String()
	case "job.failed":
		return bobv1.ConversionTaskStatus_CONVERSION_TASK_STATUS_FAILED.String()
	default:
		return bobv1.ConversionTaskStatus_CONVERSION_TASK_STATUS_INVALID.String()
	}
}
