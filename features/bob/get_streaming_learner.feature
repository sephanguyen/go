Feature: get streaming learners who publishing an uploading stream from client
    Background: 
        Given a valid lesson in database
             And some valid learners in database
    
    Scenario: Return the currently uploading learner ids for a given lesson id
        Given "4" learner prepared publish in the lesson
        When get streaming learners 
        Then return "4" learner ids, who are currently uploading in the lesson
    
    Scenario: Students publish and unpublish uploading stream in a lesson as the same time
        Given a lesson with arbitrary number of student publishing
        When students publish and unpublish as the same time
        Then the number of publishing students must be record correctly