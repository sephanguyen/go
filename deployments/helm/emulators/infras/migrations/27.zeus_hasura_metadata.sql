\connect zeus_hasura_metadata;

GRANT USAGE ON SCHEMA public TO zeus_hasura;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO zeus_hasura;
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO zeus_hasura;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO zeus_hasura;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT SELECT, INSERT, UPDATE ON TABLES TO zeus_hasura;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO zeus_hasura;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO zeus_hasura;
