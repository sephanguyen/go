@quarantined
Feature: users get tags by filter
    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations

    Scenario: user see correct page offsets
        Given school admin creates "<total>" tag with "random" name
        When school admin search with filter of "<limit>" result at position "<offset>"
        Then response have correct "<prev_offset>" and "<next_offset>"
        Examples:
            | total | limit | offset | prev_offset | next_offset |
            | 100   | 10    | 0      | 0           | 10          |
            | 50    | 5     | 10     | 5           | 15          |
            | 3     | 5     | 2      | 0           | 3           |
            | 100   | 10    | 100    | 90          | 100         |
            | 100   | 10    | 110    | 100         | 110         |

    Scenario: user search for tags by keyword and see correct result
        Given school admin creates "<total>" tag with "random" name
        And school admin creates "<num>" tag with "<keyword>" name
        And school admin delete "<num_delete>" of those tag
        When school admin search with filter of "<limit>" result at position "<offset>"
        Then school admin see "<num_result>" in total of "<total_result>"
        Examples:
            | total | num | keyword      | num_delete | limit | offset | num_result | total_result |
            | 100   | 18  | manabie-tag  | 5          | 10    | 0      | 10         | 13           |
            | 50    | 10  | notification | 0          | 5     | 9      | 1          | 10           |
            | 3     | 5   | tag2         | 3          | 5     | 0      | 2          | 2            |

    Scenario: user search for tags have been archived and will not see archived tags
        Given school admin creates "5" tag with "random" name
        And school admin creates "3" tag with "archived" name
        And school admin archived those tag
        When school admin search with filter of "100" result at position "0"
        Then school admin see "0" in total of "0"