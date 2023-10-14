ALTER TABLE ONLY public.invoice_action_log
  DROP CONSTRAINT IF EXISTS invoice_action_log_pk;

ALTER TABLE public.invoice_action_log
  ALTER COLUMN invoice_action_id TYPE text;

ALTER TABLE ONLY public.invoice_action_log
  ADD CONSTRAINT invoice_action_log_pk PRIMARY KEY (invoice_action_id);