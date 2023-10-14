DROP FUNCTION calculate_purchased_slot_total;
CREATE OR REPLACE FUNCTION calculate_purchased_slot_total_v2(freq smallint, a_start_date date,a_end_date date,a_course_id text,a_location_id text, a_student_id text)  
	RETURNS smallint  
	LANGUAGE plpgsql  
	as  
	$$  
	DECLARE  
	   weeks SMALLINT; 
	   rec RECORD;
	   weeks_closed_day smallint := 0;
	   counts SMALLINT := 0;
	BEGIN    
	   SELECT  ARRAY_LENGTH(ARRAY(
	 			 SELECT unnest(ARRAY(SELECT week_order::text FROM academic_week WHERE location_id = a_location_id AND end_date >= a_start_date and start_date <= a_end_date)) 
				 INTERSECT
	             SELECT unnest(ARRAY(SELECT c.academic_weeks FROM course_location_schedule c WHERE c.course_id = a_course_id AND c.location_id = a_location_id))),1)
	   INTO weeks;

      IF EXISTS (select 1 from courses c join course_type ct on ct.course_type_id = c.course_type_id  where course_id = a_course_id and ct."name"  = 'Regular' ) THEN
      		for rec in  select s.start_date ,s.end_date
				   from scheduler s 
				   where s.freq  = 'weekly' and exists (select 1 from lessons l join lesson_members lm on lm.lesson_id = l.lesson_id where l.scheduler_id = s.scheduler_id
														and user_id = a_student_id and lm.course_id = a_course_id and l.deleted_at is null and lm.deleted_at is null) and s.deleted_at is null
	      	loop 
		    	select count(*)
		    	into counts
		    	from day_info where location_id = a_location_id and deleted_at is null 
	   						  and "date" in (select DATE(GENERATE_SERIES(rec.start_date at time zone time_zone, rec.end_date at time zone time_zone , '1 weeks'::INTERVAL)));
	   		 	weeks_closed_day := weeks_closed_day + counts;
	      	end loop;
      END IF;
	   RETURN COALESCE(weeks * freq - weeks_closed_day, 0);   
	END;   
	$$;
