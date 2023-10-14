### Restart a service

```sh
# Restart deployment for a service (bob) in local, manabie
export ORG=manabie ENV=local ARGUMENT=restart/deployment/bob
./scripts/restart_service.bash
# Delete pod for a service (tom) in prod, manabie
export ORG=manabie ENV=prod ARGUMENT=delete/pod/tom-0
./scripts/restart_service.bash
# Delete job for a job (fatima-migrate-student-subscriptions) in uat, manabie
export ORG=manabie ENV=uat ARGUMENT=delete/job/fatima-migrate-student-subscriptions
./scripts/restart_service.bash

```