#!/bin/bash
# create a list of triggered services based on the author's joined squad
# this is used for running a subset of integration tests

set -eu
GITHUB_OUTPUT=${GITHUB_OUTPUT:-/dev/stdout}
# test case: TEAMS='["func-backend","dev","squad-syllabus","squad-ddd","squad-ddd-be","squad-syllabus-be"]' ./get_triggered_svcs.bash 
function remove_brackets() {
  v="$1"
  v="${v//[/}"
  v="${v//]/}"
  v="${v//$2/ }"
  v="${v//\"/}"
  echo "$v"
}

team_list=${TEAMS:-}
team_list=$(remove_brackets "$team_list" ",")

if [ $# -ne 1 ]; then
  echo >&2 "ERROR: get_triggered_svcs.bash requires exactly 1 arguments"
  exit 1
fi

run_test_by_label=false
triggered_svcs=()

label_list=$(remove_brackets "$1" ",")
read -a label_list <<< "$label_list"
for label in "${label_list[@]}"; do
  if [[ "$label" == test:* ]]; then
    serviceString=${label#"test:"}
    triggered_svcs=$(remove_brackets "$serviceString" ";")
    run_test_by_label=true
    echo "has prefix test:"
    break
  fi
done

if [[ "$run_test_by_label" != "true" ]]; then
  read -a team_list <<< "$team_list"
  echo "Team list: ${team_list[*]}"
  for team in "${team_list[@]}"; do
    case $team in
      squad-communication)
        triggered_svcs+=("tom" "communication")
        ;;
      squad-adobo)
        triggered_svcs+=("entryexitmgmt" "invoicemgmt")
        ;;
      squad-lesson)
        triggered_svcs+=("bob" "fatima" "lessonmgmt" "mastermgmt" "enigma" "virtualclassroom" "yasuo" "calendar")
        ;;
      squad-payment)
        triggered_svcs+=("fatima" "payment" "mastermgmt")
        ;;
      squad-syllabus)
        triggered_svcs+=("eureka" "syllabus")
        ;;
      squad-user-management)
        triggered_svcs+=("bob" "fatima" "usermgmt" "yasuo")
        ;;
      squad-auth)
        triggered_svcs+=("bob" "fatima" "usermgmt" "yasuo")
        ;;    
      squad-conversationmgmt)
        triggered_svcs+=("tom" "communication" "bob" "eureka" "syllabus" "yasuo")
        ;;
      squad-architecture-be)
        triggered_svcs+=("bob" "fatima" "lessonmgmt" "mastermgmt" "enigma" "virtualclassroom" "yasuo" "accesscontrol")
        ;;
      squad-calendar)
        triggered_svcs+=("calendar" "lessonmgmt")
        ;;
      *)
        echo "Team $team not yet included and will be ignored"
        ;;
    esac
  done
fi


echo "Triggered service list: ${triggered_svcs[*]}"
echo "triggered_svcs=${triggered_svcs[*]}" >> $GITHUB_OUTPUT
