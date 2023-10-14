INSERT INTO day_type (day_type_id, display_name, resource_path, is_archived, created_at, updated_at, deleted_at)
SELECT date_type_id, display_name, resource_path,is_archived, created_at, updated_at, deleted_at FROM date_type;

INSERT INTO day_info (date,location_id, day_type_id,opening_time,status,resource_path,time_zone, created_at, updated_at, deleted_at)
SELECT date, location_id, date_type_id,opening_time ,status::text::day_info_status ,resource_path,time_zone,created_at, updated_at, deleted_at FROM date_info;
