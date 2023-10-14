INSERT INTO public.configuration_key(
    config_key,
    value_type, 
    default_value, 
    configuration_type, 
    created_at, 
    updated_at
)
VALUES(
    'arch.master_management.import_rules', 
    'string', 
    '{"location":[{"name":"partner_internal_id","rules":"required|max:20"},{"name":"name","rules":"required|max:20"},{"name":"location_type","rules":"required|max:20"},{"name":"partner_internal_parent_id","rules":"max:20"}],"location_type":[{"name":"name","rules":"required|max:20"},{"name":"display_name","rules":"required"},{"name":"level","rules":"max:20"}],"grade":[{"name":"grade_id","rules":""},{"name":"grade_partner_id","rules":"required|max:20"},{"name":"name","rules":"required|max:30"},{"name":"sequence","rules":"required"},{"name":"remarks","rules":""}],"course":[{"name":"course_id","rules":""},{"name":"course_name","rules":"required|max:20"},{"name":"course_type_id","rules":""},{"name":"remarks","rules":""}]}', 
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
);
