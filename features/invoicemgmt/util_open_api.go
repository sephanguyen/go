package invoicemgmt

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/golibs/crypt"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	http_port "github.com/manabie-com/backend/internal/invoicemgmt/services/http"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type BillingAddressInfo struct {
	StudentID    string
	PayerName    string
	PostalCode   string
	PrefectureID string
	City         string
	Street1      string
	Street2      string
}

type UpsertStudentBankRequestInfo struct {
	StudentBankRequestInfo StudentBankRequestInfoProfile `json:"student_bank_info"`
}

type StudentBankRequestInfoProfile struct {
	ExternalUserID    pgtype.Text `json:"external_user_id"`
	BankCode          pgtype.Text `json:"bank_code"`
	BankBranchCode    pgtype.Text `json:"bank_branch_code"`
	BankAccountNumber pgtype.Text `json:"bank_account_number"`
	BankAccountHolder pgtype.Text `json:"bank_account_holder"`
	BankAccountType   pgtype.Int4 `json:"bank_account_type"`
	IsVerified        pgtype.Bool `json:"is_verified"`
}

const (
	InvoicemgmtConfigKeyPrefix = "invoice.invoicemgmt."
	AESKey                     = "yYdV82bsdXO%Cl2Uq5F^^19GUh8%^W3j"
	AESIV                      = "^30F6l#gm0C!@oD7"
)

func validateBillingAddressInfoFromStudentEvent(userAddress *upb.UserAddress, billingAddressInfo *BillingAddressInfo) error {
	if billingAddressInfo.PrefectureID != userAddress.Prefecture {
		return fmt.Errorf("error expecting prefecture: %v but got: %v", userAddress.Prefecture, billingAddressInfo.PrefectureID)
	}

	if billingAddressInfo.City != userAddress.City {
		return fmt.Errorf("error expecting city: %v but got: %v", userAddress.City, billingAddressInfo.City)
	}

	if billingAddressInfo.Street1 != userAddress.FirstStreet {
		return fmt.Errorf("error expecting street 1: %v but got: %v", userAddress.FirstStreet, billingAddressInfo.Street1)
	}

	if billingAddressInfo.Street2 != userAddress.SecondStreet {
		return fmt.Errorf("error expecting street 2: %v but got: %v", userAddress.SecondStreet, billingAddressInfo.Street2)
	}
	return nil
}

func (s *suite) invoicemgmtInternalConfigIs(ctx context.Context, configKey, toggle string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := `UPDATE internal_configuration_value SET config_value = $1 
                WHERE config_key = $2  
				AND resource_path = $3 
				AND deleted_at is NULL;`

	_, err := s.MasterMgmtDBTrace.Exec(ctx, stmt, database.Text(toggle), database.Text(InvoicemgmtConfigKeyPrefix+configKey), database.Text(stepState.ResourcePath))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) setupAPIKey(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)

	_, err := s.aSignedInStaff(ctx, []string{constant.RoleOpenAPI})
	if err != nil {
		return errors.Wrap(err, "s.aSignedInStaff")
	}

	_, err = s.systemRunJobToGenerateAPIKeyWithOrganization(ctx)
	if err != nil {
		return errors.Wrap(err, "s.systemRunJobToGenerateAPIKeyWithOrganization")
	}

	var publicKey, privateKey string

	// adding try do on selecting records as sometimes it becomes flaky and has error no rows in result set
	query := `SELECT public_key, private_key FROM api_keypair WHERE user_id = $1 AND resource_path = $2`
	if err := try.Do(func(attempt int) (bool, error) {
		err = database.Select(ctx, s.AuthPostgresDBTrace, query, stepState.CurrentUserID, stepState.ResourcePath).ScanFields(&publicKey, &privateKey)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, errors.Wrap(err, "error on selecting api keypair record")
		}

		if err == nil && publicKey != "" && privateKey != "" {
			return false, nil
		}

		time.Sleep(invoiceConst.ReselectSleepDuration)
		return attempt < 10, fmt.Errorf("error on selecting api keypair record on attempt: %v", attempt)
	}); err != nil {
		return err
	}
	data, err := json.Marshal(s.Request)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}

	if stepState.BankOpenAPIPrivateKey != "" {
		privateKey = stepState.BankOpenAPIPrivateKey
	}

	if stepState.BankOpenAPIPublicKey != "" {
		publicKey = stepState.BankOpenAPIPublicKey
	}

	privateKeyByte, _ := crypt.AESDecryptBase64(privateKey, []byte(AESKey), []byte(AESIV))
	sig := hmac.New(sha256.New, privateKeyByte)
	if _, err := sig.Write(data); err != nil {
		return errors.Wrap(err, "sig.Write")
	}

	s.ManabiePublicKey = publicKey
	s.ManabieSignature = hex.EncodeToString(sig.Sum(nil))

	return nil
}

func (s *suite) aSignedInStaff(ctx context.Context, roles []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.AuthToken == "" {
		token, err := s.generateExchangeToken(stepState.CurrentUserID, constant.UserGroupSchoolAdmin, int64(stepState.CurrentSchoolID))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.AuthToken = token
	}

	createUserGroupResp, err := s.createUserGroupWithRoleNames(ctx, roles)
	if err != nil {
		return nil, err
	}

	if err := assignUserGroupToUser(ctx, s.BobDBTrace, stepState.CurrentUserID, stepState.ResourcePath, []string{createUserGroupResp.UserGroupId}); err != nil {
		return nil, err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createUserGroupWithRoleNames(ctx context.Context, roleNames []string) (*upb.CreateUserGroupResponse, error) {
	stepState := StepStateFromContext(ctx)
	req := &upb.CreateUserGroupRequest{
		UserGroupName: fmt.Sprintf("user-group_%s", idutil.ULIDNow()),
	}

	stmt := "SELECT role_id FROM role WHERE deleted_at IS NULL AND role_name = ANY($1) and resource_path = $2 LIMIT $3"
	rows, err := s.BobDBTrace.Query(ctx, stmt, roleNames, stepState.ResourcePath, len(roleNames))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roleIDs []string
	for rows.Next() {
		roleID := ""
		if err := rows.Scan(&roleID); err != nil {
			return nil, fmt.Errorf("rows.Err: %w", err)
		}
		roleIDs = append(roleIDs, roleID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	for _, roleID := range roleIDs {
		req.RoleWithLocations = append(
			req.RoleWithLocations,
			&upb.RoleWithLocations{
				RoleId:      roleID,
				LocationIds: []string{stepState.LocationID},
			},
		)
	}

	resp, err := upb.NewUserGroupMgmtServiceClient(s.UserMgmtConn).CreateUserGroup(contextWithToken(ctx), req)
	if err != nil {
		return nil, fmt.Errorf("createUserGroupWithRoleNames: %w", err)
	}

	return resp, nil
}

type userOption func(u *entity.LegacyUser)

func withID(id string) userOption {
	return func(u *entity.LegacyUser) {
		_ = u.ID.Set(id)
	}
}

func withRole(group string) userOption {
	return func(u *entity.LegacyUser) {
		_ = u.Group.Set(group)
	}
}

func assignUserGroupToUser(ctx context.Context, dbBob database.QueryExecer, userID, resourcePath string, userGroupIDs []string) error {
	userGroupMembers := make([]*entity.UserGroupMember, 0)
	for _, userGroupID := range userGroupIDs {
		userGroupMem := &entity.UserGroupMember{}
		database.AllNullEntity(userGroupMem)
		if err := multierr.Combine(
			userGroupMem.UserID.Set(userID),
			userGroupMem.UserGroupID.Set(userGroupID),
			userGroupMem.ResourcePath.Set(resourcePath),
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

func (s *suite) systemRunJobToGenerateAPIKeyWithOrganization(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.OrganizationID = fmt.Sprint(constants.ManabieSchool)

	zLogger := logger.NewZapLogger("warn", s.Cfg.Common.Environment == "local")

	err := usermgmt.RunGenerateAPIKeypair(ctx, &configurations.Config{
		OpenAPI: configurations.OpenAPI{
			AESIV:  AESIV,
			AESKey: AESKey,
		},
		Common:     s.Cfg.Common,
		PostgresV2: s.Cfg.PostgresV2,
	}, s.BobPostgresDB, zLogger, stepState.CurrentUserID, stepState.ResourcePath)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "systemRunJobToGenerateAPIKeyWithOrganization failed")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) makeHTTPRequest(method, url string, bodyRequest []byte) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyRequest))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("manabie-public-key", s.ManabiePublicKey)
	req.Header.Set("manabie-signature", s.ManabieSignature)

	if err != nil {
		return nil, err
	}

	client := http.Client{Timeout: time.Duration(30) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	respStruct := http_port.Response{}
	err = json.Unmarshal(body, &respStruct)
	if err != nil {
		return nil, err
	}
	s.Response = respStruct

	return body, nil
}
