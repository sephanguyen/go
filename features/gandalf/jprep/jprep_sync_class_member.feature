Feature: JPREP sync class member in separate in API call

    Background: Prepare class for sync class member from JPREP
        Given jprep a valid admin
        And jprep a valid academic year
        And a request new class with m_regular_course payload
        And a valid JPREP signature in its header
        And the request master registration is performed
    
    Scenario: Send student join class with m_student payload
        Given a request new class member with m_student payload
        And a valid JPREP signature in its header
        When the request user registration is performed
        Then a "200" status is returned
        And the students must be store in our system
        And the class members must be store in our system
        And jprep Eureka must store class member with action ""
        And jprep Eureka must store course students with action ""

    Scenario: Send teacher join class with m_student and m_staff payload
        Given a request new teacher with m_staff payload
        And a valid JPREP signature in its header
        And the request user registration is performed
        And a request exist class member with m_student payload and action "upserted"
        And a valid JPREP signature in its header
        When the request user registration is performed
        Then a "200" status is returned
        And the students must be store in our system
        And the teachers must be store in our system
        And the class members must be store in our system
        And jprep Eureka must store class member with action ""
        And jprep Eureka must store course students with action ""

    Scenario: Send student leave class with m_student payload
        Given a request new class member with m_student payload
        And a valid JPREP signature in its header
        And the request user registration is performed
        And a request exist class member with m_student payload and action "upserted"
        And a valid JPREP signature in its header
        And the request user registration is performed
        And a request exist class member with m_student payload and action "deleted"
        And a valid JPREP signature in its header
        When the request user registration is performed
        Then a "200" status is returned
        And the students must be store in our system
        And the class members must be store in our system
        And jprep Eureka must store class member with action "deleted"
        And jprep Eureka must store course students with action "deleted"

    Scenario: Send teacher leave class with m_student payload
        Given a request new teacher with m_staff payload
        And a valid JPREP signature in its header
        And the request user registration is performed
        And a request exist class member with m_student payload and action "upserted"
        And a valid JPREP signature in its header
        And the request user registration is performed
        And a request exist class member with m_student payload and action "deleted"
        And a valid JPREP signature in its header
        When the request user registration is performed
        Then a "200" status is returned
        And the students must be store in our system
        And the teachers must be store in our system
        And the class members must be store in our system
        And jprep Eureka must store class member with action "deleted"
        And jprep Eureka must store course students with action "deleted"

    Scenario Outline: Send new teacher with only m_staff payload missing field
        Given a request new teacher with m_staff payload missing "<field>"
        And a valid JPREP signature in its header
        When the request user registration is performed
        Then a "400" status is returned
        And the teachers must not be store in our system
        Examples:
        | field        |
        | action kind  |
        | staff id     |
        | staff name   |

    Scenario Outline: Send new class member with only m_student payload missing field
        Given a request new class member with m_student payload missing "<field>"
        And a valid JPREP signature in its header
        When the request user registration is performed
        Then a "400" status is returned
        And the students must not be store in our system
        Examples:
        | field                     |
        | action kind               |
        | studentdivs.mstudentdivid |
        | student id                |
        | last name                 |
        | given name                |
        | regularcourses.mcourseid  |
        | regularcourses.startdate  |
        | regularcourses.enddate    |
