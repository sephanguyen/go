Feature: Create order enrollment

  Scenario Outline: Create order enrollment success
    Given prepare data for create order enrollment with "<valid data>"
    When "<signed-in user>" submit order
    Then order enrollment is created successfully
    And receives "OK" status code

    Examples:
      | signed-in user | valid data                             |
      | school admin   | order with single billed at order item |
      | hq staff       | order with single billed at order item |
      | centre manager | order with single billed at order item |
      | centre staff   | order with single billed at order item |
      # | school admin and teacher | order with single billed at order item |
      | hq staff       | order with single billed at order item |

  Scenario Outline: Create order enrollment failure
    Given request for create order enrollment with "<invalid data>"
    When "school admin" submit order
    Then receives "<status code>" status code
    And receives "<error message>" error message for create order enrollment with "<invalid data>"

    Examples:
      | invalid data                | status code        | error message                                      |
      | student is already enrolled | FailedPrecondition | invalid_student_enrollment_status_already_enrolled |
      | student is LOA in location  | FailedPrecondition | invalid_student_enrollment_status_on_loa           |
