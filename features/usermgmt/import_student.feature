@quarantined
Feature: Import Student

  Scenario Outline: Import valid csv file
    Given a student valid request payload with "<row condition>"
    When "<signed-in user>" importing student
    Then the valid student lines are imported successfully
    And returns "OK" status code

    Examples:
      | signed-in user | row condition       |
      | school admin   | no row              |
      | school admin   | valid rows          |
      | school admin   | only mandatory rows |
      | school admin   | 1000 rows           |

  Scenario Outline: Import valid csv file with home address
    Given a student valid request payload home address with "<row condition>"
    When "<signed-in user>" importing student
    Then the valid student lines with home address are imported successfully
    And returns "OK" status code

    Examples:
      | signed-in user | row condition               | toggle  |
      | school admin   | no row                      | disable |
      | school admin   | valid rows                  | disable |
      | school admin   | only mandatory rows         | disable |
      | school admin   | valid row with grade master | disable |

  Scenario Outline: Import valid csv file with school history
    Given generate grade master
    And a student valid request payload school history with "<row condition>"
    When "<signed-in user>" importing student
    Then the valid student lines with school history are imported successfully
    And returns "OK" status code

    Examples:
      | signed-in user | row condition       | toggle  |
      | school admin   | no row              | disable |
      | school admin   | valid rows          | disable |
      | school admin   | only mandatory rows | disable |

  Scenario Outline: Import valid csv file with student phone number
    Given a student valid request payload home address with "<row condition>"
    When "<signed-in user>" importing student
    Then the valid student lines with student phone number imported successfully
    And returns "OK" status code

    Examples:
      | signed-in user | row condition                       | toggle  |
      | school admin   | valid row with student phone number | disable |

  Scenario Outline: Import invalid csv file
    Given a student info invalid "<invalid format>" request payload
    When "<signed-in user>" importing student
    Then the invalid student lines are returned with error code "<error code>"

    Examples:
      | signed-in user | invalid format                                          | error code                      |
      | school admin   | no data                                                 | emptyFile                       |
      | school admin   | 1001 rows                                               | invalidNumberRow                |
      | school admin   | with first name last name invalid rows                  | notFollowTemplate               |
      | school admin   | with first name last name email duplication rows        | duplicationRow                  |
      | school admin   | with first name last name phone_number duplication rows | duplicationRow                  |
      | school admin   | with first name last name missing mandatory             | missingMandatory                |
      | school admin   | with home address with invalid prefecture value         | notFollowTemplate               |
      | school admin   | with school history with invalid schoolID value         | notMatchDataRecordSchoolHistory |


  Scenario Outline: Only school admin can import school info
    Given a student valid request payload home address with "valid rows"
    When "<signed-in user>" importing student
    Then returns "<code>" status code

    Examples:
      | signed-in user | toggle  | code             |
      | school admin   | disable | OK               |
      | teacher        | disable | PermissionDenied |
      | student        | disable | PermissionDenied |
      | parent         | disable | PermissionDenied |


  Scenario Outline: Only school admin can import student support tag
    Given a student valid request payload tag with "<row condition>"
    When "<signed-in user>" importing student
    Then returns "OK" status code
    And after import amount tag have students

    Examples:
      | signed-in user | row condition       | toggle  |
      | school admin   | no row              | disable |
      | school admin   | only mandatory rows | disable |
      | school admin   | 1000 rows           | disable |
      | school admin   | valid row           | disable |