package repositories

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type PresetStudyPlanRepo struct{}

func (r *PresetStudyPlanRepo) RetrievePresetStudyPlans(ctx context.Context, db database.QueryExecer, name, country, subject string, grade int) ([]*entities.PresetStudyPlan, error) {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanRepo.RetrievePresetStudyPlans")
	defer span.End()

	fields := database.GetFieldNames(&entities.PresetStudyPlan{})
	var args []interface{}
	query := fmt.Sprintf("SELECT %s FROM preset_study_plans", strings.Join(fields, ","))

	conditions := ""
	if name != "" {
		conditions += fmt.Sprintf(" AND name = $%d", len(args)+1)
		args = append(args, name)
	}
	if country != "" {
		conditions += fmt.Sprintf(" AND country = $%d", len(args)+1)
		args = append(args, country)
	}
	if subject != "" {
		conditions += fmt.Sprintf(" AND subject = $%d", len(args)+1)
		args = append(args, subject)
	}
	if grade != -1 {
		conditions += fmt.Sprintf(" AND grade = $%d", len(args)+1)
		args = append(args, grade)
	}

	if strings.HasPrefix(conditions, " AND") {
		// replace prefix AND = WHERE
		conditions = " WHERE" + conditions[4:]
	}

	query += conditions + " ORDER BY subject DESC"
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var pp []*entities.PresetStudyPlan
	for rows.Next() {
		p := new(entities.PresetStudyPlan)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp = append(pp, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return pp, nil
}

func (r *PresetStudyPlanRepo) RetrievePresetStudyPlanWeeklies(ctx context.Context, db database.QueryExecer, presetStudyPlanID pgtype.Text) ([]*entities.PresetStudyPlanWeekly, error) {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanRepo.RetrievePresetStudyPlanWeeklies")
	defer span.End()

	fields := database.GetFieldNames(&entities.PresetStudyPlanWeekly{})
	query := fmt.Sprintf("SELECT %s FROM preset_study_plans_weekly WHERE preset_study_plan_id = $1 ORDER BY week ASC", strings.Join(fields, ","))

	rows, err := db.Query(ctx, query, &presetStudyPlanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pp []*entities.PresetStudyPlanWeekly
	for rows.Next() {
		p := new(entities.PresetStudyPlanWeekly)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, err
		}
		pp = append(pp, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pp, nil
}

func (r *PresetStudyPlanRepo) UpdatePresetStudyPlanWeeklyEndTime(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, endTime pgtype.Timestamptz) error {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanRepo.UpdatePresetStudyPlanWeeklyEndTime")
	defer span.End()

	cmdTag, err := db.Exec(ctx, `UPDATE preset_study_plans_weekly SET end_date =$1 WHERE lesson_id =$2 AND end_date > $1`, &endTime, &lessonID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("UpdatePresetStudyPlanWeeklyEndTime: Can't update lesson end time")
	}
	return nil
}

type PlanDetail struct {
	PresetStudyPlanID       pgtype.Text
	PresetStudyPlanWeeklyID pgtype.Text
	TopicID                 pgtype.Text
	StartDate               pgtype.Timestamptz
}

type PlanWithStartDate struct {
	*entities.PresetStudyPlan
	Week      *pgtype.Int2
	StartDate *pgtype.Timestamptz
}

func (r *PresetStudyPlanRepo) RetrieveStudentPresetStudyPlans(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz) ([]*PlanWithStartDate, error) {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanRepo.RetrieveStudentPresetStudyPlans")
	defer span.End()

	p := new(PlanWithStartDate)
	p.PresetStudyPlan = new(entities.PresetStudyPlan)
	pFields := database.GetFieldNames(p.PresetStudyPlan)

	fields := make([]string, 0, len(pFields))
	for _, f := range pFields {
		fields = append(fields, p.TableName()+"."+f)
	}

	args := []interface{}{&studentID}
	sub := `SELECT assignments.preset_study_plan_id, MIN(preset_study_plans_weekly.week) as week , MIN(assignments.start_date) as start_date
			FROM assignments
			JOIN student_assignments ON student_assignments.assignment_id = assignments.assignment_id
			JOIN preset_study_plans_weekly ON (preset_study_plans_weekly.preset_study_plan_weekly_id = assignments.preset_study_plan_weekly_id)
			WHERE student_id=$1 AND assignments.deleted_at IS NULL`

	if from != nil {
		sub += fmt.Sprintf(" AND assignments.start_date >= $%d", len(args)+1)
		args = append(args, from)
	}
	if to != nil {
		sub += fmt.Sprintf(" AND assignments.start_date <= $%d", len(args)+1)
		args = append(args, to)
	}
	sub += ` GROUP BY assignments.preset_study_plan_id`

	query := fmt.Sprintf(`SELECT %s, sub.start_date, sub.week FROM preset_study_plans JOIN (%s) sub
		ON preset_study_plans.preset_study_plan_id = sub.preset_study_plan_id`, strings.Join(fields, ","), sub)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var pp []*PlanWithStartDate
	for rows.Next() {
		p := &PlanWithStartDate{
			PresetStudyPlan: new(entities.PresetStudyPlan),
			Week:            new(pgtype.Int2),
			StartDate:       new(pgtype.Timestamptz),
		}

		f := append(database.GetScanFields(p.PresetStudyPlan, pFields), p.StartDate, p.Week)
		if err := rows.Scan(f...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp = append(pp, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return pp, nil
}

type Topic struct {
	Topic            entities.Topic
	StartDate        pgtype.Timestamptz
	TotalLOs         pgtype.Int4
	TotalFinishedLOs pgtype.Int4
	AssignedBy       entities.User
	EndDate          pgtype.Timestamptz
}

func (r *PresetStudyPlanRepo) RetrieveStudentPresetStudyPlanWeeklies(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, isActive bool) ([]Topic, error) {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanRepo.RetrieveStudentPresetStudyPlanWeeklies")
	defer span.End()

	args := []interface{}{&studentID}
	query := `SELECT topics.topic_id, topics.name, start_date, topics.total_los, stc.total_finished_los
		FROM students_study_plans_weekly sspw
		JOIN preset_study_plans_weekly pspw ON sspw.preset_study_plan_weekly_id = pspw.preset_study_plan_weekly_id
		JOIN topics ON topics.topic_id = pspw.topic_id
		LEFT JOIN students_topics_completeness stc ON stc.student_id = sspw.student_id AND stc.topic_id = topics.topic_id
		WHERE sspw.student_id = $1 AND start_date IS NOT NULL AND topics.deleted_at IS NOT NULL`

	if isActive {
		query += ` AND (stc.is_completed IS NULL OR stc.is_completed=FALSE)`
	}
	if from != nil {
		args = append(args, from)
		query += fmt.Sprintf(" AND start_date >= $%d", len(args))
	}
	if to != nil {
		args = append(args, to)
		query += fmt.Sprintf(" AND start_date <= $%d", len(args))
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var pp []Topic
	for rows.Next() {
		var p Topic
		if err := rows.Scan(&p.Topic.ID, &p.Topic.Name, &p.StartDate, &p.TotalLOs, &p.TotalFinishedLOs); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp = append(pp, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return pp, nil
}

func (r *PresetStudyPlanRepo) diffPlans(before, after map[string][]PlanData) (changedPlans []string) {
	for k := range before {
		sort.Slice(before[k], func(i, j int) bool { return before[k][i].Week < before[k][j].Week })
	}
	for k := range after {
		sort.Slice(after[k], func(i, j int) bool { return after[k][i].Week < after[k][j].Week })
	}

	for planID, v := range after {
		if b, ok := before[planID]; ok && !reflect.DeepEqual(b, v) {
			changedPlans = append(changedPlans, planID)
		}
	}
	return
}

func (r *PresetStudyPlanRepo) getChangedPlans(ctx context.Context, db database.QueryExecer, presetStudyPlanWeeklies []*entities.PresetStudyPlanWeekly) ([]string, error) {
	// get preset study plan data before update
	plansBefore, err := r.retrievePresetStudyPlanData(ctx, db)
	if err != nil {
		return nil, errors.Wrap(err, "s.PresetStudyPlanRepo.RetrievePresetStudyPlanData")
	}

	plansAfter := make(map[string][]PlanData)
	for _, w := range presetStudyPlanWeeklies {
		planID := w.PresetStudyPlanID.String
		plansAfter[planID] = append(plansAfter[planID], PlanData{
			PresetStudyPlanWeeklyID: w.ID.String,
			TopicID:                 w.TopicID.String,
			Week:                    w.Week.Int,
		})
	}

	changedPlans := r.diffPlans(plansBefore, plansAfter)
	if len(changedPlans) == 0 {
		return nil, nil
	}
	return changedPlans, nil
}

func (r *PresetStudyPlanRepo) CreatePresetStudyPlan(ctx context.Context, db database.Ext, presetStudyPlans []*entities.PresetStudyPlan) error {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanRepo.CreatePresetStudyPlan")
	defer span.End()

	queueInsertPresetStudyPlanFn := func(b *pgx.Batch, p *entities.PresetStudyPlan) {
		fieldNames := []string{"preset_study_plan_id", "name", "country", "grade", "subject", "created_at", "updated_at", "start_date"}
		placeHolders := "$1, $2, $3, $4, $5, $6, $7, $8"

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT preset_study_plans_pk DO UPDATE SET updated_at = $7, start_date = $8", p.TableName(), strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(p, fieldNames)...)
	}

	now := time.Now()

	b := &pgx.Batch{}
	for _, presetStudyPlan := range presetStudyPlans {
		_ = presetStudyPlan.CreatedAt.Set(now)
		_ = presetStudyPlan.UpdatedAt.Set(now)
		queueInsertPresetStudyPlanFn(b, presetStudyPlan)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(presetStudyPlans); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("preset study plan not inserted")
		}
	}

	return nil
}

func (r *PresetStudyPlanRepo) CreatePresetStudyPlanWeekly(ctx context.Context, db database.QueryExecer, presetStudyPlanWeeklies []*entities.PresetStudyPlanWeekly) error {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanRepo.CreatePresetStudyPlanWeekly")
	defer span.End()

	// calculate which preset plans are changed to
	// reassign those changed plans to students
	changedPlans, err := r.getChangedPlans(ctx, db, presetStudyPlanWeeklies)
	if err != nil {
		return errors.Wrap(err, "r.getChangedPlans")
	}

	if len(changedPlans) > 0 {
		rows, err := db.Query(ctx, "SELECT preset_study_plan_weekly_id FROM preset_study_plans_weekly WHERE preset_study_plan_id = ANY($1)", &changedPlans)
		if err != nil {
			return errors.Wrap(err, "tx.QueryEx")
		}
		defer rows.Close()

		var weeklyIDs []string
		for rows.Next() {
			var id pgtype.Text
			if err := rows.Scan(&id); err != nil {
				return errors.Wrap(err, "rows.Scan")
			}
			weeklyIDs = append(weeklyIDs, id.String)
		}
		if err := rows.Err(); err != nil {
			return errors.Wrap(err, "rows.Err")
		}

		if _, err := db.Exec(ctx, "DELETE FROM preset_study_plans_weekly WHERE preset_study_plan_id = ANY($1)", &changedPlans); err != nil {
			return errors.Wrap(err, "tx.ExecEx")
		}
	}

	queueInsertPresetStudyPlanWeekliesFn := func(b *pgx.Batch, p *entities.PresetStudyPlanWeekly) {
		fieldNames := []string{"preset_study_plan_weekly_id", "preset_study_plan_id", "topic_id", "week", "created_at", "updated_at"}
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT weekly_preset_study_plans_pk DO UPDATE SET week = $4, updated_at = $6", p.TableName(), strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(p, fieldNames)...)
	}

	now := time.Now()

	batch := func() error {
		b := &pgx.Batch{}

		for _, presetStudyPlanWeekly := range presetStudyPlanWeeklies {
			_ = presetStudyPlanWeekly.CreatedAt.Set(now)
			_ = presetStudyPlanWeekly.UpdatedAt.Set(now)
			queueInsertPresetStudyPlanWeekliesFn(b, presetStudyPlanWeekly)
		}

		batchResults := db.SendBatch(ctx, b)
		defer batchResults.Close()

		for i := 0; i < len(presetStudyPlanWeeklies); i++ {
			ct, err := batchResults.Exec()
			if err != nil {
				return errors.Wrap(err, "batchResults.Exec")
			}
			if ct.RowsAffected() != 1 {
				return fmt.Errorf("preset study plan weeklies not inserted")
			}
		}

		return nil
	}

	if err := batch(); err != nil {
		return err
	}

	return nil
}

func (r *PresetStudyPlanRepo) insertPresetStudyPlanBatch(ctx context.Context, db database.QueryExecer, presetStudyPlans []*entities.PresetStudyPlan) error {
	now := time.Now()
	queueInsertPresetStudyPlanFn := func(b *pgx.Batch, p *entities.PresetStudyPlan) {
		fieldNames := []string{"preset_study_plan_id", "name", "country", "grade", "subject", "created_at", "updated_at", "start_date"}
		placeHolders := "$1, $2, $3, $4, $5, $6, $7, $8"

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT preset_study_plans_pk DO UPDATE SET updated_at = $7, start_date = $8", p.TableName(), strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(p, fieldNames)...)
	}

	b := &pgx.Batch{}
	for _, presetStudyPlan := range presetStudyPlans {
		_ = presetStudyPlan.CreatedAt.Set(now)
		_ = presetStudyPlan.UpdatedAt.Set(now)
		queueInsertPresetStudyPlanFn(b, presetStudyPlan)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(presetStudyPlans); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("preset study plan not inserted")
		}
	}
	return nil
}

func (r *PresetStudyPlanRepo) insertPresetStudyPlanWeekliesBatch(ctx context.Context, db database.QueryExecer, presetStudyPlanWeeklies []*entities.PresetStudyPlanWeekly) error {
	now := time.Now()
	queueInsertPresetStudyPlanWeekliesFn := func(b *pgx.Batch, p *entities.PresetStudyPlanWeekly) {
		fieldNames := []string{"preset_study_plan_weekly_id", "preset_study_plan_id", "topic_id", "week", "created_at", "updated_at"}
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT weekly_preset_study_plans_pk DO UPDATE SET week = $4, updated_at = $6", p.TableName(), strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(p, fieldNames)...)
	}

	b := &pgx.Batch{}
	for _, presetStudyPlanWeekly := range presetStudyPlanWeeklies {
		_ = presetStudyPlanWeekly.CreatedAt.Set(now)
		_ = presetStudyPlanWeekly.UpdatedAt.Set(now)
		queueInsertPresetStudyPlanWeekliesFn(b, presetStudyPlanWeekly)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(presetStudyPlanWeeklies); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("preset study plan weeklies not inserted")
		}
	}

	return nil
}

func (r *PresetStudyPlanRepo) BulkImport(ctx context.Context, db database.QueryExecer, presetStudyPlans []*entities.PresetStudyPlan, presetStudyPlanWeeklies []*entities.PresetStudyPlanWeekly) error {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanRepo.BulkImport")
	defer span.End()

	// calculate which preset plans are changed to
	// reassign those changed plans to students
	changedPlans, err := r.getChangedPlans(ctx, db, presetStudyPlanWeeklies)
	if err != nil {
		return errors.Wrap(err, "r.getChangedPlans")
	}

	if len(changedPlans) > 0 {
		rows, err := db.Query(ctx, "SELECT preset_study_plan_weekly_id FROM preset_study_plans_weekly WHERE preset_study_plan_id = ANY($1)", &changedPlans)
		if err != nil {
			return errors.Wrap(err, "tx.QueryEx")
		}
		defer rows.Close()

		var weeklyIDs []string
		for rows.Next() {
			var id pgtype.Text
			if err := rows.Scan(&id); err != nil {
				return errors.Wrap(err, "rows.Scan")
			}
			weeklyIDs = append(weeklyIDs, id.String)
		}
		if err := rows.Err(); err != nil {
			return errors.Wrap(err, "rows.Err")
		}
		if _, err := db.Exec(ctx, "DELETE FROM preset_study_plans_weekly WHERE preset_study_plan_id = ANY($1)", &changedPlans); err != nil {
			return errors.Wrap(err, "tx.ExecEx")
		}
	}

	if err := r.insertPresetStudyPlanBatch(ctx, db, presetStudyPlans); err != nil {
		return err
	}

	if err := r.insertPresetStudyPlanWeekliesBatch(ctx, db, presetStudyPlanWeeklies); err != nil {
		return err
	}

	return nil
}

type AheadTopic struct {
	TopicID pgtype.Text
	Week    pgtype.Int2
}

func (r *PresetStudyPlanRepo) retrieveAssignedTopics(ctx context.Context, db database.QueryExecer, studentID, presetStudyPlanID pgtype.Text, startDate pgtype.Date) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanRepo.retrieveAssignedTopics")
	defer span.End()

	query := `SELECT topic_id FROM students_study_plans_weekly sspw
	INNER JOIN preset_study_plans_weekly pspw ON sspw.preset_study_plan_weekly_id = pspw.preset_study_plan_weekly_id
	WHERE student_id = $1 AND DATE(start_date) = $2 AND pspw.preset_study_plan_id = $3`

	rows, err := db.Query(ctx, query, &studentID, &startDate, &presetStudyPlanID)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var topics []string
	for rows.Next() {
		var topicID pgtype.Text
		if err := rows.Scan(&topicID); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		topics = append(topics, topicID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return topics, nil
}

func (r *PresetStudyPlanRepo) retrieveFinishedTopics(ctx context.Context, db database.QueryExecer, studentID, presetStudyPlanID pgtype.Text, startDate pgtype.Date) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanRepo.retrieveFinishedTopics")
	defer span.End()

	query := `SELECT pspw.topic_id FROM preset_study_plans_weekly pspw
	INNER JOIN students_study_plans_weekly sspw ON sspw.preset_study_plan_weekly_id = pspw.preset_study_plan_weekly_id
	INNER JOIN (
		SELECT lo.topic_id, COUNT(*) AS finished_los
		FROM learning_objectives lo
		INNER JOIN students_learning_objectives_completeness sloc ON sloc.lo_id = lo.lo_id
		WHERE sloc.is_finished_quiz IS TRUE AND sloc.student_id = $1 AND lo.deleted_at IS NULL
		GROUP BY lo.topic_id
	) sub ON sub.topic_id = pspw.topic_id
	INNER JOIN (
		SELECT topic_id, COUNT(lo_id) AS los FROM learning_objectives WHERE learning_objectives.deleted_at IS NULL GROUP BY topic_id
	) sub2 ON sub2.topic_id = sub.topic_id AND sub.finished_los = sub2.los
	WHERE DATE(start_date) = $2 AND preset_study_plan_id = $3
	GROUP BY pspw.topic_id`

	rows, err := db.Query(ctx, query, &studentID, &startDate, &presetStudyPlanID)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var topics []string
	for rows.Next() {
		var topicID pgtype.Text
		if err := rows.Scan(&topicID); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		topics = append(topics, topicID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return topics, nil
}

func (r *PresetStudyPlanRepo) RetrieveStudyAheadTopicsOfPresetStudyPlan(ctx context.Context, db database.QueryExecer, studentID, presetStudyPlanID pgtype.Text) ([]AheadTopic, error) {
	query := `SELECT MAX(DATE(start_date)) FROM students_study_plans_weekly sspw
	INNER JOIN preset_study_plans_weekly pspw ON sspw.preset_study_plan_weekly_id = sspw.preset_study_plan_weekly_id
	WHERE student_id = $1 AND preset_study_plan_id = $2`

	var maxStartDate pgtype.Date
	if err := db.QueryRow(ctx, query, &studentID, &presetStudyPlanID).Scan(&maxStartDate); err != nil {
		return nil, errors.Wrap(err, "r.Wrapper.QueryRowEx")
	}

	startDateThisWeek := timeutil.StartWeek()

	var startDate pgtype.Date
	startDate.Set(startDateThisWeek)

	inArray := func(s string, ss []string) bool {
		for _, r := range ss {
			if r == s {
				return true
			}
		}
		return false
	}
	isEqual := func(assignedTopics, finishedTopics []string) bool {
		for _, a := range assignedTopics {
			if !inArray(a, finishedTopics) {
				return false
			}
		}
		return true
	}

	for {
		assignedTopics, err := r.retrieveAssignedTopics(ctx, db, studentID, presetStudyPlanID, startDate)
		if err != nil {
			return nil, errors.Wrap(err, "r.retrieveAssignedTopics")
		}
		finishedTopics, err := r.retrieveFinishedTopics(ctx, db, studentID, presetStudyPlanID, startDate)
		if err != nil {
			return nil, errors.Wrap(err, "r.retrieveFinishedTopics")
		}

		if len(assignedTopics) == len(finishedTopics) && isEqual(assignedTopics, finishedTopics) {
			startDate.Set(startDate.Time.Add(7 * 24 * time.Hour))
			if startDate.Time.After(maxStartDate.Time) {
				return nil, nil
			}
			continue
		}

		switch {
		case startDate.Time.Equal(startDateThisWeek):
			return nil, nil

		case startDate.Time.After(startDateThisWeek):
			var aheadTopics []AheadTopic
			for _, a := range assignedTopics {
				if !inArray(a, finishedTopics) {
					aheadTopics = append(aheadTopics, AheadTopic{
						TopicID: database.Text(a),
					})
				}
			}
			return aheadTopics, nil
		default:
			return nil, nil
		}
	}
}

type PlanData struct {
	PresetStudyPlanWeeklyID string
	TopicID                 string
	Week                    int16
}

func (r *PresetStudyPlanRepo) retrievePresetStudyPlanData(ctx context.Context, db database.QueryExecer) (map[string][]PlanData, error) {
	query := "SELECT psp.preset_study_plan_id, preset_study_plan_weekly_id, topic_id, week FROM preset_study_plans psp INNER JOIN preset_study_plans_weekly pspw ON psp.preset_study_plan_id = pspw.preset_study_plan_id"
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	plans := make(map[string][]PlanData) // preset study plan ID => plan data
	for rows.Next() {
		var (
			planID   = new(pgtype.Text)
			weeklyID = new(pgtype.Text)
			topicID  = new(pgtype.Text)
			week     = new(pgtype.Int2)
		)
		if err := rows.Scan(planID, weeklyID, topicID, week); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		plans[planID.String] = append(plans[planID.String], PlanData{
			PresetStudyPlanWeeklyID: weeklyID.String,
			TopicID:                 topicID.String,
			Week:                    week.Int,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return plans, nil
}

type AssignedPresetStudyPlan struct {
	PresetStudyPlanID *pgtype.Text
	StartWeek         *pgtype.Int2
	StartDate         *pgtype.Timestamptz
}

func (r *PresetStudyPlanRepo) RetrieveStudentCompletedTopics(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, topicIDs pgtype.TextArray) ([]*entities.Topic, error) {
	ctx, span := interceptors.StartSpan(ctx, "PresetStudyPlanRepo.RetrieveStudentCompletedTopic")
	defer span.End()

	topicEn := &entities.Topic{}
	topicComplete := &entities.StudentTopicCompleteness{}

	fields := database.GetFieldNames(topicEn)

	args := []interface{}{&studentID, &topicIDs}
	query := fmt.Sprintf(`SELECT topics.%s FROM %s as stc JOIN %s as topics ON topics.topic_id = stc.topic_id WHERE stc.student_id = $1 AND stc.topic_id = ANY($2) and is_completed=true AND topics.deleted_at IS NULL ORDER BY updated_at desc`,
		strings.Join(fields, ", topics."), topicComplete.TableName(), topicEn.TableName())

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	topics := make([]*entities.Topic, 0, len(topicIDs.Elements))
	for rows.Next() {
		topic := &entities.Topic{}
		if err := rows.Scan(database.GetScanFields(topic, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		topics = append(topics, topic)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return topics, nil
}
