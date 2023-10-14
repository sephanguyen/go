const { isThisHostAlive } = require("./ping-url");
const { isSelfHostRunnerAvailable } = require("./check-hcmo-runners");

const hosts = [
  "web-api.staging-green.manabie.io",
  "teacher.staging.manabie.io",
  "learner.staging.manabie.io",
  "backoffice.staging.manabie.io",
];

async function getRunnerLabelsForBlocker(github, repo, workflow) {
//   const selfHosts = await isSelfHostRunnerAvailable(
//     github,
//     "Linux",
//     "amd"
//   );

//   if (selfHosts) return ["self-hosted", "amd"];

  // hosts.forEach(function (host) {
  //   const isAlive = isThisHostAlive(host);
  //   if (!isAlive) return ["self-hosted", "8-32-non-persistent-large-runner"];
  // });

  // return ["self-hosted", "amd5950x"];

  return ["self-hosted", "8-32-non-persistent-large-runner"];
}

exports.getRunnerLabelsForBlocker = getRunnerLabelsForBlocker;
