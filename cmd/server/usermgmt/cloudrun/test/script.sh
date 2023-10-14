#!/bin/bash
set -euo pipefail

# In production, consider printing commands as they are executed.
# This helps with debugging if things go wrong and you only
# have the logs.
#
# Add -x:
# `set -euox pipefail`

CLOUD_RUN_TASK_INDEX=${CLOUD_RUN_TASK_INDEX:=0}
CLOUD_RUN_TASK_ATTEMPT=${CLOUD_RUN_TASK_ATTEMPT:=0}

echo "Starting Task #${CLOUD_RUN_TASK_INDEX}, Attempt #${CLOUD_RUN_TASK_ATTEMPT}..."

# SLEEP_MS and FAIL_RATE should be a decimal
# numbers. parse and format the input using
# printf.
#
# printf validates the input since it
# quits on invalid input, as shown here:
#
#   $: printf '%.1f' "abc"
#   bash: printf: abc: invalid number
#
SLEEP_MS=$(printf '%.1f' "${SLEEP_MS:=0}")
FAIL_RATE=$(printf '%.1f' "${FAIL_RATE:=0}")

# Wait for a specific amount of time to simulate
# performing some work
#SLEEP_SEC=$(echo print\("${SLEEP_MS}"/1000\) | perl)
#sleep "$SLEEP_SEC" # sleep accepts seconds, not milliseconds

# Fail the task with a likelihood of $FAIL_RATE

# Bash does not do floating point arithmetic. Use perl
# to convert into integer and multiply by 100.
FAIL_RATE_INT=$(echo print\("int(${FAIL_RATE:=0}*100"\)\) | perl)

sshpass -v -p "${PASSWORD}" scp -v -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "${USERNAME}"@210.149.14.116:/home/"${USERNAME}"/send_test/N1-M1_users20220115.tsv /workspace
cat /workspace/N1-M1_users20220115.tsv
