@blocker
Feature: Retrieve Student Payment Method Record
    Payment service method to retrieve bulk student default payment method
    # Currently we only support Convenience Store and Direct Debit for Student's default payment method

  Scenario Outline: Retrieve students default payment method record successfully
    Given there are existing "<count>" students with "<payment-method>" default payment method
    And "<signed-in user>" logins to backoffice app
    When the RetrieveBulkStudentPaymentMethod endpoint is called for these students
    Then receives "OK" status code
    And payment methods for these students are retrieve successfully

    Examples:
      | signed-in user | payment-method                                   | count |
      | school admin   | CONVENIENCE_STORE-DIRECT_DEBIT                   | 2     |
      | hq staff       | DIRECT_DEBIT-DIRECT_DEBIT-DIRECT_DEBIT           | 3     |
      | centre staff   | CONVENIENCE_STORE-CONVENIENCE_STORE-DIRECT_DEBIT | 3     |
      | centre manager | CONVENIENCE_STORE-NO_DEFAULT_PAYMENT             | 2     |
      | centre manager | NO_DEFAULT_PAYMENT                               | 1     |
