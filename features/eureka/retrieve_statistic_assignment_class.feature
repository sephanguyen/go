@quarantined
Feature: Retrieve statistic assignment class

    Scenario: the teacher retrieve statistic assignment class
    Given some students are assigned some valid study plans
        And some students submit their assignments
        And some students join in a class 
        And a signed in "teacher"
    When the teacher retrieve statistic assignment class
    Then our system have to return statistic assignment class correctly