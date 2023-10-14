const fs = require('fs');
const path = require("path");
const yaml = require('js-yaml');
const { isEqual, toPairs, differenceWith } = require("lodash");

function getValueInLocal(org) {
    let fileName = `feature-values.yaml`;
    if (org === "jprep") {
        fileName = `jprep-feature-values.yaml`;
    }

    const squads = fs.readdirSync(path.resolve(__dirname, `../../feature_flags`))

    var result = {
        features: {},
        tags: {}
    }

    squads.forEach((squad) => {
        const data = yaml.load(fs.readFileSync(path.resolve(__dirname, `../../feature_flags/${squad}/${fileName}`), 'utf8'));
        const flags = data?.unleash?.featureFlags?.[squad] || [];

        flags.forEach((current) => {
            result.features = {
                ...result.features,
                [current.name]: pick({
                    ...current,
                    strategies: sortStrategies(current.strategies || []),
                }, ["name", "description", "type", "stale", "variants", "enabled", "strategies"])
            }

            result.tags = {
                ...result.tags,
                [current.name]: {
                    type: "team",
                    value: squad
                }
            }
        })
    })

    return result;
}

function sortStrategies(strategies) {
    const cloneStrategies = JSON.parse(JSON.stringify((strategies || [])))
    return cloneStrategies.sort((a, b) => {
        if (a.name != b.name) return b.name.localeCompare(a.name)

        return JSON.stringify(b.parameters).localeCompare(JSON.stringify(a.parameters))
    }) || [];
}

function pick(obj, keys) {
    var result = {}
    keys.forEach(function (key) {
        result[key] = obj[key]
    })

    return result
}

function compareDiff(src, dest) {
    const diff = differenceWith(
        Array.isArray(src) ? src : toPairs(src),
        Array.isArray(dest) ? dest: toPairs(dest),
        isEqual
    );

    return diff
}
    
module.exports = {
    pick, getValueInLocal, sortStrategies, compareDiff
}