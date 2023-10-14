function normalizeEnvironmentString(env) {
    switch (env) {
        case 'stag':
            return 'staging'
        case 'dorp':
            return 'preproduction'
        case 'prod':
            return 'production'
        default:
            return env
    }
}

module.exports = {
    normalizeEnvironmentString,
}
