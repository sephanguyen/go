Feature: Edit assignment time
  Background:
    Given "school admin" logins "CMS"
    And "student" logins "Learner App"
    And "school admin" has created a content book
    And "school admin" has created a studyplan exact match with the book content for student

  Scenario Outline: Admin update assignment time with valid data
    Given admin update assignment time with "<update_type>"
    Then assignment time was updated with according update_type "<update_type>"

    Examples:
      | update_type |
      | start       |
      | end         |
      | start_end   |
      | no_type     |

  Scenario Outline: Admin update assignment time with invalid data
    Given admin update assignment time with "<update_type>"
    Then return with error "<error>"

    Examples:
      | update_type                | error            |
      | missing_studyplanitem_ids  | InvalidArgument  |
      | missing_student_id         | InvalidArgument  |
      | invalid_date               | InvalidArgument  |
      | another_invalid_date       | InvalidArgument  |

  Scenario Outline: Admin update assignemnt time with null data
    Given admin update assignment time with null data and "<update_type>"
    Then assignment time was updated with null data and according update_type "<update_type>"
  
    Examples:
      | update_type |
      | start       |
      | end         |
      | start_end   |
      | no_type     |

  Scenario Outline: Admin reupdate assignment time with valid data
    Given admin update assignment time with "<update_type>"
    Then assignment time was updated with according update_type "<update_type>"
    When admin re-update assignment time with "<update_type>"
    Then assignment time was updated with new data and according update_type "<update_type>"

    Examples:
      | update_type |
      | start       |
      | end         |
      | start_end   |
      | no_type     |

    