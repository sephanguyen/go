Feature: download mapping file

    Scenario Outline: authenticate download mapping file
        Given <learning_history_data_sync>a signed in "<role>"
        When user download mapping file
        Then <learning_history_data_sync>returns "<msg>" status code

        Examples:
            | role         | msg |
            | school admin | OK  |
            | parent       | OK  |
            | student      | OK  |
            | teacher      | OK  |
            | hq staff     | OK  |
    # | centre lead    | OK   |
    # | centre manager | OK   |
    # | teacher lead   | OK   |

    Scenario: download mapping file
        Given <learning_history_data_sync>a signed in "school admin"
        When user download mapping file
        Then <learning_history_data_sync>returns "OK" status code
        And returns url of mapping file correctly