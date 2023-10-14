Feature: Edit assignment time
  Background:
    Given <study_plan_item> a signed in "school admin"
    And <study_plan_item> "school admin" has created a content book
    And <study_plan_item> "school admin" has created a study plan exact match with the book content for 2 student

  Scenario Outline: Authentication for update assignment time
    Given <study_plan_item> a signed in "<role>"
    When <study_plan_item> user send update assignment time request
    Then <study_plan_item> returns "<status>" status code

    Examples:
      | role           | status |
      | school admin   | OK     |
      | student        | OK     |
      | parent         | OK     |
      | teacher        | OK     |
      | hq staff       | OK     |
      | center lead    | OK     |
      | center manager | OK     |
      | center staff   | OK     |

  Scenario Outline: Admin update assignment time with valid data
    Given admin update assignment time with "<update_type>"
    Then assignment time was updated with according update_type "<update_type>"

    Examples:
      | update_type |
      | start       |
      | end         |
      | start_end   |
