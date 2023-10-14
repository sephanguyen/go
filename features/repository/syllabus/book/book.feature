Feature: Book repository

    Scenario: FindByID
        Given a user insert a book to database
        When user get book by call FindByID
        Then our system return the book correctly