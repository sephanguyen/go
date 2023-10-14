CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

ALTER TABLE ONLY books ADD COLUMN IF NOT EXISTS copied_from text;

ALTER TABLE ONLY chapters ADD COLUMN IF NOT EXISTS copied_from text;

ALTER TABLE ONLY learning_objectives ADD COLUMN IF NOT EXISTS copied_from text;
