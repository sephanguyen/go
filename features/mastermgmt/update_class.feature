Feature: Update a class
    Background: 
        Given some centers
        And have some courses
        And have a class

    Scenario: admin try to edit class
        Given "staff granted role school admin" signin system
        When admin update class
        Then returns "OK" status code
        And class have updated successfully
        And Mastermgmt must push msg "UpdateClass" subject "MasterMgmt.Class.Upserted" to nats 