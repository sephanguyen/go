package repositories

import (
	"context"
	"fmt"

	entities "github.com/manabie-com/backend/internal/eureka/entities/learning_history_data_sync"
	dbeureka "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type LearningHistoryDataSyncRepo struct{}

func (r *LearningHistoryDataSyncRepo) RetrieveMappingCourseID(ctx context.Context, db database.QueryExecer) ([]*entities.MappingCourseID, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningHistoryDataSyncRepo.RetrieveMappingCourseID")
	defer span.End()

	query := `
	select manabie_course_id, withus_course_id, last_updated_date, last_updated_by, is_archived
	from withus_mapping_course_id;
	`

	resp := make([]*entities.MappingCourseID, 0)
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingCourseID.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		r := new(entities.MappingCourseID)
		err := rows.Scan(&r.ManabieCourseID, &r.WithusCourseID, &r.LastUpdatedDate, &r.LastUpdatedBy, &r.IsArchived)
		if err != nil {
			return nil, fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingCourseID.Scan: %w", err)
		}
		resp = append(resp, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingCourseID.Err: %w", err)
	}

	return resp, nil
}

func (r *LearningHistoryDataSyncRepo) RetrieveMappingExamLoID(ctx context.Context, db database.QueryExecer) ([]*entities.MappingExamLoID, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningHistoryDataSyncRepo.RetrieveMappingExamLoID")
	defer span.End()

	query := `
	select exam_lo_id, material_code, last_updated_date, last_updated_by, is_archived
	from withus_mapping_exam_lo_id;
	`

	resp := make([]*entities.MappingExamLoID, 0)
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingExamLoID.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		r := new(entities.MappingExamLoID)
		err := rows.Scan(&r.ExamLoID, &r.MaterialCode, &r.LastUpdatedDate, &r.LastUpdatedBy, &r.IsArchived)
		if err != nil {
			return nil, fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingExamLoID.Scan: %w", err)
		}
		resp = append(resp, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingExamLoID.Err: %w", err)
	}

	return resp, nil
}

func (r *LearningHistoryDataSyncRepo) RetrieveMappingQuestionTag(ctx context.Context, db database.QueryExecer) ([]*entities.MappingQuestionTag, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningHistoryDataSyncRepo.RetrieveMappingQuestionTag")
	defer span.End()

	query := `
	select manabie_tag_id, manabie_tag_name, withus_tag_name, last_updated_date, last_updated_by, is_archived
	from withus_mapping_question_tag
	order by last_updated_date;`

	resp := make([]*entities.MappingQuestionTag, 0)
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingQuestionTag.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		r := new(entities.MappingQuestionTag)
		err := rows.Scan(&r.ManabieTagID, &r.ManabieTagName, &r.WithusTagName, &r.LastUpdatedDate, &r.LastUpdatedBy, &r.IsArchived)
		if err != nil {
			return nil, fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingQuestionTag.Scan: %w", err)
		}
		resp = append(resp, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingQuestionTag.Err: %w", err)
	}

	return resp, nil
}

func (r *LearningHistoryDataSyncRepo) RetrieveFailedSyncEmailRecipient(ctx context.Context, db database.QueryExecer) ([]*entities.FailedSyncEmailRecipient, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningHistoryDataSyncRepo.RetrieveFailedSyncEmailRecipient")
	defer span.End()

	query := `
	select recipient_id, email_address, last_updated_date, last_updated_by, is_archived
	from public.withus_failed_sync_email_recipient;
	`

	resp := make([]*entities.FailedSyncEmailRecipient, 0)
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveFailedSyncEmailRecipient.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		r := new(entities.FailedSyncEmailRecipient)
		err := rows.Scan(&r.RecipientID, &r.EmailAddress, &r.LastUpdatedDate, &r.LastUpdatedBy, &r.IsArchived)
		if err != nil {
			return nil, fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveFailedSyncEmailRecipient.Scan: %w", err)
		}
		resp = append(resp, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveFailedSyncEmailRecipient.Err: %w", err)
	}

	return resp, nil
}

func (r *LearningHistoryDataSyncRepo) BulkUpsertMappingCourseID(ctx context.Context, db database.QueryExecer, items []*entities.MappingCourseID) error {
	ctx, span := interceptors.StartSpan(ctx, "LearningHistoryDataSyncRepo.BulkUpsertMappingCourseID")
	defer span.End()

	query := `
			INSERT INTO %s (%s) VALUES %s
			ON CONFLICT ON CONSTRAINT withus_mapping_course_id_pk
			DO UPDATE SET
				withus_course_id = EXCLUDED.withus_course_id,
				last_updated_date = EXCLUDED.last_updated_date,
				last_updated_by = EXCLUDED.last_updated_by,
				is_archived = EXCLUDED.is_archived
		`
	err := dbeureka.BulkUpsert(ctx, db, query, items)
	if err != nil {
		return fmt.Errorf("LearningHistoryDataSyncRepo.BulkUpsertMappingCourseID: %w", err)
	}

	return nil
}

func (r *LearningHistoryDataSyncRepo) BulkUpsertMappingExamLoID(ctx context.Context, db database.QueryExecer, items []*entities.MappingExamLoID) error {
	ctx, span := interceptors.StartSpan(ctx, "LearningHistoryDataSyncRepo.BulkUpsertMappingExamLoID")
	defer span.End()

	query := `
			INSERT INTO %s (%s) VALUES %s
			ON CONFLICT ON CONSTRAINT withus_mapping_exam_lo_id_pk
			DO UPDATE SET
				material_code = EXCLUDED.material_code,
				last_updated_date = EXCLUDED.last_updated_date,
				last_updated_by = EXCLUDED.last_updated_by,
				is_archived = EXCLUDED.is_archived
		`
	err := dbeureka.BulkUpsert(ctx, db, query, items)
	if err != nil {
		return fmt.Errorf("LearningHistoryDataSyncRepo.BulkUpsertMappingExamLoID: %w", err)
	}

	return nil
}

func (r *LearningHistoryDataSyncRepo) BulkUpsertMappingQuestionTag(ctx context.Context, db database.QueryExecer, items []*entities.MappingQuestionTag) error {
	ctx, span := interceptors.StartSpan(ctx, "LearningHistoryDataSyncRepo.BulkUpsertMappingQuestionTag")
	defer span.End()

	query := `
			INSERT INTO %s (%s) VALUES %s
			ON CONFLICT ON CONSTRAINT withus_mapping_question_tag_pk
			DO UPDATE SET
				withus_tag_name = EXCLUDED.withus_tag_name,
				last_updated_date = EXCLUDED.last_updated_date,
				last_updated_by = EXCLUDED.last_updated_by,
				is_archived = EXCLUDED.is_archived
		`
	err := dbeureka.BulkUpsert(ctx, db, query, items)
	if err != nil {
		return fmt.Errorf("LearningHistoryDataSyncRepo.BulkUpsertMappingQuestionTag: %w", err)
	}

	return nil
}

func (r *LearningHistoryDataSyncRepo) BulkUpsertFailedSyncEmailRecipient(ctx context.Context, db database.QueryExecer, items []*entities.FailedSyncEmailRecipient) error {
	ctx, span := interceptors.StartSpan(ctx, "LearningHistoryDataSyncRepo.BulkUpsertFailedSyncEmailRecipient")
	defer span.End()

	query := `
			INSERT INTO %s (%s) VALUES %s
			ON CONFLICT ON CONSTRAINT withus_failed_sync_email_recipient_pk
			DO UPDATE SET
				email_address = EXCLUDED.email_address,
				last_updated_date = EXCLUDED.last_updated_date,
				last_updated_by = EXCLUDED.last_updated_by,
				is_archived = EXCLUDED.is_archived
		`
	err := dbeureka.BulkUpsert(ctx, db, query, items)
	if err != nil {
		return fmt.Errorf("LearningHistoryDataSyncRepo.BulkUpsertFailedSyncEmailRecipient: %w", err)
	}

	return nil
}

type WithusDataRow struct {
	CustomerNumber    pgtype.Text
	StudentNumber     pgtype.Text
	MaterialCode      pgtype.Text
	PaperCount        pgtype.Text
	Score             pgtype.Int4
	DateSubmitted     pgtype.Text
	ApproverID        pgtype.Text
	PaperApprovalDate pgtype.Text
	PerspectiveScore  pgtype.JSONB
	IsResubmission    pgtype.Bool
}

func (r *LearningHistoryDataSyncRepo) RetrieveWithusData(ctx context.Context, db database.QueryExecer) ([]*WithusDataRow, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningHistoryDataSyncRepo.RetrieveWithusData")
	defer span.End()

	query := `
	-- Step 1: Get scores per tag per submission
	with scores_per_tag AS(
			select  s.student_external_id,
							s.student_id,
							substring(u.email from '(.*)@')         as customer_number,
							c.course_partner_id                   	as material_code,
							els.learning_material_id                as lo_id,
							wmeli.material_code                     as paper_count,
							els.total_point                         as score,
							els.created_at                          as date_submitted,
							els.last_action_by                      as approver_manabie_user_id,
							els.last_action_at                      as paper_approval_date,
							els.last_action,
							els.status                              as submission_status,
							sum(q.point)                            as total_points_per_tag,
							sum(coalesce(elss.point, elsa.point))   as earned_points_per_tag,
							nullif(wmqt.manabie_tag_id, '')         as tag_id
	
			from		students s
			join        users u on (s.student_id = u.user_id)
			join        course_students cs on (cs.student_id = s.student_id)
			join        course_study_plans csp on (csp.course_id = cs.course_id)
			join        courses c on (csp.course_id = c.course_id)
			join        individual_study_plans_view ispv on (s.student_id = ispv.student_id and ispv.study_plan_id = csp.study_plan_id)
			join        exam_lo el on (el.learning_material_id = ispv.learning_material_id)
			join        withus_mapping_exam_lo_id wmeli on (wmeli.exam_lo_id = el.learning_material_id)
			join        exam_lo_submission els on (el.learning_material_id = els.learning_material_id and els.student_id = cs.student_id and csp.study_plan_id = els.study_plan_id)
			join        exam_lo_submission_answer elsa on (elsa.submission_id = els.submission_id)
			join        quizzes q on q.external_id=elsa.quiz_id
			left join   exam_lo_submission_score elss on (elsa.quiz_id = elss.quiz_id and elsa.submission_id = elss.submission_id)
			left join  	withus_mapping_question_tag wmqt on wmqt.manabie_tag_id = any(q.question_tag_ids)
	
			-- Course ID should have a Material Code defined in Master Management
			where c.course_partner_id <> ''
			-- LO ID should have a Paper Count defined in Master Management
			and wmeli.material_code <> ''
			-- Include only Exam LOs with Approve Grading ENABLED
			and el.approve_grading = true
			-- Include only PASSED/COMPLETED (FAILED submissions should not be sent), Include REMANDED submissions
			and (
				els."result" = 'EXAM_LO_SUBMISSION_PASSED' or
				els."result" = 'EXAM_LO_SUBMISSION_COMPLETED' or
				(els."result" = 'EXAM_LO_SUBMISSION_WAITING_FOR_GRADE' and (els.last_action = 'APPROVE_ACTION_APPROVED' or els.last_action='APPROVE_ACTION_REJECTED'))
			)
			-- Include only submissions made in the last 14 days
			and els.updated_at >= (((now() at time zone 'jst')::date + time '04:30' - interval '14' day) at time zone 'jst')::timestamp
			and els.updated_at < (((now() at time zone 'jst')::date + time '04:30') at time zone 'jst')::timestamp
			-- Include only quizzes status APPROVED
			and q.status = 'QUIZ_STATUS_APPROVED'
			
			group by    s.student_external_id, 
									s.student_id, 
									substring(u.email from '(.*)@'), 
									c.course_partner_id, 
									els.learning_material_id,
									wmeli.material_code, 
									els.total_point, 
									els.created_at,
									els.last_action_by, 
									els.last_action_at, 
									els.last_action,
									els.status, 
									tag_id
	), learning_history_denormalized as(
	-- Step 2: Concatenate scores per tag to format as [TagName]:[EarnedPoints]/[TotalPoints]
			select  r.student_external_id,
						r.student_id,
						r.customer_number,
						r.material_code,
						r.paper_count,
						r.earned_points_per_tag,
						r.date_submitted,
						r.submission_status,
						r.approver_manabie_user_id,
						r.paper_approval_date,
						r.last_action,
						(jsonb_agg(
                            jsonb_build_object('tag_id', r.tag_id, 'score', ':' || r.earned_points_per_tag || '/' || r.total_points_per_tag)) filter ( where r.tag_id is not null )
                        ) as perspective_score,
                        (jsonb_agg(
                            jsonb_build_object('tag_id', r.tag_id, 'score', ':' || '/' || r.total_points_per_tag)) filter ( where r.tag_id is not null )
                        ) as perspective_score_resubmmit,
						(select exists (select 1 from (select   * 
																					from     exam_lo_submission els2 
																	where    els2.learning_material_id = r.lo_id
																	and      els2.student_id = r.student_id
																	and      els2.created_at < r.date_submitted
																	order by created_at asc
																	limit 1) as first_submit
											where first_submit.result = 'EXAM_LO_SUBMISSION_FAILED')) as is_resubmission
				from    scores_per_tag as r
				group by    r.student_external_id, 
										r.student_id, 
										r.customer_number, 
										r.material_code, 
										r.lo_id, r.paper_count, 
										r.earned_points_per_tag,
										r.submission_status,
										r.date_submitted, 
										r.approver_manabie_user_id, 
										r.paper_approval_date, 
										r.last_action,
										r.tag_id
	)
	-- Step 3: Concatenate all perspective scores in one row as  [TagName]:[EarnedPoints]/[TotalPoints]$[TagName]:[EarnedPoints]/[TotalPoints]$[TagName]:[EarnedPoints]/[TotalPoints]
	select  lh.customer_number,
				lh.student_external_id as student_number,
				lh.material_code,
				lh.paper_count,
				-- If student only passed during resubmission, set score as 30
				-- If student is remanded, set score as 999
				cast(case   when (lh.is_resubmission and lh.submission_status = 'SUBMISSION_STATUS_RETURNED') then 30
										when ((lh.submission_status = 'SUBMISSION_STATUS_NOT_MARKED' or lh.submission_status = 'SUBMISSION_STATUS_IN_PROGRESS' or lh.submission_status = 'SUBMISSION_STATUS_MARKED') and lh.last_action = 'APPROVE_ACTION_REJECTED') then 999
										else sum(lh.earned_points_per_tag)
						end as integer) as score,
				to_char(lh.date_submitted at time zone 'jst', 'YYYY/FMMM/FMDD') as date_submitted,
				case    
					when (lh.submission_status = 'SUBMISSION_STATUS_RETURNED' and last_action = 'APPROVE_ACTION_APPROVED') or lh.last_action = 'APPROVE_ACTION_REJECTED'
						then lh.approver_manabie_user_id
					else null
				end as approver_manabie_user_id,
				case    
					when (lh.submission_status = 'SUBMISSION_STATUS_RETURNED' and last_action = 'APPROVE_ACTION_APPROVED') or lh.last_action = 'APPROVE_ACTION_REJECTED'
						then to_char(lh.paper_approval_date at time zone 'jst', 'YYYY/FMMM/FMDD')
					else    null
				end as paper_aproveal_date,
				case    
					when lh.is_resubmission
						then coalesce(jsonb_agg(lh.perspective_score_resubmmit->0) filter ( where lh.perspective_score_resubmmit is not null), '[]') ::jsonb    
					else coalesce(jsonb_agg(lh.perspective_score->0) filter ( where lh.perspective_score is not null), '[]') ::jsonb
				end as perspective_score,
				lh.is_resubmission
	from learning_history_denormalized lh

	group by    lh.customer_number,
				lh.student_external_id,
				lh.material_code,
				lh.paper_count,
				lh.is_resubmission,
				lh.submission_status,
				lh.last_action,
				lh.date_submitted,
				lh.approver_manabie_user_id,
				lh.paper_approval_date;
	`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveWithusData.Query: %w", err)
	}
	defer rows.Close()

	resp := make([]*WithusDataRow, 0)
	for rows.Next() {
		r := new(WithusDataRow)
		err = rows.Scan(&r.CustomerNumber, &r.StudentNumber, &r.MaterialCode, &r.PaperCount, &r.Score, &r.DateSubmitted, &r.ApproverID, &r.PaperApprovalDate, &r.PerspectiveScore, &r.IsResubmission)
		if err != nil {
			return nil, fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveWithusData.Scan: %w", err)
		}
		resp = append(resp, r)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveWithusData.Err: %w", err)
	}

	return resp, nil
}
