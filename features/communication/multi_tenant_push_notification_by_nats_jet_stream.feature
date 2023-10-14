Feature: Push notification by publish message to nats jet stream

    Background:
        Given "2" schools with respective "staff granted role school admin" for each school login to CMS
        And school admin 0 has created 1 student with grade, course
        And school admin 1 has created 1 student with grade, course
        And student of school 1 login to Learner App

    @blocker
    Scenario Outline: cross-notification push through Nats Jetstream between tenants is prohibited
        Given school admin 0 push "<type>" notification to student of school admin 1
        And wait to "<type>" notification send
        Then student of school admin 1 must not receive notification
        Examples:
            | type      |
            | immediate |
            | schedule  |
