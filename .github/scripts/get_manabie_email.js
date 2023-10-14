const hclParser = require("js-hcl-parser")
const { globSync } = require('glob')
const fs = require('fs').promises

const ACCESS_CONTROL_DIR = 'deployments/terraform/live/workspace/access-control/*_members.hcl'

async function hclToJson(hclFilePath) {
    const data = await fs.readFile(hclFilePath, "utf8")
    const jsonText = hclParser.parse(data)
    return JSON.parse(jsonText)
}

async function findEmail(githubName) {
    const files = globSync(ACCESS_CONTROL_DIR)
    var email
    for (const file of files) {
        const jsonData = await hclToJson(file)
        const members = jsonData.locals[0].members
        for (const member of members) {
            const github = member.github[0]
            if (!github.account) continue
            if (github.account.toLowerCase() == githubName.toLowerCase()) {
                console.log('User', githubName, 'email is', member.email)
                email = member.email
                break
            }
        }
        if (email) break
    }
    if (email) return email
    console.log('No email found for user', githubName)
}

module.exports = {
    findEmail,
}
