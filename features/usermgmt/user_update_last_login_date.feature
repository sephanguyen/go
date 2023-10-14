Feature: Update User Last Login Date

    Scenario Outline: User update last login date
    Given a "<kind of user>" signed in user with "<role>"
    When user update last login date with "<value condition>" value
    Then user last login date "<update result>"
    And returns "<status code>" status code

    Examples:
      | role         | kind of user         | value condition | update result  | status code     |
      | school admin | newly created        | valid           | is updated     | OK              |
      | teacher      | newly created        | valid           | is updated     | OK              |
      | student      | newly created        | valid           | is updated     | OK              |
      | parent       | newly created        | valid           | is updated     | OK              |
      | school admin | newly created        | missing         | is not updated | InvalidArgument |
      | teacher      | newly created        | missing         | is not updated | InvalidArgument |
      | student      | newly created        | missing         | is not updated | InvalidArgument |
      | parent       | newly created        | missing         | is not updated | InvalidArgument |
      | school admin | newly created        | zero            | is not updated | InvalidArgument |
      | teacher      | newly created        | zero            | is not updated | InvalidArgument |
      | student      | newly created        | zero            | is not updated | InvalidArgument |
      | parent       | newly created        | zero            | is not updated | InvalidArgument |
      | school admin | has signed in before | valid           | is updated     | OK              |
      | teacher      | has signed in before | valid           | is updated     | OK              |
      | student      | has signed in before | valid           | is updated     | OK              |
      | parent       | has signed in before | valid           | is updated     | OK              |
      | school admin | has signed in before | missing         | is not updated | InvalidArgument |
      | teacher      | has signed in before | missing         | is not updated | InvalidArgument |
      | student      | has signed in before | missing         | is not updated | InvalidArgument |
      | parent       | has signed in before | missing         | is not updated | InvalidArgument |
      | school admin | has signed in before | zero            | is not updated | InvalidArgument |
      | teacher      | has signed in before | zero            | is not updated | InvalidArgument |
      | student      | has signed in before | zero            | is not updated | InvalidArgument |
      | parent       | has signed in before | zero            | is not updated | InvalidArgument |
