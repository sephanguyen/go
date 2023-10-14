package usermgmt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	enigma_entites "github.com/manabie-com/backend/internal/enigma/entities"
	golibs_auth "github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	location_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc/importstudent"
	http_port "github.com/manabie-com/backend/internal/usermgmt/modules/user/port/http"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	pkg_unleash "github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/gocarina/gocsv"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// func CreateARandomSchoolAdmin(ctx context.Context, db database.Ext, tenantManager internal_auth_tenant.TenantManager) (entity.DomainSchoolAdmin, error) {
// 	// Random school admin
// 	randomSchoolAdmin := &SchoolAdmin{
// 		randomID: newID(),
// 	}
// 	// Use domain service to create school admin
// 	schoolAdminDomainService := service.DomainSchoolAdmin{
// 		DB:                    db,
// 		TenantManager:         tenantManager,
// 		DomainSchoolAdminRepo: repository.NewDomainSchoolAdminRepo(),
// 	}
// 	err := schoolAdminDomainService.CreateSchoolAdmin(ctx, randomSchoolAdmin)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "CreateSchoolAdmin")
// 	}
// 	return randomSchoolAdmin, nil
// }

// // SignedAsSchoolAdmin func create a random school admin by following our business logic,
// // also login in that user on identity platform to get id token and exchange it with shamir
// // service to get final exchanged token
// func SignedAsSchoolAdmin(ctx context.Context, db database.Ext, tenantManager internal_auth_tenant.TenantManager, shamirConn grpc.ClientConnInterface, apiKey string, fakeFirebaseAddress string) (string, error) {
// 	// We have an initial school admin that is created manually
// 	// use that school admin to log in and use its token to call API
// 	// ......

// 	// Random school admin
// 	randomSchoolAdmin, err := CreateARandomSchoolAdmin(ctx, db, tenantManager)
// 	if err != nil {
// 		return "", errors.Wrap(err, "CreateARandomSchoolAdmin")
// 	}

// 	idToken, err := common.GenerateAuthenticationToken(fakeFirebaseAddress, randomSchoolAdmin.UserID().String(), entity.UserGroupSchoolAdmin)
// 	if err != nil {
// 		return "", errors.Wrap(err, "generateExchangeToken")
// 	}

// 	/*//Login in to get id token
// 	idToken, err := LoginInAuthPlatform(ctx, apiKey, "", randomSchoolAdmin.Email().String(), randomSchoolAdmin.Password().String())
// 	if err != nil {
// 		return "", err
// 	}*/

// 	exchangedToken, err := ExchangeToken(ctx, shamirConn, randomSchoolAdmin.Email().String(), randomSchoolAdmin.Password().String(), idToken)
// 	if err != nil {
// 		return "", err
// 	}

// 	return exchangedToken, nil
// }

// func CreateInitialSchoolAdmin(ctx context.Context, db *pgxpool.Pool, tenantManager internal_auth_tenant.TenantManager, jwtApplicant string, apiKey string, orgID int32) (string, error) {
// 	organization := &Organization{
// 		organizationID: strconv.Itoa(int(orgID)),
// 		schoolID:       orgID,
// 	}
// 	schoolAdminProfile := &SchoolAdmin{randomID: idutil.ULIDNow()}
// 	schoolAdmin := entity.SchoolAdminToDelegate{
// 		DomainSchoolAdminProfile: schoolAdminProfile,
// 		HasSchoolID:              organization,
// 		HasOrganizationID:        organization,
// 		HasCountry: &repository.User{
// 			UserAttribute: repository.UserAttribute{
// 				Country: field.NewString(cpb.Country_COUNTRY_VN.String()),
// 			},
// 		},
// 		HasUserID: schoolAdminProfile,
// 	}
// 	legacyUserGroup := &entity.LegacyUserGroupWillBeDelegated{
// 		LegacyUserGroupAttribute: &entity.SchoolAdminLegacyUserGroup{},
// 		HasOrganizationID:        organization,
// 		HasUserID:                schoolAdmin,
// 	}
// 	// Initial school admins
// 	schoolAdminAggregate := aggregate.DomainSchoolAdmin{
// 		DomainSchoolAdmin: schoolAdmin,
// 		LegacyUserGroups:  entity.LegacyUserGroups{legacyUserGroup},
// 	}
// 	err := repository.NewDomainSchoolAdminRepo().Create(internal_auth.InjectFakeJwtToken(ctx, ""), db, schoolAdminAggregate)
// 	if err != nil {
// 		return "", err
// 	}

// 	tenantID, err := new(repository.OrganizationRepo).WithDefaultValue("local").GetTenantIDByOrgID(ctx, db, organization.OrganizationID().String())
// 	if err != nil {
// 		return "", err
// 	}

// 	backwardCompatibleAuthUser := &entity.User{
// 		ID:    database.Text(schoolAdmin.UserID().String()),
// 		Email: database.Text(schoolAdmin.Email().String()),
// 		UserAdditionalInfo: entity.UserAdditionalInfo{
// 			Password: schoolAdmin.Password().String(),
// 		},
// 	}

// 	err = service.CreateUsersInIdentityPlatform(ctx, tenantManager, tenantID, entity.Users{backwardCompatibleAuthUser}, int64(organization.SchoolID().Int32()), true)
// 	if err != nil {
// 		return "", err
// 	}

// 	idToken, err := LoginInAuthPlatform(ctx, apiKey, tenantID, schoolAdmin.Email().String(), schoolAdmin.Password().String())
// 	if err != nil {
// 		return "", err
// 	}

// 	exchangedToken, err := ExchangeToken(ctx, connections.ShamirConn, jwtApplicant, schoolAdmin.UserID().String(), idToken)
// 	if err != nil {
// 		return "", err
// 	}

// 	return exchangedToken, nil
// }

type LoginInAuthPlatformResult struct {
	IDToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
}

// LoginInAuthPlatform login in auth platform, returns id token
func LoginInAuthPlatform(ctx context.Context, apiKey string, tenantID string, email string, password string) (*LoginInAuthPlatformResult, error) {
	url := fmt.Sprintf("%s%s", IdentityToolkitURL, apiKey)

	loginInfo := struct {
		TenantID          string `json:"tenantId,omitempty"`
		Email             string `json:"email"`
		Password          string `json:"password"`
		ReturnSecureToken bool   `json:"returnSecureToken"`
	}{
		TenantID:          tenantID,
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}
	body, err := json.Marshal(&loginInfo)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to login firebase and failed to decode error")
	}

	if resp.StatusCode == http.StatusOK {
		r := &LoginInAuthPlatformResult{}
		if err := json.Unmarshal(data, &r); err != nil {
			return nil, errors.Wrap(err, "failed to login and failed to decode error")
		}
		return r, nil
	}

	return nil, errors.New("failed to login firebase" + string(data))
}

type ExchangeIDTokenByRefreshTokenResult struct {
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
}

func ExchangeIDTokenByRefreshToken(ctx context.Context, apiKey string, refreshToken string) (*ExchangeIDTokenByRefreshTokenResult, error) {
	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/token?key=%s", apiKey)

	reqBody := struct {
		GrantType    string `json:"grant_type"`
		RefreshToken string `json:"refresh_token"`
	}{
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
	}
	body, err := json.Marshal(&reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to exchange id token by refresh token")
	}

	if resp.StatusCode == http.StatusOK {
		r := &ExchangeIDTokenByRefreshTokenResult{}
		if err := json.Unmarshal(data, &r); err != nil {
			return nil, errors.Wrap(err, "failed to exchange id token by refresh token")
		}
		return r, nil
	}

	return nil, errors.New("failed to exchange id token by refresh token" + string(data))
}

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

// ExchangeToken exchange id token with shamir service to get manabie token
func ExchangeToken(_ context.Context, conn grpc.ClientConnInterface, applicant, userID, originalToken string, listenerFuncs ...AuthUserListener) (string, error) {
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
					Applicant: applicant,
					UserId:    userID,
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
				Applicant: applicant,
				UserId:    userID,
			},
			OriginalToken: originalToken,
		})
		if err == nil {
			rsp = resp
			return false, nil
		}
		if attempt < retryTimes {
			time.Sleep(time.Millisecond * 500)
			return true, fmt.Errorf("spb.NewTokenReaderServiceClient(conn).ExchangeToken %v", err)
		}
		return false, fmt.Errorf("exceed retryTimes %v", err)
	})
	if err != nil {
		return "", err
	}
	return rsp.NewToken, nil*/
}

func GenFakeIDToken(firebaseAddr, userID, template string) (string, error) {
	resp, err := http.Get("http://" + firebaseAddr + "/token?template=" + template + "&UserID=" + userID)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot generate new user token, err: %v", err)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()

	return string(bodyResp), nil
}

func (s *suite) signedAsAccount(ctx context.Context, group string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, group)
	stepState.CurrentUserID = s.MapOrgStaff[constants.ManabieSchool][GetRoleFromConstant(group)].UserID
	stepState.AuthToken = s.MapOrgStaff[constants.ManabieSchool][GetRoleFromConstant(group)].Token
	stepState.OrganizationID = strconv.Itoa(constants.ManabieSchool)
	stepState.CurrentSchoolID = constants.ManabieSchool
	ctx = interceptors.ContextWithUserID(ctx, stepState.CurrentUserID)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aSignedInAsInOrganization(ctx context.Context, group string, org string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	orgID := 0

	switch org {
	case "manabie":
		orgID = constants.ManabieSchool
	case "jprep":
		orgID = constants.JPREPSchool
	case "kec-demo":
		orgID = constants.KECDemo
	}
	ctx = s.signedIn(ctx, orgID, group)
	stepState.CurrentUserID = s.MapOrgStaff[orgID][GetRoleFromConstant(group)].UserID
	stepState.AuthToken = s.MapOrgStaff[orgID][GetRoleFromConstant(group)].Token
	stepState.OrganizationID = strconv.Itoa(orgID)
	stepState.CurrentSchoolID = int32(orgID)
	ctx = interceptors.ContextWithUserID(ctx, stepState.CurrentUserID)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) signedIn(ctx context.Context, orgID int, role string) context.Context {
	authInfo := s.getAuthInfo(orgID, role)
	ctx = contextWithTokenV2(ctx, authInfo.Token)
	ctx = interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: strconv.Itoa(orgID),
			UserID:       authInfo.UserID,
			UserGroup:    GetLegacyUserGroupFromConstant(role),
		},
	})
	return ctx
}

func (s *suite) getAuthInfo(orgID int, account string) common.AuthInfo {
	switch role := GetRoleFromConstant(account); role {
	case unauthenticatedType:
		return common.AuthInfo{
			UserID: newID(),
			Token:  invalidToken,
		}

	default:
		return s.MapOrgStaff[orgID][role]
	}
}

func createPartnerSyncDataLog(ctx context.Context, db database.Ext, signature string, hours time.Duration) (*enigma_entites.PartnerSyncDataLog, error) {
	stepState := StepStateFromContext(ctx)

	partnerSyncDataLog := &enigma_entites.PartnerSyncDataLog{}
	now := time.Now()
	newPartnerSyncDataLogID := idutil.ULIDNow()
	stepState.PartnerSyncDataLogID = newPartnerSyncDataLogID

	err := multierr.Combine(
		partnerSyncDataLog.PartnerSyncDataLogID.Set(newPartnerSyncDataLogID),
		partnerSyncDataLog.Signature.Set(signature),
		partnerSyncDataLog.Payload.Set([]byte("{}")),
		partnerSyncDataLog.CreatedAt.Set(now.Add(-hours*time.Hour)),
		partnerSyncDataLog.UpdatedAt.Set(now.Add(-hours*time.Hour)),
	)
	if err != nil {
		return nil, fmt.Errorf("createPartnerSyncDataLog: %s", err.Error())
	}
	if _, err = database.InsertIgnoreConflict(ctx, partnerSyncDataLog, db.Exec); err != nil {
		return nil, fmt.Errorf("insert partner sync data log err: %w", err)
	}
	return partnerSyncDataLog, nil
}

func createLogSyncDataSplit(ctx context.Context, db database.Ext, kind string) (*enigma_entites.PartnerSyncDataLogSplit, error) {
	stepState := StepStateFromContext(ctx)
	partnerSyncDataLogSplit := &enigma_entites.PartnerSyncDataLogSplit{}

	database.AllNullEntity(partnerSyncDataLogSplit)
	now := time.Now()

	newPartnerSyncDataLogSplitID := idutil.ULIDNow()
	stepState.PartnerSyncDataLogSplitID = newPartnerSyncDataLogSplitID
	err := multierr.Combine(
		partnerSyncDataLogSplit.PartnerSyncDataLogSplitID.Set(newPartnerSyncDataLogSplitID),
		partnerSyncDataLogSplit.PartnerSyncDataLogID.Set(stepState.PartnerSyncDataLogID),
		partnerSyncDataLogSplit.Payload.Set([]byte("{}")),
		partnerSyncDataLogSplit.Kind.Set(kind),
		partnerSyncDataLogSplit.Status.Set(string(enigma_entites.StatusPending)),
		partnerSyncDataLogSplit.RetryTimes.Set(0),
		partnerSyncDataLogSplit.CreatedAt.Set(now),
		partnerSyncDataLogSplit.UpdatedAt.Set(now),
	)
	if err != nil {
		return nil, fmt.Errorf("createLogSyncDataSplit: %s", err.Error())
	}

	if _, err = database.InsertIgnoreConflict(ctx, partnerSyncDataLogSplit, db.Exec); err != nil {
		return nil, fmt.Errorf("insert partner sync data log split id err: %w", err)
	}
	return partnerSyncDataLogSplit, nil
}

func (s *suite) createTagsType(ctx context.Context, userTagType string) ([]string, []string, error) {
	var userTag, userDiscountTag string

	switch userTagType {
	case studentType:
		userTag = upb.UserTagType_USER_TAG_TYPE_STUDENT.String()
		userDiscountTag = upb.UserTagType_USER_TAG_TYPE_STUDENT_DISCOUNT.String()
	case parentType:
		userTag = upb.UserTagType_USER_TAG_TYPE_PARENT.String()
		userDiscountTag = upb.UserTagType_USER_TAG_TYPE_PARENT_DISCOUNT.String()
	default:
		return nil, nil, nil
	}

	tagIDs, tagPartnerIDs, err := s.createAmountTags(ctx, amountSampleTestElement, userTag, fmt.Sprint(constants.ManabieSchool))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "createAmountTags for %s", userTag)
	}

	discountTagIDs, discountTagPartnerIDs, err := s.createAmountTags(ctx, amountSampleTestElement, userDiscountTag, fmt.Sprint(constants.ManabieSchool))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "createAmountTags for %s", userDiscountTag)
	}
	return append(tagIDs, discountTagIDs...),
		append(tagPartnerIDs, discountTagPartnerIDs...),
		nil
}

func getChildrenLocation(orgID int) []string {
	switch orgID {
	case constants.ManabieSchool:
		return []string{existingLocations[0].LocationID.String, existingLocations[2].LocationID.String}
	case constants.JPREPSchool:
		return []string{existingLocations[1].LocationID.String}
	case constants.TestingSchool:
		return []string{existingLocations[3].LocationID.String}
	default:
		return nil
	}
}

func insertLocations(ctx context.Context, locations []*location_repo.Location, db database.QueryExecer) error {
	queue := func(b *pgx.Batch, location *location_repo.Location) {
		fieldNames := database.GetFieldNames(location)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf("INSERT INTO locations (%s) "+
			"VALUES (%s) ON CONFLICT(location_id) DO "+
			"UPDATE SET updated_at = now(), name = $2, location_type=$3, parent_location_id=$4,partner_internal_id=$5,partner_internal_parent_id=$6, deleted_at = NULL", strings.Join(fieldNames, ", "), placeHolders)
		b.Queue(query, database.GetScanFields(location, fieldNames)...)
	}

	now := time.Now()
	batch := &pgx.Batch{}

	for _, location := range locations {
		if err := multierr.Combine(
			location.CreatedAt.Set(now),
			location.UpdatedAt.Set(now),
		); err != nil {
			return err
		}

		queue(batch, location)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}

	return nil
}

func assignUserGroupToUser(ctx context.Context, dbBob database.QueryExecer, userID string, userGroupIDs []string) error {
	userGroupMembers := make([]*entity.UserGroupMember, 0)
	currentSchoolID := StepStateFromContext(ctx).CurrentSchoolID
	if currentSchoolID == 0 {
		currentSchoolID = constants.ManabieSchool
	}

	for _, userGroupID := range userGroupIDs {
		userGroupMem := &entity.UserGroupMember{}
		database.AllNullEntity(userGroupMem)
		if err := multierr.Combine(
			userGroupMem.UserID.Set(userID),
			userGroupMem.UserGroupID.Set(userGroupID),
			userGroupMem.ResourcePath.Set(fmt.Sprint(currentSchoolID)),
		); err != nil {
			return err
		}
		userGroupMembers = append(userGroupMembers, userGroupMem)
	}

	if err := (&repository.UserGroupsMemberRepo{}).UpsertBatch(ctx, dbBob, userGroupMembers); err != nil {
		return errors.Wrapf(err, "assignUserGroupToUser")
	}
	return nil
}

func newUserEntity() (*entity.LegacyUser, error) {
	userID := newID()
	now := time.Now()
	user := new(entity.LegacyUser)
	firstName := fmt.Sprintf("user-first-name-%s", userID)
	lastName := fmt.Sprintf("user-last-name-%s", userID)
	fullName := helper.CombineFirstNameAndLastNameToFullName(firstName, lastName)
	database.AllNullEntity(user)
	database.AllNullEntity(&user.AppleUser)
	if err := multierr.Combine(
		user.ID.Set(userID),
		user.Email.Set(fmt.Sprintf("valid-user-%s@email.com", userID)),
		user.Avatar.Set(fmt.Sprintf("http://valid-user-%s", userID)),
		user.IsTester.Set(false),
		user.FacebookID.Set(userID),
		user.PhoneVerified.Set(false),
		user.AllowNotification.Set(true),
		user.EmailVerified.Set(false),
		user.FullName.Set(fullName),
		user.FirstName.Set(firstName),
		user.LastName.Set(lastName),
		user.Country.Set(cpb.Country_COUNTRY_VN.String()),
		user.Group.Set(entity.UserGroupStudent),
		user.Birthday.Set(now),
		user.Gender.Set(upb.Gender_FEMALE.String()),
		user.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
		user.CreatedAt.Set(now),
		user.UpdatedAt.Set(now),
		user.DeletedAt.Set(nil),
	); err != nil {
		return nil, errors.Wrap(err, "set value user")
	}

	user.UserAdditionalInfo = entity.UserAdditionalInfo{
		CustomClaims: map[string]interface{}{
			"external-info": "example-info",
		},
	}
	return user, nil
}

func newStudentEntity() (*entity.LegacyStudent, error) {
	now := time.Now()
	student := new(entity.LegacyStudent)
	database.AllNullEntity(student)
	database.AllNullEntity(&student.LegacyUser)
	user, err := newUserEntity()
	if err != nil {
		return nil, errors.Wrap(err, "newUserEntity")
	}
	student.LegacyUser = *user
	schoolID, err := strconv.ParseInt(student.LegacyUser.ResourcePath.String, 10, 32)
	if err != nil {
		return nil, errors.Wrap(err, "strconv.ParseInt")
	}

	if err := multierr.Combine(
		student.ID.Set(student.LegacyUser.ID),
		student.SchoolID.Set(schoolID),
		student.EnrollmentStatus.Set(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED),
		student.StudentExternalID.Set(student.LegacyUser.ID),
		student.StudentNote.Set("this is the note"),
		student.CurrentGrade.Set(1),
		student.TargetUniversity.Set("HUST"),
		student.TotalQuestionLimit.Set(32),
		student.OnTrial.Set(false),
		student.BillingDate.Set(now),
		student.CreatedAt.Set(student.LegacyUser.CreatedAt),
		student.UpdatedAt.Set(student.LegacyUser.UpdatedAt),
		student.DeletedAt.Set(student.LegacyUser.DeletedAt),
		student.PreviousGrade.Set(12),
	); err != nil {
		return nil, errors.Wrap(err, "set value student")
	}

	return student, nil
}

func (s *suite) createStudentUser(ctx context.Context) (entity.User, error) {
	orgID := OrgIDFromCtx(ctx)
	studentResp, err := CreateStudent(ctx, s.UserMgmtConn, nil, getChildrenLocation(orgID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create student")
	}

	return &repository.User{ID: field.NewString(studentResp.StudentProfile.Student.UserProfile.UserId)}, nil
}

func (s *suite) createParentUser(ctx context.Context) (entity.User, error) {
	user, err := s.createStudentUser(ctx)
	if err != nil {
		return nil, err
	}

	req := createParentReq(user.UserID().String())
	resp, err := upb.NewUserModifierServiceClient(s.UserMgmtConn).CreateParentsAndAssignToStudent(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create parent")
	}

	return &repository.User{ID: field.NewString(resp.ParentProfiles[0].Parent.UserProfile.UserId)}, nil
}

func (s *suite) createStaffUser(ctx context.Context) (entity.User, error) {
	roleWithLocationTeacher := RoleWithLocation{
		RoleName:    constant.RoleTeacher,
		LocationIDs: []string{constants.ManabieOrgLocation},
	}

	orgID := OrgIDFromCtx(ctx)
	resp, err := CreateStaff(ctx, s.BobDBTrace, s.UserMgmtConn, nil, []RoleWithLocation{roleWithLocationTeacher}, getChildrenLocation(orgID))
	if err != nil {
		return nil, fmt.Errorf("failed to create staff: %w", err)
	}

	return &repository.User{ID: field.NewString(resp.Staff.StaffId)}, nil
}

func (s *suite) createUserByRole(ctx context.Context, role string) (context.Context, error) {
	var (
		user entity.User
		err  error
	)

	switch role {
	case constant.RoleStudent:
		user, err = s.createStudentUser(ctx)
		s.CurrentUserGroup = constant.UserGroupStudent
	case constant.RoleParent:
		user, err = s.createParentUser(ctx)
		s.CurrentUserGroup = constant.UserGroupParent
	case "Staff":
		user, err = s.createStaffUser(ctx)
		s.CurrentUserGroup = constant.UserGroupTeacher
	}

	if err != nil {
		return ctx, fmt.Errorf("failed to create %s user: %w", role, err)
	}

	s.UserId = user.UserID().String()
	return ctx, nil
}

func (s *suite) verifyStudentsInBD(ctx context.Context, students []aggregate.DomainStudent) (context.Context, error) {
	for _, reqStudent := range students {
		userID := reqStudent.UserID().String()
		dbUsers, err := (&repository.DomainUserRepo{}).GetByIDs(ctx, s.BobPostgresDB, []string{userID})

		if err != nil {
			return ctx, errors.Wrap(err, "(&repository.DomainUserRepo{}).GetByEmails")
		}
		if len(dbUsers) == 0 {
			return ctx, fmt.Errorf("user %s is not found", userID)
		}
		dbUser := dbUsers[0]

		if _, err = s.verifyEnrollmentStatusHistories(ctx, reqStudent.EnrollmentStatusHistories, dbUser); err != nil {
			return ctx, fmt.Errorf("verifyEnrollmentStatusHistories err: %s", err)
		}
		if _, err = s.verifyGeneralInformationStudent(ctx, reqStudent.DomainStudent, dbUser, userID); err != nil {
			return ctx, fmt.Errorf("verifyGeneralInformationStudent err: %s", err)
		}
		if _, err = s.verifyUserAddress(ctx, reqStudent.UserAddress, dbUser); err != nil {
			return ctx, fmt.Errorf("verifyUserAddress err: %s", err)
		}
		if _, err = s.verifyUserPhoneNumbers(ctx, reqStudent, dbUser); err != nil {
			return ctx, fmt.Errorf("verifyUserPhoneNumbers err: %s", err)
		}
		if _, err = s.verifySchoolHistories(ctx, reqStudent, dbUser); err != nil {
			return ctx, fmt.Errorf("verifySchoolHistories err: %s", err)
		}
		if _, err = s.verifyStudentTags(ctx, reqStudent.Tags, dbUser); err != nil {
			return ctx, fmt.Errorf("verifyStudentTags err: %s", err)
		}
	}
	return ctx, nil
}

func (s *suite) verifyEnrollmentStatusHistories(ctx context.Context, reqEnrollmentStatusHistories entity.DomainEnrollmentStatusHistories, dbUser entity.User) (context.Context, error) {
	for _, reqEnrollmentStatusHistory := range reqEnrollmentStatusHistories {
		locations, err := (&repository.DomainLocationRepo{}).GetByIDs(ctx, s.BobPostgresDB, []string{reqEnrollmentStatusHistory.LocationID().String()})
		if err != nil {
			return ctx, errors.Wrap(err, "(&repository.DomainLocationRepo{}).GetByIDs")
		}
		enrollmentStatusHistoryRepo := &repository.DomainEnrollmentStatusHistoryRepo{}
		enrollmentStatusHistoryRes, err := enrollmentStatusHistoryRepo.
			GetByStudentIDAndLocationID(ctx, s.BobPostgresDB,
				dbUser.UserID().String(),
				locations[0].LocationID().String(),
				false,
			)

		if err != nil {
			return ctx, errors.Wrap(err, "(&repository.DomainEnrollmentStatusHistoryRepo{}).GetByStudentIDAndLocationID")
		}
		enrollmentStatusHistoryWillBeDelegated := entity.EnrollmentStatusHistoryWillBeDelegated{
			EnrollmentStatusHistory: reqEnrollmentStatusHistory,
			HasLocationID:           locations[0],
		}
		if len(enrollmentStatusHistoryRes) != 0 {
			currentEnrollmentStatus := enrollmentStatusHistoryRes.GetExactly(reqEnrollmentStatusHistory)
			if currentEnrollmentStatus == nil {
				return ctx, fmt.Errorf("enrollmentStatusHistory is not found, studentID %s", dbUser.UserID().String())
			}
			err = compareEntityEnrollmentStatusHistory(enrollmentStatusHistoryWillBeDelegated, currentEnrollmentStatus)
			if err != nil {
				return ctx, err
			}
		} else {
			return ctx, fmt.Errorf("enrollmentStatusHistory is not found")
		}
	}
	return ctx, nil
}

func compareEntityEnrollmentStatusHistory(expected entity.DomainEnrollmentStatusHistory, actual entity.DomainEnrollmentStatusHistory) error {
	switch {
	case expected.EnrollmentStatus().String() != actual.EnrollmentStatus().String():
		return fmt.Errorf(`compareEntityEnrollmentStatusHistory: expected upserted "enrollment_status": %v but actual is %v`,
			expected.EnrollmentStatus().String(),
			actual.EnrollmentStatus().String())
	case expected.LocationID().String() != actual.LocationID().String():
		return fmt.Errorf(`compareEntityEnrollmentStatusHistory: expected upserted "location": %v but actual is %v`,
			expected.LocationID().String(),
			actual.LocationID().String())
	case expected.StartDate().Time().Format(constant.DateLayout) != actual.StartDate().Time().Format(constant.DateLayout):
		if !expected.StartDate().Time().IsZero() {
			return fmt.Errorf(`compareEntityEnrollmentStatusHistory: expected upserted "start_date": %v but actual is %v`,
				expected.StartDate().Time().Format(constant.DateLayout),
				actual.StartDate().Time().Format(constant.DateLayout))
		}
	case expected.EndDate().Time().Format(constant.DateLayout) != actual.EndDate().Time().Format(constant.DateLayout):
		return fmt.Errorf(`compareEntityEnrollmentStatusHistory: expected upserted "end_date": %v but actual is %v`,
			expected.EndDate().Time().Format(constant.DateLayout),
			actual.EndDate().Time().Format(constant.DateLayout))
	}

	return nil
}

func (s *suite) verifyGeneralInformationStudent(ctx context.Context, reqStudent entity.DomainStudent, dbUser entity.User, userID string) (context.Context, error) {
	fullName := field.NewString(helper.CombineFirstNameAndLastNameToFullName(reqStudent.FirstName().String(), reqStudent.LastName().String()))
	fullNamePhonetic := field.NewString(helper.CombineFirstNamePhoneticAndLastNamePhoneticToFullName(reqStudent.FirstNamePhonetic().String(), reqStudent.LastNamePhonetic().String()))

	if field.IsPresent(reqStudent.ExternalStudentID()) {
		trimmedExternalUserID := strings.TrimSpace(reqStudent.ExternalStudentID().String())

		if dbUser.ExternalUserID().String() != trimmedExternalUserID {
			return ctx, fmt.Errorf(`validateUpsertStudents: expected upserted "external_user_id": %v but actual is %v`, trimmedExternalUserID, dbUser.ExternalUserID().String())
		}
	}
	switch {
	case userID != dbUser.UserID().String():
		return ctx, fmt.Errorf(`validateUpsertStudents: expected upserted "user_id": %v but actual is %v`, userID, dbUser.UserID().String())
	case reqStudent.FirstName().String() != dbUser.FirstName().String():
		return ctx, fmt.Errorf(`validateUpsertStudents: expected upserted "first_name": %v but actual is %v`, reqStudent.FirstName().String(), dbUser.FirstName().String())
	case fullName.String() != dbUser.FullName().String():
		return ctx, fmt.Errorf(`validateUpsertStudents: expected upserted "full_name": %v but actual is %v`, fullName.String(), dbUser.FullName().String())
	case reqStudent.LastName().String() != dbUser.LastName().String():
		return ctx, fmt.Errorf(`validateUpsertStudents: expected upserted "last_name": %v but actual is %v`, reqStudent.LastName().String(), dbUser.LastName().String())
	case reqStudent.FirstNamePhonetic().IsEmpty() && field.IsPresent(dbUser.FirstNamePhonetic()):
		return ctx, fmt.Errorf(`validateUpsertStudents: expected "first_name_phonetic" should be null, but actual:%v`, dbUser.FirstNamePhonetic().String())
	case reqStudent.LastNamePhonetic().IsEmpty() && field.IsPresent(dbUser.LastNamePhonetic()):
		return ctx, fmt.Errorf(`validateUpsertStudents: expected "last_name_phonetic" should be null, but actual:%v`, dbUser.LastNamePhonetic().String())
	case fullNamePhonetic.IsEmpty() && field.IsPresent(dbUser.FullNamePhonetic()):
		return ctx, fmt.Errorf(`validateUpsertStudents: expected "full_name_phonetic" should be null, but actual:%v`, dbUser.FullNamePhonetic().String())
	case reqStudent.FirstNamePhonetic().String() != dbUser.FirstNamePhonetic().String():
		return ctx, fmt.Errorf(`validateUpsertStudents: expected upserted "first_name_phonetic": %v but actual is %v`, reqStudent.FirstNamePhonetic().String(), dbUser.FirstNamePhonetic().String())
	case reqStudent.LastNamePhonetic().String() != dbUser.LastNamePhonetic().String():
		return ctx, fmt.Errorf(`validateUpsertStudents: expected upserted "last_name_phonetic": %v but actual is %v`, reqStudent.LastNamePhonetic().String(), dbUser.LastNamePhonetic().String())
	case fullNamePhonetic.String() != dbUser.FullNamePhonetic().String():
		return ctx, fmt.Errorf(`validateUpsertStudents: expected upserted "full_name_phonetic": %v but actual is %v`, fullNamePhonetic.String(), dbUser.FullNamePhonetic().String())
	case reqStudent.Gender().String() != dbUser.Gender().String():
		return ctx, fmt.Errorf(`validateUpsertStudents: expected upserted "gender": %v but actual is %v`, reqStudent.Gender().String(), dbUser.Gender().String())
	case reqStudent.Remarks().String() != dbUser.Remarks().String():
		return ctx, fmt.Errorf(`validateUpsertStudents: expected upserted "student_note": %v but actual is %v`, reqStudent.Remarks().String(), dbUser.Remarks().String())
	case reqStudent.Email().String() != dbUser.Email().String():
		return ctx, fmt.Errorf(`validateUpsertStudents: expected upserted "email": %v but actual is %v`, reqStudent.Email().String(), dbUser.Email().String())
	case reqStudent.Birthday().Date().Format(constant.DateLayout) != dbUser.Birthday().Date().Format(constant.DateLayout):
		return ctx, fmt.Errorf(`validateUpsertStudents: expected upserted "birthday": %v but actual is %v`, reqStudent.Birthday().Date().Format(constant.DateLayout), dbUser.Birthday().Date().Format(constant.DateLayout))
	case dbUser.UserRole().String() != string(constant.UserRoleStudent):
		return ctx, fmt.Errorf(`validateUpsertStudents: expected upserted "user_role": %s but actual is %s`, string(constant.UserRoleStudent), dbUser.UserRole().String())
	}
	isEnableUsername, err := isFeatureToggleEnabled(ctx, s.UnleashSuite, pkg_unleash.FeatureToggleUserNameStudentParent)
	if err != nil {
		return ctx, errors.Wrap(err, fmt.Sprintf("Get feature toggle error(%s)", pkg_unleash.FeatureToggleUserNameStudentParent))
	}
	if err := assertUpsertUsername(isEnableUsername, assertUsername{
		requestUsername:    reqStudent.UserName().String(),
		requestEmail:       reqStudent.Email().String(),
		databaseUsername:   dbUser.UserName().String(),
		requestLoginEmail:  userID + constant.LoginEmailPostfix,
		databaseLoginEmail: dbUser.LoginEmail().String(),
	}); err != nil {
		return ctx, err
	}

	orgID := OrgIDFromCtx(ctx)
	if !reqStudent.Password().IsEmpty() {
		email := reqStudent.Email().String()
		if isEnableUsername {
			email = userID + constant.LoginEmailPostfix
		}
		if err := s.loginIdentityPlatform(ctx, golibs_auth.LocalTenants[orgID], email, reqStudent.Password().String()); err != nil {
			return ctx, errors.Wrap(err, "loginIdentityPlatform")
		}
	}

	if field.IsPresent(reqStudent.Birthday()) {
		if reqStudent.Birthday().Date().Format(constant.DateLayout) != dbUser.Birthday().Date().Format(constant.DateLayout) {
			return ctx, fmt.Errorf(`validateUpsertStudents: expected upserted "birthday": %v but actual is %v`, reqStudent.Birthday().Date(), dbUser.Birthday().Date())
		}
	} else {
		if field.IsPresent(dbUser.Birthday()) {
			return ctx, fmt.Errorf(`validateUpsertStudents: expected upserted "birthday": not present but actual is presented`)
		}
	}

	if err := s.validateUsersHasUserGroupWithRole(ctx, []string{dbUser.UserID().String()}, fmt.Sprint(orgID), constant.RoleStudent); err != nil {
		return ctx, fmt.Errorf("s.userHasUserGroupWithRole: %v, user_id: %s", err, dbUser.UserID().String())
	}

	grades, err := (&repository.DomainGradeRepo{}).GetByIDs(ctx, s.BobPostgresDB, []string{reqStudent.GradeID().String()})
	if err != nil {
		return ctx, errors.Wrap(err, "(&repository.DomainUserRepo{}).GetByEmails")
	}

	if len(grades) == 0 {
		return ctx, fmt.Errorf("grade %s is not found with user_id %s", reqStudent.GradeID().String(), userID)
	}
	return ctx, nil
}

func (s *suite) verifyUserAddress(ctx context.Context, reqUserAddress entity.DomainUserAddress, dbUser entity.User) (context.Context, error) {
	if reqUserAddress != nil {
		pbUserAddress := []*upb.UserAddress{
			{
				PostalCode:   reqUserAddress.PostalCode().String(),
				City:         reqUserAddress.City().String(),
				FirstStreet:  reqUserAddress.FirstStreet().String(),
				SecondStreet: reqUserAddress.SecondStreet().String(),
				Prefecture:   reqUserAddress.PrefectureID().String(),
			},
		}
		err := validateHomeAddressesInDB(ctx, s.BobPostgresDBTrace, pbUserAddress, dbUser.UserID().String(), fmt.Sprint(OrgIDFromCtx(ctx)))
		if err != nil {
			return ctx, errors.Wrap(err, "validateHomeAddressesInDB")
		}
	} else {
		userAddresses, err := (&repository.UserAddressRepo{}).GetByUserID(ctx, s.BobPostgresDBTrace, database.Text(dbUser.UserID().String()))
		if err != nil {
			return ctx, fmt.Errorf("userAddressRepo.GetByUserID: %v", err)
		}

		if len(userAddresses) != 0 {
			return ctx, fmt.Errorf("expect to address get removed")
		}
	}
	return ctx, nil
}

func (s *suite) verifyUserPhoneNumbers(ctx context.Context, reqStudent aggregate.DomainStudent, dbUser entity.User) (context.Context, error) {
	studentRepo := repository.StudentRepo{}
	studentInDB, err := studentRepo.Find(ctx, s.BobDBTrace, database.Text(dbUser.UserID().String()))
	if err != nil {
		return ctx, errors.Wrap(err, "studentRepo.Find")
	}

	if studentInDB.ContactPreference.String != reqStudent.ContactPreference().String() {
		return ctx, fmt.Errorf(`ContactPreference: %v but actual is %v`, reqStudent.ContactPreference().String(), studentInDB.ContactPreference.String)
	}

	if len(reqStudent.UserPhoneNumbers) != 0 {
		stringPhoneNumbers := []string{}
		for _, phoneNumber := range reqStudent.UserPhoneNumbers {
			if !phoneNumber.PhoneNumber().IsEmpty() {
				stringPhoneNumbers = append(stringPhoneNumbers, phoneNumber.PhoneNumber().String())
			}
		}
		if len(stringPhoneNumbers) != 0 {
			userPhoneNumberRepoRepo := repository.UserPhoneNumberRepo{}
			userPhoneNumbers, err := userPhoneNumberRepoRepo.FindByUserID(ctx, s.BobDBTrace, database.Text(dbUser.UserID().String()))
			if err != nil {
				return ctx, errors.Wrap(err, "userPhoneNumberRepoRepo.FindByUserID")
			}

			if len(userPhoneNumbers) == 0 {
				return ctx, fmt.Errorf(`len of userPhoneNumbers actual is %v, UserID: %s`, len(userPhoneNumbers), dbUser.UserID().String())
			}
			mapPhoneNumberTypeAndPhoneNumber := make(map[string]string, len(userPhoneNumbers))

			for _, userPhoneNumber := range userPhoneNumbers {
				mapPhoneNumberTypeAndPhoneNumber[userPhoneNumber.PhoneNumberType.String] = userPhoneNumber.PhoneNumber.String
			}

			for _, phoneNumber := range reqStudent.UserPhoneNumbers {
				if mapPhoneNumberTypeAndPhoneNumber[phoneNumber.Type().String()] != phoneNumber.PhoneNumber().String() {
					return ctx, fmt.Errorf(`expect phone number: %v but actual is %v`, mapPhoneNumberTypeAndPhoneNumber[phoneNumber.Type().String()], phoneNumber.PhoneNumber().String())
				}
			}
		}
	} else {
		// check this user won't have any record of phone numbers
		if err := validateStudentDontHavePhoneNumber(ctx, s.BobPostgresDBTrace, dbUser.UserID().String()); err != nil {
			return ctx, errors.Wrap(err, "validateStudentDontHavePhoneNumber")
		}
	}
	return ctx, nil
}

func (s *suite) verifySchoolHistories(ctx context.Context, reqStudent aggregate.DomainStudent, dbStudent entity.User) (context.Context, error) {
	if len(reqStudent.SchoolHistories) == 0 {
		schoolHistories, err := (&repository.SchoolHistoryRepo{}).GetByStudentID(ctx, s.BobPostgresDBTrace, database.Text(dbStudent.UserID().String()))
		if err != nil {
			return ctx, errors.Wrap(err, "(&repository.SchoolHistoryRepo{}).GetByStudentID")
		}
		if len(schoolHistories) != 0 {
			return ctx, fmt.Errorf("expect to school histories get removed")
		}
	} else {
		schoolRepo := repository.DomainSchoolRepo{}
		schools, err := schoolRepo.GetByIDsAndGradeID(ctx, s.BobDBTrace, reqStudent.SchoolHistories.SchoolIDs(), reqStudent.GradeID().String())
		if err != nil {
			return ctx, errors.Wrap(err, "schoolRepo.GetByIDsAndGradeID")
		}
		pbSchoolHistories := make([]*upb.SchoolHistory, 0, len(reqStudent.SchoolHistories))
		for _, schoolHistory := range reqStudent.SchoolHistories {
			startDate, _ := time.Parse(constant.DateLayout, schoolHistory.StartDate().Time().Format(constant.DateLayout))
			endDate, _ := time.Parse(constant.DateLayout, schoolHistory.EndDate().Time().Format(constant.DateLayout))

			pbSchoolHistories = append(pbSchoolHistories, &upb.SchoolHistory{
				SchoolId:       schoolHistory.SchoolID().String(),
				SchoolCourseId: schoolHistory.SchoolCourseID().String(),
				StartDate:      timestamppb.New(startDate),
				EndDate:        timestamppb.New(endDate),
			})

		}
		if err := validateSchoolHistoriesInDB(ctx, s.BobDBTrace, pbSchoolHistories, dbStudent.UserID().String(), fmt.Sprint(OrgIDFromCtx(ctx)), schools); err != nil {
			return ctx, fmt.Errorf("validateSchoolHistoriesInDB: %v", err)
		}
	}
	return ctx, nil
}

// Only create unsuccessfully. TODO: should change to username
func (s *suite) verifyUsersNotInBD(ctx context.Context, emails []string) (context.Context, error) {
	userRepo := &repository.UserRepo{}
	for _, email := range emails {
		users, err := userRepo.GetByEmail(ctx, s.BobDBTrace, database.TextArray([]string{email}))
		if err != nil {
			return ctx, fmt.Errorf("userRepo.GetByEmail err: %v", err)
		}
		if len(users) > 0 {
			return ctx, fmt.Errorf("can find user with email: %s", email)
		}
	}
	return ctx, nil
}

func (s *suite) verifyLocationInNatsEvent(ctx context.Context, userIDs []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	timer := time.NewTimer(time.Second * 10)
	defer timer.Stop()

	type result struct {
		userID string
		err    error
	}
	checkResults := make(chan result, len(userIDs))

	for _, userID := range userIDs {
		go func(userID string) {
			var locationIDsInMsg []string

			msg := <-stepState.FoundChanForJetStream

			switch msg := msg.(type) {
			case *upb.EvtUser_UpdateStudent_:
				locationIDsInMsg = msg.UpdateStudent.LocationIds
			case *upb.EvtUser_CreateStudent_:
				locationIDsInMsg = msg.CreateStudent.LocationIds
			}

			userAccessPathsInDB, err := new(repository.DomainUserAccessPathRepo).GetByUserID(ctx, s.BobDBTrace, field.NewString(userID))
			if err != nil {
				checkResults <- result{
					userID: userID,
					err:    errors.Wrap(err, "GetByUserID()"),
				}
				return
			}
			if err := utils.CompareStringsRegardlessOrder(userAccessPathsInDB.LocationIDs(), locationIDsInMsg); err != nil {
				checkResults <- result{
					userID: userID,
					err:    errors.Wrap(err, fmt.Sprintf("CompareStringsRegardlessOrder: user_id %s", userID)),
				}
				return
			}
			checkResults <- result{userID: userID}
		}(userID)
	}

	checkedUserIds := make([]string, 0, len(userIDs))
	for {
		select {
		case checkResult := <-checkResults:
			if err := checkResult.err; err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			checkedUserIds = append(checkedUserIds, checkResult.userID)

			if len(checkedUserIds) == len(userIDs) {
				return StepStateToContext(ctx, stepState), nil
			}
		case <-timer.C:
			return ctx, fmt.Errorf("timeout waiting for event to be published")
		}
	}
}

func (s *suite) createSubscriptionForCreatedStudentByGRPC(ctx context.Context, req *upb.UpsertStudentRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 2)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}
	handleUpsertUser := func(ctx context.Context, data []byte) (bool, error) {
		evtUser := &upb.EvtUser{}
		if err := proto.Unmarshal(data, evtUser); err != nil {
			return false, err
		}
		switch msg := evtUser.Message.(type) {
		case *upb.EvtUser_CreateStudent_:
			for _, student := range req.GetStudentProfiles() {
				if student.FirstName == msg.CreateStudent.StudentFirstName && student.LastName == msg.CreateStudent.StudentLastName {
					locations := make([]string, 0, len(student.EnrollmentStatusHistories))
					for _, enrollmentStatusHistory := range student.EnrollmentStatusHistories {
						locations = append(locations, enrollmentStatusHistory.LocationId)
					}
					if len(locations) != len(msg.CreateStudent.LocationIds) && len(student.LocationIds) != len(msg.CreateStudent.LocationIds) {
						return true, nil
					}
					if len(student.UserAddresses) != 0 && msg.CreateStudent.UserAddress != nil {
						for _, userAddress := range student.UserAddresses {
							if userAddress != nil && msg.CreateStudent.UserAddress != nil {
								if userAddress.PostalCode != msg.CreateStudent.UserAddress.PostalCode ||
									userAddress.City != msg.CreateStudent.UserAddress.City ||
									userAddress.FirstStreet != msg.CreateStudent.UserAddress.FirstStreet ||
									userAddress.SecondStreet != msg.CreateStudent.UserAddress.SecondStreet {
									return true, nil
								}
							}
						}
					}

					evtUserCreateStudent := evtUser.GetCreateStudent()
					if err := assertEnrollmentStatusHistoryEvent(ctx, s.BobDBTrace, evtUserCreateStudent.StudentId, evtUserCreateStudent.EnrollmentStatusHistories); err != nil {
						return true, err
					}

					stepState.FoundChanForJetStream <- evtUser.Message
					return false, nil
				}
			}
		case *upb.EvtUser_UpdateStudent_:
			for _, student := range req.StudentProfiles {
				if student.FirstName == msg.UpdateStudent.StudentFirstName && student.LastName == msg.UpdateStudent.StudentLastName {
					locations := make([]string, 0, len(student.EnrollmentStatusHistories))
					for _, enrollmentStatusHistory := range student.EnrollmentStatusHistories {
						locations = append(locations, enrollmentStatusHistory.LocationId)
					}
					if len(locations) != len(msg.UpdateStudent.LocationIds) && len(student.LocationIds) != len(msg.UpdateStudent.LocationIds) {
						return true, nil
					}
					if len(student.UserAddresses) != 0 && msg.UpdateStudent.UserAddress != nil {
						for _, userAddress := range student.UserAddresses {
							if userAddress != nil && msg.UpdateStudent.UserAddress != nil {
								if userAddress.PostalCode != msg.UpdateStudent.UserAddress.PostalCode ||
									userAddress.City != msg.UpdateStudent.UserAddress.City ||
									userAddress.FirstStreet != msg.UpdateStudent.UserAddress.FirstStreet ||
									userAddress.SecondStreet != msg.UpdateStudent.UserAddress.SecondStreet {
									return true, nil
								}
							}
						}
					}

					evtUserUpdateStudent := evtUser.GetUpdateStudent()
					if err := assertEnrollmentStatusHistoryEvent(ctx, s.BobDBTrace, evtUserUpdateStudent.StudentId, evtUserUpdateStudent.EnrollmentStatusHistories); err != nil {
						return true, err
					}

					stepState.FoundChanForJetStream <- evtUser.Message
					return false, nil
				}
			}
		}
		return false, nil
	}

	subs, err := s.JSM.Subscribe(constants.SubjectUserCreated, opts, handleUpsertUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("createSubscriptionForCreatedStudentByGRPC: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)

	subs, err = s.JSM.Subscribe(constants.SubjectUserUpdated, opts, handleUpsertUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("createSubscriptionForCreatedStudentByGRPC: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createSubscriptionForCreatedStudentByOpenAPI(ctx context.Context, req *http_port.UpsertStudentsRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 2)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}
	handleUpsertUser := func(ctx context.Context, data []byte) (bool, error) {
		evtUser := &upb.EvtUser{}
		if err := proto.Unmarshal(data, evtUser); err != nil {
			return false, err
		}
		switch msg := evtUser.Message.(type) {
		case *upb.EvtUser_CreateStudent_:
			for _, student := range req.Students {
				if student.FirstName.String() == msg.CreateStudent.StudentFirstName && student.LastName.String() == msg.CreateStudent.StudentLastName {
					locations := make([]string, 0, len(student.EnrollmentStatusHistories))
					for _, enrollmentStatusHistory := range student.EnrollmentStatusHistories {
						locations = append(locations, enrollmentStatusHistory.Location.String())
					}
					if len(locations) != len(msg.CreateStudent.LocationIds) && len(student.Locations) != len(msg.CreateStudent.LocationIds) {
						return true, nil
					}
					if student.Address != nil && msg.CreateStudent.UserAddress != nil {
						if student.Address.PostalCode.String() != msg.CreateStudent.UserAddress.PostalCode ||
							student.Address.City.String() != msg.CreateStudent.UserAddress.City ||
							student.Address.FirstStreet.String() != msg.CreateStudent.UserAddress.FirstStreet ||
							student.Address.SecondStreet.String() != msg.CreateStudent.UserAddress.SecondStreet {
							return true, nil
						}
					}

					evtUserCreateStudent := evtUser.GetCreateStudent()
					if err := assertEnrollmentStatusHistoryEvent(ctx, s.BobDBTrace, evtUserCreateStudent.StudentId, evtUserCreateStudent.EnrollmentStatusHistories); err != nil {
						return true, err
					}

					stepState.FoundChanForJetStream <- evtUser.Message
					return false, nil
				}
			}
		case *upb.EvtUser_UpdateStudent_:
			for _, student := range req.Students {
				if student.FirstName.String() == msg.UpdateStudent.StudentFirstName && student.LastName.String() == msg.UpdateStudent.StudentLastName {
					locations := make([]string, 0, len(student.EnrollmentStatusHistories))
					for _, enrollmentStatusHistory := range student.EnrollmentStatusHistories {
						locations = append(locations, enrollmentStatusHistory.Location.String())
					}
					if len(locations) != len(msg.UpdateStudent.LocationIds) && len(student.Locations) != len(msg.UpdateStudent.LocationIds) {
						return true, nil
					}
					if student.Address != nil && msg.UpdateStudent.UserAddress != nil {
						if student.Address.PostalCode.String() != msg.UpdateStudent.UserAddress.PostalCode ||
							student.Address.City.String() != msg.UpdateStudent.UserAddress.City ||
							student.Address.FirstStreet.String() != msg.UpdateStudent.UserAddress.FirstStreet ||
							student.Address.SecondStreet.String() != msg.UpdateStudent.UserAddress.SecondStreet {
							return true, nil
						}
					}

					evtUserUpdateStudent := evtUser.GetUpdateStudent()
					if err := assertEnrollmentStatusHistoryEvent(ctx, s.BobDBTrace, evtUserUpdateStudent.StudentId, evtUserUpdateStudent.EnrollmentStatusHistories); err != nil {
						return true, err
					}

					stepState.FoundChanForJetStream <- evtUser.Message
					return false, nil
				}
			}
		}
		return false, nil
	}

	subs, err := s.JSM.Subscribe(constants.SubjectUserCreated, opts, handleUpsertUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("createSubscriptionForCreatedStudentByOpenAPI: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)

	subs, err = s.JSM.Subscribe(constants.SubjectUserUpdated, opts, handleUpsertUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("createSubscriptionForCreatedStudentByOpenAPI: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createSubscriptionForCreatedStudentByImport(ctx context.Context, req *upb.ImportStudentRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 2)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}
	handleUpsertUser := func(ctx context.Context, data []byte) (bool, error) {
		evtUser := &upb.EvtUser{}
		if err := proto.Unmarshal(data, evtUser); err != nil {
			return false, err
		}
		studentCSVs := []importstudent.StudentCSV{}
		err := gocsv.UnmarshalBytes(req.Payload, &studentCSVs)
		if err != nil {
			return true, errors.Wrap(err, "gocsv.UnmarshalBytes")
		}
		switch msg := evtUser.Message.(type) {
		case *upb.EvtUser_CreateStudent_:
			for _, student := range studentCSVs {
				if student.FirstNameAttr.String() == msg.CreateStudent.StudentFirstName && student.LastNameAttr.String() == msg.CreateStudent.StudentLastName {
					locations := []string{}
					if !student.LocationAttr.IsEmpty() {
						locations = strings.Split(student.LocationAttr.String(), ";")
					}
					if len(locations) != len(msg.CreateStudent.LocationIds) {
						return true, nil
					}
					if student.PostalCodeAttr.String() != msg.CreateStudent.UserAddress.PostalCode ||
						student.CityAttr.String() != msg.CreateStudent.UserAddress.City ||
						student.FirstStreetAttr.String() != msg.CreateStudent.UserAddress.FirstStreet ||
						student.SecondStreetAttr.String() != msg.CreateStudent.UserAddress.SecondStreet {
						return true, nil
					}
					stepState.FoundChanForJetStream <- evtUser.Message
					return false, nil
				}
			}
		case *upb.EvtUser_UpdateStudent_:
			for _, student := range studentCSVs {
				if student.FirstNameAttr.String() == msg.UpdateStudent.StudentFirstName && student.LastNameAttr.String() == msg.UpdateStudent.StudentLastName {
					locations := []string{}
					if !student.LocationAttr.IsEmpty() {
						locations = strings.Split(student.LocationAttr.String(), ";")
					}
					if len(locations) != len(msg.UpdateStudent.LocationIds) {
						return true, nil
					}
					if student.PostalCodeAttr.String() != msg.UpdateStudent.UserAddress.PostalCode ||
						student.CityAttr.String() != msg.UpdateStudent.UserAddress.City ||
						student.FirstStreetAttr.String() != msg.UpdateStudent.UserAddress.FirstStreet ||
						student.SecondStreetAttr.String() != msg.UpdateStudent.UserAddress.SecondStreet {
						return true, nil
					}
					stepState.FoundChanForJetStream <- evtUser.Message
					return false, nil
				}
			}
		}
		return false, nil
	}

	subs, err := s.JSM.Subscribe(constants.SubjectUserCreated, opts, handleUpsertUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("createSubscriptionForCreatedStudentByImport: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)

	subs, err = s.JSM.Subscribe(constants.SubjectUserUpdated, opts, handleUpsertUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("createSubscriptionForCreatedStudentByImport: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func assertEnrollmentStatusHistoryEvent(ctx context.Context, db database.QueryExecer, studentID string, enrollmentStatusHistoriesEvt []*upb.EnrollmentStatusHistory) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	return TryUntilSuccess(ctx, defaultInterval*time.Millisecond, func(ctx context.Context) (bool, error) {
		enrollmentStatusHistoriesDB, err := getAllEnrollmentStatusHistoriesOfStudent(ctx, db, studentID)
		if err != nil {
			return false, err
		}

		for _, enrollmentStatusHistoryDB := range enrollmentStatusHistoriesDB {
			if !enrollmentStatusHistoryDB.DeletedAtAttr.IsZero() {
				continue
			}

			found := 0
			errStr := ""
			for _, enrollmentStatusHistoryEvt := range enrollmentStatusHistoriesEvt {
				locationIDDB := enrollmentStatusHistoryDB.LocationID().String()
				locationIDEvt := enrollmentStatusHistoryEvt.LocationId
				if locationIDDB != locationIDEvt {
					errStr = fmt.Sprintf("failed to match locationID, expected: %s, got: %s", locationIDDB, locationIDEvt)
					continue
				}

				enrollmentStatusDB := enrollmentStatusHistoryDB.EnrollmentStatus().String()
				enrollmentStatusEvt := enrollmentStatusHistoryEvt.EnrollmentStatus.String()
				if enrollmentStatusDB != enrollmentStatusEvt {
					errStr = fmt.Sprintf("failed to match enrollmentStatus, expected: %s, got: %s", enrollmentStatusDB, enrollmentStatusEvt)
					continue
				}

				startDateInDB := enrollmentStatusHistoryDB.StartDate().Time()
				startDateInEvt := enrollmentStatusHistoryEvt.StartDate.AsTime()
				if startDateInDB != startDateInEvt {
					errStr = fmt.Sprintf("failed to match startDate, expected: %s, got: %s", startDateInDB, startDateInEvt)
					continue
				}

				endDateInDB := enrollmentStatusHistoryDB.EndDate().Time()
				endDateInEvt := enrollmentStatusHistoryEvt.EndDate.AsTime()
				if field.IsPresent(enrollmentStatusHistoryDB.EndDate()) {
					if endDateInDB.Format(time.ANSIC) != endDateInEvt.Format(time.ANSIC) {
						errStr = fmt.Sprintf("failed to match endDate, expected: %s, got: %s", endDateInDB.Format(time.ANSIC), endDateInEvt.Format(time.ANSIC))
						continue
					}
				} else {
					if endDateInDB.Nanosecond() != endDateInEvt.Nanosecond() {
						errStr = "failed to match endDate, expected: none, got: existed"
						continue
					}
				}

				found++
			}

			if found == 0 {
				return true, fmt.Errorf("assertEnrollmentStatusHistoryEvent: failed to match enrollmentStatusHistory: %s of user %s", errStr, studentID)
			}
		}
		return false, nil
	})
}

func (s *suite) verifyStudentTags(ctx context.Context, reqTags entity.DomainTags, dbUser entity.User) (context.Context, error) {
	if len(reqTags.TagIDs()) != 0 {
		err := s.validateUserTags(ctx, dbUser.UserID().String(), reqTags.TagIDs())
		if err != nil {
			return ctx, errors.Wrapf(err, "s.validateUserTags, UserID: %s, tag_ids: %v", dbUser.UserID().String(), reqTags.TagIDs())
		}
	}
	return ctx, nil
}

func (s *suite) validateUserTags(ctx context.Context, userID string, tagIDs []string) error {
	taggedUserRepo := repository.DomainTaggedUserRepo{}

	taggedUsers, err := taggedUserRepo.GetByUserIDs(ctx, s.BobDBTrace, []string{userID})

	if err != nil {
		return errors.Wrap(err, "taggedUserRepo.GetByUserIDs")
	}

	createUserTagIDs := map[string]struct{}{}
	queriedTagIDs := []string{}

	for _, taggedUser := range taggedUsers {
		createUserTagIDs[taggedUser.TagID().String()] = struct{}{}
		queriedTagIDs = append(queriedTagIDs, taggedUser.TagID().String())
	}

	if len(queriedTagIDs) != len(tagIDs) {
		return fmt.Errorf("number of tags in db is invalid")
	}
	return nil
}

type assertUsername struct {
	requestUsername    string
	requestEmail       string
	databaseUsername   string
	requestLoginEmail  string
	databaseLoginEmail string
}

func assertUpsertUsername(isEnableUsername bool, assertParams assertUsername) error {
	if isEnableUsername {
		if !strings.EqualFold(assertParams.requestUsername, assertParams.databaseUsername) {
			return fmt.Errorf(`assertUpsertUsername: expected upserted "username": %v but actual is %v`, assertParams.requestUsername, assertParams.databaseUsername)
		}
		if !strings.EqualFold(assertParams.requestLoginEmail, assertParams.databaseLoginEmail) {
			return fmt.Errorf(`assertUpsertUsername: expected upserted "login_email": %v but actual is %v`, assertParams.requestLoginEmail, assertParams.databaseLoginEmail)
		}
	} else {
		// for logic in username CRU, if feature username is disabled, we will use email as username
		if !strings.EqualFold(assertParams.requestEmail, assertParams.databaseUsername) {
			return fmt.Errorf(`assertUpsertUsername: expected upserted "username" same with email: %v but actual is %v`, assertParams.requestEmail, assertParams.databaseUsername)
		}
		if !strings.EqualFold(assertParams.requestEmail, assertParams.databaseLoginEmail) {
			return fmt.Errorf(`assertUpsertUsername: expected upserted "login_email" same with email: %v but actual is %v`, assertParams.requestEmail, assertParams.databaseLoginEmail)
		}
	}
	return nil
}
