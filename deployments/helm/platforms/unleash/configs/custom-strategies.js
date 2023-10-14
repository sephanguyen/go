const unleash = require("unleash-client");

class EnvironmentStrategy extends unleash.Strategy {
  constructor() {
    super("strategy_environment");
  }

  isEnabled(parameters, context) {
    var env = context.properties.env;
    if (env === 'local') {
      env = 'stag';
    }
    return parameters.environments.split(",").map((e) => e.trim()).includes(env);
  }
}

class OrganizationStrategy extends unleash.Strategy {
  constructor() {
    super("strategy_organization");
  }

  isEnabled(parameters, context) {
    return parameters.organizations.split(",").includes(context.properties.org);
  }
}

class VariantStrategy extends unleash.Strategy {
  constructor() {
    super("strategy_variant");
  }

  isEnabled(parameters, context) {
    return parameters.variants.split(",").includes(context.properties.var);
  }
}

module.exports = [new EnvironmentStrategy(), new OrganizationStrategy(), new VariantStrategy()];
