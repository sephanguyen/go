Feature: Create retry quiz test

   Background: create quiz background
        Given <shuffled_quiz_set> a signed in "school admin"
        And <shuffled_quiz_set> a valid book content
        And user create a quiz using v2
        And <shuffled_quiz_set> a signed in "student"
        And school admin add student to a course have a study plan
    
    Scenario Outline: auth create a retry quiz test
        Given <shuffled_quiz_set> a signed in "<role>"
        And school admin add student to a course have a study plan
        And user create quiz test v2 
        And <shuffled_quiz_set> returns "OK" status code
        When user create retry quiz test v2
        Then <shuffled_quiz_set> returns "<status code>" status code
        Examples:
            | role           | status code |
            | school admin   | OK          |
            | admin          | OK          |
            | teacher        | OK          |
            | student        | OK          |
            | hq staff       | OK          |
            | center lead    | OK          |
            | center manager | OK          |
            | center staff   | OK          |
            | lead teacher   | OK          |

    Scenario: school admin create a retry quiz test
        Given user create quiz test v2
        And <shuffled_quiz_set> returns "OK" status code
        When user create retry quiz test v2
        Then <shuffled_quiz_set> returns "OK" status code 
        And retry shuffled quiz test have been stored

    Scenario Outline: auth create a retry quiz test for LO having question groups
        Given <shuffled_quiz_set> a signed in "<role>"
        Given <2> existing question group
        Given <2> quiz belong to question group
        And school admin add student to a course have a study plan
        And user create quiz test v2 
        And <shuffled_quiz_set> returns "OK" status code
        When user create retry quiz test v2
        Then <shuffled_quiz_set> returns "<status code>" status code
        And retry shuffled quiz test have been stored
        And user got quiz test response
        Examples:
            | role           | status code |
            | school admin   | OK          |
            | admin          | OK          |
            | teacher        | OK          |
            | student        | OK          |
            | hq staff       | OK          |
            | center lead    | OK          |
            | center manager | OK          |
            | center staff   | OK          |
            | lead teacher   | OK          |