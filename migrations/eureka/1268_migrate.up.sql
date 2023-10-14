create or replace function get_exam_lo_returned_scores() returns table (
        student_id text,
        study_plan_id text,
        learning_material_id text,
        submission_id text,
        graded_point smallint,
        total_point smallint,
        status text,
        result text,
        created_at timestamptz
    ) language sql stable as $$
select els.student_id,
       els.study_plan_id,
       els.learning_material_id,
       els.submission_id,
       --   when teacher manual grading, we should use with score from teacher
        (
            case el.review_option
                when 'EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE' then
                        case
                            when (select coalesce(
                                    (   select isp.end_date
                                        from individual_study_plan isp
                                        where isp.deleted_at is null
                                            and isp.student_id = els.student_id
                                            and isp.learning_material_id = els.learning_material_id
                                            and isp.study_plan_id = els.study_plan_id
                                    ),
                                    (   select msp.end_date
                                        from master_study_plan msp
                                        where msp.deleted_at is null
                                            and msp.study_plan_id = els.study_plan_id
                                            and msp.learning_material_id = els.learning_material_id
                                    )
                                ) < now())
                            then (els.status = 'SUBMISSION_STATUS_RETURNED')::BOOLEAN::INT * sum(coalesce(elss.point, elsa.point))
                        end
                else (els.status = 'SUBMISSION_STATUS_RETURNED')::BOOLEAN::INT * sum(coalesce(elss.point, elsa.point))
            end
        )::smallint as graded_point,
       els.total_point::smallint as total_point,
       els.status,
       els.result,
       els.created_at
from exam_lo_submission els
    join exam_lo el using (learning_material_id)
    join exam_lo_submission_answer elsa using (submission_id)
    left join exam_lo_submission_score elss using (submission_id, quiz_id)
group by  els.submission_id, el.review_option $$;
