ALTER TABLE IF EXISTS public.question_group
    ALTER COLUMN created_at SET DEFAULT (now() at time zone 'utc'),
    ALTER COLUMN updated_at SET DEFAULT (now() at time zone 'utc'),
    DROP CONSTRAINT IF EXISTS question_group_pk,
    ADD CONSTRAINT question_group_pk PRIMARY KEY (question_group_id);
