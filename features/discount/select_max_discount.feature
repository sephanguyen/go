Feature: Auto select max discount based on user_discount_tag
  @blocker
  Scenario: Auto select max discount with update order
    Given prepare data for max discount selection with "<discount tag>"
    When "school admin" added "<discount tag>" to student
    Then system selects max discount for student
    And event is received for update product

    Examples:
      | discount tag        |
      | single parent       |
      | employee full time  |
      | employee part time  |
      | family              |
