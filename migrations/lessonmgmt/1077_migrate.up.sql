create OR REPLACE FUNCTION not_update_related_lesson_lock() RETURNS trigger
   LANGUAGE plpgsql AS
$$BEGIN
   IF EXISTS (SELECT 1 FROM lessons l WHERE l.lesson_id = OLD.lesson_id AND l.is_locked = true) THEN
      RETURN OLD;
END IF;
RETURN NEW;
END;$$;

create OR REPLACE FUNCTION not_update_related_lesson_lock_lesson_member() RETURNS trigger
   LANGUAGE plpgsql AS
$$BEGIN
   IF EXISTS (SELECT 1 FROM lessons l WHERE l.lesson_id = OLD.lesson_id AND l.is_locked = true) THEN
         IF old.attendance_status = 'STUDENT_ATTEND_STATUS_ABSENT' and new.attendance_status = 'STUDENT_ATTEND_STATUS_REALLOCATE' THEN
            old.attendance_status = new.attendance_status;
            old.updated_at = new.updated_at;
END IF;
RETURN OLD;
END IF;
RETURN NEW;
END;$$;

DROP TRIGGER IF EXISTS not_update_lesson_member_when_lesson_lock on public.lesson_members;

create trigger not_update_lesson_member_when_lesson_lock before
    update
    on
        public.lesson_members for each row execute function not_update_related_lesson_lock_lesson_member();
