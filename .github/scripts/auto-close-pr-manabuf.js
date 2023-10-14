module.exports = async ({ context, github, repo, exec, tempBranch }) => {
    const resp = await github.rest.pulls.list({
        owner: context.repo.owner,
        repo: repo,
        state: "open"
    });
    let pullRequests = [];
    if (resp.data && resp.data.length > 0) {
        resp.data.forEach(pr => {
            pullRequests.push({
                branch: pr.head.ref,
                pull_number: pr.number
            });
        })
    }
    if (pullRequests.length > 0) {
        console.log("pullRequests", pullRequests);
        for (const pr of pullRequests) {
            const branchName = pr.branch;
            const commitHash = branchName.split("_")[1];
            if (commitHash) {
                let output = "";
                let error = "";

                const options = {};
                options.listeners = {
                    stdout: (data) => {
                        output += data.toString();
                    },
                    stderr: (data) => {
                        error += data.toString();
                    }
                };
                try {
                    await exec.exec(`git branch ${tempBranch} --contains ${commitHash} | echo not found`, [], options);
                    if (output.includes(tempBranch)) {
                        const resp = await github.rest.pulls.update({
                            owner: context.repo.owner,
                            repo: repo,
                            pull_number: pr.pull_number,
                            state: "close"
                        });
                        console.log(resp)
                    }
                } catch (error) {
                    console.log(error);
                }
            }
        }
    }
}
