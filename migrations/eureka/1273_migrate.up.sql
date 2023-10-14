CREATE OR REPLACE FUNCTION public.find_question_by_lo_id(id VARCHAR) RETURNS SETOF public.quizzes
LANGUAGE sql STABLE
AS $$
WITH question_ids AS (
    SELECT qei.id, ROW_NUMBER() OVER (ORDER BY qei.path) as display_order from (
        SELECT qh.js ->> 'id' as id,
        ARRAY[idx] as path 
        FROM quiz_sets qs,
        unnest(qs.question_hierarchy) with ordinality as qh(js, idx)
        where qs.lo_id  = id
        and qs.deleted_at is null
        UNION 
        SELECT qhci.cids as id,
        ARRAY[qh.idx, qhci.idx] as path
        FROM quiz_sets qs,
        unnest(qs.question_hierarchy) with ordinality as qh(js, idx),
        jsonb_array_elements_text(qh.js -> 'children_ids') with ordinality as qhci(cids, idx)
        where qs.lo_id  = id
        and qs.deleted_at is null
        and jsonb_typeof(qh.js -> 'children_ids') = 'array'
    ) qei
)
SELECT 
    quiz_id,
    country,
    school_id,
    external_id,
    kind,
    question,
    explanation,
    options,
    tagged_los,
    difficulty_level,
    created_by,
    approved_by,
    status,
    coalesce(q.updated_at, qr.updated_at) as updated_at ,
    coalesce(q.created_at, qr.created_at) as created_at,
    coalesce(q.deleted_at, qr.deleted_at) as deleted_at,
    lo_ids,
    coalesce(q.resource_path, qr.resource_path) as resource_path,
    point,
    coalesce(q.question_group_id, qr.question_group_id) as question_group_id,
    question_tag_ids,
    label_type
FROM question_ids qi
LEFT JOIN quizzes q on qi.id = q.external_id
LEFT JOIN question_group qr on qi.id = qr.question_group_id
order by qi.display_order ASC
$$;
