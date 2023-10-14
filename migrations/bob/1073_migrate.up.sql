ALTER TABLE teachers
    DROP CONSTRAINT IF EXISTS teachers__teacher_id__fk;
ALTER TABLE teachers
    ADD CONSTRAINT teachers__teacher_id__fk FOREIGN KEY (teacher_id) REFERENCES users(user_id);
