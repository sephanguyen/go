Feature: user publish topics

Background:
  Given a signed in "school admin"
  And user has created an empty book
  And user create a valid chapter
  And user has created some valid topics

Scenario: public missing topics
  When user public some missing topics
  Then returns "Internal" status code