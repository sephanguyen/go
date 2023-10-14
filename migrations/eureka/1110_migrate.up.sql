create or replace function check_study_plan_item_time(master_updated_at timestamptz, 
                      student_updated_at timestamptz,
                      master_time timestamptz,
                      student_time timestamptz) 
RETURNS timestamptz LANGUAGE 'plpgsql' SECURITY invoker
AS $$
begin
    case 
      when master_updated_at is null or student_updated_at is null
          then return coalesce(master_time, student_time);
      else 
        case
              when master_updated_at >= student_updated_at 
                then return master_time;
              else return student_time;
        end case;
    end case;
end;
$$;
 
create or replace view student_study_plans_view as
select  ssp.student_id,
		m.study_plan_id,
		m.book_id,
		m.chapter_id,
		m.chapter_display_order,
		m.topic_id,
		m.topic_display_order,
		m.learning_material_id ,
		m.lm_display_order,
		check_study_plan_item_time(m.updated_at, isp.updated_at, m.start_date, isp.start_date) as start_date,
		check_study_plan_item_time(m.updated_at, isp.updated_at, m.end_date, isp.end_date) as end_date,
		check_study_plan_item_time(m.updated_at, isp.updated_at, m.available_from, isp.available_from) as available_from,
		check_study_plan_item_time(m.updated_at, isp.updated_at, m.available_to, isp.available_to) as available_to,
		check_study_plan_item_time(m.updated_at, isp.updated_at, m.school_date, isp.school_date) as school_date,
		check_study_plan_item_time(m.updated_at, isp.updated_at, m.updated_at, isp.updated_at) as updated_at,
		case 
			when m.updated_at is null or isp.updated_at is null then coalesce(m.status, isp.status)
		else case
				when m.updated_at >= isp.updated_at then m.status
				else isp.status
			end
		end as status
from master_study_plan_view m
join student_study_plans ssp  on (m.study_plan_id = ssp.master_study_plan_id or (ssp.master_study_plan_id is null and m.study_plan_id=ssp.study_plan_id))
left join individual_study_plan isp on ssp.student_id  = isp.student_id  and m.learning_material_id = isp.learning_material_id and m.study_plan_id = isp.study_plan_id