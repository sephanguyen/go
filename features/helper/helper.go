package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type SyncedAuthUser struct {
	UserID string `json:"user_id"`

	Err error `json:"-"`
}

type AuthUserListener func() (<-chan SyncedAuthUser, func(), error)

func NewAuthUserListener(ctx context.Context, authPostgresDB *pgxpool.Pool) AuthUserListener {
	// Temporary approach to deal with delay time by data syncing from bob db to auth db
	// After we switched to auth db completely, we can safely remove this
	return func() (<-chan SyncedAuthUser, func(), error) {
		listener := make(chan SyncedAuthUser, 1000)

		acquiredConn, err := authPostgresDB.Acquire(ctx)
		if err != nil {
			return nil, nil, errors.Wrap(err, "authPostgresDB.Acquire")
		}
		_, err = acquiredConn.Exec(ctx, "LISTEN synced_auth_user")
		if err != nil {
			return nil, nil, errors.Wrap(err, "acquiredConn.Exec")
		}

		// Acquired connection will use this cancelable context
		// When caller cancel this context, WaitForNotification() will return err (ctx canceled)
		// and the goroutine will exit
		listenerCtx, stop := context.WithCancel(context.Background())
		go func(ctx context.Context) {
			defer func() {
				close(listener)
				acquiredConn.Release()
			}()
			for {
				// WaitForNotification will block until receive a notification or an error
				notification, err := acquiredConn.Conn().WaitForNotification(ctx)
				if err != nil {
					return
				}
				// fmt.Println("PID:", notification.PID, "Channel:", notification.Channel, "Payload:", notification.Payload)

				var syncedAuthUser SyncedAuthUser
				if err := json.Unmarshal([]byte(notification.Payload), &syncedAuthUser); err != nil {
					listener <- SyncedAuthUser{
						Err: err,
					}
					continue
				}

				select {
				case listener <- syncedAuthUser:
					continue
				case <-time.After(5 * time.Second):
					// slow consumer, current policy is drop message
					continue
				}
			}
		}(listenerCtx)

		// Give the caller a closure to cancel the context that acquired connection uses
		shutdownFunc := func() {
			stop()
		}

		return listener, shutdownFunc, nil
	}
}

func ExchangeToken(originalToken, userID, userGroup, applicantID string, schoolID int64, conn grpc.ClientConnInterface, listenerFuncs ...AuthUserListener) (string, error) {
	// 5s timeout to avoid DeadlineExceeded error
	/*ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()*/

	fanInSyncedUsers := make(chan SyncedAuthUser, 10000)
	stopSyncedUserFanInChan := make(chan struct{})
	clearResourceFuncs := make([]func(), 0, len(listenerFuncs))

	for _, listenerFunc := range listenerFuncs {
		listener, stop, err := listenerFunc()
		if err != nil {
			return "", errors.Wrap(err, "listenerFunc()")
		}
		clearResourceFuncs = append(clearResourceFuncs, stop)

		// listenerFuncs is a slice, so we need to perform a fan-in
		go func() {
			for {
				select {
				case syncedAuthUser := <-listener:
					if syncedAuthUser.Err != nil {
						// fmt.Println("syncedAuthUser.Err", syncedAuthUser.Err)
						// current policy is ignoring data have invalid format
						continue
					}
					fanInSyncedUsers <- syncedAuthUser
				case <-stopSyncedUserFanInChan:
					return
				}
			}
		}()
	}

	defer func() {
		for _, clearResourceFunc := range clearResourceFuncs {
			clearResourceFunc()
		}
		close(stopSyncedUserFanInChan)
	}()

	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	// fmt.Printf("expect user %v \n", userID)

	attempts := 0
	for {
		// skip this for first time
		if attempts > 0 {
			select {
			case syncedAuthUser := <-fanInSyncedUsers:
				if syncedAuthUser.UserID != userID {
					continue
				}
				// fmt.Printf("found user %v \n", userID)
			case <-ticker.C:
				// We can wait for a notification, but occasionally we may have a slow
				// consumer issue and current policy is dropping message.
				// This prevents the process blocks forever if the message is never
				// delivered to this listener

				// fmt.Printf("still not found user: %v \n", userID)
			}
		}

		token, err := func() (string, error) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			resp, err := spb.NewTokenReaderServiceClient(conn).ExchangeToken(ctx, &spb.ExchangeTokenRequest{
				NewTokenInfo: &spb.ExchangeTokenRequest_TokenInfo{
					Applicant:    applicantID,
					UserId:       userID,
					DefaultRole:  userGroup,
					AllowedRoles: []string{userGroup},
					SchoolIds:    []int64{schoolID},
				},
				OriginalToken: originalToken,
			})
			if err != nil {
				return "", err
			}
			return resp.NewToken, err
		}()

		if err == nil {
			return token, nil
		}
		if err != nil {
			// fmt.Println(err)
			attempts++
			// fmt.Println("attempts:", attempts)
			if attempts >= 120 {
				// fmt.Println("end attempts:", attempts, err)
				return "", err
			}
		}
	}

	/*var rsp *spb.ExchangeTokenResponse
	err := try.Do(func(attempt int) (bool, error) {
		resp, err := spb.NewTokenReaderServiceClient(conn).ExchangeToken(ctx, &spb.ExchangeTokenRequest{
			NewTokenInfo: &spb.ExchangeTokenRequest_TokenInfo{
				Applicant:    applicantID,
				UserId:       userID,
				DefaultRole:  userGroup,
				AllowedRoles: []string{userGroup},
				SchoolIds:    []int64{schoolID},
			},
			OriginalToken: originalToken,
		})
		if err == nil {
			rsp = resp
			return false, nil
		}
		if attempt < 10 {
			time.Sleep(time.Millisecond * 200)
			return true, fmt.Errorf("spb.NewTokenReaderServiceClient(conn).ExchangeToken %v", err)
		}
		return false, fmt.Errorf("exceed retryTimes %v", err)
	})
	if err != nil {
		return "", err
	}
	return rsp.NewToken, nil*/
}

var (
	JapaneseKatakana = []int64{12449, 12531}
	JapaneseHiragana = []int64{12353, 12435}
	English          = []int64{65, 90}
)

func RandInt(start, end int64) int64 {
	rand.Seed(time.Now().UnixNano())
	return (start + rand.Int63n(end-start)) //nolint:gosec
}

func GenerateRandomRune(size int, start, end int64) string {
	randRune := make([]rune, size)
	for i := range randRune {
		randRune[i] = rune(RandInt(start, end))
	}
	return string(randRune)
}

// Deprecated: use BuildRegexpMapV2 with checking step signature instead
func BuildRegexpMap(steps map[string]interface{}) map[string]*regexp.Regexp {
	m := make(map[string]*regexp.Regexp, len(steps))
	for k := range steps {
		m[k] = regexp.MustCompile(k)
	}
	return m
}

func BuildRegexpMapV2(steps map[string]interface{}) map[string]*regexp.Regexp {
	m := make(map[string]*regexp.Regexp, len(steps))
	for stepDef, stepFunc := range steps {
		if err := CheckStepFunc(stepFunc); err != nil {
			panic(fmt.Errorf("BuildRegexpMap error: %w (step: %s)", err, stepDef))
		}

		m[stepDef] = regexp.MustCompile(stepDef)
	}

	return m
}

var (
	errorInterface   = reflect.TypeOf((*error)(nil)).Elem()
	contextInterface = reflect.TypeOf((*context.Context)(nil)).Elem()
)

func CheckStepFunc(stepFunc interface{}) error {
	v := reflect.ValueOf(stepFunc)
	typ := v.Type()
	if typ.Kind() != reflect.Func {
		return fmt.Errorf("expected handler to be func, but got: %T", stepFunc)
	}

	if typ.NumOut() != 2 {
		return fmt.Errorf("expected handler to return two values (context, error), but it has: %d", typ.NumOut())
	}

	if firstOut := typ.Out(0); firstOut.Kind() != reflect.Interface || !firstOut.Implements(contextInterface) {
		return fmt.Errorf("expected handler to return first value is a context, but it has %v", firstOut.Kind())
	}

	if secondOut := typ.Out(1); secondOut.Kind() != reflect.Interface || !secondOut.Implements(errorInterface) {
		return fmt.Errorf("expected handler to return second value is an error, but it has %v", secondOut.Kind())
	}

	return nil
}

func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[RandInt(0, int64(len(letters)))]
	}
	return string(s)
}

func GRPCContext(ctx context.Context, key, val string) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.Pairs(key, val, "pkg", "com.manabie.liz", "version", "1.0.0"))
}

func CombineFirstNameAndLastNameToFullName(firstName, lastName string) string {
	return lastName + " " + firstName
}
func CombineFirstNamePhoneticAndLastNamePhoneticToFullName(firstNamePhonetic, lastNamePhonetic string) string {
	return strings.TrimSpace(lastNamePhonetic + " " + firstNamePhonetic)
}
func SplitNameToFirstNameAndLastName(fullname string) (firstName, lastName string) {
	if fullname == "" {
		return
	}
	splitNames := regexp.MustCompile(" +|　+").Split(fullname, 2)
	lastName = splitNames[0]

	if len(splitNames) == 2 {
		firstName = splitNames[1]
	}
	return
}

func CreateMultipleSchoolAdmins(ctx context.Context, db database.QueryExecer, schoolAdmins []*entity.SchoolAdmin) error {
	ctx, span := interceptors.StartSpan(ctx, "SchoolAdmin.CreateMultiple")
	defer span.End()

	queueFn := func(batch *pgx.Batch, u *entity.SchoolAdmin) {
		fields, values := u.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT DO NOTHING",
			u.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)

		batch.Queue(stmt, values...)
	}

	batch := &pgx.Batch{}
	now := time.Now()

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	for _, u := range schoolAdmins {
		_ = u.UpdatedAt.Set(now)
		_ = u.CreatedAt.Set(now)
		if u.ResourcePath.Status == pgtype.Null {
			err := u.ResourcePath.Set(resourcePath)
			if err != nil {
				return err
			}
		}
		queueFn(batch, u)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < len(schoolAdmins); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "schoolAdmin not inserted, batchResults.Exec")
		}
	}

	return nil
}

func CreateUser(ctx context.Context, db database.QueryExecer, user *entity.LegacyUser) error {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		user.UpdatedAt.Set(now),
		user.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	if user.ResourcePath.Status == pgtype.Null {
		err := user.ResourcePath.Set(resourcePath)
		if err != nil {
			return err
		}
	}
	_, err := database.InsertExceptOnConflictDoNothing(ctx, user, []string{"remarks"}, db.Exec)
	if err != nil {
		return fmt.Errorf("user not inserted: %w", err)
	}

	return nil
}
