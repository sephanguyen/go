package payment

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/features/payment/entities"
	"github.com/manabie-com/backend/features/usermgmt"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/try"
	payment_entities "github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/metadata"
)

const (
	UserGroupStudent             = "student"
	UserGroupAdmin               = "admin"
	UserGroupTeacher             = "teacher"
	UserGroupParent              = "parent"
	UserGroupSchoolAdmin         = "school admin"
	UserGroupOrganizationManager = "organization manager"
	UserGroupUnauthenticated     = "unauthenticated"
	UserGroupHQStaff             = "hq staff"
	UserGroupCentreLead          = "centre lead"
	UserGroupCentreManager       = "centre manager"
	UserGroupCentreStaff         = "centre staff"
	UserGroupTeacherLead         = "teacher lead"
)

type userOption func(u *entities.User)

func withID(id string) userOption {
	return func(u *entities.User) {
		_ = u.ID.Set(id)
	}
}

func withRole(group string) userOption {
	return func(u *entities.User) {
		_ = u.Group.Set(group)
	}
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

// signedAsAccount user root account of ManabieSchool to sign in a user on ManabieOrgLocation location with specific user group
// Make sure user is synced to fatima, if not, insert user in fatima
func (s *suite) signedAsAccount(ctx context.Context, group string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = common.ValidContext(ctx, constants.ManabieSchool, s.RootAccount[constants.ManabieSchool].UserID, s.RootAccount[constants.ManabieSchool].Token)
	roleWithLocation := usermgmt.RoleWithLocation{
		LocationIDs: []string{constants.ManabieOrgLocation},
	}
	stepState.CurrentSchoolID = constants.ManabieSchool
	switch group {
	case UserGroupSchoolAdmin:
		roleWithLocation.RoleName = constant.RoleSchoolAdmin
	case UserGroupHQStaff:
		roleWithLocation.RoleName = constant.RoleHQStaff
	case UserGroupCentreLead:
		roleWithLocation.RoleName = constant.RoleCentreLead
	case UserGroupCentreManager:
		roleWithLocation.RoleName = constant.RoleCentreManager
	case UserGroupCentreStaff:
		roleWithLocation.RoleName = constant.RoleCentreStaff
	case UserGroupTeacher:
		roleWithLocation.RoleName = constant.RoleTeacher
	case UserGroupTeacherLead:
		roleWithLocation.RoleName = constant.RoleTeacherLead
	case UserGroupStudent:
		roleWithLocation.RoleName = constant.RoleStudent
	case UserGroupParent:
		roleWithLocation.RoleName = constant.UserGroupParent
	case UserGroupOrganizationManager:
		roleWithLocation.RoleName = constant.UserGroupOrganizationManager
	default:
		return StepStateToContext(ctx, stepState), errors.New("user group is invalid")
	}

	authInfo, err := usermgmt.SignIn(ctx, s.BobDBTrace, s.AuthPostgresDB, s.ShamirConn, s.Cfg.JWTApplicant, s.FirebaseAddress, s.UserMgmtConn, roleWithLocation, []string{constants.ManabieOrgLocation})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.CurrentUserID = authInfo.UserID
	stepState.AuthToken = authInfo.Token
	stepState.LocationID = constants.ManabieOrgLocation
	stepState.CurrentUserGroup = group
	ctx = common.ValidContext(ctx, constants.ManabieSchool, authInfo.UserID, authInfo.Token)

	err = try.Do(func(attempt int) (bool, error) {
		time.Sleep(1 * time.Second)
		err = s.getAdmin(ctx, stepState.CurrentUserID)
		if err == nil {
			return false, nil
		}
		retry := attempt <= 5
		if retry {
			return true, nil
		}
		err = s.insertAdmin(ctx, stepState.CurrentUserID, fmt.Sprintf("name-user-id-%s", authInfo.UserID))
		if err != nil {
			return false, fmt.Errorf("error when user info have not been synced from bob to fatima: %s", err.Error())
		}
		return false, nil
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedInAdmin(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := idutil.ULIDNow()
	var err error

	stepState.CurrentUserID = id
	stepState.CurrentUserGroup = constant.UserGroupAdmin

	ctx, err = s.aValidUser(StepStateToContext(ctx, stepState), withID(id), withRole(constant.UserGroupAdmin))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken, err = s.generateValidAuthenticationToken(id, constant.UserGroupAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) generateValidAuthenticationToken(sub, userGroup string) (string, error) {
	firebaseToken, err := generateAuthenticationToken(sub, "templates/"+userGroup+".template")
	if err != nil {
		return "", err
	}
	token, err := helper.ExchangeToken(firebaseToken, sub, userGroup, applicantID, s.getSchool(), s.ShamirConn)
	if err != nil {
		return "", err
	}
	return token, nil
}

func generateAuthenticationToken(sub string, template string) (string, error) {
	resp, err := http.Get("http://" + firebaseAddr + "/token?template=" + template + "&UserID=" + sub)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken:cannot generate new user token, err: %v", err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()

	return string(b), nil
}

func (s *suite) aValidUser(ctx context.Context, opts ...userOption) (context.Context, error) {
	num := rand.Int()
	ctx, err := s.aValidUserInDB(ctx, s.BobDBTrace, num, opts...)
	if err != nil {
		return ctx, err
	}
	ctx, err = s.aValidUserInDB(ctx, s.FatimaDBTrace, num, opts...)
	if err != nil {
		return ctx, err
	}
	return ctx, err
}

func (s *suite) aValidUserInDB(ctx context.Context, dbConn *database.DBTrace, randomNumber int, opts ...userOption) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	u := &entities.User{}
	database.AllNullEntity(u)
	firstName := fmt.Sprintf("valid-user-first-name-%d", randomNumber)
	lastName := fmt.Sprintf("valid-user-last-name-%d", randomNumber)

	err := multierr.Combine(
		u.FullName.Set(helper.CombineFirstNameAndLastNameToFullName(firstName, lastName)),
		u.FirstName.Set(firstName),
		u.LastName.Set(lastName),
		u.PhoneNumber.Set(fmt.Sprintf("+848%d", randomNumber)),
		u.Email.Set(fmt.Sprintf("valid-user-%d@email.com", randomNumber)),
		u.Country.Set(cpb.Country_COUNTRY_VN.String()),
		u.Group.Set(constant.UserGroupAdmin),
		u.Avatar.Set(fmt.Sprintf("http://valid-user-%d", randomNumber)),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, opt := range opts {
		opt(u)
	}

	err = database.ExecInTx(ctx, dbConn, func(ctx context.Context, tx pgx.Tx) error {
		err := s.createUser(ctx, tx, u)
		if err != nil {
			return fmt.Errorf("cannot create user: %w", err)
		}
		schoolID := int64(stepState.CurrentSchoolID)
		if schoolID == 0 {
			schoolID = constants.ManabieSchool
		}
		if u.Group.String == constant.UserGroupTeacher {
			teacherRepo := repository.TeacherRepo{}
			t := &entity.Teacher{}
			database.AllNullEntity(t)
			t.ID = u.ID
			err := t.SchoolIDs.Set([]int64{schoolID})
			if err != nil {
				return err
			}

			err = teacherRepo.CreateMultiple(ctx, tx, []*entity.Teacher{t})
			if err != nil {
				return fmt.Errorf("cannot create teacher:%w", err)
			}
		} else if u.Group.String == constant.UserGroupSchoolAdmin {
			// schoolAdminRepo := repository.SchoolAdminRepo{}
			schoolAdminAccount := &entity.SchoolAdmin{}
			database.AllNullEntity(schoolAdminAccount)
			err := multierr.Combine(
				schoolAdminAccount.SchoolAdminID.Set(u.ID.String),
				schoolAdminAccount.SchoolID.Set(schoolID),
			)
			if err != nil {
				return err
			}
			err = helper.CreateMultipleSchoolAdmins(ctx, tx, []*entity.SchoolAdmin{schoolAdminAccount})
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	uGroup := &entity.UserGroup{}
	database.AllNullEntity(uGroup)

	err = multierr.Combine(
		uGroup.GroupID.Set(u.Group.String),
		uGroup.UserID.Set(u.ID.String),
		uGroup.IsOrigin.Set(true),
		uGroup.Status.Set("USER_GROUP_STATUS_ACTIVE"),
		uGroup.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	userGroupRepo := &repository.UserGroupRepo{}
	err = userGroupRepo.Upsert(ctx, dbConn, uGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("userGroupRepo.Upsert: %w %s", err, u.Group.String)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedInStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := idutil.ULIDNow()
	var err error
	stepState.AuthToken, err = s.generateValidAuthenticationToken(id, "phone")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentUserID = id
	stepState.CurrentUserGroup = constant.UserGroupStudent

	ctx, err = s.aValidStudentInDB(StepStateToContext(ctx, stepState), id)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken, err = s.generateValidAuthenticationToken(id, "phone")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) aValidStudentInDB(ctx context.Context, id string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentRepo := repository.StudentRepo{}
	now := time.Now()
	student := &entity.LegacyStudent{}
	database.AllNullEntity(student)
	err := multierr.Combine(
		student.ID.Set(id),
		student.CurrentGrade.Set(12),
		student.OnTrial.Set(true),
		student.TotalQuestionLimit.Set(10),
		student.SchoolID.Set(constants.ManabieSchool),
		student.CreatedAt.Set(now),
		student.UpdatedAt.Set(now),
		student.BillingDate.Set(now),
		student.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED"),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	err = studentRepo.Create(ctx, s.BobDBTrace, student)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return s.aValidUser(StepStateToContext(ctx, stepState), withID(student.ID.String), withRole(constant.UserGroupStudent))
}

func contextWithToken(ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)

	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", stepState.AuthToken)
}

func parseToDate(value string) (time.Time, error) {
	const layoutISO string = "2006-01-02"
	var (
		timeElement time.Time
		err         error
	)
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return time.Time{}, fmt.Errorf("empty string")
	}

	if len(trimmedValue) == len(layoutISO) {
		timeElement, err = time.Parse(layoutISO, trimmedValue)
	} else {
		timeElement, err = time.Parse(time.RFC3339, trimmedValue)
	}

	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing %v: %w", trimmedValue, err)
	}
	return timeElement, nil
}

func IsEqualNumericAndFloat32(numeric pgtype.Numeric, float32Value float32) bool {
	tmpFloatValue := float32(numeric.Int.Int64()) * float32(math.Pow10(int(numeric.Exp)))
	return fmt.Sprintf("%.2f", tmpFloatValue) == fmt.Sprintf("%.2f", float32Value)
}

func (s *suite) getSchool() int64 {
	if s.SchoolID != "" {
		intSchool, _ := strconv.ParseInt(s.SchoolID, 10, 64)
		return intSchool
	}

	return constants.ManabieSchool
}

func countOrderItemForRecurringProduct(dbOrderItems []payment_entities.OrderItem, orderItems []*pb.OrderItem) int {
	foundOrderItem := 0
	for _, item := range orderItems {
		for _, dbItem := range dbOrderItems {
			if item.ProductId == dbItem.ProductID.String &&
				((dbItem.DiscountID.Status == pgtype.Null) ||
					(dbItem.DiscountID.Status == pgtype.Present && dbItem.DiscountID.String == item.DiscountId.Value)) &&
				((dbItem.StartDate.Status == pgtype.Null) ||
					(dbItem.StartDate.Status == pgtype.Present && dbItem.StartDate.Time.Equal(item.StartDate.AsTime()))) {
				foundOrderItem++
			}
		}
	}
	return foundOrderItem
}

func countBillItemForRecurringProduct(billItems []payment_entities.BillItem, billingItems []*pb.BillingItem, billingStatus pb.BillingStatus, billingType pb.BillingType, locationID string) int {
	foundBillItem := 0
	for _, item := range billingItems {
		for _, dbItem := range billItems {
			if item.ProductId == dbItem.ProductID.String &&
				dbItem.BillStatus.String == billingStatus.String() &&
				dbItem.BillType.String == billingType.String() &&
				IsEqualNumericAndFloat32(dbItem.FinalPrice, item.FinalPrice) &&
				dbItem.LocationID.String == locationID {
				foundBillItem++
			}
		}
	}
	return foundBillItem
}

func getInclusivePercentTax(priceAfterDiscount float32, taxPercent float32) float32 {
	return float32(float64(priceAfterDiscount*taxPercent) / float64(100+taxPercent))
}

func getProratedPrice(priceOrder float32, numerator int, denominator int) float32 {
	return (priceOrder * float32(numerator)) / float32(denominator)
}

func getPercentDiscountedPrice(priceOrder float32, percentDiscount int) float32 {
	return priceOrder - getPercentDiscountValue(priceOrder, percentDiscount)
}

func getPercentDiscountValue(priceOrder float32, percentDiscount int) float32 {
	discountValue := priceOrder * (float32(percentDiscount) / 100)
	return float32(math.Round(float64(discountValue)))
}

func getFrequencyBasePrice(price float32, slots int) float32 {
	return price * float32(slots)
}

func getScheduleBasePrice(price float32, totalWeight int) float32 {
	return price * float32(totalWeight)
}

func assignUserGroupToUser(ctx context.Context, dbBob database.QueryExecer, userID string, userGroupIDs []string) error {
	userGroupMembers := make([]*entity.UserGroupMember, 0)
	for _, userGroupID := range userGroupIDs {
		userGroupMem := &entity.UserGroupMember{}
		database.AllNullEntity(userGroupMem)
		if err := multierr.Combine(
			userGroupMem.UserID.Set(userID),
			userGroupMem.UserGroupID.Set(userGroupID),
			userGroupMem.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
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

func (s *suite) createUserGroupWithRoleNames(ctx context.Context, roleNames []string) (*entity.UserGroupV2, error) {
	req := &upb.CreateUserGroupRequest{
		UserGroupName: fmt.Sprintf("user-group_%s", idutil.ULIDNow()),
	}

	stmt := "SELECT role_id FROM role WHERE deleted_at IS NULL AND role_name = ANY($1) LIMIT $2"
	rows, err := s.BobDB.Query(ctx, stmt, roleNames, len(roleNames))
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
				LocationIds: []string{constants.ManabieOrgLocation},
			},
		)
	}

	resourcePath, _ := strconv.Atoi(golibs.ResourcePathFromCtx(ctx))

	resp, err := s.HandleCreateUserGroup(ctx, s.FatimaDBTrace, req, resourcePath)
	if err != nil {
		return nil, fmt.Errorf("HandleCreateUserGroup: %w", err)
	}

	return resp, nil
}

func (s *suite) aSignedInStaff(ctx context.Context, roles []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := idutil.ULIDNow()
	ctx, err := s.aValidUser(StepStateToContext(ctx, stepState), withID(id), withRole(constant.UserGroupSchoolAdmin))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUser - school admin: %w", err)
	}

	token, err := s.generateValidAuthenticationToken(id, constant.UserGroupSchoolAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = token

	createUserGroupResp, err := s.createUserGroupWithRoleNames(ctx, roles)
	if err != nil {
		return nil, err
	}

	if err := assignUserGroupToUser(ctx, s.FatimaDBTrace, id, []string{createUserGroupResp.UserGroupID.String}); err != nil {
		return nil, err
	}

	stepState.CurrentUserID = id
	return ctx, nil
}

func userGroupPayloadToUserGroupEnt(payload *upb.CreateUserGroupRequest, resourcePath string) (*entity.UserGroupV2, error) {
	userGroup := &entity.UserGroupV2{}
	database.AllNullEntity(userGroup)
	if err := multierr.Combine(
		userGroup.UserGroupID.Set(idutil.ULIDNow()),
		userGroup.UserGroupName.Set(payload.UserGroupName),
		userGroup.ResourcePath.Set(resourcePath),
		userGroup.OrgLocationID.Set(constants.ManabieOrgLocation),
		userGroup.IsSystem.Set(false),
	); err != nil {
		return nil, fmt.Errorf("set user group fail: %w", err)
	}

	return userGroup, nil
}

func roleWithLocationsPayloadToGrantedRole(payload *upb.RoleWithLocations, userGroupID string, resourcePath string) (*entity.GrantedRole, error) {
	grantedRole := &entity.GrantedRole{}
	database.AllNullEntity(grantedRole)
	if err := multierr.Combine(
		grantedRole.GrantedRoleID.Set(idutil.ULIDNow()),
		grantedRole.UserGroupID.Set(userGroupID),
		grantedRole.RoleID.Set(payload.RoleId),
		grantedRole.ResourcePath.Set(resourcePath),
	); err != nil {
		return nil, fmt.Errorf("set granted role fail: %w", err)
	}

	return grantedRole, nil
}

func (s *suite) HandleCreateUserGroup(ctx context.Context, tx *database.DBTrace, req *upb.CreateUserGroupRequest, resourcePath int) (*entity.UserGroupV2, error) {
	var userGroup *entity.UserGroupV2
	var err error

	userGroupV2Repo := &repository.UserGroupV2Repo{}
	grantedRoleRepo := &repository.GrantedRoleRepo{}
	// convert payload to entity

	if userGroup, err = userGroupPayloadToUserGroupEnt(req, fmt.Sprint(resourcePath)); err != nil {
		return nil, fmt.Errorf("s.UserGroupPayloadToUserGroupEnts: %w", err)
	}
	userGroup.UserGroupID.Set(idutil.ULIDNow())

	if err = database.ExecInTx(ctx, tx, func(ctx context.Context, tx pgx.Tx) error {
		// create usergroup first
		if err = userGroupV2Repo.Create(ctx, tx, userGroup); err != nil {
			return fmt.Errorf("userGroupV2Repo.Create: %w", err)
		}

		var grantedRole *entity.GrantedRole
		for _, roleWithLocations := range req.RoleWithLocations {
			// convert payload to entity
			if grantedRole, err = roleWithLocationsPayloadToGrantedRole(roleWithLocations, userGroup.UserGroupID.String, fmt.Sprint(resourcePath)); err != nil {
				return fmt.Errorf("s.RoleWithLocationsPayloadToGrantedRole: %w", err)
			}
			grantedRole.GrantedRoleID.Set(idutil.ULIDNow())
			// create granted_role
			if err = grantedRoleRepo.Create(ctx, tx, grantedRole); err != nil {
				return fmt.Errorf("grantedRoleRepo.Create: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("database.ExecInTx: %w", err)
	}

	return userGroup, nil
}

func (s *suite) checkCreatedOrderDetailsAndActionLogs(ctx context.Context, orderType pb.OrderType) error {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return stepState.ResponseErr
	}

	var (
		successful   bool
		orderID      string
		studentID    string
		locationID   string
		orderComment string
		reqOrderType pb.OrderType
	)

	switch orderType {
	case pb.OrderType_ORDER_TYPE_NEW:
		req := stepState.Request.(*pb.CreateOrderRequest)
		res := stepState.Response.(*pb.CreateOrderResponse)

		studentID = req.StudentId
		orderComment = req.OrderComment
		locationID = req.LocationId
		reqOrderType = req.OrderType

		successful = res.Successful
		orderID = res.OrderId
	case pb.OrderType_ORDER_TYPE_CUSTOM_BILLING:
		req := stepState.Request.(*pb.CreateCustomBillingRequest)
		res := stepState.Response.(*pb.CreateCustomBillingResponse)

		studentID = req.StudentId
		orderComment = req.OrderComment
		locationID = req.LocationId
		reqOrderType = req.OrderType

		successful = res.Successful
		orderID = res.OrderId
	default:
		return fmt.Errorf("undefined create order type")
	}
	if !successful {
		return fmt.Errorf("create order with type: %v failed", reqOrderType.String())
	}
	order, err := s.getOrder(ctx, orderID)
	if err != nil {
		return err
	}

	if order.OrderComment.String != orderComment ||
		order.LocationID.String != locationID ||
		order.StudentID.String != studentID ||
		order.OrderType.String != reqOrderType.String() {
		return fmt.Errorf("create order with type: %v error wrong data", reqOrderType.String())
	}

	orderActionLogs, err := s.getOrderActionLogs(ctx, orderID)
	if err != nil {
		return err
	}
	if len(orderActionLogs) == 0 {
		return fmt.Errorf("create order action log fail")
	}

	if !(orderActionLogs[0].Action.String == pb.OrderActionStatus_ORDER_ACTION_SUBMITTED.String() &&
		orderActionLogs[0].Comment.String == order.OrderComment.String &&
		orderActionLogs[0].UserID.String == stepState.CurrentUserID) {
		return fmt.Errorf("create order action log incorrect content")
	}

	return nil
}

func (s *suite) validateCreatedOrderItemsAndBillItemsForRecurringProducts(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.CreateOrderRequest)
	res := stepState.Response.(*pb.CreateOrderResponse)

	orderItems, err := s.getOrderItems(ctx, res.OrderId)
	if err != nil {
		return err
	}

	foundOrderItem := countOrderItemForRecurringProduct(orderItems, req.OrderItems)
	if foundOrderItem < len(req.OrderItems) {
		return fmt.Errorf("missing order item")
	}

	billingItems, err := s.getBillItems(ctx, res.OrderId)
	if err != nil {
		return err
	}

	foundBillItem := countBillItemForRecurringProduct(billingItems, req.BillingItems, pb.BillingStatus_BILLING_STATUS_BILLED, pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER, req.LocationId)
	if foundBillItem < len(req.BillingItems) {
		return fmt.Errorf("missing billing item")
	}

	foundUpcomingBillItem := countBillItemForRecurringProduct(billingItems, req.UpcomingBillingItems, pb.BillingStatus_BILLING_STATUS_PENDING, pb.BillingType_BILLING_TYPE_UPCOMING_BILLING, req.LocationId)
	if foundUpcomingBillItem < len(req.UpcomingBillingItems) {
		return fmt.Errorf("missing upcoming billing item")
	}

	return nil
}

func (s *suite) validateCreatedOrderItemsAndBillItemsForOneTimeProducts(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.CreateOrderRequest)
	res := stepState.Response.(*pb.CreateOrderResponse)

	orderItems, err := s.getOrderItems(ctx, res.OrderId)
	if err != nil {
		return err
	}
	foundOrderItem := countOrderItem(orderItems, req.OrderItems)
	if foundOrderItem < len(req.OrderItems) {
		return fmt.Errorf("create miss orderItem")
	}

	billItems, err := s.getBillItems(ctx, res.OrderId)
	if err != nil {
		return err
	}
	foundBillItem := countBillItem(billItems, req.BillingItems, req.LocationId)
	if foundBillItem < len(req.BillingItems) {
		return fmt.Errorf("create miss billItem")
	}

	return nil
}

func (s *suite) createUser(ctx context.Context, db database.QueryExecer, user *entities.User) error {
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
	_, err := database.Insert(ctx, user, db.Exec)
	if err != nil {
		return fmt.Errorf("user not inserted: %w", err)
	}

	return nil
}

func StartOfDate(t time.Time) time.Time {
	return t.Truncate(24 * time.Hour)
}

func EndOfDate(t time.Time) time.Time {
	nextDate := StartOfDate(t).AddDate(0, 0, 1)
	endOfDate := nextDate.Add(-1 * time.Second)
	return endOfDate
}
