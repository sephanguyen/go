Feature: JPREP sync live lesson in separate API call

    Background: Prepare course for sync live lesson from JPREP
        Given a request new class with m_regular_course payload
        And a valid JPREP signature in its header
        And the request master registration is performed
        And a "200" status is returned

    Scenario: Send new live lesson with m_lesson payload
        Given a request new live lesson with m_lesson payload
        And a valid JPREP signature in its header
        When the request master registration is performed
        Then a "200" status is returned
        And the lesson must be store in our system with action "upserted"
        And the lesson group must be store in our system
        And the preset study plan weekly must be store in our system
        And the topics must be store in our system with action "upserted"
        And the preset study plan must be store in our system
        And the course must be update in our system
        And jprep Tom must store conversation lesson
        And jprep Tom must store conversation with status "CONVERSATION_STATUS_NONE"

    Scenario: Send delete live lesson with m_lesson payload
        Given a request new live lesson with m_lesson payload
        And a valid JPREP signature in its header
        And the request master registration is performed
        And a request exist live lesson with m_lesson payload and action "deleted"
        And a valid JPREP signature in its header
        When the request master registration is performed
        Then a "200" status is returned
        And the lesson must be store in our system with action "deleted"
        And the preset study plan weekly must be delete in our system
        And the topics must be store in our system with action "deleted"
        
    Scenario: Send exist live lesson with m_lesson payload
        Given a request new live lesson with m_lesson payload
        And a valid JPREP signature in its header
        And the request master registration is performed
        And a request exist live lesson with m_lesson payload and action "upserted"
        And a valid JPREP signature in its header
        When the request master registration is performed
        Then a "200" status is returned
        And the lesson must be store in our system with action "upserted"
        And the lesson group must be store in our system
        And the preset study plan weekly must be store in our system
        And the topics must be store in our system with action "upserted"
        And the course must be update in our system

     Scenario Outline: Send new live lesson with m_lesson payload missing field
        Given a request new live lesson with m_lesson payload missing "<field>"
        And a valid JPREP signature in its header
        When the request master registration is performed
        Then a "400" status is returned
        And the lesson must not be store in our system
        And the lesson group must not be store in our system
        And the preset study plan must not be store in our system
        And the preset study plan weekly must not be store in our system
        And the topic must not be store in our system

        Examples:
        | field             |
        | action kind       |
        | lesson id         |
        | lesson name       |
        | lesson type       |
        | course id         |
        | start datetime    |
        | end datetime      |