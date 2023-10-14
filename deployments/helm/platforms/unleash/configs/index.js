'use strict';

const unleash = require('unleash-server');
const fs = require('fs')
const yaml = require('js-yaml')

async function upsertAdminUser(app, config) {
    // We probably need transactional serialization here,
    // but it's extremely unlikely for this function to get called that often.
    const logger = app.config.getLogger('index.js')
    var adminUserID = await app.stores.userStore.hasUser({ username: "admin" })
    if (adminUserID === undefined) {
        logger.info(`Admin user not found, creating a new admin user`)
        const user = await app.services.userService.createUser({
            username: "admin",
            rootRole: "Admin",
        }).catch(err => { throw new Error(`Failed to create admin user: ${err}`) })
        adminUserID = user.id
    }

    logger.info(`Updating admin user password hash`)
    const adminPasswordHash = config.admin_password
    await app.stores.userStore
        .setPasswordHash(adminUserID, adminPasswordHash)
        .catch(err => { throw new Error(`Failed to set password hash for admin user: ${err}`) })
}

async function upsertClientAPIToken(app) {
    // We probably need transactional serialization here, too
    const logger = app.config.getLogger('index.js')
    logger.info(`Upserting client API token`)
    const apiToken = process.env.PROXY_API_TOKEN
    const apiTokenExists = await app.stores.apiTokenStore.exists(apiToken)
    if (!apiTokenExists) {
        logger.info(`Client API token not exist, creating using token: ${apiToken}`)
        await app.stores.apiTokenStore.insert({
            secret: apiToken,
            username: 'proxy-client',
            type: 'client',
        }).catch(err => { throw new Error(`Failed to insert API token: ${err}`) })
        await app.services.apiTokenService.fetchActiveTokens();
    }
}

async function upsertAdminAPIToken(app) {
    // We probably need transactional serialization here, too
    const logger = app.config.getLogger('index.js')
    logger.info(`Upserting admin API token`)
    const apiToken = config.adminAPIToken;
    const apiTokenExists = await app.stores.apiTokenStore.exists(apiToken)
    if (!apiTokenExists) {
        logger.info(`Admin API token not exist, creating using token: ${apiToken}`)
        await app.stores.apiTokenStore.insert({
            secret: apiToken,
            username: 'backend-admin',
            type: 'admin',
        }).catch(err => { throw new Error(`Failed to insert API token: ${err}`) })
        await app.services.apiTokenService.fetchActiveTokens();
    }
}

// client token for unleash-jira Forge app
async function upsertJiraAppAPIToken(app) {
    const logger = app.config.getLogger('index.js')
    logger.info(`Upserting API token for jira-unleash app`)
    const apiToken = config.unleashJiraToken;
    const apiTokenExists = await app.stores.apiTokenStore.exists(apiToken)
    if (!apiTokenExists) {
        logger.info(`unleash-jira API token not exist, creating using token: ${apiToken}`)
        await app.stores.apiTokenStore.insert({
            secret: apiToken,
            username: 'unleash-jira',
            type: 'client',
        }).catch(err => { throw new Error(`Failed to insert API token: ${err}`) })
        await app.services.apiTokenService.fetchActiveTokens();
    }
}

const config = yaml.load(fs.readFileSync(process.env.FILE_SECRETS))
let options = {
    authentication: {
        createAdminUser: false, // we will create an admin user ourselves
    },
    databaseUrl: config.db_connection,
    email: {
        host: 'smtp.gmail.com',
        smtpuser: config.smtpuser,
        smtppass: config.smtppass,
        sender: config.sender,
    },
};

async function importIfNotExists(app) {
    const { stateService } = app.services;
    console.info('Starting import if not exists')
    const exportedData = await stateService.export({
        includeStrategies: true,
        includeFeatureToggles: true,
        includeTags: true,
        includeProjects: true,
    });

    if (exportedData.features === undefined
        || exportedData.features.length == 0
        || process.env.FORCE_IMPORT == 'true') {
        console.info(`Starting import with file ${process.env.UNLEASH_IMPORT_FILE}`)
        await stateService.importFile({
            file: process.env.UNLEASH_IMPORT_FILE,
            keepExisting: process.env.KEEP_EXISTING == 'true',
            dropBeforeImport: process.env.DROP_BEFORE_IMPORT == 'true',
        })
    }
    console.info('End import if not exists')
}

async function addCustomContexts(app, newContexts) {
    const { contextService } = app.services
    const allContexts = await contextService.getAll()
    for (let index = 0; index < newContexts.length; index++) {
        const { name, description, stickiness } = newContexts[index]
        if (allContexts.some((context) => context.name === name)) {
            console.info('[addCustomContexts] Context', name, 'already exists.')
            continue
        }
        console.info('[addCustomContexts] Adding new context', name)
        await contextService.createContextField(
            { name, description, stickiness },
            'admin'
        )
    }
}

async function startUnleash() {
    var app = await unleash.start(options);
    await upsertAdminUser(app, config);
    await upsertClientAPIToken(app);
    if (config.adminAPIToken) {
        await upsertAdminAPIToken(app);
    }
    if (config.unleashJiraToken) {
        await upsertJiraAppAPIToken(app)
    }
    const newContexts = [
        { name: 'env', description: 'properties.env in client. Allows you to constrain on environment', stickiness: true },
        { name: 'org', description: 'properties.org in client. Allows you to constrain on organization', stickiness: true }
    ]
    await addCustomContexts(app, newContexts)
    await importIfNotExists(app);
}

startUnleash();
