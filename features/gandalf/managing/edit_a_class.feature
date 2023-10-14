Feature: Edit a class

  Background:
    Given a signed in admin
    And a random number
    # a random number will be converted to string and will be appended to school_name
    # we need to do that because we have a unique constraint (school_name, country, city_id, district_id) in DB
    And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
    And a school name "S2", country "COUNTRY_VN", city "Hồ Chí Minh", district "3"
    And admin inserts schools
    And admin create a class with school name "S1" and expired at "2150-12-12"

  Scenario: owner class try to edit a class
    Given a teacher who is owner current class
    And a EditClassRequest with class name is "edit-class-name"
    And a "valid" classId in EditClassRequest
    When user edit a class
    Then Bob returns "OK" status code
    And Bob must update class in db
    And Bob must push msg "EditClass" subject "Class.Upserted" to nats

  Scenario Outline: user try to edit class with valid data
    Given a signed in "<role>" with school name "<school name>"
    And a EditClassRequest with class name is "<class name>"
    And a "<class id>" classId in EditClassRequest
    When user edit a class
    Then Bob returns "OK" status code
    And Bob must update class in db
    And Bob must push msg "EditClass" subject "Class.Upserted" to nats

    Examples:
      | role         | school name | class name | class id |
      | admin        | S1          | edit name  | valid    |
      | school admin | S1          | edit name  | valid    |
