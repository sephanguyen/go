Feature: Update a exam lo

  Background:a valid book content
    Given <exam_lo>a signed in "school admin"
    And <exam_lo>a valid book content
    And there are exam LOs existed in topic

  Scenario Outline: Authentication <role> for update exam LO
    Given <exam_lo>a signed in "<role>"
    When user update a valid exam LO
    Then <exam_lo>returns "<status>" status code

    Examples:
      | role           | status           |
      | school admin   | OK               |
      | student        | PermissionDenied |
      | parent         | PermissionDenied |
      | teacher        | PermissionDenied |
      | hq staff       | OK               |
      | center lead    | PermissionDenied |
      | center manager | OK               |
      | center staff   | PermissionDenied |

  Scenario: admin update an exam LO
    When user update a valid exam LO
    Then our system update exam LO correctly

  Scenario Outline: admin update exam LO
    Given there are exam LOs existed in topic
    When user update a exam LO with "<field>"
    Then <exam_lo>returns "OK" status code
    And our system must update exam LO with "<field>" correctly

    Examples:
      | field           |
      | instruction     |
      | grade_to_pass   |
      | manual_grading  |
      | time_limit      |
      | maximum_attempt |
      | approve_grading |
      | grade_capping   |
      | review_option   |
