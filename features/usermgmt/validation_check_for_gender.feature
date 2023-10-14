Feature: validation check for gender when create user

  Scenario Outline: Create a student with valid gender student info
    Given only student info with "<gender>"
    When "<signed-in user>" create new student account
    Then new student account created success with student info
    And receives "OK" status code

    Examples:
      | signed-in user | gender |
      | school admin   | MALE   |
      | school admin   | FEMALE |

  Scenario Outline: Cannot create student account with invalid gender
    Given only student info with "<gender>"
    When "<signed-in user>" create new student account
    Then "<signed-in user>" cannot create that account

    Examples:
      | signed-in user | gender
      | school admin   | INVALID_GENDER
