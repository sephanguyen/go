Feature: JPREP sync course class

    Background: Generate CourseClass
        Given <jprep_sync_course_class>generate course class

    Scenario: Sync courrse classf
        Given <jprep_sync>a signed in "<role>"
        When <jprep_sync_course_class>NAT JETSTREAM send a request
        Then <jprep_sync_course_class> store correct request to our system
