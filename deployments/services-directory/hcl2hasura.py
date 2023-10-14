from entities import Service, _service
import yaml


# TODO: accepts an ENV argument to decide which of stag/uat/prod definition to use.
def main() -> None:
    filepath: str = "./deployments/decl/stag-defs.yaml"

    with open(filepath, 'r') as f:
        data: list[_service] = yaml.safe_load(f)
        hasura_services: list[str] = []
        for svc_def in data:
            s: Service = Service(svc_def)
            if s.hasura.enabled:
                hasura_services.append(s.name)

    hasura_services.sort()
    print(" ".join(hasura_services))


if __name__ == '__main__':
    main()
