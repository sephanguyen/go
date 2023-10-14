Feature: JPREP master registration

    Scenario Outline: Save logs data master sync
        Given a request with <n_course> course, <n_lesson> lesson, <n_class> class and <n_academic_year> academic year
        And a valid JPREP signature in its header
        When the request master registration is performed
        And returns "OK" status code
        Then a signed in as "school admin" with school "<school>"
        And a partner "<school>" data sync log already exists in DB
        And a partner "<school>" data sync logs split already exists <n_logs_split> rows in DB
        And a payload of "<school>" data sync logs split match with request
        Examples:
        # 500 items per log
        | n_course    | n_lesson | n_class | n_academic_year | school       | n_logs_split |
        | 2           | 2        | 2       | 2               | -2147483647  | 4            |
        | 501         | 2        | 2       | 2               | -2147483647  | 5            |

    Scenario Outline: Don't save logs data master sync because payload invalid
        Given a request with <n_course> course, <n_lesson> lesson, <n_class> class and <n_academic_year> academic year with payload invalid
        And a valid JPREP signature in its header
        When the request master registration is performed
        And returns "OK" status code
        Then a signed in as "school admin" with school "<school>"
        And a partner "<school>" data sync log not exists in DB
        Examples:
        # 500 items per log
        | n_course    | n_lesson | n_class | n_academic_year | school       |
        | 2           | 2        | 2       | 2               | -2147483647  | 
        | 501         | 2        | 2       | 2               | -2147483647  | 