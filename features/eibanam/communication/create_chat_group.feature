@chat
Feature: Create chat group

    Background:
        Given "school admin" logins CMS
        And "teacher" logins Teacher App

    Scenario: 1 student many parent 
        When school admin has created student with many parents' info
        And "student" logins Learner App successfully with credentials which school admin gives
        And "parents" login Learner App successfully with credentials which school admin gives
        Then student sees student chat group on Learner App
        And all student's parents see same parent chat group on Learner App
        And teacher sees both student chat group & parent chat group in Unjoined tab on Teacher App

    Scenario: 1 parent many student 
        Given school admin has created student with parent info
        When school admin created new student with same parent info
        And "newly created student" logins Learner App successfully with credentials which school admin gives
        And "parent" logins Learner App successfully with credentials which school admin gives
        Then "newly created student" sees new chat group on Learner App
        And parent sees 2 chat groups on Learner App
        And teacher sees 4 new chat groups in Unjoined tab on Teacher App
