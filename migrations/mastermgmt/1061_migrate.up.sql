-- Remove config key is not used syllabus.grade_book.is_enabled
DELETE FROM internal_configuration_value WHERE config_key = 'syllabus.grade_book.is_enabled';
DELETE FROM external_configuration_value WHERE config_key = 'syllabus.grade_book.is_enabled';
DELETE FROM configuration_key WHERE config_key = 'syllabus.grade_book.is_enabled';

-- Update default value of syllabus.grade_book.management
UPDATE configuration_key SET default_value = 'true', updated_at = now() WHERE config_key = 'syllabus.grade_book.management';
UPDATE configuration_key SET default_value = 'on', updated_at = now() WHERE config_key = 'syllabus.learning_material.content_lo';
