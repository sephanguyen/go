Feature: Run job to migrate full name into last name

  Scenario: Migrate full name into first name of students
    Given some random student with full name only
    When system run job to migrate student full name into last name
    Then full name migrated to last name successfully