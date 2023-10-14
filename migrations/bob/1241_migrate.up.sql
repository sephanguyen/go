UPDATE public.lessons 
SET scheduling_status = 'LESSON_SCHEDULING_STATUS_PUBLISHED'
WHERE scheduling_status is null or scheduling_status = '';