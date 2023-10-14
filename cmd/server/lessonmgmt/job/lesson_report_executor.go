package job

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	lesson_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/application/services"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/application/services/form_partner"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/infrastructure"
	lesson_report_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/infrastructure/repo"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

const LESSON_REPORT_EXECUTOR = LabelExecutor("LESSON_REPORT_EXECUTOR")

type RWMap struct {
	sync.RWMutex
	m map[string]form_partner.EvictionPartner
}

func (r *RWMap) Get(key string) form_partner.EvictionPartner {
	r.RLock()
	defer r.RUnlock()
	return r.m[key]
}

func (r *RWMap) Set(key string, val form_partner.EvictionPartner) {
	r.Lock()
	defer r.Unlock()
	r.m[key] = val
}

type LessonReportExecutor struct {
	service               *services.UpdaterIndividualLessonReport
	db                    database.Ext
	logger                *zap.Logger
	lessonReportRepo      infrastructure.LessonReportRepo
	partnerFormConfigRepo infrastructure.PartnerFormConfigRepo
	mapEvictionPartner    *RWMap
}

func InitLessonReportExecutor(db database.Ext, logger *zap.Logger) *LessonReportExecutor {
	lessonReportRepo := &lesson_report_repo.LessonReportRepo{}
	partnerFormConfigRepo := &lesson_report_repo.PartnerFormConfigRepo{}
	return &LessonReportExecutor{
		db:                    db,
		logger:                logger,
		lessonReportRepo:      lessonReportRepo,
		partnerFormConfigRepo: partnerFormConfigRepo,
		mapEvictionPartner:    &RWMap{m: make(map[string]form_partner.EvictionPartner)},
		service: &services.UpdaterIndividualLessonReport{
			DB:                     db,
			LessonReportRepo:       lessonReportRepo,
			LessonReportDetailRepo: &lesson_report_repo.LessonReportDetailRepo{},
			PartnerFormConfigRepo:  partnerFormConfigRepo,
			LessonMemberRepo:       &lesson_repo.LessonMemberRepo{},
		},
	}
}

var MapResourceIdOfBestcoAndRenseikai = map[string]bool{
	"-2147483645": true,
	"-2147483644": true, // different resource path bestco uat and production
	"-2147483648": true, // test bestco
}

func (l *LessonReportExecutor) GetTotal(ctx context.Context) (int, error) {
	query := `SELECT COUNT(1) FROM lesson_reports l
				JOIN partner_form_configs pfc 
				ON l.form_config_id = pfc.form_config_id 
					AND pfc.feature_name = 'FEATURE_NAME_INDIVIDUAL_LESSON_REPORT'
					AND pfc.deleted_at IS NULL
				WHERE l.deleted_at IS NULL
				AND l.resource_path = $1`

	var totalLesson int
	if err := l.db.QueryRow(ctx, query, golibs.ResourcePathFromCtx(ctx)).Scan(&totalLesson); err != nil && err != pgx.ErrNoRows {
		return 0, fmt.Errorf("row.Scan: %w", err)
	}
	return totalLesson, nil
}

func (l *LessonReportExecutor) ExecuteJob(ctx context.Context, limit int, offSet int) error {
	resourceId := golibs.ResourcePathFromCtx(ctx)
	// trick allways pre set offset = 0 for get data because form config id will change after run
	offSet = 0
	lessonReports, err := l.lessonReportRepo.FindByResourcePath(ctx, l.db, golibs.ResourcePathFromCtx(ctx), limit, offSet)
	if err != nil {
		return fmt.Errorf("fail Sync Lesson Report in org: %s, offset: %d, limit: %d: %s", golibs.ResourcePathFromCtx(ctx), offSet, limit, err)
	}
	evictionPartner := l.mapEvictionPartner.Get(resourceId)
	for _, v := range lessonReports {
		err = l.service.Update(evictionPartner, ctx, v)
		if err != nil {
			return fmt.Errorf("fail update lesson report: %s", err)
		}
	}

	l.logger.Info(fmt.Sprintf("the total of sync lessons report success in org: %s, offset: %d, limit: %d: %d", golibs.ResourcePathFromCtx(ctx), offSet, limit, len(lessonReports)))
	return nil
}

func (l *LessonReportExecutor) PreExecute(ctx context.Context) error {
	resourceId := golibs.ResourcePathFromCtx(ctx)
	var evictionPartner form_partner.EvictionPartner
	formPartner := form_partner.InitFormPartner(resourceId)
	evictionPartner = &form_partner.AllPartner{FormPartner: formPartner}
	if MapResourceIdOfBestcoAndRenseikai[resourceId] {
		evictionPartner = &form_partner.ReseikaiAndBestcoPartner{FormPartner: formPartner}
	}

	numResourceId, _ := strconv.Atoi(resourceId)
	now := time.Now()
	partnerFormConfig := &domain.PartnerFormConfig{
		FormConfigID:   evictionPartner.GetConfigFormId(),
		PartnerID:      numResourceId,
		FeatureName:    evictionPartner.GetConfigFormName(),
		CreatedAt:      now,
		UpdatedAt:      now,
		FormConfigData: []byte(evictionPartner.GetConfigForm()),
	}
	l.mapEvictionPartner.Set(resourceId, evictionPartner)
	return l.partnerFormConfigRepo.CreatePartnerFormConfig(ctx, l.db, partnerFormConfig)
}
