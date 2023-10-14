ALTER TABLE tagged_user ADD CONSTRAINT tagged_user__parent_id__fk FOREIGN KEY(user_id) REFERENCES parents(parent_id);
