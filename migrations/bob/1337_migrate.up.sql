-- start create function ---

CREATE OR REPLACE FUNCTION not_update_lesson_lock() RETURNS trigger
   LANGUAGE plpgsql AS
$$BEGIN
	IF new.is_locked = false THEN
	   RETURN NEW;
	END IF;
   IF OLD.is_locked = true THEN
      RETURN OLD;
   END IF;
   RETURN NEW;
END;$$;


create OR REPLACE FUNCTION not_update_related_lesson_lock() RETURNS trigger
   LANGUAGE plpgsql AS
$$BEGIN
   IF EXISTS (SELECT 1 FROM lessons l WHERE l.lesson_id = OLD.lesson_id AND l.is_locked = true) THEN
      RETURN OLD;
   END IF;
   RETURN NEW;
END;$$;

-- end create function ---

-- start create trigger ---

DROP TRIGGER IF EXISTS NOT_UPDATE_LESSON_LOCK
  ON lessons;

CREATE TRIGGER NOT_UPDATE_LESSON_LOCK
    BEFORE UPDATE ON lessons FOR EACH ROW
    EXECUTE PROCEDURE not_update_lesson_lock();


DROP TRIGGER IF EXISTS NOT_UPDATE_LESSON_MEMBER_WHEN_LESSON_LOCK
  ON lesson_members;

CREATE trigger NOT_UPDATE_LESSON_MEMBER_WHEN_LESSON_LOCK
    BEFORE UPDATE ON lesson_members FOR EACH ROW
    EXECUTE PROCEDURE not_update_related_lesson_lock();


  
DROP TRIGGER IF EXISTS NOT_UPDATE_LESSON_TEACHER_WHEN_LESSON_LOCK
  ON lessons_teachers;

CREATE trigger NOT_UPDATE_LESSON_TEACHER_WHEN_LESSON_LOCK
    BEFORE UPDATE ON lessons_teachers FOR EACH ROW
    EXECUTE PROCEDURE not_update_related_lesson_lock();
    
 
DROP TRIGGER IF EXISTS NOT_UPDATE_LESSON_REPORT_WHEN_LESSON_LOCK
  ON lesson_reports;
CREATE trigger NOT_UPDATE_LESSON_REPORT_WHEN_LESSON_LOCK
    BEFORE UPDATE ON lesson_reports FOR EACH ROW
    EXECUTE PROCEDURE not_update_related_lesson_lock();
  
 -- end create trigger ---
