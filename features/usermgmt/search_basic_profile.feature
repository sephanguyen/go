Feature: Seach basic profile

    With input are a list user id and search_text(name)
    I want to retrieve all basic profile satisfy the input.

    Background: Prepare student
        Given prepare students data

    Scenario Outline: retrieve basic profile
        Given search basic profile request with student_ids and "<filter>"
        When  "<signed-in-user>" search basic profile
        Then returns a list basic profile correctly

        Examples:
            | signed-in-user | filter                                              |
            | school admin   | none                                                |
            | school admin   | paging                                              |
            | school admin   | location_ids                                        |
            | school admin   | search_text                                         |
            | teacher        | none                                                |
            | teacher        | paging                                              |
            | teacher        | location_ids                                        |
            | teacher        | search_text                                         |
            | student        | none                                                |
            | student        | paging                                              |
            | student        | location_ids                                        |
            | student        | search_text                                         |
            | parent         | none                                                |
            | parent         | paging                                              |
            | parent         | location_ids                                        |
            | parent         | search_text                                         |
            | parent         | search_text_phonetic_name                           |
            | parent         | search_text_combine_full_name_and_phonetic_name     |
            | parent         | search_text_only_first_name_or_first_name_phonetic  |

