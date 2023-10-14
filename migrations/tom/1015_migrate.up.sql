ALTER TABLE ONLY public.conversation_students
    ADD CONSTRAINT student_id_conversation_type_un UNIQUE (student_id, conversation_type);