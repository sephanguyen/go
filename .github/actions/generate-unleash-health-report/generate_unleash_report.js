const subProcess = require('child_process')
const EXPECTED_LIFETIME = 120 * 24 * 3600 * 1000 // 120 days
const MAX_LIFETIME_AT_FIRST_REQUEST = 30 * 24 * 3600 * 1000 // 30 days

function unleashHealthReport(admin_token, api_url) {
    const now = new Date()

    const features = getFeature(admin_token, api_url)
    if (!features || features.length == 0) {
        console.log('No feature toggle was found')
        return
    }
    const health_report = {}

    var stale_features = 0
    var no_usage_features = 0
    var healthy_features = 0

    features.forEach(feature => {
        const lifeTime = getTimeDiff(now, Date.parse(feature.createdAt))
        const lastSeenAt = getTimeDiff(now, Date.parse(feature.lastSeenAt))

        let status = []
        if (lifeTime > EXPECTED_LIFETIME) {
            status.push(`Lifetime exceeded ${msecToTime(EXPECTED_LIFETIME)}`);
            stale_features++
        }
        if (lastSeenAt == -1 && lifeTime > MAX_LIFETIME_AT_FIRST_REQUEST) {
            status.push(`No usage reported from last ${msecToTime(MAX_LIFETIME_AT_FIRST_REQUEST)}`);
            no_usage_features++
        }

        if (status.length == 0) {
            healthy_features++
            return
        }

        // const squad = feature.tags[0].value || 'other';
        const squad = feature.name.split('_')[0] == 'BACKEND' ? feature.name.split('_')[1] : feature.name.split('_')[0]

        if (!health_report[squad]) health_report[squad] = []

        health_report[squad].push({
            "name": feature.name,
            "lifeTime": msecToTime(lifeTime),
            "lastSeenAt": msecToTime(lastSeenAt),
            "status": status.join(' + ')
        })
    })
    const total_features = features.length;
    const health_rating = Math.floor(healthy_features / total_features * 100)
    const unhealthy_features = total_features - healthy_features
    const message = `Health rating: ${health_rating}%. ${total_features} active toggles. &#9888; ${unhealthy_features} feature toggles need attention`

    console.log('Number of features: ' + total_features)
    console.log(`Number of potentially stale features: ${stale_features} (${Math.floor(stale_features / total_features * 100)}%)`)
    console.log(`Number of no usage features: ${no_usage_features} (${Math.floor(no_usage_features / total_features * 100)}%)`)

    return {
        report: arrayToTable(health_report),
        message: message
    }
}

function getFeature(admin_token, api_url) {
    const get_all_feature_command =
        `
        curl -H "Content-Type: application/json" \
            -H "Authorization: ${admin_token}" \
            -X GET \
            ${api_url}
        `
    try {
        const stdout = subProcess.execSync(get_all_feature_command)
        return JSON.parse(stdout).features
    } catch (e) {
        console.log(`Status Code: ${e.status} with '${e.message}'`);
    }
}

function getTimeDiff(now, date) {
    if (!date) return -1
    return now.getTime() - date;
}

function msecToTime(msec) {
    if (msec == -1) return "-"

    var seconds = Math.floor(msec / 1000);
    var minutes = Math.floor(seconds / 60);
    var hours = Math.floor(minutes / 60);
    var days = Math.floor(hours / 24);

    if (days) return `${days} days`
    if (hours) return `${hours} hours`
    if (minutes) return `${minutes} minutes`
    if (seconds) return `${seconds} seconds`
}

function arrayToTable(array) {
    let tables = []
    for (const squadData in array) {
        const header = Object.keys(array[squadData][0])
        const table = [
            `<details>`,
            `<summary> ${squadData} - ${array[squadData].length} features </summary>`,
            ``,
            `| ${header.join(' | ')} |`,
            `| --- | --- | --- | --- |`,
            ...array[squadData].map(row => `| ${header.map(fieldName => row[fieldName]).join(' | ')} |`),
            `</details>`,
            ``
        ].join('\r\n')
        tables.push(table)
    }
    return (tables.join('\n'))
}

module.exports = {
    unleashHealthReport
}
