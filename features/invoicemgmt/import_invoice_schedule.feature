@major
Feature: Import invoice schedule
  As an HQ manager or admin
  I am able to upload import invoice schedule

  # File content types used in the scenarios below determine the contents of the imported CSV file
  # - multiple-valid-dates - contains valid future dates, including next day and following months
  # - one-invoice-to-archive - contains an existing invoice schedule record for archival
  # - empty-file - contains no contents
  # - invalid-col-count - contains wrong number of CSV headers
  # - invalid-header - contains invalid CSV header name
  # - multiple-valid-and-invalid-dates - contains valid future dates and invalid dates (past and current day)

  Scenario Outline: Admin imports invoice schedule successfully
    Given "<signed-in user>" from "<org>" org in "<country>" country logins to backoffice app
    And there is no existing import schedule
    When "<signed-in user>" signed-in user imports invoice schedule file with "multiple-valid-dates" file content type
    Then receives "OK" status code
    And error list is empty
    And import schedule reflects in the DB based on "multiple-valid-dates" file content type
    And imported invoice schedules are converted in "JST"
    And the scheduled date is one day ahead of invoice_date

    Examples:
      | signed-in user | org         | country    |
      | school admin   | -2147483631 | COUNTRY_JP |
      | hq staff       | -2147483640 | COUNTRY_VN |
      | hq staff       | -2147483647 | NO_COUNTRY |


  Scenario Outline: Admin archives an invoice schedule successfully
    Given "<signed-in user>" from "<org>" org in "<country>" country logins to backoffice app
    And there is an existing import schedule
    When "<signed-in user>" signed-in user imports invoice schedule file with "one-invoice-to-archive" file content type
    Then receives "OK" status code
    And error list is empty
    And import schedule reflects in the DB based on "one-invoice-to-archive" file content type
    And the scheduled date is one day ahead of invoice_date

    Examples:
      | signed-in user | org         | country    |
      | school admin   | -2147483631 | COUNTRY_JP |

  Scenario Outline: Admin failed to import invoice schedule due to invalid format
    Given "<signed-in user>" from "<org>" org in "<country>" country logins to backoffice app
    When "<signed-in user>" signed-in user imports invoice schedule file with "<file-content-type>" file content type
    Then receives "InvalidArgument" status code
    And receives "<import-error>" import error

    Examples:
      | signed-in user | org         | country    | file-content-type | import-error                                              |
      | school admin   | -2147483631 | COUNTRY_JP | empty-file        | No data in CSV file                                       |
      | school admin   | -2147483631 | COUNTRY_JP | invalid-col-count | Invalid CSV format: number of column should be 4          |
      | school admin   | -2147483631 | COUNTRY_JP | invalid-header    | Invalid CSV format: first column should be 'invoice date' |

  Scenario: Admin failed to import invoice due to errors in CSV contents
    Given "<signed-in user>" from "<org>" org in "<country>" country logins to backoffice app
    When "<signed-in user>" signed-in user imports invoice schedule file with "multiple-valid-and-invalid-dates" file content type
    Then receives "OK" status code
    And error list is correct

    Examples:
      | signed-in user | org         | country    |
      | school admin   | -2147483631 | COUNTRY_JP |

  Scenario Outline: Admin imports duplicate invoice schedule
    Given "school admin" from "-2147483645" org in "COUNTRY_VN" country logins to backoffice app
    When "school admin" signed-in user imports invoice schedule file with "duplicate-valid-dates" file content type
    Then receives "OK" status code
    And error list is empty
    And import schedule reflects in the DB based on "duplicate-valid-dates" file content type
    And the scheduled date is one day ahead of invoice_date