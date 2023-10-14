CREATE OR REPLACE FUNCTION not_update_lesson_lock() RETURNS trigger
   LANGUAGE plpgsql AS
$$BEGIN
   IF new.is_locked = false THEN
	  RETURN NEW;
   END IF;
   IF OLD.is_locked = true THEN
      IF old.attendance_status = 'STUDENT_ATTEND_STATUS_ABSENT' and new.attendance_status = 'STUDENT_ATTEND_STATUS_REALLOCATE' THEN
            old.attendance_status = new.attendance_status;
            old.updated_at = new.updated_at;
        END IF;
   	  IF new.scheduler_id <> old.scheduler_id THEN
		    old.scheduler_id = new.scheduler_id;
            old.updated_at = new.updated_at;
        END IF;
      RETURN OLD;
   END IF;
   RETURN NEW;
END;$$;
