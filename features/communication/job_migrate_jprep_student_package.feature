@quarantined @disabled
Feature: notification system sync Jprep student package data from Nats even of Yasuo
    Scenario: notificatiom systen create Jprep student package
        Given some valid student package in fatima database 
        When run migration for jprep student package
        And waiting for notification system sync data
        Then notification system must store student course data correctly
