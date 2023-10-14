Feature: Retrieve course statistic

    Background: user sign in
        Given <exam_lo>a signed in "school admin"

    Scenario Outline: students do quiz one time and view grade book
        Given <exam_lo>"<num_student>" student login
        And <exam_lo>a school admin login
        And <exam_lo>a teacher login
        And <exam_lo>"school admin" has created a book with each "<num_los>" los with "<grade_to_pass_point>" point, "<num_ass>" assignments, "<num_topic>" topics, "<num_chap>" chapters, "<num_quiz>" quizzes
        And <exam_lo>"school admin" has created a course with a book
        And <exam_lo>"school admin" has created a studyplan for all student
        And <exam_lo>"school admin" has updated course duration for student
        And <exam_lo>"<num_student_do_test>" students do test and each student done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
        And <exam_lo>teacher updates "<result_1>" of exam lo
        When <exam_lo>retrieve topic statistic
        Then <exam_lo>our system returns correct topic statistic
        And <exam_lo>topic total assigned student is <topic_assigned>, completed students is <topic_completed_student>, average score is <topic_average_score>
        Then <exam_lo>retrieve grade book with "<setting>"
        And <exam_lo>our system returns "<num_student_grade>" items with "<exam_los>" exam los, "<completed_los>" completed los, "<grade_to_pass>" grade to pass items, "<passed>" items and "<num_exam_results>" items with "<point>" point, "<grade_point>" grade point, "<total_attempts>" total attempts

        # case 3 passed -> grade point = point of first exam lo
        Examples:
            | num_student | num_student_in_class | done_los | correct_quizzes | done_assignments | assignment_mark | skipped_topics | topic_assigned | topic_completed_student | topic_average_score | num_los | num_ass | num_topic | num_chap | num_quiz | num_student_do_test | setting             | num_student_grade | exam_los | completed_los | grade_to_pass | passed | num_exam_results | point | grade_point | total_attempts | grade_to_pass_point | result_1                     |
            | 2           | 0                    | 2        | 5               | 0                | 0               | 0              | 2              | 2                       | 100                 | 2       | 0       | 1         | 1        | 5        | 2                   | LATEST_SCORE        | 2                 | 2        | 2             | 2             | 0      | 2                | 5     | 5           | 1              | 4                   | EXAM_LO_SUBMISSION_COMPLETED |
            | 2           | 0                    | 2        | 3               | 0                | 0               | 0              | 2              | 2                       | 60                  | 2       | 0       | 1         | 1        | 5        | 2                   | LATEST_SCORE        | 2                 | 2        | 2             | 2             | 0      | 2                | 5     | 3           | 1              | 3                   | EXAM_LO_SUBMISSION_COMPLETED |
            | 2           | 0                    | 2        | 5               | 0                | 0               | 0              | 2              | 2                       | 100                 | 2       | 0       | 1         | 1        | 5        | 2                   | GRADE_TO_PASS_SCORE | 2                 | 2        | 2             | 2             | 2      | 2                | 5     | 5           | 1              | 3                   | EXAM_LO_SUBMISSION_PASSED    |

    Scenario Outline: students do quiz 2 times and view grade book
        Given <exam_lo>"<num_student>" student login
        And <exam_lo>a school admin login
        And <exam_lo>a teacher login
        And <exam_lo>"school admin" has created a book with each "<num_los>" los with "<grade_to_pass_point>" point, "<num_ass>" assignments, "<num_topic>" topics, "<num_chap>" chapters, "<num_quiz>" quizzes
        And <exam_lo>"school admin" has created a course with a book
        And <exam_lo>"school admin" has created a studyplan for all student
        And <exam_lo>"school admin" has updated course duration for student
        And <exam_lo>"<num_student_do_test>" students do test and each student done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
        And <exam_lo>"<num_student_do_test>" students do test and each student done "<done_los>" los with "<correct_quizzes_2>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
        And <exam_lo>teacher updates "<result_2>" of exam lo
        When <exam_lo>retrieve topic statistic
        Then <exam_lo>our system returns correct topic statistic
        And <exam_lo>topic total assigned student is <topic_assigned>, completed students is <topic_completed_student>, average score is <topic_average_score>
        Then <exam_lo>retrieve grade book with "<setting>"
        And <exam_lo>our system returns "<num_student_grade>" items with "<exam_los>" exam los, "<completed_los>" completed los, "<grade_to_pass>" grade to pass items, "<passed>" items and "<num_exam_results>" items with "<point>" point, "<grade_point>" grade point, "<total_attempts>" total attempts

        Examples:
            | num_student | num_student_in_class | done_los | correct_quizzes | done_assignments | assignment_mark | skipped_topics | topic_assigned | topic_completed_student | topic_average_score | num_los | num_ass | num_topic | num_chap | num_quiz | num_student_do_test | setting      | num_student_grade | exam_los | completed_los | grade_to_pass | passed | num_exam_results | point | grade_point | total_attempts | grade_to_pass_point | correct_quizzes_2 | result_2 |
            | 2           | 0                    | 2        | 1               | 0                | 0               | 0              | 2              | 2                       | 40                  | 2       | 0       | 1         | 1        | 5        | 2                   | LATEST_SCORE | 2                 | 2        | 2             | 2             | 2      | 2                | 5     | 2           | 2              | 4                   | 2                 | EXAM_LO_SUBMISSION_PASSED |

    Scenario Outline: students do quiz 2 times, teacher updates result of submission and view grade book
        Given <exam_lo>"<num_student>" student login
        And <exam_lo>a school admin login
        And <exam_lo>a teacher login
        And <exam_lo>"school admin" has created a book with each "<num_los>" los with "<grade_to_pass_point>" point, "<num_ass>" assignments, "<num_topic>" topics, "<num_chap>" chapters, "<num_quiz>" quizzes
        And <exam_lo>"school admin" has created a course with a book
        And <exam_lo>"school admin" has created a studyplan for all student
        And <exam_lo>"school admin" has updated course duration for student
        And <exam_lo>"<num_student_do_test>" students do test and each student done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
        And <exam_lo>teacher updates "<result_1>" of exam lo
        And <exam_lo>"<num_student_do_test>" students do test and each student done "<done_los>" los with "<correct_quizzes_2>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
        And <exam_lo>teacher updates "<result_2>" of exam lo
        When <exam_lo>retrieve topic statistic
        Then <exam_lo>our system returns correct topic statistic
        And <exam_lo>topic total assigned student is <topic_assigned>, completed students is <topic_completed_student>, average score is <topic_average_score>
        Then <exam_lo>retrieve grade book with "<setting>"
        And <exam_lo>our system returns "<num_student_grade>" items with "<exam_los>" exam los, "<completed_los>" completed los, "<grade_to_pass>" grade to pass items, "<passed>" items and "<num_exam_results>" items with "<point>" point, "<grade_point>" grade point, "<total_attempts>" total attempts

        Examples:
        # case 1 failed failed -> grade point = latest 
        # case 2 failed pass -> grade point = grade to pass point
            | num_student | num_student_in_class | done_los | correct_quizzes | done_assignments | assignment_mark | skipped_topics | topic_assigned | topic_completed_student | topic_average_score | num_los | num_ass | num_topic | num_chap | num_quiz | num_student_do_test | setting             | num_student_grade | exam_los | completed_los | grade_to_pass | passed | num_exam_results | point | grade_point | total_attempts | grade_to_pass_point | correct_quizzes_2 | result_1                  | result_2                  |
            | 2           | 0                    | 2        | 1               | 0                | 0               | 0              | 2              | 2                       | 60                  | 2       | 0       | 1         | 1        | 5        | 2                   | GRADE_TO_PASS_SCORE | 2                 | 2        | 2             | 2             | 0      | 2                | 5     | 3           | 2              | 4                   | 3                 | EXAM_LO_SUBMISSION_FAILED | EXAM_LO_SUBMISSION_FAILED |
            | 2           | 0                    | 2        | 1               | 0                | 0               | 0              | 2              | 2                       | 80                  | 2       | 0       | 1         | 1        | 5        | 2                   | GRADE_TO_PASS_SCORE | 2                 | 2        | 2             | 2             | 2      | 2                | 5     | 4           | 2              | 3                   | 4                 | EXAM_LO_SUBMISSION_FAILED | EXAM_LO_SUBMISSION_PASSED |

    Scenario Outline: Authentication for upsert grade book setting
        Given <exam_lo>a signed in "<role>"
        When <exam_lo>user upsert grade book setting
        Then <exam_lo>returns "<status>" status code

        Examples:
            | role           | status           |
            | school admin   | OK               |
            | student        | PermissionDenied |
            | parent         | PermissionDenied |
            | teacher        | PermissionDenied |
            | hq staff       | OK               |
            | center lead    | PermissionDenied |
            | center manager | PermissionDenied |
            | center staff   | PermissionDenied |

    Scenario Outline:
        Given <exam_lo>"4" student login
        And <exam_lo>a school admin login
        And admin respectively create 3 books with 0 exam lo, 1 exam lo and 2 exam los
        And admin create 1st course with 3 study plans using 3 books
        And admin create 2nd course with a study plan with book have 2 exam los
        And admin create 3rd course with a study plan with book have 0 exam lo
        And admin create 4th course with no study plan
        And admin create 1st student at grade 5 join all courses
        And admin create 2nd student at grade 6 join 1st course
        And admin create 3rd student at grade 7 join 3rd course
        And admin create 4th student at grade 5 join no course
        And admin get list grade book with "<course>" and "<grade>" and "<student>" and "<record_per_page>"
        Then <exam_lo>returns "OK" status code
        And returns correct "<total_item>"
        Examples:
        # total_item is number of response item (not total student)
            | record_per_page | course | grade | student | total_item |
            |                 | 1      |       |         | 2          |
            |                 | 1,3    |       |         | 2          |
            |                 |        | 5     |         | 3          |
            |                 |        | 5,6   |         | 3          |
            |                 |        |       | 3       | 0          |
            |                 |        |       | 3,1     | 2          |
            | 2               |        |       | 3,1     | 2          |
