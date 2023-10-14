const { importAccounts } = require('./accounts.js');
const { upsertFeatureFlag } = require("./feature-flags-v2.js")



function main() {
    if (process.env.IS_CREATE_ACCOUNT) {
        return importAccounts();
    }

    return upsertFeatureFlag({dryRun: process.env.DRY_RUN === 'true'});

}

main()