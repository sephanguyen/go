# quarantined is for create multiple order with the same course, duplicate time range
@quarantined
Feature: Get Orders Items List

  Scenario: Get order items list Success
    Given prepare data for get list order items create "<type of order>" "<type of product>"
    When "<signed-in user for creation>" create "<type of order>" orders data for get list order items
    Then "<signed-in user for getting>" get list order items with "<filter>"

  Examples:
    | signed-in user for creation | signed-in user for getting | type of product                   | type of order    | filter                     |
    | school admin                | school admin               | material type one time            | new              | without filter             |
    | hq staff                    | hq staff                   | material type one time            | new              | without filter             |
    | centre manager              | centre manager             | material type one time            | new              | without filter             |
    | centre staff                | centre staff               | material type one time            | new              | without filter             |
    | centre staff                | centre lead                | material type one time            | new              | without filter             |
    # | school admin and teacher    | school admin and teacher   | material type one time            | new              | without filter             |
    | school admin                | school admin               | package type one time             | new              | without filter             |
    | hq staff                    | hq staff                   | package type one time             | new              | without filter             |
    | centre manager              | centre manager             | package type one time             | new              | without filter             |
    | centre staff                | centre staff               | package type one time             | new              | without filter             |
    | centre staff                | centre lead                | package type one time             | new              | without filter             |
    # | school admin and teacher    | school admin and teacher   | package type one time             | new              | without filter             |
    | school admin                | school admin               | fee type one time                 | new              | without filter             |
    | hq staff                    | hq staff                   | fee type one time                 | new              | without filter             |
    | centre manager              | centre manager             | fee type one time                 | new              | without filter             |
    | centre staff                | centre staff               | fee type one time                 | new              | without filter             |
    | centre staff                | centre lead                | fee type one time                 | new              | without filter             |
    # | school admin and teacher    | school admin and teacher   | fee type one time                 | new              | without filter             |
    | school admin                | school admin               | material type one time            | new              | empty filter               |
    | hq staff                    | hq staff                   | package type one time             | new              | empty filter               |
    | centre manager              | centre manager             | fee type one time                 | new              | empty filter               |
    # | school admin and teacher    | school admin and teacher   | package type one time             | new              | empty filter               |
    | school admin                | school admin               | material type one time            | new              | filter with empty response |
    | hq staff                    | hq staff                   | package type one time             | new              | filter with empty response |
    | centre manager              | centre manager             | fee type one time                 | new              | filter with empty response |
    # | school admin and teacher    | school admin and teacher   | package type one time             | new              | filter with empty response |
