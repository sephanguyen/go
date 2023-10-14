@quarantined
Feature: Import school info

  Scenario Outline: Import valid csv file
    Given an school info valid request payload with "<row condition>"
    When "<signed-in user>" importing school info
    Then the valid school info lines are imported successfully
    And the invalid school info lines are returned with error
    And returns "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |
      | school admin   | 500 rows               |

  Scenario Outline: Import invalid csv file
    Given an school info invalid "<invalid format>" request payload
    When "<signed-in user>" importing school info
    Then returns "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                    |
      | school admin   | no data                                           |
      | school admin   | header only                                       |
      | school admin   | number of column is not equal 6                   |
      | school admin   | mismatched number of fields in header and content |
      | school admin   | wrong school_id column name in header             |
      | school admin   | wrong school_name column name in header           |
      | school admin   | wrong school_name_phonetic column name in header  |
      | school admin   | wrong school_level_id column name in header            |
      | school admin   | wrong address column name in header                  |
      | school admin   | wrong is_archived column name in header           |

  Scenario Outline: Only school admin can import school info
    Given an school info valid request payload with "all valid rows"
    When "<signed-in user>" importing school info
    Then returns "<code>" status code

    Examples:
      | signed-in user | code             |
      | school admin   | OK               |
      | teacher        | PermissionDenied |
      | student        | PermissionDenied |
      | parent         | PermissionDenied |
