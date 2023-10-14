Feature: Create scheduler

    Scenario Outline: user create scheduler
        Given signed as "school admin" account
        When user creates a scheduler "<start_date>", "<end_date>", "<frequency>"
        Then returns "OK" status code
        And scheduler has been added to the database

        Examples:
            |  start_date                  |  end_date                    | frequency  |
            |  2022-07-31T00:00:00+07:00   |  2022-08-31T23:59:00+07:00   | WEEKLY     |
            |  2022-07-31T00:00:00+07:00   |  2022-07-31T23:59:00+07:00   | ONCE       |