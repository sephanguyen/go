
async function run(github, context, core, files, codeOwnersFiles) {
  if (!!!files || !!!codeOwnersFiles) {
    return
  }  
  
  // fileOwners: array of
  // {
  //   "name": ".github/workflows/check_reviewers.js",
  //   "rule_match": "/.github/",
  //   "owners": [
  //     "@manabie-com/squad-platform"
  //   ]
  // },

  const fileOwners = files.map(f => Object.assign({ 
    name: f 
  }, codeOwnersFiles.fileMatches[f]))

  let ownersRequired = {}
  const orgPrefix = '@manabie-com/'
  fileOwners.forEach(
    v => v.owners.map(
      owner => ownersRequired[owner.substring(orgPrefix.length)] = true
    )
  )
  ownersRequired = Object.keys(ownersRequired)

  core.info("Reviews Required According to CODEOWNERS File and File Changes")
  core.info(JSON.stringify(ownersRequired, null, 2))
  
  let reviewedByTeams = await getTeamsReviewed({ github, context, core })

  for (let i = 0; i < ownersRequired.length; i++) {
    const name = ownersRequired[i]            
    if (!!reviewedByTeams[name]) {
      core.info("Required Review from " + name + " Exists")
      pass = true
    } else {
      core.info("Required Review from " + name + " Missing")
      pass = false
    }
    await setStatus({ github, context, core, pass, teamName: name })
  }
}


async function getTeamsReviewed({ github, context, core }) {
    const { issue } = context

    const pull_number = issue.number
    let pass = true

    const query = `query($owner:String!, $name:String!, $number:Int!) {
        repository(owner:$owner, name:$name){
          pullRequest(number:$number) {
            reviews (first: 100, states: APPROVED) {
              nodes {                    
                author {
                  login
                },
                state
                onBehalfOf (first:100) {
                  nodes {
                    name
                  }
                }
              }
            }
          }
        }
      }`;
    const variables = {
        owner: context.repo.owner,
        name: context.repo.repo,
        number: pull_number,
    }
    let response = await github.graphql(query, variables)
    core.info("Github API Approved Reviews GraphQL Response")
    core.info(JSON.stringify(response, null, 2))

    let teamsReviewed = {}

    if (!!response.repository.pullRequest.reviews && !!response.repository.pullRequest.reviews.nodes) {
        response.repository.pullRequest.reviews.nodes.map(
            review => review.onBehalfOf.nodes.forEach(team => {
                teamsReviewed[team.name] = true
            }
        ));
    }
    core.info("Teams Reviewed")
    core.info(JSON.stringify(Object.keys(teamsReviewed), null, 2))

    return teamsReviewed
}

async function setStatus({ github, context, core, pass, teamName }) {


    let sha = !!context.payload.pull_request && !!context.payload.pull_request.head ?
        context.payload.pull_request.head.sha :
        context.sha

    core.info('Commit Status for SHA: ' + sha)

    try {
        let res = await github.rest.repos.createCommitStatus({
            ...context.repo,
            sha,
            state: pass ? 'success' : 'failure',
            target_url: `${context.serverUrl}/${context.repo.owner}/${context.repo.repo}/actions/runs/${context.runId}`,
            context: `Reviewed by ${teamName}`,
            description: `Check that a member from ${teamName} has approved this PR`,
        })
    } catch (error) {
        core.info("Error: ", error.message)
        core.setFailed(error.message)
    }
}

module.exports = {
  run,
}