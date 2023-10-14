async function createDeploymentStatus({ github, core }) {
    const repo = process.env.REPO;
    const tag = process.env.RELEASE_TAG;
    const env = process.env.ENVIRONMENT;
    let state = process.env.STATE;
    const deploymentId = process.env.DEPLOYMENT_ID;
    const owner = "manabie-com";
    if (
        ![
            "error",
            "failure",
            "inactive",
            "in_progress",
            "pending",
            "queued",
            "success",
        ].includes(state)
    ) {
        state = 'pending'
    }

    if (state === "in_progress") {
        const deployment = await github.rest.repos.createDeployment({
            owner: owner,
            repo: repo,
            ref: tag,
            auto_merge: false,
            environment: env,
            production_environment: true,
            description: tag,
            required_contexts: [],
        });
        core.setOutput("deployment_id", deployment.data.id);
    } else {
        if (deploymentId === "" || deploymentId === undefined) {
            core.setFailed("Missed DEPLOYMENT_ID env!");
            return;
        }
        await github.rest.repos.createDeploymentStatus({
            owner: owner,
            repo: repo,
            deployment_id: deploymentId,
            state: state,
        });
    }
}

module.exports = {
    createDeploymentStatus,
};
