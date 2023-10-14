locals {
  project_id = "staging-manabie-online"
  region     = "asia-southeast1"

  # ID of the pool to be created
  pool_id = "gh-action-pool"

  # Attribute condition for the deploy bot.
  # Only these workflows may use the WIF to use the deploy bot.
  deploy_bot_attribute_condition = <<EOT
    attribute.workflow == 'ci.auto_deploy_monitoring.yml' ||
    attribute.workflow == 'deployment.ad-hoc' ||
    attribute.workflow == 'deployment.deploy-job' ||
    attribute.workflow == 'deployment.platform.yml' ||
    attribute.workflow == 'deployment.run_j4' ||
    attribute.workflow == 'deployment.preproduction' ||
    attribute.workflow == 'tbd.deploy' ||
    attribute.workflow == 'mfe.deploy' ||
    attribute.workflow == 'deployment.set-log-level' ||
    attribute.workflow == 'deployment.k8s-runner.yml' ||
    attribute.workflow == 'deployment.uninstall_preproduction' ||
    attribute.workflow == 'deployment.disable_platform.yml' ||
    attribute.workflow == 'check_certs' ||
    attribute.workflow == 'tbd.build-and-deploy-learnosity'
  EOT
}
