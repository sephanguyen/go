#1/bin/bash

CURRENT_DATE=$(date +"%s")
MESSAGE=""


# Loop over kubectl get certificates output and check for issues
# Append to MESSAGE any problems
while read x; do
    NOT_AFTER=$(date -d $(echo $x | cut -d\  -f3) +"%s")
    READY=$(echo $x | cut -d\  -f4)
    echo "$x"
    if [[ "$READY" != "Ready" ]]; then
        MESSAGE="NOT READY: ${x}\n${MESSAGE}"
    elif [[ $CURRENT_DATE -ge $((NOT_AFTER - 5 * 3600 * 24)) ]]; then
        MESSAGE="EXPIRES SOON: $x\n${MESSAGE}"
    fi
done < <(kubectl get certificates.cert-manager.io \
    -o=custom-columns='NAMESPACE:.metadata.namespace,NAME:.metadata.name,NOT_AFTER:status.notAfter,READY:status.conditions[0].type' \
    --sort-by=status.notAfter \
    --all-namespaces --no-headers)

# If message not empty
if [[ -n "$MESSAGE" ]]; then
    echo 'SLACK_MESSAGE<<EOF'
    echo -e "$MESSAGE"
    echo 'EOF'

    # If github_env is not empty
    if [[ -n "$GITHUB_ENV" ]]; then
        # Set Slack_Message for next action
        echo 'SLACK_MESSAGE<<EOF' >> $GITHUB_ENV
        echo -e "$MESSAGE" >> $GITHUB_ENV
        echo 'EOF' >> $GITHUB_ENV
    fi

    exit 1
fi