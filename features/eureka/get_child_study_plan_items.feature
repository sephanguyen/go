Feature: Get child study plan items

    Scenario: teacher get child study plan items
    Given a course and assigned this course to some students
    And a signed in "teacher"
    When teacher get child study plan items
    Then our system have to return child study plan items correctly