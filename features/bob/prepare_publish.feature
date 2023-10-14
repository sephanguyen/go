@quarantined
Feature: prepare publish an uploading stream from client

    Background: 
        Given a valid lesson in database
            And some valid learners in database

    Scenario: The learner currently does not publish any uploading stream in the lesson and lesson's leaner counter less than maximum limit
    Given a "default" number of stream of the lesson
        And the arbitrary learner does not publishing any uploading stream in the lesson
    When the learner prepare to publish 
    Then the number of stream of the lesson have to increasing
        And new record indicating that the "default" learner is publishing an upload stream in the lesson 
        And returns "OK" status code
    
    Scenario: The learner currently does not pulish any uploading stream in the lesson and lesson's learner counter equal than maximum limit
    Given a "maximum" number of stream of the lesson
        And the arbitrary learner does not publishing any uploading stream in the lesson
    When the learner prepare to publish 
    Then the number of stream of the lesson have to unchanged
        And the learner is not allowed to publish any uploading stream in the lesson
        And returns the response "PREPARE_TO_PUBLISH_STATUS_REACHED_MAX_UPSTREAM_LIMIT"
    
    Scenario: The learner is currently publishing an uploading stream in the lesson
    Given an arbitrary learner publishing an uploading stream in the lesson
    When the learner prepare to publish 
    Then the number of stream of the lesson have to unchanged
        And the learner is still publishing an uploading stream in the lesson
        And returns the response "PREPARE_TO_PUBLISH_STATUS_PREPARED_BEFORE"
    
    Scenario: (Concurrent)Two learners concurrently prepare to publish an uploading stream to the same lesson and lesson's learner counter is default
    Given a "default" number of stream of the lesson
        And the "first" learner currently does not publish any uploading stream in the lesson
        And the "second" learner currently does not publish any uploading stream in the lesson
    When two learners prepare to publish in concurrently
    Then the lesson's learner counter have to increasing two
        And new record indicating that the "first" learner is publishing an upload stream in the lesson
        And new record indicating that the "second" learner is publishing an upload stream in the lesson 
        And returns "OK" status for both requests
    
    Scenario: (Concurrent)Two learners concurrently prepare to publish an uploading stream to the same lesson and there is only one available publish slot
    Given a "second maximum" number of stream of the lesson
        And the "first" learner currently does not publish any uploading stream in the lesson
        And the "second" learner currently does not publish any uploading stream in the lesson
    When two learners prepare to publish in concurrently
    Then the lesson's learner counter have to maximum
        And new record indicating that the either first learner or second is publishing an upload stream in the lesson
        And returns "OK" for the user who is granted
        And returns the response "PREPARE_TO_PUBLISH_STATUS_REACHED_MAX_UPSTREAM_LIMIT" for the user who is rejected
    
    Scenario: (Concurrent)The learners concurrently prepare to publish an uploading stream twice to the same lesson and lesson's learner counter equal than maximum limit
     Given a "second maximum" number of stream of the lesson
        And the arbitrary learner does not publishing any uploading stream in the lesson
    When the learner prepare to publish twice in concurrently
    Then the lesson's learner counter have to maximum
        And new record indicating that the "default" learner is publishing an upload stream in the lesson
        And returns "OK" for the one
        And returns the response "PREPARE_TO_PUBLISH_STATUS_PREPARED_BEFORE" for another
    
    Scenario: (Concurrent)The learner prepare to publish requests with the same lesson and the learner are send concurrently
        Given a "default" number of stream of the lesson
        And the arbitrary learner does not publishing any uploading stream in the lesson
        When the learner prepare to publish twice in concurrently
        Then the number of stream of the lesson have to become "1"
            And new record indicating that the "default" learner is publishing an upload stream in the lesson
            And returns "OK" for the one
            And returns the response "PREPARE_TO_PUBLISH_STATUS_PREPARED_BEFORE" for another
    