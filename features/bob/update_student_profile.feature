@quarantined
Feature: Update Student Profile
    In order to update my own profile
    As a student
    I need to perform profile update

    Scenario: unauthenticated user updates profile
        Given an invalid authentication token
            And a valid updates profile request
        When user updates profile
        Then Bob must not update student's profile
            And returns "Unauthenticated" status code

    Scenario: registered student updates profile
        Given a signed in student
            And a valid updates profile request
            And his owned student UUID
        When user updates profile
        Then Bob must records student's profile update
            And returns "OK" status code
            And Bob must publish event to user_device_token channel

    Scenario: registered student updates school using existed school
        Given "staff granted role school admin" signin system
            And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
            And a school name "S2", country "COUNTRY_VN", city "Hồ Chí Minh", district "3"
            And a school name "S3", country "COUNTRY_VN", city "Hồ Chí Minh", district "Bình Thạnh"
            And admin inserts schools

            And a signed in student
            And a valid updates profile request
            And student selects school in country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
            And his owned student UUID
        When user updates profile
        Then Bob must records student's profile update
            And returns "OK" status code
            And Bob must publish event to user_device_token channel

    Scenario: registered student updates school by input new school
        Given "staff granted role school admin" signin system
            And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
            And a school name "S2", country "COUNTRY_VN", city "Hồ Chí Minh", district "3"
            And a school name "S3", country "COUNTRY_VN", city "Hồ Chí Minh", district "Bình Thạnh"
            And admin inserts schools

            And a signed in student
            And a valid updates profile request
            And student inputs new school in country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
            And his owned student UUID
        When user updates profile
        Then Bob must records student's profile update
            And returns "OK" status code
            And Bob must publish event to user_device_token channel
    
    Scenario: non-registered student updates profile
        Given "staff granted role school admin" signin system
        And a valid updates profile request
        And his owned student UUID
        When user updates profile
        Then Bob must not update student's profile
        And returns "PermissionDenied" status code