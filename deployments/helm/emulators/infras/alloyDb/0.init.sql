CREATE DATABASE alloydb;
CREATE USER "kafka_connector" REPLICATION PASSWORD 'example';
\connect alloydb;

-- schema public
GRANT USAGE ON SCHEMA public TO kafka_connector;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public to kafka_connector;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO kafka_connector;

-- schema bob
CREATE SCHEMA bob;
GRANT CREATE, USAGE ON SCHEMA bob TO kafka_connector;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA bob to kafka_connector;
ALTER DEFAULT PRIVILEGES IN SCHEMA bob GRANT ALL PRIVILEGES ON TABLES TO kafka_connector;

-- schema timesheet
CREATE SCHEMA timesheet;
GRANT CREATE, USAGE ON SCHEMA timesheet TO kafka_connector;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA timesheet to kafka_connector;
ALTER DEFAULT PRIVILEGES IN SCHEMA timesheet GRANT ALL PRIVILEGES ON TABLES TO kafka_connector;

-- schema invoicemgmt
CREATE SCHEMA invoicemgmt;
GRANT USAGE ON SCHEMA invoicemgmt TO kafka_connector;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA invoicemgmt to kafka_connector;
ALTER DEFAULT PRIVILEGES IN SCHEMA invoicemgmt GRANT ALL PRIVILEGES ON TABLES TO kafka_connector;

-- schema fatima
CREATE SCHEMA fatima;
GRANT USAGE ON SCHEMA fatima TO kafka_connector;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA fatima to kafka_connector;
ALTER DEFAULT PRIVILEGES IN SCHEMA fatima GRANT ALL PRIVILEGES ON TABLES TO kafka_connector;

-- schema mastermgmt
CREATE SCHEMA mastermgmt;
GRANT USAGE ON SCHEMA mastermgmt TO kafka_connector;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA mastermgmt to kafka_connector;
ALTER DEFAULT PRIVILEGES IN SCHEMA mastermgmt GRANT ALL PRIVILEGES ON TABLES TO kafka_connector;

-- schema lessonmgmt
CREATE SCHEMA lessonmgmt;
GRANT USAGE ON SCHEMA lessonmgmt TO kafka_connector;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA lessonmgmt to kafka_connector;
ALTER DEFAULT PRIVILEGES IN SCHEMA lessonmgmt GRANT ALL PRIVILEGES ON TABLES TO kafka_connector;

-- schema calendar
CREATE SCHEMA calendar;
GRANT USAGE ON SCHEMA calendar TO kafka_connector;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA calendar to kafka_connector;
ALTER DEFAULT PRIVILEGES IN SCHEMA calendar GRANT ALL PRIVILEGES ON TABLES TO kafka_connector;
