ALTER TABLE ONLY public.course_location_schedule
    ADD COLUMN IF NOT EXISTS "academic_weeks" TEXT[];

ALTER TABLE course_location_schedule  
DROP COLUMN academic_week;

CREATE UNIQUE INDEX course_location_schedule_idx ON course_location_schedule (course_id, location_id);

ALTER TABLE course_location_schedule 
ADD CONSTRAINT unique_course_location_schedule
UNIQUE
USING INDEX course_location_schedule_idx;