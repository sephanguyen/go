package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	libdatabase "github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var ERPEnrollmentStatus = []string{
	upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
	upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
	upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL.String(),
}

func (service *DomainStudent) getConfiguration(ctx context.Context) (*mpb.Configuration, error) {
	zapLogger := ctxzap.Extract(ctx)
	getConfigReq := &mpb.GetConfigurationByKeyRequest{Key: constant.KeyEnrollmentStatusHistoryConfig}
	resp, err := service.ConfigurationClient.GetConfigurationByKey(ctx, getConfigReq)
	if err != nil {
		zapLogger.Error(
			"error when gettingconfig",
			zap.Error(err),
			zap.String("Function", "getConfiguration"),
		)
		if strings.Contains(err.Error(), "wrong token") {
			resp = &mpb.GetConfigurationByKeyResponse{
				Configuration: &mpb.Configuration{
					ConfigKey:   constant.KeyEnrollmentStatusHistoryConfig,
					ConfigValue: constant.ConfigValueOff,
				},
			}
		} else {
			return nil, errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, "service.ConfigurationClient.GetConfigurationByKey"),
			}
		}
	}

	// Get configuration to detect logic lms or erp
	configurationResp := resp.GetConfiguration()
	if configurationResp == nil {
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  fmt.Errorf("not found config"),
		}
	}
	return configurationResp, nil
}

func validateEnrollmentStatusCreateStudent(enrollmentStatus entity.DomainEnrollmentStatusHistory, hasActivatedEnrollment bool, idx int) error {
	enrollmentStatusStr := enrollmentStatus.EnrollmentStatus().String()
	if enrollmentStatusStr == upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String() && !hasActivatedEnrollment {
		return errcode.Error{
			Code:      errcode.InvalidData,
			FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.enrollment_status", idx),
			Index:     idx,
		}
	}
	return nil
}

func (service *DomainStudent) validateEnrollmentStatusUpdateStudent(ctx context.Context, enrollmentStatus entity.DomainEnrollmentStatusHistory, idx int, isOrderFlow bool) error {
	zapLogger := ctxzap.Extract(ctx)
	// There is only one currentEnrollmentStatus of one student in one location at a time
	userID := enrollmentStatus.UserID().String()
	locationID := enrollmentStatus.LocationID().String()

	currentEnrollmentStatus, err := service.EnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID(ctx, service.DB, userID, locationID, true)
	if err != nil {
		zapLogger.Error(
			"cannot get current enrollment status histories",
			zap.Error(err),
			zap.String("Repo", "EnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID"),
			zap.String("studentID", userID),
			zap.String("locationID", locationID),
		)
		return errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID"),
		}
	}
	latestStatus, err := service.EnrollmentStatusHistoryRepo.GetLatestEnrollmentStudentOfLocation(ctx, service.DB, userID, locationID)
	if err != nil {
		zapLogger.Error(
			"cannot get latest enrollment status history",
			zap.Error(err),
			zap.String("Repo", "EnrollmentStatusHistoryRepo.GetLatestEnrollmentStudentOfLocation"),
			zap.String("studentID", userID),
			zap.String("locationID", locationID),
		)
		return errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.GetLatestEnrollmentStudentOfLocation"),
		}
	}
	if len(currentEnrollmentStatus) == 0 {
		currentEnrollmentStatus = latestStatus
		if len(currentEnrollmentStatus) == 0 {
			zapLogger.Error(
				"current or last enrollment status history cannot be empty",
				zap.String("studentID", userID),
				zap.String("locationID", locationID),
			)
			return errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: "location_id",
				Err:       errors.New("cannot get current and last with this location_id"),
			}
		}
	}
	// non erp status can not change to others at order flow
	if isOrderFlow && !golibs.InArrayString(currentEnrollmentStatus[0].EnrollmentStatus().String(), ERPEnrollmentStatus) {
		return errcode.Error{
			Code:      errcode.InvalidData,
			FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.enrollment_status", idx),
			Index:     idx,
		}
	}
	if err := validateEntityEnrollmentStatusHistory(currentEnrollmentStatus[0], enrollmentStatus, latestStatus[0], idx); err != nil {
		zapLogger.Error(
			"validateEnrollmentStatusUpdateStudent.validateEntityEnrollmentStatusHistory",
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (service *DomainStudent) setEnrollmentStatusHistories(organization valueobj.HasOrganizationID, studentsToUpsert ...aggregate.DomainStudent) {
	for _, student := range studentsToUpsert {
		if len(student.EnrollmentStatusHistories) == 0 {
			continue
		}

		enrollmentStatusHistoryWillBeDelegated := make([]entity.EnrollmentStatusHistoryWillBeDelegated, 0, len(student.EnrollmentStatusHistories))
		for _, enrollmentStatusHistory := range student.EnrollmentStatusHistories {
			enrollmentStatusHistoryWillBeDelegated = append(enrollmentStatusHistoryWillBeDelegated, entity.EnrollmentStatusHistoryWillBeDelegated{
				EnrollmentStatusHistory: enrollmentStatusHistory,
				HasLocationID:           enrollmentStatusHistory,
				HasUserID:               student,
				HasOrganizationID:       organization,
			})
		}

		for j := range enrollmentStatusHistoryWillBeDelegated {
			student.EnrollmentStatusHistories[j] = &enrollmentStatusHistoryWillBeDelegated[j]
		}
	}
}

func (service *DomainStudent) upsertEnrollmentStatusHistories(ctx context.Context, db libdatabase.QueryExecer, studentsToUpsert ...aggregate.DomainStudent) error {
	zapLogger := ctxzap.Extract(ctx)
	for _, student := range studentsToUpsert {
		// Skip when updating student without EnrollmentStatusHistories
		if len(student.EnrollmentStatusHistories) == 0 {
			continue
		}
		for _, enrollmentStatusHistory := range student.EnrollmentStatusHistories {
			existedEnrollmentStatusHistory, err := service.EnrollmentStatusHistoryRepo.GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate(ctx, service.DB, enrollmentStatusHistory)
			if err != nil {
				zapLogger.Error(
					"cannot get existing enrollment status histories when upsert",
					zap.Error(err),
					zap.String("LocationID", enrollmentStatusHistory.LocationID().String()),
					zap.String("EnrollmentStatus", enrollmentStatusHistory.EnrollmentStatus().String()),
					zap.String("UserID", enrollmentStatusHistory.UserID().String()),
				)
				return errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate")
			}
			// Skip if there is a record in BD
			if len(existedEnrollmentStatusHistory) != 0 {
				continue
			}
			if err := service.upsertEnrollmentStatusHistory(ctx, db, enrollmentStatusHistory); err != nil {
				zapLogger.Error(
					"upsertEnrollmentStatusHistories.upsertEnrollmentStatusHistory",
					zap.Error(err),
				)
				return err
			}
		}
	}
	return nil
}

func (service *DomainStudent) hasActivatedEnrollmentStatusHistory(ctx context.Context, db libdatabase.QueryExecer, reqEnrollmentStatusHistories entity.DomainEnrollmentStatusHistories, studentID string) (bool, error) {
	zapLogger := ctxzap.Extract(ctx)
	currentEnrollmentStatusHistoriesOfStudent, err := service.EnrollmentStatusHistoryRepo.GetByStudentID(ctx, db, studentID, true)
	if err != nil {
		zapLogger.Error(
			"cannot get enrollment status histories",
			zap.Error(err),
			zap.String("Repo", "EnrollmentStatusHistoryRepo.GetByStudentID"),
			zap.String("studentId", studentID),
		)
		return false, errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.GetByStudentID")
	}
	for _, enrollmentStatusHistory := range append(currentEnrollmentStatusHistoriesOfStudent, reqEnrollmentStatusHistories...) {
		if enrollmentStatusHistory.EnrollmentStatus().String() != upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String() {
			return true, nil
		}
	}
	return false, nil
}

func (service *DomainStudent) upsertEnrollmentStatusHistory(ctx context.Context, db libdatabase.QueryExecer, enrollmentStatusReq entity.DomainEnrollmentStatusHistory) error {
	zapLogger := ctxzap.Extract(ctx)
	// check location of student is exists (new or not)
	enrollmentStatusHistoriesOfStudent, err := service.EnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID(ctx, db, enrollmentStatusReq.UserID().String(), enrollmentStatusReq.LocationID().String(), false)
	enrollReqStr := enrollmentStatusReq.EnrollmentStatus().String()
	if err != nil {
		zapLogger.Error(
			"cannot get enrollment status histories",
			zap.Error(err),
			zap.String("Repo", "EnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID"),
			zap.String("studentID", enrollmentStatusReq.UserID().String()),
			zap.String("locationID", enrollmentStatusReq.LocationID().String()),
		)
		return errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID"),
		}
	}
	if len(enrollmentStatusHistoriesOfStudent) == 0 { // Create logic
		if err := service.EnrollmentStatusHistoryRepo.Create(ctx, db, enrollmentStatusReq); err != nil {
			zapLogger.Error(
				"cannot create enrollment status history",
				zap.Error(err),
				zap.String("Repo", "EnrollmentStatusHistoryRepo.Create"),
			)
			return errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.Create"),
			}
		}
	} else { // Update logic
		currentEnrollmentStatusHistories, err := service.EnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID(ctx, db, enrollmentStatusReq.UserID().String(), enrollmentStatusReq.LocationID().String(), true)
		if err != nil {
			zapLogger.Error(
				"cannot get enrollment status history",
				zap.Error(err),
				zap.String("Repo", "EnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID"),
				zap.String("studentID", enrollmentStatusReq.UserID().String()),
				zap.String("locationID", enrollmentStatusReq.LocationID().String()),
			)
			return errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID"),
			}
		}

		latestEnrollmentStatusHistories, err := service.EnrollmentStatusHistoryRepo.GetLatestEnrollmentStudentOfLocation(ctx, db, enrollmentStatusReq.UserID().String(), enrollmentStatusReq.LocationID().String())
		if err != nil {
			zapLogger.Error(
				"cannot get latest enrollment status history",
				zap.Error(err),
				zap.String("Repo", "EnrollmentStatusHistoryRepo.GetLatestEnrollmentStudentOfLocation"),
				zap.String("studentID", enrollmentStatusReq.UserID().String()),
				zap.String("locationID", enrollmentStatusReq.LocationID().String()),
			)
			return errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.GetLatestEnrollmentStudentOfLocation"),
			}
		}

		if len(currentEnrollmentStatusHistories) == 0 {
			if len(latestEnrollmentStatusHistories) == 0 {
				return nil
			}
			currentEnrollmentStatusHistories = latestEnrollmentStatusHistories
		}

		currentEnrollmentStatusHistory := currentEnrollmentStatusHistories[0]
		currentEnrollStr := currentEnrollmentStatusHistories[0].EnrollmentStatus().String()

		latestEnrollmentStatusHistory := latestEnrollmentStatusHistories[0]
		latestEnrollStr := latestEnrollmentStatusHistory.EnrollmentStatus().String()

		if currentEnrollStr == enrollReqStr {
			// Update end_date of status temporary
			if enrollReqStr == upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String() {
				if err := service.EnrollmentStatusHistoryRepo.Update(ctx, db,
					currentEnrollmentStatusHistory,
					enrollmentStatusReq,
				); err != nil {
					zapLogger.Error(
						"cannot update enrollment status histories",
						zap.Error(err),
						zap.String("Repo", "EnrollmentStatusHistoryRepo.Update"),
					)
					return errcode.Error{
						Code: errcode.InternalError,
						Err:  errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.Update"),
					}
				}
			}
			return nil
		}

		endDate := enrollmentStatusReq.StartDate().Time().Add(-1 * time.Second)
		if enrollmentStatusReq.StartDate().Time().IsZero() {
			endDate = time.Now().Add(-1 * time.Second)
		}

		if err := service.EnrollmentStatusHistoryRepo.DeactivateEnrollmentStatus(ctx, db,
			currentEnrollmentStatusHistory,
			endDate,
		); err != nil {
			zapLogger.Error(
				"cannot deactivate enrollment status history",
				zap.Error(err),
				zap.String("Repo", "EnrollmentStatusHistoryRepo.DeactivateEnrollmentStatus"),
			)
			return errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.DeactivateEnrollmentStatus"),
			}
		}

		// User only want to update last enrollment
		if latestEnrollStr != currentEnrollStr {
			if err := service.EnrollmentStatusHistoryRepo.Update(ctx, db, latestEnrollmentStatusHistory, enrollmentStatusReq); err != nil {
				zapLogger.Error(
					"cannot update enrollment status history",
					zap.Error(err),
					zap.String("Repo", "EnrollmentStatusHistoryRepo.Update"),
				)
				return errcode.Error{
					Code: errcode.InternalError,
					Err:  errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.Update"),
				}
			}
			return nil
		}
		if err := service.EnrollmentStatusHistoryRepo.Create(ctx, db, enrollmentStatusReq); err != nil {
			zapLogger.Error(
				"cannot create enrollment status history",
				zap.Error(err),
				zap.String("Repo", "EnrollmentStatusHistoryRepo.Create"),
			)
			return errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.Create"),
			}
		}
		return nil
	}
	return nil
}

func (service *DomainStudent) updateEnrollmentStatusHistories(ctx context.Context, db libdatabase.QueryExecer, studentsToUpdate aggregate.DomainStudents) error {
	if len(studentsToUpdate) == 0 {
		return nil
	}
	existingEnrollmentStatusHistories, err := service.EnrollmentStatusHistoryRepo.GetByStudentIDs(ctx, db, studentsToUpdate.StudentIDs())
	if err != nil {
		return err
	}
	for _, student := range studentsToUpdate {
		// Skip when updating student without EnrollmentStatusHistories
		if len(student.EnrollmentStatusHistories) == 0 {
			continue
		}
		for _, reqEnrollmentStatusHistory := range student.EnrollmentStatusHistories {
			existedEnrollmentStatusHistory := existingEnrollmentStatusHistories.GetExactly(reqEnrollmentStatusHistory)
			// Skip if there is a the same record enrollment status histories in BD
			if existedEnrollmentStatusHistory != nil {
				continue
			}

			enrollReqStr := reqEnrollmentStatusHistory.EnrollmentStatus().String()
			activatedEnrollmentStatusHistory := existingEnrollmentStatusHistories.GetActivatedByUserIDLocationID(reqEnrollmentStatusHistory.UserID(), reqEnrollmentStatusHistory.LocationID())
			latestEnrollmentStatusHistory := existingEnrollmentStatusHistories.GetLatestByUserIDLocationID(reqEnrollmentStatusHistory.UserID(), reqEnrollmentStatusHistory.LocationID())

			if activatedEnrollmentStatusHistory == nil {
				if latestEnrollmentStatusHistory == nil {
					continue
				}
				activatedEnrollmentStatusHistory = latestEnrollmentStatusHistory
			}
			activatedEnrollStr := activatedEnrollmentStatusHistory.EnrollmentStatus().String()
			latestEnrollStr := latestEnrollmentStatusHistory.EnrollmentStatus().String()

			if activatedEnrollStr == enrollReqStr {
				// Update end_date of status temporary
				if enrollReqStr == constant.StudentEnrollmentStatusTemporary {
					if err := service.EnrollmentStatusHistoryRepo.Update(ctx, db,
						activatedEnrollmentStatusHistory,
						reqEnrollmentStatusHistory,
					); err != nil {
						return err
					}
				}
				return nil
			}

			endDate := reqEnrollmentStatusHistory.StartDate().Time().Add(-1 * time.Second)
			if reqEnrollmentStatusHistory.StartDate().Time().IsZero() {
				endDate = time.Now().Add(-1 * time.Second)
			}

			// Deactivate current enrollment status to create new one
			if err := service.EnrollmentStatusHistoryRepo.DeactivateEnrollmentStatus(ctx, db,
				activatedEnrollmentStatusHistory,
				endDate,
			); err != nil {
				return err
			}
			// User only want to update last enrollment
			if latestEnrollStr != activatedEnrollStr {
				return service.EnrollmentStatusHistoryRepo.Update(ctx, db, latestEnrollmentStatusHistory, reqEnrollmentStatusHistory)
			}
			return service.EnrollmentStatusHistoryRepo.Create(ctx, db, reqEnrollmentStatusHistory)
		}
	}
	return nil
}

func validateEntityEnrollmentStatusHistory(currentEnrollmentStatus entity.DomainEnrollmentStatusHistory, reqEnrollmentStatus entity.DomainEnrollmentStatusHistory, latestEnrollmentStatus entity.DomainEnrollmentStatusHistory, idx int) error {
	currentStatus := currentEnrollmentStatus.EnrollmentStatus().String()
	reqStatus := reqEnrollmentStatus.EnrollmentStatus().String()
	reqEnrollmentStatusStartDateTruncated := reqEnrollmentStatus.StartDate().Time().Truncate(time.Second)
	reqEnrollmentStatusEndDateTruncated := reqEnrollmentStatus.EndDate().Time().Truncate(time.Second)

	currentEnrollmentStatusStartDateTruncated := currentEnrollmentStatus.StartDate().Time().Truncate(time.Second)
	currentEnrollmentStatusEndDateTruncated := currentEnrollmentStatus.EndDate().Time().Truncate(time.Second)
	switch currentStatus {
	// status potential can change to any status
	case upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String():
		break
	// status non-potential can't change to any status
	case upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL.String():
		if reqStatus != upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL.String() {
			return errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.enrollment_status", idx),
				Index:     idx,
			}
		}
	// status temporary can't change end date when having a status in future
	case upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String():
		if reqStatus != currentStatus {
			break
		}
		latestStatus := latestEnrollmentStatus.EnrollmentStatus().String()
		if latestStatus != currentStatus && !currentEnrollmentStatusEndDateTruncated.Equal(reqEnrollmentStatusEndDateTruncated) {
			return errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.end_date", idx),
				Index:     idx,
			}
		}
	}

	if reqEnrollmentStatusStartDateTruncated.IsZero() {
		return nil
	}

	// Skip if nothing change
	if currentStatus == reqStatus {
		if !currentEnrollmentStatusStartDateTruncated.Equal(reqEnrollmentStatusStartDateTruncated) {
			return errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.start_date", idx),
				Index:     idx,
			}
		}
		return nil
	}

	if currentEnrollmentStatusStartDateTruncated.After(reqEnrollmentStatusStartDateTruncated) {
		return errcode.Error{
			Code:      errcode.InvalidData,
			FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.start_date", idx),
			Index:     idx,
		}
	}

	// start_date is the same, but latestStatus != reqStatus
	if currentEnrollmentStatusStartDateTruncated.Equal(reqEnrollmentStatusStartDateTruncated) {
		return errcode.Error{
			Code:      errcode.InvalidData,
			FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.start_date", idx),
			Index:     idx,
		}
	}

	return nil
}

func validateEnrollmentStatusCreateRequestFromOrder(enrollmentStatus string, index int) error {
	for _, validEnrollment := range ERPEnrollmentStatus {
		if enrollmentStatus == validEnrollment {
			return nil
		}
	}

	return errcode.Error{
		Code:      errcode.InvalidData,
		FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.enrollment_status", index),
	}
}

func (service *DomainStudent) validateEnrollmentStatusHistories(ctx context.Context, students ...aggregate.DomainStudent) error {
	zapLogger := ctxzap.Extract(ctx)
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		zapLogger.Error(
			"validateEnrollmentStatusHistories.OrganizationFromContext",
			zap.Error(err),
		)
		return err
	}

	configEnrollmentUpdateManual := ""
	if unleash.IsFeatureUsingMasterReplicatedTable(service.UnleashClient, service.Env, organization) {
		config, err := service.InternalConfigurationRepo.GetByKey(ctx, service.DB, constant.KeyEnrollmentStatusHistoryConfig)
		if err != nil {
			zapLogger.Error(
				"validateEnrollmentStatusHistories.getConfiguration",
				zap.Error(err),
			)
			return err
		}
		configEnrollmentUpdateManual = config.ConfigValue().String()
	} else {
		configurationResp, err := service.getConfiguration(ctx)
		if err != nil {
			zapLogger.Error(
				"validateEnrollmentStatusHistories.getConfiguration",
				zap.Error(err),
			)
			return err
		}
		configEnrollmentUpdateManual = configurationResp.GetConfigValue()
	}
	isOrderFlow := configEnrollmentUpdateManual == constant.ConfigValueOff
	for idx, student := range students {
		userID := student.UserID().String()
		hasActivatedEnrollment, err := service.hasActivatedEnrollmentStatusHistory(ctx, service.DB, student.EnrollmentStatusHistories, userID)
		if err != nil {
			zapLogger.Error(
				"validateEnrollmentStatusHistories.hasActivatedEnrollmentStatusHistory",
				zap.Error(err),
			)
			return errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, "service.hasActivatedEnrollmentStatusHistory"),
			}
		}

		for _, enrollmentStatusHistory := range student.EnrollmentStatusHistories {
			locationID := enrollmentStatusHistory.LocationID().String()
			enrollmentStatusStr := enrollmentStatusHistory.EnrollmentStatus().String()
			// validate config
			if isOrderFlow {
				existingEnrollmentStatusHistories, err := service.EnrollmentStatusHistoryRepo.GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate(ctx, service.DB, enrollmentStatusHistory)
				if err != nil {
					zapLogger.Error(
						"cannot get existing enrollment status histories in order flow",
						zap.Error(err),
						zap.String("LocationID", enrollmentStatusHistory.LocationID().String()),
						zap.String("EnrollmentStatus", enrollmentStatusHistory.EnrollmentStatus().String()),
						zap.String("UserID", enrollmentStatusHistory.UserID().String()),
						zap.String("StartDate", enrollmentStatusHistory.StartDate().Time().String()),
					)
					return errcode.Error{
						Code: errcode.InternalError,
						Err:  errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate"),
					}
				}
				if len(existingEnrollmentStatusHistories) != 0 {
					return nil
				}
				if err := validateEnrollmentStatusCreateRequestFromOrder(enrollmentStatusStr, idx); err != nil {
					zapLogger.Error(
						"validateEnrollmentStatusHistories.validateEnrollmentStatusCreateRequestFromOrder",
						zap.Error(err),
					)
					return err
				}
			}
			// validate upsert student
			enrollmentStatusHistories, err := service.EnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID(ctx, service.DB, userID, locationID, false)
			if err != nil {
				zapLogger.Error(
					"cannot get enrollment status histories",
					zap.Error(err),
					zap.String("Repo", "EnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID"),
					zap.String("studentID", userID),
					zap.String("locationID", locationID),
				)
				return errcode.Error{
					Code: errcode.InternalError,
					Err:  errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID"),
				}
			}

			if len(enrollmentStatusHistories) == 0 {
				// validate create student
				if err := validateEnrollmentStatusCreateStudent(enrollmentStatusHistory, hasActivatedEnrollment, idx); err != nil {
					zapLogger.Error(
						"validateEnrollmentStatusHistories.validateEnrollmentStatusCreateStudent",
						zap.Error(err),
					)
					return err
				}
			} else {
				existedEnrollmentStatusHistory, err := service.EnrollmentStatusHistoryRepo.GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate(ctx, service.DB, enrollmentStatusHistory)
				if err != nil {
					zapLogger.Error(
						"cannot get existing enrollment status histories in LMS flow",
						zap.Error(err),
						zap.String("LocationID", enrollmentStatusHistory.LocationID().String()),
						zap.String("EnrollmentStatus", enrollmentStatusHistory.EnrollmentStatus().String()),
						zap.String("UserID", enrollmentStatusHistory.UserID().String()),
					)
					return errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate")
				} // Skip if there is a record in BD
				if len(existedEnrollmentStatusHistory) != 0 {
					continue
				}
				// validate update student
				if err := service.validateEnrollmentStatusUpdateStudent(ctx, enrollmentStatusHistory, idx, isOrderFlow); err != nil {
					zapLogger.Error(
						"validateEnrollmentStatusHistories.validateEnrollmentStatusUpdateStudent",
						zap.Error(err),
					)
					return err
				}
			}
		}
	}

	return nil
}

func validateEnrollmentStatusHistoriesBeforeCreating(ctx context.Context, studentsToUpsert ...aggregate.DomainStudent) error {
	zapLogger := ctxzap.Extract(ctx)
	location, err := time.LoadLocation(constant.JpTimeZone)
	if err != nil {
		zapLogger.Error(
			"can not load location of timezone",
			zap.Error(err),
			zap.String("Timezone", constant.JpTimeZone),
		)
		return err
	}
	for idx, student := range studentsToUpsert {
		// Potential/Temporary/Non-Potential status start date can not be after current date
		for j, enrollmentStatus := range student.EnrollmentStatusHistories {
			if enrollmentStatus.EnrollmentStatus().IsEmpty() {
				return errcode.Error{
					// There is an enrollment status that has been mapped with an invalid enum, resulting in an empty string for the status.
					// Therefore, we need to display an error message stating that the data is invalid.
					Code:      errcode.InvalidData,
					FieldName: fmt.Sprintf("students[%d].enrollment_status_histories[%d].enrollment_status", student.IndexAttr, j),
					Index:     idx,
				}
			}

			startDateTime := enrollmentStatus.StartDate().Time()
			statusStr := enrollmentStatus.EnrollmentStatus().String()
			endDateTime := enrollmentStatus.EndDate().Time()
			if !endDateTime.IsZero() {
				if endDateTime.Before(startDateTime) || endDateTime.Equal(startDateTime) {
					return errcode.Error{
						Code:      errcode.InvalidData,
						FieldName: fmt.Sprintf("students[%d].enrollment_status_histories[%d].end_date", student.IndexAttr, j),
						Index:     idx,
					}
				}
			}
			if startDateTime.IsZero() || !golibs.InArrayString(statusStr, ERPEnrollmentStatus) {
				continue
			}
			roundedStartDateTime := utils.TruncateTimeToStartOfDay(startDateTime.In(location))
			roundedCurrentDateTime := utils.TruncateTimeToStartOfDay(time.Now().In(location))
			if roundedStartDateTime.After(roundedCurrentDateTime) {
				return errcode.Error{
					Code:      errcode.InvalidData,
					FieldName: fmt.Sprintf("students[%d].enrollment_status_histories[%d].start_date", student.IndexAttr, j),
					Index:     idx,
				}
			}
		}
		switch {
		// Skip if case update student
		case student.UserID().String() != "", len(student.EnrollmentStatusHistories) != 0:
			continue
		case len(student.EnrollmentStatusHistories) == 0:
			if len(student.UserAccessPaths.LocationIDs()) == 0 {
				return errcode.Error{
					Code:      errcode.MissingMandatory,
					FieldName: fmt.Sprintf("students[%d].locations", student.IndexAttr),
					Index:     idx,
				}
			}
		}
	}
	return nil
}

type StudentActivationStatusManager struct {
	EnrollmentStatusHistoryRepo interface {
		GetInactiveAndActiveStudents(ctx context.Context, db libdatabase.QueryExecer, studentIDs, enrollmentStatuses []string) ([]entity.DomainEnrollmentStatusHistory, error)
	}
	UserRepo interface {
		UpdateActivation(ctx context.Context, db libdatabase.QueryExecer, users entity.Users) error
	}
}

// Temporarily leave this struct here, refactor later when find the best way to map entity to another entity
type DomainUserImpl struct {
	entity.EmptyUser

	UserIDAttr        field.String
	DeactivatedAtAttr field.Time
}

func (u DomainUserImpl) UserID() field.String {
	return u.UserIDAttr
}

func (u DomainUserImpl) DeactivatedAt() field.Time {
	return u.DeactivatedAtAttr
}

func (manager *StudentActivationStatusManager) DeactivateAndReactivateStudents(ctx context.Context, db libdatabase.QueryExecer, studentIDs, enrollmentStatuses []string) error {
	zapLogger := ctxzap.Extract(ctx)
	now := time.Now()
	enrollmentStatusHistories, err := manager.EnrollmentStatusHistoryRepo.GetInactiveAndActiveStudents(ctx, db, studentIDs, enrollmentStatuses)
	if err != nil {
		zapLogger.Error(
			"cannot get enrollment status histories",
			zap.Error(err),
			zap.String("Manager", "DeactivateAndReactivateStudentManager"),
			zap.String("Repo", "enrollmentStatusHistoryRepo.GetInactiveAndActiveStudents"),
			zap.Strings("studentIDs", studentIDs),
		)
		return errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.GetInactiveAndActiveStudents"),
		}
	}
	zapLogger.Warn(
		"--end insert DeactivateAndReactivateStudents-GetInactiveAndActiveStudents--",
		zap.Int64("time-end", time.Since(now).Milliseconds()),
	)
	users := EnrollmentStatusHistoryStartDateToUserDeactivatedAt(enrollmentStatusHistories)
	now = time.Now()
	if err := manager.UserRepo.UpdateActivation(ctx, db, users); err != nil {
		return err
	}
	zapLogger.Warn(
		"--end insert DeactivateAndReactivateStudents-UpdateActivation--",
		zap.Int64("time-end", time.Since(now).Milliseconds()),
	)
	return nil
}

func EnrollmentStatusHistoryStartDateToUserDeactivatedAt(enrollmentStatusHistories entity.DomainEnrollmentStatusHistories) entity.Users {
	domainUsers := make(entity.Users, 0, len(enrollmentStatusHistories))
	for _, enrollmentStatus := range enrollmentStatusHistories {
		domainUsers = append(domainUsers, &DomainUserImpl{
			UserIDAttr:        enrollmentStatus.UserID(),
			DeactivatedAtAttr: enrollmentStatus.StartDate(),
		})
	}
	return domainUsers
}

type OrderFlowEnrollmentStatusManager interface {
	HandleEnrollmentStatusUpdate(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error)
	HandleForNewLocation(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error)
	HandleExistedLocations(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error)
	HandleVoidEnrollmentStatus(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error)
}

var _ OrderFlowEnrollmentStatusManager = (*HandelOrderFlowEnrollmentStatus)(nil)

type HandelOrderFlowEnrollmentStatus struct {
	Logger                            *zap.Logger
	DomainEnrollmentStatusHistoryRepo DomainEnrollmentStatusHistoryRepo
	DomainUserAccessPathRepo          DomainUserAccessPathRepo

	SyncEnrollmentStatusHistory     func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog, enrollmentStatus string) error
	DeactivateAndReactivateStudents func(ctx context.Context, db libdatabase.Ext, studentIDs []string) error
}

func (s *HandelOrderFlowEnrollmentStatus) HandleForNewLocation(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
	defaultStatusForNewLocation := constant.StudentEnrollmentStatusPotential
	err := s.SyncEnrollmentStatusHistory(ctx, db, req, defaultStatusForNewLocation)
	if err != nil {
		s.Logger.Error("HandelOrderFlowEnrollmentStatus.HandleForNewLocation",
			zap.String("EnrollmentStatus", defaultStatusForNewLocation),
			zap.String("OrderStatus", req.OrderStatus),
			zap.String("OrderType", req.OrderType),
			zap.Error(err),
		)
		// return true for retry
		return true, err
	}
	return false, nil
}

func (s *HandelOrderFlowEnrollmentStatus) HandleExistedLocations(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
	// Get current enrollment status (start date < current date , end date > current date)
	currentEnrollmentStatus, err := s.DomainEnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID(ctx, db, req.StudentID, req.LocationID, true)
	if err != nil {
		s.Logger.Error("HandelOrderFlowEnrollmentStatus.HandleExistedLocations",
			zap.String("StudentID", req.StudentID),
			zap.String("LocationID", req.LocationID),
			zap.String("OrderStatus", req.OrderStatus),
			zap.String("OrderType", req.OrderType),
			zap.Error(err),
		)
		// return true to retry
		return true, err
	}
	if len(currentEnrollmentStatus) != 0 {
		if currentEnrollmentStatus[0].EnrollmentStatus().String() == constant.StudentEnrollmentStatusTemporary {
			// set end date = start date - 1 second
			endDate := req.StartDate.Add(-1 * time.Second)
			err := s.DomainEnrollmentStatusHistoryRepo.
				DeactivateEnrollmentStatus(
					ctx, db,
					currentEnrollmentStatus[0],
					endDate,
				)
			if err != nil {
				s.Logger.Error("HandelOrderFlowEnrollmentStatus.HandleExistedLocations",
					zap.String("StartDate", req.StartDate.String()),
					zap.String("CurrentEnrollmentStatus", currentEnrollmentStatus[0].EnrollmentStatus().String()),
					zap.String("OrderStatus", req.OrderStatus),
					zap.String("OrderType", req.OrderType),
					zap.Error(err),
				)
				// return true to retry
				return true, err
			}

			if currentEnrollmentStatus[0].EndDate().Ptr().BeforeTime(req.EndDate) &&
				currentEnrollmentStatus[0].EndDate().Ptr().Status() != field.StatusNull {
				req.EndDate = currentEnrollmentStatus[0].EndDate().Time()
			}
			if err := s.SyncEnrollmentStatusHistory(ctx, db, req, constant.StudentEnrollmentStatusPotential); err != nil {
				s.Logger.Error("HandelOrderFlowEnrollmentStatus.HandleExistedLocations",
					zap.String("EnrollmentStatus", constant.StudentEnrollmentStatusEnrolled),
					zap.String("OrderStatus", req.OrderStatus),
					zap.String("OrderType", req.OrderType),
					zap.Error(err),
				)
				// return true to retry
				return true, err
			}
		}
	}
	return false, nil
}

func (s *HandelOrderFlowEnrollmentStatus) HandleEnrollmentStatusUpdate(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
	currentEnrollmentStatus, err := s.DomainEnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID(ctx, db, req.StudentID, req.LocationID, true)
	if err != nil {
		s.Logger.Error("HandelOrderFlowEnrollmentStatus.HandleEnrollmentStatusUpdate",
			zap.String("StudentID", req.StudentID),
			zap.String("LocationID", req.LocationID),
			zap.String("OrderStatus", req.OrderStatus),
			zap.String("OrderType", req.OrderType),
			zap.Error(err),
		)
		return true, err
	}

	if len(currentEnrollmentStatus) != 0 {
		// If location exists -> deactivate current enrollment and
		// set end date = start date - 1 second
		if currentEnrollmentStatus[0].EnrollmentStatus().String() == MapOrderTypeAndEnrollmentStatus[req.OrderType] {
			return false, nil
		}
		endDate := req.StartDate.Add(-1 * time.Second)
		err := s.DomainEnrollmentStatusHistoryRepo.
			DeactivateEnrollmentStatus(
				ctx, db,
				currentEnrollmentStatus[0],
				endDate,
			)
		if err != nil {
			s.Logger.Error("HandelOrderFlowEnrollmentStatus.HandleEnrollmentStatusUpdate",
				zap.String("StartDate", req.StartDate.String()),
				zap.String("CurrentEnrollmentStatus", currentEnrollmentStatus[0].EnrollmentStatus().String()),
				zap.String("OrderStatus", req.OrderStatus),
				zap.String("OrderType", req.OrderType),
				zap.Error(err),
			)
			return true, err
		}
	}

	newEnrollmentStatus := MapOrderTypeAndEnrollmentStatus[req.OrderType]
	if err := s.SyncEnrollmentStatusHistory(ctx, db, req, newEnrollmentStatus); err != nil {
		s.Logger.Error("HandelOrderFlowEnrollmentStatus.HandleEnrollmentStatusUpdate",
			zap.String("EnrollmentStatus", newEnrollmentStatus),
			zap.String("OrderStatus", req.OrderStatus),
			zap.String("OrderType", req.OrderType),
			zap.Error(err),
		)
		return true, err
	}

	return false, nil
}

func (s *HandelOrderFlowEnrollmentStatus) HandleVoidEnrollmentStatus(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
	latestEnrollmentStatus, err := s.DomainEnrollmentStatusHistoryRepo.GetLatestEnrollmentStudentOfLocation(ctx, db, req.StudentID, req.LocationID)
	if err != nil {
		s.Logger.Error("HandelOrderFlowEnrollmentStatus.HandleVoidEnrollmentStatus",
			zap.String("StudentID", req.StudentID),
			zap.String("LocationID", req.LocationID),
			zap.String("OrderStatus", req.OrderStatus),
			zap.String("OrderType", req.OrderType),
			zap.Error(err),
		)
		return true, err
	}
	if len(latestEnrollmentStatus) == 0 {
		return false, err
	}

	// Delete if enrollment status corresponds (don't have any status between submitted and voided)
	// Select 2 first record sort DESC -> [0] is latest record , [1] is previous record
	if req.OrderID == latestEnrollmentStatus[0].OrderID().String() {
		err := s.DomainEnrollmentStatusHistoryRepo.SoftDeleteEnrollments(ctx, db, latestEnrollmentStatus[0])
		if err != nil {
			s.Logger.Error("HandelOrderFlowEnrollmentStatus.HandleVoidEnrollmentStatus",
				zap.String("OrderID", latestEnrollmentStatus[0].OrderID().String()),
				zap.String("currentEnrollmentStatus", latestEnrollmentStatus[0].EnrollmentStatus().String()),
				zap.String("OrderStatus", req.OrderStatus),
				zap.String("OrderType", req.OrderType),
				zap.Error(err),
			)
			return true, err
		}
		if len(latestEnrollmentStatus) == 1 {
			err := s.DomainUserAccessPathRepo.SoftDeleteByUserIDAndLocationIDs(
				ctx, db,
				latestEnrollmentStatus[0].UserID().String(),
				latestEnrollmentStatus[0].OrganizationID().String(),
				[]string{latestEnrollmentStatus[0].LocationID().String()})
			if err != nil {
				s.Logger.Error("HandelOrderFlowEnrollmentStatus.HandleVoidEnrollmentStatus",
					zap.String("OrderID", latestEnrollmentStatus[0].OrderID().String()),
					zap.String("currentEnrollmentStatus", latestEnrollmentStatus[0].EnrollmentStatus().String()),
					zap.String("OrderStatus", req.OrderStatus),
					zap.String("OrderType", req.OrderType),
					zap.Error(err),
				)
				return true, err
			}
		}

		if len(latestEnrollmentStatus) == 2 {
			// Set previous record value end date to NULL
			err := s.DomainEnrollmentStatusHistoryRepo.DeactivateEnrollmentStatus(
				ctx, db,
				latestEnrollmentStatus[1],
				time.Time{},
			)
			if err != nil {
				s.Logger.Error("HandelOrderFlowEnrollmentStatus.HandleVoidEnrollmentStatus",
					zap.String("StartDate", req.StartDate.String()),
					zap.String("CurrentEnrollmentStatus", latestEnrollmentStatus[1].EnrollmentStatus().String()),
					zap.String("OrderStatus", req.OrderStatus),
					zap.String("OrderType", req.OrderType),
					zap.Error(err),
				)
				return true, err
			}
		}
		if err := s.DeactivateAndReactivateStudents(ctx, db, []string{req.StudentID}); err != nil {
			return true, err
		}
	}

	return false, nil
}

type EnrollmentStatusHistoryStartDateModifierFn func(context.Context, libdatabase.Ext, DomainEnrollmentStatusHistoryRepo, entity.DomainEnrollmentStatusHistory) (entity.DomainEnrollmentStatusHistory, error)

// EnrollmentStatusHistoryStartDateModifier will modify start date of enrollment status history if it is future date
// if there are duplicate enrollment status history with same start date but different order id, it will modify start date to be satisfied primary key constraint
// "pk__student_enrollment_status_history" PRIMARY KEY, btree (student_id, location_id, enrollment_status, start_date)
//
// if there are duplicate enrollment status history with same start date, order id and enrollment status, it should be throw error
func EnrollmentStatusHistoryStartDateModifier(ctx context.Context, db libdatabase.Ext, enrollmentStatusHistoryRepo DomainEnrollmentStatusHistoryRepo, enrollmentStatusHistory entity.DomainEnrollmentStatusHistory) (entity.DomainEnrollmentStatusHistory, error) {
	// Check if the start date of the enrollment status history is in the future
	if !utils.IsFutureDate(enrollmentStatusHistory.StartDate()) {
		return enrollmentStatusHistory, nil
	}

	// Get the enrollment status histories that have the same start date as the given one
	sameStartDateEnrollmentStatusHistories, err := enrollmentStatusHistoryRepo.GetSameStartDateEnrollmentStatusHistory(ctx, db, enrollmentStatusHistory)
	if err != nil {
		return nil, entity.InternalError{RawErr: err}
	}

	// If there are no enrollment status histories with the same start date, return the original one and nil error
	if len(sameStartDateEnrollmentStatusHistories) == 0 {
		return enrollmentStatusHistory, nil
	}

	// Create a slice of start dates from the same start date enrollment status histories
	startDates := make([]time.Time, 0, len(sameStartDateEnrollmentStatusHistories))
	for _, sameStartDateEnrollmentStatusHistory := range sameStartDateEnrollmentStatusHistories {
		// If the order ID of the same start date enrollment status history is equal to the given one, return nil and an existing data error
		if sameStartDateEnrollmentStatusHistory.OrderID().Equal(enrollmentStatusHistory.OrderID()) {
			return nil, entity.ExistingDataError{
				FieldName:  string(entity.FieldEnrollmentStatusHistoryOrderID),
				EntityName: entity.Entity(entity.EnrollmentStatusHistories),
			}
		}

		// Append the start date of the same start date enrollment status history to the slice
		startDates = append(startDates, sameStartDateEnrollmentStatusHistory.StartDate().Time())
	}

	// If there are some start dates in the slice, find the maximum one
	if len(startDates) > 0 {
		maxTime := utils.MaxTime(startDates)
		// The minimum unit in postgresql is Microsecond
		deltaTime := maxTime.Sub(enrollmentStatusHistory.StartDate().Time()) + time.Microsecond

		// Create a new enrollment status history with the modified start date by adding the delta time to the original one
		newTimeStartTime := enrollmentStatusHistory.StartDate().Time().Add(deltaTime)
		enrollmentStatusHistory = entity.NewEnrollmentStatusHistoryWithStartDate(enrollmentStatusHistory, field.NewTime(newTimeStartTime))
	}

	return enrollmentStatusHistory, nil
}

var _ EnrollmentStatusHistoryStartDateModifierFn = EnrollmentStatusHistoryStartDateModifier
