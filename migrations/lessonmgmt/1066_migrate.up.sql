CREATE INDEX lesson_reports_lesson_id_idx ON public.lesson_reports USING btree (lesson_id);
CREATE INDEX lessons__class_id__idx ON public.lessons USING btree (class_id);
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
create trigger not_update_lesson_lock before
    update
    on
        public.lessons for each row execute function not_update_lesson_lock();
