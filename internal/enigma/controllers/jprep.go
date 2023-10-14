package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/manabie-com/backend/internal/enigma/dto"
	"github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/enigma/middlewares"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const layout string = "2006-01-02"
const maxRetryTimes int32 = 3
const limitDateRange int32 = 720

var (
	ErrTimestampIsRequired = fmt.Errorf("Timestamp is required")
)

type JPREPController struct {
	Logger                 *zap.Logger
	JSM                    nats.JetStreamManagement
	DB                     database.Ext
	PartnerSyncDataLogRepo interface {
		Create(ctx context.Context, db database.QueryExecer, a *entities.PartnerSyncDataLog) error
		UpdateTime(ctx context.Context, db database.QueryExecer, logID string) error
	}
	PartnerSyncDataLogSplitRepo interface {
		Create(ctx context.Context, db database.QueryExecer, a *entities.PartnerSyncDataLogSplit) error
		GetLogsBySignature(ctx context.Context, db database.QueryExecer, signature pgtype.Text) ([]*entities.PartnerSyncDataLogSplit, error)
		GetLogsReportByDate(ctx context.Context, db database.QueryExecer, fromDate, toDate pgtype.Date) ([]*entities.PartnerSyncDataLogReport, error)
		GetLogsByDateRange(ctx context.Context, db database.QueryExecer, fromDate, toDate pgtype.Date) ([]*entities.PartnerSyncDataLogSplit, error)
		UpdateLogsStatusAndRetryTime(ctx context.Context, db database.QueryExecer, logs []*entities.PartnerSyncDataLogSplit) error
	}
}

type LogStructure struct {
	PartnerSyncDataLogID string
	Payload              []byte
	Kind                 string
}

const MaxRecordProcessPertime = 500

func RegisterJPREPController(r *gin.RouterGroup, c *JPREPController) {
	r.PUT("/user-registration", c.UserRegistration)
	r.PUT("/user-course", c.SyncUserCourse)
	r.PUT("/master-registration", c.MasterRegistration)
	r.POST("/partner-log", c.PartnerLog)
	r.POST("/partner-log/report", c.PartnerLogReport)
	r.POST("/partner-log/recover", c.PartnerLogRecover)
}

func (j *JPREPController) SyncUserCourse(c *gin.Context) {
	event, err := j.toEventSyncUserCourse(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: fmt.Sprint(constants.JPREPSchool),
			UserID:       constants.SyncAccount,
		},
	}
	ctx := interceptors.ContextWithJWTClaims(c, claim)
	parentLog, err := j.logSyncData(ctx, event.RawPayload, event.Signature)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	// chunk the student lessons
	j.Logger.Info(fmt.Sprintf("Start sync %d lesson_members", len(event.StudentLessons)))
	err = nats.ChunkHandler(len(event.StudentLessons), MaxRecordProcessPertime, func(start, end int) error {
		payload, err := json.Marshal(event.StudentLessons[start:end])
		if err != nil {
			return fmt.Errorf("json.Marshal StudentLessons: %w", err)
		}
		log, err := j.logSyncDataSplit(ctx, &LogStructure{
			PartnerSyncDataLogID: parentLog.PartnerSyncDataLogID.String,
			Payload:              payload,
			Kind:                 string(entities.KindStudentLessons),
		})
		if err != nil {
			return err
		}
		msg, _ := proto.Marshal(&npb.EventSyncUserCourse{
			Signature:      event.Signature,
			Timestamp:      event.Timestamp,
			StudentLessons: event.StudentLessons[start:end],
			LogId:          log.PartnerSyncDataLogSplitID.String,
		})

		_, err = j.JSM.PublishContext(ctx, constants.SubjectJPREPSyncUserCourseNatsJS, msg)
		return err
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (j *JPREPController) UserRegistration(c *gin.Context) {
	event, err := j.toEventUserRegistration(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: fmt.Sprint(constants.JPREPSchool),
			UserID:       constants.SyncAccount,
		},
	}
	ctx := interceptors.ContextWithJWTClaims(c, claim)
	parentLog, err := j.logSyncData(ctx, event.RawPayload, event.Signature)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	// chunk the students
	j.Logger.Info(fmt.Sprintf("Start sync %d students", len(event.Students)))
	err = nats.ChunkHandler(len(event.Students), MaxRecordProcessPertime, func(start, end int) error {
		payload, err := json.Marshal(event.Students[start:end])
		if err != nil {
			return fmt.Errorf("json.Marshal Students: %w", err)
		}
		log, err := j.logSyncDataSplit(ctx, &LogStructure{
			PartnerSyncDataLogID: parentLog.PartnerSyncDataLogID.String,
			Payload:              payload,
			Kind:                 string(entities.KindStudent),
		})
		if err != nil {
			return err
		}
		msg, _ := proto.Marshal(&npb.EventUserRegistration{
			Signature: event.Signature,
			Timestamp: event.Timestamp,
			Students:  event.Students[start:end],
			LogId:     log.PartnerSyncDataLogSplitID.String,
		})

		_, err = j.JSM.PublishContext(ctx, constants.SubjectUserRegistrationNatsJS, msg)

		return err
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// chunk the staffs
	j.Logger.Info(fmt.Sprintf("Start sync %d Staffs", len(event.Staffs)))
	err = nats.ChunkHandler(len(event.Staffs), MaxRecordProcessPertime, func(start, end int) error {
		payload, err := json.Marshal(event.Staffs[start:end])
		if err != nil {
			return fmt.Errorf("json.Marshal Staffs: %w", err)
		}
		log, err := j.logSyncDataSplit(ctx, &LogStructure{
			PartnerSyncDataLogID: parentLog.PartnerSyncDataLogID.String,
			Payload:              payload,
			Kind:                 string(entities.KindStaff),
		})
		if err != nil {
			return err
		}
		msg, _ := proto.Marshal(&npb.EventUserRegistration{
			Signature: event.Signature,
			Timestamp: event.Timestamp,
			Staffs:    event.Staffs[start:end],
			LogId:     log.PartnerSyncDataLogSplitID.String,
		})

		_, err = j.JSM.PublishContext(ctx, constants.SubjectUserRegistrationNatsJS, msg)
		return err
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (j *JPREPController) MasterRegistration(c *gin.Context) {
	event, err := j.toEventMasterRegistration(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: fmt.Sprint(constants.JPREPSchool),
			UserID:       constants.SyncAccount,
		},
	}
	ctx := interceptors.ContextWithJWTClaims(c, claim)
	parentLog, err := j.logSyncData(ctx, event.RawPayload, event.Signature)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	// chunk the courses
	j.Logger.Info(fmt.Sprintf("Start sync %d courses", len(event.Courses)))
	err = nats.ChunkHandler(len(event.Courses), MaxRecordProcessPertime, func(start, end int) error {
		payload, err := json.Marshal(event.Courses[start:end])
		if err != nil {
			return fmt.Errorf("json.Marshal Courses: %w", err)
		}
		log, err := j.logSyncDataSplit(ctx, &LogStructure{
			PartnerSyncDataLogID: parentLog.PartnerSyncDataLogID.String,
			Payload:              payload,
			Kind:                 string(entities.KindCourse),
		})
		if err != nil {
			return err
		}
		msg, _ := proto.Marshal(&npb.EventMasterRegistration{
			Signature: event.Signature,
			Timestamp: event.Timestamp,
			Courses:   event.Courses[start:end],
			LogId:     log.PartnerSyncDataLogSplitID.String,
		})

		_, err = j.JSM.PublishAsyncContext(ctx, constants.SubjectSyncMasterRegistration, msg)
		return err
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// chunk the classes
	j.Logger.Info(fmt.Sprintf("Start sync %d class", len(event.Classes)))
	err = nats.ChunkHandler(len(event.Classes), MaxRecordProcessPertime, func(start, end int) error {
		payload, err := json.Marshal(event.Classes[start:end])
		if err != nil {
			return fmt.Errorf("json.Marshal Classes: %w", err)
		}
		log, err := j.logSyncDataSplit(ctx, &LogStructure{
			PartnerSyncDataLogID: parentLog.PartnerSyncDataLogID.String,
			Payload:              payload,
			Kind:                 string(entities.KindClass),
		})
		if err != nil {
			return err
		}
		msg, _ := proto.Marshal(&npb.EventMasterRegistration{
			Signature: event.Signature,
			Timestamp: event.Timestamp,
			Classes:   event.Classes[start:end],
			LogId:     log.PartnerSyncDataLogSplitID.String,
		})

		_, err = j.JSM.PublishAsyncContext(ctx, constants.SubjectSyncMasterRegistration, msg)
		return err
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// chunk the lessons
	j.Logger.Info(fmt.Sprintf("Start sync %d lessons", len(event.Lessons)))
	err = nats.ChunkHandler(len(event.Lessons), MaxRecordProcessPertime, func(start, end int) error {
		payload, err := json.Marshal(event.Lessons[start:end])
		if err != nil {
			return fmt.Errorf("json.Marshal Lessons: %w", err)
		}
		log, err := j.logSyncDataSplit(ctx, &LogStructure{
			PartnerSyncDataLogID: parentLog.PartnerSyncDataLogID.String,
			Payload:              payload,
			Kind:                 string(entities.KindLesson),
		})
		if err != nil {
			return err
		}
		msg, _ := proto.Marshal(&npb.EventMasterRegistration{
			Signature: event.Signature,
			Timestamp: event.Timestamp,
			Lessons:   event.Lessons[start:end],
			LogId:     log.PartnerSyncDataLogSplitID.String,
		})

		_, err = j.JSM.PublishAsyncContext(ctx, constants.SubjectSyncMasterRegistration, msg)
		return err
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// chunk the academicYears
	j.Logger.Info(fmt.Sprintf("Start sync %d academicYears", len(event.AcademicYears)))
	err = nats.ChunkHandler(len(event.AcademicYears), MaxRecordProcessPertime, func(start, end int) error {
		payload, err := json.Marshal(event.AcademicYears[start:end])
		if err != nil {
			return fmt.Errorf("json.Marshal AcademicYears: %w", err)
		}
		log, err := j.logSyncDataSplit(ctx, &LogStructure{
			PartnerSyncDataLogID: parentLog.PartnerSyncDataLogID.String,
			Payload:              payload,
			Kind:                 string(entities.KindAcademicYear),
		})
		if err != nil {
			return err
		}
		msg, _ := proto.Marshal(&npb.EventMasterRegistration{
			Signature:     event.Signature,
			Timestamp:     event.Timestamp,
			AcademicYears: event.AcademicYears[start:end],
			LogId:         log.PartnerSyncDataLogSplitID.String,
		})

		_, err = j.JSM.PublishAsyncContext(ctx, constants.SubjectSyncMasterRegistration, msg)
		return err
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (j JPREPController) PartnerLog(c *gin.Context) {
	payload := middlewares.PayloadFromContext(c)
	if len(payload) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "payload body is empty",
		})
		return
	}

	req := &dto.PartnerLogRequest{}
	if err := json.Unmarshal(payload, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if req.Timestamp == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": ErrTimestampIsRequired,
		})
		return
	}
	signature := req.Payload.Signature
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Errorf("signature empty"),
		})
		return
	}

	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: fmt.Sprint(constants.JPREPSchool),
			UserID:       constants.SyncAccount,
		},
	}

	ctx := interceptors.ContextWithJWTClaims(c, claim)
	partnerLogSplits, err := j.PartnerSyncDataLogSplitRepo.GetLogsBySignature(ctx, j.DB, database.Text(signature))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp := []*dto.PartnerLogResponse{}
	for _, log := range partnerLogSplits {
		logResp := &dto.PartnerLogResponse{
			PartnerSyncDataLogSplitID: log.PartnerSyncDataLogSplitID.String,
			Status:                    log.Status.String,
			UpdatedAt:                 log.UpdatedAt.Time.Unix(),
		}

		resp = append(resp, logResp)
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (j JPREPController) PartnerLogReport(c *gin.Context) {
	payload := middlewares.PayloadFromContext(c)
	if len(payload) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("payload body is empty"),
		})
		return
	}

	req := &dto.PartnerLogRequestByDate{}
	if err := json.Unmarshal(payload, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: fmt.Sprint(constants.JPREPSchool),
			UserID:       constants.SyncAccount,
		},
	}
	ctx := interceptors.ContextWithJWTClaims(c, claim)
	res, err := j.getPartnerLogReport(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": res})
}

func (j JPREPController) PartnerLogRecover(c *gin.Context) {
	payload := middlewares.PayloadFromContext(c)
	if len(payload) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("payload body is empty"),
		})
		return
	}

	req := &dto.PartnerLogRequestByDate{}
	if err := json.Unmarshal(payload, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("json.Unmarshal err"),
		})
		return
	}
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: fmt.Sprint(constants.JPREPSchool),
			UserID:       constants.SyncAccount,
		},
	}
	ctx := interceptors.ContextWithJWTClaims(c, claim)
	err := j.recoverData(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (j *JPREPController) recoverData(ctx context.Context, req *dto.PartnerLogRequestByDate) error {
	if req.Timestamp == 0 {
		return ErrTimestampIsRequired
	}

	fromDate := req.Payload.FromDate
	toDate := req.Payload.ToDate
	startDate, endDate, err := validateAndParseDate(fromDate, toDate)
	if err != nil {
		return err
	}
	logs, err := j.PartnerSyncDataLogSplitRepo.GetLogsByDateRange(ctx, j.DB, startDate, endDate)
	if err != nil {
		return err
	}
	for _, log := range logs {
		if log.RetryTimes.Int >= maxRetryTimes {
			err = log.Status.Set(string(entities.StatusFailed))
			if err != nil {
				return err
			}
		} else {
			err = log.RetryTimes.Set(log.RetryTimes.Int + 1)
			if err != nil {
				return err
			}
		}
	}
	err = j.PartnerSyncDataLogSplitRepo.UpdateLogsStatusAndRetryTime(ctx, j.DB, logs)
	if err != nil {
		return err
	}
	if len(logs) > 0 {
		err = j.PartnerSyncDataLogRepo.UpdateTime(ctx, j.DB, logs[0].PartnerSyncDataLogID.String)
		if err != nil {
			return err
		}
	}
	return j.publishMessageToRecover(ctx, logs)
}

func (j *JPREPController) publishMessageToRecover(ctx context.Context, logs []*entities.PartnerSyncDataLogSplit) error {
	for _, log := range logs {
		if log.RetryTimes.Int <= maxRetryTimes {
			switch log.Kind.String {
			case string(entities.KindStudent):
				var students []*npb.EventUserRegistration_Student
				err := json.Unmarshal(log.Payload.Bytes, &students)
				if err == nil {
					msg, err := proto.Marshal(&npb.EventUserRegistration{
						Students: students,
						LogId:    log.PartnerSyncDataLogSplitID.String,
					})
					if err == nil {
						if _, err = j.JSM.PublishContext(ctx, constants.SubjectUserRegistrationNatsJS, msg); err != nil {
							j.Logger.Error(fmt.Sprintf("err recover student %v", err))
						}
					}
				}
			case string(entities.KindStaff):
				var staffs []*npb.EventUserRegistration_Staff
				err := json.Unmarshal(log.Payload.Bytes, &staffs)
				if err == nil {
					msg, err := proto.Marshal(&npb.EventUserRegistration{
						Staffs: staffs,
						LogId:  log.PartnerSyncDataLogSplitID.String,
					})
					if err == nil {
						if _, err = j.JSM.PublishContext(ctx, constants.SubjectUserRegistrationNatsJS, msg); err != nil {
							j.Logger.Error(fmt.Sprintf("err recover staff %v", err))
						}
					}
				}
			case string(entities.KindLesson):
				var lessons []*npb.EventMasterRegistration_Lesson
				err := json.Unmarshal(log.Payload.Bytes, &lessons)
				if err == nil {
					msg, err := proto.Marshal(&npb.EventMasterRegistration{
						Lessons: lessons,
						LogId:   log.PartnerSyncDataLogSplitID.String,
					})
					if err == nil {
						if _, err = j.JSM.PublishContext(ctx, constants.SubjectSyncMasterRegistration, msg); err != nil {
							j.Logger.Error(fmt.Sprintf("err recover lesson %v", err))
						}
					}
				}
			case string(entities.KindCourse):
				var courses []*npb.EventMasterRegistration_Course
				err := json.Unmarshal(log.Payload.Bytes, &courses)
				if err == nil {
					msg, err := proto.Marshal(&npb.EventMasterRegistration{
						Courses: courses,
						LogId:   log.PartnerSyncDataLogSplitID.String,
					})
					if err == nil {
						if _, err = j.JSM.PublishContext(ctx, constants.SubjectSyncMasterRegistration, msg); err != nil {
							j.Logger.Error(fmt.Sprintf("err recover course %v", err))
						}
					}
				}
			case string(entities.KindClass):
				var classes []*npb.EventMasterRegistration_Class
				err := json.Unmarshal(log.Payload.Bytes, &classes)
				if err == nil {
					msg, err := proto.Marshal(&npb.EventMasterRegistration{
						Classes: classes,
						LogId:   log.PartnerSyncDataLogSplitID.String,
					})
					if err == nil {
						if _, err = j.JSM.PublishContext(ctx, constants.SubjectSyncMasterRegistration, msg); err != nil {
							j.Logger.Error(fmt.Sprintf("err recover class %v", err))
						}
					}
				}
			case string(entities.KindAcademicYear):
				var academicYears []*npb.EventMasterRegistration_AcademicYear
				err := json.Unmarshal(log.Payload.Bytes, &academicYears)
				if err == nil {
					msg, err := proto.Marshal(&npb.EventMasterRegistration{
						AcademicYears: academicYears,
						LogId:         log.PartnerSyncDataLogSplitID.String,
					})
					if err == nil {
						if _, err = j.JSM.PublishContext(ctx, constants.SubjectSyncMasterRegistration, msg); err != nil {
							j.Logger.Error(fmt.Sprintf("err recover academic year %v", err))
						}
					}
				}
			case string(entities.KindStudentLessons):
				var studentLessons []*npb.EventSyncUserCourse_StudentLesson
				err := json.Unmarshal(log.Payload.Bytes, &studentLessons)
				if err == nil {
					msg, err := proto.Marshal(&npb.EventSyncUserCourse{
						StudentLessons: studentLessons,
						LogId:          log.PartnerSyncDataLogSplitID.String,
					})
					if err == nil {
						if _, err = j.JSM.PublishContext(ctx, constants.SubjectJPREPSyncUserCourseNatsJS, msg); err != nil {
							j.Logger.Error(fmt.Sprintf("err recover student lesson %v", err))
						}
					}
				}
			}
		}
	}
	return nil
}

func (j *JPREPController) getPartnerLogReport(ctx context.Context, req *dto.PartnerLogRequestByDate) (map[string]map[string]int, error) {
	if req.Timestamp == 0 {
		return nil, ErrTimestampIsRequired
	}

	fromDate := req.Payload.FromDate
	toDate := req.Payload.ToDate
	startDate, endDate, err := validateAndParseDate(fromDate, toDate)
	if err != nil {
		return nil, err
	}
	reports, err := j.PartnerSyncDataLogSplitRepo.GetLogsReportByDate(ctx, j.DB, startDate, endDate)
	if err != nil {
		return nil, err
	}
	res := make(map[string]map[string]int)
	for _, report := range reports {
		if val, ok := res[report.CreatedAt.Time.Format(layout)]; ok {
			val[report.Status.String] = int(report.Total.Int)
		} else {
			res[report.CreatedAt.Time.Format(layout)] = map[string]int{
				string(entities.StatusPending):    0,
				string(entities.StatusProcessing): 0,
				string(entities.StatusSuccess):    0,
				string(entities.StatusFailed):     0,
			}
			res[report.CreatedAt.Time.Format(layout)][report.Status.String] = int(report.Total.Int)
		}
	}
	return res, nil
}

func (j *JPREPController) toEventUserRegistration(c *gin.Context) (*npb.EventUserRegistration, error) {
	payload := middlewares.PayloadFromContext(c)
	if len(payload) == 0 {
		return nil, fmt.Errorf("payload body is empty")
	}

	req := &dto.UserRegistrationRequest{}
	if err := json.Unmarshal(payload, req); err != nil {
		return nil, err
	}

	if req.Timestamp == 0 {
		return nil, ErrTimestampIsRequired
	}

	signature := c.GetHeader(middlewares.JPREPHeaderKey)
	logger := j.Logger.With(zap.String("signature", signature))

	j.Logger.Info(fmt.Sprintf("req.Payload.Students total %d", len(req.Payload.Students)))
	students, err := toPbStudents(req.Payload.Students, logger)
	if err != nil {
		return nil, fmt.Errorf("toPbStudents: %w", err)
	}

	j.Logger.Info(fmt.Sprintf("req.Payload.Staffs total %d", len(req.Payload.Staffs)))
	staffs, err := toPbStaffs(req.Payload.Staffs, logger)
	if err != nil {
		return nil, fmt.Errorf("toPbStaffs: %w", err)
	}

	return &npb.EventUserRegistration{
		Signature:  signature,
		RawPayload: payload,
		Timestamp: &timestamppb.Timestamp{
			Seconds: int64(req.Timestamp),
		},
		Students: students,
		Staffs:   staffs,
	}, nil
}

func (j *JPREPController) toEventSyncUserCourse(c *gin.Context) (*npb.EventSyncUserCourse, error) {
	payload := middlewares.PayloadFromContext(c)
	if len(payload) == 0 {
		return nil, fmt.Errorf("payload body is empty")
	}

	req := &dto.SyncUserCourseRequest{}
	if err := json.Unmarshal(payload, req); err != nil {
		return nil, err
	}

	if req.Timestamp == 0 {
		return nil, fmt.Errorf("timestamp is required")
	}

	signature := c.GetHeader(middlewares.JPREPHeaderKey)
	logger := j.Logger.With(zap.String("signature", signature))
	j.Logger.Info(fmt.Sprintf("req.Payload.StudentLessons total %d", len(req.Payload.StudentLessons)))
	studentLessons, err := toPbStudentLessons(req.Payload.StudentLessons, logger)
	if err != nil {
		return nil, fmt.Errorf("toPbStudentLessons: %w", err)
	}

	return &npb.EventSyncUserCourse{
		Signature:  signature,
		RawPayload: payload,
		Timestamp: &timestamppb.Timestamp{
			Seconds: int64(req.Timestamp),
		},
		StudentLessons: studentLessons,
	}, nil
}

func (j *JPREPController) toEventMasterRegistration(c *gin.Context) (*npb.EventMasterRegistration, error) {
	payload := middlewares.PayloadFromContext(c)
	if len(payload) == 0 {
		return nil, fmt.Errorf("payload body is empty")
	}

	req := &dto.MasterRegistrationRequest{}
	if err := json.Unmarshal(payload, req); err != nil {
		return nil, err
	}

	if req.Timestamp == 0 {
		return nil, fmt.Errorf("timestamp is required")
	}

	signature := c.GetHeader(middlewares.JPREPHeaderKey)
	logger := j.Logger.With(zap.String("signature", signature))
	j.Logger.Info(fmt.Sprintf("req.Payload.Lessons total %d", len(req.Payload.Lessons)))
	lessons, err := toPbLessons(req.Payload.Lessons, logger)
	if err != nil {
		return nil, fmt.Errorf("toPbLessons: %w", err)
	}

	j.Logger.Info(fmt.Sprintf("req.Payload.Courses total %d", len(req.Payload.Courses)))
	courses, err := toPbCourses(req.Payload.Courses, logger)
	if err != nil {
		return nil, fmt.Errorf("toPbCourses: %w", err)
	}

	j.Logger.Info(fmt.Sprintf("req.Payload.Classes total %d", len(req.Payload.Classes)))
	classes, err := toPbClasses(req.Payload.Classes, logger)
	if err != nil {
		return nil, fmt.Errorf("toPbClasses: %w", err)
	}

	j.Logger.Info(fmt.Sprintf("req.Payload.AcademicYears total %d", len(req.Payload.AcademicYears)))
	academicYears, err := toAcademicYears(req.Payload.AcademicYears, logger)
	if err != nil {
		return nil, fmt.Errorf("toAcademicYears: %w", err)
	}

	return &npb.EventMasterRegistration{
		Signature:  signature,
		RawPayload: payload,
		Timestamp: &timestamppb.Timestamp{
			Seconds: int64(req.Timestamp),
		},
		Courses:       courses,
		Classes:       classes,
		Lessons:       lessons,
		AcademicYears: academicYears,
	}, nil
}

func (j *JPREPController) logSyncData(ctx context.Context, payload []byte, signature string) (*entities.PartnerSyncDataLog, error) {
	e := &entities.PartnerSyncDataLog{}
	database.AllNullEntity(e)
	err := multierr.Combine(
		e.PartnerSyncDataLogID.Set(idutil.ULIDNow()),
		e.Signature.Set(signature),
		e.Payload.Set(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("err Set PartnerSyncDataLog: %w", err)
	}
	err = j.PartnerSyncDataLogRepo.Create(ctx, j.DB, e)
	if err != nil {
		return nil, fmt.Errorf("err PartnerSyncDataLogRepo.Create: %w", err)
	}
	return e, nil
}

func (j *JPREPController) logSyncDataSplit(ctx context.Context, logStruct *LogStructure) (*entities.PartnerSyncDataLogSplit, error) {
	e := &entities.PartnerSyncDataLogSplit{}
	database.AllNullEntity(e)
	err := multierr.Combine(
		e.PartnerSyncDataLogSplitID.Set(idutil.ULIDNow()),
		e.PartnerSyncDataLogID.Set(logStruct.PartnerSyncDataLogID),
		e.Payload.Set(logStruct.Payload),
		e.Kind.Set(logStruct.Kind),
		e.Status.Set(string(entities.StatusPending)),
		e.RetryTimes.Set(0),
	)
	if err != nil {
		return nil, fmt.Errorf("err Set PartnerSyncDataLogSplit: %w", err)
	}
	err = j.PartnerSyncDataLogSplitRepo.Create(ctx, j.DB, e)
	if err != nil {
		return nil, fmt.Errorf("err PartnerSyncDataLogSplitRepo.Create: %w", err)
	}
	return e, nil
}

func validateAndParseDate(fromDate, toDate string) (from, to pgtype.Date, err error) {
	if fromDate == "" || toDate == "" {
		return from, to, fmt.Errorf("from_date and to_date is required")
	}
	start, err := time.Parse(layout, fromDate)
	if err != nil {
		return from, to, err
	}
	end, err := time.Parse(layout, toDate)
	if err != nil {
		return from, to, err
	}

	if int32(end.Sub(start).Hours()) > limitDateRange {
		return from, to, fmt.Errorf("Please choose a period less than or equal to 30 days")
	}
	if end.Sub(start).Hours() < 0 {
		return from, to, fmt.Errorf("Start date must come before End date")
	}

	var startDate, endDate pgtype.Date
	err = multierr.Combine(
		startDate.Set(start),
		endDate.Set(end),
	)
	if err != nil {
		return from, to, err
	}
	return startDate, endDate, nil
}

func toPbLessons(lessons []dto.Lesson, logger *zap.Logger) ([]*npb.EventMasterRegistration_Lesson, error) {
	mapCount := map[dto.Action]int{}

	results := make([]*npb.EventMasterRegistration_Lesson, 0, len(lessons))
	for i, l := range lessons {
		lessonType := toLessonType(l.LessonType)
		if lessonType == cpb.LessonType_LESSON_TYPE_NONE {
			return nil, fmt.Errorf("payload.m_lesson[%d].lesson_type should be oneof [online, offline, hybrid]", i)
		}

		action := toPbActionKind(l.ActionKind)
		if action == npb.ActionKind_ACTION_KIND_NONE {
			return nil, fmt.Errorf("payload.m_lesson[%d].action_kind should be oneof [deleted, upserted]", i)
		}

		mapCount[l.ActionKind]++

		if l.LessonID == 0 {
			return nil, fmt.Errorf("payload.m_lesson[%d].m_lesson_id is required", i)
		}

		if l.CourseID == 0 {
			return nil, fmt.Errorf("payload.m_lesson[%d].m_course_name_id is required", i)
		}

		if l.ClassName == "" {
			return nil, fmt.Errorf("payload.m_lesson[%d].m_class_name is required", i)
		}

		if action == npb.ActionKind_ACTION_KIND_UPSERTED {
			if l.StartDatetime == 0 {
				return nil, fmt.Errorf("payload.m_lesson[%d].start_datetime is required", i)
			}

			if l.EndDatetime == 0 {
				return nil, fmt.Errorf("payload.m_lesson[%d].end_datetime is required", i)
			}
		}

		results = append(results, &npb.EventMasterRegistration_Lesson{
			ActionKind: action,
			CourseId:   toJprepCourseID(l.CourseID),
			LessonId:   toJprepLessonID(l.LessonID),
			StartDate: &timestamppb.Timestamp{
				Seconds: int64(l.StartDatetime),
			},
			EndDate: &timestamppb.Timestamp{
				Seconds: int64(l.EndDatetime),
			},
			LessonGroup: l.Week,
			ClassName:   l.ClassName,
			LessonType:  lessonType,
		})
	}

	logger.Info("parsed lessons", zap.Any("stats", mapCount))
	return results, nil
}

func toPbStudentLessons(studentLessons []dto.StudentLesson, logger *zap.Logger) ([]*npb.EventSyncUserCourse_StudentLesson, error) {
	mapCount := map[dto.Action]int{}

	results := make([]*npb.EventSyncUserCourse_StudentLesson, 0, len(studentLessons))
	for i, s := range studentLessons {
		action := toPbActionKind(s.ActionKind)
		if action != npb.ActionKind_ACTION_KIND_UPSERTED {
			return nil, fmt.Errorf("payload.student_lesson[%d].action_kind should be upserted", i)
		}

		mapCount[s.ActionKind]++

		if s.StudentID == "" {
			return nil, fmt.Errorf("payload.student_id is required")
		}

		lessonIDs := make([]string, 0, len(s.LessonIDs))
		for _, l := range s.LessonIDs {
			lessonIDs = append(lessonIDs, toJprepLessonID(l))
		}

		results = append(results, &npb.EventSyncUserCourse_StudentLesson{
			ActionKind: action,
			StudentId:  s.StudentID,
			LessonIds:  lessonIDs,
		})
	}

	logger.Info("parsed studentLessons", zap.Any("stats", mapCount))
	return results, nil
}

func toPbActionKind(s dto.Action) npb.ActionKind {
	switch s {
	case dto.ActionKindUpserted:
		return npb.ActionKind_ACTION_KIND_UPSERTED
	case dto.ActionKindDeleted:
		return npb.ActionKind_ACTION_KIND_DELETED
	default:
		return npb.ActionKind_ACTION_KIND_NONE
	}
}

func toJprepAcedemicYearID(v int) string {
	return toJprepID("ACADEMIC_YEAR", v)
}

func toJprepCourseID(v int) string {
	return toJprepID("COURSE", v)
}

func toJprepLessonID(v int) string {
	return toJprepID("LESSON", v)
}

func toJprepID(typeID string, v int) string {
	return fmt.Sprintf("JPREP_%s_%09d", typeID, v)
}

func toPbCourses(courses []dto.Course, logger *zap.Logger) ([]*npb.EventMasterRegistration_Course, error) {
	mapCount := map[dto.Action]int{}

	results := make([]*npb.EventMasterRegistration_Course, 0, len(courses))
	for i, c := range courses {
		mapCount[c.ActionKind]++

		action := toPbActionKind(c.ActionKind)
		if action == npb.ActionKind_ACTION_KIND_NONE {
			return nil, fmt.Errorf("payload.m_course_name[%d].action_kind should be oneof [deleted, upserted]", i)
		}

		if c.CourseID == 0 {
			return nil, fmt.Errorf("payload.m_course_name[%d].m_course_name_id is required", i)
		}

		if action == npb.ActionKind_ACTION_KIND_UPSERTED {
			if c.CourseName == "" {
				return nil, fmt.Errorf("payload.m_course_name[%d].course_name is required", i)
			}
		}

		status := cpb.CourseStatus_COURSE_STATUS_ACTIVE
		if c.CourseStudentDivID != dto.CourseIDKid && c.CourseStudentDivID != dto.CourseIDAPlus {
			status = cpb.CourseStatus_COURSE_STATUS_INACTIVE
		}

		course := &npb.EventMasterRegistration_Course{
			ActionKind: action,
			CourseId:   toJprepCourseID(c.CourseID),
			CourseName: c.CourseName,
			Status:     status,
		}

		results = append(results, course)
	}

	logger.Info("parsed courses", zap.Any("stats", mapCount))
	return results, nil
}

func toPbClasses(classes []dto.Class, logger *zap.Logger) ([]*npb.EventMasterRegistration_Class, error) {
	mapCount := map[dto.Action]int{}

	results := make([]*npb.EventMasterRegistration_Class, 0, len(classes))
	for i, c := range classes {
		mapCount[c.ActionKind]++

		action := toPbActionKind(c.ActionKind)
		if action == npb.ActionKind_ACTION_KIND_NONE {
			return nil, fmt.Errorf("payload.m_regular_course[%d].action_kind should be oneof [deleted, upserted]", i)
		}

		if c.ClassID == 0 {
			return nil, fmt.Errorf("payload.m_regular_course[%d].m_course_id is required", i)
		}

		if c.CourseID == 0 {
			return nil, fmt.Errorf("payload.m_regular_course[%d].m_course_name_id is required", i)
		}

		var startDate, endDate time.Time
		if action == npb.ActionKind_ACTION_KIND_UPSERTED {
			if c.ClassName == "" {
				return nil, fmt.Errorf("payload.m_regular_course[%d].class_name is required", i)
			}

			if c.StartDate == "" {
				return nil, fmt.Errorf("payload.m_regular_course[%d].startdate is required", i)
			}

			if c.EndDate == "" {
				return nil, fmt.Errorf("enddate is required")
			}

			var err error
			startDate, err = time.Parse("2006/01/02", c.StartDate)
			if err != nil {
				return nil, fmt.Errorf("payload.m_regular_course[%d].startdate time.Parse: %w", i, err)
			}
			// convert JST to UTC
			startDate = convJSTStartDate(startDate)
			endDate, err = time.Parse("2006/01/02", c.EndDate)
			if err != nil {
				return nil, fmt.Errorf("payload.m_regular_course[%d].enddate time.Parse: %w", i, err)
			}
			// convert JST to UCT, sub a second to make sure the end date is not the beginning of a day
			endDate = convJSTEndDate(endDate)
		}

		course := &npb.EventMasterRegistration_Class{
			ActionKind: action,
			ClassName:  c.ClassName,
			ClassId:    uint64(c.ClassID),
			CourseId:   toJprepCourseID(c.CourseID),
			StartDate: &timestamppb.Timestamp{
				Seconds: startDate.Unix(),
			},
			EndDate: &timestamppb.Timestamp{
				Seconds: endDate.Unix(),
			},
			AcademicYearId: toJprepAcedemicYearID(c.AcademicYearID),
		}

		results = append(results, course)
	}

	logger.Info("parsed classes", zap.Any("stats", mapCount))
	return results, nil
}

func toAcademicYears(academicYears []dto.AcademicYear, logger *zap.Logger) ([]*npb.EventMasterRegistration_AcademicYear, error) {
	mapCount := map[dto.Action]int{}

	results := make([]*npb.EventMasterRegistration_AcademicYear, 0, len(academicYears))
	for i, a := range academicYears {
		mapCount[a.ActionKind]++

		action := toPbActionKind(a.ActionKind)
		if action == npb.ActionKind_ACTION_KIND_NONE {
			return nil, fmt.Errorf("payload.m_academic_year[%d].action_kind should be oneof [deleted, upserted]", i)
		}

		if a.AcademicYearID == 0 {
			return nil, fmt.Errorf("payload.m_academic_year[%d].m_academic_year_id is required", i)
		}

		if action == npb.ActionKind_ACTION_KIND_UPSERTED {
			if a.Name == "" {
				return nil, fmt.Errorf("payload.m_academic_year[%d].year_name is required", i)
			}

			if a.StartYearDate == 0 {
				return nil, fmt.Errorf("payload.m_academic_year[%d].start_year_date is required", i)
			}

			if a.EndYearDate == 0 {
				return nil, fmt.Errorf("payload.m_academic_year[%d].end_year_date is required", i)
			}
		}

		academicYear := &npb.EventMasterRegistration_AcademicYear{
			ActionKind:     action,
			AcademicYearId: toJprepAcedemicYearID(a.AcademicYearID),
			Name:           a.Name,
			StartYearDate: &timestamppb.Timestamp{
				Seconds: a.StartYearDate,
			},
			EndYearDate: &timestamppb.Timestamp{
				Seconds: a.EndYearDate,
			},
		}

		results = append(results, academicYear)
	}

	logger.Info("parsed academicYears", zap.Any("stats", mapCount))
	return results, nil
}

func toPbStudents(students []dto.Student, logger *zap.Logger) ([]*npb.EventUserRegistration_Student, error) {
	mapCount := map[dto.Action]int{}

	results := make([]*npb.EventUserRegistration_Student, 0, len(students))
	for i, s := range students {
		mapCount[s.ActionKind]++

		action := toPbActionKind(s.ActionKind)
		if action == npb.ActionKind_ACTION_KIND_NONE {
			return nil, fmt.Errorf("payload.m_student[%d].action_kind in should be oneof [deleted, upserted]", i)
		}

		if s.StudentID == "" {
			return nil, fmt.Errorf("payload.m_student[%d].student_id is required", i)
		}

		studentDivs := []int64{}
		if action == npb.ActionKind_ACTION_KIND_UPSERTED {
			for j, d := range s.StudentDivs {
				if d.MStudentDivID == 0 {
					return nil, fmt.Errorf("payload.m_student[%d].student_divs[%d].m_student_div_id is required", i, j)
				}

				studentDivs = append(studentDivs, int64(d.MStudentDivID))
			}

			if s.LastName == "" {
				return nil, fmt.Errorf("payload.m_student[%d].last_name is required", i)
			}

			if s.GivenName == "" {
				return nil, fmt.Errorf("payload.m_student[%d].given_name is required", i)
			}
		}

		packages := make([]*npb.EventUserRegistration_Student_Package, 0, len(s.Regularcourses))
		for j, p := range s.Regularcourses {
			if p.ClassID == 0 {
				return nil, fmt.Errorf("payload.m_student[%d].regularcourses[%d].m_course_id is required", i, j)
			}

			if p.Startdate == "" {
				return nil, fmt.Errorf("payload.m_student[%d].regularcourses[%d].startdate is required", i, j)
			}

			if p.Enddate == "" {
				return nil, fmt.Errorf("payload.m_student[%d].regularcourses[%d].enddate is required", i, j)
			}

			startDate, err := time.Parse("2006/01/02", p.Startdate)
			if err != nil {
				return nil, fmt.Errorf("payload.m_student[%d].startdate time.Parse: %w", i, err)
			}

			endDate, err := time.Parse("2006/01/02", p.Enddate)
			if err != nil {
				return nil, fmt.Errorf("payload.m_student[%d].enddate time.Parse: %w", i, err)
			}

			startDate = convJSTStartDate(startDate)
			endDate = convJSTEndDate(endDate)

			packages = append(packages, &npb.EventUserRegistration_Student_Package{
				ClassId: int64(p.ClassID),
				StartDate: &timestamppb.Timestamp{
					Seconds: startDate.Unix(),
				},
				EndDate: &timestamppb.Timestamp{
					Seconds: endDate.Unix(),
				},
			})
		}

		results = append(results, &npb.EventUserRegistration_Student{
			ActionKind:  action,
			StudentId:   s.StudentID,
			StudentDivs: studentDivs,
			LastName:    s.LastName,
			GivenName:   s.GivenName,
			Packages:    packages,
		})
	}

	logger.Info("parsed students", zap.Any("stats", mapCount))
	return results, nil
}

func toPbStaffs(staffs []dto.Staff, logger *zap.Logger) ([]*npb.EventUserRegistration_Staff, error) {
	mapCount := map[dto.Action]int{}

	results := make([]*npb.EventUserRegistration_Staff, 0, len(staffs))
	for i, s := range staffs {
		mapCount[s.ActionKind]++

		action := toPbActionKind(s.ActionKind)
		if action == npb.ActionKind_ACTION_KIND_NONE {
			return nil, fmt.Errorf("payload.m_staff[%d].action_kind in should be oneof [deleted, upserted]", i)
		}

		if s.StaffID == "" {
			return nil, fmt.Errorf("payload.m_staff[%d].staff_id is required", i)
		}

		if action == npb.ActionKind_ACTION_KIND_UPSERTED {
			if s.Name == "" {
				return nil, fmt.Errorf("payload.m_staff[%d].name is required", i)
			}
		}

		results = append(results, &npb.EventUserRegistration_Staff{
			ActionKind: action,
			StaffId:    s.StaffID,
			Name:       s.Name,
		})
	}

	logger.Info("parsed staffs", zap.Any("stats", mapCount))
	return results, nil
}

func toLessonType(l string) cpb.LessonType {
	switch l {
	case "online":
		return cpb.LessonType_LESSON_TYPE_ONLINE
	case "offline":
		return cpb.LessonType_LESSON_TYPE_OFFLINE
	case "hybrid":
		return cpb.LessonType_LESSON_TYPE_HYBRID
	default:
		return cpb.LessonType_LESSON_TYPE_NONE
	}
}

//convJSTStartDate convert JST start date to UTC start date
func convJSTStartDate(startDate time.Time) time.Time {
	return startDate.Add(-time.Hour * 9)
}

// convJSTEndDate convert JST end date to UCT end date, sub a second to make sure the end date is not the beginning of a day
func convJSTEndDate(endDate time.Time) time.Time {
	return endDate.Add(time.Hour*15 - time.Second)
}
