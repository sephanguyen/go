Feature: Delete a class
    Background: 
        Given some centers
        And have some courses
        And have a class

    Scenario: admin try to delete class
        Given "staff granted role school admin" signin system
        When admin delete class
        Then returns "OK" status code
        And class have deleted successfully
        And Mastermgmt must push msg "DeleteClass" subject "MasterMgmt.Class.Upserted" to nats 