package helpers

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/conversationmgmt/common/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

func (helper *ConversationMgmtHelper) CreateGradeMasterForOrgazination(ctx context.Context, orgID string) ([]*entities.GradeMaster, error) {
	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: orgID,
		},
	})
	grades, err := helper.getGradeMasterByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if len(grades) > 0 {
		return grades, nil
	}

	grades = []*entities.GradeMaster{}
	for i := 0; i < 17; i++ {
		grade := &entities.GradeMaster{
			ID:                idutil.ULIDNow(),
			Name:              fmt.Sprint(i),
			PartnerInternalID: fmt.Sprintf("id-%d", i),
		}

		stmtGrade := `
			INSERT INTO public.grade
			("name", is_archived, updated_at, created_at, resource_path, grade_id, partner_internal_id, deleted_at, "sequence")
			VALUES($1, false, now(), now(), autofillresourcepath(), $2, $3, NULL, $4);
		`
		if _, err := helper.BobDBConn.Exec(ctx2, stmtGrade, grade.Name, grade.ID, grade.PartnerInternalID, i); err != nil {
			return nil, err
		}

		stmtGradeOrg := `
			INSERT INTO public.grade_organization
			(grade_organization_id, grade_id, grade_value, created_at, updated_at, deleted_at, resource_path)
			VALUES($1, $2, $3, now(), now(), NULL, autofillresourcepath());
		`
		if _, err := helper.BobDBConn.Exec(ctx2, stmtGradeOrg, idutil.ULIDNow(), grade.ID, i); err != nil {
			return nil, err
		}

		grades = append(grades, grade)
	}
	return grades, nil
}

func (helper *ConversationMgmtHelper) getGradeMasterByOrg(ctx context.Context, orgID string) ([]*entities.GradeMaster, error) {
	grades := []*entities.GradeMaster{}
	query := `
		SELECT grade_id, "name", partner_internal_id
		FROM public.grade
		WHERE resource_path = $1 AND deleted_at IS NULL
	`
	rows, err := helper.BobDBConn.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		grade := &entities.GradeMaster{}
		err = rows.Scan(&grade.ID, &grade.Name, &grade.PartnerInternalID)
		if err != nil {
			return nil, err
		}
		grades = append(grades, grade)
	}
	return grades, nil
}
