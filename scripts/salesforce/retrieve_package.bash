#!/bin/bash
package_name=$1
org_alias=$2

if [ -z $package_name ]
  then
    echo "Please add the path to the package.xml as the first argument"
    exit 1
fi

if [ -z $org_alias ]
  then
    echo "Please add alias of org as the second argument"
    exit 1
fi

mv salesforce/sfdx-project.json salesforce/original-sfdx-project.json

node scripts/salesforce/change_default_path.js $package_name

if [ $? -ne 0 ]; then
  mv salesforce/original-sfdx-project.json salesforce/sfdx-project.json
  exit 1
fi

cd salesforce && sf project retrieve start --manifest $package_name/package.xml -o $org_alias

rm sfdx-project.json

mv original-sfdx-project.json sfdx-project.json