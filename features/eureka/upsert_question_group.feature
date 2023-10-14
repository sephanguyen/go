Feature: Upsert question group

  Background: given a quizset of a learning objective
    Given a quiz set with "2" quizzes in Learning Objective belonged to a "TOPIC_TYPE_EXAM" topic

  Scenario Outline: Insert question group with rich description
    When insert question group with "<type>" rich description
    Then returns "OK" status code

    Examples:
      | type  |
      | empty |
      | null  |
      | full  |


