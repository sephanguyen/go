CREATE OR REPLACE RULE LESSON_MEMBER_LOCK AS ON UPDATE TO lesson_members
    WHERE EXISTS (SELECT 1 FROM lessons l WHERE l.lesson_id = OLD.lesson_id AND l.is_locked = true)
    DO INSTEAD NOTHING;
   
CREATE OR REPLACE RULE LESSON_TEACHER_LOCK AS ON UPDATE TO lessons_teachers
    WHERE EXISTS (SELECT 1 FROM lessons l WHERE l.lesson_id = OLD.lesson_id AND l.is_locked = true)
    DO INSTEAD NOTHING;

CREATE OR REPLACE RULE LESSON_REPORT_LOCK AS ON UPDATE TO lesson_reports
    WHERE EXISTS (SELECT 1 FROM lessons l WHERE l.lesson_id = OLD.lesson_id AND l.is_locked = true)
    DO INSTEAD NOTHING;   
 
CREATE OR REPLACE RULE LESSON_LOCK AS ON UPDATE TO lessons
    WHERE old.is_locked = true
    DO INSTEAD NOTHING; 
