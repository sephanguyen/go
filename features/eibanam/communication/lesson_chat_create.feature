@chat
Feature: Create live lesson chat group
    Scenario: Participants see lesson chat group when teacher starts live lesson before student joins
        Given "school admin" logins CMS
        When school admin has created student with student info only
        And "teacher" logins Teacher App
        And "student" logins Learner App
        And school admin has created live lesson including student and teacher
        Given "teacher" has joined lesson
        Given "student" has joined lesson
        Then "teacher" sees lesson chat group is created
        And "student" sees lesson chat group is created


    Scenario: Participants see lesson chat group when teacher starts live lesson before student joins
        Given "jprep school admin" logins CMS
        When school admin has created student with student info only
        And "teacher" logins Teacher App
        And "student" logins Learner App
        And jprep school admin has created live lesson including student
        Given "teacher" has joined lesson
        Given "student" has joined lesson
        Then "teacher" sees lesson chat group is created
        And "student" sees lesson chat group is created