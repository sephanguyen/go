Feature: Upload file using presign url

    Scenario: Client call generate presign url to put object
        Given "staff granted role school admin" signin system
        And a file information to generate put object url
        When generate presign url to put object
        Then returns "OK" status code
        And return presign put object url 
        And the file can be uploaded using the returned url

    Scenario: Client call generate resumable upload url
        Given "staff granted role school admin" signin system
        And a file information to generate resumable upload url
        When generate resumable upload url
        Then returns "OK" status code
        And return resumable upload url