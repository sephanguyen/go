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


--
-- Name: conversation_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.conversation_status AS ENUM (
    'CONVERSATION_STATUS_NONE',
    'CONVERSATION_STATUS_CLOSE'
);


--
-- Name: message_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.message_type AS ENUM (
    'MESSAGE_TYPE_TEXT',
    'MESSAGE_TYPE_IMAGE',
    'MESSAGE_TYPE_VIDEO',
    'MESSAGE_TYPE_SYSTEM',
    'MESSAGE_TYPE_BUTTON',
    'MESSAGE_TYPE_COACH_AUTO'
);


--
-- Name: reset_student_account(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.reset_student_account(u_id text) RETURNS text
    LANGUAGE plpgsql
    AS $$
declare
	c_ids uuid[];
begin
	select array(select conversation_id from conversation_statuses where user_id = u_id) into c_ids;
 	if (c_ids = '{}')then
 		return 'conversation for user not found';
 	end if;

 	delete from user_device_tokens where user_id = u_id;
	delete from conversation_statuses where conversation_id = any(c_ids);
	delete from messages where conversation_id = any(c_ids);
	delete from conversations where student_id = u_id;
	DELETE FROM conversations WHERE conversation_id IN (
		SELECT c.conversation_id
	FROM conversations AS c LEFT JOIN conversation_statuses AS s ON c.conversation_id = s.conversation_id
	WHERE s.conversation_id IS NULL
	);

 	return 'success';
END;
$$;


SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: conversation_statuses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.conversation_statuses (
    conversation_statuses_id uuid NOT NULL,
    user_id text NOT NULL,
    conversation_id uuid NOT NULL,
    seen_at timestamp with time zone,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    last_notify_at timestamp with time zone,
    role text,
    status text,
    CONSTRAINT conversation_statuses_role_check CHECK ((role = ANY ('{USER_GROUP_STUDENT,USER_GROUP_COACH,USER_GROUP_TUTOR,USER_GROUP_TEACHER,USER_GROUP_PARENT}'::text[]))),
    CONSTRAINT conversation_statuses_status_check CHECK ((status = ANY ('{CONVERSATION_STATUS_ACTIVE,CONVERSATION_STATUS_INACTIVE}'::text[])))
);


--
-- Name: COLUMN conversation_statuses.role; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.conversation_statuses.role IS 'role: USER_GROUP_STUDENT, USER_GROUP_COACH, USER_GROUP_TUTOR, USER_GROUP_TEACHER, USER_GROUP_PARENT ...';


--
-- Name: COLUMN conversation_statuses.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.conversation_statuses.status IS 'status: CONVERSATION_STATUS_ACTIVE, CONVERSATION_STATUS_INACTIVE';


--
-- Name: conversations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.conversations (
    conversation_id uuid NOT NULL,
    student_id text,
    coach_id text,
    guest_ids text[],
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    tutor_id text,
    student_question_id text,
    status text DEFAULT 'CONVERSATION_STATUS_NONE'::text,
    previous_coach_ids text[] DEFAULT '{}'::text[],
    class_id integer,
    name text,
    CONSTRAINT conversations_status_check CHECK ((status = ANY ('{CONVERSATION_STATUS_NONE,CONVERSATION_STATUS_CLOSE}'::text[])))
);


--
-- Name: messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.messages (
    message_id uuid NOT NULL,
    conversation_id uuid NOT NULL,
    user_id text,
    message text,
    url_media text,
    type text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT messages_type_check CHECK ((type = ANY ('{MESSAGE_TYPE_TEXT,MESSAGE_TYPE_IMAGE,MESSAGE_TYPE_VIDEO,MESSAGE_TYPE_SYSTEM,MESSAGE_TYPE_BUTTON,MESSAGE_TYPE_COACH_AUTO}'::text[])))
);


--
-- Name: online_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.online_users (
    user_id text NOT NULL,
    node_name text NOT NULL,
    last_active_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: user_device_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_device_tokens (
    user_device_token_id integer NOT NULL,
    user_id text NOT NULL,
    token text,
    allow_notification boolean,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    user_name text
);


--
-- Name: user_device_tokens_user_device_token_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_device_tokens_user_device_token_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_device_tokens_user_device_token_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_device_tokens_user_device_token_id_seq OWNED BY public.user_device_tokens.user_device_token_id;


--
-- Name: user_device_tokens user_device_token_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_device_tokens ALTER COLUMN user_device_token_id SET DEFAULT nextval('public.user_device_tokens_user_device_token_id_seq'::regclass);


--
-- Name: conversation_statuses conversation_statuses__user_id__conversation_id_un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversation_statuses
    ADD CONSTRAINT conversation_statuses__user_id__conversation_id_un UNIQUE (user_id, conversation_id);


--
-- Name: conversation_statuses conversation_statuses_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversation_statuses
    ADD CONSTRAINT conversation_statuses_pk PRIMARY KEY (conversation_statuses_id);


--
-- Name: conversations conversations__student_id__coach_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversations
    ADD CONSTRAINT conversations__student_id__coach_id UNIQUE (student_id, coach_id);


--
-- Name: conversations conversations_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversations
    ADD CONSTRAINT conversations_pk PRIMARY KEY (conversation_id);


--
-- Name: messages messages_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pk PRIMARY KEY (message_id);


--
-- Name: online_users online_users_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.online_users
    ADD CONSTRAINT online_users_pk PRIMARY KEY (user_id, node_name);


--
-- Name: user_device_tokens user_device_tokens_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_device_tokens
    ADD CONSTRAINT user_device_tokens_pk PRIMARY KEY (user_device_token_id);


--
-- Name: user_device_tokens user_id_un; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_device_tokens
    ADD CONSTRAINT user_id_un UNIQUE (user_id);


--
-- Name: conversations_idx__guest_ids; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX conversations_idx__guest_ids ON public.conversations USING btree (guest_ids);


--
-- Name: conversation_statuses conversation_statuses__conversation_id__fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversation_statuses
    ADD CONSTRAINT conversation_statuses__conversation_id__fk FOREIGN KEY (conversation_id) REFERENCES public.conversations(conversation_id);


--
-- Name: messages messages__conversation_id__fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages__conversation_id__fk FOREIGN KEY (conversation_id) REFERENCES public.conversations(conversation_id);


--
-- PostgreSQL database dump complete
--

