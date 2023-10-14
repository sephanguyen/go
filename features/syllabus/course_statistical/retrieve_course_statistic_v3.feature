Feature: Retrieve course statistic v3

    Background: user sign in
        Given <course_statistical>a signed in "school admin"

    Scenario Outline: authenticate create book
        Given <course_statistical>a signed in "<role>"
        When user create a book
        Then <course_statistical>returns "<msg>" status code
        Examples:
            | role         | msg              |
            | student      | PermissionDenied |
            | school admin | OK               |

    Scenario Outline: retrieve course statistic with no class filter v3
        Given "<num_student>" student login
        And <course_statistical>a school admin login
        And students exists in school history in DB
        And tag users valid exists in DB
        And <course_statistical>a teacher login
        And "school admin" has created a book with each "<num_los>" los, "<num_ass>" assignments, "<num_topic>" topics, "<num_chap>" chapters, "<num_quiz>" quizzes
        And "school admin" has created a course with a book
        And "school admin" has created a studyplan for all student
        And "school admin" has updated course duration for student
        And "<num_student_do_test>" students do test and each student done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics V2
        When retrieve course statistic v3 with no class, "<school>", "<tag>" filter
        And <course_statistical_v3>topic total assigned student is <topic_assigned>, completed students is <topic_completed_student>, average score is <topic_average_score>

        Examples:
            | num_student | num_student_in_class | done_los | correct_quizzes | done_assignments | assignment_mark | skipped_topics | topic_assigned | topic_completed_student | topic_average_score | num_los | num_ass | num_topic | num_chap | num_quiz | num_student_do_test | school     | tag    |
            # Testcase 1:
            # 2 student finish their assigne
            # 2 los => 4 studyplanItem + 2 master studyplanItem
            # corrrect_quizzes = 5/5 quizzes total and two each assigment has 7 point
            # => avg scorre each master studyplanItem (100+100+70+70)/4 = 85
            # => ave score of topic (85+85)/2 = 85
            # => Final result is 85, completed  student 2
            | 2           | 0                    | 2        | 5               | 2                | 7               | 0              | 2              | 2                       | 85                  | 2       | 2       | 1         | 1        | 5        | 2                   | all        | all    |
            # Testcase 2:
            # 2 student finish their assigne
            # 2 los => 4 studyplanItem + 2 master studyplanItem
            # corrrect_quizzes = 2/5 quizzes total and two each assigment has 7 point
            # => avg scorre each master studyplanItem (40+40+70+70)/4 = 55
            # => ave score of topic (55+55)/2 = 55
            # => Final result is 85, completed  student 2
            | 2           | 0                    | 1        | 2               | 1                | 7               | 0              | 2              | 2                       | 55                  | 1       | 1       | 1         | 1        | 5        | 2                   | all        | all    |
            # Testcase 3:
            # 1 student finish their assigne lo
            # 2 los => 4 studyplanItem + 2 master studyplanItem
            # corrrect_quizzes = 5/5 quizzes total and two each assigment has 7 point
            # => avg scorre each master studyplanItem (60+70)/2 = 65
            # => ave score of topic 65/1 = 65
            # => Final result is 65, completed  student 1
            | 2           | 0                    | 2        | 3               | 2                | 7               | 0              | 2              | 1                       | 65                  | 2       | 2       | 1         | 1        | 5        | 1                   | all        | all    |
            | 2           | 0                    | 2        | 3               | 2                | 7               | 0              | 1              | 1                       | 65                  | 2       | 2       | 1         | 1        | 5        | 1                   | school_id  | tag_id |
            | 2           | 0                    | 2        | 3               | 2                | 7               | 0              | 0              | 0                       | 0                   | 2       | 2       | 1         | 1        | 5        | 1                   | unassigned | all    |


    Scenario Outline: retrieve course statistic with class filter
        Given "<num_student>" student login
        And <course_statistical> "<num_student_in_class>" in a "<type>" class
        And <course_statistical>a school admin login
        And <course_statistical>a teacher login
        And "school admin" has created a book with each "<num_los>" los, "<num_ass>" assignments, "<num_topic>" topics, "<num_chap>" chapters, "<num_quiz>" quizzes
        And "school admin" has created a course with a book
        And course "<have>" a class
        And "school admin" has created a studyplan for all student
        And "school admin" has updated course duration for student
        And "<num_student_do_test>" students do test and each student done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
        When retrieve course statistic v3 with class filter
        Then our system returns correct topic statistic
        And topic total assigned student is <topic_assigned>, completed students is <topic_completed_student>, average score is <topic_average_score>

        Examples:
            | num_student | num_student_in_class | done_los | correct_quizzes | done_assignments | assignment_mark | skipped_topics | topic_assigned | topic_completed_student | topic_average_score | num_los | num_ass | num_topic | num_chap | num_quiz | num_student_do_test | type      | have  |
            # Testcase 1:
            # 2 student finish their assigne in different class
            # 2 los => 4 studyplanItem + 2 master studyplanItem
            # corrrect_quizzes = 5/5 quizzes total and two each assigment has 7 point
            # => avg scorre each master studyplanItem (100+70)/2 = 85
            # => ave score of topic (85)/1 = 85
            # => Final result is 85, completed  student 1
            | 2           | 1                    | 2        | 5               | 2                | 7               | 0              | 1              | 1                       | 85                  | 2       | 2       | 1         | 1        | 5        | 2                   | different | true  |
            # Testcase 2:
            # 2 student finish their assigne in the same class
            # 2 los => 4 studyplanItem + 2 master studyplanItem
            # corrrect_quizzes = 5/5 quizzes total and two each assigment has 7 point
            # => avg scorre each master studyplanItem (100+100+70+70)/2 = 85
            # => ave score of topic (85+85)/2 = 85
            # => Final result is 85, completed  student 1
            | 2           | 2                    | 2        | 5               | 2                | 7               | 0              | 2              | 2                       | 85                  | 2       | 2       | 1         | 1        | 5        | 2                   | same      | true  |
            # Testcase 3:
            # 2 student finish their assigne in the same class
            # Course does not have a class
            # => Final result nothing
            | 2           | 2                    | 0        | 0               | 0                | 0               | 0              | 0              | 0                       | 0                   | 2       | 2       | 1         | 1        | 5        | 2                   | same      | false |

