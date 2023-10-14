Feature: Run job to increase grade of students

  Scenario Outline: Increase grade of students
    Given list students with different grade and different school
    When system run job to increase grade of students
    Then grade of students was increased by "<level>" level

     Examples:
      | level |
      | 1     |
