Feature: Run Job migrate enrollment status success
  i want to import data enrollment status from csv file

  Scenario: Run Job migrate enrollment status success
    Given a signed in "staff granted role school admin"
    And setup data to migrate enrollment status
    When run job migrate enrollment status
    Then check data job migration enrollment status correct
