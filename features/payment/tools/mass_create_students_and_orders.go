/*
 *
 * HOW TO USE
 *   Step 1:
 *     Use FE script located at school-portal-admin/src/squads/payment/csv/
 *     and ImportAllForTest API to import master data
 *
 *   Step 2:
 *     Update correct values for target env in "GLOBAL VARS" section
 *     using the imported master data from Step 1
 *
 *   Step 3:
 *     Run this script:
 *       go run mass_create_students_and_orders.go
 *
 */

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	ipb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	// port forward these if test with local environment
	// kubectl -n backend port-forward $(kubectl get pods -l app.kubernetes.io/name=bob -n backend -o=name) 5050:5050
	BOB_LOCAL_API_URL        = "127.0.0.1:5050"
	ORDER_LOCAL_API_URL      = "127.0.0.1:6250"
	INVOICE_LOCAL_API_URL    = "127.0.0.1:6650"
	USERMGMT_LOCAL_API_URL   = "127.0.0.1:6150"
	MASTERMGMT_LOCAL_API_URL = "127.0.0.1:6450"

	STAG_API_URL       = "web-api.staging-green.manabie.io:443"
	UAT_API_URL        = "api.uat.manabie.io:443"
	PROD_TOKYO_API_URL = "https://web-api.prod.tokyo.manabie.io:31400"

	IDENTITY_TOOLKIT_URL = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key="
	LOCAL_API_KEY        = "AIzaSyAlE26SQ0OMGjmr4IiF9D6CJkK0eRvV6HA"
	STAG_API_KEY         = "AIzaSyA7h5F1D1irKjtxd5Uj8A1OTMRmoc1ANRs"
	UAT_API_KEY          = "AIzaSyBULNKqiy-4kJTTsyLoTA6bwAaSFc_7g9M"
	PROD_API_KEY         = "AIzaSyAX_hkFpXOfLzf5NWOVdvqLctPsaX3NdQ8"
)

// GLOBAL VARS
//
//	Remember to update correct values for target environment before running this script
//	Remember to enable the "User_StudentManagement_UsingMasterReplicatedTable" feature flag in the respective environment.
//	  for ex: for stag:
//	    https://admin.staging-green.manabie.io/unleash/projects/default/features/User_StudentManagement_UsingMasterReplicatedTable
var (
	Env = "uat" // local, stag, uat

	Local_tenantID = "end-to-end-dopvo" // select tenant_id from organizations where domain_name = 'e2e';
	Local_email    = "thu.vo+e2eschool@manabie.com"
	Local_pwd      = "M@nabie123"

	Stag_tenantID = "manabie-p7muf" // select tenant_id from organizations where domain_name = 'manabie';
	Stag_email    = "phuc.chau+manabieschooladmin@manabie.com"
	Stag_pwd      = "Manabie@2021"

	Uat_tenantID = "manabie-9h0ng" // select tenant_id from organizations where domain_name = 'manabie';
	Uat_email    = "quangkhai.nguyen+uatmanabie+schooladmin@gmail.com"
	Uat_pwd      = "123456"

	NumStudents      = 6000
	NumStudentOffset = 0

	// If UseStudentsFromFile = true, the script will use student info from the script instead of calling UpsertStudent API
	// The file content just needs to have id and first_name information, you can get it directly from the env database
	// For example, you can use this command to export the file:
	// psql -h localhost -p 5432 -U quangtien.pham@manabie.com -d uat_fatima -t -A -F","
	//   -c "SET ROLE bypass_rls_role;select user_id,first_name from users where first_name like '%stress test student%' order by created_at desc;"
	//   -o ./features/payment/tools/students_ids_and_names.csv
	UseStudentsFromFile = true

	// This will just take effect when UseStudentsFromFile = true
	// If UseStudentsFromFile = fale, the script will always upsert student payment method
	// If UseStudentsFromFile = true and UpsertStudentPaymentMethod = true
	//   then the script will re-upsert students' payment method
	UpsertStudentPaymentMethod = false

	OrderType = pb.OrderType_ORDER_TYPE_NEW // NEW, ENROLLMENT
	Fees      = []pb.FeeType{               // Just support one time fee now
		pb.FeeType_FEE_TYPE_ONE_TIME,
	}
	Materials = []pb.MaterialType{ // Just support recurring material now
		pb.MaterialType_MATERIAL_TYPE_RECURRING,
	}
	Packages = []pb.PackageType{ // Support all
		pb.PackageType_PACKAGE_TYPE_ONE_TIME,
		pb.PackageType_PACKAGE_TYPE_SLOT_BASED,
		pb.PackageType_PACKAGE_TYPE_SCHEDULED,
		// pb.PackageType_PACKAGE_TYPE_FREQUENCY,
	}

	LocationID = "01H2SQAE91TZBRZ0V72F6FGEEY"
	GradeID    = "01GC5VBS7T8C34TP2Q1W4N8PQQ"

	OneTimeMaterial_taxID              = "8e4e280e-c7a6-45dd-a582-fe93612abbc1"
	OneTimeMaterial_discountID         = "abb801a1-30c8-48b5-b5e9-aee373f54cb6"
	OneTimeMaterial_materialID         = "15780bd8-b39d-4ab2-93b6-392bf1ea315f"
	OneTimeMaterial_price              = float32(100)
	OneTimeMaterial_taxPercentage      = float32(10)
	OneTimeMaterial_discountPercentage = float32(10)

	OneTimeFee_taxID              = "cb1884e8-4114-4469-aef8-91aff1d26cea"
	OneTimeFee_discountID         = "4fc4dc9c-8151-498e-a224-7ef69eb561a6"
	OneTimeFee_feeID              = "7cd4331a-751f-4697-915b-c9d7c9681708"
	OneTimeFee_price              = float32(100)
	OneTimeFee_taxPercentage      = float32(10)
	OneTimeFee_discountPercentage = float32(10)

	RecurringMaterial_taxID                       = "6b9821fd-c085-4b53-9005-5dbc7d31203d"
	RecurringMaterial_discountID                  = "1941b157-338b-45ca-9448-7c2345a94d54"
	RecurringMaterial_materialID                  = "20bf7651-9a33-4136-ac40-be4072609a2c"
	RecurringMaterial_billingScheduleCurrPeriodID = "7859f51c-af71-4bd9-a5c4-9b5c61e8376c"
	RecurringMaterial_billingScheduleNextPeriodID = "41033263-2da0-45d5-8bb4-a959ebba144b"
	RecurringMaterial_priceCurrPeriod             = float32(300)
	RecurringMaterial_priceNextPeriod             = float32(600)
	RecurringMaterial_taxPercentage               = float32(10)
	RecurringMaterial_discountPercentage          = float32(10)
	RecurringMaterial_startDate                   = timestamppb.New(time.Now())

	RecurringFee_taxID                       = "3dd660f2-311e-4c54-8839-46deab8385de"
	RecurringFee_discountID                  = "a1a11dad-449c-443a-8565-5a4ac5946862"
	RecurringFee_feeID                       = "82b42dd1-b20d-414c-a1ee-ab50bf68db98"
	RecurringFee_billingScheduleCurrPeriodID = "d47d697c-ce29-438c-9489-74bd7030d734"
	RecurringFee_billingScheduleNextPeriodID = "bf7773f0-bb68-4f9f-844b-226d4034787b"
	RecurringFee_priceCurrPeriod             = float32(300)
	RecurringFee_priceNextPeriod             = float32(600)
	RecurringFee_taxPercentage               = float32(10)
	RecurringFee_discountPercentage          = float32(10)
	RecurringFee_startDate                   = timestamppb.New(time.Now())

	OneTimePackage_taxID              = "0b6ed98f-f982-40ba-a80f-0b0b68779e68"
	OneTimePackage_discountID         = "06888efd-75d7-4e5a-be54-913f5d3c4b1a"
	OneTimePackage_packageID          = "e6ba93f0-c31d-43fa-8e6e-b8973926a8be"
	OneTimePackage_price              = float32(200)
	OneTimePackage_taxPercentage      = float32(10)
	OneTimePackage_discountPercentage = float32(10)
	OneTimePackage_course1ID          = "01H2SRE1APYDBKMB44EZST23F0"
	OneTimePackage_course2ID          = "01H2SREMYFX2XKQ5WFFZ2QSPE6"
	OneTimePackage_course1Weight      = int32(1)
	OneTimePackage_course2Weight      = int32(1)

	SlotBasedPackage_taxID              = "46fe461c-8423-4884-92bb-e02c06c20d4e"
	SlotBasedPackage_discountID         = "33683366-0f53-4c80-aebb-8cdb2ef0b962"
	SlotBasedPackage_packageID          = "6476e76e-da43-40f3-82ed-d0318ea658e0"
	SlotBasedPackage_price              = float32(200)
	SlotBasedPackage_taxPercentage      = float32(10)
	SlotBasedPackage_discountPercentage = float32(10)
	SlotBasedPackage_course1ID          = "01H2SRG08N5X3434YG5AJ7Z4J7"
	SlotBasedPackage_course2ID          = "01H2SRGESWBXNCAPB9BBNBH5M1"
	SlotBasedPackage_course1Slots       = int32(1)
	SlotBasedPackage_course2Slots       = int32(1)

	FrequencyBasedPackage_packageID                   = "64a7863e-2150-45ee-97a3-3212e56bf269"
	FrequencyBasedPackage_billingScheduleCurrPeriodID = "14b05c69-a55c-4bd7-8b2f-0dea8e3c0bb8"
	FrequencyBasedPackage_billingScheduleNextPeriodID = "92469950-c1b2-41bf-91af-801af6649145"
	FrequencyBasedPackage_priceCurrPeriod             = float32(200)
	FrequencyBasedPackage_priceNextPeriod             = float32(200)
	FrequencyBasedPackage_course1ID                   = "01H2SRHX38JWP37N7Y7YXRBFDB"
	FrequencyBasedPackage_course2ID                   = "01H2SRJD1PXF6GB1X2862H63RS7"
	FrequencyBasedPackage_course1Slots                = int32(1)
	FrequencyBasedPackage_course2Slots                = int32(1)
	FrequencyBasedPackage_taxID                       = "8bb6db67-c366-49e2-8c5c-ac499a1dd12e"
	FrequencyBasedPackage_taxPercentage               = float32(10)
	FrequencyBasedPackage_discountID                  = "c6efb4bc-c2d6-4ac1-acc6-e51335c6c9af"
	FrequencyBasedPackage_discountPercentage          = float32(10)
	FrequencyBasedPackage_startDate                   = timestamppb.New(time.Now())

	ScheduleBasedPackage_packageID                   = "c6d0773c-74c5-48d7-87f7-c2e7cd6a2b36"
	ScheduleBasedPackage_billingScheduleCurrPeriodID = "84658341-266d-49d6-b6e3-5e7327ca1961"
	ScheduleBasedPackage_billingScheduleNextPeriodID = "b2c72a5f-3c03-46fb-9080-a83daad95332"
	ScheduleBasedPackage_priceCurrPeriod             = float32(200)
	ScheduleBasedPackage_priceNextPeriod             = float32(200)
	ScheduleBasedPackage_course1ID                   = "01H2SRGYPY8S3R12CKHYW24F9D"
	ScheduleBasedPackage_course2ID                   = "01H2SRHG020Z9QK3R3KQ7X7HE6"
	ScheduleBasedPackage_course1Weight               = int32(1)
	ScheduleBasedPackage_course2Weight               = int32(1)
	ScheduleBasedPackage_taxID                       = "27ccc104-41f7-4b8d-a35f-e090f45065f5"
	ScheduleBasedPackage_taxPercentage               = float32(10)
	ScheduleBasedPackage_discountID                  = "1b153218-16d5-4a2b-b59e-a9816d59c77b"
	ScheduleBasedPackage_discountPercentage          = float32(10)
	ScheduleBasedPackage_startDate                   = timestamppb.New(time.Now())

	/** Use the below SQL script to insert values for bank IDs in local env

	INSERT INTO partner_bank
	  (partner_bank_id, bank_number, bank_name, bank_branch_number, bank_branch_name, deposit_items, account_number, created_at, updated_at, consignor_code, consignor_name, resource_path, is_archived, record_limit)
	VALUES
	  ('01H30XHKFHNBE8N8E2CRZTF4NP', '1234', 'bank-01H30XHKFHNBE8N8E2CRZTF4NP', '123', 'bank-branch-01H30XHKFHNBE8N8E2CRZTF4NP', 'ORDINARY_BANK_ACCOUNT', '1234567', now(), now(), '12345', 'consignor-01H30XHKFHNBE8N8E2CRZTF4NP', '-2147483644', false, 1000),
	  ('01H30XR34HYZ9373P2PDMP9R38', '1234', 'bank-01H30XR34HYZ9373P2PDMP9R38', '123', 'bank-branch-01H30XR34HYZ9373P2PDMP9R38', 'ORDINARY_BANK_ACCOUNT', '1234567', now(), now(), '12345', 'consignor-01H30XR34HYZ9373P2PDMP9R38', '-2147483644', false, 1000);

	INSERT INTO bank
	  (bank_id, bank_code, bank_name, bank_name_phonetic, created_at, updated_at, resource_path, is_archived)
	VALUES
	  ('01H30XR1Z9M7WC5Z0Z8HZ1S75A', '1234', 'bank-01H30XR1Z9M7WC5Z0Z8HZ1S75A', 'phonetic-01H30XR1Z9M7WC5Z0Z8HZ1S75A', now(), now(), '-2147483644', false),
	  ('01H30XR34HYZ9373P2PDMP9R38', '1234', 'bank-01H30XR34HYZ9373P2PDMP9R38', 'phonetic-01H30XR34HYZ9373P2PDMP9R38	', now(), now(), '-2147483644', false);


	INSERT INTO bank_branch
	  (bank_branch_id, bank_branch_code, bank_branch_name, bank_branch_phonetic_name, bank_id, created_at, updated_at, resource_path, is_archived)
	VALUES
	  ('01H30XW6TE4KP1BE038XGE68A5', '123', 'bank-branch-01H30XW6TE4KP1BE038XGE68A5', 'phonetic-01H30XW6TE4KP1BE038XGE68A5', '01H30XR1Z9M7WC5Z0Z8HZ1S75A', now(), now(), '-2147483644', false),
	  ('01H30XXT6E3K827T1X75Y8TBDJ', '123', 'bank-branch-01H30XXT6E3K827T1X75Y8TBDJ', 'phonetic-01H30XXT6E3K827T1X75Y8TBDJ', '01H30XR34HYZ9373P2PDMP9R38', now(), now(), '-2147483644', false);


	INSERT INTO bank_mapping
	  (bank_mapping_id, bank_id, partner_bank_id, created_at, updated_at, resource_path, is_archived)
	VALUES
	  ('01H30Y02J2CWQ3J55F94NZ3CQF', '01H30XR1Z9M7WC5Z0Z8HZ1S75A', '01H30XHKFHNBE8N8E2CRZTF4NP', now(), now(), '-2147483644', false),
	  ('01H30Y177Y2BVQJYSYQS2D9BB4', '01H30XR34HYZ9373P2PDMP9R38', '01H30XR34HYZ9373P2PDMP9R38', now(), now(), '-2147483644', false);

	**/
	Local_BankID1       = "01H30XR1Z9M7WC5Z0Z8HZ1S75A"
	Local_BankID2       = "01H30XR34HYZ9373P2PDMP9R38"
	Local_BankBranchID1 = "01H30XW6TE4KP1BE038XGE68A5"
	Local_BankBranchID2 = "01H30XXT6E3K827T1X75Y8TBDJ"
	Local_IsVerified    = true

	Stag_BankID1       = "01GNBKK38MQ40FC5YJ7QPGFA1N"
	Stag_BankID2       = "01GKMZGJJ3A1710Q0RTKD4K3K2"
	Stag_BankBranchID1 = "01GNBKZ79DJKD2Y7E1MP4VF192"
	Stag_BankBranchID2 = "01GKMZHJ1QHNRJQQSPGBXDFKWT"
	Stag_IsVerified    = true

	UAT_BankID1       = "01GNZWDA363EN7FFD5V9TTX555"
	UAT_BankID2       = "01GNZWDA363EN7FFD5V9TTX557"
	UAT_BankBranchID1 = "01GNZX6H9V2BWWVM4JSCSAQP87"
	UAT_BankBranchID2 = "01GNZXB55RR3KAHSR4DPCC49YF"
	UAT_IsVerified    = true

	DIRECT_DEBIT_BANK_A = fmt.Sprintf("%s-Bank-A", ipb.PaymentMethod_DIRECT_DEBIT.String())
	DIRECT_DEBIT_BANK_B = fmt.Sprintf("%s-Bank-B", ipb.PaymentMethod_DIRECT_DEBIT.String())

	CONVENIENCE_STORE_PERCENTAGE   = 0.1
	DIRECT_DEBIT_BANK_A_PERCENTAGE = 0.7
)

type respB struct {
	Kind         string `json:"kind"`
	LocalID      string `json:"localId"`
	Email        string `json:"email"`
	DisplayName  string `json:"displayName"`
	IDToken      string `json:"idToken"`
	Registered   bool   `json:"registered"`
	RefreshToken string `json:"refreshToken"`
	ExprireIn    string `json:"expiresIn"`
}

type bankAccountInfoRelatedIDs struct {
	BankID1       string
	BankID2       string
	BankBranchID1 string
	BankBranchID2 string
	IsVerified    bool
}

func main() {
	ctx := context.Background()

	// Local
	tenantID := Local_tenantID
	email := Local_email
	pwd := Local_pwd
	identityPlatformApiKey := LOCAL_API_KEY
	bobApiURL := BOB_LOCAL_API_URL         // for token (for local, need to port forward bob and input 127.0.0.1:<bob-port>)
	orderServApiURL := ORDER_LOCAL_API_URL // for calling service (for local, need to port forward and input 127.0.0.1:<service-port>)
	invoiceServApiURL := INVOICE_LOCAL_API_URL
	usermgmtServApiURL := USERMGMT_LOCAL_API_URL
	// mastermgmtServApiURL := MASTERMGMT_LOCAL_API_URL
	isLocal := true // set to "true" for local environment

	bankID1 := Local_BankID1
	bankID2 := Local_BankID2
	bankBranchID1 := Local_BankBranchID1
	bankBranchID2 := Local_BankBranchID2
	isVerified := Local_IsVerified

	switch Env {
	case "local":
		// see above
	case "stag":
		tenantID = Stag_tenantID
		email = Stag_email
		pwd = Stag_pwd
		identityPlatformApiKey = STAG_API_KEY
		bobApiURL = STAG_API_URL
		orderServApiURL = STAG_API_URL
		invoiceServApiURL = STAG_API_URL
		usermgmtServApiURL = STAG_API_URL
		// mastermgmtServApiURL = STAG_API_URL
		isLocal = false

		bankID1 = Stag_BankID1
		bankID2 = Stag_BankID2
		bankBranchID1 = Stag_BankBranchID1
		bankBranchID2 = Stag_BankBranchID2
		isVerified = Stag_IsVerified
	case "uat":
		tenantID = Uat_tenantID
		email = Uat_email
		pwd = Uat_pwd
		identityPlatformApiKey = UAT_API_KEY
		bobApiURL = UAT_API_URL
		orderServApiURL = UAT_API_URL
		invoiceServApiURL = UAT_API_URL
		usermgmtServApiURL = UAT_API_URL
		// mastermgmtServApiURL = UAT_API_URL
		isLocal = false

		bankID1 = UAT_BankID1
		bankID2 = UAT_BankID2
		bankBranchID1 = UAT_BankBranchID1
		bankBranchID2 = UAT_BankBranchID2
		isVerified = UAT_IsVerified
	default:
		log.Fatal("does not support running this script on this env: ", Env)
	}

	rB, err := LoginIdentityPlatform(ctx, tenantID, email, pwd, identityPlatformApiKey)
	connBob := dialHost(bobApiURL, isLocal)
	defer connBob.Close()
	rsp, err := bpb.NewUserModifierServiceClient(connBob).ExchangeToken(contextWithValidVersion(ctx), &bpb.ExchangeTokenRequest{
		Token: rB.IDToken,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	token := rsp.Token
	orderConn := dialHost(orderServApiURL, isLocal)
	defer orderConn.Close()
	invoiceConn := dialHost(invoiceServApiURL, isLocal)
	defer invoiceConn.Close()
	usermgmtConn := dialHost(usermgmtServApiURL, isLocal)
	defer usermgmtConn.Close()
	// mastermgmtConn := dialHost(mastermgmtServApiURL, isLocal)
	// defer mastermgmtConn.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Minute*100)
	defer cancel()

	ctx = ctxWIthToken(ctx, token)

	bankRelatedIDs := &bankAccountInfoRelatedIDs{
		BankID1:       bankID1,
		BankID2:       bankID2,
		BankBranchID1: bankBranchID1,
		BankBranchID2: bankBranchID2,
		IsVerified:    isVerified,
	}

	err = importStudentsAndCreateOrders(ctx, LocationID, GradeID,
		orderConn, usermgmtConn, invoiceConn, bankRelatedIDs)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func getOrderList(ctx context.Context, conn *grpc.ClientConn) error {
	now := time.Now().UTC()
	req := &pb.RetrieveListOfOrdersRequest{
		CurrentTime: timestamppb.New(time.Now()),
		//OrderStatus: pb.OrderStatus_ORDER_STATUS_ALL,
		//LocationIds: []string{"01GV2MBJRW9AS88X5S4C5DXCE7"},
		Filter: &pb.RetrieveListOfOrdersFilter{
			OrderTypes: []pb.OrderType{
				//pb.OrderType_ORDER_TYPE_NEW,
			},
			CreatedFrom: timestamppb.New(now.AddDate(-100, 0, 0)),
			CreatedTo:   timestamppb.New(now.AddDate(1, 0, 0)),
		},
		Paging: &cpb.Paging{
			Limit: 10,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 0,
			},
		},
	}

	client := pb.NewOrderServiceClient(conn)
	resp, err := client.RetrieveListOfOrders(ctx, req)
	if err != nil {
		return fmt.Errorf("getOrderList: %v", err)
	}
	fmt.Println(resp.TotalItems)
	beautifiedArrBytes, err := json.MarshalIndent(resp.Items, "", "    ")
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(beautifiedArrBytes))
	return nil
}

func dialHost(host string, insecure bool) *grpc.ClientConn {
	dialWithTransportSecurityOption := grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	if insecure {
		dialWithTransportSecurityOption = grpc.WithInsecure()
	}

	conn, err := grpc.Dial(host, dialWithTransportSecurityOption)
	if err != nil {
		panic(err.Error())
	}

	return conn
}

func ctxWIthToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx,
		"pkg", "com.manabie.liz",
		"version", "1.0.0",
		"token", token)
}

func LoginIdentityPlatform(ctx context.Context, tenantID string, email string, password string, apiKey string) (*respB, error) {
	url := fmt.Sprintf("%s%s", IDENTITY_TOOLKIT_URL, apiKey)

	loginInfo := struct {
		TenantID          string `json:"tenantId"`
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

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		rB := &respB{}
		json.NewDecoder(resp.Body).Decode(rB)
		return rB, nil
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to login identity platform: and failed to decode error")
	}
	return nil, errors.New("failed to login identity platform:" + string(data))
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func importStudentsAndCreateOrders(ctx context.Context,
	locationID, gradeID string,
	orderConn, uconn, iconn *grpc.ClientConn, bankAccountIDs *bankAccountInfoRelatedIDs,
) error {
	createdStudents := []*upb.StudentProfileV2{}
	if UseStudentsFromFile {
		csvBytes, err := os.ReadFile("students_ids_and_names.csv")
		if err != nil && os.IsNotExist(err) {
			return fmt.Errorf("students_ids_and_names.csv not exist: %v", err)
		}
		bytesReader := bytes.NewReader(csvBytes)
		bufReader := bufio.NewReader(bytesReader)
		for {
			lineByte, _, err := bufReader.ReadLine()
			if err == io.EOF {
				break
			} else if err != nil {
				return status.Error(codes.InvalidArgument, "bufReader.ReadLine() err: "+err.Error())
			}
			// lbs is line before splited into csv format
			lbs := strings.TrimSpace(string(lineByte))
			line := strings.Split(lbs, ",")
			studentProfile := &upb.StudentProfileV2{Id: line[0], FirstName: line[1]}
			createdStudents = append(createdStudents, studentProfile)
		}
	} else {
		numStudents := NumStudents
		numStudentOffset := NumStudentOffset
		for numStudents > 0 {
			studentProfiles := []*upb.StudentProfileV2{}
			for {
				studentProfile := &upb.StudentProfileV2{
					FirstName: fmt.Sprintf("stress test student %d", numStudents+numStudentOffset),
					LastName:  "Payment",
					Email:     fmt.Sprintf("pmststudent-%d@email.com", numStudents+numStudentOffset),
					GradeId:   gradeID,
					EnrollmentStatusHistories: []*upb.EnrollmentStatusHistory{
						{
							LocationId:       locationID,
							EnrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL,
							StartDate:        timestamppb.New(time.Now()),
						},
					},
					Password: "123456",
					StudentPhoneNumbers: &upb.StudentPhoneNumbers{
						ContactPreference: upb.StudentContactPreference_STUDENT_HOME_PHONE_NUMBER,
					},
				}
				studentProfiles = append(studentProfiles, studentProfile)
				numStudents--
				fmt.Println("adding student profiles to bulk upsert, numStudents left:", numStudents)
				if (numStudents%50) == 0 || numStudents <= 0 {
					break
				}
			}
			if len(studentProfiles) > 0 {
				req := &upb.UpsertStudentRequest{
					StudentProfiles: studentProfiles,
				}

				resp, err := upb.NewStudentServiceClient(uconn).UpsertStudent(ctx, req)
				if err != nil {
					return fmt.Errorf("UpsertStudent: %v", err)
				}
				fmt.Println("bulk upserted", len(studentProfiles), "students")
				createdStudents = append(createdStudents, resp.StudentProfiles...)
			}

			fmt.Println("numStudents left:", numStudents)
			time.Sleep(1 * time.Second)
		}
	}

	if !(UseStudentsFromFile && !UpsertStudentPaymentMethod) {
		// Get the students sorted by payment method
		sortedStudentsByPaymentMethod := getStudentsSortedByPaymentMethod(createdStudents)

		// Create CONVENIENCE_STORE payment method for students
		for _, student := range sortedStudentsByPaymentMethod[ipb.PaymentMethod_CONVENIENCE_STORE.String()] {
			if err := createStudentConvenienceStorePaymentMethod(ctx, iconn, student); err != nil {
				return err
			}
		}

		// Create DIRECT_DEBIT payment method for students using bank A
		for _, student := range sortedStudentsByPaymentMethod[DIRECT_DEBIT_BANK_A] {
			if err := createStudentDirectDebitPaymentMethod(ctx, iconn, student, bankAccountIDs.BankID1, bankAccountIDs.BankBranchID1, bankAccountIDs.IsVerified); err != nil {
				return err
			}
		}

		// Create DIRECT_DEBIT payment method for students using bank B
		for _, student := range sortedStudentsByPaymentMethod[DIRECT_DEBIT_BANK_B] {
			if err := createStudentDirectDebitPaymentMethod(ctx, iconn, student, bankAccountIDs.BankID2, bankAccountIDs.BankBranchID2, bankAccountIDs.IsVerified); err != nil {
				return err
			}
		}
	}
	// Create orders
	for _, student := range createdStudents {
		studentID := student.Id
		err := createOrders(ctx, locationID, studentID, orderConn)
		if err != nil {
			return fmt.Errorf("createOrders: %v", err)
		}
		fmt.Println("orders created for", student.FirstName)
	}

	return nil
}

func createStudentConvenienceStorePaymentMethod(ctx context.Context, iconn *grpc.ClientConn, student *upb.StudentProfileV2) error {
	studentID := student.Id
	request := &ipb.UpsertStudentPaymentInfoRequest{
		StudentId: studentID,
		BillingInfo: &ipb.BillingInformation{
			StudentPaymentDetailId: "",
			PayerName:              fmt.Sprintf("%s-payer_name", studentID),
			PayerPhoneNumber:       "",
			BillingAddress: &ipb.BillingAddress{
				BillingAddressId: "",
				PostalCode:       "7",
				PrefectureCode:   "01",
				City:             "hcm",
				Street1:          "hcm",
				Street2:          "",
			},
		},
	}
	_, err := ipb.NewEditPaymentDetailServiceClient(iconn).UpsertStudentPaymentInfo(ctx, request)
	if err != nil {
		return fmt.Errorf("UpsertStudentPaymentInfo: %v", err)
	}
	fmt.Println("added CONVENIENCE_STORE payment method for", student.FirstName)

	return nil
}

func createStudentDirectDebitPaymentMethod(ctx context.Context, iconn *grpc.ClientConn, student *upb.StudentProfileV2, bankID, bankBranchID string, isVerified bool) error {
	studentID := student.Id
	request := &ipb.UpsertStudentPaymentInfoRequest{
		StudentId: studentID,
		BillingInfo: &ipb.BillingInformation{
			StudentPaymentDetailId: "",
			PayerName:              fmt.Sprintf("%s-payer_name", studentID),
			PayerPhoneNumber:       "",
			BillingAddress: &ipb.BillingAddress{
				BillingAddressId: "",
				PostalCode:       "7",
				PrefectureCode:   "01",
				City:             "hcm",
				Street1:          "hcm",
				Street2:          "",
			},
		},
		BankAccountInfo: &ipb.BankAccountInformation{
			BankAccountId:     "",
			BankId:            bankID,
			BankBranchId:      bankBranchID,
			BankAccountHolder: "ACCOUNT HOLDER",
			BankAccountNumber: "1234567",
			BankAccountType:   ipb.BankAccountType(1),
			IsVerified:        isVerified,
		},
	}
	_, err := ipb.NewEditPaymentDetailServiceClient(iconn).UpsertStudentPaymentInfo(ctx, request)
	if err != nil {
		return fmt.Errorf("UpsertStudentPaymentInfo: %v", err)
	}
	fmt.Printf("added DIRECT_DEBIT payment method with Bank %s for %s \n", bankID, student.FirstName)

	return nil
}

// please note that this func can divide them by percentage properly if the number of slice is divisible by 10
// it will still work for other number of students but it can not assure the exact percentage
func getStudentsSortedByPaymentMethod(createdStudents []*upb.StudentProfileV2) map[string][]*upb.StudentProfileV2 {
	m := make(map[string][]*upb.StudentProfileV2)
	m[ipb.PaymentMethod_CONVENIENCE_STORE.String()] = []*upb.StudentProfileV2{}
	m[DIRECT_DEBIT_BANK_A] = []*upb.StudentProfileV2{}
	m[DIRECT_DEBIT_BANK_B] = []*upb.StudentProfileV2{}

	numberOfCSStudents := math.Ceil((float64(len(createdStudents)) * CONVENIENCE_STORE_PERCENTAGE))
	numberOfBankAStudents := math.Floor((float64(len(createdStudents)) * DIRECT_DEBIT_BANK_A_PERCENTAGE))
	numberOfBankBStudents := len(createdStudents) - int(numberOfBankAStudents) - int(numberOfCSStudents)

	if numberOfCSStudents == 0 {
		return m
	}

	m[ipb.PaymentMethod_CONVENIENCE_STORE.String()] = createdStudents[:int(numberOfCSStudents)]

	if numberOfBankAStudents > 0 {
		m[DIRECT_DEBIT_BANK_A] = createdStudents[int(numberOfCSStudents):int(numberOfCSStudents+numberOfBankAStudents)]
	}

	if numberOfBankBStudents > 0 {
		m[DIRECT_DEBIT_BANK_B] = createdStudents[int(numberOfCSStudents+numberOfBankAStudents):]
	}

	return m
}

func createOrders(ctx context.Context, locationID, studentID string, orderConn *grpc.ClientConn) error {
	// req := reqForOneTimeFee(locationID, studentID)
	// req := reqForOneTimePackage(locationID, studentID)
	// req := reqForSlotBasedPackage(locationID, studentID)
	// req := reqForRecurringMaterial(locationID, studentID)
	// req := reqForScheduleBasedPackage(locationID, studentID)
	// req := reqForFrequencyBasedPackage(locationID, studentID)

	// Create fees
	for _, fee := range Fees {
		var req *pb.CreateOrderRequest
		switch fee {
		case pb.FeeType_FEE_TYPE_ONE_TIME:
			req = reqForOneTimeFee(locationID, studentID)
		default:
			fmt.Println("Not supported fee type:", fee)
			continue
		}
		_, err := pb.NewOrderServiceClient(orderConn).CreateOrder(ctx, req)
		if err != nil {
			return fmt.Errorf("CreateOrder: %v", err)
		}
	}

	// Create materials
	for _, material := range Materials {
		var req *pb.CreateOrderRequest
		switch material {
		case pb.MaterialType_MATERIAL_TYPE_RECURRING:
			req = reqForRecurringMaterial(locationID, studentID)
		default:
			fmt.Println("Not supported material type:", material)
			continue
		}
		_, err := pb.NewOrderServiceClient(orderConn).CreateOrder(ctx, req)
		if err != nil {
			return fmt.Errorf("CreateOrder: %v", err)
		}
	}

	// Create packages
	for _, pkg := range Packages {
		var req *pb.CreateOrderRequest
		switch pkg {
		case pb.PackageType_PACKAGE_TYPE_ONE_TIME:
			req = reqForOneTimePackage(locationID, studentID)
		case pb.PackageType_PACKAGE_TYPE_SLOT_BASED:
			req = reqForSlotBasedPackage(locationID, studentID)
		case pb.PackageType_PACKAGE_TYPE_SCHEDULED:
			req = reqForScheduleBasedPackage(locationID, studentID)
		case pb.PackageType_PACKAGE_TYPE_FREQUENCY:
			req = reqForFrequencyBasedPackage(locationID, studentID)
		default:
			fmt.Println("Not supported package type:", pkg)
			continue
		}
		_, err := pb.NewOrderServiceClient(orderConn).CreateOrder(ctx, req)
		if err != nil {
			return fmt.Errorf("CreateOrder: %v", err)
		}
	}

	return nil
}

func reqForOneTimeFee(locationID, studentID string) *pb.CreateOrderRequest {
	var (
		req pb.CreateOrderRequest
	)
	taxID := OneTimeFee_taxID
	discountID := OneTimeFee_discountID
	feeID := OneTimeFee_feeID
	price := OneTimeFee_price
	taxPercentage := OneTimeFee_taxPercentage
	discountPercentage := OneTimeFee_discountPercentage

	// startDate := timestamppb.New(time.Now().AddDate(1, 1, 0))
	finalPrice := price - price*taxPercentage/100

	orderItems := []*pb.OrderItem{}
	billingItems := []*pb.BillingItem{}

	orderItems = append(orderItems, &pb.OrderItem{
		ProductId:  feeID,
		DiscountId: &wrapperspb.StringValue{Value: discountID},
		// StartDate:  startDate,
	})
	billingItems = append(billingItems, &pb.BillingItem{
		ProductId: feeID,
		Price:     price,
		// Quantity:  &wrapperspb.Int32Value{Value: 1},
		DiscountItem: &pb.DiscountBillItem{
			DiscountId:          discountID,
			DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
			DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
			DiscountAmountValue: discountPercentage,
			DiscountAmount:      price * discountPercentage / 100,
		},
		TaxItem: &pb.TaxBillItem{
			TaxId:         taxID,
			TaxPercentage: taxPercentage,
			TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
			TaxAmount:     finalPrice * taxPercentage / (taxPercentage + 100),
		},
		FinalPrice: finalPrice,
	})

	req.StudentId = studentID
	req.LocationId = locationID
	req.OrderComment = " stress test create order one time fee"
	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.OrderType = OrderType

	return &req
}

func reqForOneTimePackage(locationID, studentID string) *pb.CreateOrderRequest {
	var (
		req pb.CreateOrderRequest
	)
	taxID := OneTimePackage_taxID
	discountID := OneTimePackage_discountID
	packageID := OneTimePackage_packageID
	price := OneTimePackage_price
	taxPercentage := OneTimePackage_taxPercentage
	discountPercentage := OneTimePackage_discountPercentage
	course1Weight := OneTimePackage_course1Weight
	course2Weight := OneTimePackage_course2Weight

	// startDate := timestamppb.New(time.Now().AddDate(1, 1, 0))
	finalPrice := price - price*taxPercentage/100
	totalWeight := course1Weight + course2Weight

	orderItems := []*pb.OrderItem{}
	billingItems := []*pb.BillingItem{}
	courseItems := []*pb.CourseItem{
		{
			CourseId:   OneTimePackage_course1ID,
			CourseName: "Payment strest test course 1",
			Weight:     &wrapperspb.Int32Value{Value: course1Weight},
		},
		{
			CourseId:   OneTimePackage_course2ID,
			CourseName: "Payment strest test course 2",
			Weight:     &wrapperspb.Int32Value{Value: course2Weight},
		},
	}

	orderItems = append(orderItems, &pb.OrderItem{
		ProductId:   packageID,
		DiscountId:  &wrapperspb.StringValue{Value: discountID},
		CourseItems: courseItems,
		// StartDate:   startDate,
	})
	billingItems = append(billingItems, &pb.BillingItem{
		ProductId:   packageID,
		Price:       price,
		Quantity:    &wrapperspb.Int32Value{Value: totalWeight}, // price for totalWeight
		CourseItems: courseItems,
		DiscountItem: &pb.DiscountBillItem{
			DiscountId:          discountID,
			DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
			DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
			DiscountAmountValue: discountPercentage,
			DiscountAmount:      price * discountPercentage / 100,
		},
		TaxItem: &pb.TaxBillItem{
			TaxId:         taxID,
			TaxPercentage: taxPercentage,
			TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
			TaxAmount:     finalPrice * taxPercentage / (taxPercentage + 100),
		},
		FinalPrice: finalPrice,
	})

	req.StudentId = studentID
	req.LocationId = locationID
	req.OrderComment = " stress test create order one time package"
	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.OrderType = OrderType

	return &req
}

func reqForSlotBasedPackage(locationID, studentID string) *pb.CreateOrderRequest {
	var (
		req pb.CreateOrderRequest
	)
	taxID := SlotBasedPackage_taxID
	discountID := SlotBasedPackage_discountID
	packageID := SlotBasedPackage_packageID
	price := SlotBasedPackage_price
	taxPercentage := SlotBasedPackage_taxPercentage
	discountPercentage := SlotBasedPackage_discountPercentage
	course1Slots := SlotBasedPackage_course1Slots
	course2Slots := SlotBasedPackage_course2Slots

	// startDate := timestamppb.New(time.Now().AddDate(1, 1, 0))
	finalPrice := price - price*taxPercentage/100
	totalSlots := course1Slots + course2Slots

	orderItems := []*pb.OrderItem{}
	billingItems := []*pb.BillingItem{}
	courseItems := []*pb.CourseItem{
		{
			CourseId:   SlotBasedPackage_course1ID,
			CourseName: "Payment strest test course 1",
			Slot:       &wrapperspb.Int32Value{Value: course1Slots},
		},
		{
			CourseId:   SlotBasedPackage_course2ID,
			CourseName: "Payment strest test course 2",
			Slot:       &wrapperspb.Int32Value{Value: course2Slots},
		},
	}

	orderItems = append(orderItems, &pb.OrderItem{
		ProductId:   packageID,
		DiscountId:  &wrapperspb.StringValue{Value: discountID},
		CourseItems: courseItems,
		// StartDate:   startDate,
	})
	billingItems = append(billingItems, &pb.BillingItem{
		ProductId:   packageID,
		Price:       price,
		Quantity:    &wrapperspb.Int32Value{Value: totalSlots}, // price for totalSlots
		CourseItems: courseItems,
		DiscountItem: &pb.DiscountBillItem{
			DiscountId:          discountID,
			DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
			DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
			DiscountAmountValue: discountPercentage,
			DiscountAmount:      price * discountPercentage / 100,
		},
		TaxItem: &pb.TaxBillItem{
			TaxId:         taxID,
			TaxPercentage: taxPercentage,
			TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
			TaxAmount:     finalPrice * taxPercentage / (taxPercentage + 100),
		},
		FinalPrice: finalPrice,
	})

	req.StudentId = studentID
	req.LocationId = locationID
	req.OrderComment = " stress test create order slot based package"
	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.OrderType = OrderType

	return &req
}

func reqForRecurringMaterial(locationID, studentID string) *pb.CreateOrderRequest {
	var (
		req pb.CreateOrderRequest
	)
	taxID := RecurringMaterial_taxID
	discountID := RecurringMaterial_discountID
	materialID := RecurringMaterial_materialID
	billingScheduleCurrPeriodID := RecurringMaterial_billingScheduleCurrPeriodID
	billingScheduleNextPeriodID := RecurringMaterial_billingScheduleNextPeriodID
	priceCurrPeriod := RecurringMaterial_priceCurrPeriod
	priceNextPeriod := RecurringMaterial_priceNextPeriod
	taxPercentage := RecurringMaterial_taxPercentage
	discountPercentage := RecurringMaterial_discountPercentage
	startDate := RecurringMaterial_startDate

	finalPriceCurrPeriod := priceCurrPeriod - priceCurrPeriod*taxPercentage/100
	finalPriceNextPeriod := priceNextPeriod - priceNextPeriod*taxPercentage/100

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:  materialID,
			DiscountId: &wrapperspb.StringValue{Value: discountID},
			StartDate:  startDate,
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               materialID,
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: billingScheduleCurrPeriodID},
			Price:                   priceCurrPeriod,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: taxPercentage,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     finalPriceCurrPeriod * taxPercentage / (taxPercentage + 100),
			},
			FinalPrice: finalPriceCurrPeriod,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountID,
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: discountPercentage,
				DiscountAmount:      priceCurrPeriod * discountPercentage / 100,
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               materialID,
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: billingScheduleNextPeriodID},
			Price:                   priceNextPeriod,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: taxPercentage,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     finalPriceNextPeriod * taxPercentage / (taxPercentage + 100),
			},
			FinalPrice: finalPriceNextPeriod,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountID,
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: discountPercentage,
				DiscountAmount:      priceNextPeriod * discountPercentage / 100,
			},
		},
	)

	req.StudentId = studentID
	req.LocationId = locationID
	req.OrderComment = " stress test create order recurring material"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = OrderType

	return &req
}

func reqForScheduleBasedPackage(locationID, studentID string) *pb.CreateOrderRequest {
	var (
		req pb.CreateOrderRequest
	)
	taxID := ScheduleBasedPackage_taxID
	discountID := ScheduleBasedPackage_discountID
	packageID := ScheduleBasedPackage_packageID
	billingScheduleCurrPeriodID := ScheduleBasedPackage_billingScheduleCurrPeriodID
	billingScheduleNextPeriodID := ScheduleBasedPackage_billingScheduleNextPeriodID
	priceCurrPeriod := ScheduleBasedPackage_priceCurrPeriod
	priceNextPeriod := ScheduleBasedPackage_priceNextPeriod
	taxPercentage := ScheduleBasedPackage_taxPercentage
	discountPercentage := ScheduleBasedPackage_discountPercentage
	course1Weight := ScheduleBasedPackage_course1Weight
	course2Weight := ScheduleBasedPackage_course2Weight
	startDate := ScheduleBasedPackage_startDate

	finalPriceCurrPeriod := priceCurrPeriod - priceCurrPeriod*taxPercentage/100
	finalPriceNextPeriod := priceNextPeriod - priceNextPeriod*taxPercentage/100
	totalWeight := course1Weight + course2Weight

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}
	courseItems := []*pb.CourseItem{
		{
			CourseId:   ScheduleBasedPackage_course1ID,
			CourseName: "Payment strest test course 1",
			Weight:     &wrapperspb.Int32Value{Value: course1Weight},
		},
		{
			CourseId:   ScheduleBasedPackage_course2ID,
			CourseName: "Payment strest test course 2",
			Weight:     &wrapperspb.Int32Value{Value: course2Weight},
		},
	}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:   packageID,
			DiscountId:  &wrapperspb.StringValue{Value: discountID},
			CourseItems: courseItems,
			StartDate:   startDate,
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               packageID,
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: billingScheduleCurrPeriodID},
			Price:                   priceCurrPeriod,
			Quantity:                &wrapperspb.Int32Value{Value: totalWeight},
			CourseItems:             courseItems,
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: taxPercentage,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     finalPriceCurrPeriod * taxPercentage / (taxPercentage + 100),
			},
			FinalPrice: finalPriceCurrPeriod,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountID,
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: discountPercentage,
				DiscountAmount:      priceCurrPeriod * discountPercentage / 100,
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               packageID,
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: billingScheduleNextPeriodID},
			Price:                   priceNextPeriod,
			Quantity:                &wrapperspb.Int32Value{Value: totalWeight},
			CourseItems:             courseItems,
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: taxPercentage,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     finalPriceNextPeriod * taxPercentage / (taxPercentage + 100),
			},
			FinalPrice: finalPriceNextPeriod,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountID,
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: discountPercentage,
				DiscountAmount:      priceNextPeriod * discountPercentage / 100,
			},
		},
	)

	req.StudentId = studentID
	req.LocationId = locationID
	req.OrderComment = " stress test create order schedule based package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = OrderType

	return &req
}

func reqForFrequencyBasedPackage(locationID, studentID string) *pb.CreateOrderRequest {
	var (
		req pb.CreateOrderRequest
	)
	taxID := FrequencyBasedPackage_taxID
	discountID := FrequencyBasedPackage_discountID
	packageID := FrequencyBasedPackage_packageID
	billingScheduleCurrPeriodID := FrequencyBasedPackage_billingScheduleCurrPeriodID
	billingScheduleNextPeriodID := FrequencyBasedPackage_billingScheduleNextPeriodID
	priceCurrPeriod := FrequencyBasedPackage_priceCurrPeriod
	priceNextPeriod := FrequencyBasedPackage_priceNextPeriod
	taxPercentage := FrequencyBasedPackage_taxPercentage
	discountPercentage := FrequencyBasedPackage_discountPercentage
	course1Slots := FrequencyBasedPackage_course1Slots
	course2Slots := FrequencyBasedPackage_course2Slots
	startDate := FrequencyBasedPackage_startDate

	finalPriceCurrPeriod := priceCurrPeriod - priceCurrPeriod*taxPercentage/100
	finalPriceNextPeriod := priceNextPeriod - priceNextPeriod*taxPercentage/100
	totalSlots := course1Slots + course2Slots

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}
	courseItems := []*pb.CourseItem{
		{
			CourseId:   FrequencyBasedPackage_course1ID,
			CourseName: "Payment strest test course 1",
			Slot:       &wrapperspb.Int32Value{Value: course1Slots},
		},
		{
			CourseId:   FrequencyBasedPackage_course2ID,
			CourseName: "Payment strest test course 2",
			Slot:       &wrapperspb.Int32Value{Value: course2Slots},
		},
	}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:   packageID,
			DiscountId:  &wrapperspb.StringValue{Value: discountID},
			CourseItems: courseItems,
			StartDate:   startDate,
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               packageID,
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: billingScheduleCurrPeriodID},
			Price:                   priceCurrPeriod,
			Quantity:                &wrapperspb.Int32Value{Value: totalSlots},
			CourseItems:             courseItems,
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: taxPercentage,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     finalPriceCurrPeriod * taxPercentage / (taxPercentage + 100),
			},
			FinalPrice: finalPriceCurrPeriod,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountID,
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: discountPercentage,
				DiscountAmount:      priceCurrPeriod * discountPercentage / 100,
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               packageID,
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: billingScheduleNextPeriodID},
			Price:                   priceNextPeriod,
			Quantity:                &wrapperspb.Int32Value{Value: totalSlots},
			CourseItems:             courseItems,
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: taxPercentage,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     finalPriceNextPeriod * taxPercentage / (taxPercentage + 100),
			},
			FinalPrice: finalPriceNextPeriod,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountID,
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: discountPercentage,
				DiscountAmount:      priceNextPeriod * discountPercentage / 100,
			},
		},
	)

	req.StudentId = studentID
	req.LocationId = locationID
	req.OrderComment = " stress test create order frequency based package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = OrderType

	return &req
}
