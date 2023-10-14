CREATE UNIQUE INDEX granted_role_granted_role_id_key ON public.granted_role USING btree (granted_role_id);
CREATE INDEX granted_role_user_group_id_idx ON public.granted_role USING btree (user_group_id);
CREATE INDEX lesson_classrooms__classroom_id__idx ON public.lesson_classrooms USING btree (classroom_id);
CREATE INDEX lesson_members__user_id__idx ON public.lesson_members USING btree (user_id);
create OR REPLACE FUNCTION not_update_related_lesson_lock() RETURNS trigger
   LANGUAGE plpgsql AS
$$BEGIN
   IF EXISTS (SELECT 1 FROM lessons l WHERE l.lesson_id = OLD.lesson_id AND l.is_locked = true) THEN
      RETURN OLD;
END IF;
RETURN NEW;
END;$$;
create trigger not_update_lesson_member_when_lesson_lock before
    update
    on
        public.lesson_members for each row execute function not_update_related_lesson_lock();

create trigger not_update_lesson_teacher_when_lesson_lock before
    update
    on
        public.lessons_teachers for each row execute function not_update_related_lesson_lock();
