UPDATE study_plans sp SET resource_path = sp.school_id::TEXT
WHERE
	(sp.resource_path IS NULL OR LENGTH(sp.resource_path)=0);

UPDATE student_latest_submissions sls SET resource_path = sps.resource_path
FROM study_plan_items sps
WHERE sls.study_plan_item_id = sps.study_plan_item_id
  AND(sps.resource_path IS NOT NULL AND LENGTH(sps.resource_path)!=0)
	AND(sls.resource_path IS NULL OR LENGTH(sls.resource_path)=0);

UPDATE student_study_plans ssp SET resource_path = sp.school_id::TEXT
FROM study_plans sp
WHERE ssp.study_plan_id = sp.study_plan_id
  AND sp.school_id IS NOT NULL
	AND(ssp.resource_path IS NULL OR LENGTH(ssp.resource_path)=0);

UPDATE course_study_plans csp SET resource_path = sp.school_id::TEXT
FROM study_plans sp
WHERE csp.study_plan_id = sp.study_plan_id
  AND sp.school_id IS NOT NULL
	AND(csp.resource_path IS NULL OR LENGTH(csp.resource_path)=0);

UPDATE course_students cs SET resource_path = csp.resource_path
FROM course_study_plans csp
WHERE csp.course_id = cs.course_id
  AND(csp.resource_path IS NOT NULL AND LENGTH(csp.resource_path)!=0)
	AND(cs.resource_path IS NULL OR LENGTH(cs.resource_path)=0);

UPDATE student_submissions ss SET resource_path = spi.resource_path
FROM study_plan_items spi
WHERE ss.study_plan_item_id = spi.study_plan_item_id
  AND(spi.resource_path IS NOT NULL AND LENGTH(spi.resource_path)!=0)
	AND(ss.resource_path IS NULL OR LENGTH(ss.resource_path)=0);

UPDATE topics_assignments ta SET resource_path = aspi.resource_path
FROM assignment_study_plan_items aspi
WHERE ta.assignment_id = aspi.assignment_id
  AND(aspi.resource_path IS NOT NULL AND LENGTH(aspi.resource_path)!=0)
	AND(ta.resource_path IS NULL OR LENGTH(ta.resource_path)=0);

UPDATE course_classes cc SET resource_path = csp.resource_path
FROM course_study_plans csp
WHERE csp.course_id = cc.course_id
  AND(csp.resource_path IS NOT NULL AND LENGTH(csp.resource_path)!=0)
	AND(cc.resource_path IS NULL OR LENGTH(cc.resource_path)=0);

UPDATE class_students cs SET resource_path = cc.resource_path
FROM course_classes cc
WHERE cs.class_id = cc.class_id
  AND(cc.resource_path IS NOT NULL AND LENGTH(cc.resource_path)!=0)
	AND(cs.resource_path IS NULL OR LENGTH(cs.resource_path)=0);

UPDATE assign_study_plan_tasks aspt SET resource_path = sp.resource_path
FROM study_plans sp
WHERE aspt.study_plan_ids[1] = sp.study_plan_id
    AND (sp.resource_path IS NOT NULL AND LENGTH(sp.resource_path)!=0)
    AND (aspt.resource_path IS NULL OR LENGTH(aspt.resource_path)=0);

UPDATE student_submission_grades ssg SET resource_path = ss.resource_path
FROM student_submissions ss
WHERE ssg.student_submission_id = ss.student_submission_id
    AND (ss.resource_path IS NOT NULL AND LENGTH(ss.resource_path)!=0)
    AND (ssg.resource_path IS NULL OR LENGTH(ssg.resource_path)=0);

UPDATE assignments a SET resource_path = ta.resource_path
FROM topics_assignments ta
WHERE ta.topic_id = a.content ->> 'topic_id'
    AND (ta.resource_path IS NOT NULL AND LENGTH(ta.resource_path)!=0)
    AND (a.resource_path IS NULL OR LENGTH(a.resource_path)=0);
