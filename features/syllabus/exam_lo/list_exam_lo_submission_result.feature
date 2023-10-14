Feature: List exam lo submission result

    # Scenario Outline: authenticate when list exam lo submission result
    #     Given <exam_lo>a signed in "<role>"
    #     And there are exam lo submissions existed
    #     When user list exam lo submission result
    #     Then <exam_lo>returns "<status code>" status code

    #     Examples:
    #         | role           | status code      |
    #         | school admin   | PermissionDenied |
    #         | admin          | PermissionDenied |
    #         | teacher        | OK               |
    #         | student        | OK               |
    #         | hq staff       | PermissionDenied |
    #         | center lead    | PermissionDenied |
    #         | center manager | PermissionDenied |
    #         | center staff   | PermissionDenied |

    Scenario Outline: authenticate when list exam lo submission result
        Given <exam_lo>a signed in "teacher"
        And there are exam lo submission scores existed
        When user list exam lo submission result
        Then <exam_lo>returns "OK" status code
        And our system must returns list exam lo submission result correctly
