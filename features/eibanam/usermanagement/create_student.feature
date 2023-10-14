@runsequence
Feature: Create Student on CMS

    Background:
        Given "school admin" logins CMS
        And "teacher" logins Teacher App

    Scenario: Create student with only student info
        When school admin creates a new student with student info
        Then school admin sees newly created student on CMS
        And student logins Learner App successfully with credentials which school admin gives

    Scenario: Create student with new parent
        When school admin creates a new student with parent info
        Then school admin sees newly created student on CMS
        And student logins Learner App successfully with credentials which school admin gives
        And new parent logins Learner App successfully with credentials which school admin gives
        And parent sees 1 student's stats on Learner App

    Scenario: Create student with existed parent
        Given school admin has created a student with parent info
        When school admin creates a new student with existed parent info
        Then school admin sees newly created student on CMS
        And student logins Learner App successfully with credentials which school admin gives
        And existed parent logins Learner App successfully with his existed credentials
        And parent sees 2 student's stats on Learner App

    Scenario Outline: Create student and associate student with course which has <condition>
        When school admin creates a new student with course which has "<condition>"
        Then school admin sees newly created student on CMS
        And teacher sees newly created student on Teacher App
        And student logins Learner App successfully with credentials which school admin gives
        And student "<result>" the course on Learner App when "<condition>"

        Examples:
            | condition                              | result       |
            | start date <= current date <= end date | sees         |
            | start date > current date              | does not see |
            | end time < current date                | does not see |

    Scenario: Create student with parent and associate student with course
        Given school admin has created a student with parent info
        When school admin creates a new student with new parent, existed parent and visible course
        Then school admin sees newly created student on CMS
        And teacher sees newly created student on Teacher App
        And student logins Learner App successfully with credentials which school admin gives
        And student sees the course on Learner App
        And new parent logins Learner App successfully with credentials which school admin gives
        And existed parent logins Learner App successfully with his existed credentials
        And all parent sees student's stats on Learner App