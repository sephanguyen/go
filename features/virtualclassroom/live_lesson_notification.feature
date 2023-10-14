Feature: Send notifications to students and parents from upcoming live lessons

    Background:
        Given user signed in as school admin
        When enter a school
        And have some centers
        And have some teacher accounts
        And have some student accounts
        And have some courses
        And have parent accounts for students
        And students have the country "COUNTRY_JP"
        And student and parent accounts have device tokens
        And "enable" Unleash feature with feature name "Virtual_Classroom_SwitchNewDBConnection_Switch_DB_To_LessonManagement"

    Scenario: upcoming live lesson notifications are sent every "<interval>"
        Given user creates a live lesson with start time after "<interval>" for newly created students
        When wait for upcoming live lesson notification cronjob to run
        Then live lesson participants should receive notifications with message for "<interval>" interval
        Examples:
            | interval |
            | 24h      |
            | 15m      |