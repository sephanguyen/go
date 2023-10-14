@quarantined
Feature:  unpublish an uploading stream from client

    Background: 
        Given a valid lesson in database
            And some valid learners in database

    Scenario: Unpublish the learner who prepared to publish before
    Given a "default" number of stream of the lesson
        And a learner prepared publish in the lesson
    When the learner unpublish 
    Then the number of stream of the lesson have to equal to "default" number
        And no record indicating that the "default" learner is unpublish an upload stream in the lesson 
        And returns "OK" status code
    
    Scenario: Unpublish the learner who does not prepare to publish before
    Given a "default" number of stream of the lesson
        And the arbitrary learner does not publishing any uploading stream in the lesson
    When the learner unpublish 
    Then the number of stream of the lesson have to unchanged
        # And no new record indicating that the "default" learner is unpublish an upload stream in the lesson 
        And unpublish returns the response "UNPUBLISH_STATUS_UNPUBLISHED_BEFORE"
        And returns "OK" status code

    Scenario: (Concurrent)Two learners unpublish their streams in the same lesson at the same time
    Given a "default" number of stream of the lesson
        And two learner prepared publish in the lesson
        And record indicating the two learner prepared publish in the lesson
    When two learners unpublish in concurrently
    Then the number of stream of the lesson have to equal to "default" number
        And no record indicating that the "first" learner prepared to publish an upload stream in the lesson 
        And no record indicating that the "second" learner prepared to publish an upload stream in the lesson 
        And unpublish returns "OK" status code for both requests
    
    Scenario: (Concurrent)A learner ubpublishes uploading her streams twice at the same time 
     Given a "default" number of stream of the lesson
        And a learner prepared publish in the lesson
    When the learner unpublish twice in concurrently
    Then the number of stream of the lesson have to equal to "default" number
        And no record indicating that the "default" learner prepared to publish an upload stream in the lesson  
        And returns OK for the one
        And unpublish returns the response "UNPUBLISH_STATUS_UNPUBLISHED_BEFORE" for another
    
    Scenario: (Concurrent)The learner requests unpublish twice with the same lesson concurrently
        Given a "second maximum" number of stream of the lesson
            And the arbitrary learner does not publishing any uploading stream in the lesson
            And a learner prepared publish in the lesson
        When the learner unpublish twice in concurrently
        Then the number of stream of the lesson have to decrease "1"
            And no record indicating that the "default" learner prepared to publish an upload stream in the lesson
            And returns "OK" for the one
            And returns the response "UNPUBLISH_STATUS_UNPUBLISHED_BEFORE" for another

