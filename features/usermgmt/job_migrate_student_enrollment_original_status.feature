Feature: Run migration job to migrate update student enrollment original status to new status in our system

  Scenario Outline: Migrate update student enrollment original status to new status in our system
    Given students with enrollment original status in our system
    When system run job to migrate update student enrollment status "<original status>" to "<new status>" in our system "<resource path>"
    Then students have enrollment status "<original status>" are updated to "<new status>" with "<resource path>"

    Examples:
      | original status                | new status                          | resource path  |
      | STUDENT_ENROLLMENT_STATUS_LOA | STUDENT_ENROLLMENT_STATUS_WITHDRAWN | MANABIE_SCHOOL |


