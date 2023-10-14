module.exports = {
  username: process.env.IMD_USERNAME,
  password: process.env.IMD_PASSWORD,
  manifestFormat: "importmap",
  locations: {
    jprep: "gs://import-map-deployer-production/jprep.json",
    synersia: "gs://import-map-deployer-production/synersia.json",
    renseikai: "gs://import-map-deployer-production/renseikai.json",
    ga: "gs://import-map-deployer-production/ga.json",
    aic: "gs://import-map-deployer-production/aic.json",
    manabie: "gs://import-map-deployer-production/tokyo.json",
    tokyo: "gs://import-map-deployer-production/tokyo.json",
    default: "gs://import-map-deployer-production/default.json",
  },
};
