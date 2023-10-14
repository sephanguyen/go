Feature: List Books

  Scenario: List books by ids
    Given some books are existed in DB
    When student list books by ids
    Then return a list of books
