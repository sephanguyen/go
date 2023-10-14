CREATE SEQUENCE public.applied_slot_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.applied_slot_id_seq OWNED BY public.applied_slot.id;
ALTER TABLE ONLY public.applied_slot ALTER COLUMN id SET DEFAULT nextval('public.applied_slot_id_seq'::regclass);

CREATE SEQUENCE public.center_opening_slot_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.center_opening_slot_id_seq OWNED BY public.center_opening_slot.id;
ALTER TABLE ONLY public.center_opening_slot ALTER COLUMN id SET DEFAULT nextval('public.center_opening_slot_id_seq'::regclass);

CREATE SEQUENCE public.student_available_slot_master_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.student_available_slot_master_id_seq OWNED BY public.student_available_slot_master.id;
ALTER TABLE ONLY public.student_available_slot_master ALTER COLUMN id SET DEFAULT nextval('public.student_available_slot_master_id_seq'::regclass);

CREATE SEQUENCE public.teacher_available_slot_master_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.teacher_available_slot_master_id_seq OWNED BY public.teacher_available_slot_master.id;
ALTER TABLE ONLY public.teacher_available_slot_master ALTER COLUMN id SET DEFAULT nextval('public.teacher_available_slot_master_id_seq'::regclass);

CREATE SEQUENCE public.teacher_subject_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.teacher_subject_id_seq OWNED BY public.teacher_subject.id;
ALTER TABLE ONLY public.teacher_subject ALTER COLUMN id SET DEFAULT nextval('public.teacher_subject_id_seq'::regclass);

CREATE SEQUENCE public.time_slot_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.time_slot_id_seq OWNED BY public.time_slot.id;
ALTER TABLE ONLY public.time_slot ALTER COLUMN id SET DEFAULT nextval('public.time_slot_id_seq'::regclass);
