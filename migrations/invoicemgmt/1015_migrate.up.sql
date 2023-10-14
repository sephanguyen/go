CREATE SEQUENCE public.payment_id_seq
    AS integer;

ALTER SEQUENCE public.payment_id_seq OWNED BY public.payment.payment_id;
ALTER TABLE ONLY public.payment ALTER COLUMN payment_id SET DEFAULT nextval('public.payment_id_seq'::regclass);
ALTER TABLE ONLY public.payment ALTER COLUMN payment_date DROP NOT NULL;