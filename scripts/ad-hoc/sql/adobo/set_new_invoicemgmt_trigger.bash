#!/bin/bash

# This script truncates data in some invoicemgmt-related tables.

set -euo pipefail

DB_NAME="invoicemgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

CREATE OR REPLACE FUNCTION fill_seq_invoice() RETURNS TRIGGER 
AS \$\$
DECLARE
	resourcePath text;
BEGIN
	resourcePath := current_setting('permission.resource_path', 't');
    -- set an advisory_lock with resource_path as key
    PERFORM pg_advisory_lock(hashtext(resourcePath));
    -- set next value within the resource_path
    SELECT coalesce(max(invoice_sequence_number), 0) + 1 INTO NEW.invoice_sequence_number from public.invoice where resource_path = resourcePath;
    RETURN NEW;
END \$\$ LANGUAGE plpgsql;

EOF
