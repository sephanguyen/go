Feature: Upsert student course packages
  As a school staff
  I need to be able to upsert student course packages

  Scenario Outline: Upsert student course packages
    Given student exist in our system
    And assign course package to exist student
    When "<signed-in user>" upsert student course packages
    Then upsert student course packages successfully
    And receives "OK" status code

    Examples:
      | signed-in user |
      | school admin   |

  Scenario Outline: student, teacher, parent don't have permission to upsert student course package
    Given student exist in our system
    And assign course package to exist student
    When "<signed-in user>" upsert student course packages
    Then receives "<msg>" status code

    Examples:
      | signed-in user  | msg              |
      | student         | PermissionDenied |
      | teacher         | PermissionDenied |
      | parent          | PermissionDenied |


  Scenario Outline: Upsert student course packages only add new course
    Given student exist in our system
    When "<signed-in user>" upsert student course packages with only "new course"
    Then upsert student course packages successfully
    And receives "OK" status code

    Examples:
      | signed-in user |
      | school admin   |
  
  Scenario Outline: Upsert student course packages only edit existed course
    Given student exist in our system
    And assign course package to exist student
    When "<signed-in user>" upsert student course packages with only "edit existed course"
    Then upsert student course packages successfully
    And receives "OK" status code

    Examples:
      | signed-in user |
      | school admin   |

  Scenario Outline: Upsert student course packages with student id invalid
    Given student exist in our system
    When "<signed-in user>" upsert student course packages with "<student id>" invalid
    Then "<signed-in user>" cannot upsert student course packages 
    And receives "<msg>" status code

    Examples:
      | signed-in user | student id | msg             |
      | school admin   | empty      | InvalidArgument |
      | school admin   | non-exist  | Internal        |

  Scenario Outline: Upsert student course packages with course invalid start date and end date
    Given student exist in our system
    When "<signed-in user>" upsert student course packages with course invalid start date and end date
    Then "<signed-in user>" cannot upsert student course packages
    And receives "InvalidArgument" status code

    Examples:
      | signed-in user |
      | school admin   |

  Scenario Outline: Upsert student course packages with package id empty
    Given student exist in our system
    When "<signed-in user>" upsert student course packages with package id empty
    Then "<signed-in user>" cannot upsert student course packages
    And receives "InvalidArgument" status code

    Examples:
      | signed-in user |
      | school admin   |

  Scenario Outline: Upsert student course packages with student package extra
    Given student exist in our system
    When "<signed-in user>" upsert student course packages with student package course extra
    Then upsert student course packages successfully with student package extra
    And receives "OK" status code

    Examples:
      | signed-in user |
      | school admin   |
#   Scenario Outline: Upsert student course packages with location ids empty
#     Given student exist in our system
#     When "<signed-in user>" upsert student course packages with location ids empty
#     Then "<signed-in user>" cannot upsert student course packages
#     And receives "InvalidArgument" status code
#
#     Examples:
#       | signed-in user |
#       | school admin   |