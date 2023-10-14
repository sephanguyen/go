Feature: Register Student Class
  As a admin
  I want to register class for a student course

  Scenario Outline: Register class for a student
    Given student exist in our system
    And assign student package with class empty to exist student
    When "<signed-in user>" register class for a student
    Then student package class must store in database
    And receives "OK" status code

    Examples:
      | signed-in user |
      | school admin   |