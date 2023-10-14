set -euo pipefail

DB_NAME="mastermgmt"

ORG_ID=$1
ORG_NAME=$2
TENANT_ID=$3
LOGO_URL=$4
DOMAIN_NAME=$5

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
--- Organization ---
INSERT INTO organizations (organization_id, tenant_id,name,resource_path, domain_name, logo_url, country,  created_at, updated_at, deleted_at)
VALUES('${ORG_ID}','${TENANT_ID}', '${ORG_NAME}', '${ORG_ID}', '${DOMAIN_NAME}', '${LOGO_URL}', 'COUNTRY_JP', now(),  now(), null) ON CONFLICT DO NOTHING;
EOF
