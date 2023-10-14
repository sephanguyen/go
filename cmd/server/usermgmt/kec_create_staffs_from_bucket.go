package usermgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	userInterceptor "github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"cloud.google.com/go/storage"
	"github.com/gocarina/gocsv"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/square/go-jose.v2/jwt"
)

var (
	mapTimeSheetSettingToggle = map[string]bool{
		"on":  true,
		"off": false,
	}

	token string
)

type KecStaff struct {
	Row               field.Int32  `csv:"Row"`
	Field             field.String `csv:"Field"`
	Message           field.String `csv:"Message"`
	LastName          field.String `csv:"Last Name"`
	FirstName         field.String `csv:"First Name"`
	LastNamePhonetic  field.String `csv:"Last Name (Phonetic)"`
	FirstNamePhonetic field.String `csv:"First Name (Phonetic)"`
	Email             field.String `csv:"Email"`
	PrimaryPhone      field.String `csv:"Primary Phone Number"`
	SecondaryPhone    field.String `csv:"Secondary Phone Number"`
	Birthday          field.String `csv:"Birthday"`
	Gender            field.Int32  `csv:"Gender"`
	WorkingStatus     field.String `csv:"Working Status"`
	StartDate         field.String `csv:"Start Date"`
	EndDate           field.String `csv:"End Date"`
	UserGroup         field.String `csv:"User Group"`
	Location          field.String `csv:"Location"`
	Remarks           field.String `csv:"Remarks"`
	TimesheetSetting  field.String `csv:"Timesheet Settings"`
	Tags              field.String `csv:"Staff Tags"`
}

func init() {
	bootstrap.RegisterJob("usermgmt_kec_create_staffs_from_bucket", runKecCreateStaffsFromBucket).
		Desc("Cmd kec create staffs from bucket").
		StringVar(&token, "token", "", "Token kec create staffs from bucket").
		StringVar(&bucketName, "bucketName", "", "Bucket name kec create staffs from bucket").
		StringVar(&objectName, "objectName", "", "object name kec create staffs from bucket")
}

func runKecCreateStaffsFromBucket(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	return RunKecCreateStaffsFromBucket(ctx, c, rsc, bucketName, objectName)
}

func RunKecCreateStaffsFromBucket(ctx context.Context, _ configurations.Config, rsc *bootstrap.Resources, bucketName, objectName string) error {
	zLogger := rsc.Logger()
	zLogger.Sugar().Info("-----KEC create staffs from bucket-----")

	dbPool := rsc.DBWith("bob")

	var kecStaffs []*KecStaff
	err := getKecData(ctx, bucketName, objectName, &kecStaffs)
	if err != nil {
		return fmt.Errorf("cannot get kec data: %s", err)
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = client.Close()
	}()

	obj := client.Bucket(bucketName).Object("kec_create_staffs_from_bucket_error.csv")

	parsedToken, err := jwt.ParseSigned(token)
	if err != nil {
		return fmt.Errorf("err parse token: %w", err)
	}

	claims := &interceptors.CustomClaims{}
	err = parsedToken.UnsafeClaimsWithoutVerification(claims)
	if err != nil {
		return fmt.Errorf("err parse UnsafeClaimsWithoutVerification: %w", err)
	}

	autoCreateTimesheetServiceClient := tpb.NewAutoCreateTimesheetServiceClient(rsc.GRPCDial("timesheet"))
	staffServiceClient := upb.NewStaffServiceClient(rsc.GRPCDial("usermgmt"))

	ctx = userInterceptor.GRPCContext(ctx, "token", token)
	ctx = interceptors.ContextWithUserID(ctx, claims.Manabie.UserID)
	ctx = interceptors.ContextWithJWTClaims(ctx, claims)

	locations := []string{}
	tags := []string{}
	userGroups := []string{}
	emails := []string{}
	for _, user := range kecStaffs {
		locations = append(locations, strings.Split(user.Location.String(), ";")...)
		tags = append(tags, strings.Split(user.Tags.String(), ";")...)
		userGroups = append(userGroups, strings.Split(user.UserGroup.String(), ";")...)
		emails = append(emails, user.Email.String())
	}

	domainUsers, err := (&repository.DomainUserRepo{}).GetByEmails(ctx, dbPool, emails)
	if err != nil {
		return fmt.Errorf("err GetByEmails: %w", err)
	}
	domainLocations, err := getLocationsByNames(ctx, dbPool, locations)
	if err != nil {
		return fmt.Errorf("err getLocationsByNames: %w", err)
	}
	domainUserGroups, err := getUserGroupsByNames(ctx, dbPool, userGroups)
	if err != nil {
		return fmt.Errorf("err getUserGroupsByNames: %w", err)
	}
	domainTags, err := getTagsByNames(ctx, dbPool, tags)
	if err != nil {
		return fmt.Errorf("err getTagsByNames: %w", err)
	}

	for i, user := range kecStaffs {
		locations := []string{}
		if user.Location.String() != "" {
			locations = strings.Split(user.Location.String(), ";")
		}
		locationIDs, err := getLocationIDsByNames(domainLocations, locations)
		if err != nil {
			user.Row = field.NewInt32(int32(i + 2))
			user.Field = field.NewString("Location")
			user.Message = field.NewString(err.Error())
			continue
		}
		userGroups := []string{}
		if user.UserGroup.String() != "" {
			userGroups = strings.Split(user.UserGroup.String(), ";")
		}
		userGroupIDs, err := getUserGroupIDsByNames(domainUserGroups, userGroups)
		if err != nil {
			user.Row = field.NewInt32(int32(i + 2))
			user.Field = field.NewString("User Group")
			user.Message = field.NewString(err.Error())
			continue
		}

		tags := []string{}
		if user.Tags.String() != "" {
			tags = strings.Split(user.Tags.String(), ";")
		}
		tagIDs, err := getTagIDsByNames(domainTags, tags)
		if err != nil {
			user.Row = field.NewInt32(int32(i + 2))
			user.Field = field.NewString("Staff Tags")
			user.Message = field.NewString(err.Error())
			continue
		}

		var birthday *timestamppb.Timestamp
		if field.IsPresent(user.Birthday) {
			t, err := time.Parse(constant.DateLayout, user.Birthday.String())
			if err != nil {
				user.Row = field.NewInt32(int32(i + 2))
				user.Field = field.NewString("Birthday")
				user.Message = field.NewString("date type should have yyyy/mm/dd format")
				continue
			}
			birthday = timestamppb.New(t)
		}

		var startDate *timestamppb.Timestamp
		if field.IsPresent(user.StartDate) {
			t, err := time.Parse(constant.DateLayout, user.StartDate.String())
			if err != nil {
				user.Row = field.NewInt32(int32(i + 2))
				user.Field = field.NewString("StartDate")
				user.Message = field.NewString("date type should have yyyy/mm/dd format")
				continue
			}
			startDate = timestamppb.New(t)
		}

		var endDate *timestamppb.Timestamp
		if field.IsPresent(user.EndDate) {
			t, err := time.Parse(constant.DateLayout, user.EndDate.String())
			if err != nil {
				user.Row = field.NewInt32(int32(i + 2))
				user.Field = field.NewString("EndDate")
				user.Message = field.NewString("date type should have yyyy/mm/dd format")
				continue
			}
			endDate = timestamppb.New(t)
		}

		var userID string
		for _, domainUser := range domainUsers {
			if domainUser.Email().String() == user.Email.String() {
				userID = domainUser.UserID().String()
				break
			}
		}
		if userID == "" {
			resp, err := staffServiceClient.CreateStaff(ctx, &upb.CreateStaffRequest{
				Staff: &upb.CreateStaffRequest_StaffProfile{
					UserNameFields: &upb.UserNameFields{
						FirstName:         user.FirstName.String(),
						LastName:          user.LastName.String(),
						FirstNamePhonetic: user.FirstNamePhonetic.String(),
						LastNamePhonetic:  user.LastNamePhonetic.String(),
					},
					Email: user.Email.String(),
					StaffPhoneNumber: []*upb.StaffPhoneNumber{
						{
							PhoneNumber:     user.PrimaryPhone.String(),
							PhoneNumberType: upb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
						},
						{
							PhoneNumber:     user.SecondaryPhone.String(),
							PhoneNumberType: upb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER,
						},
					},
					Birthday:       birthday,
					StartDate:      startDate,
					EndDate:        endDate,
					WorkingStatus:  upb.StaffWorkingStatus(upb.StaffWorkingStatus_value[user.WorkingStatus.String()]),
					Gender:         upb.Gender(user.Gender.Int32()),
					UserGroup:      upb.UserGroup_USER_GROUP_TEACHER,
					OrganizationId: claims.OrganizationID().String(),
					Country:        cpb.Country_COUNTRY_JP,
					LocationIds:    locationIDs,
					UserGroupIds:   userGroupIDs,
					Remarks:        user.Remarks.String(),
					TagIds:         tagIDs,
				},
			})
			if err != nil {
				user.Row = field.NewInt32(int32(i + 2))
				user.Message = field.NewString(fmt.Sprintf("CreateStaff %v", err))
				continue
			}
			userID = resp.Staff.StaffId
		}

		// for cpu taking a deep breath
		time.Sleep(time.Second)

		_, err = autoCreateTimesheetServiceClient.UpdateAutoCreateTimesheetFlag(ctx, &tpb.UpdateAutoCreateTimesheetFlagRequest{
			StaffId: userID,
			FlagOn:  mapTimeSheetSettingToggle[strings.ToLower(user.TimesheetSetting.String())],
		})
		if err != nil {
			user.Row = field.NewInt32(int32(i + 2))
			user.Message = field.NewString(fmt.Sprintf("UpdateAutoCreateTimesheetFlag %v", err))
			continue
		}
	}

	err = writeToCSVInBucket(ctx, obj, kecStaffs)
	if err != nil {
		return fmt.Errorf("err writeToCSVInBucket: %v", err)
	}

	return nil
}

func writeToCSVInBucket(ctx context.Context, obj *storage.ObjectHandle, kecStaffs []*KecStaff) error {
	writer := obj.NewWriter(ctx)

	csvContent, err := gocsv.MarshalString(kecStaffs)
	if err != nil {
		return fmt.Errorf("error marshal csv: %s", err.Error())
	}

	if _, err := writer.Write([]byte(csvContent)); err != nil {
		_ = writer.Close()
		return fmt.Errorf("error write csv to storage: %s", err.Error())
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("error closing writer: %s", err.Error())
	}

	return nil
}

func getLocationIDsByNames(listOfLocations entity.DomainLocations, names []string) ([]string, error) {
	locationIDs := []string{}
	for _, name := range names {
		var locationID string
		for _, location := range listOfLocations {
			if name == location.Name().String() {
				locationID = location.LocationID().String()
			}
		}
		if locationID == "" {
			return nil, fmt.Errorf("cannot find location name: %s", name)
		}
		locationIDs = append(locationIDs, locationID)
	}
	return locationIDs, nil
}

func getTagIDsByNames(listOfTags entity.DomainTags, names []string) ([]string, error) {
	tagIDs := []string{}
	for _, name := range names {
		var tagID string
		for _, tag := range listOfTags {
			if name == tag.TagName().String() {
				tagID = tag.TagID().String()
			}
		}
		if tagID == "" {
			return nil, fmt.Errorf("cannot find tag name: %s", name)
		}
		tagIDs = append(tagIDs, tagID)
	}
	return tagIDs, nil
}

func getUserGroupIDsByNames(listOfUserGroups []entity.DomainUserGroup, names []string) ([]string, error) {
	userGroupIDs := []string{}
	for _, name := range names {
		var userGroupID string
		for _, userGroup := range listOfUserGroups {
			if name == userGroup.Name().String() {
				userGroupID = userGroup.UserGroupID().String()
			}
		}
		if userGroupID == "" {
			return nil, fmt.Errorf("cannot find userGroup name: %s", name)
		}
		userGroupIDs = append(userGroupIDs, userGroupID)
	}
	return userGroupIDs, nil
}

func getUserGroupsByNames(ctx context.Context, db database.QueryExecer, names []string) ([]entity.DomainUserGroup, error) {
	userGroupRepoEntity := repository.UserGroup{}
	field, _ := userGroupRepoEntity.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM public.%s WHERE user_group_name = ANY($1)", strings.Join(field, ","), userGroupRepoEntity.TableName())

	rows, err := db.Query(ctx, query, names)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userGroups []entity.DomainUserGroup
	for rows.Next() {
		userGroupRepoEntity := &repository.UserGroup{}
		_, value := userGroupRepoEntity.FieldMap()
		err := rows.Scan(value...)
		if err != nil {
			return nil, err
		}
		userGroups = append(userGroups, userGroupRepoEntity)
	}

	return userGroups, nil
}

func getLocationsByNames(ctx context.Context, db database.QueryExecer, names []string) (entity.DomainLocations, error) {
	locationRepoEntity := repository.Location{}
	field, _ := locationRepoEntity.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM public.%s WHERE name = ANY($1)", strings.Join(field, ","), locationRepoEntity.TableName())

	rows, err := db.Query(ctx, query, names)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations entity.DomainLocations
	for rows.Next() {
		locationRepoEntity := &repository.Location{}
		_, value := locationRepoEntity.FieldMap()
		err := rows.Scan(value...)
		if err != nil {
			return nil, err
		}
		locations = append(locations, locationRepoEntity)
	}

	return locations, nil
}

func getTagsByNames(ctx context.Context, db database.QueryExecer, names []string) (entity.DomainTags, error) {
	tagRepoEntity := repository.Tag{}
	field, _ := tagRepoEntity.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM public.%s WHERE user_tag_name = ANY($1)", strings.Join(field, ","), tagRepoEntity.TableName())

	rows, err := db.Query(ctx, query, names)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags entity.DomainTags
	for rows.Next() {
		tagRepoEntity := &repository.Tag{}
		_, value := tagRepoEntity.FieldMap()
		err := rows.Scan(value...)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tagRepoEntity)
	}

	return tags, nil
}
