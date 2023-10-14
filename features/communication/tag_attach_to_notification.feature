Feature: school admin want to create, edit notification with tags attached
    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "1" students
        And school admin create some tags named "tag1,tag2,tag3,tag4,tag5"
   
    @blocker
    Scenario: Admin use upsert notification to manage notification with attached tags
        Given a valid composed notification
        And notification is attached with tags "<attached_tag_names>" in database
        When user send upsert notification request to attach "<req_tag_names_attach>" tags to notification
        Then "<num_tag_result>" attached tags data and names are correctly stored database
        Examples:
            | attached_tag_names | req_tag_names_attach | num_tag_result |
            | tag2,tag3,tag5     | tag4,tag1            | 2              |
            | tag1,tag3,tag2     | tag1,tag2,tag3       | 3              |
            | NULL               | tag1,tag4,tag3       | 3              |
            | tag5,tag3          | NULL                 | 0              |
            | tag1,tag2,tag3     | tag2,tag2,tag3,tag4  | 3              |
            | tag5,tag4,tag3     | tag0,tag2,tag3,tag4  | -1             |

    Scenario: Admin use upsert notification to soft delete associations and re-assign the same tags
        Given a valid composed notification
        And notification is attached with tags "<attached_tag_names>" in database
        When user send upsert notification request to remove "<delete_tag_names>" tags from notification
        Then association with "<delete_tag_names>" tags are soft deleted in database
        When user send upsert notification request to attach "<reassign_tag_names>" tags to notification
        Then association with "<reassign_tag_names>" tags are reenabled in database
        Examples:
            | attached_tag_names  | delete_tag_names    | reassign_tag_names  |
            | tag1,tag2,tag3      | tag2,tag3           | tag3                |
            | tag3,tag2,tag5,tag4 | tag4,tag3,tag2      | tag5,tag4           |
            | tag3,tag2,tag5,tag4 | tag3,tag2,tag5,tag4 | tag3,tag2,tag5,tag4 |

    Scenario: Admin use upsert notification with tag and discard it
        Given a valid composed notification
        And notification is attached with tags "tag1,tag2,tag3" in database
        When current staff discards notification
        Then notification is discarded

    Scenario: Admin upsert notification with archived tags
        Given a valid composed notification
        And admin archived "<archived_tag_names>" tags
        And notification is attached with tags "<attached_tag_names>" in database
        Then returns "InvalidArgument" status code
        Examples:
            | attached_tag_names  | archived_tag_names  |
            | tag1,tag2,tag3      | tag2                |
            | tag3,tag2,tag5,tag4 | tag4,tag3,tag2      |
            | tag3,tag2,tag5,tag4 | tag3,tag2,tag5,tag4 |
