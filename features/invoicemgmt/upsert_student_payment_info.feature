@blocker
Feature: Upsert student payment info
  As an HQ manager or admin
  I am able to upsert payment info of a student

  Scenario Outline: Create a new billing information for a student doesn't have billing info yet
    Given an existing student with student payment "no billing info yet" info
    When "<signed-in user>" create a billing information that "<type of info>" for existing student
    Then "<result>" billing information
    And receives "<responded code>" status code

    Examples:
      | signed-in user  | type of info                                      | result               | responded code     |
      | unauthenticated | missing student id                                | failed to create     | Unauthenticated    |
      | school admin    | missing student id                                | failed to create     | InvalidArgument    |
      | unauthenticated | missing both billing information and bank account | failed to create     | Unauthenticated    |
      | school admin    | missing both billing information and bank account | failed to create     | InvalidArgument    |
      | unauthenticated | missing billing information                       | failed to create     | Unauthenticated    |
      | student         | missing billing information                       | failed to create     | PermissionDenied   |
      | parent          | missing billing information                       | failed to create     | PermissionDenied   |
      | teacher         | missing billing information                       | failed to create     | PermissionDenied   |
      | school admin    | missing billing information                       | failed to create     | InvalidArgument    |
      | hq staff        | missing billing information                       | failed to create     | InvalidArgument    |
      | centre manager  | missing billing information                       | failed to create     | InvalidArgument    |
      | centre staff    | missing billing information                       | failed to create     | InvalidArgument    |
      | school admin    | missing payer name                                | failed to create     | InvalidArgument    |
      | hq staff        | missing payer name                                | failed to create     | InvalidArgument    |
      | school admin    | missing billing address                           | failed to create     | InvalidArgument    |
      | centre manager  | missing billing address                           | failed to create     | InvalidArgument    |
      | school admin    | missing postal code                               | failed to create     | InvalidArgument    |
      | centre staff    | missing postal code                               | failed to create     | InvalidArgument    |
      | school admin    | missing prefecture code                           | failed to create     | InvalidArgument    |
      | school admin    | has non-exist prefecture code                     | failed to create     | FailedPrecondition |
      | school admin    | missing city                                      | failed to create     | InvalidArgument    |
      | school admin    | is valid                                          | successfully created | OK                 |
      | hq staff        | is valid                                          | successfully created | OK                 |
      | centre manager  | is valid                                          | successfully created | OK                 |
      | centre staff    | is valid                                          | successfully created | OK                 |
      | school admin    | missing payer phone number                        | successfully created | OK                 |
      | hq staff        | missing payer phone number                        | successfully created | OK                 |
      | school admin    | missing street 2                                  | successfully created | OK                 |
      | centre manager  | missing street 2                                  | successfully created | OK                 |


  Scenario Outline: Create more billing information for a student already have billing info
    Given an existing student with student payment "billing address" info
    When "<signed-in user>" create a billing information that "<type of info>" for existing student
    Then "<result>" billing information
    And receives "<responded code>" status code

    Examples:
      | signed-in user  | type of info                                      | result           | responded code     |
      | unauthenticated | missing student id                                | failed to create | Unauthenticated    |
      | school admin    | missing student id                                | failed to create | InvalidArgument    |
      | unauthenticated | missing both billing information and bank account | failed to create | Unauthenticated    |
      | school admin    | missing both billing information and bank account | failed to create | InvalidArgument    |
      | unauthenticated | is valid                                          | failed to create | Unauthenticated    |
      | student         | is valid                                          | failed to create | PermissionDenied   |
      | parent          | is valid                                          | failed to create | PermissionDenied   |
      | teacher         | is valid                                          | failed to create | PermissionDenied   |
      | school admin    | is valid                                          | failed to create | FailedPrecondition |
      | hq staff        | is valid                                          | failed to create | FailedPrecondition |
      | centre manager  | is valid                                          | failed to create | FailedPrecondition |
      | centre staff    | is valid                                          | failed to create | FailedPrecondition |
      | school admin    | missing payer name                                | failed to create | InvalidArgument    |
      | hq staff        | missing payer name                                | failed to create | InvalidArgument    |
      | school admin    | missing postal code                               | failed to create | FailedPrecondition |
      | centre manager  | missing postal code                               | failed to create | FailedPrecondition |
      | school admin    | missing payer phone number                        | failed to create | FailedPrecondition |
      | centre staff    | missing payer phone number                        | failed to create | FailedPrecondition |
      | school admin    | missing street 2                                  | failed to create | FailedPrecondition |


  Scenario Outline: Update billing information for a student already have billing info
    Given an existing student with student payment "billing address" info
    When "<signed-in user>" update with new billing information that "<type of info>" for existing student
    Then "<result>" billing information
    And receives "<responded code>" status code

    Examples:
      | signed-in user  | type of info                                      | result               | responded code     |
      | unauthenticated | missing student id                                | failed to update     | Unauthenticated    |
      | school admin    | missing student id                                | failed to update     | InvalidArgument    |
      | unauthenticated | missing both billing information and bank account | failed to update     | Unauthenticated    |
      | school admin    | missing both billing information and bank account | failed to update     | InvalidArgument    |
      | unauthenticated | missing billing information                       | failed to update     | Unauthenticated    |
      | student         | missing billing information                       | failed to update     | PermissionDenied   |
      | parent          | missing billing information                       | failed to update     | PermissionDenied   |
      | teacher         | missing billing information                       | failed to update     | PermissionDenied   |
      | school admin    | missing billing information                       | failed to update     | InvalidArgument    |
      | hq staff        | missing billing information                       | failed to update     | InvalidArgument    |
      | centre manager  | missing billing information                       | failed to update     | InvalidArgument    |
      | centre staff    | missing billing information                       | failed to update     | InvalidArgument    |
      | school admin    | missing payer name                                | failed to update     | InvalidArgument    |
      | hq staff        | missing payer name                                | failed to update     | InvalidArgument    |
      | school admin    | missing billing address                           | failed to update     | InvalidArgument    |
      | centre manager  | missing billing address                           | failed to update     | InvalidArgument    |
      | school admin    | missing postal code                               | failed to update     | InvalidArgument    |
      | centre staff    | missing postal code                               | failed to update     | InvalidArgument    |
      | school admin    | missing prefecture code                           | failed to update     | InvalidArgument    |
      | school admin    | has non-exist prefecture code                     | failed to update     | FailedPrecondition |
      | school admin    | missing city                                      | failed to update     | InvalidArgument    |
      | school admin    | has non-exist payment detail id                   | failed to update     | FailedPrecondition |
      | hq staff        | has non-exist payment detail id                   | failed to update     | FailedPrecondition |
      | centre manager  | has non-exist payment detail id                   | failed to update     | FailedPrecondition |
      | centre staff    | has non-exist payment detail id                   | failed to update     | FailedPrecondition |
      | school admin    | has non-exist billing address id                  | failed to update     | FailedPrecondition |
      | hq staff        | has non-exist billing address id                  | failed to update     | FailedPrecondition |
      | centre manager  | has non-exist billing address id                  | failed to update     | FailedPrecondition |
      | centre staff    | has non-exist billing address id                  | failed to update     | FailedPrecondition |
      | school admin    | is valid                                          | successfully updated | OK                 |
      | hq staff        | is valid                                          | successfully updated | OK                 |
      | centre manager  | is valid                                          | successfully updated | OK                 |
      | centre staff    | is valid                                          | successfully updated | OK                 |
      | school admin    | missing payer phone number                        | successfully updated | OK                 |
      | hq staff        | missing payer phone number                        | successfully updated | OK                 |
      | school admin    | missing street 2                                  | successfully updated | OK                 |
      | centre manager  | missing street 2                                  | successfully updated | OK                 |


  Scenario Outline: Create a new bank account for a student
    Given an existing student with student payment "<billing info>" info
    When "<signed-in user>" create a bank account that "<type of info>" for existing student
    Then "<result>" bank account
    And receives "<responded code>" status code

    Examples:
      | billing info        | signed-in user  | type of info                                                          | result               | responded code     |
      | no billing info yet | unauthenticated | missing student id                                                    | failed to create     | Unauthenticated    |
      | no billing info yet | school admin    | missing student id                                                    | failed to create     | InvalidArgument    |
      | no billing info yet | student         | missing both billing information and bank account                     | failed to create     | PermissionDenied   |
      | no billing info yet | school admin    | missing both billing information and bank account                     | failed to create     | InvalidArgument    |
      | no billing info yet | parent          | missing bank id and has unverified status                             | failed to create     | PermissionDenied   |
      | no billing info yet | school admin    | missing bank id and has unverified status                             | failed to create     | FailedPrecondition |
      | no billing info yet | teacher         | has non-exist bank id and unverified status                           | failed to create     | PermissionDenied   |
      | no billing info yet | school admin    | has non-exist bank id and unverified status                           | failed to create     | FailedPrecondition |
      | no billing info yet | hq staff        | missing bank branch id and has unverified status                      | failed to create     | FailedPrecondition |
      | no billing info yet | school admin    | missing bank branch id and has unverified status                      | failed to create     | FailedPrecondition |
      | no billing info yet | centre manager  | has non-exist bank branch id and unverified status                    | failed to create     | FailedPrecondition |
      | no billing info yet | school admin    | has non-exist bank branch id and unverified status                    | failed to create     | FailedPrecondition |
      | no billing info yet | centre staff    | missing bank account holder and has unverified status                 | failed to create     | FailedPrecondition |
      | no billing info yet | school admin    | missing bank account holder and has unverified status                 | failed to create     | FailedPrecondition |
      | no billing info yet | school admin    | has invalid alphabet in bank account holder and has unverified status | failed to create     | FailedPrecondition |
      | no billing info yet | school admin    | has invalid kana in bank account holder and has unverified status     | failed to create     | FailedPrecondition |
      | no billing info yet | school admin    | has invalid symbol in bank account holder and has unverified status   | failed to create     | FailedPrecondition |
      | no billing info yet | school admin    | missing bank account number and has unverified status                 | failed to create     | FailedPrecondition |
      | no billing info yet | school admin    | missing bank account type and has unverified status                   | failed to create     | FailedPrecondition |
      | no billing info yet | school admin    | is valid and has unverified status                                    | failed to create     | FailedPrecondition |
      | no billing info yet | school admin    | missing bank id and has verified status                               | failed to create     | InvalidArgument    |
      | no billing info yet | school admin    | has non-exist bank id and verified status                             | failed to create     | FailedPrecondition |
      | no billing info yet | school admin    | missing bank branch id and has verified status                        | failed to create     | InvalidArgument    |
      | no billing info yet | school admin    | has non-exist bank branch id and verified status                      | failed to create     | FailedPrecondition |
      | no billing info yet | school admin    | missing bank account holder and has verified status                   | failed to create     | InvalidArgument    |
      | no billing info yet | school admin    | missing bank account number and has verified status                   | failed to create     | InvalidArgument    |
      | no billing info yet | school admin    | missing bank account type and has verified status                     | failed to create     | InvalidArgument    |
      | no billing info yet | school admin    | is valid and has verified status                                      | failed to create     | FailedPrecondition |
      | billing address     | unauthenticated | missing student id                                                    | failed to create     | Unauthenticated    |
      | billing address     | school admin    | missing student id                                                    | failed to create     | InvalidArgument    |
      | billing address     | student         | missing both billing information and bank account                     | failed to create     | PermissionDenied   |
      | billing address     | school admin    | missing both billing information and bank account                     | failed to create     | InvalidArgument    |
      | billing address     | parent          | missing bank id and has unverified status                             | failed to create     | PermissionDenied   |
      | billing address     | school admin    | missing bank id and has unverified status                             | successfully created | OK                 |
      | billing address     | teacher         | has non-exist bank id and unverified status                           | failed to create     | PermissionDenied   |
      | billing address     | school admin    | has non-exist bank id and unverified status                           | successfully created | OK                 |
      | billing address     | hq staff        | missing bank branch id and has unverified status                      | successfully created | OK                 |
      | billing address     | school admin    | missing bank branch id and has unverified status                      | successfully created | OK                 |
      | billing address     | centre manager  | has non-exist bank branch id and unverified status                    | successfully created | OK                 |
      | billing address     | school admin    | has non-exist bank branch id and unverified status                    | successfully created | OK                 |
      | billing address     | centre staff    | missing bank account holder and has unverified status                 | successfully created | OK                 |
      | billing address     | school admin    | missing bank account holder and has unverified status                 | successfully created | OK                 |
      | billing address     | school admin    | has invalid alphabet in bank account holder and has unverified status | successfully created | OK                 |
      | billing address     | school admin    | has invalid kana in bank account holder and has unverified status     | successfully created | OK                 |
      | billing address     | school admin    | has invalid symbol in bank account holder and has unverified status   | successfully created | OK                 |
      | billing address     | hq staff        | missing bank account number and has unverified status                 | successfully created | OK                 |
      | billing address     | centre manager  | missing bank account type and has unverified status                   | successfully created | OK                 |
      | billing address     | centre staff    | is valid and has unverified status                                    | successfully created | OK                 |
      | billing address     | school admin    | missing bank id and has verified status                               | failed to create     | InvalidArgument    |
      | billing address     | hq staff        | has non-exist bank id and verified status                             | failed to create     | FailedPrecondition |
      | billing address     | centre manager  | missing bank branch id and has verified status                        | failed to create     | InvalidArgument    |
      | billing address     | centre staff    | has non-exist bank branch id and verified status                      | failed to create     | FailedPrecondition |
      | billing address     | school admin    | missing bank account holder and has verified status                   | failed to create     | InvalidArgument    |
      | billing address     | school admin    | has invalid alphabet in bank account holder and has verified status   | failed to create     | InvalidArgument    |
      | billing address     | school admin    | has invalid kana in bank account holder and has verified status       | failed to create     | InvalidArgument    |
      | billing address     | school admin    | has invalid symbol in bank account holder and has verified status     | failed to create     | InvalidArgument    |
      | billing address     | hq staff        | missing bank account number and has verified status                   | failed to create     | InvalidArgument    |
      | billing address     | centre manager  | missing bank account type and has verified status                     | failed to create     | InvalidArgument    |
      | billing address     | centre staff    | is valid and has verified status                                      | successfully created | OK                 |


  Scenario Outline: Update a existing bank account of a student
    Given an existing student with student payment "billing address and bank account" info
    When "<signed-in user>" update with new bank account that "<type of info>" for existing student
    Then "<result>" bank account
    And receives "<responded code>" status code

    Examples:
      | signed-in user  | type of info                                                          | result               | responded code     |
      | unauthenticated | missing student id                                                    | failed to update     | Unauthenticated    |
      | school admin    | missing student id                                                    | failed to update     | InvalidArgument    |
      | student         | missing both billing information and bank account                     | failed to update     | PermissionDenied   |
      | school admin    | missing both billing information and bank account                     | failed to update     | InvalidArgument    |
      | parent          | missing bank id and has unverified status                             | failed to update     | PermissionDenied   |
      | school admin    | missing bank id and has unverified status                             | successfully updated | OK                 |
      | teacher         | has non-exist bank id and unverified status                           | failed to update     | PermissionDenied   |
      | school admin    | has non-exist bank id and unverified status                           | successfully updated | OK                 |
      | hq staff        | missing bank branch id and has unverified status                      | successfully updated | OK                 |
      | centre manager  | has non-exist bank branch id and unverified status                    | successfully updated | OK                 |
      | centre staff    | missing bank account holder and has unverified status                 | successfully updated | OK                 |
      | centre staff    | has invalid alphabet in bank account holder and has unverified status | successfully updated | OK                 |
      | centre staff    | has invalid kana in bank account holder and has unverified status     | successfully updated | OK                 |
      | centre staff    | has invalid symbol in bank account holder and has unverified status   | successfully updated | OK                 |
      | school admin    | missing bank account number and has unverified status                 | successfully updated | OK                 |
      | hq staff        | missing bank account type and has unverified status                   | successfully updated | OK                 |
      | centre manager  | is valid and has unverified status                                    | successfully updated | OK                 |
      | centre staff    | missing bank id and has verified status                               | failed to update     | InvalidArgument    |
      | school admin    | has non-exist bank id and verified status                             | failed to update     | FailedPrecondition |
      | hq staff        | missing bank branch id and has verified status                        | failed to update     | InvalidArgument    |
      | centre manager  | has non-exist bank branch id and verified status                      | failed to update     | FailedPrecondition |
      | centre staff    | missing bank account holder and has verified status                   | failed to update     | InvalidArgument    |
      | centre staff    | missing bank account holder and has verified status                   | failed to update     | InvalidArgument    |
      | centre staff    | has invalid alphabet in bank account holder and has verified status   | failed to update     | InvalidArgument    |
      | centre staff    | has invalid kana in bank account holder and has verified status       | failed to update     | InvalidArgument    |
      | school admin    | has invalid symbol in bank account holder and has verified status     | failed to update     | InvalidArgument    |
      | hq staff        | missing bank account type and has verified status                     | failed to update     | InvalidArgument    |
      | centre manager  | is valid and has verified status                                      | successfully updated | OK                 |

  Scenario Outline: Update the verified status of bank account of the student
    Given an existing student with student payment "billing address and bank account" info
    And this student bank account is verified
    And the student default payment method was set to "DIRECT_DEBIT"
    When "<signed-in user>" update with new bank account that "<type of info>" for existing student
    Then "successfully updated" bank account
    And receives "OK" status code
    And the student default payment method was set to "<payment-method>"

    Examples:
      | signed-in user | type of info                                                          | payment-method    |
      | school admin   | is valid and has verified status                                      | DIRECT_DEBIT      |
      | centre manager | is valid and has unverified status                                    | CONVENIENCE_STORE |
      | hq staff       | missing bank account number and has unverified status                 | CONVENIENCE_STORE |
      | centre staff   | missing bank account type and has unverified status                   | CONVENIENCE_STORE |
      | school admin   | has non-exist bank id and unverified status                           | CONVENIENCE_STORE |
      | hq staff       | missing bank branch id and has unverified status                      | CONVENIENCE_STORE |
      | centre manager | has non-exist bank branch id and unverified status                    | CONVENIENCE_STORE |
      | centre staff   | missing bank account holder and has unverified status                 | CONVENIENCE_STORE |
      | centre staff   | has invalid alphabet in bank account holder and has unverified status | CONVENIENCE_STORE |
      | centre staff   | has invalid kana in bank account holder and has unverified status     | CONVENIENCE_STORE |
      | centre staff   | has invalid symbol in bank account holder and has unverified status   | CONVENIENCE_STORE |
      | school admin   | missing bank account number and has unverified status                 | CONVENIENCE_STORE |
      | hq staff       | missing bank account type and has unverified status                   | CONVENIENCE_STORE |

  Scenario Outline: Update Billing Address and Bank Information with Action Log Created successfully
    Given an existing student with student payment "<payment-info>" info
    And "<signed-in user>" logins to backoffice app
    And request "<update-action>" updates on student payment info with "<update-info>" information
    When admin updates the student payment information
    Then receives "OK" status code
    And student payment information updated successfully with "<update-action>" student payment detail action log record

    Examples:
      | signed-in user | payment-info                     | update-info                                     | update-action                    |
      | school admin   | billing address                  | PostalCode-City-Street1-PrefectureCode          | UPDATED_BILLING_DETAILS          |
      | school admin   | billing address                  | PostalCode-City-Street2-PrefectureCode          | UPDATED_BILLING_DETAILS          |
      | hq staff       | billing address                  | PostalCode-City-PrefectureCode                  | UPDATED_BILLING_DETAILS          |
      | centre manager | billing address                  | Street1-City                                    | UPDATED_BILLING_DETAILS          |
      | centre staff   | billing address                  | PayerName-PayerPhone                            | UPDATED_BILLING_DETAILS          |
      | school admin   | billing address and bank account | BankId-BankBranchId                             | UPDATED_BANK_DETAILS             |
      | hq staff       | billing address and bank account | BankAccountHolder-BankAccountNumber-NotVerified | UPDATED_PAYMENT_METHOD           |
      | centre manager | billing address and bank account | BankAccountType-Verified                        | UPDATED_PAYMENT_METHOD           |
      | centre staff   | billing address and bank account | PostalCode-City-BankAccountHolder-BankId        | UPDATED_BILLING_AND_BANK_DETAILS |

  Scenario Outline: No Action Log Created for on Billing Address and Bank Information with no updates
    Given an existing student with student payment "<payment-info>" info
    And "<signed-in user>" logins to backoffice app
    When admin updates student payment "<update-action>" info with same information
    Then receives "OK" status code
    And no student payment detail action log recorded

    Examples:
      | signed-in user | payment-info                     | update-action                    |
      | school admin   | billing address                  | UPDATED_BILLING_DETAILS          |
      | centre manager | billing address and bank account | UPDATED_BANK_DETAILS             |
      | hq staff       | billing address and bank account | UPDATED_BILLING_AND_BANK_DETAILS |

  Scenario: Update billing info of student with existing verfied bank account
    Given an existing student with student payment "billing address and bank account" info
    And this student bank account is verified
    And this student billing address was removed
    And this student payment method was removed
    And request "UPDATED_BILLING_AND_BANK_DETAILS" updates on student payment info with "PostalCode-City-Street1-PrefectureCode-PayerName-Verified" information
    When admin updates the student payment information
    Then receives "OK" status code
    And the student default payment method was set to "DIRECT_DEBIT"