UPDATE public.lessons
SET     teaching_method = CASE  
                            WHEN teaching_model = 'LESSON_TEACHING_MODEL_INDIVIDUAL' THEN 'LESSON_TEACHING_METHOD_INDIVIDUAL' 
                            WHEN teaching_model = 'LESSON_TEACHING_MODEL_GROUP' THEN 'LESSON_TEACHING_METHOD_GROUP' 
                            ELSE teaching_model
                        END 
WHERE teaching_model IN ('LESSON_TEACHING_MODEL_INDIVIDUAL', 'LESSON_TEACHING_MODEL_GROUP') and teaching_method is null;


UPDATE public.lessons
SET   teaching_medium = CASE  
                            WHEN lesson_type = 'LESSON_TYPE_ONLINE' THEN 'LESSON_TEACHING_MEDIUM_ONLINE' 
                            WHEN lesson_type = 'LESSON_TYPE_OFFLINE' THEN 'LESSON_TEACHING_MEDIUM_OFFLINE' 
                            WHEN lesson_type = 'LESSON_TYPE_HYBRID' THEN 'LESSON_TEACHING_MEDIUM_HYBRID' 
                            ELSE lesson_type
                        END 
WHERE lesson_type IN ('LESSON_TYPE_ONLINE', 'LESSON_TYPE_OFFLINE', 'LESSON_TYPE_HYBRID') and teaching_medium is null;

UPDATE public.lessons SET scheduling_status = 'LESSON_SCHEDULING_STATUS_PUBLISHED';
