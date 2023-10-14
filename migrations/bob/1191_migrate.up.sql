ALTER TABLE ONLY public.organizations DROP CONSTRAINT IF EXISTS organization__domain_name__un;
ALTER TABLE public.organizations
    ADD CONSTRAINT organization__domain_name__un UNIQUE (domain_name);

ALTER TABLE ONLY public.organizations DROP CONSTRAINT IF EXISTS organization__domain_name__check;
ALTER TABLE public.organizations
    ADD CONSTRAINT organization__domain_name__check CHECK (domain_name ~ '^[^-\s].[a-z0-9-]*$');

-- Regex follow Validation rule:
--
--   Valid characters for hostnames are ASCII(7) letters from a to z, the digits from 0 to 9, and the hyphen (-). A hostname may not start with a hyphen.
--   hostname7 (https://man7.org/linux/man-pages/man7/hostname.7.html#:~:text=Valid%20characters%20for%20hostnames%20are,to%20an%20address%20for%20use)
-- Dissection ^[^_\s].[a-z0-9-]*$:
--   ^[^-\s]: not allow start with -
--   [a-z0-9-]*: accept only char inside range a-z, 0-9
