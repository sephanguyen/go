Feature: JPREP registers course and class in separate API call

    Scenario: Send with only m_course_name payload
        Given a request with m_course_name payload
        And a valid JPREP signature in its header
        When the request master registration is performed
        Then a "200" status is returned
        And the course must be registered in the system

    Scenario Outline: Send with only m_course_name payload missing field
        Given a request with m_course_name payload missing "<field>"
        And a valid JPREP signature in its header
        When the request master registration is performed
        Then a "400" status is returned
        And the course must not be registered in the system
        Examples:
        | field         |
        | action kind   |
        | course id     |
        | course name   |

    Scenario: Send delete course with only m_course_name payload
        Given a request with m_course_name payload
        And a valid JPREP signature in its header
        And the request master registration is performed
        And a request with current m_course_name payload and action "deleted"
        And a valid JPREP signature in its header
        When the request master registration is performed
        Then a "200" status is returned
        And the course must be registered in the system with action "deleted"