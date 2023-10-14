@quarantined
Feature: Retrieve Learning Progress
    In order to track the learning time details
    As an student
    I need to retrieve my Learning Progress (LP)

    Scenario: unauthenticated student try to retrieve LP
        Given an invalid authentication token
        When student retrieves LP
        Then returns "Unauthenticated" status code

    Scenario: invalid filter range
        Given a signed in "student" with filter range is "invalid" 
        When student retrieves LP
        Then returns "InvalidArgument" status code

    Scenario: student try to retrieve LP of another student
        Given a signed in "student" with filter range is "valid"
            And an other student profile in DB
        When student retrieves LP
        Then returns "NotFound" status code

    Scenario: student try to retrieve LP but hasn't learned any LO
        Given a signed in "student" with filter range is "valid" use his owned student UUID
            And student hasn't learned any LO
        When student retrieves LP
        Then returns LP with all total_time_spent_in_day equal to zero

    Scenario: student has been learning an LO and try to retrieve LP
        Given a signed in "student" with filter range is "valid" use his owned student UUID
            And a list of learning_objective event logs
            And a student inserts a list of event logs
        When student retrieves LP
        Then returns "OK" status code
            And returns LP with some total_time_spent_in_day larger than zero

    Scenario: student has been learning an LO and try to retrieve LP
        Given a signed in "student" with filter range is "valid" use his owned student UUID
            And a random number

            And a "started" learning objective event log with session "S1" at "2019-12-12T19:00:00Z"
            And a student inserts a list of event logs

        When student retrieves LP
        Then returns "OK" status code
            And returns LP with all total_time_spent_in_day equal to zero

    Scenario: student has been learning an LO and try to retrieve LP
        Given a signed in "student" with filter range is "valid" use his owned student UUID
            And a random number

            And a "started" learning objective event log with session "S1" at "2019-12-12T10:00:00Z"
            And a student inserts a list of event logs

            And previous request data is reset

            And a "completed" learning objective event log with session "S1" at "2019-12-12T10:03:00Z"
            And a student inserts a list of event logs

        When student retrieves LP
        Then returns "OK" status code
            And total_learning_time at "2019-12-11T17:00:00Z" must be "180"
            And total_learning_time at "2019-12-12T17:00:00Z" must be "0"

    Scenario: student has been learning an LO and try to retrieve LP
        Given a signed in "student" with filter range is "valid" use his owned student UUID

            And a random number

            And a "started" learning objective event log with session "S1" at "2019-12-12T19:00:00Z"
            And a student inserts a list of event logs

            And previous request data is reset

            And a "completed" learning objective event log with session "S1" at "2019-12-12T19:03:00Z"
            And a student inserts a list of event logs

        When student retrieves LP
        Then returns "OK" status code
            And total_learning_time at "2019-12-11T17:00:00Z" must be "0"
            And total_learning_time at "2019-12-12T17:00:00Z" must be "180"

    Scenario: student has been learning an LO and try to retrieve LP
        Given a signed in "student" with filter range is "valid" use his owned student UUID

            And a random number

            And a "started" learning objective event log with session "S1" at "2019-12-12T19:00:00Z"
            And a student inserts a list of event logs

            And previous request data is reset

            And a "exited" learning objective event log with session "S1" at "2019-12-12T19:03:00Z"
            And a student inserts a list of event logs

        When student retrieves LP
        Then returns "OK" status code
            And total_learning_time at "2019-12-11T17:00:00Z" must be "0"
            And total_learning_time at "2019-12-12T17:00:00Z" must be "180"

    Scenario: student has been learning an LO and try to retrieve LP
        Given a signed in "student" with filter range is "valid" use his owned student UUID

            And a random number

            And a "started" learning objective event log with session "S1" at "2019-12-12T19:00:00Z"
            And a student inserts a list of event logs

            And previous request data is reset

            And a "completed" learning objective event log with session "S1" at "2019-12-12T19:03:00Z"
            And a student inserts a list of event logs

            And student retrieves LP
            And total_learning_time at "2019-12-12T17:00:00Z" must be "180"

            And previous request data is reset

            And a "exited" learning objective event log with session "S1" at "2019-12-12T19:03:25Z"
            And a student inserts a list of event logs

        When student retrieves LP
        Then returns "OK" status code
            And total_learning_time at "2019-12-11T17:00:00Z" must be "0"
            And total_learning_time at "2019-12-12T17:00:00Z" must be "180"

    Scenario: student has been learning an LO and try to retrieve LP
        Given a signed in "student" with filter range is "valid" use his owned student UUID

            And a random number

            And a "started" learning objective event log with session "S1" at "2019-12-12T19:00:00Z"
            And a student inserts a list of event logs

            And previous request data is reset

            And a "completed" learning objective event log with session "S1" at "2019-12-12T19:03:00Z"
            And a "exited" learning objective event log with session "S1" at "2019-12-12T19:03:25Z"
            And a student inserts a list of event logs

        When student retrieves LP
        Then returns "OK" status code
            And total_learning_time at "2019-12-11T17:00:00Z" must be "0"
            And total_learning_time at "2019-12-12T17:00:00Z" must be "180"

    Scenario: student has been learning an LO and try to retrieve LP
        Given a signed in "student" with use his owned student UUID

            And a random number
            And a filter range with "from" is "2019-12-09T00:00:00Z"
            And a filter range with "to" is "2019-12-15T23:59:59Z"

            And a "started" learning objective event log with session "S1" at "2019-12-12T19:00:00Z"
            And a student inserts a list of event logs

            And previous request data is reset

            And a "completed" learning objective event log with session "S1" at "2019-12-12T19:03:00Z"
            And a student inserts a list of event logs

        When student retrieves LP
        Then returns "OK" status code
            And total_learning_time at "2019-12-11T17:00:00Z" must be "0"
            And total_learning_time at "2019-12-12T17:00:00Z" must be "180"

    Scenario: student has been learning an LO and try to retrieve LP
        Given a signed in "student" with use his owned student UUID

            And a random number
            And a filter range with "from" is "2019-12-09T00:00:00Z"
            And a filter range with "to" is "2019-12-15T23:59:59Z"

            And a "started" learning objective event log with session "S1" at "2019-12-12T19:00:00Z"
            And a "paused" learning objective event log with session "S1" at "2019-12-12T19:02:00Z"
            And a "paused" learning objective event log with session "S1" at "2019-12-12T19:02:30Z"
            And a "resumed" learning objective event log with session "S1" at "2019-12-12T19:02:10Z"
            And a "completed" learning objective event log with session "S1" at "2019-12-12T19:03:10Z"
            And a student inserts a list of event logs

        When student retrieves LP
        Then returns "OK" status code
            And total_learning_time at "2019-12-11T17:00:00Z" must be "0"
            And total_learning_time at "2019-12-12T17:00:00Z" must be "180"

    Scenario: student has been learning an LO and try to retrieve LP
        Given a signed in "student" with use his owned student UUID

            And a random number
            And a filter range with "from" is "2019-12-09T00:00:00Z"
            And a filter range with "to" is "2019-12-15T23:59:59Z"

            And a "started" learning objective event log with session "S1" at "2019-12-12T19:00:00Z"
            And a "paused" learning objective event log with session "S1" at "2019-12-12T19:02:00Z"
            And a "resumed" learning objective event log with session "S1" at "2019-12-12T19:02:10Z"
            And a "resumed" learning objective event log with session "S1" at "2019-12-12T19:02:15Z"
            And a "completed" learning objective event log with session "S1" at "2019-12-12T19:03:10Z"
            And a student inserts a list of event logs

        When student retrieves LP
        Then returns "OK" status code
            And total_learning_time at "2019-12-11T17:00:00Z" must be "0"
            And total_learning_time at "2019-12-12T17:00:00Z" must be "180"
