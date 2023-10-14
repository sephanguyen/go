@major
Feature: Import partner bank master
  As an HQ manager or admin
  I am able to upload import partner bank on master management

  Scenario Outline: Admin imports partner bank successfully
    Given a request payload file with "not-exist" partner bank "<record-type>" record
    When "<signed-in user>" logins to backoffice app 
    And imports partner bank "<record-type>" records
    Then receives "OK" status code 
    And partner bank csv is imported successfully

    Examples:
      | signed-in user   | record-type                 |
      | school admin     | single-valid                |
      | hq staff         | single-valid                |
      | school admin     | single-valid-limit          |
      | hq staff         | single-valid-limit          |
      | school admin     | multiple-with-default-valid |
      | hq staff         | multiple-with-default-valid |

  Scenario Outline: Admin archive partner bank successfully
    Given a request payload file with "existing" partner bank "<record-type>" record
    When "<signed-in user>" logins to backoffice app 
    And archives "<record-type>" partner bank records
    Then receives "OK" status code 
    And partner bank csv is archived successfully

    Examples:
      | signed-in user   | record-type                 |
      | school admin     | single-valid                |
      | hq staff         | single-valid                |
      | school admin     | single-valid-limit          |
      | hq staff         | single-valid-limit          |
      | school admin     | multiple-with-default-valid |
      | hq staff         | multiple-with-default-valid |

  Scenario: Admin failed to import partner bank to errors in CSV contents
    Given a request payload file with "not-existing" partner bank "<record-type>" record
    When "<signed-in user>" logins to backoffice app
    And imports invalid partner bank records
    Then receives "OK" status code 
    And partner bank csv "<record-type>" record is imported unsuccessfully
  
    Examples:
      | signed-in user      |  record-type                            |
      | school admin        |  multiple-with-default-invalid-required |
      | hq staff            |  multiple-with-default-invalid-format   |