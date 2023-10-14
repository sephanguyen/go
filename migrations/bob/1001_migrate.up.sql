--
-- PostgreSQL database dump
--

-- Dumped from database version 11.6
-- Dumped by pg_dump version 11.7 (Ubuntu 11.7-0ubuntu0.19.10.1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET search_path TO public;
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA IF NOT EXISTS public;


--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA public IS 'standard public schema';


SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    user_id text NOT NULL,
    country text NOT NULL,
    name text NOT NULL,
    avatar text,
    phone_number text NOT NULL,
    email text,
    device_token text,
    allow_notification boolean,
    user_group text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    is_tester boolean,
    facebook_id text,
    platform text,
    phone_verified boolean,
    email_verified boolean,
    CONSTRAINT users_platform_check CHECK ((platform = ANY ('{PLATFORM_NONE,PLATFORM_IOS,PLATFORM_ANDROID}'::text[]))),
    CONSTRAINT users_user_group_check CHECK ((user_group = ANY (ARRAY['USER_GROUP_STUDENT'::text, 'USER_GROUP_COACH'::text, 'USER_GROUP_TUTOR'::text, 'USER_GROUP_STAFF'::text, 'USER_GROUP_ADMIN'::text, 'USER_GROUP_TEACHER'::text, 'USER_GROUP_PARENT'::text, 'USER_GROUP_CONTENT_ADMIN'::text, 'USER_GROUP_CONTENT_STAFF'::text, 'USER_GROUP_SALES_ADMIN'::text, 'USER_GROUP_SALES_STAFF'::text, 'USER_GROUP_CS_ADMIN'::text, 'USER_GROUP_CS_STAFF'::text, 'USER_GROUP_SCHOOL_ADMIN'::text, 'USER_GROUP_SCHOOL_STAFF'::text])))
);


--
-- Name: COLUMN users.is_tester; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.users.is_tester IS 'to distinguish our staff using app as a student or tester testing app as coach, tutor';


--
-- Name: find_teacher_by_school_id(integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.find_teacher_by_school_id(school_id integer) RETURNS SETOF public.users
    LANGUAGE sql STABLE
    AS $$
    select u.* from  teachers t join users u on u.user_id = t.teacher_id where t.school_ids @> ARRAY[school_id]
$$;


--
-- Name: getdate(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.getdate() RETURNS timestamp with time zone
    LANGUAGE sql STABLE
    AS $$select now()$$;


--
-- Name: a; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.a AS
 SELECT users.name,
    users.updated_at,
    count(users.updated_at) OVER (PARTITION BY users.updated_at ORDER BY users.updated_at DESC) AS count
   FROM public.users;


--
-- Name: activity_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activity_logs (
    activity_log_id text NOT NULL,
    user_id text NOT NULL,
    action_type text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    payload jsonb
);


--
-- Name: assignments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.assignments (
    assignment_id text NOT NULL,
    assignment_type text DEFAULT 'ASSIGNMENT_TYPE_NONE'::text,
    assigned_by text NOT NULL,
    topic_id text NOT NULL,
    preset_study_plan_id text,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    preset_study_plan_weekly_id text,
    class_id integer
);


--
-- Name: billing_histories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.billing_histories (
    billing_id text NOT NULL,
    student_id text NOT NULL,
    generator_id text NOT NULL,
    generator_email text NOT NULL,
    billing_from timestamp with time zone NOT NULL,
    billing_to timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: chapters; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chapters (
    chapter_id text NOT NULL,
    name text NOT NULL,
    country text,
    subject text,
    grade smallint NOT NULL,
    display_order smallint DEFAULT 0,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    school_id integer DEFAULT '-2147483648'::integer NOT NULL
);


--
-- Name: cities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cities (
    city_id integer NOT NULL,
    name text NOT NULL,
    country text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    display_order smallint DEFAULT 0 NOT NULL
);


--
-- Name: cities_city_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.cities_city_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: cities_city_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.cities_city_id_seq OWNED BY public.cities.city_id;


--
-- Name: class_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.class_members (
    class_member_id text NOT NULL,
    class_id integer NOT NULL,
    user_id text NOT NULL,
    status text DEFAULT 'CLASS_MEMBER_STATUS_NONE'::text NOT NULL,
    user_group text NOT NULL,
    is_owner boolean DEFAULT false NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    student_subscription_id text,
    CONSTRAINT class_members_status_check CHECK ((status = ANY (ARRAY['CLASS_MEMBER_STATUS_NONE'::text, 'CLASS_MEMBER_STATUS_ACTIVE'::text, 'CLASS_MEMBER_STATUS_INACTIVE'::text]))),
    CONSTRAINT class_members_user_group_check CHECK ((user_group = ANY (ARRAY['USER_GROUP_STUDENT'::text, 'USER_GROUP_COACH'::text, 'USER_GROUP_TUTOR'::text, 'USER_GROUP_STAFF'::text, 'USER_GROUP_ADMIN'::text, 'USER_GROUP_TEACHER'::text, 'USER_GROUP_PARENT'::text, 'USER_GROUP_CONTENT_ADMIN'::text, 'USER_GROUP_CONTENT_STAFF'::text, 'USER_GROUP_SALES_ADMIN'::text, 'USER_GROUP_SALES_STAFF'::text, 'USER_GROUP_CS_ADMIN'::text, 'USER_GROUP_CS_STAFF'::text, 'USER_GROUP_SCHOOL_ADMIN'::text, 'USER_GROUP_SCHOOL_STAFF'::text])))
);


--
-- Name: class_preset_study_plans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.class_preset_study_plans (
    class_preset_study_plan_id text NOT NULL,
    class_id integer NOT NULL,
    preset_study_plan_id text NOT NULL,
    deleted_at timestamp with time zone,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: classes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.classes (
    class_id integer NOT NULL,
    school_id integer NOT NULL,
    avatar text NOT NULL,
    name text NOT NULL,
    subjects text[],
    grades integer[],
    status text DEFAULT 'CLASS_STATUS_NONE'::text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    plan_id text,
    country text,
    plan_expired_at timestamp with time zone,
    plan_duration smallint,
    class_code text,
    CONSTRAINT classes_status_check CHECK ((status = ANY (ARRAY['CLASS_STATUS_NONE'::text, 'CLASS_STATUS_ACTIVE'::text, 'CLASS_STATUS_INACTIVE'::text]))),
    CONSTRAINT classes_subjects_check CHECK ((subjects <@ ARRAY['SUBJECT_MATHS'::text, 'SUBJECT_BIOLOGY'::text, 'SUBJECT_PHYSICS'::text, 'SUBJECT_CHEMISTRY'::text, 'SUBJECT_GEOGRAPHY'::text, 'SUBJECT_ENGLISH'::text, 'SUBJECT_ENGLISH_2'::text]))
);


--
-- Name: classes_class_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.classes_class_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: classes_class_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.classes_class_id_seq OWNED BY public.classes.class_id;


--
-- Name: classes_school_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.classes_school_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: classes_school_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.classes_school_id_seq OWNED BY public.classes.school_id;


--
-- Name: coaches; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.coaches (
    coach_id text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: cod_orders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cod_orders (
    cod_order_id text NOT NULL,
    student_order_id integer NOT NULL,
    cod_secret_code text NOT NULL,
    customer_name text NOT NULL,
    customer_phone_number text NOT NULL,
    customer_address text NOT NULL,
    address_tree text[] NOT NULL,
    status text DEFAULT 'COD_ORDER_STATUS_NONE'::text NOT NULL,
    expected_delivery_time timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: cod_orders_student_order_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.cod_orders_student_order_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: cod_orders_student_order_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.cod_orders_student_order_id_seq OWNED BY public.cod_orders.student_order_id;


--
-- Name: configs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.configs (
    config_key text NOT NULL,
    config_group text NOT NULL,
    country text NOT NULL,
    config_value text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: courses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.courses (
    course_id text NOT NULL,
    name text NOT NULL,
    country text,
    subject text,
    grade smallint NOT NULL,
    display_order smallint DEFAULT 0,
    chapter_ids text[],
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    school_id integer DEFAULT '-2147483648'::integer NOT NULL
);


--
-- Name: courses_classes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.courses_classes (
    course_id text NOT NULL,
    class_id integer NOT NULL,
    status text DEFAULT 'COURSE_CLASS_STATUS_ACTIVE'::text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT courses_classes_status_check CHECK ((status = ANY (ARRAY['COURSE_CLASS_STATUS_ACTIVE'::text, 'COURSE_CLASS_STATUS_INACTIVE'::text])))
);


--
-- Name: districts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.districts (
    district_id integer NOT NULL,
    name text NOT NULL,
    country text NOT NULL,
    city_id integer NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: districts_district_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.districts_district_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: districts_district_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.districts_district_id_seq OWNED BY public.districts.district_id;


--
-- Name: hub_tours; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.hub_tours (
    hub_id integer NOT NULL,
    student_id text NOT NULL,
    parent_phone_number character varying(100),
    status character varying(50) DEFAULT 'new'::character varying NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: hubs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.hubs (
    hub_id integer NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    phone_number character varying(100),
    address text,
    country text NOT NULL,
    city_id integer,
    district_id integer,
    point point,
    images text[],
    opening_hours text[],
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    events jsonb
);


--
-- Name: hubs_hub_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.hubs_hub_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: hubs_hub_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.hubs_hub_id_seq OWNED BY public.hubs.hub_id;


--
-- Name: ios_transactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ios_transactions (
    ios_transaction_id text NOT NULL,
    student_id text NOT NULL,
    receipt_data text NOT NULL,
    status text NOT NULL,
    is_manual_verify boolean NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: learning_objectives; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.learning_objectives (
    lo_id text NOT NULL,
    name text NOT NULL,
    country text,
    grade smallint NOT NULL,
    subject text NOT NULL,
    topic_id text,
    master_lo_id text,
    display_order smallint,
    prerequisites text[],
    video text,
    study_guide text,
    video_script text,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    school_id integer DEFAULT '-2147483648'::integer NOT NULL
);


--
-- Name: notification_messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notification_messages (
    notification_message_id integer NOT NULL,
    country text NOT NULL,
    key character varying(255) NOT NULL,
    receiver_group character varying(100) NOT NULL,
    title text,
    body text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: notification_messages_notification_message_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.notification_messages_notification_message_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: notification_messages_notification_message_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.notification_messages_notification_message_id_seq OWNED BY public.notification_messages.notification_message_id;


--
-- Name: notification_targets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notification_targets (
    target_id text NOT NULL,
    name text NOT NULL,
    conditions json,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    notification_id text NOT NULL,
    title text NOT NULL,
    description text NOT NULL,
    type text NOT NULL,
    data jsonb,
    target text,
    schedule_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT notifications_type_check CHECK ((type = ANY ('{NOTIFICATION_TYPE_PROMO_NONE,NOTIFICATION_TYPE_TEXT,NOTIFICATION_TYPE_PROMO_CODE}'::text[])))
);


--
-- Name: package_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.package_items (
    package_item_id text NOT NULL,
    package_id integer NOT NULL,
    plan_id text NOT NULL,
    country text,
    expired_at timestamp with time zone,
    duration smallint,
    subject text[],
    grades integer[],
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    price integer,
    discounted_price integer
);


--
-- Name: COLUMN package_items.duration; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.package_items.duration IS 'unit is day';


--
-- Name: packages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.packages (
    package_id integer NOT NULL,
    country text,
    name text NOT NULL,
    description text[],
    price integer NOT NULL,
    discounted_price integer,
    prioritize_level smallint DEFAULT 0,
    upgradable_from integer[],
    is_recommended boolean NOT NULL,
    is_enabled boolean NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    ios_bundle_id text
);


--
-- Name: packages_package_id_seq1; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.packages_package_id_seq1
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: packages_package_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.packages_package_id_seq1 OWNED BY public.packages.package_id;


--
-- Name: plans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.plans (
    plan_id text NOT NULL,
    country text NOT NULL,
    description text,
    plan_privileges text[] NOT NULL,
    is_purchasable boolean NOT NULL,
    prioritize_level smallint DEFAULT 0,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    benefits text[]
);


--
-- Name: prac; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.prac AS
 SELECT users.name,
    users.updated_at,
    count(users.updated_at) OVER (PARTITION BY users.updated_at ORDER BY users.updated_at DESC) AS count
   FROM public.users;


--
-- Name: preset_study_plans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.preset_study_plans (
    preset_study_plan_id text NOT NULL,
    name text NOT NULL,
    country text,
    grade smallint NOT NULL,
    subject text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    start_date timestamp with time zone
);


--
-- Name: preset_study_plans_weekly; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.preset_study_plans_weekly (
    preset_study_plan_weekly_id text NOT NULL,
    preset_study_plan_id text NOT NULL,
    topic_id text NOT NULL,
    week smallint NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: promotion_rules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.promotion_rules (
    promotion_rule_id integer NOT NULL,
    promotion_id integer NOT NULL,
    promo_type text NOT NULL,
    discount_type text,
    discount_amount numeric(12,2) NOT NULL,
    conditions jsonb,
    rewards jsonb,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: promotion_rules_promotion_rule_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.promotion_rules_promotion_rule_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: promotion_rules_promotion_rule_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.promotion_rules_promotion_rule_id_seq OWNED BY public.promotion_rules.promotion_rule_id;


--
-- Name: promotions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.promotions (
    promotion_id integer NOT NULL,
    owner_id text NOT NULL,
    prefix_sequence_number integer DEFAULT 1 NOT NULL,
    country text NOT NULL,
    code_prefix character varying(100) NOT NULL,
    code character varying(100) NOT NULL,
    started_date timestamp with time zone,
    expired_date timestamp with time zone,
    status text NOT NULL,
    redemption_limit_per_code integer DEFAULT 0 NOT NULL,
    redemption_limit_per_user integer DEFAULT 0 NOT NULL,
    total_redemptions integer DEFAULT 0 NOT NULL,
    notes text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: promotions_promotion_id_seq1; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.promotions_promotion_id_seq1
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: promotions_promotion_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.promotions_promotion_id_seq1 OWNED BY public.promotions.promotion_id;


--
-- Name: questions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.questions (
    question_id text NOT NULL,
    country text NOT NULL,
    master_question_id text,
    question text NOT NULL,
    answers text[],
    explanation text,
    difficulty_level smallint DEFAULT 1 NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    question_rendered text,
    answers_rendered text[],
    explanation_rendered text,
    is_waiting_for_render boolean,
    explanation_wrong_answer text[],
    explanation_wrong_answer_rendered text[]
);


--
-- Name: questions_tagged_learning_objectives; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.questions_tagged_learning_objectives (
    question_id text NOT NULL,
    lo_id text NOT NULL,
    display_order integer DEFAULT 0 NOT NULL
);


--
-- Name: quizsets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.quizsets (
    lo_id text NOT NULL,
    question_id text NOT NULL,
    display_order integer,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: school_admins; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.school_admins (
    school_admin_id text NOT NULL,
    school_id integer NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: school_configs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.school_configs (
    school_id integer NOT NULL,
    plan_id text NOT NULL,
    country text,
    plan_expired_at timestamp with time zone,
    plan_duration smallint,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    privileges text[],
    CONSTRAINT school_configs_privileges_check CHECK ((privileges <@ ARRAY['CAN_ACCESS_LEARNING_TOPICS'::text, 'CAN_ACCESS_PRACTICE_TOPICS'::text, 'CAN_ACCESS_MOCK_EXAMS'::text, 'CAN_ACCESS_ALL_LOS'::text, 'CAN_ACCESS_SOME_LOS'::text, 'CAN_WATCH_VIDEOS'::text, 'CAN_READ_STUDY_GUIDES'::text, 'CAN_SKIP_VIDEOS'::text, 'CAN_CHAT_WITH_TEACHER'::text])),
    CONSTRAINT school_configs_privileges_check1 CHECK ((privileges <@ ARRAY['CAN_ACCESS_LEARNING_TOPICS'::text, 'CAN_ACCESS_PRACTICE_TOPICS'::text, 'CAN_ACCESS_ALL_LOS'::text, 'CAN_WATCH_VIDEOS'::text, 'CAN_READ_STUDY_GUIDES'::text, 'CAN_CHAT_WITH_TEACHER'::text])),
    CONSTRAINT school_configs_privileges_check2 CHECK ((privileges <@ ARRAY['CAN_ACCESS_LEARNING_TOPICS'::text, 'CAN_ACCESS_PRACTICE_TOPICS'::text, 'CAN_ACCESS_ALL_LOS'::text, 'CAN_WATCH_VIDEOS'::text, 'CAN_READ_STUDY_GUIDES'::text, 'CAN_CHAT_WITH_TEACHER'::text]))
);


--
-- Name: school_configs_school_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.school_configs_school_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: school_configs_school_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.school_configs_school_id_seq OWNED BY public.school_configs.school_id;


--
-- Name: schools; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schools (
    school_id integer NOT NULL,
    name text NOT NULL,
    country text NOT NULL,
    city_id integer NOT NULL,
    district_id integer NOT NULL,
    point point,
    is_system_school boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    is_merge boolean DEFAULT false,
    phone_number text
);


--
-- Name: schools_school_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.schools_school_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: schools_school_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.schools_school_id_seq OWNED BY public.schools.school_id;


--
-- Name: student_assignments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.student_assignments (
    student_id text NOT NULL,
    assignment_id text NOT NULL,
    assignment_status text DEFAULT 'STUDENT_ASSIGNMENT_STATUS_ACTIVE'::text,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    completed_at timestamp with time zone
);


--
-- Name: student_comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.student_comments (
    comment_id text NOT NULL,
    student_id text NOT NULL,
    coach_id text NOT NULL,
    comment_content text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: student_event_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.student_event_logs (
    student_event_log_id integer NOT NULL,
    student_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    event_type character varying(100) NOT NULL,
    payload jsonb,
    event_id character varying(50)
);


--
-- Name: student_event_logs_student_event_log_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.student_event_logs_student_event_log_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: student_event_logs_student_event_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.student_event_logs_student_event_log_id_seq OWNED BY public.student_event_logs.student_event_log_id;


--
-- Name: student_learning_time_by_daily; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.student_learning_time_by_daily (
    learning_time_id integer NOT NULL,
    student_id text NOT NULL,
    learning_time integer DEFAULT 0 NOT NULL,
    day timestamp with time zone NOT NULL,
    sessions text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: COLUMN student_learning_time_by_daily.learning_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.student_learning_time_by_daily.learning_time IS 'learning time in seconds unit';


--
-- Name: student_learning_time_by_daily_learning_time_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.student_learning_time_by_daily_learning_time_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: student_learning_time_by_daily_learning_time_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.student_learning_time_by_daily_learning_time_id_seq OWNED BY public.student_learning_time_by_daily.learning_time_id;


--
-- Name: student_orders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.student_orders (
    student_order_id integer NOT NULL,
    amount numeric(12,2) NOT NULL,
    currency text NOT NULL,
    payment_method text,
    student_id text NOT NULL,
    package_id integer,
    package_name text,
    coupon text,
    coupon_amount numeric(12,2),
    gateway_response text,
    gateway_full_feedback text,
    gateway_link text,
    country text,
    gateway_name text,
    is_manual_created boolean,
    created_by_email text,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    status text,
    ios_transaction_id text,
    inapp_transaction_id text,
    reference_number text,
    CONSTRAINT student_order_status CHECK ((status = ANY (ARRAY['ORDER_STATUS_NONE'::text, 'ORDER_STATUS_WAITING_FOR_PAYMENT'::text, 'ORDER_STATUS_PROCESSING_PAYMENT'::text, 'ORDER_STATUS_SUCCESSFULLY'::text, 'ORDER_STATUS_FAILED'::text, 'ORDER_STATUS_CANCELED'::text, 'ORDER_STATUS_DELETED'::text, 'ORDER_STATUS_DISABLED'::text])))
);


--
-- Name: student_orders_student_order_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.student_orders_student_order_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: student_orders_student_order_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.student_orders_student_order_id_seq OWNED BY public.student_orders.student_order_id;


--
-- Name: student_questions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.student_questions (
    student_question_id text NOT NULL,
    student_id text NOT NULL,
    tutor_id text,
    quiz_id text NOT NULL,
    content text NOT NULL,
    url_medias text[],
    history_assigned_tutor_ids text[],
    status text DEFAULT 'QUESTION_STATUS_WAITING_FOR_ASSIGN'::text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    rate text,
    rate_at timestamp with time zone,
    history_changed_status text[],
    is_processing boolean DEFAULT false
);


--
-- Name: student_statistics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.student_statistics (
    student_id text NOT NULL,
    total_lo_finished integer DEFAULT 0,
    total_learning_time integer DEFAULT 0,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    last_time_completed_lo timestamp with time zone,
    additional_data jsonb
);


--
-- Name: student_submission_scores; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.student_submission_scores (
    student_submission_score_id text NOT NULL,
    teacher_id text NOT NULL,
    student_submission_id text NOT NULL,
    given_score numeric NOT NULL,
    total_score numeric NOT NULL,
    created_at timestamp with time zone NOT NULL,
    notes text
);


--
-- Name: student_submissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.student_submissions (
    student_submission_id text NOT NULL,
    student_id text NOT NULL,
    topic_id text NOT NULL,
    content text,
    attachment_names text[],
    attachment_urls text[],
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: student_subscriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.student_subscriptions (
    student_subscription_id text NOT NULL,
    student_order_id integer,
    student_id text NOT NULL,
    plan_id text NOT NULL,
    country text,
    start_time timestamp with time zone NOT NULL,
    end_time timestamp with time zone NOT NULL,
    subject text[],
    grades integer[],
    status text DEFAULT 'SUBSCRIPTION_STATUS_NONE'::text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    amount numeric(12,2),
    coupon_amount numeric(12,2),
    extend_from text
);


--
-- Name: students; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.students (
    student_id text NOT NULL,
    current_grade smallint,
    target_university text,
    on_trial boolean DEFAULT true NOT NULL,
    billing_date timestamp with time zone NOT NULL,
    birthday date,
    biography text,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    total_question_limit smallint DEFAULT 20,
    school_id integer
);


--
-- Name: COLUMN students.billing_date; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.students.billing_date IS 'students need to pay before this day';


--
-- Name: students_assigned_coaches; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.students_assigned_coaches (
    student_id text NOT NULL,
    coach_id text NOT NULL,
    is_active boolean DEFAULT false NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: students_learning_objectives_completeness; Type: TABLE; Schema: public; Owner: -
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
    highest_quiz_score real
);


--
-- Name: students_study_plans_weekly; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.students_study_plans_weekly (
    preset_study_plan_weekly_id text NOT NULL,
    student_id text NOT NULL,
    start_date timestamp with time zone,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: students_topics_completeness; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.students_topics_completeness (
    student_id text NOT NULL,
    topic_id text NOT NULL,
    total_finished_los integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    is_completed boolean DEFAULT false
);


--
-- Name: students_topics_overdue; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.students_topics_overdue (
    topic_id text NOT NULL,
    student_id text NOT NULL,
    due_date timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: teachers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.teachers (
    teacher_id text NOT NULL,
    school_ids integer[],
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    school_name text
);


--
-- Name: teacher_by_school_id; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.teacher_by_school_id AS
 SELECT unnest(t.school_ids) AS school_id,
    t.teacher_id,
    t.created_at,
    t.updated_at
   FROM public.teachers t;


--
-- Name: topics; Type: TABLE; Schema: public; Owner: -
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
    attachment_names text[]
);


--
-- Name: tutors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tutors (
    tutor_id text NOT NULL,
    skill_set text[] NOT NULL,
    status text DEFAULT 'TUTOR_STATUS_NONE'::text,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    grades smallint[],
    current_active_questions smallint DEFAULT 0,
    open_questions smallint DEFAULT 0,
    total_resolved_questions integer DEFAULT 0
);


--
-- Name: user_notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_notifications (
    user_notification_id text NOT NULL,
    notification_id text NOT NULL,
    status text DEFAULT 'USER_NOTIFICATION_STATUS_NEW'::text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    user_id text NOT NULL,
    CONSTRAINT user_notifications_status_check CHECK ((status = ANY ('{USER_NOTIFICATION_STATUS_NEW,USER_NOTIFICATION_STATUS_SEEN,USER_NOTIFICATION_STATUS_READ,USER_NOTIFICATION_STATUS_FAILED}'::text[])))
);


--
-- Name: cities city_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cities ALTER COLUMN city_id SET DEFAULT nextval('public.cities_city_id_seq'::regclass);


--
-- Name: classes class_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.classes ALTER COLUMN class_id SET DEFAULT nextval('public.classes_class_id_seq'::regclass);


--
-- Name: classes school_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.classes ALTER COLUMN school_id SET DEFAULT nextval('public.classes_school_id_seq'::regclass);


--
-- Name: cod_orders student_order_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cod_orders ALTER COLUMN student_order_id SET DEFAULT nextval('public.cod_orders_student_order_id_seq'::regclass);


--
-- Name: districts district_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.districts ALTER COLUMN district_id SET DEFAULT nextval('public.districts_district_id_seq'::regclass);


--
-- Name: hubs hub_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.hubs ALTER COLUMN hub_id SET DEFAULT nextval('public.hubs_hub_id_seq'::regclass);


--
-- Name: notification_messages notification_message_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_messages ALTER COLUMN notification_message_id SET DEFAULT nextval('public.notification_messages_notification_message_id_seq'::regclass);


--
-- Name: packages package_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.packages ALTER COLUMN package_id SET DEFAULT nextval('public.packages_package_id_seq1'::regclass);


--
-- Name: promotion_rules promotion_rule_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promotion_rules ALTER COLUMN promotion_rule_id SET DEFAULT nextval('public.promotion_rules_promotion_rule_id_seq'::regclass);


--
-- Name: promotions promotion_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promotions ALTER COLUMN promotion_id SET DEFAULT nextval('public.promotions_promotion_id_seq1'::regclass);


--
-- Name: school_configs school_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.school_configs ALTER COLUMN school_id SET DEFAULT nextval('public.school_configs_school_id_seq'::regclass);


--
-- Name: schools school_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schools ALTER COLUMN school_id SET DEFAULT nextval('public.schools_school_id_seq'::regclass);


--
-- Name: student_event_logs student_event_log_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_event_logs ALTER COLUMN student_event_log_id SET DEFAULT nextval('public.student_event_logs_student_event_log_id_seq'::regclass);


--
-- Name: student_learning_time_by_daily learning_time_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_learning_time_by_daily ALTER COLUMN learning_time_id SET DEFAULT nextval('public.student_learning_time_by_daily_learning_time_id_seq'::regclass);


--
-- Name: student_orders student_order_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_orders ALTER COLUMN student_order_id SET DEFAULT nextval('public.student_orders_student_order_id_seq'::regclass);


--
-- Name: activity_logs activity_logs_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activity_logs
    ADD CONSTRAINT activity_logs_pk PRIMARY KEY (activity_log_id);


--
-- Name: assignments assignments_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.assignments
    ADD CONSTRAINT assignments_pk PRIMARY KEY (assignment_id);


--
-- Name: billing_histories billing_histories_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_histories
    ADD CONSTRAINT billing_histories_pk PRIMARY KEY (billing_id);


--
-- Name: chapters chapters_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chapters
    ADD CONSTRAINT chapters_pk PRIMARY KEY (chapter_id);


--
-- Name: cities city_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cities
    ADD CONSTRAINT city_pk PRIMARY KEY (city_id);


--
-- Name: cities city_un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cities
    ADD CONSTRAINT city_un UNIQUE (country, name);


--
-- Name: class_members class_members_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.class_members
    ADD CONSTRAINT class_members_pk PRIMARY KEY (class_member_id);


--
-- Name: class_preset_study_plans class_preset_study_plans__class_id__preset_study_plan_id__un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.class_preset_study_plans
    ADD CONSTRAINT class_preset_study_plans__class_id__preset_study_plan_id__un UNIQUE (class_id, preset_study_plan_id);


--
-- Name: class_preset_study_plans class_preset_study_plans_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.class_preset_study_plans
    ADD CONSTRAINT class_preset_study_plans_pk PRIMARY KEY (class_preset_study_plan_id);


--
-- Name: classes classes_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.classes
    ADD CONSTRAINT classes_pk PRIMARY KEY (class_id);


--
-- Name: classes classes_un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.classes
    ADD CONSTRAINT classes_un UNIQUE (class_code);


--
-- Name: coaches coaches_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.coaches
    ADD CONSTRAINT coaches_pk PRIMARY KEY (coach_id);


--
-- Name: cod_orders cod_orders_pk1; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cod_orders
    ADD CONSTRAINT cod_orders_pk1 PRIMARY KEY (cod_order_id);


--
-- Name: configs config_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.configs
    ADD CONSTRAINT config_pk PRIMARY KEY (config_key, config_group, country);


--
-- Name: courses_classes courses_classes_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.courses_classes
    ADD CONSTRAINT courses_classes_pk PRIMARY KEY (course_id, class_id);


--
-- Name: courses courses_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.courses
    ADD CONSTRAINT courses_pk PRIMARY KEY (course_id);


--
-- Name: districts district_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.districts
    ADD CONSTRAINT district_pk PRIMARY KEY (district_id);


--
-- Name: districts district_un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.districts
    ADD CONSTRAINT district_un UNIQUE (country, city_id, name);


--
-- Name: student_event_logs event_id_un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_event_logs
    ADD CONSTRAINT event_id_un UNIQUE (event_id);


--
-- Name: student_event_logs event_log_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_event_logs
    ADD CONSTRAINT event_log_pk PRIMARY KEY (student_event_log_id);


--
-- Name: hubs hub_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.hubs
    ADD CONSTRAINT hub_pk PRIMARY KEY (hub_id);


--
-- Name: hub_tours hub_tour_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.hub_tours
    ADD CONSTRAINT hub_tour_pk PRIMARY KEY (hub_id, student_id);


--
-- Name: student_orders inapp_transaction_id_un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_orders
    ADD CONSTRAINT inapp_transaction_id_un UNIQUE (inapp_transaction_id);


--
-- Name: ios_transactions ios_transactions_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ios_transactions
    ADD CONSTRAINT ios_transactions_pk PRIMARY KEY (ios_transaction_id);


--
-- Name: learning_objectives learning_objectives_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.learning_objectives
    ADD CONSTRAINT learning_objectives_pk PRIMARY KEY (lo_id);


--
-- Name: notification_messages notification_messages_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_messages
    ADD CONSTRAINT notification_messages_pk PRIMARY KEY (notification_message_id);


--
-- Name: notification_messages notification_messages_un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_messages
    ADD CONSTRAINT notification_messages_un UNIQUE (country, key, receiver_group);


--
-- Name: notification_targets notification_target_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_targets
    ADD CONSTRAINT notification_target_pk PRIMARY KEY (target_id);


--
-- Name: notifications notifications_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pk PRIMARY KEY (notification_id);


--
-- Name: package_items package_item_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.package_items
    ADD CONSTRAINT package_item_pk PRIMARY KEY (package_item_id);


--
-- Name: packages packages_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.packages
    ADD CONSTRAINT packages_pk PRIMARY KEY (package_id);


--
-- Name: plans plans_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plans
    ADD CONSTRAINT plans_pk PRIMARY KEY (plan_id, country);


--
-- Name: preset_study_plans preset_study_plans_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.preset_study_plans
    ADD CONSTRAINT preset_study_plans_pk PRIMARY KEY (preset_study_plan_id);


--
-- Name: promotion_rules promotion_rule_id_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promotion_rules
    ADD CONSTRAINT promotion_rule_id_pk PRIMARY KEY (promotion_rule_id);


--
-- Name: promotions promotions_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT promotions_pk PRIMARY KEY (promotion_id);


--
-- Name: promotions promotions_un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT promotions_un UNIQUE (country, code);


--
-- Name: questions questions_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.questions
    ADD CONSTRAINT questions_pk PRIMARY KEY (question_id);


--
-- Name: questions_tagged_learning_objectives questions_tagged_learning_objectives_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.questions_tagged_learning_objectives
    ADD CONSTRAINT questions_tagged_learning_objectives_pk PRIMARY KEY (question_id, lo_id);


--
-- Name: quizsets quizsets_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.quizsets
    ADD CONSTRAINT quizsets_pk PRIMARY KEY (lo_id, question_id);


--
-- Name: school_admins school_admins_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.school_admins
    ADD CONSTRAINT school_admins_pk PRIMARY KEY (school_admin_id);


--
-- Name: school_configs school_configs_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.school_configs
    ADD CONSTRAINT school_configs_pk PRIMARY KEY (school_id);


--
-- Name: schools school_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schools
    ADD CONSTRAINT school_pk PRIMARY KEY (school_id);


--
-- Name: schools school_un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schools
    ADD CONSTRAINT school_un UNIQUE (country, city_id, district_id, name);


--
-- Name: student_statistics statistics_student_un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_statistics
    ADD CONSTRAINT statistics_student_un UNIQUE (student_id);


--
-- Name: student_comments student_comments_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_comments
    ADD CONSTRAINT student_comments_pk PRIMARY KEY (comment_id);


--
-- Name: student_learning_time_by_daily student_learning_time_by_daily_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_learning_time_by_daily
    ADD CONSTRAINT student_learning_time_by_daily_pk PRIMARY KEY (learning_time_id);


--
-- Name: student_learning_time_by_daily student_learning_time_by_daily_un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_learning_time_by_daily
    ADD CONSTRAINT student_learning_time_by_daily_un UNIQUE (student_id, day);


--
-- Name: student_orders student_orders_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_orders
    ADD CONSTRAINT student_orders_pk PRIMARY KEY (student_order_id);


--
-- Name: student_questions student_questions_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_questions
    ADD CONSTRAINT student_questions_pk PRIMARY KEY (student_question_id);


--
-- Name: student_submission_scores student_submission_scores_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_submission_scores
    ADD CONSTRAINT student_submission_scores_pk PRIMARY KEY (student_submission_score_id);


--
-- Name: student_submissions student_submissions_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_submissions
    ADD CONSTRAINT student_submissions_pk PRIMARY KEY (student_submission_id);


--
-- Name: student_subscriptions student_subscriptions_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_subscriptions
    ADD CONSTRAINT student_subscriptions_pk PRIMARY KEY (student_subscription_id);


--
-- Name: student_assignments students_assignments_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_assignments
    ADD CONSTRAINT students_assignments_pk PRIMARY KEY (student_id, assignment_id);


--
-- Name: students_learning_objectives_completeness students_learning_objectives_completeness_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.students_learning_objectives_completeness
    ADD CONSTRAINT students_learning_objectives_completeness_pk PRIMARY KEY (student_id, lo_id);


--
-- Name: students students_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.students
    ADD CONSTRAINT students_pk PRIMARY KEY (student_id);


--
-- Name: students_topics_overdue students_topic_overdue_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.students_topics_overdue
    ADD CONSTRAINT students_topic_overdue_pk PRIMARY KEY (topic_id, student_id);


--
-- Name: students_topics_completeness students_topics_completeness_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.students_topics_completeness
    ADD CONSTRAINT students_topics_completeness_pk UNIQUE (student_id, topic_id);


--
-- Name: teachers teachers_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teachers
    ADD CONSTRAINT teachers_pk PRIMARY KEY (teacher_id);


--
-- Name: topics topics_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topics
    ADD CONSTRAINT topics_pk PRIMARY KEY (topic_id);


--
-- Name: tutors tutors_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tutors
    ADD CONSTRAINT tutors_pk PRIMARY KEY (tutor_id);


--
-- Name: user_notifications user_notifications_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notifications
    ADD CONSTRAINT user_notifications_pk PRIMARY KEY (user_notification_id);


--
-- Name: user_notifications user_notifications_un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notifications
    ADD CONSTRAINT user_notifications_un UNIQUE (user_id, notification_id);


--
-- Name: users users_email_un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_un UNIQUE (email);


--
-- Name: users users_fb_id_un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_fb_id_un UNIQUE (facebook_id);


--
-- Name: users users_phone_un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_phone_un UNIQUE (phone_number);


--
-- Name: users users_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pk PRIMARY KEY (user_id);


--
-- Name: preset_study_plans_weekly weekly_preset_study_plans_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.preset_study_plans_weekly
    ADD CONSTRAINT weekly_preset_study_plans_pk PRIMARY KEY (preset_study_plan_weekly_id);


--
-- Name: preset_study_plans_weekly weekly_preset_study_plans_un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.preset_study_plans_weekly
    ADD CONSTRAINT weekly_preset_study_plans_un UNIQUE (preset_study_plan_id, topic_id, week);


--
-- Name: students_study_plans_weekly weekly_study_plans_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.students_study_plans_weekly
    ADD CONSTRAINT weekly_study_plans_pk PRIMARY KEY (student_id, preset_study_plan_weekly_id);


--
-- Name: activity_logs_payload; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activity_logs_payload ON public.activity_logs USING gin (payload);


--
-- Name: billing_histories_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX billing_histories_created_at_idx ON public.billing_histories USING btree (created_at DESC);


--
-- Name: class_members_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX class_members_user_id_idx ON public.class_members USING btree (user_id);


--
-- Name: event_logs_student_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX event_logs_student_id_idx ON public.student_event_logs USING btree (student_id);


--
-- Name: learning_objectives_topic_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX learning_objectives_topic_id_idx ON public.learning_objectives USING btree (topic_id);


--
-- Name: quizsets_lo_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX quizsets_lo_id_idx ON public.quizsets USING btree (lo_id);


--
-- Name: student_orders_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX student_orders_created_at_idx ON public.student_orders USING btree (created_at DESC);


--
-- Name: student_orders_created_by_email_idx_created_by_email_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX student_orders_created_by_email_idx_created_by_email_idx ON public.student_orders USING btree (created_by_email);


--
-- Name: student_orders_student_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX student_orders_student_id_idx ON public.student_orders USING btree (student_id);


--
-- Name: student_subscriptions_order_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX student_subscriptions_order_idx ON public.student_subscriptions USING btree (student_order_id);


--
-- Name: students_assigned_coaches_coach_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX students_assigned_coaches_coach_id_idx ON public.students_assigned_coaches USING btree (coach_id);


--
-- Name: students_assigned_coaches_student_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX students_assigned_coaches_student_id_idx ON public.students_assigned_coaches USING btree (student_id);


--
-- Name: students_study_plans_weekly_weekly_id_student_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX students_study_plans_weekly_weekly_id_student_id ON public.students_study_plans_weekly USING btree (preset_study_plan_weekly_id, student_id);


--
-- Name: tutors_statistics_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX tutors_statistics_idx ON public.tutors USING btree (open_questions, current_active_questions, total_resolved_questions);


--
-- Name: students_assigned_coaches assigned_coaches_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.students_assigned_coaches
    ADD CONSTRAINT assigned_coaches_fk FOREIGN KEY (student_id) REFERENCES public.students(student_id);


--
-- Name: students_assigned_coaches assigned_coaches_fk_1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.students_assigned_coaches
    ADD CONSTRAINT assigned_coaches_fk_1 FOREIGN KEY (coach_id) REFERENCES public.coaches(coach_id);


--
-- Name: assignments assignments_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.assignments
    ADD CONSTRAINT assignments_fk FOREIGN KEY (topic_id) REFERENCES public.topics(topic_id);


--
-- Name: billing_histories billing_histories_generator_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_histories
    ADD CONSTRAINT billing_histories_generator_fk FOREIGN KEY (generator_id) REFERENCES public.users(user_id);


--
-- Name: billing_histories billing_histories_students_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_histories
    ADD CONSTRAINT billing_histories_students_fk FOREIGN KEY (student_id) REFERENCES public.students(student_id);


--
-- Name: topics chapter_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topics
    ADD CONSTRAINT chapter_id_fk FOREIGN KEY (chapter_id) REFERENCES public.chapters(chapter_id);


--
-- Name: chapters chapters_school_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chapters
    ADD CONSTRAINT chapters_school_id_fk FOREIGN KEY (school_id) REFERENCES public.schools(school_id);


--
-- Name: districts city_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.districts
    ADD CONSTRAINT city_id_fk FOREIGN KEY (city_id) REFERENCES public.cities(city_id);


--
-- Name: schools city_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schools
    ADD CONSTRAINT city_id_fk FOREIGN KEY (city_id) REFERENCES public.cities(city_id);


--
-- Name: class_members class_members__class_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.class_members
    ADD CONSTRAINT class_members__class_id_fk FOREIGN KEY (class_id) REFERENCES public.classes(class_id);


--
-- Name: class_members class_members__student_subscription_id__fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.class_members
    ADD CONSTRAINT class_members__student_subscription_id__fk FOREIGN KEY (student_subscription_id) REFERENCES public.student_subscriptions(student_subscription_id);


--
-- Name: class_members class_members__user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.class_members
    ADD CONSTRAINT class_members__user_id_fk FOREIGN KEY (user_id) REFERENCES public.users(user_id);


--
-- Name: class_preset_study_plans class_preset_study_plans__preset_study_plan_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.class_preset_study_plans
    ADD CONSTRAINT class_preset_study_plans__preset_study_plan_id_fk FOREIGN KEY (preset_study_plan_id) REFERENCES public.preset_study_plans(preset_study_plan_id);


--
-- Name: classes classes__plans_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.classes
    ADD CONSTRAINT classes__plans_fk FOREIGN KEY (plan_id, country) REFERENCES public.plans(plan_id, country);


--
-- Name: classes classes__school_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.classes
    ADD CONSTRAINT classes__school_id_fk FOREIGN KEY (school_id) REFERENCES public.schools(school_id);


--
-- Name: cod_orders cod_orders__student_orders_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cod_orders
    ADD CONSTRAINT cod_orders__student_orders_fk FOREIGN KEY (student_order_id) REFERENCES public.student_orders(student_order_id);


--
-- Name: courses courses_school_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.courses
    ADD CONSTRAINT courses_school_id_fk FOREIGN KEY (school_id) REFERENCES public.schools(school_id);


--
-- Name: schools district_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schools
    ADD CONSTRAINT district_id_fk FOREIGN KEY (district_id) REFERENCES public.districts(district_id);


--
-- Name: student_event_logs event_logs_student_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_event_logs
    ADD CONSTRAINT event_logs_student_fk FOREIGN KEY (student_id) REFERENCES public.students(student_id);


--
-- Name: hub_tours hub_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.hub_tours
    ADD CONSTRAINT hub_id_fk FOREIGN KEY (hub_id) REFERENCES public.hubs(hub_id);


--
-- Name: ios_transactions ios_transactions__student_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ios_transactions
    ADD CONSTRAINT ios_transactions__student_id_fk FOREIGN KEY (student_id) REFERENCES public.students(student_id);


--
-- Name: learning_objectives learning_objectives_master_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.learning_objectives
    ADD CONSTRAINT learning_objectives_master_fk FOREIGN KEY (master_lo_id) REFERENCES public.learning_objectives(lo_id);


--
-- Name: learning_objectives learning_objectives_school_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.learning_objectives
    ADD CONSTRAINT learning_objectives_school_id_fk FOREIGN KEY (school_id) REFERENCES public.schools(school_id);


--
-- Name: learning_objectives lo_topic_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.learning_objectives
    ADD CONSTRAINT lo_topic_fk FOREIGN KEY (topic_id) REFERENCES public.topics(topic_id);


--
-- Name: notifications notifications_notification_targets_target_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_notification_targets_target_id_fk FOREIGN KEY (target) REFERENCES public.notification_targets(target_id);


--
-- Name: package_items package_item__packages_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.package_items
    ADD CONSTRAINT package_item__packages_fk FOREIGN KEY (package_id) REFERENCES public.packages(package_id);


--
-- Name: package_items package_item__plans_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.package_items
    ADD CONSTRAINT package_item__plans_fk FOREIGN KEY (plan_id, country) REFERENCES public.plans(plan_id, country);


--
-- Name: promotion_rules promotion_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promotion_rules
    ADD CONSTRAINT promotion_id_fk FOREIGN KEY (promotion_id) REFERENCES public.promotions(promotion_id);


--
-- Name: promotions promotions_owner_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT promotions_owner_id_fk FOREIGN KEY (owner_id) REFERENCES public.users(user_id);


--
-- Name: questions questions_master_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.questions
    ADD CONSTRAINT questions_master_fk FOREIGN KEY (master_question_id) REFERENCES public.questions(question_id);


--
-- Name: questions_tagged_learning_objectives questions_tagged_learning_objectives_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.questions_tagged_learning_objectives
    ADD CONSTRAINT questions_tagged_learning_objectives_fk FOREIGN KEY (question_id) REFERENCES public.questions(question_id);


ALTER TABLE ONLY public.questions_tagged_learning_objectives
    ADD CONSTRAINT questions_tagged_learning_objectives_fk_1 FOREIGN KEY (lo_id) REFERENCES public.learning_objectives(lo_id);


--
-- Name: quizsets quizsets_lo_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.quizsets
    ADD CONSTRAINT quizsets_lo_fk FOREIGN KEY (lo_id) REFERENCES public.learning_objectives(lo_id);


--
-- Name: quizsets quizsets_question_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.quizsets
    ADD CONSTRAINT quizsets_question_fk FOREIGN KEY (question_id) REFERENCES public.questions(question_id);


--
-- Name: school_admins school_admin_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.school_admins
    ADD CONSTRAINT school_admin_id_fk FOREIGN KEY (school_admin_id) REFERENCES public.users(user_id);


--
-- Name: school_configs school_configs__school_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.school_configs
    ADD CONSTRAINT school_configs__school_id_fk FOREIGN KEY (school_id) REFERENCES public.schools(school_id);


--
-- Name: students school_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.students
    ADD CONSTRAINT school_id_fk FOREIGN KEY (school_id) REFERENCES public.schools(school_id);


--
-- Name: school_admins school_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.school_admins
    ADD CONSTRAINT school_id_fk FOREIGN KEY (school_id) REFERENCES public.schools(school_id);


--
-- Name: student_statistics statistics_student_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_statistics
    ADD CONSTRAINT statistics_student_fk FOREIGN KEY (student_id) REFERENCES public.students(student_id);


--
-- Name: student_assignments student_assignment_student_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_assignments
    ADD CONSTRAINT student_assignment_student_id_fk FOREIGN KEY (student_id) REFERENCES public.students(student_id);


--
-- Name: student_assignments student_assignments_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_assignments
    ADD CONSTRAINT student_assignments_fk FOREIGN KEY (assignment_id) REFERENCES public.assignments(assignment_id);


--
-- Name: student_comments student_comments_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_comments
    ADD CONSTRAINT student_comments_fk FOREIGN KEY (student_id) REFERENCES public.students(student_id);


--
-- Name: student_comments student_comments_fk1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_comments
    ADD CONSTRAINT student_comments_fk1 FOREIGN KEY (coach_id) REFERENCES public.coaches(coach_id);


--
-- Name: hub_tours student_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.hub_tours
    ADD CONSTRAINT student_id_fk FOREIGN KEY (student_id) REFERENCES public.students(student_id);


--
-- Name: student_learning_time_by_daily student_learning_time_by_daily_student_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_learning_time_by_daily
    ADD CONSTRAINT student_learning_time_by_daily_student_id_fk FOREIGN KEY (student_id) REFERENCES public.students(student_id);


--
-- Name: student_orders student_orders__ios_transaction_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_orders
    ADD CONSTRAINT student_orders__ios_transaction_id_fk FOREIGN KEY (ios_transaction_id) REFERENCES public.ios_transactions(ios_transaction_id);


--
-- Name: student_orders student_orders_packages_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_orders
    ADD CONSTRAINT student_orders_packages_fk FOREIGN KEY (package_id) REFERENCES public.packages(package_id);


--
-- Name: student_orders student_orders_students_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_orders
    ADD CONSTRAINT student_orders_students_fk FOREIGN KEY (student_id) REFERENCES public.students(student_id);


--
-- Name: student_questions student_questions_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_questions
    ADD CONSTRAINT student_questions_fk FOREIGN KEY (student_id) REFERENCES public.students(student_id);


--
-- Name: student_questions student_questions_fk1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_questions
    ADD CONSTRAINT student_questions_fk1 FOREIGN KEY (tutor_id) REFERENCES public.tutors(tutor_id);


--
-- Name: student_questions student_questions_fk2; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_questions
    ADD CONSTRAINT student_questions_fk2 FOREIGN KEY (quiz_id) REFERENCES public.questions(question_id);


--
-- Name: student_submissions student_submissions_fk_1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_submissions
    ADD CONSTRAINT student_submissions_fk_1 FOREIGN KEY (student_id) REFERENCES public.students(student_id);


--
-- Name: student_subscriptions student_subscriptions__plans_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_subscriptions
    ADD CONSTRAINT student_subscriptions__plans_fk FOREIGN KEY (plan_id, country) REFERENCES public.plans(plan_id, country);


--
-- Name: student_subscriptions student_subscriptions__student_order_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_subscriptions
    ADD CONSTRAINT student_subscriptions__student_order_fk FOREIGN KEY (student_order_id) REFERENCES public.student_orders(student_order_id);


--
-- Name: student_subscriptions student_subscriptions__students_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_subscriptions
    ADD CONSTRAINT student_subscriptions__students_fk FOREIGN KEY (student_id) REFERENCES public.students(student_id);


--
-- Name: students_learning_objectives_completeness students_learning_objectives_completeness_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.students_learning_objectives_completeness
    ADD CONSTRAINT students_learning_objectives_completeness_fk FOREIGN KEY (student_id, preset_study_plan_weekly_id) REFERENCES public.students_study_plans_weekly(student_id, preset_study_plan_weekly_id);


--
-- Name: students_learning_objectives_completeness students_learning_objectives_completeness_lo_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.students_learning_objectives_completeness
    ADD CONSTRAINT students_learning_objectives_completeness_lo_fk FOREIGN KEY (lo_id) REFERENCES public.learning_objectives(lo_id);


--
-- Name: students_learning_objectives_completeness students_learning_objectives_completeness_students_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.students_learning_objectives_completeness
    ADD CONSTRAINT students_learning_objectives_completeness_students_fk FOREIGN KEY (student_id) REFERENCES public.students(student_id);


--
-- Name: students_topics_overdue students_topic_overdue_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.students_topics_overdue
    ADD CONSTRAINT students_topic_overdue_fk FOREIGN KEY (student_id) REFERENCES public.students(student_id);


--
-- Name: students_topics_overdue students_topic_overdue_fk1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.students_topics_overdue
    ADD CONSTRAINT students_topic_overdue_fk1 FOREIGN KEY (topic_id) REFERENCES public.topics(topic_id);


--
-- Name: student_submission_scores submission_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_submission_scores
    ADD CONSTRAINT submission_fk FOREIGN KEY (student_submission_id) REFERENCES public.student_submissions(student_submission_id);


--
-- Name: student_submission_scores submission_scores_teacher_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_submission_scores
    ADD CONSTRAINT submission_scores_teacher_fk FOREIGN KEY (teacher_id) REFERENCES public.teachers(teacher_id);


--
-- Name: student_submissions submissions_topic_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student_submissions
    ADD CONSTRAINT submissions_topic_fk FOREIGN KEY (topic_id) REFERENCES public.topics(topic_id);


--
-- Name: topics topics_school_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topics
    ADD CONSTRAINT topics_school_id_fk FOREIGN KEY (school_id) REFERENCES public.schools(school_id);


--
-- Name: user_notifications user_notifications__notification_id__fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notifications
    ADD CONSTRAINT user_notifications__notification_id__fk FOREIGN KEY (notification_id) REFERENCES public.notifications(notification_id);


--
-- Name: user_notifications user_notifications__user_id__fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notifications
    ADD CONSTRAINT user_notifications__user_id__fk FOREIGN KEY (user_id) REFERENCES public.users(user_id);


--
-- Name: preset_study_plans_weekly weekly_preset_study_plans_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.preset_study_plans_weekly
    ADD CONSTRAINT weekly_preset_study_plans_fk FOREIGN KEY (topic_id) REFERENCES public.topics(topic_id);


--
-- Name: preset_study_plans_weekly weekly_preset_study_plans_fk_1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.preset_study_plans_weekly
    ADD CONSTRAINT weekly_preset_study_plans_fk_1 FOREIGN KEY (preset_study_plan_id) REFERENCES public.preset_study_plans(preset_study_plan_id);


--
-- Name: students_study_plans_weekly weekly_study_plans_student_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.students_study_plans_weekly
    ADD CONSTRAINT weekly_study_plans_student_fk FOREIGN KEY (student_id) REFERENCES public.students(student_id);


--
-- PostgreSQL database dump complete
--


INSERT INTO public.cities ("name",country,created_at,updated_at) VALUES
('Thành phố Hồ Chí Minh','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Hà Nội','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Hải Phòng','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Đà Nẵng','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Cần Thơ','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Phú Yên','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Yên Bái','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Vĩnh Phúc','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Vĩnh Long','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Tuyên Quang','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tình Trà Vinh','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Tiền Giang','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Thừa Thiên Huế','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Thanh Hóa','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Thái Nguyên','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Thái Bình','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Tây Ninh','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Sơn La','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Sóc Trăng','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Quảng Trị','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Quảng Ninh','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Quảng Ngãi','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Quảng Nam','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Quảng Bình','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Phú Thọ','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Ninh Thuận','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Ninh Bình','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Nghệ An','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Nam Định','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Lạng Sơn','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Lào Cai','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Long An','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh An Giang','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Bà Rịa - Vũng Tàu','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Bắc Giang','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Bắc Kạn','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Bạc Liêu','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Bắc Ninh','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Bến Tre','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Bình Định','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Bình Dương','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Bình Phước','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Bình Thuận','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Cà Mau','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Cao Bằng','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Đắk Lắk','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Đắk Nông','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Điện Biên','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Đồng Nai','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Đồng Tháp','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Gia Lai','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Hà Giang','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Hà Nam','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Hà Tĩnh','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Hải Dương','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Hậu Giang','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Hòa Bình','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Hưng Yên','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Khánh Hòa','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Kiên Giang','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Kon Tum','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Lai Châu','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Tỉnh Lâm Đồng','COUNTRY_VN','2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
;

INSERT INTO public.districts ("name",country,city_id,created_at,updated_at) VALUES
('Quận 1','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận 2','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận 3','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận 4','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận 5','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận 6','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận 7','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận 8','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận 9','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận 10','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận 11','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận 12','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Bình Tân','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Bình Thạnh','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Gò Vấp','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Tân Bình','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Phú Nhuận','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Thủ Đức','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Tân Phú','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cần Giờ','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Củ Chi','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nhà Bè','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bình Chánh','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hóc Môn','COUNTRY_VN',1,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Ba Đình','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Hoàn Kiếm','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Hai Bà Trưng','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Đống Đa','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Tây Hồ','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Cầu Giấy','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Thanh Xuân','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Hoàng Mai','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Long Biên','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Bắc Từ Liêm','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Nam Từ Liêm','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thanh Trì','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Gia Lâm','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đông Anh','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Sóc Sơn','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Hà Đông','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Sơn Tây','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ba Vì','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phúc Thọ','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thạch Thất','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Quốc Oai','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Chương Mỹ','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đan Phượng','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hoài Đức','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thanh Oai','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mỹ Đức','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ứng Hòa','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thường Tín','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phú Xuyên','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mê Linh','COUNTRY_VN',2,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Đồ Sơn','COUNTRY_VN',3,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Dương Kinh','COUNTRY_VN',3,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Hải An','COUNTRY_VN',3,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Hồng Bàng','COUNTRY_VN',3,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Kiến An','COUNTRY_VN',3,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Lê Chân','COUNTRY_VN',3,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Ngô Quyền','COUNTRY_VN',3,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện An Dương','COUNTRY_VN',3,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện An Lão','COUNTRY_VN',3,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cát Hải','COUNTRY_VN',3,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Kiến Thụy','COUNTRY_VN',3,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thủy Nguyên','COUNTRY_VN',3,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tiên Lãng','COUNTRY_VN',3,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyễn Vĩnh Bảo','COUNTRY_VN',3,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Cẩm Lệ','COUNTRY_VN',4,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Hải Châu','COUNTRY_VN',4,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Liên Chiểu','COUNTRY_VN',4,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Ngũ Hành Sơn','COUNTRY_VN',4,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Sơn Trà','COUNTRY_VN',4,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Thanh Khê','COUNTRY_VN',4,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hòa Vang','COUNTRY_VN',4,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cờ Đỏ','COUNTRY_VN',5,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phong Điền','COUNTRY_VN',5,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thới Lai','COUNTRY_VN',5,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thốt Nốt','COUNTRY_VN',5,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vĩnh Thạnh','COUNTRY_VN',5,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Bình Thủy','COUNTRY_VN',5,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Cái Răng','COUNTRY_VN',5,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Ninh Kiều','COUNTRY_VN',5,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Ô Môn','COUNTRY_VN',5,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Quận Thốt Nốt','COUNTRY_VN',5,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đông Hòa','COUNTRY_VN',6,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đồng Xuân','COUNTRY_VN',6,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phú Hòa','COUNTRY_VN',6,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Sơn Hòa','COUNTRY_VN',6,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Sông Cầu','COUNTRY_VN',6,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Sông Hinh','COUNTRY_VN',6,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tây Hòa','COUNTRY_VN',6,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tuy An','COUNTRY_VN',6,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Tuy Hòa','COUNTRY_VN',6,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lục Yên','COUNTRY_VN',7,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mù Căng Chải','COUNTRY_VN',7,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Trạm Tấu','COUNTRY_VN',7,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Trấn Yên','COUNTRY_VN',7,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Văn Chấn','COUNTRY_VN',7,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Văn Yên','COUNTRY_VN',7,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Yên Bình','COUNTRY_VN',7,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Yên Bái','COUNTRY_VN',7,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Nghĩa Lộ','COUNTRY_VN',7,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bình Xuyên','COUNTRY_VN',8,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lập Thạch','COUNTRY_VN',8,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Sông Lô','COUNTRY_VN',8,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tam Dương','COUNTRY_VN',8,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tam Đảo','COUNTRY_VN',8,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vĩnh Tường','COUNTRY_VN',8,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Yên Lạc','COUNTRY_VN',8,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Vĩnh Yên','COUNTRY_VN',8,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Phúc Yên','COUNTRY_VN',8,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bình Minh','COUNTRY_VN',9,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bình Tân','COUNTRY_VN',9,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Long Hồ','COUNTRY_VN',9,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mang Thít','COUNTRY_VN',9,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tam Bình','COUNTRY_VN',9,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Trà Ôn','COUNTRY_VN',9,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vũng Liêm','COUNTRY_VN',9,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Vĩnh Long','COUNTRY_VN',9,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Bình Minh','COUNTRY_VN',9,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Chiêm Hóa','COUNTRY_VN',10,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hàm Yên','COUNTRY_VN',10,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lâm Bình','COUNTRY_VN',10,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Na Hang','COUNTRY_VN',10,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Sơn Dương','COUNTRY_VN',10,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Yên Sơn','COUNTRY_VN',10,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Tuyên Quang','COUNTRY_VN',10,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Càng Long','COUNTRY_VN',11,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cầu Kè','COUNTRY_VN',11,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cầu Ngang','COUNTRY_VN',11,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Châu Thành','COUNTRY_VN',11,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Duyên Hải','COUNTRY_VN',11,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tiểu Cần','COUNTRY_VN',11,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Trà Cú','COUNTRY_VN',11,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Trà Vinh','COUNTRY_VN',11,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cái Bè','COUNTRY_VN',12,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Cai Lậy','COUNTRY_VN',12,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Châu Thành','COUNTRY_VN',12,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Chợ Gạo','COUNTRY_VN',12,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Gò Công Đông','COUNTRY_VN',12,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Gò Công Tây','COUNTRY_VN',12,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tân Phú Đông','COUNTRY_VN',12,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tân Phước','COUNTRY_VN',12,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Mỹ Tho','COUNTRY_VN',12,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Gò Công','COUNTRY_VN',12,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện A Lưới','COUNTRY_VN',13,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hương Trà','COUNTRY_VN',13,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nam Đông','COUNTRY_VN',13,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phong Điền','COUNTRY_VN',13,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phú Lộc','COUNTRY_VN',13,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phú Vang','COUNTRY_VN',13,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Quảng Điền','COUNTRY_VN',13,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Huế','COUNTRY_VN',13,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Hương Thủy','COUNTRY_VN',13,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Hương Trà','COUNTRY_VN',13,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bá Thuớc','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cẩm Thủy','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đông Sơn','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hà Trung','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hậu Lộc','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hoằng Hóa','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lang Chánh','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mường Lát','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nga Sơn','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ngọc Lặc','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Như Thanh','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Như Xuân','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nông Cống','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Quan Hóa','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Quan Sơn','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Quảng Xương','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thạch Thành','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thiệu Hóa','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thọ Xuân','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thống Nhất','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thường Xuân','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tĩnh Gia','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Triệu Sơn','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vĩnh Lộc','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Yên Định','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Thanh Hóa','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Bỉm Sơn','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Sầm Sơn','COUNTRY_VN',14,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đại Từ','COUNTRY_VN',15,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Định Hóa','COUNTRY_VN',15,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đồng Hỷ','COUNTRY_VN',15,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phổ Yên','COUNTRY_VN',15,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phú Bình','COUNTRY_VN',15,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phú Lương','COUNTRY_VN',15,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Võ Nhai','COUNTRY_VN',15,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Sông Công','COUNTRY_VN',15,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Thái Nguyên','COUNTRY_VN',15,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Phổ Yên','COUNTRY_VN',15,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đông Hưng','COUNTRY_VN',16,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hưng Hà','COUNTRY_VN',16,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Kiến Xương','COUNTRY_VN',16,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Quỳnh Phụ','COUNTRY_VN',16,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thái Thụy','COUNTRY_VN',16,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tiền Hải','COUNTRY_VN',16,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vũ Thư','COUNTRY_VN',16,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Thái Bình','COUNTRY_VN',16,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện An Biên','COUNTRY_VN',17,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bến cầu','COUNTRY_VN',17,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Châu Thành','COUNTRY_VN',17,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Dương Minh Châu','COUNTRY_VN',17,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Gò Dầu','COUNTRY_VN',17,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hòa Thành','COUNTRY_VN',17,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tân Biên','COUNTRY_VN',17,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tân Châu','COUNTRY_VN',17,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Trảng Bàng','COUNTRY_VN',17,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Tây Ninh','COUNTRY_VN',17,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Tân Châu','COUNTRY_VN',17,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bắc Yên','COUNTRY_VN',18,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mai Sơn','COUNTRY_VN',18,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mộc Châu','COUNTRY_VN',18,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mường La','COUNTRY_VN',18,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phù Yên','COUNTRY_VN',18,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Quỳnh Nhai','COUNTRY_VN',18,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Sông Mã','COUNTRY_VN',18,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Sốp Cộp','COUNTRY_VN',18,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thuận Châu','COUNTRY_VN',18,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vân Hồ','COUNTRY_VN',18,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Yên Châu','COUNTRY_VN',18,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Sơn La','COUNTRY_VN',18,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Châu Thành','COUNTRY_VN',19,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mỹ Tú','COUNTRY_VN',19,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thạnh Trị','COUNTRY_VN',19,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Trần Đề','COUNTRY_VN',19,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Sóc Trăng','COUNTRY_VN',19,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Vĩnh Châu','COUNTRY_VN',19,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cam Lộ','COUNTRY_VN',20,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện ĐaKrông','COUNTRY_VN',20,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Gio Linh','COUNTRY_VN',20,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hải Lăng','COUNTRY_VN',20,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hướng Hóa','COUNTRY_VN',20,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Triệu Phong','COUNTRY_VN',20,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vĩnh Linh','COUNTRY_VN',20,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Đông Hà','COUNTRY_VN',20,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Quảng Trị','COUNTRY_VN',20,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ba Chẽ','COUNTRY_VN',21,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bình Liêu','COUNTRY_VN',21,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cô Tô','COUNTRY_VN',21,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đầm Hà','COUNTRY_VN',21,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện đảo Vân Đồn','COUNTRY_VN',21,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hải Hà','COUNTRY_VN',21,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hoành Bồ','COUNTRY_VN',21,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tiên Yên','COUNTRY_VN',21,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Cẩm Phả','COUNTRY_VN',21,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Hạ Long','COUNTRY_VN',21,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Móng Cái','COUNTRY_VN',21,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Uông Bí','COUNTRY_VN',21,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Đông Triều','COUNTRY_VN',21,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Quảng Yên','COUNTRY_VN',21,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ba Tơ','COUNTRY_VN',22,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bình Sơn','COUNTRY_VN',22,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện đảo Lý Sơn','COUNTRY_VN',22,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đức Phổ','COUNTRY_VN',22,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Minh Long','COUNTRY_VN',22,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mộ Đức','COUNTRY_VN',22,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nghĩa Hành','COUNTRY_VN',22,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Sơn Hà','COUNTRY_VN',22,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Sơn Tây','COUNTRY_VN',22,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tây Trà','COUNTRY_VN',22,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Trà Bồng','COUNTRY_VN',22,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tư Nghĩa','COUNTRY_VN',22,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Quảng Ngãi','COUNTRY_VN',22,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bắc Trà My','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Duy Xuyên','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đại Lộc','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Điện Bàn','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đông Giang','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hiệp Đức','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nam Giang','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nam Trà My','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nông Sơn','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Núi Thành','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phú Ninh','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phước Sơn','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Quế Sơn','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tây Giang','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thăng Bình','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tiên Phước','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Hội An','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phổ Tam Kỳ','COUNTRY_VN',23,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bố Trạch','COUNTRY_VN',24,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lệ Thủy','COUNTRY_VN',24,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Minh Hóa','COUNTRY_VN',24,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Quảng Ninh','COUNTRY_VN',24,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Quảng Trạch','COUNTRY_VN',24,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tuyên Hóa','COUNTRY_VN',24,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Đồng Hới','COUNTRY_VN',24,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Ba Đồn','COUNTRY_VN',24,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cẩm Khê','COUNTRY_VN',25,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đoan Hùng','COUNTRY_VN',25,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hạ Hòa','COUNTRY_VN',25,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lâm Thao','COUNTRY_VN',25,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phú Ninh','COUNTRY_VN',25,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tam Nông','COUNTRY_VN',25,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tân Sơn','COUNTRY_VN',25,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thanh Ba','COUNTRY_VN',25,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thanh Sơn','COUNTRY_VN',25,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thanh Thủy','COUNTRY_VN',25,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Yẽn Lập','COUNTRY_VN',25,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Việt Trì','COUNTRY_VN',25,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Phú Thọ','COUNTRY_VN',25,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bác Ái','COUNTRY_VN',26,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ninh Hải','COUNTRY_VN',26,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ninh Phước','COUNTRY_VN',26,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ninh Sơn','COUNTRY_VN',26,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thuận Bắc','COUNTRY_VN',26,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thuận Nam','COUNTRY_VN',26,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Phan Rang-Tháp Chàm','COUNTRY_VN',26,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Gia Viễn','COUNTRY_VN',27,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hoa Lư','COUNTRY_VN',27,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Kim Sơn','COUNTRY_VN',27,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nho Quan','COUNTRY_VN',27,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Yên Khánh','COUNTRY_VN',27,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Yên Mô','COUNTRY_VN',27,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Ninh Bình','COUNTRY_VN',27,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Tam Điệp','COUNTRY_VN',27,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Anh Sơn','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Con Cuông','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Diễn Châu','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đô Lương','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hưng Nguyên','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Kỳ Sơn','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nam Đàn','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nghi Lộc','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nghĩa Đàn','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Quế Phong','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Quỳ châu','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Quỳ Hợp','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Quỳnh Lưu','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tân Kỳ','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thanh chương','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tương Dương','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Yên Thành','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Vinh','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Cửa Lò','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Thái Hòa','COUNTRY_VN',28,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Giao Thủy','COUNTRY_VN',29,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hải Hậu','COUNTRY_VN',29,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mỹ Lộc','COUNTRY_VN',29,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nam Trục','COUNTRY_VN',29,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nghĩa Hưng','COUNTRY_VN',29,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Trục Ninh','COUNTRY_VN',29,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vụ Bản','COUNTRY_VN',29,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Xuân Trường','COUNTRY_VN',29,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ý Yên','COUNTRY_VN',29,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Nam Định','COUNTRY_VN',29,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bắc Sơn','COUNTRY_VN',30,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bình Gia','COUNTRY_VN',30,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cao Lộc','COUNTRY_VN',30,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Chi Lăng','COUNTRY_VN',30,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đình Lập','COUNTRY_VN',30,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hữu Lũng','COUNTRY_VN',30,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lộc Bình','COUNTRY_VN',30,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tràng Định','COUNTRY_VN',30,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Văn Lãng','COUNTRY_VN',30,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Văn Quan','COUNTRY_VN',30,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Lạng Sơn','COUNTRY_VN',30,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bắc Hà','COUNTRY_VN',31,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bảo Thắng','COUNTRY_VN',31,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bảo Yên','COUNTRY_VN',31,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bát Xát','COUNTRY_VN',31,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mường Khương','COUNTRY_VN',31,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Sa Pa','COUNTRY_VN',31,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Si Ma Cai','COUNTRY_VN',31,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Văn Bàn','COUNTRY_VN',31,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành Phố Lào Cai','COUNTRY_VN',31,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bến Lức','COUNTRY_VN',32,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cần Đước','COUNTRY_VN',32,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cần Giuộc','COUNTRY_VN',32,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Châu Thành','COUNTRY_VN',32,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đức Hòa','COUNTRY_VN',32,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đức Huệ','COUNTRY_VN',32,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Kiến Tường','COUNTRY_VN',32,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mộc Hóa','COUNTRY_VN',32,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tân Hưng','COUNTRY_VN',32,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tân Thành','COUNTRY_VN',32,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tân Trụ','COUNTRY_VN',32,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thạnh Hóa','COUNTRY_VN',32,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thủ Thừa','COUNTRY_VN',32,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vĩnh Hưng','COUNTRY_VN',32,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Tân An','COUNTRY_VN',32,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Thạnh Hóa','COUNTRY_VN',32,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện An Phú','COUNTRY_VN',33,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Châu Đốc','COUNTRY_VN',33,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Châu Phú','COUNTRY_VN',33,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Châu Thành','COUNTRY_VN',33,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Chợ Mới','COUNTRY_VN',33,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Long Xuyên','COUNTRY_VN',33,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phú Tân','COUNTRY_VN',33,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Tân Châu','COUNTRY_VN',33,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thoại Sơn','COUNTRY_VN',33,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tịnh Biên','COUNTRY_VN',33,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tri Tôn','COUNTRY_VN',33,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Châu Đốc','COUNTRY_VN',33,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Long Xuyên','COUNTRY_VN',33,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Châu Đức','COUNTRY_VN',34,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Côn Đảo','COUNTRY_VN',34,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đất Đỏ','COUNTRY_VN',34,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Long Điền','COUNTRY_VN',34,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tân Thành','COUNTRY_VN',34,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Xuyên Mộc','COUNTRY_VN',34,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Bà Rịa','COUNTRY_VN',34,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Vũng Tàu','COUNTRY_VN',34,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hiệp Hòa','COUNTRY_VN',35,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lạng Giang','COUNTRY_VN',35,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lục Nam','COUNTRY_VN',35,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lục Ngạn','COUNTRY_VN',35,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Sơn Động','COUNTRY_VN',35,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tân Yên','COUNTRY_VN',35,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Việt Yên','COUNTRY_VN',35,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Yên Dũng','COUNTRY_VN',35,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyên Yên Thế','COUNTRY_VN',35,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Bắc Giang','COUNTRY_VN',35,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ba Bể','COUNTRY_VN',36,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Bắc Kạn','COUNTRY_VN',36,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bạch Thông','COUNTRY_VN',36,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Chợ Đồn','COUNTRY_VN',36,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Chợ Mới','COUNTRY_VN',36,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Na Rì','COUNTRY_VN',36,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ngân Sơn','COUNTRY_VN',36,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Pác Nặm','COUNTRY_VN',36,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đông Hải','COUNTRY_VN',37,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Giá Rai','COUNTRY_VN',37,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hòa Bình','COUNTRY_VN',37,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hồng Dân','COUNTRY_VN',37,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Phước Long','COUNTRY_VN',37,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vĩnh Lợi','COUNTRY_VN',37,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Bạc Liêu','COUNTRY_VN',37,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Giá Rai','COUNTRY_VN',37,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Gia Bình','COUNTRY_VN',38,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('huyện Lương Tài','COUNTRY_VN',38,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('huyện Quế Võ','COUNTRY_VN',38,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('huyện Thuận Thành','COUNTRY_VN',38,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('huyện Tiên Du','COUNTRY_VN',38,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('huyện Yên Phong','COUNTRY_VN',38,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('thành phố Bắc Ninh','COUNTRY_VN',38,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('thị xã Từ Sơn','COUNTRY_VN',38,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ba Tri','COUNTRY_VN',39,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bình Đại','COUNTRY_VN',39,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Châu Thành','COUNTRY_VN',39,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Chợ Lách','COUNTRY_VN',39,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Giồng Trôm','COUNTRY_VN',39,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mỏ Cày Bắc','COUNTRY_VN',39,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mỏ Cày Nam','COUNTRY_VN',39,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thạnh Phú','COUNTRY_VN',39,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Bến Tre','COUNTRY_VN',39,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện An Lão','COUNTRY_VN',40,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hoài Ân','COUNTRY_VN',40,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hoài Nhơn','COUNTRY_VN',40,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phù Cát','COUNTRY_VN',40,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phù Mỹ','COUNTRY_VN',40,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tây Sơn','COUNTRY_VN',40,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tây Phước','COUNTRY_VN',40,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vân Canh','COUNTRY_VN',40,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vĩnh Thạnh','COUNTRY_VN',40,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Quy Nhơn','COUNTRY_VN',40,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã An Nhơn','COUNTRY_VN',40,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Hoài Nhơn','COUNTRY_VN',40,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bắc Tân Uyên','COUNTRY_VN',41,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bàu Bàng','COUNTRY_VN',41,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Dầu Tiếng','COUNTRY_VN',41,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Dĩ An','COUNTRY_VN',41,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phú Giáo','COUNTRY_VN',41,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Thủ Dầu Một','COUNTRY_VN',41,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Bến Cát','COUNTRY_VN',41,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Dĩ An','COUNTRY_VN',41,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Tân Uyên','COUNTRY_VN',41,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Thuận An','COUNTRY_VN',41,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện An Phú','COUNTRY_VN',42,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bù Đăng','COUNTRY_VN',42,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bù Đốp','COUNTRY_VN',42,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bù Gia Mập','COUNTRY_VN',42,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Chơn Thành','COUNTRY_VN',42,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đồng Phú','COUNTRY_VN',42,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hớn Quản','COUNTRY_VN',42,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lộc Ninh','COUNTRY_VN',42,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Bình Long','COUNTRY_VN',42,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Đồng Xoài','COUNTRY_VN',42,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Phước Long','COUNTRY_VN',42,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bắc Bình','COUNTRY_VN',43,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện đảo Phú Quý','COUNTRY_VN',43,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đức Linh','COUNTRY_VN',43,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hàm Tâm','COUNTRY_VN',43,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hàm Thuận Bắc','COUNTRY_VN',43,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hàm Thuận Nam','COUNTRY_VN',43,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lagi','COUNTRY_VN',43,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tánh Linh','COUNTRY_VN',43,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tuy Phong','COUNTRY_VN',43,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Phan Thiết','COUNTRY_VN',43,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cái Nước','COUNTRY_VN',44,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đầm Dơi','COUNTRY_VN',44,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Năm Căn','COUNTRY_VN',44,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ngọc Hiển','COUNTRY_VN',44,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phú Tân','COUNTRY_VN',44,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thới Bình','COUNTRY_VN',44,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Trần Văn Thời','COUNTRY_VN',44,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện U Minh','COUNTRY_VN',44,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Cà Mau','COUNTRY_VN',44,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bảo Lạc','COUNTRY_VN',45,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bảo Lâm','COUNTRY_VN',45,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hạ Lang','COUNTRY_VN',45,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hà Quảng','COUNTRY_VN',45,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hòa An','COUNTRY_VN',45,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nguyên Bình','COUNTRY_VN',45,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phục Hòa','COUNTRY_VN',45,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Quảng Uyên','COUNTRY_VN',45,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thạch An','COUNTRY_VN',45,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thông Nông','COUNTRY_VN',45,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Trà Lĩnh','COUNTRY_VN',45,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Trùng Khánh','COUNTRY_VN',45,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Cao Bằng','COUNTRY_VN',45,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Buôn Đôn','COUNTRY_VN',46,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cư Kuin','COUNTRY_VN',46,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cư M''gar','COUNTRY_VN',46,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ea H''leo','COUNTRY_VN',46,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ea Kar','COUNTRY_VN',46,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ea Súp','COUNTRY_VN',46,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Krông Ana','COUNTRY_VN',46,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Krông Bông','COUNTRY_VN',46,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Krông Búk','COUNTRY_VN',46,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Krông Năng','COUNTRY_VN',46,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Krông Pắc','COUNTRY_VN',46,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lắk','COUNTRY_VN',46,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện M''Drắc','COUNTRY_VN',46,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Buôn Mê Thuột','COUNTRY_VN',46,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Buôn Hồ','COUNTRY_VN',46,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cư Jút','COUNTRY_VN',47,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đắk Giong','COUNTRY_VN',47,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đăk Mil','COUNTRY_VN',47,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đắk R''lấp','COUNTRY_VN',47,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đắk Song','COUNTRY_VN',47,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Krông Nô','COUNTRY_VN',47,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tuy Đức','COUNTRY_VN',47,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Gia Nghĩa','COUNTRY_VN',47,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Điện Biên','COUNTRY_VN',48,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Điện Biên Đông','COUNTRY_VN',48,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mường Ảng','COUNTRY_VN',48,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mường Chà','COUNTRY_VN',48,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mường Nhé','COUNTRY_VN',48,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nậm Pồ','COUNTRY_VN',48,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tủa Chùa','COUNTRY_VN',48,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tuần Giáo','COUNTRY_VN',48,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Điện Biên Phủ','COUNTRY_VN',48,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Mường Lay','COUNTRY_VN',48,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện An Phú','COUNTRY_VN',49,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cẩm Mỹ','COUNTRY_VN',49,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Định Quán','COUNTRY_VN',49,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Long Thành','COUNTRY_VN',49,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nhơn Trạch','COUNTRY_VN',49,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tân Phú','COUNTRY_VN',49,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thống Nhất','COUNTRY_VN',49,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Trảng Bom','COUNTRY_VN',49,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vĩnh Cửu','COUNTRY_VN',49,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Xuân Lộc','COUNTRY_VN',49,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Biên Hòa','COUNTRY_VN',49,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('thành phố Cao Lãnh','COUNTRY_VN',50,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Châu Thành','COUNTRY_VN',50,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Hồng Ngự','COUNTRY_VN',50,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Lai Vung','COUNTRY_VN',50,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lấp Vò','COUNTRY_VN',50,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tam Nông','COUNTRY_VN',50,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tân Hồng','COUNTRY_VN',50,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thanh Bình','COUNTRY_VN',50,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tháp Mười','COUNTRY_VN',50,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Sa Đéc','COUNTRY_VN',50,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Chư Păh','COUNTRY_VN',51,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Chư Prông','COUNTRY_VN',51,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Chư Pưh','COUNTRY_VN',51,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Chư Sê','COUNTRY_VN',51,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đăk Đoa','COUNTRY_VN',51,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đak Pơ','COUNTRY_VN',51,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đức Cơ','COUNTRY_VN',51,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện la Grai','COUNTRY_VN',51,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện La Pa','COUNTRY_VN',51,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện K''Bang','COUNTRY_VN',51,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Kông Chro','COUNTRY_VN',51,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Krông Pa','COUNTRY_VN',51,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mang Yang','COUNTRY_VN',51,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phú Thiện','COUNTRY_VN',51,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Pleiku','COUNTRY_VN',51,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã An Khê','COUNTRY_VN',51,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Ayun Pa','COUNTRY_VN',51,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bắc Mê','COUNTRY_VN',52,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bắc Quang','COUNTRY_VN',52,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đồng Văn','COUNTRY_VN',52,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hoàng Su Phì','COUNTRY_VN',52,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mèo Vạc','COUNTRY_VN',52,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Quản Bạ','COUNTRY_VN',52,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Quang Bình','COUNTRY_VN',52,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vị Xuyên','COUNTRY_VN',52,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Xín Mần','COUNTRY_VN',52,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Yên Mân','COUNTRY_VN',52,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Hà Giang','COUNTRY_VN',52,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bình Lục','COUNTRY_VN',53,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Duy Tiên','COUNTRY_VN',53,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Kim Bảng','COUNTRY_VN',53,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lý Nhân','COUNTRY_VN',53,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thanh Liêm','COUNTRY_VN',53,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Phủ lý','COUNTRY_VN',53,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cẩm Xuyên','COUNTRY_VN',54,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Can Lộc','COUNTRY_VN',54,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đức Thọ','COUNTRY_VN',54,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hương Khê','COUNTRY_VN',54,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Hương Sơn','COUNTRY_VN',54,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Kỳ Ảnh','COUNTRY_VN',54,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lộc Hà','COUNTRY_VN',54,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nghi Xuân','COUNTRY_VN',54,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thạch Hà','COUNTRY_VN',54,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vũ Quang','COUNTRY_VN',54,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Hà Tĩnh','COUNTRY_VN',54,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bình Giang','COUNTRY_VN',55,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cẩm Giàng','COUNTRY_VN',55,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Chí Linh','COUNTRY_VN',55,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Gia Lộc','COUNTRY_VN',55,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Kim Thành','COUNTRY_VN',55,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Kinh Môn','COUNTRY_VN',55,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nam Sách','COUNTRY_VN',55,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ninh Giang','COUNTRY_VN',55,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thanh Hà','COUNTRY_VN',55,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Thanh Miện','COUNTRY_VN',55,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tứ Kỳ','COUNTRY_VN',55,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Hải Dương','COUNTRY_VN',55,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Châu Thành','COUNTRY_VN',56,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Châu Thành A','COUNTRY_VN',56,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Long Mỹ','COUNTRY_VN',56,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phụng Hiệp','COUNTRY_VN',56,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vị Thủy','COUNTRY_VN',56,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Mỹ Thanh','COUNTRY_VN',56,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Ngã Bảy','COUNTRY_VN',56,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cao Phong','COUNTRY_VN',57,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đà Bắc','COUNTRY_VN',57,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Hòa Bình','COUNTRY_VN',57,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Kim Bôi','COUNTRY_VN',57,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Kỳ Sơn','COUNTRY_VN',57,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lạc Sơn','COUNTRY_VN',57,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lạc Thủy','COUNTRY_VN',57,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lương Sơn','COUNTRY_VN',57,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mai Châu','COUNTRY_VN',57,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tân Lạc','COUNTRY_VN',57,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Yên Thủy','COUNTRY_VN',57,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ân Thi','COUNTRY_VN',58,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đồng Hỷ','COUNTRY_VN',58,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Khoái Châu','COUNTRY_VN',58,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Kim Động','COUNTRY_VN',58,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mỹ Hào','COUNTRY_VN',58,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phù Cừ','COUNTRY_VN',58,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tiên Lữ','COUNTRY_VN',58,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Văn Giang','COUNTRY_VN',58,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Văn Lâm','COUNTRY_VN',58,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Yên Mỹ','COUNTRY_VN',58,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Hưng Yên','COUNTRY_VN',58,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cam Lâm','COUNTRY_VN',59,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Diên Khánh','COUNTRY_VN',59,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Khánh Sơn','COUNTRY_VN',59,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Khánh Vĩnh','COUNTRY_VN',59,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vạn Ninh','COUNTRY_VN',59,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Cam Ranh','COUNTRY_VN',59,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Nha Trang','COUNTRY_VN',59,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện An Biên','COUNTRY_VN',60,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện An Minh','COUNTRY_VN',60,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Châu Thành','COUNTRY_VN',60,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện đảo Kiên Hải','COUNTRY_VN',60,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện đảo Phú Quốc','COUNTRY_VN',60,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Giang Thành','COUNTRY_VN',60,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Giồng Riềng','COUNTRY_VN',60,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Gò Quao','COUNTRY_VN',60,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Gò Đất','COUNTRY_VN',60,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Kiên Hải','COUNTRY_VN',60,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Kiên Lương','COUNTRY_VN',60,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phú Quốc','COUNTRY_VN',60,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Rạch Giá','COUNTRY_VN',60,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tân Hiệp','COUNTRY_VN',60,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện U Minh Thượng','COUNTRY_VN',60,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Vĩnh Thuận','COUNTRY_VN',60,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Hà Tiên','COUNTRY_VN',60,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đắk Glei','COUNTRY_VN',61,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đắk Hà','COUNTRY_VN',61,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đăk Tô','COUNTRY_VN',61,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Kon Plông','COUNTRY_VN',61,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Kon Rẫy','COUNTRY_VN',61,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Ngọc Hồi','COUNTRY_VN',61,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Sa Thầy','COUNTRY_VN',61,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tu Mơ Rông','COUNTRY_VN',61,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Kon Tum','COUNTRY_VN',61,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Mường Tè','COUNTRY_VN',62,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Nậm Nhùn','COUNTRY_VN',62,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Phong Thổ','COUNTRY_VN',62,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Sình Hồ','COUNTRY_VN',62,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Tam Đường','COUNTRY_VN',62,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thị xã Tân Uyên','COUNTRY_VN',62,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Than Uyên','COUNTRY_VN',62,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Lai Châu','COUNTRY_VN',62,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Bảo Lâm','COUNTRY_VN',63,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Cát Tiên','COUNTRY_VN',63,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Di Linh','COUNTRY_VN',63,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đạ Huoai','COUNTRY_VN',63,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đạ Tẻh','COUNTRY_VN',63,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đam Rông','COUNTRY_VN',63,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đơn Dương','COUNTRY_VN',63,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Đức Trọng','COUNTRY_VN',63,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lạc Dương','COUNTRY_VN',63,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Huyện Lâm Hà','COUNTRY_VN',63,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Bảo Lộc','COUNTRY_VN',63,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
,('Thành phố Đà Lạt','COUNTRY_VN',63,'2019-09-27 14:34:56.970','2019-09-27 14:34:56.970')
;

INSERT INTO public.schools ("name",country,city_id,district_id,point,is_system_school,created_at,updated_at) VALUES
('TH-THCS-THPT Song ngữ Quốc tế Horizon','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:56.992','2019-09-27 14:34:56.992')
,('TH-THCS-THPT Úc châu','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:56.998','2019-09-27 14:34:56.998')
,('TH-THCS-THPT Quốc tế Á châu','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:57.000','2019-09-27 14:34:57.000')
,('THPT Bùi Thị Xuân','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:57.002','2019-09-27 14:34:57.002')
,('THPT Châu Á Thái Bình Dương','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:57.003','2019-09-27 14:34:57.003')
,('THPT Lương Thế Vinh','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:57.004','2019-09-27 14:34:57.004')
,('THPT Tenlơman','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:57.005','2019-09-27 14:34:57.005')
,('THPT Trần Đại Nghĩa','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:57.006','2019-09-27 14:34:57.006')
,('THPT Trưng Vương','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:57.007','2019-09-27 14:34:57.007')
,('TTGDTX Lê Quý Đôn','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:57.008','2019-09-27 14:34:57.008')
,('TTGDTX Quận 1','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:57.009','2019-09-27 14:34:57.009')
,('THPT Giồng Ông Tố','COUNTRY_VN',1,2,NULL,true,'2019-09-27 14:34:57.012','2019-09-27 14:34:57.012')
,('THPT Thủ Thiêm','COUNTRY_VN',1,2,NULL,true,'2019-09-27 14:34:57.014','2019-09-27 14:34:57.014')
,('TTGDTX Quận 2','COUNTRY_VN',1,2,NULL,true,'2019-09-27 14:34:57.016','2019-09-27 14:34:57.016')
,('THPT Lê Quý Đôn','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:57.018','2019-09-27 14:34:57.018')
,('THPT Marie-Curie','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:57.019','2019-09-27 14:34:57.019')
,('THPT Nguyễn Thị Diệu','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:57.019','2019-09-27 14:34:57.019')
,('THPT Nguyễn Thị Minh Khai','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:57.021','2019-09-27 14:34:57.021')
,('THPT Lê Thị Hồng Gấm','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:57.022','2019-09-27 14:34:57.022')
,('TTGDTX Quận 3','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:57.023','2019-09-27 14:34:57.023')
,('THPT Nguyễn Hữu Thọ','COUNTRY_VN',1,4,NULL,true,'2019-09-27 14:34:57.026','2019-09-27 14:34:57.026')
,('THPT Nguyễn Trãi','COUNTRY_VN',1,4,NULL,true,'2019-09-27 14:34:57.027','2019-09-27 14:34:57.027')
,('TTGDTX Quận 4','COUNTRY_VN',1,4,NULL,true,'2019-09-27 14:34:57.029','2019-09-27 14:34:57.029')
,('THPT Trần Hữu Trang','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:57.032','2019-09-27 14:34:57.032')
,('THPT Trần Khai Nguyên','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:57.033','2019-09-27 14:34:57.033')
,('THPT Văn Lang','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:57.034','2019-09-27 14:34:57.034')
,('THPT Thực hành/ĐHSP','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:57.034','2019-09-27 14:34:57.034')
,('THTH Sài Gòn','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:57.035','2019-09-27 14:34:57.035')
,('TTGDTX Chu Văn An','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:57.036','2019-09-27 14:34:57.036')
,('TTGDTX Quận 5','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:57.036','2019-09-27 14:34:57.036')
,('THCS-THPT Khai Trí','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:57.037','2019-09-27 14:34:57.037')
,('THPT Hùng Vương','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:57.038','2019-09-27 14:34:57.038')
,('THPT Lê Hồng Phong','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:57.038','2019-09-27 14:34:57.038')
,('THPT Tân Nam Mỹ','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:57.039','2019-09-27 14:34:57.039')
,('THPT Thăng Long','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:57.039','2019-09-27 14:34:57.039')
,('Phổ thông Năng khiếu ĐHQG-HCM','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:57.040','2019-09-27 14:34:57.040')
,('THPT Mạc Đĩnh Chi','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:57.043','2019-09-27 14:34:57.043')
,('THPT Nguyễn Tất Thành','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:57.043','2019-09-27 14:34:57.043')
,('THPT Phạm Phú Thứ','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:57.044','2019-09-27 14:34:57.044')
,('THPT Phan Bội Châu','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:57.046','2019-09-27 14:34:57.046')
,('THPT Phú Lâm','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:57.048','2019-09-27 14:34:57.048')
,('THPT Quốc Trí','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:57.049','2019-09-27 14:34:57.049')
,('TTGDTX Quận 6','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:57.050','2019-09-27 14:34:57.050')
,('THPT Bình Phú','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:57.051','2019-09-27 14:34:57.051')
,('THPT Lê Thánh Tôn','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:57.053','2019-09-27 14:34:57.053')
,('THPT Nam Sài gòn','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:57.053','2019-09-27 14:34:57.053')
,('THPT Ngô Quyền','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:57.054','2019-09-27 14:34:57.054')
,('THPT Quốc tế Khai Sáng','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:57.054','2019-09-27 14:34:57.054')
,('THPT Tân Phong','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:57.054','2019-09-27 14:34:57.054')
,('TTGDTX Quận 7','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:57.055','2019-09-27 14:34:57.055')
,('TH-THCS-THPT Quốc tế Canada','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:57.055','2019-09-27 14:34:57.055')
,('THPT Lương Văn Can','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:57.057','2019-09-27 14:34:57.057')
,('THPT Ngô Gia Tự','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:57.059','2019-09-27 14:34:57.059')
,('THPT Nguyễn Văn Linh','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:57.060','2019-09-27 14:34:57.060')
,('THPT NKTDTT Nguyễn Thị Định','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:57.061','2019-09-27 14:34:57.061')
,('THPT Tạ Quang Bủu','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:57.064','2019-09-27 14:34:57.064')
,('TH-THCS-THPT Nam Mỹ','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:57.066','2019-09-27 14:34:57.066')
,('TTGDTX Quận 8','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:57.066','2019-09-27 14:34:57.066')
,('THPT Hoa Sen','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:57.067','2019-09-27 14:34:57.067')
,('THPT Long Trường','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:57.068','2019-09-27 14:34:57.068')
,('THPT Nguyễn Huệ','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:57.068','2019-09-27 14:34:57.068')
,('THPT Nguyễn Văn Tăng','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:57.068','2019-09-27 14:34:57.068')
,('THPT Phước Long','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:57.069','2019-09-27 14:34:57.069')
,('TTGDTX Quận 9','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:57.069','2019-09-27 14:34:57.069')
,('THPT Nguyễn An Ninh','COUNTRY_VN',1,10,NULL,true,'2019-09-27 14:34:57.071','2019-09-27 14:34:57.071')
,('THPT Nguyễn Du','COUNTRY_VN',1,10,NULL,true,'2019-09-27 14:34:57.071','2019-09-27 14:34:57.071')
,('THPT Nguyễn Khuyến','COUNTRY_VN',1,10,NULL,true,'2019-09-27 14:34:57.071','2019-09-27 14:34:57.071')
,('THPT Việt Úc','COUNTRY_VN',1,10,NULL,true,'2019-09-27 14:34:57.072','2019-09-27 14:34:57.072')
,('TTGDTX Quận 10','COUNTRY_VN',1,10,NULL,true,'2019-09-27 14:34:57.072','2019-09-27 14:34:57.072')
,('THPT Quốc tế APU','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:57.074','2019-09-27 14:34:57.074')
,('THPT Trần Nhân Tông','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:57.076','2019-09-27 14:34:57.076')
,('THPT Trần Quang Khải','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:57.077','2019-09-27 14:34:57.077')
,('THPT Trần Quốc Tuấn','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:57.079','2019-09-27 14:34:57.079')
,('THPT Việt Mỹ Anh','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:57.081','2019-09-27 14:34:57.081')
,('TTGDTX Quận 11','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:57.082','2019-09-27 14:34:57.082')
,('THPT Nam Kỳ Khởi Nghĩa','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:57.083','2019-09-27 14:34:57.083')
,('THPT Nguyễn Hiền','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:57.084','2019-09-27 14:34:57.084')
,('THPT Thạnh Lộc','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:57.086','2019-09-27 14:34:57.086')
,('THPT Trường Chinh','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:57.087','2019-09-27 14:34:57.087')
,('THPT Võ Trường Toàn','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:57.087','2019-09-27 14:34:57.087')
,('TTGDTX Quận 12','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:57.088','2019-09-27 14:34:57.088')
,('TH-THCS-THPT Mỹ việt','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:57.088','2019-09-27 14:34:57.088')
,('THPT Chu Văn An','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:57.089','2019-09-27 14:34:57.089')
,('THPT Hàm Nghi','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:57.090','2019-09-27 14:34:57.090')
,('THPT Ngôi Sao','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:57.091','2019-09-27 14:34:57.091')
,('THPT Nguyễn Hữu cảnh','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:57.093','2019-09-27 14:34:57.093')
,('THPT Phan Châu Trinh','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:57.094','2019-09-27 14:34:57.094')
,('THPT Vĩnh Lộc','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:57.094','2019-09-27 14:34:57.094')
,('TTGDTX Quận Bình Tân','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:57.096','2019-09-27 14:34:57.096')
,('THPT An Lạc','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:57.098','2019-09-27 14:34:57.098')
,('THPT Bình Hưng Hòa','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:57.099','2019-09-27 14:34:57.099')
,('THPT Bình Tân','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:57.099','2019-09-27 14:34:57.099')
,('THPT Lam Sơn','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:57.101','2019-09-27 14:34:57.101')
,('THPT Phan Đăng Lưu','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:57.101','2019-09-27 14:34:57.101')
,('THPT Thanh Đa','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:57.102','2019-09-27 14:34:57.102')
,('THPT Trần Văn Giàu','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:57.102','2019-09-27 14:34:57.102')
,('THPT Võ Thị Sáu','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:57.103','2019-09-27 14:34:57.103')
,('TTGDTX Gia Định','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:57.103','2019-09-27 14:34:57.103')
,('TTGDTX Quận Bình Thạnh','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:57.104','2019-09-27 14:34:57.104')
,('THPT Đông Đô','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:57.105','2019-09-27 14:34:57.105')
,('THPT Gia Định','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:57.105','2019-09-27 14:34:57.105')
,('THPT Hoàng Hoa Thám','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:57.106','2019-09-27 14:34:57.106')
,('THPT Hưng Đạo','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:57.106','2019-09-27 14:34:57.106')
,('THPT Nguyễn Tri Phương','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:57.109','2019-09-27 14:34:57.109')
,('THPT Nguyễn Trung Trực','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:57.110','2019-09-27 14:34:57.110')
,('THPT Phùng Hưng','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:57.111','2019-09-27 14:34:57.111')
,('THPT Trần Hưng Đạo','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:57.113','2019-09-27 14:34:57.113')
,('THPT Việt Âu','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:57.115','2019-09-27 14:34:57.115')
,('TTGDTX Quận Gò vấp','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:57.116','2019-09-27 14:34:57.116')
,('THCS-THPT Hồng Hà','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:57.116','2019-09-27 14:34:57.116')
,('THPT Đào Duy Từ','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:57.117','2019-09-27 14:34:57.117')
,('THPT Đông Dương','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:57.118','2019-09-27 14:34:57.118')
,('THPT Gò Vấp','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:57.119','2019-09-27 14:34:57.119')
,('THPT Hermann Gmeiner','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:57.119','2019-09-27 14:34:57.119')
,('THPT Lý Thái Tổ','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:57.120','2019-09-27 14:34:57.120')
,('THPT Nguyễn Công Trứ','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:57.120','2019-09-27 14:34:57.120')
,('TH-THCS-THPT Đại việt','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:57.121','2019-09-27 14:34:57.121')
,('THCS-THPT Âu Lạc','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:57.121','2019-09-27 14:34:57.121')
,('THCS-THPT Phạm Ngũ Lão','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:57.121','2019-09-27 14:34:57.121')
,('THCS-THPT Phan Huy Ích','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:57.122','2019-09-27 14:34:57.122')
,('THPT Nguyễn Thái Bình','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:57.122','2019-09-27 14:34:57.122')
,('THPT Nguyễn Thượng Hiền','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:57.123','2019-09-27 14:34:57.123')
,('THPT Tân Trào','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:57.123','2019-09-27 14:34:57.123')
,('THPT Thanh Bình','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:57.124','2019-09-27 14:34:57.124')
,('THPT Thủ Khoa Huân','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:57.127','2019-09-27 14:34:57.127')
,('TTGDTX Quận Tân Bình','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:57.130','2019-09-27 14:34:57.130')
,('TTGDTX TN xung phong','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:57.133','2019-09-27 14:34:57.133')
,('THCS-THPT Hiền Vương','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:57.134','2019-09-27 14:34:57.134')
,('THPT Hai Bà Trung','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:57.136','2019-09-27 14:34:57.136')
,('THPT Lý Tự Trọng','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:57.137','2019-09-27 14:34:57.137')
,('THPT Nguyễn Chí Thanh','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:57.137','2019-09-27 14:34:57.137')
,('THCS-THPT Hoàng Diệu','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:57.139','2019-09-27 14:34:57.139')
,('THPT Hàn Thuyên','COUNTRY_VN',1,17,NULL,true,'2019-09-27 14:34:57.146','2019-09-27 14:34:57.146')
,('THPT Phú Nhuận','COUNTRY_VN',1,17,NULL,true,'2019-09-27 14:34:57.148','2019-09-27 14:34:57.148')
,('THPT Quốc tế Việt úc','COUNTRY_VN',1,17,NULL,true,'2019-09-27 14:34:57.149','2019-09-27 14:34:57.149')
,('TH-THCS-THPT Quốc tế','COUNTRY_VN',1,17,NULL,true,'2019-09-27 14:34:57.150','2019-09-27 14:34:57.150')
,('TTGDTX Quận Phú Nhuận','COUNTRY_VN',1,17,NULL,true,'2019-09-27 14:34:57.150','2019-09-27 14:34:57.150')
,('THPT Đào Sơn Tây','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:57.151','2019-09-27 14:34:57.151')
,('THPT Hiệp Bình','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:57.152','2019-09-27 14:34:57.152')
,('THPT Nguyễn Hữu Huân','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:57.152','2019-09-27 14:34:57.152')
,('THPT Phương Nam','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:57.153','2019-09-27 14:34:57.153')
,('THPT Tam Phú','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:57.153','2019-09-27 14:34:57.153')
,('THPT Thủ Đúc','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:57.154','2019-09-27 14:34:57.154')
,('TTGDTX Quận Thủ Đức','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:57.154','2019-09-27 14:34:57.154')
,('THPT Bách Việt','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:57.155','2019-09-27 14:34:57.155')
,('THPT Tân Bình','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:57.156','2019-09-27 14:34:57.156')
,('THPT Tây Thạnh','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:57.156','2019-09-27 14:34:57.156')
,('THPT Trần Cao Vân','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:57.157','2019-09-27 14:34:57.157')
,('THPT Trần Phú','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:57.158','2019-09-27 14:34:57.158')
,('THPT Trần Quốc Toàn','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:57.160','2019-09-27 14:34:57.160')
,('THPT Vĩnh Viễn','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:57.160','2019-09-27 14:34:57.160')
,('TTGDTX Quận Tân Phú','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:57.161','2019-09-27 14:34:57.161')
,('THCS-THPT Nhân văn','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:57.163','2019-09-27 14:34:57.163')
,('THPT Đông Du','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:57.165','2019-09-27 14:34:57.165')
,('THPT Huỳnh Thúc Kháng','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:57.166','2019-09-27 14:34:57.166')
,('THPT Minh Đức','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:57.166','2019-09-27 14:34:57.166')
,('THPT Nam Việt','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:57.167','2019-09-27 14:34:57.167')
,('THPT Nhân việt','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:57.167','2019-09-27 14:34:57.167')
,('TH-THCS-THPT Hoà Bình','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:57.168','2019-09-27 14:34:57.168')
,('TH-THCS-THPT Quốc văn Sài Gòn','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:57.168','2019-09-27 14:34:57.168')
,('THPT An Nghĩa','COUNTRY_VN',1,20,NULL,true,'2019-09-27 14:34:57.169','2019-09-27 14:34:57.169')
,('THPT Bình Khánh','COUNTRY_VN',1,20,NULL,true,'2019-09-27 14:34:57.170','2019-09-27 14:34:57.170')
,('THPT Cần Thạnh','COUNTRY_VN',1,20,NULL,true,'2019-09-27 14:34:57.170','2019-09-27 14:34:57.170')
,('TTGDTX Huyện Cần Giờ','COUNTRY_VN',1,20,NULL,true,'2019-09-27 14:34:57.170','2019-09-27 14:34:57.170')
,('THPT Củ Chi','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:57.171','2019-09-27 14:34:57.171')
,('THPT Phú Hòa','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:57.172','2019-09-27 14:34:57.172')
,('THPT Quang Trung','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:57.172','2019-09-27 14:34:57.172')
,('THPT Tân Thông Hội','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:57.173','2019-09-27 14:34:57.173')
,('THPT Trung Lập','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:57.173','2019-09-27 14:34:57.173')
,('THPT Trung Phú','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:57.173','2019-09-27 14:34:57.173')
,('TTGDTX Huyện Củ Chi','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:57.174','2019-09-27 14:34:57.174')
,('THPT An Nhơn Tây','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:57.176','2019-09-27 14:34:57.176')
,('THPT Dương Văn Dương','COUNTRY_VN',1,22,NULL,true,'2019-09-27 14:34:57.177','2019-09-27 14:34:57.177')
,('THPT Long Thới','COUNTRY_VN',1,22,NULL,true,'2019-09-27 14:34:57.178','2019-09-27 14:34:57.178')
,('THPT Phước Kiến','COUNTRY_VN',1,22,NULL,true,'2019-09-27 14:34:57.179','2019-09-27 14:34:57.179')
,('TTGDTX Huyện Nhà Bè','COUNTRY_VN',1,22,NULL,true,'2019-09-27 14:34:57.180','2019-09-27 14:34:57.180')
,('THPT Bắc Mỹ','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:57.181','2019-09-27 14:34:57.181')
,('THPT Bình Chánh','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:57.181','2019-09-27 14:34:57.181')
,('THPT Đa Phước','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:57.182','2019-09-27 14:34:57.182')
,('THPT Lê Minh Xuân','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:57.182','2019-09-27 14:34:57.182')
,('THPT TânTúc','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:57.183','2019-09-27 14:34:57.183')
,('THPT Vĩnh Lộc B','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:57.183','2019-09-27 14:34:57.183')
,('TTGDTX Huyện Bình Chánh','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:57.183','2019-09-27 14:34:57.183')
,('THPT Bà Điểm','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:57.184','2019-09-27 14:34:57.184')
,('THPT Lý Thường Kiệt','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:57.185','2019-09-27 14:34:57.185')
,('THPT Nguyễn Hữu Cầu','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:57.185','2019-09-27 14:34:57.185')
,('THPT Nguyễn Hữu Tiến','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:57.186','2019-09-27 14:34:57.186')
,('THPT Nguyễn Văn Cừ','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:57.186','2019-09-27 14:34:57.186')
,('THPT Phạm Văn Sáng','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:57.186','2019-09-27 14:34:57.186')
,('TTGDTX Huyện Hóc Môn','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:57.187','2019-09-27 14:34:57.187')
,('THPT Đinh Tiên Hoàng','COUNTRY_VN',2,25,NULL,true,'2019-09-27 14:34:57.188','2019-09-27 14:34:57.188')
,('THPT Hồ Tùng Mậu','COUNTRY_VN',2,25,NULL,true,'2019-09-27 14:34:57.188','2019-09-27 14:34:57.188')
,('THPT Nguyễn Trãi','COUNTRY_VN',2,25,NULL,true,'2019-09-27 14:34:57.188','2019-09-27 14:34:57.188')
,('THPT Phạm Hồng Thái','COUNTRY_VN',2,25,NULL,true,'2019-09-27 14:34:57.189','2019-09-27 14:34:57.189')
,('THPT Phan Đình Phùng','COUNTRY_VN',2,25,NULL,true,'2019-09-27 14:34:57.189','2019-09-27 14:34:57.189')
,('THPT Thực nghiệm','COUNTRY_VN',2,25,NULL,true,'2019-09-27 14:34:57.189','2019-09-27 14:34:57.189')
,('THPT Văn Lang','COUNTRY_VN',2,25,NULL,true,'2019-09-27 14:34:57.190','2019-09-27 14:34:57.190')
,('THCS-THPT Hà Thành','COUNTRY_VN',2,25,NULL,true,'2019-09-27 14:34:57.190','2019-09-27 14:34:57.190')
,('TTGDTX Quận Ba Đình','COUNTRY_VN',2,25,NULL,true,'2019-09-27 14:34:57.191','2019-09-27 14:34:57.191')
,('TTGDTX Nguyễn Văn Tố','COUNTRY_VN',2,26,NULL,true,'2019-09-27 14:34:57.193','2019-09-27 14:34:57.193')
,('THPT Marie Curie','COUNTRY_VN',2,26,NULL,true,'2019-09-27 14:34:57.193','2019-09-27 14:34:57.193')
,('THPT Trần Phú','COUNTRY_VN',2,26,NULL,true,'2019-09-27 14:34:57.194','2019-09-27 14:34:57.194')
,('THPT Văn Hiến','COUNTRY_VN',2,26,NULL,true,'2019-09-27 14:34:57.194','2019-09-27 14:34:57.194')
,('THPT Việt Đức','COUNTRY_VN',2,26,NULL,true,'2019-09-27 14:34:57.195','2019-09-27 14:34:57.195')
,('THPT Đông Kinh','COUNTRY_VN',2,27,NULL,true,'2019-09-27 14:34:57.196','2019-09-27 14:34:57.196')
,('THPT Hoàng Diệu','COUNTRY_VN',2,27,NULL,true,'2019-09-27 14:34:57.196','2019-09-27 14:34:57.196')
,('THPT Hồng Hà','COUNTRY_VN',2,27,NULL,true,'2019-09-27 14:34:57.197','2019-09-27 14:34:57.197')
,('THPT Mai Hắc Đế','COUNTRY_VN',2,27,NULL,true,'2019-09-27 14:34:57.197','2019-09-27 14:34:57.197')
,('THPT Ngô Gia Tự','COUNTRY_VN',2,27,NULL,true,'2019-09-27 14:34:57.198','2019-09-27 14:34:57.198')
,('THPT Thăng Long','COUNTRY_VN',2,27,NULL,true,'2019-09-27 14:34:57.198','2019-09-27 14:34:57.198')
,('THPT Đoàn Kết','COUNTRY_VN',2,27,NULL,true,'2019-09-27 14:34:57.199','2019-09-27 14:34:57.199')
,('TTGDTX Quận Hai Bà Trưng','COUNTRY_VN',2,27,NULL,true,'2019-09-27 14:34:57.199','2019-09-27 14:34:57.199')
,('THPT Hoàng cầu','COUNTRY_VN',2,28,NULL,true,'2019-09-27 14:34:57.200','2019-09-27 14:34:57.200')
,('THPT Kim Liên','COUNTRY_VN',2,28,NULL,true,'2019-09-27 14:34:57.201','2019-09-27 14:34:57.201')
,('THPT Lê Quý Đôn','COUNTRY_VN',2,28,NULL,true,'2019-09-27 14:34:57.201','2019-09-27 14:34:57.201')
,('THPT Nguyễn Văn Huyên','COUNTRY_VN',2,28,NULL,true,'2019-09-27 14:34:57.201','2019-09-27 14:34:57.201')
,('THPT Phan Huy Chú','COUNTRY_VN',2,28,NULL,true,'2019-09-27 14:34:57.202','2019-09-27 14:34:57.202')
,('THPT Quang Trung','COUNTRY_VN',2,28,NULL,true,'2019-09-27 14:34:57.202','2019-09-27 14:34:57.202')
,('THPT Tô Hiến Thành','COUNTRY_VN',2,28,NULL,true,'2019-09-27 14:34:57.202','2019-09-27 14:34:57.202')
,('THCS-THPT Alfred Nobel','COUNTRY_VN',2,28,NULL,true,'2019-09-27 14:34:57.203','2019-09-27 14:34:57.203')
,('THPT Băc Hà','COUNTRY_VN',2,28,NULL,true,'2019-09-27 14:34:57.203','2019-09-27 14:34:57.203')
,('THPT Đống Đa','COUNTRY_VN',2,28,NULL,true,'2019-09-27 14:34:57.204','2019-09-27 14:34:57.204')
,('THPT Einstein','COUNTRY_VN',2,28,NULL,true,'2019-09-27 14:34:57.204','2019-09-27 14:34:57.204')
,('TTGDTX Quận Đống Đa','COUNTRY_VN',2,28,NULL,true,'2019-09-27 14:34:57.204','2019-09-27 14:34:57.204')
,('TH-THCS-THPT Song ngữ Quốc tế Horizon','COUNTRY_VN',2,29,NULL,true,'2019-09-27 14:34:57.205','2019-09-27 14:34:57.205')
,('THPT Chu Văn An','COUNTRY_VN',2,29,NULL,true,'2019-09-27 14:34:57.205','2019-09-27 14:34:57.205')
,('THPT Đông Đô','COUNTRY_VN',2,29,NULL,true,'2019-09-27 14:34:57.206','2019-09-27 14:34:57.206')
,('THPT Hà Nội Academy','COUNTRY_VN',2,29,NULL,true,'2019-09-27 14:34:57.206','2019-09-27 14:34:57.206')
,('THPT Phan Chu Trinh','COUNTRY_VN',2,29,NULL,true,'2019-09-27 14:34:57.207','2019-09-27 14:34:57.207')
,('THPT Tây Hồ','COUNTRY_VN',2,29,NULL,true,'2019-09-27 14:34:57.207','2019-09-27 14:34:57.207')
,('TTGDTX Quận Tây Hồ','COUNTRY_VN',2,29,NULL,true,'2019-09-27 14:34:57.208','2019-09-27 14:34:57.208')
,('THPT Lương Thế Vinh','COUNTRY_VN',2,30,NULL,true,'2019-09-27 14:34:57.210','2019-09-27 14:34:57.210')
,('THPT Lý Thái Tổ','COUNTRY_VN',2,30,NULL,true,'2019-09-27 14:34:57.210','2019-09-27 14:34:57.210')
,('THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',2,30,NULL,true,'2019-09-27 14:34:57.211','2019-09-27 14:34:57.211')
,('THPT Nguyễn Siêu','COUNTRY_VN',2,30,NULL,true,'2019-09-27 14:34:57.211','2019-09-27 14:34:57.211')
,('THPT Phạm Văn Đồng','COUNTRY_VN',2,30,NULL,true,'2019-09-27 14:34:57.212','2019-09-27 14:34:57.212')
,('THPT Yên Hoà','COUNTRY_VN',2,30,NULL,true,'2019-09-27 14:34:57.212','2019-09-27 14:34:57.212')
,('THPT Cầu Giấy','COUNTRY_VN',2,30,NULL,true,'2019-09-27 14:34:57.213','2019-09-27 14:34:57.213')
,('THPT Chuyên Đại học Sư phạm','COUNTRY_VN',2,30,NULL,true,'2019-09-27 14:34:57.214','2019-09-27 14:34:57.214')
,('THPT Chuyên Hà Nội','COUNTRY_VN',2,30,NULL,true,'2019-09-27 14:34:57.214','2019-09-27 14:34:57.214')
,('THPT Chuyên Ngữ Đại học Ngoại ngữ','COUNTRY_VN',2,30,NULL,true,'2019-09-27 14:34:57.215','2019-09-27 14:34:57.215')
,('THPT Hermann Gmeiner','COUNTRY_VN',2,30,NULL,true,'2019-09-27 14:34:57.215','2019-09-27 14:34:57.215')
,('THPT Hồng Bàng','COUNTRY_VN',2,30,NULL,true,'2019-09-27 14:34:57.216','2019-09-27 14:34:57.216')
,('THCS-THPT Nguyễn Tất Thành','COUNTRY_VN',2,30,NULL,true,'2019-09-27 14:34:57.213','2019-09-27 14:34:57.216')
,('TTGDTX Quận Cầu Giấy','COUNTRY_VN',2,30,NULL,true,'2019-09-27 14:34:57.217','2019-09-27 14:34:57.217')
,('THPT Hồ Xuân Hương','COUNTRY_VN',2,31,NULL,true,'2019-09-27 14:34:57.218','2019-09-27 14:34:57.218')
,('THPT Huỳnh Thúc Kháng','COUNTRY_VN',2,31,NULL,true,'2019-09-27 14:34:57.218','2019-09-27 14:34:57.218')
,('THPT Lương Văn Can','COUNTRY_VN',2,31,NULL,true,'2019-09-27 14:34:57.218','2019-09-27 14:34:57.218')
,('THPT Nguyễn Trường Tộ','COUNTRY_VN',2,31,NULL,true,'2019-09-27 14:34:57.219','2019-09-27 14:34:57.219')
,('THPT Nhân Chính','COUNTRY_VN',2,31,NULL,true,'2019-09-27 14:34:57.219','2019-09-27 14:34:57.219')
,('THPT Phan Bội Châu','COUNTRY_VN',2,31,NULL,true,'2019-09-27 14:34:57.220','2019-09-27 14:34:57.220')
,('THPT Trần Hung Đạo','COUNTRY_VN',2,31,NULL,true,'2019-09-27 14:34:57.220','2019-09-27 14:34:57.220')
,('THPT Chuyên KHTN','COUNTRY_VN',2,31,NULL,true,'2019-09-27 14:34:57.221','2019-09-27 14:34:57.221')
,('THPT Dân lập Hà Nội','COUNTRY_VN',2,31,NULL,true,'2019-09-27 14:34:57.221','2019-09-27 14:34:57.221')
,('THPT Đại Việt','COUNTRY_VN',2,31,NULL,true,'2019-09-27 14:34:57.222','2019-09-27 14:34:57.222')
,('THPT Đào Duy Từ','COUNTRY_VN',2,31,NULL,true,'2019-09-27 14:34:57.222','2019-09-27 14:34:57.222')
,('THPT Đông Nam Á','COUNTRY_VN',2,31,NULL,true,'2019-09-27 14:34:57.223','2019-09-27 14:34:57.223')
,('TTGDTX Quận Thanh Xuân','COUNTRY_VN',2,31,NULL,true,'2019-09-27 14:34:57.223','2019-09-27 14:34:57.223')
,('THCS-THPT Quốc tế Thăng Long','COUNTRY_VN',2,32,NULL,true,'2019-09-27 14:34:57.225','2019-09-27 14:34:57.225')
,('THPT Dân lập Trần Quang Khải','COUNTRY_VN',2,32,NULL,true,'2019-09-27 14:34:57.226','2019-09-27 14:34:57.226')
,('THPT Hoàng Văn Thụ','COUNTRY_VN',2,32,NULL,true,'2019-09-27 14:34:57.226','2019-09-27 14:34:57.226')
,('THPT Nguyễn Đình Chiểu','COUNTRY_VN',2,32,NULL,true,'2019-09-27 14:34:57.227','2019-09-27 14:34:57.227')
,('THPT Phương Nam','COUNTRY_VN',2,32,NULL,true,'2019-09-27 14:34:57.227','2019-09-27 14:34:57.227')
,('THPT Trương Định','COUNTRY_VN',2,32,NULL,true,'2019-09-27 14:34:57.228','2019-09-27 14:34:57.228')
,('THPT Việt Nam-Ba Lan','COUNTRY_VN',2,32,NULL,true,'2019-09-27 14:34:57.229','2019-09-27 14:34:57.229')
,('TTGDTX Quận Hoàng Mai','COUNTRY_VN',2,32,NULL,true,'2019-09-27 14:34:57.229','2019-09-27 14:34:57.229')
,('THPT Lý Thường Kiệt','COUNTRY_VN',2,33,NULL,true,'2019-09-27 14:34:57.230','2019-09-27 14:34:57.230')
,('THPT Nguyễn Gia Thiều','COUNTRY_VN',2,33,NULL,true,'2019-09-27 14:34:57.231','2019-09-27 14:34:57.231')
,('THPT Tây Sơn','COUNTRY_VN',2,33,NULL,true,'2019-09-27 14:34:57.231','2019-09-27 14:34:57.231')
,('THPT Thạch Bàn','COUNTRY_VN',2,33,NULL,true,'2019-09-27 14:34:57.232','2019-09-27 14:34:57.232')
,('THPT Vạn Xuân','COUNTRY_VN',2,33,NULL,true,'2019-09-27 14:34:57.232','2019-09-27 14:34:57.232')
,('THPT Wellspring','COUNTRY_VN',2,33,NULL,true,'2019-09-27 14:34:57.233','2019-09-27 14:34:57.233')
,('THPT Lê Văn Thiêm','COUNTRY_VN',2,33,NULL,true,'2019-09-27 14:34:57.233','2019-09-27 14:34:57.233')
,('TTGDTX Quận Long Biên','COUNTRY_VN',2,33,NULL,true,'2019-09-27 14:34:57.234','2019-09-27 14:34:57.234')
,('THPT Dân lập Đoàn Thị Điểm','COUNTRY_VN',2,34,NULL,true,'2019-09-27 14:34:57.235','2019-09-27 14:34:57.235')
,('THPT Khoa học giáo dục','COUNTRY_VN',2,34,NULL,true,'2019-09-27 14:34:57.236','2019-09-27 14:34:57.236')
,('THPT Lê Thánh Tông','COUNTRY_VN',2,34,NULL,true,'2019-09-27 14:34:57.236','2019-09-27 14:34:57.236')
,('THPT Nguyễn Thị Minh Khai','COUNTRY_VN',2,34,NULL,true,'2019-09-27 14:34:57.237','2019-09-27 14:34:57.237')
,('THPT Tây Đô','COUNTRY_VN',2,34,NULL,true,'2019-09-27 14:34:57.237','2019-09-27 14:34:57.237')
,('THPT Thượng Cát','COUNTRY_VN',2,34,NULL,true,'2019-09-27 14:34:57.238','2019-09-27 14:34:57.238')
,('THPT Xuân Đỉnh','COUNTRY_VN',2,34,NULL,true,'2019-09-27 14:34:57.238','2019-09-27 14:34:57.238')
,('THPT M.V.Lômônôxốp','COUNTRY_VN',2,35,NULL,true,'2019-09-27 14:34:57.239','2019-09-27 14:34:57.239')
,('THPT Olympia','COUNTRY_VN',2,35,NULL,true,'2019-09-27 14:34:57.240','2019-09-27 14:34:57.240')
,('THPT Trần Thánh Tông','COUNTRY_VN',2,35,NULL,true,'2019-09-27 14:34:57.241','2019-09-27 14:34:57.241')
,('THPT Trí Đức','COUNTRY_VN',2,35,NULL,true,'2019-09-27 14:34:57.242','2019-09-27 14:34:57.242')
,('THPT Trung Văn','COUNTRY_VN',2,35,NULL,true,'2019-09-27 14:34:57.243','2019-09-27 14:34:57.243')
,('THPT Việt Úc Hà Nội','COUNTRY_VN',2,35,NULL,true,'2019-09-27 14:34:57.244','2019-09-27 14:34:57.244')
,('THPT Xuân Thuỷ','COUNTRY_VN',2,35,NULL,true,'2019-09-27 14:34:57.244','2019-09-27 14:34:57.244')
,('THCS-THPT Newton','COUNTRY_VN',2,35,NULL,true,'2019-09-27 14:34:57.245','2019-09-27 14:34:57.245')
,('THCS-THPT Trần Quốc Tuấn','COUNTRY_VN',2,35,NULL,true,'2019-09-27 14:34:57.245','2019-09-27 14:34:57.245')
,('THPT Đại Mổ','COUNTRY_VN',2,35,NULL,true,'2019-09-27 14:34:57.245','2019-09-27 14:34:57.245')
,('TTGDTX Từ Liêm','COUNTRY_VN',2,35,NULL,true,'2019-09-27 14:34:57.246','2019-09-27 14:34:57.246')
,('THPT Ngọc Hồi','COUNTRY_VN',2,36,NULL,true,'2019-09-27 14:34:57.247','2019-09-27 14:34:57.247')
,('THPT Ngô Thì Nhậm','COUNTRY_VN',2,36,NULL,true,'2019-09-27 14:34:57.248','2019-09-27 14:34:57.248')
,('TTGDTX Huyện Thanh Trì','COUNTRY_VN',2,36,NULL,true,'2019-09-27 14:34:57.249','2019-09-27 14:34:57.249')
,('TTGDTX Đông Mỹ','COUNTRY_VN',2,36,NULL,true,'2019-09-27 14:34:57.249','2019-09-27 14:34:57.249')
,('THPT Cao Bá Quát','COUNTRY_VN',2,37,NULL,true,'2019-09-27 14:34:57.251','2019-09-27 14:34:57.251')
,('THPT Dương Xá','COUNTRY_VN',2,37,NULL,true,'2019-09-27 14:34:57.251','2019-09-27 14:34:57.251')
,('THPT Lê Ngọc Hân','COUNTRY_VN',2,37,NULL,true,'2019-09-27 14:34:57.252','2019-09-27 14:34:57.252')
,('THPT Lý Thánh Tông','COUNTRY_VN',2,37,NULL,true,'2019-09-27 14:34:57.252','2019-09-27 14:34:57.252')
,('THPT Nguyễn Văn Cừ','COUNTRY_VN',2,37,NULL,true,'2019-09-27 14:34:57.253','2019-09-27 14:34:57.253')
,('THPT Tô Hiệu','COUNTRY_VN',2,37,NULL,true,'2019-09-27 14:34:57.253','2019-09-27 14:34:57.253')
,('THPT Yên Viên','COUNTRY_VN',2,37,NULL,true,'2019-09-27 14:34:57.253','2019-09-27 14:34:57.253')
,('TTGDTX Đình Xuyên','COUNTRY_VN',2,37,NULL,true,'2019-09-27 14:34:57.254','2019-09-27 14:34:57.254')
,('TTGDTX Phú Thị','COUNTRY_VN',2,37,NULL,true,'2019-09-27 14:34:57.255','2019-09-27 14:34:57.255')
,('THPT Bắc Đuống','COUNTRY_VN',2,37,NULL,true,'2019-09-27 14:34:57.255','2019-09-27 14:34:57.255')
,('THPT Hoàng Long','COUNTRY_VN',2,38,NULL,true,'2019-09-27 14:34:57.255','2019-09-27 14:34:57.255')
,('THPT Lê Hồng Phong','COUNTRY_VN',2,38,NULL,true,'2019-09-27 14:34:57.256','2019-09-27 14:34:57.256')
,('THPT Liên Hà','COUNTRY_VN',2,38,NULL,true,'2019-09-27 14:34:57.256','2019-09-27 14:34:57.256')
,('THPT Ngô Quyền','COUNTRY_VN',2,38,NULL,true,'2019-09-27 14:34:57.257','2019-09-27 14:34:57.257')
,('THPT Ngô Tất Tố','COUNTRY_VN',2,38,NULL,true,'2019-09-27 14:34:57.257','2019-09-27 14:34:57.257')
,('THPT Phạm Ngũ Lão','COUNTRY_VN',2,38,NULL,true,'2019-09-27 14:34:57.259','2019-09-27 14:34:57.259')
,('TTGDTX Huyện Đông Anh','COUNTRY_VN',2,38,NULL,true,'2019-09-27 14:34:57.259','2019-09-27 14:34:57.259')
,('THPT An Dương Vương','COUNTRY_VN',2,38,NULL,true,'2019-09-27 14:34:57.260','2019-09-27 14:34:57.260')
,('THPT Bắc Thăng Long','COUNTRY_VN',2,38,NULL,true,'2019-09-27 14:34:57.260','2019-09-27 14:34:57.260')
,('THPT Cổ Loa','COUNTRY_VN',2,38,NULL,true,'2019-09-27 14:34:57.261','2019-09-27 14:34:57.261')
,('THPT Đông Anh','COUNTRY_VN',2,38,NULL,true,'2019-09-27 14:34:57.261','2019-09-27 14:34:57.261')
,('THPT Lam Hồng','COUNTRY_VN',2,39,NULL,true,'2019-09-27 14:34:57.262','2019-09-27 14:34:57.262')
,('THPT Mạc Đĩnh Chi','COUNTRY_VN',2,39,NULL,true,'2019-09-27 14:34:57.263','2019-09-27 14:34:57.263')
,('THPT Minh Phú','COUNTRY_VN',2,39,NULL,true,'2019-09-27 14:34:57.263','2019-09-27 14:34:57.263')
,('THPT Minh Trí','COUNTRY_VN',2,39,NULL,true,'2019-09-27 14:34:57.264','2019-09-27 14:34:57.264')
,('THPT Sóc Sơn','COUNTRY_VN',2,39,NULL,true,'2019-09-27 14:34:57.265','2019-09-27 14:34:57.265')
,('THPT Trung Giã','COUNTRY_VN',2,39,NULL,true,'2019-09-27 14:34:57.265','2019-09-27 14:34:57.265')
,('THPT Xuân Giang','COUNTRY_VN',2,39,NULL,true,'2019-09-27 14:34:57.266','2019-09-27 14:34:57.266')
,('THPT Dân lập Đặng Thai Mai','COUNTRY_VN',2,39,NULL,true,'2019-09-27 14:34:57.266','2019-09-27 14:34:57.266')
,('THPT Dân lập Nguyễn Thượng Hiền','COUNTRY_VN',2,39,NULL,true,'2019-09-27 14:34:57.266','2019-09-27 14:34:57.266')
,('THPT Dân lập Phùng Khăc Khoan','COUNTRY_VN',2,39,NULL,true,'2019-09-27 14:34:57.267','2019-09-27 14:34:57.267')
,('THPT Đa Phúc','COUNTRY_VN',2,39,NULL,true,'2019-09-27 14:34:57.267','2019-09-27 14:34:57.267')
,('THPT Kim Anh','COUNTRY_VN',2,39,NULL,true,'2019-09-27 14:34:57.268','2019-09-27 14:34:57.268')
,('THPT Lạc Long Quân','COUNTRY_VN',2,39,NULL,true,'2019-09-27 14:34:57.268','2019-09-27 14:34:57.268')
,('TTGDTX Huyện Sóc Sơn','COUNTRY_VN',2,39,NULL,true,'2019-09-27 14:34:57.269','2019-09-27 14:34:57.269')
,('THPT Hà Đông','COUNTRY_VN',2,40,NULL,true,'2019-09-27 14:34:57.270','2019-09-27 14:34:57.270')
,('THPT Lê Lợi','COUNTRY_VN',2,40,NULL,true,'2019-09-27 14:34:57.270','2019-09-27 14:34:57.270')
,('THPT Lê Quý Đôn','COUNTRY_VN',2,40,NULL,true,'2019-09-27 14:34:57.271','2019-09-27 14:34:57.271')
,('THPT Phùng Hưng','COUNTRY_VN',2,40,NULL,true,'2019-09-27 14:34:57.271','2019-09-27 14:34:57.271')
,('THPT Quang Trung','COUNTRY_VN',2,40,NULL,true,'2019-09-27 14:34:57.272','2019-09-27 14:34:57.272')
,('THPT Trần Hưng Đạo','COUNTRY_VN',2,40,NULL,true,'2019-09-27 14:34:57.272','2019-09-27 14:34:57.272')
,('THPT Xa La','COUNTRY_VN',2,40,NULL,true,'2019-09-27 14:34:57.273','2019-09-27 14:34:57.273')
,('THPT Chuyên Nguyễn Huệ','COUNTRY_VN',2,40,NULL,true,'2019-09-27 14:34:57.273','2019-09-27 14:34:57.273')
,('THPT Quốc tế Việt Nam','COUNTRY_VN',2,40,NULL,true,'2019-09-27 14:34:57.274','2019-09-27 14:34:57.274')
,('TTGDTX Hà Tây','COUNTRY_VN',2,40,NULL,true,'2019-09-27 14:34:57.276','2019-09-27 14:34:57.276')
,('THPT Xuân Khanh','COUNTRY_VN',2,41,NULL,true,'2019-09-27 14:34:57.277','2019-09-27 14:34:57.277')
,('THPT Tùng Thiện','COUNTRY_VN',2,41,NULL,true,'2019-09-27 14:34:57.278','2019-09-27 14:34:57.278')
,('THPT Sơn Tây','COUNTRY_VN',2,41,NULL,true,'2019-09-27 14:34:57.278','2019-09-27 14:34:57.278')
,('THPT Nguyễn Tất Thành','COUNTRY_VN',2,41,NULL,true,'2019-09-27 14:34:57.279','2019-09-27 14:34:57.279')
,('THPT Võ thuật Bảo Long','COUNTRY_VN',2,41,NULL,true,'2019-09-27 14:34:57.279','2019-09-27 14:34:57.279')
,('TTGDTX Thị xã Sơn Tây','COUNTRY_VN',2,41,NULL,true,'2019-09-27 14:34:57.280','2019-09-27 14:34:57.280')
,('TTGDTX Huyện Ba Vì','COUNTRY_VN',2,42,NULL,true,'2019-09-27 14:34:57.281','2019-09-27 14:34:57.281')
,('THPT Dân tộc Nội trú Hà Nội','COUNTRY_VN',2,42,NULL,true,'2019-09-27 14:34:57.282','2019-09-27 14:34:57.282')
,('THPT Ba Vì','COUNTRY_VN',2,42,NULL,true,'2019-09-27 14:34:57.282','2019-09-27 14:34:57.282')
,('THPT Bất Bạt','COUNTRY_VN',2,42,NULL,true,'2019-09-27 14:34:57.283','2019-09-27 14:34:57.283')
,('THPT Lương Thế Vinh','COUNTRY_VN',2,42,NULL,true,'2019-09-27 14:34:57.283','2019-09-27 14:34:57.283')
,('THPT Ngô Quyền','COUNTRY_VN',2,42,NULL,true,'2019-09-27 14:34:57.284','2019-09-27 14:34:57.284')
,('THPT Quảng Oai','COUNTRY_VN',2,42,NULL,true,'2019-09-27 14:34:57.284','2019-09-27 14:34:57.284')
,('THPT Trần Phú','COUNTRY_VN',2,42,NULL,true,'2019-09-27 14:34:57.284','2019-09-27 14:34:57.284')
,('TTGDTX Huyện Phúc Thọ','COUNTRY_VN',2,43,NULL,true,'2019-09-27 14:34:57.286','2019-09-27 14:34:57.286')
,('Hũu Nghị T78','COUNTRY_VN',2,43,NULL,true,'2019-09-27 14:34:57.286','2019-09-27 14:34:57.286')
,('THPT Hồng Đức','COUNTRY_VN',2,43,NULL,true,'2019-09-27 14:34:57.286','2019-09-27 14:34:57.286')
,('THPT Ngọc Tảo','COUNTRY_VN',2,43,NULL,true,'2019-09-27 14:34:57.287','2019-09-27 14:34:57.287')
,('THPT Phúc Thọ','COUNTRY_VN',2,43,NULL,true,'2019-09-27 14:34:57.287','2019-09-27 14:34:57.287')
,('THPT Vân Cốc','COUNTRY_VN',2,43,NULL,true,'2019-09-27 14:34:57.288','2019-09-27 14:34:57.288')
,('THPT Băc Lương Sơn','COUNTRY_VN',2,44,NULL,true,'2019-09-27 14:34:57.288','2019-09-27 14:34:57.288')
,('THPT FPT','COUNTRY_VN',2,44,NULL,true,'2019-09-27 14:34:57.289','2019-09-27 14:34:57.289')
,('THPT Hai Bà Trưng','COUNTRY_VN',2,44,NULL,true,'2019-09-27 14:34:57.289','2019-09-27 14:34:57.289')
,('THPT Phan Huy Chú','COUNTRY_VN',2,44,NULL,true,'2019-09-27 14:34:57.289','2019-09-27 14:34:57.289')
,('THPT Phùng Khắc Khoan','COUNTRY_VN',2,44,NULL,true,'2019-09-27 14:34:57.290','2019-09-27 14:34:57.290')
,('THPT Thạch Thất','COUNTRY_VN',2,44,NULL,true,'2019-09-27 14:34:57.290','2019-09-27 14:34:57.290')
,('TTGDTX Huyện Thạch Thất','COUNTRY_VN',2,44,NULL,true,'2019-09-27 14:34:57.290','2019-09-27 14:34:57.290')
,('THPT Phú Bình','COUNTRY_VN',2,44,NULL,true,'2019-09-27 14:34:57.292','2019-09-27 14:34:57.292')
,('TTGDTX Huyện Quốc Oai','COUNTRY_VN',2,45,NULL,true,'2019-09-27 14:34:57.295','2019-09-27 14:34:57.295')
,('THPT Nguyễn Trực','COUNTRY_VN',2,45,NULL,true,'2019-09-27 14:34:57.295','2019-09-27 14:34:57.295')
,('THPT Cao Bá Quát','COUNTRY_VN',2,45,NULL,true,'2019-09-27 14:34:57.296','2019-09-27 14:34:57.296')
,('THPT Minh Khai','COUNTRY_VN',2,45,NULL,true,'2019-09-27 14:34:57.296','2019-09-27 14:34:57.296')
,('THPT Quốc Oai','COUNTRY_VN',2,45,NULL,true,'2019-09-27 14:34:57.297','2019-09-27 14:34:57.297')
,('THPT Tư thục Minh Khai','COUNTRY_VN',2,45,NULL,true,'2019-09-27 14:34:57.297','2019-09-27 14:34:57.297')
,('THPT Chúc Động','COUNTRY_VN',2,46,NULL,true,'2019-09-27 14:34:57.298','2019-09-27 14:34:57.298')
,('THPT Chương Mỹ A','COUNTRY_VN',2,46,NULL,true,'2019-09-27 14:34:57.298','2019-09-27 14:34:57.298')
,('THPT Chương Mỹ B','COUNTRY_VN',2,46,NULL,true,'2019-09-27 14:34:57.299','2019-09-27 14:34:57.299')
,('THPT Đặng Tiến Đông','COUNTRY_VN',2,46,NULL,true,'2019-09-27 14:34:57.300','2019-09-27 14:34:57.300')
,('THPT Ngô Sỹ Liên','COUNTRY_VN',2,46,NULL,true,'2019-09-27 14:34:57.300','2019-09-27 14:34:57.300')
,('THPT Trần Đại Nghĩa','COUNTRY_VN',2,46,NULL,true,'2019-09-27 14:34:57.301','2019-09-27 14:34:57.301')
,('THPT Xuân Mai','COUNTRY_VN',2,46,NULL,true,'2019-09-27 14:34:57.301','2019-09-27 14:34:57.301')
,('TTGDTX Huyện Chương Mỹ','COUNTRY_VN',2,46,NULL,true,'2019-09-27 14:34:57.302','2019-09-27 14:34:57.302')
,('TTGDTX Huyện Đan Phượng','COUNTRY_VN',2,47,NULL,true,'2019-09-27 14:34:57.302','2019-09-27 14:34:57.302')
,('THPT Đan Phượng','COUNTRY_VN',2,47,NULL,true,'2019-09-27 14:34:57.303','2019-09-27 14:34:57.303')
,('THPT Hồng Thái','COUNTRY_VN',2,47,NULL,true,'2019-09-27 14:34:57.303','2019-09-27 14:34:57.303')
,('THPT Tân Lập','COUNTRY_VN',2,47,NULL,true,'2019-09-27 14:34:57.304','2019-09-27 14:34:57.304')
,('TTGDTX Huyện Hoài Đức','COUNTRY_VN',2,48,NULL,true,'2019-09-27 14:34:57.304','2019-09-27 14:34:57.304')
,('THPT Bình Minh','COUNTRY_VN',2,48,NULL,true,'2019-09-27 14:34:57.305','2019-09-27 14:34:57.305')
,('THPT Hoài Đức A','COUNTRY_VN',2,48,NULL,true,'2019-09-27 14:34:57.305','2019-09-27 14:34:57.305')
,('THPT Hoài Đức B','COUNTRY_VN',2,48,NULL,true,'2019-09-27 14:34:57.306','2019-09-27 14:34:57.306')
,('THPT Vạn Xuân','COUNTRY_VN',2,48,NULL,true,'2019-09-27 14:34:57.306','2019-09-27 14:34:57.306')
,('TTGDTX Huyện Thanh Oai','COUNTRY_VN',2,49,NULL,true,'2019-09-27 14:34:57.308','2019-09-27 14:34:57.308')
,('THPT Bắc Hà','COUNTRY_VN',2,49,NULL,true,'2019-09-27 14:34:57.309','2019-09-27 14:34:57.309')
,('THPT Nguyễn Du','COUNTRY_VN',2,49,NULL,true,'2019-09-27 14:34:57.310','2019-09-27 14:34:57.310')
,('THPT Thanh Oai A','COUNTRY_VN',2,49,NULL,true,'2019-09-27 14:34:57.311','2019-09-27 14:34:57.311')
,('THPT Thanh Oai B','COUNTRY_VN',2,49,NULL,true,'2019-09-27 14:34:57.311','2019-09-27 14:34:57.311')
,('THPT Thanh Xuân','COUNTRY_VN',2,49,NULL,true,'2019-09-27 14:34:57.313','2019-09-27 14:34:57.313')
,('TTGDTX Huyện Mỹ Đức','COUNTRY_VN',2,50,NULL,true,'2019-09-27 14:34:57.315','2019-09-27 14:34:57.315')
,('THPT Đinh Tiên Hoàng','COUNTRY_VN',2,50,NULL,true,'2019-09-27 14:34:57.316','2019-09-27 14:34:57.316')
,('THPT Hợp Thanh','COUNTRY_VN',2,50,NULL,true,'2019-09-27 14:34:57.316','2019-09-27 14:34:57.316')
,('THPT Mỹ Đức A','COUNTRY_VN',2,50,NULL,true,'2019-09-27 14:34:57.317','2019-09-27 14:34:57.317')
,('THPT Mỹ Đức B','COUNTRY_VN',2,50,NULL,true,'2019-09-27 14:34:57.317','2019-09-27 14:34:57.317')
,('THPT Mỹ Đức C','COUNTRY_VN',2,50,NULL,true,'2019-09-27 14:34:57.318','2019-09-27 14:34:57.318')
,('TTGDTX Huyện Ứng Hoà','COUNTRY_VN',2,51,NULL,true,'2019-09-27 14:34:57.318','2019-09-27 14:34:57.318')
,('THPT Đại Cường','COUNTRY_VN',2,51,NULL,true,'2019-09-27 14:34:57.319','2019-09-27 14:34:57.319')
,('THPT Lưu Hoàng','COUNTRY_VN',2,51,NULL,true,'2019-09-27 14:34:57.319','2019-09-27 14:34:57.319')
,('THPT Nguyễn Thuợng Hiền','COUNTRY_VN',2,51,NULL,true,'2019-09-27 14:34:57.319','2019-09-27 14:34:57.319')
,('THPT Trần Đăng Ninh','COUNTRY_VN',2,51,NULL,true,'2019-09-27 14:34:57.320','2019-09-27 14:34:57.320')
,('THPT Ứng Hoà A','COUNTRY_VN',2,51,NULL,true,'2019-09-27 14:34:57.320','2019-09-27 14:34:57.320')
,('THPT Ứng Hoà B','COUNTRY_VN',2,51,NULL,true,'2019-09-27 14:34:57.320','2019-09-27 14:34:57.320')
,('TTGDTX Huyện Thường Tín','COUNTRY_VN',2,52,NULL,true,'2019-09-27 14:34:57.321','2019-09-27 14:34:57.321')
,('THPT Lý Tử Tấn','COUNTRY_VN',2,52,NULL,true,'2019-09-27 14:34:57.321','2019-09-27 14:34:57.321')
,('THPT Nguyễn Trãi','COUNTRY_VN',2,52,NULL,true,'2019-09-27 14:34:57.322','2019-09-27 14:34:57.322')
,('THPT Thường Tín','COUNTRY_VN',2,52,NULL,true,'2019-09-27 14:34:57.322','2019-09-27 14:34:57.322')
,('THPT Tô Hiệu','COUNTRY_VN',2,52,NULL,true,'2019-09-27 14:34:57.323','2019-09-27 14:34:57.323')
,('THPT Vân Tảo','COUNTRY_VN',2,52,NULL,true,'2019-09-27 14:34:57.323','2019-09-27 14:34:57.323')
,('TTGDTX Huyện Phú Xuyên','COUNTRY_VN',2,53,NULL,true,'2019-09-27 14:34:57.324','2019-09-27 14:34:57.324')
,('THPT Đồng Quan','COUNTRY_VN',2,53,NULL,true,'2019-09-27 14:34:57.326','2019-09-27 14:34:57.326')
,('THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',2,53,NULL,true,'2019-09-27 14:34:57.327','2019-09-27 14:34:57.327')
,('THPT Phú Xuyên A','COUNTRY_VN',2,53,NULL,true,'2019-09-27 14:34:57.328','2019-09-27 14:34:57.328')
,('THPT Phú Xuyên B','COUNTRY_VN',2,53,NULL,true,'2019-09-27 14:34:57.329','2019-09-27 14:34:57.329')
,('THPT Tân Dân','COUNTRY_VN',2,53,NULL,true,'2019-09-27 14:34:57.330','2019-09-27 14:34:57.330')
,('TTGDTX Huyện Mê Linh','COUNTRY_VN',2,54,NULL,true,'2019-09-27 14:34:57.331','2019-09-27 14:34:57.331')
,('THPT Mê Linh','COUNTRY_VN',2,54,NULL,true,'2019-09-27 14:34:57.332','2019-09-27 14:34:57.332')
,('THPT Nguyễn Du','COUNTRY_VN',2,54,NULL,true,'2019-09-27 14:34:57.333','2019-09-27 14:34:57.333')
,('THPT Quang Minh','COUNTRY_VN',2,54,NULL,true,'2019-09-27 14:34:57.333','2019-09-27 14:34:57.333')
,('THPT Tiền Phong','COUNTRY_VN',2,54,NULL,true,'2019-09-27 14:34:57.334','2019-09-27 14:34:57.334')
,('THPT Tiến Thịnh','COUNTRY_VN',2,54,NULL,true,'2019-09-27 14:34:57.334','2019-09-27 14:34:57.334')
,('THPT Tự Lập','COUNTRY_VN',2,54,NULL,true,'2019-09-27 14:34:57.335','2019-09-27 14:34:57.335')
,('THPT Yên Lăng','COUNTRY_VN',2,54,NULL,true,'2019-09-27 14:34:57.335','2019-09-27 14:34:57.335')
,('THPT Đồ Sơn','COUNTRY_VN',3,55,NULL,true,'2019-09-27 14:34:57.337','2019-09-27 14:34:57.337')
,('THPT Nội Trú Đồ Sơn','COUNTRY_VN',3,55,NULL,true,'2019-09-27 14:34:57.337','2019-09-27 14:34:57.337')
,('TTGDTX Quận Đồ Sơn','COUNTRY_VN',3,55,NULL,true,'2019-09-27 14:34:57.338','2019-09-27 14:34:57.338')
,('THPT Mạc Đĩnh Chi','COUNTRY_VN',3,56,NULL,true,'2019-09-27 14:34:57.339','2019-09-27 14:34:57.339')
,('TTGDTX Quận Dương Kinh','COUNTRY_VN',3,56,NULL,true,'2019-09-27 14:34:57.339','2019-09-27 14:34:57.339')
,('THPT Hải An','COUNTRY_VN',3,57,NULL,true,'2019-09-27 14:34:57.340','2019-09-27 14:34:57.340')
,('THPT Lê Quý Đôn','COUNTRY_VN',3,57,NULL,true,'2019-09-27 14:34:57.340','2019-09-27 14:34:57.340')
,('THPT Phan Chu Trinh','COUNTRY_VN',3,57,NULL,true,'2019-09-27 14:34:57.340','2019-09-27 14:34:57.340')
,('TTGDTX Quận Hải An','COUNTRY_VN',3,57,NULL,true,'2019-09-27 14:34:57.342','2019-09-27 14:34:57.342')
,('THPT Hồng Bàng','COUNTRY_VN',3,58,NULL,true,'2019-09-27 14:34:57.343','2019-09-27 14:34:57.343')
,('THPT Lê Hồng Phong','COUNTRY_VN',3,58,NULL,true,'2019-09-27 14:34:57.344','2019-09-27 14:34:57.344')
,('THPT Lương Thế Vinh','COUNTRY_VN',3,58,NULL,true,'2019-09-27 14:34:57.344','2019-09-27 14:34:57.344')
,('TTGDTX Quận Hồng Bàng','COUNTRY_VN',3,58,NULL,true,'2019-09-27 14:34:57.345','2019-09-27 14:34:57.345')
,('THPT Đồng Hòa','COUNTRY_VN',3,59,NULL,true,'2019-09-27 14:34:57.346','2019-09-27 14:34:57.346')
,('THPT Kiến An','COUNTRY_VN',3,59,NULL,true,'2019-09-27 14:34:57.347','2019-09-27 14:34:57.347')
,('THPT Phan Đăng Lưu','COUNTRY_VN',3,59,NULL,true,'2019-09-27 14:34:57.347','2019-09-27 14:34:57.347')
,('TTGDTX Quận Kiến An','COUNTRY_VN',3,59,NULL,true,'2019-09-27 14:34:57.349','2019-09-27 14:34:57.349')
,('THPT NCH Nguyễn Tất Thành','COUNTRY_VN',3,60,NULL,true,'2019-09-27 14:34:57.350','2019-09-27 14:34:57.350')
,('THPT Lê Chân','COUNTRY_VN',3,60,NULL,true,'2019-09-27 14:34:57.351','2019-09-27 14:34:57.351')
,('THPT Lý Thái Tổ','COUNTRY_VN',3,60,NULL,true,'2019-09-27 14:34:57.351','2019-09-27 14:34:57.351')
,('THPT Ngô Quyền','COUNTRY_VN',3,60,NULL,true,'2019-09-27 14:34:57.352','2019-09-27 14:34:57.352')
,('THPT Trần Nguyên Hãn','COUNTRY_VN',3,60,NULL,true,'2019-09-27 14:34:57.352','2019-09-27 14:34:57.352')
,('TTGDTX Quận Lê Chân','COUNTRY_VN',3,60,NULL,true,'2019-09-27 14:34:57.353','2019-09-27 14:34:57.353')
,('TTGDTX Thành phố Hải Phòng','COUNTRY_VN',3,60,NULL,true,'2019-09-27 14:34:57.353','2019-09-27 14:34:57.353')
,('THPT Anhxtanh','COUNTRY_VN',3,61,NULL,true,'2019-09-27 14:34:57.354','2019-09-27 14:34:57.354')
,('THPT Chuyên Trần Phú','COUNTRY_VN',3,61,NULL,true,'2019-09-27 14:34:57.355','2019-09-27 14:34:57.355')
,('THPT Hàng Hải','COUNTRY_VN',3,61,NULL,true,'2019-09-27 14:34:57.355','2019-09-27 14:34:57.355')
,('THPT Hermann Gmeiner','COUNTRY_VN',3,61,NULL,true,'2019-09-27 14:34:57.355','2019-09-27 14:34:57.355')
,('THPT Lương Khánh Thiện','COUNTRY_VN',3,61,NULL,true,'2019-09-27 14:34:57.356','2019-09-27 14:34:57.356')
,('THPT Marie Curie','COUNTRY_VN',3,61,NULL,true,'2019-09-27 14:34:57.356','2019-09-27 14:34:57.356')
,('THPT Thái Phiên','COUNTRY_VN',3,61,NULL,true,'2019-09-27 14:34:57.356','2019-09-27 14:34:57.356')
,('THPT Thăng Long','COUNTRY_VN',3,61,NULL,true,'2019-09-27 14:34:57.357','2019-09-27 14:34:57.357')
,('TTGDTX Quận Ngô Quyền','COUNTRY_VN',3,61,NULL,true,'2019-09-27 14:34:57.358','2019-09-27 14:34:57.358')
,('THPT An Dương Vương','COUNTRY_VN',3,62,NULL,true,'2019-09-27 14:34:57.360','2019-09-27 14:34:57.360')
,('THPT An Hải','COUNTRY_VN',3,62,NULL,true,'2019-09-27 14:34:57.360','2019-09-27 14:34:57.360')
,('THPT Nguyễn Trãi','COUNTRY_VN',3,62,NULL,true,'2019-09-27 14:34:57.361','2019-09-27 14:34:57.361')
,('THPT Tân An','COUNTRY_VN',3,62,NULL,true,'2019-09-27 14:34:57.361','2019-09-27 14:34:57.361')
,('TTGDTX Huyện An Dương','COUNTRY_VN',3,62,NULL,true,'2019-09-27 14:34:57.362','2019-09-27 14:34:57.362')
,('THPT An Lão','COUNTRY_VN',3,63,NULL,true,'2019-09-27 14:34:57.364','2019-09-27 14:34:57.364')
,('THPT Quốc Tuấn','COUNTRY_VN',3,63,NULL,true,'2019-09-27 14:34:57.364','2019-09-27 14:34:57.364')
,('THPT Trần Hưng Đạo','COUNTRY_VN',3,63,NULL,true,'2019-09-27 14:34:57.365','2019-09-27 14:34:57.365')
,('THPT Trần Tất Văn','COUNTRY_VN',3,63,NULL,true,'2019-09-27 14:34:57.366','2019-09-27 14:34:57.366')
,('TTGDTX Huyện An Lão','COUNTRY_VN',3,63,NULL,true,'2019-09-27 14:34:57.366','2019-09-27 14:34:57.366')
,('THPT Cát Bà','COUNTRY_VN',3,64,NULL,true,'2019-09-27 14:34:57.367','2019-09-27 14:34:57.367')
,('THPT Cát Hải','COUNTRY_VN',3,64,NULL,true,'2019-09-27 14:34:57.367','2019-09-27 14:34:57.367')
,('TTGDTX Huyện Cát Hải','COUNTRY_VN',3,64,NULL,true,'2019-09-27 14:34:57.368','2019-09-27 14:34:57.368')
,('THPT Kiến Thụy','COUNTRY_VN',3,65,NULL,true,'2019-09-27 14:34:57.369','2019-09-27 14:34:57.369')
,('THPT Nguyễn Đức Cảnh','COUNTRY_VN',3,65,NULL,true,'2019-09-27 14:34:57.369','2019-09-27 14:34:57.369')
,('THPT Nguyễn Huệ','COUNTRY_VN',3,65,NULL,true,'2019-09-27 14:34:57.370','2019-09-27 14:34:57.370')
,('THPT Thụy Hương','COUNTRY_VN',3,65,NULL,true,'2019-09-27 14:34:57.370','2019-09-27 14:34:57.370')
,('TTGDTX Huyện Kiến Thụy','COUNTRY_VN',3,65,NULL,true,'2019-09-27 14:34:57.371','2019-09-27 14:34:57.371')
,('THPT 25/10','COUNTRY_VN',3,66,NULL,true,'2019-09-27 14:34:57.371','2019-09-27 14:34:57.371')
,('THPT Bạch Đằng','COUNTRY_VN',3,66,NULL,true,'2019-09-27 14:34:57.372','2019-09-27 14:34:57.372')
,('THPT Lê Ích Mộc','COUNTRY_VN',3,66,NULL,true,'2019-09-27 14:34:57.372','2019-09-27 14:34:57.372')
,('THPT Lý Thường Kiệt','COUNTRY_VN',3,66,NULL,true,'2019-09-27 14:34:57.373','2019-09-27 14:34:57.373')
,('THPT Nam Triệu','COUNTRY_VN',3,66,NULL,true,'2019-09-27 14:34:57.373','2019-09-27 14:34:57.373')
,('THPT Phạm Ngũ Lão','COUNTRY_VN',3,66,NULL,true,'2019-09-27 14:34:57.374','2019-09-27 14:34:57.374')
,('THPT Quang Trung','COUNTRY_VN',3,66,NULL,true,'2019-09-27 14:34:57.376','2019-09-27 14:34:57.376')
,('THPT Thủy Sơn','COUNTRY_VN',3,66,NULL,true,'2019-09-27 14:34:57.376','2019-09-27 14:34:57.376')
,('TTGDTX Huyện Thủy Nguyên','COUNTRY_VN',3,66,NULL,true,'2019-09-27 14:34:57.377','2019-09-27 14:34:57.377')
,('THPT Hùng Thắng','COUNTRY_VN',3,67,NULL,true,'2019-09-27 14:34:57.378','2019-09-27 14:34:57.378')
,('THPT Nhữ Văn Lan','COUNTRY_VN',3,67,NULL,true,'2019-09-27 14:34:57.378','2019-09-27 14:34:57.378')
,('THPT Tiên Lãng','COUNTRY_VN',3,67,NULL,true,'2019-09-27 14:34:57.380','2019-09-27 14:34:57.380')
,('THPT Toàn Thắng','COUNTRY_VN',3,67,NULL,true,'2019-09-27 14:34:57.381','2019-09-27 14:34:57.381')
,('TTGDTX Huyện Tiên Lãng','COUNTRY_VN',3,67,NULL,true,'2019-09-27 14:34:57.382','2019-09-27 14:34:57.382')
,('THPT Cộng Hiền','COUNTRY_VN',3,68,NULL,true,'2019-09-27 14:34:57.383','2019-09-27 14:34:57.383')
,('THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',3,68,NULL,true,'2019-09-27 14:34:57.383','2019-09-27 14:34:57.383')
,('THPT Nguyễn Khuyến','COUNTRY_VN',3,68,NULL,true,'2019-09-27 14:34:57.384','2019-09-27 14:34:57.384')
,('THPT Tô Hiệu','COUNTRY_VN',3,68,NULL,true,'2019-09-27 14:34:57.385','2019-09-27 14:34:57.385')
,('TPT Vĩnh Bảo','COUNTRY_VN',3,68,NULL,true,'2019-09-27 14:34:57.386','2019-09-27 14:34:57.386')
,('TTGDTX Huyện Vĩnh Bảo','COUNTRY_VN',3,68,NULL,true,'2019-09-27 14:34:57.386','2019-09-27 14:34:57.386')
,('THPT Cẩm Lệ','COUNTRY_VN',4,69,NULL,true,'2019-09-27 14:34:57.387','2019-09-27 14:34:57.387')
,('THPT Hòa Vang','COUNTRY_VN',4,69,NULL,true,'2019-09-27 14:34:57.387','2019-09-27 14:34:57.387')
,('TTGDTX Quận Cẩm Lệ','COUNTRY_VN',4,69,NULL,true,'2019-09-27 14:34:57.388','2019-09-27 14:34:57.388')
,('THPT Nguyễn Hiền','COUNTRY_VN',4,70,NULL,true,'2019-09-27 14:34:57.388','2019-09-27 14:34:57.388')
,('THPT Phan Châu Trinh','COUNTRY_VN',4,70,NULL,true,'2019-09-27 14:34:57.389','2019-09-27 14:34:57.389')
,('THPT Trần Phú','COUNTRY_VN',4,70,NULL,true,'2019-09-27 14:34:57.389','2019-09-27 14:34:57.389')
,('THPT Tư thục Diên Hồng','COUNTRY_VN',4,70,NULL,true,'2019-09-27 14:34:57.389','2019-09-27 14:34:57.389')
,('TTGDTX Quận Hải Châu','COUNTRY_VN',4,70,NULL,true,'2019-09-27 14:34:57.390','2019-09-27 14:34:57.390')
,('THPT Nguyễn Thượng Hiền','COUNTRY_VN',4,71,NULL,true,'2019-09-27 14:34:57.391','2019-09-27 14:34:57.391')
,('THPT Nguyễn Trãi','COUNTRY_VN',4,71,NULL,true,'2019-09-27 14:34:57.392','2019-09-27 14:34:57.392')
,('THPT Tư thục Khai Trí','COUNTRY_VN',4,71,NULL,true,'2019-09-27 14:34:57.393','2019-09-27 14:34:57.393')
,('TTGDTX Quận Liên Chiểu','COUNTRY_VN',4,71,NULL,true,'2019-09-27 14:34:57.394','2019-09-27 14:34:57.394')
,('THPT Dân lập Hermann Gmeiner','COUNTRY_VN',4,72,NULL,true,'2019-09-27 14:34:57.396','2019-09-27 14:34:57.396')
,('THPT Ngũ Hành Sơn','COUNTRY_VN',4,72,NULL,true,'2019-09-27 14:34:57.398','2019-09-27 14:34:57.398')
,('TTGDTX Quận Ngũ Hành Sơn','COUNTRY_VN',4,72,NULL,true,'2019-09-27 14:34:57.399','2019-09-27 14:34:57.399')
,('THPT Chuyên Lê Quý Đôn','COUNTRY_VN',4,73,NULL,true,'2019-09-27 14:34:57.401','2019-09-27 14:34:57.401')
,('THPT Hoàng Hoa Thám','COUNTRY_VN',4,73,NULL,true,'2019-09-27 14:34:57.401','2019-09-27 14:34:57.401')
,('THPT Ngô Quyền','COUNTRY_VN',4,73,NULL,true,'2019-09-27 14:34:57.402','2019-09-27 14:34:57.402')
,('THPT Tôn Thất Tùng','COUNTRY_VN',4,73,NULL,true,'2019-09-27 14:34:57.402','2019-09-27 14:34:57.402')
,('TTGDTX Thành phố Đà Nẵng','COUNTRY_VN',4,73,NULL,true,'2019-09-27 14:34:57.402','2019-09-27 14:34:57.402')
,('THPT Thái Phiên','COUNTRY_VN',4,74,NULL,true,'2019-09-27 14:34:57.403','2019-09-27 14:34:57.403')
,('THPT Thanh Khê','COUNTRY_VN',4,74,NULL,true,'2019-09-27 14:34:57.404','2019-09-27 14:34:57.404')
,('THPT Tư thục Quang Trung','COUNTRY_VN',4,74,NULL,true,'2019-09-27 14:34:57.404','2019-09-27 14:34:57.404')
,('TTGDTX Quận Thanh Khê','COUNTRY_VN',4,74,NULL,true,'2019-09-27 14:34:57.404','2019-09-27 14:34:57.404')
,('THPT Ông Ích Khiêm','COUNTRY_VN',4,75,NULL,true,'2019-09-27 14:34:57.405','2019-09-27 14:34:57.405')
,('THPT Phạm Phú Thứ','COUNTRY_VN',4,75,NULL,true,'2019-09-27 14:34:57.406','2019-09-27 14:34:57.406')
,('THPT Phan Thành Tài','COUNTRY_VN',4,75,NULL,true,'2019-09-27 14:34:57.406','2019-09-27 14:34:57.406')
,('TTGDTX Huyện Hòa Vang','COUNTRY_VN',4,75,NULL,true,'2019-09-27 14:34:57.406','2019-09-27 14:34:57.406')
,('THCS-THPT Trần Ngọc Hoằng','COUNTRY_VN',5,76,NULL,true,'2019-09-27 14:34:57.411','2019-09-27 14:34:57.411')
,('THPT Hà Huy Giáo','COUNTRY_VN',5,76,NULL,true,'2019-09-27 14:34:57.411','2019-09-27 14:34:57.411')
,('THPT Trung An','COUNTRY_VN',5,76,NULL,true,'2019-09-27 14:34:57.412','2019-09-27 14:34:57.412')
,('TTGDTX Huyện Cờ Đỏ','COUNTRY_VN',5,76,NULL,true,'2019-09-27 14:34:57.413','2019-09-27 14:34:57.413')
,('THPT Giai Xuân','COUNTRY_VN',5,77,NULL,true,'2019-09-27 14:34:57.415','2019-09-27 14:34:57.415')
,('THPT Phan Văn Trị','COUNTRY_VN',5,77,NULL,true,'2019-09-27 14:34:57.416','2019-09-27 14:34:57.416')
,('TTGDTX Huyện Phong Điền','COUNTRY_VN',5,77,NULL,true,'2019-09-27 14:34:57.417','2019-09-27 14:34:57.417')
,('THCS-THPT Trường Xuân','COUNTRY_VN',5,78,NULL,true,'2019-09-27 14:34:57.418','2019-09-27 14:34:57.418')
,('THPT Thới Lai','COUNTRY_VN',5,78,NULL,true,'2019-09-27 14:34:57.418','2019-09-27 14:34:57.418')
,('TTGDTX Huyện Thới Lai','COUNTRY_VN',5,78,NULL,true,'2019-09-27 14:34:57.418','2019-09-27 14:34:57.418')
,('THPT Thốt Nốt','COUNTRY_VN',5,79,NULL,true,'2019-09-27 14:34:57.419','2019-09-27 14:34:57.419')
,('THPT Thuận Hưng','COUNTRY_VN',5,79,NULL,true,'2019-09-27 14:34:57.419','2019-09-27 14:34:57.419')
,('TTGDTX Huyện Thốt Nốt','COUNTRY_VN',5,79,NULL,true,'2019-09-27 14:34:57.420','2019-09-27 14:34:57.420')
,('THPT Thạnh An','COUNTRY_VN',5,80,NULL,true,'2019-09-27 14:34:57.420','2019-09-27 14:34:57.420')
,('THPT Vĩnh Thạnh','COUNTRY_VN',5,80,NULL,true,'2019-09-27 14:34:57.421','2019-09-27 14:34:57.421')
,('TTGDTX Huyện Vĩnh Thạnh','COUNTRY_VN',5,80,NULL,true,'2019-09-27 14:34:57.421','2019-09-27 14:34:57.421')
,('THPT Bình Thủy','COUNTRY_VN',5,81,NULL,true,'2019-09-27 14:34:57.422','2019-09-27 14:34:57.422')
,('THPT Bùi Hữu Nghĩa','COUNTRY_VN',5,81,NULL,true,'2019-09-27 14:34:57.422','2019-09-27 14:34:57.422')
,('THPT Chuyên Lý Tự Trọng','COUNTRY_VN',5,81,NULL,true,'2019-09-27 14:34:57.422','2019-09-27 14:34:57.422')
,('TTGDTX Quận Bình Thủy','COUNTRY_VN',5,81,NULL,true,'2019-09-27 14:34:57.423','2019-09-27 14:34:57.423')
,('THPT Nguyễn Việt Dũng','COUNTRY_VN',5,82,NULL,true,'2019-09-27 14:34:57.423','2019-09-27 14:34:57.423')
,('THPT Trần Đại Nghĩa','COUNTRY_VN',5,82,NULL,true,'2019-09-27 14:34:57.425','2019-09-27 14:34:57.425')
,('TTGDTX Quận Cái Răng','COUNTRY_VN',5,82,NULL,true,'2019-09-27 14:34:57.426','2019-09-27 14:34:57.426')
,('Phổ thông năng khiếu Thể dục Thể thao','COUNTRY_VN',5,83,NULL,true,'2019-09-27 14:34:57.427','2019-09-27 14:34:57.427')
,('THPT Việt Mỹ','COUNTRY_VN',5,83,NULL,true,'2019-09-27 14:34:57.428','2019-09-27 14:34:57.428')
,('TH-THCS-THPT Quốc văn Sài Gòn','COUNTRY_VN',5,83,NULL,true,'2019-09-27 14:34:57.428','2019-09-27 14:34:57.428')
,('THPT Châu Văn Liêm','COUNTRY_VN',5,83,NULL,true,'2019-09-27 14:34:57.429','2019-09-27 14:34:57.429')
,('THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',5,83,NULL,true,'2019-09-27 14:34:57.430','2019-09-27 14:34:57.430')
,('THPT Nguyễn Việt Hồng','COUNTRY_VN',5,83,NULL,true,'2019-09-27 14:34:57.430','2019-09-27 14:34:57.430')
,('THPT Phan Ngọc Hiển','COUNTRY_VN',5,83,NULL,true,'2019-09-27 14:34:57.431','2019-09-27 14:34:57.431')
,('THPT Thái Bình Dương','COUNTRY_VN',5,83,NULL,true,'2019-09-27 14:34:57.432','2019-09-27 14:34:57.432')
,('THPT Thực hành sư phạm - ĐHCT','COUNTRY_VN',5,83,NULL,true,'2019-09-27 14:34:57.432','2019-09-27 14:34:57.432')
,('TTGDTX Thành phố Cần Thơ','COUNTRY_VN',5,83,NULL,true,'2019-09-27 14:34:57.433','2019-09-27 14:34:57.433')
,('TTGDTX Quận Ninh Kiều','COUNTRY_VN',5,83,NULL,true,'2019-09-27 14:34:57.433','2019-09-27 14:34:57.433')
,('THPT Dân tộc Nội trú Ô Môn','COUNTRY_VN',5,84,NULL,true,'2019-09-27 14:34:57.434','2019-09-27 14:34:57.434')
,('THPT Thới Long','COUNTRY_VN',5,84,NULL,true,'2019-09-27 14:34:57.434','2019-09-27 14:34:57.434')
,('THPT Lưu Hữu Phước','COUNTRY_VN',5,84,NULL,true,'2019-09-27 14:34:57.435','2019-09-27 14:34:57.435')
,('THPT Lương Định Của','COUNTRY_VN',5,84,NULL,true,'2019-09-27 14:34:57.435','2019-09-27 14:34:57.435')
,('TTGDTX Quận Ô Môn','COUNTRY_VN',5,84,NULL,true,'2019-09-27 14:34:57.435','2019-09-27 14:34:57.435')
,('THPT Thốt Nốt','COUNTRY_VN',5,85,NULL,true,'2019-09-27 14:34:57.436','2019-09-27 14:34:57.436')
,('THPT Thuận Hưng','COUNTRY_VN',5,85,NULL,true,'2019-09-27 14:34:57.436','2019-09-27 14:34:57.436')
,('TTGDTX Quận Thốt Nốt','COUNTRY_VN',5,85,NULL,true,'2019-09-27 14:34:57.437','2019-09-27 14:34:57.437')
,('THPT Dân lập Lê Thánh Tôn','COUNTRY_VN',6,86,NULL,true,'2019-09-27 14:34:57.438','2019-09-27 14:34:57.438')
,('THPT Lê Trung Kiên','COUNTRY_VN',6,86,NULL,true,'2019-09-27 14:34:57.439','2019-09-27 14:34:57.439')
,('THPT Nguyễn Công Trứ','COUNTRY_VN',6,86,NULL,true,'2019-09-27 14:34:57.439','2019-09-27 14:34:57.439')
,('THPT Nguyễn Văn Linh','COUNTRY_VN',6,86,NULL,true,'2019-09-27 14:34:57.440','2019-09-27 14:34:57.440')
,('THCS-THPT Chu Văn An','COUNTRY_VN',6,87,NULL,true,'2019-09-27 14:34:57.442','2019-09-27 14:34:57.442')
,('THPT Lê Lợi','COUNTRY_VN',6,87,NULL,true,'2019-09-27 14:34:57.444','2019-09-27 14:34:57.444')
,('THPT Nguyễn Thái Bình','COUNTRY_VN',6,87,NULL,true,'2019-09-27 14:34:57.445','2019-09-27 14:34:57.445')
,('TTGDTX Huyện Đồng Xuân','COUNTRY_VN',6,87,NULL,true,'2019-09-27 14:34:57.446','2019-09-27 14:34:57.446')
,('THPT Trần Bình Trọng','COUNTRY_VN',6,88,NULL,true,'2019-09-27 14:34:57.446','2019-09-27 14:34:57.446')
,('THPT Trần Quốc Tuấn','COUNTRY_VN',6,88,NULL,true,'2019-09-27 14:34:57.447','2019-09-27 14:34:57.447')
,('THPT Trần Suyền','COUNTRY_VN',6,88,NULL,true,'2019-09-27 14:34:57.448','2019-09-27 14:34:57.448')
,('TTGDTX Huyện Phú Hòa','COUNTRY_VN',6,88,NULL,true,'2019-09-27 14:34:57.449','2019-09-27 14:34:57.449')
,('THCS-THPT Nguyễn Bá Ngọc','COUNTRY_VN',6,89,NULL,true,'2019-09-27 14:34:57.450','2019-09-27 14:34:57.450')
,('THPT Phan Bội Châu','COUNTRY_VN',6,89,NULL,true,'2019-09-27 14:34:57.451','2019-09-27 14:34:57.451')
,('TTGDTX Huyện Sơn Hòa','COUNTRY_VN',6,89,NULL,true,'2019-09-27 14:34:57.451','2019-09-27 14:34:57.451')
,('THPT Phan Chu Trinh','COUNTRY_VN',6,90,NULL,true,'2019-09-27 14:34:57.452','2019-09-27 14:34:57.452')
,('THCS-THPT Nguyễn Khuyến','COUNTRY_VN',6,90,NULL,true,'2019-09-27 14:34:57.452','2019-09-27 14:34:57.452')
,('THCS-THPT Võ Nguyên Giáp','COUNTRY_VN',6,90,NULL,true,'2019-09-27 14:34:57.453','2019-09-27 14:34:57.453')
,('THPT Phan Đình Phùng','COUNTRY_VN',6,90,NULL,true,'2019-09-27 14:34:57.453','2019-09-27 14:34:57.453')
,('TTGDTX Thị xã Sông Cầu','COUNTRY_VN',6,90,NULL,true,'2019-09-27 14:34:57.454','2019-09-27 14:34:57.454')
,('THCS-THPT Võ Văn Kiệt','COUNTRY_VN',6,91,NULL,true,'2019-09-27 14:34:57.454','2019-09-27 14:34:57.454')
,('THPT Nguyễn Du','COUNTRY_VN',6,91,NULL,true,'2019-09-27 14:34:57.455','2019-09-27 14:34:57.455')
,('THPT Tôn Đức Thắng','COUNTRY_VN',6,91,NULL,true,'2019-09-27 14:34:57.455','2019-09-27 14:34:57.455')
,('TTGDTX Huyện Sông Hinh','COUNTRY_VN',6,91,NULL,true,'2019-09-27 14:34:57.455','2019-09-27 14:34:57.455')
,('THPT Lê Hồng Phong','COUNTRY_VN',6,92,NULL,true,'2019-09-27 14:34:57.456','2019-09-27 14:34:57.456')
,('THPT Nguyễn Thị Minh Khai','COUNTRY_VN',6,92,NULL,true,'2019-09-27 14:34:57.456','2019-09-27 14:34:57.456')
,('THPT Phạm Văn Đồng','COUNTRY_VN',6,92,NULL,true,'2019-09-27 14:34:57.457','2019-09-27 14:34:57.457')
,('THCS-THPT Nguyễn Viết Xuân','COUNTRY_VN',6,93,NULL,true,'2019-09-27 14:34:57.460','2019-09-27 14:34:57.460')
,('THCS-THPT Võ Thị Sáu','COUNTRY_VN',6,93,NULL,true,'2019-09-27 14:34:57.461','2019-09-27 14:34:57.461')
,('THPT Lê Thành Phương','COUNTRY_VN',6,93,NULL,true,'2019-09-27 14:34:57.462','2019-09-27 14:34:57.462')
,('THPT Trần Phú','COUNTRY_VN',6,93,NULL,true,'2019-09-27 14:34:57.463','2019-09-27 14:34:57.463')
,('TTGDTX Huyện Tuy An','COUNTRY_VN',6,93,NULL,true,'2019-09-27 14:34:57.464','2019-09-27 14:34:57.464')
,('THPT Tư thục Duy Tân','COUNTRY_VN',6,94,NULL,true,'2019-09-27 14:34:57.466','2019-09-27 14:34:57.466')
,('THPT Chuyên Lương Văn Chánh','COUNTRY_VN',6,94,NULL,true,'2019-09-27 14:34:57.467','2019-09-27 14:34:57.467')
,('THPT Dân tộc Nội trú Phú Yên','COUNTRY_VN',6,94,NULL,true,'2019-09-27 14:34:57.468','2019-09-27 14:34:57.468')
,('THPT Dân lập Nguyễn Bỉnh Khiêm','COUNTRY_VN',6,94,NULL,true,'2019-09-27 14:34:57.468','2019-09-27 14:34:57.468')
,('THPT Ngô Gia Tự','COUNTRY_VN',6,94,NULL,true,'2019-09-27 14:34:57.469','2019-09-27 14:34:57.469')
,('THPT Nguyễn Huệ','COUNTRY_VN',6,94,NULL,true,'2019-09-27 14:34:57.469','2019-09-27 14:34:57.469')
,('THPT Nguyễn Trãi','COUNTRY_VN',6,94,NULL,true,'2019-09-27 14:34:57.470','2019-09-27 14:34:57.470')
,('THPT Nguyễn Trường Tộ','COUNTRY_VN',6,94,NULL,true,'2019-09-27 14:34:57.470','2019-09-27 14:34:57.470')
,('TTGDTX Tỉnh Phú Yên','COUNTRY_VN',6,94,NULL,true,'2019-09-27 14:34:57.471','2019-09-27 14:34:57.471')
,('TTGDTX Thành phố Tuy Hòa','COUNTRY_VN',6,94,NULL,true,'2019-09-27 14:34:57.471','2019-09-27 14:34:57.471')
,('THPT Hoàng Văn Thụ','COUNTRY_VN',7,95,NULL,true,'2019-09-27 14:34:57.473','2019-09-27 14:34:57.473')
,('THPT Hồng Quang','COUNTRY_VN',7,95,NULL,true,'2019-09-27 14:34:57.473','2019-09-27 14:34:57.473')
,('THPT Mai Sơn','COUNTRY_VN',7,95,NULL,true,'2019-09-27 14:34:57.474','2019-09-27 14:34:57.474')
,('TTGDTX Hồ Tùng Mậu','COUNTRY_VN',7,95,NULL,true,'2019-09-27 14:34:57.475','2019-09-27 14:34:57.475')
,('THPT Mù Cang Chải','COUNTRY_VN',7,96,NULL,true,'2019-09-27 14:34:57.477','2019-09-27 14:34:57.477')
,('TTGDTX Huyện Mù Cang Chải','COUNTRY_VN',7,96,NULL,true,'2019-09-27 14:34:57.477','2019-09-27 14:34:57.477')
,('THPT Trạm Tấu','COUNTRY_VN',7,97,NULL,true,'2019-09-27 14:34:57.478','2019-09-27 14:34:57.478')
,('TTGDTX Huyện Trạm Tấu','COUNTRY_VN',7,97,NULL,true,'2019-09-27 14:34:57.479','2019-09-27 14:34:57.479')
,('THCS-THPT Trấn Yên II','COUNTRY_VN',7,98,NULL,true,'2019-09-27 14:34:57.481','2019-09-27 14:34:57.481')
,('THPT Lê Quý Đôn','COUNTRY_VN',7,98,NULL,true,'2019-09-27 14:34:57.482','2019-09-27 14:34:57.482')
,('TTGDTX Huyện Trấn Yên','COUNTRY_VN',7,98,NULL,true,'2019-09-27 14:34:57.482','2019-09-27 14:34:57.482')
,('THPT Sơn Thịnh','COUNTRY_VN',7,99,NULL,true,'2019-09-27 14:34:57.483','2019-09-27 14:34:57.483')
,('THPT Văn Chấn','COUNTRY_VN',7,99,NULL,true,'2019-09-27 14:34:57.484','2019-09-27 14:34:57.484')
,('TTGDTX Huyện Văn Chấn','COUNTRY_VN',7,99,NULL,true,'2019-09-27 14:34:57.484','2019-09-27 14:34:57.484')
,('THPT Chu Văn An','COUNTRY_VN',7,100,NULL,true,'2019-09-27 14:34:57.485','2019-09-27 14:34:57.485')
,('THPT Nguyễn Lương Bằng','COUNTRY_VN',7,100,NULL,true,'2019-09-27 14:34:57.486','2019-09-27 14:34:57.486')
,('THPT Trần Phú','COUNTRY_VN',7,100,NULL,true,'2019-09-27 14:34:57.486','2019-09-27 14:34:57.486')
,('TTGDTX Huyện Văn Yên','COUNTRY_VN',7,100,NULL,true,'2019-09-27 14:34:57.486','2019-09-27 14:34:57.486')
,('THPT Cảm Ân','COUNTRY_VN',7,101,NULL,true,'2019-09-27 14:34:57.487','2019-09-27 14:34:57.487')
,('THPT Cảm Nhân','COUNTRY_VN',7,101,NULL,true,'2019-09-27 14:34:57.487','2019-09-27 14:34:57.487')
,('THPT Thác Bà','COUNTRY_VN',7,101,NULL,true,'2019-09-27 14:34:57.488','2019-09-27 14:34:57.488')
,('THPT Trần Nhật Duật','COUNTRY_VN',7,101,NULL,true,'2019-09-27 14:34:57.488','2019-09-27 14:34:57.488')
,('TTGDTX Huyện Yên Bình','COUNTRY_VN',7,101,NULL,true,'2019-09-27 14:34:57.488','2019-09-27 14:34:57.488')
,('THPT Chuyên Nguyễn Tất Thành','COUNTRY_VN',7,102,NULL,true,'2019-09-27 14:34:57.490','2019-09-27 14:34:57.490')
,('THPT Dân tộc Nội trú Yên Bái','COUNTRY_VN',7,102,NULL,true,'2019-09-27 14:34:57.490','2019-09-27 14:34:57.490')
,('THPT Đồng Tâm','COUNTRY_VN',7,102,NULL,true,'2019-09-27 14:34:57.491','2019-09-27 14:34:57.491')
,('THPT Hoàng Quốc Việt','COUNTRY_VN',7,102,NULL,true,'2019-09-27 14:34:57.492','2019-09-27 14:34:57.492')
,('THPT Lý Thường Kiệt','COUNTRY_VN',7,102,NULL,true,'2019-09-27 14:34:57.492','2019-09-27 14:34:57.492')
,('THPT Nguyễn Huệ','COUNTRY_VN',7,102,NULL,true,'2019-09-27 14:34:57.493','2019-09-27 14:34:57.493')
,('TTGDTX Thành phố Yên Bái','COUNTRY_VN',7,102,NULL,true,'2019-09-27 14:34:57.493','2019-09-27 14:34:57.493')
,('TTGDTX Tỉnh Yên Bái','COUNTRY_VN',7,102,NULL,true,'2019-09-27 14:34:57.494','2019-09-27 14:34:57.494')
,('THPT Dân tộc Nội trú Miền Tây','COUNTRY_VN',7,103,NULL,true,'2019-09-27 14:34:57.495','2019-09-27 14:34:57.495')
,('THPT Nghĩa Lộ','COUNTRY_VN',7,103,NULL,true,'2019-09-27 14:34:57.496','2019-09-27 14:34:57.496')
,('THPT Nguyễn Trãi','COUNTRY_VN',7,103,NULL,true,'2019-09-27 14:34:57.497','2019-09-27 14:34:57.497')
,('TTGDTX Thị xã Nghĩa Lộ','COUNTRY_VN',7,103,NULL,true,'2019-09-27 14:34:57.498','2019-09-27 14:34:57.498')
,('THPT Bình Xuyên','COUNTRY_VN',8,104,NULL,true,'2019-09-27 14:34:57.499','2019-09-27 14:34:57.499')
,('THPT Nguyễn Duy Thì','COUNTRY_VN',8,104,NULL,true,'2019-09-27 14:34:57.500','2019-09-27 14:34:57.500')
,('THPT Quang Hà','COUNTRY_VN',8,104,NULL,true,'2019-09-27 14:34:57.500','2019-09-27 14:34:57.500')
,('THPT Võ Thị Sáu','COUNTRY_VN',8,104,NULL,true,'2019-09-27 14:34:57.500','2019-09-27 14:34:57.500')
,('TTGDTX Huyện Bình Xuyên','COUNTRY_VN',8,104,NULL,true,'2019-09-27 14:34:57.501','2019-09-27 14:34:57.501')
,('THPT Liễn Sơn','COUNTRY_VN',8,105,NULL,true,'2019-09-27 14:34:57.502','2019-09-27 14:34:57.502')
,('THPT Ngô Gia Tự','COUNTRY_VN',8,105,NULL,true,'2019-09-27 14:34:57.502','2019-09-27 14:34:57.502')
,('THPT Thái Hòa','COUNTRY_VN',8,105,NULL,true,'2019-09-27 14:34:57.503','2019-09-27 14:34:57.503')
,('THPT Trần Nguyễn Hãn','COUNTRY_VN',8,105,NULL,true,'2019-09-27 14:34:57.503','2019-09-27 14:34:57.503')
,('THPT Triệu Thái','COUNTRY_VN',8,105,NULL,true,'2019-09-27 14:34:57.503','2019-09-27 14:34:57.503')
,('THPT Văn Quán','COUNTRY_VN',8,105,NULL,true,'2019-09-27 14:34:57.504','2019-09-27 14:34:57.504')
,('TTGDTX Huyện Lập Thạch','COUNTRY_VN',8,105,NULL,true,'2019-09-27 14:34:57.504','2019-09-27 14:34:57.504')
,('THPT Bình Sơn','COUNTRY_VN',8,106,NULL,true,'2019-09-27 14:34:57.505','2019-09-27 14:34:57.505')
,('THPT Sáng Sơn','COUNTRY_VN',8,106,NULL,true,'2019-09-27 14:34:57.506','2019-09-27 14:34:57.506')
,('THPT Sông Lô','COUNTRY_VN',8,106,NULL,true,'2019-09-27 14:34:57.506','2019-09-27 14:34:57.506')
,('THPT Tam Dương','COUNTRY_VN',8,107,NULL,true,'2019-09-27 14:34:57.507','2019-09-27 14:34:57.507')
,('THPT Tam Dương 2','COUNTRY_VN',8,107,NULL,true,'2019-09-27 14:34:57.508','2019-09-27 14:34:57.508')
,('THPT Trần Hưng Đạo','COUNTRY_VN',8,107,NULL,true,'2019-09-27 14:34:57.509','2019-09-27 14:34:57.509')
,('TTGDTX Huyện Tam Dương','COUNTRY_VN',8,107,NULL,true,'2019-09-27 14:34:57.510','2019-09-27 14:34:57.510')
,('THPT Tam Đảo','COUNTRY_VN',8,108,NULL,true,'2019-09-27 14:34:57.511','2019-09-27 14:34:57.511')
,('THPT Tam Đảo 2','COUNTRY_VN',8,108,NULL,true,'2019-09-27 14:34:57.511','2019-09-27 14:34:57.511')
,('TTGDTX Huyện Tam Đảo','COUNTRY_VN',8,108,NULL,true,'2019-09-27 14:34:57.511','2019-09-27 14:34:57.511')
,('THPT Đội Cấn','COUNTRY_VN',8,109,NULL,true,'2019-09-27 14:34:57.513','2019-09-27 14:34:57.513')
,('THPT Hồ Xuân Hương','COUNTRY_VN',8,109,NULL,true,'2019-09-27 14:34:57.514','2019-09-27 14:34:57.514')
,('THPT Lê Xoay','COUNTRY_VN',8,109,NULL,true,'2019-09-27 14:34:57.515','2019-09-27 14:34:57.515')
,('THPT Nguyễn Thị Giang','COUNTRY_VN',8,109,NULL,true,'2019-09-27 14:34:57.515','2019-09-27 14:34:57.515')
,('THPT Nguyễn Viết Xuân','COUNTRY_VN',8,109,NULL,true,'2019-09-27 14:34:57.516','2019-09-27 14:34:57.516')
,('THPT Vĩnh Tường','COUNTRY_VN',8,109,NULL,true,'2019-09-27 14:34:57.516','2019-09-27 14:34:57.516')
,('TTGDTX Huyện Vĩnh Tường','COUNTRY_VN',8,109,NULL,true,'2019-09-27 14:34:57.517','2019-09-27 14:34:57.517')
,('THPT Đồng Đậu','COUNTRY_VN',8,110,NULL,true,'2019-09-27 14:34:57.518','2019-09-27 14:34:57.518')
,('THPT Phạm Công Bình','COUNTRY_VN',8,110,NULL,true,'2019-09-27 14:34:57.518','2019-09-27 14:34:57.518')
,('THPT Yên Lạc','COUNTRY_VN',8,110,NULL,true,'2019-09-27 14:34:57.518','2019-09-27 14:34:57.518')
,('THPT Yên Lạc 2','COUNTRY_VN',8,110,NULL,true,'2019-09-27 14:34:57.519','2019-09-27 14:34:57.519')
,('TTGDTX Huyện Yên Lạc','COUNTRY_VN',8,110,NULL,true,'2019-09-27 14:34:57.519','2019-09-27 14:34:57.519')
,('THPT Chuyên Vĩnh Phúc','COUNTRY_VN',8,111,NULL,true,'2019-09-27 14:34:57.520','2019-09-27 14:34:57.520')
,('THPT Dân tộc Nội trú Vĩnh Phúc','COUNTRY_VN',8,111,NULL,true,'2019-09-27 14:34:57.520','2019-09-27 14:34:57.520')
,('THPT Liên Bảo','COUNTRY_VN',8,111,NULL,true,'2019-09-27 14:34:57.520','2019-09-27 14:34:57.520')
,('THPT Nguyễn Thái Học','COUNTRY_VN',8,111,NULL,true,'2019-09-27 14:34:57.521','2019-09-27 14:34:57.521')
,('THPT Trần Phú','COUNTRY_VN',8,111,NULL,true,'2019-09-27 14:34:57.521','2019-09-27 14:34:57.521')
,('THPT Vĩnh Yên','COUNTRY_VN',8,111,NULL,true,'2019-09-27 14:34:57.521','2019-09-27 14:34:57.521')
,('TTGDTX Tỉnh Vĩnh Phúc','COUNTRY_VN',8,111,NULL,true,'2019-09-27 14:34:57.522','2019-09-27 14:34:57.522')
,('THPT Bến Tre','COUNTRY_VN',8,112,NULL,true,'2019-09-27 14:34:57.522','2019-09-27 14:34:57.522')
,('THPT Hai Bà Trưng','COUNTRY_VN',8,112,NULL,true,'2019-09-27 14:34:57.523','2019-09-27 14:34:57.523')
,('THPT Phúc Yên','COUNTRY_VN',8,112,NULL,true,'2019-09-27 14:34:57.523','2019-09-27 14:34:57.523')
,('THPT Xuân Hòa','COUNTRY_VN',8,112,NULL,true,'2019-09-27 14:34:57.524','2019-09-27 14:34:57.524')
,('TTGDTX Thị xã Phúc Yên','COUNTRY_VN',8,112,NULL,true,'2019-09-27 14:34:57.525','2019-09-27 14:34:57.525')
,('THPT Bình Minh','COUNTRY_VN',9,113,NULL,true,'2019-09-27 14:34:57.527','2019-09-27 14:34:57.527')
,('THPT Hoàng Thái Hiếu','COUNTRY_VN',9,113,NULL,true,'2019-09-27 14:34:57.528','2019-09-27 14:34:57.528')
,('TTGDTX Huyện Bình Minh','COUNTRY_VN',9,113,NULL,true,'2019-09-27 14:34:57.528','2019-09-27 14:34:57.528')
,('THCS-THPT Mỹ Thuận','COUNTRY_VN',9,114,NULL,true,'2019-09-27 14:34:57.530','2019-09-27 14:34:57.530')
,('THPT Tân Lược','COUNTRY_VN',9,114,NULL,true,'2019-09-27 14:34:57.530','2019-09-27 14:34:57.530')
,('THPT Tân Quới','COUNTRY_VN',9,114,NULL,true,'2019-09-27 14:34:57.531','2019-09-27 14:34:57.531')
,('TTGDTX Huyện Bình Tân','COUNTRY_VN',9,114,NULL,true,'2019-09-27 14:34:57.532','2019-09-27 14:34:57.532')
,('THCS-THPT Phú Quới','COUNTRY_VN',9,115,NULL,true,'2019-09-27 14:34:57.533','2019-09-27 14:34:57.533')
,('THPT Bán công Long Hồ','COUNTRY_VN',9,115,NULL,true,'2019-09-27 14:34:57.534','2019-09-27 14:34:57.534')
,('THPT Hòa Ninh','COUNTRY_VN',9,115,NULL,true,'2019-09-27 14:34:57.534','2019-09-27 14:34:57.534')
,('THPT Phạm Hùng','COUNTRY_VN',9,115,NULL,true,'2019-09-27 14:34:57.534','2019-09-27 14:34:57.534')
,('TTGDTX Huyện Long Hồ','COUNTRY_VN',9,115,NULL,true,'2019-09-27 14:34:57.535','2019-09-27 14:34:57.535')
,('THCS-THPT Mỹ Phước','COUNTRY_VN',9,116,NULL,true,'2019-09-27 14:34:57.536','2019-09-27 14:34:57.536')
,('THPT Mang Thít','COUNTRY_VN',9,116,NULL,true,'2019-09-27 14:34:57.536','2019-09-27 14:34:57.536')
,('THPT Nguyễn Văn Thiệt','COUNTRY_VN',9,116,NULL,true,'2019-09-27 14:34:57.536','2019-09-27 14:34:57.536')
,('TTGDTX Huyện Mang Thít','COUNTRY_VN',9,116,NULL,true,'2019-09-27 14:34:57.537','2019-09-27 14:34:57.537')
,('THCS-THPT Long Phú','COUNTRY_VN',9,117,NULL,true,'2019-09-27 14:34:57.538','2019-09-27 14:34:57.538')
,('THCS-THPT Phú Thịnh','COUNTRY_VN',9,117,NULL,true,'2019-09-27 14:34:57.538','2019-09-27 14:34:57.538')
,('THPT Dân tộc Nội trú Vĩnh Long','COUNTRY_VN',9,117,NULL,true,'2019-09-27 14:34:57.539','2019-09-27 14:34:57.539')
,('THPT Phan Văn Hòa','COUNTRY_VN',9,117,NULL,true,'2019-09-27 14:34:57.539','2019-09-27 14:34:57.539')
,('THPT Tam Bình','COUNTRY_VN',9,117,NULL,true,'2019-09-27 14:34:57.540','2019-09-27 14:34:57.540')
,('THPT Trần Đại Nghịa','COUNTRY_VN',9,117,NULL,true,'2019-09-27 14:34:57.540','2019-09-27 14:34:57.540')
,('TTGDTX Huyện Tam Bình','COUNTRY_VN',9,117,NULL,true,'2019-09-27 14:34:57.540','2019-09-27 14:34:57.540')
,('THCS-THPT Hòa Bình','COUNTRY_VN',9,118,NULL,true,'2019-09-27 14:34:57.543','2019-09-27 14:34:57.543')
,('THPT Hựu Thành','COUNTRY_VN',9,118,NULL,true,'2019-09-27 14:34:57.544','2019-09-27 14:34:57.544')
,('THPT Lê Thanh Mừng','COUNTRY_VN',9,118,NULL,true,'2019-09-27 14:34:57.544','2019-09-27 14:34:57.544')
,('THPT Trà Ôn','COUNTRY_VN',9,118,NULL,true,'2019-09-27 14:34:57.546','2019-09-27 14:34:57.546')
,('THPT Vĩnh Xuân','COUNTRY_VN',9,118,NULL,true,'2019-09-27 14:34:57.547','2019-09-27 14:34:57.547')
,('TTGDTX Huyện Trà Ôn','COUNTRY_VN',9,118,NULL,true,'2019-09-27 14:34:57.548','2019-09-27 14:34:57.548')
,('THCS-THPT Hiếu Nhơn','COUNTRY_VN',9,119,NULL,true,'2019-09-27 14:34:57.549','2019-09-27 14:34:57.549')
,('THPT Hiếu Phụng','COUNTRY_VN',9,119,NULL,true,'2019-09-27 14:34:57.550','2019-09-27 14:34:57.550')
,('THPT Nguyễn Hiếu Tự','COUNTRY_VN',9,119,NULL,true,'2019-09-27 14:34:57.551','2019-09-27 14:34:57.551')
,('THPT Võ Văn Kiệt','COUNTRY_VN',9,119,NULL,true,'2019-09-27 14:34:57.551','2019-09-27 14:34:57.551')
,('TTGDTX Huyện Vũng Liêm','COUNTRY_VN',9,119,NULL,true,'2019-09-27 14:34:57.552','2019-09-27 14:34:57.552')
,('THCS-THPT Trưng Vương','COUNTRY_VN',9,120,NULL,true,'2019-09-27 14:34:57.553','2019-09-27 14:34:57.553')
,('THPT Chuyên Nguyễn Bỉnh Khiêm','COUNTRY_VN',9,120,NULL,true,'2019-09-27 14:34:57.553','2019-09-27 14:34:57.553')
,('THPT Lưu Văn Liệt','COUNTRY_VN',9,120,NULL,true,'2019-09-27 14:34:57.554','2019-09-27 14:34:57.554')
,('THPT Nguyễn Thông','COUNTRY_VN',9,120,NULL,true,'2019-09-27 14:34:57.554','2019-09-27 14:34:57.554')
,('THPT Vình Long','COUNTRY_VN',9,120,NULL,true,'2019-09-27 14:34:57.554','2019-09-27 14:34:57.554')
,('TTGDTX Thành phố Vĩnh Long','COUNTRY_VN',9,120,NULL,true,'2019-09-27 14:34:57.555','2019-09-27 14:34:57.555')
,('THPT Bình Minh','COUNTRY_VN',9,121,NULL,true,'2019-09-27 14:34:57.555','2019-09-27 14:34:57.555')
,('THPT Hoàng Thái Hiếu','COUNTRY_VN',9,121,NULL,true,'2019-09-27 14:34:57.556','2019-09-27 14:34:57.556')
,('TTGDTX Thị xã Bình Minh','COUNTRY_VN',9,121,NULL,true,'2019-09-27 14:34:57.556','2019-09-27 14:34:57.556')
,('THPT Chiêm Hóa','COUNTRY_VN',10,122,NULL,true,'2019-09-27 14:34:57.559','2019-09-27 14:34:57.559')
,('THPT Đầm Hồng','COUNTRY_VN',10,122,NULL,true,'2019-09-27 14:34:57.561','2019-09-27 14:34:57.561')
,('THPT Hà Lang','COUNTRY_VN',10,122,NULL,true,'2019-09-27 14:34:57.561','2019-09-27 14:34:57.561')
,('THPT Hòa Phú','COUNTRY_VN',10,122,NULL,true,'2019-09-27 14:34:57.562','2019-09-27 14:34:57.562')
,('THPT Kim Bình','COUNTRY_VN',10,122,NULL,true,'2019-09-27 14:34:57.562','2019-09-27 14:34:57.562')
,('THPT Minh Quang','COUNTRY_VN',10,122,NULL,true,'2019-09-27 14:34:57.563','2019-09-27 14:34:57.563')
,('THPT Hàm Yên','COUNTRY_VN',10,123,NULL,true,'2019-09-27 14:34:57.565','2019-09-27 14:34:57.565')
,('THPT Phù Lưu','COUNTRY_VN',10,123,NULL,true,'2019-09-27 14:34:57.566','2019-09-27 14:34:57.566')
,('THPT Thái Hòa','COUNTRY_VN',10,123,NULL,true,'2019-09-27 14:34:57.566','2019-09-27 14:34:57.566')
,('THPT Lâm Bình','COUNTRY_VN',10,124,NULL,true,'2019-09-27 14:34:57.567','2019-09-27 14:34:57.567')
,('THPT Thượng Lâm','COUNTRY_VN',10,124,NULL,true,'2019-09-27 14:34:57.567','2019-09-27 14:34:57.567')
,('THPT Na Nang','COUNTRY_VN',10,125,NULL,true,'2019-09-27 14:34:57.568','2019-09-27 14:34:57.568')
,('THPT Yên Hoa','COUNTRY_VN',10,125,NULL,true,'2019-09-27 14:34:57.569','2019-09-27 14:34:57.569')
,('THPT ATK Tân Trào','COUNTRY_VN',10,126,NULL,true,'2019-09-27 14:34:57.570','2019-09-27 14:34:57.570')
,('THPT Đông Thọ','COUNTRY_VN',10,126,NULL,true,'2019-09-27 14:34:57.570','2019-09-27 14:34:57.570')
,('THPT Kháng Nhật','COUNTRY_VN',10,126,NULL,true,'2019-09-27 14:34:57.570','2019-09-27 14:34:57.570')
,('THPT Kim Xuyên','COUNTRY_VN',10,126,NULL,true,'2019-09-27 14:34:57.571','2019-09-27 14:34:57.571')
,('THPT Sơn Dương','COUNTRY_VN',10,126,NULL,true,'2019-09-27 14:34:57.572','2019-09-27 14:34:57.572')
,('THPT Sơn Nam','COUNTRY_VN',10,126,NULL,true,'2019-09-27 14:34:57.572','2019-09-27 14:34:57.572')
,('THPT Tháng 10','COUNTRY_VN',10,127,NULL,true,'2019-09-27 14:34:57.574','2019-09-27 14:34:57.574')
,('THPT Trung Sơn','COUNTRY_VN',10,127,NULL,true,'2019-09-27 14:34:57.575','2019-09-27 14:34:57.575')
,('THPT Xuân Huy','COUNTRY_VN',10,127,NULL,true,'2019-09-27 14:34:57.576','2019-09-27 14:34:57.576')
,('THPT Xuân Vân','COUNTRY_VN',10,127,NULL,true,'2019-09-27 14:34:57.577','2019-09-27 14:34:57.577')
,('THPT Dân tộc Nội trú Tuyên Quang','COUNTRY_VN',10,128,NULL,true,'2019-09-27 14:34:57.578','2019-09-27 14:34:57.578')
,('THPT Chuyên Tuyên Quang','COUNTRY_VN',10,128,NULL,true,'2019-09-27 14:34:57.578','2019-09-27 14:34:57.578')
,('THPT Nguyễn Văn Huyên','COUNTRY_VN',10,128,NULL,true,'2019-09-27 14:34:57.578','2019-09-27 14:34:57.578')
,('THPT Sông Lô','COUNTRY_VN',10,128,NULL,true,'2019-09-27 14:34:57.579','2019-09-27 14:34:57.579')
,('THPT Tân Trào','COUNTRY_VN',10,128,NULL,true,'2019-09-27 14:34:57.581','2019-09-27 14:34:57.581')
,('THPT Ỷ La','COUNTRY_VN',10,128,NULL,true,'2019-09-27 14:34:57.582','2019-09-27 14:34:57.582')
,('TTGDTX tỉnh Tuyên Quang','COUNTRY_VN',10,128,NULL,true,'2019-09-27 14:34:57.582','2019-09-27 14:34:57.582')
,('THPT Bùi Hữu Nghĩa','COUNTRY_VN',11,129,NULL,true,'2019-09-27 14:34:57.584','2019-09-27 14:34:57.584')
,('THPT Dương Háo Học','COUNTRY_VN',11,129,NULL,true,'2019-09-27 14:34:57.585','2019-09-27 14:34:57.585')
,('THPT Hồ Thị Nhâm','COUNTRY_VN',11,129,NULL,true,'2019-09-27 14:34:57.585','2019-09-27 14:34:57.585')
,('THPT Nguyễn Đáng','COUNTRY_VN',11,129,NULL,true,'2019-09-27 14:34:57.586','2019-09-27 14:34:57.586')
,('THPT Nguyễn Văn Hai','COUNTRY_VN',11,129,NULL,true,'2019-09-27 14:34:57.586','2019-09-27 14:34:57.586')
,('TTGDTX Huyện Càng Long','COUNTRY_VN',11,129,NULL,true,'2019-09-27 14:34:57.586','2019-09-27 14:34:57.586')
,('THPT Cầu Kè','COUNTRY_VN',11,130,NULL,true,'2019-09-27 14:34:57.587','2019-09-27 14:34:57.587')
,('THPT Phong Phú','COUNTRY_VN',11,130,NULL,true,'2019-09-27 14:34:57.588','2019-09-27 14:34:57.588')
,('THPT Tam Ngãi','COUNTRY_VN',11,130,NULL,true,'2019-09-27 14:34:57.588','2019-09-27 14:34:57.588')
,('TTGDTX Huyện Cầu Kè','COUNTRY_VN',11,130,NULL,true,'2019-09-27 14:34:57.588','2019-09-27 14:34:57.588')
,('THPT Cầu Ngang A','COUNTRY_VN',11,131,NULL,true,'2019-09-27 14:34:57.589','2019-09-27 14:34:57.589')
,('THPT Cầu Ngang B','COUNTRY_VN',11,131,NULL,true,'2019-09-27 14:34:57.590','2019-09-27 14:34:57.590')
,('THPT Dương Quang Đông','COUNTRY_VN',11,131,NULL,true,'2019-09-27 14:34:57.590','2019-09-27 14:34:57.590')
,('THPT Nhị Trường','COUNTRY_VN',11,131,NULL,true,'2019-09-27 14:34:57.592','2019-09-27 14:34:57.592')
,('TTGDTX Huyện Cầu Ngang','COUNTRY_VN',11,131,NULL,true,'2019-09-27 14:34:57.594','2019-09-27 14:34:57.594')
,('THPT Hòa Lợi','COUNTRY_VN',11,132,NULL,true,'2019-09-27 14:34:57.595','2019-09-27 14:34:57.595')
,('THPT Hòa Minh','COUNTRY_VN',11,132,NULL,true,'2019-09-27 14:34:57.597','2019-09-27 14:34:57.597')
,('THPT Lương Hòa A','COUNTRY_VN',11,132,NULL,true,'2019-09-27 14:34:57.598','2019-09-27 14:34:57.598')
,('THPT Vũ Đình Liệu','COUNTRY_VN',11,132,NULL,true,'2019-09-27 14:34:57.599','2019-09-27 14:34:57.599')
,('TTGDTX Huyện Châu Thành','COUNTRY_VN',11,132,NULL,true,'2019-09-27 14:34:57.600','2019-09-27 14:34:57.600')
,('THPT Duyên Hải','COUNTRY_VN',11,133,NULL,true,'2019-09-27 14:34:57.601','2019-09-27 14:34:57.601')
,('THPT Long Hữu','COUNTRY_VN',11,133,NULL,true,'2019-09-27 14:34:57.601','2019-09-27 14:34:57.601')
,('THPT Long Khánh','COUNTRY_VN',11,133,NULL,true,'2019-09-27 14:34:57.602','2019-09-27 14:34:57.602')
,('TTGDTX Huyện Duyên Hải','COUNTRY_VN',11,133,NULL,true,'2019-09-27 14:34:57.602','2019-09-27 14:34:57.602')
,('THPT Cầu Quan','COUNTRY_VN',11,134,NULL,true,'2019-09-27 14:34:57.603','2019-09-27 14:34:57.603')
,('THPT Hiếu Tử','COUNTRY_VN',11,134,NULL,true,'2019-09-27 14:34:57.603','2019-09-27 14:34:57.603')
,('THPT Tiểu Cần','COUNTRY_VN',11,134,NULL,true,'2019-09-27 14:34:57.604','2019-09-27 14:34:57.604')
,('TTGDTX Huyện Tiểu Cần','COUNTRY_VN',11,134,NULL,true,'2019-09-27 14:34:57.604','2019-09-27 14:34:57.604')
,('THCS-THPT Dân tộc Nội trú Trà Cú','COUNTRY_VN',11,135,NULL,true,'2019-09-27 14:34:57.605','2019-09-27 14:34:57.605')
,('THPT Đại An','COUNTRY_VN',11,135,NULL,true,'2019-09-27 14:34:57.605','2019-09-27 14:34:57.605')
,('THPT Đôn Châu','COUNTRY_VN',11,135,NULL,true,'2019-09-27 14:34:57.605','2019-09-27 14:34:57.605')
,('THPT Hàm Giang','COUNTRY_VN',11,135,NULL,true,'2019-09-27 14:34:57.606','2019-09-27 14:34:57.606')
,('THPT Long Hiệp','COUNTRY_VN',11,135,NULL,true,'2019-09-27 14:34:57.606','2019-09-27 14:34:57.606')
,('THPT Tập Sơn','COUNTRY_VN',11,135,NULL,true,'2019-09-27 14:34:57.607','2019-09-27 14:34:57.607')
,('THPT Trà Cú','COUNTRY_VN',11,135,NULL,true,'2019-09-27 14:34:57.607','2019-09-27 14:34:57.607')
,('TTGDTX Huyện Trà Cú','COUNTRY_VN',11,135,NULL,true,'2019-09-27 14:34:57.610','2019-09-27 14:34:57.610')
,('THPT Dân tộc Nội trú Trà Vinh','COUNTRY_VN',11,136,NULL,true,'2019-09-27 14:34:57.611','2019-09-27 14:34:57.611')
,('THPT Chuyên Nguyễn Thiện Thành','COUNTRY_VN',11,136,NULL,true,'2019-09-27 14:34:57.612','2019-09-27 14:34:57.612')
,('THPT Phạm Thái Bường','COUNTRY_VN',11,136,NULL,true,'2019-09-27 14:34:57.612','2019-09-27 14:34:57.612')
,('THPT Thành phố Trà Vinh','COUNTRY_VN',11,136,NULL,true,'2019-09-27 14:34:57.613','2019-09-27 14:34:57.613')
,('THPT Thực hành sư phạm - ĐHCT','COUNTRY_VN',11,136,NULL,true,'2019-09-27 14:34:57.614','2019-09-27 14:34:57.614')
,('TTGDTX Thành phố Trà Vinh','COUNTRY_VN',11,136,NULL,true,'2019-09-27 14:34:57.614','2019-09-27 14:34:57.614')
,('THPT Cái Bè','COUNTRY_VN',12,137,NULL,true,'2019-09-27 14:34:57.616','2019-09-27 14:34:57.616')
,('THPT Huỳnh văn Sâm','COUNTRY_VN',12,137,NULL,true,'2019-09-27 14:34:57.617','2019-09-27 14:34:57.617')
,('THPT Lê Thanh Hiền','COUNTRY_VN',12,137,NULL,true,'2019-09-27 14:34:57.617','2019-09-27 14:34:57.617')
,('THPT Ngô Văn Nhạc','COUNTRY_VN',12,137,NULL,true,'2019-09-27 14:34:57.617','2019-09-27 14:34:57.617')
,('THPT Phạm Thành Trung','COUNTRY_VN',12,137,NULL,true,'2019-09-27 14:34:57.618','2019-09-27 14:34:57.618')
,('THPT Thiên Hộ Dương','COUNTRY_VN',12,137,NULL,true,'2019-09-27 14:34:57.618','2019-09-27 14:34:57.618')
,('THPT Đốc Binh Kiều','COUNTRY_VN',12,138,NULL,true,'2019-09-27 14:34:57.619','2019-09-27 14:34:57.619')
,('THPT Lê Văn Phẩm','COUNTRY_VN',12,138,NULL,true,'2019-09-27 14:34:57.619','2019-09-27 14:34:57.619')
,('THPT Lưu Tấn Phát','COUNTRY_VN',12,138,NULL,true,'2019-09-27 14:34:57.620','2019-09-27 14:34:57.620')
,('THPT Phan Việt Thống','COUNTRY_VN',12,138,NULL,true,'2019-09-27 14:34:57.621','2019-09-27 14:34:57.621')
,('THPT Tứ Kiệt','COUNTRY_VN',12,138,NULL,true,'2019-09-27 14:34:57.621','2019-09-27 14:34:57.621')
,('THPT Dưỡng Điềm','COUNTRY_VN',12,139,NULL,true,'2019-09-27 14:34:57.622','2019-09-27 14:34:57.622')
,('THPT Nam Kỳ Khởi Nghĩa','COUNTRY_VN',12,139,NULL,true,'2019-09-27 14:34:57.623','2019-09-27 14:34:57.623')
,('THPT Rạch Gầm-Xoài Mút','COUNTRY_VN',12,139,NULL,true,'2019-09-27 14:34:57.623','2019-09-27 14:34:57.623')
,('THPT Tân Hiệp','COUNTRY_VN',12,139,NULL,true,'2019-09-27 14:34:57.623','2019-09-27 14:34:57.623')
,('THPT Vĩnh Kim','COUNTRY_VN',12,139,NULL,true,'2019-09-27 14:34:57.625','2019-09-27 14:34:57.625')
,('TTGDTX Huyện Châu Thành','COUNTRY_VN',12,139,NULL,true,'2019-09-27 14:34:57.626','2019-09-27 14:34:57.626')
,('THPT Bình Phục Nhứt','COUNTRY_VN',12,140,NULL,true,'2019-09-27 14:34:57.627','2019-09-27 14:34:57.627')
,('THPT Chợ Gạo','COUNTRY_VN',12,140,NULL,true,'2019-09-27 14:34:57.628','2019-09-27 14:34:57.628')
,('THPT Thủ Khoa Huân','COUNTRY_VN',12,140,NULL,true,'2019-09-27 14:34:57.628','2019-09-27 14:34:57.628')
,('THPT Trần Văn Hoài','COUNTRY_VN',12,140,NULL,true,'2019-09-27 14:34:57.629','2019-09-27 14:34:57.629')
,('TTGDTX Huyện Chợ Gạo','COUNTRY_VN',12,140,NULL,true,'2019-09-27 14:34:57.630','2019-09-27 14:34:57.630')
,('THPT Gò Công Đông','COUNTRY_VN',12,141,NULL,true,'2019-09-27 14:34:57.632','2019-09-27 14:34:57.632')
,('THPT Nguyễn Văn Côn','COUNTRY_VN',12,141,NULL,true,'2019-09-27 14:34:57.633','2019-09-27 14:34:57.633')
,('TTGDTX Huyện Gò Công Đông','COUNTRY_VN',12,141,NULL,true,'2019-09-27 14:34:57.633','2019-09-27 14:34:57.633')
,('THPT Long Bình','COUNTRY_VN',12,142,NULL,true,'2019-09-27 14:34:57.634','2019-09-27 14:34:57.634')
,('THPT Nguyễn Văn Thìn','COUNTRY_VN',12,142,NULL,true,'2019-09-27 14:34:57.634','2019-09-27 14:34:57.634')
,('THPT Vĩnh Bình','COUNTRY_VN',12,142,NULL,true,'2019-09-27 14:34:57.634','2019-09-27 14:34:57.634')
,('TTGDTX Huyện Gò Công Tây','COUNTRY_VN',12,142,NULL,true,'2019-09-27 14:34:57.635','2019-09-27 14:34:57.635')
,('THPT Phú Thạnh','COUNTRY_VN',12,143,NULL,true,'2019-09-27 14:34:57.636','2019-09-27 14:34:57.636')
,('THPT Nguyễn Văn Tiếp','COUNTRY_VN',12,144,NULL,true,'2019-09-27 14:34:57.636','2019-09-27 14:34:57.636')
,('THPT Tân Phước','COUNTRY_VN',12,145,NULL,true,'2019-09-27 14:34:57.637','2019-09-27 14:34:57.637')
,('TTGDTX Huyện Tân Phước','COUNTRY_VN',12,145,NULL,true,'2019-09-27 14:34:57.637','2019-09-27 14:34:57.637')
,('THPT Âp Bắc','COUNTRY_VN',12,145,NULL,true,'2019-09-27 14:34:57.638','2019-09-27 14:34:57.638')
,('THPT Chuyên Tiền Giang','COUNTRY_VN',12,145,NULL,true,'2019-09-27 14:34:57.638','2019-09-27 14:34:57.638')
,('THPT Nguyễn Đình Chiếu','COUNTRY_VN',12,145,NULL,true,'2019-09-27 14:34:57.638','2019-09-27 14:34:57.638')
,('THPT NK TDTT','COUNTRY_VN',12,145,NULL,true,'2019-09-27 14:34:57.639','2019-09-27 14:34:57.639')
,('THPT Phước Thạnh','COUNTRY_VN',12,145,NULL,true,'2019-09-27 14:34:57.639','2019-09-27 14:34:57.639')
,('THPT Trần Hung Đạo','COUNTRY_VN',12,145,NULL,true,'2019-09-27 14:34:57.639','2019-09-27 14:34:57.639')
,('TTGDTX Thành phố Mỹ Tho','COUNTRY_VN',12,145,NULL,true,'2019-09-27 14:34:57.640','2019-09-27 14:34:57.640')
,('THPT Bình Đông','COUNTRY_VN',12,146,NULL,true,'2019-09-27 14:34:57.641','2019-09-27 14:34:57.641')
,('THPT Gò Công Đông','COUNTRY_VN',12,146,NULL,true,'2019-09-27 14:34:57.642','2019-09-27 14:34:57.642')
,('THPT Trương Định','COUNTRY_VN',12,146,NULL,true,'2019-09-27 14:34:57.643','2019-09-27 14:34:57.643')
,('THPT A Lưới','COUNTRY_VN',13,147,NULL,true,'2019-09-27 14:34:57.644','2019-09-27 14:34:57.644')
,('THPT Hồng Vân','COUNTRY_VN',13,147,NULL,true,'2019-09-27 14:34:57.645','2019-09-27 14:34:57.645')
,('THPT Hương Lâm','COUNTRY_VN',13,147,NULL,true,'2019-09-27 14:34:57.645','2019-09-27 14:34:57.645')
,('TTGDTX A Lưới','COUNTRY_VN',13,147,NULL,true,'2019-09-27 14:34:57.646','2019-09-27 14:34:57.646')
,('THPT Bình Điền','COUNTRY_VN',13,148,NULL,true,'2019-09-27 14:34:57.647','2019-09-27 14:34:57.647')
,('THPT Đặng Huy Trứ','COUNTRY_VN',13,148,NULL,true,'2019-09-27 14:34:57.648','2019-09-27 14:34:57.648')
,('THPT Hương Trà','COUNTRY_VN',13,148,NULL,true,'2019-09-27 14:34:57.649','2019-09-27 14:34:57.649')
,('THPT Hương Vinh','COUNTRY_VN',13,148,NULL,true,'2019-09-27 14:34:57.649','2019-09-27 14:34:57.649')
,('TTGDTX Hương Trà','COUNTRY_VN',13,148,NULL,true,'2019-09-27 14:34:57.650','2019-09-27 14:34:57.650')
,('THPT Hương Giang','COUNTRY_VN',13,149,NULL,true,'2019-09-27 14:34:57.650','2019-09-27 14:34:57.650')
,('THPT Nam Đông','COUNTRY_VN',13,149,NULL,true,'2019-09-27 14:34:57.651','2019-09-27 14:34:57.651')
,('TTGDTX Nam Đông','COUNTRY_VN',13,149,NULL,true,'2019-09-27 14:34:57.651','2019-09-27 14:34:57.651')
,('THPT Nguyễn Đình Chiểu','COUNTRY_VN',13,150,NULL,true,'2019-09-27 14:34:57.652','2019-09-27 14:34:57.652')
,('THPT Phong Điền','COUNTRY_VN',13,150,NULL,true,'2019-09-27 14:34:57.653','2019-09-27 14:34:57.653')
,('THPT Tam Giang','COUNTRY_VN',13,150,NULL,true,'2019-09-27 14:34:57.653','2019-09-27 14:34:57.653')
,('THPT Trần Văn Kỷ','COUNTRY_VN',13,150,NULL,true,'2019-09-27 14:34:57.653','2019-09-27 14:34:57.653')
,('TTGDTX Phong Điền','COUNTRY_VN',13,150,NULL,true,'2019-09-27 14:34:57.654','2019-09-27 14:34:57.654')
,('THPT An Lương Đông','COUNTRY_VN',13,151,NULL,true,'2019-09-27 14:34:57.654','2019-09-27 14:34:57.654')
,('THPT Phú Lộc','COUNTRY_VN',13,151,NULL,true,'2019-09-27 14:34:57.655','2019-09-27 14:34:57.655')
,('THPT Thừa Lưu','COUNTRY_VN',13,151,NULL,true,'2019-09-27 14:34:57.655','2019-09-27 14:34:57.655')
,('THPT Tư thục Thế Hệ Mới','COUNTRY_VN',13,151,NULL,true,'2019-09-27 14:34:57.655','2019-09-27 14:34:57.655')
,('THPT Vinh Lộc','COUNTRY_VN',13,151,NULL,true,'2019-09-27 14:34:57.656','2019-09-27 14:34:57.656')
,('TTGDTX Phú Lộc','COUNTRY_VN',13,151,NULL,true,'2019-09-27 14:34:57.656','2019-09-27 14:34:57.656')
,('THPT Huế star','COUNTRY_VN',13,152,NULL,true,'2019-09-27 14:34:57.657','2019-09-27 14:34:57.657')
,('THPT Hà Trung','COUNTRY_VN',13,152,NULL,true,'2019-09-27 14:34:57.658','2019-09-27 14:34:57.658')
,('THPT Nguyễn Sinh Cung','COUNTRY_VN',13,152,NULL,true,'2019-09-27 14:34:57.659','2019-09-27 14:34:57.659')
,('THPT Phan Đăng Luu','COUNTRY_VN',13,152,NULL,true,'2019-09-27 14:34:57.660','2019-09-27 14:34:57.660')
,('THPT Thuận An','COUNTRY_VN',13,152,NULL,true,'2019-09-27 14:34:57.660','2019-09-27 14:34:57.660')
,('THPT Vinh Xuân','COUNTRY_VN',13,152,NULL,true,'2019-09-27 14:34:57.661','2019-09-27 14:34:57.661')
,('TTGDTX Phú Vang','COUNTRY_VN',13,152,NULL,true,'2019-09-27 14:34:57.661','2019-09-27 14:34:57.661')
,('THPT Hóa Châu','COUNTRY_VN',13,153,NULL,true,'2019-09-27 14:34:57.662','2019-09-27 14:34:57.662')
,('THPT Nguyễn Chí Thanh','COUNTRY_VN',13,153,NULL,true,'2019-09-27 14:34:57.663','2019-09-27 14:34:57.663')
,('THPT Tố Hữu','COUNTRY_VN',13,153,NULL,true,'2019-09-27 14:34:57.665','2019-09-27 14:34:57.665')
,('TTGDTX Quảng Điền','COUNTRY_VN',13,153,NULL,true,'2019-09-27 14:34:57.666','2019-09-27 14:34:57.666')
,('THPT Dân lập Trần Hưng Đạo','COUNTRY_VN',13,154,NULL,true,'2019-09-27 14:34:57.666','2019-09-27 14:34:57.666')
,('THPT Đặng Trần Côn','COUNTRY_VN',13,154,NULL,true,'2019-09-27 14:34:57.667','2019-09-27 14:34:57.667')
,('THPT Gia Hội','COUNTRY_VN',13,154,NULL,true,'2019-09-27 14:34:57.667','2019-09-27 14:34:57.667')
,('THPT Hai Bà Trưng','COUNTRY_VN',13,154,NULL,true,'2019-09-27 14:34:57.667','2019-09-27 14:34:57.667')
,('THPT Nguyễn Huệ','COUNTRY_VN',13,154,NULL,true,'2019-09-27 14:34:57.668','2019-09-27 14:34:57.668')
,('THPT Nguyễn Trường Tộ','COUNTRY_VN',13,154,NULL,true,'2019-09-27 14:34:57.668','2019-09-27 14:34:57.668')
,('TTGDTX Thành phố Huế','COUNTRY_VN',13,154,NULL,true,'2019-09-27 14:34:57.669','2019-09-27 14:34:57.669')
,('THPT Bùi Thị Xuân','COUNTRY_VN',13,154,NULL,true,'2019-09-27 14:34:57.669','2019-09-27 14:34:57.669')
,('THPT Cao Thắng','COUNTRY_VN',13,154,NULL,true,'2019-09-27 14:34:57.670','2019-09-27 14:34:57.670')
,('THPT Chi Lăng','COUNTRY_VN',13,154,NULL,true,'2019-09-27 14:34:57.670','2019-09-27 14:34:57.670')
,('THPT Chuyên Quốc Học','COUNTRY_VN',13,154,NULL,true,'2019-09-27 14:34:57.670','2019-09-27 14:34:57.670')
,('THPT Dân tộc Nội trú Thừa Thiên Huế','COUNTRY_VN',13,154,NULL,true,'2019-09-27 14:34:57.671','2019-09-27 14:34:57.671')
,('THPT Hương Thủy','COUNTRY_VN',13,155,NULL,true,'2019-09-27 14:34:57.671','2019-09-27 14:34:57.671')
,('THPT Nguyễn Trãi','COUNTRY_VN',13,155,NULL,true,'2019-09-27 14:34:57.672','2019-09-27 14:34:57.672')
,('THPT Phú Bài','COUNTRY_VN',13,155,NULL,true,'2019-09-27 14:34:57.672','2019-09-27 14:34:57.672')
,('TTGDTX Hương Thủy','COUNTRY_VN',13,155,NULL,true,'2019-09-27 14:34:57.673','2019-09-27 14:34:57.673')
,('THPT Bình Điền','COUNTRY_VN',13,156,NULL,true,'2019-09-27 14:34:57.673','2019-09-27 14:34:57.673')
,('THPT Đặng Huy Trứ','COUNTRY_VN',13,156,NULL,true,'2019-09-27 14:34:57.674','2019-09-27 14:34:57.674')
,('THPT Hương Trà','COUNTRY_VN',13,156,NULL,true,'2019-09-27 14:34:57.675','2019-09-27 14:34:57.675')
,('THPT Hương Vinh','COUNTRY_VN',13,156,NULL,true,'2019-09-27 14:34:57.676','2019-09-27 14:34:57.676')
,('TTGDTX Hương Trà','COUNTRY_VN',13,156,NULL,true,'2019-09-27 14:34:57.677','2019-09-27 14:34:57.677')
,('THPT Bá Thước','COUNTRY_VN',14,157,NULL,true,'2019-09-27 14:34:57.678','2019-09-27 14:34:57.678')
,('THPT Bá Thước 3','COUNTRY_VN',14,157,NULL,true,'2019-09-27 14:34:57.679','2019-09-27 14:34:57.679')
,('THPT Hà Văn Mao','COUNTRY_VN',14,157,NULL,true,'2019-09-27 14:34:57.679','2019-09-27 14:34:57.679')
,('TTGDTX Huyện Bá Thước','COUNTRY_VN',14,157,NULL,true,'2019-09-27 14:34:57.680','2019-09-27 14:34:57.680')
,('THPT Cẩm Thủy 1','COUNTRY_VN',14,158,NULL,true,'2019-09-27 14:34:57.681','2019-09-27 14:34:57.681')
,('THPT Cẩm Thủy 2','COUNTRY_VN',14,158,NULL,true,'2019-09-27 14:34:57.682','2019-09-27 14:34:57.682')
,('THPT Cẩm Thủy 3','COUNTRY_VN',14,158,NULL,true,'2019-09-27 14:34:57.683','2019-09-27 14:34:57.683')
,('TTGDTX Huyện Cẩm Thủy','COUNTRY_VN',14,158,NULL,true,'2019-09-27 14:34:57.683','2019-09-27 14:34:57.683')
,('THPT Đông Sơn 1','COUNTRY_VN',14,159,NULL,true,'2019-09-27 14:34:57.685','2019-09-27 14:34:57.685')
,('THPT Đông Sơn 2','COUNTRY_VN',14,159,NULL,true,'2019-09-27 14:34:57.685','2019-09-27 14:34:57.685')
,('THPT Nguyễn Mộng Tuân','COUNTRY_VN',14,159,NULL,true,'2019-09-27 14:34:57.686','2019-09-27 14:34:57.686')
,('TTGDTX Huyện Đông Sơn','COUNTRY_VN',14,159,NULL,true,'2019-09-27 14:34:57.686','2019-09-27 14:34:57.686')
,('THPT Hà Trung','COUNTRY_VN',14,160,NULL,true,'2019-09-27 14:34:57.687','2019-09-27 14:34:57.687')
,('THPT Hoàng Lệ Kha','COUNTRY_VN',14,160,NULL,true,'2019-09-27 14:34:57.687','2019-09-27 14:34:57.687')
,('THPT Nguyễn Hoàng','COUNTRY_VN',14,160,NULL,true,'2019-09-27 14:34:57.687','2019-09-27 14:34:57.687')
,('TTGDTX Huyện Hà Trung','COUNTRY_VN',14,160,NULL,true,'2019-09-27 14:34:57.688','2019-09-27 14:34:57.688')
,('THPT Đnh Chương Dương','COUNTRY_VN',14,160,NULL,true,'2019-09-27 14:34:57.688','2019-09-27 14:34:57.688')
,('THPT Hậu Lộc 1','COUNTRY_VN',14,161,NULL,true,'2019-09-27 14:34:57.689','2019-09-27 14:34:57.689')
,('THPT Hậu Lộc 2','COUNTRY_VN',14,161,NULL,true,'2019-09-27 14:34:57.689','2019-09-27 14:34:57.689')
,('THPT Hậu Lộc 3','COUNTRY_VN',14,161,NULL,true,'2019-09-27 14:34:57.690','2019-09-27 14:34:57.690')
,('THPT Hậu Lộc 4','COUNTRY_VN',14,161,NULL,true,'2019-09-27 14:34:57.690','2019-09-27 14:34:57.690')
,('TTGDTX Huyện Hậu Lộc','COUNTRY_VN',14,161,NULL,true,'2019-09-27 14:34:57.691','2019-09-27 14:34:57.691')
,('THPT Hoằng Hóa','COUNTRY_VN',14,162,NULL,true,'2019-09-27 14:34:57.693','2019-09-27 14:34:57.693')
,('THPT Hoằng Hóa 2','COUNTRY_VN',14,162,NULL,true,'2019-09-27 14:34:57.693','2019-09-27 14:34:57.693')
,('THPT Hoằng Hóa 3','COUNTRY_VN',14,162,NULL,true,'2019-09-27 14:34:57.694','2019-09-27 14:34:57.694')
,('THPT Hoằng Hóa 4','COUNTRY_VN',14,162,NULL,true,'2019-09-27 14:34:57.694','2019-09-27 14:34:57.694')
,('THPT Lê Viết Tạo','COUNTRY_VN',14,162,NULL,true,'2019-09-27 14:34:57.695','2019-09-27 14:34:57.695')
,('THPT Lương Đắc Bằng','COUNTRY_VN',14,162,NULL,true,'2019-09-27 14:34:57.696','2019-09-27 14:34:57.696')
,('THPT Lưu Đình Chất','COUNTRY_VN',14,162,NULL,true,'2019-09-27 14:34:57.696','2019-09-27 14:34:57.696')
,('TTGDTX Huyện Hoằng Hóa','COUNTRY_VN',14,162,NULL,true,'2019-09-27 14:34:57.697','2019-09-27 14:34:57.697')
,('THPT Lang Chánh','COUNTRY_VN',14,163,NULL,true,'2019-09-27 14:34:57.698','2019-09-27 14:34:57.698')
,('TTGDTX Huyện Lang Chánh','COUNTRY_VN',14,163,NULL,true,'2019-09-27 14:34:57.699','2019-09-27 14:34:57.699')
,('THPT Mường Lát','COUNTRY_VN',14,164,NULL,true,'2019-09-27 14:34:57.700','2019-09-27 14:34:57.700')
,('TTGDTX Huyện Mường Lát','COUNTRY_VN',14,164,NULL,true,'2019-09-27 14:34:57.700','2019-09-27 14:34:57.700')
,('THPT Ba Đình','COUNTRY_VN',14,165,NULL,true,'2019-09-27 14:34:57.701','2019-09-27 14:34:57.701')
,('THPT Mai Anh Tuấn','COUNTRY_VN',14,165,NULL,true,'2019-09-27 14:34:57.701','2019-09-27 14:34:57.701')
,('THPT Nga Sơn','COUNTRY_VN',14,165,NULL,true,'2019-09-27 14:34:57.702','2019-09-27 14:34:57.702')
,('THPT Trần Phú','COUNTRY_VN',14,165,NULL,true,'2019-09-27 14:34:57.702','2019-09-27 14:34:57.702')
,('TTGDTX Huyện Nga Sơn','COUNTRY_VN',14,165,NULL,true,'2019-09-27 14:34:57.703','2019-09-27 14:34:57.703')
,('THPT Bắc Sơn','COUNTRY_VN',14,166,NULL,true,'2019-09-27 14:34:57.704','2019-09-27 14:34:57.704')
,('THPT Lê Lai','COUNTRY_VN',14,166,NULL,true,'2019-09-27 14:34:57.704','2019-09-27 14:34:57.704')
,('THPT Ngọc Lặc','COUNTRY_VN',14,166,NULL,true,'2019-09-27 14:34:57.705','2019-09-27 14:34:57.705')
,('TTGDTX Huyện Ngọc Lặc','COUNTRY_VN',14,166,NULL,true,'2019-09-27 14:34:57.705','2019-09-27 14:34:57.705')
,('THCS-THPT Như Thanh','COUNTRY_VN',14,167,NULL,true,'2019-09-27 14:34:57.706','2019-09-27 14:34:57.706')
,('THPT Như Thanh','COUNTRY_VN',14,167,NULL,true,'2019-09-27 14:34:57.706','2019-09-27 14:34:57.706')
,('THPT Như Thanh 2','COUNTRY_VN',14,167,NULL,true,'2019-09-27 14:34:57.707','2019-09-27 14:34:57.707')
,('TTGDTX Huyện Như Thanh','COUNTRY_VN',14,167,NULL,true,'2019-09-27 14:34:57.708','2019-09-27 14:34:57.708')
,('THPT Như Xuân','COUNTRY_VN',14,168,NULL,true,'2019-09-27 14:34:57.709','2019-09-27 14:34:57.709')
,('THPT Như Xuân 2','COUNTRY_VN',14,168,NULL,true,'2019-09-27 14:34:57.710','2019-09-27 14:34:57.710')
,('TTGDTX Huyện Như Xuân','COUNTRY_VN',14,168,NULL,true,'2019-09-27 14:34:57.710','2019-09-27 14:34:57.710')
,('THPT Nông Cống','COUNTRY_VN',14,169,NULL,true,'2019-09-27 14:34:57.711','2019-09-27 14:34:57.711')
,('THPT Nông Cống 1','COUNTRY_VN',14,169,NULL,true,'2019-09-27 14:34:57.711','2019-09-27 14:34:57.711')
,('THPT Nông Cống 2','COUNTRY_VN',14,169,NULL,true,'2019-09-27 14:34:57.712','2019-09-27 14:34:57.712')
,('THPT Nông Cống 3','COUNTRY_VN',14,169,NULL,true,'2019-09-27 14:34:57.712','2019-09-27 14:34:57.712')
,('THPT Nông Cống 4','COUNTRY_VN',14,169,NULL,true,'2019-09-27 14:34:57.713','2019-09-27 14:34:57.713')
,('THPT Triệu Thị Trinh','COUNTRY_VN',14,169,NULL,true,'2019-09-27 14:34:57.714','2019-09-27 14:34:57.714')
,('TTGDTX Huyện Nông Cống','COUNTRY_VN',14,169,NULL,true,'2019-09-27 14:34:57.714','2019-09-27 14:34:57.714')
,('THCS-THPT Quan Hóa','COUNTRY_VN',14,170,NULL,true,'2019-09-27 14:34:57.715','2019-09-27 14:34:57.715')
,('THPT Quan Hóa','COUNTRY_VN',14,170,NULL,true,'2019-09-27 14:34:57.716','2019-09-27 14:34:57.716')
,('TTGDTX Huyện Quan Hóa','COUNTRY_VN',14,170,NULL,true,'2019-09-27 14:34:57.717','2019-09-27 14:34:57.717')
,('THPT Quan Sơn','COUNTRY_VN',14,171,NULL,true,'2019-09-27 14:34:57.717','2019-09-27 14:34:57.717')
,('THPT Quan Sơn 2','COUNTRY_VN',14,171,NULL,true,'2019-09-27 14:34:57.718','2019-09-27 14:34:57.718')
,('TTGDTX Huyện Quan Sơn','COUNTRY_VN',14,171,NULL,true,'2019-09-27 14:34:57.718','2019-09-27 14:34:57.718')
,('THPT Đặng Thai Mai','COUNTRY_VN',14,172,NULL,true,'2019-09-27 14:34:57.719','2019-09-27 14:34:57.719')
,('THPT Nguyễn Xuân Nguyên','COUNTRY_VN',14,172,NULL,true,'2019-09-27 14:34:57.719','2019-09-27 14:34:57.719')
,('THPT Quảng Xương 1','COUNTRY_VN',14,172,NULL,true,'2019-09-27 14:34:57.720','2019-09-27 14:34:57.720')
,('THPT Quảng Xương 2','COUNTRY_VN',14,172,NULL,true,'2019-09-27 14:34:57.720','2019-09-27 14:34:57.720')
,('THPT Quảng Xương 3','COUNTRY_VN',14,172,NULL,true,'2019-09-27 14:34:57.720','2019-09-27 14:34:57.720')
,('THPT Quảng Xương 4','COUNTRY_VN',14,172,NULL,true,'2019-09-27 14:34:57.721','2019-09-27 14:34:57.721')
,('TTGDTX Huyện Quảng Xương','COUNTRY_VN',14,172,NULL,true,'2019-09-27 14:34:57.721','2019-09-27 14:34:57.721')
,('THPT Thạch Thành 1','COUNTRY_VN',14,173,NULL,true,'2019-09-27 14:34:57.722','2019-09-27 14:34:57.722')
,('THPT Thạch Thành 2','COUNTRY_VN',14,173,NULL,true,'2019-09-27 14:34:57.722','2019-09-27 14:34:57.722')
,('THPT Thạch Thành 3','COUNTRY_VN',14,173,NULL,true,'2019-09-27 14:34:57.723','2019-09-27 14:34:57.723')
,('THPT Thạch Thành 4','COUNTRY_VN',14,173,NULL,true,'2019-09-27 14:34:57.723','2019-09-27 14:34:57.723')
,('TTGDTX Thạch Thành','COUNTRY_VN',14,173,NULL,true,'2019-09-27 14:34:57.723','2019-09-27 14:34:57.723')
,('THPT Dương Đình Nghệ','COUNTRY_VN',14,174,NULL,true,'2019-09-27 14:34:57.726','2019-09-27 14:34:57.726')
,('THPT Lê Văn Hưu','COUNTRY_VN',14,174,NULL,true,'2019-09-27 14:34:57.727','2019-09-27 14:34:57.727')
,('THPT Nguyễn Quán Nho','COUNTRY_VN',14,174,NULL,true,'2019-09-27 14:34:57.727','2019-09-27 14:34:57.727')
,('THPT Thiệu Hóa','COUNTRY_VN',14,174,NULL,true,'2019-09-27 14:34:57.728','2019-09-27 14:34:57.728')
,('TTGDTX Huyện Thiệu Hóa','COUNTRY_VN',14,174,NULL,true,'2019-09-27 14:34:57.729','2019-09-27 14:34:57.729')
,('THPT Lam Kinh','COUNTRY_VN',14,175,NULL,true,'2019-09-27 14:34:57.730','2019-09-27 14:34:57.730')
,('THPT Lê Hoàn','COUNTRY_VN',14,175,NULL,true,'2019-09-27 14:34:57.730','2019-09-27 14:34:57.730')
,('THPT Lê Lợi','COUNTRY_VN',14,175,NULL,true,'2019-09-27 14:34:57.731','2019-09-27 14:34:57.731')
,('THPT Lê Văn Linh','COUNTRY_VN',14,175,NULL,true,'2019-09-27 14:34:57.731','2019-09-27 14:34:57.731')
,('THPT Thọ Xuân 4','COUNTRY_VN',14,175,NULL,true,'2019-09-27 14:34:57.732','2019-09-27 14:34:57.732')
,('THPT Thọ Xuân 5','COUNTRY_VN',14,175,NULL,true,'2019-09-27 14:34:57.732','2019-09-27 14:34:57.732')
,('TTGDTX Huyện Thọ Xuân','COUNTRY_VN',14,175,NULL,true,'2019-09-27 14:34:57.733','2019-09-27 14:34:57.733')
,('THCS-THPT Thống Nhất','COUNTRY_VN',14,176,NULL,true,'2019-09-27 14:34:57.734','2019-09-27 14:34:57.734')
,('THPT Cầm Bá Thước','COUNTRY_VN',14,177,NULL,true,'2019-09-27 14:34:57.735','2019-09-27 14:34:57.735')
,('THPT Thường Xuân 2','COUNTRY_VN',14,177,NULL,true,'2019-09-27 14:34:57.736','2019-09-27 14:34:57.736')
,('THPT Thường Xuân 3','COUNTRY_VN',14,177,NULL,true,'2019-09-27 14:34:57.736','2019-09-27 14:34:57.736')
,('TTGDTX Huyện Thường Xuân','COUNTRY_VN',14,177,NULL,true,'2019-09-27 14:34:57.737','2019-09-27 14:34:57.737')
,('THCS-THPT Nghi Sơn','COUNTRY_VN',14,178,NULL,true,'2019-09-27 14:34:57.737','2019-09-27 14:34:57.737')
,('THPT Tĩnh Gia 1','COUNTRY_VN',14,178,NULL,true,'2019-09-27 14:34:57.738','2019-09-27 14:34:57.738')
,('THPT Tĩnh Gia 2','COUNTRY_VN',14,178,NULL,true,'2019-09-27 14:34:57.738','2019-09-27 14:34:57.738')
,('THPT Tĩnh Gia 3','COUNTRY_VN',14,178,NULL,true,'2019-09-27 14:34:57.739','2019-09-27 14:34:57.739')
,('THPT Tĩnh Gia 4','COUNTRY_VN',14,178,NULL,true,'2019-09-27 14:34:57.739','2019-09-27 14:34:57.739')
,('THPT Tĩnh Gia 5','COUNTRY_VN',14,178,NULL,true,'2019-09-27 14:34:57.739','2019-09-27 14:34:57.739')
,('TTGDTX Huyện Tĩnh Gia','COUNTRY_VN',14,178,NULL,true,'2019-09-27 14:34:57.740','2019-09-27 14:34:57.740')
,('THPT Triệu Sơn','COUNTRY_VN',14,179,NULL,true,'2019-09-27 14:34:57.742','2019-09-27 14:34:57.742')
,('THPT Triệu Sơn 1','COUNTRY_VN',14,179,NULL,true,'2019-09-27 14:34:57.743','2019-09-27 14:34:57.743')
,('THPT Triệu Sơn 2','COUNTRY_VN',14,179,NULL,true,'2019-09-27 14:34:57.743','2019-09-27 14:34:57.743')
,('THPT Triệu Sơn 3','COUNTRY_VN',14,179,NULL,true,'2019-09-27 14:34:57.744','2019-09-27 14:34:57.744')
,('THPT Triệu Sơn 4','COUNTRY_VN',14,179,NULL,true,'2019-09-27 14:34:57.744','2019-09-27 14:34:57.744')
,('THPT Triệu Sơn 5','COUNTRY_VN',14,179,NULL,true,'2019-09-27 14:34:57.745','2019-09-27 14:34:57.745')
,('THPT Triệu Sơn 6','COUNTRY_VN',14,179,NULL,true,'2019-09-27 14:34:57.745','2019-09-27 14:34:57.745')
,('TTGDTX Huyện Triệu Sơn','COUNTRY_VN',14,179,NULL,true,'2019-09-27 14:34:57.746','2019-09-27 14:34:57.746')
,('THPT Tống Duy Tân','COUNTRY_VN',14,180,NULL,true,'2019-09-27 14:34:57.748','2019-09-27 14:34:57.748')
,('THPT Trần Khát Chân','COUNTRY_VN',14,180,NULL,true,'2019-09-27 14:34:57.748','2019-09-27 14:34:57.748')
,('THPT Vĩnh Lộc','COUNTRY_VN',14,180,NULL,true,'2019-09-27 14:34:57.749','2019-09-27 14:34:57.749')
,('TTGDTX Huyện Vĩnh Lộc','COUNTRY_VN',14,180,NULL,true,'2019-09-27 14:34:57.749','2019-09-27 14:34:57.749')
,('THPT Hà Tông Huân','COUNTRY_VN',14,181,NULL,true,'2019-09-27 14:34:57.751','2019-09-27 14:34:57.751')
,('THPT Trần Ân Chiêm','COUNTRY_VN',14,181,NULL,true,'2019-09-27 14:34:57.752','2019-09-27 14:34:57.752')
,('THPT Yên Định 1','COUNTRY_VN',14,181,NULL,true,'2019-09-27 14:34:57.753','2019-09-27 14:34:57.753')
,('THPT Yên Định 2','COUNTRY_VN',14,181,NULL,true,'2019-09-27 14:34:57.753','2019-09-27 14:34:57.753')
,('THPT Yên Định 3','COUNTRY_VN',14,181,NULL,true,'2019-09-27 14:34:57.753','2019-09-27 14:34:57.753')
,('TTGDTX Huyện Yên Định','COUNTRY_VN',14,181,NULL,true,'2019-09-27 14:34:57.754','2019-09-27 14:34:57.754')
,('THPT Chuyên Lam Sơn','COUNTRY_VN',14,182,NULL,true,'2019-09-27 14:34:57.755','2019-09-27 14:34:57.755')
,('THPT Dân Tộc Nội trú Thanh Hóa','COUNTRY_VN',14,182,NULL,true,'2019-09-27 14:34:57.755','2019-09-27 14:34:57.755')
,('THPT Đào Duy Anh','COUNTRY_VN',14,182,NULL,true,'2019-09-27 14:34:57.756','2019-09-27 14:34:57.756')
,('THPT Đào Duy Từ','COUNTRY_VN',14,182,NULL,true,'2019-09-27 14:34:57.756','2019-09-27 14:34:57.756')
,('THPT Đống Sơn','COUNTRY_VN',14,182,NULL,true,'2019-09-27 14:34:57.757','2019-09-27 14:34:57.757')
,('THPT Hàm Rồng','COUNTRY_VN',14,182,NULL,true,'2019-09-27 14:34:57.758','2019-09-27 14:34:57.758')
,('THPT Lý Thường Kiệt','COUNTRY_VN',14,182,NULL,true,'2019-09-27 14:34:57.760','2019-09-27 14:34:57.760')
,('THPT Nguyễn Huệ','COUNTRY_VN',14,182,NULL,true,'2019-09-27 14:34:57.760','2019-09-27 14:34:57.760')
,('THPT Nguyễn Trãi','COUNTRY_VN',14,182,NULL,true,'2019-09-27 14:34:57.761','2019-09-27 14:34:57.761')
,('THPT Tô Hiến Thành','COUNTRY_VN',14,182,NULL,true,'2019-09-27 14:34:57.761','2019-09-27 14:34:57.761')
,('THPT Trường Thi','COUNTRY_VN',14,182,NULL,true,'2019-09-27 14:34:57.763','2019-09-27 14:34:57.763')
,('TTGDTX Tỉnh Thanh Hoá','COUNTRY_VN',14,182,NULL,true,'2019-09-27 14:34:57.763','2019-09-27 14:34:57.763')
,('TTGDTX Thành phố Thanh Hoá','COUNTRY_VN',14,182,NULL,true,'2019-09-27 14:34:57.764','2019-09-27 14:34:57.764')
,('THPT Bỉm Sơn','COUNTRY_VN',14,183,NULL,true,'2019-09-27 14:34:57.765','2019-09-27 14:34:57.765')
,('THPT Lê Hồng Phong','COUNTRY_VN',14,183,NULL,true,'2019-09-27 14:34:57.766','2019-09-27 14:34:57.766')
,('TTGDTX Thị xã Bỉm Sơn','COUNTRY_VN',14,183,NULL,true,'2019-09-27 14:34:57.766','2019-09-27 14:34:57.766')
,('THPT Nguyễn Thị Lợi','COUNTRY_VN',14,184,NULL,true,'2019-09-27 14:34:57.767','2019-09-27 14:34:57.767')
,('THPT Sầm Sơn','COUNTRY_VN',14,184,NULL,true,'2019-09-27 14:34:57.767','2019-09-27 14:34:57.767')
,('TTGDTX Thị xã Sầm Sơn','COUNTRY_VN',14,184,NULL,true,'2019-09-27 14:34:57.768','2019-09-27 14:34:57.768')
,('THPT Đại Từ','COUNTRY_VN',15,185,NULL,true,'2019-09-27 14:34:57.770','2019-09-27 14:34:57.770')
,('THPT Lưu Nhân Chú','COUNTRY_VN',15,185,NULL,true,'2019-09-27 14:34:57.770','2019-09-27 14:34:57.770')
,('THPT Nguyễn Huệ','COUNTRY_VN',15,185,NULL,true,'2019-09-27 14:34:57.771','2019-09-27 14:34:57.771')
,('TTGDTX Huyện Đại Từ','COUNTRY_VN',15,185,NULL,true,'2019-09-27 14:34:57.771','2019-09-27 14:34:57.771')
,('THPT Bình Yên','COUNTRY_VN',15,186,NULL,true,'2019-09-27 14:34:57.772','2019-09-27 14:34:57.772')
,('THPT Định Hóa','COUNTRY_VN',15,186,NULL,true,'2019-09-27 14:34:57.772','2019-09-27 14:34:57.772')
,('TTGDTX Huyện Định Hóa','COUNTRY_VN',15,186,NULL,true,'2019-09-27 14:34:57.773','2019-09-27 14:34:57.773')
,('THPT Đồng Hỷ','COUNTRY_VN',15,187,NULL,true,'2019-09-27 14:34:57.773','2019-09-27 14:34:57.773')
,('THPT Trại Cau','COUNTRY_VN',15,187,NULL,true,'2019-09-27 14:34:57.774','2019-09-27 14:34:57.774')
,('THPT Trần Quốc Tuấn','COUNTRY_VN',15,187,NULL,true,'2019-09-27 14:34:57.777','2019-09-27 14:34:57.777')
,('TTGDTX Huyện Đồng Hỷ','COUNTRY_VN',15,187,NULL,true,'2019-09-27 14:34:57.778','2019-09-27 14:34:57.778')
,('THPT Bắc Sơn','COUNTRY_VN',15,188,NULL,true,'2019-09-27 14:34:57.779','2019-09-27 14:34:57.779')
,('THPT Lê Hồng Phong','COUNTRY_VN',15,188,NULL,true,'2019-09-27 14:34:57.779','2019-09-27 14:34:57.779')
,('THPT Phổ Yên','COUNTRY_VN',15,188,NULL,true,'2019-09-27 14:34:57.780','2019-09-27 14:34:57.780')
,('TTGDTX Huyện Phổ Yên','COUNTRY_VN',15,188,NULL,true,'2019-09-27 14:34:57.781','2019-09-27 14:34:57.781')
,('THPT Điềm Thụy','COUNTRY_VN',15,189,NULL,true,'2019-09-27 14:34:57.782','2019-09-27 14:34:57.782')
,('THPT Lương Phú','COUNTRY_VN',15,189,NULL,true,'2019-09-27 14:34:57.783','2019-09-27 14:34:57.783')
,('THPT Phú Bình','COUNTRY_VN',15,189,NULL,true,'2019-09-27 14:34:57.784','2019-09-27 14:34:57.784')
,('TTGDTX Huyện Phú Bình','COUNTRY_VN',15,189,NULL,true,'2019-09-27 14:34:57.785','2019-09-27 14:34:57.785')
,('THPT Khánh Hòa','COUNTRY_VN',15,190,NULL,true,'2019-09-27 14:34:57.785','2019-09-27 14:34:57.785')
,('THPT Phú Lương','COUNTRY_VN',15,190,NULL,true,'2019-09-27 14:34:57.786','2019-09-27 14:34:57.786')
,('THPT Yên Ninh','COUNTRY_VN',15,190,NULL,true,'2019-09-27 14:34:57.787','2019-09-27 14:34:57.787')
,('TTGDTX Huyện Phú Lương','COUNTRY_VN',15,190,NULL,true,'2019-09-27 14:34:57.787','2019-09-27 14:34:57.787')
,('THPT Hoàng Quốc Việt','COUNTRY_VN',15,191,NULL,true,'2019-09-27 14:34:57.788','2019-09-27 14:34:57.788')
,('THPT Trần Phú','COUNTRY_VN',15,191,NULL,true,'2019-09-27 14:34:57.788','2019-09-27 14:34:57.788')
,('THPT Võ Nhai','COUNTRY_VN',15,191,NULL,true,'2019-09-27 14:34:57.789','2019-09-27 14:34:57.789')
,('TTGDTX Huyện Võ Nhai','COUNTRY_VN',15,191,NULL,true,'2019-09-27 14:34:57.790','2019-09-27 14:34:57.790')
,('THPT Lương Thế Vinh','COUNTRY_VN',15,192,NULL,true,'2019-09-27 14:34:57.790','2019-09-27 14:34:57.790')
,('THPT Sông Công','COUNTRY_VN',15,192,NULL,true,'2019-09-27 14:34:57.791','2019-09-27 14:34:57.791')
,('TTGDTX Thị xã Sông Công','COUNTRY_VN',15,192,NULL,true,'2019-09-27 14:34:57.792','2019-09-27 14:34:57.792')
,('THPT vùng cao Việt Bắc','COUNTRY_VN',15,193,NULL,true,'2019-09-27 14:34:57.793','2019-09-27 14:34:57.793')
,('THPT Dân tộc Nội trú Thái Nguyên','COUNTRY_VN',15,193,NULL,true,'2019-09-27 14:34:57.793','2019-09-27 14:34:57.793')
,('THPT Bưu chính viễn thông và CNTT Miên Núi','COUNTRY_VN',15,193,NULL,true,'2019-09-27 14:34:57.793','2019-09-27 14:34:57.793')
,('THPT Chu Văn An','COUNTRY_VN',15,193,NULL,true,'2019-09-27 14:34:57.794','2019-09-27 14:34:57.794')
,('THPT Chuyên Thái Nguyên','COUNTRY_VN',15,193,NULL,true,'2019-09-27 14:34:57.795','2019-09-27 14:34:57.795')
,('THPT Dương Tự Minh','COUNTRY_VN',15,193,NULL,true,'2019-09-27 14:34:57.795','2019-09-27 14:34:57.795')
,('THPT Đào Duy Từ','COUNTRY_VN',15,193,NULL,true,'2019-09-27 14:34:57.796','2019-09-27 14:34:57.796')
,('THPT Gang Thép','COUNTRY_VN',15,193,NULL,true,'2019-09-27 14:34:57.797','2019-09-27 14:34:57.797')
,('THPT Lê Quý Đôn','COUNTRY_VN',15,193,NULL,true,'2019-09-27 14:34:57.797','2019-09-27 14:34:57.797')
,('THPT Lương Ngọc Quyến','COUNTRY_VN',15,193,NULL,true,'2019-09-27 14:34:57.798','2019-09-27 14:34:57.798')
,('THPT Ngô Quyền','COUNTRY_VN',15,193,NULL,true,'2019-09-27 14:34:57.798','2019-09-27 14:34:57.798')
,('THPT Thái Nguyên','COUNTRY_VN',15,193,NULL,true,'2019-09-27 14:34:57.799','2019-09-27 14:34:57.799')
,('TTGDTX Tỉnh Thái Nguyên','COUNTRY_VN',15,193,NULL,true,'2019-09-27 14:34:57.799','2019-09-27 14:34:57.799')
,('THPT Bắc Sơn','COUNTRY_VN',15,194,NULL,true,'2019-09-27 14:34:57.801','2019-09-27 14:34:57.801')
,('THPT Lê Hồng Phong','COUNTRY_VN',15,194,NULL,true,'2019-09-27 14:34:57.801','2019-09-27 14:34:57.801')
,('THPT Phổ Yên','COUNTRY_VN',15,194,NULL,true,'2019-09-27 14:34:57.801','2019-09-27 14:34:57.801')
,('TTGDTX Huyện Phổ Yên','COUNTRY_VN',15,194,NULL,true,'2019-09-27 14:34:57.802','2019-09-27 14:34:57.802')
,('THPT Bắc Đông Quan','COUNTRY_VN',16,195,NULL,true,'2019-09-27 14:34:57.803','2019-09-27 14:34:57.803')
,('THPT Đông Quan','COUNTRY_VN',16,195,NULL,true,'2019-09-27 14:34:57.803','2019-09-27 14:34:57.803')
,('THPT Mê Linh','COUNTRY_VN',16,195,NULL,true,'2019-09-27 14:34:57.804','2019-09-27 14:34:57.804')
,('THPT Nam Đông Quan','COUNTRY_VN',16,195,NULL,true,'2019-09-27 14:34:57.804','2019-09-27 14:34:57.804')
,('THPT Tiên Hưng','COUNTRY_VN',16,195,NULL,true,'2019-09-27 14:34:57.804','2019-09-27 14:34:57.804')
,('THPT Tư thục Đông Hưng','COUNTRY_VN',16,195,NULL,true,'2019-09-27 14:34:57.805','2019-09-27 14:34:57.805')
,('TTGDTX Huyện Đông Hưng','COUNTRY_VN',16,195,NULL,true,'2019-09-27 14:34:57.805','2019-09-27 14:34:57.805')
,('THPT Bắc Duyên Hà','COUNTRY_VN',16,196,NULL,true,'2019-09-27 14:34:57.806','2019-09-27 14:34:57.806')
,('THPT Đông Hưng Hà','COUNTRY_VN',16,196,NULL,true,'2019-09-27 14:34:57.806','2019-09-27 14:34:57.806')
,('THPT Hưng Nhân','COUNTRY_VN',16,196,NULL,true,'2019-09-27 14:34:57.807','2019-09-27 14:34:57.807')
,('THPT Nam Duyên Hà','COUNTRY_VN',16,196,NULL,true,'2019-09-27 14:34:57.807','2019-09-27 14:34:57.807')
,('THPT Trần Thị Dung','COUNTRY_VN',16,196,NULL,true,'2019-09-27 14:34:57.808','2019-09-27 14:34:57.808')
,('TTGDTX Huyện Hưng Hà','COUNTRY_VN',16,196,NULL,true,'2019-09-27 14:34:57.809','2019-09-27 14:34:57.809')
,('THPT Bắc Kiến Xương','COUNTRY_VN',16,197,NULL,true,'2019-09-27 14:34:57.810','2019-09-27 14:34:57.810')
,('THPT Bình Thanh','COUNTRY_VN',16,197,NULL,true,'2019-09-27 14:34:57.810','2019-09-27 14:34:57.810')
,('THPT Chu Văn An','COUNTRY_VN',16,197,NULL,true,'2019-09-27 14:34:57.811','2019-09-27 14:34:57.811')
,('THPT Hồng Đức','COUNTRY_VN',16,197,NULL,true,'2019-09-27 14:34:57.811','2019-09-27 14:34:57.811')
,('THPT Nguyễn Du','COUNTRY_VN',16,197,NULL,true,'2019-09-27 14:34:57.812','2019-09-27 14:34:57.812')
,('TTGDTX Huyện Kiến Xương','COUNTRY_VN',16,197,NULL,true,'2019-09-27 14:34:57.812','2019-09-27 14:34:57.812')
,('THPT Nguyễn Huệ','COUNTRY_VN',16,198,NULL,true,'2019-09-27 14:34:57.814','2019-09-27 14:34:57.814')
,('THPT Phụ Dực','COUNTRY_VN',16,198,NULL,true,'2019-09-27 14:34:57.814','2019-09-27 14:34:57.814')
,('THPT Quỳnh Cõi','COUNTRY_VN',16,198,NULL,true,'2019-09-27 14:34:57.815','2019-09-27 14:34:57.815')
,('THPT Quỳnh Thọ','COUNTRY_VN',16,198,NULL,true,'2019-09-27 14:34:57.815','2019-09-27 14:34:57.815')
,('THPT Trần Hưng Đạo','COUNTRY_VN',16,198,NULL,true,'2019-09-27 14:34:57.816','2019-09-27 14:34:57.816')
,('TTGDTX Huyện Quỳnh Phụ I','COUNTRY_VN',16,198,NULL,true,'2019-09-27 14:34:57.816','2019-09-27 14:34:57.816')
,('TTGDTX Huyện Quỳnh Phụ II','COUNTRY_VN',16,198,NULL,true,'2019-09-27 14:34:57.817','2019-09-27 14:34:57.817')
,('THPT Đông Thụy Anh','COUNTRY_VN',16,199,NULL,true,'2019-09-27 14:34:57.817','2019-09-27 14:34:57.817')
,('THPT Tây Thụy Anh','COUNTRY_VN',16,199,NULL,true,'2019-09-27 14:34:57.818','2019-09-27 14:34:57.818')
,('THPT Thái Ninh','COUNTRY_VN',16,199,NULL,true,'2019-09-27 14:34:57.818','2019-09-27 14:34:57.818')
,('THPT Thái Phúc','COUNTRY_VN',16,199,NULL,true,'2019-09-27 14:34:57.819','2019-09-27 14:34:57.819')
,('TTGDTX Huyện Thái Thụy I','COUNTRY_VN',16,199,NULL,true,'2019-09-27 14:34:57.819','2019-09-27 14:34:57.819')
,('TTGDTX Huyện Thái Thụy II','COUNTRY_VN',16,199,NULL,true,'2019-09-27 14:34:57.819','2019-09-27 14:34:57.819')
,('THPT Đông Tiền Hải','COUNTRY_VN',16,200,NULL,true,'2019-09-27 14:34:57.820','2019-09-27 14:34:57.820')
,('THPT Hoàng Văn Thái','COUNTRY_VN',16,200,NULL,true,'2019-09-27 14:34:57.820','2019-09-27 14:34:57.820')
,('THPT Nam Tiền Hải','COUNTRY_VN',16,200,NULL,true,'2019-09-27 14:34:57.821','2019-09-27 14:34:57.821')
,('THPT Tây Tiền Hải','COUNTRY_VN',16,200,NULL,true,'2019-09-27 14:34:57.821','2019-09-27 14:34:57.821')
,('TTGDTX Huyện Tiền Hải','COUNTRY_VN',16,200,NULL,true,'2019-09-27 14:34:57.821','2019-09-27 14:34:57.821')
,('THPT Hùng Vương','COUNTRY_VN',16,201,NULL,true,'2019-09-27 14:34:57.822','2019-09-27 14:34:57.822')
,('THPT Lý Bôn','COUNTRY_VN',16,201,NULL,true,'2019-09-27 14:34:57.822','2019-09-27 14:34:57.822')
,('THPT Nguyễn Trãi','COUNTRY_VN',16,201,NULL,true,'2019-09-27 14:34:57.823','2019-09-27 14:34:57.823')
,('THPT Phạm Quang Thẩm','COUNTRY_VN',16,201,NULL,true,'2019-09-27 14:34:57.823','2019-09-27 14:34:57.823')
,('THPT Vũ Tiên','COUNTRY_VN',16,201,NULL,true,'2019-09-27 14:34:57.823','2019-09-27 14:34:57.823')
,('TTGDTX Huyện Vũ Thư','COUNTRY_VN',16,201,NULL,true,'2019-09-27 14:34:57.825','2019-09-27 14:34:57.825')
,('THPT Chuyên Thái Bình','COUNTRY_VN',16,202,NULL,true,'2019-09-27 14:34:57.826','2019-09-27 14:34:57.826')
,('THPT Diêm Điền','COUNTRY_VN',16,202,NULL,true,'2019-09-27 14:34:57.827','2019-09-27 14:34:57.827')
,('THPT Lê Quý Đôn','COUNTRY_VN',16,202,NULL,true,'2019-09-27 14:34:57.827','2019-09-27 14:34:57.827')
,('THPT Nguyễn Công Trứ','COUNTRY_VN',16,202,NULL,true,'2019-09-27 14:34:57.828','2019-09-27 14:34:57.828')
,('THPT Nguyễn Đức Cảnh','COUNTRY_VN',16,202,NULL,true,'2019-09-27 14:34:57.829','2019-09-27 14:34:57.829')
,('THPT Nguyễn Thái Bình','COUNTRY_VN',16,202,NULL,true,'2019-09-27 14:34:57.830','2019-09-27 14:34:57.830')
,('TTGDTX Thành phố Thái Bình','COUNTRY_VN',16,202,NULL,true,'2019-09-27 14:34:57.830','2019-09-27 14:34:57.830')
,('THPT Nguyễn An Ninh','COUNTRY_VN',17,203,NULL,true,'2019-09-27 14:34:57.832','2019-09-27 14:34:57.832')
,('THPT Trần Phú','COUNTRY_VN',17,203,NULL,true,'2019-09-27 14:34:57.833','2019-09-27 14:34:57.833')
,('TTGDTX Tân Biên','COUNTRY_VN',17,203,NULL,true,'2019-09-27 14:34:57.833','2019-09-27 14:34:57.833')
,('THPT Huỳnh Thúc Kháng','COUNTRY_VN',17,204,NULL,true,'2019-09-27 14:34:57.834','2019-09-27 14:34:57.834')
,('THPT Nguyễn Huệ','COUNTRY_VN',17,204,NULL,true,'2019-09-27 14:34:57.834','2019-09-27 14:34:57.834')
,('TTGDTX Huyện Bến Cầu','COUNTRY_VN',17,204,NULL,true,'2019-09-27 14:34:57.835','2019-09-27 14:34:57.835')
,('THPT Châu Thành','COUNTRY_VN',17,205,NULL,true,'2019-09-27 14:34:57.835','2019-09-27 14:34:57.835')
,('THPT Hoàng Văn Thụ','COUNTRY_VN',17,205,NULL,true,'2019-09-27 14:34:57.836','2019-09-27 14:34:57.836')
,('THPT Lê Hồng Phong','COUNTRY_VN',17,205,NULL,true,'2019-09-27 14:34:57.836','2019-09-27 14:34:57.836')
,('TTGDTX Châu Thành','COUNTRY_VN',17,205,NULL,true,'2019-09-27 14:34:57.837','2019-09-27 14:34:57.837')
,('THPT Dương Minh Châu','COUNTRY_VN',17,206,NULL,true,'2019-09-27 14:34:57.837','2019-09-27 14:34:57.837')
,('THPT Nguyễn Đình Chiểu','COUNTRY_VN',17,206,NULL,true,'2019-09-27 14:34:57.837','2019-09-27 14:34:57.837')
,('THPT Nguyễn Thái Bình','COUNTRY_VN',17,206,NULL,true,'2019-09-27 14:34:57.838','2019-09-27 14:34:57.838')
,('TTGDTX Huyện Dương Minh Châu','COUNTRY_VN',17,206,NULL,true,'2019-09-27 14:34:57.838','2019-09-27 14:34:57.838')
,('THPT Ngô Gia Tự','COUNTRY_VN',17,207,NULL,true,'2019-09-27 14:34:57.839','2019-09-27 14:34:57.839')
,('THPT Nguyễn Văn Trỗi','COUNTRY_VN',17,207,NULL,true,'2019-09-27 14:34:57.839','2019-09-27 14:34:57.839')
,('THPT Quang Trung','COUNTRY_VN',17,207,NULL,true,'2019-09-27 14:34:57.840','2019-09-27 14:34:57.840')
,('THPT Trần Quốc Đại','COUNTRY_VN',17,207,NULL,true,'2019-09-27 14:34:57.840','2019-09-27 14:34:57.840')
,('TTGDTX Huyện Gò Dầu','COUNTRY_VN',17,207,NULL,true,'2019-09-27 14:34:57.841','2019-09-27 14:34:57.841')
,('THPT Lý Thường Kiệt','COUNTRY_VN',17,208,NULL,true,'2019-09-27 14:34:57.842','2019-09-27 14:34:57.842')
,('THPT Nguyễn Chí Thanh','COUNTRY_VN',17,208,NULL,true,'2019-09-27 14:34:57.843','2019-09-27 14:34:57.843')
,('THPT Nguyễn Trung Trực','COUNTRY_VN',17,208,NULL,true,'2019-09-27 14:34:57.843','2019-09-27 14:34:57.843')
,('TTGDTX Huyện Hòa Thành','COUNTRY_VN',17,208,NULL,true,'2019-09-27 14:34:57.844','2019-09-27 14:34:57.844')
,('TTGDTX Tỉnh Tây Ninh','COUNTRY_VN',17,208,NULL,true,'2019-09-27 14:34:57.844','2019-09-27 14:34:57.844')
,('THPT Lương Thế Vinh','COUNTRY_VN',17,209,NULL,true,'2019-09-27 14:34:57.845','2019-09-27 14:34:57.845')
,('THPT Lê Duẩn','COUNTRY_VN',17,210,NULL,true,'2019-09-27 14:34:57.847','2019-09-27 14:34:57.847')
,('THPT Tân Châu','COUNTRY_VN',17,210,NULL,true,'2019-09-27 14:34:57.848','2019-09-27 14:34:57.848')
,('THPT Tân Đông','COUNTRY_VN',17,210,NULL,true,'2019-09-27 14:34:57.849','2019-09-27 14:34:57.849')
,('THPT Tân Hưng','COUNTRY_VN',17,210,NULL,true,'2019-09-27 14:34:57.850','2019-09-27 14:34:57.850')
,('TTGDTX Huyện Tân Châu','COUNTRY_VN',17,210,NULL,true,'2019-09-27 14:34:57.850','2019-09-27 14:34:57.850')
,('THPT Bình Thạnh','COUNTRY_VN',17,211,NULL,true,'2019-09-27 14:34:57.851','2019-09-27 14:34:57.851')
,('THPT Lộc Hưng','COUNTRY_VN',17,211,NULL,true,'2019-09-27 14:34:57.851','2019-09-27 14:34:57.851')
,('THPT Nguyễn Trãi','COUNTRY_VN',17,211,NULL,true,'2019-09-27 14:34:57.852','2019-09-27 14:34:57.852')
,('THPT Trảng Bàng','COUNTRY_VN',17,211,NULL,true,'2019-09-27 14:34:57.852','2019-09-27 14:34:57.852')
,('TTGDTX Huyện Trảng Bàng','COUNTRY_VN',17,211,NULL,true,'2019-09-27 14:34:57.853','2019-09-27 14:34:57.853')
,('THPT Dân tộc Nội trú Tây Ninh','COUNTRY_VN',17,212,NULL,true,'2019-09-27 14:34:57.853','2019-09-27 14:34:57.853')
,('THPT Chuyên Hoàng Lê Kha','COUNTRY_VN',17,212,NULL,true,'2019-09-27 14:34:57.854','2019-09-27 14:34:57.854')
,('THPT Lê Quý Đôn','COUNTRY_VN',17,212,NULL,true,'2019-09-27 14:34:57.854','2019-09-27 14:34:57.854')
,('THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',17,212,NULL,true,'2019-09-27 14:34:57.855','2019-09-27 14:34:57.855')
,('THPT Tây Ninh','COUNTRY_VN',17,212,NULL,true,'2019-09-27 14:34:57.855','2019-09-27 14:34:57.855')
,('THPT Trần Đại Nghĩa','COUNTRY_VN',17,212,NULL,true,'2019-09-27 14:34:57.855','2019-09-27 14:34:57.855')
,('TTGDTX Thành phố Tây Ninh','COUNTRY_VN',17,212,NULL,true,'2019-09-27 14:34:57.855','2019-09-27 14:34:57.855')
,('THPT Lê Duẩn','COUNTRY_VN',17,213,NULL,true,'2019-09-27 14:34:57.856','2019-09-27 14:34:57.856')
,('THPT Tân Châu','COUNTRY_VN',17,213,NULL,true,'2019-09-27 14:34:57.857','2019-09-27 14:34:57.857')
,('THPT Tân Đông','COUNTRY_VN',17,213,NULL,true,'2019-09-27 14:34:57.857','2019-09-27 14:34:57.857')
,('THPT Tân Hưng','COUNTRY_VN',17,213,NULL,true,'2019-09-27 14:34:57.859','2019-09-27 14:34:57.859')
,('TTGDTX Thị xã Tân Châu','COUNTRY_VN',17,213,NULL,true,'2019-09-27 14:34:57.860','2019-09-27 14:34:57.860')
,('THPT Bắc Yên','COUNTRY_VN',18,214,NULL,true,'2019-09-27 14:34:57.861','2019-09-27 14:34:57.861')
,('TTGDTX Huyện Bắc Yên','COUNTRY_VN',18,214,NULL,true,'2019-09-27 14:34:57.861','2019-09-27 14:34:57.861')
,('THPT Chu Văn Thịnh','COUNTRY_VN',18,215,NULL,true,'2019-09-27 14:34:57.863','2019-09-27 14:34:57.863')
,('THPT Cò Nòi','COUNTRY_VN',18,215,NULL,true,'2019-09-27 14:34:57.864','2019-09-27 14:34:57.864')
,('THPT Mai Sơn','COUNTRY_VN',18,215,NULL,true,'2019-09-27 14:34:57.865','2019-09-27 14:34:57.865')
,('TTGDTX Huyện Mai Sơn','COUNTRY_VN',18,215,NULL,true,'2019-09-27 14:34:57.866','2019-09-27 14:34:57.866')
,('THPT Chiềng Sơn','COUNTRY_VN',18,216,NULL,true,'2019-09-27 14:34:57.867','2019-09-27 14:34:57.867')
,('THPT Mộc Lỵ','COUNTRY_VN',18,216,NULL,true,'2019-09-27 14:34:57.867','2019-09-27 14:34:57.867')
,('THPT Tân Lập','COUNTRY_VN',18,216,NULL,true,'2019-09-27 14:34:57.867','2019-09-27 14:34:57.867')
,('THPT Thảo Nguyên','COUNTRY_VN',18,216,NULL,true,'2019-09-27 14:34:57.868','2019-09-27 14:34:57.868')
,('Trung tâm GDTX Mộc Châu','COUNTRY_VN',18,216,NULL,true,'2019-09-27 14:34:57.868','2019-09-27 14:34:57.868')
,('THPT Mường Bú','COUNTRY_VN',18,217,NULL,true,'2019-09-27 14:34:57.869','2019-09-27 14:34:57.869')
,('THPT Mường La','COUNTRY_VN',18,217,NULL,true,'2019-09-27 14:34:57.869','2019-09-27 14:34:57.869')
,('TTGDTX Huyện Mường La','COUNTRY_VN',18,217,NULL,true,'2019-09-27 14:34:57.870','2019-09-27 14:34:57.870')
,('THPT Gia Phù','COUNTRY_VN',18,218,NULL,true,'2019-09-27 14:34:57.870','2019-09-27 14:34:57.870')
,('THPT Phù Yên','COUNTRY_VN',18,218,NULL,true,'2019-09-27 14:34:57.871','2019-09-27 14:34:57.871')
,('THPT Tân Lang','COUNTRY_VN',18,218,NULL,true,'2019-09-27 14:34:57.871','2019-09-27 14:34:57.871')
,('TTGDTX Huyện Phù Yên','COUNTRY_VN',18,218,NULL,true,'2019-09-27 14:34:57.872','2019-09-27 14:34:57.872')
,('THPT Mường Giòn','COUNTRY_VN',18,219,NULL,true,'2019-09-27 14:34:57.873','2019-09-27 14:34:57.873')
,('THPT Quỳnh Nhai','COUNTRY_VN',18,219,NULL,true,'2019-09-27 14:34:57.873','2019-09-27 14:34:57.873')
,('TTGDTX Huyện Quỳnh Nhai','COUNTRY_VN',18,219,NULL,true,'2019-09-27 14:34:57.873','2019-09-27 14:34:57.873')
,('THPT Chiềng Khương','COUNTRY_VN',18,220,NULL,true,'2019-09-27 14:34:57.875','2019-09-27 14:34:57.875')
,('THPT Mường Lầm THPT Sông Mã','COUNTRY_VN',18,220,NULL,true,'2019-09-27 14:34:57.876','2019-09-27 14:34:57.876')
,('TTGDTX Huyện Sông Mã','COUNTRY_VN',18,220,NULL,true,'2019-09-27 14:34:57.877','2019-09-27 14:34:57.877')
,('THPT Sốp Cộp','COUNTRY_VN',18,221,NULL,true,'2019-09-27 14:34:57.878','2019-09-27 14:34:57.878')
,('TTGDTX Huyện Sốp Cộp','COUNTRY_VN',18,221,NULL,true,'2019-09-27 14:34:57.879','2019-09-27 14:34:57.879')
,('THPT Bình Thuận','COUNTRY_VN',18,222,NULL,true,'2019-09-27 14:34:57.881','2019-09-27 14:34:57.881')
,('THPT Co Mạ','COUNTRY_VN',18,222,NULL,true,'2019-09-27 14:34:57.882','2019-09-27 14:34:57.882')
,('THPT Thuận Châu','COUNTRY_VN',18,222,NULL,true,'2019-09-27 14:34:57.883','2019-09-27 14:34:57.883')
,('THPT Tông Lệnh','COUNTRY_VN',18,222,NULL,true,'2019-09-27 14:34:57.883','2019-09-27 14:34:57.883')
,('TTGDTX Huyện Thuận Châu','COUNTRY_VN',18,222,NULL,true,'2019-09-27 14:34:57.883','2019-09-27 14:34:57.883')
,('THPT Mộc Hạ','COUNTRY_VN',18,223,NULL,true,'2019-09-27 14:34:57.884','2019-09-27 14:34:57.884')
,('THPT Phiêng Khoài','COUNTRY_VN',18,224,NULL,true,'2019-09-27 14:34:57.885','2019-09-27 14:34:57.885')
,('THPT Yên Châu','COUNTRY_VN',18,224,NULL,true,'2019-09-27 14:34:57.886','2019-09-27 14:34:57.886')
,('TTGDTX Huyện Yên Châu','COUNTRY_VN',18,224,NULL,true,'2019-09-27 14:34:57.887','2019-09-27 14:34:57.887')
,('THPT Dân tộc Nội trú Sơn La','COUNTRY_VN',18,225,NULL,true,'2019-09-27 14:34:57.888','2019-09-27 14:34:57.888')
,('THPT Chiềng Sinh','COUNTRY_VN',18,225,NULL,true,'2019-09-27 14:34:57.888','2019-09-27 14:34:57.888')
,('THPT Chuyên Sơn La','COUNTRY_VN',18,225,NULL,true,'2019-09-27 14:34:57.889','2019-09-27 14:34:57.889')
,('THPT Nguyễn Du','COUNTRY_VN',18,225,NULL,true,'2019-09-27 14:34:57.889','2019-09-27 14:34:57.889')
,('THPT Tô Hiệu','COUNTRY_VN',18,225,NULL,true,'2019-09-27 14:34:57.889','2019-09-27 14:34:57.889')
,('TTGDTX Thành phố Sơn La','COUNTRY_VN',18,225,NULL,true,'2019-09-27 14:34:57.890','2019-09-27 14:34:57.890')
,('THCS-THPT An Ninh','COUNTRY_VN',19,226,NULL,true,'2019-09-27 14:34:57.893','2019-09-27 14:34:57.893')
,('THCS-THPT Mỹ Thuận','COUNTRY_VN',19,227,NULL,true,'2019-09-27 14:34:57.895','2019-09-27 14:34:57.895')
,('THCS-THPT Hưng Lợi','COUNTRY_VN',19,228,NULL,true,'2019-09-27 14:34:57.897','2019-09-27 14:34:57.897')
,('THCS-THPT Thạnh Tân','COUNTRY_VN',19,228,NULL,true,'2019-09-27 14:34:57.898','2019-09-27 14:34:57.898')
,('THCS-THPT Trần Đề','COUNTRY_VN',19,229,NULL,true,'2019-09-27 14:34:57.899','2019-09-27 14:34:57.899')
,('TTGDTX Tỉnh Sóc Trăng','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.900','2019-09-27 14:34:57.900')
,('THPT Phan Văn Hùng','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.901','2019-09-27 14:34:57.901')
,('THPT Phú Tâm','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.901','2019-09-27 14:34:57.901')
,('THPT Thiều Văn Chỏi','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.902','2019-09-27 14:34:57.902')
,('THPT Thuận Hòa','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.902','2019-09-27 14:34:57.902')
,('THPT Trần Văn Bảy','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.903','2019-09-27 14:34:57.903')
,('THPT Văn Ngọc Chính','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.903','2019-09-27 14:34:57.903')
,('THPT Vĩnh Hải','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.903','2019-09-27 14:34:57.903')
,('THPT Lịch Hội Thượng','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.904','2019-09-27 14:34:57.904')
,('THPT Lương Định của','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.904','2019-09-27 14:34:57.904')
,('THPT Mai Thanh Thế','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.905','2019-09-27 14:34:57.905')
,('THPT Mỹ Hương','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.905','2019-09-27 14:34:57.905')
,('THPT Mỹ Xuyên','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.905','2019-09-27 14:34:57.905')
,('THPT Ngọc Tố','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.906','2019-09-27 14:34:57.906')
,('THPT Nguyễn Khuyến','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.906','2019-09-27 14:34:57.906')
,('THPT Đại Ngãi','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.907','2019-09-27 14:34:57.907')
,('THPT Đoàn văn Tố','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.909','2019-09-27 14:34:57.909')
,('THPT Hòa Tú','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.911','2019-09-27 14:34:57.911')
,('THPT Hoàng Diệu','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.912','2019-09-27 14:34:57.912')
,('THPT Huỳnh Hữu Nghĩa','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.913','2019-09-27 14:34:57.913')
,('THPT Kế Sách','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.914','2019-09-27 14:34:57.914')
,('THPT Lẽ Văn Tám','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.915','2019-09-27 14:34:57.915')
,('THCS-THPT Thạnh Tân','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.916','2019-09-27 14:34:57.916')
,('THCS-THPT Trần Đề','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.917','2019-09-27 14:34:57.917')
,('THPT An Lạc Thôn','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.917','2019-09-27 14:34:57.917')
,('THPT An Ninh','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.918','2019-09-27 14:34:57.918')
,('THPT An Thạnh 3','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.918','2019-09-27 14:34:57.918')
,('THPT Chuyên Nguyễn Thị Minh Khai','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.919','2019-09-27 14:34:57.919')
,('THPT Dân tộc Nội trú Huỳnh Cương','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.920','2019-09-27 14:34:57.920')
,('THCS-THPT Dân tộc Nội trú Vĩnh châu','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.920','2019-09-27 14:34:57.920')
,('THCS-THPT Hưng Lợi','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.921','2019-09-27 14:34:57.921')
,('THCS-THPT iSchool Sóc Trăng','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.921','2019-09-27 14:34:57.921')
,('THCS-THPT Khánh Hoà','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.922','2019-09-27 14:34:57.922')
,('THCS-THPT Lai Hòa','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.922','2019-09-27 14:34:57.922')
,('THCS-THPT Lê Hồng Phong','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.922','2019-09-27 14:34:57.922')
,('THCS-THPT Mỹ Thuận','COUNTRY_VN',19,230,NULL,true,'2019-09-27 14:34:57.923','2019-09-27 14:34:57.923')
,('HCS-THPT Dân tộc Nội trú Vĩnh Châu','COUNTRY_VN',19,231,NULL,true,'2019-09-27 14:34:57.924','2019-09-27 14:34:57.924')
,('THCSC-THPT Khánh Hòa','COUNTRY_VN',19,231,NULL,true,'2019-09-27 14:34:57.927','2019-09-27 14:34:57.927')
,('THCS-THPT Lai Hòa','COUNTRY_VN',19,231,NULL,true,'2019-09-27 14:34:57.928','2019-09-27 14:34:57.928')
,('THPT Cam Lộ','COUNTRY_VN',20,232,NULL,true,'2019-09-27 14:34:57.932','2019-09-27 14:34:57.932')
,('THPT Chế Lan Viên','COUNTRY_VN',20,232,NULL,true,'2019-09-27 14:34:57.933','2019-09-27 14:34:57.933')
,('THPT Lê Thế Hiếu','COUNTRY_VN',20,232,NULL,true,'2019-09-27 14:34:57.934','2019-09-27 14:34:57.934')
,('THPT Tân Lâm','COUNTRY_VN',20,232,NULL,true,'2019-09-27 14:34:57.934','2019-09-27 14:34:57.934')
,('TTGDTX Huyện Cam Lộ','COUNTRY_VN',20,232,NULL,true,'2019-09-27 14:34:57.935','2019-09-27 14:34:57.935')
,('THPT Đakrông','COUNTRY_VN',20,233,NULL,true,'2019-09-27 14:34:57.936','2019-09-27 14:34:57.936')
,('THPT Số 2 Đakrông','COUNTRY_VN',20,233,NULL,true,'2019-09-27 14:34:57.937','2019-09-27 14:34:57.937')
,('TTGDTX Huyện Đakrông','COUNTRY_VN',20,233,NULL,true,'2019-09-27 14:34:57.937','2019-09-27 14:34:57.937')
,('THPT Cồn Tiên','COUNTRY_VN',20,234,NULL,true,'2019-09-27 14:34:57.938','2019-09-27 14:34:57.938')
,('THPT Gio Linh','COUNTRY_VN',20,234,NULL,true,'2019-09-27 14:34:57.939','2019-09-27 14:34:57.939')
,('THPT Nguyễn Du','COUNTRY_VN',20,234,NULL,true,'2019-09-27 14:34:57.939','2019-09-27 14:34:57.939')
,('TTGDTX Huyện Gio Linh','COUNTRY_VN',20,234,NULL,true,'2019-09-27 14:34:57.940','2019-09-27 14:34:57.940')
,('THPT Bùi Dục Tài','COUNTRY_VN',20,235,NULL,true,'2019-09-27 14:34:57.943','2019-09-27 14:34:57.943')
,('THPT Hải Lăng','COUNTRY_VN',20,235,NULL,true,'2019-09-27 14:34:57.945','2019-09-27 14:34:57.945')
,('THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',20,235,NULL,true,'2019-09-27 14:34:57.945','2019-09-27 14:34:57.945')
,('THPT Trần Thị Tâm','COUNTRY_VN',20,235,NULL,true,'2019-09-27 14:34:57.946','2019-09-27 14:34:57.946')
,('TTGDTX Huyện Hải Lăng','COUNTRY_VN',20,235,NULL,true,'2019-09-27 14:34:57.947','2019-09-27 14:34:57.947')
,('THPT A Túc','COUNTRY_VN',20,236,NULL,true,'2019-09-27 14:34:57.950','2019-09-27 14:34:57.950')
,('THPT Huớng Hoá','COUNTRY_VN',20,236,NULL,true,'2019-09-27 14:34:57.950','2019-09-27 14:34:57.950')
,('THPT Hướng Phùng','COUNTRY_VN',20,236,NULL,true,'2019-09-27 14:34:57.951','2019-09-27 14:34:57.951')
,('THPT Lao Bảo','COUNTRY_VN',20,236,NULL,true,'2019-09-27 14:34:57.951','2019-09-27 14:34:57.951')
,('TTGDTX Huyện Hướng Hóa','COUNTRY_VN',20,236,NULL,true,'2019-09-27 14:34:57.952','2019-09-27 14:34:57.952')
,('THPT Chu Văn An','COUNTRY_VN',20,237,NULL,true,'2019-09-27 14:34:57.953','2019-09-27 14:34:57.953')
,('THPT Nguyễn Hữu Thận','COUNTRY_VN',20,237,NULL,true,'2019-09-27 14:34:57.953','2019-09-27 14:34:57.953')
,('THPT Triệu Phong','COUNTRY_VN',20,237,NULL,true,'2019-09-27 14:34:57.954','2019-09-27 14:34:57.954')
,('THPT Vĩnh Định','COUNTRY_VN',20,237,NULL,true,'2019-09-27 14:34:57.954','2019-09-27 14:34:57.954')
,('TTGDTX Huyện Triệu Phong','COUNTRY_VN',20,237,NULL,true,'2019-09-27 14:34:57.954','2019-09-27 14:34:57.954')
,('THPT Bến Quan','COUNTRY_VN',20,238,NULL,true,'2019-09-27 14:34:57.955','2019-09-27 14:34:57.955')
,('THPT Cửa Tùng','COUNTRY_VN',20,238,NULL,true,'2019-09-27 14:34:57.956','2019-09-27 14:34:57.956')
,('THPT Nguyễn Công Trứ','COUNTRY_VN',20,238,NULL,true,'2019-09-27 14:34:57.956','2019-09-27 14:34:57.956')
,('THPT Vĩnh Linh','COUNTRY_VN',20,238,NULL,true,'2019-09-27 14:34:57.956','2019-09-27 14:34:57.956')
,('TTGDTX Huyện Vĩnh Linh','COUNTRY_VN',20,238,NULL,true,'2019-09-27 14:34:57.957','2019-09-27 14:34:57.957')
,('TH-THCS-THPT Trưng Vương','COUNTRY_VN',20,239,NULL,true,'2019-09-27 14:34:57.959','2019-09-27 14:34:57.959')
,('THPT Chuyên Lê Quý Đôn','COUNTRY_VN',20,239,NULL,true,'2019-09-27 14:34:57.960','2019-09-27 14:34:57.960')
,('THPT Đông Hà','COUNTRY_VN',20,239,NULL,true,'2019-09-27 14:34:57.960','2019-09-27 14:34:57.960')
,('THPT Lê Lợi','COUNTRY_VN',20,239,NULL,true,'2019-09-27 14:34:57.961','2019-09-27 14:34:57.961')
,('THPT Phan Châu Trinh','COUNTRY_VN',20,239,NULL,true,'2019-09-27 14:34:57.961','2019-09-27 14:34:57.961')
,('TTGDTX Thị xã Đông Hà','COUNTRY_VN',20,239,NULL,true,'2019-09-27 14:34:57.962','2019-09-27 14:34:57.962')
,('THPT Dân tộc Nội trú Quảng Trị','COUNTRY_VN',20,240,NULL,true,'2019-09-27 14:34:57.965','2019-09-27 14:34:57.965')
,('THPT Nguyễn Huệ','COUNTRY_VN',20,240,NULL,true,'2019-09-27 14:34:57.965','2019-09-27 14:34:57.965')
,('THPT Thị xã Quảng Trị','COUNTRY_VN',20,240,NULL,true,'2019-09-27 14:34:57.966','2019-09-27 14:34:57.966')
,('TTGDTX Thị xã Quảng Trị','COUNTRY_VN',20,240,NULL,true,'2019-09-27 14:34:57.967','2019-09-27 14:34:57.967')
,('THPT Ba Chẽ','COUNTRY_VN',21,241,NULL,true,'2019-09-27 14:34:57.968','2019-09-27 14:34:57.968')
,('TTGDTX Huyện Ba Chẽ','COUNTRY_VN',21,241,NULL,true,'2019-09-27 14:34:57.968','2019-09-27 14:34:57.968')
,('THCS-THPT Hoành Mô','COUNTRY_VN',21,242,NULL,true,'2019-09-27 14:34:57.969','2019-09-27 14:34:57.969')
,('THPT Bình Liêu','COUNTRY_VN',21,242,NULL,true,'2019-09-27 14:34:57.969','2019-09-27 14:34:57.969')
,('TTGDTX Huyện Bình Liêu','COUNTRY_VN',21,242,NULL,true,'2019-09-27 14:34:57.969','2019-09-27 14:34:57.969')
,('THPT Cô Tô','COUNTRY_VN',21,243,NULL,true,'2019-09-27 14:34:57.970','2019-09-27 14:34:57.970')
,('TTGDTX Huyện Cô Tô','COUNTRY_VN',21,243,NULL,true,'2019-09-27 14:34:57.971','2019-09-27 14:34:57.971')
,('THCS-THPT Lê Lợi','COUNTRY_VN',21,244,NULL,true,'2019-09-27 14:34:57.971','2019-09-27 14:34:57.971')
,('THPT Đầm Hà','COUNTRY_VN',21,244,NULL,true,'2019-09-27 14:34:57.972','2019-09-27 14:34:57.972')
,('TTGDTX Huyện Đầm Hà','COUNTRY_VN',21,244,NULL,true,'2019-09-27 14:34:57.972','2019-09-27 14:34:57.972')
,('THPT Hải Đảo','COUNTRY_VN',21,245,NULL,true,'2019-09-27 14:34:57.973','2019-09-27 14:34:57.973')
,('THPT Quan Lạn','COUNTRY_VN',21,245,NULL,true,'2019-09-27 14:34:57.973','2019-09-27 14:34:57.973')
,('THPT Trần Khánh Dư','COUNTRY_VN',21,245,NULL,true,'2019-09-27 14:34:57.974','2019-09-27 14:34:57.974')
,('TTGDTX Huyện đảo Vân Đồn','COUNTRY_VN',21,245,NULL,true,'2019-09-27 14:34:57.976','2019-09-27 14:34:57.976')
,('THCS-THPT Đường Hoa Cương','COUNTRY_VN',21,246,NULL,true,'2019-09-27 14:34:57.977','2019-09-27 14:34:57.977')
,('THPT Nguyễn Du','COUNTRY_VN',21,246,NULL,true,'2019-09-27 14:34:57.977','2019-09-27 14:34:57.977')
,('THPT Quảng Hà','COUNTRY_VN',21,246,NULL,true,'2019-09-27 14:34:57.978','2019-09-27 14:34:57.978')
,('TTGDTX Huyện Hải Hà','COUNTRY_VN',21,246,NULL,true,'2019-09-27 14:34:57.979','2019-09-27 14:34:57.979')
,('THPT Hoành Bồ','COUNTRY_VN',21,247,NULL,true,'2019-09-27 14:34:57.981','2019-09-27 14:34:57.981')
,('THPT Quảng La','COUNTRY_VN',21,247,NULL,true,'2019-09-27 14:34:57.982','2019-09-27 14:34:57.982')
,('THPT Thống Nhất','COUNTRY_VN',21,247,NULL,true,'2019-09-27 14:34:57.982','2019-09-27 14:34:57.982')
,('TTGDTX Huyện Hoành Bồ','COUNTRY_VN',21,247,NULL,true,'2019-09-27 14:34:57.983','2019-09-27 14:34:57.983')
,('THPT Dân tộc Nội trú Tiên Yên','COUNTRY_VN',21,248,NULL,true,'2019-09-27 14:34:57.984','2019-09-27 14:34:57.984')
,('THPT Hải Đông','COUNTRY_VN',21,248,NULL,true,'2019-09-27 14:34:57.984','2019-09-27 14:34:57.984')
,('THPT Nguyễn Trãi','COUNTRY_VN',21,248,NULL,true,'2019-09-27 14:34:57.985','2019-09-27 14:34:57.985')
,('THPT Tiên Yên','COUNTRY_VN',21,248,NULL,true,'2019-09-27 14:34:57.985','2019-09-27 14:34:57.985')
,('TTGDTX Huyện Tiên Yên','COUNTRY_VN',21,248,NULL,true,'2019-09-27 14:34:57.985','2019-09-27 14:34:57.985')
,('THPT Cẩm Phả','COUNTRY_VN',21,249,NULL,true,'2019-09-27 14:34:57.986','2019-09-27 14:34:57.986')
,('THPT Cửa Ông','COUNTRY_VN',21,249,NULL,true,'2019-09-27 14:34:57.986','2019-09-27 14:34:57.986')
,('THPT Hùng Vương','COUNTRY_VN',21,249,NULL,true,'2019-09-27 14:34:57.986','2019-09-27 14:34:57.986')
,('THPT Lê Hồng Phong','COUNTRY_VN',21,249,NULL,true,'2019-09-27 14:34:57.987','2019-09-27 14:34:57.987')
,('THPT Lê Quý Đôn','COUNTRY_VN',21,249,NULL,true,'2019-09-27 14:34:57.987','2019-09-27 14:34:57.987')
,('THPT Lương Thế vinh','COUNTRY_VN',21,249,NULL,true,'2019-09-27 14:34:57.987','2019-09-27 14:34:57.987')
,('THPT Mông Dương','COUNTRY_VN',21,249,NULL,true,'2019-09-27 14:34:57.988','2019-09-27 14:34:57.988')
,('TTGDTX Thành phố Cẩm Phả','COUNTRY_VN',21,249,NULL,true,'2019-09-27 14:34:57.988','2019-09-27 14:34:57.988')
,('THCS-THPT Lê Thánh Tông','COUNTRY_VN',21,250,NULL,true,'2019-09-27 14:34:57.989','2019-09-27 14:34:57.989')
,('THPT Bãi Cháy','COUNTRY_VN',21,250,NULL,true,'2019-09-27 14:34:57.989','2019-09-27 14:34:57.989')
,('THPT Chuyên Hạ Long','COUNTRY_VN',21,250,NULL,true,'2019-09-27 14:34:57.990','2019-09-27 14:34:57.990')
,('THPT Hạ Long','COUNTRY_VN',21,250,NULL,true,'2019-09-27 14:34:57.990','2019-09-27 14:34:57.990')
,('THPT Hòn Gai','COUNTRY_VN',21,250,NULL,true,'2019-09-27 14:34:57.991','2019-09-27 14:34:57.991')
,('THPT Ngô Quyền','COUNTRY_VN',21,250,NULL,true,'2019-09-27 14:34:57.992','2019-09-27 14:34:57.992')
,('THPT Nquyễn Bỉnh Khiêm','COUNTRY_VN',21,250,NULL,true,'2019-09-27 14:34:57.993','2019-09-27 14:34:57.993')
,('THPT Vũ Văn Hiếu','COUNTRY_VN',21,250,NULL,true,'2019-09-27 14:34:57.994','2019-09-27 14:34:57.994')
,('THPT Dân tộc Nội trú Quảng Ninh','COUNTRY_VN',21,250,NULL,true,'2019-09-27 14:34:57.995','2019-09-27 14:34:57.995')
,('TH-THCS-THPT Đoàn Thị Điểm','COUNTRY_VN',21,250,NULL,true,'2019-09-27 14:34:57.997','2019-09-27 14:34:57.997')
,('TH-THCS-THPT Văn Lang','COUNTRY_VN',21,250,NULL,true,'2019-09-27 14:34:57.998','2019-09-27 14:34:57.998')
,('TTGDTX Thành phố Hạ Long','COUNTRY_VN',21,250,NULL,true,'2019-09-27 14:34:57.999','2019-09-27 14:34:57.999')
,('THCS-THPT Chu Văn An','COUNTRY_VN',21,251,NULL,true,'2019-09-27 14:34:58.000','2019-09-27 14:34:58.000')
,('THPT Lý Thường Kiệt','COUNTRY_VN',21,251,NULL,true,'2019-09-27 14:34:58.001','2019-09-27 14:34:58.001')
,('THPT Trần Phú','COUNTRY_VN',21,251,NULL,true,'2019-09-27 14:34:58.001','2019-09-27 14:34:58.001')
,('TTGDTX Thành phố Móng Cái','COUNTRY_VN',21,251,NULL,true,'2019-09-27 14:34:58.002','2019-09-27 14:34:58.002')
,('THPT Hoàng Văn Thụ','COUNTRY_VN',21,252,NULL,true,'2019-09-27 14:34:58.003','2019-09-27 14:34:58.003')
,('THPT Hồng Đức','COUNTRY_VN',21,252,NULL,true,'2019-09-27 14:34:58.003','2019-09-27 14:34:58.003')
,('THPT Nguyễn Tất Thành','COUNTRY_VN',21,252,NULL,true,'2019-09-27 14:34:58.003','2019-09-27 14:34:58.003')
,('THPT Uông Bí','COUNTRY_VN',21,252,NULL,true,'2019-09-27 14:34:58.004','2019-09-27 14:34:58.004')
,('TTGDTX Thành phố Uông Bí','COUNTRY_VN',21,252,NULL,true,'2019-09-27 14:34:58.004','2019-09-27 14:34:58.004')
,('THCS-THPT Trần Nhân Tông','COUNTRY_VN',21,253,NULL,true,'2019-09-27 14:34:58.005','2019-09-27 14:34:58.005')
,('THCS-THPT Nguyễn Bình','COUNTRY_VN',21,253,NULL,true,'2019-09-27 14:34:58.005','2019-09-27 14:34:58.005')
,('THPT Đông Triều','COUNTRY_VN',21,253,NULL,true,'2019-09-27 14:34:58.006','2019-09-27 14:34:58.006')
,('THPT Hoàng Hoa Thám','COUNTRY_VN',21,253,NULL,true,'2019-09-27 14:34:58.006','2019-09-27 14:34:58.006')
,('THPT Hoàng Quốc Việt','COUNTRY_VN',21,253,NULL,true,'2019-09-27 14:34:58.006','2019-09-27 14:34:58.006')
,('THPT Lê Chân','COUNTRY_VN',21,253,NULL,true,'2019-09-27 14:34:58.007','2019-09-27 14:34:58.007')
,('TH-THCS-THPT Trần Hưng Đạo','COUNTRY_VN',21,253,NULL,true,'2019-09-27 14:34:58.007','2019-09-27 14:34:58.007')
,('THPT Bạch Đằng','COUNTRY_VN',21,254,NULL,true,'2019-09-27 14:34:58.011','2019-09-27 14:34:58.011')
,('THPT Đống Thành','COUNTRY_VN',21,254,NULL,true,'2019-09-27 14:34:58.012','2019-09-27 14:34:58.012')
,('THPT Minh Hà','COUNTRY_VN',21,254,NULL,true,'2019-09-27 14:34:58.012','2019-09-27 14:34:58.012')
,('THPT Ngô Gia Tự','COUNTRY_VN',21,254,NULL,true,'2019-09-27 14:34:58.013','2019-09-27 14:34:58.013')
,('THPT Trần Quốc Tuấn','COUNTRY_VN',21,254,NULL,true,'2019-09-27 14:34:58.014','2019-09-27 14:34:58.014')
,('THPT Yên Hưng','COUNTRY_VN',21,254,NULL,true,'2019-09-27 14:34:58.015','2019-09-27 14:34:58.015')
,('TTGDTX Thị xã Quảng Yên','COUNTRY_VN',21,254,NULL,true,'2019-09-27 14:34:58.016','2019-09-27 14:34:58.016')
,('THPT Ba Tơ','COUNTRY_VN',22,255,NULL,true,'2019-09-27 14:34:58.017','2019-09-27 14:34:58.017')
,('THPT Phạm Kiệt','COUNTRY_VN',22,255,NULL,true,'2019-09-27 14:34:58.018','2019-09-27 14:34:58.018')
,('TTGDTX Huyện Ba Tơ','COUNTRY_VN',22,255,NULL,true,'2019-09-27 14:34:58.018','2019-09-27 14:34:58.018')
,('THPT Bình Sơn','COUNTRY_VN',22,256,NULL,true,'2019-09-27 14:34:58.019','2019-09-27 14:34:58.019')
,('THPT Lê Quý Đôn','COUNTRY_VN',22,256,NULL,true,'2019-09-27 14:34:58.019','2019-09-27 14:34:58.019')
,('THPT Trần Kỳ Phong','COUNTRY_VN',22,256,NULL,true,'2019-09-27 14:34:58.020','2019-09-27 14:34:58.020')
,('THPT Vạn Tường','COUNTRY_VN',22,256,NULL,true,'2019-09-27 14:34:58.020','2019-09-27 14:34:58.020')
,('TTGDTX Huyện Bình Sơn','COUNTRY_VN',22,256,NULL,true,'2019-09-27 14:34:58.020','2019-09-27 14:34:58.020')
,('THPT Lý Sơn','COUNTRY_VN',22,257,NULL,true,'2019-09-27 14:34:58.021','2019-09-27 14:34:58.021')
,('TTGDTX Huyện đảo Lý Sơn','COUNTRY_VN',22,257,NULL,true,'2019-09-27 14:34:58.021','2019-09-27 14:34:58.021')
,('THPT Lương Thế vinh','COUNTRY_VN',22,258,NULL,true,'2019-09-27 14:34:58.022','2019-09-27 14:34:58.022')
,('THPT Số 1 Đức Phổ','COUNTRY_VN',22,258,NULL,true,'2019-09-27 14:34:58.022','2019-09-27 14:34:58.022')
,('THPT Số 2 Đức Phổ','COUNTRY_VN',22,258,NULL,true,'2019-09-27 14:34:58.023','2019-09-27 14:34:58.023')
,('TTGDTX Huyện Đức Phổ','COUNTRY_VN',22,258,NULL,true,'2019-09-27 14:34:58.023','2019-09-27 14:34:58.023')
,('THPT Minh Long','COUNTRY_VN',22,259,NULL,true,'2019-09-27 14:34:58.024','2019-09-27 14:34:58.024')
,('TTGDTX Huyện Minh Long','COUNTRY_VN',22,259,NULL,true,'2019-09-27 14:34:58.025','2019-09-27 14:34:58.025')
,('THPT Nguyễn Công Trứ','COUNTRY_VN',22,260,NULL,true,'2019-09-27 14:34:58.027','2019-09-27 14:34:58.027')
,('THPT Phạm Văn Đồng','COUNTRY_VN',22,260,NULL,true,'2019-09-27 14:34:58.027','2019-09-27 14:34:58.027')
,('THPT Số 2 Mộ Đức','COUNTRY_VN',22,260,NULL,true,'2019-09-27 14:34:58.028','2019-09-27 14:34:58.028')
,('THPT Trần Quang Diệu','COUNTRY_VN',22,260,NULL,true,'2019-09-27 14:34:58.029','2019-09-27 14:34:58.029')
,('TTGDTX Huyện Mộ Đức','COUNTRY_VN',22,260,NULL,true,'2019-09-27 14:34:58.030','2019-09-27 14:34:58.030')
,('THP Nguyễn Công Phương','COUNTRY_VN',22,261,NULL,true,'2019-09-27 14:34:58.032','2019-09-27 14:34:58.032')
,('THPT Số 1 Nghĩa Hành','COUNTRY_VN',22,261,NULL,true,'2019-09-27 14:34:58.033','2019-09-27 14:34:58.033')
,('THPT Số 2 Nghĩa Hành','COUNTRY_VN',22,261,NULL,true,'2019-09-27 14:34:58.033','2019-09-27 14:34:58.033')
,('TTGDTX Huyện Nghĩa Hành','COUNTRY_VN',22,261,NULL,true,'2019-09-27 14:34:58.033','2019-09-27 14:34:58.033')
,('THCS-THPT Phạm Kiệt','COUNTRY_VN',22,262,NULL,true,'2019-09-27 14:34:58.034','2019-09-27 14:34:58.034')
,('THPT Quang Trung','COUNTRY_VN',22,262,NULL,true,'2019-09-27 14:34:58.035','2019-09-27 14:34:58.035')
,('THPT Sơn Hà','COUNTRY_VN',22,262,NULL,true,'2019-09-27 14:34:58.035','2019-09-27 14:34:58.035')
,('TTGDTX Huyện Sơn Hà','COUNTRY_VN',22,262,NULL,true,'2019-09-27 14:34:58.035','2019-09-27 14:34:58.035')
,('THPT Đinh Tiên Hoàng','COUNTRY_VN',22,263,NULL,true,'2019-09-27 14:34:58.036','2019-09-27 14:34:58.036')
,('TTGDTX Huyện Sơn Tây','COUNTRY_VN',22,263,NULL,true,'2019-09-27 14:34:58.036','2019-09-27 14:34:58.036')
,('THPT Ba Gia','COUNTRY_VN',22,264,NULL,true,'2019-09-27 14:34:58.037','2019-09-27 14:34:58.037')
,('THPT Tư thục Trương Định','COUNTRY_VN',22,264,NULL,true,'2019-09-27 14:34:58.037','2019-09-27 14:34:58.037')
,('THPT Tây Trà','COUNTRY_VN',22,264,NULL,true,'2019-09-27 14:34:58.038','2019-09-27 14:34:58.038')
,('TTGDTX Huyện Tây Trà','COUNTRY_VN',22,264,NULL,true,'2019-09-27 14:34:58.038','2019-09-27 14:34:58.038')
,('THPT Trà Bồng','COUNTRY_VN',22,265,NULL,true,'2019-09-27 14:34:58.039','2019-09-27 14:34:58.039')
,('TTGDTX Huyện Trà Bồng','COUNTRY_VN',22,265,NULL,true,'2019-09-27 14:34:58.039','2019-09-27 14:34:58.039')
,('THPT Chu Văn An','COUNTRY_VN',22,266,NULL,true,'2019-09-27 14:34:58.040','2019-09-27 14:34:58.040')
,('THPT Số 1 Tư Nghĩa','COUNTRY_VN',22,266,NULL,true,'2019-09-27 14:34:58.040','2019-09-27 14:34:58.040')
,('THPT Số 2 Tư Nghía','COUNTRY_VN',22,266,NULL,true,'2019-09-27 14:34:58.042','2019-09-27 14:34:58.042')
,('THPT Thu Xà','COUNTRY_VN',22,266,NULL,true,'2019-09-27 14:34:58.043','2019-09-27 14:34:58.043')
,('TTGDTX Huyện Tư Nghĩa','COUNTRY_VN',22,266,NULL,true,'2019-09-27 14:34:58.043','2019-09-27 14:34:58.043')
,('THPT Chuyên Lê Khiết','COUNTRY_VN',22,267,NULL,true,'2019-09-27 14:34:58.044','2019-09-27 14:34:58.044')
,('THPT Dân tộc Nội trú Quảng Ngãi','COUNTRY_VN',22,267,NULL,true,'2019-09-27 14:34:58.045','2019-09-27 14:34:58.045')
,('THPT Huỳnh Thúc Kháng','COUNTRY_VN',22,267,NULL,true,'2019-09-27 14:34:58.045','2019-09-27 14:34:58.045')
,('THPT Lê Trung Đình','COUNTRY_VN',22,267,NULL,true,'2019-09-27 14:34:58.046','2019-09-27 14:34:58.046')
,('THPT Sơn Mỹ','COUNTRY_VN',22,267,NULL,true,'2019-09-27 14:34:58.047','2019-09-27 14:34:58.047')
,('THPT Trần Quốc Tuấn','COUNTRY_VN',22,267,NULL,true,'2019-09-27 14:34:58.048','2019-09-27 14:34:58.048')
,('THPT Tư thục Hoàng Văn Thụ','COUNTRY_VN',22,267,NULL,true,'2019-09-27 14:34:58.049','2019-09-27 14:34:58.049')
,('THPT Tư thục Nguyễn Bỉnh Khiêm','COUNTRY_VN',22,267,NULL,true,'2019-09-27 14:34:58.049','2019-09-27 14:34:58.049')
,('THPT Võ Nguyên Giáp','COUNTRY_VN',22,267,NULL,true,'2019-09-27 14:34:58.050','2019-09-27 14:34:58.050')
,('TTGDTX Huyện Sơn Tịnh','COUNTRY_VN',22,267,NULL,true,'2019-09-27 14:34:58.050','2019-09-27 14:34:58.050')
,('TTGDTX Tỉnh Quảng Ngãi','COUNTRY_VN',22,267,NULL,true,'2019-09-27 14:34:58.051','2019-09-27 14:34:58.051')
,('THPT Dân tộc Nội trú Nước Oa','COUNTRY_VN',23,268,NULL,true,'2019-09-27 14:34:58.052','2019-09-27 14:34:58.052')
,('THPT Bắc Trà My','COUNTRY_VN',23,268,NULL,true,'2019-09-27 14:34:58.052','2019-09-27 14:34:58.052')
,('TTGDTX Huyện Bắc Trà My','COUNTRY_VN',23,268,NULL,true,'2019-09-27 14:34:58.053','2019-09-27 14:34:58.053')
,('THPT Lê Hồng Phong','COUNTRY_VN',23,269,NULL,true,'2019-09-27 14:34:58.054','2019-09-27 14:34:58.054')
,('THPT Nguyễn Hiền','COUNTRY_VN',23,269,NULL,true,'2019-09-27 14:34:58.054','2019-09-27 14:34:58.054')
,('THPT Sào Nam','COUNTRY_VN',23,269,NULL,true,'2019-09-27 14:34:58.055','2019-09-27 14:34:58.055')
,('TTGDTX Huyện Duy Xuyên','COUNTRY_VN',23,269,NULL,true,'2019-09-27 14:34:58.055','2019-09-27 14:34:58.055')
,('THPT Chu Văn An','COUNTRY_VN',23,270,NULL,true,'2019-09-27 14:34:58.056','2019-09-27 14:34:58.056')
,('THPT Đỗ Văn Tuyển','COUNTRY_VN',23,270,NULL,true,'2019-09-27 14:34:58.057','2019-09-27 14:34:58.057')
,('THPT Huỳnh Ngọc Huệ','COUNTRY_VN',23,270,NULL,true,'2019-09-27 14:34:58.057','2019-09-27 14:34:58.057')
,('THPT Lương Thúc Kỳ','COUNTRY_VN',23,270,NULL,true,'2019-09-27 14:34:58.059','2019-09-27 14:34:58.059')
,('TTGDTX Huyện Đại Lộc','COUNTRY_VN',23,270,NULL,true,'2019-09-27 14:34:58.060','2019-09-27 14:34:58.060')
,('TH-THCS-THPT Quảng Đông','COUNTRY_VN',23,271,NULL,true,'2019-09-27 14:34:58.061','2019-09-27 14:34:58.061')
,('THPT Hoàng Diệu','COUNTRY_VN',23,271,NULL,true,'2019-09-27 14:34:58.061','2019-09-27 14:34:58.061')
,('THPT Lương Thế Vinh','COUNTRY_VN',23,271,NULL,true,'2019-09-27 14:34:58.062','2019-09-27 14:34:58.062')
,('THPT Nguyễn Duy Hiệu','COUNTRY_VN',23,271,NULL,true,'2019-09-27 14:34:58.064','2019-09-27 14:34:58.064')
,('THPT Nguyễn Khuyến','COUNTRY_VN',23,271,NULL,true,'2019-09-27 14:34:58.065','2019-09-27 14:34:58.065')
,('THPT Phạm Phú Thứ','COUNTRY_VN',23,271,NULL,true,'2019-09-27 14:34:58.066','2019-09-27 14:34:58.066')
,('TH-THCS-THPT Hoàng Sa','COUNTRY_VN',23,271,NULL,true,'2019-09-27 14:34:58.066','2019-09-27 14:34:58.066')
,('TTGDTX Huyện Điện Bàn','COUNTRY_VN',23,271,NULL,true,'2019-09-27 14:34:58.067','2019-09-27 14:34:58.067')
,('THPT Âu Cơ','COUNTRY_VN',23,272,NULL,true,'2019-09-27 14:34:58.068','2019-09-27 14:34:58.068')
,('THPT Quang Trung','COUNTRY_VN',23,272,NULL,true,'2019-09-27 14:34:58.068','2019-09-27 14:34:58.068')
,('THPT Hiệp Đức','COUNTRY_VN',23,273,NULL,true,'2019-09-27 14:34:58.069','2019-09-27 14:34:58.069')
,('THPT Trần Phú','COUNTRY_VN',23,273,NULL,true,'2019-09-27 14:34:58.069','2019-09-27 14:34:58.069')
,('TTGDTX Huyện Hiệp Đức','COUNTRY_VN',23,273,NULL,true,'2019-09-27 14:34:58.070','2019-09-27 14:34:58.070')
,('THPT Nam Giang','COUNTRY_VN',23,274,NULL,true,'2019-09-27 14:34:58.071','2019-09-27 14:34:58.071')
,('THPT Nguyễn Văn Trỗi','COUNTRY_VN',23,274,NULL,true,'2019-09-27 14:34:58.071','2019-09-27 14:34:58.071')
,('TTGDTX Huyện Nam Giang','COUNTRY_VN',23,274,NULL,true,'2019-09-27 14:34:58.072','2019-09-27 14:34:58.072')
,('THPT Nam Trà My','COUNTRY_VN',23,275,NULL,true,'2019-09-27 14:34:58.073','2019-09-27 14:34:58.073')
,('TTGDTX Huyện Nam Trà My','COUNTRY_VN',23,275,NULL,true,'2019-09-27 14:34:58.073','2019-09-27 14:34:58.073')
,('THPT Nông Sơn','COUNTRY_VN',23,276,NULL,true,'2019-09-27 14:34:58.075','2019-09-27 14:34:58.075')
,('THPT Cao Bá Quát','COUNTRY_VN',23,277,NULL,true,'2019-09-27 14:34:58.077','2019-09-27 14:34:58.077')
,('THPT Nguyễn Huệ','COUNTRY_VN',23,277,NULL,true,'2019-09-27 14:34:58.077','2019-09-27 14:34:58.077')
,('THPT Núi Thành','COUNTRY_VN',23,277,NULL,true,'2019-09-27 14:34:58.078','2019-09-27 14:34:58.078')
,('TTGDTX Huyện Núi Thành','COUNTRY_VN',23,277,NULL,true,'2019-09-27 14:34:58.080','2019-09-27 14:34:58.080')
,('THPT Nguyễn Dục','COUNTRY_VN',23,278,NULL,true,'2019-09-27 14:34:58.082','2019-09-27 14:34:58.082')
,('THPT Trần Văn Dư','COUNTRY_VN',23,278,NULL,true,'2019-09-27 14:34:58.083','2019-09-27 14:34:58.083')
,('TTGDTX Huyện Phú Ninh','COUNTRY_VN',23,278,NULL,true,'2019-09-27 14:34:58.084','2019-09-27 14:34:58.084')
,('THPT Khâm Đức','COUNTRY_VN',23,279,NULL,true,'2019-09-27 14:34:58.085','2019-09-27 14:34:58.085')
,('TTGDTX Huyện Phước Sơn','COUNTRY_VN',23,279,NULL,true,'2019-09-27 14:34:58.086','2019-09-27 14:34:58.086')
,('THPT Dân lập Phạm Văn Đồng','COUNTRY_VN',23,279,NULL,true,'2019-09-27 14:34:58.086','2019-09-27 14:34:58.086')
,('THPT Nguyễn Văn Cừ','COUNTRY_VN',23,280,NULL,true,'2019-09-27 14:34:58.087','2019-09-27 14:34:58.087')
,('THPT Quế Sơn','COUNTRY_VN',23,280,NULL,true,'2019-09-27 14:34:58.088','2019-09-27 14:34:58.088')
,('THPT Trần Đại Nghĩa','COUNTRY_VN',23,280,NULL,true,'2019-09-27 14:34:58.088','2019-09-27 14:34:58.088')
,('TTGDTX Huyện Quế Sơn','COUNTRY_VN',23,280,NULL,true,'2019-09-27 14:34:58.089','2019-09-27 14:34:58.089')
,('THPT Tây Giang','COUNTRY_VN',23,281,NULL,true,'2019-09-27 14:34:58.090','2019-09-27 14:34:58.090')
,('THPT Hùng Vương','COUNTRY_VN',23,282,NULL,true,'2019-09-27 14:34:58.092','2019-09-27 14:34:58.092')
,('THPT Lý Tự Trọng','COUNTRY_VN',23,282,NULL,true,'2019-09-27 14:34:58.093','2019-09-27 14:34:58.093')
,('THPT Nguyễn Thái Bình','COUNTRY_VN',23,282,NULL,true,'2019-09-27 14:34:58.094','2019-09-27 14:34:58.094')
,('THPT Thái Phiên','COUNTRY_VN',23,282,NULL,true,'2019-09-27 14:34:58.094','2019-09-27 14:34:58.094')
,('THPT Tiểu La','COUNTRY_VN',23,282,NULL,true,'2019-09-27 14:34:58.095','2019-09-27 14:34:58.095')
,('TTGDTX Huyện Thăng Bình','COUNTRY_VN',23,282,NULL,true,'2019-09-27 14:34:58.096','2019-09-27 14:34:58.096')
,('THPT Huỳnh Thúc Kháng','COUNTRY_VN',23,283,NULL,true,'2019-09-27 14:34:58.098','2019-09-27 14:34:58.098')
,('THPT Phan Châu Trinh','COUNTRY_VN',23,283,NULL,true,'2019-09-27 14:34:58.099','2019-09-27 14:34:58.099')
,('TTGDTX Huyện Tiên Phước','COUNTRY_VN',23,283,NULL,true,'2019-09-27 14:34:58.099','2019-09-27 14:34:58.099')
,('THPT Dân tộc Nội trú Quảng Nam','COUNTRY_VN',23,284,NULL,true,'2019-09-27 14:34:58.102','2019-09-27 14:34:58.102')
,('THPT Chuyên Lê Thánh Tông','COUNTRY_VN',23,284,NULL,true,'2019-09-27 14:34:58.102','2019-09-27 14:34:58.102')
,('THPT Nguyễn Trãi','COUNTRY_VN',23,284,NULL,true,'2019-09-27 14:34:58.103','2019-09-27 14:34:58.103')
,('THPT Trần Hưng Đạo','COUNTRY_VN',23,284,NULL,true,'2019-09-27 14:34:58.103','2019-09-27 14:34:58.103')
,('THPT Trần Quý Cáp','COUNTRY_VN',23,284,NULL,true,'2019-09-27 14:34:58.104','2019-09-27 14:34:58.104')
,('TTGDTX Thành phố Hội An','COUNTRY_VN',23,284,NULL,true,'2019-09-27 14:34:58.104','2019-09-27 14:34:58.104')
,('THPT Chuyên Nguyễn Bỉnh Khiêm','COUNTRY_VN',23,285,NULL,true,'2019-09-27 14:34:58.105','2019-09-27 14:34:58.105')
,('THPT Dân lập Hà Huy Tập','COUNTRY_VN',23,285,NULL,true,'2019-09-27 14:34:58.105','2019-09-27 14:34:58.105')
,('THPT Duy Tân','COUNTRY_VN',23,285,NULL,true,'2019-09-27 14:34:58.106','2019-09-27 14:34:58.106')
,('THPT Lê Quý Đôn','COUNTRY_VN',23,285,NULL,true,'2019-09-27 14:34:58.106','2019-09-27 14:34:58.106')
,('THPT Phan Bội Châu','COUNTRY_VN',23,285,NULL,true,'2019-09-27 14:34:58.106','2019-09-27 14:34:58.106')
,('THPT Trần Cao Vân','COUNTRY_VN',23,285,NULL,true,'2019-09-27 14:34:58.107','2019-09-27 14:34:58.107')
,('TTGDTX Tỉnh Quảng Nam','COUNTRY_VN',23,285,NULL,true,'2019-09-27 14:34:58.107','2019-09-27 14:34:58.107')
,('THCS-THPT Việt Trung','COUNTRY_VN',24,286,NULL,true,'2019-09-27 14:34:58.112','2019-09-27 14:34:58.112')
,('THPT Số 1 Bố Trạch','COUNTRY_VN',24,286,NULL,true,'2019-09-27 14:34:58.112','2019-09-27 14:34:58.112')
,('THPT Số 2 Bố Trạch','COUNTRY_VN',24,286,NULL,true,'2019-09-27 14:34:58.114','2019-09-27 14:34:58.114')
,('THPT SỐ 3 Bố Trạch','COUNTRY_VN',24,286,NULL,true,'2019-09-27 14:34:58.115','2019-09-27 14:34:58.115')
,('THPT Số 4 Bố trạch','COUNTRY_VN',24,286,NULL,true,'2019-09-27 14:34:58.116','2019-09-27 14:34:58.116')
,('THPT SỐ 5 Bố Trạch','COUNTRY_VN',24,286,NULL,true,'2019-09-27 14:34:58.117','2019-09-27 14:34:58.117')
,('TTGDTX Huyện Bố Trạch','COUNTRY_VN',24,286,NULL,true,'2019-09-27 14:34:58.117','2019-09-27 14:34:58.117')
,('THCS-THPT Dương Văn An','COUNTRY_VN',24,287,NULL,true,'2019-09-27 14:34:58.118','2019-09-27 14:34:58.118')
,('THPT Hoàng Hoa Thám','COUNTRY_VN',24,287,NULL,true,'2019-09-27 14:34:58.119','2019-09-27 14:34:58.119')
,('THPT KT Lệ Thuỷ','COUNTRY_VN',24,287,NULL,true,'2019-09-27 14:34:58.120','2019-09-27 14:34:58.120')
,('THPT Lệ Thuỷ','COUNTRY_VN',24,287,NULL,true,'2019-09-27 14:34:58.120','2019-09-27 14:34:58.120')
,('THPT Nguyễn Chí Thanh','COUNTRY_VN',24,287,NULL,true,'2019-09-27 14:34:58.120','2019-09-27 14:34:58.120')
,('THPT Trần Hưng Đạo','COUNTRY_VN',24,287,NULL,true,'2019-09-27 14:34:58.121','2019-09-27 14:34:58.121')
,('TTGDTX Huyện Lệ Thủy','COUNTRY_VN',24,287,NULL,true,'2019-09-27 14:34:58.121','2019-09-27 14:34:58.121')
,('THCS-THPT Hóa Tiến','COUNTRY_VN',24,288,NULL,true,'2019-09-27 14:34:58.122','2019-09-27 14:34:58.122')
,('THCS-THPT Trung Hóa','COUNTRY_VN',24,288,NULL,true,'2019-09-27 14:34:58.122','2019-09-27 14:34:58.122')
,('THPT Minh Hóa','COUNTRY_VN',24,288,NULL,true,'2019-09-27 14:34:58.123','2019-09-27 14:34:58.123')
,('TTGDTX Huyện Minh Hóa','COUNTRY_VN',24,288,NULL,true,'2019-09-27 14:34:58.123','2019-09-27 14:34:58.123')
,('THPT Nguyễn Hữu Cảnh','COUNTRY_VN',24,289,NULL,true,'2019-09-27 14:34:58.124','2019-09-27 14:34:58.124')
,('THPT Ninh Châu','COUNTRY_VN',24,289,NULL,true,'2019-09-27 14:34:58.126','2019-09-27 14:34:58.126')
,('THPT Quảng Ninh','COUNTRY_VN',24,289,NULL,true,'2019-09-27 14:34:58.127','2019-09-27 14:34:58.127')
,('TTGDTX Huyện Quảng Ninh','COUNTRY_VN',24,289,NULL,true,'2019-09-27 14:34:58.127','2019-09-27 14:34:58.127')
,('THPT Quang Trung','COUNTRY_VN',24,290,NULL,true,'2019-09-27 14:34:58.129','2019-09-27 14:34:58.129')
,('TTGDTX Huyện Quảng Trạch','COUNTRY_VN',24,290,NULL,true,'2019-09-27 14:34:58.130','2019-09-27 14:34:58.130')
,('THCS-THPT Bắc Sơn','COUNTRY_VN',24,291,NULL,true,'2019-09-27 14:34:58.132','2019-09-27 14:34:58.132')
,('THPT Lê Trực','COUNTRY_VN',24,291,NULL,true,'2019-09-27 14:34:58.133','2019-09-27 14:34:58.133')
,('THPT Phan Bội Châu','COUNTRY_VN',24,291,NULL,true,'2019-09-27 14:34:58.133','2019-09-27 14:34:58.133')
,('THPT Tuyên Hóa','COUNTRY_VN',24,291,NULL,true,'2019-09-27 14:34:58.134','2019-09-27 14:34:58.134')
,('TTGDTX Huyện Tuyên Hóa','COUNTRY_VN',24,291,NULL,true,'2019-09-27 14:34:58.134','2019-09-27 14:34:58.134')
,('THPT Dân tộc Nội trú Quảng Bình','COUNTRY_VN',24,292,NULL,true,'2019-09-27 14:34:58.135','2019-09-27 14:34:58.135')
,('THPT Chuyên Võ Nguyên Giáp','COUNTRY_VN',24,292,NULL,true,'2019-09-27 14:34:58.136','2019-09-27 14:34:58.136')
,('THPT Đào Duy Từ','COUNTRY_VN',24,292,NULL,true,'2019-09-27 14:34:58.136','2019-09-27 14:34:58.136')
,('THPT Đồng Hới','COUNTRY_VN',24,292,NULL,true,'2019-09-27 14:34:58.137','2019-09-27 14:34:58.137')
,('THPT Phan Đình Phùng','COUNTRY_VN',24,292,NULL,true,'2019-09-27 14:34:58.137','2019-09-27 14:34:58.137')
,('TTGDTX Thành phố Đồng Hới','COUNTRY_VN',24,292,NULL,true,'2019-09-27 14:34:58.138','2019-09-27 14:34:58.138')
,('THPT Lê Hồng Phong','COUNTRY_VN',24,293,NULL,true,'2019-09-27 14:34:58.138','2019-09-27 14:34:58.138')
,('THPT Lê Lợi','COUNTRY_VN',24,293,NULL,true,'2019-09-27 14:34:58.139','2019-09-27 14:34:58.139')
,('THPT Lương Thế Vinh','COUNTRY_VN',24,293,NULL,true,'2019-09-27 14:34:58.139','2019-09-27 14:34:58.139')
,('THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',24,293,NULL,true,'2019-09-27 14:34:58.140','2019-09-27 14:34:58.140')
,('TTGDTX Thị xã Ba Đồn','COUNTRY_VN',24,293,NULL,true,'2019-09-27 14:34:58.141','2019-09-27 14:34:58.141')
,('THPT Bán công Cẩm Khê','COUNTRY_VN',25,294,NULL,true,'2019-09-27 14:34:58.145','2019-09-27 14:34:58.145')
,('THPT Cẩm Khê','COUNTRY_VN',25,294,NULL,true,'2019-09-27 14:34:58.146','2019-09-27 14:34:58.146')
,('THPT Hiền Đa','COUNTRY_VN',25,294,NULL,true,'2019-09-27 14:34:58.147','2019-09-27 14:34:58.147')
,('THPT Phương Xá','COUNTRY_VN',25,294,NULL,true,'2019-09-27 14:34:58.148','2019-09-27 14:34:58.148')
,('TTGDTX Huyện Cẩm Khê','COUNTRY_VN',25,294,NULL,true,'2019-09-27 14:34:58.149','2019-09-27 14:34:58.149')
,('THPT Bán công Đoan Hùng','COUNTRY_VN',25,295,NULL,true,'2019-09-27 14:34:58.150','2019-09-27 14:34:58.150')
,('THPT Chân Mộng','COUNTRY_VN',25,295,NULL,true,'2019-09-27 14:34:58.151','2019-09-27 14:34:58.151')
,('THPT Đoan Hùng','COUNTRY_VN',25,295,NULL,true,'2019-09-27 14:34:58.152','2019-09-27 14:34:58.152')
,('THPT Quế Lâm','COUNTRY_VN',25,295,NULL,true,'2019-09-27 14:34:58.152','2019-09-27 14:34:58.152')
,('TTGDTX Huyện Đoan Hùng','COUNTRY_VN',25,295,NULL,true,'2019-09-27 14:34:58.153','2019-09-27 14:34:58.153')
,('THPT Hạ Hoà','COUNTRY_VN',25,296,NULL,true,'2019-09-27 14:34:58.154','2019-09-27 14:34:58.154')
,('THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',25,296,NULL,true,'2019-09-27 14:34:58.154','2019-09-27 14:34:58.154')
,('THPT Vĩnh chân','COUNTRY_VN',25,296,NULL,true,'2019-09-27 14:34:58.155','2019-09-27 14:34:58.155')
,('THPT Xuân Áng','COUNTRY_VN',25,296,NULL,true,'2019-09-27 14:34:58.155','2019-09-27 14:34:58.155')
,('TTGDTX Huyện Hạ Hòa','COUNTRY_VN',25,296,NULL,true,'2019-09-27 14:34:58.155','2019-09-27 14:34:58.155')
,('THPT Bán công Phong Châu','COUNTRY_VN',25,297,NULL,true,'2019-09-27 14:34:58.156','2019-09-27 14:34:58.156')
,('THPT Lâm Thao','COUNTRY_VN',25,297,NULL,true,'2019-09-27 14:34:58.156','2019-09-27 14:34:58.156')
,('THPT Long Châu Sa','COUNTRY_VN',25,297,NULL,true,'2019-09-27 14:34:58.157','2019-09-27 14:34:58.157')
,('THPT Phong Châu','COUNTRY_VN',25,297,NULL,true,'2019-09-27 14:34:58.158','2019-09-27 14:34:58.158')
,('TTGDTX Huyện Lâm Thao','COUNTRY_VN',25,297,NULL,true,'2019-09-27 14:34:58.159','2019-09-27 14:34:58.159')
,('THPT Bán công Phù Ninh','COUNTRY_VN',25,298,NULL,true,'2019-09-27 14:34:58.160','2019-09-27 14:34:58.160')
,('THPT Nguyễn Huệ','COUNTRY_VN',25,298,NULL,true,'2019-09-27 14:34:58.161','2019-09-27 14:34:58.161')
,('THPT Phan Đăng Luu','COUNTRY_VN',25,298,NULL,true,'2019-09-27 14:34:58.161','2019-09-27 14:34:58.161')
,('THPT Phù Ninh','COUNTRY_VN',25,298,NULL,true,'2019-09-27 14:34:58.162','2019-09-27 14:34:58.162')
,('THPT Trung Giáp','COUNTRY_VN',25,298,NULL,true,'2019-09-27 14:34:58.163','2019-09-27 14:34:58.163')
,('THPT Tử Đà','COUNTRY_VN',25,298,NULL,true,'2019-09-27 14:34:58.164','2019-09-27 14:34:58.164')
,('TTGDTX Huyện Phù Ninh','COUNTRY_VN',25,298,NULL,true,'2019-09-27 14:34:58.165','2019-09-27 14:34:58.165')
,('THPT Bán công Tam Nông','COUNTRY_VN',25,299,NULL,true,'2019-09-27 14:34:58.166','2019-09-27 14:34:58.166')
,('THPT Hưng Hoá','COUNTRY_VN',25,299,NULL,true,'2019-09-27 14:34:58.166','2019-09-27 14:34:58.166')
,('THPT Mỹ Văn','COUNTRY_VN',25,299,NULL,true,'2019-09-27 14:34:58.167','2019-09-27 14:34:58.167')
,('THPT Tam Nông','COUNTRY_VN',25,299,NULL,true,'2019-09-27 14:34:58.167','2019-09-27 14:34:58.167')
,('TTGDTX Huyện Tam Nông','COUNTRY_VN',25,299,NULL,true,'2019-09-27 14:34:58.167','2019-09-27 14:34:58.167')
,('THPT Minh Đài','COUNTRY_VN',25,300,NULL,true,'2019-09-27 14:34:58.168','2019-09-27 14:34:58.168')
,('THPT Thạch Kiệ','COUNTRY_VN',25,300,NULL,true,'2019-09-27 14:34:58.168','2019-09-27 14:34:58.168')
,('TTGDTX Huyện Tân Sơn','COUNTRY_VN',25,300,NULL,true,'2019-09-27 14:34:58.169','2019-09-27 14:34:58.169')
,('THPT Bán công Thanh Ba','COUNTRY_VN',25,301,NULL,true,'2019-09-27 14:34:58.169','2019-09-27 14:34:58.169')
,('THPT Thanh Ba','COUNTRY_VN',25,301,NULL,true,'2019-09-27 14:34:58.170','2019-09-27 14:34:58.170')
,('THPT Yển Khê','COUNTRY_VN',25,301,NULL,true,'2019-09-27 14:34:58.170','2019-09-27 14:34:58.170')
,('TTGDTX Huyện Thanh Ba','COUNTRY_VN',25,301,NULL,true,'2019-09-27 14:34:58.170','2019-09-27 14:34:58.170')
,('THPT Bán công Thanh Sơn','COUNTRY_VN',25,302,NULL,true,'2019-09-27 14:34:58.171','2019-09-27 14:34:58.171')
,('THPT Hương Cần','COUNTRY_VN',25,302,NULL,true,'2019-09-27 14:34:58.172','2019-09-27 14:34:58.172')
,('THPT Thanh Sơn','COUNTRY_VN',25,302,NULL,true,'2019-09-27 14:34:58.172','2019-09-27 14:34:58.172')
,('THPT Văn Miếu','COUNTRY_VN',25,302,NULL,true,'2019-09-27 14:34:58.172','2019-09-27 14:34:58.172')
,('TTGDTX Huyện Thanh Sơn','COUNTRY_VN',25,302,NULL,true,'2019-09-27 14:34:58.173','2019-09-27 14:34:58.173')
,('THPT Tản Đà','COUNTRY_VN',25,303,NULL,true,'2019-09-27 14:34:58.173','2019-09-27 14:34:58.173')
,('THPT Thanh Thủy','COUNTRY_VN',25,303,NULL,true,'2019-09-27 14:34:58.174','2019-09-27 14:34:58.174')
,('THPT Trungg Nghĩa','COUNTRY_VN',25,303,NULL,true,'2019-09-27 14:34:58.175','2019-09-27 14:34:58.175')
,('TTGDTX Huyện Thanh Thủy','COUNTRY_VN',25,303,NULL,true,'2019-09-27 14:34:58.176','2019-09-27 14:34:58.176')
,('THPT Lương Sơn','COUNTRY_VN',25,304,NULL,true,'2019-09-27 14:34:58.178','2019-09-27 14:34:58.178')
,('THPT Minh Hòa','COUNTRY_VN',25,304,NULL,true,'2019-09-27 14:34:58.179','2019-09-27 14:34:58.179')
,('THPT Yên Lập','COUNTRY_VN',25,304,NULL,true,'2019-09-27 14:34:58.181','2019-09-27 14:34:58.181')
,('TTGDTX Huyện Yên Lập','COUNTRY_VN',25,304,NULL,true,'2019-09-27 14:34:58.182','2019-09-27 14:34:58.182')
,('THPT Bán công Công nghiệp Việt Trì','COUNTRY_VN',25,305,NULL,true,'2019-09-27 14:34:58.183','2019-09-27 14:34:58.183')
,('THPT Chuyên Hùng Vương','COUNTRY_VN',25,305,NULL,true,'2019-09-27 14:34:58.183','2019-09-27 14:34:58.183')
,('THPT Công nghiệp Việt Trì','COUNTRY_VN',25,305,NULL,true,'2019-09-27 14:34:58.183','2019-09-27 14:34:58.183')
,('THPT Dân lập Âu Cơ','COUNTRY_VN',25,305,NULL,true,'2019-09-27 14:34:58.184','2019-09-27 14:34:58.184')
,('THPT Dân lập Vân Phú','COUNTRY_VN',25,305,NULL,true,'2019-09-27 14:34:58.184','2019-09-27 14:34:58.184')
,('THPT Herman','COUNTRY_VN',25,305,NULL,true,'2019-09-27 14:34:58.185','2019-09-27 14:34:58.185')
,('THPT Kĩ thuật Việt Trì','COUNTRY_VN',25,305,NULL,true,'2019-09-27 14:34:58.185','2019-09-27 14:34:58.185')
,('THPT Lê Quý Đôn','COUNTRY_VN',25,305,NULL,true,'2019-09-27 14:34:58.185','2019-09-27 14:34:58.185')
,('THPT Nguyễn Tất Thành','COUNTRY_VN',25,306,NULL,true,'2019-09-27 14:34:58.186','2019-09-27 14:34:58.186')
,('THPT Trần Phú','COUNTRY_VN',25,306,NULL,true,'2019-09-27 14:34:58.187','2019-09-27 14:34:58.187')
,('THPT Dân tộc Nội trú Phú Thọ','COUNTRY_VN',25,306,NULL,true,'2019-09-27 14:34:58.187','2019-09-27 14:34:58.187')
,('THPT Bán công Hùng Vương','COUNTRY_VN',25,306,NULL,true,'2019-09-27 14:34:58.188','2019-09-27 14:34:58.188')
,('THPT Hùng Vương','COUNTRY_VN',25,306,NULL,true,'2019-09-27 14:34:58.188','2019-09-27 14:34:58.188')
,('THPT Thị xã Phú Thọ','COUNTRY_VN',25,306,NULL,true,'2019-09-27 14:34:58.188','2019-09-27 14:34:58.188')
,('THPT Trường Chinh','COUNTRY_VN',25,306,NULL,true,'2019-09-27 14:34:58.189','2019-09-27 14:34:58.189')
,('TTGDTX Thị xã Phú Thọ','COUNTRY_VN',25,306,NULL,true,'2019-09-27 14:34:58.189','2019-09-27 14:34:58.189')
,('THPT Dân tộc Nội trú Pinăng Tắc','COUNTRY_VN',26,307,NULL,true,'2019-09-27 14:34:58.190','2019-09-27 14:34:58.190')
,('THPT Bác Ái','COUNTRY_VN',26,307,NULL,true,'2019-09-27 14:34:58.192','2019-09-27 14:34:58.192')
,('THPT Ninh Hải','COUNTRY_VN',26,308,NULL,true,'2019-09-27 14:34:58.194','2019-09-27 14:34:58.194')
,('THPT Phan Chu Trinh','COUNTRY_VN',26,308,NULL,true,'2019-09-27 14:34:58.194','2019-09-27 14:34:58.194')
,('THPT Tôn Đức Thắng','COUNTRY_VN',26,308,NULL,true,'2019-09-27 14:34:58.195','2019-09-27 14:34:58.195')
,('THPT An Phuớc','COUNTRY_VN',26,309,NULL,true,'2019-09-27 14:34:58.198','2019-09-27 14:34:58.198')
,('THPT Nguyễn Huệ','COUNTRY_VN',26,310,NULL,true,'2019-09-27 14:34:58.199','2019-09-27 14:34:58.199')
,('THPT Phạm Văn Đồng','COUNTRY_VN',26,310,NULL,true,'2019-09-27 14:34:58.200','2019-09-27 14:34:58.200')
,('TTGDTX Huyện Ninh Phước','COUNTRY_VN',26,310,NULL,true,'2019-09-27 14:34:58.201','2019-09-27 14:34:58.201')
,('THPT Lê Duẩn','COUNTRY_VN',26,310,NULL,true,'2019-09-27 14:34:58.201','2019-09-27 14:34:58.201')
,('THPT Nguyễn Du','COUNTRY_VN',26,310,NULL,true,'2019-09-27 14:34:58.202','2019-09-27 14:34:58.202')
,('THPT Trường Chinh','COUNTRY_VN',26,310,NULL,true,'2019-09-27 14:34:58.202','2019-09-27 14:34:58.202')
,('TTGDTX Huyện Ninh Sơn','COUNTRY_VN',26,310,NULL,true,'2019-09-27 14:34:58.202','2019-09-27 14:34:58.202')
,('THPT Phan Bội Châu','COUNTRY_VN',26,311,NULL,true,'2019-09-27 14:34:58.203','2019-09-27 14:34:58.203')
,('THPT Nguyễn Văn Linh','COUNTRY_VN',26,312,NULL,true,'2019-09-27 14:34:58.204','2019-09-27 14:34:58.204')
,('THPT Chuyên Lê Quý Đôn','COUNTRY_VN',26,312,NULL,true,'2019-09-27 14:34:58.205','2019-09-27 14:34:58.205')
,('THPT Dân tộc Nội trú Ninh Thuận','COUNTRY_VN',26,313,NULL,true,'2019-09-27 14:34:58.206','2019-09-27 14:34:58.206')
,('THPT Ischool','COUNTRY_VN',26,313,NULL,true,'2019-09-27 14:34:58.206','2019-09-27 14:34:58.206')
,('THPT Nguyễn Trãi','COUNTRY_VN',26,313,NULL,true,'2019-09-27 14:34:58.207','2019-09-27 14:34:58.207')
,('THPT Tháp Chàm','COUNTRY_VN',26,313,NULL,true,'2019-09-27 14:34:58.208','2019-09-27 14:34:58.208')
,('TTGDTX Tỉnh Ninh Thuận','COUNTRY_VN',26,313,NULL,true,'2019-09-27 14:34:58.211','2019-09-27 14:34:58.211')
,('THPT Chu Văn An','COUNTRY_VN',26,313,NULL,true,'2019-09-27 14:34:58.211','2019-09-27 14:34:58.211')
,('THPT Gia Viễn A','COUNTRY_VN',27,314,NULL,true,'2019-09-27 14:34:58.214','2019-09-27 14:34:58.214')
,('THPT Gia Viễn B','COUNTRY_VN',27,314,NULL,true,'2019-09-27 14:34:58.215','2019-09-27 14:34:58.215')
,('THPT Gia Viễn C','COUNTRY_VN',27,314,NULL,true,'2019-09-27 14:34:58.216','2019-09-27 14:34:58.216')
,('THPT Gia Viễn D','COUNTRY_VN',27,314,NULL,true,'2019-09-27 14:34:58.217','2019-09-27 14:34:58.217')
,('TTGDTX Huyện Gia Viễn','COUNTRY_VN',27,314,NULL,true,'2019-09-27 14:34:58.218','2019-09-27 14:34:58.218')
,('THPT Hoa Lư A','COUNTRY_VN',27,315,NULL,true,'2019-09-27 14:34:58.219','2019-09-27 14:34:58.219')
,('THPT Trương Hàn Siêu','COUNTRY_VN',27,315,NULL,true,'2019-09-27 14:34:58.219','2019-09-27 14:34:58.219')
,('TTGDTX Huyện Hoa Lư','COUNTRY_VN',27,315,NULL,true,'2019-09-27 14:34:58.219','2019-09-27 14:34:58.219')
,('THPT Bình Minh','COUNTRY_VN',27,316,NULL,true,'2019-09-27 14:34:58.220','2019-09-27 14:34:58.220')
,('THPT Kim Sơn A','COUNTRY_VN',27,316,NULL,true,'2019-09-27 14:34:58.220','2019-09-27 14:34:58.220')
,('THPT Kim Sơn B','COUNTRY_VN',27,316,NULL,true,'2019-09-27 14:34:58.221','2019-09-27 14:34:58.221')
,('THPT Kim Sơn C','COUNTRY_VN',27,316,NULL,true,'2019-09-27 14:34:58.221','2019-09-27 14:34:58.221')
,('TTGDTX Huyện Kim Sơn','COUNTRY_VN',27,316,NULL,true,'2019-09-27 14:34:58.222','2019-09-27 14:34:58.222')
,('THPT Dân tộc Nội trú Ninh Bình','COUNTRY_VN',27,316,NULL,true,'2019-09-27 14:34:58.222','2019-09-27 14:34:58.222')
,('THPT Nho Quan A','COUNTRY_VN',27,317,NULL,true,'2019-09-27 14:34:58.223','2019-09-27 14:34:58.223')
,('THPT Nho Quan B','COUNTRY_VN',27,317,NULL,true,'2019-09-27 14:34:58.224','2019-09-27 14:34:58.225')
,('TTGDTX Huyện Nho Quan','COUNTRY_VN',27,317,NULL,true,'2019-09-27 14:34:58.226','2019-09-27 14:34:58.226')
,('THPT Vũ Duy Thanh','COUNTRY_VN',27,318,NULL,true,'2019-09-27 14:34:58.227','2019-09-27 14:34:58.227')
,('THPT Yên Khánh A','COUNTRY_VN',27,318,NULL,true,'2019-09-27 14:34:58.227','2019-09-27 14:34:58.227')
,('THPT Yên Khánh B','COUNTRY_VN',27,318,NULL,true,'2019-09-27 14:34:58.228','2019-09-27 14:34:58.228')
,('THPT Yên Khánh C','COUNTRY_VN',27,318,NULL,true,'2019-09-27 14:34:58.229','2019-09-27 14:34:58.229')
,('TTGDTX Huyện Yên Khánh','COUNTRY_VN',27,318,NULL,true,'2019-09-27 14:34:58.229','2019-09-27 14:34:58.229')
,('THPT Tạ Uyên','COUNTRY_VN',27,319,NULL,true,'2019-09-27 14:34:58.232','2019-09-27 14:34:58.232')
,('THPT Yên Mô A','COUNTRY_VN',27,319,NULL,true,'2019-09-27 14:34:58.232','2019-09-27 14:34:58.232')
,('THPT Yên Mô B','COUNTRY_VN',27,319,NULL,true,'2019-09-27 14:34:58.233','2019-09-27 14:34:58.233')
,('TTGDTX Huyện Yên Mô','COUNTRY_VN',27,319,NULL,true,'2019-09-27 14:34:58.233','2019-09-27 14:34:58.233')
,('THPT Chuyên Lương Văn Tụy','COUNTRY_VN',27,320,NULL,true,'2019-09-27 14:34:58.234','2019-09-27 14:34:58.234')
,('THPT Đinh Tiên Hoàng','COUNTRY_VN',27,320,NULL,true,'2019-09-27 14:34:58.235','2019-09-27 14:34:58.235')
,('THPT Nguyễn Công Trứ','COUNTRY_VN',27,320,NULL,true,'2019-09-27 14:34:58.235','2019-09-27 14:34:58.235')
,('THPT Ninh Bình Bạc Liêu','COUNTRY_VN',27,320,NULL,true,'2019-09-27 14:34:58.236','2019-09-27 14:34:58.236')
,('THPT Trần Hưng Đạo','COUNTRY_VN',27,320,NULL,true,'2019-09-27 14:34:58.236','2019-09-27 14:34:58.236')
,('TTGDTX Thành phố Ninh Bình','COUNTRY_VN',27,320,NULL,true,'2019-09-27 14:34:58.237','2019-09-27 14:34:58.237')
,('THPT Ngô Thì Nhậm','COUNTRY_VN',27,321,NULL,true,'2019-09-27 14:34:58.239','2019-09-27 14:34:58.239')
,('THPT Nguyễn Huệ','COUNTRY_VN',27,321,NULL,true,'2019-09-27 14:34:58.241','2019-09-27 14:34:58.241')
,('TTGDTX Thị xã Tam Điệp','COUNTRY_VN',27,321,NULL,true,'2019-09-27 14:34:58.242','2019-09-27 14:34:58.242')
,('THPT Anh Sơn 1','COUNTRY_VN',28,322,NULL,true,'2019-09-27 14:34:58.246','2019-09-27 14:34:58.246')
,('THPT Anh Sơn 2','COUNTRY_VN',28,322,NULL,true,'2019-09-27 14:34:58.246','2019-09-27 14:34:58.246')
,('THPT Anh Sơn 3','COUNTRY_VN',28,322,NULL,true,'2019-09-27 14:34:58.247','2019-09-27 14:34:58.247')
,('TTGDTX Huyện Anh Sơn','COUNTRY_VN',28,322,NULL,true,'2019-09-27 14:34:58.247','2019-09-27 14:34:58.247')
,('THPT Con Cuông','COUNTRY_VN',28,323,NULL,true,'2019-09-27 14:34:58.249','2019-09-27 14:34:58.249')
,('THPT Mường Quạ','COUNTRY_VN',28,323,NULL,true,'2019-09-27 14:34:58.250','2019-09-27 14:34:58.250')
,('TTGDTX Huyện Con Cuông','COUNTRY_VN',28,323,NULL,true,'2019-09-27 14:34:58.251','2019-09-27 14:34:58.251')
,('THPT Diễn Châu 2','COUNTRY_VN',28,324,NULL,true,'2019-09-27 14:34:58.252','2019-09-27 14:34:58.252')
,('THPT Diễn Châu 3','COUNTRY_VN',28,324,NULL,true,'2019-09-27 14:34:58.252','2019-09-27 14:34:58.252')
,('THPT Diễn Châu 4','COUNTRY_VN',28,324,NULL,true,'2019-09-27 14:34:58.253','2019-09-27 14:34:58.253')
,('THPT Diễn Châu 5','COUNTRY_VN',28,324,NULL,true,'2019-09-27 14:34:58.253','2019-09-27 14:34:58.253')
,('THPT Ngô Trí Hoà','COUNTRY_VN',28,324,NULL,true,'2019-09-27 14:34:58.254','2019-09-27 14:34:58.254')
,('THPT Nguyễn Du','COUNTRY_VN',28,324,NULL,true,'2019-09-27 14:34:58.254','2019-09-27 14:34:58.254')
,('THPT Nguyễn Văn Tố','COUNTRY_VN',28,324,NULL,true,'2019-09-27 14:34:58.254','2019-09-27 14:34:58.254')
,('THPT Nguyễn Xuân Ôn','COUNTRY_VN',28,324,NULL,true,'2019-09-27 14:34:58.255','2019-09-27 14:34:58.255')
,('THPT Quang Trung','COUNTRY_VN',28,324,NULL,true,'2019-09-27 14:34:58.255','2019-09-27 14:34:58.255')
,('TTGDTX Huyện Diễn Châu','COUNTRY_VN',28,324,NULL,true,'2019-09-27 14:34:58.256','2019-09-27 14:34:58.256')
,('THPT Duy Tân','COUNTRY_VN',28,325,NULL,true,'2019-09-27 14:34:58.256','2019-09-27 14:34:58.256')
,('THPT Đô Lương 1','COUNTRY_VN',28,325,NULL,true,'2019-09-27 14:34:58.257','2019-09-27 14:34:58.257')
,('THPT Đô Lương 2','COUNTRY_VN',28,325,NULL,true,'2019-09-27 14:34:58.258','2019-09-27 14:34:58.258')
,('THPT Đô Lương 3','COUNTRY_VN',28,325,NULL,true,'2019-09-27 14:34:58.259','2019-09-27 14:34:58.259')
,('THPT Đô Lương 4','COUNTRY_VN',28,325,NULL,true,'2019-09-27 14:34:58.260','2019-09-27 14:34:58.260')
,('THPT Văn Tràng','COUNTRY_VN',28,325,NULL,true,'2019-09-27 14:34:58.260','2019-09-27 14:34:58.260')
,('TTGDTX Huyện Đô Lương','COUNTRY_VN',28,325,NULL,true,'2019-09-27 14:34:58.261','2019-09-27 14:34:58.261')
,('THPT Đinh Bạt Tuy','COUNTRY_VN',28,326,NULL,true,'2019-09-27 14:34:58.262','2019-09-27 14:34:58.262')
,('THPT Lê Hồng Phong','COUNTRY_VN',28,326,NULL,true,'2019-09-27 14:34:58.263','2019-09-27 14:34:58.263')
,('THPT Nguyễn Trường Tộ','COUNTRY_VN',28,326,NULL,true,'2019-09-27 14:34:58.264','2019-09-27 14:34:58.264')
,('THPT Phạm Hồng Thái','COUNTRY_VN',28,326,NULL,true,'2019-09-27 14:34:58.265','2019-09-27 14:34:58.265')
,('THPT Thái Lão','COUNTRY_VN',28,326,NULL,true,'2019-09-27 14:34:58.265','2019-09-27 14:34:58.265')
,('TTGDTX Huyện Hưng Nguyên','COUNTRY_VN',28,326,NULL,true,'2019-09-27 14:34:58.266','2019-09-27 14:34:58.266')
,('THPT Kỳ Sơn','COUNTRY_VN',28,327,NULL,true,'2019-09-27 14:34:58.266','2019-09-27 14:34:58.266')
,('TTGDTX Huyện Kỳ Sơn','COUNTRY_VN',28,327,NULL,true,'2019-09-27 14:34:58.267','2019-09-27 14:34:58.267')
,('THPT Kim Liên','COUNTRY_VN',28,328,NULL,true,'2019-09-27 14:34:58.267','2019-09-27 14:34:58.267')
,('THPT Mai Hắc Đế','COUNTRY_VN',28,328,NULL,true,'2019-09-27 14:34:58.268','2019-09-27 14:34:58.268')
,('THPT Nam Đàn 1','COUNTRY_VN',28,328,NULL,true,'2019-09-27 14:34:58.268','2019-09-27 14:34:58.268')
,('THPT Nam Đàn 2','COUNTRY_VN',28,328,NULL,true,'2019-09-27 14:34:58.268','2019-09-27 14:34:58.268')
,('THPT Sào Nam','COUNTRY_VN',28,328,NULL,true,'2019-09-27 14:34:58.269','2019-09-27 14:34:58.269')
,('TTGDTX Huyện Nam Đàn','COUNTRY_VN',28,328,NULL,true,'2019-09-27 14:34:58.269','2019-09-27 14:34:58.269')
,('THPT Nghi Lộc 2','COUNTRY_VN',28,329,NULL,true,'2019-09-27 14:34:58.270','2019-09-27 14:34:58.270')
,('THPT Nghi Lộc 3','COUNTRY_VN',28,329,NULL,true,'2019-09-27 14:34:58.270','2019-09-27 14:34:58.270')
,('THPT Nghi Lộc 4','COUNTRY_VN',28,329,NULL,true,'2019-09-27 14:34:58.270','2019-09-27 14:34:58.270')
,('THPT Nghi Lộc 5','COUNTRY_VN',28,329,NULL,true,'2019-09-27 14:34:58.271','2019-09-27 14:34:58.271')
,('THPT Nguyễn Duy Trinh','COUNTRY_VN',28,329,NULL,true,'2019-09-27 14:34:58.271','2019-09-27 14:34:58.271')
,('THPT Nguyễn Thức Tự','COUNTRY_VN',28,329,NULL,true,'2019-09-27 14:34:58.271','2019-09-27 14:34:58.271')
,('TTGDTX Huyện Nghi Lộc','COUNTRY_VN',28,329,NULL,true,'2019-09-27 14:34:58.272','2019-09-27 14:34:58.272')
,('THPT 1/5','COUNTRY_VN',28,330,NULL,true,'2019-09-27 14:34:58.272','2019-09-27 14:34:58.272')
,('THPT Cờ Đỏ','COUNTRY_VN',28,330,NULL,true,'2019-09-27 14:34:58.273','2019-09-27 14:34:58.273')
,('TTGDTX Huyện Nghĩa Đàn','COUNTRY_VN',28,330,NULL,true,'2019-09-27 14:34:58.273','2019-09-27 14:34:58.273')
,('THPT Quế Phong','COUNTRY_VN',28,331,NULL,true,'2019-09-27 14:34:58.274','2019-09-27 14:34:58.274')
,('TTGDTX Huyện Quế Phong','COUNTRY_VN',28,331,NULL,true,'2019-09-27 14:34:58.275','2019-09-27 14:34:58.275')
,('THPT Quỳ Châu','COUNTRY_VN',28,332,NULL,true,'2019-09-27 14:34:58.277','2019-09-27 14:34:58.277')
,('TTGDTX Huyện Quỳ Châu','COUNTRY_VN',28,332,NULL,true,'2019-09-27 14:34:58.277','2019-09-27 14:34:58.277')
,('THPT Qùy Hợp 1','COUNTRY_VN',28,333,NULL,true,'2019-09-27 14:34:58.279','2019-09-27 14:34:58.279')
,('THPT Qùy Hợp 2','COUNTRY_VN',28,333,NULL,true,'2019-09-27 14:34:58.280','2019-09-27 14:34:58.280')
,('THPT Qùy Hợp 3','COUNTRY_VN',28,333,NULL,true,'2019-09-27 14:34:58.281','2019-09-27 14:34:58.281')
,('TTGDTX Huyện Quỳ Hợp','COUNTRY_VN',28,333,NULL,true,'2019-09-27 14:34:58.282','2019-09-27 14:34:58.282')
,('THPT Bắc Quỳnh Lưu','COUNTRY_VN',28,333,NULL,true,'2019-09-27 14:34:58.283','2019-09-27 14:34:58.283')
,('THPT Cù Chính Lan','COUNTRY_VN',28,333,NULL,true,'2019-09-27 14:34:58.283','2019-09-27 14:34:58.283')
,('THPT Hoàng Mai','COUNTRY_VN',28,333,NULL,true,'2019-09-27 14:34:58.284','2019-09-27 14:34:58.284')
,('THPT Lý Tự Trọng','COUNTRY_VN',28,333,NULL,true,'2019-09-27 14:34:58.284','2019-09-27 14:34:58.284')
,('THPT Nguyễn Đức Mậu','COUNTRY_VN',28,333,NULL,true,'2019-09-27 14:34:58.284','2019-09-27 14:34:58.284')
,('THPT Quỳnh Lưu 1','COUNTRY_VN',28,334,NULL,true,'2019-09-27 14:34:58.286','2019-09-27 14:34:58.286')
,('THPT Quỳnh Lưu 2','COUNTRY_VN',28,334,NULL,true,'2019-09-27 14:34:58.286','2019-09-27 14:34:58.286')
,('THPT Quỳnh Lưu 3','COUNTRY_VN',28,334,NULL,true,'2019-09-27 14:34:58.286','2019-09-27 14:34:58.286')
,('THPT Quỳnh Lun 4','COUNTRY_VN',28,334,NULL,true,'2019-09-27 14:34:58.287','2019-09-27 14:34:58.287')
,('TTGDTX Huyện Quỳnh Lưu','COUNTRY_VN',28,334,NULL,true,'2019-09-27 14:34:58.287','2019-09-27 14:34:58.287')
,('THPT Lê Lợi','COUNTRY_VN',28,335,NULL,true,'2019-09-27 14:34:58.288','2019-09-27 14:34:58.288')
,('THPT Tân Kỳ','COUNTRY_VN',28,335,NULL,true,'2019-09-27 14:34:58.289','2019-09-27 14:34:58.289')
,('THPT Tân Kỳ 3','COUNTRY_VN',28,335,NULL,true,'2019-09-27 14:34:58.289','2019-09-27 14:34:58.289')
,('TTGDTX Huyện Tân Kỳ','COUNTRY_VN',28,335,NULL,true,'2019-09-27 14:34:58.290','2019-09-27 14:34:58.290')
,('THPT Cát Ngạn','COUNTRY_VN',28,336,NULL,true,'2019-09-27 14:34:58.291','2019-09-27 14:34:58.291')
,('THPT Đặng Thai Mai','COUNTRY_VN',28,336,NULL,true,'2019-09-27 14:34:58.293','2019-09-27 14:34:58.293')
,('THPT Đặng Thúc Hứa','COUNTRY_VN',28,336,NULL,true,'2019-09-27 14:34:58.293','2019-09-27 14:34:58.293')
,('THPT Nguyễn Cảnh Chân','COUNTRY_VN',28,336,NULL,true,'2019-09-27 14:34:58.294','2019-09-27 14:34:58.294')
,('THPT Nguyễn Sỹ Sách','COUNTRY_VN',28,336,NULL,true,'2019-09-27 14:34:58.294','2019-09-27 14:34:58.294')
,('THPT Thanh Chương 1','COUNTRY_VN',28,336,NULL,true,'2019-09-27 14:34:58.295','2019-09-27 14:34:58.295')
,('THPT Thanh Chương 3','COUNTRY_VN',28,336,NULL,true,'2019-09-27 14:34:58.296','2019-09-27 14:34:58.296')
,('TTGDTX Huyện Thanh Chương','COUNTRY_VN',28,336,NULL,true,'2019-09-27 14:34:58.297','2019-09-27 14:34:58.297')
,('THPT Tương Dương 1','COUNTRY_VN',28,337,NULL,true,'2019-09-27 14:34:58.299','2019-09-27 14:34:58.299')
,('THPT Tương Dương 2','COUNTRY_VN',28,337,NULL,true,'2019-09-27 14:34:58.300','2019-09-27 14:34:58.300')
,('TTGDTX Huyện Tương Dương','COUNTRY_VN',28,337,NULL,true,'2019-09-27 14:34:58.300','2019-09-27 14:34:58.300')
,('THPT Băc Yên Thành','COUNTRY_VN',28,338,NULL,true,'2019-09-27 14:34:58.301','2019-09-27 14:34:58.301')
,('THPT Lê Doãn Nhã','COUNTRY_VN',28,338,NULL,true,'2019-09-27 14:34:58.301','2019-09-27 14:34:58.301')
,('THPT Nam Yên Thành','COUNTRY_VN',28,338,NULL,true,'2019-09-27 14:34:58.301','2019-09-27 14:34:58.301')
,('THPT Phan Đăng Luu','COUNTRY_VN',28,338,NULL,true,'2019-09-27 14:34:58.302','2019-09-27 14:34:58.302')
,('THPT Phan Thúc Trực','COUNTRY_VN',28,338,NULL,true,'2019-09-27 14:34:58.302','2019-09-27 14:34:58.302')
,('THPT Trần Đình Phong','COUNTRY_VN',28,338,NULL,true,'2019-09-27 14:34:58.302','2019-09-27 14:34:58.302')
,('THPT Yên Thành 2','COUNTRY_VN',28,338,NULL,true,'2019-09-27 14:34:58.303','2019-09-27 14:34:58.303')
,('THPT Yên Thành 3','COUNTRY_VN',28,338,NULL,true,'2019-09-27 14:34:58.303','2019-09-27 14:34:58.303')
,('TTGDTX Huyện Yên Thành','COUNTRY_VN',28,338,NULL,true,'2019-09-27 14:34:58.304','2019-09-27 14:34:58.304')
,('THPT Chuyên Toán ĐH Vinh','COUNTRY_VN',28,339,NULL,true,'2019-09-27 14:34:58.305','2019-09-27 14:34:58.305')
,('Phổ thông năng khiếu TDTT Nghệ An','COUNTRY_VN',28,339,NULL,true,'2019-09-27 14:34:58.305','2019-09-27 14:34:58.305')
,('THPT Chuyên Phan Bội Châu','COUNTRY_VN',28,339,NULL,true,'2019-09-27 14:34:58.305','2019-09-27 14:34:58.305')
,('THPT Dân tộc Nội trú Số 2','COUNTRY_VN',28,339,NULL,true,'2019-09-27 14:34:58.306','2019-09-27 14:34:58.306')
,('THPT Dân tộc Nội trú Nghệ An','COUNTRY_VN',28,339,NULL,true,'2019-09-27 14:34:58.306','2019-09-27 14:34:58.306')
,('THPT Hà Huy Tập','COUNTRY_VN',28,339,NULL,true,'2019-09-27 14:34:58.306','2019-09-27 14:34:58.306')
,('THPT Hermann Gmeiner','COUNTRY_VN',28,339,NULL,true,'2019-09-27 14:34:58.307','2019-09-27 14:34:58.307')
,('THPT Huỳnh Thúc Kháng','COUNTRY_VN',28,339,NULL,true,'2019-09-27 14:34:58.309','2019-09-27 14:34:58.309')
,('THPT Lê Viết Thuật','COUNTRY_VN',28,339,NULL,true,'2019-09-27 14:34:58.310','2019-09-27 14:34:58.310')
,('THPT Nguyễn Huệ','COUNTRY_VN',28,339,NULL,true,'2019-09-27 14:34:58.312','2019-09-27 14:34:58.312')
,('THPT Nguyễn Trãi','COUNTRY_VN',28,339,NULL,true,'2019-09-27 14:34:58.313','2019-09-27 14:34:58.313')
,('THPT Nguyễn Truờng Tộ','COUNTRY_VN',28,339,NULL,true,'2019-09-27 14:34:58.314','2019-09-27 14:34:58.314')
,('THPT VTC','COUNTRY_VN',28,339,NULL,true,'2019-09-27 14:34:58.315','2019-09-27 14:34:58.315')
,('TTGDTX Thành phố Vinh','COUNTRY_VN',28,339,NULL,true,'2019-09-27 14:34:58.316','2019-09-27 14:34:58.316')
,('THPT Cửa Lò','COUNTRY_VN',28,340,NULL,true,'2019-09-27 14:34:58.317','2019-09-27 14:34:58.317')
,('THPT Cửa Lò 2','COUNTRY_VN',28,340,NULL,true,'2019-09-27 14:34:58.318','2019-09-27 14:34:58.318')
,('TTGDTX Số 2','COUNTRY_VN',28,340,NULL,true,'2019-09-27 14:34:58.318','2019-09-27 14:34:58.318')
,('THPT Đông Hiếu','COUNTRY_VN',28,341,NULL,true,'2019-09-27 14:34:58.319','2019-09-27 14:34:58.319')
,('THPT Sông Hiếu','COUNTRY_VN',28,341,NULL,true,'2019-09-27 14:34:58.319','2019-09-27 14:34:58.319')
,('THPT Tây Hiếu','COUNTRY_VN',28,341,NULL,true,'2019-09-27 14:34:58.320','2019-09-27 14:34:58.320')
,('THPT Thái Hòa','COUNTRY_VN',28,341,NULL,true,'2019-09-27 14:34:58.320','2019-09-27 14:34:58.320')
,('TTGDTX Thị xã Thái Hòa','COUNTRY_VN',28,341,NULL,true,'2019-09-27 14:34:58.321','2019-09-27 14:34:58.321')
,('THPT Giao Thủy','COUNTRY_VN',29,342,NULL,true,'2019-09-27 14:34:58.322','2019-09-27 14:34:58.322')
,('THPT Giao Thủy B','COUNTRY_VN',29,342,NULL,true,'2019-09-27 14:34:58.322','2019-09-27 14:34:58.322')
,('THPT Giao Thủy C','COUNTRY_VN',29,342,NULL,true,'2019-09-27 14:34:58.322','2019-09-27 14:34:58.322')
,('THPT Quất Lâm','COUNTRY_VN',29,342,NULL,true,'2019-09-27 14:34:58.323','2019-09-27 14:34:58.323')
,('THPT Thiên Trường','COUNTRY_VN',29,342,NULL,true,'2019-09-27 14:34:58.323','2019-09-27 14:34:58.323')
,('TGDTX Huyện Giao Thủy','COUNTRY_VN',29,342,NULL,true,'2019-09-27 14:34:58.324','2019-09-27 14:34:58.324')
,('THPT A Hải Hậu','COUNTRY_VN',29,343,NULL,true,'2019-09-27 14:34:58.328','2019-09-27 14:34:58.328')
,('THPT An Phúc','COUNTRY_VN',29,343,NULL,true,'2019-09-27 14:34:58.329','2019-09-27 14:34:58.329')
,('THPT B Hải Hậu','COUNTRY_VN',29,343,NULL,true,'2019-09-27 14:34:58.329','2019-09-27 14:34:58.329')
,('THPT C Hải Hậu','COUNTRY_VN',29,343,NULL,true,'2019-09-27 14:34:58.330','2019-09-27 14:34:58.330')
,('THPT Thịnh Long','COUNTRY_VN',29,343,NULL,true,'2019-09-27 14:34:58.331','2019-09-27 14:34:58.331')
,('THPT Tô Hiến Thành','COUNTRY_VN',29,343,NULL,true,'2019-09-27 14:34:58.333','2019-09-27 14:34:58.333')
,('THPT Trần Quốc Tuấn','COUNTRY_VN',29,343,NULL,true,'2019-09-27 14:34:58.333','2019-09-27 14:34:58.333')
,('THPT Vũ Văn Hiếu','COUNTRY_VN',29,343,NULL,true,'2019-09-27 14:34:58.334','2019-09-27 14:34:58.334')
,('TTGDTX Huyện Hải Hậu','COUNTRY_VN',29,343,NULL,true,'2019-09-27 14:34:58.334','2019-09-27 14:34:58.334')
,('TTGDTX Hải Cường','COUNTRY_VN',29,343,NULL,true,'2019-09-27 14:34:58.335','2019-09-27 14:34:58.335')
,('THPT Mỹ Lộc','COUNTRY_VN',29,344,NULL,true,'2019-09-27 14:34:58.337','2019-09-27 14:34:58.337')
,('THPT Trần Văn Lan','COUNTRY_VN',29,344,NULL,true,'2019-09-27 14:34:58.337','2019-09-27 14:34:58.337')
,('TTGDTX Huyện Mỹ Lộc','COUNTRY_VN',29,344,NULL,true,'2019-09-27 14:34:58.337','2019-09-27 14:34:58.337')
,('THPT Lý Tự Trọng','COUNTRY_VN',29,345,NULL,true,'2019-09-27 14:34:58.338','2019-09-27 14:34:58.338')
,('THPT Nam Trực','COUNTRY_VN',29,345,NULL,true,'2019-09-27 14:34:58.339','2019-09-27 14:34:58.339')
,('THPT Nguyễn Du','COUNTRY_VN',29,345,NULL,true,'2019-09-27 14:34:58.339','2019-09-27 14:34:58.339')
,('THPT Phan Bội Châu','COUNTRY_VN',29,345,NULL,true,'2019-09-27 14:34:58.340','2019-09-27 14:34:58.340')
,('THPT Quang Trung','COUNTRY_VN',29,345,NULL,true,'2019-09-27 14:34:58.340','2019-09-27 14:34:58.340')
,('THPT Trần Văn Bảo','COUNTRY_VN',29,345,NULL,true,'2019-09-27 14:34:58.341','2019-09-27 14:34:58.341')
,('TTGDTX Vũ Tuấn Chiêu','COUNTRY_VN',29,345,NULL,true,'2019-09-27 14:34:58.344','2019-09-27 14:34:58.344')
,('TTGDTX Huyện Nam Trực','COUNTRY_VN',29,345,NULL,true,'2019-09-27 14:34:58.345','2019-09-27 14:34:58.345')
,('THPT A Nghĩa Hưng','COUNTRY_VN',29,346,NULL,true,'2019-09-27 14:34:58.346','2019-09-27 14:34:58.346')
,('THPT B Nghĩa Hưng','COUNTRY_VN',29,346,NULL,true,'2019-09-27 14:34:58.346','2019-09-27 14:34:58.346')
,('THPT C Nghĩa Hưng','COUNTRY_VN',29,346,NULL,true,'2019-09-27 14:34:58.347','2019-09-27 14:34:58.347')
,('THPT Nghĩa Hưng','COUNTRY_VN',29,346,NULL,true,'2019-09-27 14:34:58.348','2019-09-27 14:34:58.348')
,('THPT Nghĩa Minh','COUNTRY_VN',29,346,NULL,true,'2019-09-27 14:34:58.349','2019-09-27 14:34:58.349')
,('THPT Trần Nhân Tông','COUNTRY_VN',29,346,NULL,true,'2019-09-27 14:34:58.350','2019-09-27 14:34:58.350')
,('TTGDTX Nghĩa Tân','COUNTRY_VN',29,346,NULL,true,'2019-09-27 14:34:58.350','2019-09-27 14:34:58.350')
,('TTGDTX Huyện Nghĩa Hưng','COUNTRY_VN',29,346,NULL,true,'2019-09-27 14:34:58.351','2019-09-27 14:34:58.351')
,('THPT Đoàn Kết','COUNTRY_VN',29,347,NULL,true,'2019-09-27 14:34:58.352','2019-09-27 14:34:58.352')
,('THPT Lê Quý Đôn','COUNTRY_VN',29,347,NULL,true,'2019-09-27 14:34:58.352','2019-09-27 14:34:58.352')
,('THPT Nguyễn Trãi','COUNTRY_VN',29,347,NULL,true,'2019-09-27 14:34:58.353','2019-09-27 14:34:58.353')
,('THPT Trực Ninh','COUNTRY_VN',29,347,NULL,true,'2019-09-27 14:34:58.353','2019-09-27 14:34:58.353')
,('THPT Trực Ninh B','COUNTRY_VN',29,347,NULL,true,'2019-09-27 14:34:58.354','2019-09-27 14:34:58.354')
,('TTGDTX A Trực Ninh','COUNTRY_VN',29,347,NULL,true,'2019-09-27 14:34:58.354','2019-09-27 14:34:58.354')
,('TTGDTX B Trực Ninh','COUNTRY_VN',29,347,NULL,true,'2019-09-27 14:34:58.354','2019-09-27 14:34:58.354')
,('THPT Hoàng Văn Thụ','COUNTRY_VN',29,348,NULL,true,'2019-09-27 14:34:58.355','2019-09-27 14:34:58.355')
,('THPT Hùng Vương','COUNTRY_VN',29,348,NULL,true,'2019-09-27 14:34:58.355','2019-09-27 14:34:58.355')
,('THPT Lương Thế Vinh','COUNTRY_VN',29,348,NULL,true,'2019-09-27 14:34:58.356','2019-09-27 14:34:58.356')
,('THPT Ngô Quyền','COUNTRY_VN',29,348,NULL,true,'2019-09-27 14:34:58.356','2019-09-27 14:34:58.356')
,('THPT Nguyễn Bính','COUNTRY_VN',29,348,NULL,true,'2019-09-27 14:34:58.356','2019-09-27 14:34:58.356')
,('THPT Nguyễn Đức Thuận','COUNTRY_VN',29,348,NULL,true,'2019-09-27 14:34:58.358','2019-09-27 14:34:58.358')
,('TTGDTX Liên Minh','COUNTRY_VN',29,348,NULL,true,'2019-09-27 14:34:58.359','2019-09-27 14:34:58.359')
,('THPT Cao Phong','COUNTRY_VN',29,349,NULL,true,'2019-09-27 14:34:58.360','2019-09-27 14:34:58.360')
,('THPT Nguyễn Trường Thúy','COUNTRY_VN',29,349,NULL,true,'2019-09-27 14:34:58.361','2019-09-27 14:34:58.361')
,('THPT Xuân Trường A','COUNTRY_VN',29,349,NULL,true,'2019-09-27 14:34:58.361','2019-09-27 14:34:58.361')
,('THPT Xuân Trường B','COUNTRY_VN',29,349,NULL,true,'2019-09-27 14:34:58.362','2019-09-27 14:34:58.362')
,('THPT Xuân Trường C','COUNTRY_VN',29,349,NULL,true,'2019-09-27 14:34:58.363','2019-09-27 14:34:58.363')
,('TTGDTX Huyện Xuân Trường','COUNTRY_VN',29,349,NULL,true,'2019-09-27 14:34:58.363','2019-09-27 14:34:58.363')
,('THPT Đại An','COUNTRY_VN',29,350,NULL,true,'2019-09-27 14:34:58.365','2019-09-27 14:34:58.365')
,('THPT Đỗ Huy Liêu','COUNTRY_VN',29,350,NULL,true,'2019-09-27 14:34:58.366','2019-09-27 14:34:58.366')
,('THPT Lý Nhân Tông','COUNTRY_VN',29,350,NULL,true,'2019-09-27 14:34:58.366','2019-09-27 14:34:58.366')
,('THPT Mỹ Tho','COUNTRY_VN',29,350,NULL,true,'2019-09-27 14:34:58.367','2019-09-27 14:34:58.367')
,('THPT Phạm Văn Nghị','COUNTRY_VN',29,350,NULL,true,'2019-09-27 14:34:58.367','2019-09-27 14:34:58.367')
,('THPT Tống Văn Trân','COUNTRY_VN',29,350,NULL,true,'2019-09-27 14:34:58.368','2019-09-27 14:34:58.368')
,('THPT Ý Yên','COUNTRY_VN',29,350,NULL,true,'2019-09-27 14:34:58.368','2019-09-27 14:34:58.368')
,('TTGDTX Huyện Ý Yên','COUNTRY_VN',29,350,NULL,true,'2019-09-27 14:34:58.368','2019-09-27 14:34:58.368')
,('THPT Chuyên Lê Hồng Phong','COUNTRY_VN',29,351,NULL,true,'2019-09-27 14:34:58.369','2019-09-27 14:34:58.369')
,('THPT Dân lập Trần Nhật Duật','COUNTRY_VN',29,351,NULL,true,'2019-09-27 14:34:58.369','2019-09-27 14:34:58.369')
,('THPT Nguyễn Công Trứ','COUNTRY_VN',29,351,NULL,true,'2019-09-27 14:34:58.370','2019-09-27 14:34:58.370')
,('THPT Nguyễn Huệ','COUNTRY_VN',29,351,NULL,true,'2019-09-27 14:34:58.370','2019-09-27 14:34:58.370')
,('THPT Nguyễn Khuyến','COUNTRY_VN',29,351,NULL,true,'2019-09-27 14:34:58.371','2019-09-27 14:34:58.371')
,('THPT Trần Hưng Đạo','COUNTRY_VN',29,351,NULL,true,'2019-09-27 14:34:58.371','2019-09-27 14:34:58.371')
,('THPT Trần Quang Khải','COUNTRY_VN',29,351,NULL,true,'2019-09-27 14:34:58.372','2019-09-27 14:34:58.372')
,('TTGDTX Tỉnh Nam Định','COUNTRY_VN',29,351,NULL,true,'2019-09-27 14:34:58.372','2019-09-27 14:34:58.372')
,('TTGDTX Trần Phú','COUNTRY_VN',29,351,NULL,true,'2019-09-27 14:34:58.372','2019-09-27 14:34:58.372')
,('THPT Bắc Sơn','COUNTRY_VN',30,352,NULL,true,'2019-09-27 14:34:58.373','2019-09-27 14:34:58.373')
,('THPT Vũ Lễ','COUNTRY_VN',30,352,NULL,true,'2019-09-27 14:34:58.374','2019-09-27 14:34:58.374')
,('TT GDTX Bắc Sơn','COUNTRY_VN',30,352,NULL,true,'2019-09-27 14:34:58.376','2019-09-27 14:34:58.376')
,('TT GDTX Bình Gia','COUNTRY_VN',30,353,NULL,true,'2019-09-27 14:34:58.377','2019-09-27 14:34:58.377')
,('Phổ thông DTNT - THCS','COUNTRY_VN',30,353,NULL,true,'2019-09-27 14:34:58.377','2019-09-27 14:34:58.377')
,('PT DT Nội Trú - THCS huyện Cao Lộc','COUNTRY_VN',30,354,NULL,true,'2019-09-27 14:34:58.378','2019-09-27 14:34:58.378')
,('THPT Cao Lộc','COUNTRY_VN',30,354,NULL,true,'2019-09-27 14:34:58.379','2019-09-27 14:34:58.379')
,('THPT Đồng Đăng','COUNTRY_VN',30,354,NULL,true,'2019-09-27 14:34:58.380','2019-09-27 14:34:58.380')
,('TT GDTX Cao Lộc','COUNTRY_VN',30,354,NULL,true,'2019-09-27 14:34:58.381','2019-09-27 14:34:58.381')
,('THPT Chi Lăng','COUNTRY_VN',30,355,NULL,true,'2019-09-27 14:34:58.383','2019-09-27 14:34:58.383')
,('THPT Đồng Bành','COUNTRY_VN',30,355,NULL,true,'2019-09-27 14:34:58.383','2019-09-27 14:34:58.383')
,('THPT Hòa Bình','COUNTRY_VN',30,355,NULL,true,'2019-09-27 14:34:58.384','2019-09-27 14:34:58.384')
,('TT GDTX Chi Lăng','COUNTRY_VN',30,355,NULL,true,'2019-09-27 14:34:58.384','2019-09-27 14:34:58.384')
,('THPT Đình Lập','COUNTRY_VN',30,356,NULL,true,'2019-09-27 14:34:58.385','2019-09-27 14:34:58.385')
,('TT GDTX Đình Lập','COUNTRY_VN',30,356,NULL,true,'2019-09-27 14:34:58.385','2019-09-27 14:34:58.385')
,('CĐ nghề và công nghệ Nông Lâm Đông Bắc','COUNTRY_VN',30,357,NULL,true,'2019-09-27 14:34:58.386','2019-09-27 14:34:58.386')
,('THPT Hữu Lũng','COUNTRY_VN',30,357,NULL,true,'2019-09-27 14:34:58.386','2019-09-27 14:34:58.386')
,('THPT Vân Nham','COUNTRY_VN',30,357,NULL,true,'2019-09-27 14:34:58.386','2019-09-27 14:34:58.386')
,('TT GDTX 2 tỉnh','COUNTRY_VN',30,357,NULL,true,'2019-09-27 14:34:58.387','2019-09-27 14:34:58.387')
,('THPT Lộc Bình','COUNTRY_VN',30,358,NULL,true,'2019-09-27 14:34:58.388','2019-09-27 14:34:58.388')
,('THPT Na Dương','COUNTRY_VN',30,358,NULL,true,'2019-09-27 14:34:58.388','2019-09-27 14:34:58.388')
,('THPT Tú Đoạn','COUNTRY_VN',30,358,NULL,true,'2019-09-27 14:34:58.388','2019-09-27 14:34:58.388')
,('TT GDTX Lộc Bình','COUNTRY_VN',30,358,NULL,true,'2019-09-27 14:34:58.389','2019-09-27 14:34:58.389')
,('THPT Binh Độ','COUNTRY_VN',30,359,NULL,true,'2019-09-27 14:34:58.389','2019-09-27 14:34:58.389')
,('THPT Tràng Định','COUNTRY_VN',30,359,NULL,true,'2019-09-27 14:34:58.390','2019-09-27 14:34:58.390')
,('TT GDTX Tràng Định','COUNTRY_VN',30,359,NULL,true,'2019-09-27 14:34:58.390','2019-09-27 14:34:58.390')
,('THPT Vãn Lãng','COUNTRY_VN',30,360,NULL,true,'2019-09-27 14:34:58.393','2019-09-27 14:34:58.393')
,('TT GDTX Vãn Lãng','COUNTRY_VN',30,360,NULL,true,'2019-09-27 14:34:58.393','2019-09-27 14:34:58.393')
,('THPT Lương Văn Tri','COUNTRY_VN',30,361,NULL,true,'2019-09-27 14:34:58.396','2019-09-27 14:34:58.396')
,('THPT Văn Quan','COUNTRY_VN',30,361,NULL,true,'2019-09-27 14:34:58.398','2019-09-27 14:34:58.398')
,('TT GDTX Văn Quan','COUNTRY_VN',30,361,NULL,true,'2019-09-27 14:34:58.399','2019-09-27 14:34:58.399')
,('Cao đắng nghề Lạng Sơn','COUNTRY_VN',30,362,NULL,true,'2019-09-27 14:34:58.400','2019-09-27 14:34:58.400')
,('THPT Chuyên Chu văn An','COUNTRY_VN',30,362,NULL,true,'2019-09-27 14:34:58.401','2019-09-27 14:34:58.401')
,('THPT DT Nội trú tỉnh','COUNTRY_VN',30,362,NULL,true,'2019-09-27 14:34:58.401','2019-09-27 14:34:58.401')
,('THPT Ngô Thì Sỹ','COUNTRY_VN',30,362,NULL,true,'2019-09-27 14:34:58.402','2019-09-27 14:34:58.402')
,('THPT Việt Bắc','COUNTRY_VN',30,362,NULL,true,'2019-09-27 14:34:58.402','2019-09-27 14:34:58.402')
,('TT GDTX 1 tỉnh','COUNTRY_VN',30,362,NULL,true,'2019-09-27 14:34:58.402','2019-09-27 14:34:58.402')
,('PTDT nội trú THCS và THPT H. Bắc Hà','COUNTRY_VN',31,363,NULL,true,'2019-09-27 14:34:58.404','2019-09-27 14:34:58.404')
,('THPT số 1 Bắc Hà','COUNTRY_VN',31,363,NULL,true,'2019-09-27 14:34:58.405','2019-09-27 14:34:58.405')
,('THPT số 2 Bắc Hà','COUNTRY_VN',31,363,NULL,true,'2019-09-27 14:34:58.405','2019-09-27 14:34:58.405')
,('TT Dạy nghề và GDTX Bắc Hà','COUNTRY_VN',31,363,NULL,true,'2019-09-27 14:34:58.406','2019-09-27 14:34:58.406')
,('TT GDTX Bắc Hà','COUNTRY_VN',31,363,NULL,true,'2019-09-27 14:34:58.406','2019-09-27 14:34:58.406')
,('THPT sõ 1 Bảo Thắng','COUNTRY_VN',31,364,NULL,true,'2019-09-27 14:34:58.407','2019-09-27 14:34:58.407')
,('THPT sô 2 Bảo Thắng','COUNTRY_VN',31,364,NULL,true,'2019-09-27 14:34:58.410','2019-09-27 14:34:58.410')
,('THPT sô 3 Bảo Thắng','COUNTRY_VN',31,364,NULL,true,'2019-09-27 14:34:58.411','2019-09-27 14:34:58.411')
,('TT Dạy nghề và GDTX Bảo Thắng','COUNTRY_VN',31,364,NULL,true,'2019-09-27 14:34:58.412','2019-09-27 14:34:58.412')
,('TT GDTX Bảo Thắng','COUNTRY_VN',31,364,NULL,true,'2019-09-27 14:34:58.413','2019-09-27 14:34:58.413')
,('THPT sõ 1 Bảo Yên','COUNTRY_VN',31,365,NULL,true,'2019-09-27 14:34:58.415','2019-09-27 14:34:58.415')
,('THPT sõ 2 Bảo Yên','COUNTRY_VN',31,365,NULL,true,'2019-09-27 14:34:58.416','2019-09-27 14:34:58.416')
,('THPT sõ 3 Bảo Yên','COUNTRY_VN',31,365,NULL,true,'2019-09-27 14:34:58.417','2019-09-27 14:34:58.417')
,('TT Dạy nghề và GDTX Bảo Yên','COUNTRY_VN',31,365,NULL,true,'2019-09-27 14:34:58.418','2019-09-27 14:34:58.418')
,('TT GDTX Bảo Yên','COUNTRY_VN',31,365,NULL,true,'2019-09-27 14:34:58.418','2019-09-27 14:34:58.418')
,('THCS và THPT huyện Bát xát','COUNTRY_VN',31,366,NULL,true,'2019-09-27 14:34:58.419','2019-09-27 14:34:58.419')
,('THPT Sõ 1 Bát Xát','COUNTRY_VN',31,366,NULL,true,'2019-09-27 14:34:58.420','2019-09-27 14:34:58.420')
,('THPT sõ 2 Bát Xát','COUNTRY_VN',31,366,NULL,true,'2019-09-27 14:34:58.420','2019-09-27 14:34:58.420')
,('TT Dạy nghề và GDTX Bát xát','COUNTRY_VN',31,366,NULL,true,'2019-09-27 14:34:58.420','2019-09-27 14:34:58.420')
,('TT GDTX Bát xát','COUNTRY_VN',31,366,NULL,true,'2019-09-27 14:34:58.421','2019-09-27 14:34:58.421')
,('THPT số 1 Mường Khương','COUNTRY_VN',31,367,NULL,true,'2019-09-27 14:34:58.422','2019-09-27 14:34:58.422')
,('THPT sỗ 2 Mường Khương','COUNTRY_VN',31,367,NULL,true,'2019-09-27 14:34:58.422','2019-09-27 14:34:58.422')
,('THPT sỗ 3 Mường Khương','COUNTRY_VN',31,367,NULL,true,'2019-09-27 14:34:58.423','2019-09-27 14:34:58.423')
,('TT Dạy nghề và GDTX Muờng Khương','COUNTRY_VN',31,367,NULL,true,'2019-09-27 14:34:58.423','2019-09-27 14:34:58.423')
,('TT GDTX Mường Khương','COUNTRY_VN',31,367,NULL,true,'2019-09-27 14:34:58.424','2019-09-27 14:34:58.424')
,('PTDT nội trú THCS và THPT H.Sa Pa','COUNTRY_VN',31,368,NULL,true,'2019-09-27 14:34:58.426','2019-09-27 14:34:58.426')
,('THPT SỐ 1 Sa Pa','COUNTRY_VN',31,368,NULL,true,'2019-09-27 14:34:58.427','2019-09-27 14:34:58.427')
,('THPT số 2 Sa Pa','COUNTRY_VN',31,368,NULL,true,'2019-09-27 14:34:58.427','2019-09-27 14:34:58.427')
,('TT Dạy nghề và GDTX Sa Pa','COUNTRY_VN',31,368,NULL,true,'2019-09-27 14:34:58.427','2019-09-27 14:34:58.427')
,('TT GDTX Sa Pa','COUNTRY_VN',31,368,NULL,true,'2019-09-27 14:34:58.428','2019-09-27 14:34:58.428')
,('PTDT nội trú THCS và THPT H.si Ma Cai','COUNTRY_VN',31,369,NULL,true,'2019-09-27 14:34:58.430','2019-09-27 14:34:58.430')
,('THPT SỐ1 Si Ma Cai','COUNTRY_VN',31,369,NULL,true,'2019-09-27 14:34:58.431','2019-09-27 14:34:58.431')
,('THPT SỐ 2 Si ma cai','COUNTRY_VN',31,369,NULL,true,'2019-09-27 14:34:58.432','2019-09-27 14:34:58.432')
,('TT Dạy nghề và GDTX Si Ma Cai','COUNTRY_VN',31,369,NULL,true,'2019-09-27 14:34:58.432','2019-09-27 14:34:58.432')
,('TT GDTX Si Ma Cai','COUNTRY_VN',31,369,NULL,true,'2019-09-27 14:34:58.433','2019-09-27 14:34:58.433')
,('THPT số 1 Văn Bàn','COUNTRY_VN',31,370,NULL,true,'2019-09-27 14:34:58.434','2019-09-27 14:34:58.434')
,('THPT số 2 Văn Bàn','COUNTRY_VN',31,370,NULL,true,'2019-09-27 14:34:58.434','2019-09-27 14:34:58.434')
,('THPT số 3 Văn Bàn','COUNTRY_VN',31,370,NULL,true,'2019-09-27 14:34:58.434','2019-09-27 14:34:58.434')
,('THPT số 4 Văn Bàn','COUNTRY_VN',31,370,NULL,true,'2019-09-27 14:34:58.435','2019-09-27 14:34:58.435')
,('TT Dạy nghề và GDTX Văn Bàn','COUNTRY_VN',31,370,NULL,true,'2019-09-27 14:34:58.435','2019-09-27 14:34:58.435')
,('TT GDTX Văn Bàn','COUNTRY_VN',31,370,NULL,true,'2019-09-27 14:34:58.436','2019-09-27 14:34:58.436')
,('CĐ nghề tỉnh Lào Cai','COUNTRY_VN',31,371,NULL,true,'2019-09-27 14:34:58.436','2019-09-27 14:34:58.436')
,('THPT Chuyên tỉnh Lào Cai','COUNTRY_VN',31,371,NULL,true,'2019-09-27 14:34:58.437','2019-09-27 14:34:58.437')
,('THPT DTNT tỉnh','COUNTRY_VN',31,371,NULL,true,'2019-09-27 14:34:58.437','2019-09-27 14:34:58.437')
,('THPT số 1 Tp Lào Cai','COUNTRY_VN',31,371,NULL,true,'2019-09-27 14:34:58.438','2019-09-27 14:34:58.438')
,('THPT SỐ 2 Tp Lào Cai','COUNTRY_VN',31,371,NULL,true,'2019-09-27 14:34:58.438','2019-09-27 14:34:58.438')
,('THPT số 3 Tp Lào Cai','COUNTRY_VN',31,371,NULL,true,'2019-09-27 14:34:58.438','2019-09-27 14:34:58.438')
,('THPT số 4 Tp Lào Cai','COUNTRY_VN',31,371,NULL,true,'2019-09-27 14:34:58.439','2019-09-27 14:34:58.439')
,('TT Dạy nghề và GDTX TP Lào Cai','COUNTRY_VN',31,371,NULL,true,'2019-09-27 14:34:58.439','2019-09-27 14:34:58.439')
,('TT GDTX số 1 TP Lào Cai','COUNTRY_VN',31,371,NULL,true,'2019-09-27 14:34:58.439','2019-09-27 14:34:58.439')
,('TT GDTX SỐ2TP Lào Cai','COUNTRY_VN',31,371,NULL,true,'2019-09-27 14:34:58.440','2019-09-27 14:34:58.440')
,('TTKT-TH-HN-DN & GDTX tỉnh','COUNTRY_VN',31,371,NULL,true,'2019-09-27 14:34:58.440','2019-09-27 14:34:58.440')
,('CĐ nghề Tây Sài Gòn','COUNTRY_VN',32,372,NULL,true,'2019-09-27 14:34:58.443','2019-09-27 14:34:58.443')
,('TC KT-KT Long An','COUNTRY_VN',32,372,NULL,true,'2019-09-27 14:34:58.443','2019-09-27 14:34:58.443')
,('THCS & THPT Lương Hòa','COUNTRY_VN',32,372,NULL,true,'2019-09-27 14:34:58.444','2019-09-27 14:34:58.444')
,('THCS & THPT iSCHOOL Long An','COUNTRY_VN',32,372,NULL,true,'2019-09-27 14:34:58.444','2019-09-27 14:34:58.444')
,('THPT GÒ Đen','COUNTRY_VN',32,372,NULL,true,'2019-09-27 14:34:58.444','2019-09-27 14:34:58.444')
,('THPT Nguyễn Hữu Thọ','COUNTRY_VN',32,372,NULL,true,'2019-09-27 14:34:58.445','2019-09-27 14:34:58.445')
,('TT.GDTX &KTTH-HN Bến Lức','COUNTRY_VN',32,372,NULL,true,'2019-09-27 14:34:58.446','2019-09-27 14:34:58.446')
,('THCS & THPT Long Cang','COUNTRY_VN',32,373,NULL,true,'2019-09-27 14:34:58.447','2019-09-27 14:34:58.447')
,('THCS & THPT Long Hụu Đông','COUNTRY_VN',32,373,NULL,true,'2019-09-27 14:34:58.448','2019-09-27 14:34:58.448')
,('THPT Cần Đước','COUNTRY_VN',32,373,NULL,true,'2019-09-27 14:34:58.449','2019-09-27 14:34:58.449')
,('THPT Chu Văn An','COUNTRY_VN',32,373,NULL,true,'2019-09-27 14:34:58.449','2019-09-27 14:34:58.449')
,('THPT Long Hòa','COUNTRY_VN',32,373,NULL,true,'2019-09-27 14:34:58.450','2019-09-27 14:34:58.450')
,('THPT Rạch Kiến','COUNTRY_VN',32,373,NULL,true,'2019-09-27 14:34:58.450','2019-09-27 14:34:58.450')
,('TT.GDTX &KTTH-HN cần Đưức','COUNTRY_VN',32,373,NULL,true,'2019-09-27 14:34:58.451','2019-09-27 14:34:58.451')
,('TC nghề cần Giuộc','COUNTRY_VN',32,374,NULL,true,'2019-09-27 14:34:58.452','2019-09-27 14:34:58.452')
,('THCS & THPT Long Thượng','COUNTRY_VN',32,374,NULL,true,'2019-09-27 14:34:58.452','2019-09-27 14:34:58.452')
,('THPT Cần Giuộc','COUNTRY_VN',32,374,NULL,true,'2019-09-27 14:34:58.453','2019-09-27 14:34:58.453')
,('THPT Đông Thạnh','COUNTRY_VN',32,374,NULL,true,'2019-09-27 14:34:58.453','2019-09-27 14:34:58.453')
,('THPT Nguyễn Đình Chiểu','COUNTRY_VN',32,374,NULL,true,'2019-09-27 14:34:58.453','2019-09-27 14:34:58.453')
,('TT.GDTX &KTTH-HN cần Giuộc','COUNTRY_VN',32,374,NULL,true,'2019-09-27 14:34:58.454','2019-09-27 14:34:58.454')
,('THPT Châu Thành','COUNTRY_VN',32,375,NULL,true,'2019-09-27 14:34:58.454','2019-09-27 14:34:58.454')
,('THPT Nguyễn Thống','COUNTRY_VN',32,375,NULL,true,'2019-09-27 14:34:58.455','2019-09-27 14:34:58.455')
,('THPT Phan Văn Đạt','COUNTRY_VN',32,375,NULL,true,'2019-09-27 14:34:58.455','2019-09-27 14:34:58.455')
,('TT.GDTX &KTTH-HN châu Thành','COUNTRY_VN',32,375,NULL,true,'2019-09-27 14:34:58.456','2019-09-27 14:34:58.456')
,('TC nghề Đức Hòa','COUNTRY_VN',32,376,NULL,true,'2019-09-27 14:34:58.456','2019-09-27 14:34:58.456')
,('THPT An Ninh','COUNTRY_VN',32,376,NULL,true,'2019-09-27 14:34:58.457','2019-09-27 14:34:58.457')
,('THPT Đức Hòa','COUNTRY_VN',32,376,NULL,true,'2019-09-27 14:34:58.458','2019-09-27 14:34:58.458')
,('THPT Hậu Nghĩa','COUNTRY_VN',32,376,NULL,true,'2019-09-27 14:34:58.458','2019-09-27 14:34:58.458')
,('THPT Năng khiếu Đại học Tân Tạo','COUNTRY_VN',32,376,NULL,true,'2019-09-27 14:34:58.459','2019-09-27 14:34:58.459')
,('THPT Nguyễn Công Trứ','COUNTRY_VN',32,376,NULL,true,'2019-09-27 14:34:58.459','2019-09-27 14:34:58.459')
,('THCS & THPT Mỹ Quý','COUNTRY_VN',32,377,NULL,true,'2019-09-27 14:34:58.462','2019-09-27 14:34:58.462')
,('THCS & THPT Mỹ Bình','COUNTRY_VN',32,377,NULL,true,'2019-09-27 14:34:58.464','2019-09-27 14:34:58.464')
,('THPT Đức Huệ','COUNTRY_VN',32,377,NULL,true,'2019-09-27 14:34:58.464','2019-09-27 14:34:58.464')
,('TT.GDTX &KTTH-HN Đírc Huệ','COUNTRY_VN',32,377,NULL,true,'2019-09-27 14:34:58.466','2019-09-27 14:34:58.466')
,('THCS & THPT Mỹ Quý','COUNTRY_VN',32,378,NULL,true,'2019-09-27 14:34:58.469','2019-09-27 14:34:58.469')
,('THCS & THPT Mỹ Bình','COUNTRY_VN',32,378,NULL,true,'2019-09-27 14:34:58.471','2019-09-27 14:34:58.471')
,('THPT Đức Huệ','COUNTRY_VN',32,378,NULL,true,'2019-09-27 14:34:58.472','2019-09-27 14:34:58.472')
,('TT.GDTX &KTTH-HN Đírc Huệ','COUNTRY_VN',32,378,NULL,true,'2019-09-27 14:34:58.472','2019-09-27 14:34:58.472')
,('THCS & THPT Bình Phong Thạnh','COUNTRY_VN',32,379,NULL,true,'2019-09-27 14:34:58.474','2019-09-27 14:34:58.474')
,('THPT Tân Hưng','COUNTRY_VN',32,380,NULL,true,'2019-09-27 14:34:58.477','2019-09-27 14:34:58.477')
,('TT.GDTX &KTTH-HN Tân Hưng','COUNTRY_VN',32,380,NULL,true,'2019-09-27 14:34:58.478','2019-09-27 14:34:58.478')
,('THCS & THPT Hậu Thạnh Đông','COUNTRY_VN',32,381,NULL,true,'2019-09-27 14:34:58.481','2019-09-27 14:34:58.481')
,('THPT Tân Thạnh','COUNTRY_VN',32,381,NULL,true,'2019-09-27 14:34:58.482','2019-09-27 14:34:58.482')
,('TT.GDTX &KTTH-HN Tân Thạnh','COUNTRY_VN',32,381,NULL,true,'2019-09-27 14:34:58.483','2019-09-27 14:34:58.483')
,('THPT Nguyễn Trung Trục','COUNTRY_VN',32,382,NULL,true,'2019-09-27 14:34:58.484','2019-09-27 14:34:58.484')
,('THPT Tân Trụ','COUNTRY_VN',32,382,NULL,true,'2019-09-27 14:34:58.485','2019-09-27 14:34:58.485')
,('TT.GDTX &KTTH-HN Tân Trụ','COUNTRY_VN',32,382,NULL,true,'2019-09-27 14:34:58.485','2019-09-27 14:34:58.485')
,('THPT Thạnh Hóa','COUNTRY_VN',32,383,NULL,true,'2019-09-27 14:34:58.486','2019-09-27 14:34:58.486')
,('TT.GDTX &KTTH-HN Thạnh Hoá','COUNTRY_VN',32,383,NULL,true,'2019-09-27 14:34:58.487','2019-09-27 14:34:58.487')
,('THTHCS & THPT Bồ Đề Phương Duy','COUNTRY_VN',32,384,NULL,true,'2019-09-27 14:34:58.488','2019-09-27 14:34:58.488')
,('THCS & THPT Mỹ Lạc','COUNTRY_VN',32,384,NULL,true,'2019-09-27 14:34:58.489','2019-09-27 14:34:58.489')
,('THPT Thủ Khoa Thừa','COUNTRY_VN',32,384,NULL,true,'2019-09-27 14:34:58.489','2019-09-27 14:34:58.489')
,('THPT Thủ Thừa','COUNTRY_VN',32,384,NULL,true,'2019-09-27 14:34:58.490','2019-09-27 14:34:58.490')
,('TT.GDTX &KTTH-HN Thủ Thừa','COUNTRY_VN',32,384,NULL,true,'2019-09-27 14:34:58.491','2019-09-27 14:34:58.491')
,('THCS & THPT Khánh Hưng','COUNTRY_VN',32,385,NULL,true,'2019-09-27 14:34:58.494','2019-09-27 14:34:58.494')
,('THPT Vĩnh Hưng','COUNTRY_VN',32,385,NULL,true,'2019-09-27 14:34:58.494','2019-09-27 14:34:58.494')
,('TT.GDTX &KTTH-HN Vĩnh Hưng','COUNTRY_VN',32,385,NULL,true,'2019-09-27 14:34:58.495','2019-09-27 14:34:58.495')
,('CĐN Kỹ thuật Công nghệ LADEC','COUNTRY_VN',32,386,NULL,true,'2019-09-27 14:34:58.497','2019-09-27 14:34:58.497')
,('CĐN Long An','COUNTRY_VN',32,386,NULL,true,'2019-09-27 14:34:58.497','2019-09-27 14:34:58.497')
,('TC Việt-Nhật','COUNTRY_VN',32,386,NULL,true,'2019-09-27 14:34:58.498','2019-09-27 14:34:58.498')
,('TDTT Tỉnh Long An','COUNTRY_VN',32,386,NULL,true,'2019-09-27 14:34:58.498','2019-09-27 14:34:58.498')
,('THCS & THPT Hà Long','COUNTRY_VN',32,386,NULL,true,'2019-09-27 14:34:58.499','2019-09-27 14:34:58.499')
,('THCS & THPT Nguyễn Văn Rành','COUNTRY_VN',32,386,NULL,true,'2019-09-27 14:34:58.499','2019-09-27 14:34:58.499')
,('THPT chuyên Long An','COUNTRY_VN',32,386,NULL,true,'2019-09-27 14:34:58.500','2019-09-27 14:34:58.500')
,('THPT Hùng Vương','COUNTRY_VN',32,386,NULL,true,'2019-09-27 14:34:58.501','2019-09-27 14:34:58.501')
,('THPT Huỳnh Ngọc','COUNTRY_VN',32,386,NULL,true,'2019-09-27 14:34:58.501','2019-09-27 14:34:58.501')
,('THPT Lê Quý Đôn','COUNTRY_VN',32,386,NULL,true,'2019-09-27 14:34:58.502','2019-09-27 14:34:58.502')
,('THPT Tân An','COUNTRY_VN',32,386,NULL,true,'2019-09-27 14:34:58.502','2019-09-27 14:34:58.502')
,('TT.GDTX Long An','COUNTRY_VN',32,386,NULL,true,'2019-09-27 14:34:58.503','2019-09-27 14:34:58.503')
,('TT.GDTX Tp. Tân An','COUNTRY_VN',32,386,NULL,true,'2019-09-27 14:34:58.503','2019-09-27 14:34:58.503')
,('TT.KTTH-HN Long An','COUNTRY_VN',32,386,NULL,true,'2019-09-27 14:34:58.504','2019-09-27 14:34:58.504')
,('THPT Thạnh Hóa','COUNTRY_VN',32,387,NULL,true,'2019-09-27 14:34:58.504','2019-09-27 14:34:58.504')
,('TT.GDTX &KTTH-HN Thạnh Hoá','COUNTRY_VN',32,387,NULL,true,'2019-09-27 14:34:58.505','2019-09-27 14:34:58.505')
,('TC Nghề Tân Châu','COUNTRY_VN',33,388,NULL,true,'2019-09-27 14:34:58.506','2019-09-27 14:34:58.506')
,('THPT An Phú','COUNTRY_VN',33,388,NULL,true,'2019-09-27 14:34:58.506','2019-09-27 14:34:58.506')
,('THPT An Phú 2','COUNTRY_VN',33,388,NULL,true,'2019-09-27 14:34:58.507','2019-09-27 14:34:58.507')
,('THPT Nguyễn Quang Diêu','COUNTRY_VN',33,388,NULL,true,'2019-09-27 14:34:58.508','2019-09-27 14:34:58.508')
,('THPT Quốc Thái','COUNTRY_VN',33,388,NULL,true,'2019-09-27 14:34:58.510','2019-09-27 14:34:58.510')
,('THPT Vĩnh Lộc','COUNTRY_VN',33,388,NULL,true,'2019-09-27 14:34:58.511','2019-09-27 14:34:58.511')
,('TTDN-GDTX An Phú','COUNTRY_VN',33,388,NULL,true,'2019-09-27 14:34:58.512','2019-09-27 14:34:58.512')
,('TC Nghề Châu Đốc','COUNTRY_VN',33,389,NULL,true,'2019-09-27 14:34:58.513','2019-09-27 14:34:58.513')
,('THPT Thủ Khoa Nghiã','COUNTRY_VN',33,389,NULL,true,'2019-09-27 14:34:58.514','2019-09-27 14:34:58.514')
,('THPT Võ Thị Sáu','COUNTRY_VN',33,389,NULL,true,'2019-09-27 14:34:58.515','2019-09-27 14:34:58.515')
,('TT GDTX Châu Đốc','COUNTRY_VN',33,389,NULL,true,'2019-09-27 14:34:58.516','2019-09-27 14:34:58.516')
,('Phổ thông Bình Long','COUNTRY_VN',33,390,NULL,true,'2019-09-27 14:34:58.518','2019-09-27 14:34:58.518')
,('TC Kinh tế-Kỹ thuật An Giang','COUNTRY_VN',33,390,NULL,true,'2019-09-27 14:34:58.518','2019-09-27 14:34:58.518')
,('THPT Bình Mỹ','COUNTRY_VN',33,390,NULL,true,'2019-09-27 14:34:58.519','2019-09-27 14:34:58.519')
,('THPT Châu Phú','COUNTRY_VN',33,390,NULL,true,'2019-09-27 14:34:58.519','2019-09-27 14:34:58.519')
,('THPT Thạnh Mỹ Tây','COUNTRY_VN',33,390,NULL,true,'2019-09-27 14:34:58.519','2019-09-27 14:34:58.519')
,('THPT Trần Văn Thành','COUNTRY_VN',33,390,NULL,true,'2019-09-27 14:34:58.520','2019-09-27 14:34:58.520')
,('TTDN-GDTX châu Phú','COUNTRY_VN',33,390,NULL,true,'2019-09-27 14:34:58.520','2019-09-27 14:34:58.520')
,('THPT Cần Đăng','COUNTRY_VN',33,391,NULL,true,'2019-09-27 14:34:58.521','2019-09-27 14:34:58.521')
,('THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',33,391,NULL,true,'2019-09-27 14:34:58.522','2019-09-27 14:34:58.522')
,('THPT Vĩnh Bình','COUNTRY_VN',33,391,NULL,true,'2019-09-27 14:34:58.522','2019-09-27 14:34:58.522')
,('TTDN-GDTX châu Thành','COUNTRY_VN',33,391,NULL,true,'2019-09-27 14:34:58.523','2019-09-27 14:34:58.523')
,('THPT Châu Văn Liêm','COUNTRY_VN',33,392,NULL,true,'2019-09-27 14:34:58.524','2019-09-27 14:34:58.524')
,('THPT Hòa Bình','COUNTRY_VN',33,392,NULL,true,'2019-09-27 14:34:58.525','2019-09-27 14:34:58.525')
,('THPT Huỳnh Thị Huùng','COUNTRY_VN',33,392,NULL,true,'2019-09-27 14:34:58.527','2019-09-27 14:34:58.527')
,('THPT Long Kiến','COUNTRY_VN',33,392,NULL,true,'2019-09-27 14:34:58.527','2019-09-27 14:34:58.527')
,('THPT Mỹ Hiệp','COUNTRY_VN',33,392,NULL,true,'2019-09-27 14:34:58.528','2019-09-27 14:34:58.528')
,('THPT Mỹ Hội Đông','COUNTRY_VN',33,392,NULL,true,'2019-09-27 14:34:58.528','2019-09-27 14:34:58.528')
,('CĐ Nghề An Giang','COUNTRY_VN',33,393,NULL,true,'2019-09-27 14:34:58.530','2019-09-27 14:34:58.530')
,('Năng khiếu thể thao','COUNTRY_VN',33,393,NULL,true,'2019-09-27 14:34:58.531','2019-09-27 14:34:58.531')
,('Phổ thông Quốc tế GIS','COUNTRY_VN',33,393,NULL,true,'2019-09-27 14:34:58.531','2019-09-27 14:34:58.531')
,('Phổ thông Thưc hành Sư phạm','COUNTRY_VN',33,393,NULL,true,'2019-09-27 14:34:58.533','2019-09-27 14:34:58.533')
,('TC Nghề KTKT công Đoàn An Giang','COUNTRY_VN',33,393,NULL,true,'2019-09-27 14:34:58.533','2019-09-27 14:34:58.533')
,('TH Y Tế','COUNTRY_VN',33,393,NULL,true,'2019-09-27 14:34:58.534','2019-09-27 14:34:58.534')
,('THPT Bình Khánh','COUNTRY_VN',33,393,NULL,true,'2019-09-27 14:34:58.534','2019-09-27 14:34:58.534')
,('phổ thông Phú Tân','COUNTRY_VN',33,394,NULL,true,'2019-09-27 14:34:58.535','2019-09-27 14:34:58.535')
,('THPT Bình Thạnh Đông','COUNTRY_VN',33,394,NULL,true,'2019-09-27 14:34:58.536','2019-09-27 14:34:58.536')
,('THPT Chu Văn An','COUNTRY_VN',33,394,NULL,true,'2019-09-27 14:34:58.536','2019-09-27 14:34:58.536')
,('THPT Hoà Lạc','COUNTRY_VN',33,394,NULL,true,'2019-09-27 14:34:58.536','2019-09-27 14:34:58.536')
,('THPT Nguyễn Chí Thanh','COUNTRY_VN',33,394,NULL,true,'2019-09-27 14:34:58.537','2019-09-27 14:34:58.537')
,('TTDN-GDTX Phú Tân','COUNTRY_VN',33,394,NULL,true,'2019-09-27 14:34:58.537','2019-09-27 14:34:58.537')
,('THPT Châu Phong','COUNTRY_VN',33,395,NULL,true,'2019-09-27 14:34:58.538','2019-09-27 14:34:58.538')
,('THPT Đức Trí','COUNTRY_VN',33,395,NULL,true,'2019-09-27 14:34:58.539','2019-09-27 14:34:58.539')
,('THPT Tân châu','COUNTRY_VN',33,395,NULL,true,'2019-09-27 14:34:58.539','2019-09-27 14:34:58.539')
,('THPT Vĩnh Xương','COUNTRY_VN',33,395,NULL,true,'2019-09-27 14:34:58.540','2019-09-27 14:34:58.540')
,('TT GDTX Tân châu','COUNTRY_VN',33,395,NULL,true,'2019-09-27 14:34:58.540','2019-09-27 14:34:58.540')
,('THPT Nguyễn Khuyến','COUNTRY_VN',33,396,NULL,true,'2019-09-27 14:34:58.545','2019-09-27 14:34:58.545')
,('THPT Nguyễn Văn Thoại','COUNTRY_VN',33,396,NULL,true,'2019-09-27 14:34:58.545','2019-09-27 14:34:58.545')
,('THPT Vĩnh Trạch','COUNTRY_VN',33,396,NULL,true,'2019-09-27 14:34:58.546','2019-09-27 14:34:58.546')
,('THPT Vọng Thê','COUNTRY_VN',33,396,NULL,true,'2019-09-27 14:34:58.548','2019-09-27 14:34:58.548')
,('TTDN-GDTX Thoại Sơn','COUNTRY_VN',33,396,NULL,true,'2019-09-27 14:34:58.550','2019-09-27 14:34:58.550')
,('THPT Chi Lăng','COUNTRY_VN',33,397,NULL,true,'2019-09-27 14:34:58.551','2019-09-27 14:34:58.551')
,('THPT Tịnh Biên','COUNTRY_VN',33,397,NULL,true,'2019-09-27 14:34:58.552','2019-09-27 14:34:58.552')
,('THPT Xuân Tô','COUNTRY_VN',33,397,NULL,true,'2019-09-27 14:34:58.552','2019-09-27 14:34:58.552')
,('TTDN-GDTX Tịnh Biên','COUNTRY_VN',33,397,NULL,true,'2019-09-27 14:34:58.553','2019-09-27 14:34:58.553')
,('Phổ thông Cô Tô','COUNTRY_VN',33,398,NULL,true,'2019-09-27 14:34:58.554','2019-09-27 14:34:58.554')
,('TC Nghề Dân tộc Nội Trú An Giang','COUNTRY_VN',33,398,NULL,true,'2019-09-27 14:34:58.554','2019-09-27 14:34:58.554')
,('THPT Ba Chúc','COUNTRY_VN',33,398,NULL,true,'2019-09-27 14:34:58.555','2019-09-27 14:34:58.555')
,('THPT Dân Tộc Nội Trú','COUNTRY_VN',33,398,NULL,true,'2019-09-27 14:34:58.555','2019-09-27 14:34:58.555')
,('THPT Nguyễn Trung Trục','COUNTRY_VN',33,398,NULL,true,'2019-09-27 14:34:58.556','2019-09-27 14:34:58.556')
,('TT GDTX Tri Tôn','COUNTRY_VN',33,398,NULL,true,'2019-09-27 14:34:58.557','2019-09-27 14:34:58.557')
,('TC Nghề Châu Đốc','COUNTRY_VN',33,399,NULL,true,'2019-09-27 14:34:58.559','2019-09-27 14:34:58.559')
,('THPT Thủ Khoa Nghiã','COUNTRY_VN',33,399,NULL,true,'2019-09-27 14:34:58.560','2019-09-27 14:34:58.560')
,('THPT Võ Thị Sáu','COUNTRY_VN',33,399,NULL,true,'2019-09-27 14:34:58.561','2019-09-27 14:34:58.561')
,('TT GDTX Châu Đốc','COUNTRY_VN',33,399,NULL,true,'2019-09-27 14:34:58.562','2019-09-27 14:34:58.562')
,('CĐ Nghề An Giang','COUNTRY_VN',33,400,NULL,true,'2019-09-27 14:34:58.565','2019-09-27 14:34:58.565')
,('Năng khiếu thể thao','COUNTRY_VN',33,400,NULL,true,'2019-09-27 14:34:58.566','2019-09-27 14:34:58.566')
,('Phổ thông Quốc tế GIS','COUNTRY_VN',33,400,NULL,true,'2019-09-27 14:34:58.567','2019-09-27 14:34:58.567')
,('Phổ thông Thực hành Sư phạm','COUNTRY_VN',33,400,NULL,true,'2019-09-27 14:34:58.568','2019-09-27 14:34:58.568')
,('TC Nghề KTKT công Đoàn An Giang','COUNTRY_VN',33,400,NULL,true,'2019-09-27 14:34:58.569','2019-09-27 14:34:58.569')
,('TH Y Tế','COUNTRY_VN',33,400,NULL,true,'2019-09-27 14:34:58.569','2019-09-27 14:34:58.569')
,('Phổ thông dân tộc nội trú tỉnh','COUNTRY_VN',34,401,NULL,true,'2019-09-27 14:34:58.571','2019-09-27 14:34:58.571')
,('THPT Ngô Quyền','COUNTRY_VN',34,401,NULL,true,'2019-09-27 14:34:58.571','2019-09-27 14:34:58.571')
,('THPT Nguyễn Du','COUNTRY_VN',34,401,NULL,true,'2019-09-27 14:34:58.572','2019-09-27 14:34:58.572')
,('THPT Nguyễn Trãi','COUNTRY_VN',34,401,NULL,true,'2019-09-27 14:34:58.572','2019-09-27 14:34:58.572')
,('THPT Nguyễn Văn Cừ','COUNTRY_VN',34,401,NULL,true,'2019-09-27 14:34:58.573','2019-09-27 14:34:58.573')
,('THPT Trần Phú','COUNTRY_VN',34,401,NULL,true,'2019-09-27 14:34:58.573','2019-09-27 14:34:58.573')
,('TT GDTX -DN-GTVL châu Đức','COUNTRY_VN',34,401,NULL,true,'2019-09-27 14:34:58.575','2019-09-27 14:34:58.575')
,('THCS-THPT Võ Thị sáu','COUNTRY_VN',34,402,NULL,true,'2019-09-27 14:34:58.578','2019-09-27 14:34:58.578')
,('TT GDTX Côn Đảo','COUNTRY_VN',34,402,NULL,true,'2019-09-27 14:34:58.579','2019-09-27 14:34:58.579')
,('THPT Dương Bạch Mai','COUNTRY_VN',34,403,NULL,true,'2019-09-27 14:34:58.581','2019-09-27 14:34:58.581')
,('THPT Võ Thị Sáu','COUNTRY_VN',34,403,NULL,true,'2019-09-27 14:34:58.582','2019-09-27 14:34:58.582')
,('TT GDTX-HN Đất Đỏ','COUNTRY_VN',34,403,NULL,true,'2019-09-27 14:34:58.583','2019-09-27 14:34:58.583')
,('THPT Long Hải - Phuớc tỉnh','COUNTRY_VN',34,404,NULL,true,'2019-09-27 14:34:58.584','2019-09-27 14:34:58.584')
,('THPT Minh Đạm','COUNTRY_VN',34,404,NULL,true,'2019-09-27 14:34:58.585','2019-09-27 14:34:58.585')
,('THPT Trần Quang Khải','COUNTRY_VN',34,404,NULL,true,'2019-09-27 14:34:58.585','2019-09-27 14:34:58.585')
,('THPT Trần Văn Quan','COUNTRY_VN',34,404,NULL,true,'2019-09-27 14:34:58.585','2019-09-27 14:34:58.585')
,('TT GDTX Long Điền','COUNTRY_VN',34,404,NULL,true,'2019-09-27 14:34:58.586','2019-09-27 14:34:58.586')
,('CĐ nghề quốc tế Hồng Lam','COUNTRY_VN',34,405,NULL,true,'2019-09-27 14:34:58.587','2019-09-27 14:34:58.587')
,('THPT Hắc Dịch','COUNTRY_VN',34,405,NULL,true,'2019-09-27 14:34:58.587','2019-09-27 14:34:58.587')
,('THPT Phú Mỹ','COUNTRY_VN',34,405,NULL,true,'2019-09-27 14:34:58.588','2019-09-27 14:34:58.588')
,('THPT Trần Hưng Đạo','COUNTRY_VN',34,405,NULL,true,'2019-09-27 14:34:58.588','2019-09-27 14:34:58.588')
,('TT GDTX Tân Thành','COUNTRY_VN',34,405,NULL,true,'2019-09-27 14:34:58.589','2019-09-27 14:34:58.589')
,('THPT Bưng Riềng','COUNTRY_VN',34,406,NULL,true,'2019-09-27 14:34:58.589','2019-09-27 14:34:58.589')
,('THPT Hòa Bình','COUNTRY_VN',34,406,NULL,true,'2019-09-27 14:34:58.590','2019-09-27 14:34:58.590')
,('THPT Hoà Hội','COUNTRY_VN',34,406,NULL,true,'2019-09-27 14:34:58.590','2019-09-27 14:34:58.590')
,('THPT Phước Bửu','COUNTRY_VN',34,406,NULL,true,'2019-09-27 14:34:58.592','2019-09-27 14:34:58.592')
,('THPT Xuyên Mộc','COUNTRY_VN',34,406,NULL,true,'2019-09-27 14:34:58.593','2019-09-27 14:34:58.593')
,('TT GDTX -DN-GTVL Xuyên Mộc','COUNTRY_VN',34,406,NULL,true,'2019-09-27 14:34:58.594','2019-09-27 14:34:58.594')
,('THPT Bà Rịa','COUNTRY_VN',34,407,NULL,true,'2019-09-27 14:34:58.596','2019-09-27 14:34:58.596')
,('THPT Châu Thành','COUNTRY_VN',34,407,NULL,true,'2019-09-27 14:34:58.597','2019-09-27 14:34:58.597')
,('THPT DL Chu văn An','COUNTRY_VN',34,407,NULL,true,'2019-09-27 14:34:58.598','2019-09-27 14:34:58.598')
,('THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',34,407,NULL,true,'2019-09-27 14:34:58.599','2019-09-27 14:34:58.599')
,('TT GDTX -DN-GTVL Bà Rịa','COUNTRY_VN',34,407,NULL,true,'2019-09-27 14:34:58.600','2019-09-27 14:34:58.600')
,('BTVH Cấp 2.3 Nguyễn Thái Học','COUNTRY_VN',34,408,NULL,true,'2019-09-27 14:34:58.601','2019-09-27 14:34:58.601')
,('CĐ nghề Dầu khí','COUNTRY_VN',34,408,NULL,true,'2019-09-27 14:34:58.602','2019-09-27 14:34:58.602')
,('CĐ nghề Du lịch Vũng Tàu','COUNTRY_VN',34,408,NULL,true,'2019-09-27 14:34:58.602','2019-09-27 14:34:58.602')
,('CĐ nghề tỉnh Bà Rịa-Vũng Tàu','COUNTRY_VN',34,408,NULL,true,'2019-09-27 14:34:58.603','2019-09-27 14:34:58.603')
,('TC Công nghệ thông tin TM. COMPUTER','COUNTRY_VN',34,408,NULL,true,'2019-09-27 14:34:58.604','2019-09-27 14:34:58.604')
,('TC nghề Giao thông vận tải','COUNTRY_VN',34,408,NULL,true,'2019-09-27 14:34:58.604','2019-09-27 14:34:58.604')
,('TC nghề KTKT công đoàn Bà Rja - VT','COUNTRY_VN',34,408,NULL,true,'2019-09-27 14:34:58.604','2019-09-27 14:34:58.604')
,('THCS-THPT Song ngữ','COUNTRY_VN',34,408,NULL,true,'2019-09-27 14:34:58.605','2019-09-27 14:34:58.605')
,('THPT Chuyên Lê Quý Đôn','COUNTRY_VN',34,408,NULL,true,'2019-09-27 14:34:58.605','2019-09-27 14:34:58.605')
,('THPT Đinh Tiên Hoàng','COUNTRY_VN',34,408,NULL,true,'2019-09-27 14:34:58.606','2019-09-27 14:34:58.606')
,('THPT Lê Hồng Phong','COUNTRY_VN',34,408,NULL,true,'2019-09-27 14:34:58.606','2019-09-27 14:34:58.606')
,('THPT Nguyễn Huệ','COUNTRY_VN',34,408,NULL,true,'2019-09-27 14:34:58.606','2019-09-27 14:34:58.606')
,('THPT Nguyễn Thị Minh Khai','COUNTRY_VN',34,408,NULL,true,'2019-09-27 14:34:58.607','2019-09-27 14:34:58.607')
,('THPT Trần Nguyên Hãn','COUNTRY_VN',34,408,NULL,true,'2019-09-27 14:34:58.608','2019-09-27 14:34:58.608')
,('THPT Vũng Tàu','COUNTRY_VN',34,408,NULL,true,'2019-09-27 14:34:58.611','2019-09-27 14:34:58.611')
,('TSTD Vũng Tàu','COUNTRY_VN',34,408,NULL,true,'2019-09-27 14:34:58.611','2019-09-27 14:34:58.611')
,('TT GDTX-HN Vũng Tàu','COUNTRY_VN',34,408,NULL,true,'2019-09-27 14:34:58.612','2019-09-27 14:34:58.612')
,('THPT Dân lập Hiệp Hoà 1','COUNTRY_VN',35,409,NULL,true,'2019-09-27 14:34:58.614','2019-09-27 14:34:58.614')
,('THPT Dân lập Hiệp Hoà 2','COUNTRY_VN',35,409,NULL,true,'2019-09-27 14:34:58.615','2019-09-27 14:34:58.615')
,('THPT Hiệp Hoà 1','COUNTRY_VN',35,409,NULL,true,'2019-09-27 14:34:58.616','2019-09-27 14:34:58.616')
,('THPT Hiệp Hoà 2','COUNTRY_VN',35,409,NULL,true,'2019-09-27 14:34:58.617','2019-09-27 14:34:58.617')
,('THPT Hiệp Hoà 3','COUNTRY_VN',35,409,NULL,true,'2019-09-27 14:34:58.617','2019-09-27 14:34:58.617')
,('THPT Hiệp Hòa 4','COUNTRY_VN',35,409,NULL,true,'2019-09-27 14:34:58.618','2019-09-27 14:34:58.618')
,('TT GDTX H. Hiệp Hoà','COUNTRY_VN',35,409,NULL,true,'2019-09-27 14:34:58.618','2019-09-27 14:34:58.618')
,('TC nghề số 12 Bộ Quốc phòng','COUNTRY_VN',35,410,NULL,true,'2019-09-27 14:34:58.619','2019-09-27 14:34:58.619')
,('THPT Dân Lập Phi Mô','COUNTRY_VN',35,410,NULL,true,'2019-09-27 14:34:58.620','2019-09-27 14:34:58.620')
,('THPT Dân lập Thái Đào','COUNTRY_VN',35,410,NULL,true,'2019-09-27 14:34:58.620','2019-09-27 14:34:58.620')
,('THPT Lạng Giang 1','COUNTRY_VN',35,410,NULL,true,'2019-09-27 14:34:58.621','2019-09-27 14:34:58.621')
,('THPT Lạng Giang 2','COUNTRY_VN',35,410,NULL,true,'2019-09-27 14:34:58.621','2019-09-27 14:34:58.621')
,('THPT Lạng Giang 3','COUNTRY_VN',35,410,NULL,true,'2019-09-27 14:34:58.622','2019-09-27 14:34:58.622')
,('TT GDTX H. Lạng Giang','COUNTRY_VN',35,410,NULL,true,'2019-09-27 14:34:58.622','2019-09-27 14:34:58.622')
,('THPT Cẩm Lý','COUNTRY_VN',35,411,NULL,true,'2019-09-27 14:34:58.623','2019-09-27 14:34:58.623')
,('THPT Dân lập Đồi Ngô','COUNTRY_VN',35,411,NULL,true,'2019-09-27 14:34:58.624','2019-09-27 14:34:58.624')
,('THPT Lục Nam','COUNTRY_VN',35,411,NULL,true,'2019-09-27 14:34:58.626','2019-09-27 14:34:58.626')
,('THPT Phương Sơn','COUNTRY_VN',35,411,NULL,true,'2019-09-27 14:34:58.626','2019-09-27 14:34:58.626')
,('THPT Tứ Sơn','COUNTRY_VN',35,411,NULL,true,'2019-09-27 14:34:58.627','2019-09-27 14:34:58.627')
,('THPT Tư thục Thanh Hồ','COUNTRY_VN',35,411,NULL,true,'2019-09-27 14:34:58.627','2019-09-27 14:34:58.627')
,('TT GDTX H. Lục Nam','COUNTRY_VN',35,411,NULL,true,'2019-09-27 14:34:58.628','2019-09-27 14:34:58.628')
,('DTNT H. Lục Ngạn','COUNTRY_VN',35,412,NULL,true,'2019-09-27 14:34:58.630','2019-09-27 14:34:58.630')
,('THPT bán công Lục Ngạn','COUNTRY_VN',35,412,NULL,true,'2019-09-27 14:34:58.631','2019-09-27 14:34:58.631')
,('THPT Lục Ngạn 1','COUNTRY_VN',35,412,NULL,true,'2019-09-27 14:34:58.632','2019-09-27 14:34:58.632')
,('THPT Lục Ngạn 2','COUNTRY_VN',35,412,NULL,true,'2019-09-27 14:34:58.632','2019-09-27 14:34:58.632')
,('THPT Lục ngạn 3','COUNTRY_VN',35,412,NULL,true,'2019-09-27 14:34:58.633','2019-09-27 14:34:58.633')
,('Trung THPT Lục Ngạn số 4','COUNTRY_VN',35,412,NULL,true,'2019-09-27 14:34:58.633','2019-09-27 14:34:58.633')
,('TT GDTX H. Lục Ngạn','COUNTRY_VN',35,412,NULL,true,'2019-09-27 14:34:58.634','2019-09-27 14:34:58.634')
,('DTNTH.Scm Động','COUNTRY_VN',35,413,NULL,true,'2019-09-27 14:34:58.635','2019-09-27 14:34:58.635')
,('THPT Sơn Động','COUNTRY_VN',35,413,NULL,true,'2019-09-27 14:34:58.636','2019-09-27 14:34:58.636')
,('THPT Sơn Động 2','COUNTRY_VN',35,413,NULL,true,'2019-09-27 14:34:58.636','2019-09-27 14:34:58.636')
,('THPT Sơn Động 3','COUNTRY_VN',35,413,NULL,true,'2019-09-27 14:34:58.636','2019-09-27 14:34:58.636')
,('TT GDTX H. Sơn Động','COUNTRY_VN',35,413,NULL,true,'2019-09-27 14:34:58.637','2019-09-27 14:34:58.637')
,('THPT Dân lập Tân Yên','COUNTRY_VN',35,414,NULL,true,'2019-09-27 14:34:58.638','2019-09-27 14:34:58.638')
,('THPT Nhã Nam','COUNTRY_VN',35,414,NULL,true,'2019-09-27 14:34:58.638','2019-09-27 14:34:58.638')
,('THPT Tân Yên 1','COUNTRY_VN',35,414,NULL,true,'2019-09-27 14:34:58.638','2019-09-27 14:34:58.638')
,('THPT Tân Yên 2','COUNTRY_VN',35,414,NULL,true,'2019-09-27 14:34:58.639','2019-09-27 14:34:58.639')
,('TT GDTX H. Tân Yên','COUNTRY_VN',35,414,NULL,true,'2019-09-27 14:34:58.639','2019-09-27 14:34:58.639')
,('THPT Lý Thường Kiệt','COUNTRY_VN',35,415,NULL,true,'2019-09-27 14:34:58.641','2019-09-27 14:34:58.641')
,('THPT Tư thục Việt Yên','COUNTRY_VN',35,415,NULL,true,'2019-09-27 14:34:58.642','2019-09-27 14:34:58.642')
,('THPT Việt Yên 1','COUNTRY_VN',35,415,NULL,true,'2019-09-27 14:34:58.643','2019-09-27 14:34:58.643')
,('THPT Việt Yên 2','COUNTRY_VN',35,415,NULL,true,'2019-09-27 14:34:58.643','2019-09-27 14:34:58.643')
,('TT GDTX H. Việt Yên','COUNTRY_VN',35,415,NULL,true,'2019-09-27 14:34:58.644','2019-09-27 14:34:58.644')
,('THPT Dân lập Quang Trung','COUNTRY_VN',35,416,NULL,true,'2019-09-27 14:34:58.645','2019-09-27 14:34:58.645')
,('THPT Dân lập Yên Dũng 1','COUNTRY_VN',35,416,NULL,true,'2019-09-27 14:34:58.646','2019-09-27 14:34:58.646')
,('THPT Tư thục Thái Sơn','COUNTRY_VN',35,416,NULL,true,'2019-09-27 14:34:58.648','2019-09-27 14:34:58.648')
,('THPT Yẻn Dũng 1','COUNTRY_VN',35,416,NULL,true,'2019-09-27 14:34:58.649','2019-09-27 14:34:58.649')
,('THPT Yẻn Dũng 2','COUNTRY_VN',35,416,NULL,true,'2019-09-27 14:34:58.649','2019-09-27 14:34:58.649')
,('THPT Yẻn Dũng 3','COUNTRY_VN',35,416,NULL,true,'2019-09-27 14:34:58.650','2019-09-27 14:34:58.650')
,('TT GDTX H. Yên Dũng','COUNTRY_VN',35,416,NULL,true,'2019-09-27 14:34:58.650','2019-09-27 14:34:58.650')
,('TC nghề MN Yên Thế','COUNTRY_VN',35,417,NULL,true,'2019-09-27 14:34:58.651','2019-09-27 14:34:58.651')
,('THPT Bổ Hạ','COUNTRY_VN',35,417,NULL,true,'2019-09-27 14:34:58.651','2019-09-27 14:34:58.651')
,('THPT Mỏ Trạng','COUNTRY_VN',35,417,NULL,true,'2019-09-27 14:34:58.652','2019-09-27 14:34:58.652')
,('THPT Yên Thế','COUNTRY_VN',35,417,NULL,true,'2019-09-27 14:34:58.652','2019-09-27 14:34:58.652')
,('TT GDTX H. Yên Thế','COUNTRY_VN',35,417,NULL,true,'2019-09-27 14:34:58.653','2019-09-27 14:34:58.653')
,('CĐ Kỹ thuật Công nghiệp','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.654','2019-09-27 14:34:58.654')
,('CĐ nghề Bắc Giang','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.654','2019-09-27 14:34:58.654')
,('TC nghề GTVT','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.654','2019-09-27 14:34:58.654')
,('TC nghề Lái xe sổ 1','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.655','2019-09-27 14:34:58.655')
,('TC nghề Thủ công mỹ nghệ 19.5','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.655','2019-09-27 14:34:58.655')
,('TC Văn hóa-Thế thao và Du lịch','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.656','2019-09-27 14:34:58.656')
,('THPT Chuyên Bắc Giang','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.656','2019-09-27 14:34:58.656')
,('THPT Dân lập Hồ Tùng Mậu','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.656','2019-09-27 14:34:58.656')
,('THPT Dân lập Nguyên Hồng','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.657','2019-09-27 14:34:58.657')
,('THPT DTNT tỉnh','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.657','2019-09-27 14:34:58.657')
,('THPT Giáp Hải','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.659','2019-09-27 14:34:58.659')
,('THPT Ngô Sỹ Liên','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.659','2019-09-27 14:34:58.659')
,('THPT Thái Thuận','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.660','2019-09-27 14:34:58.660')
,('Tiểu học, THCS, THPT Thu Hương','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.660','2019-09-27 14:34:58.660')
,('Tr CĐ Công nghệ Việt Hàn Bắc Giang','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.661','2019-09-27 14:34:58.661')
,('TT GDTX tỉnh','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.663','2019-09-27 14:34:58.663')
,('TT Ngoại ngữ-Tin học BG','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.664','2019-09-27 14:34:58.664')
,('TTGD KTTH Hướng nghiệp','COUNTRY_VN',35,418,NULL,true,'2019-09-27 14:34:58.665','2019-09-27 14:34:58.665')
,('THPT Ba Bể','COUNTRY_VN',36,419,NULL,true,'2019-09-27 14:34:58.666','2019-09-27 14:34:58.666')
,('THPT Quảng Khê','COUNTRY_VN',36,419,NULL,true,'2019-09-27 14:34:58.667','2019-09-27 14:34:58.667')
,('TT GDTX H. Ba Bể tỉnh Bắc Kạn','COUNTRY_VN',36,419,NULL,true,'2019-09-27 14:34:58.667','2019-09-27 14:34:58.667')
,('TC nghề Bắc Kạn','COUNTRY_VN',36,420,NULL,true,'2019-09-27 14:34:58.668','2019-09-27 14:34:58.668')
,('THPT Bắc Kạn','COUNTRY_VN',36,420,NULL,true,'2019-09-27 14:34:58.669','2019-09-27 14:34:58.669')
,('THPT Chuyên Bắc Kạn','COUNTRY_VN',36,420,NULL,true,'2019-09-27 14:34:58.669','2019-09-27 14:34:58.669')
,('THPT Dân lập Hùng Vương','COUNTRY_VN',36,420,NULL,true,'2019-09-27 14:34:58.669','2019-09-27 14:34:58.669')
,('TT GDTX tỉnh Bắc Kạn','COUNTRY_VN',36,420,NULL,true,'2019-09-27 14:34:58.670','2019-09-27 14:34:58.670')
,('TT Kỹ thuật TH-HN Bắc Kạn','COUNTRY_VN',36,420,NULL,true,'2019-09-27 14:34:58.670','2019-09-27 14:34:58.670')
,('THPT Phủ Thông','COUNTRY_VN',36,421,NULL,true,'2019-09-27 14:34:58.671','2019-09-27 14:34:58.671')
,('TT GDTX H. Bạch Thông, tỉnh Bắc Kạn','COUNTRY_VN',36,421,NULL,true,'2019-09-27 14:34:58.671','2019-09-27 14:34:58.671')
,('THPT Bình Trung','COUNTRY_VN',36,422,NULL,true,'2019-09-27 14:34:58.672','2019-09-27 14:34:58.672')
,('THPT Chợ Đồn','COUNTRY_VN',36,422,NULL,true,'2019-09-27 14:34:58.672','2019-09-27 14:34:58.672')
,('TT GDTX -DN H. chợ Đồn, tỉnh Bắc Kạn','COUNTRY_VN',36,422,NULL,true,'2019-09-27 14:34:58.673','2019-09-27 14:34:58.673')
,('THPT Chợ Mới','COUNTRY_VN',36,423,NULL,true,'2019-09-27 14:34:58.674','2019-09-27 14:34:58.674')
,('THPT Yên Hân','COUNTRY_VN',36,423,NULL,true,'2019-09-27 14:34:58.675','2019-09-27 14:34:58.675')
,('TT GDTX H. Chợ Mới, tỉnh Bắc Kạn','COUNTRY_VN',36,423,NULL,true,'2019-09-27 14:34:58.676','2019-09-27 14:34:58.676')
,('THPT Na Rì','COUNTRY_VN',36,424,NULL,true,'2019-09-27 14:34:58.677','2019-09-27 14:34:58.677')
,('TT GDTX H. Na Rì. tỉnh Băc Kạn','COUNTRY_VN',36,424,NULL,true,'2019-09-27 14:34:58.678','2019-09-27 14:34:58.678')
,('THPT Nà Phặc','COUNTRY_VN',36,425,NULL,true,'2019-09-27 14:34:58.682','2019-09-27 14:34:58.682')
,('THPT Ngân Sơn','COUNTRY_VN',36,425,NULL,true,'2019-09-27 14:34:58.682','2019-09-27 14:34:58.682')
,('TT GDTX H. Ngân Sơn, tỉnh Bắc Kạn','COUNTRY_VN',36,425,NULL,true,'2019-09-27 14:34:58.683','2019-09-27 14:34:58.683')
,('THPT Bộc Bố','COUNTRY_VN',36,426,NULL,true,'2019-09-27 14:34:58.684','2019-09-27 14:34:58.684')
,('TT GDTX H. Pác Nặm, tỉnh Băc Kạn','COUNTRY_VN',36,426,NULL,true,'2019-09-27 14:34:58.685','2019-09-27 14:34:58.685')
,('THPT Điền Hải','COUNTRY_VN',37,427,NULL,true,'2019-09-27 14:34:58.686','2019-09-27 14:34:58.686')
,('THPT Định Thành','COUNTRY_VN',37,427,NULL,true,'2019-09-27 14:34:58.687','2019-09-27 14:34:58.687')
,('THPT Gành Hào','COUNTRY_VN',37,427,NULL,true,'2019-09-27 14:34:58.687','2019-09-27 14:34:58.687')
,('Trung tâm GD&DN Đông Hải','COUNTRY_VN',37,427,NULL,true,'2019-09-27 14:34:58.688','2019-09-27 14:34:58.688')
,('THPT Giá Rai','COUNTRY_VN',37,428,NULL,true,'2019-09-27 14:34:58.689','2019-09-27 14:34:58.689')
,('THPT Nguyễn Trung Trục','COUNTRY_VN',37,428,NULL,true,'2019-09-27 14:34:58.689','2019-09-27 14:34:58.689')
,('THPT Tân Phong','COUNTRY_VN',37,428,NULL,true,'2019-09-27 14:34:58.690','2019-09-27 14:34:58.690')
,('Trung tâm GD&DN Giá Rai','COUNTRY_VN',37,428,NULL,true,'2019-09-27 14:34:58.691','2019-09-27 14:34:58.691')
,('Phổ thông Dân tộc Nội trú tỉnh Bạc Liêu','COUNTRY_VN',37,429,NULL,true,'2019-09-27 14:34:58.693','2019-09-27 14:34:58.693')
,('THPT Lê Thị Riêng','COUNTRY_VN',37,429,NULL,true,'2019-09-27 14:34:58.694','2019-09-27 14:34:58.694')
,('Trung tâm GD&DN Hòa Bình','COUNTRY_VN',37,429,NULL,true,'2019-09-27 14:34:58.695','2019-09-27 14:34:58.695')
,('THPT Ngan Dừa','COUNTRY_VN',37,430,NULL,true,'2019-09-27 14:34:58.697','2019-09-27 14:34:58.697')
,('THPT Ninh Quới','COUNTRY_VN',37,430,NULL,true,'2019-09-27 14:34:58.698','2019-09-27 14:34:58.698')
,('THPT Ninh Thạnh Lợi','COUNTRY_VN',37,430,NULL,true,'2019-09-27 14:34:58.699','2019-09-27 14:34:58.699')
,('Trung tâm GD&DN Hồng Dân','COUNTRY_VN',37,430,NULL,true,'2019-09-27 14:34:58.699','2019-09-27 14:34:58.699')
,('THPT Trần Văn Bảy','COUNTRY_VN',37,431,NULL,true,'2019-09-27 14:34:58.700','2019-09-27 14:34:58.700')
,('THPT Võ Văn Kiệt','COUNTRY_VN',37,431,NULL,true,'2019-09-27 14:34:58.701','2019-09-27 14:34:58.701')
,('Trung tâm GD&DN Phước Long','COUNTRY_VN',37,431,NULL,true,'2019-09-27 14:34:58.701','2019-09-27 14:34:58.701')
,('THPT Lê Văn Đẩu','COUNTRY_VN',37,432,NULL,true,'2019-09-27 14:34:58.702','2019-09-27 14:34:58.702')
,('THPT Vĩnh Hưng','COUNTRY_VN',37,432,NULL,true,'2019-09-27 14:34:58.703','2019-09-27 14:34:58.703')
,('Trung tâm GD&DN Vĩnh Lợi','COUNTRY_VN',37,432,NULL,true,'2019-09-27 14:34:58.703','2019-09-27 14:34:58.703')
,('THCS & THPT Trần văn Lắm','COUNTRY_VN',37,433,NULL,true,'2019-09-27 14:34:58.704','2019-09-27 14:34:58.704')
,('THPT Bạc Liêu','COUNTRY_VN',37,433,NULL,true,'2019-09-27 14:34:58.705','2019-09-27 14:34:58.705')
,('THPT Chuyên Bạc Liêu','COUNTRY_VN',37,433,NULL,true,'2019-09-27 14:34:58.705','2019-09-27 14:34:58.705')
,('THPT Hiệp Thành','COUNTRY_VN',37,433,NULL,true,'2019-09-27 14:34:58.705','2019-09-27 14:34:58.705')
,('THPT Phan Ngọc Hiển','COUNTRY_VN',37,433,NULL,true,'2019-09-27 14:34:58.706','2019-09-27 14:34:58.706')
,('TT GDTX tỉnh Bạc Liêu','COUNTRY_VN',37,433,NULL,true,'2019-09-27 14:34:58.706','2019-09-27 14:34:58.706')
,('THPT Giá Rai','COUNTRY_VN',37,434,NULL,true,'2019-09-27 14:34:58.708','2019-09-27 14:34:58.708')
,('THPT Nguyễn Trung Trục','COUNTRY_VN',37,434,NULL,true,'2019-09-27 14:34:58.711','2019-09-27 14:34:58.711')
,('THPT Tân Phong','COUNTRY_VN',37,434,NULL,true,'2019-09-27 14:34:58.712','2019-09-27 14:34:58.712')
,('Trung tâm GD&DN Giá Rai','COUNTRY_VN',37,434,NULL,true,'2019-09-27 14:34:58.712','2019-09-27 14:34:58.712')
,('THPT Gia Bình 1','COUNTRY_VN',38,435,NULL,true,'2019-09-27 14:34:58.715','2019-09-27 14:34:58.715')
,('THPT Gia Bình 3','COUNTRY_VN',38,435,NULL,true,'2019-09-27 14:34:58.716','2019-09-27 14:34:58.716')
,('THPT Lê Văn Thịnh','COUNTRY_VN',38,435,NULL,true,'2019-09-27 14:34:58.717','2019-09-27 14:34:58.717')
,('TT GDTX Gia Bình','COUNTRY_VN',38,435,NULL,true,'2019-09-27 14:34:58.717','2019-09-27 14:34:58.717')
,('THPT Hải á','COUNTRY_VN',38,436,NULL,true,'2019-09-27 14:34:58.718','2019-09-27 14:34:58.718')
,('THPT Lương Tài 1','COUNTRY_VN',38,436,NULL,true,'2019-09-27 14:34:58.719','2019-09-27 14:34:58.719')
,('THPT Lương Tài 2','COUNTRY_VN',38,436,NULL,true,'2019-09-27 14:34:58.719','2019-09-27 14:34:58.719')
,('THPT Lương Tài 3','COUNTRY_VN',38,436,NULL,true,'2019-09-27 14:34:58.720','2019-09-27 14:34:58.720')
,('TT GDTX Lương Tài','COUNTRY_VN',38,436,NULL,true,'2019-09-27 14:34:58.720','2019-09-27 14:34:58.720')
,('THPT Phố Mới','COUNTRY_VN',38,437,NULL,true,'2019-09-27 14:34:58.721','2019-09-27 14:34:58.721')
,('THPT Quế Võ 1','COUNTRY_VN',38,437,NULL,true,'2019-09-27 14:34:58.722','2019-09-27 14:34:58.722')
,('THPT Quế Võ 2','COUNTRY_VN',38,437,NULL,true,'2019-09-27 14:34:58.722','2019-09-27 14:34:58.722')
,('THPT Quế Võ 3','COUNTRY_VN',38,437,NULL,true,'2019-09-27 14:34:58.722','2019-09-27 14:34:58.722')
,('THPT Trần Hung Đạo','COUNTRY_VN',38,437,NULL,true,'2019-09-27 14:34:58.723','2019-09-27 14:34:58.723')
,('TT GDTX số 2 tỉnh Bắc Ninh','COUNTRY_VN',38,437,NULL,true,'2019-09-27 14:34:58.723','2019-09-27 14:34:58.723')
,('THPT Kinh Bắc','COUNTRY_VN',38,438,NULL,true,'2019-09-27 14:34:58.728','2019-09-27 14:34:58.728')
,('THPT Thiên Đức','COUNTRY_VN',38,438,NULL,true,'2019-09-27 14:34:58.729','2019-09-27 14:34:58.729')
,('THPT Thuận Thành 1','COUNTRY_VN',38,438,NULL,true,'2019-09-27 14:34:58.729','2019-09-27 14:34:58.729')
,('THPT Thuận Thành 2','COUNTRY_VN',38,438,NULL,true,'2019-09-27 14:34:58.731','2019-09-27 14:34:58.731')
,('THPT Thuận Thành 3','COUNTRY_VN',38,438,NULL,true,'2019-09-27 14:34:58.733','2019-09-27 14:34:58.733')
,('TT GDTX Thuận Thành','COUNTRY_VN',38,438,NULL,true,'2019-09-27 14:34:58.733','2019-09-27 14:34:58.733')
,('THPT Lê Quý Đôn','COUNTRY_VN',38,439,NULL,true,'2019-09-27 14:34:58.735','2019-09-27 14:34:58.735')
,('THPT Nguyễn Đăng Đạo','COUNTRY_VN',38,439,NULL,true,'2019-09-27 14:34:58.735','2019-09-27 14:34:58.735')
,('THPT Tiên Du 1','COUNTRY_VN',38,439,NULL,true,'2019-09-27 14:34:58.736','2019-09-27 14:34:58.736')
,('THPT Trần Nhân Tông','COUNTRY_VN',38,439,NULL,true,'2019-09-27 14:34:58.737','2019-09-27 14:34:58.737')
,('TT GDTX Tiên Du','COUNTRY_VN',38,439,NULL,true,'2019-09-27 14:34:58.737','2019-09-27 14:34:58.737')
,('THPT Nguyễn Trãi','COUNTRY_VN',38,440,NULL,true,'2019-09-27 14:34:58.738','2019-09-27 14:34:58.738')
,('THPT Yên Phong 1','COUNTRY_VN',38,440,NULL,true,'2019-09-27 14:34:58.739','2019-09-27 14:34:58.739')
,('THPT Yên Phong 2','COUNTRY_VN',38,440,NULL,true,'2019-09-27 14:34:58.739','2019-09-27 14:34:58.739')
,('TT GDTX Yên Phong','COUNTRY_VN',38,440,NULL,true,'2019-09-27 14:34:58.739','2019-09-27 14:34:58.739')
,('CĐ Nghề Cơ điện Xây dựng Bắc Ninh','COUNTRY_VN',38,441,NULL,true,'2019-09-27 14:34:58.742','2019-09-27 14:34:58.742')
,('CĐ Nghề Kinh tế Kỳ thuật Băc Ninh','COUNTRY_VN',38,441,NULL,true,'2019-09-27 14:34:58.744','2019-09-27 14:34:58.744')
,('PT có nhiều cấp học Quốc tế Kinh Băc','COUNTRY_VN',38,441,NULL,true,'2019-09-27 14:34:58.745','2019-09-27 14:34:58.745')
,('TC nghề KT KT Liên đoàn Lao động','COUNTRY_VN',38,441,NULL,true,'2019-09-27 14:34:58.745','2019-09-27 14:34:58.745')
,('THPT Chuyên Bắc Ninh','COUNTRY_VN',38,441,NULL,true,'2019-09-27 14:34:58.746','2019-09-27 14:34:58.746')
,('THPT Hàm Long','COUNTRY_VN',38,441,NULL,true,'2019-09-27 14:34:58.747','2019-09-27 14:34:58.747')
,('THPT Hàn Thuyên','COUNTRY_VN',38,441,NULL,true,'2019-09-27 14:34:58.748','2019-09-27 14:34:58.748')
,('THPT Hoàng Quốc việt','COUNTRY_VN',38,441,NULL,true,'2019-09-27 14:34:58.749','2019-09-27 14:34:58.749')
,('THPT Lý Nhân Tông','COUNTRY_VN',38,441,NULL,true,'2019-09-27 14:34:58.750','2019-09-27 14:34:58.750')
,('THPT Lý Thường Kiệt','COUNTRY_VN',38,441,NULL,true,'2019-09-27 14:34:58.750','2019-09-27 14:34:58.750')
,('THPT Nguyễn Du','COUNTRY_VN',38,441,NULL,true,'2019-09-27 14:34:58.751','2019-09-27 14:34:58.751')
,('TT GDTX tỉnh Bắc Ninh','COUNTRY_VN',38,441,NULL,true,'2019-09-27 14:34:58.751','2019-09-27 14:34:58.751')
,('CĐ Công nghiệp Hưng Yên (cơ sở 2)','COUNTRY_VN',38,442,NULL,true,'2019-09-27 14:34:58.753','2019-09-27 14:34:58.753')
,('CĐ Thủy sản','COUNTRY_VN',38,442,NULL,true,'2019-09-27 14:34:58.753','2019-09-27 14:34:58.753')
,('PT năng khiếu TDTT Olympic','COUNTRY_VN',38,442,NULL,true,'2019-09-27 14:34:58.754','2019-09-27 14:34:58.754')
,('THPT Lý Thái Tổ','COUNTRY_VN',38,442,NULL,true,'2019-09-27 14:34:58.754','2019-09-27 14:34:58.754')
,('THPT Ngô Gia Tự','COUNTRY_VN',38,442,NULL,true,'2019-09-27 14:34:58.754','2019-09-27 14:34:58.754')
,('THPT Nguyễn Văn Cừ','COUNTRY_VN',38,442,NULL,true,'2019-09-27 14:34:58.755','2019-09-27 14:34:58.755')
,('THPT Từ Sơn','COUNTRY_VN',38,442,NULL,true,'2019-09-27 14:34:58.755','2019-09-27 14:34:58.755')
,('THPT Bán công Ba Tri','COUNTRY_VN',39,443,NULL,true,'2019-09-27 14:34:58.757','2019-09-27 14:34:58.757')
,('THPT Phan Liêm','COUNTRY_VN',39,443,NULL,true,'2019-09-27 14:34:58.758','2019-09-27 14:34:58.758')
,('THPT Phan Ngọc Tòng','COUNTRY_VN',39,443,NULL,true,'2019-09-27 14:34:58.759','2019-09-27 14:34:58.759')
,('THPT Phan Thanh Giản','COUNTRY_VN',39,443,NULL,true,'2019-09-27 14:34:58.760','2019-09-27 14:34:58.760')
,('THPT Sương Nguyệt Ánh','COUNTRY_VN',39,443,NULL,true,'2019-09-27 14:34:58.761','2019-09-27 14:34:58.761')
,('THPT Tán Kế','COUNTRY_VN',39,443,NULL,true,'2019-09-27 14:34:58.761','2019-09-27 14:34:58.761')
,('Trung tâm GDTX Ba Tri','COUNTRY_VN',39,443,NULL,true,'2019-09-27 14:34:58.762','2019-09-27 14:34:58.762')
,('THPT Bán công Bình Đại','COUNTRY_VN',39,444,NULL,true,'2019-09-27 14:34:58.765','2019-09-27 14:34:58.765')
,('THPT Bán công Lộc Thuận','COUNTRY_VN',39,444,NULL,true,'2019-09-27 14:34:58.766','2019-09-27 14:34:58.766')
,('THPT Huỳnh Tấn Phát','COUNTRY_VN',39,444,NULL,true,'2019-09-27 14:34:58.766','2019-09-27 14:34:58.766')
,('THPT Lê Hoàng chiếu','COUNTRY_VN',39,444,NULL,true,'2019-09-27 14:34:58.766','2019-09-27 14:34:58.766')
,('THPT Lê Qúy Đôn','COUNTRY_VN',39,444,NULL,true,'2019-09-27 14:34:58.767','2019-09-27 14:34:58.767')
,('Trung tâm GDTX Bình Đại','COUNTRY_VN',39,444,NULL,true,'2019-09-27 14:34:58.767','2019-09-27 14:34:58.767')
,('THPT BC Châu Thành A','COUNTRY_VN',39,445,NULL,true,'2019-09-27 14:34:58.768','2019-09-27 14:34:58.768')
,('THPT BC Châu Thành B','COUNTRY_VN',39,445,NULL,true,'2019-09-27 14:34:58.769','2019-09-27 14:34:58.769')
,('THPT Diệp Minh châu','COUNTRY_VN',39,445,NULL,true,'2019-09-27 14:34:58.769','2019-09-27 14:34:58.769')
,('THPT Mạc Đĩnh Chi','COUNTRY_VN',39,445,NULL,true,'2019-09-27 14:34:58.770','2019-09-27 14:34:58.770')
,('THPT Nguyễn Huệ','COUNTRY_VN',39,445,NULL,true,'2019-09-27 14:34:58.770','2019-09-27 14:34:58.770')
,('THPT Trần Văn ơn','COUNTRY_VN',39,445,NULL,true,'2019-09-27 14:34:58.771','2019-09-27 14:34:58.771')
,('Trung tâm GDTX Châu Thành','COUNTRY_VN',39,445,NULL,true,'2019-09-27 14:34:58.771','2019-09-27 14:34:58.771')
,('THPT Bán công Chợ Lách','COUNTRY_VN',39,446,NULL,true,'2019-09-27 14:34:58.772','2019-09-27 14:34:58.772')
,('THPT Bán công Vĩnh Thành','COUNTRY_VN',39,446,NULL,true,'2019-09-27 14:34:58.773','2019-09-27 14:34:58.773')
,('THPT Trần Văn Kiết','COUNTRY_VN',39,446,NULL,true,'2019-09-27 14:34:58.773','2019-09-27 14:34:58.773')
,('THPT Trương Vĩnh Ký','COUNTRY_VN',39,446,NULL,true,'2019-09-27 14:34:58.774','2019-09-27 14:34:58.774')
,('Trung tâm GDTX Chợ Lách','COUNTRY_VN',39,446,NULL,true,'2019-09-27 14:34:58.775','2019-09-27 14:34:58.775')
,('THPT Bán công Giồng Trôm','COUNTRY_VN',39,447,NULL,true,'2019-09-27 14:34:58.777','2019-09-27 14:34:58.777')
,('THPT Dân lập Giồng Trôm','COUNTRY_VN',39,447,NULL,true,'2019-09-27 14:34:58.777','2019-09-27 14:34:58.777')
,('THPT Nguyễn Ngọc Thăng','COUNTRY_VN',39,447,NULL,true,'2019-09-27 14:34:58.778','2019-09-27 14:34:58.778')
,('THPT Nguyễn Thị Định','COUNTRY_VN',39,447,NULL,true,'2019-09-27 14:34:58.778','2019-09-27 14:34:58.778')
,('THPT Nguyễn Trãi','COUNTRY_VN',39,447,NULL,true,'2019-09-27 14:34:58.779','2019-09-27 14:34:58.779')
,('THPT Phan Văn Trị','COUNTRY_VN',39,447,NULL,true,'2019-09-27 14:34:58.781','2019-09-27 14:34:58.781')
,('Trung tâm GDTX huyện Giồng Trôm','COUNTRY_VN',39,447,NULL,true,'2019-09-27 14:34:58.782','2019-09-27 14:34:58.782')
,('THPT Bán công Phước Mỹ Trung','COUNTRY_VN',39,448,NULL,true,'2019-09-27 14:34:58.784','2019-09-27 14:34:58.784')
,('THPT Lê Anh Xuân','COUNTRY_VN',39,448,NULL,true,'2019-09-27 14:34:58.784','2019-09-27 14:34:58.784')
,('THPT Ngố Văn cấn','COUNTRY_VN',39,448,NULL,true,'2019-09-27 14:34:58.785','2019-09-27 14:34:58.785')
,('Trung tâm GDTX Mỏ Cày Bắc','COUNTRY_VN',39,448,NULL,true,'2019-09-27 14:34:58.785','2019-09-27 14:34:58.785')
,('THPT Bán công Mỏ Cày','COUNTRY_VN',39,449,NULL,true,'2019-09-27 14:34:58.786','2019-09-27 14:34:58.786')
,('THPT Ca Văn Thỉnh','COUNTRY_VN',39,449,NULL,true,'2019-09-27 14:34:58.787','2019-09-27 14:34:58.787')
,('THPT Chê-Ghêvara','COUNTRY_VN',39,449,NULL,true,'2019-09-27 14:34:58.787','2019-09-27 14:34:58.787')
,('THPT Nguyễn Thị Minh Khai','COUNTRY_VN',39,449,NULL,true,'2019-09-27 14:34:58.787','2019-09-27 14:34:58.787')
,('THPT Quản Trọng Hoàng','COUNTRY_VN',39,449,NULL,true,'2019-09-27 14:34:58.788','2019-09-27 14:34:58.788')
,('Trung tâm GDTX huyện Mỏ Cày Nam','COUNTRY_VN',39,449,NULL,true,'2019-09-27 14:34:58.789','2019-09-27 14:34:58.789')
,('THPT Bán công Thạnh Phú','COUNTRY_VN',39,450,NULL,true,'2019-09-27 14:34:58.790','2019-09-27 14:34:58.790')
,('THPT Đoàn Thị Điểm','COUNTRY_VN',39,450,NULL,true,'2019-09-27 14:34:58.790','2019-09-27 14:34:58.790')
,('THPT Lê Hoài Đôn','COUNTRY_VN',39,450,NULL,true,'2019-09-27 14:34:58.792','2019-09-27 14:34:58.792')
,('THPT Trần Trường Sinh','COUNTRY_VN',39,450,NULL,true,'2019-09-27 14:34:58.793','2019-09-27 14:34:58.793')
,('Trung tâm GDTX Thạnh Phú','COUNTRY_VN',39,450,NULL,true,'2019-09-27 14:34:58.794','2019-09-27 14:34:58.794')
,('CĐ nghề Đồng Khởi','COUNTRY_VN',39,451,NULL,true,'2019-09-27 14:34:58.795','2019-09-27 14:34:58.795')
,('Năng khiếu TDTT Bến Tre','COUNTRY_VN',39,451,NULL,true,'2019-09-27 14:34:58.797','2019-09-27 14:34:58.797')
,('Phổ thõng Hermann Gmeiner','COUNTRY_VN',39,451,NULL,true,'2019-09-27 14:34:58.798','2019-09-27 14:34:58.798')
,('TC nghề Bến Tre','COUNTRY_VN',39,451,NULL,true,'2019-09-27 14:34:58.799','2019-09-27 14:34:58.799')
,('THPT Bán công Thị Xã','COUNTRY_VN',39,451,NULL,true,'2019-09-27 14:34:58.800','2019-09-27 14:34:58.800')
,('THPT Chuyên Bến Tre','COUNTRY_VN',39,451,NULL,true,'2019-09-27 14:34:58.801','2019-09-27 14:34:58.801')
,('THPT Lạc Long Quân','COUNTRY_VN',39,451,NULL,true,'2019-09-27 14:34:58.801','2019-09-27 14:34:58.801')
,('THPT Nguyễn Đình Chiểu','COUNTRY_VN',39,451,NULL,true,'2019-09-27 14:34:58.802','2019-09-27 14:34:58.802')
,('THPT Võ Trường Toản','COUNTRY_VN',39,451,NULL,true,'2019-09-27 14:34:58.802','2019-09-27 14:34:58.802')
,('Trung cấp Y Tế Bến Tre','COUNTRY_VN',39,451,NULL,true,'2019-09-27 14:34:58.803','2019-09-27 14:34:58.803')
,('Trung tâm GDTX thành phố Bến Tre','COUNTRY_VN',39,451,NULL,true,'2019-09-27 14:34:58.803','2019-09-27 14:34:58.803')
,('THPT An Lão','COUNTRY_VN',40,452,NULL,true,'2019-09-27 14:34:58.804','2019-09-27 14:34:58.804')
,('THPT Số 2 An Lão','COUNTRY_VN',40,452,NULL,true,'2019-09-27 14:34:58.805','2019-09-27 14:34:58.805')
,('TT GDTX-HN An Lão','COUNTRY_VN',40,452,NULL,true,'2019-09-27 14:34:58.805','2019-09-27 14:34:58.805')
,('THPT Hoài Ân','COUNTRY_VN',40,453,NULL,true,'2019-09-27 14:34:58.806','2019-09-27 14:34:58.806')
,('THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',40,453,NULL,true,'2019-09-27 14:34:58.806','2019-09-27 14:34:58.806')
,('THPT Trần Quang Diệu','COUNTRY_VN',40,453,NULL,true,'2019-09-27 14:34:58.807','2019-09-27 14:34:58.807')
,('THPT Võ Giữ','COUNTRY_VN',40,453,NULL,true,'2019-09-27 14:34:58.810','2019-09-27 14:34:58.810')
,('TT GDTX-HN Hoài Ân','COUNTRY_VN',40,453,NULL,true,'2019-09-27 14:34:58.811','2019-09-27 14:34:58.811')
,('THPT Nguyễn Du','COUNTRY_VN',40,454,NULL,true,'2019-09-27 14:34:58.812','2019-09-27 14:34:58.812')
,('THPT Nguyễn Trân','COUNTRY_VN',40,454,NULL,true,'2019-09-27 14:34:58.813','2019-09-27 14:34:58.813')
,('THPT Phan Bội Châu','COUNTRY_VN',40,454,NULL,true,'2019-09-27 14:34:58.814','2019-09-27 14:34:58.814')
,('THPT Tam Quan','COUNTRY_VN',40,454,NULL,true,'2019-09-27 14:34:58.815','2019-09-27 14:34:58.815')
,('THPT Tăng Bạt Hổ','COUNTRY_VN',40,454,NULL,true,'2019-09-27 14:34:58.816','2019-09-27 14:34:58.816')
,('TT GDTX-HN Hoài Nhơn','COUNTRY_VN',40,454,NULL,true,'2019-09-27 14:34:58.817','2019-09-27 14:34:58.817')
,('THPT Nguyễn Hồng Đạo','COUNTRY_VN',40,455,NULL,true,'2019-09-27 14:34:58.818','2019-09-27 14:34:58.818')
,('THPT Nguyễn Hũu Quang','COUNTRY_VN',40,455,NULL,true,'2019-09-27 14:34:58.818','2019-09-27 14:34:58.818')
,('THPT SỐ1 Phù Cát','COUNTRY_VN',40,455,NULL,true,'2019-09-27 14:34:58.819','2019-09-27 14:34:58.819')
,('THPT Số 2 Phù Cát','COUNTRY_VN',40,455,NULL,true,'2019-09-27 14:34:58.819','2019-09-27 14:34:58.819')
,('THPT Số 3 Phù Cát','COUNTRY_VN',40,455,NULL,true,'2019-09-27 14:34:58.820','2019-09-27 14:34:58.820')
,('TT GDTX-HN Phù cát','COUNTRY_VN',40,455,NULL,true,'2019-09-27 14:34:58.820','2019-09-27 14:34:58.820')
,('THPT An Lương','COUNTRY_VN',40,456,NULL,true,'2019-09-27 14:34:58.821','2019-09-27 14:34:58.821')
,('THPT Bình Dương','COUNTRY_VN',40,456,NULL,true,'2019-09-27 14:34:58.821','2019-09-27 14:34:58.821')
,('THPT Mỹ Thọ','COUNTRY_VN',40,456,NULL,true,'2019-09-27 14:34:58.822','2019-09-27 14:34:58.822')
,('THPT Nguyễn Trung Trực','COUNTRY_VN',40,456,NULL,true,'2019-09-27 14:34:58.822','2019-09-27 14:34:58.822')
,('THPT SỐ1 Phù Mỹ','COUNTRY_VN',40,456,NULL,true,'2019-09-27 14:34:58.823','2019-09-27 14:34:58.823')
,('THPT Số 2 Phù Mỹ','COUNTRY_VN',40,456,NULL,true,'2019-09-27 14:34:58.823','2019-09-27 14:34:58.823')
,('TT GDTX-HN Phù Mỹ','COUNTRY_VN',40,456,NULL,true,'2019-09-27 14:34:58.825','2019-09-27 14:34:58.825')
,('THPT Nguyễn Huệ','COUNTRY_VN',40,457,NULL,true,'2019-09-27 14:34:58.827','2019-09-27 14:34:58.827')
,('THPT Quang Trung','COUNTRY_VN',40,457,NULL,true,'2019-09-27 14:34:58.827','2019-09-27 14:34:58.827')
,('THPT Tây Sơn','COUNTRY_VN',40,457,NULL,true,'2019-09-27 14:34:58.829','2019-09-27 14:34:58.829')
,('THPT Võ Lai','COUNTRY_VN',40,457,NULL,true,'2019-09-27 14:34:58.830','2019-09-27 14:34:58.830')
,('TT GDTX-HN Tây Sơn','COUNTRY_VN',40,457,NULL,true,'2019-09-27 14:34:58.832','2019-09-27 14:34:58.832')
,('THPT Nguyễn Diêu','COUNTRY_VN',40,458,NULL,true,'2019-09-27 14:34:58.833','2019-09-27 14:34:58.833')
,('THPT SỐ 1 Tuy Phước','COUNTRY_VN',40,458,NULL,true,'2019-09-27 14:34:58.834','2019-09-27 14:34:58.834')
,('THPT Số 2 Tuy Phước','COUNTRY_VN',40,458,NULL,true,'2019-09-27 14:34:58.835','2019-09-27 14:34:58.835')
,('THPT Xuân Diệu','COUNTRY_VN',40,458,NULL,true,'2019-09-27 14:34:58.835','2019-09-27 14:34:58.835')
,('TT GDTX-HN Tuy Phước','COUNTRY_VN',40,458,NULL,true,'2019-09-27 14:34:58.836','2019-09-27 14:34:58.836')
,('THPT DTNT vân Canh','COUNTRY_VN',40,459,NULL,true,'2019-09-27 14:34:58.837','2019-09-27 14:34:58.837')
,('THPT Vân Canh','COUNTRY_VN',40,459,NULL,true,'2019-09-27 14:34:58.837','2019-09-27 14:34:58.837')
,('TT GDTX-HN vân Canh','COUNTRY_VN',40,459,NULL,true,'2019-09-27 14:34:58.838','2019-09-27 14:34:58.838')
,('THPT DTNT Vĩnh Thạnh','COUNTRY_VN',40,460,NULL,true,'2019-09-27 14:34:58.839','2019-09-27 14:34:58.839')
,('THPT Vĩnh Thạnh','COUNTRY_VN',40,460,NULL,true,'2019-09-27 14:34:58.839','2019-09-27 14:34:58.839')
,('CĐ nghề cơ điện xây dụng và Nông lâm','COUNTRY_VN',40,461,NULL,true,'2019-09-27 14:34:58.842','2019-09-27 14:34:58.842')
,('CĐ nghề Quy Nhơn','COUNTRY_VN',40,461,NULL,true,'2019-09-27 14:34:58.843','2019-09-27 14:34:58.843')
,('THPT Chu Văn An','COUNTRY_VN',40,461,NULL,true,'2019-09-27 14:34:58.843','2019-09-27 14:34:58.843')
,('THPT chuyên Lê Quý Đôn','COUNTRY_VN',40,461,NULL,true,'2019-09-27 14:34:58.844','2019-09-27 14:34:58.844')
,('THPT DTNT Tỉnh','COUNTRY_VN',40,461,NULL,true,'2019-09-27 14:34:58.845','2019-09-27 14:34:58.845')
,('THPT Hùng Vương','COUNTRY_VN',40,461,NULL,true,'2019-09-27 14:34:58.847','2019-09-27 14:34:58.847')
,('THPT Nguyễn Thái Học','COUNTRY_VN',40,461,NULL,true,'2019-09-27 14:34:58.848','2019-09-27 14:34:58.848')
,('THPT Hòa Bình','COUNTRY_VN',40,462,NULL,true,'2019-09-27 14:34:58.849','2019-09-27 14:34:58.849')
,('THPT Nguyễn Đình Chiểu','COUNTRY_VN',40,462,NULL,true,'2019-09-27 14:34:58.849','2019-09-27 14:34:58.849')
,('THPT Nguyễn Trường Tộ','COUNTRY_VN',40,462,NULL,true,'2019-09-27 14:34:58.850','2019-09-27 14:34:58.850')
,('THPT SỐ1 An Nhơn','COUNTRY_VN',40,462,NULL,true,'2019-09-27 14:34:58.850','2019-09-27 14:34:58.850')
,('THPT Số 2 An Nhơn','COUNTRY_VN',40,462,NULL,true,'2019-09-27 14:34:58.851','2019-09-27 14:34:58.851')
,('THPT Số 3 An Nhơn','COUNTRY_VN',40,462,NULL,true,'2019-09-27 14:34:58.851','2019-09-27 14:34:58.851')
,('TT GDTX-HN An Nhơn','COUNTRY_VN',40,462,NULL,true,'2019-09-27 14:34:58.851','2019-09-27 14:34:58.851')
,('THPT Lý Tự Trọng','COUNTRY_VN',40,463,NULL,true,'2019-09-27 14:34:58.852','2019-09-27 14:34:58.852')
,('THPT Nguyễn Du','COUNTRY_VN',40,463,NULL,true,'2019-09-27 14:34:58.853','2019-09-27 14:34:58.853')
,('THPT Nguyễn Trân','COUNTRY_VN',40,463,NULL,true,'2019-09-27 14:34:58.853','2019-09-27 14:34:58.853')
,('THPT Phan Bội Châu','COUNTRY_VN',40,463,NULL,true,'2019-09-27 14:34:58.853','2019-09-27 14:34:58.853')
,('THPT Tam Quan','COUNTRY_VN',40,463,NULL,true,'2019-09-27 14:34:58.854','2019-09-27 14:34:58.854')
,('THPT Tăng Bạt Hổ','COUNTRY_VN',40,463,NULL,true,'2019-09-27 14:34:58.854','2019-09-27 14:34:58.854')
,('TT GDTX-HN Hoài Nhơn','COUNTRY_VN',40,463,NULL,true,'2019-09-27 14:34:58.854','2019-09-27 14:34:58.854')
,('THPT Lê Lợi','COUNTRY_VN',41,464,NULL,true,'2019-09-27 14:34:58.855','2019-09-27 14:34:58.855')
,('THPT Tân Bình','COUNTRY_VN',41,464,NULL,true,'2019-09-27 14:34:58.856','2019-09-27 14:34:58.856')
,('THPT Thường Tân','COUNTRY_VN',41,464,NULL,true,'2019-09-27 14:34:58.856','2019-09-27 14:34:58.856')
,('THPT Bàu Bàng','COUNTRY_VN',41,465,NULL,true,'2019-09-27 14:34:58.857','2019-09-27 14:34:58.857')
,('THPT Dầu Tiếng','COUNTRY_VN',41,466,NULL,true,'2019-09-27 14:34:58.859','2019-09-27 14:34:58.859')
,('THPT Phan Bội Châu','COUNTRY_VN',41,466,NULL,true,'2019-09-27 14:34:58.860','2019-09-27 14:34:58.860')
,('THPT Thanh Tuyền','COUNTRY_VN',41,466,NULL,true,'2019-09-27 14:34:58.860','2019-09-27 14:34:58.860')
,('TT GDTX - KTHN H. Dầu Tiếng','COUNTRY_VN',41,466,NULL,true,'2019-09-27 14:34:58.861','2019-09-27 14:34:58.861')
,('CĐN Công nghệ và NL Nam Bộ','COUNTRY_VN',41,467,NULL,true,'2019-09-27 14:34:58.862','2019-09-27 14:34:58.862')
,('CĐN Đông An','COUNTRY_VN',41,467,NULL,true,'2019-09-27 14:34:58.864','2019-09-27 14:34:58.864')
,('Phân hiệu CĐN Đường sắt phía Nam','COUNTRY_VN',41,467,NULL,true,'2019-09-27 14:34:58.865','2019-09-27 14:34:58.865')
,('TCN Dĩ An','COUNTRY_VN',41,467,NULL,true,'2019-09-27 14:34:58.865','2019-09-27 14:34:58.865')
,('TCN Khu Công nghiệp','COUNTRY_VN',41,467,NULL,true,'2019-09-27 14:34:58.866','2019-09-27 14:34:58.866')
,('THPT Bình An','COUNTRY_VN',41,467,NULL,true,'2019-09-27 14:34:58.866','2019-09-27 14:34:58.866')
,('THPT Dĩ An','COUNTRY_VN',41,467,NULL,true,'2019-09-27 14:34:58.867','2019-09-27 14:34:58.867')
,('THPT Nguyễn An Ninh','COUNTRY_VN',41,467,NULL,true,'2019-09-27 14:34:58.867','2019-09-27 14:34:58.867')
,('TT GDTX - KTHN TX. Dĩ An','COUNTRY_VN',41,467,NULL,true,'2019-09-27 14:34:58.867','2019-09-27 14:34:58.867')
,('Tư thục THPT Phan Chu Trinh','COUNTRY_VN',41,467,NULL,true,'2019-09-27 14:34:58.868','2019-09-27 14:34:58.868')
,('THPT Nguyễn Huệ','COUNTRY_VN',41,468,NULL,true,'2019-09-27 14:34:58.869','2019-09-27 14:34:58.869')
,('THPT Phước Vĩnh','COUNTRY_VN',41,468,NULL,true,'2019-09-27 14:34:58.869','2019-09-27 14:34:58.869')
,('THPT Tây Sơn','COUNTRY_VN',41,468,NULL,true,'2019-09-27 14:34:58.869','2019-09-27 14:34:58.869')
,('TT GDTX - KTHN H. Phú Giáo','COUNTRY_VN',41,468,NULL,true,'2019-09-27 14:34:58.870','2019-09-27 14:34:58.870')
,('TCN Nghiệp vụ Bình Dương','COUNTRY_VN',41,469,NULL,true,'2019-09-27 14:34:58.870','2019-09-27 14:34:58.870')
,('TCN tỉnh Bình Dương','COUNTRY_VN',41,469,NULL,true,'2019-09-27 14:34:58.871','2019-09-27 14:34:58.871')
,('TCN Việt Hàn Bình Dương','COUNTRY_VN',41,469,NULL,true,'2019-09-27 14:34:58.871','2019-09-27 14:34:58.871')
,('THPT An Mỹ','COUNTRY_VN',41,469,NULL,true,'2019-09-27 14:34:58.871','2019-09-27 14:34:58.871')
,('THPT Bình Phú','COUNTRY_VN',41,469,NULL,true,'2019-09-27 14:34:58.872','2019-09-27 14:34:58.872')
,('THPT chuyên Hùng Vương','COUNTRY_VN',41,469,NULL,true,'2019-09-27 14:34:58.872','2019-09-27 14:34:58.872')
,('THPT Nguyễn Đình Chiểu','COUNTRY_VN',41,469,NULL,true,'2019-09-27 14:34:58.872','2019-09-27 14:34:58.872')
,('THPT Võ Minh Đức','COUNTRY_VN',41,469,NULL,true,'2019-09-27 14:34:58.873','2019-09-27 14:34:58.873')
,('Trung tâm GDTX tỉnh Bình Dương','COUNTRY_VN',41,469,NULL,true,'2019-09-27 14:34:58.873','2019-09-27 14:34:58.873')
,('Tư thục THCS-THPT Nguyễn Khuyến','COUNTRY_VN',41,469,NULL,true,'2019-09-27 14:34:58.874','2019-09-27 14:34:58.874')
,('Tư thục Trung tiểu học Ngô Thời Nhiệm','COUNTRY_VN',41,469,NULL,true,'2019-09-27 14:34:58.876','2019-09-27 14:34:58.876')
,('Tư thục Trung Tiểu học PETRUS -KY','COUNTRY_VN',41,469,NULL,true,'2019-09-27 14:34:58.877','2019-09-27 14:34:58.877')
,('Tư thục Trung tiểu học Việt Anh','COUNTRY_VN',41,469,NULL,true,'2019-09-27 14:34:58.878','2019-09-27 14:34:58.878')
,('THPT Bến Cát','COUNTRY_VN',41,470,NULL,true,'2019-09-27 14:34:58.881','2019-09-27 14:34:58.881')
,('THPT Tây Nam','COUNTRY_VN',41,470,NULL,true,'2019-09-27 14:34:58.882','2019-09-27 14:34:58.882')
,('TT GDTX - KTHN H. Bến cát','COUNTRY_VN',41,470,NULL,true,'2019-09-27 14:34:58.883','2019-09-27 14:34:58.883')
,('CĐN Công nghệ và NL Nam Bộ','COUNTRY_VN',41,471,NULL,true,'2019-09-27 14:34:58.884','2019-09-27 14:34:58.884')
,('CĐN Đồng An','COUNTRY_VN',41,471,NULL,true,'2019-09-27 14:34:58.885','2019-09-27 14:34:58.885')
,('Phân hiệu CĐN Đường sắt phía Nam','COUNTRY_VN',41,471,NULL,true,'2019-09-27 14:34:58.886','2019-09-27 14:34:58.886')
,('TCN Dĩ An','COUNTRY_VN',41,471,NULL,true,'2019-09-27 14:34:58.886','2019-09-27 14:34:58.886')
,('TCN Khu Công nghiệp','COUNTRY_VN',41,471,NULL,true,'2019-09-27 14:34:58.887','2019-09-27 14:34:58.887')
,('THPT Bình An','COUNTRY_VN',41,471,NULL,true,'2019-09-27 14:34:58.887','2019-09-27 14:34:58.887')
,('THPT Dĩ An','COUNTRY_VN',41,471,NULL,true,'2019-09-27 14:34:58.887','2019-09-27 14:34:58.887')
,('TCN Tân Uyên','COUNTRY_VN',41,472,NULL,true,'2019-09-27 14:34:58.888','2019-09-27 14:34:58.888')
,('THPT Huỳnh văn Nghệ','COUNTRY_VN',41,472,NULL,true,'2019-09-27 14:34:58.889','2019-09-27 14:34:58.889')
,('THPT Tân Phước Khánh','COUNTRY_VN',41,472,NULL,true,'2019-09-27 14:34:58.889','2019-09-27 14:34:58.889')
,('THPT Thái Hoà','COUNTRY_VN',41,472,NULL,true,'2019-09-27 14:34:58.889','2019-09-27 14:34:58.889')
,('TT GDTX - KTHN H. Tân Uyên','COUNTRY_VN',41,472,NULL,true,'2019-09-27 14:34:58.890','2019-09-27 14:34:58.890')
,('CĐN Việt Nam - Singapore','COUNTRY_VN',41,473,NULL,true,'2019-09-27 14:34:58.893','2019-09-27 14:34:58.893')
,('TCN KT và NV công đoàn','COUNTRY_VN',41,473,NULL,true,'2019-09-27 14:34:58.895','2019-09-27 14:34:58.895')
,('THPT Nguyễn Trãi','COUNTRY_VN',41,473,NULL,true,'2019-09-27 14:34:58.895','2019-09-27 14:34:58.895')
,('THPT Trần Văn ơn','COUNTRY_VN',41,473,NULL,true,'2019-09-27 14:34:58.896','2019-09-27 14:34:58.896')
,('THPT Trịnh Hoài Đức','COUNTRY_VN',41,473,NULL,true,'2019-09-27 14:34:58.897','2019-09-27 14:34:58.897')
,('TT GDTX -KTHN TX. Thuận An','COUNTRY_VN',41,473,NULL,true,'2019-09-27 14:34:58.898','2019-09-27 14:34:58.898')
,('Tư thục Trung tiểu học Đức Trí','COUNTRY_VN',41,473,NULL,true,'2019-09-27 14:34:58.899','2019-09-27 14:34:58.899')
,('THPT Đồng Phú','COUNTRY_VN',42,474,NULL,true,'2019-09-27 14:34:58.901','2019-09-27 14:34:58.901')
,('THCS & THPT Lương Thế vinh','COUNTRY_VN',42,475,NULL,true,'2019-09-27 14:34:58.902','2019-09-27 14:34:58.902')
,('THPT Bù Đăng','COUNTRY_VN',42,475,NULL,true,'2019-09-27 14:34:58.903','2019-09-27 14:34:58.903')
,('THPT Lê Quý Đôn','COUNTRY_VN',42,475,NULL,true,'2019-09-27 14:34:58.903','2019-09-27 14:34:58.903')
,('THPT Thống Nhất','COUNTRY_VN',42,475,NULL,true,'2019-09-27 14:34:58.904','2019-09-27 14:34:58.904')
,('TT GDTX Bù Đăng','COUNTRY_VN',42,475,NULL,true,'2019-09-27 14:34:58.904','2019-09-27 14:34:58.904')
,('THCS & THPT Đăng Hà','COUNTRY_VN',42,476,NULL,true,'2019-09-27 14:34:58.905','2019-09-27 14:34:58.905')
,('THCS & THPT Tân Tiến','COUNTRY_VN',42,476,NULL,true,'2019-09-27 14:34:58.905','2019-09-27 14:34:58.905')
,('THPT Thanh Hòa','COUNTRY_VN',42,476,NULL,true,'2019-09-27 14:34:58.905','2019-09-27 14:34:58.905')
,('TT GDTX -DN Bù Đốp','COUNTRY_VN',42,476,NULL,true,'2019-09-27 14:34:58.906','2019-09-27 14:34:58.906')
,('THCS & THPT Đa Kia','COUNTRY_VN',42,477,NULL,true,'2019-09-27 14:34:58.906','2019-09-27 14:34:58.906')
,('THCS & THPT VÕ Thị sáu','COUNTRY_VN',42,477,NULL,true,'2019-09-27 14:34:58.907','2019-09-27 14:34:58.907')
,('THPT Đắc ơ','COUNTRY_VN',42,477,NULL,true,'2019-09-27 14:34:58.907','2019-09-27 14:34:58.907')
,('THPT Ngô Quyền','COUNTRY_VN',42,477,NULL,true,'2019-09-27 14:34:58.909','2019-09-27 14:34:58.909')
,('THPT Nguyễn Khuyến','COUNTRY_VN',42,477,NULL,true,'2019-09-27 14:34:58.911','2019-09-27 14:34:58.911')
,('THPT Phú Riềng','COUNTRY_VN',42,477,NULL,true,'2019-09-27 14:34:58.911','2019-09-27 14:34:58.911')
,('TC Nghề Tôn Đức Thắng','COUNTRY_VN',42,478,NULL,true,'2019-09-27 14:34:58.912','2019-09-27 14:34:58.912')
,('THCS & THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',42,478,NULL,true,'2019-09-27 14:34:58.913','2019-09-27 14:34:58.913')
,('THPT Chơn Thành','COUNTRY_VN',42,478,NULL,true,'2019-09-27 14:34:58.914','2019-09-27 14:34:58.914')
,('THPT Chu Văn An','COUNTRY_VN',42,478,NULL,true,'2019-09-27 14:34:58.915','2019-09-27 14:34:58.915')
,('TT GDTX Chơn Thành','COUNTRY_VN',42,478,NULL,true,'2019-09-27 14:34:58.916','2019-09-27 14:34:58.916')
,('THCS & THPT Đồng Tiến','COUNTRY_VN',42,479,NULL,true,'2019-09-27 14:34:58.917','2019-09-27 14:34:58.917')
,('TT GDTX Đồng Phú','COUNTRY_VN',42,479,NULL,true,'2019-09-27 14:34:58.918','2019-09-27 14:34:58.918')
,('THPT Nguyễn Hũu cảnh','COUNTRY_VN',42,480,NULL,true,'2019-09-27 14:34:58.918','2019-09-27 14:34:58.918')
,('THPT Trần Phú','COUNTRY_VN',42,480,NULL,true,'2019-09-27 14:34:58.919','2019-09-27 14:34:58.919')
,('THPT LỘC Hiệp','COUNTRY_VN',42,481,NULL,true,'2019-09-27 14:34:58.920','2019-09-27 14:34:58.920')
,('THPT LỘC Ninh','COUNTRY_VN',42,481,NULL,true,'2019-09-27 14:34:58.920','2019-09-27 14:34:58.920')
,('THPT Lộc Thái','COUNTRY_VN',42,481,NULL,true,'2019-09-27 14:34:58.920','2019-09-27 14:34:58.920')
,('TT GDTX -DN Lộc Ninh','COUNTRY_VN',42,481,NULL,true,'2019-09-27 14:34:58.921','2019-09-27 14:34:58.921')
,('THPT Bình Long','COUNTRY_VN',42,482,NULL,true,'2019-09-27 14:34:58.922','2019-09-27 14:34:58.922')
,('THPT Chuyên Bình Long','COUNTRY_VN',42,482,NULL,true,'2019-09-27 14:34:58.922','2019-09-27 14:34:58.922')
,('THPT Nguyễn Huệ','COUNTRY_VN',42,482,NULL,true,'2019-09-27 14:34:58.923','2019-09-27 14:34:58.923')
,('TT GDTX Bình Long','COUNTRY_VN',42,482,NULL,true,'2019-09-27 14:34:58.923','2019-09-27 14:34:58.923')
,('DTNT THPT TỈnh','COUNTRY_VN',42,483,NULL,true,'2019-09-27 14:34:58.926','2019-09-27 14:34:58.926')
,('THPT Chuyên Quang Trung','COUNTRY_VN',42,483,NULL,true,'2019-09-27 14:34:58.927','2019-09-27 14:34:58.927')
,('THPT Đồng Xoài','COUNTRY_VN',42,483,NULL,true,'2019-09-27 14:34:58.928','2019-09-27 14:34:58.928')
,('THPT Hùng Vương','COUNTRY_VN',42,483,NULL,true,'2019-09-27 14:34:58.928','2019-09-27 14:34:58.928')
,('THPT Nguyễn Du','COUNTRY_VN',42,483,NULL,true,'2019-09-27 14:34:58.929','2019-09-27 14:34:58.929')
,('TT GDTX Tinh','COUNTRY_VN',42,483,NULL,true,'2019-09-27 14:34:58.930','2019-09-27 14:34:58.930')
,('THPT Phước Bình','COUNTRY_VN',42,484,NULL,true,'2019-09-27 14:34:58.932','2019-09-27 14:34:58.932')
,('THPT Phước Long','COUNTRY_VN',42,484,NULL,true,'2019-09-27 14:34:58.933','2019-09-27 14:34:58.933')
,('TT GDTX Phước Long','COUNTRY_VN',42,484,NULL,true,'2019-09-27 14:34:58.934','2019-09-27 14:34:58.934')
,('THPT Bắc Bình','COUNTRY_VN',43,485,NULL,true,'2019-09-27 14:34:58.937','2019-09-27 14:34:58.937')
,('THPT Nguyễn Thị Minh Khai','COUNTRY_VN',43,485,NULL,true,'2019-09-27 14:34:58.937','2019-09-27 14:34:58.937')
,('TT GDTX-HN Bắc Bình','COUNTRY_VN',43,485,NULL,true,'2019-09-27 14:34:58.937','2019-09-27 14:34:58.937')
,('THPT Ngô Quyền','COUNTRY_VN',43,486,NULL,true,'2019-09-27 14:34:58.938','2019-09-27 14:34:58.938')
,('THPT Chu Văn An','COUNTRY_VN',43,487,NULL,true,'2019-09-27 14:34:58.939','2019-09-27 14:34:58.939')
,('THPT Đức Linh','COUNTRY_VN',43,487,NULL,true,'2019-09-27 14:34:58.939','2019-09-27 14:34:58.939')
,('THPT Hùng Vương','COUNTRY_VN',43,487,NULL,true,'2019-09-27 14:34:58.940','2019-09-27 14:34:58.940')
,('THPT Quang Trung','COUNTRY_VN',43,487,NULL,true,'2019-09-27 14:34:58.941','2019-09-27 14:34:58.941')
,('TT GDTX-HN Đức Linh','COUNTRY_VN',43,487,NULL,true,'2019-09-27 14:34:58.942','2019-09-27 14:34:58.942')
,('THPT Đức Tân','COUNTRY_VN',43,488,NULL,true,'2019-09-27 14:34:58.943','2019-09-27 14:34:58.943')
,('THPT Hàm Tân','COUNTRY_VN',43,488,NULL,true,'2019-09-27 14:34:58.944','2019-09-27 14:34:58.944')
,('THPT Huỳnh Thúc Kháng','COUNTRY_VN',43,488,NULL,true,'2019-09-27 14:34:58.944','2019-09-27 14:34:58.944')
,('THPT Dân tộc nội trú Tỉnh','COUNTRY_VN',43,489,NULL,true,'2019-09-27 14:34:58.946','2019-09-27 14:34:58.946')
,('THPT Hàm Thuận Bắc','COUNTRY_VN',43,489,NULL,true,'2019-09-27 14:34:58.947','2019-09-27 14:34:58.947')
,('THPT Nguyễn Văn Linh','COUNTRY_VN',43,489,NULL,true,'2019-09-27 14:34:58.948','2019-09-27 14:34:58.948')
,('THPT Hàm Thuận Nam','COUNTRY_VN',43,490,NULL,true,'2019-09-27 14:34:58.949','2019-09-27 14:34:58.949')
,('THPT Lương Thế Vinh','COUNTRY_VN',43,490,NULL,true,'2019-09-27 14:34:58.950','2019-09-27 14:34:58.950')
,('THPT Lý Thường Kiệt','COUNTRY_VN',43,491,NULL,true,'2019-09-27 14:34:58.951','2019-09-27 14:34:58.951')
,('THPT Nguyễn Huệ','COUNTRY_VN',43,491,NULL,true,'2019-09-27 14:34:58.952','2019-09-27 14:34:58.952')
,('THPT Nguyễn Trường Tộ','COUNTRY_VN',43,491,NULL,true,'2019-09-27 14:34:58.952','2019-09-27 14:34:58.952')
,('TT GDTX-HN La Gi','COUNTRY_VN',43,491,NULL,true,'2019-09-27 14:34:58.953','2019-09-27 14:34:58.953')
,('THPT Nguyễn Văn Trỗi','COUNTRY_VN',43,492,NULL,true,'2019-09-27 14:34:58.954','2019-09-27 14:34:58.954')
,('THPT Tánh Linh','COUNTRY_VN',43,492,NULL,true,'2019-09-27 14:34:58.954','2019-09-27 14:34:58.954')
,('TT GDTX-HN Tánh Linh','COUNTRY_VN',43,492,NULL,true,'2019-09-27 14:34:58.955','2019-09-27 14:34:58.955')
,('THPT Hòa Đa','COUNTRY_VN',43,493,NULL,true,'2019-09-27 14:34:58.956','2019-09-27 14:34:58.956')
,('THPT Lê Quý Đôn','COUNTRY_VN',43,493,NULL,true,'2019-09-27 14:34:58.956','2019-09-27 14:34:58.956')
,('THPT Nguyễn Khuyến','COUNTRY_VN',43,493,NULL,true,'2019-09-27 14:34:58.956','2019-09-27 14:34:58.956')
,('THPT Tuy Phong','COUNTRY_VN',43,493,NULL,true,'2019-09-27 14:34:58.957','2019-09-27 14:34:58.957')
,('CĐ Cộng đồng Bình Thuận','COUNTRY_VN',43,494,NULL,true,'2019-09-27 14:34:58.958','2019-09-27 14:34:58.958')
,('CĐ Nghề Bình Thuận','COUNTRY_VN',43,494,NULL,true,'2019-09-27 14:34:58.959','2019-09-27 14:34:58.959')
,('CĐ Ytế Bình Thuận','COUNTRY_VN',43,494,NULL,true,'2019-09-27 14:34:58.960','2019-09-27 14:34:58.960')
,('Đại học Phan Thiết','COUNTRY_VN',43,494,NULL,true,'2019-09-27 14:34:58.960','2019-09-27 14:34:58.960')
,('TC Du lịch Mũi Né','COUNTRY_VN',43,494,NULL,true,'2019-09-27 14:34:58.961','2019-09-27 14:34:58.961')
,('TC Nghề Kinh tế - Kỳ thuật CĐ Bình Thuận','COUNTRY_VN',43,494,NULL,true,'2019-09-27 14:34:58.961','2019-09-27 14:34:58.961')
,('TH Bổ túc Phan Bội Châu','COUNTRY_VN',43,494,NULL,true,'2019-09-27 14:34:58.962','2019-09-27 14:34:58.962')
,('TH, THCS và THPT Lê Quý Đôn','COUNTRY_VN',43,494,NULL,true,'2019-09-27 14:34:58.963','2019-09-27 14:34:58.963')
,('TH. THCS, THPT châu Á Thái Bình Dương','COUNTRY_VN',43,494,NULL,true,'2019-09-27 14:34:58.965','2019-09-27 14:34:58.965')
,('THCS và THPT Lê Lợi','COUNTRY_VN',43,494,NULL,true,'2019-09-27 14:34:58.965','2019-09-27 14:34:58.965')
,('THPT Bùi Thị Xuân','COUNTRY_VN',43,494,NULL,true,'2019-09-27 14:34:58.966','2019-09-27 14:34:58.966')
,('THPT Chuyên Trần Hưng Đạo','COUNTRY_VN',43,494,NULL,true,'2019-09-27 14:34:58.966','2019-09-27 14:34:58.966')
,('THPT Phan Bội Châu','COUNTRY_VN',43,494,NULL,true,'2019-09-27 14:34:58.967','2019-09-27 14:34:58.967')
,('THPT Phan Chu Trinh','COUNTRY_VN',43,494,NULL,true,'2019-09-27 14:34:58.967','2019-09-27 14:34:58.967')
,('THPT Phan Thiết','COUNTRY_VN',43,494,NULL,true,'2019-09-27 14:34:58.968','2019-09-27 14:34:58.968')
,('THPT Cái Nước','COUNTRY_VN',44,495,NULL,true,'2019-09-27 14:34:58.969','2019-09-27 14:34:58.969')
,('THPT Nguyễn Mai','COUNTRY_VN',44,495,NULL,true,'2019-09-27 14:34:58.970','2019-09-27 14:34:58.970')
,('THPT Phú Hưng','COUNTRY_VN',44,495,NULL,true,'2019-09-27 14:34:58.970','2019-09-27 14:34:58.970')
,('TT GDTX Cái Nước','COUNTRY_VN',44,495,NULL,true,'2019-09-27 14:34:58.971','2019-09-27 14:34:58.971')
,('THPT Đầm Dơi','COUNTRY_VN',44,496,NULL,true,'2019-09-27 14:34:58.971','2019-09-27 14:34:58.971')
,('THPT Tân Đức','COUNTRY_VN',44,496,NULL,true,'2019-09-27 14:34:58.972','2019-09-27 14:34:58.972')
,('THPT Thái Thanh Hoà','COUNTRY_VN',44,496,NULL,true,'2019-09-27 14:34:58.973','2019-09-27 14:34:58.973')
,('TT GDTX Đầm Dơi','COUNTRY_VN',44,496,NULL,true,'2019-09-27 14:34:58.973','2019-09-27 14:34:58.973')
,('THPT Phan Ngọc Hiển','COUNTRY_VN',44,497,NULL,true,'2019-09-27 14:34:58.975','2019-09-27 14:34:58.975')
,('TT GDTX Năm căn','COUNTRY_VN',44,497,NULL,true,'2019-09-27 14:34:58.976','2019-09-27 14:34:58.976')
,('THPT Viên An','COUNTRY_VN',44,498,NULL,true,'2019-09-27 14:34:58.977','2019-09-27 14:34:58.977')
,('TT GDTX Ngọc Hiến','COUNTRY_VN',44,498,NULL,true,'2019-09-27 14:34:58.978','2019-09-27 14:34:58.978')
,('THPT Nguyễn Thị Minh Khai','COUNTRY_VN',44,499,NULL,true,'2019-09-27 14:34:58.980','2019-09-27 14:34:58.980')
,('THPT Phú Tân','COUNTRY_VN',44,499,NULL,true,'2019-09-27 14:34:58.981','2019-09-27 14:34:58.981')
,('TT GDTX Phú Tân','COUNTRY_VN',44,499,NULL,true,'2019-09-27 14:34:58.982','2019-09-27 14:34:58.982')
,('THPT Lê Công Nhân','COUNTRY_VN',44,500,NULL,true,'2019-09-27 14:34:58.983','2019-09-27 14:34:58.983')
,('THPT Nguyễn Văn Nguyễn','COUNTRY_VN',44,500,NULL,true,'2019-09-27 14:34:58.983','2019-09-27 14:34:58.983')
,('THPT Thới Bình','COUNTRY_VN',44,500,NULL,true,'2019-09-27 14:34:58.984','2019-09-27 14:34:58.984')
,('TT GDTX Thới Bình','COUNTRY_VN',44,500,NULL,true,'2019-09-27 14:34:58.984','2019-09-27 14:34:58.984')
,('THPT Huỳnh Phi Hùng','COUNTRY_VN',44,501,NULL,true,'2019-09-27 14:34:58.985','2019-09-27 14:34:58.985')
,('THPT Khánh Hưng','COUNTRY_VN',44,501,NULL,true,'2019-09-27 14:34:58.985','2019-09-27 14:34:58.985')
,('THPT Sông Đốc','COUNTRY_VN',44,501,NULL,true,'2019-09-27 14:34:58.986','2019-09-27 14:34:58.986')
,('THPT Trần Văn Thời','COUNTRY_VN',44,501,NULL,true,'2019-09-27 14:34:58.986','2019-09-27 14:34:58.986')
,('TT GDTX Trần văn Thời','COUNTRY_VN',44,501,NULL,true,'2019-09-27 14:34:58.986','2019-09-27 14:34:58.986')
,('THPT Khánh An','COUNTRY_VN',44,502,NULL,true,'2019-09-27 14:34:58.987','2019-09-27 14:34:58.987')
,('THPT Khánh Lâm','COUNTRY_VN',44,502,NULL,true,'2019-09-27 14:34:58.987','2019-09-27 14:34:58.987')
,('THPT u Minh','COUNTRY_VN',44,502,NULL,true,'2019-09-27 14:34:58.988','2019-09-27 14:34:58.988')
,('TT GDTX u Minh','COUNTRY_VN',44,502,NULL,true,'2019-09-27 14:34:58.988','2019-09-27 14:34:58.988')
,('Phổ thông Hermann Gmeiner','COUNTRY_VN',44,503,NULL,true,'2019-09-27 14:34:58.989','2019-09-27 14:34:58.989')
,('PT Dân tộc nội trú','COUNTRY_VN',44,503,NULL,true,'2019-09-27 14:34:58.989','2019-09-27 14:34:58.989')
,('TC Nghề Cà Mau','COUNTRY_VN',44,503,NULL,true,'2019-09-27 14:34:58.989','2019-09-27 14:34:58.989')
,('THPT Cà Mau','COUNTRY_VN',44,503,NULL,true,'2019-09-27 14:34:58.990','2019-09-27 14:34:58.990')
,('THPT Chuyên Phan Ngọc Hiển','COUNTRY_VN',44,503,NULL,true,'2019-09-27 14:34:58.990','2019-09-27 14:34:58.990')
,('THPT Hồ Thị Kỷ','COUNTRY_VN',44,503,NULL,true,'2019-09-27 14:34:58.992','2019-09-27 14:34:58.992')
,('THPT Ngọc Hiển','COUNTRY_VN',44,503,NULL,true,'2019-09-27 14:34:58.993','2019-09-27 14:34:58.993')
,('THPT Nguyễn Việt Khải','COUNTRY_VN',44,503,NULL,true,'2019-09-27 14:34:58.994','2019-09-27 14:34:58.994')
,('THPT Tắc Vân','COUNTRY_VN',44,503,NULL,true,'2019-09-27 14:34:58.995','2019-09-27 14:34:58.995')
,('THPT Thanh Bình Cà Mau','COUNTRY_VN',44,503,NULL,true,'2019-09-27 14:34:58.996','2019-09-27 14:34:58.996')
,('THPT Võ Thị Hồng','COUNTRY_VN',44,503,NULL,true,'2019-09-27 14:34:58.998','2019-09-27 14:34:58.998')
,('TT GDTX TP.Cà Mau','COUNTRY_VN',44,503,NULL,true,'2019-09-27 14:34:58.999','2019-09-27 14:34:58.999')
,('THPT Bản Ngà','COUNTRY_VN',45,504,NULL,true,'2019-09-27 14:34:59.000','2019-09-27 14:34:59.000')
,('THPT Bảo Lạc','COUNTRY_VN',45,504,NULL,true,'2019-09-27 14:34:59.002','2019-09-27 14:34:59.002')
,('TT GDTX Bảo Lạc','COUNTRY_VN',45,504,NULL,true,'2019-09-27 14:34:59.003','2019-09-27 14:34:59.003')
,('THPT Bảo Lâm','COUNTRY_VN',45,505,NULL,true,'2019-09-27 14:34:59.004','2019-09-27 14:34:59.004')
,('THPT Lý Bôn','COUNTRY_VN',45,505,NULL,true,'2019-09-27 14:34:59.004','2019-09-27 14:34:59.004')
,('TT GDTX Bảo Lâm','COUNTRY_VN',45,505,NULL,true,'2019-09-27 14:34:59.005','2019-09-27 14:34:59.005')
,('THPT Bằng Ca','COUNTRY_VN',45,506,NULL,true,'2019-09-27 14:34:59.006','2019-09-27 14:34:59.006')
,('THPT Hạ Lang','COUNTRY_VN',45,506,NULL,true,'2019-09-27 14:34:59.006','2019-09-27 14:34:59.006')
,('TT GDTX Hạ Lang','COUNTRY_VN',45,506,NULL,true,'2019-09-27 14:34:59.006','2019-09-27 14:34:59.006')
,('THPT Hà Quảng','COUNTRY_VN',45,507,NULL,true,'2019-09-27 14:34:59.007','2019-09-27 14:34:59.007')
,('THPT Lục Khu','COUNTRY_VN',45,507,NULL,true,'2019-09-27 14:34:59.009','2019-09-27 14:34:59.009')
,('THPT Nà Giàng','COUNTRY_VN',45,507,NULL,true,'2019-09-27 14:34:59.011','2019-09-27 14:34:59.011')
,('TT GDTX Hà Quảng','COUNTRY_VN',45,507,NULL,true,'2019-09-27 14:34:59.011','2019-09-27 14:34:59.011')
,('THPT Hoà An','COUNTRY_VN',45,508,NULL,true,'2019-09-27 14:34:59.012','2019-09-27 14:34:59.012')
,('TT GDTX Hoà An','COUNTRY_VN',45,508,NULL,true,'2019-09-27 14:34:59.013','2019-09-27 14:34:59.013')
,('THPT Nà Bao','COUNTRY_VN',45,509,NULL,true,'2019-09-27 14:34:59.015','2019-09-27 14:34:59.015')
,('THPT Nguyên Bình','COUNTRY_VN',45,509,NULL,true,'2019-09-27 14:34:59.015','2019-09-27 14:34:59.015')
,('THPT Tĩnh Túc','COUNTRY_VN',45,509,NULL,true,'2019-09-27 14:34:59.016','2019-09-27 14:34:59.016')
,('TT GDTX Nguyên Bình','COUNTRY_VN',45,509,NULL,true,'2019-09-27 14:34:59.017','2019-09-27 14:34:59.017')
,('THPT Cách Linh','COUNTRY_VN',45,510,NULL,true,'2019-09-27 14:34:59.018','2019-09-27 14:34:59.018')
,('THPT Phục Hoà','COUNTRY_VN',45,510,NULL,true,'2019-09-27 14:34:59.018','2019-09-27 14:34:59.018')
,('TT GDTX Phục Hoà','COUNTRY_VN',45,510,NULL,true,'2019-09-27 14:34:59.019','2019-09-27 14:34:59.019')
,('THPT Đống Đa','COUNTRY_VN',45,511,NULL,true,'2019-09-27 14:34:59.020','2019-09-27 14:34:59.020')
,('THPT Quảng Uyên','COUNTRY_VN',45,511,NULL,true,'2019-09-27 14:34:59.020','2019-09-27 14:34:59.020')
,('TT GDTX Quảng Uyên','COUNTRY_VN',45,511,NULL,true,'2019-09-27 14:34:59.021','2019-09-27 14:34:59.021')
,('THPT Canh Tân','COUNTRY_VN',45,512,NULL,true,'2019-09-27 14:34:59.022','2019-09-27 14:34:59.022')
,('THPT Thạch An','COUNTRY_VN',45,512,NULL,true,'2019-09-27 14:34:59.023','2019-09-27 14:34:59.023')
,('TT GDTX Thạch An','COUNTRY_VN',45,512,NULL,true,'2019-09-27 14:34:59.023','2019-09-27 14:34:59.023')
,('THPT Thông Nông','COUNTRY_VN',45,513,NULL,true,'2019-09-27 14:34:59.026','2019-09-27 14:34:59.026')
,('TT GDTX Thông Nông','COUNTRY_VN',45,513,NULL,true,'2019-09-27 14:34:59.026','2019-09-27 14:34:59.026')
,('THPT Quang Trung','COUNTRY_VN',45,514,NULL,true,'2019-09-27 14:34:59.027','2019-09-27 14:34:59.027')
,('THPT Trà Lĩnh','COUNTRY_VN',45,514,NULL,true,'2019-09-27 14:34:59.028','2019-09-27 14:34:59.028')
,('TT GDTX Trà Lĩnh','COUNTRY_VN',45,514,NULL,true,'2019-09-27 14:34:59.029','2019-09-27 14:34:59.029')
,('THPT Pò Tấu','COUNTRY_VN',45,515,NULL,true,'2019-09-27 14:34:59.030','2019-09-27 14:34:59.030')
,('THPT Thông Huề','COUNTRY_VN',45,515,NULL,true,'2019-09-27 14:34:59.031','2019-09-27 14:34:59.031')
,('THPT Trùng Khánh','COUNTRY_VN',45,515,NULL,true,'2019-09-27 14:34:59.031','2019-09-27 14:34:59.031')
,('TT GDTX Trùng Khánh','COUNTRY_VN',45,515,NULL,true,'2019-09-27 14:34:59.032','2019-09-27 14:34:59.032')
,('TC nghề Cao Bằng','COUNTRY_VN',45,516,NULL,true,'2019-09-27 14:34:59.033','2019-09-27 14:34:59.033')
,('THPT Bế Văn Đàn','COUNTRY_VN',45,516,NULL,true,'2019-09-27 14:34:59.033','2019-09-27 14:34:59.033')
,('THPT Cao Bình','COUNTRY_VN',45,516,NULL,true,'2019-09-27 14:34:59.034','2019-09-27 14:34:59.034')
,('THPT Chuyên Cao Bằng','COUNTRY_VN',45,516,NULL,true,'2019-09-27 14:34:59.034','2019-09-27 14:34:59.034')
,('THPT DTNT Cao Bằng','COUNTRY_VN',45,516,NULL,true,'2019-09-27 14:34:59.034','2019-09-27 14:34:59.034')
,('THPT Thành phố Cao Bằng','COUNTRY_VN',45,516,NULL,true,'2019-09-27 14:34:59.035','2019-09-27 14:34:59.035')
,('TT GDTX Thành phố Cao Bằng','COUNTRY_VN',45,516,NULL,true,'2019-09-27 14:34:59.035','2019-09-27 14:34:59.035')
,('TT KTTH-HN tỉnh Cao Bằng','COUNTRY_VN',45,516,NULL,true,'2019-09-27 14:34:59.036','2019-09-27 14:34:59.036')
,('TT GDTX Tỉnh','COUNTRY_VN',45,516,NULL,true,'2019-09-27 14:34:59.036','2019-09-27 14:34:59.036')
,('THPT Buôn Đốn','COUNTRY_VN',46,517,NULL,true,'2019-09-27 14:34:59.037','2019-09-27 14:34:59.037')
,('THPT Trần Đại Nghĩa','COUNTRY_VN',46,517,NULL,true,'2019-09-27 14:34:59.037','2019-09-27 14:34:59.037')
,('TT GDTX Buôn Đốn','COUNTRY_VN',46,517,NULL,true,'2019-09-27 14:34:59.038','2019-09-27 14:34:59.038')
,('THPT Việt ĐỨC','COUNTRY_VN',46,518,NULL,true,'2019-09-27 14:34:59.039','2019-09-27 14:34:59.039')
,('THPT Y Jut','COUNTRY_VN',46,518,NULL,true,'2019-09-27 14:34:59.039','2019-09-27 14:34:59.039')
,('TT GDTX CưKuin','COUNTRY_VN',46,518,NULL,true,'2019-09-27 14:34:59.041','2019-09-27 14:34:59.041')
,('THPT Cư M''Gar','COUNTRY_VN',46,519,NULL,true,'2019-09-27 14:34:59.044','2019-09-27 14:34:59.044')
,('THPT Lê Hữu Trác','COUNTRY_VN',46,519,NULL,true,'2019-09-27 14:34:59.045','2019-09-27 14:34:59.045')
,('THPT Nguyễn Trãi','COUNTRY_VN',46,519,NULL,true,'2019-09-27 14:34:59.046','2019-09-27 14:34:59.046')
,('THPT Trần Quang Khải','COUNTRY_VN',46,519,NULL,true,'2019-09-27 14:34:59.046','2019-09-27 14:34:59.046')
,('TT GDTX Cư M’Gar','COUNTRY_VN',46,519,NULL,true,'2019-09-27 14:34:59.047','2019-09-27 14:34:59.047')
,('THPT Ea H’leo','COUNTRY_VN',46,520,NULL,true,'2019-09-27 14:34:59.049','2019-09-27 14:34:59.049')
,('THPT Phan Chu Trinh','COUNTRY_VN',46,520,NULL,true,'2019-09-27 14:34:59.050','2019-09-27 14:34:59.050')
,('THPT Trường Chinh','COUNTRY_VN',46,520,NULL,true,'2019-09-27 14:34:59.051','2019-09-27 14:34:59.051')
,('TT GDTX Ea H’Leo','COUNTRY_VN',46,520,NULL,true,'2019-09-27 14:34:59.052','2019-09-27 14:34:59.052')
,('THPT Ngô Gia Tự','COUNTRY_VN',46,521,NULL,true,'2019-09-27 14:34:59.053','2019-09-27 14:34:59.053')
,('THPT Nguyễn Thái Bình','COUNTRY_VN',46,521,NULL,true,'2019-09-27 14:34:59.053','2019-09-27 14:34:59.053')
,('THPT Trần Nhân Tông','COUNTRY_VN',46,521,NULL,true,'2019-09-27 14:34:59.054','2019-09-27 14:34:59.054')
,('THPT Trần Quốc Toản','COUNTRY_VN',46,521,NULL,true,'2019-09-27 14:34:59.055','2019-09-27 14:34:59.055')
,('TT GDTX Ea Kar','COUNTRY_VN',46,521,NULL,true,'2019-09-27 14:34:59.055','2019-09-27 14:34:59.055')
,('THPT Ea Rốk','COUNTRY_VN',46,522,NULL,true,'2019-09-27 14:34:59.056','2019-09-27 14:34:59.056')
,('THPT Ea Súp','COUNTRY_VN',46,522,NULL,true,'2019-09-27 14:34:59.056','2019-09-27 14:34:59.056')
,('TT GDTX Ea súp','COUNTRY_VN',46,522,NULL,true,'2019-09-27 14:34:59.057','2019-09-27 14:34:59.057')
,('THPT Hùng Vương','COUNTRY_VN',46,523,NULL,true,'2019-09-27 14:34:59.059','2019-09-27 14:34:59.059')
,('THPT Krông Ana','COUNTRY_VN',46,523,NULL,true,'2019-09-27 14:34:59.059','2019-09-27 14:34:59.059')
,('THPT Phạm Văn Đồng','COUNTRY_VN',46,523,NULL,true,'2019-09-27 14:34:59.060','2019-09-27 14:34:59.060')
,('TT GDTX Krông Ana','COUNTRY_VN',46,523,NULL,true,'2019-09-27 14:34:59.060','2019-09-27 14:34:59.060')
,('THPT Krông Bông','COUNTRY_VN',46,524,NULL,true,'2019-09-27 14:34:59.061','2019-09-27 14:34:59.061')
,('THPT Trần Hung Đạo','COUNTRY_VN',46,524,NULL,true,'2019-09-27 14:34:59.062','2019-09-27 14:34:59.062')
,('TT GDTX Krông Bông','COUNTRY_VN',46,524,NULL,true,'2019-09-27 14:34:59.062','2019-09-27 14:34:59.062')
,('THPT Nguyễn Văn Cừ','COUNTRY_VN',46,525,NULL,true,'2019-09-27 14:34:59.063','2019-09-27 14:34:59.063')
,('THPT Phan Đăng Lưu','COUNTRY_VN',46,525,NULL,true,'2019-09-27 14:34:59.064','2019-09-27 14:34:59.064')
,('THPT Lý Tự Trọng','COUNTRY_VN',46,526,NULL,true,'2019-09-27 14:34:59.065','2019-09-27 14:34:59.065')
,('THPT Nguyễn Huệ','COUNTRY_VN',46,526,NULL,true,'2019-09-27 14:34:59.066','2019-09-27 14:34:59.066')
,('THPT Phan Bội Châu','COUNTRY_VN',46,526,NULL,true,'2019-09-27 14:34:59.066','2019-09-27 14:34:59.066')
,('THPT Tôn Đức Thắng','COUNTRY_VN',46,526,NULL,true,'2019-09-27 14:34:59.067','2019-09-27 14:34:59.067')
,('TT GDTX Krông Năng','COUNTRY_VN',46,526,NULL,true,'2019-09-27 14:34:59.067','2019-09-27 14:34:59.067')
,('THPT Lê Hồng Phong','COUNTRY_VN',46,527,NULL,true,'2019-09-27 14:34:59.068','2019-09-27 14:34:59.068')
,('THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',46,527,NULL,true,'2019-09-27 14:34:59.069','2019-09-27 14:34:59.069')
,('THPT Nguyễn Công Trứ','COUNTRY_VN',46,527,NULL,true,'2019-09-27 14:34:59.069','2019-09-27 14:34:59.069')
,('THPT Nguyễn Thị Minh Khai','COUNTRY_VN',46,527,NULL,true,'2019-09-27 14:34:59.069','2019-09-27 14:34:59.069')
,('THPT Phan Đình Phùng','COUNTRY_VN',46,527,NULL,true,'2019-09-27 14:34:59.070','2019-09-27 14:34:59.070')
,('THPT Quang Trung','COUNTRY_VN',46,527,NULL,true,'2019-09-27 14:34:59.070','2019-09-27 14:34:59.070')
,('TT GDTX Krông Pắk','COUNTRY_VN',46,527,NULL,true,'2019-09-27 14:34:59.071','2019-09-27 14:34:59.071')
,('THPT Lăk','COUNTRY_VN',46,528,NULL,true,'2019-09-27 14:34:59.072','2019-09-27 14:34:59.072')
,('TT GDTX Lăk','COUNTRY_VN',46,528,NULL,true,'2019-09-27 14:34:59.072','2019-09-27 14:34:59.072')
,('THPT Nguyễn Trường Tộ','COUNTRY_VN',46,529,NULL,true,'2019-09-27 14:34:59.073','2019-09-27 14:34:59.073')
,('THPT Nguyễn Tất Thành','COUNTRY_VN',46,529,NULL,true,'2019-09-27 14:34:59.073','2019-09-27 14:34:59.073')
,('TT GDTX M’Drắk','COUNTRY_VN',46,529,NULL,true,'2019-09-27 14:34:59.074','2019-09-27 14:34:59.074')
,('CĐ Nghề TN Dân Tộc, Đêk Lăk','COUNTRY_VN',46,530,NULL,true,'2019-09-27 14:34:59.076','2019-09-27 14:34:59.076')
,('năng khiếu Thể dục Thể thao','COUNTRY_VN',46,530,NULL,true,'2019-09-27 14:34:59.077','2019-09-27 14:34:59.077')
,('TC Kinh tế Kỹ thuật Đắk lắk','COUNTRY_VN',46,530,NULL,true,'2019-09-27 14:34:59.077','2019-09-27 14:34:59.077')
,('TC nghề Đăk Lăk','COUNTRY_VN',46,530,NULL,true,'2019-09-27 14:34:59.078','2019-09-27 14:34:59.078')
,('THPT Buôn Ma Thuột','COUNTRY_VN',46,530,NULL,true,'2019-09-27 14:34:59.078','2019-09-27 14:34:59.078')
,('THPT Cao Bá Quát','COUNTRY_VN',46,530,NULL,true,'2019-09-27 14:34:59.079','2019-09-27 14:34:59.079')
,('THPT Chu Văn An','COUNTRY_VN',46,530,NULL,true,'2019-09-27 14:34:59.080','2019-09-27 14:34:59.080')
,('THPT Buôn Hồ','COUNTRY_VN',46,531,NULL,true,'2019-09-27 14:34:59.082','2019-09-27 14:34:59.082')
,('THPT Hai Bà Trung','COUNTRY_VN',46,531,NULL,true,'2019-09-27 14:34:59.083','2019-09-27 14:34:59.083')
,('THPT Huỳnh Thúc Kháng','COUNTRY_VN',46,531,NULL,true,'2019-09-27 14:34:59.083','2019-09-27 14:34:59.083')
,('TT GDTX Buôn Hồ','COUNTRY_VN',46,531,NULL,true,'2019-09-27 14:34:59.084','2019-09-27 14:34:59.084')
,('Phổ thông DTNT Cư Jút','COUNTRY_VN',47,532,NULL,true,'2019-09-27 14:34:59.085','2019-09-27 14:34:59.085')
,('THPT Đào Duy Từ','COUNTRY_VN',47,532,NULL,true,'2019-09-27 14:34:59.086','2019-09-27 14:34:59.086')
,('THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',47,532,NULL,true,'2019-09-27 14:34:59.086','2019-09-27 14:34:59.086')
,('THPT Phan Bội Châu','COUNTRY_VN',47,532,NULL,true,'2019-09-27 14:34:59.086','2019-09-27 14:34:59.086')
,('THPT Phan Chu Trinh','COUNTRY_VN',47,532,NULL,true,'2019-09-27 14:34:59.087','2019-09-27 14:34:59.087')
,('TT GDTX Cư Jút','COUNTRY_VN',47,532,NULL,true,'2019-09-27 14:34:59.087','2019-09-27 14:34:59.087')
,('Phổ thông DTNT Đăk Giong','COUNTRY_VN',47,533,NULL,true,'2019-09-27 14:34:59.088','2019-09-27 14:34:59.088')
,('THPT Đăk Giong','COUNTRY_VN',47,533,NULL,true,'2019-09-27 14:34:59.088','2019-09-27 14:34:59.088')
,('Trung học cs và THPT Lê Duẩn','COUNTRY_VN',47,533,NULL,true,'2019-09-27 14:34:59.088','2019-09-27 14:34:59.088')
,('Phổ thông DTNT Đăk Mil','COUNTRY_VN',47,534,NULL,true,'2019-09-27 14:34:59.089','2019-09-27 14:34:59.089')
,('THPT ĐăkMil','COUNTRY_VN',47,534,NULL,true,'2019-09-27 14:34:59.089','2019-09-27 14:34:59.089')
,('THPT Nguyễn Du','COUNTRY_VN',47,534,NULL,true,'2019-09-27 14:34:59.090','2019-09-27 14:34:59.090')
,('THPT Quang Trung','COUNTRY_VN',47,534,NULL,true,'2019-09-27 14:34:59.090','2019-09-27 14:34:59.090')
,('THPT Trần Hung Đạo','COUNTRY_VN',47,534,NULL,true,'2019-09-27 14:34:59.092','2019-09-27 14:34:59.092')
,('TT GDTX Đăk Mil','COUNTRY_VN',47,534,NULL,true,'2019-09-27 14:34:59.093','2019-09-27 14:34:59.093')
,('Phổ thông DTNT Đăk RLấp','COUNTRY_VN',47,535,NULL,true,'2019-09-27 14:34:59.094','2019-09-27 14:34:59.094')
,('THPT Nguyễn Đình Chiểu','COUNTRY_VN',47,535,NULL,true,'2019-09-27 14:34:59.095','2019-09-27 14:34:59.095')
,('THPT Nguyễn Tất Thành','COUNTRY_VN',47,535,NULL,true,'2019-09-27 14:34:59.096','2019-09-27 14:34:59.096')
,('THPT Phạm Văn Đồng','COUNTRY_VN',47,535,NULL,true,'2019-09-27 14:34:59.097','2019-09-27 14:34:59.097')
,('THPT Trường Chinh','COUNTRY_VN',47,535,NULL,true,'2019-09-27 14:34:59.098','2019-09-27 14:34:59.098')
,('TT GDTX Đăk RLấp','COUNTRY_VN',47,535,NULL,true,'2019-09-27 14:34:59.098','2019-09-27 14:34:59.098')
,('Phổ thông DTNT Đăk Song','COUNTRY_VN',47,536,NULL,true,'2019-09-27 14:34:59.099','2019-09-27 14:34:59.099')
,('THPT Đăk Song','COUNTRY_VN',47,536,NULL,true,'2019-09-27 14:34:59.100','2019-09-27 14:34:59.100')
,('THPT Phan Đình Phùng','COUNTRY_VN',47,536,NULL,true,'2019-09-27 14:34:59.100','2019-09-27 14:34:59.100')
,('TT GDTX Đăk Song','COUNTRY_VN',47,536,NULL,true,'2019-09-27 14:34:59.101','2019-09-27 14:34:59.101')
,('Phổ thông DTNT Krông Nô','COUNTRY_VN',47,537,NULL,true,'2019-09-27 14:34:59.102','2019-09-27 14:34:59.102')
,('THPT Hùng Vuưng','COUNTRY_VN',47,537,NULL,true,'2019-09-27 14:34:59.102','2019-09-27 14:34:59.102')
,('THPT Krông Nô','COUNTRY_VN',47,537,NULL,true,'2019-09-27 14:34:59.103','2019-09-27 14:34:59.103')
,('THPT Trần Phú','COUNTRY_VN',47,537,NULL,true,'2019-09-27 14:34:59.103','2019-09-27 14:34:59.103')
,('TT GDTX Krông Nô','COUNTRY_VN',47,537,NULL,true,'2019-09-27 14:34:59.104','2019-09-27 14:34:59.104')
,('THPT Lê Quý Đôn','COUNTRY_VN',47,538,NULL,true,'2019-09-27 14:34:59.104','2019-09-27 14:34:59.104')
,('TC Nghề Đăk Nông','COUNTRY_VN',47,539,NULL,true,'2019-09-27 14:34:59.105','2019-09-27 14:34:59.105')
,('THPT Chu Văn An','COUNTRY_VN',47,539,NULL,true,'2019-09-27 14:34:59.105','2019-09-27 14:34:59.105')
,('THPT Chuyên Nguyễn Chí Thanh','COUNTRY_VN',47,539,NULL,true,'2019-09-27 14:34:59.106','2019-09-27 14:34:59.106')
,('THPT DTNT NTrang Lơng tỉnh Đắk Nông','COUNTRY_VN',47,539,NULL,true,'2019-09-27 14:34:59.106','2019-09-27 14:34:59.106')
,('THPT Gia Nghía','COUNTRY_VN',47,539,NULL,true,'2019-09-27 14:34:59.106','2019-09-27 14:34:59.106')
,('TT GDTX tỉnh','COUNTRY_VN',47,539,NULL,true,'2019-09-27 14:34:59.107','2019-09-27 14:34:59.107')
,('Cao đẳng Nghề Điện Biên','COUNTRY_VN',48,540,NULL,true,'2019-09-27 14:34:59.111','2019-09-27 14:34:59.111')
,('THPT huyện Điện Biên','COUNTRY_VN',48,540,NULL,true,'2019-09-27 14:34:59.112','2019-09-27 14:34:59.112')
,('THPT Mường Nhà','COUNTRY_VN',48,540,NULL,true,'2019-09-27 14:34:59.112','2019-09-27 14:34:59.112')
,('THPT Nà Tấu','COUNTRY_VN',48,540,NULL,true,'2019-09-27 14:34:59.112','2019-09-27 14:34:59.112')
,('THPT Thanh Chăn','COUNTRY_VN',48,540,NULL,true,'2019-09-27 14:34:59.113','2019-09-27 14:34:59.113')
,('THPT Thanh Nưa','COUNTRY_VN',48,540,NULL,true,'2019-09-27 14:34:59.113','2019-09-27 14:34:59.113')
,('Trung tâm GDTX huyện Điện Biên','COUNTRY_VN',48,540,NULL,true,'2019-09-27 14:34:59.114','2019-09-27 14:34:59.114')
,('PT DTN THPT Điện Biên Đông','COUNTRY_VN',48,541,NULL,true,'2019-09-27 14:34:59.115','2019-09-27 14:34:59.115')
,('THPT Mường Luân','COUNTRY_VN',48,541,NULL,true,'2019-09-27 14:34:59.115','2019-09-27 14:34:59.115')
,('THPT Trần Can','COUNTRY_VN',48,541,NULL,true,'2019-09-27 14:34:59.116','2019-09-27 14:34:59.116')
,('Trung tâm GDTX huyện Điện Biên Đông','COUNTRY_VN',48,541,NULL,true,'2019-09-27 14:34:59.117','2019-09-27 14:34:59.117')
,('PT DTN THPT Mường Ảng','COUNTRY_VN',48,542,NULL,true,'2019-09-27 14:34:59.118','2019-09-27 14:34:59.118')
,('THPT Búng Lao','COUNTRY_VN',48,542,NULL,true,'2019-09-27 14:34:59.119','2019-09-27 14:34:59.119')
,('THPT Mường ảng','COUNTRY_VN',48,542,NULL,true,'2019-09-27 14:34:59.119','2019-09-27 14:34:59.119')
,('Trung tâm GDTX huyện Mường Ảng','COUNTRY_VN',48,542,NULL,true,'2019-09-27 14:34:59.120','2019-09-27 14:34:59.120')
,('PT DTN THPT Mường chà','COUNTRY_VN',48,543,NULL,true,'2019-09-27 14:34:59.120','2019-09-27 14:34:59.120')
,('THPT Mường Chà','COUNTRY_VN',48,543,NULL,true,'2019-09-27 14:34:59.121','2019-09-27 14:34:59.121')
,('Trung tâm GDTX huyện Mường Chà','COUNTRY_VN',48,543,NULL,true,'2019-09-27 14:34:59.121','2019-09-27 14:34:59.121')
,('THPT DTNT H. Mường Nhé','COUNTRY_VN',48,544,NULL,true,'2019-09-27 14:34:59.122','2019-09-27 14:34:59.122')
,('THPT Mường Nhé','COUNTRY_VN',48,544,NULL,true,'2019-09-27 14:34:59.122','2019-09-27 14:34:59.122')
,('Trung tâm GDTX huyện Mường Nhé','COUNTRY_VN',48,544,NULL,true,'2019-09-27 14:34:59.123','2019-09-27 14:34:59.123')
,('THPT Chà Cang','COUNTRY_VN',48,545,NULL,true,'2019-09-27 14:34:59.124','2019-09-27 14:34:59.124')
,('PT DTN THPT huyện Tủa Chùa','COUNTRY_VN',48,546,NULL,true,'2019-09-27 14:34:59.127','2019-09-27 14:34:59.127')
,('THPT Tả Sìn Thàng','COUNTRY_VN',48,546,NULL,true,'2019-09-27 14:34:59.127','2019-09-27 14:34:59.127')
,('THPT Tủa Chùa','COUNTRY_VN',48,546,NULL,true,'2019-09-27 14:34:59.128','2019-09-27 14:34:59.128')
,('Trung tâm GDTX huyện Tủa Chùa','COUNTRY_VN',48,546,NULL,true,'2019-09-27 14:34:59.130','2019-09-27 14:34:59.130')
,('PT DTN THPT huyện Tuần Giáo','COUNTRY_VN',48,547,NULL,true,'2019-09-27 14:34:59.131','2019-09-27 14:34:59.131')
,('THPT Mùn Chung','COUNTRY_VN',48,547,NULL,true,'2019-09-27 14:34:59.132','2019-09-27 14:34:59.132')
,('THPT Tuần Giáo','COUNTRY_VN',48,547,NULL,true,'2019-09-27 14:34:59.133','2019-09-27 14:34:59.133')
,('Trung tâm GDTX huyện Tuần Giáo','COUNTRY_VN',48,547,NULL,true,'2019-09-27 14:34:59.133','2019-09-27 14:34:59.133')
,('Phổ thông Dân tộc Nội Trú Tỉnh','COUNTRY_VN',48,548,NULL,true,'2019-09-27 14:34:59.134','2019-09-27 14:34:59.134')
,('PT DTN THPT huyện Điện Biên','COUNTRY_VN',48,548,NULL,true,'2019-09-27 14:34:59.135','2019-09-27 14:34:59.135')
,('THPT Chuyên Lê Quý Đôn','COUNTRY_VN',48,548,NULL,true,'2019-09-27 14:34:59.135','2019-09-27 14:34:59.135')
,('THPT Phan Đình Giót','COUNTRY_VN',48,548,NULL,true,'2019-09-27 14:34:59.136','2019-09-27 14:34:59.136')
,('THPT thành phố Điện Biên Phủ','COUNTRY_VN',48,548,NULL,true,'2019-09-27 14:34:59.136','2019-09-27 14:34:59.136')
,('Trung tâm GDTX Tỉnh','COUNTRY_VN',48,548,NULL,true,'2019-09-27 14:34:59.136','2019-09-27 14:34:59.136')
,('THPT Thị xã Mường Lay','COUNTRY_VN',48,549,NULL,true,'2019-09-27 14:34:59.137','2019-09-27 14:34:59.137')
,('THCS-THPT Ngọc Lâm','COUNTRY_VN',49,550,NULL,true,'2019-09-27 14:34:59.138','2019-09-27 14:34:59.138')
,('THPT Đắc Lua','COUNTRY_VN',49,550,NULL,true,'2019-09-27 14:34:59.139','2019-09-27 14:34:59.139')
,('THPT Đoàn Kết','COUNTRY_VN',49,550,NULL,true,'2019-09-27 14:34:59.139','2019-09-27 14:34:59.139')
,('THPT Thanh Bình','COUNTRY_VN',49,550,NULL,true,'2019-09-27 14:34:59.139','2019-09-27 14:34:59.139')
,('THPT Tôn Đức Thắng','COUNTRY_VN',49,550,NULL,true,'2019-09-27 14:34:59.139','2019-09-27 14:34:59.139')
,('TT GDTX Tân Phú','COUNTRY_VN',49,550,NULL,true,'2019-09-27 14:34:59.140','2019-09-27 14:34:59.140')
,('THPT Sông Ray','COUNTRY_VN',49,551,NULL,true,'2019-09-27 14:34:59.143','2019-09-27 14:34:59.143')
,('THPT Võ Trường Toản','COUNTRY_VN',49,551,NULL,true,'2019-09-27 14:34:59.144','2019-09-27 14:34:59.144')
,('THPT Xuân Mỹ','COUNTRY_VN',49,551,NULL,true,'2019-09-27 14:34:59.145','2019-09-27 14:34:59.145')
,('TT GDTX Cẩm Mỹ','COUNTRY_VN',49,551,NULL,true,'2019-09-27 14:34:59.145','2019-09-27 14:34:59.145')
,('THCS-THPT Lạc Long Quân','COUNTRY_VN',49,552,NULL,true,'2019-09-27 14:34:59.146','2019-09-27 14:34:59.146')
,('THCS-THPT Tây Sơn','COUNTRY_VN',49,552,NULL,true,'2019-09-27 14:34:59.147','2019-09-27 14:34:59.147')
,('THPT Điểu Cải','COUNTRY_VN',49,552,NULL,true,'2019-09-27 14:34:59.147','2019-09-27 14:34:59.147')
,('THPT Định Quán','COUNTRY_VN',49,552,NULL,true,'2019-09-27 14:34:59.148','2019-09-27 14:34:59.148')
,('THPT Phú Ngọc','COUNTRY_VN',49,552,NULL,true,'2019-09-27 14:34:59.149','2019-09-27 14:34:59.149')
,('THPT Tân Phú','COUNTRY_VN',49,552,NULL,true,'2019-09-27 14:34:59.150','2019-09-27 14:34:59.150')
,('TT GDTX Định Quán','COUNTRY_VN',49,552,NULL,true,'2019-09-27 14:34:59.150','2019-09-27 14:34:59.150')
,('CĐ nghề KV Long Thành-Nhơn Trạch','COUNTRY_VN',49,553,NULL,true,'2019-09-27 14:34:59.151','2019-09-27 14:34:59.151')
,('CĐ nghề LiLaMa2','COUNTRY_VN',49,553,NULL,true,'2019-09-27 14:34:59.151','2019-09-27 14:34:59.151')
,('TC nghề Tri Thức','COUNTRY_VN',49,553,NULL,true,'2019-09-27 14:34:59.152','2019-09-27 14:34:59.152')
,('THPT Bình Sơn','COUNTRY_VN',49,553,NULL,true,'2019-09-27 14:34:59.152','2019-09-27 14:34:59.152')
,('THPT Long Phước','COUNTRY_VN',49,553,NULL,true,'2019-09-27 14:34:59.152','2019-09-27 14:34:59.152')
,('THPT Long Thành','COUNTRY_VN',49,553,NULL,true,'2019-09-27 14:34:59.153','2019-09-27 14:34:59.153')
,('THPT Nguyễn Đình Chiểu','COUNTRY_VN',49,553,NULL,true,'2019-09-27 14:34:59.153','2019-09-27 14:34:59.153')
,('TC Kinh tế- Kỹ thuật Đồng Nai','COUNTRY_VN',49,554,NULL,true,'2019-09-27 14:34:59.154','2019-09-27 14:34:59.154')
,('THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',49,554,NULL,true,'2019-09-27 14:34:59.154','2019-09-27 14:34:59.154')
,('THPT Nhơn Trạch','COUNTRY_VN',49,554,NULL,true,'2019-09-27 14:34:59.154','2019-09-27 14:34:59.154')
,('THPT Phước Thiền','COUNTRY_VN',49,554,NULL,true,'2019-09-27 14:34:59.155','2019-09-27 14:34:59.155')
,('TT GDTX Nhơn Trạch','COUNTRY_VN',49,554,NULL,true,'2019-09-27 14:34:59.155','2019-09-27 14:34:59.155')
,('TH-THCS-THPT Lê Quý Đôn-Tân Phú','COUNTRY_VN',49,555,NULL,true,'2019-09-27 14:34:59.156','2019-09-27 14:34:59.156')
,('THPT Dầu Giây','COUNTRY_VN',49,556,NULL,true,'2019-09-27 14:34:59.157','2019-09-27 14:34:59.157')
,('THPT Kiệm Tân','COUNTRY_VN',49,556,NULL,true,'2019-09-27 14:34:59.157','2019-09-27 14:34:59.157')
,('THPT Thống Nhất B','COUNTRY_VN',49,556,NULL,true,'2019-09-27 14:34:59.159','2019-09-27 14:34:59.159')
,('TT GDTX Thống Nhất','COUNTRY_VN',49,556,NULL,true,'2019-09-27 14:34:59.159','2019-09-27 14:34:59.159')
,('CĐ nghề Cơ giới - Thủy lợi','COUNTRY_VN',49,557,NULL,true,'2019-09-27 14:34:59.160','2019-09-27 14:34:59.160')
,('ĐH Lâm Nghiệp (cơ sở 2)','COUNTRY_VN',49,557,NULL,true,'2019-09-27 14:34:59.161','2019-09-27 14:34:59.161')
,('TC Bách khoa Đồng Nai','COUNTRY_VN',49,557,NULL,true,'2019-09-27 14:34:59.162','2019-09-27 14:34:59.162')
,('TC nghề Hòa Bình','COUNTRY_VN',49,557,NULL,true,'2019-09-27 14:34:59.162','2019-09-27 14:34:59.162')
,('TC nghề Tân Mai','COUNTRY_VN',49,557,NULL,true,'2019-09-27 14:34:59.163','2019-09-27 14:34:59.163')
,('THCSTHPT Bàu Hàm','COUNTRY_VN',49,557,NULL,true,'2019-09-27 14:34:59.164','2019-09-27 14:34:59.164')
,('THPT Dân Tộc Nội Trú tỉnh','COUNTRY_VN',49,557,NULL,true,'2019-09-27 14:34:59.165','2019-09-27 14:34:59.165')
,('TC nghề Cơ Điện Đông Nam Bộ','COUNTRY_VN',49,558,NULL,true,'2019-09-27 14:34:59.166','2019-09-27 14:34:59.166')
,('TH-THCS-THPT Hùng Vuơng','COUNTRY_VN',49,558,NULL,true,'2019-09-27 14:34:59.166','2019-09-27 14:34:59.166')
,('THCS-THPT Huỳnh văn nghệ','COUNTRY_VN',49,558,NULL,true,'2019-09-27 14:34:59.167','2019-09-27 14:34:59.167')
,('THPT Trị An','COUNTRY_VN',49,558,NULL,true,'2019-09-27 14:34:59.167','2019-09-27 14:34:59.167')
,('THPT Vĩnh củu','COUNTRY_VN',49,558,NULL,true,'2019-09-27 14:34:59.167','2019-09-27 14:34:59.167')
,('TT GDTX Vĩnh cửu','COUNTRY_VN',49,558,NULL,true,'2019-09-27 14:34:59.168','2019-09-27 14:34:59.168')
,('THPT DL Hồng Bàng','COUNTRY_VN',49,559,NULL,true,'2019-09-27 14:34:59.169','2019-09-27 14:34:59.169')
,('THPT Xuân Hưng','COUNTRY_VN',49,559,NULL,true,'2019-09-27 14:34:59.169','2019-09-27 14:34:59.169')
,('THPT Xuân Lộc','COUNTRY_VN',49,559,NULL,true,'2019-09-27 14:34:59.170','2019-09-27 14:34:59.170')
,('THPT Xuân Thọ','COUNTRY_VN',49,559,NULL,true,'2019-09-27 14:34:59.170','2019-09-27 14:34:59.170')
,('TT GDTX Xuân Lộc','COUNTRY_VN',49,559,NULL,true,'2019-09-27 14:34:59.170','2019-09-27 14:34:59.170')
,('Bổ Túc Văn Hóa Tỉnh','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.171','2019-09-27 14:34:59.171')
,('CĐ nghề Đồng Nai','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.171','2019-09-27 14:34:59.171')
,('CĐ nghề Miền Đông Nam Bộ','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.171','2019-09-27 14:34:59.171')
,('ĐH Đồng Nai','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.172','2019-09-27 14:34:59.172')
,('PT Năng Khiếu Thể Thao','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.172','2019-09-27 14:34:59.172')
,('TC Miền Đông','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.173','2019-09-27 14:34:59.173')
,('TC nghề 26/3','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.173','2019-09-27 14:34:59.173')
,('TC nghề Đinh Tiên Hoàng','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.173','2019-09-27 14:34:59.173')
,('TC nghề GTVT Đồng Nai','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.174','2019-09-27 14:34:59.174')
,('TC nghề Kinh tế - Kỹ thuật số 2','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.176','2019-09-27 14:34:59.176')
,('TH-THCS-THPT Nguyễn Văn Trỗi','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.176','2019-09-27 14:34:59.176')
,('TH-THCS-THPT Song Ngữ Lạc Hồng','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.177','2019-09-27 14:34:59.177')
,('THCS-THPT Châu á Thái Bình Dương','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.177','2019-09-27 14:34:59.177')
,('THCS-THPT và DN Tân Hòa','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.178','2019-09-27 14:34:59.178')
,('THPT Chu Văn An','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.178','2019-09-27 14:34:59.178')
,('THPT Chuyên Lương Thế Vinh','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.179','2019-09-27 14:34:59.179')
,('THPT DL Bùi Thị Xuân','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.181','2019-09-27 14:34:59.181')
,('THPT Đinh Tiên Hoàng','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.182','2019-09-27 14:34:59.182')
,('THPT Lê Hồng Phong','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.182','2019-09-27 14:34:59.182')
,('THPT Nam Hà','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.183','2019-09-27 14:34:59.183')
,('THPT Ngô Quyền','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.184','2019-09-27 14:34:59.184')
,('THPT Nguyễn Hữu Cảnh','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.184','2019-09-27 14:34:59.184')
,('THPT Nguyễn Trãi','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.184','2019-09-27 14:34:59.184')
,('THPT Tam Hiệp','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.185','2019-09-27 14:34:59.185')
,('THPT Tam Phước','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.185','2019-09-27 14:34:59.185')
,('THPT Trấn Biên','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.186','2019-09-27 14:34:59.186')
,('THPT tư thục Đức Trí','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.186','2019-09-27 14:34:59.186')
,('THPT Tư thục Lê Quý Đôn','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.186','2019-09-27 14:34:59.186')
,('THPT Tư thục Nguyễn Khuyến','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.187','2019-09-27 14:34:59.187')
,('TT GDTX Biên Hòa','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.187','2019-09-27 14:34:59.187')
,('TT GDTX tỉnh Đồng Nai','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.187','2019-09-27 14:34:59.187')
,('TT KTTH Hướng nghiệp Đồng Nai','COUNTRY_VN',49,560,NULL,true,'2019-09-27 14:34:59.188','2019-09-27 14:34:59.188')
,('Năng khiếu TDTT','COUNTRY_VN',50,561,NULL,true,'2019-09-27 14:34:59.189','2019-09-27 14:34:59.189')
,('TC nghề GTVT','COUNTRY_VN',50,561,NULL,true,'2019-09-27 14:34:59.189','2019-09-27 14:34:59.189')
,('THCS và THPT Nguyễn Văn Khải','COUNTRY_VN',50,561,NULL,true,'2019-09-27 14:34:59.189','2019-09-27 14:34:59.189')
,('THPT Cao Lãnh 1','COUNTRY_VN',50,561,NULL,true,'2019-09-27 14:34:59.190','2019-09-27 14:34:59.190')
,('THPT Cao Lãnh 2','COUNTRY_VN',50,561,NULL,true,'2019-09-27 14:34:59.190','2019-09-27 14:34:59.190')
,('THPT chuyên Nguyễn Quang Diêu','COUNTRY_VN',50,561,NULL,true,'2019-09-27 14:34:59.192','2019-09-27 14:34:59.192')
,('THPT Châu Thành 1','COUNTRY_VN',50,562,NULL,true,'2019-09-27 14:34:59.194','2019-09-27 14:34:59.194')
,('THPT Châu Thành 2','COUNTRY_VN',50,562,NULL,true,'2019-09-27 14:34:59.195','2019-09-27 14:34:59.195')
,('THPT Nha Mân','COUNTRY_VN',50,562,NULL,true,'2019-09-27 14:34:59.197','2019-09-27 14:34:59.197')
,('THPT Tân Phú Trung','COUNTRY_VN',50,562,NULL,true,'2019-09-27 14:34:59.198','2019-09-27 14:34:59.198')
,('TT Dạy nghề - GDTX Châu Thành','COUNTRY_VN',50,562,NULL,true,'2019-09-27 14:34:59.199','2019-09-27 14:34:59.199')
,('TC Nghề Hồng Ngự','COUNTRY_VN',50,563,NULL,true,'2019-09-27 14:34:59.200','2019-09-27 14:34:59.200')
,('THPT Chu Văn An','COUNTRY_VN',50,563,NULL,true,'2019-09-27 14:34:59.200','2019-09-27 14:34:59.200')
,('THPT Hồng Ngự 1','COUNTRY_VN',50,563,NULL,true,'2019-09-27 14:34:59.201','2019-09-27 14:34:59.201')
,('THPT Hồng Ngự 2','COUNTRY_VN',50,563,NULL,true,'2019-09-27 14:34:59.202','2019-09-27 14:34:59.202')
,('THPT Hồng Ngự 3','COUNTRY_VN',50,563,NULL,true,'2019-09-27 14:34:59.202','2019-09-27 14:34:59.202')
,('THPT Long Khánh A','COUNTRY_VN',50,563,NULL,true,'2019-09-27 14:34:59.203','2019-09-27 14:34:59.203')
,('Trung cấp nghề - GDTX Hồng Ngự','COUNTRY_VN',50,563,NULL,true,'2019-09-27 14:34:59.203','2019-09-27 14:34:59.203')
,('THPT Lai Vung 1','COUNTRY_VN',50,564,NULL,true,'2019-09-27 14:34:59.204','2019-09-27 14:34:59.204')
,('THPT Lai Vung 2','COUNTRY_VN',50,564,NULL,true,'2019-09-27 14:34:59.205','2019-09-27 14:34:59.205')
,('THPT Lai Vung 3','COUNTRY_VN',50,564,NULL,true,'2019-09-27 14:34:59.205','2019-09-27 14:34:59.205')
,('THPT Phan Văn Bảy','COUNTRY_VN',50,564,NULL,true,'2019-09-27 14:34:59.206','2019-09-27 14:34:59.206')
,('TT Day nghề - GDTX Lai Vung','COUNTRY_VN',50,564,NULL,true,'2019-09-27 14:34:59.206','2019-09-27 14:34:59.206')
,('THCS & THPT Bình Thạnh Trung','COUNTRY_VN',50,565,NULL,true,'2019-09-27 14:34:59.208','2019-09-27 14:34:59.208')
,('THPT Lấp Vò 1','COUNTRY_VN',50,565,NULL,true,'2019-09-27 14:34:59.210','2019-09-27 14:34:59.210')
,('THPT Lấp Vò 2','COUNTRY_VN',50,565,NULL,true,'2019-09-27 14:34:59.211','2019-09-27 14:34:59.211')
,('THPT Lấp Vò 3','COUNTRY_VN',50,565,NULL,true,'2019-09-27 14:34:59.211','2019-09-27 14:34:59.211')
,('THPT Nguyễn Trãi','COUNTRY_VN',50,565,NULL,true,'2019-09-27 14:34:59.212','2019-09-27 14:34:59.212')
,('TT Dạy nghề - GDTX Lấp Vò','COUNTRY_VN',50,565,NULL,true,'2019-09-27 14:34:59.213','2019-09-27 14:34:59.213')
,('THCS và THPT Hòa Bình','COUNTRY_VN',50,566,NULL,true,'2019-09-27 14:34:59.214','2019-09-27 14:34:59.214')
,('THPT Tam Nông','COUNTRY_VN',50,566,NULL,true,'2019-09-27 14:34:59.215','2019-09-27 14:34:59.215')
,('THPT Tràm Chim','COUNTRY_VN',50,566,NULL,true,'2019-09-27 14:34:59.215','2019-09-27 14:34:59.215')
,('TT Dạy nghề - GDTX Tam Nông','COUNTRY_VN',50,566,NULL,true,'2019-09-27 14:34:59.216','2019-09-27 14:34:59.216')
,('THPT Giồng Thị Đam','COUNTRY_VN',50,567,NULL,true,'2019-09-27 14:34:59.217','2019-09-27 14:34:59.217')
,('THPT Tân Hồng','COUNTRY_VN',50,567,NULL,true,'2019-09-27 14:34:59.217','2019-09-27 14:34:59.217')
,('THPT Tân Thành','COUNTRY_VN',50,567,NULL,true,'2019-09-27 14:34:59.218','2019-09-27 14:34:59.218')
,('TT Dạy nghề - GDTX Tân Hồng','COUNTRY_VN',50,567,NULL,true,'2019-09-27 14:34:59.218','2019-09-27 14:34:59.218')
,('TC Nghề Thanh Bình','COUNTRY_VN',50,568,NULL,true,'2019-09-27 14:34:59.219','2019-09-27 14:34:59.219')
,('THPT Thanh Bình 1','COUNTRY_VN',50,568,NULL,true,'2019-09-27 14:34:59.219','2019-09-27 14:34:59.219')
,('THPT Thanh Bình 2','COUNTRY_VN',50,568,NULL,true,'2019-09-27 14:34:59.220','2019-09-27 14:34:59.220')
,('THPT Trần Văn Năng','COUNTRY_VN',50,568,NULL,true,'2019-09-27 14:34:59.220','2019-09-27 14:34:59.220')
,('Trung cấp nghề - GDTX Thanh Bình','COUNTRY_VN',50,568,NULL,true,'2019-09-27 14:34:59.220','2019-09-27 14:34:59.220')
,('TC Nghề Tháp Mười','COUNTRY_VN',50,569,NULL,true,'2019-09-27 14:34:59.221','2019-09-27 14:34:59.221')
,('THPT Đốc Bình Kiều','COUNTRY_VN',50,569,NULL,true,'2019-09-27 14:34:59.221','2019-09-27 14:34:59.221')
,('THPT Mỹ Quý','COUNTRY_VN',50,569,NULL,true,'2019-09-27 14:34:59.221','2019-09-27 14:34:59.221')
,('THPT Phú Điền','COUNTRY_VN',50,569,NULL,true,'2019-09-27 14:34:59.222','2019-09-27 14:34:59.222')
,('THPT Tháp Mười','COUNTRY_VN',50,569,NULL,true,'2019-09-27 14:34:59.222','2019-09-27 14:34:59.222')
,('THPT Trường Xuân','COUNTRY_VN',50,569,NULL,true,'2019-09-27 14:34:59.222','2019-09-27 14:34:59.222')
,('Trung cấp nghề - GDTX Tháp Mười','COUNTRY_VN',50,569,NULL,true,'2019-09-27 14:34:59.223','2019-09-27 14:34:59.223')
,('CĐ nghề Đồng Tháp','COUNTRY_VN',50,570,NULL,true,'2019-09-27 14:34:59.224','2019-09-27 14:34:59.224')
,('THPT Chuyên Nguyễn Đình Chiểu','COUNTRY_VN',50,570,NULL,true,'2019-09-27 14:34:59.225','2019-09-27 14:34:59.225')
,('THPT Nguyễn Du','COUNTRY_VN',50,570,NULL,true,'2019-09-27 14:34:59.226','2019-09-27 14:34:59.226')
,('THPT Thành phố Sa Đéc','COUNTRY_VN',50,570,NULL,true,'2019-09-27 14:34:59.226','2019-09-27 14:34:59.226')
,('TT GDTX Thành phố Sa Đéc','COUNTRY_VN',50,570,NULL,true,'2019-09-27 14:34:59.227','2019-09-27 14:34:59.227')
,('THPT la Ly','COUNTRY_VN',51,571,NULL,true,'2019-09-27 14:34:59.228','2019-09-27 14:34:59.228')
,('THPT Mạc Đĩnh Chi','COUNTRY_VN',51,571,NULL,true,'2019-09-27 14:34:59.229','2019-09-27 14:34:59.229')
,('THPT Phạm Hồng Thái','COUNTRY_VN',51,571,NULL,true,'2019-09-27 14:34:59.230','2019-09-27 14:34:59.230')
,('TT GDTX Chư Păh','COUNTRY_VN',51,571,NULL,true,'2019-09-27 14:34:59.230','2019-09-27 14:34:59.230')
,('THPT Lê Quý Đôn','COUNTRY_VN',51,572,NULL,true,'2019-09-27 14:34:59.232','2019-09-27 14:34:59.232')
,('THPT Pleime','COUNTRY_VN',51,572,NULL,true,'2019-09-27 14:34:59.232','2019-09-27 14:34:59.232')
,('THPT Trần Phú','COUNTRY_VN',51,572,NULL,true,'2019-09-27 14:34:59.233','2019-09-27 14:34:59.233')
,('TT DN & GDTX Chư Prông','COUNTRY_VN',51,572,NULL,true,'2019-09-27 14:34:59.233','2019-09-27 14:34:59.233')
,('THPT Nguyễn Thái Học','COUNTRY_VN',51,573,NULL,true,'2019-09-27 14:34:59.234','2019-09-27 14:34:59.234')
,('TT GDTX-HN Chư Pưh','COUNTRY_VN',51,573,NULL,true,'2019-09-27 14:34:59.234','2019-09-27 14:34:59.234')
,('THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',51,574,NULL,true,'2019-09-27 14:34:59.235','2019-09-27 14:34:59.235')
,('THPT Nguyễn Văn Cừ','COUNTRY_VN',51,574,NULL,true,'2019-09-27 14:34:59.235','2019-09-27 14:34:59.235')
,('THPT Trường Chinh','COUNTRY_VN',51,574,NULL,true,'2019-09-27 14:34:59.236','2019-09-27 14:34:59.236')
,('TT GDTX Chư Sê','COUNTRY_VN',51,574,NULL,true,'2019-09-27 14:34:59.236','2019-09-27 14:34:59.236')
,('THPT Lê Hồng Phong','COUNTRY_VN',51,575,NULL,true,'2019-09-27 14:34:59.237','2019-09-27 14:34:59.237')
,('THPT Nguyễn Huệ','COUNTRY_VN',51,575,NULL,true,'2019-09-27 14:34:59.238','2019-09-27 14:34:59.238')
,('THPT Nguyễn Thị Minh Khai','COUNTRY_VN',51,575,NULL,true,'2019-09-27 14:34:59.238','2019-09-27 14:34:59.238')
,('TT GDTX Đăk Đoa','COUNTRY_VN',51,575,NULL,true,'2019-09-27 14:34:59.239','2019-09-27 14:34:59.239')
,('THPT Y Đôn','COUNTRY_VN',51,576,NULL,true,'2019-09-27 14:34:59.240','2019-09-27 14:34:59.240')
,('TT GDTX Đak Pơ','COUNTRY_VN',51,576,NULL,true,'2019-09-27 14:34:59.240','2019-09-27 14:34:59.240')
,('THPT Lê Hoàn','COUNTRY_VN',51,577,NULL,true,'2019-09-27 14:34:59.242','2019-09-27 14:34:59.242')
,('THPT Nguyễn Trường Tộ','COUNTRY_VN',51,577,NULL,true,'2019-09-27 14:34:59.243','2019-09-27 14:34:59.243')
,('THPT Tôn Đức Thắng','COUNTRY_VN',51,577,NULL,true,'2019-09-27 14:34:59.243','2019-09-27 14:34:59.243')
,('TT GDTX Đúc Cơ','COUNTRY_VN',51,577,NULL,true,'2019-09-27 14:34:59.244','2019-09-27 14:34:59.244')
,('THPT Huỳnh Thúc Kháng','COUNTRY_VN',51,578,NULL,true,'2019-09-27 14:34:59.245','2019-09-27 14:34:59.245')
,('THPT Phạm Văn Đồng','COUNTRY_VN',51,578,NULL,true,'2019-09-27 14:34:59.245','2019-09-27 14:34:59.245')
,('TT DN & GDTX la Grai','COUNTRY_VN',51,578,NULL,true,'2019-09-27 14:34:59.246','2019-09-27 14:34:59.246')
,('THPT Nguyễn Tất Thành','COUNTRY_VN',51,579,NULL,true,'2019-09-27 14:34:59.248','2019-09-27 14:34:59.248')
,('THPT Phan Chu Trinh','COUNTRY_VN',51,579,NULL,true,'2019-09-27 14:34:59.249','2019-09-27 14:34:59.249')
,('TT GDTX-HN la Pa','COUNTRY_VN',51,579,NULL,true,'2019-09-27 14:34:59.249','2019-09-27 14:34:59.249')
,('THPT Anh hùng Núp','COUNTRY_VN',51,580,NULL,true,'2019-09-27 14:34:59.250','2019-09-27 14:34:59.250')
,('THPT Lương Thế Vinh','COUNTRY_VN',51,580,NULL,true,'2019-09-27 14:34:59.250','2019-09-27 14:34:59.250')
,('TT DN & GDTX KBang','COUNTRY_VN',51,580,NULL,true,'2019-09-27 14:34:59.250','2019-09-27 14:34:59.250')
,('THPT Hà Huy Tập','COUNTRY_VN',51,581,NULL,true,'2019-09-27 14:34:59.251','2019-09-27 14:34:59.251')
,('TT GDTX Kông chro','COUNTRY_VN',51,581,NULL,true,'2019-09-27 14:34:59.251','2019-09-27 14:34:59.251')
,('THPT Chu Văn An','COUNTRY_VN',51,582,NULL,true,'2019-09-27 14:34:59.252','2019-09-27 14:34:59.252')
,('THPT Đinh Tiên Hoàng','COUNTRY_VN',51,582,NULL,true,'2019-09-27 14:34:59.253','2019-09-27 14:34:59.253')
,('THPT Nguyễn Du','COUNTRY_VN',51,582,NULL,true,'2019-09-27 14:34:59.253','2019-09-27 14:34:59.253')
,('TT GDTX Krông Pa','COUNTRY_VN',51,582,NULL,true,'2019-09-27 14:34:59.253','2019-09-27 14:34:59.253')
,('THCS & THPT Kpă Klong','COUNTRY_VN',51,583,NULL,true,'2019-09-27 14:34:59.254','2019-09-27 14:34:59.254')
,('THPT Trần Hưng Đạo','COUNTRY_VN',51,583,NULL,true,'2019-09-27 14:34:59.254','2019-09-27 14:34:59.254')
,('TT DN & GDTX Mang Yang','COUNTRY_VN',51,583,NULL,true,'2019-09-27 14:34:59.255','2019-09-27 14:34:59.255')
,('THPT Trần Quốc Tuấn','COUNTRY_VN',51,584,NULL,true,'2019-09-27 14:34:59.255','2019-09-27 14:34:59.255')
,('THPT Võ Văn Kiệt','COUNTRY_VN',51,584,NULL,true,'2019-09-27 14:34:59.256','2019-09-27 14:34:59.256')
,('TT GDTX Phú Thiện','COUNTRY_VN',51,584,NULL,true,'2019-09-27 14:34:59.256','2019-09-27 14:34:59.256')
,('Cao đẳng nghề Gia Lai','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.258','2019-09-27 14:34:59.258')
,('CĐ nghề số 05 Chi nhánh Gia Lai','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.259','2019-09-27 14:34:59.259')
,('PT Dân tộc Nội trú tỉnh','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.259','2019-09-27 14:34:59.259')
,('Quốc tế Châu Á Thái Bình Dương - Gia Lai','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.260','2019-09-27 14:34:59.260')
,('TC nghề số 15','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.260','2019-09-27 14:34:59.260')
,('TC nghề số 21','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.261','2019-09-27 14:34:59.261')
,('TC VH-NT Gia Lai','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.262','2019-09-27 14:34:59.262')
,('TC Y tế Gia Lai','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.262','2019-09-27 14:34:59.262')
,('TH, THCS, THPT Nguyễn văn Linh','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.263','2019-09-27 14:34:59.263')
,('Thiếu sinh quân-Quân khu V','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.263','2019-09-27 14:34:59.263')
,('THPT Chuyên Hùng Vương','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.264','2019-09-27 14:34:59.264')
,('THPT Hoàng Hoa Thám','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.264','2019-09-27 14:34:59.264')
,('THPT Lê Lợi','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.265','2019-09-27 14:34:59.265')
,('THPT Nguyễn Chí Thanh','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.266','2019-09-27 14:34:59.266')
,('THPT Phan Bội Châu','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.266','2019-09-27 14:34:59.266')
,('THPT Pleiku','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.266','2019-09-27 14:34:59.266')
,('TT GDTX tỉnh','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.267','2019-09-27 14:34:59.267')
,('TT Kỹ thuật - Tổng hợp - Hướng nghiệp','COUNTRY_VN',51,585,NULL,true,'2019-09-27 14:34:59.267','2019-09-27 14:34:59.267')
,('TC nghề An Khê','COUNTRY_VN',51,586,NULL,true,'2019-09-27 14:34:59.268','2019-09-27 14:34:59.268')
,('THPT Nguyễn Khuyến','COUNTRY_VN',51,586,NULL,true,'2019-09-27 14:34:59.268','2019-09-27 14:34:59.268')
,('THPT Nguyễn Trãi','COUNTRY_VN',51,586,NULL,true,'2019-09-27 14:34:59.268','2019-09-27 14:34:59.268')
,('THPT Quang Trung','COUNTRY_VN',51,586,NULL,true,'2019-09-27 14:34:59.269','2019-09-27 14:34:59.269')
,('TT GDTX An Khê','COUNTRY_VN',51,586,NULL,true,'2019-09-27 14:34:59.269','2019-09-27 14:34:59.269')
,('TC nghề Ayun Pa','COUNTRY_VN',51,587,NULL,true,'2019-09-27 14:34:59.270','2019-09-27 14:34:59.270')
,('THPT Lê Thánh Tông','COUNTRY_VN',51,587,NULL,true,'2019-09-27 14:34:59.270','2019-09-27 14:34:59.270')
,('THPT Lý Thường Kiệt','COUNTRY_VN',51,587,NULL,true,'2019-09-27 14:34:59.270','2019-09-27 14:34:59.270')
,('TT GDTX Ayun Pa','COUNTRY_VN',51,587,NULL,true,'2019-09-27 14:34:59.271','2019-09-27 14:34:59.271')
,('GDTX Bắc Mê','COUNTRY_VN',52,588,NULL,true,'2019-09-27 14:34:59.272','2019-09-27 14:34:59.272')
,('THCS và THPT Minh Ngọc','COUNTRY_VN',52,588,NULL,true,'2019-09-27 14:34:59.273','2019-09-27 14:34:59.273')
,('THPT Bắc Mê','COUNTRY_VN',52,588,NULL,true,'2019-09-27 14:34:59.273','2019-09-27 14:34:59.273')
,('GDTX Bắc Quang','COUNTRY_VN',52,589,NULL,true,'2019-09-27 14:34:59.275','2019-09-27 14:34:59.275')
,('PT Cấp 2-3 Tân Quang','COUNTRY_VN',52,589,NULL,true,'2019-09-27 14:34:59.276','2019-09-27 14:34:59.276')
,('PT DTNT cấp 2-3 Bắc Quang','COUNTRY_VN',52,589,NULL,true,'2019-09-27 14:34:59.276','2019-09-27 14:34:59.276')
,('THPT Đồng Yên','COUNTRY_VN',52,589,NULL,true,'2019-09-27 14:34:59.277','2019-09-27 14:34:59.277')
,('THPT Hùng An','COUNTRY_VN',52,589,NULL,true,'2019-09-27 14:34:59.277','2019-09-27 14:34:59.277')
,('THPT Kim Ngọc','COUNTRY_VN',52,589,NULL,true,'2019-09-27 14:34:59.278','2019-09-27 14:34:59.278')
,('THPT Liên Hiệp','COUNTRY_VN',52,589,NULL,true,'2019-09-27 14:34:59.278','2019-09-27 14:34:59.278')
,('GDTX Đồng Văn','COUNTRY_VN',52,590,NULL,true,'2019-09-27 14:34:59.279','2019-09-27 14:34:59.279')
,('THPT Đồng Văn','COUNTRY_VN',52,590,NULL,true,'2019-09-27 14:34:59.280','2019-09-27 14:34:59.280')
,('GDTX Hoàng Su Phì','COUNTRY_VN',52,591,NULL,true,'2019-09-27 14:34:59.281','2019-09-27 14:34:59.281')
,('THPT Hoàng Su Phì','COUNTRY_VN',52,591,NULL,true,'2019-09-27 14:34:59.282','2019-09-27 14:34:59.282')
,('THPT Thông Nguyên','COUNTRY_VN',52,591,NULL,true,'2019-09-27 14:34:59.282','2019-09-27 14:34:59.282')
,('GDTX Mèo Vạc','COUNTRY_VN',52,592,NULL,true,'2019-09-27 14:34:59.283','2019-09-27 14:34:59.283')
,('THPT Mèo Vạc','COUNTRY_VN',52,592,NULL,true,'2019-09-27 14:34:59.283','2019-09-27 14:34:59.283')
,('GDTX Quản Bạ','COUNTRY_VN',52,593,NULL,true,'2019-09-27 14:34:59.284','2019-09-27 14:34:59.284')
,('THPT Quản Bạ','COUNTRY_VN',52,593,NULL,true,'2019-09-27 14:34:59.285','2019-09-27 14:34:59.285')
,('THPT Quyết Tiến','COUNTRY_VN',52,593,NULL,true,'2019-09-27 14:34:59.285','2019-09-27 14:34:59.285')
,('GDTX Quang Bình','COUNTRY_VN',52,594,NULL,true,'2019-09-27 14:34:59.286','2019-09-27 14:34:59.286')
,('THPT Quang Bình','COUNTRY_VN',52,594,NULL,true,'2019-09-27 14:34:59.286','2019-09-27 14:34:59.286')
,('THPT Xuân Giang','COUNTRY_VN',52,594,NULL,true,'2019-09-27 14:34:59.287','2019-09-27 14:34:59.287')
,('GDTX Vị Xuyên','COUNTRY_VN',52,595,NULL,true,'2019-09-27 14:34:59.288','2019-09-27 14:34:59.288')
,('PT Cấp 2-3 Phương Tiến','COUNTRY_VN',52,595,NULL,true,'2019-09-27 14:34:59.288','2019-09-27 14:34:59.288')
,('THCS & THPT Linh Hồ','COUNTRY_VN',52,595,NULL,true,'2019-09-27 14:34:59.289','2019-09-27 14:34:59.289')
,('THCS & THPT Tùng Bá','COUNTRY_VN',52,595,NULL,true,'2019-09-27 14:34:59.289','2019-09-27 14:34:59.289')
,('THPT Vị Xuyên','COUNTRY_VN',52,595,NULL,true,'2019-09-27 14:34:59.289','2019-09-27 14:34:59.289')
,('THPT Việt Lâm','COUNTRY_VN',52,595,NULL,true,'2019-09-27 14:34:59.290','2019-09-27 14:34:59.290')
,('GDTX Xín Mần','COUNTRY_VN',52,596,NULL,true,'2019-09-27 14:34:59.292','2019-09-27 14:34:59.292')
,('THCS và THPT Nà chì','COUNTRY_VN',52,596,NULL,true,'2019-09-27 14:34:59.293','2019-09-27 14:34:59.293')
,('THPT Xín Mần','COUNTRY_VN',52,596,NULL,true,'2019-09-27 14:34:59.293','2019-09-27 14:34:59.293')
,('GDTX Yên Minh','COUNTRY_VN',52,597,NULL,true,'2019-09-27 14:34:59.294','2019-09-27 14:34:59.294')
,('PT DTNT cấp 2-3 Yên Minh','COUNTRY_VN',52,597,NULL,true,'2019-09-27 14:34:59.295','2019-09-27 14:34:59.295')
,('THPT Mậu Duệ','COUNTRY_VN',52,597,NULL,true,'2019-09-27 14:34:59.296','2019-09-27 14:34:59.296')
,('THPT Yên Minh','COUNTRY_VN',52,597,NULL,true,'2019-09-27 14:34:59.297','2019-09-27 14:34:59.297')
,('CĐ Nghề Hà Giang','COUNTRY_VN',52,598,NULL,true,'2019-09-27 14:34:59.298','2019-09-27 14:34:59.298')
,('CĐSP Hà Giang','COUNTRY_VN',52,598,NULL,true,'2019-09-27 14:34:59.299','2019-09-27 14:34:59.299')
,('GDTX Tỉnh','COUNTRY_VN',52,598,NULL,true,'2019-09-27 14:34:59.299','2019-09-27 14:34:59.299')
,('PTDT Nội trú tỉnh','COUNTRY_VN',52,598,NULL,true,'2019-09-27 14:34:59.299','2019-09-27 14:34:59.299')
,('THPT Chuyên','COUNTRY_VN',52,598,NULL,true,'2019-09-27 14:34:59.300','2019-09-27 14:34:59.300')
,('THPT Lê Hồng Phong','COUNTRY_VN',52,598,NULL,true,'2019-09-27 14:34:59.300','2019-09-27 14:34:59.300')
,('THPT Ngọc Hà','COUNTRY_VN',52,598,NULL,true,'2019-09-27 14:34:59.301','2019-09-27 14:34:59.301')
,('THPT ABình Lục','COUNTRY_VN',53,599,NULL,true,'2019-09-27 14:34:59.302','2019-09-27 14:34:59.302')
,('THPT B Bình Lục','COUNTRY_VN',53,599,NULL,true,'2019-09-27 14:34:59.303','2019-09-27 14:34:59.303')
,('THPT CBình Lục','COUNTRY_VN',53,599,NULL,true,'2019-09-27 14:34:59.303','2019-09-27 14:34:59.303')
,('THPT Dân lập Bình Lục','COUNTRY_VN',53,599,NULL,true,'2019-09-27 14:34:59.303','2019-09-27 14:34:59.303')
,('THPT Nguyễn Khuyến','COUNTRY_VN',53,599,NULL,true,'2019-09-27 14:34:59.304','2019-09-27 14:34:59.304')
,('Trung tâm GDTX Bình Lục','COUNTRY_VN',53,599,NULL,true,'2019-09-27 14:34:59.304','2019-09-27 14:34:59.304')
,('THPT ADuy Tiên','COUNTRY_VN',53,600,NULL,true,'2019-09-27 14:34:59.305','2019-09-27 14:34:59.305')
,('THPT BDuy Tiên','COUNTRY_VN',53,600,NULL,true,'2019-09-27 14:34:59.305','2019-09-27 14:34:59.305')
,('THPT CDuy Tiên','COUNTRY_VN',53,600,NULL,true,'2019-09-27 14:34:59.306','2019-09-27 14:34:59.306')
,('THPT Nguyễn HQu Tiến','COUNTRY_VN',53,600,NULL,true,'2019-09-27 14:34:59.306','2019-09-27 14:34:59.306')
,('Trung tâm GDTX Duy Tiên','COUNTRY_VN',53,600,NULL,true,'2019-09-27 14:34:59.306','2019-09-27 14:34:59.306')
,('THPT A Kim Bảng','COUNTRY_VN',53,601,NULL,true,'2019-09-27 14:34:59.308','2019-09-27 14:34:59.308')
,('THPT B Kim Bầng','COUNTRY_VN',53,601,NULL,true,'2019-09-27 14:34:59.309','2019-09-27 14:34:59.309')
,('THPT CKim Bảng','COUNTRY_VN',53,601,NULL,true,'2019-09-27 14:34:59.310','2019-09-27 14:34:59.310')
,('THPT Lý Thường Kiệt','COUNTRY_VN',53,601,NULL,true,'2019-09-27 14:34:59.310','2019-09-27 14:34:59.310')
,('Trung tâm GDTX Kim Bảng','COUNTRY_VN',53,601,NULL,true,'2019-09-27 14:34:59.311','2019-09-27 14:34:59.311')
,('THPT Bắc Lý','COUNTRY_VN',53,602,NULL,true,'2019-09-27 14:34:59.312','2019-09-27 14:34:59.312')
,('THPT Dân lập Trần Hung Đạo','COUNTRY_VN',53,602,NULL,true,'2019-09-27 14:34:59.313','2019-09-27 14:34:59.313')
,('THPT Lý Nhân','COUNTRY_VN',53,602,NULL,true,'2019-09-27 14:34:59.314','2019-09-27 14:34:59.314')
,('THPT Nam Cao','COUNTRY_VN',53,602,NULL,true,'2019-09-27 14:34:59.314','2019-09-27 14:34:59.314')
,('THPT Nam Lý','COUNTRY_VN',53,602,NULL,true,'2019-09-27 14:34:59.315','2019-09-27 14:34:59.315')
,('Trung tâm GDTX Lý Nhân','COUNTRY_VN',53,602,NULL,true,'2019-09-27 14:34:59.316','2019-09-27 14:34:59.316')
,('THPT A Thanh Liêm','COUNTRY_VN',53,603,NULL,true,'2019-09-27 14:34:59.317','2019-09-27 14:34:59.317')
,('THPT B Thanh Liêm','COUNTRY_VN',53,603,NULL,true,'2019-09-27 14:34:59.317','2019-09-27 14:34:59.317')
,('THPT C Thanh Liêm','COUNTRY_VN',53,603,NULL,true,'2019-09-27 14:34:59.318','2019-09-27 14:34:59.318')
,('THPT Dân lập Thanh Liêm','COUNTRY_VN',53,603,NULL,true,'2019-09-27 14:34:59.318','2019-09-27 14:34:59.318')
,('THPT Lê Hoàn','COUNTRY_VN',53,603,NULL,true,'2019-09-27 14:34:59.318','2019-09-27 14:34:59.318')
,('Trung tâm GDTX Thanh Liêm','COUNTRY_VN',53,603,NULL,true,'2019-09-27 14:34:59.319','2019-09-27 14:34:59.319')
,('Cao đẳng nghề Hà Nam','COUNTRY_VN',53,604,NULL,true,'2019-09-27 14:34:59.319','2019-09-27 14:34:59.319')
,('THPT APhủLý','COUNTRY_VN',53,604,NULL,true,'2019-09-27 14:34:59.320','2019-09-27 14:34:59.320')
,('THPT B Phủ Lý','COUNTRY_VN',53,604,NULL,true,'2019-09-27 14:34:59.320','2019-09-27 14:34:59.320')
,('THPT c Phủ Lý','COUNTRY_VN',53,604,NULL,true,'2019-09-27 14:34:59.321','2019-09-27 14:34:59.321')
,('THPT Chuyên Biên Hòa','COUNTRY_VN',53,604,NULL,true,'2019-09-27 14:34:59.321','2019-09-27 14:34:59.321')
,('THPT Dân lập Lương Thế Vinh','COUNTRY_VN',53,604,NULL,true,'2019-09-27 14:34:59.322','2019-09-27 14:34:59.322')
,('Trung tâm GDTX Tỉnh Hà Nam','COUNTRY_VN',53,604,NULL,true,'2019-09-27 14:34:59.322','2019-09-27 14:34:59.322')
,('THPT Cẩm Bình','COUNTRY_VN',54,605,NULL,true,'2019-09-27 14:34:59.323','2019-09-27 14:34:59.323')
,('THPT Cẩm Xuyên','COUNTRY_VN',54,605,NULL,true,'2019-09-27 14:34:59.323','2019-09-27 14:34:59.323')
,('THPT Hà Huy Tập','COUNTRY_VN',54,605,NULL,true,'2019-09-27 14:34:59.325','2019-09-27 14:34:59.325')
,('THPT Nguyễn Đình Liền','COUNTRY_VN',54,605,NULL,true,'2019-09-27 14:34:59.326','2019-09-27 14:34:59.326')
,('THPT Phan Đình Giót','COUNTRY_VN',54,605,NULL,true,'2019-09-27 14:34:59.327','2019-09-27 14:34:59.327')
,('TT DN-HN và GDTX cẩm Xuyên','COUNTRY_VN',54,605,NULL,true,'2019-09-27 14:34:59.328','2019-09-27 14:34:59.328')
,('THPT Can Lộc','COUNTRY_VN',54,606,NULL,true,'2019-09-27 14:34:59.330','2019-09-27 14:34:59.330')
,('THPT DL Can Lộc','COUNTRY_VN',54,606,NULL,true,'2019-09-27 14:34:59.331','2019-09-27 14:34:59.331')
,('THPT Đồng Lộc','COUNTRY_VN',54,606,NULL,true,'2019-09-27 14:34:59.332','2019-09-27 14:34:59.332')
,('THPT nghền','COUNTRY_VN',54,606,NULL,true,'2019-09-27 14:34:59.332','2019-09-27 14:34:59.332')
,('TT DN-HN và GDTX Can Lộc','COUNTRY_VN',54,606,NULL,true,'2019-09-27 14:34:59.333','2019-09-27 14:34:59.333')
,('THPT Đức Thọ','COUNTRY_VN',54,607,NULL,true,'2019-09-27 14:34:59.334','2019-09-27 14:34:59.334')
,('THPT Lê Hồng Phong','COUNTRY_VN',54,607,NULL,true,'2019-09-27 14:34:59.334','2019-09-27 14:34:59.334')
,('THPT Nguyễn Thị Minh Khai','COUNTRY_VN',54,607,NULL,true,'2019-09-27 14:34:59.335','2019-09-27 14:34:59.335')
,('THPT Trần Phú','COUNTRY_VN',54,607,NULL,true,'2019-09-27 14:34:59.336','2019-09-27 14:34:59.336')
,('TT DN-HN và GDTX Đức Thọ','COUNTRY_VN',54,607,NULL,true,'2019-09-27 14:34:59.336','2019-09-27 14:34:59.336')
,('THPT Gia Phố','COUNTRY_VN',54,608,NULL,true,'2019-09-27 14:34:59.337','2019-09-27 14:34:59.337')
,('THPT Hàm Nghi','COUNTRY_VN',54,608,NULL,true,'2019-09-27 14:34:59.337','2019-09-27 14:34:59.337')
,('THPT Huưng Khê','COUNTRY_VN',54,608,NULL,true,'2019-09-27 14:34:59.337','2019-09-27 14:34:59.337')
,('THPT Phúc Trạch','COUNTRY_VN',54,608,NULL,true,'2019-09-27 14:34:59.338','2019-09-27 14:34:59.338')
,('TT DN-HN và GDTX Hương Khê','COUNTRY_VN',54,608,NULL,true,'2019-09-27 14:34:59.338','2019-09-27 14:34:59.338')
,('THPT Cao Thắng','COUNTRY_VN',54,609,NULL,true,'2019-09-27 14:34:59.339','2019-09-27 14:34:59.339')
,('THPT Huưng Sơn','COUNTRY_VN',54,609,NULL,true,'2019-09-27 14:34:59.339','2019-09-27 14:34:59.339')
,('THPT Lê Hữu Trác','COUNTRY_VN',54,609,NULL,true,'2019-09-27 14:34:59.340','2019-09-27 14:34:59.340')
,('THPT Lý Chính Thẳng','COUNTRY_VN',54,609,NULL,true,'2019-09-27 14:34:59.340','2019-09-27 14:34:59.340')
,('THPT DL Nguyễn Khắc Viện','COUNTRY_VN',54,609,NULL,true,'2019-09-27 14:34:59.341','2019-09-27 14:34:59.341')
,('TT DN-HN và GDTX Hương Sơn','COUNTRY_VN',54,609,NULL,true,'2019-09-27 14:34:59.344','2019-09-27 14:34:59.344')
,('THPT Kỳ Anh','COUNTRY_VN',54,610,NULL,true,'2019-09-27 14:34:59.345','2019-09-27 14:34:59.345')
,('THPT Kỳ Lâm','COUNTRY_VN',54,610,NULL,true,'2019-09-27 14:34:59.346','2019-09-27 14:34:59.346')
,('THPT Lê Quảng Chí','COUNTRY_VN',54,610,NULL,true,'2019-09-27 14:34:59.346','2019-09-27 14:34:59.346')
,('THPT Nguyễn Huệ','COUNTRY_VN',54,610,NULL,true,'2019-09-27 14:34:59.348','2019-09-27 14:34:59.348')
,('THPT Nguyễn Thị Bích Châu','COUNTRY_VN',54,610,NULL,true,'2019-09-27 14:34:59.349','2019-09-27 14:34:59.349')
,('TT DN-HN và GDTX Kỳ Anh','COUNTRY_VN',54,610,NULL,true,'2019-09-27 14:34:59.350','2019-09-27 14:34:59.350')
,('THPT Mai Thúc Loan','COUNTRY_VN',54,611,NULL,true,'2019-09-27 14:34:59.352','2019-09-27 14:34:59.352')
,('THPT Nguyễn Đổng Chi','COUNTRY_VN',54,611,NULL,true,'2019-09-27 14:34:59.352','2019-09-27 14:34:59.352')
,('THPT Nguyễn Văn Trỗi','COUNTRY_VN',54,611,NULL,true,'2019-09-27 14:34:59.353','2019-09-27 14:34:59.353')
,('TT DN-HN và GDTX Lộc Hà','COUNTRY_VN',54,611,NULL,true,'2019-09-27 14:34:59.353','2019-09-27 14:34:59.353')
,('THPT Nghi Xuân','COUNTRY_VN',54,612,NULL,true,'2019-09-27 14:34:59.354','2019-09-27 14:34:59.354')
,('THPT Nguyễn Công Trứ','COUNTRY_VN',54,612,NULL,true,'2019-09-27 14:34:59.355','2019-09-27 14:34:59.355')
,('THPT Nguyễn Du','COUNTRY_VN',54,612,NULL,true,'2019-09-27 14:34:59.355','2019-09-27 14:34:59.355')
,('TT DN-HN và GDTX Nghi Xuân','COUNTRY_VN',54,612,NULL,true,'2019-09-27 14:34:59.356','2019-09-27 14:34:59.356')
,('THPT Lê Quý Đôn','COUNTRY_VN',54,613,NULL,true,'2019-09-27 14:34:59.356','2019-09-27 14:34:59.356')
,('THPT Lý Tự Trọng','COUNTRY_VN',54,613,NULL,true,'2019-09-27 14:34:59.357','2019-09-27 14:34:59.357')
,('THPT Mai Kính','COUNTRY_VN',54,613,NULL,true,'2019-09-27 14:34:59.359','2019-09-27 14:34:59.359')
,('THPT Nguyễn Trung Thiên','COUNTRY_VN',54,613,NULL,true,'2019-09-27 14:34:59.360','2019-09-27 14:34:59.360')
,('Trung tâm DN-HN và GDTX Thạch Hà','COUNTRY_VN',54,613,NULL,true,'2019-09-27 14:34:59.360','2019-09-27 14:34:59.360')
,('THPT CÙ Huy cận','COUNTRY_VN',54,614,NULL,true,'2019-09-27 14:34:59.361','2019-09-27 14:34:59.361')
,('THPT Vũ Quang','COUNTRY_VN',54,614,NULL,true,'2019-09-27 14:34:59.362','2019-09-27 14:34:59.362')
,('TT DN-HN và GDTX vũ Quang','COUNTRY_VN',54,614,NULL,true,'2019-09-27 14:34:59.363','2019-09-27 14:34:59.363')
,('Cao đẳng Nghề công nghệ Hà Tĩnh','COUNTRY_VN',54,615,NULL,true,'2019-09-27 14:34:59.364','2019-09-27 14:34:59.364')
,('Cao đẳng nghề Việt Đức Hà Tĩnh','COUNTRY_VN',54,615,NULL,true,'2019-09-27 14:34:59.364','2019-09-27 14:34:59.364')
,('THPT Chuyên Hà Tĩnh','COUNTRY_VN',54,615,NULL,true,'2019-09-27 14:34:59.365','2019-09-27 14:34:59.365')
,('THPT ISCHOOL Hà Tĩnh','COUNTRY_VN',54,615,NULL,true,'2019-09-27 14:34:59.366','2019-09-27 14:34:59.366')
,('THPT Phan Đình Phùng','COUNTRY_VN',54,615,NULL,true,'2019-09-27 14:34:59.366','2019-09-27 14:34:59.366')
,('THPT Thành Sen','COUNTRY_VN',54,615,NULL,true,'2019-09-27 14:34:59.367','2019-09-27 14:34:59.367')
,('Trung cấp Nghề Hà Tĩnh','COUNTRY_VN',54,615,NULL,true,'2019-09-27 14:34:59.368','2019-09-27 14:34:59.368')
,('TT BDNVSP và GDTX tỉnh Hà Tĩnh','COUNTRY_VN',54,615,NULL,true,'2019-09-27 14:34:59.368','2019-09-27 14:34:59.368')
,('TT DN- HN và GDTX TP Hà Tĩnh','COUNTRY_VN',54,615,NULL,true,'2019-09-27 14:34:59.369','2019-09-27 14:34:59.369')
,('THPT Bình Giang','COUNTRY_VN',55,616,NULL,true,'2019-09-27 14:34:59.370','2019-09-27 14:34:59.370')
,('THPT Đường An','COUNTRY_VN',55,616,NULL,true,'2019-09-27 14:34:59.370','2019-09-27 14:34:59.370')
,('THPT Kẻ Sặt','COUNTRY_VN',55,616,NULL,true,'2019-09-27 14:34:59.370','2019-09-27 14:34:59.370')
,('THPT Vũ Ngọc Phan','COUNTRY_VN',55,616,NULL,true,'2019-09-27 14:34:59.371','2019-09-27 14:34:59.371')
,('TT GDTX Bình Giang','COUNTRY_VN',55,616,NULL,true,'2019-09-27 14:34:59.371','2019-09-27 14:34:59.371')
,('THPT Cẩm Giàng','COUNTRY_VN',55,617,NULL,true,'2019-09-27 14:34:59.372','2019-09-27 14:34:59.372')
,('THPT Cẩm Giàng II','COUNTRY_VN',55,617,NULL,true,'2019-09-27 14:34:59.372','2019-09-27 14:34:59.372')
,('THPT Tuệ Tĩnh','COUNTRY_VN',55,617,NULL,true,'2019-09-27 14:34:59.373','2019-09-27 14:34:59.373')
,('TT GDTX Cẩm Giàng','COUNTRY_VN',55,617,NULL,true,'2019-09-27 14:34:59.373','2019-09-27 14:34:59.373')
,('THPT Bến Tắm','COUNTRY_VN',55,618,NULL,true,'2019-09-27 14:34:59.374','2019-09-27 14:34:59.374')
,('THPT Chí Linh','COUNTRY_VN',55,618,NULL,true,'2019-09-27 14:34:59.376','2019-09-27 14:34:59.376')
,('THPT Phả Lại','COUNTRY_VN',55,618,NULL,true,'2019-09-27 14:34:59.377','2019-09-27 14:34:59.377')
,('THPT Trần Phú','COUNTRY_VN',55,618,NULL,true,'2019-09-27 14:34:59.378','2019-09-27 14:34:59.378')
,('TT GDTX-HN-DN Chí Linh','COUNTRY_VN',55,618,NULL,true,'2019-09-27 14:34:59.378','2019-09-27 14:34:59.378')
,('THPT Đoàn Thượng','COUNTRY_VN',55,619,NULL,true,'2019-09-27 14:34:59.380','2019-09-27 14:34:59.380')
,('THPT Gia Lộc','COUNTRY_VN',55,619,NULL,true,'2019-09-27 14:34:59.381','2019-09-27 14:34:59.381')
,('THPT Gia Lộc II','COUNTRY_VN',55,619,NULL,true,'2019-09-27 14:34:59.382','2019-09-27 14:34:59.382')
,('TT GDTX Gia Lộc','COUNTRY_VN',55,619,NULL,true,'2019-09-27 14:34:59.382','2019-09-27 14:34:59.382')
,('THPT Đồng Gia','COUNTRY_VN',55,620,NULL,true,'2019-09-27 14:34:59.384','2019-09-27 14:34:59.384')
,('THPT Kim Thành','COUNTRY_VN',55,620,NULL,true,'2019-09-27 14:34:59.384','2019-09-27 14:34:59.384')
,('THPT Kim Thành II','COUNTRY_VN',55,620,NULL,true,'2019-09-27 14:34:59.384','2019-09-27 14:34:59.384')
,('THPT Phú Thái','COUNTRY_VN',55,620,NULL,true,'2019-09-27 14:34:59.385','2019-09-27 14:34:59.385')
,('TT GDTX Kim Thành','COUNTRY_VN',55,620,NULL,true,'2019-09-27 14:34:59.385','2019-09-27 14:34:59.385')
,('THPT Kinh Môn','COUNTRY_VN',55,621,NULL,true,'2019-09-27 14:34:59.386','2019-09-27 14:34:59.386')
,('THPT Kinh Môn II','COUNTRY_VN',55,621,NULL,true,'2019-09-27 14:34:59.387','2019-09-27 14:34:59.387')
,('THPT Nhị Chiểu','COUNTRY_VN',55,621,NULL,true,'2019-09-27 14:34:59.387','2019-09-27 14:34:59.387')
,('THPT Phúc Thành','COUNTRY_VN',55,621,NULL,true,'2019-09-27 14:34:59.387','2019-09-27 14:34:59.387')
,('THPT Quang Thành','COUNTRY_VN',55,621,NULL,true,'2019-09-27 14:34:59.388','2019-09-27 14:34:59.388')
,('THPT Trần Quang Khải','COUNTRY_VN',55,621,NULL,true,'2019-09-27 14:34:59.388','2019-09-27 14:34:59.388')
,('TT GDTX Kinh Môn','COUNTRY_VN',55,621,NULL,true,'2019-09-27 14:34:59.388','2019-09-27 14:34:59.388')
,('THPT Mạc Đĩnh Chi','COUNTRY_VN',55,622,NULL,true,'2019-09-27 14:34:59.389','2019-09-27 14:34:59.389')
,('THPT Nam Sách','COUNTRY_VN',55,622,NULL,true,'2019-09-27 14:34:59.389','2019-09-27 14:34:59.389')
,('THPT Nam Sách II','COUNTRY_VN',55,622,NULL,true,'2019-09-27 14:34:59.390','2019-09-27 14:34:59.390')
,('THPT Phan Bội Châu','COUNTRY_VN',55,622,NULL,true,'2019-09-27 14:34:59.390','2019-09-27 14:34:59.390')
,('TT GDTX Nam Sách','COUNTRY_VN',55,622,NULL,true,'2019-09-27 14:34:59.392','2019-09-27 14:34:59.392')
,('THPT Hồng Đức','COUNTRY_VN',55,623,NULL,true,'2019-09-27 14:34:59.394','2019-09-27 14:34:59.394')
,('THPT Khúc Thừa Dụ','COUNTRY_VN',55,623,NULL,true,'2019-09-27 14:34:59.394','2019-09-27 14:34:59.394')
,('THPT Ninh Giang','COUNTRY_VN',55,623,NULL,true,'2019-09-27 14:34:59.395','2019-09-27 14:34:59.395')
,('THPT Ninh Giang II','COUNTRY_VN',55,623,NULL,true,'2019-09-27 14:34:59.396','2019-09-27 14:34:59.396')
,('THPT Quang Trung','COUNTRY_VN',55,623,NULL,true,'2019-09-27 14:34:59.397','2019-09-27 14:34:59.397')
,('TT GDTX Ninh Giang','COUNTRY_VN',55,623,NULL,true,'2019-09-27 14:34:59.398','2019-09-27 14:34:59.398')
,('THPT Ha Bắc','COUNTRY_VN',55,624,NULL,true,'2019-09-27 14:34:59.399','2019-09-27 14:34:59.399')
,('THPT Ha Dong','COUNTRY_VN',55,624,NULL,true,'2019-09-27 14:34:59.399','2019-09-27 14:34:59.399')
,('THPT Thanh Binh','COUNTRY_VN',55,624,NULL,true,'2019-09-27 14:34:59.400','2019-09-27 14:34:59.400')
,('THPT Thanh Ha','COUNTRY_VN',55,624,NULL,true,'2019-09-27 14:34:59.400','2019-09-27 14:34:59.400')
,('TT GDTX Thanh Ha','COUNTRY_VN',55,624,NULL,true,'2019-09-27 14:34:59.401','2019-09-27 14:34:59.401')
,('THPT Lê Quý Đôn','COUNTRY_VN',55,625,NULL,true,'2019-09-27 14:34:59.402','2019-09-27 14:34:59.402')
,('THPT Thanh Miện','COUNTRY_VN',55,625,NULL,true,'2019-09-27 14:34:59.402','2019-09-27 14:34:59.402')
,('THPT Thanh Miện 2','COUNTRY_VN',55,625,NULL,true,'2019-09-27 14:34:59.403','2019-09-27 14:34:59.403')
,('THPT Thanh Miện 3','COUNTRY_VN',55,625,NULL,true,'2019-09-27 14:34:59.403','2019-09-27 14:34:59.403')
,('TT GDTX Thanh Miện','COUNTRY_VN',55,625,NULL,true,'2019-09-27 14:34:59.404','2019-09-27 14:34:59.404')
,('THPT Cầu Xe','COUNTRY_VN',55,626,NULL,true,'2019-09-27 14:34:59.405','2019-09-27 14:34:59.405')
,('THPT Hung Đạo','COUNTRY_VN',55,626,NULL,true,'2019-09-27 14:34:59.405','2019-09-27 14:34:59.405')
,('THPT Tứ Kỳ','COUNTRY_VN',55,626,NULL,true,'2019-09-27 14:34:59.405','2019-09-27 14:34:59.405')
,('THPT Tứ Kỳ II','COUNTRY_VN',55,626,NULL,true,'2019-09-27 14:34:59.406','2019-09-27 14:34:59.406')
,('TT GDTX Tứ Kỳ','COUNTRY_VN',55,626,NULL,true,'2019-09-27 14:34:59.406','2019-09-27 14:34:59.406')
,('THPT ái Quốc','COUNTRY_VN',55,627,NULL,true,'2019-09-27 14:34:59.410','2019-09-27 14:34:59.410')
,('THPT Hoàng Văn Thụ','COUNTRY_VN',55,627,NULL,true,'2019-09-27 14:34:59.411','2019-09-27 14:34:59.411')
,('THPT Hồng Quang','COUNTRY_VN',55,627,NULL,true,'2019-09-27 14:34:59.412','2019-09-27 14:34:59.412')
,('THPT Lương Thế Vinh','COUNTRY_VN',55,627,NULL,true,'2019-09-27 14:34:59.412','2019-09-27 14:34:59.412')
,('THPT Marie Curie','COUNTRY_VN',55,627,NULL,true,'2019-09-27 14:34:59.413','2019-09-27 14:34:59.413')
,('THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',55,627,NULL,true,'2019-09-27 14:34:59.414','2019-09-27 14:34:59.414')
,('THPT Nguyễn Du','COUNTRY_VN',55,627,NULL,true,'2019-09-27 14:34:59.414','2019-09-27 14:34:59.414')
,('THPT Nguyễn Trãi','COUNTRY_VN',55,627,NULL,true,'2019-09-27 14:34:59.415','2019-09-27 14:34:59.415')
,('THPT Thành Đông','COUNTRY_VN',55,627,NULL,true,'2019-09-27 14:34:59.415','2019-09-27 14:34:59.415')
,('TT GDTX TP.Hải Dương','COUNTRY_VN',55,627,NULL,true,'2019-09-27 14:34:59.416','2019-09-27 14:34:59.416')
,('CĐ Nghề Trần Đại Nghĩa','COUNTRY_VN',56,628,NULL,true,'2019-09-27 14:34:59.418','2019-09-27 14:34:59.418')
,('THPT Ngã sáu','COUNTRY_VN',56,628,NULL,true,'2019-09-27 14:34:59.418','2019-09-27 14:34:59.418')
,('THPT Phú Hữu','COUNTRY_VN',56,628,NULL,true,'2019-09-27 14:34:59.418','2019-09-27 14:34:59.418')
,('TT GDTX H. Châu Thành','COUNTRY_VN',56,628,NULL,true,'2019-09-27 14:34:59.419','2019-09-27 14:34:59.419')
,('THPT Cái Tắc','COUNTRY_VN',56,629,NULL,true,'2019-09-27 14:34:59.419','2019-09-27 14:34:59.419')
,('THPT Châu Thành A','COUNTRY_VN',56,629,NULL,true,'2019-09-27 14:34:59.420','2019-09-27 14:34:59.420')
,('THPT Tâm Vu','COUNTRY_VN',56,629,NULL,true,'2019-09-27 14:34:59.420','2019-09-27 14:34:59.420')
,('THPT Trường Long Tây','COUNTRY_VN',56,629,NULL,true,'2019-09-27 14:34:59.420','2019-09-27 14:34:59.420')
,('TT GDTX H. Châu Thành A','COUNTRY_VN',56,629,NULL,true,'2019-09-27 14:34:59.421','2019-09-27 14:34:59.421')
,('Phổ thông Dân tộc nội trú','COUNTRY_VN',56,630,NULL,true,'2019-09-27 14:34:59.422','2019-09-27 14:34:59.422')
,('THPT Long Mỹ','COUNTRY_VN',56,630,NULL,true,'2019-09-27 14:34:59.422','2019-09-27 14:34:59.422')
,('THPT Lương Tâm','COUNTRY_VN',56,630,NULL,true,'2019-09-27 14:34:59.422','2019-09-27 14:34:59.422')
,('THPT Tân Phú','COUNTRY_VN',56,630,NULL,true,'2019-09-27 14:34:59.423','2019-09-27 14:34:59.423')
,('THPT Tây Đô','COUNTRY_VN',56,630,NULL,true,'2019-09-27 14:34:59.423','2019-09-27 14:34:59.423')
,('TT GDTX H. Long Mỹ','COUNTRY_VN',56,630,NULL,true,'2019-09-27 14:34:59.424','2019-09-27 14:34:59.424')
,('THPT Cây Dương','COUNTRY_VN',56,631,NULL,true,'2019-09-27 14:34:59.426','2019-09-27 14:34:59.426')
,('THPT Hòa An','COUNTRY_VN',56,631,NULL,true,'2019-09-27 14:34:59.427','2019-09-27 14:34:59.427')
,('THPT Lương Thế Vinh','COUNTRY_VN',56,631,NULL,true,'2019-09-27 14:34:59.427','2019-09-27 14:34:59.427')
,('THPT Tân Long','COUNTRY_VN',56,631,NULL,true,'2019-09-27 14:34:59.428','2019-09-27 14:34:59.428')
,('TT GDTX H. Phụng Hiệp','COUNTRY_VN',56,631,NULL,true,'2019-09-27 14:34:59.430','2019-09-27 14:34:59.430')
,('THPT Lê Hồng Phong','COUNTRY_VN',56,632,NULL,true,'2019-09-27 14:34:59.432','2019-09-27 14:34:59.432')
,('THPT Vị Thủy','COUNTRY_VN',56,632,NULL,true,'2019-09-27 14:34:59.433','2019-09-27 14:34:59.433')
,('THPT Vĩnh Tường','COUNTRY_VN',56,632,NULL,true,'2019-09-27 14:34:59.433','2019-09-27 14:34:59.433')
,('TT GDTX H. Vị Thuỷ','COUNTRY_VN',56,632,NULL,true,'2019-09-27 14:34:59.434','2019-09-27 14:34:59.434')
,('TC nghề tỉnh Hậu Giang','COUNTRY_VN',56,633,NULL,true,'2019-09-27 14:34:59.435','2019-09-27 14:34:59.435')
,('THPT Chiêm Thành Tấn','COUNTRY_VN',56,633,NULL,true,'2019-09-27 14:34:59.435','2019-09-27 14:34:59.435')
,('THPT chuyên Vị Thanh','COUNTRY_VN',56,633,NULL,true,'2019-09-27 14:34:59.436','2019-09-27 14:34:59.436')
,('THPT Vị Thanh','COUNTRY_VN',56,633,NULL,true,'2019-09-27 14:34:59.436','2019-09-27 14:34:59.436')
,('TT GDTX thành phố vị Thanh','COUNTRY_VN',56,633,NULL,true,'2019-09-27 14:34:59.437','2019-09-27 14:34:59.437')
,('TC nghề Ngã Bảy','COUNTRY_VN',56,634,NULL,true,'2019-09-27 14:34:59.438','2019-09-27 14:34:59.438')
,('THPT Lê Quý Đôn','COUNTRY_VN',56,634,NULL,true,'2019-09-27 14:34:59.438','2019-09-27 14:34:59.438')
,('THPT Nguyễn Minh Quang','COUNTRY_VN',56,634,NULL,true,'2019-09-27 14:34:59.438','2019-09-27 14:34:59.438')
,('TT GDTX thị xã Ngã Bảy','COUNTRY_VN',56,634,NULL,true,'2019-09-27 14:34:59.439','2019-09-27 14:34:59.439')
,('THPT Cao Phong','COUNTRY_VN',57,635,NULL,true,'2019-09-27 14:34:59.442','2019-09-27 14:34:59.442')
,('THPT Thạch Yên','COUNTRY_VN',57,635,NULL,true,'2019-09-27 14:34:59.443','2019-09-27 14:34:59.443')
,('TT GDTX Cao Phong','COUNTRY_VN',57,635,NULL,true,'2019-09-27 14:34:59.443','2019-09-27 14:34:59.443')
,('THPT Đà Bắc','COUNTRY_VN',57,636,NULL,true,'2019-09-27 14:34:59.444','2019-09-27 14:34:59.444')
,('THPT Mường Chiềng','COUNTRY_VN',57,636,NULL,true,'2019-09-27 14:34:59.445','2019-09-27 14:34:59.445')
,('THPT Yên Hoà','COUNTRY_VN',57,636,NULL,true,'2019-09-27 14:34:59.446','2019-09-27 14:34:59.446')
,('TT GDTX Đà Bắc','COUNTRY_VN',57,636,NULL,true,'2019-09-27 14:34:59.447','2019-09-27 14:34:59.447')
,('CĐ nghề Cơ điện Tây Băc','COUNTRY_VN',57,637,NULL,true,'2019-09-27 14:34:59.449','2019-09-27 14:34:59.449')
,('CĐ nghề Hòa Bình','COUNTRY_VN',57,637,NULL,true,'2019-09-27 14:34:59.449','2019-09-27 14:34:59.449')
,('CĐ nghề Sông Đà','COUNTRY_VN',57,637,NULL,true,'2019-09-27 14:34:59.450','2019-09-27 14:34:59.450')
,('Phổ thông Dân tộc nội trú','COUNTRY_VN',57,637,NULL,true,'2019-09-27 14:34:59.450','2019-09-27 14:34:59.450')
,('THPT chuyên Hoàng Văn Thụ','COUNTRY_VN',57,637,NULL,true,'2019-09-27 14:34:59.450','2019-09-27 14:34:59.450')
,('THPT Công Nghiệp','COUNTRY_VN',57,637,NULL,true,'2019-09-27 14:34:59.451','2019-09-27 14:34:59.451')
,('THPT Lạc Long Quân','COUNTRY_VN',57,637,NULL,true,'2019-09-27 14:34:59.451','2019-09-27 14:34:59.451')
,('THPT Ngô Quyền','COUNTRY_VN',57,637,NULL,true,'2019-09-27 14:34:59.452','2019-09-27 14:34:59.452')
,('THPT Nguyễn Du','COUNTRY_VN',57,637,NULL,true,'2019-09-27 14:34:59.452','2019-09-27 14:34:59.452')
,('Trung học Kinh tế-Kỹ Thuật HB','COUNTRY_VN',57,637,NULL,true,'2019-09-27 14:34:59.452','2019-09-27 14:34:59.452')
,('TT GDTX thành phố HB','COUNTRY_VN',57,637,NULL,true,'2019-09-27 14:34:59.453','2019-09-27 14:34:59.453')
,('TT GDTX tỉnh Hoà Bình','COUNTRY_VN',57,637,NULL,true,'2019-09-27 14:34:59.453','2019-09-27 14:34:59.453')
,('THPT 19/5','COUNTRY_VN',57,638,NULL,true,'2019-09-27 14:34:59.454','2019-09-27 14:34:59.454')
,('THPT Bắc Son','COUNTRY_VN',57,638,NULL,true,'2019-09-27 14:34:59.454','2019-09-27 14:34:59.454')
,('THPT Kim Boi','COUNTRY_VN',57,638,NULL,true,'2019-09-27 14:34:59.455','2019-09-27 14:34:59.455')
,('THPT Sào Báy','COUNTRY_VN',57,638,NULL,true,'2019-09-27 14:34:59.455','2019-09-27 14:34:59.455')
,('TT GDTX Kim Bôi','COUNTRY_VN',57,638,NULL,true,'2019-09-27 14:34:59.455','2019-09-27 14:34:59.455')
,('THPT Kỳ Sơn','COUNTRY_VN',57,639,NULL,true,'2019-09-27 14:34:59.456','2019-09-27 14:34:59.456')
,('THPT Phú Cường','COUNTRY_VN',57,639,NULL,true,'2019-09-27 14:34:59.457','2019-09-27 14:34:59.457')
,('TT GDTX &DN Kỳ Sơn','COUNTRY_VN',57,639,NULL,true,'2019-09-27 14:34:59.458','2019-09-27 14:34:59.458')
,('THPT Cộng Hoà','COUNTRY_VN',57,640,NULL,true,'2019-09-27 14:34:59.459','2019-09-27 14:34:59.459')
,('THPT Đại Đồng','COUNTRY_VN',57,640,NULL,true,'2019-09-27 14:34:59.460','2019-09-27 14:34:59.460')
,('THPT Lạc Sơn','COUNTRY_VN',57,640,NULL,true,'2019-09-27 14:34:59.461','2019-09-27 14:34:59.461')
,('THPT Ngọc Sơn','COUNTRY_VN',57,640,NULL,true,'2019-09-27 14:34:59.465','2019-09-27 14:34:59.465')
,('THPT Quyết Thắng','COUNTRY_VN',57,640,NULL,true,'2019-09-27 14:34:59.466','2019-09-27 14:34:59.466')
,('TT GDTX &DN Lạc Sơn','COUNTRY_VN',57,640,NULL,true,'2019-09-27 14:34:59.467','2019-09-27 14:34:59.467')
,('THPT Lạc Thuỷ A','COUNTRY_VN',57,641,NULL,true,'2019-09-27 14:34:59.468','2019-09-27 14:34:59.468')
,('THPT Lạc Thuỷ B','COUNTRY_VN',57,641,NULL,true,'2019-09-27 14:34:59.469','2019-09-27 14:34:59.469')
,('THPT Lạc Thuỷ C','COUNTRY_VN',57,641,NULL,true,'2019-09-27 14:34:59.470','2019-09-27 14:34:59.470')
,('THPT Thanh Hà','COUNTRY_VN',57,641,NULL,true,'2019-09-27 14:34:59.470','2019-09-27 14:34:59.470')
,('TT GDTX Lạc Thuỷ','COUNTRY_VN',57,641,NULL,true,'2019-09-27 14:34:59.471','2019-09-27 14:34:59.471')
,('THPT CÙ Chính Lan','COUNTRY_VN',57,642,NULL,true,'2019-09-27 14:34:59.472','2019-09-27 14:34:59.472')
,('THPT Lương Sơn','COUNTRY_VN',57,642,NULL,true,'2019-09-27 14:34:59.472','2019-09-27 14:34:59.472')
,('THPT Nam Lương Sơn','COUNTRY_VN',57,642,NULL,true,'2019-09-27 14:34:59.473','2019-09-27 14:34:59.473')
,('THPT Nguyễn Trãi','COUNTRY_VN',57,642,NULL,true,'2019-09-27 14:34:59.473','2019-09-27 14:34:59.473')
,('TT GDTX Lương Sơn','COUNTRY_VN',57,642,NULL,true,'2019-09-27 14:34:59.474','2019-09-27 14:34:59.474')
,('THPT Mai Châu A','COUNTRY_VN',57,643,NULL,true,'2019-09-27 14:34:59.476','2019-09-27 14:34:59.476')
,('THPT Mai Châu B','COUNTRY_VN',57,643,NULL,true,'2019-09-27 14:34:59.477','2019-09-27 14:34:59.477')
,('TT GDTX Mai Châu','COUNTRY_VN',57,643,NULL,true,'2019-09-27 14:34:59.478','2019-09-27 14:34:59.478')
,('THPT Đoàn Kết','COUNTRY_VN',57,644,NULL,true,'2019-09-27 14:34:59.479','2019-09-27 14:34:59.479')
,('THPT Lũng Vân','COUNTRY_VN',57,644,NULL,true,'2019-09-27 14:34:59.481','2019-09-27 14:34:59.481')
,('THPT Mường Bi','COUNTRY_VN',57,644,NULL,true,'2019-09-27 14:34:59.482','2019-09-27 14:34:59.482')
,('THPT Tân Lạc','COUNTRY_VN',57,644,NULL,true,'2019-09-27 14:34:59.482','2019-09-27 14:34:59.482')
,('TT GDTX &DN Tân Lac','COUNTRY_VN',57,644,NULL,true,'2019-09-27 14:34:59.482','2019-09-27 14:34:59.482')
,('THPT Yên Thuỷ A','COUNTRY_VN',57,645,NULL,true,'2019-09-27 14:34:59.483','2019-09-27 14:34:59.483')
,('THPT Yên Thuỷ B','COUNTRY_VN',57,645,NULL,true,'2019-09-27 14:34:59.484','2019-09-27 14:34:59.484')
,('THPT Yên Thuỷ C','COUNTRY_VN',57,645,NULL,true,'2019-09-27 14:34:59.484','2019-09-27 14:34:59.484')
,('TT GDTX Yên Thuỷ','COUNTRY_VN',57,645,NULL,true,'2019-09-27 14:34:59.485','2019-09-27 14:34:59.485')
,('THPT Ân Thi','COUNTRY_VN',58,646,NULL,true,'2019-09-27 14:34:59.486','2019-09-27 14:34:59.486')
,('THPT Lê Quý Đôn','COUNTRY_VN',58,646,NULL,true,'2019-09-27 14:34:59.486','2019-09-27 14:34:59.486')
,('THPT Nguyễn Trung Ngạn','COUNTRY_VN',58,646,NULL,true,'2019-09-27 14:34:59.487','2019-09-27 14:34:59.487')
,('THPT Phạm Ngũ Lão','COUNTRY_VN',58,646,NULL,true,'2019-09-27 14:34:59.487','2019-09-27 14:34:59.487')
,('TT GDTX Ân Thi','COUNTRY_VN',58,646,NULL,true,'2019-09-27 14:34:59.488','2019-09-27 14:34:59.488')
,('TT KT-TH Ân Thi','COUNTRY_VN',58,646,NULL,true,'2019-09-27 14:34:59.488','2019-09-27 14:34:59.488')
,('THPT Nguyễn Trãi','COUNTRY_VN',58,647,NULL,true,'2019-09-27 14:34:59.489','2019-09-27 14:34:59.489')
,('CĐ Nghề Cơ điện và Thủy lợi','COUNTRY_VN',58,648,NULL,true,'2019-09-27 14:34:59.490','2019-09-27 14:34:59.490')
,('THPT Khoái châu','COUNTRY_VN',58,648,NULL,true,'2019-09-27 14:34:59.490','2019-09-27 14:34:59.490')
,('THPT Nam Khoái Châu','COUNTRY_VN',58,648,NULL,true,'2019-09-27 14:34:59.492','2019-09-27 14:34:59.492')
,('THPT Nguyễn Siêu','COUNTRY_VN',58,648,NULL,true,'2019-09-27 14:34:59.493','2019-09-27 14:34:59.493')
,('THPT Phùng Hưng','COUNTRY_VN',58,648,NULL,true,'2019-09-27 14:34:59.494','2019-09-27 14:34:59.494')
,('THPT Trần Quang Khải','COUNTRY_VN',58,648,NULL,true,'2019-09-27 14:34:59.494','2019-09-27 14:34:59.494')
,('TT KT-TH Khoái Châu','COUNTRY_VN',58,648,NULL,true,'2019-09-27 14:34:59.495','2019-09-27 14:34:59.495')
,('TT GDTX Khoái Châu','COUNTRY_VN',58,648,NULL,true,'2019-09-27 14:34:59.497','2019-09-27 14:34:59.497')
,('THPT Đúc Hợp','COUNTRY_VN',58,649,NULL,true,'2019-09-27 14:34:59.499','2019-09-27 14:34:59.499')
,('THPT Kim Động','COUNTRY_VN',58,649,NULL,true,'2019-09-27 14:34:59.500','2019-09-27 14:34:59.500')
,('THPT Nghĩa Dân','COUNTRY_VN',58,649,NULL,true,'2019-09-27 14:34:59.500','2019-09-27 14:34:59.500')
,('TT GDTX Kim Động','COUNTRY_VN',58,649,NULL,true,'2019-09-27 14:34:59.501','2019-09-27 14:34:59.501')
,('THPT Hồng Đức','COUNTRY_VN',58,650,NULL,true,'2019-09-27 14:34:59.502','2019-09-27 14:34:59.502')
,('THPT Mỹ Hào','COUNTRY_VN',58,650,NULL,true,'2019-09-27 14:34:59.502','2019-09-27 14:34:59.502')
,('THPT Nguyễn Thiện Thuật','COUNTRY_VN',58,650,NULL,true,'2019-09-27 14:34:59.503','2019-09-27 14:34:59.503')
,('TT GDTX Mỹ Hào','COUNTRY_VN',58,650,NULL,true,'2019-09-27 14:34:59.503','2019-09-27 14:34:59.503')
,('THPT Nam Phù Cừ','COUNTRY_VN',58,651,NULL,true,'2019-09-27 14:34:59.504','2019-09-27 14:34:59.504')
,('THPT Nguyễn Du','COUNTRY_VN',58,651,NULL,true,'2019-09-27 14:34:59.504','2019-09-27 14:34:59.504')
,('THPT Phù Cừ','COUNTRY_VN',58,651,NULL,true,'2019-09-27 14:34:59.505','2019-09-27 14:34:59.505')
,('TT GDTX Phù cừ','COUNTRY_VN',58,651,NULL,true,'2019-09-27 14:34:59.505','2019-09-27 14:34:59.505')
,('THPT Hoàng Hoa Thám','COUNTRY_VN',58,652,NULL,true,'2019-09-27 14:34:59.506','2019-09-27 14:34:59.506')
,('THPT Ngô Quyền','COUNTRY_VN',58,652,NULL,true,'2019-09-27 14:34:59.506','2019-09-27 14:34:59.506')
,('THPT Tiên Lữ','COUNTRY_VN',58,652,NULL,true,'2019-09-27 14:34:59.507','2019-09-27 14:34:59.507')
,('THPT Trần Hung Đạo','COUNTRY_VN',58,652,NULL,true,'2019-09-27 14:34:59.511','2019-09-27 14:34:59.511')
,('TT GDTX Tiên Lữ','COUNTRY_VN',58,652,NULL,true,'2019-09-27 14:34:59.512','2019-09-27 14:34:59.512')
,('TT-KT-TH Tiên Lữ','COUNTRY_VN',58,652,NULL,true,'2019-09-27 14:34:59.512','2019-09-27 14:34:59.512')
,('PT Đoàn thị Điểm Ecopark','COUNTRY_VN',58,653,NULL,true,'2019-09-27 14:34:59.515','2019-09-27 14:34:59.515')
,('THPT Dương Quảng Hàm','COUNTRY_VN',58,653,NULL,true,'2019-09-27 14:34:59.516','2019-09-27 14:34:59.516')
,('THPT Nguyễn Công Hoan','COUNTRY_VN',58,653,NULL,true,'2019-09-27 14:34:59.517','2019-09-27 14:34:59.517')
,('THPT Văn Giang','COUNTRY_VN',58,653,NULL,true,'2019-09-27 14:34:59.518','2019-09-27 14:34:59.518')
,('TT GDTX Văn Giang','COUNTRY_VN',58,653,NULL,true,'2019-09-27 14:34:59.519','2019-09-27 14:34:59.519')
,('THPT Hùng Vương','COUNTRY_VN',58,654,NULL,true,'2019-09-27 14:34:59.520','2019-09-27 14:34:59.520')
,('THPT Lương Tài','COUNTRY_VN',58,654,NULL,true,'2019-09-27 14:34:59.520','2019-09-27 14:34:59.520')
,('THPT Trung Vương','COUNTRY_VN',58,654,NULL,true,'2019-09-27 14:34:59.521','2019-09-27 14:34:59.521')
,('THPT Văn Lâm','COUNTRY_VN',58,654,NULL,true,'2019-09-27 14:34:59.521','2019-09-27 14:34:59.521')
,('TT GDTX Văn Lâm','COUNTRY_VN',58,654,NULL,true,'2019-09-27 14:34:59.522','2019-09-27 14:34:59.522')
,('CĐ Công Nghiệp Hưng Yên','COUNTRY_VN',58,655,NULL,true,'2019-09-27 14:34:59.522','2019-09-27 14:34:59.522')
,('THPT Hồng Bàng','COUNTRY_VN',58,655,NULL,true,'2019-09-27 14:34:59.523','2019-09-27 14:34:59.523')
,('THPT Minh Châu','COUNTRY_VN',58,655,NULL,true,'2019-09-27 14:34:59.524','2019-09-27 14:34:59.524')
,('THPT Triệu Quang Phục','COUNTRY_VN',58,655,NULL,true,'2019-09-27 14:34:59.527','2019-09-27 14:34:59.527')
,('THPT Yên Mỹ','COUNTRY_VN',58,655,NULL,true,'2019-09-27 14:34:59.529','2019-09-27 14:34:59.529')
,('TT GDTX Phố Nối','COUNTRY_VN',58,655,NULL,true,'2019-09-27 14:34:59.531','2019-09-27 14:34:59.531')
,('TC Nghề Hung Yên','COUNTRY_VN',58,656,NULL,true,'2019-09-27 14:34:59.533','2019-09-27 14:34:59.533')
,('TC Văn hóa Ng.Thuật và D. Lịch HY','COUNTRY_VN',58,656,NULL,true,'2019-09-27 14:34:59.533','2019-09-27 14:34:59.533')
,('THPT Chuyên tỉnh Hưng Yên','COUNTRY_VN',58,656,NULL,true,'2019-09-27 14:34:59.534','2019-09-27 14:34:59.534')
,('THPT Quang Trung','COUNTRY_VN',58,656,NULL,true,'2019-09-27 14:34:59.535','2019-09-27 14:34:59.535')
,('THPT Tô Hiệu','COUNTRY_VN',58,656,NULL,true,'2019-09-27 14:34:59.536','2019-09-27 14:34:59.536')
,('THPT TP Hung Yên','COUNTRY_VN',58,656,NULL,true,'2019-09-27 14:34:59.536','2019-09-27 14:34:59.536')
,('TT GDTX TP. Hưng Yên','COUNTRY_VN',58,656,NULL,true,'2019-09-27 14:34:59.537','2019-09-27 14:34:59.537')
,('THPT Nguyễn Huệ','COUNTRY_VN',59,657,NULL,true,'2019-09-27 14:34:59.539','2019-09-27 14:34:59.539')
,('THPT Trần Bình Trọng','COUNTRY_VN',59,657,NULL,true,'2019-09-27 14:34:59.539','2019-09-27 14:34:59.539')
,('TT GDTX Cam Lâm','COUNTRY_VN',59,657,NULL,true,'2019-09-27 14:34:59.540','2019-09-27 14:34:59.540')
,('THPT BC Lê Lợi','COUNTRY_VN',59,658,NULL,true,'2019-09-27 14:34:59.544','2019-09-27 14:34:59.544')
,('THPT BC Nguyễn Bỉnh Khiêm','COUNTRY_VN',59,658,NULL,true,'2019-09-27 14:34:59.545','2019-09-27 14:34:59.545')
,('THPT Đoàn Thị Điểm','COUNTRY_VN',59,658,NULL,true,'2019-09-27 14:34:59.546','2019-09-27 14:34:59.546')
,('THPT Hoàng Hoa Thám','COUNTRY_VN',59,658,NULL,true,'2019-09-27 14:34:59.547','2019-09-27 14:34:59.547')
,('THPT Nguyễn Thái Học','COUNTRY_VN',59,658,NULL,true,'2019-09-27 14:34:59.549','2019-09-27 14:34:59.549')
,('TT GDTX Diên Khánh','COUNTRY_VN',59,658,NULL,true,'2019-09-27 14:34:59.550','2019-09-27 14:34:59.550')
,('Cấp2,3 Khánh Sơn','COUNTRY_VN',59,659,NULL,true,'2019-09-27 14:34:59.551','2019-09-27 14:34:59.551')
,('TT GDTX Khánh Sơn','COUNTRY_VN',59,659,NULL,true,'2019-09-27 14:34:59.552','2019-09-27 14:34:59.552')
,('THPT Lạc Long Quân','COUNTRY_VN',59,660,NULL,true,'2019-09-27 14:34:59.553','2019-09-27 14:34:59.553')
,('TT GDTX Khánh Vĩnh','COUNTRY_VN',59,660,NULL,true,'2019-09-27 14:34:59.553','2019-09-27 14:34:59.553')
,('TC Nghề Vạn Ninh','COUNTRY_VN',59,661,NULL,true,'2019-09-27 14:34:59.554','2019-09-27 14:34:59.554')
,('THPT Huỳnh Thúc Kháng','COUNTRY_VN',59,661,NULL,true,'2019-09-27 14:34:59.554','2019-09-27 14:34:59.554')
,('THPT Lê Hồng Phong','COUNTRY_VN',59,661,NULL,true,'2019-09-27 14:34:59.555','2019-09-27 14:34:59.555')
,('THPT Nguyễn Thị Minh Khai','COUNTRY_VN',59,661,NULL,true,'2019-09-27 14:34:59.555','2019-09-27 14:34:59.555')
,('THPT Tô Văn Ơn','COUNTRY_VN',59,661,NULL,true,'2019-09-27 14:34:59.556','2019-09-27 14:34:59.556')
,('TT GDTX Vạn Ninh','COUNTRY_VN',59,661,NULL,true,'2019-09-27 14:34:59.556','2019-09-27 14:34:59.556')
,('Hệ GDTX tại THPT Ngô Gia Tự','COUNTRY_VN',59,662,NULL,true,'2019-09-27 14:34:59.559','2019-09-27 14:34:59.559')
,('TC nghề Cam Ranh','COUNTRY_VN',59,662,NULL,true,'2019-09-27 14:34:59.560','2019-09-27 14:34:59.560')
,('THPT Ngô Gia Tự','COUNTRY_VN',59,662,NULL,true,'2019-09-27 14:34:59.561','2019-09-27 14:34:59.561')
,('THPT Phan Bội Châu','COUNTRY_VN',59,662,NULL,true,'2019-09-27 14:34:59.562','2019-09-27 14:34:59.562')
,('THPT Thăng Long','COUNTRY_VN',59,662,NULL,true,'2019-09-27 14:34:59.564','2019-09-27 14:34:59.564')
,('THPT Trần Hưng Đạo','COUNTRY_VN',59,662,NULL,true,'2019-09-27 14:34:59.565','2019-09-27 14:34:59.565')
,('TT GDTX Cam Ranh','COUNTRY_VN',59,662,NULL,true,'2019-09-27 14:34:59.566','2019-09-27 14:34:59.566')
,('APC Nha Trang','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.566','2019-09-27 14:34:59.566')
,('BTTH Nha Trang 2','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.567','2019-09-27 14:34:59.567')
,('CĐ nghề Nha Trang','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.567','2019-09-27 14:34:59.567')
,('CĐ nghề Quốc tế Nam Việt','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.568','2019-09-27 14:34:59.568')
,('dự bị ĐH Dân tộc TW Nha Trang','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.568','2019-09-27 14:34:59.568')
,('PT Dân tộc Nội trú tỉnh KH','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.569','2019-09-27 14:34:59.569')
,('Quốc Tế Hoàn cầu Nha Trang','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.569','2019-09-27 14:34:59.569')
,('TC Kinh tế Khánh Hòa','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.570','2019-09-27 14:34:59.570')
,('TC KTKT Trần Đại Nghĩa','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.570','2019-09-27 14:34:59.570')
,('TC nghề Nha Trang','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.570','2019-09-27 14:34:59.570')
,('THCS & THPT iSchool Nha Trang','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.571','2019-09-27 14:34:59.571')
,('THPT Đạmri -Đạ Huoai','COUNTRY_VN',63,701,NULL,true,'2019-09-27 14:34:59.702','2019-09-27 14:34:59.702')
,('THPT BC Nguyễn Trường Tộ','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.571','2019-09-27 14:34:59.571')
,('THPT chuyên Lê Quý Đôn','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.572','2019-09-27 14:34:59.572')
,('THPT DL Lê Thánh Tôn','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.572','2019-09-27 14:34:59.572')
,('THPT DL Nguyễn Thiện Thuật','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.573','2019-09-27 14:34:59.573')
,('THPT Đại Việt','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.573','2019-09-27 14:34:59.573')
,('THPT Hà Huy Tập','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.574','2019-09-27 14:34:59.574')
,('THPT Hermann Gmeiner','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.575','2019-09-27 14:34:59.575')
,('THPT Hoàng Văn Thụ','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.575','2019-09-27 14:34:59.575')
,('THPT Lý Tự Trọng','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.576','2019-09-27 14:34:59.576')
,('THPT Nguyễn Văn Trỗi','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.576','2019-09-27 14:34:59.576')
,('TT GDTX Nha Trang','COUNTRY_VN',59,663,NULL,true,'2019-09-27 14:34:59.577','2019-09-27 14:34:59.577')
,('THPT An Biên','COUNTRY_VN',60,664,NULL,true,'2019-09-27 14:34:59.579','2019-09-27 14:34:59.579')
,('THPT Đông Thái','COUNTRY_VN',60,664,NULL,true,'2019-09-27 14:34:59.581','2019-09-27 14:34:59.581')
,('THPT Nam Yên','COUNTRY_VN',60,664,NULL,true,'2019-09-27 14:34:59.582','2019-09-27 14:34:59.582')
,('Trung tâm GDTX An Biên','COUNTRY_VN',60,664,NULL,true,'2019-09-27 14:34:59.582','2019-09-27 14:34:59.582')
,('THPT An Minh','COUNTRY_VN',60,665,NULL,true,'2019-09-27 14:34:59.583','2019-09-27 14:34:59.583')
,('THPT Nguyễn Văn Xiên','COUNTRY_VN',60,665,NULL,true,'2019-09-27 14:34:59.583','2019-09-27 14:34:59.583')
,('THPT Vân Khánh','COUNTRY_VN',60,665,NULL,true,'2019-09-27 14:34:59.584','2019-09-27 14:34:59.584')
,('Trung tâm GDTX An Minh','COUNTRY_VN',60,665,NULL,true,'2019-09-27 14:34:59.584','2019-09-27 14:34:59.584')
,('THPT Châu Thành','COUNTRY_VN',60,666,NULL,true,'2019-09-27 14:34:59.585','2019-09-27 14:34:59.585')
,('THPT Mong Thọ','COUNTRY_VN',60,666,NULL,true,'2019-09-27 14:34:59.585','2019-09-27 14:34:59.585')
,('Trung tâm GDTX Châu Thành','COUNTRY_VN',60,666,NULL,true,'2019-09-27 14:34:59.585','2019-09-27 14:34:59.585')
,('THCS An Sơn','COUNTRY_VN',60,667,NULL,true,'2019-09-27 14:34:59.586','2019-09-27 14:34:59.586')
,('THPT Kiên Hải','COUNTRY_VN',60,667,NULL,true,'2019-09-27 14:34:59.587','2019-09-27 14:34:59.587')
,('THPT Lại Sơn','COUNTRY_VN',60,667,NULL,true,'2019-09-27 14:34:59.587','2019-09-27 14:34:59.587')
,('THPT An Thới','COUNTRY_VN',60,668,NULL,true,'2019-09-27 14:34:59.588','2019-09-27 14:34:59.588')
,('THPT Dương Đông','COUNTRY_VN',60,668,NULL,true,'2019-09-27 14:34:59.588','2019-09-27 14:34:59.588')
,('THPT Phú Quốc','COUNTRY_VN',60,668,NULL,true,'2019-09-27 14:34:59.588','2019-09-27 14:34:59.588')
,('Trung tâm GDTX Phú Quốc','COUNTRY_VN',60,668,NULL,true,'2019-09-27 14:34:59.589','2019-09-27 14:34:59.589')
,('THPT Thoại Ngọc Hầu','COUNTRY_VN',60,669,NULL,true,'2019-09-27 14:34:59.589','2019-09-27 14:34:59.589')
,('Trung tâm GDTX Giang Thành','COUNTRY_VN',60,669,NULL,true,'2019-09-27 14:34:59.590','2019-09-27 14:34:59.590')
,('THCS Thạnh Phước','COUNTRY_VN',60,670,NULL,true,'2019-09-27 14:34:59.592','2019-09-27 14:34:59.592')
,('THPT Bàn Tân Định','COUNTRY_VN',60,670,NULL,true,'2019-09-27 14:34:59.593','2019-09-27 14:34:59.593')
,('THPT Giồng Riềng','COUNTRY_VN',60,670,NULL,true,'2019-09-27 14:34:59.594','2019-09-27 14:34:59.594')
,('THPT Hoà Hưng','COUNTRY_VN',60,670,NULL,true,'2019-09-27 14:34:59.594','2019-09-27 14:34:59.594')
,('THPT Hòa Thuận','COUNTRY_VN',60,670,NULL,true,'2019-09-27 14:34:59.597','2019-09-27 14:34:59.597')
,('THPT Long Thạnh','COUNTRY_VN',60,670,NULL,true,'2019-09-27 14:34:59.598','2019-09-27 14:34:59.598')
,('THPT Thạnh Lộc','COUNTRY_VN',60,670,NULL,true,'2019-09-27 14:34:59.599','2019-09-27 14:34:59.599')
,('Trung cấp Nghề DTNT tỉnh Kiên Giang','COUNTRY_VN',60,670,NULL,true,'2019-09-27 14:34:59.600','2019-09-27 14:34:59.600')
,('Trung tâm GDTX Giồng Riềng','COUNTRY_VN',60,670,NULL,true,'2019-09-27 14:34:59.601','2019-09-27 14:34:59.601')
,('THPT Định An','COUNTRY_VN',60,671,NULL,true,'2019-09-27 14:34:59.602','2019-09-27 14:34:59.602')
,('THPT GÒ Quao','COUNTRY_VN',60,671,NULL,true,'2019-09-27 14:34:59.602','2019-09-27 14:34:59.602')
,('THPT Thới Quản','COUNTRY_VN',60,671,NULL,true,'2019-09-27 14:34:59.603','2019-09-27 14:34:59.603')
,('THPT Vĩnh Hoà Hưng Bắc','COUNTRY_VN',60,671,NULL,true,'2019-09-27 14:34:59.603','2019-09-27 14:34:59.603')
,('THPT Vĩnh Thẳng','COUNTRY_VN',60,671,NULL,true,'2019-09-27 14:34:59.604','2019-09-27 14:34:59.604')
,('Trung tâm GDTX Gò Quao','COUNTRY_VN',60,671,NULL,true,'2019-09-27 14:34:59.604','2019-09-27 14:34:59.604')
,('THPT Bình Sơn','COUNTRY_VN',60,672,NULL,true,'2019-09-27 14:34:59.605','2019-09-27 14:34:59.605')
,('THPT Hòn Đất','COUNTRY_VN',60,672,NULL,true,'2019-09-27 14:34:59.606','2019-09-27 14:34:59.606')
,('THPT Nam Thái Sơn','COUNTRY_VN',60,672,NULL,true,'2019-09-27 14:34:59.606','2019-09-27 14:34:59.606')
,('THPT Nguyễn Hùng Hiệp','COUNTRY_VN',60,672,NULL,true,'2019-09-27 14:34:59.607','2019-09-27 14:34:59.607')
,('THPT Phan Thị Ràng','COUNTRY_VN',60,672,NULL,true,'2019-09-27 14:34:59.608','2019-09-27 14:34:59.608')
,('THPT Sóc Sơn','COUNTRY_VN',60,672,NULL,true,'2019-09-27 14:34:59.611','2019-09-27 14:34:59.611')
,('Trung tâm GDTX Hòn Đất','COUNTRY_VN',60,672,NULL,true,'2019-09-27 14:34:59.612','2019-09-27 14:34:59.612')
,('THCS An Sơn','COUNTRY_VN',60,673,NULL,true,'2019-09-27 14:34:59.615','2019-09-27 14:34:59.615')
,('THPT Kiên Hải','COUNTRY_VN',60,673,NULL,true,'2019-09-27 14:34:59.616','2019-09-27 14:34:59.616')
,('THPT Lại Sơn','COUNTRY_VN',60,673,NULL,true,'2019-09-27 14:34:59.617','2019-09-27 14:34:59.617')
,('THPT Ba Hòn','COUNTRY_VN',60,674,NULL,true,'2019-09-27 14:34:59.617','2019-09-27 14:34:59.617')
,('THPT Kiên Lương','COUNTRY_VN',60,674,NULL,true,'2019-09-27 14:34:59.618','2019-09-27 14:34:59.618')
,('Trung tâm GDTX Kiên Lương','COUNTRY_VN',60,674,NULL,true,'2019-09-27 14:34:59.618','2019-09-27 14:34:59.618')
,('THPT An Thới','COUNTRY_VN',60,675,NULL,true,'2019-09-27 14:34:59.619','2019-09-27 14:34:59.619')
,('THPT Dương Đông','COUNTRY_VN',60,675,NULL,true,'2019-09-27 14:34:59.620','2019-09-27 14:34:59.620')
,('THPT Phú Quốc','COUNTRY_VN',60,675,NULL,true,'2019-09-27 14:34:59.620','2019-09-27 14:34:59.620')
,('Trung tâm GDTX Phú Quốc','COUNTRY_VN',60,675,NULL,true,'2019-09-27 14:34:59.621','2019-09-27 14:34:59.621')
,('Cao đẳng Nghề tỉnh Kiên Giang','COUNTRY_VN',60,676,NULL,true,'2019-09-27 14:34:59.622','2019-09-27 14:34:59.622')
,('PT Dân tộc Nội trú Tỉnh','COUNTRY_VN',60,676,NULL,true,'2019-09-27 14:34:59.622','2019-09-27 14:34:59.622')
,('THPT chuyên Huỳnh Mẫn Đạt','COUNTRY_VN',60,676,NULL,true,'2019-09-27 14:34:59.622','2019-09-27 14:34:59.622')
,('THPT iSCHOOL Rạch Giá','COUNTRY_VN',60,676,NULL,true,'2019-09-27 14:34:59.623','2019-09-27 14:34:59.623')
,('THPT Ngô Sĩ Liên','COUNTRY_VN',60,676,NULL,true,'2019-09-27 14:34:59.624','2019-09-27 14:34:59.624')
,('THPT Nguyễn Hùng Sơn','COUNTRY_VN',60,676,NULL,true,'2019-09-27 14:34:59.626','2019-09-27 14:34:59.626')
,('THPT Nguyễn Trung Trục','COUNTRY_VN',60,676,NULL,true,'2019-09-27 14:34:59.626','2019-09-27 14:34:59.626')
,('THPT Phó Cơ Điều','COUNTRY_VN',60,676,NULL,true,'2019-09-27 14:34:59.627','2019-09-27 14:34:59.627')
,('Trung cấp Kỹ thuật-Nghiệp vụ Kiên Giang','COUNTRY_VN',60,676,NULL,true,'2019-09-27 14:34:59.627','2019-09-27 14:34:59.627')
,('Trung tâm GDTX Tỉnh','COUNTRY_VN',60,676,NULL,true,'2019-09-27 14:34:59.628','2019-09-27 14:34:59.628')
,('THPT Cây Dương','COUNTRY_VN',60,677,NULL,true,'2019-09-27 14:34:59.631','2019-09-27 14:34:59.631')
,('THPT Tân Hiệp','COUNTRY_VN',60,677,NULL,true,'2019-09-27 14:34:59.632','2019-09-27 14:34:59.632')
,('THPT Thạnh Đông','COUNTRY_VN',60,677,NULL,true,'2019-09-27 14:34:59.632','2019-09-27 14:34:59.632')
,('THPT Thạnh Tây','COUNTRY_VN',60,677,NULL,true,'2019-09-27 14:34:59.633','2019-09-27 14:34:59.633')
,('Trung tâm GDTX Tân Hiệp','COUNTRY_VN',60,677,NULL,true,'2019-09-27 14:34:59.633','2019-09-27 14:34:59.633')
,('THPT Minh Thuận','COUNTRY_VN',60,678,NULL,true,'2019-09-27 14:34:59.634','2019-09-27 14:34:59.634')
,('THPT U Minh Thượng','COUNTRY_VN',60,678,NULL,true,'2019-09-27 14:34:59.634','2019-09-27 14:34:59.634')
,('THPT Vĩnh Hoà','COUNTRY_VN',60,678,NULL,true,'2019-09-27 14:34:59.635','2019-09-27 14:34:59.635')
,('THPT Vĩnh Bình Bắc','COUNTRY_VN',60,679,NULL,true,'2019-09-27 14:34:59.636','2019-09-27 14:34:59.636')
,('THPT Vĩnh Phong','COUNTRY_VN',60,679,NULL,true,'2019-09-27 14:34:59.636','2019-09-27 14:34:59.636')
,('THPT Vĩnh Thuận','COUNTRY_VN',60,679,NULL,true,'2019-09-27 14:34:59.636','2019-09-27 14:34:59.636')
,('Trung tâm GDTX Vĩnh Thuận','COUNTRY_VN',60,679,NULL,true,'2019-09-27 14:34:59.637','2019-09-27 14:34:59.637')
,('THPT Nguyễn Thần Hiến','COUNTRY_VN',60,680,NULL,true,'2019-09-27 14:34:59.637','2019-09-27 14:34:59.637')
,('Trung tâm GDTX TX Hà Tiên','COUNTRY_VN',60,680,NULL,true,'2019-09-27 14:34:59.638','2019-09-27 14:34:59.638')
,('PT DTNT Đăk Glei','COUNTRY_VN',61,681,NULL,true,'2019-09-27 14:34:59.639','2019-09-27 14:34:59.639')
,('THPT Lương Thế Vinh','COUNTRY_VN',61,681,NULL,true,'2019-09-27 14:34:59.639','2019-09-27 14:34:59.639')
,('TT GDTX Đăk Glei','COUNTRY_VN',61,681,NULL,true,'2019-09-27 14:34:59.639','2019-09-27 14:34:59.639')
,('PT DTNT Đăk Hà','COUNTRY_VN',61,682,NULL,true,'2019-09-27 14:34:59.640','2019-09-27 14:34:59.640')
,('THPT Nguyễn Du','COUNTRY_VN',61,682,NULL,true,'2019-09-27 14:34:59.642','2019-09-27 14:34:59.642')
,('THPT Trần Quốc Tuấn','COUNTRY_VN',61,682,NULL,true,'2019-09-27 14:34:59.642','2019-09-27 14:34:59.642')
,('TT GDTX Đăk Hà','COUNTRY_VN',61,682,NULL,true,'2019-09-27 14:34:59.642','2019-09-27 14:34:59.642')
,('PT DTNT Đăk Tô','COUNTRY_VN',61,683,NULL,true,'2019-09-27 14:34:59.643','2019-09-27 14:34:59.643')
,('THPT Nguyễn Văn Cừ','COUNTRY_VN',61,683,NULL,true,'2019-09-27 14:34:59.644','2019-09-27 14:34:59.644')
,('TT GDTX Đăk Tô','COUNTRY_VN',61,683,NULL,true,'2019-09-27 14:34:59.644','2019-09-27 14:34:59.644')
,('PT DTNT Kon Plong','COUNTRY_VN',61,684,NULL,true,'2019-09-27 14:34:59.646','2019-09-27 14:34:59.646')
,('PT DTNT Kon Rẫy','COUNTRY_VN',61,685,NULL,true,'2019-09-27 14:34:59.649','2019-09-27 14:34:59.649')
,('THPT Chu Văn An','COUNTRY_VN',61,685,NULL,true,'2019-09-27 14:34:59.649','2019-09-27 14:34:59.649')
,('TT GDTX Kon Ray','COUNTRY_VN',61,685,NULL,true,'2019-09-27 14:34:59.650','2019-09-27 14:34:59.650')
,('PT DTNT Ngọc Hồi','COUNTRY_VN',61,686,NULL,true,'2019-09-27 14:34:59.650','2019-09-27 14:34:59.650')
,('THPT Nguyễn Trãi','COUNTRY_VN',61,686,NULL,true,'2019-09-27 14:34:59.651','2019-09-27 14:34:59.651')
,('THPT Phan Chu Trinh','COUNTRY_VN',61,686,NULL,true,'2019-09-27 14:34:59.651','2019-09-27 14:34:59.651')
,('TT GDTX Ngọc Hồi','COUNTRY_VN',61,686,NULL,true,'2019-09-27 14:34:59.652','2019-09-27 14:34:59.652')
,('PTDTNT Sa Thầy','COUNTRY_VN',61,687,NULL,true,'2019-09-27 14:34:59.653','2019-09-27 14:34:59.653')
,('THPT Quang Trung','COUNTRY_VN',61,687,NULL,true,'2019-09-27 14:34:59.653','2019-09-27 14:34:59.653')
,('TT GDTX Sa Thầy','COUNTRY_VN',61,687,NULL,true,'2019-09-27 14:34:59.654','2019-09-27 14:34:59.654')
,('PT DTNT Tu Mơ Rông','COUNTRY_VN',61,688,NULL,true,'2019-09-27 14:34:59.654','2019-09-27 14:34:59.654')
,('CĐ Kinh tế- Kỹ thuật Kon Tum','COUNTRY_VN',61,689,NULL,true,'2019-09-27 14:34:59.655','2019-09-27 14:34:59.655')
,('CĐ Sư phạm Kon Tum','COUNTRY_VN',61,689,NULL,true,'2019-09-27 14:34:59.655','2019-09-27 14:34:59.655')
,('PT DTNT tỉnh Kon Tum','COUNTRY_VN',61,689,NULL,true,'2019-09-27 14:34:59.655','2019-09-27 14:34:59.655')
,('TC Nghề Kon Tum','COUNTRY_VN',61,689,NULL,true,'2019-09-27 14:34:59.656','2019-09-27 14:34:59.656')
,('THPT chuyên Nguyễn Tất Thành','COUNTRY_VN',61,689,NULL,true,'2019-09-27 14:34:59.656','2019-09-27 14:34:59.656')
,('THPT Duy Tân','COUNTRY_VN',61,689,NULL,true,'2019-09-27 14:34:59.656','2019-09-27 14:34:59.656')
,('THPT Kon Tum','COUNTRY_VN',61,689,NULL,true,'2019-09-27 14:34:59.657','2019-09-27 14:34:59.657')
,('THPT Lê Lợi','COUNTRY_VN',61,689,NULL,true,'2019-09-27 14:34:59.658','2019-09-27 14:34:59.658')
,('THPT Ngô Mây','COUNTRY_VN',61,689,NULL,true,'2019-09-27 14:34:59.659','2019-09-27 14:34:59.659')
,('THPT Phan Bội Châu','COUNTRY_VN',61,689,NULL,true,'2019-09-27 14:34:59.660','2019-09-27 14:34:59.660')
,('THPT Trường Chinh','COUNTRY_VN',61,689,NULL,true,'2019-09-27 14:34:59.660','2019-09-27 14:34:59.660')
,('Trung học Y tế Kon Tum','COUNTRY_VN',61,689,NULL,true,'2019-09-27 14:34:59.661','2019-09-27 14:34:59.661')
,('TT GDTX Tỉnh','COUNTRY_VN',61,689,NULL,true,'2019-09-27 14:34:59.661','2019-09-27 14:34:59.661')
,('THPT Dân tộc Nội trú Ka Lăng','COUNTRY_VN',62,690,NULL,true,'2019-09-27 14:34:59.666','2019-09-27 14:34:59.666')
,('THPT Mường Tè','COUNTRY_VN',62,690,NULL,true,'2019-09-27 14:34:59.666','2019-09-27 14:34:59.666')
,('Trung tâm GDTX Mường Tè','COUNTRY_VN',62,690,NULL,true,'2019-09-27 14:34:59.667','2019-09-27 14:34:59.667')
,('THPT Nậm Nhùn','COUNTRY_VN',62,691,NULL,true,'2019-09-27 14:34:59.668','2019-09-27 14:34:59.668')
,('THPT Dào San','COUNTRY_VN',62,692,NULL,true,'2019-09-27 14:34:59.669','2019-09-27 14:34:59.669')
,('THPT Phong Thổ','COUNTRY_VN',62,692,NULL,true,'2019-09-27 14:34:59.669','2019-09-27 14:34:59.669')
,('Trung tâm GDTX huyện Phong Thổ','COUNTRY_VN',62,692,NULL,true,'2019-09-27 14:34:59.670','2019-09-27 14:34:59.670')
,('PTDTNT huyện Sìn Hồ','COUNTRY_VN',62,693,NULL,true,'2019-09-27 14:34:59.671','2019-09-27 14:34:59.671')
,('THPT Nậm Tăm','COUNTRY_VN',62,693,NULL,true,'2019-09-27 14:34:59.671','2019-09-27 14:34:59.671')
,('THPT Sìn Hồ','COUNTRY_VN',62,693,NULL,true,'2019-09-27 14:34:59.672','2019-09-27 14:34:59.672')
,('Trung tâm GDTX huyện Sìn Hồ','COUNTRY_VN',62,693,NULL,true,'2019-09-27 14:34:59.672','2019-09-27 14:34:59.672')
,('THPT Bình Lư','COUNTRY_VN',62,694,NULL,true,'2019-09-27 14:34:59.673','2019-09-27 14:34:59.673')
,('Trung tâm GDTX huyện Tam Đường','COUNTRY_VN',62,694,NULL,true,'2019-09-27 14:34:59.674','2019-09-27 14:34:59.674')
,('THPT Tân uyên','COUNTRY_VN',62,695,NULL,true,'2019-09-27 14:34:59.675','2019-09-27 14:34:59.675')
,('THPT Trung Đồng','COUNTRY_VN',62,695,NULL,true,'2019-09-27 14:34:59.677','2019-09-27 14:34:59.677')
,('Trung tâm GDTX huyện Tân Uyên','COUNTRY_VN',62,695,NULL,true,'2019-09-27 14:34:59.677','2019-09-27 14:34:59.677')
,('PTDTNT huyện Than Uyên','COUNTRY_VN',62,696,NULL,true,'2019-09-27 14:34:59.678','2019-09-27 14:34:59.678')
,('THPT Mường Kim','COUNTRY_VN',62,696,NULL,true,'2019-09-27 14:34:59.680','2019-09-27 14:34:59.680')
,('THPT Mường Than','COUNTRY_VN',62,696,NULL,true,'2019-09-27 14:34:59.681','2019-09-27 14:34:59.681')
,('THPT Than Uyên','COUNTRY_VN',62,696,NULL,true,'2019-09-27 14:34:59.681','2019-09-27 14:34:59.681')
,('Trung tâm GDTX huyện Than Uyên','COUNTRY_VN',62,696,NULL,true,'2019-09-27 14:34:59.682','2019-09-27 14:34:59.682')
,('THPT Chuyên Lê Quý Đôn','COUNTRY_VN',62,697,NULL,true,'2019-09-27 14:34:59.684','2019-09-27 14:34:59.684')
,('THPT Dân tộc Nội trú Tỉnh','COUNTRY_VN',62,697,NULL,true,'2019-09-27 14:34:59.685','2019-09-27 14:34:59.685')
,('THPT Mường So','COUNTRY_VN',62,697,NULL,true,'2019-09-27 14:34:59.685','2019-09-27 14:34:59.685')
,('THPT Quyết Thắng','COUNTRY_VN',62,697,NULL,true,'2019-09-27 14:34:59.685','2019-09-27 14:34:59.685')
,('THPT Thành Phố','COUNTRY_VN',62,697,NULL,true,'2019-09-27 14:34:59.686','2019-09-27 14:34:59.686')
,('Trung cấp nghề Lai Châu','COUNTRY_VN',62,697,NULL,true,'2019-09-27 14:34:59.686','2019-09-27 14:34:59.686')
,('Trung tâm GDTX - Hướng nghiệp Tỉnh','COUNTRY_VN',62,697,NULL,true,'2019-09-27 14:34:59.687','2019-09-27 14:34:59.687')
,('THPT Bảo Lâm','COUNTRY_VN',63,698,NULL,true,'2019-09-27 14:34:59.688','2019-09-27 14:34:59.688')
,('THPT LỘC An -Bảo Lâm','COUNTRY_VN',63,698,NULL,true,'2019-09-27 14:34:59.688','2019-09-27 14:34:59.688')
,('THPT Lộc Băc Bảo Lâm','COUNTRY_VN',63,698,NULL,true,'2019-09-27 14:34:59.689','2019-09-27 14:34:59.689')
,('THPT Lộc Thành -Bảo Lâm','COUNTRY_VN',63,698,NULL,true,'2019-09-27 14:34:59.689','2019-09-27 14:34:59.689')
,('TT GDTX Bảo Lâm','COUNTRY_VN',63,698,NULL,true,'2019-09-27 14:34:59.690','2019-09-27 14:34:59.690')
,('THPT Cát Tiên','COUNTRY_VN',63,699,NULL,true,'2019-09-27 14:34:59.691','2019-09-27 14:34:59.691')
,('THPT Gia Viễn-Cát Tiên','COUNTRY_VN',63,699,NULL,true,'2019-09-27 14:34:59.693','2019-09-27 14:34:59.693')
,('THPT Quang Trung -Cát Tiên','COUNTRY_VN',63,699,NULL,true,'2019-09-27 14:34:59.694','2019-09-27 14:34:59.694')
,('TT GDTX Cát Tiên','COUNTRY_VN',63,699,NULL,true,'2019-09-27 14:34:59.695','2019-09-27 14:34:59.695')
,('THPT Di Linh','COUNTRY_VN',63,700,NULL,true,'2019-09-27 14:34:59.696','2019-09-27 14:34:59.696')
,('THPT Hòa Ninh Di Linh','COUNTRY_VN',63,700,NULL,true,'2019-09-27 14:34:59.697','2019-09-27 14:34:59.697')
,('THPT Lê Hồng Phong','COUNTRY_VN',63,700,NULL,true,'2019-09-27 14:34:59.698','2019-09-27 14:34:59.698')
,('THPT Nguyễn Huệ - Di Linh','COUNTRY_VN',63,700,NULL,true,'2019-09-27 14:34:59.699','2019-09-27 14:34:59.699')
,('THPT Nguyễn Viết Xuân','COUNTRY_VN',63,700,NULL,true,'2019-09-27 14:34:59.699','2019-09-27 14:34:59.699')
,('THPT Phan Bội Châu','COUNTRY_VN',63,700,NULL,true,'2019-09-27 14:34:59.700','2019-09-27 14:34:59.700')
,('TT KTTH-HN Di Linh','COUNTRY_VN',63,700,NULL,true,'2019-09-27 14:34:59.701','2019-09-27 14:34:59.701')
,('THPT Đạ Huoai','COUNTRY_VN',63,701,NULL,true,'2019-09-27 14:34:59.701','2019-09-27 14:34:59.701')
,('TT KTTH-HN Đạ Huoai','COUNTRY_VN',63,701,NULL,true,'2019-09-27 14:34:59.702','2019-09-27 14:34:59.702')
,('THCS & THPT DTNT Liên huyện phía Nam','COUNTRY_VN',63,702,NULL,true,'2019-09-27 14:34:59.703','2019-09-27 14:34:59.703')
,('THPT Đạ Tẻh','COUNTRY_VN',63,702,NULL,true,'2019-09-27 14:34:59.704','2019-09-27 14:34:59.704')
,('THPT Lê Quý Đôn -Đạ Tẻh','COUNTRY_VN',63,702,NULL,true,'2019-09-27 14:34:59.704','2019-09-27 14:34:59.704')
,('THPT TT Nguyễn Khuyến -Đạ Tẻh','COUNTRY_VN',63,702,NULL,true,'2019-09-27 14:34:59.704','2019-09-27 14:34:59.704')
,('TT KTTH-HN Đạ Tẻh','COUNTRY_VN',63,702,NULL,true,'2019-09-27 14:34:59.705','2019-09-27 14:34:59.705')
,('THPT ĐạTống','COUNTRY_VN',63,703,NULL,true,'2019-09-27 14:34:59.706','2019-09-27 14:34:59.706')
,('THPT Nguyễn Chí Thanh - Đam Rông','COUNTRY_VN',63,703,NULL,true,'2019-09-27 14:34:59.706','2019-09-27 14:34:59.706')
,('THPT Phan Đình Phùng','COUNTRY_VN',63,703,NULL,true,'2019-09-27 14:34:59.706','2019-09-27 14:34:59.706')
,('Trung tâm GDTX Đam Rông','COUNTRY_VN',63,703,NULL,true,'2019-09-27 14:34:59.707','2019-09-27 14:34:59.707')
,('THCS & THPT Ngô Gia Tự','COUNTRY_VN',63,704,NULL,true,'2019-09-27 14:34:59.710','2019-09-27 14:34:59.710')
,('THPT Đơn Dương','COUNTRY_VN',63,704,NULL,true,'2019-09-27 14:34:59.711','2019-09-27 14:34:59.711')
,('THPT Hùng Vương','COUNTRY_VN',63,704,NULL,true,'2019-09-27 14:34:59.712','2019-09-27 14:34:59.712')
,('THPT Lê Lợi -Đơn Dương','COUNTRY_VN',63,704,NULL,true,'2019-09-27 14:34:59.713','2019-09-27 14:34:59.713')
,('THPT Próh -Đơn Dương','COUNTRY_VN',63,704,NULL,true,'2019-09-27 14:34:59.714','2019-09-27 14:34:59.714')
,('TT KTTH-HN Đơn Dương','COUNTRY_VN',63,704,NULL,true,'2019-09-27 14:34:59.715','2019-09-27 14:34:59.715')
,('TC KT-KT Quốc việt','COUNTRY_VN',63,705,NULL,true,'2019-09-27 14:34:59.716','2019-09-27 14:34:59.716')
,('THPT Chu Văn An-Đức Trọng','COUNTRY_VN',63,705,NULL,true,'2019-09-27 14:34:59.717','2019-09-27 14:34:59.717')
,('THPT Đức Trọng','COUNTRY_VN',63,705,NULL,true,'2019-09-27 14:34:59.717','2019-09-27 14:34:59.717')
,('THPT Hoàng Hoa Thám','COUNTRY_VN',63,705,NULL,true,'2019-09-27 14:34:59.718','2019-09-27 14:34:59.718')
,('THPT Lương Thế Vinh','COUNTRY_VN',63,705,NULL,true,'2019-09-27 14:34:59.718','2019-09-27 14:34:59.718')
,('THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',63,705,NULL,true,'2019-09-27 14:34:59.718','2019-09-27 14:34:59.718')
,('THPT Nguyễn Thái Bình','COUNTRY_VN',63,705,NULL,true,'2019-09-27 14:34:59.719','2019-09-27 14:34:59.719')
,('THPT Đạ Sar Lạc Dương','COUNTRY_VN',63,706,NULL,true,'2019-09-27 14:34:59.720','2019-09-27 14:34:59.720')
,('THPT Lang Biang','COUNTRY_VN',63,706,NULL,true,'2019-09-27 14:34:59.720','2019-09-27 14:34:59.720')
,('Trung tâm GDTX Lac Dương','COUNTRY_VN',63,706,NULL,true,'2019-09-27 14:34:59.721','2019-09-27 14:34:59.721')
,('THPT Huỳnh Thúc Kháng','COUNTRY_VN',63,707,NULL,true,'2019-09-27 14:34:59.722','2019-09-27 14:34:59.722')
,('THPT Lâm Hà','COUNTRY_VN',63,707,NULL,true,'2019-09-27 14:34:59.722','2019-09-27 14:34:59.722')
,('THPT Lê Quý Đôn -Lâm Hà','COUNTRY_VN',63,707,NULL,true,'2019-09-27 14:34:59.722','2019-09-27 14:34:59.722')
,('THPT Tân Hà-Lâm Hà','COUNTRY_VN',63,707,NULL,true,'2019-09-27 14:34:59.723','2019-09-27 14:34:59.723')
,('THPT Thăng Long -Lâm Hà','COUNTRY_VN',63,707,NULL,true,'2019-09-27 14:34:59.723','2019-09-27 14:34:59.723')
,('TT KTTH-HN Lâm Hà','COUNTRY_VN',63,707,NULL,true,'2019-09-27 14:34:59.723','2019-09-27 14:34:59.723')
,('CĐ Công nghệ & Kinh tế Bảo Lộc','COUNTRY_VN',63,708,NULL,true,'2019-09-27 14:34:59.726','2019-09-27 14:34:59.726')
,('Dân lập Lê Lợi -Bảo Lộc','COUNTRY_VN',63,708,NULL,true,'2019-09-27 14:34:59.727','2019-09-27 14:34:59.727')
,('TC Nghề Bảo Lộc','COUNTRY_VN',63,708,NULL,true,'2019-09-27 14:34:59.728','2019-09-27 14:34:59.728')
,('THPT Bá Thiên - Bảo Lộc','COUNTRY_VN',63,708,NULL,true,'2019-09-27 14:34:59.730','2019-09-27 14:34:59.730')
,('THPT Bảo Lộc','COUNTRY_VN',63,708,NULL,true,'2019-09-27 14:34:59.731','2019-09-27 14:34:59.731')
,('THPT BC Nguyễn Du -Bảo Lộc','COUNTRY_VN',63,708,NULL,true,'2019-09-27 14:34:59.732','2019-09-27 14:34:59.732')
,('THPT chuyên Bảo Lộc','COUNTRY_VN',63,708,NULL,true,'2019-09-27 14:34:59.733','2019-09-27 14:34:59.733')
,('THPT Lê Thị Pha -Bảo Lộc','COUNTRY_VN',63,708,NULL,true,'2019-09-27 14:34:59.734','2019-09-27 14:34:59.734')
,('THPT Lộc Phát-Bảo Lộc','COUNTRY_VN',63,708,NULL,true,'2019-09-27 14:34:59.735','2019-09-27 14:34:59.735')
,('THPT Lộc Thanh','COUNTRY_VN',63,708,NULL,true,'2019-09-27 14:34:59.735','2019-09-27 14:34:59.735')
,('THPT Nguyễn Tri Phương','COUNTRY_VN',63,708,NULL,true,'2019-09-27 14:34:59.736','2019-09-27 14:34:59.736')
,('THPT TT Duy Tân Bảo Lộc','COUNTRY_VN',63,708,NULL,true,'2019-09-27 14:34:59.737','2019-09-27 14:34:59.737')
,('TT GDTX Lâm Đồng','COUNTRY_VN',63,708,NULL,true,'2019-09-27 14:34:59.737','2019-09-27 14:34:59.737')
,('CĐ KT-KT Lâm Đồng','COUNTRY_VN',63,709,NULL,true,'2019-09-27 14:34:59.738','2019-09-27 14:34:59.738')
,('CĐ Y tế Lâm Đồng','COUNTRY_VN',63,709,NULL,true,'2019-09-27 14:34:59.739','2019-09-27 14:34:59.739')
,('Hermann Gmeiner','COUNTRY_VN',63,709,NULL,true,'2019-09-27 14:34:59.739','2019-09-27 14:34:59.739')
,('Phân hiệu TC Văn thư lưu trữ TVV','COUNTRY_VN',63,709,NULL,true,'2019-09-27 14:34:59.740','2019-09-27 14:34:59.740')
,('TC Du Lịch Dalat','COUNTRY_VN',63,709,NULL,true,'2019-09-27 14:34:59.740','2019-09-27 14:34:59.740')
,('THCS &THPT Nguyễn Du -Đà Lạt','COUNTRY_VN',63,709,NULL,true,'2019-09-27 14:34:59.743','2019-09-27 14:34:59.743')
,('THPT Chi Lăng','COUNTRY_VN',63,709,NULL,true,'2019-09-27 14:34:59.744','2019-09-27 14:34:59.744')
,('THPT chuyên Thăng Long -Đà lạt','COUNTRY_VN',63,709,NULL,true,'2019-09-27 14:34:59.744','2019-09-27 14:34:59.744')
,('THPT DTNT Tỉnh','COUNTRY_VN',63,709,NULL,true,'2019-09-27 14:34:59.745','2019-09-27 14:34:59.745')
,('THPT Đống Đa','COUNTRY_VN',63,709,NULL,true,'2019-09-27 14:34:59.746','2019-09-27 14:34:59.746')
,('THPT Phù Đổng','COUNTRY_VN',63,709,NULL,true,'2019-09-27 14:34:59.746','2019-09-27 14:34:59.746')
,('THPT Tà Nung-Đà Lạt','COUNTRY_VN',63,709,NULL,true,'2019-09-27 14:34:59.747','2019-09-27 14:34:59.747')
,('THPT Tây Sơn','COUNTRY_VN',63,709,NULL,true,'2019-09-27 14:34:59.748','2019-09-27 14:34:59.748')
,('THPT Trần Phú -Đà Lạt','COUNTRY_VN',63,709,NULL,true,'2019-09-27 14:34:59.749','2019-09-27 14:34:59.749')
,('THPT Xuân Trường','COUNTRY_VN',63,709,NULL,true,'2019-09-27 14:34:59.749','2019-09-27 14:34:59.749')
,('THPT Yersin -Đà Lạt','COUNTRY_VN',63,709,NULL,true,'2019-09-27 14:34:59.750','2019-09-27 14:34:59.750')
,('TT GDTX Đà Lạt','COUNTRY_VN',63,709,NULL,true,'2019-09-27 14:34:59.750','2019-09-27 14:34:59.750')
,('THCS-THPT Chuyên Trần Đại Nghĩa','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:59.751','2019-09-27 14:34:59.751')
,('THCS Chu Văn An','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:59.751','2019-09-27 14:34:59.751')
,('THCS Đồng Khởi','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:59.752','2019-09-27 14:34:59.752')
,('THCS Đức Trí','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:59.752','2019-09-27 14:34:59.752')
,('THCS Huỳnh Khương Ninh','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:59.753','2019-09-27 14:34:59.753')
,('THCS Minh Đức','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:59.753','2019-09-27 14:34:59.753')
,('THCS Nguyễn Du','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:59.753','2019-09-27 14:34:59.753')
,('THCS Trần Văn Ơn','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:59.754','2019-09-27 14:34:59.754')
,('THCS Văn Lang','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:59.754','2019-09-27 14:34:59.754')
,('THCS Võ Trường Toản','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:59.754','2019-09-27 14:34:59.754')
,('THCS-THPT Đăng Khoa','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:57.001','2019-09-27 14:34:59.755')
,('TH-THCS-THPT Quốc tế Á Châu','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:59.755','2019-09-27 14:34:59.755')
,('TH-THCS-THPT Úc Châu','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:59.756','2019-09-27 14:34:59.756')
,('THCS-THPT Châu Á Thái Bình Dương','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:59.756','2019-09-27 14:34:59.756')
,('TH-THCS-THPT Nam Mỹ','COUNTRY_VN',1,1,NULL,true,'2019-09-27 14:34:59.756','2019-09-27 14:34:59.756')
,('THCS An Phú','COUNTRY_VN',1,2,NULL,true,'2019-09-27 14:34:59.757','2019-09-27 14:34:59.757')
,('THCS Bình An','COUNTRY_VN',1,2,NULL,true,'2019-09-27 14:34:59.758','2019-09-27 14:34:59.758')
,('THCS Giồng Ông Tố','COUNTRY_VN',1,2,NULL,true,'2019-09-27 14:34:59.759','2019-09-27 14:34:59.759')
,('THCS Lương Định Của','COUNTRY_VN',1,2,NULL,true,'2019-09-27 14:34:59.760','2019-09-27 14:34:59.760')
,('THCS Nguyễn Thị Định','COUNTRY_VN',1,2,NULL,true,'2019-09-27 14:34:59.761','2019-09-27 14:34:59.761')
,('THCS Nguyễn Văn Trỗi','COUNTRY_VN',1,2,NULL,true,'2019-09-27 14:34:59.762','2019-09-27 14:34:59.762')
,('TH-THCS Tuệ Đức','COUNTRY_VN',1,2,NULL,true,'2019-09-27 14:34:59.763','2019-09-27 14:34:59.763')
,('THCS Thạnh Mỹ Lợi','COUNTRY_VN',1,2,NULL,true,'2019-09-27 14:34:59.764','2019-09-27 14:34:59.764')
,('THCS Trần Quốc Toản','COUNTRY_VN',1,2,NULL,true,'2019-09-27 14:34:59.765','2019-09-27 14:34:59.765')
,('THCS Cát Lái','COUNTRY_VN',1,2,NULL,true,'2019-09-27 14:34:59.766','2019-09-27 14:34:59.766')
,('Song ngữ Quốc tế Horizon','COUNTRY_VN',1,2,NULL,true,'2019-09-27 14:34:59.766','2019-09-27 14:34:59.766')
,('THCS Bạch Đằng','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:59.767','2019-09-27 14:34:59.767')
,('THCS Bàn Cờ','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:59.767','2019-09-27 14:34:59.767')
,('THCS Colette','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:59.768','2019-09-27 14:34:59.768')
,('THCS Đoàn Thị Điểm','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:59.768','2019-09-27 14:34:59.768')
,('THCS Hai Bà Trưng','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:59.768','2019-09-27 14:34:59.768')
,('THCS Kiến Thiết','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:59.769','2019-09-27 14:34:59.769')
,('THCS Lê Lợi','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:59.769','2019-09-27 14:34:59.769')
,('THCS Lê Quý Đôn','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:59.769','2019-09-27 14:34:59.769')
,('THCS Lương Thế Vinh','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:59.770','2019-09-27 14:34:59.770')
,('THCS Phan Sào Nam','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:59.770','2019-09-27 14:34:59.770')
,('THCS Thăng Long','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:59.770','2019-09-27 14:34:59.770')
,('THCS-THPT Nguyễn Bỉnh Khiêm','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:57.022','2019-09-27 14:34:59.771')
,('TH-THCS-THPT Tây Úc','COUNTRY_VN',1,3,NULL,true,'2019-09-27 14:34:59.771','2019-09-27 14:34:59.771')
,('THCS Chi Lăng','COUNTRY_VN',1,4,NULL,true,'2019-09-27 14:34:59.771','2019-09-27 14:34:59.771')
,('THCS Khánh Hội A','COUNTRY_VN',1,4,NULL,true,'2019-09-27 14:34:59.772','2019-09-27 14:34:59.772')
,('THCS Nguyễn Huệ','COUNTRY_VN',1,4,NULL,true,'2019-09-27 14:34:59.773','2019-09-27 14:34:59.773')
,('THCS Quang Trung','COUNTRY_VN',1,4,NULL,true,'2019-09-27 14:34:59.773','2019-09-27 14:34:59.773')
,('THCS Tăng Bạt Hổ A','COUNTRY_VN',1,4,NULL,true,'2019-09-27 14:34:59.773','2019-09-27 14:34:59.773')
,('THCS Vân Đồn','COUNTRY_VN',1,4,NULL,true,'2019-09-27 14:34:59.775','2019-09-27 14:34:59.775')
,('THCS Ba Đình','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:59.776','2019-09-27 14:34:59.776')
,('THCS Hồng Bàng','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:59.776','2019-09-27 14:34:59.776')
,('THCS Kim Đồng','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:59.777','2019-09-27 14:34:59.777')
,('THCS Lý Phong','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:59.777','2019-09-27 14:34:59.777')
,('THCS Mạch Kiếm Hùng','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:59.778','2019-09-27 14:34:59.778')
,('THCS Trần Bội Cơ','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:59.778','2019-09-27 14:34:59.778')
,('Trung học thực hành Sài Gòn','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:59.779','2019-09-27 14:34:59.779')
,('THCS-THPT An Đông','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:57.036','2019-09-27 14:34:59.781')
,('THCS-THPT Quang Trung Nguyễn Huệ','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:57.038','2019-09-27 14:34:59.782')
,('TH-THCS-THPT Văn Lang','COUNTRY_VN',1,5,NULL,true,'2019-09-27 14:34:59.782','2019-09-27 14:34:59.782')
,('THCS Bình Tây','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:59.783','2019-09-27 14:34:59.783')
,('THCS Đoàn Kết','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:59.784','2019-09-27 14:34:59.784')
,('THCS Hậu Giang','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:59.784','2019-09-27 14:34:59.784')
,('THCS Hoàng Lê Kha','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:59.785','2019-09-27 14:34:59.785')
,('THCS Lam Sơn','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:59.785','2019-09-27 14:34:59.785')
,('THCS Nguyễn Đức Cảnh','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:59.786','2019-09-27 14:34:59.786')
,('THCS Nguyễn Văn Luông','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:59.786','2019-09-27 14:34:59.786')
,('THCS Phạm Đình Hổ','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:59.787','2019-09-27 14:34:59.787')
,('THCS Phú Định','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:59.787','2019-09-27 14:34:59.787')
,('THCS Văn Thân','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:59.788','2019-09-27 14:34:59.788')
,('THCS-THPT Đào Duy Anh','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:57.050','2019-09-27 14:34:59.788')
,('THCS-THPT Phan Bội Châu','COUNTRY_VN',1,6,NULL,true,'2019-09-27 14:34:59.789','2019-09-27 14:34:59.789')
,('THCS Hoàng Quốc Việt','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:59.789','2019-09-27 14:34:59.789')
,('THCS Huỳnh Tấn Phát','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:59.790','2019-09-27 14:34:59.790')
,('THCS Nguyễn Hiền','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:59.790','2019-09-27 14:34:59.790')
,('THCS Nguyễn Hữu Thọ','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:59.792','2019-09-27 14:34:59.792')
,('THCS Nguyễn Thị Thập','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:59.793','2019-09-27 14:34:59.793')
,('THCS Phạm Hữu Lầu','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:59.794','2019-09-27 14:34:59.794')
,('THCS Trần Quốc Tuấn','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:59.795','2019-09-27 14:34:59.795')
,('TH-THCS-THPT Nam Sài Gòn','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:59.797','2019-09-27 14:34:59.797')
,('THCS-THPT Đức Trí','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:57.056','2019-09-27 14:34:59.798')
,('THCS-THPT Đinh Thiện Lý','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:57.056','2019-09-27 14:34:59.799')
,('THCS-THPT Sao Việt','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:57.052','2019-09-27 14:34:59.799')
,('THCS-THPT Quốc tế Canada','COUNTRY_VN',1,7,NULL,true,'2019-09-27 14:34:59.800','2019-09-27 14:34:59.800')
,('THCS Bình An','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:59.801','2019-09-27 14:34:59.801')
,('THCS Bình Đông','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:59.801','2019-09-27 14:34:59.801')
,('THCS Chánh Hưng','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:59.802','2019-09-27 14:34:59.802')
,('THCS Dương Bá Trạc','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:59.802','2019-09-27 14:34:59.802')
,('THCS Khánh Bình','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:59.803','2019-09-27 14:34:59.803')
,('THCS Lê Lai','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:59.803','2019-09-27 14:34:59.803')
,('THCS Lý Thánh Tông','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:59.803','2019-09-27 14:34:59.803')
,('THCS Phan Đăng Lưu','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:59.804','2019-09-27 14:34:59.804')
,('THCS Phú Lợi','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:59.805','2019-09-27 14:34:59.805')
,('THCS Sương Nguyệt Anh','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:59.805','2019-09-27 14:34:59.805')
,('THCS Trần Danh Ninh','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:59.805','2019-09-27 14:34:59.805')
,('THCS Tùng Thiện Vương','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:59.806','2019-09-27 14:34:59.806')
,('THPT chuyên NK TDTT Nguyễn Thị Định','COUNTRY_VN',1,8,NULL,true,'2019-09-27 14:34:59.806','2019-09-27 14:34:59.806')
,('THCS Đặng Tấn Tài','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:59.807','2019-09-27 14:34:59.807')
,('THCS Hoa Lư','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:59.810','2019-09-27 14:34:59.810')
,('THCS Hưng Bình','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:59.811','2019-09-27 14:34:59.811')
,('THCS Long Bình','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:59.812','2019-09-27 14:34:59.812')
,('THCS Long Phước','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:59.812','2019-09-27 14:34:59.812')
,('THCS Long Trường','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:59.813','2019-09-27 14:34:59.813')
,('THCS Phú Hữu','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:59.814','2019-09-27 14:34:59.814')
,('THCS Phước Bình','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:59.815','2019-09-27 14:34:59.815')
,('THCS Tân Phú','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:59.816','2019-09-27 14:34:59.816')
,('THCS Tăng Nhơn Phú B','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:59.817','2019-09-27 14:34:59.817')
,('THCS Trần Quốc Toản','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:59.817','2019-09-27 14:34:59.817')
,('THCS Trường Thạnh','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:59.817','2019-09-27 14:34:59.817')
,('TH-THCS-THPT Ngô Thời Nhiệm','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:57.067','2019-09-27 14:34:59.818')
,('THCS-THPT Hoa Sen','COUNTRY_VN',1,9,NULL,true,'2019-09-27 14:34:59.818','2019-09-27 14:34:59.818')
,('THCS Cách Mạng Tháng Tám','COUNTRY_VN',1,10,NULL,true,'2019-09-27 14:34:59.818','2019-09-27 14:34:59.818')
,('THCS Duy Tân','COUNTRY_VN',1,10,NULL,true,'2019-09-27 14:34:59.819','2019-09-27 14:34:59.819')
,('THCS Hoàng Văn Thụ','COUNTRY_VN',1,10,NULL,true,'2019-09-27 14:34:59.819','2019-09-27 14:34:59.819')
,('THCS Lạc Hồng','COUNTRY_VN',1,10,NULL,true,'2019-09-27 14:34:59.819','2019-09-27 14:34:59.819')
,('THCS Nguyễn Tri Phương','COUNTRY_VN',1,10,NULL,true,'2019-09-27 14:34:59.820','2019-09-27 14:34:59.820')
,('THCS Nguyễn Văn Tố','COUNTRY_VN',1,10,NULL,true,'2019-09-27 14:34:59.820','2019-09-27 14:34:59.820')
,('THCS Trần Phú','COUNTRY_VN',1,10,NULL,true,'2019-09-27 14:34:59.821','2019-09-27 14:34:59.821')
,('THCS-THPT Sương Nguyệt Anh','COUNTRY_VN',1,10,NULL,true,'2019-09-27 14:34:57.072','2019-09-27 14:34:59.821')
,('THCS-THPT Diên Hồng','COUNTRY_VN',1,10,NULL,true,'2019-09-27 14:34:57.070','2019-09-27 14:34:59.821')
,('TH-THCS-THPT Vạn Hạnh','COUNTRY_VN',1,10,NULL,true,'2019-09-27 14:34:57.073','2019-09-27 14:34:59.822')
,('THCS-THPT Duy Tân','COUNTRY_VN',1,10,NULL,true,'2019-09-27 14:34:57.073','2019-09-27 14:34:59.822')
,('TH-THCS-THPT Việt Úc','COUNTRY_VN',1,10,NULL,true,'2019-09-27 14:34:59.822','2019-09-27 14:34:59.822')
,('THCS Chu Văn An','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:59.823','2019-09-27 14:34:59.823')
,('THCS Hậu Giang','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:59.823','2019-09-27 14:34:59.823')
,('THCS Lê Anh Xuân','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:59.823','2019-09-27 14:34:59.823')
,('THCS Lê Quý Đôn','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:59.825','2019-09-27 14:34:59.825')
,('THCS Nguyễn Huệ','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:59.826','2019-09-27 14:34:59.826')
,('THCS Nguyễn Minh Hoàng','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:59.826','2019-09-27 14:34:59.826')
,('THCS Nguyễn Văn Phú','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:59.827','2019-09-27 14:34:59.827')
,('THCS Phú Thọ','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:59.827','2019-09-27 14:34:59.827')
,('THCS Việt Mỹ','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:59.828','2019-09-27 14:34:59.828')
,('THCS Lữ Gia','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:59.829','2019-09-27 14:34:59.829')
,('TH-THCS-THPT Trương Vĩnh Ký','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:57.083','2019-09-27 14:34:59.830')
,('THCS-THPT Quốc tế APU','COUNTRY_VN',1,11,NULL,true,'2019-09-27 14:34:59.830','2019-09-27 14:34:59.830')
,('THCS An Phú Đông','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:59.831','2019-09-27 14:34:59.831')
,('THCS Lương Thế Vinh','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:59.832','2019-09-27 14:34:59.832')
,('THCS Nguyễn An Ninh','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:59.832','2019-09-27 14:34:59.832')
,('THCS Nguyễn Ảnh Thủ','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:59.833','2019-09-27 14:34:59.833')
,('THCS Nguyễn Chí Thanh','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:59.833','2019-09-27 14:34:59.833')
,('THCS Nguyễn Huệ','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:59.834','2019-09-27 14:34:59.834')
,('THCS Nguyễn Trung Trực','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:59.834','2019-09-27 14:34:59.834')
,('THCS Nguyễn Vĩnh Nghiệp','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:59.834','2019-09-27 14:34:59.834')
,('THCS Phan Bội Châu','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:59.835','2019-09-27 14:34:59.835')
,('THCS Trần Hưng Đạo','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:59.835','2019-09-27 14:34:59.835')
,('THCS Trần Quang Khải','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:59.836','2019-09-27 14:34:59.836')
,('THCS Nguyễn Hiền','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:59.836','2019-09-27 14:34:59.836')
,('THCS Hà Huy Tập','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:59.836','2019-09-27 14:34:59.836')
,('THCS-THPT Bắc Sơn','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:57.085','2019-09-27 14:34:59.837')
,('THCS-THPT Bạch Đằng','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:59.837','2019-09-27 14:34:59.837')
,('THCS-THPT Lạc Hồng','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:57.085','2019-09-27 14:34:59.838')
,('THCS-THPT Hoa Lư','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:57.086','2019-09-27 14:34:59.838')
,('TH-THCS-THPT Mỹ Việt','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:59.839','2019-09-27 14:34:59.839')
,('THCS-THPT Tuệ Đức','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:59.840','2019-09-27 14:34:59.840')
,('THCS-THPT Ngọc Viễn Đông','COUNTRY_VN',1,12,NULL,true,'2019-09-27 14:34:59.842','2019-09-27 14:34:59.842')
,('THCS Bình Chánh','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.843','2019-09-27 14:34:59.843')
,('THCS Đa Phước','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.844','2019-09-27 14:34:59.844')
,('THCS Đồng Đen','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.844','2019-09-27 14:34:59.844')
,('THCS Gò Xoài','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.845','2019-09-27 14:34:59.845')
,('THCS Hưng Long','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.846','2019-09-27 14:34:59.846')
,('THCS Lê Minh Xuân','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.847','2019-09-27 14:34:59.847')
,('THCS Nguyễn Thái Bình','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.847','2019-09-27 14:34:59.847')
,('THCS Nguyễn Văn Linh','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.849','2019-09-27 14:34:59.849')
,('THCS Phạm Văn Hai','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.849','2019-09-27 14:34:59.849')
,('THCS Phong Phú','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.850','2019-09-27 14:34:59.850')
,('THCS Qui Đức','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.851','2019-09-27 14:34:59.851')
,('THCS Tân Kiên','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.852','2019-09-27 14:34:59.852')
,('THCS Tân Nhựt','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.852','2019-09-27 14:34:59.852')
,('THCS Tân Quý Tây','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.853','2019-09-27 14:34:59.853')
,('THCS Tân Túc','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.853','2019-09-27 14:34:59.853')
,('TH-THCS Thế Giới Trẻ Em','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.854','2019-09-27 14:34:59.854')
,('THCS Vĩnh Lộc A','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.854','2019-09-27 14:34:59.854')
,('THCS Vĩnh Lộc B','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.854','2019-09-27 14:34:59.854')
,('THCS Võ Văn Vân','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.855','2019-09-27 14:34:59.855')
,('THCS-THPT Bắc Mỹ','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.856','2019-09-27 14:34:59.856')
,('TH-THCS-THPT Albert Einstein','COUNTRY_VN',1,23,NULL,true,'2019-09-27 14:34:59.856','2019-09-27 14:34:59.856')
,('THCS An Lạc','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.857','2019-09-27 14:34:59.857')
,('THCS Bình Hưng Hòa','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.858','2019-09-27 14:34:59.858')
,('THCS Bình Tân','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.859','2019-09-27 14:34:59.859')
,('THCS Bình Trị Đông','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.860','2019-09-27 14:34:59.860')
,('THCS Bình Trị Đông A','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.860','2019-09-27 14:34:59.860')
,('THCS Hồ Văn Long','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.861','2019-09-27 14:34:59.861')
,('THCS Huỳnh Văn Nghệ','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.862','2019-09-27 14:34:59.862')
,('THCS Lê Tấn Bê','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.862','2019-09-27 14:34:59.862')
,('THCS Lý Thường Kiệt','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.863','2019-09-27 14:34:59.863')
,('THCS Nguyễn Trãi','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.864','2019-09-27 14:34:59.864')
,('THCS Tân Tạo','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.865','2019-09-27 14:34:59.865')
,('THCS Tân Tạo A','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.866','2019-09-27 14:34:59.866')
,('THCS Trần Quốc Toản','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.866','2019-09-27 14:34:59.866')
,('THCS Trí Tuệ Việt','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.867','2019-09-27 14:34:59.867')
,('TH-THCS-THPT Chu Văn An','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.867','2019-09-27 14:34:59.867')
,('THCS-THPT Phan Châu Trinh','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.868','2019-09-27 14:34:59.868')
,('THCS-THPT Ngôi Sao','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.868','2019-09-27 14:34:59.868')
,('TH-THCS-THPT Ngôi Sao Nhỏ','COUNTRY_VN',1,13,NULL,true,'2019-09-27 14:34:59.868','2019-09-27 14:34:59.868')
,('THCS Bình Lợi Trung','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.869','2019-09-27 14:34:59.869')
,('THCS Bình Quới Tây','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.869','2019-09-27 14:34:59.869')
,('THCS Cù Chính Lan','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.870','2019-09-27 14:34:59.870')
,('THCS Cửu Long','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.870','2019-09-27 14:34:59.870')
,('THCS Điện Biên','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.870','2019-09-27 14:34:59.870')
,('THCS Đống Đa','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.871','2019-09-27 14:34:59.871')
,('THCS Hà Huy Tập','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.871','2019-09-27 14:34:59.871')
,('THCS Lam Sơn','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.872','2019-09-27 14:34:59.872')
,('THCS Lê Văn Tám','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.872','2019-09-27 14:34:59.872')
,('THCS Nguyễn Văn Bé','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.872','2019-09-27 14:34:59.872')
,('THCS Phú Mỹ','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.873','2019-09-27 14:34:59.873')
,('THCS Rạng Đông','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.873','2019-09-27 14:34:59.873')
,('THCS Thanh Đa','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.874','2019-09-27 14:34:59.874')
,('THCS Trương Công Định','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.875','2019-09-27 14:34:59.875')
,('THCS Yên Thế','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.876','2019-09-27 14:34:59.876')
,('TH-THCS-THPT Mùa Xuân','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.876','2019-09-27 14:34:59.876')
,('TH-THCS-THPT Vinschool','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.877','2019-09-27 14:34:59.877')
,('TH-THCS-THPT Anh Quốc','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.878','2019-09-27 14:34:59.878')
,('TH-THCS-THPT Hoàng Gia','COUNTRY_VN',1,14,NULL,true,'2019-09-27 14:34:59.878','2019-09-27 14:34:59.878')
,('THCS An Thới Đông','COUNTRY_VN',1,20,NULL,true,'2019-09-27 14:34:59.879','2019-09-27 14:34:59.879')
,('THCS Bình Khánh','COUNTRY_VN',1,20,NULL,true,'2019-09-27 14:34:59.880','2019-09-27 14:34:59.880')
,('THCS Cần Thạnh','COUNTRY_VN',1,20,NULL,true,'2019-09-27 14:34:59.881','2019-09-27 14:34:59.881')
,('THCS Doi Lầu','COUNTRY_VN',1,20,NULL,true,'2019-09-27 14:34:59.882','2019-09-27 14:34:59.882')
,('THCS Long Hòa','COUNTRY_VN',1,20,NULL,true,'2019-09-27 14:34:59.882','2019-09-27 14:34:59.882')
,('THCS Lý Nhơn','COUNTRY_VN',1,20,NULL,true,'2019-09-27 14:34:59.882','2019-09-27 14:34:59.882')
,('THCS Tam Thôn Hiệp','COUNTRY_VN',1,20,NULL,true,'2019-09-27 14:34:59.883','2019-09-27 14:34:59.883')
,('THCS An Nhơn Tây','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.883','2019-09-27 14:34:59.883')
,('THCS An Phú','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.883','2019-09-27 14:34:59.883')
,('THCS Bình Hòa','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.884','2019-09-27 14:34:59.884')
,('THCS Hòa Phú','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.884','2019-09-27 14:34:59.884')
,('THCS Nguyễn Văn Xơ','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.885','2019-09-27 14:34:59.885')
,('THCS Nhuận Đức','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.885','2019-09-27 14:34:59.885')
,('THCS Phạm Văn Cội','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.885','2019-09-27 14:34:59.885')
,('THCS Phú Hòa Đông','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.886','2019-09-27 14:34:59.886')
,('THCS Phú Mỹ Hưng','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.886','2019-09-27 14:34:59.886')
,('THCS Phước Hiệp','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.886','2019-09-27 14:34:59.886')
,('THCS Phước Thạnh','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.887','2019-09-27 14:34:59.887')
,('THCS Phước Vĩnh An','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.887','2019-09-27 14:34:59.887')
,('THCS Tân An Hội','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.887','2019-09-27 14:34:59.887')
,('THCS Tân Phú Trung','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.888','2019-09-27 14:34:59.888')
,('THCS Tân Thạnh Đông','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.888','2019-09-27 14:34:59.888')
,('THCS Tân Thạnh Tây','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.888','2019-09-27 14:34:59.888')
,('THCS Tân Thông Hội','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.889','2019-09-27 14:34:59.889')
,('THCS Tân Tiến','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.889','2019-09-27 14:34:59.889')
,('THCS Thị Trấn','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.890','2019-09-27 14:34:59.890')
,('THCS Thị Trấn 2','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.890','2019-09-27 14:34:59.890')
,('THCS Trung An','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.892','2019-09-27 14:34:59.892')
,('THCS Trung Lập','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.893','2019-09-27 14:34:59.893')
,('THCS Trung Lập Hạ','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.894','2019-09-27 14:34:59.894')
,('TH-THCS Tân Trung','COUNTRY_VN',1,21,NULL,true,'2019-09-27 14:34:59.895','2019-09-27 14:34:59.895')
,('THCS An Nhơn','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.896','2019-09-27 14:34:59.896')
,('THCS Gò Vấp','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.898','2019-09-27 14:34:59.898')
,('THCS Huỳnh Văn Nghệ','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.898','2019-09-27 14:34:59.898')
,('THCS Lý Tự Trọng','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.899','2019-09-27 14:34:59.899')
,('THCS Nguyễn Du','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.900','2019-09-27 14:34:59.900')
,('THCS Nguyễn Trãi','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.901','2019-09-27 14:34:59.901')
,('THCS Nguyễn Văn Nghi','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.901','2019-09-27 14:34:59.901')
,('THCS Nguyễn Văn Trỗi','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.901','2019-09-27 14:34:59.901')
,('THCS Phạm Văn Chiêu','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.902','2019-09-27 14:34:59.902')
,('THCS Phan Tây Hồ','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.902','2019-09-27 14:34:59.902')
,('THCS Phan Văn Trị','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.903','2019-09-27 14:34:59.903')
,('THCS Quang Trung','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.903','2019-09-27 14:34:59.903')
,('THCS Tân Sơn','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.903','2019-09-27 14:34:59.903')
,('THCS Thông Tây Hội','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.904','2019-09-27 14:34:59.904')
,('THCS Trường Sơn','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.904','2019-09-27 14:34:59.904')
,('THCS-THPT Hermann Gmeiner','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.904','2019-09-27 14:34:59.904')
,('TH-THCS-THPT Đại Việt','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.905','2019-09-27 14:34:59.905')
,('THCS-THPT Nguyễn Tri Phương','COUNTRY_VN',1,15,NULL,true,'2019-09-27 14:34:59.906','2019-09-27 14:34:59.906')
,('THCS Đặng Công Bỉnh','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:59.906','2019-09-27 14:34:59.906')
,('THCS Đỗ Văn Dậy','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:59.907','2019-09-27 14:34:59.907')
,('THCS Đông Thạnh','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:59.909','2019-09-27 14:34:59.909')
,('THCS Lý Chinh Thắng 1','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:59.910','2019-09-27 14:34:59.910')
,('THCS Nguyễn An Khương','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:59.911','2019-09-27 14:34:59.911')
,('THCS Nguyễn Hồng Đào','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:59.911','2019-09-27 14:34:59.911')
,('THCS Phan Công Hớn','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:59.912','2019-09-27 14:34:59.912')
,('THCS Tam Đông 1','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:59.913','2019-09-27 14:34:59.913')
,('THCS Tân Xuân','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:59.914','2019-09-27 14:34:59.914')
,('THCS Thị Trấn','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:59.915','2019-09-27 14:34:59.915')
,('THCS Tô Ký','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:59.916','2019-09-27 14:34:59.916')
,('THCS Trung Mỹ Tây 1','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:59.917','2019-09-27 14:34:59.917')
,('THCS Xuân Thới Thượng','COUNTRY_VN',1,24,NULL,true,'2019-09-27 14:34:59.918','2019-09-27 14:34:59.918')
,('THCS Hai Bà Trưng','COUNTRY_VN',1,22,NULL,true,'2019-09-27 14:34:59.918','2019-09-27 14:34:59.918')
,('THCS Hiệp Phước','COUNTRY_VN',1,22,NULL,true,'2019-09-27 14:34:59.919','2019-09-27 14:34:59.919')
,('THCS Lê Văn Hưu','COUNTRY_VN',1,22,NULL,true,'2019-09-27 14:34:59.919','2019-09-27 14:34:59.919')
,('THCS Nguyễn Bỉnh Khiêm','COUNTRY_VN',1,22,NULL,true,'2019-09-27 14:34:59.920','2019-09-27 14:34:59.920')
,('THCS Nguyễn Văn Quỳ','COUNTRY_VN',1,22,NULL,true,'2019-09-27 14:34:59.920','2019-09-27 14:34:59.920')
,('THCS Phước Lộc','COUNTRY_VN',1,22,NULL,true,'2019-09-27 14:34:59.921','2019-09-27 14:34:59.921')
,('THCS Lê Thành Công','COUNTRY_VN',1,22,NULL,true,'2019-09-27 14:34:59.921','2019-09-27 14:34:59.921')
,('THCS Nguyễn Thị Hương','COUNTRY_VN',1,22,NULL,true,'2019-09-27 14:34:59.922','2019-09-27 14:34:59.922')
,('TH-THCS-THPT Ngân Hà','COUNTRY_VN',1,22,NULL,true,'2019-09-27 14:34:59.922','2019-09-27 14:34:59.922')
,('THCS Cầu Kiệu','COUNTRY_VN',1,17,NULL,true,'2019-09-27 14:34:59.923','2019-09-27 14:34:59.923')
,('THCS Châu Văn Liêm','COUNTRY_VN',1,17,NULL,true,'2019-09-27 14:34:59.924','2019-09-27 14:34:59.924')
,('THCS Độc Lập','COUNTRY_VN',1,17,NULL,true,'2019-09-27 14:34:59.926','2019-09-27 14:34:59.926')
,('THCS Ngô Tất Tố','COUNTRY_VN',1,17,NULL,true,'2019-09-27 14:34:59.926','2019-09-27 14:34:59.926')
,('THCS Trần Huy Liệu','COUNTRY_VN',1,17,NULL,true,'2019-09-27 14:34:59.927','2019-09-27 14:34:59.927')
,('THCS Đào Duy Anh','COUNTRY_VN',1,17,NULL,true,'2019-09-27 14:34:59.928','2019-09-27 14:34:59.928')
,('TH-THCS-THPT Quốc Tế','COUNTRY_VN',1,17,NULL,true,'2019-09-27 14:34:59.930','2019-09-27 14:34:59.930')
,('THCS-THPT Hồng Hà','COUNTRY_VN',1,17,NULL,true,'2019-09-27 14:34:59.931','2019-09-27 14:34:59.931')
,('THCS-THPT Việt Mỹ','COUNTRY_VN',1,17,NULL,true,'2019-09-27 14:34:57.141','2019-09-27 14:34:59.932')
,('THCS-THPT Việt Anh','COUNTRY_VN',1,17,NULL,true,'2019-09-27 14:34:57.144','2019-09-27 14:34:59.934')
,('THCS Hoàng Hoa Thám','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:59.934','2019-09-27 14:34:59.934')
,('THCS Lý Thường Kiệt','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:59.935','2019-09-27 14:34:59.935')
,('THCS Ngô Quyền','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:59.936','2019-09-27 14:34:59.936')
,('THCS Ngô Sĩ Liên','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:59.936','2019-09-27 14:34:59.936')
,('THCS Nguyễn Gia Thiều','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:59.936','2019-09-27 14:34:59.936')
,('THCS Phạm Ngọc Thạch','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:59.937','2019-09-27 14:34:59.937')
,('THCS Quang Trung','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:59.937','2019-09-27 14:34:59.937')
,('THCS Tân Bình','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:59.938','2019-09-27 14:34:59.938')
,('THCS Âu Lạc','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:59.938','2019-09-27 14:34:59.938')
,('THCS Trần Văn Đang','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:59.938','2019-09-27 14:34:59.938')
,('THCS Trường Chinh','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:59.939','2019-09-27 14:34:59.939')
,('THCS Võ Văn Tần','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:59.939','2019-09-27 14:34:59.939')
,('THCS Trần Văn Quang','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:59.939','2019-09-27 14:34:59.939')
,('THCS-THPT Việt Thanh','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:57.135','2019-09-27 14:34:59.940')
,('THCS-THPT Thái Bình','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:57.136','2019-09-27 14:34:59.940')
,('TH-THCS-THPT Thanh Bình','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:59.943','2019-09-27 14:34:59.943')
,('THCS-THPT Nguyễn Khuyến','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:57.135','2019-09-27 14:34:59.944')
,('THCS-THPT Bác Ái','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:57.139','2019-09-27 14:34:59.945')
,('THCS-THPT Hai Bà Trưng','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:59.946','2019-09-27 14:34:59.946')
,('TH-THCS-THPT Thái Bình Dương','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:57.138','2019-09-27 14:34:59.946')
,('THCS-THPT Văn hóa Việt','COUNTRY_VN',1,16,NULL,true,'2019-09-27 14:34:59.948','2019-09-27 14:34:59.948')
,('THCS Đặng Trần Côn','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.949','2019-09-27 14:34:59.949')
,('THCS Đồng Khởi','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.950','2019-09-27 14:34:59.950')
,('THCS Hoàng Diệu','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.950','2019-09-27 14:34:59.950')
,('THCS Hùng Vương','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.950','2019-09-27 14:34:59.950')
,('THCS Lê Anh Xuân','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.951','2019-09-27 14:34:59.951')
,('THCS Lê Lợi','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.952','2019-09-27 14:34:59.952')
,('THCS Nguyễn Huệ','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.952','2019-09-27 14:34:59.952')
,('THCS Phan Bội Châu','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.952','2019-09-27 14:34:59.952')
,('THCS Tân Thới Hòa','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.953','2019-09-27 14:34:59.953')
,('THCS Thoại Ngọc Hầu','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.953','2019-09-27 14:34:59.953')
,('THCS Trần Quang Khải','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.953','2019-09-27 14:34:59.953')
,('THCS Võ Thành Trang','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.954','2019-09-27 14:34:59.954')
,('THCS Tôn Thất Tùng','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.954','2019-09-27 14:34:59.954')
,('TH-THCS Hồng Ngọc','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.955','2019-09-27 14:34:59.955')
,('TH-THCS-THPT Quốc Văn Sài Gòn','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.955','2019-09-27 14:34:59.955')
,('THCS-THPT Hồng Đức','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.956','2019-09-27 14:34:59.956')
,('THCS-THPT Trí Đức','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.956','2019-09-27 14:34:59.956')
,('THCS-THPT Nhân Văn','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.958','2019-09-27 14:34:59.958')
,('THCS-THPT Khai Minh','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:57.169','2019-09-27 14:34:59.959')
,('THCS-THPT Tân Phú','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:57.165','2019-09-27 14:34:59.960')
,('THCS-THPT Đinh Tiên Hoàng','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.961','2019-09-27 14:34:59.961')
,('TH-THCS-THPT Hòa Bình','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.962','2019-09-27 14:34:59.962')
,('THCS-THPT Nam Việt','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.964','2019-09-27 14:34:59.964')
,('THCS-THPT Trần Cao Vân','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.965','2019-09-27 14:34:59.965')
,('TH, THCS và THPT Lê Thánh Tông','COUNTRY_VN',1,19,NULL,true,'2019-09-27 14:34:59.965','2019-09-27 14:34:59.965')
,('THCS Bình Chiểu','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:59.966','2019-09-27 14:34:59.966')
,('THCS Bình Thọ','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:59.966','2019-09-27 14:34:59.966')
,('THCS Hiệp Bình','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:59.967','2019-09-27 14:34:59.967')
,('THCS Lê Quý Đôn','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:59.967','2019-09-27 14:34:59.967')
,('THCS Lê Văn Việt','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:59.968','2019-09-27 14:34:59.968')
,('THCS Linh Đông','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:59.968','2019-09-27 14:34:59.968')
,('THCS Linh Trung','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:59.968','2019-09-27 14:34:59.968')
,('THCS Ngô Chí Quốc','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:59.969','2019-09-27 14:34:59.969')
,('THCS Nguyễn Văn Bá','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:59.969','2019-09-27 14:34:59.969')
,('THCS Tam Bình','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:59.970','2019-09-27 14:34:59.970')
,('THCS Thái Văn Lung','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:59.970','2019-09-27 14:34:59.970')
,('THCS Trường Thọ','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:59.970','2019-09-27 14:34:59.970')
,('THCS Trương Văn Ngư','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:59.971','2019-09-27 14:34:59.971')
,('THCS Xuân Trường','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:59.971','2019-09-27 14:34:59.971')
,('THCS Dương Văn Thì','COUNTRY_VN',1,18,NULL,true,'2019-09-27 14:34:59.972','2019-09-27 14:34:59.972')
;


INSERT INTO public.hubs ("name",description,phone_number,address,country,city_id,district_id,point,images,opening_hours,created_at,updated_at,events) VALUES
('Manabie Lý Thường Kiệt',NULL,'0916 501 517','373/3C-D Lý Thường Kiệt, Phường 9, Quận Tân Bình, Thành phố Hồ Chí Minh, Việt Nam','COUNTRY_VN',1,16,POINT(10.7805731,106.6530848),'{https://storage.googleapis.com/manabie-content/hubs/h1/1.jpg,https://storage.googleapis.com/manabie-content/hubs/h1/2.jpg,https://storage.googleapis.com/manabie-content/hubs/h1/3.jpg,https://storage.googleapis.com/manabie-content/hubs/h1/4.jpg,https://storage.googleapis.com/manabie-content/hubs/h1/5.jpg,https://storage.googleapis.com/manabie-content/hubs/h1/6.jpg,https://storage.googleapis.com/manabie-content/hubs/h1/7.jpg}','{Mon-Fri 12:00 - 21:00,Sat-Sun 8:00 - 21:00}','2019-10-02 09:14:00.000','2019-10-02 09:14:00.000',NULL)
,('Manabie Hoàng Hoa Thám',NULL,'0916 211 517','7 Hoàng Hoa Thám, Phường 13, Quận Tân Bình, Thành phố Hồ Chí Minh, Việt Nam','COUNTRY_VN',1,16,POINT(10.797163,106.6446453),'{https://storage.googleapis.com/manabie-content/hubs/h2/1.jpg,https://storage.googleapis.com/manabie-content/hubs/h2/2.jpg,https://storage.googleapis.com/manabie-content/hubs/h2/3.jpg,https://storage.googleapis.com/manabie-content/hubs/h2/4.jpg}','{Mon-Fri 12:00 - 21:00,Sat-Sun 8:00 - 21:00}','2019-10-02 09:14:00.000','2019-10-02 09:14:00.000',NULL)
,('Manabie Lê Đức Thọ',NULL,'0919 041 315','208 Lê Đức Thọ, Phường 6, Quận Gò Vấp, Thành phố Hồ Chí Minh, Việt Nam','COUNTRY_VN',1,15,POINT(10.8454481,106.6695028),'{https://storage.googleapis.com/manabie-content/hubs/h3/1.jpg,https://storage.googleapis.com/manabie-content/hubs/h3/2.jpg,https://storage.googleapis.com/manabie-content/hubs/h3/3.jpg,https://storage.googleapis.com/manabie-content/hubs/h3/4.jpg}','{Mon-Fri 12:00 - 21:00,Sat-Sun 8:00 - 21:00}','2019-10-02 09:14:00.000','2019-10-02 09:14:00.000',NULL)
,('Manabie Nguyễn Huy Tự',NULL,'0914 501 315','29 Nguyễn Huy Tự, Phường Đa Kao, Quận 1, Thành phố Hồ Chí Minh, Việt Nam','COUNTRY_VN',1,1,POINT(10.7925148,106.6942215),NULL,'{Mon-Fri 12:00 - 21:00,Sat-Sun 8:00 - 21:00}','2019-10-02 09:14:00.000','2019-10-02 09:14:00.000',NULL)
,('Manabie Dương Đình Nghệ',NULL,'0916 851 517','19A Dương Đình Nghệ, Phường 8, Quận 11, Thành phố Hồ Chí Minh, Việt Nam','COUNTRY_VN',1,11,POINT(10.7603188,106.6475329),NULL,'{Mon-Fri 12:00 - 21:00,Sat-Sun: 8;00 - 21:00}','2019-10-02 09:14:00.000','2019-10-02 09:14:00.000',NULL)
;

INSERT INTO configs VALUES
        ('secureHash', 'payment', 'COUNTRY_ID', 'DpBOSuGh9qUuxVIGrIN755d34bMgqLf9', now(), now()),
        ('endpoint', 'payment', 'COUNTRY_ID', 'https://test.paydollar.com/b2cDemo/eng/payment/payForm.jsp', now(), now()),
        ('merchantID', 'payment', 'COUNTRY_ID', '74001088', now(), now()),
        ('currCode', 'payment', 'COUNTRY_ID', '360', now(), now()),
        ('minimumAmount', 'payment', 'COUNTRY_ID', '50000', now(), now()),
        ('successUrl', 'payment', 'COUNTRY_ID', 'https://student-coach-e1e95.firebaseapp.com/response/success', now(), now()),
        ('failUrl', 'payment', 'COUNTRY_ID', 'https://student-coach-e1e95.firebaseapp.com/response/fail', now(), now()),
        ('cancelUrl', 'payment', 'COUNTRY_ID', 'https://student-coach-e1e95.firebaseapp.com/response/cancel', now(), now()),
        ('interfaceLang', 'payment', 'COUNTRY_ID', 'V', now(), now()),
        ('trialPeriod', 'payment', 'COUNTRY_ID', '30', now(), now())
        ON CONFLICT DO NOTHING;

INSERT INTO configs VALUES
        ('secureHash', 'payment', 'COUNTRY_SG', 'DpBOSuGh9qUuxVIGrIN755d34bMgqLf9', now(), now()),
        ('endpoint', 'payment', 'COUNTRY_SG', 'https://test.paydollar.com/b2cDemo/eng/payment/payForm.jsp', now(), now()),
        ('merchantID', 'payment', 'COUNTRY_SG', '74001088', now(), now()),
        ('currCode', 'payment', 'COUNTRY_SG', '702', now(), now()),
        ('minimumAmount', 'payment', 'COUNTRY_SG', '50000', now(), now()),
        ('successUrl', 'payment', 'COUNTRY_SG', 'https://student-coach-e1e95.firebaseapp.com/response/success', now(), now()),
        ('failUrl', 'payment', 'COUNTRY_SG', 'https://student-coach-e1e95.firebaseapp.com/response/fail', now(), now()),
        ('cancelUrl', 'payment', 'COUNTRY_SG', 'https://student-coach-e1e95.firebaseapp.com/response/cancel', now(), now()),
        ('interfaceLang', 'payment', 'COUNTRY_SG', 'V', now(), now()),
        ('trialPeriod', 'payment', 'COUNTRY_SG', '30', now(), now())
        ON CONFLICT DO NOTHING;

INSERT INTO configs VALUES
        ('secureHash', 'payment', 'COUNTRY_JP', 'DpBOSuGh9qUuxVIGrIN755d34bMgqLf9', now(), now()),
        ('endpoint', 'payment', 'COUNTRY_JP', 'https://test.paydollar.com/b2cDemo/eng/payment/payForm.jsp', now(), now()),
        ('merchantID', 'payment', 'COUNTRY_JP', '74001088', now(), now()),
        ('currCode', 'payment', 'COUNTRY_JP', '392', now(), now()),
        ('minimumAmount', 'payment', 'COUNTRY_JP', '50000', now(), now()),
        ('successUrl', 'payment', 'COUNTRY_JP', 'https://student-coach-e1e95.firebaseapp.com/response/success', now(), now()),
        ('failUrl', 'payment', 'COUNTRY_JP', 'https://student-coach-e1e95.firebaseapp.com/response/fail', now(), now()),
        ('cancelUrl', 'payment', 'COUNTRY_JP', 'https://student-coach-e1e95.firebaseapp.com/response/cancel', now(), now()),
        ('interfaceLang', 'payment', 'COUNTRY_JP', 'V', now(), now()),
        ('trialPeriod', 'payment', 'COUNTRY_JP', '30', now(), now())
        ON CONFLICT DO NOTHING;

INSERT INTO configs VALUES
        ('secureHash', 'payment', 'COUNTRY_MASTER', 'DpBOSuGh9qUuxVIGrIN755d34bMgqLf9', now(), now()),
        ('endpoint', 'payment', 'COUNTRY_MASTER', 'https://test.paydollar.com/b2cDemo/eng/payment/payForm.jsp', now(), now()),
        ('merchantID', 'payment', 'COUNTRY_MASTER', '74001088', now(), now()),
        ('currCode', 'payment', 'COUNTRY_MASTER', '392', now(), now()),
        ('minimumAmount', 'payment', 'COUNTRY_MASTER', '50000', now(), now()),
        ('successUrl', 'payment', 'COUNTRY_MASTER', 'https://student-coach-e1e95.firebaseapp.com/response/success', now(), now()),
        ('failUrl', 'payment', 'COUNTRY_MASTER', 'https://student-coach-e1e95.firebaseapp.com/response/fail', now(), now()),
        ('cancelUrl', 'payment', 'COUNTRY_MASTER', 'https://student-coach-e1e95.firebaseapp.com/response/cancel', now(), now()),
        ('interfaceLang', 'payment', 'COUNTRY_MASTER', 'V', now(), now()),
        ('trialPeriod', 'payment', 'COUNTRY_MASTER', '30', now(), now())
        ON CONFLICT DO NOTHING;


INSERT INTO public."plans"
(plan_id, country, description, plan_privileges, is_purchasable, prioritize_level, created_at, updated_at, benefits)
VALUES('Trial', 'COUNTRY_JP', 'Trial plans', '{}', false, 0, now(), now(), '{test benefits 1,test benefits 2,test benefits 3}');
INSERT INTO public."plans"
(plan_id, country, description, plan_privileges, is_purchasable, prioritize_level, created_at, updated_at, benefits)
VALUES('Expired', 'COUNTRY_JP', 'place holder package', '{}', false, 0, now(), now(), '{test benefits 1,test benefits 2,test benefits 3}');
INSERT INTO public."plans"
(plan_id, country, description, plan_privileges, is_purchasable, prioritize_level, created_at, updated_at, benefits)
VALUES('School', 'COUNTRY_JP', NULL, '{}', false, 1, now(), now(), NULL);

INSERT INTO public."plans"
(plan_id, country, description, plan_privileges, is_purchasable, prioritize_level, created_at, updated_at, benefits)
VALUES('Trial', 'COUNTRY_SG', 'Trial plans', '{}', false, 0, now(), now(), '{test benefits 1,test benefits 2,test benefits 3}');
INSERT INTO public."plans"
(plan_id, country, description, plan_privileges, is_purchasable, prioritize_level, created_at, updated_at, benefits)
VALUES('Expired', 'COUNTRY_SG', 'place holder package', '{}', false, 0, now(), now(), '{test benefits 1,test benefits 2,test benefits 3}');
INSERT INTO public."plans"
(plan_id, country, description, plan_privileges, is_purchasable, prioritize_level, created_at, updated_at, benefits)
VALUES('School', 'COUNTRY_SG', NULL, '{}', false, 1, now(), now(), NULL);

INSERT INTO public."plans"
(plan_id, country, description, plan_privileges, is_purchasable, prioritize_level, created_at, updated_at, benefits)
VALUES('Trial', 'COUNTRY_ID', 'Trial plans', '{}', false, 0, now(), now(), '{test benefits 1,test benefits 2,test benefits 3}');
INSERT INTO public."plans"
(plan_id, country, description, plan_privileges, is_purchasable, prioritize_level, created_at, updated_at, benefits)
VALUES('Expired', 'COUNTRY_ID', 'place holder package', '{}', false, 0, now(), now(), '{test benefits 1,test benefits 2,test benefits 3}');
INSERT INTO public."plans"
(plan_id, country, description, plan_privileges, is_purchasable, prioritize_level, created_at, updated_at, benefits)
VALUES('School', 'COUNTRY_ID', NULL, '{}', false, 1, now(), now(), NULL);

INSERT INTO public."plans"
(plan_id, country, description, plan_privileges, is_purchasable, prioritize_level, created_at, updated_at, benefits)
VALUES('Trial', 'COUNTRY_MASTER', 'Trial plans', '{}', false, 0, now(), now(), '{test benefits 1,test benefits 2,test benefits 3}');
INSERT INTO public."plans"
(plan_id, country, description, plan_privileges, is_purchasable, prioritize_level, created_at, updated_at, benefits)
VALUES('Expired', 'COUNTRY_MASTER', 'place holder package', '{}', false, 0, now(), now(), '{test benefits 1,test benefits 2,test benefits 3}');
INSERT INTO public."plans"
(plan_id, country, description, plan_privileges, is_purchasable, prioritize_level, created_at, updated_at, benefits)
VALUES('School', 'COUNTRY_MASTER', NULL, '{}', false, 1, now(), now(), NULL);

INSERT INTO configs VALUES
        ('iosBundleID', 'iap', 'COUNTRY_VN', 'com.manabie.ios', now(), now())
        ON CONFLICT DO NOTHING;


INSERT INTO configs VALUES
        ('1', 'class_avatar', 'COUNTRY_MASTER', 'https://storage.googleapis.com/manabie-backend/class/ico_class_default.png', now(), now()),
        ('2', 'class_avatar', 'COUNTRY_MASTER', 'https://storage.googleapis.com/manabie-backend/class/ico_class_default_2.png', now(), now()),
        ('planName', 'class_plan', 'COUNTRY_VN', 'School', now(), now()),
        ('planPeriod', 'class_plan', 'COUNTRY_VN', '2020-06-30', now(), now())
        ON CONFLICT DO NOTHING;

INSERT INTO configs VALUES
        ('contactName', 'ghn', 'COUNTRY_VN', 'Manabie VN', now(), now()),
        ('contactPhone', 'ghn', 'COUNTRY_VN', '012345789', now(), now()),
        ('contactAddress', 'ghn', 'COUNTRY_VN', 'TNR Tower 15th Floor, 180-192 Nguyen Cong Tru Str., 1st Dist., Ho Chi Minh City, Vietnam', now(), now()),
        ('district', 'ghn', 'COUNTRY_VN', '1442', now(), now()),
        ('wardCode', 'ghn', 'COUNTRY_VN', '20108', now(), now()),
        ('externalReturnCode', 'ghn', 'COUNTRY_VN', 'GHN', now(), now()),
        ('baseUrl', 'ghn', 'COUNTRY_VN', 'http://gandalf:5889', now(), now()),
        ('token', 'ghn', 'COUNTRY_VN', '3a4533b4322b41ba99d1aae2ae31a85d', now(), now())
        ON CONFLICT DO NOTHING;

INSERT INTO cities (city_id, "name", country, created_at, updated_at, display_order)
VALUES(-2147483648, 'Manabie City', 'COUNTRY_MASTER', now(), now(), 0) ON CONFLICT DO NOTHING;

INSERT INTO districts
(district_id, "name", country, city_id, created_at, updated_at)
VALUES(-2147483648, 'Manabie District', 'COUNTRY_MASTER', -2147483648, now(), now()) ON CONFLICT DO NOTHING;

INSERT INTO public."plans"
(plan_id, description, country, plan_privileges, created_at, updated_at, is_purchasable, prioritize_level, benefits)
VALUES('School', NULL, 'COUNTRY_VN', '{}', now(), now(), false, 1, NULL)
ON CONFLICT DO NOTHING;

INSERT INTO configs VALUES
        ('secureHash', 'payment', 'COUNTRY_VN', 'DpBOSuGh9qUuxVIGrIN755d34bMgqLf9', now(), now()),
        ('endpoint', 'payment', 'COUNTRY_VN', 'https://test.paydollar.com/b2cDemo/eng/payment/payForm.jsp', now(), now()),
        ('merchantID', 'payment', 'COUNTRY_VN', '74001088', now(), now()),
        ('currCode', 'payment', 'COUNTRY_VN', '704', now(), now()),
        ('minimumAmount', 'payment', 'COUNTRY_VN', '50000', now(), now()),
        ('successUrl', 'payment', 'COUNTRY_VN', 'https://student-coach-e1e95.firebaseapp.com/response/success', now(), now()),
        ('failUrl', 'payment', 'COUNTRY_VN', 'https://student-coach-e1e95.firebaseapp.com/response/fail', now(), now()),
        ('cancelUrl', 'payment', 'COUNTRY_VN', 'https://student-coach-e1e95.firebaseapp.com/response/cancel', now(), now()),
        ('interfaceLang', 'payment', 'COUNTRY_VN', 'V', now(), now()),
        ('trialPeriod', 'payment', 'COUNTRY_VN', '30', now(), now())
        ON CONFLICT DO NOTHING;

INSERT INTO public.notification_messages (country,"key",receiver_group,title,body,created_at,updated_at) VALUES
('COUNTRY_VN','QUESTION_TRANSITION_ASSIGN','USER_GROUP_TUTOR','[G{{.StudentGrade}}] {{.StudentName}}','Chào bạn, bạn có thể giải thích câu hỏi này được không?','2019-12-22 06:41:24.489','2019-12-22 06:41:24.489')
,('COUNTRY_VN','QUESTION_TRANSITION_DISAGREE_RESOLVED','USER_GROUP_TUTOR','[G{{.StudentGrade}}] {{.StudentName}} vẫn chưa hiểu câu trả lời của bạn','','2019-12-22 06:41:24.500','2019-12-22 06:41:24.500')
,('COUNTRY_VN','COACH_AUTO_EVENT_FINISH_FIRST_LO','USER_GROUP_STUDENT','','Chào bạn trẻ,
Chúc mừng em vừa hoàn thành bài học đầu tiên trên ứng dụng Manabie! [emoji][emoji][emoji]
Để đánh dấu cột mốc này, đội Cố vấn học tập của Manabie xin dành tặng cho em một món quà cực hấp dẫn: [emoji][emoji] 40% ƯU ĐÃI áp dụng cho TẤT CẢ các gói học tại Manabie - sử dụng bằng cách nhập ngay MÃ KHUYẾN MÃI XXXXXX.
Hãy dùng món quà này để khám phá thêm các video cực "kool" của Manabie và xóa bỏ những "điểm mù" với kiến thức của các bài học trên lớp nhé.
Mã Khuyến mãi chỉ có hiệu lực tới ngày 26/12 thôi, nhanh chân lên em!!
Mã Khuyến mãi: XXXXXX',now(),now())
,('COUNTRY_VN','COACH_AUTO_EVENT_FINISH_FIRST_TOPIC','USER_GROUP_STUDENT','','Chào bạn trẻ,
Chúc mừng em vừa hoàn thành bài học đầu tiên trên ứng dụng Manabie! [emoji][emoji][emoji]
Để đánh dấu cột mốc này, đội Cố vấn học tập của Manabie xin dành tặng cho em một món quà cực hấp dẫn: [emoji][emoji] 40% ƯU ĐÃI áp dụng cho TẤT CẢ các gói học tại Manabie - sử dụng bằng cách nhập ngay MÃ KHUYẾN MÃI XXXXXX.
Hãy dùng món quà này để khám phá thêm các video cực "kool" của Manabie và xóa bỏ những "điểm mù" với kiến thức của các bài học trên lớp nhé.
Mã Khuyến mãi chỉ có hiệu lực tới ngày 26/12 thôi, nhanh chân lên em!!
Mã Khuyến mãi: XXXXXX',now(),now())
,('COUNTRY_MASTER','QUESTION_TRANSITION_ASSIGN','USER_GROUP_TUTOR','[G{{.StudentGrade}}] {{.StudentName}}','Hi, could you please help me with this question?',now(),now())
,('COUNTRY_MASTER','QUESTION_TRANSITION_DISAGREE_RESOLVED','USER_GROUP_TUTOR','[G{{.StudentGrade}}] {{.StudentName}} still has a doubt','',now(),now())
,('COUNTRY_MASTER','COACH_AUTO_EVENT_FINISH_FIRST_LO','USER_GROUP_STUDENT','','Hi there,
Congratulations on finishing your first lesson on Manabie app! [emoji][emoji][emoji]
I am a personal coach at Manabie.
To celebrate our very first milestone, we have this gift specially for you: [emoji][emoji] 40% DISCOUNT on ALL Manabie premium packages by applying the code XXXXXX on payment screen.
Let''s grab this chance to explore our cool videos and let our subject master help you achieve the academic excellence
Only limited code is given till December 26 so hurry up!!
Promo code: XXXXXX',now(),now())
,('COUNTRY_MASTER','COACH_AUTO_EVENT_FINISH_FIRST_TOPIC','USER_GROUP_STUDENT','','Hi there,
Congratulations on finishing your first lesson on Manabie app! [emoji][emoji][emoji]
I am a personal coach at Manabie.
To celebrate our very first milestone, we have this gift specially for you: [emoji][emoji] 40% DISCOUNT on ALL Manabie premium packages by applying the code XXXXXX on payment screen.
Let''s grab this chance to explore our cool videos and let our subject master help you achieve the academic excellence
Only limited code is given till December 26 so hurry up!!
Promo code: XXXXXX',now(),now())
;

INSERT INTO public.notification_messages (country,"key",receiver_group,title,body,created_at,updated_at) VALUES
('COUNTRY_VN','STUDENT_FINISH_FIRST_THREE_LO_EVENT','USER_GROUP_STUDENT','🎁 Em được gửi tặng 1 món quà đặc biệt! 🎁','Chúc mừng em vừa hoàn thành bài học đầu tiên trên Manabie! 🎉🎉🎉 Để đánh dấu cột mốc này, Manabie xin gửi tặng em 1 mã khuyến mãi [GIẢM GIÁ 30%] khi mua bất kì gói học nào. Hãy nhập mã [{{.PromotionCode}}] khi thanh toán gói học để được khấu trừ và tiếp tục đồng hành cùng Manabie nha!',now(),now())
,('COUNTRY_MASTER','STUDENT_FINISH_FIRST_THREE_LO_EVENT','USER_GROUP_STUDENT','🎁 You''ve got a special gift! 🎁','You''ve just completed 1ST LESSON with Manabie, congratulations! 🎉🎉🎉 A special gift voucher of [30% DISCOUNT] on any package plans is given to you to celebrate this 1st milestone. Let''s use the code [{{.PromotionCode}}] to claim your discount and achieve much more together!',now(),now());

INSERT INTO public.schools
(school_id, name, country, city_id, district_id, point, is_system_school, created_at, updated_at, is_merge)
VALUES(-2147483648, 'Manabie School', 'COUNTRY_MASTER', 1, 1, NULL, false, now(), now(), false);

INSERT INTO configs VALUES
        ('default_privileges', 'school', 'COUNTRY_VN', 'CAN_ACCESS_LEARNING_TOPICS,CAN_ACCESS_PRACTICE_TOPICS,CAN_ACCESS_ALL_LOS,CAN_WATCH_VIDEOS,CAN_READ_STUDY_GUIDES,CAN_CHAT_WITH_TEACHER', now(), now())
        ON CONFLICT DO NOTHING;

ALTER TABLE ONLY public.classes DROP CONSTRAINT classes_subjects_check;

ALTER TABLE public.classes ADD CHECK (subjects <@ ARRAY[
                'SUBJECT_MATHS',
                'SUBJECT_BIOLOGY',
                'SUBJECT_PHYSICS',
                'SUBJECT_CHEMISTRY',
                'SUBJECT_GEOGRAPHY',
                'SUBJECT_ENGLISH',
                'SUBJECT_ENGLISH_2',
                'SUBJECT_JAPANESE',
                'SUBJECT_SCIENCE',
                'SUBJECT_SOCIAL_STUDIES'
            ]);
