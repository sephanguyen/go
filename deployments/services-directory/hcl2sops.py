import yaml

from entities import Service

KMS_DB_MIGRATION = """  # TODO: Remove when finish create SA to migrate database instead postgres's password
  - path_regex: secrets\\/manabie\\/prod\\/([^\\/]+)_migrate
    gcp_kms: projects/student-coach-e1e95/locations/asia-southeast1/keyRings/manabie/cryptoKeys/prod-manabie

  - path_regex: secrets\\/jprep\\/prod\\/([^\\/]+)_migrate
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jprep/cryptoKeys/prod-jprep

  - path_regex: secrets\\/synersia\\/prod\\/([^\\/]+)_migrate
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-synersia

  - path_regex: secrets\\/renseikai\\/prod\\/([^\\/]+)_migrate
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-renseikai

  - path_regex: secrets\\/ga\\/prod\\/([^\\/]+)_migrate
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-ga

  - path_regex: secrets\\/kec\\/prod\\/([^\\/]+)_migrate
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-kec

  - path_regex: secrets\\/aic\\/prod\\/([^\\/]+)_migrate
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-aic

  - path_regex: secrets\\/nsg\\/prod\\/([^\\/]+)_migrate
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-nsg

  - path_regex: secrets\\/tokyo\\/prod\\/([^\\/]+)_migrate
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/prod-tokyo/cryptoKeys/prod-tokyo\n\n"""

KMS_DEFAULT = """  # Rule to match appsmith secrets.
  - path_regex: appsmith\\/([^\\/]+)\\/secrets\\/([^\\/]+)\\/local
    gcp_kms: projects/dev-manabie-online/locations/global/keyRings/deployments/cryptoKeys/github-actions

  - path_regex: appsmith\\/appsmith\\/secrets\\/([^\\/]+)\\/stag
    gcp_kms: projects/staging-manabie-online/locations/global/keyRings/backend-services/cryptoKeys/stag-appsmith

  - path_regex: appsmith\\/mongodb\\/secrets\\/([^\\/]+)\\/stag
    gcp_kms: projects/staging-manabie-online/locations/global/keyRings/backend-services/cryptoKeys/stag-mongodb

  - path_regex: appsmith\\/appsmith\\/secrets\\/([^\\/]+)\\/uat
    gcp_kms: projects/staging-manabie-online/locations/global/keyRings/backend-services/cryptoKeys/uat-appsmith

  - path_regex: appsmith\\/mongodb\\/secrets\\/([^\\/]+)\\/uat
    gcp_kms: projects/staging-manabie-online/locations/global/keyRings/backend-services/cryptoKeys/uat-mongodb

  - path_regex: appsmith\\/appsmith\\/secrets\\/([^\\/]+)\\/prod
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/backend-services/cryptoKeys/prod-appsmith

  - path_regex: appsmith\\/mongodb\\/secrets\\/([^\\/]+)\\/prod
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/backend-services/cryptoKeys/prod-mongodb

  # The following rules match various environments.
  # Note that only platform members can access and manage these UAT/production keys.
  - path_regex: secrets\\/e2e\\/local\\/
    gcp_kms: projects/dev-manabie-online/locations/global/keyRings/deployments/cryptoKeys/github-actions

  - path_regex: secrets\\/manabie\\/local\\/
    gcp_kms: projects/dev-manabie-online/locations/global/keyRings/deployments/cryptoKeys/github-actions

  - path_regex: secrets\\/manabie\\/stag\\/
    gcp_kms: projects/staging-manabie-online/locations/global/keyRings/deployments/cryptoKeys/github-actions

  - path_regex: secrets\\/manabie\\/uat\\/
    gcp_kms: projects/staging-manabie-online/locations/global/keyRings/deployments/cryptoKeys/uat-manabie

  - path_regex: secrets\\/jprep\\/stag\\/
    gcp_kms: projects/staging-manabie-online/locations/global/keyRings/deployments/cryptoKeys/stag-jprep

  - path_regex: secrets\\/jprep\\/uat\\/
    gcp_kms: projects/staging-manabie-online/locations/global/keyRings/deployments/cryptoKeys/uat-jprep

  - path_regex: secrets\\/jprep\\/prod\\/
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jprep/cryptoKeys/prod-jprep

  - path_regex: secrets\\/synersia\\/prod
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-synersia

  - path_regex: secrets\\/renseikai\\/prod
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-renseikai

  - path_regex: secrets\\/ga\\/prod
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-ga

  - path_regex: secrets\\/aic\\/prod
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-aic

  - path_regex: secrets\\/tokyo\\/prod
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/prod-tokyo/cryptoKeys/prod-tokyo

  # This file contains the raw passwords for unleash. It is not used in any deployment anywhere.
  # Admins can decrypt this file to get the passwords to login to unleash.
  - path_regex: deployments\\/helm\\/platforms\\/unleash\\/secrets\\/unleash_raw_passwords.secrets.encrypted.yaml
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/backend-services/cryptoKeys/prod-unleash
  - path_regex: deployments\\/helm\\/platforms\\/unleash\\/secrets\\/unleash_admin_tokens.secrets.encrypted.yaml
    gcp_kms: projects/student-coach-e1e95/locations/asia-southeast1/keyRings/manabie/cryptoKeys/prod-manabie

  # production data warehouse kafka
  - path_regex: helm\\/data-warehouse\\/kafka-connect\\/secrets\\/tokyo\\/(prod|dorp)
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/backend-services/cryptoKeys/prod-dwh-kafka-connect

  - path_regex: helm\\/data-warehouse\\/cp-schema-registry\\/secrets\\/tokyo\\/(prod|dorp)
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/backend-services/cryptoKeys/prod-dwh-cp-schema-registry

  - path_regex: helm\\/data-warehouse\\/kafka\\/secrets\\/tokyo\\/(prod|dorp)
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/backend-services/cryptoKeys/prod-dwh-kafka

  - path_regex: helm\\/data-warehouse\\/ksql-server\\/secrets\\/tokyo\\/(prod|dorp)
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/backend-services/cryptoKeys/prod-dwh-cp-ksql-server

  # Camel
  - path_regex: camel\\/integrations\\/quarkus\\/secrets\\/([^\\/]+)\\/local
    gcp_kms: projects/dev-manabie-online/locations/global/keyRings/deployments/cryptoKeys/github-actions
  - path_regex: camel\\/integrations\\/quarkus\\/secrets\\/([^\\/]+)\\/stag
    gcp_kms: projects/staging-manabie-online/locations/global/keyRings/deployments/cryptoKeys/github-actions

  # This fixes `no matching creation rule found` when decrypting
  - path_regex: ""
    gcp_kms: ""
"""


class Sops:
    def __init__(self, **kwargs) -> None:
        self._service_names: set[str] = set()

        self._fp: str = kwargs.get('fp', '.sops.yaml')
        with open(self._fp, 'w') as f:
            f.write('# File is generated by `python deployments/services-directory/hcl2sops.py` command. DO NOT EDIT.\n')
            f.write('creation_rules:\n')

        self._services: dict[str, set[str]] = {}
        for e in ['prod', 'uat', 'stag']:
            fp: str = kwargs.get(f'{e}_def_filepath',
                                 f'deployments/decl/{e}-defs.yaml')
            with open(fp, 'r') as f:
                service_list: list[Service] = [
                    Service(s) for s in yaml.safe_load(f)
                ]
                self._services[e] = set(
                    s.name for s in service_list if not s.disable_iam)

    def _config(self, environment) -> tuple[str, str]:
        match environment:
            case 'stag':
                project_id: str = 'staging-manabie-online'
                location: str = 'global'
            case 'uat':
                project_id: str = 'staging-manabie-online'
                location: str = 'global'
            case 'prod':
                project_id: str = 'student-coach-e1e95'
                location: str = 'asia-northeast1'
            case _:
                raise Exception(f'invalid environment {environment}\n')
        return project_id, location

    def _add_dorp_env_for_prod(self, environment:str, service:str) -> str:
      if environment == "prod" and service in ['hephaestus','data-warehouse', 'import-map-deployer'] :
        return f'({environment}|dorp)'
      return environment

    def generate(self) -> None:
        def _rule(environment: str, service: str) -> str:
            project_id, location = self._config(environment)
            return f"""  - path_regex: {service}\\/secrets\\/([^\\/]+)\\/{self._add_dorp_env_for_prod(environment,service)}\\/
    gcp_kms: projects/{project_id}/locations/{location}/keyRings/backend-services/cryptoKeys/{environment}-{service}\n\n"""

        envs: list[str] = ['stag', 'uat', 'prod']
        with open(self._fp, 'a') as f:
            f.write(KMS_DB_MIGRATION)
            for e in envs:
                f.write(
                    f"""  # Rules to match {e} secrets for backend and platform services.\n""")
                for s in sorted(self._services[e]):
                    f.write(_rule(e, s))
            f.write(KMS_DEFAULT)

    def generate_for_nats(self) -> None:
        def _rule(environment: str, service: str) -> str:
            project_id, location = self._config(environment)
            return f"""  - path_regex: helm\\/platforms\\/nats-jetstream\\/secrets\\/([^\\/]+)\\/{self._add_dorp_env_for_prod(environment,service)}\\/{service}
    gcp_kms: projects/{project_id}/locations/{location}/keyRings/backend-services/cryptoKeys/{environment}-{service}\n\n"""

        envs: list[str] = ['prod', 'uat', 'stag']
        with open(self._fp, 'a') as f:
            for e in envs:
                f.write(f'  # Rules for NATS Jetstream in {e}\n')
                for s in sorted(self._services[e]):
                    f.write(_rule(e, s))

            f.write("""  # Match other secrets in NATS
  - path_regex: helm\\/platforms\\/nats-jetstream\\/secrets\\/([^\\/]+)\\/prod\\/
    gcp_kms: projects/student-coach-e1e95/locations/asia-northeast1/keyRings/backend-services/cryptoKeys/prod-nats-jetstream
  - path_regex: helm\\/platforms\\/nats-jetstream\\/secrets\\/([^\\/]+)\\/uat\\/
    gcp_kms: projects/staging-manabie-online/locations/global/keyRings/backend-services/cryptoKeys/uat-nats-jetstream
  - path_regex: helm\\/platforms\\/nats-jetstream\\/secrets\\/([^\\/]+)\\/stag\\/
    gcp_kms: projects/staging-manabie-online/locations/global/keyRings/backend-services/cryptoKeys/stag-nats-jetstream
  - path_regex: helm\\/platforms\\/nats-jetstream\\/secrets\\/([^\\/]+)\\/local\\/
    gcp_kms: projects/dev-manabie-online/locations/global/keyRings/deployments/cryptoKeys/github-actions\n\n""")


def main() -> None:
    s = Sops()
    s.generate_for_nats()
    s.generate()


if __name__ == '__main__':
    main()
