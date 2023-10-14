CREATE OR REPLACE FUNCTION not_update_lesson_lock() RETURNS trigger
   LANGUAGE plpgsql AS
$$BEGIN
	IF new.is_locked = false THEN
	   RETURN NEW;
	END IF;
   IF OLD.is_locked = true THEN
   	  IF new.scheduler_id <> old.scheduler_id THEN
		old.scheduler_id = new.scheduler_id;
	  END IF;
      RETURN OLD;
   END IF;
   RETURN NEW;
END;$$;