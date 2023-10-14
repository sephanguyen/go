Feature: user gets system notifications

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations

    Scenario Outline: staff get system notification by filter
        Given some staffs with random roles and granted organization location of current organization
        And staff create system notification with "<num new>" new and "<num done>" done and "<num unenabled>" unenabled
        And waiting for kafka sync data
        When staff get system notifications with status "<status>" and lang "<lang>" and limit "<limit>" and offset "<offset>"
        Then staff check response "<num new>" new and "<num done>" done and "<status>" status and "<total count>" count
        Examples:
            | num new | num done | num unenabled | limit | offset | status | lang | total count |
            | 10      | 15       | 5             | 10    | 0      | new    | en   | 25          |
            | 5       | 10       | 5             | 5     | 0      | done   | jp   | 15          |
            | 3       | 7        | 5             | 5     | 5      | all    | en   | 10          |
