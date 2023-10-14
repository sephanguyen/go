Feature: List Chapters

  Background:
    Given a signed in "school admin"

  Scenario: List chapters by ids
    Given some chapters are existed in DB
    When student list chapters by ids
    Then return a list of chapters
