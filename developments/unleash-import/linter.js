const { getValueInLocal, compareDiff } = require("./utils.js")
const { violateFeatureFlags } = require("./feature-violate.js")

const fs = require('fs');
const path = require('path');

function linterFeatureFlags(features) {
    Object.keys(features).forEach((name) => {
        const feature = features[name];

        if (!feature.name) {
            throw new Error(`Feature ${name} is missing name: EX: name: "feature_name"`);
        }

        if (!feature.type) {
            throw new Error(`Feature ${name} is missing type. EX: type: "release"`);
        }

        if (!feature.description) {
            throw new Error(`Feature ${name} is missing description. EX: description: "This is description"`);
        }

        if (typeof feature.enabled !== "boolean") {
            throw new Error(`Feature ${name} is missing enabled. EX: enabled: true/false`);
        }


        if (!feature.variants) {
            throw new Error(`Feature ${name} is missing variants. EX: variants: []`);
        }

        if (!feature.strategies) {
            throw new Error(`Feature ${name} is missing strategies, If you dont have any strategies, please use empty array. EX strategies: []`);
        }

        feature.strategies.forEach((strategy) => {
            if (!strategy.name) {
                throw new Error(`Feature ${name} is missing strategy name. EX: name: "strategy_environment"`);
            }

            if (!strategy.parameters) {
                if (strategy.name != 'default') {
                    throw new Error(`Feature ${name} is missing strategy parameters. EX: parameters: { "environments": 'stag' }`);
                }
            }

            for (const key in strategy.parameters) {
                if (!strategy.parameters.hasOwnProperty(key)) {
                    throw new Error(`Feature ${name} is missing strategy parameters key. EX: parameters: { "environments": 'stag' }`);
                }
                
                if (!strategy.parameters[key]) {
                    throw new Error(`Feature ${name} is missing strategy parameters value. EX: parameters: { "environments": 'stag' }`);
                }
            }
            
            strategy.constraints.forEach((constraint) => {
                for (const key in constraint) {
                    if (!constraint[key] && typeof constraint[key] != "boolean") {
                        throw new Error(`Feature ${name} is missing value for '${key}' key.`);
                    }
                }
            })

        })

    })
}

function linterValueInLocal() {
    const manabie = getValueInLocal('manabie')
    linterFeatureFlags(manabie.features)

    // we will remove jprep unleash later
    const jprep = getValueInLocal('jprep')
    linterFeatureFlags(jprep.features)

    const CHECK_VIOLATE_FLAGS = process.env.CHECK_VIOLATE_FLAGS || "false";

    if(CHECK_VIOLATE_FLAGS === "true") {
        const src = JSON.parse(fs.readFileSync(path.resolve(__dirname, 'violated-feature-flags.json'), {
            encoding: 'utf8',
        }))
        const dest = violateFeatureFlags();

        const manabieDiffs = compareDiff(dest["manabie"], src["manabie"]) || []
        const jprepDiffs = compareDiff(dest["jprep"], src["jprep"]) || []


        let violateFlags = {}
        for (obj of manabieDiffs) {
            const [key, feature] = obj;

            if(!src['manabie']?.[key]) {
                violateFlags[key] = feature
            }
        }
        for (obj of jprepDiffs) {
            const [key, feature] = obj;

            if(!src['jprep']?.[key]) {
                violateFlags[key] = feature
            }
        }

        if(Object.keys(violateFlags).length > 0) {
            throw new Error(`Violate feature flags: ${Object.keys(violateFlags).join(', ')}`)
        }
    }
}

linterValueInLocal()