module.exports = {
  username: process.env.IMD_USERNAME,
  password: process.env.IMD_PASSWORD,
  manifestFormat: "importmap",
  locations: {
    jprep: "gs://import-map-deployer-staging/jprep.json",
    manabie: "gs://import-map-deployer-staging/manabie.json",
    default: "gs://import-map-deployer-staging/default.json",
  },
};
