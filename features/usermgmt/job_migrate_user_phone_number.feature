Feature: Run job to migrate user phone number from users table into user_phone_number table

  Scenario: Migrate phone number of user into phone number of "<userType>"
    Given some random "<userType>" with phone number
    When The system runs a job to migrate the phone number of "<userType>" into user_phone_number's table
    Then The phone number of "<userType>" was successfully migrated to the user_phone_number table

    Examples:
      | userType | 
      | staff    |
      | student  |
      | parent   |
