@quarantined
Feature: Teacher join all conversation
    Background: default manabie resource path
        Given resource path of school "Manabie" is applied

    Scenario: Migrate conversation locations for student with no locations
        Given create student with no locations
        And insert org location access path to student
        When migrate conversation locations
        Then conversation location is inserted