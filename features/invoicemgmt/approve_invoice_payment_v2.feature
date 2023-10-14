@major
Feature: Approve Invoice Payment V2
  As an HQ manager or admin
  I am able to approve a payment on an invoice

  Scenario Outline: Admin approves an invoice payment successfully
    Given "<signed-in user>" logins to backoffice app
    And there is an existing invoice with "ISSUED" status
    And admin already requested payment with amount same on invoice outstanding balance
    And this payment has payment "<payment-method>" payment method
    And admin added the requested payment on the invoice
    And admin sets the approve payment form with "<payment-date>" payment date
    When admin submits the approve payment form with remarks "<remarks>" using v2 endpoint
    Then receives "OK" status code
    And latest payment record has "SUCCESSFUL" payment status
    And latest payment record has receipt date today
    And invoice has "PAID" invoice status
    And invoice amount paid is equal to payment amount
    And invoice has zero outstanding balance
    And action log record is recorded with "PAYMENT_APPROVED" action and "<remarks>" remarks

    Examples:
      | signed-in user | payment-method | remarks | payment-date |
      | school admin   | CASH           | any     | TODAY        |
      | hq staff       | BANK_TRANSFER  |         | TODAY        |

  Scenario Outline: Admin failed to approve a payment with an invalid invoice status
    Given "<signed-in user>" logins to backoffice app
    And there is an existing invoice with "<invoice-status>" status
    And there is "PENDING" payment history with "<payment-method>" payment method
    And admin sets the approve payment form with "<payment-date>" payment date
    When admin submits the approve payment form with remarks "<remarks>" using v2 endpoint
    Then receives "InvalidArgument" status code
    
    Examples:
      | signed-in user | invoice-status | payment-method | remarks |
      | school admin   | PAID           | CASH           | any     |
      | school admin   | VOID           | CASH           | any     |
      | hq staff       | DRAFT          | BANK_TRANSFER  |         |
      | hq staff       | FAILED         | BANK_TRANSFER  | any     |
  
  Scenario Outline: Admin failed to approve a payment with a negative invoice total
    Given "<signed-in user>" logins to backoffice app
    And there is an existing invoice with "ISSUED" status
    And this invoice has "-500" total amount
    And there is "PENDING" payment history with "<payment-method>" payment method
    And admin sets the approve payment form with "<payment-date>" payment date
    When admin submits the approve payment form with remarks "<remarks>" using v2 endpoint
    Then receives "InvalidArgument" status code
    
    Examples:
      | signed-in user | payment-method | remarks |
      | school admin   | CASH           | any     |
      | hq staff       | BANK_TRANSFER  |         |

  Scenario Outline: Admin failed to approve a non-pending payment status
    Given "<signed-in user>" logins to backoffice app
    And there is an existing invoice with "ISSUED" status
    And there is "<payment-status>" payment history with "<payment-method>" payment method
    And admin sets the approve payment form with "<payment-date>" payment date
    When admin submits the approve payment form with remarks "<remarks>" using v2 endpoint
    Then receives "InvalidArgument" status code
    
    Examples:
      | signed-in user | payment-method | remarks | payment-status |
      | school admin   | CASH           | any     | SUCCESSFUL     |
      | hq staff       | BANK_TRANSFER  |         | FAILED         |
