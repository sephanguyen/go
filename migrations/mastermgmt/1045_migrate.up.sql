UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'lesson.assigned_student_list' and resource_path  IN ('-2147483648','-2147483646','-2147483645','-2147483643','-2147483642','-2147483641','-2147483640','-2147483639','-2147483638');