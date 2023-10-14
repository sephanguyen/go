Feature: JPREP user course registration

    Scenario Outline: Save logs data user course sync
        Given a request with <n_student_lessons> student lessons 
        And a valid JPREP signature in its header
        When the request user course registration is performed
        And returns "OK" status code
        Then a signed in as "school admin" with school "<school>"
        And a partner "<school>" data sync log already exists in DB
        And a partner "<school>" data sync logs split already exists <n_logs_split> rows in DB
        And a payload of "<school>" data sync logs split match with request
        Examples:
        # 500 items per log
        | n_student_lessons | school       | n_logs_split |
        | 2                 | -2147483647  | 1            |
        | 501               | -2147483647  | 2            |

    Scenario Outline: Don't save logs data student lessons sync because payload invalid
        Given a request with <n_student_lessons> student lessons with payload invalid
        And a valid JPREP signature in its header
        When the request user course registration is performed
        And returns "OK" status code
        Then a signed in as "school admin" with school "<school>"
        And a partner "<school>" data sync log not exists in DB
        Examples:
        # 500 items per log
        | n_student_lessons | school       |
        | 2                 | -2147483647  | 
        | 501               | -2147483647  |