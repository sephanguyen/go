package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pbc "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"firebase.google.com/go/v4/auth"
	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/metadata"
)

const (
	StableTemplateImportStudentHeaders = "user_id,external_user_id,last_name,first_name,last_name_phonetic,first_name_phonetic,email,grade,birthday,gender,remarks,student_tag,student_phone_number,home_phone_number,contact_preference,postal_code,prefecture,city,first_street,second_street,school,school_course,start_date,end_date,enrollment_status,location,status_start_date"
	StableTemplateImportStudentValues  = "uuid,externaluserid,lastname,firstname,lastname_phonetic,firstname_phonetic,student@email.com,1,2000/01/12,1,remarks,tag_partner_id_1;tag_partner_id_2,123456789,123456789,1,7000000,prefecture value,city value,street1 value,street2 value,school_partner_id_1;school_partner_id_2,school_course_partner_id_1;school_course_partner_id_2,2000/01/08;2000/01/09,2000/01/10;2000/01/11,1;2,location_partner_id_1;location_partner_id_2,2000/01/12;2000/01/13"

	PhoneNumberPattern = `^\d{7,20}$`
	emailPattern       = `^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`
)

var (
	studentEnrollmentStatusMap = map[string]string{
		"1": constant.StudentEnrollmentStatusPotential,
		"2": constant.StudentEnrollmentStatusEnrolled,
		"3": constant.StudentEnrollmentStatusWithdrawn,
		"4": constant.StudentEnrollmentStatusGraduated,
		"5": constant.StudentEnrollmentStatusLOA,
		"6": constant.StudentEnrollmentStatusTemporary,
		"7": constant.StudentEnrollmentStatusNonPotential,
	}
)

type ParentProfile interface {
	GetTagIds() []string
}

func UpdateUserEmail(ctx context.Context, authClient multitenant.TenantClient, userID string, email string) error {
	_, err := authClient.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	userToUpdate := (&auth.UserToUpdate{}).Email(email)

	_, err = authClient.LegacyUpdateUser(ctx, userID, userToUpdate)
	if err != nil {
		return errors.Wrapf(err, "overrideUserEmail userID=%v, email=%v", userID, email)
	}
	return nil
}

func countElementOccurrencesInSlice(arr []int) map[int]int {
	keys := make(map[int]int)
	for _, entry := range arr {
		if _, value := keys[entry]; !value {
			keys[entry] = 1
		} else {
			keys[entry]++
		}
	}

	return keys
}

func findUniqueValue(keys map[int]int, lenArrCheck int) int {
	min := lenArrCheck
	uniqueValue := 0

	// get the element that appears least in the array
	for key, value := range keys {
		if value < min {
			min = value
			uniqueValue = key
		}
	}

	return uniqueValue
}

func overrideUserPassword(ctx context.Context, authClient multitenant.TenantClient, userID string, password string) error {
	_, err := authClient.GetUser(ctx, userID)
	if err != nil {
		if auth.IsUserNotFound(err) {
			err = user.ErrUserNotFound
		}
		return err
	}

	userToUpdate := (&auth.UserToUpdate{}).Password(password)

	_, err = authClient.LegacyUpdateUser(ctx, userID, userToUpdate)
	if err != nil {
		return errors.Wrap(err, "overrideUserPassword()")
	}
	return nil
}

func prependBeforeColumn(currentHeaders, currentValues, anchorHeader, headerToPrepend, valueToPrepend string) (string, string) {
	separatedChar := ","

	headers := strings.Split(currentHeaders, separatedChar)
	values := strings.Split(currentValues, separatedChar)

	resultHeaders := make([]string, len(headers)+1, cap(headers)+1)
	resultValues := make([]string, len(values)+1, cap(values)+1)

	for index, header := range headers {
		if header == anchorHeader {
			// append headers
			copy(resultHeaders, headers[:index])
			resultHeaders[index] = headerToPrepend
			copy(resultHeaders[index+1:], headers[index:])

			// append values
			copy(resultValues, values[:index])
			resultValues[index] = valueToPrepend
			copy(resultValues[index+1:], values[index:])

			// after appending, we return directly
			return strings.Join(resultHeaders, separatedChar), strings.Join(resultValues, separatedChar)
		}
	}

	// after loop, we can not find position to append, return default value
	return strings.Join(headers, separatedChar), strings.Join(values, separatedChar)
}

func convertDataTemplateCSVToBase64(data string) string {
	encodedStr := base64.StdEncoding.EncodeToString([]byte(data))

	return encodedStr
}

func signCtx(ctx context.Context) context.Context {
	headers, ok := metadata.FromIncomingContext(ctx)
	var pkg, token, version string
	if ok {
		pkg = headers["pkg"][0]
		token = headers["token"][0]
		version = headers["version"][0]
	}
	return metadata.AppendToOutgoingContext(ctx, "pkg", pkg, "version", version, "token", token)
}

func toSchoolPb(schoolIn *entity.School) *pb.School {
	if schoolIn == nil {
		return nil
	}

	schoolOut := &pb.School{
		Id:      schoolIn.ID.Int,
		Name:    schoolIn.Name.String,
		Country: pbc.Country(pbc.Country_value[schoolIn.Country.String]),
	}

	if schoolIn.City == nil {
		schoolOut.City = &pb.City{
			Id: schoolIn.CityID.Int,
		}
	} else {
		schoolOut.City = toCityPb(schoolIn.City)
	}

	if schoolIn.District == nil {
		schoolOut.District = &pb.District{
			Id: schoolIn.DistrictID.Int,
		}
	} else {
		schoolOut.District = toDistrictPb(schoolIn.District)
	}

	if schoolIn.Point.Status == pgtype.Present {
		schoolOut.Point = &pb.Point{
			Lat:  schoolIn.Point.P.X,
			Long: schoolIn.Point.P.Y,
		}
	}

	return schoolOut
}

func toCityPb(cityIn *entity.City) *pb.City {
	if cityIn == nil {
		return nil
	}

	return &pb.City{
		Id:      cityIn.ID.Int,
		Name:    cityIn.Name.String,
		Country: pbc.Country(pbc.Country_value[cityIn.Country.String]),
	}
}

func toDistrictPb(districtIn *entity.District) *pb.District {
	if districtIn == nil {
		return nil
	}

	districtOut := &pb.District{
		Id:      districtIn.ID.Int,
		Name:    districtIn.Name.String,
		Country: pbc.Country(pbc.Country_value[districtIn.Country.String]),
	}

	if districtIn.City == nil {
		districtOut.City = &pb.City{
			Id: districtIn.CityID.Int,
		}
	} else {
		districtOut.City = toCityPb(districtIn.City)
	}

	return districtOut
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

func MatchingRegex(pattern string, str string) error {
	match, err := regexp.MatchString(pattern, str)
	if err != nil {
		return fmt.Errorf("error regexp.MatchString: %v", err)
	}
	if !match {
		return fmt.Errorf("error regexp.MatchString: doesn't match")
	}
	return nil
}

func toUserAccessPathEntities(locations []*domain.Location, userIDs []string) ([]*entity.UserAccessPath, error) {
	userAccessPaths := []*entity.UserAccessPath{}

	for _, userID := range userIDs {
		for _, location := range locations {
			userAccessPathEnt := &entity.UserAccessPath{}
			database.AllNullEntity(userAccessPathEnt)

			if err := multierr.Combine(
				userAccessPathEnt.UserID.Set(userID),
				userAccessPathEnt.LocationID.Set(location.LocationID),
			); err != nil {
				return nil, errors.Wrap(err, "multierr.Combine")
			}

			userAccessPaths = append(userAccessPaths, userAccessPathEnt)
		}
	}

	return userAccessPaths, nil
}

func getTagIDsFromParentProfiles(req interface{}) []string {
	tagIDs := []string{}

	switch pbProfile := req.(type) {
	case *pb.CreateParentsAndAssignToStudentRequest:
		for _, profile := range pbProfile.GetParentProfiles() {
			tagIDs = append(tagIDs, profile.GetTagIds()...)
		}
	case *pb.UpdateParentsAndFamilyRelationshipRequest:
		for _, profile := range pbProfile.GetParentProfiles() {
			tagIDs = append(tagIDs, profile.GetTagIds()...)
		}
	}

	return tagIDs
}
