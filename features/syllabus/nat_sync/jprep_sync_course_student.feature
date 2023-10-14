Feature: JPREP sync course student

    Background: Generate CourseStudent
        Given <jprep_sync_course_student>generate course student

    Scenario: Sync courrse student
        When <jprep_sync_course_student>NAT JETSTREAM send a request
        Then <jprep_sync_course_student> store correct request to our system
