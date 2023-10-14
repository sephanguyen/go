Feature: Run publish import user events

    Scenario: Publish import student user events
        Given generate import student event records
        When system run task to publish import user events
        Then downstream services consume the events
        And status of import user events get updated

    Scenario: Publish import parent user events
        Given generate import parent event records
        When system run task to publish import user events
        Then downstream services consume the events
        And status of import user events get updated