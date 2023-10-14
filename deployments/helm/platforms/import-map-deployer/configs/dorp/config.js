module.exports = {
  username: process.env.IMD_USERNAME,
  password: process.env.IMD_PASSWORD,
  manifestFormat: "importmap",
  locations: {
    tokyo: "gs://import-map-deployer-preproduction/tokyo.json",
    default: "gs://import-map-deployer-preproduction/default.json",
  },
};
