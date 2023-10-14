#!/bin/bash

get_id() {
    MESSAGE="$1"
    CURRENT_BRANCH="$2"
    get_ticket_id() {
        regexp_message="(\b(LT|lt))-([0-9]{1,6})"
        [[ $MESSAGE =~ $regexp_message ]]
        TICKET_ID_MESSAGE="${BASH_REMATCH[0]^^}"
        if [ -z "$TICKET_ID_MESSAGE" ]; then
            [[ $CURRENT_BRANCH =~ $regexp_message ]]
            TICKET_ID="${BASH_REMATCH[0]^^}"
        fi
    }

    get_ticket_id

    if [ -n "$TICKET_ID" ]; then 
        sed -i.bak -e "1s/^/[$TICKET_ID] /" $3
    fi
}