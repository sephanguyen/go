Feature: Test book hasura

    Scenario: BooksList
        Given a user insert some books to database
        When user call BooksTitle
        Then our system return BooksTitle correctly