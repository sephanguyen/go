Feature: Get logs report partner data sync

    @critical
    Scenario Outline:  Get logs report partner data sync
        Given a request with <n_student> student and <n_staff> staff
        And a valid JPREP signature in its header
        When the request user registration is performed
        And returns "OK" status code
        Then a signed in as "school admin" with school "<school>"
        And a partner "<school>" data sync log already exists in DB
        And a partner "<school>" data sync logs split already exists <n_logs_split> rows in DB
        And a payload of "<school>" data sync logs split match with request
        And a request get partner data logs report
        And a valid JPREP signature in its header
        When the request get partner log report is performed
        And returns "OK" status code
        And a response of "<school>" partner log report match with DB
        Examples:
        | n_student   | n_staff | n_logs_split | school       | 
        | 2           | 2       | 2            | -2147483647  |
        | 2           | 0       | 1            | -2147483647  |