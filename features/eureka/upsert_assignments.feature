Feature: Upsert Assignments

   Scenario: Create assignments
      Given a valid "teacher" token
      When user create new assignments
      Then returns "OK" status code
      Then eureka must store correct assignments

   Scenario: Update assignments
      Given a valid "teacher" token
      When user update assignments
      Then returns "OK" status code
      Then eureka must store correct assignments

   Scenario: Create assignments
      Given a valid "teacher" token
      When user create assignment with empty assignment_id
      Then returns "OK" status code
      Then eureka must store correct assignment when create assignment with empty assignment_id


   Scenario: Create some assignments same topic and time
      Given a valid "teacher" token
      And user create some assignments same topic and time
      When retrieve assignments with that topic
      Then returns a assignment list with different display order

   Scenario: Create assignments with display order
      Given a valid "teacher" token
      And user create assignments with display order
      When retrieve assignments
      Then returns assignment list with display order correctly

   Scenario: Create assignments without display order
      Given a valid "teacher" token
      And user create assignments without display order
      When retrieve assignments with that topic
      Then returns a assignment list with different display order

