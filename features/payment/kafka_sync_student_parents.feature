Feature: Student Parent sync data from Kafka
    @quarantined
    Scenario: Sync data of student_parents table from bob to fatima
        When a record is inserted in student parent in bob
        Then the student parent must be recorded in payment
