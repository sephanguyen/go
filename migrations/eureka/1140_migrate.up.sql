CREATE OR REPLACE FUNCTION retrieve_study_plan_identity(study_plan_item_ids text[]) 
returns table (
    student_id text,
    lm_id text,
    study_plan_id text,
    study_plan_item_id text
) AS
$BODY$

select coalesce(ssp.master_study_plan_id, spi.study_plan_id) as study_plan_id,
       coalesce(nullif(content_structure ->>'lo_id', ''), content_structure->>'assignment_id') as lm_id,
       ssp.student_id,
       spi.study_plan_item_id
from study_plan_items spi
join student_study_plans ssp on spi.study_plan_id = ssp.study_plan_id
where study_plan_item_id = ANY(study_plan_item_ids)

$BODY$
LANGUAGE 'sql';
