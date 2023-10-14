DROP INDEX IF EXISTS users__lower_email__idx;
CREATE INDEX users__lower_email__idx ON public.users (LOWER(email));