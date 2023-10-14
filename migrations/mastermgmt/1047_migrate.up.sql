--  ['-2147483635': KEC-DEMO], ['-2147483642': KEC], ['-2147483641': AIC]

-- Create configuration_key and trigger rule create internal_configuration_value for all partners with value is 'true'
INSERT INTO configuration_key (config_key, value_type, default_value, configuration_type, created_at, updated_at)
VALUES ('syllabus.grade_book.management', 'boolean', 'true', 'CONFIGURATION_TYPE_INTERNAL', NOW(), NOW());

-- Disable syllabus.grade_book.management for KEC, AIC
UPDATE internal_configuration_value
SET config_value = 'false'
WHERE config_key = 'syllabus.grade_book.management' and resource_path IN ('-2147483635', '-2147483641', '-2147483642');

-- Update default value of syllabus.grade_book.management to false
UPDATE configuration_key
SET default_value = 'false', updated_at = now()
WHERE config_key = 'syllabus.grade_book.management';
