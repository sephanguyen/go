Feature: Recover data sync

    Scenario Outline:  Recover data sync
        Given a random number
        And a data sync log already of "<kind>" with "<current-status>" and <current-retry-times> exists in DB at "<date>"
        And request with recover data sync at "<date>"
        And a valid JPREP signature in its header
        When the request recover data sync is performed
        And returns "OK" status code
        Then a signed in as "school admin" with school "<school>"
        And a partner "<school>" data sync log split match with <expect-retry-times>

        Examples:
        | kind            | current-status | current-retry-times | expect-retry-times | school       | date       |
        | STUDENT         | PROCESSING     | 0                   | 1                  | -2147483647  | 2022-04-01 |
        | STAFF           | PROCESSING     | 1                   | 2                  | -2147483647  | 2022-04-02 |
        | COURSE          | PROCESSING     | 2                   | 3                  | -2147483647  | 2022-04-03 |
        | CLASS           | PROCESSING     | 2                   | 3                  | -2147483647  | 2022-04-04 |
        | LESSON          | PROCESSING     | 0                   | 1                  | -2147483647  | 2022-04-05 |
        | ACADEMICYEAR    | PROCESSING     | 1                   | 2                  | -2147483647  | 2022-04-06 |
        | STUDENT_LESSONS | PROCESSING     | 2                   | 3                  | -2147483647  | 2022-04-07 |
        | STUDENT         | PROCESSING     | 3                   | 3                  | -2147483647  | 2022-04-08 |
        | STUDENT         | PENDING        | 0                   | 1                  | -2147483647  | 2022-04-09 |
        | STUDENT         | SUCCESS        | 1                   | 2                  | -2147483647  | 2022-04-10 |
        | STUDENT         | FAILED         | 3                   | 3                  | -2147483647  | 2022-04-11 |
