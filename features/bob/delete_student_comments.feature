Feature: delete comment in student

    The teacher delete student's comment 

    Scenario: a teacher delete write comment for a student
    Given "staff granted role teacher" signin system
    And a student with some comments
    When the teacher delete student's comment
    And our systems have to store comment correctly
