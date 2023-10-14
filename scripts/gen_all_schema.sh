#!/bin/sh

DB_POD=$(kubectl get pods -l app.kubernetes.io/name=postgres-infras -n emulator -o=name)

CURDIR=$(pwd)
DB_USER="postgres"
MODULES="eureka"

for module in $MODULES; do
    kubectl exec -it -n emulator "$DB_POD" -- /bin/bash -c "pg_dump -d $module -U $DB_USER --schema=public --schema-only > $module.sql"
    kubectl exec -n emulator "$DB_POD" -- tar cf - "$module".sql | tar xf - -C "$CURDIR"/internal/"$module"/
done
