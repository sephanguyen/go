Feature: Upsert topic with flow automatically excutes display order

    Background:
        Given "school admin" logins
        And user has created an empty book
        And user create a valid chapter
    
    Scenario Outline: school admin has create some topics-new and old (both new, old topic) with another topic created before
        Given school admin has created some topics before
        When school admin has create some "<type>" topics
        Then our system have to save the topics correctly

        Examples:
            | type        | 
            | new         |
            | new and old |  

    Scenario: two school admin have create some topics in same time
        Given another school admin logins
        When two school admin create some topics
        Then our system have to store the topics in concurrency correctly

     Scenario: school admin upsert topics on old flow and new flow
        Given school admin has created some topics before by old flow
        When school admin has create some "new" topics
        Then our system have to save topics on both old and new flow correctly