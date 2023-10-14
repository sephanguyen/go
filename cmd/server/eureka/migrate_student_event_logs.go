package eureka

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
)

func init() {
	bootstrap.RegisterJob("eureka_update_student_event_logs_study_plan_item_identity", runUpdateStudentEventLogsStudyPlanItemIdentity).
		Desc("eureka update_student_event_logs study_plan_item_identity")
}

func runUpdateStudentEventLogsStudyPlanItemIdentity(ctx context.Context, _ configurations.Config, rsc *bootstrap.Resources) error {
	db := rsc.DBWith("eureka")

	start := time.Now()
	defer func() {
		fmt.Println("migration completed of student_event_logs: ", time.Since(start))
	}()

	if _, err := db.Exec(ctx, `
		DO $$
		DECLARE
			batch_size int := 10000;
			min_id bigint;
			max_id bigint;
		BEGIN
			SELECT max(student_event_log_id), min(student_event_log_id) INTO max_id, min_id FROM student_event_logs;
		
			FOR idx IN min_id..max_id BY batch_size LOOP
				UPDATE student_event_logs sel
				SET
					learning_material_id = sel.payload->>'lo_id',
					study_plan_id = (
						SELECT COALESCE(sp.master_study_plan_id, sp.study_plan_id)
						FROM study_plan_items spi
						JOIN study_plans sp ON sp.study_plan_id = spi.study_plan_id
						WHERE spi.study_plan_item_id = sel.study_plan_item_id
					)
				WHERE student_event_log_id >= idx AND student_event_log_id < idx+batch_size
					AND event_type = ANY(ARRAY[
						'study_guide_finished',
						'video_finished',
						'learning_objective',
						'quiz_answer_selected'
					]);
				RAISE INFO 'committing data from % to % at %', idx, idx+batch_size, now();
				COMMIT;
			END LOOP;
		END;
		$$;`,
	); err != nil {
		return fmt.Errorf("UpdateStudentEventLogsStudyPlanItemIdentity: database.Exec: %s", err)
	}
	return nil
}
