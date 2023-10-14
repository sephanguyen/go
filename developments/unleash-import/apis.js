
const axios = require('axios');
const https = require('https');
const { exit } = require('process');

var org = process.env.ORG ?? 'manabie';
var env = process.env.ENV ?? 'local';
var adminToken = process.env.ADMIN_TOKEN ?? '2570e064cb6996e4258e9d6a8206300e4286b088a3fd3ee02450cc56'; // default is backend-admin token on local
var slackToken = process.env.SLACK_TOKEN;
const { WebClient } = require("@slack/web-api");

// WebClient instantiates a client that can call API methods
// When using Bolt, you can use either `app.client` or the `client` passed to listeners.
const client = new WebClient(slackToken);

const urls = {
    'local': {
        'manabie': 'https://admin.local-green.manabie.io:31600/unleash',
    },
    'stag': {
        'manabie': 'https://admin.staging.manabie.io/unleash',
        'jprep': 'https://admin.staging.jprep.manabie.io/unleash'
    },
    'uat': {
        'manabie': 'https://admin.uat.manabie.io/unleash',
        'jprep': 'https://admin.uat.jprep.manabie.io/unleash',
    },
    'dorp': {
        'tokyo': 'https://admin.prep.tokyo.manabie.io/unleash',
    },
    'prod': {
        'tokyo': 'https://admin.prod.tokyo.manabie.io/unleash',
        'jprep': 'https://admin.prod.jprep.manabie.io/unleash',
        // 'ga': 'https://admin.prod.ga.manabie.io/unleash',
        // 'aic': 'https://admin.prod.aic.manabie.io/unleash',
        // 'renseikai': 'https://admin.prod.renseikai.manabie.io/unleash',
        // 'synersia': 'https://admin.synersia.manabie.io/unleash'
    }
}

const roles = {
    1: 'Admin',
    2: 'Editor',
    3: 'Viewer',
};

const url = urls[env][org];

console.log(`======= ${org} - ${env} ========`)
console.log(`======= url: ${url} ========`)

axios.interceptors.request.use(function (config) {
    config.headers.Authorization = adminToken;
    config.headers['Content-Type'] = 'application/json';
    config.httpsAgent = new https.Agent({
        rejectUnauthorized: false,
    });
    return config;
});

axios.interceptors.response.use((response) => {
    return response;
}, (error) => {
    return Promise.reject(error.message);
})


async function removeFeatures(features) {
    for (let index = 0; index < features.length; index++) {
        const feature = features[index];
        info(`Deleting: ${feature}`)
        await axios.delete(`${url}/api/admin/features/${feature}`).then(function (response) {
            // console.log(JSON.stringify(response.data));
            info(`Archived: ${feature}`)
        }).catch(function (e) {
            error(e);
        });
        await axios.delete(`${url}/api/admin/archive/${feature}`).then(function (response) {
            // info(JSON.stringify(response.data));
            info(`Deleted: ${feature}`)
        }).catch(function (e) {
            error(e);
        });
    }
}

async function updateFeatures(features) {
    for (let index = 0; index < features.length; index++) {
        const feature = features[index];
        info(`Updating: ${feature['name']}`)
        await axios.put(`${url}/api/admin/features/${feature['name']}`, JSON.stringify(feature)).then(function (response) {
            // console.log(JSON.stringify(response.data));
            info(`Updated: ${feature['name']}`)
        }).catch(function (e) {
            error(e);
        });
    }
}


async function updateFeatureTags(featureTags) {
    Object.keys(featureTags).forEach(async (featureName) => {
        await axios.post(`${url}/api/admin/features/${featureName}/tags`, JSON.stringify(featureTags[featureName]))
        .then(function (response) {
            info(`Updated tags for ${featureName}`)
        })
        .catch(function (e) {
            error(e);
        });
    });
}

async function createFeatures(features) {
    for (let index = 0; index < features.length; index++) {
        const feature = features[index];
        info(`Creating: ${feature['name']}`)
        await axios.post(`${url}/api/admin/features`, JSON.stringify(feature)).then(function (response) {
            // console.log(JSON.stringify(response.data));
            info(`Created: ${feature['name']}`)
        }).catch(function (e) {
            error(e);
        });
    }
}

function info(message) {
    console.info(`INFO(${org}, ${env}) << ${message}`);
}

function error(message) {
    console.error(`ERROR(${org}, ${env}) << ${message}`);
    exit(1)
}

async function getFeatures() {
    info(`Getting remote features`)
    var response = await axios.get(`${url}/api/admin/features`).catch(function (e) {
        error(e);
    });

    return response['data'];
}

async function getUsers() {
    info(`Getting remote users`)
    var response = await axios.get(`${url}/api/admin/user-admin`).catch(function (e) {
        error(e);
    });
    return response['data'];
}

async function removeUsers(users) {
    for (let index = 0; index < users.length; index++) {
        const user = users[index];
        info(`Deleting: ${user['name']}`)
        await axios.delete(`${url}/api/admin/user-admin/${user['id']}`).then(function (response) {
            // info(JSON.stringify(response.data));
            info(`Deleted: ${user['name']}`)
        }).catch(function (e) {
            error(e);
        });
    }
}

async function updateUsers(users) {
    for (let index = 0; index < users.length; index++) {
        const user = users[index];
        info(`Updating: ${user['name']}`)
        await axios.put(`${url}/api/admin/user-admin/${user['id']}`, JSON.stringify(user)).then(function (response) {
            // console.log(JSON.stringify(response.data));
            info(`Updated: ${user['name']}`)
        }).catch(function (e) {
            error(e);
        });
    }
}

async function createUsers(users) {
    for (let index = 0; index < users.length; index++) {
        const user = users[index];
        info(`Creating: ${user['name']}`)
        info(JSON.stringify(user))
        await axios.post(`${url}/api/admin/user-admin`, JSON.stringify(user)).then(async function (response) {
            info(`Created: ${user['name']}`);
            console.log(response);
            await finderUserIdByEmailAndSendMessage(response['data']['email'], response['data']['inviteLink'], roles[user['rootRole']]);
        }).catch(function (e) {
            info(`User email: ${user.email}`)
            error(e);
        });
    }
}

async function finderUserIdByEmailAndSendMessage(email, inviteLink, role) {
    try {
        info(`Sending message to: ${email}`)
        info(`Find userid by: ${email}`)
        const result = await client.users.lookupByEmail({
            'email': email,
        });
        info(`Found userid: ${result.user.id}`)
        await client.chat.postMessage({
            channel: result.user.id,
            mrkdwn: true,
            text: `
Hi ${result.user.real_name}, I just invited you as ${role} in UNLEASH ${org.toUpperCase()} ${env.toUpperCase()}, Please click the invite link below to reset your password and log in.
:point_right: ${inviteLink}
If you got error Invalid Token, you can use forget password to reset the password.
Thank you!`,
        });
        info(`Sent message to: ${email}`)
    }
    catch (e) {
        error(e);
    }
}



module.exports = {
    updateFeatures,
    updateFeatureTags,
    removeFeatures,
    createFeatures,
    getFeatures,
    getUsers,
    removeUsers,
    createUsers,
    updateUsers,
    finderUserIdByEmailAndSendMessage,
    env,
    org,
    info,
    error
};
