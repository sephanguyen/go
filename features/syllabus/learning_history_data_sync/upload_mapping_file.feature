Feature: upload mapping file

    Background:
        Given <learning_history_data_sync>a signed in "school admin"
        And <learning_history_data_sync>valid course, exam_lo, question_tag in db

    Scenario Outline: authenticate upload mapping file
        Given <learning_history_data_sync>a signed in "<role>"
        When user upload mapping file
        Then <learning_history_data_sync>returns "<msg>" status code

        Examples:
            | role           | msg |
            | school admin   | OK  |
            | parent         | OK  |
            | student        | OK  |
            | teacher        | OK  |
            | hq staff       | OK  |
            | centre lead    | OK  |
            | centre manager | OK  |
            | teacher lead   | OK  |

    Scenario Outline:
        Given <learning_history_data_sync>a signed in "school admin"
        When user upload mapping file
        Then csv file is uploaded