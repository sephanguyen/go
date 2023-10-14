Feature: Check quiz correctness
  I'm doing the quiz test, i need to answer the quiz and check the correctness of the quiz

  Scenario: student choose option of the quiz
    Given a quiz test include "9" multiple choice quizzes with "3" quizzes per page and do quiz test

    When student choose option "1, 3" of the quiz "1"
    Then returns "OK" status code
    And returns expected result multiple choice type

  Scenario: student choose option of the quiz many time
    Given a quiz test include "9" multiple choice quizzes with "3" quizzes per page and do quiz test

    When student choose option "1, 2, 3, 4, 5" of the quiz "1"
    Then returns "OK" status code
    And returns expected result multiple choice type

    When student choose option "1, 3, 4" of the quiz "2"
    Then returns "OK" status code
    And returns expected result multiple choice type

    When student choose option "1, 4" of the quiz "3"
    Then returns "OK" status code
    And returns expected result multiple choice type

    When student choose option "1, 4" of the quiz "3"
    Then returns "OK" status code
    And returns expected result multiple choice type

  Scenario: student fill text of the quiz
    Given a quiz test "9" fill in the blank quizzes with "3" quizzes per page and do quiz test

    When student fill text "hello, goodbye" of the quiz "1"
    Then returns "OK" status code
    And returns expected result fill in the blank type

  Scenario: student fill text of the quiz many times
    Given a quiz test "9" fill in the blank quizzes with "3" quizzes per page and do quiz test

    When student fill text "hello, goodbye, meeting, fine, bye" of the quiz "1"
    Then returns "OK" status code
    And returns expected result fill in the blank type

    When student fill text "hello, goodbye, meeting, fine, bye" of the quiz "2"
    Then returns "OK" status code
    And returns expected result fill in the blank type

    When student fill text "hello, goodbye," of the quiz "3"
    Then returns "OK" status code
    And returns expected result fill in the blank type

  Scenario: student choose option with index out of range
    Given a quiz test include "9" multiple choice quizzes with "3" quizzes per page and do quiz test

    When student choose option "1, 100" of the quiz "1"
    Then returns "FailedPrecondition" status code

  Scenario: student does not choose any option
    Given a quiz test include "9" multiple choice quizzes with "3" quizzes per page and do quiz test

    When student choose option "0" of the quiz "1"
    Then returns "FailedPrecondition" status code

  Scenario: student does not choose a quiz
    Given a quiz test include "9" multiple choice quizzes with "3" quizzes per page and do quiz test

    When student missing quiz id in request
    Then returns "InvalidArgument" status code

  Scenario: student do pair of word quizzes
    Given a quiz test include "9" pair of word quizzes with "3" quizzes per page and do quiz test
    When student answer pair of word quizzes
    Then returns "OK" status code
    And returns expected result pair of word quizzes

  Scenario: student do term and definition quizzes
    Given a quiz test include "9" term and definition quizzes with "3" quizzes per page and do quiz test
    When student answer term and definition quizzes
    Then returns "OK" status code
    And returns expected result term and definition quizzes

  Scenario: student do fill in the blank quizzes with ocr
    Given a quiz test "9" fill in the blank quizzes with "3" quizzes per page and do quiz test
    When student answer fill in the blank quiz with ocr
    Then returns "OK" status code
    And returns expected result fill in the blank quiz

  @blocker
  Scenario Outline: student do quizzes
    Given a quiz test include "<total_quizzes>" with "<quiz_type>" quizzes with "<limit>" quizzes per page and do quiz test
    When student answer "<quiz_type>" quizzes
    Then returns "OK" status code
    And returns expected result "<quiz_type>" quizzes

    Examples:
      | total_quizzes | quiz_type             | limit |
      | 9             | term and definition   | 3     |
      | 15            | fill in the blank new | 7     |
      | 15            | fill in the blank old | 11    |
      | 13            | term and definition   | 4     |
      | 19            | multiple choice       | 11    |
      | 29            | pair of word          | 11    |
      | 41            | manual input          | 13    |
      | 71            | manual input          | 22    |
      | 9             | ordering              | 5     |
