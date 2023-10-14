--
-- PostgreSQL database dump
--

-- Dumped from database version 11.9

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA public;


ALTER SCHEMA public OWNER TO postgres;

--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: postgres
--

COMMENT ON SCHEMA public IS 'standard public schema';


--
-- Name: japanese_collation; Type: COLLATION; Schema: public; Owner: postgres
--

CREATE COLLATION public.japanese_collation (provider = icu, locale = 'en-u-kn-true-kr-digit-en-ja_JP');


ALTER COLLATION public.japanese_collation OWNER TO postgres;

--
-- Name: rating_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.rating_type AS ENUM (
    'neutral',
    'positive',
    'negative'
);


ALTER TYPE public.rating_type OWNER TO postgres;

--
-- Name: assignment_graded_score_v2(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.assignment_graded_score_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, student_submission_id text, graded_point smallint, total_point smallint, status text, passed boolean, created_at timestamp with time zone)
    LANGUAGE sql STABLE
    AS $$
select 
    ss.student_id,
    ss.study_plan_id,
    ss.learning_material_id,
    ss.student_submission_id,
    ssg.grade::smallint as graded_point,
    a.max_grade::smallint as total_point,
    ss.status,
    ss.understanding_level != 'SUBMISSION_UNDERSTANDING_LEVEL_SAD' as passed,
    ss.created_at 
from student_submissions ss
join student_submission_grades ssg on ss.student_submission_id = ssg.student_submission_id
join assignment a using (learning_material_id)
where ssg.grade != -1 and ss.status = 'SUBMISSION_STATUS_RETURNED'
group by ss.student_id,
         ss.study_plan_id,
         ss.learning_material_id,
         ss.student_submission_id,
         ssg.grade::smallint,
         a.max_grade::smallint,
         ss.status,
         ss.understanding_level,
         ss.created_at;
$$;


ALTER FUNCTION public.assignment_graded_score_v2() OWNER TO postgres;

--
-- Name: autofillresourcepath(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.autofillresourcepath() RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE
		resource_path text;
BEGIN
	resource_path := current_setting('permission.resource_path', 't');

	RETURN resource_path;
END $$;


ALTER FUNCTION public.autofillresourcepath() OWNER TO postgres;

--
-- Name: book_tree_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.book_tree_fn() RETURNS TABLE(book_id text, chapter_id text, chapter_display_order smallint, topic_id text, topic_display_order smallint, learning_material_id text, lm_display_order smallint)
    LANGUAGE sql STABLE
    AS $$
SELECT
  b.book_id,
  c.chapter_id,
  c.display_order AS chapter_display_order,
  t.topic_id,
  t.display_order AS topic_display_order,
  lm.learning_material_id,
  lm.display_order AS lm_display_order
FROM
  books b
  JOIN chapters c USING (book_id)
  JOIN topics t USING (chapter_id)
  JOIN learning_material lm USING (topic_id)
WHERE
  COALESCE(
    b.deleted_at,
    c.deleted_at,
    t.deleted_at,
    lm.deleted_at
  ) IS NULL;
$$;


ALTER FUNCTION public.book_tree_fn() OWNER TO postgres;

SET default_tablespace = '';

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    user_id text NOT NULL,
    name text NOT NULL,
    user_group text NOT NULL,
    resource_path text DEFAULT public.autofillresourcepath(),
    country text NOT NULL,
    avatar text,
    phone_number text,
    email text,
    device_token text,
    allow_notification boolean,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    is_tester boolean,
    facebook_id text,
    platform text,
    phone_verified boolean,
    email_verified boolean,
    deleted_at timestamp with time zone,
    given_name text,
    last_login_date timestamp with time zone,
    birthday date,
    gender text,
    first_name text DEFAULT ''::text NOT NULL,
    last_name text DEFAULT ''::text NOT NULL,
    first_name_phonetic text,
    last_name_phonetic text,
    full_name_phonetic text
);

ALTER TABLE ONLY public.users FORCE ROW LEVEL SECURITY;


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: bypass_rls_search_name_user_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.bypass_rls_search_name_user_fn() RETURNS SETOF public.users
    LANGUAGE sql STABLE SECURITY DEFINER
    AS $$
    SELECT
        *
    FROM
        public.users
$$;


ALTER FUNCTION public.bypass_rls_search_name_user_fn() OWNER TO postgres;

--
-- Name: calculate_learning_time(text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.calculate_learning_time(_student_id text) RETURNS TABLE(learning_time_by_minutes integer, student_id text, sessions text, date date, day timestamp with time zone, assignment_duration integer, submit_learning_material_id text)
    LANGUAGE sql STABLE
    AS $$
with data as (select created_at                              as date
                   , to_date(created_at::text, 'YYYY-MM-DD') as date_normalize
                   , payload ->> 'session_id'                as session_id
                   , payload ->> 'event'                     as event
                   , student_id
              from student_event_logs
              where 1 = 1
                and payload ->> 'session_id' is not null
                and payload ->> 'event' is not null
                and event_type = 'learning_objective'
                and student_id = _student_id
              order by created_at)
   , pause_resume_normalize as (select student_id,
                                       session_id,
                                       event,
                                       event_at,
                                       next_event,
                                       next_event_at,
                                       date_normalize
                                from (select student_id,
                                             session_id,
                                             date_normalize
                                              ,
                                             lead(date_normalize) over (order by session_id, date) as next_date_normalize
                                              ,
                                             lag(date_normalize) over (order by session_id, date)  as prev_date_normalize
                                              ,
                                             lead(session_id) over (order by session_id, date)     as next_session_id,
                                             lag(session_id) over (order by session_id, date)      as prev_session_id,
                                             event,
                                             date                                                  as event_at,
                                             lag(event) over (order by session_id, date)           as prev_event,
                                             lead(event) over (order by session_id, date)          as next_event,
                                             lead(date) over (order by session_id, date)           as next_event_at
                                      from (select session_id
                                                 , date
                                                 , date_normalize
                                                 , event
                                                 , lead(date_normalize) over (order by session_id, date) as next_date_normalize
                                                 , lag(date_normalize) over (order by session_id, date)  as prev_date_normalize
                                                 , lead(session_id) over (order by session_id, date)     as next_session_id
                                                 , lag(session_id) over (order by session_id, date)      as prev_session_id
                                                 , lead(event) over (order by session_id, date)          as next_event
                                                 , lag(event) over (order by session_id, date)           as prev_event
                                                 , student_id
                                            from data
                                            where event = 'paused'
                                               or event = 'resumed'
                                               or event = 'started'
                                               or event = 'exited') pause_resume
                                      where ((event = 'paused' and
                                              (prev_event is null or
                                               (not (prev_event = 'paused' or prev_event = 'exited') and
                                                session_id = prev_session_id) or
                                               (prev_event = 'paused' and session_id != prev_session_id)) and
                                              (next_event is not null and
                                               (not next_event = 'started' or prev_event = 'started')) and
                                              ((date_normalize = next_date_normalize or
                                                date_normalize = prev_date_normalize) and
                                               (session_id = next_session_id or session_id = prev_session_id))) or
                                             (event = 'resumed' and next_event is null and
                                              (session_id = next_session_id or session_id = prev_session_id)) or
                                             (event = 'resumed' and (not next_event = 'resumed') and
                                              (prev_event is not null and not prev_event = 'resumed') and
                                              (session_id = next_session_id or session_id = prev_session_id)))
                                      order by student_id, session_id, date) pause_resume_filter
                                where event = 'paused'
                                  and (next_event = 'resumed' and next_session_id = session_id)
                                  and (session_id = next_session_id or session_id = prev_session_id))
   , start_complete_normalize as (select student_id,
                                         session_id,
                                         event,
                                         event_at,
                                         next_event,
                                         next_event_at,
                                         date_normalize
                                  from (select session_id,
                                               date_normalize
                                                ,
                                               lead(date_normalize) over (order by session_id, date) as next_date_normalize
                                                ,
                                               lag(date_normalize) over (order by session_id, date)  as prev_date_normalize
                                                ,
                                               lead(session_id) over (order by session_id, date)     as next_session_id,
                                               lag(session_id) over (order by session_id, date)      as prev_session_id,
                                               event,
                                               date                                                  as event_at,
                                               lag(event) over (order by session_id, date)           as prev_event,
                                               lead(event) over (order by session_id, date)          as next_event,
                                               lead(date) over (order by session_id, date)           as next_event_at,
                                               student_id
                                        from (select session_id
                                                   , date
                                                   , date_normalize
                                                   , event
                                                   , lead(date_normalize) over (order by session_id, date) as next_date_normalize
                                                   , lag(date_normalize) over (order by session_id, date)  as prev_date_normalize
                                                   , lead(session_id) over (order by session_id, date)     as next_session_id
                                                   , lag(session_id) over (order by session_id, date)      as prev_session_id
                                                   , lead(event) over (order by session_id, date)          as next_event
                                                   , lag(event) over (order by session_id, date)           as prev_event
                                                   , student_id
                                              from data
                                              where event = 'started'
                                                 or event = 'completed') start_complete
                                        where (event = 'started' and (prev_event is null or prev_event != 'started' or
                                                                      (prev_event = 'started' and session_id != prev_session_id))
                                            and ((date_normalize = next_date_normalize or
                                                  date_normalize = prev_date_normalize) and
                                                 (session_id = next_session_id or session_id = prev_session_id)))
                                           or (event = 'completed' and next_event is null and
                                               ((date_normalize = next_date_normalize or
                                                 date_normalize = prev_date_normalize) and
                                                (session_id = next_session_id or session_id = prev_session_id)))
                                           or (event = 'completed' and not next_event = 'completed' and
                                               prev_event is not null and ((date_normalize = next_date_normalize or
                                                                            date_normalize = prev_date_normalize) and
                                                                           (session_id = next_session_id or session_id = prev_session_id)))) start_complete_filter
                                  where event = 'started'
                                    and (next_event = 'completed' and next_session_id = session_id)
                                    and (session_id = next_session_id or session_id = prev_session_id))
   , total_sc_by_date as (select sum(time_by_sec)      as time_by_sec,
                                 student_id,
                                 date_normalize,
                                 array_agg(session_id) as sessions
                          from (select (coalesce(EXTRACT(epoch FROM (sum(next_event_at - event_at))), 0)) as time_by_sec,
                                       student_id,
                                       session_id,
                                       date_normalize
                                from start_complete_normalize
                                group by student_id, session_id, date_normalize) total_sc_by_session
                          group by student_id, date_normalize)
   , total_pr_by_date as (select sum(time_by_sec)      as time_by_sec,
                                 student_id,
                                 date_normalize,
                                 array_agg(session_id) as sessions
                          from (select (coalesce(EXTRACT(epoch FROM (sum(next_event_at - event_at))), 0)) as time_by_sec,
                                       student_id,
                                       session_id,
                                       date_normalize
                                from pause_resume_normalize
                                group by student_id, session_id, date_normalize) total_pr_by_session
                          group by student_id, date_normalize)
   , learning_time as (select distinct ((a.time_by_sec - coalesce(b.time_by_sec, 0)) / 60)::int as minute,
                                       a.student_id,
                                       (select array_agg(sessions)::text
                                        from (select distinct unnest(aa.sessions || b.sessions) as sessions
                                              from total_sc_by_date aa
                                                       left join total_pr_by_date b using (student_id, date_normalize)
                                              where a.student_id = aa.student_id
                                                and a.date_normalize = aa.date_normalize) subq) as sessions,
                                       a.date_normalize                                         as date,
                                       a.date_normalize::timestamp AT TIME ZONE 'UTC'           as day
                       from total_sc_by_date a
                                left join total_pr_by_date b using (student_id, date_normalize)
                       group by a.student_id, a.date_normalize, a.time_by_sec, b.time_by_sec)
select (learning_time.minute + coalesce(sum(student_submissions.duration),0) / 60)::int,
       learning_time.student_id,
       learning_time.sessions,
       learning_time.date,
       learning_time.day,
       (sum(student_submissions.duration) / 60)::int as assignment_duration,
       (array_agg(learning_material_id))::text             as learning_material_ids
from learning_time
         left join student_submissions on learning_time.student_id = student_submissions.student_id and learning_time.date = to_date(student_submissions.created_at::text, 'YYYY-MM-DD')
group by learning_time.minute, learning_time.student_id, learning_time.sessions, learning_time.date, learning_time.day
$$;


ALTER FUNCTION public.calculate_learning_time(_student_id text) OWNER TO postgres;

--
-- Name: check_study_plan_item_time(timestamp with time zone, timestamp with time zone, timestamp with time zone, timestamp with time zone); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.check_study_plan_item_time(master_updated_at timestamp with time zone, student_updated_at timestamp with time zone, master_time timestamp with time zone, student_time timestamp with time zone) RETURNS timestamp with time zone
    LANGUAGE plpgsql
    AS $$
begin
    case 
      when master_updated_at is null or student_updated_at is null
          then return coalesce(master_time, student_time);
      else 
        case
              when master_updated_at >= student_updated_at 
                then return master_time;
              else return student_time;
        end case;
    end case;
end;
$$;


ALTER FUNCTION public.check_study_plan_item_time(master_updated_at timestamp with time zone, student_updated_at timestamp with time zone, master_time timestamp with time zone, student_time timestamp with time zone) OWNER TO postgres;

--
-- Name: create_assignment_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.create_assignment_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ 
BEGIN IF new.type = 'ASSIGNMENT_TYPE_LEARNING_OBJECTIVE' THEN
    INSERT INTO
        assignment (
            learning_material_id,
            topic_id,
            name,
            type,
            display_order,
            attachments,
            max_grade,
            instruction,
            is_required_grade,
            allow_resubmission,
            require_attachment,
            allow_late_submission,
            require_assignment_note,
            require_video_submission,
            created_at,
            updated_at,
            deleted_at,
            resource_path
        )
    VALUES
        (
            new.assignment_id,
            new.content ->> 'topic_id',
            new.name,
            'LEARNING_MATERIAL_GENERAL_ASSIGNMENT',
            new.display_order,
            new.attachment,
            new.max_grade,
            new.instruction,
            new.is_required_grade,
            (COALESCE(new.settings ->> 'allow_resubmission', 'false'))::boolean,
        	(COALESCE(new.settings ->> 'require_attachment', 'false'))::boolean,
        	(COALESCE(new.settings ->> 'allow_late_submission', 'false'))::boolean,
        	(COALESCE(new.settings ->> 'require_assignment_note', 'false'))::boolean,
        	(COALESCE(new.settings ->> 'require_video_submission', 'false'))::boolean,
            new.created_at,
            new.updated_at,
            new.deleted_at,
            new.resource_path
        ) ON CONFLICT 
    ON CONSTRAINT assignment_pk
    DO UPDATE
    SET
        name = new.name,
        display_order = new.display_order,
        attachments = new.attachment,
        max_grade = new.max_grade,
        instruction = new.instruction,
        is_required_grade = new.is_required_grade,
        allow_resubmission = (COALESCE(new.settings ->> 'allow_resubmission', 'false'))::boolean,
        require_attachment = (COALESCE(new.settings ->> 'require_attachment', 'false'))::boolean,
        allow_late_submission = (COALESCE(new.settings ->> 'allow_late_submission', 'false'))::boolean,
        require_assignment_note = (COALESCE(new.settings ->> 'require_assignment_note', 'false'))::boolean,
        require_video_submission = (COALESCE(new.settings ->> 'require_video_submission', 'false'))::boolean,
        updated_at = new.updated_at,
        deleted_at = new.deleted_at;
    END IF;
RETURN NULL;
END;
$$;


ALTER FUNCTION public.create_assignment_fn() OWNER TO postgres;

--
-- Name: create_flash_card_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.create_flash_card_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ 
BEGIN IF new.type = 'LEARNING_OBJECTIVE_TYPE_FLASH_CARD' THEN
    INSERT INTO
        flash_card (
            learning_material_id,
            topic_id,
            name,
            type,
            display_order,
            created_at,
            updated_at,
            resource_path,
            deleted_at
        )
    VALUES
        (
            new.lo_id,
            new.topic_id,
            new.name,
            'LEARNING_MATERIAL_FLASH_CARD',
            new.display_order,
            new.created_at,
            new.updated_at,
            new.resource_path,
            new.deleted_at
        ) ON CONFLICT 
    ON CONSTRAINT flash_card_pk
    DO UPDATE
    SET
        updated_at = new.updated_at,
        name = new.name,
        display_order = new.display_order,
        deleted_at = new.deleted_at;
        END IF;
RETURN NULL;
END;
$$;


ALTER FUNCTION public.create_flash_card_fn() OWNER TO postgres;

--
-- Name: create_learning_objective_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.create_learning_objective_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ 
BEGIN IF new.type = 'LEARNING_OBJECTIVE_TYPE_LEARNING' THEN
    INSERT INTO
        learning_objective (
            learning_material_id,
            topic_id,
            name,
            type,
            display_order,
            created_at,
            updated_at,
            resource_path,
            video,
            study_guide,
            video_script,
            vendor_type,
            vendor_reference_id,
            deleted_at
        )
    VALUES
        (
            new.lo_id,
            new.topic_id,
            new.name,
            'LEARNING_MATERIAL_LEARNING_OBJECTIVE',
            new.display_order,
            new.created_at,
            new.updated_at,
            new.resource_path,
            new.video,
            new.study_guide,
            new.video_script,
            new.vendor_type,
            new.vendor_reference_id,
            new.deleted_at
        ) ON CONFLICT 
    ON CONSTRAINT learning_objective_pk
    DO UPDATE
    SET
        updated_at = new.updated_at,
        name = new.name,
        display_order = new.display_order,
        video = new.video,
        study_guide = new.study_guide,
        video_script = new.video_script,
        vendor_type = new.vendor_type,
        vendor_reference_id = new.vendor_reference_id,
        deleted_at = new.deleted_at;
        END IF;
RETURN NULL;
END;
$$;


ALTER FUNCTION public.create_learning_objective_fn() OWNER TO postgres;

--
-- Name: create_task_assignment_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.create_task_assignment_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ 
BEGIN IF new.type = 'ASSIGNMENT_TYPE_TASK' THEN
    INSERT INTO
        task_assignment (
            learning_material_id,
            topic_id,
            name,
            type,
            display_order,
            attachments,
            instruction,
            require_duration,
            require_complete_date,
            require_understanding_level,
            require_correctness,
            require_attachment,
            require_assignment_note,
            created_at,
            updated_at,
            deleted_at,
            resource_path
        )
    VALUES
        (
            new.assignment_id,
            new.content ->> 'topic_id',
            new.name,
            'LEARNING_MATERIAL_TASK_ASSIGNMENT',
            new.display_order,
            new.attachment,
            new.instruction,
            (COALESCE(new.settings ->> 'require_duration', 'false'))::boolean,
            (COALESCE(new.settings ->> 'require_complete_date', 'false'))::boolean,
            (COALESCE(new.settings ->> 'require_understanding_level', 'false'))::boolean,
            (COALESCE(new.settings ->> 'require_correctness', 'false'))::boolean,
            (COALESCE(new.settings ->> 'require_attachment', 'false'))::boolean,
            (COALESCE(new.settings ->> 'require_assignment_note', 'false'))::boolean,
            new.created_at,
            new.updated_at,
            new.deleted_at,
            new.resource_path
        ) ON CONFLICT 
    ON CONSTRAINT task_assignment_pk
    DO UPDATE
    SET
        name = new.name,
        display_order = new.display_order,
        attachments = new.attachment,
        instruction = new.instruction,
        require_duration = (COALESCE(new.settings ->> 'require_duration', 'false'))::boolean,
        require_complete_date = (COALESCE(new.settings ->> 'require_complete_date', 'false'))::boolean,
        require_understanding_level = (COALESCE(new.settings ->> 'require_understanding_level', 'false'))::boolean,
        require_correctness = (COALESCE(new.settings ->> 'require_correctness', 'false'))::boolean,
        require_attachment = (COALESCE(new.settings ->> 'require_attachment', 'false'))::boolean,
        require_assignment_note = (COALESCE(new.settings ->> 'require_assignment_note', 'false'))::boolean,
        updated_at = new.updated_at,
        deleted_at = new.deleted_at;
    END IF;
RETURN NULL;
END;
$$;


ALTER FUNCTION public.create_task_assignment_fn() OWNER TO postgres;

--
-- Name: exam_lo_graded_score_v2(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.exam_lo_graded_score_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, graded_point smallint, total_point smallint, status text, result text, created_at timestamp with time zone)
    LANGUAGE sql STABLE
    AS $$
 select els.student_id,
    els.study_plan_id,
    els.learning_material_id,
    els.submission_id,
    sum(coalesce(elss.point, elsa.point))::smallint as graded_point,
    els.total_point::smallint as total_point,
    els.status,
    els.result,
    els.created_at
from exam_lo_submission els
    join exam_lo_submission_answer elsa using (submission_id)
    left join exam_lo_submission_score elss using (submission_id, quiz_id)
    where els.status = 'SUBMISSION_STATUS_RETURNED' AND els.deleted_at IS NULL
group by els.submission_id;
$$;


ALTER FUNCTION public.exam_lo_graded_score_v2() OWNER TO postgres;

--
-- Name: fc_answer(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fc_answer() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, external_quiz_id text, is_accepted boolean, point integer)
    LANGUAGE sql STABLE
    AS $$
select student_id,
       study_plan_id,
       learning_material_id,
       submission_id,
       external_quiz_id,
       bool_or(is_accepted) as is_accepted,
       q.point
from (
    select student_id,
           study_plan_id,
           learning_material_id,
           submission_id,
           (jsonb_array_elements(submission_history) ->> 'quiz_id')::text     as external_quiz_id,
           (jsonb_array_elements(submission_history) ->> 'is_accepted')::bool as is_accepted
    from fc_raw_answer()
) raw_answer 
join quizzes q
    on raw_answer.external_quiz_id = q.external_id
group by student_id,
         study_plan_id,
         learning_material_id,
         submission_id,
         external_quiz_id,
         q.point
$$;


ALTER FUNCTION public.fc_answer() OWNER TO postgres;

--
-- Name: fc_answer_v2(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fc_answer_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, external_quiz_id text, is_accepted boolean, point integer, total_point integer)
    LANGUAGE sql STABLE
    AS $$
select  sa.student_id,
        sa.study_plan_id,
        sa.learning_material_id,
        sa.submission_id,
        sa.quiz_id,
        sa.is_accepted,
        point,
        s.total_point
from flash_card_submission_answer sa
join flash_card_submission s using(submission_id)
join get_student_completion_learning_material() clm on
    clm.student_id = sa.student_id and clm.study_plan_id = sa.study_plan_id and clm.learning_material_id = sa.learning_material_id
$$;


ALTER FUNCTION public.fc_answer_v2() OWNER TO postgres;

--
-- Name: fc_graded_score(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fc_graded_score() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, graded_points smallint, total_points smallint, status text)
    LANGUAGE sql STABLE
    AS $$
select student_id,
       study_plan_id,
       learning_material_id,
       submission_id,
       sum(is_accepted::int * point)::smallint as graded_score,
       sum(point)::smallint                    as total_scores,
       'S'
from fc_answer()
group by student_id,
         study_plan_id,
         learning_material_id,
         submission_id
$$;


ALTER FUNCTION public.fc_graded_score() OWNER TO postgres;

--
-- Name: fc_graded_score_v2(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fc_graded_score_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, graded_point smallint, total_point smallint, status text)
    LANGUAGE sql STABLE
    AS $$
 select sa.student_id,
        sa.study_plan_id,
        sa.learning_material_id,
        sa.submission_id,
        sum(point)::smallint as graded_point,
        max(s.total_point)::smallint as total_point,
        'S'
from flash_card_submission_answer sa
join flash_card_submission s using (submission_id)
where s.is_submitted is true and s.deleted_at is null
group by sa.student_id,
         sa.study_plan_id,
         sa.learning_material_id,
         sa.submission_id
$$;


ALTER FUNCTION public.fc_graded_score_v2() OWNER TO postgres;

--
-- Name: fc_raw_answer(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fc_raw_answer() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, submission_history jsonb, quiz_external_ids text[])
    LANGUAGE sql STABLE
    AS $$

select student_id,
       study_plan_id,
       learning_material_id,
       shuffled_quiz_set_id as submission_id,
       submission_history,
       quiz_external_ids
from shuffled_quiz_sets
    join flash_card using (learning_material_id)
$$;


ALTER FUNCTION public.fc_raw_answer() OWNER TO postgres;

--
-- Name: learning_material; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.learning_material (
    learning_material_id text NOT NULL,
    topic_id text NOT NULL,
    name text NOT NULL,
    type text,
    display_order smallint,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    vendor_type text DEFAULT 'LM_VENDOR_TYPE_MANABIE'::text NOT NULL,
    vendor_reference_id text,
    is_published boolean DEFAULT false NOT NULL,
    CONSTRAINT vendor_type_check CHECK ((vendor_type = ANY (ARRAY['LM_VENDOR_TYPE_MANABIE'::text, 'LM_VENDOR_TYPE_LEARNOSITY'::text])))
);

ALTER TABLE ONLY public.learning_material FORCE ROW LEVEL SECURITY;


ALTER TABLE public.learning_material OWNER TO postgres;

--
-- Name: exam_lo; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.exam_lo (
    instruction text,
    grade_to_pass integer,
    manual_grading boolean DEFAULT false,
    time_limit integer,
    maximum_attempt integer,
    approve_grading boolean DEFAULT false NOT NULL,
    grade_capping boolean DEFAULT false NOT NULL,
    review_option text DEFAULT 'EXAM_LO_REVIEW_OPTION_IMMEDIATELY'::text NOT NULL,
    CONSTRAINT exam_lo_review_option_check CHECK ((review_option = ANY (ARRAY['EXAM_LO_REVIEW_OPTION_IMMEDIATELY'::text, 'EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE'::text]))),
    CONSTRAINT exam_lo_type_check CHECK ((type = 'LEARNING_MATERIAL_EXAM_LO'::text))
)
INHERITS (public.learning_material);

ALTER TABLE ONLY public.exam_lo FORCE ROW LEVEL SECURITY;


ALTER TABLE public.exam_lo OWNER TO postgres;

--
-- Name: filter_rls_search_name_exam_lo_fn(text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.filter_rls_search_name_exam_lo_fn(search_name text) RETURNS SETOF public.exam_lo
    LANGUAGE sql STABLE
    AS $$
    SELECT
        el.*
    FROM
        private_search_name_exam_lo_fn(search_name) el
    JOIN public.exam_lo USING(learning_material_id)
$$;


ALTER FUNCTION public.filter_rls_search_name_exam_lo_fn(search_name text) OWNER TO postgres;

--
-- Name: filter_rls_search_name_lm_fn(text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.filter_rls_search_name_lm_fn(search_name text) RETURNS SETOF public.learning_material
    LANGUAGE sql STABLE
    AS $$
    SELECT
        sl.*
    FROM
        private_search_name_lm_fn(search_name) AS sl
        JOIN public.learning_material ON sl.learning_material_id = learning_material.learning_material_id 
$$;


ALTER FUNCTION public.filter_rls_search_name_lm_fn(search_name text) OWNER TO postgres;

--
-- Name: filter_rls_search_name_user_fn(text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.filter_rls_search_name_user_fn(search_name text) RETURNS SETOF public.users
    LANGUAGE sql STABLE
    AS $$
    SELECT
        us.*
    FROM
        private_search_name_user_fn(search_name) us
    JOIN users USING(user_id)
$$;


ALTER FUNCTION public.filter_rls_search_name_user_fn(search_name text) OWNER TO postgres;

--
-- Name: filter_rls_user_fn(text, text[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.filter_rls_user_fn(search_name text, student_ids text[]) RETURNS SETOF public.users
    LANGUAGE sql STABLE
    AS $$
    SELECT
        us.*
    FROM
        private_search_name_user_fn(search_name) us
    JOIN users USING(user_id)
    WHERE (student_ids is null or user_id = any(student_ids))
$$;


ALTER FUNCTION public.filter_rls_user_fn(search_name text, student_ids text[]) OWNER TO postgres;

--
-- Name: quizzes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.quizzes (
    quiz_id text NOT NULL,
    country text NOT NULL,
    school_id integer NOT NULL,
    external_id text NOT NULL,
    kind text NOT NULL,
    question jsonb NOT NULL,
    explanation jsonb NOT NULL,
    options jsonb NOT NULL,
    tagged_los text[],
    difficulty_level integer,
    created_by text NOT NULL,
    approved_by text NOT NULL,
    status text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    lo_ids text[],
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    point integer DEFAULT 1,
    question_group_id text,
    question_tag_ids text[],
    label_type text DEFAULT 'QUIZ_LABEL_TYPE_NONE'::text,
    CONSTRAINT label_type_check CHECK ((label_type = ANY (ARRAY['QUIZ_LABEL_TYPE_NONE'::text, 'QUIZ_LABEL_TYPE_WITHOUT_LABEL'::text, 'QUIZ_LABEL_TYPE_CUSTOM'::text, 'QUIZ_LABEL_TYPE_NUMBER'::text, 'QUIZ_LABEL_TYPE_TEXT_LOWERCASE'::text, 'QUIZ_LABEL_TYPE_TEXT_UPPERCASE'::text])))
);

ALTER TABLE ONLY public.quizzes FORCE ROW LEVEL SECURITY;


ALTER TABLE public.quizzes OWNER TO postgres;

--
-- Name: find_a_quiz_in_quiz_set(text, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.find_a_quiz_in_quiz_set(quizid text, loid text) RETURNS SETOF public.quizzes
    LANGUAGE sql STABLE
    AS $$

SELECT quiz.* FROM public.quizzes AS quiz 
    WHERE 
        quiz.quiz_id = quizId
    AND quiz.deleted_at IS NULL
    AND (
        EXISTS (
            SELECT qs.* FROM public.quiz_sets AS qs
                where qs.lo_id = lOId
                AND qs.deleted_at IS NULL
                AND qs.quiz_external_ids && ARRAY[quiz.external_id]
        )
    );

$$;


ALTER FUNCTION public.find_a_quiz_in_quiz_set(quizid text, loid text) OWNER TO postgres;

--
-- Name: assignments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.assignments (
    assignment_id text NOT NULL,
    content jsonb,
    attachment text[],
    settings jsonb,
    check_list jsonb,
    name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    max_grade integer,
    status text,
    instruction text,
    type text,
    is_required_grade boolean,
    display_order integer DEFAULT 0,
    original_topic text,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    topic_id text
);

ALTER TABLE ONLY public.assignments FORCE ROW LEVEL SECURITY;


ALTER TABLE public.assignments OWNER TO postgres;

--
-- Name: find_assignment_by_topic_id(text[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.find_assignment_by_topic_id(ids text[]) RETURNS SETOF public.assignments
    LANGUAGE sql STABLE
    AS $$
 select * from assignments a where a."content"->>'topic_id' = any(ids::text[]);
$$;


ALTER FUNCTION public.find_assignment_by_topic_id(ids text[]) OWNER TO postgres;

--
-- Name: find_question_by_lo_id(character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.find_question_by_lo_id(id character varying) RETURNS SETOF public.quizzes
    LANGUAGE sql STABLE
    AS $$
WITH question_ids AS (
    SELECT qei.id, ROW_NUMBER() OVER (ORDER BY qei.path) as display_order from (
        SELECT qh.js ->> 'id' as id,
        ARRAY[idx] as path 
        FROM quiz_sets qs,
        unnest(qs.question_hierarchy) with ordinality as qh(js, idx)
        where qs.lo_id  = id
        and qs.deleted_at is null
        UNION 
        SELECT qhci.cids as id,
        ARRAY[qh.idx, qhci.idx] as path
        FROM quiz_sets qs,
        unnest(qs.question_hierarchy) with ordinality as qh(js, idx),
        jsonb_array_elements_text(qh.js -> 'children_ids') with ordinality as qhci(cids, idx)
        where qs.lo_id  = id
        and qs.deleted_at is null
        and jsonb_typeof(qh.js -> 'children_ids') = 'array'
    ) qei
)
SELECT 
    quiz_id,
    country,
    school_id,
    external_id,
    kind,
    question,
    explanation,
    options,
    tagged_los,
    difficulty_level,
    created_by,
    approved_by,
    status,
    coalesce(q.updated_at, qr.updated_at) as updated_at ,
    coalesce(q.created_at, qr.created_at) as created_at,
    coalesce(q.deleted_at, qr.deleted_at) as deleted_at,
    lo_ids,
    coalesce(q.resource_path, qr.resource_path) as resource_path,
    point,
    coalesce(q.question_group_id, qr.question_group_id) as question_group_id,
    question_tag_ids,
    label_type
FROM question_ids qi
LEFT JOIN quizzes q on qi.id = q.external_id
LEFT JOIN question_group qr on qi.id = qr.question_group_id
order by qi.display_order ASC
$$;


ALTER FUNCTION public.find_question_by_lo_id(id character varying) OWNER TO postgres;

--
-- Name: find_quiz_by_lo_id(character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.find_quiz_by_lo_id(id character varying) RETURNS SETOF public.quizzes
    LANGUAGE sql STABLE
    AS $$
select q.* from quiz_sets qs, unnest(qs.quiz_external_ids) WITH ORDINALITY AS search_quiz_external_ids(quiz_external_id, ordinality)
                                  join quizzes q on (q.external_id::TEXT = search_quiz_external_ids.quiz_external_id )
where qs.deleted_at IS NULL AND q.deleted_at IS NULL and qs.lo_id = id
order by search_quiz_external_ids.ordinality ASC
    $$;


ALTER FUNCTION public.find_quiz_by_lo_id(id character varying) OWNER TO postgres;

--
-- Name: generate_ulid(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.generate_ulid() RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE
  -- Crockford's Base32
  encoding   BYTEA = '0123456789ABCDEFGHJKMNPQRSTVWXYZ';
  timestamp  BYTEA = E'\\000\\000\\000\\000\\000\\000';
  output     TEXT = '';

  unix_time  BIGINT;
  ulid       BYTEA;
BEGIN
  -- 6 timestamp bytes
  unix_time = (EXTRACT(EPOCH FROM NOW()) * 1000)::BIGINT;
  timestamp = SET_BYTE(timestamp, 0, (unix_time >> 40)::BIT(8)::INTEGER);
  timestamp = SET_BYTE(timestamp, 1, (unix_time >> 32)::BIT(8)::INTEGER);
  timestamp = SET_BYTE(timestamp, 2, (unix_time >> 24)::BIT(8)::INTEGER);
  timestamp = SET_BYTE(timestamp, 3, (unix_time >> 16)::BIT(8)::INTEGER);
  timestamp = SET_BYTE(timestamp, 4, (unix_time >> 8)::BIT(8)::INTEGER);
  timestamp = SET_BYTE(timestamp, 5, unix_time::BIT(8)::INTEGER);

  -- 10 entropy bytes
  ulid = timestamp || gen_random_bytes(10);

  -- Encode the timestamp
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 0) & 224) >> 5));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 0) & 31)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 1) & 248) >> 3));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 1) & 7) << 2) | ((GET_BYTE(ulid, 2) & 192) >> 6)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 2) & 62) >> 1));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 2) & 1) << 4) | ((GET_BYTE(ulid, 3) & 240) >> 4)));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 3) & 15) << 1) | ((GET_BYTE(ulid, 4) & 128) >> 7)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 4) & 124) >> 2));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 4) & 3) << 3) | ((GET_BYTE(ulid, 5) & 224) >> 5)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 5) & 31)));

  -- Encode the entropy
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 6) & 248) >> 3));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 6) & 7) << 2) | ((GET_BYTE(ulid, 7) & 192) >> 6)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 7) & 62) >> 1));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 7) & 1) << 4) | ((GET_BYTE(ulid, 8) & 240) >> 4)));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 8) & 15) << 1) | ((GET_BYTE(ulid, 9) & 128) >> 7)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 9) & 124) >> 2));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 9) & 3) << 3) | ((GET_BYTE(ulid, 10) & 224) >> 5)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 10) & 31)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 11) & 248) >> 3));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 11) & 7) << 2) | ((GET_BYTE(ulid, 12) & 192) >> 6)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 12) & 62) >> 1));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 12) & 1) << 4) | ((GET_BYTE(ulid, 13) & 240) >> 4)));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 13) & 15) << 1) | ((GET_BYTE(ulid, 14) & 128) >> 7)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 14) & 124) >> 2));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 14) & 3) << 3) | ((GET_BYTE(ulid, 15) & 224) >> 5)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 15) & 31)));

  RETURN output;
END
$$;


ALTER FUNCTION public.generate_ulid() OWNER TO postgres;

--
-- Name: get_assignment_scores(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_assignment_scores() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, student_submission_id text, graded_point smallint, total_point smallint, status text, passed boolean, created_at timestamp with time zone)
    LANGUAGE sql STABLE
    AS $$
select 
    ss.student_id,
    ss.study_plan_id,
    ss.learning_material_id,
    ss.student_submission_id,
    ss.correct_score::smallint as graded_point,
    ss.total_score::smallint as total_point,
    ss.status,
    ss.understanding_level != 'SUBMISSION_UNDERSTANDING_LEVEL_SAD' as passed,
    ss.created_at 
from student_submissions ss
    join assignment a using (learning_material_id)
order by ss.student_id, ss.study_plan_id, ss.learning_material_id, ss.created_at;
$$;


ALTER FUNCTION public.get_assignment_scores() OWNER TO postgres;

--
-- Name: get_assignment_scores_v2(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_assignment_scores_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, student_submission_id text, graded_point smallint, total_point smallint, status text, passed boolean, created_at timestamp with time zone)
    LANGUAGE sql STABLE
    AS $$
select 
    ss.student_id,
    ss.study_plan_id,
    ss.learning_material_id,
    ss.student_submission_id,
    ssg.grade::smallint as graded_point,
    a.max_grade::smallint as total_point,
    ss.status,
    ss.understanding_level != 'SUBMISSION_UNDERSTANDING_LEVEL_SAD' as passed,
    ss.created_at 
from student_submissions ss
join student_submission_grades ssg on ss.student_submission_id = ssg.student_submission_id
join assignment a using (learning_material_id)
where ssg.grade != -1
order by ss.student_id, ss.study_plan_id, ss.learning_material_id, ss.created_at;
$$;


ALTER FUNCTION public.get_assignment_scores_v2() OWNER TO postgres;

--
-- Name: get_exam_lo_returned_scores(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_exam_lo_returned_scores() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, graded_point smallint, total_point smallint, status text, result text, created_at timestamp with time zone)
    LANGUAGE sql STABLE
    AS $$
select els.student_id,
    els.study_plan_id,
    els.learning_material_id,
    els.submission_id,
    --   when teacher manual grading, we should use with score from teacher
    (
        (els.status = 'SUBMISSION_STATUS_RETURNED')::BOOLEAN::INT * sum(coalesce(elss.point, elsa.point))
    )::smallint as graded_point,
    els.total_point::smallint as total_point,
    els.status,
    els.result,
    els.created_at
from exam_lo_submission els
    join exam_lo_submission_answer elsa using (submission_id)
    left join exam_lo_submission_score elss using (submission_id, quiz_id)
group by els.submission_id $$;


ALTER FUNCTION public.get_exam_lo_returned_scores() OWNER TO postgres;

--
-- Name: get_exam_lo_scores(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_exam_lo_scores() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, graded_point smallint, total_point smallint, status text, result text, created_at timestamp with time zone)
    LANGUAGE sql STABLE
    AS $$
select els.student_id,
       els.study_plan_id,
       els.learning_material_id,
       els.submission_id,
    --   when teacher manual grading, we should use with score from teacher
       sum(coalesce(elss.point, elsa.point))::smallint as graded_point,
       els.total_point::smallint                       as total_point,
       els.status,
       els.result,
       els.created_at
from exam_lo_submission els
         join exam_lo_submission_answer elsa using (submission_id)
         left join exam_lo_submission_score elss using (submission_id, quiz_id)
where els.deleted_at is null and
    elsa.deleted_at is null and
    elss.deleted_at is null
group by els.student_id,
         els.study_plan_id,
         els.learning_material_id,
         els.submission_id
$$;


ALTER FUNCTION public.get_exam_lo_scores() OWNER TO postgres;

--
-- Name: get_exam_lo_scores_grade_to_pass(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_exam_lo_scores_grade_to_pass() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, graded_point smallint, total_point smallint, status text, passed boolean, total_attempts smallint)
    LANGUAGE sql STABLE
    AS $$
select distinct on (student_id, study_plan_id, learning_material_id) student_id,
                                                                     study_plan_id,
                                                                     learning_material_id,
                                                                     -- graded point is calculated
                                                                     -- if all submission are fails choose latest score
                                                                     -- if a submission is passed choose grade_to_pass from exam_lo setting
                                                                     -- if a submission is passed from 2nd
                                                                     -- ex : s1 failed, s2 pass, s3 pass
                                                                     -- we will calculate s2 = pass * count(s1,s2,s3) > 1
                                                                     coalesce(NULLIF(
                                                                                      (e.grade_to_pass *
                                                                                       (result = 'EXAM_LO_SUBMISSION_PASSED')::integer *
                                                                                       (count(*) over (
                                                                                           partition by student_id,
                                                                                               study_plan_id,
                                                                                               learning_material_id
                                                                                           ) > 1)::integer *
                                                                                       e.grade_capping::integer)::smallint,
                                                                                      0)
                                                                         , s.graded_point)                  as graded_point,
                                                                     total_point,
                                                                     status,
                                                                     (result = 'EXAM_LO_SUBMISSION_PASSED') as passed,
                                                                     count(*) over (partition by student_id,
                                                                         study_plan_id,
                                                                         learning_material_id)::smallint    as total_attempts
from get_exam_lo_scores() s
         join exam_lo e using (learning_material_id)
-- order by the submissions are passed -> then latest
order by student_id, study_plan_id, learning_material_id, (result = 'EXAM_LO_SUBMISSION_PASSED') desc, s.created_at desc
$$;


ALTER FUNCTION public.get_exam_lo_scores_grade_to_pass() OWNER TO postgres;

--
-- Name: get_exam_lo_scores_latest_score(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_exam_lo_scores_latest_score() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, graded_point smallint, total_point smallint, status text, passed boolean, total_attempts smallint)
    LANGUAGE sql STABLE
    AS $$
select distinct on (student_id, study_plan_id, learning_material_id) student_id,
                                                                     study_plan_id,
                                                                     learning_material_id,
                                                                     graded_point,
                                                                     total_point,
                                                                     status,
                                                                    --  true when a submission is passed
                                                                     bool_or(result = 'EXAM_LO_SUBMISSION_PASSED')
                                                                     over ( partition by student_id,
                                                                         study_plan_id,
                                                                         learning_material_id )          as passed,
                                                                     count(*) over (partition by student_id,
                                                                         study_plan_id,
                                                                         learning_material_id)::smallint as total_attempts
from get_exam_lo_scores()
-- get the latest scores by created_at
order by student_id, study_plan_id, learning_material_id, created_at desc;
$$;


ALTER FUNCTION public.get_exam_lo_scores_latest_score() OWNER TO postgres;

--
-- Name: course_students; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.course_students (
    course_id text NOT NULL,
    student_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    course_student_id text NOT NULL,
    start_at timestamp with time zone,
    end_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    vendor_synced_at timestamp with time zone
);

ALTER TABLE ONLY public.course_students FORCE ROW LEVEL SECURITY;


ALTER TABLE public.course_students OWNER TO postgres;

--
-- Name: get_list_course_student_study_plans_by_filter(text, text, text[], text, integer[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_list_course_student_study_plans_by_filter(_course_id text, search text, _book_ids text[], _status text, _grades integer[]) RETURNS SETOF public.course_students
    LANGUAGE sql STABLE
    AS $$

SELECT DISTINCT c_student.* FROM public.course_students as c_student

LEFT JOIN student_study_plans as s_study_plans
    USING(student_id)
LEFT JOIN study_plans as st
    USING(course_id, study_plan_id)

WHERE c_student.course_id = _course_id

AND c_student.deleted_at IS NULL

AND (
    -- When user don't apply filter should return matched
    (_book_ids = '{}' AND _grades = '{}' AND search = '')
    OR (
        st.deleted_at IS NULL
        AND st.name ILIKE ('%' || search || '%')
        AND st.status = _status
        AND (
            _book_ids = '{}' OR
            st.book_id = ANY(_book_ids)
        )
        AND (
            _grades = '{}' OR
        	EXISTS (SELECT * FROM UNNEST(grades) WHERE unnest = ANY(_grades))
        )
    )
)

$$;


ALTER FUNCTION public.get_list_course_student_study_plans_by_filter(_course_id text, search text, _book_ids text[], _status text, _grades integer[]) OWNER TO postgres;

--
-- Name: get_list_course_student_study_plans_by_filter_v2(text, text, text[], text, integer[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_list_course_student_study_plans_by_filter_v2(_course_id text, search text, _book_ids text[], _status text, _grades integer[]) RETURNS SETOF public.course_students
    LANGUAGE sql STABLE
    AS $$

SELECT DISTINCT c_student.* FROM public.course_students as c_student
LEFT JOIN student_study_plans as s_study_plans
    USING(student_id)
LEFT JOIN study_plans as st
    USING(course_id, study_plan_id)
WHERE c_student.course_id = _course_id
AND c_student.deleted_at IS NULL
AND (
    -- When user don't apply filter should return matched
    (_book_ids = '{}' AND _grades = '{}' AND search = '')
    OR (
        st.deleted_at IS NULL
        AND st.name ILIKE ('%' || search || '%')
        AND st.status = _status
        AND (
            _book_ids = '{}' OR
            st.book_id = ANY(_book_ids)
        )
        AND (
            _grades = '{}' OR
        	EXISTS (SELECT * FROM UNNEST(grades) WHERE unnest = ANY(_grades))
        )
    )
)

$$;


ALTER FUNCTION public.get_list_course_student_study_plans_by_filter_v2(_course_id text, search text, _book_ids text[], _status text, _grades integer[]) OWNER TO postgres;

--
-- Name: course_study_plans; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.course_study_plans (
    course_id text NOT NULL,
    study_plan_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.course_study_plans FORCE ROW LEVEL SECURITY;


ALTER TABLE public.course_study_plans OWNER TO postgres;

--
-- Name: get_list_course_study_plan_by_filter(text, text, text[], text[], integer[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_list_course_study_plan_by_filter(_course_id text, search text, _book_ids text[], _status text[], _grades integer[]) RETURNS SETOF public.course_study_plans
    LANGUAGE sql STABLE
    AS $$
SELECT  cs.* FROM public.course_study_plans as cs  JOIN study_plans as st 
USING(study_plan_id,course_id)

WHERE (_status = '{}' OR st.status = ANY(_status))
AND (_book_ids = '{}' OR st.book_id = ANY(_book_ids))
AND cs.deleted_at IS NULL
AND st.deleted_at IS NULL
AND cs.course_id = _course_id
AND st.name ilike ('%' || search || '%')
AND (
	EXISTS (SELECT * FROM UNNEST(grades) WHERE unnest = ANY(_grades)) 
	OR _grades = '{}'
)

$$;


ALTER FUNCTION public.get_list_course_study_plan_by_filter(_course_id text, search text, _book_ids text[], _status text[], _grades integer[]) OWNER TO postgres;

--
-- Name: get_max_assignment_scores(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_max_assignment_scores() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, student_submission_id text, graded_point smallint, total_point smallint, status text, passed boolean, total_attempts smallint, created_at timestamp with time zone)
    LANGUAGE sql STABLE
    AS $$
select distinct on (gas.student_id,gas.study_plan_id,gas.learning_material_id)
    gas.student_id,
    gas.study_plan_id,
    gas.learning_material_id,
    gas.student_submission_id,
    gas.graded_point,
    gas.total_point,
    gas.status,
    gas.passed,
     count(*)
    over (partition by gas.student_id, gas.study_plan_id, gas.learning_material_id)::smallint as total_attempts,
    gas.created_at 
from get_assignment_scores() gas
order by gas.student_id, gas.study_plan_id, gas.learning_material_id,gas.graded_point * 1.0 / coalesce(nullif(gas.total_point,0),1) desc;
$$;


ALTER FUNCTION public.get_max_assignment_scores() OWNER TO postgres;

--
-- Name: get_student_chapter_progress(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_student_chapter_progress() RETURNS TABLE(student_id text, study_plan_id text, chapter_id text, average_score smallint)
    LANGUAGE sql STABLE
    AS $$
select student_id,
       study_plan_id,
       chapter_id,
       avg(average_score)::smallint
from get_student_topic_progress()
group by student_id, study_plan_id, chapter_id
$$;


ALTER FUNCTION public.get_student_chapter_progress() OWNER TO postgres;

--
-- Name: get_student_completion_learning_material(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_student_completion_learning_material() RETURNS TABLE(study_plan_id text, student_id text, learning_material_id text, completed_at timestamp with time zone)
    LANGUAGE sql STABLE
    AS $$
(
    SELECT sel.study_plan_id,
        sel.student_id,
        sel.learning_material_id,
        max(sel.created_at) as completed_at
    FROM student_event_logs sel
    JOIN learning_material lm USING (learning_material_id)
    WHERE sel.payload ->> 'event' = ANY (ARRAY ['completed', 'exited'])
        AND CASE
            -- Exam LO type:
            WHEN lm.type = 'LEARNING_MATERIAL_EXAM_LO' THEN
                EXISTS(SELECT 1
                    FROM exam_lo_submission els
                    WHERE els.study_plan_id = sel.study_plan_id
                        AND els.student_id = sel.student_id
                        AND els.learning_material_id = sel.learning_material_id
                        AND els.deleted_at IS NULL
                )
            -- LO, FLASH_CARD type
            ELSE TRUE
        END
       GROUP BY sel.study_plan_id, sel.student_id, sel.learning_material_id
)
UNION ALL
(
    SELECT study_plan_id,
        student_id,
        learning_material_id,
        max(complete_date) as completed_at
    FROM student_submissions
    WHERE deleted_at IS NULL
    GROUP BY study_plan_id, student_id, learning_material_id
)
$$;


ALTER FUNCTION public.get_student_completion_learning_material() OWNER TO postgres;

--
-- Name: study_plans; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.study_plans (
    study_plan_id text NOT NULL,
    master_study_plan_id text,
    name text,
    study_plan_type text,
    school_id integer,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    course_id text,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    book_id text,
    status text DEFAULT 'STUDY_PLAN_STATUS_ACTIVE'::text,
    track_school_progress boolean DEFAULT false,
    grades integer[] DEFAULT '{}'::integer[],
    CONSTRAINT study_plan_status_check CHECK ((status = ANY (ARRAY['STUDY_PLAN_STATUS_NONE'::text, 'STUDY_PLAN_STATUS_ACTIVE'::text, 'STUDY_PLAN_STATUS_ARCHIVED'::text])))
);

ALTER TABLE ONLY public.study_plans FORCE ROW LEVEL SECURITY;


ALTER TABLE public.study_plans OWNER TO postgres;

--
-- Name: get_student_study_plans_by_filter(text, text, text[], text, integer[], text[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_student_study_plans_by_filter(_course_id text, search text, _book_ids text[], _status text, _grades integer[], _student_ids text[]) RETURNS SETOF public.study_plans
    LANGUAGE sql STABLE
    AS $$

select st.* from public.student_study_plans as s_study_plans join study_plans as st

USING(study_plan_id)

WHERE

    s_study_plans.deleted_at IS NULL
    AND st.deleted_at IS NULL

    AND s_study_plans.student_id = ANY(_student_ids)
    AND st.course_id = _course_id

    AND st.name ILIKE ('%' || search || '%')
    AND st.status = _status
    AND (
        _book_ids = '{}' OR
        st.book_id = ANY(_book_ids)
    )
    AND (
        _grades = '{}' OR
	    EXISTS (SELECT * FROM UNNEST(grades) WHERE unnest = ANY(_grades))
    );

$$;


ALTER FUNCTION public.get_student_study_plans_by_filter(_course_id text, search text, _book_ids text[], _status text, _grades integer[], _student_ids text[]) OWNER TO postgres;

--
-- Name: get_student_study_plans_by_filter_v2(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_student_study_plans_by_filter_v2() RETURNS TABLE(study_plan_id text, master_study_plan_id text, name text, study_plan_type text, school_id integer, created_at timestamp with time zone, updated_at timestamp with time zone, deleted_at timestamp with time zone, course_id text, resource_path text, book_id text, status text, track_school_progress boolean, grades integer[], student_id text)
    LANGUAGE sql STABLE
    AS $$

select st.*, s_study_plans.student_id
from public.individual_study_plan_fn() as s_study_plans
         join study_plans as st
              USING (study_plan_id)
$$;


ALTER FUNCTION public.get_student_study_plans_by_filter_v2() OWNER TO postgres;

--
-- Name: get_student_topic_progress(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_student_topic_progress() RETURNS TABLE(student_id text, study_plan_id text, chapter_id text, topic_id text, completed_sp_item smallint, total_sp_item smallint, average_score smallint)
    LANGUAGE sql STABLE
    AS $$
select student_id,
       study_plan_id,
       chapter_id,
       topic_id,
       (select count(*) from get_student_completion_learning_material() gsclm
       			join learning_material lm using (learning_material_id)
       			where gsclm.student_id = lalm.student_id
       			and gsclm.study_plan_id = lalm.study_plan_id
       			and lm.topic_id = lalm.topic_id)::smallint   as completed_sp_item,
       count(*)::smallint  as total_sp_item,
       (avg(gs.graded_points * 1.0 / gs.total_points) * 100)::smallint as average_score
from list_available_learning_material() lalm
         left join max_graded_score() gs
                   using (student_id, study_plan_id, learning_material_id)
group by student_id, study_plan_id, chapter_id, topic_id
$$;


ALTER FUNCTION public.get_student_topic_progress() OWNER TO postgres;

--
-- Name: get_submission_history(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_submission_history() RETURNS TABLE(student_id text, shuffled_quiz_set_id text, quiz_id text, student_text_answer text[], correct_text_answer text[], student_index_answer integer[], correct_index_answer integer[], submitted_keys_answer text[], correct_keys_answer text[], is_correct boolean[], is_accepted boolean, point integer, created_at timestamp with time zone, updated_at timestamp with time zone, deleted_at timestamp with time zone, resource_path text)
    LANGUAGE sql STABLE
    AS $$
    SELECT DISTINCT ON (shuffled_quiz_set_id, quiz_id)
 		student_id,
        shuffled_quiz_set_id,
        sh.quiz_id,
        sh.filled_text AS student_text_answer,
		sh.correct_text AS correct_text_answer,
        sh.selected_index AS student_index_answer,
        sh.correct_index AS correct_index_answer,
        sh.submitted_keys AS submitted_keys_answer,
        sh.correct_keys AS correct_keys_answer,
       	sh.correctness AS is_correct,
        sh.is_accepted AS is_accepted,
		COALESCE((sh.is_accepted)::BOOLEAN::INT*(SELECT point FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = sh.quiz_id LIMIT 1), 0) as point,
        (sh.submitted_at)::timestamp with time zone as created_at,
		updated_at,
		deleted_at,
		resource_path
    FROM shuffled_quiz_sets AS sqs, jsonb_to_recordset(sqs.submission_history) AS sh (
        quiz_id text,
        correctness BOOLEAN[],
        filled_text TEXT[],
        is_accepted BOOLEAN,
        correct_text TEXT[], 
        submitted_at timestamp with time zone,
        correct_index INTEGER[],
        selected_index INTEGER[],
        submitted_keys TEXT[],
        correct_keys TEXT[])
	WHERE sqs.deleted_at is NULL
    ORDER BY shuffled_quiz_set_id, quiz_id, created_at DESC
$$;


ALTER FUNCTION public.get_submission_history() OWNER TO postgres;

--
-- Name: get_task_assignment_scores(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_task_assignment_scores() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, student_submission_id text, graded_point smallint, total_point smallint, status text, passed boolean, created_at timestamp with time zone)
    LANGUAGE sql STABLE
    AS $$
select
    ss.student_id,
    ss.study_plan_id,
    ss.learning_material_id,
    ss.student_submission_id,
    ss.correct_score::smallint as graded_point,
        ss.total_score::smallint as total_point,
        ss.status,
    ss.understanding_level != 'SUBMISSION_UNDERSTANDING_LEVEL_SAD' as passed,
    ss.created_at
from student_submissions ss
    join task_assignment ta using (learning_material_id);
$$;


ALTER FUNCTION public.get_task_assignment_scores() OWNER TO postgres;

--
-- Name: graded_score(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.graded_score() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, graded_points smallint, total_points smallint, status text)
    LANGUAGE sql STABLE
    AS $$
select student_id,
    study_plan_id,
    learning_material_id,
    submission_id,
    graded_points,
    total_points,
    status
from lo_graded_score()
union all
select student_id,
    study_plan_id,
    learning_material_id,
    submission_id,
    graded_points,
    total_points,
    status
from fc_graded_score()
union all
select student_id,
    study_plan_id,
    learning_material_id,
    student_submission_id,
    graded_point,
    total_point,
    status
from get_assignment_scores()
union all
select student_id,
    study_plan_id,
    learning_material_id,
    student_submission_id,
    graded_point,
    total_point,
    status
from get_task_assignment_scores()
union all
select student_id,
    study_plan_id,
    learning_material_id,
    submission_id,
    graded_point,
    total_point,
    status
from get_exam_lo_returned_scores() where status = 'SUBMISSION_STATUS_RETURNED'
$$;


ALTER FUNCTION public.graded_score() OWNER TO postgres;

--
-- Name: graded_score_v2(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.graded_score_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, graded_point smallint, total_point smallint, status text)
    LANGUAGE sql STABLE
    AS $$
select student_id,
    study_plan_id,
    learning_material_id,
    submission_id,
    graded_point,
    total_point,
    status
from lo_graded_score_v2()
union all
select student_id,
    study_plan_id,
    learning_material_id,
    submission_id,
    graded_point,
    total_point,
    status
from fc_graded_score_v2()
union all
select student_id,
    study_plan_id,
    learning_material_id,
    student_submission_id,
    graded_point,
    total_point,
    status
from assignment_graded_score_v2()
union all
select student_id,
    study_plan_id,
    learning_material_id,
    student_submission_id,
    graded_point,
    total_point,
    status
from task_assignment_graded_score_v2()
union all
select student_id,
    study_plan_id,
    learning_material_id,
    submission_id,
    graded_point,
    total_point,
    status
from exam_lo_graded_score_v2() 
$$;


ALTER FUNCTION public.graded_score_v2() OWNER TO postgres;

--
-- Name: individual_study_plan_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.individual_study_plan_fn() RETURNS TABLE(student_id text, study_plan_id text, book_id text, chapter_id text, chapter_display_order smallint, topic_id text, topic_display_order smallint, learning_material_id text, lm_display_order smallint, start_date timestamp with time zone, end_date timestamp with time zone, available_from timestamp with time zone, available_to timestamp with time zone, school_date timestamp with time zone, updated_at timestamp with time zone, status text, resource_path text)
    LANGUAGE sql STABLE
    AS $$
SELECT ssp.student_id,
       m.study_plan_id,
       m.book_id,
       m.chapter_id,
       m.chapter_display_order,
       m.topic_id,
       m.topic_display_order,
       m.learning_material_id,
       m.lm_display_order,
       public.check_study_plan_item_time(m.updated_at, isp.updated_at, m.start_date, isp.start_date) AS start_date,
       public.check_study_plan_item_time(m.updated_at, isp.updated_at, m.end_date,
                                         isp.end_date)                                               AS end_date,
       public.check_study_plan_item_time(m.updated_at, isp.updated_at, m.available_from,
                                         isp.available_from)                                         AS available_from,
       public.check_study_plan_item_time(m.updated_at, isp.updated_at, m.available_to,
                                         isp.available_to)                                           AS available_to,
       public.check_study_plan_item_time(m.updated_at, isp.updated_at, m.school_date,
                                         isp.school_date)                                            AS school_date,
       public.check_study_plan_item_time(m.updated_at, isp.updated_at, m.updated_at,
                                         isp.updated_at)                                             AS updated_at,
       CASE
           WHEN ((m.updated_at IS NULL) OR (isp.updated_at IS NULL)) THEN COALESCE(m.status, isp.status)
           ELSE
               CASE
                   WHEN (m.updated_at >= isp.updated_at) THEN m.status
                   ELSE isp.status
                   END
           END                                                                                       AS status,
       m.resource_path
FROM (public.master_study_plan_view m
    join (select cs.course_id, cs.student_id, sp.study_plan_id, sp.study_plan_type, cs.deleted_at
            from study_plans sp
            join course_students cs on cs.course_id = sp.course_id
            where sp.master_study_plan_id is null
        ) ssp ON m.study_plan_id = ssp.study_plan_id
    left join student_study_plans ssp_a
       on ssp.student_id = ssp_a.student_id and ssp.study_plan_id = ssp_a.study_plan_id
    LEFT JOIN public.individual_study_plan isp
      ON (((ssp.student_id = isp.student_id) AND (m.learning_material_id = isp.learning_material_id) AND
           (m.study_plan_id = isp.study_plan_id))))
    where (ssp_a.student_id is not null OR ssp.study_plan_type = 'STUDY_PLAN_TYPE_COURSE')
$$;


ALTER FUNCTION public.individual_study_plan_fn() OWNER TO postgres;

--
-- Name: is_table_in_publication(text, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.is_table_in_publication(publication_name text, table_name text) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN EXISTS (
    SELECT 1
    FROM pg_publication_rel pr
    JOIN pg_class c ON pr.prrelid = c.oid
    JOIN pg_namespace n ON c.relnamespace = n.oid
    JOIN pg_publication p ON pr.prpubid = p.oid
    WHERE p.pubname = publication_name AND c.relname = table_name
  );
END;
$$;


ALTER FUNCTION public.is_table_in_publication(publication_name text, table_name text) OWNER TO postgres;

--
-- Name: list_available_learning_material(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.list_available_learning_material() RETURNS TABLE(student_id text, study_plan_id text, book_id text, chapter_id text, chapter_display_order smallint, topic_id text, topic_display_order smallint, learning_material_id text, lm_display_order smallint, available_from timestamp with time zone, available_to timestamp with time zone, start_date timestamp with time zone, end_date timestamp with time zone, status text, school_date timestamp with time zone)
    LANGUAGE sql STABLE
    AS $$
select student_id,
       study_plan_id,
       book_id,
       chapter_id,
       chapter_display_order,
       topic_id,
       topic_display_order,
       learning_material_id,
       lm_display_order,
       available_from,
       available_to,
       start_date,
       end_date,
       status,
       school_date
from individual_study_plan_fn()
where (now() between available_from and available_to)
   or (available_from <= now() and available_to is null)
$$;


ALTER FUNCTION public.list_available_learning_material() OWNER TO postgres;

--
-- Name: list_individual_study_plan_item(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.list_individual_study_plan_item() RETURNS TABLE(study_plan_id text, learning_material_id text, student_id text, available_from timestamp with time zone, available_to timestamp with time zone, start_date timestamp with time zone, end_date timestamp with time zone, status text, school_date timestamp with time zone, completed_at timestamp with time zone, lm_display_order smallint, scorce smallint, type text)
    LANGUAGE sql STABLE
    AS $$
select distinct on (study_plan_id, learning_material_id, student_id)
    isp.study_plan_id,
    isp.learning_material_id,
    isp.student_id,
    isp.available_from,
    isp.available_to,
    isp.start_date,
    isp.end_date,
    isp.status,
    isp.school_date,
    gsl.completed_at,
    isp.lm_display_order,
	coalesce((gs.graded_points * 1.0 / gs.total_points) * 100, null)::smallint AS scorce,
    lm.type
FROM list_available_learning_material() AS isp
INNER JOIN learning_material lm using (learning_material_id)
LEFT JOIN get_student_completion_learning_material() gsl using(student_id, study_plan_id, learning_material_id)
LEFT JOIN max_graded_score() gs using (student_id, study_plan_id, learning_material_id)
where
	lm.deleted_at is null
	AND
    isp.status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE' $$;


ALTER FUNCTION public.list_individual_study_plan_item() OWNER TO postgres;

--
-- Name: lo_answer(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.lo_answer() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, external_quiz_id text, is_accepted boolean, point integer)
    LANGUAGE sql STABLE
    AS $$
select student_id,
       study_plan_id,
       learning_material_id,
       submission_id,
       external_quiz_id,
       bool_or(is_accepted) as is_accepted,
       q.point
from (select student_id,
             study_plan_id,
             learning_material_id,
             submission_id,
             (jsonb_array_elements(submission_history) ->> 'quiz_id')::text     as external_quiz_id,
             (jsonb_array_elements(submission_history) ->> 'is_accepted')::bool as is_accepted
      from lo_raw_answer() sqs
      union
      select student_id,
             study_plan_id,
             learning_material_id,
             submission_id,
             unnest(quiz_external_ids) as external_quiz_id,
--          not in submission, so answer default is wrong
             false                     as is_accepted
      from lo_raw_answer() sqs) raw_answer
         join quizzes q
              on raw_answer.external_quiz_id = q.external_id
group by student_id,
         study_plan_id,
         learning_material_id,
         submission_id,
         external_quiz_id,
         q.point
$$;


ALTER FUNCTION public.lo_answer() OWNER TO postgres;

--
-- Name: lo_answer_v2(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.lo_answer_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, external_quiz_id text, is_accepted boolean, point integer, total_point integer)
    LANGUAGE sql STABLE
    AS $$
select  sa.student_id,
        sa.study_plan_id,
        sa.learning_material_id,
        sa.submission_id,
        sa.quiz_id,
        sa.is_accepted,
        point,
        s.total_point
from lo_submission_answer sa
join lo_submission s using (submission_id)
$$;


ALTER FUNCTION public.lo_answer_v2() OWNER TO postgres;

--
-- Name: lo_graded_score(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.lo_graded_score() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, graded_points smallint, total_points smallint, status text)
    LANGUAGE sql STABLE
    AS $$
select student_id,
       study_plan_id,
       learning_material_id,
       submission_id,
       sum(is_accepted::int * point)::smallint as graded_score,
       sum(point)::smallint                    as total_scores,
       'S'
from lo_answer()
group by student_id,
         study_plan_id,
         learning_material_id,
         submission_id
$$;


ALTER FUNCTION public.lo_graded_score() OWNER TO postgres;

--
-- Name: lo_graded_score_v2(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.lo_graded_score_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, graded_point smallint, total_point smallint, status text)
    LANGUAGE sql STABLE
    AS $$
select  sa.student_id,
        sa.study_plan_id,
        sa.learning_material_id,
        sa.submission_id,
        sum(point)::smallint as graded_point,
        max(s.total_point)::smallint as total_point,
        'S'
from lo_submission_answer sa
join lo_submission s using (submission_id)
where s.is_submitted is true and s.deleted_at is null
group by sa.student_id,
         sa.study_plan_id,
         sa.learning_material_id,
         sa.submission_id
$$;


ALTER FUNCTION public.lo_graded_score_v2() OWNER TO postgres;

--
-- Name: lo_raw_answer(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.lo_raw_answer() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, submission_history jsonb, quiz_external_ids text[])
    LANGUAGE sql STABLE
    AS $$
select student_id,
       study_plan_id,
       learning_material_id,
       shuffled_quiz_set_id as submission_id,
       submission_history,
       quiz_external_ids
from shuffled_quiz_sets
join learning_objective using (learning_material_id)
$$;


ALTER FUNCTION public.lo_raw_answer() OWNER TO postgres;

--
-- Name: master_study_plan_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.master_study_plan_fn() RETURNS TABLE(study_plan_id text, book_id text, chapter_id text, chapter_display_order smallint, topic_id text, topic_display_order smallint, learning_material_id text, lm_display_order smallint, resource_path text, start_date timestamp with time zone, end_date timestamp with time zone, available_from timestamp with time zone, available_to timestamp with time zone, school_date timestamp with time zone, updated_at timestamp with time zone, status text)
    LANGUAGE sql STABLE
    AS $$
SELECT sp.*, m.start_date, m.end_date, m.available_from, m.available_to, m.school_date, m.updated_at, m.status
FROM study_plan_tree sp
         LEFT JOIN master_study_plan m USING (study_plan_id, learning_material_id);
$$;


ALTER FUNCTION public.master_study_plan_fn() OWNER TO postgres;

--
-- Name: max_graded_score(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.max_graded_score() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, graded_points smallint, total_points smallint)
    LANGUAGE sql STABLE
    AS $$
select distinct on (student_id, 
    study_plan_id ,
    learning_material_id) student_id,
                          study_plan_id,
                          learning_material_id,
                          graded_points,
                          total_points
from
    graded_score()
where total_points > 0
order by student_id, study_plan_id, learning_material_id, graded_points * 1.0 / total_points desc
$$;


ALTER FUNCTION public.max_graded_score() OWNER TO postgres;

--
-- Name: max_graded_score_v2(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.max_graded_score_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, graded_point smallint, total_point smallint)
    LANGUAGE sql STABLE
    AS $$
select distinct on (student_id, 
    study_plan_id ,
    learning_material_id) student_id,
                          study_plan_id,
                          learning_material_id,
                          graded_point,
                          total_point
from
    graded_score_v2()
where total_point > 0
order by student_id, study_plan_id, learning_material_id, graded_point * 1.0 / total_point desc
$$;


ALTER FUNCTION public.max_graded_score_v2() OWNER TO postgres;

--
-- Name: migrate_learning_objectives_to_exam_lo_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.migrate_learning_objectives_to_exam_lo_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    INSERT INTO exam_lo (
        learning_material_id,
        topic_id,
        name,
        type,
        display_order,
        created_at,
        updated_at,
        deleted_at,
        resource_path,
        instruction,
        grade_to_pass,
        manual_grading,
        time_limit,
        maximum_attempt,
        approve_grading,
        grade_capping,
        review_option
    )
    VALUES (
        NEW.lo_id,
        NEW.topic_id,
        NEW.name,
        'LEARNING_MATERIAL_EXAM_LO',
        NEW.display_order,
        NEW.created_at,
        NEW.updated_at,
        NEW.deleted_at,
        NEW.resource_path,
        NEW.instruction,
        NEW.grade_to_pass,
        NEW.manual_grading,
        NEW.time_limit,
        NEW.maximum_attempt,
        NEW.approve_grading,
        NEW.grade_capping,
        NEW.review_option
    )
    ON CONFLICT ON CONSTRAINT exam_lo_pk DO UPDATE SET
        topic_id = NEW.topic_id,
        name = NEW.name,
        display_order = NEW.display_order,
        updated_at = NEW.updated_at,
        deleted_at = NEW.deleted_at,
        instruction = NEW.instruction,
        grade_to_pass = NEW.grade_to_pass,
        manual_grading = NEW.manual_grading,
        time_limit = NEW.time_limit,
        maximum_attempt = NEW.maximum_attempt,
        approve_grading = NEW.approve_grading,
        grade_capping = NEW.grade_capping,
        review_option = NEW.review_option;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.migrate_learning_objectives_to_exam_lo_fn() OWNER TO postgres;

--
-- Name: migrate_study_plan_items_to_master_study_plan_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.migrate_study_plan_items_to_master_study_plan_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
-- Condition to specify storing master study plan items
-- 1: UPDATE only
-- WHY don't store insert, which always insert null values
    WITH temp_table as (
        SELECT nt.*
        FROM public.study_plans sp
        JOIN new_table nt 
            ON nt.study_plan_id = sp.study_plan_id
        WHERE sp.master_study_plan_id IS NULL
          AND sp.study_plan_type = 'STUDY_PLAN_TYPE_COURSE'
          AND nt.created_at <> nt.updated_at
    )
    INSERT INTO master_study_plan (
        study_plan_id,
        learning_material_id,
        status,
        start_date,
        end_date,
        available_from,
        available_to,
        created_at,
        updated_at,
        deleted_at,
        school_date,
        resource_path
    )
    SELECT
        study_plan_id,
        coalesce(NULLIF(content_structure ->> 'lo_id',''),content_structure->>'assignment_id'),
        status,
        start_date,
        end_date,
        available_from,
        available_to,
        created_at,
        updated_at,
        deleted_at,
        school_date,
        resource_path
    FROM temp_table

    ON CONFLICT ON CONSTRAINT learning_material_id_study_plan_id_pk DO UPDATE SET
      start_date = EXCLUDED.start_date,
      end_date = EXCLUDED.end_date,
      available_from = EXCLUDED.available_from,
      available_to = EXCLUDED.available_to,
      school_date = EXCLUDED.school_date,
      updated_at = EXCLUDED.updated_at,
      status = EXCLUDED.status,
      deleted_at = EXCLUDED.deleted_at;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.migrate_study_plan_items_to_master_study_plan_fn() OWNER TO postgres;

--
-- Name: migrate_to_exam_lo_submission_and_answer_once_submitted(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.migrate_to_exam_lo_submission_and_answer_once_submitted() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
_submission_id text;
BEGIN
    -- It's from Exam LO type
    -- and student actually submitted their answer (trigger after update of updated_at)
    IF EXISTS(
        SELECT 1 FROM exam_lo WHERE learning_material_id = NEW.learning_material_id
    )
    THEN
        INSERT INTO public.exam_lo_submission (
            submission_id,
            student_id,
            study_plan_id,
            learning_material_id,
            shuffled_quiz_set_id,
            status,
            result,
            created_at,
            updated_at,
            deleted_at,
            total_point,
            resource_path
        )
        VALUES (
            generate_ulid(),
            NEW.student_id,
            NEW.study_plan_id,
            NEW.learning_material_id,
            NEW.shuffled_quiz_set_id,
            'SUBMISSION_STATUS_RETURNED',
            'EXAM_LO_SUBMISSION_COMPLETED',
            NEW.updated_at, -- els.created_at == trigger by sqs.updated_at.
            NEW.updated_at, -- els.created_at & els.updated_at is the same.
            NEW.deleted_at,
            COALESCE((SELECT SUM(point) FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = ANY(NEW.quiz_external_ids)), 0),
            NEW.resource_path
        )
        ON CONFLICT ON CONSTRAINT shuffled_quiz_set_id_un DO UPDATE SET
        student_id = EXCLUDED.student_id,
        study_plan_id = EXCLUDED.study_plan_id,
        learning_material_id = EXCLUDED.learning_material_id,
        status = EXCLUDED.status,
        result = EXCLUDED.result,
        created_at = EXCLUDED.created_at,
        updated_at = EXCLUDED.updated_at,
        deleted_at = EXCLUDED.deleted_at,
        total_point = EXCLUDED.total_point
        RETURNING submission_id into _submission_id;

INSERT INTO public.exam_lo_submission_answer (
    student_id,
    submission_id,
    study_plan_id,
    learning_material_id,
    shuffled_quiz_set_id,
    quiz_id,
    student_text_answer,
    correct_text_answer,
    student_index_answer,
    correct_index_answer,
    submitted_keys_answer,
    correct_keys_answer,
    is_correct,
    is_accepted,
    point,
    created_at,
    updated_at,
    deleted_at,
    resource_path
)
SELECT NEW.student_id,
       _submission_id AS submission_id,
       NEW.study_plan_id,
       NEW.learning_material_id,
       NEW.shuffled_quiz_set_id,
       SA.quiz_id,
       (SELECT array_agg(obj) FROM jsonb_array_elements_text((SA.quiz_history->>'filled_text')::jsonb) X(obj)) AS student_text_answer,
       (SELECT array_agg(obj) FROM jsonb_array_elements_text((SA.quiz_history->>'correct_text')::jsonb) X(obj)) AS correct_text_answer,
       (SELECT array_agg(obj) FROM jsonb_array_elements_text((SA.quiz_history->>'selected_index')::jsonb) X(obj))::INTEGER[] AS student_index_answer,
        (SELECT array_agg(obj) FROM jsonb_array_elements_text((SA.quiz_history->>'correct_index')::jsonb) X(obj))::INTEGER[] AS correct_index_answer,
        (SELECT array_agg(obj) FROM jsonb_array_elements_text((SA.quiz_history->>'submitted_keys')::jsonb) X(obj)) AS submitted_keys_answer,
       (SELECT array_agg(obj) FROM jsonb_array_elements_text((SA.quiz_history->>'correct_keys')::jsonb) X(obj)) AS correct_keys_answer,
       ARRAY(SELECT jsonb_array_elements_text((SA.quiz_history->>'correctness')::jsonb))::BOOLEAN[] AS is_correct,
        (SA.quiz_history->>'is_accepted')::BOOLEAN AS is_accepted,
        CASE WHEN SA.quiz_history IS NOT NULL THEN
                 COALESCE((SA.quiz_history->>'is_accepted')::BOOLEAN::INT*(SELECT point FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = SA.quiz_id), 0)
             ELSE 0
            END AS point, -- If there is no answer for the question, then 0 point as default.
       NEW.updated_at, -- elsa.created_at == trigger by sqs.updated_at
       NEW.updated_at, -- els.created_at & els.updated_at is the same.
       NEW.deleted_at,
       NEW.resource_path
       -- The table contains the latest quiz_history by quiz_id, which in submission_history column.
FROM (SELECT quiz_id,
             (SELECT DISTINCT ON (X.obj ->> 'quiz_id') X.obj
      FROM public.shuffled_quiz_sets Y
          CROSS JOIN jsonb_array_elements(Y.submission_history) X(obj)
      WHERE Y.shuffled_quiz_set_id = SQ.shuffled_quiz_set_id
        AND X.obj ->> 'quiz_id' = SQ.quiz_id
      ORDER BY X.obj ->> 'quiz_id', X.obj->>'submitted_at' DESC) AS quiz_history
     -- For each record in exam_lo_submission table, there will be respectively n records in exam_lo_submission_answer table
     -- based on shuffled_quiz_sets.quiz_external_ids column.
     -- In case of quiz_external_ids is null, there is no record created (UNNEST deals with it).
    FROM (SELECT shuffled_quiz_set_id,
                UNNEST(quiz_external_ids) AS quiz_id
                FROM public.shuffled_quiz_sets
                WHERE shuffled_quiz_set_id = NEW.shuffled_quiz_set_id) SQ
               ) SA
ON CONFLICT ON CONSTRAINT exam_lo_submission_answer_pk DO UPDATE SET
    study_plan_id = EXCLUDED.study_plan_id,
    learning_material_id = EXCLUDED.learning_material_id,
    shuffled_quiz_set_id = EXCLUDED.shuffled_quiz_set_id,
    student_text_answer = EXCLUDED.student_text_answer,
    correct_text_answer = EXCLUDED.correct_text_answer,
    student_index_answer = EXCLUDED.student_index_answer,
    correct_index_answer = EXCLUDED.correct_index_answer,
    submitted_keys_answer = EXCLUDED.submitted_keys_answer,
    correct_keys_answer = EXCLUDED.correct_keys_answer,
    is_correct = EXCLUDED.is_correct,
    is_accepted = EXCLUDED.is_accepted,
    point = EXCLUDED.point,
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at,
    deleted_at = EXCLUDED.deleted_at;
END IF;
RETURN NULL;
END;
$$;


ALTER FUNCTION public.migrate_to_exam_lo_submission_and_answer_once_submitted() OWNER TO postgres;

--
-- Name: migrate_to_exam_lo_submission_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.migrate_to_exam_lo_submission_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM public.shuffled_quiz_sets SQ
         WHERE SQ.shuffled_quiz_set_id = NEW.shuffled_quiz_set_id
           AND SQ.original_shuffle_quiz_set_id IS NULL
           AND EXISTS (SELECT 1
                         FROM public.study_plan_items SP
                        WHERE SP.study_plan_item_id = SQ.study_plan_item_id
                          AND SP.completed_at IS NOT NULL
                          AND EXISTS (SELECT 1 FROM public.exam_lo WHERE learning_material_id = SP.content_structure->>'lo_id'))
    )
    THEN
        INSERT INTO public.exam_lo_submission (
            submission_id,
            student_id,
            study_plan_id,
            learning_material_id,
            shuffled_quiz_set_id,
            status,
            result,
            created_at,
            updated_at,
            deleted_at,
            total_point,
            resource_path
        )
        VALUES (
            generate_ulid(),
            NEW.student_id,
            (SELECT SP.master_study_plan_id
               FROM study_plans SP
                   LEFT JOIN study_plan_items SPI
                       ON SPI.study_plan_id = SP.study_plan_id
              WHERE SPI.study_plan_item_id = NEW.study_plan_item_id),
            (SELECT content_structure->>'lo_id' FROM public.study_plan_items WHERE study_plan_item_id = NEW.study_plan_item_id),
            NEW.shuffled_quiz_set_id,
            'SUBMISSION_STATUS_RETURNED',
            'EXAM_LO_SUBMISSION_COMPLETED',
            NEW.created_at,
            NEW.updated_at,
            NEW.deleted_at,
            COALESCE((SELECT SUM(point) FROM public.quizzes WHERE external_id = ANY(NEW.quiz_external_ids)), 0),
            NEW.resource_path
        )
        ON CONFLICT ON CONSTRAINT shuffled_quiz_set_id_un DO UPDATE SET
            student_id = NEW.student_id,
            study_plan_id = (SELECT SP.master_study_plan_id
                               FROM study_plans SP
                                   LEFT JOIN study_plan_items SPI
                                       ON SPI.study_plan_id = SP.study_plan_id
                              WHERE SPI.study_plan_item_id = NEW.study_plan_item_id),
            learning_material_id = (SELECT content_structure->>'lo_id' FROM public.study_plan_items WHERE study_plan_item_id = NEW.study_plan_item_id),
            status = 'SUBMISSION_STATUS_RETURNED',
            result = 'EXAM_LO_SUBMISSION_COMPLETED',
            created_at = NEW.created_at,
            updated_at = NEW.updated_at,
            deleted_at = NEW.deleted_at,
            total_point = COALESCE((SELECT SUM(point) FROM public.quizzes WHERE external_id = ANY(NEW.quiz_external_ids)), 0);
    END IF;
RETURN NULL;
END;
$$;


ALTER FUNCTION public.migrate_to_exam_lo_submission_fn() OWNER TO postgres;

--
-- Name: migrate_to_flash_card_submission_and_flash_card_submission_answ(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.migrate_to_flash_card_submission_and_flash_card_submission_answ() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    _submission_id text;
BEGIN
    IF EXISTS(
      SELECT 1 FROM flash_card fc WHERE fc.learning_material_id = NEW.learning_material_id
    )
    THEN
        INSERT INTO public.flash_card_submission (
            submission_id,
            student_id,
            study_plan_id,
            learning_material_id,
            shuffled_quiz_set_id,
            created_at,
            updated_at,
            deleted_at,
            total_point,
            resource_path
        )
        VALUES (
            generate_ulid(),
            NEW.student_id,
            NEW.study_plan_id,
            NEW.learning_material_id,
            NEW.shuffled_quiz_set_id,
            NEW.created_at,
            NEW.updated_at,
            NEW.deleted_at,
            COALESCE((SELECT SUM(point) FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = ANY(NEW.quiz_external_ids)), 0),
            NEW.resource_path
        )
        ON CONFLICT ON CONSTRAINT flash_card_submission_shuffled_quiz_set_id_un DO UPDATE SET
            updated_at = EXCLUDED.updated_at,
            deleted_at = EXCLUDED.deleted_at,
            total_point = EXCLUDED.total_point
        RETURNING submission_id into _submission_id;

	    INSERT INTO flash_card_submission_answer(
            student_id,
            quiz_id,
            submission_id,
            study_plan_id,
            learning_material_id,
            shuffled_quiz_set_id,
            student_text_answer,
            correct_text_answer,
            student_index_answer,
            correct_index_answer, 
            is_correct,
            is_accepted,
            point,
            created_at,
            updated_at,
            deleted_at,
            resource_path
        ) SELECT 
            NEW.student_id,
            fca.quiz_id,
            _submission_id AS submission_id,
            NEW.study_plan_id       ,
            NEW.learning_material_id,
            NEW.shuffled_quiz_set_id,
            fca.filled_text AS student_text_answer,
            fca.correct_text AS correct_text_answer,
            fca.selected_index AS student_index_answer,
            fca.correct_index AS correct_index_answer,
	        COALESCE(fca.correctness, '{}')::BOOLEAN[] AS is_correct, -- to avoid not null constraint
            fca.is_accepted AS is_accepted,
            COALESCE(fca.is_accepted::INT*(SELECT point FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = fca.quiz_id), 0) AS point, 
            COALESCE(fca.submitted_at, NEW.created_at) AS created_at, -- to avoid not null constraint
            NEW.updated_at,
            NEW.deleted_at,
            NEW.resource_path
	    FROM (
            SELECT DISTINCT ON(quiz_id) shuffled_quiz_set_id,
                x.* FROM shuffled_quiz_sets sqs, jsonb_to_recordset(sqs.submission_history)
            AS x (quiz_id TEXT,  
                filled_text text[],
                correct_text TEXT[], 
                selected_index INTEGER[], 
                correct_index INTEGER[],
                correctness BOOLEAN[], 
                submitted_at timestamp with time zone,
                is_accepted BOOLEAN)
            WHERE sqs.shuffled_quiz_set_id = NEW.shuffled_quiz_set_id
            ORDER BY quiz_id, submitted_at desc
             -- sort by submitted_at desc to ensure get the lastest answer
        ) fca
	    ON CONFLICT ON CONSTRAINT flash_card_submission_answer_pk DO UPDATE SET
            student_text_answer = EXCLUDED.student_text_answer,
            correct_text_answer = EXCLUDED.correct_text_answer,
            student_index_answer = EXCLUDED.student_index_answer,
            correct_index_answer = EXCLUDED.correct_index_answer,
            is_correct = EXCLUDED.is_correct,
            is_accepted = EXCLUDED.is_accepted,
            point = EXCLUDED.point,
            updated_at = EXCLUDED.updated_at,
            deleted_at = EXCLUDED.deleted_at;	  
    END IF;
RETURN NULL;
END;   
$$;


ALTER FUNCTION public.migrate_to_flash_card_submission_and_flash_card_submission_answ() OWNER TO postgres;

--
-- Name: migrate_to_lo_submission_and_answer_fnc(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.migrate_to_lo_submission_and_answer_fnc() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  -- insert to lo submission first
    IF EXISTS (
        SELECT 1 
        FROM public.learning_objective LO 
        WHERE LO.learning_material_id = NEW.learning_material_id and NEW.submission_history::text != '[]'::text
    )
    THEN
        INSERT INTO public.lo_submission (
            submission_id,
            student_id,
            study_plan_id,
            learning_material_id,
            shuffled_quiz_set_id,
            total_point,
            created_at,
            updated_at,
            deleted_at,
            resource_path
        )
        VALUES (
            generate_ulid(),
            NEW.student_id,
            NEW.study_plan_id,
            NEW.learning_material_id,
            NEW.shuffled_quiz_set_id,
            COALESCE(
                (
                    SELECT SUM(point) 
                    FROM public.quizzes 
                    WHERE quizzes.deleted_at IS NULL AND quizzes.external_id = ANY(SELECT unnest(quiz_external_ids) FROM quiz_sets qs WHERE 
                    qs.quiz_set_id = new.original_quiz_set_id)
                    ), 0),
            NEW.created_at,
            NEW.updated_at,
            NEW.deleted_at,
            NEW.resource_path
        )
        ON CONFLICT ON CONSTRAINT shuffled_quiz_set_id_lo_submission_un DO UPDATE SET
            updated_at = EXCLUDED.updated_at,
            deleted_at = EXCLUDED.deleted_at,
            total_point = EXCLUDED.total_point;

  -- continue insert to lo answer 
        INSERT INTO public.lo_submission_answer(
        student_id,
        quiz_id,
        submission_id,
        study_plan_id,
        learning_material_id,
        shuffled_quiz_set_id,
        student_text_answer,
        correct_text_answer,
        student_index_answer,
        correct_index_answer,
        submitted_keys_answer,
        correct_keys_answer,
        point,
        is_correct,
        is_accepted,
        created_at,
        updated_at,
        deleted_at,
        resource_path
    )
    SELECT 
        sh.student_id,
        sh.quiz_id,
        ls.submission_id,
        ls.study_plan_id,
        ls.learning_material_id,
        sh.shuffled_quiz_set_id,
        sh.student_text_answer,
        sh.correct_text_answer,
        sh.student_index_answer,
        sh.correct_index_answer,
        sh.submitted_keys_answer,
        sh.correct_keys_answer,
        sh.point,
        sh.is_correct,
        sh.is_accepted,
        sh.created_at,
        sh.updated_at,
        sh.deleted_at,
        sh.resource_path
    FROM get_submission_history() AS sh
    JOIN lo_submission ls USING(shuffled_quiz_set_id)
    JOIN quizzes q ON q.external_id = sh.quiz_id
    WHERE ls.deleted_at IS NULL
        AND q.deleted_at IS NULL
        AND sh.shuffled_quiz_set_id = NEW.shuffled_quiz_set_id
    ON CONFLICT ON CONSTRAINT lo_submission_answer_pk DO UPDATE SET
        student_text_answer = EXCLUDED.student_text_answer,
        correct_text_answer = EXCLUDED.correct_text_answer,
        student_index_answer = EXCLUDED.student_index_answer,
        correct_index_answer = EXCLUDED.correct_index_answer,
        submitted_keys_answer = EXCLUDED.submitted_keys_answer,
        correct_keys_answer = EXCLUDED.correct_keys_answer,
        point = EXCLUDED.point,
        is_correct = EXCLUDED.is_correct,
        is_accepted = EXCLUDED.is_accepted,
        updated_at = EXCLUDED.updated_at,
        deleted_at = EXCLUDED.deleted_at;
    END IF;
  
RETURN NULL;
END;
$$;


ALTER FUNCTION public.migrate_to_lo_submission_and_answer_fnc() OWNER TO postgres;

--
-- Name: migrate_to_lo_submission_answer_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.migrate_to_lo_submission_answer_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    INSERT INTO public.lo_submission_answer(
        student_id,
        quiz_id,
        submission_id,
        study_plan_id,
        learning_material_id,
        shuffled_quiz_set_id,
        student_text_answer,
        correct_text_answer,
        student_index_answer,
        correct_index_answer,
        submitted_keys_answer,
        correct_keys_answer,
        point,
        is_correct,
        is_accepted,
        created_at,
        updated_at,
        deleted_at,
        resource_path
    )
    SELECT 
        sh.student_id,
        sh.quiz_id,
        ls.submission_id,
        ls.study_plan_id,
        ls.learning_material_id,
        sh.shuffled_quiz_set_id,
        sh.student_text_answer,
        sh.correct_text_answer,
        sh.student_index_answer,
        sh.correct_index_answer,
        sh.submitted_keys_answer,
        sh.correct_keys_answer,
        sh.point,
        sh.is_correct,
        sh.is_accepted,
        sh.created_at,
        sh.updated_at,
        sh.deleted_at,
        sh.resource_path
    FROM get_submission_history() AS sh
    JOIN lo_submission ls USING(shuffled_quiz_set_id)
    JOIN quizzes q ON q.external_id = sh.quiz_id
    WHERE ls.deleted_at IS NULL
        AND q.deleted_at IS NULL
        AND sh.shuffled_quiz_set_id = NEW.shuffled_quiz_set_id
    ON CONFLICT ON CONSTRAINT lo_submission_answer_pk DO UPDATE SET
        student_text_answer = EXCLUDED.student_text_answer,
        correct_text_answer = EXCLUDED.correct_text_answer,
        student_index_answer = EXCLUDED.student_index_answer,
        correct_index_answer = EXCLUDED.correct_index_answer,
        submitted_keys_answer = EXCLUDED.submitted_keys_answer,
        correct_keys_answer = EXCLUDED.correct_keys_answer,
        point = EXCLUDED.point,
        is_correct = EXCLUDED.is_correct,
        is_accepted = EXCLUDED.is_accepted,
        updated_at = EXCLUDED.updated_at,
        deleted_at = EXCLUDED.deleted_at;
RETURN NULL;
END;
$$;


ALTER FUNCTION public.migrate_to_lo_submission_answer_fn() OWNER TO postgres;

--
-- Name: migrate_to_lo_submission_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.migrate_to_lo_submission_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF EXISTS (
        SELECT 1 
        FROM public.learning_objective LO 
        WHERE LO.learning_material_id = NEW.learning_material_id
    )
    THEN
        INSERT INTO public.lo_submission (
            submission_id,
            student_id,
            study_plan_id,
            learning_material_id,
            shuffled_quiz_set_id,
            total_point,
            created_at,
            updated_at,
            deleted_at,
            resource_path
        )
        VALUES (
            generate_ulid(),
            NEW.student_id,
            NEW.study_plan_id,
            NEW.learning_material_id,
            NEW.shuffled_quiz_set_id,
            COALESCE((SELECT SUM(point) FROM public.quizzes WHERE quizzes.deleted_at IS NULL AND quizzes.external_id = ANY(NEW.quiz_external_ids)), 0),
            NEW.created_at,
            NEW.updated_at,
            NEW.deleted_at,
            NEW.resource_path
        )
        ON CONFLICT ON CONSTRAINT shuffled_quiz_set_id_lo_submission_un DO UPDATE SET
            updated_at = EXCLUDED.updated_at,
            deleted_at = EXCLUDED.deleted_at,
            total_point = EXCLUDED.total_point;
    END IF;
RETURN NULL;
END;
$$;


ALTER FUNCTION public.migrate_to_lo_submission_fn() OWNER TO postgres;

--
-- Name: migrate_withus_mapping_course_id(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.migrate_withus_mapping_course_id() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    INSERT INTO public.withus_mapping_course_id (
        manabie_course_id,
        resource_path
    )
    SELECT 
        NEW.course_id,
        NEW.resource_path
    ON CONFLICT ON CONSTRAINT withus_mapping_course_id_pk DO NOTHING;
RETURN NULL;
END;
$$;


ALTER FUNCTION public.migrate_withus_mapping_course_id() OWNER TO postgres;

--
-- Name: migrate_withus_mapping_exam_lo_id(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.migrate_withus_mapping_exam_lo_id() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    INSERT INTO public.withus_mapping_exam_lo_id (
        exam_lo_id,
        resource_path
    )
    SELECT 
        NEW.learning_material_id,
        NEW.resource_path
    ON CONFLICT ON CONSTRAINT withus_mapping_exam_lo_id_pk DO NOTHING;
RETURN NULL;
END;
$$;


ALTER FUNCTION public.migrate_withus_mapping_exam_lo_id() OWNER TO postgres;

--
-- Name: migrate_withus_mapping_question_tag(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.migrate_withus_mapping_question_tag() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    INSERT INTO public.withus_mapping_question_tag (
        manabie_tag_id,
        manabie_tag_name,
        resource_path
    )
    SELECT 
        NEW.question_tag_id,
        NEW.name,
        NEW.resource_path
    ON CONFLICT ON CONSTRAINT withus_mapping_question_tag_pk DO UPDATE SET
        manabie_tag_name = NEW.name;
RETURN NULL;
END;
$$;


ALTER FUNCTION public.migrate_withus_mapping_question_tag() OWNER TO postgres;

--
-- Name: permission_check(text, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.permission_check(resource_path text, table_name text) RETURNS boolean
    LANGUAGE sql STABLE
    AS $_$
    select ($1 = current_setting('permission.resource_path') )::BOOLEAN
$_$;


ALTER FUNCTION public.permission_check(resource_path text, table_name text) OWNER TO postgres;

--
-- Name: private_search_name_exam_lo_fn(text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.private_search_name_exam_lo_fn(search_name text) RETURNS SETOF public.exam_lo
    LANGUAGE sql STABLE SECURITY DEFINER
    AS $$
    SELECT
        *
    FROM
        public.exam_lo
    WHERE
        name ilike '%' || search_name || '%'
$$;


ALTER FUNCTION public.private_search_name_exam_lo_fn(search_name text) OWNER TO postgres;

--
-- Name: private_search_name_lm_fn(text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.private_search_name_lm_fn(search_name text) RETURNS SETOF public.learning_material
    LANGUAGE sql STABLE SECURITY DEFINER
    AS $$
SELECT
    *
FROM
    public.learning_material
WHERE
    name ilike search_name OR search_name IS NULL 
$$;


ALTER FUNCTION public.private_search_name_lm_fn(search_name text) OWNER TO postgres;

--
-- Name: private_search_name_user_fn(text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.private_search_name_user_fn(search_name text) RETURNS SETOF public.users
    LANGUAGE sql STABLE SECURITY DEFINER
    AS $$
    SELECT
        *
    FROM
        public.users
    WHERE
        name ilike '%' || search_name || '%'
$$;


ALTER FUNCTION public.private_search_name_user_fn(search_name text) OWNER TO postgres;

--
-- Name: retrieve_study_plan_identity(text[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.retrieve_study_plan_identity(study_plan_item_ids text[]) RETURNS TABLE(study_plan_id text, lm_id text, student_id text, study_plan_item_id text)
    LANGUAGE sql STABLE
    AS $$

select coalesce(ssp.master_study_plan_id, spi.study_plan_id) as study_plan_id,
       coalesce(nullif(content_structure ->>'lo_id', ''), content_structure->>'assignment_id') as lm_id,
       ssp.student_id,
       spi.study_plan_item_id
from study_plan_items spi
join student_study_plans ssp on spi.study_plan_id = ssp.study_plan_id
where study_plan_item_id = ANY(study_plan_item_ids)

$$;


ALTER FUNCTION public.retrieve_study_plan_identity(study_plan_item_ids text[]) OWNER TO postgres;

--
-- Name: study_plan_tree_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.study_plan_tree_fn() RETURNS TABLE(study_plan_id text, book_id text, chapter_id text, chapter_display_order smallint, topic_id text, topic_display_order smallint, learning_material_id text, lm_display_order smallint, resource_path text)
    LANGUAGE sql STABLE
    AS $$
SELECT sp.study_plan_id , bt.book_id , bt.chapter_id  , bt.chapter_display_order , bt.topic_id, bt.topic_display_order, bt.learning_material_id, bt.lm_display_order, sp.resource_path
FROM study_plans sp JOIN book_tree bt USING (book_id);
$$;


ALTER FUNCTION public.study_plan_tree_fn() OWNER TO postgres;

--
-- Name: task_assignment_graded_score_v2(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.task_assignment_graded_score_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, student_submission_id text, graded_point smallint, total_point smallint, status text, passed boolean, created_at timestamp with time zone)
    LANGUAGE sql STABLE
    AS $$
select
    ss.student_id,
    ss.study_plan_id,
    ss.learning_material_id,
    ss.student_submission_id,
    ss.correct_score::smallint as graded_point,
    ss.total_score::smallint as total_point,
    ss.status,
    ss.understanding_level != 'SUBMISSION_UNDERSTANDING_LEVEL_SAD' as passed,
    ss.created_at
from student_submissions ss
    join task_assignment ta using (learning_material_id)
where ss.correct_score > 0 and ss.deleted_at is null;
$$;


ALTER FUNCTION public.task_assignment_graded_score_v2() OWNER TO postgres;

--
-- Name: trigger_student_event_logs_fill_new_identity_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.trigger_student_event_logs_fill_new_identity_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE current_study_plan_id text;
        current_learning_material_id text;
BEGIN 
IF new.event_type = ANY(ARRAY[
    'study_guide_finished',
    'video_finished',
    'learning_objective',
    'quiz_answer_selected'
    ]) THEN
    current_study_plan_id =(
        SELECT
        COALESCE(sp.master_study_plan_id, sp.study_plan_id)
        FROM
        study_plan_items spi
        JOIN study_plans sp ON spi.study_plan_id = sp.study_plan_id
        WHERE
        spi.study_plan_item_id = new.payload ->> 'study_plan_item_id'
    );
    current_learning_material_id = new.payload->>'lo_id';
UPDATE
    student_event_logs sel
SET
    study_plan_item_id = (new.payload ->> 'study_plan_item_id'),
    learning_material_id = current_learning_material_id,
    study_plan_id = current_study_plan_id
WHERE
    student_event_log_id = new.student_event_log_id;
END IF;
IF (new.event_type = 'learning_objective' and NEW.payload ->> 'event' = 'completed')
THEN
    IF exists (select 1 from flash_card where learning_material_id = current_learning_material_id) THEN
        update flash_card_submission set is_submitted = true, updated_at = now()    
        where student_id = new.student_id 
            and learning_material_id = current_learning_material_id
            and study_plan_id = current_study_plan_id 
            and shuffled_quiz_set_id = (select shuffled_quiz_set_id from shuffled_quiz_sets where session_id = new.payload ->> 'session_id' ORDER BY created_at desc limit 1);
    END IF;
    
    if exists (select 1 from learning_objective where learning_material_id = current_learning_material_id) THEN
        update lo_submission set is_submitted = true, updated_at = now()
        where student_id = new.student_id 
            and learning_material_id = current_learning_material_id
            and study_plan_id = current_study_plan_id 
            and shuffled_quiz_set_id = (select shuffled_quiz_set_id from shuffled_quiz_sets where session_id = new.payload ->> 'session_id' ORDER BY created_at desc limit 1);
    END IF;

	call upsert_highest_score(current_study_plan_id, current_learning_material_id, new.student_id, new.resource_path);
END IF;
RETURN NULL;
END;
$$;


ALTER FUNCTION public.trigger_student_event_logs_fill_new_identity_fn() OWNER TO postgres;

--
-- Name: trigger_student_latest_submissions_fill_new_identity_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.trigger_student_latest_submissions_fill_new_identity_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE sp_id TEXT;
BEGIN
  IF NEW.study_plan_item_id IS NULL THEN
    RETURN NEW;
  END IF;

  SELECT nid.study_plan_id INTO sp_id FROM public.retrieve_study_plan_identity(ARRAY[NEW.study_plan_item_id]) as nid;
  NEW.study_plan_id = sp_id;
  NEW.learning_material_id = NEW.assignment_id;

  RETURN NEW;
END;
$$;


ALTER FUNCTION public.trigger_student_latest_submissions_fill_new_identity_fn() OWNER TO postgres;

--
-- Name: trigger_student_submissions_fill_new_identity_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.trigger_student_submissions_fill_new_identity_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE current_study_plan_id text;
        current_learning_material_id text;
        current_student_id text;
        current_resource_path text;
        tmp record;
BEGIN
  current_study_plan_id = (
  SELECT
    COALESCE(sp.master_study_plan_id, sp.study_plan_id)
  FROM
    study_plan_items spi
  JOIN study_plans sp ON spi.study_plan_id = sp.study_plan_id
  WHERE
    spi.study_plan_item_id = new.study_plan_item_id
  );
  current_learning_material_id = new.assignment_id;
  current_student_id = new.student_id;
  current_resource_path = new.resource_path;
  IF TG_OP = 'INSERT' THEN 
    UPDATE
    student_submissions ss
    SET
    study_plan_id = current_study_plan_id,
    learning_material_id = current_learning_material_id
    WHERE
    study_plan_item_id = NEW.study_plan_item_id;
  ELSE 
     -- if TG_OP != 'INSERT', in some cases we don't have new.study_plan_item_id so we need to use student_submission_id for query to get identity  
    IF (current_study_plan_id IS NULL) AND (new.student_submission_id is not null) THEN
        SELECT
        ss.study_plan_id,
        ss.learning_material_id,
        ss.student_id,
        ss.resource_path
        FROM
        student_submissions ss
        WHERE
        ss.student_submission_id = new.student_submission_id
        INTO tmp;
    current_study_plan_id = tmp.study_plan_id;
    current_learning_material_id = tmp.learning_material_id;
    current_student_id = tmp.student_id;
    current_resource_path = tmp.resource_path;
    END IF;
  END IF ;
  
	call upsert_highest_score(current_study_plan_id, current_learning_material_id, current_student_id, current_resource_path);
RETURN NULL;
END;
$$;


ALTER FUNCTION public.trigger_student_submissions_fill_new_identity_fn() OWNER TO postgres;

--
-- Name: trigger_study_plan_items_completed_at_to_exam_lo_submission_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.trigger_study_plan_items_completed_at_to_exam_lo_submission_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.completed_at IS NOT NULL
    THEN
        INSERT INTO public.exam_lo_submission (
            submission_id,
            student_id,
            study_plan_id,
            learning_material_id,
            shuffled_quiz_set_id,
            status,
            result,
            created_at,
            updated_at,
            deleted_at,
            total_point,
            resource_path
        )
        SELECT generate_ulid() AS submission_id,
               SQ.student_id AS student_id,
               (SELECT SP.master_study_plan_id
                  FROM study_plans SP
                      LEFT JOIN study_plan_items SPI
                          ON SPI.study_plan_id = SP.study_plan_id
                 WHERE SPI.study_plan_item_id = SQ.study_plan_item_id) AS study_plan_id,
               (SELECT content_structure->>'lo_id' FROM public.study_plan_items WHERE study_plan_item_id = SQ.study_plan_item_id) AS learning_material_id,
               SQ.shuffled_quiz_set_id AS shuffled_quiz_set_id,
               'SUBMISSION_STATUS_RETURNED' AS status,
               'EXAM_LO_SUBMISSION_COMPLETED' AS result,
               SQ.created_at AS created_at,
               SQ.updated_at AS updated_at,
               SQ.deleted_at AS deleted_at,
               COALESCE((SELECT SUM(point) FROM public.quizzes WHERE external_id = ANY(SQ.quiz_external_ids)), 0) AS total_point,
               SQ.resource_path AS resource_path
          FROM public.shuffled_quiz_sets SQ
         WHERE SQ.original_shuffle_quiz_set_id IS NULL
           AND EXISTS (SELECT 1
                         FROM public.study_plan_items SP
                        WHERE SP.study_plan_item_id = NEW.study_plan_item_id
                          AND SP.study_plan_item_id = SQ.study_plan_item_id
                          AND SP.completed_at IS NOT NULL
                          AND EXISTS (SELECT 1 FROM public.exam_lo WHERE learning_material_id = SP.content_structure->>'lo_id'))
                          -- When study_plan_items.completed_at was updated to IS NOT NULL, new data which match shuffled_quiz_sets should be migrated
                          AND NOT EXISTS (SELECT 1 FROM exam_lo_submission WHERE shuffled_quiz_set_id = SQ.shuffled_quiz_set_id)
        ON CONFLICT ON CONSTRAINT shuffled_quiz_set_id_un DO NOTHING;
    END IF;
RETURN NULL;
END;
$$;


ALTER FUNCTION public.trigger_study_plan_items_completed_at_to_exam_lo_submission_fn() OWNER TO postgres;

--
-- Name: trigger_study_plan_items_to_individual_study_plan(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.trigger_study_plan_items_to_individual_study_plan() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Two conditions to specify storing individual study plan items
-- 1 - This individual item should differ master item when insert
-- 2 - This individual item is task assigment (without  master) when insert
    WITH temp_table as (
        SELECT nt.*, spi.student_id, COALESCE(spi.master_study_plan_id, spi.study_plan_id) as r_study_plan_id
        FROM study_plans sp
         JOIN new_table nt
              ON nt.study_plan_id = sp.study_plan_id
         JOIN student_study_plans spi ON nt.study_plan_id = spi.study_plan_id
         LEFT JOIN study_plan_items master_spi
              ON master_spi.study_plan_item_id = nt.copy_study_plan_item_id
        WHERE   
                -- 1 - This individual item should differ master item
                (
                    sp.master_study_plan_id is not null
                    -- this is should be differ master
                    AND nt.created_at <> nt.updated_at
                )
                -- 2 - This individual item is task assigment (without master)
                OR (
                        sp.study_plan_type = 'STUDY_PLAN_TYPE_INDIVIDUAL'
                    )
    )
    INSERT INTO public.individual_study_plan (
        study_plan_id,
        learning_material_id,
        student_id,
        status,
        start_date,
        end_date,
        available_from,
        available_to,
        created_at,
        updated_at,
        deleted_at,
        school_date,
        resource_path
    )
    SELECT
        r_study_plan_id,
        COALESCE(NULLIF(content_structure->>'lo_id', ''), content_structure->>'assignment_id'),
        student_id,
        status,
        start_date,
        end_date,
        available_from,
        available_to,
        created_at,
        updated_at,
        deleted_at,
        school_date,
        resource_path
    FROM temp_table

    ON CONFLICT ON CONSTRAINT learning_material_id_student_id_study_plan_id_pk DO UPDATE SET
     status = EXCLUDED.status,
     start_date = EXCLUDED.start_date,
     end_date = EXCLUDED.end_date,
     available_from = EXCLUDED.available_from,
     available_to = EXCLUDED.available_to,
     school_date = EXCLUDED.school_date,
     updated_at = EXCLUDED.updated_at,
     deleted_at = EXCLUDED.deleted_at;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.trigger_study_plan_items_to_individual_study_plan() OWNER TO postgres;

--
-- Name: update_allocate_marker_when_exam_lo_submission_was_created(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_allocate_marker_when_exam_lo_submission_was_created() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    UPDATE allocate_marker
    SET teacher_id = NULL
    WHERE study_plan_id = NEW.study_plan_id
    AND student_id = NEW.student_id 
    AND learning_material_id = NEW.learning_material_id;

    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_allocate_marker_when_exam_lo_submission_was_created() OWNER TO postgres;

--
-- Name: update_book_id_for_chapters_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_book_id_for_chapters_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ 
BEGIN 
-- IF new.book_id != old.book_id THEN
    UPDATE public.chapters 
    SET book_id = new.book_id
    WHERE chapter_id = new.chapter_id;
    -- END IF;
RETURN NULL;
END;
$$;


ALTER FUNCTION public.update_book_id_for_chapters_fn() OWNER TO postgres;

--
-- Name: update_content_structure_flatten_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_content_structure_flatten_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
  DECLARE
    lo_id text := NEW.lo_id;
    content_structure jsonb := (SELECT content_structure FROM study_plan_items WHERE study_plan_item_id = NEW.study_plan_item_id);
    course_id text := content_structure->>'course_id';
    book_id text := content_structure->>'book_id';
    chapter_id text := content_structure->>'chapter_id';
    topic_id text := content_structure->>'topic_id';
  BEGIN
    UPDATE study_plan_items
    SET content_structure_flatten = 'book::' || book_id || 'topic::' || topic_id || 'chapter::' || chapter_id || 'course::' || course_id || 'lo::' || lo_id
    WHERE study_plan_item_id = NEW.study_plan_item_id;
    RETURN NULL;
  END;
$$;


ALTER FUNCTION public.update_content_structure_flatten_fn() OWNER TO postgres;

--
-- Name: update_content_structure_flatten_on_assignment_study_plan_item_(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_content_structure_flatten_on_assignment_study_plan_item_() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
  DECLARE
    assignment_id text := NEW.assignment_id;
    content_structure jsonb := (SELECT content_structure FROM study_plan_items WHERE study_plan_item_id = NEW.study_plan_item_id);
    course_id text := content_structure->>'course_id';
    book_id text := content_structure->>'book_id';
    chapter_id text := content_structure->>'chapter_id';
    topic_id text := content_structure->>'topic_id';
  BEGIN
    UPDATE study_plan_items
    SET content_structure_flatten = 'book::' || book_id || 'topic::' || topic_id || 'chapter::' || chapter_id || 'course::' || course_id || 'assignment::' || assignment_id
    WHERE study_plan_item_id = NEW.study_plan_item_id;
    RETURN NULL;
  END;
$$;


ALTER FUNCTION public.update_content_structure_flatten_on_assignment_study_plan_item_() OWNER TO postgres;

--
-- Name: update_content_structure_fnc(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_content_structure_fnc() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
    if new.content_structure_flatten is not null then 
        -- add assignment_id into content_structure when content_structure_flatten contains 'assignment::'
        if new.content_structure_flatten like '%assignment::%' then
            new.content_structure = new.content_structure || concat('{"assignment_id": "', split_part(new.content_structure_flatten, '::', 6), '"}')::jsonb;
        -- add lo_id into content_structure when content_structure_flatten contains 'lo::'
        elsif new.content_structure_flatten like '%lo::%' then 
            new.content_structure = new.content_structure || concat('{"lo_id": "', split_part(new.content_structure_flatten, '::', 6), '"}')::jsonb;
        end if;
    end if;
RETURN NEW;

END;

$$;


ALTER FUNCTION public.update_content_structure_fnc() OWNER TO postgres;

--
-- Name: update_master_study_plan_id_on_student_study_plan_created_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_master_study_plan_id_on_student_study_plan_created_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
  BEGIN
    UPDATE student_study_plans
    SET master_study_plan_id = (SELECT master_study_plan_id FROM study_plans WHERE study_plan_id=NEW.study_plan_id)
    WHERE study_plan_id = NEW.study_plan_id;
    RETURN NULL;
  END;
$$;


ALTER FUNCTION public.update_master_study_plan_id_on_student_study_plan_created_fn() OWNER TO postgres;

--
-- Name: update_max_score_exam_lo_once_exam_lo_submission_status_change(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_max_score_exam_lo_once_exam_lo_submission_status_change() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
		WITH mss AS (
	        SELECT
				coalesce(mgs.graded_point,0) AS graded_point,
				coalesce(mgs.total_point,0) AS total_point,
				coalesce((mgs.graded_point * 1.0 / mgs.total_point) * 100, 0)::smallint AS max_percentage
	        FROM
	            max_graded_score_v2() AS mgs
	        WHERE
	            mgs.learning_material_id = NEW.learning_material_id AND 
	            mgs.study_plan_id = NEW.study_plan_id AND 
	            mgs.student_id = NEW.student_id),
	            update_when_mss_empty AS (
               	UPDATE max_score_submission
			    SET updated_at = now(), max_score = NULL, max_percentage = NULL
			    WHERE NOT EXISTS (SELECT 1 FROM mss) AND learning_material_id = NEW.learning_material_id AND study_plan_id = NEW.study_plan_id AND student_id = NEW.student_id)
	           
		INSERT
		INTO
		max_score_submission AS target (
		student_id,
		study_plan_id,
		learning_material_id,
		max_score,
		total_score,
	    max_percentage,
		created_at,
		updated_at,
		deleted_at,
		resource_path)
		SELECT
		NEW.student_id,
		NEW.study_plan_id,
		NEW.learning_material_id,
		mss.graded_point,
		mss.total_point,
		mss.max_percentage,
		now(),
		now(),
		NULL,
		NEW.resource_path
	FROM
		mss
	ON CONFLICT ON CONSTRAINT max_score_submission_study_plan_item_identity_pk DO
	UPDATE
	SET
		max_score = EXCLUDED.max_score,
	    total_score = EXCLUDED.total_score,
	    max_percentage = EXCLUDED.max_percentage,
	    updated_at = now();
RETURN NULL;
END;
$$;


ALTER FUNCTION public.update_max_score_exam_lo_once_exam_lo_submission_status_change() OWNER TO postgres;

--
-- Name: update_question_hierarchy_quiz_sets_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_question_hierarchy_quiz_sets_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ 
BEGIN 
    UPDATE public.quiz_sets
    SET question_hierarchy = (
        CASE WHEN ARRAY_LENGTH(quiz_external_ids, 1) IS NULL
        THEN ARRAY[]::JSONB[]
        ELSE
        (
            SELECT ARRAY_AGG(
                TO_JSONB(qei)
            ) FROM (
                SELECT UNNEST(quiz_external_ids) as id, 'QUESTION' as type
            ) qei
        )
        END
    )
    WHERE quiz_set_id=NEW.quiz_set_id
    AND NEW.question_hierarchy IS NULL
    AND NOT EXISTS(
        SELECT 1 FROM flash_card WHERE learning_material_id=NEW.lo_id AND deleted_at IS NULL
    );
RETURN NULL;
END;
$$;


ALTER FUNCTION public.update_question_hierarchy_quiz_sets_fn() OWNER TO postgres;

--
-- Name: update_question_hierarchy_shuffled_quiz_sets_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_question_hierarchy_shuffled_quiz_sets_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
UPDATE public.shuffled_quiz_sets
SET question_hierarchy = (
    CASE WHEN ARRAY_LENGTH(quiz_external_ids, 1) IS NULL
             THEN ARRAY[]::JSONB[]
         ELSE
             (
                 SELECT ARRAY_AGG(
                                TO_JSONB(qei)
                            ) FROM (
                                       SELECT UNNEST(quiz_external_ids) as id, 'QUESTION' as type
                                   ) qei
             )
        END
    )
WHERE shuffled_quiz_set_id=NEW.shuffled_quiz_set_id
  AND NEW.question_hierarchy IS NULL
  AND NOT EXISTS(
        SELECT 1 FROM flash_card WHERE learning_material_id=NEW.learning_material_id AND deleted_at IS NULL
    );
RETURN NULL;
END;
$$;


ALTER FUNCTION public.update_question_hierarchy_shuffled_quiz_sets_fn() OWNER TO postgres;

--
-- Name: update_study_plan_item_identity_for_flashcard_progression_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_study_plan_item_identity_for_flashcard_progression_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ 
BEGIN 
    UPDATE public.flashcard_progressions sqs
    SET study_plan_id =  COALESCE(sp.master_study_plan_id, sp.study_plan_id),
        learning_material_id = CASE
            WHEN content_structure ->> 'lo_id' != ANY(ARRAY['', NULL]) THEN content_structure ->> 'lo_id'
            WHEN content_structure ->> 'assignment_id' != ANY(ARRAY['', NULL]) THEN content_structure ->> 'assignment_id'
            ELSE NULL 
            END 
    FROM public.study_plan_items spi
    JOIN public.study_plans sp
    USING(study_plan_id)
    WHERE sqs.study_plan_item_id = new.study_plan_item_id AND sqs.study_plan_item_id = spi.study_plan_item_id;
RETURN NULL;
END;
$$;


ALTER FUNCTION public.update_study_plan_item_identity_for_flashcard_progression_fn() OWNER TO postgres;

--
-- Name: update_study_plan_item_identity_for_shuffled_quiz_set_fn(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_study_plan_item_identity_for_shuffled_quiz_set_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ 
BEGIN 
    UPDATE public.shuffled_quiz_sets sqs
    SET study_plan_id =  COALESCE(sp.master_study_plan_id, sp.study_plan_id),
        learning_material_id = CASE
            WHEN content_structure ->> 'lo_id' != ANY(ARRAY['', NULL]) THEN content_structure ->> 'lo_id'
            WHEN content_structure ->> 'assignment_id' != ANY(ARRAY['', NULL]) THEN content_structure ->> 'assignment_id'
            ELSE NULL 
            END 
    FROM public.study_plan_items spi
    JOIN public.study_plans sp
    USING(study_plan_id)
    WHERE sqs.study_plan_item_id = new.study_plan_item_id AND sqs.study_plan_item_id = spi.study_plan_item_id;
RETURN NULL;
END;
$$;


ALTER FUNCTION public.update_study_plan_item_identity_for_shuffled_quiz_set_fn() OWNER TO postgres;

--
-- Name: upsert_highest_score(text, text, text, text); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.upsert_highest_score(current_study_plan_id text, current_learning_material_id text, current_student_id text, current_resource_path text)
    LANGUAGE plpgsql
    AS $$
declare tmp record;
BEGIN 
	select  
		coalesce(graded_point,0) as max_score,
        coalesce(total_point,0) as total_score, 
        coalesce((graded_point * 1.0 / total_point) * 100, 0)::smallint max_percentage 
	from max_graded_score_v2()
	where study_plan_id = current_study_plan_id 
	and learning_material_id = current_learning_material_id
	and student_id = current_student_id
	into tmp;
	insert into max_score_submission (study_plan_id, learning_material_id, student_id, max_score, total_score, max_percentage, created_at, updated_at, deleted_at, resource_path)
	values(current_study_plan_id, current_learning_material_id, current_student_id,tmp.max_score, tmp.total_score, tmp.max_percentage, now(), now(), null, current_resource_path) 
	ON CONFLICT ON constraint max_score_submission_study_plan_item_identity_pk 
	do update set max_score = tmp.max_score,
				  total_score = tmp.total_score,
                  max_percentage = tmp.max_percentage,
				  updated_at = now();
end; 
$$;


ALTER PROCEDURE public.upsert_highest_score(current_study_plan_id text, current_learning_material_id text, current_student_id text, current_resource_path text) OWNER TO postgres;

--
-- Name: withus_check_valid_course_id(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.withus_check_valid_course_id() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.manabie_course_id IS NOT NULL THEN
        IF NOT EXISTS (SELECT 1 FROM public.course_students WHERE course_id = NEW.manabie_course_id) THEN
            RAISE EXCEPTION 'manabie_course_id % does not exist', NEW.manabie_course_id;
        END IF;
    END IF;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.withus_check_valid_course_id() OWNER TO postgres;

--
-- Name: academic_year; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.academic_year (
    academic_year_id text NOT NULL,
    name text NOT NULL,
    start_date date NOT NULL,
    end_date date NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.academic_year FORCE ROW LEVEL SECURITY;


ALTER TABLE public.academic_year OWNER TO postgres;

--
-- Name: allocate_marker; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.allocate_marker (
    allocate_marker_id text NOT NULL,
    teacher_id text,
    student_id text NOT NULL,
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    created_by text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath()
);


ALTER TABLE public.allocate_marker OWNER TO postgres;

--
-- Name: alloydb_dbz_signal; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.alloydb_dbz_signal (
    id text NOT NULL,
    type text,
    data text
);


ALTER TABLE public.alloydb_dbz_signal OWNER TO postgres;

--
-- Name: assessment; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.assessment (
    id text NOT NULL,
    course_id text NOT NULL,
    learning_material_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    ref_table character varying(20) NOT NULL
);

ALTER TABLE ONLY public.assessment FORCE ROW LEVEL SECURITY;


ALTER TABLE public.assessment OWNER TO postgres;

--
-- Name: assessment_session; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.assessment_session (
    session_id text NOT NULL,
    user_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    assessment_id text,
    status text,
    study_plan_assessment_id text,
    CONSTRAINT check_one_of_fk_assessment_not_null CHECK (((
CASE
    WHEN (assessment_id IS NULL) THEN 0
    ELSE 1
END +
CASE
    WHEN (study_plan_assessment_id IS NULL) THEN 0
    ELSE 1
END) = 1))
);

ALTER TABLE ONLY public.assessment_session FORCE ROW LEVEL SECURITY;


ALTER TABLE public.assessment_session OWNER TO postgres;

--
-- Name: assessment_submission; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.assessment_submission (
    id text NOT NULL,
    session_id text NOT NULL,
    assessment_id text,
    student_id text NOT NULL,
    grading_status text NOT NULL,
    max_score integer DEFAULT 0 NOT NULL,
    graded_score integer DEFAULT 0 NOT NULL,
    allocated_marker_id text,
    marked_by text,
    marked_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    completed_at timestamp with time zone NOT NULL,
    study_plan_assessment_id text,
    CONSTRAINT check_one_of_fk_assessment_not_null CHECK (((
CASE
    WHEN (assessment_id IS NULL) THEN 0
    ELSE 1
END +
CASE
    WHEN (study_plan_assessment_id IS NULL) THEN 0
    ELSE 1
END) = 1)),
    CONSTRAINT grading_status_check CHECK ((grading_status = ANY (ARRAY['NOT_MARKED'::text, 'IN_PROGRESS'::text, 'MARKED'::text, 'RETURNED'::text])))
);

ALTER TABLE ONLY public.assessment_submission FORCE ROW LEVEL SECURITY;


ALTER TABLE public.assessment_submission OWNER TO postgres;

--
-- Name: assign_study_plan_tasks; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.assign_study_plan_tasks (
    id text NOT NULL,
    study_plan_ids text[] NOT NULL,
    status text,
    course_id text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    error_detail text,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.assign_study_plan_tasks FORCE ROW LEVEL SECURITY;


ALTER TABLE public.assign_study_plan_tasks OWNER TO postgres;

--
-- Name: assignment; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.assignment (
    attachments text[],
    max_grade integer,
    instruction text,
    is_required_grade boolean,
    allow_resubmission boolean,
    require_attachment boolean,
    allow_late_submission boolean,
    require_assignment_note boolean,
    require_video_submission boolean,
    CONSTRAINT assignment_type_check CHECK ((type = 'LEARNING_MATERIAL_GENERAL_ASSIGNMENT'::text))
)
INHERITS (public.learning_material);

ALTER TABLE ONLY public.assignment FORCE ROW LEVEL SECURITY;


ALTER TABLE public.assignment OWNER TO postgres;

--
-- Name: assignment_study_plan_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.assignment_study_plan_items (
    assignment_id text NOT NULL,
    study_plan_item_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.assignment_study_plan_items FORCE ROW LEVEL SECURITY;


ALTER TABLE public.assignment_study_plan_items OWNER TO postgres;

--
-- Name: book_tree; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.book_tree AS
 SELECT book_tree_fn.book_id,
    book_tree_fn.chapter_id,
    book_tree_fn.chapter_display_order,
    book_tree_fn.topic_id,
    book_tree_fn.topic_display_order,
    book_tree_fn.learning_material_id,
    book_tree_fn.lm_display_order
   FROM public.book_tree_fn() book_tree_fn(book_id, chapter_id, chapter_display_order, topic_id, topic_display_order, learning_material_id, lm_display_order);


ALTER TABLE public.book_tree OWNER TO postgres;

--
-- Name: books; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.books (
    book_id text NOT NULL,
    name text NOT NULL,
    country text,
    subject text,
    grade smallint,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    school_id integer DEFAULT '-2147483648'::integer,
    copied_from text,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    current_chapter_display_order integer DEFAULT 0 NOT NULL,
    book_type text DEFAULT 'BOOK_TYPE_GENERAL'::text,
    is_v2 boolean DEFAULT false,
    CONSTRAINT book_type_check CHECK ((book_type = ANY (ARRAY['BOOK_TYPE_NONE'::text, 'BOOK_TYPE_GENERAL'::text, 'BOOK_TYPE_ADHOC'::text])))
);

ALTER TABLE ONLY public.books FORCE ROW LEVEL SECURITY;


ALTER TABLE public.books OWNER TO postgres;

--
-- Name: books_chapters; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.books_chapters (
    book_id text NOT NULL,
    chapter_id text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.books_chapters FORCE ROW LEVEL SECURITY;


ALTER TABLE public.books_chapters OWNER TO postgres;

--
-- Name: cerebry_classes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cerebry_classes (
    id text NOT NULL,
    class_code text NOT NULL,
    class_name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath()
);

ALTER TABLE ONLY public.cerebry_classes FORCE ROW LEVEL SECURITY;


ALTER TABLE public.cerebry_classes OWNER TO postgres;

--
-- Name: chapters; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.chapters (
    chapter_id text NOT NULL,
    name text NOT NULL,
    country text,
    subject text,
    grade smallint,
    display_order smallint DEFAULT 0,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    school_id integer DEFAULT '-2147483648'::integer NOT NULL,
    deleted_at timestamp with time zone,
    copied_from text,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    current_topic_display_order integer DEFAULT 0,
    book_id text
);

ALTER TABLE ONLY public.chapters FORCE ROW LEVEL SECURITY;


ALTER TABLE public.chapters OWNER TO postgres;

--
-- Name: class_students; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.class_students (
    student_id text NOT NULL,
    class_id text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.class_students FORCE ROW LEVEL SECURITY;


ALTER TABLE public.class_students OWNER TO postgres;

--
-- Name: class_study_plans; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.class_study_plans (
    class_id integer NOT NULL,
    study_plan_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.class_study_plans FORCE ROW LEVEL SECURITY;


ALTER TABLE public.class_study_plans OWNER TO postgres;

--
-- Name: content_bank_medias; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.content_bank_medias (
    id text NOT NULL,
    name text NOT NULL,
    resource text,
    type text,
    file_size_bytes bigint DEFAULT 0,
    created_by text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    CONSTRAINT media_type_check CHECK ((type = ANY (ARRAY['MEDIA_TYPE_IMAGE'::text])))
);

ALTER TABLE ONLY public.content_bank_medias FORCE ROW LEVEL SECURITY;


ALTER TABLE public.content_bank_medias OWNER TO postgres;

--
-- Name: course_access_paths; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.course_access_paths (
    course_id text NOT NULL,
    location_id text NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);


ALTER TABLE public.course_access_paths OWNER TO postgres;

--
-- Name: course_classes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.course_classes (
    course_id text NOT NULL,
    class_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    course_class_id text NOT NULL,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.course_classes FORCE ROW LEVEL SECURITY;


ALTER TABLE public.course_classes OWNER TO postgres;

--
-- Name: course_student_subscriptions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.course_student_subscriptions (
    course_student_subscription_id text NOT NULL,
    course_student_id text NOT NULL,
    course_id text NOT NULL,
    student_id text NOT NULL,
    start_at timestamp with time zone,
    end_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.course_student_subscriptions FORCE ROW LEVEL SECURITY;


ALTER TABLE public.course_student_subscriptions OWNER TO postgres;

--
-- Name: course_students_access_paths; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.course_students_access_paths (
    course_student_id text NOT NULL,
    location_id text NOT NULL,
    course_id text NOT NULL,
    student_id text NOT NULL,
    access_path text,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.course_students_access_paths FORCE ROW LEVEL SECURITY;


ALTER TABLE public.course_students_access_paths OWNER TO postgres;

--
-- Name: courses; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.courses (
    course_id text NOT NULL,
    name text NOT NULL,
    country text,
    subject text,
    grade smallint,
    display_order smallint DEFAULT 0,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    school_id integer DEFAULT '-2147483648'::integer NOT NULL,
    deleted_at timestamp with time zone,
    course_type text,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    teacher_ids text[],
    preset_study_plan_id text,
    icon text,
    status text DEFAULT 'COURSE_STATUS_NONE'::text,
    resource_path text DEFAULT public.autofillresourcepath(),
    teaching_method text,
    course_type_id text,
    remarks text,
    is_archived boolean DEFAULT false,
    course_partner_id text,
    is_adaptive boolean DEFAULT false,
    vendor_id text
);

ALTER TABLE ONLY public.courses FORCE ROW LEVEL SECURITY;


ALTER TABLE public.courses OWNER TO postgres;

--
-- Name: courses_books; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.courses_books (
    book_id text NOT NULL,
    course_id text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.courses_books FORCE ROW LEVEL SECURITY;


ALTER TABLE public.courses_books OWNER TO postgres;

--
-- Name: dbz_signals; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.dbz_signals (
    id text NOT NULL,
    type text,
    data text
);


ALTER TABLE public.dbz_signals OWNER TO postgres;

--
-- Name: exam_lo_submission; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.exam_lo_submission (
    submission_id text NOT NULL,
    student_id text NOT NULL,
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    shuffled_quiz_set_id text NOT NULL,
    status text,
    result text,
    teacher_feedback text,
    teacher_id text,
    marked_at timestamp with time zone,
    removed_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    total_point integer DEFAULT 0,
    last_action text DEFAULT 'APPROVE_ACTION_NONE'::text NOT NULL,
    last_action_at timestamp with time zone,
    last_action_by text,
    CONSTRAINT last_action_check CHECK ((last_action = ANY (ARRAY['APPROVE_ACTION_NONE'::text, 'APPROVE_ACTION_APPROVED'::text, 'APPROVE_ACTION_REJECTED'::text])))
);

ALTER TABLE ONLY public.exam_lo_submission FORCE ROW LEVEL SECURITY;


ALTER TABLE public.exam_lo_submission OWNER TO postgres;

--
-- Name: exam_lo_submission_answer; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.exam_lo_submission_answer (
    student_id text NOT NULL,
    quiz_id text NOT NULL,
    submission_id text NOT NULL,
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    shuffled_quiz_set_id text NOT NULL,
    student_text_answer text[],
    correct_text_answer text[],
    student_index_answer integer[],
    correct_index_answer integer[],
    is_correct boolean[] DEFAULT '{}'::boolean[] NOT NULL,
    is_accepted boolean,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    point integer DEFAULT 0,
    submitted_keys_answer text[],
    correct_keys_answer text[]
);

ALTER TABLE ONLY public.exam_lo_submission_answer FORCE ROW LEVEL SECURITY;


ALTER TABLE public.exam_lo_submission_answer OWNER TO postgres;

--
-- Name: exam_lo_submission_score; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.exam_lo_submission_score (
    submission_id text NOT NULL,
    quiz_id text NOT NULL,
    teacher_id text NOT NULL,
    teacher_comment text,
    is_correct boolean[] DEFAULT '{}'::boolean[] NOT NULL,
    is_accepted boolean,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    point integer DEFAULT 0,
    shuffled_quiz_set_id text NOT NULL
);

ALTER TABLE ONLY public.exam_lo_submission_score FORCE ROW LEVEL SECURITY;


ALTER TABLE public.exam_lo_submission_score OWNER TO postgres;

--
-- Name: feedback_session; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.feedback_session (
    id text NOT NULL,
    submission_id text NOT NULL,
    created_by text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath()
);

ALTER TABLE ONLY public.feedback_session FORCE ROW LEVEL SECURITY;


ALTER TABLE public.feedback_session OWNER TO postgres;

--
-- Name: flash_card; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.flash_card (
    CONSTRAINT learning_objective_type_check CHECK ((type = 'LEARNING_MATERIAL_FLASH_CARD'::text))
)
INHERITS (public.learning_material);

ALTER TABLE ONLY public.flash_card FORCE ROW LEVEL SECURITY;


ALTER TABLE public.flash_card OWNER TO postgres;

--
-- Name: flash_card_submission; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.flash_card_submission (
    submission_id text NOT NULL,
    student_id text NOT NULL,
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    shuffled_quiz_set_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    total_point integer DEFAULT 0,
    is_submitted boolean DEFAULT false NOT NULL
);

ALTER TABLE ONLY public.flash_card_submission FORCE ROW LEVEL SECURITY;


ALTER TABLE public.flash_card_submission OWNER TO postgres;

--
-- Name: flash_card_submission_answer; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.flash_card_submission_answer (
    student_id text NOT NULL,
    quiz_id text NOT NULL,
    submission_id text NOT NULL,
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    shuffled_quiz_set_id text NOT NULL,
    student_text_answer text[],
    correct_text_answer text[],
    student_index_answer integer[],
    correct_index_answer integer[],
    is_correct boolean[] DEFAULT '{}'::boolean[] NOT NULL,
    is_accepted boolean DEFAULT false,
    point integer DEFAULT 0,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath()
);

ALTER TABLE ONLY public.flash_card_submission_answer FORCE ROW LEVEL SECURITY;


ALTER TABLE public.flash_card_submission_answer OWNER TO postgres;

--
-- Name: flashcard_progressions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.flashcard_progressions (
    study_set_id text NOT NULL,
    original_study_set_id text,
    student_id text NOT NULL,
    study_plan_item_id text,
    lo_id text,
    quiz_external_ids text[],
    studying_index integer,
    skipped_question_ids text[],
    remembered_question_ids text[],
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    completed_at timestamp with time zone,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    original_quiz_set_id text,
    study_plan_id text,
    learning_material_id text
);

ALTER TABLE ONLY public.flashcard_progressions FORCE ROW LEVEL SECURITY;


ALTER TABLE public.flashcard_progressions OWNER TO postgres;

--
-- Name: flashcard_speeches; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.flashcard_speeches (
    speech_id text NOT NULL,
    sentence text NOT NULL,
    link text NOT NULL,
    type text NOT NULL,
    quiz_id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by text,
    updated_by text,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    settings jsonb
);

ALTER TABLE ONLY public.flashcard_speeches FORCE ROW LEVEL SECURITY;


ALTER TABLE public.flashcard_speeches OWNER TO postgres;

--
-- Name: get_student_study_plans_by_filter_view; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.get_student_study_plans_by_filter_view AS
 SELECT get_student_study_plans_by_filter_v2.study_plan_id,
    get_student_study_plans_by_filter_v2.master_study_plan_id,
    get_student_study_plans_by_filter_v2.name,
    get_student_study_plans_by_filter_v2.study_plan_type,
    get_student_study_plans_by_filter_v2.school_id,
    get_student_study_plans_by_filter_v2.created_at,
    get_student_study_plans_by_filter_v2.updated_at,
    get_student_study_plans_by_filter_v2.deleted_at,
    get_student_study_plans_by_filter_v2.course_id,
    get_student_study_plans_by_filter_v2.resource_path,
    get_student_study_plans_by_filter_v2.book_id,
    get_student_study_plans_by_filter_v2.status,
    get_student_study_plans_by_filter_v2.track_school_progress,
    get_student_study_plans_by_filter_v2.grades,
    get_student_study_plans_by_filter_v2.student_id
   FROM public.get_student_study_plans_by_filter_v2() get_student_study_plans_by_filter_v2(study_plan_id, master_study_plan_id, name, study_plan_type, school_id, created_at, updated_at, deleted_at, course_id, resource_path, book_id, status, track_school_progress, grades, student_id);


ALTER TABLE public.get_student_study_plans_by_filter_view OWNER TO postgres;

--
-- Name: grade; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.grade (
    name text NOT NULL,
    is_archived boolean DEFAULT false NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    grade_id text NOT NULL,
    deleted_at timestamp with time zone,
    sequence integer
);

ALTER TABLE ONLY public.grade FORCE ROW LEVEL SECURITY;


ALTER TABLE public.grade OWNER TO postgres;

--
-- Name: grade_book_setting; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.grade_book_setting (
    setting text NOT NULL,
    updated_by text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    CONSTRAINT grade_book_setting_check CHECK ((setting = ANY (ARRAY['LATEST_SCORE'::text, 'GRADE_TO_PASS_SCORE'::text])))
);


ALTER TABLE public.grade_book_setting OWNER TO postgres;

--
-- Name: granted_permission; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.granted_permission (
    user_group_id text NOT NULL,
    user_group_name text NOT NULL,
    role_id text NOT NULL,
    role_name text NOT NULL,
    permission_id text NOT NULL,
    permission_name text NOT NULL,
    location_id text NOT NULL,
    resource_path text NOT NULL
);

ALTER TABLE ONLY public.granted_permission FORCE ROW LEVEL SECURITY;


ALTER TABLE public.granted_permission OWNER TO postgres;

--
-- Name: granted_role; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.granted_role (
    granted_role_id text NOT NULL,
    user_group_id text NOT NULL,
    role_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.granted_role FORCE ROW LEVEL SECURITY;


ALTER TABLE public.granted_role OWNER TO postgres;

--
-- Name: granted_role_access_path; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.granted_role_access_path (
    granted_role_id text NOT NULL,
    location_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.granted_role_access_path FORCE ROW LEVEL SECURITY;


ALTER TABLE public.granted_role_access_path OWNER TO postgres;

--
-- Name: locations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.locations (
    location_id text NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('UTC'::text, now()) NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    location_type text,
    partner_internal_id text,
    partner_internal_parent_id text,
    parent_location_id text,
    is_archived boolean DEFAULT false NOT NULL,
    access_path text
);

ALTER TABLE ONLY public.locations FORCE ROW LEVEL SECURITY;


ALTER TABLE public.locations OWNER TO postgres;

--
-- Name: permission; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.permission (
    permission_id text NOT NULL,
    permission_name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.permission FORCE ROW LEVEL SECURITY;


ALTER TABLE public.permission OWNER TO postgres;

--
-- Name: permission_role; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.permission_role (
    permission_id text NOT NULL,
    role_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.permission_role FORCE ROW LEVEL SECURITY;


ALTER TABLE public.permission_role OWNER TO postgres;

--
-- Name: role; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.role (
    role_id text NOT NULL,
    role_name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    is_system boolean DEFAULT false
);

ALTER TABLE ONLY public.role FORCE ROW LEVEL SECURITY;


ALTER TABLE public.role OWNER TO postgres;

--
-- Name: user_group; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_group (
    user_group_id text NOT NULL,
    user_group_name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    org_location_id text,
    is_system boolean DEFAULT false
);

ALTER TABLE ONLY public.user_group FORCE ROW LEVEL SECURITY;


ALTER TABLE public.user_group OWNER TO postgres;

--
-- Name: user_group_member; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_group_member (
    user_id text NOT NULL,
    user_group_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.user_group_member FORCE ROW LEVEL SECURITY;


ALTER TABLE public.user_group_member OWNER TO postgres;

--
-- Name: granted_permissions; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.granted_permissions AS
 SELECT ugm.user_id,
    p.permission_name,
    l1.location_id,
    ugm.resource_path,
    p.permission_id
   FROM ((((((((public.user_group_member ugm
     JOIN public.user_group ug ON ((ugm.user_group_id = ug.user_group_id)))
     JOIN public.granted_role gr ON ((ug.user_group_id = gr.user_group_id)))
     JOIN public.role r ON ((gr.role_id = r.role_id)))
     JOIN public.permission_role pr ON ((r.role_id = pr.role_id)))
     JOIN public.permission p ON ((p.permission_id = pr.permission_id)))
     JOIN public.granted_role_access_path grap ON ((gr.granted_role_id = grap.granted_role_id)))
     JOIN public.locations l ON ((l.location_id = grap.location_id)))
     JOIN public.locations l1 ON ((l1.access_path ~~ (l.access_path || '%'::text))))
  WHERE ((ugm.deleted_at IS NULL) AND (ug.deleted_at IS NULL) AND (gr.deleted_at IS NULL) AND (r.deleted_at IS NULL) AND (pr.deleted_at IS NULL) AND (p.deleted_at IS NULL) AND (grap.deleted_at IS NULL) AND (l.deleted_at IS NULL) AND (l1.deleted_at IS NULL));


ALTER TABLE public.granted_permissions OWNER TO postgres;

--
-- Name: groups; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.groups (
    group_id text NOT NULL,
    name text NOT NULL,
    description text,
    privileges jsonb,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.groups FORCE ROW LEVEL SECURITY;


ALTER TABLE public.groups OWNER TO postgres;

--
-- Name: import_study_plan_task; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.import_study_plan_task (
    task_id text NOT NULL,
    study_plan_id text NOT NULL,
    status text NOT NULL,
    error_detail text,
    imported_by text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT public.autofillresourcepath()
);

ALTER TABLE ONLY public.import_study_plan_task FORCE ROW LEVEL SECURITY;


ALTER TABLE public.import_study_plan_task OWNER TO postgres;

--
-- Name: individual_study_plan; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.individual_study_plan (
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    student_id text NOT NULL,
    status text NOT NULL,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    available_from timestamp with time zone,
    available_to timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    school_date timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath()
);

ALTER TABLE ONLY public.individual_study_plan FORCE ROW LEVEL SECURITY;


ALTER TABLE public.individual_study_plan OWNER TO postgres;

--
-- Name: individual_study_plans_view; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.individual_study_plans_view AS
 SELECT individual_study_plan_fn.student_id,
    individual_study_plan_fn.study_plan_id,
    individual_study_plan_fn.book_id,
    individual_study_plan_fn.chapter_id,
    individual_study_plan_fn.chapter_display_order,
    individual_study_plan_fn.topic_id,
    individual_study_plan_fn.topic_display_order,
    individual_study_plan_fn.learning_material_id,
    individual_study_plan_fn.lm_display_order,
    individual_study_plan_fn.start_date,
    individual_study_plan_fn.end_date,
    individual_study_plan_fn.available_from,
    individual_study_plan_fn.available_to,
    individual_study_plan_fn.school_date,
    individual_study_plan_fn.updated_at,
    individual_study_plan_fn.status,
    individual_study_plan_fn.resource_path
   FROM public.individual_study_plan_fn() individual_study_plan_fn(student_id, study_plan_id, book_id, chapter_id, chapter_display_order, topic_id, topic_display_order, learning_material_id, lm_display_order, start_date, end_date, available_from, available_to, school_date, updated_at, status, resource_path);


ALTER TABLE public.individual_study_plans_view OWNER TO postgres;

--
-- Name: learning_objective; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.learning_objective (
    video text,
    study_guide text,
    video_script text,
    manual_grading boolean DEFAULT false,
    CONSTRAINT learning_objective_type_check CHECK ((type = 'LEARNING_MATERIAL_LEARNING_OBJECTIVE'::text))
)
INHERITS (public.learning_material);

ALTER TABLE ONLY public.learning_objective FORCE ROW LEVEL SECURITY;


ALTER TABLE public.learning_objective OWNER TO postgres;

--
-- Name: learning_objectives; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.learning_objectives (
    lo_id text NOT NULL,
    name text NOT NULL,
    country text,
    grade smallint,
    subject text,
    topic_id text,
    master_lo_id text,
    display_order smallint,
    prerequisites text[],
    video text,
    study_guide text,
    video_script text,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    school_id integer DEFAULT '-2147483648'::integer NOT NULL,
    deleted_at timestamp with time zone,
    copied_from text,
    type text,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    instruction text,
    grade_to_pass integer,
    manual_grading boolean DEFAULT false,
    time_limit integer,
    maximum_attempt integer,
    approve_grading boolean DEFAULT false NOT NULL,
    grade_capping boolean DEFAULT false NOT NULL,
    review_option text DEFAULT 'EXAM_LO_REVIEW_OPTION_IMMEDIATELY'::text NOT NULL,
    vendor_type text DEFAULT 'LM_VENDOR_TYPE_MANABIE'::text NOT NULL,
    vendor_reference_id text,
    CONSTRAINT learning_objectives_review_option_check CHECK ((review_option = ANY (ARRAY['EXAM_LO_REVIEW_OPTION_IMMEDIATELY'::text, 'EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE'::text]))),
    CONSTRAINT vendor_type_check CHECK ((vendor_type = ANY (ARRAY['LM_VENDOR_TYPE_MANABIE'::text, 'LM_VENDOR_TYPE_LEARNOSITY'::text])))
);

ALTER TABLE ONLY public.learning_objectives FORCE ROW LEVEL SECURITY;


ALTER TABLE public.learning_objectives OWNER TO postgres;

--
-- Name: lms_learning_material_list; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.lms_learning_material_list (
    lm_list_id text NOT NULL,
    lm_ids text[],
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath()
);

ALTER TABLE ONLY public.lms_learning_material_list FORCE ROW LEVEL SECURITY;


ALTER TABLE public.lms_learning_material_list OWNER TO postgres;

--
-- Name: lms_student_study_plan_item; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.lms_student_study_plan_item (
    student_id text NOT NULL,
    lm_list_id text NOT NULL,
    study_plan_id text NOT NULL,
    type text DEFAULT 'STATIC'::text,
    status text DEFAULT 'STUDY_PLAN_STATUS_ACTIVE'::text,
    display_order integer DEFAULT 0 NOT NULL,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    study_plan_item_id text NOT NULL,
    master_study_plan_item_id text,
    CONSTRAINT lms_student_study_plan_item_status_check CHECK ((status = ANY (ARRAY['STUDY_PLAN_STATUS_NONE'::text, 'STUDY_PLAN_STATUS_ACTIVE'::text, 'STUDY_PLAN_STATUS_ARCHIVED'::text])))
);

ALTER TABLE ONLY public.lms_student_study_plan_item FORCE ROW LEVEL SECURITY;


ALTER TABLE public.lms_student_study_plan_item OWNER TO postgres;

--
-- Name: lms_student_study_plans; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.lms_student_study_plans (
    student_id text NOT NULL,
    study_plan_id text NOT NULL,
    status text DEFAULT 'STUDY_PLAN_STATUS_ACTIVE'::text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    CONSTRAINT student_study_plans_status_check CHECK ((status = ANY (ARRAY['STUDY_PLAN_STATUS_NONE'::text, 'STUDY_PLAN_STATUS_ACTIVE'::text, 'STUDY_PLAN_STATUS_ARCHIVED'::text])))
);

ALTER TABLE ONLY public.lms_student_study_plans FORCE ROW LEVEL SECURITY;


ALTER TABLE public.lms_student_study_plans OWNER TO postgres;

--
-- Name: lms_study_plan_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.lms_study_plan_items (
    study_plan_item_id text NOT NULL,
    study_plan_id text NOT NULL,
    lm_list_id text NOT NULL,
    name text NOT NULL,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    display_order integer DEFAULT 0 NOT NULL,
    status text DEFAULT 'STUDY_PLAN_ITEM_STATUS_ACTIVE'::text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    CONSTRAINT study_plan_item_status_check CHECK ((status = ANY (ARRAY['STUDY_PLAN_STATUS_NONE'::text, 'STUDY_PLAN_STATUS_ACTIVE'::text, 'STUDY_PLAN_STATUS_ARCHIVED'::text])))
);

ALTER TABLE ONLY public.lms_study_plan_items FORCE ROW LEVEL SECURITY;


ALTER TABLE public.lms_study_plan_items OWNER TO postgres;

--
-- Name: lms_study_plans; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.lms_study_plans (
    study_plan_id text NOT NULL,
    name text NOT NULL,
    course_id text NOT NULL,
    academic_year text,
    status text DEFAULT 'STUDY_PLAN_STATUS_ACTIVE'::text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    CONSTRAINT study_plan_status_check CHECK ((status = ANY (ARRAY['STUDY_PLAN_STATUS_NONE'::text, 'STUDY_PLAN_STATUS_ACTIVE'::text, 'STUDY_PLAN_STATUS_ARCHIVED'::text])))
);

ALTER TABLE ONLY public.lms_study_plans FORCE ROW LEVEL SECURITY;


ALTER TABLE public.lms_study_plans OWNER TO postgres;

--
-- Name: lo_progression; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.lo_progression (
    progression_id text NOT NULL,
    shuffled_quiz_set_id text NOT NULL,
    student_id text NOT NULL,
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    quiz_external_ids text[],
    last_index integer NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    session_id text
);

ALTER TABLE ONLY public.lo_progression FORCE ROW LEVEL SECURITY;


ALTER TABLE public.lo_progression OWNER TO postgres;

--
-- Name: lo_progression_answer; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.lo_progression_answer (
    progression_answer_id text NOT NULL,
    shuffled_quiz_set_id text NOT NULL,
    quiz_external_id text NOT NULL,
    progression_id text NOT NULL,
    student_id text NOT NULL,
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    student_text_answer text[],
    student_index_answer integer[],
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    submitted_keys_answer text[]
);

ALTER TABLE ONLY public.lo_progression_answer FORCE ROW LEVEL SECURITY;


ALTER TABLE public.lo_progression_answer OWNER TO postgres;

--
-- Name: lo_study_plan_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.lo_study_plan_items (
    lo_id text NOT NULL,
    study_plan_item_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.lo_study_plan_items FORCE ROW LEVEL SECURITY;


ALTER TABLE public.lo_study_plan_items OWNER TO postgres;

--
-- Name: lo_submission; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.lo_submission (
    submission_id text NOT NULL,
    student_id text NOT NULL,
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    shuffled_quiz_set_id text NOT NULL,
    total_point integer DEFAULT 0,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    is_submitted boolean DEFAULT false NOT NULL
);

ALTER TABLE ONLY public.lo_submission FORCE ROW LEVEL SECURITY;


ALTER TABLE public.lo_submission OWNER TO postgres;

--
-- Name: lo_submission_answer; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.lo_submission_answer (
    student_id text NOT NULL,
    quiz_id text NOT NULL,
    submission_id text NOT NULL,
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    shuffled_quiz_set_id text NOT NULL,
    student_text_answer text[],
    correct_text_answer text[],
    student_index_answer integer[],
    correct_index_answer integer[],
    point integer DEFAULT 0,
    is_correct boolean[] DEFAULT '{}'::boolean[] NOT NULL,
    is_accepted boolean DEFAULT false,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    submitted_keys_answer text[],
    correct_keys_answer text[]
);

ALTER TABLE ONLY public.lo_submission_answer FORCE ROW LEVEL SECURITY;


ALTER TABLE public.lo_submission_answer OWNER TO postgres;

--
-- Name: lo_video_rating; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.lo_video_rating (
    lo_id text NOT NULL,
    video_id text NOT NULL,
    learner_id text NOT NULL,
    rating_value public.rating_type NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath()
);

ALTER TABLE ONLY public.lo_video_rating FORCE ROW LEVEL SECURITY;


ALTER TABLE public.lo_video_rating OWNER TO postgres;

--
-- Name: master_study_plan; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.master_study_plan (
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    status text NOT NULL,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    available_from timestamp with time zone,
    available_to timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    school_date timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath()
);

ALTER TABLE ONLY public.master_study_plan FORCE ROW LEVEL SECURITY;


ALTER TABLE public.master_study_plan OWNER TO postgres;

--
-- Name: master_study_plan_view; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.master_study_plan_view AS
 SELECT master_study_plan_fn.study_plan_id,
    master_study_plan_fn.book_id,
    master_study_plan_fn.chapter_id,
    master_study_plan_fn.chapter_display_order,
    master_study_plan_fn.topic_id,
    master_study_plan_fn.topic_display_order,
    master_study_plan_fn.learning_material_id,
    master_study_plan_fn.lm_display_order,
    master_study_plan_fn.resource_path,
    master_study_plan_fn.start_date,
    master_study_plan_fn.end_date,
    master_study_plan_fn.available_from,
    master_study_plan_fn.available_to,
    master_study_plan_fn.school_date,
    master_study_plan_fn.updated_at,
    master_study_plan_fn.status
   FROM public.master_study_plan_fn() master_study_plan_fn(study_plan_id, book_id, chapter_id, chapter_display_order, topic_id, topic_display_order, learning_material_id, lm_display_order, resource_path, start_date, end_date, available_from, available_to, school_date, updated_at, status);


ALTER TABLE public.master_study_plan_view OWNER TO postgres;

--
-- Name: max_score_submission; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.max_score_submission (
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    student_id text NOT NULL,
    max_score integer,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    total_score integer,
    max_percentage integer
);

ALTER TABLE ONLY public.max_score_submission FORCE ROW LEVEL SECURITY;


ALTER TABLE public.max_score_submission OWNER TO postgres;

--
-- Name: question_group; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.question_group (
    question_group_id text NOT NULL,
    learning_material_id text NOT NULL,
    name text,
    description text,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    rich_description jsonb
);

ALTER TABLE ONLY public.question_group FORCE ROW LEVEL SECURITY;


ALTER TABLE public.question_group OWNER TO postgres;

--
-- Name: question_tag; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.question_tag (
    question_tag_id text NOT NULL,
    name text NOT NULL,
    question_tag_type_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath()
);

ALTER TABLE ONLY public.question_tag FORCE ROW LEVEL SECURITY;


ALTER TABLE public.question_tag OWNER TO postgres;

--
-- Name: question_tag_type; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.question_tag_type (
    question_tag_type_id text NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath()
);

ALTER TABLE ONLY public.question_tag_type FORCE ROW LEVEL SECURITY;


ALTER TABLE public.question_tag_type OWNER TO postgres;

--
-- Name: quiz_sets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.quiz_sets (
    quiz_set_id text NOT NULL,
    lo_id text NOT NULL,
    quiz_external_ids text[] NOT NULL,
    status text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    question_hierarchy jsonb[] DEFAULT ARRAY[]::jsonb[]
);

ALTER TABLE ONLY public.quiz_sets FORCE ROW LEVEL SECURITY;


ALTER TABLE public.quiz_sets OWNER TO postgres;

--
-- Name: shuffled_quiz_sets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.shuffled_quiz_sets (
    shuffled_quiz_set_id text NOT NULL,
    original_quiz_set_id text,
    quiz_external_ids text[],
    status text,
    random_seed text,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    student_id text NOT NULL,
    study_plan_item_id text,
    total_correctness integer DEFAULT 0 NOT NULL,
    submission_history jsonb DEFAULT '[]'::jsonb NOT NULL,
    session_id text,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    original_shuffle_quiz_set_id text,
    study_plan_id text,
    learning_material_id text,
    question_hierarchy jsonb[] DEFAULT ARRAY[]::jsonb[]
);

ALTER TABLE ONLY public.shuffled_quiz_sets FORCE ROW LEVEL SECURITY;


ALTER TABLE public.shuffled_quiz_sets OWNER TO postgres;

--
-- Name: student_event_logs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.student_event_logs (
    student_event_log_id integer NOT NULL,
    student_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    event_type character varying(100) NOT NULL,
    payload jsonb,
    event_id character varying(50),
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    study_plan_item_id text,
    study_plan_id text,
    learning_material_id text
);


ALTER TABLE public.student_event_logs OWNER TO postgres;

--
-- Name: student_event_logs_student_event_log_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.student_event_logs_student_event_log_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.student_event_logs_student_event_log_id_seq OWNER TO postgres;

--
-- Name: student_event_logs_student_event_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.student_event_logs_student_event_log_id_seq OWNED BY public.student_event_logs.student_event_log_id;


--
-- Name: student_latest_submissions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.student_latest_submissions (
    study_plan_item_id text,
    assignment_id text,
    student_id text NOT NULL,
    student_submission_id text NOT NULL,
    submission_content jsonb,
    check_list jsonb,
    status text,
    note text,
    editor_id text,
    student_submission_grade_id text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    deleted_by text,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    complete_date timestamp with time zone,
    duration integer,
    correct_score real,
    total_score real,
    understanding_level text,
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL
);

ALTER TABLE ONLY public.student_latest_submissions FORCE ROW LEVEL SECURITY;


ALTER TABLE public.student_latest_submissions OWNER TO postgres;

--
-- Name: student_learning_time_by_daily; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.student_learning_time_by_daily (
    learning_time_id integer NOT NULL,
    student_id text NOT NULL,
    learning_time integer DEFAULT 0 NOT NULL,
    day timestamp with time zone NOT NULL,
    sessions text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    assignment_learning_time integer DEFAULT 0 NOT NULL,
    assignment_submission_ids text[]
);

ALTER TABLE ONLY public.student_learning_time_by_daily FORCE ROW LEVEL SECURITY;


ALTER TABLE public.student_learning_time_by_daily OWNER TO postgres;

--
-- Name: COLUMN student_learning_time_by_daily.learning_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.student_learning_time_by_daily.learning_time IS 'learning time in seconds unit';


--
-- Name: student_learning_time_by_daily_learning_time_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.student_learning_time_by_daily_learning_time_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.student_learning_time_by_daily_learning_time_id_seq OWNER TO postgres;

--
-- Name: student_learning_time_by_daily_learning_time_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.student_learning_time_by_daily_learning_time_id_seq OWNED BY public.student_learning_time_by_daily.learning_time_id;


--
-- Name: student_study_plans; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.student_study_plans (
    study_plan_id text NOT NULL,
    student_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    master_study_plan_id text
);

ALTER TABLE ONLY public.student_study_plans FORCE ROW LEVEL SECURITY;


ALTER TABLE public.student_study_plans OWNER TO postgres;

--
-- Name: student_submission_grades; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.student_submission_grades (
    student_submission_grade_id text NOT NULL,
    student_submission_id text NOT NULL,
    grade_content jsonb,
    grader_id text,
    grader_comment text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    grade numeric(10,2),
    status text,
    editor_id text,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.student_submission_grades FORCE ROW LEVEL SECURITY;


ALTER TABLE public.student_submission_grades OWNER TO postgres;

--
-- Name: student_submissions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.student_submissions (
    student_submission_id text NOT NULL,
    study_plan_item_id text,
    assignment_id text,
    student_id text NOT NULL,
    submission_content jsonb,
    check_list jsonb,
    status text,
    note text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    student_submission_grade_id text,
    editor_id text,
    deleted_by text,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    complete_date timestamp with time zone,
    duration integer,
    correct_score real,
    total_score real,
    understanding_level text,
    study_plan_id text,
    learning_material_id text
);

ALTER TABLE ONLY public.student_submissions FORCE ROW LEVEL SECURITY;


ALTER TABLE public.student_submissions OWNER TO postgres;

--
-- Name: students; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.students (
    student_id text NOT NULL,
    current_grade smallint,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    student_external_id text,
    grade_id text
);

ALTER TABLE ONLY public.students FORCE ROW LEVEL SECURITY;


ALTER TABLE public.students OWNER TO postgres;

--
-- Name: students_learning_objectives_completeness; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.students_learning_objectives_completeness (
    student_id text NOT NULL,
    lo_id text NOT NULL,
    preset_study_plan_weekly_id text,
    first_attempt_score smallint DEFAULT 0 NOT NULL,
    is_finished_quiz boolean DEFAULT false NOT NULL,
    is_finished_video boolean DEFAULT false NOT NULL,
    is_finished_study_guide boolean DEFAULT false NOT NULL,
    first_quiz_correctness real,
    finished_quiz_at timestamp with time zone,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    highest_quiz_score real,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.students_learning_objectives_completeness FORCE ROW LEVEL SECURITY;


ALTER TABLE public.students_learning_objectives_completeness OWNER TO postgres;

--
-- Name: students_topics_completeness; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.students_topics_completeness (
    student_id text NOT NULL,
    topic_id text NOT NULL,
    total_finished_los integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    is_completed boolean DEFAULT false,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.students_topics_completeness FORCE ROW LEVEL SECURITY;


ALTER TABLE public.students_topics_completeness OWNER TO postgres;

--
-- Name: study_plan_assessment; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.study_plan_assessment (
    id text NOT NULL,
    study_plan_item_id text NOT NULL,
    learning_material_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    ref_table character varying(20) NOT NULL
);

ALTER TABLE ONLY public.study_plan_assessment FORCE ROW LEVEL SECURITY;


ALTER TABLE public.study_plan_assessment OWNER TO postgres;

--
-- Name: study_plan_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.study_plan_items (
    study_plan_item_id text NOT NULL,
    study_plan_id text,
    available_from timestamp with time zone,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    deleted_at timestamp with time zone,
    available_to timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    copy_study_plan_item_id text,
    content_structure jsonb,
    completed_at timestamp with time zone,
    display_order integer DEFAULT 0 NOT NULL,
    content_structure_flatten text,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    status text DEFAULT 'STUDY_PLAN_ITEM_STATUS_ACTIVE'::text,
    school_date timestamp with time zone,
    CONSTRAINT status_check CHECK ((status = ANY (ARRAY['STUDY_PLAN_ITEM_STATUS_NONE'::text, 'STUDY_PLAN_ITEM_STATUS_ACTIVE'::text, 'STUDY_PLAN_ITEM_STATUS_ARCHIVED'::text])))
);

ALTER TABLE ONLY public.study_plan_items FORCE ROW LEVEL SECURITY;


ALTER TABLE public.study_plan_items OWNER TO postgres;

--
-- Name: study_plan_monitors; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.study_plan_monitors (
    study_plan_monitor_id text NOT NULL,
    student_id text,
    course_id text,
    type text,
    payload jsonb,
    level text,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    auto_upserted_at timestamp with time zone
);

ALTER TABLE ONLY public.study_plan_monitors FORCE ROW LEVEL SECURITY;


ALTER TABLE public.study_plan_monitors OWNER TO postgres;

--
-- Name: study_plan_tree; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.study_plan_tree AS
 SELECT study_plan_tree_fn.study_plan_id,
    study_plan_tree_fn.book_id,
    study_plan_tree_fn.chapter_id,
    study_plan_tree_fn.chapter_display_order,
    study_plan_tree_fn.topic_id,
    study_plan_tree_fn.topic_display_order,
    study_plan_tree_fn.learning_material_id,
    study_plan_tree_fn.lm_display_order,
    study_plan_tree_fn.resource_path
   FROM public.study_plan_tree_fn() study_plan_tree_fn(study_plan_id, book_id, chapter_id, chapter_display_order, topic_id, topic_display_order, learning_material_id, lm_display_order, resource_path);


ALTER TABLE public.study_plan_tree OWNER TO postgres;

--
-- Name: tagged_user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tagged_user (
    user_id text NOT NULL,
    tag_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text
);

ALTER TABLE ONLY public.tagged_user FORCE ROW LEVEL SECURITY;


ALTER TABLE public.tagged_user OWNER TO postgres;

--
-- Name: task_assignment; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.task_assignment (
    attachments text[],
    instruction text,
    require_duration boolean,
    require_complete_date boolean,
    require_understanding_level boolean,
    require_correctness boolean,
    require_attachment boolean,
    require_assignment_note boolean,
    CONSTRAINT task_assignment_type_check CHECK ((type = 'LEARNING_MATERIAL_TASK_ASSIGNMENT'::text))
)
INHERITS (public.learning_material);

ALTER TABLE ONLY public.task_assignment FORCE ROW LEVEL SECURITY;


ALTER TABLE public.task_assignment OWNER TO postgres;

--
-- Name: topics; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.topics (
    topic_id text NOT NULL,
    name text NOT NULL,
    country text,
    grade smallint NOT NULL,
    subject text NOT NULL,
    topic_type text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    status text DEFAULT 'TOPIC_STATUS_PUBLISHED'::text,
    display_order smallint,
    published_at timestamp with time zone,
    total_los integer DEFAULT 0 NOT NULL,
    chapter_id text,
    icon_url text,
    school_id integer DEFAULT '-2147483648'::integer NOT NULL,
    attachment_urls text[],
    instruction text,
    copied_topic_id text,
    essay_required boolean DEFAULT false NOT NULL,
    attachment_names text[],
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    lo_display_order_counter integer DEFAULT 0
);

ALTER TABLE ONLY public.topics FORCE ROW LEVEL SECURITY;


ALTER TABLE public.topics OWNER TO postgres;

--
-- Name: topics_assignments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.topics_assignments (
    topic_id text NOT NULL,
    assignment_id text NOT NULL,
    display_order smallint,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.topics_assignments FORCE ROW LEVEL SECURITY;


ALTER TABLE public.topics_assignments OWNER TO postgres;

--
-- Name: topics_learning_objectives; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.topics_learning_objectives (
    topic_id text NOT NULL,
    lo_id text NOT NULL,
    display_order smallint,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.topics_learning_objectives FORCE ROW LEVEL SECURITY;


ALTER TABLE public.topics_learning_objectives OWNER TO postgres;

--
-- Name: user_access_paths; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_access_paths (
    user_id text NOT NULL,
    location_id text NOT NULL,
    access_path text,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.user_access_paths FORCE ROW LEVEL SECURITY;


ALTER TABLE public.user_access_paths OWNER TO postgres;

--
-- Name: user_tag; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_tag (
    user_tag_id text NOT NULL,
    user_tag_name text NOT NULL,
    user_tag_type text NOT NULL,
    is_archived boolean NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    user_tag_partner_id text NOT NULL
);

ALTER TABLE ONLY public.user_tag FORCE ROW LEVEL SECURITY;


ALTER TABLE public.user_tag OWNER TO postgres;

--
-- Name: users_groups; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users_groups (
    user_id text NOT NULL,
    group_id text NOT NULL,
    is_origin boolean NOT NULL,
    status text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.users_groups FORCE ROW LEVEL SECURITY;


ALTER TABLE public.users_groups OWNER TO postgres;

--
-- Name: withus_failed_sync_email_recipient; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.withus_failed_sync_email_recipient (
    recipient_id text NOT NULL,
    email_address text NOT NULL,
    last_updated_date timestamp with time zone,
    last_updated_by text,
    is_archived boolean DEFAULT false,
    resource_path text DEFAULT public.autofillresourcepath()
);

ALTER TABLE ONLY public.withus_failed_sync_email_recipient FORCE ROW LEVEL SECURITY;


ALTER TABLE public.withus_failed_sync_email_recipient OWNER TO postgres;

--
-- Name: withus_mapping_course_id; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.withus_mapping_course_id (
    manabie_course_id text NOT NULL,
    withus_course_id text DEFAULT ''::text NOT NULL,
    last_updated_date timestamp with time zone,
    last_updated_by text,
    is_archived boolean DEFAULT false,
    resource_path text DEFAULT public.autofillresourcepath()
);

ALTER TABLE ONLY public.withus_mapping_course_id FORCE ROW LEVEL SECURITY;


ALTER TABLE public.withus_mapping_course_id OWNER TO postgres;

--
-- Name: withus_mapping_exam_lo_id; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.withus_mapping_exam_lo_id (
    exam_lo_id text NOT NULL,
    material_code text DEFAULT ''::text NOT NULL,
    last_updated_date timestamp with time zone,
    last_updated_by text,
    is_archived boolean DEFAULT false,
    resource_path text DEFAULT public.autofillresourcepath()
);

ALTER TABLE ONLY public.withus_mapping_exam_lo_id FORCE ROW LEVEL SECURITY;


ALTER TABLE public.withus_mapping_exam_lo_id OWNER TO postgres;

--
-- Name: withus_mapping_question_tag; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.withus_mapping_question_tag (
    manabie_tag_id text NOT NULL,
    manabie_tag_name text NOT NULL,
    withus_tag_name text DEFAULT ''::text NOT NULL,
    last_updated_date timestamp with time zone,
    last_updated_by text,
    is_archived boolean DEFAULT false,
    resource_path text DEFAULT public.autofillresourcepath()
);

ALTER TABLE ONLY public.withus_mapping_question_tag FORCE ROW LEVEL SECURITY;


ALTER TABLE public.withus_mapping_question_tag OWNER TO postgres;

--
-- Name: assignment resource_path; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assignment ALTER COLUMN resource_path SET DEFAULT public.autofillresourcepath();


--
-- Name: assignment vendor_type; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assignment ALTER COLUMN vendor_type SET DEFAULT 'LM_VENDOR_TYPE_MANABIE'::text;


--
-- Name: assignment is_published; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assignment ALTER COLUMN is_published SET DEFAULT false;


--
-- Name: exam_lo resource_path; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.exam_lo ALTER COLUMN resource_path SET DEFAULT public.autofillresourcepath();


--
-- Name: exam_lo vendor_type; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.exam_lo ALTER COLUMN vendor_type SET DEFAULT 'LM_VENDOR_TYPE_MANABIE'::text;


--
-- Name: exam_lo is_published; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.exam_lo ALTER COLUMN is_published SET DEFAULT false;


--
-- Name: flash_card resource_path; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.flash_card ALTER COLUMN resource_path SET DEFAULT public.autofillresourcepath();


--
-- Name: flash_card vendor_type; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.flash_card ALTER COLUMN vendor_type SET DEFAULT 'LM_VENDOR_TYPE_MANABIE'::text;


--
-- Name: flash_card is_published; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.flash_card ALTER COLUMN is_published SET DEFAULT false;


--
-- Name: learning_objective resource_path; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.learning_objective ALTER COLUMN resource_path SET DEFAULT public.autofillresourcepath();


--
-- Name: learning_objective vendor_type; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.learning_objective ALTER COLUMN vendor_type SET DEFAULT 'LM_VENDOR_TYPE_MANABIE'::text;


--
-- Name: learning_objective is_published; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.learning_objective ALTER COLUMN is_published SET DEFAULT false;


--
-- Name: student_event_logs student_event_log_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_event_logs ALTER COLUMN student_event_log_id SET DEFAULT nextval('public.student_event_logs_student_event_log_id_seq'::regclass);


--
-- Name: student_learning_time_by_daily learning_time_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_learning_time_by_daily ALTER COLUMN learning_time_id SET DEFAULT nextval('public.student_learning_time_by_daily_learning_time_id_seq'::regclass);


--
-- Name: task_assignment resource_path; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.task_assignment ALTER COLUMN resource_path SET DEFAULT public.autofillresourcepath();


--
-- Name: task_assignment vendor_type; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.task_assignment ALTER COLUMN vendor_type SET DEFAULT 'LM_VENDOR_TYPE_MANABIE'::text;


--
-- Name: task_assignment is_published; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.task_assignment ALTER COLUMN is_published SET DEFAULT false;


--
-- Name: alloydb_dbz_signal alloydb_dbz_signal_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alloydb_dbz_signal
    ADD CONSTRAINT alloydb_dbz_signal_pkey PRIMARY KEY (id);


--
-- Name: assessment_session assessment_session_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assessment_session
    ADD CONSTRAINT assessment_session_pk PRIMARY KEY (session_id);


--
-- Name: assessment_submission assessment_submission_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assessment_submission
    ADD CONSTRAINT assessment_submission_pk PRIMARY KEY (id);


--
-- Name: assign_study_plan_tasks assign_study_plan_tasks_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assign_study_plan_tasks
    ADD CONSTRAINT assign_study_plan_tasks_pk PRIMARY KEY (id);


--
-- Name: assignment assignment_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assignment
    ADD CONSTRAINT assignment_pk PRIMARY KEY (learning_material_id);


--
-- Name: assignment_study_plan_items assignment_study_plan_items_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assignment_study_plan_items
    ADD CONSTRAINT assignment_study_plan_items_pk PRIMARY KEY (study_plan_item_id, assignment_id);


--
-- Name: assignment_study_plan_items assignment_study_plan_items_study_plan_item_id_un; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assignment_study_plan_items
    ADD CONSTRAINT assignment_study_plan_items_study_plan_item_id_un UNIQUE (study_plan_item_id);


--
-- Name: assignments assignments_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assignments
    ADD CONSTRAINT assignments_pk PRIMARY KEY (assignment_id);


--
-- Name: books_chapters books_chapters_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.books_chapters
    ADD CONSTRAINT books_chapters_pk PRIMARY KEY (book_id, chapter_id);


--
-- Name: books books_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.books
    ADD CONSTRAINT books_pk PRIMARY KEY (book_id);


--
-- Name: cerebry_classes cerebry_classes_name_un; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cerebry_classes
    ADD CONSTRAINT cerebry_classes_name_un UNIQUE (class_code);


--
-- Name: cerebry_classes cerebry_classes_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cerebry_classes
    ADD CONSTRAINT cerebry_classes_pk PRIMARY KEY (id);


--
-- Name: chapters chapters_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chapters
    ADD CONSTRAINT chapters_pk PRIMARY KEY (chapter_id);


--
-- Name: class_students class_students_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.class_students
    ADD CONSTRAINT class_students_pk PRIMARY KEY (student_id, class_id);


--
-- Name: class_study_plans class_study_plans_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.class_study_plans
    ADD CONSTRAINT class_study_plans_pk PRIMARY KEY (class_id, study_plan_id);


--
-- Name: course_access_paths course_access_paths_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.course_access_paths
    ADD CONSTRAINT course_access_paths_pk PRIMARY KEY (course_id, location_id);


--
-- Name: course_classes course_classes_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.course_classes
    ADD CONSTRAINT course_classes_pk PRIMARY KEY (course_id, class_id);


--
-- Name: course_students course_student_id_un; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.course_students
    ADD CONSTRAINT course_student_id_un UNIQUE (course_student_id);


--
-- Name: course_students course_student_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.course_students
    ADD CONSTRAINT course_student_pk PRIMARY KEY (course_id, student_id);


--
-- Name: course_students_access_paths course_students_access_paths_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.course_students_access_paths
    ADD CONSTRAINT course_students_access_paths_pk PRIMARY KEY (course_student_id, location_id);


--
-- Name: course_student_subscriptions course_students_subscriptions_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.course_student_subscriptions
    ADD CONSTRAINT course_students_subscriptions_pk PRIMARY KEY (course_student_subscription_id);


--
-- Name: course_study_plans course_study_plans_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.course_study_plans
    ADD CONSTRAINT course_study_plans_pk PRIMARY KEY (course_id, study_plan_id);


--
-- Name: courses_books courses_books_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.courses_books
    ADD CONSTRAINT courses_books_pk PRIMARY KEY (book_id, course_id);


--
-- Name: courses courses_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.courses
    ADD CONSTRAINT courses_pk PRIMARY KEY (course_id);


--
-- Name: dbz_signals dbz_signals_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dbz_signals
    ADD CONSTRAINT dbz_signals_pkey PRIMARY KEY (id);


--
-- Name: student_event_logs event_id_un; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_event_logs
    ADD CONSTRAINT event_id_un UNIQUE (event_id);


--
-- Name: student_event_logs event_log_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_event_logs
    ADD CONSTRAINT event_log_pk PRIMARY KEY (student_event_log_id);


--
-- Name: exam_lo exam_lo_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.exam_lo
    ADD CONSTRAINT exam_lo_pk PRIMARY KEY (learning_material_id);


--
-- Name: exam_lo_submission_answer exam_lo_submission_answer_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.exam_lo_submission_answer
    ADD CONSTRAINT exam_lo_submission_answer_pk PRIMARY KEY (student_id, quiz_id, submission_id);


--
-- Name: exam_lo_submission exam_lo_submission_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.exam_lo_submission
    ADD CONSTRAINT exam_lo_submission_pk PRIMARY KEY (submission_id);


--
-- Name: exam_lo_submission_score exam_lo_submission_score_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.exam_lo_submission_score
    ADD CONSTRAINT exam_lo_submission_score_pk PRIMARY KEY (submission_id, quiz_id);


--
-- Name: feedback_session feedback_session_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.feedback_session
    ADD CONSTRAINT feedback_session_pk PRIMARY KEY (id);


--
-- Name: flash_card flash_card_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.flash_card
    ADD CONSTRAINT flash_card_pk PRIMARY KEY (learning_material_id);


--
-- Name: flash_card_submission_answer flash_card_submission_answer_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.flash_card_submission_answer
    ADD CONSTRAINT flash_card_submission_answer_pk PRIMARY KEY (student_id, quiz_id, submission_id);


--
-- Name: flash_card_submission flash_card_submission_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.flash_card_submission
    ADD CONSTRAINT flash_card_submission_pk PRIMARY KEY (submission_id);


--
-- Name: flash_card_submission flash_card_submission_shuffled_quiz_set_id_un; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.flash_card_submission
    ADD CONSTRAINT flash_card_submission_shuffled_quiz_set_id_un UNIQUE (shuffled_quiz_set_id);


--
-- Name: flashcard_progressions flashcard_progressions_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.flashcard_progressions
    ADD CONSTRAINT flashcard_progressions_pk PRIMARY KEY (study_set_id);


--
-- Name: flashcard_speeches flashcard_speeches_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.flashcard_speeches
    ADD CONSTRAINT flashcard_speeches_pk PRIMARY KEY (speech_id);


--
-- Name: grade_book_setting grade_book_setting_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.grade_book_setting
    ADD CONSTRAINT grade_book_setting_pk PRIMARY KEY (setting);


--
-- Name: grade grade_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.grade
    ADD CONSTRAINT grade_pk PRIMARY KEY (grade_id);


--
-- Name: granted_permission granted_permission__pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.granted_permission
    ADD CONSTRAINT granted_permission__pk PRIMARY KEY (user_group_id, role_id, permission_id, location_id);


--
-- Name: import_study_plan_task import_study_plan_task_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.import_study_plan_task
    ADD CONSTRAINT import_study_plan_task_pk PRIMARY KEY (task_id);


--
-- Name: individual_study_plan learning_material_id_student_id_study_plan_id_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.individual_study_plan
    ADD CONSTRAINT learning_material_id_student_id_study_plan_id_pk PRIMARY KEY (learning_material_id, student_id, study_plan_id);


--
-- Name: master_study_plan learning_material_id_study_plan_id_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.master_study_plan
    ADD CONSTRAINT learning_material_id_study_plan_id_pk PRIMARY KEY (learning_material_id, study_plan_id);


--
-- Name: learning_material learning_material_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.learning_material
    ADD CONSTRAINT learning_material_pk PRIMARY KEY (learning_material_id);


--
-- Name: learning_objective learning_objective_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.learning_objective
    ADD CONSTRAINT learning_objective_pk PRIMARY KEY (learning_material_id);


--
-- Name: learning_objectives learning_objectives_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.learning_objectives
    ADD CONSTRAINT learning_objectives_pk PRIMARY KEY (lo_id);


--
-- Name: lms_learning_material_list lms_learning_material_list_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lms_learning_material_list
    ADD CONSTRAINT lms_learning_material_list_pkey PRIMARY KEY (lm_list_id);


--
-- Name: lms_student_study_plan_item lms_student_study_plan_item_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lms_student_study_plan_item
    ADD CONSTRAINT lms_student_study_plan_item_pkey PRIMARY KEY (study_plan_item_id);


--
-- Name: lms_student_study_plans lms_student_study_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lms_student_study_plans
    ADD CONSTRAINT lms_student_study_plans_pkey PRIMARY KEY (student_id, study_plan_id);


--
-- Name: lms_study_plan_items lms_study_plan_items_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lms_study_plan_items
    ADD CONSTRAINT lms_study_plan_items_pkey PRIMARY KEY (study_plan_item_id);


--
-- Name: lms_study_plans lms_study_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lms_study_plans
    ADD CONSTRAINT lms_study_plans_pkey PRIMARY KEY (study_plan_id);


--
-- Name: lo_progression_answer lo_progression_answer_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lo_progression_answer
    ADD CONSTRAINT lo_progression_answer_pk PRIMARY KEY (progression_answer_id);


--
-- Name: lo_progression_answer lo_progression_answer_un; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lo_progression_answer
    ADD CONSTRAINT lo_progression_answer_un UNIQUE (progression_id, quiz_external_id);


--
-- Name: lo_progression lo_progression_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lo_progression
    ADD CONSTRAINT lo_progression_pk PRIMARY KEY (progression_id);


--
-- Name: lo_study_plan_items lo_study_plan_items_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lo_study_plan_items
    ADD CONSTRAINT lo_study_plan_items_pk PRIMARY KEY (study_plan_item_id, lo_id);


--
-- Name: lo_study_plan_items lo_study_plan_items_study_plan_item_id_un; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lo_study_plan_items
    ADD CONSTRAINT lo_study_plan_items_study_plan_item_id_un UNIQUE (study_plan_item_id);


--
-- Name: lo_submission_answer lo_submission_answer_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lo_submission_answer
    ADD CONSTRAINT lo_submission_answer_pk PRIMARY KEY (student_id, quiz_id, submission_id);


--
-- Name: lo_submission lo_submission_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lo_submission
    ADD CONSTRAINT lo_submission_pkey PRIMARY KEY (submission_id);


--
-- Name: lo_video_rating lo_video_rating_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lo_video_rating
    ADD CONSTRAINT lo_video_rating_pk PRIMARY KEY (lo_id, video_id);


--
-- Name: locations locations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.locations
    ADD CONSTRAINT locations_pkey PRIMARY KEY (location_id);


--
-- Name: max_score_submission max_score_submission_study_plan_item_identity_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.max_score_submission
    ADD CONSTRAINT max_score_submission_study_plan_item_identity_pk PRIMARY KEY (learning_material_id, student_id, study_plan_id);


--
-- Name: academic_year pk__academic_year; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.academic_year
    ADD CONSTRAINT pk__academic_year PRIMARY KEY (academic_year_id);


--
-- Name: assessment pk__assessment_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assessment
    ADD CONSTRAINT pk__assessment_id PRIMARY KEY (id);


--
-- Name: granted_role pk__granted_role; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.granted_role
    ADD CONSTRAINT pk__granted_role PRIMARY KEY (granted_role_id);


--
-- Name: granted_role_access_path pk__granted_role_access_path; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.granted_role_access_path
    ADD CONSTRAINT pk__granted_role_access_path PRIMARY KEY (granted_role_id, location_id);


--
-- Name: groups pk__groups; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.groups
    ADD CONSTRAINT pk__groups PRIMARY KEY (group_id);


--
-- Name: permission pk__permission; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.permission
    ADD CONSTRAINT pk__permission PRIMARY KEY (permission_id);


--
-- Name: permission_role pk__permission_role; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.permission_role
    ADD CONSTRAINT pk__permission_role PRIMARY KEY (permission_id, role_id, resource_path);


--
-- Name: study_plan_assessment pk__sp_assessment_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.study_plan_assessment
    ADD CONSTRAINT pk__sp_assessment_id PRIMARY KEY (id);


--
-- Name: user_group pk__user_group; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_group
    ADD CONSTRAINT pk__user_group PRIMARY KEY (user_group_id);


--
-- Name: user_group_member pk__user_group_member; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_group_member
    ADD CONSTRAINT pk__user_group_member PRIMARY KEY (user_id, user_group_id);


--
-- Name: users_groups pk__users_groups; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users_groups
    ADD CONSTRAINT pk__users_groups PRIMARY KEY (user_id, group_id);


--
-- Name: allocate_marker pk_allocate_marker; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.allocate_marker
    ADD CONSTRAINT pk_allocate_marker PRIMARY KEY (student_id, study_plan_id, learning_material_id);


--
-- Name: question_group question_group_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.question_group
    ADD CONSTRAINT question_group_pk PRIMARY KEY (question_group_id);


--
-- Name: question_tag question_tag_id_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.question_tag
    ADD CONSTRAINT question_tag_id_pk PRIMARY KEY (question_tag_id);


--
-- Name: question_tag_type question_tag_type_id_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.question_tag_type
    ADD CONSTRAINT question_tag_type_id_pk PRIMARY KEY (question_tag_type_id);


--
-- Name: quiz_sets quiz_sets_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quiz_sets
    ADD CONSTRAINT quiz_sets_pk PRIMARY KEY (quiz_set_id);


--
-- Name: quizzes quizs_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quizzes
    ADD CONSTRAINT quizs_pk PRIMARY KEY (quiz_id);


--
-- Name: role role__pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role
    ADD CONSTRAINT role__pk PRIMARY KEY (role_id, resource_path);


--
-- Name: assessment_submission session_id_un; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assessment_submission
    ADD CONSTRAINT session_id_un UNIQUE (session_id);


--
-- Name: lo_submission shuffled_quiz_set_id_lo_submission_un; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lo_submission
    ADD CONSTRAINT shuffled_quiz_set_id_lo_submission_un UNIQUE (shuffled_quiz_set_id);


--
-- Name: exam_lo_submission shuffled_quiz_set_id_un; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.exam_lo_submission
    ADD CONSTRAINT shuffled_quiz_set_id_un UNIQUE (shuffled_quiz_set_id);


--
-- Name: shuffled_quiz_sets shuffled_quiz_sets_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shuffled_quiz_sets
    ADD CONSTRAINT shuffled_quiz_sets_pkey PRIMARY KEY (shuffled_quiz_set_id);


--
-- Name: student_latest_submissions student_latest_submissions_old_uk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_latest_submissions
    ADD CONSTRAINT student_latest_submissions_old_uk UNIQUE (student_id, study_plan_item_id, assignment_id);


--
-- Name: student_latest_submissions student_latest_submissions_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_latest_submissions
    ADD CONSTRAINT student_latest_submissions_pk PRIMARY KEY (student_id, study_plan_id, learning_material_id);


--
-- Name: student_learning_time_by_daily student_learning_time_by_daily_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_learning_time_by_daily
    ADD CONSTRAINT student_learning_time_by_daily_pk PRIMARY KEY (learning_time_id);


--
-- Name: student_learning_time_by_daily student_learning_time_by_daily_un; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_learning_time_by_daily
    ADD CONSTRAINT student_learning_time_by_daily_un UNIQUE (student_id, day);


--
-- Name: student_study_plans student_master_study_plan; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_study_plans
    ADD CONSTRAINT student_master_study_plan UNIQUE (student_id, master_study_plan_id);


--
-- Name: student_study_plans student_study_plans_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_study_plans
    ADD CONSTRAINT student_study_plans_pk PRIMARY KEY (study_plan_id, student_id);


--
-- Name: student_submission_grades student_submission_grades_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_submission_grades
    ADD CONSTRAINT student_submission_grades_pk PRIMARY KEY (student_submission_grade_id);


--
-- Name: student_submissions student_submissions_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_submissions
    ADD CONSTRAINT student_submissions_pk PRIMARY KEY (student_submission_id);


--
-- Name: students_learning_objectives_completeness students_learning_objectives_completeness_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.students_learning_objectives_completeness
    ADD CONSTRAINT students_learning_objectives_completeness_pk PRIMARY KEY (student_id, lo_id);


--
-- Name: students students_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.students
    ADD CONSTRAINT students_pk PRIMARY KEY (student_id);


--
-- Name: students_topics_completeness students_topics_completeness_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.students_topics_completeness
    ADD CONSTRAINT students_topics_completeness_pk UNIQUE (student_id, topic_id);


--
-- Name: study_plan_items study_plan_items_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.study_plan_items
    ADD CONSTRAINT study_plan_items_pk PRIMARY KEY (study_plan_item_id);


--
-- Name: study_plan_monitors study_plan_monitor_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.study_plan_monitors
    ADD CONSTRAINT study_plan_monitor_pk PRIMARY KEY (study_plan_monitor_id);


--
-- Name: study_plans study_plans_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.study_plans
    ADD CONSTRAINT study_plans_pk PRIMARY KEY (study_plan_id);


--
-- Name: feedback_session submission_id_un; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.feedback_session
    ADD CONSTRAINT submission_id_un UNIQUE (submission_id);


--
-- Name: tagged_user tagged_user_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tagged_user
    ADD CONSTRAINT tagged_user_pk PRIMARY KEY (user_id, tag_id);


--
-- Name: task_assignment task_assignment_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.task_assignment
    ADD CONSTRAINT task_assignment_pk PRIMARY KEY (learning_material_id);


--
-- Name: topics_assignments topics_assignments_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.topics_assignments
    ADD CONSTRAINT topics_assignments_pk PRIMARY KEY (topic_id, assignment_id);


--
-- Name: topics_learning_objectives topics_learning_objectives_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.topics_learning_objectives
    ADD CONSTRAINT topics_learning_objectives_pk PRIMARY KEY (topic_id, lo_id);


--
-- Name: topics topics_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.topics
    ADD CONSTRAINT topics_pk PRIMARY KEY (topic_id);


--
-- Name: assessment un_lm_course; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assessment
    ADD CONSTRAINT un_lm_course UNIQUE (learning_material_id, course_id);


--
-- Name: study_plan_assessment un_lm_sp_item; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.study_plan_assessment
    ADD CONSTRAINT un_lm_sp_item UNIQUE (learning_material_id, study_plan_item_id);


--
-- Name: user_access_paths user_access_paths_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_access_paths
    ADD CONSTRAINT user_access_paths_pk PRIMARY KEY (user_id, location_id);


--
-- Name: user_tag user_tag_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_tag
    ADD CONSTRAINT user_tag_pk PRIMARY KEY (user_tag_id);


--
-- Name: users users_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pk PRIMARY KEY (user_id);


--
-- Name: withus_failed_sync_email_recipient withus_failed_sync_email_recipient_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.withus_failed_sync_email_recipient
    ADD CONSTRAINT withus_failed_sync_email_recipient_pk PRIMARY KEY (recipient_id);


--
-- Name: withus_mapping_course_id withus_mapping_course_id_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.withus_mapping_course_id
    ADD CONSTRAINT withus_mapping_course_id_pk PRIMARY KEY (manabie_course_id);


--
-- Name: withus_mapping_exam_lo_id withus_mapping_exam_lo_id_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.withus_mapping_exam_lo_id
    ADD CONSTRAINT withus_mapping_exam_lo_id_pk PRIMARY KEY (exam_lo_id);


--
-- Name: withus_mapping_question_tag withus_mapping_question_tag_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.withus_mapping_question_tag
    ADD CONSTRAINT withus_mapping_question_tag_pk PRIMARY KEY (manabie_tag_id);


--
-- Name: assignment_name_gist_trgm_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX assignment_name_gist_trgm_idx ON public.assignment USING gist (name public.gist_trgm_ops);


--
-- Name: assignment_topic_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX assignment_topic_id_idx ON public.assignment USING btree (topic_id);


--
-- Name: assignments_topic_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX assignments_topic_id_idx ON public.assignments USING btree (topic_id);


--
-- Name: book_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX book_idx ON public.study_plans USING btree (book_id);


--
-- Name: books_chapters_book_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX books_chapters_book_id_idx ON public.books_chapters USING btree (book_id);


--
-- Name: books_chapters_chapter_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX books_chapters_chapter_id_idx ON public.books_chapters USING btree (chapter_id);


--
-- Name: chapters_book_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX chapters_book_id_idx ON public.chapters USING btree (book_id);


--
-- Name: content_bank_medias_name_resource_path_unique_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX content_bank_medias_name_resource_path_unique_idx ON public.content_bank_medias USING btree (name, resource_path) WHERE (deleted_at IS NULL);


--
-- Name: copy_study_plan_item_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX copy_study_plan_item_id_idx ON public.study_plan_items USING btree (copy_study_plan_item_id);


--
-- Name: course_students_access_paths_course_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX course_students_access_paths_course_id_idx ON public.course_students_access_paths USING btree (course_id);


--
-- Name: course_students_access_paths_course_id_student_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX course_students_access_paths_course_id_student_id_idx ON public.course_students_access_paths USING btree (course_id, student_id);


--
-- Name: course_students_access_paths_location_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX course_students_access_paths_location_id_idx ON public.course_students_access_paths USING btree (location_id);


--
-- Name: course_students_access_paths_student_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX course_students_access_paths_student_id_idx ON public.course_students_access_paths USING btree (student_id);


--
-- Name: course_students_course_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX course_students_course_id_idx ON public.course_students USING btree (course_id);


--
-- Name: course_students_student_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX course_students_student_id_idx ON public.course_students USING btree (student_id);


--
-- Name: course_study_plans_course_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX course_study_plans_course_id_idx ON public.course_study_plans USING btree (course_id);


--
-- Name: course_study_plans_study_plan_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX course_study_plans_study_plan_id_idx ON public.course_study_plans USING btree (study_plan_id);


--
-- Name: courses_books_course_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX courses_books_course_id_idx ON public.courses_books USING btree (course_id);


--
-- Name: event_logs_student_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX event_logs_student_id_idx ON public.student_event_logs USING btree (student_id);


--
-- Name: exam_lo_name_gin_trgm_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX exam_lo_name_gin_trgm_idx ON public.exam_lo USING gin (name public.gin_trgm_ops);


--
-- Name: exam_lo_submission_answer_submission_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX exam_lo_submission_answer_submission_id_idx ON public.exam_lo_submission_answer USING btree (submission_id);


--
-- Name: exam_lo_submission_status_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX exam_lo_submission_status_idx ON public.exam_lo_submission USING btree (status);


--
-- Name: exam_lo_submission_student_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX exam_lo_submission_student_id_idx ON public.exam_lo_submission USING btree (student_id);


--
-- Name: exam_lo_submission_study_plan_item_identity_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX exam_lo_submission_study_plan_item_identity_idx ON public.exam_lo_submission USING btree (study_plan_id, student_id, learning_material_id);


--
-- Name: exam_lo_topic_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX exam_lo_topic_id_idx ON public.exam_lo USING btree (topic_id);


--
-- Name: flash_card_name_gist_trgm_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX flash_card_name_gist_trgm_idx ON public.flash_card USING gist (name public.gist_trgm_ops);


--
-- Name: flash_card_submission_answer_study_plan_item_identity_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX flash_card_submission_answer_study_plan_item_identity_idx ON public.flash_card_submission_answer USING btree (student_id, study_plan_id, learning_material_id);


--
-- Name: flash_card_topic_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX flash_card_topic_id_idx ON public.flash_card USING btree (topic_id);


--
-- Name: granted_permission__permission_name__idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX granted_permission__permission_name__idx ON public.granted_permission USING btree (permission_name);


--
-- Name: granted_permission__role_name__idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX granted_permission__role_name__idx ON public.granted_permission USING btree (role_name);


--
-- Name: granted_permission__user_group_id__idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX granted_permission__user_group_id__idx ON public.granted_permission USING btree (user_group_id);


--
-- Name: individual_study_plan_study_plan_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX individual_study_plan_study_plan_id_idx ON public.individual_study_plan USING btree (study_plan_id);


--
-- Name: latest_session_by_identity_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX latest_session_by_identity_idx ON public.assessment_session USING btree (assessment_id, user_id, created_at DESC);


--
-- Name: learning_material_name_gist_trgm_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX learning_material_name_gist_trgm_idx ON public.learning_material USING gist (name public.gist_trgm_ops);


--
-- Name: learning_material_topic_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX learning_material_topic_id_idx ON public.learning_material USING btree (topic_id);


--
-- Name: learning_objective_name_gist_trgm_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX learning_objective_name_gist_trgm_idx ON public.learning_objective USING gist (name public.gist_trgm_ops);


--
-- Name: learning_objective_topic_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX learning_objective_topic_id_idx ON public.learning_objective USING btree (topic_id);


--
-- Name: learning_objectives_name_idx_gin_trgm; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX learning_objectives_name_idx_gin_trgm ON public.learning_objectives USING gin (name public.gin_trgm_ops);


--
-- Name: learning_objectives_topic_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX learning_objectives_topic_id_idx ON public.learning_objectives USING btree (topic_id);


--
-- Name: learning_objectives_type_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX learning_objectives_type_idx ON public.learning_objectives USING btree (type);


--
-- Name: lo_progression_study_plan_item_identity_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX lo_progression_study_plan_item_identity_idx ON public.lo_progression USING btree (student_id, study_plan_id, learning_material_id);


--
-- Name: lo_progression_study_plan_item_identity_un; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX lo_progression_study_plan_item_identity_un ON public.lo_progression USING btree (student_id, study_plan_id, learning_material_id) WHERE (deleted_at IS NULL);


--
-- Name: lo_study_plan_items_lo_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX lo_study_plan_items_lo_id_idx ON public.lo_study_plan_items USING btree (lo_id);


--
-- Name: lo_submission_answer_study_plan_item_identity_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX lo_submission_answer_study_plan_item_identity_idx ON public.lo_submission_answer USING btree (student_id, study_plan_id, learning_material_id);


--
-- Name: master_study_plan_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX master_study_plan_id_idx ON public.study_plans USING btree (master_study_plan_id);


--
-- Name: master_study_plan_item_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX master_study_plan_item_id_idx ON public.study_plans USING btree (master_study_plan_id);


--
-- Name: master_study_plan_lm_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX master_study_plan_lm_id_idx ON public.master_study_plan USING btree (learning_material_id);


--
-- Name: master_study_plan_study_plan_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX master_study_plan_study_plan_id_idx ON public.master_study_plan USING btree (study_plan_id);


--
-- Name: question_group_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX question_group_id_idx ON public.quizzes USING hash (question_group_id) WHERE ((question_group_id IS NOT NULL) AND (deleted_at IS NULL));


--
-- Name: quiz_sets_approved_lo_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX quiz_sets_approved_lo_id_idx ON public.quiz_sets USING btree (lo_id) WHERE ((status = 'QUIZSET_STATUS_APPROVED'::text) AND (deleted_at IS NULL));


--
-- Name: quiz_sets_lo_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX quiz_sets_lo_id_idx ON public.quiz_sets USING btree (lo_id);


--
-- Name: quizzes_external_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX quizzes_external_id_idx ON public.quizzes USING btree (external_id);


--
-- Name: shuffled_quiz_original_quiz_set_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX shuffled_quiz_original_quiz_set_id_idx ON public.shuffled_quiz_sets USING btree (original_quiz_set_id);


--
-- Name: shuffled_quiz_session_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX shuffled_quiz_session_id_idx ON public.shuffled_quiz_sets USING btree (session_id);


--
-- Name: shuffled_quiz_sets_learning_material_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX shuffled_quiz_sets_learning_material_idx ON public.shuffled_quiz_sets USING btree (learning_material_id);


--
-- Name: shuffled_quiz_sets_study_plan_item_identity_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX shuffled_quiz_sets_study_plan_item_identity_idx ON public.shuffled_quiz_sets USING btree (student_id, study_plan_id, learning_material_id);


--
-- Name: shuffled_quiz_sets_study_plan_item_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX shuffled_quiz_sets_study_plan_item_idx ON public.shuffled_quiz_sets USING btree (study_plan_item_id);


--
-- Name: student_event_logs_event_type_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_event_logs_event_type_idx ON public.student_event_logs USING btree (event_type);


--
-- Name: student_event_logs_learning_material_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_event_logs_learning_material_id_idx ON public.student_event_logs USING btree (learning_material_id);


--
-- Name: student_event_logs_payload_session_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_event_logs_payload_session_id_idx ON public.student_event_logs USING btree (((payload ->> 'session_id'::text)));


--
-- Name: student_event_logs_payload_study_plan_item_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_event_logs_payload_study_plan_item_id_idx ON public.student_event_logs USING btree (((payload ->> 'study_plan_item_id'::text))) WHERE ((payload ->> 'study_plan_item_id'::text) IS NOT NULL);


--
-- Name: student_event_logs_study_plan_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_event_logs_study_plan_id_idx ON public.student_event_logs USING btree (study_plan_id);


--
-- Name: student_event_logs_study_plan_item_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_event_logs_study_plan_item_id_idx ON public.student_event_logs USING btree (study_plan_item_id);


--
-- Name: student_event_logs_study_plan_item_identity_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_event_logs_study_plan_item_identity_idx ON public.student_event_logs USING btree (study_plan_id, learning_material_id, student_id);


--
-- Name: student_latest_submissions_student_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_latest_submissions_student_id_idx ON public.student_latest_submissions USING btree (student_id);


--
-- Name: student_latest_submissions_submission_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX student_latest_submissions_submission_id ON public.student_latest_submissions USING btree (student_submission_id DESC);


--
-- Name: student_learning_time_by_daily_student_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_learning_time_by_daily_student_id_idx ON public.student_learning_time_by_daily USING btree (student_id);


--
-- Name: student_study_plans_master_study_plan_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_study_plans_master_study_plan_id_idx ON public.student_study_plans USING btree (master_study_plan_id);


--
-- Name: student_study_plans_student_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_study_plans_student_id_idx ON public.student_study_plans USING btree (student_id);


--
-- Name: student_submission_grades_student_submission_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_submission_grades_student_submission_id ON public.student_submission_grades USING btree (student_submission_id);


--
-- Name: student_submissions_learning_material_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_submissions_learning_material_id_idx ON public.student_submissions USING btree (learning_material_id);


--
-- Name: student_submissions_student_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_submissions_student_id_idx ON public.student_submissions USING btree (student_id);


--
-- Name: student_submissions_student_submission_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_submissions_student_submission_id_idx ON public.student_submissions USING btree (student_submission_id);


--
-- Name: student_submissions_study_plan_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_submissions_study_plan_id_idx ON public.student_submissions USING btree (study_plan_id);


--
-- Name: student_submissions_study_plan_item_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_submissions_study_plan_item_id_idx ON public.student_submissions USING btree (study_plan_item_id);


--
-- Name: student_submissions_study_plan_item_identity_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX student_submissions_study_plan_item_identity_idx ON public.student_submissions USING btree (study_plan_id, learning_material_id, student_id);


--
-- Name: students_current_grade_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX students_current_grade_idx ON public.students USING btree (current_grade);


--
-- Name: study_plan_content_structure_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX study_plan_content_structure_idx ON public.study_plan_items USING btree (study_plan_id, content_structure_flatten);


--
-- Name: study_plan_items_study_plan_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX study_plan_items_study_plan_id_idx ON public.study_plan_items USING btree (study_plan_id);


--
-- Name: study_plans_course_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX study_plans_course_id_idx ON public.study_plans USING btree (course_id);


--
-- Name: task_assignment_name_gist_trgm_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX task_assignment_name_gist_trgm_idx ON public.task_assignment USING gist (name public.gist_trgm_ops);


--
-- Name: task_assignment_topic_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX task_assignment_topic_id_idx ON public.task_assignment USING btree (topic_id);


--
-- Name: topic_assignments; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX topic_assignments ON public.assignments USING btree (((content ->> 'topic_id'::text)));


--
-- Name: topics_assignments_assignment_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX topics_assignments_assignment_id_idx ON public.topics_assignments USING btree (assignment_id);


--
-- Name: topics_chapter_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX topics_chapter_id_idx ON public.topics USING btree (chapter_id);


--
-- Name: topics_learning_objectives_lo_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX topics_learning_objectives_lo_id_idx ON public.topics_learning_objectives USING btree (lo_id);


--
-- Name: user_access_paths__location_id__idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX user_access_paths__location_id__idx ON public.user_access_paths USING btree (location_id);


--
-- Name: user_access_paths__user_id__idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX user_access_paths__user_id__idx ON public.user_access_paths USING btree (user_id);


--
-- Name: user_group_member_user_group_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX user_group_member_user_group_idx ON public.user_group_member USING btree (user_group_id);


--
-- Name: user_name_gin_trgm_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX user_name_gin_trgm_idx ON public.users USING gin (name public.gin_trgm_ops);


--
-- Name: assignments create_assignment; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER create_assignment AFTER INSERT OR UPDATE ON public.assignments FOR EACH ROW EXECUTE PROCEDURE public.create_assignment_fn();


--
-- Name: learning_objectives create_flash_card; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER create_flash_card AFTER INSERT OR UPDATE ON public.learning_objectives FOR EACH ROW EXECUTE PROCEDURE public.create_flash_card_fn();


--
-- Name: learning_objectives create_learning_objective; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER create_learning_objective AFTER INSERT OR UPDATE ON public.learning_objectives FOR EACH ROW EXECUTE PROCEDURE public.create_learning_objective_fn();


--
-- Name: assignments create_task_assignment; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER create_task_assignment AFTER INSERT OR UPDATE ON public.assignments FOR EACH ROW EXECUTE PROCEDURE public.create_task_assignment_fn();


--
-- Name: student_event_logs fill_new_identity; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER fill_new_identity AFTER INSERT ON public.student_event_logs FOR EACH ROW EXECUTE PROCEDURE public.trigger_student_event_logs_fill_new_identity_fn();


--
-- Name: student_latest_submissions fill_new_identity; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER fill_new_identity BEFORE INSERT ON public.student_latest_submissions FOR EACH ROW EXECUTE PROCEDURE public.trigger_student_latest_submissions_fill_new_identity_fn();


--
-- Name: student_submissions fill_new_identity; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER fill_new_identity AFTER INSERT OR UPDATE ON public.student_submissions FOR EACH ROW EXECUTE PROCEDURE public.trigger_student_submissions_fill_new_identity_fn();


--
-- Name: learning_objectives migrate_to_exam_lo; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER migrate_to_exam_lo AFTER INSERT OR UPDATE ON public.learning_objectives FOR EACH ROW WHEN ((new.type = 'LEARNING_OBJECTIVE_TYPE_EXAM_LO'::text)) EXECUTE PROCEDURE public.migrate_learning_objectives_to_exam_lo_fn();


--
-- Name: shuffled_quiz_sets migrate_to_exam_lo_submission_and_answer_once_submitted; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER migrate_to_exam_lo_submission_and_answer_once_submitted AFTER UPDATE OF updated_at ON public.shuffled_quiz_sets FOR EACH ROW EXECUTE PROCEDURE public.migrate_to_exam_lo_submission_and_answer_once_submitted();


--
-- Name: shuffled_quiz_sets migrate_to_flash_card_submission_and_flash_card_submission_answ; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER migrate_to_flash_card_submission_and_flash_card_submission_answ AFTER UPDATE OF updated_at ON public.shuffled_quiz_sets FOR EACH ROW EXECUTE PROCEDURE public.migrate_to_flash_card_submission_and_flash_card_submission_answ();


--
-- Name: shuffled_quiz_sets migrate_to_lo_submission_and_answer; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER migrate_to_lo_submission_and_answer AFTER UPDATE OF updated_at ON public.shuffled_quiz_sets FOR EACH ROW EXECUTE PROCEDURE public.migrate_to_lo_submission_and_answer_fnc();


--
-- Name: study_plan_items migrate_to_master_study_plan_ins; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER migrate_to_master_study_plan_ins AFTER INSERT ON public.study_plan_items REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE PROCEDURE public.migrate_study_plan_items_to_master_study_plan_fn();


--
-- Name: study_plan_items migrate_to_master_study_plan_udt; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER migrate_to_master_study_plan_udt AFTER UPDATE ON public.study_plan_items REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE PROCEDURE public.migrate_study_plan_items_to_master_study_plan_fn();


--
-- Name: course_students migrate_withus_mapping_course_id; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER migrate_withus_mapping_course_id AFTER INSERT ON public.course_students FOR EACH ROW EXECUTE PROCEDURE public.migrate_withus_mapping_course_id();


--
-- Name: exam_lo migrate_withus_mapping_exam_lo_id; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER migrate_withus_mapping_exam_lo_id AFTER INSERT ON public.exam_lo FOR EACH ROW EXECUTE PROCEDURE public.migrate_withus_mapping_exam_lo_id();


--
-- Name: question_tag migrate_withus_mapping_question_tag; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER migrate_withus_mapping_question_tag AFTER INSERT OR UPDATE ON public.question_tag FOR EACH ROW EXECUTE PROCEDURE public.migrate_withus_mapping_question_tag();


--
-- Name: study_plan_items trigger_study_plan_items_to_individual_study_plan_ins; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trigger_study_plan_items_to_individual_study_plan_ins AFTER INSERT ON public.study_plan_items REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE PROCEDURE public.trigger_study_plan_items_to_individual_study_plan();


--
-- Name: study_plan_items trigger_study_plan_items_to_individual_study_plan_udt; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trigger_study_plan_items_to_individual_study_plan_udt AFTER UPDATE ON public.study_plan_items REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE PROCEDURE public.trigger_study_plan_items_to_individual_study_plan();


--
-- Name: exam_lo_submission update_allocate_marker_once_created; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_allocate_marker_once_created AFTER INSERT ON public.exam_lo_submission FOR EACH ROW EXECUTE PROCEDURE public.update_allocate_marker_when_exam_lo_submission_was_created();


--
-- Name: books_chapters update_book_id_for_chapters; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_book_id_for_chapters AFTER INSERT OR UPDATE ON public.books_chapters FOR EACH ROW EXECUTE PROCEDURE public.update_book_id_for_chapters_fn();


--
-- Name: study_plan_items update_content_structure; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_content_structure BEFORE INSERT OR UPDATE ON public.study_plan_items FOR EACH ROW EXECUTE PROCEDURE public.update_content_structure_fnc();


--
-- Name: assignment_study_plan_items update_content_structure_flatten; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_content_structure_flatten AFTER INSERT ON public.assignment_study_plan_items FOR EACH ROW EXECUTE PROCEDURE public.update_content_structure_flatten_on_assignment_study_plan_item_();


--
-- Name: lo_study_plan_items update_content_structure_flatten; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_content_structure_flatten AFTER INSERT ON public.lo_study_plan_items FOR EACH ROW EXECUTE PROCEDURE public.update_content_structure_flatten_fn();


--
-- Name: student_study_plans update_master_study_plan_student_study_plan; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_master_study_plan_student_study_plan AFTER INSERT ON public.student_study_plans FOR EACH ROW EXECUTE PROCEDURE public.update_master_study_plan_id_on_student_study_plan_created_fn();


--
-- Name: exam_lo_submission update_max_score_exam_lo_once_exam_lo_submission_status_change; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_max_score_exam_lo_once_exam_lo_submission_status_change AFTER UPDATE OF status, deleted_at ON public.exam_lo_submission FOR EACH ROW EXECUTE PROCEDURE public.update_max_score_exam_lo_once_exam_lo_submission_status_change();


--
-- Name: flashcard_progressions update_study_plan_item_identity_for_flashcard_progression; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_study_plan_item_identity_for_flashcard_progression AFTER INSERT ON public.flashcard_progressions FOR EACH ROW EXECUTE PROCEDURE public.update_study_plan_item_identity_for_flashcard_progression_fn();


--
-- Name: shuffled_quiz_sets update_study_plan_item_identity_for_shuffled_quiz_set; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_study_plan_item_identity_for_shuffled_quiz_set AFTER INSERT ON public.shuffled_quiz_sets FOR EACH ROW EXECUTE PROCEDURE public.update_study_plan_item_identity_for_shuffled_quiz_set_fn();


--
-- Name: withus_mapping_course_id withus_check_valid_course_id; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER withus_check_valid_course_id BEFORE INSERT OR UPDATE ON public.withus_mapping_course_id FOR EACH ROW EXECUTE PROCEDURE public.withus_check_valid_course_id();


--
-- Name: assessment_submission assessment_submission_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assessment_submission
    ADD CONSTRAINT assessment_submission_fk FOREIGN KEY (session_id) REFERENCES public.assessment_session(session_id);


--
-- Name: assignment_study_plan_items assignment_study_plan_items_assignment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assignment_study_plan_items
    ADD CONSTRAINT assignment_study_plan_items_assignment_id_fkey FOREIGN KEY (assignment_id) REFERENCES public.assignments(assignment_id);


--
-- Name: assignment_study_plan_items assignment_study_plan_items_study_plan_item_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assignment_study_plan_items
    ADD CONSTRAINT assignment_study_plan_items_study_plan_item_id_fkey FOREIGN KEY (study_plan_item_id) REFERENCES public.study_plan_items(study_plan_item_id);


--
-- Name: assignment assignment_topic_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assignment
    ADD CONSTRAINT assignment_topic_id_fk FOREIGN KEY (topic_id) REFERENCES public.topics(topic_id);


--
-- Name: class_study_plans class_study_plans_study_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.class_study_plans
    ADD CONSTRAINT class_study_plans_study_plan_id_fkey FOREIGN KEY (study_plan_id) REFERENCES public.study_plans(study_plan_id);


--
-- Name: course_students_access_paths course_students_access_paths_course_students_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.course_students_access_paths
    ADD CONSTRAINT course_students_access_paths_course_students_fk FOREIGN KEY (course_student_id) REFERENCES public.course_students(course_student_id);


--
-- Name: course_student_subscriptions course_students_subscriptions_course_students_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.course_student_subscriptions
    ADD CONSTRAINT course_students_subscriptions_course_students_fk FOREIGN KEY (course_student_id) REFERENCES public.course_students(course_student_id);


--
-- Name: course_study_plans course_study_plans_study_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.course_study_plans
    ADD CONSTRAINT course_study_plans_study_plan_id_fkey FOREIGN KEY (study_plan_id) REFERENCES public.study_plans(study_plan_id);


--
-- Name: withus_mapping_exam_lo_id exam_lo_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.withus_mapping_exam_lo_id
    ADD CONSTRAINT exam_lo_id_fk FOREIGN KEY (exam_lo_id) REFERENCES public.exam_lo(learning_material_id);


--
-- Name: exam_lo_submission_answer exam_lo_submission_answer_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.exam_lo_submission_answer
    ADD CONSTRAINT exam_lo_submission_answer_fk FOREIGN KEY (submission_id) REFERENCES public.exam_lo_submission(submission_id);


--
-- Name: exam_lo_submission_score exam_lo_submission_score_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.exam_lo_submission_score
    ADD CONSTRAINT exam_lo_submission_score_fk FOREIGN KEY (submission_id) REFERENCES public.exam_lo_submission(submission_id);


--
-- Name: exam_lo exam_lo_topic_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.exam_lo
    ADD CONSTRAINT exam_lo_topic_id_fk FOREIGN KEY (topic_id) REFERENCES public.topics(topic_id);


--
-- Name: feedback_session feedback_session_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.feedback_session
    ADD CONSTRAINT feedback_session_fk FOREIGN KEY (submission_id) REFERENCES public.assessment_submission(id);


--
-- Name: assessment fk__course_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assessment
    ADD CONSTRAINT fk__course_id FOREIGN KEY (course_id) REFERENCES public.courses(course_id);


--
-- Name: study_plan_assessment fk__learning_material_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.study_plan_assessment
    ADD CONSTRAINT fk__learning_material_id FOREIGN KEY (learning_material_id) REFERENCES public.learning_material(learning_material_id);


--
-- Name: study_plan_assessment fk__sp_item_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.study_plan_assessment
    ADD CONSTRAINT fk__sp_item_id FOREIGN KEY (study_plan_item_id) REFERENCES public.lms_study_plan_items(study_plan_item_id);


--
-- Name: assessment_session fk_assessment_assessment_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assessment_session
    ADD CONSTRAINT fk_assessment_assessment_id FOREIGN KEY (assessment_id) REFERENCES public.assessment(id);


--
-- Name: courses fk_cerebry_class_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.courses
    ADD CONSTRAINT fk_cerebry_class_id FOREIGN KEY (vendor_id) REFERENCES public.cerebry_classes(id);


--
-- Name: lms_student_study_plan_item fk_lm_list_id_lms_student_study_plan_item; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lms_student_study_plan_item
    ADD CONSTRAINT fk_lm_list_id_lms_student_study_plan_item FOREIGN KEY (lm_list_id) REFERENCES public.lms_learning_material_list(lm_list_id);


--
-- Name: lms_study_plan_items fk_lm_list_id_lms_study_plan_items; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lms_study_plan_items
    ADD CONSTRAINT fk_lm_list_id_lms_study_plan_items FOREIGN KEY (lm_list_id) REFERENCES public.lms_learning_material_list(lm_list_id);


--
-- Name: assessment_session fk_sp_assessment_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.assessment_session
    ADD CONSTRAINT fk_sp_assessment_id FOREIGN KEY (study_plan_assessment_id) REFERENCES public.study_plan_assessment(id);


--
-- Name: lms_student_study_plan_item fk_study_plan_id_lms_student_study_plan_item; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lms_student_study_plan_item
    ADD CONSTRAINT fk_study_plan_id_lms_student_study_plan_item FOREIGN KEY (study_plan_id) REFERENCES public.lms_study_plans(study_plan_id);


--
-- Name: lms_student_study_plans fk_study_plan_id_lms_student_study_plans; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lms_student_study_plans
    ADD CONSTRAINT fk_study_plan_id_lms_student_study_plans FOREIGN KEY (study_plan_id) REFERENCES public.lms_study_plans(study_plan_id);


--
-- Name: lms_study_plan_items fk_study_plan_id_lms_study_plan_items; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lms_study_plan_items
    ADD CONSTRAINT fk_study_plan_id_lms_study_plan_items FOREIGN KEY (study_plan_id) REFERENCES public.lms_study_plans(study_plan_id);


--
-- Name: flash_card_submission_answer flash_card_submission_answer_flash_card_submission_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.flash_card_submission_answer
    ADD CONSTRAINT flash_card_submission_answer_flash_card_submission_fk FOREIGN KEY (submission_id) REFERENCES public.flash_card_submission(submission_id);


--
-- Name: flash_card_submission flash_card_submission_flash_card_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.flash_card_submission
    ADD CONSTRAINT flash_card_submission_flash_card_fk FOREIGN KEY (learning_material_id) REFERENCES public.flash_card(learning_material_id);


--
-- Name: flash_card_submission flash_card_submission_shuffled_quiz_sets_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.flash_card_submission
    ADD CONSTRAINT flash_card_submission_shuffled_quiz_sets_fk FOREIGN KEY (shuffled_quiz_set_id) REFERENCES public.shuffled_quiz_sets(shuffled_quiz_set_id);


--
-- Name: flash_card_submission flash_card_submission_study_plans_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.flash_card_submission
    ADD CONSTRAINT flash_card_submission_study_plans_fk FOREIGN KEY (study_plan_id) REFERENCES public.study_plans(study_plan_id);


--
-- Name: lo_progression_answer lo_progression_answer_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lo_progression_answer
    ADD CONSTRAINT lo_progression_answer_fk FOREIGN KEY (progression_id) REFERENCES public.lo_progression(progression_id);


--
-- Name: lo_study_plan_items lo_study_plan_items_study_plan_item_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lo_study_plan_items
    ADD CONSTRAINT lo_study_plan_items_study_plan_item_id_fkey FOREIGN KEY (study_plan_item_id) REFERENCES public.study_plan_items(study_plan_item_id);


--
-- Name: lo_submission_answer lo_submission_answer_learning_objective_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lo_submission_answer
    ADD CONSTRAINT lo_submission_answer_learning_objective_fk FOREIGN KEY (learning_material_id) REFERENCES public.learning_objective(learning_material_id);


--
-- Name: lo_submission_answer lo_submission_answer_shuffled_quiz_sets_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lo_submission_answer
    ADD CONSTRAINT lo_submission_answer_shuffled_quiz_sets_fk FOREIGN KEY (shuffled_quiz_set_id) REFERENCES public.shuffled_quiz_sets(shuffled_quiz_set_id);


--
-- Name: lo_submission_answer lo_submission_answer_study_plans_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lo_submission_answer
    ADD CONSTRAINT lo_submission_answer_study_plans_fk FOREIGN KEY (study_plan_id) REFERENCES public.study_plans(study_plan_id);


--
-- Name: lo_submission lo_submission_learning_objective_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lo_submission
    ADD CONSTRAINT lo_submission_learning_objective_fk FOREIGN KEY (learning_material_id) REFERENCES public.learning_objective(learning_material_id);


--
-- Name: lo_submission lo_submission_shuffled_quiz_sets_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lo_submission
    ADD CONSTRAINT lo_submission_shuffled_quiz_sets_fk FOREIGN KEY (shuffled_quiz_set_id) REFERENCES public.shuffled_quiz_sets(shuffled_quiz_set_id);


--
-- Name: lo_submission lo_submission_study_plans_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lo_submission
    ADD CONSTRAINT lo_submission_study_plans_fk FOREIGN KEY (study_plan_id) REFERENCES public.study_plans(study_plan_id);


--
-- Name: withus_mapping_question_tag question_tag_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.withus_mapping_question_tag
    ADD CONSTRAINT question_tag_id_fk FOREIGN KEY (manabie_tag_id) REFERENCES public.question_tag(question_tag_id);


--
-- Name: question_tag question_tag_type_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.question_tag
    ADD CONSTRAINT question_tag_type_id_fk FOREIGN KEY (question_tag_type_id) REFERENCES public.question_tag_type(question_tag_type_id);


--
-- Name: student_latest_submissions student_latest_submission_assigment_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_latest_submissions
    ADD CONSTRAINT student_latest_submission_assigment_fk FOREIGN KEY (assignment_id) REFERENCES public.assignments(assignment_id);


--
-- Name: student_latest_submissions student_latest_submission_study_plan_item_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_latest_submissions
    ADD CONSTRAINT student_latest_submission_study_plan_item_fk FOREIGN KEY (study_plan_item_id) REFERENCES public.study_plan_items(study_plan_item_id);


--
-- Name: student_study_plans student_study_plans_study_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_study_plans
    ADD CONSTRAINT student_study_plans_study_plan_id_fkey FOREIGN KEY (study_plan_id) REFERENCES public.study_plans(study_plan_id);


--
-- Name: student_submissions student_submission_assigment_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_submissions
    ADD CONSTRAINT student_submission_assigment_fk FOREIGN KEY (assignment_id) REFERENCES public.assignments(assignment_id);


--
-- Name: student_submission_grades student_submission_grades_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_submission_grades
    ADD CONSTRAINT student_submission_grades_fk FOREIGN KEY (student_submission_id) REFERENCES public.student_submissions(student_submission_id);


--
-- Name: student_submissions student_submission_study_plan_item_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_submissions
    ADD CONSTRAINT student_submission_study_plan_item_fk FOREIGN KEY (study_plan_item_id) REFERENCES public.study_plan_items(study_plan_item_id);


--
-- Name: student_submissions student_submissions_grades_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_submissions
    ADD CONSTRAINT student_submissions_grades_fk FOREIGN KEY (student_submission_grade_id) REFERENCES public.student_submission_grades(student_submission_grade_id);


--
-- Name: master_study_plan study_plan_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.master_study_plan
    ADD CONSTRAINT study_plan_id_fk FOREIGN KEY (study_plan_id) REFERENCES public.study_plans(study_plan_id);


--
-- Name: individual_study_plan study_plan_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.individual_study_plan
    ADD CONSTRAINT study_plan_id_fk FOREIGN KEY (study_plan_id) REFERENCES public.study_plans(study_plan_id);


--
-- Name: study_plan_items study_plan_items_study_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.study_plan_items
    ADD CONSTRAINT study_plan_items_study_plan_id_fkey FOREIGN KEY (study_plan_id) REFERENCES public.study_plans(study_plan_id);


--
-- Name: study_plans study_plans_master_study_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.study_plans
    ADD CONSTRAINT study_plans_master_study_plan_id_fkey FOREIGN KEY (master_study_plan_id) REFERENCES public.study_plans(study_plan_id);


--
-- Name: task_assignment task_assignment_topic_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.task_assignment
    ADD CONSTRAINT task_assignment_topic_id_fk FOREIGN KEY (topic_id) REFERENCES public.topics(topic_id);


--
-- Name: learning_material topic_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.learning_material
    ADD CONSTRAINT topic_id_fk FOREIGN KEY (topic_id) REFERENCES public.topics(topic_id);


--
-- Name: learning_objective topic_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.learning_objective
    ADD CONSTRAINT topic_id_fk FOREIGN KEY (topic_id) REFERENCES public.topics(topic_id);


--
-- Name: flash_card topic_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.flash_card
    ADD CONSTRAINT topic_id_fk FOREIGN KEY (topic_id) REFERENCES public.topics(topic_id);


--
-- Name: topics_assignments topics_assignments_assignment_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.topics_assignments
    ADD CONSTRAINT topics_assignments_assignment_fk FOREIGN KEY (assignment_id) REFERENCES public.assignments(assignment_id);


--
-- Name: academic_year; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.academic_year ENABLE ROW LEVEL SECURITY;

--
-- Name: assessment; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.assessment ENABLE ROW LEVEL SECURITY;

--
-- Name: assessment_session; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.assessment_session ENABLE ROW LEVEL SECURITY;

--
-- Name: assessment_submission; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.assessment_submission ENABLE ROW LEVEL SECURITY;

--
-- Name: assign_study_plan_tasks; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.assign_study_plan_tasks ENABLE ROW LEVEL SECURITY;

--
-- Name: assignment; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.assignment ENABLE ROW LEVEL SECURITY;

--
-- Name: assignment_study_plan_items; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.assignment_study_plan_items ENABLE ROW LEVEL SECURITY;

--
-- Name: assignments; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.assignments ENABLE ROW LEVEL SECURITY;

--
-- Name: books; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.books ENABLE ROW LEVEL SECURITY;

--
-- Name: books_chapters; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.books_chapters ENABLE ROW LEVEL SECURITY;

--
-- Name: cerebry_classes; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.cerebry_classes ENABLE ROW LEVEL SECURITY;

--
-- Name: chapters; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.chapters ENABLE ROW LEVEL SECURITY;

--
-- Name: class_students; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.class_students ENABLE ROW LEVEL SECURITY;

--
-- Name: class_study_plans; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.class_study_plans ENABLE ROW LEVEL SECURITY;

--
-- Name: content_bank_medias; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.content_bank_medias ENABLE ROW LEVEL SECURITY;

--
-- Name: course_classes; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.course_classes ENABLE ROW LEVEL SECURITY;

--
-- Name: course_student_subscriptions; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.course_student_subscriptions ENABLE ROW LEVEL SECURITY;

--
-- Name: course_students; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.course_students ENABLE ROW LEVEL SECURITY;

--
-- Name: course_students_access_paths; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.course_students_access_paths ENABLE ROW LEVEL SECURITY;

--
-- Name: course_study_plans; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.course_study_plans ENABLE ROW LEVEL SECURITY;

--
-- Name: courses; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.courses ENABLE ROW LEVEL SECURITY;

--
-- Name: courses_books; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.courses_books ENABLE ROW LEVEL SECURITY;

--
-- Name: exam_lo; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.exam_lo ENABLE ROW LEVEL SECURITY;

--
-- Name: exam_lo_submission; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.exam_lo_submission ENABLE ROW LEVEL SECURITY;

--
-- Name: exam_lo_submission_answer; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.exam_lo_submission_answer ENABLE ROW LEVEL SECURITY;

--
-- Name: exam_lo_submission_score; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.exam_lo_submission_score ENABLE ROW LEVEL SECURITY;

--
-- Name: feedback_session; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.feedback_session ENABLE ROW LEVEL SECURITY;

--
-- Name: flash_card; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.flash_card ENABLE ROW LEVEL SECURITY;

--
-- Name: flash_card_submission; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.flash_card_submission ENABLE ROW LEVEL SECURITY;

--
-- Name: flash_card_submission_answer; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.flash_card_submission_answer ENABLE ROW LEVEL SECURITY;

--
-- Name: flashcard_progressions; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.flashcard_progressions ENABLE ROW LEVEL SECURITY;

--
-- Name: flashcard_speeches; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.flashcard_speeches ENABLE ROW LEVEL SECURITY;

--
-- Name: grade; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.grade ENABLE ROW LEVEL SECURITY;

--
-- Name: granted_permission; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.granted_permission ENABLE ROW LEVEL SECURITY;

--
-- Name: granted_role; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.granted_role ENABLE ROW LEVEL SECURITY;

--
-- Name: granted_role_access_path; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.granted_role_access_path ENABLE ROW LEVEL SECURITY;

--
-- Name: groups; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.groups ENABLE ROW LEVEL SECURITY;

--
-- Name: import_study_plan_task; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.import_study_plan_task ENABLE ROW LEVEL SECURITY;

--
-- Name: individual_study_plan; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.individual_study_plan ENABLE ROW LEVEL SECURITY;

--
-- Name: learning_material; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.learning_material ENABLE ROW LEVEL SECURITY;

--
-- Name: learning_objective; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.learning_objective ENABLE ROW LEVEL SECURITY;

--
-- Name: learning_objectives; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.learning_objectives ENABLE ROW LEVEL SECURITY;

--
-- Name: lms_learning_material_list; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.lms_learning_material_list ENABLE ROW LEVEL SECURITY;

--
-- Name: lms_student_study_plan_item; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.lms_student_study_plan_item ENABLE ROW LEVEL SECURITY;

--
-- Name: lms_student_study_plans; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.lms_student_study_plans ENABLE ROW LEVEL SECURITY;

--
-- Name: lms_study_plan_items; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.lms_study_plan_items ENABLE ROW LEVEL SECURITY;

--
-- Name: lms_study_plans; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.lms_study_plans ENABLE ROW LEVEL SECURITY;

--
-- Name: lo_progression; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.lo_progression ENABLE ROW LEVEL SECURITY;

--
-- Name: lo_progression_answer; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.lo_progression_answer ENABLE ROW LEVEL SECURITY;

--
-- Name: lo_study_plan_items; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.lo_study_plan_items ENABLE ROW LEVEL SECURITY;

--
-- Name: lo_submission; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.lo_submission ENABLE ROW LEVEL SECURITY;

--
-- Name: lo_submission_answer; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.lo_submission_answer ENABLE ROW LEVEL SECURITY;

--
-- Name: lo_video_rating; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.lo_video_rating ENABLE ROW LEVEL SECURITY;

--
-- Name: locations; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.locations ENABLE ROW LEVEL SECURITY;

--
-- Name: master_study_plan; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.master_study_plan ENABLE ROW LEVEL SECURITY;

--
-- Name: max_score_submission; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.max_score_submission ENABLE ROW LEVEL SECURITY;

--
-- Name: permission; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.permission ENABLE ROW LEVEL SECURITY;

--
-- Name: permission_role; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.permission_role ENABLE ROW LEVEL SECURITY;

--
-- Name: question_group; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.question_group ENABLE ROW LEVEL SECURITY;

--
-- Name: question_tag; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.question_tag ENABLE ROW LEVEL SECURITY;

--
-- Name: question_tag_type; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.question_tag_type ENABLE ROW LEVEL SECURITY;

--
-- Name: quiz_sets; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.quiz_sets ENABLE ROW LEVEL SECURITY;

--
-- Name: quizzes; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.quizzes ENABLE ROW LEVEL SECURITY;

--
-- Name: academic_year rls_academic_year; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_academic_year ON public.academic_year USING (public.permission_check(resource_path, 'academic_year'::text)) WITH CHECK (public.permission_check(resource_path, 'academic_year'::text));


--
-- Name: academic_year rls_academic_year_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_academic_year_restrictive ON public.academic_year AS RESTRICTIVE USING (public.permission_check(resource_path, 'academic_year'::text)) WITH CHECK (public.permission_check(resource_path, 'academic_year'::text));


--
-- Name: allocate_marker rls_allocate_marker; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_allocate_marker ON public.allocate_marker USING (public.permission_check(resource_path, 'allocate_marker'::text)) WITH CHECK (public.permission_check(resource_path, 'allocate_marker'::text));


--
-- Name: allocate_marker rls_allocate_marker_location; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_allocate_marker_location ON public.allocate_marker AS RESTRICTIVE USING ((1 <= ( SELECT count(*) AS count
   FROM (public.granted_permissions p
     JOIN public.user_access_paths usp ON ((usp.location_id = p.location_id)))
  WHERE ((p.user_id = current_setting('app.user_id'::text)) AND (p.permission_id IN ( SELECT p2.permission_id
           FROM public.permission p2
          WHERE ((p2.permission_name = 'syllabus.allocate_marker.read'::text) AND (p2.resource_path = current_setting('permission.resource_path'::text))))) AND (usp.deleted_at IS NULL))
 LIMIT 1))) WITH CHECK ((1 <= ( SELECT count(*) AS count
   FROM (public.granted_permissions p
     JOIN public.user_access_paths usp ON ((usp.location_id = p.location_id)))
  WHERE ((p.user_id = current_setting('app.user_id'::text)) AND (p.permission_id IN ( SELECT p2.permission_id
           FROM public.permission p2
          WHERE ((p2.permission_name = 'syllabus.allocate_marker.write'::text) AND (p2.resource_path = current_setting('permission.resource_path'::text))))) AND (usp.user_id = allocate_marker.created_by) AND (usp.deleted_at IS NULL))
 LIMIT 1)));


--
-- Name: assessment rls_assessment; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_assessment ON public.assessment USING (public.permission_check(resource_path, 'assessment'::text)) WITH CHECK (public.permission_check(resource_path, 'assessment'::text));


--
-- Name: assessment rls_assessment_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_assessment_restrictive ON public.assessment AS RESTRICTIVE USING (public.permission_check(resource_path, 'assessment'::text)) WITH CHECK (public.permission_check(resource_path, 'assessment'::text));


--
-- Name: assessment_session rls_assessment_session; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_assessment_session ON public.assessment_session USING (public.permission_check(resource_path, 'assessment_session'::text)) WITH CHECK (public.permission_check(resource_path, 'assessment_session'::text));


--
-- Name: assessment_session rls_assessment_session_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_assessment_session_restrictive ON public.assessment_session AS RESTRICTIVE USING (public.permission_check(resource_path, 'assessment_session'::text)) WITH CHECK (public.permission_check(resource_path, 'assessment_session'::text));


--
-- Name: assessment_submission rls_assessment_submission; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_assessment_submission ON public.assessment_submission USING (public.permission_check(resource_path, 'assessment_submission'::text)) WITH CHECK (public.permission_check(resource_path, 'assessment_submission'::text));


--
-- Name: assessment_submission rls_assessment_submission_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_assessment_submission_restrictive ON public.assessment_submission AS RESTRICTIVE USING (public.permission_check(resource_path, 'assessment_submission'::text)) WITH CHECK (public.permission_check(resource_path, 'assessment_submission'::text));


--
-- Name: assign_study_plan_tasks rls_assign_study_plan_tasks; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_assign_study_plan_tasks ON public.assign_study_plan_tasks USING (public.permission_check(resource_path, 'assign_study_plan_tasks'::text)) WITH CHECK (public.permission_check(resource_path, 'assign_study_plan_tasks'::text));


--
-- Name: assign_study_plan_tasks rls_assign_study_plan_tasks_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_assign_study_plan_tasks_restrictive ON public.assign_study_plan_tasks AS RESTRICTIVE USING (public.permission_check(resource_path, 'assign_study_plan_tasks'::text)) WITH CHECK (public.permission_check(resource_path, 'assign_study_plan_tasks'::text));


--
-- Name: assignment rls_assignment; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_assignment ON public.assignment USING (public.permission_check(resource_path, 'assignment'::text)) WITH CHECK (public.permission_check(resource_path, 'assignment'::text));


--
-- Name: assignment rls_assignment_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_assignment_restrictive ON public.assignment AS RESTRICTIVE USING (public.permission_check(resource_path, 'assignment'::text)) WITH CHECK (public.permission_check(resource_path, 'assignment'::text));


--
-- Name: assignment_study_plan_items rls_assignment_study_plan_items; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_assignment_study_plan_items ON public.assignment_study_plan_items USING (public.permission_check(resource_path, 'assignment_study_plan_items'::text)) WITH CHECK (public.permission_check(resource_path, 'assignment_study_plan_items'::text));


--
-- Name: assignment_study_plan_items rls_assignment_study_plan_items_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_assignment_study_plan_items_restrictive ON public.assignment_study_plan_items AS RESTRICTIVE USING (public.permission_check(resource_path, 'assignment_study_plan_items'::text)) WITH CHECK (public.permission_check(resource_path, 'assignment_study_plan_items'::text));


--
-- Name: assignments rls_assignments; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_assignments ON public.assignments USING (public.permission_check(resource_path, 'assignments'::text)) WITH CHECK (public.permission_check(resource_path, 'assignments'::text));


--
-- Name: assignments rls_assignments_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_assignments_restrictive ON public.assignments AS RESTRICTIVE USING (public.permission_check(resource_path, 'assignments'::text)) WITH CHECK (public.permission_check(resource_path, 'assignments'::text));


--
-- Name: books rls_books; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_books ON public.books USING (public.permission_check(resource_path, 'books'::text)) WITH CHECK (public.permission_check(resource_path, 'books'::text));


--
-- Name: books_chapters rls_books_chapters; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_books_chapters ON public.books_chapters USING (public.permission_check(resource_path, 'books_chapters'::text)) WITH CHECK (public.permission_check(resource_path, 'books_chapters'::text));


--
-- Name: books_chapters rls_books_chapters_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_books_chapters_restrictive ON public.books_chapters AS RESTRICTIVE USING (public.permission_check(resource_path, 'books_chapters'::text)) WITH CHECK (public.permission_check(resource_path, 'books_chapters'::text));


--
-- Name: books rls_books_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_books_restrictive ON public.books AS RESTRICTIVE USING (public.permission_check(resource_path, 'books'::text)) WITH CHECK (public.permission_check(resource_path, 'books'::text));


--
-- Name: cerebry_classes rls_cerebry_classes; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_cerebry_classes ON public.cerebry_classes USING (public.permission_check(resource_path, 'cerebry_classes'::text)) WITH CHECK (public.permission_check(resource_path, 'cerebry_classes'::text));


--
-- Name: cerebry_classes rls_cerebry_classes_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_cerebry_classes_restrictive ON public.cerebry_classes AS RESTRICTIVE USING (public.permission_check(resource_path, 'cerebry_classes'::text)) WITH CHECK (public.permission_check(resource_path, 'cerebry_classes'::text));


--
-- Name: chapters rls_chapters; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_chapters ON public.chapters USING (public.permission_check(resource_path, 'chapters'::text)) WITH CHECK (public.permission_check(resource_path, 'chapters'::text));


--
-- Name: chapters rls_chapters_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_chapters_restrictive ON public.chapters AS RESTRICTIVE USING (public.permission_check(resource_path, 'chapters'::text)) WITH CHECK (public.permission_check(resource_path, 'chapters'::text));


--
-- Name: class_students rls_class_students; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_class_students ON public.class_students USING (public.permission_check(resource_path, 'class_students'::text)) WITH CHECK (public.permission_check(resource_path, 'class_students'::text));


--
-- Name: class_students rls_class_students_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_class_students_restrictive ON public.class_students AS RESTRICTIVE USING (public.permission_check(resource_path, 'class_students'::text)) WITH CHECK (public.permission_check(resource_path, 'class_students'::text));


--
-- Name: class_study_plans rls_class_study_plans; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_class_study_plans ON public.class_study_plans USING (public.permission_check(resource_path, 'class_study_plans'::text)) WITH CHECK (public.permission_check(resource_path, 'class_study_plans'::text));


--
-- Name: class_study_plans rls_class_study_plans_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_class_study_plans_restrictive ON public.class_study_plans AS RESTRICTIVE USING (public.permission_check(resource_path, 'class_study_plans'::text)) WITH CHECK (public.permission_check(resource_path, 'class_study_plans'::text));


--
-- Name: content_bank_medias rls_content_bank_medias; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_content_bank_medias ON public.content_bank_medias USING (public.permission_check(resource_path, 'content_bank_medias'::text)) WITH CHECK (public.permission_check(resource_path, 'content_bank_medias'::text));


--
-- Name: content_bank_medias rls_content_bank_medias_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_content_bank_medias_restrictive ON public.content_bank_medias AS RESTRICTIVE USING (public.permission_check(resource_path, 'content_bank_medias'::text)) WITH CHECK (public.permission_check(resource_path, 'content_bank_medias'::text));


--
-- Name: course_access_paths rls_course_access_paths; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_course_access_paths ON public.course_access_paths USING (public.permission_check(resource_path, 'course_access_paths'::text)) WITH CHECK (public.permission_check(resource_path, 'course_access_paths'::text));


--
-- Name: course_classes rls_course_classes; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_course_classes ON public.course_classes USING (public.permission_check(resource_path, 'course_classes'::text)) WITH CHECK (public.permission_check(resource_path, 'course_classes'::text));


--
-- Name: course_classes rls_course_classes_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_course_classes_restrictive ON public.course_classes AS RESTRICTIVE USING (public.permission_check(resource_path, 'course_classes'::text)) WITH CHECK (public.permission_check(resource_path, 'course_classes'::text));


--
-- Name: course_student_subscriptions rls_course_student_subscriptions; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_course_student_subscriptions ON public.course_student_subscriptions USING (public.permission_check(resource_path, 'course_student_subscriptions'::text)) WITH CHECK (public.permission_check(resource_path, 'course_student_subscriptions'::text));


--
-- Name: course_student_subscriptions rls_course_student_subscriptions_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_course_student_subscriptions_restrictive ON public.course_student_subscriptions AS RESTRICTIVE USING (public.permission_check(resource_path, 'course_student_subscriptions'::text)) WITH CHECK (public.permission_check(resource_path, 'course_student_subscriptions'::text));


--
-- Name: course_students rls_course_students; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_course_students ON public.course_students USING (public.permission_check(resource_path, 'course_students'::text)) WITH CHECK (public.permission_check(resource_path, 'course_students'::text));


--
-- Name: course_students_access_paths rls_course_students_access_paths; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_course_students_access_paths ON public.course_students_access_paths USING (public.permission_check(resource_path, 'course_students_access_paths'::text)) WITH CHECK (public.permission_check(resource_path, 'course_students_access_paths'::text));


--
-- Name: course_students_access_paths rls_course_students_access_paths_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_course_students_access_paths_restrictive ON public.course_students_access_paths AS RESTRICTIVE USING (public.permission_check(resource_path, 'course_students_access_paths'::text)) WITH CHECK (public.permission_check(resource_path, 'course_students_access_paths'::text));


--
-- Name: course_students rls_course_students_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_course_students_restrictive ON public.course_students AS RESTRICTIVE USING (public.permission_check(resource_path, 'course_students'::text)) WITH CHECK (public.permission_check(resource_path, 'course_students'::text));


--
-- Name: course_study_plans rls_course_study_plans; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_course_study_plans ON public.course_study_plans USING (public.permission_check(resource_path, 'course_study_plans'::text)) WITH CHECK (public.permission_check(resource_path, 'course_study_plans'::text));


--
-- Name: course_study_plans rls_course_study_plans_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_course_study_plans_restrictive ON public.course_study_plans AS RESTRICTIVE USING (public.permission_check(resource_path, 'course_study_plans'::text)) WITH CHECK (public.permission_check(resource_path, 'course_study_plans'::text));


--
-- Name: courses rls_courses; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_courses ON public.courses USING (public.permission_check(resource_path, 'courses'::text)) WITH CHECK (public.permission_check(resource_path, 'courses'::text));


--
-- Name: courses_books rls_courses_books; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_courses_books ON public.courses_books USING (public.permission_check(resource_path, 'courses_books'::text)) WITH CHECK (public.permission_check(resource_path, 'courses_books'::text));


--
-- Name: courses_books rls_courses_books_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_courses_books_restrictive ON public.courses_books AS RESTRICTIVE USING (public.permission_check(resource_path, 'courses_books'::text)) WITH CHECK (public.permission_check(resource_path, 'courses_books'::text));


--
-- Name: courses rls_courses_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_courses_restrictive ON public.courses AS RESTRICTIVE USING (public.permission_check(resource_path, 'courses'::text)) WITH CHECK (public.permission_check(resource_path, 'courses'::text));


--
-- Name: exam_lo rls_exam_lo; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_exam_lo ON public.exam_lo USING (public.permission_check(resource_path, 'exam_lo'::text)) WITH CHECK (public.permission_check(resource_path, 'exam_lo'::text));


--
-- Name: exam_lo rls_exam_lo_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_exam_lo_restrictive ON public.exam_lo AS RESTRICTIVE USING (public.permission_check(resource_path, 'exam_lo'::text)) WITH CHECK (public.permission_check(resource_path, 'exam_lo'::text));


--
-- Name: exam_lo_submission rls_exam_lo_submission; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_exam_lo_submission ON public.exam_lo_submission USING (public.permission_check(resource_path, 'exam_lo_submission'::text)) WITH CHECK (public.permission_check(resource_path, 'exam_lo_submission'::text));


--
-- Name: exam_lo_submission_answer rls_exam_lo_submission_answer; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_exam_lo_submission_answer ON public.exam_lo_submission_answer USING (public.permission_check(resource_path, 'exam_lo_submission_answer'::text)) WITH CHECK (public.permission_check(resource_path, 'exam_lo_submission_answer'::text));


--
-- Name: exam_lo_submission_answer rls_exam_lo_submission_answer_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_exam_lo_submission_answer_restrictive ON public.exam_lo_submission_answer AS RESTRICTIVE USING (public.permission_check(resource_path, 'exam_lo_submission_answer'::text)) WITH CHECK (public.permission_check(resource_path, 'exam_lo_submission_answer'::text));


--
-- Name: exam_lo_submission rls_exam_lo_submission_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_exam_lo_submission_restrictive ON public.exam_lo_submission AS RESTRICTIVE USING (public.permission_check(resource_path, 'exam_lo_submission'::text)) WITH CHECK (public.permission_check(resource_path, 'exam_lo_submission'::text));


--
-- Name: exam_lo_submission_score rls_exam_lo_submission_score; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_exam_lo_submission_score ON public.exam_lo_submission_score USING (public.permission_check(resource_path, 'exam_lo_submission_score'::text)) WITH CHECK (public.permission_check(resource_path, 'exam_lo_submission_score'::text));


--
-- Name: exam_lo_submission_score rls_exam_lo_submission_score_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_exam_lo_submission_score_restrictive ON public.exam_lo_submission_score AS RESTRICTIVE USING (public.permission_check(resource_path, 'exam_lo_submission_score'::text)) WITH CHECK (public.permission_check(resource_path, 'exam_lo_submission_score'::text));


--
-- Name: feedback_session rls_feedback_session; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_feedback_session ON public.feedback_session USING (public.permission_check(resource_path, 'feedback_session'::text)) WITH CHECK (public.permission_check(resource_path, 'feedback_session'::text));


--
-- Name: feedback_session rls_feedback_session_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_feedback_session_restrictive ON public.feedback_session AS RESTRICTIVE USING (public.permission_check(resource_path, 'feedback_session'::text)) WITH CHECK (public.permission_check(resource_path, 'feedback_session'::text));


--
-- Name: flash_card rls_flash_card; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_flash_card ON public.flash_card USING (public.permission_check(resource_path, 'flash_card'::text)) WITH CHECK (public.permission_check(resource_path, 'flash_card'::text));


--
-- Name: flash_card rls_flash_card_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_flash_card_restrictive ON public.flash_card AS RESTRICTIVE USING (public.permission_check(resource_path, 'flash_card'::text)) WITH CHECK (public.permission_check(resource_path, 'flash_card'::text));


--
-- Name: flash_card_submission rls_flash_card_submission; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_flash_card_submission ON public.flash_card_submission USING (public.permission_check(resource_path, 'flash_card_submission'::text)) WITH CHECK (public.permission_check(resource_path, 'flash_card_submission'::text));


--
-- Name: flash_card_submission_answer rls_flash_card_submission_answer; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_flash_card_submission_answer ON public.flash_card_submission_answer USING (public.permission_check(resource_path, 'flash_card_submission_answer'::text)) WITH CHECK (public.permission_check(resource_path, 'flash_card_submission_answer'::text));


--
-- Name: flash_card_submission_answer rls_flash_card_submission_answer_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_flash_card_submission_answer_restrictive ON public.flash_card_submission_answer AS RESTRICTIVE USING (public.permission_check(resource_path, 'flash_card_submission_answer'::text)) WITH CHECK (public.permission_check(resource_path, 'flash_card_submission_answer'::text));


--
-- Name: flash_card_submission rls_flash_card_submission_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_flash_card_submission_restrictive ON public.flash_card_submission AS RESTRICTIVE USING (public.permission_check(resource_path, 'flash_card_submission'::text)) WITH CHECK (public.permission_check(resource_path, 'flash_card_submission'::text));


--
-- Name: flashcard_progressions rls_flashcard_progressions; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_flashcard_progressions ON public.flashcard_progressions USING (public.permission_check(resource_path, 'flashcard_progressions'::text)) WITH CHECK (public.permission_check(resource_path, 'flashcard_progressions'::text));


--
-- Name: flashcard_progressions rls_flashcard_progressions_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_flashcard_progressions_restrictive ON public.flashcard_progressions AS RESTRICTIVE USING (public.permission_check(resource_path, 'flashcard_progressions'::text)) WITH CHECK (public.permission_check(resource_path, 'flashcard_progressions'::text));


--
-- Name: flashcard_speeches rls_flashcard_speeches; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_flashcard_speeches ON public.flashcard_speeches USING (public.permission_check(resource_path, 'flashcard_speeches'::text)) WITH CHECK (public.permission_check(resource_path, 'flashcard_speeches'::text));


--
-- Name: flashcard_speeches rls_flashcard_speeches_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_flashcard_speeches_restrictive ON public.flashcard_speeches AS RESTRICTIVE USING (public.permission_check(resource_path, 'flashcard_speeches'::text)) WITH CHECK (public.permission_check(resource_path, 'flashcard_speeches'::text));


--
-- Name: grade rls_grade; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_grade ON public.grade USING (public.permission_check(resource_path, 'grade'::text)) WITH CHECK (public.permission_check(resource_path, 'grade'::text));


--
-- Name: grade_book_setting rls_grade_book_setting; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_grade_book_setting ON public.grade_book_setting USING (public.permission_check(resource_path, 'grade_book_setting'::text)) WITH CHECK (public.permission_check(resource_path, 'grade_book_setting'::text));


--
-- Name: granted_permission rls_granted_permission; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_granted_permission ON public.granted_permission USING (public.permission_check(resource_path, 'granted_permission'::text)) WITH CHECK (public.permission_check(resource_path, 'granted_permission'::text));


--
-- Name: granted_permission rls_granted_permission_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_granted_permission_restrictive ON public.granted_permission AS RESTRICTIVE USING (public.permission_check(resource_path, 'granted_permission'::text)) WITH CHECK (public.permission_check(resource_path, 'granted_permission'::text));


--
-- Name: granted_role rls_granted_role; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_granted_role ON public.granted_role USING (public.permission_check(resource_path, 'granted_role'::text)) WITH CHECK (public.permission_check(resource_path, 'granted_role'::text));


--
-- Name: granted_role_access_path rls_granted_role_access_path; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_granted_role_access_path ON public.granted_role_access_path USING (public.permission_check(resource_path, 'granted_role_access_path'::text)) WITH CHECK (public.permission_check(resource_path, 'granted_role_access_path'::text));


--
-- Name: granted_role_access_path rls_granted_role_access_path_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_granted_role_access_path_restrictive ON public.granted_role_access_path AS RESTRICTIVE USING (public.permission_check(resource_path, 'granted_role_access_path'::text)) WITH CHECK (public.permission_check(resource_path, 'granted_role_access_path'::text));


--
-- Name: granted_role rls_granted_role_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_granted_role_restrictive ON public.granted_role AS RESTRICTIVE USING (public.permission_check(resource_path, 'granted_role'::text)) WITH CHECK (public.permission_check(resource_path, 'granted_role'::text));


--
-- Name: groups rls_groups; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_groups ON public.groups USING (public.permission_check(resource_path, 'groups'::text)) WITH CHECK (public.permission_check(resource_path, 'groups'::text));


--
-- Name: groups rls_groups_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_groups_restrictive ON public.groups AS RESTRICTIVE USING (public.permission_check(resource_path, 'groups'::text)) WITH CHECK (public.permission_check(resource_path, 'groups'::text));


--
-- Name: import_study_plan_task rls_import_study_plan_task; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_import_study_plan_task ON public.import_study_plan_task USING (public.permission_check(resource_path, 'import_study_plan_task'::text)) WITH CHECK (public.permission_check(resource_path, 'import_study_plan_task'::text));


--
-- Name: individual_study_plan rls_individual_study_plan; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_individual_study_plan ON public.individual_study_plan USING (public.permission_check(resource_path, 'individual_study_plan'::text)) WITH CHECK (public.permission_check(resource_path, 'individual_study_plan'::text));


--
-- Name: individual_study_plan rls_individual_study_plan_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_individual_study_plan_restrictive ON public.individual_study_plan AS RESTRICTIVE USING (public.permission_check(resource_path, 'individual_study_plan'::text)) WITH CHECK (public.permission_check(resource_path, 'individual_study_plan'::text));


--
-- Name: learning_material rls_learning_material; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_learning_material ON public.learning_material USING (public.permission_check(resource_path, 'learning_material'::text)) WITH CHECK (public.permission_check(resource_path, 'learning_material'::text));


--
-- Name: learning_material rls_learning_material_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_learning_material_restrictive ON public.learning_material AS RESTRICTIVE USING (public.permission_check(resource_path, 'learning_material'::text)) WITH CHECK (public.permission_check(resource_path, 'learning_material'::text));


--
-- Name: learning_objective rls_learning_objective; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_learning_objective ON public.learning_objective USING (public.permission_check(resource_path, 'learning_objective'::text)) WITH CHECK (public.permission_check(resource_path, 'learning_objective'::text));


--
-- Name: learning_objective rls_learning_objective_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_learning_objective_restrictive ON public.learning_objective AS RESTRICTIVE USING (public.permission_check(resource_path, 'learning_objective'::text)) WITH CHECK (public.permission_check(resource_path, 'learning_objective'::text));


--
-- Name: learning_objectives rls_learning_objectives; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_learning_objectives ON public.learning_objectives USING (public.permission_check(resource_path, 'learning_objectives'::text)) WITH CHECK (public.permission_check(resource_path, 'learning_objectives'::text));


--
-- Name: learning_objectives rls_learning_objectives_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_learning_objectives_restrictive ON public.learning_objectives AS RESTRICTIVE USING (public.permission_check(resource_path, 'learning_objectives'::text)) WITH CHECK (public.permission_check(resource_path, 'learning_objectives'::text));


--
-- Name: lms_learning_material_list rls_lms_learning_material_list; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lms_learning_material_list ON public.lms_learning_material_list USING (public.permission_check(resource_path, 'lms_learning_material_list'::text)) WITH CHECK (public.permission_check(resource_path, 'lms_learning_material_list'::text));


--
-- Name: lms_learning_material_list rls_lms_learning_material_list_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lms_learning_material_list_restrictive ON public.lms_learning_material_list AS RESTRICTIVE USING (public.permission_check(resource_path, 'lms_learning_material_list'::text)) WITH CHECK (public.permission_check(resource_path, 'lms_learning_material_list'::text));


--
-- Name: lms_student_study_plan_item rls_lms_student_study_plan_item; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lms_student_study_plan_item ON public.lms_student_study_plan_item USING (public.permission_check(resource_path, 'lms_student_study_plan_item'::text)) WITH CHECK (public.permission_check(resource_path, 'lms_student_study_plan_item'::text));


--
-- Name: lms_student_study_plan_item rls_lms_student_study_plan_item_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lms_student_study_plan_item_restrictive ON public.lms_student_study_plan_item AS RESTRICTIVE USING (public.permission_check(resource_path, 'lms_student_study_plan_item'::text)) WITH CHECK (public.permission_check(resource_path, 'lms_student_study_plan_item'::text));


--
-- Name: lms_student_study_plans rls_lms_student_study_plans; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lms_student_study_plans ON public.lms_student_study_plans USING (public.permission_check(resource_path, 'lms_student_study_plans'::text)) WITH CHECK (public.permission_check(resource_path, 'lms_student_study_plans'::text));


--
-- Name: lms_student_study_plans rls_lms_student_study_plans_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lms_student_study_plans_restrictive ON public.lms_student_study_plans AS RESTRICTIVE USING (public.permission_check(resource_path, 'lms_student_study_plans'::text)) WITH CHECK (public.permission_check(resource_path, 'lms_student_study_plans'::text));


--
-- Name: lms_study_plan_items rls_lms_study_plan_items; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lms_study_plan_items ON public.lms_study_plan_items USING (public.permission_check(resource_path, 'lms_study_plan_items'::text)) WITH CHECK (public.permission_check(resource_path, 'lms_study_plan_items'::text));


--
-- Name: lms_study_plan_items rls_lms_study_plan_items_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lms_study_plan_items_restrictive ON public.lms_study_plan_items AS RESTRICTIVE USING (public.permission_check(resource_path, 'lms_study_plan_items'::text)) WITH CHECK (public.permission_check(resource_path, 'lms_study_plan_items'::text));


--
-- Name: lms_study_plans rls_lms_study_plans; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lms_study_plans ON public.lms_study_plans USING (public.permission_check(resource_path, 'lms_study_plans'::text)) WITH CHECK (public.permission_check(resource_path, 'lms_study_plans'::text));


--
-- Name: lms_study_plans rls_lms_study_plans_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lms_study_plans_restrictive ON public.lms_study_plans AS RESTRICTIVE USING (public.permission_check(resource_path, 'lms_study_plans'::text)) WITH CHECK (public.permission_check(resource_path, 'lms_study_plans'::text));


--
-- Name: lo_progression rls_lo_progression; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lo_progression ON public.lo_progression USING (public.permission_check(resource_path, 'lo_progression'::text)) WITH CHECK (public.permission_check(resource_path, 'lo_progression'::text));


--
-- Name: lo_progression_answer rls_lo_progression_answer; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lo_progression_answer ON public.lo_progression_answer USING (public.permission_check(resource_path, 'lo_progression_answer'::text)) WITH CHECK (public.permission_check(resource_path, 'lo_progression_answer'::text));


--
-- Name: lo_progression_answer rls_lo_progression_answer_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lo_progression_answer_restrictive ON public.lo_progression_answer AS RESTRICTIVE USING (public.permission_check(resource_path, 'lo_progression_answer'::text)) WITH CHECK (public.permission_check(resource_path, 'lo_progression_answer'::text));


--
-- Name: lo_progression rls_lo_progression_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lo_progression_restrictive ON public.lo_progression AS RESTRICTIVE USING (public.permission_check(resource_path, 'lo_progression'::text)) WITH CHECK (public.permission_check(resource_path, 'lo_progression'::text));


--
-- Name: lo_study_plan_items rls_lo_study_plan_items; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lo_study_plan_items ON public.lo_study_plan_items USING (public.permission_check(resource_path, 'lo_study_plan_items'::text)) WITH CHECK (public.permission_check(resource_path, 'lo_study_plan_items'::text));


--
-- Name: lo_study_plan_items rls_lo_study_plan_items_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lo_study_plan_items_restrictive ON public.lo_study_plan_items AS RESTRICTIVE USING (public.permission_check(resource_path, 'lo_study_plan_items'::text)) WITH CHECK (public.permission_check(resource_path, 'lo_study_plan_items'::text));


--
-- Name: lo_submission rls_lo_submission; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lo_submission ON public.lo_submission USING (public.permission_check(resource_path, 'lo_submission'::text)) WITH CHECK (public.permission_check(resource_path, 'lo_submission'::text));


--
-- Name: lo_submission_answer rls_lo_submission_answer; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lo_submission_answer ON public.lo_submission_answer USING (public.permission_check(resource_path, 'lo_submission_answer'::text)) WITH CHECK (public.permission_check(resource_path, 'lo_submission_answer'::text));


--
-- Name: lo_submission_answer rls_lo_submission_answer_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lo_submission_answer_restrictive ON public.lo_submission_answer AS RESTRICTIVE USING (public.permission_check(resource_path, 'lo_submission_answer'::text)) WITH CHECK (public.permission_check(resource_path, 'lo_submission_answer'::text));


--
-- Name: lo_submission rls_lo_submission_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lo_submission_restrictive ON public.lo_submission AS RESTRICTIVE USING (public.permission_check(resource_path, 'lo_submission'::text)) WITH CHECK (public.permission_check(resource_path, 'lo_submission'::text));


--
-- Name: lo_video_rating rls_lo_video_rating; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lo_video_rating ON public.lo_video_rating USING (public.permission_check(resource_path, 'lo_video_rating'::text)) WITH CHECK (public.permission_check(resource_path, 'lo_video_rating'::text));


--
-- Name: lo_video_rating rls_lo_video_rating_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_lo_video_rating_restrictive ON public.lo_video_rating AS RESTRICTIVE USING (public.permission_check(resource_path, 'lo_video_rating'::text)) WITH CHECK (public.permission_check(resource_path, 'lo_video_rating'::text));


--
-- Name: locations rls_locations; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_locations ON public.locations USING (public.permission_check(resource_path, 'locations'::text)) WITH CHECK (public.permission_check(resource_path, 'locations'::text));


--
-- Name: locations rls_locations_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_locations_restrictive ON public.locations AS RESTRICTIVE USING (public.permission_check(resource_path, 'locations'::text)) WITH CHECK (public.permission_check(resource_path, 'locations'::text));


--
-- Name: master_study_plan rls_master_study_plan; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_master_study_plan ON public.master_study_plan USING (public.permission_check(resource_path, 'master_study_plan'::text)) WITH CHECK (public.permission_check(resource_path, 'master_study_plan'::text));


--
-- Name: master_study_plan rls_master_study_plan_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_master_study_plan_restrictive ON public.master_study_plan AS RESTRICTIVE USING (public.permission_check(resource_path, 'master_study_plan'::text)) WITH CHECK (public.permission_check(resource_path, 'master_study_plan'::text));


--
-- Name: max_score_submission rls_max_score_submission; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_max_score_submission ON public.max_score_submission USING (public.permission_check(resource_path, 'max_score_submission'::text)) WITH CHECK (public.permission_check(resource_path, 'max_score_submission'::text));


--
-- Name: permission rls_permission; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_permission ON public.permission USING (public.permission_check(resource_path, 'permission'::text)) WITH CHECK (public.permission_check(resource_path, 'permission'::text));


--
-- Name: permission rls_permission_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_permission_restrictive ON public.permission AS RESTRICTIVE USING (public.permission_check(resource_path, 'permission'::text)) WITH CHECK (public.permission_check(resource_path, 'permission'::text));


--
-- Name: permission_role rls_permission_role; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_permission_role ON public.permission_role USING (public.permission_check(resource_path, 'permission_role'::text)) WITH CHECK (public.permission_check(resource_path, 'permission_role'::text));


--
-- Name: permission_role rls_permission_role_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_permission_role_restrictive ON public.permission_role AS RESTRICTIVE USING (public.permission_check(resource_path, 'permission_role'::text)) WITH CHECK (public.permission_check(resource_path, 'permission_role'::text));


--
-- Name: question_group rls_question_group; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_question_group ON public.question_group USING (public.permission_check(resource_path, 'question_group'::text)) WITH CHECK (public.permission_check(resource_path, 'question_group'::text));


--
-- Name: question_group rls_question_group_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_question_group_restrictive ON public.question_group AS RESTRICTIVE USING (public.permission_check(resource_path, 'question_group'::text)) WITH CHECK (public.permission_check(resource_path, 'question_group'::text));


--
-- Name: question_tag rls_question_tag; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_question_tag ON public.question_tag USING (public.permission_check(resource_path, 'question_tag'::text)) WITH CHECK (public.permission_check(resource_path, 'question_tag'::text));


--
-- Name: question_tag rls_question_tag_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_question_tag_restrictive ON public.question_tag AS RESTRICTIVE USING (public.permission_check(resource_path, 'question_tag'::text)) WITH CHECK (public.permission_check(resource_path, 'question_tag'::text));


--
-- Name: question_tag_type rls_question_tag_type; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_question_tag_type ON public.question_tag_type USING (public.permission_check(resource_path, 'question_tag_type'::text)) WITH CHECK (public.permission_check(resource_path, 'question_tag_type'::text));


--
-- Name: question_tag_type rls_question_tag_type_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_question_tag_type_restrictive ON public.question_tag_type AS RESTRICTIVE USING (public.permission_check(resource_path, 'question_tag_type'::text)) WITH CHECK (public.permission_check(resource_path, 'question_tag_type'::text));


--
-- Name: quiz_sets rls_quiz_sets; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_quiz_sets ON public.quiz_sets USING (public.permission_check(resource_path, 'quiz_sets'::text)) WITH CHECK (public.permission_check(resource_path, 'quiz_sets'::text));


--
-- Name: quiz_sets rls_quiz_sets_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_quiz_sets_restrictive ON public.quiz_sets AS RESTRICTIVE USING (public.permission_check(resource_path, 'quiz_sets'::text)) WITH CHECK (public.permission_check(resource_path, 'quiz_sets'::text));


--
-- Name: quizzes rls_quizzes; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_quizzes ON public.quizzes USING (public.permission_check(resource_path, 'quizzes'::text)) WITH CHECK (public.permission_check(resource_path, 'quizzes'::text));


--
-- Name: quizzes rls_quizzes_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_quizzes_restrictive ON public.quizzes AS RESTRICTIVE USING (public.permission_check(resource_path, 'quizzes'::text)) WITH CHECK (public.permission_check(resource_path, 'quizzes'::text));


--
-- Name: role rls_role; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_role ON public.role USING (public.permission_check(resource_path, 'role'::text)) WITH CHECK (public.permission_check(resource_path, 'role'::text));


--
-- Name: role rls_role_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_role_restrictive ON public.role AS RESTRICTIVE USING (public.permission_check(resource_path, 'role'::text)) WITH CHECK (public.permission_check(resource_path, 'role'::text));


--
-- Name: shuffled_quiz_sets rls_shuffled_quiz_sets; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_shuffled_quiz_sets ON public.shuffled_quiz_sets USING (public.permission_check(resource_path, 'shuffled_quiz_sets'::text)) WITH CHECK (public.permission_check(resource_path, 'shuffled_quiz_sets'::text));


--
-- Name: shuffled_quiz_sets rls_shuffled_quiz_sets_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_shuffled_quiz_sets_restrictive ON public.shuffled_quiz_sets AS RESTRICTIVE USING (public.permission_check(resource_path, 'shuffled_quiz_sets'::text)) WITH CHECK (public.permission_check(resource_path, 'shuffled_quiz_sets'::text));


--
-- Name: student_latest_submissions rls_student_latest_submissions; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_student_latest_submissions ON public.student_latest_submissions USING (public.permission_check(resource_path, 'student_latest_submissions'::text)) WITH CHECK (public.permission_check(resource_path, 'student_latest_submissions'::text));


--
-- Name: student_latest_submissions rls_student_latest_submissions_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_student_latest_submissions_restrictive ON public.student_latest_submissions AS RESTRICTIVE USING (public.permission_check(resource_path, 'student_latest_submissions'::text)) WITH CHECK (public.permission_check(resource_path, 'student_latest_submissions'::text));


--
-- Name: student_learning_time_by_daily rls_student_learning_time_by_daily; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_student_learning_time_by_daily ON public.student_learning_time_by_daily USING (public.permission_check(resource_path, 'student_learning_time_by_daily'::text)) WITH CHECK (public.permission_check(resource_path, 'student_learning_time_by_daily'::text));


--
-- Name: student_learning_time_by_daily rls_student_learning_time_by_daily_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_student_learning_time_by_daily_restrictive ON public.student_learning_time_by_daily AS RESTRICTIVE USING (public.permission_check(resource_path, 'student_learning_time_by_daily'::text)) WITH CHECK (public.permission_check(resource_path, 'student_learning_time_by_daily'::text));


--
-- Name: student_study_plans rls_student_study_plans; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_student_study_plans ON public.student_study_plans USING (public.permission_check(resource_path, 'student_study_plans'::text)) WITH CHECK (public.permission_check(resource_path, 'student_study_plans'::text));


--
-- Name: student_study_plans rls_student_study_plans_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_student_study_plans_restrictive ON public.student_study_plans AS RESTRICTIVE USING (public.permission_check(resource_path, 'student_study_plans'::text)) WITH CHECK (public.permission_check(resource_path, 'student_study_plans'::text));


--
-- Name: student_submission_grades rls_student_submission_grades; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_student_submission_grades ON public.student_submission_grades USING (public.permission_check(resource_path, 'student_submission_grades'::text)) WITH CHECK (public.permission_check(resource_path, 'student_submission_grades'::text));


--
-- Name: student_submission_grades rls_student_submission_grades_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_student_submission_grades_restrictive ON public.student_submission_grades AS RESTRICTIVE USING (public.permission_check(resource_path, 'student_submission_grades'::text)) WITH CHECK (public.permission_check(resource_path, 'student_submission_grades'::text));


--
-- Name: student_submissions rls_student_submissions; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_student_submissions ON public.student_submissions USING (public.permission_check(resource_path, 'student_submissions'::text)) WITH CHECK (public.permission_check(resource_path, 'student_submissions'::text));


--
-- Name: student_submissions rls_student_submissions_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_student_submissions_restrictive ON public.student_submissions AS RESTRICTIVE USING (public.permission_check(resource_path, 'student_submissions'::text)) WITH CHECK (public.permission_check(resource_path, 'student_submissions'::text));


--
-- Name: students rls_students; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_students ON public.students USING (public.permission_check(resource_path, 'students'::text)) WITH CHECK (public.permission_check(resource_path, 'students'::text));


--
-- Name: students_learning_objectives_completeness rls_students_learning_objectives_completeness; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_students_learning_objectives_completeness ON public.students_learning_objectives_completeness USING (public.permission_check(resource_path, 'students_learning_objectives_completeness'::text)) WITH CHECK (public.permission_check(resource_path, 'students_learning_objectives_completeness'::text));


--
-- Name: students_learning_objectives_completeness rls_students_learning_objectives_completeness_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_students_learning_objectives_completeness_restrictive ON public.students_learning_objectives_completeness AS RESTRICTIVE USING (public.permission_check(resource_path, 'students_learning_objectives_completeness'::text)) WITH CHECK (public.permission_check(resource_path, 'students_learning_objectives_completeness'::text));


--
-- Name: students_topics_completeness rls_students_topics_completeness; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_students_topics_completeness ON public.students_topics_completeness USING (public.permission_check(resource_path, 'students_topics_completeness'::text)) WITH CHECK (public.permission_check(resource_path, 'students_topics_completeness'::text));


--
-- Name: students_topics_completeness rls_students_topics_completeness_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_students_topics_completeness_restrictive ON public.students_topics_completeness AS RESTRICTIVE USING (public.permission_check(resource_path, 'students_topics_completeness'::text)) WITH CHECK (public.permission_check(resource_path, 'students_topics_completeness'::text));


--
-- Name: study_plan_assessment rls_study_plan_assessment; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_study_plan_assessment ON public.study_plan_assessment USING (public.permission_check(resource_path, 'study_plan_assessment'::text)) WITH CHECK (public.permission_check(resource_path, 'study_plan_assessment'::text));


--
-- Name: study_plan_assessment rls_study_plan_assessment_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_study_plan_assessment_restrictive ON public.study_plan_assessment AS RESTRICTIVE USING (public.permission_check(resource_path, 'study_plan_assessment'::text)) WITH CHECK (public.permission_check(resource_path, 'study_plan_assessment'::text));


--
-- Name: study_plan_items rls_study_plan_items; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_study_plan_items ON public.study_plan_items USING (public.permission_check(resource_path, 'study_plan_items'::text)) WITH CHECK (public.permission_check(resource_path, 'study_plan_items'::text));


--
-- Name: study_plan_items rls_study_plan_items_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_study_plan_items_restrictive ON public.study_plan_items AS RESTRICTIVE USING (public.permission_check(resource_path, 'study_plan_items'::text)) WITH CHECK (public.permission_check(resource_path, 'study_plan_items'::text));


--
-- Name: study_plan_monitors rls_study_plan_monitors; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_study_plan_monitors ON public.study_plan_monitors USING (public.permission_check(resource_path, 'study_plan_monitors'::text)) WITH CHECK (public.permission_check(resource_path, 'study_plan_monitors'::text));


--
-- Name: study_plan_monitors rls_study_plan_monitors_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_study_plan_monitors_restrictive ON public.study_plan_monitors AS RESTRICTIVE USING (public.permission_check(resource_path, 'study_plan_monitors'::text)) WITH CHECK (public.permission_check(resource_path, 'study_plan_monitors'::text));


--
-- Name: study_plans rls_study_plans; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_study_plans ON public.study_plans USING (public.permission_check(resource_path, 'study_plans'::text)) WITH CHECK (public.permission_check(resource_path, 'study_plans'::text));


--
-- Name: study_plans rls_study_plans_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_study_plans_restrictive ON public.study_plans AS RESTRICTIVE USING (public.permission_check(resource_path, 'study_plans'::text)) WITH CHECK (public.permission_check(resource_path, 'study_plans'::text));


--
-- Name: tagged_user rls_tagged_user; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_tagged_user ON public.tagged_user USING (public.permission_check(resource_path, 'tagged_user'::text)) WITH CHECK (public.permission_check(resource_path, 'tagged_user'::text));


--
-- Name: tagged_user rls_tagged_user_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_tagged_user_restrictive ON public.tagged_user AS RESTRICTIVE USING (public.permission_check(resource_path, 'tagged_user'::text)) WITH CHECK (public.permission_check(resource_path, 'tagged_user'::text));


--
-- Name: task_assignment rls_task_assignment; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_task_assignment ON public.task_assignment USING (public.permission_check(resource_path, 'task_assignment'::text)) WITH CHECK (public.permission_check(resource_path, 'task_assignment'::text));


--
-- Name: task_assignment rls_task_assignment_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_task_assignment_restrictive ON public.task_assignment AS RESTRICTIVE USING (public.permission_check(resource_path, 'task_assignment'::text)) WITH CHECK (public.permission_check(resource_path, 'task_assignment'::text));


--
-- Name: topics rls_topics; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_topics ON public.topics USING (public.permission_check(resource_path, 'topics'::text)) WITH CHECK (public.permission_check(resource_path, 'topics'::text));


--
-- Name: topics_assignments rls_topics_assignments; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_topics_assignments ON public.topics_assignments USING (public.permission_check(resource_path, 'topics_assignments'::text)) WITH CHECK (public.permission_check(resource_path, 'topics_assignments'::text));


--
-- Name: topics_assignments rls_topics_assignments_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_topics_assignments_restrictive ON public.topics_assignments AS RESTRICTIVE USING (public.permission_check(resource_path, 'topics_assignments'::text)) WITH CHECK (public.permission_check(resource_path, 'topics_assignments'::text));


--
-- Name: topics_learning_objectives rls_topics_learning_objectives; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_topics_learning_objectives ON public.topics_learning_objectives USING (public.permission_check(resource_path, 'topics_learning_objectives'::text)) WITH CHECK (public.permission_check(resource_path, 'topics_learning_objectives'::text));


--
-- Name: topics_learning_objectives rls_topics_learning_objectives_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_topics_learning_objectives_restrictive ON public.topics_learning_objectives AS RESTRICTIVE USING (public.permission_check(resource_path, 'topics_learning_objectives'::text)) WITH CHECK (public.permission_check(resource_path, 'topics_learning_objectives'::text));


--
-- Name: topics rls_topics_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_topics_restrictive ON public.topics AS RESTRICTIVE USING (public.permission_check(resource_path, 'topics'::text)) WITH CHECK (public.permission_check(resource_path, 'topics'::text));


--
-- Name: user_access_paths rls_user_access_paths; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_user_access_paths ON public.user_access_paths USING (public.permission_check(resource_path, 'user_access_paths'::text)) WITH CHECK (public.permission_check(resource_path, 'user_access_paths'::text));


--
-- Name: user_group rls_user_group; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_user_group ON public.user_group USING (public.permission_check(resource_path, 'user_group'::text)) WITH CHECK (public.permission_check(resource_path, 'user_group'::text));


--
-- Name: user_group_member rls_user_group_member; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_user_group_member ON public.user_group_member USING (public.permission_check(resource_path, 'user_group_member'::text)) WITH CHECK (public.permission_check(resource_path, 'user_group_member'::text));


--
-- Name: user_group_member rls_user_group_member_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_user_group_member_restrictive ON public.user_group_member AS RESTRICTIVE USING (public.permission_check(resource_path, 'user_group_member'::text)) WITH CHECK (public.permission_check(resource_path, 'user_group_member'::text));


--
-- Name: user_group rls_user_group_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_user_group_restrictive ON public.user_group AS RESTRICTIVE USING (public.permission_check(resource_path, 'user_group'::text)) WITH CHECK (public.permission_check(resource_path, 'user_group'::text));


--
-- Name: user_tag rls_user_tag; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_user_tag ON public.user_tag USING (public.permission_check(resource_path, 'user_tag'::text)) WITH CHECK (public.permission_check(resource_path, 'user_tag'::text));


--
-- Name: user_tag rls_user_tag_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_user_tag_restrictive ON public.user_tag AS RESTRICTIVE USING (public.permission_check(resource_path, 'user_tag'::text)) WITH CHECK (public.permission_check(resource_path, 'user_tag'::text));


--
-- Name: users rls_users; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_users ON public.users USING (public.permission_check(resource_path, 'users'::text)) WITH CHECK (public.permission_check(resource_path, 'users'::text));


--
-- Name: users_groups rls_users_groups; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_users_groups ON public.users_groups USING (public.permission_check(resource_path, 'users_groups'::text)) WITH CHECK (public.permission_check(resource_path, 'users_groups'::text));


--
-- Name: users_groups rls_users_groups_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_users_groups_restrictive ON public.users_groups AS RESTRICTIVE USING (public.permission_check(resource_path, 'users_groups'::text)) WITH CHECK (public.permission_check(resource_path, 'users_groups'::text));


--
-- Name: users rls_users_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_users_restrictive ON public.users AS RESTRICTIVE USING (public.permission_check(resource_path, 'users'::text)) WITH CHECK (public.permission_check(resource_path, 'users'::text));


--
-- Name: withus_failed_sync_email_recipient rls_withus_failed_sync_email_recipient; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_withus_failed_sync_email_recipient ON public.withus_failed_sync_email_recipient USING (public.permission_check(resource_path, 'withus_failed_sync_email_recipient'::text)) WITH CHECK (public.permission_check(resource_path, 'withus_failed_sync_email_recipient'::text));


--
-- Name: withus_failed_sync_email_recipient rls_withus_failed_sync_email_recipient_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_withus_failed_sync_email_recipient_restrictive ON public.withus_failed_sync_email_recipient AS RESTRICTIVE USING (public.permission_check(resource_path, 'withus_failed_sync_email_recipient'::text)) WITH CHECK (public.permission_check(resource_path, 'withus_failed_sync_email_recipient'::text));


--
-- Name: withus_mapping_course_id rls_withus_mapping_course_id; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_withus_mapping_course_id ON public.withus_mapping_course_id USING (public.permission_check(resource_path, 'withus_mapping_course_id'::text)) WITH CHECK (public.permission_check(resource_path, 'withus_mapping_course_id'::text));


--
-- Name: withus_mapping_course_id rls_withus_mapping_course_id_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_withus_mapping_course_id_restrictive ON public.withus_mapping_course_id AS RESTRICTIVE USING (public.permission_check(resource_path, 'withus_mapping_course_id'::text)) WITH CHECK (public.permission_check(resource_path, 'withus_mapping_course_id'::text));


--
-- Name: withus_mapping_exam_lo_id rls_withus_mapping_exam_lo_id; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_withus_mapping_exam_lo_id ON public.withus_mapping_exam_lo_id USING (public.permission_check(resource_path, 'withus_mapping_exam_lo_id'::text)) WITH CHECK (public.permission_check(resource_path, 'withus_mapping_exam_lo_id'::text));


--
-- Name: withus_mapping_exam_lo_id rls_withus_mapping_exam_lo_id_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_withus_mapping_exam_lo_id_restrictive ON public.withus_mapping_exam_lo_id AS RESTRICTIVE USING (public.permission_check(resource_path, 'withus_mapping_exam_lo_id'::text)) WITH CHECK (public.permission_check(resource_path, 'withus_mapping_exam_lo_id'::text));


--
-- Name: withus_mapping_question_tag rls_withus_mapping_question_tag; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_withus_mapping_question_tag ON public.withus_mapping_question_tag USING (public.permission_check(resource_path, 'withus_mapping_question_tag'::text)) WITH CHECK (public.permission_check(resource_path, 'withus_mapping_question_tag'::text));


--
-- Name: withus_mapping_question_tag rls_withus_mapping_question_tag_restrictive; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY rls_withus_mapping_question_tag_restrictive ON public.withus_mapping_question_tag AS RESTRICTIVE USING (public.permission_check(resource_path, 'withus_mapping_question_tag'::text)) WITH CHECK (public.permission_check(resource_path, 'withus_mapping_question_tag'::text));


--
-- Name: role; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.role ENABLE ROW LEVEL SECURITY;

--
-- Name: shuffled_quiz_sets; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.shuffled_quiz_sets ENABLE ROW LEVEL SECURITY;

--
-- Name: student_latest_submissions; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.student_latest_submissions ENABLE ROW LEVEL SECURITY;

--
-- Name: student_learning_time_by_daily; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.student_learning_time_by_daily ENABLE ROW LEVEL SECURITY;

--
-- Name: student_study_plans; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.student_study_plans ENABLE ROW LEVEL SECURITY;

--
-- Name: student_submission_grades; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.student_submission_grades ENABLE ROW LEVEL SECURITY;

--
-- Name: student_submissions; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.student_submissions ENABLE ROW LEVEL SECURITY;

--
-- Name: students; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.students ENABLE ROW LEVEL SECURITY;

--
-- Name: students_learning_objectives_completeness; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.students_learning_objectives_completeness ENABLE ROW LEVEL SECURITY;

--
-- Name: students_topics_completeness; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.students_topics_completeness ENABLE ROW LEVEL SECURITY;

--
-- Name: study_plan_assessment; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.study_plan_assessment ENABLE ROW LEVEL SECURITY;

--
-- Name: study_plan_items; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.study_plan_items ENABLE ROW LEVEL SECURITY;

--
-- Name: study_plan_monitors; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.study_plan_monitors ENABLE ROW LEVEL SECURITY;

--
-- Name: study_plans; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.study_plans ENABLE ROW LEVEL SECURITY;

--
-- Name: tagged_user; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.tagged_user ENABLE ROW LEVEL SECURITY;

--
-- Name: task_assignment; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.task_assignment ENABLE ROW LEVEL SECURITY;

--
-- Name: topics; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.topics ENABLE ROW LEVEL SECURITY;

--
-- Name: topics_assignments; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.topics_assignments ENABLE ROW LEVEL SECURITY;

--
-- Name: topics_learning_objectives; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.topics_learning_objectives ENABLE ROW LEVEL SECURITY;

--
-- Name: user_access_paths; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.user_access_paths ENABLE ROW LEVEL SECURITY;

--
-- Name: user_group; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.user_group ENABLE ROW LEVEL SECURITY;

--
-- Name: user_group_member; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.user_group_member ENABLE ROW LEVEL SECURITY;

--
-- Name: user_tag; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.user_tag ENABLE ROW LEVEL SECURITY;

--
-- Name: users; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;

--
-- Name: users_groups; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.users_groups ENABLE ROW LEVEL SECURITY;

--
-- Name: withus_failed_sync_email_recipient; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.withus_failed_sync_email_recipient ENABLE ROW LEVEL SECURITY;

--
-- Name: withus_mapping_course_id; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.withus_mapping_course_id ENABLE ROW LEVEL SECURITY;

--
-- Name: withus_mapping_exam_lo_id; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.withus_mapping_exam_lo_id ENABLE ROW LEVEL SECURITY;

--
-- Name: withus_mapping_question_tag; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.withus_mapping_question_tag ENABLE ROW LEVEL SECURITY;

--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: postgres
--

GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

