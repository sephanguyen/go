INSERT INTO public.configs
(config_key, config_group, country, config_value, updated_at, created_at, deleted_at, resource_path)
VALUES
('blacklistedCourseIDs', 'lesson', 'COUNTRY_MASTER', 
'JPREP_COURSE_000000312,JPREP_COURSE_000000313,JPREP_COURSE_000000314,
JPREP_COURSE_000000315,JPREP_COURSE_000000316,JPREP_COURSE_000000317,
JPREP_COURSE_000000318,JPREP_COURSE_000000319,JPREP_COURSE_000000320,
JPREP_COURSE_000000321,JPREP_COURSE_000000322,JPREP_COURSE_000000323,
JPREP_COURSE_000000324,JPREP_COURSE_000000325,JPREP_COURSE_000000326,
JPREP_COURSE_000000327,JPREP_COURSE_000000328,JPREP_COURSE_000000329,
JPREP_COURSE_000000330,JPREP_COURSE_000000331,JPREP_COURSE_000000332,
JPREP_COURSE_000000333,JPREP_COURSE_000000334,JPREP_COURSE_000000335,
JPREP_COURSE_000000336,JPREP_COURSE_000000337,JPREP_COURSE_000000338,
JPREP_COURSE_000000339,JPREP_COURSE_000000340,JPREP_COURSE_000000341,
JPREP_COURSE_000000342,JPREP_COURSE_000000343,JPREP_COURSE_000000344,
JPREP_COURSE_000000345,JPREP_COURSE_000000346,JPREP_COURSE_000000347'
, NOW(), NOW(), NULL, '-2147483647')
ON CONFLICT DO NOTHING;
