Feature: Run job to migrate current_grade to grade_id

  Scenario Outline: Migrate current_grade to grade_id
    Given generate grade master
    And list students with grade master
    When system run job to migrate current_grade to grade_id
    Then grade of students was migrated to grade_id
