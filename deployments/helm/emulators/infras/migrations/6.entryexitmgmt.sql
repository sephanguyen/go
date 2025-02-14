\connect entryexitmgmt;

GRANT USAGE ON SCHEMA public TO entryexitmgmt;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO entryexitmgmt;
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO entryexitmgmt;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO entryexitmgmt;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO entryexitmgmt;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO entryexitmgmt;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO entryexitmgmt;

GRANT USAGE ON SCHEMA public TO hasura;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO hasura;
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO hasura;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO hasura;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT SELECT, INSERT, UPDATE ON TABLES TO hasura;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO hasura;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO hasura;

ALTER ROLE kafka_connector BYPASSRLS;
GRANT USAGE ON SCHEMA public TO kafka_connector;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO kafka_connector;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO kafka_connector;

GRANT USAGE ON SCHEMA public TO hephaestus;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO hephaestus;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT SELECT, INSERT, UPDATE ON TABLES TO hephaestus;