-- users
CREATE INDEX user__external_id__not_null__idx ON users (user_external_id) WHERE user_external_id IS NOT NULL;

-- student_packages
CREATE INDEX student_packages__student_id__idx on student_packages (student_id);
