Feature: notification system sync Jprep student class data from Nats event of Bob
    @blocker
    Scenario: notificatiom systen create Jprep student class
        Given a new "staff granted role school admin" of Jprep organization with default location logged in Back Office
        And school admin creates "random" students
        And school admin creates "random" courses with "random" classes
        Given some valid upsert event sync student class from Bob
        When nats events are published
        Then notification system must store student class data correctly

    @blocker
    Scenario: notification system delete Jprep student class
        Given a new "staff granted role school admin" of Jprep organization with default location logged in Back Office
        And school admin creates "random" students
        And school admin creates "random" courses with "random" classes
        Given some valid upsert event sync student class from Bob
        When nats events are published
        Then notification system must store student class data correctly
        Given school admin "delete" some student class
        When nats events are published
        Then notification system must store student class data correctly
