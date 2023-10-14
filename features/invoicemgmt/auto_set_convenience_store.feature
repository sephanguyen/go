Feature: Auto set Convenience Store Payment Method

  Scenario: Student account is created with convenience store payment method successfully
    Given unleash feature flag is "enable" with feature name "BACKEND_Invoice_InvoiceManagement_AutoSetConvenienceStore"
    And invoicemgmt internal config "enable_auto_default_convenience_store" is "on"
    And there is an event create student request with user address info
    When yasuo send the create student event request
    Then student payment detail record is successfully created
    And student payer name successfully updated
    And student billing address record is successfully created
    And billing address is the same as user address
    And no student payment detail action log recorded

  @quarantined
  Scenario: Student account with initial payment and billing info is included in the create student event
    Given unleash feature flag is "enable" with feature name "BACKEND_Invoice_InvoiceManagement_AutoSetConvenienceStore"
    And invoicemgmt internal config "enable_auto_default_convenience_store" is "on"
    And there is an existing student with "existing" billing address and "existing" payment detail
    And there is an event create student request with user address info
    When yasuo send the create student event request
    Then student payment detail record is successfully created
    And student payer name successfully updated
    And student billing address record is successfully created
    And billing address is the same as user address
    And student payment information updated successfully with "UPDATED_BILLING_DETAILS" student payment detail action log record

  @quarantined
  Scenario Outline: no student convenience store payment method is set when feature flag is disable
    Given unleash feature flag is "disable" with feature name "BACKEND_Invoice_InvoiceManagement_AutoSetConvenienceStore"
    And invoicemgmt internal config "enable_auto_default_convenience_store" is "<internal-config-status>"
    And there is an event create student request with user address info
    When yasuo send the create student event request
    Then no student payment detail record created
    And no student billing address record created

    Examples:
      | internal-config-status |
      | on                     |
      | off                    |

  @quarantined
  Scenario: no student convenience store payment method is set when feature flag is enable and internal config is disable
    Given unleash feature flag is "enable" with feature name "BACKEND_Invoice_InvoiceManagement_AutoSetConvenienceStore"
    And invoicemgmt internal config "enable_auto_default_convenience_store" is "off"
    And there is an event create student request with user address info
    When yasuo send the create student event request
    Then no student payment detail record created
    And no student billing address record created
