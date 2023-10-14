This module is used to grant privileges for Cloud SQL instance roles, like
read_only_role and read_write_role, to a list of databases.

For each database:
  - read_only_role:
    - Grant SELECT on table.
    - Grant SELECT on sequence.

  - read_write_role:
    - Grant SELECT, INSERT, UPDATE on table.
    - Grant SELECT, USAGE, UPDATE on sequence.
