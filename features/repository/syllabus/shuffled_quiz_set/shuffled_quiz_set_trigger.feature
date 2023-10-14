Feature: trigger shuffle quiz set

    Scenario:trigger study plan item identity of shuffle quiz set
        Given user create a study plan of assignment to database
        When user create a shuffle quiz set
        Then our system stored study plan item identity of shuffle quiz set correctly
