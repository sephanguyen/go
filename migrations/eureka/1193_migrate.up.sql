drop function if exists list_available_learning_material();

create or replace function list_available_learning_material()
    returns table
            (
                student_id            text,
                study_plan_id         text,
                book_id               text,
                chapter_id            text,
                chapter_display_order smallint,
                topic_id              text,
                topic_display_order   smallint,
                learning_material_id  text,
                lm_display_order      smallint,
                available_from 		  timestamp with time zone,
			    available_to	 	  timestamp with time zone,
			    start_date			  timestamp with time zone,
			    end_date			  timestamp with time zone,
			    status				  text,
			    school_date 		  timestamp with time zone
            )
    language sql
    stable
as
$$
select student_id,
       study_plan_id,
       book_id,
       chapter_id,
       chapter_display_order,
       topic_id,
       topic_display_order,
       learning_material_id,
       lm_display_order,
       available_from,
       available_to,
       start_date,
       end_date,
       status,
       school_date
from individual_study_plan_fn()
where (now() between available_from and available_to)
   or (available_from <= now() and available_to is null)
$$;