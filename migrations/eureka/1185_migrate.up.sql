create or replace function get_student_topic_progress()
    returns table
            (
                student_id        text,
                study_plan_id     text,
                chapter_id        text,
                topic_id          text,
                completed_sp_item smallint,
                total_sp_item     smallint,
                average_score     smallint
            )
    language sql
    stable
as
$$
select student_id,
       study_plan_id,
       chapter_id,
       topic_id,
       (select count(*) from get_student_completion_learning_material() gsclm
       			join learning_material lm using (learning_material_id)
       			where gsclm.student_id = lalm.student_id
       			and gsclm.study_plan_id = lalm.study_plan_id
       			and lm.topic_id = lalm.topic_id)::smallint   as completed_sp_item,
       count(*)::smallint  as total_sp_item,
       (avg(gs.graded_points * 1.0 / gs.total_points) * 100)::smallint as average_score
from list_available_learning_material() lalm
         left join max_graded_score() gs
                   using (student_id, study_plan_id, learning_material_id)
group by student_id, study_plan_id, chapter_id, topic_id
$$;

create or replace function get_student_chapter_progress()
    returns table
            (
                student_id    text,
                study_plan_id text,
                chapter_id    text,
                average_score smallint
            )
    language sql
    stable
as
$$
select student_id,
       study_plan_id,
       chapter_id,
       avg(average_score)::smallint
from get_student_topic_progress()
group by student_id, study_plan_id, chapter_id
$$;