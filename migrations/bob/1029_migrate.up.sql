CREATE TABLE IF NOT EXISTS public.quizzes (
	quiz_id text NOT NULL,
	country text NOT NULL,
	school_id int4 NOT NULL,
	external_id text NOT NULL,
	kind text NOT NULL,
  question JSONB NOT NULL,
  explanation JSONB NOT NULL,
  options JSONB NOT NULL,
  tagged_los text[] NOT NULL,
	difficulty_level int4 NULL,
	created_by TEXT NOT NULL,
	approved_by TEXT NOT NULL,
	status TEXT NOT NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
  CONSTRAINT quizs_pk PRIMARY KEY (quiz_id)
);

CREATE INDEX IF NOT EXISTS quizzes_external_id_idx ON public.quizzes (external_id);

CREATE TABLE IF NOT EXISTS quiz_sets (
	quiz_set_id text NOT NULL,
	lo_id text NOT NULL,
	quiz_external_ids _text NOT NULL,
	status TEXT NOT NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	CONSTRAINT quiz_sets_pk PRIMARY KEY (quiz_set_id),
	CONSTRAINT quiz_sets_fk FOREIGN KEY (lo_id) REFERENCES public.learning_objectives(lo_id)
);

CREATE INDEX IF NOT EXISTS quiz_sets_lo_id_idx ON public.quiz_sets (lo_id);
