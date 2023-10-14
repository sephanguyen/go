const hclParser = require("js-hcl-parser")
const { globSync } = require('glob')
const fs = require('fs');

const {
    org,
    env,
    getUsers,
    updateUsers,
    removeUsers,
    createUsers,
} = require('./apis.js');

const ACCESS_CONTROL_DIR = 'deployments/terraform/live/workspace/access-control/*_members.hcl'

async function hclToJson(hclFilePath) {
    const data = fs.readFileSync(hclFilePath, "utf8")
    const jsonText = hclParser.parse(data)
    return JSON.parse(jsonText)
}

async function getUsersLocal() {
    const files = globSync(ACCESS_CONTROL_DIR)
    const users = [];
    for (const file of files) {
        const jsonData = await hclToJson(file);
        var members = jsonData.locals[0].members
        const devs = members.filter(member => !member.functions.find(func => func.name === 'pdm'));
        for (const member of devs) {
            var role = 2;
            if (env == 'prod') {
                role = 3; // Default value is viewer on PROD
                for (const squad of member.squads) {
                    if (squad.role == 'manager') {
                        role = 2; // Editor
                        break;
                    }
                }
                for (const func of member.functions) {
                    if (func.name == 'techlead') {
                        role = 2; // Editor
                        break;
                    }
                }
            }
            users.push({
                'email': member['email'],
                'name': member['name'],
                'rootRole': role,
            });
        }
    }
    return users;
}


function convertToMap(users) {
    return users.reduce(function (map, obj) {
        map[obj.email] = obj;
        return map;
    }, {});
}

function compareValues(usersMap, usersLocalMap) {
    var usersCreated = [];
    var usersUpdated = [];
    var usersDeleted = [];

    for (const [key, value] of Object.entries(usersMap)) {
        if (value['username'] == 'admin') {
            continue;
        }
        var userLocal = usersLocalMap[key];

        if (userLocal === undefined) {
            usersDeleted.push(value);
            continue;
        }
        userLocal['id'] = value['id'];

        if (isDiff(value, userLocal)) {
            usersUpdated.push(userLocal);
        }
    }

    for (const [key, value] of Object.entries(usersLocalMap)) {
        var feature = usersMap[key];

        if (feature === undefined) {
            usersCreated.push(value);
        }
    }

    return {
        'users_update': usersUpdated,
        'users_create': usersCreated,
        'users_delete': usersDeleted,
    }
}

function isDiff(user, userLocal) {
    if (userLocal['name'] != user['name']
        || userLocal['rootRole'] != user['rootRole']) {
        return true;
    }

    return false;
}


async function importAccounts() {
    var users = await getUsers();
    var usersLocal = await getUsersLocal();
    var usersLocalMap = convertToMap(usersLocal);
    var usersMap = convertToMap(users['users']);
    var data = compareValues(usersMap, usersLocalMap);
    console.info(JSON.stringify(data))
    console.info("======== Start call unleash API ========")
    await updateUsers(data['users_update']);
    await createUsers(data['users_create']);
    await removeUsers(data['users_delete']);
    console.info("======== End call unleash API ========")
    fs.writeFileSync(`./${org}_${env}_diff.md`,
        `#### unleash ${org} ${env}: accounts
\`\`\`json
${JSON.stringify(data)}
\`\`\`
    `
    );
}

module.exports = {
    importAccounts
}
