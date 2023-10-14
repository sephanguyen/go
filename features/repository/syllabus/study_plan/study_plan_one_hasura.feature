Feature: Get Study Plan

    Scenario: Get Study Plan
        Given a valid StudyPlan in database
        When a user call StudyPlanOne
        Then our System will return all information about StudyPlan 
