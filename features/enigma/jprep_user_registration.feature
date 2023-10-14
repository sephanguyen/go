Feature: JPREP registers user

    Scenario Outline: Save logs data user sync
        Given a request with <n_student> student and <n_staff> staff
        And a valid JPREP signature in its header
        When the request user registration is performed
        And returns "OK" status code
        Then a signed in as "school admin" with school "<school>"
        And a partner "<school>" data sync log already exists in DB
        And a partner "<school>" data sync logs split already exists <n_logs_split> rows in DB
        And a payload of "<school>" data sync logs split match with request
        Examples:
        # 500 items per log
        | n_student   | n_staff | n_logs_split | school       |
        | 2           | 2       | 2            | -2147483647  |
        | 501         | 2       | 3            | -2147483647  |

    Scenario Outline: Don't save logs data user sync because payload invalid
        Given a request with <n_student> student and <n_staff> staff with payload invalid
        And a valid JPREP signature in its header
        When the request user registration is performed
        And returns "OK" status code
        Then a signed in as "school admin" with school "<school>"
        And a partner "<school>" data sync log not exists in DB
        Examples:
        | n_student   | n_staff | school       |
        | 2           | 2       | -2147483647  |
        | 501         | 2       | -2147483647  |