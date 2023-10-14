Feature: Upsert study plan item feature
     Background: Background has study plan
          Given a valid course and study plan background
          And a study plan name "test-study-plan" in db

     Scenario: Upsert study plan item
          Given a valid "teacher" token
          When user upsert a list of study plan item
          Then returns "OK" status code
          Then eureka must store correct study plan item

     Scenario: Upsert study plan item when study plan does not have any study plan items
          Given a valid "teacher" token
          And a study plan does not have any study plan items
          When user upsert a list of study plan item
          Then returns "OK" status code
          And book of study plan has stored correctly