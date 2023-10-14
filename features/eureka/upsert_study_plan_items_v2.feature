Feature: Upsert study plan item feature v2

  Background: Background has study plan
    Given a study plan name "test-study-plan" in db

  Scenario Outline: Authentication for upsert study plan item v2
    Given a signed in "<role>"
    When user upsert a list of study plan item v2
    Then returns "<status>" status code

    Examples: 
      | role           | status           |
      | school admin   | OK               |
      | student        | PermissionDenied |
      | parent         | PermissionDenied |
      | teacher        | OK               |
      | hq staff       | OK               |
      # | center lead    | OK               |
      # | center manager | OK               |
      # | center staff   | OK               |

#      Scenario: Upsert study plan item v2
#           Given a valid "teacher" token
#           When user upsert a list of study plan item v2
#           Then returns "OK" status code
#           Then eureka must store correct study plan item v2