#!/bin/bash

set -u

. ./common.bash

test_status=success

# test check_message()
success_cases=(
    "LT-1 Ticket with small number"
    "LT-999999 Ticket with large number"
    "chore: LT-5678 refactor to decrease memory footprint"
    "[user_login][config] Fix typo making users could not login for LT-910111"
    "Merge branch 'develop' into feature/feature-branch"
    "Merge remote-tracking branch 'origin' into features/do-stuffs"
    "Merge branch 'feature/refactor-platform' of github.com:manabie-com/backend into feature/refactor-platform"
    "Merge pull request #2154 from manabie-com/fix/synersia-package-name"
    "Merge pull request #2159 from manabie-com/feature/LT-5495-add-chat-name-to-msg"
)
for message in "${success_cases[@]}"; do
    check_message "${message}"

    if [[ $? != 0 ]]; then
        >&2 echo "check_message: expected status 0 for \"${message}\", got 1"
        test_status=failed
    fi
done

failure_cases=(
    "This message has no ticket IDs"
    "chore: 5678 this message's ticket ID has an invalid format"
    "This message has too many ticket IDs LT-9101, LT-1121"
    "[LT-1234567] The ticket number is too large"
    "LT-123Missing space"
)
for message in "${failure_cases[@]}"; do
    check_message "${message}" > /dev/null 2>&1

    if [[ $? == 0 ]]; then
        >&2 echo "check_message: expected status 1 for \"${message}\", got 0"
        test_status=failed
    fi
done

# test check_pr_title()
success_cases=(
    "LT-1 Ticket with small number"
    "LT-999999 Ticket with large number"
    "chore: LT-5678 refactor to decrease memory footprint"
    "[user_login][config] Fix typo making users could not login for LT-910111"
)
for title in "${success_cases[@]}"; do
    check_pr_title "${title}"

    if [[ $? != 0 ]]; then
        >&2 echo "check_pr_title: expected status 0 for \"${title}\", got 1"
        test_status=failed
    fi
done

failure_cases=(
    "This message has no ticket IDs"
    "chore: 5678 this message's ticket ID has an invalid format"
    "This message has too many ticket IDs LT-9101, LT-1121"
    "[LT-1234567] The ticket number is too large"
    "LT-123Missing space"
    "Merge branch 'develop' into feature/feature-branch"
    "Merge remote-tracking branch 'origin' into features/do-stuffs"
    "Merge branch 'feature/refactor-platform' of github.com:manabie-com/backend into feature/refactor-platform"
)
for title in "${failure_cases[@]}"; do
    check_pr_title "${title}" > /dev/null 2>&1

    if [[ $? == 0 ]]; then
        >&2 echo "check_pr_title: expected status 1 for \"${title}\", got 0"
        test_status=failed
    fi
done

if [[ "${test_status}" == "success" ]]; then
    echo "All tests completed successfully"
fi
