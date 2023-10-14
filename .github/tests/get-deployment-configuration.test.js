// This is a simple test suite for .github/scripts/get-deployment-configuration.js
// To run:
//   node .github/tests/get-deployment-configuration.test.js
const getDeploymentConfiguration = require('../scripts/get-deployment-configuration.js');
const assert = require('assert')

async function testPreProductionConfiguration(organization) {
    const prepConfig = await getDeploymentConfiguration({ organization: organization, environment: 'preproduction' })
    const prodConfig = await getDeploymentConfiguration({ organization: organization, environment: 'production' })
    assert.ok(prepConfig.cluster === prodConfig.cluster, `expected ${prodConfig.cluster} cluster, got ${prepConfig.cluster} for ${organization}`)
    assert.ok(prepConfig.region === prodConfig.region, `expected ${prodConfig.region} region, got ${prepConfig.region} for ${organization}`)
    assert.ok(prodConfig.environment === "prod", `expected prod environment, got ${prodConfig.environment} for ${organization}`)
    assert.ok(prepConfig.environment === "dorp", `expected dorp environment, got ${prepConfig.environment} for ${organization}`)
    assert.ok(prepConfig.organization === prodConfig.organization, `expected ${prodConfig.organization} organization, got ${prepConfig.organization} for ${organization}`)
    assert.ok(prepConfig.projectId === prodConfig.projectId, `expected ${prodConfig.projectId} projectId, got ${prepConfig.projectId} for ${organization}`)
    assert.ok(prepConfig.serviceAccountEmailSuffix === prodConfig.serviceAccountEmailSuffix, `expected ${prodConfig.serviceAccountEmailSuffix} serviceAccountEmailSuffix, got ${prepConfig.serviceAccountEmailSuffix} for ${organization}`)
    assert.ok(prepConfig.namespace.replace(/^dorp-/, "prod-") === prodConfig.namespace, `expected ${prodConfig.namespace} namespace, got ${prepConfig.namespace} for ${organization}`)
    assert.ok(prepConfig.kmsPath === prodConfig.kmsPath, `expected ${prodConfig.kmsPath} kmsPath, got ${prepConfig.kmsPath} for ${organization}`)
    assert.ok(prepConfig.sqlProxySourceConnectionName === prodConfig.sqlProxyConnectionName, `expected ${prodConfig.sqlProxyConnectionName} sqlProxySourceConnectionName, got ${prepConfig.sqlProxySourceConnectionName} for ${organization}`)

    // Pre-production-only configs
    if (organization !== "jprep") {
        await assertSQLProxyConnectionName(prepConfig.sqlProxyConnectionName, prodConfig.sqlProxyConnectionName)
    } else {
        assert(prepConfig.sqlProxyConnectionName === 'student-coach-e1e95:asia-northeast1:clone-jprep-6a98')
        assert(prepConfig.sqlProxySourceConnectionName === 'live-manabie:asia-northeast1:jprep-6a98')
    }

    // Pre-production elastic configs. Storage sizes are not tested since they vary.
    if (organization !== "jprep") {
        assert(prepConfig.elasticNamespace === `dorp-${organization}-elastic`)
        assert(prepConfig.elasticReleaseName === 'elastic')
        assert(prepConfig.elasticNameOverride === '')
        assert(prepConfig.elasticReplicas === 3)
        assert(prepConfig.elasticStorageClass === 'premium-rwo')
        assert(prepConfig.elasticSnapshotEnabled === 'true')
        assert(prepConfig.elasticSnapshotStorageClass === 'premium-rwo')
        assert(prepConfig.elasticCreateServiceAccount === 'true')
        assert(prepConfig.elasticInitIndices === 'true')
    } else {
        assert(prepConfig.elasticNamespace == 'dorp-jprep-elastic')
        assert(prepConfig.elasticReleaseName === 'dorp-jprep')
        assert(prepConfig.elasticNameOverride === 'dorp-jprep')
        assert(prepConfig.elasticReplicas === 3)
        assert(prepConfig.elasticStorageClass === 'ssd')
        assert(prepConfig.elasticSnapshotEnabled === 'true')
        assert(prepConfig.elasticSnapshotStorageClass === 'standard')
        assert(prepConfig.elasticCreateServiceAccount === 'true')
        assert(prepConfig.elasticInitIndices === 'true')
    }
}

async function assertSQLProxyConnectionName(prepConnName, prodConnName) {
    const prepComponents = prepConnName.split(':')
    const prodComponents = prodConnName.split(':')
    assert(prepComponents[0] === prodComponents[0])
    assert(prepComponents[1] === prodComponents[1])
    assert(prepComponents[2] === `clone-${prodComponents[2]}`)
}

const orgList = [
    'jprep',
    'synersia',
    'renseikai',
    'ga',
    'aic',
    'tokyo'
]

for (let org of orgList) {
    testPreProductionConfiguration(org)
}