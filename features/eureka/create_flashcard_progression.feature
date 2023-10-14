Feature: Create The Flashcard study
    In order to finish learning an objective or doing an exam
    As a student
    I need to take the quiz

    Background: given a quizet of an learning objective
        Given a quizset with "21" quizzes in Learning Objective belonged to a "TOPIC_TYPE_EXAM" topic
        And a signed in "school admin"

    Scenario: unauthenticated user try to create the flashcard study test
        Given an invalid authentication token
        When user create flashcard study with valid request and limit "3" in the first time
        Then returns "Unauthenticated" status code

    Scenario: a admin try to create the flashcard study first time
        Given a signed in "school admin"
        When user create flashcard study with valid request and limit "3" in the first time
        Then returns "PermissionDenied" status code

    Scenario: a student try to create the flashcard study first time
        Given a signed in "student"
        And a study plan item id
        When user create flashcard study with valid request and limit "3" in the first time
        Then return list of "3" flashcard study items

    Scenario: a student try to create the flashcard study and get next page
        Given a signed in "student"
        And a study plan item id
        When student doing a long exam with "5" flashcard study quizzes per page
        Then that student can fetch the list of flashcard study quizzes page by page using limit "5" flashcard study quizzes per page

    Scenario: a student try to create the flashcard study but missing loID
        Given a signed in "student"
        When user create flashcard study without loID
        Then returns "InvalidArgument" status code

    Scenario: a student try to create the flashcard study but missing paging
        Given a signed in "student"
        When user create flashcard study without paging
        Then returns "InvalidArgument" status code

    Scenario: a student try to create the flashcard with offset larger than number of quizzes
        Given a signed in "student"
        And a study plan item id
        When user create flashcard study with valid request and offset "1000" and limit "100"
        Then returns empty flashcard study items

    Scenario: a student try to create the flashcard study and get next page
        Given a signed in "student"
        # cause study plan item is from service enigma. So it's assumed that we have the study plan id
        And a study plan item id
        When student doing a long exam with "5" flashcard study quizzes per page
        Then that student can fetch the list of flashcard study quizzes page by page using limit "5" flashcard study quizzes per page

    Scenario Outline: a student try to create the quiz test
        Given a signed in "student"
        And a study plan item id
        When user create flashcard study with "<action>" and "<limit>" flashcard study quizzes per page
        Then returns expected list of flashcard study quizzes with "<action>"

        Examples:
            | action     | limit |
            | shuffled   | 1     |
            | keep order | 2     |
            | shuffled   | 3     |
            | keep order | 4     |
            | keep order | 5     |
            | keep order | 6     |
            | shuffled   | 7     |
            | shuffled   | 13    |
