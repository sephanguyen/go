@quarantined
Feature: Import study plan features

   Background: valid course background
      Given a valid "teacher" token
      And a valid course background
      And valid assignment in db

   Scenario: Import course study plan
      When user import a "valid" study plan to course
      Then returns "OK" status code
      And wait for assign study plan task to completed
      And eureka must store and assign study plan correctly

   Scenario: Import individual study plan
      Given a valid "teacher" token
      When user import a individual study plan to a student
      Then returns "OK" status code
      And wait for assign study plan task to completed
      And eureka must store and assign study plan for individual student correctly

   Scenario: Import course study plan and update
      Given user import a generated study plan to course
      When user update course study plan
      Then returns "OK" status code
      And wait for assign study plan task to completed
      And course study plan and individual student study plan must be update

   Scenario: Import course study plan and remove one row from study plan
      Given user import a generated study plan to course
      When user remove one row from study plan
      And user update course study plan
      Then returns "OK" status code
      And wait for assign study plan task to completed
      And course study plan and individual student study plan must be update

   Scenario: Import course study plan and insert one row from study plan
      Given user import a generated study plan to course
      When user insert one row to study plan
      And user update course study plan
      Then returns "OK" status code
      And wait for assign study plan task to completed
      And course study plan and individual student study plan must be update

   Scenario: Import individual study plan and update
      Given user import a generated study plan to course
      And user download a student's study plan
      When user update individual study plan
      Then returns "OK" status code
      And wait for assign study plan task to completed
      And individual study plan should be update

   Scenario: Import individual study plan and insert one row
      Given user import a generated study plan to course
      And user download a student's study plan
      When user insert one row to study plan
      And user update individual study plan
      Then returns "OK" status code
      And wait for assign study plan task to completed
      And individual study plan should be update

   Scenario: Import individual study plan and remove one row
      Given user import a generated study plan to course
      And user download a student's study plan
      When user remove one row from study plan
      And user update individual study plan
      Then returns "OK" status code
      And wait for assign study plan task to completed
      And individual study plan should be update

   Scenario: Import course study plan and change study plan item display order
      Given user import a generated study plan to course
      And user download a student's study plan
      When user change study plan item order
      And user update individual study plan
      Then returns "OK" status code
      And wait for assign study plan task to completed
      And individual study plan should be update

   Scenario: The user update study plan in type individual
      Given some study plans to individual
      When the user update study plans
      Then our system have to update study plan correctly

   Scenario: Import multi course with duplicate id
      When user import a "invalid" study plan to course
      Then returns "InvalidArgument" status code
         And our system have to handle error study plan correctly

   Scenario: Import course with more than one book
      When user import a "invalid with books" study plan to course
      Then returns "InvalidArgument" status code

   Scenario: Import course study plan and update
      Given some study plans to individual
      And make study plan item completed
      When user update course study plan with times
      Then returns "OK" status code
      Then study plan item still completed

   Scenario: Import course study plan and update with study_plan_items don't belong to study_plan
      Given some study plans to individual
      And make study plan item completed
      When user update course study plan with times with study_plan_items don't belong to study_plan
      Then returns "InvalidArgument" status code

   Scenario: Import course study plan and upsert new LOs
      When user import a "valid" study plan to course
      Then returns "OK" status code
      And wait for assign study plan task to completed
      And eureka must store and assign study plan correctly
      And user try to upsert "valid" learning objectives using APIv1
      And returns "OK" status code
      And all study plan items were created with status when upsert new los

   Scenario: Import course study plan and upsert new assignments
      When user import a "valid" study plan to course
      Then returns "OK" status code
      And wait for assign study plan task to completed
      And eureka must store and assign study plan correctly
      And user try to upsert assignments
      And returns "OK" status code
      And all study plan items were created with status when upsert new assignments

   Scenario: Import course study plan and update
      When some study plans to individual
      Then all study plan items were created with status active after import study plan with type create