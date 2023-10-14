--Drop table

-- DROP TABLE public.books;

CREATE TABLE IF NOT EXISTS public.books (
	book_id text NOT NULL,
	"name" text NOT NULL,
	country text NULL,
	subject text NULL,
	grade int2 NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	school_id int4 NOT NULL DEFAULT '-2147483648'::integer,
	copied_from text NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	current_chapter_display_order int4 NOT NULL DEFAULT 0,
	CONSTRAINT books_pk PRIMARY KEY (book_id)
);

-- Drop table

-- DROP TABLE public.books_chapters;

CREATE TABLE IF NOT EXISTS public.books_chapters (
	book_id text NOT NULL,
	chapter_id text NOT NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	CONSTRAINT books_chapters_pk PRIMARY KEY (book_id, chapter_id)
);

-- Drop table

-- DROP TABLE public.chapters;

CREATE TABLE IF NOT EXISTS public.chapters (
	chapter_id text NOT NULL,
	"name" text NOT NULL,
	country text NULL,
	subject text NULL,
	grade int2 NULL,
	display_order int2 NULL DEFAULT 0,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	school_id int4 NOT NULL DEFAULT '-2147483648'::integer,
	deleted_at timestamptz NULL,
	copied_from text NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	current_topic_display_order int4 NULL DEFAULT 0,
	CONSTRAINT chapters_pk PRIMARY KEY (chapter_id)
);

-- Drop table

-- DROP TABLE public.courses_books;

CREATE TABLE IF NOT EXISTS public.courses_books (
	book_id text NOT NULL,
	course_id text NOT NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	CONSTRAINT courses_books_pk PRIMARY KEY (book_id, course_id)
);

-- Drop table

-- DROP TABLE public.flashcard_progressions;

CREATE TABLE IF NOT EXISTS public.flashcard_progressions (
	study_set_id text NOT NULL,
	original_study_set_id text NULL,
	student_id text NOT NULL,
	study_plan_item_id text NOT NULL,
	lo_id text NOT NULL,
	quiz_external_ids _text NULL,
	studying_index int4 NULL,
	skipped_question_ids _text NULL,
	remembered_question_ids _text NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	completed_at timestamptz NULL,
	deleted_at timestamptz NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	original_quiz_set_id text NULL,
	CONSTRAINT flashcard_progressions_pk PRIMARY KEY (study_set_id)
);

-- Drop table

-- DROP TABLE public.flashcard_speeches;

CREATE TABLE IF NOT EXISTS public.flashcard_speeches (
	speech_id text NOT NULL,
	sentence text NOT NULL,
	link text NOT NULL,
	"type" text NOT NULL,
	quiz_id text NOT NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	created_by text NULL,
	updated_by text NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	settings jsonb NULL,
	CONSTRAINT flashcard_speeches_pk PRIMARY KEY (speech_id)
);

-- Drop table

-- DROP TABLE public.learning_objectives;

CREATE TABLE IF NOT EXISTS public.learning_objectives (
	lo_id text NOT NULL,
	"name" text NOT NULL,
	country text NULL,
	grade int2 NULL,
	subject text NULL,
	topic_id text NULL,
	master_lo_id text NULL,
	display_order int2 NULL,
	prerequisites _text NULL,
	video text NULL,
	study_guide text NULL,
	video_script text NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	school_id int4 NOT NULL DEFAULT '-2147483648'::integer,
	deleted_at timestamptz NULL,
	copied_from text NULL,
	"type" text NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	CONSTRAINT learning_objectives_pk PRIMARY KEY (lo_id)
);
CREATE INDEX IF NOT EXISTS learning_objectives_topic_id_idx ON public.learning_objectives USING btree (topic_id);


-- Drop table

-- DROP TABLE public.quiz_sets;

CREATE TABLE IF NOT EXISTS public.quiz_sets (
	quiz_set_id text NOT NULL,
	lo_id text NOT NULL,
	quiz_external_ids _text NOT NULL,
	status text NOT NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	CONSTRAINT quiz_sets_pk PRIMARY KEY (quiz_set_id)
);
CREATE UNIQUE INDEX IF NOT EXISTS quiz_sets_approved_lo_id_idx ON public.quiz_sets USING btree (lo_id) WHERE ((status = 'QUIZSET_STATUS_APPROVED'::text) AND (deleted_at IS NULL));
CREATE INDEX IF NOT EXISTS quiz_sets_lo_id_idx ON public.quiz_sets USING btree (lo_id);

-- Drop table

-- DROP TABLE public.quizzes;

CREATE TABLE IF NOT EXISTS public.quizzes (
	quiz_id text NOT NULL,
	country text NOT NULL,
	school_id int4 NOT NULL,
	external_id text NOT NULL,
	kind text NOT NULL,
	question jsonb NOT NULL,
	explanation jsonb NOT NULL,
	"options" jsonb NOT NULL,
	tagged_los _text NULL,
	difficulty_level int4 NULL,
	created_by text NOT NULL,
	approved_by text NOT NULL,
	status text NOT NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	lo_ids _text NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	CONSTRAINT quizs_pk PRIMARY KEY (quiz_id)
);
CREATE INDEX IF NOT EXISTS quizzes_external_id_idx ON public.quizzes USING btree (external_id);

-- Drop table

-- DROP TABLE public.shuffled_quiz_sets;

CREATE TABLE IF NOT EXISTS public.shuffled_quiz_sets (
	shuffled_quiz_set_id text NOT NULL,
	original_quiz_set_id text NULL,
	quiz_external_ids _text NULL,
	status text NULL,
	random_seed text NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	student_id text NOT NULL,
	study_plan_item_id text NULL,
	total_correctness int4 NOT NULL DEFAULT 0,
	submission_history jsonb NOT NULL DEFAULT '[]'::jsonb,
	session_id text NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	original_shuffle_quiz_set_id text NULL,
	CONSTRAINT shuffled_quiz_sets_pkey PRIMARY KEY (shuffled_quiz_set_id)
);
CREATE INDEX IF NOT EXISTS shuffled_quiz_original_quiz_set_id_idx ON public.shuffled_quiz_sets USING btree (original_quiz_set_id);
CREATE INDEX IF NOT EXISTS shuffled_quiz_sets_study_plan_item_idx ON public.shuffled_quiz_sets USING btree (study_plan_item_id);

-- Drop table

-- DROP TABLE public.student_event_logs;

CREATE TABLE IF NOT EXISTS public.student_event_logs (
	student_event_log_id serial4 NOT NULL,
	student_id text NOT NULL,
	created_at timestamptz NOT NULL,
	event_type varchar(100) NOT NULL,
	payload jsonb NULL,
	event_id varchar(50) NULL,
	deleted_at timestamptz NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	CONSTRAINT event_id_un UNIQUE (event_id),
	CONSTRAINT event_log_pk PRIMARY KEY (student_event_log_id)
);
CREATE INDEX IF NOT EXISTS event_logs_student_id_idx ON public.student_event_logs USING btree (student_id);
CREATE INDEX IF NOT EXISTS student_event_logs_event_type_idx ON public.student_event_logs USING btree (event_type);
CREATE INDEX IF NOT EXISTS student_event_logs_payload_session_id_idx ON public.student_event_logs USING btree (((payload ->> 'session_id'::text)));
CREATE INDEX IF NOT EXISTS student_event_logs_payload_study_plan_item_id_idx ON public.student_event_logs USING btree (((payload ->> 'study_plan_item_id'::text))) WHERE ((payload ->> 'study_plan_item_id'::text) IS NOT NULL);


-- Drop table

-- DROP TABLE public.student_learning_time_by_daily;

CREATE TABLE IF NOT EXISTS public.student_learning_time_by_daily (
	learning_time_id serial4 NOT NULL,
	student_id text NOT NULL,
	learning_time int4 NOT NULL DEFAULT 0,
	"day" timestamptz NOT NULL,
	sessions text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	CONSTRAINT student_learning_time_by_daily_pk PRIMARY KEY (learning_time_id)
);

-- Drop table

-- DROP TABLE public.students_learning_objectives_completeness;

CREATE TABLE IF NOT EXISTS public.students_learning_objectives_completeness (
	student_id text NOT NULL,
	lo_id text NOT NULL,
	preset_study_plan_weekly_id text NULL,
	first_attempt_score int2 NOT NULL DEFAULT 0,
	is_finished_quiz bool NOT NULL DEFAULT false,
	is_finished_video bool NOT NULL DEFAULT false,
	is_finished_study_guide bool NOT NULL DEFAULT false,
	first_quiz_correctness float4 NULL,
	finished_quiz_at timestamptz NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	highest_quiz_score float4 NULL,
	deleted_at timestamptz NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	CONSTRAINT students_learning_objectives_completeness_pk PRIMARY KEY (student_id, lo_id)
);

-- Drop table

-- DROP TABLE public.students_learning_objectives_records;

CREATE TABLE IF NOT EXISTS public.students_learning_objectives_records (
	record_id text NOT NULL,
	lo_id text NOT NULL,
	study_plan_item_id text NOT NULL,
	student_id text NOT NULL,
	accuracy numeric(5, 2) NULL,
	learning_time int4 NULL,
	completed_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	is_offline bool NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	CONSTRAINT students_learning_objectives_records_pk PRIMARY KEY (record_id)
);

-- Drop table

-- DROP TABLE public.students_topics_completeness;

CREATE TABLE IF NOT EXISTS public.students_topics_completeness (
	student_id text NOT NULL,
	topic_id text NOT NULL,
	total_finished_los int4 NOT NULL DEFAULT 0,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	is_completed bool NULL DEFAULT false,
	deleted_at timestamptz NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	CONSTRAINT students_topics_completeness_pk UNIQUE (student_id, topic_id)
);

-- Drop table

-- DROP TABLE public.topics;

CREATE TABLE IF NOT EXISTS public.topics (
	topic_id text NOT NULL,
	"name" text NOT NULL,
	country text NULL,
	grade int2 NOT NULL,
	subject text NOT NULL,
	topic_type text NOT NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	status text NULL DEFAULT 'TOPIC_STATUS_PUBLISHED'::text,
	display_order int2 NULL,
	published_at timestamptz NULL,
	total_los int4 NOT NULL DEFAULT 0,
	chapter_id text NULL,
	icon_url text NULL,
	school_id int4 NOT NULL DEFAULT '-2147483648'::integer,
	attachment_urls _text NULL,
	instruction text NULL,
	copied_topic_id text NULL,
	essay_required bool NOT NULL DEFAULT false,
	attachment_names _text NULL,
	deleted_at timestamptz NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	lo_display_order_counter int4 NULL DEFAULT 0,
	CONSTRAINT topics_pk PRIMARY KEY (topic_id)
);

-- Drop table

-- DROP TABLE public.topics_learning_objectives;

CREATE TABLE IF NOT EXISTS public.topics_learning_objectives (
	topic_id text NOT NULL,
	lo_id text NOT NULL,
	display_order int2 NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	CONSTRAINT topics_learning_objectives_pk PRIMARY KEY (topic_id, lo_id)
);




