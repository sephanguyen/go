Feature: The student's learning material completion time

  Scenario Outline: The student's learning material completion time
    Given a valid "<status>" student submission in DB
    When <student_submissions>checking completed time of learning material
    Then <student_submissions>a valid "<completed time>" of learning material
    Examples:
      | status      | completed time |
      | completed   | valid          |
      | uncompleted | null           |
