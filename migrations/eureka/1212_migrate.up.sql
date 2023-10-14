update individual_study_plan isp 
set study_plan_id = temp_table.master_study_plan_id 
from (
select isp.study_plan_id,ssp.master_study_plan_id, isp.learning_material_id, isp.student_id  from individual_study_plan isp join student_study_plans ssp on isp.study_plan_id = ssp.study_plan_id
left join individual_study_plan isp2 on isp2.study_plan_id = ssp.master_study_plan_id and isp2.learning_material_id = isp.learning_material_id and isp2.student_id = isp.student_id 
where ssp.master_study_plan_id is not null and isp2.study_plan_id is null and isp2.student_id is null and isp2.learning_material_id is null
) temp_table
where  isp.study_plan_id = temp_table.study_plan_id and isp.learning_material_id = temp_table.learning_material_id and isp.student_id = temp_table.student_id;