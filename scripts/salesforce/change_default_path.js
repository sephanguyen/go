const fs = require('fs');
const sfdxProjectJson = fs.readFileSync('salesforce/original-sfdx-project.json', 'utf8');
const namePath = process.argv[2];

const namePathSplit = namePath.split('/');

const packageName = namePathSplit[namePathSplit.length - 1];

const fileJSON = JSON.parse(sfdxProjectJson)


const allPaths = fileJSON.packageDirectories.map((item) => {
    return item.path;
})

if (!allPaths.includes(packageName)) {
  throw new Error(`Path ${packageName} not found in sfdx-project.json`);
}

const changedDefaultPath = fileJSON.packageDirectories.map((item) => {
    if (item.path === packageName) {
        item.default = true;
    }else{
        item.default = false;
    }
    return item;

});
fileJSON.packageDirectories = changedDefaultPath;

const updatedJSON = JSON.stringify(fileJSON, null, 2);
fs.writeFileSync('salesforce/sfdx-project.json', updatedJSON, 'utf8');