Feature: Trigger fill exam lo table

  Scenario Outline: Trigger insert fill exam lo table
    When insert a valid learning objectives with "<type>"
    Then there is a "<row>" in the exam lo table
    Examples:
      | type                            | row     |
      | LEARNING_OBJECTIVE_TYPE_EXAM_LO | valid   |
      | LEARNING_OBJECTIVE_TYPE_NONE    | invalid |

  Scenario Outline: Trigger update fill exam lo table
    Given insert a valid learning objectives with "<type>"
    When update a valid record learning objectives
    Then there is a "<row>" in the exam lo table
    Examples:
      | type                            | row     |
      | LEARNING_OBJECTIVE_TYPE_EXAM_LO | valid   |
      | LEARNING_OBJECTIVE_TYPE_NONE    | invalid |
