
const { getValueInLocal } = require("./utils.js")
const fs = require('fs');
const path = require('path');


function findViolateFeatures(features) {
    let result = {}
    Object.keys(features).forEach((name) => {
        const feature = features[name];

        feature.strategies.forEach((strategy) => {
            if(strategy.name == 'strategy_organization' || strategy.name == 'strategy_variant') {
                result[name] = feature
                return
            }

            strategy.constraints.forEach((constraint) => {
                for (const key in constraint) {
                    if (key == 'contextName') {
                        if (constraint[key] == 'org') {
                            result[name] = feature
                            return
                        }
                    }
                }
            })

        })
    })
    
    return result
}

const violateFeatureFlags = (filePath) => {
    let result = {}
    const manabie = getValueInLocal('manabie')
    result['manabie'] = findViolateFeatures(manabie.features)

    // we will remove jprep unleash later
    const jprep = getValueInLocal('jprep')
    result['jprep'] = findViolateFeatures(jprep.features)


    console.log("Manabie have violated feature flags: ", Object.keys(result['manabie']).length)
    console.log("Jprep have violated feature flags: ", Object.keys(result['jprep']).length)
    
    if(filePath) {
        fs.writeFileSync(path.resolve(__dirname, filePath), JSON.stringify(result, null, 2))
    }

    return result
};

exports.violateFeatureFlags = violateFeatureFlags
exports.findViolateFeatures = findViolateFeatures

// Enable this line if you would like to generate violated-feature-flags.json
// violateFeatureFlags('violated-feature-flags.json')