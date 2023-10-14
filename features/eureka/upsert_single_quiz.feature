Feature: Upsert quiz

  Background:
    Given a signed in "school admin"
    And a question tag type existed in database
    And a question tag existed in database

  Scenario Outline: Validate upsert quiz
    When user upsert "<validity>" single quiz
    Then returns "<status>" status code

    Examples:
      | validity            | status          |
      | valid               | OK              |
      | missing ExternalId  | InvalidArgument |
      | missing Question    | InvalidArgument |
      | missing Explanation | InvalidArgument |
      | missing Options     | InvalidArgument |
      | missing Attribute   | OK              |

  Scenario Outline: Upsert valid quiz
    When user upsert a valid "<quiz>" single quiz
    Then returns "OK" status code
    And quiz created successfully with new version
    And quiz_set also updated with new version

    Examples:
      | quiz               |
      | valid              |
      | existed            |
      | emptyTagLO         |
      | multipleChoiceQuiz |
      | fillInTheBlank     |
      | pairOfWordQuiz     |
      | termAndDefinition  |
      | manualInputQuiz    |
      | multiAnswerQuiz    |
      | orderingQuiz       |
      | essayQuiz          |
      | emptyPoint         |
      | withLabelType      |

  Scenario Outline: Upsert valid quiz with question group
    Given user upsert a valid "<quiz>" single quiz
    And returns "OK" status code
    And existing question group
    When user upsert a valid "questionGroup" single quiz
    Then quiz created successfully with new version
    And quiz_set also updated with new version

    Examples:
      | quiz               |
      | valid              |
      | existed            |
      | emptyTagLO         |
      | multipleChoiceQuiz |
      | fillInTheBlank     |
      | pairOfWordQuiz     |
      | termAndDefinition  |
      | manualInputQuiz    |
      | multiAnswerQuiz    |
      | orderingQuiz       |
      | essayQuiz          |
      | emptyPoint         |
      | withLabelType      |
