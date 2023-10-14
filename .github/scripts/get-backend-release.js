module.exports = async ({ context, github, environment }) => {
    environment = normalizeEnvironmentString(environment)
    switch (environment) {
        case 'staging':
            return await getLatestRCTag(context, github)
        case 'uat':
        case 'preproduction':
        case 'production':
            return await getLatestTag(context, github)
        default:
            throw new Error(`Unknown environment ${environment}`)
    }
}

// just for convenience
function normalizeEnvironmentString(environment) {
    switch (environment) {
        case 'stag':
            return 'staging'
        case 'dorp':
            return 'preproduction'
        case 'prod':
            return 'production'
        default:
            return environment
    }
}

const repo = 'backend'

// getLatestRCTag queries from Github API, sorts, then get the latest tag with -rc
async function getLatestRCTag(context, github) {
    const response = await github.rest.repos.listReleases({
        owner: context.repo.owner,
        repo: repo,
        per_page: 100,
    })
    const allTags = response.data.map((release) => release.tag_name)
    console.log(`allTags: ${allTags}`)
    const rcTags = allTags.filter((tag) => tag.includes('-rc'))
    if (rcTags.length === 0) {
        throw new Error(`No rc tags found in repository ${repo} (checking first 100 releases)`)
    }
    const sortedRCTags = rcTags.sort(customSort);
    console.log(`sortedTags: ${sortedRCTags}`)
    return sortedRCTags[0]
}

// getLatestTag queries from Github API, sorts, then get the latest tag (without -rc)
async function getLatestTag(context, github) {
    const response = await github.rest.repos.listReleases({
        owner: context.repo.owner,
        repo: repo,
        per_page: 100,
    })
    const allTags = response.data.map((release) => release.tag_name)
    console.log(`allTags: ${allTags}`)
    const nonRCTags = allTags.filter((tag) => !tag.includes('-rc'))
    if (nonRCTags.length === 0) {
        throw new Error(`No tags found in repository ${repo} (checking first 100 releases)`)
    }
    const sortedTags = nonRCTags.sort(customSort);
    console.log(`sortedTags: ${sortedTags}`)
    return sortedTags[0]
}

var customSort = function (a, b) {
    let startNumberA = Number((a.match(/(\d+)/g) || [])[0]);
    let startNumberB = Number((b.match(/(\d+)/g) || [])[0]);
    let endNumberA = Number((a.match(/(\d+$)/g) || [])[0]);
    let endNumberB = Number((b.match(/(\d+$)/g) || [])[0]);
    let startResult = startNumberB - startNumberA;
    if (startResult === 0) {
        return endNumberB - endNumberA;
    } else {
        return startResult;
    }
}
