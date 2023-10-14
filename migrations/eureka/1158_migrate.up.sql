create or replace function lo_raw_answer()
    returns table
            (
                student_id           text,
                study_plan_id        text,
                learning_material_id text,
                submission_id        text,
                submission_history   jsonb,
                quiz_external_ids    text[]
            )
    LANGUAGE sql
    STABLE
AS
$$
select student_id,
       study_plan_id,
       learning_material_id,
       shuffled_quiz_set_id as submission_id,
       submission_history,
       quiz_external_ids
from shuffled_quiz_sets
join learning_objective using (learning_material_id)
$$;


create or replace function lo_answer()
    returns table
            (
                student_id           text,
                study_plan_id        text,
                learning_material_id text,
                submission_id        text,
                external_quiz_id     text,
                is_accepted          bool,
                point                integer
            )
    LANGUAGE sql
    STABLE
AS
$$
select student_id,
       study_plan_id,
       learning_material_id,
       submission_id,
       external_quiz_id,
       bool_or(is_accepted) as is_accepted,
       q.point
from (select student_id,
             study_plan_id,
             learning_material_id,
             submission_id,
             (jsonb_array_elements(submission_history) ->> 'quiz_id')::text     as external_quiz_id,
             (jsonb_array_elements(submission_history) ->> 'is_accepted')::bool as is_accepted
      from lo_raw_answer() sqs
      union
      select student_id,
             study_plan_id,
             learning_material_id,
             submission_id,
             unnest(quiz_external_ids) as external_quiz_id,
--          not in submission, so answer default is wrong
             false                     as is_accepted
      from lo_raw_answer() sqs) raw_answer
         join quizzes q
              on raw_answer.external_quiz_id = q.external_id
group by student_id,
         study_plan_id,
         learning_material_id,
         submission_id,
         external_quiz_id,
         q.point
$$;

create or replace function lo_graded_score()
    returns table
            (
                student_id           text,
                study_plan_id        text,
                learning_material_id text,
                submission_id        text,
                graded_points        smallint,
                total_points         smallint,
                status               text
            )
    LANGUAGE sql
    STABLE
AS
$$
select student_id,
       study_plan_id,
       learning_material_id,
       submission_id,
       sum(is_accepted::int * point)::smallint as graded_score,
       sum(point)::smallint                    as total_scores,
       'S'
from lo_answer()
group by student_id,
         study_plan_id,
         learning_material_id,
         submission_id
$$;