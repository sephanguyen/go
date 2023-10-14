ALTER TABLE ONLY "e2e_instances" ADD COLUMN IF NOT EXISTS "squad_tags" TEXT[] DEFAULT '{}';
