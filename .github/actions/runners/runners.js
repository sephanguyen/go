const { getRunnerLabelsForBlocker } = require("./pre-merge");
const { isSelfHostRunnerAvailable } = require("./check-hcmo-runners");

function dynamicRunners(k8sRunner) {
  return async function (github, repo, workflow) {
    const selfHosts = await isSelfHostRunnerAvailable(
      github,
      "Linux",
      "arc-runner-hcm-8-32",
    );

    if (selfHosts) return ["arc-runner-hcm-8-32"];

    return k8sRunner;
  };
}

function getRegressionRunners() {
  return async function (github) {
    const availableSelfHosts = await isSelfHostRunnerAvailable(
      github,
      "Linux",
      "arc-runner-hcm-8-32"
    );

    const isScheduled = process.env.GITHUB_EVENT_NAME === "schedule"
    const isLocal = process.env.ENV === "local"

    if (availableSelfHosts >=1 && isScheduled)
      return ["arc-runner-hcm-8-32"]

    if (isLocal) return ["arc-runner-hcm-8-32"]

    return ["self-hosted", "8-32-non-persistent-large-runner"]
  }
}

const runnersJson = {
  eibanam: {
    "e2e.trigger": {
      "run-e2e": getRunnerLabelsForBlocker,
      "pr-description": ["self-hosted", "100m-400mi", "spot", "persistent"],
    },
    "tiered.post-merge": {
      "run-e2e": ["arc-runner-hcm-8-32"],
    },
    "tiered.pre-merge": {
      "run-e2e-extra-test": getRunnerLabelsForBlocker,
      "check-commit-messages": [
        "self-hosted",
        "100m-400mi",
        "spot",
        "persistent",
      ],
      "install-deps": dynamicRunners(["self-hosted", "4-16-large-runner"]),
      lint: dynamicRunners(["self-hosted", "4-16-large-runner"]),
      "pr-description": ["self-hosted", "100m-400mi", "spot", "persistent"],
      "run-e2e-blocker-test": getRunnerLabelsForBlocker,
      "run-e2e-blocker-test-all": getRunnerLabelsForBlocker,
      "unit-test": dynamicRunners(["self-hosted", "4-16-large-runner"]),
    },
    "tiered.regression": {
      "run-e2e": getRegressionRunners(),
    },
  },
  backend: {
    "tbd.build": {
      "build-me-android": ["self-hosted", "2-16-large-runner"],
      "build-me-ios": ["self-hosted", "macos"],
      "build-fe": ["self-hosted", "2-16-large-runner"],
      "build-me-web-teacher": ["self-hosted", "2-16-large-runner"],
      "build-me-web-learner": ["self-hosted", "2-16-large-runner"],
      "build-backend": ["self-hosted", "4-8-large-runner"],
      "mfe-root-shell": ["self-hosted", "4-16-large-runner"],
      "mfe-entrypoint": ["self-hosted", "4-8-large-runner"],
    },
    "dummy_workflow": {
      staging: {
        "deploy-learner-android": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-learner-ios": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-learner-web": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-teacher-web": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-frontend": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-k8s": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-backoffice-k8s": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-teacher-k8s": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-learner-k8s": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-appsmith": ["self-hosted", "100m-400mi", "spot", "persistent"]
      },
      uat: {
        "deploy-learner-android": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-learner-ios": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-learner-web": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-teacher-web": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-frontend": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-k8s": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-backoffice-k8s": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-teacher-k8s": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-learner-k8s": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-appsmith": ["self-hosted", "100m-400mi", "spot", "persistent"]
      },
      production: {
        "deploy-learner-android": ["self-hosted", "medium-runner"],
        "deploy-learner-ios": ["self-hosted", "medium-runner"],
        "deploy-learner-web": ["self-hosted", "medium-runner"],
        "deploy-teacher-web": ["self-hosted", "medium-runner"],
        "deploy-frontend": ["self-hosted", "medium-runner"],
        "deploy-k8s": ["self-hosted", "large-runner"],
        "deploy-backoffice-k8s": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-teacher-k8s": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-learner-k8s": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "deploy-appsmith": ["self-hosted", "100m-400mi", "spot", "persistent"]
      },
    },
    "tbd.deploy": {
      staging: {
        "deploy-learner-android": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-learner-ios": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-learner-web": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-teacher-web": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-frontend": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-k8s": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-backoffice-k8s": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-teacher-k8s": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-learner-k8s": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-appsmith": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "mfe-entrypoint": ["self-hosted", "2-4-runner"],
        "mfe-root-shell": ["self-hosted", "custom-runner", "medium-runner"],
      },
      uat: {
        "deploy-learner-android": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-learner-ios": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-learner-web": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-teacher-web": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-frontend": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-k8s": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-backoffice-k8s": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-teacher-k8s": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-learner-k8s": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-appsmith": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "mfe-entrypoint": ["self-hosted", "2-4-runner"],
        "mfe-root-shell": ["self-hosted", "custom-runner", "medium-runner"],
      },
      production: {
        "deploy-learner-android": ["self-hosted", "medium-runner"],
        "deploy-learner-ios": ["self-hosted", "medium-runner"],
        "deploy-learner-web": ["self-hosted", "medium-runner"],
        "deploy-teacher-web": ["self-hosted", "medium-runner"],
        "deploy-frontend": ["self-hosted", "medium-runner"],
        "deploy-k8s": ["self-hosted", "large-runner"],
        "deploy-backoffice-k8s": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-teacher-k8s": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-learner-k8s": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-appsmith": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "mfe-entrypoint": ["self-hosted", "2-4-runner"],
        "mfe-root-shell": ["self-hosted", "custom-runner", "medium-runner"],
      },
      preproduction: {
        "deploy-learner-android": ["self-hosted", "medium-runner"],
        "deploy-learner-ios": ["self-hosted", "medium-runner"],
        "deploy-learner-web": ["self-hosted", "medium-runner"],
        "deploy-teacher-web": ["self-hosted", "medium-runner"],
        "deploy-frontend": ["self-hosted", "medium-runner"],
        "deploy-k8s": ["self-hosted", "large-runner"],
        "deploy-backoffice-k8s": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-teacher-k8s": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-learner-k8s": ["self-hosted", "custom-runner", "medium-runner"],
        "deploy-appsmith": ["self-hosted", "100m-400mi", "spot", "persistent"],
        "mfe-entrypoint": ["self-hosted", "2-4-runner"],
        "mfe-root-shell": ["self-hosted", "custom-runner", "medium-runner"],
      },
    },
    "tiered.regression": {
      "run-integration-test": ["self-hosted", "8-32-non-persistent-large-runner"],
    },
    "tiered.post_merge_integration_test": {
      "run-integration-test": ["self-hosted", "8-32-non-persistent-large-runner"],
    },
    "tiered.pre_merge": {
      "check-commit-messages": ["self-hosted", "100m-400mi", "spot", "persistent"],
      "conclusion": ["self-hosted", "100m-400mi", "spot", "persistent"],
      "dbschema-test": ["self-hosted", "2-4-runner"],
      "hasura-metadata-test": ["self-hosted", "custom-runner", "medium-runner"],
      "helm-test": ["self-hosted", "100m-400mi", "spot", "persistent"],
      "integration-blocker-test": dynamicRunners(["self-hosted", "8-32-non-persistent-large-runner"]),
      "lint": dynamicRunners(["self-hosted", "4-16-large-runner"]),
      "proto-check": ["self-hosted", "100m-400mi", "spot", "persistent"],
      "requirements": "!! requirements is the bootstrap job in this workflow, thus cannot use runners.json",
      "skaffold-test": dynamicRunners(["self-hosted", "4-8-large-runner"]),
      "unit-test": dynamicRunners(["self-hosted", "8-16-large-runner"]),
      "unleash-flags-only": ["self-hosted", "100m-400mi", "spot", "persistent"],
    },
    "mfe.build": {
      "install-deps": ["self-hosted", "custom-runner", "medium-runner"],
      "mfe-root-shell": ["self-hosted", "4-16-large-runner"],
      "mfe-entrypoint": ["self-hosted", "2-4-runner"],
    },
    "mfe.deploy": {
      "install-deps": ["self-hosted", "2-4-runner"],
      "mfe-entrypoint": ["self-hosted", "2-4-runner"],
      "mfe-root-shell": ["self-hosted", "custom-runner", "medium-runner"],
    },
  },
  "student-app": {
    "e2e.trigger": {
      "run-e2e": getRunnerLabelsForBlocker,
    },
    "tiered.pre-merge": {
      "run-e2e-extra-tests": getRunnerLabelsForBlocker,
      "run-e2e-blocker": getRunnerLabelsForBlocker,
      "run-e2e-blocker-test-all": getRunnerLabelsForBlocker,
      "run-e2e-critical": ["arc-runner-hcm-8-32"],
    },
    "tiered.post-merge": {
      "run-e2e-extra-tests": ["arc-runner-hcm-8-32"],
      "run-e2e-blocker": ["arc-runner-hcm-8-32"],
      "run-e2e-critical": ["arc-runner-hcm-8-32"],
    },
  },
  "school-portal-admin": {
    "e2e.trigger": {
      "pr-description": ["self-hosted", "100m-400mi", "spot", "persistent"],
      "run-e2e": getRunnerLabelsForBlocker,
    },
    "tiered.pre-merge": {
      "pr-description": ["self-hosted", "100m-400mi", "spot", "persistent"],
      "run-e2e-blocker-test": getRunnerLabelsForBlocker,
      "run-e2e-blocker-test-all": getRunnerLabelsForBlocker,
      "run-e2e-extra-test": getRunnerLabelsForBlocker,
      build: ["self-hosted", "4-16-large-runner"],
      "install-deps": ["self-hosted", "2-4-runner"],
      "unit-test-small-size": ["self-hosted", "2-4-runner"],
      "unit-test-medium-size": ["self-hosted", "4-16-large-runner"],
      "unit-test-large-size": ["self-hosted", "8-16-large-runner"],
      "unit-test": ["self-hosted", "8-16-large-runner"],
      linter: ["self-hosted", "4-16-large-runner"],
      eslint: ["self-hosted", "8-16-large-runner"],
      prettier: ["self-hosted", "4-16-large-runner"],
      "danger-depcheck-translation-depcruise": [
        "self-hosted",
        "4-16-large-runner",
      ],
      "sync-noti": ["self-hosted", "100m-400mi", "spot", "persistent"],
    },
    "tiered.post-merge": {
      "run-e2e-critical-test": ["arc-runner-hcm-8-32"],
      build: dynamicRunners(["self-hosted", "4-16-large-runner"]),
      "install-deps": dynamicRunners(["self-hosted", "2-4-runner"]),
      "unit-test-small-size": dynamicRunners(["self-hosted", "2-4-runner"]),
      "unit-test-medium-size": dynamicRunners([
        "self-hosted",
        "4-16-large-runner",
      ]),
      "unit-test-large-size": dynamicRunners([
        "self-hosted",
        "8-16-large-runner",
      ]),
      "unit-test": ["self-hosted", "8-16-large-runner"],
      linter: dynamicRunners(["self-hosted", "4-16-large-runner"]),
      eslint: dynamicRunners(["self-hosted", "8-16-large-runner"]),
      prettier: dynamicRunners(["self-hosted", "4-16-large-runner"]),
      "sync-noti": ["self-hosted", "100m-400mi", "spot", "persistent"],
    },
  },
};

async function getRunners({ repo, workflow, option, github }) {
  let runners = runnersJson[repo][workflow];
  if (option) runners = runners[option];

  console.log("original runner", runners);
  if (typeof runners === "function")
    return await runners(github, repo, workflow);

  for (const key of Object.keys(runners)) {
    if (typeof runners[key] === "function") {
      runners[key] = await runners[key](github, repo, workflow);
    }
  }

  return runners;
}

exports.getRunners = getRunners;
