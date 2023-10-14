Feature: Check quiz correctness
  I'm doing the quiz test, i need to submit quiz answers

  Background: Valid study plan item and exam lo
    Given a signed in "school admin"
    And user create a study plan of exam lo to database

  Scenario: student choose option of the quiz (do all quizzes)
    Given a quiz test include "5" multiple choice quizzes with "5" quizzes per page and do quiz test
    When student choose option "1, 2, 3" of the quiz "1" for submit quiz answers
    And student choose option "1, 2" of the quiz "2" for submit quiz answers
    And student choose option "1, 3" of the quiz "3" for submit quiz answers
    And student choose option "2, 3" of the quiz "4" for submit quiz answers
    And student choose option "1, 4" of the quiz "5" for submit quiz answers
    And student submit quiz answers
    Then returns "OK" status code
    And returns expected result multiple choice type for submit quiz answers

  Scenario: student choose option of the quiz (skip some quizzes)
    Given a quiz test include "5" multiple choice quizzes with "5" quizzes per page and do quiz test
    When student choose option "1, 2, 3" of the quiz "1" for submit quiz answers
    And student choose option "1, 2" of the quiz "2" for submit quiz answers
    And student choose option "1, 4" of the quiz "5" for submit quiz answers
    And student submit quiz answers
    Then returns "OK" status code
    And returns expected result multiple choice type for submit quiz answers

  Scenario: student fill text of the quiz many times (do all quizzes)
    Given a quiz test "5" fill in the blank quizzes with "5" quizzes per page and do quiz test
    When student fill text "hello, goodbye, meeting, fine, bye" of the quiz "1"
    When student fill text "hello, goodbye, fine, bye" of the quiz "2"
    When student fill text "hello, meeting, fine, bye" of the quiz "3"
    When student fill text "goodbye, meeting, fine, bye" of the quiz "4"
    When student fill text "hello, bye" of the quiz "5"
    And student submit quiz answers
    Then returns "OK" status code
    And returns expected result fill in the blank type for submit quiz answers

  Scenario: student fill text of the quiz many times (skip all quizzes)
    Given a quiz test "5" fill in the blank quizzes with "5" quizzes per page and do quiz test
    When student fill text "hello, goodbye, meeting, fine, bye" of the quiz "1"
    When student fill text "hello, goodbye, fine, bye" of the quiz "2"
    When student fill text "hello, meeting, fine, bye" of the quiz "3"
    When student fill text "goodbye, meeting, fine, bye" of the quiz "4"
    When student fill text "hello, bye" of the quiz "5"
    And student submit quiz answers
    Then returns "OK" status code
    And returns expected result fill in the blank type for submit quiz answers

  Scenario: student do pair of word quizzes
    Given a quiz test include "5" pair of word quizzes with "5" quizzes per page and do quiz test
    When student answer pair of word quizzes for submit quiz answers
    And student submit quiz answers
    Then returns "OK" status code
    And returns expected result pair of word quizzes for submit quiz answers

  Scenario: student do term and definition quizzes
    Given a quiz test include "5" term and definition quizzes with "5" quizzes per page and do quiz test
    When student answer term and definition quizzes for submit quiz answers
    And student submit quiz answers
    Then returns "OK" status code
    And returns expected result term and definition quizzes for submit quiz answers

  Scenario: student fill text of the quiz many times (skip all quizzes)
    Given a quiz test "5" fill in the blank quizzes with "5" quizzes per page and do quiz test
    And student submit quiz answers
    Then returns "OK" status code
    And returns expected result fill in the blank type for submit quiz answers

  Scenario: student do all ordering quizzes
    Given a quiz test "5" "ordering" quizzes with "5" quizzes per page and do quiz test
    When student answer correct order options for all quizzes
    And student submit quiz answers
    Then returns "OK" status code
    And returns result all correct in submit quiz answers for ordering question
  
    Scenario: student do all essay quizzes
    Given a quiz test "5" "essay" quizzes with "5" quizzes per page and do quiz test
    When student finish essay questions
    And student submit quiz answers
    Then returns "OK" status code
    And returns essay quiz answers
