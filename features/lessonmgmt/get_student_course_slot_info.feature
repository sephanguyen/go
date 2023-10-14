Feature: Receive student course slot info through an event 

  Background:
    Given user signed in as school admin
    And have some centers
    And have some student accounts
    And have some courses

  Scenario Outline: Receive student course slot info
    Given a signed in admin
    When a message is published to student course event sync
    Then receive student course slot info successfully
