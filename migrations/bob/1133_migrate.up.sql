UPDATE books b SET resource_path = b.school_id
WHERE (b.resource_path IS NULL OR LENGTH(b.resource_path)=0);

UPDATE chapters c SET resource_path = c.school_id
WHERE (c.resource_path IS NULL OR LENGTH(c.resource_path)=0);

UPDATE topics t SET resource_path = t.school_id
WHERE (t.resource_path IS NULL OR LENGTH(t.resource_path)=0);

UPDATE learning_objectives lo SET resource_path = lo.school_id
WHERE (lo.resource_path IS NULL OR LENGTH(lo.resource_path)=0);

UPDATE quizzes q SET resource_path = q.school_id
WHERE (q.resource_path IS NULL OR LENGTH(q.resource_path)=0);

UPDATE quiz_sets qs SET resource_path = lo.resource_path
FROM learning_objectives lo
WHERE  qs.lo_id = lo.lo_id
    AND (lo.resource_path IS NOT NULL AND LENGTH(lo.resource_path)!=0)
    AND (qs.resource_path IS NULL OR LENGTH(qs.resource_path)=0);

UPDATE shuffled_quiz_sets sqs SET resource_path = qs.resource_path
FROM quiz_sets qs
WHERE sqs.original_quiz_set_id = qs.quiz_set_id
    AND (qs.resource_path IS NOT NULL AND LENGTH(qs.resource_path)!=0)
    AND (sqs.resource_path IS NULL OR LENGTH(sqs.resource_path)=0);

UPDATE students_learning_objectives_completeness sloc SET resource_path = lo.resource_path
FROM learning_objectives lo
WHERE lo.lo_id = sloc.lo_id
    AND (lo.resource_path IS NOT NULL AND LENGTH(lo.resource_path)!=0)
    AND (sloc.resource_path IS NULL OR LENGTH(sloc.resource_path)=0);

UPDATE students_learning_objectives_records slor SET resource_path = lo.resource_path
FROM learning_objectives lo
WHERE lo.lo_id = slor.lo_id
    AND (lo.resource_path IS NOT NULL AND LENGTH(lo.resource_path)!=0)
    AND (slor.resource_path IS NULL OR LENGTH(slor.resource_path)=0);

UPDATE flashcard_speeches fs SET resource_path = q.resource_path
FROM quizzes q
WHERE q.quiz_id = fs.quiz_id
    AND (q.resource_path IS NOT NULL AND LENGTH(q.resource_path)!=0)
    AND (fs.resource_path IS NULL OR LENGTH(fs.resource_path)=0);

UPDATE flashcard_progressions fp SET resource_path = lo.resource_path
FROM learning_objectives lo
WHERE lo.lo_id = fp.lo_id
    AND (lo.resource_path IS NOT NULL AND LENGTH(lo.resource_path)!=0)
    AND (fp.resource_path IS NULL OR LENGTH(fp.resource_path)=0);

UPDATE student_event_logs sel SET resource_path = lo.resource_path
FROM learning_objectives lo
WHERE lo.lo_id = sel.payload ->> 'lo_id'
    AND (lo.resource_path IS NOT NULL AND LENGTH(lo.resource_path)!=0)
    AND (sel.resource_path IS NULL OR LENGTH(sel.resource_path)=0);

 UPDATE books_chapters bc SET resource_path = b.resource_path
FROM books b
WHERE b.book_id = bc.book_id
  AND (b.resource_path IS NOT NULL AND LENGTH(b.resource_path)!=0)
  AND (bc.resource_path IS NULL OR LENGTH(bc.resource_path)=0);

UPDATE courses_books cb SET resource_path = b.resource_path
FROM books b
WHERE b.book_id = cb.book_id
  AND (b.resource_path IS NOT NULL AND LENGTH(b.resource_path)!=0)
  AND (cb.resource_path IS NULL OR LENGTH(cb.resource_path)=0);

UPDATE topics_learning_objectives tlo SET resource_path = lo.resource_path
FROM learning_objectives lo
WHERE lo.lo_id = tlo.lo_id
  AND (lo.resource_path IS NOT NULL AND LENGTH(lo.resource_path)!=0)
  AND (tlo.resource_path IS NULL OR LENGTH(tlo.resource_path)=0);

UPDATE student_learning_time_by_daily sltbd SET resource_path = s.school_id::TEXT
FROM students s
WHERE s.student_id = sltbd.student_id
    AND s.school_id IS NOT NULL
    AND(sltbd.resource_path IS NULL OR LENGTH(sltbd.resource_path)=0); 
