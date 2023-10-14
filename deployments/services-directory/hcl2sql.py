from typing import Tuple

import yaml

from entities import Service, _service

GRANT_PERMISSION_STRING = """
GRANT USAGE ON SCHEMA public TO {0};
GRANT SELECT, INSERT, UPDATE{1}ON ALL TABLES IN SCHEMA public TO {0};
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO {0};
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO {0};
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT SELECT, INSERT, UPDATE{1}ON TABLES TO {0};
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO {0};
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO {0};
"""
GRANT_PERMISSION_KAFKA = """
GRANT USAGE ON SCHEMA public TO kafka_connector;
GRANT SELECT, INSERT, UPDATE{0} ON ALL TABLES IN SCHEMA public TO kafka_connector;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE{0} ON TABLES TO kafka_connector;
"""


class SQL:
    def __init__(self) -> None:
        self._databases: set[str] = set()
        self._users: set[str] = set()
        self._bypassrls_users: set[str] = set()
        self._grant_all_privileges: set[Tuple[str, str]] = set()
        self._grant_create_on_databases: set[Tuple[str, str]] = set()
        self._database_grants: set[str] = set()
        # Some special cases need to be added manually
        self._bypassrls_users.add("hasura")
        self._users.add("hasura")

    def set_value(self, service: Service):
        self._add_createdatabases(service)
        if not service.disable_iam:
            self._add_createusers(service)
            self._add_grants(service)
        self._add_database_grants(service)

    def add_grant_all_privileges(self, dbname: str, dbuser: str) -> None:
        self._grant_all_privileges.add((dbname, dbuser))

    def add_grant_create_on_database(self, dbname: str, dbuser: str) -> None:
        self._grant_create_on_databases.add((dbname, dbuser))

    def _add_createdatabases(self, service: Service) -> None:
        if service.postgresql.createdb:
            self._databases.add(service.name)

    def _add_createusers(self, service: Service) -> None:
        self._users.add(service.name)
        if service.hasura.v2_enabled:
            self._users.add(service.hasura_name())

    def _add_grants(self, service: Service) -> None:
        # Allow hasura v1 to create metadata schema
        if service.hasura.enabled:
            self._grant_all_privileges.add((service.name, "hasura"))

        # Allow hasura v2 to create metadata schema
        if service.hasura.v2_enabled:
            self._grant_create_on_databases.add((
                service.hasura_metadata_database(),
                service.hasura_name())
            )

        # Add user bypassrls
        if service.postgresql.bypassrls:
            self._bypassrls_users.add(service.name)

    def _add_database_grants(self, service: Service) -> None:
        if len(service.postgresql.grants) or service.hasura.enabled or service.kafka.enabled:
            def _generate_hasura_grants() -> str:
                if service.hasura.enabled:
                    return GRANT_PERMISSION_STRING.format("hasura", " ")
                return ""

            def _generate_kafka_grants() -> str:
                if service.kafka.enabled:
                    return GRANT_PERMISSION_KAFKA.format(service.kafka.delete_string)
                return ""

            def _generate_common_grants() -> str:
                return "\n".join(
                    GRANT_PERMISSION_STRING.format(
                        grant.dbname, grant.delete_string)
                    for grant in service.postgresql.grants
                )
            self._database_grants.add(f"\\connect {service.name};" + "\n".join([
                _generate_common_grants(),
                _generate_hasura_grants(),
                _generate_kafka_grants(),
            ]))

    def generate(self, outpath: str) -> None:
        def _generate_createdatabase() -> str:
            sorted_databases = sorted(self._databases)
            return '\n'.join(
                f'CREATE DATABASE {dbname};' for dbname in sorted_databases
            )

        def _generate_create_hasura_v2_database() -> str:
            sorted_databases = sorted(self._grant_create_on_databases)
            return '\n'.join(
                f'CREATE DATABASE {dbname};' for (dbname, _) in sorted_databases
            )

        def _generate_createusers() -> str:
            sorted_users = sorted(self._users)
            return '\n'.join(
                f'CREATE USER \"{dbuser}\" WITH PASSWORD \'example\';'
                for dbuser in sorted_users
            )

        def _generate_bypassrls() -> str:
            sorted_bypassrls_users = sorted(self._bypassrls_users)
            return '\n'.join(
                f'ALTER ROLE \"{dbuser}\" BYPASSRLS;'
                for dbuser in sorted_bypassrls_users
            )

        def _generate_grant_all_privileges() -> str:
            sorted_list = sorted(self._grant_all_privileges)
            return '\n'.join(
                f'GRANT ALL PRIVILEGES ON DATABASE \"{dbname}\" TO {dbuser};'
                for (dbname, dbuser) in sorted_list
            )

        def _generate_grant_create_on_database() -> str:
            sorted_list = sorted(self._grant_create_on_databases)
            return '\n'.join(
                f'GRANT CREATE ON DATABASE \"{dbname}\" TO \"{dbuser}\";'
                for (dbname, dbuser) in sorted_list
            )

        def _generate_grant_for_hasura_metadata() -> str:
            sorted_list = sorted(self._grant_create_on_databases)
            return '\n'.join(
                f'\\connect {dbname};' +
                GRANT_PERMISSION_STRING.format(dbuser, " ")
                for (dbname, dbuser) in sorted_list
            )

        def _generate_database_grants() -> str:
            sorted_grants = sorted(self._database_grants)
            return '\n'.join(sorted_grants)
        out = '\n\n'.join([
            _generate_createdatabase(),
            _generate_create_hasura_v2_database(),
            _generate_createusers(),
            _generate_bypassrls(),
            _generate_grant_all_privileges(),
            _generate_grant_create_on_database(),
            _generate_database_grants(),
            _generate_grant_for_hasura_metadata(),
        ])
        with open(outpath, 'w') as f:
            f.write(out)


def main() -> None:
    defs_path: str = "./deployments/decl/stag-defs.yaml"
    sql_path: str = "./deployments/helm/emulators/infras/migrations/init.sql"

    with open(defs_path, 'r') as f:
        service_defs: list[_service] = yaml.safe_load(f)
    service_defs.sort(key=lambda x: x['name'])

    sql = SQL()
    for s in service_defs:
        sql.set_value(Service(s))
    _ = sql.generate(sql_path)


if __name__ == '__main__':
    main()
