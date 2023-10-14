

module.exports = async ({ organization, environment }) => {
  switch (environment) {
    case 'staging':
      switch (organization) {
        case 'manabie':
          return ManabieDeploymentConfiguration.staging;
        case 'jprep':
          return JPrepDeploymentConfiguration.staging;
      }
      break;

    case 'uat':
      switch (organization) {
        case 'manabie':
          return ManabieDeploymentConfiguration.uat;

        case 'jprep':
          return JPrepDeploymentConfiguration.uat;
      }
      break;

    case 'preproduction':
      switch (organization) {
        case 'jprep':
          return JPrepDeploymentConfiguration.preproduction;
        case 'synersia':
          return SynersiaDeploymentConfiguration.preproduction;
        case 'renseikai':
          return RenseikaiDeploymentConfiguration.preproduction;
        case 'ga':
          return GADeploymentConfiguration.preproduction;
        case 'aic':
          return AICDeploymentConfiguration.preproduction;
        case 'tokyo':
          return TokyoDeploymentConfiguration.preproduction;
      }

    case 'production':
      switch (organization) {
        case 'manabie':
          return ManabieDeploymentConfiguration.production;
        case 'jprep':
          return JPrepDeploymentConfiguration.production;
        case 'synersia':
          return SynersiaDeploymentConfiguration.production;
        case 'renseikai':
          return RenseikaiDeploymentConfiguration.production;
        case 'ga':
          return GADeploymentConfiguration.production;
        case 'aic':
          return AICDeploymentConfiguration.production;
        case 'tokyo':
          return TokyoDeploymentConfiguration.production;
      }
      break;
  }
  return null;
}

class ManabieDeploymentConfiguration {
  static get staging() {
    return {
      cluster: 'staging-2',
      region: 'asia-southeast1',
      zone: 'asia-southeast1-b',
      environment: 'stag',
      env: 'stag',
      organization: 'manabie',
      org: 'manabie',
      projectId: 'staging-manabie-online',
      serviceAccountEmailSuffix: 'staging-manabie-online',
      namespace: 'stag-manabie-backend',
      kmsPath: 'projects/staging-manabie-online/locations/global/keyRings/deployments/cryptoKeys/github-actions',
      sqlProxyConnectionName: 'staging-manabie-online:asia-southeast1:manabie-common-88e1ee71',
      deployServiceAccount: 'stag-deploy-bot@staging-manabie-online.iam.gserviceaccount.com',
      provider: 'projects/456005132078/locations/global/workloadIdentityPools/gh-action-pool/providers/gh-action-provider',
      platformServices: "scheduling aphelios mlflow kserve mlmodel unistall_model monitoring kiali import-map-deployer backbone unleash elastic ksql data-warehouse redash appsmith gateway runner-controller kafka connect cp-schema-registry learnosity-web-view",
    }
  }

  static get uat() {
    return {
      cluster: 'staging-2',
      region: 'asia-southeast1',
      zone: 'asia-southeast1-b',
      environment: 'uat',
      env: 'uat',
      organization: 'manabie',
      org: 'manabie',
      projectId: 'staging-manabie-online',
      serviceAccountEmailSuffix: 'uat-manabie',
      namespace: 'uat-manabie-services',
      kmsPath: 'projects/staging-manabie-online/locations/global/keyRings/deployments/cryptoKeys/uat-manabie',
      sqlProxyConnectionName: 'staging-manabie-online:asia-southeast1:manabie-common-88e1ee71',
      deployServiceAccount: 'stag-deploy-bot@staging-manabie-online.iam.gserviceaccount.com',
      provider: 'projects/456005132078/locations/global/workloadIdentityPools/gh-action-pool/providers/gh-action-provider',
      platformServices: "import-map-deployer backbone unleash elastic ksql appsmith gateway kafka connect cp-schema-registry",
    }
  }

  static get production() {
    return {
      cluster: 'manabie',
      region: 'asia-southeast1',
      environment: 'prod',
      env: 'prod',
      organization: 'manabie',
      org: 'manabie',
      projectId: 'student-coach-e1e95',
      serviceAccountEmailSuffix: 'production-manabie-vn',
      namespace: 'prod-manabie-services',
      kmsPath: 'projects/student-coach-e1e95/locations/asia-southeast1/keyRings/manabie/cryptoKeys/prod-manabie',
      sqlProxyConnectionName: 'student-coach-e1e95:asia-southeast1:manabie-2db8',
      deployServiceAccount: 'prod-deploy-bot@student-coach-e1e95.iam.gserviceaccount.com',
      provider: 'projects/418860883682/locations/global/workloadIdentityPools/gh-action-pool/providers/gh-action-provider',
      platformServices: "monitoring",
    }
  }
}

class JPrepDeploymentConfiguration {
  static get staging() {
    return {
      cluster: 'staging-2',
      region: 'asia-southeast1',
      zone: 'asia-southeast1-b',
      environment: 'stag',
      env: 'stag',
      organization: 'jprep',
      org: 'jprep',
      projectId: 'staging-manabie-online',
      serviceAccountEmailSuffix: 'staging-manabie-online',
      namespace: 'stag-jprep-backend',
      kmsPath: 'projects/staging-manabie-online/locations/global/keyRings/deployments/cryptoKeys/github-actions',
      sqlProxyConnectionName: 'staging-manabie-online:asia-southeast1:jprep-uat',
      deployServiceAccount: 'stag-deploy-bot@staging-manabie-online.iam.gserviceaccount.com',
      provider: 'projects/456005132078/locations/global/workloadIdentityPools/gh-action-pool/providers/gh-action-provider',
      platformServices: "backbone ksql unleash elastic gateway kafka connect cp-schema-registry",
    }
  }

  static get uat() {
    return {
      cluster: 'staging-2',
      region: 'asia-southeast1',
      zone: 'asia-southeast1-b',
      environment: 'uat',
      env: 'uat',
      organization: 'jprep',
      org: 'jprep',
      projectId: 'staging-manabie-online',
      serviceAccountEmailSuffix: 'staging-manabie-online',
      namespace: 'uat-jprep-services',
      kmsPath: 'projects/staging-manabie-online/locations/global/keyRings/deployments/cryptoKeys/uat-jprep',
      sqlProxyConnectionName: 'staging-manabie-online:asia-southeast1:jprep-uat',
      deployServiceAccount: 'stag-deploy-bot@staging-manabie-online.iam.gserviceaccount.com',
      provider: 'projects/456005132078/locations/global/workloadIdentityPools/gh-action-pool/providers/gh-action-provider',
      platformServices: "backbone unleash elastic ksql kiali gateway kafka connect cp-schema-registry",
    }
  }

  static get preproduction() {
    return {
      cluster: 'production',
      region: 'asia-northeast1-c',
      environment: 'dorp',
      env: 'dorp',
      organization: 'jprep',
      org: 'jprep',
      projectId: 'live-manabie',
      serviceAccountEmailSuffix: 'live-manabie',
      namespace: 'dorp-jprep-services',
      kmsPath: 'projects/live-manabie/locations/global/keyRings/deployments/cryptoKeys/prod-jprep',
      sqlProxyConnectionName: 'student-coach-e1e95:asia-northeast1:clone-jprep-6a98',
      sqlProxySourceConnectionName: 'live-manabie:asia-northeast1:jprep-6a98',
      elasticNamespace: 'dorp-jprep-elastic',
      elasticReleaseName: 'dorp-jprep',
      elasticNameOverride: 'dorp-jprep',
      elasticReplicas: 3,
      elasticStorageClass: 'ssd',
      elasticStorageSize: '20Gi',
      elasticSnapshotEnabled: 'true',
      elasticSnapshotStorageClass: 'standard',
      elasticSnapshotStorageSize: '20Gi',
      elasticCreateServiceAccount: 'true',
      elasticInitIndices: 'true',
      deployServiceAccount: 'prod-deploy-bot@live-manabie.iam.gserviceaccount.com',
      provider: 'projects/719550208908/locations/global/workloadIdentityPools/gh-action-pool/providers/gh-action-provider',
      platformServices: "unleash nats kafka backbone ksql gateway",
    }
  }

  static get production() {
    return {
      cluster: 'tokyo',
      region: 'asia-northeast1',
      environment: 'prod',
      env: 'prod',
      organization: 'jprep',
      org: 'jprep',
      projectId: 'student-coach-e1e95',
      serviceAccountEmailSuffix: 'student-coach-e1e95',
      namespace: 'prod-jprep-services',
      kmsPath: 'projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jprep/cryptoKeys/prod-jprep',
      sqlProxyConnectionName: 'student-coach-e1e95:asia-northeast1:prod-jprep-d995522c',
      deployServiceAccount: 'prod-deploy-bot@student-coach-e1e95.iam.gserviceaccount.com',
      provider: 'projects/418860883682/locations/global/workloadIdentityPools/gh-action-pool/providers/gh-action-provider',
      platformServices: "unleash nats kafka backbone ksql gateway connect",
    }
  }
}

class SynersiaDeploymentConfiguration {
  static get preproduction() {
    return {
      cluster: 'jp-partners',
      region: 'asia-northeast1',
      environment: 'dorp',
      env: 'dorp',
      organization: 'synersia',
      org: 'synersia',
      projectId: 'student-coach-e1e95',
      serviceAccountEmailSuffix: 'synersia',
      namespace: 'dorp-synersia-services',
      kmsPath: 'projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-synersia',
      sqlProxyConnectionName: 'student-coach-e1e95:asia-northeast1:clone-prod-tokyo',
      sqlProxySourceConnectionName: 'student-coach-e1e95:asia-northeast1:prod-tokyo',
      elasticNamespace: 'dorp-synersia-elastic',
      elasticReleaseName: 'elastic',
      elasticNameOverride: '',
      elasticReplicas: 3,
      elasticStorageClass: 'premium-rwo',
      elasticStorageSize: '20Gi',
      elasticSnapshotEnabled: 'true',
      elasticSnapshotStorageClass: 'premium-rwo',
      elasticSnapshotStorageSize: '20Gi',
      elasticCreateServiceAccount: 'true',
      elasticInitIndices: 'true',
      deployServiceAccount: 'prod-deploy-bot@student-coach-e1e95.iam.gserviceaccount.com',
      provider: 'projects/418860883682/locations/global/workloadIdentityPools/gh-action-pool/providers/gh-action-provider',
      platformServices: "monitoring backbone unleash ksql kiali gateway",
    }
  }

  static get production() {
    return {
      cluster: 'jp-partners',
      region: 'asia-northeast1',
      environment: 'prod',
      organization: 'synersia',
      projectId: 'student-coach-e1e95',
      serviceAccountEmailSuffix: 'synersia',
      namespace: 'prod-synersia-services',
      kmsPath: 'projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-synersia',
      sqlProxyConnectionName: 'student-coach-e1e95:asia-northeast1:prod-tokyo',
      deployServiceAccount: 'prod-deploy-bot@student-coach-e1e95.iam.gserviceaccount.com',
      provider: 'projects/418860883682/locations/global/workloadIdentityPools/gh-action-pool/providers/gh-action-provider',
      platformServices: "monitoring backbone unleash ksql kiali gateway appsmith connect",
    }
  }
}

class RenseikaiDeploymentConfiguration {
  static get preproduction() {
    return {
      cluster: 'jp-partners',
      region: 'asia-northeast1',
      environment: 'dorp',
      organization: 'renseikai',
      projectId: 'student-coach-e1e95',
      serviceAccountEmailSuffix: 'production-renseikai',
      namespace: 'dorp-renseikai-services',
      kmsPath: 'projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-renseikai',
      sqlProxyConnectionName: 'student-coach-e1e95:asia-northeast1:clone-prod-tokyo',
      sqlProxySourceConnectionName: 'student-coach-e1e95:asia-northeast1:prod-tokyo',
      elasticNamespace: 'dorp-renseikai-elastic',
      elasticReleaseName: 'elastic',
      elasticNameOverride: '',
      elasticReplicas: 3,
      elasticStorageClass: 'premium-rwo',
      elasticStorageSize: '10Gi',
      elasticSnapshotEnabled: 'true',
      elasticSnapshotStorageClass: 'premium-rwo',
      elasticSnapshotStorageSize: '10Gi',
      elasticCreateServiceAccount: 'true',
      elasticInitIndices: 'true',
      deployServiceAccount: 'prod-deploy-bot@student-coach-e1e95.iam.gserviceaccount.com',
      provider: 'projects/418860883682/locations/global/workloadIdentityPools/gh-action-pool/providers/gh-action-provider',
      platformServices: "backbone unleash ksql gateway",
    }
  }

  static get production() {
    return {
      cluster: 'jp-partners',
      region: 'asia-northeast1',
      environment: 'prod',
      env: 'prod',
      organization: 'renseikai',
      org: 'renseikai',
      projectId: 'student-coach-e1e95',
      serviceAccountEmailSuffix: 'production-renseikai',
      namespace: 'prod-renseikai-services',
      kmsPath: 'projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-renseikai',
      sqlProxyConnectionName: 'production-renseikai:asia-northeast1:renseikai-83fc',
      deployServiceAccount: 'prod-deploy-bot@student-coach-e1e95.iam.gserviceaccount.com',
      provider: 'projects/418860883682/locations/global/workloadIdentityPools/gh-action-pool/providers/gh-action-provider',
      platformServices: "backbone unleash ksql gateway appsmith connecte elastic",
    }
  }
}

class GADeploymentConfiguration {
  static get preproduction() {
    return {
      cluster: 'jp-partners',
      region: 'asia-northeast1',
      environment: 'dorp',
      env: 'dorp',
      organization: 'ga',
      org: 'ga',
      projectId: 'student-coach-e1e95',
      serviceAccountEmailSuffix: 'production-ga',
      namespace: 'dorp-ga-services',
      kmsPath: 'projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-ga',
      sqlProxyConnectionName: 'student-coach-e1e95:asia-northeast1:clone-prod-tokyo',
      sqlProxySourceConnectionName: 'student-coach-e1e95:asia-northeast1:prod-tokyo',
      elasticNamespace: 'dorp-ga-elastic',
      elasticReleaseName: 'elastic',
      elasticNameOverride: '',
      elasticReplicas: 3,
      elasticStorageClass: 'premium-rwo',
      elasticStorageSize: '10Gi',
      elasticSnapshotEnabled: 'true',
      elasticSnapshotStorageClass: 'premium-rwo',
      elasticSnapshotStorageSize: '10Gi',
      elasticCreateServiceAccount: 'true',
      elasticInitIndices: 'true',
      deployServiceAccount: 'prod-deploy-bot@student-coach-e1e95.iam.gserviceaccount.com',
      provider: 'projects/418860883682/locations/global/workloadIdentityPools/gh-action-pool/providers/gh-action-provider',
      platformServices: "backbone unleash ksql gateway",
    }
  }

  static get production() {
    return {
      cluster: 'jp-partners',
      region: 'asia-northeast1',
      environment: 'prod',
      env: 'prod',
      organization: 'ga',
      org: 'ga',
      projectId: 'student-coach-e1e95',
      serviceAccountEmailSuffix: 'production-ga',
      namespace: 'prod-ga-services',
      kmsPath: 'projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-ga',
      sqlProxyConnectionName: 'student-coach-e1e95:asia-northeast1:jp-partners-b04fbb69',
      deployServiceAccount: 'prod-deploy-bot@student-coach-e1e95.iam.gserviceaccount.com',
      provider: 'projects/418860883682/locations/global/workloadIdentityPools/gh-action-pool/providers/gh-action-provider',
      platformServices: "backbone unleash ksql gateway appsmith connect",
    }
  }
}

class AICDeploymentConfiguration {
  static get preproduction() {
    return {
      cluster: 'jp-partners',
      region: 'asia-northeast1',
      environment: 'dorp',
      env: 'dorp',
      organization: 'aic',
      org: 'aic',
      projectId: 'student-coach-e1e95',
      serviceAccountEmailSuffix: 'production-aic',
      namespace: 'dorp-aic-services',
      kmsPath: 'projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-aic',
      sqlProxyConnectionName: 'student-coach-e1e95:asia-northeast1:clone-prod-tokyo',
      sqlProxySourceConnectionName: 'student-coach-e1e95:asia-northeast1:prod-tokyo',
      elasticNamespace: 'dorp-aic-elastic',
      elasticReleaseName: 'elastic',
      elasticNameOverride: '',
      elasticReplicas: 3,
      elasticStorageClass: 'premium-rwo',
      elasticStorageSize: '10Gi',
      elasticSnapshotEnabled: 'true',
      elasticSnapshotStorageClass: 'premium-rwo',
      elasticSnapshotStorageSize: '10Gi',
      elasticCreateServiceAccount: 'true',
      elasticInitIndices: 'true',
      deployServiceAccount: 'prod-deploy-bot@student-coach-e1e95.iam.gserviceaccount.com',
      provider: 'projects/418860883682/locations/global/workloadIdentityPools/gh-action-pool/providers/gh-action-provider',
      platformServices: "backbone unleash ksql gateway connect",
    }
  }

  static get production() {
    return {
      cluster: 'jp-partners',
      region: 'asia-northeast1',
      environment: 'prod',
      env: 'prod',
      organization: 'aic',
      org: 'aic',
      projectId: 'student-coach-e1e95',
      serviceAccountEmailSuffix: 'production-aic',
      namespace: 'prod-aic-services',
      kmsPath: 'projects/student-coach-e1e95/locations/asia-northeast1/keyRings/jp-partners/cryptoKeys/prod-aic',
      sqlProxyConnectionName: 'student-coach-e1e95:asia-northeast1:jp-partners-b04fbb69',
      deployServiceAccount: 'prod-deploy-bot@student-coach-e1e95.iam.gserviceaccount.com',
      provider: 'projects/418860883682/locations/global/workloadIdentityPools/gh-action-pool/providers/gh-action-provider',
      platformServices: "backbone unleash ksql gateway appsmith connect",
    }
  }
}

class TokyoDeploymentConfiguration {
  static get preproduction() {
    return {
      cluster: 'tokyo',
      region: 'asia-northeast1',
      environment: 'dorp',
      env: 'dorp',
      organization: 'tokyo',
      org: 'tokyo',
      projectId: 'student-coach-e1e95',
      serviceAccountEmailSuffix: 'student-coach-e1e95',
      namespace: 'dorp-tokyo-services',
      kmsPath: 'projects/student-coach-e1e95/locations/asia-northeast1/keyRings/prod-tokyo/cryptoKeys/prod-tokyo',
      sqlProxyConnectionName: 'student-coach-e1e95:asia-northeast1:clone-prod-tokyo',
      sqlProxySourceConnectionName: 'student-coach-e1e95:asia-northeast1:prod-tokyo',
      elasticNamespace: 'dorp-tokyo-elastic',
      elasticReleaseName: 'elastic',
      elasticNameOverride: '',
      elasticReplicas: 3,
      elasticStorageClass: 'premium-rwo',
      elasticStorageSize: '10Gi',
      elasticSnapshotEnabled: 'true',
      elasticSnapshotStorageClass: 'premium-rwo',
      elasticSnapshotStorageSize: '10Gi',
      elasticCreateServiceAccount: 'true',
      elasticInitIndices: 'true',
      deployServiceAccount: 'prod-deploy-bot@student-coach-e1e95.iam.gserviceaccount.com',
      provider: 'projects/418860883682/locations/global/workloadIdentityPools/gh-action-pool/providers/gh-action-provider',
      platformServices: "backbone unleash ksql monitoring kiali redash import-map-deployer gateway dwh-with-auth connect replication kafka",
    }
  }

  static get production() {
    return {
      cluster: 'tokyo',
      region: 'asia-northeast1',
      environment: 'prod',
      env: 'prod',
      organization: 'tokyo',
      org: 'tokyo',
      projectId: 'student-coach-e1e95',
      serviceAccountEmailSuffix: 'student-coach-e1e95',
      namespace: 'prod-tokyo-services',
      kmsPath: 'projects/student-coach-e1e95/locations/asia-northeast1/keyRings/prod-tokyo/cryptoKeys/prod-tokyo',
      sqlProxyConnectionName: 'student-coach-e1e95:asia-northeast1:prod-tokyo',
      deployServiceAccount: 'prod-deploy-bot@student-coach-e1e95.iam.gserviceaccount.com',
      provider: 'projects/418860883682/locations/global/workloadIdentityPools/gh-action-pool/providers/gh-action-provider',
      platformServices: "backbone unleash connect ksql monitoring kiali redash import-map-deployer gateway appsmith data-warehouse dwh-with-auth replication",
    }
  }
}
