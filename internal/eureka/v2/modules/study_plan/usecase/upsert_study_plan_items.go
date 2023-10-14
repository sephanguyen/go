package usecase

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/database"
)

func (c *StudyPlanItemUseCase) UpsertStudyPlanItems(ctx context.Context, studyPlanItems []domain.StudyPlanItem) error {
	studyPlanItemDtos := []*dto.StudyPlanItemDto{}
	lmListDtos := []*dto.LmListDto{}
	for _, studyPlanItem := range studyPlanItems {
		eLmList := dto.LmListDto{}
		// always create new lmList
		lmListDto, err := eLmList.ToLmListEntity(domain.LmList{
			LmListID:  idutil.ULIDNow(),
			LmIDs:     studyPlanItem.LmList,
			CreatedAt: studyPlanItem.CreatedAt,
			UpdatedAt: studyPlanItem.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("error convert lmList: %w", err)
		}
		lmListDtos = append(lmListDtos, lmListDto)

		// must guarantee that study plan item list order is correspond to lmList
		e := dto.StudyPlanItemDto{}
		studyPlanItemDto, err := e.FromEntity(studyPlanItem)
		if err != nil {
			return fmt.Errorf("convertStudyPlanItemToDto: %w", err)
		}
		studyPlanItemDto.LmListID = lmListDto.LmListID
		studyPlanItemDtos = append(studyPlanItemDtos, studyPlanItemDto)
	}

	err := database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx database.Tx) error {
		txErr := c.LearningMaterialRepo.UpsertLearningMaterialsIDList(ctx, lmListDtos)
		if txErr != nil {
			return fmt.Errorf("StudyPlanItemRepo.UpsertStudyPlanItems: %w", txErr)
		}

		txErr = c.StudyPlanItemRepo.UpsertStudyPlanItems(ctx, studyPlanItemDtos)
		if txErr != nil {
			return fmt.Errorf("StudyPlanItemRepo.UpsertStudyPlanItems: %w", txErr)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecInTx: %w", err)
	}
	return nil
}
