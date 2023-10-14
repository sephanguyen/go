Feature: Retrieve quizzes of learning objective

  Background: given a quiz set of an learning objective
    Given a quizset with "9" quizzes in Learning Objective belonged to a "TOPIC_TYPE_EXAM" topic

  Scenario Outline: Teacher list quizzes of learning objectives
    When teacher get quizzes of "<status>" lo
    Then returns expected totalQuestion when retrieveLO
    Then returns expected "<status>" list of quizzes
    And each item have returned addition fields for flashcard

    Examples:
      | status      |
      | existed     |
      | not existed |

  Scenario Outline: Teacher list quizzes of learning objectives
    When teacher get quizzes of "<status>" lo
    Then returns expected totalQuestion when retrieveLOV1
    Then returns expected "<status>" list of quizzes
    And each item have returned addition fields for flashcard

    Examples:
      | status      |
      | existed     |
      | not existed |

  Scenario Outline: Teacher list quizzes of learning objectives having question groups
    Given existing question group
    And user upsert a valid "questionGroup" single quiz
    When teacher get quizzes of "<status>" lo
    Then returns expected totalQuestion when retrieveLO
    And returns expected "<status>" list of quizzes
    And each item have returned addition fields for flashcard
    And question group existed in response

    Examples:
      | status      |
      | existed     |
      | not existed |