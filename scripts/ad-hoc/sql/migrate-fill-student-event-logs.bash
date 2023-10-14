#!/bin/bash

# This script migrate StudyPlanItemIdentity for student_event_logs table.

set -euo pipefail

DB_NAME="eureka"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
BEGIN;
    LOCK TABLE public.student_event_logs IN EXCLUSIVE MODE;

    ALTER TABLE public.student_event_logs DISABLE TRIGGER fill_new_identity;

    UPDATE
        student_event_logs sel
    SET
        learning_material_id = sel.payload->>'lo_id',
        study_plan_id = (
            SELECT COALESCE(sp.master_study_plan_id, sp.study_plan_id)
            FROM study_plan_items spi
            JOIN study_plans sp ON sp.study_plan_id = spi.study_plan_id
            WHERE spi.study_plan_item_id = sel.study_plan_item_id
        )
    WHERE sel.event_type = ANY(ARRAY[
        'study_guide_finished',
        'video_finished',
        'learning_objective',
        'quiz_answer_selected'
    ]);

    ALTER TABLE public.student_event_logs ENABLE TRIGGER fill_new_identity;
COMMIT;
EOF
