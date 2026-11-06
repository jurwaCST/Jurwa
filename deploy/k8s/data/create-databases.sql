-- Runs once, when the PostgreSQL volume is empty. One database per service
-- (docs/design/database-schema.md section 2). Each service creates its own tables
-- at start-up (Migrate), from docs/design/schema/<service>.sql.
CREATE DATABASE report_history;
CREATE DATABASE location_taxonomy;
CREATE DATABASE routing_rules;
CREATE DATABASE department_directory;
CREATE DATABASE category_taxonomy;
CREATE DATABASE status_lifecycle;
CREATE DATABASE sla_policy;
CREATE DATABASE reassignment;
CREATE DATABASE identity;
CREATE DATABASE authorization_db;
CREATE DATABASE notification;
CREATE DATABASE feedback;
CREATE DATABASE analytics;
