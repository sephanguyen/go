DROP FUNCTION retrieve_study_plan_identity;
CREATE OR REPLACE FUNCTION retrieve_study_plan_identity(study_plan_item_ids text[]) 
returns table (
    study_plan_id text, lm_id text, student_id text, study_plan_item_id text
) AS
$BODY$

select isp.study_plan_id,
       coalesce(nullif(content_structure ->>'lo_id', ''), content_structure->>'assignment_id') as lm_id,
       isp.student_id,
       spi.study_plan_item_id
from study_plan_items spi
join individual_study_plans_view isp on coalesce(nullif(spi.content_structure ->>'lo_id', ''), spi.content_structure->>'assignment_id') = isp.learning_material_id
where study_plan_item_id = ANY(study_plan_item_ids)

$BODY$
LANGUAGE 'sql';
