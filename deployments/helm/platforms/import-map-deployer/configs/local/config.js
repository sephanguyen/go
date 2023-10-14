module.exports = {
  username: process.env.IMD_USERNAME,
  password: process.env.IMD_PASSWORD,
  manifestFormat: "importmap",
  locations: {
    jprep: "/www/importmap.json",
    manabie: "/www/importmap.json",
    default: "/www/importmap.json",
  },
};
