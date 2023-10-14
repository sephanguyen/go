Feature: notification system sync Jprep student package data from Nats even of Yasuo
    @blocker
    Scenario: notificatiom systen create Jprep student package
        Given a new "staff granted role school admin" of Jprep organization with default location logged in Back Office
        And school admin creates "random" students
        And school admin creates "random" courses
        Given some valid upsert event sync student course from Yasuo
        When nats events are published
        Then notification system must store student course data correctly with type "upsert"

    @blocker
    Scenario: notification system update Jprep student package
        Given a new "staff granted role school admin" of Jprep organization with default location logged in Back Office
        And school admin creates "random" students
        And school admin creates "random" courses
        Given some valid upsert event sync student course from Yasuo
        When nats events are published
        Then notification system must store student course data correctly with type "upsert"
        Given school admin "update" some student course
        When nats events are published
        Then notification system must store student course data correctly with type "upsert"

    @blocker
    Scenario: notification system delete Jprep student package
        Given a new "staff granted role school admin" of Jprep organization with default location logged in Back Office
        And school admin creates "random" students
        And school admin creates "random" courses
        Given some valid upsert event sync student course from Yasuo
        When nats events are published
        Then notification system must store student course data correctly with type "upsert"
        Given school admin "delete" some student course
        When nats events are published
        Then notification system must store student course data correctly with type "delete"
