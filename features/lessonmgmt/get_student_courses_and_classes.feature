Feature: User get student's courses and classes

  Background:
    Given enter a school
    When have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some classes assign to courses
    And have some student subscriptions

  Scenario: Get student's courses and classes
    Given insert student class member
    And user signed in as teacher
    When get student's courses and classes
    Then returns "OK" status code
    And must get correct courses and classes of students

  Scenario: Get empty student's courses and classes
    And user signed in as teacher
    When get student's courses and classes
    Then returns "OK" status code
    And must get correct courses and classes of students
