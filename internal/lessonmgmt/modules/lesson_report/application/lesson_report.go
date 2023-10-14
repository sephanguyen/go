package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lesson_infrastructure "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	lesson_report_consts "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/constant"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/infrastructure"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type LessonReportCommand struct {
	LessonReportRepo       infrastructure.LessonReportRepo
	LessonReportDetailRepo infrastructure.LessonReportDetailRepo
	LessonRepo             lesson_infrastructure.LessonRepo
	LessonMemberRepo       lesson_infrastructure.LessonMemberRepo
	PartnerFormConfigRepo  infrastructure.PartnerFormConfigRepo
	ReallocationRepo       lesson_infrastructure.ReallocationRepo
	MasterDataPort         lesson_infrastructure.MasterDataPort
	Logger                 *zap.Logger
}

func (l *LessonReportCommand) SaveDraft(ctx context.Context, db database.Ext, lessonReport *domain.LessonReport) error {
	lessonReport.ReportSubmittingStatus = lesson_report_consts.ReportSubmittingStatusSaved
	err := l.preStoreToDB(ctx, db, lessonReport)
	if err != nil {
		return fmt.Errorf("preStore to DB: %s", err)
	}
	return database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) (err error) {
		if err := l.storeToDB(ctx, tx, lessonReport); err != nil {
			return fmt.Errorf("store to DB: %s", err)
		}
		return nil
	})
}

func (l *LessonReportCommand) Submit(ctx context.Context, db database.Ext, lessonReport *domain.LessonReport) error {
	lessonReport.ReportSubmittingStatus = lesson_report_consts.ReportSubmittingStatusSubmitted
	err := l.preStoreToDB(ctx, db, lessonReport)
	if err != nil {
		return fmt.Errorf("preStore to DB: %s", err)
	}
	return database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) (err error) {
		if err := l.storeToDB(ctx, tx, lessonReport); err != nil {
			return fmt.Errorf("store to DB: %s", err)
		}
		return nil
	})
}

func (l *LessonReportCommand) preStoreToDB(ctx context.Context, db database.Ext, lessonReport *domain.LessonReport) error {
	isDraft := lessonReport.ReportSubmittingStatus == lesson_report_consts.ReportSubmittingStatusSaved
	if lessonReport.UnleashToggles["Lesson_LessonManagement_BackOffice_OptimisticLockingLessonReport"] {
		if err := l.validateOptimisticLockingLessonReport(ctx, db, lessonReport); err != nil {
			return fmt.Errorf("l.validateOptimisticLockingLessonReport: %v", err)
		}
	}

	// when update current lesson report, we need
	// check existence of id and didn't be changed lesson id
	if len(lessonReport.LessonReportID) != 0 {
		currentLessonReport, err := l.LessonReportRepo.FindByID(ctx, db, lessonReport.LessonReportID)
		if err != nil {
			return fmt.Errorf("LessonReportRepo.FindByID: %v", err)
		}
		lessonReport.LessonID = currentLessonReport.LessonID
	} else {
		// if there is no lesson report id (creating), check there is any current lesson report
		currentLessonReport, err := l.LessonReportRepo.FindByLessonID(ctx, db, lessonReport.LessonID)
		if err == nil {
			if isDraft {
				// if ReportSubmittingStatus == save draft, assign report_id
				lessonReport.LessonReportID = currentLessonReport.LessonReportID
				lessonReport.LessonID = currentLessonReport.LessonID
			} else {
				// if ReportSubmittingStatus == submit, don't assign report_id
				return fmt.Errorf("each lesson must only have a lesson report")
			}
		} else if err != nil && err.Error() != domain.NotFoundDBErr {
			return fmt.Errorf("LessonReportRepo.FindByLessonID: %v", err)
		}
	}

	lessonReport, err := l.Normalize(ctx, db, lessonReport)
	if err != nil {
		return err
	}

	// check permission when submit report
	if !isDraft && lessonReport.UnleashToggles[lesson_report_consts.PermissionToSubmitReport] {
		err = l.CheckSubmitReportPermission(ctx, db, lessonReport.Lesson.LocationID)
		if err != nil {
			return fmt.Errorf("unable submit this report: %w", err)
		}
	}
	_, isExists := lessonReport.FormConfig.FormConfigData.GetFieldByID(string(lesson_report_consts.SystemDefinedFieldAttendanceStatus))
	if !isExists {
		if lessonReport.UnleashToggles["Lesson_LessonManagement_BackOffice_ValidationLessonBeforeCompleted"] {
			lessonMembers, err := l.LessonMemberRepo.GetLessonMembersInLessons(ctx, db, []string{lessonReport.LessonID})
			if err != nil {
				return fmt.Errorf("LessonMemberRepo.GetLessonMembersInLessons: %v", err)
			}
			mapMemberInfo := make(map[string]*lesson_domain.LessonMember)
			for _, member := range lessonMembers {
				mapMemberInfo[member.StudentID] = member
			}

			for _, detail := range lessonReport.Details {
				if lessonReport.UnleashToggles["Lesson_LessonManagement_BackOffice_ValidationLessonBeforeCompleted"] {
					if detail.AttendanceStatus == "" || detail.AttendanceStatus == lesson_report_consts.StudentAttendStatusEmpty {
						if memberInfo, ok := mapMemberInfo[detail.StudentID]; ok {
							detail.AttendanceStatus = lesson_report_consts.StudentAttendStatus(memberInfo.AttendanceStatus)
							detail.AttendanceRemark = memberInfo.AttendanceRemark
							detail.AttendanceNote = memberInfo.AttendanceNote
							detail.AttendanceNotice = lesson_report_consts.StudentAttendanceNotice(memberInfo.AttendanceNotice)
							detail.AttendanceReason = lesson_report_consts.StudentAttendanceReason(memberInfo.AttendanceReason)
						}
					}
				}
			}
		} else {
			err = lessonReport.Details.RemoveAttendanceInfo()
			if err != nil {
				return fmt.Errorf("LessonReportDetail.RemoveAttendanceInfo.Err: %v", err)
			}
			lessonReport.IsUpdateMembersInfo = false
		}
	}

	if err := lessonReport.IsValid(ctx, db, isDraft); err != nil {
		return err
	}

	return nil
}

// Normalize will normalize fields in LessonReport struct and also fill missing fields
//   - Set default value for SubmittingStatus if it empties
//   - Fill FormConfig's data if it empties
//   - Remove lesson details which has StudentID not belong to lesson
//   - Normalize Details
func (l *LessonReportCommand) Normalize(ctx context.Context, db database.Ext, lessonReport *domain.LessonReport) (*domain.LessonReport, error) {
	if len(lessonReport.ReportSubmittingStatus) == 0 {
		lessonReport.ReportSubmittingStatus = lesson_report_consts.ReportSubmittingStatusSaved
	}

	// check lesson id
	lesson, err := l.LessonRepo.GetLessonByID(ctx, db, lessonReport.LessonID)
	if err != nil {
		return lessonReport, fmt.Errorf("LessonRepo.FindByID: %v", err)
	}
	lessonReport.Lesson = lesson
	// fill form config
	if lessonReport.FormConfig == nil {
		formCfg, err := l.getFormConfig(ctx, db, lesson, lessonReport.FeatureName)
		if err != nil || formCfg == nil {
			return lessonReport, fmt.Errorf("could not get form config: %v", err)
		}
		lessonReport.FormConfig = formCfg
	}

	// normalize details
	lessonReport.Details.Normalize()

	return lessonReport, nil
}

func (l *LessonReportCommand) storeToDB(ctx context.Context, db database.Ext, lessonReport *domain.LessonReport) error {
	var err error
	isCreated := len(lessonReport.LessonReportID) == 0

	if isCreated {
		// insert lesson report record
		lessonReport.CreatedAt, lessonReport.UpdatedAt = time.Now(), time.Now()
		e, err := l.LessonReportRepo.Create(ctx, db, lessonReport)
		if err != nil {
			return fmt.Errorf("LessonReportRepo.Create: %v", err)
		}
		lessonReport.LessonReportID = e.LessonReportID
	} else {
		lessonReport.UpdatedAt = time.Now()
		_, err := l.LessonReportRepo.Update(ctx, db, lessonReport)
		if err != nil {
			return fmt.Errorf("LessonReportRepo.Update: %v", err)
		}
	}

	// upsert lesson report details record
	details, err := lessonReport.Details.ToLessonReportDetailsDomain(lessonReport.LessonReportID)
	if err != nil {
		return fmt.Errorf("could not convert to lesson report details entity: %v", err)
	}

	if lessonReport.UnleashToggles["Lesson_LessonManagement_BackOffice_OptimisticLockingLessonReport"] {
		if lessonReport.IsSavePerStudent {
			if err = l.LessonReportDetailRepo.UpsertOne(ctx, db, lessonReport.LessonReportID, *details[0]); err != nil {
				return fmt.Errorf("LessonReportDetailRepo.UpsertOne: %v", err)
			}
		} else {
			if err = l.LessonReportDetailRepo.UpsertWithVersion(ctx, db, lessonReport.LessonReportID, details); err != nil {
				return fmt.Errorf("LessonReportDetailRepo.UpsertWithVersion: %v", err)
			}
		}
	} else {
		if err = l.LessonReportDetailRepo.Upsert(ctx, db, lessonReport.LessonReportID, details); err != nil {
			return fmt.Errorf("LessonReportDetailRepo.Upsert: %v", err)
		}
	}

	// get lesson report detail ids
	details, err = l.LessonReportDetailRepo.GetByLessonReportID(ctx, db, lessonReport.LessonReportID)
	if err != nil {
		return fmt.Errorf("LessonReportDetailRepo.GetByLessonReportID: %v", err)
	}

	detailByStudentID := make(map[string]*domain.LessonReportDetail)
	for i := range details {
		detailByStudentID[details[i].StudentID] = details[i]
	}

	// upsert field values for each lesson report detail
	lessonReportDetailIDs := []string{}
	var fieldValues []*domain.PartnerDynamicFormFieldValue
	studentAttendanceStatus := make(map[string]lesson_domain.StudentAttendStatus)
	for _, detail := range lessonReport.Details {
		studentAttendanceStatus[detail.StudentID] = lesson_domain.StudentAttendStatus(detail.AttendanceStatus)
		id := detailByStudentID[detail.StudentID].LessonReportDetailID
		e, err := detail.Fields.ToPartnerDynamicFormFieldValueEntities(id)
		if err != nil {
			return fmt.Errorf("could not convert to partner dynamic form field value: %v", err)
		}
		fieldValues = append(fieldValues, e...)
		lessonReportDetailIDs = append(lessonReportDetailIDs, id)
	}

	// when lesson is locked, the report can not change the attendance status of lesson member
	if !lessonReport.Lesson.IsLocked && lessonReport.IsUpdateMembersInfo {
		// upsert lesson member's fields
		members := lessonReport.Details.ToLessonMembersEntity(lessonReport.LessonID)
		if err != nil {
			return fmt.Errorf("could not convert to lesson member: %v", err)
		}
		if err = l.LessonMemberRepo.UpdateLessonMembersFields(
			ctx,
			db,
			members,
			repo.UpdateLessonMemberFields{
				repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceRemark),
				repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceStatus),
				repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNotice),
				repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceReason),
				repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNote),
			},
		); err != nil {
			return fmt.Errorf("LessonMemberRepo.UpdateLessonMembersFields: %v", err)
		}
	}
	// for update: delete field values before upserting
	if len(lessonReportDetailIDs) > 0 {
		lessonReportDetailIDs = golibs.Uniq(lessonReportDetailIDs)
		if err = l.PartnerFormConfigRepo.DeleteByLessonReportDetailIDs(ctx, db, lessonReportDetailIDs); err != nil {
			return fmt.Errorf("PartnerFormConfigRepo.DeleteByLessonReportDetailIDs: %v", err)
		}
	}

	if err = l.LessonReportDetailRepo.UpsertFieldValues(ctx, db, fieldValues); err != nil {
		return fmt.Errorf("LessonReportDetailRepo.UpsertFieldValues: %v", err)
	}

	lesson := lessonReport.Lesson
	learners := lesson.Learners
	studentReallocate := learners.GetStudentReallocate(studentAttendanceStatus)
	if len(studentReallocate) > 0 {
		reallocation := make([]*lesson_domain.Reallocation, 0, len(studentReallocate))
		learnerMap := learners.GroupByLearnerID()
		for _, studentID := range studentReallocate {
			reallocation = append(reallocation, &lesson_domain.Reallocation{
				OriginalLessonID: lesson.LessonID,
				StudentID:        studentID,
				CourseID:         learnerMap[studentID].CourseID,
			})
		}
		if err = l.ReallocationRepo.UpsertReallocation(ctx, db, lesson.LessonID, reallocation); err != nil {
			return fmt.Errorf("l.ReallocationRepo.UpsertReallocation: %w", err)
		}
	}
	studentUnReallocate := learners.GetStudentUnReallocate(studentAttendanceStatus)
	if len(studentUnReallocate) > 0 {
		reallocations, err := l.ReallocationRepo.GetFollowingReallocation(ctx, db, lesson.LessonID, studentUnReallocate)
		if err != nil {
			return fmt.Errorf("l.ReallocationRepo.GetFollowingReallocation: %w", err)
		}
		studentReallocateRemoved := []string{}
		lessonMemberRemoved := make([]*lesson_domain.LessonMember, 0)
		for _, r := range reallocations {
			studentReallocateRemoved = append(studentReallocateRemoved, r.StudentID, r.OriginalLessonID)
			if slices.Contains(studentUnReallocate, r.StudentID) && r.OriginalLessonID == lesson.LessonID {
				continue
			}
			lessonMemberRemoved = append(lessonMemberRemoved, &lesson_domain.LessonMember{
				LessonID:  r.OriginalLessonID,
				StudentID: r.StudentID,
			})
		}
		if len(studentReallocateRemoved) > 0 {
			if err := l.ReallocationRepo.SoftDelete(ctx, db, studentReallocateRemoved, true); err != nil {
				return fmt.Errorf("l.ReallocationRepo.SoftDelete: %w", err)
			}
		}
		if len(lessonMemberRemoved) > 0 {
			if err = l.LessonMemberRepo.DeleteLessonMembers(ctx, db, lessonMemberRemoved); err != nil {
				return fmt.Errorf("l.LessonMemberRepo.DeleteLessonMembers: %w", err)
			}
		}
	}
	return nil
}

func (l *LessonReportCommand) getFormConfig(ctx context.Context, db database.Ext, lesson *lesson_domain.Lesson, featureName string) (*domain.FormConfig, error) {
	schoolID := golibs.ResourcePathFromCtx(ctx)
	schoolIDInt, err := strconv.Atoi(schoolID)
	if err != nil {
		return nil, fmt.Errorf("getFormConfig strconv.Atoi err: %v", err)
	}
	var feature string
	if len(featureName) > 0 {
		feature = featureName
	} else {
		//if FE doesn't input anything for feature name, we will based on lesson teaching method
		switch lesson.TeachingMethod {
		case lesson_domain.LessonTeachingMethodGroup:
			feature = string(lesson_report_consts.FeatureNameGroupLessonReport)
		case lesson_domain.LessonTeachingMethodIndividual:
			feature = string(lesson_report_consts.FeatureNameIndividualUpdateLessonReport)
		default:
			return nil, fmt.Errorf("LessonReportCommand.getFormConfig: lesson id doesn't have a teaching method %s and featureName cannot be emptied", lesson.LessonID)
		}
	}
	cf, err := l.PartnerFormConfigRepo.FindByPartnerAndFeatureName(ctx, db, schoolIDInt, feature)
	if err != nil {
		return nil, fmt.Errorf("PartnerFormConfigRepo.FindByPartnerAndFeatureName: %v", err)
	}

	formCfg, err := domain.NewFormConfigByPartnerFormConfig(cf)
	if err != nil {
		return nil, err
	}

	return formCfg, nil
}

func (l *LessonReportCommand) validateOptimisticLockingLessonReport(ctx context.Context, db database.Ext, lessonReport *domain.LessonReport) error {
	lessonReportDetailDB, err := l.LessonReportDetailRepo.GetReportVersionByLessonID(ctx, db, lessonReport.LessonID)

	if err != nil {
		return fmt.Errorf("LessonReportDetailRepo.GetReportVersionByLessonID: %v", err)
	}

	if len(lessonReportDetailDB) > 0 {
		mapStudentToLessonReportVersionDB := make(map[string]int, len(lessonReportDetailDB))

		for _, detail := range lessonReportDetailDB {
			mapStudentToLessonReportVersionDB[detail.StudentID] = detail.ReportVersion
		}
		// case per student check version by index == 0 in array detail
		if lessonReport.IsSavePerStudent && len(lessonReport.Details) == 1 {
			if reportVersion, oke := mapStudentToLessonReportVersionDB[lessonReport.Details[0].StudentID]; oke {
				if lessonReport.Details[0].ReportVersion != reportVersion {
					l.Logger.Info("[validateOptimisticLockingLessonReport][IsSavePerStudent]: the version lesson report is out of date",
						zap.String("studentID", lessonReport.Details[0].StudentID),
						zap.String("reportVersionDB", fmt.Sprintf(`%d`, reportVersion)),
						zap.String("reportVersionFE", fmt.Sprintf(`%d`, lessonReport.Details[0].ReportVersion)),
					)

					return domain.ErrReportVersionIsOutOfDate
				}
			}
		} else {
			// case check all version in array detail
			if len(lessonReport.Details) != 0 {
				for _, reportDetail := range lessonReport.Details {
					if reportVersion, oke := mapStudentToLessonReportVersionDB[reportDetail.StudentID]; oke {
						if reportDetail.ReportVersion != reportVersion {
							l.Logger.Info("[validateOptimisticLockingLessonReport][CheckAllVersion]: the version lesson report is out of date",
								zap.String("studentID", reportDetail.StudentID),
								zap.String("reportVersionDB", fmt.Sprintf(`%d`, reportVersion)),
								zap.String("reportVersionFE", fmt.Sprintf(`%d`, reportDetail.ReportVersion)),
							)
							return domain.ErrReportVersionIsOutOfDate
						}
					}
				}
			}
		}
	}

	return nil
}

func (l *LessonReportCommand) CheckSubmitReportPermission(ctx context.Context, db database.Ext, locationID string) error {
	userID := interceptors.UserIDFromContext(ctx)
	permissionNames := []string{lesson_report_consts.ReportReviewPermission, lesson_report_consts.LessonWritePermission}
	grantedPermissions, err := l.MasterDataPort.FindPermissionByNamesAndUserID(ctx, db, permissionNames, userID)
	if err != nil {
		return fmt.Errorf("cannot get permission of this user(%s): %s", userID, err)
	}

	if !slices.Contains(grantedPermissions.Permissions, lesson_report_consts.ReportReviewPermission) {
		return fmt.Errorf("this user don't have permission to submit the report")
	}

	// if !slices.Contains(grantedPermissions.GrantedLocations[lesson_report_consts.LessonWritePermission], locationID) {
	// 	return fmt.Errorf("this user don't have permission on this location")
	// }

	return nil
}
