ALTER TABLE IF EXISTS conversations
    ADD COLUMN IF NOT EXISTS last_message_id TEXT
        REFERENCES messages(message_id) ON DELETE SET NULL;

-- Fill in data to new column last_message_id
WITH matcher_table AS (
	SELECT m.conversation_id, m.message_id
	FROM messages m
	JOIN (
		SELECT conversation_id, max(created_at) AS last_created_at
		FROM messages
		GROUP BY conversation_id
	) tmp
	ON m.conversation_id = tmp.conversation_id
		AND m.created_at = tmp.last_created_at
)
UPDATE conversations AS c
    SET last_message_id = (
        SELECT mt.message_id
        FROM matcher_table mt
        WHERE mt.conversation_id = c.conversation_id
		LIMIT 1
    );

-- Create function and trigger to automatically set new last_message_id for each conversation
CREATE OR REPLACE FUNCTION insert_message_in_conversation_fn()
	RETURNS TRIGGER
	LANGUAGE PLPGSQL
	AS
	$$
	BEGIN
		UPDATE conversations c
		SET last_message_id = NEW.message_id
		WHERE c.conversation_id = NEW.conversation_id
			AND (
				c.last_message_id IS NULL
				OR
				NEW.created_at > (
					SELECT created_at
					FROM messages
					WHERE message_id = c.last_message_id
				)
			);
		RETURN NEW;
	END
	$$
    VOLATILE
    STRICT;

CREATE TRIGGER insert_message_in_conversation
    AFTER INSERT
	ON messages
    FOR EACH ROW
    EXECUTE FUNCTION insert_message_in_conversation_fn();

-- Create a function and trigger to re-update last_message_id if the lastest message is deleted
CREATE OR REPLACE FUNCTION delete_message_in_conversation_fn()
	RETURNS TRIGGER
	LANGUAGE PLPGSQL
	AS
	$$
    DECLARE
		convo_last_message_id text := (SELECT last_message_id FROM conversations WHERE conversation_id = OLD.conversation_id);
	BEGIN
		-- Check if the deleted message is the last_message_id before update
		IF (convo_last_message_id IS NULL OR convo_last_message_id = OLD.message_id) 
		THEN
			UPDATE conversations
			SET last_message_id = (
				SELECT message_id
				FROM messages
				WHERE conversation_id = OLD.conversation_id
					AND message_id <> OLD.message_id
					AND deleted_at IS NULL
					AND deleted_by IS NULL
				ORDER BY created_at DESC
				LIMIT 1
			)
			WHERE conversation_id = OLD.conversation_id;
		END IF;
		RETURN NEW;
	END
	$$
	VOLATILE
	STRICT;

CREATE TRIGGER soft_delete_message_in_conversation
	AFTER UPDATE OF deleted_by, deleted_at
	ON messages
	FOR EACH ROW
	EXECUTE FUNCTION delete_message_in_conversation_fn();

CREATE TRIGGER delete_message_in_conversation
	AFTER DELETE
	ON messages
	FOR EACH ROW
	EXECUTE FUNCTION delete_message_in_conversation_fn();