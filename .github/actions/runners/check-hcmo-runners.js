async function isSelfHostRunnerAvailable(github, os, label) {
  if (!process.env.GITHUB_TOKEN) {
    console.log("skip check self-hosted runners because inputs.token is empty");
    return 0;
  }

  const result = await github.rest.actions.listSelfHostedRunnersForOrg({
    org: "manabie-com",
    per_page: 100,
  });

  if (!result?.data || !result?.data?.runners) return 0;

  const selfHosts = result.data.runners.filter(
    (r) =>
      String(r.os).indexOf(os) !== -1 &&
      r.status === "online" &&
      String(r.name).indexOf(label) !== -1 &&
      r.busy === false
  );

  console.log("How many self-hosted runners are available?", selfHosts);

  return selfHosts.length;
}

exports.isSelfHostRunnerAvailable = isSelfHostRunnerAvailable;
