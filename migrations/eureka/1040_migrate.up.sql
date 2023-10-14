-- migrate content_structure from content_structure_flatten
-- sample content_structure 'book::book-id-1topic::topic-id-1chapter::chapter-id-1course::course-id-1lo::01FGNQRHWX2WYSKF41SC8A8WNE'
-- update content_structure assignment_id when content_structure_flatten has pattern '%assignment::%'
update study_plan_items
set content_structure = content_structure || concat('{"assignment_id": "', split_part(content_structure_flatten, '::', 6), '"}')::jsonb
where content_structure_flatten like '%assignment::%';

-- update content_structure lo_id when content_structure_flatten has pattern '%lo::%'
update study_plan_items
set content_structure = content_structure || concat('{"lo_id": "', split_part(content_structure_flatten, '::', 6), '"}')::jsonb
where content_structure_flatten like '%lo::%';

-- trigger to update content_structure (lo_id, assignment_id) from content_structure_flatten
DROP TRIGGER IF EXISTS update_content_structure ON public.study_plan_items;

CREATE OR REPLACE FUNCTION update_content_structure_fnc()
  RETURNS trigger AS
$$
begin
    if new.content_structure_flatten is not null then 
        -- add assignment_id into content_structure when content_structure_flatten contains 'assignment::'
        if new.content_structure_flatten like '%assignment::%' then
            new.content_structure = new.content_structure || concat('{"assignment_id": "', split_part(new.content_structure_flatten, '::', 6), '"}')::jsonb;
        -- add lo_id into content_structure when content_structure_flatten contains 'lo::'
        elsif new.content_structure_flatten like '%lo::%' then 
            new.content_structure = new.content_structure || concat('{"lo_id": "', split_part(new.content_structure_flatten, '::', 6), '"}')::jsonb;
        end if;
    end if;
RETURN NEW;

END;

$$

LANGUAGE 'plpgsql';

CREATE TRIGGER update_content_structure
BEFORE INSERT OR UPDATE
ON study_plan_items
for each row 
execute procedure update_content_structure_fnc();

-- migrate missing status records
UPDATE study_plan_items
SET status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE'::TEXT
WHERE status IS NULL;
