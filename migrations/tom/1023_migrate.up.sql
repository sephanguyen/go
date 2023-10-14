CREATE OR REPLACE FUNCTION insert_message_in_conversation_fn()
	RETURNS TRIGGER
	LANGUAGE PLPGSQL
	AS
	$$
	BEGIN
		UPDATE conversations c
		SET last_message_id = NEW.message_id
		WHERE 
			(NEW.type != 'MESSAGE_TYPE_SYSTEM' OR NEW.message = 'CODES_MESSAGE_TYPE_CREATED_CONVERSATION') 
			AND c.conversation_id = NEW.conversation_id
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