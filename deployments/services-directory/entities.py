from typing import Any, Tuple, TypedDict
from typing_extensions import NotRequired


class _grant(TypedDict):
    dbname: str
    grant_delete: NotRequired[bool]


class _postgresql(TypedDict):
    createdb: NotRequired[bool]
    grants: NotRequired['list[_grant]']
    bypassrls: NotRequired[bool]


class _hasura(TypedDict):
    enabled: NotRequired[bool]
    v2_enabled: NotRequired[bool]


class _kafka(TypedDict):
    enabled: NotRequired[bool]
    grant_delete: NotRequired[bool]


class _service(TypedDict):
    name: str
    postgresql: NotRequired[_postgresql]
    hasura: NotRequired[_hasura]
    kafka: NotRequired[_kafka]
    disable_iam: NotRequired[bool]


class Grant:
    def __init__(self, v: _grant) -> None:
        self.dbname = v['dbname']
        self.delete_string = ", DELETE " if v.get(
            'grant_delete', False) else " "


class Postgresql:
    def __init__(self, v: _postgresql) -> None:
        self.createdb = v.get('createdb', False)
        self.grants = [Grant(g) for g in v.get('grants', [])]
        self.bypassrls = v.get('bypassrls', False)


class Hasura:
    def __init__(self, v: _hasura) -> None:
        self.enabled = v.get('enabled', False)
        self.v2_enabled = v.get('v2_enabled', False)


class Kafka:
    def __init__(self, v: _kafka) -> None:
        self.enabled = v.get('enabled', False)
        self.delete_string = ", DELETE " if v.get(
            'grant_delete', False) else " "


class Service:
    def __init__(self, s: _service) -> None:
        self.name = s['name']
        self.postgresql = Postgresql(s.get('postgresql', {}))
        self.hasura = Hasura(s.get('hasura', {}))
        self.kafka = Kafka(s.get('kafka', {}))
        self.disable_iam = s.get('disable_iam', False)

    # def update(self, sql: SQL):
    #     self._add_createdatabases(sql)
    #     if not self.disable_iam:
    #         self._add_createusers(sql)
    #         self._add_grants(sql)
    #     self._add_database_grants(sql)

    def hasura_name(self) -> str:
        return f'{self.name}_hasura'

    def hasura_metadata_database(self) -> str:
        return f'{self.name}_hasura_metadata'
