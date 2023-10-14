Feature: Multiple School Admin Interact with student entry exit records

    Scenario: School admin creates a new entry and exit record
        Given "school admin" logins CMS App with resource path from "organization -2147483648"
        And this school admin creates a new student entry and exit record
        When another "school admin" logins CMS App with resource path from "<organization>"
        Then this school admin "<result>" see the new student entry and exit record 
       
         Examples:
            | result  | organization              |
            | can     | organization -2147483648  |
            | cannot  | organization -2147483646  |
            | cannot  | organization -2147483645  |