Feature: Retrieve school history by student in course

    Scenario Outline: authenticate Retrieve school history by student in course
        Given <course_statistical>a signed in "<role>"
        When retrieve school history by student in course
        Then <course_statistical>returns "<msg>" status code
        Examples:
            | role    | msg              |
            | student | PermissionDenied |

    Scenario Outline: retrieve school history by student in course
        Given "<num_student>" student login
        And <course_statistical>a signed in "school admin"
        And students exists in school history in DB
        And <course_statistical>a school admin login
        And <course_statistical>a teacher login
        And "school admin" has created a book with each "<num_los>" los, "<num_ass>" assignments, "<num_topic>" topics, "<num_chap>" chapters, "<num_quiz>" quizzes
        And "school admin" has created a course with a book
        And "school admin" has created a studyplan for all student
        And "school admin" has updated course duration for student
        When <course_statistical>a teacher login
        And retrieve school history by student in course
        Then <course_statistical>returns "OK" status code
        And there are 2 school information
        Examples:
            | num_student | num_los | num_ass | num_topic | num_chap | num_quiz |
            | 2           | 2       | 2       | 1         | 1        | 5        |
