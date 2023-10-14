ALTER TABLE ONLY public.bank
  DROP CONSTRAINT IF EXISTS bank__bank_code__unique;

ALTER TABLE ONLY public.bank_branch
  DROP CONSTRAINT IF EXISTS bank_branch__bank_branch_code__unique;