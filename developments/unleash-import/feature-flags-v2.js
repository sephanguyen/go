const {
    getFeatures,
    removeFeatures,
    updateFeatures,
    createFeatures,
    updateFeatureTags,
    org,
    env,
} = require('./apis.js');
const { pick, getValueInLocal, sortStrategies } = require("./utils.js")
const fs = require('fs');
const { execSync } = require("node:child_process");


const _ = require('lodash');

async function getValueInRemote() {
    const { features } = await getFeatures();

    var result = {
        features: {},
        tags: {}
    }

    features.forEach((current) => {
        result.features = {
            ...result.features,
            [current.name]: {
                ...pick({
                    ...current,
                    strategies: sortStrategies((current.strategies || []).map(item => {
                        // clone to another object to remove id
                        const captureItem = Object.assign({}, item);
                        delete captureItem.id
                        return captureItem
                    }))
                }, ["name", "description", "type", "stale", "variants", "enabled", "strategies"]),
            }
        }

        result.tags = {
            ...result.tags,
            // pick first tag only
            [current.name]: Array.isArray(current.tags) ? pick(current.tags[0], ["type", "value"]) : {}
        }
    })

    return result;
}


function getObjectDiff(localObject, remoteObject) {
    const createObj = {};
    const removeObj = {};
    const updateObj = {};

    // Check properties in remoteObject that are not in localObject (remove)
    for (const prop in remoteObject) {
        if (!localObject.hasOwnProperty(prop)) {
            removeObj[prop] = remoteObject[prop];
        }
    }

    // Check properties in localObject that are not in remoteObject (create)
    for (const prop in localObject) {
        if (!remoteObject.hasOwnProperty(prop)) {
            createObj[prop] = localObject[prop];
            delete localObject[prop]; // remove from localObject to avoid update in next step
        }
    }

    // Check properties in both objects and identify updates
    for (const prop in localObject) {
        if (!_.isEqual(localObject[prop], remoteObject[prop])) {
            updateObj[prop] = localObject[prop];

            console.log("remoteObject", JSON.stringify(remoteObject[prop], null, 4))
            console.log("localObject", JSON.stringify(localObject[prop], null, 4))
        }
    }

    return {
        create: createObj,
        remove: removeObj,
        update: updateObj
    };
}

async function upsertFeatureFlag({ dryRun = false }) {
    // 1. features from unleash local
    const localValues = filterConstraint(getValueInLocal(org))

    // 2. features from unleash remote
    const remoteValues = await getValueInRemote()

    // 3. compare and update features
    const diff = getObjectDiff(localValues.features, remoteValues.features)

    console.info("Diff features: ", JSON.stringify(diff, null, 2))

    fs.writeFileSync(`./${org}_${env}_diff.md`,
`#### unleash ${org} ${env}: feature flags
\`\`\`json
${JSON.stringify(diff, null, 2)}
\`\`\`
`
    );

    // run exec to apply diff to $GITHUB_STEP_SUMMARY
    execSync(`cat ${org}_${env}_diff.md >> $GITHUB_STEP_SUMMARY`)

    if(dryRun) return;

    await createFeatures(Object.values(diff['create']));
    await updateFeatures(Object.values(diff['update']));
    await removeFeatures(Object.keys(diff['remove']));


    // 4. compare and create/update tags on new flags only
    const diffTags = getObjectDiff(localValues.tags, remoteValues.tags)
    console.info("Diff tags: ", diffTags)

    await updateFeatureTags({
        ...diffTags.create,
        ...diffTags.update
    });

}

function filterConstraint(features) {
    if (process.env.ENABLE_CONSTRAINT === 'true') return features
    const localFeatures = JSON.parse(JSON.stringify(features))

    for (const feature in localFeatures.features) {
        const value = localFeatures.features[feature]
        if (value.strategies.length > 0) {
            value.strategies.forEach((strategy) => {
                strategy.constraints = []
            })
        }
    }
    return localFeatures
}


module.exports = {
    upsertFeatureFlag,
    getObjectDiff
}
