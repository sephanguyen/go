ALTER TABLE public.academic_week ALTER column week_order SET NOT NULL;

CREATE OR REPLACE FUNCTION calculate_purchased_slot_total(freq int, a_start_date date,a_end_date date,a_course_id text,a_location_id text)  
RETURNS INT  
LANGUAGE plpgsql  
as  
$$  
DECLARE  
   weeks integer; 
   frequency integer;
   academic_weeks text[];
BEGIN 
   SELECT c.frequency, c.academic_weeks 
   INTO frequency,academic_weeks 
   FROM course_location_schedule c WHERE c.course_id = a_course_id AND c.location_id = a_location_id;
   freq := COALESCE(freq, frequency);
   SELECT ARRAY_LENGTH(ARRAY(
 			 SELECT unnest(ARRAY(SELECT week_order::text FROM academic_week WHERE location_id = a_location_id AND start_date between a_start_date and a_end_date)) 
			 INTERSECT
             SELECT unnest(academic_weeks)),1)
   INTO weeks;
   RETURN COALESCE(weeks * freq, 0);
END;  
$$;
