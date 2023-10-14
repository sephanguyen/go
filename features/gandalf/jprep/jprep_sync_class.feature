Feature: JPREP sync class in separate API call

    Background: Prepare academic year for sync class from JPREP
        Given jprep a valid admin
        And jprep a valid academic year

    Scenario: Send new class with only m_regular_course payload
        Given a request new class with m_regular_course payload
        And a valid JPREP signature in its header
        When the request master registration is performed
        Then a "200" status is returned
        And the class must be store in our system

    Scenario: Send exist class with only m_regular_course payload and action upserted
        Given a request new class with m_regular_course payload
        And a valid JPREP signature in its header
        And the request master registration is performed
        And a request exist class with m_regular_course payload with action "upserted"
        And a valid JPREP signature in its header
        When the request master registration is performed
        Then a "200" status is returned
        And the class must be store in our system

    Scenario: Send exist class with only m_regular_course payload and action deleted
        Given a request new class with m_regular_course payload
        And a valid JPREP signature in its header
        And the request master registration is performed
        And a request exist class with m_regular_course payload with action "deleted"
        And a valid JPREP signature in its header
        When the request master registration is performed
        Then a "200" status is returned
        And the class must be store in our system

    Scenario Outline: Send new class with only m_regular_course payload missing field
        Given a request new class with m_regular_course payload missing "<field>"
        And a valid JPREP signature in its header
        When the request master registration is performed
        Then a "400" status is returned
        And the class must not be store in our system
        And the course class must not be store in our system
        And the course academic year must not be store in our system
        Examples:
        | field             |
        | action kind       |
        | course id         |
        | class id          |
        | class name        |
        | start date        |
        | end date          |
    
    # TODO 1: Check course id is existing in our system
    # TODO 2: Check academic year id is existing in our system
