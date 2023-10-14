Feature: Import classes

   Background: 
      Given some centers
      And have some courses

   Scenario: admin try to import classes success
      Given "staff granted role school admin" signin system
      And a valid classes payload
      When user import classes by csv file
      Then returns "OK" status code
      And the valid classes was created
      And Mastermgmt must push msg "CreateClass" subject "MasterMgmt.Class.Upserted" to nats 

   Scenario: admin try to import classes failed
      Given "staff granted role school admin" signin system
      And a valid and invalid classes payload
      When user import classes by csv file
      Then returns "InvalidArgument" status code
    