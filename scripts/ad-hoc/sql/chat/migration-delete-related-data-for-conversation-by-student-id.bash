#!/bin/bash

set -euo pipefail

DB_NAME="tom"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --echo-all <<EOF
DO \$\$
DECLARE
affected_rows int;
BEGIN
    CREATE TEMP TABLE will_delete_conversation_ids AS (
        SELECT conversation_id 
        FROM conversation_members cm 
        JOIN users u ON cm.user_id = u.user_id
        WHERE u.user_id = ANY (
            SELECT user_id
            FROM users 
            WHERE email > '1004@with-us.co.jp' 
                AND email < '2001@with-us.co.jp' 
                AND email LIKE '%@with-us.co.jp'
                AND "name" LIKE '生徒 テスト%'
                AND resource_path = '-2147483629'
        )
    );
    
    UPDATE conversations SET last_message_id = NULL
    WHERE conversation_id = ANY (
        SELECT * FROM will_delete_conversation_ids
    );
    GET DIAGNOSTICS affected_rows = ROW_COUNT;
   	RAISE INFO '% rows of [conversations] table is/are updated last_message_id to NULL', affected_rows;
   	affected_rows := 0;
    DELETE FROM messages
    WHERE conversation_id = ANY (
        SELECT * FROM will_delete_conversation_ids
    );
    GET DIAGNOSTICS affected_rows = ROW_COUNT;
   	RAISE INFO '% rows of [messages] table is/are deleted', affected_rows;
   	affected_rows := 0;
    DELETE FROM conversation_students
    WHERE conversation_id = ANY (
        SELECT * FROM will_delete_conversation_ids
    );
    GET DIAGNOSTICS affected_rows = ROW_COUNT;
   	RAISE INFO '% rows of [conversation_students] table is/are deleted', affected_rows;
   	affected_rows := 0;
    DELETE FROM conversation_members 
    WHERE conversation_id = ANY (
        SELECT * FROM will_delete_conversation_ids
    );
    GET DIAGNOSTICS affected_rows = ROW_COUNT;
   	RAISE INFO '% rows of [conversation_members] table is/are deleted', affected_rows;
   	affected_rows := 0;
    DELETE FROM conversation_locations 
    WHERE conversation_id = ANY (
        SELECT * FROM will_delete_conversation_ids
    );
    GET DIAGNOSTICS affected_rows = ROW_COUNT;
   	RAISE INFO '% rows of [conversation_locations] table is/are deleted', affected_rows;
   	affected_rows := 0;
    DELETE FROM conversations 
    WHERE conversation_id = ANY (
        SELECT * FROM will_delete_conversation_ids
    );
    GET DIAGNOSTICS affected_rows = ROW_COUNT;
   	RAISE INFO '% rows of [conversations] table is/are deleted', affected_rows;
    COMMIT;
	
    DROP TABLE IF EXISTS will_delete_conversation_ids;
END \$\$;
EOF
