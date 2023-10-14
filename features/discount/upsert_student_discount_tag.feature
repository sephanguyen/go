@blocker
Feature: Upsert Student Discount Tag
  As an HQ Staff, Admin,  Center Manager, Center Staff and Center Lead
  I can upsert sudent discount tag

  Background:
    Given there is an existing discount master data with discount tag

  Scenario Outline: Admin upserts user discount tag successfully (create new records with no existing records)
    Given there is an existing student with active products
    And this student has "no-existing" user discount tag "" records
    And a request payload for upsert user discount tag
    When "<signed-in user>" logins to backoffice app
    And upserts user discount tag "<discount-types>" records for this student
    And apply the upsert discount tags on the student
    Then receives "OK" status code
    And this student has correct user discount tag "<correct-discount-types>" records

    Examples:
      | signed-in user | discount-types                                             | correct-discount-types                                     |
      | hq staff       | single parent                                              | single parent                                              |
      | centre staff   | employee full time/employee part time                      | employee full time/employee part time                      |
      | centre manager | employee full time/employee part time/single parent/family | employee full time/employee part time/single parent/family |

  Scenario Outline: Admin upserts user discount tag successfully (create new records with existing records)
    Given there is an existing student with active products
    And this student has "existing" user discount tag "<existing-discount-types>" records
    And a request payload for upsert user discount tag
    When "<signed-in user>" logins to backoffice app
    And upserts user discount tag "<discount-types>" records for this student
    And apply the upsert discount tags on the student
    Then receives "OK" status code
    And this student has correct user discount tag "<correct-discount-types>" records

    Examples:
      | signed-in user | existing-discount-types                                                  | discount-types                                                           | correct-discount-types                                                   |
      | centre lead    | single parent                                                            | single parent/family                                                     | single parent/family                                                     |
      | centre staff   | combo/sibling                                                            | combo/sibling/employee full time/employee part time/single parent/family | combo/sibling/employee full time/employee part time/single parent/family |
      | hq staff       | combo/sibling/employee full time/single parent/family                    | combo/sibling/employee part time/employee full time/single parent/family | combo/sibling/employee full time/employee part time/single parent/family |
      | school admin   | combo/sibling/employee full time/employee part time/single parent/family | combo/sibling/employee full time/employee part time/single parent/family | combo/sibling/employee full time/employee part time/single parent/family |

  Scenario: Admin upserts user discount tag successfully w/ no existing records and no discount types selected
    Given there is an existing student with active products
    And this student has "no-existing" user discount tag "" records
    And a request payload for upsert user discount tag
    When "school admin" logins to backoffice app
    And upserts user discount tag "" records for this student
    And apply the upsert discount tags on the student
    Then receives "OK" status code
    And this student has no user discount tag records

  Scenario Outline: Admin upserts user discount tag successfully w/ deleting all existing records
    Given there is an existing student with active products
    And this student has "existing" user discount tag "<existing-discount-types>" records
    And a request payload for upsert user discount tag
    When "school admin" logins to backoffice app
    And upserts user discount tag "" records for this student
    And apply the upsert discount tags on the student
    Then receives "OK" status code
    And this student has no user discount tag records

    Examples:
      | signed-in user | existing-discount-types                                            |
      | school admin   | sibling/single parent                                              |
      | hq staff       | combo/employee full time/employee part time/single parent/family   |
      | hq staff       | combo/sibling                                                      |


  Scenario Outline: Admin upserts user discount tag successfully w/ existing records (create and delete)
    Given there is an existing student with active products
    And this student has "existing" user discount tag "<existing-discount-types>" records
    And a request payload for upsert user discount tag
    When "<signed-in user>" logins to backoffice app
    And upserts user discount tag "<discount-types>" records for this student
    And apply the upsert discount tags on the student
    Then receives "OK" status code
    And this student has correct user discount tag "<correct-discount-types>" records

    Examples:
      | signed-in user | existing-discount-types                                                  | discount-types                                               | correct-discount-types                                     |
      | centre lead    | single parent                                                            | family                                                       | family                                                     |
      | centre manager | combo/sibling/employee full time/employee part time/single parent/family | employee full time/employee part time/single parent/family   | employee full time/employee part time/single parent/family |

  Scenario Outline: Retrieve Active User Discount Tag with invalid student
    Given there is an invalid "<condition>" payload request for upsert user discount tag
    When "school admin" logins to backoffice app
    And upserts user discount tag "" records for this student
    And apply the upsert discount tags on the student
    Then receives "<status-code>" status code

    Examples:
      | condition      | status-code        |
      | empty student  | FailedPrecondition |
      | invalid data   | Internal           |




